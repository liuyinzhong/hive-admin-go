package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const registrationTypeDictType = "MED_REGISTRATION_TYPE"

var medicalBusinessLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type MedicalRegistrationFeeService struct{}

func NewMedicalRegistrationFeeService() *MedicalRegistrationFeeService {
	return &MedicalRegistrationFeeService{}
}

type registrationFeeRuleListRow struct {
	models.MedRegistrationFeeRule
	DoctorNo       string `gorm:"column:doctor_no"`
	DoctorName     string `gorm:"column:doctor_name"`
	DepartmentCode string `gorm:"column:department_code"`
	DepartmentName string `gorm:"column:department_name"`
}

func (s *MedicalRegistrationFeeService) GetRegistrationFeeRuleList(req models.RegistrationFeeRuleListRequest) (*utils.PageResult, error) {
	query := database.DB.Table("med_registration_fee_rule AS fee").
		Select("fee.*, doctor.doctor_no, doctor.name AS doctor_name, department.department_code, department.department_name").
		Joins("JOIN med_doctor AS doctor ON doctor.doctor_id = fee.doctor_id AND doctor.del_flag = 0").
		Joins("JOIN med_department AS department ON department.department_id = fee.department_id AND department.del_flag = 0").
		Where("fee.del_flag = 0")

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("doctor.name LIKE ? OR doctor.doctor_no LIKE ? OR department.department_name LIKE ? OR department.department_code LIKE ?", like, like, like, like)
	}
	if req.DoctorID != "" {
		if err := validateMedicalUUID(req.DoctorID, "医生ID"); err != nil {
			return nil, err
		}
		query = query.Where("fee.doctor_id = ?", req.DoctorID)
	}
	if req.DepartmentID != "" {
		if err := validateMedicalUUID(req.DepartmentID, "临床科室ID"); err != nil {
			return nil, err
		}
		query = query.Where("fee.department_id = ?", req.DepartmentID)
	}
	if req.RegistrationType != "" {
		query = query.Where("fee.registration_type = ?", req.RegistrationType)
	}

	today := medicalToday()
	switch req.PeriodStatus {
	case "":
	case "current":
		query = query.Where("fee.effective_date <= ? AND (fee.expiry_date IS NULL OR fee.expiry_date >= ?)", today, today)
	case "future":
		query = query.Where("fee.effective_date > ?", today)
	case "expired":
		query = query.Where("fee.expiry_date IS NOT NULL AND fee.expiry_date < ?", today)
	default:
		return nil, fmt.Errorf("%w: 有效期状态不正确", ErrMedicalInvalidInput)
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"doctorName":     "doctor.name",
		"departmentName": "department.department_name",
		"feeAmount":      "fee.fee_amount",
		"effectiveDate":  "fee.effective_date",
		"expiryDate":     "fee.expiry_date",
		"version":        "fee.version",
		"createDate":     "fee.create_date",
	})
	if order == "" {
		order = "fee.effective_date desc, fee.version desc, fee.create_date desc"
	}
	query = query.Order(order)

	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var rows []registrationFeeRuleListRow
	pageResult, err := utils.Paginate(query, req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}

	items := make([]*models.RegistrationFeeRuleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, registrationFeeRuleToResponse(row, today))
	}
	pageResult.Items = items
	return pageResult, nil
}

func (s *MedicalRegistrationFeeService) CreateRegistrationFeeRule(req models.CreateRegistrationFeeRuleRequest, operatorID string) error {
	feeAmount, err := normalizeRegistrationFeeAmount(req.FeeAmount)
	if err != nil {
		return err
	}
	effectiveDate, expiryDate, err := validateRegistrationFeePeriod(req.EffectiveDate, req.ExpiryDate)
	if err != nil {
		return err
	}
	if medicalDateBefore(effectiveDate, medicalToday()) {
		return fmt.Errorf("%w: 生效日期不能早于今天", ErrMedicalInvalidInput)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.validateRegistrationFeeDimension(tx, req.DoctorID, req.DepartmentID, req.RegistrationType, effectiveDate, expiryDate); err != nil {
			return err
		}

		rules, err := lockRegistrationFeeRules(tx, req.DoctorID, req.DepartmentID, req.RegistrationType)
		if err != nil {
			return err
		}
		if registrationFeeRulesOverlap(rules, effectiveDate, expiryDate, "") {
			return fmt.Errorf("%w: 同一医生、科室和挂号类型的有效期不能重叠", ErrMedicalConflict)
		}

		now := time.Now()
		rule := models.MedRegistrationFeeRule{
			FeeRuleID:        utils.GenerateUUID(),
			DoctorID:         req.DoctorID,
			DepartmentID:     req.DepartmentID,
			RegistrationType: req.RegistrationType,
			FeeAmount:        feeAmount,
			EffectiveDate:    effectiveDate,
			ExpiryDate:       expiryDate,
			Version:          nextRegistrationFeeVersion(rules),
			Remark:           normalizeMedicalOptionalString(req.Remark),
			CreatorID:        optionalOperatorID(operatorID),
			UpdaterID:        optionalOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		}
		if err := tx.Create(&rule).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *MedicalRegistrationFeeService) AdjustRegistrationFeeRule(feeRuleID string, req models.AdjustRegistrationFeeRuleRequest, operatorID string) error {
	if err := validateMedicalUUID(feeRuleID, "挂号费规则ID"); err != nil {
		return err
	}
	feeAmount, err := normalizeRegistrationFeeAmount(req.FeeAmount)
	if err != nil {
		return err
	}
	effectiveDate, _, err := validateRegistrationFeePeriod(req.EffectiveDate, nil)
	if err != nil {
		return err
	}
	if medicalDateBefore(effectiveDate, medicalToday()) {
		return fmt.Errorf("%w: 调价生效日期不能早于今天", ErrMedicalInvalidInput)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var baseRule models.MedRegistrationFeeRule
		if err := tx.Where("fee_rule_id = ? AND del_flag = 0", feeRuleID).First(&baseRule).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("%w: 挂号费规则不存在", ErrMedicalNotFound)
			}
			return err
		}
		if err := s.validateRegistrationFeeDimension(tx, baseRule.DoctorID, baseRule.DepartmentID, baseRule.RegistrationType, effectiveDate, baseRule.ExpiryDate); err != nil {
			return err
		}
		rules, err := lockRegistrationFeeRules(tx, baseRule.DoctorID, baseRule.DepartmentID, baseRule.RegistrationType)
		if err != nil {
			return err
		}
		currentBaseRule := findRegistrationFeeRule(rules, feeRuleID)
		if currentBaseRule == nil {
			return fmt.Errorf("%w: 挂号费规则不存在", ErrMedicalNotFound)
		}
		baseRule = *currentBaseRule
		if !medicalDateAfter(effectiveDate, baseRule.EffectiveDate) {
			return fmt.Errorf("%w: 调价生效日期必须晚于原规则生效日期", ErrMedicalInvalidInput)
		}
		if baseRule.ExpiryDate != nil && medicalDateAfter(effectiveDate, *baseRule.ExpiryDate) {
			return fmt.Errorf("%w: 调价生效日期不能晚于原规则失效日期", ErrMedicalInvalidInput)
		}
		if registrationFeeRulesOverlap(rules, effectiveDate, baseRule.ExpiryDate, baseRule.FeeRuleID) {
			return fmt.Errorf("%w: 调价后的有效期与现有规则重叠", ErrMedicalConflict)
		}

		now := time.Now()
		baseExpiryDate := effectiveDate.AddDate(0, 0, -1)
		if err := tx.Model(&models.MedRegistrationFeeRule{}).
			Where("fee_rule_id = ? AND del_flag = 0", baseRule.FeeRuleID).
			Updates(map[string]interface{}{
				"expiry_date": baseExpiryDate,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error; err != nil {
			return err
		}

		newRule := models.MedRegistrationFeeRule{
			FeeRuleID:        utils.GenerateUUID(),
			DoctorID:         baseRule.DoctorID,
			DepartmentID:     baseRule.DepartmentID,
			RegistrationType: baseRule.RegistrationType,
			FeeAmount:        feeAmount,
			EffectiveDate:    effectiveDate,
			ExpiryDate:       baseRule.ExpiryDate,
			Version:          nextRegistrationFeeVersion(rules),
			Remark:           normalizeMedicalOptionalString(req.Remark),
			CreatorID:        optionalOperatorID(operatorID),
			UpdaterID:        optionalOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		}
		if err := tx.Create(&newRule).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *MedicalRegistrationFeeService) validateRegistrationFeeDimension(tx *gorm.DB, doctorID, departmentID, registrationType string, effectiveDate time.Time, expiryDate *time.Time) error {
	if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
		return err
	}
	if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
		return err
	}
	if strings.TrimSpace(registrationType) == "" {
		return fmt.Errorf("%w: 挂号类型不能为空", ErrMedicalInvalidInput)
	}

	var doctorCount int64
	if err := tx.Model(&models.MedDoctor{}).Where("doctor_id = ? AND status = 1 AND del_flag = 0", doctorID).Count(&doctorCount).Error; err != nil {
		return err
	}
	if doctorCount == 0 {
		return fmt.Errorf("%w: 医生不存在或已停用", ErrMedicalInvalidInput)
	}

	var departmentCount int64
	if err := tx.Model(&models.MedDepartment{}).Where("department_id = ? AND status = 1 AND del_flag = 0", departmentID).Count(&departmentCount).Error; err != nil {
		return err
	}
	if departmentCount == 0 {
		return fmt.Errorf("%w: 临床科室不存在或已停用", ErrMedicalInvalidInput)
	}

	var relation models.MedDoctorDepartment
	// 以稳定存在的医生科室关系行作为费用维度锁，避免首次并发创建时空范围锁互相兼容。
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("doctor_id = ? AND department_id = ? AND status = 1 AND appointment_enabled = 1 AND del_flag = 0", doctorID, departmentID).First(&relation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%w: 医生未绑定该出诊科室或该科室未开放预约", ErrMedicalInvalidInput)
		}
		return err
	}
	if relation.ValidFrom != nil && medicalDateBefore(effectiveDate, *relation.ValidFrom) {
		return fmt.Errorf("%w: 挂号费生效日期不能早于医生科室关系生效日期", ErrMedicalInvalidInput)
	}
	if relation.ValidTo != nil && (expiryDate == nil || medicalDateAfter(*expiryDate, *relation.ValidTo)) {
		return fmt.Errorf("%w: 挂号费有效期不能超过医生科室关系有效期", ErrMedicalInvalidInput)
	}

	var dictCount int64
	if err := tx.Table("sys_dict AS item").
		Joins("JOIN sys_dict AS root ON root.id = item.pid AND root.del_flag = 0 AND root.status = 1").
		Where("item.type = ? AND item.value = ? AND item.del_flag = 0 AND item.status = 1", registrationTypeDictType, registrationType).
		Count(&dictCount).Error; err != nil {
		return err
	}
	if dictCount == 0 {
		return fmt.Errorf("%w: 挂号类型字典值不存在或已停用", ErrMedicalInvalidInput)
	}
	return nil
}

func lockRegistrationFeeRules(tx *gorm.DB, doctorID, departmentID, registrationType string) ([]models.MedRegistrationFeeRule, error) {
	var rules []models.MedRegistrationFeeRule
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("doctor_id = ? AND department_id = ? AND registration_type = ? AND del_flag = 0", doctorID, departmentID, registrationType).
		Order("version asc").
		Find(&rules).Error
	return rules, err
}

func findRegistrationFeeRule(rules []models.MedRegistrationFeeRule, feeRuleID string) *models.MedRegistrationFeeRule {
	for index := range rules {
		if rules[index].FeeRuleID == feeRuleID {
			return &rules[index]
		}
	}
	return nil
}

func registrationFeeRulesOverlap(rules []models.MedRegistrationFeeRule, effectiveDate time.Time, expiryDate *time.Time, excludeFeeRuleID string) bool {
	for _, rule := range rules {
		if rule.FeeRuleID == excludeFeeRuleID {
			continue
		}
		if registrationFeePeriodsOverlap(rule.EffectiveDate, rule.ExpiryDate, effectiveDate, expiryDate) {
			return true
		}
	}
	return false
}

func registrationFeePeriodsOverlap(leftStart time.Time, leftEnd *time.Time, rightStart time.Time, rightEnd *time.Time) bool {
	if leftEnd != nil && medicalDateBefore(*leftEnd, rightStart) {
		return false
	}
	if rightEnd != nil && medicalDateBefore(*rightEnd, leftStart) {
		return false
	}
	return true
}

func nextRegistrationFeeVersion(rules []models.MedRegistrationFeeRule) int {
	version := 0
	for _, rule := range rules {
		if rule.Version > version {
			version = rule.Version
		}
	}
	return version + 1
}

func normalizeRegistrationFeeAmount(value string) (string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > 8 {
		return "", fmt.Errorf("%w: 挂号费金额格式不正确", ErrMedicalInvalidInput)
	}
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("%w: 挂号费金额格式不正确", ErrMedicalInvalidInput)
		}
	}
	if len(parts[0]) > 1 && parts[0][0] == '0' {
		return "", fmt.Errorf("%w: 挂号费金额格式不正确", ErrMedicalInvalidInput)
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) == 0 || len(fraction) > 2 {
			return "", fmt.Errorf("%w: 挂号费金额最多保留两位小数", ErrMedicalInvalidInput)
		}
		for _, ch := range fraction {
			if ch < '0' || ch > '9' {
				return "", fmt.Errorf("%w: 挂号费金额格式不正确", ErrMedicalInvalidInput)
			}
		}
	}

	yuan, _ := strconv.ParseInt(parts[0], 10, 64)
	for len(fraction) < 2 {
		fraction += "0"
	}
	cent := int64(0)
	if fraction != "" {
		cent, _ = strconv.ParseInt(fraction, 10, 64)
	}
	if yuan == 0 && cent == 0 {
		return "", fmt.Errorf("%w: 挂号费金额必须大于0", ErrMedicalInvalidInput)
	}
	return fmt.Sprintf("%d.%02d", yuan, cent), nil
}

func validateRegistrationFeePeriod(effectiveDateValue string, expiryDateValue *string) (time.Time, *time.Time, error) {
	effectiveDate, err := parseMedicalDate(&effectiveDateValue, "生效日期")
	if err != nil || effectiveDate == nil {
		if err != nil {
			return time.Time{}, nil, err
		}
		return time.Time{}, nil, fmt.Errorf("%w: 生效日期不能为空", ErrMedicalInvalidInput)
	}
	expiryDate, err := parseMedicalDate(expiryDateValue, "失效日期")
	if err != nil {
		return time.Time{}, nil, err
	}
	if expiryDate != nil && medicalDateBefore(*expiryDate, *effectiveDate) {
		return time.Time{}, nil, fmt.Errorf("%w: 失效日期不能早于生效日期", ErrMedicalInvalidInput)
	}
	return *effectiveDate, expiryDate, nil
}

func medicalToday() time.Time {
	now := time.Now().In(medicalBusinessLocation)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, medicalBusinessLocation)
}

func medicalDateBefore(left, right time.Time) bool {
	return left.Format("2006-01-02") < right.Format("2006-01-02")
}

func medicalDateAfter(left, right time.Time) bool {
	return left.Format("2006-01-02") > right.Format("2006-01-02")
}

func registrationFeeRuleToResponse(row registrationFeeRuleListRow, today time.Time) *models.RegistrationFeeRuleResponse {
	periodStatus := "current"
	if medicalDateAfter(row.EffectiveDate, today) {
		periodStatus = "future"
	} else if row.ExpiryDate != nil && medicalDateBefore(*row.ExpiryDate, today) {
		periodStatus = "expired"
	}
	return &models.RegistrationFeeRuleResponse{
		FeeRuleID:        row.FeeRuleID,
		DoctorID:         row.DoctorID,
		DoctorNo:         row.DoctorNo,
		DoctorName:       row.DoctorName,
		DepartmentID:     row.DepartmentID,
		DepartmentCode:   row.DepartmentCode,
		DepartmentName:   row.DepartmentName,
		RegistrationType: row.RegistrationType,
		FeeAmount:        row.FeeAmount,
		EffectiveDate:    row.EffectiveDate.Format("2006-01-02"),
		ExpiryDate:       timeToDateStringPtr(row.ExpiryDate),
		Version:          row.Version,
		PeriodStatus:     periodStatus,
		Remark:           row.Remark,
		CreateDate:       models.TimeToStringPtr(row.CreateDate),
		UpdateDate:       models.TimeToStringPtr(row.UpdateDate),
	}
}

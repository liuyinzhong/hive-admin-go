package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const (
	patientGenderDictType     = "GENDER"
	patientIDTypeDictType     = "MED_PATIENT_ID_TYPE"
	patientResidentIDCardType = "ID_CARD"
)

var patientPhonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

type MedicalPatientService struct{}

type preparedPatientRequest struct {
	name                     string
	gender                   string
	birthDate                time.Time
	idType                   string
	idNumber                 string
	phone                    string
	address                  *string
	emergencyContactName     *string
	emergencyContactRelation *string
	emergencyContactPhone    *string
	remark                   *string
}

func NewMedicalPatientService() *MedicalPatientService {
	return &MedicalPatientService{}
}

func (s *MedicalPatientService) GetPatientList(req models.PatientListRequest, permission datapermission.Permission) (*utils.PageResult, error) {
	query := database.DB.Model(&models.MedPatient{}).Where("med_patient.del_flag = 0")
	query = permission.Apply(query, "med_patient.creator_id")
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			"med_patient.patient_no LIKE ? OR med_patient.name LIKE ? OR med_patient.phone LIKE ? OR med_patient.id_number LIKE ?",
			like, like, like, like,
		)
	}
	if strings.TrimSpace(req.Gender) != "" {
		query = query.Where("med_patient.gender = ?", strings.TrimSpace(req.Gender))
	}
	if req.Status != nil {
		if err := validateMedicalStatus(*req.Status); err != nil {
			return nil, err
		}
		query = query.Where("med_patient.status = ?", *req.Status)
	}
	if req.CreateDateFrom != "" {
		from, err := parsePatientQueryDate(req.CreateDateFrom, "创建开始日期", false)
		if err != nil {
			return nil, err
		}
		query = query.Where("med_patient.create_date >= ?", from)
	}
	if req.CreateDateTo != "" {
		to, err := parsePatientQueryDate(req.CreateDateTo, "创建结束日期", true)
		if err != nil {
			return nil, err
		}
		query = query.Where("med_patient.create_date < ?", to)
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"patientNo":  "patient_no",
		"name":       "name",
		"birthDate":  "birth_date",
		"status":     "status",
		"createDate": "create_date",
	})
	if order == "" {
		order = "create_date desc"
	}
	query = query.Order(order)

	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var patients []models.MedPatient
	pageResult, err := utils.Paginate(query, req.Page, pageSize, &patients)
	if err != nil {
		return nil, err
	}
	items := make([]*models.PatientResponse, 0, len(patients))
	for _, patient := range patients {
		items = append(items, patientToResponse(patient, false))
	}
	pageResult.Items = items
	return pageResult, nil
}

func (s *MedicalPatientService) GetPatientDetail(patientID string, showSensitive bool, permission datapermission.Permission) (*models.PatientResponse, error) {
	if err := validateMedicalUUID(patientID, "患者ID"); err != nil {
		return nil, err
	}
	var patient models.MedPatient
	query := database.DB.Model(&models.MedPatient{}).Where("patient_id = ? AND del_flag = 0", patientID)
	if err := permission.Apply(query, "med_patient.creator_id").First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 患者不存在", ErrMedicalNotFound)
		}
		return nil, err
	}
	return patientToResponse(patient, showSensitive), nil
}

func (s *MedicalPatientService) CreatePatient(req models.SavePatientRequest, operatorID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		prepared, err := s.preparePatientRequest(tx, req, "")
		if err != nil {
			return err
		}
		patientNo, err := NewBaseCodeSequenceService().NextBusinessCode(tx, "PATIENT", "PAT", 6)
		if err != nil {
			return err
		}
		now := time.Now()
		patient := models.MedPatient{
			PatientID:                utils.GenerateUUID(),
			PatientNo:                patientNo,
			Name:                     prepared.name,
			Gender:                   prepared.gender,
			BirthDate:                prepared.birthDate,
			IDType:                   prepared.idType,
			IDNumber:                 prepared.idNumber,
			Phone:                    prepared.phone,
			Address:                  prepared.address,
			EmergencyContactName:     prepared.emergencyContactName,
			EmergencyContactRelation: prepared.emergencyContactRelation,
			EmergencyContactPhone:    prepared.emergencyContactPhone,
			Remark:                   prepared.remark,
			Status:                   1,
			CreatorID:                optionalOperatorID(operatorID),
			UpdaterID:                optionalOperatorID(operatorID),
			CreateDate:               &now,
			UpdateDate:               &now,
			DelFlag:                  0,
		}
		return tx.Create(&patient).Error
	})
}

func (s *MedicalPatientService) UpdatePatient(patientID string, req models.SavePatientRequest, operatorID string, permission datapermission.Permission) error {
	if err := validateMedicalUUID(patientID, "患者ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var patient models.MedPatient
		query := tx.Model(&models.MedPatient{}).Where("patient_id = ? AND del_flag = 0", patientID)
		if err := permission.Apply(query, "med_patient.creator_id").First(&patient).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 患者不存在", ErrMedicalNotFound)
			}
			return err
		}
		prepared, err := s.preparePatientRequest(tx, req, patientID)
		if err != nil {
			return err
		}
		return tx.Model(&models.MedPatient{}).Where("patient_id = ?", patientID).Updates(map[string]interface{}{
			"name":                       prepared.name,
			"gender":                     prepared.gender,
			"birth_date":                 prepared.birthDate,
			"id_type":                    prepared.idType,
			"id_number":                  prepared.idNumber,
			"phone":                      prepared.phone,
			"address":                    prepared.address,
			"emergency_contact_name":     prepared.emergencyContactName,
			"emergency_contact_relation": prepared.emergencyContactRelation,
			"emergency_contact_phone":    prepared.emergencyContactPhone,
			"remark":                     prepared.remark,
			"updater_id":                 optionalOperatorID(operatorID),
			"update_date":                time.Now(),
		}).Error
	})
}

func (s *MedicalPatientService) UpdatePatientStatus(patientID string, status int, operatorID string, permission datapermission.Permission) error {
	if err := validateMedicalUUID(patientID, "患者ID"); err != nil {
		return err
	}
	if err := validateMedicalStatus(status); err != nil {
		return err
	}
	query := database.DB.Model(&models.MedPatient{}).Where("patient_id = ? AND del_flag = 0", patientID)
	result := permission.Apply(query, "med_patient.creator_id").
		Updates(map[string]interface{}{
			"status":      status,
			"updater_id":  optionalOperatorID(operatorID),
			"update_date": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 患者不存在", ErrMedicalNotFound)
	}
	return nil
}

func (s *MedicalPatientService) preparePatientRequest(tx *gorm.DB, req models.SavePatientRequest, excludePatientID string) (*preparedPatientRequest, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 患者姓名不能为空", ErrMedicalInvalidInput)
	}
	if utf8.RuneCountInString(name) > 64 || containsControlRune(name) {
		return nil, fmt.Errorf("%w: 患者姓名长度或字符不合法", ErrMedicalInvalidInput)
	}

	gender := strings.TrimSpace(req.Gender)
	if err := s.validatePatientDictValue(tx, patientGenderDictType, gender, true); err != nil {
		return nil, err
	}
	idType := strings.TrimSpace(req.IDType)
	if err := s.validatePatientDictValue(tx, patientIDTypeDictType, idType, true); err != nil {
		return nil, err
	}
	idNumber := strings.ToUpper(strings.TrimSpace(req.IDNumber))
	if idNumber == "" || len(idNumber) > 128 || containsControlRune(idNumber) {
		return nil, fmt.Errorf("%w: 证件号码不能为空且长度不能超过128个字符", ErrMedicalInvalidInput)
	}
	if idType == patientResidentIDCardType && !validateResidentIDCard(idNumber) {
		return nil, fmt.Errorf("%w: 居民身份证号码格式或校验码错误", ErrMedicalInvalidInput)
	}
	if err := s.ensurePatientIdentityUnique(tx, idType, idNumber, excludePatientID); err != nil {
		return nil, err
	}

	phone := strings.TrimSpace(req.Phone)
	if !patientPhonePattern.MatchString(phone) {
		return nil, fmt.Errorf("%w: 手机号必须是11位中国大陆手机号", ErrMedicalInvalidInput)
	}
	birthDate, err := parseMedicalDate(&req.BirthDate, "出生日期")
	if err != nil || birthDate == nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%w: 出生日期不能为空", ErrMedicalInvalidInput)
	}
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	if birthDate.After(today) {
		return nil, fmt.Errorf("%w: 出生日期不能晚于当前日期", ErrMedicalInvalidInput)
	}

	address := normalizeMedicalOptionalString(req.Address)
	if err := validatePatientOptionalLength(address, 512, "常住地址"); err != nil {
		return nil, err
	}
	emergencyName := normalizeMedicalOptionalString(req.EmergencyContactName)
	emergencyRelation := normalizeMedicalOptionalString(req.EmergencyContactRelation)
	emergencyPhone := normalizeMedicalOptionalString(req.EmergencyContactPhone)
	emergencyFilled := emergencyName != nil || emergencyRelation != nil || emergencyPhone != nil
	if emergencyFilled && (emergencyName == nil || emergencyRelation == nil || emergencyPhone == nil) {
		return nil, fmt.Errorf("%w: 紧急联系人姓名、关系和手机号需要全部填写", ErrMedicalInvalidInput)
	}
	if err := validatePatientOptionalLength(emergencyName, 64, "紧急联系人姓名"); err != nil {
		return nil, err
	}
	if err := validatePatientOptionalLength(emergencyRelation, 64, "紧急联系人关系"); err != nil {
		return nil, err
	}
	if emergencyPhone != nil && !patientPhonePattern.MatchString(*emergencyPhone) {
		return nil, fmt.Errorf("%w: 紧急联系人手机号必须是11位中国大陆手机号", ErrMedicalInvalidInput)
	}
	remark := normalizeMedicalOptionalString(req.Remark)
	if err := validatePatientOptionalLength(remark, 512, "备注"); err != nil {
		return nil, err
	}

	return &preparedPatientRequest{
		name:                     name,
		gender:                   gender,
		birthDate:                *birthDate,
		idType:                   idType,
		idNumber:                 idNumber,
		phone:                    phone,
		address:                  address,
		emergencyContactName:     emergencyName,
		emergencyContactRelation: emergencyRelation,
		emergencyContactPhone:    emergencyPhone,
		remark:                   remark,
	}, nil
}

func (s *MedicalPatientService) ensurePatientIdentityUnique(tx *gorm.DB, idType, idNumber, excludePatientID string) error {
	query := tx.Model(&models.MedPatient{}).Where("id_type = ? AND id_number = ? AND del_flag = 0", idType, idNumber)
	if excludePatientID != "" {
		query = query.Where("patient_id != ?", excludePatientID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 证件类型和证件号码已关联其他患者档案", ErrMedicalConflict)
	}
	return nil
}

func (s *MedicalPatientService) validatePatientDictValue(tx *gorm.DB, dictType, value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return fmt.Errorf("%w: 字典值不能为空", ErrMedicalInvalidInput)
		}
		return nil
	}
	var count int64
	if err := tx.Model(&models.SysDict{}).Where("type = ? AND value = ? AND status = 1 AND del_flag = 0", dictType, value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: 字典%s中不存在值%s", ErrMedicalInvalidInput, dictType, value)
	}
	return nil
}

func parsePatientQueryDate(value, label string, endOfDay bool) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s格式必须为YYYY-MM-DD", ErrMedicalInvalidInput, label)
	}
	if endOfDay {
		return parsed.AddDate(0, 0, 1), nil
	}
	return parsed, nil
}

func validatePatientOptionalLength(value *string, maxLength int, label string) error {
	if value != nil && utf8.RuneCountInString(*value) > maxLength {
		return fmt.Errorf("%w: %s不能超过%d个字符", ErrMedicalInvalidInput, label, maxLength)
	}
	return nil
}

func containsControlRune(value string) bool {
	for _, char := range value {
		if unicode.IsControl(char) {
			return true
		}
	}
	return false
}

func validateResidentIDCard(value string) bool {
	if len(value) != 18 {
		return false
	}
	for index := 0; index < 17; index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	if value[17] != 'X' && (value[17] < '0' || value[17] > '9') {
		return false
	}
	if _, err := time.Parse("20060102", value[6:14]); err != nil {
		return false
	}
	weights := [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checks := "10X98765432"
	sum := 0
	for index, weight := range weights {
		sum += int(value[index]-'0') * weight
	}
	return value[17] == checks[sum%11]
}

func patientToResponse(patient models.MedPatient, showSensitive bool) *models.PatientResponse {
	name := patient.Name
	idNumber := patient.IDNumber
	phone := patient.Phone
	if !showSensitive {
		name = maskPatientName(name)
		idNumber = maskPatientIDNumber(idNumber)
		phone = maskPatientPhone(phone)
	}
	return &models.PatientResponse{
		PatientID:                patient.PatientID,
		PatientNo:                patient.PatientNo,
		Name:                     name,
		Gender:                   patient.Gender,
		BirthDate:                patient.BirthDate.Format("2006-01-02"),
		IDType:                   patient.IDType,
		IDNumber:                 idNumber,
		Phone:                    phone,
		Address:                  patient.Address,
		EmergencyContactName:     patient.EmergencyContactName,
		EmergencyContactRelation: patient.EmergencyContactRelation,
		EmergencyContactPhone:    patient.EmergencyContactPhone,
		Remark:                   patient.Remark,
		Status:                   patient.Status,
		CreateDate:               models.TimeToStringPtr(patient.CreateDate),
		UpdateDate:               models.TimeToStringPtr(patient.UpdateDate),
	}
}

func maskPatientName(value string) string {
	runes := []rune(value)
	if len(runes) <= 1 {
		return "*"
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}

func maskPatientIDNumber(value string) string {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	if len(value) <= 7 {
		return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
	}
	return value[:3] + strings.Repeat("*", len(value)-7) + value[len(value)-4:]
}

func maskPatientPhone(value string) string {
	if len(value) != 11 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + "****" + value[len(value)-4:]
}

package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const (
	maxScheduleGenerationDays = 90
	scheduleTimeLayout        = "15:04:05"
)

type preparedScheduleTemplate struct {
	templateName     string
	doctorID         string
	departmentID     string
	registrationType string
	weekday          int
	startTime        string
	endTime          string
	defaultSlotQuota int
	slotQuotaConfig  []models.ScheduleSlotQuotaRequest
	totalQuota       int
	effectiveDate    time.Time
	expiryDate       *time.Time
	status           int
	remark           *string
}

type preparedSchedule struct {
	doctorID         string
	departmentID     string
	registrationType string
	scheduleDate     time.Time
	startTime        string
	endTime          string
	defaultSlotQuota int
	slotQuotaConfig  []models.ScheduleSlotQuotaRequest
	totalQuota       int
	remark           *string
}

func prepareScheduleTemplateRequest(req models.SaveScheduleTemplateRequest) (*preparedScheduleTemplate, error) {
	templateName := strings.TrimSpace(req.TemplateName)
	if templateName == "" || len([]rune(templateName)) > 64 {
		return nil, fmt.Errorf("%w: 模板名称不能为空且最多64个字符", ErrMedicalInvalidInput)
	}
	if req.Weekday < 1 || req.Weekday > 7 {
		return nil, fmt.Errorf("%w: 星期必须在1到7之间", ErrMedicalInvalidInput)
	}
	if req.DefaultSlotQuota < 1 || req.DefaultSlotQuota > 99 {
		return nil, fmt.Errorf("%w: 每半小时默认容量必须在1到99之间", ErrMedicalInvalidInput)
	}
	if err := validateMedicalStatus(req.Status); err != nil {
		return nil, err
	}
	startTime, endTime, err := validateScheduleTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	_, totalQuota, normalizedConfig, err := buildScheduleSlotDrafts("", startTime, endTime, req.DefaultSlotQuota, req.SlotQuotaConfig, "")
	if err != nil {
		return nil, err
	}
	effectiveDate, expiryDate, err := validateRegistrationFeePeriod(req.EffectiveDate, req.ExpiryDate)
	if err != nil {
		return nil, err
	}
	return &preparedScheduleTemplate{
		templateName:     templateName,
		doctorID:         req.DoctorID,
		departmentID:     req.DepartmentID,
		registrationType: strings.TrimSpace(req.RegistrationType),
		weekday:          req.Weekday,
		startTime:        startTime,
		endTime:          endTime,
		defaultSlotQuota: req.DefaultSlotQuota,
		slotQuotaConfig:  normalizedConfig,
		totalQuota:       totalQuota,
		effectiveDate:    effectiveDate,
		expiryDate:       expiryDate,
		status:           req.Status,
		remark:           normalizeMedicalOptionalString(req.Remark),
	}, nil
}

func prepareScheduleRequest(req models.SaveScheduleRequest) (*preparedSchedule, error) {
	if req.DefaultSlotQuota < 1 || req.DefaultSlotQuota > 99 {
		return nil, fmt.Errorf("%w: 每半小时默认容量必须在1到99之间", ErrMedicalInvalidInput)
	}
	startTime, endTime, err := validateScheduleTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	_, totalQuota, normalizedConfig, err := buildScheduleSlotDrafts("", startTime, endTime, req.DefaultSlotQuota, req.SlotQuotaConfig, "")
	if err != nil {
		return nil, err
	}
	scheduleDate, err := parseRequiredMedicalDate(req.ScheduleDate, "出诊日期")
	if err != nil {
		return nil, err
	}
	if medicalDateBefore(scheduleDate, medicalToday()) {
		return nil, fmt.Errorf("%w: 出诊日期不能早于今天", ErrMedicalInvalidInput)
	}
	return &preparedSchedule{
		doctorID:         req.DoctorID,
		departmentID:     req.DepartmentID,
		registrationType: strings.TrimSpace(req.RegistrationType),
		scheduleDate:     scheduleDate,
		startTime:        startTime,
		endTime:          endTime,
		defaultSlotQuota: req.DefaultSlotQuota,
		slotQuotaConfig:  normalizedConfig,
		totalQuota:       totalQuota,
		remark:           normalizeMedicalOptionalString(req.Remark),
	}, nil
}

func validateScheduleTimeRange(startValue, endValue string) (string, string, error) {
	startTime, err := normalizeScheduleTime(startValue, "开始时间")
	if err != nil {
		return "", "", err
	}
	endTime, err := normalizeScheduleTime(endValue, "结束时间")
	if err != nil {
		return "", "", err
	}
	if startTime >= endTime {
		return "", "", fmt.Errorf("%w: 结束时间必须晚于开始时间", ErrMedicalInvalidInput)
	}
	return startTime, endTime, nil
}

func normalizeScheduleTime(value, label string) (string, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04", scheduleTimeLayout} {
		parsed, err := time.ParseInLocation(layout, value, medicalBusinessLocation)
		if err == nil {
			return parsed.Format(scheduleTimeLayout), nil
		}
	}
	return "", fmt.Errorf("%w: %s格式必须为HH:mm或HH:mm:ss", ErrMedicalInvalidInput, label)
}

func buildScheduleSlotDrafts(scheduleID, startTime, endTime string, defaultQuota int, overrides []models.ScheduleSlotQuotaRequest, operatorID string) ([]models.MedScheduleSlot, int, []models.ScheduleSlotQuotaRequest, error) {
	if defaultQuota < 1 || defaultQuota > 99 {
		return nil, 0, nil, fmt.Errorf("%w: 每半小时默认容量必须在1到99之间", ErrMedicalInvalidInput)
	}
	start, err := time.ParseInLocation(scheduleTimeLayout, startTime, medicalBusinessLocation)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("%w: 开始时间格式不正确", ErrMedicalInvalidInput)
	}
	end, err := time.ParseInLocation(scheduleTimeLayout, endTime, medicalBusinessLocation)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("%w: 结束时间格式不正确", ErrMedicalInvalidInput)
	}
	if start.Minute()%30 != 0 || start.Second() != 0 || end.Minute()%30 != 0 || end.Second() != 0 {
		return nil, 0, nil, fmt.Errorf("%w: 排班时间必须为整点或半点", ErrMedicalInvalidInput)
	}
	if !end.After(start) || end.Sub(start) < 30*time.Minute {
		return nil, 0, nil, fmt.Errorf("%w: 排班时长至少为30分钟", ErrMedicalInvalidInput)
	}

	overrideMap := make(map[string]int, len(overrides))
	for _, override := range overrides {
		normalized, err := normalizeScheduleTime(override.StartTime, "档位开始时间")
		if err != nil {
			return nil, 0, nil, err
		}
		parsed, _ := time.ParseInLocation(scheduleTimeLayout, normalized, medicalBusinessLocation)
		if parsed.Minute()%30 != 0 || override.Quota < 0 || override.Quota > 99 {
			return nil, 0, nil, fmt.Errorf("%w: 单档容量必须在0到99之间且时间为整点或半点", ErrMedicalInvalidInput)
		}
		if _, exists := overrideMap[normalized]; exists {
			return nil, 0, nil, fmt.Errorf("%w: 单档容量配置不能重复", ErrMedicalInvalidInput)
		}
		overrideMap[normalized] = override.Quota
	}

	now := time.Now()
	slots := make([]models.MedScheduleSlot, 0, int(end.Sub(start)/(30*time.Minute)))
	normalizedOverrides := make([]models.ScheduleSlotQuotaRequest, 0, len(overrides))
	totalQuota := 0
	for current := start; current.Before(end); current = current.Add(30 * time.Minute) {
		key := current.Format(scheduleTimeLayout)
		quota := defaultQuota
		if override, exists := overrideMap[key]; exists {
			quota = override
			normalizedOverrides = append(normalizedOverrides, models.ScheduleSlotQuotaRequest{StartTime: trimScheduleTime(key), Quota: quota})
			delete(overrideMap, key)
		}
		slotID := ""
		if scheduleID != "" {
			slotID = utils.GenerateUUID()
		}
		slots = append(slots, models.MedScheduleSlot{
			SlotID:     slotID,
			ScheduleID: scheduleID,
			StartTime:  key,
			EndTime:    current.Add(30 * time.Minute).Format(scheduleTimeLayout),
			Quota:      quota,
			CreatorID:  optionalOperatorID(operatorID),
			UpdaterID:  optionalOperatorID(operatorID),
			CreateDate: &now,
			UpdateDate: &now,
		})
		totalQuota += quota
	}
	if len(overrideMap) > 0 {
		return nil, 0, nil, fmt.Errorf("%w: 单档容量时间必须位于排班范围内", ErrMedicalInvalidInput)
	}
	if totalQuota == 0 {
		return nil, 0, nil, fmt.Errorf("%w: 整段排班至少需要一个可预约号源", ErrMedicalInvalidInput)
	}
	sort.Slice(normalizedOverrides, func(i, j int) bool { return normalizedOverrides[i].StartTime < normalizedOverrides[j].StartTime })
	return slots, totalQuota, normalizedOverrides, nil
}

func marshalScheduleSlotQuotaConfig(config []models.ScheduleSlotQuotaRequest) (*string, error) {
	if len(config) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	value := string(encoded)
	return &value, nil
}

func unmarshalScheduleSlotQuotaConfig(value *string) ([]models.ScheduleSlotQuotaRequest, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return []models.ScheduleSlotQuotaRequest{}, nil
	}
	var config []models.ScheduleSlotQuotaRequest
	if err := json.Unmarshal([]byte(*value), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func parseRequiredMedicalDate(value, label string) (time.Time, error) {
	parsed, err := parseMedicalDate(&value, label)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, fmt.Errorf("%w: %s不能为空", ErrMedicalInvalidInput, label)
	}
	return *parsed, nil
}

func scheduleTimesOverlap(leftStart, leftEnd, rightStart, rightEnd string) bool {
	return leftStart < rightEnd && leftEnd > rightStart
}

func scheduleDatePeriodsOverlap(leftStart time.Time, leftEnd *time.Time, rightStart time.Time, rightEnd *time.Time) bool {
	if leftEnd != nil && medicalDateBefore(*leftEnd, rightStart) {
		return false
	}
	if rightEnd != nil && medicalDateBefore(*rightEnd, leftStart) {
		return false
	}
	return true
}

func isoWeekday(value time.Time) int {
	weekday := int(value.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

func scheduleDateTime(scheduleDate time.Time, scheduleTime string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 "+scheduleTimeLayout, scheduleDate.Format("2006-01-02")+" "+scheduleTime, medicalBusinessLocation)
}

func validateScheduleStatus(status int) error {
	if status < models.MedScheduleStatusDraft || status > models.MedScheduleStatusFinished {
		return fmt.Errorf("%w: 排班状态不正确", ErrMedicalInvalidInput)
	}
	return nil
}

func normalizeScheduleUUIDs(values []string, label string, max int) ([]string, error) {
	if len(values) == 0 || len(values) > max {
		return nil, fmt.Errorf("%w: %s数量必须在1到%d之间", ErrMedicalInvalidInput, label, max)
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if err := validateMedicalUUID(value, label); err != nil {
			return nil, err
		}
		if _, exists := seen[value]; exists {
			return nil, fmt.Errorf("%w: %s不能重复", ErrMedicalInvalidInput, label)
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func lockScheduleDoctors(tx *gorm.DB, doctorIDs []string) error {
	unique := make(map[string]struct{}, len(doctorIDs))
	ids := make([]string, 0, len(doctorIDs))
	for _, doctorID := range doctorIDs {
		if _, exists := unique[doctorID]; exists {
			continue
		}
		unique[doctorID] = struct{}{}
		ids = append(ids, doctorID)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return fmt.Errorf("%w: 医生不能为空", ErrMedicalInvalidInput)
	}
	var doctors []models.MedDoctor
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("doctor_id IN ? AND status = 1 AND del_flag = 0", ids).
		Order("doctor_id asc").
		Find(&doctors).Error; err != nil {
		return err
	}
	if len(doctors) != len(ids) {
		return fmt.Errorf("%w: 医生不存在或已停用", ErrMedicalInvalidInput)
	}
	return nil
}

func validateScheduleDimension(tx *gorm.DB, doctorID, departmentID, registrationType string, scheduleDate *time.Time) (*models.MedDoctorDepartment, error) {
	if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
		return nil, err
	}
	if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(registrationType) == "" {
		return nil, fmt.Errorf("%w: 挂号类型不能为空", ErrMedicalInvalidInput)
	}

	var departmentCount int64
	if err := tx.Model(&models.MedDepartment{}).Where("department_id = ? AND status = 1 AND del_flag = 0", departmentID).Count(&departmentCount).Error; err != nil {
		return nil, err
	}
	if departmentCount == 0 {
		return nil, fmt.Errorf("%w: 临床科室不存在或已停用", ErrMedicalInvalidInput)
	}

	var relation models.MedDoctorDepartment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("doctor_id = ? AND department_id = ? AND status = 1 AND appointment_enabled = 1 AND del_flag = 0", doctorID, departmentID).
		First(&relation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: 医生未绑定该出诊科室或该科室未开放预约", ErrMedicalInvalidInput)
		}
		return nil, err
	}
	if scheduleDate != nil {
		if relation.ValidFrom != nil && medicalDateBefore(*scheduleDate, *relation.ValidFrom) {
			return nil, fmt.Errorf("%w: 出诊日期早于医生科室关系生效日期", ErrMedicalInvalidInput)
		}
		if relation.ValidTo != nil && medicalDateAfter(*scheduleDate, *relation.ValidTo) {
			return nil, fmt.Errorf("%w: 出诊日期晚于医生科室关系失效日期", ErrMedicalInvalidInput)
		}
	}

	var dictCount int64
	if err := tx.Table("sys_dict AS item").
		Joins("JOIN sys_dict AS root ON root.id = item.pid AND root.type = ? AND root.del_flag = 0 AND root.status = 1", registrationTypeDictType).
		Where("item.type = ? AND item.value = ? AND item.del_flag = 0 AND item.status = 1", registrationTypeDictType, registrationType).
		Count(&dictCount).Error; err != nil {
		return nil, err
	}
	if dictCount == 0 {
		return nil, fmt.Errorf("%w: 挂号类型字典值不存在或已停用", ErrMedicalInvalidInput)
	}
	return &relation, nil
}

func resolveScheduleFeeRule(tx *gorm.DB, doctorID, departmentID, registrationType string, scheduleDate time.Time) (*models.MedRegistrationFeeRule, error) {
	var rules []models.MedRegistrationFeeRule
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("doctor_id = ? AND department_id = ? AND registration_type = ? AND effective_date <= ? AND (expiry_date IS NULL OR expiry_date >= ?) AND del_flag = 0", doctorID, departmentID, registrationType, scheduleDate, scheduleDate).
		Order("effective_date desc, version desc").
		Limit(2).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("%w: 出诊日期没有可用的挂号费规则", ErrMedicalConflict)
	}
	if len(rules) > 1 {
		return nil, fmt.Errorf("%w: 出诊日期匹配到重叠的挂号费规则", ErrMedicalConflict)
	}
	return &rules[0], nil
}

func ensureScheduleTimeAvailable(tx *gorm.DB, doctorID string, scheduleDate time.Time, startTime, endTime, excludeScheduleID string) error {
	query := tx.Model(&models.MedSchedule{}).
		Where("doctor_id = ? AND schedule_date = ? AND status IN ? AND del_flag = 0", doctorID, scheduleDate, []int{models.MedScheduleStatusDraft, models.MedScheduleStatusPublished}).
		Where("start_time < ? AND end_time > ?", endTime, startTime)
	if excludeScheduleID != "" {
		query = query.Where("schedule_id != ?", excludeScheduleID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 医生在该日期和时间段已有排班", ErrMedicalConflict)
	}
	return nil
}

func scheduleGenerationRequestHash(templateIDs []string, startDate, endDate time.Time) (string, string, error) {
	templateJSON, err := json.Marshal(templateIDs)
	if err != nil {
		return "", "", err
	}
	payload := string(templateJSON) + "|" + startDate.Format("2006-01-02") + "|" + endDate.Format("2006-01-02")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:]), string(templateJSON), nil
}

func validateScheduleGenerationRange(startValue, endValue string) (time.Time, time.Time, error) {
	startDate, err := parseRequiredMedicalDate(startValue, "生成开始日期")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := parseRequiredMedicalDate(endValue, "生成结束日期")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if medicalDateBefore(startDate, medicalToday()) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: 生成开始日期不能早于今天", ErrMedicalInvalidInput)
	}
	if medicalDateBefore(endDate, startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: 生成结束日期不能早于开始日期", ErrMedicalInvalidInput)
	}
	if medicalDateAfter(endDate, medicalToday().AddDate(0, 0, maxScheduleGenerationDays-1)) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: 只能生成今天起未来%d天内的排班", ErrMedicalInvalidInput, maxScheduleGenerationDays)
	}
	return startDate, endDate, nil
}

func trimScheduleTime(value string) string {
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

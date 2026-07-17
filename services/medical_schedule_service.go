package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

type MedicalScheduleService struct{}

func NewMedicalScheduleService() *MedicalScheduleService {
	return &MedicalScheduleService{}
}

type scheduleTemplateListRow struct {
	models.MedScheduleTemplate
	DoctorNo       string `gorm:"column:doctor_no"`
	DoctorName     string `gorm:"column:doctor_name"`
	DepartmentCode string `gorm:"column:department_code"`
	DepartmentName string `gorm:"column:department_name"`
}

type scheduleListRow struct {
	models.MedSchedule
	DoctorNo       string `gorm:"column:doctor_no"`
	DoctorName     string `gorm:"column:doctor_name"`
	DepartmentCode string `gorm:"column:department_code"`
	DepartmentName string `gorm:"column:department_name"`
}

func (s *MedicalScheduleService) GetScheduleTemplateList(req models.ScheduleTemplateListRequest) (*utils.PageResult, error) {
	query := database.DB.Table("med_schedule_template AS template").
		Select("template.*, doctor.doctor_no, doctor.name AS doctor_name, department.department_code, department.department_name").
		Joins("JOIN med_doctor AS doctor ON doctor.doctor_id = template.doctor_id AND doctor.del_flag = 0").
		Joins("JOIN med_department AS department ON department.department_id = template.department_id AND department.del_flag = 0").
		Where("template.del_flag = 0")
	if req.DoctorID != "" {
		if err := validateMedicalUUID(req.DoctorID, "医生ID"); err != nil {
			return nil, err
		}
		query = query.Where("template.doctor_id = ?", req.DoctorID)
	}
	if req.DepartmentID != "" {
		if err := validateMedicalUUID(req.DepartmentID, "临床科室ID"); err != nil {
			return nil, err
		}
		query = query.Where("template.department_id = ?", req.DepartmentID)
	}
	if req.Weekday != nil {
		if *req.Weekday < 1 || *req.Weekday > 7 {
			return nil, fmt.Errorf("%w: 星期必须在1到7之间", ErrMedicalInvalidInput)
		}
		query = query.Where("template.weekday = ?", *req.Weekday)
	}
	if req.Status != nil {
		if err := validateMedicalStatus(*req.Status); err != nil {
			return nil, err
		}
		query = query.Where("template.status = ?", *req.Status)
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"doctorName":     "doctor.name",
		"departmentName": "department.department_name",
		"weekday":        "template.weekday",
		"startTime":      "template.start_time",
		"effectiveDate":  "template.effective_date",
		"status":         "template.status",
		"createDate":     "template.create_date",
	})
	if order == "" {
		order = "template.weekday asc, template.start_time asc, template.create_date desc, template.template_id asc"
	} else {
		order += ", template.template_id asc"
	}
	query = query.Order(order)
	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var rows []scheduleTemplateListRow
	pageResult, err := utils.Paginate(query, req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	items := make([]*models.ScheduleTemplateResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, scheduleTemplateToResponse(row))
	}
	pageResult.Items = items
	return pageResult, nil
}

func (s *MedicalScheduleService) CreateScheduleTemplate(req models.SaveScheduleTemplateRequest, operatorID string) error {
	prepared, err := prepareScheduleTemplateRequest(req)
	if err != nil {
		return err
	}
	slotQuotaConfig, err := marshalScheduleSlotQuotaConfig(prepared.slotQuotaConfig)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockScheduleDoctors(tx, []string{prepared.doctorID}); err != nil {
			return err
		}
		if _, err := validateScheduleDimension(tx, prepared.doctorID, prepared.departmentID, prepared.registrationType, nil); err != nil {
			return err
		}
		if prepared.status == 1 {
			if err := ensureScheduleTemplateAvailable(tx, prepared, ""); err != nil {
				return err
			}
		}
		now := time.Now()
		template := models.MedScheduleTemplate{
			TemplateID:       utils.GenerateUUID(),
			TemplateName:     prepared.templateName,
			DoctorID:         prepared.doctorID,
			DepartmentID:     prepared.departmentID,
			RegistrationType: prepared.registrationType,
			Weekday:          prepared.weekday,
			StartTime:        prepared.startTime,
			EndTime:          prepared.endTime,
			DefaultSlotQuota: prepared.defaultSlotQuota,
			SlotQuotaConfig:  slotQuotaConfig,
			TotalQuota:       prepared.totalQuota,
			EffectiveDate:    prepared.effectiveDate,
			ExpiryDate:       prepared.expiryDate,
			Status:           prepared.status,
			Remark:           prepared.remark,
			CreatorID:        optionalOperatorID(operatorID),
			UpdaterID:        optionalOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		}
		return tx.Create(&template).Error
	})
}

func (s *MedicalScheduleService) UpdateScheduleTemplate(templateID string, req models.SaveScheduleTemplateRequest, operatorID string) error {
	if err := validateMedicalUUID(templateID, "排班模板ID"); err != nil {
		return err
	}
	prepared, err := prepareScheduleTemplateRequest(req)
	if err != nil {
		return err
	}
	slotQuotaConfig, err := marshalScheduleSlotQuotaConfig(prepared.slotQuotaConfig)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var initial models.MedScheduleTemplate
		if err := tx.Where("template_id = ? AND del_flag = 0", templateID).First(&initial).Error; err != nil {
			return scheduleRecordError(err, "排班模板不存在")
		}
		lockedDoctorIDs := []string{initial.DoctorID, prepared.doctorID}
		if err := lockScheduleDoctors(tx, lockedDoctorIDs); err != nil {
			return err
		}
		var current models.MedScheduleTemplate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("template_id = ? AND del_flag = 0", templateID).First(&current).Error; err != nil {
			return scheduleRecordError(err, "排班模板不存在")
		}
		if current.DoctorID != initial.DoctorID {
			return fmt.Errorf("%w: 排班模板已被并发修改，请重试", ErrMedicalConflict)
		}
		if _, err := validateScheduleDimension(tx, prepared.doctorID, prepared.departmentID, prepared.registrationType, nil); err != nil {
			return err
		}
		if prepared.status == 1 {
			if err := ensureScheduleTemplateAvailable(tx, prepared, templateID); err != nil {
				return err
			}
		}
		return tx.Model(&models.MedScheduleTemplate{}).
			Where("template_id = ? AND del_flag = 0", templateID).
			Updates(map[string]interface{}{
				"template_name":      prepared.templateName,
				"doctor_id":          prepared.doctorID,
				"department_id":      prepared.departmentID,
				"registration_type":  prepared.registrationType,
				"weekday":            prepared.weekday,
				"start_time":         prepared.startTime,
				"end_time":           prepared.endTime,
				"default_slot_quota": prepared.defaultSlotQuota,
				"slot_quota_config":  slotQuotaConfig,
				"total_quota":        prepared.totalQuota,
				"effective_date":     prepared.effectiveDate,
				"expiry_date":        prepared.expiryDate,
				"status":             prepared.status,
				"remark":             prepared.remark,
				"updater_id":         optionalOperatorID(operatorID),
				"update_date":        time.Now(),
			}).Error
	})
}

func (s *MedicalScheduleService) UpdateScheduleTemplateStatus(templateID string, status int, operatorID string) error {
	if err := validateMedicalUUID(templateID, "排班模板ID"); err != nil {
		return err
	}
	if err := validateMedicalStatus(status); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var initial models.MedScheduleTemplate
		if err := tx.Where("template_id = ? AND del_flag = 0", templateID).First(&initial).Error; err != nil {
			return scheduleRecordError(err, "排班模板不存在")
		}
		if err := lockScheduleDoctors(tx, []string{initial.DoctorID}); err != nil {
			return err
		}
		var current models.MedScheduleTemplate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("template_id = ? AND del_flag = 0", templateID).First(&current).Error; err != nil {
			return scheduleRecordError(err, "排班模板不存在")
		}
		if current.DoctorID != initial.DoctorID {
			return fmt.Errorf("%w: 排班模板已被并发修改，请重试", ErrMedicalConflict)
		}
		if status == 1 {
			prepared, err := templateToPrepared(current)
			if err != nil {
				return err
			}
			if _, err := validateScheduleDimension(tx, current.DoctorID, current.DepartmentID, current.RegistrationType, nil); err != nil {
				return err
			}
			if err := ensureScheduleTemplateAvailable(tx, prepared, templateID); err != nil {
				return err
			}
		}
		return tx.Model(&models.MedScheduleTemplate{}).
			Where("template_id = ? AND del_flag = 0", templateID).
			Updates(map[string]interface{}{
				"status":      status,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": time.Now(),
			}).Error
	})
}

func (s *MedicalScheduleService) GetScheduleList(req models.ScheduleListRequest) (*utils.PageResult, error) {
	query := database.DB.Table("med_schedule AS schedule").
		Select("schedule.*, doctor.doctor_no, doctor.name AS doctor_name, department.department_code, department.department_name").
		Joins("JOIN med_doctor AS doctor ON doctor.doctor_id = schedule.doctor_id AND doctor.del_flag = 0").
		Joins("JOIN med_department AS department ON department.department_id = schedule.department_id AND department.del_flag = 0").
		Where("schedule.del_flag = 0")
	if req.DoctorID != "" {
		if err := validateMedicalUUID(req.DoctorID, "医生ID"); err != nil {
			return nil, err
		}
		query = query.Where("schedule.doctor_id = ?", req.DoctorID)
	}
	if req.DepartmentID != "" {
		if err := validateMedicalUUID(req.DepartmentID, "临床科室ID"); err != nil {
			return nil, err
		}
		query = query.Where("schedule.department_id = ?", req.DepartmentID)
	}
	if req.RegistrationType != "" {
		query = query.Where("schedule.registration_type = ?", req.RegistrationType)
	}
	var startDate *time.Time
	if req.StartDate != "" {
		parsed, err := parseRequiredMedicalDate(req.StartDate, "开始日期")
		if err != nil {
			return nil, err
		}
		startDate = &parsed
		query = query.Where("schedule.schedule_date >= ?", parsed)
	}
	var endDate *time.Time
	if req.EndDate != "" {
		parsed, err := parseRequiredMedicalDate(req.EndDate, "结束日期")
		if err != nil {
			return nil, err
		}
		endDate = &parsed
		query = query.Where("schedule.schedule_date <= ?", parsed)
	}
	if startDate != nil && endDate != nil && medicalDateBefore(*endDate, *startDate) {
		return nil, fmt.Errorf("%w: 结束日期不能早于开始日期", ErrMedicalInvalidInput)
	}
	if req.Status != nil {
		if err := validateScheduleStatus(*req.Status); err != nil {
			return nil, err
		}
		query = query.Where("schedule.status = ?", *req.Status)
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"doctorName":     "doctor.name",
		"departmentName": "department.department_name",
		"scheduleDate":   "schedule.schedule_date",
		"startTime":      "schedule.start_time",
		"feeAmount":      "schedule.fee_amount",
		"totalQuota":     "schedule.total_quota",
		"status":         "schedule.status",
		"createDate":     "schedule.create_date",
	})
	if order == "" {
		order = "schedule.schedule_date asc, schedule.start_time asc, schedule.create_date desc, schedule.schedule_id asc"
	} else {
		order += ", schedule.schedule_id asc"
	}
	query = query.Order(order)
	pageSize := req.PageSize
	if pageSize > 200 {
		pageSize = 200
	}
	var rows []scheduleListRow
	pageResult, err := utils.Paginate(query, req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	scheduleIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		scheduleIDs = append(scheduleIDs, row.ScheduleID)
	}
	slotsBySchedule := make(map[string][]models.MedScheduleSlot, len(scheduleIDs))
	if len(scheduleIDs) > 0 {
		var slots []models.MedScheduleSlot
		if err := database.DB.Where("schedule_id IN ? AND del_flag = 0", scheduleIDs).
			Order("schedule_id asc, start_time asc").Find(&slots).Error; err != nil {
			return nil, err
		}
		for _, slot := range slots {
			slotsBySchedule[slot.ScheduleID] = append(slotsBySchedule[slot.ScheduleID], slot)
		}
	}
	items := make([]*models.ScheduleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, scheduleToResponse(row, slotsBySchedule[row.ScheduleID]))
	}
	pageResult.Items = items
	return pageResult, nil
}

func (s *MedicalScheduleService) CreateSchedule(req models.SaveScheduleRequest, operatorID string) error {
	prepared, err := prepareScheduleRequest(req)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockScheduleDoctors(tx, []string{prepared.doctorID}); err != nil {
			return err
		}
		if _, err := validateScheduleDimension(tx, prepared.doctorID, prepared.departmentID, prepared.registrationType, &prepared.scheduleDate); err != nil {
			return err
		}
		if err := ensureScheduleTimeAvailable(tx, prepared.doctorID, prepared.scheduleDate, prepared.startTime, prepared.endTime, ""); err != nil {
			return err
		}
		now := time.Now()
		scheduleID := utils.GenerateUUID()
		slots, totalQuota, _, err := buildScheduleSlotDrafts(scheduleID, prepared.startTime, prepared.endTime, prepared.defaultSlotQuota, prepared.slotQuotaConfig, operatorID)
		if err != nil {
			return err
		}
		schedule := models.MedSchedule{
			ScheduleID:       scheduleID,
			DoctorID:         prepared.doctorID,
			DepartmentID:     prepared.departmentID,
			RegistrationType: prepared.registrationType,
			ScheduleDate:     prepared.scheduleDate,
			StartTime:        prepared.startTime,
			EndTime:          prepared.endTime,
			DefaultSlotQuota: prepared.defaultSlotQuota,
			TotalQuota:       totalQuota,
			Status:           models.MedScheduleStatusDraft,
			Remark:           prepared.remark,
			CreatorID:        optionalOperatorID(operatorID),
			UpdaterID:        optionalOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		}
		if err := tx.Create(&schedule).Error; err != nil {
			return err
		}
		return tx.CreateInBatches(&slots, 100).Error
	})
}

func (s *MedicalScheduleService) UpdateSchedule(scheduleID string, req models.SaveScheduleRequest, operatorID string) error {
	if err := validateMedicalUUID(scheduleID, "排班ID"); err != nil {
		return err
	}
	prepared, err := prepareScheduleRequest(req)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var initial models.MedSchedule
		if err := tx.Where("schedule_id = ? AND del_flag = 0", scheduleID).First(&initial).Error; err != nil {
			return scheduleRecordError(err, "排班不存在")
		}
		if err := lockScheduleDoctors(tx, []string{initial.DoctorID, prepared.doctorID}); err != nil {
			return err
		}
		var current models.MedSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("schedule_id = ? AND del_flag = 0", scheduleID).First(&current).Error; err != nil {
			return scheduleRecordError(err, "排班不存在")
		}
		if current.DoctorID != initial.DoctorID {
			return fmt.Errorf("%w: 排班已被并发修改，请重试", ErrMedicalConflict)
		}
		if current.Status != models.MedScheduleStatusDraft {
			return fmt.Errorf("%w: 只有草稿排班可以编辑", ErrMedicalConflict)
		}
		if _, err := validateScheduleDimension(tx, prepared.doctorID, prepared.departmentID, prepared.registrationType, &prepared.scheduleDate); err != nil {
			return err
		}
		if err := ensureScheduleTimeAvailable(tx, prepared.doctorID, prepared.scheduleDate, prepared.startTime, prepared.endTime, scheduleID); err != nil {
			return err
		}
		slots, totalQuota, _, err := buildScheduleSlotDrafts(scheduleID, prepared.startTime, prepared.endTime, prepared.defaultSlotQuota, prepared.slotQuotaConfig, operatorID)
		if err != nil {
			return err
		}
		// 草稿档位尚未进入预约链路，编辑时物理替换，避免布尔删除标记与
		// (schedule_id, start_time, del_flag) 唯一键在多次编辑时冲突。
		if err := tx.Where("schedule_id = ? AND del_flag = 0", scheduleID).
			Delete(&models.MedScheduleSlot{}).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&slots, 100).Error; err != nil {
			return err
		}
		return tx.Model(&models.MedSchedule{}).
			Where("schedule_id = ? AND del_flag = 0", scheduleID).
			Updates(map[string]interface{}{
				"doctor_id":          prepared.doctorID,
				"department_id":      prepared.departmentID,
				"registration_type":  prepared.registrationType,
				"schedule_date":      prepared.scheduleDate,
				"start_time":         prepared.startTime,
				"end_time":           prepared.endTime,
				"fee_rule_id":        nil,
				"fee_rule_version":   nil,
				"fee_amount":         nil,
				"default_slot_quota": prepared.defaultSlotQuota,
				"total_quota":        totalQuota,
				"remark":             prepared.remark,
				"updater_id":         optionalOperatorID(operatorID),
				"update_date":        time.Now(),
			}).Error
	})
}

func (s *MedicalScheduleService) PublishSchedules(req models.PublishSchedulesRequest, operatorID string) error {
	scheduleIDs, err := normalizeScheduleUUIDs(req.ScheduleIDs, "排班ID", 100)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return publishScheduleIDsTx(tx, scheduleIDs, operatorID)
	})
}

func publishScheduleIDsTx(tx *gorm.DB, scheduleIDs []string, operatorID string) error {
	var initial []models.MedSchedule
	if err := tx.Where("schedule_id IN ? AND del_flag = 0", scheduleIDs).Find(&initial).Error; err != nil {
		return err
	}
	if len(initial) != len(scheduleIDs) {
		return fmt.Errorf("%w: 部分排班不存在", ErrMedicalNotFound)
	}
	doctorIDs := make([]string, 0, len(initial))
	for _, schedule := range initial {
		doctorIDs = append(doctorIDs, schedule.DoctorID)
	}
	if err := lockScheduleDoctors(tx, doctorIDs); err != nil {
		return err
	}
	var schedules []models.MedSchedule
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("schedule_id IN ? AND del_flag = 0", scheduleIDs).
		Order("doctor_id asc, department_id asc, registration_type asc, schedule_date asc, start_time asc, schedule_id asc").
		Find(&schedules).Error; err != nil {
		return err
	}
	if len(schedules) != len(scheduleIDs) {
		return fmt.Errorf("%w: 部分排班不存在", ErrMedicalNotFound)
	}
	var slots []models.MedScheduleSlot
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("schedule_id IN ? AND del_flag = 0", scheduleIDs).
		Order("schedule_id asc, start_time asc").Find(&slots).Error; err != nil {
		return err
	}
	slotsBySchedule := make(map[string][]models.MedScheduleSlot, len(scheduleIDs))
	for _, slot := range slots {
		slotsBySchedule[slot.ScheduleID] = append(slotsBySchedule[slot.ScheduleID], slot)
	}
	now := time.Now()
	for _, schedule := range schedules {
		if schedule.Status != models.MedScheduleStatusDraft {
			return fmt.Errorf("%w: 只有草稿排班可以发布", ErrMedicalConflict)
		}
		startAt, err := scheduleDateTime(schedule.ScheduleDate, schedule.StartTime)
		if err != nil {
			return err
		}
		if !startAt.After(now.In(medicalBusinessLocation)) {
			return fmt.Errorf("%w: 已开始或已过期的排班不能发布", ErrMedicalConflict)
		}
		scheduleSlots := slotsBySchedule[schedule.ScheduleID]
		if len(scheduleSlots) == 0 {
			return fmt.Errorf("%w: 排班没有号源档位", ErrMedicalConflict)
		}
		totalQuota := 0
		bookedQuota := 0
		for _, slot := range scheduleSlots {
			if slot.Quota < slot.BookedQuota {
				return fmt.Errorf("%w: 档位容量不能小于已预约数量", ErrMedicalConflict)
			}
			totalQuota += slot.Quota
			bookedQuota += slot.BookedQuota
		}
		if totalQuota == 0 {
			return fmt.Errorf("%w: 整段排班至少需要一个可预约号源", ErrMedicalConflict)
		}
		if _, err := validateScheduleDimension(tx, schedule.DoctorID, schedule.DepartmentID, schedule.RegistrationType, &schedule.ScheduleDate); err != nil {
			return err
		}
		feeRule, err := resolveScheduleFeeRule(tx, schedule.DoctorID, schedule.DepartmentID, schedule.RegistrationType, schedule.ScheduleDate)
		if err != nil {
			return err
		}
		feeRuleID := feeRule.FeeRuleID
		feeRuleVersion := feeRule.Version
		feeAmount := feeRule.FeeAmount
		if err := tx.Model(&models.MedSchedule{}).
			Where("schedule_id = ? AND status = ? AND del_flag = 0", schedule.ScheduleID, models.MedScheduleStatusDraft).
			Updates(map[string]interface{}{
				"fee_rule_id":      &feeRuleID,
				"fee_rule_version": &feeRuleVersion,
				"fee_amount":       &feeAmount,
				"total_quota":      totalQuota,
				"booked_quota":     bookedQuota,
				"status":           models.MedScheduleStatusPublished,
				"published_at":     now,
				"updater_id":       optionalOperatorID(operatorID),
				"update_date":      now,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *MedicalScheduleService) StopSchedule(scheduleID string, req models.StopScheduleRequest, operatorID string) error {
	if err := validateMedicalUUID(scheduleID, "排班ID"); err != nil {
		return err
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" || len([]rune(reason)) > 512 {
		return fmt.Errorf("%w: 停诊原因不能为空且最多512个字符", ErrMedicalInvalidInput)
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var schedule models.MedSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("schedule_id = ? AND del_flag = 0", scheduleID).First(&schedule).Error; err != nil {
			return scheduleRecordError(err, "排班不存在")
		}
		if schedule.Status != models.MedScheduleStatusPublished {
			return fmt.Errorf("%w: 只有已发布排班可以停诊", ErrMedicalConflict)
		}
		now := time.Now()
		return tx.Model(&models.MedSchedule{}).
			Where("schedule_id = ? AND del_flag = 0", scheduleID).
			Updates(map[string]interface{}{
				"status":      models.MedScheduleStatusStopped,
				"stop_reason": reason,
				"stopped_at":  now,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error
	})
}

func (s *MedicalScheduleService) DeleteDraftSchedules(scheduleIDs []string, operatorID string) error {
	normalizedIDs, err := normalizeScheduleUUIDs(scheduleIDs, "排班ID", 100)
	if err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var schedules []models.MedSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("schedule_id IN ? AND del_flag = 0", normalizedIDs).
			Order("schedule_id asc").Find(&schedules).Error; err != nil {
			return err
		}
		if len(schedules) != len(normalizedIDs) {
			return fmt.Errorf("%w: 部分排班不存在", ErrMedicalNotFound)
		}
		for _, schedule := range schedules {
			if schedule.Status != models.MedScheduleStatusDraft {
				return fmt.Errorf("%w: 只有草稿排班可以删除", ErrMedicalConflict)
			}
		}
		now := time.Now()
		if err := tx.Model(&models.MedScheduleSlot{}).Where("schedule_id IN ? AND del_flag = 0", normalizedIDs).
			Updates(map[string]interface{}{"del_flag": 1, "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error; err != nil {
			return err
		}
		return tx.Model(&models.MedSchedule{}).Where("schedule_id IN ? AND status = ? AND del_flag = 0", normalizedIDs, models.MedScheduleStatusDraft).
			Updates(map[string]interface{}{"del_flag": 1, "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error
	})
}

func (s *MedicalScheduleService) FinishSchedule(scheduleID string, operatorID string) error {
	if err := validateMedicalUUID(scheduleID, "排班ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var schedule models.MedSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("schedule_id = ? AND del_flag = 0", scheduleID).First(&schedule).Error; err != nil {
			return scheduleRecordError(err, "排班不存在")
		}
		if schedule.Status != models.MedScheduleStatusPublished {
			return fmt.Errorf("%w: 只有已发布排班可以结束", ErrMedicalConflict)
		}
		endAt, err := scheduleDateTime(schedule.ScheduleDate, schedule.EndTime)
		if err != nil {
			return err
		}
		now := time.Now()
		if endAt.After(now.In(medicalBusinessLocation)) {
			return fmt.Errorf("%w: 排班尚未结束", ErrMedicalConflict)
		}
		return tx.Model(&models.MedSchedule{}).
			Where("schedule_id = ? AND del_flag = 0", scheduleID).
			Updates(map[string]interface{}{
				"status":      models.MedScheduleStatusFinished,
				"finished_at": now,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error
	})
}

func ensureScheduleTemplateAvailable(tx *gorm.DB, prepared *preparedScheduleTemplate, excludeTemplateID string) error {
	query := tx.Model(&models.MedScheduleTemplate{}).
		Where("doctor_id = ? AND weekday = ? AND status = 1 AND del_flag = 0", prepared.doctorID, prepared.weekday).
		Where("start_time < ? AND end_time > ?", prepared.endTime, prepared.startTime).
		Where("expiry_date IS NULL OR expiry_date >= ?", prepared.effectiveDate)
	if prepared.expiryDate != nil {
		query = query.Where("effective_date <= ?", *prepared.expiryDate)
	}
	if excludeTemplateID != "" {
		query = query.Where("template_id != ?", excludeTemplateID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 医生存在有效期和时间段重叠的周期模板", ErrMedicalConflict)
	}
	return nil
}

func templateToPrepared(template models.MedScheduleTemplate) (*preparedScheduleTemplate, error) {
	config, err := unmarshalScheduleSlotQuotaConfig(template.SlotQuotaConfig)
	if err != nil {
		return nil, err
	}
	return &preparedScheduleTemplate{
		templateName:     template.TemplateName,
		doctorID:         template.DoctorID,
		departmentID:     template.DepartmentID,
		registrationType: template.RegistrationType,
		weekday:          template.Weekday,
		startTime:        template.StartTime,
		endTime:          template.EndTime,
		defaultSlotQuota: template.DefaultSlotQuota,
		slotQuotaConfig:  config,
		totalQuota:       template.TotalQuota,
		effectiveDate:    template.EffectiveDate,
		expiryDate:       template.ExpiryDate,
		status:           template.Status,
		remark:           template.Remark,
	}, nil
}

func scheduleTemplateToResponse(row scheduleTemplateListRow) *models.ScheduleTemplateResponse {
	config, _ := unmarshalScheduleSlotQuotaConfig(row.SlotQuotaConfig)
	return &models.ScheduleTemplateResponse{
		TemplateID:       row.TemplateID,
		TemplateName:     row.TemplateName,
		DoctorID:         row.DoctorID,
		DoctorNo:         row.DoctorNo,
		DoctorName:       row.DoctorName,
		DepartmentID:     row.DepartmentID,
		DepartmentCode:   row.DepartmentCode,
		DepartmentName:   row.DepartmentName,
		RegistrationType: row.RegistrationType,
		Weekday:          row.Weekday,
		StartTime:        trimScheduleTime(row.StartTime),
		EndTime:          trimScheduleTime(row.EndTime),
		DefaultSlotQuota: row.DefaultSlotQuota,
		SlotQuotaConfig:  config,
		TotalQuota:       row.TotalQuota,
		EffectiveDate:    row.EffectiveDate.Format("2006-01-02"),
		ExpiryDate:       timeToDateStringPtr(row.ExpiryDate),
		Status:           row.Status,
		Remark:           row.Remark,
		CreateDate:       models.TimeToStringPtr(row.CreateDate),
		UpdateDate:       models.TimeToStringPtr(row.UpdateDate),
	}
}

func scheduleToResponse(row scheduleListRow, slots []models.MedScheduleSlot) *models.ScheduleResponse {
	remainingQuota := row.TotalQuota - row.BookedQuota
	if remainingQuota < 0 {
		remainingQuota = 0
	}
	feeSnapshotStatus := "pending"
	if row.FeeRuleID != nil && row.FeeRuleVersion != nil && row.FeeAmount != nil {
		feeSnapshotStatus = "fixed"
	}
	slotResponses := make([]models.ScheduleSlotResponse, 0, len(slots))
	now := time.Now().In(medicalBusinessLocation)
	for _, slot := range slots {
		remaining := slot.Quota - slot.BookedQuota
		if remaining < 0 {
			remaining = 0
		}
		status := "draft"
		canBook := false
		switch row.Status {
		case models.MedScheduleStatusStopped:
			status = "stopped"
		case models.MedScheduleStatusFinished:
			status = "finished"
		case models.MedScheduleStatusPublished:
			startAt, err := scheduleDateTime(row.ScheduleDate, slot.StartTime)
			switch {
			case err != nil || !now.Before(startAt):
				status = "closed"
			case remaining == 0:
				status = "full"
			default:
				status = "available"
				canBook = true
			}
		}
		slotResponses = append(slotResponses, models.ScheduleSlotResponse{
			SlotID: slot.SlotID, StartTime: trimScheduleTime(slot.StartTime), EndTime: trimScheduleTime(slot.EndTime),
			Quota: slot.Quota, BookedQuota: slot.BookedQuota, RemainingQuota: remaining, BookingStatus: status, CanBook: canBook,
		})
	}
	return &models.ScheduleResponse{
		ScheduleID:        row.ScheduleID,
		TemplateID:        row.TemplateID,
		GenerationBatchID: row.GenerationBatchID,
		DoctorID:          row.DoctorID,
		DoctorNo:          row.DoctorNo,
		DoctorName:        row.DoctorName,
		DepartmentID:      row.DepartmentID,
		DepartmentCode:    row.DepartmentCode,
		DepartmentName:    row.DepartmentName,
		RegistrationType:  row.RegistrationType,
		ScheduleDate:      row.ScheduleDate.Format("2006-01-02"),
		StartTime:         trimScheduleTime(row.StartTime),
		EndTime:           trimScheduleTime(row.EndTime),
		FeeRuleID:         row.FeeRuleID,
		FeeRuleVersion:    row.FeeRuleVersion,
		FeeAmount:         row.FeeAmount,
		FeeSnapshotStatus: feeSnapshotStatus,
		DefaultSlotQuota:  row.DefaultSlotQuota,
		TotalQuota:        row.TotalQuota,
		BookedQuota:       row.BookedQuota,
		RemainingQuota:    remainingQuota,
		Status:            row.Status,
		StopReason:        row.StopReason,
		PublishedAt:       models.TimeToStringPtr(row.PublishedAt),
		StoppedAt:         models.TimeToStringPtr(row.StoppedAt),
		FinishedAt:        models.TimeToStringPtr(row.FinishedAt),
		Remark:            row.Remark,
		CreateDate:        models.TimeToStringPtr(row.CreateDate),
		UpdateDate:        models.TimeToStringPtr(row.UpdateDate),
		Slots:             slotResponses,
	}
}

func scheduleRecordError(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("%w: %s", ErrMedicalNotFound, message)
	}
	return err
}

func scheduleIDsFromRows(schedules []models.MedSchedule) []string {
	ids := make([]string, 0, len(schedules))
	for _, schedule := range schedules {
		ids = append(ids, schedule.ScheduleID)
	}
	sort.Strings(ids)
	return ids
}

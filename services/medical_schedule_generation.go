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

func (s *MedicalScheduleService) GenerateSchedules(req models.GenerateSchedulesRequest, operatorID string) (*models.GenerateSchedulesResponse, error) {
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" || len([]rune(idempotencyKey)) > 64 {
		return nil, fmt.Errorf("%w: 幂等键不能为空且最多64个字符", ErrMedicalInvalidInput)
	}
	templateIDs, err := normalizeScheduleUUIDs(req.TemplateIDs, "排班模板ID", 100)
	if err != nil {
		return nil, err
	}
	startDate, endDate, err := validateScheduleGenerationRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	requestHash, templateJSON, err := scheduleGenerationRequestHash(templateIDs, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var response *models.GenerateSchedulesResponse
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		candidateBatch := models.MedScheduleGenerationBatch{
			BatchID:        utils.GenerateUUID(),
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
			TemplateIDs:    templateJSON,
			StartDate:      startDate,
			EndDate:        endDate,
			Status:         models.MedScheduleBatchStatusProcessing,
			CreatorID:      optionalOperatorID(operatorID),
			CreateDate:     &now,
			UpdateDate:     &now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "idempotency_key"}},
			DoNothing: true,
		}).Create(&candidateBatch).Error; err != nil {
			return err
		}

		var batch models.MedScheduleGenerationBatch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("idempotency_key = ?", idempotencyKey).First(&batch).Error; err != nil {
			return err
		}
		if batch.RequestHash != requestHash {
			return fmt.Errorf("%w: 相同幂等键不能用于不同的生成参数", ErrMedicalConflict)
		}
		if batch.BatchID != candidateBatch.BatchID {
			if batch.Status != models.MedScheduleBatchStatusCompleted {
				return fmt.Errorf("%w: 该生成请求正在处理中", ErrMedicalConflict)
			}
			result, err := loadScheduleGenerationResult(tx, batch, true)
			if err != nil {
				return err
			}
			response = result
			return nil
		}

		templates, relations, existingSchedules, err := prepareScheduleGenerationData(tx, templateIDs, startDate, endDate)
		if err != nil {
			return err
		}
		createdSchedules, createdSlots, skippedCount, err := buildGeneratedSchedules(
			templates,
			relations,
			existingSchedules,
			batch.BatchID,
			startDate,
			endDate,
			operatorID,
		)
		if err != nil {
			return err
		}
		if len(createdSchedules) > 0 {
			if err := tx.CreateInBatches(&createdSchedules, 200).Error; err != nil {
				return err
			}
			if err := tx.CreateInBatches(&createdSlots, 500).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.MedScheduleGenerationBatch{}).
			Where("batch_id = ?", batch.BatchID).
			Updates(map[string]interface{}{
				"status":          models.MedScheduleBatchStatusCompleted,
				"generated_count": len(createdSchedules),
				"skipped_count":   skippedCount,
				"update_date":     time.Now(),
			}).Error; err != nil {
			return err
		}
		response = &models.GenerateSchedulesResponse{
			BatchID:        batch.BatchID,
			Idempotent:     false,
			GeneratedCount: len(createdSchedules),
			SkippedCount:   skippedCount,
			ScheduleIDs:    scheduleIDsFromRows(createdSchedules),
		}
		return nil
	})
	return response, err
}

func prepareScheduleGenerationData(tx *gorm.DB, templateIDs []string, startDate, endDate time.Time) ([]models.MedScheduleTemplate, map[string]models.MedDoctorDepartment, []models.MedSchedule, error) {
	var initialTemplates []models.MedScheduleTemplate
	if err := tx.Where("template_id IN ? AND del_flag = 0", templateIDs).Order("template_id asc").Find(&initialTemplates).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(initialTemplates) != len(templateIDs) {
		return nil, nil, nil, fmt.Errorf("%w: 部分排班模板不存在", ErrMedicalNotFound)
	}
	doctorIDs := make([]string, 0, len(initialTemplates))
	initialDoctorByTemplate := make(map[string]string, len(initialTemplates))
	for _, template := range initialTemplates {
		doctorIDs = append(doctorIDs, template.DoctorID)
		initialDoctorByTemplate[template.TemplateID] = template.DoctorID
	}
	if err := lockScheduleDoctors(tx, doctorIDs); err != nil {
		return nil, nil, nil, err
	}

	var templates []models.MedScheduleTemplate
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("template_id IN ? AND del_flag = 0", templateIDs).Order("template_id asc").Find(&templates).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(templates) != len(templateIDs) {
		return nil, nil, nil, fmt.Errorf("%w: 部分排班模板不存在", ErrMedicalNotFound)
	}
	for _, template := range templates {
		if template.DoctorID != initialDoctorByTemplate[template.TemplateID] {
			return nil, nil, nil, fmt.Errorf("%w: 排班模板已被并发修改，请重试", ErrMedicalConflict)
		}
		if template.Status != 1 {
			return nil, nil, nil, fmt.Errorf("%w: 停用的排班模板不能生成排班", ErrMedicalConflict)
		}
	}
	sort.Slice(templates, func(i, j int) bool {
		left := templates[i]
		right := templates[j]
		leftKey := left.DoctorID + "|" + left.DepartmentID + "|" + fmt.Sprintf("%d|%s|%s|%s", left.Weekday, left.StartTime, left.EndTime, left.TemplateID)
		rightKey := right.DoctorID + "|" + right.DepartmentID + "|" + fmt.Sprintf("%d|%s|%s|%s", right.Weekday, right.StartTime, right.EndTime, right.TemplateID)
		return leftKey < rightKey
	})

	relations, err := loadScheduleGenerationDimensions(tx, templates)
	if err != nil {
		return nil, nil, nil, err
	}

	uniqueDoctorIDs := uniqueScheduleStrings(doctorIDs)
	var existingSchedules []models.MedSchedule
	if err := tx.Where("doctor_id IN ? AND schedule_date BETWEEN ? AND ? AND del_flag = 0", uniqueDoctorIDs, startDate, endDate).
		Order("doctor_id asc, schedule_date asc, start_time asc, schedule_id asc").
		Find(&existingSchedules).Error; err != nil {
		return nil, nil, nil, err
	}
	return templates, relations, existingSchedules, nil
}

func loadScheduleGenerationDimensions(tx *gorm.DB, templates []models.MedScheduleTemplate) (map[string]models.MedDoctorDepartment, error) {
	departmentIDs := make([]string, 0, len(templates))
	registrationTypes := make([]string, 0, len(templates))
	pairKeys := make([]string, 0, len(templates))
	pairValues := make(map[string][]interface{}, len(templates))
	for _, template := range templates {
		departmentIDs = append(departmentIDs, template.DepartmentID)
		registrationTypes = append(registrationTypes, template.RegistrationType)
		pairKey := scheduleDoctorDepartmentKey(template.DoctorID, template.DepartmentID)
		if _, exists := pairValues[pairKey]; !exists {
			pairKeys = append(pairKeys, pairKey)
			pairValues[pairKey] = []interface{}{template.DoctorID, template.DepartmentID}
		}
	}
	departmentIDs = uniqueScheduleStrings(departmentIDs)
	registrationTypes = uniqueScheduleStrings(registrationTypes)
	sort.Strings(pairKeys)

	var departments []models.MedDepartment
	if err := tx.Where("department_id IN ? AND status = 1 AND del_flag = 0", departmentIDs).
		Order("department_id asc").
		Find(&departments).Error; err != nil {
		return nil, err
	}
	if len(departments) != len(departmentIDs) {
		return nil, fmt.Errorf("%w: 部分临床科室不存在或已停用", ErrMedicalInvalidInput)
	}

	pairs := make([][]interface{}, 0, len(pairKeys))
	for _, pairKey := range pairKeys {
		pairs = append(pairs, pairValues[pairKey])
	}
	var relationRows []models.MedDoctorDepartment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("(doctor_id, department_id) IN ? AND status = 1 AND appointment_enabled = 1 AND del_flag = 0", pairs).
		Order("doctor_id asc, department_id asc").
		Find(&relationRows).Error; err != nil {
		return nil, err
	}
	relationsByPair := make(map[string]models.MedDoctorDepartment, len(relationRows))
	for _, relation := range relationRows {
		relationsByPair[scheduleDoctorDepartmentKey(relation.DoctorID, relation.DepartmentID)] = relation
	}
	if len(relationsByPair) != len(pairKeys) {
		return nil, fmt.Errorf("%w: 部分医生未绑定出诊科室或科室未开放预约", ErrMedicalInvalidInput)
	}

	var dictValues []string
	if err := tx.Table("sys_dict AS item").
		Joins("JOIN sys_dict AS root ON root.id = item.pid AND root.type = ? AND root.del_flag = 0 AND root.status = 1", registrationTypeDictType).
		Where("item.type = ? AND item.value IN ? AND item.del_flag = 0 AND item.status = 1", registrationTypeDictType, registrationTypes).
		Distinct("item.value").
		Pluck("item.value", &dictValues).Error; err != nil {
		return nil, err
	}
	if len(dictValues) != len(registrationTypes) {
		return nil, fmt.Errorf("%w: 部分挂号类型字典值不存在或已停用", ErrMedicalInvalidInput)
	}

	result := make(map[string]models.MedDoctorDepartment, len(templates))
	for _, template := range templates {
		relation := relationsByPair[scheduleDoctorDepartmentKey(template.DoctorID, template.DepartmentID)]
		result[scheduleDimensionKey(template.DoctorID, template.DepartmentID, template.RegistrationType)] = relation
	}
	return result, nil
}

func buildGeneratedSchedules(templates []models.MedScheduleTemplate, relations map[string]models.MedDoctorDepartment, existingSchedules []models.MedSchedule, batchID string, startDate, endDate time.Time, operatorID string) ([]models.MedSchedule, []models.MedScheduleSlot, int, error) {
	schedulesByDoctorDate := make(map[string][]models.MedSchedule)
	for _, schedule := range existingSchedules {
		key := scheduleDoctorDateKey(schedule.DoctorID, schedule.ScheduleDate)
		schedulesByDoctorDate[key] = append(schedulesByDoctorDate[key], schedule)
	}
	created := make([]models.MedSchedule, 0)
	createdSlots := make([]models.MedScheduleSlot, 0)
	skippedCount := 0
	now := time.Now()
	for _, template := range templates {
		dimensionKey := scheduleDimensionKey(template.DoctorID, template.DepartmentID, template.RegistrationType)
		relation := relations[dimensionKey]
		for scheduleDate := startDate; !medicalDateAfter(scheduleDate, endDate); scheduleDate = scheduleDate.AddDate(0, 0, 1) {
			if isoWeekday(scheduleDate) != template.Weekday || medicalDateBefore(scheduleDate, template.EffectiveDate) || (template.ExpiryDate != nil && medicalDateAfter(scheduleDate, *template.ExpiryDate)) {
				continue
			}
			if relation.ValidFrom != nil && medicalDateBefore(scheduleDate, *relation.ValidFrom) {
				continue
			}
			if relation.ValidTo != nil && medicalDateAfter(scheduleDate, *relation.ValidTo) {
				continue
			}

			doctorDateKey := scheduleDoctorDateKey(template.DoctorID, scheduleDate)
			existing := schedulesByDoctorDate[doctorDateKey]
			alreadyGenerated := false
			for _, schedule := range existing {
				if schedule.TemplateID != nil && *schedule.TemplateID == template.TemplateID {
					alreadyGenerated = true
					break
				}
			}
			if alreadyGenerated {
				skippedCount++
				continue
			}
			for _, schedule := range existing {
				if (schedule.Status == models.MedScheduleStatusDraft || schedule.Status == models.MedScheduleStatusPublished) && scheduleTimesOverlap(schedule.StartTime, schedule.EndTime, template.StartTime, template.EndTime) {
					return nil, nil, 0, fmt.Errorf("%w: 医生在%s %s-%s已有排班", ErrMedicalConflict, scheduleDate.Format("2006-01-02"), trimScheduleTime(template.StartTime), trimScheduleTime(template.EndTime))
				}
			}
			config, err := unmarshalScheduleSlotQuotaConfig(template.SlotQuotaConfig)
			if err != nil {
				return nil, nil, 0, err
			}
			templateID := template.TemplateID
			batchIDValue := batchID
			scheduleID := utils.GenerateUUID()
			slots, totalQuota, _, err := buildScheduleSlotDrafts(scheduleID, template.StartTime, template.EndTime, template.DefaultSlotQuota, config, operatorID)
			if err != nil {
				return nil, nil, 0, err
			}
			schedule := models.MedSchedule{
				ScheduleID:        scheduleID,
				TemplateID:        &templateID,
				GenerationBatchID: &batchIDValue,
				DoctorID:          template.DoctorID,
				DepartmentID:      template.DepartmentID,
				RegistrationType:  template.RegistrationType,
				ScheduleDate:      scheduleDate,
				StartTime:         template.StartTime,
				EndTime:           template.EndTime,
				DefaultSlotQuota:  template.DefaultSlotQuota,
				TotalQuota:        totalQuota,
				Status:            models.MedScheduleStatusDraft,
				Remark:            template.Remark,
				CreatorID:         optionalOperatorID(operatorID),
				UpdaterID:         optionalOperatorID(operatorID),
				CreateDate:        &now,
				UpdateDate:        &now,
			}
			created = append(created, schedule)
			createdSlots = append(createdSlots, slots...)
			schedulesByDoctorDate[doctorDateKey] = append(existing, schedule)
		}
	}
	return created, createdSlots, skippedCount, nil
}

func selectGeneratedScheduleFeeRule(rules []models.MedRegistrationFeeRule, scheduleDate time.Time) (*models.MedRegistrationFeeRule, error) {
	var selected *models.MedRegistrationFeeRule
	for index := range rules {
		rule := &rules[index]
		if medicalDateAfter(rule.EffectiveDate, scheduleDate) || (rule.ExpiryDate != nil && medicalDateBefore(*rule.ExpiryDate, scheduleDate)) {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s匹配到重叠的挂号费规则", ErrMedicalConflict, scheduleDate.Format("2006-01-02"))
		}
		selected = rule
	}
	if selected == nil {
		return nil, fmt.Errorf("%w: %s没有可用的挂号费规则", ErrMedicalConflict, scheduleDate.Format("2006-01-02"))
	}
	return selected, nil
}

func loadScheduleGenerationResult(tx *gorm.DB, batch models.MedScheduleGenerationBatch, idempotent bool) (*models.GenerateSchedulesResponse, error) {
	var schedules []models.MedSchedule
	if err := tx.Where("generation_batch_id = ? AND del_flag = 0", batch.BatchID).Order("schedule_id asc").Find(&schedules).Error; err != nil {
		return nil, err
	}
	return &models.GenerateSchedulesResponse{
		BatchID:        batch.BatchID,
		Idempotent:     idempotent,
		GeneratedCount: batch.GeneratedCount,
		SkippedCount:   batch.SkippedCount,
		ScheduleIDs:    scheduleIDsFromRows(schedules),
	}, nil
}

func scheduleDimensionKey(doctorID, departmentID, registrationType string) string {
	return doctorID + "|" + departmentID + "|" + registrationType
}

func scheduleDoctorDepartmentKey(doctorID, departmentID string) string {
	return doctorID + "|" + departmentID
}

func scheduleDoctorDateKey(doctorID string, scheduleDate time.Time) string {
	return doctorID + "|" + scheduleDate.Format("2006-01-02")
}

func uniqueScheduleStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

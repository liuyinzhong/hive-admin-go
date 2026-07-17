package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func StartMedicalScheduleAutoScheduler() {
	go func() {
		service := NewMedicalScheduleService()
		for {
			if err := markInterruptedScheduleAutoTasks(time.Now()); err != nil {
				log.Printf("清理中断的出诊排班自动任务失败: %v", err)
			}
			now := time.Now().In(medicalBusinessLocation)
			nextRun := nextScheduleAutomationRun(now)
			timer := time.NewTimer(time.Until(nextRun))
			<-timer.C
			if err := service.RunWeeklyScheduleAutomation(nextRun); err != nil {
				log.Printf("出诊排班每周自动任务执行失败: %v", err)
			}
		}
	}()
}

func nextScheduleAutomationRun(now time.Time) time.Time {
	localNow := now.In(medicalBusinessLocation)
	daysUntilSunday := (int(time.Sunday) - int(localNow.Weekday()) + 7) % 7
	candidateDate := localNow.AddDate(0, 0, daysUntilSunday)
	candidate := time.Date(candidateDate.Year(), candidateDate.Month(), candidateDate.Day(), 20, 0, 0, 0, medicalBusinessLocation)
	if !candidate.After(localNow) {
		candidate = candidate.AddDate(0, 0, 7)
	}
	return candidate
}

func scheduleWeekMonday(value time.Time) time.Time {
	local := value.In(medicalBusinessLocation)
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, medicalBusinessLocation)
	weekday := isoWeekday(date)
	return date.AddDate(0, 0, -(weekday - 1))
}

func (s *MedicalScheduleService) RunWeeklyScheduleAutomation(executedAt time.Time) error {
	currentWeekStart := scheduleWeekMonday(executedAt)
	nextWeekStart := currentWeekStart.AddDate(0, 0, 7)
	nextWeekEnd := nextWeekStart.AddDate(0, 0, 6)
	followingWeekStart := nextWeekStart.AddDate(0, 0, 7)
	followingWeekEnd := followingWeekStart.AddDate(0, 0, 6)
	if err := s.runAutomaticPublish(nextWeekStart, nextWeekEnd, executedAt); err != nil {
		return err
	}
	return s.runAutomaticGeneration(followingWeekStart, followingWeekEnd, executedAt)
}

func (s *MedicalScheduleService) runAutomaticPublish(startDate, endDate, executedAt time.Time) (resultErr error) {
	taskKey := "publish:" + startDate.Format("2006-01-02")
	claimed, err := claimScheduleAutoTask(taskKey, models.MedScheduleAutoTaskTypePublish, startDate, endDate, executedAt)
	if err != nil || !claimed {
		return err
	}
	defer func() {
		if resultErr != nil {
			if markErr := failScheduleAutoTask(taskKey, resultErr); markErr != nil {
				resultErr = fmt.Errorf("%v; 自动任务失败状态保存失败: %w", resultErr, markErr)
			}
		}
	}()
	var schedules []models.MedSchedule
	if err := database.DB.Where("schedule_date BETWEEN ? AND ? AND status = ? AND del_flag = 0", startDate, endDate, models.MedScheduleStatusDraft).
		Order("doctor_id asc, schedule_date asc, start_time asc, schedule_id asc").Find(&schedules).Error; err != nil {
		return err
	}
	idsByDoctor := make(map[string][]string)
	for _, schedule := range schedules {
		idsByDoctor[schedule.DoctorID] = append(idsByDoctor[schedule.DoctorID], schedule.ScheduleID)
	}
	doctorIDs := sortedScheduleMapKeys(idsByDoctor)
	doctorNames, err := loadScheduleDoctorNames(doctorIDs)
	if err != nil {
		return err
	}
	failures := make([]models.ScheduleAutoTaskFailure, 0)
	successCount := 0
	for _, doctorID := range doctorIDs {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			return publishScheduleIDsTx(tx, idsByDoctor[doctorID], "")
		})
		if err != nil {
			failures = append(failures, models.ScheduleAutoTaskFailure{DoctorID: doctorID, DoctorName: doctorNames[doctorID], Reason: err.Error()})
			continue
		}
		successCount++
	}
	return saveScheduleAutoTask(taskKey, models.MedScheduleAutoTaskTypePublish, startDate, endDate, executedAt, successCount, failures)
}

func (s *MedicalScheduleService) runAutomaticGeneration(startDate, endDate, executedAt time.Time) (resultErr error) {
	taskKey := "generate:" + startDate.Format("2006-01-02")
	claimed, err := claimScheduleAutoTask(taskKey, models.MedScheduleAutoTaskTypeGenerate, startDate, endDate, executedAt)
	if err != nil || !claimed {
		return err
	}
	defer func() {
		if resultErr != nil {
			if markErr := failScheduleAutoTask(taskKey, resultErr); markErr != nil {
				resultErr = fmt.Errorf("%v; 自动任务失败状态保存失败: %w", resultErr, markErr)
			}
		}
	}()
	var templates []models.MedScheduleTemplate
	if err := database.DB.Where("status = 1 AND effective_date <= ? AND (expiry_date IS NULL OR expiry_date >= ?) AND del_flag = 0", endDate, startDate).
		Order("doctor_id asc, template_id asc").Find(&templates).Error; err != nil {
		return err
	}
	templateIDsByDoctor := make(map[string][]string)
	for _, template := range templates {
		templateIDsByDoctor[template.DoctorID] = append(templateIDsByDoctor[template.DoctorID], template.TemplateID)
	}
	doctorIDs := sortedScheduleMapKeys(templateIDsByDoctor)
	doctorNames, err := loadScheduleDoctorNames(doctorIDs)
	if err != nil {
		return err
	}
	failures := make([]models.ScheduleAutoTaskFailure, 0)
	successCount := 0
	for _, doctorID := range doctorIDs {
		_, err := s.GenerateSchedules(models.GenerateSchedulesRequest{
			IdempotencyKey: fmt.Sprintf("auto:%s:%s", startDate.Format("2006-01-02"), doctorID),
			TemplateIDs:    templateIDsByDoctor[doctorID],
			StartDate:      startDate.Format("2006-01-02"),
			EndDate:        endDate.Format("2006-01-02"),
		}, "")
		if err != nil {
			failures = append(failures, models.ScheduleAutoTaskFailure{DoctorID: doctorID, DoctorName: doctorNames[doctorID], Reason: err.Error()})
			continue
		}
		successCount++
	}
	return saveScheduleAutoTask(taskKey, models.MedScheduleAutoTaskTypeGenerate, startDate, endDate, executedAt, successCount, failures)
}

func claimScheduleAutoTask(taskKey, taskType string, startDate, endDate, executedAt time.Time) (bool, error) {
	details := "[]"
	now := time.Now()
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_key"}},
		DoNothing: true,
	}).Create(&models.MedScheduleAutoTask{
		TaskID: utils.GenerateUUID(), TaskKey: taskKey, TaskType: taskType,
		TargetWeekStart: startDate, TargetWeekEnd: endDate, Status: models.MedScheduleAutoTaskStatusProcessing,
		Details: &details, ExecutedAt: executedAt, CreateDate: &now, UpdateDate: &now,
	})
	return result.RowsAffected == 1, result.Error
}

func saveScheduleAutoTask(taskKey, taskType string, startDate, endDate, executedAt time.Time, successCount int, failures []models.ScheduleAutoTaskFailure) error {
	status := models.MedScheduleAutoTaskStatusSuccess
	if len(failures) > 0 && successCount > 0 {
		status = models.MedScheduleAutoTaskStatusPartial
	} else if len(failures) > 0 {
		status = models.MedScheduleAutoTaskStatusFailed
	}
	details, err := json.Marshal(failures)
	if err != nil {
		return err
	}
	detailValue := string(details)
	return database.DB.Model(&models.MedScheduleAutoTask{}).
		Where("task_key = ? AND task_type = ?", taskKey, taskType).
		Updates(map[string]interface{}{
			"status": status, "success_doctor_count": successCount,
			"failure_doctor_count": len(failures), "details": detailValue, "update_date": time.Now(),
		}).Error
}

func failScheduleAutoTask(taskKey string, taskErr error) error {
	details, err := json.Marshal([]models.ScheduleAutoTaskFailure{{Reason: taskErr.Error()}})
	if err != nil {
		return err
	}
	return database.DB.Model(&models.MedScheduleAutoTask{}).
		Where("task_key = ?", taskKey).
		Updates(map[string]interface{}{
			"status": models.MedScheduleAutoTaskStatusFailed, "failure_doctor_count": 1,
			"details": string(details), "update_date": time.Now(),
		}).Error
}

func markInterruptedScheduleAutoTasks(now time.Time) error {
	details, err := json.Marshal([]models.ScheduleAutoTaskFailure{{Reason: "服务执行中断，请管理员检查草稿并手工处理"}})
	if err != nil {
		return err
	}
	// 自动任务正常执行远低于六小时；只标记陈旧记录，不自动补跑。
	return database.DB.Model(&models.MedScheduleAutoTask{}).
		Where("status = ? AND update_date < ?", models.MedScheduleAutoTaskStatusProcessing, now.Add(-6*time.Hour)).
		Updates(map[string]interface{}{
			"status": models.MedScheduleAutoTaskStatusFailed, "failure_doctor_count": 1,
			"details": string(details), "update_date": now,
		}).Error
}

func loadScheduleDoctorNames(doctorIDs []string) (map[string]string, error) {
	result := make(map[string]string, len(doctorIDs))
	if len(doctorIDs) == 0 {
		return result, nil
	}
	var doctors []models.MedDoctor
	if err := database.DB.Select("doctor_id", "name").Where("doctor_id IN ? AND del_flag = 0", doctorIDs).Find(&doctors).Error; err != nil {
		return nil, err
	}
	for _, doctor := range doctors {
		result[doctor.DoctorID] = doctor.Name
	}
	return result, nil
}

func sortedScheduleMapKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *MedicalScheduleService) GetScheduleAutoTaskList(req models.ScheduleAutoTaskListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.MedScheduleAutoTask{})
	if req.TaskType != "" {
		if req.TaskType != models.MedScheduleAutoTaskTypePublish && req.TaskType != models.MedScheduleAutoTaskTypeGenerate {
			return nil, fmt.Errorf("%w: 自动任务类型不正确", ErrMedicalInvalidInput)
		}
		query = query.Where("task_type = ?", req.TaskType)
	}
	if req.Status != nil {
		if *req.Status < models.MedScheduleAutoTaskStatusSuccess || *req.Status > models.MedScheduleAutoTaskStatusProcessing {
			return nil, fmt.Errorf("%w: 自动任务状态不正确", ErrMedicalInvalidInput)
		}
		query = query.Where("status = ?", *req.Status)
	}
	var startDate *time.Time
	if req.StartDate != "" {
		parsed, err := parseRequiredMedicalDate(req.StartDate, "开始日期")
		if err != nil {
			return nil, err
		}
		startDate = &parsed
		query = query.Where("target_week_start >= ?", parsed)
	}
	var endDate *time.Time
	if req.EndDate != "" {
		parsed, err := parseRequiredMedicalDate(req.EndDate, "结束日期")
		if err != nil {
			return nil, err
		}
		endDate = &parsed
		query = query.Where("target_week_end <= ?", parsed)
	}
	if startDate != nil && endDate != nil && medicalDateBefore(*endDate, *startDate) {
		return nil, fmt.Errorf("%w: 结束日期不能早于开始日期", ErrMedicalInvalidInput)
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{"executedAt": "executed_at", "targetWeekStart": "target_week_start", "status": "status"})
	if order == "" {
		order = "executed_at desc, task_id asc"
	} else {
		order += ", task_id asc"
	}
	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var rows []models.MedScheduleAutoTask
	pageResult, err := utils.Paginate(query.Order(order), req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	items := make([]models.ScheduleAutoTaskResponse, 0, len(rows))
	for _, row := range rows {
		failures := []models.ScheduleAutoTaskFailure{}
		if row.Details != nil && *row.Details != "" {
			if err := json.Unmarshal([]byte(*row.Details), &failures); err != nil {
				return nil, err
			}
		}
		items = append(items, models.ScheduleAutoTaskResponse{
			TaskID: row.TaskID, TaskType: row.TaskType,
			TargetWeekStart: row.TargetWeekStart.Format("2006-01-02"), TargetWeekEnd: row.TargetWeekEnd.Format("2006-01-02"),
			Status: row.Status, SuccessDoctorCount: row.SuccessDoctorCount, FailureDoctorCount: row.FailureDoctorCount,
			Failures: failures, ExecutedAt: row.ExecutedAt.Format("2006-01-02 15:04:05"),
		})
	}
	pageResult.Items = items
	return pageResult, nil
}

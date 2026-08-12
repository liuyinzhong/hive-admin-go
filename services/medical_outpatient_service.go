package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

type MedicalOutpatientService struct{}

type workbenchQueueRow struct {
	models.MedVisitQueue
	RecordID       *string `gorm:"column:record_id"`
	PatientNo      string  `gorm:"column:patient_no"`
	PatientName    string  `gorm:"column:patient_name"`
	PatientPhone   string  `gorm:"column:patient_phone"`
	RegistrationNo string  `gorm:"column:registration_no"`
	StartTime      string  `gorm:"column:start_time"`
	EndTime        string  `gorm:"column:end_time"`
}

type outpatientRecordRow struct {
	models.MedOutpatientRecord
	RegistrationNo   string    `gorm:"column:registration_no"`
	PatientNo        string    `gorm:"column:patient_no"`
	PatientName      string    `gorm:"column:patient_name"`
	PatientGender    string    `gorm:"column:patient_gender"`
	PatientBirthDate time.Time `gorm:"column:patient_birth_date"`
	PatientPhone     string    `gorm:"column:patient_phone"`
	DoctorName       string    `gorm:"column:doctor_name"`
	DepartmentName   string    `gorm:"column:department_name"`
	QueueID          string    `gorm:"column:queue_id"`
	QueueSequence    int       `gorm:"column:queue_sequence"`
	QueueStatus      int       `gorm:"column:queue_status"`
}

func NewMedicalOutpatientService() *MedicalOutpatientService {
	return &MedicalOutpatientService{}
}

func (s *MedicalOutpatientService) GetWorkbench(userID string) (*models.DoctorWorkbenchResponse, error) {
	doctor, err := requireWorkbenchDoctor(database.DB, userID)
	if err != nil {
		return nil, err
	}
	var schedules []models.MedSchedule
	if err := database.DB.Where("doctor_id = ? AND schedule_date = ? AND status = ? AND del_flag = 0", doctor.DoctorID, medicalToday(), models.MedScheduleStatusPublished).
		Order("start_time asc, schedule_id asc").Find(&schedules).Error; err != nil {
		return nil, err
	}
	result := &models.DoctorWorkbenchResponse{
		DoctorID: doctor.DoctorID, DoctorNo: doctor.DoctorNo, DoctorName: doctor.Name,
		Schedules: make([]models.DoctorWorkbenchScheduleResponse, 0, len(schedules)),
	}
	if len(schedules) == 0 {
		return result, nil
	}
	scheduleIDs := make([]string, 0, len(schedules))
	departmentIDs := make([]string, 0, len(schedules))
	for _, schedule := range schedules {
		scheduleIDs = append(scheduleIDs, schedule.ScheduleID)
		departmentIDs = append(departmentIDs, schedule.DepartmentID)
	}
	var departments []models.MedDepartment
	if err := database.DB.Select("department_id", "department_name").Where("department_id IN ?", departmentIDs).Find(&departments).Error; err != nil {
		return nil, err
	}
	departmentNames := make(map[string]string, len(departments))
	for _, department := range departments {
		departmentNames[department.DepartmentID] = department.DepartmentName
	}
	var queueRows []workbenchQueueRow
	if err := database.DB.Table("med_visit_queue AS queue").
		Select("queue.*, record.record_id, registration.patient_no, registration.patient_name, registration.patient_phone, registration.registration_no, registration.start_time, registration.end_time").
		Joins("JOIN med_registration AS registration ON registration.registration_id = queue.registration_id").
		Joins("LEFT JOIN med_outpatient_record AS record ON record.registration_id = registration.registration_id").
		Where("queue.schedule_id IN ?", scheduleIDs).
		Order("queue.schedule_id asc, queue.queue_sequence asc").Scan(&queueRows).Error; err != nil {
		return nil, err
	}
	queuesBySchedule := make(map[string][]models.DoctorWorkbenchQueueResponse, len(schedules))
	for _, schedule := range schedules {
		queuesBySchedule[schedule.ScheduleID] = make([]models.DoctorWorkbenchQueueResponse, 0)
	}
	for _, row := range queueRows {
		queuesBySchedule[row.ScheduleID] = append(queuesBySchedule[row.ScheduleID], models.DoctorWorkbenchQueueResponse{
			QueueID: row.QueueID, RegistrationID: row.RegistrationID, RecordID: row.RecordID,
			QueueSequence: row.QueueSequence, QueueStatus: row.QueueStatus, CallCount: row.CallCount,
			PatientNo: row.PatientNo, PatientName: maskPatientName(row.PatientName), PatientPhone: maskPatientPhone(row.PatientPhone),
			RegistrationNo: row.RegistrationNo, StartTime: trimScheduleTime(row.StartTime), EndTime: trimScheduleTime(row.EndTime),
			CheckInTime: row.CreateDate.Format("2006-01-02 15:04:05"),
		})
	}
	for _, schedule := range schedules {
		result.Schedules = append(result.Schedules, models.DoctorWorkbenchScheduleResponse{
			ScheduleID: schedule.ScheduleID, ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
			StartTime: trimScheduleTime(schedule.StartTime), EndTime: trimScheduleTime(schedule.EndTime),
			DepartmentID: schedule.DepartmentID, DepartmentName: departmentNames[schedule.DepartmentID],
			RegistrationType: schedule.RegistrationType, Queues: queuesBySchedule[schedule.ScheduleID],
		})
	}
	return result, nil
}

func (s *MedicalOutpatientService) CallNext(scheduleID, userID string) error {
	if err := validateMedicalUUID(scheduleID, "排班ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		if err := lockOwnedPublishedSchedule(tx, scheduleID, doctor.DoctorID); err != nil {
			return err
		}
		if err := ensureScheduleQueueIdle(tx, scheduleID, ""); err != nil {
			return err
		}
		var queue models.MedVisitQueue
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("schedule_id = ? AND queue_status = ?", scheduleID, models.MedVisitQueueStatusWaiting).
			Order("queue_sequence asc").First(&queue).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 当前没有待叫号患者", ErrMedicalConflict)
			}
			return err
		}
		return tx.Model(&queue).Updates(map[string]interface{}{
			"queue_status": models.MedVisitQueueStatusCalled,
			"call_count":   gorm.Expr("call_count + 1"),
		}).Error
	})
}

func (s *MedicalOutpatientService) RepeatCall(queueID, userID string) error {
	return s.transitionQueue(queueID, userID, models.MedVisitQueueStatusCalled, models.MedVisitQueueStatusCalled, true)
}

func (s *MedicalOutpatientService) PassQueue(queueID, userID string) error {
	return s.transitionQueue(queueID, userID, models.MedVisitQueueStatusCalled, models.MedVisitQueueStatusPassed, false)
}

func (s *MedicalOutpatientService) RecallQueue(queueID, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		scheduleID, err := findOwnedQueueScheduleID(tx, queueID, doctor.DoctorID)
		if err != nil {
			return err
		}
		if err := lockOwnedPublishedSchedule(tx, scheduleID, doctor.DoctorID); err != nil {
			return err
		}
		queue, err := lockOwnedQueue(tx, queueID, doctor.DoctorID)
		if err != nil {
			return err
		}
		if queue.QueueStatus != models.MedVisitQueueStatusPassed {
			return fmt.Errorf("%w: 只有已过号患者可以重新叫号", ErrMedicalConflict)
		}
		if err := ensureScheduleQueueIdle(tx, queue.ScheduleID, queue.QueueID); err != nil {
			return err
		}
		return tx.Model(queue).Updates(map[string]interface{}{
			"queue_status": models.MedVisitQueueStatusCalled,
			"call_count":   gorm.Expr("call_count + 1"),
		}).Error
	})
}

func (s *MedicalOutpatientService) StartConsultation(queueID, userID string) (*models.OutpatientRecordResponse, error) {
	var recordID string
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		queue, err := lockOwnedQueue(tx, queueID, doctor.DoctorID)
		if err != nil {
			return err
		}
		if queue.QueueStatus != models.MedVisitQueueStatusCalled {
			return fmt.Errorf("%w: 只有已叫号患者可以开始接诊", ErrMedicalConflict)
		}
		var registration models.MedRegistration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("registration_id = ?", queue.RegistrationID).First(&registration).Error; err != nil {
			return err
		}
		if registration.Status != models.MedRegistrationStatusCheckedIn {
			return fmt.Errorf("%w: 挂号单状态不允许开始接诊", ErrMedicalConflict)
		}
		var existing models.MedOutpatientRecord
		if err := tx.Where("registration_id = ?", registration.RegistrationID).First(&existing).Error; err == nil {
			recordID = existing.RecordID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		} else {
			now := time.Now()
			recordID = utils.GenerateUUID()
			record := models.MedOutpatientRecord{
				RecordID: recordID, RegistrationID: registration.RegistrationID, PatientID: registration.PatientID,
				DoctorID: registration.DoctorID, DepartmentID: registration.DepartmentID, StartDate: now,
				CreatorID: optionalOperatorID(userID), UpdaterID: optionalOperatorID(userID), CreateDate: &now, UpdateDate: &now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return tx.Model(queue).Update("queue_status", models.MedVisitQueueStatusConsulting).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetOutpatientRecord(recordID, userID)
}

func (s *MedicalOutpatientService) GetOutpatientRecord(recordID, userID string) (*models.OutpatientRecordResponse, error) {
	doctor, err := requireWorkbenchDoctor(database.DB, userID)
	if err != nil {
		return nil, err
	}
	return loadOutpatientRecordResponse(database.DB, recordID, doctor.DoctorID)
}

func (s *MedicalOutpatientService) SaveOutpatientRecord(recordID, userID string, req models.SaveOutpatientRecordRequest) (*models.OutpatientRecordResponse, error) {
	if err := validateMedicalUUID(recordID, "门诊病历ID"); err != nil {
		return nil, err
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		var record models.MedOutpatientRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("record_id = ? AND doctor_id = ?", recordID, doctor.DoctorID).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 门诊病历不存在或不属于当前医生", ErrMedicalNotFound)
			}
			return err
		}
		if record.EndDate != nil {
			return fmt.Errorf("%w: 已完成的门诊病历不能修改", ErrMedicalConflict)
		}
		var queue models.MedVisitQueue
		if err := tx.Where("registration_id = ? AND queue_status = ?", record.RegistrationID, models.MedVisitQueueStatusConsulting).First(&queue).Error; err != nil {
			return fmt.Errorf("%w: 当前患者不在接诊中", ErrMedicalConflict)
		}
		updates := outpatientRecordUpdates(req, userID)
		if err := tx.Model(&record).Updates(updates).Error; err != nil {
			return err
		}
		return replaceOutpatientDiagnoses(tx, recordID, userID, req.Diagnoses)
	})
	if err != nil {
		return nil, err
	}
	return s.GetOutpatientRecord(recordID, userID)
}

func (s *MedicalOutpatientService) CompleteOutpatientRecord(recordID, userID string) (*models.OutpatientRecordResponse, error) {
	if err := validateMedicalUUID(recordID, "门诊病历ID"); err != nil {
		return nil, err
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		var record models.MedOutpatientRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("record_id = ? AND doctor_id = ?", recordID, doctor.DoctorID).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 门诊病历不存在或不属于当前医生", ErrMedicalNotFound)
			}
			return err
		}
		if record.EndDate != nil {
			return fmt.Errorf("%w: 门诊病历已完成", ErrMedicalConflict)
		}
		if record.VisitType == nil || blank(record.ChiefComplaint) || blank(record.PresentIllness) || blank(record.TreatmentPlan) {
			return fmt.Errorf("%w: 完成接诊前请填写初复诊、主诉、现病史和处理方案", ErrMedicalInvalidInput)
		}
		var primaryCount int64
		if err := tx.Model(&models.MedOutpatientDiagnosis{}).Where("record_id = ? AND is_primary = 1", recordID).Count(&primaryCount).Error; err != nil {
			return err
		}
		if primaryCount != 1 {
			return fmt.Errorf("%w: 完成接诊前必须且只能选择一个主要诊断", ErrMedicalInvalidInput)
		}
		var blockedPrescriptionCount int64
		if err := tx.Model(&models.MedPrescription{}).
			Where("record_id = ? AND status IN ?", recordID, []int{models.MedPrescriptionStatusDraft, models.MedPrescriptionStatusRejected}).
			Count(&blockedPrescriptionCount).Error; err != nil {
			return err
		}
		if blockedPrescriptionCount > 0 {
			return fmt.Errorf("%w: 草稿或审核不通过的处方必须提交或作废后才能完成接诊", ErrMedicalConflict)
		}
		var queue models.MedVisitQueue
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("registration_id = ?", record.RegistrationID).First(&queue).Error; err != nil {
			return err
		}
		if queue.QueueStatus != models.MedVisitQueueStatusConsulting {
			return fmt.Errorf("%w: 候诊状态不允许完成接诊", ErrMedicalConflict)
		}
		var registration models.MedRegistration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("registration_id = ?", record.RegistrationID).First(&registration).Error; err != nil {
			return err
		}
		if registration.Status != models.MedRegistrationStatusCheckedIn {
			return fmt.Errorf("%w: 挂号单状态不允许完成接诊", ErrMedicalConflict)
		}
		now := time.Now()
		if err := tx.Model(&record).Updates(map[string]interface{}{"end_date": now, "updater_id": optionalOperatorID(userID), "update_date": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&queue).Update("queue_status", models.MedVisitQueueStatusCompleted).Error; err != nil {
			return err
		}
		if err := tx.Model(&registration).Updates(map[string]interface{}{"status": models.MedRegistrationStatusCompleted, "updater_id": optionalOperatorID(userID), "update_date": now}).Error; err != nil {
			return err
		}
		fromStatus := registration.Status
		return createRegistrationLog(tx, registration.RegistrationID, &fromStatus, models.MedRegistrationStatusCompleted, userID, now, nil, nil)
	})
	if err != nil {
		return nil, err
	}
	return s.GetOutpatientRecord(recordID, userID)
}

func (s *MedicalOutpatientService) GetPatientHistory(recordID, userID string) ([]models.OutpatientRecordResponse, error) {
	current, err := s.GetOutpatientRecord(recordID, userID)
	if err != nil {
		return nil, err
	}
	if current.QueueStatus != models.MedVisitQueueStatusConsulting && current.QueueStatus != models.MedVisitQueueStatusCompleted {
		return nil, fmt.Errorf("%w: 开始接诊后才能查看患者历史病历", ErrMedicalConflict)
	}
	var ids []string
	if err := database.DB.Table("med_outpatient_record AS record").
		Joins("JOIN med_registration AS registration ON registration.registration_id = record.registration_id").
		Where("record.patient_id = ? AND record.record_id <> ? AND record.end_date IS NOT NULL AND registration.status = ?", current.PatientID, recordID, models.MedRegistrationStatusCompleted).
		Order("record.end_date desc").Limit(50).Pluck("record.record_id", &ids).Error; err != nil {
		return nil, err
	}
	return loadOutpatientRecordResponses(database.DB, ids, "")
}

func (s *MedicalOutpatientService) transitionQueue(queueID, userID string, fromStatus, toStatus int, incrementCall bool) error {
	if err := validateMedicalUUID(queueID, "候诊记录ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		queue, err := lockOwnedQueue(tx, queueID, doctor.DoctorID)
		if err != nil {
			return err
		}
		if queue.QueueStatus != fromStatus {
			return fmt.Errorf("%w: 当前候诊状态不允许执行该操作", ErrMedicalConflict)
		}
		updates := map[string]interface{}{"queue_status": toStatus}
		if incrementCall {
			updates["call_count"] = gorm.Expr("call_count + 1")
		}
		return tx.Model(queue).Updates(updates).Error
	})
}

func findOwnedQueueScheduleID(db *gorm.DB, queueID, doctorID string) (string, error) {
	if err := validateMedicalUUID(queueID, "候诊记录ID"); err != nil {
		return "", err
	}
	var row struct {
		ScheduleID string `gorm:"column:schedule_id"`
	}
	err := db.Table("med_visit_queue AS queue").Select("queue.schedule_id").
		Joins("JOIN med_schedule AS schedule ON schedule.schedule_id = queue.schedule_id AND schedule.del_flag = 0").
		Where("queue.queue_id = ? AND schedule.doctor_id = ? AND schedule.schedule_date = ? AND schedule.status = ?", queueID, doctorID, medicalToday(), models.MedScheduleStatusPublished).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("%w: 候诊记录不存在或不属于当前医生", ErrMedicalNotFound)
		}
		return "", err
	}
	return row.ScheduleID, nil
}

func requireWorkbenchDoctor(db *gorm.DB, userID string) (*models.MedDoctor, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: 未获取到当前用户", ErrMedicalInvalidInput)
	}
	var doctor models.MedDoctor
	if err := db.Where("user_id = ? AND status = 1 AND del_flag = 0", userID).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 当前用户未绑定启用的医生档案", ErrMedicalConflict)
		}
		return nil, err
	}
	return &doctor, nil
}

func lockOwnedPublishedSchedule(tx *gorm.DB, scheduleID, doctorID string) error {
	var schedule models.MedSchedule
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("schedule_id = ? AND doctor_id = ? AND schedule_date = ? AND status = ? AND del_flag = 0", scheduleID, doctorID, medicalToday(), models.MedScheduleStatusPublished).
		First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 排班不存在、不是今日排班或不属于当前医生", ErrMedicalNotFound)
		}
		return err
	}
	return nil
}

func lockOwnedQueue(tx *gorm.DB, queueID, doctorID string) (*models.MedVisitQueue, error) {
	if err := validateMedicalUUID(queueID, "候诊记录ID"); err != nil {
		return nil, err
	}
	var queue models.MedVisitQueue
	if err := tx.Table("med_visit_queue AS queue").Select("queue.*").
		Joins("JOIN med_schedule AS schedule ON schedule.schedule_id = queue.schedule_id AND schedule.del_flag = 0").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("queue.queue_id = ? AND schedule.doctor_id = ? AND schedule.schedule_date = ? AND schedule.status = ?", queueID, doctorID, medicalToday(), models.MedScheduleStatusPublished).
		First(&queue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 候诊记录不存在或不属于当前医生", ErrMedicalNotFound)
		}
		return nil, err
	}
	return &queue, nil
}

func ensureScheduleQueueIdle(tx *gorm.DB, scheduleID, excludeQueueID string) error {
	query := tx.Model(&models.MedVisitQueue{}).
		Where("schedule_id = ? AND queue_status IN ?", scheduleID, []int{models.MedVisitQueueStatusCalled, models.MedVisitQueueStatusConsulting})
	if excludeQueueID != "" {
		query = query.Where("queue_id <> ?", excludeQueueID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前排班已有叫号或接诊中的患者", ErrMedicalConflict)
	}
	return nil
}

func outpatientRecordUpdates(req models.SaveOutpatientRecordRequest, userID string) map[string]interface{} {
	return map[string]interface{}{
		"visit_type": req.VisitType, "informant": normalizeMedicalOptionalString(req.Informant),
		"temperature": normalizeMedicalOptionalString(req.Temperature), "pulse": req.Pulse, "respiratory_rate": req.RespiratoryRate,
		"systolic_pressure": req.SystolicPressure, "diastolic_pressure": req.DiastolicPressure,
		"height": normalizeMedicalOptionalString(req.Height), "weight": normalizeMedicalOptionalString(req.Weight),
		"chief_complaint": normalizeMedicalOptionalString(req.ChiefComplaint), "present_illness": normalizeMedicalOptionalString(req.PresentIllness),
		"past_history": normalizeMedicalOptionalString(req.PastHistory), "personal_history": normalizeMedicalOptionalString(req.PersonalHistory),
		"family_history": normalizeMedicalOptionalString(req.FamilyHistory), "allergy_history": normalizeMedicalOptionalString(req.AllergyHistory),
		"marital_reproductive": normalizeMedicalOptionalString(req.MaritalReproductive), "menstrual_history": normalizeMedicalOptionalString(req.MenstrualHistory),
		"physical_examination": normalizeMedicalOptionalString(req.PhysicalExamination), "specialist_examination": normalizeMedicalOptionalString(req.SpecialistExamination),
		"auxiliary_examination": normalizeMedicalOptionalString(req.AuxiliaryExamination), "treatment_plan": normalizeMedicalOptionalString(req.TreatmentPlan),
		"medical_advice": normalizeMedicalOptionalString(req.MedicalAdvice), "follow_up_advice": normalizeMedicalOptionalString(req.FollowUpAdvice),
		"remark": normalizeMedicalOptionalString(req.Remark), "updater_id": optionalOperatorID(userID), "update_date": time.Now(),
	}
}

func replaceOutpatientDiagnoses(tx *gorm.DB, recordID, userID string, requests []models.SaveOutpatientDiagnosisRequest) error {
	seen := make(map[string]struct{}, len(requests))
	primaryCount := 0
	rows := make([]models.MedOutpatientDiagnosis, 0, len(requests))
	now := time.Now()
	for index, request := range requests {
		if err := validateMedicalUUID(request.DiagnosisID, "诊断ID"); err != nil {
			return err
		}
		if _, exists := seen[request.DiagnosisID]; exists {
			return fmt.Errorf("%w: 同一份病历不能重复选择疾病诊断", ErrMedicalInvalidInput)
		}
		seen[request.DiagnosisID] = struct{}{}
		if request.IsPrimary == 1 {
			primaryCount++
		}
		var diagnosis models.MedDiagnosis
		if err := tx.Where("diagnosis_id = ? AND status = 1 AND del_flag = 0", request.DiagnosisID).First(&diagnosis).Error; err != nil {
			return fmt.Errorf("%w: 选择的疾病诊断不存在或已停用", ErrMedicalInvalidInput)
		}
		rows = append(rows, models.MedOutpatientDiagnosis{
			RecordDiagnosisID: utils.GenerateUUID(), RecordID: recordID, DiagnosisID: diagnosis.DiagnosisID,
			ICDCode: diagnosis.ICDCode, ICDName: diagnosis.ICDName, IsPrimary: request.IsPrimary,
			Sort: index, CreatorID: optionalOperatorID(userID), CreateDate: now,
		})
	}
	if primaryCount > 1 {
		return fmt.Errorf("%w: 一份门诊病历最多一个主要诊断", ErrMedicalInvalidInput)
	}
	if err := tx.Where("record_id = ?", recordID).Delete(&models.MedOutpatientDiagnosis{}).Error; err != nil {
		return err
	}
	if len(rows) > 0 {
		return tx.Create(&rows).Error
	}
	return nil
}

func loadOutpatientRecordResponse(db *gorm.DB, recordID, doctorID string) (*models.OutpatientRecordResponse, error) {
	if err := validateMedicalUUID(recordID, "门诊病历ID"); err != nil {
		return nil, err
	}
	items, err := loadOutpatientRecordResponses(db, []string{recordID}, doctorID)
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func loadOutpatientRecordResponses(db *gorm.DB, recordIDs []string, doctorID string) ([]models.OutpatientRecordResponse, error) {
	result := make([]models.OutpatientRecordResponse, 0, len(recordIDs))
	if len(recordIDs) == 0 {
		return result, nil
	}
	query := db.Table("med_outpatient_record AS record").
		Select("record.*, registration.registration_no, registration.patient_no, registration.patient_name, registration.patient_gender, registration.patient_birth_date, registration.patient_phone, registration.doctor_name, registration.department_name, queue.queue_id, queue.queue_sequence, queue.queue_status").
		Joins("JOIN med_registration AS registration ON registration.registration_id = record.registration_id").
		Joins("JOIN med_visit_queue AS queue ON queue.registration_id = record.registration_id").
		Where("record.record_id IN ?", recordIDs)
	if doctorID != "" {
		query = query.Where("record.doctor_id = ?", doctorID)
	}
	var rows []outpatientRecordRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	rowsByID := make(map[string]outpatientRecordRow, len(rows))
	for _, row := range rows {
		rowsByID[row.RecordID] = row
	}
	if len(rowsByID) != len(recordIDs) {
		return nil, fmt.Errorf("%w: 门诊病历不存在或不属于当前医生", ErrMedicalNotFound)
	}

	var diagnoses []models.MedOutpatientDiagnosis
	if err := db.Where("record_id IN ?", recordIDs).Order("record_id asc, is_primary desc, sort asc").Find(&diagnoses).Error; err != nil {
		return nil, err
	}
	diagnosesByRecord := make(map[string][]models.OutpatientDiagnosisResponse, len(recordIDs))
	for _, recordID := range recordIDs {
		diagnosesByRecord[recordID] = make([]models.OutpatientDiagnosisResponse, 0)
	}
	for _, diagnosis := range diagnoses {
		diagnosesByRecord[diagnosis.RecordID] = append(diagnosesByRecord[diagnosis.RecordID], models.OutpatientDiagnosisResponse{
			RecordDiagnosisID: diagnosis.RecordDiagnosisID, DiagnosisID: diagnosis.DiagnosisID,
			ICDCode: diagnosis.ICDCode, ICDName: diagnosis.ICDName, IsPrimary: diagnosis.IsPrimary, Sort: diagnosis.Sort,
		})
	}
	prescriptionRows, err := queryPrescriptionRows(db.Where("prescription.record_id IN ?", recordIDs))
	if err != nil {
		return nil, err
	}
	prescriptions, err := buildPrescriptionResponses(db, prescriptionRows)
	if err != nil {
		return nil, err
	}
	prescriptionsByRecord := make(map[string][]models.PrescriptionResponse, len(recordIDs))
	for _, recordID := range recordIDs {
		prescriptionsByRecord[recordID] = make([]models.PrescriptionResponse, 0)
	}
	for _, prescription := range prescriptions {
		prescriptionsByRecord[prescription.RecordID] = append(prescriptionsByRecord[prescription.RecordID], prescription)
	}
	for _, recordID := range recordIDs {
		row := rowsByID[recordID]
		result = append(result, models.OutpatientRecordResponse{
			MedOutpatientRecord: row.MedOutpatientRecord, RegistrationNo: row.RegistrationNo,
			PatientNo: row.PatientNo, PatientName: row.PatientName, PatientGender: row.PatientGender,
			PatientBirthDate: row.PatientBirthDate.Format("2006-01-02"), PatientPhone: row.PatientPhone,
			DoctorName: row.DoctorName, DepartmentName: row.DepartmentName, QueueID: row.QueueID,
			QueueSequence: row.QueueSequence, QueueStatus: row.QueueStatus,
			Diagnoses: diagnosesByRecord[recordID], Prescriptions: prescriptionsByRecord[recordID],
		})
	}
	return result, nil
}

func blank(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

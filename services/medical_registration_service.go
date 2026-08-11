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

type MedicalRegistrationService struct {
	codeSequenceService *BaseCodeSequenceService
}

type registrationLogRow struct {
	models.MedRegistrationLog
	OperatorName *string `gorm:"column:operator_name"`
}

type visitQueueListRow struct {
	models.MedVisitQueue
	PatientNo      string `gorm:"column:patient_no"`
	PatientName    string `gorm:"column:patient_name"`
	PatientPhone   string `gorm:"column:patient_phone"`
	RegistrationNo string `gorm:"column:registration_no"`
	StartTime      string `gorm:"column:start_time"`
	EndTime        string `gorm:"column:end_time"`
}

func NewMedicalRegistrationService() *MedicalRegistrationService {
	return &MedicalRegistrationService{codeSequenceService: NewBaseCodeSequenceService()}
}

func ValidateRegistrationTransition(fromStatus, toStatus int) error {
	allowed := map[int]map[int]struct{}{
		models.MedRegistrationStatusPendingPayment: {
			models.MedRegistrationStatusPaid: {}, models.MedRegistrationStatusCanceled: {},
		},
		models.MedRegistrationStatusPaid: {
			models.MedRegistrationStatusCheckedIn: {}, models.MedRegistrationStatusNoShow: {}, models.MedRegistrationStatusRefundStarted: {},
		},
		models.MedRegistrationStatusCheckedIn:     {models.MedRegistrationStatusCompleted: {}},
		models.MedRegistrationStatusRefundStarted: {models.MedRegistrationStatusRefunding: {}},
		models.MedRegistrationStatusRefunding:     {models.MedRegistrationStatusRefunded: {}},
	}
	if targets, ok := allowed[fromStatus]; ok {
		if _, ok := targets[toStatus]; ok {
			return nil
		}
	}
	return fmt.Errorf("%w: 当前挂号单状态不允许执行该操作", ErrMedicalConflict)
}

func RegistrationTransitionReleasesQuota(fromStatus, toStatus int) bool {
	return fromStatus == models.MedRegistrationStatusPendingPayment && toStatus == models.MedRegistrationStatusCanceled ||
		fromStatus == models.MedRegistrationStatusRefunding && toStatus == models.MedRegistrationStatusRefunded
}

func (s *MedicalRegistrationService) GetRegistrationList(req models.RegistrationListRequest, showSensitive bool) (*utils.PageResult, error) {
	query := database.DB.Model(&models.MedRegistration{})
	if value := strings.TrimSpace(req.RegistrationNo); value != "" {
		query = query.Where("registration_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(req.PatientKeyword); value != "" {
		like := "%" + value + "%"
		query = query.Where("patient_no LIKE ? OR patient_name LIKE ? OR patient_phone LIKE ?", like, like, like)
	}
	if req.StartDate != "" {
		value, err := parseMedicalQueryDate(req.StartDate, "就诊开始日期")
		if err != nil {
			return nil, err
		}
		query = query.Where("schedule_date >= ?", value)
	}
	if req.EndDate != "" {
		value, err := parseMedicalQueryDate(req.EndDate, "就诊结束日期")
		if err != nil {
			return nil, err
		}
		query = query.Where("schedule_date <= ?", value)
	}
	if value := strings.TrimSpace(req.DepartmentID); value != "" {
		if err := validateMedicalUUID(value, "科室ID"); err != nil {
			return nil, err
		}
		query = query.Where("department_id = ?", value)
	}
	if value := strings.TrimSpace(req.DoctorID); value != "" {
		if err := validateMedicalUUID(value, "医生ID"); err != nil {
			return nil, err
		}
		query = query.Where("doctor_id = ?", value)
	}
	if value := strings.TrimSpace(req.RegistrationType); value != "" {
		query = query.Where("registration_type = ?", value)
	}
	if req.RegistrationMethod != nil {
		if !validRegistrationMethod(*req.RegistrationMethod) {
			return nil, fmt.Errorf("%w: 挂号方式不正确", ErrMedicalInvalidInput)
		}
		query = query.Where("registration_method = ?", *req.RegistrationMethod)
	}
	if req.Status != nil {
		if !validRegistrationStatus(*req.Status) {
			return nil, fmt.Errorf("%w: 挂号状态不正确", ErrMedicalInvalidInput)
		}
		query = query.Where("status = ?", *req.Status)
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"registrationNo": "registration_no", "patientName": "patient_name", "scheduleDate": "schedule_date",
		"startTime": "start_time", "feeAmount": "fee_amount", "status": "status", "createDate": "create_date",
	})
	if order == "" {
		order = "create_date desc"
	}
	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var registrations []models.MedRegistration
	result, err := utils.Paginate(query.Order(order), req.Page, pageSize, &registrations)
	if err != nil {
		return nil, err
	}
	items := make([]*models.RegistrationResponse, 0, len(registrations))
	for _, registration := range registrations {
		items = append(items, registrationToResponse(registration, showSensitive))
	}
	result.Items = items
	return result, nil
}

func (s *MedicalRegistrationService) GetRegistrationDetail(registrationID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.getRegistrationDetail(database.DB, registrationID, showSensitive, true)
}

// GetVisitQueueList 按签到序号返回实际排班下的完整候诊队列。
// 队列场景固定返回脱敏后的患者姓名和手机号，不开放敏感信息权限例外。
func (s *MedicalRegistrationService) GetVisitQueueList(scheduleID string) ([]models.VisitQueueListItemResponse, error) {
	if err := validateMedicalUUID(scheduleID, "排班ID"); err != nil {
		return nil, err
	}
	var schedule models.MedSchedule
	if err := database.DB.Select("schedule_id").Where("schedule_id = ? AND del_flag = 0", scheduleID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 排班不存在", ErrMedicalNotFound)
		}
		return nil, err
	}

	var rows []visitQueueListRow
	if err := database.DB.Table("med_visit_queue AS visit_queue").
		Select("visit_queue.*, registration.patient_no, registration.patient_name, registration.patient_phone, registration.registration_no, registration.start_time, registration.end_time").
		Joins("JOIN med_registration AS registration ON registration.registration_id = visit_queue.registration_id").
		Where("visit_queue.schedule_id = ?", scheduleID).
		Order("visit_queue.queue_sequence ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.VisitQueueListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.VisitQueueListItemResponse{
			QueueID: row.QueueID, QueueSequence: row.QueueSequence, QueueStatus: row.QueueStatus, CallCount: row.CallCount,
			PatientNo: row.PatientNo, PatientName: maskPatientName(row.PatientName), PatientPhone: maskPatientPhone(row.PatientPhone),
			RegistrationNo: row.RegistrationNo, StartTime: trimScheduleTime(row.StartTime), EndTime: trimScheduleTime(row.EndTime),
			CheckInTime: row.CreateDate.In(medicalBusinessLocation).Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}

func (s *MedicalRegistrationService) getRegistrationDetail(db *gorm.DB, registrationID string, showSensitive, includeQueue bool) (*models.RegistrationResponse, error) {
	if err := validateMedicalUUID(registrationID, "挂号单ID"); err != nil {
		return nil, err
	}
	var registration models.MedRegistration
	if err := db.Where("registration_id = ?", registrationID).First(&registration).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 挂号单不存在", ErrMedicalNotFound)
		}
		return nil, err
	}
	response := registrationToResponse(registration, showSensitive)
	lifecycles, err := loadRegistrationLogs(db, registrationID)
	if err != nil {
		return nil, err
	}
	response.LifecycleRecords = lifecycles
	if includeQueue {
		queueInfo, err := loadVisitQueue(db, registrationID)
		if err != nil {
			return nil, err
		}
		response.QueueInfo = queueInfo
	}
	return response, nil
}

func (s *MedicalRegistrationService) CreateRegistration(req models.CreateRegistrationRequest, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	patientID := strings.TrimSpace(req.PatientID)
	slotID := strings.TrimSpace(req.SlotID)
	if err := validateMedicalUUID(patientID, "患者ID"); err != nil {
		return nil, err
	}
	if err := validateMedicalUUID(slotID, "号源时段ID"); err != nil {
		return nil, err
	}
	if req.RegistrationMethod == nil || !validRegistrationMethod(*req.RegistrationMethod) {
		return nil, fmt.Errorf("%w: 挂号方式不正确", ErrMedicalInvalidInput)
	}
	remark, err := normalizeRegistrationText(req.Remark, 512, "备注")
	if err != nil {
		return nil, err
	}

	var response *models.RegistrationResponse
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var patient models.MedPatient
		if err := tx.Where("patient_id = ? AND status = 1 AND del_flag = 0", patientID).First(&patient).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 患者不存在或已停用", ErrMedicalInvalidInput)
			}
			return err
		}
		var slot models.MedScheduleSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("slot_id = ? AND del_flag = 0", slotID).First(&slot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 号源时段不存在", ErrMedicalNotFound)
			}
			return err
		}
		var schedule models.MedSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("schedule_id = ? AND del_flag = 0", slot.ScheduleID).First(&schedule).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 排班不存在", ErrMedicalNotFound)
			}
			return err
		}
		startAt, parseErr := scheduleDateTime(schedule.ScheduleDate, slot.StartTime)
		if parseErr != nil {
			return parseErr
		}
		if schedule.Status != models.MedScheduleStatusPublished || !startAt.After(time.Now().In(medicalBusinessLocation)) || slot.BookedQuota >= slot.Quota {
			return fmt.Errorf("%w: 当前号源不可预约", ErrMedicalConflict)
		}
		if schedule.FeeAmount == nil || strings.TrimSpace(*schedule.FeeAmount) == "" {
			return fmt.Errorf("%w: 排班费用快照缺失", ErrMedicalConflict)
		}
		var duplicateCount int64
		if err := tx.Model(&models.MedRegistration{}).
			Where("patient_id = ? AND slot_id = ? AND status NOT IN ?", patientID, slotID, []int{models.MedRegistrationStatusCanceled, models.MedRegistrationStatusRefunded}).
			Count(&duplicateCount).Error; err != nil {
			return err
		}
		if duplicateCount > 0 {
			return fmt.Errorf("%w: 该患者已占用当前号源", ErrMedicalConflict)
		}

		var doctor models.MedDoctor
		if err := tx.Select("doctor_id", "name").Where("doctor_id = ? AND del_flag = 0", schedule.DoctorID).First(&doctor).Error; err != nil {
			return err
		}
		var department models.MedDepartment
		if err := tx.Select("department_id", "department_name").Where("department_id = ? AND del_flag = 0", schedule.DepartmentID).First(&department).Error; err != nil {
			return err
		}
		var registrationType models.SysDict
		if err := tx.Select("label").Where("type = ? AND value = ? AND status = 1 AND del_flag = 0", registrationTypeDictType, schedule.RegistrationType).First(&registrationType).Error; err != nil {
			return fmt.Errorf("%w: 挂号类型字典值不存在或已停用", ErrMedicalConflict)
		}
		registrationNo, err := s.codeSequenceService.NextBusinessCode(tx, "REGISTRATION", "REG", 6)
		if err != nil {
			return err
		}
		now := time.Now()
		registrationID := utils.GenerateUUID()
		registration := models.MedRegistration{
			RegistrationID: registrationID, RegistrationNo: registrationNo,
			PatientID: patient.PatientID, PatientNo: patient.PatientNo, PatientName: patient.Name, PatientGender: patient.Gender,
			PatientBirthDate: patient.BirthDate, PatientIDType: patient.IDType, PatientIDNumber: patient.IDNumber, PatientPhone: patient.Phone,
			ScheduleID: schedule.ScheduleID, SlotID: slot.SlotID, DoctorID: schedule.DoctorID, DoctorName: doctor.Name,
			DepartmentID: schedule.DepartmentID, DepartmentName: department.DepartmentName, ScheduleDate: schedule.ScheduleDate,
			StartTime: slot.StartTime, EndTime: slot.EndTime, RegistrationType: schedule.RegistrationType,
			RegistrationTypeName: strings.TrimSpace(stringValue(registrationType.Label)), RegistrationMethod: *req.RegistrationMethod,
			FeeRuleID: schedule.FeeRuleID, FeeRuleVersion: schedule.FeeRuleVersion, FeeAmount: *schedule.FeeAmount,
			Status: models.MedRegistrationStatusPendingPayment, Remark: remark, CreatorID: optionalOperatorID(operatorID),
			UpdaterID: optionalOperatorID(operatorID), CreateDate: &now, UpdateDate: &now,
		}
		if err := tx.Create(&registration).Error; err != nil {
			return err
		}
		if err := tx.Model(&slot).Updates(map[string]interface{}{"booked_quota": gorm.Expr("booked_quota + 1"), "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&schedule).Updates(map[string]interface{}{"booked_quota": gorm.Expr("booked_quota + 1"), "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error; err != nil {
			return err
		}
		if err := createRegistrationLog(tx, registrationID, nil, models.MedRegistrationStatusPendingPayment, operatorID, now, nil, nil); err != nil {
			return err
		}
		response, err = s.getRegistrationDetail(tx, registrationID, showSensitive, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *MedicalRegistrationService) ConfirmPayment(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusPaid, operatorID, nil, false, showSensitive)
}
func (s *MedicalRegistrationService) Cancel(id, operatorID, reason string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusCanceled, operatorID, &reason, false, showSensitive)
}
func (s *MedicalRegistrationService) CheckIn(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusCheckedIn, operatorID, nil, false, showSensitive)
}
func (s *MedicalRegistrationService) Complete(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusCompleted, operatorID, nil, false, showSensitive)
}
func (s *MedicalRegistrationService) NoShow(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusNoShow, operatorID, nil, true, showSensitive)
}
func (s *MedicalRegistrationService) StartRefund(id, operatorID, reason string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusRefundStarted, operatorID, &reason, false, showSensitive)
}
func (s *MedicalRegistrationService) ProcessRefund(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusRefunding, operatorID, nil, false, showSensitive)
}
func (s *MedicalRegistrationService) CompleteRefund(id, operatorID string, showSensitive bool) (*models.RegistrationResponse, error) {
	return s.transition(id, models.MedRegistrationStatusRefunded, operatorID, nil, false, showSensitive)
}

func (s *MedicalRegistrationService) transition(id string, toStatus int, operatorID string, reason *string, requireSlotEnded, showSensitive bool) (*models.RegistrationResponse, error) {
	if err := validateMedicalUUID(id, "挂号单ID"); err != nil {
		return nil, err
	}
	normalizedReason, err := normalizeRegistrationText(reason, 512, "原因")
	if err != nil {
		return nil, err
	}
	if (toStatus == models.MedRegistrationStatusCanceled || toStatus == models.MedRegistrationStatusRefundStarted) && normalizedReason == nil {
		return nil, fmt.Errorf("%w: 原因不能为空", ErrMedicalInvalidInput)
	}
	var response *models.RegistrationResponse
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var registration models.MedRegistration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("registration_id = ?", id).First(&registration).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 挂号单不存在", ErrMedicalNotFound)
			}
			return err
		}
		if err := ValidateRegistrationTransition(registration.Status, toStatus); err != nil {
			return err
		}
		if requireSlotEnded {
			endAt, err := scheduleDateTime(registration.ScheduleDate, registration.EndTime)
			if err != nil {
				return err
			}
			if time.Now().In(medicalBusinessLocation).Before(endAt) {
				return fmt.Errorf("%w: 号源结束后才能标记爽约", ErrMedicalConflict)
			}
		}
		now := time.Now()
		if RegistrationTransitionReleasesQuota(registration.Status, toStatus) {
			if err := releaseRegistrationQuota(tx, registration, operatorID, now); err != nil {
				return err
			}
		}
		if toStatus == models.MedRegistrationStatusCheckedIn {
			if err := createVisitQueue(tx, registration, operatorID, now); err != nil {
				return err
			}
		}
		if toStatus == models.MedRegistrationStatusCompleted {
			if err := completeVisitQueue(tx, registration.RegistrationID); err != nil {
				return err
			}
		}
		fromStatus := registration.Status
		if err := tx.Model(&registration).Updates(map[string]interface{}{"status": toStatus, "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error; err != nil {
			return err
		}
		var refundAmount *string
		if toStatus == models.MedRegistrationStatusRefundStarted {
			refundAmount = &registration.FeeAmount
		}
		if err := createRegistrationLog(tx, id, &fromStatus, toStatus, operatorID, now, normalizedReason, refundAmount); err != nil {
			return err
		}
		response, err = s.getRegistrationDetail(tx, id, showSensitive, toStatus == models.MedRegistrationStatusCheckedIn)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func releaseRegistrationQuota(tx *gorm.DB, registration models.MedRegistration, operatorID string, now time.Time) error {
	var slot models.MedScheduleSlot
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("slot_id = ? AND del_flag = 0", registration.SlotID).First(&slot).Error; err != nil {
		return err
	}
	var schedule models.MedSchedule
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("schedule_id = ? AND del_flag = 0", registration.ScheduleID).First(&schedule).Error; err != nil {
		return err
	}
	if slot.BookedQuota <= 0 || schedule.BookedQuota <= 0 {
		return fmt.Errorf("%w: 号源占用数量异常", ErrMedicalConflict)
	}
	if err := tx.Model(&slot).Updates(map[string]interface{}{"booked_quota": gorm.Expr("booked_quota - 1"), "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error; err != nil {
		return err
	}
	return tx.Model(&schedule).Updates(map[string]interface{}{"booked_quota": gorm.Expr("booked_quota - 1"), "updater_id": optionalOperatorID(operatorID), "update_date": now}).Error
}

func createRegistrationLog(tx *gorm.DB, registrationID string, fromStatus *int, toStatus int, operatorID string, operatedAt time.Time, reason, refundAmount *string) error {
	return tx.Create(&models.MedRegistrationLog{LogID: utils.GenerateUUID(), RegistrationID: registrationID, FromStatus: fromStatus, ToStatus: toStatus, OperatorID: optionalOperatorID(operatorID), OperatedAt: operatedAt, Reason: reason, RefundAmount: refundAmount}).Error
}

func createVisitQueue(tx *gorm.DB, registration models.MedRegistration, operatorID string, createdAt time.Time) error {
	var schedule models.MedSchedule
	if err := tx.Select("schedule_id").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("schedule_id = ? AND del_flag = 0", registration.ScheduleID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 排班不存在", ErrMedicalNotFound)
		}
		return err
	}

	var existing models.MedVisitQueue
	if err := tx.Select("queue_id").Where("registration_id = ?", registration.RegistrationID).First(&existing).Error; err == nil {
		return fmt.Errorf("%w: 挂号单已生成候诊序号", ErrMedicalConflict)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var maxSequence int
	if err := tx.Model(&models.MedVisitQueue{}).
		Where("schedule_id = ?", registration.ScheduleID).
		Select("COALESCE(MAX(queue_sequence), 0)").Scan(&maxSequence).Error; err != nil {
		return err
	}

	queue := models.MedVisitQueue{
		QueueID:        utils.GenerateUUID(),
		RegistrationID: registration.RegistrationID,
		ScheduleID:     registration.ScheduleID,
		QueueSequence:  maxSequence + 1,
		QueueStatus:    models.MedVisitQueueStatusWaiting,
		CallCount:      0,
		CreateDate:     createdAt,
		CreatorID:      optionalOperatorID(operatorID),
	}
	return tx.Create(&queue).Error
}

func completeVisitQueue(tx *gorm.DB, registrationID string) error {
	result := tx.Model(&models.MedVisitQueue{}).
		Where("registration_id = ? AND queue_status = ?", registrationID, models.MedVisitQueueStatusWaiting).
		Update("queue_status", models.MedVisitQueueStatusCompleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("%w: 候诊记录不存在或状态不允许完成", ErrMedicalConflict)
	}
	return nil
}

func loadRegistrationLogs(db *gorm.DB, registrationID string) ([]models.RegistrationLifecycleResponse, error) {
	var rows []registrationLogRow
	err := db.Table("med_registration_log AS registration_log").
		Select("registration_log.*, COALESCE(user.real_name, user.username) AS operator_name").
		Joins("LEFT JOIN sys_user AS user ON user.user_id = registration_log.operator_id AND user.del_flag = 0").
		Where("registration_log.registration_id = ?", registrationID).Order("registration_log.operated_at ASC, registration_log.log_id ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]models.RegistrationLifecycleResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.RegistrationLifecycleResponse{LifecycleID: row.LogID, FromStatus: row.FromStatus, ToStatus: row.ToStatus, OperatorID: row.OperatorID, OperatorName: row.OperatorName, OperatedAt: row.OperatedAt.In(medicalBusinessLocation).Format("2006-01-02 15:04:05"), Reason: row.Reason, RefundAmount: row.RefundAmount})
	}
	return result, nil
}

func loadVisitQueue(db *gorm.DB, registrationID string) (*models.VisitQueueResponse, error) {
	var queue models.MedVisitQueue
	if err := db.Where("registration_id = ?", registrationID).First(&queue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &models.VisitQueueResponse{
		QueueID:       queue.QueueID,
		QueueSequence: queue.QueueSequence,
		QueueStatus:   queue.QueueStatus,
		CallCount:     queue.CallCount,
		CreateDate:    queue.CreateDate.In(medicalBusinessLocation).Format("2006-01-02 15:04:05"),
		CreatorID:     queue.CreatorID,
	}, nil
}

func registrationToResponse(value models.MedRegistration, showSensitive bool) *models.RegistrationResponse {
	response := &models.RegistrationResponse{
		RegistrationID: value.RegistrationID, RegistrationNo: value.RegistrationNo, PatientID: value.PatientID, PatientNo: value.PatientNo,
		PatientName: value.PatientName, PatientGender: value.PatientGender, PatientBirthDate: value.PatientBirthDate.Format("2006-01-02"),
		PatientIDType: value.PatientIDType, PatientIDNumber: value.PatientIDNumber, PatientPhone: value.PatientPhone,
		ScheduleID: value.ScheduleID, SlotID: value.SlotID, DoctorID: value.DoctorID, DoctorName: value.DoctorName,
		DepartmentID: value.DepartmentID, DepartmentName: value.DepartmentName, ScheduleDate: value.ScheduleDate.Format("2006-01-02"),
		StartTime: value.StartTime, EndTime: value.EndTime, RegistrationType: value.RegistrationType, RegistrationTypeName: value.RegistrationTypeName,
		RegistrationMethod: value.RegistrationMethod, FeeRuleID: value.FeeRuleID, FeeRuleVersion: value.FeeRuleVersion, FeeAmount: value.FeeAmount,
		Status: value.Status, Remark: value.Remark, CreatorID: value.CreatorID, CreateDate: formatMedicalTime(value.CreateDate), UpdateDate: formatMedicalTime(value.UpdateDate),
		LifecycleRecords: []models.RegistrationLifecycleResponse{},
	}
	if !showSensitive {
		response.PatientName = maskPatientName(response.PatientName)
		response.PatientIDNumber = maskPatientIDNumber(response.PatientIDNumber)
		response.PatientPhone = maskPatientPhone(response.PatientPhone)
	}
	return response
}

func parseMedicalQueryDate(value, label string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), medicalBusinessLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s格式必须为YYYY-MM-DD", ErrMedicalInvalidInput, label)
	}
	return parsed, nil
}

func normalizeRegistrationText(value *string, max int, label string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, nil
	}
	if len([]rune(normalized)) > max {
		return nil, fmt.Errorf("%w: %s不能超过%d个字符", ErrMedicalInvalidInput, label, max)
	}
	return &normalized, nil
}

func validRegistrationMethod(value int) bool {
	return value == models.MedRegistrationMethodOnSite || value == models.MedRegistrationMethodAppointment
}

func validRegistrationStatus(value int) bool {
	switch value {
	case models.MedRegistrationStatusPendingPayment, models.MedRegistrationStatusPaid, models.MedRegistrationStatusCheckedIn,
		models.MedRegistrationStatusCompleted, models.MedRegistrationStatusCanceled, models.MedRegistrationStatusNoShow,
		models.MedRegistrationStatusRefundStarted, models.MedRegistrationStatusRefunding, models.MedRegistrationStatusRefunded:
		return true
	default:
		return false
	}
}

func formatMedicalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.In(medicalBusinessLocation).Format("2006-01-02 15:04:05")
	return &formatted
}

package services

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const prescriptionSequenceType = "MED_PRESCRIPTION"

var medicationFrequencyMultiplier = map[string]float64{
	"QD": 1, "BID": 2, "TID": 3, "QID": 4, "QAM": 1, "QN": 1,
	"Q8H": 3, "Q12H": 2, "QOD": 0.5, "QW": 1.0 / 7, "STAT": 0,
}

type MedicalPrescriptionService struct {
	codeSequenceService *BaseCodeSequenceService
}

type prescriptionListRow struct {
	models.MedPrescription
	PatientNo      string `gorm:"column:patient_no"`
	PatientName    string `gorm:"column:patient_name"`
	DoctorName     string `gorm:"column:doctor_name"`
	DepartmentName string `gorm:"column:department_name"`
	RegistrationNo string `gorm:"column:registration_no"`
}

type drugSnapshotRow struct {
	SkuID           string  `gorm:"column:sku_id"`
	SkuCode         string  `gorm:"column:sku_code"`
	ProductName     string  `gorm:"column:product_name"`
	ProductType     string  `gorm:"column:product_type"`
	SpecName        string  `gorm:"column:spec_name"`
	DosageForm      *string `gorm:"column:dosage_form"`
	EnterpriseName  string  `gorm:"column:enterprise_name"`
	ApprovalNo      string  `gorm:"column:approval_no"`
	PackageSpecName string  `gorm:"column:package_spec_name"`
	PackConversion  int     `gorm:"column:pack_conversion"`
	MinUnitName     string  `gorm:"column:min_unit_name"`
	PackageUnitName string  `gorm:"column:package_unit_name"`
	AllowSplit      int     `gorm:"column:allow_split"`
}

func NewMedicalPrescriptionService() *MedicalPrescriptionService {
	return &MedicalPrescriptionService{codeSequenceService: NewBaseCodeSequenceService()}
}

func (s *MedicalPrescriptionService) CreatePrescription(recordID, userID string, req models.SavePrescriptionRequest) (*models.PrescriptionResponse, error) {
	if err := validateMedicalUUID(recordID, "门诊病历ID"); err != nil {
		return nil, err
	}
	var prescriptionID string
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, record, err := lockEditableRecord(tx, recordID, userID, false)
		if err != nil {
			return err
		}
		if req.PrescriptionType != models.MedPrescriptionTypeOrdinary {
			return fmt.Errorf("%w: 首期只支持普通处方", ErrMedicalInvalidInput)
		}
		prescriptionNo, err := s.codeSequenceService.NextBusinessCode(tx, prescriptionSequenceType, "RX", 8)
		if err != nil {
			return err
		}
		now := time.Now()
		prescriptionID = utils.GenerateUUID()
		prescription := models.MedPrescription{
			PrescriptionID: prescriptionID, PrescriptionNo: prescriptionNo, RecordID: record.RecordID,
			RegistrationID: record.RegistrationID, PatientID: record.PatientID, DoctorID: doctor.DoctorID,
			PrescriptionType: req.PrescriptionType, Status: models.MedPrescriptionStatusDraft,
			Remark: normalizeMedicalOptionalString(req.Remark), CreatorID: optionalOperatorID(userID), UpdaterID: optionalOperatorID(userID),
			CreateDate: &now, UpdateDate: &now,
		}
		if err := tx.Create(&prescription).Error; err != nil {
			return err
		}
		return replacePrescriptionItems(tx, prescriptionID, userID, req.Items)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPrescription(prescriptionID)
}

func (s *MedicalPrescriptionService) UpdatePrescription(id, userID string, req models.SavePrescriptionRequest) (*models.PrescriptionResponse, error) {
	if err := validateMedicalUUID(id, "处方ID"); err != nil {
		return nil, err
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		var prescription models.MedPrescription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("prescription_id = ? AND doctor_id = ?", id, doctor.DoctorID).First(&prescription).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 处方不存在或不属于当前医生", ErrMedicalNotFound)
			}
			return err
		}
		if prescription.Status != models.MedPrescriptionStatusDraft && prescription.Status != models.MedPrescriptionStatusRejected {
			return fmt.Errorf("%w: 当前处方状态不允许修改", ErrMedicalConflict)
		}
		if req.PrescriptionType != models.MedPrescriptionTypeOrdinary {
			return fmt.Errorf("%w: 首期只支持普通处方", ErrMedicalInvalidInput)
		}
		if err := tx.Model(&prescription).Updates(map[string]interface{}{
			"prescription_type": req.PrescriptionType, "remark": normalizeMedicalOptionalString(req.Remark),
			"updater_id": optionalOperatorID(userID), "update_date": time.Now(),
		}).Error; err != nil {
			return err
		}
		return replacePrescriptionItems(tx, id, userID, req.Items)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPrescription(id)
}

func (s *MedicalPrescriptionService) GetPrescription(id string) (*models.PrescriptionResponse, error) {
	return s.getPrescription(id, "", false)
}

func (s *MedicalPrescriptionService) GetDoctorPrescription(id, userID string) (*models.PrescriptionResponse, error) {
	doctor, err := requireWorkbenchDoctor(database.DB, userID)
	if err != nil {
		return nil, err
	}
	return s.getPrescription(id, doctor.DoctorID, false)
}

func (s *MedicalPrescriptionService) GetReviewPrescription(id string) (*models.PrescriptionResponse, error) {
	return s.getPrescription(id, "", true)
}

func (s *MedicalPrescriptionService) getPrescription(id, doctorID string, reviewOnly bool) (*models.PrescriptionResponse, error) {
	if err := validateMedicalUUID(id, "处方ID"); err != nil {
		return nil, err
	}
	query := database.DB.Where("prescription.prescription_id = ?", id)
	if doctorID != "" {
		query = query.Where("prescription.doctor_id = ?", doctorID)
	}
	if reviewOnly {
		query = query.Where("prescription.status IN ?", []int{
			models.MedPrescriptionStatusPendingReview,
			models.MedPrescriptionStatusApproved,
			models.MedPrescriptionStatusRejected,
		})
	}
	rows, err := queryPrescriptionRows(query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: 处方不存在", ErrMedicalNotFound)
	}
	items, err := buildPrescriptionResponses(database.DB, rows)
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (s *MedicalPrescriptionService) SubmitPrescription(id, userID string) (*models.PrescriptionResponse, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		doctor, err := requireWorkbenchDoctor(tx, userID)
		if err != nil {
			return err
		}
		var prescription models.MedPrescription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("prescription_id = ? AND doctor_id = ?", id, doctor.DoctorID).First(&prescription).Error; err != nil {
			return fmt.Errorf("%w: 处方不存在或不属于当前医生", ErrMedicalNotFound)
		}
		if prescription.Status != models.MedPrescriptionStatusDraft && prescription.Status != models.MedPrescriptionStatusRejected {
			return fmt.Errorf("%w: 当前处方状态不允许提交", ErrMedicalConflict)
		}
		var items []models.MedPrescriptionItem
		if err := tx.Where("prescription_id = ?", id).Order("sort asc").Find(&items).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("%w: 处方至少包含一条药品明细", ErrMedicalInvalidInput)
		}
		for _, item := range items {
			if _, err := loadEnabledDrugSnapshot(tx, item.SkuID); err != nil {
				return err
			}
		}
		var diagnoses []models.MedOutpatientDiagnosis
		if err := tx.Where("record_id = ?", prescription.RecordID).Order("is_primary desc, sort asc").Find(&diagnoses).Error; err != nil {
			return err
		}
		primaryCount := 0
		for _, diagnosis := range diagnoses {
			if diagnosis.IsPrimary == 1 {
				primaryCount++
			}
		}
		if primaryCount != 1 {
			return fmt.Errorf("%w: 提交处方前必须且只能选择一个主要诊断", ErrMedicalInvalidInput)
		}
		var record models.MedOutpatientRecord
		if err := tx.Where("record_id = ?", prescription.RecordID).First(&record).Error; err != nil {
			return err
		}
		version := prescription.CurrentVersion + 1
		now := time.Now()
		submissionID := utils.GenerateUUID()
		submission := models.MedPrescriptionSubmission{
			SubmissionID: submissionID, PrescriptionID: id, Version: version,
			SubmissionStatus: models.MedPrescriptionSubmissionPending, AllergyHistory: record.AllergyHistory,
			SubmittedBy: userID, SubmittedAt: now,
		}
		if err := tx.Create(&submission).Error; err != nil {
			return err
		}
		snapshotItems := make([]models.MedPrescriptionSubmissionItem, 0, len(items))
		for _, item := range items {
			snapshotItems = append(snapshotItems, prescriptionItemSnapshot(submissionID, item))
		}
		if err := tx.Create(&snapshotItems).Error; err != nil {
			return err
		}
		snapshotDiagnoses := make([]models.MedPrescriptionSubmissionDiagnosis, 0, len(diagnoses))
		for _, diagnosis := range diagnoses {
			snapshotDiagnoses = append(snapshotDiagnoses, models.MedPrescriptionSubmissionDiagnosis{
				SubmissionDiagnosisID: utils.GenerateUUID(), SubmissionID: submissionID, DiagnosisID: diagnosis.DiagnosisID,
				ICDCode: diagnosis.ICDCode, ICDName: diagnosis.ICDName, IsPrimary: diagnosis.IsPrimary, Sort: diagnosis.Sort,
			})
		}
		if err := tx.Create(&snapshotDiagnoses).Error; err != nil {
			return err
		}
		return tx.Model(&prescription).Updates(map[string]interface{}{
			"status": models.MedPrescriptionStatusPendingReview, "current_version": version,
			"updater_id": optionalOperatorID(userID), "update_date": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetPrescription(id)
}

func (s *MedicalPrescriptionService) WithdrawPrescription(id, userID string) (*models.PrescriptionResponse, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		prescription, submission, err := lockPendingPrescription(tx, id)
		if err != nil {
			return err
		}
		if prescription.CreatorID == nil || *prescription.CreatorID != userID {
			return fmt.Errorf("%w: 只能撤回本人提交的处方", ErrMedicalConflict)
		}
		now := time.Now()
		if err := tx.Model(submission).Update("submission_status", models.MedPrescriptionSubmissionWithdrawn).Error; err != nil {
			return err
		}
		return tx.Model(prescription).Updates(map[string]interface{}{"status": models.MedPrescriptionStatusDraft, "updater_id": optionalOperatorID(userID), "update_date": now}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetPrescription(id)
}

func (s *MedicalPrescriptionService) VoidPrescription(id, userID string) (*models.PrescriptionResponse, error) {
	if err := validateMedicalUUID(id, "处方ID"); err != nil {
		return nil, err
	}
	doctor, err := requireWorkbenchDoctor(database.DB, userID)
	if err != nil {
		return nil, err
	}
	result := database.DB.Model(&models.MedPrescription{}).
		Where("prescription_id = ? AND doctor_id = ? AND status IN ?", id, doctor.DoctorID, []int{models.MedPrescriptionStatusDraft, models.MedPrescriptionStatusRejected}).
		Updates(map[string]interface{}{"status": models.MedPrescriptionStatusVoided, "updater_id": optionalOperatorID(userID), "update_date": time.Now()})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("%w: 处方不存在或当前状态不允许作废", ErrMedicalConflict)
	}
	return s.GetPrescription(id)
}

func (s *MedicalPrescriptionService) GetReviewList(req models.PrescriptionListRequest) (*utils.PageResult, error) {
	query := prescriptionBaseQuery(database.DB)
	if value := strings.TrimSpace(req.PrescriptionNo); value != "" {
		query = query.Where("prescription.prescription_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(req.PatientKeyword); value != "" {
		like := "%" + value + "%"
		query = query.Where("registration.patient_no LIKE ? OR registration.patient_name LIKE ?", like, like)
	}
	if req.Status != nil {
		if *req.Status != models.MedPrescriptionStatusPendingReview &&
			*req.Status != models.MedPrescriptionStatusApproved &&
			*req.Status != models.MedPrescriptionStatusRejected {
			return nil, fmt.Errorf("%w: 审核列表状态只允许10、20或30", ErrMedicalInvalidInput)
		}
		query = query.Where("prescription.status = ?", *req.Status)
	} else {
		query = query.Where("prescription.status IN ?", []int{models.MedPrescriptionStatusPendingReview, models.MedPrescriptionStatusApproved, models.MedPrescriptionStatusRejected})
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"prescriptionNo": "prescription.prescription_no", "patientName": "registration.patient_name",
		"doctorName": "registration.doctor_name", "status": "prescription.status", "createDate": "prescription.create_date",
	})
	if order == "" {
		order = "prescription.update_date desc"
	}
	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var rows []prescriptionListRow
	page, err := utils.Paginate(query.Order(order), req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	items, err := buildPrescriptionResponses(database.DB, rows)
	if err != nil {
		return nil, err
	}
	page.Items = items
	return page, nil
}

func (s *MedicalPrescriptionService) ReviewPrescription(id, reviewerID string, req models.ReviewPrescriptionRequest) (*models.PrescriptionResponse, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		prescription, submission, err := lockPendingPrescription(tx, id)
		if err != nil {
			return err
		}
		if submission.SubmittedBy == reviewerID {
			return fmt.Errorf("%w: 审核人不能审核本人开具的处方", ErrMedicalConflict)
		}
		opinion := normalizeMedicalOptionalString(req.Opinion)
		if req.Approved == 0 && opinion == nil {
			return fmt.Errorf("%w: 审核不通过时必须填写审核意见", ErrMedicalInvalidInput)
		}
		now := time.Now()
		submissionStatus := models.MedPrescriptionSubmissionRejected
		prescriptionStatus := models.MedPrescriptionStatusRejected
		if req.Approved == 1 {
			submissionStatus = models.MedPrescriptionSubmissionApproved
			prescriptionStatus = models.MedPrescriptionStatusApproved
		}
		if err := tx.Model(submission).Updates(map[string]interface{}{
			"submission_status": submissionStatus, "reviewer_id": reviewerID,
			"review_opinion": opinion, "reviewed_at": now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(prescription).Updates(map[string]interface{}{
			"status": prescriptionStatus, "updater_id": reviewerID, "update_date": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetPrescription(id)
}

func lockEditableRecord(tx *gorm.DB, recordID, userID string, allowCompleted bool) (*models.MedDoctor, *models.MedOutpatientRecord, error) {
	doctor, err := requireWorkbenchDoctor(tx, userID)
	if err != nil {
		return nil, nil, err
	}
	var record models.MedOutpatientRecord
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("record_id = ? AND doctor_id = ?", recordID, doctor.DoctorID).First(&record).Error; err != nil {
		return nil, nil, fmt.Errorf("%w: 门诊病历不存在或不属于当前医生", ErrMedicalNotFound)
	}
	if !allowCompleted && record.EndDate != nil {
		return nil, nil, fmt.Errorf("%w: 已完成接诊不能新建处方", ErrMedicalConflict)
	}
	return doctor, &record, nil
}

func replacePrescriptionItems(tx *gorm.DB, prescriptionID, userID string, requests []models.SavePrescriptionItemRequest) error {
	seen := make(map[string]struct{}, len(requests))
	rows := make([]models.MedPrescriptionItem, 0, len(requests))
	now := time.Now()
	for index, request := range requests {
		if _, exists := seen[request.SkuID]; exists {
			return fmt.Errorf("%w: 同一张处方不能重复选择相同药品", ErrMedicalInvalidInput)
		}
		seen[request.SkuID] = struct{}{}
		snapshot, err := loadEnabledDrugSnapshot(tx, request.SkuID)
		if err != nil {
			return err
		}
		if err := validateMedicationRoute(tx, request.MedicationRoute); err != nil {
			return err
		}
		singleDose, err := parsePositiveDecimal3(request.SingleDose, "单次用量")
		if err != nil {
			return err
		}
		totalMinQuantity, dispenseQuantity, dispenseUnit, err := calculateDispenseQuantity(singleDose, request.Frequency, request.CourseDays, request.TotalMinQuantity, *snapshot)
		if err != nil {
			return err
		}
		rows = append(rows, models.MedPrescriptionItem{
			ItemID: utils.GenerateUUID(), PrescriptionID: prescriptionID, SkuID: snapshot.SkuID, SkuCode: snapshot.SkuCode,
			ProductName: snapshot.ProductName, SpecName: snapshot.SpecName, DosageForm: snapshot.DosageForm,
			EnterpriseName: snapshot.EnterpriseName, ApprovalNo: snapshot.ApprovalNo, PackageSpecName: snapshot.PackageSpecName,
			PackConversion: snapshot.PackConversion, MinUnitName: snapshot.MinUnitName, PackageUnitName: snapshot.PackageUnitName,
			AllowSplit: snapshot.AllowSplit, SingleDose: formatDecimal3(singleDose), DoseUnit: snapshot.MinUnitName,
			MedicationRoute: strings.TrimSpace(request.MedicationRoute), Frequency: strings.ToUpper(strings.TrimSpace(request.Frequency)),
			CourseDays: request.CourseDays, TotalMinQuantity: totalMinQuantity, DispenseQuantity: dispenseQuantity, DispenseUnit: dispenseUnit,
			UsageInstructions: normalizeMedicalOptionalString(request.UsageInstructions), Remark: normalizeMedicalOptionalString(request.Remark),
			Sort: index, CreatorID: optionalOperatorID(userID), UpdaterID: optionalOperatorID(userID), CreateDate: &now, UpdateDate: &now,
		})
	}
	if err := tx.Where("prescription_id = ?", prescriptionID).Delete(&models.MedPrescriptionItem{}).Error; err != nil {
		return err
	}
	if len(rows) > 0 {
		return tx.Create(&rows).Error
	}
	return nil
}

func loadEnabledDrugSnapshot(tx *gorm.DB, skuID string) (*drugSnapshotRow, error) {
	if err := validateMedicalUUID(skuID, "药品SKU ID"); err != nil {
		return nil, err
	}
	var row drugSnapshotRow
	err := tx.Table("product_sku AS sku").
		Select("sku.sku_id, sku.sku_code, spu.product_name, spu.product_type, rp.spec_name, rp.dosage_form, enterprise.enterprise_name, mp.approval_no, sku.package_spec_name, sku.pack_conversion, sku.min_unit_name, sku.package_unit_name, sku.allow_split").
		Joins("JOIN product_mp AS mp ON mp.mp_id = sku.mp_id AND mp.status = 1 AND mp.del_flag = 0").
		Joins("JOIN product_rp AS rp ON rp.rp_id = mp.rp_id AND rp.status = 1 AND rp.del_flag = 0").
		Joins("JOIN product_spu AS spu ON spu.spu_id = rp.spu_id AND spu.status = 1 AND spu.del_flag = 0").
		Joins("JOIN base_enterprise AS enterprise ON enterprise.enterprise_id = mp.enterprise_id AND enterprise.status = 1 AND enterprise.del_flag = 0").
		Where("sku.sku_id = ? AND sku.status = 1 AND sku.del_flag = 0 AND spu.product_type = ?", skuID, models.ProductTypeDrug).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 药品不存在、已停用或档案链路不可用", ErrMedicalInvalidInput)
		}
		return nil, err
	}
	return &row, nil
}

func validateMedicationRoute(tx *gorm.DB, route string) error {
	value := strings.TrimSpace(route)
	if value == "" {
		return fmt.Errorf("%w: 给药途径不能为空", ErrMedicalInvalidInput)
	}
	var count int64
	if err := tx.Table("sys_dict AS item").Joins("JOIN sys_dict AS root ON root.id = item.pid AND root.status = 1 AND root.del_flag = 0").
		Where("item.type = ? AND item.value = ? AND item.status = 1 AND item.del_flag = 0", "MED_MEDICATION_ROUTE", value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: 给药途径不存在或已停用", ErrMedicalInvalidInput)
	}
	return nil
}

func calculateDispenseQuantity(singleDose float64, frequency string, courseDays int, manualTotal *string, snapshot drugSnapshotRow) (string, string, string, error) {
	frequency = strings.ToUpper(strings.TrimSpace(frequency))
	var total float64
	if frequency == "PRN" || frequency == "SOS" {
		if manualTotal == nil {
			return "", "", "", fmt.Errorf("%w: 必要时或需要时用药必须填写发药总量", ErrMedicalInvalidInput)
		}
		value, err := parsePositiveDecimal3(*manualTotal, "发药总量")
		if err != nil {
			return "", "", "", err
		}
		total = value
	} else if frequency == "STAT" {
		total = singleDose
	} else {
		multiplier, exists := medicationFrequencyMultiplier[frequency]
		if !exists {
			return "", "", "", fmt.Errorf("%w: 用药频次不正确", ErrMedicalInvalidInput)
		}
		total = singleDose * multiplier * float64(courseDays)
		if frequency == "QOD" || frequency == "QW" {
			total = singleDose * math.Ceil(multiplier*float64(courseDays))
		}
	}
	if total <= 0 || total > 99999999999 {
		return "", "", "", fmt.Errorf("%w: 发药数量超出允许范围", ErrMedicalInvalidInput)
	}
	totalText := formatDecimal3(total)
	if snapshot.AllowSplit == 1 {
		return totalText, totalText, snapshot.MinUnitName, nil
	}
	if snapshot.PackConversion <= 0 {
		return "", "", "", fmt.Errorf("%w: 药品包装换算系数不正确", ErrMedicalInvalidInput)
	}
	packages := math.Ceil(total / float64(snapshot.PackConversion))
	return totalText, formatDecimal3(packages), snapshot.PackageUnitName, nil
}

func parsePositiveDecimal3(value, field string) (float64, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts) == 0 || parts[0] == "" || (len(parts) == 2 && len(parts[1]) > 3) {
		return 0, fmt.Errorf("%w: %s最多保留三位小数", ErrMedicalInvalidInput, field)
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("%w: %s必须大于0", ErrMedicalInvalidInput, field)
	}
	return number, nil
}

func formatDecimal3(value float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 3, 64), "0"), ".")
}

func prescriptionItemSnapshot(submissionID string, item models.MedPrescriptionItem) models.MedPrescriptionSubmissionItem {
	return models.MedPrescriptionSubmissionItem{
		SubmissionItemID: utils.GenerateUUID(), SubmissionID: submissionID, SkuID: item.SkuID, SkuCode: item.SkuCode,
		ProductName: item.ProductName, SpecName: item.SpecName, DosageForm: item.DosageForm,
		EnterpriseName: item.EnterpriseName, ApprovalNo: item.ApprovalNo, PackageSpecName: item.PackageSpecName,
		PackConversion: item.PackConversion, MinUnitName: item.MinUnitName, PackageUnitName: item.PackageUnitName,
		AllowSplit: item.AllowSplit, SingleDose: item.SingleDose, DoseUnit: item.DoseUnit,
		MedicationRoute: item.MedicationRoute, Frequency: item.Frequency, CourseDays: item.CourseDays,
		TotalMinQuantity: item.TotalMinQuantity, DispenseQuantity: item.DispenseQuantity, DispenseUnit: item.DispenseUnit,
		UsageInstructions: item.UsageInstructions, Remark: item.Remark, Sort: item.Sort,
	}
}

func lockPendingPrescription(tx *gorm.DB, id string) (*models.MedPrescription, *models.MedPrescriptionSubmission, error) {
	if err := validateMedicalUUID(id, "处方ID"); err != nil {
		return nil, nil, err
	}
	var prescription models.MedPrescription
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("prescription_id = ?", id).First(&prescription).Error; err != nil {
		return nil, nil, fmt.Errorf("%w: 处方不存在", ErrMedicalNotFound)
	}
	if prescription.Status != models.MedPrescriptionStatusPendingReview {
		return nil, nil, fmt.Errorf("%w: 当前处方不是待审核状态", ErrMedicalConflict)
	}
	var submission models.MedPrescriptionSubmission
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("prescription_id = ? AND version = ? AND submission_status = ?", id, prescription.CurrentVersion, models.MedPrescriptionSubmissionPending).
		First(&submission).Error; err != nil {
		return nil, nil, fmt.Errorf("%w: 待审核提交版本不存在", ErrMedicalConflict)
	}
	return &prescription, &submission, nil
}

func prescriptionBaseQuery(db *gorm.DB) *gorm.DB {
	return db.Table("med_prescription AS prescription").
		Select("prescription.*, registration.patient_no, registration.patient_name, registration.doctor_name, registration.department_name, registration.registration_no").
		Joins("JOIN med_registration AS registration ON registration.registration_id = prescription.registration_id")
}

func queryPrescriptionRows(db *gorm.DB) ([]prescriptionListRow, error) {
	var rows []prescriptionListRow
	err := prescriptionBaseQuery(db).Order("prescription.create_date desc").Scan(&rows).Error
	return rows, err
}

func loadPrescriptionResponses(db *gorm.DB, recordID string) ([]models.PrescriptionResponse, error) {
	rows, err := queryPrescriptionRows(db.Where("prescription.record_id = ?", recordID))
	if err != nil {
		return nil, err
	}
	return buildPrescriptionResponses(db, rows)
}

func buildPrescriptionResponses(db *gorm.DB, rows []prescriptionListRow) ([]models.PrescriptionResponse, error) {
	result := make([]models.PrescriptionResponse, 0, len(rows))
	if len(rows) == 0 {
		return result, nil
	}
	prescriptionIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		prescriptionIDs = append(prescriptionIDs, row.PrescriptionID)
	}

	var items []models.MedPrescriptionItem
	if err := db.Where("prescription_id IN ?", prescriptionIDs).Order("prescription_id asc, sort asc").Find(&items).Error; err != nil {
		return nil, err
	}
	itemsByPrescription := make(map[string][]models.MedPrescriptionItem, len(rows))
	for _, prescriptionID := range prescriptionIDs {
		itemsByPrescription[prescriptionID] = make([]models.MedPrescriptionItem, 0)
	}
	for _, item := range items {
		itemsByPrescription[item.PrescriptionID] = append(itemsByPrescription[item.PrescriptionID], item)
	}

	var submissions []models.MedPrescriptionSubmission
	if err := db.Where("prescription_id IN ?", prescriptionIDs).Order("prescription_id asc, version desc").Find(&submissions).Error; err != nil {
		return nil, err
	}
	latestByPrescription := make(map[string]models.MedPrescriptionSubmission, len(rows))
	submissionIDs := make([]string, 0, len(rows))
	for _, submission := range submissions {
		if _, exists := latestByPrescription[submission.PrescriptionID]; exists {
			continue
		}
		latestByPrescription[submission.PrescriptionID] = submission
		submissionIDs = append(submissionIDs, submission.SubmissionID)
	}

	submissionItemsByID := make(map[string][]models.MedPrescriptionSubmissionItem, len(submissionIDs))
	submissionDiagnosesByID := make(map[string][]models.MedPrescriptionSubmissionDiagnosis, len(submissionIDs))
	for _, submissionID := range submissionIDs {
		submissionItemsByID[submissionID] = make([]models.MedPrescriptionSubmissionItem, 0)
		submissionDiagnosesByID[submissionID] = make([]models.MedPrescriptionSubmissionDiagnosis, 0)
	}
	if len(submissionIDs) > 0 {
		var submissionItems []models.MedPrescriptionSubmissionItem
		if err := db.Where("submission_id IN ?", submissionIDs).Order("submission_id asc, sort asc").Find(&submissionItems).Error; err != nil {
			return nil, err
		}
		for _, item := range submissionItems {
			submissionItemsByID[item.SubmissionID] = append(submissionItemsByID[item.SubmissionID], item)
		}
		var submissionDiagnoses []models.MedPrescriptionSubmissionDiagnosis
		if err := db.Where("submission_id IN ?", submissionIDs).Order("submission_id asc, is_primary desc, sort asc").Find(&submissionDiagnoses).Error; err != nil {
			return nil, err
		}
		for _, diagnosis := range submissionDiagnoses {
			submissionDiagnosesByID[diagnosis.SubmissionID] = append(submissionDiagnosesByID[diagnosis.SubmissionID], diagnosis)
		}
	}

	for _, row := range rows {
		response := models.PrescriptionResponse{
			MedPrescription: row.MedPrescription, PatientNo: row.PatientNo, PatientName: row.PatientName,
			DoctorName: row.DoctorName, DepartmentName: row.DepartmentName, RegistrationNo: row.RegistrationNo,
			Items: itemsByPrescription[row.PrescriptionID], SubmissionItems: []models.MedPrescriptionSubmissionItem{},
			SubmissionDiagnoses: []models.MedPrescriptionSubmissionDiagnosis{},
		}
		if submission, exists := latestByPrescription[row.PrescriptionID]; exists {
			response.LatestSubmission = &submission
			response.SubmissionItems = submissionItemsByID[submission.SubmissionID]
			response.SubmissionDiagnoses = submissionDiagnosesByID[submission.SubmissionID]
		}
		result = append(result, response)
	}
	return result, nil
}

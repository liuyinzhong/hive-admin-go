package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

type MedicalDiagnosisService struct{}

func NewMedicalDiagnosisService() *MedicalDiagnosisService {
	return &MedicalDiagnosisService{}
}

func (s *MedicalDiagnosisService) GetDiagnosisList(req models.DiagnosisListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.MedDiagnosis{}).Where("del_flag = 0")
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("icd_code LIKE ? OR icd_name LIKE ? OR name_pinyin LIKE ?", like, like, like)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"icdCode": "icd_code", "icdName": "icd_name", "sort": "sort",
		"status": "status", "createDate": "create_date", "updateDate": "update_date",
	})
	if order == "" {
		order = "sort asc, icd_code asc"
	}
	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var rows []models.MedDiagnosis
	page, err := utils.Paginate(query.Order(order), req.Page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	items := make([]models.DiagnosisResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, diagnosisToResponse(row))
	}
	page.Items = items
	return page, nil
}

func (s *MedicalDiagnosisService) GetDiagnosisOptions(req models.DiagnosisOptionRequest) ([]models.DiagnosisResponse, error) {
	limit := req.PageSize
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := database.DB.Model(&models.MedDiagnosis{}).Where("status = 1 AND del_flag = 0")
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("icd_code LIKE ? OR icd_name LIKE ? OR name_pinyin LIKE ?", like, like, like)
	}
	var rows []models.MedDiagnosis
	if err := query.Order("sort asc, icd_code asc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]models.DiagnosisResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, diagnosisToResponse(row))
	}
	return items, nil
}

func (s *MedicalDiagnosisService) GetDiagnosisDetail(id string) (*models.DiagnosisResponse, error) {
	row, err := loadDiagnosis(database.DB, id)
	if err != nil {
		return nil, err
	}
	result := diagnosisToResponse(*row)
	return &result, nil
}

func (s *MedicalDiagnosisService) CreateDiagnosis(req models.SaveDiagnosisRequest, operatorID string) error {
	prepared, err := prepareDiagnosis(req)
	if err != nil {
		return err
	}
	var count int64
	if err := database.DB.Model(&models.MedDiagnosis{}).
		Where("icd_code = ? AND del_flag = 0", prepared.ICDCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: ICD编码已存在", ErrMedicalConflict)
	}
	now := time.Now()
	prepared.DiagnosisID = utils.GenerateUUID()
	prepared.CreatorID = optionalOperatorID(operatorID)
	prepared.UpdaterID = optionalOperatorID(operatorID)
	prepared.CreateDate = &now
	prepared.UpdateDate = &now
	return database.DB.Create(prepared).Error
}

func (s *MedicalDiagnosisService) UpdateDiagnosis(id string, req models.SaveDiagnosisRequest, operatorID string) error {
	if err := validateMedicalUUID(id, "诊断ID"); err != nil {
		return err
	}
	prepared, err := prepareDiagnosis(req)
	if err != nil {
		return err
	}
	var count int64
	if err := database.DB.Model(&models.MedDiagnosis{}).
		Where("icd_code = ? AND diagnosis_id <> ? AND del_flag = 0", prepared.ICDCode, id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: ICD编码已存在", ErrMedicalConflict)
	}
	result := database.DB.Model(&models.MedDiagnosis{}).Where("diagnosis_id = ? AND del_flag = 0", id).Updates(map[string]interface{}{
		"icd_code": prepared.ICDCode, "icd_name": prepared.ICDName, "name_pinyin": prepared.NamePinyin,
		"status": prepared.Status, "sort": prepared.Sort, "remark": prepared.Remark,
		"updater_id": optionalOperatorID(operatorID), "update_date": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 疾病诊断不存在", ErrMedicalNotFound)
	}
	return nil
}

func (s *MedicalDiagnosisService) UpdateDiagnosisStatus(id string, status int, operatorID string) error {
	if err := validateMedicalUUID(id, "诊断ID"); err != nil {
		return err
	}
	result := database.DB.Model(&models.MedDiagnosis{}).Where("diagnosis_id = ? AND del_flag = 0", id).
		Updates(map[string]interface{}{"status": status, "updater_id": optionalOperatorID(operatorID), "update_date": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 疾病诊断不存在", ErrMedicalNotFound)
	}
	return nil
}

func (s *MedicalDiagnosisService) DeleteDiagnosis(id, operatorID string) error {
	if err := validateMedicalUUID(id, "诊断ID"); err != nil {
		return err
	}
	result := database.DB.Model(&models.MedDiagnosis{}).Where("diagnosis_id = ? AND del_flag = 0", id).
		Updates(map[string]interface{}{"del_flag": 1, "updater_id": optionalOperatorID(operatorID), "update_date": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 疾病诊断不存在", ErrMedicalNotFound)
	}
	return nil
}

func loadDiagnosis(db *gorm.DB, id string) (*models.MedDiagnosis, error) {
	if err := validateMedicalUUID(id, "诊断ID"); err != nil {
		return nil, err
	}
	var row models.MedDiagnosis
	if err := db.Where("diagnosis_id = ? AND del_flag = 0", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 疾病诊断不存在", ErrMedicalNotFound)
		}
		return nil, err
	}
	return &row, nil
}

func prepareDiagnosis(req models.SaveDiagnosisRequest) (*models.MedDiagnosis, error) {
	code := strings.ToUpper(strings.TrimSpace(req.ICDCode))
	name := strings.TrimSpace(req.ICDName)
	if code == "" || name == "" {
		return nil, fmt.Errorf("%w: ICD编码和名称不能为空", ErrMedicalInvalidInput)
	}
	return &models.MedDiagnosis{
		ICDCode: code, ICDName: name, NamePinyin: normalizeMedicalOptionalString(req.NamePinyin),
		Status: req.Status, Sort: req.Sort, Remark: normalizeMedicalOptionalString(req.Remark),
	}, nil
}

func diagnosisToResponse(row models.MedDiagnosis) models.DiagnosisResponse {
	return models.DiagnosisResponse{
		DiagnosisID: row.DiagnosisID, ICDCode: row.ICDCode, ICDName: row.ICDName,
		NamePinyin: row.NamePinyin, Status: row.Status, Sort: row.Sort, Remark: row.Remark,
		CreateDate: models.TimeToStringPtr(row.CreateDate), UpdateDate: models.TimeToStringPtr(row.UpdateDate),
	}
}

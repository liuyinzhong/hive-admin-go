package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrMedicalInvalidInput = errors.New("医疗管理参数错误")
	ErrMedicalNotFound     = errors.New("医疗管理数据不存在")
	ErrMedicalConflict     = errors.New("医疗管理数据冲突")
)

type MedicalDepartmentService struct{}

func NewMedicalDepartmentService() *MedicalDepartmentService {
	return &MedicalDepartmentService{}
}

func (s *MedicalDepartmentService) GetDepartmentTree(req models.MedicalDepartmentListRequest) ([]*models.MedicalDepartmentTreeResponse, error) {
	var departments []models.MedDepartment
	if err := database.DB.Where("del_flag = 0").Order("sort asc, create_date asc").Find(&departments).Error; err != nil {
		return nil, err
	}

	filtered := filterMedicalDepartments(departments, strings.TrimSpace(req.Keyword), req.Status)
	return buildMedicalDepartmentTree(filtered), nil
}

func (s *MedicalDepartmentService) GetAllDepartments() ([]*models.MedicalDepartmentTreeResponse, error) {
	var departments []models.MedDepartment
	if err := database.DB.Where("del_flag = 0 AND status = 1").Order("sort asc, create_date asc").Find(&departments).Error; err != nil {
		return nil, err
	}
	return buildMedicalDepartmentTree(departments), nil
}

func (s *MedicalDepartmentService) CreateDepartment(req models.CreateMedicalDepartmentRequest, operatorID string) error {
	if err := validateMedicalStatus(req.Status); err != nil {
		return err
	}
	if err := s.validateDepartmentParent(database.DB, req.Pid, ""); err != nil {
		return err
	}
	if err := s.ensureDepartmentUnique(database.DB, req.DepartmentCode, req.DepartmentName, ""); err != nil {
		return err
	}

	now := time.Now()
	department := models.MedDepartment{
		DepartmentID:   utils.GenerateUUID(),
		DepartmentCode: strings.TrimSpace(req.DepartmentCode),
		DepartmentName: strings.TrimSpace(req.DepartmentName),
		Pid:            normalizeMedicalOptionalString(req.Pid),
		Sort:           req.Sort,
		Status:         req.Status,
		Remark:         normalizeMedicalOptionalString(req.Remark),
		CreatorID:      optionalOperatorID(operatorID),
		UpdaterID:      optionalOperatorID(operatorID),
		CreateDate:     &now,
		UpdateDate:     &now,
		DelFlag:        0,
	}
	return database.DB.Create(&department).Error
}

func (s *MedicalDepartmentService) GetDepartmentDetail(departmentID string) (*models.MedicalDepartmentTreeResponse, error) {
	if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
		return nil, err
	}
	var department models.MedDepartment
	if err := database.DB.Where("department_id = ? AND del_flag = 0", departmentID).First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 临床科室不存在", ErrMedicalNotFound)
		}
		return nil, err
	}
	return medicalDepartmentToResponse(department), nil
}

func (s *MedicalDepartmentService) UpdateDepartment(departmentID string, req models.UpdateMedicalDepartmentRequest, operatorID string) error {
	if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
		return err
	}
	if err := validateMedicalStatus(req.Status); err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var department models.MedDepartment
		if err := tx.Where("department_id = ? AND del_flag = 0", departmentID).First(&department).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 临床科室不存在", ErrMedicalNotFound)
			}
			return err
		}
		if err := s.validateDepartmentParent(tx, req.Pid, departmentID); err != nil {
			return err
		}
		if err := s.ensureDepartmentUnique(tx, req.DepartmentCode, req.DepartmentName, departmentID); err != nil {
			return err
		}

		now := time.Now()
		updates := map[string]interface{}{
			"department_code": strings.TrimSpace(req.DepartmentCode),
			"department_name": strings.TrimSpace(req.DepartmentName),
			"pid":             normalizeMedicalOptionalString(req.Pid),
			"sort":            req.Sort,
			"status":          req.Status,
			"remark":          normalizeMedicalOptionalString(req.Remark),
			"updater_id":      optionalOperatorID(operatorID),
			"update_date":     now,
		}
		return tx.Model(&models.MedDepartment{}).Where("department_id = ?", departmentID).Updates(updates).Error
	})
}

func (s *MedicalDepartmentService) UpdateDepartmentStatus(departmentID string, status int, operatorID string) error {
	if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
		return err
	}
	if err := validateMedicalStatus(status); err != nil {
		return err
	}

	result := database.DB.Model(&models.MedDepartment{}).
		Where("department_id = ? AND del_flag = 0", departmentID).
		Updates(map[string]interface{}{"status": status, "updater_id": optionalOperatorID(operatorID), "update_date": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 临床科室不存在", ErrMedicalNotFound)
	}
	return nil
}

func (s *MedicalDepartmentService) DeleteDepartments(departmentIDs []string, operatorID string) error {
	if len(departmentIDs) == 0 {
		return fmt.Errorf("%w: 请选择要删除的临床科室", ErrMedicalInvalidInput)
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, departmentID := range departmentIDs {
			if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
				return err
			}
			var department models.MedDepartment
			if err := tx.Where("department_id = ? AND del_flag = 0", departmentID).First(&department).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("%w: 临床科室不存在", ErrMedicalNotFound)
				}
				return err
			}

			var childrenCount int64
			if err := tx.Model(&models.MedDepartment{}).Where("pid = ? AND del_flag = 0", departmentID).Count(&childrenCount).Error; err != nil {
				return err
			}
			if childrenCount > 0 {
				return fmt.Errorf("%w: 临床科室存在下级科室，不能删除", ErrMedicalConflict)
			}

			var doctorCount int64
			if err := tx.Model(&models.MedDoctorDepartment{}).
				Where("department_id = ? AND status = 1 AND del_flag = 0", departmentID).
				Count(&doctorCount).Error; err != nil {
				return err
			}
			if doctorCount > 0 {
				return fmt.Errorf("%w: 临床科室已关联医生，不能删除", ErrMedicalConflict)
			}

			now := time.Now()
			if err := tx.Model(&models.MedDepartment{}).Where("department_id = ?", departmentID).Updates(map[string]interface{}{
				"status":      0,
				"del_flag":    1,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *MedicalDepartmentService) ensureDepartmentUnique(tx *gorm.DB, code, name, excludeID string) error {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return fmt.Errorf("%w: 临床科室编码和名称不能为空", ErrMedicalInvalidInput)
	}

	query := tx.Model(&models.MedDepartment{}).Where("del_flag = 0 AND (department_code = ? OR department_name = ?)", code, name)
	if excludeID != "" {
		query = query.Where("department_id != ?", excludeID)
	}
	var departments []models.MedDepartment
	if err := query.Select("department_code", "department_name").Find(&departments).Error; err != nil {
		return err
	}
	for _, item := range departments {
		if item.DepartmentCode == code {
			return fmt.Errorf("%w: 临床科室编码已存在", ErrMedicalConflict)
		}
		if item.DepartmentName == name {
			return fmt.Errorf("%w: 临床科室名称已存在", ErrMedicalConflict)
		}
	}
	return nil
}

func (s *MedicalDepartmentService) validateDepartmentParent(tx *gorm.DB, pid *string, departmentID string) error {
	pid = normalizeMedicalOptionalString(pid)
	if pid == nil {
		return nil
	}
	if err := validateMedicalUUID(*pid, "上级临床科室ID"); err != nil {
		return err
	}
	if *pid == departmentID {
		return fmt.Errorf("%w: 上级临床科室不能选择自身", ErrMedicalInvalidInput)
	}

	var parent models.MedDepartment
	if err := tx.Where("department_id = ? AND del_flag = 0", *pid).First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 上级临床科室不存在", ErrMedicalNotFound)
		}
		return err
	}

	visited := map[string]struct{}{departmentID: {}}
	current := parent
	for {
		if _, exists := visited[current.DepartmentID]; exists {
			return fmt.Errorf("%w: 临床科室层级不能形成循环", ErrMedicalInvalidInput)
		}
		visited[current.DepartmentID] = struct{}{}
		if current.Pid == nil || *current.Pid == "" {
			break
		}
		if err := tx.Where("department_id = ? AND del_flag = 0", *current.Pid).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 上级临床科室层级不完整", ErrMedicalConflict)
			}
			return err
		}
	}
	return nil
}

func filterMedicalDepartments(departments []models.MedDepartment, keyword string, status *int) []models.MedDepartment {
	if keyword == "" && status == nil {
		return departments
	}
	byID := make(map[string]models.MedDepartment, len(departments))
	for _, department := range departments {
		byID[department.DepartmentID] = department
	}
	keep := make(map[string]struct{})
	keyword = strings.ToLower(keyword)
	for _, department := range departments {
		matchesKeyword := keyword == "" || strings.Contains(strings.ToLower(department.DepartmentName), keyword) || strings.Contains(strings.ToLower(department.DepartmentCode), keyword)
		matchesStatus := status == nil || department.Status == *status
		if !matchesKeyword || !matchesStatus {
			continue
		}
		current := department
		for {
			keep[current.DepartmentID] = struct{}{}
			if current.Pid == nil {
				break
			}
			parent, exists := byID[*current.Pid]
			if !exists {
				break
			}
			current = parent
		}
	}
	result := make([]models.MedDepartment, 0, len(keep))
	for _, department := range departments {
		if _, exists := keep[department.DepartmentID]; exists {
			result = append(result, department)
		}
	}
	return result
}

func buildMedicalDepartmentTree(departments []models.MedDepartment) []*models.MedicalDepartmentTreeResponse {
	departmentMap := make(map[string]*models.MedicalDepartmentTreeResponse, len(departments))
	for _, department := range departments {
		departmentMap[department.DepartmentID] = medicalDepartmentToResponse(department)
	}
	roots := make([]*models.MedicalDepartmentTreeResponse, 0)
	for _, department := range departments {
		node := departmentMap[department.DepartmentID]
		if department.Pid == nil || *department.Pid == "" {
			roots = append(roots, node)
			continue
		}
		if parent, exists := departmentMap[*department.Pid]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots
}

func medicalDepartmentToResponse(department models.MedDepartment) *models.MedicalDepartmentTreeResponse {
	return &models.MedicalDepartmentTreeResponse{
		DepartmentID:   department.DepartmentID,
		DepartmentCode: department.DepartmentCode,
		DepartmentName: department.DepartmentName,
		Pid:            department.Pid,
		Sort:           department.Sort,
		Status:         department.Status,
		Remark:         department.Remark,
		CreateDate:     models.TimeToStringPtr(department.CreateDate),
		UpdateDate:     models.TimeToStringPtr(department.UpdateDate),
		Children:       make([]*models.MedicalDepartmentTreeResponse, 0),
	}
}

func validateMedicalUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrMedicalInvalidInput, label)
	}
	return nil
}

func validateMedicalStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrMedicalInvalidInput)
	}
	return nil
}

func normalizeMedicalOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalOperatorID(operatorID string) *string {
	if operatorID == "" {
		return nil
	}
	return &operatorID
}

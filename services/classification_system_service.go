package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// 分类体系相关错误
var (
	ErrClassificationInvalidInput = errors.New("分类体系参数错误")
	ErrClassificationNotFound     = errors.New("分类体系数据不存在")
	ErrClassificationConflict     = errors.New("分类体系数据冲突")
)

// ClassificationSystemService 分类体系服务
type ClassificationSystemService struct{}

// NewClassificationSystemService 创建分类体系服务实例
func NewClassificationSystemService() *ClassificationSystemService {
	return &ClassificationSystemService{}
}

// GetAllSystems 查询全部分类体系，按 sort 升序、创建时间升序排列
func (s *ClassificationSystemService) GetAllSystems() ([]models.ClassificationSystemResponse, error) {
	var systems []models.BaseClassificationSystem
	if err := database.DB.Where("del_flag = 0").
		Order("sort asc, create_date asc").
		Find(&systems).Error; err != nil {
		return nil, err
	}
	responses := make([]models.ClassificationSystemResponse, 0, len(systems))
	for _, system := range systems {
		responses = append(responses, *systemToResponse(system))
	}
	return responses, nil
}

// GetSystemDetail 查询分类体系详情
func (s *ClassificationSystemService) GetSystemDetail(systemID string) (*models.ClassificationSystemResponse, error) {
	if err := validateClassificationUUID(systemID, "分类体系ID"); err != nil {
		return nil, err
	}
	system, err := s.getSystemByID(database.DB, systemID)
	if err != nil {
		return nil, err
	}
	return systemToResponse(*system), nil
}

// CreateSystem 创建分类体系
func (s *ClassificationSystemService) CreateSystem(req models.SaveClassificationSystemRequest, operatorID string) (*models.ClassificationSystemResponse, error) {
	normalized, err := normalizeSystemSave(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureSystemCodeUnique(tx, normalized.SystemCode, ""); err != nil {
			return err
		}
		now := time.Now()
		createdID = utils.GenerateUUID()
		system := models.BaseClassificationSystem{
			ClassificationSystemID: createdID,
			SystemCode:             normalized.SystemCode,
			SystemName:             normalized.SystemName,
			Sort:                   normalized.Sort,
			Remark:                 normalized.Remark,
			RowVersion:             1,
			CreatorID:              optionalBaseOperatorID(operatorID),
			UpdaterID:              optionalBaseOperatorID(operatorID),
			CreateDate:             &now,
			UpdateDate:             &now,
			DelFlag:                0,
		}
		return tx.Create(&system).Error
	}); err != nil {
		return nil, err
	}
	return s.GetSystemDetail(createdID)
}

// UpdateSystem 修改分类体系
func (s *ClassificationSystemService) UpdateSystem(systemID string, req models.SaveClassificationSystemRequest, operatorID string) (*models.ClassificationSystemResponse, error) {
	if err := validateClassificationUUID(systemID, "分类体系ID"); err != nil {
		return nil, err
	}
	normalized, err := normalizeSystemSave(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		system, err := s.lockSystem(tx, systemID)
		if err != nil {
			return err
		}
		if system.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 分类体系已被其他人修改，请刷新后重试", ErrClassificationConflict)
		}
		if err := s.ensureSystemCodeUnique(tx, normalized.SystemCode, systemID); err != nil {
			return err
		}
		now := time.Now()
		updates := map[string]interface{}{
			"system_code": normalized.SystemCode,
			"system_name": normalized.SystemName,
			"sort":        normalized.Sort,
			"remark":      normalized.Remark,
			"row_version": system.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}
		return tx.Model(&models.BaseClassificationSystem{}).Where("classification_system_id = ?", systemID).Updates(updates).Error
	}); err != nil {
		return nil, err
	}
	return s.GetSystemDetail(systemID)
}

// DeleteSystem 单条删除分类体系（有节点时禁止删除）
func (s *ClassificationSystemService) DeleteSystem(systemID string, operatorID string) error {
	if err := validateClassificationUUID(systemID, "分类体系ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		system, err := s.lockSystem(tx, systemID)
		if err != nil {
			return err
		}
		var nodeCount int64
		if err := tx.Model(&models.BaseClassificationNode{}).
			Where("classification_system_id = ? AND del_flag = 0", systemID).
			Count(&nodeCount).Error; err != nil {
			return err
		}
		if nodeCount > 0 {
			return fmt.Errorf("%w: 分类体系存在节点，不能删除", ErrClassificationConflict)
		}
		now := time.Now()
		return tx.Model(&models.BaseClassificationSystem{}).Where("classification_system_id = ?", systemID).Updates(map[string]interface{}{
			"del_flag":    1,
			"row_version": system.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

// getSystemByID 按ID查询体系（未删除）
func (s *ClassificationSystemService) getSystemByID(tx *gorm.DB, systemID string) (*models.BaseClassificationSystem, error) {
	var system models.BaseClassificationSystem
	if err := tx.Where("classification_system_id = ? AND del_flag = 0", systemID).First(&system).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类体系不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &system, nil
}

// GetSystemByCode 按体系编码查询体系（未删除），供节点服务复用
func (s *ClassificationSystemService) GetSystemByCode(systemCode string) (*models.BaseClassificationSystem, error) {
	code := strings.TrimSpace(systemCode)
	if code == "" {
		return nil, fmt.Errorf("%w: 体系编码不能为空", ErrClassificationInvalidInput)
	}
	var system models.BaseClassificationSystem
	if err := database.DB.Where("system_code = ? AND del_flag = 0", code).First(&system).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类体系不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &system, nil
}

// lockSystem 事务内加行锁查询体系
func (s *ClassificationSystemService) lockSystem(tx *gorm.DB, systemID string) (*models.BaseClassificationSystem, error) {
	var system models.BaseClassificationSystem
	if err := tx.Where("classification_system_id = ? AND del_flag = 0", systemID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&system).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类体系不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &system, nil
}

// ensureSystemCodeUnique 校验体系编码全局唯一
func (s *ClassificationSystemService) ensureSystemCodeUnique(tx *gorm.DB, code, excludeID string) error {
	query := tx.Model(&models.BaseClassificationSystem{}).Where("del_flag = 0 AND system_code = ?", code)
	if excludeID != "" {
		query = query.Where("classification_system_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 体系编码已存在", ErrClassificationConflict)
	}
	return nil
}

// normalizedSystemSave 体系保存请求归一化结果
type normalizedSystemSave struct {
	SystemCode         string
	SystemName         string
	Sort               int
	Remark             *string
	ExpectedRowVersion int
}

// normalizeSystemSave 归一化体系保存请求并校验基础字段
func normalizeSystemSave(req models.SaveClassificationSystemRequest, requireVersion bool) (*normalizedSystemSave, error) {
	code := strings.TrimSpace(req.SystemCode)
	if code == "" {
		return nil, fmt.Errorf("%w: 体系编码不能为空", ErrClassificationInvalidInput)
	}
	if len([]rune(code)) > 64 {
		return nil, fmt.Errorf("%w: 体系编码不能超过64个字符", ErrClassificationInvalidInput)
	}
	name := strings.TrimSpace(req.SystemName)
	if name == "" {
		return nil, fmt.Errorf("%w: 体系名称不能为空", ErrClassificationInvalidInput)
	}
	if len([]rune(name)) > 128 {
		return nil, fmt.Errorf("%w: 体系名称不能超过128个字符", ErrClassificationInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrClassificationInvalidInput)
	}
	return &normalizedSystemSave{
		SystemCode:         code,
		SystemName:         name,
		Sort:               req.Sort,
		Remark:             normalizeBaseOptionalString(req.Remark),
		ExpectedRowVersion: req.ExpectedRowVersion,
	}, nil
}

// validateClassificationUUID 校验UUID格式
func validateClassificationUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrClassificationInvalidInput, label)
	}
	return nil
}

// validateClassificationStatus 校验状态值（节点使用）
func validateClassificationStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrClassificationInvalidInput)
	}
	return nil
}

// systemToResponse 体系模型转响应
func systemToResponse(system models.BaseClassificationSystem) *models.ClassificationSystemResponse {
	return &models.ClassificationSystemResponse{
		ClassificationSystemID: system.ClassificationSystemID,
		SystemCode:             system.SystemCode,
		SystemName:             system.SystemName,
		Sort:                   system.Sort,
		Remark:                 system.Remark,
		RowVersion:             system.RowVersion,
		CreateDate:             models.TimeToStringPtr(system.CreateDate),
		UpdateDate:             models.TimeToStringPtr(system.UpdateDate),
	}
}

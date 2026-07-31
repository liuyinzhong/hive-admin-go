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

var (
	ErrErpWarehouseInvalidInput = errors.New("仓库参数错误")
	ErrErpWarehouseNotFound     = errors.New("仓库数据不存在")
	ErrErpWarehouseConflict     = errors.New("仓库数据冲突")
)

type ErpWarehouseService struct{}

func NewErpWarehouseService() *ErpWarehouseService {
	return &ErpWarehouseService{}
}

func (s *ErpWarehouseService) GetWarehouseList(req models.ErpWarehouseListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpWarehousePage(req.Page, req.PageSize, 20, 100)
	query := database.DB.Model(&models.ErpWarehouse{}).Where("del_flag = 0")
	var err error
	query, err = s.applyWarehouseFilters(query, req.Keyword, req.StorageType, req.BusinessScope, req.Status)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"warehouseCode": "warehouse_code",
		"warehouseName": "warehouse_name",
		"storageType":   "storage_type",
		"businessScope": "business_scope",
		"status":        "status",
		"createDate":    "create_date",
		"updateDate":    "update_date",
	})
	if order == "" {
		order = "update_date desc, create_date desc"
	}

	var warehouses []models.ErpWarehouse
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&warehouses).Error; err != nil {
		return nil, err
	}

	responses := erpWarehousesToResponses(warehouses)
	if err := s.fillWarehouseZoneCounts(database.DB, responses); err != nil {
		return nil, err
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

func (s *ErpWarehouseService) GetWarehouseDetail(warehouseID string) (*models.ErpWarehouseResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}

	var warehouse models.ErpWarehouse
	if err := database.DB.Where("warehouse_id = ? AND del_flag = 0", warehouseID).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 仓库不存在", ErrErpWarehouseNotFound)
		}
		return nil, err
	}
	response := erpWarehouseToResponse(warehouse)
	responses := []models.ErpWarehouseResponse{*response}
	if err := s.fillWarehouseZoneCounts(database.DB, responses); err != nil {
		return nil, err
	}
	response.ZoneCount = responses[0].ZoneCount
	return response, nil
}

func (s *ErpWarehouseService) GetWarehouseOptions(req models.ErpWarehouseOptionsRequest) ([]models.ErpWarehouseOptionResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.ErpWarehouse{}).Where("del_flag = 0 AND status = 1")
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.ToLower(strings.TrimSpace(req.Keyword)) + "%"
		query = query.Where("LOWER(warehouse_code) LIKE ? OR LOWER(warehouse_name) LIKE ?", keyword, keyword)
	}

	var warehouses []models.ErpWarehouse
	if err := query.Order("warehouse_code asc").Limit(pageSize).Find(&warehouses).Error; err != nil {
		return nil, err
	}

	options := make([]models.ErpWarehouseOptionResponse, 0, len(warehouses))
	for _, warehouse := range warehouses {
		options = append(options, models.ErpWarehouseOptionResponse{
			WarehouseID:   warehouse.WarehouseID,
			WarehouseCode: warehouse.WarehouseCode,
			WarehouseName: warehouse.WarehouseName,
			StorageType:   warehouse.StorageType,
			BusinessScope: warehouse.BusinessScope,
		})
	}
	return options, nil
}

func (s *ErpWarehouseService) CreateWarehouse(req models.SaveErpWarehouseRequest, operatorID string) (*models.ErpWarehouseResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureWarehouseNameUnique(tx, normalized.WarehouseNameNormalized, ""); err != nil {
			return err
		}
		code, err := s.nextWarehouseCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		warehouse := models.ErpWarehouse{
			WarehouseID:             createdID,
			WarehouseCode:           code,
			WarehouseName:           normalized.WarehouseName,
			WarehouseNameNormalized: normalized.WarehouseNameNormalized,
			StorageType:             normalized.StorageType,
			BusinessScope:           normalized.BusinessScope,
			Address:                 normalized.Address,
			Status:                  normalized.Status,
			Remark:                  normalized.Remark,
			RowVersion:              1,
			CreatorID:               optionalErpWarehouseOperatorID(operatorID),
			UpdaterID:               optionalErpWarehouseOperatorID(operatorID),
			CreateDate:              &now,
			UpdateDate:              &now,
			DelFlag:                 0,
		}
		return tx.Create(&warehouse).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseDetail(createdID)
}

func (s *ErpWarehouseService) UpdateWarehouse(warehouseID string, req models.SaveErpWarehouseRequest, operatorID string) (*models.ErpWarehouseResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var current models.ErpWarehouse
		if err := tx.Where("warehouse_id = ? AND del_flag = 0", warehouseID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 仓库不存在", ErrErpWarehouseNotFound)
			}
			return err
		}
		if current.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 仓库已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}
		if err := s.ensureWarehouseNameUnique(tx, normalized.WarehouseNameNormalized, warehouseID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ErpWarehouse{}).Where("warehouse_id = ?", warehouseID).Updates(map[string]interface{}{
			"warehouse_name":            normalized.WarehouseName,
			"warehouse_name_normalized": normalized.WarehouseNameNormalized,
			"storage_type":              normalized.StorageType,
			"business_scope":            normalized.BusinessScope,
			"address":                   normalized.Address,
			"status":                    normalized.Status,
			"remark":                    normalized.Remark,
			"row_version":               current.RowVersion + 1,
			"updater_id":                optionalErpWarehouseOperatorID(operatorID),
			"update_date":               now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseDetail(warehouseID)
}

func (s *ErpWarehouseService) UpdateWarehouseStatus(warehouseID string, req models.UpdateErpWarehouseStatusRequest, operatorID string) (*models.ErpWarehouseResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var current models.ErpWarehouse
		if err := tx.Where("warehouse_id = ? AND del_flag = 0", warehouseID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 仓库不存在", ErrErpWarehouseNotFound)
			}
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 仓库已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}

		now := time.Now()
		return tx.Model(&models.ErpWarehouse{}).Where("warehouse_id = ?", warehouseID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalErpWarehouseOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseDetail(warehouseID)
}

func (s *ErpWarehouseService) DeleteWarehouse(warehouseID string, req models.DeleteErpWarehouseRequest, operatorID string) error {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return err
	}
	if req.ExpectedRowVersion <= 0 {
		return fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var current models.ErpWarehouse
		if err := tx.Where("warehouse_id = ? AND del_flag = 0", warehouseID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 仓库不存在", ErrErpWarehouseNotFound)
			}
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 仓库已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}

		now := time.Now()
		return tx.Model(&models.ErpWarehouse{}).Where("warehouse_id = ?", warehouseID).Updates(map[string]interface{}{
			"del_flag":    1,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalErpWarehouseOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

type normalizedErpWarehouseSave struct {
	WarehouseName           string
	WarehouseNameNormalized string
	StorageType             string
	BusinessScope           string
	Address                 *string
	Status                  int
	Remark                  *string
	ExpectedRowVersion      int
}

func (s *ErpWarehouseService) normalizeSaveRequest(req models.SaveErpWarehouseRequest, requireVersion bool) (*normalizedErpWarehouseSave, error) {
	warehouseName := strings.TrimSpace(req.WarehouseName)
	if warehouseName == "" {
		return nil, fmt.Errorf("%w: 仓库名称不能为空", ErrErpWarehouseInvalidInput)
	}
	if len([]rune(warehouseName)) > 128 {
		return nil, fmt.Errorf("%w: 仓库名称不能超过128个字符", ErrErpWarehouseInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}
	if err := validateErpWarehouseStatus(req.Status); err != nil {
		return nil, err
	}
	storageType, err := normalizeErpWarehouseStorageType(req.StorageType)
	if err != nil {
		return nil, err
	}
	businessScope, err := normalizeErpWarehouseBusinessScope(req.BusinessScope)
	if err != nil {
		return nil, err
	}

	return &normalizedErpWarehouseSave{
		WarehouseName:           warehouseName,
		WarehouseNameNormalized: normalizeErpWarehouseText(warehouseName),
		StorageType:             storageType,
		BusinessScope:           businessScope,
		Address:                 normalizeErpWarehouseOptionalString(req.Address),
		Status:                  req.Status,
		Remark:                  normalizeErpWarehouseOptionalString(req.Remark),
		ExpectedRowVersion:      req.ExpectedRowVersion,
	}, nil
}

func (s *ErpWarehouseService) applyWarehouseFilters(query *gorm.DB, keyword, storageType, businessScope string, status *int) (*gorm.DB, error) {
	if strings.TrimSpace(keyword) != "" {
		kw := "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
		query = query.Where("LOWER(warehouse_code) LIKE ? OR LOWER(warehouse_name) LIKE ?", kw, kw)
	}
	if strings.TrimSpace(storageType) != "" {
		normalized, err := normalizeErpWarehouseStorageType(storageType)
		if err != nil {
			return nil, err
		}
		query = query.Where("storage_type = ?", normalized)
	}
	if strings.TrimSpace(businessScope) != "" {
		normalized, err := normalizeErpWarehouseBusinessScope(businessScope)
		if err != nil {
			return nil, err
		}
		query = query.Where("business_scope = ?", normalized)
	}
	if status != nil {
		if err := validateErpWarehouseStatus(*status); err != nil {
			return nil, err
		}
		query = query.Where("status = ?", *status)
	}
	return query, nil
}

func (s *ErpWarehouseService) ensureWarehouseNameUnique(tx *gorm.DB, nameNormalized, excludeID string) error {
	query := tx.Model(&models.ErpWarehouse{}).Where("del_flag = 0 AND warehouse_name_normalized = ?", nameNormalized)
	if excludeID != "" {
		query = query.Where("warehouse_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 仓库名称已存在", ErrErpWarehouseConflict)
	}
	return nil
}

func (s *ErpWarehouseService) nextWarehouseCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_WAREHOUSE", "WH", 6)
}

func erpWarehousesToResponses(warehouses []models.ErpWarehouse) []models.ErpWarehouseResponse {
	responses := make([]models.ErpWarehouseResponse, 0, len(warehouses))
	for _, warehouse := range warehouses {
		responses = append(responses, *erpWarehouseToResponse(warehouse))
	}
	return responses
}

func erpWarehouseToResponse(warehouse models.ErpWarehouse) *models.ErpWarehouseResponse {
	return &models.ErpWarehouseResponse{
		WarehouseID:   warehouse.WarehouseID,
		WarehouseCode: warehouse.WarehouseCode,
		WarehouseName: warehouse.WarehouseName,
		StorageType:   warehouse.StorageType,
		BusinessScope: warehouse.BusinessScope,
		Address:       warehouse.Address,
		Status:        warehouse.Status,
		Remark:        warehouse.Remark,
		RowVersion:    warehouse.RowVersion,
		CreateDate:    models.TimeToStringPtr(warehouse.CreateDate),
		UpdateDate:    models.TimeToStringPtr(warehouse.UpdateDate),
	}
}

func validateErpWarehouseUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrErpWarehouseInvalidInput, label)
	}
	return nil
}

func validateErpWarehouseStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrErpWarehouseInvalidInput)
	}
	return nil
}

func normalizeErpWarehouseStorageType(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.WarehouseStorageTypeNormal,
		models.WarehouseStorageTypeRefrigerated,
		models.WarehouseStorageTypeFrozen,
		models.WarehouseStorageTypeCool,
		models.WarehouseStorageTypeHazardous:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 仓库储存类型不支持", ErrErpWarehouseInvalidInput)
	}
}

func normalizeErpWarehouseBusinessScope(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.WarehouseBusinessScopeDrug,
		models.WarehouseBusinessScopeConsumable,
		models.WarehouseBusinessScopeDevice,
		models.WarehouseBusinessScopeComprehensive:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 仓库业务范围不支持", ErrErpWarehouseInvalidInput)
	}
}

func normalizeErpWarehouseOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeErpWarehouseText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func normalizeErpWarehousePage(page, pageSize, defaultSize, maxSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return page, pageSize
}

func optionalErpWarehouseOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

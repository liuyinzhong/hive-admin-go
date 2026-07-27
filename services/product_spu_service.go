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
	ErrProductSpuInvalidInput = errors.New("通用产品参数错误")
	ErrProductSpuNotFound     = errors.New("通用产品数据不存在")
	ErrProductSpuConflict     = errors.New("通用产品数据冲突")
)

type ProductSpuService struct{}

func NewProductSpuService() *ProductSpuService {
	return &ProductSpuService{}
}

func (s *ProductSpuService) GetProductSpuList(req models.ProductSpuListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeProductSpuPage(req.Page, req.PageSize, 20, 100)
	query := database.DB.Model(&models.ProductSpu{}).Where("del_flag = 0")
	query = s.applyProductSpuFilters(query, req.Keyword, req.ProductType, req.Status)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"spuCode":     "spu_code",
		"productName": "product_name",
		"productType": "product_type",
		"status":      "status",
		"createDate":  "create_date",
		"updateDate":  "update_date",
	})
	if order == "" {
		order = "create_date desc"
	}

	var spus []models.ProductSpu
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&spus).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: s.productSpusToResponses(spus), Total: total}, nil
}

func (s *ProductSpuService) GetProductSpuDetail(spuID string) (*models.ProductSpuResponse, error) {
	if err := validateProductSpuUUID(spuID, "通用产品ID"); err != nil {
		return nil, err
	}

	var spu models.ProductSpu
	if err := database.DB.Where("spu_id = ? AND del_flag = 0", spuID).First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 通用产品不存在", ErrProductSpuNotFound)
		}
		return nil, err
	}
	return s.productSpuToResponse(spu), nil
}

func (s *ProductSpuService) CreateProductSpu(req models.SaveProductSpuRequest, operatorID string) (*models.ProductSpuResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureProductSpuUnique(tx, normalized.ProductNameNormalized, normalized.ProductType, ""); err != nil {
			return err
		}
		code, err := s.nextProductSpuCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		spu := models.ProductSpu{
			SpuID:                 createdID,
			SpuCode:               code,
			ProductName:           normalized.ProductName,
			ProductNameNormalized: normalized.ProductNameNormalized,
			ShortName:             normalized.ShortName,
			ShortNameNormalized:   normalized.ShortNameNormalized,
			ProductType:           normalized.ProductType,
			Description:           normalized.Description,
			Status:                normalized.Status,
			RowVersion:            1,
			CreatorID:             optionalProductSpuOperatorID(operatorID),
			UpdaterID:             optionalProductSpuOperatorID(operatorID),
			CreateDate:            &now,
			UpdateDate:            &now,
			DelFlag:               0,
		}
		return tx.Create(&spu).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSpuDetail(createdID)
}

func (s *ProductSpuService) UpdateProductSpu(spuID string, req models.SaveProductSpuRequest, operatorID string) (*models.ProductSpuResponse, error) {
	if err := validateProductSpuUUID(spuID, "通用产品ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var spu models.ProductSpu
		if err := tx.Where("spu_id = ? AND del_flag = 0", spuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&spu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 通用产品不存在", ErrProductSpuNotFound)
			}
			return err
		}
		if spu.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 通用产品已被其他人修改，请刷新后重试", ErrProductSpuConflict)
		}
		if err := s.ensureProductSpuUnique(tx, normalized.ProductNameNormalized, normalized.ProductType, spuID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ProductSpu{}).Where("spu_id = ?", spuID).Updates(map[string]interface{}{
			"product_name":            normalized.ProductName,
			"product_name_normalized": normalized.ProductNameNormalized,
			"short_name":              normalized.ShortName,
			"short_name_normalized":   normalized.ShortNameNormalized,
			"product_type":            normalized.ProductType,
			"description":             normalized.Description,
			"status":                  normalized.Status,
			"row_version":             spu.RowVersion + 1,
			"updater_id":              optionalProductSpuOperatorID(operatorID),
			"update_date":             now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSpuDetail(spuID)
}

func (s *ProductSpuService) UpdateProductSpuStatus(spuID string, req models.UpdateProductSpuStatusRequest, operatorID string) (*models.ProductSpuResponse, error) {
	if err := validateProductSpuUUID(spuID, "通用产品ID"); err != nil {
		return nil, err
	}
	if err := validateProductSpuStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var spu models.ProductSpu
		if err := tx.Where("spu_id = ? AND del_flag = 0", spuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&spu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 通用产品不存在", ErrProductSpuNotFound)
			}
			return err
		}
		if spu.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 通用产品已被其他人修改，请刷新后重试", ErrProductSpuConflict)
		}

		now := time.Now()
		return tx.Model(&models.ProductSpu{}).Where("spu_id = ?", spuID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": spu.RowVersion + 1,
			"updater_id":  optionalProductSpuOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSpuDetail(spuID)
}

func (s *ProductSpuService) GetProductSpuOptions(req models.ProductSpuOptionsRequest) ([]models.ProductSpuOptionResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := database.DB.Model(&models.ProductSpu{}).Where("del_flag = 0 AND status = 1")
	query = s.applyProductSpuFilters(query, req.Keyword, req.ProductType, nil)

	var spus []models.ProductSpu
	if err := query.Order("spu_code asc").Limit(pageSize).Find(&spus).Error; err != nil {
		return nil, err
	}

	options := make([]models.ProductSpuOptionResponse, 0, len(spus))
	for _, spu := range spus {
		options = append(options, models.ProductSpuOptionResponse{
			SpuID:       spu.SpuID,
			SpuCode:     spu.SpuCode,
			ProductName: spu.ProductName,
			ShortName:   spu.ShortName,
			ProductType: spu.ProductType,
		})
	}
	return options, nil
}

type normalizedProductSpuSave struct {
	ProductName           string
	ProductNameNormalized string
	ShortName             *string
	ShortNameNormalized   *string
	ProductType           string
	Description           *string
	Status                int
	ExpectedRowVersion    int
}

func (s *ProductSpuService) normalizeSaveRequest(req models.SaveProductSpuRequest, requireVersion bool) (*normalizedProductSpuSave, error) {
	productName := strings.TrimSpace(req.ProductName)
	if productName == "" {
		return nil, fmt.Errorf("%w: 通用名称不能为空", ErrProductSpuInvalidInput)
	}
	if len([]rune(productName)) > 128 {
		return nil, fmt.Errorf("%w: 通用名称不能超过128个字符", ErrProductSpuInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrProductSpuInvalidInput)
	}
	if err := validateProductSpuStatus(req.Status); err != nil {
		return nil, err
	}
	productType, err := normalizeProductType(req.ProductType)
	if err != nil {
		return nil, err
	}
	shortName := normalizeProductSpuOptionalString(req.ShortName)
	if shortName != nil && len([]rune(*shortName)) > 64 {
		return nil, fmt.Errorf("%w: 简称不能超过64个字符", ErrProductSpuInvalidInput)
	}
	var shortNameNormalized *string
	if shortName != nil {
		value := normalizeProductSpuText(*shortName)
		shortNameNormalized = &value
	}
	description := normalizeProductSpuOptionalString(req.Description)
	if description != nil && len([]rune(*description)) > 2000 {
		return nil, fmt.Errorf("%w: 描述不能超过2000个字符", ErrProductSpuInvalidInput)
	}

	return &normalizedProductSpuSave{
		ProductName:           productName,
		ProductNameNormalized: normalizeProductSpuText(productName),
		ShortName:             shortName,
		ShortNameNormalized:   shortNameNormalized,
		ProductType:           productType,
		Description:           description,
		Status:                req.Status,
		ExpectedRowVersion:    req.ExpectedRowVersion,
	}, nil
}

func (s *ProductSpuService) applyProductSpuFilters(query *gorm.DB, keyword, productType string, status *int) *gorm.DB {
	if strings.TrimSpace(keyword) != "" {
		kw := "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
		query = query.Where(
			"LOWER(product_name) LIKE ? OR LOWER(spu_code) LIKE ? OR LOWER(IFNULL(short_name, '')) LIKE ?",
			kw, kw, kw,
		)
	}
	if strings.TrimSpace(productType) != "" {
		if normalized, err := normalizeProductType(productType); err == nil {
			query = query.Where("product_type = ?", normalized)
		}
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	return query
}

func (s *ProductSpuService) ensureProductSpuUnique(tx *gorm.DB, productNameNormalized, productType, excludeID string) error {
	query := tx.Model(&models.ProductSpu{}).
		Where("del_flag = 0 AND product_name_normalized = ? AND product_type = ?", productNameNormalized, productType)
	if excludeID != "" {
		query = query.Where("spu_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 同类型通用名称已存在", ErrProductSpuConflict)
	}
	return nil
}

func (s *ProductSpuService) nextProductSpuCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "PRODUCT_SPU", "SPU", 6)
}

func (s *ProductSpuService) productSpusToResponses(spus []models.ProductSpu) []models.ProductSpuResponse {
	responses := make([]models.ProductSpuResponse, 0, len(spus))
	for _, spu := range spus {
		responses = append(responses, *s.productSpuToResponse(spu))
	}
	return responses
}

func (s *ProductSpuService) productSpuToResponse(spu models.ProductSpu) *models.ProductSpuResponse {
	return &models.ProductSpuResponse{
		SpuID:       spu.SpuID,
		SpuCode:     spu.SpuCode,
		ProductName: spu.ProductName,
		ShortName:   spu.ShortName,
		ProductType: spu.ProductType,
		Description: spu.Description,
		Status:      spu.Status,
		RowVersion:  spu.RowVersion,
		CreateDate:  models.TimeToStringPtr(spu.CreateDate),
		UpdateDate:  models.TimeToStringPtr(spu.UpdateDate),
	}
}

func validateProductSpuUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrProductSpuInvalidInput, label)
	}
	return nil
}

func validateProductSpuStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrProductSpuInvalidInput)
	}
	return nil
}

func normalizeProductType(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.ProductTypeDrug,
		models.ProductTypeDevice,
		models.ProductTypeConsumable,
		models.ProductTypeFSMP,
		models.ProductTypeOther:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 产品类型不支持", ErrProductSpuInvalidInput)
	}
}

func normalizeProductSpuOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeProductSpuText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func normalizeProductSpuPage(page, pageSize, defaultSize, maxSize int) (int, int) {
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

func optionalProductSpuOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

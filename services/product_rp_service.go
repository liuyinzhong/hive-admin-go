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
	ErrProductRpInvalidInput = errors.New("规格产品参数错误")
	ErrProductRpNotFound     = errors.New("规格产品数据不存在")
	ErrProductRpConflict     = errors.New("规格产品数据冲突")
)

type ProductRpService struct{}

func NewProductRpService() *ProductRpService {
	return &ProductRpService{}
}

func (s *ProductRpService) GetProductRpList(req models.ProductRpListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeProductRpPage(req.Page, req.PageSize, 20, 100)
	query, err := s.applyProductRpFilters(s.baseProductRpQuery(), req.Keyword, req.SpuID, req.ProductType, req.Status)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"rpCode":     "product_rp.rp_code",
		"specName":   "product_rp.spec_name",
		"createDate": "product_rp.create_date",
		"updateDate": "product_rp.update_date",
	})
	if order == "" {
		order = "product_rp.create_date desc"
	}

	var rows []productRpQueryRow
	if err := query.Select(productRpSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: productRpRowsToResponses(rows), Total: total}, nil
}

func (s *ProductRpService) GetProductRpDetail(rpID string) (*models.ProductRpResponse, error) {
	if err := validateProductRpUUID(rpID, "规格产品ID"); err != nil {
		return nil, err
	}

	var row productRpQueryRow
	if err := s.baseProductRpQuery().
		Select(productRpSelectFields()).
		Where("product_rp.rp_id = ?", rpID).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.RpID == "" {
		return nil, fmt.Errorf("%w: 规格产品不存在", ErrProductRpNotFound)
	}
	return productRpRowToResponse(row), nil
}

func (s *ProductRpService) CreateProductRp(req models.SaveProductRpRequest, operatorID string) (*models.ProductRpResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.getEditableProductSpu(tx, normalized.SpuID); err != nil {
			return err
		}
		if err := s.ensureProductRpUnique(tx, normalized.SpuID, normalized.SpecNameNormalized, ""); err != nil {
			return err
		}
		code, err := s.nextProductRpCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		rp := models.ProductRp{
			RpID:               createdID,
			RpCode:             code,
			SpuID:              normalized.SpuID,
			SpecName:           normalized.SpecName,
			SpecNameNormalized: normalized.SpecNameNormalized,
			DosageForm:         normalized.DosageForm,
			StrengthText:       normalized.StrengthText,
			Description:        normalized.Description,
			Status:             normalized.Status,
			RowVersion:         1,
			CreatorID:          optionalProductRpOperatorID(operatorID),
			UpdaterID:          optionalProductRpOperatorID(operatorID),
			CreateDate:         &now,
			UpdateDate:         &now,
			DelFlag:            0,
		}
		return tx.Create(&rp).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductRpDetail(createdID)
}

func (s *ProductRpService) UpdateProductRp(rpID string, req models.SaveProductRpRequest, operatorID string) (*models.ProductRpResponse, error) {
	if err := validateProductRpUUID(rpID, "规格产品ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var rp models.ProductRp
		if err := tx.Where("rp_id = ? AND del_flag = 0", rpID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&rp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 规格产品不存在", ErrProductRpNotFound)
			}
			return err
		}
		if rp.SpuID != normalized.SpuID {
			return fmt.Errorf("%w: 不允许更换所属通用产品", ErrProductRpInvalidInput)
		}
		if rp.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 规格产品已被其他人修改，请刷新后重试", ErrProductRpConflict)
		}
		if _, err := s.getEditableProductSpu(tx, rp.SpuID); err != nil {
			return err
		}
		if err := s.ensureProductRpUnique(tx, rp.SpuID, normalized.SpecNameNormalized, rpID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ProductRp{}).Where("rp_id = ?", rpID).Updates(map[string]interface{}{
			"spec_name":            normalized.SpecName,
			"spec_name_normalized": normalized.SpecNameNormalized,
			"dosage_form":          normalized.DosageForm,
			"strength_text":        normalized.StrengthText,
			"description":          normalized.Description,
			"status":               normalized.Status,
			"row_version":          rp.RowVersion + 1,
			"updater_id":           optionalProductRpOperatorID(operatorID),
			"update_date":          now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductRpDetail(rpID)
}

func (s *ProductRpService) UpdateProductRpStatus(rpID string, req models.UpdateProductRpStatusRequest, operatorID string) (*models.ProductRpResponse, error) {
	if err := validateProductRpUUID(rpID, "规格产品ID"); err != nil {
		return nil, err
	}
	if err := validateProductRpStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var rp models.ProductRp
		if err := tx.Where("rp_id = ? AND del_flag = 0", rpID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&rp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 规格产品不存在", ErrProductRpNotFound)
			}
			return err
		}
		if rp.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 规格产品已被其他人修改，请刷新后重试", ErrProductRpConflict)
		}
		if _, err := s.getEditableProductSpu(tx, rp.SpuID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ProductRp{}).Where("rp_id = ?", rpID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": rp.RowVersion + 1,
			"updater_id":  optionalProductRpOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductRpDetail(rpID)
}

func (s *ProductRpService) GetProductRpOptions(req models.ProductRpOptionsRequest) ([]models.ProductRpOptionResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query, err := s.applyProductRpFilters(s.baseProductRpQuery(), req.Keyword, req.SpuID, req.ProductType, nil)
	if err != nil {
		return nil, err
	}
	query = query.Where("product_rp.status = 1 AND product_spu.status = 1")

	var rows []productRpQueryRow
	if err := query.Select(productRpSelectFields()).
		Order("product_rp.rp_code asc").
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	options := make([]models.ProductRpOptionResponse, 0, len(rows))
	for _, row := range rows {
		options = append(options, models.ProductRpOptionResponse{
			RpID:        row.RpID,
			RpCode:      row.RpCode,
			SpuID:       row.SpuID,
			SpuCode:     row.SpuCode,
			ProductName: row.ProductName,
			ProductType: row.ProductType,
			SpecName:    row.SpecName,
		})
	}
	return options, nil
}

type normalizedProductRpSave struct {
	SpuID              string
	SpecName           string
	SpecNameNormalized string
	DosageForm         *string
	StrengthText       *string
	Description        *string
	Status             int
	ExpectedRowVersion int
}

type productRpQueryRow struct {
	RpID         string
	RpCode       string
	SpuID        string
	SpuCode      string
	ProductName  string
	ProductType  string
	SpecName     string
	DosageForm   *string
	StrengthText *string
	Description  *string
	Status       int
	RowVersion   int
	CreateDate   *time.Time
	UpdateDate   *time.Time
}

func (s *ProductRpService) normalizeSaveRequest(req models.SaveProductRpRequest, requireVersion bool) (*normalizedProductRpSave, error) {
	if err := validateProductRpUUID(req.SpuID, "通用产品ID"); err != nil {
		return nil, err
	}
	specName := strings.TrimSpace(req.SpecName)
	if specName == "" {
		return nil, fmt.Errorf("%w: 规格名称不能为空", ErrProductRpInvalidInput)
	}
	if len([]rune(specName)) > 128 {
		return nil, fmt.Errorf("%w: 规格名称不能超过128个字符", ErrProductRpInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrProductRpInvalidInput)
	}
	if err := validateProductRpStatus(req.Status); err != nil {
		return nil, err
	}

	dosageForm := normalizeProductRpOptionalString(req.DosageForm)
	if dosageForm != nil && len([]rune(*dosageForm)) > 64 {
		return nil, fmt.Errorf("%w: 剂型/形态不能超过64个字符", ErrProductRpInvalidInput)
	}
	strengthText := normalizeProductRpOptionalString(req.StrengthText)
	if strengthText != nil && len([]rune(*strengthText)) > 128 {
		return nil, fmt.Errorf("%w: 含量/规格文本不能超过128个字符", ErrProductRpInvalidInput)
	}
	description := normalizeProductRpOptionalString(req.Description)
	if description != nil && len([]rune(*description)) > 2000 {
		return nil, fmt.Errorf("%w: 描述不能超过2000个字符", ErrProductRpInvalidInput)
	}

	return &normalizedProductRpSave{
		SpuID:              req.SpuID,
		SpecName:           specName,
		SpecNameNormalized: normalizeProductRpText(specName),
		DosageForm:         dosageForm,
		StrengthText:       strengthText,
		Description:        description,
		Status:             req.Status,
		ExpectedRowVersion: req.ExpectedRowVersion,
	}, nil
}

func (s *ProductRpService) baseProductRpQuery() *gorm.DB {
	return database.DB.Table("product_rp").
		Joins("INNER JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Where("product_rp.del_flag = 0")
}

func (s *ProductRpService) applyProductRpFilters(query *gorm.DB, keyword, spuID, productType string, status *int) (*gorm.DB, error) {
	if strings.TrimSpace(keyword) != "" {
		kw := "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
		query = query.Where(
			"LOWER(product_rp.rp_code) LIKE ? OR LOWER(product_rp.spec_name) LIKE ? OR LOWER(IFNULL(product_rp.dosage_form, '')) LIKE ? OR LOWER(IFNULL(product_rp.strength_text, '')) LIKE ? OR LOWER(product_spu.spu_code) LIKE ? OR LOWER(product_spu.product_name) LIKE ?",
			kw, kw, kw, kw, kw, kw,
		)
	}
	if strings.TrimSpace(spuID) != "" {
		if err := validateProductRpUUID(spuID, "通用产品ID"); err != nil {
			return nil, err
		}
		query = query.Where("product_rp.spu_id = ?", strings.TrimSpace(spuID))
	}
	if strings.TrimSpace(productType) != "" {
		normalized, err := normalizeProductType(productType)
		if err != nil {
			return nil, fmt.Errorf("%w: 产品类型不支持", ErrProductRpInvalidInput)
		}
		query = query.Where("product_spu.product_type = ?", normalized)
	}
	if status != nil {
		query = query.Where("product_rp.status = ?", *status)
	}
	return query, nil
}

func (s *ProductRpService) getEditableProductSpu(tx *gorm.DB, spuID string) (*models.ProductSpu, error) {
	var spu models.ProductSpu
	if err := tx.Where("spu_id = ? AND del_flag = 0", spuID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 所属通用产品不存在", ErrProductRpNotFound)
		}
		return nil, err
	}
	if spu.Status != 1 {
		return nil, fmt.Errorf("%w: 所属通用产品已停用，不能维护规格产品", ErrProductRpConflict)
	}
	return &spu, nil
}

func (s *ProductRpService) ensureProductRpUnique(tx *gorm.DB, spuID, specNameNormalized, excludeID string) error {
	query := tx.Model(&models.ProductRp{}).
		Where("del_flag = 0 AND spu_id = ? AND spec_name_normalized = ?", spuID, specNameNormalized)
	if excludeID != "" {
		query = query.Where("rp_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前通用产品下规格名称已存在", ErrProductRpConflict)
	}
	return nil
}

func (s *ProductRpService) nextProductRpCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "PRODUCT_RP", "RP", 6)
}

func productRpSelectFields() string {
	return strings.Join([]string{
		"product_rp.rp_id",
		"product_rp.rp_code",
		"product_rp.spu_id",
		"product_spu.spu_code",
		"product_spu.product_name",
		"product_spu.product_type",
		"product_rp.spec_name",
		"product_rp.dosage_form",
		"product_rp.strength_text",
		"product_rp.description",
		"product_rp.status",
		"product_rp.row_version",
		"product_rp.create_date",
		"product_rp.update_date",
	}, ", ")
}

func productRpRowsToResponses(rows []productRpQueryRow) []models.ProductRpResponse {
	responses := make([]models.ProductRpResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, *productRpRowToResponse(row))
	}
	return responses
}

func productRpRowToResponse(row productRpQueryRow) *models.ProductRpResponse {
	return &models.ProductRpResponse{
		RpID:         row.RpID,
		RpCode:       row.RpCode,
		SpuID:        row.SpuID,
		SpuCode:      row.SpuCode,
		ProductName:  row.ProductName,
		ProductType:  row.ProductType,
		SpecName:     row.SpecName,
		DosageForm:   row.DosageForm,
		StrengthText: row.StrengthText,
		Description:  row.Description,
		Status:       row.Status,
		RowVersion:   row.RowVersion,
		CreateDate:   models.TimeToStringPtr(row.CreateDate),
		UpdateDate:   models.TimeToStringPtr(row.UpdateDate),
	}
}

func validateProductRpUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrProductRpInvalidInput, label)
	}
	return nil
}

func validateProductRpStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrProductRpInvalidInput)
	}
	return nil
}

func normalizeProductRpOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeProductRpText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func normalizeProductRpPage(page, pageSize, defaultSize, maxSize int) (int, int) {
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

func optionalProductRpOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

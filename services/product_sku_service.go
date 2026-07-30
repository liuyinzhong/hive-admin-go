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
	ErrProductSkuInvalidInput = errors.New("SKU参数错误")
	ErrProductSkuNotFound     = errors.New("SKU数据不存在")
	ErrProductSkuConflict     = errors.New("SKU数据冲突")
)

type ProductSkuService struct{}

func NewProductSkuService() *ProductSkuService {
	return &ProductSkuService{}
}

func (s *ProductSkuService) GetProductSkuList(req models.ProductSkuListRequest) (*utils.PaginationResponse, error) {
	if err := validateProductSkuUUID(req.MpID, "厂家产品ID"); err != nil {
		return nil, err
	}

	page, pageSize := normalizeProductSkuPage(req.Page, req.PageSize, 20, 100)
	query := s.baseProductSkuQuery().Where("product_sku.mp_id = ?", strings.TrimSpace(req.MpID))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"skuCode":         "product_sku.sku_code",
		"packageSpecName": "product_sku.package_spec_name",
		"cartonSpecName":  "product_sku.carton_spec_name",
		"status":          "product_sku.status",
		"createDate":      "product_sku.create_date",
		"updateDate":      "product_sku.update_date",
	})
	if order == "" {
		order = "product_sku.create_date desc"
	}

	var rows []productSkuQueryRow
	if err := query.Select(productSkuSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: productSkuRowsToResponses(rows), Total: total}, nil
}

func (s *ProductSkuService) GetProductSkuDetail(skuID string) (*models.ProductSkuResponse, error) {
	if err := validateProductSkuUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}

	var row productSkuQueryRow
	if err := s.baseProductSkuQuery().
		Select(productSkuSelectFields()).
		Where("product_sku.sku_id = ?", skuID).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.SkuID == "" {
		return nil, fmt.Errorf("%w: SKU不存在", ErrProductSkuNotFound)
	}
	return productSkuRowToResponse(row), nil
}

func (s *ProductSkuService) CreateProductSku(req models.SaveProductSkuRequest, operatorID string) (*models.ProductSkuResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.getExistingProductMp(tx, normalized.MpID); err != nil {
			return err
		}
		if err := s.ensureProductSkuUnique(tx, normalized, ""); err != nil {
			return err
		}
		code, err := s.nextProductSkuCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		sku := models.ProductSku{
			SkuID:             createdID,
			SkuCode:           code,
			MpID:              normalized.MpID,
			PackageSpecName:   normalized.PackageSpecName,
			PackConversion:    normalized.PackConversion,
			MinUnitName:       normalized.MinUnitName,
			PackageUnitName:   normalized.PackageUnitName,
			CartonUnitName:    normalized.CartonUnitName,
			CartonConversion:  normalized.CartonConversion,
			CartonSpecName:    normalized.CartonSpecName,
			FullChainSpecName: normalized.FullChainSpecName,
			Barcode:           normalized.Barcode,
			Gtin:              normalized.Gtin,
			UdiDi:             normalized.UdiDi,
			AllowSplit:        normalized.AllowSplit,
			Description:       normalized.Description,
			Status:            normalized.Status,
			RowVersion:        1,
			CreatorID:         optionalProductSkuOperatorID(operatorID),
			UpdaterID:         optionalProductSkuOperatorID(operatorID),
			CreateDate:        &now,
			UpdateDate:        &now,
			DelFlag:           0,
		}
		return tx.Create(&sku).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuDetail(createdID)
}

func (s *ProductSkuService) UpdateProductSku(skuID string, req models.SaveProductSkuRequest, operatorID string) (*models.ProductSkuResponse, error) {
	if err := validateProductSkuUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var sku models.ProductSku
		if err := tx.Where("sku_id = ? AND del_flag = 0", skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&sku).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU不存在", ErrProductSkuNotFound)
			}
			return err
		}
		if sku.MpID != normalized.MpID {
			return fmt.Errorf("%w: 不允许更换所属厂家产品", ErrProductSkuInvalidInput)
		}
		if sku.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: SKU已被其他人修改，请刷新后重试", ErrProductSkuConflict)
		}
		if err := s.ensureProductSkuUnique(tx, normalized, skuID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ProductSku{}).Where("sku_id = ?", skuID).Updates(map[string]interface{}{
			"package_spec_name":    normalized.PackageSpecName,
			"pack_conversion":      normalized.PackConversion,
			"min_unit_name":        normalized.MinUnitName,
			"package_unit_name":    normalized.PackageUnitName,
			"carton_unit_name":     normalized.CartonUnitName,
			"carton_conversion":    normalized.CartonConversion,
			"carton_spec_name":     normalized.CartonSpecName,
			"full_chain_spec_name": normalized.FullChainSpecName,
			"barcode":              normalized.Barcode,
			"gtin":                 normalized.Gtin,
			"udi_di":               normalized.UdiDi,
			"allow_split":          normalized.AllowSplit,
			"description":          normalized.Description,
			"status":               normalized.Status,
			"row_version":          sku.RowVersion + 1,
			"updater_id":           optionalProductSkuOperatorID(operatorID),
			"update_date":          now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuDetail(skuID)
}

func (s *ProductSkuService) UpdateProductSkuStatus(skuID string, req models.UpdateProductSkuStatusRequest, operatorID string) (*models.ProductSkuResponse, error) {
	if err := validateProductSkuUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if err := validateProductSkuStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var sku models.ProductSku
		if err := tx.Where("sku_id = ? AND del_flag = 0", skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&sku).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU不存在", ErrProductSkuNotFound)
			}
			return err
		}
		if sku.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: SKU已被其他人修改，请刷新后重试", ErrProductSkuConflict)
		}

		now := time.Now()
		return tx.Model(&models.ProductSku{}).Where("sku_id = ?", skuID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": sku.RowVersion + 1,
			"updater_id":  optionalProductSkuOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuDetail(skuID)
}

func (s *ProductSkuService) GetProductSkuOptions(req models.ProductSkuOptionsRequest) ([]models.ProductSkuOptionResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	query := s.baseProductSkuQuery().Where("product_sku.status = 1")
	if strings.TrimSpace(req.MpID) != "" {
		if err := validateProductSkuUUID(req.MpID, "厂家产品ID"); err != nil {
			return nil, err
		}
		query = query.Where("product_sku.mp_id = ?", strings.TrimSpace(req.MpID))
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"LOWER(product_sku.sku_code) LIKE ? OR LOWER(product_sku.package_spec_name) LIKE ? OR LOWER(product_sku.carton_spec_name) LIKE ? OR LOWER(product_sku.full_chain_spec_name) LIKE ? OR LOWER(IFNULL(product_sku.barcode, '')) LIKE ? OR LOWER(IFNULL(product_sku.gtin, '')) LIKE ? OR LOWER(IFNULL(product_sku.udi_di, '')) LIKE ? OR LOWER(product_mp.mp_code) LIKE ? OR LOWER(base_enterprise.enterprise_name) LIKE ? OR LOWER(product_spu.product_name) LIKE ?",
			like, like, like, like, like, like, like, like, like, like,
		)
	}

	var rows []productSkuQueryRow
	if err := query.Select(productSkuSelectFields()).
		Order("product_sku.sku_code asc").
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	options := make([]models.ProductSkuOptionResponse, 0, len(rows))
	for _, row := range rows {
		options = append(options, models.ProductSkuOptionResponse{
			SpuID:           row.SpuID,
			SpuCode:         row.SpuCode,
			ProductName:     row.ProductName,
			ProductType:     row.ProductType,
			RpID:            row.RpID,
			RpCode:          row.RpCode,
			SpecName:        row.SpecName,
			MpID:            row.MpID,
			MpCode:          row.MpCode,
			EnterpriseID:    row.EnterpriseID,
			EnterpriseName:  row.EnterpriseName,
			ApprovalNo:      row.ApprovalNo,
			BrandName:       row.BrandName,
			SkuID:           row.SkuID,
			SkuCode:         row.SkuCode,
			PackageSpecName: row.PackageSpecName,
		})
	}
	return options, nil
}

type normalizedProductSkuSave struct {
	MpID               string
	PackageSpecName    string
	PackConversion     int
	MinUnitName        string
	PackageUnitName    string
	CartonUnitName     string
	CartonConversion   int
	CartonSpecName     string
	FullChainSpecName  string
	Barcode            *string
	Gtin               *string
	UdiDi              *string
	AllowSplit         int
	Description        *string
	Status             int
	ExpectedRowVersion int
}

type productSkuQueryRow struct {
	SpuID             string
	SpuCode           string
	ProductName       string
	ProductType       string
	RpID              string
	RpCode            string
	SpecName          string
	MpID              string
	MpCode            string
	EnterpriseID      string
	EnterpriseCode    string
	EnterpriseName    string
	ApprovalNo        string
	BrandName         *string
	SkuID             string
	SkuCode           string
	PackageSpecName   string
	PackConversion    int
	MinUnitName       string
	PackageUnitName   string
	CartonUnitName    string
	CartonConversion  int
	CartonSpecName    string
	FullChainSpecName string
	Barcode           *string
	Gtin              *string
	UdiDi             *string
	AllowSplit        int
	Description       *string
	Status            int
	RowVersion        int
	CreateDate        *time.Time
	UpdateDate        *time.Time
}

func (s *ProductSkuService) normalizeSaveRequest(req models.SaveProductSkuRequest, requireVersion bool) (*normalizedProductSkuSave, error) {
	if err := validateProductSkuUUID(req.MpID, "厂家产品ID"); err != nil {
		return nil, err
	}
	if req.PackConversion <= 0 || req.PackConversion > 999999 {
		return nil, fmt.Errorf("%w: 包装换算系数必须在1到999999之间", ErrProductSkuInvalidInput)
	}
	minUnitName := strings.TrimSpace(req.MinUnitName)
	if minUnitName == "" {
		return nil, fmt.Errorf("%w: 最小单位不能为空", ErrProductSkuInvalidInput)
	}
	if len([]rune(minUnitName)) > 32 {
		return nil, fmt.Errorf("%w: 最小单位不能超过32个字符", ErrProductSkuInvalidInput)
	}
	packageUnitName := strings.TrimSpace(req.PackageUnitName)
	if packageUnitName == "" {
		return nil, fmt.Errorf("%w: 包装单位不能为空", ErrProductSkuInvalidInput)
	}
	if len([]rune(packageUnitName)) > 32 {
		return nil, fmt.Errorf("%w: 包装单位不能超过32个字符", ErrProductSkuInvalidInput)
	}
	cartonUnitName := strings.TrimSpace(req.CartonUnitName)
	if cartonUnitName == "" {
		return nil, fmt.Errorf("%w: 大包装单位不能为空", ErrProductSkuInvalidInput)
	}
	if len([]rune(cartonUnitName)) > 32 {
		return nil, fmt.Errorf("%w: 大包装单位不能超过32个字符", ErrProductSkuInvalidInput)
	}
	if req.CartonConversion <= 0 || req.CartonConversion > 999999 {
		return nil, fmt.Errorf("%w: 大包装换算系数必须在1到999999之间", ErrProductSkuInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrProductSkuInvalidInput)
	}
	if err := validateProductSkuStatus(req.Status); err != nil {
		return nil, err
	}
	if err := validateProductSkuBoolFlag(req.AllowSplit, "是否允许拆零"); err != nil {
		return nil, err
	}

	barcode := normalizeProductSkuOptionalString(req.Barcode)
	if barcode != nil && len([]rune(*barcode)) > 64 {
		return nil, fmt.Errorf("%w: 商品条码不能超过64个字符", ErrProductSkuInvalidInput)
	}
	gtin := normalizeProductSkuOptionalString(req.Gtin)
	if gtin != nil && len([]rune(*gtin)) > 64 {
		return nil, fmt.Errorf("%w: GTIN不能超过64个字符", ErrProductSkuInvalidInput)
	}
	udiDi := normalizeProductSkuOptionalString(req.UdiDi)
	if udiDi != nil && len([]rune(*udiDi)) > 128 {
		return nil, fmt.Errorf("%w: UDI-DI不能超过128个字符", ErrProductSkuInvalidInput)
	}
	description := normalizeProductSkuOptionalString(req.Description)
	if description != nil && len([]rune(*description)) > 2000 {
		return nil, fmt.Errorf("%w: 描述不能超过2000个字符", ErrProductSkuInvalidInput)
	}

	packageSpecName := buildProductSkuPackageSpecName(req.PackConversion, minUnitName, packageUnitName)
	cartonSpecName := buildProductSkuCartonSpecName(req.CartonConversion, packageUnitName, cartonUnitName)
	fullChainSpecName := buildProductSkuFullChainSpecName(req.CartonConversion, cartonUnitName, packageUnitName, req.PackConversion, minUnitName)

	return &normalizedProductSkuSave{
		MpID:               strings.TrimSpace(req.MpID),
		PackageSpecName:    packageSpecName,
		PackConversion:     req.PackConversion,
		MinUnitName:        minUnitName,
		PackageUnitName:    packageUnitName,
		CartonUnitName:     cartonUnitName,
		CartonConversion:   req.CartonConversion,
		CartonSpecName:     cartonSpecName,
		FullChainSpecName:  fullChainSpecName,
		Barcode:            barcode,
		Gtin:               gtin,
		UdiDi:              udiDi,
		AllowSplit:         req.AllowSplit,
		Description:        description,
		Status:             req.Status,
		ExpectedRowVersion: req.ExpectedRowVersion,
	}, nil
}

func (s *ProductSkuService) baseProductSkuQuery() *gorm.DB {
	return database.DB.Table("product_sku").
		Joins("INNER JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0").
		Joins("INNER JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("INNER JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("INNER JOIN base_enterprise ON base_enterprise.enterprise_id = product_mp.enterprise_id AND base_enterprise.del_flag = 0").
		Where("product_sku.del_flag = 0")
}

func (s *ProductSkuService) getExistingProductMp(tx *gorm.DB, mpID string) (*models.ProductMp, error) {
	var mp models.ProductMp
	if err := tx.Where("mp_id = ? AND del_flag = 0", mpID).First(&mp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 所属厂家产品不存在", ErrProductSkuNotFound)
		}
		return nil, err
	}
	return &mp, nil
}

func (s *ProductSkuService) ensureProductSkuUnique(tx *gorm.DB, normalized *normalizedProductSkuSave, excludeID string) error {
	query := tx.Model(&models.ProductSku{}).
		Where(
			"del_flag = 0 AND mp_id = ? AND pack_conversion = ? AND min_unit_name = ? AND package_unit_name = ? AND carton_conversion = ? AND carton_unit_name = ?",
			normalized.MpID,
			normalized.PackConversion,
			normalized.MinUnitName,
			normalized.PackageUnitName,
			normalized.CartonConversion,
			normalized.CartonUnitName,
		)
	if excludeID != "" {
		query = query.Where("sku_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前厂家产品下包装链路已存在", ErrProductSkuConflict)
	}
	return nil
}

func (s *ProductSkuService) nextProductSkuCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "PRODUCT_SKU", "SKU", 6)
}

func productSkuSelectFields() string {
	return strings.Join([]string{
		"product_spu.spu_id",
		"product_spu.spu_code",
		"product_spu.product_name",
		"product_spu.product_type",
		"product_rp.rp_id",
		"product_rp.rp_code",
		"product_rp.spec_name",
		"product_mp.mp_id",
		"product_mp.mp_code",
		"product_mp.enterprise_id",
		"base_enterprise.enterprise_code",
		"base_enterprise.enterprise_name",
		"product_mp.approval_no",
		"product_mp.brand_name",
		"product_sku.sku_id",
		"product_sku.sku_code",
		"product_sku.package_spec_name",
		"product_sku.pack_conversion",
		"product_sku.min_unit_name",
		"product_sku.package_unit_name",
		"product_sku.carton_unit_name",
		"product_sku.carton_conversion",
		"product_sku.carton_spec_name",
		"product_sku.full_chain_spec_name",
		"product_sku.barcode",
		"product_sku.gtin",
		"product_sku.udi_di",
		"product_sku.allow_split",
		"product_sku.description",
		"product_sku.status",
		"product_sku.row_version",
		"product_sku.create_date",
		"product_sku.update_date",
	}, ", ")
}

func productSkuRowsToResponses(rows []productSkuQueryRow) []models.ProductSkuResponse {
	responses := make([]models.ProductSkuResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, *productSkuRowToResponse(row))
	}
	return responses
}

func productSkuRowToResponse(row productSkuQueryRow) *models.ProductSkuResponse {
	return &models.ProductSkuResponse{
		SpuID:             row.SpuID,
		SpuCode:           row.SpuCode,
		ProductName:       row.ProductName,
		ProductType:       row.ProductType,
		RpID:              row.RpID,
		RpCode:            row.RpCode,
		SpecName:          row.SpecName,
		MpID:              row.MpID,
		MpCode:            row.MpCode,
		EnterpriseID:      row.EnterpriseID,
		EnterpriseCode:    row.EnterpriseCode,
		EnterpriseName:    row.EnterpriseName,
		ApprovalNo:        row.ApprovalNo,
		BrandName:         row.BrandName,
		SkuID:             row.SkuID,
		SkuCode:           row.SkuCode,
		PackageSpecName:   row.PackageSpecName,
		PackConversion:    row.PackConversion,
		MinUnitName:       row.MinUnitName,
		PackageUnitName:   row.PackageUnitName,
		CartonUnitName:    row.CartonUnitName,
		CartonConversion:  row.CartonConversion,
		CartonSpecName:    row.CartonSpecName,
		FullChainSpecName: row.FullChainSpecName,
		Barcode:           row.Barcode,
		Gtin:              row.Gtin,
		UdiDi:             row.UdiDi,
		AllowSplit:        row.AllowSplit,
		Description:       row.Description,
		Status:            row.Status,
		RowVersion:        row.RowVersion,
		CreateDate:        models.TimeToStringPtr(row.CreateDate),
		UpdateDate:        models.TimeToStringPtr(row.UpdateDate),
	}
}

func validateProductSkuUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrProductSkuInvalidInput, label)
	}
	return nil
}

func validateProductSkuStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrProductSkuInvalidInput)
	}
	return nil
}

func validateProductSkuBoolFlag(value int, label string) error {
	if value != 0 && value != 1 {
		return fmt.Errorf("%w: %s只能是0或1", ErrProductSkuInvalidInput, label)
	}
	return nil
}

func buildProductSkuPackageSpecName(packConversion int, minUnitName, packageUnitName string) string {
	return fmt.Sprintf("%d%s/%s", packConversion, minUnitName, packageUnitName)
}

func buildProductSkuCartonSpecName(cartonConversion int, packageUnitName, cartonUnitName string) string {
	return fmt.Sprintf("%d%s/%s", cartonConversion, packageUnitName, cartonUnitName)
}

func buildProductSkuFullChainSpecName(cartonConversion int, cartonUnitName, packageUnitName string, packConversion int, minUnitName string) string {
	totalMinUnitQuantity := int64(cartonConversion) * int64(packConversion)
	return fmt.Sprintf("1%s/%d%s/%d%s", cartonUnitName, cartonConversion, packageUnitName, totalMinUnitQuantity, minUnitName)
}

func normalizeProductSkuOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeProductSkuPage(page, pageSize, defaultSize, maxSize int) (int, int) {
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

func optionalProductSkuOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

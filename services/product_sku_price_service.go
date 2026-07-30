package services

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	ErrProductSkuPriceInvalidInput = errors.New("SKU价格参数错误")
	ErrProductSkuPriceNotFound     = errors.New("SKU价格数据不存在")
	ErrProductSkuPriceConflict     = errors.New("SKU价格数据冲突")
)

var productPriceLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type ProductSkuPriceService struct{}

func NewProductSkuPriceService() *ProductSkuPriceService {
	return &ProductSkuPriceService{}
}

func (s *ProductSkuPriceService) GetProductSkuPriceList(skuID string) ([]models.ProductSkuPriceResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if _, err := s.getExistingProductSku(database.DB, skuID); err != nil {
		return nil, err
	}

	var rows []productSkuPriceQueryRow
	if err := s.baseProductSkuPriceQuery().
		Select(productSkuPriceSelectFields()).
		Where("product_sku_price.sku_id = ?", skuID).
		Order("product_sku_price.effective_start desc, product_sku_price.create_date desc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return productSkuPriceRowsToResponses(rows), nil
}

func (s *ProductSkuPriceService) GetProductSkuPriceTiers(skuID, priceID string) ([]models.ProductSkuPriceTierResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if err := validateProductSkuPriceUUID(priceID, "价格ID"); err != nil {
		return nil, err
	}
	if _, err := s.getExistingProductSkuPrice(database.DB, skuID, priceID); err != nil {
		return nil, err
	}

	var tiers []models.ProductSkuPriceTier
	if err := database.DB.Where("price_id = ? AND del_flag = 0", priceID).
		Order("min_quantity asc, tier_id asc").
		Find(&tiers).Error; err != nil {
		return nil, err
	}
	return productSkuPriceTiersToResponses(tiers), nil
}

func (s *ProductSkuPriceService) CreateProductSkuPrice(skuID string, req models.SaveProductSkuPriceRequest, operatorID string) (*models.ProductSkuPriceResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}
	if productSkuPriceTimeBeforeNow(normalized.EffectiveStart) {
		return nil, fmt.Errorf("%w: 生效开始时间不能早于当前时间", ErrProductSkuPriceInvalidInput)
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockExistingProductSku(tx, skuID); err != nil {
			return err
		}
		prices, err := s.lockProductSkuPrices(tx, skuID, normalized.PriceType, normalized.ScopeType, normalized.ScopeID)
		if err != nil {
			return err
		}
		if productSkuPricePeriodsOverlap(prices, normalized.EffectiveStart, normalized.EffectiveEnd, "") {
			return fmt.Errorf("%w: 同一SKU、价格类型和价格范围下生效期不能重叠", ErrProductSkuPriceConflict)
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		price := models.ProductSkuPrice{
			PriceID:        createdID,
			SkuID:          skuID,
			PriceType:      normalized.PriceType,
			ScopeType:      normalized.ScopeType,
			ScopeID:        normalized.ScopeID,
			CurrencyCode:   normalized.CurrencyCode,
			Price:          normalized.Price,
			TaxIncluded:    normalized.TaxIncluded,
			EffectiveStart: normalized.EffectiveStart,
			EffectiveEnd:   normalized.EffectiveEnd,
			Status:         normalized.Status,
			Remark:         normalized.Remark,
			RowVersion:     1,
			CreatorID:      optionalProductSkuPriceOperatorID(operatorID),
			UpdaterID:      optionalProductSkuPriceOperatorID(operatorID),
			CreateDate:     &now,
			UpdateDate:     &now,
			DelFlag:        0,
		}
		return tx.Create(&price).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuPriceDetail(skuID, createdID)
}

func (s *ProductSkuPriceService) UpdateProductSkuPrice(skuID, priceID string, req models.SaveProductSkuPriceRequest, operatorID string) (*models.ProductSkuPriceResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if err := validateProductSkuPriceUUID(priceID, "价格ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockExistingProductSku(tx, skuID); err != nil {
			return err
		}

		var current models.ProductSkuPrice
		if err := tx.Where("price_id = ? AND sku_id = ? AND del_flag = 0", priceID, skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
			}
			return err
		}
		if current.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: SKU价格已被其他人修改，请刷新后重试", ErrProductSkuPriceConflict)
		}
		if !current.EffectiveStart.Equal(normalized.EffectiveStart) && productSkuPriceTimeBeforeNow(normalized.EffectiveStart) {
			return fmt.Errorf("%w: 生效开始时间不能早于当前时间", ErrProductSkuPriceInvalidInput)
		}

		prices, err := s.lockProductSkuPrices(tx, skuID, normalized.PriceType, normalized.ScopeType, normalized.ScopeID)
		if err != nil {
			return err
		}
		if productSkuPricePeriodsOverlap(prices, normalized.EffectiveStart, normalized.EffectiveEnd, priceID) {
			return fmt.Errorf("%w: 同一SKU、价格类型和价格范围下生效期不能重叠", ErrProductSkuPriceConflict)
		}

		now := time.Now()
		return tx.Model(&models.ProductSkuPrice{}).Where("price_id = ?", priceID).Updates(map[string]interface{}{
			"price_type":      normalized.PriceType,
			"scope_type":      normalized.ScopeType,
			"scope_id":        normalized.ScopeID,
			"currency_code":   normalized.CurrencyCode,
			"price":           normalized.Price,
			"tax_included":    normalized.TaxIncluded,
			"effective_start": normalized.EffectiveStart,
			"effective_end":   normalized.EffectiveEnd,
			"status":          normalized.Status,
			"remark":          normalized.Remark,
			"row_version":     current.RowVersion + 1,
			"updater_id":      optionalProductSkuPriceOperatorID(operatorID),
			"update_date":     now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuPriceDetail(skuID, priceID)
}

func (s *ProductSkuPriceService) UpdateProductSkuPriceStatus(skuID, priceID string, req models.UpdateProductSkuPriceStatusRequest, operatorID string) (*models.ProductSkuPriceResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if err := validateProductSkuPriceUUID(priceID, "价格ID"); err != nil {
		return nil, err
	}
	if req.Status == nil {
		return nil, fmt.Errorf("%w: 状态不能为空", ErrProductSkuPriceInvalidInput)
	}
	if err := validateProductSkuPriceStatus(*req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockExistingProductSku(tx, skuID); err != nil {
			return err
		}

		var current models.ProductSkuPrice
		if err := tx.Where("price_id = ? AND sku_id = ? AND del_flag = 0", priceID, skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
			}
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: SKU价格已被其他人修改，请刷新后重试", ErrProductSkuPriceConflict)
		}
		if *req.Status == 1 {
			prices, err := s.lockProductSkuPrices(tx, skuID, current.PriceType, current.ScopeType, current.ScopeID)
			if err != nil {
				return err
			}
			if productSkuPricePeriodsOverlap(prices, current.EffectiveStart, current.EffectiveEnd, priceID) {
				return fmt.Errorf("%w: 启用后的SKU价格生效期与现有价格重叠", ErrProductSkuPriceConflict)
			}
		}

		now := time.Now()
		return tx.Model(&models.ProductSkuPrice{}).Where("price_id = ?", priceID).Updates(map[string]interface{}{
			"status":      *req.Status,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalProductSkuPriceOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuPriceDetail(skuID, priceID)
}

func (s *ProductSkuPriceService) DeleteProductSkuPrice(skuID, priceID string, req models.DeleteProductSkuPriceRequest, operatorID string) error {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return err
	}
	if err := validateProductSkuPriceUUID(priceID, "价格ID"); err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockExistingProductSku(tx, skuID); err != nil {
			return err
		}

		var current models.ProductSkuPrice
		if err := tx.Where("price_id = ? AND sku_id = ? AND del_flag = 0", priceID, skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
			}
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: SKU价格已被其他人修改，请刷新后重试", ErrProductSkuPriceConflict)
		}

		now := time.Now()
		if err := tx.Model(&models.ProductSkuPriceTier{}).Where("price_id = ? AND del_flag = 0", priceID).Updates(map[string]interface{}{
			"del_flag":    1,
			"updater_id":  optionalProductSkuPriceOperatorID(operatorID),
			"update_date": now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProductSkuPrice{}).Where("price_id = ?", priceID).Updates(map[string]interface{}{
			"del_flag":    1,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalProductSkuPriceOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

func (s *ProductSkuPriceService) SaveProductSkuPriceTiers(skuID, priceID string, req models.SaveProductSkuPriceTiersRequest, operatorID string) (*models.ProductSkuPriceResponse, error) {
	if err := validateProductSkuPriceUUID(skuID, "SKU ID"); err != nil {
		return nil, err
	}
	if err := validateProductSkuPriceUUID(priceID, "价格ID"); err != nil {
		return nil, err
	}
	normalized, err := normalizeProductSkuPriceTierItems(req.Tiers)
	if err != nil {
		return nil, err
	}
	if req.ExpectedPriceRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少价格版本号", ErrProductSkuPriceInvalidInput)
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockExistingProductSku(tx, skuID); err != nil {
			return err
		}

		var current models.ProductSkuPrice
		if err := tx.Where("price_id = ? AND sku_id = ? AND del_flag = 0", priceID, skuID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
			}
			return err
		}
		if current.RowVersion != req.ExpectedPriceRowVersion {
			return fmt.Errorf("%w: SKU价格已被其他人修改，请刷新后重试", ErrProductSkuPriceConflict)
		}

		existing, err := s.lockProductSkuPriceTiers(tx, priceID)
		if err != nil {
			return err
		}
		existingMap := make(map[string]models.ProductSkuPriceTier, len(existing))
		for _, tier := range existing {
			existingMap[tier.TierID] = tier
		}

		now := time.Now()
		keptTierIDs := make(map[string]struct{}, len(normalized))
		for _, tier := range normalized {
			if tier.TierID == nil {
				row := models.ProductSkuPriceTier{
					TierID:      utils.GenerateUUID(),
					PriceID:     priceID,
					MinQuantity: tier.MinQuantity,
					MaxQuantity: tier.MaxQuantity,
					TierPrice:   tier.TierPrice,
					CreatorID:   optionalProductSkuPriceOperatorID(operatorID),
					UpdaterID:   optionalProductSkuPriceOperatorID(operatorID),
					CreateDate:  &now,
					UpdateDate:  &now,
					DelFlag:     0,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
				keptTierIDs[row.TierID] = struct{}{}
				continue
			}

			existingTier, exists := existingMap[*tier.TierID]
			if !exists {
				return fmt.Errorf("%w: 阶梯价格不存在或不属于当前价格", ErrProductSkuPriceInvalidInput)
			}
			if err := tx.Model(&models.ProductSkuPriceTier{}).Where("tier_id = ?", existingTier.TierID).Updates(map[string]interface{}{
				"min_quantity": tier.MinQuantity,
				"max_quantity": tier.MaxQuantity,
				"tier_price":   tier.TierPrice,
				"updater_id":   optionalProductSkuPriceOperatorID(operatorID),
				"update_date":  now,
			}).Error; err != nil {
				return err
			}
			keptTierIDs[existingTier.TierID] = struct{}{}
		}

		for _, tier := range existing {
			if _, kept := keptTierIDs[tier.TierID]; kept {
				continue
			}
			if err := tx.Model(&models.ProductSkuPriceTier{}).Where("tier_id = ?", tier.TierID).Updates(map[string]interface{}{
				"del_flag":    1,
				"updater_id":  optionalProductSkuPriceOperatorID(operatorID),
				"update_date": now,
			}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&models.ProductSkuPrice{}).Where("price_id = ?", priceID).Updates(map[string]interface{}{
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalProductSkuPriceOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductSkuPriceDetail(skuID, priceID)
}

func (s *ProductSkuPriceService) GetProductSkuPriceDetail(skuID, priceID string) (*models.ProductSkuPriceResponse, error) {
	var row productSkuPriceQueryRow
	if err := s.baseProductSkuPriceQuery().
		Select(productSkuPriceSelectFields()).
		Where("product_sku_price.sku_id = ? AND product_sku_price.price_id = ?", skuID, priceID).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.PriceID == "" {
		return nil, fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
	}
	return productSkuPriceRowToResponse(row), nil
}

type normalizedProductSkuPriceSave struct {
	PriceType          string
	ScopeType          string
	ScopeID            *string
	CurrencyCode       string
	Price              string
	TaxIncluded        int
	EffectiveStart     time.Time
	EffectiveEnd       *time.Time
	Status             int
	Remark             *string
	ExpectedRowVersion int
}

type productSkuPriceQueryRow struct {
	PriceID        string
	SkuID          string
	SkuCode        string
	PriceType      string
	ScopeType      string
	ScopeID        *string
	ScopeName      *string
	CurrencyCode   string
	Price          string
	TaxIncluded    int
	EffectiveStart time.Time
	EffectiveEnd   *time.Time
	Status         int
	Remark         *string
	RowVersion     int
	TierCount      int
	CreateDate     *time.Time
	UpdateDate     *time.Time
}

type normalizedProductSkuPriceTier struct {
	TierID      *string
	MinQuantity int
	MaxQuantity *int
	TierPrice   string
}

func (s *ProductSkuPriceService) normalizeSaveRequest(req models.SaveProductSkuPriceRequest, requireVersion bool) (*normalizedProductSkuPriceSave, error) {
	priceType := strings.TrimSpace(req.PriceType)
	if priceType == "" || len([]rune(priceType)) > 32 {
		return nil, fmt.Errorf("%w: 价格类型不能为空且不能超过32个字符", ErrProductSkuPriceInvalidInput)
	}
	scopeType := strings.TrimSpace(req.ScopeType)
	if scopeType == "" || len([]rune(scopeType)) > 32 {
		return nil, fmt.Errorf("%w: 价格范围不能为空且不能超过32个字符", ErrProductSkuPriceInvalidInput)
	}
	scopeID := normalizeProductSkuPriceOptionalString(req.ScopeID)
	if strings.EqualFold(scopeType, "GLOBAL") {
		scopeID = nil
	} else if scopeID == nil {
		return nil, fmt.Errorf("%w: 非全局价格必须选择范围对象", ErrProductSkuPriceInvalidInput)
	}
	if scopeID != nil && len([]rune(*scopeID)) > 36 {
		return nil, fmt.Errorf("%w: 范围对象ID不能超过36个字符", ErrProductSkuPriceInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrProductSkuPriceInvalidInput)
	}
	if req.Status == nil {
		return nil, fmt.Errorf("%w: 状态不能为空", ErrProductSkuPriceInvalidInput)
	}
	if err := validateProductSkuPriceStatus(*req.Status); err != nil {
		return nil, err
	}
	if req.TaxIncluded == nil {
		return nil, fmt.Errorf("%w: 是否含税不能为空", ErrProductSkuPriceInvalidInput)
	}
	if *req.TaxIncluded != 0 && *req.TaxIncluded != 1 {
		return nil, fmt.Errorf("%w: 是否含税只能是0或1", ErrProductSkuPriceInvalidInput)
	}
	currencyCode := strings.ToUpper(strings.TrimSpace(req.CurrencyCode))
	if currencyCode == "" {
		currencyCode = "CNY"
	}
	if currencyCode != "CNY" {
		return nil, fmt.Errorf("%w: 第一版仅支持人民币", ErrProductSkuPriceInvalidInput)
	}
	price, err := normalizeProductSkuPriceAmount(req.Price)
	if err != nil {
		return nil, err
	}
	effectiveStart, effectiveEnd, err := normalizeProductSkuPricePeriod(req.EffectiveStart, req.EffectiveEnd)
	if err != nil {
		return nil, err
	}
	remark := normalizeProductSkuPriceOptionalString(req.Remark)
	if remark != nil && len([]rune(*remark)) > 512 {
		return nil, fmt.Errorf("%w: 备注不能超过512个字符", ErrProductSkuPriceInvalidInput)
	}

	return &normalizedProductSkuPriceSave{
		PriceType:          priceType,
		ScopeType:          scopeType,
		ScopeID:            scopeID,
		CurrencyCode:       currencyCode,
		Price:              price,
		TaxIncluded:        *req.TaxIncluded,
		EffectiveStart:     effectiveStart,
		EffectiveEnd:       effectiveEnd,
		Status:             *req.Status,
		Remark:             remark,
		ExpectedRowVersion: req.ExpectedRowVersion,
	}, nil
}

func (s *ProductSkuPriceService) baseProductSkuPriceQuery() *gorm.DB {
	return database.DB.Table("product_sku_price").
		Joins("INNER JOIN product_sku ON product_sku.sku_id = product_sku_price.sku_id AND product_sku.del_flag = 0").
		Joins("LEFT JOIN base_enterprise scope_enterprise ON scope_enterprise.enterprise_id = product_sku_price.scope_id AND scope_enterprise.del_flag = 0").
		Joins("LEFT JOIN (?) AS tier_stat ON tier_stat.price_id = product_sku_price.price_id", productSkuPriceTierCountSubquery()).
		Where("product_sku_price.del_flag = 0")
}

func (s *ProductSkuPriceService) getExistingProductSku(tx *gorm.DB, skuID string) (*models.ProductSku, error) {
	var sku models.ProductSku
	if err := tx.Where("sku_id = ? AND del_flag = 0", skuID).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: SKU不存在", ErrProductSkuPriceNotFound)
		}
		return nil, err
	}
	return &sku, nil
}

func (s *ProductSkuPriceService) lockExistingProductSku(tx *gorm.DB, skuID string) (*models.ProductSku, error) {
	var sku models.ProductSku
	if err := tx.Where("sku_id = ? AND del_flag = 0", skuID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: SKU不存在", ErrProductSkuPriceNotFound)
		}
		return nil, err
	}
	return &sku, nil
}

func (s *ProductSkuPriceService) getExistingProductSkuPrice(tx *gorm.DB, skuID, priceID string) (*models.ProductSkuPrice, error) {
	var price models.ProductSkuPrice
	if err := tx.Where("price_id = ? AND sku_id = ? AND del_flag = 0", priceID, skuID).First(&price).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: SKU价格不存在", ErrProductSkuPriceNotFound)
		}
		return nil, err
	}
	return &price, nil
}

func (s *ProductSkuPriceService) lockProductSkuPriceTiers(tx *gorm.DB, priceID string) ([]models.ProductSkuPriceTier, error) {
	var tiers []models.ProductSkuPriceTier
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("price_id = ? AND del_flag = 0", priceID).
		Order("min_quantity asc, tier_id asc").
		Find(&tiers).Error
	return tiers, err
}

func (s *ProductSkuPriceService) lockProductSkuPrices(tx *gorm.DB, skuID, priceType, scopeType string, scopeID *string) ([]models.ProductSkuPrice, error) {
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sku_id = ? AND price_type = ? AND scope_type = ? AND del_flag = 0", skuID, priceType, scopeType)
	if scopeID == nil {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", *scopeID)
	}
	var prices []models.ProductSkuPrice
	err := query.Order("effective_start asc, price_id asc").Find(&prices).Error
	return prices, err
}

func productSkuPriceSelectFields() string {
	return strings.Join([]string{
		"product_sku_price.price_id",
		"product_sku_price.sku_id",
		"product_sku.sku_code",
		"product_sku_price.price_type",
		"product_sku_price.scope_type",
		"product_sku_price.scope_id",
		"scope_enterprise.enterprise_name AS scope_name",
		"product_sku_price.currency_code",
		"product_sku_price.price",
		"product_sku_price.tax_included",
		"product_sku_price.effective_start",
		"product_sku_price.effective_end",
		"product_sku_price.status",
		"product_sku_price.remark",
		"product_sku_price.row_version",
		"COALESCE(tier_stat.tier_count, 0) AS tier_count",
		"product_sku_price.create_date",
		"product_sku_price.update_date",
	}, ", ")
}

func productSkuPriceTierCountSubquery() *gorm.DB {
	return database.DB.Table("product_sku_price_tier").
		Select("price_id, COUNT(*) AS tier_count").
		Where("del_flag = 0").
		Group("price_id")
}

func productSkuPriceRowsToResponses(rows []productSkuPriceQueryRow) []models.ProductSkuPriceResponse {
	responses := make([]models.ProductSkuPriceResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, *productSkuPriceRowToResponse(row))
	}
	return responses
}

func productSkuPriceRowToResponse(row productSkuPriceQueryRow) *models.ProductSkuPriceResponse {
	return &models.ProductSkuPriceResponse{
		PriceID:        row.PriceID,
		SkuID:          row.SkuID,
		SkuCode:        row.SkuCode,
		PriceType:      row.PriceType,
		ScopeType:      row.ScopeType,
		ScopeID:        row.ScopeID,
		ScopeName:      row.ScopeName,
		CurrencyCode:   row.CurrencyCode,
		Price:          row.Price,
		TaxIncluded:    row.TaxIncluded,
		EffectiveStart: row.EffectiveStart.Format("2006-01-02 15:04:05"),
		EffectiveEnd:   models.TimeToStringPtr(row.EffectiveEnd),
		Status:         row.Status,
		Remark:         row.Remark,
		RowVersion:     row.RowVersion,
		TierCount:      row.TierCount,
		CreateDate:     models.TimeToStringPtr(row.CreateDate),
		UpdateDate:     models.TimeToStringPtr(row.UpdateDate),
	}
}

func productSkuPriceTiersToResponses(tiers []models.ProductSkuPriceTier) []models.ProductSkuPriceTierResponse {
	responses := make([]models.ProductSkuPriceTierResponse, 0, len(tiers))
	for _, tier := range tiers {
		responses = append(responses, models.ProductSkuPriceTierResponse{
			TierID:      tier.TierID,
			PriceID:     tier.PriceID,
			MinQuantity: tier.MinQuantity,
			MaxQuantity: tier.MaxQuantity,
			TierPrice:   tier.TierPrice,
			CreateDate:  models.TimeToStringPtr(tier.CreateDate),
			UpdateDate:  models.TimeToStringPtr(tier.UpdateDate),
		})
	}
	return responses
}

func productSkuPricePeriodsOverlap(prices []models.ProductSkuPrice, effectiveStart time.Time, effectiveEnd *time.Time, excludePriceID string) bool {
	for _, price := range prices {
		if price.PriceID == excludePriceID {
			continue
		}
		if productSkuPricePeriodOverlap(price.EffectiveStart, price.EffectiveEnd, effectiveStart, effectiveEnd) {
			return true
		}
	}
	return false
}

func productSkuPricePeriodOverlap(leftStart time.Time, leftEnd *time.Time, rightStart time.Time, rightEnd *time.Time) bool {
	if leftEnd != nil && !leftEnd.After(rightStart) {
		return false
	}
	if rightEnd != nil && !rightEnd.After(leftStart) {
		return false
	}
	return true
}

func normalizeProductSkuPriceAmount(value string) (string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > 14 {
		return "", fmt.Errorf("%w: 价格金额格式不正确", ErrProductSkuPriceInvalidInput)
	}
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("%w: 价格金额格式不正确", ErrProductSkuPriceInvalidInput)
		}
	}
	intPart := strings.TrimLeft(parts[0], "0")
	if intPart == "" {
		intPart = "0"
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) == 0 || len(fraction) > 4 {
			return "", fmt.Errorf("%w: 价格金额最多保留四位小数", ErrProductSkuPriceInvalidInput)
		}
		for _, ch := range fraction {
			if ch < '0' || ch > '9' {
				return "", fmt.Errorf("%w: 价格金额格式不正确", ErrProductSkuPriceInvalidInput)
			}
		}
	}
	for len(fraction) < 4 {
		fraction += "0"
	}
	fractionValue, _ := strconv.ParseInt(fraction, 10, 64)
	if intPart == "0" && fractionValue == 0 {
		return "", fmt.Errorf("%w: 价格金额必须大于0", ErrProductSkuPriceInvalidInput)
	}
	return fmt.Sprintf("%s.%s", intPart, fraction), nil
}

func normalizeProductSkuPriceTierItems(items []models.SaveProductSkuPriceTierItem) ([]normalizedProductSkuPriceTier, error) {
	if items == nil {
		return nil, fmt.Errorf("%w: 阶梯价格数组不能为空", ErrProductSkuPriceInvalidInput)
	}
	tiers := make([]normalizedProductSkuPriceTier, 0, len(items))
	for _, item := range items {
		tierID := normalizeProductSkuPriceOptionalString(item.TierID)
		if tierID != nil {
			if err := validateProductSkuPriceUUID(*tierID, "阶梯价格ID"); err != nil {
				return nil, err
			}
		}
		if item.MinQuantity <= 0 {
			return nil, fmt.Errorf("%w: 起始数量必须为正整数", ErrProductSkuPriceInvalidInput)
		}
		if item.MaxQuantity != nil && *item.MaxQuantity < item.MinQuantity {
			return nil, fmt.Errorf("%w: 结束数量必须大于等于起始数量", ErrProductSkuPriceInvalidInput)
		}
		tierPrice, err := normalizeProductSkuPriceAmount(item.TierPrice)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, normalizedProductSkuPriceTier{
			TierID:      tierID,
			MinQuantity: item.MinQuantity,
			MaxQuantity: item.MaxQuantity,
			TierPrice:   tierPrice,
		})
	}
	sort.Slice(tiers, func(i, j int) bool {
		if tiers[i].MinQuantity == tiers[j].MinQuantity {
			return optionalProductSkuPriceMaxQuantityValue(tiers[i].MaxQuantity) < optionalProductSkuPriceMaxQuantityValue(tiers[j].MaxQuantity)
		}
		return tiers[i].MinQuantity < tiers[j].MinQuantity
	})
	if err := validateProductSkuPriceTierPeriods(tiers); err != nil {
		return nil, err
	}
	return tiers, nil
}

func validateProductSkuPriceTierPeriods(tiers []normalizedProductSkuPriceTier) error {
	seenMinQuantity := make(map[int]struct{}, len(tiers))
	var previous *normalizedProductSkuPriceTier
	for index := range tiers {
		current := tiers[index]
		if _, exists := seenMinQuantity[current.MinQuantity]; exists {
			return fmt.Errorf("%w: 阶梯起始数量不能重复", ErrProductSkuPriceConflict)
		}
		seenMinQuantity[current.MinQuantity] = struct{}{}
		if previous != nil {
			if previous.MaxQuantity == nil || current.MinQuantity <= *previous.MaxQuantity {
				return fmt.Errorf("%w: 阶梯数量区间不能重叠", ErrProductSkuPriceConflict)
			}
		}
		previous = &tiers[index]
	}
	return nil
}

func optionalProductSkuPriceMaxQuantityValue(value *int) int {
	if value == nil {
		return int(^uint(0) >> 1)
	}
	return *value
}

func normalizeProductSkuPricePeriod(effectiveStartValue string, effectiveEndValue *string) (time.Time, *time.Time, error) {
	effectiveStart, err := parseProductSkuPriceTime(effectiveStartValue, "生效开始时间")
	if err != nil {
		return time.Time{}, nil, err
	}
	effectiveEnd, err := parseProductSkuPriceOptionalTime(effectiveEndValue, "生效结束时间")
	if err != nil {
		return time.Time{}, nil, err
	}
	if effectiveEnd != nil && !effectiveEnd.After(effectiveStart) {
		return time.Time{}, nil, fmt.Errorf("%w: 生效结束时间必须晚于生效开始时间", ErrProductSkuPriceInvalidInput)
	}
	return effectiveStart, effectiveEnd, nil
}

func productSkuPriceTimeBeforeNow(value time.Time) bool {
	return value.Before(time.Now().In(productPriceLocation))
}

func parseProductSkuPriceTime(value, label string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%w: %s不能为空", ErrProductSkuPriceInvalidInput, label)
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, productPriceLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s格式错误，请使用 2006-01-02 15:04:05", ErrProductSkuPriceInvalidInput, label)
	}
	return parsed, nil
}

func parseProductSkuPriceOptionalTime(value *string, label string) (*time.Time, error) {
	normalized := normalizeProductSkuPriceOptionalString(value)
	if normalized == nil {
		return nil, nil
	}
	parsed, err := parseProductSkuPriceTime(*normalized, label)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func validateProductSkuPriceUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrProductSkuPriceInvalidInput, label)
	}
	return nil
}

func validateProductSkuPriceStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrProductSkuPriceInvalidInput)
	}
	return nil
}

func normalizeProductSkuPriceOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalProductSkuPriceOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

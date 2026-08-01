package services

import (
	"errors"
	"fmt"
	"math/big"
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
	ErrErpPurchaseInboundInvalidInput = errors.New("采购入库参数错误")
	ErrErpPurchaseInboundNotFound     = errors.New("采购入库数据不存在")
	ErrErpPurchaseInboundConflict     = errors.New("采购入库数据冲突")
)

type ErpPurchaseInboundService struct {
	inventoryService *ErpInventoryService
}

func NewErpPurchaseInboundService() *ErpPurchaseInboundService {
	return &ErpPurchaseInboundService{
		inventoryService: NewErpInventoryService(),
	}
}

func (s *ErpPurchaseInboundService) GetPurchaseInboundList(req models.ErpPurchaseInboundListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query, err := s.applyPurchaseInboundFilters(s.basePurchaseInboundQuery(database.DB), req)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"inboundNo":     "erp_purchase_inbound.inbound_no",
		"supplierName":  "supplier.enterprise_name",
		"warehouseName": "erp_warehouse.warehouse_name",
		"inboundDate":   "erp_purchase_inbound.inbound_date",
		"lineCount":     "COALESCE(item_stat.line_count, 0)",
		"totalAmount":   "COALESCE(item_stat.total_amount, 0)",
		"createDate":    "erp_purchase_inbound.create_date",
	})
	if order == "" {
		order = "erp_purchase_inbound.inbound_date desc, erp_purchase_inbound.create_date desc"
	}

	var rows []erpPurchaseInboundListQueryRow
	if err := query.Select(erpPurchaseInboundListSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{
		Items: erpPurchaseInboundListRowsToResponses(rows),
		Total: total,
	}, nil
}

func (s *ErpPurchaseInboundService) GetPurchaseInboundDetail(inboundID string) (*models.ErpPurchaseInboundResponse, error) {
	if err := validateErpPurchaseInboundUUID(inboundID, "采购入库单ID"); err != nil {
		return nil, err
	}
	return s.getPurchaseInboundDetail(database.DB, strings.TrimSpace(inboundID))
}

func (s *ErpPurchaseInboundService) CreatePurchaseInbound(req models.CreateErpPurchaseInboundRequest, operatorID string) (*models.ErpPurchaseInboundResponse, error) {
	supplierID := strings.TrimSpace(req.SupplierID)
	warehouseID := strings.TrimSpace(req.WarehouseID)
	if err := validateErpPurchaseInboundUUID(supplierID, "供应商ID"); err != nil {
		return nil, err
	}
	if err := validateErpPurchaseInboundUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}

	inboundDate, err := parsePurchaseInboundBusinessDate(req.InboundDate)
	if err != nil {
		return nil, err
	}
	items, err := normalizePurchaseInboundItems(req.Items)
	if err != nil {
		return nil, err
	}
	remark := normalizePurchaseInboundOptionalString(req.Remark)
	if remark != nil && len([]rune(*remark)) > 512 {
		return nil, fmt.Errorf("%w: 单据备注不能超过512个字符", ErrErpPurchaseInboundInvalidInput)
	}

	inboundID := utils.GenerateUUID()
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.lockEnabledSupplier(tx, supplierID); err != nil {
			return err
		}
		warehouse, err := s.inventoryService.lockEnabledWarehouse(tx, warehouseID)
		if err != nil {
			return err
		}
		inventoryItems := make([]normalizedErpInventoryInboundItem, 0, len(items))
		for _, item := range items {
			inventoryItems = append(inventoryItems, item.normalizedErpInventoryInboundItem)
		}
		skuMap, err := s.inventoryService.lockEnabledSkuMap(tx, inventoryItems)
		if err != nil {
			return err
		}

		inboundNo, err := s.nextPurchaseInboundNo(tx)
		if err != nil {
			return err
		}
		now := time.Now().In(erpInventoryLocation)
		inbound := models.ErpPurchaseInbound{
			InboundID:   inboundID,
			InboundNo:   inboundNo,
			SupplierID:  supplierID,
			WarehouseID: warehouseID,
			InboundDate: inboundDate,
			Remark:      remark,
			CreatorID:   optionalErpPurchaseInboundOperatorID(operatorID),
			CreateDate:  &now,
		}
		if err := tx.Create(&inbound).Error; err != nil {
			return err
		}

		for _, item := range items {
			itemRow := models.ErpPurchaseInboundItem{
				InboundItemID: utils.GenerateUUID(),
				InboundID:     inboundID,
				LineNo:        item.LineNo,
				SkuID:         item.SkuID,
				BatchNo:       item.BatchNo,
				ExpiryDate:    item.ExpiryDate,
				UnitCost:      item.UnitCost,
				Quantity:      item.Quantity,
				Remark:        item.Remark,
				CreateDate:    &now,
			}
			if err := tx.Create(&itemRow).Error; err != nil {
				return err
			}

			sourceBillID := inboundID
			if err := s.inventoryService.createInventoryInMovement(tx, *warehouse, skuMap[item.SkuID], item.normalizedErpInventoryInboundItem, erpInventoryInMovementContext{
				SourceBillType: models.InventorySourceBillTypePurchaseInbound,
				SourceBillID:   &sourceBillID,
				SourceBillNo:   inboundNo,
				MovementType:   models.InventoryMovementTypePurchaseIn,
			}, operatorID); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetPurchaseInboundDetail(inboundID)
}

type normalizedErpPurchaseInboundItem struct {
	normalizedErpInventoryInboundItem
	LineNo int
}

type erpPurchaseInboundListQueryRow struct {
	InboundID     string
	InboundNo     string
	SupplierID    string
	SupplierName  string
	WarehouseID   string
	WarehouseName string
	InboundDate   time.Time
	LineCount     int
	TotalAmount   string
	Remark        *string
	CreatorID     *string
	CreateDate    *time.Time
}

type erpPurchaseInboundDetailQueryRow struct {
	InboundID     string
	InboundNo     string
	SupplierID    string
	SupplierName  string
	WarehouseID   string
	WarehouseName string
	InboundDate   time.Time
	Remark        *string
	CreatorID     *string
	CreateDate    *time.Time
}

type erpPurchaseInboundItemQueryRow struct {
	InboundItemID   string
	LineNo          int
	SkuID           string
	SkuCode         string
	ProductName     string
	SpecName        string
	EnterpriseName  string
	PackageSpecName string
	PackageUnitName string
	MinUnitName     string
	BatchNo         string
	ExpiryDate      time.Time
	UnitCost        string
	Quantity        int
	Remark          *string
}

func (s *ErpPurchaseInboundService) basePurchaseInboundQuery(db *gorm.DB) *gorm.DB {
	return db.Table("erp_purchase_inbound").
		Joins("LEFT JOIN base_enterprise supplier ON supplier.enterprise_id = erp_purchase_inbound.supplier_id").
		Joins("LEFT JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_purchase_inbound.warehouse_id").
		Joins("LEFT JOIN (?) AS item_stat ON item_stat.inbound_id = erp_purchase_inbound.inbound_id", erpPurchaseInboundItemStatSubquery(db))
}

func (s *ErpPurchaseInboundService) applyPurchaseInboundFilters(query *gorm.DB, req models.ErpPurchaseInboundListRequest) (*gorm.DB, error) {
	if value := strings.TrimSpace(req.InboundNo); value != "" {
		query = query.Where("LOWER(erp_purchase_inbound.inbound_no) LIKE ?", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.SupplierID); value != "" {
		if err := validateErpPurchaseInboundUUID(value, "供应商ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_purchase_inbound.supplier_id = ?", value)
	}
	if value := strings.TrimSpace(req.WarehouseID); value != "" {
		if err := validateErpPurchaseInboundUUID(value, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_purchase_inbound.warehouse_id = ?", value)
	}
	if value := strings.TrimSpace(req.SkuCode); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM erp_purchase_inbound_item filter_item INNER JOIN product_sku filter_sku ON filter_sku.sku_id = filter_item.sku_id WHERE filter_item.inbound_id = erp_purchase_inbound.inbound_id AND LOWER(filter_sku.sku_code) LIKE ?)", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.BatchNo); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM erp_purchase_inbound_item filter_item WHERE filter_item.inbound_id = erp_purchase_inbound.inbound_id AND LOWER(filter_item.batch_no) LIKE ?)", "%"+strings.ToLower(value)+"%")
	}

	from, to, err := normalizePurchaseInboundDateRange(req.InboundDateFrom, req.InboundDateTo)
	if err != nil {
		return nil, err
	}
	if from != nil {
		query = query.Where("erp_purchase_inbound.inbound_date >= ?", *from)
	}
	if to != nil {
		query = query.Where("erp_purchase_inbound.inbound_date <= ?", *to)
	}
	return query, nil
}

func (s *ErpPurchaseInboundService) getPurchaseInboundDetail(db *gorm.DB, inboundID string) (*models.ErpPurchaseInboundResponse, error) {
	var inbound erpPurchaseInboundDetailQueryRow
	if err := s.basePurchaseInboundQuery(db).
		Select(erpPurchaseInboundDetailSelectFields()).
		Where("erp_purchase_inbound.inbound_id = ?", inboundID).
		First(&inbound).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 采购入库单不存在", ErrErpPurchaseInboundNotFound)
		}
		return nil, err
	}

	var items []erpPurchaseInboundItemQueryRow
	if err := db.Table("erp_purchase_inbound_item").
		Joins("LEFT JOIN product_sku ON product_sku.sku_id = erp_purchase_inbound_item.sku_id").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id").
		Joins("LEFT JOIN base_enterprise product_enterprise ON product_enterprise.enterprise_id = product_mp.enterprise_id").
		Select(erpPurchaseInboundItemSelectFields()).
		Where("erp_purchase_inbound_item.inbound_id = ?", inboundID).
		Order("erp_purchase_inbound_item.line_no asc").
		Scan(&items).Error; err != nil {
		return nil, err
	}

	itemResponses := make([]models.ErpPurchaseInboundItemResponse, 0, len(items))
	totalAmount := "0.0000"
	for _, item := range items {
		amount := multiplyErpInventoryAmount(item.UnitCost, item.Quantity)
		itemResponses = append(itemResponses, models.ErpPurchaseInboundItemResponse{
			InboundItemID:   item.InboundItemID,
			LineNo:          item.LineNo,
			SkuID:           item.SkuID,
			SkuCode:         item.SkuCode,
			ProductName:     item.ProductName,
			SpecName:        item.SpecName,
			EnterpriseName:  item.EnterpriseName,
			PackageSpecName: item.PackageSpecName,
			PackageUnitName: item.PackageUnitName,
			MinUnitName:     item.MinUnitName,
			BatchNo:         item.BatchNo,
			ExpiryDate:      formatErpInventoryDate(item.ExpiryDate),
			UnitCost:        item.UnitCost,
			Quantity:        item.Quantity,
			Amount:          amount,
			Remark:          item.Remark,
		})
		totalAmount = addErpInventoryAmounts(totalAmount, amount)
	}

	return &models.ErpPurchaseInboundResponse{
		InboundID:     inbound.InboundID,
		InboundNo:     inbound.InboundNo,
		SupplierID:    inbound.SupplierID,
		SupplierName:  inbound.SupplierName,
		WarehouseID:   inbound.WarehouseID,
		WarehouseName: inbound.WarehouseName,
		InboundDate:   formatErpInventoryDate(inbound.InboundDate),
		Remark:        inbound.Remark,
		CreatorID:     inbound.CreatorID,
		CreateDate:    models.TimeToStringPtr(inbound.CreateDate),
		LineCount:     len(itemResponses),
		TotalAmount:   totalAmount,
		Items:         itemResponses,
	}, nil
}

func erpPurchaseInboundItemStatSubquery(db *gorm.DB) *gorm.DB {
	return db.Table("erp_purchase_inbound_item").
		Select("inbound_id, COUNT(*) AS line_count, CAST(COALESCE(SUM(quantity * unit_cost), 0) AS CHAR) AS total_amount").
		Group("inbound_id")
}

func erpPurchaseInboundListSelectFields() string {
	return strings.Join([]string{
		"erp_purchase_inbound.inbound_id",
		"erp_purchase_inbound.inbound_no",
		"erp_purchase_inbound.supplier_id",
		"COALESCE(supplier.enterprise_name, '') AS supplier_name",
		"erp_purchase_inbound.warehouse_id",
		"COALESCE(erp_warehouse.warehouse_name, '') AS warehouse_name",
		"erp_purchase_inbound.inbound_date",
		"COALESCE(item_stat.line_count, 0) AS line_count",
		"COALESCE(item_stat.total_amount, '0.0000') AS total_amount",
		"erp_purchase_inbound.remark",
		"erp_purchase_inbound.creator_id",
		"erp_purchase_inbound.create_date",
	}, ", ")
}

func erpPurchaseInboundDetailSelectFields() string {
	return strings.Join([]string{
		"erp_purchase_inbound.inbound_id",
		"erp_purchase_inbound.inbound_no",
		"erp_purchase_inbound.supplier_id",
		"COALESCE(supplier.enterprise_name, '') AS supplier_name",
		"erp_purchase_inbound.warehouse_id",
		"COALESCE(erp_warehouse.warehouse_name, '') AS warehouse_name",
		"erp_purchase_inbound.inbound_date",
		"erp_purchase_inbound.remark",
		"erp_purchase_inbound.creator_id",
		"erp_purchase_inbound.create_date",
	}, ", ")
}

func erpPurchaseInboundItemSelectFields() string {
	return strings.Join([]string{
		"erp_purchase_inbound_item.inbound_item_id",
		"erp_purchase_inbound_item.line_no",
		"erp_purchase_inbound_item.sku_id",
		"COALESCE(product_sku.sku_code, '') AS sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"COALESCE(product_enterprise.enterprise_name, '') AS enterprise_name",
		"COALESCE(product_sku.package_spec_name, '') AS package_spec_name",
		"COALESCE(product_sku.package_unit_name, '') AS package_unit_name",
		"COALESCE(product_sku.min_unit_name, '') AS min_unit_name",
		"erp_purchase_inbound_item.batch_no",
		"erp_purchase_inbound_item.expiry_date",
		"erp_purchase_inbound_item.unit_cost",
		"erp_purchase_inbound_item.quantity",
		"erp_purchase_inbound_item.remark",
	}, ", ")
}

func erpPurchaseInboundListRowsToResponses(rows []erpPurchaseInboundListQueryRow) []models.ErpPurchaseInboundListResponse {
	responses := make([]models.ErpPurchaseInboundListResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, models.ErpPurchaseInboundListResponse{
			InboundID:     row.InboundID,
			InboundNo:     row.InboundNo,
			SupplierID:    row.SupplierID,
			SupplierName:  row.SupplierName,
			WarehouseID:   row.WarehouseID,
			WarehouseName: row.WarehouseName,
			InboundDate:   formatErpInventoryDate(row.InboundDate),
			LineCount:     row.LineCount,
			TotalAmount:   normalizePurchaseInboundAmount(row.TotalAmount),
			Remark:        row.Remark,
			CreatorID:     row.CreatorID,
			CreateDate:    models.TimeToStringPtr(row.CreateDate),
		})
	}
	return responses
}

func (s *ErpPurchaseInboundService) lockEnabledSupplier(tx *gorm.DB, supplierID string) error {
	var supplier models.BaseEnterprise
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Joins("INNER JOIN base_enterprise_role ON base_enterprise_role.enterprise_id = base_enterprise.enterprise_id AND base_enterprise_role.role_type = ?", models.EnterpriseRoleSupplier).
		Where("base_enterprise.enterprise_id = ? AND base_enterprise.del_flag = 0 AND base_enterprise.status = 1", supplierID).
		First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 供应商不存在、未启用或未配置供应商角色", ErrErpPurchaseInboundNotFound)
		}
		return err
	}
	return nil
}

func (s *ErpPurchaseInboundService) nextPurchaseInboundNo(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_PURCHASE_INBOUND", "PIN", 8)
}

func normalizePurchaseInboundItems(items []models.CreateErpPurchaseInboundItem) ([]normalizedErpPurchaseInboundItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: 采购入库明细不能为空", ErrErpPurchaseInboundInvalidInput)
	}
	if len(items) > 100 {
		return nil, fmt.Errorf("%w: 采购入库明细不能超过100行", ErrErpPurchaseInboundInvalidInput)
	}

	normalized := make([]normalizedErpPurchaseInboundItem, 0, len(items))
	seen := make(map[string]int, len(items))
	for index, item := range items {
		lineNo := index + 1
		skuID := strings.TrimSpace(item.SkuID)
		if err := validateErpPurchaseInboundUUID(skuID, fmt.Sprintf("第%d行SKU ID", lineNo)); err != nil {
			return nil, err
		}
		batchNo := strings.TrimSpace(item.BatchNo)
		if batchNo == "" {
			return nil, fmt.Errorf("%w: 第%d行批号不能为空", ErrErpPurchaseInboundInvalidInput, lineNo)
		}
		if len([]rune(batchNo)) > 64 {
			return nil, fmt.Errorf("%w: 第%d行批号不能超过64个字符", ErrErpPurchaseInboundInvalidInput, lineNo)
		}
		expiryDate, err := parsePurchaseInboundDateOnly(item.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("%w: 第%d行%s", ErrErpPurchaseInboundInvalidInput, lineNo, strings.TrimPrefix(err.Error(), ErrErpPurchaseInboundInvalidInput.Error()+": "))
		}
		unitCost, err := normalizeErpInventoryAmount(item.UnitCost)
		if err != nil {
			return nil, fmt.Errorf("%w: 第%d行成本价不合法", ErrErpPurchaseInboundInvalidInput, lineNo)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: 第%d行入库数量必须为正整数", ErrErpPurchaseInboundInvalidInput, lineNo)
		}
		if item.Quantity > 999999999 {
			return nil, fmt.Errorf("%w: 第%d行入库数量不能超过999999999", ErrErpPurchaseInboundInvalidInput, lineNo)
		}
		remark := normalizePurchaseInboundOptionalString(item.Remark)
		if remark != nil && len([]rune(*remark)) > 512 {
			return nil, fmt.Errorf("%w: 第%d行备注不能超过512个字符", ErrErpPurchaseInboundInvalidInput, lineNo)
		}

		duplicateKey := strings.Join([]string{skuID, batchNo, expiryDate.Format("2006-01-02"), unitCost}, "|")
		if previousLine, exists := seen[duplicateKey]; exists {
			return nil, fmt.Errorf("%w: 第%d行与第%d行的SKU、批号、有效期和成本价重复，请合并数量", ErrErpPurchaseInboundConflict, previousLine, lineNo)
		}
		seen[duplicateKey] = lineNo

		normalized = append(normalized, normalizedErpPurchaseInboundItem{
			normalizedErpInventoryInboundItem: normalizedErpInventoryInboundItem{
				SkuID:      skuID,
				BatchNo:    batchNo,
				ExpiryDate: expiryDate,
				UnitCost:   unitCost,
				Quantity:   item.Quantity,
				Remark:     remark,
			},
			LineNo: lineNo,
		})
	}
	return normalized, nil
}

func parsePurchaseInboundDateOnly(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%w: 日期不能为空", ErrErpPurchaseInboundInvalidInput)
	}
	parsed, err := time.ParseInLocation("2006-01-02", trimmed, erpInventoryLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: 日期格式错误，请使用 2006-01-02", ErrErpPurchaseInboundInvalidInput)
	}
	return parsed, nil
}

func parsePurchaseInboundBusinessDate(value string) (time.Time, error) {
	parsed, err := parsePurchaseInboundDateOnly(value)
	if err != nil {
		return time.Time{}, err
	}
	today := time.Now().In(erpInventoryLocation)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, erpInventoryLocation)
	if parsed.After(todayDate) {
		return time.Time{}, fmt.Errorf("%w: 入库日期不能晚于当前日期", ErrErpPurchaseInboundInvalidInput)
	}
	return parsed, nil
}

func normalizePurchaseInboundDateRange(fromValue, toValue string) (*time.Time, *time.Time, error) {
	var from, to *time.Time
	if strings.TrimSpace(fromValue) != "" {
		parsed, err := parsePurchaseInboundDateOnly(strings.TrimSpace(fromValue))
		if err != nil {
			return nil, nil, fmt.Errorf("%w: 入库开始日期%s", ErrErpPurchaseInboundInvalidInput, strings.TrimPrefix(err.Error(), ErrErpPurchaseInboundInvalidInput.Error()+": "))
		}
		from = &parsed
	}
	if strings.TrimSpace(toValue) != "" {
		parsed, err := parsePurchaseInboundDateOnly(strings.TrimSpace(toValue))
		if err != nil {
			return nil, nil, fmt.Errorf("%w: 入库结束日期%s", ErrErpPurchaseInboundInvalidInput, strings.TrimPrefix(err.Error(), ErrErpPurchaseInboundInvalidInput.Error()+": "))
		}
		to = &parsed
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("%w: 入库开始日期不能晚于结束日期", ErrErpPurchaseInboundInvalidInput)
	}
	return from, to, nil
}

func normalizePurchaseInboundOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalErpPurchaseInboundOperatorID(operatorID string) *string {
	return normalizePurchaseInboundOptionalString(&operatorID)
}

func validateErpPurchaseInboundUUID(value, label string) error {
	if _, err := uuid.Parse(strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrErpPurchaseInboundInvalidInput, label)
	}
	return nil
}

func normalizePurchaseInboundAmount(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0.0000"
	}
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		return parts[0] + ".0000"
	}
	fraction := parts[1] + "0000"
	return parts[0] + "." + fraction[:4]
}

func addErpInventoryAmounts(left, right string) string {
	leftRat, leftOK := new(big.Rat).SetString(left)
	rightRat, rightOK := new(big.Rat).SetString(right)
	if !leftOK || !rightOK {
		return "0.0000"
	}
	return leftRat.Add(leftRat, rightRat).FloatString(4)
}

package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrErpOtherOutboundInvalidInput = errors.New("其它出库参数错误")
	ErrErpOtherOutboundNotFound     = errors.New("其它出库数据不存在")
	ErrErpOtherOutboundConflict     = errors.New("其它出库数据冲突")
)

type ErpOtherOutboundService struct {
	inventoryService *ErpInventoryService
}

func NewErpOtherOutboundService() *ErpOtherOutboundService {
	return &ErpOtherOutboundService{
		inventoryService: NewErpInventoryService(),
	}
}

func (s *ErpOtherOutboundService) GetOtherOutboundList(req models.ErpOtherOutboundListRequest, permission datapermission.Permission) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := permission.Apply(s.baseOtherOutboundQuery(database.DB), "erp_other_outbound.creator_id")
	query, err := s.applyOtherOutboundFilters(query, req)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"outboundNo":    "erp_other_outbound.outbound_no",
		"warehouseName": "erp_warehouse.warehouse_name",
		"outboundDate":  "erp_other_outbound.outbound_date",
		"lineCount":     "COALESCE(item_stat.line_count, 0)",
		"createDate":    "erp_other_outbound.create_date",
	})
	if order == "" {
		order = "erp_other_outbound.outbound_date desc, erp_other_outbound.create_date desc"
	}

	var rows []erpOtherOutboundListQueryRow
	if err := query.Select(erpOtherOutboundListSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{
		Items: erpOtherOutboundListRowsToResponses(rows),
		Total: total,
	}, nil
}

func (s *ErpOtherOutboundService) GetOtherOutboundDetail(outboundID string, permission datapermission.Permission) (*models.ErpOtherOutboundResponse, error) {
	if err := validateErpOtherOutboundUUID(outboundID, "其它出库单ID"); err != nil {
		return nil, err
	}
	return s.getOtherOutboundDetail(database.DB, strings.TrimSpace(outboundID), permission)
}

func (s *ErpOtherOutboundService) CreateOtherOutbound(req models.CreateErpOtherOutboundRequest, operatorID string, permission datapermission.Permission) (*models.ErpOtherOutboundResponse, error) {
	warehouseID := strings.TrimSpace(req.WarehouseID)
	if err := validateErpOtherOutboundUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}

	outboundDate, err := parseOtherOutboundBusinessDate(req.OutboundDate)
	if err != nil {
		return nil, err
	}
	items, err := normalizeOtherOutboundItems(req.Items)
	if err != nil {
		return nil, err
	}
	remark := normalizeErpOtherOutboundOptionalString(req.Remark)
	if remark != nil && len([]rune(*remark)) > 500 {
		return nil, fmt.Errorf("%w: 单据备注不能超过500个字符", ErrErpOtherOutboundInvalidInput)
	}

	outboundID := utils.GenerateUUID()
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		warehouse, err := s.inventoryService.lockEnabledWarehouse(tx, warehouseID)
		if err != nil {
			return err
		}

		outboundNo, err := s.nextOtherOutboundNo(tx)
		if err != nil {
			return err
		}
		now := time.Now().In(erpInventoryLocation)
		outbound := models.ErpOtherOutbound{
			OutboundID:   outboundID,
			OutboundNo:   outboundNo,
			WarehouseID:  warehouseID,
			OutboundDate: outboundDate,
			Remark:       remark,
			CreatorID:    optionalErpOtherOutboundOperatorID(operatorID),
			CreateDate:   &now,
		}
		if err := tx.Create(&outbound).Error; err != nil {
			return err
		}

		processingItems := append([]normalizedErpOtherOutboundItem(nil), items...)
		sort.Slice(processingItems, func(i, j int) bool {
			return processingItems[i].BalanceID < processingItems[j].BalanceID
		})
		sourceBillID := outboundID
		for _, item := range processingItems {
			if err := s.inventoryService.ensureInventoryBalanceAccess(tx, item.BalanceID, permission); err != nil {
				return fmt.Errorf("第%d行：%w", item.LineNo, err)
			}
			itemRow := models.ErpOtherOutboundItem{
				OutboundItemID: utils.GenerateUUID(),
				OutboundID:     outboundID,
				LineNo:         item.LineNo,
				BalanceID:      item.BalanceID,
				Quantity:       item.Quantity,
				Remark:         item.Remark,
				CreateDate:     &now,
			}
			if err := tx.Create(&itemRow).Error; err != nil {
				return fmt.Errorf("第%d行：%w", item.LineNo, err)
			}

			if err := s.inventoryService.createInventoryOutMovement(tx, *warehouse, item.BalanceID, item.Quantity, item.TraceCodes, item.Remark, erpInventoryInMovementContext{
				SourceBillType: models.InventorySourceBillTypeOtherOutbound,
				SourceBillID:   &sourceBillID,
				SourceBillNo:   outboundNo,
				MovementType:   models.InventoryMovementTypeOtherOut,
			}, operatorID); err != nil {
				return fmt.Errorf("第%d行：%w", item.LineNo, err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 返回本次已成功创建的记录，不改变后续独立查询的数据范围。
	return s.getOtherOutboundDetail(database.DB, outboundID, datapermission.Permission{All: true})
}

type normalizedErpOtherOutboundItem struct {
	LineNo     int
	BalanceID  string
	Quantity   int
	TraceCodes []string
	Remark     *string
}

type erpOtherOutboundListQueryRow struct {
	OutboundID    string
	OutboundNo    string
	WarehouseID   string
	WarehouseName string
	OutboundDate  time.Time
	LineCount     int
	Remark        *string
	CreatorID     *string
	CreateDate    *time.Time
}

type erpOtherOutboundDetailQueryRow struct {
	OutboundID    string
	OutboundNo    string
	WarehouseID   string
	WarehouseName string
	OutboundDate  time.Time
	Remark        *string
	CreatorID     *string
	CreateDate    *time.Time
}

type erpOtherOutboundItemQueryRow struct {
	OutboundItemID  string
	LineNo          int
	BalanceID       string
	SkuID           string
	SkuCode         string
	ProductName     string
	SpecName        string
	EnterpriseName  string
	PackageSpecName string
	PackageUnitName string
	BatchNo         string
	ExpiryDate      time.Time
	UnitCost        string
	Quantity        int
	Remark          *string
}

func (s *ErpOtherOutboundService) baseOtherOutboundQuery(db *gorm.DB) *gorm.DB {
	return db.Table("erp_other_outbound").
		Joins("LEFT JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_other_outbound.warehouse_id").
		Joins("LEFT JOIN (?) AS item_stat ON item_stat.outbound_id = erp_other_outbound.outbound_id", erpOtherOutboundItemStatSubquery(db))
}

func (s *ErpOtherOutboundService) applyOtherOutboundFilters(query *gorm.DB, req models.ErpOtherOutboundListRequest) (*gorm.DB, error) {
	if value := strings.TrimSpace(req.OutboundNo); value != "" {
		query = query.Where("LOWER(erp_other_outbound.outbound_no) LIKE ?", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.WarehouseID); value != "" {
		if err := validateErpOtherOutboundUUID(value, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_other_outbound.warehouse_id = ?", value)
	}
	if value := strings.TrimSpace(req.SkuCode); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM erp_other_outbound_item filter_item INNER JOIN erp_inventory_balance filter_balance ON filter_balance.balance_id = filter_item.balance_id INNER JOIN product_sku filter_sku ON filter_sku.sku_id = filter_balance.sku_id WHERE filter_item.outbound_id = erp_other_outbound.outbound_id AND LOWER(filter_sku.sku_code) LIKE ?)", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.BatchNo); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM erp_other_outbound_item filter_item INNER JOIN erp_inventory_balance filter_balance ON filter_balance.balance_id = filter_item.balance_id INNER JOIN erp_inventory_batch filter_batch ON filter_batch.batch_id = filter_balance.batch_id WHERE filter_item.outbound_id = erp_other_outbound.outbound_id AND LOWER(filter_batch.batch_no) LIKE ?)", "%"+strings.ToLower(value)+"%")
	}

	from, to, err := normalizeOtherOutboundDateRange(req.OutboundDateFrom, req.OutboundDateTo)
	if err != nil {
		return nil, err
	}
	if from != nil {
		query = query.Where("erp_other_outbound.outbound_date >= ?", *from)
	}
	if to != nil {
		query = query.Where("erp_other_outbound.outbound_date <= ?", *to)
	}
	return query, nil
}

func (s *ErpOtherOutboundService) getOtherOutboundDetail(db *gorm.DB, outboundID string, permission datapermission.Permission) (*models.ErpOtherOutboundResponse, error) {
	var outbound erpOtherOutboundDetailQueryRow
	query := permission.Apply(s.baseOtherOutboundQuery(db), "erp_other_outbound.creator_id")
	if err := query.
		Select(erpOtherOutboundDetailSelectFields()).
		Where("erp_other_outbound.outbound_id = ?", outboundID).
		First(&outbound).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 其它出库单不存在", ErrErpOtherOutboundNotFound)
		}
		return nil, err
	}

	var items []erpOtherOutboundItemQueryRow
	if err := db.Table("erp_other_outbound_item").
		Joins("INNER JOIN erp_inventory_balance ON erp_inventory_balance.balance_id = erp_other_outbound_item.balance_id").
		Joins("INNER JOIN erp_inventory_batch ON erp_inventory_batch.batch_id = erp_inventory_balance.batch_id").
		Joins("LEFT JOIN product_sku ON product_sku.sku_id = erp_inventory_balance.sku_id").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id").
		Joins("LEFT JOIN base_enterprise product_enterprise ON product_enterprise.enterprise_id = product_mp.enterprise_id").
		Select(erpOtherOutboundItemSelectFields()).
		Where("erp_other_outbound_item.outbound_id = ?", outboundID).
		Order("erp_other_outbound_item.line_no asc").
		Scan(&items).Error; err != nil {
		return nil, err
	}

	itemResponses := make([]models.ErpOtherOutboundItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, models.ErpOtherOutboundItemResponse{
			OutboundItemID:  item.OutboundItemID,
			LineNo:          item.LineNo,
			BalanceID:       item.BalanceID,
			SkuID:           item.SkuID,
			SkuCode:         item.SkuCode,
			ProductName:     item.ProductName,
			SpecName:        item.SpecName,
			EnterpriseName:  item.EnterpriseName,
			PackageSpecName: item.PackageSpecName,
			PackageUnitName: item.PackageUnitName,
			BatchNo:         item.BatchNo,
			ExpiryDate:      formatErpInventoryDate(item.ExpiryDate),
			UnitCost:        item.UnitCost,
			Quantity:        item.Quantity,
			Remark:          item.Remark,
		})
	}

	return &models.ErpOtherOutboundResponse{
		OutboundID:    outbound.OutboundID,
		OutboundNo:    outbound.OutboundNo,
		WarehouseID:   outbound.WarehouseID,
		WarehouseName: outbound.WarehouseName,
		OutboundDate:  formatErpInventoryDate(outbound.OutboundDate),
		LineCount:     len(itemResponses),
		Remark:        outbound.Remark,
		CreatorID:     outbound.CreatorID,
		CreateDate:    models.TimeToStringPtr(outbound.CreateDate),
		Items:         itemResponses,
	}, nil
}

func erpOtherOutboundItemStatSubquery(db *gorm.DB) *gorm.DB {
	return db.Table("erp_other_outbound_item").
		Select("outbound_id, COUNT(*) AS line_count").
		Group("outbound_id")
}

func erpOtherOutboundListSelectFields() string {
	return strings.Join([]string{
		"erp_other_outbound.outbound_id",
		"erp_other_outbound.outbound_no",
		"erp_other_outbound.warehouse_id",
		"COALESCE(erp_warehouse.warehouse_name, '') AS warehouse_name",
		"erp_other_outbound.outbound_date",
		"COALESCE(item_stat.line_count, 0) AS line_count",
		"erp_other_outbound.remark",
		"erp_other_outbound.creator_id",
		"erp_other_outbound.create_date",
	}, ", ")
}

func erpOtherOutboundDetailSelectFields() string {
	return strings.Join([]string{
		"erp_other_outbound.outbound_id",
		"erp_other_outbound.outbound_no",
		"erp_other_outbound.warehouse_id",
		"COALESCE(erp_warehouse.warehouse_name, '') AS warehouse_name",
		"erp_other_outbound.outbound_date",
		"erp_other_outbound.remark",
		"erp_other_outbound.creator_id",
		"erp_other_outbound.create_date",
	}, ", ")
}

func erpOtherOutboundItemSelectFields() string {
	return strings.Join([]string{
		"erp_other_outbound_item.outbound_item_id",
		"erp_other_outbound_item.line_no",
		"erp_other_outbound_item.balance_id",
		"erp_inventory_balance.sku_id",
		"COALESCE(product_sku.sku_code, '') AS sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"COALESCE(product_enterprise.enterprise_name, '') AS enterprise_name",
		"COALESCE(product_sku.package_spec_name, '') AS package_spec_name",
		"COALESCE(erp_inventory_balance.package_unit_name, product_sku.package_unit_name, '') AS package_unit_name",
		"erp_inventory_batch.batch_no",
		"erp_inventory_batch.expiry_date",
		"erp_inventory_batch.unit_cost",
		"erp_other_outbound_item.quantity",
		"erp_other_outbound_item.remark",
	}, ", ")
}

func erpOtherOutboundListRowsToResponses(rows []erpOtherOutboundListQueryRow) []models.ErpOtherOutboundListResponse {
	responses := make([]models.ErpOtherOutboundListResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, models.ErpOtherOutboundListResponse{
			OutboundID:    row.OutboundID,
			OutboundNo:    row.OutboundNo,
			WarehouseID:   row.WarehouseID,
			WarehouseName: row.WarehouseName,
			OutboundDate:  formatErpInventoryDate(row.OutboundDate),
			LineCount:     row.LineCount,
			Remark:        row.Remark,
			CreatorID:     row.CreatorID,
			CreateDate:    models.TimeToStringPtr(row.CreateDate),
		})
	}
	return responses
}

func normalizeOtherOutboundItems(items []models.CreateErpOtherOutboundItem) ([]normalizedErpOtherOutboundItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: 其它出库明细不能为空", ErrErpOtherOutboundInvalidInput)
	}
	if len(items) > 100 {
		return nil, fmt.Errorf("%w: 其它出库明细不能超过100行", ErrErpOtherOutboundInvalidInput)
	}

	normalized := make([]normalizedErpOtherOutboundItem, 0, len(items))
	seen := make(map[string]int, len(items))
	traceCodeSeen := make(map[string]int)
	for lineNo, item := range items {
		currentLineNo := lineNo + 1
		balanceID := strings.TrimSpace(item.BalanceID)
		if err := validateErpOtherOutboundUUID(balanceID, fmt.Sprintf("第%d行库存余额ID", currentLineNo)); err != nil {
			return nil, err
		}
		if previousLineNo, exists := seen[balanceID]; exists {
			return nil, fmt.Errorf("%w: 第%d行与第%d行重复选择同一个库存余额", ErrErpOtherOutboundConflict, currentLineNo, previousLineNo)
		}
		seen[balanceID] = currentLineNo
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: 第%d行出库数量必须为正整数", ErrErpOtherOutboundInvalidInput, currentLineNo)
		}
		if item.Quantity > 999999999 {
			return nil, fmt.Errorf("%w: 第%d行出库数量不能超过999999999", ErrErpOtherOutboundInvalidInput, currentLineNo)
		}
		traceCodes, err := normalizeErpInventoryTraceCodes(item.TraceCodes)
		if err != nil {
			return nil, fmt.Errorf("%w: 第%d行%s", ErrErpOtherOutboundInvalidInput, currentLineNo, strings.TrimPrefix(err.Error(), ErrErpInventoryInvalidInput.Error()+": "))
		}
		for _, traceCode := range traceCodes {
			if previousLine, exists := traceCodeSeen[traceCode]; exists {
				return nil, fmt.Errorf("%w: 第%d行与第%d行存在重复追溯码%s", ErrErpOtherOutboundConflict, currentLineNo, previousLine, traceCode)
			}
			traceCodeSeen[traceCode] = currentLineNo
		}
		remark := normalizeErpOtherOutboundOptionalString(item.Remark)
		if remark != nil && len([]rune(*remark)) > 500 {
			return nil, fmt.Errorf("%w: 第%d行备注不能超过500个字符", ErrErpOtherOutboundInvalidInput, currentLineNo)
		}
		normalized = append(normalized, normalizedErpOtherOutboundItem{
			LineNo:     currentLineNo,
			BalanceID:  balanceID,
			Quantity:   item.Quantity,
			TraceCodes: traceCodes,
			Remark:     remark,
		})
	}
	return normalized, nil
}

func parseOtherOutboundBusinessDate(value string) (time.Time, error) {
	parsed, err := parseErpInventoryDate(value, "出库日期")
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s", ErrErpOtherOutboundInvalidInput, err.Error())
	}
	today := time.Now().In(erpInventoryLocation)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, erpInventoryLocation)
	if parsed.After(todayDate) {
		return time.Time{}, fmt.Errorf("%w: 出库日期不能晚于当前日期", ErrErpOtherOutboundInvalidInput)
	}
	return parsed, nil
}

func normalizeOtherOutboundDateRange(fromValue, toValue string) (*time.Time, *time.Time, error) {
	var from, to *time.Time
	if strings.TrimSpace(fromValue) != "" {
		parsed, err := parseErpInventoryDate(strings.TrimSpace(fromValue), "出库开始日期")
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s", ErrErpOtherOutboundInvalidInput, err.Error())
		}
		from = &parsed
	}
	if strings.TrimSpace(toValue) != "" {
		parsed, err := parseErpInventoryDate(strings.TrimSpace(toValue), "出库结束日期")
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s", ErrErpOtherOutboundInvalidInput, err.Error())
		}
		to = &parsed
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("%w: 出库开始日期不能晚于结束日期", ErrErpOtherOutboundInvalidInput)
	}
	return from, to, nil
}

func normalizeErpOtherOutboundOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalErpOtherOutboundOperatorID(operatorID string) *string {
	return normalizeErpOtherOutboundOptionalString(&operatorID)
}

func validateErpOtherOutboundUUID(value, label string) error {
	if err := validateErpInventoryUUID(strings.TrimSpace(value), label); err != nil {
		return fmt.Errorf("%w: %s", ErrErpOtherOutboundInvalidInput, err.Error())
	}
	return nil
}

func (s *ErpOtherOutboundService) nextOtherOutboundNo(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_OTHER_OUTBOUND", "OOUT", 8)
}

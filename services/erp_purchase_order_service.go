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
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrErpPurchaseOrderInvalidInput = errors.New("采购单参数错误")
	ErrErpPurchaseOrderNotFound     = errors.New("采购单数据不存在")
	ErrErpPurchaseOrderConflict     = errors.New("采购单数据冲突")
)

type ErpPurchaseOrderService struct{}

func NewErpPurchaseOrderService() *ErpPurchaseOrderService { return &ErpPurchaseOrderService{} }

type normalizedErpPurchaseOrderSave struct {
	SupplierID          string
	WarehouseID         string
	OrderDate           time.Time
	ExpectedArrivalDate *time.Time
	Remark              *string
	ExpectedRowVersion  int
	Items               []normalizedErpPurchaseOrderItem
}

type normalizedErpPurchaseOrderItem struct {
	LineNo          int
	SkuID           string
	OrderedQuantity int
	UnitPrice       string
	Remark          *string
}

type erpPurchaseOrderSkuSnapshot struct {
	SkuID           string
	SkuCode         string
	ProductName     string
	SpecName        string
	PackageSpecName string
	PackageUnitName string
	TraceMode       string
}

type erpPurchaseOrderListRow struct {
	PurchaseOrderID     string
	PurchaseOrderNo     string
	SupplierID          string
	SupplierName        string
	WarehouseID         string
	WarehouseName       string
	OrderDate           time.Time
	ExpectedArrivalDate *time.Time
	Status              string
	LineCount           int
	TotalAmount         string
	Remark              *string
	RowVersion          int
	CreatorID           *string
	ConfirmedBy         *string
	ConfirmedAt         *time.Time
	CreateDate          *time.Time
	UpdateDate          *time.Time
}

type erpPurchaseOrderItemRow struct {
	PurchaseOrderItemID string
	LineNo              int
	SkuID               string
	SkuCode             string
	ProductName         string
	SpecName            string
	PackageSpecName     string
	PackageUnitName     string
	TraceMode           string
	OrderedQuantity     int
	InboundQuantity     int
	UnitPrice           string
	Remark              *string
}

type erpPurchaseOrderLogRow struct {
	PurchaseOrderLogID string
	ActionType         string
	FromStatus         *string
	ToStatus           *string
	Summary            string
	RelatedInboundID   *string
	RelatedInboundNo   *string
	Reason             *string
	OperatorID         *string
	OperatorName       *string
	OperatedAt         time.Time
}

func (s *ErpPurchaseOrderService) GetPurchaseOrderList(req models.ErpPurchaseOrderListRequest, permission datapermission.Permission) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := permission.Apply(s.basePurchaseOrderQuery(database.DB), "erp_purchase_order.creator_id")
	query, err := s.applyPurchaseOrderFilters(query, req)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"purchaseOrderNo": "erp_purchase_order.purchase_order_no",
		"supplierName":    "erp_purchase_order.supplier_name_snapshot",
		"warehouseName":   "erp_warehouse.warehouse_name",
		"orderDate":       "erp_purchase_order.order_date",
		"status":          "erp_purchase_order.status",
		"lineCount":       "COALESCE(item_stat.line_count, 0)",
		"totalAmount":     "COALESCE(item_stat.total_amount, 0)",
		"createDate":      "erp_purchase_order.create_date",
		"updateDate":      "erp_purchase_order.update_date",
	})
	if order == "" {
		order = "erp_purchase_order.order_date desc, erp_purchase_order.create_date desc"
	}
	var rows []erpPurchaseOrderListRow
	if err := query.Select(erpPurchaseOrderSelectFields()).Order(order).
		Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &utils.PaginationResponse{Items: erpPurchaseOrderRowsToResponses(rows), Total: total}, nil
}

func (s *ErpPurchaseOrderService) GetPurchaseOrderDetail(purchaseOrderID string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	if err := validateErpPurchaseOrderUUID(purchaseOrderID, "采购单ID"); err != nil {
		return nil, err
	}
	return s.getPurchaseOrderDetail(database.DB, strings.TrimSpace(purchaseOrderID), permission)
}

func (s *ErpPurchaseOrderService) GetPurchaseOrderLogs(purchaseOrderID string, permission datapermission.Permission) ([]models.ErpPurchaseOrderLogResponse, error) {
	if err := validateErpPurchaseOrderUUID(purchaseOrderID, "采购单ID"); err != nil {
		return nil, err
	}
	var count int64
	parentQuery := database.DB.Model(&models.ErpPurchaseOrder{}).Where("purchase_order_id = ?", strings.TrimSpace(purchaseOrderID))
	if err := permission.Apply(parentQuery, "erp_purchase_order.creator_id").Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("%w: 采购单不存在", ErrErpPurchaseOrderNotFound)
	}
	var rows []erpPurchaseOrderLogRow
	if err := database.DB.Table("erp_purchase_order_log log").
		Joins("LEFT JOIN sys_user operator ON operator.user_id = log.operator_id").
		Select(strings.Join([]string{
			"log.purchase_order_log_id", "log.action_type", "log.from_status", "log.to_status", "log.summary",
			"log.related_inbound_id", "log.related_inbound_no", "log.reason", "log.operator_id",
			"COALESCE(operator.real_name, operator.username) AS operator_name", "log.operated_at",
		}, ", ")).
		Where("log.purchase_order_id = ?", strings.TrimSpace(purchaseOrderID)).
		Order("log.operated_at asc, log.purchase_order_log_id asc").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]models.ErpPurchaseOrderLogResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.ErpPurchaseOrderLogResponse{
			PurchaseOrderLogID: row.PurchaseOrderLogID, ActionType: row.ActionType, FromStatus: row.FromStatus,
			ToStatus: row.ToStatus, Summary: row.Summary, RelatedInboundID: row.RelatedInboundID,
			RelatedInboundNo: row.RelatedInboundNo, Reason: row.Reason, OperatorID: row.OperatorID,
			OperatorName: row.OperatorName, OperatedAt: row.OperatedAt.In(erpInventoryLocation).Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *ErpPurchaseOrderService) CreatePurchaseOrder(req models.SaveErpPurchaseOrderRequest, operatorID string) (*models.ErpPurchaseOrderResponse, error) {
	normalized, err := normalizeErpPurchaseOrderSave(req, false)
	if err != nil {
		return nil, err
	}
	purchaseOrderID := utils.GenerateUUID()
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		supplier, warehouse, skuMap, err := s.lockPurchaseOrderReferences(tx, normalized)
		if err != nil {
			return err
		}
		purchaseOrderNo, err := NewBaseCodeSequenceService().NextBusinessCode(tx, "PURCHASE_ORDER", "PO", 6)
		if err != nil {
			return err
		}
		now := time.Now().In(erpInventoryLocation)
		row := models.ErpPurchaseOrder{
			PurchaseOrderID: purchaseOrderID, PurchaseOrderNo: purchaseOrderNo, SupplierID: normalized.SupplierID,
			SupplierNameSnapshot: supplier.EnterpriseName, WarehouseID: warehouse.WarehouseID, OrderDate: normalized.OrderDate,
			ExpectedArrivalDate: normalized.ExpectedArrivalDate, Status: models.ErpPurchaseOrderStatusDraft,
			Remark: normalized.Remark, RowVersion: 1, CreatorID: optionalErpPurchaseInboundOperatorID(operatorID),
			UpdaterID: optionalErpPurchaseInboundOperatorID(operatorID), CreateDate: &now, UpdateDate: &now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if err := s.replacePurchaseOrderItems(tx, purchaseOrderID, normalized.Items, skuMap, now); err != nil {
			return err
		}
		return s.createPurchaseOrderLog(tx, purchaseOrderID, models.ErpPurchaseOrderLogActionCreate, nil,
			purchaseOrderStringPtr(models.ErpPurchaseOrderStatusDraft), "创建采购单草稿", nil, nil, operatorID, now)
	}); err != nil {
		return nil, err
	}
	// 返回本次已成功创建的记录，不改变后续独立查询的数据范围。
	return s.getPurchaseOrderDetail(database.DB, purchaseOrderID, datapermission.Permission{All: true})
}

func (s *ErpPurchaseOrderService) UpdatePurchaseOrder(purchaseOrderID string, req models.SaveErpPurchaseOrderRequest, operatorID string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	if err := validateErpPurchaseOrderUUID(purchaseOrderID, "采购单ID"); err != nil {
		return nil, err
	}
	normalized, err := normalizeErpPurchaseOrderSave(req, true)
	if err != nil {
		return nil, err
	}
	purchaseOrderID = strings.TrimSpace(purchaseOrderID)
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var order models.ErpPurchaseOrder
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.ErpPurchaseOrder{}).Where("purchase_order_id = ?", purchaseOrderID)
		if err := permission.Apply(query, "erp_purchase_order.creator_id").First(&order).Error; err != nil {
			return erpPurchaseOrderRecordError(err)
		}
		if order.Status != models.ErpPurchaseOrderStatusDraft {
			return fmt.Errorf("%w: 只有草稿采购单可以修改", ErrErpPurchaseOrderConflict)
		}
		if order.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 采购单已被其他人修改，请刷新后重试", ErrErpPurchaseOrderConflict)
		}
		supplier, warehouse, skuMap, err := s.lockPurchaseOrderReferences(tx, normalized)
		if err != nil {
			return err
		}
		now := time.Now().In(erpInventoryLocation)
		updates := map[string]any{
			"supplier_id": normalized.SupplierID, "supplier_name_snapshot": supplier.EnterpriseName,
			"warehouse_id": warehouse.WarehouseID, "order_date": normalized.OrderDate,
			"expected_arrival_date": normalized.ExpectedArrivalDate, "remark": normalized.Remark,
			"row_version": order.RowVersion + 1, "updater_id": optionalErpPurchaseInboundOperatorID(operatorID), "update_date": now,
		}
		if err := tx.Model(&models.ErpPurchaseOrder{}).Where("purchase_order_id = ?", purchaseOrderID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("purchase_order_id = ?", purchaseOrderID).Delete(&models.ErpPurchaseOrderItem{}).Error; err != nil {
			return err
		}
		if err := s.replacePurchaseOrderItems(tx, purchaseOrderID, normalized.Items, skuMap, now); err != nil {
			return err
		}
		return s.createPurchaseOrderLog(tx, purchaseOrderID, models.ErpPurchaseOrderLogActionUpdate,
			purchaseOrderStringPtr(order.Status), purchaseOrderStringPtr(order.Status), "修改采购单草稿", nil, nil, operatorID, now)
	}); err != nil {
		return nil, err
	}
	return s.GetPurchaseOrderDetail(purchaseOrderID, permission)
}

func (s *ErpPurchaseOrderService) ConfirmPurchaseOrder(purchaseOrderID, operatorID string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	return s.changePurchaseOrderStatus(purchaseOrderID, operatorID, "", models.ErpPurchaseOrderLogActionConfirm, permission)
}

func (s *ErpPurchaseOrderService) CancelPurchaseOrder(purchaseOrderID, operatorID, reason string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	return s.changePurchaseOrderStatus(purchaseOrderID, operatorID, reason, models.ErpPurchaseOrderLogActionCancel, permission)
}

func (s *ErpPurchaseOrderService) ClosePurchaseOrder(purchaseOrderID, operatorID, reason string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	return s.changePurchaseOrderStatus(purchaseOrderID, operatorID, reason, models.ErpPurchaseOrderLogActionClose, permission)
}

func (s *ErpPurchaseOrderService) changePurchaseOrderStatus(purchaseOrderID, operatorID, reason, action string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	if err := validateErpPurchaseOrderUUID(purchaseOrderID, "采购单ID"); err != nil {
		return nil, err
	}
	purchaseOrderID = strings.TrimSpace(purchaseOrderID)
	normalizedReason := normalizePurchaseInboundOptionalString(&reason)
	if action != models.ErpPurchaseOrderLogActionConfirm {
		if normalizedReason == nil {
			return nil, fmt.Errorf("%w: 原因不能为空", ErrErpPurchaseOrderInvalidInput)
		}
		if len([]rune(*normalizedReason)) > 500 {
			return nil, fmt.Errorf("%w: 原因不能超过500个字符", ErrErpPurchaseOrderInvalidInput)
		}
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var order models.ErpPurchaseOrder
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.ErpPurchaseOrder{}).Where("purchase_order_id = ?", purchaseOrderID)
		if err := permission.Apply(query, "erp_purchase_order.creator_id").First(&order).Error; err != nil {
			return erpPurchaseOrderRecordError(err)
		}
		fromStatus := order.Status
		toStatus := ""
		summary := ""
		switch action {
		case models.ErpPurchaseOrderLogActionConfirm:
			if order.Status != models.ErpPurchaseOrderStatusDraft {
				return fmt.Errorf("%w: 只有草稿采购单可以确认", ErrErpPurchaseOrderConflict)
			}
			if err := s.refreshAndValidatePurchaseOrderSnapshots(tx, &order); err != nil {
				return err
			}
			toStatus, summary = models.ErpPurchaseOrderStatusWaitingReceipt, "确认采购单，进入待收货"
		case models.ErpPurchaseOrderLogActionCancel:
			if order.Status != models.ErpPurchaseOrderStatusDraft && order.Status != models.ErpPurchaseOrderStatusWaitingReceipt {
				return fmt.Errorf("%w: 只有草稿或待收货采购单可以取消", ErrErpPurchaseOrderConflict)
			}
			toStatus, summary = models.ErpPurchaseOrderStatusCancelled, "取消采购单"
		case models.ErpPurchaseOrderLogActionClose:
			if order.Status != models.ErpPurchaseOrderStatusPartialReceipt {
				return fmt.Errorf("%w: 只有部分入库采购单可以关闭", ErrErpPurchaseOrderConflict)
			}
			toStatus, summary = models.ErpPurchaseOrderStatusClosed, "关闭采购单，剩余数量不再收货"
		default:
			return fmt.Errorf("%w: 未知采购单操作", ErrErpPurchaseOrderInvalidInput)
		}
		now := time.Now().In(erpInventoryLocation)
		updates := map[string]any{"status": toStatus, "row_version": order.RowVersion + 1, "updater_id": optionalErpPurchaseInboundOperatorID(operatorID), "update_date": now}
		if action == models.ErpPurchaseOrderLogActionConfirm {
			updates["confirmed_by"] = optionalErpPurchaseInboundOperatorID(operatorID)
			updates["confirmed_at"] = now
		}
		if err := tx.Model(&models.ErpPurchaseOrder{}).Where("purchase_order_id = ?", purchaseOrderID).Updates(updates).Error; err != nil {
			return err
		}
		return s.createPurchaseOrderLog(tx, purchaseOrderID, action, purchaseOrderStringPtr(fromStatus), purchaseOrderStringPtr(toStatus), summary, nil, normalizedReason, operatorID, now)
	}); err != nil {
		return nil, err
	}
	return s.GetPurchaseOrderDetail(purchaseOrderID, permission)
}

func (s *ErpPurchaseOrderService) basePurchaseOrderQuery(db *gorm.DB) *gorm.DB {
	return db.Table("erp_purchase_order").
		Joins("LEFT JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_purchase_order.warehouse_id").
		Joins("LEFT JOIN (?) item_stat ON item_stat.purchase_order_id = erp_purchase_order.purchase_order_id", db.Table("erp_purchase_order_item").Select("purchase_order_id, COUNT(*) line_count, CAST(COALESCE(SUM(ordered_quantity * unit_price), 0) AS CHAR) total_amount").Group("purchase_order_id"))
}

func (s *ErpPurchaseOrderService) applyPurchaseOrderFilters(query *gorm.DB, req models.ErpPurchaseOrderListRequest) (*gorm.DB, error) {
	if value := strings.TrimSpace(req.PurchaseOrderNo); value != "" {
		query = query.Where("LOWER(erp_purchase_order.purchase_order_no) LIKE ?", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.SupplierID); value != "" {
		if err := validateErpPurchaseOrderUUID(value, "供应商ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_purchase_order.supplier_id = ?", value)
	}
	if value := strings.TrimSpace(req.WarehouseID); value != "" {
		if err := validateErpPurchaseOrderUUID(value, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_purchase_order.warehouse_id = ?", value)
	}
	if value := strings.TrimSpace(req.SkuCode); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM erp_purchase_order_item filter_item WHERE filter_item.purchase_order_id = erp_purchase_order.purchase_order_id AND LOWER(filter_item.sku_code_snapshot) LIKE ?)", "%"+strings.ToLower(value)+"%")
	}
	if value := strings.TrimSpace(req.Status); value != "" {
		if !isErpPurchaseOrderStatus(value) {
			return nil, fmt.Errorf("%w: 采购单状态不合法", ErrErpPurchaseOrderInvalidInput)
		}
		query = query.Where("erp_purchase_order.status = ?", value)
	}
	from, to, err := normalizePurchaseInboundDateRange(req.OrderDateFrom, req.OrderDateTo)
	if err != nil {
		return nil, fmt.Errorf("%w: 采购日期范围不合法", ErrErpPurchaseOrderInvalidInput)
	}
	if from != nil {
		query = query.Where("erp_purchase_order.order_date >= ?", *from)
	}
	if to != nil {
		query = query.Where("erp_purchase_order.order_date <= ?", *to)
	}
	return query, nil
}

func (s *ErpPurchaseOrderService) getPurchaseOrderDetail(db *gorm.DB, purchaseOrderID string, permission datapermission.Permission) (*models.ErpPurchaseOrderResponse, error) {
	var row erpPurchaseOrderListRow
	query := s.basePurchaseOrderQuery(db).Where("erp_purchase_order.purchase_order_id = ?", purchaseOrderID)
	if err := permission.Apply(query, "erp_purchase_order.creator_id").Select(erpPurchaseOrderSelectFields()).First(&row).Error; err != nil {
		return nil, erpPurchaseOrderRecordError(err)
	}
	var itemRows []erpPurchaseOrderItemRow
	if err := db.Table("erp_purchase_order_item item").
		Joins("LEFT JOIN product_sku ON product_sku.sku_id = item.sku_id").
		Select(strings.Join([]string{"item.purchase_order_item_id", "item.line_no", "item.sku_id", "item.sku_code_snapshot AS sku_code", "item.product_name_snapshot AS product_name", "item.spec_name_snapshot AS spec_name", "item.package_spec_snapshot AS package_spec_name", "item.package_unit_snapshot AS package_unit_name", "COALESCE(product_sku.trace_mode, 'NONE') AS trace_mode", "item.ordered_quantity", "item.inbound_quantity", "item.unit_price", "item.remark"}, ", ")).
		Where("item.purchase_order_id = ?", purchaseOrderID).Order("item.line_no asc").Scan(&itemRows).Error; err != nil {
		return nil, err
	}
	list := erpPurchaseOrderRowToResponse(row)
	items := make([]models.ErpPurchaseOrderItemResponse, 0, len(itemRows))
	for _, item := range itemRows {
		remaining := item.OrderedQuantity - item.InboundQuantity
		if remaining < 0 {
			remaining = 0
		}
		items = append(items, models.ErpPurchaseOrderItemResponse{
			PurchaseOrderItemID: item.PurchaseOrderItemID, LineNo: item.LineNo, SkuID: item.SkuID,
			SkuCode: item.SkuCode, ProductName: item.ProductName, SpecName: item.SpecName,
			PackageSpecName: item.PackageSpecName, PackageUnitName: item.PackageUnitName, TraceMode: item.TraceMode,
			OrderedQuantity: item.OrderedQuantity, InboundQuantity: item.InboundQuantity, RemainingQuantity: remaining,
			UnitPrice: item.UnitPrice, Amount: multiplyErpInventoryAmount(item.UnitPrice, item.OrderedQuantity), Remark: item.Remark,
		})
	}
	return &models.ErpPurchaseOrderResponse{ErpPurchaseOrderListResponse: list, ConfirmedBy: row.ConfirmedBy, ConfirmedAt: models.TimeToStringPtr(row.ConfirmedAt), Items: items}, nil
}

func (s *ErpPurchaseOrderService) lockPurchaseOrderReferences(tx *gorm.DB, normalized *normalizedErpPurchaseOrderSave) (*models.BaseEnterprise, *models.ErpWarehouse, map[string]erpPurchaseOrderSkuSnapshot, error) {
	var supplier models.BaseEnterprise
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Joins("INNER JOIN base_enterprise_role ON base_enterprise_role.enterprise_id = base_enterprise.enterprise_id AND base_enterprise_role.role_type = ?", models.EnterpriseRoleSupplier).
		Where("base_enterprise.enterprise_id = ? AND base_enterprise.del_flag = 0 AND base_enterprise.status = 1", normalized.SupplierID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, fmt.Errorf("%w: 供应商不存在、未启用或未配置供应商角色", ErrErpPurchaseOrderNotFound)
		}
		return nil, nil, nil, err
	}
	var warehouse models.ErpWarehouse
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("warehouse_id = ? AND del_flag = 0 AND status = 1", normalized.WarehouseID).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, fmt.Errorf("%w: 仓库不存在或未启用", ErrErpPurchaseOrderNotFound)
		}
		return nil, nil, nil, err
	}
	skuIDs := make([]string, 0, len(normalized.Items))
	for _, item := range normalized.Items {
		skuIDs = append(skuIDs, item.SkuID)
	}
	skuMap, err := s.loadEnabledSkuSnapshots(tx, skuIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	return &supplier, &warehouse, skuMap, nil
}

func (s *ErpPurchaseOrderService) loadEnabledSkuSnapshots(tx *gorm.DB, skuIDs []string) (map[string]erpPurchaseOrderSkuSnapshot, error) {
	var rows []erpPurchaseOrderSkuSnapshot
	if err := tx.Table("product_sku").Clauses(clause.Locking{Strength: "UPDATE"}).Joins("INNER JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0 AND product_mp.status = 1").Joins("INNER JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0 AND product_rp.status = 1").Joins("INNER JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0 AND product_spu.status = 1").
		Select(strings.Join([]string{"product_sku.sku_id", "product_sku.sku_code", "product_spu.product_name", "product_rp.spec_name", "product_sku.package_spec_name", "product_sku.package_unit_name", "product_sku.trace_mode"}, ", ")).
		Where("product_sku.sku_id IN ? AND product_sku.del_flag = 0 AND product_sku.status = 1", skuIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]erpPurchaseOrderSkuSnapshot, len(rows))
	for _, row := range rows {
		result[row.SkuID] = row
	}
	if len(result) != len(skuIDs) {
		return nil, fmt.Errorf("%w: SKU不存在或未启用", ErrErpPurchaseOrderNotFound)
	}
	return result, nil
}

func (s *ErpPurchaseOrderService) replacePurchaseOrderItems(tx *gorm.DB, purchaseOrderID string, items []normalizedErpPurchaseOrderItem, skuMap map[string]erpPurchaseOrderSkuSnapshot, now time.Time) error {
	for _, item := range items {
		snapshot := skuMap[item.SkuID]
		row := models.ErpPurchaseOrderItem{
			PurchaseOrderItemID: utils.GenerateUUID(), PurchaseOrderID: purchaseOrderID, LineNo: item.LineNo,
			SkuID: item.SkuID, OrderedQuantity: item.OrderedQuantity, InboundQuantity: 0, UnitPrice: item.UnitPrice,
			SkuCodeSnapshot: snapshot.SkuCode, ProductNameSnapshot: snapshot.ProductName, SpecNameSnapshot: snapshot.SpecName,
			PackageSpecSnapshot: snapshot.PackageSpecName, PackageUnitSnapshot: snapshot.PackageUnitName,
			Remark: item.Remark, CreateDate: &now, UpdateDate: &now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *ErpPurchaseOrderService) refreshAndValidatePurchaseOrderSnapshots(tx *gorm.DB, order *models.ErpPurchaseOrder) error {
	normalized := &normalizedErpPurchaseOrderSave{SupplierID: order.SupplierID, WarehouseID: order.WarehouseID}
	var items []models.ErpPurchaseOrderItem
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("purchase_order_id = ?", order.PurchaseOrderID).Order("line_no asc").Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("%w: 采购单明细不能为空", ErrErpPurchaseOrderConflict)
	}
	for _, item := range items {
		normalized.Items = append(normalized.Items, normalizedErpPurchaseOrderItem{SkuID: item.SkuID})
	}
	supplier, _, skuMap, err := s.lockPurchaseOrderReferences(tx, normalized)
	if err != nil {
		return err
	}
	if err := tx.Model(order).Update("supplier_name_snapshot", supplier.EnterpriseName).Error; err != nil {
		return err
	}
	for _, item := range items {
		snapshot := skuMap[item.SkuID]
		if err := tx.Model(&models.ErpPurchaseOrderItem{}).Where("purchase_order_item_id = ?", item.PurchaseOrderItemID).Updates(map[string]any{
			"sku_code_snapshot": snapshot.SkuCode, "product_name_snapshot": snapshot.ProductName,
			"spec_name_snapshot": snapshot.SpecName, "package_spec_snapshot": snapshot.PackageSpecName,
			"package_unit_snapshot": snapshot.PackageUnitName,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *ErpPurchaseOrderService) createPurchaseOrderLog(tx *gorm.DB, purchaseOrderID, action string, fromStatus, toStatus *string, summary string, relatedInbound *models.ErpPurchaseInbound, reason *string, operatorID string, now time.Time) error {
	var relatedID, relatedNo *string
	if relatedInbound != nil {
		relatedID, relatedNo = &relatedInbound.InboundID, &relatedInbound.InboundNo
	}
	return tx.Create(&models.ErpPurchaseOrderLog{
		PurchaseOrderLogID: utils.GenerateUUID(), PurchaseOrderID: purchaseOrderID, ActionType: action,
		FromStatus: fromStatus, ToStatus: toStatus, Summary: summary, RelatedInboundID: relatedID,
		RelatedInboundNo: relatedNo, Reason: reason, OperatorID: optionalErpPurchaseInboundOperatorID(operatorID), OperatedAt: now,
	}).Error
}

func normalizeErpPurchaseOrderSave(req models.SaveErpPurchaseOrderRequest, requireVersion bool) (*normalizedErpPurchaseOrderSave, error) {
	supplierID, warehouseID := strings.TrimSpace(req.SupplierID), strings.TrimSpace(req.WarehouseID)
	if err := validateErpPurchaseOrderUUID(supplierID, "供应商ID"); err != nil {
		return nil, err
	}
	if err := validateErpPurchaseOrderUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	orderDate, err := parsePurchaseInboundDateOnly(req.OrderDate)
	if err != nil {
		return nil, fmt.Errorf("%w: 采购日期格式错误", ErrErpPurchaseOrderInvalidInput)
	}
	var expectedArrivalDate *time.Time
	if req.ExpectedArrivalDate != nil && strings.TrimSpace(*req.ExpectedArrivalDate) != "" {
		parsed, err := parsePurchaseInboundDateOnly(*req.ExpectedArrivalDate)
		if err != nil {
			return nil, fmt.Errorf("%w: 预计到货日期格式错误", ErrErpPurchaseOrderInvalidInput)
		}
		expectedArrivalDate = &parsed
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 采购单版本号不能为空", ErrErpPurchaseOrderInvalidInput)
	}
	remark := normalizePurchaseInboundOptionalString(req.Remark)
	if remark != nil && len([]rune(*remark)) > 512 {
		return nil, fmt.Errorf("%w: 单据备注不能超过512个字符", ErrErpPurchaseOrderInvalidInput)
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: 采购明细不能为空", ErrErpPurchaseOrderInvalidInput)
	}
	if len(req.Items) > 100 {
		return nil, fmt.Errorf("%w: 采购明细不能超过100行", ErrErpPurchaseOrderInvalidInput)
	}
	items := make([]normalizedErpPurchaseOrderItem, 0, len(req.Items))
	seen := make(map[string]int, len(req.Items))
	for index, item := range req.Items {
		lineNo := index + 1
		skuID := strings.TrimSpace(item.SkuID)
		if err := validateErpPurchaseOrderUUID(skuID, fmt.Sprintf("第%d行SKU ID", lineNo)); err != nil {
			return nil, err
		}
		if previous, exists := seen[skuID]; exists {
			return nil, fmt.Errorf("%w: 第%d行与第%d行SKU重复", ErrErpPurchaseOrderConflict, lineNo, previous)
		}
		seen[skuID] = lineNo
		if item.OrderedQuantity <= 0 || item.OrderedQuantity > 999999999 {
			return nil, fmt.Errorf("%w: 第%d行采购数量必须为1至999999999的整数", ErrErpPurchaseOrderInvalidInput, lineNo)
		}
		unitPrice, err := normalizeErpInventoryAmount(item.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("%w: 第%d行采购单价不合法", ErrErpPurchaseOrderInvalidInput, lineNo)
		}
		itemRemark := normalizePurchaseInboundOptionalString(item.Remark)
		if itemRemark != nil && len([]rune(*itemRemark)) > 512 {
			return nil, fmt.Errorf("%w: 第%d行备注不能超过512个字符", ErrErpPurchaseOrderInvalidInput, lineNo)
		}
		items = append(items, normalizedErpPurchaseOrderItem{LineNo: lineNo, SkuID: skuID, OrderedQuantity: item.OrderedQuantity, UnitPrice: unitPrice, Remark: itemRemark})
	}
	return &normalizedErpPurchaseOrderSave{SupplierID: supplierID, WarehouseID: warehouseID, OrderDate: orderDate, ExpectedArrivalDate: expectedArrivalDate, Remark: remark, ExpectedRowVersion: req.ExpectedRowVersion, Items: items}, nil
}

func erpPurchaseOrderSelectFields() string {
	return strings.Join([]string{
		"erp_purchase_order.purchase_order_id", "erp_purchase_order.purchase_order_no", "erp_purchase_order.supplier_id",
		"erp_purchase_order.supplier_name_snapshot AS supplier_name", "erp_purchase_order.warehouse_id",
		"COALESCE(erp_warehouse.warehouse_name, '') AS warehouse_name", "erp_purchase_order.order_date",
		"erp_purchase_order.expected_arrival_date", "erp_purchase_order.status", "COALESCE(item_stat.line_count, 0) AS line_count",
		"COALESCE(item_stat.total_amount, '0.0000') AS total_amount", "erp_purchase_order.remark", "erp_purchase_order.row_version",
		"erp_purchase_order.creator_id", "erp_purchase_order.confirmed_by", "erp_purchase_order.confirmed_at",
		"erp_purchase_order.create_date", "erp_purchase_order.update_date",
	}, ", ")
}

func erpPurchaseOrderRowsToResponses(rows []erpPurchaseOrderListRow) []models.ErpPurchaseOrderListResponse {
	result := make([]models.ErpPurchaseOrderListResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, erpPurchaseOrderRowToResponse(row))
	}
	return result
}

func erpPurchaseOrderRowToResponse(row erpPurchaseOrderListRow) models.ErpPurchaseOrderListResponse {
	return models.ErpPurchaseOrderListResponse{
		PurchaseOrderID: row.PurchaseOrderID, PurchaseOrderNo: row.PurchaseOrderNo, SupplierID: row.SupplierID,
		SupplierName: row.SupplierName, WarehouseID: row.WarehouseID, WarehouseName: row.WarehouseName,
		OrderDate: formatErpInventoryDate(row.OrderDate), ExpectedArrivalDate: formatOptionalErpPurchaseOrderDate(row.ExpectedArrivalDate),
		Status: row.Status, LineCount: row.LineCount, TotalAmount: normalizePurchaseInboundAmount(row.TotalAmount),
		Remark: row.Remark, RowVersion: row.RowVersion, CreatorID: row.CreatorID,
		CreateDate: models.TimeToStringPtr(row.CreateDate), UpdateDate: models.TimeToStringPtr(row.UpdateDate),
	}
}

func formatOptionalErpPurchaseOrderDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatErpInventoryDate(*value)
	return &formatted
}

func validateErpPurchaseOrderUUID(value, label string) error {
	if _, err := uuid.Parse(strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrErpPurchaseOrderInvalidInput, label)
	}
	return nil
}

func isErpPurchaseOrderStatus(value string) bool {
	switch value {
	case models.ErpPurchaseOrderStatusDraft, models.ErpPurchaseOrderStatusWaitingReceipt,
		models.ErpPurchaseOrderStatusPartialReceipt, models.ErpPurchaseOrderStatusCompleted,
		models.ErpPurchaseOrderStatusCancelled, models.ErpPurchaseOrderStatusClosed:
		return true
	default:
		return false
	}
}

func erpPurchaseOrderRecordError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: 采购单不存在", ErrErpPurchaseOrderNotFound)
	}
	return err
}

func purchaseOrderStringPtr(value string) *string { return &value }

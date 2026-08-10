package services

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrErpInventoryInvalidInput = errors.New("库存参数错误")
	ErrErpInventoryNotFound     = errors.New("库存数据不存在")
	ErrErpInventoryConflict     = errors.New("库存数据冲突")
)

type ErpInventoryInsufficientError struct {
	SkuCode         string
	BatchNo         string
	Requested       int
	Available       int
	PackageUnitName string
}

func (e *ErpInventoryInsufficientError) Error() string {
	return fmt.Sprintf("SKU %s 批号 %s 请求出库%d%s，当前可用%d%s", e.SkuCode, e.BatchNo, e.Requested, e.PackageUnitName, e.Available, e.PackageUnitName)
}

func (e *ErpInventoryInsufficientError) Unwrap() error {
	return ErrErpInventoryConflict
}

var erpInventoryLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type ErpInventoryService struct{}

func NewErpInventoryService() *ErpInventoryService {
	return &ErpInventoryService{}
}

func (s *ErpInventoryService) GetInventoryBalanceList(req models.ErpInventoryBalanceListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := s.baseInventoryBalanceQuery()
	var err error
	query, err = s.applyInventoryBalanceFilters(query, req)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := buildInventoryBalanceOrder(req.Sorts)
	if order == "" {
		order = "erp_inventory_balance.update_date desc, erp_inventory_balance.create_date desc"
	}

	var rows []erpInventoryBalanceQueryRow
	if err := query.Select(erpInventoryBalanceSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: erpInventoryBalanceRowsToResponses(rows), Total: total}, nil
}

func buildInventoryBalanceOrder(sorts string) string {
	return utils.BuildOrderBy(sorts, map[string]string{
		"warehouseCode":    "erp_warehouse.warehouse_code",
		"warehouseName":    "erp_warehouse.warehouse_name",
		"skuCode":          "product_sku.sku_code",
		"productName":      "product_spu.product_name",
		"specName":         "product_rp.spec_name",
		"enterpriseName":   "base_enterprise.enterprise_name",
		"batchNo":          "erp_inventory_batch.batch_no",
		"expiryDate":       "erp_inventory_batch.expiry_date",
		"unitCost":         "erp_inventory_batch.unit_cost",
		"packageUnitCount": "erp_inventory_balance.package_unit_count",
		"minUnitCount":     "erp_inventory_balance.min_unit_count",
		"inventoryAmount":  "erp_inventory_balance.package_unit_count * erp_inventory_batch.unit_cost",
		"movementCount":    "COALESCE(movement_stat.movement_count, 0)",
		"createDate":       "erp_inventory_balance.create_date",
		"updateDate":       "erp_inventory_balance.update_date",
	})
}

func (s *ErpInventoryService) GetInventoryMovements(balanceID string, req models.ErpInventoryMovementListRequest) (*utils.PaginationResponse, error) {
	if err := validateErpInventoryUUID(balanceID, "库存余额ID"); err != nil {
		return nil, err
	}
	if _, err := s.getExistingInventoryBalance(database.DB, balanceID); err != nil {
		return nil, err
	}

	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := s.baseInventoryMovementQuery().Where("erp_inventory_movement.balance_id = ?", balanceID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"sourceBillType": "erp_inventory_movement.source_bill_type",
		"sourceBillNo":   "erp_inventory_movement.source_bill_no",
		"movementType":   "erp_inventory_movement.movement_type",
		"direction":      "erp_inventory_movement.direction",
		"createDate":     "erp_inventory_movement.create_date",
	})
	if order == "" {
		order = "erp_inventory_movement.create_date desc"
	}

	var rows []erpInventoryMovementQueryRow
	if err := query.Select(erpInventoryMovementSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: erpInventoryMovementRowsToResponses(rows), Total: total}, nil
}

func (s *ErpInventoryService) GetInventoryMovementsBySource(req models.ErpInventorySourceMovementListRequest) (*utils.PaginationResponse, error) {
	if strings.TrimSpace(req.SourceBillType) == "" {
		return nil, fmt.Errorf("%w: 来源单据类型不能为空", ErrErpInventoryInvalidInput)
	}
	if err := validateErpInventoryUUID(req.SourceBillID, "来源单据ID"); err != nil {
		return nil, err
	}

	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := s.baseInventoryMovementQuery().Where(
		"erp_inventory_movement.source_bill_type = ? AND erp_inventory_movement.source_bill_id = ?",
		strings.TrimSpace(req.SourceBillType),
		strings.TrimSpace(req.SourceBillID),
	)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"sourceBillType": "erp_inventory_movement.source_bill_type",
		"sourceBillNo":   "erp_inventory_movement.source_bill_no",
		"movementType":   "erp_inventory_movement.movement_type",
		"direction":      "erp_inventory_movement.direction",
		"createDate":     "erp_inventory_movement.create_date",
	})
	if order == "" {
		order = "erp_inventory_movement.create_date desc"
	}

	var rows []erpInventoryMovementQueryRow
	if err := query.Select(erpInventoryMovementSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: erpInventoryMovementRowsToResponses(rows), Total: total}, nil
}

func (s *ErpInventoryService) GetInventoryTraceCodeList(req models.ErpInventoryTraceCodeListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := s.baseInventoryTraceCodeQuery()
	if traceCode := strings.TrimSpace(req.TraceCode); traceCode != "" {
		query = query.Where("erp_inventory_trace_code.trace_code = ?", traceCode)
	}
	if skuCode := strings.TrimSpace(req.SkuCode); skuCode != "" {
		query = query.Where("product_sku.sku_code LIKE ?", "%"+skuCode+"%")
	}
	if batchNo := strings.TrimSpace(req.BatchNo); batchNo != "" {
		query = query.Where("erp_inventory_batch.batch_no LIKE ?", "%"+batchNo+"%")
	}
	if warehouseID := strings.TrimSpace(req.WarehouseID); warehouseID != "" {
		if err := validateErpInventoryUUID(warehouseID, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_inventory_balance.warehouse_id = ?", warehouseID)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		if status != models.InventoryTraceCodeStatusInStock && status != models.InventoryTraceCodeStatusOutbound {
			return nil, fmt.Errorf("%w: 追溯码状态不正确", ErrErpInventoryInvalidInput)
		}
		query = query.Where("erp_inventory_trace_code.status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"traceCode":  "erp_inventory_trace_code.trace_code",
		"skuCode":    "product_sku.sku_code",
		"batchNo":    "erp_inventory_batch.batch_no",
		"status":     "erp_inventory_trace_code.status",
		"createDate": "erp_inventory_trace_code.create_date",
		"updateDate": "erp_inventory_trace_code.update_date",
	})
	if order == "" {
		order = "erp_inventory_trace_code.update_date desc"
	}
	var rows []erpInventoryTraceCodeQueryRow
	if err := query.Select(erpInventoryTraceCodeSelectFields()).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &utils.PaginationResponse{Items: erpInventoryTraceCodeRowsToResponses(rows), Total: total}, nil
}

func (s *ErpInventoryService) GetInventoryTraceCodeMovements(traceID string, req models.ErpInventoryMovementListRequest) (*utils.PaginationResponse, error) {
	if err := validateErpInventoryUUID(traceID, "追溯码ID"); err != nil {
		return nil, err
	}
	var count int64
	if err := database.DB.Model(&models.ErpInventoryTraceCode{}).Where("trace_id = ?", traceID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("%w: 追溯码不存在", ErrErpInventoryNotFound)
	}
	page, pageSize := normalizeErpInventoryPage(req.Page, req.PageSize, 20, 100)
	query := s.baseInventoryMovementQuery().
		Joins("INNER JOIN erp_inventory_movement_trace_code ON erp_inventory_movement_trace_code.movement_id = erp_inventory_movement.movement_id").
		Where("erp_inventory_movement_trace_code.trace_id = ?", traceID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"sourceBillType": "erp_inventory_movement.source_bill_type",
		"sourceBillNo":   "erp_inventory_movement.source_bill_no",
		"movementType":   "erp_inventory_movement.movement_type",
		"direction":      "erp_inventory_movement.direction",
		"createDate":     "erp_inventory_movement.create_date",
	})
	if order == "" {
		order = "erp_inventory_movement.create_date desc"
	}
	var rows []erpInventoryMovementQueryRow
	if err := query.Select(erpInventoryMovementSelectFields()).Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &utils.PaginationResponse{Items: erpInventoryMovementRowsToResponses(rows), Total: total}, nil
}

func (s *ErpInventoryService) CreateInitialStocks(req models.CreateErpInventoryInitialStockRequest, operatorID string) (*models.CreateErpInventoryInitialStockResponse, error) {
	if err := validateErpInventoryUUID(req.WarehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	items, err := normalizeErpInventoryInitialStockItems(req.Items)
	if err != nil {
		return nil, err
	}

	var result models.CreateErpInventoryInitialStockResponse
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		warehouse, err := s.lockEnabledWarehouse(tx, strings.TrimSpace(req.WarehouseID))
		if err != nil {
			return err
		}
		skuMap, err := s.lockEnabledSkuMap(tx, items)
		if err != nil {
			return err
		}
		sourceBillNo, err := s.nextInitialStockSourceBillNo(tx)
		if err != nil {
			return err
		}

		movementCount := 0
		for _, item := range items {
			sku := skuMap[item.SkuID]
			if err := s.createInventoryInMovement(tx, *warehouse, sku, item, erpInventoryInMovementContext{
				SourceBillType: models.InventorySourceBillTypeInitialStock,
				SourceBillNo:   sourceBillNo,
				MovementType:   models.InventoryMovementTypeInitialIn,
			}, operatorID); err != nil {
				return err
			}
			movementCount++
		}

		result = models.CreateErpInventoryInitialStockResponse{
			SourceBillNo:  sourceBillNo,
			MovementCount: movementCount,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &result, nil
}

type normalizedErpInventoryInboundItem struct {
	SkuID      string
	BatchNo    string
	ExpiryDate time.Time
	UnitCost   string
	Quantity   int
	TraceCodes []string
	Remark     *string
}

type erpInventoryBalanceQueryRow struct {
	BalanceID        string
	WarehouseID      string
	WarehouseCode    string
	WarehouseName    string
	SkuID            string
	SkuCode          string
	ProductName      string
	SpecName         string
	EnterpriseName   string
	ApprovalNo       string
	BrandName        *string
	PackageSpecName  string
	TraceMode        string
	BatchID          string
	BatchNo          string
	ExpiryDate       time.Time
	UnitCost         string
	PackageUnitCount int
	PackageUnitName  string
	MinUnitCount     int64
	MinUnitName      string
	MovementCount    int
	RowVersion       int
	CreateDate       *time.Time
	UpdateDate       *time.Time
}

type erpInventoryMovementQueryRow struct {
	MovementID             string
	BalanceID              string
	WarehouseID            string
	WarehouseCode          string
	WarehouseName          string
	SkuID                  string
	SkuCode                string
	ProductName            string
	SpecName               string
	EnterpriseName         string
	ApprovalNo             string
	BrandName              *string
	PackageSpecName        string
	BatchID                string
	BatchNo                string
	ExpiryDate             time.Time
	UnitCost               string
	SourceBillType         string
	SourceBillID           *string
	SourceBillNo           string
	MovementType           string
	Direction              string
	BeforePackageUnitCount int
	ChangePackageUnitCount int
	AfterPackageUnitCount  int
	BeforeMinUnitCount     int64
	ChangeMinUnitCount     int64
	AfterMinUnitCount      int64
	PackageUnitName        string
	MinUnitName            string
	Remark                 *string
	CreateDate             *time.Time
}

type erpInventoryTraceCodeQueryRow struct {
	TraceID          string
	TraceCode        string
	SkuID            string
	SkuCode          string
	ProductName      string
	SpecName         string
	PackageSpecName  string
	BatchID          string
	BatchNo          string
	ExpiryDate       time.Time
	CurrentBalanceID *string
	WarehouseID      *string
	WarehouseName    *string
	Status           string
	PackageLevel     string
	ParentTraceID    *string
	RowVersion       int
	CreateDate       *time.Time
	UpdateDate       *time.Time
}

func (s *ErpInventoryService) applyInventoryBalanceFilters(query *gorm.DB, req models.ErpInventoryBalanceListRequest) (*gorm.DB, error) {
	if strings.TrimSpace(req.WarehouseID) != "" {
		if err := validateErpInventoryUUID(req.WarehouseID, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_inventory_balance.warehouse_id = ?", strings.TrimSpace(req.WarehouseID))
	}
	if strings.TrimSpace(req.BalanceIDs) != "" {
		balanceIDs, err := normalizeErpInventoryBalanceIDs(req.BalanceIDs)
		if err != nil {
			return nil, err
		}
		query = query.Where("erp_inventory_balance.balance_id IN ?", balanceIDs)
	}
	if strings.TrimSpace(req.SkuCode) != "" {
		skuCode := "%" + strings.ToLower(strings.TrimSpace(req.SkuCode)) + "%"
		query = query.Where("LOWER(product_sku.sku_code) LIKE ?", skuCode)
	}
	if strings.TrimSpace(req.BatchNo) != "" {
		batchNo := "%" + strings.ToLower(strings.TrimSpace(req.BatchNo)) + "%"
		query = query.Where("LOWER(erp_inventory_batch.batch_no) LIKE ?", batchNo)
	}
	if req.OnlyPositive {
		query = query.Where("erp_inventory_balance.package_unit_count > 0")
	}
	return query, nil
}

func (s *ErpInventoryService) baseInventoryBalanceQuery() *gorm.DB {
	return s.baseInventoryBalanceDataQuery().
		Joins("LEFT JOIN (?) AS movement_stat ON movement_stat.balance_id = erp_inventory_balance.balance_id", erpInventoryMovementCountSubquery())
}

func (s *ErpInventoryService) baseInventoryBalanceExportQuery() *gorm.DB {
	return s.baseInventoryBalanceDataQuery()
}

func (s *ErpInventoryService) baseInventoryBalanceDataQuery() *gorm.DB {
	return database.DB.Table("erp_inventory_balance").
		Joins("INNER JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_inventory_balance.warehouse_id AND erp_warehouse.del_flag = 0").
		Joins("INNER JOIN product_sku ON product_sku.sku_id = erp_inventory_balance.sku_id AND product_sku.del_flag = 0").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("LEFT JOIN base_enterprise ON base_enterprise.enterprise_id = product_mp.enterprise_id AND base_enterprise.del_flag = 0").
		Joins("INNER JOIN erp_inventory_batch ON erp_inventory_batch.batch_id = erp_inventory_balance.batch_id")
}

func (s *ErpInventoryService) baseInventoryMovementQuery() *gorm.DB {
	return database.DB.Table("erp_inventory_movement").
		Joins("INNER JOIN erp_inventory_balance ON erp_inventory_balance.balance_id = erp_inventory_movement.balance_id").
		Joins("INNER JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_inventory_movement.warehouse_id AND erp_warehouse.del_flag = 0").
		Joins("INNER JOIN product_sku ON product_sku.sku_id = erp_inventory_movement.sku_id AND product_sku.del_flag = 0").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("LEFT JOIN base_enterprise ON base_enterprise.enterprise_id = product_mp.enterprise_id AND base_enterprise.del_flag = 0").
		Joins("INNER JOIN erp_inventory_batch ON erp_inventory_batch.batch_id = erp_inventory_movement.batch_id")
}

func (s *ErpInventoryService) baseInventoryTraceCodeQuery() *gorm.DB {
	return database.DB.Table("erp_inventory_trace_code").
		Joins("INNER JOIN product_sku ON product_sku.sku_id = erp_inventory_trace_code.sku_id AND product_sku.del_flag = 0").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("INNER JOIN erp_inventory_batch ON erp_inventory_batch.batch_id = erp_inventory_trace_code.batch_id").
		Joins("LEFT JOIN erp_inventory_balance ON erp_inventory_balance.balance_id = erp_inventory_trace_code.current_balance_id").
		Joins("LEFT JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_inventory_balance.warehouse_id AND erp_warehouse.del_flag = 0")
}

func erpInventoryBalanceSelectFields() string {
	return strings.Join([]string{
		"erp_inventory_balance.balance_id",
		"erp_inventory_balance.warehouse_id",
		"erp_warehouse.warehouse_code",
		"erp_warehouse.warehouse_name",
		"erp_inventory_balance.sku_id",
		"product_sku.sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"COALESCE(base_enterprise.enterprise_name, '') AS enterprise_name",
		"COALESCE(product_mp.approval_no, '') AS approval_no",
		"product_mp.brand_name",
		"product_sku.package_spec_name",
		"product_sku.trace_mode",
		"erp_inventory_balance.batch_id",
		"erp_inventory_batch.batch_no",
		"erp_inventory_batch.expiry_date",
		"erp_inventory_batch.unit_cost",
		"erp_inventory_balance.package_unit_count",
		"erp_inventory_balance.package_unit_name",
		"erp_inventory_balance.min_unit_count",
		"erp_inventory_balance.min_unit_name",
		"COALESCE(movement_stat.movement_count, 0) AS movement_count",
		"erp_inventory_balance.row_version",
		"erp_inventory_balance.create_date",
		"erp_inventory_balance.update_date",
	}, ", ")
}

func erpInventoryBalanceExportSelectFields() string {
	return strings.Join([]string{
		"erp_inventory_balance.balance_id",
		"erp_inventory_balance.warehouse_id",
		"erp_warehouse.warehouse_code",
		"erp_warehouse.warehouse_name",
		"erp_inventory_balance.sku_id",
		"product_sku.sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"COALESCE(base_enterprise.enterprise_name, '') AS enterprise_name",
		"COALESCE(product_mp.approval_no, '') AS approval_no",
		"product_sku.package_spec_name",
		"erp_inventory_batch.batch_no",
		"erp_inventory_batch.expiry_date",
		"erp_inventory_batch.unit_cost",
		"erp_inventory_balance.package_unit_count",
		"erp_inventory_balance.package_unit_name",
		"erp_inventory_balance.min_unit_count",
		"erp_inventory_balance.min_unit_name",
		"erp_inventory_balance.update_date",
	}, ", ")
}

func erpInventoryMovementSelectFields() string {
	return strings.Join([]string{
		"erp_inventory_movement.movement_id",
		"erp_inventory_movement.balance_id",
		"erp_inventory_movement.warehouse_id",
		"erp_warehouse.warehouse_code",
		"erp_warehouse.warehouse_name",
		"erp_inventory_movement.sku_id",
		"product_sku.sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"COALESCE(base_enterprise.enterprise_name, '') AS enterprise_name",
		"COALESCE(product_mp.approval_no, '') AS approval_no",
		"product_mp.brand_name",
		"product_sku.package_spec_name",
		"erp_inventory_movement.batch_id",
		"erp_inventory_batch.batch_no",
		"erp_inventory_batch.expiry_date",
		"erp_inventory_batch.unit_cost",
		"erp_inventory_movement.source_bill_type",
		"erp_inventory_movement.source_bill_id",
		"erp_inventory_movement.source_bill_no",
		"erp_inventory_movement.movement_type",
		"erp_inventory_movement.direction",
		"erp_inventory_movement.before_package_unit_count",
		"erp_inventory_movement.change_package_unit_count",
		"erp_inventory_movement.after_package_unit_count",
		"erp_inventory_movement.before_min_unit_count",
		"erp_inventory_movement.change_min_unit_count",
		"erp_inventory_movement.after_min_unit_count",
		"erp_inventory_movement.package_unit_name",
		"erp_inventory_movement.min_unit_name",
		"erp_inventory_movement.remark",
		"erp_inventory_movement.create_date",
	}, ", ")
}

func erpInventoryTraceCodeSelectFields() string {
	return strings.Join([]string{
		"erp_inventory_trace_code.trace_id",
		"erp_inventory_trace_code.trace_code",
		"erp_inventory_trace_code.sku_id",
		"product_sku.sku_code",
		"COALESCE(product_spu.product_name, '') AS product_name",
		"COALESCE(product_rp.spec_name, '') AS spec_name",
		"product_sku.package_spec_name",
		"erp_inventory_trace_code.batch_id",
		"erp_inventory_batch.batch_no",
		"erp_inventory_batch.expiry_date",
		"erp_inventory_trace_code.current_balance_id",
		"erp_inventory_balance.warehouse_id",
		"erp_warehouse.warehouse_name",
		"erp_inventory_trace_code.status",
		"erp_inventory_trace_code.package_level",
		"erp_inventory_trace_code.parent_trace_id",
		"erp_inventory_trace_code.row_version",
		"erp_inventory_trace_code.create_date",
		"erp_inventory_trace_code.update_date",
	}, ", ")
}

func erpInventoryMovementCountSubquery() *gorm.DB {
	return database.DB.Table("erp_inventory_movement").
		Select("balance_id, COUNT(*) AS movement_count").
		Group("balance_id")
}

func (s *ErpInventoryService) getExistingInventoryBalance(tx *gorm.DB, balanceID string) (*models.ErpInventoryBalance, error) {
	var balance models.ErpInventoryBalance
	if err := tx.Where("balance_id = ?", balanceID).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 库存余额不存在", ErrErpInventoryNotFound)
		}
		return nil, err
	}
	return &balance, nil
}

func (s *ErpInventoryService) lockEnabledWarehouse(tx *gorm.DB, warehouseID string) (*models.ErpWarehouse, error) {
	var warehouse models.ErpWarehouse
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND del_flag = 0 AND status = 1", warehouseID).
		First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 仓库不存在或未启用", ErrErpInventoryNotFound)
		}
		return nil, err
	}
	return &warehouse, nil
}

func (s *ErpInventoryService) lockEnabledSkuMap(tx *gorm.DB, items []normalizedErpInventoryInboundItem) (map[string]models.ProductSku, error) {
	skuIDSet := make(map[string]struct{}, len(items))
	for _, item := range items {
		skuIDSet[item.SkuID] = struct{}{}
	}
	skuIDs := make([]string, 0, len(skuIDSet))
	for skuID := range skuIDSet {
		skuIDs = append(skuIDs, skuID)
	}
	sort.Strings(skuIDs)

	var skus []models.ProductSku
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sku_id IN ? AND del_flag = 0 AND status = 1", skuIDs).
		Find(&skus).Error; err != nil {
		return nil, err
	}
	if len(skus) != len(skuIDs) {
		return nil, fmt.Errorf("%w: SKU不存在或未启用", ErrErpInventoryNotFound)
	}

	skuMap := make(map[string]models.ProductSku, len(skus))
	for _, sku := range skus {
		if sku.PackConversion <= 0 {
			return nil, fmt.Errorf("%w: SKU包装换算系数无效", ErrErpInventoryInvalidInput)
		}
		skuMap[sku.SkuID] = sku
	}
	return skuMap, nil
}

// lockReferencedSkuMap loads SKU rows already fixed by a confirmed business document.
// A later status change on the master data must not invalidate the existing document.
func (s *ErpInventoryService) lockReferencedSkuMap(tx *gorm.DB, items []normalizedErpInventoryInboundItem) (map[string]models.ProductSku, error) {
	skuIDSet := make(map[string]struct{}, len(items))
	for _, item := range items {
		skuIDSet[item.SkuID] = struct{}{}
	}
	skuIDs := make([]string, 0, len(skuIDSet))
	for skuID := range skuIDSet {
		skuIDs = append(skuIDs, skuID)
	}
	sort.Strings(skuIDs)

	var skus []models.ProductSku
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sku_id IN ?", skuIDs).Find(&skus).Error; err != nil {
		return nil, err
	}
	if len(skus) != len(skuIDs) {
		return nil, fmt.Errorf("%w: 采购单关联的SKU不存在", ErrErpInventoryNotFound)
	}
	skuMap := make(map[string]models.ProductSku, len(skus))
	for _, sku := range skus {
		if sku.PackConversion <= 0 {
			return nil, fmt.Errorf("%w: SKU包装换算系数无效", ErrErpInventoryInvalidInput)
		}
		skuMap[sku.SkuID] = sku
	}
	return skuMap, nil
}

type erpInventoryInMovementContext struct {
	SourceBillType string
	SourceBillID   *string
	SourceBillNo   string
	MovementType   string
}

func (s *ErpInventoryService) createInventoryInMovement(tx *gorm.DB, warehouse models.ErpWarehouse, sku models.ProductSku, item normalizedErpInventoryInboundItem, context erpInventoryInMovementContext, operatorID string) error {
	now := time.Now().In(erpInventoryLocation)
	if err := validateErpInventoryTraceMode(sku.TraceMode, item.Quantity, item.TraceCodes); err != nil {
		return err
	}
	batch, err := s.getOrCreateInventoryBatch(tx, sku, item, operatorID, now)
	if err != nil {
		return err
	}

	changePackageCount := item.Quantity
	changeMinCount := int64(item.Quantity) * int64(sku.PackConversion)
	if changeMinCount <= 0 {
		return fmt.Errorf("%w: 库存换算数量无效", ErrErpInventoryInvalidInput)
	}

	balance, err := s.lockInventoryBalance(tx, warehouse.WarehouseID, batch.BatchID)
	if err != nil {
		return err
	}

	beforePackageCount := 0
	beforeMinCount := int64(0)
	afterPackageCount := changePackageCount
	afterMinCount := changeMinCount
	balanceID := utils.GenerateUUID()
	if balance != nil {
		if sku.TraceMode == models.ProductSkuTraceModeRequired {
			if err := s.ensureTraceBalanceConsistency(tx, balance.BalanceID, balance.PackageUnitCount); err != nil {
				return err
			}
		}
		beforePackageCount = balance.PackageUnitCount
		beforeMinCount = balance.MinUnitCount
		afterPackageCount = beforePackageCount + changePackageCount
		afterMinCount = beforeMinCount + changeMinCount
		balanceID = balance.BalanceID
		if err := tx.Model(&models.ErpInventoryBalance{}).Where("balance_id = ?", balance.BalanceID).Updates(map[string]interface{}{
			"package_unit_count": afterPackageCount,
			"package_unit_name":  sku.PackageUnitName,
			"min_unit_count":     afterMinCount,
			"min_unit_name":      sku.MinUnitName,
			"row_version":        balance.RowVersion + 1,
			"updater_id":         optionalErpInventoryOperatorID(operatorID),
			"update_date":        now,
		}).Error; err != nil {
			return err
		}
	} else {
		row := models.ErpInventoryBalance{
			BalanceID:        balanceID,
			WarehouseID:      warehouse.WarehouseID,
			SkuID:            sku.SkuID,
			BatchID:          batch.BatchID,
			PackageUnitCount: afterPackageCount,
			PackageUnitName:  sku.PackageUnitName,
			MinUnitCount:     afterMinCount,
			MinUnitName:      sku.MinUnitName,
			RowVersion:       1,
			CreatorID:        optionalErpInventoryOperatorID(operatorID),
			UpdaterID:        optionalErpInventoryOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		}
		if err := tx.Create(&row).Error; err != nil {
			if isErpInventoryDuplicateKeyError(err) {
				return fmt.Errorf("%w: 库存余额被并发创建，请刷新后重试", ErrErpInventoryConflict)
			}
			return err
		}
	}

	movement := models.ErpInventoryMovement{
		MovementID:             utils.GenerateUUID(),
		BalanceID:              balanceID,
		WarehouseID:            warehouse.WarehouseID,
		SkuID:                  sku.SkuID,
		BatchID:                batch.BatchID,
		SourceBillType:         context.SourceBillType,
		SourceBillID:           context.SourceBillID,
		SourceBillNo:           context.SourceBillNo,
		MovementType:           context.MovementType,
		Direction:              models.InventoryMovementDirectionIn,
		BeforePackageUnitCount: beforePackageCount,
		ChangePackageUnitCount: changePackageCount,
		AfterPackageUnitCount:  afterPackageCount,
		BeforeMinUnitCount:     beforeMinCount,
		ChangeMinUnitCount:     changeMinCount,
		AfterMinUnitCount:      afterMinCount,
		PackageUnitName:        sku.PackageUnitName,
		MinUnitName:            sku.MinUnitName,
		Remark:                 item.Remark,
		OperatorID:             optionalErpInventoryOperatorID(operatorID),
		CreateDate:             &now,
	}
	if err := tx.Create(&movement).Error; err != nil {
		return err
	}
	if err := s.createInboundTraceCodes(tx, sku, batch.BatchID, balanceID, movement.MovementID, item.TraceCodes, operatorID, now); err != nil {
		return err
	}
	if sku.TraceMode == models.ProductSkuTraceModeRequired {
		return s.ensureTraceBalanceConsistency(tx, balanceID, afterPackageCount)
	}
	return nil
}

func (s *ErpInventoryService) createInventoryOutMovement(tx *gorm.DB, warehouse models.ErpWarehouse, balanceID string, quantity int, traceCodes []string, remark *string, context erpInventoryInMovementContext, operatorID string) error {
	if quantity <= 0 {
		return fmt.Errorf("%w: 出库数量必须为正整数", ErrErpInventoryInvalidInput)
	}

	var balance models.ErpInventoryBalance
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("balance_id = ? AND warehouse_id = ?", balanceID, warehouse.WarehouseID).
		First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 库存余额不存在或不属于当前仓库", ErrErpInventoryNotFound)
		}
		return err
	}

	var sku models.ProductSku
	if err := tx.Where("sku_id = ? AND del_flag = 0", balance.SkuID).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: SKU不存在", ErrErpInventoryNotFound)
		}
		return err
	}
	if sku.PackConversion <= 0 {
		return fmt.Errorf("%w: SKU包装换算系数无效", ErrErpInventoryInvalidInput)
	}
	if err := validateErpInventoryTraceMode(sku.TraceMode, quantity, traceCodes); err != nil {
		return err
	}

	var batch models.ErpInventoryBatch
	if err := tx.Where("batch_id = ?", balance.BatchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 库存批次不存在", ErrErpInventoryNotFound)
		}
		return err
	}

	packageUnitName := balance.PackageUnitName
	if packageUnitName == "" {
		packageUnitName = sku.PackageUnitName
	}
	if balance.PackageUnitCount < quantity {
		return &ErpInventoryInsufficientError{
			SkuCode:         sku.SkuCode,
			BatchNo:         batch.BatchNo,
			Requested:       quantity,
			Available:       balance.PackageUnitCount,
			PackageUnitName: packageUnitName,
		}
	}
	var traceRows []models.ErpInventoryTraceCode
	if sku.TraceMode == models.ProductSkuTraceModeRequired {
		if err := s.ensureTraceBalanceConsistency(tx, balance.BalanceID, balance.PackageUnitCount); err != nil {
			return err
		}
		lockedRows, err := s.lockOutboundTraceCodes(tx, balance, traceCodes)
		if err != nil {
			return err
		}
		traceRows = lockedRows
	}

	changeMinCount := int64(quantity) * int64(sku.PackConversion)
	if changeMinCount <= 0 || balance.MinUnitCount < changeMinCount {
		return fmt.Errorf("%w: SKU %s 批号 %s 的最小单位库存数据不一致", ErrErpInventoryConflict, sku.SkuCode, batch.BatchNo)
	}

	now := time.Now().In(erpInventoryLocation)
	beforePackageCount := balance.PackageUnitCount
	beforeMinCount := balance.MinUnitCount
	afterPackageCount := beforePackageCount - quantity
	afterMinCount := beforeMinCount - changeMinCount
	if err := tx.Model(&models.ErpInventoryBalance{}).
		Where("balance_id = ?", balance.BalanceID).
		Updates(map[string]interface{}{
			"package_unit_count": afterPackageCount,
			"min_unit_count":     afterMinCount,
			"row_version":        balance.RowVersion + 1,
			"updater_id":         optionalErpInventoryOperatorID(operatorID),
			"update_date":        now,
		}).Error; err != nil {
		return err
	}

	minUnitName := balance.MinUnitName
	if minUnitName == "" {
		minUnitName = sku.MinUnitName
	}
	movement := models.ErpInventoryMovement{
		MovementID:             utils.GenerateUUID(),
		BalanceID:              balance.BalanceID,
		WarehouseID:            warehouse.WarehouseID,
		SkuID:                  balance.SkuID,
		BatchID:                balance.BatchID,
		SourceBillType:         context.SourceBillType,
		SourceBillID:           context.SourceBillID,
		SourceBillNo:           context.SourceBillNo,
		MovementType:           context.MovementType,
		Direction:              models.InventoryMovementDirectionOut,
		BeforePackageUnitCount: beforePackageCount,
		ChangePackageUnitCount: -quantity,
		AfterPackageUnitCount:  afterPackageCount,
		BeforeMinUnitCount:     beforeMinCount,
		ChangeMinUnitCount:     -changeMinCount,
		AfterMinUnitCount:      afterMinCount,
		PackageUnitName:        packageUnitName,
		MinUnitName:            minUnitName,
		Remark:                 remark,
		OperatorID:             optionalErpInventoryOperatorID(operatorID),
		CreateDate:             &now,
	}
	if err := tx.Create(&movement).Error; err != nil {
		return err
	}
	if err := s.markOutboundTraceCodes(tx, movement.MovementID, traceRows, operatorID, now); err != nil {
		return err
	}
	if sku.TraceMode == models.ProductSkuTraceModeRequired {
		return s.ensureTraceBalanceConsistency(tx, balance.BalanceID, afterPackageCount)
	}
	return nil
}

func validateErpInventoryTraceMode(traceMode string, quantity int, traceCodes []string) error {
	switch traceMode {
	case models.ProductSkuTraceModeNone:
		if len(traceCodes) > 0 {
			return fmt.Errorf("%w: 非追溯SKU不能录入追溯码", ErrErpInventoryInvalidInput)
		}
	case models.ProductSkuTraceModeRequired:
		if len(traceCodes) != quantity {
			return fmt.Errorf("%w: 追溯码数量%d必须与包装单位数量%d一致", ErrErpInventoryInvalidInput, len(traceCodes), quantity)
		}
	default:
		return fmt.Errorf("%w: SKU追溯码管理模式不正确", ErrErpInventoryInvalidInput)
	}
	return nil
}

func normalizeErpInventoryTraceCodes(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		traceCode := strings.TrimSpace(value)
		if traceCode == "" {
			return nil, fmt.Errorf("%w: 第%d个追溯码不能为空", ErrErpInventoryInvalidInput, index+1)
		}
		if len(traceCode) > 64 {
			return nil, fmt.Errorf("%w: 第%d个追溯码不能超过64位", ErrErpInventoryInvalidInput, index+1)
		}
		for _, char := range traceCode {
			if char < '0' || char > '9' {
				return nil, fmt.Errorf("%w: 第%d个追溯码只能包含数字", ErrErpInventoryInvalidInput, index+1)
			}
		}
		if _, exists := seen[traceCode]; exists {
			return nil, fmt.Errorf("%w: 追溯码%s重复", ErrErpInventoryConflict, traceCode)
		}
		seen[traceCode] = struct{}{}
		result = append(result, traceCode)
	}
	return result, nil
}

func (s *ErpInventoryService) createInboundTraceCodes(tx *gorm.DB, sku models.ProductSku, batchID, balanceID, movementID string, traceCodes []string, operatorID string, now time.Time) error {
	if len(traceCodes) == 0 {
		return nil
	}
	var existingCount int64
	if err := tx.Model(&models.ErpInventoryTraceCode{}).Where("trace_code IN ?", traceCodes).Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount > 0 {
		return fmt.Errorf("%w: 追溯码已存在，不能重复入库", ErrErpInventoryConflict)
	}

	rows := make([]models.ErpInventoryTraceCode, 0, len(traceCodes))
	links := make([]models.ErpInventoryMovementTraceCode, 0, len(traceCodes))
	for _, traceCode := range traceCodes {
		traceID := utils.GenerateUUID()
		currentBalanceID := balanceID
		rows = append(rows, models.ErpInventoryTraceCode{
			TraceID:          traceID,
			TraceCode:        traceCode,
			SkuID:            sku.SkuID,
			BatchID:          batchID,
			CurrentBalanceID: &currentBalanceID,
			Status:           models.InventoryTraceCodeStatusInStock,
			PackageLevel:     models.InventoryTraceCodePackageLevelSmall,
			RowVersion:       1,
			CreatorID:        optionalErpInventoryOperatorID(operatorID),
			UpdaterID:        optionalErpInventoryOperatorID(operatorID),
			CreateDate:       &now,
			UpdateDate:       &now,
		})
		links = append(links, models.ErpInventoryMovementTraceCode{
			MovementTraceCodeID: utils.GenerateUUID(),
			MovementID:          movementID,
			TraceID:             traceID,
			CreateDate:          &now,
		})
	}
	if err := tx.Create(&rows).Error; err != nil {
		if isErpInventoryDuplicateKeyError(err) {
			return fmt.Errorf("%w: 追溯码已存在，不能重复入库", ErrErpInventoryConflict)
		}
		return err
	}
	return tx.Create(&links).Error
}

func (s *ErpInventoryService) lockOutboundTraceCodes(tx *gorm.DB, balance models.ErpInventoryBalance, traceCodes []string) ([]models.ErpInventoryTraceCode, error) {
	var rows []models.ErpInventoryTraceCode
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("trace_code IN ?", traceCodes).
		Order("trace_code asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(traceCodes) {
		return nil, fmt.Errorf("%w: 存在未入库的追溯码", ErrErpInventoryNotFound)
	}
	for _, row := range rows {
		if row.Status != models.InventoryTraceCodeStatusInStock || row.CurrentBalanceID == nil {
			return nil, fmt.Errorf("%w: 追溯码%s当前不在库", ErrErpInventoryConflict, row.TraceCode)
		}
		if *row.CurrentBalanceID != balance.BalanceID || row.SkuID != balance.SkuID || row.BatchID != balance.BatchID {
			return nil, fmt.Errorf("%w: 追溯码%s不属于当前库存余额", ErrErpInventoryConflict, row.TraceCode)
		}
	}
	return rows, nil
}

func (s *ErpInventoryService) markOutboundTraceCodes(tx *gorm.DB, movementID string, rows []models.ErpInventoryTraceCode, operatorID string, now time.Time) error {
	if len(rows) == 0 {
		return nil
	}
	traceIDs := make([]string, 0, len(rows))
	links := make([]models.ErpInventoryMovementTraceCode, 0, len(rows))
	for _, row := range rows {
		traceIDs = append(traceIDs, row.TraceID)
		links = append(links, models.ErpInventoryMovementTraceCode{
			MovementTraceCodeID: utils.GenerateUUID(),
			MovementID:          movementID,
			TraceID:             row.TraceID,
			CreateDate:          &now,
		})
	}
	if err := tx.Model(&models.ErpInventoryTraceCode{}).
		Where("trace_id IN ? AND status = ?", traceIDs, models.InventoryTraceCodeStatusInStock).
		Updates(map[string]interface{}{
			"current_balance_id": nil,
			"status":             models.InventoryTraceCodeStatusOutbound,
			"row_version":        gorm.Expr("row_version + 1"),
			"updater_id":         optionalErpInventoryOperatorID(operatorID),
			"update_date":        now,
		}).Error; err != nil {
		return err
	}
	return tx.Create(&links).Error
}

func (s *ErpInventoryService) ensureTraceBalanceConsistency(tx *gorm.DB, balanceID string, expectedCount int) error {
	var count int64
	if err := tx.Model(&models.ErpInventoryTraceCode{}).
		Where("current_balance_id = ? AND status = ?", balanceID, models.InventoryTraceCodeStatusInStock).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(expectedCount) {
		return fmt.Errorf("%w: 库存余额数量与在库追溯码数量不一致", ErrErpInventoryConflict)
	}
	return nil
}

func (s *ErpInventoryService) getOrCreateInventoryBatch(tx *gorm.DB, sku models.ProductSku, item normalizedErpInventoryInboundItem, operatorID string, now time.Time) (*models.ErpInventoryBatch, error) {
	var batch models.ErpInventoryBatch
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sku_id = ? AND batch_no = ? AND expiry_date = ? AND unit_cost = ?", sku.SkuID, item.BatchNo, item.ExpiryDate, item.UnitCost).
		First(&batch).Error
	if err == nil {
		return &batch, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	batch = models.ErpInventoryBatch{
		BatchID:    utils.GenerateUUID(),
		SkuID:      sku.SkuID,
		BatchNo:    item.BatchNo,
		ExpiryDate: item.ExpiryDate,
		UnitCost:   item.UnitCost,
		CreatorID:  optionalErpInventoryOperatorID(operatorID),
		UpdaterID:  optionalErpInventoryOperatorID(operatorID),
		CreateDate: &now,
		UpdateDate: &now,
	}
	if err := tx.Create(&batch).Error; err != nil {
		if isErpInventoryDuplicateKeyError(err) {
			return nil, fmt.Errorf("%w: 库存批次被并发创建，请刷新后重试", ErrErpInventoryConflict)
		}
		return nil, err
	}
	return &batch, nil
}

func (s *ErpInventoryService) lockInventoryBalance(tx *gorm.DB, warehouseID, batchID string) (*models.ErpInventoryBalance, error) {
	var balance models.ErpInventoryBalance
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND batch_id = ?", warehouseID, batchID).
		First(&balance).Error
	if err == nil {
		return &balance, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, err
}

func (s *ErpInventoryService) nextInitialStockSourceBillNo(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_INVENTORY_INITIAL_STOCK", "INIT", 8)
}

func normalizeErpInventoryInitialStockItems(items []models.ErpInventoryInitialStockItem) ([]normalizedErpInventoryInboundItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: 初始库存明细不能为空", ErrErpInventoryInvalidInput)
	}
	if len(items) > 100 {
		return nil, fmt.Errorf("%w: 初始库存明细不能超过100行", ErrErpInventoryInvalidInput)
	}

	normalized := make([]normalizedErpInventoryInboundItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	traceCodeSeen := make(map[string]struct{})
	for _, item := range items {
		skuID := strings.TrimSpace(item.SkuID)
		if err := validateErpInventoryUUID(skuID, "SKU ID"); err != nil {
			return nil, err
		}
		batchNo := strings.TrimSpace(item.BatchNo)
		if batchNo == "" {
			return nil, fmt.Errorf("%w: 批号不能为空", ErrErpInventoryInvalidInput)
		}
		if len([]rune(batchNo)) > 64 {
			return nil, fmt.Errorf("%w: 批号不能超过64个字符", ErrErpInventoryInvalidInput)
		}
		expiryDate, err := parseErpInventoryDate(item.ExpiryDate, "有效期")
		if err != nil {
			return nil, err
		}
		unitCost, err := normalizeErpInventoryAmount(item.UnitCost)
		if err != nil {
			return nil, err
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: 入库数量必须为正整数", ErrErpInventoryInvalidInput)
		}
		if item.Quantity > 999999999 {
			return nil, fmt.Errorf("%w: 入库数量不能超过999999999", ErrErpInventoryInvalidInput)
		}
		traceCodes, err := normalizeErpInventoryTraceCodes(item.TraceCodes)
		if err != nil {
			return nil, err
		}
		for _, traceCode := range traceCodes {
			if _, exists := traceCodeSeen[traceCode]; exists {
				return nil, fmt.Errorf("%w: 同次提交中追溯码%s重复", ErrErpInventoryConflict, traceCode)
			}
			traceCodeSeen[traceCode] = struct{}{}
		}
		remark := normalizeErpInventoryOptionalString(item.Remark)
		if remark != nil && len([]rune(*remark)) > 512 {
			return nil, fmt.Errorf("%w: 备注不能超过512个字符", ErrErpInventoryInvalidInput)
		}

		duplicateKey := strings.Join([]string{skuID, batchNo, expiryDate.Format("2006-01-02"), unitCost}, "|")
		if _, exists := seen[duplicateKey]; exists {
			return nil, fmt.Errorf("%w: 同次提交中相同SKU、批号、有效期和成本价不能重复", ErrErpInventoryConflict)
		}
		seen[duplicateKey] = struct{}{}

		normalized = append(normalized, normalizedErpInventoryInboundItem{
			SkuID:      skuID,
			BatchNo:    batchNo,
			ExpiryDate: expiryDate,
			UnitCost:   unitCost,
			Quantity:   item.Quantity,
			TraceCodes: traceCodes,
			Remark:     remark,
		})
	}
	return normalized, nil
}

func erpInventoryBalanceRowsToResponses(rows []erpInventoryBalanceQueryRow) []models.ErpInventoryBalanceResponse {
	responses := make([]models.ErpInventoryBalanceResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, models.ErpInventoryBalanceResponse{
			BalanceID:        row.BalanceID,
			WarehouseID:      row.WarehouseID,
			WarehouseCode:    row.WarehouseCode,
			WarehouseName:    row.WarehouseName,
			SkuID:            row.SkuID,
			SkuCode:          row.SkuCode,
			ProductName:      row.ProductName,
			SpecName:         row.SpecName,
			EnterpriseName:   row.EnterpriseName,
			ApprovalNo:       row.ApprovalNo,
			BrandName:        row.BrandName,
			PackageSpecName:  row.PackageSpecName,
			TraceMode:        row.TraceMode,
			BatchID:          row.BatchID,
			BatchNo:          row.BatchNo,
			ExpiryDate:       formatErpInventoryDate(row.ExpiryDate),
			UnitCost:         row.UnitCost,
			PackageUnitCount: row.PackageUnitCount,
			PackageUnitName:  row.PackageUnitName,
			MinUnitCount:     row.MinUnitCount,
			MinUnitName:      row.MinUnitName,
			InventoryAmount:  multiplyErpInventoryAmount(row.UnitCost, row.PackageUnitCount),
			MovementCount:    row.MovementCount,
			RowVersion:       row.RowVersion,
			CreateDate:       models.TimeToStringPtr(row.CreateDate),
			UpdateDate:       models.TimeToStringPtr(row.UpdateDate),
		})
	}
	return responses
}

func erpInventoryMovementRowsToResponses(rows []erpInventoryMovementQueryRow) []models.ErpInventoryMovementResponse {
	responses := make([]models.ErpInventoryMovementResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, models.ErpInventoryMovementResponse{
			MovementID:             row.MovementID,
			BalanceID:              row.BalanceID,
			WarehouseID:            row.WarehouseID,
			WarehouseCode:          row.WarehouseCode,
			WarehouseName:          row.WarehouseName,
			SkuID:                  row.SkuID,
			SkuCode:                row.SkuCode,
			ProductName:            row.ProductName,
			SpecName:               row.SpecName,
			EnterpriseName:         row.EnterpriseName,
			ApprovalNo:             row.ApprovalNo,
			BrandName:              row.BrandName,
			PackageSpecName:        row.PackageSpecName,
			BatchID:                row.BatchID,
			BatchNo:                row.BatchNo,
			ExpiryDate:             formatErpInventoryDate(row.ExpiryDate),
			UnitCost:               row.UnitCost,
			SourceBillType:         row.SourceBillType,
			SourceBillID:           row.SourceBillID,
			SourceBillNo:           row.SourceBillNo,
			MovementType:           row.MovementType,
			Direction:              row.Direction,
			BeforePackageUnitCount: row.BeforePackageUnitCount,
			ChangePackageUnitCount: row.ChangePackageUnitCount,
			AfterPackageUnitCount:  row.AfterPackageUnitCount,
			BeforeMinUnitCount:     row.BeforeMinUnitCount,
			ChangeMinUnitCount:     row.ChangeMinUnitCount,
			AfterMinUnitCount:      row.AfterMinUnitCount,
			PackageUnitName:        row.PackageUnitName,
			MinUnitName:            row.MinUnitName,
			Remark:                 row.Remark,
			CreateDate:             models.TimeToStringPtr(row.CreateDate),
		})
	}
	return responses
}

func erpInventoryTraceCodeRowsToResponses(rows []erpInventoryTraceCodeQueryRow) []models.ErpInventoryTraceCodeResponse {
	responses := make([]models.ErpInventoryTraceCodeResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, models.ErpInventoryTraceCodeResponse{
			TraceID:          row.TraceID,
			TraceCode:        row.TraceCode,
			SkuID:            row.SkuID,
			SkuCode:          row.SkuCode,
			ProductName:      row.ProductName,
			SpecName:         row.SpecName,
			PackageSpecName:  row.PackageSpecName,
			BatchID:          row.BatchID,
			BatchNo:          row.BatchNo,
			ExpiryDate:       formatErpInventoryDate(row.ExpiryDate),
			CurrentBalanceID: row.CurrentBalanceID,
			WarehouseID:      row.WarehouseID,
			WarehouseName:    row.WarehouseName,
			Status:           row.Status,
			PackageLevel:     row.PackageLevel,
			ParentTraceID:    row.ParentTraceID,
			RowVersion:       row.RowVersion,
			CreateDate:       models.TimeToStringPtr(row.CreateDate),
			UpdateDate:       models.TimeToStringPtr(row.UpdateDate),
		})
	}
	return responses
}

func normalizeErpInventoryAmount(value string) (string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > 14 {
		return "", fmt.Errorf("%w: 成本价格式不正确", ErrErpInventoryInvalidInput)
	}
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("%w: 成本价格式不正确", ErrErpInventoryInvalidInput)
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
			return "", fmt.Errorf("%w: 成本价最多保留四位小数", ErrErpInventoryInvalidInput)
		}
		for _, ch := range fraction {
			if ch < '0' || ch > '9' {
				return "", fmt.Errorf("%w: 成本价格式不正确", ErrErpInventoryInvalidInput)
			}
		}
	}
	for len(fraction) < 4 {
		fraction += "0"
	}
	if intPart == "0" && fraction == "0000" {
		return "", fmt.Errorf("%w: 成本价必须大于0", ErrErpInventoryInvalidInput)
	}
	return fmt.Sprintf("%s.%s", intPart, fraction), nil
}

func multiplyErpInventoryAmount(unitCost string, quantity int) string {
	rat, ok := new(big.Rat).SetString(unitCost)
	if !ok {
		return "0.0000"
	}
	rat.Mul(rat, big.NewRat(int64(quantity), 1))
	return rat.FloatString(4)
}

func parseErpInventoryDate(value, label string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%w: %s不能为空", ErrErpInventoryInvalidInput, label)
	}
	parsed, err := time.ParseInLocation("2006-01-02", trimmed, erpInventoryLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s格式错误，请使用 2006-01-02", ErrErpInventoryInvalidInput, label)
	}
	return parsed, nil
}

func formatErpInventoryDate(value time.Time) string {
	return value.In(erpInventoryLocation).Format("2006-01-02")
}

func validateErpInventoryUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrErpInventoryInvalidInput, label)
	}
	return nil
}

func normalizeErpInventoryBalanceIDs(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	ids := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		balanceID := strings.TrimSpace(part)
		if balanceID == "" {
			continue
		}
		if err := validateErpInventoryUUID(balanceID, "库存余额ID"); err != nil {
			return nil, err
		}
		if _, exists := seen[balanceID]; exists {
			continue
		}
		seen[balanceID] = struct{}{}
		ids = append(ids, balanceID)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: 库存余额ID不能为空", ErrErpInventoryInvalidInput)
	}
	if len(ids) > 100 {
		return nil, fmt.Errorf("%w: 库存余额ID不能超过100个", ErrErpInventoryInvalidInput)
	}
	return ids, nil
}

func isErpInventoryDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func normalizeErpInventoryOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeErpInventoryPage(page, pageSize, defaultSize, maxSize int) (int, int) {
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

func optionalErpInventoryOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

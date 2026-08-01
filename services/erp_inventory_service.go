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

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
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

func (s *ErpInventoryService) applyInventoryBalanceFilters(query *gorm.DB, req models.ErpInventoryBalanceListRequest) (*gorm.DB, error) {
	if strings.TrimSpace(req.WarehouseID) != "" {
		if err := validateErpInventoryUUID(req.WarehouseID, "仓库ID"); err != nil {
			return nil, err
		}
		query = query.Where("erp_inventory_balance.warehouse_id = ?", strings.TrimSpace(req.WarehouseID))
	}
	if strings.TrimSpace(req.SkuCode) != "" {
		skuCode := "%" + strings.ToLower(strings.TrimSpace(req.SkuCode)) + "%"
		query = query.Where("LOWER(product_sku.sku_code) LIKE ?", skuCode)
	}
	if strings.TrimSpace(req.BatchNo) != "" {
		batchNo := "%" + strings.ToLower(strings.TrimSpace(req.BatchNo)) + "%"
		query = query.Where("LOWER(erp_inventory_batch.batch_no) LIKE ?", batchNo)
	}
	return query, nil
}

func (s *ErpInventoryService) baseInventoryBalanceQuery() *gorm.DB {
	return database.DB.Table("erp_inventory_balance").
		Joins("INNER JOIN erp_warehouse ON erp_warehouse.warehouse_id = erp_inventory_balance.warehouse_id AND erp_warehouse.del_flag = 0").
		Joins("INNER JOIN product_sku ON product_sku.sku_id = erp_inventory_balance.sku_id AND product_sku.del_flag = 0").
		Joins("LEFT JOIN product_mp ON product_mp.mp_id = product_sku.mp_id AND product_mp.del_flag = 0").
		Joins("LEFT JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("LEFT JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("LEFT JOIN base_enterprise ON base_enterprise.enterprise_id = product_mp.enterprise_id AND base_enterprise.del_flag = 0").
		Joins("INNER JOIN erp_inventory_batch ON erp_inventory_batch.batch_id = erp_inventory_balance.batch_id").
		Joins("LEFT JOIN (?) AS movement_stat ON movement_stat.balance_id = erp_inventory_balance.balance_id", erpInventoryMovementCountSubquery())
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

type erpInventoryInMovementContext struct {
	SourceBillType string
	SourceBillID   *string
	SourceBillNo   string
	MovementType   string
}

func (s *ErpInventoryService) createInventoryInMovement(tx *gorm.DB, warehouse models.ErpWarehouse, sku models.ProductSku, item normalizedErpInventoryInboundItem, context erpInventoryInMovementContext, operatorID string) error {
	now := time.Now().In(erpInventoryLocation)
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
	return tx.Create(&movement).Error
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

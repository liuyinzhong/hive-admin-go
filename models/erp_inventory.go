package models

import "time"

const (
	InventorySourceBillTypeInitialStock    = "INITIAL_STOCK"
	InventorySourceBillTypePurchaseInbound = "PURCHASE_INBOUND"
	InventoryMovementTypeInitialIn         = "INITIAL_IN"
	InventoryMovementTypePurchaseIn        = "PURCHASE_IN"
	InventoryMovementDirectionIn           = "IN"
)

type ErpInventoryBatch struct {
	BatchID    string     `gorm:"column:batch_id;type:char(36);primaryKey" json:"batchId"`
	SkuID      string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	BatchNo    string     `gorm:"column:batch_no;type:varchar(64)" json:"batchNo"`
	ExpiryDate time.Time  `gorm:"column:expiry_date;type:date" json:"expiryDate"`
	UnitCost   string     `gorm:"column:unit_cost;type:decimal(18,4)" json:"unitCost"`
	CreatorID  *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID  *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (ErpInventoryBatch) TableName() string {
	return "erp_inventory_batch"
}

type ErpInventoryBalance struct {
	BalanceID        string     `gorm:"column:balance_id;type:char(36);primaryKey" json:"balanceId"`
	WarehouseID      string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	SkuID            string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	BatchID          string     `gorm:"column:batch_id;type:char(36)" json:"batchId"`
	PackageUnitCount int        `gorm:"column:package_unit_count;type:int" json:"packageUnitCount"`
	PackageUnitName  string     `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	MinUnitCount     int64      `gorm:"column:min_unit_count;type:bigint" json:"minUnitCount"`
	MinUnitName      string     `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	RowVersion       int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID        *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID        *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (ErpInventoryBalance) TableName() string {
	return "erp_inventory_balance"
}

type ErpInventoryMovement struct {
	MovementID             string     `gorm:"column:movement_id;type:char(36);primaryKey" json:"movementId"`
	BalanceID              string     `gorm:"column:balance_id;type:char(36)" json:"balanceId"`
	WarehouseID            string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	SkuID                  string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	BatchID                string     `gorm:"column:batch_id;type:char(36)" json:"batchId"`
	SourceBillType         string     `gorm:"column:source_bill_type;type:varchar(32)" json:"sourceBillType"`
	SourceBillID           *string    `gorm:"column:source_bill_id;type:char(36)" json:"sourceBillId"`
	SourceBillNo           string     `gorm:"column:source_bill_no;type:varchar(32)" json:"sourceBillNo"`
	MovementType           string     `gorm:"column:movement_type;type:varchar(32)" json:"movementType"`
	Direction              string     `gorm:"column:direction;type:varchar(8)" json:"direction"`
	BeforePackageUnitCount int        `gorm:"column:before_package_unit_count;type:int" json:"beforePackageUnitCount"`
	ChangePackageUnitCount int        `gorm:"column:change_package_unit_count;type:int" json:"changePackageUnitCount"`
	AfterPackageUnitCount  int        `gorm:"column:after_package_unit_count;type:int" json:"afterPackageUnitCount"`
	BeforeMinUnitCount     int64      `gorm:"column:before_min_unit_count;type:bigint" json:"beforeMinUnitCount"`
	ChangeMinUnitCount     int64      `gorm:"column:change_min_unit_count;type:bigint" json:"changeMinUnitCount"`
	AfterMinUnitCount      int64      `gorm:"column:after_min_unit_count;type:bigint" json:"afterMinUnitCount"`
	PackageUnitName        string     `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	MinUnitName            string     `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	Remark                 *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	OperatorID             *string    `gorm:"column:operator_id;type:char(36)" json:"operatorId"`
	CreateDate             *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (ErpInventoryMovement) TableName() string {
	return "erp_inventory_movement"
}

type ErpInventoryBalanceListRequest struct {
	Page        int    `form:"page" example:"1"`                                           // 页码
	PageSize    int    `form:"pageSize" example:"20"`                                      // 每页数量
	WarehouseID string `form:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 仓库ID
	SkuCode     string `form:"skuCode" example:"SKU000001"`                                // SKU编码
	BatchNo     string `form:"batchNo" example:"B20260731001"`                             // 批号
	Sorts       string `form:"sorts" example:"updateDate,desc"`                            // 排序
}

type ErpInventoryMovementListRequest struct {
	Page     int    `form:"page" example:"1"`                // 页码
	PageSize int    `form:"pageSize" example:"20"`           // 每页数量
	Sorts    string `form:"sorts" example:"createDate,desc"` // 排序
}

type ErpInventorySourceMovementListRequest struct {
	SourceBillType string `form:"sourceBillType" example:"PURCHASE_INBOUND"`
	SourceBillID   string `form:"sourceBillId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Page           int    `form:"page" example:"1"`
	PageSize       int    `form:"pageSize" example:"20"`
	Sorts          string `form:"sorts" example:"createDate,desc"`
}

type CreateErpInventoryInitialStockRequest struct {
	WarehouseID string                         `json:"warehouseId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 入库仓库ID
	Items       []ErpInventoryInitialStockItem `json:"items" binding:"required"`                                                      // 初始库存明细
}

type ErpInventoryInitialStockItem struct {
	SkuID      string  `json:"skuId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // SKU ID
	BatchNo    string  `json:"batchNo" binding:"required,max=64" example:"B20260731001"`                // 批号
	ExpiryDate string  `json:"expiryDate" binding:"required" example:"2028-12-31"`                      // 有效期，格式YYYY-MM-DD
	UnitCost   string  `json:"unitCost" binding:"required" example:"12.0000"`                           // 包装单位成本价
	Quantity   int     `json:"quantity" binding:"required,min=1" example:"20"`                          // 包装单位入库数量
	Remark     *string `json:"remark" binding:"omitempty,max=512" example:"初始库存导入"`                     // 备注
}

type CreateErpInventoryInitialStockResponse struct {
	SourceBillNo  string `json:"sourceBillNo" example:"INIT00000001"` // 初始库存来源批次号
	MovementCount int    `json:"movementCount" example:"2"`           // 写入流水数量
}

type ErpInventoryBalanceResponse struct {
	BalanceID        string  `json:"balanceId" example:"550e8400-e29b-41d4-a716-446655440000"`   // 库存余额ID
	WarehouseID      string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 仓库ID
	WarehouseCode    string  `json:"warehouseCode" example:"WH000001"`                           // 仓库编码
	WarehouseName    string  `json:"warehouseName" example:"中心库"`                                // 仓库名称
	SkuID            string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`       // SKU ID
	SkuCode          string  `json:"skuCode" example:"SKU000001"`                                // SKU编码
	ProductName      string  `json:"productName" example:"阿莫西林胶囊"`                               // 通用名称
	SpecName         string  `json:"specName" example:"0.25g"`                                   // 规格名称
	EnterpriseName   string  `json:"enterpriseName" example:"张三药业"`                              // 生产企业
	ApprovalNo       string  `json:"approvalNo" example:"国药准字H20260001"`                         // 批准文号
	BrandName        *string `json:"brandName" example:"品牌名"`                                    // 品牌名
	PackageSpecName  string  `json:"packageSpecName" example:"10粒/盒"`                            // 包装规格
	BatchID          string  `json:"batchId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 批次ID
	BatchNo          string  `json:"batchNo" example:"B20260731001"`                             // 批号
	ExpiryDate       string  `json:"expiryDate" example:"2028-12-31"`                            // 有效期
	UnitCost         string  `json:"unitCost" example:"12.0000"`                                 // 包装单位成本价
	PackageUnitCount int     `json:"packageUnitCount" example:"20"`                              // 包装单位库存数量
	PackageUnitName  string  `json:"packageUnitName" example:"盒"`                                // 包装单位名称
	MinUnitCount     int64   `json:"minUnitCount" example:"400"`                                 // 最小单位库存数量
	MinUnitName      string  `json:"minUnitName" example:"粒"`                                    // 最小单位名称
	InventoryAmount  string  `json:"inventoryAmount" example:"240.0000"`                         // 库存金额，按包装单位库存数量*成本价计算
	MovementCount    int     `json:"movementCount" example:"2"`                                  // 库存流水数量
	RowVersion       int     `json:"rowVersion" example:"1"`                                     // 版本号
	CreateDate       *string `json:"createDate" example:"2026-01-15 09:00:00"`                   // 创建时间
	UpdateDate       *string `json:"updateDate" example:"2026-01-15 09:00:00"`                   // 更新时间
}

type ErpInventoryMovementResponse struct {
	MovementID             string  `json:"movementId" example:"550e8400-e29b-41d4-a716-446655440000"`   // 库存流水ID
	BalanceID              string  `json:"balanceId" example:"550e8400-e29b-41d4-a716-446655440000"`    // 库存余额ID
	WarehouseID            string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`  // 仓库ID
	WarehouseCode          string  `json:"warehouseCode" example:"WH000001"`                            // 仓库编码
	WarehouseName          string  `json:"warehouseName" example:"中心库"`                                 // 仓库名称
	SkuID                  string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // SKU ID
	SkuCode                string  `json:"skuCode" example:"SKU000001"`                                 // SKU编码
	ProductName            string  `json:"productName" example:"阿莫西林胶囊"`                                // 通用名称
	SpecName               string  `json:"specName" example:"0.25g"`                                    // 规格名称
	EnterpriseName         string  `json:"enterpriseName" example:"张三药业"`                               // 生产企业
	ApprovalNo             string  `json:"approvalNo" example:"国药准字H20260001"`                          // 批准文号
	BrandName              *string `json:"brandName" example:"品牌名"`                                     // 品牌名
	PackageSpecName        string  `json:"packageSpecName" example:"10粒/盒"`                             // 包装规格
	BatchID                string  `json:"batchId" example:"550e8400-e29b-41d4-a716-446655440000"`      // 批次ID
	BatchNo                string  `json:"batchNo" example:"B20260731001"`                              // 批号
	ExpiryDate             string  `json:"expiryDate" example:"2028-12-31"`                             // 有效期
	UnitCost               string  `json:"unitCost" example:"12.0000"`                                  // 包装单位成本价
	SourceBillType         string  `json:"sourceBillType" example:"INITIAL_STOCK"`                      // 来源单据类型
	SourceBillID           *string `json:"sourceBillId" example:"550e8400-e29b-41d4-a716-446655440000"` // 来源单据ID
	SourceBillNo           string  `json:"sourceBillNo" example:"INIT00000001"`                         // 来源单据号
	MovementType           string  `json:"movementType" example:"INITIAL_IN"`                           // 库存业务类型
	Direction              string  `json:"direction" example:"IN"`                                      // 方向：IN入库
	BeforePackageUnitCount int     `json:"beforePackageUnitCount" example:"0"`                          // 变更前包装单位数量
	ChangePackageUnitCount int     `json:"changePackageUnitCount" example:"20"`                         // 变更包装单位数量
	AfterPackageUnitCount  int     `json:"afterPackageUnitCount" example:"20"`                          // 变更后包装单位数量
	BeforeMinUnitCount     int64   `json:"beforeMinUnitCount" example:"0"`                              // 变更前最小单位数量
	ChangeMinUnitCount     int64   `json:"changeMinUnitCount" example:"400"`                            // 变更最小单位数量
	AfterMinUnitCount      int64   `json:"afterMinUnitCount" example:"400"`                             // 变更后最小单位数量
	PackageUnitName        string  `json:"packageUnitName" example:"盒"`                                 // 包装单位名称
	MinUnitName            string  `json:"minUnitName" example:"粒"`                                     // 最小单位名称
	Remark                 *string `json:"remark" example:"初始库存导入"`                                     // 备注
	CreateDate             *string `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
}

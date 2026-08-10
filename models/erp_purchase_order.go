package models

import "time"

const (
	ErpPurchaseOrderStatusDraft          = "DRAFT"
	ErpPurchaseOrderStatusWaitingReceipt = "WAITING_RECEIPT"
	ErpPurchaseOrderStatusPartialReceipt = "PARTIAL_RECEIPT"
	ErpPurchaseOrderStatusCompleted      = "COMPLETED"
	ErpPurchaseOrderStatusCancelled      = "CANCELLED"
	ErpPurchaseOrderStatusClosed         = "CLOSED"
)

const (
	ErpPurchaseOrderLogActionCreate  = "CREATE"
	ErpPurchaseOrderLogActionUpdate  = "UPDATE"
	ErpPurchaseOrderLogActionConfirm = "CONFIRM"
	ErpPurchaseOrderLogActionInbound = "INBOUND"
	ErpPurchaseOrderLogActionCancel  = "CANCEL"
	ErpPurchaseOrderLogActionClose   = "CLOSE"
)

type ErpPurchaseOrder struct {
	PurchaseOrderID      string     `gorm:"column:purchase_order_id;type:char(36);primaryKey" json:"purchaseOrderId"`
	PurchaseOrderNo      string     `gorm:"column:purchase_order_no;type:varchar(32)" json:"purchaseOrderNo"`
	SupplierID           string     `gorm:"column:supplier_id;type:char(36)" json:"supplierId"`
	SupplierNameSnapshot string     `gorm:"column:supplier_name_snapshot;type:varchar(128)" json:"supplierNameSnapshot"`
	WarehouseID          string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	OrderDate            time.Time  `gorm:"column:order_date;type:date" json:"orderDate"`
	ExpectedArrivalDate  *time.Time `gorm:"column:expected_arrival_date;type:date" json:"expectedArrivalDate"`
	Status               string     `gorm:"column:status;type:varchar(32)" json:"status"`
	Remark               *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion           int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID            *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID            *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	ConfirmedBy          *string    `gorm:"column:confirmed_by;type:char(36)" json:"confirmedBy"`
	ConfirmedAt          *time.Time `gorm:"column:confirmed_at" json:"confirmedAt"`
	CreateDate           *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate           *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (ErpPurchaseOrder) TableName() string { return "erp_purchase_order" }

type ErpPurchaseOrderItem struct {
	PurchaseOrderItemID string     `gorm:"column:purchase_order_item_id;type:char(36);primaryKey" json:"purchaseOrderItemId"`
	PurchaseOrderID     string     `gorm:"column:purchase_order_id;type:char(36)" json:"purchaseOrderId"`
	LineNo              int        `gorm:"column:line_no;type:int" json:"lineNo"`
	SkuID               string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	OrderedQuantity     int        `gorm:"column:ordered_quantity;type:int" json:"orderedQuantity"`
	InboundQuantity     int        `gorm:"column:inbound_quantity;type:int;default:0" json:"inboundQuantity"`
	UnitPrice           string     `gorm:"column:unit_price;type:decimal(18,4)" json:"unitPrice"`
	SkuCodeSnapshot     string     `gorm:"column:sku_code_snapshot;type:varchar(16)" json:"skuCodeSnapshot"`
	ProductNameSnapshot string     `gorm:"column:product_name_snapshot;type:varchar(128)" json:"productNameSnapshot"`
	SpecNameSnapshot    string     `gorm:"column:spec_name_snapshot;type:varchar(128)" json:"specNameSnapshot"`
	PackageSpecSnapshot string     `gorm:"column:package_spec_snapshot;type:varchar(128)" json:"packageSpecSnapshot"`
	PackageUnitSnapshot string     `gorm:"column:package_unit_snapshot;type:varchar(32)" json:"packageUnitSnapshot"`
	Remark              *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreateDate          *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate          *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (ErpPurchaseOrderItem) TableName() string { return "erp_purchase_order_item" }

type ErpPurchaseOrderLog struct {
	PurchaseOrderLogID string    `gorm:"column:purchase_order_log_id;type:char(36);primaryKey" json:"purchaseOrderLogId"`
	PurchaseOrderID    string    `gorm:"column:purchase_order_id;type:char(36)" json:"purchaseOrderId"`
	ActionType         string    `gorm:"column:action_type;type:varchar(32)" json:"actionType"`
	FromStatus         *string   `gorm:"column:from_status;type:varchar(32)" json:"fromStatus"`
	ToStatus           *string   `gorm:"column:to_status;type:varchar(32)" json:"toStatus"`
	Summary            string    `gorm:"column:summary;type:varchar(512)" json:"summary"`
	RelatedInboundID   *string   `gorm:"column:related_inbound_id;type:char(36)" json:"relatedInboundId"`
	RelatedInboundNo   *string   `gorm:"column:related_inbound_no;type:varchar(32)" json:"relatedInboundNo"`
	Reason             *string   `gorm:"column:reason;type:varchar(500)" json:"reason"`
	OperatorID         *string   `gorm:"column:operator_id;type:char(36)" json:"operatorId"`
	OperatedAt         time.Time `gorm:"column:operated_at" json:"operatedAt"`
}

func (ErpPurchaseOrderLog) TableName() string { return "erp_purchase_order_log" }

type ErpPurchaseOrderListRequest struct {
	Page            int    `form:"page" example:"1"`
	PageSize        int    `form:"pageSize" example:"20"`
	PurchaseOrderNo string `form:"purchaseOrderNo" example:"PO000001"`
	SupplierID      string `form:"supplierId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseID     string `form:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode         string `form:"skuCode" example:"SKU000001"`
	Status          string `form:"status" example:"WAITING_RECEIPT"`
	OrderDateFrom   string `form:"orderDateFrom" example:"2026-08-01"`
	OrderDateTo     string `form:"orderDateTo" example:"2026-08-31"`
	Sorts           string `form:"sorts" example:"orderDate,desc;createDate,desc"`
}

type SaveErpPurchaseOrderRequest struct {
	SupplierID          string                     `json:"supplierId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseID         string                     `json:"warehouseId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrderDate           string                     `json:"orderDate" binding:"required" example:"2026-08-08"`
	ExpectedArrivalDate *string                    `json:"expectedArrivalDate" example:"2026-08-15"`
	Remark              *string                    `json:"remark" binding:"omitempty,max=512" example:"常规采购"`
	ExpectedRowVersion  int                        `json:"expectedRowVersion" example:"1"`
	Items               []SaveErpPurchaseOrderItem `json:"items" binding:"required"`
}

type SaveErpPurchaseOrderItem struct {
	SkuID           string  `json:"skuId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	OrderedQuantity int     `json:"orderedQuantity" binding:"required,min=1" example:"100"`
	UnitPrice       string  `json:"unitPrice" binding:"required" example:"12.0000"`
	Remark          *string `json:"remark" binding:"omitempty,max=512" example:"采购明细备注"`
}

type ErpPurchaseOrderReasonRequest struct {
	Reason string `json:"reason" binding:"required,max=500" example:"供应商无法继续供货"`
}

type ErpPurchaseOrderListResponse struct {
	PurchaseOrderID     string  `json:"purchaseOrderId" example:"550e8400-e29b-41d4-a716-446655440000"`
	PurchaseOrderNo     string  `json:"purchaseOrderNo" example:"PO000001"`
	SupplierID          string  `json:"supplierId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SupplierName        string  `json:"supplierName" example:"示例供应商"`
	WarehouseID         string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseName       string  `json:"warehouseName" example:"中心库"`
	OrderDate           string  `json:"orderDate" example:"2026-08-08"`
	ExpectedArrivalDate *string `json:"expectedArrivalDate" example:"2026-08-15"`
	Status              string  `json:"status" example:"WAITING_RECEIPT"`
	LineCount           int     `json:"lineCount" example:"2"`
	TotalAmount         string  `json:"totalAmount" example:"1200.0000"`
	Remark              *string `json:"remark" example:"常规采购"`
	RowVersion          int     `json:"rowVersion" example:"1"`
	CreatorID           *string `json:"creatorId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateDate          *string `json:"createDate" example:"2026-08-08 10:00:00"`
	UpdateDate          *string `json:"updateDate" example:"2026-08-08 10:30:00"`
}

type ErpPurchaseOrderItemResponse struct {
	PurchaseOrderItemID string  `json:"purchaseOrderItemId" example:"550e8400-e29b-41d4-a716-446655440001"`
	LineNo              int     `json:"lineNo" example:"1"`
	SkuID               string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440001"`
	SkuCode             string  `json:"skuCode" example:"SKU000001"`
	ProductName         string  `json:"productName" example:"阿莫西林胶囊"`
	SpecName            string  `json:"specName" example:"0.25g"`
	PackageSpecName     string  `json:"packageSpecName" example:"10粒/盒"`
	PackageUnitName     string  `json:"packageUnitName" example:"盒"`
	TraceMode           string  `json:"traceMode" example:"REQUIRED"`
	OrderedQuantity     int     `json:"orderedQuantity" example:"100"`
	InboundQuantity     int     `json:"inboundQuantity" example:"40"`
	RemainingQuantity   int     `json:"remainingQuantity" example:"60"`
	UnitPrice           string  `json:"unitPrice" example:"12.0000"`
	Amount              string  `json:"amount" example:"1200.0000"`
	Remark              *string `json:"remark" example:"采购明细备注"`
}

type ErpPurchaseOrderResponse struct {
	ErpPurchaseOrderListResponse
	ConfirmedBy *string                        `json:"confirmedBy" example:"550e8400-e29b-41d4-a716-446655440000"`
	ConfirmedAt *string                        `json:"confirmedAt" example:"2026-08-08 11:00:00"`
	Items       []ErpPurchaseOrderItemResponse `json:"items"`
}

type ErpPurchaseOrderLogResponse struct {
	PurchaseOrderLogID string  `json:"purchaseOrderLogId" example:"550e8400-e29b-41d4-a716-446655440002"`
	ActionType         string  `json:"actionType" example:"INBOUND"`
	FromStatus         *string `json:"fromStatus" example:"WAITING_RECEIPT"`
	ToStatus           *string `json:"toStatus" example:"PARTIAL_RECEIPT"`
	Summary            string  `json:"summary" example:"采购入库单PIN00000001，涉及2个SKU、3条入库明细"`
	RelatedInboundID   *string `json:"relatedInboundId" example:"550e8400-e29b-41d4-a716-446655440003"`
	RelatedInboundNo   *string `json:"relatedInboundNo" example:"PIN00000001"`
	Reason             *string `json:"reason" example:"供应商无法继续供货"`
	OperatorID         *string `json:"operatorId" example:"550e8400-e29b-41d4-a716-446655440004"`
	OperatorName       *string `json:"operatorName" example:"仓库管理员"`
	OperatedAt         string  `json:"operatedAt" example:"2026-08-08 12:00:00"`
}

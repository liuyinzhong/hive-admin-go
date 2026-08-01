package models

import "time"

type ErpPurchaseInbound struct {
	InboundID   string     `gorm:"column:inbound_id;type:char(36);primaryKey" json:"inboundId"`
	InboundNo   string     `gorm:"column:inbound_no;type:varchar(32)" json:"inboundNo"`
	SupplierID  string     `gorm:"column:supplier_id;type:char(36)" json:"supplierId"`
	WarehouseID string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	InboundDate time.Time  `gorm:"column:inbound_date;type:date" json:"inboundDate"`
	Remark      *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID   *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate  *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (ErpPurchaseInbound) TableName() string {
	return "erp_purchase_inbound"
}

type ErpPurchaseInboundItem struct {
	InboundItemID string     `gorm:"column:inbound_item_id;type:char(36);primaryKey" json:"inboundItemId"`
	InboundID     string     `gorm:"column:inbound_id;type:char(36)" json:"inboundId"`
	LineNo        int        `gorm:"column:line_no;type:int" json:"lineNo"`
	SkuID         string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	BatchNo       string     `gorm:"column:batch_no;type:varchar(64)" json:"batchNo"`
	ExpiryDate    time.Time  `gorm:"column:expiry_date;type:date" json:"expiryDate"`
	UnitCost      string     `gorm:"column:unit_cost;type:decimal(18,4)" json:"unitCost"`
	Quantity      int        `gorm:"column:quantity;type:int" json:"quantity"`
	Remark        *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreateDate    *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (ErpPurchaseInboundItem) TableName() string {
	return "erp_purchase_inbound_item"
}

type ErpPurchaseInboundListRequest struct {
	Page            int    `form:"page" example:"1"`
	PageSize        int    `form:"pageSize" example:"20"`
	InboundNo       string `form:"inboundNo" example:"PIN00000001"`
	SupplierID      string `form:"supplierId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseID     string `form:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode         string `form:"skuCode" example:"SKU000001"`
	BatchNo         string `form:"batchNo" example:"B20260731001"`
	InboundDateFrom string `form:"inboundDateFrom" example:"2026-07-01"`
	InboundDateTo   string `form:"inboundDateTo" example:"2026-07-31"`
	Sorts           string `form:"sorts" example:"inboundDate,desc;createDate,desc"`
}

type CreateErpPurchaseInboundRequest struct {
	SupplierID  string                         `json:"supplierId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseID string                         `json:"warehouseId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	InboundDate string                         `json:"inboundDate" binding:"required" example:"2026-07-31"`
	Remark      *string                        `json:"remark" binding:"omitempty,max=512" example:"采购到货入库"`
	Items       []CreateErpPurchaseInboundItem `json:"items" binding:"required"`
}

type CreateErpPurchaseInboundItem struct {
	SkuID      string  `json:"skuId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchNo    string  `json:"batchNo" binding:"required,max=64" example:"B20260731001"`
	ExpiryDate string  `json:"expiryDate" binding:"required" example:"2028-12-31"`
	UnitCost   string  `json:"unitCost" binding:"required" example:"12.0000"`
	Quantity   int     `json:"quantity" binding:"required,min=1" example:"20"`
	Remark     *string `json:"remark" binding:"omitempty,max=512" example:"首批到货"`
}

type ErpPurchaseInboundListResponse struct {
	InboundID     string  `json:"inboundId" example:"550e8400-e29b-41d4-a716-446655440000"`
	InboundNo     string  `json:"inboundNo" example:"PIN00000001"`
	SupplierID    string  `json:"supplierId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SupplierName  string  `json:"supplierName" example:"示例供应商"`
	WarehouseID   string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseName string  `json:"warehouseName" example:"中心库"`
	InboundDate   string  `json:"inboundDate" example:"2026-07-31"`
	LineCount     int     `json:"lineCount" example:"2"`
	TotalAmount   string  `json:"totalAmount" example:"240.0000"`
	Remark        *string `json:"remark" example:"采购到货入库"`
	CreatorID     *string `json:"creatorId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateDate    *string `json:"createDate" example:"2026-07-31 10:00:00"`
}

type ErpPurchaseInboundItemResponse struct {
	InboundItemID   string  `json:"inboundItemId" example:"550e8400-e29b-41d4-a716-446655440000"`
	LineNo          int     `json:"lineNo" example:"1"`
	SkuID           string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode         string  `json:"skuCode" example:"SKU000001"`
	ProductName     string  `json:"productName" example:"阿莫西林胶囊"`
	SpecName        string  `json:"specName" example:"0.25g"`
	EnterpriseName  string  `json:"enterpriseName" example:"示例药业"`
	PackageSpecName string  `json:"packageSpecName" example:"10粒/盒"`
	PackageUnitName string  `json:"packageUnitName" example:"盒"`
	MinUnitName     string  `json:"minUnitName" example:"粒"`
	BatchNo         string  `json:"batchNo" example:"B20260731001"`
	ExpiryDate      string  `json:"expiryDate" example:"2028-12-31"`
	UnitCost        string  `json:"unitCost" example:"12.0000"`
	Quantity        int     `json:"quantity" example:"20"`
	Amount          string  `json:"amount" example:"240.0000"`
	Remark          *string `json:"remark" example:"首批到货"`
}

type ErpPurchaseInboundResponse struct {
	InboundID     string                           `json:"inboundId" example:"550e8400-e29b-41d4-a716-446655440000"`
	InboundNo     string                           `json:"inboundNo" example:"PIN00000001"`
	SupplierID    string                           `json:"supplierId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SupplierName  string                           `json:"supplierName" example:"示例供应商"`
	WarehouseID   string                           `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseName string                           `json:"warehouseName" example:"中心库"`
	InboundDate   string                           `json:"inboundDate" example:"2026-07-31"`
	Remark        *string                          `json:"remark" example:"采购到货入库"`
	CreatorID     *string                          `json:"creatorId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateDate    *string                          `json:"createDate" example:"2026-07-31 10:00:00"`
	LineCount     int                              `json:"lineCount" example:"2"`
	TotalAmount   string                           `json:"totalAmount" example:"240.0000"`
	Items         []ErpPurchaseInboundItemResponse `json:"items"`
}

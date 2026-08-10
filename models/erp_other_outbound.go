package models

import "time"

type ErpOtherOutbound struct {
	OutboundID   string     `gorm:"column:outbound_id;type:char(36);primaryKey" json:"outboundId"`
	OutboundNo   string     `gorm:"column:outbound_no;type:varchar(32)" json:"outboundNo"`
	WarehouseID  string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	OutboundDate time.Time  `gorm:"column:outbound_date;type:date" json:"outboundDate"`
	Remark       *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID    *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (ErpOtherOutbound) TableName() string {
	return "erp_other_outbound"
}

type ErpOtherOutboundItem struct {
	OutboundItemID string     `gorm:"column:outbound_item_id;type:char(36);primaryKey" json:"outboundItemId"`
	OutboundID     string     `gorm:"column:outbound_id;type:char(36)" json:"outboundId"`
	LineNo         int        `gorm:"column:line_no;type:int" json:"lineNo"`
	BalanceID      string     `gorm:"column:balance_id;type:char(36)" json:"balanceId"`
	Quantity       int        `gorm:"column:quantity;type:int" json:"quantity"`
	Remark         *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (ErpOtherOutboundItem) TableName() string {
	return "erp_other_outbound_item"
}

type ErpOtherOutboundListRequest struct {
	Page             int    `form:"page" example:"1"`
	PageSize         int    `form:"pageSize" example:"20"`
	OutboundNo       string `form:"outboundNo" example:"OOUT00000001"`
	WarehouseID      string `form:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode          string `form:"skuCode" example:"SKU000001"`
	BatchNo          string `form:"batchNo" example:"B20260731001"`
	OutboundDateFrom string `form:"outboundDateFrom" example:"2026-07-01"`
	OutboundDateTo   string `form:"outboundDateTo" example:"2026-07-31"`
	Sorts            string `form:"sorts" example:"outboundDate,desc;createDate,desc"`
}

type CreateErpOtherOutboundRequest struct {
	WarehouseID  string                       `json:"warehouseId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OutboundDate string                       `json:"outboundDate" binding:"required" example:"2026-07-31"`
	Remark       *string                      `json:"remark" binding:"omitempty,max=500" example:"领用出库"`
	Items        []CreateErpOtherOutboundItem `json:"items" binding:"required"`
}

type CreateErpOtherOutboundItem struct {
	BalanceID  string   `json:"balanceId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity   int      `json:"quantity" binding:"required,min=1" example:"2"`
	TraceCodes []string `json:"traceCodes" example:"[\"81000000000000000001\"]"`
	Remark     *string  `json:"remark" binding:"omitempty,max=500" example:"日常领用"`
}

type ErpOtherOutboundListResponse struct {
	OutboundID    string  `json:"outboundId" example:"550e8400-e29b-41d4-a716-446655440000"`
	OutboundNo    string  `json:"outboundNo" example:"OOUT00000001"`
	WarehouseID   string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseName string  `json:"warehouseName" example:"中心库"`
	OutboundDate  string  `json:"outboundDate" example:"2026-07-31"`
	LineCount     int     `json:"lineCount" example:"2"`
	Remark        *string `json:"remark" example:"领用出库"`
	CreatorID     *string `json:"creatorId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateDate    *string `json:"createDate" example:"2026-07-31 10:00:00"`
}

type ErpOtherOutboundItemResponse struct {
	OutboundItemID  string  `json:"outboundItemId" example:"550e8400-e29b-41d4-a716-446655440000"`
	LineNo          int     `json:"lineNo" example:"1"`
	BalanceID       string  `json:"balanceId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuID           string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode         string  `json:"skuCode" example:"SKU000001"`
	ProductName     string  `json:"productName" example:"阿莫西林胶囊"`
	SpecName        string  `json:"specName" example:"0.25g"`
	EnterpriseName  string  `json:"enterpriseName" example:"示例药业"`
	PackageSpecName string  `json:"packageSpecName" example:"10粒/盒"`
	PackageUnitName string  `json:"packageUnitName" example:"盒"`
	BatchNo         string  `json:"batchNo" example:"B20260731001"`
	ExpiryDate      string  `json:"expiryDate" example:"2028-12-31"`
	UnitCost        string  `json:"unitCost" example:"12.0000"`
	Quantity        int     `json:"quantity" example:"2"`
	Remark          *string `json:"remark" example:"日常领用"`
}

type ErpOtherOutboundResponse struct {
	OutboundID    string                         `json:"outboundId" example:"550e8400-e29b-41d4-a716-446655440000"`
	OutboundNo    string                         `json:"outboundNo" example:"OOUT00000001"`
	WarehouseID   string                         `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	WarehouseName string                         `json:"warehouseName" example:"中心库"`
	OutboundDate  string                         `json:"outboundDate" example:"2026-07-31"`
	LineCount     int                            `json:"lineCount" example:"2"`
	Remark        *string                        `json:"remark" example:"领用出库"`
	CreatorID     *string                        `json:"creatorId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateDate    *string                        `json:"createDate" example:"2026-07-31 10:00:00"`
	Items         []ErpOtherOutboundItemResponse `json:"items"`
}

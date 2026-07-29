package models

import "time"

type ProductSkuPrice struct {
	PriceID        string     `gorm:"column:price_id;type:char(36);primaryKey" json:"priceId"`
	SkuID          string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	PriceType      string     `gorm:"column:price_type;type:varchar(32)" json:"priceType"`
	ScopeType      string     `gorm:"column:scope_type;type:varchar(32)" json:"scopeType"`
	ScopeID        *string    `gorm:"column:scope_id;type:char(36)" json:"scopeId"`
	CurrencyCode   string     `gorm:"column:currency_code;type:char(3);default:CNY" json:"currencyCode"`
	Price          string     `gorm:"column:price;type:decimal(18,4)" json:"price"`
	TaxIncluded    int        `gorm:"column:tax_included;type:tinyint;default:1" json:"taxIncluded"`
	EffectiveStart time.Time  `gorm:"column:effective_start" json:"effectiveStart"`
	EffectiveEnd   *time.Time `gorm:"column:effective_end" json:"effectiveEnd"`
	Status         int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark         *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion     int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID      *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductSkuPrice) TableName() string {
	return "product_sku_price"
}

type SaveProductSkuPriceRequest struct {
	PriceType          string  `json:"priceType" binding:"required,max=32" example:"RETAIL"`                       // 价格类型
	ScopeType          string  `json:"scopeType" binding:"required,max=32" example:"GLOBAL"`                       // 价格范围
	ScopeID            *string `json:"scopeId" binding:"omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // 范围对象ID
	CurrencyCode       string  `json:"currencyCode" binding:"omitempty,max=3" example:"CNY"`                       // 币种，第一版仅支持CNY
	Price              string  `json:"price" binding:"required" example:"29.8000"`                                 // 价格金额，最多4位小数
	TaxIncluded        *int    `json:"taxIncluded" binding:"required,oneof=0 1" example:"1"`                       // 是否含税
	EffectiveStart     string  `json:"effectiveStart" binding:"required" example:"2026-08-01 00:00:00"`            // 生效开始时间
	EffectiveEnd       *string `json:"effectiveEnd" binding:"omitempty" example:"2026-09-01 00:00:00"`             // 生效结束时间
	Status             *int    `json:"status" binding:"required,oneof=0 1" example:"1"`                            // 状态
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"价格维护说明"`                        // 备注
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                                             // 期望行版本号
}

type UpdateProductSkuPriceStatusRequest struct {
	Status             *int `json:"status" binding:"required,oneof=0 1" example:"1"`         // 状态
	ExpectedRowVersion int  `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type DeleteProductSkuPriceRequest struct {
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ProductSkuPriceResponse struct {
	PriceID        string  `json:"priceId" example:"550e8400-e29b-41d4-a716-446655440000"` // 价格ID
	SkuID          string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`   // SKU ID
	SkuCode        string  `json:"skuCode" example:"SKU000001"`                            // SKU编码
	PriceType      string  `json:"priceType" example:"RETAIL"`                             // 价格类型
	ScopeType      string  `json:"scopeType" example:"GLOBAL"`                             // 价格范围
	ScopeID        *string `json:"scopeId" example:"550e8400-e29b-41d4-a716-446655440000"` // 范围对象ID
	ScopeName      *string `json:"scopeName" example:"张三企业"`                               // 范围对象名称
	CurrencyCode   string  `json:"currencyCode" example:"CNY"`                             // 币种
	Price          string  `json:"price" example:"29.8000"`                                // 价格金额
	TaxIncluded    int     `json:"taxIncluded" example:"1"`                                // 是否含税
	EffectiveStart string  `json:"effectiveStart" example:"2026-08-01 00:00:00"`           // 生效开始时间
	EffectiveEnd   *string `json:"effectiveEnd" example:"2026-09-01 00:00:00"`             // 生效结束时间
	Status         int     `json:"status" example:"1"`                                     // 状态
	Remark         *string `json:"remark" example:"价格维护说明"`                                // 备注
	RowVersion     int     `json:"rowVersion" example:"1"`                                 // 版本号
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"`               // 创建时间
	UpdateDate     *string `json:"updateDate" example:"2026-01-15 09:00:00"`               // 更新时间
}

package models

import "time"

type ProductSku struct {
	SkuID                     string     `gorm:"column:sku_id;type:char(36);primaryKey" json:"skuId"`
	SkuCode                   string     `gorm:"column:sku_code;type:varchar(16)" json:"skuCode"`
	MpID                      string     `gorm:"column:mp_id;type:char(36)" json:"mpId"`
	PackageSpecName           string     `gorm:"column:package_spec_name;type:varchar(128)" json:"packageSpecName"`
	PackageSpecNameNormalized string     `gorm:"column:package_spec_name_normalized;type:varchar(128)" json:"packageSpecNameNormalized"`
	PackageQuantity           int        `gorm:"column:package_quantity;type:int" json:"packageQuantity"`
	MinUnitName               string     `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	PackageUnitName           string     `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	Barcode                   *string    `gorm:"column:barcode;type:varchar(64)" json:"barcode"`
	Gtin                      *string    `gorm:"column:gtin;type:varchar(64)" json:"gtin"`
	UdiDi                     *string    `gorm:"column:udi_di;type:varchar(128)" json:"udiDi"`
	AllowSplit                int        `gorm:"column:allow_split;type:tinyint;default:0" json:"allowSplit"`
	Description               *string    `gorm:"column:description;type:varchar(2000)" json:"description"`
	Status                    int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	RowVersion                int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID                 *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID                 *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate                *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate                *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                   int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductSku) TableName() string {
	return "product_sku"
}

type ProductSkuListRequest struct {
	MpID     string `form:"mpId" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Sorts    string `form:"sorts"`
}

type SaveProductSkuRequest struct {
	MpID               string  `json:"mpId" binding:"required"`
	PackageSpecName    string  `json:"packageSpecName" binding:"required,max=128"`
	PackageQuantity    int     `json:"packageQuantity" binding:"required,min=1,max=999999"`
	MinUnitName        string  `json:"minUnitName" binding:"required,max=32"`
	PackageUnitName    string  `json:"packageUnitName" binding:"required,max=32"`
	Barcode            *string `json:"barcode" binding:"omitempty,max=64"`
	Gtin               *string `json:"gtin" binding:"omitempty,max=64"`
	UdiDi              *string `json:"udiDi" binding:"omitempty,max=128"`
	AllowSplit         int     `json:"allowSplit" binding:"oneof=0 1"`
	Description        *string `json:"description" binding:"omitempty,max=2000"`
	Status             int     `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int     `json:"expectedRowVersion"`
}

type UpdateProductSkuStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1"`
}

type ProductSkuOptionsRequest struct {
	MpID     string `form:"mpId"`
	Keyword  string `form:"keyword"`
	PageSize int    `form:"pageSize"`
}

type ProductSkuResponse struct {
	SpuID           string  `json:"spuId"`
	SpuCode         string  `json:"spuCode"`
	ProductName     string  `json:"productName"`
	ProductType     string  `json:"productType"`
	RpID            string  `json:"rpId"`
	RpCode          string  `json:"rpCode"`
	SpecName        string  `json:"specName"`
	MpID            string  `json:"mpId"`
	MpCode          string  `json:"mpCode"`
	EnterpriseID    string  `json:"enterpriseId"`
	EnterpriseCode  string  `json:"enterpriseCode"`
	EnterpriseName  string  `json:"enterpriseName"`
	ApprovalNo      string  `json:"approvalNo"`
	BrandName       *string `json:"brandName"`
	SkuID           string  `json:"skuId"`
	SkuCode         string  `json:"skuCode"`
	PackageSpecName string  `json:"packageSpecName"`
	PackageQuantity int     `json:"packageQuantity"`
	MinUnitName     string  `json:"minUnitName"`
	PackageUnitName string  `json:"packageUnitName"`
	Barcode         *string `json:"barcode"`
	Gtin            *string `json:"gtin"`
	UdiDi           *string `json:"udiDi"`
	AllowSplit      int     `json:"allowSplit"`
	Description     *string `json:"description"`
	Status          int     `json:"status"`
	RowVersion      int     `json:"rowVersion"`
	CreateDate      *string `json:"createDate"`
	UpdateDate      *string `json:"updateDate"`
}

type ProductSkuOptionResponse struct {
	SpuID           string  `json:"spuId"`
	SpuCode         string  `json:"spuCode"`
	ProductName     string  `json:"productName"`
	ProductType     string  `json:"productType"`
	RpID            string  `json:"rpId"`
	RpCode          string  `json:"rpCode"`
	SpecName        string  `json:"specName"`
	MpID            string  `json:"mpId"`
	MpCode          string  `json:"mpCode"`
	EnterpriseID    string  `json:"enterpriseId"`
	EnterpriseName  string  `json:"enterpriseName"`
	ApprovalNo      string  `json:"approvalNo"`
	BrandName       *string `json:"brandName"`
	SkuID           string  `json:"skuId"`
	SkuCode         string  `json:"skuCode"`
	PackageSpecName string  `json:"packageSpecName"`
}

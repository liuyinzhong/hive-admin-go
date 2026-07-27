package models

import "time"

const (
	ProductTypeDrug       = "DRUG"
	ProductTypeDevice     = "DEVICE"
	ProductTypeConsumable = "CONSUMABLE"
	ProductTypeFSMP       = "FSMP"
	ProductTypeOther      = "OTHER"
)

type ProductSpu struct {
	SpuID                 string     `gorm:"column:spu_id;type:char(36);primaryKey" json:"spuId"`
	SpuCode               string     `gorm:"column:spu_code;type:varchar(16)" json:"spuCode"`
	ProductName           string     `gorm:"column:product_name;type:varchar(128)" json:"productName"`
	ProductNameNormalized string     `gorm:"column:product_name_normalized;type:varchar(128)" json:"productNameNormalized"`
	ShortName             *string    `gorm:"column:short_name;type:varchar(64)" json:"shortName"`
	ShortNameNormalized   *string    `gorm:"column:short_name_normalized;type:varchar(64)" json:"shortNameNormalized"`
	ProductType           string     `gorm:"column:product_type;type:varchar(32)" json:"productType"`
	Description           *string    `gorm:"column:description;type:varchar(2000)" json:"description"`
	Status                int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	RowVersion            int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID             *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID             *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate            *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate            *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag               int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductSpu) TableName() string {
	return "product_spu"
}

type ProductSpuListRequest struct {
	Page        int    `form:"page" example:"1"`
	PageSize    int    `form:"pageSize" example:"20"`
	Keyword     string `form:"keyword"`
	ProductType string `form:"productType"`
	Status      *int   `form:"status"`
	Sorts       string `form:"sorts"`
}

type SaveProductSpuRequest struct {
	ProductName        string  `json:"productName" binding:"required,max=128"`
	ShortName          *string `json:"shortName" binding:"omitempty,max=64"`
	ProductType        string  `json:"productType" binding:"required,max=32"`
	Description        *string `json:"description" binding:"omitempty,max=2000"`
	Status             int     `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int     `json:"expectedRowVersion"`
}

type UpdateProductSpuStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1"`
}

type ProductSpuOptionsRequest struct {
	Keyword     string `form:"keyword"`
	ProductType string `form:"productType"`
	PageSize    int    `form:"pageSize"`
}

type ProductSpuResponse struct {
	SpuID       string  `json:"spuId"`
	SpuCode     string  `json:"spuCode"`
	ProductName string  `json:"productName"`
	ShortName   *string `json:"shortName"`
	ProductType string  `json:"productType"`
	Description *string `json:"description"`
	Status      int     `json:"status"`
	RowVersion  int     `json:"rowVersion"`
	CreateDate  *string `json:"createDate"`
	UpdateDate  *string `json:"updateDate"`
}

type ProductSpuOptionResponse struct {
	SpuID       string  `json:"spuId"`
	SpuCode     string  `json:"spuCode"`
	ProductName string  `json:"productName"`
	ShortName   *string `json:"shortName"`
	ProductType string  `json:"productType"`
}

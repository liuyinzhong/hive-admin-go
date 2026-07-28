package models

import "time"

type ProductRp struct {
	RpID               string     `gorm:"column:rp_id;type:char(36);primaryKey" json:"rpId"`
	RpCode             string     `gorm:"column:rp_code;type:varchar(16)" json:"rpCode"`
	SpuID              string     `gorm:"column:spu_id;type:char(36)" json:"spuId"`
	SpecName           string     `gorm:"column:spec_name;type:varchar(128)" json:"specName"`
	SpecNameNormalized string     `gorm:"column:spec_name_normalized;type:varchar(128)" json:"specNameNormalized"`
	DosageForm         *string    `gorm:"column:dosage_form;type:varchar(64)" json:"dosageForm"`
	StrengthText       *string    `gorm:"column:strength_text;type:varchar(128)" json:"strengthText"`
	Description        *string    `gorm:"column:description;type:varchar(2000)" json:"description"`
	Status             int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	RowVersion         int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID          *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID          *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate         *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate         *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag            int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductRp) TableName() string {
	return "product_rp"
}

type ProductRpListRequest struct {
	Page        int    `form:"page" example:"1"`
	PageSize    int    `form:"pageSize" example:"20"`
	Keyword     string `form:"keyword"`
	SpuID       string `form:"spuId"`
	ProductType string `form:"productType"`
	Status      *int   `form:"status"`
	Sorts       string `form:"sorts"`
}

type SaveProductRpRequest struct {
	SpuID              string  `json:"spuId" binding:"required"`
	SpecName           string  `json:"specName" binding:"required,max=128"`
	DosageForm         *string `json:"dosageForm" binding:"omitempty,max=64"`
	StrengthText       *string `json:"strengthText" binding:"omitempty,max=128"`
	Description        *string `json:"description" binding:"omitempty,max=2000"`
	Status             int     `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int     `json:"expectedRowVersion"`
}

type UpdateProductRpStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1"`
}

type ProductRpOptionsRequest struct {
	Keyword     string `form:"keyword"`
	SpuID       string `form:"spuId"`
	ProductType string `form:"productType"`
	PageSize    int    `form:"pageSize"`
}

type ProductRpResponse struct {
	RpID         string  `json:"rpId"`
	RpCode       string  `json:"rpCode"`
	SpuID        string  `json:"spuId"`
	SpuCode      string  `json:"spuCode"`
	ProductName  string  `json:"productName"`
	ProductType  string  `json:"productType"`
	SpecName     string  `json:"specName"`
	DosageForm   *string `json:"dosageForm"`
	StrengthText *string `json:"strengthText"`
	Description  *string `json:"description"`
	Status       int     `json:"status"`
	RowVersion   int     `json:"rowVersion"`
	CreateDate   *string `json:"createDate"`
	UpdateDate   *string `json:"updateDate"`
}

type ProductRpOptionResponse struct {
	RpID        string `json:"rpId"`
	RpCode      string `json:"rpCode"`
	SpuID       string `json:"spuId"`
	SpuCode     string `json:"spuCode"`
	ProductName string `json:"productName"`
	ProductType string `json:"productType"`
	SpecName    string `json:"specName"`
}

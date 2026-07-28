package models

import "time"

type ProductMp struct {
	MpID                 string     `gorm:"column:mp_id;type:char(36);primaryKey" json:"mpId"`
	MpCode               string     `gorm:"column:mp_code;type:varchar(16)" json:"mpCode"`
	RpID                 string     `gorm:"column:rp_id;type:char(36)" json:"rpId"`
	EnterpriseID         string     `gorm:"column:enterprise_id;type:char(36)" json:"enterpriseId"`
	ApprovalNo           string     `gorm:"column:approval_no;type:varchar(128)" json:"approvalNo"`
	ApprovalNoNormalized string     `gorm:"column:approval_no_normalized;type:varchar(128)" json:"approvalNoNormalized"`
	BrandName            *string    `gorm:"column:brand_name;type:varchar(128)" json:"brandName"`
	Description          *string    `gorm:"column:description;type:varchar(2000)" json:"description"`
	Status               int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	RowVersion           int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID            *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID            *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate           *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate           *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag              int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductMp) TableName() string {
	return "product_mp"
}

type ProductMpListRequest struct {
	RpID string `form:"rpId" binding:"required"`
}

type SaveProductMpRequest struct {
	RpID               string  `json:"rpId" binding:"required"`
	EnterpriseID       string  `json:"enterpriseId" binding:"required"`
	ApprovalNo         string  `json:"approvalNo" binding:"required,max=128"`
	BrandName          *string `json:"brandName" binding:"omitempty,max=128"`
	Description        *string `json:"description" binding:"omitempty,max=2000"`
	Status             int     `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int     `json:"expectedRowVersion"`
}

type UpdateProductMpStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1"`
}

type ProductMpResponse struct {
	MpID           string  `json:"mpId"`
	MpCode         string  `json:"mpCode"`
	RpID           string  `json:"rpId"`
	RpCode         string  `json:"rpCode"`
	SpecName       string  `json:"specName"`
	SpuID          string  `json:"spuId"`
	SpuCode        string  `json:"spuCode"`
	ProductName    string  `json:"productName"`
	ProductType    string  `json:"productType"`
	EnterpriseID   string  `json:"enterpriseId"`
	EnterpriseCode string  `json:"enterpriseCode"`
	EnterpriseName string  `json:"enterpriseName"`
	ApprovalNo     string  `json:"approvalNo"`
	BrandName      *string `json:"brandName"`
	Description    *string `json:"description"`
	Status         int     `json:"status"`
	RowVersion     int     `json:"rowVersion"`
	CreateDate     *string `json:"createDate"`
	UpdateDate     *string `json:"updateDate"`
}

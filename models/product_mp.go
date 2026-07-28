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
	RpID string `form:"rpId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品规格ID
}

type SaveProductMpRequest struct {
	RpID               string  `json:"rpId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品规格ID
	EnterpriseID       string  `json:"enterpriseId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	ApprovalNo         string  `json:"approvalNo" binding:"required,max=128" example:"国药准字H20260001"`                  // 批准文号
	BrandName          *string `json:"brandName" binding:"omitempty,max=128" example:"品牌名"`                            // 品牌名
	Description        *string `json:"description" binding:"omitempty,max=2000" example:"产品描述"`                        // 产品描述
	Status             int     `json:"status" binding:"oneof=0 1" example:"1"`                                         // 状态
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                                                 // 期望行版本号
}

type UpdateProductMpStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ProductMpResponse struct {
	MpID           string  `json:"mpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品MP ID
	MpCode         string  `json:"mpCode" example:"MP001"`                                      // MP编码
	RpID           string  `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品规格ID
	RpCode         string  `json:"rpCode" example:"RP001"`                                      // 规格编码
	SpecName       string  `json:"specName" example:"0.25g"`                                    // 规格名称
	SpuID          string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SPU ID
	SpuCode        string  `json:"spuCode" example:"SPU001"`                                    // SPU编码
	ProductName    string  `json:"productName" example:"阿莫西林胶囊"`                                // 产品名称
	ProductType    string  `json:"productType" example:"DRUG"`                                  // 产品类型
	EnterpriseID   string  `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode string  `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName string  `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ApprovalNo     string  `json:"approvalNo" example:"国药准字H20260001"`                          // 批准文号
	BrandName      *string `json:"brandName" example:"品牌名"`                                     // 品牌名
	Description    *string `json:"description" example:"产品描述"`                                  // 产品描述
	Status         int     `json:"status" example:"1"`                                          // 状态
	RowVersion     int     `json:"rowVersion" example:"1"`                                      // 版本号
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate     *string `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
}

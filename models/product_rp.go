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
	Page        int    `form:"page" example:"1"`                                     // 页码
	PageSize    int    `form:"pageSize" example:"20"`                                // 每页数量
	Keyword     string `form:"keyword" example:"0.25g"`                              // 关键词搜索
	SpuID       string `form:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	ProductType string `form:"productType" example:"DRUG"`                           // 产品类型
	Status      *int   `form:"status" example:"1"`                                   // 状态
	Sorts       string `form:"sorts" example:"specName,desc;createDate,desc"`        // 排序
}

type SaveProductRpRequest struct {
	SpuID              string  `json:"spuId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpecName           string  `json:"specName" binding:"required,max=128" example:"0.25g"`                     // 规格名称
	DosageForm         *string `json:"dosageForm" binding:"omitempty,max=64" example:"胶囊剂"`                     // 剂型
	StrengthText       *string `json:"strengthText" binding:"omitempty,max=128" example:"0.25g"`                // 规格含量
	Description        *string `json:"description" binding:"omitempty,max=2000" example:"规格描述"`                 // 规格描述
	Status             int     `json:"status" binding:"oneof=0 1" example:"1"`                                  // 状态
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                                          // 期望行版本号
}

type UpdateProductRpStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ProductRpOptionsRequest struct {
	Keyword     string `form:"keyword" example:"0.25g"`                              // 关键词搜索
	SpuID       string `form:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	ProductType string `form:"productType" example:"DRUG"`                           // 产品类型
	PageSize    int    `form:"pageSize" example:"20"`                                // 每页数量
}

type ProductRpResponse struct {
	RpID         string  `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`  // 产品规格ID
	RpCode       string  `json:"rpCode" example:"RP001"`                               // 规格编码
	SpuID        string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpuCode      string  `json:"spuCode" example:"SPU001"`                             // SPU编码
	ProductName  string  `json:"productName" example:"阿莫西林胶囊"`                         // 产品名称
	ProductType  string  `json:"productType" example:"DRUG"`                           // 产品类型
	SpecName     string  `json:"specName" example:"0.25g"`                             // 规格名称
	DosageForm   *string `json:"dosageForm" example:"胶囊剂"`                             // 剂型
	StrengthText *string `json:"strengthText" example:"0.25g"`                         // 规格含量
	Description  *string `json:"description" example:"规格描述"`                           // 规格描述
	Status       int     `json:"status" example:"1"`                                   // 状态
	RowVersion   int     `json:"rowVersion" example:"1"`                               // 版本号
	CreateDate   *string `json:"createDate" example:"2026-01-15 09:00:00"`             // 创建时间
	UpdateDate   *string `json:"updateDate" example:"2026-01-15 09:00:00"`             // 更新时间
}

type ProductRpOptionResponse struct {
	RpID        string `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`  // 产品规格ID
	RpCode      string `json:"rpCode" example:"RP001"`                               // 规格编码
	SpuID       string `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpuCode     string `json:"spuCode" example:"SPU001"`                             // SPU编码
	ProductName string `json:"productName" example:"阿莫西林胶囊"`                         // 产品名称
	ProductType string `json:"productType" example:"DRUG"`                           // 产品类型
	SpecName    string `json:"specName" example:"0.25g"`                             // 规格名称
}

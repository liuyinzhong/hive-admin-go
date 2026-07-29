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
	Page        int    `form:"page" example:"1"`                                 // 页码
	PageSize    int    `form:"pageSize" example:"20"`                            // 每页数量
	Keyword     string `form:"keyword" example:"阿莫西林"`                           // 关键词搜索
	ProductType string `form:"productType" example:"DRUG"`                       // 产品类型
	Status      *int   `form:"status" example:"1"`                               // 状态
	Sorts       string `form:"sorts" example:"productName,desc;createDate,desc"` // 排序
}

type SaveProductSpuRequest struct {
	ProductName        string  `json:"productName" binding:"required,max=128" example:"阿莫西林胶囊"` // 产品名称
	ShortName          *string `json:"shortName" binding:"omitempty,max=64" example:"阿莫西林"`     // 产品简称
	ProductType        string  `json:"productType" binding:"required,max=32" example:"DRUG"`    // 产品类型
	Description        *string `json:"description" binding:"omitempty,max=2000" example:"产品描述"` // 产品描述
	Status             int     `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                          // 期望行版本号
}

type UpdateProductSpuStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ProductSpuOptionsRequest struct {
	Keyword     string `form:"keyword" example:"阿莫西林"`     // 关键词搜索
	ProductType string `form:"productType" example:"DRUG"` // 产品类型
	PageSize    int    `form:"pageSize" example:"20"`      // 每页数量
}

type ProductSpuResponse struct {
	SpuID       string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpuCode     string  `json:"spuCode" example:"SPU001"`                             // SPU编码
	ProductName string  `json:"productName" example:"阿莫西林胶囊"`                         // 产品名称
	ShortName   *string `json:"shortName" example:"阿莫西林"`                             // 产品简称
	ProductType string  `json:"productType" example:"DRUG"`                           // 产品类型
	Description *string `json:"description" example:"产品描述"`                           // 产品描述
	Status      int     `json:"status" example:"1"`                                   // 状态
	RowVersion  int     `json:"rowVersion" example:"1"`                               // 版本号
	CreateDate  *string `json:"createDate" example:"2026-01-15 09:00:00"`             // 创建时间
	UpdateDate  *string `json:"updateDate" example:"2026-01-15 09:00:00"`             // 更新时间
}

type ProductSpuDetailResponse struct {
	SpuID       string                        `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpuCode     string                        `json:"spuCode" example:"SPU001"`                             // SPU编码
	ProductName string                        `json:"productName" example:"阿莫西林胶囊"`                         // 产品名称
	ShortName   *string                       `json:"shortName" example:"阿莫西林"`                             // 产品简称
	ProductType string                        `json:"productType" example:"DRUG"`                           // 产品类型
	Description *string                       `json:"description" example:"产品描述"`                           // 产品描述
	Status      int                           `json:"status" example:"1"`                                   // SPU状态：0停用 1启用
	RowVersion  int                           `json:"rowVersion" example:"1"`                               // SPU版本号
	CreateDate  *string                       `json:"createDate" example:"2026-01-15 09:00:00"`             // 创建时间
	UpdateDate  *string                       `json:"updateDate" example:"2026-01-15 09:00:00"`             // 更新时间
	Rows        []ProductSpuDetailRowResponse `json:"rows"`                                                 // 合并单元格表格行，status仅代表SKU状态
}

type ProductSpuDetailRowResponse struct {
	SpuID           string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SPU ID
	SpuCode         string  `json:"spuCode" example:"SPU001"`                                    // SPU编码
	ProductName     string  `json:"productName" example:"阿莫西林胶囊"`                                // 产品名称
	ShortName       *string `json:"shortName" example:"阿莫西林"`                                    // 产品简称
	ProductType     string  `json:"productType" example:"DRUG"`                                  // 产品类型
	RpID            *string `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 规格产品ID，未维护时为空
	RpCode          *string `json:"rpCode" example:"RP001"`                                      // RP编码
	SpecName        *string `json:"specName" example:"0.25g"`                                    // 规格名称
	DosageForm      *string `json:"dosageForm" example:"胶囊剂"`                                    // 剂型/形态
	StrengthText    *string `json:"strengthText" example:"0.25g"`                                // 含量/规格文本
	RpRowVersion    *int    `json:"rpRowVersion" example:"1"`                                    // RP版本号
	MpID            *string `json:"mpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 厂家产品ID，未维护时为空
	MpCode          *string `json:"mpCode" example:"MP001"`                                      // MP编码
	EnterpriseID    *string `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode  *string `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName  *string `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ApprovalNo      *string `json:"approvalNo" example:"国药准字H20260001"`                          // 批准文号/注册证号/备案号
	BrandName       *string `json:"brandName" example:"品牌名"`                                     // 品牌/商品名
	MpRowVersion    *int    `json:"mpRowVersion" example:"1"`                                    // MP版本号
	SkuID           *string `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // SKU ID，未维护时为空
	SkuCode         *string `json:"skuCode" example:"SKU001"`                                    // SKU编码
	PackageSpecName *string `json:"packageSpecName" example:"10粒/盒"`                             // 包装规格名称
	PackageQuantity *int    `json:"packageQuantity" example:"10"`                                // 包装数量
	MinUnitName     *string `json:"minUnitName" example:"粒"`                                     // 最小单位
	PackageUnitName *string `json:"packageUnitName" example:"盒"`                                 // 包装单位
	Barcode         *string `json:"barcode" example:"6901234567890"`                             // 商品条码
	Gtin            *string `json:"gtin" example:"06901234567890"`                               // GTIN
	UdiDi           *string `json:"udiDi" example:"(01)06901234567890"`                          // UDI-DI
	AllowSplit      *int    `json:"allowSplit" example:"0"`                                      // 是否允许拆零：0否 1是
	SkuRowVersion   *int    `json:"skuRowVersion" example:"1"`                                   // SKU版本号
	Status          *int    `json:"status" example:"1"`                                          // SKU状态：0停用 1启用；无SKU时为空
}

type ProductSpuOptionResponse struct {
	SpuID       string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品SPU ID
	SpuCode     string  `json:"spuCode" example:"SPU001"`                             // SPU编码
	ProductName string  `json:"productName" example:"阿莫西林胶囊"`                         // 产品名称
	ShortName   *string `json:"shortName" example:"阿莫西林"`                             // 产品简称
	ProductType string  `json:"productType" example:"DRUG"`                           // 产品类型
}

package models

import "time"

const (
	ProductSkuTraceModeNone     = "NONE"
	ProductSkuTraceModeRequired = "REQUIRED"
)

type ProductSku struct {
	SkuID             string     `gorm:"column:sku_id;type:char(36);primaryKey" json:"skuId"`
	SkuCode           string     `gorm:"column:sku_code;type:varchar(16)" json:"skuCode"`
	MpID              string     `gorm:"column:mp_id;type:char(36)" json:"mpId"`
	PackageSpecName   string     `gorm:"column:package_spec_name;type:varchar(128)" json:"packageSpecName"`
	PackConversion    int        `gorm:"column:pack_conversion;type:int" json:"packConversion"`
	MinUnitName       string     `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	PackageUnitName   string     `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	CartonUnitName    string     `gorm:"column:carton_unit_name;type:varchar(32)" json:"cartonUnitName"`
	CartonConversion  int        `gorm:"column:carton_conversion;type:int" json:"cartonConversion"`
	CartonSpecName    string     `gorm:"column:carton_spec_name;type:varchar(128)" json:"cartonSpecName"`
	FullChainSpecName string     `gorm:"column:full_chain_spec_name;type:varchar(256)" json:"fullChainSpecName"`
	Barcode           *string    `gorm:"column:barcode;type:varchar(64)" json:"barcode"`
	Gtin              *string    `gorm:"column:gtin;type:varchar(64)" json:"gtin"`
	UdiDi             *string    `gorm:"column:udi_di;type:varchar(128)" json:"udiDi"`
	AllowSplit        int        `gorm:"column:allow_split;type:tinyint;default:0" json:"allowSplit"`
	TraceMode         string     `gorm:"column:trace_mode;type:varchar(16);default:NONE" json:"traceMode"`
	Description       *string    `gorm:"column:description;type:varchar(2000)" json:"description"`
	Status            int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	RowVersion        int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID         *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID         *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag           int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ProductSku) TableName() string {
	return "product_sku"
}

type ProductSkuListRequest struct {
	MpID     string `form:"mpId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品MP ID
	Page     int    `form:"page" example:"1"`                                                       // 页码
	PageSize int    `form:"pageSize" example:"20"`                                                  // 每页数量
	Sorts    string `form:"sorts" example:"packageSpecName,desc;createDate,desc"`                   // 排序
}

type SaveProductSkuRequest struct {
	MpID               string  `json:"mpId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品MP ID
	PackConversion     int     `json:"packConversion" binding:"required,min=1,max=999999" example:"10"`        // 包装换算系数
	MinUnitName        string  `json:"minUnitName" binding:"required,max=32" example:"粒"`                      // 最小单位名称
	PackageUnitName    string  `json:"packageUnitName" binding:"required,max=32" example:"盒"`                  // 包装单位名称
	CartonUnitName     string  `json:"cartonUnitName" binding:"required,max=32" example:"箱"`                   // 大包装单位名称
	CartonConversion   int     `json:"cartonConversion" binding:"required,min=1,max=999999" example:"20"`      // 大包装换算系数
	Barcode            *string `json:"barcode" binding:"omitempty,max=64" example:"6901234567890"`             // 条形码
	Gtin               *string `json:"gtin" binding:"omitempty,max=64" example:"06901234567890"`               // GTIN码
	UdiDi              *string `json:"udiDi" binding:"omitempty,max=128" example:"(01)06901234567890"`         // UDI-DI码
	AllowSplit         int     `json:"allowSplit" binding:"oneof=0 1" example:"0"`                             // 是否允许拆零
	TraceMode          string  `json:"traceMode" binding:"required,oneof=NONE REQUIRED" example:"NONE"`        // 追溯码管理模式
	Description        *string `json:"description" binding:"omitempty,max=2000" example:"包装描述"`                // 包装描述
	Status             int     `json:"status" binding:"oneof=0 1" example:"1"`                                 // 状态
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                                         // 期望行版本号
}

type UpdateProductSkuStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ProductSkuOptionsRequest struct {
	MpID     string `form:"mpId" example:"550e8400-e29b-41d4-a716-446655440000"` // 产品MP ID
	Keyword  string `form:"keyword" example:"10粒"`                               // 关键词搜索
	PageSize int    `form:"pageSize" example:"20"`                               // 每页数量，默认50，最大100
}

type ProductSkuResponse struct {
	SpuID             string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SPU ID
	SpuCode           string  `json:"spuCode" example:"SPU001"`                                    // SPU编码
	ProductName       string  `json:"productName" example:"阿莫西林胶囊"`                                // 产品名称
	ProductType       string  `json:"productType" example:"DRUG"`                                  // 产品类型
	RpID              string  `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品规格ID
	RpCode            string  `json:"rpCode" example:"RP001"`                                      // 规格编码
	SpecName          string  `json:"specName" example:"0.25g"`                                    // 规格名称
	MpID              string  `json:"mpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品MP ID
	MpCode            string  `json:"mpCode" example:"MP001"`                                      // MP编码
	EnterpriseID      string  `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode    string  `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName    string  `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ApprovalNo        string  `json:"approvalNo" example:"国药准字H20260001"`                          // 批准文号
	BrandName         *string `json:"brandName" example:"品牌名"`                                     // 品牌名
	SkuID             string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SKU ID
	SkuCode           string  `json:"skuCode" example:"SKU001"`                                    // SKU编码
	PackageSpecName   string  `json:"packageSpecName" example:"10粒/盒"`                             // 包装规格名称
	PackConversion    int     `json:"packConversion" example:"10"`                                 // 包装换算系数
	MinUnitName       string  `json:"minUnitName" example:"粒"`                                     // 最小单位名称
	PackageUnitName   string  `json:"packageUnitName" example:"盒"`                                 // 包装单位名称
	CartonUnitName    string  `json:"cartonUnitName" example:"箱"`                                  // 大包装单位名称
	CartonConversion  int     `json:"cartonConversion" example:"20"`                               // 大包装换算系数
	CartonSpecName    string  `json:"cartonSpecName" example:"20盒/箱"`                              // 大包装规格名称
	FullChainSpecName string  `json:"fullChainSpecName" example:"1箱/20盒/200粒"`                     // 全链路规格名称
	Barcode           *string `json:"barcode" example:"6901234567890"`                             // 条形码
	Gtin              *string `json:"gtin" example:"06901234567890"`                               // GTIN码
	UdiDi             *string `json:"udiDi" example:"(01)06901234567890"`                          // UDI-DI码
	AllowSplit        int     `json:"allowSplit" example:"0"`                                      // 是否允许拆零
	TraceMode         string  `json:"traceMode" example:"REQUIRED"`                                // 追溯码管理模式
	Description       *string `json:"description" example:"包装描述"`                                  // 包装描述
	Status            int     `json:"status" example:"1"`                                          // 状态
	RowVersion        int     `json:"rowVersion" example:"1"`                                      // 版本号
	CreateDate        *string `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate        *string `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
}

type ProductSkuOptionResponse struct {
	SpuID             string  `json:"spuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SPU ID
	SpuCode           string  `json:"spuCode" example:"SPU001"`                                    // SPU编码
	ProductName       string  `json:"productName" example:"阿莫西林胶囊"`                                // 产品名称
	ShortName         *string `json:"shortName" example:"阿莫西林"`                                    // 产品简称
	ProductType       string  `json:"productType" example:"DRUG"`                                  // 产品类型
	SpuDescription    *string `json:"spuDescription" example:"产品描述"`                               // SPU描述
	RpID              string  `json:"rpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品规格ID
	RpCode            string  `json:"rpCode" example:"RP001"`                                      // 规格编码
	SpecName          string  `json:"specName" example:"0.25g"`                                    // 规格名称
	DosageForm        *string `json:"dosageForm" example:"胶囊剂"`                                    // 剂型
	StrengthText      *string `json:"strengthText" example:"0.25g"`                                // 规格含量
	RpDescription     *string `json:"rpDescription" example:"规格描述"`                                // RP描述
	MpID              string  `json:"mpId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 产品MP ID
	MpCode            string  `json:"mpCode" example:"MP001"`                                      // MP编码
	EnterpriseID      string  `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode    string  `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName    string  `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ApprovalNo        string  `json:"approvalNo" example:"国药准字H20260001"`                          // 批准文号
	BrandName         *string `json:"brandName" example:"品牌名"`                                     // 品牌名
	MpDescription     *string `json:"mpDescription" example:"厂家产品描述"`                              // MP描述
	SkuID             string  `json:"skuId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 产品SKU ID
	SkuCode           string  `json:"skuCode" example:"SKU001"`                                    // SKU编码
	PackageSpecName   string  `json:"packageSpecName" example:"10粒/盒"`                             // 包装规格名称
	PackConversion    int     `json:"packConversion" example:"10"`                                 // 包装换算系数
	MinUnitName       string  `json:"minUnitName" example:"粒"`                                     // 最小单位名称
	PackageUnitName   string  `json:"packageUnitName" example:"盒"`                                 // 包装单位名称
	CartonUnitName    string  `json:"cartonUnitName" example:"箱"`                                  // 大包装单位名称
	CartonConversion  int     `json:"cartonConversion" example:"20"`                               // 大包装换算系数
	CartonSpecName    string  `json:"cartonSpecName" example:"20盒/箱"`                              // 大包装规格名称
	FullChainSpecName string  `json:"fullChainSpecName" example:"1箱/20盒/200粒"`                     // 全链路规格名称
	Barcode           *string `json:"barcode" example:"6901234567890"`                             // 条形码
	Gtin              *string `json:"gtin" example:"06901234567890"`                               // GTIN码
	UdiDi             *string `json:"udiDi" example:"(01)06901234567890"`                          // UDI-DI码
	AllowSplit        int     `json:"allowSplit" example:"0"`                                      // 是否允许拆零
	TraceMode         string  `json:"traceMode" example:"REQUIRED"`                                // 追溯码管理模式
	Description       *string `json:"description" example:"包装描述"`                                  // SKU描述
	RowVersion        int     `json:"rowVersion" example:"1"`                                      // SKU版本号
	CreateDate        *string `json:"createDate" example:"2026-01-15 09:00:00"`                    // SKU创建时间
	UpdateDate        *string `json:"updateDate" example:"2026-01-15 09:00:00"`                    // SKU更新时间
}

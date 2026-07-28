package models

import "time"

const (
	EnterpriseTypeEnterprise        = "ENTERPRISE"
	EnterpriseTypeMedicalOrg        = "MEDICAL_ORG"
	EnterpriseTypeIndividual        = "INDIVIDUAL"
	EnterpriseTypePublicInstitution = "PUBLIC_INSTITUTION"
	EnterpriseTypeOther             = "OTHER"

	EnterpriseRoleManufacturer = "MANUFACTURER"
	EnterpriseRoleMAH          = "MAH"
	EnterpriseRoleRegistrant   = "REGISTRANT"
	EnterpriseRoleFiler        = "FILER"
	EnterpriseRoleImportAgent  = "IMPORT_AGENT"
	EnterpriseRoleSupplier     = "SUPPLIER"
	EnterpriseRoleDistributor  = "DISTRIBUTOR"
	EnterpriseRoleDealer       = "DEALER"
	EnterpriseRoleCustomer     = "CUSTOMER"
)

type BaseEnterprise struct {
	EnterpriseID             string     `gorm:"column:enterprise_id;type:char(36);primaryKey" json:"enterpriseId"`
	EnterpriseCode           string     `gorm:"column:enterprise_code;type:varchar(16)" json:"enterpriseCode"`
	EnterpriseName           string     `gorm:"column:enterprise_name;type:varchar(128)" json:"enterpriseName"`
	EnterpriseNameNormalized string     `gorm:"column:enterprise_name_normalized;type:varchar(128)" json:"enterpriseNameNormalized"`
	ShortName                *string    `gorm:"column:short_name;type:varchar(64)" json:"shortName"`
	ShortNameNormalized      *string    `gorm:"column:short_name_normalized;type:varchar(64)" json:"shortNameNormalized"`
	UnifiedCreditCode        *string    `gorm:"column:unified_credit_code;type:varchar(32)" json:"unifiedCreditCode"`
	EnterpriseType           string     `gorm:"column:enterprise_type;type:varchar(32)" json:"enterpriseType"`
	ContactName              *string    `gorm:"column:contact_name;type:varchar(64)" json:"contactName"`
	ContactPhone             *string    `gorm:"column:contact_phone;type:varchar(32)" json:"contactPhone"`
	Address                  *string    `gorm:"column:address;type:varchar(512)" json:"address"`
	Status                   int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark                   *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion               int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID                *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID                *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate               *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate               *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                  int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (BaseEnterprise) TableName() string {
	return "base_enterprise"
}

type BaseEnterpriseRole struct {
	EnterpriseRoleID string     `gorm:"column:enterprise_role_id;type:char(36);primaryKey" json:"enterpriseRoleId"`
	EnterpriseID     string     `gorm:"column:enterprise_id;type:char(36)" json:"enterpriseId"`
	RoleType         string     `gorm:"column:role_type;type:varchar(32)" json:"roleType"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (BaseEnterpriseRole) TableName() string {
	return "base_enterprise_role"
}

type BaseCodeSequence struct {
	SequenceType string     `gorm:"column:sequence_type;type:varchar(32);primaryKey" json:"sequenceType"`
	Prefix       string     `gorm:"column:prefix;type:varchar(16)" json:"prefix"`
	CurrentValue int        `gorm:"column:current_value;type:int;default:0" json:"currentValue"`
	NumberLength int        `gorm:"column:number_length;type:tinyint;default:6" json:"numberLength"`
	Remark       *string    `gorm:"column:remark;type:varchar(256)" json:"remark"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (BaseCodeSequence) TableName() string {
	return "base_code_sequence"
}

type EnterpriseListRequest struct {
	Page           int    `form:"page" example:"1"`                                    // 页码
	PageSize       int    `form:"pageSize" example:"20"`                               // 每页数量
	Keyword        string `form:"keyword" example:"张三企业"`                              // 关键词搜索
	EnterpriseType string `form:"enterpriseType" example:"ENTERPRISE"`                 // 企业类型
	RoleTypes      string `form:"roleTypes" example:"MANUFACTURER,MAH"`                // 角色类型（多个逗号分隔）
	Status         *int   `form:"status" example:"1"`                                  // 状态
	Sorts          string `form:"sorts" example:"enterpriseName,desc;createDate,desc"` // 排序
}

type SaveEnterpriseRequest struct {
	EnterpriseName     string   `json:"enterpriseName" binding:"required,max=128" example:"张三企业"`                  // 企业名称
	ShortName          *string  `json:"shortName" binding:"omitempty,max=64" example:"张企"`                         // 企业简称
	UnifiedCreditCode  *string  `json:"unifiedCreditCode" binding:"omitempty,max=32" example:"91110108MA01XXXXXX"` // 统一社会信用代码
	EnterpriseType     string   `json:"enterpriseType" binding:"required,max=32" example:"ENTERPRISE"`             // 企业类型
	Roles              []string `json:"roles" binding:"required,min=1" example:"[\"MANUFACTURER\",\"MAH\"]"`       // 企业角色列表
	ContactName        *string  `json:"contactName" binding:"omitempty,max=64" example:"张三"`                       // 联系人姓名
	ContactPhone       *string  `json:"contactPhone" binding:"omitempty,max=32" example:"13800138000"`             // 联系人电话
	Address            *string  `json:"address" binding:"omitempty,max=512" example:"北京市朝阳区"`                      // 地址
	Status             int      `json:"status" binding:"oneof=0 1" example:"1"`                                    // 状态
	Remark             *string  `json:"remark" binding:"omitempty,max=512" example:"备注信息"`                         // 备注
	ExpectedRowVersion int      `json:"expectedRowVersion" example:"1"`                                            // 期望行版本号
}

type UpdateEnterpriseStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type EnterpriseOptionsRequest struct {
	Keyword  string `form:"keyword" example:"张三企业"`          // 关键词搜索
	RoleType string `form:"roleType" example:"MANUFACTURER"` // 角色类型
	PageSize int    `form:"pageSize" example:"20"`           // 每页数量
}

type EnterpriseResponse struct {
	EnterpriseID      string   `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode    string   `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName    string   `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ShortName         *string  `json:"shortName" example:"张企"`                                      // 企业简称
	UnifiedCreditCode *string  `json:"unifiedCreditCode" example:"91110108MA01XXXXXX"`              // 统一社会信用代码
	EnterpriseType    string   `json:"enterpriseType" example:"ENTERPRISE"`                         // 企业类型
	Roles             []string `json:"roles" example:"[\"id1\",\"id2\"]"`                           // 企业角色列表
	ContactName       *string  `json:"contactName" example:"张三"`                                    // 联系人姓名
	ContactPhone      *string  `json:"contactPhone" example:"13800138000"`                          // 联系人电话
	Address           *string  `json:"address" example:"北京市朝阳区"`                                    // 地址
	Status            int      `json:"status" example:"1"`                                          // 状态
	Remark            *string  `json:"remark" example:"备注信息"`                                       // 备注
	RowVersion        int      `json:"rowVersion" example:"1"`                                      // 版本号
	CreateDate        *string  `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate        *string  `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
}

type EnterpriseOptionResponse struct {
	EnterpriseID   string   `json:"enterpriseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 企业主体ID
	EnterpriseCode string   `json:"enterpriseCode" example:"ENT001"`                             // 企业编码
	EnterpriseName string   `json:"enterpriseName" example:"张三企业"`                               // 企业名称
	ShortName      *string  `json:"shortName" example:"张企"`                                      // 企业简称
	EnterpriseType string   `json:"enterpriseType" example:"ENTERPRISE"`                         // 企业类型
	Roles          []string `json:"roles" example:"[\"id1\",\"id2\"]"`                           // 企业角色列表
}

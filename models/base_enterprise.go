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
	Page           int    `form:"page" example:"1"`
	PageSize       int    `form:"pageSize" example:"20"`
	Keyword        string `form:"keyword"`
	EnterpriseType string `form:"enterpriseType"`
	RoleTypes      string `form:"roleTypes"`
	Status         *int   `form:"status"`
	Sorts          string `form:"sorts"`
}

type SaveEnterpriseRequest struct {
	EnterpriseName     string   `json:"enterpriseName" binding:"required,max=128"`
	ShortName          *string  `json:"shortName" binding:"omitempty,max=64"`
	UnifiedCreditCode  *string  `json:"unifiedCreditCode" binding:"omitempty,max=32"`
	EnterpriseType     string   `json:"enterpriseType" binding:"required,max=32"`
	Roles              []string `json:"roles" binding:"required,min=1"`
	ContactName        *string  `json:"contactName" binding:"omitempty,max=64"`
	ContactPhone       *string  `json:"contactPhone" binding:"omitempty,max=32"`
	Address            *string  `json:"address" binding:"omitempty,max=512"`
	Status             int      `json:"status" binding:"oneof=0 1"`
	Remark             *string  `json:"remark" binding:"omitempty,max=512"`
	ExpectedRowVersion int      `json:"expectedRowVersion"`
}

type UpdateEnterpriseStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1"`
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1"`
}

type EnterpriseOptionsRequest struct {
	Keyword  string `form:"keyword"`
	RoleType string `form:"roleType"`
	PageSize int    `form:"pageSize"`
}

type EnterpriseResponse struct {
	EnterpriseID      string   `json:"enterpriseId"`
	EnterpriseCode    string   `json:"enterpriseCode"`
	EnterpriseName    string   `json:"enterpriseName"`
	ShortName         *string  `json:"shortName"`
	UnifiedCreditCode *string  `json:"unifiedCreditCode"`
	EnterpriseType    string   `json:"enterpriseType"`
	Roles             []string `json:"roles"`
	ContactName       *string  `json:"contactName"`
	ContactPhone      *string  `json:"contactPhone"`
	Address           *string  `json:"address"`
	Status            int      `json:"status"`
	Remark            *string  `json:"remark"`
	RowVersion        int      `json:"rowVersion"`
	CreateDate        *string  `json:"createDate"`
	UpdateDate        *string  `json:"updateDate"`
}

type EnterpriseOptionResponse struct {
	EnterpriseID   string   `json:"enterpriseId"`
	EnterpriseCode string   `json:"enterpriseCode"`
	EnterpriseName string   `json:"enterpriseName"`
	ShortName      *string  `json:"shortName"`
	EnterpriseType string   `json:"enterpriseType"`
	Roles          []string `json:"roles"`
}

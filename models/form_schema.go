package models

import (
	"encoding/json"
	"time"
)

const (
	FormSchemaLayoutSingle = "single"
	FormSchemaLayoutDouble = "double"
	FormSchemaLayoutTriple = "triple"
)

// SysFormSchema 保存一份可复用的 Vben 表单 Schema。
type SysFormSchema struct {
	FormSchemaID string     `gorm:"column:form_schema_id;type:char(36);primaryKey" json:"formSchemaId"`
	SchemaKey    string     `gorm:"column:schema_key;type:varchar(128);uniqueIndex" json:"schemaKey"`
	SchemaName   string     `gorm:"column:schema_name;type:varchar(128)" json:"schemaName"`
	Category     *string    `gorm:"column:category;type:varchar(64)" json:"category"`
	Layout       string     `gorm:"column:layout;type:varchar(16);default:single" json:"layout"`
	SchemaJSON   string     `gorm:"column:schema_json;type:longtext" json:"-"`
	Status       int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark       *string    `gorm:"column:remark;type:varchar(256)" json:"remark"`
	CreatorID    *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (SysFormSchema) TableName() string { return "sys_form_schema" }

// FormSchemaRule 是后端执行提交校验所需的规则投影。
type FormSchemaRule struct {
	HandlerKey *string       `json:"handlerKey,omitempty"`
	Integer    bool          `json:"integer,omitempty"`
	Length     *int          `json:"length,omitempty"`
	Max        *float64      `json:"max,omitempty"`
	Message    *string       `json:"message,omitempty"`
	Min        *float64      `json:"min,omitempty"`
	Pattern    *string       `json:"pattern,omitempty"`
	Type       string        `json:"type"`
	Values     []interface{} `json:"values,omitempty"`
}

// FormSchemaField 是完整 Vben Schema 的后端最小字段投影。
type FormSchemaField struct {
	Component      string                 `json:"component"`
	ComponentProps map[string]interface{} `json:"componentProps,omitempty"`
	DefaultValue   interface{}            `json:"defaultValue,omitempty"`
	Dependencies   json.RawMessage        `json:"dependencies,omitempty" swaggertype:"object"`
	FieldName      string                 `json:"fieldName"`
	Label          string                 `json:"label,omitempty"`
	Rules          []FormSchemaRule       `json:"rules,omitempty"`
}

// FormSchemaResponse 返回表单基本信息和原始 Schema JSON。
type FormSchemaResponse struct {
	FormSchemaID string          `json:"formSchemaId" example:"UUID"`
	SchemaKey    string          `json:"schemaKey" example:"expense_apply"`
	SchemaName   string          `json:"schemaName" example:"报销申请表"`
	Category     *string         `json:"category" example:"workflow"`
	Layout       string          `json:"layout" example:"single"`
	Schema       json.RawMessage `json:"schema" swaggertype:"array,object"`
	Status       string          `json:"status" example:"1"`
	Remark       *string         `json:"remark" example:"报销流程申请表"`
	CreatorID    *string         `json:"creatorId" example:"UUID"`
	CreatorName  *string         `json:"creatorName" example:"管理员"`
	CreateDate   *string         `json:"createDate" example:"2026-07-15 15:30:26"`
	UpdateDate   *string         `json:"updateDate" example:"2026-07-15 15:30:26"`
}

// UpsertFormSchemaRequest 创建或更新一份表单 Schema。
type UpsertFormSchemaRequest struct {
	SchemaKey  string          `json:"schemaKey" binding:"required" example:"expense_apply"`
	SchemaName string          `json:"schemaName" binding:"required" example:"报销申请表"`
	Category   *string         `json:"category" example:"workflow"`
	Layout     string          `json:"layout" binding:"required" example:"single"`
	Schema     json.RawMessage `json:"schema" binding:"required" swaggertype:"array,object"`
	Status     *string         `json:"status" example:"1"`
	Remark     *string         `json:"remark" example:"报销流程申请表"`
}

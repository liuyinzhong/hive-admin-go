package models

import (
	"encoding/json"
	"time"
)

const (
	PrintDocumentTypePurchaseInbound = "PURCHASE_INBOUND"
	PrintTemplateStatusDraft         = "DRAFT"
	PrintTemplateStatusPublished     = "PUBLISHED"
)

// PrintTemplate 保存单据打印模板的当前草稿和当前已发布内容。
// published_layout 不承担历史版本职责，发布时直接覆盖当前已发布内容。
type PrintTemplate struct {
	TemplateID      string     `gorm:"column:template_id;type:char(36);primaryKey" json:"templateId"`
	DocumentType    string     `gorm:"column:document_type;type:varchar(64);not null" json:"documentType"`
	TemplateName    string     `gorm:"column:template_name;type:varchar(128);not null" json:"templateName"`
	DraftLayout     string     `gorm:"column:draft_layout;type:longtext;not null" json:"draftLayout"`
	PublishedLayout *string    `gorm:"column:published_layout;type:longtext" json:"publishedLayout"`
	Status          string     `gorm:"column:status;type:varchar(16);not null" json:"status"`
	RowVersion      int        `gorm:"column:row_version;type:int;not null;default:1" json:"rowVersion"`
	CreatorID       *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID       *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate      *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate      *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (PrintTemplate) TableName() string {
	return "print_template"
}

type PrintTemplateListRequest struct {
	Page         int    `form:"page" example:"1"`
	PageSize     int    `form:"pageSize" example:"20"`
	DocumentType string `form:"documentType" example:"PURCHASE_INBOUND"`
	Status       string `form:"status" example:"PUBLISHED"`
	Sorts        string `form:"sorts" example:"updateDate,desc"`
}

type CreatePrintTemplateRequest struct {
	DocumentType string          `json:"documentType" binding:"required" example:"PURCHASE_INBOUND"`
	TemplateName string          `json:"templateName" binding:"required,max=128" example:"采购入库单默认模板"`
	DraftLayout  json.RawMessage `json:"draftLayout" binding:"required" swaggertype:"object"`
}

type UpdatePrintTemplateRequest struct {
	TemplateName string          `json:"templateName" binding:"required,max=128" example:"采购入库单默认模板"`
	DraftLayout  json.RawMessage `json:"draftLayout" binding:"required" swaggertype:"object"`
	RowVersion   int             `json:"rowVersion" binding:"required,min=1" example:"3"`
}

type PublishPrintTemplateRequest struct {
	RowVersion int `json:"rowVersion" binding:"required,min=1" example:"3"`
}

type PrintTemplateListResponse struct {
	TemplateID   string  `json:"templateId" example:"550e8400-e29b-41d4-a716-446655440000"`
	DocumentType string  `json:"documentType" example:"PURCHASE_INBOUND"`
	TemplateName string  `json:"templateName" example:"采购入库单默认模板"`
	Status       string  `json:"status" example:"PUBLISHED"`
	RowVersion   int     `json:"rowVersion" example:"3"`
	HasDraft     bool    `json:"hasDraft" example:"true"`
	HasPublished bool    `json:"hasPublished" example:"true"`
	CreateDate   *string `json:"createDate" example:"2026-08-03 10:00:00"`
	UpdateDate   *string `json:"updateDate" example:"2026-08-03 10:00:00"`
}

type PrintTemplateResponse struct {
	TemplateID      string           `json:"templateId" example:"550e8400-e29b-41d4-a716-446655440000"`
	DocumentType    string           `json:"documentType" example:"PURCHASE_INBOUND"`
	TemplateName    string           `json:"templateName" example:"采购入库单默认模板"`
	Status          string           `json:"status" example:"PUBLISHED"`
	DraftLayout     json.RawMessage  `json:"draftLayout" swaggertype:"object"`
	PublishedLayout *json.RawMessage `json:"publishedLayout" swaggertype:"object"`
	RowVersion      int              `json:"rowVersion" example:"3"`
	CreateDate      *string          `json:"createDate" example:"2026-08-03 10:00:00"`
	UpdateDate      *string          `json:"updateDate" example:"2026-08-03 10:00:00"`
}

type PrintFieldGroup struct {
	Code   string                 `json:"code" example:"header"`
	Name   string                 `json:"name" example:"单据头"`
	Fields []PrintFieldDefinition `json:"fields"`
}

type PrintFieldDefinition struct {
	Path     string `json:"path" example:"header.inboundNo"`
	Label    string `json:"label" example:"入库单号"`
	DataType string `json:"dataType" example:"string"`
	Scope    string `json:"scope" example:"header"`
	Example  string `json:"example" example:"PIN00000001"`
}

type PrintTemplateMetadataResponse struct {
	DocumentTypes []PrintDocumentTypeDefinition `json:"documentTypes"`
	FieldGroups   []PrintFieldGroup             `json:"fieldGroups"`
}

type PrintDocumentTypeDefinition struct {
	Code string `json:"code" example:"PURCHASE_INBOUND"`
	Name string `json:"name" example:"采购入库单"`
}

// PrintDocumentResponse 是统一打印数据协议。
// header/summary 为对象，items 为明细数组，system 为打印时刻等系统注入字段；
// 与 worm-vue3-print 表达式上下文（{header.inboundNo} 等）的取值路径一致。
type PrintDocumentResponse struct {
	DocumentType  string                   `json:"documentType" example:"PURCHASE_INBOUND"`
	SchemaVersion int                      `json:"schemaVersion" example:"1"`
	Header        map[string]interface{}   `json:"header"`
	Items         []map[string]interface{} `json:"items"`
	Summary       map[string]interface{}   `json:"summary"`
	System        map[string]interface{}   `json:"system"`
}

type PrintDocumentBundleResponse struct {
	Template *PrintTemplateResponse `json:"template"`
	Data     *PrintDocumentResponse `json:"data"`
}

package models

import (
	"encoding/json"
	"time"
)

// 自动化动作类型。
const (
	// ActionTypeUpdateField 修改字段值:把当前关联业务对象的指定状态字段改为固定值,仅可挂载被动触发流程。
	ActionTypeUpdateField = "update_field"
	// ActionTypeInsertRecord 插入记录:按字段映射向目标业务表插入一条新记录,仅可挂载手动发起流程。
	ActionTypeInsertRecord = "insert_record"
)

// 插入记录的值来源类型。
const (
	// InsertSourceFixed 固定值。
	InsertSourceFixed = "fixed"
	// InsertSourceForm 表单字段:运行时取流程实例变量中该字段名的值。
	InsertSourceForm = "form"
)

// WfAutomation 工作流自动化动作库。
// 动作是可复用的无状态配置模板,由流程设计器按业务类型筛选后挂载到节点画布(发布时快照);
// 动作库后续修改不影响已挂载的快照,重新挂载或重新发布才生效。
type WfAutomation struct {
	AutomationID   string     `gorm:"column:automation_id;type:char(36);primaryKey" json:"automationId"`
	AutomationName string     `gorm:"column:automation_name;type:varchar(128)" json:"automationName"`
	BusinessType   string     `gorm:"column:business_type;type:varchar(36);index" json:"businessType"` // 业务类型,字典BUSINESS_TYPE的值,作为设计器挂载过滤维度
	ActionType     string     `gorm:"column:action_type;type:varchar(64)" json:"actionType"`           // 动作类型:update_field修改字段值 insert_record插入记录
	ActionConfig   string     `gorm:"column:action_config;type:json" json:"actionConfig"`              // 动作参数JSON,结构随动作类型变化
	Status         int        `gorm:"column:status;type:tinyint;default:0" json:"status"`              // 状态:0启用 1停用
	Remark         *string    `gorm:"column:remark;type:varchar(256)" json:"remark"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfAutomation) TableName() string { return "wf_automation" }

// UpdateFieldConfig 修改字段值动作参数。
type UpdateFieldConfig struct {
	TargetField string `json:"targetField" example:"story_status"` // 目标字段,须在业务类型状态字段白名单内
	TargetValue string `json:"targetValue" example:"10"`           // 目标值,字典字段须为对应字典的合法值
}

// InsertRecordMapping 插入记录的字段映射行。
type InsertRecordMapping struct {
	Field      string `json:"field" example:"story_title"`     // 目标字段,须在动作业务类型可插字段目录内
	SourceType string `json:"sourceType" example:"form"`       // 值来源:fixed固定值 form表单字段
	Value      string `json:"value" example:"0"`               // 固定值,字典字段须为对应字典的合法值
	FormField  string `json:"formField" example:"story_title"` // 表单字段名,弱引用,发布时按流程绑定表单校验存在性
}

// InsertRecordConfig 插入记录动作参数。
// 插入目标即动作自身的业务类型(BusinessType),不单独配置目标业务类型。
type InsertRecordConfig struct {
	Mappings []InsertRecordMapping `json:"mappings"` // 字段映射,动作业务类型必填字段必须全部配置
}

// CreateAutomationRequest 创建自动化动作请求。
// ActionConfig 为动作参数 JSON,结构随 ActionType 变化,由 Service 按类型反序列化校验。
type CreateAutomationRequest struct {
	AutomationName string          `json:"automationName" binding:"required" example:"需求评审通过"`   // 动作名称
	BusinessType   string          `json:"businessType" binding:"required" example:"0"`          // 业务类型,字典BUSINESS_TYPE的值
	ActionType     string          `json:"actionType" binding:"required" example:"update_field"` // 动作类型
	ActionConfig   json.RawMessage `json:"actionConfig" binding:"required"`                      // 动作参数,结构随动作类型变化
	Status         *int            `json:"status" example:"0"`                                   // 状态:0启用 1停用,默认启用
	Remark         *string         `json:"remark" example:"评审节点使用"`                              // 备注
}

// UpdateAutomationRequest 更新自动化动作请求。
type UpdateAutomationRequest struct {
	AutomationName string          `json:"automationName" binding:"required" example:"需求评审通过"`
	BusinessType   string          `json:"businessType" binding:"required" example:"0"`
	ActionType     string          `json:"actionType" binding:"required" example:"update_field"`
	ActionConfig   json.RawMessage `json:"actionConfig" binding:"required"`
	Status         *int            `json:"status" example:"0"`
	Remark         *string         `json:"remark" example:"评审节点使用"`
}

// AutomationUpdateFieldResponse 修改字段值动作的响应参数,附字段元数据供摘要与字典翻译。
type AutomationUpdateFieldResponse struct {
	TargetField      string `json:"targetField" example:"story_status"`
	TargetValue      string `json:"targetValue" example:"10"`
	TargetFieldLabel string `json:"targetFieldLabel" example:"需求状态"`
	DictType         string `json:"dictType" example:"STORY_STATUS"`
}

// AutomationInsertFieldResponse 插入记录映射行的响应,附目录元数据供前端渲染。
type AutomationInsertFieldResponse struct {
	Field      string `json:"field" example:"story_title"`
	FieldLabel string `json:"fieldLabel" example:"需求名称"`
	Required   bool   `json:"required" example:"true"`
	DictType   string `json:"dictType" example:"STORY_STATUS"`
	IsRefField bool   `json:"isRefField" example:"false"`
	SourceType string `json:"sourceType" example:"form"`
	Value      string `json:"value" example:"0"`
	FormField  string `json:"formField" example:"story_title"`
}

// AutomationInsertRecordResponse 插入记录动作的响应参数。
// 目标业务类型即动作自身的业务类型(BusinessType),不重复携带。
type AutomationInsertRecordResponse struct {
	Mappings []AutomationInsertFieldResponse `json:"mappings"`
}

// AutomationResponse 自动化动作响应。
// 参数按动作类型挂载在 UpdateField 或 InsertRecord 上,前端按 ActionType 渲染。
type AutomationResponse struct {
	AutomationID   *string                         `json:"automationId" example:"UUID"`
	AutomationName string                          `json:"automationName" example:"需求评审通过"`
	BusinessType   string                          `json:"businessType" example:"0"`
	ActionType     string                          `json:"actionType" example:"update_field"`
	UpdateField    *AutomationUpdateFieldResponse  `json:"updateField,omitempty"`
	InsertRecord   *AutomationInsertRecordResponse `json:"insertRecord,omitempty"`
	Status         string                          `json:"status" example:"0"`
	Remark         *string                         `json:"remark" example:"评审节点使用"`
	CreatorID      *string                         `json:"creatorId" example:"UUID"`
	CreatorName    *string                         `json:"creatorName" example:"管理员"`
	CreateDate     *string                         `json:"createDate" example:"2026-09-03 10:00:00"`
	UpdateDate     *string                         `json:"updateDate" example:"2026-09-03 10:00:00"`
}

// AutomationFieldMeta 业务类型字段元数据,供动作库表单的目标字段下拉。
// Required 仅可插字段目录返回时携带,标记表级非空字段。
type AutomationFieldMeta struct {
	Field    string `json:"field" example:"story_status"`
	Label    string `json:"label" example:"需求状态"`
	DictType string `json:"dictType" example:"STORY_STATUS"`
	Required bool   `json:"required" example:"false"`
}

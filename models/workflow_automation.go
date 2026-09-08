package models

import "time"

// WfAutomation 工作流自动化动作库。
// 动作是可复用的无状态配置模板,由流程设计器按业务类型筛选后挂载到节点画布(发布时快照);
// 动作库后续修改不影响已挂载的快照,重新挂载或重新发布才生效。
type WfAutomation struct {
	AutomationID   string     `gorm:"column:automation_id;type:char(36);primaryKey" json:"automationId"`
	AutomationName string     `gorm:"column:automation_name;type:varchar(128)" json:"automationName"`
	BusinessType   string     `gorm:"column:business_type;type:varchar(36);index" json:"businessType"` // 业务类型,字典BUSINESS_TYPE的值
	ActionType     string     `gorm:"column:action_type;type:varchar(64)" json:"actionType"`           // 动作类型:update_field修改字段值;预留insert_record插入记录
	ActionConfig   string     `gorm:"column:action_config;type:json" json:"actionConfig"`              // 动作参数JSON,结构随动作类型变化
	Status         int        `gorm:"column:status;type:tinyint;default:0" json:"status"`              // 状态:0启用 1停用
	Remark         *string    `gorm:"column:remark;type:varchar(256)" json:"remark"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfAutomation) TableName() string { return "wf_automation" }

// AutomationActionConfig 自动化动作参数。
// 版本1仅支持修改字段值;插入记录(insert_record)等动作类型扩展时在对应结构上增加字段,
// 存储统一序列化为 action_config JSON。
type AutomationActionConfig struct {
	TargetField string `json:"targetField" example:"story_status"` // 目标字段,须在业务类型可写字段白名单内
	TargetValue string `json:"targetValue" example:"10"`           // 目标值,字典字段须为对应字典的合法值
}

// CreateAutomationRequest 创建自动化动作请求。
type CreateAutomationRequest struct {
	AutomationName string                 `json:"automationName" binding:"required" example:"需求评审通过"`   // 动作名称
	BusinessType   string                 `json:"businessType" binding:"required" example:"0"`          // 业务类型,字典BUSINESS_TYPE的值
	ActionType     string                 `json:"actionType" binding:"required" example:"update_field"` // 动作类型
	ActionConfig   AutomationActionConfig `json:"actionConfig" binding:"required"`                      // 动作参数
	Status         *int                   `json:"status" example:"0"`                                   // 状态:0启用 1停用,默认启用
	Remark         *string                `json:"remark" example:"评审节点使用"`                              // 备注
}

// UpdateAutomationRequest 更新自动化动作请求。
type UpdateAutomationRequest struct {
	AutomationName string                 `json:"automationName" binding:"required" example:"需求评审通过"`
	BusinessType   string                 `json:"businessType" binding:"required" example:"0"`
	ActionType     string                 `json:"actionType" binding:"required" example:"update_field"`
	ActionConfig   AutomationActionConfig `json:"actionConfig" binding:"required"`
	Status         *int                   `json:"status" example:"0"`
	Remark         *string                `json:"remark" example:"评审节点使用"`
}

// AutomationResponse 自动化动作响应。
// TargetFieldLabel 与 DictType 由后端按业务类型注册表填充,供前端渲染摘要和字典下拉。
type AutomationResponse struct {
	AutomationID     *string `json:"automationId" example:"UUID"`
	AutomationName   string  `json:"automationName" example:"需求评审通过"`
	BusinessType     string  `json:"businessType" example:"0"`
	ActionType       string  `json:"actionType" example:"update_field"`
	TargetField      string  `json:"targetField" example:"story_status"`
	TargetValue      string  `json:"targetValue" example:"10"`
	TargetFieldLabel string  `json:"targetFieldLabel" example:"需求状态"`
	DictType         string  `json:"dictType" example:"STORY_STATUS"`
	Status           string  `json:"status" example:"0"`
	Remark           *string `json:"remark" example:"评审节点使用"`
	CreatorID        *string `json:"creatorId" example:"UUID"`
	CreatorName      *string `json:"creatorName" example:"管理员"`
	CreateDate       *string `json:"createDate" example:"2026-09-03 10:00:00"`
	UpdateDate       *string `json:"updateDate" example:"2026-09-03 10:00:00"`
}

// AutomationFieldMeta 业务类型可写字段元数据,供动作库表单的目标字段下拉。
type AutomationFieldMeta struct {
	Field    string `json:"field" example:"story_status"`
	Label    string `json:"label" example:"需求状态"`
	DictType string `json:"dictType" example:"STORY_STATUS"`
}

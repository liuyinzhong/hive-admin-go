package models

import "time"

type WfProcessDefinition struct {
	DefinitionID   string     `gorm:"column:definition_id;type:char(36);primaryKey" json:"definitionId"`
	DefinitionKey  string     `gorm:"column:definition_key;type:varchar(128)" json:"definitionKey"`
	DefinitionName string     `gorm:"column:definition_name;type:varchar(128)" json:"definitionName"`
	Category       *string    `gorm:"column:category;type:varchar(64)" json:"category"`
	Status         int        `gorm:"column:status;type:tinyint;default:0" json:"status"`
	Version        int        `gorm:"column:version;type:int;default:0" json:"version"`
	FlowData       *string    `gorm:"column:flow_data;type:longtext" json:"flowData"`
	Remark         *string    `gorm:"column:remark;type:varchar(256)" json:"remark"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessDefinition) TableName() string {
	return "wf_process_definition"
}

type WorkflowDefinitionResponse struct {
	DefinitionID   *string `json:"definitionId" example:"UUID"`                    // 流程定义ID
	DefinitionKey  string  `json:"definitionKey" example:"story_approval"`         // 流程标识
	DefinitionName string  `json:"definitionName" example:"需求审批流程"`                // 流程名称
	Category       *string `json:"category" example:"dev"`                         // 流程分类
	Status         string  `json:"status" example:"0"`                             // 流程状态：0草稿 1已发布 2已停用
	Version        int     `json:"version" example:"1"`                            // 发布版本号
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"` // LogicFlow画布JSON
	Remark         *string `json:"remark" example:"流程说明"`                          // 备注
	CreatorID      *string `json:"creatorId" example:"UUID"`                       // 创建人ID
	CreatorName    *string `json:"creatorName" example:"管理员"`                      // 创建人姓名
	CreateDate     *string `json:"createDate" example:"2026-05-18 15:30:26"`       // 创建时间
	UpdateDate     *string `json:"updateDate" example:"2026-05-18 15:30:26"`       // 更新时间
}

type CreateWorkflowDefinitionRequest struct {
	DefinitionKey  string  `json:"definitionKey" binding:"required" example:"story_approval"` // 流程标识，系统内唯一
	DefinitionName string  `json:"definitionName" binding:"required" example:"需求审批流程"`        // 流程名称
	Category       *string `json:"category" example:"dev"`                                    // 流程分类
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"`            // LogicFlow画布JSON
	Remark         *string `json:"remark" example:"流程说明"`                                     // 备注
}

type UpdateWorkflowDefinitionRequest struct {
	DefinitionKey  string  `json:"definitionKey" binding:"required" example:"story_approval"` // 流程标识，系统内唯一
	DefinitionName string  `json:"definitionName" binding:"required" example:"需求审批流程"`        // 流程名称
	Category       *string `json:"category" example:"dev"`                                    // 流程分类
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"`            // LogicFlow画布JSON
	Remark         *string `json:"remark" example:"流程说明"`                                     // 备注
}

type UpdateWorkflowCanvasRequest struct {
	FlowData string `json:"flowData" binding:"required" example:"{\"nodes\":[],\"edges\":[]}"` // LogicFlow画布JSON
}

type UpdateWorkflowStatusRequest struct {
	Status string `json:"status" binding:"required" example:"2"` // 流程状态：0草稿 1已发布 2已停用
}

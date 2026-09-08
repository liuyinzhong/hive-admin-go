package models

import "time"

// 流程定义启动类型。
const (
	// WorkflowStartTypeManual 手动发起流程:出现在"发起申请"入口,纯流程实例,不绑定业务对象,发布校验禁止挂载修改字段值动作。
	WorkflowStartTypeManual = 0
	// WorkflowStartTypePassive 被动触发流程:由业务对象发起(创建后自动匹配默认流程或业务侧手动发起),必须携带业务对象。
	WorkflowStartTypePassive = 1
)

type WfProcessDefinition struct {
	DefinitionID   string     `gorm:"column:definition_id;type:char(36);primaryKey" json:"definitionId"`
	DefinitionKey  string     `gorm:"column:definition_key;type:varchar(128)" json:"definitionKey"`
	DefinitionName string     `gorm:"column:definition_name;type:varchar(128)" json:"definitionName"`
	Category       *string    `gorm:"column:category;type:varchar(64)" json:"category"`
	BusinessType   *string    `gorm:"column:business_type;type:varchar(64);index" json:"businessType"` // 业务类型,字典BUSINESS_TYPE的值(0需求/10任务/20缺陷/30版本)
	StartType      int        `gorm:"column:start_type;type:tinyint;default:0" json:"startType"`       // 启动类型:0手动发起流程 1被动触发流程
	IsDefault      int        `gorm:"column:is_default;type:tinyint;default:0" json:"isDefault"`       // 默认流程标志:同业务类型唯一,仅被动触发流程可设,业务对象创建后按此匹配自动发起
	Status         int        `gorm:"column:status;type:tinyint;default:0" json:"status"`
	Version        int        `gorm:"column:version;type:int;default:0" json:"version"`
	FlowData       *string    `gorm:"column:flow_data;type:longtext" json:"flowData"`
	FormSchemaID   *string    `gorm:"column:form_schema_id;type:char(36);index" json:"formSchemaId"`
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
	BusinessType   *string `json:"businessType" example:"0"`                       // 业务类型,字典BUSINESS_TYPE的值
	StartType      int     `json:"startType" example:"0"`                          // 启动类型:0手动发起流程 1被动触发流程
	IsDefault      bool    `json:"isDefault" example:"false"`                      // 默认流程标志(同业务类型唯一,仅被动触发流程可设)
	Status         string  `json:"status" example:"0"`                             // 流程状态：0草稿 1已发布 2已停用
	Version        int     `json:"version" example:"1"`                            // 发布版本号
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"` // LogicFlow画布JSON
	FormSchemaID   *string `json:"formSchemaId" example:"UUID"`                    // 关联表单Schema ID
	Remark         *string `json:"remark" example:"流程说明"`                          // 备注
	CreatorID      *string `json:"creatorId" example:"UUID"`                       // 创建人ID
	CreatorName    *string `json:"creatorName" example:"管理员"`                      // 创建人姓名
	CreateDate     *string `json:"createDate" example:"2026-05-18 15:30:26"`       // 创建时间
	UpdateDate     *string `json:"updateDate" example:"2026-05-18 15:30:26"`       // 更新时间
}

// CreateWorkflowDefinitionRequest 创建流程定义请求。DefinitionKey 由后端通过公共编码流水自动生成，不接受前端传入。
// BusinessType 必填且须为已注册业务类型(字典BUSINESS_TYPE);StartType 默认手动发起;
// IsDefault 仅被动触发流程可设为 true,同业务类型唯一,保存时自动顶掉该类型原默认。
type CreateWorkflowDefinitionRequest struct {
	DefinitionName string  `json:"definitionName" binding:"required" example:"需求审批流程"` // 流程名称
	Category       *string `json:"category" example:"dev"`                             // 流程分类
	BusinessType   string  `json:"businessType" binding:"required" example:"0"`        // 业务类型,字典BUSINESS_TYPE的值
	StartType      *int    `json:"startType" example:"1"`                              // 启动类型:0手动发起(默认) 1被动触发
	IsDefault      *bool   `json:"isDefault" example:"true"`                           // 默认流程标志,仅被动触发流程可设,同业务类型唯一
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"`     // LogicFlow画布JSON
	Remark         *string `json:"remark" example:"流程说明"`                              // 备注
}

// UpdateWorkflowDefinitionRequest 更新流程定义请求。DefinitionKey 由后端自动生成且不可修改，不接受前端传入。
// IsDefault 仅被动触发流程可设为 true,同业务类型唯一,保存时自动顶掉该类型原默认。
type UpdateWorkflowDefinitionRequest struct {
	DefinitionName string  `json:"definitionName" binding:"required" example:"需求审批流程"` // 流程名称
	Category       *string `json:"category" example:"dev"`                             // 流程分类
	BusinessType   string  `json:"businessType" binding:"required" example:"0"`        // 业务类型,字典BUSINESS_TYPE的值
	StartType      *int    `json:"startType" example:"1"`                              // 启动类型:0手动发起(默认) 1被动触发
	IsDefault      *bool   `json:"isDefault" example:"true"`                           // 默认流程标志,仅被动触发流程可设,同业务类型唯一
	FlowData       *string `json:"flowData" example:"{\"nodes\":[],\"edges\":[]}"`     // LogicFlow画布JSON
	Remark         *string `json:"remark" example:"流程说明"`                              // 备注
}

type UpdateWorkflowCanvasRequest struct {
	FlowData string `json:"flowData" binding:"required" example:"{\"nodes\":[],\"edges\":[]}"` // LogicFlow画布JSON
}

// UpdateWorkflowFormSchemaRequest 绑定流程使用的独立表单 Schema。
type UpdateWorkflowFormSchemaRequest struct {
	FormSchemaID string `json:"formSchemaId" binding:"required" example:"UUID"`
}

type UpdateWorkflowStatusRequest struct {
	Status string `json:"status" binding:"required" example:"2"` // 流程状态：0草稿 1已发布 2已停用
}

package models

import (
	"encoding/json"
	"time"
)

const (
	WorkflowInstanceStatusRunning   = 0
	WorkflowInstanceStatusCompleted = 1
	WorkflowInstanceStatusRejected  = 2
	WorkflowInstanceStatusCanceled  = 3

	WorkflowTaskStatusPending  = 0
	WorkflowTaskStatusApproved = 1
	WorkflowTaskStatusRejected = 2
	WorkflowTaskStatusCanceled = 3

	WorkflowCopyStatusUnread = 0
	WorkflowCopyStatusRead   = 1

	WorkflowNodeStatusPlanned    = 0
	WorkflowNodeStatusActive     = 1
	WorkflowNodeStatusCompleted  = 2
	WorkflowNodeStatusTerminated = 3
	WorkflowNodeStatusSuperseded = 4
)

// WfProcessInstance 保存一次流程运行实例及其定义快照。
type WfProcessInstance struct {
	InstanceID        string     `gorm:"column:instance_id;type:char(36);primaryKey" json:"instanceId"`
	InstanceNo        string     `gorm:"column:instance_no;type:varchar(32);uniqueIndex" json:"instanceNo"`
	DefinitionID      string     `gorm:"column:definition_id;type:char(36);index" json:"definitionId"`
	DefinitionKey     string     `gorm:"column:definition_key;type:varchar(128)" json:"definitionKey"`
	DefinitionName    string     `gorm:"column:definition_name;type:varchar(128)" json:"definitionName"`
	DefinitionVersion int        `gorm:"column:definition_version;type:int" json:"definitionVersion"`
	Title             string     `gorm:"column:title;type:varchar(128)" json:"title"`
	StarterID         string     `gorm:"column:starter_id;type:char(36);index" json:"starterId"`
	StarterName       string     `gorm:"column:starter_name;type:varchar(36)" json:"starterName"`
	Status            int        `gorm:"column:status;type:tinyint;default:0;index" json:"status"`
	Variables         string     `gorm:"column:variables;type:longtext" json:"variables"`
	FlowSnapshot      string     `gorm:"column:flow_snapshot;type:longtext" json:"flowSnapshot"`
	FormSnapshot      *string    `gorm:"column:form_snapshot;type:longtext" json:"formSnapshot"`
	FormLayout        string     `gorm:"column:form_layout;type:varchar(16)" json:"formLayout"`
	StartDate         *time.Time `gorm:"column:start_date" json:"startDate"`
	EndDate           *time.Time `gorm:"column:end_date" json:"endDate"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag           int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessInstance) TableName() string { return "wf_process_instance" }

// WfProcessNodeInstance 保存一次具体节点执行及其预计参与人快照。
type WfProcessNodeInstance struct {
	NodeInstanceID   string     `gorm:"column:node_instance_id;type:char(36);primaryKey" json:"nodeInstanceId"`
	InstanceID       string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	NodeID           string     `gorm:"column:node_id;type:varchar(128);index" json:"nodeId"`
	NodeName         string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	NodeType         string     `gorm:"column:node_type;type:varchar(32);index" json:"nodeType"`
	Sequence         int        `gorm:"column:sequence;type:int;index" json:"sequence"`
	RouteVersion     int        `gorm:"column:route_version;type:int;index" json:"routeVersion"`
	Status           int        `gorm:"column:status;type:tinyint;index" json:"status"`
	ApprovalMode     *string    `gorm:"column:approval_mode;type:varchar(16)" json:"approvalMode"`
	BranchEdgeID     *string    `gorm:"column:branch_edge_id;type:varchar(128)" json:"branchEdgeId"`
	ActorIDs         string     `gorm:"column:actor_ids;type:longtext" json:"actorIds"`
	ActorNames       string     `gorm:"column:actor_names;type:longtext" json:"actorNames"`
	FieldPermissions string     `gorm:"column:field_permissions;type:longtext" json:"fieldPermissions"`
	StartDate        *time.Time `gorm:"column:start_date" json:"startDate"`
	EndDate          *time.Time `gorm:"column:end_date" json:"endDate"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (WfProcessNodeInstance) TableName() string { return "wf_process_node_instance" }

// WfProcessTask 保存审批节点为每位审批人生成的任务。
type WfProcessTask struct {
	TaskID         string     `gorm:"column:task_id;type:char(36);primaryKey" json:"taskId"`
	TaskGroupID    string     `gorm:"column:task_group_id;type:char(36);index" json:"taskGroupId"`
	NodeInstanceID string     `gorm:"column:node_instance_id;type:char(36);index" json:"nodeInstanceId"`
	InstanceID     string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	NodeID         string     `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName       string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	AssigneeID     string     `gorm:"column:assignee_id;type:char(36);index" json:"assigneeId"`
	AssigneeName   string     `gorm:"column:assignee_name;type:varchar(36)" json:"assigneeName"`
	ApprovalMode   string     `gorm:"column:approval_mode;type:varchar(16)" json:"approvalMode"`
	Status         int        `gorm:"column:status;type:tinyint;default:0;index" json:"status"`
	Comment        *string    `gorm:"column:comment;type:varchar(512)" json:"comment"`
	FinishDate     *time.Time `gorm:"column:finish_date" json:"finishDate"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessTask) TableName() string { return "wf_process_task" }

// WfProcessRecord 保存流程运行过程中的不可变操作记录。
type WfProcessRecord struct {
	RecordID       string     `gorm:"column:record_id;type:char(36);primaryKey" json:"recordId"`
	NodeInstanceID string     `gorm:"column:node_instance_id;type:char(36);index" json:"nodeInstanceId"`
	InstanceID     string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	TaskID         *string    `gorm:"column:task_id;type:char(36)" json:"taskId"`
	NodeID         *string    `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName       *string    `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	Action         string     `gorm:"column:action;type:varchar(32)" json:"action"`
	OperatorID     *string    `gorm:"column:operator_id;type:char(36)" json:"operatorId"`
	OperatorName   *string    `gorm:"column:operator_name;type:varchar(36)" json:"operatorName"`
	Comment        *string    `gorm:"column:comment;type:varchar(512)" json:"comment"`
	CreateDate     *time.Time `gorm:"column:create_date;index" json:"createDate"`
}

func (WfProcessRecord) TableName() string { return "wf_process_record" }

// WfProcessCopy 保存抄送接收人与已读状态。
type WfProcessCopy struct {
	CopyID         string     `gorm:"column:copy_id;type:char(36);primaryKey" json:"copyId"`
	NodeInstanceID string     `gorm:"column:node_instance_id;type:char(36);index" json:"nodeInstanceId"`
	InstanceID     string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	NodeID         string     `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName       string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	ReceiverID     string     `gorm:"column:receiver_id;type:char(36);index" json:"receiverId"`
	ReceiverName   string     `gorm:"column:receiver_name;type:varchar(36)" json:"receiverName"`
	Status         int        `gorm:"column:status;type:tinyint;default:0;index" json:"status"`
	ReadDate       *time.Time `gorm:"column:read_date" json:"readDate"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessCopy) TableName() string { return "wf_process_copy" }

// StartWorkflowInstanceRequest 是发起流程请求。
// BusinessID 为业务对象主键(如 dev_story.story_id),由业务发起方(如 CreateStory)传入;
// 为空时表示发起纯流程实例(不绑定业务)。BusinessType 不由前端传入,后端从流程定义读取并校验。
type StartWorkflowInstanceRequest struct {
	DefinitionID string                 `json:"definitionId" binding:"required" example:"UUID"` // 流程定义ID
	BusinessID   *string                `json:"businessId" example:"UUID"`                      // 业务对象主键,可空
	Variables    map[string]interface{} `json:"variables"`                                      // 条件判断业务变量
}

// StartStoryWorkflowRequest 为需求发起流程的专用请求。
// businessId 不由前端传入,后端从路径参数 storyId 取,保证业务对象与 URL 一致。
type StartStoryWorkflowRequest struct {
	DefinitionID string                 `json:"definitionId" binding:"required" example:"UUID"` // 已发布的流程定义ID
	Variables    map[string]interface{} `json:"variables"`                                      // 流程变量
}

// WorkflowTaskActionRequest 是审批操作请求。
type WorkflowTaskActionRequest struct {
	Comment   *string                `json:"comment" example:"同意"` // 审批意见
	Variables map[string]interface{} `json:"variables"`            // 当前节点允许编辑的表单字段
}

// WorkflowTaskTransferRequest 是转交待办任务请求。
type WorkflowTaskTransferRequest struct {
	TargetUserID string  `json:"targetUserId" binding:"required" example:"UUID"` // 目标用户ID
	Comment      *string `json:"comment" example:"请协助处理"`                        // 转交说明
}

// WorkflowTaskAddSignRequest 是向当前审批组并行加签请求。
type WorkflowTaskAddSignRequest struct {
	UserIDs []string `json:"userIds" binding:"required"` // 加签用户ID
	Comment *string  `json:"comment" example:"增加财务复核"`   // 加签说明
}

// WorkflowTaskRemoveSignRequest 是从当前审批组减签请求。
type WorkflowTaskRemoveSignRequest struct {
	TaskIDs []string `json:"taskIds" binding:"required"` // 待取消的任务ID
	Comment *string  `json:"comment" example:"无需重复审批"`   // 减签说明
}

// WorkflowTaskReturnRequest 是退回历史审批节点请求，目标为空时退回上一审批节点。
type WorkflowTaskReturnRequest struct {
	TargetNodeID *string `json:"targetNodeId" example:"approve_manager"` // 历史审批节点ID
	Comment      *string `json:"comment" example:"请补充材料"`                // 退回说明
}

// WorkflowReturnTargetResponse 是当前任务允许退回的历史审批节点。
type WorkflowReturnTargetResponse struct {
	NodeID   string `json:"nodeId" example:"approve_manager"` // 节点ID
	NodeName string `json:"nodeName" example:"部门主管审批"`        // 节点名称
}

// WorkflowInstanceResponse 是流程实例列表与详情基础数据。
type WorkflowInstanceResponse struct {
	InstanceID        string                 `json:"instanceId" example:"550e8400-e29b-41d4-a716-446655440000"`   // 流程实例ID
	InstanceNo        string                 `json:"instanceNo" example:"WI000001"`                               // 流程编号
	DefinitionID      string                 `json:"definitionId" example:"550e8400-e29b-41d4-a716-446655440000"` // 流程定义ID
	DefinitionKey     string                 `json:"definitionKey" example:"WF000001"`                            // 流程标识
	DefinitionName    string                 `json:"definitionName" example:"需求审批流程"`                             // 流程名称
	DefinitionVersion int                    `json:"definitionVersion" example:"1"`                               // 流程定义版本
	Title             string                 `json:"title" example:"XX功能需求审批"`                                    // 流程标题
	StarterID         string                 `json:"starterId" example:"550e8400-e29b-41d4-a716-446655440000"`    // 发起人ID
	StarterName       string                 `json:"starterName" example:"张三"`                                    // 发起人姓名
	Status            string                 `json:"status" example:"0"`                                          // 流程状态：0运行中 1已完成 2已拒绝 3已取消
	Variables         map[string]interface{} `json:"variables"`                                                   // 条件判断业务变量
	FormSchema        json.RawMessage        `json:"formSchema" swaggertype:"array,object"`                       // 表单Schema快照
	FormLayout        string                 `json:"formLayout" example:"single"`                                 // 表单布局
	StartDate         *string                `json:"startDate" example:"2026-01-15 09:00:00"`                     // 开始时间
	EndDate           *string                `json:"endDate" example:"2026-01-15 18:00:00"`                       // 结束时间
	CreateDate        *string                `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
}

// WorkflowTaskResponse 是用户待办和已办任务数据。
type WorkflowTaskResponse struct {
	TaskID         string  `json:"taskId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 任务ID
	TaskGroupID    string  `json:"taskGroupId" example:"550e8400-e29b-41d4-a716-446655440000"`    // 任务组ID
	NodeInstanceID string  `json:"nodeInstanceId" example:"550e8400-e29b-41d4-a716-446655440000"` // 节点实例ID
	InstanceID     string  `json:"instanceId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 流程实例ID
	InstanceTitle  string  `json:"instanceTitle" example:"XX功能需求审批"`                              // 流程标题
	NodeID         string  `json:"nodeId" example:"approve_manager"`                              // 节点ID
	NodeName       string  `json:"nodeName" example:"部门主管审批"`                                     // 节点名称
	AssigneeID     string  `json:"assigneeId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 审批人ID
	AssigneeName   string  `json:"assigneeName" example:"李四"`                                     // 审批人姓名
	ApprovalMode   string  `json:"approvalMode" example:"or"`                                     // 审批模式
	Status         string  `json:"status" example:"0"`                                            // 任务状态：0待办 1已审批 2已拒绝 3已取消
	Comment        *string `json:"comment" example:"同意通过"`                                        // 审批意见
	StarterName    string  `json:"starterName" example:"张三"`                                      // 发起人姓名
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"`                      // 创建时间
	FinishDate     *string `json:"finishDate" example:"2026-01-15 18:00:00"`                      // 完成时间
}

// WorkflowCopyResponse 是用户抄送列表数据。
type WorkflowCopyResponse struct {
	CopyID         string  `json:"copyId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 抄送ID
	NodeInstanceID string  `json:"nodeInstanceId" example:"550e8400-e29b-41d4-a716-446655440000"` // 节点实例ID
	InstanceID     string  `json:"instanceId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 流程实例ID
	InstanceTitle  string  `json:"instanceTitle" example:"XX功能需求审批"`                              // 流程标题
	NodeID         string  `json:"nodeId" example:"approve_manager"`                              // 节点ID
	NodeName       string  `json:"nodeName" example:"部门主管审批"`                                     // 节点名称
	ReceiverID     string  `json:"receiverId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 接收人ID
	ReceiverName   string  `json:"receiverName" example:"王五"`                                     // 接收人姓名
	StarterName    string  `json:"starterName" example:"张三"`                                      // 发起人姓名
	Status         string  `json:"status" example:"0"`                                            // 抄送状态：0未读 1已读
	ReadDate       *string `json:"readDate" example:"2026-01-15 15:00:00"`                        // 已读时间
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"`                      // 创建时间
}

// WorkflowRecordResponse 是流程操作记录响应。
type WorkflowRecordResponse struct {
	RecordID       string  `json:"recordId" example:"550e8400-e29b-41d4-a716-446655440000"`       // 记录ID
	NodeInstanceID string  `json:"nodeInstanceId" example:"550e8400-e29b-41d4-a716-446655440000"` // 节点实例ID
	TaskID         *string `json:"taskId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 任务ID
	NodeID         *string `json:"nodeId" example:"approve_manager"`                              // 节点ID
	NodeName       *string `json:"nodeName" example:"部门主管审批"`                                     // 节点名称
	Action         string  `json:"action" example:"approve"`                                      // 操作类型
	OperatorID     *string `json:"operatorId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 操作人ID
	OperatorName   *string `json:"operatorName" example:"张三"`                                     // 操作人姓名
	Comment        *string `json:"comment" example:"同意通过"`                                        // 操作意见
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"`                      // 操作时间
}

// WorkflowNodeActorResponse 是节点预计参与人快照。
type WorkflowNodeActorResponse struct {
	UserID   string `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"` // 用户ID
	UserName string `json:"userName" example:"张三"`                                 // 用户姓名
}

// WorkflowNodeInstanceResponse 聚合单次节点执行及其关联数据。
type WorkflowNodeInstanceResponse struct {
	NodeInstanceID   string                      `json:"nodeInstanceId" example:"550e8400-e29b-41d4-a716-446655440000"` // 节点实例ID
	NodeID           string                      `json:"nodeId" example:"approve_manager"`                              // 节点ID
	NodeName         string                      `json:"nodeName" example:"部门主管审批"`                                     // 节点名称
	NodeType         string                      `json:"nodeType" example:"approval"`                                   // 节点类型
	Sequence         int                         `json:"sequence" example:"1"`                                          // 节点序号
	RouteVersion     int                         `json:"routeVersion" example:"1"`                                      // 路由版本
	Status           string                      `json:"status" example:"1"`                                            // 节点状态：0计划中 1激活 2已完成 3终止 4被取代
	Action           string                      `json:"action" example:"approve"`                                      // 节点操作
	ApprovalMode     *string                     `json:"approvalMode" example:"or"`                                     // 审批模式
	BranchEdgeID     *string                     `json:"branchEdgeId" example:"edge_001"`                               // 分支连线ID
	FieldPermissions map[string]string           `json:"fieldPermissions"`                                              // 字段读写权限映射
	Actors           []WorkflowNodeActorResponse `json:"actors"`                                                        // 预计参与人列表
	Tasks            []WorkflowTaskResponse      `json:"tasks"`                                                         // 任务列表
	Copies           []WorkflowCopyResponse      `json:"copies"`                                                        // 抄送列表
	Records          []WorkflowRecordResponse    `json:"records"`                                                       // 操作记录列表
	StartDate        *string                     `json:"startDate" example:"2026-01-15 09:00:00"`                       // 开始时间
	EndDate          *string                     `json:"endDate" example:"2026-01-15 18:00:00"`                         // 结束时间
	DurationSeconds  *int64                      `json:"durationSeconds" example:"4500"`                                // 耗时秒数
}

// WorkflowInstanceDetailResponse 聚合实例和按流转顺序排列的节点实例。
type WorkflowInstanceDetailResponse struct {
	Instance WorkflowInstanceResponse       `json:"instance"` // 流程实例信息
	Nodes    []WorkflowNodeInstanceResponse `json:"nodes"`    // 节点实例列表
}

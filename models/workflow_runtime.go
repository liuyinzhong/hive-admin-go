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
	BusinessKey       *string    `gorm:"column:business_key;type:varchar(128);index" json:"businessKey"`
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
type StartWorkflowInstanceRequest struct {
	DefinitionID string                 `json:"definitionId" binding:"required" example:"UUID"` // 流程定义ID
	BusinessKey  *string                `json:"businessKey" example:"purchase:UUID"`            // 业务对象唯一标识
	Variables    map[string]interface{} `json:"variables"`                                      // 条件判断业务变量
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
	NodeID   string `json:"nodeId"`
	NodeName string `json:"nodeName"`
}

// WorkflowInstanceResponse 是流程实例列表与详情基础数据。
type WorkflowInstanceResponse struct {
	InstanceID        string                 `json:"instanceId"`
	InstanceNo        string                 `json:"instanceNo"`
	DefinitionID      string                 `json:"definitionId"`
	DefinitionKey     string                 `json:"definitionKey"`
	DefinitionName    string                 `json:"definitionName"`
	DefinitionVersion int                    `json:"definitionVersion"`
	Title             string                 `json:"title"`
	BusinessKey       *string                `json:"businessKey"`
	StarterID         string                 `json:"starterId"`
	StarterName       string                 `json:"starterName"`
	Status            string                 `json:"status"`
	Variables         map[string]interface{} `json:"variables"`
	FormSchema        json.RawMessage        `json:"formSchema" swaggertype:"array,object"`
	FormLayout        string                 `json:"formLayout" example:"single"`
	StartDate         *string                `json:"startDate"`
	EndDate           *string                `json:"endDate"`
	CreateDate        *string                `json:"createDate"`
}

// WorkflowTaskResponse 是用户待办和已办任务数据。
type WorkflowTaskResponse struct {
	TaskID         string  `json:"taskId"`
	TaskGroupID    string  `json:"taskGroupId"`
	NodeInstanceID string  `json:"nodeInstanceId"`
	InstanceID     string  `json:"instanceId"`
	InstanceTitle  string  `json:"instanceTitle"`
	NodeID         string  `json:"nodeId"`
	NodeName       string  `json:"nodeName"`
	AssigneeID     string  `json:"assigneeId"`
	AssigneeName   string  `json:"assigneeName"`
	ApprovalMode   string  `json:"approvalMode"`
	Status         string  `json:"status"`
	Comment        *string `json:"comment"`
	StarterName    string  `json:"starterName"`
	CreateDate     *string `json:"createDate"`
	FinishDate     *string `json:"finishDate"`
}

// WorkflowCopyResponse 是用户抄送列表数据。
type WorkflowCopyResponse struct {
	CopyID         string  `json:"copyId"`
	NodeInstanceID string  `json:"nodeInstanceId"`
	InstanceID     string  `json:"instanceId"`
	InstanceTitle  string  `json:"instanceTitle"`
	NodeID         string  `json:"nodeId"`
	NodeName       string  `json:"nodeName"`
	ReceiverID     string  `json:"receiverId"`
	ReceiverName   string  `json:"receiverName"`
	StarterName    string  `json:"starterName"`
	Status         string  `json:"status"`
	ReadDate       *string `json:"readDate"`
	CreateDate     *string `json:"createDate"`
}

// WorkflowRecordResponse 是流程操作记录响应。
type WorkflowRecordResponse struct {
	RecordID       string  `json:"recordId"`
	NodeInstanceID string  `json:"nodeInstanceId"`
	TaskID         *string `json:"taskId"`
	NodeID         *string `json:"nodeId"`
	NodeName       *string `json:"nodeName"`
	Action         string  `json:"action"`
	OperatorID     *string `json:"operatorId"`
	OperatorName   *string `json:"operatorName"`
	Comment        *string `json:"comment"`
	CreateDate     *string `json:"createDate"`
}

// WorkflowNodeActorResponse 是节点预计参与人快照。
type WorkflowNodeActorResponse struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

// WorkflowNodeInstanceResponse 聚合单次节点执行及其关联数据。
type WorkflowNodeInstanceResponse struct {
	NodeInstanceID   string                      `json:"nodeInstanceId"`
	NodeID           string                      `json:"nodeId"`
	NodeName         string                      `json:"nodeName"`
	NodeType         string                      `json:"nodeType"`
	Sequence         int                         `json:"sequence"`
	RouteVersion     int                         `json:"routeVersion"`
	Status           string                      `json:"status"`
	Action           string                      `json:"action"`
	ApprovalMode     *string                     `json:"approvalMode"`
	BranchEdgeID     *string                     `json:"branchEdgeId"`
	FieldPermissions map[string]string           `json:"fieldPermissions"`
	Actors           []WorkflowNodeActorResponse `json:"actors"`
	Tasks            []WorkflowTaskResponse      `json:"tasks"`
	Copies           []WorkflowCopyResponse      `json:"copies"`
	Records          []WorkflowRecordResponse    `json:"records"`
	StartDate        *string                     `json:"startDate"`
	EndDate          *string                     `json:"endDate"`
	DurationSeconds  *int64                      `json:"durationSeconds" example:"4500"`
}

// WorkflowInstanceDetailResponse 聚合实例和按流转顺序排列的节点实例。
type WorkflowInstanceDetailResponse struct {
	Instance WorkflowInstanceResponse       `json:"instance"`
	Nodes    []WorkflowNodeInstanceResponse `json:"nodes"`
}

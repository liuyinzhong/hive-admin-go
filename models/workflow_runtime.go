package models

import "time"

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
)

// WfProcessInstance 保存一次流程运行实例及其定义快照。
type WfProcessInstance struct {
	InstanceID        string     `gorm:"column:instance_id;type:char(36);primaryKey" json:"instanceId"`
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
	StartDate         *time.Time `gorm:"column:start_date" json:"startDate"`
	EndDate           *time.Time `gorm:"column:end_date" json:"endDate"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag           int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessInstance) TableName() string { return "wf_process_instance" }

// WfProcessTask 保存审批节点为每位审批人生成的任务。
type WfProcessTask struct {
	TaskID       string     `gorm:"column:task_id;type:char(36);primaryKey" json:"taskId"`
	TaskGroupID  string     `gorm:"column:task_group_id;type:char(36);index" json:"taskGroupId"`
	InstanceID   string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	NodeID       string     `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName     string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	AssigneeID   string     `gorm:"column:assignee_id;type:char(36);index" json:"assigneeId"`
	AssigneeName string     `gorm:"column:assignee_name;type:varchar(36)" json:"assigneeName"`
	ApprovalMode string     `gorm:"column:approval_mode;type:varchar(16)" json:"approvalMode"`
	Status       int        `gorm:"column:status;type:tinyint;default:0;index" json:"status"`
	Comment      *string    `gorm:"column:comment;type:varchar(512)" json:"comment"`
	FinishDate   *time.Time `gorm:"column:finish_date" json:"finishDate"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessTask) TableName() string { return "wf_process_task" }

// WfProcessRecord 保存流程运行过程中的不可变操作记录。
type WfProcessRecord struct {
	RecordID     string     `gorm:"column:record_id;type:char(36);primaryKey" json:"recordId"`
	InstanceID   string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	TaskID       *string    `gorm:"column:task_id;type:char(36)" json:"taskId"`
	NodeID       *string    `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName     *string    `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	Action       string     `gorm:"column:action;type:varchar(32)" json:"action"`
	OperatorID   *string    `gorm:"column:operator_id;type:char(36)" json:"operatorId"`
	OperatorName *string    `gorm:"column:operator_name;type:varchar(36)" json:"operatorName"`
	Comment      *string    `gorm:"column:comment;type:varchar(512)" json:"comment"`
	CreateDate   *time.Time `gorm:"column:create_date;index" json:"createDate"`
}

func (WfProcessRecord) TableName() string { return "wf_process_record" }

// WfProcessCopy 保存抄送接收人与已读状态。
type WfProcessCopy struct {
	CopyID       string     `gorm:"column:copy_id;type:char(36);primaryKey" json:"copyId"`
	InstanceID   string     `gorm:"column:instance_id;type:char(36);index" json:"instanceId"`
	NodeID       string     `gorm:"column:node_id;type:varchar(128)" json:"nodeId"`
	NodeName     string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	ReceiverID   string     `gorm:"column:receiver_id;type:char(36);index" json:"receiverId"`
	ReceiverName string     `gorm:"column:receiver_name;type:varchar(36)" json:"receiverName"`
	Status       int        `gorm:"column:status;type:tinyint;default:0;index" json:"status"`
	ReadDate     *time.Time `gorm:"column:read_date" json:"readDate"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (WfProcessCopy) TableName() string { return "wf_process_copy" }

// StartWorkflowInstanceRequest 是发起流程请求。
type StartWorkflowInstanceRequest struct {
	DefinitionID string                 `json:"definitionId" binding:"required" example:"UUID"` // 流程定义ID
	Title        string                 `json:"title" binding:"required" example:"采购申请审批"`      // 实例标题
	BusinessKey  *string                `json:"businessKey" example:"purchase:UUID"`            // 业务对象唯一标识
	Variables    map[string]interface{} `json:"variables"`                                      // 条件判断业务变量
}

// WorkflowTaskActionRequest 是审批操作请求。
type WorkflowTaskActionRequest struct {
	Comment *string `json:"comment" example:"同意"` // 审批意见
}

// WorkflowInstanceResponse 是流程实例列表与详情基础数据。
type WorkflowInstanceResponse struct {
	InstanceID        string                 `json:"instanceId"`
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
	StartDate         *string                `json:"startDate"`
	EndDate           *string                `json:"endDate"`
	CreateDate        *string                `json:"createDate"`
}

// WorkflowTaskResponse 是用户待办和已办任务数据。
type WorkflowTaskResponse struct {
	TaskID        string  `json:"taskId"`
	InstanceID    string  `json:"instanceId"`
	InstanceTitle string  `json:"instanceTitle"`
	NodeID        string  `json:"nodeId"`
	NodeName      string  `json:"nodeName"`
	AssigneeID    string  `json:"assigneeId"`
	AssigneeName  string  `json:"assigneeName"`
	ApprovalMode  string  `json:"approvalMode"`
	Status        string  `json:"status"`
	Comment       *string `json:"comment"`
	StarterName   string  `json:"starterName"`
	CreateDate    *string `json:"createDate"`
	FinishDate    *string `json:"finishDate"`
}

// WorkflowCopyResponse 是用户抄送列表数据。
type WorkflowCopyResponse struct {
	CopyID        string  `json:"copyId"`
	InstanceID    string  `json:"instanceId"`
	InstanceTitle string  `json:"instanceTitle"`
	NodeID        string  `json:"nodeId"`
	NodeName      string  `json:"nodeName"`
	ReceiverID    string  `json:"receiverId"`
	ReceiverName  string  `json:"receiverName"`
	StarterName   string  `json:"starterName"`
	Status        string  `json:"status"`
	ReadDate      *string `json:"readDate"`
	CreateDate    *string `json:"createDate"`
}

// WorkflowRecordResponse 是流程操作记录响应。
type WorkflowRecordResponse struct {
	RecordID     string  `json:"recordId"`
	TaskID       *string `json:"taskId"`
	NodeID       *string `json:"nodeId"`
	NodeName     *string `json:"nodeName"`
	Action       string  `json:"action"`
	OperatorID   *string `json:"operatorId"`
	OperatorName *string `json:"operatorName"`
	Comment      *string `json:"comment"`
	CreateDate   *string `json:"createDate"`
}

// WorkflowInstanceDetailResponse 聚合实例、任务、记录和抄送信息。
type WorkflowInstanceDetailResponse struct {
	Instance WorkflowInstanceResponse `json:"instance"`
	Tasks    []WorkflowTaskResponse   `json:"tasks"`
	Records  []WorkflowRecordResponse `json:"records"`
	Copies   []WorkflowCopyResponse   `json:"copies"`
}

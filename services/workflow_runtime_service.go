package services

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const workflowMaxAutomaticSteps = 100

type workflowGraph struct {
	Nodes []workflowNode `json:"nodes"`
	Edges []workflowEdge `json:"edges"`
}

type workflowNode struct {
	ID         string                 `json:"id"`
	Text       json.RawMessage        `json:"text"`
	Properties workflowNodeProperties `json:"properties"`
}

type workflowNodeProperties struct {
	NodeType     string   `json:"nodeType"`
	AssigneeType string   `json:"assigneeType"`
	AssigneeIDs  []string `json:"assigneeIds"`
	ApprovalMode string   `json:"approvalMode"`
	CopyType     string   `json:"copyType"`
	CopyIDs      []string `json:"copyIds"`
	BranchMode   string   `json:"branchMode"`
}

type workflowEdge struct {
	ID           string                 `json:"id"`
	SourceNodeID string                 `json:"sourceNodeId"`
	TargetNodeID string                 `json:"targetNodeId"`
	Properties   workflowEdgeProperties `json:"properties"`
}

type workflowEdgeProperties struct {
	IsDefaultBranch     bool                    `json:"isDefaultBranch"`
	Priority            int                     `json:"priority"`
	ConditionLogic      string                  `json:"conditionLogic"`
	ConditionRules      []workflowConditionRule `json:"conditionRules"`
	ConditionExpression string                  `json:"conditionExpression"`
}

type workflowConditionRule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type workflowExecutionContext struct {
	graph     *workflowGraph
	instance  *models.WfProcessInstance
	variables map[string]interface{}
	steps     int
}

// StartWorkflowInstance 创建实例快照并推进到第一个人工节点。
func StartWorkflowInstance(req *models.StartWorkflowInstanceRequest, starterID string) (*models.WorkflowInstanceResponse, error) {
	var response *models.WorkflowInstanceResponse
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var definition models.WfProcessDefinition
		if err := tx.Where("definition_id = ? AND status = 1 AND del_flag = 0", req.DefinitionID).
			First(&definition).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("流程定义不存在或未发布")
			}
			return err
		}
		if definition.FlowData == nil {
			return fmt.Errorf("流程定义没有画布数据")
		}

		graph, err := parseWorkflowGraph(*definition.FlowData)
		if err != nil {
			return err
		}
		starter, err := getActiveWorkflowUser(tx, starterID)
		if err != nil {
			return fmt.Errorf("流程发起人不存在或已停用")
		}
		variables := req.Variables
		if variables == nil {
			variables = make(map[string]interface{})
		}
		variablesJSON, err := json.Marshal(variables)
		if err != nil {
			return fmt.Errorf("流程变量无法序列化")
		}

		now := time.Now()
		instance := models.WfProcessInstance{
			InstanceID:        utils.GenerateUUID(),
			DefinitionID:      definition.DefinitionID,
			DefinitionKey:     definition.DefinitionKey,
			DefinitionName:    definition.DefinitionName,
			DefinitionVersion: definition.Version,
			Title:             strings.TrimSpace(req.Title),
			BusinessKey:       normalizeOptionalString(req.BusinessKey),
			StarterID:         starter.UserID,
			StarterName:       workflowUserName(starter),
			Status:            models.WorkflowInstanceStatusRunning,
			Variables:         string(variablesJSON),
			FlowSnapshot:      *definition.FlowData,
			StartDate:         &now,
			CreateDate:        &now,
			UpdateDate:        &now,
			DelFlag:           0,
		}
		if instance.Title == "" {
			return fmt.Errorf("流程标题不能为空")
		}
		if err := tx.Create(&instance).Error; err != nil {
			return err
		}
		if err := createWorkflowRecord(tx, &instance, nil, nil, "start", &starter.UserID, stringPtr(instance.StarterName), nil); err != nil {
			return err
		}

		startNode := findWorkflowNodeByType(graph, "start")
		if startNode == nil {
			return fmt.Errorf("流程缺少开始节点")
		}
		context := &workflowExecutionContext{graph: graph, instance: &instance, variables: variables}
		if err := advanceWorkflowFromNode(tx, context, startNode.ID); err != nil {
			return err
		}

		result, err := buildWorkflowInstanceResponse(instance)
		if err != nil {
			return err
		}
		response = &result
		return nil
	})
	return response, err
}

// GetWorkflowInstances 分页获取当前用户发起的流程实例。
func GetWorkflowInstances(page, pageSize int, userID string, statuses []int) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.WfProcessInstance{}).
		Where("starter_id = ? AND del_flag = 0", userID)
	if len(statuses) > 0 {
		db = db.Where("status IN ?", statuses)
	}
	page, pageSize = normalizeWorkflowPagination(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.WfProcessInstance
	if err := db.Order("create_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	responses := make([]models.WorkflowInstanceResponse, 0, len(items))
	for _, item := range items {
		response, err := buildWorkflowInstanceResponse(item)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

// GetWorkflowTasks 分页获取当前用户的待办或已办任务。
func GetWorkflowTasks(page, pageSize int, userID string, statuses []int) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.WfProcessTask{}).
		Where("assignee_id = ? AND del_flag = 0", userID)
	if len(statuses) > 0 {
		db = db.Where("status IN ?", statuses)
	}
	page, pageSize = normalizeWorkflowPagination(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.WfProcessTask
	if err := db.Order("create_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	instances, err := loadWorkflowInstancesByTasks(items)
	if err != nil {
		return nil, err
	}
	responses := make([]models.WorkflowTaskResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildWorkflowTaskResponse(item, instances[item.InstanceID]))
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

// GetWorkflowCopies 分页获取当前用户收到的抄送记录。
func GetWorkflowCopies(page, pageSize int, userID string, statuses []int) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.WfProcessCopy{}).
		Where("receiver_id = ? AND del_flag = 0", userID)
	if len(statuses) > 0 {
		db = db.Where("status IN ?", statuses)
	}
	page, pageSize = normalizeWorkflowPagination(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.WfProcessCopy
	if err := db.Order("create_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	instances, err := loadWorkflowInstancesByCopies(items)
	if err != nil {
		return nil, err
	}
	responses := make([]models.WorkflowCopyResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildWorkflowCopyResponse(item, instances[item.InstanceID]))
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

// GetWorkflowInstanceDetail 获取参与者可见的流程实例详情。
func GetWorkflowInstanceDetail(instanceID, userID string) (*models.WorkflowInstanceDetailResponse, error) {
	var instance models.WfProcessInstance
	if err := database.DB.Where("instance_id = ? AND del_flag = 0", instanceID).First(&instance).Error; err != nil {
		return nil, fmt.Errorf("流程实例不存在")
	}
	if instance.StarterID != userID {
		var participation int64
		if err := database.DB.Model(&models.WfProcessTask{}).
			Where("instance_id = ? AND assignee_id = ? AND del_flag = 0", instanceID, userID).
			Count(&participation).Error; err != nil {
			return nil, err
		}
		if participation == 0 {
			if err := database.DB.Model(&models.WfProcessCopy{}).
				Where("instance_id = ? AND receiver_id = ? AND del_flag = 0", instanceID, userID).
				Count(&participation).Error; err != nil {
				return nil, err
			}
		}
		if participation == 0 {
			return nil, fmt.Errorf("无权查看该流程实例")
		}
	}

	var tasks []models.WfProcessTask
	var records []models.WfProcessRecord
	var copies []models.WfProcessCopy
	if err := database.DB.Where("instance_id = ? AND del_flag = 0", instanceID).Order("create_date ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Where("instance_id = ?", instanceID).Order("create_date ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Where("instance_id = ? AND del_flag = 0", instanceID).Order("create_date ASC").Find(&copies).Error; err != nil {
		return nil, err
	}

	instanceResponse, err := buildWorkflowInstanceResponse(instance)
	if err != nil {
		return nil, err
	}
	detail := &models.WorkflowInstanceDetailResponse{Instance: instanceResponse}
	detail.Tasks = make([]models.WorkflowTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		detail.Tasks = append(detail.Tasks, buildWorkflowTaskResponse(task, instance))
	}
	detail.Records = make([]models.WorkflowRecordResponse, 0, len(records))
	for _, record := range records {
		detail.Records = append(detail.Records, models.WorkflowRecordResponse{
			RecordID: record.RecordID, TaskID: record.TaskID, NodeID: record.NodeID, NodeName: record.NodeName,
			Action: record.Action, OperatorID: record.OperatorID, OperatorName: record.OperatorName,
			Comment: record.Comment, CreateDate: models.TimeToStringPtr(record.CreateDate),
		})
	}
	detail.Copies = make([]models.WorkflowCopyResponse, 0, len(copies))
	for _, copyItem := range copies {
		detail.Copies = append(detail.Copies, buildWorkflowCopyResponse(copyItem, instance))
	}
	return detail, nil
}

// ApproveWorkflowTask 审批通过任务，并按或签/会签规则推进流程。
func ApproveWorkflowTask(taskID, userID string, req *models.WorkflowTaskActionRequest) error {
	return handleWorkflowTask(taskID, userID, req.Comment, true)
}

// RejectWorkflowTask 驳回任务并终止当前流程实例。
func RejectWorkflowTask(taskID, userID string, req *models.WorkflowTaskActionRequest) error {
	return handleWorkflowTask(taskID, userID, req.Comment, false)
}

// CancelWorkflowInstance 允许发起人撤销仍在运行的实例。
func CancelWorkflowInstance(instanceID, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var instance models.WfProcessInstance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("instance_id = ? AND del_flag = 0", instanceID).First(&instance).Error; err != nil {
			return fmt.Errorf("流程实例不存在")
		}
		if instance.StarterID != userID {
			return fmt.Errorf("只有发起人可以撤销流程")
		}
		if instance.Status != models.WorkflowInstanceStatusRunning {
			return fmt.Errorf("只有运行中的流程可以撤销")
		}
		operator, err := getActiveWorkflowUser(tx, userID)
		if err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&instance).Updates(map[string]interface{}{
			"status": models.WorkflowInstanceStatusCanceled, "end_date": now, "update_date": now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.WfProcessTask{}).
			Where("instance_id = ? AND status = ? AND del_flag = 0", instanceID, models.WorkflowTaskStatusPending).
			Updates(map[string]interface{}{"status": models.WorkflowTaskStatusCanceled, "finish_date": now, "update_date": now}).Error; err != nil {
			return err
		}
		return createWorkflowRecord(tx, &instance, nil, nil, "cancel", &userID, stringPtr(workflowUserName(operator)), nil)
	})
}

// ReadWorkflowCopy 将当前用户收到的抄送记录标记为已读。
func ReadWorkflowCopy(copyID, userID string) error {
	now := time.Now()
	result := database.DB.Model(&models.WfProcessCopy{}).
		Where("copy_id = ? AND receiver_id = ? AND del_flag = 0", copyID, userID).
		Updates(map[string]interface{}{"status": models.WorkflowCopyStatusRead, "read_date": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("抄送记录不存在")
	}
	return nil
}

// handleWorkflowTask 在事务内处理审批并推进或终止实例。
func handleWorkflowTask(taskID, userID string, comment *string, approved bool) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var task models.WfProcessTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("task_id = ? AND del_flag = 0", taskID).First(&task).Error; err != nil {
			return fmt.Errorf("审批任务不存在")
		}
		if task.AssigneeID != userID {
			return fmt.Errorf("无权处理该审批任务")
		}
		if task.Status != models.WorkflowTaskStatusPending {
			return fmt.Errorf("审批任务已处理")
		}

		var instance models.WfProcessInstance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("instance_id = ? AND del_flag = 0", task.InstanceID).First(&instance).Error; err != nil {
			return fmt.Errorf("流程实例不存在")
		}
		if instance.Status != models.WorkflowInstanceStatusRunning {
			return fmt.Errorf("流程实例已结束")
		}
		operator, err := getActiveWorkflowUser(tx, userID)
		if err != nil {
			return err
		}

		now := time.Now()
		taskStatus := models.WorkflowTaskStatusRejected
		action := "reject"
		if approved {
			taskStatus = models.WorkflowTaskStatusApproved
			action = "approve"
		}
		if err := tx.Model(&task).Updates(map[string]interface{}{
			"status": taskStatus, "comment": normalizeOptionalString(comment), "finish_date": now, "update_date": now,
		}).Error; err != nil {
			return err
		}
		if err := createWorkflowRecord(tx, &instance, &task, nil, action, &userID, stringPtr(workflowUserName(operator)), normalizeOptionalString(comment)); err != nil {
			return err
		}

		if !approved {
			if err := cancelWorkflowTaskGroup(tx, task.TaskGroupID, task.TaskID, now); err != nil {
				return err
			}
			return finishWorkflowInstance(tx, &instance, models.WorkflowInstanceStatusRejected, "rejected")
		}

		if task.ApprovalMode == "any" {
			if err := cancelWorkflowTaskGroup(tx, task.TaskGroupID, task.TaskID, now); err != nil {
				return err
			}
		} else {
			var pendingCount int64
			if err := tx.Model(&models.WfProcessTask{}).
				Where("task_group_id = ? AND status = ? AND del_flag = 0", task.TaskGroupID, models.WorkflowTaskStatusPending).
				Count(&pendingCount).Error; err != nil {
				return err
			}
			if pendingCount > 0 {
				return nil
			}
		}

		graph, err := parseWorkflowGraph(instance.FlowSnapshot)
		if err != nil {
			return err
		}
		variables := make(map[string]interface{})
		if err := json.Unmarshal([]byte(instance.Variables), &variables); err != nil {
			return fmt.Errorf("流程变量解析失败")
		}
		context := &workflowExecutionContext{graph: graph, instance: &instance, variables: variables}
		return advanceWorkflowFromNode(tx, context, task.NodeID)
	})
}

// advanceWorkflowFromNode 从指定节点沿唯一出线继续执行。
func advanceWorkflowFromNode(tx *gorm.DB, context *workflowExecutionContext, nodeID string) error {
	context.steps++
	if context.steps > workflowMaxAutomaticSteps {
		return fmt.Errorf("流程自动流转超过安全上限，请检查是否存在循环")
	}
	outgoing := workflowOutgoingEdges(context.graph, nodeID)
	if len(outgoing) != 1 {
		return fmt.Errorf("节点 %s 必须且只能有一条普通出线", nodeID)
	}
	return enterWorkflowNode(tx, context, outgoing[0].TargetNodeID)
}

// enterWorkflowNode 根据节点类型执行自动节点或创建人工任务。
func enterWorkflowNode(tx *gorm.DB, context *workflowExecutionContext, nodeID string) error {
	node := findWorkflowNode(context.graph, nodeID)
	if node == nil {
		return fmt.Errorf("流程节点 %s 不存在", nodeID)
	}
	nodeName := workflowNodeName(node)
	switch node.Properties.NodeType {
	case "end":
		if err := createWorkflowRecord(tx, context.instance, nil, node, "complete", nil, nil, nil); err != nil {
			return err
		}
		return finishWorkflowInstance(tx, context.instance, models.WorkflowInstanceStatusCompleted, "completed")
	case "approve":
		return createWorkflowApprovalTasks(tx, context, node)
	case "copy":
		if err := createWorkflowCopies(tx, context, node); err != nil {
			return err
		}
		if err := createWorkflowRecord(tx, context.instance, nil, node, "copy", nil, nil, nil); err != nil {
			return err
		}
		return advanceWorkflowFromNode(tx, context, node.ID)
	case "condition":
		edge, err := selectWorkflowConditionEdge(context.graph, node.ID, context.variables)
		if err != nil {
			return fmt.Errorf("条件节点 %s：%w", nodeName, err)
		}
		if err := createWorkflowRecord(tx, context.instance, nil, node, "branch", nil, nil, stringPtr(edge.ID)); err != nil {
			return err
		}
		context.steps++
		if context.steps > workflowMaxAutomaticSteps {
			return fmt.Errorf("流程自动流转超过安全上限，请检查是否存在循环")
		}
		return enterWorkflowNode(tx, context, edge.TargetNodeID)
	case "start":
		return advanceWorkflowFromNode(tx, context, node.ID)
	default:
		return fmt.Errorf("不支持的流程节点类型：%s", node.Properties.NodeType)
	}
}

// createWorkflowApprovalTasks 解析审批人并按同一任务组创建审批任务。
func createWorkflowApprovalTasks(tx *gorm.DB, context *workflowExecutionContext, node *workflowNode) error {
	users, err := resolveWorkflowActors(tx, context.instance, node.Properties.AssigneeType, node.Properties.AssigneeIDs)
	if err != nil {
		return fmt.Errorf("审批节点 %s：%w", workflowNodeName(node), err)
	}
	approvalMode := node.Properties.ApprovalMode
	if approvalMode == "" {
		approvalMode = "any"
	}
	if approvalMode != "any" && approvalMode != "all" {
		return fmt.Errorf("审批方式必须是 any 或 all")
	}
	taskGroupID := utils.GenerateUUID()
	now := time.Now()
	for _, user := range users {
		task := models.WfProcessTask{
			TaskID: utils.GenerateUUID(), TaskGroupID: taskGroupID, InstanceID: context.instance.InstanceID,
			NodeID: node.ID, NodeName: workflowNodeName(node), AssigneeID: user.UserID,
			AssigneeName: workflowUserName(user), ApprovalMode: approvalMode,
			Status: models.WorkflowTaskStatusPending, CreateDate: &now, UpdateDate: &now,
		}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
	}
	return createWorkflowRecord(tx, context.instance, nil, node, "pending", nil, nil, nil)
}

// createWorkflowCopies 解析抄送人并生成去重后的抄送记录。
func createWorkflowCopies(tx *gorm.DB, context *workflowExecutionContext, node *workflowNode) error {
	users, err := resolveWorkflowActors(tx, context.instance, node.Properties.CopyType, node.Properties.CopyIDs)
	if err != nil {
		return fmt.Errorf("抄送节点 %s：%w", workflowNodeName(node), err)
	}
	now := time.Now()
	for _, user := range users {
		copyItem := models.WfProcessCopy{
			CopyID: utils.GenerateUUID(), InstanceID: context.instance.InstanceID,
			NodeID: node.ID, NodeName: workflowNodeName(node), ReceiverID: user.UserID,
			ReceiverName: workflowUserName(user), Status: models.WorkflowCopyStatusUnread,
			CreateDate: &now,
		}
		if err := tx.Create(&copyItem).Error; err != nil {
			return err
		}
	}
	return nil
}

// resolveWorkflowActors 按用户、角色或发起人主管解析启用用户。
func resolveWorkflowActors(tx *gorm.DB, instance *models.WfProcessInstance, actorType string, actorIDs []string) ([]models.SysUser, error) {
	userIDs := make([]string, 0)
	switch actorType {
	case "user":
		userIDs = append(userIDs, actorIDs...)
	case "role":
		if len(actorIDs) == 0 {
			return nil, fmt.Errorf("未配置角色")
		}
		if err := tx.Model(&models.SysUserRole{}).
			Where("role_id IN ? AND del_flag = 0", actorIDs).Distinct().Pluck("user_id", &userIDs).Error; err != nil {
			return nil, err
		}
	case "leader":
		var starter models.SysUser
		if err := tx.Where("user_id = ? AND del_flag = 0", instance.StarterID).First(&starter).Error; err != nil {
			return nil, err
		}
		if starter.LeaderUserID == nil || *starter.LeaderUserID == "" {
			return nil, fmt.Errorf("发起人未配置直属主管")
		}
		userIDs = append(userIDs, *starter.LeaderUserID)
	default:
		return nil, fmt.Errorf("不支持的人员类型：%s", actorType)
	}
	userIDs = uniqueStrings(userIDs)
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("没有解析到处理人")
	}
	var users []models.SysUser
	if err := tx.Where("user_id IN ? AND status = 1 AND del_flag = 0", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) != len(userIDs) {
		return nil, fmt.Errorf("部分处理人不存在或已停用")
	}
	return users, nil
}

// selectWorkflowConditionEdge 按优先级选中第一条匹配出线，否则返回默认分支。
func selectWorkflowConditionEdge(graph *workflowGraph, nodeID string, variables map[string]interface{}) (*workflowEdge, error) {
	outgoing := workflowOutgoingEdges(graph, nodeID)
	var defaultEdge *workflowEdge
	conditionEdges := make([]workflowEdge, 0)
	for index := range outgoing {
		edge := outgoing[index]
		if edge.Properties.IsDefaultBranch {
			copyEdge := edge
			defaultEdge = &copyEdge
			continue
		}
		conditionEdges = append(conditionEdges, edge)
	}
	sort.SliceStable(conditionEdges, func(i, j int) bool {
		return conditionEdges[i].Properties.Priority < conditionEdges[j].Properties.Priority
	})
	for index := range conditionEdges {
		matched, err := evaluateWorkflowEdge(conditionEdges[index], variables)
		if err != nil {
			return nil, err
		}
		if matched {
			return &conditionEdges[index], nil
		}
	}
	if defaultEdge == nil {
		return nil, fmt.Errorf("没有条件命中且未配置默认分支")
	}
	return defaultEdge, nil
}

// evaluateWorkflowEdge 计算一条出线的全部结构化条件。
func evaluateWorkflowEdge(edge workflowEdge, variables map[string]interface{}) (bool, error) {
	rules := edge.Properties.ConditionRules
	if len(rules) == 0 {
		return false, fmt.Errorf("分支 %s 没有结构化条件规则", edge.ID)
	}
	logic := edge.Properties.ConditionLogic
	if logic == "or" {
		for _, rule := range rules {
			matched, err := evaluateWorkflowRule(rule, variables)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	for _, rule := range rules {
		matched, err := evaluateWorkflowRule(rule, variables)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

// evaluateWorkflowRule 按字段、操作符和值计算单条规则。
func evaluateWorkflowRule(rule workflowConditionRule, variables map[string]interface{}) (bool, error) {
	actual, exists := workflowVariableValue(variables, rule.Field)
	switch rule.Operator {
	case "empty":
		return !exists || workflowValueEmpty(actual), nil
	case "notEmpty":
		return exists && !workflowValueEmpty(actual), nil
	}
	if !exists {
		return false, nil
	}
	switch rule.Operator {
	case "equal", "notEqual":
		matched := workflowValuesEqual(actual, rule.Value)
		if rule.Operator == "notEqual" {
			matched = !matched
		}
		return matched, nil
	case "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual":
		actualNumber, ok := workflowNumber(actual)
		if !ok {
			return false, fmt.Errorf("字段 %s 不是数字", rule.Field)
		}
		expectedNumber, err := strconv.ParseFloat(strings.TrimSpace(rule.Value), 64)
		if err != nil {
			return false, fmt.Errorf("字段 %s 的比较值不是数字", rule.Field)
		}
		switch rule.Operator {
		case "greaterThan":
			return actualNumber > expectedNumber, nil
		case "greaterThanOrEqual":
			return actualNumber >= expectedNumber, nil
		case "lessThan":
			return actualNumber < expectedNumber, nil
		default:
			return actualNumber <= expectedNumber, nil
		}
	case "contains", "notContains":
		matched := workflowValueContains(actual, rule.Value)
		if rule.Operator == "notContains" {
			matched = !matched
		}
		return matched, nil
	default:
		return false, fmt.Errorf("不支持的条件操作符：%s", rule.Operator)
	}
}

// parseWorkflowGraph 解析并验证可执行的 LogicFlow 画布。
func parseWorkflowGraph(flowData string) (*workflowGraph, error) {
	var graph workflowGraph
	if err := json.Unmarshal([]byte(flowData), &graph); err != nil {
		return nil, fmt.Errorf("流程画布数据必须是合法JSON")
	}
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("流程画布不能为空")
	}
	nodeMap := make(map[string]workflowNode)
	startCount := 0
	endCount := 0
	for _, node := range graph.Nodes {
		if node.ID == "" || node.Properties.NodeType == "" {
			return nil, fmt.Errorf("流程存在无效节点")
		}
		nodeMap[node.ID] = node
		if node.Properties.NodeType == "start" {
			startCount++
		}
		if node.Properties.NodeType == "end" {
			endCount++
		}
	}
	if startCount != 1 || endCount == 0 {
		return nil, fmt.Errorf("流程必须包含一个开始节点和至少一个结束节点")
	}
	for _, edge := range graph.Edges {
		if _, exists := nodeMap[edge.SourceNodeID]; !exists {
			return nil, fmt.Errorf("连线 %s 的源节点不存在", edge.ID)
		}
		if _, exists := nodeMap[edge.TargetNodeID]; !exists {
			return nil, fmt.Errorf("连线 %s 的目标节点不存在", edge.ID)
		}
	}
	for index := range graph.Nodes {
		node := &graph.Nodes[index]
		outgoing := workflowOutgoingEdges(&graph, node.ID)
		switch node.Properties.NodeType {
		case "end":
			if len(outgoing) != 0 {
				return nil, fmt.Errorf("结束节点不能有出线")
			}
		case "condition":
			if err := validateWorkflowConditionEdges(node, outgoing); err != nil {
				return nil, err
			}
		case "approve":
			if node.Properties.AssigneeType == "" {
				return nil, fmt.Errorf("审批节点 %s 未配置审批人", workflowNodeName(node))
			}
			if node.Properties.ApprovalMode == "" {
				node.Properties.ApprovalMode = "any"
			}
			if node.Properties.ApprovalMode != "any" && node.Properties.ApprovalMode != "all" {
				return nil, fmt.Errorf("审批节点 %s 的审批方式无效", workflowNodeName(node))
			}
			if len(outgoing) != 1 {
				return nil, fmt.Errorf("审批节点 %s 必须且只能有一条出线", workflowNodeName(node))
			}
		default:
			if len(outgoing) != 1 {
				return nil, fmt.Errorf("节点 %s 必须且只能有一条出线", workflowNodeName(node))
			}
		}
	}
	return &graph, nil
}

// validateWorkflowConditionEdges 校验条件节点的默认分支、优先级和规则。
func validateWorkflowConditionEdges(node *workflowNode, edges []workflowEdge) error {
	if len(edges) < 2 {
		return fmt.Errorf("条件节点 %s 至少需要两条出线", workflowNodeName(node))
	}
	defaultCount := 0
	priorities := make(map[int]bool)
	for _, edge := range edges {
		if edge.Properties.IsDefaultBranch {
			defaultCount++
			continue
		}
		if edge.Properties.Priority <= 0 || priorities[edge.Properties.Priority] {
			return fmt.Errorf("条件节点 %s 的分支优先级无效或重复", workflowNodeName(node))
		}
		priorities[edge.Properties.Priority] = true
		if len(edge.Properties.ConditionRules) == 0 {
			return fmt.Errorf("条件节点 %s 存在未配置结构化规则的分支", workflowNodeName(node))
		}
		if edge.Properties.ConditionLogic != "" && edge.Properties.ConditionLogic != "and" && edge.Properties.ConditionLogic != "or" {
			return fmt.Errorf("条件节点 %s 存在无效的规则关系", workflowNodeName(node))
		}
		for _, rule := range edge.Properties.ConditionRules {
			if strings.TrimSpace(rule.Field) == "" {
				return fmt.Errorf("条件节点 %s 存在未配置字段的规则", workflowNodeName(node))
			}
			if !workflowConditionOperatorSupported(rule.Operator) {
				return fmt.Errorf("条件节点 %s 存在不支持的条件操作符", workflowNodeName(node))
			}
			if rule.Operator != "empty" && rule.Operator != "notEmpty" && strings.TrimSpace(rule.Value) == "" {
				return fmt.Errorf("条件节点 %s 存在未配置比较值的规则", workflowNodeName(node))
			}
		}
	}
	if defaultCount != 1 {
		return fmt.Errorf("条件节点 %s 必须且只能有一条默认分支", workflowNodeName(node))
	}
	return nil
}

// workflowConditionOperatorSupported 判断运行引擎是否支持指定条件操作符。
func workflowConditionOperatorSupported(operator string) bool {
	switch operator {
	case "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "contains", "notContains", "empty", "notEmpty":
		return true
	default:
		return false
	}
}

// cancelWorkflowTaskGroup 取消同一审批组内尚未处理的其他任务。
func cancelWorkflowTaskGroup(tx *gorm.DB, taskGroupID, excludeTaskID string, now time.Time) error {
	return tx.Model(&models.WfProcessTask{}).
		Where("task_group_id = ? AND task_id <> ? AND status = ? AND del_flag = 0", taskGroupID, excludeTaskID, models.WorkflowTaskStatusPending).
		Updates(map[string]interface{}{"status": models.WorkflowTaskStatusCanceled, "finish_date": now, "update_date": now}).Error
}

// finishWorkflowInstance 更新实例终态并记录实例结束事件。
func finishWorkflowInstance(tx *gorm.DB, instance *models.WfProcessInstance, status int, action string) error {
	now := time.Now()
	if err := tx.Model(instance).Updates(map[string]interface{}{
		"status": status, "end_date": now, "update_date": now,
	}).Error; err != nil {
		return err
	}
	instance.Status = status
	instance.EndDate = &now
	return createWorkflowRecord(tx, instance, nil, nil, action, nil, nil, nil)
}

// createWorkflowRecord 创建一条流程审计记录。
func createWorkflowRecord(tx *gorm.DB, instance *models.WfProcessInstance, task *models.WfProcessTask, node *workflowNode, action string, operatorID, operatorName, comment *string) error {
	now := time.Now()
	record := models.WfProcessRecord{
		RecordID: utils.GenerateUUID(), InstanceID: instance.InstanceID, Action: action,
		OperatorID: operatorID, OperatorName: operatorName, Comment: comment, CreateDate: &now,
	}
	if task != nil {
		record.TaskID = &task.TaskID
		record.NodeID = &task.NodeID
		record.NodeName = &task.NodeName
	}
	if node != nil {
		record.NodeID = &node.ID
		nodeName := workflowNodeName(node)
		record.NodeName = &nodeName
	}
	return tx.Create(&record).Error
}

// getActiveWorkflowUser 查询可参与流程的启用用户。
func getActiveWorkflowUser(tx *gorm.DB, userID string) (models.SysUser, error) {
	var user models.SysUser
	err := tx.Where("user_id = ? AND status = 1 AND del_flag = 0", userID).First(&user).Error
	return user, err
}

// workflowOutgoingEdges 获取指定节点全部出线。
func workflowOutgoingEdges(graph *workflowGraph, nodeID string) []workflowEdge {
	edges := make([]workflowEdge, 0)
	for _, edge := range graph.Edges {
		if edge.SourceNodeID == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// findWorkflowNode 按节点ID查询节点。
func findWorkflowNode(graph *workflowGraph, nodeID string) *workflowNode {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == nodeID {
			return &graph.Nodes[index]
		}
	}
	return nil
}

// findWorkflowNodeByType 按节点类型查询第一个节点。
func findWorkflowNodeByType(graph *workflowGraph, nodeType string) *workflowNode {
	for index := range graph.Nodes {
		if graph.Nodes[index].Properties.NodeType == nodeType {
			return &graph.Nodes[index]
		}
	}
	return nil
}

// workflowNodeName 兼容 LogicFlow 字符串和对象文本格式。
func workflowNodeName(node *workflowNode) string {
	if len(node.Text) == 0 {
		return node.ID
	}
	var text string
	if json.Unmarshal(node.Text, &text) == nil && text != "" {
		return text
	}
	var textObject struct {
		Value string `json:"value"`
	}
	if json.Unmarshal(node.Text, &textObject) == nil && textObject.Value != "" {
		return textObject.Value
	}
	return node.ID
}

// workflowVariableValue 支持使用点号读取嵌套流程变量。
func workflowVariableValue(variables map[string]interface{}, field string) (interface{}, bool) {
	parts := strings.Split(strings.TrimSpace(field), ".")
	var current interface{} = variables
	for _, part := range parts {
		object, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// workflowValuesEqual 按实际变量类型比较规则值。
func workflowValuesEqual(actual interface{}, expected string) bool {
	if number, ok := workflowNumber(actual); ok {
		expectedNumber, err := strconv.ParseFloat(strings.TrimSpace(expected), 64)
		return err == nil && number == expectedNumber
	}
	if value, ok := actual.(bool); ok {
		expectedBool, err := strconv.ParseBool(strings.TrimSpace(expected))
		return err == nil && value == expectedBool
	}
	return fmt.Sprint(actual) == expected
}

// workflowNumber 将JSON数字和常用整数类型转换为float64。
func workflowNumber(value interface{}) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

// workflowValueEmpty 判断空值、空字符串和空集合。
func workflowValueEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len() == 0
	default:
		return false
	}
}

// workflowValueContains 判断字符串或集合是否包含规则值。
func workflowValueContains(actual interface{}, expected string) bool {
	if value, ok := actual.(string); ok {
		return strings.Contains(value, expected)
	}
	reflected := reflect.ValueOf(actual)
	if reflected.IsValid() && (reflected.Kind() == reflect.Array || reflected.Kind() == reflect.Slice) {
		for index := 0; index < reflected.Len(); index++ {
			if workflowValuesEqual(reflected.Index(index).Interface(), expected) {
				return true
			}
		}
	}
	return false
}

// buildWorkflowInstanceResponse 转换实例响应并解析变量JSON。
func buildWorkflowInstanceResponse(instance models.WfProcessInstance) (models.WorkflowInstanceResponse, error) {
	variables := make(map[string]interface{})
	if err := json.Unmarshal([]byte(instance.Variables), &variables); err != nil {
		return models.WorkflowInstanceResponse{}, fmt.Errorf("流程实例 %s 的变量数据损坏", instance.InstanceID)
	}
	return models.WorkflowInstanceResponse{
		InstanceID: instance.InstanceID, DefinitionID: instance.DefinitionID,
		DefinitionKey: instance.DefinitionKey, DefinitionName: instance.DefinitionName,
		DefinitionVersion: instance.DefinitionVersion, Title: instance.Title,
		BusinessKey: instance.BusinessKey, StarterID: instance.StarterID,
		StarterName: instance.StarterName, Status: strconv.Itoa(instance.Status),
		Variables: variables, StartDate: models.TimeToStringPtr(instance.StartDate),
		EndDate: models.TimeToStringPtr(instance.EndDate), CreateDate: models.TimeToStringPtr(instance.CreateDate),
	}, nil
}

// buildWorkflowTaskResponse 组合任务及实例摘要。
func buildWorkflowTaskResponse(task models.WfProcessTask, instance models.WfProcessInstance) models.WorkflowTaskResponse {
	return models.WorkflowTaskResponse{
		TaskID: task.TaskID, InstanceID: task.InstanceID, InstanceTitle: instance.Title,
		NodeID: task.NodeID, NodeName: task.NodeName, AssigneeID: task.AssigneeID,
		AssigneeName: task.AssigneeName, ApprovalMode: task.ApprovalMode,
		Status: strconv.Itoa(task.Status), Comment: task.Comment, StarterName: instance.StarterName,
		CreateDate: models.TimeToStringPtr(task.CreateDate), FinishDate: models.TimeToStringPtr(task.FinishDate),
	}
}

// buildWorkflowCopyResponse 组合抄送及实例摘要。
func buildWorkflowCopyResponse(copyItem models.WfProcessCopy, instance models.WfProcessInstance) models.WorkflowCopyResponse {
	return models.WorkflowCopyResponse{
		CopyID: copyItem.CopyID, InstanceID: copyItem.InstanceID, InstanceTitle: instance.Title,
		NodeID: copyItem.NodeID, NodeName: copyItem.NodeName, ReceiverID: copyItem.ReceiverID,
		ReceiverName: copyItem.ReceiverName, StarterName: instance.StarterName,
		Status: strconv.Itoa(copyItem.Status), ReadDate: models.TimeToStringPtr(copyItem.ReadDate),
		CreateDate: models.TimeToStringPtr(copyItem.CreateDate),
	}
}

// loadWorkflowInstancesByTasks 批量加载任务关联实例。
func loadWorkflowInstancesByTasks(tasks []models.WfProcessTask) (map[string]models.WfProcessInstance, error) {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.InstanceID)
	}
	return loadWorkflowInstanceMap(ids)
}

// loadWorkflowInstancesByCopies 批量加载抄送关联实例。
func loadWorkflowInstancesByCopies(copies []models.WfProcessCopy) (map[string]models.WfProcessInstance, error) {
	ids := make([]string, 0, len(copies))
	for _, copyItem := range copies {
		ids = append(ids, copyItem.InstanceID)
	}
	return loadWorkflowInstanceMap(ids)
}

// loadWorkflowInstanceMap 批量查询实例并构造ID映射。
func loadWorkflowInstanceMap(instanceIDs []string) (map[string]models.WfProcessInstance, error) {
	result := make(map[string]models.WfProcessInstance)
	if len(instanceIDs) == 0 {
		return result, nil
	}
	var instances []models.WfProcessInstance
	if err := database.DB.Where("instance_id IN ? AND del_flag = 0", uniqueStrings(instanceIDs)).Find(&instances).Error; err != nil {
		return nil, err
	}
	for _, instance := range instances {
		result[instance.InstanceID] = instance
	}
	return result, nil
}

// normalizeWorkflowPagination 规范流程列表分页并限制单页最大数量。
func normalizeWorkflowPagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// workflowUserName 返回用户展示名称。
func workflowUserName(user models.SysUser) string {
	if user.RealName != nil && *user.RealName != "" {
		return *user.RealName
	}
	if user.Username != nil {
		return *user.Username
	}
	return user.UserID
}

// uniqueStrings 对非空字符串去重。
func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// normalizeOptionalString 去除可选字符串首尾空白并将空值转为nil。
func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

// stringPtr 返回字符串指针。
func stringPtr(value string) *string { return &value }

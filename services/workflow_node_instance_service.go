package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// initializeWorkflowRoute 生成首版完整路径并从开始节点执行。
func initializeWorkflowRoute(tx *gorm.DB, context *workflowExecutionContext, startNodeID string) error {
	nodes, err := createWorkflowRoute(tx, context, startNodeID, 1, 1)
	if err != nil {
		return err
	}
	return activateWorkflowNode(tx, context, &nodes[0])
}

// rebuildWorkflowRouteAfterNode 使用最新变量替换当前节点之后尚未到达的路径。
func rebuildWorkflowRouteAfterNode(tx *gorm.DB, context *workflowExecutionContext, current *models.WfProcessNodeInstance) error {
	if err := supersedePlannedWorkflowNodes(tx, context.instance.InstanceID); err != nil {
		return err
	}
	node := findWorkflowNode(context.graph, current.NodeID)
	if node == nil {
		return fmt.Errorf("流程节点 %s 不存在", current.NodeID)
	}
	outgoing := workflowOutgoingEdges(context.graph, node.ID)
	if len(outgoing) != 1 {
		return fmt.Errorf("节点 %s 必须且只能有一条普通出线", workflowNodeName(node))
	}
	return createAndActivateWorkflowRoute(tx, context, outgoing[0].TargetNodeID)
}

// restartWorkflowRouteAtNode 从退回目标创建新路径版本并重新执行。
func restartWorkflowRouteAtNode(tx *gorm.DB, context *workflowExecutionContext, nodeID string) error {
	if err := supersedePlannedWorkflowNodes(tx, context.instance.InstanceID); err != nil {
		return err
	}
	return createAndActivateWorkflowRoute(tx, context, nodeID)
}

func createAndActivateWorkflowRoute(tx *gorm.DB, context *workflowExecutionContext, startNodeID string) error {
	routeVersion, sequence, err := nextWorkflowRoutePosition(tx, context.instance.InstanceID)
	if err != nil {
		return err
	}
	nodes, err := createWorkflowRoute(tx, context, startNodeID, routeVersion, sequence)
	if err != nil {
		return err
	}
	return activateWorkflowNode(tx, context, &nodes[0])
}

// createWorkflowRoute 根据当前变量物化从指定节点到结束节点的完整路径。
func createWorkflowRoute(tx *gorm.DB, context *workflowExecutionContext, startNodeID string, routeVersion, sequence int) ([]models.WfProcessNodeInstance, error) {
	createdAt := time.Now()
	nodes := make([]models.WfProcessNodeInstance, 0)
	currentNodeID := startNodeID
	for step := 0; step < workflowMaxAutomaticSteps; step++ {
		node := findWorkflowNode(context.graph, currentNodeID)
		if node == nil {
			return nil, fmt.Errorf("流程节点 %s 不存在", currentNodeID)
		}
		actorIDs, actorNames, err := resolveWorkflowNodeActorSnapshot(tx, context.instance, node)
		if err != nil {
			return nil, fmt.Errorf("节点 %s：%w", workflowNodeName(node), err)
		}
		actorIDsJSON, err := json.Marshal(actorIDs)
		if err != nil {
			return nil, err
		}
		actorNamesJSON, err := json.Marshal(actorNames)
		if err != nil {
			return nil, err
		}
		fieldPermissions := node.Properties.FieldPermissions
		if fieldPermissions == nil {
			fieldPermissions = map[string]string{}
		}
		fieldPermissionsJSON, err := json.Marshal(fieldPermissions)
		if err != nil {
			return nil, err
		}
		nodeInstance := models.WfProcessNodeInstance{
			NodeInstanceID: utils.GenerateUUID(), InstanceID: context.instance.InstanceID,
			NodeID: node.ID, NodeName: workflowNodeName(node), NodeType: node.Properties.NodeType,
			Sequence: sequence, RouteVersion: routeVersion, Status: models.WorkflowNodeStatusPlanned,
			ActorIDs: string(actorIDsJSON), ActorNames: string(actorNamesJSON), FieldPermissions: string(fieldPermissionsJSON),
			CreateDate: &createdAt, UpdateDate: &createdAt,
		}
		if node.Properties.NodeType == "approve" {
			approvalMode := node.Properties.ApprovalMode
			nodeInstance.ApprovalMode = &approvalMode
		}
		nextNodeID, branchEdgeID, done, err := nextWorkflowRouteNode(context.graph, node, context.variables)
		if err != nil {
			return nil, err
		}
		nodeInstance.BranchEdgeID = branchEdgeID
		nodes = append(nodes, nodeInstance)
		if done {
			break
		}
		currentNodeID = nextNodeID
		sequence++
	}
	if len(nodes) == 0 || nodes[len(nodes)-1].NodeType != "end" {
		return nil, fmt.Errorf("流程预计路径未到达结束节点")
	}
	if err := tx.Create(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func nextWorkflowRouteNode(graph *workflowGraph, node *workflowNode, variables map[string]interface{}) (string, *string, bool, error) {
	if node.Properties.NodeType == "end" {
		return "", nil, true, nil
	}
	if node.Properties.NodeType == "condition" {
		edge, err := selectWorkflowConditionEdge(graph, node.ID, variables)
		if err != nil {
			return "", nil, false, fmt.Errorf("条件节点 %s：%w", workflowNodeName(node), err)
		}
		return edge.TargetNodeID, stringPtr(edge.ID), false, nil
	}
	outgoing := workflowOutgoingEdges(graph, node.ID)
	if len(outgoing) != 1 {
		return "", nil, false, fmt.Errorf("节点 %s 必须且只能有一条普通出线", workflowNodeName(node))
	}
	return outgoing[0].TargetNodeID, nil, false, nil
}

func resolveWorkflowNodeActorSnapshot(tx *gorm.DB, instance *models.WfProcessInstance, node *workflowNode) ([]string, []string, error) {
	actorType := ""
	actorIDs := []string(nil)
	switch node.Properties.NodeType {
	case "approve":
		actorType = node.Properties.AssigneeType
		actorIDs = node.Properties.AssigneeIDs
	case "copy":
		actorType = node.Properties.CopyType
		actorIDs = node.Properties.CopyIDs
	default:
		if node.Properties.NodeType == "start" {
			return []string{instance.StarterID}, []string{instance.StarterName}, nil
		}
		return []string{}, []string{}, nil
	}
	users, err := resolveWorkflowActors(tx, instance, actorType, actorIDs)
	if err != nil {
		return nil, nil, err
	}
	userIDs := make([]string, 0, len(users))
	userNames := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.UserID)
		userNames = append(userNames, workflowUserName(user))
	}
	return userIDs, userNames, nil
}

// validateWorkflowGraphApprovers 校验整张流程图中所有审批节点均可解析到启用用户。
func validateWorkflowGraphApprovers(tx *gorm.DB, graph *workflowGraph, instance *models.WfProcessInstance) error {
	for index := range graph.Nodes {
		node := &graph.Nodes[index]
		if node.Properties.NodeType != "approve" {
			continue
		}
		if _, _, err := resolveWorkflowNodeActorSnapshot(tx, instance, node); err != nil {
			return fmt.Errorf("节点 %s：%w", workflowNodeName(node), err)
		}
	}
	return nil
}

// activateWorkflowNode 激活节点；自动节点在同一事务内继续推进。
func activateWorkflowNode(tx *gorm.DB, context *workflowExecutionContext, nodeInstance *models.WfProcessNodeInstance) error {
	now := time.Now()
	if err := tx.Model(nodeInstance).Updates(map[string]interface{}{
		"status": models.WorkflowNodeStatusActive, "start_date": now, "update_date": now,
	}).Error; err != nil {
		return err
	}
	nodeInstance.Status = models.WorkflowNodeStatusActive
	nodeInstance.StartDate = &now
	switch nodeInstance.NodeType {
	case "start":
		if err := createWorkflowRecord(tx, context.instance, nil, nodeInstance, "start", &context.instance.StarterID, &context.instance.StarterName, nil); err != nil {
			return err
		}
		return completeAndAdvanceWorkflowNode(tx, context, nodeInstance)
	case "approve":
		return createWorkflowApprovalTasks(tx, context, nodeInstance)
	case "copy":
		if err := createWorkflowCopies(tx, context.instance, nodeInstance); err != nil {
			return err
		}
		return completeAndAdvanceWorkflowNode(tx, context, nodeInstance)
	case "condition":
		if nodeInstance.BranchEdgeID == nil {
			return fmt.Errorf("条件节点 %s 未保存命中分支", nodeInstance.NodeName)
		}
		if err := createWorkflowRecord(tx, context.instance, nil, nodeInstance, "branch", nil, nil, nodeInstance.BranchEdgeID); err != nil {
			return err
		}
		return completeAndAdvanceWorkflowNode(tx, context, nodeInstance)
	case "end":
		if err := completeWorkflowNode(tx, nodeInstance); err != nil {
			return err
		}
		return finishWorkflowInstance(tx, context.instance, models.WorkflowInstanceStatusCompleted)
	default:
		return fmt.Errorf("不支持的流程节点类型：%s", nodeInstance.NodeType)
	}
}

func completeAndAdvanceWorkflowNode(tx *gorm.DB, context *workflowExecutionContext, nodeInstance *models.WfProcessNodeInstance) error {
	if err := completeWorkflowNode(tx, nodeInstance); err != nil {
		return err
	}
	var next models.WfProcessNodeInstance
	if err := tx.Where("instance_id = ? AND route_version = ? AND sequence > ? AND status = ?", nodeInstance.InstanceID, nodeInstance.RouteVersion, nodeInstance.Sequence, models.WorkflowNodeStatusPlanned).
		Order("sequence ASC").First(&next).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("节点 %s 后没有可执行节点", nodeInstance.NodeName)
		}
		return err
	}
	return activateWorkflowNode(tx, context, &next)
}

func completeWorkflowNode(tx *gorm.DB, nodeInstance *models.WfProcessNodeInstance) error {
	now := time.Now()
	if err := tx.Model(nodeInstance).Updates(map[string]interface{}{
		"status": models.WorkflowNodeStatusCompleted, "end_date": now, "update_date": now,
	}).Error; err != nil {
		return err
	}
	nodeInstance.Status = models.WorkflowNodeStatusCompleted
	nodeInstance.EndDate = &now
	return nil
}

func terminateWorkflowNode(tx *gorm.DB, nodeInstanceID string) error {
	now := time.Now()
	result := tx.Model(&models.WfProcessNodeInstance{}).
		Where("node_instance_id = ? AND status = ?", nodeInstanceID, models.WorkflowNodeStatusActive).
		Updates(map[string]interface{}{"status": models.WorkflowNodeStatusTerminated, "end_date": now, "update_date": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("当前流程节点不存在或已结束")
	}
	return nil
}

func supersedePlannedWorkflowNodes(tx *gorm.DB, instanceID string) error {
	return tx.Model(&models.WfProcessNodeInstance{}).
		Where("instance_id = ? AND status = ?", instanceID, models.WorkflowNodeStatusPlanned).
		Updates(map[string]interface{}{"status": models.WorkflowNodeStatusSuperseded, "update_date": time.Now()}).Error
}

func nextWorkflowRoutePosition(tx *gorm.DB, instanceID string) (int, int, error) {
	var latest models.WfProcessNodeInstance
	if err := tx.Where("instance_id = ?", instanceID).Order("sequence DESC").First(&latest).Error; err != nil {
		return 0, 0, err
	}
	return latest.RouteVersion + 1, latest.Sequence + 1, nil
}

func loadActiveWorkflowNode(tx *gorm.DB, nodeInstanceID string) (*models.WfProcessNodeInstance, error) {
	var nodeInstance models.WfProcessNodeInstance
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("node_instance_id = ? AND status = ?", nodeInstanceID, models.WorkflowNodeStatusActive).
		First(&nodeInstance).Error; err != nil {
		return nil, fmt.Errorf("当前流程节点不存在或已结束")
	}
	return &nodeInstance, nil
}

func createWorkflowApprovalTasks(tx *gorm.DB, context *workflowExecutionContext, nodeInstance *models.WfProcessNodeInstance) error {
	actorIDs, actorNames, err := workflowNodeActors(nodeInstance)
	if err != nil {
		return err
	}
	if nodeInstance.ApprovalMode == nil || (*nodeInstance.ApprovalMode != "any" && *nodeInstance.ApprovalMode != "all") {
		return fmt.Errorf("审批节点 %s 的审批方式无效", nodeInstance.NodeName)
	}
	previousApprovedActorIDs, err := loadAdjacentApprovedWorkflowActorIDs(tx, nodeInstance)
	if err != nil {
		return err
	}
	autoApprovedActorIDs := adjacentWorkflowAutoApprovedActorIDs(*nodeInstance.ApprovalMode, actorIDs, previousApprovedActorIDs)
	autoApprovedActorSet := make(map[string]bool, len(autoApprovedActorIDs))
	for _, actorID := range autoApprovedActorIDs {
		autoApprovedActorSet[actorID] = true
	}

	taskGroupID := utils.GenerateUUID()
	now := time.Now()
	tasks := make([]models.WfProcessTask, 0, len(actorIDs))
	for index, actorID := range actorIDs {
		task := models.WfProcessTask{
			TaskID: utils.GenerateUUID(), TaskGroupID: taskGroupID, NodeInstanceID: nodeInstance.NodeInstanceID,
			InstanceID: context.instance.InstanceID, NodeID: nodeInstance.NodeID, NodeName: nodeInstance.NodeName,
			AssigneeID: actorID, AssigneeName: actorNames[index], ApprovalMode: *nodeInstance.ApprovalMode,
			Status: models.WorkflowTaskStatusPending, CreateDate: &now, UpdateDate: &now,
		}
		if autoApprovedActorSet[actorID] {
			task.Status = models.WorkflowTaskStatusApproved
			task.Comment = stringPtr("与上一审批节点审批人相同，系统自动通过")
			task.FinishDate = &now
		} else if *nodeInstance.ApprovalMode == "any" && len(autoApprovedActorIDs) > 0 {
			task.Status = models.WorkflowTaskStatusCanceled
			task.FinishDate = &now
		}
		tasks = append(tasks, task)
	}
	if err := tx.Create(&tasks).Error; err != nil {
		return err
	}
	for index := range tasks {
		task := &tasks[index]
		if !autoApprovedActorSet[task.AssigneeID] {
			continue
		}
		if err := createWorkflowRecord(tx, context.instance, task, nil, "autoApprove", &task.AssigneeID, &task.AssigneeName, task.Comment); err != nil {
			return err
		}
	}
	if len(autoApprovedActorIDs) > 0 && (*nodeInstance.ApprovalMode == "any" || len(autoApprovedActorIDs) == len(actorIDs)) {
		return completeAndAdvanceWorkflowNode(tx, context, nodeInstance)
	}
	return nil
}

// loadAdjacentApprovedWorkflowActorIDs 返回实际执行顺序中相邻审批节点真正通过的人员。
func loadAdjacentApprovedWorkflowActorIDs(tx *gorm.DB, nodeInstance *models.WfProcessNodeInstance) ([]string, error) {
	var previousNode models.WfProcessNodeInstance
	err := tx.Where("instance_id = ? AND sequence < ? AND status IN ?", nodeInstance.InstanceID, nodeInstance.Sequence, []int{models.WorkflowNodeStatusCompleted, models.WorkflowNodeStatusTerminated}).
		Order("sequence DESC").First(&previousNode).Error
	if err == gorm.ErrRecordNotFound {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if previousNode.NodeType != "approve" || previousNode.Status != models.WorkflowNodeStatusCompleted {
		return []string{}, nil
	}
	var actorIDs []string
	if err := tx.Model(&models.WfProcessTask{}).
		Where("node_instance_id = ? AND status = ? AND del_flag = 0", previousNode.NodeInstanceID, models.WorkflowTaskStatusApproved).
		Order("create_date ASC").Pluck("assignee_id", &actorIDs).Error; err != nil {
		return nil, err
	}
	return uniqueStrings(actorIDs), nil
}

// adjacentWorkflowAutoApprovedActorIDs 按当前审批方式筛选相邻节点可自动通过的人员。
func adjacentWorkflowAutoApprovedActorIDs(approvalMode string, actorIDs, previousApprovedActorIDs []string) []string {
	previousActorSet := make(map[string]bool, len(previousApprovedActorIDs))
	for _, actorID := range previousApprovedActorIDs {
		previousActorSet[actorID] = true
	}
	autoApprovedActorIDs := make([]string, 0)
	for _, actorID := range actorIDs {
		if !previousActorSet[actorID] {
			continue
		}
		autoApprovedActorIDs = append(autoApprovedActorIDs, actorID)
		if approvalMode == "any" {
			break
		}
	}
	return autoApprovedActorIDs
}

func createWorkflowCopies(tx *gorm.DB, instance *models.WfProcessInstance, nodeInstance *models.WfProcessNodeInstance) error {
	actorIDs, actorNames, err := workflowNodeActors(nodeInstance)
	if err != nil {
		return err
	}
	now := time.Now()
	for index, actorID := range actorIDs {
		copyItem := models.WfProcessCopy{
			CopyID: utils.GenerateUUID(), NodeInstanceID: nodeInstance.NodeInstanceID, InstanceID: instance.InstanceID,
			NodeID: nodeInstance.NodeID, NodeName: nodeInstance.NodeName,
			ReceiverID: actorID, ReceiverName: actorNames[index], Status: models.WorkflowCopyStatusUnread,
			CreateDate: &now,
		}
		if err := tx.Create(&copyItem).Error; err != nil {
			return err
		}
	}
	return nil
}

func workflowNodeActors(nodeInstance *models.WfProcessNodeInstance) ([]string, []string, error) {
	var actorIDs []string
	var actorNames []string
	if err := json.Unmarshal([]byte(nodeInstance.ActorIDs), &actorIDs); err != nil {
		return nil, nil, fmt.Errorf("节点 %s 的参与人ID快照损坏", nodeInstance.NodeName)
	}
	if err := json.Unmarshal([]byte(nodeInstance.ActorNames), &actorNames); err != nil {
		return nil, nil, fmt.Errorf("节点 %s 的参与人名称快照损坏", nodeInstance.NodeName)
	}
	if len(actorIDs) == 0 || len(actorIDs) != len(actorNames) {
		return nil, nil, fmt.Errorf("节点 %s 的参与人快照无效", nodeInstance.NodeName)
	}
	return actorIDs, actorNames, nil
}

// buildWorkflowNodeResponses 将批量查询结果按节点实例聚合。
func buildWorkflowNodeResponses(nodeInstances []models.WfProcessNodeInstance, tasks []models.WfProcessTask, copies []models.WfProcessCopy, records []models.WfProcessRecord, instance models.WfProcessInstance, responseTime time.Time) ([]models.WorkflowNodeInstanceResponse, error) {
	taskMap := make(map[string][]models.WorkflowTaskResponse)
	for _, task := range tasks {
		taskMap[task.NodeInstanceID] = append(taskMap[task.NodeInstanceID], buildWorkflowTaskResponse(task, instance))
	}
	copyMap := make(map[string][]models.WorkflowCopyResponse)
	for _, copyItem := range copies {
		copyMap[copyItem.NodeInstanceID] = append(copyMap[copyItem.NodeInstanceID], buildWorkflowCopyResponse(copyItem, instance))
	}
	recordMap := make(map[string][]models.WorkflowRecordResponse)
	for _, record := range records {
		recordMap[record.NodeInstanceID] = append(recordMap[record.NodeInstanceID], buildWorkflowRecordResponse(record))
	}
	responses := make([]models.WorkflowNodeInstanceResponse, 0, len(nodeInstances))
	for _, nodeInstance := range nodeInstances {
		actorIDs := make([]string, 0)
		actorNames := make([]string, 0)
		if err := json.Unmarshal([]byte(nodeInstance.ActorIDs), &actorIDs); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(nodeInstance.ActorNames), &actorNames); err != nil {
			return nil, err
		}
		if len(actorIDs) != len(actorNames) {
			return nil, fmt.Errorf("节点 %s 的参与人快照无效", nodeInstance.NodeName)
		}
		fieldPermissions, err := workflowNodeFieldPermissions(&nodeInstance)
		if err != nil {
			return nil, err
		}
		durationSeconds, err := workflowApprovalDurationSeconds(nodeInstance, responseTime)
		if err != nil {
			return nil, err
		}
		actors := make([]models.WorkflowNodeActorResponse, 0, len(actorIDs))
		for index, actorID := range actorIDs {
			actors = append(actors, models.WorkflowNodeActorResponse{UserID: actorID, UserName: actorNames[index]})
		}
		nodeRecords := recordMap[nodeInstance.NodeInstanceID]
		responses = append(responses, models.WorkflowNodeInstanceResponse{
			NodeInstanceID: nodeInstance.NodeInstanceID, NodeID: nodeInstance.NodeID,
			NodeName: nodeInstance.NodeName, NodeType: nodeInstance.NodeType,
			Sequence: nodeInstance.Sequence, RouteVersion: nodeInstance.RouteVersion,
			Status: strconv.Itoa(nodeInstance.Status), Action: workflowNodeDisplayAction(nodeInstance, nodeRecords),
			ApprovalMode: nodeInstance.ApprovalMode, BranchEdgeID: nodeInstance.BranchEdgeID,
			FieldPermissions: fieldPermissions,
			Actors:           actors, Tasks: nonNilTasks(taskMap[nodeInstance.NodeInstanceID]),
			Copies: nonNilCopies(copyMap[nodeInstance.NodeInstanceID]), Records: nonNilRecords(nodeRecords),
			StartDate: models.TimeToStringPtr(nodeInstance.StartDate), EndDate: models.TimeToStringPtr(nodeInstance.EndDate),
			DurationSeconds: durationSeconds,
		})
	}
	return responses, nil
}

// workflowApprovalDurationSeconds 计算审批节点从到达到完成或查询时刻的停留秒数。
func workflowApprovalDurationSeconds(node models.WfProcessNodeInstance, responseTime time.Time) (*int64, error) {
	if node.NodeType != "approve" || node.Status == models.WorkflowNodeStatusPlanned || node.Status == models.WorkflowNodeStatusSuperseded {
		return nil, nil
	}
	if node.StartDate == nil {
		return nil, fmt.Errorf("审批节点 %s 缺少到达时间", node.NodeName)
	}

	endTime := responseTime
	switch node.Status {
	case models.WorkflowNodeStatusActive:
	case models.WorkflowNodeStatusCompleted, models.WorkflowNodeStatusTerminated:
		if node.EndDate == nil {
			return nil, fmt.Errorf("审批节点 %s 缺少结束时间", node.NodeName)
		}
		endTime = *node.EndDate
	default:
		return nil, fmt.Errorf("审批节点 %s 的状态无效", node.NodeName)
	}
	if endTime.Before(*node.StartDate) {
		return nil, fmt.Errorf("审批节点 %s 的结束时间早于到达时间", node.NodeName)
	}

	durationSeconds := int64(endTime.Sub(*node.StartDate).Seconds())
	return &durationSeconds, nil
}

func workflowNodeFieldPermissions(nodeInstance *models.WfProcessNodeInstance) (map[string]string, error) {
	fieldPermissions := make(map[string]string)
	if err := json.Unmarshal([]byte(nodeInstance.FieldPermissions), &fieldPermissions); err != nil {
		return nil, fmt.Errorf("节点 %s 的字段权限快照损坏", nodeInstance.NodeName)
	}
	if fieldPermissions == nil {
		return nil, fmt.Errorf("节点 %s 的字段权限快照无效", nodeInstance.NodeName)
	}
	return fieldPermissions, nil
}

func workflowNodeDisplayAction(node models.WfProcessNodeInstance, records []models.WorkflowRecordResponse) string {
	if node.Status == models.WorkflowNodeStatusPlanned {
		return "planned"
	}
	if node.Status == models.WorkflowNodeStatusActive {
		return "pending"
	}
	if len(records) > 0 {
		return records[len(records)-1].Action
	}
	switch node.NodeType {
	case "start":
		return "start"
	case "approve":
		return "approve"
	case "condition":
		return "branch"
	case "copy":
		return "copy"
	case "end":
		return "complete"
	default:
		return node.NodeType
	}
}

func nonNilTasks(items []models.WorkflowTaskResponse) []models.WorkflowTaskResponse {
	if items == nil {
		return []models.WorkflowTaskResponse{}
	}
	return items
}

func nonNilCopies(items []models.WorkflowCopyResponse) []models.WorkflowCopyResponse {
	if items == nil {
		return []models.WorkflowCopyResponse{}
	}
	return items
}

func nonNilRecords(items []models.WorkflowRecordResponse) []models.WorkflowRecordResponse {
	if items == nil {
		return []models.WorkflowRecordResponse{}
	}
	return items
}

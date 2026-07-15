package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const defaultWorkflowFlowData = `{"nodes":[],"edges":[]}`

func GetWorkflowDefinitions(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.WfProcessDefinition{}).Where("del_flag = ?", 0)

	if definitionKey, ok := params["definitionKey"].(string); ok && definitionKey != "" {
		db = db.Where("definition_key LIKE ?", "%"+definitionKey+"%")
	}
	if definitionName, ok := params["definitionName"].(string); ok && definitionName != "" {
		db = db.Where("definition_name LIKE ?", "%"+definitionName+"%")
	}
	if category, ok := params["category"].(string); ok && category != "" {
		db = db.Where("category = ?", category)
	}
	if statuses, ok := params["statuses"].([]int); ok && len(statuses) > 0 {
		db = db.Where("status IN ?", statuses)
	}

	sorts := params["sorts"].(string)
	order := utils.BuildOrderBy(sorts, map[string]string{
		"definitionKey":  "definition_key",
		"definitionName": "definition_name",
		"category":       "category",
		"status":         "status",
		"version":        "version",
		"createDate":     "create_date",
		"updateDate":     "update_date",
	})
	if order == "" {
		order = "create_date DESC"
	}

	return utils.PaginateWithTransform[models.WfProcessDefinition](db, page, pageSize, order, func(items []models.WfProcessDefinition) interface{} {
		return buildWorkflowDefinitionResponses(items)
	})
}

func GetAllWorkflowDefinitions(params map[string]interface{}) ([]models.WorkflowDefinitionResponse, error) {
	db := database.DB.Model(&models.WfProcessDefinition{}).Where("del_flag = ?", 0)

	if definitionName, ok := params["definitionName"].(string); ok && definitionName != "" {
		db = db.Where("definition_name LIKE ?", "%"+definitionName+"%")
	}
	if category, ok := params["category"].(string); ok && category != "" {
		db = db.Where("category = ?", category)
	}
	if status, ok := params["status"].(int); ok && status >= 0 {
		db = db.Where("status = ?", status)
	}

	var definitions []models.WfProcessDefinition
	err := db.Order("create_date DESC").Find(&definitions).Error
	if err != nil {
		return nil, err
	}

	return buildWorkflowDefinitionResponses(definitions), nil
}

func GetWorkflowDefinition(definitionID string) (*models.WorkflowDefinitionResponse, error) {
	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("流程定义不存在")
		}
		return nil, err
	}

	responses := buildWorkflowDefinitionResponses([]models.WfProcessDefinition{definition})
	return &responses[0], nil
}

func CreateWorkflowDefinition(req *models.CreateWorkflowDefinitionRequest, creatorID string) error {
	if err := ensureWorkflowDefinitionKeyUnique(req.DefinitionKey, ""); err != nil {
		return err
	}

	flowData := normalizeWorkflowFlowData(req.FlowData)
	if req.FlowData != nil && strings.TrimSpace(*req.FlowData) != "" {
		if err := validateWorkflowFlowData(flowData); err != nil {
			return err
		}
	}

	now := time.Now()
	definition := models.WfProcessDefinition{
		DefinitionID:   uuid.New().String(),
		DefinitionKey:  strings.TrimSpace(req.DefinitionKey),
		DefinitionName: strings.TrimSpace(req.DefinitionName),
		Category:       req.Category,
		Status:         0,
		Version:        0,
		FlowData:       &flowData,
		Remark:         req.Remark,
		CreatorID:      &creatorID,
		CreateDate:     &now,
		UpdateDate:     &now,
		DelFlag:        0,
	}

	return database.DB.Create(&definition).Error
}

func UpdateWorkflowDefinition(definitionID string, req *models.UpdateWorkflowDefinitionRequest) error {
	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}

	if err := ensureWorkflowDefinitionKeyUnique(req.DefinitionKey, definitionID); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"definition_key":  strings.TrimSpace(req.DefinitionKey),
		"definition_name": strings.TrimSpace(req.DefinitionName),
		"category":        req.Category,
		"remark":          req.Remark,
		"update_date":     time.Now(),
	}
	if req.FlowData != nil && strings.TrimSpace(*req.FlowData) != "" {
		flowData := normalizeWorkflowFlowData(req.FlowData)
		if err := validateWorkflowFlowData(flowData); err != nil {
			return err
		}
		updates["flow_data"] = flowData
		updates["status"] = 0
	}
	return database.DB.Model(&definition).Updates(updates).Error
}

func UpdateWorkflowCanvas(definitionID string, flowData string) error {
	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}

	if err := validateWorkflowFlowData(flowData); err != nil {
		return err
	}

	return database.DB.Model(&definition).Updates(map[string]interface{}{
		"flow_data":   strings.TrimSpace(flowData),
		"status":      0,
		"update_date": time.Now(),
	}).Error
}

func PublishWorkflowDefinition(definitionID string) error {
	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}

	flowData := defaultWorkflowFlowData
	if definition.FlowData != nil {
		flowData = *definition.FlowData
	}
	graph, err := parseWorkflowGraph(flowData)
	if err != nil {
		return err
	}
	formSchema, err := parseWorkflowFormSchema(definition.FormSchema)
	if err != nil {
		return err
	}
	if err := validateWorkflowConditionFields(graph, formSchema); err != nil {
		return err
	}

	version := definition.Version + 1
	if version < 1 {
		version = 1
	}

	return database.DB.Model(&definition).Updates(map[string]interface{}{
		"status":      1,
		"version":     version,
		"update_date": time.Now(),
	}).Error
}

func UpdateWorkflowDefinitionStatus(definitionID string, status int) error {
	if status != 0 && status != 1 && status != 2 {
		return fmt.Errorf("流程状态只能是 0、1、2")
	}
	if status == 1 {
		return PublishWorkflowDefinition(definitionID)
	}

	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}

	return database.DB.Model(&definition).Updates(map[string]interface{}{
		"status":      status,
		"update_date": time.Now(),
	}).Error
}

func DeleteWorkflowDefinitions(definitionIDs []string) error {
	if len(definitionIDs) == 0 {
		return nil
	}

	return database.DB.Model(&models.WfProcessDefinition{}).
		Where("definition_id IN ? AND del_flag = ?", definitionIDs, 0).
		Updates(map[string]interface{}{
			"del_flag":    1,
			"update_date": time.Now(),
		}).Error
}

func buildWorkflowDefinitionResponses(definitions []models.WfProcessDefinition) []models.WorkflowDefinitionResponse {
	creatorIDs := make([]string, 0)
	for _, definition := range definitions {
		if definition.CreatorID != nil {
			creatorIDs = append(creatorIDs, *definition.CreatorID)
		}
	}

	creatorNames := make(map[string]string)
	if len(creatorIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", creatorIDs).Find(&users)
		for _, user := range users {
			if user.RealName != nil {
				creatorNames[user.UserID] = *user.RealName
			}
		}
	}

	responses := make([]models.WorkflowDefinitionResponse, 0, len(definitions))
	for _, definition := range definitions {
		creatorName := creatorNames[utils.StringValue(definition.CreatorID)]
		responses = append(responses, models.WorkflowDefinitionResponse{
			DefinitionID:   &definition.DefinitionID,
			DefinitionKey:  definition.DefinitionKey,
			DefinitionName: definition.DefinitionName,
			Category:       definition.Category,
			Status:         fmt.Sprintf("%d", definition.Status),
			Version:        definition.Version,
			FlowData:       definition.FlowData,
			FormSchema:     definition.FormSchema,
			Remark:         definition.Remark,
			CreatorID:      definition.CreatorID,
			CreatorName:    &creatorName,
			CreateDate:     models.TimeToStringPtr(definition.CreateDate),
			UpdateDate:     models.TimeToStringPtr(definition.UpdateDate),
		})
	}

	return responses
}

func ensureWorkflowDefinitionKeyUnique(definitionKey string, excludeDefinitionID string) error {
	key := strings.TrimSpace(definitionKey)
	if key == "" {
		return fmt.Errorf("流程标识不能为空")
	}

	db := database.DB.Model(&models.WfProcessDefinition{}).
		Where("definition_key = ? AND del_flag = ?", key, 0)
	if excludeDefinitionID != "" {
		db = db.Where("definition_id <> ?", excludeDefinitionID)
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("流程标识已存在")
	}

	return nil
}

func normalizeWorkflowFlowData(flowData *string) string {
	if flowData == nil || strings.TrimSpace(*flowData) == "" {
		return defaultWorkflowFlowData
	}
	return strings.TrimSpace(*flowData)
}

func validateWorkflowFlowData(flowData string) error {
	if strings.TrimSpace(flowData) == "" {
		return fmt.Errorf("流程画布数据不能为空")
	}

	_, err := parseWorkflowGraph(flowData)
	return err
}

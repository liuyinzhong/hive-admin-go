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
	if businessType, ok := params["businessType"].(string); ok && businessType != "" {
		db = db.Where("business_type = ?", businessType)
	}
	if startType, ok := params["startType"].(int); ok && startType >= 0 {
		db = db.Where("start_type = ?", startType)
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
	if businessType, ok := params["businessType"].(string); ok && businessType != "" {
		db = db.Where("business_type = ?", businessType)
	}
	if startType, ok := params["startType"].(int); ok && startType >= 0 {
		db = db.Where("start_type = ?", startType)
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

// CreateWorkflowDefinition 创建流程定义。DefinitionKey 通过公共编码流水 WORKFLOW_DEFINITION 域自动生成，保证全局唯一且递增。
// BusinessType 必填且须为已注册业务类型;StartType 默认手动发起;
// IsDefault 仅被动触发流程可设为 true,同业务类型唯一,创建事务内自动顶掉该类型原默认。
func CreateWorkflowDefinition(req *models.CreateWorkflowDefinitionRequest, creatorID string) error {
	flowData := normalizeWorkflowFlowData(req.FlowData)
	if req.FlowData != nil && strings.TrimSpace(*req.FlowData) != "" {
		if err := validateWorkflowFlowData(flowData); err != nil {
			return err
		}
	}
	businessType, err := validateWorkflowBusinessType(req.BusinessType)
	if err != nil {
		return err
	}
	startType, err := normalizeWorkflowStartType(req.StartType)
	if err != nil {
		return err
	}
	isDefault, err := normalizeWorkflowIsDefault(req.IsDefault, startType)
	if err != nil {
		return err
	}

	now := time.Now()
	definition := models.WfProcessDefinition{
		DefinitionID:   uuid.New().String(),
		DefinitionName: strings.TrimSpace(req.DefinitionName),
		Category:       req.Category,
		BusinessType:   &businessType,
		StartType:      startType,
		IsDefault:      isDefault,
		Status:         0,
		Version:        0,
		FlowData:       &flowData,
		Remark:         req.Remark,
		CreatorID:      &creatorID,
		CreateDate:     &now,
		UpdateDate:     &now,
		DelFlag:        0,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		key, err := NewBaseCodeSequenceService().NextBusinessCode(tx, "WORKFLOW_DEFINITION", "WF", 6)
		if err != nil {
			return err
		}
		definition.DefinitionKey = key
		if err := tx.Create(&definition).Error; err != nil {
			return err
		}
		if isDefault == 1 {
			return clearWorkflowBusinessTypeDefault(tx, businessType, definition.DefinitionID)
		}
		return nil
	})
}

// UpdateWorkflowDefinition 更新流程定义。DefinitionKey 由后端自动生成且不可修改，更新时不接受前端传入的 DefinitionKey。
// 默认流程标志随表单提交维护:设为默认时在同事务内顶掉同业务类型原默认;改为手动发起时自动清空。
func UpdateWorkflowDefinition(definitionID string, req *models.UpdateWorkflowDefinitionRequest) error {
	var definition models.WfProcessDefinition
	err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}
	businessType, err := validateWorkflowBusinessType(req.BusinessType)
	if err != nil {
		return err
	}
	startType, err := normalizeWorkflowStartType(req.StartType)
	if err != nil {
		return err
	}
	isDefault, err := normalizeWorkflowIsDefault(req.IsDefault, startType)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"definition_name": strings.TrimSpace(req.DefinitionName),
		"category":        req.Category,
		"business_type":   businessType,
		"start_type":      startType,
		"remark":          req.Remark,
		"update_date":     time.Now(),
	}
	if startType == models.WorkflowStartTypeManual {
		updates["is_default"] = 0
	} else {
		updates["is_default"] = isDefault
		// 改为被动触发时自动解绑表单:被动触发流程由业务对象自动发起,没有发起人填表单环节
		updates["form_schema_id"] = gorm.Expr("NULL")
	}
	if req.FlowData != nil && strings.TrimSpace(*req.FlowData) != "" {
		flowData := normalizeWorkflowFlowData(req.FlowData)
		if err := validateWorkflowFlowData(flowData); err != nil {
			return err
		}
		updates["flow_data"] = flowData
		updates["status"] = 0
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&definition).Updates(updates).Error; err != nil {
			return err
		}
		if isDefault == 1 {
			return clearWorkflowBusinessTypeDefault(tx, businessType, definition.DefinitionID)
		}
		return nil
	})
}

// normalizeWorkflowIsDefault 规范化默认流程标志:仅被动触发流程可设为默认。
func normalizeWorkflowIsDefault(value *bool, startType int) (int, error) {
	if value == nil || !*value {
		return 0, nil
	}
	if startType != models.WorkflowStartTypePassive {
		return 0, fmt.Errorf("只有被动触发流程可以设为默认流程")
	}
	return 1, nil
}

// clearWorkflowBusinessTypeDefault 顶掉同业务类型下除当前定义外的原默认流程,保证同类型唯一。
// 先清后设在同一事务内串行执行,并发设默认时后提交者胜出,最终状态仍然唯一。
func clearWorkflowBusinessTypeDefault(tx *gorm.DB, businessType, excludeDefinitionID string) error {
	return tx.Model(&models.WfProcessDefinition{}).
		Where("business_type = ? AND is_default = 1 AND definition_id <> ? AND del_flag = 0", businessType, excludeDefinitionID).
		Updates(map[string]interface{}{"is_default": 0, "update_date": time.Now()}).Error
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
	// 表单 Schema 可选:未绑定表单的流程定义也允许发布(纯审批流程或与业务解耦的流程)
	var formFields []models.FormSchemaField
	if definition.FormSchemaID != nil && strings.TrimSpace(*definition.FormSchemaID) != "" {
		loadedFields, _, _, err := loadWorkflowFormSchema(database.DB, definition.FormSchemaID, true)
		if err != nil {
			return err
		}
		formFields = loadedFields
		hasValueField := false
		for _, field := range formFields {
			if formComponentHasValue(field.Component) {
				hasValueField = true
				break
			}
		}
		if !hasValueField {
			return fmt.Errorf("流程关联的表单 Schema 没有字段")
		}
	}
	if err := validateWorkflowConditionFields(graph, formFields); err != nil {
		return err
	}
	if err := validateWorkflowBusinessTypeRequired(&definition); err != nil {
		return err
	}
	if err := validateWorkflowAutomationMounts(&definition, graph, formFields); err != nil {
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

// validateWorkflowBusinessTypeRequired 校验发布定义的业务类型必填且已注册。
func validateWorkflowBusinessTypeRequired(definition *models.WfProcessDefinition) error {
	if definition.BusinessType == nil || strings.TrimSpace(*definition.BusinessType) == "" {
		return fmt.Errorf("流程定义必须声明业务类型后才能发布")
	}
	if _, exists := getBusinessTypeDef(strings.TrimSpace(*definition.BusinessType)); !exists {
		return fmt.Errorf("流程定义的业务类型 %s 未注册,无法发布", *definition.BusinessType)
	}
	// 被动触发流程由业务对象自动发起,表单的发起填写环节不存在,不允许携带表单发布(直接改库绕过绑定接口的兜底)
	if definition.StartType == models.WorkflowStartTypePassive &&
		definition.FormSchemaID != nil && strings.TrimSpace(*definition.FormSchemaID) != "" {
		return fmt.Errorf("被动触发流程由业务对象自动发起,不能关联表单,请先解除表单绑定")
	}
	return nil
}

// validateWorkflowAutomationMounts 校验画布节点挂载的自动化动作快照(动作类型×启动类型矩阵):
// 1. 修改字段值依赖业务流程绑定,仅被动触发流程可挂;插入记录依赖表单变量,仅手动发起流程可挂;
// 2. 修改字段值快照:目标字段在该业务类型状态字段白名单内、目标值非空;
// 3. 插入记录快照:目标业务类型已注册、必填字段映射齐全、值来源合法、表单字段名存在于当前流程绑定表单。
func validateWorkflowAutomationMounts(definition *models.WfProcessDefinition, graph *workflowGraph, formFields []models.FormSchemaField) error {
	for index := range graph.Nodes {
		node := &graph.Nodes[index]
		for _, mount := range node.Properties.Automations {
			nodeName := workflowNodeName(node)
			switch mount.ActionType {
			case models.ActionTypeUpdateField:
				if definition.StartType == models.WorkflowStartTypeManual {
					return fmt.Errorf("节点「%s」挂载了自动化动作「%s」:修改字段值需要业务对象,手动发起流程不能挂载,请将启动类型改为被动触发",
						nodeName, mount.AutomationName)
				}
				if err := validateUpdateFieldSnapshot(definition, mount, nodeName); err != nil {
					return err
				}
			case models.ActionTypeInsertRecord:
				if definition.StartType == models.WorkflowStartTypePassive {
					return fmt.Errorf("节点「%s」挂载了自动化动作「%s」:插入记录的值来源依赖表单字段,被动触发流程不能挂载,请将启动类型改为手动发起",
						nodeName, mount.AutomationName)
				}
				if err := validateInsertRecordSnapshot(definition, mount, formFields, nodeName); err != nil {
					return err
				}
			default:
				return fmt.Errorf("节点「%s」的自动化动作「%s」动作类型 %s 不受支持", nodeName, mount.AutomationName, mount.ActionType)
			}
		}
	}
	return nil
}

// validateUpdateFieldSnapshot 校验修改字段值快照:目标字段在该业务类型状态字段白名单内、目标值非空。
func validateUpdateFieldSnapshot(definition *models.WfProcessDefinition, mount workflowAutomationMount, nodeName string) error {
	if definition.BusinessType == nil {
		return nil
	}
	def, exists := getBusinessTypeDef(strings.TrimSpace(*definition.BusinessType))
	if !exists {
		return nil
	}
	for _, field := range def.StatusFields {
		if field.Field == mount.TargetField {
			if strings.TrimSpace(mount.TargetValue) == "" {
				return fmt.Errorf("节点「%s」的自动化动作「%s」目标值为空", nodeName, mount.AutomationName)
			}
			return nil
		}
	}
	return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 不在业务类型[%s]的可写字段白名单内",
		nodeName, mount.AutomationName, mount.TargetField, def.Label)
}

// validateInsertRecordSnapshot 校验插入记录快照:映射字段在动作业务类型的可插目录内、
// 必填字段映射齐全、值来源合法;表单字段名按当前流程绑定的表单字段强校验(弱引用的发布期拦截点)。
func validateInsertRecordSnapshot(definition *models.WfProcessDefinition, mount workflowAutomationMount, formFields []models.FormSchemaField, nodeName string) error {
	targetDef, exists := getBusinessTypeDef(mount.BusinessType)
	if !exists {
		return fmt.Errorf("节点「%s」的自动化动作「%s」业务类型 %s 未注册", nodeName, mount.AutomationName, mount.BusinessType)
	}
	if len(mount.Mappings) == 0 {
		return fmt.Errorf("节点「%s」的自动化动作「%s」没有字段映射", nodeName, mount.AutomationName)
	}
	insertFieldMap := make(map[string]businessTypeFieldDef, len(targetDef.InsertFields))
	for _, field := range targetDef.InsertFields {
		insertFieldMap[field.Field] = field
	}
	formFieldSet := make(map[string]bool, len(formFields))
	for _, field := range formFields {
		formFieldSet[field.FieldName] = true
	}
	mappedFields := make(map[string]bool, len(mount.Mappings))
	for _, mapping := range mount.Mappings {
		field, ok := insertFieldMap[mapping.Field]
		if !ok {
			return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 不在业务类型[%s]的可插字段目录内",
				nodeName, mount.AutomationName, mapping.Field, targetDef.Label)
		}
		if mappedFields[field.Field] {
			return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 重复映射", nodeName, mount.AutomationName, field.Field)
		}
		mappedFields[field.Field] = true
		switch mapping.SourceType {
		case models.InsertSourceFixed:
			if strings.TrimSpace(mapping.Value) == "" {
				return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 的固定值为空", nodeName, mount.AutomationName, field.Label)
			}
		case models.InsertSourceForm:
			if strings.TrimSpace(mapping.FormField) == "" {
				return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 的表单字段名为空", nodeName, mount.AutomationName, field.Label)
			}
			if !formFieldSet[mapping.FormField] {
				return fmt.Errorf("节点「%s」的自动化动作「%s」引用的表单字段 %s 不在流程绑定表单中,无法发布",
					nodeName, mount.AutomationName, mapping.FormField)
			}
		default:
			return fmt.Errorf("节点「%s」的自动化动作「%s」目标字段 %s 的值来源不合法", nodeName, mount.AutomationName, field.Field)
		}
	}
	for _, field := range targetDef.InsertFields {
		if field.Required && !mappedFields[field.Field] {
			return fmt.Errorf("节点「%s」的自动化动作「%s」缺少业务类型[%s]必填字段 %s 的映射",
				nodeName, mount.AutomationName, targetDef.Label, field.Label)
		}
	}
	return nil
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
			BusinessType:   definition.BusinessType,
			StartType:      definition.StartType,
			IsDefault:      definition.IsDefault == 1,
			Status:         fmt.Sprintf("%d", definition.Status),
			Version:        definition.Version,
			FlowData:       definition.FlowData,
			FormSchemaID:   definition.FormSchemaID,
			Remark:         definition.Remark,
			CreatorID:      definition.CreatorID,
			CreatorName:    &creatorName,
			CreateDate:     models.TimeToStringPtr(definition.CreateDate),
			UpdateDate:     models.TimeToStringPtr(definition.UpdateDate),
		})
	}

	return responses
}

// validateWorkflowBusinessType 校验业务类型必填且已注册(值为 BUSINESS_TYPE 字典值,与业务类型注册表一致)。
func validateWorkflowBusinessType(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("业务类型不能为空")
	}
	if _, exists := getBusinessTypeDef(trimmed); !exists {
		return "", fmt.Errorf("业务类型 %s 未注册", trimmed)
	}
	return trimmed, nil
}

// normalizeWorkflowStartType 规范化启动类型,未传时默认手动发起。
func normalizeWorkflowStartType(value *int) (int, error) {
	if value == nil {
		return models.WorkflowStartTypeManual, nil
	}
	if *value != models.WorkflowStartTypeManual && *value != models.WorkflowStartTypePassive {
		return 0, fmt.Errorf("启动类型只能是 0(手动发起)或 1(被动触发)")
	}
	return *value, nil
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

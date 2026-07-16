package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// UpdateWorkflowFormSchema 绑定流程定义使用的独立表单 Schema。
func UpdateWorkflowFormSchema(definitionID, formSchemaID string) error {
	if _, err := uuid.Parse(definitionID); err != nil {
		return fmt.Errorf("流程定义ID无效")
	}
	if _, err := uuid.Parse(formSchemaID); err != nil {
		return fmt.Errorf("表单 Schema ID 无效")
	}
	var definition models.WfProcessDefinition
	if err := database.DB.Where("definition_id = ? AND del_flag = 0", definitionID).First(&definition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}
	if _, _, _, err := loadWorkflowFormSchema(database.DB, &formSchemaID, true); err != nil {
		return err
	}
	return database.DB.Model(&definition).Updates(map[string]interface{}{
		"form_schema_id": formSchemaID,
		"status":         0,
		"update_date":    time.Now(),
	}).Error
}

// loadWorkflowFormSchema 加载流程绑定的可用表单并返回字段投影、原始JSON和布局。
func loadWorkflowFormSchema(db *gorm.DB, formSchemaID *string, requireEnabled bool) ([]models.FormSchemaField, string, string, error) {
	if formSchemaID == nil || strings.TrimSpace(*formSchemaID) == "" {
		return nil, "", "", fmt.Errorf("流程定义未关联表单 Schema")
	}
	query := db.Where("form_schema_id = ? AND del_flag = 0", *formSchemaID)
	if requireEnabled {
		query = query.Where("status = 1")
	}
	var schema models.SysFormSchema
	if err := query.First(&schema).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", "", fmt.Errorf("流程关联的表单 Schema 不存在或已禁用")
		}
		return nil, "", "", err
	}
	fields, _, err := parseAndValidateFormSchema(json.RawMessage(schema.SchemaJSON))
	if err != nil {
		return nil, "", "", err
	}
	return fields, schema.SchemaJSON, schema.Layout, nil
}

// parseWorkflowFormSnapshot 解析流程实例发起时保存的表单快照。
func parseWorkflowFormSnapshot(snapshot *string) ([]models.FormSchemaField, error) {
	if snapshot == nil || strings.TrimSpace(*snapshot) == "" {
		return nil, fmt.Errorf("流程实例缺少表单快照")
	}
	fields, _, err := parseAndValidateFormSchema(json.RawMessage(*snapshot))
	return fields, err
}

// validateWorkflowConditionFields 校验条件分支和节点权限引用的字段。
func validateWorkflowConditionFields(graph *workflowGraph, formFields []models.FormSchemaField) error {
	fields := make(map[string]models.FormSchemaField, len(formFields))
	for _, field := range formFields {
		if formComponentHasValue(field.Component) {
			fields[field.FieldName] = field
		}
	}
	for _, edge := range graph.Edges {
		for _, rule := range edge.Properties.ConditionRules {
			field, exists := fields[rule.Field]
			if !exists {
				return fmt.Errorf("条件分支引用了不存在的表单字段“%s”", rule.Field)
			}
			label := formFieldLabel(field)
			if isWorkflowNumericOperator(rule.Operator) && !formComponentIsNumeric(field.Component) {
				return fmt.Errorf("字段“%s”不是数字，不能使用大小比较", label)
			}
			if isWorkflowContainsOperator(rule.Operator) && !formComponentSupportsContains(field.Component) {
				return fmt.Errorf("字段“%s”不支持包含比较", label)
			}
		}
	}
	for _, node := range graph.Nodes {
		if node.Properties.NodeType != "approve" {
			continue
		}
		for fieldName, permission := range node.Properties.FieldPermissions {
			field, exists := fields[fieldName]
			if !exists {
				return fmt.Errorf("审批节点“%s”的字段权限引用了不存在的字段“%s”", workflowNodeName(&node), fieldName)
			}
			if permission != "hidden" && permission != "readonly" && permission != "editable" {
				return fmt.Errorf("审批节点“%s”的字段“%s”权限无效", workflowNodeName(&node), formFieldLabel(field))
			}
		}
	}
	return nil
}

func formFieldLabel(field models.FormSchemaField) string {
	if strings.TrimSpace(field.Label) != "" {
		return strings.TrimSpace(field.Label)
	}
	return field.FieldName
}

func formComponentIsNumeric(component string) bool {
	return component == "InputNumber" || component == "Rate"
}

func formComponentSupportsContains(component string) bool {
	switch component {
	case "AutoComplete", "CheckboxGroup", "Input", "InputPassword", "Mentions", "RichEditor", "Textarea", "VbenInput", "VbenInputPassword":
		return true
	default:
		return false
	}
}

func isWorkflowNumericOperator(operator string) bool {
	return operator == "greaterThan" || operator == "greaterThanOrEqual" || operator == "lessThan" || operator == "lessThanOrEqual"
}

func isWorkflowContainsOperator(operator string) bool {
	return operator == "contains" || operator == "notContains"
}

package services

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

const workflowFormSchemaVersion = 1

const workflowFormGridType = "grid"

var workflowFormFieldKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

var workflowFormFieldTypes = map[string]struct{}{
	"checkbox": {},
	"date":     {},
	"input":    {},
	"number":   {},
	"radio":    {},
	"select":   {},
	"switch":   {},
	"textarea": {},
}

// UpdateWorkflowForm 校验并保存流程定义绑定的申请表单结构。
func UpdateWorkflowForm(definitionID string, schema *models.WorkflowFormSchema) error {
	if err := validateWorkflowFormSchema(schema); err != nil {
		return err
	}

	var definition models.WfProcessDefinition
	if err := database.DB.Where("definition_id = ? AND del_flag = ?", definitionID, 0).First(&definition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("流程定义不存在")
		}
		return err
	}

	formSchema, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("流程表单结构无法序列化")
	}
	return database.DB.Model(&definition).Updates(map[string]interface{}{
		"form_schema": string(formSchema),
		"status":      0,
		"update_date": time.Now(),
	}).Error
}

// parseWorkflowFormSchema 解析数据库中的流程表单结构。
func parseWorkflowFormSchema(value *string) (*models.WorkflowFormSchema, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, fmt.Errorf("流程尚未配置申请表单")
	}
	var schema models.WorkflowFormSchema
	if err := json.Unmarshal([]byte(*value), &schema); err != nil {
		return nil, fmt.Errorf("流程表单结构损坏")
	}
	if err := validateWorkflowFormSchema(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// validateWorkflowFormSchema 校验表单版本、字段标识、类型和选项配置。
func validateWorkflowFormSchema(schema *models.WorkflowFormSchema) error {
	if schema == nil {
		return fmt.Errorf("流程表单不能为空")
	}
	if schema.Version != workflowFormSchemaVersion {
		return fmt.Errorf("暂不支持该流程表单版本")
	}
	if len(schema.Fields) == 0 {
		return fmt.Errorf("流程表单至少需要一个字段")
	}
	if len(schema.Fields) > 100 {
		return fmt.Errorf("流程表单元素不能超过100个")
	}

	ids := make(map[string]struct{}, len(schema.Fields))
	keys := make(map[string]struct{}, len(schema.Fields))
	fieldCount := 0
	for index := range schema.Fields {
		element := &schema.Fields[index]
		element.ID = strings.TrimSpace(element.ID)
		element.Type = strings.TrimSpace(element.Type)
		if element.Type == workflowFormGridType {
			if err := validateWorkflowFormGrid(element, index, ids, keys, &fieldCount); err != nil {
				return err
			}
			continue
		}
		fieldCount++
		if err := validateWorkflowFormField(element, index+1, ids, keys); err != nil {
			return err
		}
	}
	if fieldCount == 0 {
		return fmt.Errorf("流程表单至少需要一个字段")
	}
	if fieldCount > 100 {
		return fmt.Errorf("流程表单字段不能超过100个")
	}
	return nil
}

// validateWorkflowFormGrid 校验一个顶层栅格布局及其列内字段。
func validateWorkflowFormGrid(grid *models.WorkflowFormField, index int, ids, keys map[string]struct{}, fieldCount *int) error {
	if grid.ID == "" {
		return fmt.Errorf("第%d个栅格布局的ID不能为空", index+1)
	}
	if err := registerWorkflowFormID(grid.ID, ids); err != nil {
		return err
	}
	if len(grid.Columns) == 0 || len(grid.Columns) > 24 {
		return fmt.Errorf("第%d个栅格布局必须包含1到24列", index+1)
	}

	grid.Key = ""
	grid.Label = ""
	grid.Required = false
	grid.Placeholder = nil
	grid.DefaultValue = nil
	grid.Options = nil
	totalSpan := 0
	for columnIndex := range grid.Columns {
		column := &grid.Columns[columnIndex]
		column.ID = strings.TrimSpace(column.ID)
		if column.ID == "" {
			return fmt.Errorf("第%d个栅格布局的第%d列ID不能为空", index+1, columnIndex+1)
		}
		if err := registerWorkflowFormID(column.ID, ids); err != nil {
			return err
		}
		if column.Span < 1 || column.Span > 24 {
			return fmt.Errorf("第%d个栅格布局的第%d列宽度必须在1到24之间", index+1, columnIndex+1)
		}
		totalSpan += column.Span
		if totalSpan > 24 {
			return fmt.Errorf("第%d个栅格布局的列宽总和不能超过24", index+1)
		}
		for fieldIndex := range column.Fields {
			field := &column.Fields[fieldIndex]
			field.Type = strings.TrimSpace(field.Type)
			if field.Type == workflowFormGridType {
				return fmt.Errorf("栅格布局不支持嵌套")
			}
			*fieldCount++
			if err := validateWorkflowFormField(field, fieldIndex+1, ids, keys); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateWorkflowFormField 校验一个真实业务字段。
func validateWorkflowFormField(field *models.WorkflowFormField, index int, ids, keys map[string]struct{}) error {
	field.ID = strings.TrimSpace(field.ID)
	field.Key = strings.TrimSpace(field.Key)
	field.Label = strings.TrimSpace(field.Label)
	field.Type = strings.TrimSpace(field.Type)
	field.Columns = nil
	if field.ID == "" || field.Key == "" || field.Label == "" {
		return fmt.Errorf("第%d个表单字段的名称和标识不能为空", index)
	}
	if !workflowFormFieldKeyPattern.MatchString(field.Key) {
		return fmt.Errorf("字段“%s”的标识只能以字母开头，并包含字母、数字或下划线", field.Label)
	}
	if _, exists := workflowFormFieldTypes[field.Type]; !exists {
		return fmt.Errorf("字段“%s”的类型不受支持", field.Label)
	}
	if err := registerWorkflowFormID(field.ID, ids); err != nil {
		return err
	}
	if _, exists := keys[field.Key]; exists {
		return fmt.Errorf("字段标识“%s”不能重复", field.Key)
	}
	keys[field.Key] = struct{}{}
	if err := validateWorkflowFormOptions(field); err != nil {
		return err
	}
	if !workflowFormValueEmpty(field.DefaultValue) {
		if err := validateWorkflowFormValue(*field, field.DefaultValue); err != nil {
			return fmt.Errorf("字段“%s”的默认值无效: %w", field.Label, err)
		}
	}
	return nil
}

// registerWorkflowFormID 记录布局、列和字段共用的元素ID。
func registerWorkflowFormID(id string, ids map[string]struct{}) error {
	if _, exists := ids[id]; exists {
		return fmt.Errorf("表单元素ID不能重复")
	}
	ids[id] = struct{}{}
	return nil
}

// workflowFormFields 按结构顺序返回所有真实业务字段。
func workflowFormFields(schema *models.WorkflowFormSchema) []models.WorkflowFormField {
	fields := make([]models.WorkflowFormField, 0)
	for _, element := range schema.Fields {
		if element.Type != workflowFormGridType {
			fields = append(fields, element)
			continue
		}
		for _, column := range element.Columns {
			fields = append(fields, column.Fields...)
		}
	}
	return fields
}

// validateWorkflowFormOptions 校验选择类字段的选项完整性和唯一性。
func validateWorkflowFormOptions(field *models.WorkflowFormField) error {
	requiresOptions := field.Type == "checkbox" || field.Type == "radio" || field.Type == "select"
	if !requiresOptions {
		field.Options = nil
		return nil
	}
	if len(field.Options) == 0 {
		return fmt.Errorf("字段“%s”至少需要一个选项", field.Label)
	}
	values := make(map[string]struct{}, len(field.Options))
	for index := range field.Options {
		option := &field.Options[index]
		option.Label = strings.TrimSpace(option.Label)
		option.Value = strings.TrimSpace(option.Value)
		if option.Label == "" || option.Value == "" {
			return fmt.Errorf("字段“%s”的选项名称和值不能为空", field.Label)
		}
		if _, exists := values[option.Value]; exists {
			return fmt.Errorf("字段“%s”的选项值不能重复", field.Label)
		}
		values[option.Value] = struct{}{}
	}
	return nil
}

// validateWorkflowFormVariables 按发布时的表单结构校验申请数据。
func validateWorkflowFormVariables(schema *models.WorkflowFormSchema, variables map[string]interface{}) error {
	formFields := workflowFormFields(schema)
	fields := make(map[string]models.WorkflowFormField, len(formFields))
	for _, field := range formFields {
		fields[field.Key] = field
	}
	for key := range variables {
		if _, exists := fields[key]; !exists {
			return fmt.Errorf("申请数据包含未知字段“%s”", key)
		}
	}
	for _, field := range formFields {
		value, exists := variables[field.Key]
		if !exists || workflowFormValueEmpty(value) {
			if field.Required {
				return fmt.Errorf("请填写“%s”", field.Label)
			}
			continue
		}
		if err := validateWorkflowFormValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

// validateWorkflowFormValue 校验单个申请字段的值类型和选项范围。
func validateWorkflowFormValue(field models.WorkflowFormField, value interface{}) error {
	switch field.Type {
	case "input", "textarea", "date":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("字段“%s”的值类型错误", field.Label)
		}
	case "number":
		if _, ok := workflowNumber(value); !ok {
			return fmt.Errorf("字段“%s”必须是数字", field.Label)
		}
	case "switch":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("字段“%s”必须是布尔值", field.Label)
		}
	case "select", "radio":
		selected, ok := value.(string)
		if !ok || !workflowFormOptionExists(field.Options, selected) {
			return fmt.Errorf("字段“%s”的选项无效", field.Label)
		}
	case "checkbox":
		values := reflect.ValueOf(value)
		if values.Kind() != reflect.Slice && values.Kind() != reflect.Array {
			return fmt.Errorf("字段“%s”必须是选项数组", field.Label)
		}
		for index := 0; index < values.Len(); index++ {
			selected, ok := values.Index(index).Interface().(string)
			if !ok || !workflowFormOptionExists(field.Options, selected) {
				return fmt.Errorf("字段“%s”的选项无效", field.Label)
			}
		}
	}
	return nil
}

// workflowFormOptionExists 判断选项值是否存在于字段配置中。
func workflowFormOptionExists(options []models.WorkflowFormOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

// workflowFormValueEmpty 判断申请字段值是否为空。
func workflowFormValueEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	reflected := reflect.ValueOf(value)
	return (reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.Array) && reflected.Len() == 0
}

// validateWorkflowConditionFields 校验条件分支引用的字段存在且操作符适配字段类型。
func validateWorkflowConditionFields(graph *workflowGraph, schema *models.WorkflowFormSchema) error {
	formFields := workflowFormFields(schema)
	fields := make(map[string]models.WorkflowFormField, len(formFields))
	for _, field := range formFields {
		fields[field.Key] = field
	}
	for _, edge := range graph.Edges {
		for _, rule := range edge.Properties.ConditionRules {
			field, exists := fields[rule.Field]
			if !exists {
				return fmt.Errorf("条件分支引用了不存在的表单字段“%s”", rule.Field)
			}
			if isWorkflowNumericOperator(rule.Operator) && field.Type != "number" {
				return fmt.Errorf("字段“%s”不是数字，不能使用大小比较", field.Label)
			}
			if isWorkflowContainsOperator(rule.Operator) && field.Type != "checkbox" && field.Type != "input" && field.Type != "textarea" {
				return fmt.Errorf("字段“%s”不支持包含比较", field.Label)
			}
		}
	}
	for _, node := range graph.Nodes {
		if node.Properties.NodeType != "approve" {
			continue
		}
		for key, permission := range node.Properties.FieldPermissions {
			field, exists := fields[key]
			if !exists {
				return fmt.Errorf("审批节点“%s”的字段权限引用了不存在的字段“%s”", workflowNodeName(&node), key)
			}
			if permission != "hidden" && permission != "readonly" && permission != "editable" {
				return fmt.Errorf("审批节点“%s”的字段“%s”权限无效", workflowNodeName(&node), field.Label)
			}
		}
	}
	return nil
}

// isWorkflowNumericOperator 判断条件操作符是否为数字大小比较。
func isWorkflowNumericOperator(operator string) bool {
	return operator == "greaterThan" || operator == "greaterThanOrEqual" || operator == "lessThan" || operator == "lessThanOrEqual"
}

// isWorkflowContainsOperator 判断条件操作符是否为包含比较。
func isWorkflowContainsOperator(operator string) bool {
	return operator == "contains" || operator == "notContains"
}

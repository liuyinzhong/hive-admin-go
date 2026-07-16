package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const maxFormSchemaFields = 200

var formFieldNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)*$`)

var supportedFormComponents = map[string]struct{}{
	"ApiCascader": {}, "ApiSelect": {}, "ApiTreeSelect": {}, "AutoComplete": {},
	"Cascader": {}, "Checkbox": {}, "CheckboxGroup": {}, "CollapsibleParams": {},
	"ColorSelect": {}, "DatePicker": {}, "DefaultButton": {}, "Divider": {},
	"IconPicker": {}, "Input": {}, "InputNumber": {}, "InputPassword": {},
	"Mentions": {}, "PrimaryButton": {}, "Radio": {}, "RadioGroup": {},
	"RangePicker": {}, "Rate": {}, "RichEditor": {}, "Select": {}, "Space": {},
	"Switch": {}, "Textarea": {}, "TimePicker": {}, "TreeSelect": {}, "Upload": {},
	"VbenCheckbox": {}, "VbenInput": {}, "VbenInputPassword": {}, "VbenPinInput": {}, "VbenSelect": {},
}

func GetFormSchemas(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.SysFormSchema{}).Where("del_flag = ?", 0)
	if value, ok := params["schemaKey"].(string); ok && strings.TrimSpace(value) != "" {
		db = db.Where("schema_key LIKE ?", "%"+strings.TrimSpace(value)+"%")
	}
	if value, ok := params["schemaName"].(string); ok && strings.TrimSpace(value) != "" {
		db = db.Where("schema_name LIKE ?", "%"+strings.TrimSpace(value)+"%")
	}
	if value, ok := params["category"].(string); ok && strings.TrimSpace(value) != "" {
		db = db.Where("category = ?", strings.TrimSpace(value))
	}
	if status, ok := params["status"].(int); ok && status >= 0 {
		db = db.Where("status = ?", status)
	}
	order := utils.BuildOrderBy(params["sorts"].(string), map[string]string{
		"schemaKey": "schema_key", "schemaName": "schema_name", "category": "category",
		"status": "status", "createDate": "create_date", "updateDate": "update_date",
	})
	if order == "" {
		order = "create_date DESC"
	}
	return utils.PaginateWithTransform[models.SysFormSchema](db, page, pageSize, order, func(items []models.SysFormSchema) interface{} {
		responses, _ := buildFormSchemaResponses(items)
		return responses
	})
}

func GetAllFormSchemas(params map[string]interface{}) ([]models.FormSchemaResponse, error) {
	db := database.DB.Where("del_flag = ?", 0)
	if value, ok := params["schemaName"].(string); ok && strings.TrimSpace(value) != "" {
		db = db.Where("schema_name LIKE ?", "%"+strings.TrimSpace(value)+"%")
	}
	if status, ok := params["status"].(int); ok && status >= 0 {
		db = db.Where("status = ?", status)
	}
	var schemas []models.SysFormSchema
	if err := db.Order("create_date DESC").Find(&schemas).Error; err != nil {
		return nil, err
	}
	return buildFormSchemaResponses(schemas)
}

func GetFormSchema(formSchemaID string) (*models.FormSchemaResponse, error) {
	if _, err := uuid.Parse(formSchemaID); err != nil {
		return nil, fmt.Errorf("表单 Schema ID 无效")
	}
	var schema models.SysFormSchema
	if err := database.DB.Where("form_schema_id = ? AND del_flag = 0", formSchemaID).First(&schema).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("表单 Schema 不存在")
		}
		return nil, err
	}
	responses, err := buildFormSchemaResponses([]models.SysFormSchema{schema})
	if err != nil {
		return nil, err
	}
	return &responses[0], nil
}

func CreateFormSchema(req *models.UpsertFormSchemaRequest, creatorID string) error {
	key, name, layout, status, schemaJSON, err := normalizeFormSchemaRequest(req)
	if err != nil {
		return err
	}
	if err := ensureFormSchemaKeyUnique(key, ""); err != nil {
		return err
	}
	now := time.Now()
	return database.DB.Create(&models.SysFormSchema{
		FormSchemaID: utils.GenerateUUID(), SchemaKey: key, SchemaName: name,
		Category: normalizeOptionalString(req.Category), Layout: layout, SchemaJSON: schemaJSON, Status: status,
		Remark: normalizeOptionalString(req.Remark), CreatorID: &creatorID,
		CreateDate: &now, UpdateDate: &now, DelFlag: 0,
	}).Error
}

func UpdateFormSchema(formSchemaID string, req *models.UpsertFormSchemaRequest) error {
	if _, err := uuid.Parse(formSchemaID); err != nil {
		return fmt.Errorf("表单 Schema ID 无效")
	}
	var schema models.SysFormSchema
	if err := database.DB.Where("form_schema_id = ? AND del_flag = 0", formSchemaID).First(&schema).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("表单 Schema 不存在")
		}
		return err
	}
	key, name, layout, status, schemaJSON, err := normalizeFormSchemaRequest(req)
	if err != nil {
		return err
	}
	if err := ensureFormSchemaKeyUnique(key, formSchemaID); err != nil {
		return err
	}
	requiresRepublish := schema.SchemaJSON != schemaJSON || schema.Layout != layout || schema.Status != status
	return database.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&schema).Updates(map[string]interface{}{
			"schema_key": key, "schema_name": name, "category": normalizeOptionalString(req.Category), "layout": layout,
			"schema_json": schemaJSON, "status": status, "remark": normalizeOptionalString(req.Remark),
			"update_date": now,
		}).Error; err != nil {
			return err
		}
		if !requiresRepublish {
			return nil
		}
		return tx.Model(&models.WfProcessDefinition{}).
			Where("form_schema_id = ? AND del_flag = 0", formSchemaID).
			Updates(map[string]interface{}{"status": 0, "update_date": now}).Error
	})
}

func DeleteFormSchemas(formSchemaIDs []string) error {
	if len(formSchemaIDs) == 0 {
		return nil
	}
	for _, id := range formSchemaIDs {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("表单 Schema ID 无效")
		}
	}
	var referenced int64
	if err := database.DB.Model(&models.WfProcessDefinition{}).
		Where("form_schema_id IN ? AND del_flag = 0", formSchemaIDs).Count(&referenced).Error; err != nil {
		return err
	}
	if referenced > 0 {
		return fmt.Errorf("表单 Schema 已被流程引用，不能删除")
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, formSchemaID := range formSchemaIDs {
			if err := tx.Model(&models.SysFormSchema{}).
				Where("form_schema_id = ? AND del_flag = 0", formSchemaID).
				Updates(map[string]interface{}{
					"del_flag": 1, "schema_key": formSchemaID, "update_date": now,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func normalizeFormSchemaRequest(req *models.UpsertFormSchemaRequest) (string, string, string, int, string, error) {
	key := strings.TrimSpace(req.SchemaKey)
	name := strings.TrimSpace(req.SchemaName)
	if !formFieldNamePattern.MatchString(key) || len(key) > 128 {
		return "", "", "", 0, "", fmt.Errorf("Schema 标识只能使用字母、数字、下划线和点，且必须以字母开头")
	}
	if name == "" || utf8.RuneCountInString(name) > 128 {
		return "", "", "", 0, "", fmt.Errorf("Schema 名称不能为空且不能超过128个字符")
	}
	layout := strings.TrimSpace(req.Layout)
	if layout != models.FormSchemaLayoutSingle && layout != models.FormSchemaLayoutDouble && layout != models.FormSchemaLayoutTriple {
		return "", "", "", 0, "", fmt.Errorf("表单布局只能是single、double或triple")
	}
	status := 1
	if req.Status != nil {
		parsed, err := strconv.Atoi(*req.Status)
		if err != nil || (parsed != 0 && parsed != 1) {
			return "", "", "", 0, "", fmt.Errorf("Schema 状态只能是0或1")
		}
		status = parsed
	}
	_, schemaJSON, err := parseAndValidateFormSchema(req.Schema)
	if err != nil {
		return "", "", "", 0, "", err
	}
	return key, name, layout, status, schemaJSON, nil
}

func parseAndValidateFormSchema(raw json.RawMessage) ([]models.FormSchemaField, string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || !json.Valid(raw) {
		return nil, "", fmt.Errorf("表单 Schema 必须是有效 JSON")
	}
	var fields []models.FormSchemaField
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, "", fmt.Errorf("表单 Schema 必须是字段数组")
	}
	if len(fields) > maxFormSchemaFields {
		return nil, "", fmt.Errorf("表单字段不能超过%d个", maxFormSchemaFields)
	}
	fieldNames := make(map[string]struct{}, len(fields))
	for index := range fields {
		field := &fields[index]
		field.FieldName = strings.TrimSpace(field.FieldName)
		field.Component = strings.TrimSpace(field.Component)
		if !formFieldNamePattern.MatchString(field.FieldName) {
			return nil, "", fmt.Errorf("第%d个字段的 fieldName 无效", index+1)
		}
		if _, exists := fieldNames[field.FieldName]; exists {
			return nil, "", fmt.Errorf("字段“%s”不能重复", field.FieldName)
		}
		for existing := range fieldNames {
			if strings.HasPrefix(field.FieldName, existing+".") || strings.HasPrefix(existing, field.FieldName+".") {
				return nil, "", fmt.Errorf("字段“%s”和“%s”不能互为父子字段", field.FieldName, existing)
			}
		}
		fieldNames[field.FieldName] = struct{}{}
		if _, supported := supportedFormComponents[field.Component]; !supported {
			return nil, "", fmt.Errorf("字段“%s”的组件“%s”不受支持", field.FieldName, field.Component)
		}
		if err := validateFormSchemaRules(field); err != nil {
			return nil, "", err
		}
	}
	compact, err := json.Marshal(json.RawMessage(raw))
	if err != nil {
		return nil, "", fmt.Errorf("表单 Schema 无法序列化")
	}
	return fields, string(compact), nil
}

func validateFormSchemaRules(field *models.FormSchemaField) error {
	for _, rule := range field.Rules {
		switch rule.Type {
		case "array", "boolean", "date", "email", "enum", "number", "required", "regex", "selectRequired", "string", "url":
		case "custom":
			return fmt.Errorf("字段“%s”的自定义校验处理器尚未在后端注册", field.FieldName)
		default:
			return fmt.Errorf("字段“%s”的校验类型“%s”无效", field.FieldName, rule.Type)
		}
		if rule.Length != nil && *rule.Length < 0 {
			return fmt.Errorf("字段“%s”的固定长度不能小于0", field.FieldName)
		}
		if rule.Min != nil && rule.Max != nil && *rule.Min > *rule.Max {
			return fmt.Errorf("字段“%s”的最小值不能大于最大值", field.FieldName)
		}
		if rule.Pattern != nil {
			if len(*rule.Pattern) > 512 {
				return fmt.Errorf("字段“%s”的正则表达式不能超过512个字符", field.FieldName)
			}
			if _, err := regexp.Compile(*rule.Pattern); err != nil {
				return fmt.Errorf("字段“%s”的正则表达式无效", field.FieldName)
			}
		}
	}
	return nil
}

func validateFormSchemaValues(fields []models.FormSchemaField, values map[string]interface{}) error {
	valueFields := make(map[string]models.FormSchemaField)
	for _, field := range fields {
		if formComponentHasValue(field.Component) {
			valueFields[field.FieldName] = field
		}
	}
	if err := validateFormSchemaValuePaths(valueFields, values); err != nil {
		return err
	}
	for _, field := range valueFields {
		value, _ := formSchemaValueAtPath(values, field.FieldName)
		if err := validateFormSchemaFieldValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

type formSchemaPathNode struct {
	children map[string]*formSchemaPathNode
	terminal bool
}

func validateFormSchemaValuePaths(fields map[string]models.FormSchemaField, values map[string]interface{}) error {
	root := &formSchemaPathNode{children: make(map[string]*formSchemaPathNode)}
	for fieldName := range fields {
		node := root
		for _, part := range strings.Split(fieldName, ".") {
			if node.children[part] == nil {
				node.children[part] = &formSchemaPathNode{children: make(map[string]*formSchemaPathNode)}
			}
			node = node.children[part]
		}
		node.terminal = true
	}
	return validateFormSchemaValuePathNode(root, values, "")
}

func validateFormSchemaValuePathNode(node *formSchemaPathNode, values map[string]interface{}, parent string) error {
	for key, value := range values {
		path := key
		if parent != "" {
			path = parent + "." + key
		}
		child := node.children[key]
		if child == nil {
			return fmt.Errorf("表单数据包含未知字段“%s”", path)
		}
		if child.terminal || value == nil {
			continue
		}
		nested, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("表单字段“%s”必须是对象", path)
		}
		if err := validateFormSchemaValuePathNode(child, nested, path); err != nil {
			return err
		}
	}
	return nil
}

func formSchemaValueAtPath(values map[string]interface{}, fieldName string) (interface{}, bool) {
	var current interface{} = values
	for _, part := range strings.Split(fieldName, ".") {
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

func validateFormSchemaFieldValue(field models.FormSchemaField, value interface{}) error {
	label := field.Label
	if label == "" {
		label = field.FieldName
	}
	for _, rule := range field.Rules {
		if rule.Type == "required" || rule.Type == "selectRequired" {
			if formValueEmpty(value) {
				return formRuleError(rule, fmt.Sprintf("请填写%s", label))
			}
		}
	}
	if formValueEmpty(value) {
		return nil
	}
	for _, rule := range field.Rules {
		if err := validateFormSchemaRuleValue(rule, value, label); err != nil {
			return err
		}
	}
	return nil
}

func validateFormSchemaRuleValue(rule models.FormSchemaRule, value interface{}, label string) error {
	length := -1
	switch typed := value.(type) {
	case string:
		length = utf8.RuneCountInString(typed)
	case []interface{}:
		length = len(typed)
	}
	switch rule.Type {
	case "string", "email", "url", "regex", "date":
		if _, ok := value.(string); !ok {
			return formRuleError(rule, fmt.Sprintf("%s必须是文本", label))
		}
	case "number":
		if _, ok := formNumber(value); !ok {
			return formRuleError(rule, fmt.Sprintf("%s必须是数字", label))
		}
	case "array":
		if reflect.ValueOf(value).Kind() != reflect.Slice {
			return formRuleError(rule, fmt.Sprintf("%s必须是数组", label))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return formRuleError(rule, fmt.Sprintf("%s必须是布尔值", label))
		}
	}
	if rule.Length != nil && length >= 0 && length != *rule.Length {
		return formRuleError(rule, fmt.Sprintf("%s长度必须为%d", label, *rule.Length))
	}
	actual, numeric := formNumber(value)
	if !numeric && length >= 0 {
		actual, numeric = float64(length), true
	}
	if rule.Min != nil && numeric && actual < *rule.Min {
		return formRuleError(rule, fmt.Sprintf("%s不能小于%v", label, *rule.Min))
	}
	if rule.Max != nil && numeric && actual > *rule.Max {
		return formRuleError(rule, fmt.Sprintf("%s不能大于%v", label, *rule.Max))
	}
	if rule.Integer && numeric && math.Trunc(actual) != actual {
		return formRuleError(rule, fmt.Sprintf("%s必须是整数", label))
	}
	text, _ := value.(string)
	if rule.Type == "regex" && rule.Pattern != nil {
		matched, _ := regexp.MatchString(*rule.Pattern, text)
		if !matched {
			return formRuleError(rule, fmt.Sprintf("%s格式不正确", label))
		}
	}
	if rule.Type == "email" && !regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(text) {
		return formRuleError(rule, fmt.Sprintf("%s格式不正确", label))
	}
	if rule.Type == "url" {
		parsed, err := url.ParseRequestURI(text)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return formRuleError(rule, fmt.Sprintf("%s格式不正确", label))
		}
	}
	if rule.Type == "enum" && len(rule.Values) > 0 {
		matched := false
		for _, expected := range rule.Values {
			if reflect.DeepEqual(value, expected) {
				matched = true
				break
			}
		}
		if !matched {
			return formRuleError(rule, fmt.Sprintf("%s不在允许范围内", label))
		}
	}
	return nil
}

func formRuleError(rule models.FormSchemaRule, fallback string) error {
	if rule.Message != nil && strings.TrimSpace(*rule.Message) != "" {
		return fmt.Errorf("%s", strings.TrimSpace(*rule.Message))
	}
	return fmt.Errorf("%s", fallback)
}

func formComponentHasValue(component string) bool {
	switch component {
	case "CollapsibleParams", "DefaultButton", "Divider", "PrimaryButton", "Space":
		return false
	default:
		return true
	}
}

func formValueEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	reflected := reflect.ValueOf(value)
	return (reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.Array) && reflected.Len() == 0
}

func formNumber(value interface{}) (float64, bool) {
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

func ensureFormSchemaKeyUnique(key, excludeID string) error {
	db := database.DB.Model(&models.SysFormSchema{}).Where("schema_key = ? AND del_flag = 0", key)
	if excludeID != "" {
		db = db.Where("form_schema_id <> ?", excludeID)
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("Schema 标识已存在")
	}
	return nil
}

func buildFormSchemaResponses(schemas []models.SysFormSchema) ([]models.FormSchemaResponse, error) {
	creatorIDs := make([]string, 0)
	for _, schema := range schemas {
		if schema.CreatorID != nil {
			creatorIDs = append(creatorIDs, *schema.CreatorID)
		}
	}
	creatorNames := make(map[string]string)
	if len(creatorIDs) > 0 {
		var users []models.SysUser
		if err := database.DB.Where("user_id IN ?", creatorIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			creatorNames[user.UserID] = workflowUserName(user)
		}
	}
	responses := make([]models.FormSchemaResponse, 0, len(schemas))
	for _, schema := range schemas {
		if !json.Valid([]byte(schema.SchemaJSON)) {
			return nil, fmt.Errorf("表单 Schema“%s”数据无效", schema.SchemaName)
		}
		creatorName := creatorNames[utils.StringValue(schema.CreatorID)]
		responses = append(responses, models.FormSchemaResponse{
			FormSchemaID: schema.FormSchemaID, SchemaKey: schema.SchemaKey, SchemaName: schema.SchemaName,
			Category: schema.Category, Layout: schema.Layout, Schema: json.RawMessage(schema.SchemaJSON), Status: strconv.Itoa(schema.Status),
			Remark: schema.Remark, CreatorID: schema.CreatorID, CreatorName: &creatorName,
			CreateDate: models.TimeToStringPtr(schema.CreateDate), UpdateDate: models.TimeToStringPtr(schema.UpdateDate),
		})
	}
	return responses, nil
}

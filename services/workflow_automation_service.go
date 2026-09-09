package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// CreateAutomation 创建自动化动作。
// 按动作类型反序列化并校验参数:修改字段值校验状态字段白名单与字典值;
// 插入记录校验目标业务类型、可插字段目录、必填映射与固定值合法性(表单字段名为弱引用,不做存在性校验)。
func CreateAutomation(req *models.CreateAutomationRequest, creatorID string) error {
	configJSON, err := validateAndEncodeAutomationConfig(req.BusinessType, req.ActionType, req.ActionConfig)
	if err != nil {
		return err
	}
	status := 0
	if req.Status != nil {
		status = *req.Status
	}
	now := time.Now()
	automation := models.WfAutomation{
		AutomationID:   utils.GenerateUUID(),
		AutomationName: strings.TrimSpace(req.AutomationName),
		BusinessType:   req.BusinessType,
		ActionType:     req.ActionType,
		ActionConfig:   string(configJSON),
		Status:         status,
		Remark:         normalizeOptionalString(req.Remark),
		CreatorID:      &creatorID,
		CreateDate:     &now,
		UpdateDate:     &now,
		DelFlag:        0,
	}
	return database.DB.Create(&automation).Error
}

// UpdateAutomation 更新自动化动作。
// 已挂载到流程画布的快照不受影响,重新挂载或重新发布对应流程才使用新配置。
func UpdateAutomation(automationID string, req *models.UpdateAutomationRequest) error {
	var automation models.WfAutomation
	if err := database.DB.Where("automation_id = ? AND del_flag = 0", automationID).
		First(&automation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("自动化动作不存在")
		}
		return err
	}
	configJSON, err := validateAndEncodeAutomationConfig(req.BusinessType, req.ActionType, req.ActionConfig)
	if err != nil {
		return err
	}
	status := automation.Status
	if req.Status != nil {
		status = *req.Status
	}
	return database.DB.Model(&automation).Updates(map[string]interface{}{
		"automation_name": strings.TrimSpace(req.AutomationName),
		"business_type":   req.BusinessType,
		"action_type":     req.ActionType,
		"action_config":   string(configJSON),
		"status":          status,
		"remark":          normalizeOptionalString(req.Remark),
		"update_date":     time.Now(),
	}).Error
}

// DeleteAutomations 软删除自动化动作。已挂载到画布的快照不受影响,继续按快照执行。
func DeleteAutomations(automationIDs []string) error {
	if len(automationIDs) == 0 {
		return nil
	}
	return database.DB.Model(&models.WfAutomation{}).
		Where("automation_id IN ? AND del_flag = 0", automationIDs).
		Updates(map[string]interface{}{"del_flag": 1, "update_date": time.Now()}).Error
}

// GetAutomations 分页查询自动化动作,支持按名称、业务类型和状态筛选。
func GetAutomations(page, pageSize int, name, businessType string, status *int) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.WfAutomation{}).Where("del_flag = 0")
	if name != "" {
		db = db.Where("automation_name LIKE ?", "%"+name+"%")
	}
	if businessType != "" {
		db = db.Where("business_type = ?", businessType)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	page, pageSize = normalizeWorkflowPagination(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.WfAutomation
	if err := db.Order("create_date DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	creatorIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.CreatorID != nil {
			creatorIDs = append(creatorIDs, *item.CreatorID)
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
	responses := make([]models.AutomationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildAutomationResponse(item, creatorNames))
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

// GetAutomationOptions 按业务类型返回启用的自动化动作选项,供设计器节点挂载选择器。
// 数据为动作库元数据,不涉及业务记录,无需数据权限校验。
func GetAutomationOptions(businessType string) ([]models.AutomationResponse, error) {
	if _, exists := getBusinessTypeDef(businessType); !exists && businessType != "" {
		return nil, fmt.Errorf("业务类型未注册: %s", businessType)
	}
	db := database.DB.Model(&models.WfAutomation{}).
		Where("del_flag = 0 AND status = 0")
	if businessType != "" {
		db = db.Where("business_type = ?", businessType)
	}
	var items []models.WfAutomation
	if err := db.Order("create_date DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	responses := make([]models.AutomationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildAutomationResponse(item, nil))
	}
	return responses, nil
}

// validateAndEncodeAutomationConfig 按动作类型反序列化、校验并重新序列化动作参数。
func validateAndEncodeAutomationConfig(businessType, actionType string, raw json.RawMessage) (json.RawMessage, error) {
	switch actionType {
	case models.ActionTypeUpdateField:
		var config models.UpdateFieldConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, fmt.Errorf("动作参数格式错误")
		}
		if err := validateUpdateFieldConfig(config); err != nil {
			return nil, err
		}
		return json.Marshal(config)
	case models.ActionTypeInsertRecord:
		var config models.InsertRecordConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, fmt.Errorf("动作参数格式错误")
		}
		if err := validateInsertRecordConfig(businessType, config); err != nil {
			return nil, err
		}
		return json.Marshal(config)
	default:
		return nil, fmt.Errorf("不支持的动作类型: %s", actionType)
	}
}

// validateUpdateFieldConfig 校验修改字段值参数:业务类型已注册、目标字段在状态字段白名单内、目标值为合法字典值。
func validateUpdateFieldConfig(config models.UpdateFieldConfig) error {
	targetField := strings.TrimSpace(config.TargetField)
	if targetField == "" {
		return fmt.Errorf("目标字段不能为空")
	}
	if strings.TrimSpace(config.TargetValue) == "" {
		return fmt.Errorf("目标值不能为空")
	}
	// 目标字段可能属于任一业务类型的状态白名单,遍历注册表定位
	for _, def := range businessTypeRegistry {
		for _, field := range def.StatusFields {
			if field.Field != targetField {
				continue
			}
			return validateAutomationDictValue(field.DictType, config.TargetValue)
		}
	}
	return fmt.Errorf("目标字段 %s 不在任何业务类型的可写字段白名单内", targetField)
}

// validateInsertRecordConfig 校验插入记录参数:映射字段在动作业务类型的可插目录内、
// 必填字段映射齐全、值来源合法;固定值的字典与引用字段做配置期存在性校验,表单字段名为弱引用不校验存在性。
// 插入目标即动作自身的业务类型,不单独配置目标业务类型。
func validateInsertRecordConfig(businessType string, config models.InsertRecordConfig) error {
	def, exists := getBusinessTypeDef(businessType)
	if !exists {
		return fmt.Errorf("业务类型 %s 未注册", businessType)
	}
	if len(config.Mappings) == 0 {
		return fmt.Errorf("插入记录至少需要配置一条字段映射")
	}
	insertFieldMap := make(map[string]businessTypeFieldDef, len(def.InsertFields))
	for _, field := range def.InsertFields {
		insertFieldMap[field.Field] = field
	}
	mappedFields := make(map[string]bool, len(config.Mappings))
	for index, mapping := range config.Mappings {
		field, ok := insertFieldMap[strings.TrimSpace(mapping.Field)]
		if !ok {
			return fmt.Errorf("第%d条映射的目标字段 %s 不在业务类型[%s]的可插字段目录内", index+1, mapping.Field, def.Label)
		}
		if mappedFields[field.Field] {
			return fmt.Errorf("目标字段 %s 重复配置映射", field.Field)
		}
		mappedFields[field.Field] = true
		switch mapping.SourceType {
		case models.InsertSourceFixed:
			if strings.TrimSpace(mapping.Value) == "" {
				return fmt.Errorf("目标字段 %s 的固定值不能为空", field.Label)
			}
			if field.DictType != "" {
				if err := validateAutomationDictValue(field.DictType, mapping.Value); err != nil {
					return fmt.Errorf("目标字段 %s: %w", field.Label, err)
				}
			}
			if ref, isRef := insertRefValidators[field.Field]; isRef {
				if err := validateInsertRefExists(database.DB, field.Field, mapping.Value, ref); err != nil {
					return fmt.Errorf("目标字段 %s: %w", field.Label, err)
				}
			}
		case models.InsertSourceForm:
			if strings.TrimSpace(mapping.FormField) == "" {
				return fmt.Errorf("目标字段 %s 的表单字段名不能为空", field.Label)
			}
		default:
			return fmt.Errorf("目标字段 %s 的值来源不合法,只支持固定值或表单字段", field.Label)
		}
	}
	for _, field := range def.InsertFields {
		if field.Required && !mappedFields[field.Field] {
			return fmt.Errorf("业务类型[%s]的必填字段 %s 未配置映射", def.Label, field.Label)
		}
	}
	return nil
}

// validateInsertRefExists 校验引用字段指向的记录存在且未删除。
func validateInsertRefExists(tx *gorm.DB, field, value string, ref insertRefDef) error {
	var count int64
	if err := tx.Model(ref.Model).
		Where(field+" = ? AND del_flag = 0", value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("引用的%s(%s)不存在或已删除", ref.Label, value)
	}
	return nil
}

// validateAutomationDictValue 校验目标值是对应字典的合法值,防止配出无效状态。
func validateAutomationDictValue(dictType, value string) error {
	var count int64
	if err := database.DB.Model(&models.SysDict{}).
		Where("type = ? AND value = ? AND del_flag = 0", dictType, value).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("目标值 %s 不是字典 %s 的合法值", value, dictType)
	}
	return nil
}

// buildAutomationResponse 组装动作响应,按动作类型附带参数与目录元数据供前端渲染摘要。
func buildAutomationResponse(item models.WfAutomation, creatorNames map[string]string) models.AutomationResponse {
	response := models.AutomationResponse{
		AutomationID:   &item.AutomationID,
		AutomationName: item.AutomationName,
		BusinessType:   item.BusinessType,
		ActionType:     item.ActionType,
		Status:         fmt.Sprintf("%d", item.Status),
		Remark:         item.Remark,
		CreatorID:      item.CreatorID,
		CreateDate:     models.TimeToStringPtr(item.CreateDate),
		UpdateDate:     models.TimeToStringPtr(item.UpdateDate),
	}
	if item.CreatorID != nil {
		creatorName := creatorNames[*item.CreatorID]
		response.CreatorName = &creatorName
	}
	switch item.ActionType {
	case models.ActionTypeUpdateField:
		var config models.UpdateFieldConfig
		_ = json.Unmarshal([]byte(item.ActionConfig), &config)
		updateField := &models.AutomationUpdateFieldResponse{
			TargetField: config.TargetField,
			TargetValue: config.TargetValue,
		}
		if def, exists := getBusinessTypeDef(item.BusinessType); exists {
			for _, field := range def.StatusFields {
				if field.Field == config.TargetField {
					updateField.TargetFieldLabel = field.Label
					updateField.DictType = field.DictType
					break
				}
			}
		}
		response.UpdateField = updateField
	case models.ActionTypeInsertRecord:
		var config models.InsertRecordConfig
		_ = json.Unmarshal([]byte(item.ActionConfig), &config)
		insertRecord := &models.AutomationInsertRecordResponse{
			Mappings: make([]models.AutomationInsertFieldResponse, 0, len(config.Mappings)),
		}
		insertFieldMap := make(map[string]businessTypeFieldDef)
		if targetDef, exists := getBusinessTypeDef(item.BusinessType); exists {
			for _, field := range targetDef.InsertFields {
				insertFieldMap[field.Field] = field
			}
		}
		for _, mapping := range config.Mappings {
			row := models.AutomationInsertFieldResponse{
				Field:      mapping.Field,
				SourceType: mapping.SourceType,
				Value:      mapping.Value,
				FormField:  mapping.FormField,
			}
			if field, exists := insertFieldMap[mapping.Field]; exists {
				row.FieldLabel = field.Label
				row.Required = field.Required
				row.DictType = field.DictType
			}
			_, row.IsRefField = insertRefValidators[mapping.Field]
			insertRecord.Mappings = append(insertRecord.Mappings, row)
		}
		response.InsertRecord = insertRecord
	}
	return response
}

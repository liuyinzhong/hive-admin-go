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

// automationActionTypeUpdateField 修改当前关联业务的状态字段为固定值(版本1唯一动作类型)。
const automationActionTypeUpdateField = "update_field"

// CreateAutomation 创建自动化动作。
// 校验业务类型已注册、动作类型受支持、目标字段在该业务类型可写字段白名单内、目标值为合法字典值。
func CreateAutomation(req *models.CreateAutomationRequest, creatorID string) error {
	fieldDef, err := validateAutomationConfig(req.BusinessType, req.ActionType, req.ActionConfig)
	if err != nil {
		return err
	}
	if err := validateAutomationDictValue(fieldDef.DictType, req.ActionConfig.TargetValue); err != nil {
		return err
	}
	configJSON, err := json.Marshal(req.ActionConfig)
	if err != nil {
		return fmt.Errorf("动作参数序列化失败")
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
	fieldDef, err := validateAutomationConfig(req.BusinessType, req.ActionType, req.ActionConfig)
	if err != nil {
		return err
	}
	if err := validateAutomationDictValue(fieldDef.DictType, req.ActionConfig.TargetValue); err != nil {
		return err
	}
	configJSON, err := json.Marshal(req.ActionConfig)
	if err != nil {
		return fmt.Errorf("动作参数序列化失败")
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

// validateAutomationConfig 校验动作配置的业务类型、动作类型与目标字段白名单。
func validateAutomationConfig(businessType, actionType string, config models.AutomationActionConfig) (businessTypeFieldDef, error) {
	def, exists := getBusinessTypeDef(businessType)
	if !exists {
		return businessTypeFieldDef{}, fmt.Errorf("业务类型未注册: %s", businessType)
	}
	if actionType != automationActionTypeUpdateField {
		return businessTypeFieldDef{}, fmt.Errorf("不支持的动作类型: %s", actionType)
	}
	targetField := strings.TrimSpace(config.TargetField)
	if targetField == "" {
		return businessTypeFieldDef{}, fmt.Errorf("目标字段不能为空")
	}
	for _, field := range def.StatusFields {
		if field.Field == targetField {
			if strings.TrimSpace(config.TargetValue) == "" {
				return businessTypeFieldDef{}, fmt.Errorf("目标值不能为空")
			}
			return field, nil
		}
	}
	return businessTypeFieldDef{}, fmt.Errorf("目标字段 %s 不在业务类型[%s]的可写字段白名单内", targetField, def.Label)
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

// buildAutomationResponse 组装动作响应,附带字段元数据供前端渲染摘要与字典翻译。
func buildAutomationResponse(item models.WfAutomation, creatorNames map[string]string) models.AutomationResponse {
	var config models.AutomationActionConfig
	_ = json.Unmarshal([]byte(item.ActionConfig), &config)
	fieldLabel, dictType := "", ""
	if def, exists := getBusinessTypeDef(item.BusinessType); exists {
		for _, field := range def.StatusFields {
			if field.Field == config.TargetField {
				fieldLabel = field.Label
				dictType = field.DictType
				break
			}
		}
	}
	creatorName := ""
	if item.CreatorID != nil {
		creatorName = creatorNames[*item.CreatorID]
	}
	return models.AutomationResponse{
		AutomationID:     &item.AutomationID,
		AutomationName:   item.AutomationName,
		BusinessType:     item.BusinessType,
		ActionType:       item.ActionType,
		TargetField:      config.TargetField,
		TargetValue:      config.TargetValue,
		TargetFieldLabel: fieldLabel,
		DictType:         dictType,
		Status:           fmt.Sprintf("%d", item.Status),
		Remark:           item.Remark,
		CreatorID:        item.CreatorID,
		CreatorName:      &creatorName,
		CreateDate:       models.TimeToStringPtr(item.CreateDate),
		UpdateDate:       models.TimeToStringPtr(item.UpdateDate),
	}
}

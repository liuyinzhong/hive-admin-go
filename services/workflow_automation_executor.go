package services

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// changeBehaviorAutomation 自动化动作变更行为,CHANGE_BEHAVIOR 字典值 60。
const changeBehaviorAutomation = 60

// executeWorkflowAutomations 执行节点挂载的自动化动作快照。
// 按画布挂载顺序在同一流程事务内执行,任一动作失败返回错误,整个流程流转回滚;
// 退回重走路径时经过的自动节点动作会再次执行,两类动作均幂等安全(设固定值/插入新记录各自独立)。
// operatorID 为触发本次节点完成的操作人:审批节点取完成审批的办理人,自动节点取流程发起人,写入业务变更记录。
// nodeName 用于变更记录正文定位触发节点。
// 返回本次动作新插入的需求 ID,供事务提交成功后链式自动发起需求流程。
func executeWorkflowAutomations(tx *gorm.DB, instance *models.WfProcessInstance, mounts []workflowAutomationMount, variables map[string]interface{}, operatorID, nodeName string) ([]string, error) {
	pendingStoryIDs := make([]string, 0)
	if len(mounts) == 0 {
		return pendingStoryIDs, nil
	}
	binding, err := getWorkflowBusinessInstanceByInstanceID(tx, instance.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("查询业务流程关联失败: %w", err)
	}
	for _, mount := range mounts {
		switch mount.ActionType {
		case models.ActionTypeUpdateField:
			if binding == nil {
				return nil, fmt.Errorf("流程实例未绑定业务对象,无法执行自动化动作「%s」,请由业务对象发起本流程", mount.AutomationName)
			}
			if err := executeWorkflowAutomationUpdateField(tx, instance, binding, mount, operatorID, nodeName); err != nil {
				return nil, err
			}
		case models.ActionTypeInsertRecord:
			storyID, err := executeWorkflowAutomationInsertRecord(tx, instance, mount, variables, operatorID, nodeName)
			if err != nil {
				return nil, err
			}
			pendingStoryIDs = append(pendingStoryIDs, storyID)
		default:
			return nil, fmt.Errorf("自动化动作「%s」的动作类型 %s 不受支持", mount.AutomationName, mount.ActionType)
		}
	}
	return pendingStoryIDs, nil
}

// executeWorkflowAutomationUpdateField 执行修改字段值动作:
// 定位当前关联业务 -> 校验快照字段仍在白名单内 -> 读取旧值 -> 更新状态字段 -> 写业务变更记录
// (不写参与人、不推送通知)。
func executeWorkflowAutomationUpdateField(tx *gorm.DB, instance *models.WfProcessInstance, binding *models.WfBusinessInstance, mount workflowAutomationMount, operatorID, nodeName string) error {
	def, exists := getBusinessTypeDef(binding.BusinessType)
	if !exists {
		return fmt.Errorf("绑定的业务类型 %s 未注册,无法执行自动化动作「%s」", binding.BusinessType, mount.AutomationName)
	}
	businessTypeInt, err := strconv.Atoi(binding.BusinessType)
	if err != nil {
		return fmt.Errorf("绑定的业务类型 %s 不是有效数值", binding.BusinessType)
	}
	var fieldDef *businessTypeFieldDef
	for index := range def.StatusFields {
		if def.StatusFields[index].Field == mount.TargetField {
			fieldDef = &def.StatusFields[index]
			break
		}
	}
	if fieldDef == nil {
		return fmt.Errorf("自动化动作「%s」的目标字段 %s 不在业务类型[%s]的可写字段白名单内", mount.AutomationName, mount.TargetField, def.Label)
	}
	targetValue, err := strconv.Atoi(mount.TargetValue)
	if err != nil {
		return fmt.Errorf("自动化动作「%s」的目标值 %s 不是有效数值", mount.AutomationName, mount.TargetValue)
	}

	row := make(map[string]interface{})
	if err := tx.Model(def.Model()).Select(def.PrimaryKey, mount.TargetField).
		Where(def.PrimaryKey+" = ? AND del_flag = 0", binding.BusinessID).First(row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("业务对象 %s 不存在或已删除,无法执行自动化动作「%s」", binding.BusinessID, mount.AutomationName)
		}
		return err
	}
	oldValueText := automationValueText(row[mount.TargetField])

	now := time.Now()
	result := tx.Model(def.Model()).
		Where(def.PrimaryKey+" = ? AND del_flag = 0", binding.BusinessID).
		Updates(map[string]interface{}{
			mount.TargetField: targetValue,
			"update_date":     &now,
		})
	if result.Error != nil {
		return fmt.Errorf("自动化动作「%s」更新业务字段失败: %w", mount.AutomationName, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("业务对象 %s 不存在或已删除,无法执行自动化动作「%s」", binding.BusinessID, mount.AutomationName)
	}

	oldLabel := automationDictLabel(tx, fieldDef.DictType, oldValueText)
	newLabel := automationDictLabel(tx, fieldDef.DictType, mount.TargetValue)
	changeRichText := fmt.Sprintf("流程 %s 节点「%s」自动化动作「%s」：%s %s → %s",
		instance.InstanceNo, nodeName, mount.AutomationName, fieldDef.Label, oldLabel, newLabel)
	return createChangeHistoryTx(tx, operatorID, binding.BusinessID, businessTypeInt, changeBehaviorAutomation, changeRichText)
}

// executeWorkflowAutomationInsertRecord 执行插入记录动作:
// 按映射取值(固定值/流程表单变量) -> 校验 -> 同事务插入动作业务类型对应的业务记录 -> 写变更记录与业务流程绑定。
// 插入目标即动作自身的业务类型;目标业务为需求时返回新需求 ID,由调用方在事务提交成功后链式自动发起需求流程。
func executeWorkflowAutomationInsertRecord(tx *gorm.DB, instance *models.WfProcessInstance, mount workflowAutomationMount, variables map[string]interface{}, operatorID, nodeName string) (string, error) {
	targetType := mount.BusinessType
	def, exists := getBusinessTypeDef(targetType)
	if !exists {
		return "", fmt.Errorf("自动化动作「%s」的目标业务类型 %s 未注册", mount.AutomationName, targetType)
	}
	insertFieldMap := make(map[string]businessTypeFieldDef, len(def.InsertFields))
	for _, field := range def.InsertFields {
		insertFieldMap[field.Field] = field
	}
	if len(mount.Mappings) == 0 {
		return "", fmt.Errorf("自动化动作「%s」没有字段映射,无法插入记录", mount.AutomationName)
	}

	record := make(map[string]interface{})
	mappedFields := make(map[string]bool, len(mount.Mappings))
	for _, mapping := range mount.Mappings {
		field, ok := insertFieldMap[mapping.Field]
		if !ok {
			return "", fmt.Errorf("自动化动作「%s」的目标字段 %s 不在业务类型[%s]的可插字段目录内", mount.AutomationName, mapping.Field, def.Label)
		}
		if mappedFields[field.Field] {
			return "", fmt.Errorf("自动化动作「%s」的目标字段 %s 重复映射", mount.AutomationName, field.Field)
		}
		mappedFields[field.Field] = true
		var value string
		switch mapping.SourceType {
		case models.InsertSourceFixed:
			value = mapping.Value
		case models.InsertSourceForm:
			value = workflowVariableString(variables, mapping.FormField)
			if value == "" {
				return "", fmt.Errorf("自动化动作「%s」:流程变量中不存在表单字段 %s 的值,无法填充 %s",
					mount.AutomationName, mapping.FormField, field.Label)
			}
		default:
			return "", fmt.Errorf("自动化动作「%s」的字段 %s 值来源 %s 不合法", mount.AutomationName, field.Field, mapping.SourceType)
		}
		if ref, isRef := insertRefValidators[field.Field]; isRef {
			if err := validateInsertRefExists(tx, field.Field, value, ref); err != nil {
				return "", fmt.Errorf("自动化动作「%s」: %w", mount.AutomationName, err)
			}
		}
		record[field.Field] = value
	}
	for _, field := range def.InsertFields {
		if field.Required && !mappedFields[field.Field] {
			return "", fmt.Errorf("自动化动作「%s」缺少必填字段 %s 的映射,无法插入%s记录", mount.AutomationName, field.Label, def.Label)
		}
	}

	now := time.Now()
	record[def.PrimaryKey] = utils.GenerateUUID()
	record["create_date"] = &now
	record["update_date"] = &now
	record["del_flag"] = 0
	if err := tx.Model(def.Model()).Create(record).Error; err != nil {
		return "", fmt.Errorf("自动化动作「%s」插入%s记录失败: %w", mount.AutomationName, def.Label, err)
	}
	newBusinessID, _ := record[def.PrimaryKey].(string)

	businessTypeInt, err := strconv.Atoi(targetType)
	if err != nil {
		return "", fmt.Errorf("目标业务类型 %s 不是有效数值", targetType)
	}
	changeRichText := fmt.Sprintf("流程 %s 节点「%s」自动化动作「%s」插入一条%s记录",
		instance.InstanceNo, nodeName, mount.AutomationName, def.Label)
	if err := createChangeHistoryTx(tx, operatorID, newBusinessID, businessTypeInt, changeBehaviorAutomation, changeRichText); err != nil {
		return "", err
	}
	// 来源绑定:新记录 <-> 触发插入的审批实例,需求详情关联流程列表据此展示来源审批。
	if err := createWorkflowBusinessInstance(tx, targetType, newBusinessID, instance.InstanceID, instance.DefinitionID, instance.StarterID); err != nil {
		return "", err
	}
	return newBusinessID, nil
}

// workflowVariableString 从流程变量读取字符串值,键不存在或值为空返回空串。
func workflowVariableString(variables map[string]interface{}, field string) string {
	value, exists := variables[field]
	if !exists || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

// automationValueText 将查询出的状态字段值统一转为文本,兼容驱动返回的 []byte、整数等类型。
func automationValueText(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

// automationDictLabel 查字典标签用于变更记录正文展示,查不到时回退为原始值。
func automationDictLabel(tx *gorm.DB, dictType, value string) string {
	if value == "" {
		return "空"
	}
	var dict models.SysDict
	if err := tx.Where("type = ? AND value = ? AND del_flag = 0", dictType, value).
		First(&dict).Error; err != nil {
		return value
	}
	if dict.Label == nil || *dict.Label == "" {
		return value
	}
	return *dict.Label
}

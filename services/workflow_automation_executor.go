package services

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/models"
)

// changeBehaviorAutomation 自动化动作变更行为,CHANGE_BEHAVIOR 字典值 60。
const changeBehaviorAutomation = 60

// executeWorkflowAutomations 执行节点挂载的自动化动作快照。
// 按画布挂载顺序在同一流程事务内执行,任一动作失败返回错误,整个流程流转回滚;
// 退回重走路径时经过的自动节点动作会再次执行,update_field 幂等(设固定值),重复执行无害。
// operatorID 为触发本次节点完成的操作人:审批节点取完成审批的办理人,自动节点取流程发起人,写入业务变更记录。
// nodeName 用于变更记录正文定位触发节点。
func executeWorkflowAutomations(tx *gorm.DB, instance *models.WfProcessInstance, mounts []workflowAutomationMount, operatorID, nodeName string) error {
	if len(mounts) == 0 {
		return nil
	}
	binding, err := getWorkflowBusinessInstanceByInstanceID(tx, instance.InstanceID)
	if err != nil {
		return fmt.Errorf("查询业务流程关联失败: %w", err)
	}
	if binding == nil {
		return fmt.Errorf("流程实例未绑定业务对象,无法执行自动化动作,请由业务对象发起本流程")
	}
	def, exists := getBusinessTypeDef(binding.BusinessType)
	if !exists {
		return fmt.Errorf("绑定的业务类型 %s 未注册,无法执行自动化动作", binding.BusinessType)
	}
	businessTypeInt, err := strconv.Atoi(binding.BusinessType)
	if err != nil {
		return fmt.Errorf("绑定的业务类型 %s 不是有效数值", binding.BusinessType)
	}
	for _, mount := range mounts {
		if err := executeWorkflowAutomationUpdateField(tx, instance, def, binding.BusinessID, businessTypeInt, mount, operatorID, nodeName); err != nil {
			return err
		}
	}
	return nil
}

// executeWorkflowAutomationUpdateField 执行修改字段值动作:
// 校验快照字段仍在白名单内 -> 读取旧值 -> 更新状态字段 -> 写业务变更记录(不写参与人、不推送通知)。
func executeWorkflowAutomationUpdateField(tx *gorm.DB, instance *models.WfProcessInstance, def businessTypeDef, businessID string, businessTypeInt int, mount workflowAutomationMount, operatorID, nodeName string) error {
	if mount.ActionType != automationActionTypeUpdateField {
		return fmt.Errorf("自动化动作「%s」的动作类型 %s 不受支持", mount.AutomationName, mount.ActionType)
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
		Where(def.PrimaryKey+" = ? AND del_flag = 0", businessID).First(row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("业务对象 %s 不存在或已删除,无法执行自动化动作「%s」", businessID, mount.AutomationName)
		}
		return err
	}
	oldValueText := automationValueText(row[mount.TargetField])

	now := time.Now()
	result := tx.Model(def.Model()).
		Where(def.PrimaryKey+" = ? AND del_flag = 0", businessID).
		Updates(map[string]interface{}{
			mount.TargetField: targetValue,
			"update_date":     &now,
		})
	if result.Error != nil {
		return fmt.Errorf("自动化动作「%s」更新业务字段失败: %w", mount.AutomationName, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("业务对象 %s 不存在或已删除,无法执行自动化动作「%s」", businessID, mount.AutomationName)
	}

	oldLabel := automationDictLabel(tx, fieldDef.DictType, oldValueText)
	newLabel := automationDictLabel(tx, fieldDef.DictType, mount.TargetValue)
	changeRichText := fmt.Sprintf("流程 %s 节点「%s」自动化动作「%s」：%s %s → %s",
		instance.InstanceNo, nodeName, mount.AutomationName, fieldDef.Label, oldLabel, newLabel)
	return createChangeHistoryTx(tx, operatorID, businessID, businessTypeInt, changeBehaviorAutomation, changeRichText)
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

package services

import (
	"fmt"
	"sort"

	"gorm.io/gorm"
	"hive-admin-go/models"
)

// BusinessStateHook 业务状态钩子接口。
// 流程引擎在节点完成时调用，业务模块实现并注册，用于把流程流转同步为业务对象状态。
type BusinessStateHook interface {
	// OnNodeCompleted 在节点完成时由引擎调用。
	// tx：流程流转所在事务，业务用同一个 tx 改状态，失败则整个流程流转回滚。
	// instance：流程实例（含定义快照）。
	// businessID：业务对象 ID（由引擎从 wf_business_instance 关联表查得）。
	// nodeBusinessKey：完成的节点业务键（流程节点上配的稳定语义标识）。
	// approved：true=节点通过，false=流程拒绝（当前预留，未实现拒绝分支调用）。
	OnNodeCompleted(tx *gorm.DB, instance *models.WfProcessInstance, businessID string, nodeBusinessKey string, approved bool) error
	// BusinessLabel 返回业务类型中文名,供流程设计器业务类型下拉展示。
	BusinessLabel() string
	// SupportedNodeKeys 返回本钩子支持的节点业务键定义,供流程设计器节点业务键下拉选择。
	// 返回顺序即设计器下拉展示顺序,需与 OnNodeCompleted 内部映射保持一致。
	SupportedNodeKeys() []models.BusinessNodeKeyDef
}

// businessHookRegistry 业务状态钩子注册表，按 business_type 索引。
var businessHookRegistry = map[string]BusinessStateHook{}

// RegisterBusinessHook 注册业务状态钩子。
// businessType 为流程定义声明的业务归属标识（如 story/bug/task）。
func RegisterBusinessHook(businessType string, hook BusinessStateHook) {
	businessHookRegistry[businessType] = hook
}

// InitBusinessHooks 初始化所有业务模块的状态钩子注册。在 main 启动时调用。
func InitBusinessHooks() {
	RegisterBusinessHook("story", &storyStateHook{})
}

// getBusinessHook 按 business_type 查找已注册的钩子。
func getBusinessHook(businessType string) (BusinessStateHook, bool) {
	hook, exists := businessHookRegistry[businessType]
	return hook, exists
}

// ListBusinessHookRegistry 返回所有已注册业务状态钩子的元数据。
// 供流程设计器加载业务类型和节点业务键下拉选项。返回的是钩子元数据,不涉及业务记录,无需数据权限校验。
func ListBusinessHookRegistry() *models.BusinessHookRegistryResponse {
	items := make([]models.BusinessHookRegistryItem, 0, len(businessHookRegistry))
	for businessType, hook := range businessHookRegistry {
		items = append(items, models.BusinessHookRegistryItem{
			BusinessType: businessType,
			Label:        hook.BusinessLabel(),
			NodeKeys:     hook.SupportedNodeKeys(),
		})
	}
	// 按 businessType 排序,保证设计器下拉选项顺序稳定
	sort.Slice(items, func(i, j int) bool {
		return items[i].BusinessType < items[j].BusinessType
	})
	return &models.BusinessHookRegistryResponse{Items: items}
}

// triggerBusinessStateHook 在节点完成时触发业务状态钩子。
// 流程引擎在 completeWorkflowNode 后调用本函数：
//  1. 反查 wf_business_instance 得 business_type + business_id；
//  2. 按 business_type 找已注册 hook；
//  3. 调 hook.OnNodeCompleted 用共享 tx 改业务对象状态。
//
// 未配置 nodeBusinessKey、无业务绑定或无 hook 注册时静默跳过（支持纯流程实例）。
func triggerBusinessStateHook(tx *gorm.DB, instance *models.WfProcessInstance, nodeBusinessKey string, approved bool) error {
	if nodeBusinessKey == "" {
		return nil
	}
	binding, err := getWorkflowBusinessInstanceByInstanceID(tx, instance.InstanceID)
	if err != nil {
		return fmt.Errorf("查询业务流程关联失败: %w", err)
	}
	if binding == nil {
		return nil
	}
	hook, exists := getBusinessHook(binding.BusinessType)
	if !exists {
		return nil
	}
	return hook.OnNodeCompleted(tx, instance, binding.BusinessID, nodeBusinessKey, approved)
}

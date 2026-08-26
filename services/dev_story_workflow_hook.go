package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// storyNodeStatusMap 需求流程节点业务键到需求状态值的映射。
// 节点业务键由流程设计器在节点属性中配置(nodeBusinessKey),与 STORY_STATUS 字典值对应。
// 当前映射覆盖需求全生命周期关键节点,未配置业务键或未命中映射的节点不触发状态变更。
// 注意:键集合需与 storyNodeKeyDefinitions 保持一致,新增节点键时两处同步更新。
var storyNodeStatusMap = map[string]int{
	"review":           10, // 评审通过 -> 已评审
	"ui_design":        20, // UI设计完成 -> UI设计中(交付给设计)
	"develop":          30, // 开发完成 -> 后端开发中(交付给后端)
	"develop_frontend": 31, // 前端开发完成 -> 前端开发中
	"test":             40, // 测试通过 -> 测试中
	"product_accept":   50, // 产品验收通过 -> 产品验收中
	"business_accept":  51, // 业务验收通过 -> 业务验收中
	"close":            99, // 关闭 -> 已关闭
}

// storyNodeKeyDefinitions 需求流程节点业务键定义,供流程设计器下拉选择展示。
// 顺序即设计器下拉展示顺序,键集合与 storyNodeStatusMap 保持一致。
var storyNodeKeyDefinitions = []models.BusinessNodeKeyDef{
	{NodeKey: "review", Label: "评审通过", Description: "节点通过后需求状态改为已评审"},
	{NodeKey: "ui_design", Label: "UI设计完成", Description: "节点通过后需求状态改为UI设计中(交付给设计)"},
	{NodeKey: "develop", Label: "开发完成", Description: "节点通过后需求状态改为后端开发中(交付给后端)"},
	{NodeKey: "develop_frontend", Label: "前端开发完成", Description: "节点通过后需求状态改为前端开发中"},
	{NodeKey: "test", Label: "测试通过", Description: "节点通过后需求状态改为测试中"},
	{NodeKey: "product_accept", Label: "产品验收通过", Description: "节点通过后需求状态改为产品验收中"},
	{NodeKey: "business_accept", Label: "业务验收通过", Description: "节点通过后需求状态改为业务验收中"},
	{NodeKey: "close", Label: "关闭", Description: "节点通过后需求状态改为已关闭"},
}

// storyStateHook 需求业务状态钩子实现。
// 流程引擎在审批节点完成后,按 businessType="story" 找到本钩子并调用 OnNodeCompleted,
// 把节点业务键映射为需求状态值并更新 dev_story.story_status。
type storyStateHook struct{}

// BusinessLabel 返回需求业务类型中文名,供流程设计器业务类型下拉展示。
func (h *storyStateHook) BusinessLabel() string {
	return "需求"
}

// SupportedNodeKeys 返回需求流程支持的节点业务键定义,供流程设计器节点业务键下拉选择。
func (h *storyStateHook) SupportedNodeKeys() []models.BusinessNodeKeyDef {
	return storyNodeKeyDefinitions
}

// GetBusinessSummary 返回需求摘要,供流程实例详情页展示关联业务。
// 详情页路径用 storyNum(与前端路由 /dev/story/detail/:storyNum 一致),便于 URL 直接传播。
func (h *storyStateHook) GetBusinessSummary(businessID string) (string, string, error) {
	var story models.DevStory
	if err := database.DB.Select("story_id", "story_title", "story_num").
		Where("story_id = ? AND del_flag = 0", businessID).
		First(&story).Error; err != nil {
		return "", "", fmt.Errorf("需求不存在或已删除: %w", err)
	}
	title := ""
	if story.StoryTitle != nil {
		title = *story.StoryTitle
	}
	return title, fmt.Sprintf("/dev/story/detail/%d", story.StoryNum), nil
}

// OnNodeCompleted 实现业务状态钩子契约。
// tx: 流程引擎共享事务,需求状态更新与流程流转同事务,失败回滚。
// nodeBusinessKey: 流程节点声明的业务键(如 "review"),未命中映射时静默跳过。
// approved: 当前只处理 true(节点通过);false 预留,暂不实现拒绝回退。
func (h *storyStateHook) OnNodeCompleted(tx *gorm.DB, instance *models.WfProcessInstance, businessID string, nodeBusinessKey string, approved bool) error {
	if !approved {
		return nil
	}
	targetStatus, ok := storyNodeStatusMap[nodeBusinessKey]
	if !ok {
		return nil
	}
	now := time.Now()
	result := tx.Model(&models.DevStory{}).
		Where("story_id = ? AND del_flag = 0", businessID).
		Updates(map[string]interface{}{
			"story_status": targetStatus,
			"update_date":  &now,
		})
	if result.Error != nil {
		return fmt.Errorf("更新需求状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("需求 %s 不存在或已删除,无法同步流程状态", businessID)
	}
	return nil
}

package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// workflowStoryLandingTitleField 落地创建必须由表单同名字段提供的需求名称(dev_story 非空字段)。
const workflowStoryLandingTitleField = "story_title"

// workflowStoryLandingStringFields 落地允许同名映射为字符串的需求字段白名单。
// 表单字段名有意与需求表字段名一致,白名单外的表单字段不写库。
const workflowStoryLandingStringFields = "story_title、story_rich_text、project_id、version_id、module_id"

// workflowStoryLandingIntFields 落地允许同名映射为整数的需求字典字段白名单。
const workflowStoryLandingIntFields = "story_type、story_level、source"

// landStoryFromWorkflow 流程实例审批通过后,在同一事务按表单同名字段映射落地创建一条规划中需求。
//
// 映射规则:表单字段名与需求表字段名同名时取值写入;表单未提供的字段不提供值
// (version_id/module_id/story_rich_text 存 NULL,project_id 存空串,字典字段沿用数据库默认 0)。
// story_title 为需求表非空字段,表单未提供或为空时报错,整个审批操作回滚。
// 同事务写入:dev_story、需求变更记录(正文留痕来源实例)、wf_business_instance 来源绑定。
// 返回新需求 ID,供事务提交后链式自动发起需求流程。
func landStoryFromWorkflow(tx *gorm.DB, instance *models.WfProcessInstance) (string, error) {
	variables := make(map[string]interface{})
	if strings.TrimSpace(instance.Variables) != "" {
		if err := json.Unmarshal([]byte(instance.Variables), &variables); err != nil {
			return "", fmt.Errorf("流程变量解析失败,无法落地创建需求")
		}
	}

	title, err := workflowLandingStringField(variables, "story_title")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("流程表单未提供 %s,无法落地创建需求", workflowStoryLandingTitleField)
	}
	richText, err := workflowLandingStringField(variables, "story_rich_text")
	if err != nil {
		return "", err
	}
	projectID, err := workflowLandingStringField(variables, "project_id")
	if err != nil {
		return "", err
	}
	versionID, err := workflowLandingOptionalRefField(variables, "version_id")
	if err != nil {
		return "", err
	}
	moduleID, err := workflowLandingOptionalRefField(variables, "module_id")
	if err != nil {
		return "", err
	}
	storyType, err := workflowLandingIntField(variables, "story_type")
	if err != nil {
		return "", err
	}
	storyLevel, err := workflowLandingIntField(variables, "story_level")
	if err != nil {
		return "", err
	}
	source, err := workflowLandingIntField(variables, "source")
	if err != nil {
		return "", err
	}

	if projectID != "" {
		if err := validateWorkflowLandingReference(tx, &models.DevProject{}, "project_id", projectID, "项目"); err != nil {
			return "", err
		}
	}
	if versionID != nil {
		if err := validateWorkflowLandingReference(tx, &models.DevVersion{}, "version_id", *versionID, "版本"); err != nil {
			return "", err
		}
	}
	if moduleID != nil {
		if err := validateWorkflowLandingReference(tx, &models.DevModule{}, "module_id", *moduleID, "模块"); err != nil {
			return "", err
		}
	}

	storyID := utils.GenerateUUID()
	now := time.Now()
	story := models.DevStory{
		StoryID:       storyID,
		StoryTitle:    &title,
		StoryType:     storyType,
		StoryStatus:   0, // 规划中,字典STORY_STATUS值
		StoryLevel:    storyLevel,
		Source:        source,
		ProjectID:     projectID,
		VersionID:     versionID,
		ModuleID:      moduleID,
		CreatorID:     &instance.StarterID,
		StoryRichText: &richText,
		CreateDate:    &now,
		UpdateDate:    &now,
		DelFlag:       0,
		StoryNum:      0,
	}
	if err := tx.Create(&story).Error; err != nil {
		return "", err
	}
	changeRichText := fmt.Sprintf("由流程实例 %s 审批通过后自动创建", instance.InstanceNo)
	if err := createChangeHistoryTx(tx, instance.StarterID, storyID, 0, 0, changeRichText); err != nil {
		return "", err
	}
	// 来源绑定:需求 <-> 触发落地创建的审批实例,供需求详情页关联流程列表追溯来源。
	if err := createWorkflowBusinessInstance(tx, "story", storyID, instance.InstanceID, instance.DefinitionID, instance.StarterID); err != nil {
		return "", err
	}
	return storyID, nil
}

// workflowLandingStringField 从流程变量读取同名字段的字符串值,字段未出现时返回空串。
// 可映射的字符串字段见 workflowStoryLandingStringFields,JSON 数字统一按整数字符串映射以对齐字典值。
func workflowLandingStringField(variables map[string]interface{}, field string) (string, error) {
	value, exists := variables[field]
	if !exists || value == nil {
		return "", nil
	}
	switch typed := value.(type) {
	case string:
		return typed, nil
	case float64:
		if typed != float64(int64(typed)) {
			return "", fmt.Errorf("流程表单字段 %s 的数值不是整数,无法映射需求字段", field)
		}
		return strconv.FormatInt(int64(typed), 10), nil
	default:
		return "", fmt.Errorf("流程表单字段 %s 的类型无法映射需求字段", field)
	}
}

// workflowLandingOptionalRefField 从流程变量读取可选引用字段,空值返回 nil(落库为 NULL)。
func workflowLandingOptionalRefField(variables map[string]interface{}, field string) (*string, error) {
	value, err := workflowLandingStringField(variables, field)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return &value, nil
}

// workflowLandingIntField 从流程变量读取同名字典字段,字段未出现时返回 0(沿用数据库默认值)。
// 可映射的整数字段见 workflowStoryLandingIntFields。
func workflowLandingIntField(variables map[string]interface{}, field string) (int, error) {
	value, err := workflowLandingStringField(variables, field)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("流程表单字段 %s 的值 %q 不是有效的字典数值", field, value)
	}
	return number, nil
}

// validateWorkflowLandingReference 校验表单提供的引用字段指向未删除记录,防止落地产生悬空引用。
func validateWorkflowLandingReference(tx *gorm.DB, model interface{}, column, value, label string) error {
	var count int64
	if err := tx.Model(model).Where(column+" = ? AND del_flag = 0", value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("流程表单提供的 %s(%s)不存在或已删除,无法落地创建需求", column, label)
	}
	return nil
}

// flushWorkflowLandingAutoStart 在审批事务提交成功后,为本次落地创建的需求链式自动发起需求流程。
// 复用 autoStartStoryWorkflow 的宽松语义:未匹配到已发布 story 流程定义或发起失败时仅记日志,
// 不影响已通过的审批和已创建的需求。
func flushWorkflowLandingAutoStart(context *workflowExecutionContext) {
	if context == nil || len(context.pendingAutoStartStories) == 0 {
		return
	}
	starterID := context.instance.StarterID
	for _, storyID := range context.pendingAutoStartStories {
		autoStartStoryWorkflow(storyID, starterID)
	}
}

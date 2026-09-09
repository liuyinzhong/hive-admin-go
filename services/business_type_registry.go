package services

import (
	"fmt"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// businessTypeFieldDef 业务类型字段元数据:状态字段白名单与可插字段目录共用。
type businessTypeFieldDef struct {
	Field    string // 数据库列名,如 story_status
	Label    string // 中文名,如 需求状态
	DictType string // 值字典类型,如 STORY_STATUS;空表示非字典字段
	Required bool   // 可插字段目录中的必填标记(表级非空字段)
}

// insertRefDef 插入记录引用字段的校验定义:校验引用记录存在且未删除。
type insertRefDef struct {
	Model interface{}
	Label string
}

// insertRefValidators 插入记录引用字段的统一校验表(引用目标表与业务类型无关,全局共享)。
// 引用表主键列与字段名一致,统一按 字段名 = 值 AND del_flag = 0 校验。
var insertRefValidators = map[string]insertRefDef{
	"project_id": {Model: &models.DevProject{}, Label: "项目"},
	"version_id": {Model: &models.DevVersion{}, Label: "版本"},
	"module_id":  {Model: &models.DevModule{}, Label: "模块"},
	"story_id":   {Model: &models.DevStory{}, Label: "需求"},
}

// businessTypeDef 业务类型注册项。
// businessType 统一使用 BUSINESS_TYPE 字典值(0需求/10任务/20缺陷/30版本),
// 贯穿流程定义声明、动作库配置、业务流程绑定与执行器定位。
type businessTypeDef struct {
	Type       string
	Label      string
	Model      func() interface{}
	PrimaryKey string
	// StatusFields 修改字段值动作的可写字段白名单(仅状态字段)。
	StatusFields []businessTypeFieldDef
	// InsertFields 插入记录动作的可插字段目录(含必填标记);目录外字段不可插入。
	InsertFields []businessTypeFieldDef
	// DetailSummary 返回业务对象摘要(标题+前端详情页路径),供流程实例详情页展示关联业务。
	// 任务/缺陷等抽屉型详情暂无独立路由,返回空路径。
	DetailSummary func(tx *gorm.DB, businessID string) (title string, detailPath string, err error)
}

// businessTypeRegistry 业务类型注册表。新增业务类型时在此登记,
// 动作库配置、设计器过滤、执行器与关联业务摘要都以此表为唯一来源。
var businessTypeRegistry = map[string]businessTypeDef{
	"0": {
		Type:         "0",
		Label:        "需求",
		Model:        func() interface{} { return &models.DevStory{} },
		PrimaryKey:   "story_id",
		StatusFields: []businessTypeFieldDef{{Field: "story_status", Label: "需求状态", DictType: "STORY_STATUS"}},
		InsertFields: []businessTypeFieldDef{
			{Field: "story_title", Label: "需求名称", Required: true},
			{Field: "story_rich_text", Label: "需求描述"},
			{Field: "story_status", Label: "需求状态", DictType: "STORY_STATUS"},
			{Field: "story_type", Label: "需求类型", DictType: "STORY_TYPE"},
			{Field: "story_level", Label: "需求优先级", DictType: "STORY_LEVEL"},
			{Field: "source", Label: "需求来源", DictType: "STORY_SOURCE"},
			{Field: "project_id", Label: "所属项目"},
			{Field: "version_id", Label: "关联版本"},
			{Field: "module_id", Label: "关联模块"},
		},
		DetailSummary: func(tx *gorm.DB, businessID string) (string, string, error) {
			var story models.DevStory
			if err := tx.Select("story_id", "story_title", "story_num").
				Where("story_id = ? AND del_flag = 0", businessID).First(&story).Error; err != nil {
				return "", "", fmt.Errorf("需求不存在或已删除: %w", err)
			}
			title := ""
			if story.StoryTitle != nil {
				title = *story.StoryTitle
			}
			return title, fmt.Sprintf("/dev/story/detail/%d", story.StoryNum), nil
		},
	},
	"10": {
		Type:         "10",
		Label:        "任务",
		Model:        func() interface{} { return &models.DevTask{} },
		PrimaryKey:   "task_id",
		StatusFields: []businessTypeFieldDef{{Field: "task_status", Label: "任务状态", DictType: "TASK_STATUS"}},
		InsertFields: []businessTypeFieldDef{
			{Field: "task_title", Label: "任务名称", Required: true},
			{Field: "project_id", Label: "所属项目", Required: true},
			{Field: "task_rich_text", Label: "任务描述"},
			{Field: "task_status", Label: "任务状态", DictType: "TASK_STATUS"},
			{Field: "task_type", Label: "任务类型", DictType: "TASK_TYPE"},
		},
		DetailSummary: func(tx *gorm.DB, businessID string) (string, string, error) {
			return businessTypeRowTitle(tx, &models.DevTask{}, "task_id", "task_title", businessID, "任务")
		},
	},
	"20": {
		Type:       "20",
		Label:      "缺陷",
		Model:      func() interface{} { return &models.DevBug{} },
		PrimaryKey: "bug_id",
		StatusFields: []businessTypeFieldDef{
			{Field: "bug_status", Label: "缺陷状态", DictType: "BUG_STATUS"},
			{Field: "bug_confirm_status", Label: "缺陷确认状态", DictType: "BUG_CONFIRM_STATUS"},
		},
		InsertFields: []businessTypeFieldDef{
			{Field: "bug_title", Label: "缺陷名称", Required: true},
			{Field: "bug_rich_text", Label: "缺陷描述"},
			{Field: "bug_env", Label: "缺陷环境", DictType: "BUG_ENV"},
			{Field: "bug_status", Label: "缺陷状态", DictType: "BUG_STATUS"},
			{Field: "bug_confirm_status", Label: "缺陷确认状态", DictType: "BUG_CONFIRM_STATUS"},
			{Field: "bug_level", Label: "缺陷优先级", DictType: "BUG_LEVEL"},
		},
		DetailSummary: func(tx *gorm.DB, businessID string) (string, string, error) {
			return businessTypeRowTitle(tx, &models.DevBug{}, "bug_id", "bug_title", businessID, "缺陷")
		},
	},
	"30": {
		Type:         "30",
		Label:        "版本",
		Model:        func() interface{} { return &models.DevVersion{} },
		PrimaryKey:   "version_id",
		StatusFields: []businessTypeFieldDef{{Field: "release_status", Label: "发布状态", DictType: "RELEASE_STATUS"}},
		InsertFields: []businessTypeFieldDef{
			{Field: "version", Label: "版本号", Required: true},
			{Field: "project_id", Label: "所属项目", Required: true},
			{Field: "remark", Label: "版本备注"},
			{Field: "release_status", Label: "发布状态", DictType: "RELEASE_STATUS"},
		},
		DetailSummary: func(tx *gorm.DB, businessID string) (string, string, error) {
			return businessTypeRowTitle(tx, &models.DevVersion{}, "version_id", "version", businessID, "版本")
		},
	},
}

// businessTypeRowTitle 通用摘要:按主键查标题字段,无独立详情路由时路径返回空。
func businessTypeRowTitle(tx *gorm.DB, model interface{}, primaryKey, titleColumn, businessID, label string) (string, string, error) {
	row := make(map[string]interface{})
	if err := tx.Model(model).Select(titleColumn).
		Where(primaryKey+" = ? AND del_flag = 0", businessID).First(row).Error; err != nil {
		return "", "", fmt.Errorf("%s不存在或已删除: %w", label, err)
	}
	title := ""
	if value, ok := row[titleColumn].([]byte); ok {
		title = string(value)
	} else if value, ok := row[titleColumn].(string); ok {
		title = value
	}
	return title, "", nil
}

// getBusinessTypeDef 按业务类型(字典值)查注册项。
func getBusinessTypeDef(businessType string) (businessTypeDef, bool) {
	def, exists := businessTypeRegistry[businessType]
	return def, exists
}

// GetBusinessTypeFieldMetas 返回指定业务类型的可写字段元数据,供动作库表单下拉。
func GetBusinessTypeFieldMetas(businessType string) ([]models.AutomationFieldMeta, error) {
	def, exists := getBusinessTypeDef(businessType)
	if !exists {
		return nil, fmt.Errorf("业务类型未注册: %s", businessType)
	}
	metas := make([]models.AutomationFieldMeta, 0, len(def.StatusFields))
	for _, field := range def.StatusFields {
		metas = append(metas, models.AutomationFieldMeta{Field: field.Field, Label: field.Label, DictType: field.DictType})
	}
	return metas, nil
}

// GetBusinessTypeInsertFieldMetas 返回指定业务类型的可插字段目录(含必填标记),供插入记录动作的字段映射配置。
func GetBusinessTypeInsertFieldMetas(businessType string) ([]models.AutomationFieldMeta, error) {
	def, exists := getBusinessTypeDef(businessType)
	if !exists {
		return nil, fmt.Errorf("业务类型未注册: %s", businessType)
	}
	metas := make([]models.AutomationFieldMeta, 0, len(def.InsertFields))
	for _, field := range def.InsertFields {
		metas = append(metas, models.AutomationFieldMeta{Field: field.Field, Label: field.Label, DictType: field.DictType, Required: field.Required})
	}
	return metas, nil
}

// ListBusinessTypeMetas 返回全部已注册业务类型及可写字段,供动作库新建表单联动。
func ListBusinessTypeMetas() []models.AutomationFieldMeta {
	metas := make([]models.AutomationFieldMeta, 0)
	for _, def := range businessTypeRegistry {
		for _, field := range def.StatusFields {
			metas = append(metas, models.AutomationFieldMeta{Field: field.Field, Label: def.Label + "·" + field.Label, DictType: field.DictType})
		}
	}
	return metas
}

// buildBusinessTypeSummary 流程实例详情页"关联业务"摘要,按绑定业务类型从注册表定位。
// 注册表未登记的类型降级为仅展示业务类型和ID,不阻塞详情加载。
func buildBusinessTypeSummary(businessType, businessID string) (label, title, detailPath string) {
	def, exists := getBusinessTypeDef(businessType)
	if !exists {
		return businessType, "", ""
	}
	title, path, err := def.DetailSummary(database.DB, businessID)
	if err != nil {
		return def.Label, "", ""
	}
	return def.Label, title, path
}

// buildWorkflowBusinessSummary 组装流程实例关联业务摘要。
// 供流程实例详情页展示"关联业务"入口:纯流程实例返回 nil;
// 业务类型未注册或业务对象已删除时降级为仅展示业务类型和ID,不影响详情页整体加载。
func buildWorkflowBusinessSummary(instanceID string) *models.WorkflowBusinessSummaryResponse {
	binding, err := getWorkflowBusinessInstanceByInstanceID(database.DB, instanceID)
	if err != nil || binding == nil {
		return nil
	}
	summary := &models.WorkflowBusinessSummaryResponse{
		BusinessType: binding.BusinessType,
		BusinessID:   binding.BusinessID,
	}
	summary.BusinessLabel, summary.BusinessTitle, summary.DetailPath = buildBusinessTypeSummary(binding.BusinessType, binding.BusinessID)
	return summary
}

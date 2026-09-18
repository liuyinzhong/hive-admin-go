package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/models"
)

// changeFieldKind 变更明细字段的取值方式,决定新旧值的翻译与展示规则。
type changeFieldKind int

const (
	changeValueDict     changeFieldKind = iota // 字典值:存原始值携带 dictType,前端按字典翻译
	changeValueText                            // 普通文本:原样记录
	changeValueDate                            // 日期时间:统一格式化为 YYYY-MM-DD HH:mm:ss
	changeValueRichText                        // 富文本:只记"已更新",不记内容
	changeValueUser                            // 用户引用:写入时翻译为姓名固化
	changeValueVersion                         // 版本引用:写入时翻译为版本号固化
	changeValueModule                          // 模块引用:写入时翻译为模块名固化
	changeValueStory                           // 需求引用:写入时翻译为需求标题固化
	changeValueProject                         // 项目引用:写入时翻译为项目标题固化
	changeValueFileIDs                         // 附件ID列表:比对增删并翻译为文件名固化
)

// changeFieldDef 变更明细字段目录的注册项。
type changeFieldDef struct {
	Field    string          // 数据库列名,如 story_status
	Label    string          // 字段中文标签,写入时刻固化,如 需求状态
	DictType string          // 值字典类型,如 STORY_STATUS;空表示非字典值
	Kind     changeFieldKind // 取值方式
}

// changeFieldCatalog 变更明细字段目录,按业务类型(BUSINESS_TYPE 字典值)注册;
// 目录外字段不记明细,业务新增可变更字段时在此登记。
var changeFieldCatalog = map[int][]changeFieldDef{
	0: { // 需求
		{Field: "story_title", Label: "需求名称", Kind: changeValueText},
		{Field: "story_rich_text", Label: "需求描述", Kind: changeValueRichText},
		{Field: "story_type", Label: "需求类型", DictType: "STORY_TYPE", Kind: changeValueDict},
		{Field: "story_status", Label: "需求状态", DictType: "STORY_STATUS", Kind: changeValueDict},
		{Field: "story_level", Label: "需求优先级", DictType: "STORY_LEVEL", Kind: changeValueDict},
		{Field: "source", Label: "需求来源", DictType: "STORY_SOURCE", Kind: changeValueDict},
		{Field: "project_id", Label: "所属项目", Kind: changeValueProject},
		{Field: "version_id", Label: "关联版本", Kind: changeValueVersion},
		{Field: "module_id", Label: "关联模块", Kind: changeValueModule},
		{Field: "file_ids", Label: "附件", Kind: changeValueFileIDs},
	},
	10: { // 任务
		{Field: "task_title", Label: "任务名称", Kind: changeValueText},
		{Field: "task_rich_text", Label: "任务描述", Kind: changeValueRichText},
		{Field: "task_status", Label: "任务状态", DictType: "TASK_STATUS", Kind: changeValueDict},
		{Field: "task_type", Label: "任务类型", DictType: "TASK_TYPE", Kind: changeValueDict},
		{Field: "user_id", Label: "负责人", Kind: changeValueUser},
		{Field: "plan_hours", Label: "预估工时", Kind: changeValueText},
		{Field: "actual_hours", Label: "实际工时", Kind: changeValueText},
		{Field: "start_date", Label: "开始日期", Kind: changeValueDate},
		{Field: "end_date", Label: "结束日期", Kind: changeValueDate},
		{Field: "project_id", Label: "所属项目", Kind: changeValueProject},
		{Field: "version_id", Label: "关联版本", Kind: changeValueVersion},
		{Field: "module_id", Label: "关联模块", Kind: changeValueModule},
		{Field: "story_id", Label: "关联需求", Kind: changeValueStory},
	},
	20: { // 缺陷
		{Field: "bug_title", Label: "缺陷名称", Kind: changeValueText},
		{Field: "bug_rich_text", Label: "缺陷描述", Kind: changeValueRichText},
		{Field: "bug_status", Label: "缺陷状态", DictType: "BUG_STATUS", Kind: changeValueDict},
		{Field: "bug_confirm_status", Label: "缺陷确认状态", DictType: "BUG_CONFIRM_STATUS", Kind: changeValueDict},
		{Field: "bug_level", Label: "缺陷优先级", DictType: "BUG_LEVEL", Kind: changeValueDict},
		{Field: "bug_env", Label: "缺陷环境", DictType: "BUG_ENV", Kind: changeValueDict},
		{Field: "bug_source", Label: "缺陷来源", DictType: "BUG_SOURCE", Kind: changeValueDict},
		{Field: "bug_type", Label: "缺陷类型", DictType: "BUG_TYPE", Kind: changeValueDict},
		{Field: "fix_user_id", Label: "修复人", Kind: changeValueUser},
		{Field: "verifier_id", Label: "验证人", Kind: changeValueUser},
		{Field: "project_id", Label: "所属项目", Kind: changeValueProject},
		{Field: "version_id", Label: "关联版本", Kind: changeValueVersion},
		{Field: "module_id", Label: "关联模块", Kind: changeValueModule},
		{Field: "story_id", Label: "关联需求", Kind: changeValueStory},
		{Field: "file_ids", Label: "附件", Kind: changeValueFileIDs},
	},
}

// buildChangeItems 按变更明细目录对比新旧值生成字段级变更明细。
// oldValues/newValues 的键为数据库列名,值须为非 nil 的标量或指针(指针 nil 表示空值);
// newValues 中未提供的字段视为未修改,目录外字段与值未变化的字段不产生明细。
func buildChangeItems(tx *gorm.DB, businessType int, oldValues, newValues map[string]interface{}) []models.ChangeItem {
	catalog, exists := changeFieldCatalog[businessType]
	if !exists {
		return nil
	}
	items := make([]models.ChangeItem, 0)
	for _, def := range catalog {
		newText, newOK := normalizeChangeValue(newValues[def.Field])
		if !newOK {
			continue
		}
		oldText, _ := normalizeChangeValue(oldValues[def.Field])

		if def.Kind == changeValueFileIDs {
			if item, ok := buildFileIDsChangeItem(tx, def, oldText, newText); ok {
				items = append(items, item)
			}
			continue
		}
		if oldText == newText {
			continue
		}

		item := models.ChangeItem{
			FieldKey:   def.Field,
			FieldLabel: def.Label,
			DictType:   def.DictType,
		}
		switch def.Kind {
		case changeValueRichText:
			// 富文本不做内容级 diff,只记一条"已更新",避免时间线被长文本淹没
			item.NewValue = "已更新"
		case changeValueUser, changeValueVersion, changeValueModule, changeValueStory, changeValueProject:
			// 引用值在写入时刻翻译为可读文本固化,后续引用对象改名不影响历史明细可读性
			item.OldValue = changeRefText(tx, def.Kind, oldText)
			item.NewValue = changeRefText(tx, def.Kind, newText)
		default:
			item.OldValue = changeDisplayText(oldText)
			item.NewValue = changeDisplayText(newText)
		}
		items = append(items, item)
	}
	return items
}

// normalizeChangeValue 将变更字段的旧值或新值归一化为文本。
// 值未提供(键不存在或为 nil 接口)时第二个返回值为 false;
// 兼容业务 struct 字段、请求值与 GORM 查询返回的原始值。
func normalizeChangeValue(value interface{}) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case int:
		return strconv.Itoa(typed), true
	case int64:
		return strconv.FormatInt(typed, 10), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case time.Time:
		return typed.Format("2006-01-02 15:04:05"), true
	case *time.Time:
		if typed == nil {
			return "", true
		}
		return typed.Format("2006-01-02 15:04:05"), true
	case *string:
		if typed == nil {
			return "", true
		}
		return *typed, true
	default:
		return fmt.Sprintf("%v", typed), true
	}
}

// changeDisplayText 展示空值统一为"空"。
func changeDisplayText(text string) string {
	if text == "" {
		return "空"
	}
	return text
}

// changeRefText 将引用值翻译为可读文本(人名/版本号/模块名/需求标题/项目标题);
// 空值显示"空",引用记录查不到时回退为原始值,保证明细不因翻译失败而丢失。
func changeRefText(tx *gorm.DB, kind changeFieldKind, value string) string {
	if value == "" {
		return "空"
	}
	label := ""
	switch kind {
	case changeValueUser:
		var user models.SysUser
		if err := tx.Select("real_name").Where("user_id = ? AND del_flag = 0", value).First(&user).Error; err == nil && user.RealName != nil {
			label = *user.RealName
		}
	case changeValueVersion:
		label = changeColumnText(tx, &models.DevVersion{}, "version_id", "version", value)
	case changeValueModule:
		label = changeColumnText(tx, &models.DevModule{}, "module_id", "module_title", value)
	case changeValueStory:
		label = changeColumnText(tx, &models.DevStory{}, "story_id", "story_title", value)
	case changeValueProject:
		label = changeColumnText(tx, &models.DevProject{}, "project_id", "project_title", value)
	}
	if label == "" {
		return value
	}
	return label
}

// changeColumnText 按主键查询单列文本;记录不存在或已删除时返回空串,由调用方回退为原始值。
func changeColumnText(tx *gorm.DB, model interface{}, primaryKey, column, value string) string {
	row := make(map[string]interface{})
	if err := tx.Model(model).Select(column).
		Where(primaryKey+" = ? AND del_flag = 0", value).First(row).Error; err != nil {
		return ""
	}
	if text, ok := row[column].([]byte); ok {
		return string(text)
	}
	if text, ok := row[column].(string); ok {
		return text
	}
	return ""
}

// buildFileIDsChangeItem 对比附件ID列表差异,将新增/移除的附件翻译为文件名生成一条附件明细;
// 附件集合无变化时不产生明细。旧值侧记录移除的文件,新值侧记录新增的文件。
func buildFileIDsChangeItem(tx *gorm.DB, def changeFieldDef, oldText, newText string) (models.ChangeItem, bool) {
	oldIDs := changeFileIDList(oldText)
	newIDs := changeFileIDList(newText)
	oldSet := make(map[string]bool, len(oldIDs))
	for _, id := range oldIDs {
		oldSet[id] = true
	}
	newSet := make(map[string]bool, len(newIDs))
	for _, id := range newIDs {
		newSet[id] = true
	}

	removedIDs := make([]string, 0)
	for _, id := range oldIDs {
		if !newSet[id] {
			removedIDs = append(removedIDs, id)
		}
	}
	addedIDs := make([]string, 0)
	for _, id := range newIDs {
		if !oldSet[id] {
			addedIDs = append(addedIDs, id)
		}
	}
	if len(addedIDs) == 0 && len(removedIDs) == 0 {
		return models.ChangeItem{}, false
	}

	return models.ChangeItem{
		FieldKey:   def.Field,
		FieldLabel: def.Label,
		OldValue:   changeFileSummary(tx, "移除", removedIDs),
		NewValue:   changeFileSummary(tx, "新增", addedIDs),
	}, true
}

// changeFileIDList 解析逗号分隔的附件ID串为有序列表,自动忽略空值。
func changeFileIDList(text string) []string {
	if text == "" {
		return nil
	}
	parts := strings.Split(text, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		if id := strings.TrimSpace(part); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// storyChangeValues 从需求记录提取变更明细对比用的字段值(键为数据库列名)。
func storyChangeValues(story *models.DevStory) map[string]interface{} {
	return map[string]interface{}{
		"story_title":     story.StoryTitle,
		"story_type":      story.StoryType,
		"story_status":    story.StoryStatus,
		"story_level":     story.StoryLevel,
		"source":          story.Source,
		"project_id":      story.ProjectID,
		"version_id":      story.VersionID,
		"module_id":       story.ModuleID,
		"story_rich_text": story.StoryRichText,
		"file_ids":        story.FileIDs,
	}
}

// taskChangeValues 从任务记录提取变更明细对比用的字段值(键为数据库列名)。
func taskChangeValues(task *models.DevTask) map[string]interface{} {
	return map[string]interface{}{
		"task_title":     task.TaskTitle,
		"task_rich_text": task.TaskRichText,
		"task_status":    task.TaskStatus,
		"task_type":      task.TaskType,
		"user_id":        task.UserID,
		"plan_hours":     task.PlanHours,
		"actual_hours":   task.ActualHours,
		"start_date":     task.StartDate,
		"end_date":       task.EndDate,
		"project_id":     task.ProjectID,
		"version_id":     task.VersionID,
		"module_id":      task.ModuleID,
		"story_id":       task.StoryID,
	}
}

// bugChangeValues 从缺陷记录提取变更明细对比用的字段值(键为数据库列名)。
func bugChangeValues(bug *models.DevBug) map[string]interface{} {
	return map[string]interface{}{
		"bug_title":          bug.BugTitle,
		"bug_rich_text":      bug.BugRichText,
		"bug_status":         bug.BugStatus,
		"bug_confirm_status": bug.BugConfirmStatus,
		"bug_level":          bug.BugLevel,
		"bug_env":            bug.BugEnv,
		"bug_source":         bug.BugSource,
		"bug_type":           bug.BugType,
		"fix_user_id":        bug.FixUserID,
		"verifier_id":        bug.VerifierID,
		"project_id":         bug.ProjectID,
		"version_id":         bug.VersionID,
		"module_id":          bug.ModuleID,
		"story_id":           bug.StoryID,
		"file_ids":           bug.FileIDs,
	}
}

// changeFileSummary 将增/删的附件ID翻译为文件名,拼成"新增：a.txt、b.png"形式;
// 文件记录查不到时回退为文件ID展示,避免信息丢失;列表为空时返回空串。
func changeFileSummary(tx *gorm.DB, action string, fileIDs []string) string {
	if len(fileIDs) == 0 {
		return ""
	}
	var files []models.SysFile
	if err := tx.Select("file_id", "name").Where("file_id IN ?", fileIDs).Find(&files).Error; err != nil {
		return action + "：" + strings.Join(fileIDs, "、")
	}
	nameMap := make(map[string]string, len(files))
	for _, file := range files {
		name := file.FileID
		if file.Name != nil && *file.Name != "" {
			name = *file.Name
		}
		nameMap[file.FileID] = name
	}
	names := make([]string, 0, len(fileIDs))
	for _, id := range fileIDs {
		if name, ok := nameMap[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return action + "：" + strings.Join(names, "、")
}

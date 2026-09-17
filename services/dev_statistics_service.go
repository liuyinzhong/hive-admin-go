package services

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
)

// GetTaskFindDay 统计两个日期各小时任务创建数量，用于任务趋势对比折线图
func GetTaskFindDay(date1Str, date2Str string, permission datapermission.Permission) (*models.TaskFindDayResponse, error) {
	date1, err := parseStatisticsDate(date1Str)
	if err != nil {
		return nil, err
	}
	date2, err := parseStatisticsDate(date2Str)
	if err != nil {
		return nil, err
	}

	date1Data, err := countTaskByHour(date1, permission)
	if err != nil {
		return nil, err
	}
	date2Data, err := countTaskByHour(date2, permission)
	if err != nil {
		return nil, err
	}

	return &models.TaskFindDayResponse{
		Date1: date1Data,
		Date2: date2Data,
	}, nil
}

// GetTaskFindYear 统计指定年份各月份任务实际工时合计，用于工时总量柱状图
func GetTaskFindYear(year int, permission datapermission.Permission) (*models.TaskFindYearResponse, error) {
	if year <= 0 {
		return nil, fmt.Errorf("年份参数错误")
	}

	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.Local)

	type monthSum struct {
		Month int     `gorm:"column:month"`
		Total float64 `gorm:"column:total"`
	}
	var results []monthSum
	query := database.DB.Model(&models.DevTask{}).
		Select("MONTH(create_date) as month, COALESCE(SUM(actual_hours), 0) as total").
		Where("dev_task.del_flag = ? AND create_date >= ? AND create_date < ?", 0, yearStart, yearEnd)
	err := permission.Apply(query, "dev_task.creator_id", "dev_task.user_id").
		Group("MONTH(create_date)").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	list := make([]float64, 12)
	for _, r := range results {
		if r.Month >= 1 && r.Month <= 12 {
			list[r.Month-1] = r.Total
		}
	}
	return &models.TaskFindYearResponse{List: list}, nil
}

// GetWorkspaceEnum 统计需求、任务、缺陷的总数与待处理数量，用于工作台概览
func GetWorkspaceEnum(permission datapermission.Permission) (*models.WorkspaceEnumResponse, error) {
	var storyTotal, storyActive int64
	var taskTotal, taskActive int64
	var bugTotal, bugActive int64

	storyQuery := applyStoryPermission(database.DB.Model(&models.DevStory{}), permission)
	taskQuery := permission.Apply(database.DB.Model(&models.DevTask{}), "dev_task.creator_id", "dev_task.user_id")
	bugQuery := permission.Apply(database.DB.Model(&models.DevBug{}), "dev_bug.creator_id", "dev_bug.fix_user_id")

	if err := storyQuery.Where("dev_story.del_flag = ?", 0).Count(&storyTotal).Error; err != nil {
		return nil, err
	}
	if err := storyQuery.Session(&gorm.Session{}).Where("dev_story.del_flag = ? AND story_status = ?", 0, 0).Count(&storyActive).Error; err != nil {
		return nil, err
	}
	if err := taskQuery.Where("dev_task.del_flag = ?", 0).Count(&taskTotal).Error; err != nil {
		return nil, err
	}
	if err := taskQuery.Session(&gorm.Session{}).Where("dev_task.del_flag = ? AND task_status = ?", 0, 0).Count(&taskActive).Error; err != nil {
		return nil, err
	}
	if err := bugQuery.Where("dev_bug.del_flag = ?", 0).Count(&bugTotal).Error; err != nil {
		return nil, err
	}
	if err := bugQuery.Session(&gorm.Session{}).Where("dev_bug.del_flag = ? AND bug_status = ?", 0, 0).Count(&bugActive).Error; err != nil {
		return nil, err
	}

	return &models.WorkspaceEnumResponse{
		StoryTotalNum: storyTotal,
		StoryNum:      storyActive,
		TaskTotalNum:  taskTotal,
		TaskNum:       taskActive,
		BugTotalNum:   bugTotal,
		BugNum:        bugActive,
	}, nil
}

// parseStatisticsDate 解析统计接口的日期参数，兼容 YYYY/MM/DD 与 YYYY-MM-DD 格式
func parseStatisticsDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("日期不能为空")
	}
	normalized := strings.ReplaceAll(dateStr, "/", "-")
	t, err := time.ParseInLocation("2006-01-02", normalized, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期格式错误，请使用 YYYY/MM/DD 格式")
	}
	return t, nil
}

// countTaskByHour 统计指定日期各小时（0-23）的任务创建数量
func countTaskByHour(date time.Time, permission datapermission.Permission) ([]int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)

	type hourCount struct {
		Hour  int   `gorm:"column:hour"`
		Count int64 `gorm:"column:count"`
	}
	var results []hourCount
	query := database.DB.Model(&models.DevTask{}).
		Select("HOUR(create_date) as hour, COUNT(*) as count").
		Where("dev_task.del_flag = ? AND create_date >= ? AND create_date < ?", 0, start, end)
	err := permission.Apply(query, "dev_task.creator_id", "dev_task.user_id").
		Group("HOUR(create_date)").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	data := make([]int64, 24)
	for _, r := range results {
		if r.Hour >= 0 && r.Hour < 24 {
			data[r.Hour] = r.Count
		}
	}
	return data, nil
}

// 版本统计的完成态字典值：需求已关闭=99、任务已完成=99、缺陷已修复=20。
const (
	versionStoryDoneStatus = 99
	versionTaskDoneStatus  = 99
	versionBugFixedStatus  = 20
)

// distRow 通用分组统计原始行：key 为分组列原值（数字字典值或用户/模块ID），num 为聚合计数。
type distRow struct {
	Key string  `gorm:"column:k"`
	Num float64 `gorm:"column:num"`
}

// workloadRow 执行人工时聚合行。
type workloadRow struct {
	UserID      string  `gorm:"column:user_id"`
	PlanHours   float64 `gorm:"column:plan_hours"`
	ActualHours float64 `gorm:"column:actual_hours"`
}

// trendRow 按日期聚合完成数行。
type trendRow struct {
	Date  string `gorm:"column:date"`
	Count int64  `gorm:"column:count"`
}

// GetVersionStatistics 统计指定版本下需求、任务、缺陷的概览、完成趋势与多维分布。
// 所有聚合都先应用当前角色数据范围（需求按创建人/参与人、任务按创建人/执行人、缺陷按创建人/修复人）再按 version_id 过滤。
func GetVersionStatistics(versionID string, permission datapermission.Permission) (*models.VersionStatisticsResponse, error) {
	if versionID == "" {
		return nil, fmt.Errorf("versionId不能为空")
	}

	// 校验版本存在且在当前数据范围内，同时取版本周期用于趋势统计
	var version models.DevVersion
	versionQuery := database.DB.Model(&models.DevVersion{}).
		Where("dev_version.version_id = ? AND dev_version.del_flag = ?", versionID, 0)
	if err := permission.Apply(versionQuery, "dev_version.creator_id").First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("版本不存在")
		}
		return nil, err
	}

	dictMap := loadVersionStatDictLabels()

	// ---- 基础查询：已删除=false 且属于当前版本 ----
	storyBase := applyStoryPermission(
		database.DB.Model(&models.DevStory{}).Where("dev_story.del_flag = ? AND dev_story.version_id = ?", 0, versionID),
		permission,
	)
	taskBase := permission.Apply(
		database.DB.Model(&models.DevTask{}).Where("dev_task.del_flag = ? AND dev_task.version_id = ?", 0, versionID),
		"dev_task.creator_id", "dev_task.user_id",
	)
	bugBase := permission.Apply(
		database.DB.Model(&models.DevBug{}).Where("dev_bug.del_flag = ? AND dev_bug.version_id = ?", 0, versionID),
		"dev_bug.creator_id", "dev_bug.fix_user_id",
	)

	// ---- 概览汇总 ----
	resp := &models.VersionStatisticsResponse{
		PersonTaskDist:    []models.VersionDistItem{},
		PersonStoryDist:   []models.VersionDistItem{},
		PersonHoursDist:   []models.VersionDistItem{},
		ModuleDist:        []models.VersionDistItem{},
		StoryTypeDist:     []models.VersionDistItem{},
		StorySourceDist:   []models.VersionDistItem{},
		StoryStatusFunnel: []models.VersionDistItem{},
		TaskTypeDist:      []models.VersionDistItem{},
		BugTypeDist:       []models.VersionDistItem{},
		BugLevelDist:      []models.VersionDistItem{},
		BugSourceDist:     []models.VersionDistItem{},
		BugFixerDist:      []models.VersionDistItem{},
	}

	storyTotal, storyDone, err := countTwoStates(storyBase, "dev_story.story_status", versionStoryDoneStatus)
	if err != nil {
		return nil, err
	}
	taskTotal, taskDone, err := countTwoStates(taskBase, "dev_task.task_status", versionTaskDoneStatus)
	if err != nil {
		return nil, err
	}
	bugTotal, bugFixed, err := countTwoStates(bugBase, "dev_bug.bug_status", versionBugFixedStatus)
	if err != nil {
		return nil, err
	}
	resp.Summary = models.VersionStatisticsSummary{
		StoryTotal: storyTotal, StoryDone: storyDone,
		TaskTotal: taskTotal, TaskDone: taskDone,
		BugTotal: bugTotal, BugFixed: bugFixed,
	}

	// ---- 完成趋势：按版本周期逐日统计完成任务/修复缺陷 ----
	resp.ProgressTrend, err = buildVersionProgressTrend(version, taskBase, bugBase)
	if err != nil {
		return nil, err
	}

	// ---- 各维度原始分组（数字字典维度）----
	storyTypeRows, _ := groupDist(storyBase.Session(&gorm.Session{}), "dev_story.story_type")
	storySourceRows, _ := groupDist(storyBase.Session(&gorm.Session{}), "dev_story.source")
	storyStatusRows, _ := groupDist(storyBase.Session(&gorm.Session{}), "dev_story.story_status")
	taskTypeRows, _ := groupDist(taskBase.Session(&gorm.Session{}), "dev_task.task_type")
	bugTypeRows, _ := groupDist(bugBase.Session(&gorm.Session{}), "dev_bug.bug_type")
	bugLevelRows, _ := groupDist(bugBase.Session(&gorm.Session{}), "dev_bug.bug_level")
	bugSourceRows, _ := groupDist(bugBase.Session(&gorm.Session{}), "dev_bug.bug_source")

	// ---- 人维度与模块维度原始分组 ----
	personTaskRows, _ := groupDist(taskBase.Session(&gorm.Session{}), "dev_task.user_id")
	personHoursRows, _ := groupDist(taskBase.Session(&gorm.Session{}), "dev_task.user_id")
	moduleRows, _ := groupDist(taskBase.Session(&gorm.Session{}), "dev_task.module_id")
	bugFixerRows, _ := groupDist(bugBase.Session(&gorm.Session{}), "dev_bug.fix_user_id")

	// 需求参与人：join 参与人表，按人统计其参与的版本需求数
	personStoryRows := []distRow{}
	storyUserQuery := database.DB.Table("dev_story AS s").
		Joins("JOIN dev_story_user dsu ON dsu.story_id = s.story_id").
		Where("s.del_flag = ? AND s.version_id = ?", 0, versionID)
	storyUserQuery = applyStoryPermission(storyUserQuery, permission)
	if err := storyUserQuery.Select("dsu.user_id AS k, COUNT(DISTINCT s.story_id) AS num").
		Group("dsu.user_id").Scan(&personStoryRows).Error; err != nil {
		return nil, err
	}

	// 任务工时：按执行人汇总计划/实际工时
	workloadRows := []workloadRow{}
	if err := taskBase.Session(&gorm.Session{}).
		Select("dev_task.user_id AS user_id, COALESCE(SUM(dev_task.plan_hours),0) AS plan_hours, COALESCE(SUM(dev_task.actual_hours),0) AS actual_hours").
		Where("dev_task.user_id IS NOT NULL AND dev_task.user_id <> ''").
		Group("dev_task.user_id").Scan(&workloadRows).Error; err != nil {
		return nil, err
	}

	// ---- 批量翻译人员姓名与模块标题 ----
	userIDSet := map[string]bool{}
	collectUserIDs(userIDSet, personTaskRows, personStoryRows, personHoursRows, bugFixerRows)
	for _, w := range workloadRows {
		if w.UserID != "" {
			userIDSet[w.UserID] = true
		}
	}
	moduleIDSet := map[string]bool{}
	for _, r := range moduleRows {
		if r.Key != "" {
			moduleIDSet[r.Key] = true
		}
	}
	userNames := loadUserRealNames(userIDSet)
	moduleTitles := loadModuleTitles(moduleIDSet)

	// ---- 字典维度翻译拼装 ----
	resp.StoryTypeDist = toDictDist(storyTypeRows, dictMap, "STORY_TYPE")
	resp.StorySourceDist = toDictDist(storySourceRows, dictMap, "STORY_SOURCE")
	resp.StoryStatusFunnel = toDictDist(storyStatusRows, dictMap, "STORY_STATUS")
	resp.TaskTypeDist = toDictDist(taskTypeRows, dictMap, "TASK_TYPE")
	resp.BugTypeDist = toDictDist(bugTypeRows, dictMap, "BUG_TYPE")
	resp.BugLevelDist = toDictDist(bugLevelRows, dictMap, "BUG_LEVEL")
	resp.BugSourceDist = toDictDist(bugSourceRows, dictMap, "BUG_SOURCE")

	// ---- 人/模块维度翻译拼装 ----
	resp.PersonTaskDist = toUserDist(personTaskRows, userNames)
	resp.PersonStoryDist = toUserDist(personStoryRows, userNames)
	resp.PersonHoursDist = toUserDist(personHoursRows, userNames)
	resp.BugFixerDist = toUserDist(bugFixerRows, userNames)
	resp.ModuleDist = toModuleDist(moduleRows, moduleTitles)

	// ---- 工时对比：categories 与 workloadRows 对齐 ----
	categories := make([]string, 0, len(workloadRows))
	planHours := make([]float64, 0, len(workloadRows))
	actualHours := make([]float64, 0, len(workloadRows))
	for _, w := range workloadRows {
		name := userNames[w.UserID]
		if name == "" {
			name = "未分配"
		}
		categories = append(categories, name)
		planHours = append(planHours, w.PlanHours)
		actualHours = append(actualHours, w.ActualHours)
	}
	resp.TaskWorkload = models.VersionTaskWorkload{
		Categories: categories, PlanHours: planHours, ActualHours: actualHours,
	}

	return resp, nil
}

// countTwoStates 统计总数与指定完成态数量。
func countTwoStates(base *gorm.DB, statusCol string, doneStatus int) (total, done int64, err error) {
	if err = base.Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = base.Session(&gorm.Session{}).Where(statusCol+" = ?", doneStatus).Count(&done).Error; err != nil {
		return 0, 0, err
	}
	return total, done, nil
}

// groupDist 按指定列分组计数，空值统一归为空字符串 key。
func groupDist(base *gorm.DB, groupCol string) ([]distRow, error) {
	rows := []distRow{}
	err := base.Select("IFNULL(" + groupCol + ", '') AS k, COUNT(*) AS num").
		Group(groupCol).Scan(&rows).Error
	return rows, err
}

// buildVersionProgressTrend 生成版本周期内逐日完成趋势；无起止日期时用创建日到今天，最多 366 天。
func buildVersionProgressTrend(version models.DevVersion, taskBase, bugBase *gorm.DB) (models.VersionProgressTrend, error) {
	start := version.StartDate
	end := version.EndDate
	now := time.Now()
	if start == nil {
		start = version.CreateDate
	}
	if start == nil {
		s := now.AddDate(0, 0, -30)
		start = &s
	}
	endDate := now
	if end != nil && end.After(*start) {
		endDate = *end
	}
	// 截断到天
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.Local)
	if endDay.Before(startDay) {
		endDay = startDay
	}
	if days := int(endDay.Sub(startDay).Hours()/24) + 1; days > 366 {
		endDay = startDay.AddDate(0, 0, 365)
	}

	dates := []string{}
	taskDoneByDay := map[string]int64{}
	bugFixedByDay := map[string]int64{}

	taskRows := []trendRow{}
	if err := taskBase.Session(&gorm.Session{}).
		Select("DATE(update_date) AS date, COUNT(*) AS count").
		Where("task_status = ? AND update_date IS NOT NULL AND update_date >= ? AND update_date <= ?",
			versionTaskDoneStatus, startDay, endDay.AddDate(0, 0, 1)).
		Group("DATE(update_date)").Scan(&taskRows).Error; err != nil {
		return models.VersionProgressTrend{}, err
	}
	for _, r := range taskRows {
		taskDoneByDay[r.Date] = r.Count
	}

	bugRows := []trendRow{}
	if err := bugBase.Session(&gorm.Session{}).
		Select("DATE(update_date) AS date, COUNT(*) AS count").
		Where("bug_status = ? AND update_date IS NOT NULL AND update_date >= ? AND update_date <= ?",
			versionBugFixedStatus, startDay, endDay.AddDate(0, 0, 1)).
		Group("DATE(update_date)").Scan(&bugRows).Error; err != nil {
		return models.VersionProgressTrend{}, err
	}
	for _, r := range bugRows {
		bugFixedByDay[r.Date] = r.Count
	}

	taskDone := []int64{}
	bugFixed := []int64{}
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		dates = append(dates, key)
		taskDone = append(taskDone, taskDoneByDay[key])
		bugFixed = append(bugFixed, bugFixedByDay[key])
	}
	return models.VersionProgressTrend{Dates: dates, TaskDone: taskDone, BugFixed: bugFixed}, nil
}

// loadVersionStatDictLabels 一次性加载版本统计用到的字典 value->label 映射。
func loadVersionStatDictLabels() map[string]map[string]string {
	types := []string{"STORY_TYPE", "STORY_SOURCE", "STORY_STATUS", "TASK_TYPE", "BUG_TYPE", "BUG_LEVEL", "BUG_SOURCE"}
	var dicts []models.SysDict
	database.DB.Where("type IN ? AND del_flag = 0", types).Find(&dicts)
	m := make(map[string]map[string]string)
	for _, d := range dicts {
		if d.Value == nil || d.Label == nil || *d.Value == "" {
			continue
		}
		if m[d.Type] == nil {
			m[d.Type] = map[string]string{}
		}
		m[d.Type][*d.Value] = *d.Label
	}
	return m
}

func collectUserIDs(dst map[string]bool, rows ...[]distRow) {
	for _, rs := range rows {
		for _, r := range rs {
			if r.Key != "" {
				dst[r.Key] = true
			}
		}
	}
}

// loadUserRealNames 批量查询用户姓名。
func loadUserRealNames(idSet map[string]bool) map[string]string {
	names := make(map[string]string)
	if len(idSet) == 0 {
		return names
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var users []models.SysUser
	database.DB.Select("user_id, real_name").Where("user_id IN ?", ids).Find(&users)
	for _, u := range users {
		if u.RealName != nil {
			names[u.UserID] = *u.RealName
		}
	}
	return names
}

// loadModuleTitles 批量查询模块标题。
func loadModuleTitles(idSet map[string]bool) map[string]string {
	titles := make(map[string]string)
	if len(idSet) == 0 {
		return titles
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var modules []models.DevModule
	database.DB.Select("module_id, module_title").Where("module_id IN ?", ids).Find(&modules)
	for _, m := range modules {
		if m.ModuleTitle != nil {
			titles[m.ModuleID] = *m.ModuleTitle
		}
	}
	return titles
}

// toDictDist 把数字分组结果翻译成中文标签分布。
func toDictDist(rows []distRow, dictMap map[string]map[string]string, dictType string) []models.VersionDistItem {
	out := make([]models.VersionDistItem, 0, len(rows))
	for _, r := range rows {
		label := r.Key
		if mm, ok := dictMap[dictType]; ok {
			if l, ok := mm[r.Key]; ok {
				label = l
			}
		}
		if label == "" {
			label = "未分配"
		}
		out = append(out, models.VersionDistItem{Name: label, Value: r.Num})
	}
	return out
}

// toUserDist 把按用户分组结果翻译成用户姓名分布。
func toUserDist(rows []distRow, userNames map[string]string) []models.VersionDistItem {
	out := make([]models.VersionDistItem, 0, len(rows))
	for _, r := range rows {
		name := userNames[r.Key]
		if name == "" {
			name = "未分配"
		}
		out = append(out, models.VersionDistItem{Name: name, Value: r.Num})
	}
	return out
}

// toModuleDist 把按模块分组结果翻译成模块标题分布。
func toModuleDist(rows []distRow, moduleTitles map[string]string) []models.VersionDistItem {
	out := make([]models.VersionDistItem, 0, len(rows))
	for _, r := range rows {
		name := moduleTitles[r.Key]
		if name == "" {
			name = "未分配"
		}
		out = append(out, models.VersionDistItem{Name: name, Value: r.Num})
	}
	return out
}

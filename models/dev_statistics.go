package models

// TaskFindDayResponse 任务趋势统计响应，按小时维度对比两个日期的任务创建数量
type TaskFindDayResponse struct {
	Date1 []int64 `json:"date1"` // 日期1（昨天）各小时任务创建数量，长度 24
	Date2 []int64 `json:"date2"` // 日期2（今天）各小时任务创建数量，长度 24
}

// TaskFindYearResponse 任务年度工时统计响应，按月份返回实际工时合计
type TaskFindYearResponse struct {
	List []float64 `json:"list"` // 各月份实际工时合计，长度 12
}

// WorkspaceEnumResponse 工作空间概览统计响应，包含需求、任务、缺陷的总数与待处理数量
type WorkspaceEnumResponse struct {
	StoryTotalNum int64 `json:"storyTotalNum"` // 需求总数
	StoryNum      int64 `json:"storyNum"`      // 待评审需求数
	TaskTotalNum  int64 `json:"taskTotalNum"`  // 任务总数
	TaskNum       int64 `json:"taskNum"`       // 待执行任务数
	BugTotalNum   int64 `json:"bugTotalNum"`   // 缺陷总数
	BugNum        int64 `json:"bugNum"`        // 待处理缺陷数
}

// VersionStatisticsSummary 版本概览汇总，统计当前版本下各工作项总数与完成数
type VersionStatisticsSummary struct {
	StoryTotal int64 `json:"storyTotal"` // 需求总数
	StoryDone  int64 `json:"storyDone"`  // 已关闭需求数(story_status=99)
	TaskTotal  int64 `json:"taskTotal"`  // 任务总数
	TaskDone   int64 `json:"taskDone"`   // 已完成任务数(task_status=99)
	BugTotal   int64 `json:"bugTotal"`   // 缺陷总数
	BugFixed   int64 `json:"bugFixed"`   // 已修复缺陷数(bug_status=20)
}

// VersionDistItem 通用分布项，name 为可直接展示的文案（字典中文名、用户姓名或模块名）
type VersionDistItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// VersionProgressTrend 版本周期内按天完成趋势
type VersionProgressTrend struct {
	Dates    []string `json:"dates"`    // 日期，格式 YYYY-MM-DD
	TaskDone []int64  `json:"taskDone"` // 当日完成任务数
	BugFixed []int64  `json:"bugFixed"` // 当日修复缺陷数
}

// VersionTaskWorkload 任务工时对比，按执行人聚合计划/实际工时
type VersionTaskWorkload struct {
	Categories  []string  `json:"categories"`  // 执行人姓名
	PlanHours   []float64 `json:"planHours"`   // 计划工时合计
	ActualHours []float64 `json:"actualHours"` // 实际工时合计
}

// VersionStatisticsResponse 版本详情统计数据，对应前端 VersionStatisticsFace
type VersionStatisticsResponse struct {
	Summary           VersionStatisticsSummary `json:"summary"`
	ProgressTrend     VersionProgressTrend     `json:"progressTrend"`
	PersonTaskDist    []VersionDistItem        `json:"personTaskDist"`    // 人员任务占比(按执行人)
	PersonStoryDist   []VersionDistItem        `json:"personStoryDist"`   // 人员需求参与(按需求参与人)
	PersonHoursDist   []VersionDistItem        `json:"personHoursDist"`   // 人员工时占比(按执行人实际工时)
	ModuleDist        []VersionDistItem        `json:"moduleDist"`        // 模块任务占比(按任务所属模块)
	StoryTypeDist     []VersionDistItem        `json:"storyTypeDist"`     // 需求类型分布
	StorySourceDist   []VersionDistItem        `json:"storySourceDist"`   // 需求来源分布
	StoryStatusFunnel []VersionDistItem        `json:"storyStatusFunnel"` // 需求状态漏斗
	TaskTypeDist      []VersionDistItem        `json:"taskTypeDist"`      // 任务类型分布
	TaskWorkload      VersionTaskWorkload      `json:"taskWorkload"`      // 任务计划/实际工时对比
	BugTypeDist       []VersionDistItem        `json:"bugTypeDist"`       // Bug类型分布
	BugLevelDist      []VersionDistItem        `json:"bugLevelDist"`      // Bug级别分布
	BugSourceDist     []VersionDistItem        `json:"bugSourceDist"`     // Bug来源分布
	BugFixerDist      []VersionDistItem        `json:"bugFixerDist"`      // Bug修复人分布
}

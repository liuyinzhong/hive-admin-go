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

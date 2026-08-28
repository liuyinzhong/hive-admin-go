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
	bugQuery := permission.Apply(database.DB.Model(&models.DevBug{}), "dev_bug.creator_id", "dev_bug.user_id")

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

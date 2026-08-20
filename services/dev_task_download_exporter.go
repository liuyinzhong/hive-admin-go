package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/xuri/excelize/v2"
)

type devTaskDownloadExporter struct{}

type DevTaskExportPayload struct {
	Request models.DevTaskExportRequest `json:"request"`
}

func NewDevTaskExportPayload(request models.DevTaskExportRequest) DevTaskExportPayload {
	return DevTaskExportPayload{Request: request}
}

type devTaskDownloadRow struct {
	TaskNum      int
	ProjectTitle string
	Version      string
	ModuleTitle  string
	StoryTitle   string
	TaskTitle    string
	RealName     string
	TaskStatus   int
	TaskType     int
	PlanHours    float64
	ActualHours  float64
	StartDate    *time.Time
	EndDate      *time.Time
	CreatorName  string
	CreateDate   *time.Time
}

type devTaskExportColumn struct {
	Field string
	Title string
	Width int
}

var devTaskExportColumns = []models.DownloadExportColumn{
	{Field: "taskNum", Title: "任务编号"},
	{Field: "projectTitle", Title: "项目"},
	{Field: "version", Title: "版本"},
	{Field: "moduleTitle", Title: "模块"},
	{Field: "storyTitle", Title: "需求"},
	{Field: "taskTitle", Title: "任务标题"},
	{Field: "userList", Title: "负责人"},
	{Field: "taskStatus", Title: "状态"},
	{Field: "taskType", Title: "类型"},
	{Field: "planHours", Title: "计划工时"},
	{Field: "actualHours", Title: "实际工时"},
	{Field: "percent", Title: "进度"},
	{Field: "startDate", Title: "开始日期"},
	{Field: "endDate", Title: "结束日期"},
	{Field: "creatorName", Title: "创建人"},
	{Field: "createDate", Title: "创建时间"},
}

var devTaskExportColumnTitles = func() map[string]string {
	result := make(map[string]string, len(devTaskExportColumns))
	for _, column := range devTaskExportColumns {
		result[column.Field] = column.Title
	}
	result["realName"] = "负责人"
	return result
}()

func (e *devTaskDownloadExporter) Count(payload, creatorID string) (int64, error) {
	request, permission, err := parseDevTaskExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := buildDevTaskQuery(devTaskExportParams(request), permission).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (e *devTaskDownloadExporter) Export(payload, creatorID, filePath string, totalRows int64, onProgress func(int64)) (int64, error) {
	request, permission, err := parseDevTaskExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	columns, err := resolveDevTaskExportColumns(request)
	if err != nil {
		return 0, err
	}
	query := buildDevTaskQuery(devTaskExportParams(request), permission)
	order := buildDevTaskOrder(devTaskExportParams(request))
	dictLabels, err := loadDevTaskDownloadDictLabels()
	if err != nil {
		return 0, err
	}
	selectFields := `
		dev_task.task_num,
		COALESCE(dev_project.project_title, '') AS project_title,
		COALESCE(dev_version.version, '') AS version,
		COALESCE(dev_module.module_title, '') AS module_title,
		COALESCE(dev_story.story_title, '') AS story_title,
		COALESCE(dev_task.task_title, '') AS task_title,
		COALESCE(assignee.real_name, '') AS real_name,
		dev_task.task_status,
		dev_task.task_type,
		dev_task.plan_hours,
		dev_task.actual_hours,
		dev_task.start_date,
		dev_task.end_date,
		COALESCE(creator.real_name, '') AS creator_name,
		dev_task.create_date`
	query = query.
		Joins("LEFT JOIN dev_project ON dev_project.project_id = dev_task.project_id AND dev_project.del_flag = 0").
		Joins("LEFT JOIN dev_version ON dev_version.version_id = dev_task.version_id AND dev_version.del_flag = 0").
		Joins("LEFT JOIN dev_module ON dev_module.module_id = dev_task.module_id AND dev_module.del_flag = 0").
		Joins("LEFT JOIN dev_story ON dev_story.story_id = dev_task.story_id AND dev_story.del_flag = 0").
		Joins("LEFT JOIN sys_user assignee ON assignee.user_id = dev_task.user_id AND assignee.del_flag = 0").
		Joins("LEFT JOIN sys_user creator ON creator.user_id = dev_task.creator_id AND creator.del_flag = 0")

	includeHeader := exportBoolValue(request.IsHeader, true)
	useTitle := exportBoolValue(request.IsTitle, true)
	original := exportBoolValue(request.Original, false)
	headers := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		if useTitle {
			headers = append(headers, column.Title)
		} else {
			headers = append(headers, column.Field)
		}
	}
	sheetName := normalizeDownloadSheetName(request.SheetName)
	var processed int64
	widths := make([]int, 0, len(columns))
	for _, column := range columns {
		widths = append(widths, column.Width)
	}
	err = writeDownloadWorkbookWithWidths(filePath, sheetName, headers, includeHeader, widths, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(selectFields).Order(order).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		rowOffset := 0
		if includeHeader {
			rowOffset = 1
		}
		for rows.Next() {
			var row devTaskDownloadRow
			if err := query.ScanRows(rows, &row); err != nil {
				return err
			}
			percent := 0
			if row.PlanHours > 0 {
				percent = int(row.ActualHours / row.PlanHours * 100)
			}
			processed++
			if processed > totalRows {
				return ErrDownloadDataChanged
			}
			cell, err := excelize.CoordinatesToCellName(1, int(processed)+rowOffset)
			if err != nil {
				return err
			}
			values := make([]interface{}, 0, len(columns))
			for _, column := range columns {
				values = append(values, devTaskExportColumnValue(column.Field, row, dictLabels, percent, original))
			}
			if err := writer.SetRow(cell, values); err != nil {
				return err
			}
			onProgress(processed)
		}
		return rows.Err()
	})
	return processed, err
}

func loadDevTaskDownloadDictLabels() (map[string]string, error) {
	var dicts []models.SysDict
	if err := database.DB.Where("type IN ? AND status = 1 AND del_flag = 0", []string{"TASK_STATUS", "TASK_TYPE"}).Find(&dicts).Error; err != nil {
		return nil, err
	}
	labels := make(map[string]string, len(dicts))
	for _, dict := range dicts {
		if dict.Label == nil || dict.Value == nil {
			continue
		}
		labels[dict.Type+":"+*dict.Value] = *dict.Label
	}
	return labels, nil
}

func formatDevTaskDownloadDictValue(labels map[string]string, dictType string, value int) string {
	rawValue := fmt.Sprintf("%d", value)
	if label := labels[dictType+":"+rawValue]; label != "" {
		return label
	}
	return rawValue
}

func devTaskExportColumnValue(field string, row devTaskDownloadRow, dictLabels map[string]string, percent int, original bool) interface{} {
	switch field {
	case "taskNum":
		if original {
			return row.TaskNum
		}
		return fmt.Sprintf("#%d", row.TaskNum)
	case "projectTitle":
		return row.ProjectTitle
	case "version":
		return row.Version
	case "moduleTitle":
		return row.ModuleTitle
	case "storyTitle":
		return row.StoryTitle
	case "taskTitle":
		return row.TaskTitle
	case "userList", "realName":
		return row.RealName
	case "taskStatus":
		if original {
			return row.TaskStatus
		}
		return formatDevTaskDownloadDictValue(dictLabels, "TASK_STATUS", row.TaskStatus)
	case "taskType":
		if original {
			return row.TaskType
		}
		return formatDevTaskDownloadDictValue(dictLabels, "TASK_TYPE", row.TaskType)
	case "planHours":
		return row.PlanHours
	case "actualHours":
		return row.ActualHours
	case "percent":
		if original {
			return percent
		}
		return fmt.Sprintf("%d%%", percent)
	case "startDate":
		return formatDownloadDate(row.StartDate)
	case "endDate":
		return formatDownloadDate(row.EndDate)
	case "creatorName":
		return row.CreatorName
	case "createDate":
		return utils.TimeToString(row.CreateDate)
	default:
		return ""
	}
}

func resolveDevTaskExportColumns(request models.DevTaskExportRequest) ([]devTaskExportColumn, error) {
	selected := request.Columns
	if len(selected) == 0 {
		return nil, fmt.Errorf("导出列不能为空")
	}
	columns := make([]devTaskExportColumn, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, selectedColumn := range selected {
		field := strings.TrimSpace(selectedColumn.Field)
		title, ok := devTaskExportColumnTitles[field]
		if !ok {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		if value := strings.TrimSpace(selectedColumn.Title); value != "" {
			titleRunes := []rune(value)
			if len(titleRunes) > 255 {
				value = string(titleRunes[:255])
			}
			title = value
		}
		columns = append(columns, devTaskExportColumn{Field: field, Title: title, Width: selectedColumn.Width})
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("没有可导出的有效列")
	}
	return columns, nil
}

func parseDevTaskExportRequest(payload, creatorID string) (models.DevTaskExportRequest, datapermission.Permission, error) {
	request, err := decodeDevTaskExportRequest(payload)
	if err != nil {
		return models.DevTaskExportRequest{}, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return models.DevTaskExportRequest{}, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return models.DevTaskExportRequest{}, datapermission.Permission{}, err
	}
	return request, permission, nil
}

func devTaskExportParams(request models.DevTaskExportRequest) map[string]interface{} {
	return map[string]interface{}{
		"taskTitle":    request.TaskTitle,
		"projectId":    request.ProjectID,
		"versionId":    request.VersionID,
		"taskStatuses": request.TaskStatuses,
		"sorts":        request.Sorts,
	}
}

func decodeDevTaskExportRequest(payload string) (models.DevTaskExportRequest, error) {
	var envelope struct {
		Request json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return models.DevTaskExportRequest{}, err
	}
	requestPayload := []byte(payload)
	if len(envelope.Request) > 0 && string(envelope.Request) != "null" {
		requestPayload = envelope.Request
	}
	var request models.DevTaskExportRequest
	if err := json.Unmarshal(requestPayload, &request); err != nil {
		return models.DevTaskExportRequest{}, err
	}
	return request, nil
}

func formatDownloadDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func exportBoolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func normalizeDownloadSheetName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "任务管理"
	}
	value = strings.NewReplacer("[", "_", "]", "_", ":", "_", "*", "_", "?", "_", "/", "_", "\\", "_").Replace(value)
	runes := []rune(value)
	if len(runes) > 31 {
		value = string(runes[:31])
	}
	if strings.TrimSpace(value) == "" {
		return "任务管理"
	}
	return value
}

func normalizeDownloadFileName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.NewReplacer("<", "_", ">", "_", ":", "_", "\"", "_", "/", "_", "\\", "_", "|", "_", "?", "_", "*", "_").Replace(value)
	filtered := make([]rune, 0, len(value))
	for _, char := range value {
		if char < 32 || char == 127 {
			continue
		}
		filtered = append(filtered, char)
	}
	value = strings.Trim(string(filtered), " .")
	if value == "" {
		return ""
	}
	if runes := []rune(value); len(runes) > 180 {
		value = strings.Trim(string(runes[:180]), " .")
	}
	if !strings.HasSuffix(strings.ToLower(value), ".xlsx") {
		value += ".xlsx"
	}
	return value
}

func (e *devTaskDownloadExporter) ResolveFileName(payload string, _ time.Time) (string, error) {
	request, err := decodeDevTaskExportRequest(payload)
	if err != nil {
		return "", err
	}
	return normalizeDownloadFileName(request.Filename), nil
}

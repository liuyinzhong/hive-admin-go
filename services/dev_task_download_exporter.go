package services

import (
	"encoding/json"
	"fmt"
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

func (e *devTaskDownloadExporter) Count(payload, creatorID string) (int64, error) {
	params, permission, err := parseDevTaskExportParams(payload, creatorID)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := buildDevTaskQuery(params, permission).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (e *devTaskDownloadExporter) Export(payload, creatorID, filePath string, totalRows int64, onProgress func(int64)) (int64, error) {
	params, permission, err := parseDevTaskExportParams(payload, creatorID)
	if err != nil {
		return 0, err
	}
	query := buildDevTaskQuery(params, permission)
	order := buildDevTaskOrder(params)
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

	headers := []interface{}{
		"任务编号", "项目", "版本", "模块", "需求", "任务标题", "负责人", "状态", "类型", "计划工时", "实际工时",
		"进度", "开始日期", "结束日期", "创建人", "创建时间",
	}
	var processed int64
	err = writeDownloadWorkbook(filePath, "任务管理", headers, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(selectFields).Order(order).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

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
			cell, err := excelize.CoordinatesToCellName(1, int(processed)+1)
			if err != nil {
				return err
			}
			values := []interface{}{
				fmt.Sprintf("#%d", row.TaskNum),
				row.ProjectTitle,
				row.Version,
				row.ModuleTitle,
				row.StoryTitle,
				row.TaskTitle,
				row.RealName,
				formatDevTaskDownloadDictValue(dictLabels, "TASK_STATUS", row.TaskStatus),
				formatDevTaskDownloadDictValue(dictLabels, "TASK_TYPE", row.TaskType),
				row.PlanHours,
				row.ActualHours,
				fmt.Sprintf("%d%%", percent),
				formatDownloadDate(row.StartDate),
				formatDownloadDate(row.EndDate),
				row.CreatorName,
				utils.TimeToString(row.CreateDate),
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

func parseDevTaskExportParams(payload, creatorID string) (map[string]interface{}, datapermission.Permission, error) {
	request, err := decodeDevTaskExportRequest(payload)
	if err != nil {
		return nil, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return nil, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return nil, datapermission.Permission{}, err
	}
	return map[string]interface{}{
		"taskTitle":    request.TaskTitle,
		"projectId":    request.ProjectID,
		"versionId":    request.VersionID,
		"taskStatuses": request.TaskStatuses,
		"sorts":        request.Sorts,
	}, permission, nil
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

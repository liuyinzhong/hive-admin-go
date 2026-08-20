package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"

	"github.com/xuri/excelize/v2"
)

type loginLogDownloadExporter struct{}

type loginLogExportRequest struct {
	models.AuditLogListRequest
	Filename  string
	SheetName string
	Columns   []models.DownloadExportColumn
	IsHeader  *bool
	IsTitle   *bool
	Original  *bool
}

type loginLogExportColumn struct {
	Field string
	Title string
	Width int
}

type loginLogDownloadRow struct {
	Username   string
	EventType  string
	HTTPStatus int
	Status     int
	DurationMs int64
	IP         string
	UserAgent  string
	CreateDate time.Time
}

var loginLogExportColumnTitles = map[string]string{
	"username":   "用户名",
	"eventType":  "事件类型",
	"httpStatus": "HTTP状态",
	"status":     "执行状态",
	"durationMs": "耗时（毫秒）",
	"ip":         "客户端IP",
	"userAgent":  "浏览器 / User-Agent",
	"createDate": "发生时间",
}

type LoginLogExportPayload struct {
	Request models.LoginLogExportRequest `json:"request"`
}

func NewLoginLogExportPayload(request models.LoginLogExportRequest) LoginLogExportPayload {
	return LoginLogExportPayload{Request: request}
}

func (e *loginLogDownloadExporter) Count(payload, creatorID string) (int64, error) {
	req, permission, err := parseLoginLogExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	if _, err := resolveLoginLogExportColumns(req); err != nil {
		return 0, err
	}
	query, err := buildLoginLogQuery(req.AuditLogListRequest, permission)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (e *loginLogDownloadExporter) Export(payload, creatorID, filePath string, totalRows int64, onProgress func(int64)) (int64, error) {
	req, permission, err := parseLoginLogExportRequest(payload, creatorID)
	if err != nil {
		return 0, err
	}
	columns, err := resolveLoginLogExportColumns(req)
	if err != nil {
		return 0, err
	}
	query, err := buildLoginLogQuery(req.AuditLogListRequest, permission)
	if err != nil {
		return 0, err
	}

	includeHeader := exportBoolValue(req.IsHeader, true)
	useTitle := exportBoolValue(req.IsTitle, true)
	original := exportBoolValue(req.Original, false)
	headers := make([]interface{}, 0, len(columns))
	widths := make([]int, 0, len(columns))
	for _, column := range columns {
		if useTitle {
			headers = append(headers, column.Title)
		} else {
			headers = append(headers, column.Field)
		}
		widths = append(widths, column.Width)
	}
	sheetName := normalizeDownloadSheetName(req.SheetName)
	var processed int64
	err = writeDownloadWorkbookWithWidths(filePath, sheetName, headers, includeHeader, widths, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(
			"username", "event_type", "http_status", "status", "duration_ms", "ip", "user_agent", "create_date",
		).Order(buildLoginLogOrder(req.Sorts)).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		rowOffset := 0
		if includeHeader {
			rowOffset = 1
		}
		for rows.Next() {
			var row loginLogDownloadRow
			if err := query.ScanRows(rows, &row); err != nil {
				return err
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
				values = append(values, loginLogExportColumnValue(column.Field, row, original))
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

func parseLoginLogExportRequest(payload, creatorID string) (loginLogExportRequest, datapermission.Permission, error) {
	request, err := decodeLoginLogExportRequest(payload)
	if err != nil {
		return loginLogExportRequest{}, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return loginLogExportRequest{}, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return loginLogExportRequest{}, datapermission.Permission{}, err
	}
	return loginLogExportRequest{
		AuditLogListRequest: models.AuditLogListRequest{
			Username:  request.Username,
			EventType: request.EventType,
			Status:    request.Status,
			IP:        request.IP,
			StartDate: request.StartDate,
			EndDate:   request.EndDate,
			Sorts:     request.Sorts,
		},
		Filename:  request.Filename,
		SheetName: request.SheetName,
		Columns:   request.Columns,
		IsHeader:  request.IsHeader,
		IsTitle:   request.IsTitle,
		Original:  request.Original,
	}, permission, nil
}

func decodeLoginLogExportRequest(payload string) (models.LoginLogExportRequest, error) {
	var envelope struct {
		Request json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return models.LoginLogExportRequest{}, err
	}
	requestPayload := []byte(payload)
	if len(envelope.Request) > 0 && string(envelope.Request) != "null" {
		requestPayload = envelope.Request
	}
	var request models.LoginLogExportRequest
	if err := json.Unmarshal(requestPayload, &request); err != nil {
		return models.LoginLogExportRequest{}, err
	}
	return request, nil
}

func resolveLoginLogExportColumns(request loginLogExportRequest) ([]loginLogExportColumn, error) {
	if len(request.Columns) == 0 {
		return nil, fmt.Errorf("导出列不能为空")
	}
	columns := make([]loginLogExportColumn, 0, len(request.Columns))
	seen := make(map[string]struct{}, len(request.Columns))
	for _, selectedColumn := range request.Columns {
		field := strings.TrimSpace(selectedColumn.Field)
		title, ok := loginLogExportColumnTitles[field]
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
		columns = append(columns, loginLogExportColumn{
			Field: field,
			Title: title,
			Width: selectedColumn.Width,
		})
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("没有可导出的有效列")
	}
	return columns, nil
}

func loginLogExportColumnValue(field string, row loginLogDownloadRow, original bool) interface{} {
	switch field {
	case "username":
		return row.Username
	case "eventType":
		if original {
			return row.EventType
		}
		return formatLoginLogEventType(row.EventType)
	case "httpStatus":
		return row.HTTPStatus
	case "status":
		if original {
			return row.Status
		}
		return formatLoginLogStatus(row.Status)
	case "durationMs":
		return row.DurationMs
	case "ip":
		return row.IP
	case "userAgent":
		return row.UserAgent
	case "createDate":
		if original {
			return row.CreateDate
		}
		return row.CreateDate.In(auditLogLocation).Format("2006-01-02 15:04:05")
	default:
		return ""
	}
}

func (e *loginLogDownloadExporter) ResolveFileName(payload string, _ time.Time) (string, error) {
	request, err := decodeLoginLogExportRequest(payload)
	if err != nil {
		return "", err
	}
	return normalizeDownloadFileName(request.Filename), nil
}

func formatLoginLogEventType(value string) string {
	switch value {
	case models.LoginLogTypeLogin:
		return "登录"
	case models.LoginLogTypeLogout:
		return "退出"
	default:
		return value
	}
}

func formatLoginLogStatus(value int) string {
	if value == models.AuditLogStatusSuccess {
		return "成功"
	}
	return "失败"
}

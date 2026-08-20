package services

import (
	"encoding/json"
	"fmt"
	"time"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"

	"github.com/xuri/excelize/v2"
)

type loginLogDownloadExporter struct{}

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
	query, err := buildLoginLogQuery(req, permission)
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
	query, err := buildLoginLogQuery(req, permission)
	if err != nil {
		return 0, err
	}

	headers := []interface{}{
		"用户名", "事件类型", "HTTP状态", "执行状态", "耗时（毫秒）", "客户端IP", "浏览器 / User-Agent", "发生时间",
	}
	var processed int64
	err = writeDownloadWorkbook(filePath, "登录日志", headers, func(writer *excelize.StreamWriter) error {
		rows, err := query.Select(
			"username", "event_type", "http_status", "status", "duration_ms", "ip", "user_agent", "create_date",
		).Order(buildLoginLogOrder(req.Sorts)).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var row struct {
				Username   string
				EventType  string
				HTTPStatus int
				Status     int
				DurationMs int64
				IP         string
				UserAgent  string
				CreateDate time.Time
			}
			if err := query.ScanRows(rows, &row); err != nil {
				return err
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
				row.Username,
				formatLoginLogEventType(row.EventType),
				row.HTTPStatus,
				formatLoginLogStatus(row.Status),
				row.DurationMs,
				row.IP,
				row.UserAgent,
				row.CreateDate.In(auditLogLocation).Format("2006-01-02 15:04:05"),
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

func parseLoginLogExportRequest(payload, creatorID string) (models.AuditLogListRequest, datapermission.Permission, error) {
	request, err := decodeLoginLogExportRequest(payload)
	if err != nil {
		return models.AuditLogListRequest{}, datapermission.Permission{}, err
	}
	if creatorID == "" {
		return models.AuditLogListRequest{}, datapermission.Permission{}, fmt.Errorf("导出任务缺少创建用户")
	}
	permission, err := resolveDataPermission(creatorID)
	if err != nil {
		return models.AuditLogListRequest{}, datapermission.Permission{}, err
	}
	return models.AuditLogListRequest{
		Username:  request.Username,
		EventType: request.EventType,
		Status:    request.Status,
		IP:        request.IP,
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
		Sorts:     request.Sorts,
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

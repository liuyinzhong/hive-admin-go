package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrAuditLogInvalidInput = errors.New("日志查询参数错误")
	ErrAuditLogNotFound     = errors.New("日志不存在")
)

var auditLogLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type AuditLogService struct{}

func NewAuditLogService() *AuditLogService { return &AuditLogService{} }

func (s *AuditLogService) RecordOperation(entry models.OperationLogEntry) error {
	userID := nullableString(entry.UserID)
	username, realName, err := loadAuditActor(entry.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	now := time.Now().In(auditLogLocation)
	row := models.SysOperationLog{
		LogID: utils.GenerateUUID(), UserID: userID, Username: username, RealName: realName,
		RequestMethod: entry.RequestMethod, RequestURL: entry.RequestURL, QueryParams: entry.QueryParams,
		RequestBody: entry.RequestBody, ResponseBody: entry.ResponseBody,
		QueryTruncated: boolToInt(entry.QueryTruncated), RequestTruncated: boolToInt(entry.RequestTruncated),
		ResponseTruncated: boolToInt(entry.ResponseTruncated),
		HTTPStatus:        entry.HTTPStatus, Status: auditStatus(entry.HTTPStatus), DurationMs: entry.DurationMs,
		IP: entry.IP, UserAgent: entry.UserAgent, ContentType: entry.ContentType, CreateDate: now,
	}
	return database.DB.Create(&row).Error
}

func (s *AuditLogService) RecordLogin(entry models.LoginLogEntry) error {
	userID := entry.UserID
	username := entry.Username
	if userID == "" && username != "" {
		var user models.SysUser
		if err := database.DB.Select("user_id", "username").Where("username = ? AND del_flag = 0", username).First(&user).Error; err == nil {
			userID = user.UserID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if username == "" && userID != "" {
		resolvedUsername, _, err := loadAuditActor(userID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		username = resolvedUsername
	}
	now := time.Now().In(auditLogLocation)
	row := models.SysLoginLog{
		LogID: utils.GenerateUUID(), UserID: nullableString(userID), Username: username, EventType: entry.EventType,
		ResponseBody: entry.ResponseBody, ResponseTruncated: boolToInt(entry.ResponseTruncated),
		HTTPStatus: entry.HTTPStatus, Status: auditStatus(entry.HTTPStatus),
		DurationMs: entry.DurationMs, IP: entry.IP, UserAgent: entry.UserAgent,
		ContentType: entry.ContentType, CreateDate: now,
	}
	return database.DB.Create(&row).Error
}

func (s *AuditLogService) GetOperationLogs(req models.AuditLogListRequest) (*utils.PageResult, error) {
	query, err := applyCommonAuditFilters(database.DB.Model(&models.SysOperationLog{}), req)
	if err != nil {
		return nil, err
	}
	if req.RequestURL != "" {
		query = query.Where("request_url LIKE ?", "%"+req.RequestURL+"%")
	}
	if req.Method != "" {
		query = query.Where("request_method = ?", strings.ToUpper(req.Method))
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"createDate": "create_date", "durationMs": "duration_ms", "httpStatus": "http_status",
	})
	if order == "" {
		order = "create_date desc, log_id asc"
	} else {
		order += ", log_id asc"
	}
	var rows []models.SysOperationLog
	result, err := paginateAuditQuery(query.Select("log_id", "username", "real_name", "request_method", "request_url", "http_status", "status", "duration_ms", "ip", "create_date").Order(order), normalizePage(req.Page), normalizePageSize(req.PageSize), &rows)
	if err != nil {
		return nil, err
	}
	items := make([]models.OperationLogListResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, operationLogListResponse(row))
	}
	result.Items = items
	return result, nil
}

func (s *AuditLogService) GetOperationLog(logID string) (*models.OperationLogDetailResponse, error) {
	var row models.SysOperationLog
	if err := database.DB.Where("log_id = ?", logID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}
	return &models.OperationLogDetailResponse{
		OperationLogListResponse: operationLogListResponse(row), UserID: stringValue(row.UserID),
		QueryParams: row.QueryParams, QueryTruncated: row.QueryTruncated == 1,
		RequestBody: row.RequestBody, ResponseBody: row.ResponseBody,
		RequestTruncated: row.RequestTruncated == 1, ResponseTruncated: row.ResponseTruncated == 1,
		UserAgent: row.UserAgent, ContentType: row.ContentType,
	}, nil
}

func (s *AuditLogService) GetLoginLogs(req models.AuditLogListRequest) (*utils.PageResult, error) {
	query, err := applyCommonAuditFilters(database.DB.Model(&models.SysLoginLog{}), req)
	if err != nil {
		return nil, err
	}
	if req.EventType != "" {
		if req.EventType != models.LoginLogTypeLogin && req.EventType != models.LoginLogTypeLogout {
			return nil, fmt.Errorf("%w: 登录类型不正确", ErrAuditLogInvalidInput)
		}
		query = query.Where("event_type = ?", req.EventType)
	}
	if req.IP != "" {
		query = query.Where("ip LIKE ?", "%"+req.IP+"%")
	}
	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"createDate": "create_date", "durationMs": "duration_ms", "httpStatus": "http_status",
	})
	if order == "" {
		order = "create_date desc, log_id asc"
	} else {
		order += ", log_id asc"
	}
	var rows []models.SysLoginLog
	result, err := paginateAuditQuery(query.Select("log_id", "username", "event_type", "http_status", "status", "duration_ms", "ip", "user_agent", "create_date").Order(order), normalizePage(req.Page), normalizePageSize(req.PageSize), &rows)
	if err != nil {
		return nil, err
	}
	items := make([]models.LoginLogListResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, loginLogListResponse(row))
	}
	result.Items = items
	return result, nil
}

func (s *AuditLogService) GetLoginLog(logID string) (*models.LoginLogDetailResponse, error) {
	var row models.SysLoginLog
	if err := database.DB.Where("log_id = ?", logID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}
	return &models.LoginLogDetailResponse{
		LoginLogListResponse: loginLogListResponse(row), UserID: stringValue(row.UserID),
		ResponseBody: row.ResponseBody, ResponseTruncated: row.ResponseTruncated == 1,
		ContentType: row.ContentType,
	}, nil
}

func (s *AuditLogService) CleanupExpiredLogs(retentionDays int, now time.Time) error {
	if retentionDays <= 0 {
		retentionDays = 180
	}
	cutoff := now.In(auditLogLocation).AddDate(0, 0, -retentionDays)
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("create_date < ?", cutoff).Delete(&models.SysOperationLog{}).Error; err != nil {
			return err
		}
		return tx.Where("create_date < ?", cutoff).Delete(&models.SysLoginLog{}).Error
	})
}

func StartAuditLogCleanupScheduler(retentionDays, cleanupHour int) {
	go func() {
		service := NewAuditLogService()
		for {
			now := time.Now().In(auditLogLocation)
			nextRun := nextAuditCleanupRun(now, cleanupHour)
			timer := time.NewTimer(time.Until(nextRun))
			<-timer.C
			if err := service.CleanupExpiredLogs(retentionDays, nextRun); err != nil {
				log.Printf("清理过期审计日志失败: %v", err)
			}
		}
	}()
}

func nextAuditCleanupRun(now time.Time, cleanupHour int) time.Time {
	if cleanupHour < 0 || cleanupHour > 23 {
		cleanupHour = 3
	}
	local := now.In(auditLogLocation)
	next := time.Date(local.Year(), local.Month(), local.Day(), cleanupHour, 0, 0, 0, auditLogLocation)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func applyCommonAuditFilters(query *gorm.DB, req models.AuditLogListRequest) (*gorm.DB, error) {
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Status != nil {
		if *req.Status != models.AuditLogStatusFailed && *req.Status != models.AuditLogStatusSuccess {
			return nil, fmt.Errorf("%w: 状态不正确", ErrAuditLogInvalidInput)
		}
		query = query.Where("status = ?", *req.Status)
	}
	start, end, err := auditDateRange(req.StartDate, req.EndDate, time.Now().In(auditLogLocation))
	if err != nil {
		return nil, err
	}
	return query.Where("create_date BETWEEN ? AND ?", start, end), nil
}

func auditDateRange(startValue, endValue string, now time.Time) (time.Time, time.Time, error) {
	end := now
	start := end.AddDate(0, 0, -7)
	var err error
	if startValue != "" {
		start, err = parseAuditDate(startValue, false)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if endValue != "" {
		end, err = parseAuditDate(endValue, true)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: 结束时间不能早于开始时间", ErrAuditLogInvalidInput)
	}
	return start, end, nil
}

func parseAuditDate(value string, endOfDay bool) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"} {
		parsed, err := time.ParseInLocation(layout, value, auditLogLocation)
		if err == nil {
			if layout == "2006-01-02" && endOfDay {
				parsed = parsed.Add(24*time.Hour - time.Nanosecond)
			}
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: 时间格式不正确", ErrAuditLogInvalidInput)
}

func loadAuditActor(userID string) (string, string, error) {
	if userID == "" {
		return "", "", gorm.ErrRecordNotFound
	}
	var user models.SysUser
	err := database.DB.Select("username", "real_name").Where("user_id = ?", userID).First(&user).Error
	return utils.StringValue(user.Username), utils.StringValue(user.RealName), err
}

func operationLogListResponse(row models.SysOperationLog) models.OperationLogListResponse {
	return models.OperationLogListResponse{
		LogID: row.LogID, Username: row.Username, RealName: row.RealName,
		RequestMethod: row.RequestMethod, RequestURL: row.RequestURL,
		HTTPStatus: row.HTTPStatus, Status: row.Status, DurationMs: row.DurationMs,
		IP: row.IP, CreateDate: row.CreateDate.In(auditLogLocation).Format("2006-01-02 15:04:05"),
	}
}

func loginLogListResponse(row models.SysLoginLog) models.LoginLogListResponse {
	return models.LoginLogListResponse{
		LogID: row.LogID, Username: row.Username, EventType: row.EventType,
		HTTPStatus: row.HTTPStatus, Status: row.Status, DurationMs: row.DurationMs,
		IP: row.IP, UserAgent: row.UserAgent,
		CreateDate: row.CreateDate.In(auditLogLocation).Format("2006-01-02 15:04:05"),
	}
}

func auditStatus(httpStatus int) int {
	if httpStatus >= http.StatusBadRequest {
		return models.AuditLogStatusFailed
	}
	return models.AuditLogStatusSuccess
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func paginateAuditQuery(query *gorm.DB, page, pageSize int, destination interface{}) (*utils.PageResult, error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(destination).Error; err != nil {
		return nil, err
	}
	return &utils.PageResult{Items: destination, Total: total}, nil
}

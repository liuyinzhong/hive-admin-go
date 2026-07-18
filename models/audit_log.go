package models

import "time"

const (
	AuditLogStatusFailed  = 0
	AuditLogStatusSuccess = 1

	LoginLogTypeLogin  = "login"
	LoginLogTypeLogout = "logout"
)

type SysOperationLog struct {
	LogID             string    `gorm:"column:log_id;type:char(36);primaryKey" json:"logId"`
	UserID            *string   `gorm:"column:user_id;type:char(36)" json:"userId"`
	Username          string    `gorm:"column:username;type:varchar(36)" json:"username"`
	RealName          string    `gorm:"column:real_name;type:varchar(36)" json:"realName"`
	RequestMethod     string    `gorm:"column:request_method;type:varchar(10)" json:"requestMethod"`
	RequestURL        string    `gorm:"column:request_url;type:varchar(512)" json:"requestUrl"`
	QueryParams       string    `gorm:"column:query_params;type:longtext" json:"queryParams"`
	RequestBody       string    `gorm:"column:request_body;type:longtext" json:"requestBody"`
	ResponseBody      string    `gorm:"column:response_body;type:longtext" json:"responseBody"`
	QueryTruncated    int       `gorm:"column:query_truncated;type:tinyint;default:0" json:"queryTruncated"`
	RequestTruncated  int       `gorm:"column:request_truncated;type:tinyint;default:0" json:"requestTruncated"`
	ResponseTruncated int       `gorm:"column:response_truncated;type:tinyint;default:0" json:"responseTruncated"`
	HTTPStatus        int       `gorm:"column:http_status;type:int" json:"httpStatus"`
	Status            int       `gorm:"column:status;type:tinyint" json:"status"`
	DurationMs        int64     `gorm:"column:duration_ms;type:bigint" json:"durationMs"`
	IP                string    `gorm:"column:ip;type:varchar(64)" json:"ip"`
	UserAgent         string    `gorm:"column:user_agent;type:varchar(512)" json:"userAgent"`
	ContentType       string    `gorm:"column:content_type;type:varchar(128)" json:"contentType"`
	CreateDate        time.Time `gorm:"column:create_date" json:"createDate"`
}

func (SysOperationLog) TableName() string { return "sys_operation_log" }

type SysLoginLog struct {
	LogID             string    `gorm:"column:log_id;type:char(36);primaryKey" json:"logId"`
	UserID            *string   `gorm:"column:user_id;type:char(36)" json:"userId"`
	Username          string    `gorm:"column:username;type:varchar(36)" json:"username"`
	EventType         string    `gorm:"column:event_type;type:varchar(16)" json:"eventType"`
	ResponseBody      string    `gorm:"column:response_body;type:longtext" json:"responseBody"`
	ResponseTruncated int       `gorm:"column:response_truncated;type:tinyint;default:0" json:"responseTruncated"`
	HTTPStatus        int       `gorm:"column:http_status;type:int" json:"httpStatus"`
	Status            int       `gorm:"column:status;type:tinyint" json:"status"`
	DurationMs        int64     `gorm:"column:duration_ms;type:bigint" json:"durationMs"`
	IP                string    `gorm:"column:ip;type:varchar(64)" json:"ip"`
	UserAgent         string    `gorm:"column:user_agent;type:varchar(512)" json:"userAgent"`
	ContentType       string    `gorm:"column:content_type;type:varchar(128)" json:"contentType"`
	CreateDate        time.Time `gorm:"column:create_date" json:"createDate"`
}

func (SysLoginLog) TableName() string { return "sys_login_log" }

type OperationLogEntry struct {
	UserID            string
	RequestMethod     string
	RequestURL        string
	QueryParams       string
	QueryTruncated    bool
	RequestBody       string
	ResponseBody      string
	RequestTruncated  bool
	ResponseTruncated bool
	HTTPStatus        int
	DurationMs        int64
	IP                string
	UserAgent         string
	ContentType       string
}

type LoginLogEntry struct {
	UserID            string
	Username          string
	EventType         string
	ResponseBody      string
	ResponseTruncated bool
	HTTPStatus        int
	DurationMs        int64
	IP                string
	UserAgent         string
	ContentType       string
}

type AuditLogListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	Username   string `form:"username"`
	IP         string `form:"ip"`
	Status     *int   `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
	Sorts      string `form:"sorts"`
	RequestURL string `form:"requestUrl"`
	Method     string `form:"requestMethod"`
	EventType  string `form:"eventType"`
}

type OperationLogListResponse struct {
	LogID         string `json:"logId"`
	Username      string `json:"username"`
	RealName      string `json:"realName"`
	RequestMethod string `json:"requestMethod"`
	RequestURL    string `json:"requestUrl"`
	HTTPStatus    int    `json:"httpStatus"`
	Status        int    `json:"status"`
	DurationMs    int64  `json:"durationMs"`
	IP            string `json:"ip"`
	CreateDate    string `json:"createDate"`
}

type OperationLogDetailResponse struct {
	OperationLogListResponse
	UserID            string `json:"userId"`
	QueryParams       string `json:"queryParams"`
	QueryTruncated    bool   `json:"queryTruncated"`
	RequestBody       string `json:"requestBody"`
	ResponseBody      string `json:"responseBody"`
	RequestTruncated  bool   `json:"requestTruncated"`
	ResponseTruncated bool   `json:"responseTruncated"`
	UserAgent         string `json:"userAgent"`
	ContentType       string `json:"contentType"`
}

type LoginLogListResponse struct {
	LogID      string `json:"logId"`
	Username   string `json:"username"`
	EventType  string `json:"eventType"`
	HTTPStatus int    `json:"httpStatus"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"durationMs"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
	CreateDate string `json:"createDate"`
}

type LoginLogDetailResponse struct {
	LoginLogListResponse
	UserID            string `json:"userId"`
	ResponseBody      string `json:"responseBody"`
	ResponseTruncated bool   `json:"responseTruncated"`
	ContentType       string `json:"contentType"`
}

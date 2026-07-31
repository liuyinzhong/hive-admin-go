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
	LogID         string `json:"logId" example:"cc1a8564-37e7-47df-ad60-6c0a7f199d31"` // 日志ID
	Username      string `json:"username" example:"admin"`                             // 用户名
	RealName      string `json:"realName" example:"管理员"`                               // 用户真实姓名
	RequestMethod string `json:"requestMethod" example:"POST"`                         // 请求方法
	RequestURL    string `json:"requestUrl" example:"/api/system/users"`               // 请求URL
	HTTPStatus    int    `json:"httpStatus" example:"200"`                             // HTTP状态码
	Status        int    `json:"status" example:"1"`                                   // 操作状态 0失败 1成功
	DurationMs    int64  `json:"durationMs" example:"35"`                              // 耗时(毫秒)
	IP            string `json:"ip" example:"192.168.1.100"`                           // 客户端IP
	CreateDate    string `json:"createDate" example:"2026-07-31 15:30:26"`             // 创建时间
}

type OperationLogDetailResponse struct {
	OperationLogListResponse
	UserID            string `json:"userId" example:"cc1a8564-37e7-47df-ad60-6c0a7f199d31"` // 用户ID
	QueryParams       string `json:"queryParams" example:"page=1&pageSize=20"`              // 查询参数
	QueryTruncated    bool   `json:"queryTruncated" example:"false"`                        // 查询参数是否被截断
	RequestBody       string `json:"requestBody" example:"{\"username\":\"test\"}"`         // 请求体
	ResponseBody      string `json:"responseBody" example:"{\"code\":0,\"data\":null}"`     // 响应体
	RequestTruncated  bool   `json:"requestTruncated" example:"false"`                      // 请求体是否被截断
	ResponseTruncated bool   `json:"responseTruncated" example:"false"`                     // 响应体是否被截断
	UserAgent         string `json:"userAgent" example:"Mozilla/5.0"`                       // 用户代理
	ContentType       string `json:"contentType" example:"application/json"`                // 内容类型
}

type LoginLogListResponse struct {
	LogID      string `json:"logId" example:"cc1a8564-37e7-47df-ad60-6c0a7f199d31"` // 日志ID
	Username   string `json:"username" example:"admin"`                             // 用户名
	EventType  string `json:"eventType" example:"login"`                            // 事件类型 login登录 logout退出
	HTTPStatus int    `json:"httpStatus" example:"200"`                             // HTTP状态码
	Status     int    `json:"status" example:"1"`                                   // 操作状态 0失败 1成功
	DurationMs int64  `json:"durationMs" example:"120"`                             // 耗时(毫秒)
	IP         string `json:"ip" example:"192.168.1.100"`                           // 客户端IP
	UserAgent  string `json:"userAgent" example:"Mozilla/5.0"`                      // 用户代理
	CreateDate string `json:"createDate" example:"2026-07-31 15:30:26"`             // 创建时间
}

type LoginLogDetailResponse struct {
	LoginLogListResponse
	UserID            string `json:"userId" example:"cc1a8564-37e7-47df-ad60-6c0a7f199d31"` // 用户ID
	ResponseBody      string `json:"responseBody" example:"{\"code\":0,\"data\":null}"`     // 响应体
	ResponseTruncated bool   `json:"responseTruncated" example:"false"`                     // 响应体是否被截断
	ContentType       string `json:"contentType" example:"application/json"`                // 内容类型
}

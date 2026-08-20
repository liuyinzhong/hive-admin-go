package models

import "time"

const (
	DownloadTaskStatusPending   = "pending"
	DownloadTaskStatusRunning   = "running"
	DownloadTaskStatusSucceeded = "succeeded"
	DownloadTaskStatusFailed    = "failed"

	DownloadTaskTypeInventoryBalance = "inventoryBalance"
	DownloadTaskTypeDevTask          = "devTask"
	DownloadTaskTypeLoginLog         = "loginLog"
)

// SysDownloadTask 保存当前用户发起的异步导出任务及生成文件元数据。
type SysDownloadTask struct {
	ID             string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	CreatorID      string     `gorm:"column:creator_id;type:char(36);not null" json:"creatorId"`
	TaskType       string     `gorm:"column:task_type;type:varchar(32);not null" json:"taskType"`
	TaskName       string     `gorm:"column:task_name;type:varchar(128);not null" json:"taskName"`
	SourceModule   string     `gorm:"column:source_module;type:varchar(64);not null" json:"sourceModule"`
	RequestPayload string     `gorm:"column:request_payload;type:longtext;not null" json:"-"`
	Status         string     `gorm:"column:status;type:varchar(16);not null" json:"status"`
	TotalRows      int64      `gorm:"column:total_rows;type:bigint;not null;default:0" json:"totalRows"`
	ProcessedRows  int64      `gorm:"column:processed_rows;type:bigint;not null;default:0" json:"processedRows"`
	Progress       int        `gorm:"column:progress;type:tinyint;not null;default:0" json:"progress"`
	FileName       *string    `gorm:"column:file_name;type:varchar(255)" json:"fileName"`
	FilePath       *string    `gorm:"column:file_path;type:varchar(512)" json:"-"`
	FileSize       int64      `gorm:"column:file_size;type:bigint;not null;default:0" json:"fileSize"`
	ErrorMessage   *string    `gorm:"column:error_message;type:varchar(512)" json:"errorMessage"`
	CompletedDate  *time.Time `gorm:"column:completed_date" json:"completedDate"`
	ExpireDate     *time.Time `gorm:"column:expire_date" json:"expireDate"`
	CreateDate     *time.Time `gorm:"column:create_date;not null" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date;not null" json:"updateDate"`
}

func (SysDownloadTask) TableName() string {
	return "sys_download_task"
}

type DownloadTaskListRequest struct {
	Page            int    `form:"page" example:"1"`
	PageSize        int    `form:"pageSize" example:"20"`
	TaskName        string `form:"taskName" example:"库存余额"`
	Status          string `form:"status" example:"succeeded"`
	CreateDateStart string `form:"createDateStart" example:"2026-08-01 00:00:00"`
	CreateDateEnd   string `form:"createDateEnd" example:"2026-08-08 23:59:59"`
}

type DownloadTaskCreatedResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DownloadPreviewURLResponse 预览链接生成响应，返回前端可拼接的相对路径。
type DownloadPreviewURLResponse struct {
	PreviewURL string `json:"previewUrl" example:"/api/public/downloads/preview/eyJhbGciOiJIUzI1NiJ9.xxx"`
}

type DownloadTaskChangedEvent struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	TotalRows     int64  `json:"totalRows"`
	ProcessedRows int64  `json:"processedRows"`
	Progress      int    `json:"progress"`
}

type DownloadTaskResponse struct {
	ID            string  `json:"id"`
	TaskName      string  `json:"taskName"`
	SourceModule  string  `json:"sourceModule"`
	Status        string  `json:"status"`
	TotalRows     int64   `json:"totalRows"`
	ProcessedRows int64   `json:"processedRows"`
	Progress      int     `json:"progress"`
	FileName      *string `json:"fileName"`
	FileSize      int64   `json:"fileSize"`
	ErrorMessage  *string `json:"errorMessage"`
	CompletedDate *string `json:"completedDate"`
	ExpireDate    *string `json:"expireDate"`
	CreateDate    string  `json:"createDate"`
	UpdateDate    string  `json:"updateDate"`
}

type InventoryBalanceExportRequest struct {
	WarehouseID  string                 `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkuCode      string                 `json:"skuCode" example:"SKU000001"`
	BatchNo      string                 `json:"batchNo" example:"B20260731001"`
	OnlyPositive bool                   `json:"onlyPositive" example:"true"`
	Sorts        string                 `json:"sorts" example:"updateDate,desc"`
	Filename     string                 `json:"filename" example:"库存余额导出.xlsx"`
	SheetName    string                 `json:"sheetName" example:"库存余额"`
	Columns      []DownloadExportColumn `json:"columns"`
	IsHeader     *bool                  `json:"isHeader" example:"true"`
	IsTitle      *bool                  `json:"isTitle" example:"true"`
	Original     *bool                  `json:"original" example:"false"`
}

type DevTaskExportRequest struct {
	ProjectID    string                 `json:"projectId" example:"550e8400-e29b-41d4-a716-446655440000"`
	VersionID    string                 `json:"versionId" example:"550e8400-e29b-41d4-a716-446655440000"`
	TaskTitle    string                 `json:"taskTitle" example:"下载中心开发"`
	TaskStatuses []int                  `json:"taskStatus" example:"1,2"`
	Sorts        string                 `json:"sorts" example:"createDate,desc"`
	Filename     string                 `json:"filename" example:"开发任务导出.xlsx"`
	SheetName    string                 `json:"sheetName" example:"任务管理"`
	Columns      []DownloadExportColumn `json:"columns"`
	IsHeader     *bool                  `json:"isHeader" example:"true"`
	IsTitle      *bool                  `json:"isTitle" example:"true"`
	Original     *bool                  `json:"original" example:"false"`
}

// DownloadExportColumn 是前端导出列的最小描述；字段值只用于匹配后端白名单，不直接拼接 SQL。
type DownloadExportColumn struct {
	Field string `json:"field" example:"taskTitle"`
	Title string `json:"title" example:"任务标题"`
	Width int    `json:"width,omitempty" example:"200"`
}

type LoginLogExportRequest struct {
	Username  string                 `json:"username" example:"admin"`
	EventType string                 `json:"eventType" example:"login"`
	Status    *int                   `json:"status" example:"1"`
	IP        string                 `json:"ip" example:"192.168.1.100"`
	StartDate string                 `json:"startDate" example:"2026-08-01 00:00:00"`
	EndDate   string                 `json:"endDate" example:"2026-08-20 23:59:59"`
	Sorts     string                 `json:"sorts" example:"createDate,desc"`
	Filename  string                 `json:"filename" example:"登录日志导出.xlsx"`
	SheetName string                 `json:"sheetName" example:"登录日志"`
	Columns   []DownloadExportColumn `json:"columns"`
	IsHeader  *bool                  `json:"isHeader" example:"true"`
	IsTitle   *bool                  `json:"isTitle" example:"true"`
	Original  *bool                  `json:"original" example:"false"`
}

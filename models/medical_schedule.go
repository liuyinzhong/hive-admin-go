package models

import "time"

const (
	MedScheduleStatusDraft = iota
	MedScheduleStatusPublished
	MedScheduleStatusStopped
	MedScheduleStatusFinished
)

const (
	MedScheduleBatchStatusProcessing = iota
	MedScheduleBatchStatusCompleted
)

type MedScheduleTemplate struct {
	TemplateID       string     `gorm:"column:template_id;type:char(36);primaryKey" json:"templateId"`
	TemplateName     string     `gorm:"column:template_name;type:varchar(64)" json:"templateName"`
	DoctorID         string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DepartmentID     string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	RegistrationType string     `gorm:"column:registration_type;type:varchar(36)" json:"registrationType"`
	Weekday          int        `gorm:"column:weekday;type:tinyint" json:"-"`
	Weekdays         []int      `gorm:"-" json:"weekdays"`
	StartTime        string     `gorm:"column:start_time;type:time" json:"startTime"`
	EndTime          string     `gorm:"column:end_time;type:time" json:"endTime"`
	DefaultSlotQuota int        `gorm:"column:default_slot_quota;type:int" json:"defaultSlotQuota"`
	SlotQuotaConfig  *string    `gorm:"column:slot_quota_config;type:longtext" json:"-"`
	TotalQuota       int        `gorm:"column:total_quota;type:int" json:"totalQuota"`
	EffectiveDate    time.Time  `gorm:"column:effective_date;type:date" json:"effectiveDate"`
	ExpiryDate       *time.Time `gorm:"column:expiry_date;type:date" json:"expiryDate"`
	Status           int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark           *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID        *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID        *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag          int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedScheduleTemplate) TableName() string { return "med_schedule_template" }

type MedScheduleTemplateWeekday struct {
	TemplateID string    `gorm:"column:template_id;type:char(36);primaryKey" json:"templateId"`
	Weekday    int       `gorm:"column:weekday;type:tinyint;primaryKey" json:"weekday"`
	CreateDate time.Time `gorm:"column:create_date" json:"createDate"`
}

func (MedScheduleTemplateWeekday) TableName() string { return "med_schedule_template_weekday" }

type MedScheduleGenerationBatch struct {
	BatchID        string     `gorm:"column:batch_id;type:char(36);primaryKey" json:"batchId"`
	IdempotencyKey string     `gorm:"column:idempotency_key;type:varchar(64)" json:"idempotencyKey"`
	RequestHash    string     `gorm:"column:request_hash;type:char(64)" json:"requestHash"`
	TemplateIDs    string     `gorm:"column:template_ids;type:longtext" json:"templateIds"`
	StartDate      time.Time  `gorm:"column:start_date;type:date" json:"startDate"`
	EndDate        time.Time  `gorm:"column:end_date;type:date" json:"endDate"`
	Status         int        `gorm:"column:status;type:tinyint" json:"status"`
	GeneratedCount int        `gorm:"column:generated_count;type:int" json:"generatedCount"`
	SkippedCount   int        `gorm:"column:skipped_count;type:int" json:"skippedCount"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedScheduleGenerationBatch) TableName() string { return "med_schedule_generation_batch" }

type MedSchedule struct {
	ScheduleID        string     `gorm:"column:schedule_id;type:char(36);primaryKey" json:"scheduleId"`
	TemplateID        *string    `gorm:"column:template_id;type:char(36)" json:"templateId"`
	GenerationBatchID *string    `gorm:"column:generation_batch_id;type:char(36)" json:"generationBatchId"`
	DoctorID          string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DepartmentID      string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	RegistrationType  string     `gorm:"column:registration_type;type:varchar(36)" json:"registrationType"`
	ScheduleDate      time.Time  `gorm:"column:schedule_date;type:date" json:"scheduleDate"`
	StartTime         string     `gorm:"column:start_time;type:time" json:"startTime"`
	EndTime           string     `gorm:"column:end_time;type:time" json:"endTime"`
	FeeRuleID         *string    `gorm:"column:fee_rule_id;type:char(36)" json:"feeRuleId"`
	FeeRuleVersion    *int       `gorm:"column:fee_rule_version;type:int" json:"feeRuleVersion"`
	FeeAmount         *string    `gorm:"column:fee_amount;type:decimal(10,2)" json:"feeAmount"`
	DefaultSlotQuota  int        `gorm:"column:default_slot_quota;type:int" json:"defaultSlotQuota"`
	TotalQuota        int        `gorm:"column:total_quota;type:int" json:"totalQuota"`
	BookedQuota       int        `gorm:"column:booked_quota;type:int;default:0" json:"bookedQuota"`
	Status            int        `gorm:"column:status;type:tinyint;default:0" json:"status"`
	StopReason        *string    `gorm:"column:stop_reason;type:varchar(512)" json:"stopReason"`
	PublishedAt       *time.Time `gorm:"column:published_at" json:"publishedAt"`
	StoppedAt         *time.Time `gorm:"column:stopped_at" json:"stoppedAt"`
	FinishedAt        *time.Time `gorm:"column:finished_at" json:"finishedAt"`
	Remark            *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID         *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID         *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag           int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedSchedule) TableName() string { return "med_schedule" }

type ScheduleTemplateListRequest struct {
	Page         int    `form:"page" example:"1"`
	PageSize     int    `form:"pageSize" example:"20"`
	DoctorID     string `form:"doctorId"`
	DepartmentID string `form:"departmentId"`
	Weekday      *int   `form:"weekday"`
	Status       *int   `form:"status"`
	Sorts        string `form:"sorts"`
}

type ScheduleTemplateBaseRequest struct {
	TemplateName     string                     `json:"templateName" binding:"required,max=64"`
	DoctorID         string                     `json:"doctorId" binding:"required"`
	DepartmentID     string                     `json:"departmentId" binding:"required"`
	RegistrationType string                     `json:"registrationType" binding:"required"`
	StartTime        string                     `json:"startTime" binding:"required"`
	EndTime          string                     `json:"endTime" binding:"required"`
	DefaultSlotQuota int                        `json:"defaultSlotQuota" binding:"required,min=1,max=99"`
	SlotQuotaConfig  []ScheduleSlotQuotaRequest `json:"slotQuotaConfig" binding:"omitempty,max=48"`
	EffectiveDate    string                     `json:"effectiveDate" binding:"required"`
	ExpiryDate       *string                    `json:"expiryDate"`
	Status           int                        `json:"status" binding:"oneof=0 1"`
	Remark           *string                    `json:"remark" binding:"omitempty,max=512"`
}

type SaveScheduleTemplateRequest struct {
	ScheduleTemplateBaseRequest
	// Weekdays 星期多选值，每项范围为1（周一）至7（周日）。
	Weekdays []int `json:"weekdays" binding:"required,min=1,max=7,dive,min=1,max=7"`
}

type ScheduleTemplateResponse struct {
	TemplateID       string                     `json:"templateId"`
	TemplateName     string                     `json:"templateName"`
	DoctorID         string                     `json:"doctorId"`
	DoctorNo         string                     `json:"doctorNo"`
	DoctorName       string                     `json:"doctorName"`
	DepartmentID     string                     `json:"departmentId"`
	DepartmentCode   string                     `json:"departmentCode"`
	DepartmentName   string                     `json:"departmentName"`
	RegistrationType string                     `json:"registrationType"`
	Weekdays         []int                      `json:"weekdays"`
	StartTime        string                     `json:"startTime"`
	EndTime          string                     `json:"endTime"`
	DefaultSlotQuota int                        `json:"defaultSlotQuota"`
	SlotQuotaConfig  []ScheduleSlotQuotaRequest `json:"slotQuotaConfig"`
	TotalQuota       int                        `json:"totalQuota"`
	EffectiveDate    string                     `json:"effectiveDate"`
	ExpiryDate       *string                    `json:"expiryDate"`
	Status           int                        `json:"status"`
	Remark           *string                    `json:"remark"`
	CreateDate       *string                    `json:"createDate"`
	UpdateDate       *string                    `json:"updateDate"`
}

type ScheduleListRequest struct {
	Page             int    `form:"page" example:"1"`
	PageSize         int    `form:"pageSize" example:"20"`
	DoctorID         string `form:"doctorId"`
	DepartmentID     string `form:"departmentId"`
	RegistrationType string `form:"registrationType"`
	StartDate        string `form:"startDate"`
	EndDate          string `form:"endDate"`
	Status           *int   `form:"status"`
	Sorts            string `form:"sorts"`
}

type SaveScheduleRequest struct {
	DoctorID         string                     `json:"doctorId" binding:"required"`
	DepartmentID     string                     `json:"departmentId" binding:"required"`
	RegistrationType string                     `json:"registrationType" binding:"required"`
	ScheduleDate     string                     `json:"scheduleDate" binding:"required"`
	StartTime        string                     `json:"startTime" binding:"required"`
	EndTime          string                     `json:"endTime" binding:"required"`
	DefaultSlotQuota int                        `json:"defaultSlotQuota" binding:"required,min=1,max=99"`
	SlotQuotaConfig  []ScheduleSlotQuotaRequest `json:"slotQuotaConfig" binding:"omitempty,max=48"`
	Remark           *string                    `json:"remark" binding:"omitempty,max=512"`
}

type GenerateSchedulesRequest struct {
	IdempotencyKey string   `json:"idempotencyKey" binding:"required,max=64"`
	TemplateIDs    []string `json:"templateIds" binding:"required,min=1,max=100"`
	StartDate      string   `json:"startDate" binding:"required"`
	EndDate        string   `json:"endDate" binding:"required"`
}

type PublishSchedulesRequest struct {
	ScheduleIDs []string `json:"scheduleIds" binding:"required,min=1,max=100"`
}

type StopScheduleRequest struct {
	Reason string `json:"reason" binding:"required,max=512"`
}

type ScheduleResponse struct {
	ScheduleID        string                 `json:"scheduleId"`
	TemplateID        *string                `json:"templateId"`
	GenerationBatchID *string                `json:"generationBatchId"`
	DoctorID          string                 `json:"doctorId"`
	DoctorNo          string                 `json:"doctorNo"`
	DoctorName        string                 `json:"doctorName"`
	DepartmentID      string                 `json:"departmentId"`
	DepartmentCode    string                 `json:"departmentCode"`
	DepartmentName    string                 `json:"departmentName"`
	RegistrationType  string                 `json:"registrationType"`
	ScheduleDate      string                 `json:"scheduleDate"`
	StartTime         string                 `json:"startTime"`
	EndTime           string                 `json:"endTime"`
	FeeRuleID         *string                `json:"feeRuleId"`
	FeeRuleVersion    *int                   `json:"feeRuleVersion"`
	FeeAmount         *string                `json:"feeAmount"`
	FeeSnapshotStatus string                 `json:"feeSnapshotStatus"`
	DefaultSlotQuota  int                    `json:"defaultSlotQuota"`
	TotalQuota        int                    `json:"totalQuota"`
	BookedQuota       int                    `json:"bookedQuota"`
	RemainingQuota    int                    `json:"remainingQuota"`
	Status            int                    `json:"status"`
	StopReason        *string                `json:"stopReason"`
	PublishedAt       *string                `json:"publishedAt"`
	StoppedAt         *string                `json:"stoppedAt"`
	FinishedAt        *string                `json:"finishedAt"`
	Remark            *string                `json:"remark"`
	CreateDate        *string                `json:"createDate"`
	UpdateDate        *string                `json:"updateDate"`
	Slots             []ScheduleSlotResponse `json:"slots"`
}

type GenerateSchedulesResponse struct {
	BatchID        string   `json:"batchId"`
	Idempotent     bool     `json:"idempotent"`
	GeneratedCount int      `json:"generatedCount"`
	SkippedCount   int      `json:"skippedCount"`
	ScheduleIDs    []string `json:"scheduleIds"`
}

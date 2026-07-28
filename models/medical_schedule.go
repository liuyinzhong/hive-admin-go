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
	TemplateID       string                     `json:"templateId" example:"550e8400-e29b-41d4-a716-446655440000"`   // 模板ID
	TemplateName     string                     `json:"templateName" example:"门诊排班模板"`                               // 模板名称
	DoctorID         string                     `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 医生ID
	DoctorNo         string                     `json:"doctorNo" example:"DOC001"`                                   // 医生编号
	DoctorName       string                     `json:"doctorName" example:"张三"`                                     // 医生姓名
	DepartmentID     string                     `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 科室ID
	DepartmentCode   string                     `json:"departmentCode" example:"NK"`                                 // 科室编码
	DepartmentName   string                     `json:"departmentName" example:"内科"`                                 // 科室名称
	RegistrationType string                     `json:"registrationType" example:"普通"`                               // 号别
	Weekdays         []int                      `json:"weekdays"`                                                    // 出诊星期，每项范围为1（周一）至7（周日）
	StartTime        string                     `json:"startTime" example:"09:00"`                                   // 开始时间
	EndTime          string                     `json:"endTime" example:"17:00"`                                     // 结束时间
	DefaultSlotQuota int                        `json:"defaultSlotQuota" example:"15"`                               // 默认号源配额
	SlotQuotaConfig  []ScheduleSlotQuotaRequest `json:"slotQuotaConfig"`                                             // 时段号源配置
	TotalQuota       int                        `json:"totalQuota" example:"60"`                                     // 总号源配额
	EffectiveDate    string                     `json:"effectiveDate" example:"2026-01-15"`                          // 生效日期
	ExpiryDate       *string                    `json:"expiryDate" example:"2026-12-31"`                             // 失效日期
	Status           int                        `json:"status" example:"1"`                                          // 状态(0-禁用 1-启用)
	Remark           *string                    `json:"remark" example:"备注信息"`                                       // 备注
	CreateDate       *string                    `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate       *string                    `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
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
	ScheduleID        string                 `json:"scheduleId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 排班ID
	TemplateID        *string                `json:"templateId" example:"550e8400-e29b-41d4-a716-446655440000"`        // 模板ID
	GenerationBatchID *string                `json:"generationBatchId" example:"550e8400-e29b-41d4-a716-446655440000"` // 生成批次ID
	DoctorID          string                 `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"`          // 医生ID
	DoctorNo          string                 `json:"doctorNo" example:"DOC001"`                                        // 医生编号
	DoctorName        string                 `json:"doctorName" example:"张三"`                                          // 医生姓名
	DepartmentID      string                 `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"`      // 科室ID
	DepartmentCode    string                 `json:"departmentCode" example:"NK"`                                      // 科室编码
	DepartmentName    string                 `json:"departmentName" example:"内科"`                                      // 科室名称
	RegistrationType  string                 `json:"registrationType" example:"普通"`                                    // 号别
	ScheduleDate      string                 `json:"scheduleDate" example:"2026-01-15"`                                // 排班日期
	StartTime         string                 `json:"startTime" example:"09:00"`                                        // 开始时间
	EndTime           string                 `json:"endTime" example:"17:00"`                                          // 结束时间
	FeeRuleID         *string                `json:"feeRuleId" example:"550e8400-e29b-41d4-a716-446655440000"`         // 费用规则ID
	FeeRuleVersion    *int                   `json:"feeRuleVersion" example:"1"`                                       // 费用规则版本
	FeeAmount         *string                `json:"feeAmount" example:"50.00"`                                        // 费用金额
	FeeSnapshotStatus string                 `json:"feeSnapshotStatus" example:"valid"`                                // 费用快照状态
	DefaultSlotQuota  int                    `json:"defaultSlotQuota" example:"15"`                                    // 默认号源配额
	TotalQuota        int                    `json:"totalQuota" example:"60"`                                          // 总号源配额
	BookedQuota       int                    `json:"bookedQuota" example:"5"`                                          // 已预约号源
	RemainingQuota    int                    `json:"remainingQuota" example:"55"`                                      // 剩余号源
	Status            int                    `json:"status" example:"1"`                                               // 状态(0-草稿 1-已发布 2-已停诊 3-已结束)
	StopReason        *string                `json:"stopReason" example:"医生请假"`                                        // 停诊原因
	PublishedAt       *string                `json:"publishedAt" example:"2026-01-15 09:00:00"`                        // 发布时间
	StoppedAt         *string                `json:"stoppedAt" example:"2026-01-15 09:00:00"`                          // 停诊时间
	FinishedAt        *string                `json:"finishedAt" example:"2026-01-15 09:00:00"`                         // 结束时间
	Remark            *string                `json:"remark" example:"备注信息"`                                            // 备注
	CreateDate        *string                `json:"createDate" example:"2026-01-15 09:00:00"`                         // 创建时间
	UpdateDate        *string                `json:"updateDate" example:"2026-01-15 09:00:00"`                         // 更新时间
	Slots             []ScheduleSlotResponse `json:"slots"`                                                            // 时段列表
}

type GenerateSchedulesResponse struct {
	BatchID        string   `json:"batchId" example:"550e8400-e29b-41d4-a716-446655440000"`           // 批次ID
	Idempotent     bool     `json:"idempotent" example:"true"`                                        // 是否幂等
	GeneratedCount int      `json:"generatedCount" example:"30"`                                      // 生成数量
	SkippedCount   int      `json:"skippedCount" example:"5"`                                         // 跳过数量
	ScheduleIDs    []string `json:"scheduleIds" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"` // 排班ID列表
}

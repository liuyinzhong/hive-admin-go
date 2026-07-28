package models

import "time"

const (
	MedScheduleAutoTaskStatusSuccess = iota
	MedScheduleAutoTaskStatusPartial
	MedScheduleAutoTaskStatusFailed
	MedScheduleAutoTaskStatusProcessing
)

const (
	MedScheduleAutoTaskTypePublish  = "publish"
	MedScheduleAutoTaskTypeGenerate = "generate"
)

type ScheduleSlotQuotaRequest struct {
	StartTime string `json:"startTime" binding:"required" example:"09:00"` // 时段开始时间
	Quota     int    `json:"quota" binding:"min=0,max=99" example:"10"`    // 号源配额
}

type MedScheduleSlot struct {
	SlotID      string     `gorm:"column:slot_id;type:char(36);primaryKey" json:"slotId"`
	ScheduleID  string     `gorm:"column:schedule_id;type:char(36)" json:"scheduleId"`
	StartTime   string     `gorm:"column:start_time;type:time" json:"startTime"`
	EndTime     string     `gorm:"column:end_time;type:time" json:"endTime"`
	Quota       int        `gorm:"column:quota;type:int" json:"quota"`
	BookedQuota int        `gorm:"column:booked_quota;type:int;default:0" json:"bookedQuota"`
	CreatorID   *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID   *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate  *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate  *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag     int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedScheduleSlot) TableName() string { return "med_schedule_slot" }

type MedScheduleAutoTask struct {
	TaskID             string     `gorm:"column:task_id;type:char(36);primaryKey" json:"taskId"`
	TaskKey            string     `gorm:"column:task_key;type:varchar(96)" json:"taskKey"`
	TaskType           string     `gorm:"column:task_type;type:varchar(16)" json:"taskType"`
	TargetWeekStart    time.Time  `gorm:"column:target_week_start;type:date" json:"targetWeekStart"`
	TargetWeekEnd      time.Time  `gorm:"column:target_week_end;type:date" json:"targetWeekEnd"`
	Status             int        `gorm:"column:status;type:tinyint" json:"status"`
	SuccessDoctorCount int        `gorm:"column:success_doctor_count;type:int" json:"successDoctorCount"`
	FailureDoctorCount int        `gorm:"column:failure_doctor_count;type:int" json:"failureDoctorCount"`
	Details            *string    `gorm:"column:details;type:longtext" json:"details"`
	ExecutedAt         time.Time  `gorm:"column:executed_at" json:"executedAt"`
	CreateDate         *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate         *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedScheduleAutoTask) TableName() string { return "med_schedule_auto_task" }

type ScheduleSlotResponse struct {
	SlotID         string `json:"slotId" example:"550e8400-e29b-41d4-a716-446655440000"` // 时段ID
	StartTime      string `json:"startTime" example:"09:00"`                             // 开始时间
	EndTime        string `json:"endTime" example:"09:30"`                               // 结束时间
	Quota          int    `json:"quota" example:"5"`                                     // 号源配额
	BookedQuota    int    `json:"bookedQuota" example:"2"`                               // 已预约数量
	RemainingQuota int    `json:"remainingQuota" example:"3"`                            // 剩余可预约数量
	BookingStatus  string `json:"bookingStatus" example:"available"`                     // 预约状态
	CanBook        bool   `json:"canBook" example:"true"`                                // 是否可预约
}

type ScheduleAutoTaskListRequest struct {
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"pageSize" example:"20"`
	TaskType  string `form:"taskType"`
	Status    *int   `form:"status"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	Sorts     string `form:"sorts"`
}

type ScheduleAutoTaskFailure struct {
	DoctorID   string `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"` // 医生ID
	DoctorName string `json:"doctorName" example:"张医生"`                                // 医生姓名
	Reason     string `json:"reason" example:"该医生该时段已有排班"`                             // 失败原因
}

type ScheduleAutoTaskResponse struct {
	TaskID             string                    `json:"taskId" example:"550e8400-e29b-41d4-a716-446655440000"` // 任务ID
	TaskType           string                    `json:"taskType" example:"publish"`                            // 任务类型
	TargetWeekStart    string                    `json:"targetWeekStart" example:"2026-01-13"`                  // 目标周起始日期
	TargetWeekEnd      string                    `json:"targetWeekEnd" example:"2026-01-19"`                    // 目标周结束日期
	Status             int                       `json:"status" example:"0"`                                    // 状态(0-成功 1-部分成功 2-失败 3-处理中)
	SuccessDoctorCount int                       `json:"successDoctorCount" example:"10"`                       // 成功医生数
	FailureDoctorCount int                       `json:"failureDoctorCount" example:"1"`                        // 失败医生数
	Failures           []ScheduleAutoTaskFailure `json:"failures"`                                              // 失败详情
	ExecutedAt         string                    `json:"executedAt" example:"2026-01-15 09:00:00"`              // 执行时间
}

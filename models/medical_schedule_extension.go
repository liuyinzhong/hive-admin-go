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
	StartTime string `json:"startTime" binding:"required"`
	Quota     int    `json:"quota" binding:"min=0,max=99"`
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
	SlotID         string `json:"slotId"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	Quota          int    `json:"quota"`
	BookedQuota    int    `json:"bookedQuota"`
	RemainingQuota int    `json:"remainingQuota"`
	BookingStatus  string `json:"bookingStatus"`
	CanBook        bool   `json:"canBook"`
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
	DoctorID   string `json:"doctorId"`
	DoctorName string `json:"doctorName"`
	Reason     string `json:"reason"`
}

type ScheduleAutoTaskResponse struct {
	TaskID             string                    `json:"taskId"`
	TaskType           string                    `json:"taskType"`
	TargetWeekStart    string                    `json:"targetWeekStart"`
	TargetWeekEnd      string                    `json:"targetWeekEnd"`
	Status             int                       `json:"status"`
	SuccessDoctorCount int                       `json:"successDoctorCount"`
	FailureDoctorCount int                       `json:"failureDoctorCount"`
	Failures           []ScheduleAutoTaskFailure `json:"failures"`
	ExecutedAt         string                    `json:"executedAt"`
}

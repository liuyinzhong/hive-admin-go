package models

import "time"

const (
	MedRegistrationMethodOnSite      = 0
	MedRegistrationMethodAppointment = 10
)

const (
	MedRegistrationStatusPendingPayment = 0
	MedRegistrationStatusPaid           = 10
	MedRegistrationStatusCheckedIn      = 30
	MedRegistrationStatusCompleted      = 50
	MedRegistrationStatusCanceled       = 60
	MedRegistrationStatusNoShow         = 70
	MedRegistrationStatusRefundStarted  = 80
	MedRegistrationStatusRefunding      = 90
	MedRegistrationStatusRefunded       = 100
)

type MedRegistration struct {
	RegistrationID       string     `gorm:"column:registration_id;type:char(36);primaryKey" json:"registrationId"`
	RegistrationNo       string     `gorm:"column:registration_no;type:varchar(32)" json:"registrationNo"`
	PatientID            string     `gorm:"column:patient_id;type:char(36)" json:"patientId"`
	PatientNo            string     `gorm:"column:patient_no;type:varchar(32)" json:"patientNo"`
	PatientName          string     `gorm:"column:patient_name;type:varchar(64)" json:"patientName"`
	PatientGender        string     `gorm:"column:patient_gender;type:varchar(36)" json:"patientGender"`
	PatientBirthDate     time.Time  `gorm:"column:patient_birth_date;type:date" json:"patientBirthDate"`
	PatientIDType        string     `gorm:"column:patient_id_type;type:varchar(36)" json:"patientIdType"`
	PatientIDNumber      string     `gorm:"column:patient_id_number;type:varchar(128)" json:"patientIdNumber"`
	PatientPhone         string     `gorm:"column:patient_phone;type:varchar(20)" json:"patientPhone"`
	ScheduleID           string     `gorm:"column:schedule_id;type:char(36)" json:"scheduleId"`
	SlotID               string     `gorm:"column:slot_id;type:char(36)" json:"slotId"`
	DoctorID             string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DoctorName           string     `gorm:"column:doctor_name;type:varchar(64)" json:"doctorName"`
	DepartmentID         string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	DepartmentName       string     `gorm:"column:department_name;type:varchar(64)" json:"departmentName"`
	ScheduleDate         time.Time  `gorm:"column:schedule_date;type:date" json:"scheduleDate"`
	StartTime            string     `gorm:"column:start_time;type:time" json:"startTime"`
	EndTime              string     `gorm:"column:end_time;type:time" json:"endTime"`
	RegistrationType     string     `gorm:"column:registration_type;type:varchar(36)" json:"registrationType"`
	RegistrationTypeName string     `gorm:"column:registration_type_name;type:varchar(128)" json:"registrationTypeName"`
	RegistrationMethod   int        `gorm:"column:registration_method;type:smallint" json:"registrationMethod"`
	FeeRuleID            *string    `gorm:"column:fee_rule_id;type:char(36)" json:"feeRuleId"`
	FeeRuleVersion       *int       `gorm:"column:fee_rule_version;type:int" json:"feeRuleVersion"`
	FeeAmount            string     `gorm:"column:fee_amount;type:decimal(10,2)" json:"feeAmount"`
	Status               int        `gorm:"column:status;type:smallint" json:"status"`
	Remark               *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID            *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID            *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate           *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate           *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedRegistration) TableName() string { return "med_registration" }

type MedRegistrationLog struct {
	LogID          string    `gorm:"column:log_id;type:char(36);primaryKey" json:"logId"`
	RegistrationID string    `gorm:"column:registration_id;type:char(36)" json:"registrationId"`
	FromStatus     *int      `gorm:"column:from_status;type:smallint" json:"fromStatus"`
	ToStatus       int       `gorm:"column:to_status;type:smallint" json:"toStatus"`
	OperatorID     *string   `gorm:"column:operator_id;type:char(36)" json:"operatorId"`
	OperatedAt     time.Time `gorm:"column:operated_at" json:"operatedAt"`
	Reason         *string   `gorm:"column:reason;type:varchar(512)" json:"reason"`
	RefundAmount   *string   `gorm:"column:refund_amount;type:decimal(10,2)" json:"refundAmount"`
}

func (MedRegistrationLog) TableName() string { return "med_registration_log" }

type RegistrationListRequest struct {
	Page               int    `form:"page" example:"1"`
	PageSize           int    `form:"pageSize" example:"20"`
	RegistrationNo     string `form:"registrationNo"`
	PatientKeyword     string `form:"patientKeyword"`
	StartDate          string `form:"startDate"`
	EndDate            string `form:"endDate"`
	DepartmentID       string `form:"departmentId"`
	DoctorID           string `form:"doctorId"`
	RegistrationType   string `form:"registrationType"`
	RegistrationMethod *int   `form:"registrationMethod"`
	Status             *int   `form:"status"`
	Sorts              string `form:"sorts"`
}

type CreateRegistrationRequest struct {
	PatientID          string  `json:"patientId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	SlotID             string  `json:"slotId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	RegistrationMethod *int    `json:"registrationMethod" binding:"required,oneof=0 10" example:"10"`
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"患者预约挂号"`
}

type RegistrationReasonRequest struct {
	Reason string `json:"reason" binding:"required,max=512"`
}

type RegistrationLifecycleResponse struct {
	LifecycleID  string  `json:"lifecycleId"`
	FromStatus   *int    `json:"fromStatus"`
	ToStatus     int     `json:"toStatus"`
	OperatorID   *string `json:"operatorId"`
	OperatorName *string `json:"operatorName"`
	OperatedAt   string  `json:"operatedAt"`
	Reason       *string `json:"reason"`
	RefundAmount *string `json:"refundAmount"`
}

type RegistrationResponse struct {
	RegistrationID       string                          `json:"registrationId"`
	RegistrationNo       string                          `json:"registrationNo"`
	PatientID            string                          `json:"patientId"`
	PatientNo            string                          `json:"patientNo"`
	PatientName          string                          `json:"patientName"`
	PatientGender        string                          `json:"patientGender"`
	PatientBirthDate     string                          `json:"patientBirthDate"`
	PatientIDType        string                          `json:"patientIdType"`
	PatientIDNumber      string                          `json:"patientIdNumber"`
	PatientPhone         string                          `json:"patientPhone"`
	ScheduleID           string                          `json:"scheduleId"`
	SlotID               string                          `json:"slotId"`
	DoctorID             string                          `json:"doctorId"`
	DoctorName           string                          `json:"doctorName"`
	DepartmentID         string                          `json:"departmentId"`
	DepartmentName       string                          `json:"departmentName"`
	ScheduleDate         string                          `json:"scheduleDate"`
	StartTime            string                          `json:"startTime"`
	EndTime              string                          `json:"endTime"`
	RegistrationType     string                          `json:"registrationType"`
	RegistrationTypeName string                          `json:"registrationTypeName"`
	RegistrationMethod   int                             `json:"registrationMethod"`
	FeeRuleID            *string                         `json:"feeRuleId"`
	FeeRuleVersion       *int                            `json:"feeRuleVersion"`
	FeeAmount            string                          `json:"feeAmount"`
	Status               int                             `json:"status"`
	Remark               *string                         `json:"remark"`
	CreatorID            *string                         `json:"creatorId"`
	CreateDate           *string                         `json:"createDate"`
	UpdateDate           *string                         `json:"updateDate"`
	LifecycleRecords     []RegistrationLifecycleResponse `json:"lifecycleRecords"`
}

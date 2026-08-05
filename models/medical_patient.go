package models

import "time"

type MedPatient struct {
	PatientID                string     `gorm:"column:patient_id;type:char(36);primaryKey" json:"patientId"`
	PatientNo                string     `gorm:"column:patient_no;type:varchar(32)" json:"patientNo"`
	Name                     string     `gorm:"column:name;type:varchar(64)" json:"name"`
	Gender                   string     `gorm:"column:gender;type:varchar(36)" json:"gender"`
	BirthDate                time.Time  `gorm:"column:birth_date;type:date" json:"birthDate"`
	IDType                   string     `gorm:"column:id_type;type:varchar(36)" json:"idType"`
	IDNumber                 string     `gorm:"column:id_number;type:varchar(128)" json:"idNumber"`
	Phone                    string     `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Address                  *string    `gorm:"column:address;type:varchar(512)" json:"address"`
	EmergencyContactName     *string    `gorm:"column:emergency_contact_name;type:varchar(64)" json:"emergencyContactName"`
	EmergencyContactRelation *string    `gorm:"column:emergency_contact_relation;type:varchar(64)" json:"emergencyContactRelation"`
	EmergencyContactPhone    *string    `gorm:"column:emergency_contact_phone;type:varchar(20)" json:"emergencyContactPhone"`
	Remark                   *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	Status                   int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	CreatorID                *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID                *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate               *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate               *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                  int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedPatient) TableName() string { return "med_patient" }

type PatientListRequest struct {
	Page           int    `form:"page" example:"1"`
	PageSize       int    `form:"pageSize" example:"20"`
	Keyword        string `form:"keyword" example:"PAT000001"`
	Gender         string `form:"gender" example:"male"`
	Status         *int   `form:"status" example:"1"`
	CreateDateFrom string `form:"createDateFrom" example:"2026-01-01"`
	CreateDateTo   string `form:"createDateTo" example:"2026-12-31"`
	Sorts          string `form:"sorts" example:"createDate,desc"`
}

type SavePatientRequest struct {
	Name                     string  `json:"name" binding:"required,max=64"`
	Gender                   string  `json:"gender" binding:"required"`
	BirthDate                string  `json:"birthDate" binding:"required"`
	IDType                   string  `json:"idType" binding:"required"`
	IDNumber                 string  `json:"idNumber" binding:"required,max=128"`
	Phone                    string  `json:"phone" binding:"required"`
	Address                  *string `json:"address"`
	EmergencyContactName     *string `json:"emergencyContactName"`
	EmergencyContactRelation *string `json:"emergencyContactRelation"`
	EmergencyContactPhone    *string `json:"emergencyContactPhone"`
	Remark                   *string `json:"remark"`
}

type UpdatePatientStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1" example:"1"`
}

type PatientResponse struct {
	PatientID                string  `json:"patientId" example:"550e8400-e29b-41d4-a716-446655440000"`
	PatientNo                string  `json:"patientNo" example:"PAT000001"`
	Name                     string  `json:"name" example:"张*"`
	Gender                   string  `json:"gender" example:"male"`
	BirthDate                string  `json:"birthDate" example:"1990-01-01"`
	IDType                   string  `json:"idType" example:"ID_CARD"`
	IDNumber                 string  `json:"idNumber" example:"110***********1234"`
	Phone                    string  `json:"phone" example:"138****5678"`
	Address                  *string `json:"address" example:"上海市浦东新区"`
	EmergencyContactName     *string `json:"emergencyContactName" example:"李四"`
	EmergencyContactRelation *string `json:"emergencyContactRelation" example:"配偶"`
	EmergencyContactPhone    *string `json:"emergencyContactPhone" example:"139****5678"`
	Remark                   *string `json:"remark" example:"首次建档"`
	Status                   int     `json:"status" example:"1"`
	CreateDate               *string `json:"createDate" example:"2026-01-15 09:00:00"`
	UpdateDate               *string `json:"updateDate" example:"2026-01-15 09:00:00"`
}

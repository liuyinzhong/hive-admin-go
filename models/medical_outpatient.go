package models

import "time"

const (
	MedVisitQueueStatusWaiting    = 0
	MedVisitQueueStatusCalled     = 10
	MedVisitQueueStatusPassed     = 15
	MedVisitQueueStatusConsulting = 20
	MedVisitQueueStatusCompleted  = 30
)

const (
	MedVisitTypeInitial  = 0
	MedVisitTypeFollowUp = 10
)

const (
	MedPrescriptionTypeOrdinary = 10
	MedPrescriptionTypeHerbal   = 20

	MedPrescriptionStatusDraft         = 0
	MedPrescriptionStatusPendingReview = 10
	MedPrescriptionStatusApproved      = 20
	MedPrescriptionStatusRejected      = 30
	MedPrescriptionStatusVoided        = 40
	MedPrescriptionSubmissionPending   = 0
	MedPrescriptionSubmissionApproved  = 10
	MedPrescriptionSubmissionRejected  = 20
	MedPrescriptionSubmissionWithdrawn = 30
)

type MedDiagnosis struct {
	DiagnosisID string     `gorm:"column:diagnosis_id;type:char(36);primaryKey" json:"diagnosisId"`
	ICDCode     string     `gorm:"column:icd_code;type:varchar(32)" json:"icdCode"`
	ICDName     string     `gorm:"column:icd_name;type:varchar(128)" json:"icdName"`
	NamePinyin  *string    `gorm:"column:name_pinyin;type:varchar(256)" json:"namePinyin"`
	Status      int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Sort        int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Remark      *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID   *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID   *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate  *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate  *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag     int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedDiagnosis) TableName() string { return "med_diagnosis" }

type DiagnosisListRequest struct {
	Page     int    `form:"page" example:"1"`
	PageSize int    `form:"pageSize" example:"20"`
	Keyword  string `form:"keyword" example:"高血压"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1" example:"1"`
	Sorts    string `form:"sorts" example:"sort,asc;createDate,desc"`
}

type DiagnosisOptionRequest struct {
	Keyword  string `form:"keyword" example:"高血压"`
	PageSize int    `form:"pageSize" example:"50"`
}

type SaveDiagnosisRequest struct {
	ICDCode    string  `json:"icdCode" binding:"required,max=32" example:"I10"`
	ICDName    string  `json:"icdName" binding:"required,max=128" example:"原发性高血压"`
	NamePinyin *string `json:"namePinyin" binding:"omitempty,max=256" example:"yuanfaxinggaoxueya"`
	Status     int     `json:"status" binding:"oneof=0 1" example:"1"`
	Sort       int     `json:"sort" binding:"min=0,max=999999" example:"10"`
	Remark     *string `json:"remark" binding:"omitempty,max=512" example:"常见慢性病"`
}

type UpdateDiagnosisStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1" example:"1"`
}

type DiagnosisResponse struct {
	DiagnosisID string  `json:"diagnosisId"`
	ICDCode     string  `json:"icdCode"`
	ICDName     string  `json:"icdName"`
	NamePinyin  *string `json:"namePinyin"`
	Status      int     `json:"status"`
	Sort        int     `json:"sort"`
	Remark      *string `json:"remark"`
	CreateDate  *string `json:"createDate"`
	UpdateDate  *string `json:"updateDate"`
}

type MedOutpatientRecord struct {
	RecordID              string     `gorm:"column:record_id;type:char(36);primaryKey" json:"recordId"`
	RegistrationID        string     `gorm:"column:registration_id;type:char(36)" json:"registrationId"`
	PatientID             string     `gorm:"column:patient_id;type:char(36)" json:"patientId"`
	DoctorID              string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DepartmentID          string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	VisitType             *int       `gorm:"column:visit_type;type:smallint" json:"visitType"`
	Informant             *string    `gorm:"column:informant;type:varchar(64)" json:"informant"`
	Temperature           *string    `gorm:"column:temperature;type:decimal(4,1)" json:"temperature"`
	Pulse                 *int       `gorm:"column:pulse;type:smallint" json:"pulse"`
	RespiratoryRate       *int       `gorm:"column:respiratory_rate;type:smallint" json:"respiratoryRate"`
	SystolicPressure      *int       `gorm:"column:systolic_pressure;type:smallint" json:"systolicPressure"`
	DiastolicPressure     *int       `gorm:"column:diastolic_pressure;type:smallint" json:"diastolicPressure"`
	Height                *string    `gorm:"column:height;type:decimal(5,2)" json:"height"`
	Weight                *string    `gorm:"column:weight;type:decimal(5,2)" json:"weight"`
	ChiefComplaint        *string    `gorm:"column:chief_complaint;type:text" json:"chiefComplaint"`
	PresentIllness        *string    `gorm:"column:present_illness;type:text" json:"presentIllness"`
	PastHistory           *string    `gorm:"column:past_history;type:text" json:"pastHistory"`
	PersonalHistory       *string    `gorm:"column:personal_history;type:text" json:"personalHistory"`
	FamilyHistory         *string    `gorm:"column:family_history;type:text" json:"familyHistory"`
	AllergyHistory        *string    `gorm:"column:allergy_history;type:text" json:"allergyHistory"`
	MaritalReproductive   *string    `gorm:"column:marital_reproductive;type:text" json:"maritalReproductive"`
	MenstrualHistory      *string    `gorm:"column:menstrual_history;type:text" json:"menstrualHistory"`
	PhysicalExamination   *string    `gorm:"column:physical_examination;type:text" json:"physicalExamination"`
	SpecialistExamination *string    `gorm:"column:specialist_examination;type:text" json:"specialistExamination"`
	AuxiliaryExamination  *string    `gorm:"column:auxiliary_examination;type:text" json:"auxiliaryExamination"`
	TreatmentPlan         *string    `gorm:"column:treatment_plan;type:text" json:"treatmentPlan"`
	MedicalAdvice         *string    `gorm:"column:medical_advice;type:text" json:"medicalAdvice"`
	FollowUpAdvice        *string    `gorm:"column:follow_up_advice;type:text" json:"followUpAdvice"`
	Remark                *string    `gorm:"column:remark;type:varchar(1000)" json:"remark"`
	StartDate             time.Time  `gorm:"column:start_date" json:"startDate"`
	EndDate               *time.Time `gorm:"column:end_date" json:"endDate"`
	CreatorID             *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID             *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate            *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate            *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedOutpatientRecord) TableName() string { return "med_outpatient_record" }

type MedOutpatientDiagnosis struct {
	RecordDiagnosisID string    `gorm:"column:record_diagnosis_id;type:char(36);primaryKey" json:"recordDiagnosisId"`
	RecordID          string    `gorm:"column:record_id;type:char(36)" json:"recordId"`
	DiagnosisID       string    `gorm:"column:diagnosis_id;type:char(36)" json:"diagnosisId"`
	ICDCode           string    `gorm:"column:icd_code;type:varchar(32)" json:"icdCode"`
	ICDName           string    `gorm:"column:icd_name;type:varchar(128)" json:"icdName"`
	IsPrimary         int       `gorm:"column:is_primary;type:tinyint" json:"isPrimary"`
	Sort              int       `gorm:"column:sort;type:int" json:"sort"`
	CreatorID         *string   `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate        time.Time `gorm:"column:create_date" json:"createDate"`
}

func (MedOutpatientDiagnosis) TableName() string { return "med_outpatient_diagnosis" }

type SaveOutpatientDiagnosisRequest struct {
	DiagnosisID string `json:"diagnosisId" binding:"required"`
	IsPrimary   int    `json:"isPrimary" binding:"oneof=0 1"`
	Sort        int    `json:"sort" binding:"min=0"`
}

type SaveOutpatientRecordRequest struct {
	VisitType             *int                             `json:"visitType" binding:"omitempty,oneof=0 10"`
	Informant             *string                          `json:"informant" binding:"omitempty,max=64"`
	Temperature           *string                          `json:"temperature"`
	Pulse                 *int                             `json:"pulse" binding:"omitempty,min=1,max=300"`
	RespiratoryRate       *int                             `json:"respiratoryRate" binding:"omitempty,min=1,max=100"`
	SystolicPressure      *int                             `json:"systolicPressure" binding:"omitempty,min=1,max=300"`
	DiastolicPressure     *int                             `json:"diastolicPressure" binding:"omitempty,min=1,max=200"`
	Height                *string                          `json:"height"`
	Weight                *string                          `json:"weight"`
	ChiefComplaint        *string                          `json:"chiefComplaint"`
	PresentIllness        *string                          `json:"presentIllness"`
	PastHistory           *string                          `json:"pastHistory"`
	PersonalHistory       *string                          `json:"personalHistory"`
	FamilyHistory         *string                          `json:"familyHistory"`
	AllergyHistory        *string                          `json:"allergyHistory"`
	MaritalReproductive   *string                          `json:"maritalReproductive"`
	MenstrualHistory      *string                          `json:"menstrualHistory"`
	PhysicalExamination   *string                          `json:"physicalExamination"`
	SpecialistExamination *string                          `json:"specialistExamination"`
	AuxiliaryExamination  *string                          `json:"auxiliaryExamination"`
	TreatmentPlan         *string                          `json:"treatmentPlan"`
	MedicalAdvice         *string                          `json:"medicalAdvice"`
	FollowUpAdvice        *string                          `json:"followUpAdvice"`
	Remark                *string                          `json:"remark" binding:"omitempty,max=1000"`
	Diagnoses             []SaveOutpatientDiagnosisRequest `json:"diagnoses"`
}

type OutpatientDiagnosisResponse struct {
	RecordDiagnosisID string `json:"recordDiagnosisId"`
	DiagnosisID       string `json:"diagnosisId"`
	ICDCode           string `json:"icdCode"`
	ICDName           string `json:"icdName"`
	IsPrimary         int    `json:"isPrimary"`
	Sort              int    `json:"sort"`
}

type OutpatientRecordResponse struct {
	MedOutpatientRecord
	RegistrationNo   string                        `json:"registrationNo"`
	PatientNo        string                        `json:"patientNo"`
	PatientName      string                        `json:"patientName"`
	PatientGender    string                        `json:"patientGender"`
	PatientBirthDate string                        `json:"patientBirthDate"`
	PatientPhone     string                        `json:"patientPhone"`
	DoctorName       string                        `json:"doctorName"`
	DepartmentName   string                        `json:"departmentName"`
	QueueID          string                        `json:"queueId"`
	QueueSequence    int                           `json:"queueSequence"`
	QueueStatus      int                           `json:"queueStatus"`
	Diagnoses        []OutpatientDiagnosisResponse `json:"diagnoses"`
	Prescriptions    []PrescriptionResponse        `json:"prescriptions"`
}

type DoctorWorkbenchQueueResponse struct {
	QueueID        string  `json:"queueId"`
	RegistrationID string  `json:"registrationId"`
	RecordID       *string `json:"recordId"`
	QueueSequence  int     `json:"queueSequence"`
	QueueStatus    int     `json:"queueStatus"`
	CallCount      int     `json:"callCount"`
	PatientNo      string  `json:"patientNo"`
	PatientName    string  `json:"patientName"`
	PatientPhone   string  `json:"patientPhone"`
	RegistrationNo string  `json:"registrationNo"`
	StartTime      string  `json:"startTime"`
	EndTime        string  `json:"endTime"`
	CheckInTime    string  `json:"checkInTime"`
}

type DoctorWorkbenchScheduleResponse struct {
	ScheduleID       string                         `json:"scheduleId"`
	ScheduleDate     string                         `json:"scheduleDate"`
	StartTime        string                         `json:"startTime"`
	EndTime          string                         `json:"endTime"`
	DepartmentID     string                         `json:"departmentId"`
	DepartmentName   string                         `json:"departmentName"`
	RegistrationType string                         `json:"registrationType"`
	Queues           []DoctorWorkbenchQueueResponse `json:"queues"`
}

type DoctorWorkbenchResponse struct {
	DoctorID   string                            `json:"doctorId"`
	DoctorNo   string                            `json:"doctorNo"`
	DoctorName string                            `json:"doctorName"`
	Schedules  []DoctorWorkbenchScheduleResponse `json:"schedules"`
}

type MedPrescription struct {
	PrescriptionID   string     `gorm:"column:prescription_id;type:char(36);primaryKey" json:"prescriptionId"`
	PrescriptionNo   string     `gorm:"column:prescription_no;type:varchar(32)" json:"prescriptionNo"`
	RecordID         string     `gorm:"column:record_id;type:char(36)" json:"recordId"`
	RegistrationID   string     `gorm:"column:registration_id;type:char(36)" json:"registrationId"`
	PatientID        string     `gorm:"column:patient_id;type:char(36)" json:"patientId"`
	DoctorID         string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	PrescriptionType int        `gorm:"column:prescription_type;type:smallint" json:"prescriptionType"`
	Status           int        `gorm:"column:status;type:smallint" json:"status"`
	CurrentVersion   int        `gorm:"column:current_version;type:int" json:"currentVersion"`
	Remark           *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID        *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID        *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedPrescription) TableName() string { return "med_prescription" }

type MedPrescriptionItem struct {
	ItemID            string     `gorm:"column:item_id;type:char(36);primaryKey" json:"itemId"`
	PrescriptionID    string     `gorm:"column:prescription_id;type:char(36)" json:"prescriptionId"`
	SkuID             string     `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	SkuCode           string     `gorm:"column:sku_code;type:varchar(32)" json:"skuCode"`
	ProductName       string     `gorm:"column:product_name;type:varchar(128)" json:"productName"`
	SpecName          string     `gorm:"column:spec_name;type:varchar(128)" json:"specName"`
	DosageForm        *string    `gorm:"column:dosage_form;type:varchar(64)" json:"dosageForm"`
	EnterpriseName    string     `gorm:"column:enterprise_name;type:varchar(128)" json:"enterpriseName"`
	ApprovalNo        string     `gorm:"column:approval_no;type:varchar(128)" json:"approvalNo"`
	PackageSpecName   string     `gorm:"column:package_spec_name;type:varchar(128)" json:"packageSpecName"`
	PackConversion    int        `gorm:"column:pack_conversion;type:int" json:"packConversion"`
	MinUnitName       string     `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	PackageUnitName   string     `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	AllowSplit        int        `gorm:"column:allow_split;type:tinyint" json:"allowSplit"`
	SingleDose        string     `gorm:"column:single_dose;type:decimal(12,3)" json:"singleDose"`
	DoseUnit          string     `gorm:"column:dose_unit;type:varchar(32)" json:"doseUnit"`
	MedicationRoute   string     `gorm:"column:medication_route;type:varchar(36)" json:"medicationRoute"`
	Frequency         string     `gorm:"column:frequency;type:varchar(16)" json:"frequency"`
	CourseDays        int        `gorm:"column:course_days;type:int" json:"courseDays"`
	TotalMinQuantity  string     `gorm:"column:total_min_quantity;type:decimal(14,3)" json:"totalMinQuantity"`
	DispenseQuantity  string     `gorm:"column:dispense_quantity;type:decimal(14,3)" json:"dispenseQuantity"`
	DispenseUnit      string     `gorm:"column:dispense_unit;type:varchar(32)" json:"dispenseUnit"`
	UsageInstructions *string    `gorm:"column:usage_instructions;type:varchar(1000)" json:"usageInstructions"`
	Remark            *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	Sort              int        `gorm:"column:sort;type:int" json:"sort"`
	CreatorID         *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID         *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (MedPrescriptionItem) TableName() string { return "med_prescription_item" }

type MedPrescriptionSubmission struct {
	SubmissionID     string     `gorm:"column:submission_id;type:char(36);primaryKey" json:"submissionId"`
	PrescriptionID   string     `gorm:"column:prescription_id;type:char(36)" json:"prescriptionId"`
	Version          int        `gorm:"column:version;type:int" json:"version"`
	SubmissionStatus int        `gorm:"column:submission_status;type:smallint" json:"submissionStatus"`
	AllergyHistory   *string    `gorm:"column:allergy_history;type:text" json:"allergyHistory"`
	SubmittedBy      string     `gorm:"column:submitted_by;type:char(36)" json:"submittedBy"`
	SubmittedAt      time.Time  `gorm:"column:submitted_at" json:"submittedAt"`
	ReviewerID       *string    `gorm:"column:reviewer_id;type:char(36)" json:"reviewerId"`
	ReviewOpinion    *string    `gorm:"column:review_opinion;type:varchar(1000)" json:"reviewOpinion"`
	ReviewedAt       *time.Time `gorm:"column:reviewed_at" json:"reviewedAt"`
}

func (MedPrescriptionSubmission) TableName() string { return "med_prescription_submission" }

type MedPrescriptionSubmissionItem struct {
	SubmissionItemID  string  `gorm:"column:submission_item_id;type:char(36);primaryKey" json:"submissionItemId"`
	SubmissionID      string  `gorm:"column:submission_id;type:char(36)" json:"submissionId"`
	SkuID             string  `gorm:"column:sku_id;type:char(36)" json:"skuId"`
	SkuCode           string  `gorm:"column:sku_code;type:varchar(32)" json:"skuCode"`
	ProductName       string  `gorm:"column:product_name;type:varchar(128)" json:"productName"`
	SpecName          string  `gorm:"column:spec_name;type:varchar(128)" json:"specName"`
	DosageForm        *string `gorm:"column:dosage_form;type:varchar(64)" json:"dosageForm"`
	EnterpriseName    string  `gorm:"column:enterprise_name;type:varchar(128)" json:"enterpriseName"`
	ApprovalNo        string  `gorm:"column:approval_no;type:varchar(128)" json:"approvalNo"`
	PackageSpecName   string  `gorm:"column:package_spec_name;type:varchar(128)" json:"packageSpecName"`
	PackConversion    int     `gorm:"column:pack_conversion;type:int" json:"packConversion"`
	MinUnitName       string  `gorm:"column:min_unit_name;type:varchar(32)" json:"minUnitName"`
	PackageUnitName   string  `gorm:"column:package_unit_name;type:varchar(32)" json:"packageUnitName"`
	AllowSplit        int     `gorm:"column:allow_split;type:tinyint" json:"allowSplit"`
	SingleDose        string  `gorm:"column:single_dose;type:decimal(12,3)" json:"singleDose"`
	DoseUnit          string  `gorm:"column:dose_unit;type:varchar(32)" json:"doseUnit"`
	MedicationRoute   string  `gorm:"column:medication_route;type:varchar(36)" json:"medicationRoute"`
	Frequency         string  `gorm:"column:frequency;type:varchar(16)" json:"frequency"`
	CourseDays        int     `gorm:"column:course_days;type:int" json:"courseDays"`
	TotalMinQuantity  string  `gorm:"column:total_min_quantity;type:decimal(14,3)" json:"totalMinQuantity"`
	DispenseQuantity  string  `gorm:"column:dispense_quantity;type:decimal(14,3)" json:"dispenseQuantity"`
	DispenseUnit      string  `gorm:"column:dispense_unit;type:varchar(32)" json:"dispenseUnit"`
	UsageInstructions *string `gorm:"column:usage_instructions;type:varchar(1000)" json:"usageInstructions"`
	Remark            *string `gorm:"column:remark;type:varchar(512)" json:"remark"`
	Sort              int     `gorm:"column:sort;type:int" json:"sort"`
}

func (MedPrescriptionSubmissionItem) TableName() string { return "med_prescription_submission_item" }

type MedPrescriptionSubmissionDiagnosis struct {
	SubmissionDiagnosisID string `gorm:"column:submission_diagnosis_id;type:char(36);primaryKey" json:"submissionDiagnosisId"`
	SubmissionID          string `gorm:"column:submission_id;type:char(36)" json:"submissionId"`
	DiagnosisID           string `gorm:"column:diagnosis_id;type:char(36)" json:"diagnosisId"`
	ICDCode               string `gorm:"column:icd_code;type:varchar(32)" json:"icdCode"`
	ICDName               string `gorm:"column:icd_name;type:varchar(128)" json:"icdName"`
	IsPrimary             int    `gorm:"column:is_primary;type:tinyint" json:"isPrimary"`
	Sort                  int    `gorm:"column:sort;type:int" json:"sort"`
}

func (MedPrescriptionSubmissionDiagnosis) TableName() string {
	return "med_prescription_submission_diagnosis"
}

type SavePrescriptionItemRequest struct {
	SkuID             string  `json:"skuId" binding:"required"`
	SingleDose        string  `json:"singleDose" binding:"required"`
	MedicationRoute   string  `json:"medicationRoute" binding:"required,max=36"`
	Frequency         string  `json:"frequency" binding:"required,max=16"`
	CourseDays        int     `json:"courseDays" binding:"required,min=1,max=365"`
	TotalMinQuantity  *string `json:"totalMinQuantity"`
	UsageInstructions *string `json:"usageInstructions" binding:"omitempty,max=1000"`
	Remark            *string `json:"remark" binding:"omitempty,max=512"`
	Sort              int     `json:"sort" binding:"min=0"`
}

type SavePrescriptionRequest struct {
	PrescriptionType int                           `json:"prescriptionType" binding:"required,oneof=10"`
	Remark           *string                       `json:"remark" binding:"omitempty,max=512"`
	Items            []SavePrescriptionItemRequest `json:"items"`
}

type ReviewPrescriptionRequest struct {
	Approved int     `json:"approved" binding:"oneof=0 1"`
	Opinion  *string `json:"opinion" binding:"omitempty,max=1000"`
}

type PrescriptionListRequest struct {
	Page           int    `form:"page" example:"1"`
	PageSize       int    `form:"pageSize" example:"20"`
	PrescriptionNo string `form:"prescriptionNo"`
	PatientKeyword string `form:"patientKeyword"`
	Status         *int   `form:"status"`
	Sorts          string `form:"sorts"`
}

type PrescriptionResponse struct {
	MedPrescription
	PatientNo           string                               `json:"patientNo"`
	PatientName         string                               `json:"patientName"`
	DoctorName          string                               `json:"doctorName"`
	DepartmentName      string                               `json:"departmentName"`
	RegistrationNo      string                               `json:"registrationNo"`
	Items               []MedPrescriptionItem                `json:"items"`
	LatestSubmission    *MedPrescriptionSubmission           `json:"latestSubmission"`
	SubmissionItems     []MedPrescriptionSubmissionItem      `json:"submissionItems"`
	SubmissionDiagnoses []MedPrescriptionSubmissionDiagnosis `json:"submissionDiagnoses"`
}

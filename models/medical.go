package models

import "time"

type MedDepartment struct {
	DepartmentID   string     `gorm:"column:department_id;type:char(36);primaryKey" json:"departmentId"`
	DepartmentCode string     `gorm:"column:department_code;type:varchar(32)" json:"departmentCode"`
	DepartmentName string     `gorm:"column:department_name;type:varchar(64)" json:"departmentName"`
	Pid            *string    `gorm:"column:pid;type:char(36)" json:"pid"`
	Sort           int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Status         int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark         *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID      *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag        int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedDepartment) TableName() string { return "med_department" }

type MedDoctor struct {
	DoctorID               string     `gorm:"column:doctor_id;type:char(36);primaryKey" json:"doctorId"`
	DoctorNo               string     `gorm:"column:doctor_no;type:varchar(32)" json:"doctorNo"`
	UserID                 *string    `gorm:"column:user_id;type:char(36)" json:"userId"`
	Name                   string     `gorm:"column:name;type:varchar(64)" json:"name"`
	NamePinyin             *string    `gorm:"column:name_pinyin;type:varchar(128)" json:"namePinyin"`
	Gender                 *string    `gorm:"column:gender;type:varchar(36)" json:"gender"`
	BirthDate              *time.Time `gorm:"column:birth_date;type:date" json:"birthDate"`
	Phone                  *string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email                  *string    `gorm:"column:email;type:varchar(128)" json:"email"`
	Avatar                 *string    `gorm:"column:avatar;type:varchar(512)" json:"avatar"`
	ProfessionalTitle      string     `gorm:"column:professional_title;type:varchar(36)" json:"professionalTitle"`
	AdministrativePosition *string    `gorm:"column:administrative_position;type:varchar(64)" json:"administrativePosition"`
	EmploymentType         string     `gorm:"column:employment_type;type:varchar(36)" json:"employmentType"`
	PracticeStartDate      *time.Time `gorm:"column:practice_start_date;type:date" json:"practiceStartDate"`
	EmploymentDate         *time.Time `gorm:"column:employment_date;type:date" json:"employmentDate"`
	DepartureDate          *time.Time `gorm:"column:departure_date;type:date" json:"departureDate"`
	Expertise              *string    `gorm:"column:expertise;type:text" json:"expertise"`
	Introduction           *string    `gorm:"column:introduction;type:text" json:"introduction"`
	DefaultVisitMinutes    int        `gorm:"column:default_visit_minutes;type:smallint;default:15" json:"defaultVisitMinutes"`
	OnlineConsultation     int        `gorm:"column:online_consultation;type:tinyint;default:0" json:"onlineConsultation"`
	AppointmentEnabled     int        `gorm:"column:appointment_enabled;type:tinyint;default:1" json:"appointmentEnabled"`
	ProfileVisible         int        `gorm:"column:profile_visible;type:tinyint;default:1" json:"profileVisible"`
	Sort                   int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Status                 int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark                 *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID              *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID              *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate             *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate             *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedDoctor) TableName() string { return "med_doctor" }

type MedDoctorDepartment struct {
	DoctorDepartmentID string     `gorm:"column:doctor_department_id;type:char(36);primaryKey" json:"doctorDepartmentId"`
	DoctorID           string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DepartmentID       string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	IsPrimary          int        `gorm:"column:is_primary;type:tinyint;default:0" json:"isPrimary"`
	DepartmentPosition *string    `gorm:"column:department_position;type:varchar(64)" json:"departmentPosition"`
	AppointmentEnabled int        `gorm:"column:appointment_enabled;type:tinyint;default:1" json:"appointmentEnabled"`
	ValidFrom          *time.Time `gorm:"column:valid_from;type:date" json:"validFrom"`
	ValidTo            *time.Time `gorm:"column:valid_to;type:date" json:"validTo"`
	Sort               int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Status             int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	CreatorID          *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID          *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate         *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate         *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag            int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedDoctorDepartment) TableName() string { return "med_doctor_department" }

type MedicalDepartmentListRequest struct {
	Keyword string `form:"keyword" example:"内科"`
	Status  *int   `form:"status" example:"1"`
}

type MedicalDepartmentTreeResponse struct {
	DepartmentID   string                           `json:"departmentId"`
	DepartmentCode string                           `json:"departmentCode"`
	DepartmentName string                           `json:"departmentName"`
	Pid            *string                          `json:"pid"`
	Sort           int                              `json:"sort"`
	Status         int                              `json:"status"`
	Remark         *string                          `json:"remark"`
	CreateDate     *string                          `json:"createDate"`
	UpdateDate     *string                          `json:"updateDate"`
	Children       []*MedicalDepartmentTreeResponse `json:"children"`
}

type CreateMedicalDepartmentRequest struct {
	DepartmentCode string  `json:"departmentCode" binding:"required,max=32"`
	DepartmentName string  `json:"departmentName" binding:"required,max=64"`
	Pid            *string `json:"pid"`
	Sort           int     `json:"sort"`
	Status         int     `json:"status"`
	Remark         *string `json:"remark"`
}

type UpdateMedicalDepartmentRequest = CreateMedicalDepartmentRequest

type UpdateMedicalStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

type DoctorListRequest struct {
	Page              int    `form:"page" example:"1"`
	PageSize          int    `form:"pageSize" example:"20"`
	Keyword           string `form:"keyword"`
	DepartmentID      string `form:"departmentId"`
	ProfessionalTitle string `form:"professionalTitle"`
	EmploymentType    string `form:"employmentType"`
	Status            *int   `form:"status"`
	Sorts             string `form:"sorts"`
}

type SaveDoctorRequest struct {
	DoctorNo               string   `json:"doctorNo" binding:"required,max=32"`
	UserID                 *string  `json:"userId"`
	Name                   string   `json:"name" binding:"required,max=64"`
	NamePinyin             *string  `json:"namePinyin"`
	Gender                 *string  `json:"gender"`
	BirthDate              *string  `json:"birthDate"`
	Phone                  *string  `json:"phone"`
	Email                  *string  `json:"email"`
	Avatar                 *string  `json:"avatar"`
	ProfessionalTitle      string   `json:"professionalTitle" binding:"required"`
	AdministrativePosition *string  `json:"administrativePosition"`
	EmploymentType         string   `json:"employmentType" binding:"required"`
	PracticeStartDate      *string  `json:"practiceStartDate"`
	EmploymentDate         *string  `json:"employmentDate"`
	DepartureDate          *string  `json:"departureDate"`
	Expertise              *string  `json:"expertise"`
	Introduction           *string  `json:"introduction"`
	DefaultVisitMinutes    int      `json:"defaultVisitMinutes"`
	OnlineConsultation     int      `json:"onlineConsultation"`
	AppointmentEnabled     int      `json:"appointmentEnabled"`
	ProfileVisible         int      `json:"profileVisible"`
	Sort                   int      `json:"sort"`
	Status                 int      `json:"status"`
	Remark                 *string  `json:"remark"`
	DepartmentIDs          []string `json:"departmentIds" binding:"required,min=1"`
	PrimaryDepartmentID    string   `json:"primaryDepartmentId" binding:"required"`
}

type DoctorDepartmentResponse struct {
	DoctorDepartmentID string  `json:"doctorDepartmentId"`
	DepartmentID       string  `json:"departmentId"`
	DepartmentCode     string  `json:"departmentCode"`
	DepartmentName     string  `json:"departmentName"`
	IsPrimary          int     `json:"isPrimary"`
	DepartmentPosition *string `json:"departmentPosition"`
	AppointmentEnabled int     `json:"appointmentEnabled"`
	Status             int     `json:"status"`
}

type DoctorResponse struct {
	DoctorID               string                     `json:"doctorId"`
	DoctorNo               string                     `json:"doctorNo"`
	UserID                 *string                    `json:"userId"`
	UserName               *string                    `json:"userName"`
	Name                   string                     `json:"name"`
	NamePinyin             *string                    `json:"namePinyin"`
	Gender                 *string                    `json:"gender"`
	BirthDate              *string                    `json:"birthDate"`
	Phone                  *string                    `json:"phone"`
	Email                  *string                    `json:"email"`
	Avatar                 *string                    `json:"avatar"`
	ProfessionalTitle      string                     `json:"professionalTitle"`
	AdministrativePosition *string                    `json:"administrativePosition"`
	EmploymentType         string                     `json:"employmentType"`
	PracticeStartDate      *string                    `json:"practiceStartDate"`
	EmploymentDate         *string                    `json:"employmentDate"`
	DepartureDate          *string                    `json:"departureDate"`
	Expertise              *string                    `json:"expertise"`
	Introduction           *string                    `json:"introduction"`
	DefaultVisitMinutes    int                        `json:"defaultVisitMinutes"`
	OnlineConsultation     int                        `json:"onlineConsultation"`
	AppointmentEnabled     int                        `json:"appointmentEnabled"`
	ProfileVisible         int                        `json:"profileVisible"`
	Sort                   int                        `json:"sort"`
	Status                 int                        `json:"status"`
	Remark                 *string                    `json:"remark"`
	DepartmentIDs          []string                   `json:"departmentIds"`
	DepartmentNames        []string                   `json:"departmentNames"`
	PrimaryDepartmentID    *string                    `json:"primaryDepartmentId"`
	PrimaryDepartmentName  *string                    `json:"primaryDepartmentName"`
	Departments            []DoctorDepartmentResponse `json:"departments"`
	CreateDate             *string                    `json:"createDate"`
	UpdateDate             *string                    `json:"updateDate"`
}

type DoctorOptionResponse struct {
	DoctorID              string  `json:"doctorId"`
	DoctorNo              string  `json:"doctorNo"`
	Name                  string  `json:"name"`
	ProfessionalTitle     string  `json:"professionalTitle"`
	PrimaryDepartmentID   *string `json:"primaryDepartmentId"`
	PrimaryDepartmentName *string `json:"primaryDepartmentName"`
}

type MedRegistrationFeeRule struct {
	FeeRuleID        string     `gorm:"column:fee_rule_id;type:char(36);primaryKey" json:"feeRuleId"`
	DoctorID         string     `gorm:"column:doctor_id;type:char(36)" json:"doctorId"`
	DepartmentID     string     `gorm:"column:department_id;type:char(36)" json:"departmentId"`
	RegistrationType string     `gorm:"column:registration_type;type:varchar(36)" json:"registrationType"`
	FeeAmount        string     `gorm:"column:fee_amount;type:decimal(10,2)" json:"feeAmount"`
	EffectiveDate    time.Time  `gorm:"column:effective_date;type:date" json:"effectiveDate"`
	ExpiryDate       *time.Time `gorm:"column:expiry_date;type:date" json:"expiryDate"`
	Version          int        `gorm:"column:version;type:int" json:"version"`
	Remark           *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	CreatorID        *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID        *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag          int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (MedRegistrationFeeRule) TableName() string { return "med_registration_fee_rule" }

type RegistrationFeeRuleListRequest struct {
	Page             int    `form:"page" example:"1"`
	PageSize         int    `form:"pageSize" example:"20"`
	Keyword          string `form:"keyword"`
	DoctorID         string `form:"doctorId"`
	DepartmentID     string `form:"departmentId"`
	RegistrationType string `form:"registrationType"`
	PeriodStatus     string `form:"periodStatus"`
	Sorts            string `form:"sorts"`
}

type CreateRegistrationFeeRuleRequest struct {
	DoctorID         string  `json:"doctorId" binding:"required"`
	DepartmentID     string  `json:"departmentId" binding:"required"`
	RegistrationType string  `json:"registrationType" binding:"required"`
	FeeAmount        string  `json:"feeAmount" binding:"required"`
	EffectiveDate    string  `json:"effectiveDate" binding:"required"`
	ExpiryDate       *string `json:"expiryDate"`
	Remark           *string `json:"remark" binding:"omitempty,max=512"`
}

type AdjustRegistrationFeeRuleRequest struct {
	FeeAmount     string  `json:"feeAmount" binding:"required"`
	EffectiveDate string  `json:"effectiveDate" binding:"required"`
	Remark        *string `json:"remark" binding:"omitempty,max=512"`
}

type RegistrationFeeRuleResponse struct {
	FeeRuleID        string  `json:"feeRuleId"`
	DoctorID         string  `json:"doctorId"`
	DoctorNo         string  `json:"doctorNo"`
	DoctorName       string  `json:"doctorName"`
	DepartmentID     string  `json:"departmentId"`
	DepartmentCode   string  `json:"departmentCode"`
	DepartmentName   string  `json:"departmentName"`
	RegistrationType string  `json:"registrationType"`
	FeeAmount        string  `json:"feeAmount"`
	EffectiveDate    string  `json:"effectiveDate"`
	ExpiryDate       *string `json:"expiryDate"`
	Version          int     `json:"version"`
	PeriodStatus     string  `json:"periodStatus"`
	Remark           *string `json:"remark"`
	CreateDate       *string `json:"createDate"`
	UpdateDate       *string `json:"updateDate"`
}

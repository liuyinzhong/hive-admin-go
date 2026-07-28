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
	DepartmentID   string                           `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 科室ID
	DepartmentCode string                           `json:"departmentCode" example:"DEPT001"`                            // 科室编码
	DepartmentName string                           `json:"departmentName" example:"内科"`                                 // 科室名称
	Pid            *string                          `json:"pid" example:"550e8400-e29b-41d4-a716-446655440000"`          // 父级科室ID
	Sort           int                              `json:"sort" example:"1"`                                            // 排序号
	Status         int                              `json:"status" example:"1"`                                          // 状态(0-禁用 1-启用)
	Remark         *string                          `json:"remark" example:"备注信息"`                                       // 备注
	CreateDate     *string                          `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate     *string                          `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
	Children       []*MedicalDepartmentTreeResponse `json:"children"`                                                    // 子级科室列表
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
	DoctorDepartmentID string  `json:"doctorDepartmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 医生科室关联ID
	DepartmentID       string  `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"`       // 科室ID
	DepartmentCode     string  `json:"departmentCode" example:"NK"`                                       // 科室编码
	DepartmentName     string  `json:"departmentName" example:"内科"`                                       // 科室名称
	IsPrimary          int     `json:"isPrimary" example:"1"`                                             // 是否主科室
	DepartmentPosition *string `json:"departmentPosition" example:"主任"`                                   // 科室职务
	AppointmentEnabled int     `json:"appointmentEnabled" example:"1"`                                    // 是否开启预约
	Status             int     `json:"status" example:"1"`                                                // 状态
}

type DoctorResponse struct {
	DoctorID               string                     `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"`            // 医生ID
	DoctorNo               string                     `json:"doctorNo" example:"DOC001"`                                          // 医生编号
	UserID                 *string                    `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`              // 用户ID
	UserName               *string                    `json:"userName" example:"张三"`                                              // 用户名
	Name                   string                     `json:"name" example:"张三"`                                                  // 姓名
	NamePinyin             *string                    `json:"namePinyin" example:"zhangsan"`                                      // 姓名拼音
	Gender                 *string                    `json:"gender" example:"male"`                                              // 性别
	BirthDate              *string                    `json:"birthDate" example:"1985-06-15"`                                     // 出生日期
	Phone                  *string                    `json:"phone" example:"13800138000"`                                        // 手机号
	Email                  *string                    `json:"email" example:"zhangsan@example.com"`                               // 邮箱
	Avatar                 *string                    `json:"avatar" example:"https://example.com/avatar.jpg"`                    // 头像
	ProfessionalTitle      string                     `json:"professionalTitle" example:"主任医师"`                                   // 职称
	AdministrativePosition *string                    `json:"administrativePosition" example:"科室主任"`                              // 行政职务
	EmploymentType         string                     `json:"employmentType" example:"full-time"`                                 // 执业类型
	PracticeStartDate      *string                    `json:"practiceStartDate" example:"2005-07-01"`                             // 执业开始日期
	EmploymentDate         *string                    `json:"employmentDate" example:"2010-03-15"`                                // 入职日期
	DepartureDate          *string                    `json:"departureDate" example:"2024-12-31"`                                 // 离职日期
	Expertise              *string                    `json:"expertise" example:"心内科常见疾病诊治"`                                      // 专长
	Introduction           *string                    `json:"introduction" example:"从事心内科临床工作二十年，具有丰富的临床经验"`                      // 简介
	DefaultVisitMinutes    int                        `json:"defaultVisitMinutes" example:"15"`                                   // 默认就诊时长(分钟)
	OnlineConsultation     int                        `json:"onlineConsultation" example:"1"`                                     // 在线咨询(0-关闭 1-开启)
	AppointmentEnabled     int                        `json:"appointmentEnabled" example:"1"`                                     // 预约开启(0-关闭 1-开启)
	ProfileVisible         int                        `json:"profileVisible" example:"1"`                                         // 资料可见(0-隐藏 1-可见)
	Sort                   int                        `json:"sort" example:"0"`                                                   // 排序
	Status                 int                        `json:"status" example:"1"`                                                 // 状态(0-禁用 1-启用)
	Remark                 *string                    `json:"remark" example:"备注信息"`                                              // 备注
	DepartmentIDs          []string                   `json:"departmentIds" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"` // 科室ID列表
	DepartmentNames        []string                   `json:"departmentNames" example:"[\"内科\",\"外科\"]"`                          // 科室名称列表
	PrimaryDepartmentID    *string                    `json:"primaryDepartmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 主科室ID
	PrimaryDepartmentName  *string                    `json:"primaryDepartmentName" example:"内科"`                                 // 主科室名称
	Departments            []DoctorDepartmentResponse `json:"departments"`                                                        // 科室列表
	CreateDate             *string                    `json:"createDate" example:"2026-01-15 09:00:00"`                           // 创建时间
	UpdateDate             *string                    `json:"updateDate" example:"2026-01-15 09:00:00"`                           // 更新时间
}

type DoctorOptionResponse struct {
	DoctorID              string  `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"`            // 医生ID
	DoctorNo              string  `json:"doctorNo" example:"DOC001"`                                          // 医生编号
	Name                  string  `json:"name" example:"张三"`                                                  // 姓名
	ProfessionalTitle     string  `json:"professionalTitle" example:"主任医师"`                                   // 职称
	PrimaryDepartmentID   *string `json:"primaryDepartmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 主科室ID
	PrimaryDepartmentName *string `json:"primaryDepartmentName" example:"内科"`                                 // 主科室名称
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
	FeeRuleID        string  `json:"feeRuleId" example:"550e8400-e29b-41d4-a716-446655440000"`    // 费用规则ID
	DoctorID         string  `json:"doctorId" example:"550e8400-e29b-41d4-a716-446655440000"`     // 医生ID
	DoctorNo         string  `json:"doctorNo" example:"DOC001"`                                   // 医生工号
	DoctorName       string  `json:"doctorName" example:"张医生"`                                    // 医生姓名
	DepartmentID     string  `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"` // 科室ID
	DepartmentCode   string  `json:"departmentCode" example:"DEPT001"`                            // 科室编码
	DepartmentName   string  `json:"departmentName" example:"内科"`                                 // 科室名称
	RegistrationType string  `json:"registrationType" example:"普通"`                               // 挂号类型
	FeeAmount        string  `json:"feeAmount" example:"100.00"`                                  // 费用金额
	EffectiveDate    string  `json:"effectiveDate" example:"2026-01-15"`                          // 生效日期
	ExpiryDate       *string `json:"expiryDate" example:"2026-12-31"`                             // 失效日期
	Version          int     `json:"version" example:"1"`                                         // 版本号
	PeriodStatus     string  `json:"periodStatus" example:"active"`                               // 期间状态
	Remark           *string `json:"remark" example:"备注信息"`                                       // 备注
	CreateDate       *string `json:"createDate" example:"2026-01-15 09:00:00"`                    // 创建时间
	UpdateDate       *string `json:"updateDate" example:"2026-01-15 09:00:00"`                    // 更新时间
}

package models

import "time"

const (
	InstitutionTypeHospital = "HOSPITAL"

	InstitutionNaturePublic    = "PUBLIC"
	InstitutionNaturePrivate   = "PRIVATE"
	InstitutionNatureNonProfit = "NON_PROFIT"
	InstitutionNatureForProfit = "FOR_PROFIT"

	InstitutionCategoryGeneral        = "GENERAL"
	InstitutionCategorySpecialty      = "SPECIALTY"
	InstitutionCategoryTraditional    = "TRADITIONAL"
	InstitutionCategoryMaternalChild  = "MATERNAL_CHILD"
	InstitutionCategoryIntegrated     = "INTEGRATED"
	InstitutionCategoryRehabilitation = "REHABILITATION"
	InstitutionCategoryOther          = "OTHER"

	InstitutionLevelTertiaryA  = "TERTIARY_A"
	InstitutionLevelTertiaryB  = "TERTIARY_B"
	InstitutionLevelTertiaryC  = "TERTIARY_C"
	InstitutionLevelSecondaryA = "SECONDARY_A"
	InstitutionLevelSecondaryB = "SECONDARY_B"
	InstitutionLevelSecondaryC = "SECONDARY_C"
	InstitutionLevelPrimary    = "PRIMARY"
	InstitutionLevelUnrated    = "UNRATED"

	InstitutionContactLegalRepresentative = "LEGAL_REPRESENTATIVE"
	InstitutionContactPrincipal           = "PRINCIPAL"
	InstitutionContactMedicalQuality      = "MEDICAL_QUALITY"
	InstitutionContactInformation         = "INFORMATION"
	InstitutionContactFinance             = "FINANCE"
	InstitutionContactGeneral             = "GENERAL"

	InstitutionAddressRegistered = "REGISTERED"
	InstitutionAddressPractice   = "PRACTICE"
	InstitutionAddressMailing    = "MAILING"
)

type BaseInstitution struct {
	InstitutionID        string     `gorm:"column:institution_id;type:char(36);primaryKey" json:"institutionId"`
	SingletonKey         string     `gorm:"column:singleton_key;type:char(1);uniqueIndex" json:"-"`
	InstitutionName      string     `gorm:"column:institution_name;type:varchar(128)" json:"institutionName"`
	ShortName            *string    `gorm:"column:short_name;type:varchar(64)" json:"shortName"`
	EnglishName          *string    `gorm:"column:english_name;type:varchar(128)" json:"englishName"`
	Aliases              *string    `gorm:"column:aliases;type:varchar(256)" json:"aliases"`
	InstitutionType      string     `gorm:"column:institution_type;type:varchar(32)" json:"institutionType"`
	InstitutionNature    string     `gorm:"column:institution_nature;type:varchar(32)" json:"institutionNature"`
	HospitalCategory     string     `gorm:"column:hospital_category;type:varchar(32)" json:"hospitalCategory"`
	HospitalLevel        string     `gorm:"column:hospital_level;type:varchar(32)" json:"hospitalLevel"`
	UnifiedCreditCode    string     `gorm:"column:unified_credit_code;type:varchar(32)" json:"unifiedCreditCode"`
	EstablishmentDate    *time.Time `gorm:"column:establishment_date;type:date" json:"establishmentDate"`
	Remark               *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	LogoURL              *string    `gorm:"column:logo_url;type:varchar(512)" json:"logoUrl"`
	DisplayName          *string    `gorm:"column:display_name;type:varchar(128)" json:"displayName"`
	Slogan               *string    `gorm:"column:slogan;type:varchar(256)" json:"slogan"`
	Introduction         *string    `gorm:"column:introduction;type:text" json:"introduction"`
	DiagnosisSubjects    *string    `gorm:"column:diagnosis_subjects;type:text" json:"diagnosisSubjects"`
	KeySpecialties       *string    `gorm:"column:key_specialties;type:text" json:"keySpecialties"`
	ServiceHours         *string    `gorm:"column:service_hours;type:text" json:"serviceHours"`
	EmergencyDescription *string    `gorm:"column:emergency_description;type:text" json:"emergencyDescription"`
	ServiceFeatures      *string    `gorm:"column:service_features;type:text" json:"serviceFeatures"`
	InvoiceTitle         *string    `gorm:"column:invoice_title;type:varchar(128)" json:"invoiceTitle"`
	TaxpayerID           *string    `gorm:"column:taxpayer_id;type:varchar(32)" json:"taxpayerId"`
	TaxpayerType         *string    `gorm:"column:taxpayer_type;type:varchar(32)" json:"taxpayerType"`
	CreatorID            *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID            *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate           *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate           *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (BaseInstitution) TableName() string { return "base_institution" }

type BaseInstitutionQualification struct {
	QualificationID  string     `gorm:"column:qualification_id;type:char(36);primaryKey" json:"qualificationId"`
	InstitutionID    string     `gorm:"column:institution_id;type:char(36)" json:"institutionId"`
	CertificateName  string     `gorm:"column:certificate_name;type:varchar(128)" json:"certificateName"`
	CertificateNo    string     `gorm:"column:certificate_no;type:varchar(128)" json:"certificateNo"`
	IssuingAuthority *string    `gorm:"column:issuing_authority;type:varchar(128)" json:"issuingAuthority"`
	IssueDate        *time.Time `gorm:"column:issue_date;type:date" json:"issueDate"`
	ExpiryDate       *time.Time `gorm:"column:expiry_date;type:date" json:"expiryDate"`
	Scope            *string    `gorm:"column:scope;type:varchar(512)" json:"scope"`
	Remark           *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	Attachment       *string    `gorm:"column:attachment;type:varchar(512)" json:"attachment"`
}

func (BaseInstitutionQualification) TableName() string { return "base_institution_qualification" }

type BaseInstitutionContact struct {
	ContactID     string  `gorm:"column:contact_id;type:char(36);primaryKey" json:"contactId"`
	InstitutionID string  `gorm:"column:institution_id;type:char(36)" json:"institutionId"`
	ContactType   string  `gorm:"column:contact_type;type:varchar(32)" json:"contactType"`
	ContactName   string  `gorm:"column:contact_name;type:varchar(64)" json:"contactName"`
	JobTitle      *string `gorm:"column:job_title;type:varchar(64)" json:"jobTitle"`
	Phone         *string `gorm:"column:phone;type:varchar(32)" json:"phone"`
	Email         *string `gorm:"column:email;type:varchar(128)" json:"email"`
	IsPrimary     int     `gorm:"column:is_primary;type:tinyint;default:0" json:"isPrimary"`
	Remark        *string `gorm:"column:remark;type:varchar(512)" json:"remark"`
}

func (BaseInstitutionContact) TableName() string { return "base_institution_contact" }

type BaseInstitutionAddress struct {
	AddressID     string  `gorm:"column:address_id;type:char(36);primaryKey" json:"addressId"`
	InstitutionID string  `gorm:"column:institution_id;type:char(36)" json:"institutionId"`
	AddressType   string  `gorm:"column:address_type;type:varchar(32)" json:"addressType"`
	FullAddress   string  `gorm:"column:full_address;type:varchar(512)" json:"fullAddress"`
	PostalCode    *string `gorm:"column:postal_code;type:varchar(16)" json:"postalCode"`
	Phone         *string `gorm:"column:phone;type:varchar(32)" json:"phone"`
	IsPrimary     int     `gorm:"column:is_primary;type:tinyint;default:0" json:"isPrimary"`
	Remark        *string `gorm:"column:remark;type:varchar(512)" json:"remark"`
}

func (BaseInstitutionAddress) TableName() string { return "base_institution_address" }

type BaseInstitutionBankAccount struct {
	BankAccountID string  `gorm:"column:bank_account_id;type:char(36);primaryKey" json:"bankAccountId"`
	InstitutionID string  `gorm:"column:institution_id;type:char(36)" json:"institutionId"`
	AccountName   string  `gorm:"column:account_name;type:varchar(128)" json:"accountName"`
	BankName      string  `gorm:"column:bank_name;type:varchar(128)" json:"bankName"`
	AccountNumber string  `gorm:"column:account_number;type:varchar(64)" json:"accountNumber"`
	AccountType   *string `gorm:"column:account_type;type:varchar(32)" json:"accountType"`
	IsDefault     int     `gorm:"column:is_default;type:tinyint;default:0" json:"isDefault"`
	Remark        *string `gorm:"column:remark;type:varchar(512)" json:"remark"`
}

func (BaseInstitutionBankAccount) TableName() string { return "base_institution_bank_account" }

type SaveInstitutionRequest struct {
	InstitutionName      string                                `json:"institutionName" binding:"required,max=128" example:"示例医院"`
	ShortName            *string                               `json:"shortName" binding:"omitempty,max=64"`
	EnglishName          *string                               `json:"englishName" binding:"omitempty,max=128"`
	Aliases              *string                               `json:"aliases" binding:"omitempty,max=256"`
	InstitutionType      string                                `json:"institutionType" binding:"required,max=32" example:"HOSPITAL"`
	InstitutionNature    string                                `json:"institutionNature" binding:"required,max=32" example:"PUBLIC"`
	HospitalCategory     string                                `json:"hospitalCategory" binding:"required,max=32" example:"GENERAL"`
	HospitalLevel        string                                `json:"hospitalLevel" binding:"required,max=32" example:"TERTIARY_A"`
	UnifiedCreditCode    string                                `json:"unifiedCreditCode" binding:"required,max=32" example:"91110108MA01XXXXXX"`
	EstablishmentDate    string                                `json:"establishmentDate" binding:"omitempty" example:"2026-01-01"`
	Remark               *string                               `json:"remark" binding:"omitempty,max=512"`
	LogoURL              *string                               `json:"logoUrl" binding:"omitempty,max=512"`
	DisplayName          *string                               `json:"displayName" binding:"omitempty,max=128"`
	Slogan               *string                               `json:"slogan" binding:"omitempty,max=256"`
	Introduction         *string                               `json:"introduction" binding:"omitempty,max=5000"`
	DiagnosisSubjects    *string                               `json:"diagnosisSubjects" binding:"omitempty,max=5000"`
	KeySpecialties       *string                               `json:"keySpecialties" binding:"omitempty,max=5000"`
	ServiceHours         *string                               `json:"serviceHours" binding:"omitempty,max=2000"`
	EmergencyDescription *string                               `json:"emergencyDescription" binding:"omitempty,max=2000"`
	ServiceFeatures      *string                               `json:"serviceFeatures" binding:"omitempty,max=5000"`
	InvoiceTitle         *string                               `json:"invoiceTitle" binding:"omitempty,max=128"`
	TaxpayerID           *string                               `json:"taxpayerId" binding:"omitempty,max=32"`
	TaxpayerType         *string                               `json:"taxpayerType" binding:"omitempty,max=32"`
	Qualifications       []SaveInstitutionQualificationRequest `json:"qualifications" binding:"required"`
	Contacts             []SaveInstitutionContactRequest       `json:"contacts" binding:"required"`
	Addresses            []SaveInstitutionAddressRequest       `json:"addresses" binding:"required"`
	BankAccounts         []SaveInstitutionBankAccountRequest   `json:"bankAccounts" binding:"required"`
}

type SaveInstitutionQualificationRequest struct {
	CertificateName  string  `json:"certificateName" binding:"required,max=128"`
	CertificateNo    string  `json:"certificateNo" binding:"required,max=128"`
	IssuingAuthority *string `json:"issuingAuthority" binding:"omitempty,max=128"`
	IssueDate        string  `json:"issueDate" binding:"omitempty"`
	ExpiryDate       string  `json:"expiryDate" binding:"omitempty"`
	Scope            *string `json:"scope" binding:"omitempty,max=512"`
	Remark           *string `json:"remark" binding:"omitempty,max=512"`
	Attachment       *string `json:"attachment" binding:"omitempty,max=512"`
}

type SaveInstitutionContactRequest struct {
	ContactType string  `json:"contactType" binding:"required,max=32"`
	ContactName string  `json:"contactName" binding:"required,max=64"`
	JobTitle    *string `json:"jobTitle" binding:"omitempty,max=64"`
	Phone       *string `json:"phone" binding:"omitempty,max=32"`
	Email       *string `json:"email" binding:"omitempty,max=128"`
	IsPrimary   bool    `json:"isPrimary"`
	Remark      *string `json:"remark" binding:"omitempty,max=512"`
}

type SaveInstitutionAddressRequest struct {
	AddressType string  `json:"addressType" binding:"required,max=32"`
	FullAddress string  `json:"fullAddress" binding:"required,max=512"`
	PostalCode  *string `json:"postalCode" binding:"omitempty,max=16"`
	Phone       *string `json:"phone" binding:"omitempty,max=32"`
	IsPrimary   bool    `json:"isPrimary"`
	Remark      *string `json:"remark" binding:"omitempty,max=512"`
}

type SaveInstitutionBankAccountRequest struct {
	AccountName   string  `json:"accountName" binding:"required,max=128"`
	BankName      string  `json:"bankName" binding:"required,max=128"`
	AccountNumber string  `json:"accountNumber" binding:"required,max=64"`
	AccountType   *string `json:"accountType" binding:"omitempty,max=32"`
	IsDefault     bool    `json:"isDefault"`
	Remark        *string `json:"remark" binding:"omitempty,max=512"`
}

type InstitutionResponse struct {
	InstitutionID        string                             `json:"institutionId"`
	InstitutionName      string                             `json:"institutionName"`
	ShortName            *string                            `json:"shortName"`
	EnglishName          *string                            `json:"englishName"`
	Aliases              *string                            `json:"aliases"`
	InstitutionType      string                             `json:"institutionType"`
	InstitutionNature    string                             `json:"institutionNature"`
	HospitalCategory     string                             `json:"hospitalCategory"`
	HospitalLevel        string                             `json:"hospitalLevel"`
	UnifiedCreditCode    string                             `json:"unifiedCreditCode"`
	EstablishmentDate    *string                            `json:"establishmentDate"`
	Remark               *string                            `json:"remark"`
	LogoURL              *string                            `json:"logoUrl"`
	DisplayName          *string                            `json:"displayName"`
	Slogan               *string                            `json:"slogan"`
	Introduction         *string                            `json:"introduction"`
	DiagnosisSubjects    *string                            `json:"diagnosisSubjects"`
	KeySpecialties       *string                            `json:"keySpecialties"`
	ServiceHours         *string                            `json:"serviceHours"`
	EmergencyDescription *string                            `json:"emergencyDescription"`
	ServiceFeatures      *string                            `json:"serviceFeatures"`
	InvoiceTitle         *string                            `json:"invoiceTitle"`
	TaxpayerID           *string                            `json:"taxpayerId"`
	TaxpayerType         *string                            `json:"taxpayerType"`
	CreateDate           *string                            `json:"createDate"`
	UpdateDate           *string                            `json:"updateDate"`
	Qualifications       []InstitutionQualificationResponse `json:"qualifications"`
	Contacts             []InstitutionContactResponse       `json:"contacts"`
	Addresses            []InstitutionAddressResponse       `json:"addresses"`
	BankAccounts         []InstitutionBankAccountResponse   `json:"bankAccounts"`
}

type InstitutionQualificationResponse struct {
	QualificationID  string  `json:"qualificationId"`
	CertificateName  string  `json:"certificateName"`
	CertificateNo    string  `json:"certificateNo"`
	IssuingAuthority *string `json:"issuingAuthority"`
	IssueDate        *string `json:"issueDate"`
	ExpiryDate       *string `json:"expiryDate"`
	Scope            *string `json:"scope"`
	Remark           *string `json:"remark"`
	Attachment       *string `json:"attachment"`
}

type InstitutionContactResponse struct {
	ContactID   string  `json:"contactId"`
	ContactType string  `json:"contactType"`
	ContactName string  `json:"contactName"`
	JobTitle    *string `json:"jobTitle"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	IsPrimary   bool    `json:"isPrimary"`
	Remark      *string `json:"remark"`
}

type InstitutionAddressResponse struct {
	AddressID   string  `json:"addressId"`
	AddressType string  `json:"addressType"`
	FullAddress string  `json:"fullAddress"`
	PostalCode  *string `json:"postalCode"`
	Phone       *string `json:"phone"`
	IsPrimary   bool    `json:"isPrimary"`
	Remark      *string `json:"remark"`
}

type InstitutionBankAccountResponse struct {
	BankAccountID string  `json:"bankAccountId"`
	AccountName   string  `json:"accountName"`
	BankName      string  `json:"bankName"`
	AccountNumber string  `json:"accountNumber"`
	AccountType   *string `json:"accountType"`
	IsDefault     bool    `json:"isDefault"`
	Remark        *string `json:"remark"`
}

package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrBaseInstitutionInvalidInput = errors.New("机构资料参数错误")
	ErrBaseInstitutionNotFound     = errors.New("机构资料不存在")
	ErrBaseInstitutionConflict     = errors.New("机构资料数据冲突")
)

var institutionCreditCodePattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type BaseInstitutionService struct{}

func NewBaseInstitutionService() *BaseInstitutionService {
	return &BaseInstitutionService{}
}

// GetInstitution 获取当前系统唯一的机构资料聚合。
func (s *BaseInstitutionService) GetInstitution() (*models.InstitutionResponse, error) {
	var institution models.BaseInstitution
	if err := database.DB.Where("singleton_key = ?", "1").First(&institution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var qualifications []models.BaseInstitutionQualification
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).
		Order("qualification_id asc").Find(&qualifications).Error; err != nil {
		return nil, err
	}
	var contacts []models.BaseInstitutionContact
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).
		Order("is_primary desc, contact_type asc, contact_id asc").Find(&contacts).Error; err != nil {
		return nil, err
	}
	var addresses []models.BaseInstitutionAddress
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).
		Order("is_primary desc, address_type asc, address_id asc").Find(&addresses).Error; err != nil {
		return nil, err
	}

	var bankAccounts []models.BaseInstitutionBankAccount
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).
		Order("is_default desc, bank_account_id asc").Find(&bankAccounts).Error; err != nil {
		return nil, err
	}
	return s.institutionToResponse(institution, qualifications, contacts, addresses, bankAccounts), nil
}

// SaveInstitution 保存机构资料聚合。子资料采用当前集合全量替换，不保留历史。
func (s *BaseInstitutionService) SaveInstitution(req models.SaveInstitutionRequest, operatorID string) (*models.InstitutionResponse, error) {
	normalized, err := normalizeInstitutionSaveRequest(req)
	if err != nil {
		return nil, err
	}

	var institutionID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var institution models.BaseInstitution
		err := tx.Where("singleton_key = ?", "1").Clauses(clause.Locking{Strength: "UPDATE"}).First(&institution).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			institutionID = utils.GenerateUUID()
			institution = models.BaseInstitution{
				InstitutionID: institutionID,
				SingletonKey:  "1",
				CreatorID:     optionalInstitutionOperatorID(operatorID),
				CreateDate:    &now,
			}
		} else {
			institutionID = institution.InstitutionID
		}

		institution.InstitutionName = normalized.InstitutionName
		institution.ShortName = normalized.ShortName
		institution.EnglishName = normalized.EnglishName
		institution.Aliases = normalized.Aliases
		institution.InstitutionType = normalized.InstitutionType
		institution.InstitutionNature = normalized.InstitutionNature
		institution.HospitalCategory = normalized.HospitalCategory
		institution.HospitalLevel = normalized.HospitalLevel
		institution.UnifiedCreditCode = normalized.UnifiedCreditCode
		institution.EstablishmentDate = normalized.EstablishmentDate
		institution.Remark = normalized.Remark
		institution.LogoURL = normalized.LogoURL
		institution.DisplayName = normalized.DisplayName
		institution.Slogan = normalized.Slogan
		institution.Introduction = normalized.Introduction
		institution.DiagnosisSubjects = normalized.DiagnosisSubjects
		institution.KeySpecialties = normalized.KeySpecialties
		institution.ServiceHours = normalized.ServiceHours
		institution.EmergencyDescription = normalized.EmergencyDescription
		institution.ServiceFeatures = normalized.ServiceFeatures
		institution.InvoiceTitle = normalized.InvoiceTitle
		institution.TaxpayerID = normalized.TaxpayerID
		institution.TaxpayerType = normalized.TaxpayerType
		institution.UpdaterID = optionalInstitutionOperatorID(operatorID)
		institution.UpdateDate = &now
		if institution.CreateDate == nil {
			institution.CreateDate = &now
		}
		if err := tx.Save(&institution).Error; err != nil {
			return err
		}

		if err := s.replaceInstitutionChildren(tx, institutionID, normalized); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetInstitution()
}

type normalizedInstitutionSave struct {
	InstitutionName      string
	ShortName            *string
	EnglishName          *string
	Aliases              *string
	InstitutionType      string
	InstitutionNature    string
	HospitalCategory     string
	HospitalLevel        string
	UnifiedCreditCode    string
	EstablishmentDate    *time.Time
	Remark               *string
	LogoURL              *string
	DisplayName          *string
	Slogan               *string
	Introduction         *string
	DiagnosisSubjects    *string
	KeySpecialties       *string
	ServiceHours         *string
	EmergencyDescription *string
	ServiceFeatures      *string
	InvoiceTitle         *string
	TaxpayerID           *string
	TaxpayerType         *string
	Qualifications       []normalizedInstitutionQualification
	Contacts             []normalizedInstitutionContact
	Addresses            []normalizedInstitutionAddress
	BankAccounts         []normalizedInstitutionBankAccount
}

type normalizedInstitutionQualification struct {
	CertificateName  string
	CertificateNo    string
	IssuingAuthority *string
	IssueDate        *time.Time
	ExpiryDate       *time.Time
	Scope            *string
	Remark           *string
	Attachment       *string
}

type normalizedInstitutionContact struct {
	ContactType string
	ContactName string
	JobTitle    *string
	Phone       *string
	Email       *string
	IsPrimary   bool
	Remark      *string
}

type normalizedInstitutionAddress struct {
	AddressType string
	FullAddress string
	PostalCode  *string
	Phone       *string
	IsPrimary   bool
	Remark      *string
}

type normalizedInstitutionBankAccount struct {
	AccountName   string
	BankName      string
	AccountNumber string
	AccountType   *string
	IsDefault     bool
	Remark        *string
}

func normalizeInstitutionSaveRequest(req models.SaveInstitutionRequest) (*normalizedInstitutionSave, error) {
	if req.Qualifications == nil || req.Contacts == nil || req.Addresses == nil || req.BankAccounts == nil {
		return nil, fmt.Errorf("%w: 资质、联系人、地址和银行账户必须以完整集合提交，未填写的集合请传空数组", ErrBaseInstitutionInvalidInput)
	}

	institutionName := strings.TrimSpace(req.InstitutionName)
	if institutionName == "" {
		return nil, fmt.Errorf("%w: 机构名称不能为空", ErrBaseInstitutionInvalidInput)
	}
	institutionType, err := normalizeInstitutionEnum(req.InstitutionType, map[string]struct{}{models.InstitutionTypeHospital: {}}, "机构类型")
	if err != nil {
		return nil, err
	}
	institutionNature, err := normalizeInstitutionEnum(req.InstitutionNature, map[string]struct{}{
		models.InstitutionNaturePublic: {}, models.InstitutionNaturePrivate: {}, models.InstitutionNatureNonProfit: {}, models.InstitutionNatureForProfit: {},
	}, "机构性质")
	if err != nil {
		return nil, err
	}
	hospitalCategory, err := normalizeInstitutionEnum(req.HospitalCategory, map[string]struct{}{
		models.InstitutionCategoryGeneral: {}, models.InstitutionCategorySpecialty: {}, models.InstitutionCategoryTraditional: {},
		models.InstitutionCategoryMaternalChild: {}, models.InstitutionCategoryIntegrated: {}, models.InstitutionCategoryRehabilitation: {}, models.InstitutionCategoryOther: {},
	}, "医院类别")
	if err != nil {
		return nil, err
	}
	hospitalLevel, err := normalizeInstitutionEnum(req.HospitalLevel, map[string]struct{}{
		models.InstitutionLevelTertiaryA: {}, models.InstitutionLevelTertiaryB: {}, models.InstitutionLevelTertiaryC: {},
		models.InstitutionLevelSecondaryA: {}, models.InstitutionLevelSecondaryB: {}, models.InstitutionLevelSecondaryC: {},
		models.InstitutionLevelPrimary: {}, models.InstitutionLevelUnrated: {},
	}, "医院等级")
	if err != nil {
		return nil, err
	}
	creditCode := strings.ToUpper(strings.TrimSpace(req.UnifiedCreditCode))
	if creditCode == "" || !institutionCreditCodePattern.MatchString(creditCode) {
		return nil, fmt.Errorf("%w: 统一社会信用代码只能包含字母和数字", ErrBaseInstitutionInvalidInput)
	}
	establishmentDate, err := parseInstitutionDate(req.EstablishmentDate, "成立日期")
	if err != nil {
		return nil, err
	}

	result := &normalizedInstitutionSave{
		InstitutionName:      institutionName,
		ShortName:            normalizeInstitutionOptionalString(req.ShortName),
		EnglishName:          normalizeInstitutionOptionalString(req.EnglishName),
		Aliases:              normalizeInstitutionOptionalString(req.Aliases),
		InstitutionType:      institutionType,
		InstitutionNature:    institutionNature,
		HospitalCategory:     hospitalCategory,
		HospitalLevel:        hospitalLevel,
		UnifiedCreditCode:    creditCode,
		EstablishmentDate:    establishmentDate,
		Remark:               normalizeInstitutionOptionalString(req.Remark),
		LogoURL:              normalizeInstitutionOptionalString(req.LogoURL),
		DisplayName:          normalizeInstitutionOptionalString(req.DisplayName),
		Slogan:               normalizeInstitutionOptionalString(req.Slogan),
		Introduction:         normalizeInstitutionOptionalString(req.Introduction),
		DiagnosisSubjects:    normalizeInstitutionOptionalString(req.DiagnosisSubjects),
		KeySpecialties:       normalizeInstitutionOptionalString(req.KeySpecialties),
		ServiceHours:         normalizeInstitutionOptionalString(req.ServiceHours),
		EmergencyDescription: normalizeInstitutionOptionalString(req.EmergencyDescription),
		ServiceFeatures:      normalizeInstitutionOptionalString(req.ServiceFeatures),
		InvoiceTitle:         normalizeInstitutionOptionalString(req.InvoiceTitle),
		TaxpayerID:           normalizeInstitutionOptionalString(req.TaxpayerID),
		TaxpayerType:         normalizeInstitutionOptionalString(req.TaxpayerType),
	}

	for _, item := range req.Qualifications {
		issueDate, err := parseInstitutionDate(item.IssueDate, "资质发证日期")
		if err != nil {
			return nil, err
		}
		expiryDate, err := parseInstitutionDate(item.ExpiryDate, "资质有效期至")
		if err != nil {
			return nil, err
		}
		if issueDate != nil && expiryDate != nil && expiryDate.Before(*issueDate) {
			return nil, fmt.Errorf("%w: 资质有效期至不能早于发证日期", ErrBaseInstitutionInvalidInput)
		}
		certificateName := strings.TrimSpace(item.CertificateName)
		certificateNo := strings.TrimSpace(item.CertificateNo)
		if certificateName == "" || certificateNo == "" {
			return nil, fmt.Errorf("%w: 资质名称和证书编号不能为空", ErrBaseInstitutionInvalidInput)
		}
		result.Qualifications = append(result.Qualifications, normalizedInstitutionQualification{
			CertificateName:  certificateName,
			CertificateNo:    certificateNo,
			IssuingAuthority: normalizeInstitutionOptionalString(item.IssuingAuthority),
			IssueDate:        issueDate,
			ExpiryDate:       expiryDate,
			Scope:            normalizeInstitutionOptionalString(item.Scope),
			Remark:           normalizeInstitutionOptionalString(item.Remark),
			Attachment:       normalizeInstitutionOptionalString(item.Attachment),
		})
	}

	primaryContacts := 0
	for _, item := range req.Contacts {
		contactType, err := normalizeInstitutionEnum(item.ContactType, institutionContactTypes(), "联系人类型")
		if err != nil {
			return nil, err
		}
		contactName := strings.TrimSpace(item.ContactName)
		if contactName == "" {
			return nil, fmt.Errorf("%w: 联系人姓名不能为空", ErrBaseInstitutionInvalidInput)
		}
		if item.IsPrimary {
			primaryContacts++
		}
		result.Contacts = append(result.Contacts, normalizedInstitutionContact{
			ContactType: contactType,
			ContactName: contactName,
			JobTitle:    normalizeInstitutionOptionalString(item.JobTitle),
			Phone:       normalizeInstitutionOptionalString(item.Phone),
			Email:       normalizeInstitutionOptionalString(item.Email),
			IsPrimary:   item.IsPrimary,
			Remark:      normalizeInstitutionOptionalString(item.Remark),
		})
	}
	if primaryContacts > 1 {
		return nil, fmt.Errorf("%w: 联系方式最多只能设置一个主要联系人", ErrBaseInstitutionInvalidInput)
	}

	primaryAddresses := 0
	for _, item := range req.Addresses {
		addressType, err := normalizeInstitutionEnum(item.AddressType, institutionAddressTypes(), "地址类型")
		if err != nil {
			return nil, err
		}
		fullAddress := strings.TrimSpace(item.FullAddress)
		if fullAddress == "" {
			return nil, fmt.Errorf("%w: 地址不能为空", ErrBaseInstitutionInvalidInput)
		}
		if item.IsPrimary {
			primaryAddresses++
		}
		result.Addresses = append(result.Addresses, normalizedInstitutionAddress{
			AddressType: addressType,
			FullAddress: fullAddress,
			PostalCode:  normalizeInstitutionOptionalString(item.PostalCode),
			Phone:       normalizeInstitutionOptionalString(item.Phone),
			IsPrimary:   item.IsPrimary,
			Remark:      normalizeInstitutionOptionalString(item.Remark),
		})
	}
	if primaryAddresses > 1 {
		return nil, fmt.Errorf("%w: 地址最多只能设置一个主要地址", ErrBaseInstitutionInvalidInput)
	}

	bankAccounts, err := normalizeInstitutionBankAccounts(req.BankAccounts)
	if err != nil {
		return nil, err
	}
	result.BankAccounts = bankAccounts
	return result, nil
}

func (s *BaseInstitutionService) replaceInstitutionChildren(tx *gorm.DB, institutionID string, data *normalizedInstitutionSave) error {
	tables := []interface{}{
		&models.BaseInstitutionQualification{}, &models.BaseInstitutionContact{}, &models.BaseInstitutionAddress{},
		&models.BaseInstitutionBankAccount{},
	}
	for _, table := range tables {
		if err := tx.Where("institution_id = ?", institutionID).Delete(table).Error; err != nil {
			return err
		}
	}

	for _, item := range data.Qualifications {
		qualificationID := utils.GenerateUUID()
		row := models.BaseInstitutionQualification{
			QualificationID: qualificationID, InstitutionID: institutionID,
			CertificateName: item.CertificateName, CertificateNo: item.CertificateNo, IssuingAuthority: item.IssuingAuthority,
			IssueDate: item.IssueDate, ExpiryDate: item.ExpiryDate, Scope: item.Scope, Remark: item.Remark, Attachment: item.Attachment,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	contacts := make([]models.BaseInstitutionContact, 0, len(data.Contacts))
	for _, item := range data.Contacts {
		contacts = append(contacts, models.BaseInstitutionContact{
			ContactID: utils.GenerateUUID(), InstitutionID: institutionID, ContactType: item.ContactType,
			ContactName: item.ContactName, JobTitle: item.JobTitle, Phone: item.Phone, Email: item.Email,
			IsPrimary: boolToTinyInt(item.IsPrimary), Remark: item.Remark,
		})
	}
	if len(contacts) > 0 {
		if err := tx.Create(&contacts).Error; err != nil {
			return err
		}
	}

	addresses := make([]models.BaseInstitutionAddress, 0, len(data.Addresses))
	for _, item := range data.Addresses {
		addresses = append(addresses, models.BaseInstitutionAddress{
			AddressID: utils.GenerateUUID(), InstitutionID: institutionID, AddressType: item.AddressType,
			FullAddress: item.FullAddress, PostalCode: item.PostalCode, Phone: item.Phone,
			IsPrimary: boolToTinyInt(item.IsPrimary), Remark: item.Remark,
		})
	}
	if len(addresses) > 0 {
		if err := tx.Create(&addresses).Error; err != nil {
			return err
		}
	}

	bankAccounts := make([]models.BaseInstitutionBankAccount, 0, len(data.BankAccounts))
	for _, item := range data.BankAccounts {
		bankAccounts = append(bankAccounts, models.BaseInstitutionBankAccount{
			BankAccountID: utils.GenerateUUID(), InstitutionID: institutionID, AccountName: item.AccountName,
			BankName: item.BankName, AccountNumber: item.AccountNumber, AccountType: item.AccountType,
			IsDefault: boolToTinyInt(item.IsDefault), Remark: item.Remark,
		})
	}
	if len(bankAccounts) > 0 {
		if err := tx.Create(&bankAccounts).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *BaseInstitutionService) institutionToResponse(institution models.BaseInstitution, qualifications []models.BaseInstitutionQualification, contacts []models.BaseInstitutionContact, addresses []models.BaseInstitutionAddress, bankAccounts []models.BaseInstitutionBankAccount) *models.InstitutionResponse {
	qualificationResponses := make([]models.InstitutionQualificationResponse, 0, len(qualifications))
	for _, item := range qualifications {
		qualificationResponses = append(qualificationResponses, models.InstitutionQualificationResponse{
			QualificationID: item.QualificationID, CertificateName: item.CertificateName,
			CertificateNo: item.CertificateNo, IssuingAuthority: item.IssuingAuthority, IssueDate: institutionDateString(item.IssueDate),
			ExpiryDate: institutionDateString(item.ExpiryDate), Scope: item.Scope, Remark: item.Remark, Attachment: item.Attachment,
		})
	}
	contactResponses := make([]models.InstitutionContactResponse, 0, len(contacts))
	for _, item := range contacts {
		contactResponses = append(contactResponses, models.InstitutionContactResponse{
			ContactID: item.ContactID, ContactType: item.ContactType, ContactName: item.ContactName, JobTitle: item.JobTitle,
			Phone: item.Phone, Email: item.Email, IsPrimary: item.IsPrimary == 1, Remark: item.Remark,
		})
	}
	addressResponses := make([]models.InstitutionAddressResponse, 0, len(addresses))
	for _, item := range addresses {
		addressResponses = append(addressResponses, models.InstitutionAddressResponse{
			AddressID: item.AddressID, AddressType: item.AddressType, FullAddress: item.FullAddress, PostalCode: item.PostalCode,
			Phone: item.Phone, IsPrimary: item.IsPrimary == 1, Remark: item.Remark,
		})
	}
	bankResponses := make([]models.InstitutionBankAccountResponse, 0, len(bankAccounts))
	for _, item := range bankAccounts {
		bankResponses = append(bankResponses, models.InstitutionBankAccountResponse{
			BankAccountID: item.BankAccountID, AccountName: item.AccountName, BankName: item.BankName,
			AccountNumber: item.AccountNumber, AccountType: item.AccountType, IsDefault: item.IsDefault == 1, Remark: item.Remark,
		})
	}
	return &models.InstitutionResponse{
		InstitutionID: institution.InstitutionID, InstitutionName: institution.InstitutionName, ShortName: institution.ShortName,
		EnglishName: institution.EnglishName, Aliases: institution.Aliases, InstitutionType: institution.InstitutionType,
		InstitutionNature: institution.InstitutionNature, HospitalCategory: institution.HospitalCategory, HospitalLevel: institution.HospitalLevel,
		UnifiedCreditCode: institution.UnifiedCreditCode, EstablishmentDate: institutionDateString(institution.EstablishmentDate), Remark: institution.Remark,
		LogoURL: institution.LogoURL, DisplayName: institution.DisplayName, Slogan: institution.Slogan,
		Introduction: institution.Introduction, DiagnosisSubjects: institution.DiagnosisSubjects, KeySpecialties: institution.KeySpecialties,
		ServiceHours: institution.ServiceHours, EmergencyDescription: institution.EmergencyDescription, ServiceFeatures: institution.ServiceFeatures,
		InvoiceTitle: institution.InvoiceTitle, TaxpayerID: institution.TaxpayerID, TaxpayerType: institution.TaxpayerType,
		CreateDate: models.TimeToStringPtr(institution.CreateDate), UpdateDate: models.TimeToStringPtr(institution.UpdateDate),
		Qualifications: qualificationResponses, Contacts: contactResponses, Addresses: addressResponses, BankAccounts: bankResponses,
	}
}

func institutionContactTypes() map[string]struct{} {
	return map[string]struct{}{
		models.InstitutionContactLegalRepresentative: {}, models.InstitutionContactPrincipal: {}, models.InstitutionContactMedicalQuality: {},
		models.InstitutionContactInformation: {}, models.InstitutionContactFinance: {}, models.InstitutionContactGeneral: {},
	}
}

func institutionAddressTypes() map[string]struct{} {
	return map[string]struct{}{
		models.InstitutionAddressRegistered: {}, models.InstitutionAddressPractice: {}, models.InstitutionAddressMailing: {},
	}
}

func normalizeInstitutionBankAccounts(values []models.SaveInstitutionBankAccountRequest) ([]normalizedInstitutionBankAccount, error) {
	result := make([]normalizedInstitutionBankAccount, 0, len(values))
	defaults := 0
	for _, item := range values {
		accountName := strings.TrimSpace(item.AccountName)
		bankName := strings.TrimSpace(item.BankName)
		accountNumber := strings.TrimSpace(item.AccountNumber)
		if accountName == "" || bankName == "" || accountNumber == "" {
			return nil, fmt.Errorf("%w: 银行账户名称、开户行和账号不能为空", ErrBaseInstitutionInvalidInput)
		}
		if item.IsDefault {
			defaults++
		}
		result = append(result, normalizedInstitutionBankAccount{
			AccountName: accountName, BankName: bankName, AccountNumber: accountNumber,
			AccountType: normalizeInstitutionOptionalString(item.AccountType), IsDefault: item.IsDefault,
			Remark: normalizeInstitutionOptionalString(item.Remark),
		})
	}
	if defaults > 1 {
		return nil, fmt.Errorf("%w: 银行账户最多只能设置一个默认账户", ErrBaseInstitutionInvalidInput)
	}
	return result, nil
}

func normalizeInstitutionEnum(value string, allowed map[string]struct{}, label string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if _, ok := allowed[normalized]; !ok {
		return "", fmt.Errorf("%w: %s不支持", ErrBaseInstitutionInvalidInput, label)
	}
	return normalized, nil
}

func parseInstitutionDate(value, label string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: %s格式必须为YYYY-MM-DD", ErrBaseInstitutionInvalidInput, label)
	}
	return &parsed, nil
}

func institutionDateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	result := value.Format("2006-01-02")
	return &result
}

func normalizeInstitutionOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalInstitutionOperatorID(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func boolToTinyInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

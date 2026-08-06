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

	var overview models.BaseInstitutionOverview
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).First(&overview).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var brand models.BaseInstitutionBrand
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).First(&brand).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var settlement models.BaseInstitutionSettlement
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).First(&settlement).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var bankAccounts []models.BaseInstitutionBankAccount
	if err := database.DB.Where("institution_id = ?", institution.InstitutionID).
		Order("is_default desc, bank_account_id asc").Find(&bankAccounts).Error; err != nil {
		return nil, err
	}
	qualificationIDs := make([]string, 0, len(qualifications))
	for _, item := range qualifications {
		qualificationIDs = append(qualificationIDs, item.QualificationID)
	}
	qualificationAttachments, err := s.getAttachmentsByOwnerIDs(qualificationIDs, models.InstitutionAttachmentOwnerQualification)
	if err != nil {
		return nil, err
	}
	return s.institutionToResponse(institution, qualifications, qualificationAttachments, contacts, addresses,
		optionalInstitutionOverview(overview), optionalInstitutionBrand(brand), optionalInstitutionSettlement(settlement),
		bankAccounts), nil
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
	InstitutionName   string
	ShortName         *string
	EnglishName       *string
	Aliases           *string
	InstitutionType   string
	InstitutionNature string
	HospitalCategory  string
	HospitalLevel     string
	UnifiedCreditCode string
	EstablishmentDate *time.Time
	Remark            *string
	Qualifications    []normalizedInstitutionQualification
	Contacts          []normalizedInstitutionContact
	Addresses         []normalizedInstitutionAddress
	Overview          normalizedInstitutionOverview
	Brand             normalizedInstitutionBrand
	Settlement        normalizedInstitutionSettlement
}

type normalizedInstitutionQualification struct {
	CertificateName  string
	CertificateNo    string
	IssuingAuthority *string
	IssueDate        *time.Time
	ExpiryDate       *time.Time
	Scope            *string
	Remark           *string
	Attachment       *normalizedInstitutionAttachment
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

type normalizedInstitutionOverview struct {
	Introduction         *string
	DiagnosisSubjects    *string
	KeySpecialties       *string
	ServiceHours         *string
	EmergencyDescription *string
	ServiceFeatures      *string
}

type normalizedInstitutionBrand struct {
	LogoURL     *string
	DisplayName *string
	Slogan      *string
}

type normalizedInstitutionSettlement struct {
	InvoiceTitle *string
	TaxpayerID   *string
	TaxpayerType *string
	BankAccounts []normalizedInstitutionBankAccount
}

type normalizedInstitutionBankAccount struct {
	AccountName   string
	BankName      string
	AccountNumber string
	AccountType   *string
	IsDefault     bool
	Remark        *string
}

type normalizedInstitutionAttachment struct {
	AttachmentType *string
	FileName       string
	URL            string
	ExpiryDate     *time.Time
	Remark         *string
}

func normalizeInstitutionSaveRequest(req models.SaveInstitutionRequest) (*normalizedInstitutionSave, error) {
	if req.Qualifications == nil || req.Contacts == nil || req.Addresses == nil || req.Overview == nil || req.Brand == nil || req.Settlement == nil {
		return nil, fmt.Errorf("%w: 子资料必须以完整集合提交，未填写的集合请传空数组或空对象", ErrBaseInstitutionInvalidInput)
	}
	if req.Settlement.BankAccounts == nil {
		return nil, fmt.Errorf("%w: 结算资料的银行账户必须以完整集合提交", ErrBaseInstitutionInvalidInput)
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
		InstitutionName:   institutionName,
		ShortName:         normalizeInstitutionOptionalString(req.ShortName),
		EnglishName:       normalizeInstitutionOptionalString(req.EnglishName),
		Aliases:           normalizeInstitutionOptionalString(req.Aliases),
		InstitutionType:   institutionType,
		InstitutionNature: institutionNature,
		HospitalCategory:  hospitalCategory,
		HospitalLevel:     hospitalLevel,
		UnifiedCreditCode: creditCode,
		EstablishmentDate: establishmentDate,
		Remark:            normalizeInstitutionOptionalString(req.Remark),
		Overview: normalizedInstitutionOverview{
			Introduction:         normalizeInstitutionOptionalString(req.Overview.Introduction),
			DiagnosisSubjects:    normalizeInstitutionOptionalString(req.Overview.DiagnosisSubjects),
			KeySpecialties:       normalizeInstitutionOptionalString(req.Overview.KeySpecialties),
			ServiceHours:         normalizeInstitutionOptionalString(req.Overview.ServiceHours),
			EmergencyDescription: normalizeInstitutionOptionalString(req.Overview.EmergencyDescription),
			ServiceFeatures:      normalizeInstitutionOptionalString(req.Overview.ServiceFeatures),
		},
		Brand: normalizedInstitutionBrand{
			LogoURL:     normalizeInstitutionOptionalString(req.Brand.LogoURL),
			DisplayName: normalizeInstitutionOptionalString(req.Brand.DisplayName),
			Slogan:      normalizeInstitutionOptionalString(req.Brand.Slogan),
		},
		Settlement: normalizedInstitutionSettlement{
			InvoiceTitle: normalizeInstitutionOptionalString(req.Settlement.InvoiceTitle),
			TaxpayerID:   normalizeInstitutionOptionalString(req.Settlement.TaxpayerID),
			TaxpayerType: normalizeInstitutionOptionalString(req.Settlement.TaxpayerType),
		},
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
		attachment, err := normalizeInstitutionAttachment(item.Attachment)
		if err != nil {
			return nil, err
		}
		result.Qualifications = append(result.Qualifications, normalizedInstitutionQualification{
			CertificateName:  certificateName,
			CertificateNo:    certificateNo,
			IssuingAuthority: normalizeInstitutionOptionalString(item.IssuingAuthority),
			IssueDate:        issueDate,
			ExpiryDate:       expiryDate,
			Scope:            normalizeInstitutionOptionalString(item.Scope),
			Remark:           normalizeInstitutionOptionalString(item.Remark),
			Attachment:       attachment,
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

	bankAccounts, err := normalizeInstitutionBankAccounts(req.Settlement.BankAccounts)
	if err != nil {
		return nil, err
	}
	result.Settlement.BankAccounts = bankAccounts
	return result, nil
}

func (s *BaseInstitutionService) replaceInstitutionChildren(tx *gorm.DB, institutionID string, data *normalizedInstitutionSave) error {
	tables := []interface{}{
		&models.BaseInstitutionQualification{}, &models.BaseInstitutionContact{}, &models.BaseInstitutionAddress{},
		&models.BaseInstitutionOverview{}, &models.BaseInstitutionBrand{}, &models.BaseInstitutionSettlement{},
		&models.BaseInstitutionBankAccount{}, &models.BaseInstitutionAttachment{},
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
			IssueDate: item.IssueDate, ExpiryDate: item.ExpiryDate, Scope: item.Scope, Remark: item.Remark,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if item.Attachment != nil {
			if err := s.createInstitutionAttachment(tx, institutionID, models.InstitutionAttachmentOwnerQualification, qualificationID, *item.Attachment); err != nil {
				return err
			}
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

	overview := models.BaseInstitutionOverview{
		OverviewID: utils.GenerateUUID(), InstitutionID: institutionID, Introduction: data.Overview.Introduction,
		DiagnosisSubjects: data.Overview.DiagnosisSubjects, KeySpecialties: data.Overview.KeySpecialties,
		ServiceHours: data.Overview.ServiceHours, EmergencyDescription: data.Overview.EmergencyDescription,
		ServiceFeatures: data.Overview.ServiceFeatures,
	}
	if err := tx.Create(&overview).Error; err != nil {
		return err
	}
	brand := models.BaseInstitutionBrand{
		BrandID: utils.GenerateUUID(), InstitutionID: institutionID, LogoURL: data.Brand.LogoURL,
		DisplayName: data.Brand.DisplayName, Slogan: data.Brand.Slogan,
	}
	if err := tx.Create(&brand).Error; err != nil {
		return err
	}

	settlementID := utils.GenerateUUID()
	settlement := models.BaseInstitutionSettlement{
		SettlementID: settlementID, InstitutionID: institutionID, InvoiceTitle: data.Settlement.InvoiceTitle,
		TaxpayerID: data.Settlement.TaxpayerID, TaxpayerType: data.Settlement.TaxpayerType,
	}
	if err := tx.Create(&settlement).Error; err != nil {
		return err
	}
	bankAccounts := make([]models.BaseInstitutionBankAccount, 0, len(data.Settlement.BankAccounts))
	for _, item := range data.Settlement.BankAccounts {
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

func (s *BaseInstitutionService) createInstitutionAttachment(tx *gorm.DB, institutionID, ownerType, ownerID string, item normalizedInstitutionAttachment) error {
	return tx.Create(&models.BaseInstitutionAttachment{
		AttachmentID: utils.GenerateUUID(), InstitutionID: institutionID, OwnerType: ownerType, OwnerID: ownerID,
		AttachmentType: item.AttachmentType, FileName: item.FileName, URL: item.URL, ExpiryDate: item.ExpiryDate, Remark: item.Remark,
	}).Error
}

func (s *BaseInstitutionService) getAttachmentsByOwnerIDs(ownerIDs []string, ownerType string) (map[string][]models.BaseInstitutionAttachment, error) {
	result := make(map[string][]models.BaseInstitutionAttachment, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return result, nil
	}
	var rows []models.BaseInstitutionAttachment
	if err := database.DB.Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIDs).Order("attachment_id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.OwnerID] = append(result[row.OwnerID], row)
	}
	return result, nil
}

func (s *BaseInstitutionService) institutionToResponse(institution models.BaseInstitution, qualifications []models.BaseInstitutionQualification, qualificationAttachments map[string][]models.BaseInstitutionAttachment, contacts []models.BaseInstitutionContact, addresses []models.BaseInstitutionAddress, overview *models.BaseInstitutionOverview, brand *models.BaseInstitutionBrand, settlement *models.BaseInstitutionSettlement, bankAccounts []models.BaseInstitutionBankAccount) *models.InstitutionResponse {
	qualificationResponses := make([]models.InstitutionQualificationResponse, 0, len(qualifications))
	for _, item := range qualifications {
		qualificationResponses = append(qualificationResponses, models.InstitutionQualificationResponse{
			QualificationID: item.QualificationID, CertificateName: item.CertificateName,
			CertificateNo: item.CertificateNo, IssuingAuthority: item.IssuingAuthority, IssueDate: institutionDateString(item.IssueDate),
			ExpiryDate: institutionDateString(item.ExpiryDate), Scope: item.Scope, Remark: item.Remark,
			Attachment: institutionAttachmentResponse(qualificationAttachments[item.QualificationID]),
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
	var overviewResponse *models.InstitutionOverviewResponse
	if overview != nil {
		overviewResponse = &models.InstitutionOverviewResponse{
			OverviewID: overview.OverviewID, Introduction: overview.Introduction, DiagnosisSubjects: overview.DiagnosisSubjects,
			KeySpecialties: overview.KeySpecialties, ServiceHours: overview.ServiceHours,
			EmergencyDescription: overview.EmergencyDescription, ServiceFeatures: overview.ServiceFeatures,
		}
	}
	var brandResponse *models.InstitutionBrandResponse
	if brand != nil {
		brandResponse = &models.InstitutionBrandResponse{BrandID: brand.BrandID, LogoURL: brand.LogoURL, DisplayName: brand.DisplayName, Slogan: brand.Slogan}
	}
	var settlementResponse *models.InstitutionSettlementResponse
	if settlement != nil {
		bankResponses := make([]models.InstitutionBankAccountResponse, 0, len(bankAccounts))
		for _, item := range bankAccounts {
			bankResponses = append(bankResponses, models.InstitutionBankAccountResponse{
				BankAccountID: item.BankAccountID, AccountName: item.AccountName, BankName: item.BankName,
				AccountNumber: item.AccountNumber, AccountType: item.AccountType, IsDefault: item.IsDefault == 1, Remark: item.Remark,
			})
		}
		settlementResponse = &models.InstitutionSettlementResponse{
			SettlementID: settlement.SettlementID, InvoiceTitle: settlement.InvoiceTitle, TaxpayerID: settlement.TaxpayerID,
			TaxpayerType: settlement.TaxpayerType, BankAccounts: bankResponses,
		}
	}
	return &models.InstitutionResponse{
		InstitutionID: institution.InstitutionID, InstitutionName: institution.InstitutionName, ShortName: institution.ShortName,
		EnglishName: institution.EnglishName, Aliases: institution.Aliases, InstitutionType: institution.InstitutionType,
		InstitutionNature: institution.InstitutionNature, HospitalCategory: institution.HospitalCategory, HospitalLevel: institution.HospitalLevel,
		UnifiedCreditCode: institution.UnifiedCreditCode, EstablishmentDate: institutionDateString(institution.EstablishmentDate), Remark: institution.Remark,
		CreateDate: models.TimeToStringPtr(institution.CreateDate), UpdateDate: models.TimeToStringPtr(institution.UpdateDate),
		Qualifications: qualificationResponses, Contacts: contactResponses, Addresses: addressResponses, Overview: overviewResponse,
		Brand: brandResponse, Settlement: settlementResponse,
	}
}

func institutionAttachmentResponse(rows []models.BaseInstitutionAttachment) *models.InstitutionAttachmentResponse {
	if len(rows) == 0 {
		return nil
	}
	responses := institutionAttachmentResponses(rows)
	return &responses[0]
}

func institutionAttachmentResponses(rows []models.BaseInstitutionAttachment) []models.InstitutionAttachmentResponse {
	result := make([]models.InstitutionAttachmentResponse, 0, len(rows))
	for _, item := range rows {
		result = append(result, models.InstitutionAttachmentResponse{
			AttachmentID: item.AttachmentID, AttachmentType: item.AttachmentType, FileName: item.FileName,
			URL: item.URL, ExpiryDate: institutionDateString(item.ExpiryDate), Remark: item.Remark,
		})
	}
	return result
}

func optionalInstitutionOverview(value models.BaseInstitutionOverview) *models.BaseInstitutionOverview {
	if value.OverviewID == "" {
		return nil
	}
	return &value
}

func optionalInstitutionBrand(value models.BaseInstitutionBrand) *models.BaseInstitutionBrand {
	if value.BrandID == "" {
		return nil
	}
	return &value
}

func optionalInstitutionSettlement(value models.BaseInstitutionSettlement) *models.BaseInstitutionSettlement {
	if value.SettlementID == "" {
		return nil
	}
	return &value
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

func normalizeInstitutionAttachment(value *models.SaveInstitutionAttachmentRequest) (*normalizedInstitutionAttachment, error) {
	if value == nil {
		return nil, nil
	}
	fileName := strings.TrimSpace(value.FileName)
	url := strings.TrimSpace(value.URL)
	if fileName == "" || url == "" {
		return nil, fmt.Errorf("%w: 附件名称和URL不能为空", ErrBaseInstitutionInvalidInput)
	}
	expiryDate, err := parseInstitutionDate(value.ExpiryDate, "附件有效期至")
	if err != nil {
		return nil, err
	}
	return &normalizedInstitutionAttachment{
		AttachmentType: normalizeInstitutionOptionalString(value.AttachmentType), FileName: fileName, URL: url,
		ExpiryDate: expiryDate, Remark: normalizeInstitutionOptionalString(value.Remark),
	}, nil
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

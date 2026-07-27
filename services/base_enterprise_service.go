package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrBaseEnterpriseInvalidInput = errors.New("企业主体参数错误")
	ErrBaseEnterpriseNotFound     = errors.New("企业主体数据不存在")
	ErrBaseEnterpriseConflict     = errors.New("企业主体数据冲突")
)

var enterpriseCreditCodePattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type BaseEnterpriseService struct{}

func NewBaseEnterpriseService() *BaseEnterpriseService {
	return &BaseEnterpriseService{}
}

func (s *BaseEnterpriseService) GetEnterpriseList(req models.EnterpriseListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizeBasePage(req.Page, req.PageSize, 20, 100)
	query := database.DB.Model(&models.BaseEnterprise{}).Where("del_flag = 0")
	query = s.applyEnterpriseFilters(query, req.Keyword, req.EnterpriseType, req.RoleTypes, req.Status)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"enterpriseCode": "enterprise_code",
		"enterpriseName": "enterprise_name",
		"enterpriseType": "enterprise_type",
		"status":         "status",
		"createDate":     "create_date",
		"updateDate":     "update_date",
	})
	if order == "" {
		order = "update_date desc, create_date desc"
	}

	var enterprises []models.BaseEnterprise
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&enterprises).Error; err != nil {
		return nil, err
	}

	responses := s.enterprisesToResponses(enterprises, nil)
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

func (s *BaseEnterpriseService) GetEnterpriseDetail(enterpriseID string) (*models.EnterpriseResponse, error) {
	if err := validateBaseUUID(enterpriseID, "企业主体ID"); err != nil {
		return nil, err
	}

	var enterprise models.BaseEnterprise
	if err := database.DB.Where("enterprise_id = ? AND del_flag = 0", enterpriseID).First(&enterprise).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 企业主体不存在", ErrBaseEnterpriseNotFound)
		}
		return nil, err
	}

	roles, err := s.getRolesByEnterpriseIDs(database.DB, []string{enterpriseID})
	if err != nil {
		return nil, err
	}
	return s.enterpriseToResponse(enterprise, roles[enterpriseID]), nil
}

func (s *BaseEnterpriseService) CreateEnterprise(req models.SaveEnterpriseRequest, operatorID string) (*models.EnterpriseResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureEnterpriseUnique(tx, normalized.EnterpriseNameNormalized, normalized.UnifiedCreditCode, ""); err != nil {
			return err
		}
		code, err := s.nextEnterpriseCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		enterprise := models.BaseEnterprise{
			EnterpriseID:             createdID,
			EnterpriseCode:           code,
			EnterpriseName:           normalized.EnterpriseName,
			EnterpriseNameNormalized: normalized.EnterpriseNameNormalized,
			ShortName:                normalized.ShortName,
			ShortNameNormalized:      normalized.ShortNameNormalized,
			UnifiedCreditCode:        normalized.UnifiedCreditCode,
			EnterpriseType:           normalized.EnterpriseType,
			ContactName:              normalized.ContactName,
			ContactPhone:             normalized.ContactPhone,
			Address:                  normalized.Address,
			Status:                   normalized.Status,
			Remark:                   normalized.Remark,
			RowVersion:               1,
			CreatorID:                optionalBaseOperatorID(operatorID),
			UpdaterID:                optionalBaseOperatorID(operatorID),
			CreateDate:               &now,
			UpdateDate:               &now,
			DelFlag:                  0,
		}
		if err := tx.Create(&enterprise).Error; err != nil {
			return err
		}
		return s.replaceEnterpriseRoles(tx, createdID, normalized.Roles, now)
	}); err != nil {
		return nil, err
	}

	return s.GetEnterpriseDetail(createdID)
}

func (s *BaseEnterpriseService) UpdateEnterprise(enterpriseID string, req models.SaveEnterpriseRequest, operatorID string) (*models.EnterpriseResponse, error) {
	if err := validateBaseUUID(enterpriseID, "企业主体ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var enterprise models.BaseEnterprise
		if err := tx.Where("enterprise_id = ? AND del_flag = 0", enterpriseID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&enterprise).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 企业主体不存在", ErrBaseEnterpriseNotFound)
			}
			return err
		}
		if enterprise.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 企业主体已被其他人修改，请刷新后重试", ErrBaseEnterpriseConflict)
		}
		if err := s.ensureEnterpriseUnique(tx, normalized.EnterpriseNameNormalized, normalized.UnifiedCreditCode, enterpriseID); err != nil {
			return err
		}

		now := time.Now()
		updates := map[string]interface{}{
			"enterprise_name":            normalized.EnterpriseName,
			"enterprise_name_normalized": normalized.EnterpriseNameNormalized,
			"short_name":                 normalized.ShortName,
			"short_name_normalized":      normalized.ShortNameNormalized,
			"unified_credit_code":        normalized.UnifiedCreditCode,
			"enterprise_type":            normalized.EnterpriseType,
			"contact_name":               normalized.ContactName,
			"contact_phone":              normalized.ContactPhone,
			"address":                    normalized.Address,
			"status":                     normalized.Status,
			"remark":                     normalized.Remark,
			"row_version":                enterprise.RowVersion + 1,
			"updater_id":                 optionalBaseOperatorID(operatorID),
			"update_date":                now,
		}
		if err := tx.Model(&models.BaseEnterprise{}).Where("enterprise_id = ?", enterpriseID).Updates(updates).Error; err != nil {
			return err
		}
		return s.replaceEnterpriseRoles(tx, enterpriseID, normalized.Roles, now)
	}); err != nil {
		return nil, err
	}

	return s.GetEnterpriseDetail(enterpriseID)
}

func (s *BaseEnterpriseService) UpdateEnterpriseStatus(enterpriseID string, req models.UpdateEnterpriseStatusRequest, operatorID string) (*models.EnterpriseResponse, error) {
	if err := validateBaseUUID(enterpriseID, "企业主体ID"); err != nil {
		return nil, err
	}
	if err := validateBaseStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var enterprise models.BaseEnterprise
		if err := tx.Where("enterprise_id = ? AND del_flag = 0", enterpriseID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&enterprise).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 企业主体不存在", ErrBaseEnterpriseNotFound)
			}
			return err
		}
		if enterprise.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 企业主体已被其他人修改，请刷新后重试", ErrBaseEnterpriseConflict)
		}

		now := time.Now()
		return tx.Model(&models.BaseEnterprise{}).Where("enterprise_id = ?", enterpriseID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": enterprise.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetEnterpriseDetail(enterpriseID)
}

func (s *BaseEnterpriseService) GetEnterpriseOptions(req models.EnterpriseOptionsRequest) ([]models.EnterpriseOptionResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.BaseEnterprise{}).Where("del_flag = 0 AND status = 1")
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.ToLower(strings.TrimSpace(req.Keyword)) + "%"
		query = query.Where(
			"LOWER(enterprise_name) LIKE ? OR LOWER(enterprise_code) LIKE ? OR LOWER(IFNULL(short_name, '')) LIKE ? OR LOWER(IFNULL(unified_credit_code, '')) LIKE ?",
			keyword, keyword, keyword, keyword,
		)
	}
	if strings.TrimSpace(req.RoleType) != "" {
		role, err := normalizeEnterpriseRole(req.RoleType)
		if err != nil {
			return nil, err
		}
		query = query.Where("EXISTS (SELECT 1 FROM base_enterprise_role r WHERE r.enterprise_id = base_enterprise.enterprise_id AND r.role_type = ?)", role)
	}

	var enterprises []models.BaseEnterprise
	if err := query.Order("enterprise_code asc").Limit(pageSize).Find(&enterprises).Error; err != nil {
		return nil, err
	}
	ids := enterpriseIDs(enterprises)
	roleMap, err := s.getRolesByEnterpriseIDs(database.DB, ids)
	if err != nil {
		return nil, err
	}

	options := make([]models.EnterpriseOptionResponse, 0, len(enterprises))
	for _, enterprise := range enterprises {
		options = append(options, models.EnterpriseOptionResponse{
			EnterpriseID:   enterprise.EnterpriseID,
			EnterpriseCode: enterprise.EnterpriseCode,
			EnterpriseName: enterprise.EnterpriseName,
			ShortName:      enterprise.ShortName,
			EnterpriseType: enterprise.EnterpriseType,
			Roles:          roleMap[enterprise.EnterpriseID],
		})
	}
	return options, nil
}

type normalizedEnterpriseSave struct {
	EnterpriseName           string
	EnterpriseNameNormalized string
	ShortName                *string
	ShortNameNormalized      *string
	UnifiedCreditCode        *string
	EnterpriseType           string
	Roles                    []string
	ContactName              *string
	ContactPhone             *string
	Address                  *string
	Status                   int
	Remark                   *string
	ExpectedRowVersion       int
}

func (s *BaseEnterpriseService) normalizeSaveRequest(req models.SaveEnterpriseRequest, requireVersion bool) (*normalizedEnterpriseSave, error) {
	enterpriseName := strings.TrimSpace(req.EnterpriseName)
	if enterpriseName == "" {
		return nil, fmt.Errorf("%w: 企业名称不能为空", ErrBaseEnterpriseInvalidInput)
	}
	if len([]rune(enterpriseName)) > 128 {
		return nil, fmt.Errorf("%w: 企业名称不能超过128个字符", ErrBaseEnterpriseInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrBaseEnterpriseInvalidInput)
	}
	if err := validateBaseStatus(req.Status); err != nil {
		return nil, err
	}
	enterpriseType, err := normalizeEnterpriseType(req.EnterpriseType)
	if err != nil {
		return nil, err
	}
	roles, err := normalizeEnterpriseRoles(req.Roles)
	if err != nil {
		return nil, err
	}
	creditCode, err := normalizeUnifiedCreditCode(req.UnifiedCreditCode)
	if err != nil {
		return nil, err
	}
	shortName := normalizeBaseOptionalString(req.ShortName)
	var shortNameNormalized *string
	if shortName != nil {
		value := normalizeBaseText(*shortName)
		shortNameNormalized = &value
	}

	return &normalizedEnterpriseSave{
		EnterpriseName:           enterpriseName,
		EnterpriseNameNormalized: normalizeBaseText(enterpriseName),
		ShortName:                shortName,
		ShortNameNormalized:      shortNameNormalized,
		UnifiedCreditCode:        creditCode,
		EnterpriseType:           enterpriseType,
		Roles:                    roles,
		ContactName:              normalizeBaseOptionalString(req.ContactName),
		ContactPhone:             normalizeBaseOptionalString(req.ContactPhone),
		Address:                  normalizeBaseOptionalString(req.Address),
		Status:                   req.Status,
		Remark:                   normalizeBaseOptionalString(req.Remark),
		ExpectedRowVersion:       req.ExpectedRowVersion,
	}, nil
}

func (s *BaseEnterpriseService) applyEnterpriseFilters(query *gorm.DB, keyword, enterpriseType, roleTypes string, status *int) *gorm.DB {
	if strings.TrimSpace(keyword) != "" {
		kw := "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
		query = query.Where(
			"LOWER(enterprise_name) LIKE ? OR LOWER(enterprise_code) LIKE ? OR LOWER(IFNULL(short_name, '')) LIKE ? OR LOWER(IFNULL(unified_credit_code, '')) LIKE ?",
			kw, kw, kw, kw,
		)
	}
	if strings.TrimSpace(enterpriseType) != "" {
		if normalized, err := normalizeEnterpriseType(enterpriseType); err == nil {
			query = query.Where("enterprise_type = ?", normalized)
		}
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	roles := parseEnterpriseRoleTypes(roleTypes)
	if len(roles) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM base_enterprise_role r WHERE r.enterprise_id = base_enterprise.enterprise_id AND r.role_type IN ?)", roles)
	}
	return query
}

func (s *BaseEnterpriseService) ensureEnterpriseUnique(tx *gorm.DB, nameNormalized string, creditCode *string, excludeID string) error {
	query := tx.Model(&models.BaseEnterprise{}).Where("del_flag = 0 AND enterprise_name_normalized = ?", nameNormalized)
	if excludeID != "" {
		query = query.Where("enterprise_id != ?", excludeID)
	}
	var nameCount int64
	if err := query.Count(&nameCount).Error; err != nil {
		return err
	}
	if nameCount > 0 {
		return fmt.Errorf("%w: 企业名称已存在", ErrBaseEnterpriseConflict)
	}

	if creditCode == nil {
		return nil
	}
	creditQuery := tx.Model(&models.BaseEnterprise{}).Where("del_flag = 0 AND unified_credit_code = ?", *creditCode)
	if excludeID != "" {
		creditQuery = creditQuery.Where("enterprise_id != ?", excludeID)
	}
	var creditCount int64
	if err := creditQuery.Count(&creditCount).Error; err != nil {
		return err
	}
	if creditCount > 0 {
		return fmt.Errorf("%w: 统一社会信用代码已存在", ErrBaseEnterpriseConflict)
	}
	return nil
}

func (s *BaseEnterpriseService) nextEnterpriseCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ENTERPRISE", "ENT", 6)
}

func (s *BaseEnterpriseService) replaceEnterpriseRoles(tx *gorm.DB, enterpriseID string, roles []string, now time.Time) error {
	if err := tx.Where("enterprise_id = ?", enterpriseID).Delete(&models.BaseEnterpriseRole{}).Error; err != nil {
		return err
	}
	rows := make([]models.BaseEnterpriseRole, 0, len(roles))
	for _, role := range roles {
		rows = append(rows, models.BaseEnterpriseRole{
			EnterpriseRoleID: utils.GenerateUUID(),
			EnterpriseID:     enterpriseID,
			RoleType:         role,
			CreateDate:       &now,
		})
	}
	return tx.Create(&rows).Error
}

func (s *BaseEnterpriseService) enterprisesToResponses(enterprises []models.BaseEnterprise, roleMap map[string][]string) []models.EnterpriseResponse {
	if roleMap == nil {
		var err error
		roleMap, err = s.getRolesByEnterpriseIDs(database.DB, enterpriseIDs(enterprises))
		if err != nil {
			roleMap = map[string][]string{}
		}
	}
	responses := make([]models.EnterpriseResponse, 0, len(enterprises))
	for _, enterprise := range enterprises {
		responses = append(responses, *s.enterpriseToResponse(enterprise, roleMap[enterprise.EnterpriseID]))
	}
	return responses
}

func (s *BaseEnterpriseService) enterpriseToResponse(enterprise models.BaseEnterprise, roles []string) *models.EnterpriseResponse {
	if roles == nil {
		roles = []string{}
	}
	return &models.EnterpriseResponse{
		EnterpriseID:      enterprise.EnterpriseID,
		EnterpriseCode:    enterprise.EnterpriseCode,
		EnterpriseName:    enterprise.EnterpriseName,
		ShortName:         enterprise.ShortName,
		UnifiedCreditCode: enterprise.UnifiedCreditCode,
		EnterpriseType:    enterprise.EnterpriseType,
		Roles:             roles,
		ContactName:       enterprise.ContactName,
		ContactPhone:      enterprise.ContactPhone,
		Address:           enterprise.Address,
		Status:            enterprise.Status,
		Remark:            enterprise.Remark,
		RowVersion:        enterprise.RowVersion,
		CreateDate:        models.TimeToStringPtr(enterprise.CreateDate),
		UpdateDate:        models.TimeToStringPtr(enterprise.UpdateDate),
	}
}

func (s *BaseEnterpriseService) getRolesByEnterpriseIDs(tx *gorm.DB, enterpriseIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(enterpriseIDs))
	if len(enterpriseIDs) == 0 {
		return result, nil
	}
	var roles []models.BaseEnterpriseRole
	if err := tx.Where("enterprise_id IN ?", enterpriseIDs).Order("role_type asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	for _, role := range roles {
		result[role.EnterpriseID] = append(result[role.EnterpriseID], role.RoleType)
	}
	return result, nil
}

func enterpriseIDs(enterprises []models.BaseEnterprise) []string {
	ids := make([]string, 0, len(enterprises))
	for _, enterprise := range enterprises {
		ids = append(ids, enterprise.EnterpriseID)
	}
	return ids
}

func validateBaseUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrBaseEnterpriseInvalidInput, label)
	}
	return nil
}

func validateBaseStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrBaseEnterpriseInvalidInput)
	}
	return nil
}

func normalizeEnterpriseType(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.EnterpriseTypeEnterprise,
		models.EnterpriseTypeMedicalOrg,
		models.EnterpriseTypeIndividual,
		models.EnterpriseTypePublicInstitution,
		models.EnterpriseTypeOther:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 企业类型不支持", ErrBaseEnterpriseInvalidInput)
	}
}

func normalizeEnterpriseRole(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.EnterpriseRoleManufacturer,
		models.EnterpriseRoleMAH,
		models.EnterpriseRoleRegistrant,
		models.EnterpriseRoleFiler,
		models.EnterpriseRoleImportAgent,
		models.EnterpriseRoleSupplier,
		models.EnterpriseRoleDistributor,
		models.EnterpriseRoleDealer,
		models.EnterpriseRoleCustomer:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 企业角色不支持", ErrBaseEnterpriseInvalidInput)
	}
}

func normalizeEnterpriseRoles(values []string) ([]string, error) {
	seen := map[string]struct{}{}
	roles := make([]string, 0, len(values))
	for _, value := range values {
		role, err := normalizeEnterpriseRole(value)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("%w: 企业主体至少需要一个角色", ErrBaseEnterpriseInvalidInput)
	}
	return roles, nil
}

func parseEnterpriseRoleTypes(value string) []string {
	rawItems := strings.Split(value, ",")
	roles := make([]string, 0, len(rawItems))
	seen := map[string]struct{}{}
	for _, raw := range rawItems {
		role, err := normalizeEnterpriseRole(raw)
		if err != nil {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	return roles
}

func normalizeUnifiedCreditCode(value *string) (*string, error) {
	normalized := normalizeBaseOptionalString(value)
	if normalized == nil {
		return nil, nil
	}
	upper := strings.ToUpper(*normalized)
	if !enterpriseCreditCodePattern.MatchString(upper) {
		return nil, fmt.Errorf("%w: 统一社会信用代码只能包含字母和数字", ErrBaseEnterpriseInvalidInput)
	}
	return &upper, nil
}

func normalizeBaseOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeBaseText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func normalizeBasePage(page, pageSize, defaultSize, maxSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return page, pageSize
}

func optionalBaseOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

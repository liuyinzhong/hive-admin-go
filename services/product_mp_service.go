package services

import (
	"errors"
	"fmt"
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
	ErrProductMpInvalidInput = errors.New("厂家产品参数错误")
	ErrProductMpNotFound     = errors.New("厂家产品数据不存在")
	ErrProductMpConflict     = errors.New("厂家产品数据冲突")
)

type ProductMpService struct{}

func NewProductMpService() *ProductMpService {
	return &ProductMpService{}
}

func (s *ProductMpService) GetProductMpList(req models.ProductMpListRequest) (*utils.PaginationResponse, error) {
	if err := validateProductMpUUID(req.RpID, "规格产品ID"); err != nil {
		return nil, err
	}

	var rows []productMpQueryRow
	if err := s.baseProductMpQuery().
		Select(productMpSelectFields()).
		Where("product_mp.rp_id = ?", strings.TrimSpace(req.RpID)).
		Order("product_mp.create_date desc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: productMpRowsToResponses(rows), Total: int64(len(rows))}, nil
}

func (s *ProductMpService) GetProductMpDetail(mpID string) (*models.ProductMpResponse, error) {
	if err := validateProductMpUUID(mpID, "厂家产品ID"); err != nil {
		return nil, err
	}

	var row productMpQueryRow
	if err := s.baseProductMpQuery().
		Select(productMpSelectFields()).
		Where("product_mp.mp_id = ?", mpID).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.MpID == "" {
		return nil, fmt.Errorf("%w: 厂家产品不存在", ErrProductMpNotFound)
	}
	return productMpRowToResponse(row), nil
}

func (s *ProductMpService) CreateProductMp(req models.SaveProductMpRequest, operatorID string) (*models.ProductMpResponse, error) {
	normalized, err := s.normalizeSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.getExistingProductRp(tx, normalized.RpID); err != nil {
			return err
		}
		if _, err := s.getSelectableProductionEnterprise(tx, normalized.EnterpriseID); err != nil {
			return err
		}
		if err := s.ensureProductMpUnique(tx, normalized.RpID, normalized.EnterpriseID, normalized.ApprovalNoNormalized, ""); err != nil {
			return err
		}
		code, err := s.nextProductMpCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		mp := models.ProductMp{
			MpID:                 createdID,
			MpCode:               code,
			RpID:                 normalized.RpID,
			EnterpriseID:         normalized.EnterpriseID,
			ApprovalNo:           normalized.ApprovalNo,
			ApprovalNoNormalized: normalized.ApprovalNoNormalized,
			BrandName:            normalized.BrandName,
			Description:          normalized.Description,
			Status:               normalized.Status,
			RowVersion:           1,
			CreatorID:            optionalProductMpOperatorID(operatorID),
			UpdaterID:            optionalProductMpOperatorID(operatorID),
			CreateDate:           &now,
			UpdateDate:           &now,
			DelFlag:              0,
		}
		return tx.Create(&mp).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductMpDetail(createdID)
}

func (s *ProductMpService) UpdateProductMp(mpID string, req models.SaveProductMpRequest, operatorID string) (*models.ProductMpResponse, error) {
	if err := validateProductMpUUID(mpID, "厂家产品ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var mp models.ProductMp
		if err := tx.Where("mp_id = ? AND del_flag = 0", mpID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&mp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 厂家产品不存在", ErrProductMpNotFound)
			}
			return err
		}
		if mp.RpID != normalized.RpID {
			return fmt.Errorf("%w: 不允许更换所属规格产品", ErrProductMpInvalidInput)
		}
		if mp.EnterpriseID != normalized.EnterpriseID {
			return fmt.Errorf("%w: 不允许更换生产企业", ErrProductMpInvalidInput)
		}
		if mp.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 厂家产品已被其他人修改，请刷新后重试", ErrProductMpConflict)
		}
		if err := s.ensureProductMpUnique(tx, mp.RpID, mp.EnterpriseID, normalized.ApprovalNoNormalized, mpID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ProductMp{}).Where("mp_id = ?", mpID).Updates(map[string]interface{}{
			"approval_no":            normalized.ApprovalNo,
			"approval_no_normalized": normalized.ApprovalNoNormalized,
			"brand_name":             normalized.BrandName,
			"description":            normalized.Description,
			"status":                 normalized.Status,
			"row_version":            mp.RowVersion + 1,
			"updater_id":             optionalProductMpOperatorID(operatorID),
			"update_date":            now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductMpDetail(mpID)
}

func (s *ProductMpService) UpdateProductMpStatus(mpID string, req models.UpdateProductMpStatusRequest, operatorID string) (*models.ProductMpResponse, error) {
	if err := validateProductMpUUID(mpID, "厂家产品ID"); err != nil {
		return nil, err
	}
	if err := validateProductMpStatus(req.Status); err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var mp models.ProductMp
		if err := tx.Where("mp_id = ? AND del_flag = 0", mpID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&mp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 厂家产品不存在", ErrProductMpNotFound)
			}
			return err
		}
		if mp.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 厂家产品已被其他人修改，请刷新后重试", ErrProductMpConflict)
		}

		now := time.Now()
		return tx.Model(&models.ProductMp{}).Where("mp_id = ?", mpID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": mp.RowVersion + 1,
			"updater_id":  optionalProductMpOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetProductMpDetail(mpID)
}

type normalizedProductMpSave struct {
	RpID                 string
	EnterpriseID         string
	ApprovalNo           string
	ApprovalNoNormalized string
	BrandName            *string
	Description          *string
	Status               int
	ExpectedRowVersion   int
}

type productMpQueryRow struct {
	MpID           string
	MpCode         string
	RpID           string
	RpCode         string
	SpecName       string
	SpuID          string
	SpuCode        string
	ProductName    string
	ProductType    string
	EnterpriseID   string
	EnterpriseCode string
	EnterpriseName string
	ApprovalNo     string
	BrandName      *string
	Description    *string
	Status         int
	RowVersion     int
	CreateDate     *time.Time
	UpdateDate     *time.Time
}

func (s *ProductMpService) normalizeSaveRequest(req models.SaveProductMpRequest, requireVersion bool) (*normalizedProductMpSave, error) {
	if err := validateProductMpUUID(req.RpID, "规格产品ID"); err != nil {
		return nil, err
	}
	if err := validateProductMpUUID(req.EnterpriseID, "生产企业ID"); err != nil {
		return nil, err
	}
	approvalNo := strings.TrimSpace(req.ApprovalNo)
	if approvalNo == "" {
		return nil, fmt.Errorf("%w: 批准文号/注册证号/备案号不能为空", ErrProductMpInvalidInput)
	}
	if len([]rune(approvalNo)) > 128 {
		return nil, fmt.Errorf("%w: 批准文号/注册证号/备案号不能超过128个字符", ErrProductMpInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrProductMpInvalidInput)
	}
	if err := validateProductMpStatus(req.Status); err != nil {
		return nil, err
	}

	brandName := normalizeProductMpOptionalString(req.BrandName)
	if brandName != nil && len([]rune(*brandName)) > 128 {
		return nil, fmt.Errorf("%w: 品牌/商品名不能超过128个字符", ErrProductMpInvalidInput)
	}
	description := normalizeProductMpOptionalString(req.Description)
	if description != nil && len([]rune(*description)) > 2000 {
		return nil, fmt.Errorf("%w: 描述不能超过2000个字符", ErrProductMpInvalidInput)
	}

	return &normalizedProductMpSave{
		RpID:                 strings.TrimSpace(req.RpID),
		EnterpriseID:         strings.TrimSpace(req.EnterpriseID),
		ApprovalNo:           approvalNo,
		ApprovalNoNormalized: normalizeProductMpText(approvalNo),
		BrandName:            brandName,
		Description:          description,
		Status:               req.Status,
		ExpectedRowVersion:   req.ExpectedRowVersion,
	}, nil
}

func (s *ProductMpService) baseProductMpQuery() *gorm.DB {
	return database.DB.Table("product_mp").
		Joins("INNER JOIN product_rp ON product_rp.rp_id = product_mp.rp_id AND product_rp.del_flag = 0").
		Joins("INNER JOIN product_spu ON product_spu.spu_id = product_rp.spu_id AND product_spu.del_flag = 0").
		Joins("INNER JOIN base_enterprise ON base_enterprise.enterprise_id = product_mp.enterprise_id AND base_enterprise.del_flag = 0").
		Where("product_mp.del_flag = 0")
}

func (s *ProductMpService) getExistingProductRp(tx *gorm.DB, rpID string) (*models.ProductRp, error) {
	var rp models.ProductRp
	if err := tx.Where("rp_id = ? AND del_flag = 0", rpID).First(&rp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 所属规格产品不存在", ErrProductMpNotFound)
		}
		return nil, err
	}
	return &rp, nil
}

func (s *ProductMpService) getSelectableProductionEnterprise(tx *gorm.DB, enterpriseID string) (*models.BaseEnterprise, error) {
	var enterprise models.BaseEnterprise
	if err := tx.Where("enterprise_id = ? AND del_flag = 0 AND status = 1", enterpriseID).First(&enterprise).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 生产企业不存在或已停用", ErrProductMpNotFound)
		}
		return nil, err
	}
	var count int64
	if err := tx.Model(&models.BaseEnterpriseRole{}).
		Where("enterprise_id = ? AND role_type = ?", enterpriseID, models.EnterpriseRoleManufacturer).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("%w: 生产企业必须具备生产企业角色", ErrProductMpInvalidInput)
	}
	return &enterprise, nil
}

func (s *ProductMpService) ensureProductMpUnique(tx *gorm.DB, rpID, enterpriseID, approvalNoNormalized, excludeID string) error {
	query := tx.Model(&models.ProductMp{}).
		Where("del_flag = 0 AND rp_id = ? AND enterprise_id = ? AND approval_no_normalized = ?", rpID, enterpriseID, approvalNoNormalized)
	if excludeID != "" {
		query = query.Where("mp_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前规格产品下生产企业和批准文号组合已存在", ErrProductMpConflict)
	}
	return nil
}

func (s *ProductMpService) nextProductMpCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "PRODUCT_MP", "MP", 6)
}

func productMpSelectFields() string {
	return strings.Join([]string{
		"product_mp.mp_id",
		"product_mp.mp_code",
		"product_mp.rp_id",
		"product_rp.rp_code",
		"product_rp.spec_name",
		"product_rp.spu_id",
		"product_spu.spu_code",
		"product_spu.product_name",
		"product_spu.product_type",
		"product_mp.enterprise_id",
		"base_enterprise.enterprise_code",
		"base_enterprise.enterprise_name",
		"product_mp.approval_no",
		"product_mp.brand_name",
		"product_mp.description",
		"product_mp.status",
		"product_mp.row_version",
		"product_mp.create_date",
		"product_mp.update_date",
	}, ", ")
}

func productMpRowsToResponses(rows []productMpQueryRow) []models.ProductMpResponse {
	responses := make([]models.ProductMpResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, *productMpRowToResponse(row))
	}
	return responses
}

func productMpRowToResponse(row productMpQueryRow) *models.ProductMpResponse {
	return &models.ProductMpResponse{
		MpID:           row.MpID,
		MpCode:         row.MpCode,
		RpID:           row.RpID,
		RpCode:         row.RpCode,
		SpecName:       row.SpecName,
		SpuID:          row.SpuID,
		SpuCode:        row.SpuCode,
		ProductName:    row.ProductName,
		ProductType:    row.ProductType,
		EnterpriseID:   row.EnterpriseID,
		EnterpriseCode: row.EnterpriseCode,
		EnterpriseName: row.EnterpriseName,
		ApprovalNo:     row.ApprovalNo,
		BrandName:      row.BrandName,
		Description:    row.Description,
		Status:         row.Status,
		RowVersion:     row.RowVersion,
		CreateDate:     models.TimeToStringPtr(row.CreateDate),
		UpdateDate:     models.TimeToStringPtr(row.UpdateDate),
	}
}

func validateProductMpUUID(value, label string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s格式错误", ErrProductMpInvalidInput, label)
	}
	return nil
}

func validateProductMpStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("%w: 状态只能是0或1", ErrProductMpInvalidInput)
	}
	return nil
}

func normalizeProductMpOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeProductMpText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func optionalProductMpOperatorID(operatorID string) *string {
	if strings.TrimSpace(operatorID) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(operatorID)
	return &trimmed
}

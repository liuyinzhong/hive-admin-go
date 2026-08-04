package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

var (
	ErrPrintTemplateInvalidInput = errors.New("打印模板参数错误")
	ErrPrintTemplateNotFound     = errors.New("打印模板不存在")
	ErrPrintTemplateConflict     = errors.New("打印模板数据冲突")
	ErrPrintTemplateUnavailable  = errors.New("未配置可用打印模板")
)

type PrintTemplateService struct{}

func NewPrintTemplateService() *PrintTemplateService {
	return &PrintTemplateService{}
}

func (s *PrintTemplateService) GetPrintTemplateList(req models.PrintTemplateListRequest) (*utils.PaginationResponse, error) {
	page, pageSize := normalizePrintTemplatePage(req.Page, req.PageSize)
	query := database.DB.Model(&models.PrintTemplate{})

	if value := strings.TrimSpace(req.DocumentType); value != "" {
		if !isSupportedPrintDocumentType(value) {
			return nil, fmt.Errorf("%w: 不支持的单据类型", ErrPrintTemplateInvalidInput)
		}
		query = query.Where("document_type = ?", value)
	}
	if value := strings.TrimSpace(req.Status); value != "" {
		if !isSupportedPrintTemplateStatus(value) {
			return nil, fmt.Errorf("%w: 不支持的模板状态", ErrPrintTemplateInvalidInput)
		}
		query = query.Where("status = ?", value)
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"documentType": "document_type",
		"templateName": "template_name",
		"status":       "status",
		"createDate":   "create_date",
		"updateDate":   "update_date",
	})
	if order == "" {
		order = "update_date desc"
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var templates []models.PrintTemplate
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, err
	}

	items := make([]models.PrintTemplateListResponse, 0, len(templates))
	for _, template := range templates {
		items = append(items, printTemplateToListResponse(template))
	}
	return &utils.PaginationResponse{Items: items, Total: total}, nil
}

func (s *PrintTemplateService) GetPrintTemplateMetadata() *models.PrintTemplateMetadataResponse {
	return &models.PrintTemplateMetadataResponse{
		DocumentTypes: printDocumentTypes(),
		FieldGroups:   printFieldGroups(),
	}
}

func (s *PrintTemplateService) GetPrintTemplateDetail(templateID string) (*models.PrintTemplateResponse, error) {
	template, err := s.findTemplate(templateID)
	if err != nil {
		return nil, err
	}
	return printTemplateToResponse(template)
}

func (s *PrintTemplateService) GetPublishedPrintTemplate(documentType string) (*models.PrintTemplateResponse, error) {
	if !isSupportedPrintDocumentType(documentType) {
		return nil, fmt.Errorf("%w: 不支持的单据类型", ErrPrintTemplateInvalidInput)
	}

	var template models.PrintTemplate
	if err := database.DB.Where("document_type = ?", documentType).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrintTemplateUnavailable
		}
		return nil, err
	}
	if template.PublishedLayout == nil || strings.TrimSpace(*template.PublishedLayout) == "" {
		return nil, ErrPrintTemplateUnavailable
	}
	return printTemplateToResponse(template)
}

func (s *PrintTemplateService) CreatePrintTemplate(req models.CreatePrintTemplateRequest, operatorID string) (*models.PrintTemplateResponse, error) {
	documentType := strings.TrimSpace(req.DocumentType)
	templateName := strings.TrimSpace(req.TemplateName)
	if !isSupportedPrintDocumentType(documentType) {
		return nil, fmt.Errorf("%w: 不支持的单据类型", ErrPrintTemplateInvalidInput)
	}
	if templateName == "" {
		return nil, fmt.Errorf("%w: 模板名称不能为空", ErrPrintTemplateInvalidInput)
	}
	draftLayout, err := json.Marshal(req.DraftLayout)
	if err != nil {
		return nil, fmt.Errorf("%w: 模板布局格式错误", ErrPrintTemplateInvalidInput)
	}
	if err := validatePrintLayout(draftLayout, false); err != nil {
		return nil, err
	}

	var count int64
	if err := database.DB.Model(&models.PrintTemplate{}).Where("document_type = ?", documentType).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("%w: 该单据类型已存在模板", ErrPrintTemplateConflict)
	}

	now := time.Now()
	template := models.PrintTemplate{
		TemplateID:   utils.GenerateUUID(),
		DocumentType: documentType,
		TemplateName: templateName,
		DraftLayout:  string(draftLayout),
		Status:       models.PrintTemplateStatusDraft,
		RowVersion:   1,
		CreatorID:    optionalPrintOperatorID(operatorID),
		UpdaterID:    optionalPrintOperatorID(operatorID),
		CreateDate:   &now,
		UpdateDate:   &now,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		if isPrintTemplateDuplicateError(err) {
			return nil, fmt.Errorf("%w: 该单据类型已存在模板", ErrPrintTemplateConflict)
		}
		return nil, err
	}
	return printTemplateToResponse(template)
}

func (s *PrintTemplateService) UpdatePrintTemplate(templateID string, req models.UpdatePrintTemplateRequest, operatorID string) (*models.PrintTemplateResponse, error) {
	if err := validatePrintTemplateID(templateID); err != nil {
		return nil, err
	}
	templateName := strings.TrimSpace(req.TemplateName)
	if templateName == "" {
		return nil, fmt.Errorf("%w: 模板名称不能为空", ErrPrintTemplateInvalidInput)
	}
	draftLayout, err := json.Marshal(req.DraftLayout)
	if err != nil {
		return nil, fmt.Errorf("%w: 模板布局格式错误", ErrPrintTemplateInvalidInput)
	}
	if err := validatePrintLayout(draftLayout, false); err != nil {
		return nil, err
	}

	var template models.PrintTemplate
	if err := database.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrintTemplateNotFound
		}
		return nil, err
	}
	if template.RowVersion != req.RowVersion {
		return nil, fmt.Errorf("%w: 模板已被其他人修改，请刷新后重试", ErrPrintTemplateConflict)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"template_name": templateName,
		"draft_layout":  string(draftLayout),
		"status":        models.PrintTemplateStatusDraft,
		"updater_id":    optionalPrintOperatorID(operatorID),
		"update_date":   now,
		"row_version":   req.RowVersion + 1,
	}
	result := database.DB.Model(&models.PrintTemplate{}).
		Where("template_id = ? AND row_version = ?", templateID, req.RowVersion).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: 模板已被其他人修改，请刷新后重试", ErrPrintTemplateConflict)
	}

	template.TemplateName = templateName
	template.DraftLayout = string(draftLayout)
	template.Status = models.PrintTemplateStatusDraft
	template.UpdaterID = optionalPrintOperatorID(operatorID)
	template.UpdateDate = &now
	template.RowVersion = req.RowVersion + 1
	return printTemplateToResponse(template)
}

func (s *PrintTemplateService) PublishPrintTemplate(templateID string, req models.PublishPrintTemplateRequest, operatorID string) (*models.PrintTemplateResponse, error) {
	if err := validatePrintTemplateID(templateID); err != nil {
		return nil, err
	}

	var template models.PrintTemplate
	if err := database.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrintTemplateNotFound
		}
		return nil, err
	}
	if template.RowVersion != req.RowVersion {
		return nil, fmt.Errorf("%w: 模板已被其他人修改，请刷新后重试", ErrPrintTemplateConflict)
	}
	if err := validatePrintLayout([]byte(template.DraftLayout), true); err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"published_layout": template.DraftLayout,
		"status":           models.PrintTemplateStatusPublished,
		"updater_id":       optionalPrintOperatorID(operatorID),
		"update_date":      now,
		"row_version":      req.RowVersion + 1,
	}
	result := database.DB.Model(&models.PrintTemplate{}).
		Where("template_id = ? AND row_version = ?", templateID, req.RowVersion).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: 模板已被其他人修改，请刷新后重试", ErrPrintTemplateConflict)
	}

	template.PublishedLayout = stringPointer(template.DraftLayout)
	template.Status = models.PrintTemplateStatusPublished
	template.UpdaterID = optionalPrintOperatorID(operatorID)
	template.UpdateDate = &now
	template.RowVersion = req.RowVersion + 1
	return printTemplateToResponse(template)
}

func (s *PrintTemplateService) DeletePrintTemplate(templateID string) error {
	if err := validatePrintTemplateID(templateID); err != nil {
		return err
	}
	result := database.DB.Where("template_id = ?", templateID).Delete(&models.PrintTemplate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPrintTemplateNotFound
	}
	return nil
}

func (s *PrintTemplateService) findTemplate(templateID string) (models.PrintTemplate, error) {
	if err := validatePrintTemplateID(templateID); err != nil {
		return models.PrintTemplate{}, err
	}
	var template models.PrintTemplate
	if err := database.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PrintTemplate{}, ErrPrintTemplateNotFound
		}
		return models.PrintTemplate{}, err
	}
	return template, nil
}

func validatePrintTemplateID(templateID string) error {
	if _, err := uuid.Parse(strings.TrimSpace(templateID)); err != nil {
		return fmt.Errorf("%w: 模板ID格式错误", ErrPrintTemplateInvalidInput)
	}
	return nil
}

func validatePrintLayout(layout []byte, requirePublished bool) error {
	if len(layout) == 0 || !json.Valid(layout) {
		return fmt.Errorf("%w: 模板布局必须是合法JSON", ErrPrintTemplateInvalidInput)
	}
	var payload models.PrintLayout
	if err := json.Unmarshal(layout, &payload); err != nil {
		return fmt.Errorf("%w: 模板布局格式错误", ErrPrintTemplateInvalidInput)
	}
	if payload.Version != 1 {
		return fmt.Errorf("%w: 不支持的模板布局版本", ErrPrintTemplateInvalidInput)
	}
	if payload.Page.Size != "A4" {
		return fmt.Errorf("%w: 打印纸张只能为A4", ErrPrintTemplateInvalidInput)
	}
	if payload.Page.Orientation != "portrait" && payload.Page.Orientation != "landscape" {
		return fmt.Errorf("%w: 打印方向只能为portrait或landscape", ErrPrintTemplateInvalidInput)
	}
	if err := validatePrintMargins(payload.Page.Margin); err != nil {
		return err
	}

	if requirePublished {
		if payload.Sections.Body.Table == nil || len(payload.Sections.Body.Table.Columns) == 0 {
			return fmt.Errorf("%w: 模板必须包含明细表格和至少一列", ErrPrintTemplateInvalidInput)
		}
	}

	fieldMap := printFieldDefinitionMap()
	seenIDs := make(map[string]struct{})
	for _, section := range []models.PrintSection{
		payload.Sections.PageHeader,
		payload.Sections.DocumentHeader,
		payload.Sections.DocumentFooter,
		payload.Sections.PageFooter,
	} {
		if err := validatePrintSection(section, payload.Page, fieldMap, seenIDs, requirePublished); err != nil {
			return err
		}
	}
	if payload.Sections.Body.Table != nil {
		if err := validatePrintTable(*payload.Sections.Body.Table, payload.Page, fieldMap, requirePublished); err != nil {
			return err
		}
	}
	return nil
}

func validatePrintMargins(margin models.PrintPageMargins) error {
	for _, value := range []float64{margin.Top, margin.Right, margin.Bottom, margin.Left} {
		if value < 0 || value > 40 {
			return fmt.Errorf("%w: 页边距必须在0到40mm之间", ErrPrintTemplateInvalidInput)
		}
	}
	return nil
}

func validatePrintSection(section models.PrintSection, page models.PrintPageSettings, fieldMap map[string]models.PrintFieldDefinition, seenIDs map[string]struct{}, requirePublished bool) error {
	if requirePublished && (section.Height <= 0 || section.Height > 260) {
		return fmt.Errorf("%w: 模板分区高度无效", ErrPrintTemplateInvalidInput)
	}
	for _, element := range section.Elements {
		if requirePublished {
			if element.ID == "" || element.Width <= 0 || element.Height <= 0 {
				return fmt.Errorf("%w: 模板元素尺寸或ID无效", ErrPrintTemplateInvalidInput)
			}
			if element.X < 0 || element.Y < 0 {
				return fmt.Errorf("%w: 模板元素位置无效", ErrPrintTemplateInvalidInput)
			}
			pageWidth, pageHeight := printPageSize(page)
			if element.X+element.Width > pageWidth || element.Y+element.Height > pageHeight {
				return fmt.Errorf("%w: 模板元素超出A4页面范围", ErrPrintTemplateInvalidInput)
			}
			if _, exists := seenIDs[element.ID]; exists {
				return fmt.Errorf("%w: 模板元素ID重复", ErrPrintTemplateInvalidInput)
			}
			seenIDs[element.ID] = struct{}{}
		}
		if element.Kind != "text" && element.Kind != "field" && element.Kind != "image" && element.Kind != "line" && element.Kind != "signature" {
			return fmt.Errorf("%w: 不支持的模板元素类型", ErrPrintTemplateInvalidInput)
		}
		if requirePublished && element.Kind == "field" {
			field, ok := fieldMap[element.FieldPath]
			if !ok || field.Scope == "item" {
				return fmt.Errorf("%w: 模板绑定了未注册字段 %s", ErrPrintTemplateInvalidInput, element.FieldPath)
			}
		}
	}
	return nil
}

func validatePrintTable(table models.PrintDetailTable, page models.PrintPageSettings, fieldMap map[string]models.PrintFieldDefinition, requirePublished bool) error {
	if requirePublished {
		if table.ID == "" || table.Width <= 0 || table.Height <= 0 || len(table.Columns) == 0 {
			return fmt.Errorf("%w: 明细表格尺寸或列配置无效", ErrPrintTemplateInvalidInput)
		}
		pageWidth, pageHeight := printPageSize(page)
		if table.X < 0 || table.Y < 0 || table.X+table.Width > pageWidth || table.Y+table.Height > pageHeight {
			return fmt.Errorf("%w: 明细表格超出A4页面范围", ErrPrintTemplateInvalidInput)
		}
	}
	for _, column := range table.Columns {
		if requirePublished {
			if column.ID == "" || column.Title == "" || column.Width <= 0 {
				return fmt.Errorf("%w: 明细表格列配置无效", ErrPrintTemplateInvalidInput)
			}
			field, ok := fieldMap[column.FieldPath]
			if !ok || field.Scope != "item" {
				return fmt.Errorf("%w: 明细表格绑定了未注册明细字段 %s", ErrPrintTemplateInvalidInput, column.FieldPath)
			}
			if column.Format != "text" && column.Format != "number" && column.Format != "currency" {
				return fmt.Errorf("%w: 明细表列显示格式无效", ErrPrintTemplateInvalidInput)
			}
		}
	}
	return nil
}

func printPageSize(page models.PrintPageSettings) (float64, float64) {
	if page.Orientation == "landscape" {
		return 297, 210
	}
	return 210, 297
}

func normalizePrintTemplatePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func isSupportedPrintDocumentType(documentType string) bool {
	return documentType == models.PrintDocumentTypePurchaseInbound
}

func isSupportedPrintTemplateStatus(status string) bool {
	return status == models.PrintTemplateStatusDraft || status == models.PrintTemplateStatusPublished
}

func printTemplateToListResponse(template models.PrintTemplate) models.PrintTemplateListResponse {
	return models.PrintTemplateListResponse{
		TemplateID:   template.TemplateID,
		DocumentType: template.DocumentType,
		TemplateName: template.TemplateName,
		Status:       template.Status,
		RowVersion:   template.RowVersion,
		HasDraft:     strings.TrimSpace(template.DraftLayout) != "",
		HasPublished: template.PublishedLayout != nil && strings.TrimSpace(*template.PublishedLayout) != "",
		CreateDate:   models.TimeToStringPtr(template.CreateDate),
		UpdateDate:   models.TimeToStringPtr(template.UpdateDate),
	}
}

func printTemplateToResponse(template models.PrintTemplate) (*models.PrintTemplateResponse, error) {
	var draftLayout models.PrintLayout
	if err := json.Unmarshal([]byte(template.DraftLayout), &draftLayout); err != nil {
		return nil, fmt.Errorf("打印模板草稿布局数据损坏: %w", err)
	}
	response := &models.PrintTemplateResponse{
		TemplateID:   template.TemplateID,
		DocumentType: template.DocumentType,
		TemplateName: template.TemplateName,
		Status:       template.Status,
		DraftLayout:  draftLayout,
		RowVersion:   template.RowVersion,
		CreateDate:   models.TimeToStringPtr(template.CreateDate),
		UpdateDate:   models.TimeToStringPtr(template.UpdateDate),
	}
	if template.PublishedLayout != nil {
		var publishedLayout models.PrintLayout
		if err := json.Unmarshal([]byte(*template.PublishedLayout), &publishedLayout); err != nil {
			return nil, fmt.Errorf("打印模板已发布布局数据损坏: %w", err)
		}
		response.PublishedLayout = &publishedLayout
	}
	return response, nil
}

func optionalPrintOperatorID(operatorID string) *string {
	if _, err := uuid.Parse(strings.TrimSpace(operatorID)); err != nil {
		return nil
	}
	value := strings.TrimSpace(operatorID)
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func isPrintTemplateDuplicateError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique")
}

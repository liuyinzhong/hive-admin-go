package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
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

// printSupportedPaperSizes 与 worm-vue3-print PaperSize 枚举保持一致。
var printSupportedPaperSizes = map[string]struct{}{
	"A4": {}, "A3": {}, "A5": {}, "Letter": {}, "Legal": {}, "CUSTOM": {},
}

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
	draftLayout := []byte(req.DraftLayout)
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
	draftLayout := []byte(req.DraftLayout)
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

// worm-vue3-print TemplateData 边界校验。
// Go 侧不复刻渲染管线，只校验数据边界：页面结构合法、字段表达式全部命中注册表、
// 表达式标识符全部来自渲染引擎白名单，渲染级排版检查由设计器与浏览器预览负责。

// printTemplatePayload 只提取 worm TemplateData 中后端关心的结构性字段，
// 其余内容（元素样式、水印样式等）原样存储，不做结构化解释。
type printTemplatePayload struct {
	PaperSize    string           `json:"paperSize"`
	Orientation  string           `json:"orientation"`
	Unit         string           `json:"unit"`
	CustomWidth  float64          `json:"customWidth"`
	CustomHeight float64          `json:"customHeight"`
	Margins      printPageMargins `json:"margins"`
	Header       struct {
		Elements []json.RawMessage `json:"elements"`
	} `json:"header"`
	Footer struct {
		Elements []json.RawMessage `json:"elements"`
	} `json:"footer"`
	FirstPageOverlay struct {
		Elements []json.RawMessage `json:"elements"`
	} `json:"firstPageOverlay"`
	Elements  []json.RawMessage `json:"elements"`
	Watermark json.RawMessage   `json:"watermark"`
}

// printPageMargins 是 worm TemplateData 的页边距结构（mm）。
type printPageMargins struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// printMultiPagePayload 是 worm 1.3.0 起的多页面模板 wrapper（{ version, pages }）。
// pages 用指针区分「字段缺省」（单页 TemplateData）与「显式空数组」（非法）。
// 与 core 的 normalizeTemplate 判据一致：存在 pages 数组即按多页面模板处理。
type printMultiPagePayload struct {
	Version int                `json:"version"`
	Pages   *[]json.RawMessage `json:"pages"`
}

// printPageGeometry 是一页的纸张几何，用于多页面模板的一致性校验（core 要求各页一致）。
type printPageGeometry struct {
	PaperSize    string
	Orientation  string
	CustomWidth  float64
	CustomHeight float64
}

// printExpressionFunctions 是 worm 渲染管线注册的表达式函数白名单
// （见 worm-vue3-print print-core src/render/expression-eval.ts）。
var printExpressionFunctions = map[string]struct{}{
	"MONEY": {}, "DATE": {}, "UPPER": {}, "IF": {},
	"CONCAT": {}, "IFEMPTY": {}, "ROUND": {}, "LEN": {},
	"SUM": {}, "AVG": {}, "COUNT": {}, "MIN": {}, "MAX": {},
}

// printExpressionGlobals 是 worm 表达式求值器暴露的安全全局标识
// （见 worm-vue3-print print-core src/evaluator.ts SAFE_GLOBALS）。
var printExpressionGlobals = map[string]struct{}{
	"Math": {}, "Number": {}, "String": {}, "Boolean": {},
	"parseInt": {}, "parseFloat": {}, "isNaN": {},
	"true": {}, "false": {}, "null": {}, "undefined": {},
}

// printExpressionSystemVars 是渲染端注入的系统变量（页码/打印时间）。
var printExpressionSystemVars = map[string]struct{}{
	"pageIndex": {}, "totalPages": {}, "printDate": {}, "printTime": {},
}

func validatePrintLayout(layout []byte, requirePublished bool) error {
	if len(layout) == 0 || !json.Valid(layout) {
		return fmt.Errorf("%w: 模板布局必须是合法JSON", ErrPrintTemplateInvalidInput)
	}
	var wrapper printMultiPagePayload
	if err := json.Unmarshal(layout, &wrapper); err != nil {
		return fmt.Errorf("%w: 模板布局必须是对象结构", ErrPrintTemplateInvalidInput)
	}
	if wrapper.Pages == nil {
		if _, elements, err := validatePrintPage(layout, requirePublished); err != nil {
			return err
		} else if requirePublished && elements == 0 {
			return fmt.Errorf("%w: 模板必须包含至少一个元素", ErrPrintTemplateInvalidInput)
		}
		return nil
	}

	// 多页面模板：逐页校验，并校验各页纸张尺寸（含方向）一致——渲染端一次出图共用一份纸型。
	pages := *wrapper.Pages
	if len(pages) == 0 {
		return fmt.Errorf("%w: 多页面模板至少包含一页", ErrPrintTemplateInvalidInput)
	}
	totalElements := 0
	firstGeometry := printPageGeometry{}
	for index, page := range pages {
		geometry, elements, err := validatePrintPage(page, requirePublished)
		if err != nil {
			return err
		}
		if index == 0 {
			firstGeometry = geometry
		} else if geometry != firstGeometry {
			return fmt.Errorf("%w: 多页面模板各页纸张尺寸与方向必须一致（第 %d 页）", ErrPrintTemplateInvalidInput, index+1)
		}
		totalElements += elements
	}
	if requirePublished && totalElements == 0 {
		return fmt.Errorf("%w: 模板必须包含至少一个元素", ErrPrintTemplateInvalidInput)
	}
	return nil
}

// validatePrintPage 校验单页 TemplateData，返回该页纸张几何与元素数量。
// 「至少一个元素」交由调用方按整份模板判定：多页面模板允许某页无元素，但整份不能全空。
// 元素 ID 唯一性按页判定——各页独立渲染，跨页 ID 不冲突。
func validatePrintPage(layout []byte, requirePublished bool) (printPageGeometry, int, error) {
	geometry := printPageGeometry{}
	var payload printTemplatePayload
	if err := json.Unmarshal(layout, &payload); err != nil {
		return geometry, 0, fmt.Errorf("%w: 模板布局必须是对象结构", ErrPrintTemplateInvalidInput)
	}
	if _, ok := printSupportedPaperSizes[payload.PaperSize]; !ok {
		return geometry, 0, fmt.Errorf("%w: 不支持的打印纸张规格", ErrPrintTemplateInvalidInput)
	}
	if payload.PaperSize == "CUSTOM" && (payload.CustomWidth <= 0 || payload.CustomHeight <= 0) {
		return geometry, 0, fmt.Errorf("%w: 自定义纸张必须提供有效宽高", ErrPrintTemplateInvalidInput)
	}
	if payload.Orientation != "portrait" && payload.Orientation != "landscape" {
		return geometry, 0, fmt.Errorf("%w: 打印方向只能为portrait或landscape", ErrPrintTemplateInvalidInput)
	}
	if payload.Unit != "" && payload.Unit != "mm" {
		return geometry, 0, fmt.Errorf("%w: 模板坐标单位只支持mm", ErrPrintTemplateInvalidInput)
	}
	if err := validatePrintMargins(payload.Margins); err != nil {
		return geometry, 0, err
	}
	geometry = printPageGeometry{
		PaperSize:    payload.PaperSize,
		Orientation:  payload.Orientation,
		CustomWidth:  payload.CustomWidth,
		CustomHeight: payload.CustomHeight,
	}

	elements := make([]json.RawMessage, 0, len(payload.Elements)+
		len(payload.Header.Elements)+len(payload.Footer.Elements)+len(payload.FirstPageOverlay.Elements))
	elements = append(elements, payload.Elements...)
	elements = append(elements, payload.Header.Elements...)
	elements = append(elements, payload.Footer.Elements...)
	elements = append(elements, payload.FirstPageOverlay.Elements...)

	fieldMap := printFieldDefinitionMap()
	seenIDs := make(map[string]struct{})
	for _, raw := range elements {
		if err := validatePrintElement(raw, fieldMap, seenIDs, requirePublished); err != nil {
			return geometry, 0, err
		}
	}
	if len(payload.Watermark) > 0 && !json.Valid(payload.Watermark) {
		return geometry, 0, fmt.Errorf("%w: 水印配置必须是合法JSON", ErrPrintTemplateInvalidInput)
	}
	if err := validatePrintWatermark(payload.Watermark, fieldMap); err != nil {
		return geometry, 0, err
	}
	return geometry, len(elements), nil
}

func validatePrintElement(raw json.RawMessage, fieldMap map[string]models.PrintFieldDefinition, seenIDs map[string]struct{}, requirePublished bool) error {
	var element struct {
		ID          string                 `json:"id"`
		Type        string                 `json:"type"`
		Options     map[string]interface{} `json:"options"`
		ElementType struct {
			Type string `json:"type"`
		} `json:"printElementType"`
	}
	if err := json.Unmarshal(raw, &element); err != nil {
		return fmt.Errorf("%w: 模板元素必须是对象结构", ErrPrintTemplateInvalidInput)
	}
	if requirePublished {
		if element.ID == "" {
			return fmt.Errorf("%w: 模板元素缺少ID", ErrPrintTemplateInvalidInput)
		}
		if _, exists := seenIDs[element.ID]; exists {
			return fmt.Errorf("%w: 模板元素ID重复: %s", ErrPrintTemplateInvalidInput, element.ID)
		}
		seenIDs[element.ID] = struct{}{}
	}
	// 字段边界：只扫描渲染引擎会求值的表达式载体（元素 formatter、表格单元格 formatter），
	// 不扫 HTML 内容、测试值等非求值字段，避免内联 CSS 花括号被误判为表达式。
	for _, text := range printExpressionTexts(element.Options) {
		if err := validatePrintExpressionText(text, fieldMap); err != nil {
			return err
		}
	}
	// 明细表格：发布时数据源必须是注册的明细集合根
	if isPrintTableElement(element.Type, element.ElementType.Type, element.Options) && requirePublished {
		dataSource, _ := element.Options["dataSource"].(string)
		if strings.TrimSpace(dataSource) == "" {
			return fmt.Errorf("%w: 明细表格未配置数据源", ErrPrintTemplateInvalidInput)
		}
		field, ok := fieldMap[strings.TrimSpace(dataSource)]
		if !ok || field.DataType != "list" {
			return fmt.Errorf("%w: 明细表格数据源 %s 不是注册的明细集合", ErrPrintTemplateInvalidInput, dataSource)
		}
	}
	return nil
}

// printExpressionTexts 收集元素 options 中参与表达式求值的文本：
// 元素级 formatter 和表格行单元格 formatter。
func printExpressionTexts(options map[string]interface{}) []string {
	texts := []string{}
	if formatter, ok := options["formatter"].(string); ok {
		texts = append(texts, formatter)
	}
	rows, ok := options["tableRows"].([]interface{})
	if !ok {
		return texts
	}
	for _, row := range rows {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		cells, ok := rowMap["cells"].([]interface{})
		if !ok {
			continue
		}
		for _, cell := range cells {
			cellMap, ok := cell.(map[string]interface{})
			if !ok {
				continue
			}
			if formatter, ok := cellMap["formatter"].(string); ok {
				texts = append(texts, formatter)
			}
		}
	}
	return texts
}

// isPrintTableElement 依据顶层 type、printElementType.type 或 tableRows 判断表格元素；
// worm 序列化时顶层 type 可选，三种来源任一命中即视为表格。
func isPrintTableElement(typeField, metaType string, options map[string]interface{}) bool {
	if typeField == "table" || metaType == "table" {
		return true
	}
	return hasPrintTableRows(options)
}

func validatePrintWatermark(raw json.RawMessage, fieldMap map[string]models.PrintFieldDefinition) error {
	if len(raw) == 0 {
		return nil
	}
	var watermark map[string]interface{}
	if err := json.Unmarshal(raw, &watermark); err != nil {
		return fmt.Errorf("%w: 水印配置必须是对象结构", ErrPrintTemplateInvalidInput)
	}
	if content, ok := watermark["content"].(string); ok {
		if err := validatePrintExpressionText(content, fieldMap); err != nil {
			return err
		}
	}
	if binding, _ := watermark["binding"].(string); strings.TrimSpace(binding) != "" {
		if _, ok := fieldMap[strings.TrimSpace(binding)]; !ok {
			return fmt.Errorf("%w: 水印绑定了未注册字段 %s", ErrPrintTemplateInvalidInput, binding)
		}
	}
	return nil
}

func validatePrintMargins(margin printPageMargins) error {
	for _, value := range []float64{margin.Top, margin.Right, margin.Bottom, margin.Left} {
		if value < 0 || value > 40 {
			return fmt.Errorf("%w: 页边距必须在0到40mm之间", ErrPrintTemplateInvalidInput)
		}
	}
	return nil
}

// validatePrintExpressionText 校验一段模板文本中的 {表达式}：
// 点分路径必须命中字段注册表，其余标识符必须命中函数/全局/系统变量白名单。
func validatePrintExpressionText(text string, fieldMap map[string]models.PrintFieldDefinition) error {
	for _, expr := range extractPrintExpressions(text) {
		identifiers, err := extractPrintIdentifiers(expr)
		if err != nil {
			return fmt.Errorf("%w: 表达式 {%s} 解析失败", ErrPrintTemplateInvalidInput, expr)
		}
		for _, identifier := range identifiers {
			if strings.Contains(identifier, ".") {
				if _, ok := fieldMap[identifier]; !ok {
					return fmt.Errorf("%w: 模板绑定了未注册字段 %s", ErrPrintTemplateInvalidInput, identifier)
				}
				continue
			}
			if _, ok := printExpressionFunctions[identifier]; ok {
				continue
			}
			if _, ok := printExpressionGlobals[identifier]; ok {
				continue
			}
			if _, ok := printExpressionSystemVars[identifier]; ok {
				continue
			}
			return fmt.Errorf("%w: 表达式使用了未授权的标识符 %s", ErrPrintTemplateInvalidInput, identifier)
		}
	}
	return nil
}

// extractPrintExpressions 抽取文本中成对花括号包裹的表达式（支持嵌套）。
func extractPrintExpressions(text string) []string {
	expressions := []string{}
	depth := 0
	start := -1
	for i, r := range text {
		switch r {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && start >= 0 {
				if expr := strings.TrimSpace(text[start+1 : i]); expr != "" {
					expressions = append(expressions, expr)
				}
				start = -1
			}
		}
	}
	return expressions
}

// extractPrintIdentifiers 抽取表达式中的标识符；先剔除字符串字面量，
// 避免 DATE(header.inboundDate,'YYYY-MM-DD') 中的格式串被误判为字段。
func extractPrintIdentifiers(expr string) ([]string, error) {
	stripped := strings.Builder{}
	for i := 0; i < len(expr); {
		r := expr[i]
		if r == '\'' || r == '"' {
			quote := r
			i++
			for i < len(expr) && expr[i] != quote {
				if expr[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(expr) {
				return nil, errors.New("字符串字面量未闭合")
			}
			i++
			continue
		}
		stripped.WriteByte(r)
		i++
	}
	matches := printIdentifierPattern.FindAllString(stripped.String(), -1)
	return matches, nil
}

// printIdentifierPattern 提取标识符；点号属于标识符的一部分，
// 使 header.inboundNo 整体命中注册表路径而非被拆成两段裸标识符。
var printIdentifierPattern = regexp.MustCompile(`[A-Za-z_$][A-Za-z0-9_$.]*`)

// hasPrintTableRows 判断元素 options 是否带有表格行（兼容顶层 type 缺失的序列化形态）。
func hasPrintTableRows(options map[string]interface{}) bool {
	rows, exists := options["tableRows"]
	if !exists || rows == nil {
		return false
	}
	_, isArray := rows.([]interface{})
	return isArray
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
	draftLayout := json.RawMessage(template.DraftLayout)
	if !json.Valid(draftLayout) {
		return nil, errors.New("打印模板草稿布局数据损坏")
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
		publishedLayout := json.RawMessage(*template.PublishedLayout)
		if !json.Valid(publishedLayout) {
			return nil, errors.New("打印模板已发布布局数据损坏")
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

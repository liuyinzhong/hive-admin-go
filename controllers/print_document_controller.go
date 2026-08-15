package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type PrintDocumentController struct {
	printDocumentService *services.PrintDocumentService
	printTemplateService *services.PrintTemplateService
}

func NewPrintDocumentController() *PrintDocumentController {
	return &PrintDocumentController{
		printDocumentService: services.NewPrintDocumentService(),
		printTemplateService: services.NewPrintTemplateService(),
	}
}

// GetPurchaseInboundPrintDocument 获取采购入库单打印数据和当前已发布模板
// @Summary 获取采购入库单打印数据
// @Description 校验采购入库单当前数据范围后，返回统一打印数据协议和已发布模板；动态字段不保存打印快照
// @Tags 打印管理/打印数据
// @Produce json
// @Security ApiKeyAuth
// @Param inboundId path string true "采购入库单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.PrintDocumentBundleResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购入库单不存在"
// @Failure 409 {object} models.Response "未配置可用打印模板"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /printDocuments/purchaseInbound/{inboundId} [get]
func (ctrl *PrintDocumentController) GetPurchaseInboundPrintDocument(c *gin.Context) {
	template, err := ctrl.printTemplateService.GetPublishedPrintTemplate(models.PrintDocumentTypePurchaseInbound)
	if err != nil {
		writePrintDocumentError(c, err)
		return
	}
	data, err := ctrl.printDocumentService.GetPrintDocument(models.PrintDocumentTypePurchaseInbound, c.Param("inboundId"), currentDataPermission(c))
	if err != nil {
		writePrintDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(models.PrintDocumentBundleResponse{
		Template: template,
		Data:     data,
	}))
}

// GetPurchaseInboundPrintData 获取采购入库单打印预览数据
// @Summary 获取采购入库单打印预览数据
// @Description 校验采购入库单当前数据范围后返回真实业务单据的统一打印数据，供模板设计器实时预览
// @Tags 打印管理/打印数据
// @Produce json
// @Security ApiKeyAuth
// @Param inboundId path string true "采购入库单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.PrintDocumentResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购入库单不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /printDocuments/purchaseInbound/{inboundId}/data [get]
func (ctrl *PrintDocumentController) GetPurchaseInboundPrintData(c *gin.Context) {
	data, err := ctrl.printDocumentService.GetPrintDocument(models.PrintDocumentTypePurchaseInbound, c.Param("inboundId"), currentDataPermission(c))
	if err != nil {
		writePrintDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(data))
}

func writePrintDocumentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrPrintTemplateInvalidInput):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrPrintTemplateUnavailable):
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrErpPurchaseInboundInvalidInput):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrErpPurchaseInboundNotFound):
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, err.Error()))
	default:
		log.Printf("打印数据获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "打印数据获取失败"))
	}
}

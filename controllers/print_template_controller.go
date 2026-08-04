package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type PrintTemplateController struct {
	printTemplateService *services.PrintTemplateService
}

func NewPrintTemplateController() *PrintTemplateController {
	return &PrintTemplateController{
		printTemplateService: services.NewPrintTemplateService(),
	}
}

// GetPrintTemplateList 获取打印模板列表
// @Summary 获取打印模板列表
// @Description 查询打印管理中的模板；一个单据类型只允许一条逻辑模板记录
// @Tags 打印管理/打印模板
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，最大100"
// @Param documentType query string false "单据类型"
// @Param status query string false "模板状态"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.PrintTemplateListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /printTemplates [get]
func (ctrl *PrintTemplateController) GetPrintTemplateList(c *gin.Context) {
	var req models.PrintTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.printTemplateService.GetPrintTemplateList(req)
	if err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPrintTemplateMetadata 获取打印字段注册表
// @Summary 获取打印字段注册表
// @Description 返回可绑定的系统定义单据类型和打印字段，设计器不得使用未注册字段
// @Tags 打印管理/打印模板
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.PrintTemplateMetadataResponse} "获取成功"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Router /printTemplates/metadata [get]
func (ctrl *PrintTemplateController) GetPrintTemplateMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewSuccessResponse(ctrl.printTemplateService.GetPrintTemplateMetadata()))
}

// CreatePrintTemplate 创建打印模板
// @Summary 创建打印模板
// @Description 创建指定单据类型的唯一打印模板，初始保存为草稿
// @Tags 打印管理/打印模板
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreatePrintTemplateRequest true "打印模板"
// @Success 200 {object} models.Response{data=models.PrintTemplateResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 409 {object} models.Response "模板已存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /printTemplates [post]
func (ctrl *PrintTemplateController) CreatePrintTemplate(c *gin.Context) {
	var req models.CreatePrintTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.printTemplateService.CreatePrintTemplate(req, c.GetString("userId"))
	if err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPrintTemplateDetail 获取打印模板详情
// @Summary 获取打印模板详情
// @Description 获取当前草稿和当前已发布内容，不返回历史版本
// @Tags 打印管理/打印模板
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "模板ID"
// @Success 200 {object} models.Response{data=models.PrintTemplateResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "模板不存在"
// @Router /printTemplates/{templateId} [get]
func (ctrl *PrintTemplateController) GetPrintTemplateDetail(c *gin.Context) {
	result, err := ctrl.printTemplateService.GetPrintTemplateDetail(c.Param("templateId"))
	if err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdatePrintTemplate 更新打印模板草稿
// @Summary 更新打印模板草稿
// @Description 只更新当前草稿；已发布内容保持不变，发布前草稿不影响业务打印
// @Tags 打印管理/打印模板
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "模板ID"
// @Param request body models.UpdatePrintTemplateRequest true "打印模板草稿"
// @Success 200 {object} models.Response{data=models.PrintTemplateResponse} "保存成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "模板不存在"
// @Failure 409 {object} models.Response "模板已被修改"
// @Router /printTemplates/{templateId} [put]
func (ctrl *PrintTemplateController) UpdatePrintTemplate(c *gin.Context) {
	var req models.UpdatePrintTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.printTemplateService.UpdatePrintTemplate(c.Param("templateId"), req, c.GetString("userId"))
	if err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// PublishPrintTemplate 发布打印模板
// @Summary 发布打印模板
// @Description 校验当前草稿后覆盖当前已发布内容；不生成历史版本
// @Tags 打印管理/打印模板
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "模板ID"
// @Param request body models.PublishPrintTemplateRequest true "模板版本"
// @Success 200 {object} models.Response{data=models.PrintTemplateResponse} "发布成功"
// @Failure 400 {object} models.Response "参数错误或布局校验失败"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "模板不存在"
// @Failure 409 {object} models.Response "模板已被修改"
// @Router /printTemplates/{templateId}/publish [post]
func (ctrl *PrintTemplateController) PublishPrintTemplate(c *gin.Context) {
	var req models.PublishPrintTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.printTemplateService.PublishPrintTemplate(c.Param("templateId"), req, c.GetString("userId"))
	if err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeletePrintTemplate 删除打印模板
// @Summary 删除打印模板
// @Description 删除当前单据类型的逻辑模板；删除后业务打印将明确提示未配置模板
// @Tags 打印管理/打印模板
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "模板ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "模板不存在"
// @Router /printTemplates/{templateId} [delete]
func (ctrl *PrintTemplateController) DeletePrintTemplate(c *gin.Context) {
	if err := ctrl.printTemplateService.DeletePrintTemplate(c.Param("templateId")); err != nil {
		writePrintTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func writePrintTemplateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrPrintTemplateInvalidInput):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrPrintTemplateNotFound):
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrPrintTemplateConflict):
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrPrintTemplateUnavailable):
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
	default:
		log.Printf("打印模板操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "打印模板操作失败"))
	}
}

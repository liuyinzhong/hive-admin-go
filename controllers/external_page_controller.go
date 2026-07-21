package controllers

import (
	"errors"
	"net/http"

	"hive-admin-go/models"
	"hive-admin-go/services"

	"github.com/gin-gonic/gin"
)

type ExternalPageController struct {
	service *services.ExternalPageService
}

func NewExternalPageController() *ExternalPageController {
	return &ExternalPageController{service: services.NewExternalPageService()}
}

// GetExternalPages 获取外部页面列表。
// @Summary 获取外部页面列表
// @Tags 系统管理/外部页面
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param title query string false "管理名称"
// @Param name query string false "路由名称"
// @Param path query string false "路由地址"
// @Param status query int false "状态"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ExternalPageResponse}}
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Router /system/externalPages [get]
func (ctrl *ExternalPageController) GetExternalPages(c *gin.Context) {
	var req models.ExternalPageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.service.GetExternalPages(req)
	if err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateExternalPage 创建外部页面。
// @Summary 创建外部页面
// @Tags 系统管理/外部页面
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateExternalPageRequest true "外部页面信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Router /system/externalPages [post]
func (ctrl *ExternalPageController) CreateExternalPage(c *gin.Context) {
	var req models.CreateExternalPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.service.CreateExternalPage(req, c.GetString("userId")); err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetExternalPage 获取外部页面详情。
// @Summary 获取外部页面详情
// @Tags 系统管理/外部页面
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "外部页面ID"
// @Success 200 {object} models.Response{data=models.ExternalPageResponse}
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /system/externalPages/{id} [get]
func (ctrl *ExternalPageController) GetExternalPage(c *gin.Context) {
	result, err := ctrl.service.GetExternalPage(c.Param("id"))
	if err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateExternalPage 修改外部页面。
// @Summary 修改外部页面
// @Tags 系统管理/外部页面
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "外部页面ID"
// @Param request body models.UpdateExternalPageRequest true "外部页面信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /system/externalPages/{id} [put]
func (ctrl *ExternalPageController) UpdateExternalPage(c *gin.Context) {
	var req models.UpdateExternalPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.service.UpdateExternalPage(c.Param("id"), req); err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateExternalPageStatus 修改外部页面状态。
// @Summary 修改外部页面状态
// @Tags 系统管理/外部页面
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "外部页面ID"
// @Param request body models.UpdateExternalPageStatusRequest true "状态"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /system/externalPages/{id}/status [put]
func (ctrl *ExternalPageController) UpdateExternalPageStatus(c *gin.Context) {
	var req models.UpdateExternalPageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.service.UpdateExternalPageStatus(c.Param("id"), req.Status); err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteExternalPages 删除外部页面。
// @Summary 删除外部页面
// @Tags 系统管理/外部页面
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.DeleteExternalPagesRequest true "外部页面ID列表"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /system/externalPages [delete]
func (ctrl *ExternalPageController) DeleteExternalPages(c *gin.Context) {
	var req models.DeleteExternalPagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.service.DeleteExternalPages(req.IDs); err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetPublicExternalPage 匿名检查外部页面是否可访问。
// @Summary 检查外部页面是否可访问
// @Tags 公共接口/外部页面
// @Produce json
// @Param name path string true "路由名称"
// @Success 200 {object} models.Response{data=models.PublicExternalPageResponse}
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /public/externalPages/{name} [get]
func (ctrl *ExternalPageController) GetPublicExternalPage(c *gin.Context) {
	result, err := ctrl.service.GetPublicExternalPage(c.Param("name"))
	if err != nil {
		writeExternalPageError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeExternalPageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrExternalPageNotFound):
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, "外部页面不存在"))
	case errors.Is(err, services.ErrExternalPageInvalid),
		errors.Is(err, services.ErrRouteNameConflict),
		errors.Is(err, services.ErrRoutePathConflict):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "外部页面操作失败"))
	}
}

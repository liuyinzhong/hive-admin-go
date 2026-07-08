package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetDictTree 获取字典树
// @Summary 获取字典树
// @Description 获取字典树结构
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sorts query string false "排序参数 排序时仅支持：label、type、value"
// @Success 200 {object} models.Response{data=[]models.DictTreeResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts [get]
func (ctrl *SystemController) GetDictTree(c *gin.Context) {
	var req models.DictListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.dictService.GetDictTree(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateDict 创建字典
// @Summary 创建字典
// @Description 创建新字典
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateDictRequest true "字典信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts [post]
func (ctrl *SystemController) CreateDict(c *gin.Context) {
	var req models.CreateDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.dictService.CreateDict(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetDictDetail 获取字典详情
// @Summary 获取字典详情
// @Description 根据字典ID获取字典详情
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "字典ID"
// @Success 200 {object} models.Response{data=models.DictTreeResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts/{id} [get]
func (ctrl *SystemController) GetDictDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "字典ID不能为空"))
		return
	}

	result, err := ctrl.dictService.GetDictDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateDict 更新字典
// @Summary 更新字典
// @Description 更新字典信息
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "字典ID"
// @Param request body models.UpdateDictRequest true "字典信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts/{id} [put]
func (ctrl *SystemController) UpdateDict(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "字典ID不能为空"))
		return
	}

	var req models.UpdateDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.dictService.UpdateDict(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateDictStatus 更新字典状态
// @Summary 更新字典状态
// @Description 更新字典启用/禁用状态
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "字典ID"
// @Param request body map[string]int true "状态"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts/{id}/status [put]
func (ctrl *SystemController) UpdateDictStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "字典ID不能为空"))
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.dictService.UpdateDictStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteDicts 删除字典
// @Summary 删除字典
// @Description 批量删除字典
// @Tags 系统管理-字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "字典ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/dicts [delete]
func (ctrl *SystemController) DeleteDicts(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.dictService.DeleteDicts(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

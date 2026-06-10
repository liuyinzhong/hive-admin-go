package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetModules 获取模块列表
// @Summary 获取模块列表
// @Description 获取所有模块（不分页），可根据项目ID筛选
// @Tags 开发管理-模块管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId query string false "项目ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/modules [get]
func (dc *DevController) GetModules(c *gin.Context) {
	params := make(map[string]interface{})

	if projectID := c.Query("projectId"); projectID != "" {
		params["projectId"] = projectID
	}

	modules, err := services.GetAllModules(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(modules))
}

// GetModule 获取模块详情
// @Summary 获取模块详情
// @Description 根据模块ID获取模块详情
// @Tags 开发管理-模块管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param moduleId path string true "模块ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/modules/{moduleId} [get]
func (dc *DevController) GetModule(c *gin.Context) {
	moduleID := c.Param("moduleId")
	module, err := services.GetModuleByID(moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(module))
}

// CreateModule 创建模块
// @Summary 创建模块
// @Description 创建新模块
// @Tags 开发管理-模块管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateModuleRequest true "模块信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/modules [post]
func (dc *DevController) CreateModule(c *gin.Context) {
	var req models.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateModule(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateModule 更新模块
// @Summary 更新模块
// @Description 更新模块信息
// @Tags 开发管理-模块管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param moduleId path string true "模块ID"
// @Param request body models.UpdateModuleRequest true "模块信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/modules/{moduleId} [put]
func (dc *DevController) UpdateModule(c *gin.Context) {
	moduleID := c.Param("moduleId")

	var req models.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.UpdateModule(moduleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteModules 删除模块
// @Summary 删除模块
// @Description 批量删除模块
// @Tags 开发管理-模块管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "模块ID列表"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/modules [delete]
func (dc *DevController) DeleteModules(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.DeleteModules(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetVersions 获取版本列表
// @Summary 获取版本列表
// @Description 分页获取版本列表
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param version query string false "版本号"
// @Param projectId query string false "项目ID"
// @Param releaseStatus query int false "发布状态"
// @Param sorts query string false "排序参数"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.VersionResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions [get]
func (dc *DevController) GetVersions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	releaseStatus := -1
	if rsStr := c.Query("releaseStatus"); rsStr != "" {
		releaseStatus, _ = strconv.Atoi(rsStr)
	}

	params := map[string]interface{}{
		"version":       c.Query("version"),
		"projectId":     c.Query("projectId"),
		"releaseStatus": releaseStatus,
		"sorts":         c.Query("sorts"),
	}

	result, err := services.GetVersions(page, pageSize, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllVersions 获取所有版本
// @Summary 获取所有版本
// @Description 获取所有版本（不分页）
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param version query string false "版本号"
// @Param projectId query string false "项目ID"
// @Param releaseStatus query int false "发布状态"
// @Success 200 {object} models.Response{data=[]models.VersionResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/versions/all [get]
func (dc *DevController) GetAllVersions(c *gin.Context) {
	releaseStatus := -1
	if rsStr := c.Query("releaseStatus"); rsStr != "" {
		releaseStatus, _ = strconv.Atoi(rsStr)
	}

	params := map[string]interface{}{
		"version":       c.Query("version"),
		"projectId":     c.Query("projectId"),
		"releaseStatus": releaseStatus,
	}

	versions, err := services.GetAllVersions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(versions))
}

// GetVersion 获取版本详情
// @Summary 获取版本详情
// @Description 根据版本ID获取版本详情
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param versionId path string true "版本ID"
// @Success 200 {object} models.Response{data=models.VersionResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions/{versionId} [get]
func (dc *DevController) GetVersion(c *gin.Context) {
	versionID := c.Param("versionId")
	version, err := services.GetVersionByID(versionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(version))
}

// GetLatestVersion 获取最新版本号
// @Summary 获取最新版本号
// @Description 获取指定项目下按版本号排序后的最大版本号
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId query string true "项目ID"
// @Success 200 {object} models.Response{data=models.VersionResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions/getLastVersion [get]
func (dc *DevController) GetLatestVersion(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "projectId不能为空"))
		return
	}

	version, err := services.GetLatestVersion(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(version))
}

// CreateVersion 创建版本
// @Summary 创建版本
// @Description 创建新版本
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateVersionRequest true "版本信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions [post]
func (dc *DevController) CreateVersion(c *gin.Context) {
	var req models.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数格式错误: "+err.Error()))
		return
	}

	if req.Version == nil || *req.Version == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "版本号不能为空"))
		return
	}
	if req.ProjectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "项目ID不能为空"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateVersion(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateVersion 更新版本
// @Summary 更新版本
// @Description 更新版本信息
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param versionId path string true "版本ID"
// @Param request body models.UpdateVersionRequest true "版本信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions/{versionId} [put]
func (dc *DevController) UpdateVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	var req models.UpdateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数格式错误: "+err.Error()))
		return
	}

	if req.Version == nil || *req.Version == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "版本号不能为空"))
		return
	}
	if req.ProjectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "项目ID不能为空"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateVersion(versionID, &req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateVersionNext 版本流转状态
// @Summary 版本流转状态
// @Description 更新版本状态并记录变更历史
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param versionId path string true "版本ID"
// @Param request body models.UpdateVersionNextRequest true "版本状态更新信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions/{versionId}/next [put]
func (dc *DevController) UpdateVersionNext(c *gin.Context) {
	versionID := c.Param("versionId")

	var req models.UpdateVersionNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateVersionNext(versionID, req.ReleaseStatus, req.ChangeRichText, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteVersions 删除版本
// @Summary 删除版本
// @Description 批量删除版本
// @Tags 开发管理/版本管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "版本ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/versions [delete]
func (dc *DevController) DeleteVersions(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.DeleteVersions(ids, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetProjects 获取项目列表
// @Summary 获取项目列表
// @Description 获取所有项目（不分页）
// @Tags 开发管理-项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.ProjectResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/projects [get]
func (dc *DevController) GetProjects(c *gin.Context) {
	projects, err := services.GetAllProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(projects))
}

// GetProject 获取项目详情
// @Summary 获取项目详情
// @Description 根据项目ID获取项目详情
// @Tags 开发管理-项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string true "项目ID"
// @Success 200 {object} models.Response{data=models.ProjectResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/projects/{projectId} [get]
func (dc *DevController) GetProject(c *gin.Context) {
	projectID := c.Param("projectId")
	project, err := services.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(project))
}

// CreateProject 创建项目
// @Summary 创建项目
// @Description 创建新项目
// @Tags 开发管理-项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateProjectRequest true "项目信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/projects [post]
func (dc *DevController) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateProject(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateProject 更新项目
// @Summary 更新项目
// @Description 更新项目信息
// @Tags 开发管理-项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string true "项目ID"
// @Param request body models.UpdateProjectRequest true "项目信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/projects/{projectId} [put]
func (dc *DevController) UpdateProject(c *gin.Context) {
	projectID := c.Param("projectId")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.UpdateProject(projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

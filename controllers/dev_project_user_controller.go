package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetProjectUsers 获取项目用户列表
// @Summary 获取项目用户列表
// @Description 数据权限：公开接口,登录即可查看;返回项目成员,不做记录级数据权限过滤
// @Tags 开发管理/项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId query string true "项目ID"
// @Success 200 {object} models.Response{data=[]models.ProjectUserResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/project-users [get]
func (dc *DevController) GetProjectUsers(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "projectId不能为空"))
		return
	}

	users, err := services.GetProjectUsers(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(users))
}

// SaveProjectUsers 全量保存项目用户
// @Summary 全量保存项目用户
// @Description 数据权限：需要 dev:project:user 权限码;全删全插替换项目成员,不做记录级数据权限过滤
// @Tags 开发管理/项目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveProjectUserRequest true "项目用户信息"
// @Success 200 {object} models.Response "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/project-users [put]
func (dc *DevController) SaveProjectUsers(c *gin.Context) {
	var req models.SaveProjectUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	if err := services.SaveProjectUsers(&req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

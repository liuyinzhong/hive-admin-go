package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 分页获取角色列表
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param roleName query string false "角色名称"
// @Param status query int false "状态"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.RoleSimpleResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles [get]
func (ctrl *SystemController) GetRoleList(c *gin.Context) {
	var req models.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.roleService.GetRoleList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllRoles 获取所有角色
// @Summary 获取所有角色
// @Description 获取所有角色（不分页）
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.RoleSimpleResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles/all [get]
func (ctrl *SystemController) GetAllRoles(c *gin.Context) {
	result, err := ctrl.roleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新角色
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateRoleRequest true "角色信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles [post]
func (ctrl *SystemController) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.roleService.CreateRole(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetRoleDetail 获取角色详情
// @Summary 获取角色详情
// @Description 根据角色ID获取角色详情
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param roleId path string true "角色ID"
// @Success 200 {object} models.Response{data=models.RoleDetailResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles/{roleId} [get]
func (ctrl *SystemController) GetRoleDetail(c *gin.Context) {
	roleId := c.Param("roleId")
	if roleId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "角色ID不能为空"))
		return
	}

	result, err := ctrl.roleService.GetRoleDetail(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param roleId path string true "角色ID"
// @Param request body models.UpdateRoleRequest true "角色信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles/{roleId} [put]
func (ctrl *SystemController) UpdateRole(c *gin.Context) {
	roleId := c.Param("roleId")
	if roleId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "角色ID不能为空"))
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.roleService.UpdateRole(roleId, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateRoleStatus 更新角色状态
// @Summary 更新角色状态
// @Description 更新角色启用/禁用状态
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param roleId path string true "角色ID"
// @Param request body map[string]int true "状态"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles/{roleId}/status [put]
func (ctrl *SystemController) UpdateRoleStatus(c *gin.Context) {
	roleId := c.Param("roleId")
	if roleId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "角色ID不能为空"))
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.roleService.UpdateRoleStatus(roleId, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteRoles 删除角色
// @Summary 删除角色
// @Description 批量删除角色
// @Tags 系统管理-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "角色ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/roles [delete]
func (ctrl *SystemController) DeleteRoles(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.roleService.DeleteRoles(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

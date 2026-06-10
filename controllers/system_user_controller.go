package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param realName query string false "用户姓名"
// @Param username query string false "用户名"
// @Param phone query string false "手机号"
// @Param status query int false "状态"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users [get]
func (ctrl *SystemController) GetUserList(c *gin.Context) {
	var req models.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.userService.GetUserList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllUsers 获取所有用户
// @Summary 获取所有用户
// @Description 获取所有用户（不分页）
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param realName query string false "用户姓名"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users/all [get]
func (ctrl *SystemController) GetAllUsers(c *gin.Context) {
	realName := c.Query("realName")

	result, err := ctrl.userService.GetAllUsers(realName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users [post]
func (ctrl *SystemController) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userService.CreateUser(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetUserDetail 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详情
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users/{userId} [get]
func (ctrl *SystemController) GetUserDetail(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	result, err := ctrl.userService.GetUserDetail(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body models.UpdateUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users/{userId} [put]
func (ctrl *SystemController) UpdateUser(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userService.UpdateUser(userId, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新用户启用/禁用状态
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body map[string]int true "状态"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users/{userId}/status [put]
func (ctrl *SystemController) UpdateUserStatus(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	var req models.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userService.UpdateUserStatus(userId, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteUsers 删除用户
// @Summary 删除用户
// @Description 批量删除用户
// @Tags 系统管理-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "用户ID列表"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users [delete]
func (ctrl *SystemController) DeleteUsers(c *gin.Context) {
	var userIds []string
	if err := c.ShouldBindJSON(&userIds); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	currentUserId, _ := c.Get("userId")
	if err := ctrl.userService.DeleteUsers(userIds, currentUserId.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

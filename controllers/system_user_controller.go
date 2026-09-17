package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 按当前角色数据范围分页获取普通用户；任一启用部门可见即进入列表，越界部门和直属领导关联不返回
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param realName query string false "用户姓名"
// @Param username query string false "用户名"
// @Param phone query string false "手机号"
// @Param status query int false "状态"
// @Param deptId query string false "部门ID，查询该部门及子部门的用户"
// @Param roleId query string false "角色ID，查询拥有该角色的用户"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ProfileResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users [get]
func (ctrl *SystemController) GetUserList(c *gin.Context) {
	var req models.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.userService.GetUserList(req, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllUsers 获取所有用户
// @Summary 获取所有用户
// @Description 按当前角色数据范围获取所有启用普通用户（不分页），供人员选择使用
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param realName query string false "用户姓名"
// @Success 200 {object} models.Response{data=[]models.ProfileResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/users/all [get]
func (ctrl *SystemController) GetAllUsers(c *gin.Context) {
	realName := c.Query("realName")

	result, err := ctrl.userService.GetAllUsers(realName, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户；部门必须在当前可管理范围，受限操作者只能分配自己持有且范围不越界的角色
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateUserRequest true "用户信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users [post]
func (ctrl *SystemController) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userService.CreateUser(req, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetUserDetail 获取用户详情
// @Summary 获取用户详情
// @Description 按当前角色数据范围获取用户详情；越界部门和直属领导关联不返回
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Success 200 {object} models.Response{data=models.ProfileResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users/{userId} [get]
func (ctrl *SystemController) GetUserDetail(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	result, err := ctrl.userService.GetUserDetail(userId, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetUserPermissions 获取用户权限明细
// @Summary 获取用户权限明细
// @Description 按用户维度一次返回其全部关联角色及各自菜单授权明细。数据权限：角色数据范围，与用户详情同边界，目标用户任一启用部门在操作者范围内即可查看。明细按配置事实返回：停用角色与停用菜单照常返回并携带状态；目录节点仅作路径前缀不单独成行；多角色重复授权的权限在各角色分组内分别出现，不去重
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Success 200 {object} models.Response{data=models.UserPermissionResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users/{userId}/permissions [get]
func (ctrl *SystemController) GetUserPermissions(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	result, err := ctrl.userService.GetUserPermissions(userId, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetUserPersonalPermissions 获取用户个人权限
// @Summary 获取用户个人权限
// @Description 获取目标用户的个人权限（个人额外授权与个人禁止两个菜单ID集合，只影响功能权限，不影响数据范围）。数据权限：角色数据范围，与用户详情同边界，目标用户任一启用部门在操作者范围内即可查看
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Success 200 {object} models.Response{data=models.UserPersonalPermissionResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users/{userId}/personalPermissions [get]
func (ctrl *SystemController) GetUserPersonalPermissions(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	result, err := ctrl.userMenuService.GetPersonalPermissions(userId, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// SaveUserPersonalPermissions 保存用户个人权限
// @Summary 保存用户个人权限
// @Description 按提交集合完整替换目标用户的个人额外授权与个人禁止（空数组表示清空对应集合，两集合不允许交叉）。数据权限：角色数据范围，与用户详情同边界；授予和禁止都要求菜单有效且在操作者生效权限内，全部数据范围操作者豁免（防提升，镜像角色分配约束）；系统内置用户不能配置个人权限
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body models.SaveUserPersonalPermissionRequest true "个人权限集合"
// @Success 200 {object} models.Response "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users/{userId}/personalPermissions [put]
func (ctrl *SystemController) SaveUserPersonalPermissions(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	var req models.SaveUserPersonalPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userMenuService.SavePersonalPermissions(userId, req, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 按当前角色数据范围更新用户；目标用户全部启用部门均须可管理，新的部门、角色和直属领导不得越界
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body models.UpdateUserRequest true "用户信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
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

	if err := ctrl.userService.UpdateUser(userId, req, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// ResetUserPassword 重置用户密码
// @Summary 重置用户密码
// @Description 按当前角色数据范围为目标用户直接设置新密码，不验证旧密码；系统内置用户不能重置。成功后目标用户密码版本号递增，其全部既有会话凭证立即失效，并推送强制退出事件和站内消息。数据权限：角色数据范围，写前在同一查询中校验目标用户可见性
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body models.ResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} models.Response "重置成功，目标用户需使用新密码重新登录"
// @Failure 400 {object} map[string]interface{} "参数错误、用户不存在或新密码不满足强度策略"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users/{userId}/password [put]
func (ctrl *SystemController) ResetUserPassword(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "用户ID不能为空"))
		return
	}

	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.userService.ResetUserPassword(userId, req, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 按当前角色数据范围更新用户启用/禁用状态；目标用户全部启用部门均须可管理
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param request body map[string]int true "状态"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
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

	if err := ctrl.userService.UpdateUserStatus(userId, req.Status, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteUsers 删除用户
// @Summary 删除用户
// @Description 按当前角色数据范围批量软删除用户；任一用户不存在、越界或包含越界部门时整批失败
// @Tags 系统管理/用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "用户ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/users [delete]
func (ctrl *SystemController) DeleteUsers(c *gin.Context) {
	var userIds []string
	if err := c.ShouldBindJSON(&userIds); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	currentUserId, _ := c.Get("userId")
	if err := ctrl.userService.DeleteUsers(userIds, currentUserId.(string), currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

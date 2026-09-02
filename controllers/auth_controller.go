package controllers

import (
	"hive-admin-go/models"
	"hive-admin-go/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过用户名和密码登录
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求参数"
// @Success 200 {object} models.Response{data=models.LoginResponse} "登录成功"
// @Failure 401 {object} map[string]interface{} "用户名或密码错误"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "请求参数错误"))
		return
	}

	token, err := ctrl.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.LoginResponse{
		AccessToken: token,
	}))
}

// GetProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的信息（含头像、签名等资料）。数据权限：当前用户归属，只返回当前 Token 对应用户的资料，不经过角色数据范围
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.ProfileResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/profile [get]
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	profile, err := ctrl.authService.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(profile))
}

// UpdateProfile 更新当前用户资料
// @Summary 更新当前用户资料
// @Description 当前用户更新自己的头像、邮箱和签名图片。数据权限：当前用户归属，只允许修改当前 Token 对应用户的记录，不经过角色数据范围；登录名、真实姓名等其余字段不在此接口开放。字段为 null 表示不修改，空字符串表示清空
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.UpdateProfileRequest true "资料更新请求参数"
// @Success 200 {object} models.Response{data=models.ProfileResponse} "更新成功，返回最新用户资料"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/profile [put]
func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "请求参数错误"))
		return
	}

	profile, err := ctrl.authService.UpdateProfile(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(profile))
}

// GetMenus 获取用户菜单
// @Summary 获取用户菜单
// @Description 获取当前登录用户的菜单权限
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.MenuTreeResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/menus [get]
func (ctrl *AuthController) GetMenus(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	menus, err := ctrl.authService.GetMenus(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(menus))
}

// GetAuthCodes 获取用户权限码
// @Summary 获取用户权限码
// @Description 获取当前登录用户的权限码列表
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]string} "获取成功"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/codes [get]
func (ctrl *AuthController) GetAuthCodes(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	codes, err := ctrl.authService.GetAuthCodes(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(codes))
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出系统
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response "登出成功"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	if err := ctrl.authService.Logout(token.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(""))
}

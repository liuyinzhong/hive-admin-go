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
// @Description 获取当前登录用户的信息
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

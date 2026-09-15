package controllers

import (
	"errors"
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
// @Description 用户通过用户名和密码登录。该用户名在失败窗口内失败达到阈值时要求滑块验证：未携带有效挑战票据返回 428，携带后正常校验。数据权限：公开接口，不经过认证和角色数据范围
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求参数"
// @Success 200 {object} models.Response{data=models.LoginResponse} "登录成功"
// @Failure 400 {object} map[string]interface{} "账号密码有误或账号被禁用"
// @Failure 428 {object} map[string]interface{} "失败达到阈值，需完成滑块验证后重试"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "请求参数错误"))
		return
	}

	token, err := ctrl.authService.Login(req)
	if err != nil {
		// 需要滑块验证用 428 与普通密码错误区分，前端据此展示滑块
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrCaptchaRequired) {
			status = http.StatusPreconditionRequired
		}
		c.JSON(status, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.LoginResponse{
		AccessToken: token,
	}))
}

// IssueCaptcha 签发滑块挑战票据
// @Summary 签发滑块挑战票据
// @Description 为登录页滑块验证签发一次性挑战票据，登录时携带并由服务端消费；同一 IP 每分钟最多签发 10 次。数据权限：公开接口，不经过认证和角色数据范围，票据不含任何用户数据
// @Tags 认证管理
// @Produce json
// @Success 200 {object} models.Response{data=models.CaptchaIssueResponse} "签发成功"
// @Failure 429 {object} map[string]interface{} "同一 IP 签发过于频繁"
// @Router /public/captcha [post]
func (ctrl *AuthController) IssueCaptcha(c *gin.Context) {
	captchaID, err := services.NewCaptchaService().Issue(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusTooManyRequests, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.CaptchaIssueResponse{
		CaptchaID: captchaID,
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

// ChangePassword 当前用户修改密码
// @Summary 修改密码
// @Description 当前用户验证旧密码后设置新密码；成功后密码版本号递增使全部既有会话凭证立即失效，并推送强制退出事件。数据权限：当前用户归属，只操作当前 Token 对应用户，不经过角色数据范围
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.ChangePasswordRequest true "修改密码请求参数"
// @Success 200 {object} models.Response "修改成功，需使用新密码重新登录"
// @Failure 400 {object} map[string]interface{} "参数错误、旧密码不正确或新密码不满足强度策略"
// @Failure 401 {object} map[string]interface{} "用户未登录"
// @Router /auth/password [put]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "请求参数错误"))
		return
	}

	if err := ctrl.authService.ChangePassword(userID.(string), req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
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

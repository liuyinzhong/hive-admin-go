package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type MenuMessageController struct {
	service *services.MenuMessageService
}

func NewMenuMessageController() *MenuMessageController {
	return &MenuMessageController{service: services.NewMenuMessageService()}
}

// GetUnreadSummary 获取当前用户的菜单未读汇总。
// @Summary 获取菜单未读汇总
// @Description 获取当前登录用户按菜单聚合的未读数量
// @Tags 系统管理/消息推送
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.MenuMessageUnreadSummary} "获取成功"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "获取失败"
// @Router /system/messages/unreadSummary [get]
func (ctrl *MenuMessageController) GetUnreadSummary(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	result, err := ctrl.service.GetUnreadSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "获取未读消息失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateDemoMessages 创建 Demo 菜单消息。
// @Summary 创建 Demo 菜单消息
// @Description 为当前数据范围内的指定用户在指定叶子菜单下新增指定数量的测试未读消息
// @Tags 系统管理/消息推送
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateMenuMessageRequest true "消息参数"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "创建失败"
// @Router /system/messages/demo [post]
func (ctrl *MenuMessageController) CreateDemoMessages(c *gin.Context) {
	var req models.CreateMenuMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.service.CreateDemoMessages(req, currentDataPermission(c)); err != nil {
		if errors.Is(err, services.ErrMenuMessageInvalidMenu) || errors.Is(err, services.ErrMenuMessageInvalidUsers) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "创建消息失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// MarkRead 将当前用户指定菜单下的未读消息全部标记为已读。
// @Summary 标记菜单消息已读
// @Description 将当前登录用户指定菜单下的全部未读消息标记为已读
// @Tags 系统管理/消息推送
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.ReadMenuMessageRequest true "菜单参数"
// @Success 200 {object} models.Response "操作成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "操作失败"
// @Router /system/messages/read [post]
func (ctrl *MenuMessageController) MarkRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	var req models.ReadMenuMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.service.MarkRead(userID, req.MenuID); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "标记已读失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// StreamUnreadSummary 建立当前用户的菜单未读 SSE 连接。
// @Summary 订阅菜单未读推送
// @Description 订阅当前登录用户的菜单未读汇总变更
// @Tags 系统管理/消息推送
// @Produce text/event-stream
// @Security ApiKeyAuth
// @Success 200 {string} string "SSE 事件流"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "建立连接失败"
// @Router /system/messages/stream [get]
func (ctrl *MenuMessageController) StreamUnreadSummary(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	if err := ctrl.service.StreamUnreadSummary(c, userID); err != nil && !c.Writer.Written() {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "建立消息推送连接失败"))
	}
}

// GetRecentMessages 获取当前用户最近的消息列表。
// @Summary 获取通知中心消息列表
// @Description 获取当前登录用户最近的消息列表(含已读),按创建时间倒序,最多返回100条。数据权限:当前用户归属,仅按登录用户 user_id 过滤,不使用角色数据范围
// @Tags 系统管理/消息推送
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.MenuMessageItem} "获取成功"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "获取失败"
// @Router /system/messages [get]
func (ctrl *MenuMessageController) GetRecentMessages(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	result, err := ctrl.service.GetRecentMessages(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "获取消息列表失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ReadMessage 将当前用户的一条菜单消息标记为已读。
// @Summary 逐条标记消息已读
// @Description 将当前登录用户的一条菜单消息标记为已读。数据权限:当前用户归属,仅能操作本人消息
// @Tags 系统管理/消息推送
// @Produce json
// @Security ApiKeyAuth
// @Param messageId path string true "消息ID"
// @Success 200 {object} models.Response "操作成功"
// @Failure 400 {object} models.Response "消息不存在或已读"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "操作失败"
// @Router /system/messages/{messageId}/read [put]
func (ctrl *MenuMessageController) ReadMessage(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	if err := ctrl.service.ReadMessage(userID, c.Param("messageId")); err != nil {
		if errors.Is(err, services.ErrMenuMessageNotFound) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "标记已读失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// ReadAllMessages 将当前用户的全部未读消息标记为已读。
// @Summary 全部标记已读
// @Description 将当前登录用户的全部未读菜单消息一次性标记为已读。数据权限:当前用户归属,仅操作本人消息
// @Tags 系统管理/消息推送
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response "操作成功"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "操作失败"
// @Router /system/messages/readAll [put]
func (ctrl *MenuMessageController) ReadAllMessages(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}

	if err := ctrl.service.ReadAllMessages(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "标记已读失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func currentUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get("userId")
	if !exists {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok && userID != ""
}

package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetOperationLogs 获取操作日志列表
// @Summary 获取操作日志列表
// @Description 分页查询最近七天的操作日志
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param username query string false "用户名"
// @Param requestUrl query string false "请求URL"
// @Param requestMethod query string false "请求方法"
// @Param status query int false "状态，0失败 1成功"
// @Param startDate query string false "开始时间"
// @Param endDate query string false "结束时间"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.OperationLogListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/operationLogs [get]
func (ctrl *SystemController) GetOperationLogs(c *gin.Context) {
	var req models.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := ctrl.auditLogService.GetOperationLogs(req)
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetOperationLog 获取操作日志详情
// @Summary 获取操作日志详情
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param logId path string true "日志ID"
// @Success 200 {object} models.Response{data=models.OperationLogDetailResponse} "获取成功"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 404 {object} models.Response "日志不存在"
// @Router /system/operationLogs/{logId} [get]
func (ctrl *SystemController) GetOperationLog(c *gin.Context) {
	result, err := ctrl.auditLogService.GetOperationLog(c.Param("logId"))
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetLoginLogs 获取登录日志列表
// @Summary 获取登录日志列表
// @Description 分页查询最近七天的登录和退出日志
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param username query string false "用户名"
// @Param ip query string false "客户端IP"
// @Param eventType query string false "类型，login或logout"
// @Param status query int false "状态，0失败 1成功"
// @Param startDate query string false "开始时间"
// @Param endDate query string false "结束时间"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.LoginLogListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/loginLogs [get]
func (ctrl *SystemController) GetLoginLogs(c *gin.Context) {
	var req models.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := ctrl.auditLogService.GetLoginLogs(req)
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetLoginLog 获取登录日志详情
// @Summary 获取登录日志详情
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param logId path string true "日志ID"
// @Success 200 {object} models.Response{data=models.LoginLogDetailResponse} "获取成功"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 404 {object} models.Response "日志不存在"
// @Router /system/loginLogs/{logId} [get]
func (ctrl *SystemController) GetLoginLog(c *gin.Context) {
	result, err := ctrl.auditLogService.GetLoginLog(c.Param("logId"))
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeAuditLogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrAuditLogInvalidInput):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrAuditLogNotFound):
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, "日志不存在"))
	default:
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "日志查询失败"))
	}
}

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
// @Description 按日志用户和当前角色数据范围分页查询操作日志。未传时间范围时默认查询最近 7 天数据；日志表保留 180 天，超期会被物理删除。
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认 1" default(1)
// @Param pageSize query int false "每页大小，默认 20，最大 100" default(20)
// @Param username query string false "用户名，模糊匹配"
// @Param requestUrl query string false "请求 URL，模糊匹配"
// @Param requestMethod query string false "请求方法，精确匹配（自动转大写），例如 GET、POST、PUT、DELETE"
// @Param status query int false "操作状态：0 失败，1 成功；传非 0/1 返回 400" Enums(0,1)
// @Param startDate query string false "开始时间，支持格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、RFC3339；不传时默认为当前时间往前 7 天"
// @Param endDate query string false "结束时间，支持格式同 startDate；仅传日期时取当天 23:59:59.999；不传时默认为当前时间；不能早于 startDate"
// @Param sorts query string false "排序，格式 field,desc;field,asc；可排序字段：createDate、durationMs、httpStatus；未传时默认 createDate desc"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.OperationLogListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误：状态值非法、时间格式不正确或结束时间早于开始时间"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "日志查询失败"
// @Router /system/operationLogs [get]
func (ctrl *SystemController) GetOperationLogs(c *gin.Context) {
	var req models.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := ctrl.auditLogService.GetOperationLogs(req, currentDataPermission(c))
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetOperationLog 获取操作日志详情
// @Summary 获取操作日志详情
// @Description 按日志用户和当前角色数据范围获取操作日志详情，包含请求参数、请求体、响应体等完整信息
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param logId path string true "日志 ID（UUID，带横线）"
// @Success 200 {object} models.Response{data=models.OperationLogDetailResponse} "获取成功"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 404 {object} models.Response "日志不存在"
// @Failure 500 {object} models.Response "日志查询失败"
// @Router /system/operationLogs/{logId} [get]
func (ctrl *SystemController) GetOperationLog(c *gin.Context) {
	result, err := ctrl.auditLogService.GetOperationLog(c.Param("logId"), currentDataPermission(c))
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetLoginLogs 获取登录日志列表
// @Summary 获取登录日志列表
// @Description 按日志用户和当前角色数据范围分页查询登录与退出日志。未传时间范围时默认查询最近 7 天数据；空用户日志仅全部数据范围可见；日志表保留 180 天，超期会被物理删除。
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认 1" default(1)
// @Param pageSize query int false "每页大小，默认 20，最大 100" default(20)
// @Param username query string false "用户名，模糊匹配"
// @Param ip query string false "客户端 IP，模糊匹配"
// @Param eventType query string false "事件类型：login 登录、logout 退出；传其他值返回 400" Enums(login,logout)
// @Param status query int false "操作状态：0 失败，1 成功；传非 0/1 返回 400" Enums(0,1)
// @Param startDate query string false "开始时间，支持格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、RFC3339；不传时默认为当前时间往前 7 天"
// @Param endDate query string false "结束时间，支持格式同 startDate；仅传日期时取当天 23:59:59.999；不传时默认为当前时间；不能早于 startDate"
// @Param sorts query string false "排序，格式 field,desc;field,asc；可排序字段：createDate、durationMs、httpStatus；未传时默认 createDate desc"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.LoginLogListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误：事件类型非法、状态值非法、时间格式不正确或结束时间早于开始时间"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "日志查询失败"
// @Router /system/loginLogs [get]
func (ctrl *SystemController) GetLoginLogs(c *gin.Context) {
	var req models.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := ctrl.auditLogService.GetLoginLogs(req, currentDataPermission(c))
	if err != nil {
		writeAuditLogError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetLoginLog 获取登录日志详情
// @Summary 获取登录日志详情
// @Description 按日志用户和当前角色数据范围获取登录日志详情，包含响应体、内容类型等完整信息
// @Tags 系统管理/日志管理
// @Produce json
// @Security ApiKeyAuth
// @Param logId path string true "日志 ID（UUID，带横线）"
// @Success 200 {object} models.Response{data=models.LoginLogDetailResponse} "获取成功"
// @Failure 401 {object} models.Response "未认证"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 404 {object} models.Response "日志不存在"
// @Failure 500 {object} models.Response "日志查询失败"
// @Router /system/loginLogs/{logId} [get]
func (ctrl *SystemController) GetLoginLog(c *gin.Context) {
	result, err := ctrl.auditLogService.GetLoginLog(c.Param("logId"), currentDataPermission(c))
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

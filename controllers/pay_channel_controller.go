package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetPayChannelList 分页查询支付渠道列表
// @Summary 分页查询支付渠道列表
// @Description 分页查询支付渠道配置列表,支持按名称、类型、环境、状态、默认筛选与排序
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码,默认1" default(1)
// @Param pageSize query int false "每页大小,默认20" default(20)
// @Param channelName query string false "渠道配置名称,模糊搜索" example(微信)
// @Param channelType query string false "渠道类型 wechat/alipay" example(wechat)
// @Param envMode query string false "环境模式 development/testing/staging/production" example(production)
// @Param status query int false "启用状态 0=禁用 1=启用" example(1)
// @Param isDefault query int false "是否默认 0=否 1=是" example(1)
// @Param sorts query string false "排序参数" example(updateDate,desc)
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.PayChannelResponse}} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/payChannels [get]
func (ctrl *SystemController) GetPayChannelList(c *gin.Context) {
	var req models.PayChannelListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.payChannelService.GetPayChannelList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreatePayChannel 创建支付渠道
// @Summary 创建支付渠道
// @Description 创建新的支付渠道配置,校验渠道类型/环境模式枚举、extraConfig 合法性、默认渠道唯一性
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreatePayChannelRequest true "支付渠道信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "校验失败或创建失败"
// @Router /system/payChannels [post]
func (ctrl *SystemController) CreatePayChannel(c *gin.Context) {
	var req models.CreatePayChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.payChannelService.CreatePayChannel(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetPayChannelDetail 查询支付渠道详情
// @Summary 查询支付渠道详情
// @Description 根据ID查询支付渠道配置详情
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "支付渠道ID"
// @Success 200 {object} models.Response{data=models.PayChannelResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "支付渠道不存在"
// @Router /system/payChannels/{id} [get]
func (ctrl *SystemController) GetPayChannelDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "支付渠道ID不能为空"))
		return
	}

	result, err := ctrl.payChannelService.GetPayChannelDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdatePayChannel 更新支付渠道
// @Summary 更新支付渠道
// @Description 更新支付渠道配置,允许修改全部字段;设为默认时自动取消同组其他默认
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "支付渠道ID"
// @Param request body models.UpdatePayChannelRequest true "支付渠道信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "支付渠道不存在或校验失败"
// @Router /system/payChannels/{id} [put]
func (ctrl *SystemController) UpdatePayChannel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "支付渠道ID不能为空"))
		return
	}

	var req models.UpdatePayChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.payChannelService.UpdatePayChannel(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeletePayChannels 批量删除支付渠道
// @Summary 批量删除支付渠道
// @Description 批量逻辑删除支付渠道配置
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "支付渠道ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/payChannels [delete]
func (ctrl *SystemController) DeletePayChannels(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.payChannelService.DeletePayChannels(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdatePayChannelStatus 修改支付渠道启用状态
// @Summary 修改支付渠道启用状态
// @Description 启用或禁用支付渠道
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "支付渠道ID"
// @Param request body models.UpdatePayChannelStatusRequest true "启用状态"
// @Success 200 {object} models.Response "修改成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "支付渠道不存在"
// @Router /system/payChannels/{id}/status [put]
func (ctrl *SystemController) UpdatePayChannelStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "支付渠道ID不能为空"))
		return
	}

	var req models.UpdatePayChannelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.payChannelService.UpdatePayChannelStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdatePayChannelDefault 修改支付渠道默认标记
// @Summary 修改支付渠道默认标记
// @Description 设置或取消支付渠道默认标记,设为默认时自动取消同 channelType+envMode 下其他默认
// @Tags 系统管理/支付渠道
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "支付渠道ID"
// @Param request body models.UpdatePayChannelDefaultRequest true "默认标记"
// @Success 200 {object} models.Response "修改成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "支付渠道不存在"
// @Router /system/payChannels/{id}/default [put]
func (ctrl *SystemController) UpdatePayChannelDefault(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "支付渠道ID不能为空"))
		return
	}

	var req models.UpdatePayChannelDefaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.payChannelService.UpdatePayChannelDefault(id, req.IsDefault); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

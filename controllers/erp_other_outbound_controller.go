package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ErpOtherOutboundController struct {
	otherOutboundService *services.ErpOtherOutboundService
}

func NewErpOtherOutboundController() *ErpOtherOutboundController {
	return &ErpOtherOutboundController{
		otherOutboundService: services.NewErpOtherOutboundService(),
	}
}

// GetOtherOutboundList 获取其它出库单列表
// @Summary 获取其它出库单列表
// @Description 按出库单创建人及当前角色数据范围分页查询其它出库单，支持按单号、仓库、SKU编码、批号和出库日期范围筛选
// @Tags 进销存/其它出库
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，最大100"
// @Param outboundNo query string false "其它出库单号"
// @Param warehouseId query string false "仓库ID，UUID格式"
// @Param skuCode query string false "SKU编码"
// @Param batchNo query string false "批号"
// @Param outboundDateFrom query string false "出库开始日期，格式YYYY-MM-DD"
// @Param outboundDateTo query string false "出库结束日期，格式YYYY-MM-DD"
// @Param sorts query string false "排序，支持outboundNo、warehouseName、outboundDate、lineCount、createDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpOtherOutboundListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/otherOutbounds [get]
func (ctrl *ErpOtherOutboundController) GetOtherOutboundList(c *gin.Context) {
	var req models.ErpOtherOutboundListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.otherOutboundService.GetOtherOutboundList(req, currentDataPermission(c))
	if err != nil {
		writeErpOtherOutboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetOtherOutboundDetail 获取其它出库单详情
// @Summary 获取其它出库单详情
// @Description 按出库单创建人及当前角色数据范围获取单头、明细和当前库存关联展示信息；提交后只读
// @Tags 进销存/其它出库
// @Produce json
// @Security ApiKeyAuth
// @Param outboundId path string true "其它出库单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.ErpOtherOutboundResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "其它出库单不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/otherOutbounds/{outboundId} [get]
func (ctrl *ErpOtherOutboundController) GetOtherOutboundDetail(c *gin.Context) {
	result, err := ctrl.otherOutboundService.GetOtherOutboundDetail(c.Param("outboundId"), currentDataPermission(c))
	if err != nil {
		writeErpOtherOutboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateOtherOutbound 新增其它出库单
// @Summary 新增其它出库单
// @Description 提交即完成其它出库；每个库存余额必须处于当前数据范围，新出库单和库存流水归属当前用户，整批原子扣减库存
// @Tags 进销存/其它出库
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateErpOtherOutboundRequest true "其它出库单"
// @Success 200 {object} models.Response{data=models.ErpOtherOutboundResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "仓库、库存余额或SKU不存在"
// @Failure 409 {object} models.Response "明细重复、库存不足或库存写入冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/otherOutbounds [post]
func (ctrl *ErpOtherOutboundController) CreateOtherOutbound(c *gin.Context) {
	var req models.CreateErpOtherOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.otherOutboundService.CreateOtherOutbound(req, erpInventoryOperatorID(c), currentDataPermission(c))
	if err != nil {
		writeErpOtherOutboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeErpOtherOutboundError(c *gin.Context, err error) {
	message := "其它出库操作失败"
	switch {
	case errors.Is(err, services.ErrErpOtherOutboundInvalidInput), errors.Is(err, services.ErrErpInventoryInvalidInput):
		message = err.Error()
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrErpOtherOutboundNotFound), errors.Is(err, services.ErrErpInventoryNotFound):
		message = err.Error()
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrErpOtherOutboundConflict), errors.Is(err, services.ErrErpInventoryConflict):
		message = err.Error()
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, message))
	default:
		log.Printf("其它出库操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
	}
}

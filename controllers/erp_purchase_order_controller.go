package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ErpPurchaseOrderController struct {
	service *services.ErpPurchaseOrderService
}

func NewErpPurchaseOrderController() *ErpPurchaseOrderController {
	return &ErpPurchaseOrderController{service: services.NewErpPurchaseOrderService()}
}

// GetPurchaseOrderList 获取采购单列表
// @Summary 获取采购单列表
// @Description 按采购单创建人及当前角色数据范围分页查询采购单，支持单号、供应商、仓库、SKU编码、状态和采购日期范围筛选
// @Tags 进销存/采购单
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，最大100"
// @Param purchaseOrderNo query string false "采购单号"
// @Param supplierId query string false "供应商ID"
// @Param warehouseId query string false "仓库ID"
// @Param skuCode query string false "SKU编码"
// @Param status query string false "状态" Enums(DRAFT,WAITING_RECEIPT,PARTIAL_RECEIPT,COMPLETED,CANCELLED,CLOSED)
// @Param orderDateFrom query string false "采购日期起，YYYY-MM-DD"
// @Param orderDateTo query string false "采购日期止，YYYY-MM-DD"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpPurchaseOrderListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders [get]
func (ctrl *ErpPurchaseOrderController) GetPurchaseOrderList(c *gin.Context) {
	var req models.ErpPurchaseOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.service.GetPurchaseOrderList(req, currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPurchaseOrderDetail 获取采购单详情
// @Summary 获取采购单详情
// @Description 按采购单创建人及当前角色数据范围获取采购单头和明细
// @Tags 进销存/采购单
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId} [get]
func (ctrl *ErpPurchaseOrderController) GetPurchaseOrderDetail(c *gin.Context) {
	result, err := ctrl.service.GetPurchaseOrderDetail(c.Param("purchaseOrderId"), currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreatePurchaseOrder 创建采购单草稿
// @Summary 创建采购单草稿
// @Description 创建结构完整的采购单草稿并生成不可复用的采购单号，以当前用户作为归属人
// @Tags 进销存/采购单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveErpPurchaseOrderRequest true "采购单"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "供应商、仓库或SKU不存在或未启用"
// @Failure 409 {object} models.Response "明细重复"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders [post]
func (ctrl *ErpPurchaseOrderController) CreatePurchaseOrder(c *gin.Context) {
	var req models.SaveErpPurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.service.CreatePurchaseOrder(req, erpInventoryOperatorID(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdatePurchaseOrder 修改采购单草稿
// @Summary 修改采购单草稿
// @Description 按采购单创建人及当前角色数据范围修改草稿
// @Tags 进销存/采购单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Param request body models.SaveErpPurchaseOrderRequest true "采购单及期望版本"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "修改成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单或基础资料不存在"
// @Failure 409 {object} models.Response "状态或并发版本冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId} [put]
func (ctrl *ErpPurchaseOrderController) UpdatePurchaseOrder(c *gin.Context) {
	var req models.SaveErpPurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.service.UpdatePurchaseOrder(c.Param("purchaseOrderId"), req, erpInventoryOperatorID(c), currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ConfirmPurchaseOrder 确认采购单
// @Summary 确认采购单
// @Description 按当前数据范围将草稿确认并锁定，状态变更为待收货
// @Tags 进销存/采购单
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "确认成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单或基础资料不存在"
// @Failure 409 {object} models.Response "状态冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId}/confirm [post]
func (ctrl *ErpPurchaseOrderController) ConfirmPurchaseOrder(c *gin.Context) {
	result, err := ctrl.service.ConfirmPurchaseOrder(c.Param("purchaseOrderId"), erpInventoryOperatorID(c), currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CancelPurchaseOrder 取消采购单
// @Summary 取消采购单
// @Description 按当前数据范围取消草稿或待收货采购单
// @Tags 进销存/采购单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Param request body models.ErpPurchaseOrderReasonRequest true "取消原因"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "取消成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单不存在"
// @Failure 409 {object} models.Response "状态冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId}/cancel [post]
func (ctrl *ErpPurchaseOrderController) CancelPurchaseOrder(c *gin.Context) {
	ctrl.reasonAction(c, ctrl.service.CancelPurchaseOrder)
}

// ClosePurchaseOrder 关闭部分入库采购单
// @Summary 关闭采购单
// @Description 按当前数据范围关闭部分入库采购单，未入库的剩余数量永久失效
// @Tags 进销存/采购单
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Param request body models.ErpPurchaseOrderReasonRequest true "关闭原因"
// @Success 200 {object} models.Response{data=models.ErpPurchaseOrderResponse} "关闭成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单不存在"
// @Failure 409 {object} models.Response "状态冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId}/close [post]
func (ctrl *ErpPurchaseOrderController) ClosePurchaseOrder(c *gin.Context) {
	ctrl.reasonAction(c, ctrl.service.ClosePurchaseOrder)
}

type erpPurchaseOrderReasonAction func(string, string, string, datapermission.Permission) (*models.ErpPurchaseOrderResponse, error)

func (ctrl *ErpPurchaseOrderController) reasonAction(c *gin.Context, action erpPurchaseOrderReasonAction) {
	var req models.ErpPurchaseOrderReasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := action(c.Param("purchaseOrderId"), erpInventoryOperatorID(c), req.Reason, currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPurchaseOrderLogs 获取采购单业务日志
// @Summary 获取采购单业务日志
// @Description 校验采购单当前数据范围后，返回不可修改的创建、修改、确认、入库、取消和关闭日志
// @Tags 进销存/采购单
// @Produce json
// @Security ApiKeyAuth
// @Param purchaseOrderId path string true "采购单ID，UUID格式"
// @Success 200 {object} models.Response{data=[]models.ErpPurchaseOrderLogResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseOrders/{purchaseOrderId}/logs [get]
func (ctrl *ErpPurchaseOrderController) GetPurchaseOrderLogs(c *gin.Context) {
	result, err := ctrl.service.GetPurchaseOrderLogs(c.Param("purchaseOrderId"), currentDataPermission(c))
	if err != nil {
		writeErpPurchaseOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeErpPurchaseOrderError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "采购单操作失败"
	switch {
	case errors.Is(err, services.ErrErpPurchaseOrderInvalidInput):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, services.ErrErpPurchaseOrderNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, services.ErrErpPurchaseOrderConflict):
		status, message = http.StatusConflict, err.Error()
	default:
		log.Printf("采购单操作失败: %v", err)
	}
	c.JSON(status, models.NewErrorResponse(nil, message))
}

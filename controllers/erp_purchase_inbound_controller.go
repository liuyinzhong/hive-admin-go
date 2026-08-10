package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ErpPurchaseInboundController struct {
	purchaseInboundService *services.ErpPurchaseInboundService
}

func NewErpPurchaseInboundController() *ErpPurchaseInboundController {
	return &ErpPurchaseInboundController{
		purchaseInboundService: services.NewErpPurchaseInboundService(),
	}
}

// GetPurchaseInboundList 获取采购入库单列表
// @Summary 获取采购入库单列表
// @Description 分页查询采购入库单，支持按入库单号、采购单号、供应商、仓库、SKU编码、批号和入库日期范围筛选
// @Tags 进销存/采购入库
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量，最大100"
// @Param inboundNo query string false "采购入库单号"
// @Param purchaseOrderNo query string false "来源采购单号"
// @Param supplierId query string false "供应商ID，UUID格式"
// @Param warehouseId query string false "仓库ID，UUID格式"
// @Param skuCode query string false "SKU编码"
// @Param batchNo query string false "批号"
// @Param inboundDateFrom query string false "入库开始日期，格式YYYY-MM-DD"
// @Param inboundDateTo query string false "入库结束日期，格式YYYY-MM-DD"
// @Param sorts query string false "排序，支持inboundNo、supplierName、warehouseName、inboundDate、lineCount、totalAmount、createDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpPurchaseInboundListResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseInbounds [get]
func (ctrl *ErpPurchaseInboundController) GetPurchaseInboundList(c *gin.Context) {
	var req models.ErpPurchaseInboundListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.purchaseInboundService.GetPurchaseInboundList(req)
	if err != nil {
		writeErpPurchaseInboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPurchaseInboundDetail 获取采购入库单详情
// @Summary 获取采购入库单详情
// @Description 获取采购入库单头、明细和当前SKU展示信息；采购入库单提交后只读
// @Tags 进销存/采购入库
// @Produce json
// @Security ApiKeyAuth
// @Param inboundId path string true "采购入库单ID，UUID格式"
// @Success 200 {object} models.Response{data=models.ErpPurchaseInboundResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购入库单不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseInbounds/{inboundId} [get]
func (ctrl *ErpPurchaseInboundController) GetPurchaseInboundDetail(c *gin.Context) {
	result, err := ctrl.purchaseInboundService.GetPurchaseInboundDetail(c.Param("inboundId"))
	if err != nil {
		writeErpPurchaseInboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreatePurchaseInbound 新增采购入库单
// @Summary 新增采购入库单
// @Description 只能从待收货或部分入库采购单发起；供应商、仓库、SKU和成本价继承采购单，批量保存单据和明细并原子更新采购单收货进度、库存与采购单日志；启用追溯的SKU必须提交纯数字小包装追溯码且数量与追溯码个数一致
// @Tags 进销存/采购入库
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateErpPurchaseInboundRequest true "采购入库单"
// @Success 200 {object} models.Response{data=models.ErpPurchaseInboundResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "采购单、采购明细、仓库或SKU不存在"
// @Failure 409 {object} models.Response "采购单状态、剩余数量、明细或库存写入冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/purchaseInbounds [post]
func (ctrl *ErpPurchaseInboundController) CreatePurchaseInbound(c *gin.Context) {
	var req models.CreateErpPurchaseInboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.purchaseInboundService.CreatePurchaseInbound(req, erpInventoryOperatorID(c))
	if err != nil {
		writeErpPurchaseInboundError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeErpPurchaseInboundError(c *gin.Context, err error) {
	message := "采购入库操作失败"
	switch {
	case errors.Is(err, services.ErrErpPurchaseInboundInvalidInput), errors.Is(err, services.ErrErpInventoryInvalidInput):
		message = err.Error()
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrErpPurchaseInboundNotFound), errors.Is(err, services.ErrErpInventoryNotFound):
		message = err.Error()
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrErpPurchaseInboundConflict), errors.Is(err, services.ErrErpInventoryConflict):
		message = err.Error()
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, message))
	default:
		log.Printf("采购入库操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
	}
}

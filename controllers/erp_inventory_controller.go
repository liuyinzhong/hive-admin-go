package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ErpInventoryController struct {
	inventoryService *services.ErpInventoryService
}

func NewErpInventoryController() *ErpInventoryController {
	return &ErpInventoryController{
		inventoryService: services.NewErpInventoryService(),
	}
}

// GetInventoryBalanceList 获取库存余额列表
// @Summary 获取库存余额列表
// @Description 分页查询库存余额，第一版按仓库和库存批次展示；搜索条件支持仓库ID、库存余额ID、SKU编码和批号，默认包含零库存余额
// @Tags 进销存/库存管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param warehouseId query string false "仓库ID，UUID格式"
// @Param balanceIds query string false "库存余额ID，多个ID以逗号分隔，最多100个"
// @Param skuCode query string false "SKU编码"
// @Param batchNo query string false "批号"
// @Param onlyPositive query bool false "是否仅返回包装单位库存大于0的余额"
// @Param sorts query string false "排序，支持warehouseCode、warehouseName、skuCode、batchNo、expiryDate、unitCost、packageUnitCount、minUnitCount、inventoryAmount、movementCount、createDate、updateDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpInventoryBalanceResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/inventory/balances [get]
func (ctrl *ErpInventoryController) GetInventoryBalanceList(c *gin.Context) {
	var req models.ErpInventoryBalanceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.inventoryService.GetInventoryBalanceList(req)
	if err != nil {
		writeErpInventoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetInventoryMovements 获取库存流水列表
// @Summary 获取库存流水列表
// @Description 根据库存余额ID分页查询库存流水；流水一旦写入不可更改，第一版仅包含初始库存入库流水
// @Tags 进销存/库存管理
// @Produce json
// @Security ApiKeyAuth
// @Param balanceId path string true "库存余额ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param sorts query string false "排序，支持sourceBillType、sourceBillNo、movementType、direction、createDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpInventoryMovementResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/inventory/balances/{balanceId}/movements [get]
func (ctrl *ErpInventoryController) GetInventoryMovements(c *gin.Context) {
	var req models.ErpInventoryMovementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.inventoryService.GetInventoryMovements(c.Param("balanceId"), req)
	if err != nil {
		writeErpInventoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetInventoryMovementsBySource 获取来源单据库存流水列表
// @Summary 获取来源单据库存流水列表
// @Description 根据来源单据类型和来源单据ID分页查询库存流水，适用于查看整张采购入库单产生的全部流水
// @Tags 进销存/库存管理
// @Produce json
// @Security ApiKeyAuth
// @Param sourceBillType query string true "来源单据类型，例如PURCHASE_INBOUND"
// @Param sourceBillId query string true "来源单据ID，UUID格式"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param sorts query string false "排序，支持sourceBillType、sourceBillNo、movementType、direction、createDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpInventoryMovementResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/inventory/movements [get]
func (ctrl *ErpInventoryController) GetInventoryMovementsBySource(c *gin.Context) {
	var req models.ErpInventorySourceMovementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.inventoryService.GetInventoryMovementsBySource(req)
	if err != nil {
		writeErpInventoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateInitialStocks 新增初始库存
// @Summary 新增初始库存
// @Description 批量写入初始库存，数量按SKU包装单位录入；同一批次再次写入是追加库存而不是覆盖库存；整批提交原子成功或失败
// @Tags 进销存/库存管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateErpInventoryInitialStockRequest true "初始库存"
// @Success 200 {object} models.Response{data=models.CreateErpInventoryInitialStockResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/inventory/initialStocks [post]
func (ctrl *ErpInventoryController) CreateInitialStocks(c *gin.Context) {
	var req models.CreateErpInventoryInitialStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.inventoryService.CreateInitialStocks(req, erpInventoryOperatorID(c))
	if err != nil {
		writeErpInventoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func erpInventoryOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeErpInventoryError(c *gin.Context, err error) {
	message := "库存操作失败"
	switch {
	case errors.Is(err, services.ErrErpInventoryInvalidInput):
		message = err.Error()
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, message))
		return
	case errors.Is(err, services.ErrErpInventoryNotFound):
		message = err.Error()
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, message))
		return
	case errors.Is(err, services.ErrErpInventoryConflict):
		message = err.Error()
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, message))
		return
	default:
		log.Printf("库存操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
}

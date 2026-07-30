package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ErpWarehouseController struct {
	warehouseService *services.ErpWarehouseService
}

func NewErpWarehouseController() *ErpWarehouseController {
	return &ErpWarehouseController{
		warehouseService: services.NewErpWarehouseService(),
	}
}

// GetWarehouseList 获取仓库列表
// @Summary 获取仓库列表
// @Description 分页查询仓库基础资料列表，支持按仓库编码/名称、储存类型、业务范围和状态筛选
// @Tags 进销存/仓库管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "仓库编码或仓库名称"
// @Param storageType query string false "仓库储存类型：NORMAL普通 REFRIGERATED冷藏 FROZEN冷冻 COOL阴凉 HAZARDOUS危险品"
// @Param businessScope query string false "仓库业务范围：DRUG药品 CONSUMABLE耗材 DEVICE器械 COMPREHENSIVE综合"
// @Param status query int false "状态：0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpWarehouseResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses [get]
func (ctrl *ErpWarehouseController) GetWarehouseList(c *gin.Context) {
	var req models.ErpWarehouseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseList(req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouse 获取仓库详情
// @Summary 获取仓库详情
// @Description 根据仓库ID获取仓库基础资料详情
// @Tags 进销存/仓库管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Success 200 {object} models.Response{data=models.ErpWarehouseResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId} [get]
func (ctrl *ErpWarehouseController) GetWarehouse(c *gin.Context) {
	result, err := ctrl.warehouseService.GetWarehouseDetail(c.Param("warehouseId"))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouseOptions 获取启用仓库选项
// @Summary 获取启用仓库选项
// @Description 获取已启用的仓库基础资料选项，用于下拉选择等场景；需要登录，不做仓库按钮权限校验
// @Tags 进销存/仓库管理
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "仓库编码或仓库名称"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ErpWarehouseOptionResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/options [get]
func (ctrl *ErpWarehouseController) GetWarehouseOptions(c *gin.Context) {
	var req models.ErpWarehouseOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseOptions(req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateWarehouse 新增仓库
// @Summary 新增仓库
// @Description 创建仓库基础资料，仓库编码由后端自动生成
// @Tags 进销存/仓库管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveErpWarehouseRequest true "仓库"
// @Success 200 {object} models.Response{data=models.ErpWarehouseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses [post]
func (ctrl *ErpWarehouseController) CreateWarehouse(c *gin.Context) {
	var req models.SaveErpWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.CreateWarehouse(req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateWarehouse 更新仓库
// @Summary 更新仓库
// @Description 根据仓库ID更新仓库基础资料，仓库编码不可修改
// @Tags 进销存/仓库管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param request body models.SaveErpWarehouseRequest true "仓库"
// @Success 200 {object} models.Response{data=models.ErpWarehouseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId} [put]
func (ctrl *ErpWarehouseController) UpdateWarehouse(c *gin.Context) {
	var req models.SaveErpWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.UpdateWarehouse(c.Param("warehouseId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateWarehouseStatus 更新仓库启停状态
// @Summary 更新仓库启停状态
// @Description 根据仓库ID更新仓库启用/停用状态
// @Tags 进销存/仓库管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param request body models.UpdateErpWarehouseStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ErpWarehouseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/status [put]
func (ctrl *ErpWarehouseController) UpdateWarehouseStatus(c *gin.Context) {
	var req models.UpdateErpWarehouseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.UpdateWarehouseStatus(c.Param("warehouseId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteWarehouse 删除仓库
// @Summary 删除仓库
// @Description 根据仓库ID软删除仓库基础资料，第一版不校验库存或单据引用
// @Tags 进销存/仓库管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param request body models.DeleteErpWarehouseRequest true "删除参数"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId} [delete]
func (ctrl *ErpWarehouseController) DeleteWarehouse(c *gin.Context) {
	var req models.DeleteErpWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.warehouseService.DeleteWarehouse(c.Param("warehouseId"), req, erpWarehouseOperatorID(c)); err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func erpWarehouseOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeErpWarehouseError(c *gin.Context, err error) {
	message := "仓库操作失败"
	switch {
	case errors.Is(err, services.ErrErpWarehouseInvalidInput),
		errors.Is(err, services.ErrErpWarehouseNotFound),
		errors.Is(err, services.ErrErpWarehouseConflict):
		message = err.Error()
	default:
		log.Printf("仓库操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

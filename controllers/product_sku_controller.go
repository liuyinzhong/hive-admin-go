package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ProductSkuController struct {
	productSkuService *services.ProductSkuService
}

func NewProductSkuController() *ProductSkuController {
	return &ProductSkuController{
		productSkuService: services.NewProductSkuService(),
	}
}

// GetProductSkuList 获取SKU列表
// @Summary 获取SKU列表
// @Description 分页查询SKU列表，按所属厂家产品ID筛选及排序
// @Tags 产品档案/SKU
// @Produce json
// @Security ApiKeyAuth
// @Param mpId query string true "所属厂家产品ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ProductSkuResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus [get]
func (ctrl *ProductSkuController) GetProductSkuList(c *gin.Context) {
	var req models.ProductSkuListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuService.GetProductSkuList(req)
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductSku 获取SKU详情
// @Summary 获取SKU详情
// @Description 根据SKU ID查询SKU详细信息
// @Tags 产品档案/SKU
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Success 200 {object} models.Response{data=models.ProductSkuResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId} [get]
func (ctrl *ProductSkuController) GetProductSku(c *gin.Context) {
	result, err := ctrl.productSkuService.GetProductSkuDetail(c.Param("skuId"))
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateProductSku 新增SKU
// @Summary 新增SKU
// @Description 新增SKU记录
// @Tags 产品档案/SKU
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveProductSkuRequest true "SKU"
// @Success 200 {object} models.Response{data=models.ProductSkuResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus [post]
func (ctrl *ProductSkuController) CreateProductSku(c *gin.Context) {
	var req models.SaveProductSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuService.CreateProductSku(req, productSkuOperatorID(c))
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSku 更新SKU
// @Summary 更新SKU
// @Description 根据SKU ID更新SKU信息
// @Tags 产品档案/SKU
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param request body models.SaveProductSkuRequest true "SKU"
// @Success 200 {object} models.Response{data=models.ProductSkuResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId} [put]
func (ctrl *ProductSkuController) UpdateProductSku(c *gin.Context) {
	var req models.SaveProductSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuService.UpdateProductSku(c.Param("skuId"), req, productSkuOperatorID(c))
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSkuStatus 更新SKU启停状态
// @Summary 更新SKU启停状态
// @Description 更新SKU的启用/停用状态
// @Tags 产品档案/SKU
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param request body models.UpdateProductSkuStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ProductSkuResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/status [put]
func (ctrl *ProductSkuController) UpdateProductSkuStatus(c *gin.Context) {
	var req models.UpdateProductSkuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuService.UpdateProductSkuStatus(c.Param("skuId"), req, productSkuOperatorID(c))
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductSkuOptions 获取启用SKU选项
// @Summary 获取启用SKU选项
// @Description 获取已启用的SKU选项列表，用于下拉选择等场景
// @Tags 产品档案/SKU
// @Produce json
// @Security ApiKeyAuth
// @Param mpId query string false "所属厂家产品ID"
// @Param keyword query string false "关键字"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ProductSkuOptionResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/options [get]
func (ctrl *ProductSkuController) GetProductSkuOptions(c *gin.Context) {
	var req models.ProductSkuOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuService.GetProductSkuOptions(req)
	if err != nil {
		writeProductSkuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func productSkuOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeProductSkuError(c *gin.Context, err error) {
	message := "SKU操作失败"
	switch {
	case errors.Is(err, services.ErrProductSkuInvalidInput),
		errors.Is(err, services.ErrProductSkuNotFound),
		errors.Is(err, services.ErrProductSkuConflict):
		message = err.Error()
	default:
		log.Printf("SKU操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

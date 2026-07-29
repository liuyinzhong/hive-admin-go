package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ProductSkuPriceController struct {
	productSkuPriceService *services.ProductSkuPriceService
}

func NewProductSkuPriceController() *ProductSkuPriceController {
	return &ProductSkuPriceController{
		productSkuPriceService: services.NewProductSkuPriceService(),
	}
}

// GetProductSkuPriceList 获取SKU价格列表
// @Summary 获取SKU价格列表
// @Description 获取指定SKU下全部未删除价格资料，按生效开始时间倒序和创建时间倒序展示
// @Tags 产品档案/SKU价格
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Success 200 {object} models.Response{data=[]models.ProductSkuPriceResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/prices [get]
func (ctrl *ProductSkuPriceController) GetProductSkuPriceList(c *gin.Context) {
	result, err := ctrl.productSkuPriceService.GetProductSkuPriceList(c.Param("skuId"))
	if err != nil {
		writeProductSkuPriceError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateProductSkuPrice 新增SKU价格
// @Summary 新增SKU价格
// @Description 为指定SKU新增价格资料
// @Tags 产品档案/SKU价格
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param request body models.SaveProductSkuPriceRequest true "SKU价格"
// @Success 200 {object} models.Response{data=models.ProductSkuPriceResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/prices [post]
func (ctrl *ProductSkuPriceController) CreateProductSkuPrice(c *gin.Context) {
	var req models.SaveProductSkuPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuPriceService.CreateProductSkuPrice(c.Param("skuId"), req, productSkuPriceOperatorID(c))
	if err != nil {
		writeProductSkuPriceError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSkuPrice 更新SKU价格
// @Summary 更新SKU价格
// @Description 根据价格ID更新指定SKU下的价格资料
// @Tags 产品档案/SKU价格
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param priceId path string true "价格ID"
// @Param request body models.SaveProductSkuPriceRequest true "SKU价格"
// @Success 200 {object} models.Response{data=models.ProductSkuPriceResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/prices/{priceId} [put]
func (ctrl *ProductSkuPriceController) UpdateProductSkuPrice(c *gin.Context) {
	var req models.SaveProductSkuPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuPriceService.UpdateProductSkuPrice(c.Param("skuId"), c.Param("priceId"), req, productSkuPriceOperatorID(c))
	if err != nil {
		writeProductSkuPriceError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSkuPriceStatus 更新SKU价格状态
// @Summary 更新SKU价格状态
// @Description 更新指定SKU价格的启用/停用状态，启用时重新校验生效期重叠
// @Tags 产品档案/SKU价格
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param priceId path string true "价格ID"
// @Param request body models.UpdateProductSkuPriceStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ProductSkuPriceResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/prices/{priceId}/status [put]
func (ctrl *ProductSkuPriceController) UpdateProductSkuPriceStatus(c *gin.Context) {
	var req models.UpdateProductSkuPriceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSkuPriceService.UpdateProductSkuPriceStatus(c.Param("skuId"), c.Param("priceId"), req, productSkuPriceOperatorID(c))
	if err != nil {
		writeProductSkuPriceError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteProductSkuPrice 删除SKU价格
// @Summary 删除SKU价格
// @Description 软删除指定SKU价格资料
// @Tags 产品档案/SKU价格
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param skuId path string true "SKU ID"
// @Param priceId path string true "价格ID"
// @Param request body models.DeleteProductSkuPriceRequest true "删除参数"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 404 {object} models.Response "业务数据不存在"
// @Failure 409 {object} models.Response "业务数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/skus/{skuId}/prices/{priceId} [delete]
func (ctrl *ProductSkuPriceController) DeleteProductSkuPrice(c *gin.Context) {
	var req models.DeleteProductSkuPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.productSkuPriceService.DeleteProductSkuPrice(c.Param("skuId"), c.Param("priceId"), req, productSkuPriceOperatorID(c)); err != nil {
		writeProductSkuPriceError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func productSkuPriceOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeProductSkuPriceError(c *gin.Context, err error) {
	message := "SKU价格操作失败"
	switch {
	case errors.Is(err, services.ErrProductSkuPriceInvalidInput):
		message = err.Error()
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrProductSkuPriceNotFound):
		message = err.Error()
		c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, message))
	case errors.Is(err, services.ErrProductSkuPriceConflict):
		message = err.Error()
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, message))
	default:
		log.Printf("SKU价格操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
}

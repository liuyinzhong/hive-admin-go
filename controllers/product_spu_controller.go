package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ProductSpuController struct {
	productSpuService *services.ProductSpuService
}

func NewProductSpuController() *ProductSpuController {
	return &ProductSpuController{
		productSpuService: services.NewProductSpuService(),
	}
}

// GetProductSpuList 获取通用产品列表
// @Summary 获取通用产品列表
// @Tags 产品档案/通用产品
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "编码、通用名称或简称"
// @Param productType query string false "产品类型"
// @Param status query int false "状态：0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse}
// @Router /product/spus [get]
func (ctrl *ProductSpuController) GetProductSpuList(c *gin.Context) {
	var req models.ProductSpuListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSpuService.GetProductSpuList(req)
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductSpu 获取通用产品详情
// @Summary 获取通用产品详情
// @Tags 产品档案/通用产品
// @Produce json
// @Security ApiKeyAuth
// @Param spuId path string true "通用产品ID"
// @Success 200 {object} models.Response{data=models.ProductSpuResponse}
// @Router /product/spus/{spuId} [get]
func (ctrl *ProductSpuController) GetProductSpu(c *gin.Context) {
	result, err := ctrl.productSpuService.GetProductSpuDetail(c.Param("spuId"))
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateProductSpu 新增通用产品
// @Summary 新增通用产品
// @Tags 产品档案/通用产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveProductSpuRequest true "通用产品"
// @Success 200 {object} models.Response{data=models.ProductSpuResponse}
// @Router /product/spus [post]
func (ctrl *ProductSpuController) CreateProductSpu(c *gin.Context) {
	var req models.SaveProductSpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSpuService.CreateProductSpu(req, productSpuOperatorID(c))
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSpu 更新通用产品
// @Summary 更新通用产品
// @Tags 产品档案/通用产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param spuId path string true "通用产品ID"
// @Param request body models.SaveProductSpuRequest true "通用产品"
// @Success 200 {object} models.Response{data=models.ProductSpuResponse}
// @Router /product/spus/{spuId} [put]
func (ctrl *ProductSpuController) UpdateProductSpu(c *gin.Context) {
	var req models.SaveProductSpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSpuService.UpdateProductSpu(c.Param("spuId"), req, productSpuOperatorID(c))
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductSpuStatus 更新通用产品启停状态
// @Summary 更新通用产品启停状态
// @Tags 产品档案/通用产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param spuId path string true "通用产品ID"
// @Param request body models.UpdateProductSpuStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ProductSpuResponse}
// @Router /product/spus/{spuId}/status [put]
func (ctrl *ProductSpuController) UpdateProductSpuStatus(c *gin.Context) {
	var req models.UpdateProductSpuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSpuService.UpdateProductSpuStatus(c.Param("spuId"), req, productSpuOperatorID(c))
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductSpuOptions 获取启用通用产品选项
// @Summary 获取启用通用产品选项
// @Tags 产品档案/通用产品
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "编码、通用名称或简称"
// @Param productType query string false "产品类型"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ProductSpuOptionResponse}
// @Router /product/spus/options [get]
func (ctrl *ProductSpuController) GetProductSpuOptions(c *gin.Context) {
	var req models.ProductSpuOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productSpuService.GetProductSpuOptions(req)
	if err != nil {
		writeProductSpuError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func productSpuOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeProductSpuError(c *gin.Context, err error) {
	message := "通用产品操作失败"
	switch {
	case errors.Is(err, services.ErrProductSpuInvalidInput),
		errors.Is(err, services.ErrProductSpuNotFound),
		errors.Is(err, services.ErrProductSpuConflict):
		message = err.Error()
	default:
		log.Printf("通用产品操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

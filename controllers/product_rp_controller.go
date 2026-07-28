package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ProductRpController struct {
	productRpService *services.ProductRpService
}

func NewProductRpController() *ProductRpController {
	return &ProductRpController{
		productRpService: services.NewProductRpService(),
	}
}

// GetProductRpList 获取规格产品列表
// @Summary 获取规格产品列表
// @Description 分页查询规格产品列表，支持按编码、名称、剂型、规格、SPU、产品类型、状态筛选及排序
// @Tags 产品档案/规格产品
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "RP编码、规格名称、剂型、规格文本、SPU编码或通用名称"
// @Param spuId query string false "所属通用产品ID"
// @Param productType query string false "产品类型"
// @Param status query int false "状态：0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ProductRpResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps [get]
func (ctrl *ProductRpController) GetProductRpList(c *gin.Context) {
	var req models.ProductRpListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productRpService.GetProductRpList(req)
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductRp 获取规格产品详情
// @Summary 获取规格产品详情
// @Description 根据RP ID查询规格产品详细信息
// @Tags 产品档案/规格产品
// @Produce json
// @Security ApiKeyAuth
// @Param rpId path string true "规格产品ID"
// @Success 200 {object} models.Response{data=models.ProductRpResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps/{rpId} [get]
func (ctrl *ProductRpController) GetProductRp(c *gin.Context) {
	result, err := ctrl.productRpService.GetProductRpDetail(c.Param("rpId"))
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateProductRp 新增规格产品
// @Summary 新增规格产品
// @Description 新增规格产品记录
// @Tags 产品档案/规格产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveProductRpRequest true "规格产品"
// @Success 200 {object} models.Response{data=models.ProductRpResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps [post]
func (ctrl *ProductRpController) CreateProductRp(c *gin.Context) {
	var req models.SaveProductRpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productRpService.CreateProductRp(req, productRpOperatorID(c))
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductRp 更新规格产品
// @Summary 更新规格产品
// @Description 根据RP ID更新规格产品信息
// @Tags 产品档案/规格产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rpId path string true "规格产品ID"
// @Param request body models.SaveProductRpRequest true "规格产品"
// @Success 200 {object} models.Response{data=models.ProductRpResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps/{rpId} [put]
func (ctrl *ProductRpController) UpdateProductRp(c *gin.Context) {
	var req models.SaveProductRpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productRpService.UpdateProductRp(c.Param("rpId"), req, productRpOperatorID(c))
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductRpStatus 更新规格产品启停状态
// @Summary 更新规格产品启停状态
// @Description 更新规格产品的启用/停用状态
// @Tags 产品档案/规格产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rpId path string true "规格产品ID"
// @Param request body models.UpdateProductRpStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ProductRpResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps/{rpId}/status [put]
func (ctrl *ProductRpController) UpdateProductRpStatus(c *gin.Context) {
	var req models.UpdateProductRpStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productRpService.UpdateProductRpStatus(c.Param("rpId"), req, productRpOperatorID(c))
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductRpOptions 获取启用规格产品选项
// @Summary 获取启用规格产品选项
// @Description 获取已启用的规格产品选项列表，用于下拉选择等场景
// @Tags 产品档案/规格产品
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "RP编码、规格名称、剂型、规格文本、SPU编码或通用名称"
// @Param spuId query string false "所属通用产品ID"
// @Param productType query string false "产品类型"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ProductRpOptionResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /product/rps/options [get]
func (ctrl *ProductRpController) GetProductRpOptions(c *gin.Context) {
	var req models.ProductRpOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productRpService.GetProductRpOptions(req)
	if err != nil {
		writeProductRpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func productRpOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeProductRpError(c *gin.Context, err error) {
	message := "规格产品操作失败"
	switch {
	case errors.Is(err, services.ErrProductRpInvalidInput),
		errors.Is(err, services.ErrProductRpNotFound),
		errors.Is(err, services.ErrProductRpConflict):
		message = err.Error()
	default:
		log.Printf("规格产品操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

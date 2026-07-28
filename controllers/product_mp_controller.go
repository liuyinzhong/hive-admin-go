package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type ProductMpController struct {
	productMpService *services.ProductMpService
}

func NewProductMpController() *ProductMpController {
	return &ProductMpController{
		productMpService: services.NewProductMpService(),
	}
}

// GetProductMpList 获取厂家产品列表
// @Summary 获取厂家产品列表
// @Tags 产品档案/厂家产品
// @Produce json
// @Security ApiKeyAuth
// @Param rpId query string true "所属规格产品ID"
// @Success 200 {object} models.Response{data=utils.PaginationResponse}
// @Router /product/mps [get]
func (ctrl *ProductMpController) GetProductMpList(c *gin.Context) {
	var req models.ProductMpListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productMpService.GetProductMpList(req)
	if err != nil {
		writeProductMpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetProductMp 获取厂家产品详情
// @Summary 获取厂家产品详情
// @Tags 产品档案/厂家产品
// @Produce json
// @Security ApiKeyAuth
// @Param mpId path string true "厂家产品ID"
// @Success 200 {object} models.Response{data=models.ProductMpResponse}
// @Router /product/mps/{mpId} [get]
func (ctrl *ProductMpController) GetProductMp(c *gin.Context) {
	result, err := ctrl.productMpService.GetProductMpDetail(c.Param("mpId"))
	if err != nil {
		writeProductMpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateProductMp 新增厂家产品
// @Summary 新增厂家产品
// @Tags 产品档案/厂家产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveProductMpRequest true "厂家产品"
// @Success 200 {object} models.Response{data=models.ProductMpResponse}
// @Router /product/mps [post]
func (ctrl *ProductMpController) CreateProductMp(c *gin.Context) {
	var req models.SaveProductMpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productMpService.CreateProductMp(req, productMpOperatorID(c))
	if err != nil {
		writeProductMpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductMp 更新厂家产品
// @Summary 更新厂家产品
// @Tags 产品档案/厂家产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param mpId path string true "厂家产品ID"
// @Param request body models.SaveProductMpRequest true "厂家产品"
// @Success 200 {object} models.Response{data=models.ProductMpResponse}
// @Router /product/mps/{mpId} [put]
func (ctrl *ProductMpController) UpdateProductMp(c *gin.Context) {
	var req models.SaveProductMpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productMpService.UpdateProductMp(c.Param("mpId"), req, productMpOperatorID(c))
	if err != nil {
		writeProductMpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateProductMpStatus 更新厂家产品启停状态
// @Summary 更新厂家产品启停状态
// @Tags 产品档案/厂家产品
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param mpId path string true "厂家产品ID"
// @Param request body models.UpdateProductMpStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ProductMpResponse}
// @Router /product/mps/{mpId}/status [put]
func (ctrl *ProductMpController) UpdateProductMpStatus(c *gin.Context) {
	var req models.UpdateProductMpStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.productMpService.UpdateProductMpStatus(c.Param("mpId"), req, productMpOperatorID(c))
	if err != nil {
		writeProductMpError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func productMpOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeProductMpError(c *gin.Context, err error) {
	message := "厂家产品操作失败"
	switch {
	case errors.Is(err, services.ErrProductMpInvalidInput),
		errors.Is(err, services.ErrProductMpNotFound),
		errors.Is(err, services.ErrProductMpConflict):
		message = err.Error()
	default:
		log.Printf("厂家产品操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

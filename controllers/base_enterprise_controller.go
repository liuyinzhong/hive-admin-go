package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type BaseEnterpriseController struct {
	enterpriseService *services.BaseEnterpriseService
}

func NewBaseEnterpriseController() *BaseEnterpriseController {
	return &BaseEnterpriseController{
		enterpriseService: services.NewBaseEnterpriseService(),
	}
}

// GetEnterpriseList 获取企业主体列表
// @Summary 获取企业主体列表
// @Description 分页查询企业主体列表，支持按名称、简称、编码、信用代码、企业类型、角色类型和状态进行筛选
// @Tags 基础资料/企业主体
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "名称、简称、编码或信用代码"
// @Param enterpriseType query string false "企业类型"
// @Param roleTypes query string false "企业角色，逗号分隔，OR逻辑"
// @Param status query int false "状态：0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.EnterpriseResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises [get]
func (ctrl *BaseEnterpriseController) GetEnterpriseList(c *gin.Context) {
	var req models.EnterpriseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.enterpriseService.GetEnterpriseList(req)
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetEnterprise 获取企业主体详情
// @Summary 获取企业主体详情
// @Description 根据企业主体ID获取企业主体的详细信息
// @Tags 基础资料/企业主体
// @Produce json
// @Security ApiKeyAuth
// @Param enterpriseId path string true "企业主体ID"
// @Success 200 {object} models.Response{data=models.EnterpriseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises/{enterpriseId} [get]
func (ctrl *BaseEnterpriseController) GetEnterprise(c *gin.Context) {
	result, err := ctrl.enterpriseService.GetEnterpriseDetail(c.Param("enterpriseId"))
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateEnterprise 新增企业主体
// @Summary 新增企业主体
// @Description 创建新的企业主体记录
// @Tags 基础资料/企业主体
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveEnterpriseRequest true "企业主体"
// @Success 200 {object} models.Response{data=models.EnterpriseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises [post]
func (ctrl *BaseEnterpriseController) CreateEnterprise(c *gin.Context) {
	var req models.SaveEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.enterpriseService.CreateEnterprise(req, baseOperatorID(c))
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateEnterprise 更新企业主体
// @Summary 更新企业主体
// @Description 根据企业主体ID更新企业主体信息
// @Tags 基础资料/企业主体
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param enterpriseId path string true "企业主体ID"
// @Param request body models.SaveEnterpriseRequest true "企业主体"
// @Success 200 {object} models.Response{data=models.EnterpriseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises/{enterpriseId} [put]
func (ctrl *BaseEnterpriseController) UpdateEnterprise(c *gin.Context) {
	var req models.SaveEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.enterpriseService.UpdateEnterprise(c.Param("enterpriseId"), req, baseOperatorID(c))
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateEnterpriseStatus 更新企业主体启停状态
// @Summary 更新企业主体启停状态
// @Description 根据企业主体ID更新企业主体的启用/停用状态
// @Tags 基础资料/企业主体
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param enterpriseId path string true "企业主体ID"
// @Param request body models.UpdateEnterpriseStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.EnterpriseResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises/{enterpriseId}/status [put]
func (ctrl *BaseEnterpriseController) UpdateEnterpriseStatus(c *gin.Context) {
	var req models.UpdateEnterpriseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.enterpriseService.UpdateEnterpriseStatus(c.Param("enterpriseId"), req, baseOperatorID(c))
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetEnterpriseOptions 获取启用企业主体选项
// @Summary 获取启用企业主体选项
// @Description 获取已启用的企业主体选项列表，用于下拉选择等场景
// @Tags 基础资料/企业主体
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "名称、简称、编码或信用代码"
// @Param roleType query string false "企业角色"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.EnterpriseOptionResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/enterprises/options [get]
func (ctrl *BaseEnterpriseController) GetEnterpriseOptions(c *gin.Context) {
	var req models.EnterpriseOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.enterpriseService.GetEnterpriseOptions(req)
	if err != nil {
		writeBaseEnterpriseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func baseOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeBaseEnterpriseError(c *gin.Context, err error) {
	message := "企业主体操作失败"
	switch {
	case errors.Is(err, services.ErrBaseEnterpriseInvalidInput),
		errors.Is(err, services.ErrBaseEnterpriseNotFound),
		errors.Is(err, services.ErrBaseEnterpriseConflict):
		message = err.Error()
	default:
		log.Printf("企业主体操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

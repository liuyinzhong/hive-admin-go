package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetRegistrationFeeRuleList 获取挂号费规则列表。
// @Summary 获取挂号费规则列表
// @Description 分页查询挂号费规则列表，支持按医生、科室、挂号类型、有效期状态等条件筛选
// @Tags 医疗管理/挂号费
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param keyword query string false "医生或科室关键词"
// @Param doctorId query string false "医生ID"
// @Param departmentId query string false "临床科室ID"
// @Param registrationType query string false "挂号类型字典值"
// @Param periodStatus query string false "有效期状态 current/future/expired"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.RegistrationFeeRuleResponse}}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/registrationFeeRules [get]
func (ctrl *MedicalController) GetRegistrationFeeRuleList(c *gin.Context) {
	var req models.RegistrationFeeRuleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.registrationFeeService.GetRegistrationFeeRuleList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateRegistrationFeeRule 创建挂号费规则。
// @Summary 创建挂号费规则
// @Description 创建挂号费规则，支持配置费用、生效日期和失效日期
// @Tags 医疗管理/挂号费
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateRegistrationFeeRuleRequest true "挂号费规则"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/registrationFeeRules [post]
func (ctrl *MedicalController) CreateRegistrationFeeRule(c *gin.Context) {
	var req models.CreateRegistrationFeeRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.registrationFeeService.CreateRegistrationFeeRule(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// AdjustRegistrationFeeRule 调整挂号费并生成新版本。
// @Summary 调整挂号费并生成新版本
// @Description 对已有挂号费规则进行调价，生成新版本挂号费记录
// @Tags 医疗管理/挂号费
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param feeRuleId path string true "挂号费规则ID"
// @Param request body models.AdjustRegistrationFeeRuleRequest true "调价信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/registrationFeeRules/{feeRuleId}/adjustments [post]
func (ctrl *MedicalController) AdjustRegistrationFeeRule(c *gin.Context) {
	var req models.AdjustRegistrationFeeRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.registrationFeeService.AdjustRegistrationFeeRule(c.Param("feeRuleId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

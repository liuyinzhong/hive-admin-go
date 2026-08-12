package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetDiagnosisList 分页查询疾病诊断档案。
// @Summary 获取疾病诊断档案列表
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param keyword query string false "ICD编码、名称或拼音"
// @Param status query int false "状态 0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.DiagnosisResponse}}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses [get]
func (ctrl *MedicalController) GetDiagnosisList(c *gin.Context) {
	var req models.DiagnosisListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.diagnosisService.GetDiagnosisList(req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetDiagnosisOptions 查询启用疾病诊断选项。
// @Summary 获取疾病诊断选项
// @Tags 医疗管理/疾病诊断档案
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "ICD编码、名称或拼音"
// @Param pageSize query int false "返回数量，最大100"
// @Success 200 {object} models.Response{data=[]models.DiagnosisResponse}
// @Failure 403 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/diagnosisOptions [get]
func (ctrl *MedicalController) GetDiagnosisOptions(c *gin.Context) {
	var req models.DiagnosisOptionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.diagnosisService.GetDiagnosisOptions(req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetDiagnosisDetail 获取疾病诊断详情。
// @Summary 获取疾病诊断详情
// @Tags 医疗管理/疾病诊断档案
// @Produce json
// @Security ApiKeyAuth
// @Param diagnosisId path string true "诊断ID"
// @Success 200 {object} models.Response{data=models.DiagnosisResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses/{diagnosisId} [get]
func (ctrl *MedicalController) GetDiagnosisDetail(c *gin.Context) {
	result, err := ctrl.diagnosisService.GetDiagnosisDetail(c.Param("diagnosisId"))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateDiagnosis 创建疾病诊断档案。
// @Summary 创建疾病诊断档案
// @Tags 医疗管理/疾病诊断档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveDiagnosisRequest true "疾病诊断档案"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses [post]
func (ctrl *MedicalController) CreateDiagnosis(c *gin.Context) {
	var req models.SaveDiagnosisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.diagnosisService.CreateDiagnosis(req, medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateDiagnosis 修改疾病诊断档案。
// @Summary 修改疾病诊断档案
// @Tags 医疗管理/疾病诊断档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param diagnosisId path string true "诊断ID"
// @Param request body models.SaveDiagnosisRequest true "疾病诊断档案"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses/{diagnosisId} [put]
func (ctrl *MedicalController) UpdateDiagnosis(c *gin.Context) {
	var req models.SaveDiagnosisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.diagnosisService.UpdateDiagnosis(c.Param("diagnosisId"), req, medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateDiagnosisStatus 启停疾病诊断档案。
// @Summary 更新疾病诊断状态
// @Tags 医疗管理/疾病诊断档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param diagnosisId path string true "诊断ID"
// @Param request body models.UpdateDiagnosisStatusRequest true "状态"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses/{diagnosisId}/status [put]
func (ctrl *MedicalController) UpdateDiagnosisStatus(c *gin.Context) {
	var req models.UpdateDiagnosisStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.diagnosisService.UpdateDiagnosisStatus(c.Param("diagnosisId"), req.Status, medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteDiagnosis 软删除疾病诊断档案。
// @Summary 删除疾病诊断档案
// @Tags 医疗管理/疾病诊断档案
// @Produce json
// @Security ApiKeyAuth
// @Param diagnosisId path string true "诊断ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/diagnoses/{diagnosisId} [delete]
func (ctrl *MedicalController) DeleteDiagnosis(c *gin.Context) {
	if err := ctrl.diagnosisService.DeleteDiagnosis(c.Param("diagnosisId"), medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

const medicalPatientSensitivePermission = "medical:patient:viewSensitive"

// GetPatientList 获取患者列表。
// @Summary 获取患者列表
// @Description 分页查询患者主档案，列表中的姓名、证件号码和手机号始终脱敏
// @Tags 医疗管理/患者档案
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param keyword query string false "患者编号、姓名、手机号或证件号码"
// @Param gender query string false "性别字典值"
// @Param status query int false "状态 0停用 1启用"
// @Param createDateFrom query string false "创建开始日期，格式YYYY-MM-DD"
// @Param createDateTo query string false "创建结束日期，格式YYYY-MM-DD"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.PatientResponse}}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/patients [get]
func (ctrl *MedicalController) GetPatientList(c *gin.Context) {
	var req models.PatientListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.patientService.GetPatientList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreatePatient 创建患者档案。
// @Summary 创建患者档案
// @Description 创建完整实名患者档案，姓名、证件号码和手机号属于敏感信息
// @Tags 医疗管理/患者档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SavePatientRequest true "患者档案"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无患者敏感信息访问权限"
// @Failure 409 {object} models.Response "患者身份凭证已存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/patients [post]
func (ctrl *MedicalController) CreatePatient(c *gin.Context) {
	if !ctrl.hasPatientSensitivePermission(c) {
		return
	}
	var req models.SavePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.patientService.CreatePatient(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetPatientDetail 获取患者详情。
// @Summary 获取患者详情
// @Description 获取患者主档案详情；拥有敏感信息权限时返回姓名、证件号码和手机号完整值，否则返回脱敏值
// @Tags 医疗管理/患者档案
// @Produce json
// @Security ApiKeyAuth
// @Param patientId path string true "患者ID"
// @Success 200 {object} models.Response{data=models.PatientResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/patients/{patientId} [get]
func (ctrl *MedicalController) GetPatientDetail(c *gin.Context) {
	result, err := ctrl.patientService.GetPatientDetail(
		c.Param("patientId"),
		ctrl.hasPatientSensitivePermissionValue(c),
	)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdatePatient 更新患者档案。
// @Summary 更新患者档案
// @Description 更新患者完整实名档案，需要同时拥有修改权限和敏感信息权限
// @Tags 医疗管理/患者档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param patientId path string true "患者ID"
// @Param request body models.SavePatientRequest true "患者档案"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无患者敏感信息访问权限"
// @Failure 409 {object} models.Response "患者身份凭证已存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/patients/{patientId} [put]
func (ctrl *MedicalController) UpdatePatient(c *gin.Context) {
	if !ctrl.hasPatientSensitivePermission(c) {
		return
	}
	var req models.SavePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.patientService.UpdatePatient(c.Param("patientId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdatePatientStatus 更新患者状态。
// @Summary 更新患者状态
// @Description 启用或停用患者档案，停用不删除档案
// @Tags 医疗管理/患者档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param patientId path string true "患者ID"
// @Param request body models.UpdatePatientStatusRequest true "患者状态"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /medical/patients/{patientId}/status [put]
func (ctrl *MedicalController) UpdatePatientStatus(c *gin.Context) {
	var req models.UpdatePatientStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.patientService.UpdatePatientStatus(c.Param("patientId"), req.Status, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func (ctrl *MedicalController) hasPatientSensitivePermission(c *gin.Context) bool {
	if ctrl.hasPatientSensitivePermissionValue(c) {
		return true
	}
	c.JSON(http.StatusForbidden, models.NewErrorResponse(nil, "无患者敏感信息访问权限"))
	return false
}

func (ctrl *MedicalController) hasPatientSensitivePermissionValue(c *gin.Context) bool {
	return ctrl.permissionService.HasCode(medicalOperatorID(c), medicalPatientSensitivePermission)
}

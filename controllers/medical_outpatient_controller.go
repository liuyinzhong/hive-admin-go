package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetDoctorWorkbench 获取当前医生今日工作台。
// @Summary 获取医生工作台
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.DoctorWorkbenchResponse}
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench [get]
func (ctrl *MedicalController) GetDoctorWorkbench(c *gin.Context) {
	result, err := ctrl.outpatientService.GetWorkbench(medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CallNextPatient 按签到序号叫下一位患者。
// @Summary 叫下一位患者
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param scheduleId path string true "排班ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/schedules/{scheduleId}/callNext [post]
func (ctrl *MedicalController) CallNextPatient(c *gin.Context) {
	if err := ctrl.outpatientService.CallNext(c.Param("scheduleId"), medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// RepeatCallPatient 再次呼叫已叫号患者。
// @Summary 再次叫号
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param queueId path string true "候诊记录ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/queues/{queueId}/repeatCall [post]
func (ctrl *MedicalController) RepeatCallPatient(c *gin.Context) {
	ctrl.workbenchQueueAction(c, ctrl.outpatientService.RepeatCall)
}

// PassPatient 将已叫号患者标记为过号。
// @Summary 患者过号
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param queueId path string true "候诊记录ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/queues/{queueId}/pass [post]
func (ctrl *MedicalController) PassPatient(c *gin.Context) {
	ctrl.workbenchQueueAction(c, ctrl.outpatientService.PassQueue)
}

// RecallPatient 重新呼叫已过号患者。
// @Summary 重新叫号
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param queueId path string true "候诊记录ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/queues/{queueId}/recall [post]
func (ctrl *MedicalController) RecallPatient(c *gin.Context) {
	ctrl.workbenchQueueAction(c, ctrl.outpatientService.RecallQueue)
}

// StartConsultation 开始正式接诊并创建门诊病历。
// @Summary 开始接诊
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param queueId path string true "候诊记录ID"
// @Success 200 {object} models.Response{data=models.OutpatientRecordResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/doctorWorkbench/queues/{queueId}/start [post]
func (ctrl *MedicalController) StartConsultation(c *gin.Context) {
	result, err := ctrl.outpatientService.StartConsultation(c.Param("queueId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetOutpatientRecord 获取当前医生的门诊病历。
// @Summary 获取门诊病历
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "门诊病历ID"
// @Success 200 {object} models.Response{data=models.OutpatientRecordResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/outpatientRecords/{recordId} [get]
func (ctrl *MedicalController) GetOutpatientRecord(c *gin.Context) {
	result, err := ctrl.outpatientService.GetOutpatientRecord(c.Param("recordId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// SaveOutpatientRecord 保存接诊中的门诊病历。
// @Summary 保存门诊病历
// @Tags 医疗管理/医生工作台
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "门诊病历ID"
// @Param request body models.SaveOutpatientRecordRequest true "门诊病历"
// @Success 200 {object} models.Response{data=models.OutpatientRecordResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/outpatientRecords/{recordId} [put]
func (ctrl *MedicalController) SaveOutpatientRecord(c *gin.Context) {
	var req models.SaveOutpatientRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.outpatientService.SaveOutpatientRecord(c.Param("recordId"), medicalOperatorID(c), req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CompleteOutpatientRecord 完成正式接诊。
// @Summary 完成接诊
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "门诊病历ID"
// @Success 200 {object} models.Response{data=models.OutpatientRecordResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/outpatientRecords/{recordId}/complete [post]
func (ctrl *MedicalController) CompleteOutpatientRecord(c *gin.Context) {
	result, err := ctrl.outpatientService.CompleteOutpatientRecord(c.Param("recordId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPatientOutpatientHistory 获取当前接诊患者的历史门诊病历。
// @Summary 获取患者历史病历
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "当前门诊病历ID"
// @Success 200 {object} models.Response{data=[]models.OutpatientRecordResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/outpatientRecords/{recordId}/history [get]
func (ctrl *MedicalController) GetPatientOutpatientHistory(c *gin.Context) {
	result, err := ctrl.outpatientService.GetPatientHistory(c.Param("recordId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreatePrescription 创建普通处方草稿。
// @Summary 创建处方草稿
// @Tags 医疗管理/医生工作台
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "门诊病历ID"
// @Param request body models.SavePrescriptionRequest true "处方"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/outpatientRecords/{recordId}/prescriptions [post]
func (ctrl *MedicalController) CreatePrescription(c *gin.Context) {
	var req models.SavePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.prescriptionService.CreatePrescription(c.Param("recordId"), medicalOperatorID(c), req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPrescription 获取处方详情。
// @Summary 获取处方详情
// @Tags 医疗管理/处方
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptions/{prescriptionId} [get]
func (ctrl *MedicalController) GetPrescription(c *gin.Context) {
	result, err := ctrl.prescriptionService.GetDoctorPrescription(c.Param("prescriptionId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdatePrescription 修改处方草稿或审核不通过处方。
// @Summary 修改处方
// @Tags 医疗管理/医生工作台
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Param request body models.SavePrescriptionRequest true "处方"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptions/{prescriptionId} [put]
func (ctrl *MedicalController) UpdatePrescription(c *gin.Context) {
	var req models.SavePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.prescriptionService.UpdatePrescription(c.Param("prescriptionId"), medicalOperatorID(c), req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// SubmitPrescription 提交处方审核并固化提交快照。
// @Summary 提交处方审核
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptions/{prescriptionId}/submit [post]
func (ctrl *MedicalController) SubmitPrescription(c *gin.Context) {
	ctrl.prescriptionAction(c, ctrl.prescriptionService.SubmitPrescription)
}

// WithdrawPrescription 撤回尚未审核的处方。
// @Summary 撤回处方
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptions/{prescriptionId}/withdraw [post]
func (ctrl *MedicalController) WithdrawPrescription(c *gin.Context) {
	ctrl.prescriptionAction(c, ctrl.prescriptionService.WithdrawPrescription)
}

// VoidPrescription 作废草稿或审核不通过处方。
// @Summary 作废处方
// @Tags 医疗管理/医生工作台
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptions/{prescriptionId}/void [post]
func (ctrl *MedicalController) VoidPrescription(c *gin.Context) {
	ctrl.prescriptionAction(c, ctrl.prescriptionService.VoidPrescription)
}

// GetPrescriptionReviewList 查询处方审核列表。
// @Summary 获取处方审核列表
// @Tags 医疗管理/处方审核
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param prescriptionNo query string false "处方编号"
// @Param patientKeyword query string false "患者编号或姓名"
// @Param status query int false "处方状态"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.PrescriptionResponse}}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptionReviews [get]
func (ctrl *MedicalController) GetPrescriptionReviewList(c *gin.Context) {
	var req models.PrescriptionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.prescriptionService.GetReviewList(req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetPrescriptionReviewDetail 获取审核用处方详情。
// @Summary 获取处方审核详情
// @Tags 医疗管理/处方审核
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptionReviews/{prescriptionId} [get]
func (ctrl *MedicalController) GetPrescriptionReviewDetail(c *gin.Context) {
	result, err := ctrl.prescriptionService.GetReviewPrescription(c.Param("prescriptionId"))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ReviewPrescription 审核处方。
// @Summary 审核处方
// @Tags 医疗管理/处方审核
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param prescriptionId path string true "处方ID"
// @Param request body models.ReviewPrescriptionRequest true "审核结果"
// @Success 200 {object} models.Response{data=models.PrescriptionResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/prescriptionReviews/{prescriptionId}/review [post]
func (ctrl *MedicalController) ReviewPrescription(c *gin.Context) {
	var req models.ReviewPrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.prescriptionService.ReviewPrescription(c.Param("prescriptionId"), medicalOperatorID(c), req)
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

type workbenchQueueActionFunc func(string, string) error
type prescriptionActionFunc func(string, string) (*models.PrescriptionResponse, error)

func (ctrl *MedicalController) workbenchQueueAction(c *gin.Context, action workbenchQueueActionFunc) {
	if err := action(c.Param("queueId"), medicalOperatorID(c)); err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func (ctrl *MedicalController) prescriptionAction(c *gin.Context, action prescriptionActionFunc) {
	result, err := action(c.Param("prescriptionId"), medicalOperatorID(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

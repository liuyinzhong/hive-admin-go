package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetRegistrationList 分页查询挂号单。
// @Summary 获取挂号单列表
// @Description 按挂号单创建人及当前角色数据范围分页查询；患者敏感字段权限另行叠加
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param registrationNo query string false "挂号单号"
// @Param patientKeyword query string false "患者编号、姓名或手机号"
// @Param startDate query string false "就诊开始日期，YYYY-MM-DD"
// @Param endDate query string false "就诊结束日期，YYYY-MM-DD"
// @Param departmentId query string false "科室ID"
// @Param doctorId query string false "医生ID"
// @Param registrationType query string false "挂号类型"
// @Param registrationMethod query int false "挂号方式 0现场 10预约"
// @Param status query int false "挂号状态"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.RegistrationResponse}}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations [get]
func (ctrl *MedicalController) GetRegistrationList(c *gin.Context) {
	var req models.RegistrationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.registrationService.GetRegistrationList(req, ctrl.hasPatientSensitivePermissionValue(c), currentDataPermission(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateRegistration 创建挂号单。
// @Summary 创建挂号单
// @Description 患者和实际排班必须处于当前数据范围；新挂号单以当前用户作为归属人
// @Tags 医疗管理/挂号管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateRegistrationRequest true "挂号信息"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations [post]
func (ctrl *MedicalController) CreateRegistration(c *gin.Context) {
	var req models.CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.registrationService.CreateRegistration(req, medicalOperatorID(c), ctrl.hasPatientSensitivePermissionValue(c), currentDataPermission(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetRegistrationDetail 获取挂号单详情及生命周期。
// @Summary 获取挂号单详情
// @Description 按挂号单创建人及当前角色数据范围获取详情和生命周期；患者敏感字段权限另行叠加
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId} [get]
func (ctrl *MedicalController) GetRegistrationDetail(c *gin.Context) {
	result, err := ctrl.registrationService.GetRegistrationDetail(c.Param("registrationId"), ctrl.hasPatientSensitivePermissionValue(c), currentDataPermission(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ConfirmRegistrationPayment 确认支付。
// @Summary 确认挂号单支付
// @Description 按挂号单创建人及当前角色数据范围确认支付
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/confirmPayment [post]
func (ctrl *MedicalController) ConfirmRegistrationPayment(c *gin.Context) {
	ctrl.registrationAction(c, ctrl.registrationService.ConfirmPayment)
}

// CancelRegistration 取消待支付挂号单。
// @Summary 取消挂号单
// @Description 按挂号单创建人及当前角色数据范围取消待支付挂号单
// @Tags 医疗管理/挂号管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Param request body models.RegistrationReasonRequest true "取消原因"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/cancel [post]
func (ctrl *MedicalController) CancelRegistration(c *gin.Context) {
	ctrl.registrationReasonAction(c, ctrl.registrationService.Cancel)
}

// CheckInRegistration 签到。
// @Summary 挂号单签到
// @Description 按挂号单创建人及当前角色数据范围签到
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/checkIn [post]
func (ctrl *MedicalController) CheckInRegistration(c *gin.Context) {
	ctrl.registrationAction(c, ctrl.registrationService.CheckIn)
}

// MarkRegistrationNoShow 标记爽约。
// @Summary 标记挂号单爽约
// @Description 按挂号单创建人及当前角色数据范围标记爽约
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/noShow [post]
func (ctrl *MedicalController) MarkRegistrationNoShow(c *gin.Context) {
	ctrl.registrationAction(c, ctrl.registrationService.NoShow)
}

// StartRegistrationRefund 发起全额退款。
// @Summary 发起挂号单退款
// @Description 按挂号单创建人及当前角色数据范围发起全额退款
// @Tags 医疗管理/挂号管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Param request body models.RegistrationReasonRequest true "退款原因"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/refundStart [post]
func (ctrl *MedicalController) StartRegistrationRefund(c *gin.Context) {
	ctrl.registrationReasonAction(c, ctrl.registrationService.StartRefund)
}

// ProcessRegistrationRefund 标记退款中。
// @Summary 推进挂号单退款中
// @Description 按挂号单创建人及当前角色数据范围标记退款处理中
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/refundProcess [post]
func (ctrl *MedicalController) ProcessRegistrationRefund(c *gin.Context) {
	ctrl.registrationAction(c, ctrl.registrationService.ProcessRefund)
}

// CompleteRegistrationRefund 完成退款并释放号源。
// @Summary 完成挂号单退款
// @Description 按挂号单创建人及当前角色数据范围完成退款并释放号源
// @Tags 医疗管理/挂号管理
// @Produce json
// @Security ApiKeyAuth
// @Param registrationId path string true "挂号单ID"
// @Success 200 {object} models.Response{data=models.RegistrationResponse}
// @Failure 400 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /medical/registrations/{registrationId}/refundComplete [post]
func (ctrl *MedicalController) CompleteRegistrationRefund(c *gin.Context) {
	ctrl.registrationAction(c, ctrl.registrationService.CompleteRefund)
}

type registrationActionFunc func(string, string, bool, datapermission.Permission) (*models.RegistrationResponse, error)
type registrationReasonActionFunc func(string, string, string, bool, datapermission.Permission) (*models.RegistrationResponse, error)

func (ctrl *MedicalController) registrationAction(c *gin.Context, action registrationActionFunc) {
	result, err := action(c.Param("registrationId"), medicalOperatorID(c), ctrl.hasPatientSensitivePermissionValue(c), currentDataPermission(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func (ctrl *MedicalController) registrationReasonAction(c *gin.Context, action registrationReasonActionFunc) {
	var req models.RegistrationReasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := action(c.Param("registrationId"), medicalOperatorID(c), req.Reason, ctrl.hasPatientSensitivePermissionValue(c), currentDataPermission(c))
	if err != nil {
		writeRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeRegistrationError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "挂号管理操作失败"
	switch {
	case errors.Is(err, services.ErrMedicalInvalidInput):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, services.ErrMedicalNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, services.ErrMedicalConflict):
		status, message = http.StatusConflict, err.Error()
	default:
		log.Printf("挂号管理操作失败: %v", err)
	}
	c.JSON(status, models.NewErrorResponse(nil, message))
}

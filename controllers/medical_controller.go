package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type MedicalController struct {
	departmentService *services.MedicalDepartmentService
	doctorService     *services.MedicalDoctorService
}

func NewMedicalController() *MedicalController {
	return &MedicalController{
		departmentService: services.NewMedicalDepartmentService(),
		doctorService:     services.NewMedicalDoctorService(),
	}
}

// GetMedicalDepartmentTree 获取临床科室树。
// @Summary 获取临床科室树
// @Tags 医疗管理-临床科室
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "科室编码或名称"
// @Param status query int false "状态 0停用 1启用"
// @Success 200 {object} models.Response{data=[]models.MedicalDepartmentTreeResponse}
// @Router /medical/departments [get]
func (ctrl *MedicalController) GetMedicalDepartmentTree(c *gin.Context) {
	var req models.MedicalDepartmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.departmentService.GetDepartmentTree(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllMedicalDepartments 获取所有启用的临床科室。
// @Summary 获取所有启用的临床科室
// @Tags 医疗管理-临床科室
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.MedicalDepartmentTreeResponse}
// @Router /medical/departments/all [get]
func (ctrl *MedicalController) GetAllMedicalDepartments(c *gin.Context) {
	result, err := ctrl.departmentService.GetAllDepartments()
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateMedicalDepartment 创建临床科室。
// @Summary 创建临床科室
// @Tags 医疗管理-临床科室
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateMedicalDepartmentRequest true "临床科室信息"
// @Success 200 {object} models.Response
// @Router /medical/departments [post]
func (ctrl *MedicalController) CreateMedicalDepartment(c *gin.Context) {
	var req models.CreateMedicalDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.departmentService.CreateDepartment(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetMedicalDepartmentDetail 获取临床科室详情。
// @Summary 获取临床科室详情
// @Tags 医疗管理-临床科室
// @Produce json
// @Security ApiKeyAuth
// @Param departmentId path string true "临床科室ID"
// @Success 200 {object} models.Response{data=models.MedicalDepartmentTreeResponse}
// @Router /medical/departments/{departmentId} [get]
func (ctrl *MedicalController) GetMedicalDepartmentDetail(c *gin.Context) {
	result, err := ctrl.departmentService.GetDepartmentDetail(c.Param("departmentId"))
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateMedicalDepartment 更新临床科室。
// @Summary 更新临床科室
// @Tags 医疗管理-临床科室
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param departmentId path string true "临床科室ID"
// @Param request body models.UpdateMedicalDepartmentRequest true "临床科室信息"
// @Success 200 {object} models.Response
// @Router /medical/departments/{departmentId} [put]
func (ctrl *MedicalController) UpdateMedicalDepartment(c *gin.Context) {
	var req models.UpdateMedicalDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.departmentService.UpdateDepartment(c.Param("departmentId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateMedicalDepartmentStatus 更新临床科室状态。
// @Summary 更新临床科室状态
// @Tags 医疗管理-临床科室
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param departmentId path string true "临床科室ID"
// @Param request body models.UpdateMedicalStatusRequest true "状态"
// @Success 200 {object} models.Response
// @Router /medical/departments/{departmentId}/status [put]
func (ctrl *MedicalController) UpdateMedicalDepartmentStatus(c *gin.Context) {
	var req models.UpdateMedicalStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.departmentService.UpdateDepartmentStatus(c.Param("departmentId"), req.Status, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteMedicalDepartments 删除临床科室。
// @Summary 批量删除临床科室
// @Tags 医疗管理-临床科室
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "临床科室ID列表"
// @Success 200 {object} models.Response
// @Router /medical/departments [delete]
func (ctrl *MedicalController) DeleteMedicalDepartments(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.departmentService.DeleteDepartments(ids, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetDoctorList 获取医生列表。
// @Summary 获取医生列表
// @Tags 医疗管理-医生档案
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param keyword query string false "姓名、拼音或医生编号"
// @Param departmentId query string false "临床科室ID"
// @Param professionalTitle query string false "职称字典值"
// @Param employmentType query string false "用工类型字典值"
// @Param status query int false "状态 0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.DoctorResponse}}
// @Router /medical/doctors [get]
func (ctrl *MedicalController) GetDoctorList(c *gin.Context) {
	var req models.DoctorListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.doctorService.GetDoctorList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllDoctors 获取所有启用医生选项。
// @Summary 获取所有启用医生选项
// @Tags 医疗管理-医生档案
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.DoctorOptionResponse}
// @Router /medical/doctors/all [get]
func (ctrl *MedicalController) GetAllDoctors(c *gin.Context) {
	result, err := ctrl.doctorService.GetAllDoctors()
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateDoctor 创建医生档案。
// @Summary 创建医生档案
// @Tags 医疗管理-医生档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveDoctorRequest true "医生档案"
// @Success 200 {object} models.Response
// @Router /medical/doctors [post]
func (ctrl *MedicalController) CreateDoctor(c *gin.Context) {
	var req models.SaveDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.doctorService.CreateDoctor(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetDoctorDetail 获取医生详情。
// @Summary 获取医生详情
// @Tags 医疗管理-医生档案
// @Produce json
// @Security ApiKeyAuth
// @Param doctorId path string true "医生ID"
// @Success 200 {object} models.Response{data=models.DoctorResponse}
// @Router /medical/doctors/{doctorId} [get]
func (ctrl *MedicalController) GetDoctorDetail(c *gin.Context) {
	result, err := ctrl.doctorService.GetDoctorDetail(c.Param("doctorId"))
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateDoctor 更新医生档案。
// @Summary 更新医生档案
// @Tags 医疗管理-医生档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param doctorId path string true "医生ID"
// @Param request body models.SaveDoctorRequest true "医生档案"
// @Success 200 {object} models.Response
// @Router /medical/doctors/{doctorId} [put]
func (ctrl *MedicalController) UpdateDoctor(c *gin.Context) {
	var req models.SaveDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.doctorService.UpdateDoctor(c.Param("doctorId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateDoctorStatus 更新医生状态。
// @Summary 更新医生状态
// @Tags 医疗管理-医生档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param doctorId path string true "医生ID"
// @Param request body models.UpdateMedicalStatusRequest true "状态"
// @Success 200 {object} models.Response
// @Router /medical/doctors/{doctorId}/status [put]
func (ctrl *MedicalController) UpdateDoctorStatus(c *gin.Context) {
	var req models.UpdateMedicalStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.doctorService.UpdateDoctorStatus(c.Param("doctorId"), req.Status, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteDoctors 删除医生档案。
// @Summary 批量删除医生档案
// @Tags 医疗管理-医生档案
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "医生ID列表"
// @Success 200 {object} models.Response
// @Router /medical/doctors [delete]
func (ctrl *MedicalController) DeleteDoctors(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.doctorService.DeleteDoctors(ids, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func medicalOperatorID(c *gin.Context) string {
	value, exists := c.Get("userId")
	if !exists {
		return ""
	}
	operatorID, _ := value.(string)
	return operatorID
}

func writeMedicalError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "医疗管理操作失败"
	switch {
	case errors.Is(err, services.ErrMedicalInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, services.ErrMedicalNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, services.ErrMedicalConflict):
		status = http.StatusConflict
		message = err.Error()
	default:
		log.Printf("医疗管理操作失败: %v", err)
	}
	c.JSON(status, models.NewErrorResponse(nil, message))
}

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetScheduleTemplateList 获取周期排班模板列表。
// @Summary 获取周期排班模板列表
// @Tags 医疗管理-排班
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param doctorId query string false "医生ID"
// @Param departmentId query string false "临床科室ID"
// @Param weekday query int false "星期 1周一至7周日"
// @Param status query int false "状态 0停用 1启用"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ScheduleTemplateResponse}}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTemplates [get]
func (ctrl *MedicalController) GetScheduleTemplateList(c *gin.Context) {
	var req models.ScheduleTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.scheduleService.GetScheduleTemplateList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateScheduleTemplate 创建周期排班模板。
// @Summary 创建周期排班模板
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateScheduleTemplateRequest true "周期排班模板"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTemplates [post]
func (ctrl *MedicalController) CreateScheduleTemplate(c *gin.Context) {
	var req models.CreateScheduleTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.CreateScheduleTemplate(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateScheduleTemplate 更新周期排班模板。
// @Summary 更新周期排班模板
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "排班模板ID"
// @Param request body models.SaveScheduleTemplateRequest true "周期排班模板"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTemplates/{templateId} [put]
func (ctrl *MedicalController) UpdateScheduleTemplate(c *gin.Context) {
	var req models.SaveScheduleTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.UpdateScheduleTemplate(c.Param("templateId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateScheduleTemplateStatus 更新周期排班模板状态。
// @Summary 更新周期排班模板状态
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "排班模板ID"
// @Param request body models.UpdateMedicalStatusRequest true "状态"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTemplates/{templateId}/status [put]
func (ctrl *MedicalController) UpdateScheduleTemplateStatus(c *gin.Context) {
	var req models.UpdateMedicalStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.UpdateScheduleTemplateStatus(c.Param("templateId"), req.Status, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteScheduleTemplate 删除周期排班模板。
// @Summary 删除周期排班模板
// @Description 逻辑删除周期排班模板，不影响已生成的实际排班
// @Tags 医疗管理-排班
// @Produce json
// @Security ApiKeyAuth
// @Param templateId path string true "排班模板ID"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTemplates/{templateId} [delete]
func (ctrl *MedicalController) DeleteScheduleTemplate(c *gin.Context) {
	if err := ctrl.scheduleService.DeleteScheduleTemplate(c.Param("templateId"), medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetScheduleList 获取实际排班列表。
// @Summary 获取实际排班列表
// @Tags 医疗管理-排班
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大200"
// @Param doctorId query string false "医生ID"
// @Param departmentId query string false "临床科室ID"
// @Param registrationType query string false "挂号类型字典值"
// @Param startDate query string false "开始日期 YYYY-MM-DD"
// @Param endDate query string false "结束日期 YYYY-MM-DD"
// @Param status query int false "状态 0草稿 1已发布 2停诊 3结束"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ScheduleResponse}}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules [get]
func (ctrl *MedicalController) GetScheduleList(c *gin.Context) {
	var req models.ScheduleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.scheduleService.GetScheduleList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateSchedule 手工创建实际排班。
// @Summary 手工创建实际排班
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveScheduleRequest true "实际排班"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules [post]
func (ctrl *MedicalController) CreateSchedule(c *gin.Context) {
	var req models.SaveScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.CreateSchedule(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateSchedule 编辑草稿排班。
// @Summary 编辑草稿排班
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param scheduleId path string true "排班ID"
// @Param request body models.SaveScheduleRequest true "实际排班"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules/{scheduleId} [put]
func (ctrl *MedicalController) UpdateSchedule(c *gin.Context) {
	var req models.SaveScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.UpdateSchedule(c.Param("scheduleId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteDraftSchedules 删除草稿排班。
// @Summary 批量删除草稿排班
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "排班ID列表"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules [delete]
func (ctrl *MedicalController) DeleteDraftSchedules(c *gin.Context) {
	var scheduleIDs []string
	if err := c.ShouldBindJSON(&scheduleIDs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.DeleteDraftSchedules(scheduleIDs, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GenerateSchedules 根据周期模板批量生成未来排班。
// @Summary 批量生成未来排班
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.GenerateSchedulesRequest true "生成范围和幂等键"
// @Success 200 {object} models.Response{data=models.GenerateSchedulesResponse}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules/generate [post]
func (ctrl *MedicalController) GenerateSchedules(c *gin.Context) {
	var req models.GenerateSchedulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.scheduleService.GenerateSchedules(req, medicalOperatorID(c))
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// PublishSchedules 批量发布草稿排班。
// @Summary 批量发布草稿排班
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.PublishSchedulesRequest true "排班ID列表"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules/publish [post]
func (ctrl *MedicalController) PublishSchedules(c *gin.Context) {
	var req models.PublishSchedulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.PublishSchedules(req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// StopSchedule 停诊。
// @Summary 停诊
// @Tags 医疗管理-排班
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param scheduleId path string true "排班ID"
// @Param request body models.StopScheduleRequest true "停诊原因"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules/{scheduleId}/stop [put]
func (ctrl *MedicalController) StopSchedule(c *gin.Context) {
	var req models.StopScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.scheduleService.StopSchedule(c.Param("scheduleId"), req, medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// FinishSchedule 结束已完成出诊的排班。
// @Summary 结束已完成出诊的排班
// @Tags 医疗管理-排班
// @Produce json
// @Security ApiKeyAuth
// @Param scheduleId path string true "排班ID"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/schedules/{scheduleId}/finish [put]
func (ctrl *MedicalController) FinishSchedule(c *gin.Context) {
	if err := ctrl.scheduleService.FinishSchedule(c.Param("scheduleId"), medicalOperatorID(c)); err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetScheduleAutoTaskList 获取自动任务执行记录。
// @Summary 获取自动任务执行记录
// @Tags 医疗管理-排班
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param taskType query string false "任务类型 publish/generate"
// @Param status query int false "状态 0成功 1部分成功 2失败 3执行中"
// @Param startDate query string false "目标周开始日期"
// @Param endDate query string false "目标周结束日期"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ScheduleAutoTaskResponse}}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /medical/scheduleTasks [get]
func (ctrl *MedicalController) GetScheduleAutoTaskList(c *gin.Context) {
	var req models.ScheduleAutoTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.scheduleService.GetScheduleAutoTaskList(req)
	if err != nil {
		writeMedicalError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// StartWorkflowInstance 发起流程实例。
// @Summary 发起流程实例
// @Description 根据已发布流程定义创建运行实例
// @Tags 流程管理/流程运行
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.StartWorkflowInstanceRequest true "流程发起信息"
// @Success 200 {object} models.Response{data=models.WorkflowInstanceResponse} "发起成功"
// @Failure 400 {object} models.Response "参数或流程配置错误"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/instances [post]
func (wc *WorkflowController) StartWorkflowInstance(c *gin.Context) {
	var req models.StartWorkflowInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := services.StartWorkflowInstance(&req, c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWorkflowInstances 获取当前用户发起的流程实例。
// @Summary 获取我发起的流程
// @Tags 流程管理/流程运行
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query string false "实例状态，支持逗号分隔：0,1,2,3"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.WorkflowInstanceResponse}} "获取成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/instances [get]
func (wc *WorkflowController) GetWorkflowInstances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := services.GetWorkflowInstances(page, pageSize, c.GetString("userId"), parseWorkflowStatusList(c.Query("status")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWorkflowInstanceDetail 获取流程实例详情。
// @Summary 获取流程实例详情
// @Tags 流程管理/流程运行
// @Produce json
// @Security ApiKeyAuth
// @Param instanceId path string true "流程实例ID"
// @Success 200 {object} models.Response{data=models.WorkflowInstanceDetailResponse} "获取成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/instances/{instanceId} [get]
func (wc *WorkflowController) GetWorkflowInstanceDetail(c *gin.Context) {
	result, err := services.GetWorkflowInstanceDetail(c.Param("instanceId"), c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CancelWorkflowInstance 撤销流程实例。
// @Summary 撤销流程实例
// @Tags 流程管理/流程运行
// @Produce json
// @Security ApiKeyAuth
// @Param instanceId path string true "流程实例ID"
// @Success 200 {object} models.Response "撤销成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/instances/{instanceId}/cancel [put]
func (wc *WorkflowController) CancelWorkflowInstance(c *gin.Context) {
	if err := services.CancelWorkflowInstance(c.Param("instanceId"), c.GetString("userId")); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetWorkflowTasks 获取当前用户审批任务。
// @Summary 获取我的待办和已办
// @Tags 流程管理/流程任务
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query string false "任务状态，支持逗号分隔：0,1,2,3"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.WorkflowTaskResponse}} "获取成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks [get]
func (wc *WorkflowController) GetWorkflowTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := services.GetWorkflowTasks(page, pageSize, c.GetString("userId"), parseWorkflowStatusList(c.Query("status")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ApproveWorkflowTask 审批通过任务。
// @Summary 审批通过
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskActionRequest true "审批意见"
// @Success 200 {object} models.Response "审批成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/approve [put]
func (wc *WorkflowController) ApproveWorkflowTask(c *gin.Context) {
	var req models.WorkflowTaskActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.ApproveWorkflowTask(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// RejectWorkflowTask 审批驳回任务。
// @Summary 审批驳回
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskActionRequest true "审批意见"
// @Success 200 {object} models.Response "驳回成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/reject [put]
func (wc *WorkflowController) RejectWorkflowTask(c *gin.Context) {
	var req models.WorkflowTaskActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.RejectWorkflowTask(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// TransferWorkflowTask 转交当前用户的待办任务。
// @Summary 转交审批任务
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskTransferRequest true "转交信息"
// @Success 200 {object} models.Response "转交成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/transfer [put]
func (wc *WorkflowController) TransferWorkflowTask(c *gin.Context) {
	var req models.WorkflowTaskTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.TransferWorkflowTask(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// AddWorkflowTaskSign 向当前审批组并行加签。
// @Summary 并行加签
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskAddSignRequest true "加签信息"
// @Success 200 {object} models.Response "加签成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/addSign [put]
func (wc *WorkflowController) AddWorkflowTaskSign(c *gin.Context) {
	var req models.WorkflowTaskAddSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.AddWorkflowTaskSign(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// RemoveWorkflowTaskSign 从当前审批组减签未处理任务。
// @Summary 减签
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskRemoveSignRequest true "减签信息"
// @Success 200 {object} models.Response "减签成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/removeSign [put]
func (wc *WorkflowController) RemoveWorkflowTaskSign(c *gin.Context) {
	var req models.WorkflowTaskRemoveSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.RemoveWorkflowTaskSign(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetWorkflowTaskReturnTargets 查询当前任务可退回的历史审批节点。
// @Summary 查询可退回节点
// @Tags 流程管理/流程任务
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Success 200 {object} models.Response{data=[]models.WorkflowReturnTargetResponse} "获取成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/returnTargets [get]
func (wc *WorkflowController) GetWorkflowTaskReturnTargets(c *gin.Context) {
	result, err := services.GetWorkflowTaskReturnTargets(c.Param("taskId"), c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ReturnWorkflowTask 将当前任务退回历史审批节点。
// @Summary 退回审批任务
// @Tags 流程管理/流程任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.WorkflowTaskReturnRequest true "退回信息"
// @Success 200 {object} models.Response "退回成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/tasks/{taskId}/return [put]
func (wc *WorkflowController) ReturnWorkflowTask(c *gin.Context) {
	var req models.WorkflowTaskReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.ReturnWorkflowTask(c.Param("taskId"), c.GetString("userId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetWorkflowCopies 获取当前用户抄送记录。
// @Summary 获取我的抄送
// @Tags 流程管理/流程抄送
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query string false "已读状态，支持逗号分隔：0,1"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.WorkflowCopyResponse}} "获取成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/copies [get]
func (wc *WorkflowController) GetWorkflowCopies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := services.GetWorkflowCopies(page, pageSize, c.GetString("userId"), parseWorkflowStatusList(c.Query("status")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// ReadWorkflowCopy 标记抄送为已读。
// @Summary 标记抄送已读
// @Tags 流程管理/流程抄送
// @Produce json
// @Security ApiKeyAuth
// @Param copyId path string true "抄送ID"
// @Success 200 {object} models.Response "操作成功"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/copies/{copyId}/read [put]
func (wc *WorkflowController) ReadWorkflowCopy(c *gin.Context) {
	if err := services.ReadWorkflowCopy(c.Param("copyId"), c.GetString("userId")); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

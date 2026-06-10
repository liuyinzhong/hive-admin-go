package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 分页获取任务列表
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param taskNum query int false "任务编号"
// @Param taskTitle query string false "任务标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param taskStatus query string false "任务状态，支持多选：1,2"
// @Param storyId query string false "需求ID"
// @Param sorts query string false "排序参数 排序时仅支持：taskTitle、taskStatus、startDate、endDate 的排序"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks [get]
func (dc *DevController) GetTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	taskNum := 0
	if tnStr := c.Query("taskNum"); tnStr != "" {
		taskNum, _ = strconv.Atoi(tnStr)
	}
	taskStatuses := make([]int, 0)
	if tsStr := strings.TrimSpace(c.Query("taskStatus")); tsStr != "" {
		for _, part := range strings.Split(tsStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			status, err := strconv.Atoi(part)
			if err == nil {
				taskStatuses = append(taskStatuses, status)
			}
		}
	}

	params := map[string]interface{}{
		"taskNum":      taskNum,
		"taskTitle":    c.Query("taskTitle"),
		"projectId":    c.Query("projectId"),
		"versionId":    c.Query("versionId"),
		"taskStatuses": taskStatuses,
		"storyId":      c.Query("storyId"),
		"sorts":        c.Query("sorts"),
	}

	result, err := services.GetTasks(page, pageSize, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllTasks 获取所有任务
// @Summary 获取所有任务
// @Description 获取所有任务（不分页）
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskNum query int false "任务编号"
// @Param taskTitle query string false "任务标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param taskStatus query int false "任务状态"
// @Param storyId query string false "需求ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/all [get]
func (dc *DevController) GetAllTasks(c *gin.Context) {
	taskNum := 0
	if tnStr := c.Query("taskNum"); tnStr != "" {
		taskNum, _ = strconv.Atoi(tnStr)
	}
	taskStatus := -1
	if tsStr := c.Query("taskStatus"); tsStr != "" {
		taskStatus, _ = strconv.Atoi(tsStr)
	}

	params := map[string]interface{}{
		"taskNum":    taskNum,
		"taskTitle":  c.Query("taskTitle"),
		"projectId":  c.Query("projectId"),
		"versionId":  c.Query("versionId"),
		"taskStatus": taskStatus,
		"storyId":    c.Query("storyId"),
	}

	tasks, err := services.GetAllTasks(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(tasks))
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 根据任务编号获取任务详情
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskNum path int true "任务编号"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/{taskNum} [get]
func (dc *DevController) GetTask(c *gin.Context) {
	taskNum, err := strconv.Atoi(c.Param("taskNum"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	task, err := services.GetTaskByNum(taskNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(task))
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新任务
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateTaskRequest true "任务信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks [post]
func (dc *DevController) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateTask(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// CreateTasks 批量创建任务
// @Summary 批量创建任务
// @Description 批量创建任务
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []models.CreateTaskRequest true "任务信息列表"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/batch [post]
func (dc *DevController) CreateTasks(c *gin.Context) {
	var reqs []models.CreateTaskRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateTasks(reqs, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskRequest true "任务信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/{taskId} [put]
func (dc *DevController) UpdateTask(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTask(taskID, &req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTaskField 更新任务字段
// @Summary 更新任务字段
// @Description 更新任务的单个字段，仅可修改：userId、taskType、startDate、endDate
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskFieldRequest true "字段信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/{taskId}/field [put]
func (dc *DevController) UpdateTaskField(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTaskField(taskID, req.Key, req.Value, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTaskNext 任务流转状态
// @Summary 任务流转状态
// @Description 更新任务状态
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskNextRequest true "任务信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks/{taskId}/next [put]
func (dc *DevController) UpdateTaskNext(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTaskNext(taskID, req.TaskStatus, req.ChangeRichText, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteTasks 删除任务
// @Summary 删除任务
// @Description 批量删除任务
// @Tags 开发管理-任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "任务ID列表"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/tasks [delete]
func (dc *DevController) DeleteTasks(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.DeleteTasks(ids, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

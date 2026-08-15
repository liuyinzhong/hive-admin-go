package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// CreateTaskExport 创建任务管理导出任务。
// @Summary 创建任务管理导出任务
// @Description 按当前筛选和排序创建异步XLSX导出任务；Worker 在计数和生成时重新解析创建用户当前数据范围
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.DevTaskExportRequest true "导出条件"
// @Success 200 {object} models.Response{data=models.DownloadTaskCreatedResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 409 {object} models.Response "活动任务数量已达上限"
// @Failure 500 {object} models.Response "创建失败"
// @Router /dev/tasks/exports [post]
func (dc *DevController) CreateTaskExport(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}
	var req models.DevTaskExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := services.NewDownloadTaskService().CreateTask(
		userID,
		models.DownloadTaskTypeDevTask,
		"任务管理导出",
		"任务管理",
		services.NewDevTaskExportPayload(req),
	)
	if err != nil {
		if errors.Is(err, services.ErrDownloadTaskLimitReached) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "创建导出任务失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 按创建人或执行人及当前角色数据范围分页获取任务
// @Tags 开发管理/任务管理
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
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.TaskResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
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

	result, err := services.GetTasks(page, pageSize, params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllTasks 获取所有任务
// @Summary 获取所有任务
// @Description 按创建人或执行人及当前角色数据范围获取所有任务（不分页）
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskNum query int false "任务编号"
// @Param taskTitle query string false "任务标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param taskStatus query int false "任务状态"
// @Param storyId query string false "需求ID"
// @Success 200 {object} models.Response{data=[]models.TaskResponse} "获取成功"
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

	tasks, err := services.GetAllTasks(params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(tasks))
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 按创建人或执行人及当前角色数据范围获取任务详情
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskNum path int true "任务编号"
// @Success 200 {object} models.Response{data=models.TaskResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks/{taskNum} [get]
func (dc *DevController) GetTask(c *gin.Context) {
	taskNum, err := strconv.Atoi(c.Param("taskNum"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	task, err := services.GetTaskByNum(taskNum, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(task))
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新任务；执行人、关联版本和关联需求必须处于当前数据范围
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateTaskRequest true "任务信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks [post]
func (dc *DevController) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateTask(&req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// CreateTasks 批量创建任务
// @Summary 批量创建任务
// @Description 批量创建任务；每条记录的执行人、关联版本和关联需求均须处于当前数据范围
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []models.CreateTaskRequest true "任务信息列表"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks/batch [post]
func (dc *DevController) CreateTasks(c *gin.Context) {
	var reqs []models.CreateTaskRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateTasks(reqs, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 按当前数据范围更新任务；执行人、关联版本和关联需求不得越界
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskRequest true "任务信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks/{taskId} [put]
func (dc *DevController) UpdateTask(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTask(taskID, &req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTaskField 更新任务字段
// @Summary 更新任务字段
// @Description 按当前数据范围更新任务的单个字段，仅可修改：userId、taskType、startDate、endDate；修改执行人时校验目标用户范围
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskFieldRequest true "字段信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks/{taskId}/field [put]
func (dc *DevController) UpdateTaskField(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTaskField(taskID, req.Key, req.Value, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateTaskNext 任务流转状态
// @Summary 任务流转状态
// @Description 按当前数据范围更新任务状态
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务ID"
// @Param request body models.UpdateTaskNextRequest true "任务信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks/{taskId}/next [put]
func (dc *DevController) UpdateTaskNext(c *gin.Context) {
	taskID := c.Param("taskId")

	var req models.UpdateTaskNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateTaskNext(taskID, req.TaskStatus, req.ChangeRichText, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteTasks 删除任务
// @Summary 删除任务
// @Description 按当前数据范围批量删除任务；任一记录不存在或越界时整批失败
// @Tags 开发管理/任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "任务ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/tasks [delete]
func (dc *DevController) DeleteTasks(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.DeleteTasks(ids, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

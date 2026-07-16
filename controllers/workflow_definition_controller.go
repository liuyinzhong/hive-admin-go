package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type WorkflowController struct{}

// GetWorkflowDefinitions 获取流程定义列表
// @Summary 获取流程定义列表
// @Description 分页获取流程定义列表
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param definitionKey query string false "流程标识"
// @Param definitionName query string false "流程名称"
// @Param category query string false "流程分类"
// @Param status query string false "流程状态，支持多选：0,1,2"
// @Param sorts query string false "排序参数，支持：definitionKey、definitionName、category、status、version、createDate、updateDate"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.WorkflowDefinitionResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions [get]
func (wc *WorkflowController) GetWorkflowDefinitions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	statuses := parseWorkflowStatusList(c.Query("status"))
	params := map[string]interface{}{
		"definitionKey":  c.Query("definitionKey"),
		"definitionName": c.Query("definitionName"),
		"category":       c.Query("category"),
		"statuses":       statuses,
		"sorts":          c.Query("sorts"),
	}

	result, err := services.GetWorkflowDefinitions(page, pageSize, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllWorkflowDefinitions 获取所有流程定义
// @Summary 获取所有流程定义
// @Description 获取所有流程定义（不分页）
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionName query string false "流程名称"
// @Param category query string false "流程分类"
// @Param status query int false "流程状态"
// @Success 200 {object} models.Response{data=[]models.WorkflowDefinitionResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/all [get]
func (wc *WorkflowController) GetAllWorkflowDefinitions(c *gin.Context) {
	status := -1
	if statusStr := c.Query("status"); statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	params := map[string]interface{}{
		"definitionName": c.Query("definitionName"),
		"category":       c.Query("category"),
		"status":         status,
	}

	definitions, err := services.GetAllWorkflowDefinitions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(definitions))
}

// GetWorkflowDefinition 获取流程定义详情
// @Summary 获取流程定义详情
// @Description 根据流程定义ID获取流程定义详情
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Success 200 {object} models.Response{data=models.WorkflowDefinitionResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/{definitionId} [get]
func (wc *WorkflowController) GetWorkflowDefinition(c *gin.Context) {
	definitionID := c.Param("definitionId")
	definition, err := services.GetWorkflowDefinition(definitionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(definition))
}

// CreateWorkflowDefinition 创建流程定义
// @Summary 创建流程定义
// @Description 创建新的流程定义
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateWorkflowDefinitionRequest true "流程定义信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions [post]
func (wc *WorkflowController) CreateWorkflowDefinition(c *gin.Context) {
	var req models.CreateWorkflowDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	if err := services.CreateWorkflowDefinition(&req, creatorID); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateWorkflowDefinition 更新流程定义
// @Summary 更新流程定义
// @Description 更新流程定义基础信息和画布数据
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Param request body models.UpdateWorkflowDefinitionRequest true "流程定义信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/{definitionId} [put]
func (wc *WorkflowController) UpdateWorkflowDefinition(c *gin.Context) {
	definitionID := c.Param("definitionId")

	var req models.UpdateWorkflowDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	if err := services.UpdateWorkflowDefinition(definitionID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateWorkflowCanvas 保存流程画布
// @Summary 保存流程画布
// @Description 保存 LogicFlow 画布JSON数据
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Param request body models.UpdateWorkflowCanvasRequest true "流程画布信息"
// @Success 200 {object} models.Response "保存成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/{definitionId}/canvas [put]
func (wc *WorkflowController) UpdateWorkflowCanvas(c *gin.Context) {
	definitionID := c.Param("definitionId")

	var req models.UpdateWorkflowCanvasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	if err := services.UpdateWorkflowCanvas(definitionID, req.FlowData); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateWorkflowFormSchema 绑定流程使用的独立表单 Schema。
// @Summary 绑定流程表单 Schema
// @Description 保存流程定义关联的表单 Schema ID，并将流程恢复为草稿状态
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Param request body models.UpdateWorkflowFormSchemaRequest true "表单 Schema ID"
// @Success 200 {object} models.Response "保存成功"
// @Failure 400 {object} models.Response "参数或表单配置错误"
// @Router /workflow/definitions/{definitionId}/form-schema [put]
func (wc *WorkflowController) UpdateWorkflowFormSchema(c *gin.Context) {
	var req models.UpdateWorkflowFormSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.UpdateWorkflowFormSchema(c.Param("definitionId"), req.FormSchemaID); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// PublishWorkflowDefinition 发布流程定义
// @Summary 发布流程定义
// @Description 发布流程定义并递增版本号
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Success 200 {object} models.Response "发布成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/{definitionId}/publish [put]
func (wc *WorkflowController) PublishWorkflowDefinition(c *gin.Context) {
	definitionID := c.Param("definitionId")
	if err := services.PublishWorkflowDefinition(definitionID); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateWorkflowDefinitionStatus 更新流程定义状态
// @Summary 更新流程定义状态
// @Description 更新流程定义状态：0草稿 1已发布 2已停用
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param definitionId path string true "流程定义ID"
// @Param request body models.UpdateWorkflowStatusRequest true "状态信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions/{definitionId}/status [put]
func (wc *WorkflowController) UpdateWorkflowDefinitionStatus(c *gin.Context) {
	definitionID := c.Param("definitionId")

	var req models.UpdateWorkflowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	status, err := strconv.Atoi(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "状态参数错误"))
		return
	}

	if err := services.UpdateWorkflowDefinitionStatus(definitionID, status); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteWorkflowDefinitions 删除流程定义
// @Summary 删除流程定义
// @Description 批量删除流程定义
// @Tags 流程管理-流程定义
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "流程定义ID数组"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/definitions [delete]
func (wc *WorkflowController) DeleteWorkflowDefinitions(c *gin.Context) {
	var definitionIDs []string
	if err := c.ShouldBindJSON(&definitionIDs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	if err := services.DeleteWorkflowDefinitions(definitionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func parseWorkflowStatusList(statusStr string) []int {
	statuses := make([]int, 0)
	if strings.TrimSpace(statusStr) == "" {
		return statuses
	}

	for _, part := range strings.Split(statusStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		status, err := strconv.Atoi(part)
		if err == nil {
			statuses = append(statuses, status)
		}
	}

	return statuses
}

package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetWorkflowAutomations 分页查询自动化动作。
// @Summary 分页查询自动化动作
// @Description 按名称、业务类型、状态分页查询自动化动作库。数据权限:全局主数据,动作库是全系统共享的流程联动配置,不按创建人过滤,维护入口由接口权限控制。
// @Tags 工作流/自动化动作
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码,默认1"
// @Param pageSize query int false "每页数量,默认10"
// @Param automationName query string false "动作名称,模糊匹配"
// @Param businessType query string false "业务类型,字典BUSINESS_TYPE的值"
// @Param status query int false "状态:0启用 1停用"
// @Success 200 {object} models.Response{data=utils.PaginationResponse} "获取成功"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/automations [get]
func (wc *WorkflowController) GetWorkflowAutomations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 未传状态参数时保持 nil,查询不过滤状态;传了非法值视为未过滤
	var status *int
	if statusText := c.Query("status"); statusText != "" {
		if parsed, err := strconv.Atoi(statusText); err == nil {
			status = &parsed
		}
	}
	result, err := services.GetAutomations(page, pageSize, c.Query("automationName"), c.Query("businessType"), status)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWorkflowAutomationOptions 按业务类型返回启用的自动化动作选项,供设计器节点挂载选择器。
// @Summary 自动化动作选项
// @Description 返回指定业务类型下启用的自动化动作列表,供流程设计器节点挂载动作时选择。数据权限:全局主数据,选项为动作库元数据,不涉及业务记录。
// @Tags 工作流/自动化动作
// @Produce json
// @Security ApiKeyAuth
// @Param businessType query string false "业务类型,字典BUSINESS_TYPE的值;为空时返回全部业务类型"
// @Success 200 {object} models.Response{data=[]models.AutomationResponse} "获取成功"
// @Failure 401 {object} models.Response "未登录"
// @Router /workflow/automations/options [get]
func (wc *WorkflowController) GetWorkflowAutomationOptions(c *gin.Context) {
	result, err := services.GetAutomationOptions(c.Query("businessType"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWorkflowAutomationFields 返回业务类型的可写字段元数据,供动作库表单目标字段下拉。
// @Summary 自动化动作可写字段
// @Description 返回指定业务类型允许修改的字段清单(版本1仅状态字段)及字段值字典类型,供动作库新建/编辑表单联动。数据权限:全局主数据,字段元数据来源于后端业务类型注册表。
// @Tags 工作流/自动化动作
// @Produce json
// @Security ApiKeyAuth
// @Param businessType query string true "业务类型,字典BUSINESS_TYPE的值"
// @Success 200 {object} models.Response{data=[]models.AutomationFieldMeta} "获取成功"
// @Failure 400 {object} models.Response "业务类型未注册"
// @Failure 401 {object} models.Response "未登录"
// @Router /workflow/automations/fields [get]
func (wc *WorkflowController) GetWorkflowAutomationFields(c *gin.Context) {
	result, err := services.GetBusinessTypeFieldMetas(c.Query("businessType"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateWorkflowAutomation 创建自动化动作。
// @Summary 创建自动化动作
// @Description 创建可复用的自动化动作。校验业务类型已注册、动作类型受支持、目标字段在业务类型可写字段白名单内、目标值为对应字典合法值。数据权限:全局主数据,新增配置不区分归属人,维护入口由接口权限控制。
// @Tags 工作流/自动化动作
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateAutomationRequest true "自动化动作参数"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "参数或校验错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/automations [post]
func (wc *WorkflowController) CreateWorkflowAutomation(c *gin.Context) {
	var req models.CreateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.CreateAutomation(&req, c.GetString("userId")); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateWorkflowAutomation 更新自动化动作。
// @Summary 更新自动化动作
// @Description 更新自动化动作配置。已挂载到流程画布的快照不受影响,重新挂载或重新发布对应流程才使用新配置。数据权限:全局主数据,更新配置不区分归属人,维护入口由接口权限控制。
// @Tags 工作流/自动化动作
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param automationId path string true "自动化动作ID"
// @Param request body models.UpdateAutomationRequest true "自动化动作参数"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "参数或校验错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/automations/{automationId} [put]
func (wc *WorkflowController) UpdateWorkflowAutomation(c *gin.Context) {
	var req models.UpdateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.UpdateAutomation(c.Param("automationId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteWorkflowAutomations 批量删除自动化动作。
// @Summary 批量删除自动化动作
// @Description 软删除自动化动作。已挂载到流程画布的快照不受影响,继续按快照执行。数据权限:全局主数据,删除配置不区分归属人,维护入口由接口权限控制。
// @Tags 工作流/自动化动作
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param automationIds body []string true "自动化动作ID数组"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /workflow/automations [delete]
func (wc *WorkflowController) DeleteWorkflowAutomations(c *gin.Context) {
	var automationIDs []string
	if err := c.ShouldBindJSON(&automationIDs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.DeleteAutomations(automationIDs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

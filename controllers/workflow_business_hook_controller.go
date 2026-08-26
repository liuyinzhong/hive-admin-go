package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// ListBusinessHooks 查询业务状态钩子注册表。
// @Summary 查询业务状态钩子注册表
// @Description 返回所有已注册业务状态钩子的元数据,供流程设计器加载业务类型和节点业务键下拉选项。数据权限:公开接口(登录后可访问),不使用记录级数据权限,返回的是钩子元数据不涉及业务记录,不按创建人过滤。
// @Tags 流程管理/业务绑定
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.BusinessHookRegistryResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /workflow/business-hooks [get]
func (wc *WorkflowController) ListBusinessHooks(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewSuccessResponse(services.ListBusinessHookRegistry()))
}

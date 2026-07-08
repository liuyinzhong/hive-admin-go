package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetNodes 获取节点列表
// @Summary 获取节点列表
// @Description 获取所有节点（不分页），可根据业务ID筛选
// @Tags 开发管理-节点管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param businessId query string false "业务ID"
// @Success 200 {object} models.Response{data=[]models.NodeResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/nodes [get]
func (dc *DevController) GetNodes(c *gin.Context) {
	businessID := c.Query("businessId")

	nodes, err := services.GetAllNodes(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nodes))
}

// CreateNode 创建节点
// @Summary 创建节点
// @Description 创建新节点
// @Tags 开发管理-节点管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateNodeRequest true "节点信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/nodes [post]
func (dc *DevController) CreateNode(c *gin.Context) {
	var req models.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.CreateNode(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteNodes 删除节点
// @Summary 删除节点
// @Description 批量删除节点
// @Tags 开发管理-节点管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "节点ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/nodes [delete]
func (dc *DevController) DeleteNodes(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.DeleteNodes(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// ApproveNode 节点审批
// @Summary 节点审批
// @Description 对审批类型的节点进行审批
// @Tags 开发管理-节点管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param nodeId path string true "节点ID"
// @Param request body models.NodeApproveRequest true "审批信息"
// @Success 200 {object} models.Response "审批成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/nodes/{nodeId}/approve [put]
func (dc *DevController) ApproveNode(c *gin.Context) {
	nodeID := c.Param("nodeId")

	var req models.NodeApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	err := services.ApproveNode(nodeID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// NextNode 节点流转
// @Summary 节点流转
// @Description 将当前节点流转到下一个节点
// @Tags 开发管理-节点管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param nodeId path string true "节点ID"
// @Success 200 {object} models.Response "流转成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/nodes/{nodeId}/next [put]
func (dc *DevController) NextNode(c *gin.Context) {
	nodeID := c.Param("nodeId")

	err := services.NextNode(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}
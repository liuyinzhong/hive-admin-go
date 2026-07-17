package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetChangeHistory 获取变更记录
// @Summary 获取变更记录
// @Description 根据业务ID获取变更记录
// @Tags 开发管理-变更记录
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param businessId query string true "业务ID"
// @Success 200 {object} models.Response{data=[]models.ChangeHistoryResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/changeHistory [get]
func (dc *DevController) GetChangeHistory(c *gin.Context) {
	businessID := c.Query("businessId")
	if businessID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	histories, err := services.GetChangeHistory(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(histories))
}

// CreateChangeHistory 创建变更记录（评论）
// @Summary 创建变更记录
// @Description 创建新的变更记录或评论
// @Tags 开发管理-变更记录
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateChangeHistoryRequest true "变更记录信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/changeHistory [post]
func (dc *DevController) CreateChangeHistory(c *gin.Context) {
	var req models.CreateChangeHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateChangeHistory(&req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

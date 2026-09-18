package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetChangeHistory 获取变更记录
// @Summary 获取变更记录
// @Description 根据业务ID获取变更记录；修改、流转、确认与自动化动作产生的记录含 changeItems 字段级变更明细（字段名+旧值→新值），存量记录无明细时为空数组；访问范围继承对应需求、任务、缺陷或版本
// @Tags 开发管理/变更记录
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

	histories, err := services.GetChangeHistory(businessID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(histories))
}

// CreateChangeHistory 创建变更记录
// @Summary 创建变更记录
// @Description 创建新的变更记录或评论；写入前校验对应需求、任务、缺陷或版本的当前访问范围。数据权限：来源对象继承，按所属需求、任务、缺陷或版本父对象校验
// @Tags 开发管理/变更记录
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
	err := services.CreateChangeHistory(&req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateChangeHistory 编辑本人评论
// @Summary 编辑评论
// @Description 编辑已有评论（changeBehavior=30），仅创建人本人可编辑，只更新正文为最新内容，不追加变更记录、不保留编辑历史。数据权限：来源对象继承，按评论所属需求、任务、缺陷或版本父对象校验，并要求当前用户为评论创建人本人（当前用户归属）
// @Tags 开发管理/变更记录
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeId path string true "变更记录ID"
// @Param request body models.UpdateChangeHistoryRequest true "评论最新内容"
// @Success 200 {object} models.Response "编辑成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/changeHistory/{changeId} [put]
func (dc *DevController) UpdateChangeHistory(c *gin.Context) {
	changeID := c.Param("changeId")
	var req models.UpdateChangeHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	userID := c.GetString("userId")
	if err := services.UpdateChangeHistory(changeID, &req, userID, currentDataPermission(c)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

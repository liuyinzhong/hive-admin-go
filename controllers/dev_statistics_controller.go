package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetTaskFindDay 统计任务趋势
// @Summary 统计任务趋势
// @Description 按小时统计两个日期的任务创建数量，用于任务趋势对比折线图
// @Tags 统计/开发管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param date1 query string true "日期1（昨天），格式 YYYY/MM/DD 或 YYYY-MM-DD" default(2026/02/25)
// @Param date2 query string true "日期2（今天），格式 YYYY/MM/DD 或 YYYY-MM-DD" default(2026/02/26)
// @Success 200 {object} models.Response{data=models.TaskFindDayResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /statistics/dev/getTaskFindDay [get]
func (dc *DevController) GetTaskFindDay(c *gin.Context) {
	date1 := c.Query("date1")
	date2 := c.Query("date2")

	result, err := services.GetTaskFindDay(date1, date2)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetTaskFindYear 统计任务年度工时
// @Summary 统计任务年度工时
// @Description 按月统计指定年份的任务实际工时合计，用于工时总量柱状图
// @Tags 统计/开发管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param year query int true "年份" default(2026)
// @Success 200 {object} models.Response{data=models.TaskFindYearResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /statistics/dev/getTaskFindYear [get]
func (dc *DevController) GetTaskFindYear(c *gin.Context) {
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year <= 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "年份参数错误"))
		return
	}

	result, err := services.GetTaskFindYear(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWorkspaceEnum 获取工作空间概览统计
// @Summary 获取工作空间概览统计
// @Description 获取需求、任务、缺陷的总数与待处理数量，用于工作台概览
// @Tags 统计/开发管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.WorkspaceEnumResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} models.Response "内部错误"
// @Router /statistics/dev/getWorkspaceEnum [get]
func (dc *DevController) GetWorkspaceEnum(c *gin.Context) {
	result, err := services.GetWorkspaceEnum()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

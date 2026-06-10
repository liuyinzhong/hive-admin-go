package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetDeptTree 获取部门树
// @Summary 获取部门树
// @Description 获取部门树结构
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query int false "状态"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts [get]
func (ctrl *SystemController) GetDeptTree(c *gin.Context) {
	var req models.DeptListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.deptService.GetDeptTree(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllDepts 获取所有部门
// @Summary 获取所有部门
// @Description 获取所有部门（不分页）
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts/all [get]
func (ctrl *SystemController) GetAllDepts(c *gin.Context) {
	result, err := ctrl.deptService.GetAllDepts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateDept 创建部门
// @Summary 创建部门
// @Description 创建新部门
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateDeptRequest true "部门信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts [post]
func (ctrl *SystemController) CreateDept(c *gin.Context) {
	var req models.CreateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.deptService.CreateDept(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetDeptDetail 获取部门详情
// @Summary 获取部门详情
// @Description 根据部门ID获取部门详情
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param deptId path string true "部门ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts/{deptId} [get]
func (ctrl *SystemController) GetDeptDetail(c *gin.Context) {
	deptId := c.Param("deptId")
	if deptId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "部门ID不能为空"))
		return
	}

	result, err := ctrl.deptService.GetDeptDetail(deptId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateDept 更新部门
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param deptId path string true "部门ID"
// @Param request body models.UpdateDeptRequest true "部门信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts/{deptId} [put]
func (ctrl *SystemController) UpdateDept(c *gin.Context) {
	deptId := c.Param("deptId")
	if deptId == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "部门ID不能为空"))
		return
	}

	var req models.UpdateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.deptService.UpdateDept(deptId, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteDepts 删除部门
// @Summary 删除部门
// @Description 批量删除部门
// @Tags 系统管理-部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "部门ID列表"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/depts [delete]
func (ctrl *SystemController) DeleteDepts(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.deptService.DeleteDepts(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

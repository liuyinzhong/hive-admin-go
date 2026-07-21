package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetMenuTree 获取菜单树
// @Summary 获取菜单树
// @Description 获取菜单树结构
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query int false "状态"
// @Success 200 {object} models.Response{data=[]models.MenuTreeResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/menus [get]
func (ctrl *SystemController) GetMenuTree(c *gin.Context) {
	var req models.MenuListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.menuService.GetMenuTree(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CheckMenuNameExists 检查菜单名称是否存在
// @Summary 检查菜单名称是否存在
// @Description 检查菜单名称是否已存在
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param name query string true "菜单名称"
// @Param id query string false "排除的菜单ID"
// @Success 200 {object} models.Response{data=bool} "检查结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/menus/nameExists [get]
func (ctrl *SystemController) CheckMenuNameExists(c *gin.Context) {
	name := c.Query("name")
	id := c.Query("id")

	var excludeId *string
	if id != "" {
		excludeId = &id
	}

	result, err := ctrl.menuService.CheckNameExists(name, excludeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CheckMenuPathExists 检查菜单路径是否存在
// @Summary 检查菜单路径是否存在
// @Description 检查菜单路径是否已存在
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param path query string true "菜单路径"
// @Param id query string false "排除的菜单ID"
// @Success 200 {object} models.Response{data=bool} "检查结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/menus/pathExists [get]
func (ctrl *SystemController) CheckMenuPathExists(c *gin.Context) {
	path := c.Query("path")
	id := c.Query("id")

	var excludeId *string
	if id != "" {
		excludeId = &id
	}

	result, err := ctrl.menuService.CheckPathExists(path, excludeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Description 创建新菜单
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateMenuRequest true "菜单信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/menus [post]
func (ctrl *SystemController) CreateMenu(c *gin.Context) {
	var req models.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.menuService.CreateMenu(req); err != nil {
		writeMenuMutationError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetMenuDetail 获取菜单详情
// @Summary 获取菜单详情
// @Description 根据菜单ID获取菜单详情
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "菜单ID"
// @Success 200 {object} models.Response{data=models.MenuTreeResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/menus/{id} [get]
func (ctrl *SystemController) GetMenuDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "菜单ID不能为空"))
		return
	}

	result, err := ctrl.menuService.GetMenuDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "菜单ID"
// @Param request body models.UpdateMenuRequest true "菜单信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/menus/{id} [put]
func (ctrl *SystemController) UpdateMenu(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "菜单ID不能为空"))
		return
	}

	var req models.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.menuService.UpdateMenu(id, req); err != nil {
		writeMenuMutationError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func writeMenuMutationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidPermissionCode),
		errors.Is(err, services.ErrPermissionCodeConflict),
		errors.Is(err, services.ErrMenuNameRequired),
		errors.Is(err, services.ErrUnsupportedMenuType),
		errors.Is(err, services.ErrRouteNameConflict),
		errors.Is(err, services.ErrRoutePathConflict):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "菜单保存失败"))
	}
}

// DeleteMenus 删除菜单
// @Summary 删除菜单
// @Description 批量删除菜单
// @Tags 系统管理/菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "菜单ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/menus [delete]
func (ctrl *SystemController) DeleteMenus(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.menuService.DeleteMenus(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

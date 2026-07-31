package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetParamList 分页查询参数列表
// @Summary 分页查询参数列表
// @Description 分页查询系统参数列表,支持按参数键、类型、是否公开筛选与排序
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码,默认1" default(1)
// @Param pageSize query int false "每页大小,默认20" default(20)
// @Param paramKey query string false "参数键,模糊搜索" example(sys.session)
// @Param paramType query string false "参数类型,精确匹配 string/number/boolean/json" example(number)
// @Param isPublic query int false "是否公开 0=否 1=是" example(1)
// @Param sorts query string false "排序参数,支持 paramKey/paramType/isPublic/updateDate/createDate" example(updateDate,desc)
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.ParamResponse}} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/params [get]
func (ctrl *SystemController) GetParamList(c *gin.Context) {
	var req models.ParamListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.paramService.GetParamList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateParam 创建参数
// @Summary 创建参数
// @Description 创建新的系统参数,校验 paramKey 格式与唯一性、paramType 与 paramValue 一致性
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateParamRequest true "参数信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "参数键已存在或创建失败"
// @Router /system/params [post]
func (ctrl *SystemController) CreateParam(c *gin.Context) {
	var req models.CreateParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.paramService.CreateParam(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetParamDetail 查询参数详情
// @Summary 查询参数详情
// @Description 根据参数ID查询参数详情
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "参数ID"
// @Success 200 {object} models.Response{data=models.ParamResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "参数不存在"
// @Router /system/params/{id} [get]
func (ctrl *SystemController) GetParamDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数ID不能为空"))
		return
	}

	result, err := ctrl.paramService.GetParamDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateParam 更新参数
// @Summary 更新参数
// @Description 更新参数信息,允许修改 paramKey(校验唯一性排除自身)
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "参数ID"
// @Param request body models.UpdateParamRequest true "参数信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Failure 500 {object} models.Response "参数不存在、参数键已存在或校验失败"
// @Router /system/params/{id} [put]
func (ctrl *SystemController) UpdateParam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数ID不能为空"))
		return
	}

	var req models.UpdateParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.paramService.UpdateParam(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteParams 批量删除参数
// @Summary 批量删除参数
// @Description 批量逻辑删除参数
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "参数ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/params [delete]
func (ctrl *SystemController) DeleteParams(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	if err := ctrl.paramService.DeleteParams(ids); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetParamValues 公共参数批量查询
// @Summary 公共参数批量查询
// @Description 按参数键批量查询公开参数(isPublic=1),返回 key->value 映射。需登录但无需接口权限
// @Tags 系统管理/参数管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.ParamValuesRequest false "参数键数组,为空时返回全部公开参数"
// @Success 200 {object} models.Response{data=object} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/params/values [post]
func (ctrl *SystemController) GetParamValues(c *gin.Context) {
	var req models.ParamValuesRequest
	// 允许空 body
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.paramService.GetParamValues(req.Keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

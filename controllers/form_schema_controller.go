package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type FormSchemaController struct{}

// GetFormSchemas 获取表单 Schema 分页列表。
// @Summary 获取表单 Schema 列表
// @Tags 表单管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param schemaKey query string false "Schema 标识"
// @Param schemaName query string false "Schema 名称"
// @Param category query string false "分类"
// @Param status query int false "状态"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /form/schemas [get]
func (FormSchemaController) GetFormSchemas(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := -1
	if value := c.Query("status"); value != "" {
		status, _ = strconv.Atoi(value)
	}
	result, err := services.GetFormSchemas(page, pageSize, map[string]interface{}{
		"schemaKey": c.Query("schemaKey"), "schemaName": c.Query("schemaName"),
		"category": c.Query("category"), "status": status, "sorts": c.Query("sorts"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllFormSchemas 获取全部可选表单 Schema。
// @Summary 获取全部表单 Schema
// @Tags 表单管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.FormSchemaResponse}
// @Router /form/schemas/all [get]
func (FormSchemaController) GetAllFormSchemas(c *gin.Context) {
	status := -1
	if value := c.Query("status"); value != "" {
		status, _ = strconv.Atoi(value)
	}
	result, err := services.GetAllFormSchemas(map[string]interface{}{
		"schemaName": c.Query("schemaName"), "status": status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetFormSchema 获取表单 Schema 详情。
// @Summary 获取表单 Schema 详情
// @Tags 表单管理
// @Produce json
// @Security ApiKeyAuth
// @Param formSchemaId path string true "表单 Schema ID"
// @Success 200 {object} models.Response{data=models.FormSchemaResponse}
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /form/schemas/{formSchemaId} [get]
func (FormSchemaController) GetFormSchema(c *gin.Context) {
	result, err := services.GetFormSchema(c.Param("formSchemaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateFormSchema 创建表单 Schema。
// @Summary 创建表单 Schema
// @Tags 表单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.UpsertFormSchemaRequest true "表单 Schema"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /form/schemas [post]
func (FormSchemaController) CreateFormSchema(c *gin.Context) {
	var req models.UpsertFormSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.CreateFormSchema(&req, c.GetString("userId")); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateFormSchema 更新表单 Schema。
// @Summary 更新表单 Schema
// @Tags 表单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param formSchemaId path string true "表单 Schema ID"
// @Param request body models.UpsertFormSchemaRequest true "表单 Schema"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /form/schemas/{formSchemaId} [put]
func (FormSchemaController) UpdateFormSchema(c *gin.Context) {
	var req models.UpsertFormSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := services.UpdateFormSchema(c.Param("formSchemaId"), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteFormSchemas 删除表单 Schema。
// @Summary 删除表单 Schema
// @Tags 表单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "表单 Schema ID 数组"
// @Success 200 {object} models.Response
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /form/schemas [delete]
func (FormSchemaController) DeleteFormSchemas(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	for index := range ids {
		ids[index] = strings.TrimSpace(ids[index])
	}
	if err := services.DeleteFormSchemas(ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

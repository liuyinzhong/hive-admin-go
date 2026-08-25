package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件并以当前用户作为元数据创建人；现有 /uploads/** 静态 URL 为公开访问，不适合敏感文件
// @Tags 系统管理/文件管理
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "文件"
// @Success 200 {object} models.Response{data=models.FileResponse} "上传成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /system/upload [post]
func (ctrl *SystemController) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "请选择上传文件"))
		return
	}

	result, err := ctrl.fileService.UploadFile(fileHeader, c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetFileList 获取文件列表
// @Summary 获取文件列表
// @Description 按文件创建人和当前角色数据范围分页获取文件元数据；不改变 /uploads/** 的公开静态访问边界。支持按文件状态 status 精确过滤（0=正式，1=临时未绑定），响应包含 status 字段
// @Tags 系统管理/文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param originalName query string false "原始文件名(模糊搜索)"
// @Param type query string false "MIME类型(精确匹配)"
// @Param fileExt query string false "文件扩展名(精确匹配，如 .jpg)"
// @Param status query int false "文件状态(精确匹配，0=正式，1=临时未绑定)"
// @Param sorts query string false "排序参数(如 createDate,desc;size,asc)"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.FileResponse}} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /system/files [get]
func (ctrl *SystemController) GetFileList(c *gin.Context) {
	var req models.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}

	result, err := ctrl.fileService.GetFileList(req, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

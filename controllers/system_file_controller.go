package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到系统文件表
// @Tags 系统管理-文件管理
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
// @Description 分页获取文件列表
// @Tags 系统管理-文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param originalName query string false "原始文件名(模糊搜索)"
// @Param type query string false "MIME类型(精确匹配)"
// @Param fileExt query string false "文件扩展名(精确匹配，如 .jpg)"
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

	result, err := ctrl.fileService.GetFileList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

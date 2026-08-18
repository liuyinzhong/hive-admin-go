package controllers

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"hive-admin-go/models"
	"hive-admin-go/services"

	"github.com/gin-gonic/gin"
)

type DownloadTaskController struct {
	service *services.DownloadTaskService
}

func NewDownloadTaskController() *DownloadTaskController {
	return &DownloadTaskController{service: services.NewDownloadTaskService()}
}

// GetList 获取下载中心列表。
// @Summary 获取下载中心列表
// @Description 仅返回当前登录用户创建的下载任务，文件保留7天，任务记录保留30天
// @Tags 系统管理/下载中心
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小，最大100"
// @Param taskName query string false "任务名称，模糊匹配"
// @Param status query string false "状态" Enums(pending,running,succeeded,failed)
// @Param createDateStart query string false "创建时间起，格式 YYYY-MM-DD HH:mm:ss"
// @Param createDateEnd query string false "创建时间止，格式 YYYY-MM-DD HH:mm:ss"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.DownloadTaskResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录"
// @Failure 500 {object} models.Response "获取失败"
// @Router /system/downloads [get]
func (ctrl *DownloadTaskController) GetList(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}
	var req models.DownloadTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err, "参数错误"))
		return
	}
	result, err := ctrl.service.GetList(userID, req)
	if err != nil {
		if errors.Is(err, services.ErrDownloadInvalidDate) || errors.Is(err, services.ErrDownloadInvalidStatus) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "获取下载任务失败"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DownloadFile 下载已生成文件。
// @Summary 下载导出文件
// @Description 下载当前用户自己的、未过期且已生成的导出文件
// @Tags 系统管理/下载中心
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security ApiKeyAuth
// @Param id path string true "下载任务ID"
// @Success 200 {file} binary "文件流"
// @Failure 401 {object} models.Response "未登录"
// @Failure 404 {object} models.Response "任务不存在"
// @Failure 409 {object} models.Response "文件不可用"
// @Failure 500 {object} models.Response "下载失败"
// @Router /system/downloads/{id}/file [get]
func (ctrl *DownloadTaskController) DownloadFile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}
	task, err := ctrl.service.GetDownloadFile(userID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDownloadTaskNotFound):
			c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, err.Error()))
		case errors.Is(err, services.ErrDownloadFileUnavailable):
			c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "下载文件失败"))
		}
		return
	}
	file, err := os.Open(*task.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "下载文件失败"))
		return
	}
	defer file.Close()
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": *task.FileName}))
	c.Header("Content-Length", strconv.FormatInt(task.FileSize, 10))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("stream download task file %s failed: %v", task.ID, err)
	}
}

// GetPreviewURL 生成下载文件的短时预览链接。
// @Summary 生成下载文件预览链接
// @Description 数据权限：当前用户归属。校验任务属于当前登录用户、状态为成功且文件未过期后，生成 5 分钟有效的预览 token，返回相对路径。前端拼接后传给 kkFileView 等外部预览服务调用。
// @Tags 系统管理/下载中心
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "下载任务ID"
// @Success 200 {object} models.Response{data=models.DownloadPreviewURLResponse} "生成成功"
// @Failure 401 {object} models.Response "未登录"
// @Failure 404 {object} models.Response "任务不存在"
// @Failure 409 {object} models.Response "文件不可用"
// @Failure 500 {object} models.Response "生成失败"
// @Router /system/downloads/{id}/preview-url [get]
func (ctrl *DownloadTaskController) GetPreviewURL(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户未登录"))
		return
	}
	previewURL, err := ctrl.service.GeneratePreviewURL(userID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDownloadTaskNotFound):
			c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, err.Error()))
		case errors.Is(err, services.ErrDownloadFileUnavailable):
			c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "生成预览链接失败"))
		}
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(models.DownloadPreviewURLResponse{PreviewURL: previewURL}))
}

// PreviewFile 通过预览 token 获取文件流，供 kkFileView 等外部预览服务调用。
// @Summary 通过预览 token 获取文件
// @Description 数据权限：公开接口。通过 JWT 签名校验确保 token 由本服务签发且未过期，再复用下载校验确保任务仍属于签发用户且文件可用。kkFileView 等外部预览服务无登录态时通过此接口取文件。
// @Tags 系统管理/下载中心
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param token path string true "预览 token"
// @Success 200 {file} binary "文件流"
// @Failure 401 {object} models.Response "预览链接无效或已过期"
// @Failure 404 {object} models.Response "任务不存在"
// @Failure 409 {object} models.Response "文件不可用"
// @Failure 500 {object} models.Response "预览失败"
// @Router /public/downloads/preview/{token} [get]
func (ctrl *DownloadTaskController) PreviewFile(c *gin.Context) {
	task, err := ctrl.service.GetPreviewFile(c.Param("token"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDownloadPreviewInvalid):
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, err.Error()))
		case errors.Is(err, services.ErrDownloadTaskNotFound):
			c.JSON(http.StatusNotFound, models.NewErrorResponse(nil, err.Error()))
		case errors.Is(err, services.ErrDownloadFileUnavailable):
			c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "预览文件失败"))
		}
		return
	}
	file, err := os.Open(*task.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "预览文件失败"))
		return
	}
	defer file.Close()
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": *task.FileName}))
	c.Header("Content-Length", strconv.FormatInt(task.FileSize, 10))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("stream preview task file %s failed: %v", task.ID, err)
	}
}

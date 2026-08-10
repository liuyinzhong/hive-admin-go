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

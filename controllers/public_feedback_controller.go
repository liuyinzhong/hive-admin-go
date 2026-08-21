package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// PublicFeedbackController 外部反馈工单公开接口控制器。
// 不依赖登录态和数据权限，仅暴露工单提交和附件上传两个匿名接口。
type PublicFeedbackController struct {
	fileService *services.FileService
}

// NewPublicFeedbackController 构造公开反馈控制器，复用 FileService 完成匿名上传。
func NewPublicFeedbackController() *PublicFeedbackController {
	return &PublicFeedbackController{
		fileService: services.NewFileService(),
	}
}

// CreateFeedback 提交外部反馈工单。
// @Summary 提交外部反馈工单
// @Description 数据权限：公开接口，不按创建人过滤；type=story 写入 dev_story，type=bug 写入 dev_bug，source 固定为 10（外部反馈）。creator_id/version_id/project_id/module_id 全部留空，fileIds 必须由 /api/public/upload 公开上传产生。
// @Tags 公共接口/外部反馈
// @Accept json
// @Produce json
// @Param request body models.CreateStoryFeedbackRequest true "工单提交请求"
// @Success 200 {object} models.Response{data=models.CreateStoryFeedbackResponse}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /public/feedback [post]
func (ctrl *PublicFeedbackController) CreateFeedback(c *gin.Context) {
	var req models.CreateStoryFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	// 记录提交方浏览器 UA，仅 type=bug 时写入 bug_ua，便于复现缺陷环境
	ua := c.GetHeader("User-Agent")
	if ua != "" {
		req.UserAgent = &ua
	}
	result, err := services.CreateStoryFeedback(&req)
	if err != nil {
		writePublicFeedbackError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UploadFile 公开上传外部反馈附件。
// @Summary 公开上传外部反馈附件
// @Description 数据权限：公开接口，不按创建人过滤；文件元数据 creator_id 写入固定占位标记 external-feedback，工单提交时校验附件必须由此接口产生，避免伪造内部登录用户上传的文件 ID。
// @Tags 公共接口/外部反馈
// @Accept mpfd
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {object} models.Response{data=models.FileResponse}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /public/upload [post]
func (ctrl *PublicFeedbackController) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "请选择上传文件"))
		return
	}
	// 公开上传固定使用占位 creator_id，工单提交时据此校验附件来源
	result, err := ctrl.fileService.UploadFile(fileHeader, models.ExternalFeedbackFileCreatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// writePublicFeedbackError 将外部反馈服务错误映射为合适的 HTTP 状态码。
// 输入校验类错误返回 400，其余返回 500，遵循项目"业务错误 HTTP 200 code=-1"以外的参数错误惯例。
func writePublicFeedbackError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrExternalFeedbackInvalidType),
		errors.Is(err, services.ErrExternalFeedbackTitleRequired),
		errors.Is(err, services.ErrExternalFeedbackFileIDsInvalid):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "外部反馈提交失败"))
	}
}

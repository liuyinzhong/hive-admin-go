package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

type BaseInstitutionController struct {
	institutionService *services.BaseInstitutionService
}

func NewBaseInstitutionController() *BaseInstitutionController {
	return &BaseInstitutionController{
		institutionService: services.NewBaseInstitutionService(),
	}
}

// GetInstitution 获取当前机构资料。
// @Summary 获取机构资料
// @Description 获取当前系统唯一的机构资料聚合，未初始化时返回空数据。
// @Tags 基础资料/机构资料
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=models.InstitutionResponse} "获取成功"
// @Failure 403 {object} models.Response "无权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/institution [get]
func (ctrl *BaseInstitutionController) GetInstitution(c *gin.Context) {
	result, err := ctrl.institutionService.GetInstitution()
	if err != nil {
		writeBaseInstitutionError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// SaveInstitution 保存机构资料。
// @Summary 保存机构资料
// @Description 使用一个平铺聚合请求保存机构主档及全部当前子资料。集合型子资料会按请求全量替换，空数组表示清空，资质附件直接提交上传后返回的 URL 字符串。
// @Tags 基础资料/机构资料
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveInstitutionRequest true "机构资料"
// @Success 200 {object} models.Response{data=models.InstitutionResponse} "保存成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "无权限"
// @Failure 409 {object} models.Response "数据冲突"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/institution [put]
func (ctrl *BaseInstitutionController) SaveInstitution(c *gin.Context) {
	var req models.SaveInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.institutionService.SaveInstitution(req, baseOperatorID(c))
	if err != nil {
		writeBaseInstitutionError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

func writeBaseInstitutionError(c *gin.Context, err error) {
	message := "机构资料操作失败"
	switch {
	case errors.Is(err, services.ErrBaseInstitutionInvalidInput), errors.Is(err, services.ErrBaseInstitutionNotFound):
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, err.Error()))
	case errors.Is(err, services.ErrBaseInstitutionConflict):
		c.JSON(http.StatusConflict, models.NewErrorResponse(nil, err.Error()))
	default:
		log.Printf("机构资料操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
	}
}

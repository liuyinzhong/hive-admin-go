package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// ClassificationController 分类体系与节点控制器
type ClassificationController struct {
	systemService *services.ClassificationSystemService
	nodeService   *services.ClassificationNodeService
}

// NewClassificationController 创建分类控制器实例
func NewClassificationController() *ClassificationController {
	return &ClassificationController{
		systemService: services.NewClassificationSystemService(),
		nodeService:   services.NewClassificationNodeService(),
	}
}

// GetClassificationSystemList 获取全部分类体系
// @Summary 获取全部分类体系
// @Description 全量查询分类体系列表，按排序号和创建时间升序排列
// @Tags 基础资料/分类体系
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response{data=[]models.ClassificationSystemResponse} "获取成功"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems [get]
func (ctrl *ClassificationController) GetClassificationSystemList(c *gin.Context) {
	result, err := ctrl.systemService.GetAllSystems()
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetClassificationSystem 获取分类体系详情
// @Summary 获取分类体系详情
// @Description 根据分类体系ID获取详细信息
// @Tags 基础资料/分类体系
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类体系ID"
// @Success 200 {object} models.Response{data=models.ClassificationSystemResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/{id} [get]
func (ctrl *ClassificationController) GetClassificationSystem(c *gin.Context) {
	result, err := ctrl.systemService.GetSystemDetail(c.Param("id"))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateClassificationSystem 新增分类体系
// @Summary 新增分类体系
// @Description 创建新的分类体系记录
// @Tags 基础资料/分类体系
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveClassificationSystemRequest true "分类体系"
// @Success 200 {object} models.Response{data=models.ClassificationSystemResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems [post]
func (ctrl *ClassificationController) CreateClassificationSystem(c *gin.Context) {
	var req models.SaveClassificationSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.systemService.CreateSystem(req, baseOperatorID(c))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateClassificationSystem 更新分类体系
// @Summary 更新分类体系
// @Description 根据分类体系ID更新分类体系信息
// @Tags 基础资料/分类体系
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类体系ID"
// @Param request body models.SaveClassificationSystemRequest true "分类体系"
// @Success 200 {object} models.Response{data=models.ClassificationSystemResponse} "更新成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/{id} [put]
func (ctrl *ClassificationController) UpdateClassificationSystem(c *gin.Context) {
	var req models.SaveClassificationSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.systemService.UpdateSystem(c.Param("id"), req, baseOperatorID(c))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteClassificationSystem 删除分类体系
// @Summary 删除分类体系
// @Description 单条删除分类体系，存在节点时禁止删除
// @Tags 基础资料/分类体系
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类体系ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/{id} [delete]
func (ctrl *ClassificationController) DeleteClassificationSystem(c *gin.Context) {
	if err := ctrl.systemService.DeleteSystem(c.Param("id"), baseOperatorID(c)); err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetClassificationNodeTree 获取分类节点树
// @Summary 获取分类节点树
// @Description 按体系编码查询节点树，支持关键字过滤并保留祖先链
// @Tags 基础资料/分类节点
// @Produce json
// @Security ApiKeyAuth
// @Param systemCode query string true "体系编码"
// @Param keyword query string false "节点编码或名称关键字"
// @Param status query int false "状态：0停用 1启用"
// @Success 200 {object} models.Response{data=[]models.ClassificationNodeTreeResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes [get]
func (ctrl *ClassificationController) GetClassificationNodeTree(c *gin.Context) {
	var req models.ClassificationNodeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.nodeService.GetNodeTree(req)
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetClassificationNodeOptions 获取分类节点公共选项
// @Summary 获取分类节点公共选项
// @Description 按体系编码返回启用且未删除的节点树，需登录但无按钮权限
// @Tags 基础资料/分类节点
// @Produce json
// @Security ApiKeyAuth
// @Param systemCode query string true "体系编码"
// @Success 200 {object} models.Response{data=[]models.ClassificationNodeTreeResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/options [get]
func (ctrl *ClassificationController) GetClassificationNodeOptions(c *gin.Context) {
	var req models.ClassificationNodeOptionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.nodeService.GetNodeOptions(req)
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetClassificationNode 获取分类节点详情
// @Summary 获取分类节点详情
// @Description 根据分类节点ID获取详细信息
// @Tags 基础资料/分类节点
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类节点ID"
// @Success 200 {object} models.Response{data=models.ClassificationNodeTreeResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes/{id} [get]
func (ctrl *ClassificationController) GetClassificationNode(c *gin.Context) {
	result, err := ctrl.nodeService.GetNodeDetail(c.Param("id"))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateClassificationNode 新增分类节点
// @Summary 新增分类节点
// @Description 创建新的分类节点记录
// @Tags 基础资料/分类节点
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SaveClassificationNodeRequest true "分类节点"
// @Success 200 {object} models.Response{data=models.ClassificationNodeTreeResponse} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes [post]
func (ctrl *ClassificationController) CreateClassificationNode(c *gin.Context) {
	var req models.SaveClassificationNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.nodeService.CreateNode(req, baseOperatorID(c))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateClassificationNode 更新分类节点
// @Summary 更新分类节点
// @Description 根据分类节点ID更新节点信息，支持移动父级
// @Tags 基础资料/分类节点
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类节点ID"
// @Param request body models.SaveClassificationNodeRequest true "分类节点"
// @Success 200 {object} models.Response{data=models.ClassificationNodeTreeResponse} "更新成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes/{id} [put]
func (ctrl *ClassificationController) UpdateClassificationNode(c *gin.Context) {
	var req models.SaveClassificationNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.nodeService.UpdateNode(c.Param("id"), req, baseOperatorID(c))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateClassificationNodeStatus 更新分类节点启停状态
// @Summary 更新分类节点启停状态
// @Description 根据分类节点ID更新启用/停用状态
// @Tags 基础资料/分类节点
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类节点ID"
// @Param request body models.UpdateClassificationNodeStatusRequest true "状态"
// @Success 200 {object} models.Response{data=models.ClassificationNodeTreeResponse} "更新成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes/{id}/status [put]
func (ctrl *ClassificationController) UpdateClassificationNodeStatus(c *gin.Context) {
	var req models.UpdateClassificationNodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.nodeService.UpdateNodeStatus(c.Param("id"), req, baseOperatorID(c))
	if err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteClassificationNode 删除分类节点
// @Summary 删除分类节点
// @Description 单条删除分类节点，存在子节点时禁止删除
// @Tags 基础资料/分类节点
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "分类节点ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /base/classificationSystems/nodes/{id} [delete]
func (ctrl *ClassificationController) DeleteClassificationNode(c *gin.Context) {
	if err := ctrl.nodeService.DeleteNode(c.Param("id"), baseOperatorID(c)); err != nil {
		writeClassificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// writeClassificationError 统一处理分类体系错误响应
func writeClassificationError(c *gin.Context, err error) {
	message := "分类体系操作失败"
	switch {
	case errors.Is(err, services.ErrClassificationInvalidInput),
		errors.Is(err, services.ErrClassificationNotFound),
		errors.Is(err, services.ErrClassificationConflict):
		message = err.Error()
	default:
		log.Printf("分类体系操作失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, message))
		return
	}
	c.JSON(http.StatusOK, models.NewErrorResponse(nil, message))
}

package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetStorys 获取需求列表
// @Summary 获取需求列表
// @Description 按创建人或参与人及当前角色数据范围分页获取需求
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param storyNum query int false "需求编号"
// @Param storyTitle query string false "需求标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param moduleId query string false "模块ID"
// @Param storyStatus query string false "需求状态，支持多选：1,2"
// @Param sorts query string false "排序参数"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.StoryResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys [get]
func (dc *DevController) GetStorys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	storyNum := 0
	if snStr := c.Query("storyNum"); snStr != "" {
		storyNum, _ = strconv.Atoi(snStr)
	}

	storyStatuses := make([]int, 0)
	if ssStr := strings.TrimSpace(c.Query("storyStatus")); ssStr != "" {
		for _, part := range strings.Split(ssStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			status, err := strconv.Atoi(part)
			if err == nil {
				storyStatuses = append(storyStatuses, status)
			}
		}
	}

	params := map[string]interface{}{
		"storyNum":      storyNum,
		"storyTitle":    c.Query("storyTitle"),
		"projectId":     c.Query("projectId"),
		"versionId":     c.Query("versionId"),
		"moduleId":      c.Query("moduleId"),
		"storyStatuses": storyStatuses,
		"sorts":         c.Query("sorts"),
	}

	result, err := services.GetStorys(page, pageSize, params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllStorys 获取所有需求
// @Summary 获取所有需求
// @Description 按创建人或参与人及当前角色数据范围获取所有需求（不分页）
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param storyNum query int false "需求编号"
// @Param storyTitle query string false "需求标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param moduleId query string false "模块ID"
// @Param storyStatus query int false "需求状态"
// @Success 200 {object} models.Response{data=[]models.StoryResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/storys/all [get]
func (dc *DevController) GetAllStorys(c *gin.Context) {
	storyNum := 0
	if snStr := c.Query("storyNum"); snStr != "" {
		storyNum, _ = strconv.Atoi(snStr)
	}
	storyStatus := -1
	if ssStr := c.Query("storyStatus"); ssStr != "" {
		storyStatus, _ = strconv.Atoi(ssStr)
	}

	params := map[string]interface{}{
		"storyNum":    storyNum,
		"storyTitle":  c.Query("storyTitle"),
		"projectId":   c.Query("projectId"),
		"versionId":   c.Query("versionId"),
		"moduleId":    c.Query("moduleId"),
		"storyStatus": storyStatus,
	}

	storys, err := services.GetAllStorys(params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(storys))
}

// GetStory 获取需求详情
// @Summary 获取需求详情
// @Description 按创建人或参与人及当前角色数据范围获取需求详情，关联任务和缺陷继续按各自范围过滤
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param storyNum path int true "需求编号"
// @Success 200 {object} models.Response{data=models.StoryResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys/{storyNum} [get]
func (dc *DevController) GetStory(c *gin.Context) {
	storyNum, err := strconv.Atoi(c.Param("storyNum"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	story, err := services.GetStoryByNum(storyNum, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(story))
}

// CreateStory 创建需求
// @Summary 创建需求
// @Description 创建新需求；参与人、附件和关联版本必须处于当前数据范围
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateStoryRequest true "需求信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys [post]
func (dc *DevController) CreateStory(c *gin.Context) {
	var req models.CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateStory(&req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// CreateStorys 批量创建需求
// @Summary 批量创建需求
// @Description 批量创建需求；每条记录的参与人、附件和关联版本均须处于当前数据范围
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []models.CreateStoryRequest true "需求信息列表"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys/batch [post]
func (dc *DevController) CreateStorys(c *gin.Context) {
	var reqs []models.CreateStoryRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateStorys(reqs, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateStory 更新需求
// @Summary 更新需求
// @Description 按当前数据范围更新需求；参与人、附件和关联版本不得越界
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param storyId path string true "需求ID"
// @Param request body models.UpdateStoryRequest true "需求信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys/{storyId} [put]
func (dc *DevController) UpdateStory(c *gin.Context) {
	storyID := c.Param("storyId")

	var req models.UpdateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateStory(storyID, &req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateStoryField 更新需求字段
// @Summary 更新需求字段
// @Description 按当前数据范围更新需求单个字段；修改参与人时校验目标用户范围
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param storyId path string true "需求ID"
// @Param request body models.UpdateStoryFieldRequest true "字段信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys/{storyId}/field [put]
func (dc *DevController) UpdateStoryField(c *gin.Context) {
	storyID := c.Param("storyId")

	var req models.UpdateStoryFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateStoryField(storyID, req.Key, req.Value, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateStoryNext 需求流转状态
// @Summary 需求流转状态
// @Description 按当前数据范围更新需求状态
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param storyId path string true "需求ID"
// @Param request body models.UpdateStoryNextRequest true "需求信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys/{storyId}/next [put]
func (dc *DevController) UpdateStoryNext(c *gin.Context) {
	storyID := c.Param("storyId")

	var req models.UpdateStoryNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateStoryNext(storyID, req.StoryStatus, req.ChangeRichText, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteStorys 删除需求
// @Summary 删除需求
// @Description 按当前数据范围批量删除需求；任一记录不存在或越界时整批失败
// @Tags 开发管理/需求管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "需求ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/storys [delete]
func (dc *DevController) DeleteStorys(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.DeleteStorys(ids, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

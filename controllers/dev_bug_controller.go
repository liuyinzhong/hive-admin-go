package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
	"hive-admin-go/services"
)

// GetBugs 获取缺陷列表
// @Summary 获取缺陷列表
// @Description 按创建人或处理人及当前角色数据范围分页获取缺陷
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param bugNum query int false "缺陷编号"
// @Param bugTitle query string false "缺陷标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param moduleId query string false "模块ID"
// @Param bugStatus query string false "缺陷状态，支持多选：1,2"
// @Param storyId query string false "需求ID"
// @Param sorts query string false "排序参数"
// @Success 200 {object} models.Response{data=utils.PageResult{items=[]models.BugResponse}} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs [get]
func (dc *DevController) GetBugs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	bugNum := 0
	if bnStr := c.Query("bugNum"); bnStr != "" {
		bugNum, _ = strconv.Atoi(bnStr)
	}
	bugStatuses := make([]int, 0)
	if bsStr := strings.TrimSpace(c.Query("bugStatus")); bsStr != "" {
		for _, part := range strings.Split(bsStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			status, err := strconv.Atoi(part)
			if err == nil {
				bugStatuses = append(bugStatuses, status)
			}
		}
	}

	params := map[string]interface{}{
		"bugNum":      bugNum,
		"bugTitle":    c.Query("bugTitle"),
		"projectId":   c.Query("projectId"),
		"versionId":   c.Query("versionId"),
		"moduleId":    c.Query("moduleId"),
		"bugStatuses": bugStatuses,
		"storyId":     c.Query("storyId"),
		"sorts":       c.Query("sorts"),
	}

	result, err := services.GetBugs(page, pageSize, params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetAllBugs 获取所有缺陷
// @Summary 获取所有缺陷
// @Description 按创建人或处理人及当前角色数据范围获取所有缺陷（不分页）
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugNum query int false "缺陷编号"
// @Param bugTitle query string false "缺陷标题"
// @Param projectId query string false "项目ID"
// @Param versionId query string false "版本ID"
// @Param moduleId query string false "模块ID"
// @Param bugStatus query int false "缺陷状态"
// @Param storyId query string false "需求ID"
// @Success 200 {object} models.Response{data=[]models.BugResponse} "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /dev/bugs/all [get]
func (dc *DevController) GetAllBugs(c *gin.Context) {
	bugNum := 0
	if bnStr := c.Query("bugNum"); bnStr != "" {
		bugNum, _ = strconv.Atoi(bnStr)
	}
	bugStatus := -1
	if bsStr := c.Query("bugStatus"); bsStr != "" {
		bugStatus, _ = strconv.Atoi(bsStr)
	}

	params := map[string]interface{}{
		"bugNum":    bugNum,
		"bugTitle":  c.Query("bugTitle"),
		"projectId": c.Query("projectId"),
		"versionId": c.Query("versionId"),
		"moduleId":  c.Query("moduleId"),
		"bugStatus": bugStatus,
		"storyId":   c.Query("storyId"),
	}

	bugs, err := services.GetAllBugs(params, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(bugs))
}

// GetBug 获取缺陷详情
// @Summary 获取缺陷详情
// @Description 按创建人或处理人及当前角色数据范围获取缺陷详情
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugNum path int true "缺陷编号"
// @Success 200 {object} models.Response{data=models.BugResponse} "获取成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/{bugNum} [get]
func (dc *DevController) GetBug(c *gin.Context) {
	bugNum, err := strconv.Atoi(c.Param("bugNum"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	bug, err := services.GetBugByNum(bugNum, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(bug))
}

// CreateBug 创建缺陷
// @Summary 创建缺陷
// @Description 创建新缺陷；处理人、关联版本和关联需求必须处于当前数据范围
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateBugRequest true "缺陷信息"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs [post]
func (dc *DevController) CreateBug(c *gin.Context) {
	var req models.CreateBugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateBug(&req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// ConfirmBug 确认缺陷
// @Summary 确认缺陷
// @Description 按当前数据范围确认缺陷并更新缺陷状态
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugId path string true "缺陷ID"
// @Param request body models.ConfirmBugRequest true "缺陷确认信息"
// @Success 200 {object} models.Response "确认成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/{bugId}/confirm [put]
func (dc *DevController) ConfirmBug(c *gin.Context) {
	bugID := c.Param("bugId")

	var req models.ConfirmBugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.ConfirmBug(bugID, &req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// CreateBugs 批量创建缺陷
// @Summary 批量创建缺陷
// @Description 批量创建缺陷；每条记录的处理人、关联版本和关联需求均须处于当前数据范围
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []models.CreateBugRequest true "缺陷信息列表"
// @Success 200 {object} models.Response "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/batch [post]
func (dc *DevController) CreateBugs(c *gin.Context) {
	var reqs []models.CreateBugRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.CreateBugs(reqs, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateBug 更新缺陷
// @Summary 更新缺陷
// @Description 按当前数据范围更新缺陷；处理人、关联版本和关联需求不得越界
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugId path string true "缺陷ID"
// @Param request body models.UpdateBugRequest true "缺陷信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/{bugId} [put]
func (dc *DevController) UpdateBug(c *gin.Context) {
	bugID := c.Param("bugId")

	var req models.UpdateBugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateBug(bugID, &req, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateBugField 更新缺陷字段
// @Summary 更新缺陷字段
// @Description 按当前数据范围更新缺陷的单个字段；修改处理人时校验目标用户范围
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugId path string true "缺陷ID"
// @Param request body models.UpdateBugFieldRequest true "字段信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/{bugId}/field [put]
func (dc *DevController) UpdateBugField(c *gin.Context) {
	bugID := c.Param("bugId")

	var req models.UpdateBugFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateBugField(bugID, req.Key, req.Value, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// UpdateBugNext 缺陷流转状态
// @Summary 缺陷流转状态
// @Description 按当前数据范围更新缺陷状态
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bugId path string true "缺陷ID"
// @Param request body models.UpdateBugNextRequest true "缺陷信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs/{bugId}/next [put]
func (dc *DevController) UpdateBugNext(c *gin.Context) {
	bugID := c.Param("bugId")

	var req models.UpdateBugNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.UpdateBugNext(bugID, req.BugStatus, req.ChangeRichText, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// DeleteBugs 删除缺陷
// @Summary 删除缺陷
// @Description 按当前数据范围批量删除缺陷；任一记录不存在或越界时整批失败
// @Tags 开发管理/缺陷管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []string true "缺陷ID列表"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} models.Response "无接口访问权限"
// @Router /dev/bugs [delete]
func (dc *DevController) DeleteBugs(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}

	creatorID := c.GetString("userId")
	err := services.DeleteBugs(ids, creatorID, currentDataPermission(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

// GetWarehouseZoneList 获取库区列表
// @Summary 获取库区列表
// @Description 分页查询指定仓库下的库区基础资料，支持按库区编码/名称和库区类型筛选
// @Tags 进销存/库区管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "库区编码或库区名称"
// @Param zoneType query string false "库区类型：NORMAL普通区 PENDING_INSPECTION待验区 QUALIFIED合格品区 UNQUALIFIED不合格品区 RETURNED退货区"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpWarehouseZoneResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones [get]
func (ctrl *ErpWarehouseController) GetWarehouseZoneList(c *gin.Context) {
	var req models.ErpWarehouseZoneListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseZoneList(c.Param("warehouseId"), req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouseZone 获取库区详情
// @Summary 获取库区详情
// @Description 根据仓库ID和库区ID获取库区基础资料详情
// @Tags 进销存/库区管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Success 200 {object} models.Response{data=models.ErpWarehouseZoneResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId} [get]
func (ctrl *ErpWarehouseController) GetWarehouseZone(c *gin.Context) {
	result, err := ctrl.warehouseService.GetWarehouseZoneDetail(c.Param("warehouseId"), c.Param("zoneId"))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouseZoneOptions 获取库区选项
// @Summary 获取库区选项
// @Description 获取指定仓库下的库区选项；需要登录，不做库区按钮权限校验
// @Tags 进销存/库区管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param keyword query string false "库区编码或库区名称"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ErpWarehouseZoneOptionResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/options [get]
func (ctrl *ErpWarehouseController) GetWarehouseZoneOptions(c *gin.Context) {
	var req models.ErpWarehouseZoneOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseZoneOptions(c.Param("warehouseId"), req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateWarehouseZone 新增库区
// @Summary 新增库区
// @Description 在指定仓库下创建库区基础资料，库区编码由后端自动生成
// @Tags 进销存/库区管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param request body models.SaveErpWarehouseZoneRequest true "库区"
// @Success 200 {object} models.Response{data=models.ErpWarehouseZoneResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones [post]
func (ctrl *ErpWarehouseController) CreateWarehouseZone(c *gin.Context) {
	var req models.SaveErpWarehouseZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.CreateWarehouseZone(c.Param("warehouseId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateWarehouseZone 更新库区
// @Summary 更新库区
// @Description 根据仓库ID和库区ID更新库区基础资料，库区编码不可修改
// @Tags 进销存/库区管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param request body models.SaveErpWarehouseZoneRequest true "库区"
// @Success 200 {object} models.Response{data=models.ErpWarehouseZoneResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId} [put]
func (ctrl *ErpWarehouseController) UpdateWarehouseZone(c *gin.Context) {
	var req models.SaveErpWarehouseZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.UpdateWarehouseZone(c.Param("warehouseId"), c.Param("zoneId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteWarehouseZone 删除库区
// @Summary 删除库区
// @Description 根据仓库ID和库区ID软删除库区基础资料，第一版不校验库存或单据引用
// @Tags 进销存/库区管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param request body models.DeleteErpWarehouseZoneRequest true "删除参数"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId} [delete]
func (ctrl *ErpWarehouseController) DeleteWarehouseZone(c *gin.Context) {
	var req models.DeleteErpWarehouseZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.warehouseService.DeleteWarehouseZone(c.Param("warehouseId"), c.Param("zoneId"), req, erpWarehouseOperatorID(c)); err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// GetWarehouseLocationList 获取货位列表
// @Summary 获取货位列表
// @Description 分页查询指定仓库和库区下的货位基础资料，支持按货位编码/名称筛选
// @Tags 进销存/货位管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "货位编码或货位名称"
// @Param sorts query string false "排序"
// @Success 200 {object} models.Response{data=utils.PaginationResponse{items=[]models.ErpWarehouseLocationResponse}} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations [get]
func (ctrl *ErpWarehouseController) GetWarehouseLocationList(c *gin.Context) {
	var req models.ErpWarehouseLocationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseLocationList(c.Param("warehouseId"), c.Param("zoneId"), req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouseLocation 获取货位详情
// @Summary 获取货位详情
// @Description 根据仓库ID、库区ID和货位ID获取货位基础资料详情
// @Tags 进销存/货位管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param locationId path string true "货位ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440002)
// @Success 200 {object} models.Response{data=models.ErpWarehouseLocationResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations/{locationId} [get]
func (ctrl *ErpWarehouseController) GetWarehouseLocation(c *gin.Context) {
	result, err := ctrl.warehouseService.GetWarehouseLocationDetail(c.Param("warehouseId"), c.Param("zoneId"), c.Param("locationId"))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetWarehouseLocationOptions 获取货位选项
// @Summary 获取货位选项
// @Description 获取指定仓库和库区下的货位选项；需要登录，不做货位按钮权限校验
// @Tags 进销存/货位管理
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param keyword query string false "货位编码或货位名称"
// @Param pageSize query int false "返回数量"
// @Success 200 {object} models.Response{data=[]models.ErpWarehouseLocationOptionResponse} "获取成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations/options [get]
func (ctrl *ErpWarehouseController) GetWarehouseLocationOptions(c *gin.Context) {
	var req models.ErpWarehouseLocationOptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.GetWarehouseLocationOptions(c.Param("warehouseId"), c.Param("zoneId"), req)
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateWarehouseLocation 新增货位
// @Summary 新增货位
// @Description 在指定仓库和库区下创建货位基础资料，货位编码由后端自动生成
// @Tags 进销存/货位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param request body models.SaveErpWarehouseLocationRequest true "货位"
// @Success 200 {object} models.Response{data=models.ErpWarehouseLocationResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations [post]
func (ctrl *ErpWarehouseController) CreateWarehouseLocation(c *gin.Context) {
	var req models.SaveErpWarehouseLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.CreateWarehouseLocation(c.Param("warehouseId"), c.Param("zoneId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// UpdateWarehouseLocation 更新货位
// @Summary 更新货位
// @Description 根据仓库ID、库区ID和货位ID更新货位基础资料，货位编码不可修改
// @Tags 进销存/货位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param locationId path string true "货位ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440002)
// @Param request body models.SaveErpWarehouseLocationRequest true "货位"
// @Success 200 {object} models.Response{data=models.ErpWarehouseLocationResponse}
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations/{locationId} [put]
func (ctrl *ErpWarehouseController) UpdateWarehouseLocation(c *gin.Context) {
	var req models.SaveErpWarehouseLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	result, err := ctrl.warehouseService.UpdateWarehouseLocation(c.Param("warehouseId"), c.Param("zoneId"), c.Param("locationId"), req, erpWarehouseOperatorID(c))
	if err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// DeleteWarehouseLocation 删除货位
// @Summary 删除货位
// @Description 根据仓库ID、库区ID和货位ID软删除货位基础资料，第一版不校验库存或单据引用
// @Tags 进销存/货位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param warehouseId path string true "仓库ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440000)
// @Param zoneId path string true "库区ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440001)
// @Param locationId path string true "货位ID，UUID格式" example(550e8400-e29b-41d4-a716-446655440002)
// @Param request body models.DeleteErpWarehouseLocationRequest true "删除参数"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未登录或Token无效"
// @Failure 403 {object} models.Response "无接口权限"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /erp/warehouses/{warehouseId}/zones/{zoneId}/locations/{locationId} [delete]
func (ctrl *ErpWarehouseController) DeleteWarehouseLocation(c *gin.Context) {
	var req models.DeleteErpWarehouseLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(nil, "参数错误"))
		return
	}
	if err := ctrl.warehouseService.DeleteWarehouseLocation(c.Param("warehouseId"), c.Param("zoneId"), c.Param("locationId"), req, erpWarehouseOperatorID(c)); err != nil {
		writeErpWarehouseError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

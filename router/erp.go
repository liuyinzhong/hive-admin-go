package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerErpRoutes 注册 ERP 模块路由，包括仓库、库区库位、库存、采购入库、采购订单和其它出库。
func registerErpRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	erpWarehouseController := controllers.NewErpWarehouseController()
	erpInventoryController := controllers.NewErpInventoryController()
	erpPurchaseOrderController := controllers.NewErpPurchaseOrderController()
	erpPurchaseInboundController := controllers.NewErpPurchaseInboundController()
	erpOtherOutboundController := controllers.NewErpOtherOutboundController()
	permissionGuard := deps.PermissionGuard

	erp := api.Group("/erp", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		warehouses := erp.Group("/warehouses")
		{
			warehouses.GET("", permissionGuard.Require("erp:warehouse:list"), erpWarehouseController.GetWarehouseList)
			warehouses.GET("/options", erpWarehouseController.GetWarehouseOptions)
			warehouses.POST("", permissionGuard.Require("erp:warehouse:create"), erpWarehouseController.CreateWarehouse)
			warehouses.GET("/:warehouseId/zones", permissionGuard.Require("erp:warehouseZone:list"), erpWarehouseController.GetWarehouseZoneList)
			warehouses.GET("/:warehouseId/zones/options", erpWarehouseController.GetWarehouseZoneOptions)
			warehouses.POST("/:warehouseId/zones", permissionGuard.Require("erp:warehouseZone:create"), erpWarehouseController.CreateWarehouseZone)
			warehouses.GET("/:warehouseId/zones/:zoneId/locations", permissionGuard.Require("erp:warehouseLocation:list"), erpWarehouseController.GetWarehouseLocationList)
			warehouses.GET("/:warehouseId/zones/:zoneId/locations/options", erpWarehouseController.GetWarehouseLocationOptions)
			warehouses.POST("/:warehouseId/zones/:zoneId/locations", permissionGuard.Require("erp:warehouseLocation:create"), erpWarehouseController.CreateWarehouseLocation)
			warehouses.GET("/:warehouseId/zones/:zoneId/locations/:locationId", permissionGuard.Require("erp:warehouseLocation:detail"), erpWarehouseController.GetWarehouseLocation)
			warehouses.PUT("/:warehouseId/zones/:zoneId/locations/:locationId", permissionGuard.Require("erp:warehouseLocation:update"), erpWarehouseController.UpdateWarehouseLocation)
			warehouses.DELETE("/:warehouseId/zones/:zoneId/locations/:locationId", permissionGuard.Require("erp:warehouseLocation:delete"), erpWarehouseController.DeleteWarehouseLocation)
			warehouses.GET("/:warehouseId/zones/:zoneId", permissionGuard.Require("erp:warehouseZone:detail"), erpWarehouseController.GetWarehouseZone)
			warehouses.PUT("/:warehouseId/zones/:zoneId", permissionGuard.Require("erp:warehouseZone:update"), erpWarehouseController.UpdateWarehouseZone)
			warehouses.DELETE("/:warehouseId/zones/:zoneId", permissionGuard.Require("erp:warehouseZone:delete"), erpWarehouseController.DeleteWarehouseZone)
			warehouses.GET("/:warehouseId", permissionGuard.Require("erp:warehouse:detail"), erpWarehouseController.GetWarehouse)
			warehouses.PUT("/:warehouseId", permissionGuard.Require("erp:warehouse:update"), erpWarehouseController.UpdateWarehouse)
			warehouses.PUT("/:warehouseId/status", permissionGuard.Require("erp:warehouse:status"), erpWarehouseController.UpdateWarehouseStatus)
			warehouses.DELETE("/:warehouseId", permissionGuard.Require("erp:warehouse:delete"), erpWarehouseController.DeleteWarehouse)
		}

		inventory := erp.Group("/inventory")
		{
			inventory.GET("/balances", permissionGuard.Require("erp:inventoryBalance:list"), erpInventoryController.GetInventoryBalanceList)
			inventory.POST("/balances/exports", permissionGuard.Require("erp:inventoryBalance:export"), erpInventoryController.CreateInventoryBalanceExport)
			inventory.GET("/balances/:balanceId/movements", permissionGuard.Require("erp:inventoryMovement:list"), erpInventoryController.GetInventoryMovements)
			inventory.GET("/movements", permissionGuard.Require("erp:inventorySourceMovement:list"), erpInventoryController.GetInventoryMovementsBySource)
			inventory.GET("/traceCodes", permissionGuard.Require("erp:inventoryTraceCode:list"), erpInventoryController.GetInventoryTraceCodeList)
			inventory.GET("/traceCodes/:traceId/movements", permissionGuard.Require("erp:inventoryTraceCode:movements"), erpInventoryController.GetInventoryTraceCodeMovements)
			inventory.POST("/initialStocks", permissionGuard.Require("erp:inventoryInitial:create"), erpInventoryController.CreateInitialStocks)
		}

		purchaseInbounds := erp.Group("/purchaseInbounds")
		{
			purchaseInbounds.GET("", permissionGuard.Require("erp:purchaseInbound:list"), erpPurchaseInboundController.GetPurchaseInboundList)
			purchaseInbounds.POST("", permissionGuard.Require("erp:purchaseInbound:create"), erpPurchaseInboundController.CreatePurchaseInbound)
			purchaseInbounds.GET("/:inboundId", permissionGuard.Require("erp:purchaseInbound:detail"), erpPurchaseInboundController.GetPurchaseInboundDetail)
		}

		purchaseOrders := erp.Group("/purchaseOrders")
		{
			purchaseOrders.GET("", permissionGuard.Require("erp:purchaseOrder:list"), erpPurchaseOrderController.GetPurchaseOrderList)
			purchaseOrders.POST("", permissionGuard.Require("erp:purchaseOrder:create"), erpPurchaseOrderController.CreatePurchaseOrder)
			purchaseOrders.GET("/:purchaseOrderId", permissionGuard.Require("erp:purchaseOrder:detail"), erpPurchaseOrderController.GetPurchaseOrderDetail)
			purchaseOrders.PUT("/:purchaseOrderId", permissionGuard.Require("erp:purchaseOrder:update"), erpPurchaseOrderController.UpdatePurchaseOrder)
			purchaseOrders.POST("/:purchaseOrderId/confirm", permissionGuard.Require("erp:purchaseOrder:confirm"), erpPurchaseOrderController.ConfirmPurchaseOrder)
			purchaseOrders.POST("/:purchaseOrderId/cancel", permissionGuard.Require("erp:purchaseOrder:cancel"), erpPurchaseOrderController.CancelPurchaseOrder)
			purchaseOrders.POST("/:purchaseOrderId/close", permissionGuard.Require("erp:purchaseOrder:close"), erpPurchaseOrderController.ClosePurchaseOrder)
			purchaseOrders.GET("/:purchaseOrderId/logs", permissionGuard.Require("erp:purchaseOrder:logs"), erpPurchaseOrderController.GetPurchaseOrderLogs)
		}

		otherOutbounds := erp.Group("/otherOutbounds")
		{
			otherOutbounds.GET("", permissionGuard.Require("erp:otherOutbound:list"), erpOtherOutboundController.GetOtherOutboundList)
			otherOutbounds.POST("", permissionGuard.Require("erp:otherOutbound:create"), erpOtherOutboundController.CreateOtherOutbound)
			otherOutbounds.GET("/:outboundId", permissionGuard.Require("erp:otherOutbound:detail"), erpOtherOutboundController.GetOtherOutboundDetail)
		}
	}
}

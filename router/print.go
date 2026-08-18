package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerPrintRoutes 注册打印模块路由，包括打印模板管理和采购入库单打印文档。
func registerPrintRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	printTemplateController := controllers.NewPrintTemplateController()
	printDocumentController := controllers.NewPrintDocumentController()
	permissionGuard := deps.PermissionGuard

	printTemplates := api.Group("/printTemplates", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		printTemplates.GET("", permissionGuard.Require("print:template:list"), printTemplateController.GetPrintTemplateList)
		printTemplates.GET("/metadata", permissionGuard.Require("print:template:metadata"), printTemplateController.GetPrintTemplateMetadata)
		printTemplates.POST("", permissionGuard.Require("print:template:create"), printTemplateController.CreatePrintTemplate)
		printTemplates.GET("/:templateId", permissionGuard.Require("print:template:detail"), printTemplateController.GetPrintTemplateDetail)
		printTemplates.PUT("/:templateId", permissionGuard.Require("print:template:update"), printTemplateController.UpdatePrintTemplate)
		printTemplates.POST("/:templateId/publish", permissionGuard.Require("print:template:publish"), printTemplateController.PublishPrintTemplate)
		printTemplates.DELETE("/:templateId", permissionGuard.Require("print:template:delete"), printTemplateController.DeletePrintTemplate)
	}

	printDocuments := api.Group("/printDocuments", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		printDocuments.GET("/purchaseInbound/:inboundId/data", permissionGuard.Require("print:template:preview"), printDocumentController.GetPurchaseInboundPrintData)
		printDocuments.GET("/purchaseInbound/:inboundId", permissionGuard.Require("print:purchaseInbound:print"), printDocumentController.GetPurchaseInboundPrintDocument)
	}
}

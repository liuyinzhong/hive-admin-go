package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerBaseRoutes 注册基础数据模块路由，包括机构信息、企业和分类体系。
func registerBaseRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	baseEnterpriseController := controllers.NewBaseEnterpriseController()
	baseInstitutionController := controllers.NewBaseInstitutionController()
	classificationController := controllers.NewClassificationController()
	permissionGuard := deps.PermissionGuard

	base := api.Group("/base", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		institution := base.Group("/institution")
		{
			institution.GET("", permissionGuard.Require("base:institution:detail"), baseInstitutionController.GetInstitution)
			institution.PUT("", permissionGuard.Require("base:institution:update"), baseInstitutionController.SaveInstitution)
		}

		enterprises := base.Group("/enterprises")
		{
			enterprises.GET("", permissionGuard.Require("base:enterprise:list"), baseEnterpriseController.GetEnterpriseList)
			enterprises.GET("/options", baseEnterpriseController.GetEnterpriseOptions)
			enterprises.POST("", permissionGuard.Require("base:enterprise:create"), baseEnterpriseController.CreateEnterprise)
			enterprises.GET("/:enterpriseId", permissionGuard.Require("base:enterprise:detail"), baseEnterpriseController.GetEnterprise)
			enterprises.PUT("/:enterpriseId", permissionGuard.Require("base:enterprise:update"), baseEnterpriseController.UpdateEnterprise)
			enterprises.PUT("/:enterpriseId/status", permissionGuard.Require("base:enterprise:status"), baseEnterpriseController.UpdateEnterpriseStatus)
		}

		classificationSystems := base.Group("/classificationSystems")
		{
			classificationSystems.GET("", permissionGuard.Require("base:classificationSystem:list"), classificationController.GetClassificationSystemList)
			classificationSystems.POST("", permissionGuard.Require("base:classificationSystem:create"), classificationController.CreateClassificationSystem)
			// 公共选项接口：需登录但无按钮权限
			classificationSystems.GET("/options", classificationController.GetClassificationNodeOptions)
			// 节点资源
			classificationSystems.GET("/nodes", permissionGuard.Require("base:classificationNode:list"), classificationController.GetClassificationNodeTree)
			classificationSystems.POST("/nodes", permissionGuard.Require("base:classificationNode:create"), classificationController.CreateClassificationNode)
			classificationSystems.GET("/nodes/:id", permissionGuard.Require("base:classificationNode:detail"), classificationController.GetClassificationNode)
			classificationSystems.PUT("/nodes/:id", permissionGuard.Require("base:classificationNode:update"), classificationController.UpdateClassificationNode)
			classificationSystems.PUT("/nodes/:id/status", permissionGuard.Require("base:classificationNode:status"), classificationController.UpdateClassificationNodeStatus)
			classificationSystems.DELETE("/nodes/:id", permissionGuard.Require("base:classificationNode:delete"), classificationController.DeleteClassificationNode)
			// 体系详情/修改/删除
			classificationSystems.GET("/:id", permissionGuard.Require("base:classificationSystem:detail"), classificationController.GetClassificationSystem)
			classificationSystems.PUT("/:id", permissionGuard.Require("base:classificationSystem:update"), classificationController.UpdateClassificationSystem)
			classificationSystems.DELETE("/:id", permissionGuard.Require("base:classificationSystem:delete"), classificationController.DeleteClassificationSystem)
		}
	}
}

package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerFormRoutes 注册表单模块路由，包括表单 Schema 定义管理。
func registerFormRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	formSchemaController := controllers.FormSchemaController{}
	permissionGuard := deps.PermissionGuard

	form := api.Group("/form", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		schemas := form.Group("/schemas")
		{
			schemas.GET("", permissionGuard.Require("form:schema:list"), formSchemaController.GetFormSchemas)
			schemas.GET("/all", formSchemaController.GetAllFormSchemas)
			schemas.POST("", permissionGuard.Require("form:schema:create"), formSchemaController.CreateFormSchema)
			schemas.GET("/:formSchemaId", permissionGuard.Require("form:schema:detail"), formSchemaController.GetFormSchema)
			schemas.PUT("/:formSchemaId", permissionGuard.Require("form:schema:update"), formSchemaController.UpdateFormSchema)
			schemas.DELETE("", permissionGuard.Require("form:schema:delete"), formSchemaController.DeleteFormSchemas)
		}
	}
}

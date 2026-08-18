package router

import (
	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/middleware"
	"hive-admin-go/services"

	"github.com/gin-gonic/gin"
)

// RouterDeps 各业务模块路由注册时共用的公共依赖。
type RouterDeps struct {
	// PermissionGuard 接口权限校验守卫，受保护路由通过 Require(权限码) 注册中间件。
	PermissionGuard *middleware.PermissionGuard
	// DataPermissionMiddleware 数据权限解析中间件，负责解析当前用户的数据范围。
	DataPermissionMiddleware gin.HandlerFunc
}

// SetupRouter 创建 Gin 引擎并按业务模块注册全部路由。
func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Static("/uploads", "./static/uploads")

	deps := &RouterDeps{
		PermissionGuard: middleware.NewPermissionGuard(services.NewPermissionService()),
		DataPermissionMiddleware: middleware.DataPermissionMiddleware(
			datapermission.NewResolver(datapermission.NewGormAssignmentStore(database.DB)),
		),
	}

	api := router.Group("/api")
	api.Use(middleware.AuditLogMiddleware(services.NewAuditLogService()))
	{
		registerSystemRoutes(api, deps)
		registerDevRoutes(api, deps)
		registerFormRoutes(api, deps)
		registerBaseRoutes(api, deps)
		registerErpRoutes(api, deps)
		registerPrintRoutes(api, deps)
		registerProductRoutes(api, deps)
		registerMedicalRoutes(api, deps)
		registerWorkflowRoutes(api, deps)
	}

	return router
}

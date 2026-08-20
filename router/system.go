package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerSystemRoutes 注册系统管理模块路由，包括公开接口、认证接口和系统管理（用户、菜单、角色、部门、字典、参数、支付渠道、日志、下载、消息、文件）。
func registerSystemRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	authController := controllers.NewAuthController()
	systemController := controllers.NewSystemController()
	menuMessageController := controllers.NewMenuMessageController()
	downloadTaskController := controllers.NewDownloadTaskController()
	externalPageController := controllers.NewExternalPageController()
	permissionGuard := deps.PermissionGuard
	dataPermissionMiddleware := deps.DataPermissionMiddleware

	public := api.Group("/public")
	{
		public.GET("/externalPages/:name", externalPageController.GetPublicExternalPage)
		public.GET("/downloads/preview/:token", downloadTaskController.PreviewFile)
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.GET("/profile", middleware.AuthMiddleware(), authController.GetProfile)
		auth.GET("/menus", middleware.AuthMiddleware(), authController.GetMenus)
		auth.GET("/codes", middleware.AuthMiddleware(), authController.GetAuthCodes)
		auth.POST("/logout", middleware.AuthMiddleware(), authController.Logout)
	}

	system := api.Group("/system", middleware.AuthMiddleware(), dataPermissionMiddleware)
	{
		downloads := system.Group("/downloads")
		{
			downloads.GET("", downloadTaskController.GetList)
			downloads.GET("/:id/file", downloadTaskController.DownloadFile)
			downloads.GET("/:id/preview-url", downloadTaskController.GetPreviewURL)
		}

		messages := system.Group("/messages")
		{
			messages.GET("/unreadSummary", menuMessageController.GetUnreadSummary)
			messages.GET("/stream", menuMessageController.StreamUnreadSummary)
			messages.POST("/read", menuMessageController.MarkRead)
			messages.POST("/demo", permissionGuard.Require("system:messageDemo:create"), menuMessageController.CreateDemoMessages)
		}

		system.POST("/upload", systemController.UploadFile)
		system.GET("/files", permissionGuard.Require("system:file:list"), systemController.GetFileList)

		users := system.Group("/users")
		{
			users.GET("", permissionGuard.Require("system:user:list"), systemController.GetUserList)
			users.GET("/all", systemController.GetAllUsers)
			users.POST("", permissionGuard.Require("system:user:create"), systemController.CreateUser)
			users.GET("/:userId", permissionGuard.Require("system:user:detail"), systemController.GetUserDetail)
			users.PUT("/:userId", permissionGuard.Require("system:user:update"), systemController.UpdateUser)
			users.PUT("/:userId/status", permissionGuard.Require("system:user:status"), systemController.UpdateUserStatus)
			users.DELETE("", permissionGuard.Require("system:user:delete"), systemController.DeleteUsers)
		}

		menus := system.Group("/menus")
		{
			menus.GET("", permissionGuard.Require("system:menu:list"), systemController.GetMenuTree)
			menus.GET("/nameExists", systemController.CheckMenuNameExists)
			menus.GET("/pathExists", systemController.CheckMenuPathExists)
			menus.POST("", permissionGuard.Require("system:menu:create"), systemController.CreateMenu)
			menus.GET("/:id", permissionGuard.Require("system:menu:detail"), systemController.GetMenuDetail)
			menus.PUT("/:id", permissionGuard.Require("system:menu:update"), systemController.UpdateMenu)
			menus.PUT("/:id/status", permissionGuard.Require("system:menu:status"), systemController.UpdateMenuStatus)
			menus.DELETE("", permissionGuard.Require("system:menu:delete"), systemController.DeleteMenus)
		}

		externalPages := system.Group("/externalPages")
		{
			externalPages.GET("", permissionGuard.Require("system:externalPage:list"), externalPageController.GetExternalPages)
			externalPages.POST("", permissionGuard.Require("system:externalPage:create"), externalPageController.CreateExternalPage)
			externalPages.GET("/:id", permissionGuard.Require("system:externalPage:detail"), externalPageController.GetExternalPage)
			externalPages.PUT("/:id", permissionGuard.Require("system:externalPage:update"), externalPageController.UpdateExternalPage)
			externalPages.PUT("/:id/status", permissionGuard.Require("system:externalPage:status"), externalPageController.UpdateExternalPageStatus)
			externalPages.DELETE("", permissionGuard.Require("system:externalPage:delete"), externalPageController.DeleteExternalPages)
		}

		roles := system.Group("/roles")
		{
			roles.GET("", permissionGuard.Require("system:role:list"), systemController.GetRoleList)
			roles.GET("/all", systemController.GetAllRoles)
			roles.POST("", permissionGuard.Require("system:role:create"), systemController.CreateRole)
			roles.GET("/:roleId", permissionGuard.Require("system:role:detail"), systemController.GetRoleDetail)
			roles.PUT("/:roleId", permissionGuard.Require("system:role:update"), systemController.UpdateRole)
			roles.PUT("/:roleId/status", permissionGuard.Require("system:role:status"), systemController.UpdateRoleStatus)
			roles.DELETE("", permissionGuard.Require("system:role:delete"), systemController.DeleteRoles)
		}

		depts := system.Group("/depts")
		{
			depts.GET("", permissionGuard.Require("system:dept:list"), systemController.GetDeptTree)
			depts.GET("/all", systemController.GetAllDepts)
			depts.POST("", permissionGuard.Require("system:dept:create"), systemController.CreateDept)
			depts.GET("/:deptId", permissionGuard.Require("system:dept:detail"), systemController.GetDeptDetail)
			depts.PUT("/:deptId", permissionGuard.Require("system:dept:update"), systemController.UpdateDept)
			depts.DELETE("", permissionGuard.Require("system:dept:delete"), systemController.DeleteDepts)
		}

		dicts := system.Group("/dicts")
		{
			dicts.GET("", permissionGuard.Require("system:dict:list"), systemController.GetDictTree)
			dicts.POST("", permissionGuard.Require("system:dict:create"), systemController.CreateDict)
			dicts.GET("/:id", permissionGuard.Require("system:dict:detail"), systemController.GetDictDetail)
			dicts.PUT("/:id", permissionGuard.Require("system:dict:update"), systemController.UpdateDict)
			dicts.PUT("/:id/status", permissionGuard.Require("system:dict:status"), systemController.UpdateDictStatus)
			dicts.DELETE("", permissionGuard.Require("system:dict:delete"), systemController.DeleteDicts)
		}

		params := system.Group("/params")
		{
			params.GET("", permissionGuard.Require("system:param:list"), systemController.GetParamList)
			params.POST("", permissionGuard.Require("system:param:create"), systemController.CreateParam)
			params.GET("/:id", permissionGuard.Require("system:param:detail"), systemController.GetParamDetail)
			params.PUT("/:id", permissionGuard.Require("system:param:update"), systemController.UpdateParam)
			params.DELETE("", permissionGuard.Require("system:param:delete"), systemController.DeleteParams)
			// 公共参数批量查询:需登录但无接口权限
			params.POST("/values", systemController.GetParamValues)
		}

		payChannels := system.Group("/payChannels")
		{
			payChannels.GET("", permissionGuard.Require("system:payChannel:list"), systemController.GetPayChannelList)
			payChannels.POST("", permissionGuard.Require("system:payChannel:create"), systemController.CreatePayChannel)
			payChannels.GET("/:id", permissionGuard.Require("system:payChannel:detail"), systemController.GetPayChannelDetail)
			payChannels.PUT("/:id", permissionGuard.Require("system:payChannel:update"), systemController.UpdatePayChannel)
			payChannels.DELETE("", permissionGuard.Require("system:payChannel:delete"), systemController.DeletePayChannels)
			payChannels.PUT("/:id/status", permissionGuard.Require("system:payChannel:status"), systemController.UpdatePayChannelStatus)
			payChannels.PUT("/:id/default", permissionGuard.Require("system:payChannel:update"), systemController.UpdatePayChannelDefault)
		}

		operationLogs := system.Group("/operationLogs")
		{
			operationLogs.GET("", permissionGuard.Require("system:operationLog:list"), systemController.GetOperationLogs)
			operationLogs.GET("/:logId", permissionGuard.Require("system:operationLog:detail"), systemController.GetOperationLog)
		}

		loginLogs := system.Group("/loginLogs")
		{
			loginLogs.GET("", permissionGuard.Require("system:loginLog:list"), systemController.GetLoginLogs)
			loginLogs.POST("/exports", permissionGuard.Require("system:loginLog:export"), systemController.CreateLoginLogExport)
			loginLogs.GET("/:logId", permissionGuard.Require("system:loginLog:detail"), systemController.GetLoginLog)
		}
	}
}

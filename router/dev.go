package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerDevRoutes 注册开发管理模块路由，包括项目、模块、版本、需求、任务、缺陷、变更历史和开发统计。
func registerDevRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	devController := controllers.DevController{}
	permissionGuard := deps.PermissionGuard
	dataPermissionMiddleware := deps.DataPermissionMiddleware

	dev := api.Group("/dev", middleware.AuthMiddleware(), dataPermissionMiddleware)
	{
		projects := dev.Group("/projects")
		{
			projects.GET("", permissionGuard.Require("dev:project:list"), devController.GetProjects)
			projects.POST("", permissionGuard.Require("dev:project:create"), devController.CreateProject)
			projects.GET("/:projectId", permissionGuard.Require("dev:project:detail"), devController.GetProject)
			projects.PUT("/:projectId", permissionGuard.Require("dev:project:update"), devController.UpdateProject)
		}

		modules := dev.Group("/modules")
		{
			modules.GET("", permissionGuard.Require("dev:module:list"), devController.GetModules)
			modules.POST("", permissionGuard.Require("dev:module:create"), devController.CreateModule)
			modules.GET("/:moduleId", permissionGuard.Require("dev:module:detail"), devController.GetModule)
			modules.PUT("/:moduleId", permissionGuard.Require("dev:module:update"), devController.UpdateModule)
			modules.DELETE("", permissionGuard.Require("dev:module:delete"), devController.DeleteModules)
		}

		versions := dev.Group("/versions")
		{
			versions.GET("", permissionGuard.Require("dev:version:list"), devController.GetVersions)
			versions.GET("/all", devController.GetAllVersions)
			versions.GET("/getLastVersion", permissionGuard.Require("dev:version:latest"), devController.GetLatestVersion)
			versions.POST("", permissionGuard.Require("dev:version:create"), devController.CreateVersion)
			versions.PUT("/:versionId/next", permissionGuard.Require("dev:version:advance"), devController.UpdateVersionNext)
			versions.GET("/:versionId", permissionGuard.Require("dev:version:detail"), devController.GetVersion)
			versions.PUT("/:versionId", permissionGuard.Require("dev:version:update"), devController.UpdateVersion)
			versions.DELETE("", permissionGuard.Require("dev:version:delete"), devController.DeleteVersions)
		}

		storys := dev.Group("/storys")
		{
			storys.GET("", permissionGuard.Require("dev:story:list"), devController.GetStorys)
			storys.GET("/all", devController.GetAllStorys)
			storys.POST("", permissionGuard.Require("dev:story:create"), devController.CreateStory)
			storys.POST("/batch", permissionGuard.Require("dev:story:batchCreate"), devController.CreateStorys)
			storys.GET("/:storyNum", permissionGuard.Require("dev:story:detail"), devController.GetStory)
			storys.PUT("/:storyId", permissionGuard.Require("dev:story:update"), devController.UpdateStory)
			storys.PUT("/:storyId/field", permissionGuard.Require("dev:story:fieldUpdate"), devController.UpdateStoryField)
			storys.PUT("/:storyId/next", permissionGuard.Require("dev:story:advance"), devController.UpdateStoryNext)
			storys.POST("/:storyNum/workflow", permissionGuard.Require("dev:story:update"), devController.StartStoryWorkflow)
			storys.GET("/:storyNum/workflow", permissionGuard.Require("dev:story:detail"), devController.GetStoryWorkflowBinding)
			storys.DELETE("", permissionGuard.Require("dev:story:delete"), devController.DeleteStorys)
		}

		tasks := dev.Group("/tasks")
		{
			tasks.GET("", permissionGuard.Require("dev:task:list"), devController.GetTasks)
			tasks.GET("/all", devController.GetAllTasks)
			tasks.POST("/exports", permissionGuard.Require("dev:task:export"), devController.CreateTaskExport)
			tasks.POST("", permissionGuard.Require("dev:task:create"), devController.CreateTask)
			tasks.POST("/batch", permissionGuard.Require("dev:task:batchCreate"), devController.CreateTasks)
			tasks.GET("/:taskNum", permissionGuard.Require("dev:task:detail"), devController.GetTask)
			tasks.PUT("/:taskId", permissionGuard.Require("dev:task:update"), devController.UpdateTask)
			tasks.PUT("/:taskId/field", permissionGuard.Require("dev:task:fieldUpdate"), devController.UpdateTaskField)
			tasks.PUT("/:taskId/next", permissionGuard.Require("dev:task:advance"), devController.UpdateTaskNext)
			tasks.DELETE("", permissionGuard.Require("dev:task:delete"), devController.DeleteTasks)
		}

		bugs := dev.Group("/bugs")
		{
			bugs.GET("", permissionGuard.Require("dev:bug:list"), devController.GetBugs)
			bugs.GET("/all", devController.GetAllBugs)
			bugs.POST("", permissionGuard.Require("dev:bug:create"), devController.CreateBug)
			bugs.POST("/batch", permissionGuard.Require("dev:bug:batchCreate"), devController.CreateBugs)
			bugs.GET("/:bugNum", permissionGuard.Require("dev:bug:detail"), devController.GetBug)
			bugs.PUT("/:bugId", permissionGuard.Require("dev:bug:update"), devController.UpdateBug)
			bugs.PUT("/:bugId/field", permissionGuard.Require("dev:bug:fieldUpdate"), devController.UpdateBugField)
			bugs.PUT("/:bugId/next", permissionGuard.Require("dev:bug:advance"), devController.UpdateBugNext)
			bugs.PUT("/:bugId/confirm", permissionGuard.Require("dev:bug:confirm"), devController.ConfirmBug)
			bugs.DELETE("", permissionGuard.Require("dev:bug:delete"), devController.DeleteBugs)
		}

		dev.GET("/changeHistory", permissionGuard.Require("dev:changeHistory:list"), devController.GetChangeHistory)
		dev.POST("/changeHistory", permissionGuard.Require("dev:changeHistory:create"), devController.CreateChangeHistory)
	}

	statistics := api.Group("/statistics", middleware.AuthMiddleware(), dataPermissionMiddleware)
	{
		devStatistics := statistics.Group("/dev")
		{
			devStatistics.GET("/getTaskFindDay", devController.GetTaskFindDay)
			devStatistics.GET("/getTaskFindYear", devController.GetTaskFindYear)
			devStatistics.GET("/getWorkspaceEnum", devController.GetWorkspaceEnum)
		}
	}
}

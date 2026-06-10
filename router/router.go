package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Static("/uploads", "./static/uploads")

	authController := controllers.NewAuthController()
	systemController := controllers.NewSystemController()
	devController := controllers.DevController{}

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.GET("/profile", middleware.AuthMiddleware(), authController.GetProfile)
			auth.GET("/menus", middleware.AuthMiddleware(), authController.GetMenus)
			auth.GET("/codes", middleware.AuthMiddleware(), authController.GetAuthCodes)
			auth.POST("/logout", middleware.AuthMiddleware(), authController.Logout)
		}

		system := api.Group("/system", middleware.AuthMiddleware())
		{
			system.POST("/upload", systemController.UploadFile)
			system.GET("/files", systemController.GetFileList)

			users := system.Group("/users")
			{
				users.GET("", systemController.GetUserList)
				users.GET("/all", systemController.GetAllUsers)
				users.POST("", systemController.CreateUser)
				users.GET("/:userId", systemController.GetUserDetail)
				users.PUT("/:userId", systemController.UpdateUser)
				users.PUT("/:userId/status", systemController.UpdateUserStatus)
				users.DELETE("", systemController.DeleteUsers)
			}

			menus := system.Group("/menus")
			{
				menus.GET("", systemController.GetMenuTree)
				menus.GET("/name-exists", systemController.CheckMenuNameExists)
				menus.GET("/path-exists", systemController.CheckMenuPathExists)
				menus.POST("", systemController.CreateMenu)
				menus.GET("/:id", systemController.GetMenuDetail)
				menus.PUT("/:id", systemController.UpdateMenu)
				menus.DELETE("", systemController.DeleteMenus)
			}

			roles := system.Group("/roles")
			{
				roles.GET("", systemController.GetRoleList)
				roles.GET("/all", systemController.GetAllRoles)
				roles.POST("", systemController.CreateRole)
				roles.GET("/:roleId", systemController.GetRoleDetail)
				roles.PUT("/:roleId", systemController.UpdateRole)
				roles.PUT("/:roleId/status", systemController.UpdateRoleStatus)
				roles.DELETE("", systemController.DeleteRoles)
			}

			depts := system.Group("/depts")
			{
				depts.GET("", systemController.GetDeptTree)
				depts.GET("/all", systemController.GetAllDepts)
				depts.POST("", systemController.CreateDept)
				depts.GET("/:deptId", systemController.GetDeptDetail)
				depts.PUT("/:deptId", systemController.UpdateDept)
				depts.DELETE("", systemController.DeleteDepts)
			}

			dicts := system.Group("/dicts")
			{
				dicts.GET("", systemController.GetDictTree)
				dicts.POST("", systemController.CreateDict)
				dicts.GET("/:id", systemController.GetDictDetail)
				dicts.PUT("/:id", systemController.UpdateDict)
				dicts.PUT("/:id/status", systemController.UpdateDictStatus)
				dicts.DELETE("", systemController.DeleteDicts)
			}
		}

		dev := api.Group("/dev", middleware.AuthMiddleware())
		{
			projects := dev.Group("/projects")
			{
				projects.GET("", devController.GetProjects)
				projects.POST("", devController.CreateProject)
				projects.GET("/:projectId", devController.GetProject)
				projects.PUT("/:projectId", devController.UpdateProject)
			}

			modules := dev.Group("/modules")
			{
				modules.GET("", devController.GetModules)
				modules.POST("", devController.CreateModule)
				modules.GET("/:moduleId", devController.GetModule)
				modules.PUT("/:moduleId", devController.UpdateModule)
				modules.DELETE("", devController.DeleteModules)
			}

			versions := dev.Group("/versions")
			{
				versions.GET("", devController.GetVersions)
				versions.GET("/all", devController.GetAllVersions)
				versions.GET("/getLastVersion", devController.GetLatestVersion)
				versions.POST("", devController.CreateVersion)
				versions.PUT("/:versionId/next", devController.UpdateVersionNext)
				versions.GET("/:versionId", devController.GetVersion)
				versions.PUT("/:versionId", devController.UpdateVersion)
				versions.DELETE("", devController.DeleteVersions)
			}

			storys := dev.Group("/storys")
			{
				storys.GET("", devController.GetStorys)
				storys.GET("/all", devController.GetAllStorys)
				storys.POST("", devController.CreateStory)
				storys.POST("/batch", devController.CreateStorys)
				storys.GET("/:storyNum", devController.GetStory)
				storys.PUT("/:storyId", devController.UpdateStory)
				storys.PUT("/:storyId/field", devController.UpdateStoryField)
				storys.PUT("/:storyId/next", devController.UpdateStoryNext)
				storys.DELETE("", devController.DeleteStorys)
			}

			tasks := dev.Group("/tasks")
			{
				tasks.GET("", devController.GetTasks)
				tasks.GET("/all", devController.GetAllTasks)
				tasks.POST("", devController.CreateTask)
				tasks.POST("/batch", devController.CreateTasks)
				tasks.GET("/:taskNum", devController.GetTask)
				tasks.PUT("/:taskId", devController.UpdateTask)
				tasks.PUT("/:taskId/field", devController.UpdateTaskField)
				tasks.PUT("/:taskId/next", devController.UpdateTaskNext)
				tasks.DELETE("", devController.DeleteTasks)
			}

			bugs := dev.Group("/bugs")
			{
				bugs.GET("", devController.GetBugs)
				bugs.GET("/all", devController.GetAllBugs)
				bugs.POST("", devController.CreateBug)
				bugs.POST("/batch", devController.CreateBugs)
				bugs.GET("/:bugNum", devController.GetBug)
				bugs.PUT("/:bugId", devController.UpdateBug)
				bugs.PUT("/:bugId/field", devController.UpdateBugField)
				bugs.PUT("/:bugId/next", devController.UpdateBugNext)
				bugs.PUT("/:bugId/confirm", devController.ConfirmBug)
				bugs.DELETE("", devController.DeleteBugs)
			}

			dev.GET("/changeHistory", devController.GetChangeHistory)
			dev.POST("/changeHistory", devController.CreateChangeHistory)

			nodes := dev.Group("/nodes")
			{
				nodes.GET("", devController.GetNodes)
				nodes.POST("", devController.CreateNode)
				nodes.DELETE("", devController.DeleteNodes)
				nodes.PUT("/:nodeId/approve", devController.ApproveNode)
				nodes.PUT("/:nodeId/next", devController.NextNode)
			}
		}
	}

	return router
}

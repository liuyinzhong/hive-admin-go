package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"
	"hive-admin-go/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Static("/uploads", "./static/uploads")

	authController := controllers.NewAuthController()
	systemController := controllers.NewSystemController()
	externalPageController := controllers.NewExternalPageController()
	devController := controllers.DevController{}
	workflowController := controllers.WorkflowController{}
	formSchemaController := controllers.FormSchemaController{}
	medicalController := controllers.NewMedicalController()
	permissionGuard := middleware.NewPermissionGuard(services.NewPermissionService())
	auditLogService := services.NewAuditLogService()

	api := router.Group("/api")
	api.Use(middleware.AuditLogMiddleware(auditLogService))
	{
		public := api.Group("/public")
		{
			public.GET("/externalPages/:name", externalPageController.GetPublicExternalPage)
		}

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

			operationLogs := system.Group("/operationLogs")
			{
				operationLogs.GET("", permissionGuard.Require("system:operationLog:list"), systemController.GetOperationLogs)
				operationLogs.GET("/:logId", permissionGuard.Require("system:operationLog:detail"), systemController.GetOperationLog)
			}

			loginLogs := system.Group("/loginLogs")
			{
				loginLogs.GET("", permissionGuard.Require("system:loginLog:list"), systemController.GetLoginLogs)
				loginLogs.GET("/:logId", permissionGuard.Require("system:loginLog:detail"), systemController.GetLoginLog)
			}
		}

		dev := api.Group("/dev", middleware.AuthMiddleware())
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
				storys.DELETE("", permissionGuard.Require("dev:story:delete"), devController.DeleteStorys)
			}

			tasks := dev.Group("/tasks")
			{
				tasks.GET("", permissionGuard.Require("dev:task:list"), devController.GetTasks)
				tasks.GET("/all", devController.GetAllTasks)
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

		form := api.Group("/form", middleware.AuthMiddleware())
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

		medical := api.Group("/medical", middleware.AuthMiddleware())
		{
			departments := medical.Group("/departments")
			{
				departments.GET("", permissionGuard.Require("medical:department:list"), medicalController.GetMedicalDepartmentTree)
				departments.GET("/all", medicalController.GetAllMedicalDepartments)
				departments.POST("", permissionGuard.Require("medical:department:create"), medicalController.CreateMedicalDepartment)
				departments.GET("/:departmentId", permissionGuard.Require("medical:department:detail"), medicalController.GetMedicalDepartmentDetail)
				departments.PUT("/:departmentId", permissionGuard.Require("medical:department:update"), medicalController.UpdateMedicalDepartment)
				departments.PUT("/:departmentId/status", permissionGuard.Require("medical:department:status"), medicalController.UpdateMedicalDepartmentStatus)
				departments.DELETE("", permissionGuard.Require("medical:department:delete"), medicalController.DeleteMedicalDepartments)
			}

			doctors := medical.Group("/doctors")
			{
				doctors.GET("", permissionGuard.Require("medical:doctor:list"), medicalController.GetDoctorList)
				doctors.GET("/all", medicalController.GetAllDoctors)
				doctors.POST("", permissionGuard.Require("medical:doctor:create"), medicalController.CreateDoctor)
				doctors.GET("/:doctorId", permissionGuard.Require("medical:doctor:detail"), medicalController.GetDoctorDetail)
				doctors.PUT("/:doctorId", permissionGuard.Require("medical:doctor:update"), medicalController.UpdateDoctor)
				doctors.PUT("/:doctorId/status", permissionGuard.Require("medical:doctor:status"), medicalController.UpdateDoctorStatus)
				doctors.DELETE("", permissionGuard.Require("medical:doctor:delete"), medicalController.DeleteDoctors)
			}

			registrationFeeRules := medical.Group("/registrationFeeRules")
			{
				registrationFeeRules.GET("", permissionGuard.Require("medical:registrationFee:list"), medicalController.GetRegistrationFeeRuleList)
				registrationFeeRules.POST("", permissionGuard.Require("medical:registrationFee:create"), medicalController.CreateRegistrationFeeRule)
				registrationFeeRules.POST("/:feeRuleId/adjustments", permissionGuard.Require("medical:registrationFee:adjust"), medicalController.AdjustRegistrationFeeRule)
			}

			scheduleTemplates := medical.Group("/scheduleTemplates")
			{
				scheduleTemplates.GET("", permissionGuard.Require("medical:scheduleTemplate:list"), medicalController.GetScheduleTemplateList)
				scheduleTemplates.POST("", permissionGuard.Require("medical:scheduleTemplate:create"), medicalController.CreateScheduleTemplate)
				scheduleTemplates.PUT("/:templateId", permissionGuard.Require("medical:scheduleTemplate:update"), medicalController.UpdateScheduleTemplate)
				scheduleTemplates.PUT("/:templateId/status", permissionGuard.Require("medical:scheduleTemplate:status"), medicalController.UpdateScheduleTemplateStatus)
				scheduleTemplates.DELETE("/:templateId", permissionGuard.Require("medical:scheduleTemplate:delete"), medicalController.DeleteScheduleTemplate)
			}

			schedules := medical.Group("/schedules")
			{
				schedules.GET("", permissionGuard.Require("medical:schedule:list"), medicalController.GetScheduleList)
				schedules.POST("", permissionGuard.Require("medical:schedule:create"), medicalController.CreateSchedule)
				schedules.DELETE("", permissionGuard.Require("medical:schedule:delete"), medicalController.DeleteDraftSchedules)
				schedules.POST("/generate", permissionGuard.Require("medical:schedule:generate"), medicalController.GenerateSchedules)
				schedules.POST("/publish", permissionGuard.Require("medical:schedule:publish"), medicalController.PublishSchedules)
				schedules.PUT("/:scheduleId", permissionGuard.Require("medical:schedule:update"), medicalController.UpdateSchedule)
				schedules.PUT("/:scheduleId/stop", permissionGuard.Require("medical:schedule:stop"), medicalController.StopSchedule)
				schedules.PUT("/:scheduleId/finish", permissionGuard.Require("medical:schedule:finish"), medicalController.FinishSchedule)
			}

			scheduleTasks := medical.Group("/scheduleTasks")
			{
				scheduleTasks.GET("", permissionGuard.Require("medical:scheduleTask:list"), medicalController.GetScheduleAutoTaskList)
			}
		}

		workflow := api.Group("/workflow", middleware.AuthMiddleware())
		{
			definitions := workflow.Group("/definitions")
			{
				definitions.GET("", permissionGuard.Require("workflow:definition:list"), workflowController.GetWorkflowDefinitions)
				definitions.GET("/all", workflowController.GetAllWorkflowDefinitions)
				definitions.POST("", permissionGuard.Require("workflow:definition:create"), workflowController.CreateWorkflowDefinition)
				definitions.GET("/:definitionId", permissionGuard.Require("workflow:definition:detail"), workflowController.GetWorkflowDefinition)
				definitions.PUT("/:definitionId", permissionGuard.Require("workflow:definition:update"), workflowController.UpdateWorkflowDefinition)
				definitions.PUT("/:definitionId/canvas", permissionGuard.Require("workflow:definition:canvasUpdate"), workflowController.UpdateWorkflowCanvas)
				definitions.PUT("/:definitionId/formSchema", permissionGuard.Require("workflow:definition:formSchemaUpdate"), workflowController.UpdateWorkflowFormSchema)
				definitions.PUT("/:definitionId/publish", permissionGuard.Require("workflow:definition:publish"), workflowController.PublishWorkflowDefinition)
				definitions.PUT("/:definitionId/status", permissionGuard.Require("workflow:definition:status"), workflowController.UpdateWorkflowDefinitionStatus)
				definitions.DELETE("", permissionGuard.Require("workflow:definition:delete"), workflowController.DeleteWorkflowDefinitions)
			}
			instances := workflow.Group("/instances")
			{
				instances.GET("", permissionGuard.Require("workflow:instance:list"), workflowController.GetWorkflowInstances)
				instances.POST("", permissionGuard.Require("workflow:instance:start"), workflowController.StartWorkflowInstance)
				instances.GET("/:instanceId", permissionGuard.Require("workflow:instance:detail"), workflowController.GetWorkflowInstanceDetail)
				instances.PUT("/:instanceId/cancel", permissionGuard.Require("workflow:instance:cancel"), workflowController.CancelWorkflowInstance)
			}
			tasks := workflow.Group("/tasks")
			{
				tasks.GET("", permissionGuard.Require("workflow:task:list"), workflowController.GetWorkflowTasks)
				tasks.PUT("/:taskId/approve", permissionGuard.Require("workflow:task:approve"), workflowController.ApproveWorkflowTask)
				tasks.PUT("/:taskId/reject", permissionGuard.Require("workflow:task:reject"), workflowController.RejectWorkflowTask)
				tasks.PUT("/:taskId/transfer", permissionGuard.Require("workflow:task:transfer"), workflowController.TransferWorkflowTask)
				tasks.PUT("/:taskId/addSign", permissionGuard.Require("workflow:task:addSign"), workflowController.AddWorkflowTaskSign)
				tasks.PUT("/:taskId/removeSign", permissionGuard.Require("workflow:task:removeSign"), workflowController.RemoveWorkflowTaskSign)
				tasks.GET("/:taskId/returnTargets", permissionGuard.Require("workflow:task:returnTargetList"), workflowController.GetWorkflowTaskReturnTargets)
				tasks.PUT("/:taskId/return", permissionGuard.Require("workflow:task:return"), workflowController.ReturnWorkflowTask)
			}
			copies := workflow.Group("/copies")
			{
				copies.GET("", permissionGuard.Require("workflow:copy:list"), workflowController.GetWorkflowCopies)
				copies.PUT("/:copyId/read", permissionGuard.Require("workflow:copy:read"), workflowController.ReadWorkflowCopy)
			}
		}
	}

	return router
}

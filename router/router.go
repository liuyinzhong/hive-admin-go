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
	menuMessageController := controllers.NewMenuMessageController()
	downloadTaskController := controllers.NewDownloadTaskController()
	externalPageController := controllers.NewExternalPageController()
	devController := controllers.DevController{}
	workflowController := controllers.WorkflowController{}
	formSchemaController := controllers.FormSchemaController{}
	medicalController := controllers.NewMedicalController()
	baseEnterpriseController := controllers.NewBaseEnterpriseController()
	baseInstitutionController := controllers.NewBaseInstitutionController()
	erpWarehouseController := controllers.NewErpWarehouseController()
	erpInventoryController := controllers.NewErpInventoryController()
	erpPurchaseOrderController := controllers.NewErpPurchaseOrderController()
	erpPurchaseInboundController := controllers.NewErpPurchaseInboundController()
	erpOtherOutboundController := controllers.NewErpOtherOutboundController()
	printTemplateController := controllers.NewPrintTemplateController()
	printDocumentController := controllers.NewPrintDocumentController()
	productSpuController := controllers.NewProductSpuController()
	productRpController := controllers.NewProductRpController()
	productMpController := controllers.NewProductMpController()
	productSkuController := controllers.NewProductSkuController()
	productSkuPriceController := controllers.NewProductSkuPriceController()
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
			downloads := system.Group("/downloads")
			{
				downloads.GET("", downloadTaskController.GetList)
				downloads.GET("/:id/file", downloadTaskController.DownloadFile)
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

		statistics := api.Group("/statistics", middleware.AuthMiddleware())
		{
			devStatistics := statistics.Group("/dev")
			{
				devStatistics.GET("/getTaskFindDay", devController.GetTaskFindDay)
				devStatistics.GET("/getTaskFindYear", devController.GetTaskFindYear)
				devStatistics.GET("/getWorkspaceEnum", devController.GetWorkspaceEnum)
			}
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

		base := api.Group("/base", middleware.AuthMiddleware())
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

			classificationController := controllers.NewClassificationController()
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

		erp := api.Group("/erp", middleware.AuthMiddleware())
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

		printTemplates := api.Group("/printTemplates", middleware.AuthMiddleware())
		{
			printTemplates.GET("", permissionGuard.Require("print:template:list"), printTemplateController.GetPrintTemplateList)
			printTemplates.GET("/metadata", permissionGuard.Require("print:template:metadata"), printTemplateController.GetPrintTemplateMetadata)
			printTemplates.POST("", permissionGuard.Require("print:template:create"), printTemplateController.CreatePrintTemplate)
			printTemplates.GET("/:templateId", permissionGuard.Require("print:template:detail"), printTemplateController.GetPrintTemplateDetail)
			printTemplates.PUT("/:templateId", permissionGuard.Require("print:template:update"), printTemplateController.UpdatePrintTemplate)
			printTemplates.POST("/:templateId/publish", permissionGuard.Require("print:template:publish"), printTemplateController.PublishPrintTemplate)
			printTemplates.DELETE("/:templateId", permissionGuard.Require("print:template:delete"), printTemplateController.DeletePrintTemplate)
		}

		printDocuments := api.Group("/printDocuments", middleware.AuthMiddleware())
		{
			printDocuments.GET("/purchaseInbound/:inboundId/data", permissionGuard.Require("print:template:preview"), printDocumentController.GetPurchaseInboundPrintData)
			printDocuments.GET("/purchaseInbound/:inboundId", permissionGuard.Require("print:purchaseInbound:print"), printDocumentController.GetPurchaseInboundPrintDocument)
		}

		product := api.Group("/product", middleware.AuthMiddleware())
		{
			spus := product.Group("/spus")
			{
				spus.GET("", permissionGuard.Require("product:spu:list"), productSpuController.GetProductSpuList)
				spus.GET("/options", productSpuController.GetProductSpuOptions)
				spus.POST("", permissionGuard.Require("product:spu:create"), productSpuController.CreateProductSpu)
				spus.GET("/:spuId", permissionGuard.Require("product:spu:detail"), productSpuController.GetProductSpu)
				spus.PUT("/:spuId", permissionGuard.Require("product:spu:update"), productSpuController.UpdateProductSpu)
				spus.PUT("/:spuId/status", permissionGuard.Require("product:spu:status"), productSpuController.UpdateProductSpuStatus)
			}
			rps := product.Group("/rps")
			{
				rps.GET("", permissionGuard.Require("product:rp:list"), productRpController.GetProductRpList)
				rps.GET("/options", productRpController.GetProductRpOptions)
				rps.POST("", permissionGuard.Require("product:rp:create"), productRpController.CreateProductRp)
				rps.GET("/:rpId", permissionGuard.Require("product:rp:detail"), productRpController.GetProductRp)
				rps.PUT("/:rpId", permissionGuard.Require("product:rp:update"), productRpController.UpdateProductRp)
				rps.PUT("/:rpId/status", permissionGuard.Require("product:rp:status"), productRpController.UpdateProductRpStatus)
			}
			mps := product.Group("/mps")
			{
				mps.GET("", permissionGuard.Require("product:mp:list"), productMpController.GetProductMpList)
				mps.POST("", permissionGuard.Require("product:mp:create"), productMpController.CreateProductMp)
				mps.GET("/:mpId", permissionGuard.Require("product:mp:detail"), productMpController.GetProductMp)
				mps.PUT("/:mpId", permissionGuard.Require("product:mp:update"), productMpController.UpdateProductMp)
				mps.PUT("/:mpId/status", permissionGuard.Require("product:mp:status"), productMpController.UpdateProductMpStatus)
			}
			skus := product.Group("/skus")
			{
				skus.GET("", permissionGuard.Require("product:sku:list"), productSkuController.GetProductSkuList)
				skus.GET("/options", productSkuController.GetProductSkuOptions)
				skus.POST("", permissionGuard.Require("product:sku:create"), productSkuController.CreateProductSku)
				skus.GET("/:skuId/prices", permissionGuard.Require("product:skuPrice:list"), productSkuPriceController.GetProductSkuPriceList)
				skus.POST("/:skuId/prices", permissionGuard.Require("product:skuPrice:create"), productSkuPriceController.CreateProductSkuPrice)
				skus.PUT("/:skuId/prices/:priceId", permissionGuard.Require("product:skuPrice:update"), productSkuPriceController.UpdateProductSkuPrice)
				skus.PUT("/:skuId/prices/:priceId/status", permissionGuard.Require("product:skuPrice:status"), productSkuPriceController.UpdateProductSkuPriceStatus)
				skus.DELETE("/:skuId/prices/:priceId", permissionGuard.Require("product:skuPrice:delete"), productSkuPriceController.DeleteProductSkuPrice)
				skus.GET("/:skuId/prices/:priceId/tiers", permissionGuard.Require("product:skuPriceTier:list"), productSkuPriceController.GetProductSkuPriceTiers)
				skus.PUT("/:skuId/prices/:priceId/tiers", permissionGuard.Require("product:skuPriceTier:save"), productSkuPriceController.SaveProductSkuPriceTiers)
				skus.GET("/:skuId", permissionGuard.Require("product:sku:detail"), productSkuController.GetProductSku)
				skus.PUT("/:skuId", permissionGuard.Require("product:sku:update"), productSkuController.UpdateProductSku)
				skus.PUT("/:skuId/status", permissionGuard.Require("product:sku:status"), productSkuController.UpdateProductSkuStatus)
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

			patients := medical.Group("/patients")
			{
				patients.GET("", permissionGuard.Require("medical:patient:list"), medicalController.GetPatientList)
				patients.POST("", permissionGuard.Require("medical:patient:create"), medicalController.CreatePatient)
				patients.GET("/:patientId", permissionGuard.Require("medical:patient:detail"), medicalController.GetPatientDetail)
				patients.PUT("/:patientId", permissionGuard.Require("medical:patient:update"), medicalController.UpdatePatient)
				patients.PUT("/:patientId/status", permissionGuard.Require("medical:patient:status"), medicalController.UpdatePatientStatus)
			}

			registrations := medical.Group("/registrations")
			{
				registrations.GET("", permissionGuard.Require("medical:registration:list"), medicalController.GetRegistrationList)
				registrations.POST("", permissionGuard.Require("medical:registration:create"), medicalController.CreateRegistration)
				registrations.GET("/:registrationId", permissionGuard.Require("medical:registration:detail"), medicalController.GetRegistrationDetail)
				registrations.POST("/:registrationId/confirmPayment", permissionGuard.Require("medical:registration:confirmPayment"), medicalController.ConfirmRegistrationPayment)
				registrations.POST("/:registrationId/cancel", permissionGuard.Require("medical:registration:cancel"), medicalController.CancelRegistration)
				registrations.POST("/:registrationId/checkIn", permissionGuard.Require("medical:registration:checkIn"), medicalController.CheckInRegistration)
				registrations.POST("/:registrationId/complete", permissionGuard.Require("medical:registration:complete"), medicalController.CompleteRegistration)
				registrations.POST("/:registrationId/noShow", permissionGuard.Require("medical:registration:noShow"), medicalController.MarkRegistrationNoShow)
				registrations.POST("/:registrationId/refundStart", permissionGuard.Require("medical:registration:refundStart"), medicalController.StartRegistrationRefund)
				registrations.POST("/:registrationId/refundProcess", permissionGuard.Require("medical:registration:refundProcess"), medicalController.ProcessRegistrationRefund)
				registrations.POST("/:registrationId/refundComplete", permissionGuard.Require("medical:registration:refundComplete"), medicalController.CompleteRegistrationRefund)
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

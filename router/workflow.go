package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerWorkflowRoutes 注册工作流模块路由，包括流程定义、流程实例、待办任务和抄送。
func registerWorkflowRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	workflowController := controllers.WorkflowController{}
	permissionGuard := deps.PermissionGuard

	workflow := api.Group("/workflow", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
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

package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerMedicalRoutes 注册医疗模块路由，包括科室、医生、患者、诊断、挂号、挂号费、排班、医生工作台、门诊病历和处方审核。
func registerMedicalRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	medicalController := controllers.NewMedicalController()
	permissionGuard := deps.PermissionGuard

	medical := api.Group("/medical", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
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

		diagnoses := medical.Group("/diagnoses")
		{
			diagnoses.GET("", permissionGuard.Require("medical:diagnosis:list"), medicalController.GetDiagnosisList)
			diagnoses.POST("", permissionGuard.Require("medical:diagnosis:create"), medicalController.CreateDiagnosis)
			diagnoses.GET("/:diagnosisId", permissionGuard.Require("medical:diagnosis:detail"), medicalController.GetDiagnosisDetail)
			diagnoses.PUT("/:diagnosisId", permissionGuard.Require("medical:diagnosis:update"), medicalController.UpdateDiagnosis)
			diagnoses.PUT("/:diagnosisId/status", permissionGuard.Require("medical:diagnosis:status"), medicalController.UpdateDiagnosisStatus)
			diagnoses.DELETE("/:diagnosisId", permissionGuard.Require("medical:diagnosis:delete"), medicalController.DeleteDiagnosis)
		}

		registrations := medical.Group("/registrations")
		{
			registrations.GET("", permissionGuard.Require("medical:registration:list"), medicalController.GetRegistrationList)
			registrations.POST("", permissionGuard.Require("medical:registration:create"), medicalController.CreateRegistration)
			registrations.GET("/:registrationId", permissionGuard.Require("medical:registration:detail"), medicalController.GetRegistrationDetail)
			registrations.POST("/:registrationId/confirmPayment", permissionGuard.Require("medical:registration:confirmPayment"), medicalController.ConfirmRegistrationPayment)
			registrations.POST("/:registrationId/cancel", permissionGuard.Require("medical:registration:cancel"), medicalController.CancelRegistration)
			registrations.POST("/:registrationId/checkIn", permissionGuard.Require("medical:registration:checkIn"), medicalController.CheckInRegistration)
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
			schedules.GET("/:scheduleId", permissionGuard.Require("medical:schedule:detail"), medicalController.GetScheduleDetail)
			schedules.POST("", permissionGuard.Require("medical:schedule:create"), medicalController.CreateSchedule)
			schedules.DELETE("", permissionGuard.Require("medical:schedule:delete"), medicalController.DeleteDraftSchedules)
			schedules.POST("/generate", permissionGuard.Require("medical:schedule:generate"), medicalController.GenerateSchedules)
			schedules.POST("/publish", permissionGuard.Require("medical:schedule:publish"), medicalController.PublishSchedules)
			schedules.GET("/:scheduleId/visitQueues", permissionGuard.Require("medical:visitQueue:list"), medicalController.GetVisitQueueList)
			schedules.PUT("/:scheduleId", permissionGuard.Require("medical:schedule:update"), medicalController.UpdateSchedule)
			schedules.PUT("/:scheduleId/stop", permissionGuard.Require("medical:schedule:stop"), medicalController.StopSchedule)
			schedules.PUT("/:scheduleId/finish", permissionGuard.Require("medical:schedule:finish"), medicalController.FinishSchedule)
		}

		scheduleTasks := medical.Group("/scheduleTasks")
		{
			scheduleTasks.GET("", permissionGuard.Require("medical:scheduleTask:list"), medicalController.GetScheduleAutoTaskList)
		}

		doctorWorkbench := medical.Group("/doctorWorkbench", permissionGuard.Require("medical:doctorWorkbench:access"))
		{
			doctorWorkbench.GET("", medicalController.GetDoctorWorkbench)
			doctorWorkbench.GET("/diagnosisOptions", medicalController.GetDiagnosisOptions)
			doctorWorkbench.POST("/schedules/:scheduleId/callNext", medicalController.CallNextPatient)
			doctorWorkbench.POST("/queues/:queueId/repeatCall", medicalController.RepeatCallPatient)
			doctorWorkbench.POST("/queues/:queueId/pass", medicalController.PassPatient)
			doctorWorkbench.POST("/queues/:queueId/recall", medicalController.RecallPatient)
			doctorWorkbench.POST("/queues/:queueId/start", medicalController.StartConsultation)
		}

		outpatientRecords := medical.Group("/outpatientRecords", permissionGuard.Require("medical:doctorWorkbench:access"))
		{
			outpatientRecords.GET("/:recordId", medicalController.GetOutpatientRecord)
			outpatientRecords.PUT("/:recordId", medicalController.SaveOutpatientRecord)
			outpatientRecords.POST("/:recordId/complete", medicalController.CompleteOutpatientRecord)
			outpatientRecords.GET("/:recordId/history", medicalController.GetPatientOutpatientHistory)
			outpatientRecords.POST("/:recordId/prescriptions", medicalController.CreatePrescription)
		}

		prescriptions := medical.Group("/prescriptions", permissionGuard.Require("medical:doctorWorkbench:access"))
		{
			prescriptions.GET("/:prescriptionId", medicalController.GetPrescription)
			prescriptions.PUT("/:prescriptionId", medicalController.UpdatePrescription)
			prescriptions.POST("/:prescriptionId/submit", medicalController.SubmitPrescription)
			prescriptions.POST("/:prescriptionId/withdraw", medicalController.WithdrawPrescription)
			prescriptions.POST("/:prescriptionId/void", medicalController.VoidPrescription)
		}

		prescriptionReviews := medical.Group("/prescriptionReviews", permissionGuard.Require("medical:prescriptionReview:access"))
		{
			prescriptionReviews.GET("", medicalController.GetPrescriptionReviewList)
			prescriptionReviews.GET("/:prescriptionId", medicalController.GetPrescriptionReviewDetail)
			prescriptionReviews.POST("/:prescriptionId/review", medicalController.ReviewPrescription)
		}
	}
}

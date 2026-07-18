package controllers

import "hive-admin-go/services"

type SystemController struct {
	userService     *services.UserService
	menuService     *services.MenuService
	roleService     *services.RoleService
	deptService     *services.DeptService
	dictService     *services.DictService
	fileService     *services.FileService
	auditLogService *services.AuditLogService
}

func NewSystemController() *SystemController {
	return &SystemController{
		userService:     services.NewUserService(),
		menuService:     services.NewMenuService(),
		roleService:     services.NewRoleService(),
		deptService:     services.NewDeptService(),
		dictService:     services.NewDictService(),
		fileService:     services.NewFileService(),
		auditLogService: services.NewAuditLogService(),
	}
}

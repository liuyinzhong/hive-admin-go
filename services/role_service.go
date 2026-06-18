package services

import (
	"errors"
	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"time"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

func (s *RoleService) GetRoleList(req models.RoleListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.SysRole{}).Where("del_flag = 0")
	
	if req.RoleTitle != "" {
		query = query.Where("role_title LIKE ?", "%"+req.RoleTitle+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Remark != "" {
		query = query.Where("remark LIKE ?", "%"+req.Remark+"%")
	}
	if req.StartDate != "" {
		query = query.Where("create_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("create_date <= ?", req.EndDate)
	}
	
	sorts, _ := utils.ParseSortParams(req.Sorts)
	query = utils.ApplySorting(query, sorts, "create_date desc")
	
	var roles []models.SysRole
	pageResult, err := utils.Paginate(query, req.Page, req.PageSize, &roles)
	if err != nil {
		return nil, err
	}
	
	return pageResult, nil
}

func (s *RoleService) GetAllRoles() ([]*models.RoleSimpleResponse, error) {
	var roles []models.SysRole
	err := database.DB.Where("del_flag = 0").Order("create_date desc").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	
	responses := make([]*models.RoleSimpleResponse, 0)
	for _, role := range roles {
		roleTitle := ""
		if role.RoleTitle != nil {
			roleTitle = *role.RoleTitle
		}
		responses = append(responses, &models.RoleSimpleResponse{
			RoleId:    role.RoleID,
			RoleTitle: roleTitle,
			Status:    role.Status,
		})
	}
	
	return responses, nil
}

func (s *RoleService) CreateRole(req models.CreateRoleRequest) error {
	var count int64
	database.DB.Model(&models.SysRole{}).Where("role_title = ? AND del_flag = 0", req.RoleTitle).Count(&count)
	if count > 0 {
		return errors.New("角色名称已存在")
	}
	
	now := time.Now()
	role := models.SysRole{
		RoleID:     utils.GenerateUUID(),
		RoleTitle:  &req.RoleTitle,
		Status:     req.Status,
		Remark:     req.Remark,
		CreateDate: &now,
		UpdateDate: &now,
		DelFlag:    0,
	}
	
	if err := database.DB.Create(&role).Error; err != nil {
		return err
	}
	
	return s.saveRoleMenus(role.RoleID, req.Permissions)
}

func (s *RoleService) GetRoleDetail(roleId string) (*models.RoleDetailResponse, error) {
	var role models.SysRole
	err := database.DB.Where("role_id = ? AND del_flag = 0", roleId).First(&role).Error
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	
	permissions := s.getRoleMenus(roleId)
	
	roleTitle := ""
	if role.RoleTitle != nil {
		roleTitle = *role.RoleTitle
	}
	
	return &models.RoleDetailResponse{
		RoleId:      role.RoleID,
		RoleTitle:   roleTitle,
		Status:      role.Status,
		CreateDate:  models.TimeToStringPtr(role.CreateDate),
		Remark:      role.Remark,
		Permissions: permissions,
	}, nil
}

func (s *RoleService) UpdateRole(roleId string, req models.UpdateRoleRequest) error {
	var role models.SysRole
	err := database.DB.Where("role_id = ? AND del_flag = 0", roleId).First(&role).Error
	if err != nil {
		return errors.New("角色不存在")
	}
	
	var count int64
	database.DB.Model(&models.SysRole{}).Where("role_title = ? AND del_flag = 0 AND role_id != ?", req.RoleTitle, roleId).Count(&count)
	if count > 0 {
		return errors.New("角色名称已存在")
	}
	
	now := time.Now()
	role.RoleTitle = &req.RoleTitle
	role.Status = req.Status
	role.Remark = req.Remark
	role.UpdateDate = &now
	
	if err := database.DB.Save(&role).Error; err != nil {
		return err
	}
	
	return s.saveRoleMenus(roleId, req.Permissions)
}

func (s *RoleService) UpdateRoleStatus(roleId string, status int) error {
	var role models.SysRole
	err := database.DB.Where("role_id = ? AND del_flag = 0", roleId).First(&role).Error
	if err != nil {
		return errors.New("角色不存在")
	}
	
	now := time.Now()
	role.Status = status
	role.UpdateDate = &now
	
	return database.DB.Save(&role).Error
}

func (s *RoleService) DeleteRoles(roleIds []string) error {
	for _, roleId := range roleIds {
		var role models.SysRole
		err := database.DB.Where("role_id = ? AND del_flag = 0", roleId).First(&role).Error
		if err != nil {
			continue
		}
		
		database.DB.Where("role_id = ?", roleId).Delete(&models.SysRoleMenu{})
		
		now := time.Now()
		role.DelFlag = 1
		role.UpdateDate = &now
		database.DB.Save(&role)
	}
	
	return nil
}

func (s *RoleService) saveRoleMenus(roleId string, menuIds []string) error {
	database.DB.Where("role_id = ? AND del_flag = 0", roleId).Delete(&models.SysRoleMenu{})
	
	now := time.Now()
	for _, menuId := range menuIds {
		roleMenu := models.SysRoleMenu{
			ID:         utils.GenerateUUID(),
			RoleID:     roleId,
			MenuID:     menuId,
			CreateDate: &now,
			UpdateDate: &now,
			DelFlag:    0,
		}
		if err := database.DB.Create(&roleMenu).Error; err != nil {
			return err
		}
	}
	
	return nil
}

func (s *RoleService) getRoleMenus(roleId string) []string {
	var roleMenus []models.SysRoleMenu
	database.DB.Where("role_id = ? AND del_flag = 0", roleId).Find(&roleMenus)
	
	menuIds := make([]string, 0)
	for _, rm := range roleMenus {
		menuIds = append(menuIds, rm.MenuID)
	}
	
	return menuIds
}
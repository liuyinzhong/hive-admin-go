package services

import (
	"errors"
	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"time"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetUserList(req models.UserListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.SysUser{}).Where("del_flag = 0 AND is_sys = 0")

	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.RealName != "" {
		query = query.Where("real_name LIKE ?", "%"+req.RealName+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+req.Phone+"%")
	}
	if req.DeptId != "" {
		deptIds := s.getDeptAndChildren(req.DeptId)
		var userIds []string
		database.DB.Model(&models.SysUserDept{}).Where("dept_id IN ? AND del_flag = 0", deptIds).Pluck("user_id", &userIds)
		if len(userIds) > 0 {
			query = query.Where("user_id IN ?", userIds)
		} else {
			query = query.Where("1 = 0")
		}
	}

	sorts, _ := utils.ParseSortParams(req.Sorts)
	query = utils.ApplySorting(query, sorts, "create_date desc")

	var users []models.SysUser
	pageResult, err := utils.Paginate(query, req.Page, req.PageSize, &users)
	if err != nil {
		return nil, err
	}

	resultItems := make([]*models.ProfileResponse, 0)
	for _, user := range users {
		roleTitles, roleIds := s.getUserRoles(user.UserID)
		deptTitles, deptIds := s.getUserDepts(user.UserID)
		resultItems = append(resultItems, models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds))
	}

	pageResult.Items = resultItems
	return pageResult, nil
}

func (s *UserService) getDeptAndChildren(deptId string) []string {
	ids := []string{deptId}
	var children []models.SysDept
	database.DB.Where("pid = ? AND del_flag = 0", deptId).Find(&children)
	for _, child := range children {
		ids = append(ids, s.getDeptAndChildren(child.DeptID)...)
	}
	return ids
}

func (s *UserService) GetAllUsers(realName string) ([]*models.ProfileResponse, error) {
	query := database.DB.Model(&models.SysUser{}).Where("del_flag = 0 AND is_sys = 0 AND status = 1")
	
	if realName != "" {
		query = query.Where("real_name LIKE ?", "%"+realName+"%")
	}
	
	var users []models.SysUser
	err := query.Order("create_date desc").Find(&users).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]*models.ProfileResponse, 0)
	for _, user := range users {
		roleTitles, roleIds := s.getUserRoles(user.UserID)
		deptTitles, deptIds := s.getUserDepts(user.UserID)
		result = append(result, models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds))
	}
	
	return result, nil
}

func (s *UserService) CreateUser(req models.CreateUserRequest) error {
	var count int64
	database.DB.Model(&models.SysUser{}).Where("username = ? AND del_flag = 0", req.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	
	now := time.Now()
	user := models.SysUser{
		UserID:    utils.GenerateUUID(),
		Username:  &req.Username,
		RealName:  &req.RealName,
		Phone:     req.Phone,
		Password:  &req.Password,
		Desc:      req.Desc,
		Status:    1,
		CreateDate: &now,
		UpdateDate: &now,
		DelFlag:   0,
		IsSys:     0,
	}
	
	if err := database.DB.Create(&user).Error; err != nil {
		return err
	}
	
	if err := s.saveUserRoles(user.UserID, req.RoleIds); err != nil {
		return err
	}
	
	if err := s.saveUserDepts(user.UserID, req.DeptIds); err != nil {
		return err
	}
	
	return nil
}

func (s *UserService) GetUserDetail(userId string) (*models.ProfileResponse, error) {
	var user models.SysUser
	err := database.DB.Where("user_id = ? AND del_flag = 0", userId).First(&user).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	
	roleTitles, roleIds := s.getUserRoles(user.UserID)
	deptTitles, deptIds := s.getUserDepts(user.UserID)
	
	return models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds), nil
}

func (s *UserService) UpdateUser(userId string, req models.UpdateUserRequest) error {
	var user models.SysUser
	err := database.DB.Where("user_id = ? AND del_flag = 0", userId).First(&user).Error
	if err != nil {
		return errors.New("用户不存在")
	}
	
	var count int64
	database.DB.Model(&models.SysUser{}).Where("username = ? AND del_flag = 0 AND user_id != ?", req.Username, userId).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	
	now := time.Now()
	user.Username = &req.Username
	user.RealName = &req.RealName
	user.Phone = req.Phone
	user.Desc = req.Desc
	user.UpdateDate = &now
	
	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}
	
	if err := s.saveUserRoles(user.UserID, req.RoleIds); err != nil {
		return err
	}
	
	if err := s.saveUserDepts(user.UserID, req.DeptIds); err != nil {
		return err
	}
	
	return nil
}

func (s *UserService) UpdateUserStatus(userId string, status int) error {
	var user models.SysUser
	err := database.DB.Where("user_id = ? AND del_flag = 0", userId).First(&user).Error
	if err != nil {
		return errors.New("用户不存在")
	}
	
	now := time.Now()
	user.Status = status
	user.UpdateDate = &now
	
	return database.DB.Save(&user).Error
}

func (s *UserService) DeleteUsers(userIds []string, currentUserId string) error {
	for _, userId := range userIds {
		var user models.SysUser
		err := database.DB.Where("user_id = ? AND del_flag = 0", userId).First(&user).Error
		if err != nil {
			continue
		}
		
		if userId == currentUserId {
			return errors.New("不能删除当前登录用户")
		}
		
		if user.IsSys == 1 {
			return errors.New("不能删除内置用户")
		}
		
		var roleCount int64
		database.DB.Model(&models.SysUserRole{}).Where("user_id = ? AND del_flag = 0", userId).Count(&roleCount)
		if roleCount > 0 {
			database.DB.Model(&models.SysUserRole{}).Where("user_id = ?", userId).Updates(map[string]interface{}{"del_flag": 1, "update_date": time.Now()})
		}
		
		var deptCount int64
		database.DB.Model(&models.SysUserDept{}).Where("user_id = ? AND del_flag = 0", userId).Count(&deptCount)
		if deptCount > 0 {
			database.DB.Model(&models.SysUserDept{}).Where("user_id = ?", userId).Updates(map[string]interface{}{"del_flag": 1, "update_date": time.Now()})
		}
		
		now := time.Now()
		user.DelFlag = 1
		user.UpdateDate = &now
		database.DB.Save(&user)
	}
	
	return nil
}

func (s *UserService) getUserRoles(userId string) ([]string, []string) {
	var userRoles []models.SysUserRole
	database.DB.Where("user_id = ? AND del_flag = 0", userId).Find(&userRoles)
	
	var roleTitles []string
	var roleIds []string
	
	for _, ur := range userRoles {
		var role models.SysRole
		if err := database.DB.Where("role_id = ? AND del_flag = 0 AND status = 1", ur.RoleID).First(&role).Error; err == nil {
			if role.RoleTitle != nil {
				roleTitles = append(roleTitles, *role.RoleTitle)
			}
			roleIds = append(roleIds, ur.RoleID)
		}
	}
	
	return roleTitles, roleIds
}

func (s *UserService) saveUserRoles(userId string, roleIds []string) error {
	database.DB.Where("user_id = ? AND del_flag = 0", userId).Delete(&models.SysUserRole{})
	
	now := time.Now()
	for _, roleId := range roleIds {
		userRole := models.SysUserRole{
			ID:         utils.GenerateUUID(),
			UserID:     userId,
			RoleID:     roleId,
			CreateDate: &now,
			UpdateDate: &now,
			DelFlag:    0,
		}
		if err := database.DB.Create(&userRole).Error; err != nil {
			return err
		}
	}
	
	return nil
}

func (s *UserService) getUserDepts(userId string) ([]string, []string) {
	var userDepts []models.SysUserDept
	database.DB.Where("user_id = ? AND del_flag = 0", userId).Find(&userDepts)
	
	var deptTitles []string
	var deptIds []string
	
	for _, ud := range userDepts {
		var dept models.SysDept
		if err := database.DB.Where("dept_id = ? AND del_flag = 0", ud.DeptID).First(&dept).Error; err == nil {
			if dept.DeptTitle != nil {
				deptTitles = append(deptTitles, *dept.DeptTitle)
			}
			deptIds = append(deptIds, ud.DeptID)
		}
	}
	
	return deptTitles, deptIds
}

func (s *UserService) saveUserDepts(userId string, deptIds []string) error {
	database.DB.Where("user_id = ? AND del_flag = 0", userId).Delete(&models.SysUserDept{})
	
	now := time.Now()
	for _, deptId := range deptIds {
		userDept := models.SysUserDept{
			ID:         utils.GenerateUUID(),
			UserID:     userId,
			DeptID:     deptId,
			CreateDate: &now,
			UpdateDate: &now,
			DelFlag:    0,
		}
		if err := database.DB.Create(&userDept).Error; err != nil {
			return err
		}
	}
	
	return nil
}
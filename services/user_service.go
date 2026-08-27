package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetUserList(req models.UserListRequest, permission datapermission.Permission) (*utils.PageResult, error) {
	query := database.DB.Model(&models.SysUser{}).Where("del_flag = 0 AND is_sys = 0")
	query = permission.Apply(query, "sys_user.user_id")

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
	if req.RoleId != "" {
		var userIds []string
		database.DB.Model(&models.SysUserRole{}).Where("role_id = ? AND del_flag = 0", req.RoleId).Pluck("user_id", &userIds)
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
	leaderUserNames, err := s.getLeaderUserNames(users, permission)
	if err != nil {
		return nil, err
	}

	resultItems := make([]*models.ProfileResponse, 0)
	for _, user := range users {
		roleTitles, roleIds := s.getUserRoles(user.UserID)
		deptTitles, deptIds := s.getUserDepts(user.UserID, permission)
		response := models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds)
		applyVisibleLeader(response, user.LeaderUserID, leaderUserNames)
		resultItems = append(resultItems, response)
	}

	pageResult.Items = resultItems
	return pageResult, nil
}

func (s *UserService) getLeaderUserNames(users []models.SysUser, permission datapermission.Permission) (map[string]*string, error) {
	leaderUserIds := make([]string, 0)
	for _, user := range users {
		if user.LeaderUserID != nil && *user.LeaderUserID != "" {
			leaderUserIds = append(leaderUserIds, *user.LeaderUserID)
		}
	}

	leaderUserNames := make(map[string]*string)
	if len(leaderUserIds) == 0 {
		return leaderUserNames, nil
	}

	var leaders []models.SysUser
	query := database.DB.Model(&models.SysUser{}).
		Select("user_id", "real_name").
		Where("user_id IN ? AND del_flag = 0 AND is_sys = 0", leaderUserIds)
	if err := permission.Apply(query, "sys_user.user_id").
		Find(&leaders).Error; err != nil {
		return nil, err
	}
	for _, leader := range leaders {
		leaderUserNames[leader.UserID] = leader.RealName
	}
	return leaderUserNames, nil
}

func applyVisibleLeader(response *models.ProfileResponse, leaderUserID *string, leaderUserNames map[string]*string) {
	if leaderUserID == nil || *leaderUserID == "" {
		return
	}
	leaderUserName, visible := leaderUserNames[*leaderUserID]
	if !visible {
		response.LeaderUserId = nil
		response.LeaderUserName = nil
		return
	}
	response.LeaderUserName = leaderUserName
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

func (s *UserService) GetAllUsers(realName string, permission datapermission.Permission) ([]*models.ProfileResponse, error) {
	query := database.DB.Model(&models.SysUser{}).Where("del_flag = 0 AND is_sys = 0 AND status = 1")
	query = permission.Apply(query, "sys_user.user_id")

	if realName != "" {
		query = query.Where("real_name LIKE ?", "%"+realName+"%")
	}

	var users []models.SysUser
	err := query.Order("create_date desc").Find(&users).Error
	if err != nil {
		return nil, err
	}

	leaderUserNames, err := s.getLeaderUserNames(users, permission)
	if err != nil {
		return nil, err
	}

	result := make([]*models.ProfileResponse, 0)
	for _, user := range users {
		roleTitles, roleIds := s.getUserRoles(user.UserID)
		deptTitles, deptIds := s.getUserDepts(user.UserID, permission)
		response := models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds)
		applyVisibleLeader(response, user.LeaderUserID, leaderUserNames)
		result = append(result, response)
	}

	return result, nil
}

func (s *UserService) CreateUser(req models.CreateUserRequest, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateManagedDepartments(tx, req.DeptIds, permission); err != nil {
			return err
		}
		if err := validateManagedRoles(tx, req.RoleIds, req.DeptIds, permission); err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&models.SysUser{}).
			Where("username = ? AND del_flag = 0", req.Username).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("用户名已存在")
		}

		userID := utils.GenerateUUID()
		if err := s.validateLeaderUser(tx, userID, req.LeaderUserId, permission); err != nil {
			return err
		}

		now := time.Now()
		user := models.SysUser{
			UserID:       userID,
			Username:     &req.Username,
			RealName:     &req.RealName,
			Phone:        req.Phone,
			Password:     &req.Password,
			Desc:         req.Desc,
			LeaderUserID: req.LeaderUserId,
			Status:       1,
			CreateDate:   &now,
			UpdateDate:   &now,
			DelFlag:      0,
			IsSys:        0,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := s.saveUserRoles(tx, user.UserID, req.RoleIds); err != nil {
			return err
		}
		return s.saveUserDepts(tx, user.UserID, req.DeptIds)
	})
}

func (s *UserService) GetUserDetail(userId string, permission datapermission.Permission) (*models.ProfileResponse, error) {
	var user models.SysUser
	query := database.DB.Model(&models.SysUser{}).Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
	err := permission.Apply(query, "sys_user.user_id").First(&user).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	roleTitles, roleIds := s.getUserRoles(user.UserID)
	deptTitles, deptIds := s.getUserDepts(user.UserID, permission)
	response := models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds)
	leaderUserNames, err := s.getLeaderUserNames([]models.SysUser{user}, permission)
	if err != nil {
		return nil, err
	}
	applyVisibleLeader(response, user.LeaderUserID, leaderUserNames)
	return response, nil
}

func (s *UserService) UpdateUser(userId string, req models.UpdateUserRequest, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateManagedDepartments(tx, req.DeptIds, permission); err != nil {
			return err
		}
		if err := validateManagedRoles(tx, req.RoleIds, req.DeptIds, permission); err != nil {
			return err
		}

		var user models.SysUser
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Model(&models.SysUser{}).
			Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
		if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
			return errors.New("用户不存在")
		}
		if err := ensureFullyManagedUserDepartments(tx, user.UserID, permission); err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&models.SysUser{}).
			Where("username = ? AND del_flag = 0 AND user_id != ?", req.Username, userId).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("用户名已存在")
		}
		if err := s.validateLeaderUser(tx, userId, req.LeaderUserId, permission); err != nil {
			return err
		}

		now := time.Now()
		user.Username = &req.Username
		user.RealName = &req.RealName
		user.Phone = req.Phone
		user.Desc = req.Desc
		user.LeaderUserID = req.LeaderUserId
		user.UpdateDate = &now
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		if err := s.saveUserRoles(tx, user.UserID, req.RoleIds); err != nil {
			return err
		}
		return s.saveUserDepts(tx, user.UserID, req.DeptIds)
	})
}

// validateLeaderUser 校验直属上级存在、启用且不能指向用户自身。
func (s *UserService) validateLeaderUser(tx *gorm.DB, userID string, leaderUserID *string, permission datapermission.Permission) error {
	if leaderUserID == nil || *leaderUserID == "" {
		return nil
	}
	if userID == *leaderUserID {
		return errors.New("直属上级不能选择用户本人")
	}

	var count int64
	query := tx.Model(&models.SysUser{}).
		Where("user_id = ? AND del_flag = 0 AND status = 1 AND is_sys = 0", *leaderUserID)
	if err := permission.Apply(query, "sys_user.user_id").
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("直属上级不存在或已停用")
	}
	return nil
}

func (s *UserService) UpdateUserStatus(userId string, status int, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var user models.SysUser
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Model(&models.SysUser{}).
			Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
		if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
			return errors.New("用户不存在")
		}
		if err := ensureFullyManagedUserDepartments(tx, user.UserID, permission); err != nil {
			return err
		}

		return tx.Model(&user).Updates(map[string]interface{}{
			"status":      status,
			"update_date": time.Now(),
		}).Error
	})
}

func (s *UserService) DeleteUsers(userIds []string, currentUserId string, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		ids := uniqueNonEmptyStrings(userIds)
		users := make([]models.SysUser, 0, len(ids))
		for _, userId := range ids {
			if userId == currentUserId {
				return errors.New("不能删除当前登录用户")
			}
			var user models.SysUser
			query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Model(&models.SysUser{}).
				Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
			if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
				return errors.New("用户不存在或无数据权限")
			}
			if err := ensureFullyManagedUserDepartments(tx, user.UserID, permission); err != nil {
				return err
			}
			users = append(users, user)
		}

		now := time.Now()
		for _, user := range users {
			if err := tx.Model(&models.SysUserRole{}).
				Where("user_id = ? AND del_flag = 0", user.UserID).
				Updates(map[string]interface{}{"del_flag": 1, "update_date": now}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.SysUserDept{}).
				Where("user_id = ? AND del_flag = 0", user.UserID).
				Updates(map[string]interface{}{"del_flag": 1, "update_date": now}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.SysUser{}).
				Where("user_id = ?", user.UserID).
				Updates(map[string]interface{}{"del_flag": 1, "update_date": now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func validateManagedDepartments(tx *gorm.DB, departmentIDs []string, permission datapermission.Permission) error {
	departmentIDs = uniqueNonEmptyStrings(departmentIDs)
	if !permission.All && len(departmentIDs) == 0 {
		return errors.New("非全部数据权限用户必须为用户分配可管理部门")
	}
	if !permission.All && !permission.AllowsDepartments(departmentIDs) {
		return errors.New("不能将用户分配到数据权限范围之外的部门")
	}
	if len(departmentIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.SysDept{}).
		Where("dept_id IN ? AND status = 1 AND del_flag = 0", departmentIDs).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(departmentIDs)) {
		return errors.New("用户部门包含不存在或已停用的部门")
	}
	return nil
}

func validateManagedRoles(tx *gorm.DB, roleIDs, userDepartmentIDs []string, permission datapermission.Permission) error {
	roleIDs = uniqueNonEmptyStrings(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	var roles []models.SysRole
	if err := tx.Where("role_id IN ? AND status = 1 AND del_flag = 0", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	if len(roles) != len(roleIDs) {
		return errors.New("角色不存在或已停用")
	}
	if permission.All {
		return nil
	}
	if permission.UserID == "" {
		return errors.New("无法确认当前操作者的可分配角色")
	}
	var assignableRoleCount int64
	if err := tx.Table("sys_user_role AS user_role").
		Joins("JOIN sys_role AS role ON role.role_id = user_role.role_id AND role.status = 1 AND role.del_flag = 0").
		Where("user_role.user_id = ? AND user_role.role_id IN ? AND user_role.del_flag = 0", permission.UserID, roleIDs).
		Distinct("user_role.role_id").
		Count(&assignableRoleCount).Error; err != nil {
		return err
	}
	if assignableRoleCount != int64(len(roleIDs)) {
		return errors.New("不能分配当前操作者未持有的角色")
	}

	allowedDepartments := make(map[string]struct{}, len(permission.DepartmentIDs))
	for _, departmentID := range permission.DepartmentIDs {
		allowedDepartments[departmentID] = struct{}{}
	}
	for _, role := range roles {
		switch datapermission.Scope(role.DataScope) {
		case datapermission.ScopeAll:
			return errors.New("不能分配数据范围大于当前操作者的角色")
		case datapermission.ScopeCustomDepartment:
			var departmentIDs []string
			if err := tx.Model(&models.SysRoleDept{}).
				Where("role_id = ?", role.RoleID).
				Pluck("dept_id", &departmentIDs).Error; err != nil {
				return err
			}
			if !departmentSubset(departmentIDs, allowedDepartments) {
				return errors.New("不能分配数据范围大于当前操作者的角色")
			}
		case datapermission.ScopeDepartment:
			if !departmentSubset(userDepartmentIDs, allowedDepartments) {
				return errors.New("不能分配数据范围大于当前操作者的角色")
			}
		case datapermission.ScopeDepartmentAndChildren:
			expanded, err := expandDepartmentTrees(tx, userDepartmentIDs)
			if err != nil {
				return err
			}
			if !departmentSubset(expanded, allowedDepartments) {
				return errors.New("不能分配数据范围大于当前操作者的角色")
			}
		case datapermission.ScopeSelf, datapermission.ScopeNone:
			// These scopes cannot expand the operator's department visibility.
		default:
			return errors.New("角色数据范围无效")
		}
	}
	return nil
}

func ensureFullyManagedUserDepartments(tx *gorm.DB, userID string, permission datapermission.Permission) error {
	if permission.All {
		return nil
	}
	var departmentIDs []string
	if err := tx.Table("sys_user_dept AS user_dept").
		Select("user_dept.dept_id").
		Joins("JOIN sys_dept AS dept ON dept.dept_id = user_dept.dept_id AND dept.status = 1 AND dept.del_flag = 0").
		Where("user_dept.user_id = ? AND user_dept.del_flag = 0", userID).
		Pluck("user_dept.dept_id", &departmentIDs).Error; err != nil {
		return err
	}
	if len(departmentIDs) == 0 || !permission.AllowsDepartments(departmentIDs) {
		return errors.New("用户同时属于数据权限范围之外的部门，不能修改")
	}
	return nil
}

func departmentSubset(departmentIDs []string, allowed map[string]struct{}) bool {
	for _, departmentID := range uniqueNonEmptyStrings(departmentIDs) {
		if _, ok := allowed[departmentID]; !ok {
			return false
		}
	}
	return true
}

func expandDepartmentTrees(tx *gorm.DB, rootIDs []string) ([]string, error) {
	var departments []models.SysDept
	if err := tx.Where("status = 1 AND del_flag = 0").Find(&departments).Error; err != nil {
		return nil, err
	}
	childrenByParent := make(map[string][]string)
	for _, department := range departments {
		if department.Pid != nil && *department.Pid != "" {
			childrenByParent[*department.Pid] = append(childrenByParent[*department.Pid], department.DeptID)
		}
	}
	seen := make(map[string]struct{})
	queue := append([]string(nil), uniqueNonEmptyStrings(rootIDs)...)
	for len(queue) > 0 {
		departmentID := queue[0]
		queue = queue[1:]
		if _, ok := seen[departmentID]; ok {
			continue
		}
		seen[departmentID] = struct{}{}
		queue = append(queue, childrenByParent[departmentID]...)
	}
	result := make([]string, 0, len(seen))
	for departmentID := range seen {
		result = append(result, departmentID)
	}
	return result, nil
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

func (s *UserService) saveUserRoles(tx *gorm.DB, userId string, roleIds []string) error {
	if err := tx.Where("user_id = ? AND del_flag = 0", userId).Delete(&models.SysUserRole{}).Error; err != nil {
		return err
	}

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
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) getUserDepts(userId string, permission datapermission.Permission) ([]string, []string) {
	var userDepts []models.SysUserDept
	query := database.DB.Where("user_id = ? AND del_flag = 0", userId)
	if !permission.All && !(permission.IncludeSelf && permission.UserID == userId) {
		if len(permission.DepartmentIDs) == 0 {
			return []string{}, []string{}
		}
		query = query.Where("dept_id IN ?", permission.DepartmentIDs)
	}
	query.Find(&userDepts)

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

func (s *UserService) saveUserDepts(tx *gorm.DB, userId string, deptIds []string) error {
	if err := tx.Where("user_id = ? AND del_flag = 0", userId).Delete(&models.SysUserDept{}).Error; err != nil {
		return err
	}

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
		if err := tx.Create(&userDept).Error; err != nil {
			return err
		}
	}

	return nil
}

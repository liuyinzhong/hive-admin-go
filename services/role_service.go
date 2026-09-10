package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
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

	roleIDs := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}
	userCounts, err := getRoleUserCounts(roleIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*models.RoleSimpleResponse, 0, len(roles))
	for _, role := range roles {
		roleTitle := ""
		if role.RoleTitle != nil {
			roleTitle = *role.RoleTitle
		}
		items = append(items, &models.RoleSimpleResponse{
			RoleId:    role.RoleID,
			RoleTitle: roleTitle,
			Remark:    role.Remark,
			DataScope: role.DataScope,
			Status:    role.Status,
			UserCount: userCounts[role.RoleID],
		})
	}
	pageResult.Items = items
	return pageResult, nil
}

func getRoleUserCounts(roleIDs []string) (map[string]int, error) {
	result := make(map[string]int)
	if len(roleIDs) == 0 {
		return result, nil
	}

	var counts []struct {
		RoleID string `gorm:"column:role_id"`
		Count  int    `gorm:"column:count"`
	}
	err := database.DB.Table("sys_user_role AS user_role").
		Select("user_role.role_id", "COUNT(*) AS count").
		Joins("JOIN sys_user ON sys_user.user_id = user_role.user_id AND sys_user.del_flag = 0").
		Where("user_role.del_flag = 0 AND user_role.role_id IN ?", roleIDs).
		Group("user_role.role_id").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	for _, item := range counts {
		result[item.RoleID] = item.Count
	}
	return result, nil
}

func (s *RoleService) GetAllRoles() ([]*models.RoleSimpleResponse, error) {
	var roles []models.SysRole
	if err := database.DB.Where("del_flag = 0 AND status = 1").Order("create_date desc").Find(&roles).Error; err != nil {
		return nil, err
	}

	responses := make([]*models.RoleSimpleResponse, 0, len(roles))
	for _, role := range roles {
		roleTitle := ""
		if role.RoleTitle != nil {
			roleTitle = *role.RoleTitle
		}
		responses = append(responses, &models.RoleSimpleResponse{
			RoleId:    role.RoleID,
			RoleTitle: roleTitle,
			Remark:    role.Remark,
			DataScope: role.DataScope,
			Status:    role.Status,
		})
	}

	return responses, nil
}

func (s *RoleService) CreateRole(req models.CreateRoleRequest) error {
	dataScope, err := normalizeRoleDataScope(req.DataScope, datapermission.ScopeSelf)
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.SysRole{}).
			Where("role_title = ? AND del_flag = 0", req.RoleTitle).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("角色名称已存在")
		}

		departmentIDs, err := validateRoleDepartmentIDs(tx, dataScope, req.DataScopeDeptIds)
		if err != nil {
			return err
		}

		now := time.Now()
		role := models.SysRole{
			RoleID:     utils.GenerateUUID(),
			RoleTitle:  &req.RoleTitle,
			Status:     req.Status,
			Remark:     req.Remark,
			DataScope:  string(dataScope),
			CreateDate: &now,
			UpdateDate: &now,
			DelFlag:    0,
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		if err := saveRoleMenus(tx, role.RoleID, req.Permissions); err != nil {
			return err
		}
		return saveRoleDepartments(tx, role.RoleID, departmentIDs)
	})
}

func (s *RoleService) GetRoleDetail(roleID string) (*models.RoleDetailResponse, error) {
	var role models.SysRole
	if err := database.DB.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
		return nil, errors.New("角色不存在")
	}

	permissions, err := getRoleMenus(database.DB, roleID)
	if err != nil {
		return nil, err
	}
	departmentIDs, err := getRoleDepartments(database.DB, roleID)
	if err != nil {
		return nil, err
	}

	roleTitle := ""
	if role.RoleTitle != nil {
		roleTitle = *role.RoleTitle
	}

	return &models.RoleDetailResponse{
		RoleId:           role.RoleID,
		RoleTitle:        roleTitle,
		Status:           role.Status,
		CreateDate:       models.TimeToStringPtr(role.CreateDate),
		Remark:           role.Remark,
		Permissions:      permissions,
		DataScope:        role.DataScope,
		DataScopeDeptIds: departmentIDs,
	}, nil
}

func (s *RoleService) UpdateRole(roleID string, req models.UpdateRoleRequest) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var role models.SysRole
		if err := tx.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
			return errors.New("角色不存在")
		}

		dataScope, err := normalizeRoleDataScope(req.DataScope, datapermission.Scope(role.DataScope))
		if err != nil {
			return err
		}
		departmentIDs, err := validateRoleDepartmentIDs(tx, dataScope, req.DataScopeDeptIds)
		if err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&models.SysRole{}).
			Where("role_title = ? AND del_flag = 0 AND role_id != ?", req.RoleTitle, roleID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("角色名称已存在")
		}

		now := time.Now()
		if err := tx.Model(&role).Updates(map[string]interface{}{
			"role_title":  req.RoleTitle,
			"status":      req.Status,
			"remark":      req.Remark,
			"data_scope":  string(dataScope),
			"update_date": now,
		}).Error; err != nil {
			return err
		}
		if err := saveRoleMenus(tx, roleID, req.Permissions); err != nil {
			return err
		}
		return saveRoleDepartments(tx, roleID, departmentIDs)
	})
}

func (s *RoleService) UpdateRoleStatus(roleID string, status int) error {
	var role models.SysRole
	if err := database.DB.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
		return errors.New("角色不存在")
	}

	return database.DB.Model(&role).Updates(map[string]interface{}{
		"status":      status,
		"update_date": time.Now(),
	}).Error
}

func (s *RoleService) DeleteRoles(roleIDs []string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, roleID := range uniqueNonEmptyStrings(roleIDs) {
			var role models.SysRole
			if err := tx.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleMenu{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleDept{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id = ?", roleID).Delete(&models.SysUserRole{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&role).Updates(map[string]interface{}{
				"del_flag":    1,
				"update_date": time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func normalizeRoleDataScope(value string, fallback datapermission.Scope) (datapermission.Scope, error) {
	if value == "" {
		value = string(fallback)
	}
	scope := datapermission.Scope(value)
	switch scope {
	case datapermission.ScopeAll,
		datapermission.ScopeCustomDepartment,
		datapermission.ScopeDepartment,
		datapermission.ScopeDepartmentAndChildren,
		datapermission.ScopeSelf,
		datapermission.ScopeNone:
		return scope, nil
	default:
		return "", fmt.Errorf("数据范围不合法")
	}
}

func validateRoleDepartmentIDs(tx *gorm.DB, scope datapermission.Scope, departmentIDs []string) ([]string, error) {
	if scope != datapermission.ScopeCustomDepartment {
		return []string{}, nil
	}

	ids := uniqueNonEmptyStrings(departmentIDs)
	if len(ids) == 0 {
		return nil, fmt.Errorf("自定义部门数据范围至少选择一个部门")
	}
	for _, departmentID := range ids {
		if _, err := uuid.Parse(departmentID); err != nil {
			return nil, fmt.Errorf("部门ID格式错误")
		}
	}

	var count int64
	if err := tx.Model(&models.SysDept{}).
		Where("dept_id IN ? AND status = 1 AND del_flag = 0", ids).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, fmt.Errorf("自定义数据范围包含不存在或已停用的部门")
	}
	return ids, nil
}

func saveRoleMenus(tx *gorm.DB, roleID string, menuIDs []string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleMenu{}).Error; err != nil {
		return err
	}

	now := time.Now()
	for _, menuID := range uniqueNonEmptyStrings(menuIDs) {
		roleMenu := models.SysRoleMenu{
			ID:         utils.GenerateUUID(),
			RoleID:     roleID,
			MenuID:     menuID,
			CreateDate: &now,
			UpdateDate: &now,
			DelFlag:    0,
		}
		if err := tx.Create(&roleMenu).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveRoleDepartments(tx *gorm.DB, roleID string, departmentIDs []string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleDept{}).Error; err != nil {
		return err
	}

	now := time.Now()
	for _, departmentID := range departmentIDs {
		roleDepartment := models.SysRoleDept{
			ID:         utils.GenerateUUID(),
			RoleID:     roleID,
			DeptID:     departmentID,
			CreateDate: &now,
		}
		if err := tx.Create(&roleDepartment).Error; err != nil {
			return err
		}
	}
	return nil
}

func getRoleMenus(tx *gorm.DB, roleID string) ([]string, error) {
	var menuIDs []string
	err := tx.Model(&models.SysRoleMenu{}).
		Where("role_id = ? AND del_flag = 0", roleID).
		Order("create_date ASC").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func getRoleDepartments(tx *gorm.DB, roleID string) ([]string, error) {
	var departmentIDs []string
	err := tx.Model(&models.SysRoleDept{}).
		Where("role_id = ?", roleID).
		Order("create_date ASC").
		Pluck("dept_id", &departmentIDs).Error
	return departmentIDs, err
}

type roleUserRow struct {
	UserID   string     `gorm:"column:user_id"`
	Username *string    `gorm:"column:username"`
	RealName *string    `gorm:"column:real_name"`
	Status   int        `gorm:"column:status"`
	JoinDate *time.Time `gorm:"column:join_date"`
}

func (s *RoleService) GetRoleUsers(roleID string, req models.RoleUserListRequest) (*utils.PageResult, error) {
	var role models.SysRole
	if err := database.DB.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
		return nil, errors.New("角色不存在")
	}

	query := database.DB.Model(&models.SysUserRole{}).
		Select("sys_user_role.user_id", "sys_user_role.create_date AS join_date", "sys_user.username", "sys_user.real_name", "sys_user.status").
		Joins("JOIN sys_user ON sys_user.user_id = sys_user_role.user_id AND sys_user.del_flag = 0").
		Where("sys_user_role.role_id = ? AND sys_user_role.del_flag = 0", roleID)

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("sys_user.username LIKE ? OR sys_user.real_name LIKE ?", like, like)
	}
	if req.Status != nil {
		query = query.Where("sys_user.status = ?", *req.Status)
	}
	query = query.Order("sys_user_role.create_date DESC")

	var rows []roleUserRow
	pageResult, err := utils.Paginate(query, req.Page, req.PageSize, &rows)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	deptTitles, err := getRoleUserDeptTitles(userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*models.RoleUserItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &models.RoleUserItem{
			UserId:     row.UserID,
			Username:   utils.StringValue(row.Username),
			RealName:   utils.StringValue(row.RealName),
			Status:     row.Status,
			DeptTitles: deptTitles[row.UserID],
			JoinDate:   models.TimeToStringPtr(row.JoinDate),
		})
	}
	pageResult.Items = items
	return pageResult, nil
}

func getRoleUserDeptTitles(userIDs []string) (map[string][]string, error) {
	result := make(map[string][]string)
	if len(userIDs) == 0 {
		return result, nil
	}

	var userDepts []models.SysUserDept
	if err := database.DB.Where("user_id IN ? AND del_flag = 0", userIDs).Find(&userDepts).Error; err != nil {
		return nil, err
	}
	if len(userDepts) == 0 {
		return result, nil
	}

	deptIDs := make([]string, 0, len(userDepts))
	for _, userDept := range userDepts {
		deptIDs = append(deptIDs, userDept.DeptID)
	}
	var depts []models.SysDept
	if err := database.DB.Where("dept_id IN ? AND del_flag = 0", deptIDs).Find(&depts).Error; err != nil {
		return nil, err
	}
	deptTitleByID := make(map[string]string, len(depts))
	for _, dept := range depts {
		deptTitleByID[dept.DeptID] = utils.StringValue(dept.DeptTitle)
	}
	for _, userDept := range userDepts {
		if title, ok := deptTitleByID[userDept.DeptID]; ok {
			result[userDept.UserID] = append(result[userDept.UserID], title)
		}
	}
	return result, nil
}

func (s *RoleService) AddRoleUsers(roleID string, userIDs []string) (int, error) {
	ids := uniqueNonEmptyStrings(userIDs)
	if len(ids) == 0 {
		return 0, errors.New("用户ID列表不能为空")
	}
	for _, userID := range ids {
		if _, err := uuid.Parse(userID); err != nil {
			return 0, fmt.Errorf("用户ID格式错误")
		}
	}

	added := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var role models.SysRole
		if err := tx.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
			return errors.New("角色不存在")
		}

		var count int64
		if err := tx.Model(&models.SysUser{}).
			Where("user_id IN ? AND del_flag = 0", ids).
			Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return errors.New("包含不存在或已删除的用户")
		}

		var existing []string
		if err := tx.Model(&models.SysUserRole{}).
			Where("role_id = ? AND user_id IN ? AND del_flag = 0", roleID, ids).
			Pluck("user_id", &existing).Error; err != nil {
			return err
		}
		existingSet := make(map[string]struct{}, len(existing))
		for _, userID := range existing {
			existingSet[userID] = struct{}{}
		}

		now := time.Now()
		for _, userID := range ids {
			if _, ok := existingSet[userID]; ok {
				continue
			}
			userRole := models.SysUserRole{
				ID:         utils.GenerateUUID(),
				UserID:     userID,
				RoleID:     roleID,
				CreateDate: &now,
				UpdateDate: &now,
				DelFlag:    0,
			}
			if err := tx.Create(&userRole).Error; err != nil {
				return err
			}
			added++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return added, nil
}

func (s *RoleService) RemoveRoleUsers(roleID string, userIDs []string) (int, error) {
	ids := uniqueNonEmptyStrings(userIDs)
	if len(ids) == 0 {
		return 0, errors.New("用户ID列表不能为空")
	}

	var role models.SysRole
	if err := database.DB.Where("role_id = ? AND del_flag = 0", roleID).First(&role).Error; err != nil {
		return 0, errors.New("角色不存在")
	}

	result := database.DB.Where("role_id = ? AND user_id IN ?", roleID, ids).Delete(&models.SysUserRole{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

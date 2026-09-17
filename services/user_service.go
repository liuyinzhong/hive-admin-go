package services

import (
	"errors"
	"log"
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

// userManagementMenuName 用户管理页面菜单的 name，用于重置密码站内消息归属
const userManagementMenuName = "systemUser"

// ResetUserPassword 管理员在当前角色数据范围内为目标用户直接设置新密码，不验证旧密码。
// 成功后将 pwd_version 递增，使目标用户全部旧世代会话凭证立即失效，
// 并推送强制退出事件和站内消息；推送失败仅记日志，不影响重置结果。
func (s *UserService) ResetUserPassword(userId string, req models.ResetPasswordRequest, permission datapermission.Permission) error {
	var user models.SysUser
	query := database.DB.Model(&models.SysUser{}).Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
	if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}

	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := database.DB.Model(&models.SysUser{}).
		Where("user_id = ?", user.UserID).
		Updates(map[string]interface{}{
			"password":    hash,
			"pwd_version": gorm.Expr("pwd_version + 1"),
			"update_date": time.Now(),
		}).Error; err != nil {
		return err
	}

	messageService := NewMenuMessageService()
	if err := messageService.CreateMenuMessageForMenuName(user.UserID, userManagementMenuName,
		"密码已重置", "你的登录密码已被管理员重置，请使用新密码重新登录"); err != nil {
		log.Printf("[user] 重置密码站内消息推送失败: userId=%s, err=%v", user.UserID, err)
	}
	messageService.PublishForceLogout(user.UserID)
	return nil
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
	if err := fillPermissionSummary(resultItems); err != nil {
		return nil, err
	}

	pageResult.Items = resultItems
	return pageResult, nil
}

// fillPermissionSummary 为用户分页结果批量填充权限摘要计数（配置事实口径，与权限明细抽屉各分组计数一致）：
// 角色授权节点数为各未删除角色授权数之和（多角色重复授权不去重），个人计数为额外授权与禁止的节点数。
func fillPermissionSummary(items []*models.ProfileResponse) error {
	userIds := make([]string, 0, len(items))
	for _, item := range items {
		userIds = append(userIds, item.UserId)
	}
	if len(userIds) == 0 {
		return nil
	}

	var roleCounts []struct {
		UserID string `gorm:"column:user_id"`
		Count  int    `gorm:"column:count"`
	}
	if err := database.DB.Table("sys_user_role AS user_role").
		Select("user_role.user_id, COUNT(role_menu.id) AS count").
		Joins("JOIN sys_role AS role ON role.role_id = user_role.role_id AND role.del_flag = 0").
		Joins("JOIN sys_role_menu AS role_menu ON role_menu.role_id = user_role.role_id AND role_menu.del_flag = 0").
		Where("user_role.del_flag = 0 AND user_role.user_id IN ?", userIds).
		Group("user_role.user_id").
		Scan(&roleCounts).Error; err != nil {
		return err
	}

	var personalCounts []struct {
		UserID    string `gorm:"column:user_id"`
		GrantType string `gorm:"column:grant_type"`
		Count     int    `gorm:"column:count"`
	}
	if err := database.DB.Model(&models.SysUserMenu{}).
		Select("user_id, grant_type, COUNT(*) AS count").
		Where("user_id IN ? AND del_flag = 0", userIds).
		Group("user_id, grant_type").
		Scan(&personalCounts).Error; err != nil {
		return err
	}

	rolePermissionCounts := make(map[string]int, len(roleCounts))
	for _, item := range roleCounts {
		rolePermissionCounts[item.UserID] = item.Count
	}
	grantCounts := make(map[string]int, len(personalCounts))
	denyCounts := make(map[string]int, len(personalCounts))
	for _, item := range personalCounts {
		switch item.GrantType {
		case grantTypeGrant:
			grantCounts[item.UserID] = item.Count
		case grantTypeDeny:
			denyCounts[item.UserID] = item.Count
		}
	}
	for _, item := range items {
		item.RolePermissionCount = rolePermissionCounts[item.UserId]
		item.GrantCount = grantCounts[item.UserId]
		item.DenyCount = denyCounts[item.UserId]
	}
	return nil
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
	if err := utils.ValidatePassword(req.Password); err != nil {
		return err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

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
			Password:     &hash,
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
		if err := s.saveUserDepts(tx, user.UserID, req.DeptIds); err != nil {
			return err
		}
		return applyPersonalPermissionUpdate(tx, user.UserID, req.GrantMenuIds, req.DenyMenuIds, permission)
	})
}

// GetUserDetail 返回用户管理详情：基础信息之上聚合个人权限两个菜单ID集合，供用户编辑抽屉一次加载；
// 越界部门和直属领导关联不返回。
func (s *UserService) GetUserDetail(userId string, permission datapermission.Permission) (*models.UserDetailResponse, error) {
	var user models.SysUser
	query := database.DB.Model(&models.SysUser{}).Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
	err := permission.Apply(query, "sys_user.user_id").First(&user).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	roleTitles, roleIds := s.getUserRoles(user.UserID)
	deptTitles, deptIds := s.getUserDepts(user.UserID, permission)
	profile := models.SysUserToProfileResponse(user, roleTitles, roleIds, deptTitles, deptIds)
	leaderUserNames, err := s.getLeaderUserNames([]models.SysUser{user}, permission)
	if err != nil {
		return nil, err
	}
	applyVisibleLeader(profile, user.LeaderUserID, leaderUserNames)

	granted, denied, err := loadUserMenuIDs(database.DB, user.UserID)
	if err != nil {
		return nil, err
	}
	return &models.UserDetailResponse{
		ProfileResponse: *profile,
		GrantMenuIds:    granted,
		DenyMenuIds:     denied,
	}, nil
}

// GetUserPermissions 按用户维度一次聚合返回其全部关联角色及各自菜单授权明细。
// 数据范围边界与用户详情一致；明细按配置事实返回，停用角色与停用菜单照常返回并携带状态；
// 目录节点不单独成行、仅作路径前缀，多角色重复授权的权限在各角色分组内分别出现，不去重。
func (s *UserService) GetUserPermissions(userId string, permission datapermission.Permission) (*models.UserPermissionResponse, error) {
	var user models.SysUser
	query := database.DB.Model(&models.SysUser{}).Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
	if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	var userRoles []models.SysUserRole
	if err := database.DB.Where("user_id = ? AND del_flag = 0", userId).
		Order("create_date ASC").Find(&userRoles).Error; err != nil {
		return nil, err
	}

	roleIds := make([]string, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIds = append(roleIds, ur.RoleID)
	}

	roleById := make(map[string]models.SysRole, len(roleIds))
	roleMenuIds := make(map[string]map[string]struct{}, len(roleIds))
	if len(roleIds) > 0 {
		var roles []models.SysRole
		if err := database.DB.Where("role_id IN ? AND del_flag = 0", roleIds).Find(&roles).Error; err != nil {
			return nil, err
		}
		for _, role := range roles {
			roleById[role.RoleID] = role
		}

		var roleMenus []models.SysRoleMenu
		if err := database.DB.Where("role_id IN ? AND del_flag = 0", roleIds).
			Order("create_date ASC").Find(&roleMenus).Error; err != nil {
			return nil, err
		}
		for _, rm := range roleMenus {
			if roleMenuIds[rm.RoleID] == nil {
				roleMenuIds[rm.RoleID] = make(map[string]struct{})
			}
			roleMenuIds[rm.RoleID][rm.MenuID] = struct{}{}
		}
	}

	orderedItems, err := buildMenuNodeTree()
	if err != nil {
		return nil, err
	}

	granted, denied, err := loadUserMenuIDs(database.DB, user.UserID)
	if err != nil {
		return nil, err
	}

	result := &models.UserPermissionResponse{
		UserId:   user.UserID,
		RealName: utils.StringValue(user.RealName),
		Roles:    make([]models.UserPermissionRole, 0, len(userRoles)),
	}
	for _, ur := range userRoles {
		role, exists := roleById[ur.RoleID]
		if !exists {
			continue
		}
		menuIds := roleMenuIds[ur.RoleID]
		result.Roles = append(result.Roles, models.UserPermissionRole{
			RoleId:          role.RoleID,
			RoleTitle:       utils.StringValue(role.RoleTitle),
			Status:          role.Status,
			DataScope:       role.DataScope,
			Remark:          role.Remark,
			PermissionCount: len(menuIds),
			Permissions:     pruneGrantedItems(orderedItems, menuIds),
		})
	}

	if len(granted) > 0 || len(denied) > 0 {
		grantSet := make(map[string]struct{}, len(granted))
		for _, id := range granted {
			grantSet[id] = struct{}{}
		}
		denySet := make(map[string]struct{}, len(denied))
		for _, id := range denied {
			denySet[id] = struct{}{}
		}
		result.Personal = &models.UserPermissionPersonal{
			Grant:      pruneGrantedItems(orderedItems, grantSet),
			GrantCount: len(granted),
			Deny:       pruneGrantedItems(orderedItems, denySet),
			DenyCount:  len(denied),
		}
	}

	return result, nil
}

// menuNode 全量菜单树节点，供按角色裁剪授权子树使用
type menuNode struct {
	children []menuNode
	menu     models.SysMenu
}

// buildMenuNodeTree 按菜单树先序顺序构建全量菜单树；排序与菜单管理一致（order asc, create_date desc）。
func buildMenuNodeTree() ([]menuNode, error) {
	var menus []models.SysMenu
	if err := database.DB.Where("del_flag = 0 AND type != ?", externalPageType).
		Order("`order` asc, create_date desc").Find(&menus).Error; err != nil {
		return nil, err
	}

	childrenByParent := make(map[string][]models.SysMenu)
	roots := make([]models.SysMenu, 0)
	for _, menu := range menus {
		if menu.Pid == nil || *menu.Pid == "" {
			roots = append(roots, menu)
		} else {
			childrenByParent[*menu.Pid] = append(childrenByParent[*menu.Pid], menu)
		}
	}

	var attach func(menu models.SysMenu) menuNode
	attach = func(menu models.SysMenu) menuNode {
		node := menuNode{menu: menu}
		for _, child := range childrenByParent[menu.ID] {
			node.children = append(node.children, attach(child))
		}
		return node
	}

	nodes := make([]menuNode, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, attach(root))
	}
	return nodes, nil
}

// pruneGrantedItems 裁剪出授权子树：节点自身被授权，或其下存在被授权的后代时保留；
// 目录节点自身未被授权时仅作为层级结构节点出现。同一权限被多个角色授予时在各角色树内分别出现，不去重。
func pruneGrantedItems(nodes []menuNode, granted map[string]struct{}) []models.UserPermissionItem {
	items := make([]models.UserPermissionItem, 0)
	for _, node := range nodes {
		children := pruneGrantedItems(node.children, granted)
		if _, selfGranted := granted[node.menu.ID]; !selfGranted && len(children) == 0 {
			continue
		}
		items = append(items, models.UserPermissionItem{
			MenuId:   node.menu.ID,
			Title:    node.menu.Title,
			Type:     node.menu.Type,
			Path:     node.menu.Path,
			AuthCode: node.menu.AuthCode,
			Status:   node.menu.Status,
			Children: children,
		})
	}
	return items
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
		if err := s.saveUserDepts(tx, user.UserID, req.DeptIds); err != nil {
			return err
		}
		return applyPersonalPermissionUpdate(tx, user.UserID, req.GrantMenuIds, req.DenyMenuIds, permission)
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
			if err := tx.Where("user_id = ?", user.UserID).Delete(&models.SysUserRole{}).Error; err != nil {
				return err
			}
			// 个人权限随用户删除一并清理
			if err := tx.Where("user_id = ?", user.UserID).Delete(&models.SysUserMenu{}).Error; err != nil {
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

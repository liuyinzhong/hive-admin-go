package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// 个人权限类型：grant=额外授权（角色之外的附加菜单），deny=禁止（显式收回，合并时绝对优先）
const (
	grantTypeGrant = "grant"
	grantTypeDeny  = "deny"
)

type UserMenuService struct {
	menuService *MenuService
}

func NewUserMenuService() *UserMenuService {
	return &UserMenuService{menuService: NewMenuService()}
}

// effectiveMenuIDSet 计算普通用户的生效菜单ID集合：
// 全部有效角色的菜单授权 ∪ 个人额外授权 − 个人禁止（禁止绝对优先）。
// 它是动态菜单和权限码两个生效口径的唯一事实来源；系统内置用户不适用（直接取得全部启用菜单）。
func effectiveMenuIDSet(userID string) (map[string]struct{}, error) {
	var roleMenuIDs []string
	if err := database.DB.Table("sys_user_role AS user_role").
		Joins("JOIN sys_role AS role ON role.role_id = user_role.role_id AND role.status = 1 AND role.del_flag = 0").
		Joins("JOIN sys_role_menu AS role_menu ON role_menu.role_id = role.role_id AND role_menu.del_flag = 0").
		Where("user_role.user_id = ? AND user_role.del_flag = 0", userID).
		Distinct("role_menu.menu_id").
		Pluck("role_menu.menu_id", &roleMenuIDs).Error; err != nil {
		return nil, err
	}

	menuIDSet := make(map[string]struct{}, len(roleMenuIDs))
	for _, id := range roleMenuIDs {
		menuIDSet[id] = struct{}{}
	}

	var userMenus []models.SysUserMenu
	if err := database.DB.Where("user_id = ? AND del_flag = 0", userID).Find(&userMenus).Error; err != nil {
		return nil, err
	}
	for _, um := range userMenus {
		switch um.GrantType {
		case grantTypeGrant:
			menuIDSet[um.MenuID] = struct{}{}
		case grantTypeDeny:
			delete(menuIDSet, um.MenuID)
		}
	}

	return menuIDSet, nil
}

func menuIDSetToSlice(menuIDSet map[string]struct{}) []string {
	menuIDs := make([]string, 0, len(menuIDSet))
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}
	return menuIDs
}

// GetPersonalPermissions 查询用户个人权限的两个菜单ID集合，并同时返回供勾选的启用菜单树；
// 数据范围边界与用户详情一致。菜单树随本接口返回，维护个人权限不要求菜单管理列表权限。
func (s *UserMenuService) GetPersonalPermissions(userId string, permission datapermission.Permission) (*models.UserPersonalPermissionResponse, error) {
	user, err := loadManagedUser(database.DB, userId, permission)
	if err != nil {
		return nil, err
	}

	granted, denied, err := loadUserMenuIDs(database.DB, user.UserID)
	if err != nil {
		return nil, err
	}

	status := 1
	menuTree, err := s.menuService.GetMenuTree(models.MenuListRequest{Status: &status, HasButton: 1})
	if err != nil {
		return nil, err
	}

	return &models.UserPersonalPermissionResponse{
		UserId:       user.UserID,
		RealName:     utils.StringValue(user.RealName),
		GrantMenuIds: granted,
		DenyMenuIds:  denied,
		MenuTree:     menuTree,
	}, nil
}

// SavePersonalPermissions 按提交集合完整替换用户个人权限；授予和禁止都要求菜单有效且操作者持有
// （全部数据范围操作者豁免，镜像角色分配的防提升约束）。系统内置用户不能配置个人权限。
func (s *UserMenuService) SavePersonalPermissions(userId string, req models.SaveUserPersonalPermissionRequest, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := loadManagedUser(tx, userId, permission)
		if err != nil {
			return err
		}

		grants := uniqueNonEmptyStrings(req.GrantMenuIds)
		denies := uniqueNonEmptyStrings(req.DenyMenuIds)

		grantSet := make(map[string]struct{}, len(grants))
		for _, id := range grants {
			grantSet[id] = struct{}{}
		}
		for _, id := range denies {
			if _, exists := grantSet[id]; exists {
				return errors.New("同一菜单不能同时额外授权和禁止")
			}
		}

		menuIDs := append(append([]string{}, grants...), denies...)
		if err := validateExistingMenus(tx, menuIDs); err != nil {
			return err
		}
		if err := validateOperableMenus(tx, menuIDs, permission); err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", user.UserID).Delete(&models.SysUserMenu{}).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, menuID := range grants {
			if err := createUserMenu(tx, user.UserID, menuID, grantTypeGrant, now); err != nil {
				return err
			}
		}
		for _, menuID := range denies {
			if err := createUserMenu(tx, user.UserID, menuID, grantTypeDeny, now); err != nil {
				return err
			}
		}
		return nil
	})
}

// loadManagedUser 按用户详情同边界加载目标用户：角色数据范围过滤且排除系统内置用户。
func loadManagedUser(tx *gorm.DB, userId string, permission datapermission.Permission) (*models.SysUser, error) {
	var user models.SysUser
	query := tx.Model(&models.SysUser{}).Where("user_id = ? AND del_flag = 0 AND is_sys = 0", userId)
	if err := permission.Apply(query, "sys_user.user_id").First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func loadUserMenuIDs(tx *gorm.DB, userId string) (granted, denied []string, err error) {
	var userMenus []models.SysUserMenu
	if err := tx.Where("user_id = ? AND del_flag = 0", userId).Find(&userMenus).Error; err != nil {
		return nil, nil, err
	}
	granted = make([]string, 0)
	denied = make([]string, 0)
	for _, um := range userMenus {
		switch um.GrantType {
		case grantTypeGrant:
			granted = append(granted, um.MenuID)
		case grantTypeDeny:
			denied = append(denied, um.MenuID)
		}
	}
	return granted, denied, nil
}

func createUserMenu(tx *gorm.DB, userId, menuID, grantType string, now time.Time) error {
	entry := models.SysUserMenu{
		ID:         utils.GenerateUUID(),
		UserID:     userId,
		MenuID:     menuID,
		GrantType:  grantType,
		CreateDate: &now,
		UpdateDate: &now,
	}
	return tx.Create(&entry).Error
}

// validateExistingMenus 校验菜单ID均存在且未删除（含按钮节点，排除外部页面）。
func validateExistingMenus(tx *gorm.DB, menuIDs []string) error {
	if len(menuIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.SysMenu{}).
		Where("id IN ? AND del_flag = 0 AND type != ?", menuIDs, externalPageType).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(uniqueNonEmptyStrings(menuIDs))) {
		return errors.New("菜单不存在或已删除")
	}
	return nil
}

// validateOperableMenus 防提升校验：非全部数据范围的操作者只能额外授权或禁止
// 自己生效权限内且处于启用状态的菜单（停用菜单不进入任何人的生效口径）。
func validateOperableMenus(tx *gorm.DB, menuIDs []string, permission datapermission.Permission) error {
	if permission.All || len(menuIDs) == 0 {
		return nil
	}
	if permission.UserID == "" {
		return errors.New("无法确认当前操作者的生效权限")
	}
	menuIDSet, err := effectiveMenuIDSet(permission.UserID)
	if err != nil {
		return err
	}
	// 生效集合中仅启用菜单算持有
	var enabledIDs []string
	if err := tx.Model(&models.SysMenu{}).
		Where("id IN ? AND status = 1 AND del_flag = 0", menuIDSetToSlice(menuIDSet)).
		Pluck("id", &enabledIDs).Error; err != nil {
		return err
	}
	enabledSet := make(map[string]struct{}, len(enabledIDs))
	for _, id := range enabledIDs {
		enabledSet[id] = struct{}{}
	}
	for _, id := range uniqueNonEmptyStrings(menuIDs) {
		if _, held := enabledSet[id]; !held {
			return errors.New("不能调整当前操作者未持有的菜单")
		}
	}
	return nil
}

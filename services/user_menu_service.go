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

// GetPersonalPermissionMenuTree 返回个人权限可勾选的启用菜单树（含按钮），
// 是用户新建与编辑抽屉的唯一树来源，维护个人权限不要求菜单管理列表权限。
func (s *UserMenuService) GetPersonalPermissionMenuTree() (*models.PersonalPermissionMenuTreeResponse, error) {
	status := 1
	menuTree, err := s.menuService.GetMenuTree(models.MenuListRequest{Status: &status, HasButton: 1})
	if err != nil {
		return nil, err
	}
	return &models.PersonalPermissionMenuTreeResponse{MenuTree: menuTree}, nil
}

// replaceUserPersonalPermissions 在既有事务内按提交集合完整替换用户个人权限；授予和禁止都要求
// 菜单有效且操作者持有（全部数据范围操作者豁免，镜像角色分配的防提升约束）。
// 调用方须先确认目标用户存在且非系统内置用户（UpdateUser 的 is_sys=0 锁定查询、CreateUser 固定 is_sys=0），
// 并完成个人权限维护权限码校验。
func replaceUserPersonalPermissions(tx *gorm.DB, userId string, grants, denies []string, permission datapermission.Permission) error {
	grants = uniqueNonEmptyStrings(grants)
	denies = uniqueNonEmptyStrings(denies)

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

	if err := tx.Where("user_id = ?", userId).Delete(&models.SysUserMenu{}).Error; err != nil {
		return err
	}

	now := time.Now()
	for _, menuID := range grants {
		if err := createUserMenu(tx, userId, menuID, grantTypeGrant, now); err != nil {
			return err
		}
	}
	for _, menuID := range denies {
		if err := createUserMenu(tx, userId, menuID, grantTypeDeny, now); err != nil {
			return err
		}
	}
	return nil
}

// applyPersonalPermissionUpdate 在创建/更新用户事务内处理个人权限字段：
// 任一集合非 nil 即要求操作者持有 system:user:personalPermission 权限码，
// nil 的集合保持原值不变，非 nil 的集合按完整替换语义写入。
func applyPersonalPermissionUpdate(tx *gorm.DB, userId string, grantMenuIds, denyMenuIds *[]string, permission datapermission.Permission) error {
	if grantMenuIds == nil && denyMenuIds == nil {
		return nil
	}
	if permission.UserID == "" || !NewPermissionService().HasCode(permission.UserID, "system:user:personalPermission") {
		return errors.New("无个人权限维护权限")
	}

	granted, denied, err := loadUserMenuIDs(tx, userId)
	if err != nil {
		return err
	}
	if grantMenuIds != nil {
		granted = *grantMenuIds
	}
	if denyMenuIds != nil {
		denied = *denyMenuIds
	}
	return replaceUserPersonalPermissions(tx, userId, granted, denied, permission)
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

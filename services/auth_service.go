package services

import (
	"errors"
	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"net/mail"
	"strings"
	"time"
)

type AuthService struct {
	permissionService *PermissionService
}

func NewAuthService() *AuthService {
	return &AuthService{permissionService: NewPermissionService()}
}

func (s *AuthService) Login(username, password string) (string, error) {
	var user models.SysUser
	result := database.DB.Where("username = ? AND del_flag = 0", username).First(&user)
	if result.Error != nil {
		return "", errors.New("账号密码有误")
	}

	if user.Status == 0 {
		return "", errors.New("该账号已被禁用")
	}

	if user.Password == nil || *user.Password != password {
		return "", errors.New("账号密码有误")
	}

	token, err := utils.GenerateToken(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GetProfile(userID string) (*models.ProfileResponse, error) {
	var user models.SysUser
	if err := database.DB.Where("user_id = ? AND del_flag = 0", userID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	var roleTitles []string
	var roleIds []string
	var deptTitles []string
	var deptIds []string

	var userRoles []models.SysUserRole
	database.DB.Where("user_id = ? AND del_flag = 0", userID).Find(&userRoles)

	for _, ur := range userRoles {
		var role models.SysRole
		if err := database.DB.Where("role_id = ? AND del_flag = 0 AND status = 1", ur.RoleID).First(&role).Error; err == nil {
			if role.RoleTitle != nil {
				roleTitles = append(roleTitles, *role.RoleTitle)
			}
			roleIds = append(roleIds, ur.RoleID)
		}
	}

	var userDepts []models.SysUserDept
	database.DB.Where("user_id = ? AND del_flag = 0", userID).Find(&userDepts)

	for _, ud := range userDepts {
		var dept models.SysDept
		if err := database.DB.Where("dept_id = ? AND del_flag = 0", ud.DeptID).First(&dept).Error; err == nil {
			if dept.DeptTitle != nil {
				deptTitles = append(deptTitles, *dept.DeptTitle)
			}
			deptIds = append(deptIds, ud.DeptID)
		}
	}

	profile := &models.ProfileResponse{
		UserId:     user.UserID,
		Avatar:     user.Avatar,
		Username:   "",
		RealName:   "",
		Phone:      user.Phone,
		RoleTitles: roleTitles,
		RoleIds:    roleIds,
		Email:      user.Email,
		Signature:  user.Signature,
		HomePath:   user.HomePath,
		DeptTitles: deptTitles,
		DeptIds:    deptIds,
		Status:     user.Status,
	}

	if user.Username != nil {
		profile.Username = *user.Username
	}
	if user.RealName != nil {
		profile.RealName = *user.RealName
	}

	return profile, nil
}

// UpdateProfile 当前用户更新自己的头像、邮箱和签名，返回更新后的用户资料。
// 字段为 nil 表示不修改；空字符串表示清空（写入 NULL）。只允许操作当前 Token 对应的用户。
func (s *AuthService) UpdateProfile(userID string, req models.UpdateProfileRequest) (*models.ProfileResponse, error) {
	updates := map[string]interface{}{}
	if req.Avatar != nil {
		avatar := strings.TrimSpace(*req.Avatar)
		if avatar == "" {
			updates["avatar"] = nil
		} else {
			updates["avatar"] = avatar
		}
	}
	if req.Signature != nil {
		signature := strings.TrimSpace(*req.Signature)
		if signature == "" {
			updates["signature"] = nil
		} else {
			updates["signature"] = signature
		}
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			updates["email"] = nil
		} else {
			if _, err := mail.ParseAddress(email); err != nil {
				return nil, errors.New("邮箱格式错误")
			}
			updates["email"] = email
		}
	}

	if len(updates) > 0 {
		updates["update_date"] = time.Now()
		if err := database.DB.Model(&models.SysUser{}).
			Where("user_id = ? AND del_flag = 0", userID).
			Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	// 更新后重新读取完整资料；用户不存在时由 GetProfile 统一返回错误
	return s.GetProfile(userID)
}

func (s *AuthService) GetMenus(userID string) ([]*models.MenuTreeResponse, error) {
	var user models.SysUser
	if err := database.DB.Where("user_id = ? AND del_flag = 0", userID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	var menus []models.SysMenu

	if user.IsSys == 1 {
		database.DB.Where("del_flag = 0 AND status = 1 AND type IN ?", []string{"menu", "catalog", "embedded", "link"}).Find(&menus)
	} else {
		var userRoles []models.SysUserRole
		database.DB.Where("user_id = ? AND del_flag = 0", userID).Find(&userRoles)

		if len(userRoles) == 0 {
			return []*models.MenuTreeResponse{}, nil
		}

		var roleIDs []string
		for _, ur := range userRoles {
			var role models.SysRole
			if err := database.DB.Where("role_id = ? AND del_flag = 0 AND status = 1", ur.RoleID).First(&role).Error; err == nil {
				roleIDs = append(roleIDs, ur.RoleID)
			}
		}

		if len(roleIDs) == 0 {
			return []*models.MenuTreeResponse{}, nil
		}

		var roleMenus []models.SysRoleMenu
		database.DB.Where("role_id IN ? AND del_flag = 0", roleIDs).Find(&roleMenus)

		if len(roleMenus) == 0 {
			return []*models.MenuTreeResponse{}, nil
		}

		var menuIDs []string
		menuIDSet := make(map[string]bool)
		for _, rm := range roleMenus {
			menuIDSet[rm.MenuID] = true
		}
		for id := range menuIDSet {
			menuIDs = append(menuIDs, id)
		}

		database.DB.Where("id IN ? AND del_flag = 0 AND status = 1 AND type IN ?", menuIDs, []string{"menu", "catalog", "embedded", "link"}).Find(&menus)
	}

	menuTree := buildMenuTree(menus)
	return menuTree, nil
}

func (s *AuthService) GetAuthCodes(userID string) ([]string, error) {
	return s.permissionService.GetUserCodes(userID)
}

func (s *AuthService) Logout(token string) error {
	utils.AddTokenToBlacklist(token)
	return nil
}

func buildMenuTree(menus []models.SysMenu) []*models.MenuTreeResponse {
	menuMap := make(map[string]*models.MenuTreeResponse)
	roots := []*models.MenuTreeResponse{}

	for _, menu := range menus {
		treeNode := &models.MenuTreeResponse{
			ID:        menu.ID,
			Pid:       menu.Pid,
			Type:      menu.Type,
			AuthCode:  menu.AuthCode,
			Children:  []*models.MenuTreeResponse{},
			Component: menu.Component,
			Meta: models.MenuMeta{
				ActiveIcon:               menu.ActiveIcon,
				ActivePath:               menu.ActivePath,
				AffixTab:                 menu.AffixTab == 1,
				AffixTabOrder:            menu.AffixTabOrder,
				Badge:                    menu.Badge,
				BadgeType:                menu.BadgeType,
				BadgeVariants:            menu.BadgeVariants,
				HideChildrenInMenu:       menu.HideChildrenInMenu == 1,
				HideInBreadcrumb:         menu.HideInBreadcrumb == 1,
				HideInMenu:               menu.HideInMenu == 1,
				HideInTab:                menu.HideInTab == 1,
				Icon:                     menu.Icon,
				IframeSrc:                menu.IframeSrc,
				KeepAlive:                menu.KeepAlive == 1,
				Link:                     menu.Link,
				MaxNumOfOpenTab:          menu.MaxNumOfOpenTab,
				NoBasicLayout:            menu.NoBasicLayout == 1,
				OpenInNewWindow:          menu.OpenInNewWindow == 1,
				Order:                    menu.Order,
				Query:                    menu.Query,
				Title:                    menu.Title,
				DomCached:                menu.DomCached == 1,
				MenuVisibleWithForbidden: menu.MenuVisibleWithForbidden == 1,
			},
			Name:        resolveMenuName(menu.Type, menu.ID, menu.Name),
			Path:        menu.Path,
			CreatorId:   menu.CreatorID,
			CreatorName: menu.CreatorName,
			Status:      menu.Status,
		}

		if menu.CreateDate != nil {
			createDateStr := menu.CreateDate.Format("2006-01-02 15:04:05")
			treeNode.CreateDate = &createDateStr
		}
		if menu.UpdateDate != nil {
			updateDateStr := menu.UpdateDate.Format("2006-01-02 15:04:05")
			treeNode.UpdateDate = &updateDateStr
		}

		menuMap[menu.ID] = treeNode
	}

	for _, menu := range menus {
		if menu.Pid == nil || *menu.Pid == "" {
			roots = append(roots, menuMap[menu.ID])
		} else {
			if parent, exists := menuMap[*menu.Pid]; exists {
				parent.Children = append(parent.Children, menuMap[menu.ID])
			}
		}
	}

	return roots
}

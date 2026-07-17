package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MenuService struct{}

var ErrMenuNameRequired = errors.New("非按钮菜单的路由名称不能为空")

func NewMenuService() *MenuService {
	return &MenuService{}
}

func (s *MenuService) GetMenuTree(req models.MenuListRequest) ([]*models.MenuTreeResponse, error) {
	query := database.DB.Model(&models.SysMenu{}).Where("del_flag = 0")

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Path != "" {
		query = query.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var menus []models.SysMenu
	err := query.Order("`order` asc, create_date desc").Find(&menus).Error
	if err != nil {
		return nil, err
	}

	return s.buildMenuTree(menus), nil
}

func (s *MenuService) CheckNameExists(name string, excludeId *string) (bool, error) {
	query := database.DB.Model(&models.SysMenu{}).Where("name = ? AND del_flag = 0", name)
	if excludeId != nil && *excludeId != "" {
		query = query.Where("id != ?", *excludeId)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *MenuService) CheckPathExists(path string, excludeId *string) (bool, error) {
	query := database.DB.Model(&models.SysMenu{}).Where("path = ? AND del_flag = 0", path)
	if excludeId != nil && *excludeId != "" {
		query = query.Where("id != ?", *excludeId)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *MenuService) CreateMenu(req models.CreateMenuRequest) error {
	name, err := normalizeMenuName(req.Type, req.Name)
	if err != nil {
		return err
	}

	now := time.Now()
	menu := models.SysMenu{
		ID:                       utils.GenerateUUID(),
		Pid:                      req.Pid,
		Type:                     req.Type,
		Name:                     name,
		Path:                     req.Path,
		Component:                req.Component,
		Icon:                     req.Meta.Icon,
		ActiveIcon:               req.Meta.ActiveIcon,
		ActivePath:               req.Meta.ActivePath,
		KeepAlive:                btoi(req.Meta.KeepAlive),
		HideInMenu:               btoi(req.Meta.HideInMenu),
		HideInTab:                btoi(req.Meta.HideInTab),
		HideInBreadcrumb:         btoi(req.Meta.HideInBreadcrumb),
		HideChildrenInMenu:       btoi(req.Meta.HideChildrenInMenu),
		Badge:                    req.Meta.Badge,
		BadgeType:                req.Meta.BadgeType,
		BadgeVariants:            req.Meta.BadgeVariants,
		AffixTab:                 btoi(req.Meta.AffixTab),
		AffixTabOrder:            req.Meta.AffixTabOrder,
		MaxNumOfOpenTab:          req.Meta.MaxNumOfOpenTab,
		NoBasicLayout:            btoi(req.Meta.NoBasicLayout),
		OpenInNewWindow:          btoi(req.Meta.OpenInNewWindow),
		DomCached:                btoi(req.Meta.DomCached),
		Query:                    req.Meta.Query,
		MenuVisibleWithForbidden: btoi(req.Meta.MenuVisibleWithForbidden),
		Order:                    req.Meta.Order,
		Title:                    req.Meta.Title,
		Link:                     req.Meta.Link,
		IframeSrc:                req.Meta.IframeSrc,
		Status:                   req.Status,
		CreateDate:               &now,
		UpdateDate:               &now,
		DelFlag:                  0,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		authCode, err := s.normalizeAndValidateAuthCode(tx, req.AuthCode, "")
		if err != nil {
			return err
		}
		menu.AuthCode = authCode
		return tx.Create(&menu).Error
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func (s *MenuService) GetMenuDetail(id string) (*models.MenuTreeResponse, error) {
	var menu models.SysMenu
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&menu).Error
	if err != nil {
		return nil, errors.New("菜单不存在")
	}

	return s.sysMenuToResponse(menu, []*models.MenuTreeResponse{}), nil
}

func (s *MenuService) UpdateMenu(id string, req models.UpdateMenuRequest) error {
	name, err := normalizeMenuName(req.Type, req.Name)
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		authCode, err := s.normalizeAndValidateAuthCode(tx, req.AuthCode, id)
		if err != nil {
			return err
		}

		var menu models.SysMenu
		if err := tx.Where("id = ? AND del_flag = 0", id).First(&menu).Error; err != nil {
			return errors.New("菜单不存在")
		}

		now := time.Now()
		menu.Pid = req.Pid
		menu.Type = req.Type
		menu.Name = name
		menu.Path = req.Path
		menu.Component = req.Component
		menu.AuthCode = authCode
		menu.Icon = req.Meta.Icon
		menu.ActiveIcon = req.Meta.ActiveIcon
		menu.ActivePath = req.Meta.ActivePath
		menu.KeepAlive = btoi(req.Meta.KeepAlive)
		menu.HideInMenu = btoi(req.Meta.HideInMenu)
		menu.HideInTab = btoi(req.Meta.HideInTab)
		menu.HideInBreadcrumb = btoi(req.Meta.HideInBreadcrumb)
		menu.HideChildrenInMenu = btoi(req.Meta.HideChildrenInMenu)
		menu.Badge = req.Meta.Badge
		menu.BadgeType = req.Meta.BadgeType
		menu.BadgeVariants = req.Meta.BadgeVariants
		menu.AffixTab = btoi(req.Meta.AffixTab)
		menu.AffixTabOrder = req.Meta.AffixTabOrder
		menu.MaxNumOfOpenTab = req.Meta.MaxNumOfOpenTab
		menu.NoBasicLayout = btoi(req.Meta.NoBasicLayout)
		menu.OpenInNewWindow = btoi(req.Meta.OpenInNewWindow)
		menu.DomCached = btoi(req.Meta.DomCached)
		menu.Query = req.Meta.Query
		menu.MenuVisibleWithForbidden = btoi(req.Meta.MenuVisibleWithForbidden)
		menu.Order = req.Meta.Order
		menu.Title = req.Meta.Title
		menu.Link = req.Meta.Link
		menu.IframeSrc = req.Meta.IframeSrc
		menu.Status = req.Status
		menu.UpdateDate = &now

		return tx.Save(&menu).Error
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func normalizeMenuName(menuType string, raw *string) (*string, error) {
	if menuType == "button" {
		return nil, nil
	}
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, ErrMenuNameRequired
	}
	name := strings.TrimSpace(*raw)
	return &name, nil
}

func (s *MenuService) normalizeAndValidateAuthCode(tx *gorm.DB, raw *string, excludeID string) (*string, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}

	normalized, err := NormalizePermissionBundle(*raw)
	if err != nil {
		return nil, err
	}

	var menus []models.SysMenu
	if err := tx.Model(&models.SysMenu{}).
		Select("id", "auth_code").
		Where("del_flag = 0").
		Order("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&menus).Error; err != nil {
		return nil, err
	}

	existingBundles := make([]string, 0, len(menus))
	for _, menu := range menus {
		if menu.ID != excludeID && menu.AuthCode != nil && *menu.AuthCode != "" {
			existingBundles = append(existingBundles, *menu.AuthCode)
		}
	}
	if duplicate := findDuplicatePermissionCode(normalized, existingBundles); duplicate != "" {
		return nil, fmt.Errorf("%w: %s 已存在", ErrPermissionCodeConflict, duplicate)
	}

	return &normalized, nil
}

func (s *MenuService) DeleteMenus(ids []string) error {
	for _, id := range ids {
		var childrenCount int64
		database.DB.Model(&models.SysMenu{}).Where("pid = ? AND del_flag = 0", id).Count(&childrenCount)
		if childrenCount > 0 {
			return errors.New("菜单存在子菜单，不能删除")
		}

		var menu models.SysMenu
		err := database.DB.Where("id = ? AND del_flag = 0", id).First(&menu).Error
		if err != nil {
			continue
		}

		database.DB.Where("menu_id = ?", id).Delete(&models.SysRoleMenu{})

		now := time.Now()
		menu.DelFlag = 1
		menu.UpdateDate = &now
		database.DB.Save(&menu)
	}

	return nil
}

func (s *MenuService) buildMenuTree(menus []models.SysMenu) []*models.MenuTreeResponse {
	menuMap := make(map[string]*models.MenuTreeResponse)
	var roots []*models.MenuTreeResponse

	for _, menu := range menus {
		menuMap[menu.ID] = s.sysMenuToResponse(menu, []*models.MenuTreeResponse{})
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

func (s *MenuService) sysMenuToResponse(menu models.SysMenu, children []*models.MenuTreeResponse) *models.MenuTreeResponse {
	return &models.MenuTreeResponse{
		ID:          menu.ID,
		Pid:         menu.Pid,
		Type:        menu.Type,
		AuthCode:    menu.AuthCode,
		Children:    children,
		Component:   menu.Component,
		Name:        menu.Name,
		Path:        menu.Path,
		CreatorId:   menu.CreatorID,
		CreatorName: menu.CreatorName,
		Status:      menu.Status,
		CreateDate:  models.TimeToStringPtr(menu.CreateDate),
		UpdateDate:  models.TimeToStringPtr(menu.UpdateDate),
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
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

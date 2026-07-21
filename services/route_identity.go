package services

import (
	"errors"
	"strings"

	"hive-admin-go/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRouteNameConflict = errors.New("路由名称已存在")
	ErrRoutePathConflict = errors.New("路由地址已存在")
)

func ensureActiveRouteIdentityUnique(tx *gorm.DB, name, path, excludeID string) error {
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)
	if name == "" && path == "" {
		return nil
	}

	query := tx.Model(&models.SysMenu{}).
		Select("id").
		Where("del_flag = 0 AND type != ?", "button")
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if name != "" {
		var row models.SysMenu
		err := query.Where("name = ?", name).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Take(&row).Error
		if err == nil {
			return ErrRouteNameConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if path != "" {
		var row models.SysMenu
		err := query.Where("path = ?", path).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Take(&row).Error
		if err == nil {
			return ErrRoutePathConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

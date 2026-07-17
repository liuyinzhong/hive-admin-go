package services

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

const maxPermissionBundleLength = 512

var permissionCodePattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*:[a-z][a-zA-Z0-9]*:[a-z][a-zA-Z0-9]*$`)

var (
	ErrInvalidPermissionCode  = errors.New("权限码校验失败")
	ErrPermissionCodeConflict = errors.New("权限码冲突")
)

type permissionSource interface {
	LoadUserPermissionBundles(userID string) (bool, []string, error)
}

type databasePermissionSource struct{}

type PermissionService struct {
	source permissionSource
}

func NewPermissionService() *PermissionService {
	return newPermissionService(databasePermissionSource{})
}

func newPermissionService(source permissionSource) *PermissionService {
	return &PermissionService{source: source}
}

func (s *PermissionService) GetUserCodes(userID string) ([]string, error) {
	_, bundles, err := s.source.LoadUserPermissionBundles(userID)
	if err != nil {
		return nil, err
	}
	return ExpandPermissionBundles(bundles), nil
}

func (s *PermissionService) HasCode(userID, requiredCode string) bool {
	isSystemUser, bundles, err := s.source.LoadUserPermissionBundles(userID)
	if isSystemUser {
		return true
	}
	if err != nil {
		return false
	}
	for _, code := range ExpandPermissionBundles(bundles) {
		if code == requiredCode {
			return true
		}
	}
	return false
}

func NormalizePermissionBundle(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	codes := strings.Split(raw, ",")
	codeSet := make(map[string]struct{}, len(codes))
	for _, rawCode := range codes {
		code := strings.TrimSpace(rawCode)
		if code == "" || !permissionCodePattern.MatchString(code) {
			return "", fmt.Errorf("%w: 格式不正确: %s", ErrInvalidPermissionCode, rawCode)
		}
		codeSet[code] = struct{}{}
	}

	normalizedCodes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		normalizedCodes = append(normalizedCodes, code)
	}
	sort.Strings(normalizedCodes)
	normalized := strings.Join(normalizedCodes, ",")
	if len(normalized) > maxPermissionBundleLength {
		return "", fmt.Errorf("%w: 长度不能超过%d个字符", ErrInvalidPermissionCode, maxPermissionBundleLength)
	}
	return normalized, nil
}

func ExpandPermissionBundles(bundles []string) []string {
	codeSet := make(map[string]struct{})
	for _, bundle := range bundles {
		for _, rawCode := range strings.Split(bundle, ",") {
			if code := strings.TrimSpace(rawCode); code != "" {
				codeSet[code] = struct{}{}
			}
		}
	}

	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func findDuplicatePermissionCode(candidate string, existingBundles []string) string {
	existingCodeSet := make(map[string]struct{})
	for _, code := range ExpandPermissionBundles(existingBundles) {
		existingCodeSet[code] = struct{}{}
	}
	for _, code := range ExpandPermissionBundles([]string{candidate}) {
		if _, exists := existingCodeSet[code]; exists {
			return code
		}
	}
	return ""
}

func (databasePermissionSource) LoadUserPermissionBundles(userID string) (bool, []string, error) {
	if database.DB == nil {
		return false, nil, errors.New("数据库未初始化")
	}

	var user models.SysUser
	if err := database.DB.
		Select("user_id", "is_sys").
		Where("user_id = ? AND del_flag = 0", userID).
		First(&user).Error; err != nil {
		return false, nil, err
	}

	var bundles []string
	if user.IsSys == 1 {
		err := database.DB.Model(&models.SysMenu{}).
			Distinct("auth_code").
			Where("status = 1 AND del_flag = 0 AND auth_code IS NOT NULL AND auth_code != ''").
			Pluck("auth_code", &bundles).Error
		return true, bundles, err
	}

	err := database.DB.Table("sys_user_role AS user_role").
		Distinct("menu.auth_code").
		Joins("JOIN sys_role AS role ON role.role_id = user_role.role_id AND role.status = 1 AND role.del_flag = 0").
		Joins("JOIN sys_role_menu AS role_menu ON role_menu.role_id = role.role_id AND role_menu.del_flag = 0").
		Joins("JOIN sys_menu AS menu ON menu.id = role_menu.menu_id AND menu.status = 1 AND menu.del_flag = 0").
		Where("user_role.user_id = ? AND user_role.del_flag = 0", userID).
		Where("menu.auth_code IS NOT NULL AND menu.auth_code != ''").
		Pluck("menu.auth_code", &bundles).Error
	return false, bundles, err
}

package services

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
)

func validateDataPermissionUsers(tx *gorm.DB, userIDs []string, permission datapermission.Permission) ([]string, error) {
	ids := uniqueNonEmptyStrings(userIDs)
	if err := validateDataPermissionUUIDs(ids, "用户ID"); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return ids, nil
	}

	query := tx.Model(&models.SysUser{}).
		Where("user_id IN ? AND status = 1 AND del_flag = 0 AND is_sys = 0", ids)
	var count int64
	if err := permission.Apply(query, "sys_user.user_id").Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, fmt.Errorf("关联用户不存在、已停用或超出数据权限范围")
	}
	return ids, nil
}

func validateRequiredDataPermissionUser(tx *gorm.DB, userID *string, label string, permission datapermission.Permission) (string, error) {
	if userID == nil || *userID == "" {
		return "", fmt.Errorf("%s不能为空", label)
	}
	ids, err := validateDataPermissionUsers(tx, []string{*userID}, permission)
	if err != nil {
		return "", err
	}
	if len(ids) != 1 {
		return "", fmt.Errorf("%s不能为空", label)
	}
	return ids[0], nil
}

func validateDataPermissionFiles(tx *gorm.DB, fileIDs []string, permission datapermission.Permission) ([]string, error) {
	ids := uniqueNonEmptyStrings(fileIDs)
	if err := validateDataPermissionUUIDs(ids, "文件ID"); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return ids, nil
	}

	query := tx.Model(&models.SysFile{}).Where("file_id IN ?", ids)
	var count int64
	if err := permission.Apply(query, "sys_file.creator_id").Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, fmt.Errorf("附件不存在或超出数据权限范围")
	}
	return ids, nil
}

func validateDevVersionReference(tx *gorm.DB, versionID *string, permission datapermission.Permission) error {
	if versionID == nil || *versionID == "" {
		return nil
	}
	if _, err := uuid.Parse(*versionID); err != nil {
		return fmt.Errorf("版本ID格式错误")
	}
	query := tx.Model(&models.DevVersion{}).Where("version_id = ? AND del_flag = 0", *versionID)
	var count int64
	if err := permission.Apply(query, "dev_version.creator_id").Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("版本不存在或超出数据权限范围")
	}
	return nil
}

func validateDevStoryReference(tx *gorm.DB, storyID *string, permission datapermission.Permission) error {
	if storyID == nil || *storyID == "" {
		return nil
	}
	if _, err := uuid.Parse(*storyID); err != nil {
		return fmt.Errorf("需求ID格式错误")
	}
	query := tx.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = 0", *storyID)
	query = permission.ApplyWithCSVUsers(query, []string{"dev_story.creator_id"}, []string{"dev_story.user_ids"})
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("需求不存在或超出数据权限范围")
	}
	return nil
}

func validateDataPermissionUUIDs(ids []string, label string) error {
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("%s格式错误", label)
		}
	}
	return nil
}

func dataPermissionStringValue(value interface{}, label string) (string, error) {
	result, ok := value.(string)
	if !ok || result == "" {
		return "", fmt.Errorf("%s必须是非空字符串", label)
	}
	return result, nil
}

func dataPermissionStringSlice(value interface{}, label string) ([]string, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s必须是字符串数组", label)
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		item, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s必须是字符串数组", label)
		}
		result = append(result, item)
	}
	return result, nil
}

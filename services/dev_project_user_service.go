package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// GetProjectUsers 查询项目用户列表,关联 sys_user 返回姓名和头像。
// 数据权限：公开接口,登录即可查看;不做记录级数据权限过滤。
func GetProjectUsers(projectID string) ([]models.ProjectUserResponse, error) {
	var rows []models.DevProjectUser
	err := database.DB.Where("project_id = ? AND del_flag = ?", projectID, 0).
		Order("create_date ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []models.ProjectUserResponse{}, nil
	}

	userIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}

	var users []models.SysUser
	err = database.DB.Where("user_id IN ? AND del_flag = ?", userIDs, 0).Find(&users).Error
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]*models.SysUser, len(users))
	for i := range users {
		userMap[users[i].UserID] = &users[i]
	}

	responses := make([]models.ProjectUserResponse, 0, len(rows))
	for _, r := range rows {
		resp := models.ProjectUserResponse{
			UserID:      r.UserID,
			StoryStatus: r.StoryStatus,
		}
		if u, ok := userMap[r.UserID]; ok {
			resp.RealName = stringValue(u.RealName)
			resp.Avatar = u.Avatar
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// SaveProjectUsers 全量替换项目用户,先删后插;同一项目同一用户只保留一行。
// 数据权限：需要 dev:project:user 权限码;不做记录级数据权限过滤(接口级权限控制)。
func SaveProjectUsers(req *models.SaveProjectUserRequest) error {
	// 校验项目存在
	var count int64
	if err := database.DB.Model(&models.DevProject{}).Where("project_id = ? AND del_flag = ?", req.ProjectID, 0).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("项目不存在")
	}

	// 校验用户ID去重
	seen := make(map[string]bool, len(req.Users))
	rows := make([]models.DevProjectUser, 0, len(req.Users))
	now := time.Now()
	for _, u := range req.Users {
		if u.UserID == "" || seen[u.UserID] {
			continue
		}
		seen[u.UserID] = true
		rows = append(rows, models.DevProjectUser{
			ID:          uuid.New().String(),
			ProjectID:   req.ProjectID,
			UserID:      u.UserID,
			StoryStatus: u.StoryStatus,
			CreateDate:  &now,
			UpdateDate:  &now,
			DelFlag:     0,
		})
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", req.ProjectID).Delete(&models.DevProjectUser{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

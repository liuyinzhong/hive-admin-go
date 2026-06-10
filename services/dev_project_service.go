package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

func GetAllProjects() ([]models.ProjectResponse, error) {
	var projects []models.DevProject
	err := database.DB.Where("del_flag = ?", 0).Order("create_date DESC").Find(&projects).Error
	if err != nil {
		return nil, err
	}

	var responses []models.ProjectResponse
	for _, project := range projects {
		responses = append(responses, models.ProjectResponse{
			ProjectID:    &project.ProjectID,
			ProjectTitle: project.ProjectTitle,
			ProjectLogo:  project.ProjectLogo,
			Description:  project.Description,
			CreateDate:   models.TimeToStringPtr(project.CreateDate),
		})
	}
	return responses, nil
}

func GetProjectByID(projectID string) (*models.ProjectResponse, error) {
	var project models.DevProject
	err := database.DB.Where("project_id = ? AND del_flag = ?", projectID, 0).First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("项目不存在")
		}
		return nil, err
	}

	return &models.ProjectResponse{
		ProjectID:    &project.ProjectID,
		ProjectTitle: project.ProjectTitle,
		ProjectLogo:  project.ProjectLogo,
		Description:  project.Description,
		CreateDate:   models.TimeToStringPtr(project.CreateDate),
	}, nil
}

func CreateProject(req *models.CreateProjectRequest, creatorID string) error {
	var count int64
	database.DB.Model(&models.DevProject{}).Where("project_title = ? AND del_flag = ?", req.ProjectTitle, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("项目标题已存在")
	}

	now := time.Now()
	project := models.DevProject{
		ProjectID:    uuid.New().String(),
		ProjectTitle: req.ProjectTitle,
		Description:  req.Description,
		ProjectLogo:  req.ProjectLogo,
		CreateDate:   &now,
		UpdateDate:   &now,
		DelFlag:      0,
	}

	return database.DB.Create(&project).Error
}

func UpdateProject(projectID string, req *models.UpdateProjectRequest) error {
	var project models.DevProject
	err := database.DB.Where("project_id = ? AND del_flag = ?", projectID, 0).First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("项目不存在")
		}
		return err
	}

	var count int64
	database.DB.Model(&models.DevProject{}).Where("project_title = ? AND project_id != ? AND del_flag = ?", req.ProjectTitle, projectID, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("项目标题已存在")
	}

	now := time.Now()
	return database.DB.Model(&project).Updates(map[string]interface{}{
		"project_title": req.ProjectTitle,
		"description":   req.Description,
		"project_logo":  req.ProjectLogo,
		"update_date":   now,
	}).Error
}

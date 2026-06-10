package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

func GetAllModules(params map[string]interface{}) ([]models.ModuleResponse, error) {
	db := database.DB.Model(&models.DevModule{}).Where("del_flag = ?", 0)

	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}

	var modules []models.DevModule
	err := db.Order("create_date DESC").Find(&modules).Error
	if err != nil {
		return nil, err
	}

	projectIDs := make([]string, 0)
	for _, m := range modules {
		projectIDs = append(projectIDs, m.ProjectID)
	}

	projects := make(map[string]string)
	if len(projectIDs) > 0 {
		var projectList []models.DevProject
		database.DB.Where("project_id IN ?", projectIDs).Find(&projectList)
		for _, p := range projectList {
			if p.ProjectTitle != nil {
				projects[p.ProjectID] = *p.ProjectTitle
			}
		}
	}

	var responses []models.ModuleResponse
	for _, module := range modules {
		projectTitle := projects[module.ProjectID]
		responses = append(responses, models.ModuleResponse{
			ModuleID:     &module.ModuleID,
			ModuleTitle:  module.ModuleTitle,
			ProjectID:    &module.ProjectID,
			ProjectTitle: &projectTitle,
			Sort:         module.Sort,
			UpdateDate:   models.TimeToStringPtr(module.UpdateDate),
			CreateDate:   models.TimeToStringPtr(module.CreateDate),
		})
	}
	return responses, nil
}

func GetModuleByID(moduleID string) (*models.ModuleResponse, error) {
	var module models.DevModule
	err := database.DB.Where("module_id = ? AND del_flag = ?", moduleID, 0).First(&module).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模块不存在")
		}
		return nil, err
	}

	var project models.DevProject
	var projectTitle string
	database.DB.Where("project_id = ?", module.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	return &models.ModuleResponse{
		ModuleID:     &module.ModuleID,
		ModuleTitle:  module.ModuleTitle,
		ProjectID:    &module.ProjectID,
		ProjectTitle: &projectTitle,
		Sort:         module.Sort,
		UpdateDate:   models.TimeToStringPtr(module.UpdateDate),
		CreateDate:   models.TimeToStringPtr(module.CreateDate),
	}, nil
}

func CreateModule(req *models.CreateModuleRequest, creatorID string) error {
	var count int64
	database.DB.Model(&models.DevModule{}).Where("module_title = ? AND project_id = ? AND del_flag = ?", req.ModuleTitle, req.ProjectID, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("同一项目下模块标题已存在")
	}

	now := time.Now()
	module := models.DevModule{
		ModuleID:    uuid.New().String(),
		ProjectID:   req.ProjectID,
		ModuleTitle: req.ModuleTitle,
		Sort:        req.Sort,
		CreateDate:  &now,
		UpdateDate:  &now,
		DelFlag:     0,
	}

	return database.DB.Create(&module).Error
}

func UpdateModule(moduleID string, req *models.UpdateModuleRequest) error {
	var module models.DevModule
	err := database.DB.Where("module_id = ? AND del_flag = ?", moduleID, 0).First(&module).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("模块不存在")
		}
		return err
	}

	var count int64
	database.DB.Model(&models.DevModule{}).Where("module_title = ? AND module_id != ? AND del_flag = ?", req.ModuleTitle, moduleID, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("模块标题已存在")
	}

	now := time.Now()
	return database.DB.Model(&module).Updates(map[string]interface{}{
		"module_title": req.ModuleTitle,
		"sort":         req.Sort,
		"update_date":  now,
	}).Error
}

func DeleteModules(moduleIDs []string) error {
	return database.DB.Model(&models.DevModule{}).Where("module_id IN ?", moduleIDs).Updates(map[string]interface{}{
		"del_flag":    1,
		"update_date": time.Now(),
	}).Error
}

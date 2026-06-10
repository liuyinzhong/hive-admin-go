package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func GetVersions(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevVersion{}).Where("del_flag = ?", 0)

	if version, ok := params["version"].(string); ok && version != "" {
		db = db.Where("version LIKE ?", "%"+version+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if releaseStatus, ok := params["releaseStatus"].(int); ok && releaseStatus >= 0 {
		db = db.Where("release_status = ?", releaseStatus)
	}

	sorts := params["sorts"].(string)
	order := utils.BuildOrderBy(sorts, map[string]string{
		"version":   "version",
		"startDate": "start_date",
		"endDate":   "end_date",
	})
	if order == "" {
		order = "create_date DESC"
	}

	return utils.PaginateWithTransform[models.DevVersion](db, page, pageSize, order, func(items []models.DevVersion) interface{} {
		return buildVersionResponses(items)
	})
}

func GetAllVersions(params map[string]interface{}) ([]models.VersionResponse, error) {
	db := database.DB.Model(&models.DevVersion{}).Where("del_flag = ?", 0)

	if version, ok := params["version"].(string); ok && version != "" {
		db = db.Where("version LIKE ?", "%"+version+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if releaseStatus, ok := params["releaseStatus"].(int); ok && releaseStatus >= 0 {
		db = db.Where("release_status = ?", releaseStatus)
	}

	var versions []models.DevVersion
	err := db.Order("create_date DESC").Find(&versions).Error
	if err != nil {
		return nil, err
	}

	return buildVersionResponses(versions), nil
}

func buildVersionResponses(versions []models.DevVersion) []models.VersionResponse {
	creatorIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	for _, v := range versions {
		if v.CreatorID != nil {
			creatorIDs = append(creatorIDs, *v.CreatorID)
		}
		projectIDs = append(projectIDs, v.ProjectID)
	}

	creators := make(map[string]string)
	if len(creatorIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", creatorIDs).Find(&users)
		for _, u := range users {
			if u.RealName != nil {
				creators[u.UserID] = *u.RealName
			}
		}
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

	var responses []models.VersionResponse
	for _, version := range versions {
		creatorName := creators[utils.StringValue(version.CreatorID)]
		projectTitle := projects[version.ProjectID]
		responses = append(responses, models.VersionResponse{
			VersionID:         &version.VersionID,
			Version:           version.Version,
			VersionType:       intToString(version.VersionType),
			Remark:            version.Remark,
			CreatorID:         version.CreatorID,
			CreatorName:       &creatorName,
			CreateDate:        models.TimeToStringPtr(version.CreateDate),
			EndDate:           models.TimeToStringPtr(version.EndDate),
			StartDate:         models.TimeToStringPtr(version.StartDate),
			ProjectID:         &version.ProjectID,
			ProjectTitle:      &projectTitle,
			ReleaseStatus:     intToString(version.ReleaseStatus),
			ReleaseDate:       models.TimeToStringPtr(version.ReleaseDate),
			ChangeLogRichText: version.ChangeLogRichText,
			ChangeLog:         version.ChangeLog,
		})
	}
	return responses
}

func GetVersionByID(versionID string) (*models.VersionResponse, error) {
	var version models.DevVersion
	err := database.DB.Where("version_id = ? AND del_flag = ?", versionID, 0).First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("版本不存在")
		}
		return nil, err
	}

	var creatorName string
	if version.CreatorID != nil {
		var user models.SysUser
		database.DB.Where("user_id = ?", *version.CreatorID).First(&user)
		if user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	var projectTitle string
	var project models.DevProject
	database.DB.Where("project_id = ?", version.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	return &models.VersionResponse{
		VersionID:         &version.VersionID,
		Version:           version.Version,
		VersionType:       intToString(version.VersionType),
		Remark:            version.Remark,
		CreatorID:         version.CreatorID,
		CreatorName:       &creatorName,
		CreateDate:        models.TimeToStringPtr(version.CreateDate),
		EndDate:           models.TimeToStringPtr(version.EndDate),
		StartDate:         models.TimeToStringPtr(version.StartDate),
		ProjectID:         &version.ProjectID,
		ProjectTitle:      &projectTitle,
		ReleaseStatus:     intToString(version.ReleaseStatus),
		ReleaseDate:       models.TimeToStringPtr(version.ReleaseDate),
		ChangeLogRichText: version.ChangeLogRichText,
		ChangeLog:         version.ChangeLog,
	}, nil
}

func GetLatestVersion(projectID string) (*models.VersionResponse, error) {
	var versions []models.DevVersion
	err := database.DB.Where("project_id = ? AND del_flag = ?", projectID, 0).Order("create_date DESC").Find(&versions).Error
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, nil
	}

	var latestVersion *models.DevVersion
	var latestVersionStr string

	for i := range versions {
		v := &versions[i]
		if v.Version != nil {
			if latestVersion == nil || compareVersions(*v.Version, latestVersionStr) > 0 {
				latestVersion = v
				latestVersionStr = *v.Version
			}
		}
	}

	if latestVersion == nil {
		return nil, nil
	}

	var creatorName string
	if latestVersion.CreatorID != nil {
		var user models.SysUser
		database.DB.Where("user_id = ?", *latestVersion.CreatorID).First(&user)
		if user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	var projectTitle string
	var project models.DevProject
	database.DB.Where("project_id = ?", latestVersion.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	return &models.VersionResponse{
		VersionID:         &latestVersion.VersionID,
		Version:           latestVersion.Version,
		VersionType:       intToString(latestVersion.VersionType),
		Remark:            latestVersion.Remark,
		CreatorID:         latestVersion.CreatorID,
		CreatorName:       &creatorName,
		CreateDate:        models.TimeToStringPtr(latestVersion.CreateDate),
		EndDate:           models.TimeToStringPtr(latestVersion.EndDate),
		StartDate:         models.TimeToStringPtr(latestVersion.StartDate),
		ProjectID:         &latestVersion.ProjectID,
		ProjectTitle:      &projectTitle,
		ReleaseStatus:     intToString(latestVersion.ReleaseStatus),
		ReleaseDate:       models.TimeToStringPtr(latestVersion.ReleaseDate),
		ChangeLogRichText: latestVersion.ChangeLogRichText,
		ChangeLog:         latestVersion.ChangeLog,
	}, nil
}

func CreateVersion(req *models.CreateVersionRequest, creatorID string) error {
	var count int64
	database.DB.Model(&models.DevVersion{}).Where("version = ? AND project_id = ? AND del_flag = ?", req.Version, req.ProjectID, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("同一项目下版本号已存在")
	}

	versionID := uuid.New().String()
	now := time.Now()

	startDate, err := parseLocalDateTime(req.StartDate)
	if err != nil {
		return err
	}
	endDate, err := parseLocalDateTime(req.EndDate)
	if err != nil {
		return err
	}
	versionType, err := parseStringInt(req.VersionType, "versionType")
	if err != nil {
		return err
	}
	releaseStatus, err := parseStringInt(req.ReleaseStatus, "releaseStatus")
	if err != nil {
		return err
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return fmt.Errorf("结束时间不能早于开始时间")
	}

	version := models.DevVersion{
		VersionID:     versionID,
		Version:       req.Version,
		VersionType:   versionType,
		ReleaseStatus: releaseStatus,
		ProjectID:     req.ProjectID,
		Remark:        req.Remark,
		EndDate:       endDate,
		StartDate:     startDate,
		CreatorID:     &creatorID,
		CreateDate:    &now,
		UpdateDate:    &now,
		DelFlag:       0,
	}

	err = database.DB.Create(&version).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, versionID, 30, 0, "")
}

func UpdateVersion(versionID string, req *models.UpdateVersionRequest, creatorID string) error {
	var version models.DevVersion
	err := database.DB.Where("version_id = ? AND del_flag = ?", versionID, 0).First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("版本不存在")
		}
		return err
	}

	var count int64
	database.DB.Model(&models.DevVersion{}).Where("version = ? AND project_id = ? AND version_id != ? AND del_flag = ?", req.Version, req.ProjectID, versionID, 0).Count(&count)
	if count > 0 {
		return fmt.Errorf("同一项目下版本号已存在")
	}

	startDate, err := parseLocalDateTime(req.StartDate)
	if err != nil {
		return err
	}
	endDate, err := parseLocalDateTime(req.EndDate)
	if err != nil {
		return err
	}
	versionType, err := parseStringInt(req.VersionType, "versionType")
	if err != nil {
		return err
	}
	releaseStatus, err := parseStringInt(req.ReleaseStatus, "releaseStatus")
	if err != nil {
		return err
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return fmt.Errorf("结束时间不能早于开始时间")
	}

	now := time.Now()
	err = database.DB.Model(&version).Updates(map[string]interface{}{
		"version":        req.Version,
		"version_type":   versionType,
		"release_status": releaseStatus,
		"project_id":     req.ProjectID,
		"remark":         req.Remark,
		"end_date":       endDate,
		"start_date":     startDate,
		"update_date":    now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, versionID, 30, 10, "")
}

func UpdateVersionNext(versionID string, releaseStatus string, changeRichText string, creatorID string) error {
	var version models.DevVersion
	err := database.DB.Where("version_id = ? AND del_flag = ?", versionID, 0).First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("版本不存在")
		}
		return err
	}

	releaseStatusInt, err := parseStringInt(releaseStatus, "releaseStatus")
	if err != nil {
		return err
	}

	err = database.DB.Model(&version).Update("release_status", releaseStatusInt).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, versionID, 30, 40, changeRichText)
}

func DeleteVersions(versionIDs []string, creatorID string) error {
	var versions []models.DevVersion
	err := database.DB.Where("version_id IN ? AND del_flag = ?", versionIDs, 0).Find(&versions).Error
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.ReleaseStatus != 0 {
			return fmt.Errorf("只能删除发布状态为待评审的版本")
		}
	}

	err = database.DB.Model(&models.DevVersion{}).Where("version_id IN ?", versionIDs).Updates(map[string]interface{}{
		"del_flag":    1,
		"update_date": time.Now(),
	}).Error
	if err != nil {
		return err
	}

	for _, v := range versions {
		createChangeHistory(creatorID, v.VersionID, 30, 20, "")
	}
	return nil
}

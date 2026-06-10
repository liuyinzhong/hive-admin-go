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

func GetBugs(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevBug{}).Where("del_flag = ?", 0)

	if bugNum, ok := params["bugNum"].(int); ok && bugNum > 0 {
		db = db.Where("bug_num = ?", bugNum)
	}
	if bugTitle, ok := params["bugTitle"].(string); ok && bugTitle != "" {
		db = db.Where("bug_title LIKE ?", "%"+bugTitle+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if versionID, ok := params["versionId"].(string); ok && versionID != "" {
		db = db.Where("version_id = ?", versionID)
	}
	if moduleID, ok := params["moduleId"].(string); ok && moduleID != "" {
		db = db.Where("module_id = ?", moduleID)
	}
	if bugStatuses, ok := params["bugStatuses"].([]int); ok && len(bugStatuses) > 0 {
		db = db.Where("bug_status IN ?", bugStatuses)
	}
	if storyID, ok := params["storyId"].(string); ok && storyID != "" {
		db = db.Where("story_id = ?", storyID)
	}

	sorts := params["sorts"].(string)
	order := utils.BuildOrderBy(sorts, map[string]string{
		"bugTitle":         "bug_title",
		"bugStatus":        "bug_status",
		"bugConfirmStatus": "bug_confirm_status",
		"bugLevel":         "bug_level",
	})
	if order == "" {
		order = "create_date DESC"
	}

	return utils.PaginateWithTransform[models.DevBug](db, page, pageSize, order, func(items []models.DevBug) interface{} {
		return buildBugResponses(items)
	})
}

func GetAllBugs(params map[string]interface{}) ([]models.BugResponse, error) {
	db := database.DB.Model(&models.DevBug{}).Where("del_flag = ?", 0)

	if bugNum, ok := params["bugNum"].(int); ok && bugNum > 0 {
		db = db.Where("bug_num = ?", bugNum)
	}
	if bugTitle, ok := params["bugTitle"].(string); ok && bugTitle != "" {
		db = db.Where("bug_title LIKE ?", "%"+bugTitle+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if versionID, ok := params["versionId"].(string); ok && versionID != "" {
		db = db.Where("version_id = ?", versionID)
	}
	if moduleID, ok := params["moduleId"].(string); ok && moduleID != "" {
		db = db.Where("module_id = ?", moduleID)
	}
	if bugStatus, ok := params["bugStatus"].(int); ok && bugStatus >= 0 {
		db = db.Where("bug_status = ?", bugStatus)
	}
	if storyID, ok := params["storyId"].(string); ok && storyID != "" {
		db = db.Where("story_id = ?", storyID)
	}

	var bugs []models.DevBug
	err := db.Order("create_date DESC").Find(&bugs).Error
	if err != nil {
		return nil, err
	}

	return buildBugResponses(bugs), nil
}

func buildBugResponses(bugs []models.DevBug) []models.BugResponse {
	creatorIDs := make([]string, 0)
	userIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	versionIDs := make([]string, 0)
	moduleIDs := make([]string, 0)
	storyIDs := make([]string, 0)

	for _, b := range bugs {
		if b.CreatorID != nil {
			creatorIDs = append(creatorIDs, *b.CreatorID)
		}
		if b.UserID != nil {
			userIDs = append(userIDs, *b.UserID)
		}
		projectIDs = append(projectIDs, b.ProjectID)
		if b.VersionID != nil {
			versionIDs = append(versionIDs, *b.VersionID)
		}
		if b.ModuleID != nil {
			moduleIDs = append(moduleIDs, *b.ModuleID)
		}
		if b.StoryID != nil {
			storyIDs = append(storyIDs, *b.StoryID)
		}
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

	usersMap := make(map[string]models.StoryUserItem)
	if len(userIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", userIDs).Find(&users)
		for _, u := range users {
			usersMap[u.UserID] = models.StoryUserItem{
				UserID:   &u.UserID,
				Avatar:   u.Avatar,
				RealName: u.RealName,
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

	versions := make(map[string]string)
	if len(versionIDs) > 0 {
		var versionList []models.DevVersion
		database.DB.Where("version_id IN ?", versionIDs).Find(&versionList)
		for _, v := range versionList {
			if v.Version != nil {
				versions[v.VersionID] = *v.Version
			}
		}
	}

	modules := make(map[string]string)
	if len(moduleIDs) > 0 {
		var moduleList []models.DevModule
		database.DB.Where("module_id IN ?", moduleIDs).Find(&moduleList)
		for _, m := range moduleList {
			if m.ModuleTitle != nil {
				modules[m.ModuleID] = *m.ModuleTitle
			}
		}
	}

	stories := make(map[string]string)
	if len(storyIDs) > 0 {
		var storyList []models.DevStory
		database.DB.Where("story_id IN ?", storyIDs).Find(&storyList)
		for _, s := range storyList {
			if s.StoryTitle != nil {
				stories[s.StoryID] = *s.StoryTitle
			}
		}
	}

	var responses []models.BugResponse
	for _, bug := range bugs {
		creatorName := creators[utils.StringValue(bug.CreatorID)]
		projectTitle := projects[bug.ProjectID]
		version := versions[utils.StringValue(bug.VersionID)]
		moduleTitle := modules[utils.StringValue(bug.ModuleID)]
		storyTitle := stories[utils.StringValue(bug.StoryID)]

		var realName, avatar *string
		if bug.UserID != nil {
			if u, ok := usersMap[*bug.UserID]; ok {
				realName = u.RealName
				avatar = u.Avatar
			}
		}

		responses = append(responses, models.BugResponse{
			BugID:            &bug.BugID,
			BugTitle:         bug.BugTitle,
			BugNum:           bug.BugNum,
			BugStatus:        intToString(bug.BugStatus),
			BugConfirmStatus: intToString(bug.BugConfirmStatus),
			BugLevel:         intToString(bug.BugLevel),
			BugSource:        intToString(bug.BugSource),
			BugType:          intToString(bug.BugType),
			BugEnv:           intToString(bug.BugEnv),
			BugUa:            bug.BugUa,
			UserID:           bug.UserID,
			Avatar:           avatar,
			RealName:         realName,
			CreatorName:      &creatorName,
			CreatorID:        bug.CreatorID,
			VersionID:        bug.VersionID,
			Version:          &version,
			ModuleID:         bug.ModuleID,
			ModuleTitle:      &moduleTitle,
			ProjectID:        &bug.ProjectID,
			ProjectTitle:     &projectTitle,
			StoryID:          bug.StoryID,
			StoryTitle:       &storyTitle,
			UpdateDate:       models.TimeToStringPtr(bug.UpdateDate),
			CreateDate:       models.TimeToStringPtr(bug.CreateDate),
			BugRichText:      bug.BugRichText,
		})
	}
	return responses
}

func GetBugByNum(bugNum int) (*models.BugResponse, error) {
	var bug models.DevBug
	err := database.DB.Where("bug_num = ? AND del_flag = ?", bugNum, 0).First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("缺陷不存在")
		}
		return nil, err
	}

	return buildSingleBugResponse(&bug), nil
}

func GetBugsByStoryId(storyId string) ([]models.BugResponse, error) {
	var bugs []models.DevBug
	err := database.DB.Where("story_id = ? AND del_flag = ?", storyId, 0).Find(&bugs).Error
	if err != nil {
		return nil, err
	}

	var responses []models.BugResponse
	for _, bug := range bugs {
		resp := buildSingleBugResponse(&bug)
		responses = append(responses, *resp)
	}

	return responses, nil
}

func buildSingleBugResponse(bug *models.DevBug) *models.BugResponse {
	var creatorName string
	if bug.CreatorID != nil {
		var user models.SysUser
		database.DB.Where("user_id = ?", *bug.CreatorID).First(&user)
		if user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	var realName, avatar *string
	if bug.UserID != nil {
		var u models.SysUser
		database.DB.Where("user_id = ?", *bug.UserID).First(&u)
		realName = u.RealName
		avatar = u.Avatar
	}

	var projectTitle string
	var project models.DevProject
	database.DB.Where("project_id = ?", bug.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	var version string
	if bug.VersionID != nil {
		var v models.DevVersion
		database.DB.Where("version_id = ?", *bug.VersionID).First(&v)
		if v.Version != nil {
			version = *v.Version
		}
	}

	var moduleTitle string
	if bug.ModuleID != nil {
		var m models.DevModule
		database.DB.Where("module_id = ?", *bug.ModuleID).First(&m)
		if m.ModuleTitle != nil {
			moduleTitle = *m.ModuleTitle
		}
	}

	var storyTitle string
	if bug.StoryID != nil {
		var s models.DevStory
		database.DB.Where("story_id = ?", *bug.StoryID).First(&s)
		if s.StoryTitle != nil {
			storyTitle = *s.StoryTitle
		}
	}

	return &models.BugResponse{
		BugID:            &bug.BugID,
		BugTitle:         bug.BugTitle,
		BugNum:           bug.BugNum,
		BugRichText:      bug.BugRichText,
		BugStatus:        intToString(bug.BugStatus),
		BugConfirmStatus: intToString(bug.BugConfirmStatus),
		BugLevel:         intToString(bug.BugLevel),
		BugSource:        intToString(bug.BugSource),
		BugType:          intToString(bug.BugType),
		BugEnv:           intToString(bug.BugEnv),
		BugUa:            bug.BugUa,
		UserID:           bug.UserID,
		Avatar:           avatar,
		RealName:         realName,
		CreatorName:      &creatorName,
		CreatorID:        bug.CreatorID,
		VersionID:        bug.VersionID,
		Version:          &version,
		ModuleID:         bug.ModuleID,
		ModuleTitle:      &moduleTitle,
		ProjectID:        &bug.ProjectID,
		ProjectTitle:     &projectTitle,
		StoryID:          bug.StoryID,
		StoryTitle:       &storyTitle,
		UpdateDate:       models.TimeToStringPtr(bug.UpdateDate),
		CreateDate:       models.TimeToStringPtr(bug.CreateDate),
	}
}

func CreateBug(req *models.CreateBugRequest, creatorID string) error {
	bugID := uuid.New().String()
	now := time.Now()
	bugStatus, err := parseStringInt(req.BugStatus, "bugStatus")
	if err != nil {
		return err
	}
	bugLevel, err := parseStringInt(req.BugLevel, "bugLevel")
	if err != nil {
		return err
	}
	bugSource, err := parseStringInt(req.BugSource, "bugSource")
	if err != nil {
		return err
	}
	bugType, err := parseStringInt(req.BugType, "bugType")
	if err != nil {
		return err
	}
	bugEnv, err := parseStringInt(req.BugEnv, "bugEnv")
	if err != nil {
		return err
	}

	bug := models.DevBug{
		BugID:            bugID,
		BugTitle:         req.BugTitle,
		BugStatus:        bugStatus,
		BugConfirmStatus: 0,
		BugLevel:         bugLevel,
		BugSource:        bugSource,
		BugType:          bugType,
		BugEnv:           bugEnv,
		BugUa:            req.BugUa,
		ProjectID:        req.ProjectID,
		BugRichText:      req.BugRichText,
		VersionID:        req.VersionID,
		ModuleID:         req.ModuleID,
		StoryID:          req.StoryID,
		UserID:           req.UserID,
		CreatorID:        &creatorID,
		CreateDate:       &now,
		UpdateDate:       &now,
		DelFlag:          0,
	}

	err = database.DB.Create(&bug).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, bugID, 20, 0, "")
}

func CreateBugs(reqs []models.CreateBugRequest, creatorID string) error {
	for _, req := range reqs {
		err := CreateBug(&req, creatorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateBug(bugID string, req *models.UpdateBugRequest, creatorID string) error {
	var bug models.DevBug
	err := database.DB.Where("bug_id = ? AND del_flag = ?", bugID, 0).First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}
	bugStatus, err := parseStringInt(req.BugStatus, "bugStatus")
	if err != nil {
		return err
	}
	bugLevel, err := parseStringInt(req.BugLevel, "bugLevel")
	if err != nil {
		return err
	}
	bugSource, err := parseStringInt(req.BugSource, "bugSource")
	if err != nil {
		return err
	}
	bugType, err := parseStringInt(req.BugType, "bugType")
	if err != nil {
		return err
	}
	bugEnv, err := parseStringInt(req.BugEnv, "bugEnv")
	if err != nil {
		return err
	}

	now := time.Now()
	err = database.DB.Model(&bug).Updates(map[string]interface{}{
		"bug_title":     req.BugTitle,
		"bug_status":    bugStatus,
		"bug_level":     bugLevel,
		"bug_source":    bugSource,
		"bug_type":      bugType,
		"bug_env":       bugEnv,
		"bug_ua":        req.BugUa,
		"project_id":    req.ProjectID,
		"bug_rich_text": req.BugRichText,
		"version_id":    req.VersionID,
		"module_id":     req.ModuleID,
		"story_id":      req.StoryID,
		"user_id":       req.UserID,
		"update_date":   now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, bugID, 20, 10, "")
}

func UpdateBugField(bugID string, key string, value interface{}, creatorID string) error {
	var bug models.DevBug
	err := database.DB.Where("bug_id = ? AND del_flag = ?", bugID, 0).First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}

	allowedKeys := map[string]bool{"userId": true, "bugLevel": true, "bugEnv": true, "bugType": true, "bugSource": true}
	if !allowedKeys[key] {
		return fmt.Errorf("只能修改 userId、bugLevel、bugEnv、bugType、bugSource 字段")
	}

	updateMap := make(map[string]interface{})
	switch key {
	case "userId":
		updateMap["user_id"] = value
	case "bugLevel":
		updateMap["bug_level"] = value
	case "bugEnv":
		updateMap["bug_env"] = value
	case "bugType":
		updateMap["bug_type"] = value
	case "bugSource":
		updateMap["bug_source"] = value
	}

	updateMap["update_date"] = time.Now()

	err = database.DB.Model(&bug).Updates(updateMap).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, bugID, 20, 10, "")
}

func UpdateBugNext(bugID string, bugStatus string, changeRichText string, creatorID string) error {
	var bug models.DevBug
	err := database.DB.Where("bug_id = ? AND del_flag = ?", bugID, 0).First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}

	bugStatusInt, err := parseStringInt(bugStatus, "bugStatus")
	if err != nil {
		return err
	}

	now := time.Now()
	updateMap := map[string]interface{}{
		"bug_status":  bugStatusInt,
		"update_date": now,
	}

	if bugStatusInt == 0 {
		updateMap["bug_confirm_status"] = 0
	}

	err = database.DB.Model(&bug).Updates(updateMap).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, bugID, 20, 40, changeRichText)
}

func ConfirmBug(bugID string, req *models.ConfirmBugRequest, creatorID string) error {
	var bug models.DevBug
	err := database.DB.Where("bug_id = ? AND del_flag = ?", bugID, 0).First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}

	bugConfirmStatus, err := parseStringInt(req.BugConfirmStatus, "bugConfirmStatus")
	if err != nil {
		return err
	}

	now := time.Now()
	updateMap := map[string]interface{}{
		"bug_confirm_status": bugConfirmStatus,
		"update_date":        now,
	}

	if bugConfirmStatus == 1 {
		updateMap["bug_status"] = 10
	} else if bugConfirmStatus == 2 {
		updateMap["bug_status"] = 1
	}

	err = database.DB.Model(&bug).Updates(updateMap).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, bugID, 20, 50, req.ChangeRichText)
}

func DeleteBugs(bugIDs []string, creatorID string) error {
	var bugs []models.DevBug
	err := database.DB.Where("bug_id IN ? AND del_flag = ?", bugIDs, 0).Find(&bugs).Error
	if err != nil {
		return err
	}

	for _, b := range bugs {
		if b.BugStatus != 0 {
			return fmt.Errorf("只能删除状态为待确认的缺陷")
		}
	}

	err = database.DB.Model(&models.DevBug{}).Where("bug_id IN ?", bugIDs).Updates(map[string]interface{}{
		"del_flag":    1,
		"update_date": time.Now(),
	}).Error
	if err != nil {
		return err
	}

	for _, b := range bugs {
		createChangeHistory(creatorID, b.BugID, 20, 20, "")
	}
	return nil
}

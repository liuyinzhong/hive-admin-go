package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func GetBugs(page, pageSize int, params map[string]interface{}, permission datapermission.Permission) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevBug{}).Where("dev_bug.del_flag = ?", 0)
	db = permission.Apply(db, "dev_bug.creator_id", "dev_bug.fix_user_id")

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

func GetAllBugs(params map[string]interface{}, permission datapermission.Permission) ([]models.BugResponse, error) {
	db := database.DB.Model(&models.DevBug{}).Where("dev_bug.del_flag = ?", 0)
	db = permission.Apply(db, "dev_bug.creator_id", "dev_bug.fix_user_id")

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
	fixUserIDs := make([]string, 0)
	verifierIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	versionIDs := make([]string, 0)
	moduleIDs := make([]string, 0)
	storyIDs := make([]string, 0)
	allFileIDs := make([]string, 0)

	for _, b := range bugs {
		if b.CreatorID != nil {
			creatorIDs = append(creatorIDs, *b.CreatorID)
		}
		if b.FixUserID != nil {
			fixUserIDs = append(fixUserIDs, *b.FixUserID)
		}
		if b.VerifierID != nil {
			verifierIDs = append(verifierIDs, *b.VerifierID)
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
		if b.FileIDs != nil {
			ids := strings.Split(*b.FileIDs, ",")
			for _, id := range ids {
				if id != "" {
					allFileIDs = append(allFileIDs, id)
				}
			}
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

	// 修复人/验证人合并查询 sys_user，避免重复查询
	allUserIDs := make([]string, 0, len(fixUserIDs)+len(verifierIDs))
	allUserIDs = append(allUserIDs, fixUserIDs...)
	allUserIDs = append(allUserIDs, verifierIDs...)
	userItemMap := make(map[string]models.BugUserItem)
	if len(allUserIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", allUserIDs).Find(&users)
		for _, u := range users {
			userItemMap[u.UserID] = models.BugUserItem{
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

	fileMap := make(map[string]models.FileResponse)
	if len(allFileIDs) > 0 {
		var files []models.SysFile
		database.DB.Where("file_id IN ?", allFileIDs).Find(&files)

		var fileCreatorIDs []string
		for _, f := range files {
			if f.CreatorID != nil {
				fileCreatorIDs = append(fileCreatorIDs, *f.CreatorID)
			}
		}

		fileCreatorNameMap := make(map[string]string)
		if len(fileCreatorIDs) > 0 {
			var fileCreators []models.SysUser
			database.DB.Where("user_id IN ?", fileCreatorIDs).Find(&fileCreators)
			for _, c := range fileCreators {
				if c.RealName != nil {
					fileCreatorNameMap[c.UserID] = *c.RealName
				}
			}
		}

		for _, f := range files {
			fileMap[f.FileID] = *buildFileResponse(f, fileCreatorNameMap[utils.StringValue(f.CreatorID)])
		}
	}

	var responses []models.BugResponse
	for _, bug := range bugs {
		creatorName := creators[utils.StringValue(bug.CreatorID)]
		projectTitle := projects[bug.ProjectID]
		version := versions[utils.StringValue(bug.VersionID)]
		moduleTitle := modules[utils.StringValue(bug.ModuleID)]
		storyTitle := stories[utils.StringValue(bug.StoryID)]

		var fixUserInfo, verifierUserInfo *models.BugUserItem
		if bug.FixUserID != nil {
			item := models.BugUserItem{UserID: bug.FixUserID}
			if u, ok := userItemMap[*bug.FixUserID]; ok {
				item.Avatar = u.Avatar
				item.RealName = u.RealName
			}
			fixUserInfo = &item
		}
		if bug.VerifierID != nil {
			item := models.BugUserItem{UserID: bug.VerifierID}
			if u, ok := userItemMap[*bug.VerifierID]; ok {
				item.Avatar = u.Avatar
				item.RealName = u.RealName
			}
			verifierUserInfo = &item
		}

		fileIDs := make([]string, 0)
		fileList := make([]models.FileResponse, 0)
		if bug.FileIDs != nil {
			ids := strings.Split(*bug.FileIDs, ",")
			for _, id := range ids {
				if id != "" {
					fileIDs = append(fileIDs, id)
					if f, ok := fileMap[id]; ok {
						fileList = append(fileList, f)
					}
				}
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
			FixUserID:        bug.FixUserID,
			FixUserInfo:      fixUserInfo,
			VerifierID:       bug.VerifierID,
			VerifierUserInfo: verifierUserInfo,
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
			FileIDs:          fileIDs,
			FileList:         fileList,
		})
	}
	return responses
}

func GetBugByNum(bugNum int, permission datapermission.Permission) (*models.BugResponse, error) {
	var bug models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_num = ? AND del_flag = ?", bugNum, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("缺陷不存在")
		}
		return nil, err
	}

	return buildSingleBugResponse(&bug), nil
}

func GetBugsByStoryId(storyId string, permission datapermission.Permission) ([]models.BugResponse, error) {
	var bugs []models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("story_id = ? AND del_flag = ?", storyId, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").Find(&bugs).Error
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

	// 修复人信息：FixUserID 存在时返回，avatar/realName 在用户存在时填充
	var fixUserInfo *models.BugUserItem
	if bug.FixUserID != nil {
		item := models.BugUserItem{UserID: bug.FixUserID}
		var u models.SysUser
		if err := database.DB.Where("user_id = ?", *bug.FixUserID).First(&u).Error; err == nil {
			item.Avatar = u.Avatar
			item.RealName = u.RealName
		}
		fixUserInfo = &item
	}
	// 验证人信息：VerifierID 存在时返回，avatar/realName 在用户存在时填充
	var verifierUserInfo *models.BugUserItem
	if bug.VerifierID != nil {
		item := models.BugUserItem{UserID: bug.VerifierID}
		var u models.SysUser
		if err := database.DB.Where("user_id = ?", *bug.VerifierID).First(&u).Error; err == nil {
			item.Avatar = u.Avatar
			item.RealName = u.RealName
		}
		verifierUserInfo = &item
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

	fileIDs := make([]string, 0)
	fileList := make([]models.FileResponse, 0)
	if bug.FileIDs != nil {
		ids := strings.Split(*bug.FileIDs, ",")
		for _, id := range ids {
			if id != "" {
				fileIDs = append(fileIDs, id)
				var f models.SysFile
				database.DB.Where("file_id = ?", id).First(&f)
				var fileCreatorName string
				if f.CreatorID != nil {
					var u models.SysUser
					database.DB.Where("user_id = ?", *f.CreatorID).First(&u)
					if u.RealName != nil {
						fileCreatorName = *u.RealName
					}
				}
				fileList = append(fileList, *buildFileResponse(f, fileCreatorName))
			}
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
		FixUserID:        bug.FixUserID,
		FixUserInfo:      fixUserInfo,
		VerifierID:       bug.VerifierID,
		VerifierUserInfo: verifierUserInfo,
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
		FileIDs:          fileIDs,
		FileList:         fileList,
	}
}

func CreateBug(req *models.CreateBugRequest, creatorID string, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return createBugTx(tx, req, creatorID, permission)
	})
}

func createBugTx(tx *gorm.DB, req *models.CreateBugRequest, creatorID string, permission datapermission.Permission) error {
	assigneeID, err := validateRequiredDataPermissionUser(tx, req.FixUserID, "修复人", permission)
	if err != nil {
		return err
	}
	fileIDs, err := validateDataPermissionFiles(tx, req.FileIDs, permission)
	if err != nil {
		return err
	}
	if err := validateDevVersionReference(tx, req.VersionID, permission); err != nil {
		return err
	}
	if err := validateDevStoryReference(tx, req.StoryID, permission); err != nil {
		return err
	}

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

	fileIDsStr := ""
	if len(fileIDs) > 0 {
		fileIDsStr = strings.Join(fileIDs, ",")
	}
	var fileIDsPtr *string
	if fileIDsStr != "" {
		fileIDsPtr = &fileIDsStr
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
		FileIDs:          fileIDsPtr,
		VersionID:        req.VersionID,
		ModuleID:         req.ModuleID,
		StoryID:          req.StoryID,
		FixUserID:        &assigneeID,
		CreatorID:        &creatorID,
		CreateDate:       &now,
		UpdateDate:       &now,
		DelFlag:          0,
	}

	err = tx.Create(&bug).Error
	if err != nil {
		return err
	}

	return createChangeHistoryTx(tx, creatorID, bugID, 20, 0, "")
}

func CreateBugs(reqs []models.CreateBugRequest, creatorID string, permission datapermission.Permission) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for index := range reqs {
			if err := createBugTx(tx, &reqs[index], creatorID, permission); err != nil {
				return fmt.Errorf("第%d条缺陷：%w", index+1, err)
			}
		}
		return nil
	})
}

func UpdateBug(bugID string, req *models.UpdateBugRequest, creatorID string, permission datapermission.Permission) error {
	var bug models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}
	assigneeID, err := validateRequiredDataPermissionUser(database.DB, req.FixUserID, "修复人", permission)
	if err != nil {
		return err
	}
	fileIDs, err := validateDataPermissionFiles(database.DB, req.FileIDs, permission)
	if err != nil {
		return err
	}
	if err := validateDevVersionReference(database.DB, req.VersionID, permission); err != nil {
		return err
	}
	if err := validateDevStoryReference(database.DB, req.StoryID, permission); err != nil {
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

	fileIDsStr := ""
	if len(fileIDs) > 0 {
		fileIDsStr = strings.Join(fileIDs, ",")
	}

	now := time.Now()
	return updateDevRecordWithHistory(creatorID, bugID, 20, 10, "", func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
		result := permission.Apply(updateQuery, "dev_bug.creator_id", "dev_bug.fix_user_id").Updates(map[string]interface{}{
			"bug_title":     req.BugTitle,
			"bug_status":    bugStatus,
			"bug_level":     bugLevel,
			"bug_source":    bugSource,
			"bug_type":      bugType,
			"bug_env":       bugEnv,
			"bug_ua":        req.BugUa,
			"project_id":    req.ProjectID,
			"bug_rich_text": req.BugRichText,
			"file_ids": func() interface{} {
				if fileIDsStr == "" {
					return nil
				}
				return fileIDsStr
			}(),
			"version_id":  req.VersionID,
			"module_id":   req.ModuleID,
			"story_id":    req.StoryID,
			"fix_user_id": assigneeID,
			"update_date": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("缺陷不存在或无权操作")
		}
		return nil
	})
}

func UpdateBugField(bugID string, key string, value interface{}, creatorID string, permission datapermission.Permission) error {
	var bug models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").First(&bug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("缺陷不存在")
		}
		return err
	}

	allowedKeys := map[string]bool{"fixUserId": true, "bugLevel": true, "bugEnv": true, "bugType": true, "bugSource": true}
	if !allowedKeys[key] {
		return fmt.Errorf("只能修改 fixUserId、bugLevel、bugEnv、bugType、bugSource 字段")
	}

	updateMap := make(map[string]interface{})
	switch key {
	case "fixUserId":
		userID, err := dataPermissionStringValue(value, "修复人ID")
		if err != nil {
			return err
		}
		validatedID, err := validateRequiredDataPermissionUser(database.DB, &userID, "修复人", permission)
		if err != nil {
			return err
		}
		updateMap["fix_user_id"] = validatedID
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

	return updateDevRecordWithHistory(creatorID, bugID, 20, 10, "", func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
		result := permission.Apply(updateQuery, "dev_bug.creator_id", "dev_bug.fix_user_id").Updates(updateMap)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("缺陷不存在或无权操作")
		}
		return nil
	})
}

func UpdateBugNext(bugID string, bugStatus string, changeRichText string, creatorID string, permission datapermission.Permission) error {
	var bug models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").First(&bug).Error
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
	// 推进到状态 30（待验证）时，写入当前登录人为验证人
	if bugStatusInt == 30 {
		updateMap["verifier_id"] = creatorID
	}

	return updateDevRecordWithHistory(creatorID, bugID, 20, 40, changeRichText, func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
		result := permission.Apply(updateQuery, "dev_bug.creator_id", "dev_bug.fix_user_id").Updates(updateMap)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("缺陷不存在或无权操作")
		}
		return nil
	})
}

func ConfirmBug(bugID string, req *models.ConfirmBugRequest, creatorID string, permission datapermission.Permission) error {
	var bug models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").First(&bug).Error
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

	return updateDevRecordWithHistory(creatorID, bugID, 20, 50, req.ChangeRichText, func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", bugID, 0)
		result := permission.Apply(updateQuery, "dev_bug.creator_id", "dev_bug.fix_user_id").Updates(updateMap)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("缺陷不存在或无权操作")
		}
		return nil
	})
}

func DeleteBugs(bugIDs []string, creatorID string, permission datapermission.Permission) error {
	var bugs []models.DevBug
	query := database.DB.Model(&models.DevBug{}).Where("bug_id IN ? AND del_flag = ?", bugIDs, 0)
	err := permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id").Find(&bugs).Error
	if err != nil {
		return err
	}
	if len(bugs) != len(uniqueNonEmptyStrings(bugIDs)) {
		return fmt.Errorf("缺陷不存在或无权操作")
	}

	for _, b := range bugs {
		if b.BugStatus != 0 {
			return fmt.Errorf("只能删除状态为待确认的缺陷")
		}
	}

	accessibleIDs := make([]string, 0, len(bugs))
	for _, bug := range bugs {
		accessibleIDs = append(accessibleIDs, bug.BugID)
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevBug{}).Where("bug_id IN ? AND del_flag = ? AND bug_status = ?", accessibleIDs, 0, 0)
		result := permission.Apply(updateQuery, "dev_bug.creator_id", "dev_bug.fix_user_id").Updates(map[string]interface{}{
			"del_flag":    1,
			"update_date": time.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(accessibleIDs)) {
			return fmt.Errorf("缺陷不存在或无权操作")
		}
		for _, bug := range bugs {
			if err := createChangeHistoryTx(tx, creatorID, bug.BugID, 20, 20, ""); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

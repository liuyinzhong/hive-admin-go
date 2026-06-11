package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func GetStorys(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevStory{}).Where("del_flag = ?", 0)

	if storyNum, ok := params["storyNum"].(int); ok && storyNum > 0 {
		db = db.Where("story_num = ?", storyNum)
	}
	if storyTitle, ok := params["storyTitle"].(string); ok && storyTitle != "" {
		db = db.Where("story_title LIKE ?", "%"+storyTitle+"%")
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
	if storyStatuses, ok := params["storyStatuses"].([]int); ok && len(storyStatuses) > 0 {
		db = db.Where("story_status IN ?", storyStatuses)
	}

	sorts := params["sorts"].(string)
	order := utils.BuildOrderBy(sorts, map[string]string{
		"storyStatus": "story_status",
		"storyLevel":  "story_level",
		"storyTitle":  "story_title",
	})
	if order == "" {
		order = "create_date DESC"
	}

	return utils.PaginateWithTransform[models.DevStory](db, page, pageSize, order, func(items []models.DevStory) interface{} {
		return buildStoryResponses(items)
	})
}

func GetAllStorys(params map[string]interface{}) ([]models.StoryResponse, error) {
	db := database.DB.Model(&models.DevStory{}).Where("del_flag = ?", 0)

	if storyNum, ok := params["storyNum"].(int); ok && storyNum > 0 {
		db = db.Where("story_num = ?", storyNum)
	}
	if storyTitle, ok := params["storyTitle"].(string); ok && storyTitle != "" {
		db = db.Where("story_title LIKE ?", "%"+storyTitle+"%")
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
	if storyStatus, ok := params["storyStatus"].(int); ok && storyStatus >= 0 {
		db = db.Where("story_status = ?", storyStatus)
	}

	var storys []models.DevStory
	err := db.Order("create_date DESC").Find(&storys).Error
	if err != nil {
		return nil, err
	}

	return buildStoryResponses(storys), nil
}

func buildStoryResponses(storys []models.DevStory) []models.StoryResponse {
	creatorIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	versionIDs := make([]string, 0)
	moduleIDs := make([]string, 0)
	allUserIDs := make([]string, 0)
	allFileIDs := make([]string, 0)

	for _, s := range storys {
		if s.CreatorID != nil {
			creatorIDs = append(creatorIDs, *s.CreatorID)
		}
		projectIDs = append(projectIDs, s.ProjectID)
		if s.VersionID != nil {
			versionIDs = append(versionIDs, *s.VersionID)
		}
		if s.ModuleID != nil {
			moduleIDs = append(moduleIDs, *s.ModuleID)
		}
		if s.UserIDs != nil {
			ids := strings.Split(*s.UserIDs, ",")
			for _, id := range ids {
				if id != "" {
					allUserIDs = append(allUserIDs, id)
				}
			}
		}
		if s.FileIDs != nil {
			ids := strings.Split(*s.FileIDs, ",")
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

	userMap := make(map[string]models.StoryUserItem)
	if len(allUserIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", allUserIDs).Find(&users)
		for _, u := range users {
			userMap[u.UserID] = models.StoryUserItem{
				UserID:   &u.UserID,
				Avatar:   u.Avatar,
				RealName: u.RealName,
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
			var creators []models.SysUser
			database.DB.Where("user_id IN ?", fileCreatorIDs).Find(&creators)
			for _, c := range creators {
				if c.RealName != nil {
					fileCreatorNameMap[c.UserID] = *c.RealName
				}
			}
		}

		for _, f := range files {
			var creatorName string
			if f.CreatorID != nil {
				creatorName = fileCreatorNameMap[*f.CreatorID]
			}
			fileMap[f.FileID] = models.FileResponse{
				FileID:        &f.FileID,
				URL:           f.URL,
				Name:          f.Name,
				Type:          f.Type,
				Size:          f.Size,
				FileExt:       f.FileExt,
				OriginalName:  f.OriginalName,
				Path:          f.Path,
				FullPath:      f.FullPath,
				ThumbnailPath: f.ThumbnailPath,
				ThumbnailURL:  f.ThumbnailURL,
				CreatorID:     f.CreatorID,
				CreatorName:   &creatorName,
				CreateDate:    models.TimeToStringPtr(f.CreateDate),
			}
		}
	}

	var responses []models.StoryResponse
	for _, story := range storys {
		creatorName := creators[utils.StringValue(story.CreatorID)]
		projectTitle := projects[story.ProjectID]
		version := versions[utils.StringValue(story.VersionID)]
		moduleTitle := modules[utils.StringValue(story.ModuleID)]

		userIDs := make([]string, 0)
		userList := make([]models.StoryUserItem, 0)
		if story.UserIDs != nil {
			ids := strings.Split(*story.UserIDs, ",")
			for _, id := range ids {
				if id != "" {
					userIDs = append(userIDs, id)
					if u, ok := userMap[id]; ok {
						userList = append(userList, u)
					}
				}
			}
		}

		fileIDs := make([]string, 0)
		fileList := make([]models.FileResponse, 0)
		if story.FileIDs != nil {
			ids := strings.Split(*story.FileIDs, ",")
			for _, id := range ids {
				if id != "" {
					fileIDs = append(fileIDs, id)
					if f, ok := fileMap[id]; ok {
						fileList = append(fileList, f)
					}
				}
			}
		}

		responses = append(responses, models.StoryResponse{
			StoryID:       &story.StoryID,
			StoryTitle:    story.StoryTitle,
			StoryNum:      story.StoryNum,
			CreatorName:   &creatorName,
			CreatorID:     story.CreatorID,
			StoryType:     intToString(story.StoryType),
			StoryStatus:   intToString(story.StoryStatus),
			StoryLevel:    intToString(story.StoryLevel),
			VersionID:     story.VersionID,
			Version:       &version,
			ProjectID:     &story.ProjectID,
			ProjectTitle:  &projectTitle,
			ModuleID:      story.ModuleID,
			ModuleTitle:   &moduleTitle,
			Source:        intToString(story.Source),
			UpdateDate:    models.TimeToStringPtr(story.UpdateDate),
			CreateDate:    models.TimeToStringPtr(story.CreateDate),
			UserList:      userList,
			UserIDs:       userIDs,
			StoryRichText: story.StoryRichText,
			FileIDs:       fileIDs,
			FileList:      fileList,
		})
	}
	return responses
}

func GetStoryByNum(storyNum int) (*models.StoryResponse, error) {
	var story models.DevStory
	err := database.DB.Where("story_num = ? AND del_flag = ?", storyNum, 0).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("需求不存在")
		}
		return nil, err
	}

	return buildSingleStoryResponse(&story), nil
}

func buildSingleStoryResponse(story *models.DevStory) *models.StoryResponse {
	var creatorName string
	if story.CreatorID != nil {
		var user models.SysUser
		database.DB.Where("user_id = ?", *story.CreatorID).First(&user)
		if user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	var projectTitle string
	var project models.DevProject
	database.DB.Where("project_id = ?", story.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	var version string
	if story.VersionID != nil {
		var v models.DevVersion
		database.DB.Where("version_id = ?", *story.VersionID).First(&v)
		if v.Version != nil {
			version = *v.Version
		}
	}

	var moduleTitle string
	if story.ModuleID != nil {
		var m models.DevModule
		database.DB.Where("module_id = ?", *story.ModuleID).First(&m)
		if m.ModuleTitle != nil {
			moduleTitle = *m.ModuleTitle
		}
	}

	userIDs := make([]string, 0)
	userList := make([]models.StoryUserItem, 0)
	if story.UserIDs != nil {
		ids := strings.Split(*story.UserIDs, ",")
		for _, id := range ids {
			if id != "" {
				userIDs = append(userIDs, id)
				var u models.SysUser
				database.DB.Where("user_id = ?", id).First(&u)
				userList = append(userList, models.StoryUserItem{
					UserID:   &u.UserID,
					Avatar:   u.Avatar,
					RealName: u.RealName,
				})
			}
		}
	}

	fileIDs := make([]string, 0)
	fileList := make([]models.FileResponse, 0)
	if story.FileIDs != nil {
		ids := strings.Split(*story.FileIDs, ",")
		for _, id := range ids {
			if id != "" {
				fileIDs = append(fileIDs, id)
				var f models.SysFile
				database.DB.Where("file_id = ?", id).First(&f)
				var creatorName string
				if f.CreatorID != nil {
					var u models.SysUser
					database.DB.Where("user_id = ?", *f.CreatorID).First(&u)
					if u.RealName != nil {
						creatorName = *u.RealName
					}
				}
				fileList = append(fileList, models.FileResponse{
					FileID:        &f.FileID,
					URL:           f.URL,
					Name:          f.Name,
					Type:          f.Type,
					Size:          f.Size,
					FileExt:       f.FileExt,
					OriginalName:  f.OriginalName,
					Path:          f.Path,
					FullPath:      f.FullPath,
					ThumbnailPath: f.ThumbnailPath,
					ThumbnailURL:  f.ThumbnailURL,
					CreatorID:     f.CreatorID,
					CreatorName:   &creatorName,
					CreateDate:    models.TimeToStringPtr(f.CreateDate),
				})
			}
		}
	}

	taskList := make([]models.TaskResponse, 0)
	bugList := make([]models.BugResponse, 0)

	tasks, err := GetTasksByStoryId(story.StoryID)
	if err == nil && tasks != nil {
		taskList = tasks
	}

	bugs, err := GetBugsByStoryId(story.StoryID)
	if err == nil && bugs != nil {
		bugList = bugs
	}

	nodes := make([]models.NodeResponse, 0)
	var nodeList []models.DevNode
	database.DB.Where("business_id = ?", story.StoryID).Order("sort ASC").Find(&nodeList)
	nodeUserIDs := make([]string, 0)
	nodeUserMap := make(map[string]string)
	for _, n := range nodeList {
		if n.UserID != "" {
			nodeUserIDs = append(nodeUserIDs, n.UserID)
		}
	}
	if len(nodeUserIDs) > 0 {
		var nodeUsers []models.SysUser
		database.DB.Where("user_id IN ?", nodeUserIDs).Find(&nodeUsers)
		for _, u := range nodeUsers {
			if u.RealName != nil {
				nodeUserMap[u.UserID] = *u.RealName
			}
		}
	}
	for _, n := range nodeList {
		nodes = append(nodes, *models.DevNodeToNodeResponse(n, nodeUserMap[n.UserID]))
	}

	return &models.StoryResponse{
		StoryID:       &story.StoryID,
		StoryTitle:    story.StoryTitle,
		StoryNum:      story.StoryNum,
		CreatorName:   &creatorName,
		CreatorID:     story.CreatorID,
		StoryType:     intToString(story.StoryType),
		StoryStatus:   intToString(story.StoryStatus),
		StoryLevel:    intToString(story.StoryLevel),
		VersionID:     story.VersionID,
		Version:       &version,
		ProjectID:     &story.ProjectID,
		ProjectTitle:  &projectTitle,
		ModuleID:      story.ModuleID,
		ModuleTitle:   &moduleTitle,
		Source:        intToString(story.Source),
		UpdateDate:    models.TimeToStringPtr(story.UpdateDate),
		CreateDate:    models.TimeToStringPtr(story.CreateDate),
		UserList:      userList,
		UserIDs:       userIDs,
		StoryRichText: story.StoryRichText,
		FileIDs:       fileIDs,
		FileList:      fileList,
		TaskList:      taskList,
		BugList:       bugList,
		Nodes:         nodes,
	}
}

func CreateStory(req *models.CreateStoryRequest, creatorID string) error {
	storyID := uuid.New().String()
	now := time.Now()

	if len(req.Nodes) > 0 {
		hasHandleOrApprove := false
		for _, n := range req.Nodes {
			if n.NodeType == 1 || n.NodeType == 2 {
				hasHandleOrApprove = true
				if strings.TrimSpace(n.UserID) == "" {
					return fmt.Errorf("办理节点和审批节点的负责人不能为空")
				}
			}
		}
		if !hasHandleOrApprove {
			return fmt.Errorf("不能仅包含开始节点和结束节点")
		}
	}

	userIDsStr := ""
	if len(req.UserIDs) > 0 {
		userIDsStr = strings.Join(req.UserIDs, ",")
	}

	fileIDsStr := ""
	if len(req.FileIDs) > 0 {
		fileIDsStr = strings.Join(req.FileIDs, ",")
	}

	storyStatus, err := parseStringInt(req.StoryStatus, "storyStatus")
	if err != nil {
		return err
	}
	storyType, err := parseStringInt(req.StoryType, "storyType")
	if err != nil {
		return err
	}
	storyLevel, err := parseStringInt(req.StoryLevel, "storyLevel")
	if err != nil {
		return err
	}
	source, err := parseStringInt(req.Source, "source")
	if err != nil {
		return err
	}

	var userIDsPtr *string
	if userIDsStr != "" {
		userIDsPtr = &userIDsStr
	}

	var fileIDsPtr *string
	if fileIDsStr != "" {
		fileIDsPtr = &fileIDsStr
	}

	businessType := req.BusinessType
	if businessType == "" {
		businessType = "0"
	}

	story := models.DevStory{
		StoryID:       storyID,
		StoryTitle:    req.StoryTitle,
		StoryType:     storyType,
		StoryStatus:   storyStatus,
		StoryLevel:    storyLevel,
		Source:        source,
		ProjectID:     req.ProjectID,
		VersionID:     req.VersionID,
		ModuleID:      req.ModuleID,
		CreatorID:     &creatorID,
		StoryRichText: req.StoryRichText,
		UserIDs:       userIDsPtr,
		FileIDs:       fileIDsPtr,
		CreateDate:    &now,
		UpdateDate:    &now,
		DelFlag:       0,
		StoryNum:      0,
	}

	err = database.DB.Create(&story).Error
	if err != nil {
		return err
	}

	if len(req.Nodes) > 0 {
		for i, n := range req.Nodes {
			node := models.DevNode{
				NodeID:       uuid.New().String(),
				Label:        n.Label,
				Value:        n.Value,
				Sort:         n.Sort,
				UserID:       n.UserID,
				Current:      0,
				NodeType:     n.NodeType,
				Result:       0,
				Remark:       n.Remark,
				BusinessType: &businessType,
				BusinessID:   storyID,
				CreateDate:   &now,
			}
			if i == 0 {
				node.Current = 1
				node.StartDate = &now
			}
			if err := database.DB.Create(&node).Error; err != nil {
				return err
			}
		}
	}

	return createChangeHistory(creatorID, storyID, 0, 0, "")
}

func CreateStorys(reqs []models.CreateStoryRequest, creatorID string) error {
	for _, req := range reqs {
		err := CreateStory(&req, creatorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateStory(storyID string, req *models.UpdateStoryRequest, creatorID string) error {
	var story models.DevStory
	err := database.DB.Where("story_id = ? AND del_flag = ?", storyID, 0).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("需求不存在")
		}
		return err
	}

	userIDsStr := ""
	if len(req.UserIDs) > 0 {
		userIDsStr = strings.Join(req.UserIDs, ",")
	}

	fileIDsStr := ""
	if len(req.FileIDs) > 0 {
		fileIDsStr = strings.Join(req.FileIDs, ",")
	}

	storyStatus, err := parseStringInt(req.StoryStatus, "storyStatus")
	if err != nil {
		return err
	}
	storyType, err := parseStringInt(req.StoryType, "storyType")
	if err != nil {
		return err
	}
	storyLevel, err := parseStringInt(req.StoryLevel, "storyLevel")
	if err != nil {
		return err
	}
	source, err := parseStringInt(req.Source, "source")
	if err != nil {
		return err
	}

	now := time.Now()
	err = database.DB.Model(&story).Updates(map[string]interface{}{
		"story_title":     req.StoryTitle,
		"story_type":      storyType,
		"story_status":    storyStatus,
		"story_level":     storyLevel,
		"source":          source,
		"project_id":      req.ProjectID,
		"version_id":      req.VersionID,
		"module_id":       req.ModuleID,
		"story_rich_text": req.StoryRichText,
		"user_ids": func() interface{} {
			if userIDsStr == "" {
				return nil
			}
			return userIDsStr
		}(),
		"file_ids": func() interface{} {
			if fileIDsStr == "" {
				return nil
			}
			return fileIDsStr
		}(),
		"update_date": now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, storyID, 0, 10, "")
}

func UpdateStoryField(storyID string, key string, value interface{}, creatorID string) error {
	var story models.DevStory
	err := database.DB.Where("story_id = ? AND del_flag = ?", storyID, 0).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("需求不存在")
		}
		return err
	}

	allowedKeys := map[string]bool{"userIds": true, "storyType": true, "storyLevel": true, "source": true}
	if !allowedKeys[key] {
		return fmt.Errorf("只能修改 userIds、storyType、storyLevel、source 字段")
	}

	updateMap := make(map[string]interface{})
	switch key {
	case "userIds":
		if ids, ok := value.([]interface{}); ok {
			strIDs := make([]string, 0)
			for _, id := range ids {
				strIDs = append(strIDs, fmt.Sprintf("%v", id))
			}
			updateMap["user_ids"] = strings.Join(strIDs, ",")
		}
	case "storyType":
		updateMap["story_type"] = value
	case "storyLevel":
		updateMap["story_level"] = value
	case "source":
		updateMap["source"] = value
	}

	updateMap["update_date"] = time.Now()

	err = database.DB.Model(&story).Updates(updateMap).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, storyID, 0, 10, "")
}

func UpdateStoryNext(storyID string, storyStatus string, changeRichText string, creatorID string) error {
	var story models.DevStory
	err := database.DB.Where("story_id = ? AND del_flag = ?", storyID, 0).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("需求不存在")
		}
		return err
	}

	storyStatusInt, err := parseStringInt(storyStatus, "storyStatus")
	if err != nil {
		return err
	}

	now := time.Now()
	err = database.DB.Model(&story).Updates(map[string]interface{}{
		"story_status": storyStatusInt,
		"update_date":  now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, storyID, 0, 40, changeRichText)
}

func DeleteStorys(storyIDs []string, creatorID string) error {
	var storys []models.DevStory
	err := database.DB.Where("story_id IN ? AND del_flag = ?", storyIDs, 0).Find(&storys).Error
	if err != nil {
		return err
	}

	for _, s := range storys {
		if s.StoryStatus != 0 {
			return fmt.Errorf("只能删除状态为待评审的需求")
		}
	}

	err = database.DB.Model(&models.DevStory{}).Where("story_id IN ?", storyIDs).Updates(map[string]interface{}{
		"del_flag":    1,
		"update_date": time.Now(),
	}).Error
	if err != nil {
		return err
	}

	for _, s := range storys {
		createChangeHistory(creatorID, s.StoryID, 0, 20, "")
	}
	return nil
}

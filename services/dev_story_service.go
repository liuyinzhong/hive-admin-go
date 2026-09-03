package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// storyListMenuName 需求管理菜单的路由名称,流转通知的菜单未读消息推送目标。
const storyListMenuName = "devStory"

// storyStatusClosed 需求已关闭状态值(STORY_STATUS 字典),流转到该状态无需指定状态负责人。
const storyStatusClosed = 99

// applyStoryPermission 应用需求模块数据权限:创建人或 dev_story_user 参与人可见。
func applyStoryPermission(query *gorm.DB, permission datapermission.Permission) *gorm.DB {
	return permission.ApplyWithUserTable(query, []string{"dev_story.creator_id"}, "dev_story.story_id", "dev_story_user", "story_id", "user_id")
}

func GetStorys(page, pageSize int, params map[string]interface{}, permission datapermission.Permission) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevStory{}).Where("dev_story.del_flag = ?", 0)
	db = applyStoryPermission(db, permission)

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
	// 按当前负责人筛选:用户视角查询名下需求,即参与人中负责需求当前状态的用户
	if thisUserID, ok := params["thisUserId"].(string); ok && thisUserID != "" {
		db = db.Where("EXISTS (SELECT 1 FROM dev_story_user dsu WHERE dsu.story_id = dev_story.story_id AND dsu.story_status = dev_story.story_status AND dsu.user_id = ?)", thisUserID)
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
		return buildStoryResponses(items, permission)
	})
}

func GetAllStorys(params map[string]interface{}, permission datapermission.Permission) ([]models.StoryResponse, error) {
	db := database.DB.Model(&models.DevStory{}).Where("dev_story.del_flag = ?", 0)
	db = applyStoryPermission(db, permission)

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

	return buildStoryResponses(storys, permission), nil
}

func buildStoryResponses(storys []models.DevStory, permission datapermission.Permission) []models.StoryResponse {
	creatorIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	versionIDs := make([]string, 0)
	moduleIDs := make([]string, 0)
	allFileIDs := make([]string, 0)
	storyIDs := make([]string, 0, len(storys))

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
		storyIDs = append(storyIDs, s.StoryID)
		if s.FileIDs != nil {
			ids := strings.Split(*s.FileIDs, ",")
			for _, id := range ids {
				if id != "" {
					allFileIDs = append(allFileIDs, id)
				}
			}
		}
	}

	storyUserMap := loadStoryUserMap(storyIDs)

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

	userMap := make(map[string]models.SysUser)
	allUserIDs := make([]string, 0)
	for _, rows := range storyUserMap {
		for _, row := range rows {
			allUserIDs = append(allUserIDs, row.UserID)
		}
	}
	if len(allUserIDs) > 0 {
		var users []models.SysUser
		database.DB.Where("user_id IN ?", uniqueNonEmptyStrings(allUserIDs)).Find(&users)
		for _, u := range users {
			userMap[u.UserID] = u
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

		userList := make([]models.StoryUserItem, 0)
		for _, su := range storyUserMap[story.StoryID] {
			if u, ok := userMap[su.UserID]; ok {
				item := models.StoryUserItem{
					UserID:   &u.UserID,
					Avatar:   u.Avatar,
					RealName: u.RealName,
				}
				if su.StoryStatus != nil {
					status := intToString(*su.StoryStatus)
					item.StoryStatus = &status
				}
				userList = append(userList, item)
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

		// 当前状态负责人:参与人中 story_status 等于需求当前状态的用户
		thisUser := buildStoryThisUser(storyUserMap[story.StoryID], userMap, story.StoryStatus)

		responses = append(responses, models.StoryResponse{
			StoryID:       &story.StoryID,
			StoryTitle:    story.StoryTitle,
			StoryNum:      story.StoryNum,
			CreatorName:   &creatorName,
			CreatorID:     story.CreatorID,
			StoryType:     intToString(story.StoryType),
			StoryStatus:   intToString(story.StoryStatus),
			ThisUser:      thisUser,
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
			StoryRichText: story.StoryRichText,
			FileIDs:       fileIDs,
			FileList:      fileList,
		})
	}

	return responses
}

// loadStoryUserMap 批量查询需求的参与人关联行,按 story_id 分组并保持 create_date 顺序。
func loadStoryUserMap(storyIDs []string) map[string][]models.DevStoryUser {
	result := make(map[string][]models.DevStoryUser)
	ids := uniqueNonEmptyStrings(storyIDs)
	if len(ids) == 0 {
		return result
	}
	var rows []models.DevStoryUser
	if err := database.DB.Where("story_id IN ?", ids).Order("create_date, id").Find(&rows).Error; err != nil {
		log.Printf("[story] 查询需求参与人失败: err=%v", err)
		return result
	}
	for _, row := range rows {
		result[row.StoryID] = append(result[row.StoryID], row)
	}
	return result
}

// buildStoryThisUser 从参与人关联行中筛选负责指定状态的用户,返回当前状态负责人;
// 同一需求同一状态仅保留最新指定的负责人,异常多行时取最早写入的一条。
func buildStoryThisUser(rows []models.DevStoryUser, userMap map[string]models.SysUser, storyStatus int) *models.StoryUserItem {
	for _, row := range rows {
		if row.StoryStatus == nil || *row.StoryStatus != storyStatus {
			continue
		}
		if u, ok := userMap[row.UserID]; ok {
			item := models.StoryUserItem{
				UserID:   &u.UserID,
				Avatar:   u.Avatar,
				RealName: u.RealName,
			}
			status := intToString(storyStatus)
			item.StoryStatus = &status
			return &item
		}
	}
	return nil
}

func GetStoryByNum(storyNum int, permission datapermission.Permission) (*models.StoryResponse, error) {
	var story models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_num = ? AND del_flag = ?", storyNum, 0)
	err := applyStoryPermission(query, permission).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("需求不存在")
		}
		return nil, err
	}

	return buildSingleStoryResponse(&story, permission), nil
}

func buildSingleStoryResponse(story *models.DevStory, permission datapermission.Permission) *models.StoryResponse {
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

	userList := make([]models.StoryUserItem, 0)
	storyUserRows := loadStoryUserMap([]string{story.StoryID})[story.StoryID]
	userMap := make(map[string]models.SysUser)
	if len(storyUserRows) > 0 {
		userIDs := make([]string, 0, len(storyUserRows))
		for _, row := range storyUserRows {
			userIDs = append(userIDs, row.UserID)
		}
		var users []models.SysUser
		database.DB.Where("user_id IN ?", userIDs).Find(&users)
		for _, u := range users {
			userMap[u.UserID] = u
		}
		for _, row := range storyUserRows {
			u, ok := userMap[row.UserID]
			if !ok {
				continue
			}
			item := models.StoryUserItem{
				UserID:   &u.UserID,
				Avatar:   u.Avatar,
				RealName: u.RealName,
			}
			if row.StoryStatus != nil {
				status := intToString(*row.StoryStatus)
				item.StoryStatus = &status
			}
			userList = append(userList, item)
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

	tasks, err := GetTasksByStoryId(story.StoryID, permission)
	if err == nil && tasks != nil {
		taskList = tasks
	}

	bugs, err := GetBugsByStoryId(story.StoryID, permission)
	if err == nil && bugs != nil {
		bugList = bugs
	}

	return &models.StoryResponse{
		StoryID:       &story.StoryID,
		StoryTitle:    story.StoryTitle,
		StoryNum:      story.StoryNum,
		CreatorName:   &creatorName,
		CreatorID:     story.CreatorID,
		StoryType:     intToString(story.StoryType),
		StoryStatus:   intToString(story.StoryStatus),
		ThisUser:      buildStoryThisUser(storyUserRows, userMap, story.StoryStatus),
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
		StoryRichText: story.StoryRichText,
		FileIDs:       fileIDs,
		FileList:      fileList,
		TaskList:      taskList,
		BugList:       bugList,
	}
}

func CreateStory(req *models.CreateStoryRequest, creatorID string, permission datapermission.Permission) error {
	var storyID string
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		return createStoryTx(tx, req, creatorID, permission, &storyID)
	})
	if err != nil {
		return err
	}
	// 需求创建成功后,自动匹配已发布的 story 类型流程定义并发起流程,实现需求与流程的联动。
	autoStartStoryWorkflow(storyID, creatorID)
	return nil
}

// autoStartStoryWorkflow 需求创建后自动匹配并发起流程。
// 查找 business_type=story 且已发布的最新流程定义,找到则发起流程绑定到该需求;
// 未找到已发布流程定义时不报错,需求保持已创建状态(不绑流程);
// 发起失败时记日志不抛错,需求创建不受影响,可后续手动发起。
func autoStartStoryWorkflow(storyID, creatorID string) {
	var definition models.WfProcessDefinition
	err := database.DB.Where("business_type = ? AND status = ? AND del_flag = ?", "story", 1, 0).
		Order("update_date DESC").First(&definition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return // 未配置已发布的 story 流程定义,不绑流程
		}
		log.Printf("[workflow] 自动匹配需求流程失败: storyID=%s, err=%v", storyID, err)
		return
	}
	businessID := storyID
	req := &models.StartWorkflowInstanceRequest{
		DefinitionID: definition.DefinitionID,
		BusinessID:   &businessID,
	}
	if _, err := StartWorkflowInstance(req, creatorID); err != nil {
		log.Printf("[workflow] 自动发起需求流程失败: storyID=%s, definitionID=%s, err=%v",
			storyID, definition.DefinitionID, err)
	}
}

func createStoryTx(tx *gorm.DB, req *models.CreateStoryRequest, creatorID string, permission datapermission.Permission, outStoryID *string) error {
	storyID := uuid.New().String()
	if outStoryID != nil {
		*outStoryID = storyID
	}
	now := time.Now()

	fileIDs, err := validateDataPermissionFiles(tx, req.FileIDs, permission)
	if err != nil {
		return err
	}
	if err := validateDevVersionReference(tx, req.VersionID, permission); err != nil {
		return err
	}

	fileIDsStr := ""
	if len(fileIDs) > 0 {
		fileIDsStr = strings.Join(fileIDs, ",")
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

	var fileIDsPtr *string
	if fileIDsStr != "" {
		fileIDsPtr = &fileIDsStr
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
		FileIDs:       fileIDsPtr,
		CreateDate:    &now,
		UpdateDate:    &now,
		DelFlag:       0,
		StoryNum:      0,
	}

	err = tx.Create(&story).Error
	if err != nil {
		return err
	}

	return createChangeHistoryTx(tx, creatorID, storyID, 0, 0, "")
}

func CreateStorys(reqs []models.CreateStoryRequest, creatorID string, permission datapermission.Permission) error {
	storyIDs := make([]string, 0, len(reqs))
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for index := range reqs {
			var storyID string
			if err := createStoryTx(tx, &reqs[index], creatorID, permission, &storyID); err != nil {
				return fmt.Errorf("第%d条需求：%w", index+1, err)
			}
			storyIDs = append(storyIDs, storyID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 与单条创建保持一致:创建成功后自动匹配并发起流程,失败仅记日志不影响创建结果
	for _, storyID := range storyIDs {
		autoStartStoryWorkflow(storyID, creatorID)
	}
	return nil
}

// StartStoryWorkflow 为需求发起流程,自动把 businessId 绑定为 storyId。
// 前端在创建需求后调用,传入已发布的流程定义ID和流程变量。
// 流程定义声明的 BusinessType 必须为 "story",否则 StartWorkflowInstance 会拒绝绑定,
// 这样防止需求误绑定到 bug/task 等其它业务流程。返回创建好的流程实例。
func StartStoryWorkflow(storyID, definitionID, creatorID string, variables map[string]interface{}) (*models.WorkflowInstanceResponse, error) {
	var story models.DevStory
	if err := database.DB.Where("story_id = ? AND del_flag = 0", storyID).First(&story).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("需求不存在")
		}
		return nil, err
	}
	businessID := storyID
	req := &models.StartWorkflowInstanceRequest{
		DefinitionID: definitionID,
		BusinessID:   &businessID,
		Variables:    variables,
	}
	return StartWorkflowInstance(req, creatorID)
}

// GetStoryWorkflowBindings 查询需求关联的全部流程实例,用于需求详情页展示关联流程列表。
// 包含自动发起的需求流程和结束后动作落地创建的来源审批实例,按绑定时间倒序;
// 需求未关联流程时返回空列表,调用方据此决定是否显示"发起流程"按钮。
func GetStoryWorkflowBindings(storyID string) ([]WorkflowBusinessInstanceResponse, error) {
	return GetWorkflowBusinessInstanceList("story", storyID)
}

func UpdateStory(storyID string, req *models.UpdateStoryRequest, creatorID string, permission datapermission.Permission) error {
	var story models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", storyID, 0)
	err := applyStoryPermission(query, permission).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("需求不存在")
		}
		return err
	}

	fileIDs, err := validateDataPermissionFiles(database.DB, req.FileIDs, permission)
	if err != nil {
		return err
	}
	if err := validateDevVersionReference(database.DB, req.VersionID, permission); err != nil {
		return err
	}

	fileIDsStr := ""
	if len(fileIDs) > 0 {
		fileIDsStr = strings.Join(fileIDs, ",")
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
	return updateDevRecordWithHistory(creatorID, storyID, 0, 10, "", func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", storyID, 0)
		result := applyStoryPermission(updateQuery, permission).Updates(map[string]interface{}{
			"story_title":     req.StoryTitle,
			"story_type":      storyType,
			"story_status":    storyStatus,
			"story_level":     storyLevel,
			"source":          source,
			"project_id":      req.ProjectID,
			"version_id":      req.VersionID,
			"module_id":       req.ModuleID,
			"story_rich_text": req.StoryRichText,
			"file_ids": func() interface{} {
				if fileIDsStr == "" {
					return nil
				}
				return fileIDsStr
			}(),
			"update_date": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("需求不存在或无权操作")
		}
		return nil
	})
}

func UpdateStoryField(storyID string, key string, value interface{}, creatorID string, permission datapermission.Permission) error {
	var story models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", storyID, 0)
	err := applyStoryPermission(query, permission).First(&story).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("需求不存在")
		}
		return err
	}

	allowedKeys := map[string]bool{"storyType": true, "storyLevel": true, "source": true}
	if !allowedKeys[key] {
		return fmt.Errorf("只能修改 storyType、storyLevel、source 字段")
	}

	updateMap := make(map[string]interface{})
	switch key {
	case "storyType":
		updateMap["story_type"] = value
	case "storyLevel":
		updateMap["story_level"] = value
	case "source":
		updateMap["source"] = value
	}

	updateMap["update_date"] = time.Now()

	return updateDevRecordWithHistory(creatorID, storyID, 0, 10, "", func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", storyID, 0)
		result := applyStoryPermission(updateQuery, permission).Updates(updateMap)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("需求不存在或无权操作")
		}
		return nil
	})
}

func UpdateStoryNext(storyID string, storyStatus string, userID string, changeRichText string, creatorID string, permission datapermission.Permission) error {
	var story models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", storyID, 0)
	err := applyStoryPermission(query, permission).First(&story).Error
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

	// 流转到已关闭状态不指定负责人,其余状态必须指定项目成员作为目标状态负责人
	hasOwner := storyStatusInt != storyStatusClosed
	if hasOwner {
		if userID == "" {
			return fmt.Errorf("请选择状态负责人")
		}
		if err := validateStoryProjectUser(story.ProjectID, userID); err != nil {
			return err
		}
	}

	// 流转说明为空时自动生成默认内容,保证变更记录始终记录流转去向
	changeRichText = strings.TrimSpace(changeRichText)
	if changeRichText == "" {
		changeRichText = fmt.Sprintf("流转至「%s」，请及时跟进", loadStoryStatusLabel(storyStatusInt))
	}

	now := time.Now()
	err = updateDevRecordWithHistory(creatorID, storyID, 0, 40, changeRichText, func(tx *gorm.DB) error {
		return applyStoryNextTx(tx, &story, storyStatusInt, userID, hasOwner, now, permission)
	})
	if err != nil {
		return err
	}

	if hasOwner {
		notifyStoryStatusOwner(&story, storyStatusInt, userID)
	}
	return nil
}

// applyStoryNextTx 在事务内执行单条需求流转:更新需求状态并写入状态负责人行。
// 单条流转与批量流转共用,调用方负责前置校验和变更记录。
func applyStoryNextTx(tx *gorm.DB, story *models.DevStory, storyStatusInt int, userID string, hasOwner bool, now time.Time, permission datapermission.Permission) error {
	updateQuery := tx.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", story.StoryID, 0)
	result := applyStoryPermission(updateQuery, permission).Updates(map[string]interface{}{
		"story_status": storyStatusInt,
		"update_date":  now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("需求不存在或无权操作")
	}
	if !hasOwner {
		return nil
	}
	return saveStoryNextUserTx(tx, story.StoryID, userID, storyStatusInt)
}

// BatchUpdateStoryNext 批量流转需求:统一目标状态、同一位状态负责人,整批原子,
// 任一条失败整体回滚;包含已关闭需求或跨项目需求时整批拒绝。
func BatchUpdateStoryNext(req *models.BatchUpdateStoryNextRequest, creatorID string, permission datapermission.Permission) error {
	storyIDs := uniqueNonEmptyStrings(req.StoryIDs)
	if len(storyIDs) == 0 {
		return fmt.Errorf("请选择要流转的需求")
	}
	storyStatusInt, err := parseStringInt(req.StoryStatus, "storyStatus")
	if err != nil {
		return err
	}
	hasOwner := storyStatusInt != storyStatusClosed

	// 前置校验:按数据权限查出全部需求,数量不符说明存在越界或已删除的记录
	var storys []models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_id IN ? AND del_flag = ?", storyIDs, 0)
	if err := applyStoryPermission(query, permission).Find(&storys).Error; err != nil {
		return err
	}
	if len(storys) != len(storyIDs) {
		return fmt.Errorf("需求不存在或无权操作")
	}

	projectIDs := make(map[string]struct{}, 1)
	for _, story := range storys {
		if story.StoryStatus == storyStatusClosed {
			return fmt.Errorf("需求「%s」已关闭,不能再流转", utils.StringValue(story.StoryTitle))
		}
		projectIDs[story.ProjectID] = struct{}{}
	}
	if len(projectIDs) > 1 {
		return fmt.Errorf("批量流转的需求必须属于同一个项目")
	}
	if hasOwner {
		if req.UserID == "" {
			return fmt.Errorf("请选择状态负责人")
		}
		for projectID := range projectIDs {
			if err := validateStoryProjectUser(projectID, req.UserID); err != nil {
				return err
			}
		}
	}

	// 流转说明为空时自动生成默认内容,保证变更记录始终记录流转去向
	changeRichText := strings.TrimSpace(req.ChangeRichText)
	if changeRichText == "" {
		changeRichText = fmt.Sprintf("批量流转至「%s」，请及时跟进", loadStoryStatusLabel(storyStatusInt))
	}

	now := time.Now()
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for i := range storys {
			if err := applyStoryNextTx(tx, &storys[i], storyStatusInt, req.UserID, hasOwner, now, permission); err != nil {
				return fmt.Errorf("第%d条需求：%w", i+1, err)
			}
			if err := createChangeHistoryTx(tx, creatorID, storys[i].StoryID, 0, 40, changeRichText); err != nil {
				return fmt.Errorf("第%d条需求：%w", i+1, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if hasOwner {
		for i := range storys {
			notifyStoryStatusOwner(&storys[i], storyStatusInt, req.UserID)
		}
	}
	return nil
}

// validateStoryProjectUser 校验推进指定的状态负责人是需求所属项目的成员。
func validateStoryProjectUser(projectID, userID string) error {
	var count int64
	if err := database.DB.Model(&models.DevProjectUser{}).
		Where("project_id = ? AND user_id = ? AND del_flag = ?", projectID, userID, 0).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("状态负责人必须是该项目成员")
	}
	return nil
}

// saveStoryNextUserTx 保存推进指定的状态负责人,写入需求参与人关联表;
// 同一需求同一状态只保留最新指定的负责人,先删后插。
func saveStoryNextUserTx(tx *gorm.DB, storyID, userID string, storyStatus int) error {
	if err := tx.Where("story_id = ? AND story_status = ?", storyID, storyStatus).
		Delete(&models.DevStoryUser{}).Error; err != nil {
		return err
	}
	now := time.Now()
	return tx.Create(&models.DevStoryUser{
		ID:          uuid.New().String(),
		StoryID:     storyID,
		UserID:      userID,
		StoryStatus: &storyStatus,
		CreateDate:  &now,
		UpdateDate:  &now,
	}).Error
}

// notifyStoryStatusOwner 需求流转成功后,向本次指定的状态负责人推送需求管理菜单未读消息。
// 推送在事务提交后执行,失败仅记日志不影响流转结果。
func notifyStoryStatusOwner(story *models.DevStory, storyStatus int, userID string) {
	statusLabel := loadStoryStatusLabel(storyStatus)
	title := "需求流转通知"
	content := fmt.Sprintf("需求「%s」已流转至「%s」，请及时跟进", utils.StringValue(story.StoryTitle), statusLabel)

	service := NewMenuMessageService()
	if err := service.CreateMenuMessageForMenuName(userID, storyListMenuName, title, content); err != nil {
		log.Printf("[story] 流转菜单消息推送失败: storyID=%s, status=%d, userID=%s, err=%v",
			story.StoryID, storyStatus, userID, err)
	}
}

// loadStoryStatusLabel 查询 STORY_STATUS 字典中指定状态值的标签,未配置时回退为状态数字。
func loadStoryStatusLabel(storyStatus int) string {
	var dict models.SysDict
	value := strconv.Itoa(storyStatus)
	if err := database.DB.Where("type = ? AND value = ? AND status = 1 AND del_flag = 0", "STORY_STATUS", value).
		First(&dict).Error; err == nil && dict.Label != nil {
		return *dict.Label
	}
	return value
}

func DeleteStorys(storyIDs []string, creatorID string, permission datapermission.Permission) error {
	var storys []models.DevStory
	query := database.DB.Model(&models.DevStory{}).Where("story_id IN ? AND del_flag = ?", storyIDs, 0)
	err := applyStoryPermission(query, permission).Find(&storys).Error
	if err != nil {
		return err
	}
	if len(storys) != len(uniqueNonEmptyStrings(storyIDs)) {
		return fmt.Errorf("需求不存在或无权操作")
	}

	for _, s := range storys {
		if s.StoryStatus != 0 {
			return fmt.Errorf("只能删除状态为待评审的需求")
		}
	}

	accessibleIDs := make([]string, 0, len(storys))
	for _, story := range storys {
		accessibleIDs = append(accessibleIDs, story.StoryID)
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		updateQuery := tx.Model(&models.DevStory{}).Where("story_id IN ? AND del_flag = ? AND story_status = ?", accessibleIDs, 0, 0)
		result := applyStoryPermission(updateQuery, permission).Updates(map[string]interface{}{
			"del_flag":    1,
			"update_date": time.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(accessibleIDs)) {
			return fmt.Errorf("需求不存在或无权操作")
		}
		if err := tx.Where("story_id IN ?", accessibleIDs).Delete(&models.DevStoryUser{}).Error; err != nil {
			return err
		}
		for _, story := range storys {
			if err := createChangeHistoryTx(tx, creatorID, story.StoryID, 0, 20, ""); err != nil {
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

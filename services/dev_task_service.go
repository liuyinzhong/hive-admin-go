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

func GetTasks(page, pageSize int, params map[string]interface{}) (*utils.PaginationResponse, error) {
	db := database.DB.Model(&models.DevTask{}).Where("del_flag = ?", 0)

	if taskNum, ok := params["taskNum"].(int); ok && taskNum > 0 {
		db = db.Where("task_num = ?", taskNum)
	}
	if taskTitle, ok := params["taskTitle"].(string); ok && taskTitle != "" {
		db = db.Where("task_title LIKE ?", "%"+taskTitle+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if versionID, ok := params["versionId"].(string); ok && versionID != "" {
		db = db.Where("version_id = ?", versionID)
	}
	if taskStatuses, ok := params["taskStatuses"].([]int); ok && len(taskStatuses) > 0 {
		db = db.Where("task_status IN ?", taskStatuses)
	}
	if storyID, ok := params["storyId"].(string); ok && storyID != "" {
		db = db.Where("story_id = ?", storyID)
	}

	sorts := params["sorts"].(string)
	order := utils.BuildOrderBy(sorts, map[string]string{
		"taskTitle":  "task_title",
		"taskStatus": "task_status",
		"startDate":  "start_date",
		"endDate":    "end_date",
	})
	if order == "" {
		order = "create_date DESC"
	}

	return utils.PaginateWithTransform[models.DevTask](db, page, pageSize, order, func(items []models.DevTask) interface{} {
		return buildTaskResponses(items)
	})
}

func GetAllTasks(params map[string]interface{}) ([]models.TaskResponse, error) {
	db := database.DB.Model(&models.DevTask{}).Where("del_flag = ?", 0)

	if taskNum, ok := params["taskNum"].(int); ok && taskNum > 0 {
		db = db.Where("task_num = ?", taskNum)
	}
	if taskTitle, ok := params["taskTitle"].(string); ok && taskTitle != "" {
		db = db.Where("task_title LIKE ?", "%"+taskTitle+"%")
	}
	if projectID, ok := params["projectId"].(string); ok && projectID != "" {
		db = db.Where("project_id = ?", projectID)
	}
	if versionID, ok := params["versionId"].(string); ok && versionID != "" {
		db = db.Where("version_id = ?", versionID)
	}
	if taskStatus, ok := params["taskStatus"].(int); ok && taskStatus >= 0 {
		db = db.Where("task_status = ?", taskStatus)
	}
	if storyID, ok := params["storyId"].(string); ok && storyID != "" {
		db = db.Where("story_id = ?", storyID)
	}

	var tasks []models.DevTask
	err := db.Order("create_date DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	return buildTaskResponses(tasks), nil
}

func buildTaskResponses(tasks []models.DevTask) []models.TaskResponse {
	creatorIDs := make([]string, 0)
	userIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	versionIDs := make([]string, 0)
	moduleIDs := make([]string, 0)
	storyIDs := make([]string, 0)

	for _, t := range tasks {
		if t.CreatorID != nil {
			creatorIDs = append(creatorIDs, *t.CreatorID)
		}
		if t.UserID != nil {
			userIDs = append(userIDs, *t.UserID)
		}
		projectIDs = append(projectIDs, t.ProjectID)
		if t.VersionID != nil {
			versionIDs = append(versionIDs, *t.VersionID)
		}
		if t.ModuleID != nil {
			moduleIDs = append(moduleIDs, *t.ModuleID)
		}
		if t.StoryID != nil {
			storyIDs = append(storyIDs, *t.StoryID)
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

	var responses []models.TaskResponse
	for _, task := range tasks {
		creatorName := creators[utils.StringValue(task.CreatorID)]
		projectTitle := projects[task.ProjectID]
		version := versions[utils.StringValue(task.VersionID)]
		moduleTitle := modules[utils.StringValue(task.ModuleID)]
		storyTitle := stories[utils.StringValue(task.StoryID)]

		var realName, avatar *string
		if task.UserID != nil {
			if u, ok := usersMap[*task.UserID]; ok {
				realName = u.RealName
				avatar = u.Avatar
			}
		}

		percent := 0
		if task.PlanHours > 0 {
			percent = int((task.ActualHours / task.PlanHours) * 100)
		}

		responses = append(responses, models.TaskResponse{
			TaskID:       &task.TaskID,
			StoryID:      task.StoryID,
			StoryTitle:   &storyTitle,
			ModuleID:     task.ModuleID,
			ModuleTitle:  &moduleTitle,
			VersionID:    task.VersionID,
			Version:      &version,
			ProjectID:    &task.ProjectID,
			ProjectTitle: &projectTitle,
			TaskTitle:    task.TaskTitle,
			TaskNum:      task.TaskNum,
			TaskStatus:   intToString(task.TaskStatus),
			TaskType:     intToString(task.TaskType),
			PlanHours:    task.PlanHours,
			ActualHours:  task.ActualHours,
			EndDate:      models.TimeToStringPtr(task.EndDate),
			StartDate:    models.TimeToStringPtr(task.StartDate),
			CreateDate:   models.TimeToStringPtr(task.CreateDate),
			CreatorID:    task.CreatorID,
			CreatorName:  &creatorName,
			UserID:       task.UserID,
			RealName:     realName,
			Avatar:       avatar,
			Percent:      percent,
			TaskRichText: task.TaskRichText,
		})
	}
	return responses
}

func GetTaskByNum(taskNum int) (*models.TaskResponse, error) {
	var task models.DevTask
	err := database.DB.Where("task_num = ? AND del_flag = ?", taskNum, 0).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在")
		}
		return nil, err
	}

	return buildSingleTaskResponse(&task), nil
}

func GetTasksByStoryId(storyId string) ([]models.TaskResponse, error) {
	var tasks []models.DevTask
	err := database.DB.Where("story_id = ? AND del_flag = ?", storyId, 0).Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	var responses []models.TaskResponse
	for _, task := range tasks {
		resp := buildSingleTaskResponse(&task)
		responses = append(responses, *resp)
	}

	return responses, nil
}

func buildSingleTaskResponse(task *models.DevTask) *models.TaskResponse {
	var creatorName string
	if task.CreatorID != nil {
		var user models.SysUser
		database.DB.Where("user_id = ?", *task.CreatorID).First(&user)
		if user.RealName != nil {
			creatorName = *user.RealName
		}
	}

	var realName, avatar *string
	if task.UserID != nil {
		var u models.SysUser
		database.DB.Where("user_id = ?", *task.UserID).First(&u)
		realName = u.RealName
		avatar = u.Avatar
	}

	var projectTitle string
	var project models.DevProject
	database.DB.Where("project_id = ?", task.ProjectID).First(&project)
	if project.ProjectTitle != nil {
		projectTitle = *project.ProjectTitle
	}

	var version string
	if task.VersionID != nil {
		var v models.DevVersion
		database.DB.Where("version_id = ?", *task.VersionID).First(&v)
		if v.Version != nil {
			version = *v.Version
		}
	}

	var moduleTitle string
	if task.ModuleID != nil {
		var m models.DevModule
		database.DB.Where("module_id = ?", *task.ModuleID).First(&m)
		if m.ModuleTitle != nil {
			moduleTitle = *m.ModuleTitle
		}
	}

	var storyTitle string
	if task.StoryID != nil {
		var s models.DevStory
		database.DB.Where("story_id = ?", *task.StoryID).First(&s)
		if s.StoryTitle != nil {
			storyTitle = *s.StoryTitle
		}
	}

	percent := 0
	if task.PlanHours > 0 {
		percent = int((task.ActualHours / task.PlanHours) * 100)
	}

	return &models.TaskResponse{
		TaskID:       &task.TaskID,
		StoryID:      task.StoryID,
		StoryTitle:   &storyTitle,
		ModuleID:     task.ModuleID,
		ModuleTitle:  &moduleTitle,
		VersionID:    task.VersionID,
		Version:      &version,
		ProjectID:    &task.ProjectID,
		ProjectTitle: &projectTitle,
		TaskTitle:    task.TaskTitle,
		TaskNum:      task.TaskNum,
		TaskStatus:   intToString(task.TaskStatus),
		TaskType:     intToString(task.TaskType),
		PlanHours:    task.PlanHours,
		ActualHours:  task.ActualHours,
		EndDate:      models.TimeToStringPtr(task.EndDate),
		StartDate:    models.TimeToStringPtr(task.StartDate),
		CreateDate:   models.TimeToStringPtr(task.CreateDate),
		CreatorID:    task.CreatorID,
		CreatorName:  &creatorName,
		UserID:       task.UserID,
		RealName:     realName,
		Avatar:       avatar,
		Percent:      percent,
		TaskRichText: task.TaskRichText,
	}
}

func CreateTask(req *models.CreateTaskRequest, creatorID string) error {
	startDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.StartDate, time.Local)
	if err != nil {
		return fmt.Errorf("开始时间格式错误")
	}
	endDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.EndDate, time.Local)
	if err != nil {
		return fmt.Errorf("结束时间格式错误")
	}

	if endDate.Before(startDate) {
		return fmt.Errorf("结束时间必须大于或等于开始时间")
	}
	taskStatus, err := parseStringInt(req.TaskStatus, "taskStatus")
	if err != nil {
		return err
	}
	taskType, err := parseStringInt(req.TaskType, "taskType")
	if err != nil {
		return err
	}

	taskID := uuid.New().String()
	now := time.Now()

	task := models.DevTask{
		TaskID:       taskID,
		TaskTitle:    req.TaskTitle,
		TaskStatus:   taskStatus,
		TaskType:     taskType,
		PlanHours:    req.PlanHours,
		ProjectID:    req.ProjectID,
		TaskRichText: req.TaskRichText,
		VersionID:    req.VersionID,
		ModuleID:     req.ModuleID,
		StoryID:      req.StoryID,
		UserID:       req.UserID,
		EndDate:      &endDate,
		StartDate:    &startDate,
		CreatorID:    &creatorID,
		CreateDate:   &now,
		UpdateDate:   &now,
		DelFlag:      0,
	}

	err = database.DB.Create(&task).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, taskID, 10, 0, "")
}

func CreateTasks(reqs []models.CreateTaskRequest, creatorID string) error {
	for _, req := range reqs {
		err := CreateTask(&req, creatorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateTask(taskID string, req *models.UpdateTaskRequest, creatorID string) error {
	var task models.DevTask
	err := database.DB.Where("task_id = ? AND del_flag = ?", taskID, 0).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("任务不存在")
		}
		return err
	}

	startDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.StartDate, time.Local)
	if err != nil {
		return fmt.Errorf("开始时间格式错误")
	}
	endDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.EndDate, time.Local)
	if err != nil {
		return fmt.Errorf("结束时间格式错误")
	}

	if endDate.Before(startDate) {
		return fmt.Errorf("结束时间必须大于或等于开始时间")
	}
	taskStatus, err := parseStringInt(req.TaskStatus, "taskStatus")
	if err != nil {
		return err
	}
	taskType, err := parseStringInt(req.TaskType, "taskType")
	if err != nil {
		return err
	}

	now := time.Now()
	err = database.DB.Model(&task).Updates(map[string]interface{}{
		"task_title":     req.TaskTitle,
		"task_status":    taskStatus,
		"task_type":      taskType,
		"plan_hours":     req.PlanHours,
		"project_id":     req.ProjectID,
		"task_rich_text": req.TaskRichText,
		"version_id":     req.VersionID,
		"module_id":      req.ModuleID,
		"story_id":       req.StoryID,
		"user_id":        req.UserID,
		"end_date":       endDate,
		"start_date":     startDate,
		"update_date":    now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, taskID, 10, 10, "")
}

func UpdateTaskField(taskID string, key string, value interface{}, creatorID string) error {
	var task models.DevTask
	err := database.DB.Where("task_id = ? AND del_flag = ?", taskID, 0).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("任务不存在")
		}
		return err
	}

	allowedKeys := map[string]bool{"userId": true, "taskType": true, "startDate": true, "endDate": true}
	if !allowedKeys[key] {
		return fmt.Errorf("只能修改 userId、taskType、startDate、endDate 字段")
	}

	var newStartDate *time.Time = task.StartDate
	var newEndDate *time.Time = task.EndDate

	updateMap := make(map[string]interface{})
	switch key {
	case "userId":
		updateMap["user_id"] = value
	case "taskType":
		updateMap["task_type"] = value
	case "startDate":
		if dateStr, ok := value.(string); ok {
			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
			if err != nil {
				return fmt.Errorf("日期格式错误")
			}
			newStartDate = &t
			updateMap["start_date"] = t
		}
	case "endDate":
		if dateStr, ok := value.(string); ok {
			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
			if err != nil {
				return fmt.Errorf("日期格式错误")
			}
			newEndDate = &t
			updateMap["end_date"] = t
		}
	}

	if newStartDate != nil && newEndDate != nil && newEndDate.Before(*newStartDate) {
		return fmt.Errorf("结束时间必须大于或等于开始时间")
	}

	updateMap["update_date"] = time.Now()

	err = database.DB.Model(&task).Updates(updateMap).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, taskID, 10, 10, "")
}

func UpdateTaskNext(taskID string, taskStatus string, changeRichText string, creatorID string) error {
	var task models.DevTask
	err := database.DB.Where("task_id = ? AND del_flag = ?", taskID, 0).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("任务不存在")
		}
		return err
	}

	taskStatusInt, err := parseStringInt(taskStatus, "taskStatus")
	if err != nil {
		return err
	}

	now := time.Now()
	err = database.DB.Model(&task).Updates(map[string]interface{}{
		"task_status": taskStatusInt,
		"update_date": now,
	}).Error
	if err != nil {
		return err
	}

	return createChangeHistory(creatorID, taskID, 10, 40, changeRichText)
}

func DeleteTasks(taskIDs []string, creatorID string) error {
	var tasks []models.DevTask
	err := database.DB.Where("task_id IN ? AND del_flag = ?", taskIDs, 0).Find(&tasks).Error
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if t.TaskStatus != 0 {
			return fmt.Errorf("只能删除状态为待执行的任务")
		}
	}

	err = database.DB.Model(&models.DevTask{}).Where("task_id IN ?", taskIDs).Updates(map[string]interface{}{
		"del_flag":    1,
		"update_date": time.Now(),
	}).Error
	if err != nil {
		return err
	}

	for _, t := range tasks {
		createChangeHistory(creatorID, t.TaskID, 10, 20, "")
	}
	return nil
}

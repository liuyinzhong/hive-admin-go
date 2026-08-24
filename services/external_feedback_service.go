package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// externalFeedbackSource 外部反馈工单固定写入的来源值，对应字典 STORY_SOURCE/BUG_SOURCE 的 10。
const externalFeedbackSource = 10

// ErrExternalFeedbackInvalidType 工单类型不合法
var ErrExternalFeedbackInvalidType = errors.New("工单类型 type 只能是 story 或 bug")

// ErrExternalFeedbackTitleRequired 标题必填
var ErrExternalFeedbackTitleRequired = errors.New("工单标题 title 不能为空")

// ErrExternalFeedbackFileIDsInvalid 附件 ID 不是由公开上传接口产生
var ErrExternalFeedbackFileIDsInvalid = errors.New("附件 ID 不存在或不是由公开上传接口产生")

// validateExternalFeedbackFileIDs 校验 fileIds 全部由公开上传接口产生。
// 公开上传接口在 sys_file.creator_id 写入 ExternalFeedbackFileCreatorID 占位标记，
// 此处只校验 creator_id 等于该占位标记且 file_id 存在，避免外部伪造内部登录用户上传的文件 ID。
func validateExternalFeedbackFileIDs(tx *gorm.DB, fileIDs []string) error {
	ids := uniqueNonEmptyStrings(fileIDs)
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("文件ID格式错误")
		}
	}
	var count int64
	if err := tx.Model(&models.SysFile{}).
		Where("file_id IN ? AND creator_id = ?", ids, models.ExternalFeedbackFileCreatorID).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return ErrExternalFeedbackFileIDsInvalid
	}
	return nil
}

// buildExternalFeedbackFileIDsStr 将 fileIds 切片拼成逗号分隔字符串，空切片返回空字符串。
func buildExternalFeedbackFileIDsStr(fileIDs []string) (string, *string) {
	ids := uniqueNonEmptyStrings(fileIDs)
	if len(ids) == 0 {
		return "", nil
	}
	str := strings.Join(ids, ",")
	return str, &str
}

// CreateStoryFeedback 处理外部反馈工单提交。
// 不依赖登录态和数据权限，根据 type 路由到 dev_story 或 dev_bug，
// source 固定写 10，creator_id/user_id/project_id/version_id/module_id 全部留空（NULL）。
// fileIds 必须由公开上传接口 /api/public/upload 产生。
// 返回工单编号供调用方作为提交凭据。
func CreateStoryFeedback(req *models.CreateStoryFeedbackRequest) (*models.CreateStoryFeedbackResponse, error) {
	if req.Type != models.ExternalFeedbackTypeStory && req.Type != models.ExternalFeedbackTypeBug {
		return nil, ErrExternalFeedbackInvalidType
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrExternalFeedbackTitleRequired
	}

	var num int
	var err error
	switch req.Type {
	case models.ExternalFeedbackTypeStory:
		num, err = createExternalFeedbackStory(req)
	case models.ExternalFeedbackTypeBug:
		num, err = createExternalFeedbackBug(req)
	}
	if err != nil {
		return nil, err
	}
	return &models.CreateStoryFeedbackResponse{Num: num, Type: req.Type}, nil
}

// createExternalFeedbackStory 写入一条外部反馈需求记录到 dev_story。
// 不写变更历史，避免变更历史 creator_id 为空字符串污染；内部用户接手后再触发状态流转写变更历史。
func createExternalFeedbackStory(req *models.CreateStoryFeedbackRequest) (int, error) {
	storyID := uuid.New().String()
	now := time.Now()

	fileIDsStr, fileIDsPtr := buildExternalFeedbackFileIDsStr(req.FileIDs)
	_ = fileIDsStr

	var num int
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateExternalFeedbackFileIDs(tx, req.FileIDs); err != nil {
			return err
		}
		story := models.DevStory{
			StoryID:       storyID,
			StoryTitle:    &req.Title,
			StoryType:     0,
			StoryStatus:   0,
			StoryLevel:    0,
			Source:        externalFeedbackSource,
			ProjectID:     "",
			StoryRichText: req.RichText,
			FileIDs:       fileIDsPtr,
			CreateDate:    &now,
			UpdateDate:    &now,
			DelFlag:       0,
			StoryNum:      0,
		}
		if err := tx.Create(&story).Error; err != nil {
			return err
		}
		// StoryID 是手动赋值的主键(UUID)，GORM 仅回填主键；非主键自增的 story_num 需用主键回查
		if err := tx.Raw("SELECT story_num FROM dev_story WHERE story_id = ?", storyID).Scan(&num).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

// createExternalFeedbackBug 写入一条外部反馈缺陷记录到 dev_bug。
// 不写变更历史，避免变更历史 creator_id 为空字符串污染；内部用户接手后再触发状态流转写变更历史。
func createExternalFeedbackBug(req *models.CreateStoryFeedbackRequest) (int, error) {
	bugID := uuid.New().String()
	now := time.Now()

	fileIDsStr, fileIDsPtr := buildExternalFeedbackFileIDsStr(req.FileIDs)

	var num int
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateExternalFeedbackFileIDs(tx, req.FileIDs); err != nil {
			return err
		}
		bug := models.DevBug{
			BugID:            bugID,
			BugTitle:         &req.Title,
			BugStatus:        0,
			BugConfirmStatus: 0,
			BugLevel:         0,
			BugSource:        externalFeedbackSource,
			BugType:          0,
			BugEnv:           0,
			BugUa:            req.UserAgent,
			ProjectID:        "",
			BugRichText:      req.RichText,
			FileIDs:          fileIDsPtr,
			CreateDate:       &now,
			UpdateDate:       &now,
			DelFlag:          0,
		}
		if err := tx.Create(&bug).Error; err != nil {
			return err
		}
		// BugID 是手动赋值的主键(UUID)，GORM 仅回填主键；非主键自增的 bug_num 需用主键回查
		if err := tx.Raw("SELECT bug_num FROM dev_bug WHERE bug_id = ?", bugID).Scan(&num).Error; err != nil {
			return err
		}
		_ = fileIDsStr
		return nil
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

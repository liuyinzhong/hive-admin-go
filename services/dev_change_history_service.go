package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func CreateChangeHistory(req *models.CreateChangeHistoryRequest, creatorID string, permission datapermission.Permission) error {
	// 携带 changeId 时为编辑本人评论:复用创建接口,仅更新正文为最新内容
	if req.ChangeID != "" {
		return updateChangeComment(req.ChangeID, req.ChangeRichText, creatorID, permission)
	}

	changeID := uuid.New().String()
	now := time.Now()

	businessTypeInt, err := parseStringInt(req.BusinessType, "businessType")
	if err != nil {
		return err
	}
	if err := ensureDevBusinessAccess(req.BusinessID, businessTypeInt, permission); err != nil {
		return err
	}

	changeBehaviorInt, err := parseStringInt(req.ChangeBehavior, "changeBehavior")
	if err != nil {
		return err
	}

	changeHistory := models.DevChangeHistory{
		ChangeID:       changeID,
		BusinessID:     &req.BusinessID,
		BusinessType:   businessTypeInt,
		ChangeBehavior: changeBehaviorInt,
		ChangeRichText: &req.ChangeRichText,
		CreatorID:      &creatorID,
		CreateDate:     &now,
		UpdateDate:     &now,
	}

	err = database.DB.Create(&changeHistory).Error
	if err != nil {
		return err
	}

	return nil
}

// changeBehaviorComment 变更行为:评论(30);仅评论类型的记录允许编辑。
const changeBehaviorComment = 30

// updateChangeComment 编辑评论:仅评论类型(changeBehavior=30)且创建人本人可编辑,
// 只更新正文为最新内容,不追加新的变更记录,不保留编辑历史。访问范围继承评论所属业务对象。
func updateChangeComment(changeID string, changeRichText string, userID string, permission datapermission.Permission) error {
	var history models.DevChangeHistory
	if err := database.DB.Where("change_id = ?", changeID).First(&history).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("评论不存在")
		}
		return err
	}
	if history.ChangeBehavior != changeBehaviorComment {
		return fmt.Errorf("仅评论可以编辑")
	}
	if utils.StringValue(history.CreatorID) != userID {
		return fmt.Errorf("只能编辑自己的评论")
	}
	// 变更记录继承业务对象访问范围
	if err := ensureDevBusinessAccess(utils.StringValue(history.BusinessID), history.BusinessType, permission); err != nil {
		return err
	}
	now := time.Now()
	return database.DB.Model(&models.DevChangeHistory{}).
		Where("change_id = ?", changeID).
		Updates(map[string]interface{}{
			"change_rich_text": changeRichText,
			"update_date":      now,
		}).Error
}

func GetChangeHistory(businessID string, permission datapermission.Permission) ([]models.ChangeHistoryResponse, error) {
	var histories []models.DevChangeHistory
	err := database.DB.Where("business_id = ?", businessID).Order("create_date DESC").Find(&histories).Error
	if err != nil {
		return nil, err
	}
	if len(histories) > 0 {
		businessTypes := make(map[int]struct{}, len(histories))
		for _, history := range histories {
			businessTypes[history.BusinessType] = struct{}{}
		}
		for businessType := range businessTypes {
			if err := ensureDevBusinessAccess(businessID, businessType, permission); err != nil {
				return nil, err
			}
		}
	} else if err := ensureAnyDevBusinessAccess(businessID, permission); err != nil {
		return nil, err
	}

	creatorIDs := make([]string, 0)
	for _, h := range histories {
		if h.CreatorID != nil {
			creatorIDs = append(creatorIDs, *h.CreatorID)
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

	var responses []models.ChangeHistoryResponse
	for _, history := range histories {
		creatorName := creators[utils.StringValue(history.CreatorID)]
		changeItems := make([]models.ChangeItem, 0)
		if history.ChangeItems != nil && *history.ChangeItems != "" {
			// 明细JSON解析失败时降级为空明细,不因存量脏数据阻塞时间线展示
			if err := json.Unmarshal([]byte(*history.ChangeItems), &changeItems); err != nil {
				log.Printf("解析变更明细JSON失败 changeId=%s: %v", history.ChangeID, err)
				changeItems = make([]models.ChangeItem, 0)
			}
		}
		responses = append(responses, models.ChangeHistoryResponse{
			ChangeID:       &history.ChangeID,
			ChangeBehavior: intToString(history.ChangeBehavior),
			ChangeRichText: history.ChangeRichText,
			CreatorID:      history.CreatorID,
			CreatorName:    &creatorName,
			BusinessID:     history.BusinessID,
			BusinessType:   intToString(history.BusinessType),
			ExtendJson:     history.ExtendJson,
			ChangeItems:    changeItems,
			CreateDate:     models.TimeToStringPtr(history.CreateDate),
			UpdateDate:     models.TimeToStringPtr(history.UpdateDate),
		})
	}
	return responses, nil
}

func ensureAnyDevBusinessAccess(businessID string, permission datapermission.Permission) error {
	for _, businessType := range []int{0, 10, 20, 30} {
		if ensureDevBusinessAccess(businessID, businessType, permission) == nil {
			return nil
		}
	}
	return fmt.Errorf("业务数据不存在或无权访问")
}

func ensureDevBusinessAccess(businessID string, businessType int, permission datapermission.Permission) error {
	var query *gorm.DB
	switch businessType {
	case 0:
		query = database.DB.Model(&models.DevStory{}).Where("story_id = ? AND del_flag = ?", businessID, 0)
		query = applyStoryPermission(query, permission)
	case 10:
		query = database.DB.Model(&models.DevTask{}).Where("task_id = ? AND del_flag = ?", businessID, 0)
		query = permission.Apply(query, "dev_task.creator_id", "dev_task.user_id")
	case 20:
		query = database.DB.Model(&models.DevBug{}).Where("bug_id = ? AND del_flag = ?", businessID, 0)
		query = permission.Apply(query, "dev_bug.creator_id", "dev_bug.fix_user_id")
	case 30:
		query = database.DB.Model(&models.DevVersion{}).Where("version_id = ? AND del_flag = ?", businessID, 0)
		query = permission.Apply(query, "dev_version.creator_id")
	default:
		return fmt.Errorf("不支持的业务类型")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("业务数据不存在或无权访问")
	}
	return nil
}

package services

import (
	"time"

	"github.com/google/uuid"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

func CreateChangeHistory(req *models.CreateChangeHistoryRequest, creatorID string) error {
	changeID := uuid.New().String()
	now := time.Now()

	businessTypeInt, err := parseStringInt(req.BusinessType, "businessType")
	if err != nil {
		return err
	}

	changeBehaviorInt, err := parseStringInt(req.ChangeBehavior, "changeBehavior")
	if err != nil {
		return err
	}

	changeHistory := models.DevChangeHistory{
		ChangeID:        changeID,
		BusinessID:      &req.BusinessID,
		BusinessType:    businessTypeInt,
		ChangeBehavior:  changeBehaviorInt,
		ChangeRichText:  &req.ChangeRichText,
		CreatorID:       &creatorID,
		CreateDate:      &now,
		UpdateDate:      &now,
	}

	err = database.DB.Create(&changeHistory).Error
	if err != nil {
		return err
	}

	return nil
}

func GetChangeHistory(businessID string) ([]models.ChangeHistoryResponse, error) {
	var histories []models.DevChangeHistory
	err := database.DB.Where("business_id = ?", businessID).Order("create_date DESC").Find(&histories).Error
	if err != nil {
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
		responses = append(responses, models.ChangeHistoryResponse{
			ChangeID:       &history.ChangeID,
			ChangeBehavior: intToString(history.ChangeBehavior),
			ChangeRichText: history.ChangeRichText,
			CreatorID:      history.CreatorID,
			CreatorName:    &creatorName,
			BusinessID:     history.BusinessID,
			BusinessType:   intToString(history.BusinessType),
			ExtendJson:     history.ExtendJson,
			CreateDate:     models.TimeToStringPtr(history.CreateDate),
			UpdateDate:     models.TimeToStringPtr(history.UpdateDate),
		})
	}
	return responses, nil
}

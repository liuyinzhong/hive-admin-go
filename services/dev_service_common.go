package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int

		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &num1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &num2)
		}

		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	return 0
}

func parseLocalDateTime(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", *value, time.Local)
	if err != nil {
		return nil, fmt.Errorf("时间格式错误，请使用 2006-01-02 15:04:05 格式")
	}

	return &t, nil
}

func parseStringInt(value string, fieldName string) (int, error) {
	if value == "" {
		return 0, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", fieldName)
	}

	return result, nil
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func createChangeHistory(creatorID, businessID string, businessType, changeBehavior int, changeRichText string) error {
	now := time.Now()
	history := models.DevChangeHistory{
		ChangeID:       uuid.New().String(),
		ChangeBehavior: changeBehavior,
		ChangeRichText: &changeRichText,
		CreatorID:      &creatorID,
		BusinessID:     &businessID,
		BusinessType:   businessType,
		CreateDate:     &now,
		UpdateDate:     &now,
	}
	return database.DB.Create(&history).Error
}

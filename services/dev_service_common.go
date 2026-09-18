package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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

// createChangeHistoryTx 在事务内写入一条业务变更记录。
// changeItems 为可选的字段级变更明细,由调用方在写入前通过 buildChangeItems 计算差异后传入;
// 未传或为空时行为与历史版本一致,仅记录行为类型与富文本正文。
func createChangeHistoryTx(tx *gorm.DB, creatorID, businessID string, businessType, changeBehavior int, changeRichText string, changeItems ...models.ChangeItem) error {
	now := time.Now()
	var changeItemsJSON *string
	if len(changeItems) > 0 {
		data, err := json.Marshal(changeItems)
		if err != nil {
			return fmt.Errorf("序列化变更明细失败: %w", err)
		}
		text := string(data)
		changeItemsJSON = &text
	}
	history := models.DevChangeHistory{
		ChangeID:       uuid.New().String(),
		ChangeBehavior: changeBehavior,
		ChangeRichText: &changeRichText,
		CreatorID:      &creatorID,
		BusinessID:     &businessID,
		BusinessType:   businessType,
		ChangeItems:    changeItemsJSON,
		CreateDate:     &now,
		UpdateDate:     &now,
	}
	return tx.Create(&history).Error
}

// updateDevRecordWithHistory 在同一事务内执行业务更新并写入变更记录。
// changeItems 为可选的字段级变更明细,透传给 createChangeHistoryTx 一并落库。
func updateDevRecordWithHistory(
	creatorID, businessID string,
	businessType, changeBehavior int,
	changeRichText string,
	update func(*gorm.DB) error,
	changeItems ...models.ChangeItem,
) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := update(tx); err != nil {
			return err
		}
		return createChangeHistoryTx(tx, creatorID, businessID, businessType, changeBehavior, changeRichText, changeItems...)
	})
}

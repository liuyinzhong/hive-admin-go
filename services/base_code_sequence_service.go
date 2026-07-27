package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/models"
)

type BaseCodeSequenceService struct{}

func NewBaseCodeSequenceService() *BaseCodeSequenceService {
	return &BaseCodeSequenceService{}
}

func (s *BaseCodeSequenceService) NextBusinessCode(tx *gorm.DB, sequenceType, defaultPrefix string, defaultNumberLength int) (string, error) {
	now := time.Now()
	var sequence models.BaseCodeSequence
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sequence_type = ?", sequenceType).
		First(&sequence).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		sequence = models.BaseCodeSequence{
			SequenceType: sequenceType,
			Prefix:       defaultPrefix,
			CurrentValue: 1,
			NumberLength: defaultNumberLength,
			UpdateDate:   &now,
		}
		if err := tx.Create(&sequence).Error; err != nil {
			return "", err
		}
		return formatBusinessCode(sequence.Prefix, sequence.NumberLength, sequence.CurrentValue), nil
	}

	sequence.CurrentValue++
	sequence.UpdateDate = &now
	if err := tx.Save(&sequence).Error; err != nil {
		return "", err
	}
	return formatBusinessCode(sequence.Prefix, sequence.NumberLength, sequence.CurrentValue), nil
}

func formatBusinessCode(prefix string, numberLength, currentValue int) string {
	if numberLength <= 0 {
		numberLength = 6
	}
	return fmt.Sprintf("%s%0*d", prefix, numberLength, currentValue)
}

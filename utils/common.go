package utils

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func TimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func ParseSortParams(sorts string) (map[string]string, error) {
	result := make(map[string]string)
	if sorts == "" {
		return result, nil
	}
	
	pairs := strings.Split(sorts, ";")
	for _, pair := range pairs {
		parts := strings.Split(pair, ",")
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

type PageResult struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

type PaginationResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

func Paginate(db *gorm.DB, page int, pageSize int, dest interface{}) (*PageResult, error) {
	var total int64
	db.Count(&total)
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Find(dest).Error
	if err != nil {
		return nil, err
	}
	
	return &PageResult{
		Items: dest,
		Total: total,
	}, nil
}

func PaginateWithTransform[T any](db *gorm.DB, page int, pageSize int, order string, transform func(items []T) interface{}) (*PaginationResponse, error) {
	var total int64
	db.Count(&total)
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	
	var items []T
	err := db.Order(order).Offset(offset).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}
	
	return &PaginationResponse{
		Items: transform(items),
		Total: total,
	}, nil
}

func ApplySorting(db *gorm.DB, sorts map[string]string, defaultSort string) *gorm.DB {
	if len(sorts) == 0 {
		if defaultSort != "" {
			return db.Order(defaultSort)
		}
		return db
	}
	
	var orderClauses []string
	for field, direction := range sorts {
		if direction == "desc" {
			orderClauses = append(orderClauses, field+" desc")
		} else {
			orderClauses = append(orderClauses, field+" asc")
		}
	}
	
	if len(orderClauses) > 0 {
		db = db.Order(strings.Join(orderClauses, ", "))
	}
	
	return db
}

func BuildOrderBy(sorts string, fieldMap map[string]string) string {
	if sorts == "" {
		return ""
	}
	
	result := make([]string, 0)
	pairs := strings.Split(sorts, ";")
	for _, pair := range pairs {
		parts := strings.Split(pair, ",")
		if len(parts) == 2 {
			if dbField, ok := fieldMap[parts[0]]; ok {
				if parts[1] == "desc" {
					result = append(result, dbField+" desc")
				} else {
					result = append(result, dbField+" asc")
				}
			}
		}
	}
	
	return strings.Join(result, ", ")
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

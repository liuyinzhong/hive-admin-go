package services

import (
	"database/sql"
	"errors"
	"path"
	"strings"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const externalPageType = "external"

var (
	ErrExternalPageNotFound = errors.New("外部页面不存在")
	ErrExternalPageInvalid  = errors.New("外部页面参数错误")
)

type ExternalPageService struct{}

func NewExternalPageService() *ExternalPageService {
	return &ExternalPageService{}
}

func (s *ExternalPageService) GetExternalPages(req models.ExternalPageListRequest) (*utils.PageResult, error) {
	pageNumber, pageSize := normalizeExternalPagePagination(req.Page, req.PageSize)
	query := database.DB.Model(&models.SysMenu{}).
		Where("type = ? AND del_flag = 0", externalPageType)
	if value := strings.TrimSpace(req.Title); value != "" {
		query = query.Where("title LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(req.Name); value != "" {
		query = query.Where("name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(req.Path); value != "" {
		query = query.Where("path LIKE ?", "%"+value+"%")
	}
	if req.Status != nil {
		if !isBinaryStatus(*req.Status) {
			return nil, ErrExternalPageInvalid
		}
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []models.SysMenu
	if err := query.Order("create_date DESC").
		Offset((pageNumber - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.ExternalPageResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, externalPageToResponse(row))
	}
	return &utils.PageResult{Items: items, Total: total}, nil
}

func (s *ExternalPageService) GetExternalPage(id string) (*models.ExternalPageResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrExternalPageNotFound
	}
	var row models.SysMenu
	if err := database.DB.Where("id = ? AND type = ? AND del_flag = 0", id, externalPageType).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExternalPageNotFound
		}
		return nil, err
	}
	result := externalPageToResponse(row)
	return &result, nil
}

func (s *ExternalPageService) CreateExternalPage(req models.CreateExternalPageRequest, creatorID string) error {
	title, name, routePath, status, err := normalizeCreateExternalPage(req)
	if err != nil {
		return err
	}
	creatorID, creatorName := loadExternalPageCreator(creatorID)
	now := time.Now()
	nameValue := name
	pathValue := routePath
	row := models.SysMenu{
		ID:              utils.GenerateUUID(),
		Type:            externalPageType,
		Title:           title,
		Name:            &nameValue,
		Path:            &pathValue,
		Status:          status,
		NoBasicLayout:   1,
		IgnoreAccess:    1,
		CreatorID:       stringPointerOrNil(creatorID),
		CreatorName:     creatorName,
		CreateDate:      &now,
		UpdateDate:      &now,
		DelFlag:         0,
		MaxNumOfOpenTab: -1,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := ensureActiveRouteIdentityUnique(tx, name, routePath, ""); err != nil {
			return err
		}
		return tx.Create(&row).Error
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func (s *ExternalPageService) UpdateExternalPage(id string, req models.UpdateExternalPageRequest) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrExternalPageNotFound
	}
	title := strings.TrimSpace(req.Title)
	routePath := strings.TrimSpace(req.Path)
	if title == "" || !isValidExternalPath(routePath) {
		return ErrExternalPageInvalid
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var row models.SysMenu
		if err := tx.Where("id = ? AND type = ? AND del_flag = 0", id, externalPageType).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExternalPageNotFound
			}
			return err
		}
		if err := ensureActiveRouteIdentityUnique(tx, "", routePath, id); err != nil {
			return err
		}
		now := time.Now()
		return tx.Model(&row).Updates(map[string]interface{}{
			"title":       title,
			"path":        routePath,
			"update_date": &now,
		}).Error
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func (s *ExternalPageService) UpdateExternalPageStatus(id string, status int) error {
	if _, err := uuid.Parse(id); err != nil || !isBinaryStatus(status) {
		return ErrExternalPageInvalid
	}
	now := time.Now()
	result := database.DB.Model(&models.SysMenu{}).
		Where("id = ? AND type = ? AND del_flag = 0", id, externalPageType).
		Updates(map[string]interface{}{"status": status, "update_date": &now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExternalPageNotFound
	}
	return nil
}

func (s *ExternalPageService) DeleteExternalPages(ids []string) error {
	uniqueIDs := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return ErrExternalPageInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return ErrExternalPageInvalid
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var rows []models.SysMenu
		if err := tx.Where("id IN ? AND type = ? AND del_flag = 0", uniqueIDs, externalPageType).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(uniqueIDs) {
			return ErrExternalPageNotFound
		}
		now := time.Now()
		return tx.Model(&models.SysMenu{}).
			Where("id IN ? AND type = ? AND del_flag = 0", uniqueIDs, externalPageType).
			Updates(map[string]interface{}{"del_flag": 1, "update_date": &now}).Error
	})
}

func (s *ExternalPageService) GetPublicExternalPage(name string) (*models.PublicExternalPageResponse, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrExternalPageNotFound
	}
	var row models.SysMenu
	if err := database.DB.Select("name", "path").
		Where("name = ? AND type = ? AND status = 1 AND ignore_access = 1 AND del_flag = 0", name, externalPageType).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExternalPageNotFound
		}
		return nil, err
	}
	if row.Name == nil || row.Path == nil {
		return nil, ErrExternalPageNotFound
	}
	return &models.PublicExternalPageResponse{Name: *row.Name, Path: *row.Path}, nil
}

func normalizeCreateExternalPage(req models.CreateExternalPageRequest) (string, string, string, int, error) {
	title := strings.TrimSpace(req.Title)
	name := strings.TrimSpace(req.Name)
	routePath := strings.TrimSpace(req.Path)
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	if title == "" || name == "" || !isValidExternalPath(routePath) || !isBinaryStatus(status) {
		return "", "", "", 0, ErrExternalPageInvalid
	}
	return title, name, routePath, status, nil
}

func isValidExternalPath(value string) bool {
	if !strings.HasPrefix(value, "/external/") || strings.HasSuffix(value, "/") {
		return false
	}
	if strings.ContainsAny(value, "?#:*") || strings.ContainsAny(value, "\t\r\n ") {
		return false
	}
	return path.Clean(value) == value
}

func isBinaryStatus(status int) bool {
	return status == 0 || status == 1
}

func normalizeExternalPagePagination(pageNumber, pageSize int) (int, int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageNumber, pageSize
}

func loadExternalPageCreator(userID string) (string, *string) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", nil
	}
	var user models.SysUser
	if err := database.DB.Select("user_id", "real_name").
		Where("user_id = ? AND del_flag = 0", userID).
		First(&user).Error; err != nil {
		return userID, nil
	}
	return userID, user.RealName
}

func stringPointerOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func externalPageToResponse(row models.SysMenu) models.ExternalPageResponse {
	result := models.ExternalPageResponse{
		ID:          row.ID,
		Title:       row.Title,
		Status:      row.Status,
		CreatorID:   row.CreatorID,
		CreatorName: row.CreatorName,
		CreateDate:  models.TimeToStringPtr(row.CreateDate),
		UpdateDate:  models.TimeToStringPtr(row.UpdateDate),
	}
	if row.Name != nil {
		result.Name = *row.Name
	}
	if row.Path != nil {
		result.Path = *row.Path
	}
	return result
}

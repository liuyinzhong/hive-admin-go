package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	downloadMaxRowsParamKey     = "DOWNLOAD_MAX_EXPORT_ROWS"
	downloadDefaultMaxRows      = 200000
	downloadFileRetentionDays   = 7
	downloadRecordRetentionDays = 30
	downloadCenterMenuName      = "downloadCenter"
	downloadTaskDirectory       = "data/downloads"
	downloadPreviewTokenExpiry  = 5 * time.Minute
	downloadPreviewTokenIssuer  = "hive-download-preview"
	downloadPreviewURLPrefix    = "/api/public/downloads/preview/"
)

var (
	ErrDownloadTaskLimitReached = errors.New("同时等待或生成的下载任务最多为3个")
	ErrDownloadTaskNotFound     = errors.New("下载任务不存在")
	ErrDownloadFileUnavailable  = errors.New("文件尚未生成或已过期")
	ErrDownloadInvalidDate      = errors.New("日期格式错误")
	ErrDownloadInvalidStatus    = errors.New("任务状态错误")
	ErrDownloadDataChanged      = errors.New("导出数据在生成期间发生变化，请重新发起导出")
	ErrDownloadPreviewInvalid   = errors.New("预览链接无效或已过期")
)

// DownloadPreviewClaims 下载中心预览 token 的 JWT Claims，复用项目 JWT 密钥签名。
// 自包含短时 token，无需数据库存储；token 中携带 taskId 和 userId 用于后续校验。
type DownloadPreviewClaims struct {
	TaskID string `json:"taskId"`
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

type downloadTaskExporter interface {
	Count(payload, creatorID string) (int64, error)
	Export(payload, creatorID, filePath string, totalRows int64, onProgress func(int64)) (int64, error)
}

type downloadTaskManager struct {
	createMu  sync.Mutex
	signal    chan struct{}
	exporters map[string]downloadTaskExporter
}

var defaultDownloadTaskManager = &downloadTaskManager{
	signal: make(chan struct{}, 1),
	exporters: map[string]downloadTaskExporter{
		models.DownloadTaskTypeInventoryBalance: &inventoryBalanceDownloadExporter{},
		models.DownloadTaskTypeDevTask:          &devTaskDownloadExporter{},
	},
}

type DownloadTaskService struct {
	manager *downloadTaskManager
}

func NewDownloadTaskService() *DownloadTaskService {
	return &DownloadTaskService{manager: defaultDownloadTaskManager}
}

// StartDownloadTaskWorker 恢复任务状态并启动单实例串行 Worker 与清理器。
func StartDownloadTaskWorker() {
	if err := os.MkdirAll(downloadTaskDirectory, 0o755); err != nil {
		log.Printf("create download task directory failed: %v", err)
		return
	}
	now := time.Now()
	message := "服务重启，任务已终止"
	var interruptedTasks []models.SysDownloadTask
	if err := database.DB.Where("status = ?", models.DownloadTaskStatusRunning).Find(&interruptedTasks).Error; err != nil {
		log.Printf("load interrupted download tasks failed: %v", err)
	}
	if err := database.DB.Model(&models.SysDownloadTask{}).
		Where("status = ?", models.DownloadTaskStatusRunning).
		Updates(map[string]interface{}{
			"status":         models.DownloadTaskStatusFailed,
			"error_message":  message,
			"completed_date": now,
			"update_date":    now,
		}).Error; err != nil {
		log.Printf("recover download tasks failed: %v", err)
	} else {
		for _, task := range interruptedTasks {
			task.Status = models.DownloadTaskStatusFailed
			defaultDownloadTaskManager.createTerminalMessage(task, "导出失败", message)
		}
	}

	go defaultDownloadTaskManager.run()
	go defaultDownloadTaskManager.runCleanup()
	defaultDownloadTaskManager.wake()
}

func (s *DownloadTaskService) CreateTask(creatorID, taskType, taskName, sourceModule string, request interface{}) (*models.DownloadTaskCreatedResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	s.manager.createMu.Lock()
	defer s.manager.createMu.Unlock()

	var activeCount int64
	if err := database.DB.Model(&models.SysDownloadTask{}).
		Where("creator_id = ? AND status IN ?", creatorID, []string{models.DownloadTaskStatusPending, models.DownloadTaskStatusRunning}).
		Count(&activeCount).Error; err != nil {
		return nil, err
	}
	if activeCount >= 3 {
		return nil, ErrDownloadTaskLimitReached
	}

	now := time.Now()
	task := models.SysDownloadTask{
		ID:             utils.GenerateUUID(),
		CreatorID:      creatorID,
		TaskType:       taskType,
		TaskName:       taskName,
		SourceModule:   sourceModule,
		RequestPayload: string(payload),
		Status:         models.DownloadTaskStatusPending,
		CreateDate:     &now,
		UpdateDate:     &now,
	}
	if err := database.DB.Create(&task).Error; err != nil {
		return nil, err
	}
	s.manager.publish(task)
	s.manager.wake()
	return &models.DownloadTaskCreatedResponse{ID: task.ID}, nil
}

func (s *DownloadTaskService) GetList(userID string, req models.DownloadTaskListRequest) (*utils.PageResult, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.SysDownloadTask{}).Where("creator_id = ?", userID)
	if taskName := strings.TrimSpace(req.TaskName); taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		if !isDownloadTaskStatus(status) {
			return nil, ErrDownloadInvalidStatus
		}
		query = query.Where("status = ?", status)
	}
	start, err := parseDownloadDate(req.CreateDateStart)
	if err != nil {
		return nil, ErrDownloadInvalidDate
	}
	if start != nil {
		query = query.Where("create_date >= ?", *start)
	}
	end, err := parseDownloadDate(req.CreateDateEnd)
	if err != nil {
		return nil, ErrDownloadInvalidDate
	}
	if end != nil {
		query = query.Where("create_date <= ?", *end)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]models.SysDownloadTask, 0)
	if err := query.Order("create_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	responses := make([]models.DownloadTaskResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, models.DownloadTaskResponse{
			ID:            item.ID,
			TaskName:      item.TaskName,
			SourceModule:  item.SourceModule,
			Status:        item.Status,
			TotalRows:     item.TotalRows,
			ProcessedRows: item.ProcessedRows,
			Progress:      item.Progress,
			FileName:      item.FileName,
			FileSize:      item.FileSize,
			ErrorMessage:  item.ErrorMessage,
			CompletedDate: models.TimeToStringPtr(item.CompletedDate),
			ExpireDate:    models.TimeToStringPtr(item.ExpireDate),
			CreateDate:    utils.TimeToString(item.CreateDate),
			UpdateDate:    utils.TimeToString(item.UpdateDate),
		})
	}
	return &utils.PageResult{Items: responses, Total: total}, nil
}

func (s *DownloadTaskService) GetDownloadFile(userID, taskID string) (*models.SysDownloadTask, error) {
	var task models.SysDownloadTask
	if err := database.DB.Where("id = ? AND creator_id = ?", taskID, userID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDownloadTaskNotFound
		}
		return nil, err
	}
	if task.Status != models.DownloadTaskStatusSucceeded || task.FilePath == nil || task.FileName == nil || task.ExpireDate == nil || !task.ExpireDate.After(time.Now()) {
		return nil, ErrDownloadFileUnavailable
	}
	if _, err := os.Stat(*task.FilePath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDownloadFileUnavailable
		}
		return nil, err
	}
	return &task, nil
}

// GeneratePreviewURL 校验当前用户对任务拥有下载权限后，生成短时签名 token，返回前端可拼接的相对路径。
// 复用 GetDownloadFile 完成创建者、状态和文件可用性校验；token 通过 utils.SignShortLivedToken 签发，5 分钟内有效。
func (s *DownloadTaskService) GeneratePreviewURL(userID, taskID string) (string, error) {
	if _, err := s.GetDownloadFile(userID, taskID); err != nil {
		return "", err
	}
	now := time.Now()
	claims := DownloadPreviewClaims{
		TaskID: taskID,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(downloadPreviewTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    downloadPreviewTokenIssuer,
		},
	}
	signed, err := utils.SignShortLivedToken(claims)
	if err != nil {
		return "", err
	}
	return downloadPreviewURLPrefix + signed, nil
}

// GetPreviewFile 校验预览 token 并返回对应任务，供 Controller 流式返回文件。
// 通过 utils.ParseShortLivedToken 确认 token 由本服务签发且未过期；再复用 GetDownloadFile 确保 token 有效期内文件仍可下载。
func (s *DownloadTaskService) GetPreviewFile(tokenString string) (*models.SysDownloadTask, error) {
	claims := &DownloadPreviewClaims{}
	if err := utils.ParseShortLivedToken(tokenString, claims); err != nil {
		return nil, ErrDownloadPreviewInvalid
	}
	if claims.TaskID == "" || claims.UserID == "" {
		return nil, ErrDownloadPreviewInvalid
	}
	task, err := s.GetDownloadFile(claims.UserID, claims.TaskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (m *downloadTaskManager) wake() {
	select {
	case m.signal <- struct{}{}:
	default:
	}
}

func (m *downloadTaskManager) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.signal:
		case <-ticker.C:
		}
		for m.processNext() {
		}
	}
}

func (m *downloadTaskManager) processNext() bool {
	var task models.SysDownloadTask
	if err := database.DB.Where("status = ?", models.DownloadTaskStatusPending).Order("create_date ASC").First(&task).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("load pending download task failed: %v", err)
		}
		return false
	}

	now := time.Now()
	result := database.DB.Model(&models.SysDownloadTask{}).
		Where("id = ? AND status = ?", task.ID, models.DownloadTaskStatusPending).
		Updates(map[string]interface{}{"status": models.DownloadTaskStatusRunning, "update_date": now})
	if result.Error != nil || result.RowsAffected == 0 {
		if result.Error != nil {
			log.Printf("claim download task %s failed: %v", task.ID, result.Error)
		}
		return result.Error == nil
	}
	task.Status = models.DownloadTaskStatusRunning
	m.publish(task)

	exporter := m.exporters[task.TaskType]
	if exporter == nil {
		m.fail(task, "不支持的导出任务类型", nil)
		return true
	}
	totalRows, err := exporter.Count(task.RequestPayload, task.CreatorID)
	if err != nil {
		m.fail(task, "统计导出数据失败", err)
		return true
	}
	if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
		Updates(map[string]interface{}{"total_rows": totalRows, "update_date": time.Now()}).Error; err != nil {
		m.fail(task, "更新任务进度失败", err)
		return true
	}
	task.TotalRows = totalRows
	m.publish(task)
	maxRows := NewSystemParamService().GetInt(downloadMaxRowsParamKey, downloadDefaultMaxRows, 1, 1000000)
	if totalRows > int64(maxRows) {
		m.fail(task, fmt.Sprintf("导出数据共%d行，超过系统上限%d行", totalRows, maxRows), nil)
		return true
	}

	filePath := filepath.Join(downloadTaskDirectory, task.ID+".xlsx")
	lastPublished := int64(-1)
	onProgress := func(processed int64) {
		if processed != totalRows && processed-lastPublished < 500 {
			return
		}
		lastPublished = processed
		progress := 100
		if totalRows > 0 {
			progress = int(processed * 100 / totalRows)
		}
		if progress > 99 {
			progress = 99
		}
		if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{"processed_rows": processed, "progress": progress, "update_date": time.Now()}).Error; err != nil {
			log.Printf("update download task %s progress failed: %v", task.ID, err)
			return
		}
		task.ProcessedRows = processed
		task.Progress = progress
		m.publish(task)
	}
	processedRows, err := exporter.Export(task.RequestPayload, task.CreatorID, filePath, totalRows, onProgress)
	if err != nil {
		if errors.Is(err, ErrDownloadDataChanged) {
			m.fail(task, err.Error(), err)
		} else {
			m.fail(task, "生成导出文件失败", err)
		}
		return true
	}
	if processedRows != task.TotalRows {
		if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{"total_rows": processedRows, "update_date": time.Now()}).Error; err != nil {
			m.fail(task, "更新任务进度失败", err)
			return true
		}
		task.TotalRows = processedRows
	}
	m.succeed(task, filePath)
	return true
}

func (m *downloadTaskManager) succeed(task models.SysDownloadTask, filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		m.fail(task, "读取导出文件失败", err)
		return
	}
	now := time.Now()
	expireDate := now.AddDate(0, 0, downloadFileRetentionDays)
	fileName := fmt.Sprintf("%s_%s.xlsx", task.TaskName, now.Format("20060102_150405"))
	if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"status":         models.DownloadTaskStatusSucceeded,
			"processed_rows": task.TotalRows,
			"progress":       100,
			"file_name":      fileName,
			"file_path":      filePath,
			"file_size":      info.Size(),
			"error_message":  nil,
			"completed_date": now,
			"expire_date":    expireDate,
			"update_date":    now,
		}).Error; err != nil {
		m.fail(task, "保存导出结果失败", err)
		return
	}
	task.Status = models.DownloadTaskStatusSucceeded
	task.ProcessedRows = task.TotalRows
	task.Progress = 100
	m.publish(task)
	m.createTerminalMessage(task, "导出完成", "文件已生成，可前往下载中心下载")
}

func (m *downloadTaskManager) fail(task models.SysDownloadTask, message string, cause error) {
	if cause != nil {
		log.Printf("download task %s failed: %s: %v", task.ID, message, cause)
	}
	if err := os.Remove(filepath.Join(downloadTaskDirectory, task.ID+".xlsx")); err != nil && !os.IsNotExist(err) {
		log.Printf("remove failed download task file %s failed: %v", task.ID, err)
	}
	now := time.Now()
	if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"status":         models.DownloadTaskStatusFailed,
			"error_message":  message,
			"completed_date": now,
			"update_date":    now,
		}).Error; err != nil {
		log.Printf("mark download task %s failed: %v", task.ID, err)
		return
	}
	task.Status = models.DownloadTaskStatusFailed
	m.publish(task)
	m.createTerminalMessage(task, "导出失败", message)
}

func (m *downloadTaskManager) publish(task models.SysDownloadTask) {
	NewMenuMessageService().PublishDownloadTaskChanged(task.CreatorID, models.DownloadTaskChangedEvent{
		ID:            task.ID,
		Status:        task.Status,
		TotalRows:     task.TotalRows,
		ProcessedRows: task.ProcessedRows,
		Progress:      task.Progress,
	})
}

func (m *downloadTaskManager) createTerminalMessage(task models.SysDownloadTask, title, content string) {
	if err := NewMenuMessageService().CreateMenuMessageForMenuName(task.CreatorID, downloadCenterMenuName, title, task.TaskName+"："+content); err != nil {
		log.Printf("create terminal menu message for download task %s failed: %v", task.ID, err)
	}
}

func (m *downloadTaskManager) runCleanup() {
	m.cleanup()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		m.cleanup()
	}
}

func (m *downloadTaskManager) cleanup() {
	now := time.Now()
	var expired []models.SysDownloadTask
	if err := database.DB.Where("file_path IS NOT NULL AND expire_date <= ?", now).Find(&expired).Error; err != nil {
		log.Printf("load expired download files failed: %v", err)
		return
	}
	for _, task := range expired {
		if task.FilePath != nil {
			if err := os.Remove(*task.FilePath); err != nil && !os.IsNotExist(err) {
				log.Printf("remove expired download file for task %s failed: %v", task.ID, err)
				continue
			}
		}
		if err := database.DB.Model(&models.SysDownloadTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{"file_path": nil, "update_date": now}).Error; err != nil {
			log.Printf("clear expired download file path for task %s failed: %v", task.ID, err)
		}
	}

	recordDeadline := now.AddDate(0, 0, -downloadRecordRetentionDays)
	if err := database.DB.Where("create_date < ? AND file_path IS NULL", recordDeadline).Delete(&models.SysDownloadTask{}).Error; err != nil {
		log.Printf("delete expired download task records failed: %v", err)
	}
}

func parseDownloadDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func isDownloadTaskStatus(status string) bool {
	switch status {
	case models.DownloadTaskStatusPending, models.DownloadTaskStatusRunning, models.DownloadTaskStatusSucceeded, models.DownloadTaskStatusFailed:
		return true
	default:
		return false
	}
}

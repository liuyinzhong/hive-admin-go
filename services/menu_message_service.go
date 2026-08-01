package services

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrMenuMessageInvalidMenu  = errors.New("只能选择启用的叶子页面菜单")
	ErrMenuMessageInvalidUsers = errors.New("存在无效或不可用的目标用户")
)

const (
	demoMenuMessageTitle   = "测试消息"
	demoMenuMessageContent = "测试消息"
)

// MenuMessageService 负责菜单消息的持久化、汇总和实时推送。
type MenuMessageService struct {
	hub *menuMessageHub
}

func NewMenuMessageService() *MenuMessageService {
	return &MenuMessageService{hub: newMenuMessageHub()}
}

// GetUnreadSummary 返回当前用户按菜单聚合的未读数量。
func (s *MenuMessageService) GetUnreadSummary(userID string) ([]models.MenuMessageUnreadSummary, error) {
	var rows []models.MenuMessageUnreadSummary
	err := database.DB.Model(&models.SysMenuMessage{}).
		Select("sys_menu_message.menu_id, COALESCE(NULLIF(sys_menu.link, ''), sys_menu.path, '') AS menu_path, COUNT(*) AS unread_count").
		Joins("LEFT JOIN sys_menu ON sys_menu.id = sys_menu_message.menu_id").
		Where("user_id = ? AND read_at IS NULL", userID).
		Group("sys_menu_message.menu_id, sys_menu.link, sys_menu.path").
		Order("sys_menu_message.menu_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = make([]models.MenuMessageUnreadSummary, 0)
	}
	return rows, nil
}

// CreateDemoMessages 为每个目标用户新增指定数量的未读消息。
func (s *MenuMessageService) CreateDemoMessages(req models.CreateMenuMessageRequest) error {
	userIDs := uniqueNonEmptyStrings(req.UserIDs)
	if len(userIDs) == 0 {
		return ErrMenuMessageInvalidUsers
	}

	var createdUserIDs []string
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var menu models.SysMenu
		if err := tx.Where(
			"id = ? AND status = 1 AND del_flag = 0 AND type IN ?",
			req.MenuID,
			[]string{"menu", "embedded", "link"},
		).First(&menu).Error; err != nil {
			return ErrMenuMessageInvalidMenu
		}

		var childCount int64
		if err := tx.Model(&models.SysMenu{}).
			Where("pid = ? AND type != ? AND del_flag = 0", req.MenuID, "button").
			Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return ErrMenuMessageInvalidMenu
		}

		var users []models.SysUser
		if err := tx.Where(
			"user_id IN ? AND status = 1 AND del_flag = 0 AND is_sys = 0",
			userIDs,
		).Find(&users).Error; err != nil {
			return err
		}
		if len(users) != len(userIDs) {
			return ErrMenuMessageInvalidUsers
		}

		now := time.Now()
		messages := make([]models.SysMenuMessage, 0, len(users)*req.Count)
		for _, user := range users {
			for i := 0; i < req.Count; i++ {
				messages = append(messages, models.SysMenuMessage{
					ID:         utils.GenerateUUID(),
					UserID:     user.UserID,
					MenuID:     req.MenuID,
					Title:      demoMenuMessageTitle,
					Content:    demoMenuMessageContent,
					CreateDate: &now,
				})
			}
		}

		if err := tx.CreateInBatches(&messages, 500).Error; err != nil {
			return err
		}
		createdUserIDs = userIDs
		return nil
	})
	if err != nil {
		return err
	}

	for _, userID := range createdUserIDs {
		s.notifyUnreadSummary(userID)
	}
	return nil
}

// MarkRead 将当前用户在指定菜单下的全部未读消息标记为已读。
func (s *MenuMessageService) MarkRead(userID, menuID string) error {
	now := time.Now()
	if err := database.DB.Model(&models.SysMenuMessage{}).
		Where("user_id = ? AND menu_id = ? AND read_at IS NULL", userID, menuID).
		Updates(map[string]interface{}{"read_at": &now}).Error; err != nil {
		return err
	}

	s.notifyUnreadSummary(userID)
	return nil
}

// StreamUnreadSummary 将当前用户的汇总以 SSE 推送，直到客户端断开。
func (s *MenuMessageService) StreamUnreadSummary(c *gin.Context, userID string) error {
	channel, unsubscribe := s.hub.subscribe(userID)
	defer unsubscribe()

	summary, err := s.GetUnreadSummary(userID)
	if err != nil {
		return err
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return errors.New("当前响应不支持 SSE")
	}

	if err := writeMenuMessageSSE(c.Writer, summary); err != nil {
		return err
	}
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case nextSummary := <-channel:
			if err := writeMenuMessageSSE(c.Writer, nextSummary); err != nil {
				return err
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := c.Writer.Write([]byte(": heartbeat\n\n")); err != nil {
				return err
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			return nil
		}
	}
}

func (s *MenuMessageService) notifyUnreadSummary(userID string) {
	summary, err := s.GetUnreadSummary(userID)
	if err != nil {
		log.Printf("menu message summary notification failed for user %s: %v", userID, err)
		return
	}
	s.hub.publish(userID, summary)
}

func writeMenuMessageSSE(w interface{ Write([]byte) (int, error) }, summary []models.MenuMessageUnreadSummary) error {
	payload, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("event: unreadSummary\ndata: " + string(payload) + "\n\n"))
	return err
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type menuMessageHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []models.MenuMessageUnreadSummary]struct{}
}

func newMenuMessageHub() *menuMessageHub {
	return &menuMessageHub{
		subscribers: make(map[string]map[chan []models.MenuMessageUnreadSummary]struct{}),
	}
}

func (h *menuMessageHub) subscribe(userID string) (chan []models.MenuMessageUnreadSummary, func()) {
	channel := make(chan []models.MenuMessageUnreadSummary, 1)

	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[chan []models.MenuMessageUnreadSummary]struct{})
	}
	h.subscribers[userID][channel] = struct{}{}
	h.mu.Unlock()

	return channel, func() {
		h.mu.Lock()
		if subscribers := h.subscribers[userID]; subscribers != nil {
			delete(subscribers, channel)
			if len(subscribers) == 0 {
				delete(h.subscribers, userID)
			}
		}
		h.mu.Unlock()
	}
}

func (h *menuMessageHub) publish(userID string, summary []models.MenuMessageUnreadSummary) {
	h.mu.RLock()
	channels := make([]chan []models.MenuMessageUnreadSummary, 0, len(h.subscribers[userID]))
	for channel := range h.subscribers[userID] {
		channels = append(channels, channel)
	}
	h.mu.RUnlock()

	for _, channel := range channels {
		select {
		case channel <- summary:
		default:
			select {
			case <-channel:
			default:
			}
			select {
			case channel <- summary:
			default:
			}
		}
	}
}

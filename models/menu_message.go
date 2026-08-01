package models

import "time"

// SysMenuMessage 表示发给用户的菜单未读消息。
type SysMenuMessage struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	UserID     string     `gorm:"column:user_id;type:char(36);not null" json:"userId"`
	MenuID     string     `gorm:"column:menu_id;type:char(36);not null" json:"menuId"`
	Title      string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	Content    string     `gorm:"column:content;type:varchar(512);not null" json:"content"`
	ReadAt     *time.Time `gorm:"column:read_at" json:"readAt"`
	CreateDate *time.Time `gorm:"column:create_date;not null" json:"createDate"`
}

func (SysMenuMessage) TableName() string {
	return "sys_menu_message"
}

// MenuMessageUnreadSummary 是按菜单聚合的当前用户未读数量。
type MenuMessageUnreadSummary struct {
	MenuID      string `json:"menuId"`
	MenuPath    string `json:"menuPath"`
	UnreadCount int64  `json:"unreadCount"`
}

// CreateMenuMessageRequest 是 Demo 页面批量创建消息的请求参数。
type CreateMenuMessageRequest struct {
	UserIDs []string `json:"userIds" binding:"required,min=1" example:"[\"UUID\"]"`
	MenuID  string   `json:"menuId" binding:"required" example:"UUID"`
	Count   int      `json:"count" binding:"required,gt=0,lte=1000" example:"3"`
}

// ReadMenuMessageRequest 是当前用户标记菜单消息已读的请求参数。
type ReadMenuMessageRequest struct {
	MenuID string `json:"menuId" binding:"required" example:"UUID"`
}

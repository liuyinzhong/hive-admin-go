package models

import "time"

// SysTokenBlacklist 已撤销的登录凭证黑名单：登出时写入凭证唯一标识（jti）及其原过期时刻，
// 认证中间件按 jti 拦截到原过期时刻为止；过期行由审计日志清理任务顺带删除。
// 技术表，不对外序列化。
type SysTokenBlacklist struct {
	JTI        string    `gorm:"column:jti;type:char(36);primaryKey" json:"-"`
	ExpiresAt  time.Time `gorm:"column:expires_at;type:datetime" json:"-"`
	CreateDate time.Time `gorm:"column:create_date;type:datetime" json:"-"`
}

func (SysTokenBlacklist) TableName() string { return "sys_token_blacklist" }

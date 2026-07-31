package models

import (
	"time"
)

// 渠道类型枚举
const (
	ChannelTypeWechat = "wechat" // 微信支付
	ChannelTypeAlipay = "alipay" // 支付宝
)

// 环境模式枚举
const (
	EnvModeDevelopment = "development" // 开发环境
	EnvModeTesting     = "testing"     // 测试环境
	EnvModeStaging     = "staging"     // 预生产环境
	EnvModeProduction  = "production"  // 生产环境
)

// 启用状态枚举
const (
	PayChannelStatusDisabled = 0 // 禁用
	PayChannelStatusEnabled  = 1 // 启用
)

// 默认渠道枚举
const (
	PayChannelNotDefault = 0 // 非默认
	PayChannelDefault    = 1 // 默认
)

// PayChannel 支付渠道配置表模型
type PayChannel struct {
	ID          string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`                       // 渠道配置ID(UUID带横线)
	ChannelName string     `gorm:"column:channel_name;type:varchar(64);not null" json:"channelName"`   // 渠道配置名称(如"微信支付-生产")
	ChannelType string     `gorm:"column:channel_type;type:varchar(16);not null" json:"channelType"`   // 渠道类型 wechat/alipay
	EnvMode     string     `gorm:"column:env_mode;type:varchar(16);not null" json:"envMode"`           // 环境模式 development/testing/staging/production
	AppID       string     `gorm:"column:app_id;type:varchar(128);not null" json:"appId"`              // 应用ID(微信/支付宝通用)
	ExtraConfig string     `gorm:"column:extra_config;type:text" json:"extraConfig"`                   // 渠道差异化配置 JSON
	NotifyURL   *string    `gorm:"column:notify_url;type:varchar(255)" json:"notifyUrl"`               // 支付回调地址
	Status      int        `gorm:"column:status;type:tinyint;not null;default:0" json:"status"`        // 启用状态 0=禁用 1=启用
	IsDefault   int        `gorm:"column:is_default;type:tinyint;not null;default:0" json:"isDefault"` // 是否默认 0=否 1=是
	Remark      *string    `gorm:"column:remark;type:varchar(255)" json:"remark"`                      // 备注
	CreateDate  *time.Time `gorm:"column:create_date" json:"createDate"`                               // 创建时间
	UpdateDate  *time.Time `gorm:"column:update_date" json:"updateDate"`                               // 更新时间
	DelFlag     int        `gorm:"column:del_flag;type:tinyint;not null;default:0" json:"delFlag"`     // 逻辑删除标志 0=正常 1=已删除
}

// TableName 指定表名
func (PayChannel) TableName() string {
	return "pay_channel"
}

// PayChannelListRequest 支付渠道分页列表查询请求
type PayChannelListRequest struct {
	Page        int    `form:"page" example:"1"`                // 页码
	PageSize    int    `form:"pageSize" example:"20"`           // 每页大小
	ChannelName string `form:"channelName" example:"微信"`        // 渠道配置名称,模糊搜索
	ChannelType string `form:"channelType" example:"wechat"`    // 渠道类型,精确匹配
	EnvMode     string `form:"envMode" example:"production"`    // 环境模式,精确匹配
	Status      *int   `form:"status" example:"1"`              // 启用状态 0=禁用 1=启用
	IsDefault   *int   `form:"isDefault" example:"1"`           // 是否默认 0=否 1=是
	Sorts       string `form:"sorts" example:"updateDate,desc"` // 排序参数
}

// CreatePayChannelRequest 创建支付渠道请求
type CreatePayChannelRequest struct {
	ChannelName string  `json:"channelName" binding:"required" example:"微信支付-生产"` // 渠道配置名称
	ChannelType string  `json:"channelType" binding:"required" example:"wechat"`  // 渠道类型 wechat/alipay
	EnvMode     string  `json:"envMode" binding:"required" example:"production"`  // 环境模式
	AppID       string  `json:"appId" binding:"required" example:"wx1234"`        // 应用ID
	ExtraConfig string  `json:"extraConfig" example:"{\"mchId\":\"\"}"`           // 渠道差异化配置 JSON
	NotifyURL   *string `json:"notifyUrl" example:"https://api.xxx.com/pay/wx"`   // 支付回调地址
	Status      int     `json:"status" example:"1"`                               // 启用状态 0=禁用 1=启用
	IsDefault   int     `json:"isDefault" example:"1"`                            // 是否默认 0=否 1=是
	Remark      *string `json:"remark" example:"主账号"`                             // 备注
}

// UpdatePayChannelRequest 更新支付渠道请求
type UpdatePayChannelRequest struct {
	ChannelName string  `json:"channelName" binding:"required" example:"微信支付-生产"`
	ChannelType string  `json:"channelType" binding:"required" example:"wechat"`
	EnvMode     string  `json:"envMode" binding:"required" example:"production"`
	AppID       string  `json:"appId" binding:"required" example:"wx1234"`
	ExtraConfig string  `json:"extraConfig" example:"{\"mchId\":\"\"}"`
	NotifyURL   *string `json:"notifyUrl" example:"https://api.xxx.com/pay/wx"`
	Status      int     `json:"status" example:"1"`
	IsDefault   int     `json:"isDefault" example:"1"`
	Remark      *string `json:"remark" example:"主账号"`
}

// UpdatePayChannelStatusRequest 修改支付渠道启用状态请求
type UpdatePayChannelStatusRequest struct {
	Status int `json:"status" binding:"required" example:"1"` // 启用状态 0=禁用 1=启用
}

// UpdatePayChannelDefaultRequest 修改支付渠道默认标记请求
type UpdatePayChannelDefaultRequest struct {
	IsDefault int `json:"isDefault" binding:"required" example:"1"` // 是否默认 0=否 1=是
}

// PayChannelResponse 支付渠道详情响应
type PayChannelResponse struct {
	ID          string  `json:"id" example:"UUID"`                              // 渠道配置ID
	ChannelName string  `json:"channelName" example:"微信支付-生产"`                  // 渠道配置名称
	ChannelType string  `json:"channelType" example:"wechat"`                   // 渠道类型
	EnvMode     string  `json:"envMode" example:"production"`                   // 环境模式
	AppID       string  `json:"appId" example:"wx1234"`                         // 应用ID
	ExtraConfig string  `json:"extraConfig" example:"{\"mchId\":\"\"}"`         // 渠道差异化配置 JSON
	NotifyURL   *string `json:"notifyUrl" example:"https://api.xxx.com/pay/wx"` // 支付回调地址
	Status      int     `json:"status" example:"1"`                             // 启用状态
	IsDefault   int     `json:"isDefault" example:"1"`                          // 是否默认
	Remark      *string `json:"remark" example:"主账号"`                           // 备注
	CreateDate  *string `json:"createDate" example:"2026-07-31 12:00:00"`       // 创建时间
	UpdateDate  *string `json:"updateDate" example:"2026-07-31 12:00:00"`       // 更新时间
}

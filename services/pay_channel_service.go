package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// payChannelSortWhitelist 排序字段白名单,前端字段名 -> 数据库列名
var payChannelSortWhitelist = map[string]string{
	"channelName": "channel_name",
	"channelType": "channel_type",
	"envMode":     "env_mode",
	"status":      "status",
	"isDefault":   "is_default",
	"updateDate":  "update_date",
	"createDate":  "create_date",
}

// PayChannelService 支付渠道服务
type PayChannelService struct{}

// NewPayChannelService 创建支付渠道服务实例
func NewPayChannelService() *PayChannelService {
	return &PayChannelService{}
}

// GetPayChannelList 分页查询支付渠道列表
// req: 查询请求参数(含分页、筛选、排序)
// 返回分页结果和错误
func (s *PayChannelService) GetPayChannelList(req models.PayChannelListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.PayChannel{}).Where("del_flag = 0")

	if req.ChannelName != "" {
		query = query.Where("channel_name LIKE ?", "%"+req.ChannelName+"%")
	}
	if req.ChannelType != "" {
		query = query.Where("channel_type = ?", req.ChannelType)
	}
	if req.EnvMode != "" {
		query = query.Where("env_mode = ?", req.EnvMode)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsDefault)
	}

	// 白名单排序,默认按 update_date desc
	orderClause := buildPayChannelOrderClause(req.Sorts)
	if orderClause != "" {
		query = query.Order(orderClause)
	} else {
		query = query.Order("update_date desc")
	}

	var channels []models.PayChannel
	pageResult, err := utils.Paginate(query, req.Page, req.PageSize, &channels)
	if err != nil {
		return nil, err
	}

	// 转换为响应 DTO 列表
	responses := make([]*models.PayChannelResponse, 0, len(channels))
	for _, c := range channels {
		responses = append(responses, payChannelToResponse(c))
	}
	pageResult.Items = responses
	return pageResult, nil
}

// CreatePayChannel 创建支付渠道
// 校验 channelType/envMode 枚举、extraConfig 为合法 JSON、默认渠道唯一性
func (s *PayChannelService) CreatePayChannel(req models.CreatePayChannelRequest) error {
	if err := validatePayChannelType(req.ChannelType); err != nil {
		return err
	}
	if err := validatePayChannelEnvMode(req.EnvMode); err != nil {
		return err
	}
	if err := validateExtraConfig(req.ExtraConfig); err != nil {
		return err
	}

	now := time.Now()
	channel := models.PayChannel{
		ID:          utils.GenerateUUID(),
		ChannelName: req.ChannelName,
		ChannelType: req.ChannelType,
		EnvMode:     req.EnvMode,
		AppID:       req.AppID,
		ExtraConfig: req.ExtraConfig,
		NotifyURL:   req.NotifyURL,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		Remark:      req.Remark,
		CreateDate:  &now,
		UpdateDate:  &now,
		DelFlag:     0,
	}

	// 事务处理:设为默认时,先取消同 channelType+envMode 下其他默认
	if req.IsDefault == models.PayChannelDefault {
		tx := database.DB.Begin()
		if err := resetOtherDefaultInTx(tx, req.ChannelType, req.EnvMode, ""); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&channel).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	return database.DB.Create(&channel).Error
}

// GetPayChannelDetail 查询支付渠道详情
func (s *PayChannelService) GetPayChannelDetail(id string) (*models.PayChannelResponse, error) {
	var channel models.PayChannel
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&channel).Error
	if err != nil {
		return nil, errors.New("支付渠道不存在")
	}
	return payChannelToResponse(channel), nil
}

// UpdatePayChannel 更新支付渠道
// 允许修改全部字段;设为默认时事务内取消同组其他默认
func (s *PayChannelService) UpdatePayChannel(id string, req models.UpdatePayChannelRequest) error {
	var channel models.PayChannel
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&channel).Error
	if err != nil {
		return errors.New("支付渠道不存在")
	}

	if err := validatePayChannelType(req.ChannelType); err != nil {
		return err
	}
	if err := validatePayChannelEnvMode(req.EnvMode); err != nil {
		return err
	}
	if err := validateExtraConfig(req.ExtraConfig); err != nil {
		return err
	}

	now := time.Now()
	channel.ChannelName = req.ChannelName
	channel.ChannelType = req.ChannelType
	channel.EnvMode = req.EnvMode
	channel.AppID = req.AppID
	channel.ExtraConfig = req.ExtraConfig
	channel.NotifyURL = req.NotifyURL
	channel.Status = req.Status
	channel.IsDefault = req.IsDefault
	channel.Remark = req.Remark
	channel.UpdateDate = &now

	// 事务处理:设为默认时,先取消同 channelType+envMode 下其他默认(排除自身)
	if req.IsDefault == models.PayChannelDefault {
		tx := database.DB.Begin()
		if err := resetOtherDefaultInTx(tx, req.ChannelType, req.EnvMode, id); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Save(&channel).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	return database.DB.Save(&channel).Error
}

// DeletePayChannels 批量逻辑删除支付渠道
func (s *PayChannelService) DeletePayChannels(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return database.DB.Model(&models.PayChannel{}).
		Where("id IN ? AND del_flag = 0", ids).
		Updates(map[string]interface{}{
			"del_flag":    1,
			"update_date": now,
		}).Error
}

// UpdatePayChannelStatus 修改支付渠道启用状态
func (s *PayChannelService) UpdatePayChannelStatus(id string, status int) error {
	var channel models.PayChannel
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&channel).Error
	if err != nil {
		return errors.New("支付渠道不存在")
	}
	if status != models.PayChannelStatusDisabled && status != models.PayChannelStatusEnabled {
		return errors.New("启用状态只能为 0 或 1")
	}

	now := time.Now()
	channel.Status = status
	channel.UpdateDate = &now
	return database.DB.Save(&channel).Error
}

// UpdatePayChannelDefault 修改支付渠道默认标记
// 设为默认时,事务内取消同 channelType+envMode 下其他默认
func (s *PayChannelService) UpdatePayChannelDefault(id string, isDefault int) error {
	var channel models.PayChannel
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&channel).Error
	if err != nil {
		return errors.New("支付渠道不存在")
	}
	if isDefault != models.PayChannelNotDefault && isDefault != models.PayChannelDefault {
		return errors.New("默认标记只能为 0 或 1")
	}

	now := time.Now()
	channel.IsDefault = isDefault
	channel.UpdateDate = &now

	// 设为默认时,事务内取消同组其他默认(排除自身)
	if isDefault == models.PayChannelDefault {
		tx := database.DB.Begin()
		if err := resetOtherDefaultInTx(tx, channel.ChannelType, channel.EnvMode, id); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Save(&channel).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	return database.DB.Save(&channel).Error
}

// resetOtherDefaultInTx 在事务内取消同 channelType+envMode 下其他默认渠道
// excludeID: 需要排除的渠道ID(自身),空字符串表示不排除
func resetOtherDefaultInTx(tx *gorm.DB, channelType, envMode, excludeID string) error {
	query := tx.Model(&models.PayChannel{}).
		Where("channel_type = ? AND env_mode = ? AND del_flag = 0 AND is_default = 1", channelType, envMode)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	return query.Update("is_default", 0).Error
}

// validatePayChannelType 校验渠道类型枚举
func validatePayChannelType(t string) error {
	switch t {
	case models.ChannelTypeWechat, models.ChannelTypeAlipay:
		return nil
	default:
		return errors.New("渠道类型只能为 wechat 或 alipay")
	}
}

// validatePayChannelEnvMode 校验环境模式枚举
func validatePayChannelEnvMode(m string) error {
	switch m {
	case models.EnvModeDevelopment, models.EnvModeTesting, models.EnvModeStaging, models.EnvModeProduction:
		return nil
	default:
		return errors.New("环境模式只能为 development/testing/staging/production")
	}
}

// validateExtraConfig 校验 extraConfig 为合法 JSON(空字符串允许)
func validateExtraConfig(cfg string) error {
	if cfg == "" {
		return nil
	}
	var obj interface{}
	if err := json.Unmarshal([]byte(cfg), &obj); err != nil {
		return errors.New("extraConfig 必须为合法 JSON")
	}
	return nil
}

// buildPayChannelOrderClause 解析排序参数并基于白名单生成 ORDER BY 子句
// sorts 格式: field1,desc;field2,asc
func buildPayChannelOrderClause(sorts string) string {
	if sorts == "" {
		return ""
	}
	var clauses []string
	pairs := strings.Split(sorts, ";")
	for _, pair := range pairs {
		parts := strings.Split(pair, ",")
		if len(parts) != 2 {
			continue
		}
		field := strings.TrimSpace(parts[0])
		dbField, ok := payChannelSortWhitelist[field]
		if !ok {
			continue
		}
		direction := "asc"
		if strings.EqualFold(strings.TrimSpace(parts[1]), "desc") {
			direction = "desc"
		}
		clauses = append(clauses, dbField+" "+direction)
	}
	return strings.Join(clauses, ", ")
}

// payChannelToResponse 将数据库模型转换为响应 DTO
func payChannelToResponse(c models.PayChannel) *models.PayChannelResponse {
	return &models.PayChannelResponse{
		ID:          c.ID,
		ChannelName: c.ChannelName,
		ChannelType: c.ChannelType,
		EnvMode:     c.EnvMode,
		AppID:       c.AppID,
		ExtraConfig: c.ExtraConfig,
		NotifyURL:   c.NotifyURL,
		Status:      c.Status,
		IsDefault:   c.IsDefault,
		Remark:      c.Remark,
		CreateDate:  models.TimeToStringPtr(c.CreateDate),
		UpdateDate:  models.TimeToStringPtr(c.UpdateDate),
	}
}

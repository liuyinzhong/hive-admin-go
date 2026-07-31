package services

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// paramKeyRegex 参数键正则:全大写下划线分词,每段为字母开头、后跟字母或数字,段间用下划线连接
// 示例:SYS_SESSION_TIMEOUT、SYS_UPLOAD_MAX_SIZE
var paramKeyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z][A-Z0-9]*)*$`)

// paramSortWhitelist 排序字段白名单,前端字段名 -> 数据库列名
var paramSortWhitelist = map[string]string{
	"paramKey":   "param_key",
	"paramType":  "param_type",
	"isPublic":   "is_public",
	"updateDate": "update_date",
	"createDate": "create_date",
}

// SystemParamService 系统参数服务
type SystemParamService struct{}

// NewSystemParamService 创建系统参数服务实例
func NewSystemParamService() *SystemParamService {
	return &SystemParamService{}
}

// GetParamList 分页查询参数列表
// req: 查询请求参数(含分页、筛选、排序)
// 返回分页结果和错误
func (s *SystemParamService) GetParamList(req models.ParamListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.SysParam{}).Where("del_flag = 0")

	if req.ParamKey != "" {
		query = query.Where("param_key LIKE ?", "%"+req.ParamKey+"%")
	}
	if req.ParamType != "" {
		query = query.Where("param_type = ?", req.ParamType)
	}
	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}

	// 白名单排序,默认按 update_date desc
	orderClause := buildParamOrderClause(req.Sorts)
	if orderClause != "" {
		query = query.Order(orderClause)
	} else {
		query = query.Order("update_date desc")
	}

	var params []models.SysParam
	pageResult, err := utils.Paginate(query, req.Page, req.PageSize, &params)
	if err != nil {
		return nil, err
	}

	// 转换为响应 DTO 列表
	responses := make([]*models.ParamResponse, 0, len(params))
	for _, p := range params {
		responses = append(responses, sysParamToResponse(p))
	}
	pageResult.Items = responses
	return pageResult, nil
}

// CreateParam 创建参数
// 校验 paramKey 格式与唯一性、paramType 枚举、paramValue 与类型一致性
func (s *SystemParamService) CreateParam(req models.CreateParamRequest) error {
	if err := validateParamKey(req.ParamKey); err != nil {
		return err
	}
	if err := validateParamType(req.ParamType); err != nil {
		return err
	}
	if err := validateParamValue(req.ParamValue, req.ParamType); err != nil {
		return err
	}

	// 校验 paramKey 唯一性
	var count int64
	database.DB.Model(&models.SysParam{}).Where("param_key = ? AND del_flag = 0", req.ParamKey).Count(&count)
	if count > 0 {
		return errors.New("参数键已存在")
	}

	now := time.Now()
	param := models.SysParam{
		ID:         utils.GenerateUUID(),
		ParamKey:   req.ParamKey,
		ParamValue: req.ParamValue,
		ParamType:  req.ParamType,
		IsPublic:   req.IsPublic,
		Remark:     req.Remark,
		CreateDate: &now,
		UpdateDate: &now,
		DelFlag:    0,
	}
	return database.DB.Create(&param).Error
}

// GetParamDetail 查询参数详情
func (s *SystemParamService) GetParamDetail(id string) (*models.ParamResponse, error) {
	var param models.SysParam
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&param).Error
	if err != nil {
		return nil, errors.New("参数不存在")
	}
	return sysParamToResponse(param), nil
}

// UpdateParam 更新参数
// 允许修改 paramKey,需校验唯一性排除自身
func (s *SystemParamService) UpdateParam(id string, req models.UpdateParamRequest) error {
	var param models.SysParam
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&param).Error
	if err != nil {
		return errors.New("参数不存在")
	}

	if err := validateParamKey(req.ParamKey); err != nil {
		return err
	}
	if err := validateParamType(req.ParamType); err != nil {
		return err
	}
	if err := validateParamValue(req.ParamValue, req.ParamType); err != nil {
		return err
	}

	// paramKey 唯一性校验排除自身
	var count int64
	database.DB.Model(&models.SysParam{}).Where("param_key = ? AND del_flag = 0 AND id != ?", req.ParamKey, id).Count(&count)
	if count > 0 {
		return errors.New("参数键已存在")
	}

	now := time.Now()
	param.ParamKey = req.ParamKey
	param.ParamValue = req.ParamValue
	param.ParamType = req.ParamType
	param.IsPublic = req.IsPublic
	param.Remark = req.Remark
	param.UpdateDate = &now
	return database.DB.Save(&param).Error
}

// DeleteParams 批量逻辑删除参数
func (s *SystemParamService) DeleteParams(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return database.DB.Model(&models.SysParam{}).
		Where("id IN ? AND del_flag = 0", ids).
		Updates(map[string]interface{}{
			"del_flag":    1,
			"update_date": now,
		}).Error
}

// GetParamValues 公共参数批量查询
// 仅返回 isPublic=1 && del_flag=0 的参数;keys 为空时返回全部公开参数
// 返回 key->value 映射,值按 paramType 格式化:number->float64, boolean->bool, json->object, string->string
func (s *SystemParamService) GetParamValues(keys []string) (map[string]interface{}, error) {
	query := database.DB.Model(&models.SysParam{}).Where("del_flag = 0 AND is_public = 1")
	if len(keys) > 0 {
		query = query.Where("param_key IN ?", keys)
	}

	var params []models.SysParam
	if err := query.Find(&params).Error; err != nil {
		return nil, err
	}

	values := make(map[string]interface{}, len(params))
	for _, p := range params {
		values[p.ParamKey] = formatParamValue(p.ParamValue, p.ParamType)
	}
	return values, nil
}

// formatParamValue 按 paramType 将字符串值格式化为对应 Go 类型
// number->float64, boolean->bool, json->反序列化对象, string->原字符串
func formatParamValue(value, paramType string) interface{} {
	switch paramType {
	case models.ParamTypeNumber:
		// number 转为 float64(JSON 数字通用类型),转换失败时回退为原字符串
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			return num
		}
		return value
	case models.ParamTypeBoolean:
		// boolean 转为 bool,非标准值回退为原字符串
		if value == "true" {
			return true
		}
		if value == "false" {
			return false
		}
		return value
	case models.ParamTypeJSON:
		// json 反序列化为对象,失败时回退为原字符串
		var obj interface{}
		if err := json.Unmarshal([]byte(value), &obj); err == nil {
			return obj
		}
		return value
	case models.ParamTypeString:
		fallthrough
	default:
		// string 及未知类型保持原字符串
		return value
	}
}

// validateParamKey 校验参数键格式:全大写下划线分词,每段字母开头,长度 1-128
func validateParamKey(key string) error {
	if len(key) == 0 || len(key) > 128 {
		return errors.New("参数键长度需在 1-128 之间")
	}
	if !paramKeyRegex.MatchString(key) {
		return errors.New("参数键需使用全大写下划线分词,每段以字母开头且仅含字母数字")
	}
	return nil
}

// validateParamType 校验参数类型枚举
func validateParamType(t string) error {
	switch t {
	case models.ParamTypeString, models.ParamTypeNumber, models.ParamTypeBoolean, models.ParamTypeJSON:
		return nil
	default:
		return errors.New("参数类型只能为 string/number/boolean/json")
	}
}

// validateParamValue 校验参数值与类型一致性
func validateParamValue(value, t string) error {
	switch t {
	case models.ParamTypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return errors.New("参数类型为 number 时,参数值必须为合法数字")
		}
	case models.ParamTypeBoolean:
		if value != "true" && value != "false" {
			return errors.New("参数类型为 boolean 时,参数值只能为 true 或 false")
		}
	case models.ParamTypeJSON:
		var js interface{}
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return errors.New("参数类型为 json 时,参数值必须为合法 JSON")
		}
	case models.ParamTypeString:
		// string 类型不校验
	default:
		return errors.New("参数类型只能为 string/number/boolean/json")
	}
	return nil
}

// buildParamOrderClause 解析排序参数并基于白名单生成 ORDER BY 子句
// sorts 格式: field1,desc;field2,asc
func buildParamOrderClause(sorts string) string {
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
		dbField, ok := paramSortWhitelist[field]
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

// sysParamToResponse 将数据库模型转换为响应 DTO
func sysParamToResponse(p models.SysParam) *models.ParamResponse {
	return &models.ParamResponse{
		ID:         p.ID,
		ParamKey:   p.ParamKey,
		ParamValue: p.ParamValue,
		ParamType:  p.ParamType,
		IsPublic:   p.IsPublic,
		Remark:     p.Remark,
		CreateDate: models.TimeToStringPtr(p.CreateDate),
		UpdateDate: models.TimeToStringPtr(p.UpdateDate),
	}
}

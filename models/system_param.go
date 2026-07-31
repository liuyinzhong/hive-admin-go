package models

import (
	"time"
)

// 参数值类型枚举
const (
	ParamTypeString  = "string"  // 字符串类型
	ParamTypeNumber  = "number"  // 数字类型
	ParamTypeBoolean = "boolean" // 布尔类型
	ParamTypeJSON    = "json"    // JSON 类型
)

// SysParam 系统参数表模型
type SysParam struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`                     // 参数ID(UUID带横线)
	ParamKey   string     `gorm:"column:param_key;type:varchar(128);not null" json:"paramKey"`      // 参数键(点分命名,全局唯一)
	ParamValue string     `gorm:"column:param_value;type:text" json:"paramValue"`                   // 参数值(字符串存储,按 paramType 解释)
	ParamType  string     `gorm:"column:param_type;type:varchar(16);not null" json:"paramType"`     // 参数类型 string/number/boolean/json
	IsPublic   int        `gorm:"column:is_public;type:tinyint;not null;default:0" json:"isPublic"` // 是否公开 0=否 1=是
	Remark     *string    `gorm:"column:remark;type:varchar(255)" json:"remark"`                    // 备注
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`                             // 创建时间
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`                             // 更新时间
	DelFlag    int        `gorm:"column:del_flag;type:tinyint;not null;default:0" json:"delFlag"`   // 逻辑删除标志 0=正常 1=已删除
}

// TableName 指定表名
func (SysParam) TableName() string {
	return "sys_param"
}

// ParamListRequest 参数分页列表查询请求
type ParamListRequest struct {
	Page      int    `form:"page" example:"1"`                       // 页码
	PageSize  int    `form:"pageSize" example:"20"`                  // 每页大小
	ParamKey  string `form:"paramKey" example:"sys.session.timeout"` // 参数键,模糊搜索
	ParamType string `form:"paramType" example:"number"`             // 参数类型,精确匹配
	IsPublic  *int   `form:"isPublic" example:"1"`                   // 是否公开 0=否 1=是
	Sorts     string `form:"sorts" example:"updateDate,desc"`        // 排序参数,支持 paramKey/paramType/isPublic/updateDate/createDate
}

// CreateParamRequest 创建参数请求
type CreateParamRequest struct {
	ParamKey   string  `json:"paramKey" binding:"required" example:"sys.session.timeout"` // 参数键,点分命名,全局唯一
	ParamValue string  `json:"paramValue" binding:"required" example:"30"`                // 参数值
	ParamType  string  `json:"paramType" binding:"required" example:"number"`             // 参数类型 string/number/boolean/json
	IsPublic   int     `json:"isPublic" example:"1"`                                      // 是否公开 0=否 1=是
	Remark     *string `json:"remark" example:"会话超时分钟数"`                                  // 备注
}

// UpdateParamRequest 更新参数请求(paramKey 可修改,需校验唯一性排除自身)
type UpdateParamRequest struct {
	ParamKey   string  `json:"paramKey" binding:"required" example:"sys.session.timeout"` // 参数键
	ParamValue string  `json:"paramValue" binding:"required" example:"30"`                // 参数值
	ParamType  string  `json:"paramType" binding:"required" example:"number"`             // 参数类型
	IsPublic   int     `json:"isPublic" example:"1"`                                      // 是否公开
	Remark     *string `json:"remark" example:"会话超时分钟数"`                                  // 备注
}

// ParamResponse 参数详情响应
type ParamResponse struct {
	ID         string  `json:"id" example:"UUID"`                        // 参数ID
	ParamKey   string  `json:"paramKey" example:"sys.session.timeout"`   // 参数键
	ParamValue string  `json:"paramValue" example:"30"`                  // 参数值
	ParamType  string  `json:"paramType" example:"number"`               // 参数类型
	IsPublic   int     `json:"isPublic" example:"1"`                     // 是否公开 0=否 1=是
	Remark     *string `json:"remark" example:"会话超时分钟数"`                 // 备注
	CreateDate *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	UpdateDate *string `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
}

// ParamValuesRequest 公共参数批量查询请求
type ParamValuesRequest struct {
	Keys []string `json:"keys" example:"[\"SYS_SESSION_TIMEOUT\",\"SYS_UPLOAD_MAX_SIZE\"]"` // 参数键数组,为空时返回所有 isPublic=1 的参数
}

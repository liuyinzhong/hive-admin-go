package models

import "time"

// BaseClassificationSystem 分类体系表模型
type BaseClassificationSystem struct {
	ClassificationSystemID string     `gorm:"column:classification_system_id;type:char(36);primaryKey" json:"classificationSystemId"`
	SystemCode             string     `gorm:"column:system_code;type:varchar(64)" json:"systemCode"`
	SystemName             string     `gorm:"column:system_name;type:varchar(128)" json:"systemName"`
	Sort                   int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Remark                 *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion             int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID              *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID              *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate             *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate             *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

// TableName 分类体系表名
func (BaseClassificationSystem) TableName() string {
	return "base_classification_system"
}

// BaseClassificationNode 分类体系节点表模型
type BaseClassificationNode struct {
	ClassificationNodeID   string     `gorm:"column:classification_node_id;type:char(36);primaryKey" json:"classificationNodeId"`
	ClassificationSystemID string     `gorm:"column:classification_system_id;type:char(36)" json:"classificationSystemId"`
	NodeCode               string     `gorm:"column:node_code;type:varchar(64)" json:"nodeCode"`
	NodeName               string     `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	ParentID               *string    `gorm:"column:parent_id;type:char(36)" json:"parentId"`
	Status                 int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Sort                   int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	Remark                 *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion             int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID              *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID              *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate             *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate             *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

// TableName 分类体系节点表名
func (BaseClassificationNode) TableName() string {
	return "base_classification_node"
}

// SaveClassificationSystemRequest 分类体系新增/修改请求
type SaveClassificationSystemRequest struct {
	SystemCode         string  `json:"systemCode" binding:"required,max=64" example:"finance"` // 体系编码：全局唯一
	SystemName         string  `json:"systemName" binding:"required,max=128" example:"财务分类"`   // 体系名称
	Sort               int     `json:"sort" example:"1"`                                       // 排序号
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"财务分类体系"`    // 备注
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                         // 期望行版本号（修改时必填）
}

// ClassificationSystemResponse 分类体系响应
type ClassificationSystemResponse struct {
	ClassificationSystemID string  `json:"classificationSystemId" example:"550e8400-e29b-41d4-a716-446655440000"` // 分类体系ID
	SystemCode             string  `json:"systemCode" example:"finance"`                                          // 体系编码
	SystemName             string  `json:"systemName" example:"财务分类"`                                             // 体系名称
	Sort                   int     `json:"sort" example:"1"`                                                      // 排序号
	Remark                 *string `json:"remark" example:"财务分类体系"`                                               // 备注
	RowVersion             int     `json:"rowVersion" example:"1"`                                                // 数据版本号
	CreateDate             *string `json:"createDate" example:"2026-01-15 09:00:00"`                              // 创建时间
	UpdateDate             *string `json:"updateDate" example:"2026-01-15 09:00:00"`                              // 更新时间
}

// ClassificationNodeListRequest 分类节点树查询请求
type ClassificationNodeListRequest struct {
	SystemCode string `form:"systemCode" binding:"required" example:"finance"` // 体系编码
	Keyword    string `form:"keyword" example:"西药"`                            // 关键词：匹配节点编码或名称
	Status     *int   `form:"status" example:"1"`                              // 状态：0停用 1启用
}

// SaveClassificationNodeRequest 分类节点新增/修改请求
type SaveClassificationNodeRequest struct {
	ClassificationSystemID string  `json:"classificationSystemId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 所属体系ID
	NodeCode               string  `json:"nodeCode" binding:"required,max=64" example:"01"`                                          // 节点编码：体系内唯一
	NodeName               string  `json:"nodeName" binding:"required,max=128" example:"西药"`                                         // 节点名称
	ParentID               *string `json:"parentId" example:"550e8400-e29b-41d4-a716-446655440000"`                                  // 父节点ID，根节点为空
	Status                 int     `json:"status" binding:"oneof=0 1" example:"1"`                                                   // 状态
	Sort                   int     `json:"sort" example:"1"`                                                                         // 排序号
	Remark                 *string `json:"remark" binding:"omitempty,max=512" example:"备注"`                                          // 备注
	ExpectedRowVersion     int     `json:"expectedRowVersion" example:"1"`                                                           // 期望行版本号（修改时必填）
}

// UpdateClassificationNodeStatusRequest 分类节点启停请求
type UpdateClassificationNodeStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

// ClassificationNodeTreeResponse 分类节点树响应
type ClassificationNodeTreeResponse struct {
	ClassificationNodeID   string                            `json:"classificationNodeId" example:"550e8400-e29b-41d4-a716-446655440000"`   // 分类节点ID
	ClassificationSystemID string                            `json:"classificationSystemId" example:"550e8400-e29b-41d4-a716-446655440000"` // 所属体系ID
	NodeCode               string                            `json:"nodeCode" example:"01"`                                                 // 节点编码
	NodeName               string                            `json:"nodeName" example:"西药"`                                                 // 节点名称
	ParentID               *string                           `json:"parentId" example:"550e8400-e29b-41d4-a716-446655440000"`               // 父节点ID
	Status                 int                               `json:"status" example:"1"`                                                    // 状态
	Sort                   int                               `json:"sort" example:"1"`                                                      // 排序号
	Remark                 *string                           `json:"remark" example:"备注"`                                                   // 备注
	RowVersion             int                               `json:"rowVersion" example:"1"`                                                // 数据版本号
	CreateDate             *string                           `json:"createDate" example:"2026-01-15 09:00:00"`                              // 创建时间
	UpdateDate             *string                           `json:"updateDate" example:"2026-01-15 09:00:00"`                              // 更新时间
	Children               []*ClassificationNodeTreeResponse `json:"children"`                                                              // 子节点
}

// ClassificationNodeOptionRequest 分类节点公共选项查询请求
type ClassificationNodeOptionRequest struct {
	SystemCode string `form:"systemCode" binding:"required" example:"finance"` // 体系编码
}

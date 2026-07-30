package models

import "time"

const (
	WarehouseStorageTypeNormal       = "NORMAL"
	WarehouseStorageTypeRefrigerated = "REFRIGERATED"
	WarehouseStorageTypeFrozen       = "FROZEN"
	WarehouseStorageTypeCool         = "COOL"
	WarehouseStorageTypeHazardous    = "HAZARDOUS"

	WarehouseBusinessScopeDrug          = "DRUG"
	WarehouseBusinessScopeConsumable    = "CONSUMABLE"
	WarehouseBusinessScopeDevice        = "DEVICE"
	WarehouseBusinessScopeComprehensive = "COMPREHENSIVE"
)

type ErpWarehouse struct {
	WarehouseID             string     `gorm:"column:warehouse_id;type:char(36);primaryKey" json:"warehouseId"`
	WarehouseCode           string     `gorm:"column:warehouse_code;type:varchar(16)" json:"warehouseCode"`
	WarehouseName           string     `gorm:"column:warehouse_name;type:varchar(128)" json:"warehouseName"`
	WarehouseNameNormalized string     `gorm:"column:warehouse_name_normalized;type:varchar(128)" json:"warehouseNameNormalized"`
	StorageType             string     `gorm:"column:storage_type;type:varchar(32)" json:"storageType"`
	BusinessScope           string     `gorm:"column:business_scope;type:varchar(32)" json:"businessScope"`
	Address                 *string    `gorm:"column:address;type:varchar(512)" json:"address"`
	Status                  int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark                  *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion              int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID               *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID               *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate              *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate              *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                 int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ErpWarehouse) TableName() string {
	return "erp_warehouse"
}

type ErpWarehouseListRequest struct {
	Page          int    `form:"page" example:"1"`                                  // 页码
	PageSize      int    `form:"pageSize" example:"20"`                             // 每页数量
	Keyword       string `form:"keyword" example:"WH000001"`                        // 关键词搜索
	StorageType   string `form:"storageType" example:"NORMAL"`                      // 仓库储存类型
	BusinessScope string `form:"businessScope" example:"DRUG"`                      // 仓库业务范围
	Status        *int   `form:"status" example:"1"`                                // 状态
	Sorts         string `form:"sorts" example:"warehouseCode,asc;updateDate,desc"` // 排序
}

type SaveErpWarehouseRequest struct {
	WarehouseName      string  `json:"warehouseName" binding:"required,max=128" example:"中心库"` // 仓库名称
	StorageType        string  `json:"storageType" binding:"required,max=32" example:"NORMAL"` // 仓库储存类型
	BusinessScope      string  `json:"businessScope" binding:"required,max=32" example:"DRUG"` // 仓库业务范围
	Address            *string `json:"address" binding:"omitempty,max=512" example:"北京市朝阳区"`   // 地址
	Status             int     `json:"status" binding:"oneof=0 1" example:"1"`                 // 状态
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"仓库基础资料备注"`  // 备注
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                         // 期望行版本号
}

type UpdateErpWarehouseStatusRequest struct {
	Status             int `json:"status" binding:"oneof=0 1" example:"1"`                  // 状态
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type DeleteErpWarehouseRequest struct {
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ErpWarehouseOptionsRequest struct {
	Keyword  string `form:"keyword" example:"中心库"` // 关键词搜索
	PageSize int    `form:"pageSize" example:"20"` // 返回数量
}

type ErpWarehouseResponse struct {
	WarehouseID   string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 仓库ID
	WarehouseCode string  `json:"warehouseCode" example:"WH000001"`                           // 仓库编码
	WarehouseName string  `json:"warehouseName" example:"中心库"`                                // 仓库名称
	StorageType   string  `json:"storageType" example:"NORMAL"`                               // 仓库储存类型
	BusinessScope string  `json:"businessScope" example:"DRUG"`                               // 仓库业务范围
	Address       *string `json:"address" example:"北京市朝阳区"`                                   // 地址
	Status        int     `json:"status" example:"1"`                                         // 状态
	Remark        *string `json:"remark" example:"仓库基础资料备注"`                                  // 备注
	RowVersion    int     `json:"rowVersion" example:"1"`                                     // 版本号
	CreateDate    *string `json:"createDate" example:"2026-01-15 09:00:00"`                   // 创建时间
	UpdateDate    *string `json:"updateDate" example:"2026-01-15 09:00:00"`                   // 更新时间
}

type ErpWarehouseOptionResponse struct {
	WarehouseID   string `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440000"` // 仓库ID
	WarehouseCode string `json:"warehouseCode" example:"WH000001"`                           // 仓库编码
	WarehouseName string `json:"warehouseName" example:"中心库"`                                // 仓库名称
	StorageType   string `json:"storageType" example:"NORMAL"`                               // 仓库储存类型
	BusinessScope string `json:"businessScope" example:"DRUG"`                               // 仓库业务范围
}

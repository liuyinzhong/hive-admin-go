package models

import "time"

const (
	WarehouseZoneTypeNormal            = "NORMAL"
	WarehouseZoneTypePendingInspection = "PENDING_INSPECTION"
	WarehouseZoneTypeQualified         = "QUALIFIED"
	WarehouseZoneTypeUnqualified       = "UNQUALIFIED"
	WarehouseZoneTypeReturned          = "RETURNED"
)

type ErpWarehouseZone struct {
	ZoneID             string     `gorm:"column:zone_id;type:char(36);primaryKey" json:"zoneId"`
	WarehouseID        string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	ZoneCode           string     `gorm:"column:zone_code;type:varchar(16)" json:"zoneCode"`
	ZoneName           string     `gorm:"column:zone_name;type:varchar(128)" json:"zoneName"`
	ZoneNameNormalized string     `gorm:"column:zone_name_normalized;type:varchar(128)" json:"zoneNameNormalized"`
	ZoneType           string     `gorm:"column:zone_type;type:varchar(32)" json:"zoneType"`
	Remark             *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion         int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID          *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID          *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate         *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate         *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag            int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ErpWarehouseZone) TableName() string {
	return "erp_warehouse_zone"
}

type ErpWarehouseLocation struct {
	LocationID             string     `gorm:"column:location_id;type:char(36);primaryKey" json:"locationId"`
	WarehouseID            string     `gorm:"column:warehouse_id;type:char(36)" json:"warehouseId"`
	ZoneID                 string     `gorm:"column:zone_id;type:char(36)" json:"zoneId"`
	LocationCode           string     `gorm:"column:location_code;type:varchar(16)" json:"locationCode"`
	LocationName           string     `gorm:"column:location_name;type:varchar(128)" json:"locationName"`
	LocationNameNormalized string     `gorm:"column:location_name_normalized;type:varchar(128)" json:"locationNameNormalized"`
	Remark                 *string    `gorm:"column:remark;type:varchar(512)" json:"remark"`
	RowVersion             int        `gorm:"column:row_version;type:int;default:1" json:"rowVersion"`
	CreatorID              *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UpdaterID              *string    `gorm:"column:updater_id;type:char(36)" json:"updaterId"`
	CreateDate             *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate             *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (ErpWarehouseLocation) TableName() string {
	return "erp_warehouse_location"
}

type ErpWarehouseZoneListRequest struct {
	Page     int    `form:"page" example:"1"`                             // 页码
	PageSize int    `form:"pageSize" example:"20"`                        // 每页数量
	Keyword  string `form:"keyword" example:"合格品"`                        // 关键词搜索
	ZoneType string `form:"zoneType" example:"QUALIFIED"`                 // 库区类型
	Sorts    string `form:"sorts" example:"zoneCode,asc;updateDate,desc"` // 排序
}

type SaveErpWarehouseZoneRequest struct {
	ZoneName           string  `json:"zoneName" binding:"required,max=128" example:"合格品区A"`    // 库区名称
	ZoneType           string  `json:"zoneType" binding:"required,max=32" example:"QUALIFIED"` // 库区类型
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"库区备注"`      // 备注
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                         // 期望行版本号
}

type DeleteErpWarehouseZoneRequest struct {
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ErpWarehouseZoneOptionsRequest struct {
	Keyword  string `form:"keyword" example:"合格品"` // 关键词搜索
	PageSize int    `form:"pageSize" example:"20"` // 返回数量
}

type ErpWarehouseZoneResponse struct {
	ZoneID        string  `json:"zoneId" example:"550e8400-e29b-41d4-a716-446655440000"`      // 库区ID
	WarehouseID   string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440001"` // 仓库ID
	ZoneCode      string  `json:"zoneCode" example:"WZ000001"`                                // 库区编码
	ZoneName      string  `json:"zoneName" example:"合格品区A"`                                   // 库区名称
	ZoneType      string  `json:"zoneType" example:"QUALIFIED"`                               // 库区类型
	Remark        *string `json:"remark" example:"库区备注"`                                      // 备注
	LocationCount int     `json:"locationCount" example:"2"`                                  // 货位数量
	RowVersion    int     `json:"rowVersion" example:"1"`                                     // 版本号
	CreateDate    *string `json:"createDate" example:"2026-01-15 09:00:00"`                   // 创建时间
	UpdateDate    *string `json:"updateDate" example:"2026-01-15 09:00:00"`                   // 更新时间
}

type ErpWarehouseZoneOptionResponse struct {
	ZoneID      string `json:"zoneId" example:"550e8400-e29b-41d4-a716-446655440000"`      // 库区ID
	WarehouseID string `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440001"` // 仓库ID
	ZoneCode    string `json:"zoneCode" example:"WZ000001"`                                // 库区编码
	ZoneName    string `json:"zoneName" example:"合格品区A"`                                   // 库区名称
	ZoneType    string `json:"zoneType" example:"QUALIFIED"`                               // 库区类型
}

type ErpWarehouseLocationListRequest struct {
	Page     int    `form:"page" example:"1"`                                 // 页码
	PageSize int    `form:"pageSize" example:"20"`                            // 每页数量
	Keyword  string `form:"keyword" example:"A-01"`                           // 关键词搜索
	Sorts    string `form:"sorts" example:"locationCode,asc;updateDate,desc"` // 排序
}

type SaveErpWarehouseLocationRequest struct {
	LocationName       string  `json:"locationName" binding:"required,max=128" example:"A-01-01"` // 货位名称
	Remark             *string `json:"remark" binding:"omitempty,max=512" example:"货位备注"`         // 备注
	ExpectedRowVersion int     `json:"expectedRowVersion" example:"1"`                            // 期望行版本号
}

type DeleteErpWarehouseLocationRequest struct {
	ExpectedRowVersion int `json:"expectedRowVersion" binding:"required,min=1" example:"1"` // 期望行版本号
}

type ErpWarehouseLocationOptionsRequest struct {
	Keyword  string `form:"keyword" example:"A-01"` // 关键词搜索
	PageSize int    `form:"pageSize" example:"20"`  // 返回数量
}

type ErpWarehouseLocationResponse struct {
	LocationID   string  `json:"locationId" example:"550e8400-e29b-41d4-a716-446655440000"`  // 货位ID
	WarehouseID  string  `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440001"` // 仓库ID
	ZoneID       string  `json:"zoneId" example:"550e8400-e29b-41d4-a716-446655440002"`      // 库区ID
	LocationCode string  `json:"locationCode" example:"WL000001"`                            // 货位编码
	LocationName string  `json:"locationName" example:"A-01-01"`                             // 货位名称
	Remark       *string `json:"remark" example:"货位备注"`                                      // 备注
	RowVersion   int     `json:"rowVersion" example:"1"`                                     // 版本号
	CreateDate   *string `json:"createDate" example:"2026-01-15 09:00:00"`                   // 创建时间
	UpdateDate   *string `json:"updateDate" example:"2026-01-15 09:00:00"`                   // 更新时间
}

type ErpWarehouseLocationOptionResponse struct {
	LocationID   string `json:"locationId" example:"550e8400-e29b-41d4-a716-446655440000"`  // 货位ID
	WarehouseID  string `json:"warehouseId" example:"550e8400-e29b-41d4-a716-446655440001"` // 仓库ID
	ZoneID       string `json:"zoneId" example:"550e8400-e29b-41d4-a716-446655440002"`      // 库区ID
	LocationCode string `json:"locationCode" example:"WL000001"`                            // 货位编码
	LocationName string `json:"locationName" example:"A-01-01"`                             // 货位名称
}

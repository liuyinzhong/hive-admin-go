package models

import "time"

// WfBusinessInstance 业务对象与流程实例的关联记录。
// 一个流程实例只能绑定一个业务对象(通过 instance_id 唯一索引保证);
// 同一业务对象允许多次发起流程(每次发起生成新实例与新关联),历史关联通过 del_flag 软删保留。
type WfBusinessInstance struct {
	BindingID    string     `gorm:"column:binding_id;type:char(36);primaryKey" json:"bindingId"`
	BusinessType string     `gorm:"column:business_type;type:varchar(64);index" json:"businessType"` // 业务归属类型:story/bug/task,与流程定义声明的 BusinessType 一致
	BusinessID   string     `gorm:"column:business_id;type:char(36);index" json:"businessId"`        // 业务对象主键(如 dev_story.story_id)
	InstanceID   string     `gorm:"column:instance_id;type:char(36);uniqueIndex" json:"instanceId"`  // 流程实例ID,唯一索引防止一个实例绑定多个业务
	DefinitionID string     `gorm:"column:definition_id;type:char(36);index" json:"definitionId"`    // 流程定义ID(冗余,便于按定义维度查询业务发起历史)
	StarterID    string     `gorm:"column:starter_id;type:char(36);index" json:"starterId"`          // 发起人ID(冗余,便于按发起人统计业务流程)
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

// TableName 指定业务流程关联表名。
func (WfBusinessInstance) TableName() string { return "wf_business_instance" }

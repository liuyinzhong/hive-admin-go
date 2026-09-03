package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// createWorkflowBusinessInstance 在流程实例创建后写入业务关联记录。
// businessType 来自流程定义声明的 BusinessType,businessID 由业务发起方(如 CreateStory)传入。
// 调用方必须传入流程引擎所在事务 tx,保证关联记录与流程实例同生共死。
func createWorkflowBusinessInstance(tx *gorm.DB, businessType, businessID, instanceID, definitionID, starterID string) error {
	now := time.Now()
	binding := models.WfBusinessInstance{
		BindingID:    utils.GenerateUUID(),
		BusinessType: businessType,
		BusinessID:   businessID,
		InstanceID:   instanceID,
		DefinitionID: definitionID,
		StarterID:    starterID,
		CreateDate:   &now,
		UpdateDate:   &now,
		DelFlag:      0,
	}
	return tx.Create(&binding).Error
}

// getWorkflowBusinessInstanceByInstanceID 按流程实例ID反查业务关联。
// 流程引擎在节点完成时调用,用于从实例定位到业务对象。
// 返回 nil 而非错误表示该实例未绑定业务(纯流程实例),调用方应静默跳过钩子触发。
func getWorkflowBusinessInstanceByInstanceID(tx *gorm.DB, instanceID string) (*models.WfBusinessInstance, error) {
	var binding models.WfBusinessInstance
	err := tx.Where("instance_id = ? AND del_flag = 0", instanceID).First(&binding).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

// GetWorkflowBusinessInstanceByBusiness 公开按业务对象查询当前关联。
// 业务模块在需要展示"我的流程"入口时调用,例如需求详情页跳转到对应流程实例。
func GetWorkflowBusinessInstanceByBusiness(businessType, businessID string) (*models.WfBusinessInstance, error) {
	return getWorkflowBusinessInstanceByBusiness(database.DB, businessType, businessID)
}

func getWorkflowBusinessInstanceByBusiness(tx *gorm.DB, businessType, businessID string) (*models.WfBusinessInstance, error) {
	var binding models.WfBusinessInstance
	err := tx.Where("business_type = ? AND business_id = ? AND del_flag = 0", businessType, businessID).
		Order("create_date DESC").
		First(&binding).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

// GetWorkflowBusinessInstanceDetailResponse 业务关联详情响应(供前端展示)。
type WorkflowBusinessInstanceResponse struct {
	BindingID      string  `json:"bindingId" example:"UUID"`                 // 关联ID
	BusinessType   string  `json:"businessType" example:"story"`             // 业务类型
	BusinessID     string  `json:"businessId" example:"UUID"`                // 业务对象ID
	InstanceID     string  `json:"instanceId" example:"UUID"`                // 流程实例ID
	InstanceNo     string  `json:"instanceNo" example:"WI000001"`            // 流程编号
	DefinitionID   string  `json:"definitionId" example:"UUID"`              // 流程定义ID
	DefinitionName string  `json:"definitionName" example:"需求开发流程"`          // 流程定义名称:详情页展示
	Title          string  `json:"title" example:"需求开发流程-张三"`                // 流程实例标题:详情页展示
	StarterID      string  `json:"starterId" example:"UUID"`                 // 发起人ID
	StarterName    string  `json:"starterName" example:"张三"`                 // 发起人姓名
	Status         string  `json:"status" example:"0"`                       // 流程实例状态:0运行中 1已完成 2已拒绝 3已取消
	CreateDate     *string `json:"createDate" example:"2026-01-15 09:00:00"` // 关联建立时间
}

// GetWorkflowBusinessInstanceList 按业务对象查询全部关联流程实例摘要,按绑定时间倒序。
// 需求详情页展示"关联流程"列表:既包含自动发起的需求流程,也包含结束后动作落地创建时写入的来源审批实例。
// 业务对象未绑定流程时返回空列表(未绑定是正常业务状态,如流程定义未发布期间创建的需求);
// 绑定对应的流程实例缺失时返回错误(脏数据不静默跳过)。
func GetWorkflowBusinessInstanceList(businessType, businessID string) ([]WorkflowBusinessInstanceResponse, error) {
	var bindings []models.WfBusinessInstance
	if err := database.DB.Where("business_type = ? AND business_id = ? AND del_flag = 0", businessType, businessID).
		Order("create_date DESC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return []WorkflowBusinessInstanceResponse{}, nil
	}
	instanceIDs := make([]string, 0, len(bindings))
	starterIDs := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		instanceIDs = append(instanceIDs, binding.InstanceID)
		starterIDs = append(starterIDs, binding.StarterID)
	}
	var instances []models.WfProcessInstance
	if err := database.DB.Where("instance_id IN ? AND del_flag = 0", instanceIDs).Find(&instances).Error; err != nil {
		return nil, err
	}
	instanceMap := make(map[string]models.WfProcessInstance, len(instances))
	for _, instance := range instances {
		instanceMap[instance.InstanceID] = instance
	}
	// 发起人姓名批量解析,不逐条查询;包含已停用用户,展示名不依赖启用状态
	var users []models.SysUser
	if err := database.DB.Where("user_id IN ?", starterIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	userNames := make(map[string]string, len(users))
	for _, user := range users {
		userNames[user.UserID] = workflowUserName(user)
	}
	responses := make([]WorkflowBusinessInstanceResponse, 0, len(bindings))
	for _, binding := range bindings {
		instance, exists := instanceMap[binding.InstanceID]
		if !exists {
			return nil, fmt.Errorf("流程实例不存在")
		}
		responses = append(responses, WorkflowBusinessInstanceResponse{
			BindingID:      binding.BindingID,
			BusinessType:   binding.BusinessType,
			BusinessID:     binding.BusinessID,
			InstanceID:     instance.InstanceID,
			InstanceNo:     instance.InstanceNo,
			DefinitionID:   instance.DefinitionID,
			DefinitionName: instance.DefinitionName,
			Title:          instance.Title,
			StarterID:      binding.StarterID,
			StarterName:    userNames[binding.StarterID],
			Status:         fmt.Sprintf("%d", instance.Status),
			CreateDate:     models.TimeToStringPtr(binding.CreateDate),
		})
	}
	return responses, nil
}

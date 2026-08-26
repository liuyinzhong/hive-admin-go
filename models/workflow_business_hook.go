package models

// BusinessNodeKeyDef 节点业务键定义。
// 描述业务状态钩子支持的一个节点业务键及其语义,供流程设计器下拉选择展示。
type BusinessNodeKeyDef struct {
	NodeKey     string `json:"nodeKey" example:"review"`             // 节点业务键:流程节点上配置的稳定语义标识
	Label       string `json:"label" example:"评审通过"`                 // 中文名:设计器下拉展示
	Description string `json:"description" example:"节点通过后需求状态改为已评审"` // 说明:设计器下拉提示
}

// BusinessHookRegistryItem 业务状态钩子注册项。
// 对应一个已注册的业务类型及其支持的节点业务键列表。
type BusinessHookRegistryItem struct {
	BusinessType string               `json:"businessType" example:"story"` // 业务类型:流程定义声明的业务归属标识
	Label        string               `json:"label" example:"需求"`           // 业务类型中文名:设计器下拉展示
	NodeKeys     []BusinessNodeKeyDef `json:"nodeKeys"`                     // 该业务类型支持的节点业务键列表
}

// BusinessHookRegistryResponse 业务状态钩子注册表查询响应。
type BusinessHookRegistryResponse struct {
	Items []BusinessHookRegistryItem `json:"items"` // 全部已注册业务类型及其节点键
}

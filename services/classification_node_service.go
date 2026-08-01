package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

// ClassificationNodeService 分类体系节点服务
type ClassificationNodeService struct {
	systemService *ClassificationSystemService
}

// NewClassificationNodeService 创建分类节点服务实例
func NewClassificationNodeService() *ClassificationNodeService {
	return &ClassificationNodeService{
		systemService: NewClassificationSystemService(),
	}
}

// GetNodeTree 按体系编码查询节点树，支持关键字过滤并保留祖先链
func (s *ClassificationNodeService) GetNodeTree(req models.ClassificationNodeListRequest) ([]*models.ClassificationNodeTreeResponse, error) {
	system, err := s.systemService.GetSystemByCode(req.SystemCode)
	if err != nil {
		return nil, err
	}
	var nodes []models.BaseClassificationNode
	if err := database.DB.Where("classification_system_id = ? AND del_flag = 0", system.ClassificationSystemID).
		Order("sort asc, create_date asc").Find(&nodes).Error; err != nil {
		return nil, err
	}
	filtered := filterClassificationNodes(nodes, strings.TrimSpace(req.Keyword), req.Status)
	return buildClassificationNodeTree(filtered), nil
}

// GetNodeOptions 公共选项查询：按体系编码返回启用且未删除的节点树
func (s *ClassificationNodeService) GetNodeOptions(req models.ClassificationNodeOptionRequest) ([]*models.ClassificationNodeTreeResponse, error) {
	system, err := s.systemService.GetSystemByCode(req.SystemCode)
	if err != nil {
		return nil, err
	}
	var nodes []models.BaseClassificationNode
	if err := database.DB.Where("classification_system_id = ? AND del_flag = 0 AND status = 1", system.ClassificationSystemID).
		Order("sort asc, create_date asc").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return buildClassificationNodeTree(nodes), nil
}

// GetNodeDetail 查询节点详情
func (s *ClassificationNodeService) GetNodeDetail(nodeID string) (*models.ClassificationNodeTreeResponse, error) {
	if err := validateClassificationUUID(nodeID, "分类节点ID"); err != nil {
		return nil, err
	}
	node, err := s.getNodeByID(database.DB, nodeID)
	if err != nil {
		return nil, err
	}
	return classificationNodeToResponse(*node), nil
}

// CreateNode 创建分类节点
func (s *ClassificationNodeService) CreateNode(req models.SaveClassificationNodeRequest, operatorID string) (*models.ClassificationNodeTreeResponse, error) {
	normalized, err := normalizeNodeSave(req, false)
	if err != nil {
		return nil, err
	}
	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		system, err := s.lockSystemByID(tx, normalized.ClassificationSystemID)
		if err != nil {
			return err
		}
		_ = system
		if err := s.validateNodeParent(tx, normalized.ClassificationSystemID, normalized.ParentID, ""); err != nil {
			return err
		}
		if err := s.ensureNodeCodeUnique(tx, normalized.ClassificationSystemID, normalized.NodeCode, ""); err != nil {
			return err
		}
		now := time.Now()
		createdID = utils.GenerateUUID()
		node := models.BaseClassificationNode{
			ClassificationNodeID:   createdID,
			ClassificationSystemID: normalized.ClassificationSystemID,
			NodeCode:               normalized.NodeCode,
			NodeName:               normalized.NodeName,
			ParentID:               normalized.ParentID,
			Status:                 normalized.Status,
			Sort:                   normalized.Sort,
			Remark:                 normalized.Remark,
			RowVersion:             1,
			CreatorID:              optionalBaseOperatorID(operatorID),
			UpdaterID:              optionalBaseOperatorID(operatorID),
			CreateDate:             &now,
			UpdateDate:             &now,
			DelFlag:                0,
		}
		return tx.Create(&node).Error
	}); err != nil {
		return nil, err
	}
	return s.GetNodeDetail(createdID)
}

// UpdateNode 修改分类节点（含移动父级）
func (s *ClassificationNodeService) UpdateNode(nodeID string, req models.SaveClassificationNodeRequest, operatorID string) (*models.ClassificationNodeTreeResponse, error) {
	if err := validateClassificationUUID(nodeID, "分类节点ID"); err != nil {
		return nil, err
	}
	normalized, err := normalizeNodeSave(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		node, err := s.lockNode(tx, nodeID)
		if err != nil {
			return err
		}
		if node.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 分类节点已被其他人修改，请刷新后重试", ErrClassificationConflict)
		}
		// 节点不能跨体系移动
		if node.ClassificationSystemID != normalized.ClassificationSystemID {
			return fmt.Errorf("%w: 节点不能跨体系移动", ErrClassificationInvalidInput)
		}
		// 父级变更时校验环、子孙、同体系
		if !sameParentPointer(node.ParentID, normalized.ParentID) {
			if err := s.validateNodeParent(tx, node.ClassificationSystemID, normalized.ParentID, nodeID); err != nil {
				return err
			}
			if err := s.ensureNodeNotMoveToDescendant(tx, node.ClassificationSystemID, nodeID, normalized.ParentID); err != nil {
				return err
			}
		}
		if err := s.ensureNodeCodeUnique(tx, node.ClassificationSystemID, normalized.NodeCode, nodeID); err != nil {
			return err
		}
		now := time.Now()
		updates := map[string]interface{}{
			"node_code":   normalized.NodeCode,
			"node_name":   normalized.NodeName,
			"parent_id":   normalized.ParentID,
			"status":      normalized.Status,
			"sort":        normalized.Sort,
			"remark":      normalized.Remark,
			"row_version": node.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}
		return tx.Model(&models.BaseClassificationNode{}).Where("classification_node_id = ?", nodeID).Updates(updates).Error
	}); err != nil {
		return nil, err
	}
	return s.GetNodeDetail(nodeID)
}

// UpdateNodeStatus 修改分类节点启停状态
func (s *ClassificationNodeService) UpdateNodeStatus(nodeID string, req models.UpdateClassificationNodeStatusRequest, operatorID string) (*models.ClassificationNodeTreeResponse, error) {
	if err := validateClassificationUUID(nodeID, "分类节点ID"); err != nil {
		return nil, err
	}
	if err := validateClassificationStatus(req.Status); err != nil {
		return nil, err
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		node, err := s.lockNode(tx, nodeID)
		if err != nil {
			return err
		}
		if node.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 分类节点已被其他人修改，请刷新后重试", ErrClassificationConflict)
		}
		now := time.Now()
		return tx.Model(&models.BaseClassificationNode{}).Where("classification_node_id = ?", nodeID).Updates(map[string]interface{}{
			"status":      req.Status,
			"row_version": node.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}).Error
	}); err != nil {
		return nil, err
	}
	return s.GetNodeDetail(nodeID)
}

// DeleteNode 单条删除分类节点（有子节点时禁止删除）
func (s *ClassificationNodeService) DeleteNode(nodeID string, operatorID string) error {
	if err := validateClassificationUUID(nodeID, "分类节点ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		node, err := s.lockNode(tx, nodeID)
		if err != nil {
			return err
		}
		var childrenCount int64
		if err := tx.Model(&models.BaseClassificationNode{}).
			Where("parent_id = ? AND del_flag = 0", nodeID).
			Count(&childrenCount).Error; err != nil {
			return err
		}
		if childrenCount > 0 {
			return fmt.Errorf("%w: 分类节点存在下级节点，不能删除", ErrClassificationConflict)
		}
		now := time.Now()
		return tx.Model(&models.BaseClassificationNode{}).Where("classification_node_id = ?", nodeID).Updates(map[string]interface{}{
			"status":      0,
			"del_flag":    1,
			"row_version": node.RowVersion + 1,
			"updater_id":  optionalBaseOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

// getNodeByID 按ID查询节点（未删除）
func (s *ClassificationNodeService) getNodeByID(tx *gorm.DB, nodeID string) (*models.BaseClassificationNode, error) {
	var node models.BaseClassificationNode
	if err := tx.Where("classification_node_id = ? AND del_flag = 0", nodeID).First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类节点不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &node, nil
}

// lockNode 事务内加行锁查询节点
func (s *ClassificationNodeService) lockNode(tx *gorm.DB, nodeID string) (*models.BaseClassificationNode, error) {
	var node models.BaseClassificationNode
	if err := tx.Where("classification_node_id = ? AND del_flag = 0", nodeID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类节点不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &node, nil
}

// lockSystemByID 事务内加行锁查询体系（用于节点写入时确认体系存在）
func (s *ClassificationNodeService) lockSystemByID(tx *gorm.DB, systemID string) (*models.BaseClassificationSystem, error) {
	var system models.BaseClassificationSystem
	if err := tx.Where("classification_system_id = ? AND del_flag = 0", systemID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&system).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 分类体系不存在", ErrClassificationNotFound)
		}
		return nil, err
	}
	return &system, nil
}

// validateNodeParent 校验父节点：存在、同体系、未删除、不能是自身
func (s *ClassificationNodeService) validateNodeParent(tx *gorm.DB, systemID string, parentID *string, nodeID string) error {
	parentID = normalizeBaseOptionalString(parentID)
	if parentID == nil {
		return nil
	}
	if err := validateClassificationUUID(*parentID, "父节点ID"); err != nil {
		return err
	}
	if *parentID == nodeID {
		return fmt.Errorf("%w: 父节点不能选择自身", ErrClassificationInvalidInput)
	}
	var parent models.BaseClassificationNode
	if err := tx.Where("classification_node_id = ? AND del_flag = 0", *parentID).First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 父节点不存在", ErrClassificationNotFound)
		}
		return err
	}
	if parent.ClassificationSystemID != systemID {
		return fmt.Errorf("%w: 父节点必须属于同一体系", ErrClassificationInvalidInput)
	}
	return nil
}

// ensureNodeNotMoveToDescendant 校验新父不能是当前节点的子孙，避免形成环
// 通过全量拉取该体系未删除节点，内存构建父子关系，从新父向上遍历判断是否经过当前节点
func (s *ClassificationNodeService) ensureNodeNotMoveToDescendant(tx *gorm.DB, systemID, nodeID string, newParentID *string) error {
	if newParentID == nil {
		return nil
	}
	var nodes []models.BaseClassificationNode
	if err := tx.Where("classification_system_id = ? AND del_flag = 0", systemID).Find(&nodes).Error; err != nil {
		return err
	}
	parentMap := make(map[string]string, len(nodes))
	for _, n := range nodes {
		if n.ParentID != nil && *n.ParentID != "" {
			parentMap[n.ClassificationNodeID] = *n.ParentID
		}
	}
	// 从新父向上遍历，若途中遇到当前节点，说明新父是当前节点的子孙
	current := *newParentID
	visited := map[string]struct{}{nodeID: {}}
	for {
		if _, exists := visited[current]; exists {
			return fmt.Errorf("%w: 不能移动到自身子节点下", ErrClassificationInvalidInput)
		}
		visited[current] = struct{}{}
		parent, exists := parentMap[current]
		if !exists {
			break
		}
		current = parent
	}
	return nil
}

// ensureNodeCodeUnique 校验节点编码在体系内唯一
func (s *ClassificationNodeService) ensureNodeCodeUnique(tx *gorm.DB, systemID, code, excludeID string) error {
	query := tx.Model(&models.BaseClassificationNode{}).
		Where("del_flag = 0 AND classification_system_id = ? AND node_code = ?", systemID, code)
	if excludeID != "" {
		query = query.Where("classification_node_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 节点编码在该体系内已存在", ErrClassificationConflict)
	}
	return nil
}

// normalizedNodeSave 节点保存请求归一化结果
type normalizedNodeSave struct {
	ClassificationSystemID string
	NodeCode               string
	NodeName               string
	ParentID               *string
	Status                 int
	Sort                   int
	Remark                 *string
	ExpectedRowVersion     int
}

// normalizeNodeSave 归一化节点保存请求并校验基础字段
func normalizeNodeSave(req models.SaveClassificationNodeRequest, requireVersion bool) (*normalizedNodeSave, error) {
	systemID := strings.TrimSpace(req.ClassificationSystemID)
	if systemID == "" {
		return nil, fmt.Errorf("%w: 所属体系ID不能为空", ErrClassificationInvalidInput)
	}
	if err := validateClassificationUUID(systemID, "所属体系ID"); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.NodeCode)
	if code == "" {
		return nil, fmt.Errorf("%w: 节点编码不能为空", ErrClassificationInvalidInput)
	}
	if len([]rune(code)) > 64 {
		return nil, fmt.Errorf("%w: 节点编码不能超过64个字符", ErrClassificationInvalidInput)
	}
	name := strings.TrimSpace(req.NodeName)
	if name == "" {
		return nil, fmt.Errorf("%w: 节点名称不能为空", ErrClassificationInvalidInput)
	}
	if len([]rune(name)) > 128 {
		return nil, fmt.Errorf("%w: 节点名称不能超过128个字符", ErrClassificationInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrClassificationInvalidInput)
	}
	if err := validateClassificationStatus(req.Status); err != nil {
		return nil, err
	}
	return &normalizedNodeSave{
		ClassificationSystemID: systemID,
		NodeCode:               code,
		NodeName:               name,
		ParentID:               normalizeBaseOptionalString(req.ParentID),
		Status:                 req.Status,
		Sort:                   req.Sort,
		Remark:                 normalizeBaseOptionalString(req.Remark),
		ExpectedRowVersion:     req.ExpectedRowVersion,
	}, nil
}

// sameParentPointer 判断两个父ID指针是否指向同一父级（均处理为 nil 比较）
func sameParentPointer(a, b *string) bool {
	aNorm := normalizeBaseOptionalString(a)
	bNorm := normalizeBaseOptionalString(b)
	if aNorm == nil && bNorm == nil {
		return true
	}
	if aNorm == nil || bNorm == nil {
		return false
	}
	return *aNorm == *bNorm
}

// filterClassificationNodes 按关键字和状态过滤节点，保留匹配节点的祖先链
func filterClassificationNodes(nodes []models.BaseClassificationNode, keyword string, status *int) []models.BaseClassificationNode {
	if keyword == "" && status == nil {
		return nodes
	}
	byID := make(map[string]models.BaseClassificationNode, len(nodes))
	for _, node := range nodes {
		byID[node.ClassificationNodeID] = node
	}
	keep := make(map[string]struct{})
	keyword = strings.ToLower(keyword)
	for _, node := range nodes {
		matchesKeyword := keyword == "" || strings.Contains(strings.ToLower(node.NodeName), keyword) || strings.Contains(strings.ToLower(node.NodeCode), keyword)
		matchesStatus := status == nil || node.Status == *status
		if !matchesKeyword || !matchesStatus {
			continue
		}
		current := node
		for {
			keep[current.ClassificationNodeID] = struct{}{}
			if current.ParentID == nil || *current.ParentID == "" {
				break
			}
			parent, exists := byID[*current.ParentID]
			if !exists {
				break
			}
			current = parent
		}
	}
	result := make([]models.BaseClassificationNode, 0, len(keep))
	for _, node := range nodes {
		if _, exists := keep[node.ClassificationNodeID]; exists {
			result = append(result, node)
		}
	}
	return result
}

// buildClassificationNodeTree 由扁平节点列表构建树形结构
func buildClassificationNodeTree(nodes []models.BaseClassificationNode) []*models.ClassificationNodeTreeResponse {
	nodeMap := make(map[string]*models.ClassificationNodeTreeResponse, len(nodes))
	for _, node := range nodes {
		nodeMap[node.ClassificationNodeID] = classificationNodeToResponse(node)
	}
	roots := make([]*models.ClassificationNodeTreeResponse, 0)
	for _, node := range nodes {
		resp := nodeMap[node.ClassificationNodeID]
		if node.ParentID == nil || *node.ParentID == "" {
			roots = append(roots, resp)
			continue
		}
		if parent, exists := nodeMap[*node.ParentID]; exists {
			parent.Children = append(parent.Children, resp)
		} else {
			roots = append(roots, resp)
		}
	}
	return roots
}

// classificationNodeToResponse 节点模型转响应
func classificationNodeToResponse(node models.BaseClassificationNode) *models.ClassificationNodeTreeResponse {
	return &models.ClassificationNodeTreeResponse{
		ClassificationNodeID:   node.ClassificationNodeID,
		ClassificationSystemID: node.ClassificationSystemID,
		NodeCode:               node.NodeCode,
		NodeName:               node.NodeName,
		ParentID:               node.ParentID,
		Status:                 node.Status,
		Sort:                   node.Sort,
		Remark:                 node.Remark,
		RowVersion:             node.RowVersion,
		CreateDate:             models.TimeToStringPtr(node.CreateDate),
		UpdateDate:             models.TimeToStringPtr(node.UpdateDate),
		Children:               make([]*models.ClassificationNodeTreeResponse, 0),
	}
}

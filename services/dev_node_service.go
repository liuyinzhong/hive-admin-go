package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// GetAllNodes 获取节点列表
func GetAllNodes(businessID string) ([]*models.NodeResponse, error) {
	db := database.DB.Model(&models.DevNode{})

	if businessID != "" {
		db = db.Where("business_id = ?", businessID)
	}

	var nodes []models.DevNode
	err := db.Order("sort ASC").Find(&nodes).Error
	if err != nil {
		return nil, err
	}

	var responses []*models.NodeResponse
	for _, node := range nodes {
		responses = append(responses, models.DevNodeToNodeResponse(node))
	}

	return responses, nil
}

// CreateNode 创建节点
func CreateNode(req *models.CreateNodeRequest) error {
	now := time.Now()

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if req.Sort > 0 {
		if err := tx.Model(&models.DevNode{}).
			Where("business_id = ? AND sort >= ?", req.BusinessID, req.Sort).
			UpdateColumn("sort", gorm.Expr("sort + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	node := models.DevNode{
		NodeID:       uuid.New().String(),
		Label:        req.Label,
		Value:        req.Value,
		Sort:         req.Sort,
		UserID:       req.UserID,
		NodeType:     req.NodeType,
		Remark:       req.Remark,
		BusinessType: &req.BusinessType,
		BusinessID:   req.BusinessID,
		CreateDate:   &now,
	}

	if err := tx.Create(&node).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteNodes 删除节点
func DeleteNodes(nodeIDs []string) error {
	for _, nodeID := range nodeIDs {
		var node models.DevNode
		if err := database.DB.Where("node_id = ?", nodeID).First(&node).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("节点不存在")
			}
			return err
		}
		if node.NodeType == 0 || node.NodeType == 3 {
			return fmt.Errorf("开始节点和结束节点不能删除")
		}
	}

	return database.DB.Where("node_id IN ?", nodeIDs).Delete(&models.DevNode{}).Error
}

// ApproveNode 节点审批
func ApproveNode(nodeID string, req *models.NodeApproveRequest) error {
	var node models.DevNode
	err := database.DB.Where("node_id = ?", nodeID).First(&node).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("节点不存在")
		}
		return err
	}

	if node.NodeType != 2 {
		return fmt.Errorf("只有审批类型的节点才能审批")
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&node).Updates(map[string]interface{}{
		"result":           req.Result,
		"result_rich_text": req.ResultRichText,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if req.Result == 1 {
		var allNodes []models.DevNode
		tx.Where("business_id = ?", node.BusinessID).Order("sort ASC").Find(&allNodes)

		for i, n := range allNodes {
			if n.NodeID == node.NodeID {
				tx.Model(&n).Update("current", 0)
				tx.Model(&n).Update("end_date", time.Now())
				if i+1 < len(allNodes) {
					tx.Model(&allNodes[i+1]).Update("current", 1)
					tx.Model(&allNodes[i+1]).Update("start_date", time.Now())
				}
				break
			}
		}
	}

	return tx.Commit().Error
}

// NextNode 节点流转
func NextNode(nodeID string) error {
	var node models.DevNode
	err := database.DB.Where("node_id = ?", nodeID).First(&node).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("节点不存在")
		}
		return err
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var allNodes []models.DevNode
	tx.Where("business_id = ?", node.BusinessID).Order("sort ASC").Find(&allNodes)

	currentIndex := -1
	for i, n := range allNodes {
		if n.NodeID == nodeID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		tx.Rollback()
		return fmt.Errorf("节点不存在")
	}

	now := time.Now()

	tx.Model(&allNodes[currentIndex]).Update("current", 0)
	tx.Model(&allNodes[currentIndex]).Update("end_date", now)

	if currentIndex+1 < len(allNodes) {
		nextNode := allNodes[currentIndex+1]
		tx.Model(&nextNode).Update("current", 1)
		tx.Model(&nextNode).Update("start_date", now)
		if nextNode.NodeType == 3 {
			tx.Model(&nextNode).Update("end_date", now)
		}
	}

	return tx.Commit().Error
}
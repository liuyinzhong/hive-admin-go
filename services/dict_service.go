package services

import (
	"errors"
	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DictService struct{}

func NewDictService() *DictService {
	return &DictService{}
}

func (s *DictService) GetDictTree(req models.DictListRequest) ([]*models.DictTreeResponse, error) {
	query := database.DB.Model(&models.SysDict{}).Where("del_flag = 0")

	// type 为精确匹配，且同一字典树通常 type 一致，可在 SQL 层直接过滤
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var dicts []models.SysDict
	err := query.Find(&dicts).Error
	if err != nil {
		return nil, err
	}

	// label/value 为模糊搜索，需保留匹配节点的祖先链与子孙树，故在内存中做树形过滤
	if req.Label != "" || req.Value != "" {
		dicts = s.filterDictTree(dicts, req.Label, req.Value)
	}

	tree := s.buildDictTree(dicts)
	s.sortDictTree(tree, parseDictSortRules(req.Sorts), 0)
	return tree, nil
}

// GetDictValues 公共字典树查询
// 仅返回 status=1 && del_flag=0 的字典，按 value 升序排列，供全系统本地字典消费
func (s *DictService) GetDictValues() ([]*models.DictTreeResponse, error) {
	var dicts []models.SysDict
	err := database.DB.Model(&models.SysDict{}).
		Where("del_flag = 0 AND status = 1").
		Find(&dicts).Error
	if err != nil {
		return nil, err
	}

	tree := s.buildDictTree(dicts)
	s.sortDictTree(tree, []dictSortRule{{field: "value", direction: "asc"}}, 0)
	return tree, nil
}

func (s *DictService) CreateDict(req models.CreateDictRequest) error {
	var count int64
	database.DB.Model(&models.SysDict{}).Where("type = ? AND label = ? AND del_flag = 0", req.Type, req.Label).Count(&count)
	if count > 0 {
		return errors.New("同一类型下标签已存在")
	}

	now := time.Now()
	dict := models.SysDict{
		ID:         utils.GenerateUUID(),
		Pid:        req.Pid,
		Type:       req.Type,
		Label:      &req.Label,
		Value:      req.Value,
		Color:      req.Color,
		Status:     req.Status,
		Remark:     req.Remark,
		CreateDate: &now,
		UpdateDate: &now,
		DelFlag:    0,
	}

	return database.DB.Create(&dict).Error
}

func (s *DictService) GetDictDetail(id string) (*models.DictTreeResponse, error) {
	var dict models.SysDict
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&dict).Error
	if err != nil {
		return nil, errors.New("字典不存在")
	}

	return s.sysDictToResponse(dict, []*models.DictTreeResponse{}), nil
}

func (s *DictService) UpdateDict(id string, req models.UpdateDictRequest) error {
	var dict models.SysDict
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&dict).Error
	if err != nil {
		return errors.New("字典不存在")
	}

	var count int64
	database.DB.Model(&models.SysDict{}).Where("type = ? AND label = ? AND del_flag = 0 AND id != ?", req.Type, req.Label, id).Count(&count)
	if count > 0 {
		return errors.New("同一类型下标签已存在")
	}

	now := time.Now()
	dict.Pid = req.Pid
	dict.Type = req.Type
	dict.Label = &req.Label
	dict.Value = req.Value
	dict.Color = req.Color
	dict.Status = req.Status
	dict.Remark = req.Remark
	dict.UpdateDate = &now

	return database.DB.Save(&dict).Error
}

func (s *DictService) UpdateDictStatus(id string, status int) error {
	var dict models.SysDict
	err := database.DB.Where("id = ? AND del_flag = 0", id).First(&dict).Error
	if err != nil {
		return errors.New("字典不存在")
	}

	now := time.Now()

	childIds := s.getAllChildIds(id)
	childIds = append(childIds, id)

	for _, childId := range childIds {
		var childDict models.SysDict
		if err := database.DB.Where("id = ?", childId).First(&childDict).Error; err == nil {
			childDict.Status = status
			childDict.UpdateDate = &now
			database.DB.Save(&childDict)
		}
	}

	return nil
}

func (s *DictService) DeleteDicts(ids []string) error {
	for _, id := range ids {
		var childrenCount int64
		database.DB.Model(&models.SysDict{}).Where("pid = ? AND del_flag = 0", id).Count(&childrenCount)
		if childrenCount > 0 {
			return errors.New("字典存在子项，不能删除")
		}

		var dict models.SysDict
		err := database.DB.Where("id = ? AND del_flag = 0", id).First(&dict).Error
		if err != nil {
			continue
		}

		now := time.Now()
		dict.DelFlag = 1
		dict.UpdateDate = &now
		database.DB.Save(&dict)
	}

	return nil
}

func (s *DictService) getAllChildIds(id string) []string {
	var childIds []string

	var children []models.SysDict
	database.DB.Where("pid = ? AND del_flag = 0", id).Find(&children)

	for _, child := range children {
		childIds = append(childIds, child.ID)
		childIds = append(childIds, s.getAllChildIds(child.ID)...)
	}

	return childIds
}

// filterDictTree 树形搜索过滤：保留匹配节点及其祖先链与子孙树
// 规则：子级匹配则保留其父级链；父级匹配则保留其全部子孙；全树无匹配则返回空
func (s *DictService) filterDictTree(dicts []models.SysDict, label, value string) []models.SysDict {
	// 构建 id->dict 与 parentID->childIDs 索引
	dictMap := make(map[string]models.SysDict, len(dicts))
	childrenMap := make(map[string][]string, len(dicts))
	for _, d := range dicts {
		dictMap[d.ID] = d
		if d.Pid != nil && *d.Pid != "" {
			childrenMap[*d.Pid] = append(childrenMap[*d.Pid], d.ID)
		}
	}

	// 查找匹配节点（label 与 value 同时填写时取交集）
	var matchedIDs []string
	for _, d := range dicts {
		if matchesDictFilter(d, label, value) {
			matchedIDs = append(matchedIDs, d.ID)
		}
	}
	if len(matchedIDs) == 0 {
		return nil
	}

	// 收集需保留的节点ID：匹配节点 + 祖先链 + 子孙树
	keepSet := make(map[string]bool, len(dicts))
	var collectAncestors func(id string)
	collectAncestors = func(id string) {
		keepSet[id] = true
		d, ok := dictMap[id]
		if !ok || d.Pid == nil || *d.Pid == "" {
			return
		}
		if _, parentExists := dictMap[*d.Pid]; parentExists {
			collectAncestors(*d.Pid)
		}
	}
	var collectDescendants func(id string)
	collectDescendants = func(id string) {
		keepSet[id] = true
		for _, childID := range childrenMap[id] {
			collectDescendants(childID)
		}
	}
	for _, id := range matchedIDs {
		collectAncestors(id)
		collectDescendants(id)
	}

	// 按原顺序输出保留节点
	result := make([]models.SysDict, 0, len(keepSet))
	for _, d := range dicts {
		if keepSet[d.ID] {
			result = append(result, d)
		}
	}
	return result
}

// matchesDictFilter 判断节点是否匹配搜索条件（label/value 均为大小写不敏感的包含匹配，两者同时填写时取交集）
func matchesDictFilter(d models.SysDict, label, value string) bool {
	if label != "" {
		l := ""
		if d.Label != nil {
			l = *d.Label
		}
		if !strings.Contains(strings.ToLower(l), strings.ToLower(label)) {
			return false
		}
	}
	if value != "" {
		v := ""
		if d.Value != nil {
			v = *d.Value
		}
		if !strings.Contains(strings.ToLower(v), strings.ToLower(value)) {
			return false
		}
	}
	return true
}

func (s *DictService) buildDictTree(dicts []models.SysDict) []*models.DictTreeResponse {
	dictMap := make(map[string]*models.DictTreeResponse)
	var roots []*models.DictTreeResponse

	for _, dict := range dicts {
		dictMap[dict.ID] = s.sysDictToResponse(dict, []*models.DictTreeResponse{})
	}

	for _, dict := range dicts {
		if dict.Pid == nil || *dict.Pid == "" {
			roots = append(roots, dictMap[dict.ID])
		} else {
			if parent, exists := dictMap[*dict.Pid]; exists {
				parent.Children = append(parent.Children, dictMap[dict.ID])
			}
		}
	}

	return roots
}

type dictSortRule struct {
	field     string
	direction string
}

func parseDictSortRules(sorts string) []dictSortRule {
	if sorts == "" {
		return nil
	}

	rules := make([]dictSortRule, 0)
	pairs := strings.Split(sorts, ";")
	for _, pair := range pairs {
		parts := strings.Split(pair, ",")
		if len(parts) != 2 {
			continue
		}

		field := parts[0]
		if field != "label" && field != "type" && field != "value" {
			continue
		}

		direction := "asc"
		if strings.EqualFold(parts[1], "desc") {
			direction = "desc"
		}

		rules = append(rules, dictSortRule{
			field:     field,
			direction: direction,
		})
	}

	return rules
}

func (s *DictService) sortDictTree(nodes []*models.DictTreeResponse, rules []dictSortRule, level int) {
	if len(nodes) == 0 {
		return
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return compareDictNodes(nodes[i], nodes[j], rules, level) < 0
	})

	for _, node := range nodes {
		s.sortDictTree(node.Children, rules, level+1)
	}
}

func compareDictNodes(a, b *models.DictTreeResponse, rules []dictSortRule, level int) int {
	if len(rules) == 0 {
		if level == 0 {
			return compareStringForDirection(a.Label, b.Label, "desc")
		}
		return compareValueForDirection(a.Value, b.Value, "desc")
	}

	for _, rule := range rules {
		var result int
		switch rule.field {
		case "label":
			result = compareStringForDirection(a.Label, b.Label, rule.direction)
		case "type":
			result = compareStringForDirection(a.Type, b.Type, rule.direction)
		case "value":
			result = compareValueForDirection(a.Value, b.Value, rule.direction)
		}

		if result != 0 {
			return result
		}
	}

	return 0
}

func compareStringForDirection(a, b, direction string) int {
	left := strings.ToLower(strings.TrimSpace(a))
	right := strings.ToLower(strings.TrimSpace(b))

	if left == right {
		return 0
	}

	if direction == "desc" {
		if left > right {
			return -1
		}
		return 1
	}

	if left < right {
		return -1
	}
	return 1
}

func compareValueForDirection(a, b *string, direction string) int {
	left := ""
	right := ""
	if a != nil {
		left = strings.TrimSpace(*a)
	}
	if b != nil {
		right = strings.TrimSpace(*b)
	}

	leftInt, leftErr := strconv.Atoi(left)
	rightInt, rightErr := strconv.Atoi(right)

	if leftErr == nil && rightErr == nil {
		if leftInt == rightInt {
			return 0
		}
		if direction == "desc" {
			if leftInt > rightInt {
				return -1
			}
			return 1
		}
		if leftInt < rightInt {
			return -1
		}
		return 1
	}

	if leftErr == nil && rightErr != nil {
		return -1
	}
	if leftErr != nil && rightErr == nil {
		return 1
	}

	return compareStringForDirection(left, right, direction)
}

func (s *DictService) sysDictToResponse(dict models.SysDict, children []*models.DictTreeResponse) *models.DictTreeResponse {
	label := ""
	if dict.Label != nil {
		label = *dict.Label
	}

	return &models.DictTreeResponse{
		ID:         dict.ID,
		Pid:        dict.Pid,
		Type:       dict.Type,
		Label:      label,
		Value:      dict.Value,
		Color:      dict.Color,
		Status:     dict.Status,
		Remark:     dict.Remark,
		CreateDate: models.TimeToStringPtr(dict.CreateDate),
		UpdateDate: models.TimeToStringPtr(dict.UpdateDate),
		Children:   children,
	}
}

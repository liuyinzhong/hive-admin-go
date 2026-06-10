package services

import (
	"errors"
	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"time"
)

type DeptService struct{}

func NewDeptService() *DeptService {
	return &DeptService{}
}

func (s *DeptService) GetDeptTree(req models.DeptListRequest) ([]*models.DeptTreeResponse, error) {
	query := database.DB.Model(&models.SysDept{}).Where("del_flag = 0")
	
	if req.DeptTitle != "" {
		query = query.Where("dept_title LIKE ?", "%"+req.DeptTitle+"%")
	}
	
	var depts []models.SysDept
	err := query.Order("create_date desc").Find(&depts).Error
	if err != nil {
		return nil, err
	}
	
	return s.buildDeptTree(depts), nil
}

func (s *DeptService) GetAllDepts() ([]*models.DeptTreeResponse, error) {
	var depts []models.SysDept
	err := database.DB.Where("del_flag = 0").Order("create_date desc").Find(&depts).Error
	if err != nil {
		return nil, err
	}
	
	return s.buildDeptTree(depts), nil
}

func (s *DeptService) CreateDept(req models.CreateDeptRequest) error {
	var count int64
	database.DB.Model(&models.SysDept{}).Where("dept_title = ? AND del_flag = 0", req.DeptTitle).Count(&count)
	if count > 0 {
		return errors.New("部门名称已存在")
	}
	
	now := time.Now()
	dept := models.SysDept{
		DeptID:     utils.GenerateUUID(),
		Pid:        req.Pid,
		DeptTitle:  &req.DeptTitle,
		Status:     req.Status,
		Remark:     req.Remark,
		CreateDate: &now,
		UpdateDate: &now,
		DelFlag:    0,
	}
	
	return database.DB.Create(&dept).Error
}

func (s *DeptService) GetDeptDetail(deptId string) (*models.DeptTreeResponse, error) {
	var dept models.SysDept
	err := database.DB.Where("dept_id = ? AND del_flag = 0", deptId).First(&dept).Error
	if err != nil {
		return nil, errors.New("部门不存在")
	}
	
	return s.sysDeptToResponse(dept, []*models.DeptTreeResponse{}), nil
}

func (s *DeptService) UpdateDept(deptId string, req models.UpdateDeptRequest) error {
	var dept models.SysDept
	err := database.DB.Where("dept_id = ? AND del_flag = 0", deptId).First(&dept).Error
	if err != nil {
		return errors.New("部门不存在")
	}
	
	var count int64
	database.DB.Model(&models.SysDept{}).Where("dept_title = ? AND del_flag = 0 AND dept_id != ?", req.DeptTitle, deptId).Count(&count)
	if count > 0 {
		return errors.New("部门名称已存在")
	}
	
	now := time.Now()
	dept.Pid = req.Pid
	dept.DeptTitle = &req.DeptTitle
	dept.Status = req.Status
	dept.Remark = req.Remark
	dept.UpdateDate = &now
	
	return database.DB.Save(&dept).Error
}

func (s *DeptService) DeleteDepts(deptIds []string) error {
	for _, deptId := range deptIds {
		var childrenCount int64
		database.DB.Model(&models.SysDept{}).Where("pid = ? AND del_flag = 0", deptId).Count(&childrenCount)
		if childrenCount > 0 {
			return errors.New("部门存在子部门，不能删除")
		}
		
		var userCount int64
		database.DB.Model(&models.SysUserDept{}).Where("dept_id = ? AND del_flag = 0", deptId).Count(&userCount)
		if userCount > 0 {
			return errors.New("部门已被用户关联，不能删除")
		}
		
		var dept models.SysDept
		err := database.DB.Where("dept_id = ? AND del_flag = 0", deptId).First(&dept).Error
		if err != nil {
			continue
		}
		
		database.DB.Where("dept_id = ?", deptId).Delete(&models.SysUserDept{})
		
		now := time.Now()
		dept.DelFlag = 1
		dept.UpdateDate = &now
		database.DB.Save(&dept)
	}
	
	return nil
}

func (s *DeptService) buildDeptTree(depts []models.SysDept) []*models.DeptTreeResponse {
	deptMap := make(map[string]*models.DeptTreeResponse)
	var roots []*models.DeptTreeResponse
	
	for _, dept := range depts {
		deptMap[dept.DeptID] = s.sysDeptToResponse(dept, []*models.DeptTreeResponse{})
	}
	
	for _, dept := range depts {
		if dept.Pid == nil || *dept.Pid == "" {
			roots = append(roots, deptMap[dept.DeptID])
		} else {
			if parent, exists := deptMap[*dept.Pid]; exists {
				parent.Children = append(parent.Children, deptMap[dept.DeptID])
			}
		}
	}
	
	return roots
}

func (s *DeptService) sysDeptToResponse(dept models.SysDept, children []*models.DeptTreeResponse) *models.DeptTreeResponse {
	deptTitle := ""
	if dept.DeptTitle != nil {
		deptTitle = *dept.DeptTitle
	}
	
	return &models.DeptTreeResponse{
		DeptId:     dept.DeptID,
		Pid:        dept.Pid,
		DeptTitle:  deptTitle,
		Status:     dept.Status,
		CreateDate: models.TimeToStringPtr(dept.CreateDate),
		Remark:     dept.Remark,
		Children:   children,
	}
}
package datapermission

import (
	"context"

	"gorm.io/gorm"
)

type GormAssignmentStore struct {
	db *gorm.DB
}

func NewGormAssignmentStore(db *gorm.DB) *GormAssignmentStore {
	return &GormAssignmentStore{db: db}
}

func (s *GormAssignmentStore) Load(ctx context.Context, userID string) (AssignmentSnapshot, error) {
	snapshot := AssignmentSnapshot{
		CustomDepartmentIDs: make(map[string][]string),
	}

	var user struct {
		Status  int `gorm:"column:status"`
		DelFlag int `gorm:"column:del_flag"`
		IsSys   int `gorm:"column:is_sys"`
	}
	if err := s.db.WithContext(ctx).
		Table("sys_user").
		Select("status", "del_flag", "is_sys").
		Where("user_id = ?", userID).
		Take(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return snapshot, nil
		}
		return snapshot, err
	}
	snapshot.UserActive = user.Status == 1 && user.DelFlag == 0
	snapshot.SystemUser = user.IsSys == 1
	if !snapshot.UserActive || snapshot.SystemUser {
		return snapshot, nil
	}

	if err := s.db.WithContext(ctx).
		Table("sys_user_role AS user_role").
		Select("role.role_id", "role.data_scope").
		Joins("JOIN sys_role AS role ON role.role_id = user_role.role_id AND role.status = 1 AND role.del_flag = 0").
		Where("user_role.user_id = ? AND user_role.del_flag = 0", userID).
		Scan(&snapshot.Roles).Error; err != nil {
		return snapshot, err
	}

	needsUserDepartments := false
	needsDepartmentTree := false
	roleIDs := make([]string, 0, len(snapshot.Roles))
	for _, role := range snapshot.Roles {
		switch role.Scope {
		case ScopeAll:
			return snapshot, nil
		case ScopeDepartment:
			needsUserDepartments = true
		case ScopeDepartmentAndChildren:
			needsUserDepartments = true
			needsDepartmentTree = true
		case ScopeCustomDepartment:
			roleIDs = append(roleIDs, role.RoleID)
		}
	}
	if needsUserDepartments {
		if err := s.db.WithContext(ctx).
			Table("sys_user_dept AS user_dept").
			Select("user_dept.dept_id").
			Joins("JOIN sys_dept AS dept ON dept.dept_id = user_dept.dept_id AND dept.status = 1 AND dept.del_flag = 0").
			Where("user_dept.user_id = ? AND user_dept.del_flag = 0", userID).
			Pluck("user_dept.dept_id", &snapshot.UserDepartmentIDs).Error; err != nil {
			return snapshot, err
		}
	}

	if len(roleIDs) > 0 {
		var roleDepartments []struct {
			RoleID string `gorm:"column:role_id"`
			DeptID string `gorm:"column:dept_id"`
		}
		if err := s.db.WithContext(ctx).
			Table("sys_role_dept AS role_dept").
			Select("role_dept.role_id", "role_dept.dept_id").
			Joins("JOIN sys_dept AS dept ON dept.dept_id = role_dept.dept_id AND dept.status = 1 AND dept.del_flag = 0").
			Where("role_dept.role_id IN ?", roleIDs).
			Scan(&roleDepartments).Error; err != nil {
			return snapshot, err
		}
		for _, item := range roleDepartments {
			snapshot.CustomDepartmentIDs[item.RoleID] = append(snapshot.CustomDepartmentIDs[item.RoleID], item.DeptID)
		}
	}

	if needsDepartmentTree {
		var departments []struct {
			ID       string  `gorm:"column:dept_id"`
			ParentID *string `gorm:"column:pid"`
		}
		if err := s.db.WithContext(ctx).
			Table("sys_dept").
			Select("dept_id", "pid").
			Where("status = 1 AND del_flag = 0").
			Scan(&departments).Error; err != nil {
			return snapshot, err
		}
		snapshot.Departments = make([]Department, 0, len(departments))
		for _, department := range departments {
			snapshot.Departments = append(snapshot.Departments, Department{ID: department.ID, ParentID: department.ParentID})
		}
	}

	return snapshot, nil
}

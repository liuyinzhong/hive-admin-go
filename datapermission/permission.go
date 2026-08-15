package datapermission

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Scope string

const (
	ScopeAll                   Scope = "all"
	ScopeCustomDepartment      Scope = "customDepartment"
	ScopeDepartment            Scope = "department"
	ScopeDepartmentAndChildren Scope = "departmentAndChildren"
	ScopeSelf                  Scope = "self"
	ScopeNone                  Scope = "none"
)

var (
	ErrUserUnavailable = errors.New("data permission user is unavailable")
	ownerColumnPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)
)

type RoleAssignment struct {
	RoleID string `gorm:"column:role_id"`
	Scope  Scope  `gorm:"column:data_scope"`
}

type Department struct {
	ID       string
	ParentID *string
}

type AssignmentSnapshot struct {
	UserActive          bool
	SystemUser          bool
	Roles               []RoleAssignment
	UserDepartmentIDs   []string
	CustomDepartmentIDs map[string][]string
	Departments         []Department
}

type AssignmentStore interface {
	Load(ctx context.Context, userID string) (AssignmentSnapshot, error)
}

type Resolver struct {
	store AssignmentStore
}

func NewResolver(store AssignmentStore) *Resolver {
	return &Resolver{store: store}
}

type Permission struct {
	UserID        string
	All           bool
	IncludeSelf   bool
	DepartmentIDs []string
}

func (p Permission) AllowsDepartments(departmentIDs []string) bool {
	if p.All {
		return true
	}
	allowed := make(map[string]struct{}, len(p.DepartmentIDs))
	for _, departmentID := range p.DepartmentIDs {
		allowed[departmentID] = struct{}{}
	}
	for _, departmentID := range departmentIDs {
		if departmentID == "" {
			continue
		}
		if _, exists := allowed[departmentID]; !exists {
			return false
		}
	}
	return true
}

func (r *Resolver) Resolve(ctx context.Context, userID string) (Permission, error) {
	if userID == "" {
		return Permission{}, ErrUserUnavailable
	}
	snapshot, err := r.store.Load(ctx, userID)
	if err != nil {
		return Permission{}, err
	}
	if !snapshot.UserActive {
		return Permission{}, ErrUserUnavailable
	}

	permission := Permission{UserID: userID}
	if snapshot.SystemUser {
		permission.All = true
		return permission, nil
	}

	departmentIDs := make(map[string]struct{})
	childrenByParent := make(map[string][]string)
	for _, department := range snapshot.Departments {
		if department.ID == "" {
			continue
		}
		if department.ParentID != nil && *department.ParentID != "" {
			childrenByParent[*department.ParentID] = append(childrenByParent[*department.ParentID], department.ID)
		}
	}

	for _, role := range snapshot.Roles {
		switch role.Scope {
		case ScopeAll:
			permission.All = true
			permission.DepartmentIDs = nil
			return permission, nil
		case ScopeSelf:
			permission.IncludeSelf = true
		case ScopeDepartment:
			addDepartmentIDs(departmentIDs, snapshot.UserDepartmentIDs)
		case ScopeDepartmentAndChildren:
			for _, departmentID := range snapshot.UserDepartmentIDs {
				addDepartmentTree(departmentIDs, childrenByParent, departmentID)
			}
		case ScopeCustomDepartment:
			addDepartmentIDs(departmentIDs, snapshot.CustomDepartmentIDs[role.RoleID])
		case ScopeNone:
			// A role with no data scope contributes nothing to the union.
		default:
			// Unknown persisted values fail closed.
		}
	}

	permission.DepartmentIDs = make([]string, 0, len(departmentIDs))
	for departmentID := range departmentIDs {
		permission.DepartmentIDs = append(permission.DepartmentIDs, departmentID)
	}
	sort.Strings(permission.DepartmentIDs)
	return permission, nil
}

func (p Permission) Apply(query *gorm.DB, ownerColumns ...string) *gorm.DB {
	return p.apply(query, ownerColumns, nil)
}

func (p Permission) ApplyWithCSVUsers(query *gorm.DB, ownerColumns, csvUserColumns []string) *gorm.DB {
	return p.apply(query, ownerColumns, csvUserColumns)
}

func (p Permission) apply(query *gorm.DB, ownerColumns, csvUserColumns []string) *gorm.DB {
	if query == nil || p.All {
		return query
	}
	if len(ownerColumns) == 0 && len(csvUserColumns) == 0 {
		return query.Where("1 = 0")
	}
	for _, column := range append(append([]string{}, ownerColumns...), csvUserColumns...) {
		if !ownerColumnPattern.MatchString(column) {
			return query.Where("1 = 0")
		}
	}

	conditions := make([]string, 0, len(ownerColumns)*2)
	args := make([]interface{}, 0, len(ownerColumns)*2)
	if p.IncludeSelf && p.UserID != "" {
		for _, column := range ownerColumns {
			conditions = append(conditions, column+" = ?")
			args = append(args, p.UserID)
		}
		for _, column := range csvUserColumns {
			conditions = append(conditions, "FIND_IN_SET(?, "+column+") > 0")
			args = append(args, p.UserID)
		}
	}
	if len(p.DepartmentIDs) > 0 {
		for _, column := range ownerColumns {
			conditions = append(conditions, "EXISTS (SELECT 1 FROM sys_user_dept AS data_permission_user_dept WHERE data_permission_user_dept.user_id = "+column+" AND data_permission_user_dept.dept_id IN ? AND data_permission_user_dept.del_flag = 0)")
			args = append(args, p.DepartmentIDs)
		}
		for _, column := range csvUserColumns {
			conditions = append(conditions, "EXISTS (SELECT 1 FROM sys_user_dept AS data_permission_user_dept WHERE FIND_IN_SET(data_permission_user_dept.user_id, "+column+") > 0 AND data_permission_user_dept.dept_id IN ? AND data_permission_user_dept.del_flag = 0)")
			args = append(args, p.DepartmentIDs)
		}
	}
	if len(conditions) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func addDepartmentIDs(target map[string]struct{}, departmentIDs []string) {
	for _, departmentID := range departmentIDs {
		if departmentID != "" {
			target[departmentID] = struct{}{}
		}
	}
}

func addDepartmentTree(target map[string]struct{}, childrenByParent map[string][]string, rootID string) {
	if rootID == "" {
		return
	}
	queue := []string{rootID}
	for len(queue) > 0 {
		departmentID := queue[0]
		queue = queue[1:]
		if _, exists := target[departmentID]; exists {
			continue
		}
		target[departmentID] = struct{}{}
		queue = append(queue, childrenByParent[departmentID]...)
	}
}

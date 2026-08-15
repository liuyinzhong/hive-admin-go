package datapermission

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type memoryAssignmentStore struct {
	snapshot AssignmentSnapshot
	err      error
}

func (s memoryAssignmentStore) Load(_ context.Context, _ string) (AssignmentSnapshot, error) {
	return s.snapshot, s.err
}

func TestResolveSystemUserHasAllData(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{
		UserActive: true,
		SystemUser: true,
	}})

	permission, err := resolver.Resolve(context.Background(), "system-user")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !permission.All || permission.UserID != "system-user" {
		t.Fatalf("Resolve() = %#v, want unrestricted system permission", permission)
	}
}

func TestResolveCombinesRoleScopes(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{
		UserActive:        true,
		UserDepartmentIDs: []string{"dept-a"},
		Roles: []RoleAssignment{
			{RoleID: "role-self", Scope: ScopeSelf},
			{RoleID: "role-tree", Scope: ScopeDepartmentAndChildren},
			{RoleID: "role-custom", Scope: ScopeCustomDepartment},
		},
		CustomDepartmentIDs: map[string][]string{
			"role-custom": {"dept-c"},
		},
		Departments: []Department{
			{ID: "dept-a"},
			{ID: "dept-b", ParentID: stringPointer("dept-a")},
			{ID: "dept-c"},
			{ID: "dept-d", ParentID: stringPointer("dept-c")},
		},
	}})

	permission, err := resolver.Resolve(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if permission.All || !permission.IncludeSelf {
		t.Fatalf("Resolve() = %#v, want scoped permission including self", permission)
	}
	wantDepartments := []string{"dept-a", "dept-b", "dept-c"}
	if !reflect.DeepEqual(permission.DepartmentIDs, wantDepartments) {
		t.Fatalf("DepartmentIDs = %#v, want %#v", permission.DepartmentIDs, wantDepartments)
	}
}

func TestResolveRoleWithAllScopeWins(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{
		UserActive: true,
		Roles: []RoleAssignment{
			{RoleID: "role-none", Scope: ScopeNone},
			{RoleID: "role-all", Scope: ScopeAll},
		},
	}})

	permission, err := resolver.Resolve(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !permission.All {
		t.Fatalf("Resolve() = %#v, want all data", permission)
	}
}

func TestResolveNoRoleFailsClosed(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{UserActive: true}})

	permission, err := resolver.Resolve(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if permission.All || permission.IncludeSelf || len(permission.DepartmentIDs) != 0 {
		t.Fatalf("Resolve() = %#v, want no visible data", permission)
	}
}

func TestResolveNoneAndUnknownScopesFailClosed(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{
		UserActive: true,
		Roles: []RoleAssignment{
			{RoleID: "role-none", Scope: ScopeNone},
			{RoleID: "role-unknown", Scope: Scope("unexpected")},
		},
	}})

	permission, err := resolver.Resolve(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if permission.All || permission.IncludeSelf || len(permission.DepartmentIDs) != 0 {
		t.Fatalf("Resolve() = %#v, want no visible data", permission)
	}
}

func TestResolveRejectsUnavailableUser(t *testing.T) {
	resolver := NewResolver(memoryAssignmentStore{snapshot: AssignmentSnapshot{}})

	_, err := resolver.Resolve(context.Background(), "disabled-user")
	if err != ErrUserUnavailable {
		t.Fatalf("Resolve() error = %v, want %v", err, ErrUserUnavailable)
	}
}

func TestApplyRestrictsEveryOwnerColumn(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:pass@tcp(127.0.0.1:3306)/hive?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	permission := Permission{
		UserID:        "user-1",
		IncludeSelf:   true,
		DepartmentIDs: []string{"dept-a", "dept-b"},
	}
	query := permission.Apply(db.Table("dev_task"), "dev_task.creator_id", "dev_task.user_id")
	statement := query.Find(&[]struct{}{}).Statement
	if statement.Error != nil {
		t.Fatalf("Find() statement error = %v", statement.Error)
	}

	sql := statement.SQL.String()
	for _, fragment := range []string{
		"dev_task.creator_id = ?",
		"dev_task.user_id = ?",
		"FROM sys_user_dept AS data_permission_user_dept",
		"data_permission_user_dept.dept_id IN (?,?)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("SQL %q does not contain %q", sql, fragment)
		}
	}
}

func TestApplyInvalidOwnerColumnFailsClosed(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:pass@tcp(127.0.0.1:3306)/hive?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	query := (Permission{UserID: "user-1", IncludeSelf: true}).Apply(
		db.Table("dev_task"),
		"creator_id OR 1=1",
	)
	statement := query.Find(&[]struct{}{}).Statement
	if !strings.Contains(statement.SQL.String(), "1 = 0") {
		t.Fatalf("SQL %q does not fail closed", statement.SQL.String())
	}
}

func TestApplyCSVIncludesParticipantUsers(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:pass@tcp(127.0.0.1:3306)/hive?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	permission := Permission{UserID: "user-1", IncludeSelf: true, DepartmentIDs: []string{"dept-a"}}
	query := permission.ApplyWithCSVUsers(db.Table("dev_story"), []string{"dev_story.creator_id"}, []string{"dev_story.user_ids"})
	statement := query.Find(&[]struct{}{}).Statement
	sql := statement.SQL.String()
	for _, fragment := range []string{
		"dev_story.creator_id = ?",
		"FIND_IN_SET(?, dev_story.user_ids) > 0",
		"FIND_IN_SET(data_permission_user_dept.user_id, dev_story.user_ids) > 0",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("SQL %q does not contain %q", sql, fragment)
		}
	}
}

func TestAllowsDepartmentsRequiresEveryDepartment(t *testing.T) {
	permission := Permission{DepartmentIDs: []string{"dept-a", "dept-b"}}
	if !permission.AllowsDepartments([]string{"dept-a", "dept-b"}) {
		t.Fatal("AllowsDepartments() rejected an allowed department set")
	}
	if permission.AllowsDepartments([]string{"dept-a", "dept-c"}) {
		t.Fatal("AllowsDepartments() accepted a department outside the resolved scope")
	}
}

func stringPointer(value string) *string {
	return &value
}

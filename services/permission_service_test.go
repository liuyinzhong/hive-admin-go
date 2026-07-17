package services

import (
	"errors"
	"reflect"
	"testing"
)

type stubPermissionSource struct {
	isSystemUser bool
	bundles      []string
	err          error
}

func (s stubPermissionSource) LoadUserPermissionBundles(string) (bool, []string, error) {
	return s.isSystemUser, s.bundles, s.err
}

func TestNormalizePermissionBundle(t *testing.T) {
	got, err := NormalizePermissionBundle(" dev:story:batchCreate,dev:story:transition,dev:story:batchCreate ")
	if err != nil {
		t.Fatalf("NormalizePermissionBundle() error = %v", err)
	}
	if want := "dev:story:batchCreate,dev:story:transition"; got != want {
		t.Fatalf("NormalizePermissionBundle() = %q, want %q", got, want)
	}
}

func TestNormalizePermissionBundleRejectsInvalidCode(t *testing.T) {
	for _, value := range []string{
		"dev:story",
		"Dev:story:update",
		"dev:story:batch-create",
		"dev:story:update,",
		"dev::update",
	} {
		if _, err := NormalizePermissionBundle(value); !errors.Is(err, ErrInvalidPermissionCode) {
			t.Fatalf("NormalizePermissionBundle(%q) error = %v, want ErrInvalidPermissionCode", value, err)
		}
	}
}

func TestPermissionServiceCombinesAllRolePermissionBundles(t *testing.T) {
	service := newPermissionService(stubPermissionSource{bundles: []string{
		"medical:doctor:list,medical:doctor:update",
		"medical:doctor:list,medical:doctor:editorView",
	}})

	got, err := service.GetUserCodes("user-1")
	if err != nil {
		t.Fatalf("GetUserCodes() error = %v", err)
	}
	want := []string{
		"medical:doctor:editorView",
		"medical:doctor:list",
		"medical:doctor:update",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetUserCodes() = %#v, want %#v", got, want)
	}
	if !service.HasCode("user-1", "medical:doctor:update") {
		t.Fatal("HasCode() = false, want true")
	}
	if service.HasCode("user-1", "medical:doctor:delete") {
		t.Fatal("HasCode() = true for an ungranted permission")
	}
}

func TestPermissionServiceAllowsSystemUser(t *testing.T) {
	service := newPermissionService(stubPermissionSource{isSystemUser: true})
	if !service.HasCode("system-user", "workflow:definition:publish") {
		t.Fatal("HasCode() = false for a system user")
	}
}

func TestPermissionServiceAllowsSystemUserWhenBundlesCannotBeLoaded(t *testing.T) {
	service := newPermissionService(stubPermissionSource{
		isSystemUser: true,
		err:          errors.New("permission bundles unavailable"),
	})
	if !service.HasCode("system-user", "workflow:definition:publish") {
		t.Fatal("HasCode() = false for a system user when bundles cannot be loaded")
	}
}

func TestPermissionServiceFailsClosedWhenPermissionsCannotBeLoaded(t *testing.T) {
	service := newPermissionService(stubPermissionSource{err: errors.New("database unavailable")})
	if service.HasCode("user-1", "medical:doctor:update") {
		t.Fatal("HasCode() = true when permissions cannot be loaded")
	}
}

func TestFindDuplicatePermissionCodeComparesAtomicCodes(t *testing.T) {
	got := findDuplicatePermissionCode(
		"dev:story:transition,dev:story:update",
		[]string{"dev:story:update,dev:story:editorView", "medical:doctor:list"},
	)
	if want := "dev:story:update"; got != want {
		t.Fatalf("findDuplicatePermissionCode() = %q, want %q", got, want)
	}
}

func TestFindDuplicatePermissionCodeAllowsDistinctCodes(t *testing.T) {
	got := findDuplicatePermissionCode(
		"dev:story:transition,dev:story:update",
		[]string{"dev:story:editorView", "medical:doctor:list"},
	)
	if got != "" {
		t.Fatalf("findDuplicatePermissionCode() = %q, want empty", got)
	}
}

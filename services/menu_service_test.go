package services

import (
	"errors"
	"testing"
)

func TestNormalizeMenuNameClearsButtonName(t *testing.T) {
	raw := "RoutePermissionDevTaskCreate"
	name, err := normalizeMenuName("button", &raw)
	if err != nil {
		t.Fatalf("normalizeMenuName() error = %v", err)
	}
	if name != nil {
		t.Fatalf("normalizeMenuName() = %q, want nil", *name)
	}
}

func TestNormalizeMenuNameAllowsCatalogNameToBeEmpty(t *testing.T) {
	name, err := normalizeMenuName("catalog", nil)
	if err != nil {
		t.Fatalf("normalizeMenuName() error = %v", err)
	}
	if name != nil {
		t.Fatalf("normalizeMenuName() = %q, want nil", *name)
	}
}

func TestNormalizeMenuNameAllowsEmbeddedAndLinkNameToBeEmpty(t *testing.T) {
	for _, menuType := range []string{"embedded", "link"} {
		t.Run(menuType, func(t *testing.T) {
			name, err := normalizeMenuName(menuType, nil)
			if err != nil {
				t.Fatalf("normalizeMenuName() error = %v", err)
			}
			if name != nil {
				t.Fatalf("normalizeMenuName() = %q, want nil", *name)
			}
		})
	}
}

func TestNormalizeMenuNameRequiresMenuName(t *testing.T) {
	if _, err := normalizeMenuName("menu", nil); !errors.Is(err, ErrMenuNameRequired) {
		t.Fatalf("normalizeMenuName() error = %v, want ErrMenuNameRequired", err)
	}
}

func TestNormalizeMenuNameTrimsNonButtonName(t *testing.T) {
	raw := "  SystemMenu  "
	name, err := normalizeMenuName("menu", &raw)
	if err != nil {
		t.Fatalf("normalizeMenuName() error = %v", err)
	}
	if name == nil || *name != "SystemMenu" {
		t.Fatalf("normalizeMenuName() = %v, want SystemMenu", name)
	}
}

func TestNormalizeMenuRouteIdentityGeneratesEmbeddedIdentity(t *testing.T) {
	rawName := "UserInputName"
	rawPath := "/user/input/path"
	name, path, err := normalizeMenuRouteIdentity("embedded", "menu-id", &rawName, &rawPath)
	if err != nil {
		t.Fatalf("normalizeMenuRouteIdentity() error = %v", err)
	}
	if name == nil || *name != "embedded_menu-id" {
		t.Fatalf("name = %v, want embedded_menu-id", name)
	}
	if path == nil || *path != "/embedded/menu-id" {
		t.Fatalf("path = %v, want /embedded/menu-id", path)
	}
}

func TestNormalizeMenuRouteIdentityGeneratesLinkIdentity(t *testing.T) {
	rawName := "UserInputName"
	rawPath := "/user/input/path"
	name, path, err := normalizeMenuRouteIdentity("link", "menu-id", &rawName, &rawPath)
	if err != nil {
		t.Fatalf("normalizeMenuRouteIdentity() error = %v", err)
	}
	if name == nil || *name != "link_menu-id" {
		t.Fatalf("name = %v, want link_menu-id", name)
	}
	if path == nil || *path != "/link/menu-id" {
		t.Fatalf("path = %v, want /link/menu-id", path)
	}
}

func TestNormalizeAndValidateAuthCodeClearsNonButtonAuthCode(t *testing.T) {
	service := &MenuService{}
	raw := "dev:project:home"

	for _, menuType := range []string{"catalog", "menu"} {
		t.Run(menuType, func(t *testing.T) {
			authCode, err := service.normalizeAndValidateAuthCode(nil, menuType, &raw, "")
			if err != nil {
				t.Fatalf("normalizeAndValidateAuthCode() error = %v", err)
			}
			if authCode != nil {
				t.Fatalf("normalizeAndValidateAuthCode() = %q, want nil", *authCode)
			}
		})
	}
}

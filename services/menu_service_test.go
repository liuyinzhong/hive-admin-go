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

func TestNormalizeMenuNameRequiresNonButtonName(t *testing.T) {
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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubPermissionChecker struct {
	allowed bool
}

func (s stubPermissionChecker) HasCode(string, string) bool {
	return s.allowed
}

func TestPermissionGuardAllowsGrantedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	guard := NewPermissionGuard(stubPermissionChecker{allowed: true})
	router.GET("/protected", withUserID("user-1"), guard.Require("medical:doctor:list"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestPermissionGuardRejectsUngrantedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	guard := NewPermissionGuard(stubPermissionChecker{allowed: false})
	router.GET("/protected", withUserID("user-1"), guard.Require("medical:doctor:list"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func withUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userId", userID)
		c.Next()
	}
}

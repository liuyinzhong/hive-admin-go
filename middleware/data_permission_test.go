package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"hive-admin-go/datapermission"
)

type stubDataPermissionResolver struct {
	permission datapermission.Permission
	err        error
}

func (s stubDataPermissionResolver) Resolve(_ context.Context, _ string) (datapermission.Permission, error) {
	return s.permission, s.err
}

func TestDataPermissionMiddlewareStoresResolvedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	want := datapermission.Permission{UserID: "user-1", IncludeSelf: true}
	router.Use(func(c *gin.Context) {
		c.Set("userId", "user-1")
		c.Next()
	})
	router.Use(DataPermissionMiddleware(stubDataPermissionResolver{permission: want}))
	router.GET("/", func(c *gin.Context) {
		got, ok := GetDataPermission(c)
		if !ok || got.UserID != want.UserID || !got.IncludeSelf {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestDataPermissionMiddlewareAbortsUnavailableUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userId", "disabled-user")
		c.Next()
	})
	router.Use(DataPermissionMiddleware(stubDataPermissionResolver{err: datapermission.ErrUserUnavailable}))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

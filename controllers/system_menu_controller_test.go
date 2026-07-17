package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hive-admin-go/services"

	"github.com/gin-gonic/gin"
)

func TestWriteMenuMutationErrorMapsInvalidPermissionCode(t *testing.T) {
	assertMenuMutationStatus(t, services.ErrInvalidPermissionCode, http.StatusBadRequest)
}

func TestWriteMenuMutationErrorMapsPermissionCodeConflict(t *testing.T) {
	assertMenuMutationStatus(t, services.ErrPermissionCodeConflict, http.StatusBadRequest)
}

func TestWriteMenuMutationErrorMapsRequiredMenuName(t *testing.T) {
	assertMenuMutationStatus(t, services.ErrMenuNameRequired, http.StatusBadRequest)
}

func TestWriteMenuMutationErrorHidesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeMenuMutationError(context, errors.New("database connection details"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if body := recorder.Body.String(); body == "" || strings.Contains(body, "database connection details") {
		t.Fatalf("response body exposes internal error: %s", body)
	}
}

func assertMenuMutationStatus(t *testing.T, err error, want int) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeMenuMutationError(context, err)

	if recorder.Code != want {
		t.Fatalf("status = %d, want %d", recorder.Code, want)
	}
}

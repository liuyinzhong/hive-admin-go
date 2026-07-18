package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

type auditRecorderStub struct {
	operations []models.OperationLogEntry
	logins     []models.LoginLogEntry
}

func (s *auditRecorderStub) RecordOperation(entry models.OperationLogEntry) error {
	s.operations = append(s.operations, entry)
	return nil
}

func (s *auditRecorderStub) RecordLogin(entry models.LoginLogEntry) error {
	s.logins = append(s.logins, entry)
	return nil
}

func TestAuditLogMiddlewareRecordsAuthenticatedMutationWithoutSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(AuditLogMiddleware(recorder))
	router.POST("/api/system/users", func(c *gin.Context) {
		c.Set("userId", "user-1")
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"accessToken": "response-secret", "id": "created-1"}})
	})

	request := httptest.NewRequest(http.MethodPost, "/api/system/users?source=admin&token=query-secret", strings.NewReader(`{"username":"alice","password":"request-secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if len(recorder.operations) != 1 {
		t.Fatalf("operation log count = %d, want 1", len(recorder.operations))
	}
	entry := recorder.operations[0]
	if entry.UserID != "user-1" || entry.RequestURL != "/api/system/users" {
		t.Fatalf("entry identity = %#v", entry)
	}
	combined := entry.QueryParams + entry.RequestBody + entry.ResponseBody
	for _, secret := range []string{"query-secret", "request-secret", "response-secret"} {
		if strings.Contains(combined, secret) {
			t.Fatalf("audit content contains secret %q: %s", secret, combined)
		}
	}
	if !strings.Contains(entry.RequestBody, "alice") || !strings.Contains(entry.ResponseBody, "created-1") {
		t.Fatalf("audit content lost non-sensitive values: %#v", entry)
	}
}

func TestAuditLogMiddlewareIgnoresGetRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(AuditLogMiddleware(recorder))
	router.GET("/api/system/users", func(c *gin.Context) {
		c.Set("userId", "user-1")
		c.JSON(http.StatusOK, gin.H{"items": []string{}})
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/system/users", nil))

	if len(recorder.operations) != 0 || len(recorder.logins) != 0 {
		t.Fatalf("GET request produced audit logs: %#v %#v", recorder.operations, recorder.logins)
	}
}

func TestAuditLogMiddlewareClassifiesLoginAndRemovesCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(AuditLogMiddleware(recorder))
	router.POST("/api/auth/login", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, gin.H{"accessToken": "response-secret"})
	})

	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"alice","password":"request-secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if len(recorder.logins) != 1 || len(recorder.operations) != 0 {
		t.Fatalf("login classification = %#v %#v", recorder.logins, recorder.operations)
	}
	entry := recorder.logins[0]
	if entry.Username != "alice" || entry.EventType != models.LoginLogTypeLogin {
		t.Fatalf("login entry = %#v", entry)
	}
	if strings.Contains(entry.ResponseBody, "response-secret") {
		t.Fatalf("login response contains token: %s", entry.ResponseBody)
	}
}

func TestAuditLogMiddlewareMarksTruncatedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(AuditLogMiddleware(recorder))
	router.POST("/api/dev/stories", func(c *gin.Context) {
		c.Set("userId", "user-1")
		_, _ = io.Copy(io.Discard, c.Request.Body)
		c.Status(http.StatusNoContent)
	})

	requestBody := `{"password":"truncated-secret","note":"` + strings.Repeat("中", maxAuditContentBytes)
	request := httptest.NewRequest(http.MethodPost, "/api/dev/stories", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "text/plain")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if len(recorder.operations) != 1 || !recorder.operations[0].RequestTruncated {
		t.Fatalf("truncation entry = %#v", recorder.operations)
	}
	if !utf8.ValidString(recorder.operations[0].RequestBody) {
		t.Fatalf("truncated body is not valid UTF-8")
	}
	if strings.Contains(recorder.operations[0].RequestBody, "truncated-secret") {
		t.Fatalf("truncated body contains password: %s", recorder.operations[0].RequestBody[:100])
	}
}

func TestAuditLogMiddlewareStoresBinaryMetadataOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(AuditLogMiddleware(recorder))
	router.POST("/api/system/upload", func(c *gin.Context) {
		c.Set("userId", "user-1")
		_, _ = io.Copy(io.Discard, c.Request.Body)
		c.Data(http.StatusOK, "application/octet-stream", []byte{0xff, 0x00, 0x01})
	})

	request := httptest.NewRequest(http.MethodPost, "/api/system/upload", bytes.NewReader([]byte{0xff, 0x00}))
	request.Header.Set("Content-Type", "application/octet-stream")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if len(recorder.operations) != 1 {
		t.Fatalf("operation log count = %d, want 1", len(recorder.operations))
	}
	entry := recorder.operations[0]
	for _, content := range []string{entry.RequestBody, entry.ResponseBody} {
		if !strings.Contains(content, `"contentType":"application/octet-stream"`) || !strings.Contains(content, `"size":`) {
			t.Fatalf("binary metadata = %q", content)
		}
	}
}

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"hive-admin-go/models"
)

const maxAuditContentBytes = 64 * 1024

var (
	auditJSONSecretPattern = regexp.MustCompile(`(?i)("(?:[^"\\]|\\.)*(?:password|token|authorization|cookie)(?:[^"\\]|\\.)*"\s*:\s*)"(?:[^"\\]|\\.)*"`)
	auditFormSecretPattern = regexp.MustCompile(`(?i)((?:password|token|authorization|cookie)[^=&\s]*=)[^&\s]*`)
	auditBearerPattern     = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._~+/=-]+`)
)

type AuditRecorder interface {
	RecordOperation(entry models.OperationLogEntry) error
	RecordLogin(entry models.LoginLogEntry) error
}

type captureReadCloser struct {
	io.ReadCloser
	buffer    bytes.Buffer
	size      int64
	truncated bool
}

func (r *captureReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.capture(p[:n])
	return n, err
}

func (r *captureReadCloser) capture(value []byte) {
	r.size += int64(len(value))
	remaining := maxAuditContentBytes - r.buffer.Len()
	if remaining <= 0 {
		if len(value) > 0 {
			r.truncated = true
		}
		return
	}
	if len(value) > remaining {
		r.buffer.Write(value[:remaining])
		r.truncated = true
		return
	}
	r.buffer.Write(value)
}

type captureResponseWriter struct {
	gin.ResponseWriter
	buffer    bytes.Buffer
	size      int64
	truncated bool
}

func (w *captureResponseWriter) Write(value []byte) (int, error) {
	w.capture(value)
	return w.ResponseWriter.Write(value)
}

func (w *captureResponseWriter) WriteString(value string) (int, error) {
	w.capture([]byte(value))
	return w.ResponseWriter.WriteString(value)
}

func (w *captureResponseWriter) capture(value []byte) {
	w.size += int64(len(value))
	remaining := maxAuditContentBytes - w.buffer.Len()
	if remaining <= 0 {
		if len(value) > 0 {
			w.truncated = true
		}
		return
	}
	if len(value) > remaining {
		w.buffer.Write(value[:remaining])
		w.truncated = true
		return
	}
	w.buffer.Write(value)
}

func AuditLogMiddleware(recorder AuditRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !shouldObserveAuditRequest(c.Request.Method) {
			c.Next()
			return
		}

		startedAt := time.Now()
		requestCapture := &captureReadCloser{ReadCloser: c.Request.Body}
		c.Request.Body = requestCapture
		responseCapture := &captureResponseWriter{ResponseWriter: c.Writer}
		c.Writer = responseCapture

		c.Next()

		requestBody, requestTruncated := auditRequestBody(c.Request, requestCapture)
		responseBody, responseTruncated := auditResponseBody(c.Writer.Header().Get("Content-Type"), responseCapture)
		queryParams, queryTruncated := sanitizeQuery(c.Request.URL.Query())
		common := struct {
			HTTPStatus  int
			DurationMs  int64
			IP          string
			UserAgent   string
			ContentType string
		}{
			HTTPStatus:  c.Writer.Status(),
			DurationMs:  time.Since(startedAt).Milliseconds(),
			IP:          truncateRunes(c.ClientIP(), 64),
			UserAgent:   truncateRunes(c.Request.UserAgent(), 512),
			ContentType: truncateRunes(c.ContentType(), 128),
		}

		if isLoginRequest(c.Request.Method, c.Request.URL.Path) || isLogoutRequest(c.Request.Method, c.Request.URL.Path) {
			eventType := models.LoginLogTypeLogin
			if isLogoutRequest(c.Request.Method, c.Request.URL.Path) {
				eventType = models.LoginLogTypeLogout
			}
			entry := models.LoginLogEntry{
				UserID: c.GetString("userId"), Username: extractUsername(requestBody), EventType: eventType,
				ResponseBody: responseBody, ResponseTruncated: responseTruncated,
				HTTPStatus: common.HTTPStatus, DurationMs: common.DurationMs,
				IP: common.IP, UserAgent: common.UserAgent, ContentType: common.ContentType,
			}
			if err := recorder.RecordLogin(entry); err != nil {
				log.Printf("登录日志写入失败: %v", err)
			}
			return
		}

		userID := c.GetString("userId")
		if userID == "" {
			return
		}
		entry := models.OperationLogEntry{
			UserID: userID, RequestMethod: c.Request.Method, RequestURL: truncateRunes(c.Request.URL.Path, 512),
			QueryParams: queryParams, QueryTruncated: queryTruncated,
			RequestBody: requestBody, ResponseBody: responseBody,
			RequestTruncated: requestTruncated, ResponseTruncated: responseTruncated,
			HTTPStatus: common.HTTPStatus, DurationMs: common.DurationMs, IP: common.IP,
			UserAgent: common.UserAgent, ContentType: common.ContentType,
		}
		if err := recorder.RecordOperation(entry); err != nil {
			log.Printf("操作日志写入失败: %v", err)
		}
	}
}

func shouldObserveAuditRequest(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func isLoginRequest(method, path string) bool {
	return method == http.MethodPost && path == "/api/auth/login"
}

func isLogoutRequest(method, path string) bool {
	return method == http.MethodPost && path == "/api/auth/logout"
}

func auditRequestBody(request *http.Request, capture *captureReadCloser) (string, bool) {
	contentType := request.Header.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		return multipartMetadata(request.MultipartForm), capture.truncated
	}
	if !isTextAuditContent(contentType, capture.buffer.Bytes()) {
		return auditBinaryMetadata(contentType, capture.size), capture.truncated
	}
	return sanitizeAuditContent(capture.buffer.Bytes(), contentType), capture.truncated
}

func auditResponseBody(contentType string, capture *captureResponseWriter) (string, bool) {
	if !isTextAuditContent(contentType, capture.buffer.Bytes()) {
		return auditBinaryMetadata(contentType, capture.size), capture.truncated
	}
	return sanitizeAuditContent(capture.buffer.Bytes(), contentType), capture.truncated
}

func isTextAuditContent(contentType string, content []byte) bool {
	lowerType := strings.ToLower(contentType)
	if lowerType == "" {
		return utf8.Valid(content)
	}
	return strings.Contains(lowerType, "json") || strings.Contains(lowerType, "xml") ||
		strings.Contains(lowerType, "x-www-form-urlencoded") || strings.HasPrefix(lowerType, "text/")
}

func auditBinaryMetadata(contentType string, size int64) string {
	metadata, _ := json.Marshal(map[string]interface{}{"contentType": contentType, "size": size})
	return string(metadata)
}

func sanitizeAuditContent(raw []byte, contentType string) string {
	if len(raw) == 0 {
		return ""
	}
	var value interface{}
	if json.Unmarshal(raw, &value) == nil {
		value = removeSensitiveValues(value)
		encoded, err := json.Marshal(value)
		if err == nil {
			return truncateBytes(string(encoded), maxAuditContentBytes)
		}
	}
	if strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(raw))
		if err == nil {
			sanitized, _ := sanitizeQuery(values)
			return sanitized
		}
	}
	return truncateBytes(sanitizeRawAuditContent(string(raw)), maxAuditContentBytes)
}

func removeSensitiveValues(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			if isSensitiveAuditKey(key) {
				delete(typed, key)
				continue
			}
			typed[key] = removeSensitiveValues(child)
		}
	case []interface{}:
		for index, child := range typed {
			typed[index] = removeSensitiveValues(child)
		}
	}
	return value
}

func isSensitiveAuditKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", ""), "_", ""))
	return strings.Contains(normalized, "password") || strings.Contains(normalized, "token") ||
		normalized == "authorization" || strings.Contains(normalized, "cookie")
}

func sanitizeQuery(values url.Values) (string, bool) {
	result := make(map[string]interface{})
	for key, items := range values {
		if isSensitiveAuditKey(key) {
			continue
		}
		if len(items) == 1 {
			result[key] = items[0]
		} else {
			result[key] = items
		}
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", false
	}
	return truncateBytes(string(encoded), maxAuditContentBytes), len(encoded) > maxAuditContentBytes
}

func sanitizeRawAuditContent(value string) string {
	value = auditJSONSecretPattern.ReplaceAllString(value, `${1}"[REDACTED]"`)
	value = auditFormSecretPattern.ReplaceAllString(value, `${1}[REDACTED]`)
	return auditBearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
}

func multipartMetadata(form *multipart.Form) string {
	if form == nil {
		return `{"type":"multipart/form-data"}`
	}
	metadata := make(map[string]interface{})
	for key, values := range form.Value {
		if !isSensitiveAuditKey(key) {
			metadata[key] = values
		}
	}
	files := make(map[string][]map[string]interface{})
	for field, headers := range form.File {
		items := make([]map[string]interface{}, 0, len(headers))
		for _, header := range headers {
			items = append(items, map[string]interface{}{
				"name": header.Filename, "size": header.Size, "contentType": header.Header.Get("Content-Type"),
			})
		}
		files[field] = items
	}
	metadata["files"] = files
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return `{"type":"multipart/form-data"}`
	}
	return truncateBytes(string(encoded), maxAuditContentBytes)
}

func extractUsername(body string) string {
	var value map[string]interface{}
	if json.Unmarshal([]byte(body), &value) != nil {
		return ""
	}
	username, _ := value["username"].(string)
	return truncateRunes(username, 36)
}

func truncateBytes(value string, limit int) string {
	if len(value) > limit {
		value = value[:limit]
	}
	for !utf8.ValidString(value) && len(value) > 0 {
		value = value[:len(value)-1]
	}
	return value
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

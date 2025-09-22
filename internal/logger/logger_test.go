package logger

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := New(LogLevelInfo)
	if logger == nil {
		t.Error("Expected logger to be created")
		return
	}
	if logger.level != LogLevelInfo {
		t.Errorf("Expected log level %d, got %d", LogLevelInfo, logger.level)
	}
}

func TestNewWithWriters(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelDebug, &infoBuf, &errBuf)

	if logger == nil {
		t.Error("Expected logger to be created")
		return
	}
	if logger.level != LogLevelDebug {
		t.Errorf("Expected log level %d, got %d", LogLevelDebug, logger.level)
	}
}

func TestLogLevels(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelWarn, &infoBuf, &errBuf)

	// Test that debug and info messages are filtered out
	logger.LogDebug("debug message")
	logger.LogInfo("info message")

	if infoBuf.Len() > 0 {
		t.Error("Expected no output for debug/info messages at warn level")
	}

	// Test that warn and error messages are logged
	logger.LogWarn("warn message")
	logger.LogError(errors.New("test error"), "test context")

	infoOutput := infoBuf.String()
	errOutput := errBuf.String()

	if !strings.Contains(infoOutput, "warn message") {
		t.Error("Expected warn message to be logged")
	}
	if !strings.Contains(errOutput, "test error") {
		t.Error("Expected error message to be logged")
	}
}

func TestLogRequest(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelInfo, &infoBuf, &errBuf)

	req := httptest.NewRequest("GET", "/test/path?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"

	logger.LogRequest(req)

	output := infoBuf.String()
	if !strings.Contains(output, "GET") {
		t.Error("Expected method to be logged")
	}
	if !strings.Contains(output, "/test/path") {
		t.Error("Expected path to be logged")
	}
	if !strings.Contains(output, "127.0.0.1:12345") {
		t.Error("Expected remote address to be logged")
	}
}

func TestLogRequestWithBody(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelDebug, &infoBuf, &errBuf)

	body := strings.NewReader(`{"key": "value"}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")

	logger.LogRequest(req)

	output := infoBuf.String()
	if !strings.Contains(output, "POST") {
		t.Error("Expected POST method to be logged")
	}
	// In debug mode, should log request details
	if !strings.Contains(output, "Request Details") {
		t.Error("Expected request details in debug mode")
	}
}

func TestLogResponse(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelInfo, &infoBuf, &errBuf)

	req := httptest.NewRequest("GET", "/test", nil)
	responseBody := []byte(`{"status": "success"}`)
	duration := 50 * time.Millisecond

	logger.LogResponse(req, 200, responseBody, duration)

	output := infoBuf.String()
	if !strings.Contains(output, "200") {
		t.Error("Expected status code to be logged")
	}
	if !strings.Contains(output, "GET") {
		t.Error("Expected method to be logged")
	}
	if !strings.Contains(output, "/test") {
		t.Error("Expected path to be logged")
	}
	if !strings.Contains(output, "bytes") {
		t.Error("Expected response size to be logged")
	}
}

func TestLogError(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelError, &infoBuf, &errBuf)

	err := errors.New("test error")
	logger.LogError(err, "test context")

	output := errBuf.String()
	if !strings.Contains(output, "test error") {
		t.Error("Expected error message to be logged")
	}
	if !strings.Contains(output, "test context") {
		t.Error("Expected context to be logged")
	}
}

func TestLogErrorWithRequest(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelError, &infoBuf, &errBuf)

	req := httptest.NewRequest("POST", "/api/test", nil)
	err := errors.New("request error")
	logger.LogErrorWithRequest(err, req, "handling request")

	output := errBuf.String()
	if !strings.Contains(output, "request error") {
		t.Error("Expected error message to be logged")
	}
	if !strings.Contains(output, "POST") {
		t.Error("Expected method to be logged")
	}
	if !strings.Contains(output, "/api/test") {
		t.Error("Expected path to be logged")
	}
	if !strings.Contains(output, "handling request") {
		t.Error("Expected context to be logged")
	}
}

func TestIsSensitiveHeader(t *testing.T) {
	logger := New(LogLevelDebug)

	tests := []struct {
		header    string
		sensitive bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"Cookie", true},
		{"X-API-Key", true},
		{"x-api-key", true},
		{"Content-Type", false},
		{"Accept", false},
		{"User-Agent", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			result := logger.isSensitiveHeader(tt.header)
			if result != tt.sensitive {
				t.Errorf("isSensitiveHeader(%s) = %v, want %v", tt.header, result, tt.sensitive)
			}
		})
	}
}

func TestIsTextContent(t *testing.T) {
	logger := New(LogLevelDebug)

	tests := []struct {
		contentType string
		isText      bool
	}{
		{"application/json", true},
		{"application/xml", true},
		{"text/plain", true},
		{"text/html", true},
		{"application/x-www-form-urlencoded", true},
		{"image/jpeg", false},
		{"application/octet-stream", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := logger.isTextContent(tt.contentType)
			if result != tt.isText {
				t.Errorf("isTextContent(%s) = %v, want %v", tt.contentType, result, tt.isText)
			}
		})
	}
}

func TestSetAndGetLogLevel(t *testing.T) {
	logger := New(LogLevelInfo)

	// Test initial level
	if logger.GetLogLevel() != LogLevelInfo {
		t.Errorf("Expected initial log level %d, got %d", LogLevelInfo, logger.GetLogLevel())
	}

	// Test setting new level
	logger.SetLogLevel(LogLevelDebug)
	if logger.GetLogLevel() != LogLevelDebug {
		t.Errorf("Expected log level %d after setting, got %d", LogLevelDebug, logger.GetLogLevel())
	}
}

func TestMiddleware(t *testing.T) {
	var infoBuf, errBuf bytes.Buffer
	logger := NewWithWriters(LogLevelInfo, &infoBuf, &errBuf)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Errorf("Failed to write test response: %v", err)
		}
	})

	// Wrap with logging middleware
	wrappedHandler := logger.Middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", w.Body.String())
	}

	// Verify logging
	output := infoBuf.String()
	if !strings.Contains(output, "Request:") {
		t.Error("Expected request to be logged")
	}
	if !strings.Contains(output, "Response:") {
		t.Error("Expected response to be logged")
	}
	if !strings.Contains(output, "GET") {
		t.Error("Expected method to be logged")
	}
	if !strings.Contains(output, "/test") {
		t.Error("Expected path to be logged")
	}
	if !strings.Contains(output, "200") {
		t.Error("Expected status code to be logged")
	}
}

func TestResponseWriterWrapper(t *testing.T) {
	var body bytes.Buffer
	wrapper := &responseWriterWrapper{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     200,
		body:           &body,
	}

	// Test WriteHeader
	wrapper.WriteHeader(201)
	if wrapper.statusCode != 201 {
		t.Errorf("Expected status code 201, got %d", wrapper.statusCode)
	}

	// Test Write
	testData := []byte("test data")
	n, err := wrapper.Write(testData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}
	if body.String() != "test data" {
		t.Errorf("Expected body 'test data', got '%s'", body.String())
	}
}

func TestCreateRequestLog(t *testing.T) {
	logger := New(LogLevelDebug)

	body := strings.NewReader(`{"test": "data"}`)
	req := httptest.NewRequest("POST", "/api/test?param=value", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "192.168.1.1:12345"

	reqLog := logger.createRequestLog(req)

	// Verify basic fields
	if reqLog.Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", reqLog.Method)
	}
	if reqLog.Path != "/api/test" {
		t.Errorf("Expected path '/api/test', got '%s'", reqLog.Path)
	}
	if reqLog.Query != "param=value" {
		t.Errorf("Expected query 'param=value', got '%s'", reqLog.Query)
	}
	if reqLog.RemoteAddr != "192.168.1.1:12345" {
		t.Errorf("Expected remote addr '192.168.1.1:12345', got '%s'", reqLog.RemoteAddr)
	}
	if reqLog.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", reqLog.UserAgent)
	}

	// Verify headers (sensitive ones should be redacted)
	if reqLog.Headers["Authorization"][0] != "[REDACTED]" {
		t.Error("Expected Authorization header to be redacted")
	}
	if reqLog.Headers["User-Agent"][0] != "test-agent" {
		t.Error("Expected User-Agent header to be preserved")
	}

	// Verify body is captured
	if !strings.Contains(reqLog.Body, "test") {
		t.Error("Expected request body to be captured")
	}
}

func TestCreateResponseLog(t *testing.T) {
	logger := New(LogLevelDebug)

	req := httptest.NewRequest("GET", "/api/test", nil)
	responseBody := []byte(`{"result": "success"}`)
	duration := 100 * time.Millisecond

	respLog := logger.createResponseLog(req, 200, responseBody, duration)

	// Verify fields
	if respLog.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", respLog.StatusCode)
	}
	if respLog.Size != len(responseBody) {
		t.Errorf("Expected size %d, got %d", len(responseBody), respLog.Size)
	}
	if respLog.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, respLog.Duration)
	}
	if !strings.Contains(respLog.Body, "success") {
		t.Error("Expected response body to be captured")
	}
}

func BenchmarkLogRequest(b *testing.B) {
	logger := New(LogLevelInfo)
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogRequest(req)
	}
}

func BenchmarkLogResponse(b *testing.B) {
	logger := New(LogLevelInfo)
	req := httptest.NewRequest("GET", "/test", nil)
	responseBody := []byte(`{"status": "ok"}`)
	duration := 50 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogResponse(req, 200, responseBody, duration)
	}
}

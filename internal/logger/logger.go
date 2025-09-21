package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// LogLevelDebug enables debug logging
	LogLevelDebug LogLevel = iota
	// LogLevelInfo enables info logging
	LogLevelInfo
	// LogLevelWarn enables warning logging
	LogLevelWarn
	// LogLevelError enables error logging
	LogLevelError
)

// Logger handles HTTP request/response logging
type Logger struct {
	level      LogLevel
	infoLogger *log.Logger
	errLogger  *log.Logger
}

// RequestLog represents a logged HTTP request
type RequestLog struct {
	Timestamp    time.Time              `json:"timestamp"`
	Method       string                 `json:"method"`
	URL          string                 `json:"url"`
	Path         string                 `json:"path"`
	Query        string                 `json:"query,omitempty"`
	Headers      map[string][]string    `json:"headers,omitempty"`
	Body         string                 `json:"body,omitempty"`
	RemoteAddr   string                 `json:"remote_addr"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	ResponseTime time.Duration          `json:"response_time,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	ResponseSize int                    `json:"response_size,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ResponseLog represents a logged HTTP response
type ResponseLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	StatusCode int                    `json:"status_code"`
	Headers    map[string][]string    `json:"headers,omitempty"`
	Body       string                 `json:"body,omitempty"`
	Size       int                    `json:"size"`
	Duration   time.Duration          `json:"duration"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// New creates a new logger instance
func New(level LogLevel) *Logger {
	return &Logger{
		level:      level,
		infoLogger: log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmicroseconds),
		errLogger:  log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// NewWithWriters creates a new logger with custom writers
func NewWithWriters(level LogLevel, infoWriter, errorWriter io.Writer) *Logger {
	return &Logger{
		level:      level,
		infoLogger: log.New(infoWriter, "[INFO] ", log.LstdFlags|log.Lmicroseconds),
		errLogger:  log.New(errorWriter, "[ERROR] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// LogRequest logs an HTTP request with detailed information
func (l *Logger) LogRequest(req *http.Request) {
	if l.level > LogLevelInfo {
		return
	}

	reqLog := l.createRequestLog(req)

	// Log basic request info
	l.infoLogger.Printf("Request: %s %s from %s",
		reqLog.Method, reqLog.Path, reqLog.RemoteAddr)

	// Log detailed request info in debug mode
	if l.level <= LogLevelDebug {
		l.logRequestDetails(reqLog)
	}
}

// LogResponse logs an HTTP response with detailed information
func (l *Logger) LogResponse(req *http.Request, statusCode int, responseBody []byte, duration time.Duration) {
	if l.level > LogLevelInfo {
		return
	}

	respLog := l.createResponseLog(req, statusCode, responseBody, duration)

	// Log basic response info
	l.infoLogger.Printf("Response: %s %s -> %d (%v, %d bytes)",
		req.Method, req.URL.Path, statusCode, duration, respLog.Size)

	// Log detailed response info in debug mode
	if l.level <= LogLevelDebug {
		l.logResponseDetails(respLog)
	}
}

// LogError logs an error with context
func (l *Logger) LogError(err error, context string) {
	if l.level > LogLevelError {
		return
	}

	l.errLogger.Printf("Error in %s: %v", context, err)
}

// LogErrorWithRequest logs an error with request context
func (l *Logger) LogErrorWithRequest(err error, req *http.Request, context string) {
	if l.level > LogLevelError {
		return
	}

	l.errLogger.Printf("Error in %s for %s %s: %v",
		context, req.Method, req.URL.Path, err)
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(message string, args ...interface{}) {
	if l.level > LogLevelInfo {
		return
	}

	l.infoLogger.Printf(message, args...)
}

// LogDebug logs a debug message
func (l *Logger) LogDebug(message string, args ...interface{}) {
	if l.level > LogLevelDebug {
		return
	}

	l.infoLogger.Printf("[DEBUG] "+message, args...)
}

// LogWarn logs a warning message
func (l *Logger) LogWarn(message string, args ...interface{}) {
	if l.level > LogLevelWarn {
		return
	}

	l.infoLogger.Printf("[WARN] "+message, args...)
}

// createRequestLog creates a RequestLog from an HTTP request
func (l *Logger) createRequestLog(req *http.Request) *RequestLog {
	reqLog := &RequestLog{
		Timestamp:  time.Now(),
		Method:     req.Method,
		URL:        req.URL.String(),
		Path:       req.URL.Path,
		Query:      req.URL.RawQuery,
		RemoteAddr: req.RemoteAddr,
		UserAgent:  req.UserAgent(),
		Metadata:   make(map[string]interface{}),
	}

	// Copy headers (excluding sensitive ones)
	if l.level <= LogLevelDebug {
		reqLog.Headers = make(map[string][]string)
		for name, values := range req.Header {
			// Skip sensitive headers
			if l.isSensitiveHeader(name) {
				reqLog.Headers[name] = []string{"[REDACTED]"}
			} else {
				reqLog.Headers[name] = values
			}
		}
	}

	// Read and log request body for POST/PUT requests
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil {
				// Restore the body for further processing
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Log body if it's not too large and is text-based
				if len(bodyBytes) < 10240 && l.isTextContent(req.Header.Get("Content-Type")) {
					reqLog.Body = string(bodyBytes)
				} else {
					reqLog.Body = fmt.Sprintf("[BODY: %d bytes, %s]",
						len(bodyBytes), req.Header.Get("Content-Type"))
				}
			}
		}
	}

	return reqLog
}

// createResponseLog creates a ResponseLog from response data
func (l *Logger) createResponseLog(req *http.Request, statusCode int, responseBody []byte, duration time.Duration) *ResponseLog {
	respLog := &ResponseLog{
		Timestamp:  time.Now(),
		StatusCode: statusCode,
		Size:       len(responseBody),
		Duration:   duration,
		Metadata:   make(map[string]interface{}),
	}

	// Log response body if it's not too large and in debug mode
	if l.level <= LogLevelDebug && len(responseBody) < 10240 {
		respLog.Body = string(responseBody)
	} else if len(responseBody) >= 10240 {
		respLog.Body = fmt.Sprintf("[LARGE RESPONSE: %d bytes]", len(responseBody))
	}

	return respLog
}

// logRequestDetails logs detailed request information
func (l *Logger) logRequestDetails(reqLog *RequestLog) {
	details, _ := json.MarshalIndent(reqLog, "", "  ")
	l.infoLogger.Printf("Request Details:\n%s", string(details))
}

// logResponseDetails logs detailed response information
func (l *Logger) logResponseDetails(respLog *ResponseLog) {
	details, _ := json.MarshalIndent(respLog, "", "  ")
	l.infoLogger.Printf("Response Details:\n%s", string(details))
}

// isSensitiveHeader checks if a header contains sensitive information
func (l *Logger) isSensitiveHeader(headerName string) bool {
	sensitiveHeaders := []string{
		"authorization", "cookie", "set-cookie", "x-api-key",
		"x-auth-token", "x-access-token", "x-csrf-token",
	}

	headerLower := strings.ToLower(headerName)
	for _, sensitive := range sensitiveHeaders {
		if headerLower == sensitive {
			return true
		}
	}
	return false
}

// isTextContent checks if content type is text-based
func (l *Logger) isTextContent(contentType string) bool {
	textTypes := []string{
		"text/", "application/json", "application/xml",
		"application/x-www-form-urlencoded",
	}

	contentTypeLower := strings.ToLower(contentType)
	for _, textType := range textTypes {
		if strings.HasPrefix(contentTypeLower, textType) {
			return true
		}
	}
	return false
}

// Middleware returns an HTTP middleware that logs requests and responses
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request
		l.LogRequest(r)

		// Create a response writer wrapper to capture response data
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     200,
			body:           &bytes.Buffer{},
		}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Log the response
		duration := time.Since(start)
		l.LogResponse(r, wrapper.statusCode, wrapper.body.Bytes(), duration)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture response data
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body and writes it to the original writer
func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	// Capture response body for logging (limit size to prevent memory issues)
	if w.body.Len() < 10240 {
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// SetLogLevel sets the logging level
func (l *Logger) SetLogLevel(level LogLevel) {
	l.level = level
}

// GetLogLevel returns the current logging level
func (l *Logger) GetLogLevel() LogLevel {
	return l.level
}

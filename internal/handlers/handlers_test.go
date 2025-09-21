package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/walterfan/lazy-mock-server/internal/config"
	"github.com/walterfan/lazy-mock-server/internal/logger"
)

func createTestHandler() (*MockHandler, *config.Manager) {
	configManager := config.NewManager("test.yaml")

	// Add test routes
	routes := []config.Route{
		{
			Path:        "/test/simple",
			Method:      "GET",
			StatusCode:  200,
			ContentType: "text/plain",
			Response:    "Hello World",
		},
		{
			Path:        "/test/json",
			Method:      "GET",
			StatusCode:  200,
			ContentType: "application/json",
			Response: map[string]interface{}{
				"message": "Test JSON response",
				"status":  "success",
			},
		},
		{
			Path:        "/test/error",
			Method:      "GET",
			StatusCode:  404,
			ContentType: "application/json",
			Response: map[string]interface{}{
				"error": "Not found",
				"code":  404,
			},
		},
		{
			Path:        "/test/wildcard/*",
			Method:      "GET",
			StatusCode:  200,
			ContentType: "application/json",
			Response:    map[string]string{"matched": "wildcard"},
		},
		{
			Path:       "/test/params",
			Method:     "GET",
			StatusCode: 200,
			Parameters: map[string]string{"type": "user"},
			Response:   map[string]string{"type": "user search"},
		},
	}

	testConfig := &config.Config{Routes: routes}
	configManager.SetConfig(testConfig)

	log := logger.New(logger.LogLevelError) // Use error level to reduce test output
	handler := NewMockHandler(configManager, log)

	return handler, configManager
}

func TestNewMockHandler(t *testing.T) {
	configManager := config.NewManager("test.yaml")
	log := logger.New(logger.LogLevelInfo)

	handler := NewMockHandler(configManager, log)
	if handler == nil {
		t.Error("Expected handler to be created")
	}
	if handler.configManager != configManager {
		t.Error("Expected config manager to be set")
	}
	if handler.logger != log {
		t.Error("Expected logger to be set")
	}
}

func TestHandleMockEndpoint(t *testing.T) {
	handler, _ := createTestHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		contentType    string
	}{
		{
			name:           "Simple text response",
			method:         "GET",
			path:           "/test/simple",
			expectedStatus: 200,
			expectedBody:   "Hello World",
			contentType:    "text/plain",
		},
		{
			name:           "JSON response",
			method:         "GET",
			path:           "/test/json",
			expectedStatus: 200,
			expectedBody:   `{"message":"Test JSON response","status":"success"}`,
			contentType:    "application/json",
		},
		{
			name:           "Error response",
			method:         "GET",
			path:           "/test/error",
			expectedStatus: 404,
			expectedBody:   `{"code":404,"error":"Not found"}`,
			contentType:    "application/json",
		},
		{
			name:           "Wildcard match",
			method:         "GET",
			path:           "/test/wildcard/anything",
			expectedStatus: 200,
			expectedBody:   `{"matched":"wildcard"}`,
			contentType:    "application/json",
		},
		{
			name:           "Parameter match",
			method:         "GET",
			path:           "/test/params?type=user",
			expectedStatus: 200,
			expectedBody:   `{"type":"user search"}`,
			contentType:    "application/json",
		},
		{
			name:           "Not found route",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: 404,
			expectedBody:   `{"error":"Route not found","method":"GET","path":"/nonexistent"}`,
			contentType:    "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.contentType {
				t.Errorf("Expected content type '%s', got '%s'", tt.contentType, contentType)
			}

			body := strings.TrimSpace(w.Body.String())
			if tt.contentType == "application/json" {
				// For JSON, compare the parsed objects to handle ordering
				var expected, actual interface{}
				if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
					t.Fatalf("Failed to parse expected JSON: %v", err)
				}
				if err := json.Unmarshal([]byte(body), &actual); err != nil {
					t.Fatalf("Failed to parse actual JSON: %v", err)
				}

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				if string(expectedJSON) != string(actualJSON) {
					t.Errorf("Expected JSON %s, got %s", expectedJSON, actualJSON)
				}
			} else {
				if body != tt.expectedBody {
					t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestHandleGetRoutes(t *testing.T) {
	handler, _ := createTestHandler()

	req := httptest.NewRequest("GET", "/_mock/routes", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if count, ok := response["count"].(float64); !ok || count != 5 {
		t.Errorf("Expected count 5, got %v", response["count"])
	}

	routes, ok := response["routes"].([]interface{})
	if !ok {
		t.Fatal("Expected routes to be an array")
	}
	if len(routes) != 5 {
		t.Errorf("Expected 5 routes, got %d", len(routes))
	}
}

func TestHandleAddRoute(t *testing.T) {
	handler, configManager := createTestHandler()

	newRoute := config.Route{
		Path:        "/test/new",
		Method:      "POST",
		StatusCode:  201,
		ContentType: "application/json",
		Response:    map[string]string{"message": "Created"},
	}

	jsonData, _ := json.Marshal(newRoute)
	req := httptest.NewRequest("POST", "/_mock/routes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	initialCount := configManager.GetRouteCount()
	handler.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Verify the route was added
	if configManager.GetRouteCount() != initialCount+1 {
		t.Errorf("Expected %d routes after adding, got %d", initialCount+1, configManager.GetRouteCount())
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["message"] != "Route added successfully" {
		t.Errorf("Expected success message, got %v", response["message"])
	}
}

func TestHandleAddRouteInvalid(t *testing.T) {
	handler, _ := createTestHandler()

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/_mock/routes", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}

	// Test invalid route (empty path)
	invalidRoute := config.Route{
		Path:   "",
		Method: "GET",
	}
	jsonData, _ := json.Marshal(invalidRoute)
	req = httptest.NewRequest("POST", "/_mock/routes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400 for invalid route, got %d", w.Code)
	}
}

func TestHandleUpdateRoute(t *testing.T) {
	handler, _ := createTestHandler()

	updatedRoute := config.Route{
		Path:        "/test/simple",
		Method:      "GET",
		StatusCode:  201,
		ContentType: "text/plain",
		Response:    "Updated response",
	}

	jsonData, _ := json.Marshal(updatedRoute)
	req := httptest.NewRequest("PUT", "/_mock/routes/test/simple", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test the updated route
	testReq := httptest.NewRequest("GET", "/test/simple", nil)
	testW := httptest.NewRecorder()
	handler.ServeHTTP(testW, testReq)

	if testW.Code != 201 {
		t.Errorf("Expected updated status code 201, got %d", testW.Code)
	}
	if testW.Body.String() != "Updated response" {
		t.Errorf("Expected updated response, got '%s'", testW.Body.String())
	}
}

func TestHandleDeleteRoute(t *testing.T) {
	handler, configManager := createTestHandler()

	initialCount := configManager.GetRouteCount()

	req := httptest.NewRequest("DELETE", "/_mock/routes/test/simple", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify the route was deleted
	if configManager.GetRouteCount() >= initialCount {
		t.Error("Expected route count to decrease after deletion")
	}

	// Test that the route no longer works
	testReq := httptest.NewRequest("GET", "/test/simple", nil)
	testW := httptest.NewRecorder()
	handler.ServeHTTP(testW, testReq)

	if testW.Code != 404 {
		t.Errorf("Expected 404 for deleted route, got %d", testW.Code)
	}
}

func TestHandleGetConfig(t *testing.T) {
	handler, _ := createTestHandler()

	req := httptest.NewRequest("GET", "/_mock/config", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response config.Config
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse config response: %v", err)
	}

	if len(response.Routes) != 5 {
		t.Errorf("Expected 5 routes in config, got %d", len(response.Routes))
	}
}

func TestMatchesPath(t *testing.T) {
	handler, _ := createTestHandler()

	tests := []struct {
		routePath   string
		requestPath string
		expected    bool
	}{
		{"/api/users", "/api/users", true},
		{"/api/users/", "/api/users", true},
		{"/api/users", "/api/users/", true},
		{"/api/users/*", "/api/users/123", true},
		{"/api/users/*", "/api/users/profile", true},
		{"/api/users", "/api/posts", false},
		{"/api/*", "/api/anything", true},
		{"/api/*", "/different/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.routePath+" vs "+tt.requestPath, func(t *testing.T) {
			result := handler.matchesPath(tt.routePath, tt.requestPath)
			if result != tt.expected {
				t.Errorf("matchesPath(%s, %s) = %v, expected %v", tt.routePath, tt.requestPath, result, tt.expected)
			}
		})
	}
}

func TestMatchesParameters(t *testing.T) {
	handler, _ := createTestHandler()

	// Test with parameters
	req := httptest.NewRequest("GET", "/test?type=user&status=active", nil)

	routeParams := map[string]string{
		"type": "user",
	}

	if !handler.matchesParameters(routeParams, req) {
		t.Error("Expected parameters to match")
	}

	// Test with non-matching parameters
	routeParams = map[string]string{
		"type": "admin",
	}

	if handler.matchesParameters(routeParams, req) {
		t.Error("Expected parameters not to match")
	}

	// Test with missing parameters
	routeParams = map[string]string{
		"missing": "value",
	}

	if handler.matchesParameters(routeParams, req) {
		t.Error("Expected parameters not to match when missing")
	}
}

func TestProcessResponse(t *testing.T) {
	handler, _ := createTestHandler()

	req := httptest.NewRequest("GET", "/api/test?name=john&id=123", nil)

	tests := []struct {
		name     string
		response interface{}
		expected string
	}{
		{
			name:     "String with placeholders",
			response: "Method: {method}, Path: {path}, Query: {query}",
			expected: "Method: GET, Path: /api/test, Query: name=john&id=123",
		},
		{
			name:     "String with query parameter",
			response: "Hello {name}!",
			expected: "Hello john!",
		},
		{
			name:     "Non-string response",
			response: map[string]string{"message": "test"},
			expected: "map[message:test]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.processResponse(tt.response, req)
			resultStr := fmt.Sprintf("%v", result)
			if resultStr != tt.expected {
				t.Errorf("processResponse() = %v, expected %v", resultStr, tt.expected)
			}
		})
	}
}

func TestHandleWebUI(t *testing.T) {
	handler, _ := createTestHandler()

	req := httptest.NewRequest("GET", "/_mock/ui", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Since the template file might not exist in test environment,
	// we expect either success (200) or internal server error (500)
	if w.Code != 200 && w.Code != 500 {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}

	if w.Code == 200 {
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html" {
			t.Errorf("Expected content type 'text/html', got '%s'", contentType)
		}
	}
}

func TestGetters(t *testing.T) {
	configManager := config.NewManager("test.yaml")
	log := logger.New(logger.LogLevelInfo)
	handler := NewMockHandler(configManager, log)

	if handler.GetConfigManager() != configManager {
		t.Error("GetConfigManager() returned wrong config manager")
	}

	if handler.GetLogger() != log {
		t.Error("GetLogger() returned wrong logger")
	}
}

func BenchmarkHandleMockEndpoint(b *testing.B) {
	handler, _ := createTestHandler()
	req := httptest.NewRequest("GET", "/test/simple", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkFindMatchingRoute(b *testing.B) {
	handler, _ := createTestHandler()
	req := httptest.NewRequest("GET", "/test/simple", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.findMatchingRoute(req)
	}
}

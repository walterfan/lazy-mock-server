package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/walterfan/lazy-mock-server/internal/config"
	"github.com/walterfan/lazy-mock-server/internal/logger"
)

// MockHandler handles HTTP requests for mock endpoints
type MockHandler struct {
	configManager *config.Manager
	logger        *logger.Logger
	mutex         sync.RWMutex
}

// NewMockHandler creates a new mock handler
func NewMockHandler(configManager *config.Manager, logger *logger.Logger) *MockHandler {
	return &MockHandler{
		configManager: configManager,
		logger:        logger,
	}
}

// ServeHTTP handles all incoming HTTP requests
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle management API endpoints
	if strings.HasPrefix(r.URL.Path, "/_mock/") {
		h.handleManagementAPI(w, r)
		return
	}

	// Handle regular mock endpoints
	h.handleMockEndpoint(w, r)
}

// handleMockEndpoint handles regular mock API requests
func (h *MockHandler) handleMockEndpoint(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	route := h.findMatchingRoute(r)
	h.mutex.RUnlock()

	if route == nil {
		h.handleNotFound(w, r)
		return
	}

	// Set custom headers if specified
	if route.Headers != nil {
		for key, value := range route.Headers {
			w.Header().Set(key, value)
		}
	}

	// Set content type (default to application/json if not specified)
	contentType := route.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)

	// Set status code (default to 200 if not specified)
	statusCode := route.StatusCode
	if statusCode == 0 {
		statusCode = 200
	}
	w.WriteHeader(statusCode)

	// Process response body
	responseBody := h.processResponse(route.Response, r)

	// Write response based on content type
	switch contentType {
	case "application/json":
		if str, ok := responseBody.(string); ok {
			// If response is already a string, try to parse as JSON
			var jsonObj interface{}
			if err := json.Unmarshal([]byte(str), &jsonObj); err == nil {
				if err := json.NewEncoder(w).Encode(jsonObj); err != nil {
					h.logger.LogError(err, "encoding JSON object")
				}
			} else {
				// If not valid JSON, wrap in quotes
				if err := json.NewEncoder(w).Encode(str); err != nil {
					h.logger.LogError(err, "encoding string response")
				}
			}
		} else {
			if err := json.NewEncoder(w).Encode(responseBody); err != nil {
				h.logger.LogError(err, "encoding response body")
			}
		}
	default:
		// For text/plain and other content types, convert to string
		if _, err := fmt.Fprintf(w, "%v", responseBody); err != nil {
			h.logger.LogError(err, "writing response body")
		}
	}
}

// handleManagementAPI handles the management API endpoints
func (h *MockHandler) handleManagementAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/_mock/routes" && r.Method == "GET":
		h.handleGetRoutes(w, r)
	case r.URL.Path == "/_mock/routes" && r.Method == "POST":
		h.handleAddRoute(w, r)
	case strings.HasPrefix(r.URL.Path, "/_mock/routes/") && r.Method == "PUT":
		h.handleUpdateRoute(w, r)
	case strings.HasPrefix(r.URL.Path, "/_mock/routes/") && r.Method == "DELETE":
		h.handleDeleteRoute(w, r)
	case r.URL.Path == "/_mock/config" && r.Method == "GET":
		h.handleGetConfig(w, r)
	case r.URL.Path == "/_mock/config" && r.Method == "POST":
		h.handleSaveConfig(w, r)
	case r.URL.Path == "/_mock/ui" && r.Method == "GET":
		h.handleWebUI(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleGetRoutes returns all current routes
func (h *MockHandler) handleGetRoutes(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	routes := h.configManager.GetRoutes()
	count := h.configManager.GetRouteCount()
	h.mutex.RUnlock()

	response := map[string]interface{}{
		"routes": routes,
		"count":  count,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.LogErrorWithRequest(err, r, "encoding routes response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleAddRoute adds a new route
func (h *MockHandler) handleAddRoute(w http.ResponseWriter, r *http.Request) {
	var newRoute config.Route
	if err := json.NewDecoder(r.Body).Decode(&newRoute); err != nil {
		h.logger.LogErrorWithRequest(err, r, "decoding new route")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"}); encErr != nil {
			h.logger.LogError(encErr, "encoding error response")
		}
		return
	}

	// Validate the route
	if err := h.configManager.ValidateRoute(newRoute); err != nil {
		h.logger.LogErrorWithRequest(err, r, "validating new route")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encErr != nil {
			h.logger.LogError(encErr, "encoding error response")
		}
		return
	}

	h.mutex.Lock()
	h.configManager.AddRoute(newRoute)
	h.mutex.Unlock()

	h.logger.LogInfo("Added new route: %s %s", newRoute.Method, newRoute.Path)

	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"message": "Route added successfully",
		"route":   newRoute,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.LogError(err, "encoding add route response")
	}
}

// handleUpdateRoute updates an existing route
func (h *MockHandler) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid route path"}); err != nil {
			h.logger.LogError(err, "encoding error response")
		}
		return
	}

	routePath := "/" + strings.Join(pathParts[3:], "/")

	var updatedRoute config.Route
	if err := json.NewDecoder(r.Body).Decode(&updatedRoute); err != nil {
		h.logger.LogErrorWithRequest(err, r, "decoding updated route")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"}); encErr != nil {
			h.logger.LogError(encErr, "encoding error response")
		}
		return
	}

	// Validate the route
	if err := h.configManager.ValidateRoute(updatedRoute); err != nil {
		h.logger.LogErrorWithRequest(err, r, "validating updated route")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encErr != nil {
			h.logger.LogError(encErr, "encoding error response")
		}
		return
	}

	h.mutex.Lock()
	err := h.configManager.DeleteRouteByPath(routePath)
	if err == nil {
		h.configManager.AddRoute(updatedRoute)
	}
	h.mutex.Unlock()

	if err != nil {
		h.logger.LogErrorWithRequest(err, r, "updating route")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Route not found"})
		return
	}

	h.logger.LogInfo("Updated route: %s %s", updatedRoute.Method, updatedRoute.Path)

	response := map[string]interface{}{
		"message": "Route updated successfully",
		"route":   updatedRoute,
	}
	json.NewEncoder(w).Encode(response)
}

// handleDeleteRoute deletes a route
func (h *MockHandler) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid route path"})
		return
	}

	routePath := "/" + strings.Join(pathParts[3:], "/")

	h.mutex.Lock()
	err := h.configManager.DeleteRouteByPath(routePath)
	h.mutex.Unlock()

	if err != nil {
		h.logger.LogErrorWithRequest(err, r, "deleting route")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Route not found"})
		return
	}

	h.logger.LogInfo("Deleted route: %s", routePath)

	json.NewEncoder(w).Encode(map[string]string{"message": "Route deleted successfully"})
}

// handleGetConfig returns the current configuration
func (h *MockHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	config := h.configManager.GetConfig()
	h.mutex.RUnlock()

	if err := json.NewEncoder(w).Encode(config); err != nil {
		h.logger.LogErrorWithRequest(err, r, "encoding config response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSaveConfig saves the current configuration to file
func (h *MockHandler) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	err := h.configManager.Save()
	h.mutex.RUnlock()

	if err != nil {
		h.logger.LogErrorWithRequest(err, r, "saving configuration")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save configuration"})
		return
	}

	h.logger.LogInfo("Configuration saved to file: %s", h.configManager.GetConfigPath())

	json.NewEncoder(w).Encode(map[string]string{"message": "Configuration saved successfully"})
}

// handleWebUI serves the web UI
func (h *MockHandler) handleWebUI(w http.ResponseWriter, r *http.Request) {
	// Read the template file
	templatePath := "internal/templates/web_ui.html"
	htmlContent, err := os.ReadFile(templatePath)
	if err != nil {
		h.logger.LogErrorWithRequest(err, r, "reading web UI template")
		http.Error(w, "Web UI template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write(htmlContent); err != nil {
		h.logger.LogError(err, "writing web UI content")
	}
}

// handleNotFound handles requests that don't match any route
func (h *MockHandler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	response := map[string]string{
		"error":  "Route not found",
		"path":   r.URL.Path,
		"method": r.Method,
	}
	json.NewEncoder(w).Encode(response)
}

// findMatchingRoute finds the first route that matches the request
func (h *MockHandler) findMatchingRoute(r *http.Request) *config.Route {
	routes := h.configManager.GetRoutes()
	for _, route := range routes {
		if h.matchesRoute(&route, r) {
			return &route
		}
	}
	return nil
}

// matchesRoute checks if a route matches the request
func (h *MockHandler) matchesRoute(route *config.Route, r *http.Request) bool {
	// Check HTTP method
	if !strings.EqualFold(route.Method, r.Method) {
		return false
	}

	// Check path - support exact match and pattern matching
	if h.matchesPath(route.Path, r.URL.Path) {
		// Check parameters if specified
		if route.Parameters != nil {
			return h.matchesParameters(route.Parameters, r)
		}
		return true
	}

	return false
}

// matchesPath checks if the route path matches the request path
func (h *MockHandler) matchesPath(routePath, requestPath string) bool {
	// Remove trailing slashes for comparison
	routePath = strings.TrimSuffix(routePath, "/")
	requestPath = strings.TrimSuffix(requestPath, "/")

	// Exact match
	if routePath == requestPath {
		return true
	}

	// Pattern matching with wildcards
	if strings.Contains(routePath, "*") {
		pattern := strings.ReplaceAll(routePath, "*", ".*")
		matched, _ := regexp.MatchString("^"+pattern+"$", requestPath)
		return matched
	}

	// Prefix matching
	if strings.HasSuffix(routePath, "/") {
		return strings.HasPrefix(requestPath, routePath)
	}

	return false
}

// matchesParameters checks if request parameters match the route requirements
func (h *MockHandler) matchesParameters(routeParams map[string]string, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		h.logger.LogError(err, "parsing form parameters")
		return false
	}

	for key, expectedValue := range routeParams {
		actualValue := r.FormValue(key)
		if actualValue != expectedValue {
			return false
		}
	}
	return true
}

// processResponse processes the response body, handling dynamic content
func (h *MockHandler) processResponse(response interface{}, r *http.Request) interface{} {
	if str, ok := response.(string); ok {
		// Replace placeholders with request data
		str = strings.ReplaceAll(str, "{method}", r.Method)
		str = strings.ReplaceAll(str, "{path}", r.URL.Path)
		str = strings.ReplaceAll(str, "{query}", r.URL.RawQuery)

		// Replace query parameters
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				placeholder := fmt.Sprintf("{%s}", key)
				str = strings.ReplaceAll(str, placeholder, values[0])
			}
		}

		return str
	}

	return response
}

// GetConfigManager returns the configuration manager
func (h *MockHandler) GetConfigManager() *config.Manager {
	return h.configManager
}

// GetLogger returns the logger
func (h *MockHandler) GetLogger() *logger.Logger {
	return h.logger
}

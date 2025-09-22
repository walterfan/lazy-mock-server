package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("test.yaml")
	if manager == nil {
		t.Error("Expected manager to be created")
		return
	}
	if manager.configPath != "test.yaml" {
		t.Errorf("Expected config path 'test.yaml', got '%s'", manager.configPath)
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	// Create test configuration data
	configData := `routes:
  - path: "/test"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response:
      message: "test"
`

	// Write test config file
	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading
	manager := NewManager(configPath)
	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded data
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Config is nil")
	}

	if len(config.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(config.Routes))
	}

	route := config.Routes[0]
	if route.Path != "/test" {
		t.Errorf("Expected path '/test', got '%s'", route.Path)
	}
	if route.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", route.Method)
	}
	if route.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", route.StatusCode)
	}

	// Test saving - modify the route in the manager
	routes := manager.GetRoutes()
	routes[0].StatusCode = 201
	manager.GetConfig().Routes[0].StatusCode = 201

	err = manager.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load again to verify save worked
	manager2 := NewManager(configPath)
	err = manager2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	reloadedRoute := manager2.GetRoutes()[0]
	if reloadedRoute.StatusCode != 201 {
		t.Errorf("Expected status code 201 after save/reload, got %d", reloadedRoute.StatusCode)
	}
}

func TestAddRoute(t *testing.T) {
	manager := NewManager("test.yaml")

	route := Route{
		Path:        "/api/test",
		Method:      "POST",
		StatusCode:  201,
		ContentType: "application/json",
		Response:    map[string]string{"status": "created"},
	}

	manager.AddRoute(route)

	routes := manager.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].Path != "/api/test" {
		t.Errorf("Expected path '/api/test', got '%s'", routes[0].Path)
	}
}

func TestUpdateRoute(t *testing.T) {
	manager := NewManager("test.yaml")

	// Add initial route
	originalRoute := Route{
		Path:        "/api/test",
		Method:      "GET",
		StatusCode:  200,
		ContentType: "application/json",
		Response:    "original",
	}
	manager.AddRoute(originalRoute)

	// Update the route
	updatedRoute := Route{
		Path:        "/api/test",
		Method:      "GET",
		StatusCode:  201,
		ContentType: "text/plain",
		Response:    "updated",
	}

	err := manager.UpdateRoute("/api/test", "GET", updatedRoute)
	if err != nil {
		t.Fatalf("Failed to update route: %v", err)
	}

	// Verify update
	routes := manager.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route after update, got %d", len(routes))
	}

	route := routes[0]
	if route.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", route.StatusCode)
	}
	if route.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", route.ContentType)
	}
	if route.Response != "updated" {
		t.Errorf("Expected response 'updated', got '%v'", route.Response)
	}
}

func TestDeleteRoute(t *testing.T) {
	manager := NewManager("test.yaml")

	// Add routes
	route1 := Route{Path: "/api/test1", Method: "GET", StatusCode: 200}
	route2 := Route{Path: "/api/test2", Method: "POST", StatusCode: 201}

	manager.AddRoute(route1)
	manager.AddRoute(route2)

	// Verify initial count
	if manager.GetRouteCount() != 2 {
		t.Errorf("Expected 2 routes initially, got %d", manager.GetRouteCount())
	}

	// Delete one route
	err := manager.DeleteRoute("/api/test1", "GET")
	if err != nil {
		t.Fatalf("Failed to delete route: %v", err)
	}

	// Verify deletion
	if manager.GetRouteCount() != 1 {
		t.Errorf("Expected 1 route after deletion, got %d", manager.GetRouteCount())
	}

	routes := manager.GetRoutes()
	if routes[0].Path != "/api/test2" {
		t.Errorf("Expected remaining route path '/api/test2', got '%s'", routes[0].Path)
	}

	// Try to delete non-existent route
	err = manager.DeleteRoute("/api/nonexistent", "GET")
	if err == nil {
		t.Error("Expected error when deleting non-existent route")
	}
}

func TestDeleteRouteByPath(t *testing.T) {
	manager := NewManager("test.yaml")

	// Add routes with same path but different methods
	route1 := Route{Path: "/api/test", Method: "GET", StatusCode: 200}
	route2 := Route{Path: "/api/test", Method: "POST", StatusCode: 201}
	route3 := Route{Path: "/api/other", Method: "GET", StatusCode: 200}

	manager.AddRoute(route1)
	manager.AddRoute(route2)
	manager.AddRoute(route3)

	// Delete all routes with specific path
	err := manager.DeleteRouteByPath("/api/test")
	if err != nil {
		t.Fatalf("Failed to delete routes by path: %v", err)
	}

	// Verify deletion
	routes := manager.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route after deletion, got %d", len(routes))
	}
	if routes[0].Path != "/api/other" {
		t.Errorf("Expected remaining route path '/api/other', got '%s'", routes[0].Path)
	}
}

func TestFindRoute(t *testing.T) {
	manager := NewManager("test.yaml")

	route := Route{
		Path:       "/api/test",
		Method:     "GET",
		StatusCode: 200,
		Response:   "test response",
	}
	manager.AddRoute(route)

	// Find existing route
	found, err := manager.FindRoute("/api/test", "GET")
	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}
	if found.Response != "test response" {
		t.Errorf("Expected response 'test response', got '%v'", found.Response)
	}

	// Try to find non-existent route
	_, err = manager.FindRoute("/api/nonexistent", "GET")
	if err == nil {
		t.Error("Expected error when finding non-existent route")
	}
}

func TestValidateRoute(t *testing.T) {
	manager := NewManager("test.yaml")

	tests := []struct {
		name    string
		route   Route
		wantErr bool
	}{
		{
			name: "Valid route",
			route: Route{
				Path:       "/api/test",
				Method:     "GET",
				StatusCode: 200,
			},
			wantErr: false,
		},
		{
			name: "Empty path",
			route: Route{
				Path:   "",
				Method: "GET",
			},
			wantErr: true,
		},
		{
			name: "Empty method",
			route: Route{
				Path:   "/api/test",
				Method: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid method",
			route: Route{
				Path:   "/api/test",
				Method: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "Invalid status code - too low",
			route: Route{
				Path:       "/api/test",
				Method:     "GET",
				StatusCode: 99,
			},
			wantErr: true,
		},
		{
			name: "Invalid status code - too high",
			route: Route{
				Path:       "/api/test",
				Method:     "GET",
				StatusCode: 600,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateRoute(tt.route)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClone(t *testing.T) {
	manager := NewManager("test.yaml")

	route := Route{
		Path:        "/api/test",
		Method:      "GET",
		StatusCode:  200,
		ContentType: "application/json",
		Response:    map[string]string{"message": "test"},
	}
	manager.AddRoute(route)

	// Clone the configuration
	cloned := manager.Clone()
	if cloned == nil {
		t.Fatal("Cloned config is nil")
	}

	// Verify clone has same data
	if len(cloned.Routes) != 1 {
		t.Errorf("Expected 1 route in clone, got %d", len(cloned.Routes))
	}

	clonedRoute := cloned.Routes[0]
	if clonedRoute.Path != "/api/test" {
		t.Errorf("Expected cloned route path '/api/test', got '%s'", clonedRoute.Path)
	}

	// Modify original and verify clone is unchanged
	manager.AddRoute(Route{Path: "/api/test2", Method: "POST", StatusCode: 201})
	if len(cloned.Routes) != 1 {
		t.Error("Clone was modified when original changed")
	}
}

func TestLoadFromBytes(t *testing.T) {
	manager := NewManager("test.yaml")

	configData := []byte(`routes:
  - path: "/from/bytes"
    method: "GET"
    status_code: 200
    content_type: "text/plain"
    response: "loaded from bytes"
`)

	err := manager.LoadFromBytes(configData)
	if err != nil {
		t.Fatalf("Failed to load from bytes: %v", err)
	}

	routes := manager.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Path != "/from/bytes" {
		t.Errorf("Expected path '/from/bytes', got '%s'", route.Path)
	}
	if route.Response != "loaded from bytes" {
		t.Errorf("Expected response 'loaded from bytes', got '%v'", route.Response)
	}
}

func TestToBytes(t *testing.T) {
	manager := NewManager("test.yaml")

	route := Route{
		Path:        "/to/bytes",
		Method:      "GET",
		StatusCode:  200,
		ContentType: "text/plain",
		Response:    "converted to bytes",
	}
	manager.AddRoute(route)

	data, err := manager.ToBytes()
	if err != nil {
		t.Fatalf("Failed to convert to bytes: %v", err)
	}

	// Load into new manager to verify
	manager2 := NewManager("test2.yaml")
	err = manager2.LoadFromBytes(data)
	if err != nil {
		t.Fatalf("Failed to load converted bytes: %v", err)
	}

	routes := manager2.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route after conversion, got %d", len(routes))
	}

	convertedRoute := routes[0]
	if convertedRoute.Path != "/to/bytes" {
		t.Errorf("Expected converted route path '/to/bytes', got '%s'", convertedRoute.Path)
	}
	if convertedRoute.Response != "converted to bytes" {
		t.Errorf("Expected converted response 'converted to bytes', got '%v'", convertedRoute.Response)
	}
}

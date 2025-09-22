package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/walterfan/lazy-mock-server/internal/config"
	"github.com/walterfan/lazy-mock-server/internal/logger"
)

func createTestConfig(t *testing.T) string {
	configData := `routes:
  - path: "/test"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response:
      message: "test"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	return configPath
}

func TestNew(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8081,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError, // Use error level to reduce test output
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.port != 8081 {
		t.Errorf("Expected port 8081, got %d", server.port)
	}

	if server.configManager == nil {
		t.Error("Expected config manager to be set")
	}

	if server.handler == nil {
		t.Error("Expected handler to be set")
	}

	if server.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	cfg := Config{
		Port:       8081,
		ConfigPath: "/nonexistent/config.yaml",
		LogLevel:   logger.LogLevelError,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid config path")
	}
}

func TestServerMethods(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8082,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test getters
	if server.GetPort() != 8082 {
		t.Errorf("Expected port 8082, got %d", server.GetPort())
	}

	if server.GetConfigPath() != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, server.GetConfigPath())
	}

	if server.GetLogger() == nil {
		t.Error("Expected logger to be returned")
	}

	if server.GetConfigManager() == nil {
		t.Error("Expected config manager to be returned")
	}

	if server.GetHandler() == nil {
		t.Error("Expected handler to be returned")
	}

	// Test stats
	stats := server.GetStats()
	if stats["port"] != 8082 {
		t.Errorf("Expected port 8082 in stats, got %v", stats["port"])
	}

	if stats["route_count"] != 1 {
		t.Errorf("Expected 1 route in stats, got %v", stats["route_count"])
	}

	// Test version info
	version := server.GetVersion()
	if version["name"] != "Lazy Mock Server" {
		t.Errorf("Expected name 'Lazy Mock Server', got %s", version["name"])
	}

	// Test health check
	if !server.IsHealthy() {
		t.Error("Expected server to be healthy")
	}
}

func TestServerRouteManagement(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8083,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test adding route
	newRoute := config.Route{
		Path:        "/api/new",
		Method:      "POST",
		StatusCode:  201,
		ContentType: "application/json",
		Response:    map[string]string{"status": "created"},
	}

	err = server.AddRoute(newRoute)
	if err != nil {
		t.Errorf("Failed to add route: %v", err)
	}

	routes := server.GetRoutes()
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes after adding, got %d", len(routes))
	}

	// Test getting specific route
	foundRoute, err := server.GetRoute("/api/new", "POST")
	if err != nil {
		t.Errorf("Failed to get route: %v", err)
	}
	if foundRoute.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", foundRoute.StatusCode)
	}

	// Test updating route
	updatedRoute := config.Route{
		Path:        "/api/new",
		Method:      "POST",
		StatusCode:  202,
		ContentType: "application/json",
		Response:    map[string]string{"status": "updated"},
	}

	err = server.UpdateRoute("/api/new", "POST", updatedRoute)
	if err != nil {
		t.Errorf("Failed to update route: %v", err)
	}

	updatedFoundRoute, err := server.GetRoute("/api/new", "POST")
	if err != nil {
		t.Errorf("Failed to get updated route: %v", err)
	}
	if updatedFoundRoute.StatusCode != 202 {
		t.Errorf("Expected updated status code 202, got %d", updatedFoundRoute.StatusCode)
	}

	// Test deleting route
	err = server.DeleteRoute("/api/new", "POST")
	if err != nil {
		t.Errorf("Failed to delete route: %v", err)
	}

	routes = server.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route after deletion, got %d", len(routes))
	}

	// Try to get deleted route
	_, err = server.GetRoute("/api/new", "POST")
	if err == nil {
		t.Error("Expected error when getting deleted route")
	}
}

func TestServerStartStop(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8084,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestSetLogLevel(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8085,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelInfo,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test setting log level
	server.SetLogLevel(logger.LogLevelDebug)

	if server.logger.GetLogLevel() != logger.LogLevelDebug {
		t.Errorf("Expected log level %d, got %d", logger.LogLevelDebug, server.logger.GetLogLevel())
	}
}

func TestReload(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8086,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initial route count
	initialCount := server.GetConfigManager().GetRouteCount()

	// Modify config file to add another route
	newConfigData := `routes:
  - path: "/test"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response:
      message: "test"
  - path: "/test2"
    method: "POST"
    status_code: 201
    content_type: "application/json"
    response:
      message: "test2"
`

	err = os.WriteFile(configPath, []byte(newConfigData), 0644)
	if err != nil {
		t.Fatalf("Failed to update config file: %v", err)
	}

	// Reload configuration
	err = server.Reload()
	if err != nil {
		t.Errorf("Failed to reload config: %v", err)
	}

	// Verify new route count
	newCount := server.GetConfigManager().GetRouteCount()
	if newCount != initialCount+1 {
		t.Errorf("Expected %d routes after reload, got %d", initialCount+1, newCount)
	}
}

func TestSaveConfig(t *testing.T) {
	configPath := createTestConfig(t)

	cfg := Config{
		Port:       8087,
		ConfigPath: configPath,
		LogLevel:   logger.LogLevelError,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add a route
	newRoute := config.Route{
		Path:        "/api/save-test",
		Method:      "GET",
		StatusCode:  200,
		ContentType: "text/plain",
		Response:    "saved",
	}

	if err := server.AddRoute(newRoute); err != nil {
		t.Errorf("Failed to add route: %v", err)
	}

	// Save configuration
	err = server.SaveConfig()
	if err != nil {
		t.Errorf("Failed to save config: %v", err)
	}

	// Create new server and verify the route was saved
	server2, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create second server: %v", err)
	}

	routes := server2.GetRoutes()
	found := false
	for _, route := range routes {
		if route.Path == "/api/save-test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected saved route to be found in new server instance")
	}
}

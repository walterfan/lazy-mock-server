package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/walterfan/lazy-mock-server/internal/config"
	"github.com/walterfan/lazy-mock-server/internal/handlers"
	"github.com/walterfan/lazy-mock-server/internal/logger"
)

// Server represents the mock server
type Server struct {
	httpServer    *http.Server
	configManager *config.Manager
	handler       *handlers.MockHandler
	logger        *logger.Logger
	port          int
	configPath    string
}

// Config represents server configuration
type Config struct {
	Port       int
	ConfigPath string
	LogLevel   logger.LogLevel
}

// New creates a new mock server instance
func New(cfg Config) (*Server, error) {
	// Initialize logger
	log := logger.New(cfg.LogLevel)

	// Get absolute path for config file
	configPath := cfg.ConfigPath
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	// Initialize configuration manager
	configManager := config.NewManager(configPath)
	if err := configManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	log.LogInfo("Loaded configuration from: %s", configPath)
	log.LogInfo("Found %d routes in configuration", configManager.GetRouteCount())

	// Initialize handlers
	mockHandler := handlers.NewMockHandler(configManager, log)

	// Create HTTP server with logging middleware
	mux := http.NewServeMux()
	mux.Handle("/", log.Middleware(mockHandler))

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	server := &Server{
		httpServer:    httpServer,
		configManager: configManager,
		handler:       mockHandler,
		logger:        log,
		port:          cfg.Port,
		configPath:    configPath,
	}

	return server, nil
}

// Start starts the mock server
func (s *Server) Start() error {
	s.logger.LogInfo("Starting mock server on port %d", s.port)
	s.logger.LogInfo("Using configuration file: %s", s.configPath)
	s.logger.LogInfo("Web UI available at: http://localhost:%d/_mock/ui", s.port)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.LogError(err, "HTTP server error")
		}
	}()

	s.logger.LogInfo("Mock server started successfully")
	return nil
}

// Stop gracefully stops the mock server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.LogInfo("Shutting down mock server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.LogError(err, "server shutdown")
		return err
	}

	s.logger.LogInfo("Mock server stopped")
	return nil
}

// Run starts the server and waits for shutdown signals
func (s *Server) Run() error {
	// Start the server
	if err := s.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.logger.LogInfo("Received shutdown signal")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.Stop(ctx)
}

// GetPort returns the server port
func (s *Server) GetPort() int {
	return s.port
}

// GetConfigPath returns the configuration file path
func (s *Server) GetConfigPath() string {
	return s.configPath
}

// GetLogger returns the logger
func (s *Server) GetLogger() *logger.Logger {
	return s.logger
}

// GetConfigManager returns the configuration manager
func (s *Server) GetConfigManager() *config.Manager {
	return s.configManager
}

// GetHandler returns the mock handler
func (s *Server) GetHandler() *handlers.MockHandler {
	return s.handler
}

// Reload reloads the configuration from file
func (s *Server) Reload() error {
	s.logger.LogInfo("Reloading configuration...")

	if err := s.configManager.Load(); err != nil {
		s.logger.LogError(err, "reloading configuration")
		return err
	}

	s.logger.LogInfo("Configuration reloaded successfully")
	s.logger.LogInfo("Found %d routes in configuration", s.configManager.GetRouteCount())

	return nil
}

// SaveConfig saves the current configuration to file
func (s *Server) SaveConfig() error {
	s.logger.LogInfo("Saving configuration...")

	if err := s.configManager.Save(); err != nil {
		s.logger.LogError(err, "saving configuration")
		return err
	}

	s.logger.LogInfo("Configuration saved successfully")
	return nil
}

// GetStats returns server statistics
func (s *Server) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"port":        s.port,
		"config_path": s.configPath,
		"route_count": s.configManager.GetRouteCount(),
		"log_level":   s.logger.GetLogLevel(),
	}
}

// SetLogLevel sets the logging level
func (s *Server) SetLogLevel(level logger.LogLevel) {
	s.logger.SetLogLevel(level)
	s.logger.LogInfo("Log level set to: %d", level)
}

// AddRoute adds a new route to the configuration
func (s *Server) AddRoute(route config.Route) error {
	if err := s.configManager.ValidateRoute(route); err != nil {
		return err
	}

	s.configManager.AddRoute(route)
	s.logger.LogInfo("Added route: %s %s", route.Method, route.Path)
	return nil
}

// UpdateRoute updates an existing route
func (s *Server) UpdateRoute(path, method string, newRoute config.Route) error {
	if err := s.configManager.ValidateRoute(newRoute); err != nil {
		return err
	}

	if err := s.configManager.UpdateRoute(path, method, newRoute); err != nil {
		return err
	}

	s.logger.LogInfo("Updated route: %s %s", newRoute.Method, newRoute.Path)
	return nil
}

// DeleteRoute deletes a route
func (s *Server) DeleteRoute(path, method string) error {
	if err := s.configManager.DeleteRoute(path, method); err != nil {
		return err
	}

	s.logger.LogInfo("Deleted route: %s %s", method, path)
	return nil
}

// GetRoutes returns all configured routes
func (s *Server) GetRoutes() []config.Route {
	return s.configManager.GetRoutes()
}

// GetRoute finds a specific route
func (s *Server) GetRoute(path, method string) (*config.Route, error) {
	return s.configManager.FindRoute(path, method)
}

// IsHealthy checks if the server is healthy
func (s *Server) IsHealthy() bool {
	// Simple health check - server is healthy if it's running
	return s.httpServer != nil
}

// GetVersion returns the server version information
func (s *Server) GetVersion() map[string]string {
	return map[string]string{
		"version": "1.0.0",
		"name":    "Lazy Mock Server",
		"author":  "Walter Fan",
	}
}

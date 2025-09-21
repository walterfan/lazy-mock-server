package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/walterfan/lazy-mock-server/internal/logger"
	"github.com/walterfan/lazy-mock-server/internal/server"
)

// Version information
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command-line arguments
	var (
		port       = flag.Int("port", 8080, "Port to listen on")
		configPath = flag.String("config", "app/mock_response.yaml", "Path to configuration file")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Show version information
	if *version {
		fmt.Printf("Lazy Mock Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Parse log level
	var logLevelEnum logger.LogLevel
	switch *logLevel {
	case "debug":
		logLevelEnum = logger.LogLevelDebug
	case "info":
		logLevelEnum = logger.LogLevelInfo
	case "warn":
		logLevelEnum = logger.LogLevelWarn
	case "error":
		logLevelEnum = logger.LogLevelError
	default:
		log.Fatalf("Invalid log level: %s (must be debug, info, warn, or error)", *logLevel)
	}

	// Create server configuration
	serverConfig := server.Config{
		Port:       *port,
		ConfigPath: *configPath,
		LogLevel:   logLevelEnum,
	}

	// Create and start the server
	srv, err := server.New(serverConfig)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Display startup information
	fmt.Printf("🚀 Lazy Mock Server v%s\n", Version)
	fmt.Printf("📁 Config: %s\n", srv.GetConfigPath())
	fmt.Printf("🌐 Server: http://localhost:%d\n", srv.GetPort())
	fmt.Printf("🎛️  Web UI: http://localhost:%d/_mock/ui\n", srv.GetPort())
	fmt.Printf("📊 Routes: %d configured\n", srv.GetConfigManager().GetRouteCount())
	fmt.Println("🔥 Server starting...")

	// Run the server (blocks until shutdown signal)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	fmt.Println("👋 Server stopped gracefully")
}

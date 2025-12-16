# Makefile for Lazy Mock Server

# Variables
BINARY_NAME=mock-server
BINARY_UNIX=$(BINARY_NAME)-unix
BINARY_WINDOWS=$(BINARY_NAME)-windows.exe
BINARY_DARWIN=$(BINARY_NAME)-darwin
GO_FILES=$(shell find . -name "*.go" -type f)
#CONFIG_FILE=app/mock_response.yaml
CONFIG_FILE=config/mock_config.yaml

# Default target
.PHONY: all
all: clean lint test build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  build-all      Build binaries for all platforms"
	@echo "  clean          Clean build artifacts"
	@echo "  test           Run tests"
	@echo "  test-verbose   Run tests with verbose output"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  lint           Run linter (golangci-lint)"
	@echo "  fmt            Format Go code"
	@echo "  vet            Run go vet"
	@echo "  deps           Download dependencies"
	@echo "  tidy           Tidy up go.mod"
	@echo "  run            Run the server with default config"
	@echo "  run-advanced   Run the server with advanced config"
	@echo "  run-https      Run the server with HTTPS/TLS"
	@echo "  cert           Generate self-signed TLS certificate"
	@echo "  test-https     Test HTTPS endpoints"
	@echo "  install        Install the binary to GOPATH/bin"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo "  benchmark      Run benchmarks"
	@echo "  check          Run all checks (lint, vet, test)"
	@echo "  release        Create a release build"
	@echo "  all            Clean, lint, test, and build"

# Build targets
.PHONY: build
build: deps
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) main.go
	@echo "Build complete: $(BINARY_NAME)"

.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux: deps
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_UNIX) main.go

.PHONY: build-windows
build-windows: deps
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_WINDOWS) main.go

.PHONY: build-darwin
build-darwin: deps
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_DARWIN) main.go

.PHONY: release
release: clean test lint build-all
	@echo "Creating release artifacts..."
	@mkdir -p release
	@cp $(BINARY_NAME) release/
	@cp $(BINARY_UNIX) release/
	@cp $(BINARY_WINDOWS) release/
	@cp $(BINARY_DARWIN) release/
	@cp README.md release/
	@cp -r config release/
	@cp -r app release/
	@echo "Release artifacts created in release/ directory"

# Development targets
.PHONY: run
run: build
	@echo "Starting mock server with default config..."
	./$(BINARY_NAME) -config $(CONFIG_FILE) -port 8080

.PHONY: run-advanced
run-advanced: build
	@echo "Starting mock server with advanced config..."
	./$(BINARY_NAME) -config $(CONFIG_FILE) -port 8080

.PHONY: run-dev
run-dev:
	@echo "Starting mock server in development mode..."
	go run main.go -config $(CONFIG_FILE) -port 8080

.PHONY: run-https
run-https: build
	@echo "Starting mock server with HTTPS..."
	@if [ ! -f server.crt ] || [ ! -f server.key ]; then \
		echo "Generating self-signed certificate..."; \
		./generate_cert.sh; \
	fi
	./$(BINARY_NAME) -config $(CONFIG_FILE) -tls -cert server.crt -key server.key -port 8443

.PHONY: cert
cert:
	@echo "Generating self-signed TLS certificate..."
	./generate_cert.sh

.PHONY: test-https
test-https:
	@echo "Testing HTTPS endpoints..."
	./test_https.sh

.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

.PHONY: tidy
tidy:
	@echo "Tidying up go.mod..."
	go mod tidy

.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Code formatted"

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "go vet passed"

.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m --no-config; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --timeout=5m --no-config; \
	fi
	@echo "Linting passed"

.PHONY: lint-install
lint-install:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-verbose
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-short
test-short:
	@echo "Running short tests..."
	go test -short ./...

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t lazy-mock-server .

.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/config:/config lazy-mock-server

# Utility targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_UNIX)
	@rm -f $(BINARY_WINDOWS)
	@rm -f $(BINARY_DARWIN)
	@rm -f coverage.out
	@rm -f coverage.html
	@rm -rf release/
	@rm -rf vendor/
	@echo "Clean complete"

.PHONY: check
check: fmt vet lint test
	@echo "All checks passed!"

.PHONY: ci
ci: deps check build
	@echo "CI pipeline completed successfully"

# File watching for development
.PHONY: watch
watch:
	@if command -v air >/dev/null 2>&1; then \
		echo "Starting file watcher with air..."; \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to manual restart..."; \
		$(MAKE) run-dev; \
	fi

.PHONY: watch-install
watch-install:
	@echo "Installing air for file watching..."
	go install github.com/cosmtrek/air@latest

# Testing API endpoints
.PHONY: test-api
test-api:
	@echo "Testing API endpoints..."
	@echo "Make sure the server is running first with: make run"
	@echo ""
	@echo "Testing basic endpoints..."
	@curl -s http://localhost:8080/v1/metadata/sn || echo "Server not running?"
	@echo ""
	@echo "Testing management API..."
	@curl -s http://localhost:8080/_mock/routes | jq . || echo "jq not installed or server not running"
	@echo ""
	@echo "Testing web UI..."
	@echo "Open http://localhost:8080/_mock/ui in your browser"

.PHONY: test-endpoints
test-endpoints:
	@echo "Running comprehensive endpoint tests..."
	@echo "Starting server in background..."
	@./$(BINARY_NAME) -config $(CONFIG_FILE) -port 8081 &
	@SERVER_PID=$$!; \
	sleep 2; \
	echo "Testing endpoints..."; \
	curl -s http://localhost:8081/v1/metadata/sn && echo " ✓ Basic endpoint works"; \
	curl -s http://localhost:8081/_mock/routes >/dev/null && echo " ✓ Management API works"; \
	kill $$SERVER_PID; \
	echo "Endpoint tests complete"

# Security scanning
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

.PHONY: security-install
security-install:
	@echo "Installing gosec for security scanning..."
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Performance profiling
.PHONY: profile-cpu
profile-cpu:
	@echo "Running CPU profile..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

.PHONY: profile-mem
profile-mem:
	@echo "Running memory profile..."
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

.PHONY: docs-install
docs-install:
	@echo "Installing godoc..."
	go install golang.org/x/tools/cmd/godoc@latest

# Show project statistics
.PHONY: stats
stats:
	@echo "Project Statistics:"
	@echo "==================="
	@echo "Go files: $(shell find . -name '*.go' | wc -l)"
	@echo "Total lines of Go code: $(shell find . -name '*.go' -exec cat {} \; | wc -l)"
	@echo "Total lines of code (all files): $(shell find . -type f -name '*.go' -o -name '*.yaml' -o -name '*.yml' -o -name '*.md' | xargs cat | wc -l)"
	@echo "Binary size: $(shell ls -lh $(BINARY_NAME) 2>/dev/null | awk '{print $$5}' || echo 'Not built')"

# Python Poetry targets
.PHONY: poetry-install poetry-update poetry-shell poetry-run-py poetry-test-py poetry-lint-py poetry-format-py

poetry-install:
	@echo "Installing Python dependencies with Poetry..."
	poetry install

poetry-update:
	@echo "Updating Python dependencies with Poetry..."
	poetry update

poetry-shell:
	@echo "Activating Poetry shell..."
	poetry shell

poetry-run-py:
	@echo "Running Python mock server with Poetry..."
	poetry run python app/mock_server.py --port 5000

poetry-test-py:
	@echo "Running Python tests with Poetry..."
	poetry run pytest

poetry-lint-py:
	@echo "Running Python linting with Poetry..."
	poetry run flake8 app/
	poetry run mypy app/

poetry-format-py:
	@echo "Formatting Python code with Poetry..."
	poetry run black app/
	poetry run isort app/

# Development setup
.PHONY: setup setup-py
setup: deps lint-install watch-install security-install docs-install
	@echo "Go development environment setup complete!"
	@echo "Available commands:"
	@echo "  make build     - Build the Go application"
	@echo "  make run       - Run Go server with default config"
	@echo "  make test      - Run Go tests"
	@echo "  make lint      - Run Go linter"
	@echo "  make watch     - Watch files and auto-restart"
	@echo "  make help      - Show all available commands"

setup-py: poetry-install
	@echo "Python development environment setup complete!"
	@echo "Available commands:"
	@echo "  make poetry-run-py    - Run Python mock server"
	@echo "  make poetry-test-py   - Run Python tests"
	@echo "  make poetry-lint-py   - Run Python linting"
	@echo "  make poetry-format-py - Format Python code"

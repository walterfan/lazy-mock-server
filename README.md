# Lazy Mock Server

A flexible HTTP mock server available in both Python and Go implementations, supporting customized HTTP status codes, response content types, and dynamic responses with real-time management capabilities.

## 🚀 Features

### Core Capabilities
- **Flexible Route Matching**: Exact paths, wildcard patterns (`*`), and parameter-based routing
- **Custom Status Codes**: Configure any HTTP status code (200, 404, 500, etc.)
- **Multiple Content Types**: JSON, plain text, XML, HTML, and custom content types
- **Dynamic Responses**: Use placeholders to inject request data into responses
- **Custom Headers**: Add custom HTTP headers to responses
- **YAML Configuration**: Easy-to-read and maintain configuration format

### Advanced Features (Go Version)
- **Real-time Management**: Web UI and REST API for runtime configuration
- **Thread-Safe Operations**: Concurrent request handling during updates
- **Parameter Matching**: Route requests based on query parameters
- **Dynamic Placeholders**: `{method}`, `{path}`, `{query}` replacement
- **Configuration Persistence**: Save changes back to YAML files
- **HTTPS/TLS Support**: Serve mocks over HTTPS with custom certificates

## 📋 Quick Start

### Python Version

#### Using Poetry (Recommended)
```bash
# Install Poetry if not already installed
curl -sSL https://install.python-poetry.org | python3 -

# Install dependencies
poetry install

# Start server
./start_py_mock_server_poetry.sh

# Or manually with Poetry
poetry run python app/mock_server.py --port 9000

# Run tests
poetry run pytest

# Format code
poetry run black app/
poetry run isort app/
```

#### Using pip (Legacy)
```bash
# Install dependencies
pip install -r requirements.txt

# Start server
./start_py_mock_server.sh

# Or manually
cd app && python mock_server.py --port 9000
```

### Go Version
```bash
# Using Makefile (recommended)
make build          # Build the binary
make run            # Run with default config
make run-advanced   # Run with advanced config
make test           # Run tests
make lint           # Run linter
make all            # Build, test, and lint

# Manual build and run
go mod tidy
go build -o mock-server main.go
./mock-server -port 8080

# With custom config
./mock-server -config config/mock_config.yaml -port 9000

# With HTTPS/TLS enabled
./generate_cert.sh  # Generate self-signed certificate
./mock-server -tls -cert server.crt -key server.key -port 8443
```

## ⚙️ Configuration

Both versions use the same YAML configuration format:

```yaml
routes:
  - path: "/api/users"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    headers:
      X-Custom-Header: "mock-value"
      Cache-Control: "no-cache"
    response:
      users:
        - id: 1
          name: "John Doe"
        - id: 2
          name: "Jane Smith"
```

### Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `path` | URL path to match (supports `*` wildcards) | Required |
| `method` | HTTP method (GET, POST, PUT, DELETE, etc.) | Required |
| `status_code` | HTTP status code to return | 200 |
| `content_type` | Response content type | application/json |
| `headers` | Custom HTTP headers | Optional |
| `parameters` | Query parameters that must match | Optional |
| `response` | Response body (string, object, or array) | Required |

## 🎯 Examples

### 1. Basic Text Response
```yaml
routes:
  - path: "/v1/metadata/sn"
    method: "GET"
    status_code: 200
    content_type: "text/plain"
    response: "8CPKW77"
```

**Test:**
```bash
curl http://localhost:8080/v1/metadata/sn
# Output: 8CPKW77
```

### 2. JSON Response with Custom Headers
```yaml
routes:
  - path: "/api/users"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    headers:
      X-Total-Count: "100"
      Cache-Control: "max-age=3600"
    response:
      users:
        - id: 1
          name: "John Doe"
          email: "john@example.com"
```

**Test:**
```bash
curl -v http://localhost:8080/api/users
# Returns JSON with custom headers
```

### 3. Error Responses
```yaml
routes:
  - path: "/api/error/404"
    method: "GET"
    status_code: 404
    content_type: "application/json"
    response:
      error: "Not Found"
      code: 404
      message: "The requested resource was not found"

  - path: "/api/error/500"
    method: "GET"
    status_code: 500
    content_type: "application/json"
    response:
      error: "Internal Server Error"
      code: 500
      message: "Something went wrong on the server"
```

**Test:**
```bash
curl -w "%{http_code}" http://localhost:8080/api/error/404
# Returns 404 status with JSON error
```

### 4. Wildcard Path Matching (Go Version)
```yaml
routes:
  - path: "/api/users/*"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response:
      message: "User endpoint matched"
      path: "{path}"
```

**Test:**
```bash
curl http://localhost:8080/api/users/123
curl http://localhost:8080/api/users/profile
# Both match the same route
```

### 5. Parameter-Based Routing (Go Version)
```yaml
routes:
  - path: "/api/search"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    parameters:
      type: "user"
    response:
      results: "User search results"

  - path: "/api/search"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    parameters:
      type: "product"
    response:
      results: "Product search results"
```

**Test:**
```bash
curl "http://localhost:8080/api/search?type=user"
# Returns user search results

curl "http://localhost:8080/api/search?type=product"
# Returns product search results
```

### 6. Dynamic Response Placeholders (Go Version)
```yaml
routes:
  - path: "/api/echo"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response:
      method: "{method}"
      path: "{path}"
      query: "{query}"
      timestamp: "2023-12-01T10:00:00Z"
```

**Test:**
```bash
curl "http://localhost:8080/api/echo?name=john&age=30"
# Returns:
# {
#   "method": "GET",
#   "path": "/api/echo",
#   "query": "name=john&age=30",
#   "timestamp": "2023-12-01T10:00:00Z"
# }
```

### 7. Different Content Types
```yaml
routes:
  # XML Response
  - path: "/api/xml"
    method: "GET"
    status_code: 200
    content_type: "application/xml"
    response: |
      <?xml version="1.0" encoding="UTF-8"?>
      <response>
        <status>success</status>
        <data>XML response example</data>
      </response>

  # HTML Response
  - path: "/api/html"
    method: "GET"
    status_code: 200
    content_type: "text/html"
    response: |
      <!DOCTYPE html>
      <html>
      <head><title>Mock Response</title></head>
      <body><h1>Hello from Mock Server!</h1></body>
      </html>

  # Plain Text
  - path: "/api/text"
    method: "GET"
    status_code: 200
    content_type: "text/plain"
    response: "This is a plain text response"
```

### 8. POST Request Handling
```yaml
routes:
  - path: "/api/users"
    method: "POST"
    status_code: 201
    content_type: "application/json"
    response:
      id: 123
      message: "User created successfully"
      timestamp: "2023-12-01T10:00:00Z"
```

**Test:**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

## 🌐 Web UI Management (Go Version Only)

The Go version includes a beautiful web interface for real-time configuration management.

### Access Web UI
```bash
# Start server
./mock-server -port 8080

# Open browser
open http://localhost:8080/_mock/ui
```

### Web UI Features
- **Dashboard**: Route count and server status
- **Add Routes**: Form-based route creation with validation
- **Edit Routes**: Click edit to modify existing routes
- **Delete Routes**: Remove routes with confirmation
- **Live Preview**: Response body preview for each route
- **Save Configuration**: Persist changes to YAML file
- **Real-time Updates**: Changes are immediately active

## 🔧 Management API (Go Version Only)

### API Endpoints
```bash
# Get all routes
curl http://localhost:8080/_mock/routes

# Add new route
curl -X POST http://localhost:8080/_mock/routes \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/test",
    "method": "GET",
    "status_code": 200,
    "content_type": "text/plain",
    "response": "Test response"
  }'

# Update route
curl -X PUT http://localhost:8080/_mock/routes/api/test \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/test",
    "method": "GET",
    "status_code": 404,
    "content_type": "application/json",
    "response": {"error": "Not found"}
  }'

# Delete route
curl -X DELETE http://localhost:8080/_mock/routes/api/test

# Save configuration to file
curl -X POST http://localhost:8080/_mock/config
```

## 📊 Command Line Options

### Python Version
```bash
# Using Poetry
poetry run python app/mock_server.py --port 9000

# Using pip
python app/mock_server.py --port 9000
```

### Go Version
```bash
./mock-server -port 8080 -config app/mock_response.yaml
```

| Option | Description | Default |
|--------|-------------|---------|
| `-port` | Port to listen on | 8080 (Go), 5000 (Python) |
| `-config` | Path to YAML configuration | app/mock_response.yaml |
| `-log-level` | Log level (debug, info, warn, error) | info |
| `-tls` | Enable HTTPS/TLS | false |
| `-cert` | Path to TLS certificate file | server.crt |
| `-key` | Path to TLS private key file | server.key |
| `-version` | Show version information | - |

## 🔒 HTTPS/TLS Support (Go Version)

The Go version supports HTTPS/TLS for secure mock server deployments. This is useful for testing applications that require secure connections or validating SSL/TLS certificate handling.

### Quick Start with HTTPS

#### 1. Generate a Self-Signed Certificate (for testing)
```bash
# Generate certificate for localhost
./generate_cert.sh

# Or generate for a custom domain
./generate_cert.sh example.com
```

This creates two files:
- `server.crt` - TLS certificate
- `server.key` - Private key

#### 2. Start the Server with HTTPS
```bash
# Using default certificate files
./mock-server -tls

# Using custom certificate files
./mock-server -tls -cert /path/to/cert.crt -key /path/to/key.key

# With custom port
./mock-server -tls -port 8443 -cert server.crt -key server.key
```

#### 3. Test HTTPS Endpoints
```bash
# Accept self-signed certificate with -k flag
curl -k https://localhost:8080/api/users

# Or with explicit certificate
curl --cacert server.crt https://localhost:8080/api/users
```

### Using Production Certificates

For production or staging environments, use certificates from a Certificate Authority (CA):

```bash
# Using Let's Encrypt certificates
./mock-server -tls \
  -cert /etc/letsencrypt/live/yourdomain.com/fullchain.pem \
  -key /etc/letsencrypt/live/yourdomain.com/privkey.pem \
  -port 443
```

### HTTPS Configuration Examples

#### Mixed HTTP/HTTPS Setup
Run separate instances for HTTP and HTTPS:

```bash
# Terminal 1: HTTP server
./mock-server -port 8080 -config config/mock_config.yaml

# Terminal 2: HTTPS server
./mock-server -tls -port 8443 -config config/mock_config.yaml
```

#### Behind a Reverse Proxy
Use the mock server with nginx or other reverse proxies:

```nginx
server {
    listen 443 ssl;
    server_name api.example.com;
    
    ssl_certificate /path/to/server.crt;
    ssl_certificate_key /path/to/server.key;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Self-Signed Certificate Notes

⚠️ **Important**: Self-signed certificates should only be used for testing/development:

- **Browsers**: Will show security warnings that must be manually accepted
- **cURL**: Use `-k` or `--insecure` flag to skip certificate verification
- **Production**: Always use certificates from trusted Certificate Authorities

### Troubleshooting HTTPS

**Certificate not found error:**
```bash
# Check if certificate files exist
ls -la server.crt server.key

# Generate new certificates if missing
./generate_cert.sh
```

**Permission denied on port 443:**
```bash
# Use sudo for privileged ports (< 1024)
sudo ./mock-server -tls -port 443

# Or use a non-privileged port
./mock-server -tls -port 8443
```

**Certificate expired:**
```bash
# Check certificate expiration
openssl x509 -in server.crt -noout -dates

# Generate new certificate
./generate_cert.sh
```

## 🧪 Testing Examples

### Complete Testing Workflow
```bash
# 1. Start the server
./mock-server -port 8080

# 2. Test basic endpoints
curl http://localhost:8080/v1/metadata/sn
curl http://localhost:8080/v1/metadata/instanceid

# 3. Test different HTTP methods
curl -X GET http://localhost:8080/api/users
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John"}'

# 4. Test error responses
curl -w "%{http_code}" http://localhost:8080/api/error/404
curl -w "%{http_code}" http://localhost:8080/api/error/500

# 5. Test with parameters (Go version)
curl "http://localhost:8080/api/search?type=user"
curl "http://localhost:8080/api/search?type=product"

# 6. Test dynamic responses (Go version)
curl "http://localhost:8080/api/echo?name=test&id=123"

# 7. Test different content types
curl -H "Accept: application/xml" http://localhost:8080/api/xml
curl -H "Accept: text/html" http://localhost:8080/api/html
```

## 🔄 Version Comparison

| Feature | Python Version | Go Version |
|---------|----------------|------------|
| Basic Mocking | ✅ | ✅ |
| Custom Status Codes | ✅ | ✅ |
| Content Types | ✅ | ✅ |
| YAML Configuration | ✅ | ✅ |
| HTTPS/TLS Support | ❌ | ✅ |
| Wildcard Paths | ❌ | ✅ |
| Parameter Matching | ❌ | ✅ |
| Dynamic Placeholders | ❌ | ✅ |
| Web UI Management | ❌ | ✅ |
| REST API Management | ❌ | ✅ |
| Runtime Config Changes | ❌ | ✅ |
| Performance | Good | Excellent |
| Memory Usage | Higher | Lower |
| Deployment | Requires Python | Single Binary |

## 🛠️ Development Tools

### Makefile Commands
```bash
# Development
make help           # Show all available commands
make setup          # Setup development environment
make deps           # Download dependencies
make build          # Build the binary
make run            # Run with default config
make run-dev        # Run in development mode
make watch          # Watch files and auto-restart (requires air)

# Testing and Quality
make test           # Run tests
make test-coverage  # Run tests with coverage report
make test-verbose   # Run tests with verbose output
make benchmark      # Run benchmarks
make lint           # Run linter (golangci-lint)
make fmt            # Format Go code
make vet            # Run go vet
make security       # Run security scan (gosec)
make check          # Run all checks (fmt, vet, lint, test)

# Build and Release
make build-all      # Build for all platforms
make release        # Create release artifacts
make docker-build   # Build Docker image
make docker-run     # Run Docker container

# Utilities
make clean          # Clean build artifacts
make stats          # Show project statistics
```

## 🚀 Deployment

### Python Version
```bash
# Production deployment with Poetry
poetry install --only=main
poetry run gunicorn -w 4 -b 0.0.0.0:9000 app.mock_server:app

# Production deployment with pip (legacy)
pip install -r requirements.txt
gunicorn -w 4 -b 0.0.0.0:9000 app.mock_server:app
```

### Go Version
```bash
# Using Makefile
make build          # Build optimized binary
make release        # Create release for all platforms

# Manual build for production
go build -ldflags="-s -w" -o mock-server main.go

# Run in production
./mock-server -port 8080 -config production.yaml

# Docker deployment
make docker-build   # Build Docker image
make docker-run     # Run Docker container

# Or manually
docker build -t mock-server .
docker run -p 8080:8080 -v $(pwd)/config:/config mock-server
```

## 📝 TODO / Roadmap

- [x] ~~Setup Expectations by REST API~~ ✅ (Go version)
- [x] ~~Add a web UI for setup expectation and checking mock history~~ ✅ (Go version)
- [ ] Add SQLite DB support for authentication and history
- [ ] Request/Response logging and history
- [ ] Mock response templates
- [ ] Load testing capabilities
- [ ] Docker compose setup
- [ ] Kubernetes deployment manifests
- [ ] Prometheus metrics endpoint
- [ ] Health check endpoints

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and examples
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
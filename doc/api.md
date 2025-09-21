# Mock Server API Documentation

## Overview

The Go mock server now includes a comprehensive management API and web UI for dynamically managing mock routes without restarting the server.

## Management API Endpoints

All management endpoints are prefixed with `/_mock/` to avoid conflicts with your mock routes.

### 1. Get All Routes
**GET** `/_mock/routes`

Returns all currently configured mock routes.

**Response:**
```json
{
  "routes": [
    {
      "path": "/api/users",
      "method": "GET",
      "status_code": 200,
      "content_type": "application/json",
      "response": {"users": []}
    }
  ],
  "count": 1
}
```

### 2. Add New Route
**POST** `/_mock/routes`

Adds a new mock route to the server.

**Request Body:**
```json
{
  "path": "/api/new-endpoint",
  "method": "GET",
  "status_code": 200,
  "content_type": "application/json",
  "response": {"message": "Hello World"}
}
```

**Response:**
```json
{
  "message": "Route added successfully",
  "route": {
    "path": "/api/new-endpoint",
    "method": "GET",
    "status_code": 200,
    "content_type": "application/json",
    "response": {"message": "Hello World"}
  }
}
```

### 3. Update Existing Route
**PUT** `/_mock/routes{path}`

Updates an existing route. The `{path}` should be the exact path of the route to update.

**Example:** `PUT /_mock/routes/api/users`

**Request Body:**
```json
{
  "path": "/api/users",
  "method": "GET",
  "status_code": 201,
  "content_type": "application/json",
  "response": {"message": "Updated response"}
}
```

### 4. Delete Route
**DELETE** `/_mock/routes{path}`

Deletes a route by its path.

**Example:** `DELETE /_mock/routes/api/users`

**Response:**
```json
{
  "message": "Route deleted successfully"
}
```

### 5. Get Current Configuration
**GET** `/_mock/config`

Returns the complete current configuration.

**Response:**
```json
{
  "routes": [
    // ... all routes
  ]
}
```

### 6. Save Configuration to File
**POST** `/_mock/config`

Saves the current in-memory configuration back to the YAML file.

**Response:**
```json
{
  "message": "Configuration saved successfully"
}
```

### 7. Web UI
**GET** `/_mock/ui`

Serves the web-based management interface.

## Web UI Features

Access the web UI at: `http://localhost:8080/_mock/ui`

### Features:
- **Dashboard**: Shows route count and server status
- **Add Routes**: Form-based route creation with validation
- **Edit Routes**: Click edit to modify existing routes
- **Delete Routes**: Remove routes with confirmation
- **Live Preview**: See response body preview for each route
- **Save Configuration**: Persist changes to YAML file
- **Real-time Updates**: Changes are immediately active

### UI Capabilities:
1. **Form Validation**: Automatic JSON validation for JSON responses
2. **Visual Feedback**: Color-coded HTTP methods and status codes
3. **Responsive Design**: Works on desktop and mobile
4. **Error Handling**: Clear error messages for failed operations
5. **Auto-refresh**: Automatic route list updates after changes

## Usage Examples

### Using curl to manage routes:

```bash
# Get all routes
curl http://localhost:8080/_mock/routes

# Add a new route
curl -X POST http://localhost:8080/_mock/routes \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/test",
    "method": "GET", 
    "status_code": 200,
    "content_type": "text/plain",
    "response": "Test response"
  }'

# Update a route
curl -X PUT http://localhost:8080/_mock/routes/api/test \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/test",
    "method": "GET",
    "status_code": 404,
    "content_type": "application/json",
    "response": {"error": "Not found"}
  }'

# Delete a route
curl -X DELETE http://localhost:8080/_mock/routes/api/test

# Save configuration
curl -X POST http://localhost:8080/_mock/config
```

### Using the Web UI:

1. Start your server: `./mock-server -port 8080`
2. Open browser: `http://localhost:8080/_mock/ui`
3. Use the form to add/edit routes
4. Click "Save Configuration" to persist changes

## Route Configuration Options

When creating or updating routes, you can specify:

- **path**: URL path (supports wildcards with `*`)
- **method**: HTTP method (GET, POST, PUT, DELETE, PATCH)
- **status_code**: HTTP status code (100-599)
- **content_type**: Response content type
- **response**: Response body (string, object, or array)
- **headers**: Custom HTTP headers (optional)
- **parameters**: Query parameter requirements (optional)

## Thread Safety

All management operations are thread-safe using read-write mutexes, allowing:
- Concurrent read access to routes during normal operation
- Exclusive write access during configuration changes
- No downtime during route updates

## Persistence

- Changes are made in-memory first for immediate effect
- Use `POST /_mock/config` or the "Save Configuration" button to persist changes
- Configuration is saved back to the original YAML file
- Server restart will load the saved configuration

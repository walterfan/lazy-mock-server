package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Route represents a mock route configuration
type Route struct {
	Path        string            `yaml:"path" json:"path"`
	Method      string            `yaml:"method" json:"method"`
	StatusCode  int               `yaml:"status_code" json:"status_code"`
	ContentType string            `yaml:"content_type" json:"content_type"`
	Response    interface{}       `yaml:"response" json:"response"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Parameters  map[string]string `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// GetJSONSafeResponse returns a JSON-safe version of the response
func (r *Route) GetJSONSafeResponse() interface{} {
	return convertYAMLToJSON(r.Response)
}

// convertYAMLToJSON converts YAML interface{} types to JSON-compatible types
func convertYAMLToJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		// Convert map[interface{}]interface{} to map[string]interface{}
		result := make(map[string]interface{})
		for key, value := range v {
			if strKey, ok := key.(string); ok {
				result[strKey] = convertYAMLToJSON(value)
			}
		}
		return result
	case []interface{}:
		// Convert slice elements recursively
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertYAMLToJSON(item)
		}
		return result
	default:
		// Return as-is for basic types (string, int, bool, etc.)
		return v
	}
}

// Config represents the entire mock configuration
type Config struct {
	Routes []Route `yaml:"routes" json:"routes"`
}

// Manager handles configuration loading, saving, and management
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load loads the configuration from the file
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", m.configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	m.config = &config
	return nil
}

// Save saves the current configuration to the file
func (m *Manager) Save() error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", m.configPath, err)
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// SetConfig sets the configuration
func (m *Manager) SetConfig(config *Config) {
	m.config = config
}

// GetRoutes returns all routes
func (m *Manager) GetRoutes() []Route {
	if m.config == nil {
		return nil
	}
	return m.config.Routes
}

// AddRoute adds a new route to the configuration
func (m *Manager) AddRoute(route Route) {
	if m.config == nil {
		m.config = &Config{}
	}
	m.config.Routes = append(m.config.Routes, route)
}

// UpdateRoute updates an existing route by path and method
func (m *Manager) UpdateRoute(path, method string, newRoute Route) error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	for i, route := range m.config.Routes {
		if route.Path == path && route.Method == method {
			m.config.Routes[i] = newRoute
			return nil
		}
	}

	return fmt.Errorf("route not found: %s %s", method, path)
}

// DeleteRoute removes a route by path and method
func (m *Manager) DeleteRoute(path, method string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	for i, route := range m.config.Routes {
		if route.Path == path && route.Method == method {
			m.config.Routes = append(m.config.Routes[:i], m.config.Routes[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("route not found: %s %s", method, path)
}

// DeleteRouteByPath removes all routes with the specified path
func (m *Manager) DeleteRouteByPath(path string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	var newRoutes []Route
	found := false

	for _, route := range m.config.Routes {
		if route.Path != path {
			newRoutes = append(newRoutes, route)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("no routes found with path: %s", path)
	}

	m.config.Routes = newRoutes
	return nil
}

// FindRoute finds a route by path and method
func (m *Manager) FindRoute(path, method string) (*Route, error) {
	if m.config == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	for _, route := range m.config.Routes {
		if route.Path == path && route.Method == method {
			return &route, nil
		}
	}

	return nil, fmt.Errorf("route not found: %s %s", method, path)
}

// GetRouteCount returns the number of configured routes
func (m *Manager) GetRouteCount() int {
	if m.config == nil {
		return 0
	}
	return len(m.config.Routes)
}

// ValidateRoute validates a route configuration
func (m *Manager) ValidateRoute(route Route) error {
	if route.Path == "" {
		return fmt.Errorf("route path cannot be empty")
	}

	if route.Method == "" {
		return fmt.Errorf("route method cannot be empty")
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}

	if !validMethods[route.Method] {
		return fmt.Errorf("invalid HTTP method: %s", route.Method)
	}

	if route.StatusCode < 100 || route.StatusCode > 599 {
		return fmt.Errorf("invalid status code: %d", route.StatusCode)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// SetConfigPath sets the configuration file path
func (m *Manager) SetConfigPath(path string) {
	m.configPath = path
}

// Clone creates a deep copy of the configuration
func (m *Manager) Clone() *Config {
	if m.config == nil {
		return nil
	}

	// Use YAML marshal/unmarshal for deep copying
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return nil
	}

	var cloned Config
	if err := yaml.Unmarshal(data, &cloned); err != nil {
		return nil
	}

	return &cloned
}

// LoadFromBytes loads configuration from byte data
func (m *Manager) LoadFromBytes(data []byte) error {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	m.config = &config
	return nil
}

// ToBytes converts the configuration to YAML bytes
func (m *Manager) ToBytes() ([]byte, error) {
	if m.config == nil {
		return nil, fmt.Errorf("no configuration to convert")
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	return data, nil
}

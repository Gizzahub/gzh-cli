package plugins

import (
	"context"
	"time"
)

// Plugin represents the core interface that all plugins must implement
type Plugin interface {
	// Metadata returns plugin information
	GetMetadata() PluginMetadata

	// Lifecycle management
	Initialize(ctx context.Context, config PluginConfig) error
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
	Cleanup(ctx context.Context) error

	// Health check
	HealthCheck(ctx context.Context) error
}

// PluginMetadata contains plugin identification and configuration
type PluginMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Homepage    string   `json:"homepage,omitempty"`
	License     string   `json:"license,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Capabilities and requirements
	Capabilities []string           `json:"capabilities"`
	Requirements PluginRequirements `json:"requirements"`

	// Configuration schema
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`

	// Runtime information
	LoadTime time.Time    `json:"load_time"`
	LastUsed time.Time    `json:"last_used"`
	Status   PluginStatus `json:"status"`
}

// PluginRequirements defines what the plugin needs to function
type PluginRequirements struct {
	MinGZVersion    string   `json:"min_gz_version"`
	MaxGZVersion    string   `json:"max_gz_version,omitempty"`
	Dependencies    []string `json:"dependencies,omitempty"`
	Permissions     []string `json:"permissions"`
	SupportedOS     []string `json:"supported_os,omitempty"`
	RequiredEnvVars []string `json:"required_env_vars,omitempty"`
}

// PluginConfig holds runtime configuration for a plugin
type PluginConfig struct {
	Settings    map[string]interface{} `json:"settings"`
	Environment map[string]string      `json:"environment"`
	Permissions PermissionSet          `json:"permissions"`
	Limits      ResourceLimits         `json:"limits"`
}

// PermissionSet defines what the plugin is allowed to do
type PermissionSet struct {
	FileSystem   FileSystemPermissions `json:"filesystem"`
	Network      NetworkPermissions    `json:"network"`
	SystemCalls  []string              `json:"system_calls"`
	Environment  []string              `json:"environment_vars"`
	APIEndpoints []string              `json:"api_endpoints"`
}

// FileSystemPermissions controls file access
type FileSystemPermissions struct {
	ReadPaths    []string `json:"read_paths"`
	WritePaths   []string `json:"write_paths"`
	ExecutePaths []string `json:"execute_paths"`
	DenyPaths    []string `json:"deny_paths"`
}

// NetworkPermissions controls network access
type NetworkPermissions struct {
	AllowedHosts []string `json:"allowed_hosts"`
	AllowedPorts []int    `json:"allowed_ports"`
	BlockedHosts []string `json:"blocked_hosts"`
	MaxRequests  int      `json:"max_requests_per_minute"`
}

// ResourceLimits constrains plugin resource usage
type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb"`
	MaxCPUPercent    float64       `json:"max_cpu_percent"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	MaxFileHandles   int           `json:"max_file_handles"`
}

// PluginStatus represents the current state of a plugin
type PluginStatus string

const (
	PluginStatusUnloaded  PluginStatus = "unloaded"
	PluginStatusLoading   PluginStatus = "loading"
	PluginStatusReady     PluginStatus = "ready"
	PluginStatusRunning   PluginStatus = "running"
	PluginStatusError     PluginStatus = "error"
	PluginStatusDisabled  PluginStatus = "disabled"
	PluginStatusUnloading PluginStatus = "unloading"
)

// PluginAPI provides services to plugins from the host application
type PluginAPI interface {
	// Logging
	GetLogger(pluginName string) Logger

	// Configuration
	GetConfig(key string) (interface{}, error)
	SetConfig(key string, value interface{}) error

	// Event system
	EmitEvent(event Event) error
	SubscribeToEvent(eventType string, handler EventHandler) error

	// File operations (within permission boundaries)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error

	// HTTP client (within permission boundaries)
	HTTPRequest(method, url string, headers map[string]string, body []byte) (*HTTPResponse, error)

	// Inter-plugin communication
	CallPlugin(pluginName string, method string, args map[string]interface{}) (interface{}, error)

	// Host information
	GetHostInfo() HostInfo
}

// Logger interface for plugin logging
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
}

// Event represents a system event
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventHandler processes events
type EventHandler func(event Event) error

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}

// HostInfo provides information about the host system
type HostInfo struct {
	GZVersion    string `json:"gz_version"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	WorkingDir   string `json:"working_dir"`
	ConfigDir    string `json:"config_dir"`
	PluginDir    string `json:"plugin_dir"`
}

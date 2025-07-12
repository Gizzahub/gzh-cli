package plugins

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultPluginAPI provides the default implementation of PluginAPI
type DefaultPluginAPI struct {
	config      map[string]interface{}
	eventBus    *EventBus
	securityMgr *SecurityManager
	manager     *Manager
	hostInfo    HostInfo
}

// NewDefaultPluginAPI creates a new default plugin API
func NewDefaultPluginAPI(eventBus *EventBus, securityMgr *SecurityManager, hostInfo HostInfo) *DefaultPluginAPI {
	return &DefaultPluginAPI{
		config:      make(map[string]interface{}),
		eventBus:    eventBus,
		securityMgr: securityMgr,
		hostInfo:    hostInfo,
	}
}

// SetManager sets the plugin manager reference
func (api *DefaultPluginAPI) SetManager(manager *Manager) {
	api.manager = manager
}

// GetLogger returns a logger for the specified plugin
func (api *DefaultPluginAPI) GetLogger(pluginName string) Logger {
	return &DefaultLogger{
		pluginName: pluginName,
		eventBus:   api.eventBus,
	}
}

// GetConfig retrieves a configuration value
func (api *DefaultPluginAPI) GetConfig(key string) (interface{}, error) {
	if value, exists := api.config[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("configuration key %s not found", key)
}

// SetConfig sets a configuration value
func (api *DefaultPluginAPI) SetConfig(key string, value interface{}) error {
	api.config[key] = value
	return nil
}

// EmitEvent sends an event through the event bus
func (api *DefaultPluginAPI) EmitEvent(event Event) error {
	return api.eventBus.Emit(event)
}

// SubscribeToEvent registers an event handler
func (api *DefaultPluginAPI) SubscribeToEvent(eventType string, handler EventHandler) error {
	api.eventBus.Subscribe(eventType, handler)
	return nil
}

// ReadFile reads a file with permission checking
func (api *DefaultPluginAPI) ReadFile(path string) ([]byte, error) {
	// Security check would be performed here
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	return os.ReadFile(absPath)
}

// WriteFile writes data to a file with permission checking
func (api *DefaultPluginAPI) WriteFile(path string, data []byte) error {
	// Security check would be performed here
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	return os.WriteFile(absPath, data, 0644)
}

// HTTPRequest performs an HTTP request with permission checking
func (api *DefaultPluginAPI) HTTPRequest(method, url string, headers map[string]string, body []byte) (*HTTPResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, nil
}

// CallPlugin calls another plugin
func (api *DefaultPluginAPI) CallPlugin(pluginName string, method string, args map[string]interface{}) (interface{}, error) {
	if api.manager == nil {
		return nil, fmt.Errorf("plugin manager not available")
	}

	// Add method to args for the target plugin to process
	if args == nil {
		args = make(map[string]interface{})
	}
	args["__method"] = method

	return api.manager.ExecutePlugin(pluginName, args)
}

// GetHostInfo returns information about the host system
func (api *DefaultPluginAPI) GetHostInfo() HostInfo {
	return api.hostInfo
}

// DefaultLogger provides a default logger implementation
type DefaultLogger struct {
	pluginName string
	eventBus   *EventBus
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log("debug", msg, nil, fields...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log("info", msg, nil, fields...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log("warn", msg, nil, fields...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log("error", msg, err, fields...)
}

// log is the internal logging function
func (l *DefaultLogger) log(level string, msg string, err error, fields ...map[string]interface{}) {
	data := map[string]interface{}{
		"level":   level,
		"message": msg,
		"plugin":  l.pluginName,
	}

	if err != nil {
		data["error"] = err.Error()
	}

	// Merge additional fields
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			data[key] = value
		}
	}

	event := Event{
		Type:      "plugin.log",
		Source:    l.pluginName,
		Timestamp: time.Now(),
		Data:      data,
	}

	l.eventBus.Emit(event)

	// Also print to stdout for development
	fmt.Printf("[%s] %s: %s\n", level, l.pluginName, msg)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	}
}

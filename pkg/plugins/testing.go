package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	"testing"
	"time"
)

// TestFramework provides utilities for testing plugins
type TestFramework struct {
	tempDir  string
	plugins  map[string]*TestPlugin
	eventBus *EventBus
	api      *TestPluginAPI
	mu       sync.RWMutex
}

// TestPlugin wraps a plugin for testing
type TestPlugin struct {
	Plugin      Plugin
	Handle      *plugin.Plugin
	TempDir     string
	Events      []Event
	CallHistory []TestCall
	mu          sync.RWMutex
}

// TestCall records a plugin method call
type TestCall struct {
	Method    string
	Args      map[string]interface{}
	Result    interface{}
	Error     error
	Timestamp time.Time
}

// TestPluginAPI implements PluginAPI for testing
type TestPluginAPI struct {
	config    map[string]interface{}
	files     map[string][]byte
	httpMocks map[string]*HTTPResponse
	eventBus  *EventBus
	logger    *TestLogger
	hostInfo  HostInfo
	mu        sync.RWMutex
}

// TestLogger implements Logger for testing
type TestLogger struct {
	pluginName string
	logs       []TestLogEntry
	mu         sync.RWMutex
}

// TestLogEntry represents a log entry for testing
type TestLogEntry struct {
	Level     string
	Message   string
	Error     error
	Fields    map[string]interface{}
	Timestamp time.Time
}

// NewTestFramework creates a new test framework
func NewTestFramework(t *testing.T) *TestFramework {
	tempDir := t.TempDir()

	eventBus := NewEventBus()
	api := &TestPluginAPI{
		config:    make(map[string]interface{}),
		files:     make(map[string][]byte),
		httpMocks: make(map[string]*HTTPResponse),
		eventBus:  eventBus,
		logger:    &TestLogger{logs: make([]TestLogEntry, 0)},
		hostInfo: HostInfo{
			GZVersion:    "test-1.0.0",
			OS:           "test-os",
			Architecture: "test-arch",
			WorkingDir:   tempDir,
			ConfigDir:    filepath.Join(tempDir, ".config"),
			PluginDir:    filepath.Join(tempDir, "plugins"),
		},
	}

	return &TestFramework{
		tempDir:  tempDir,
		plugins:  make(map[string]*TestPlugin),
		eventBus: eventBus,
		api:      api,
	}
}

// LoadTestPlugin loads a plugin for testing (from Plugin interface, not .so file)
func (tf *TestFramework) LoadTestPlugin(name string, pluginFunc func() Plugin) error {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	pluginObj := pluginFunc()
	metadata := pluginObj.GetMetadata()

	if metadata.Name != name {
		return fmt.Errorf("plugin name mismatch: expected %s, got %s", name, metadata.Name)
	}

	testPlugin := &TestPlugin{
		Plugin:      pluginObj,
		TempDir:     filepath.Join(tf.tempDir, name),
		Events:      make([]Event, 0),
		CallHistory: make([]TestCall, 0),
	}

	// Create plugin temp directory
	if err := os.MkdirAll(testPlugin.TempDir, 0o755); err != nil {
		return fmt.Errorf("failed to create plugin temp dir: %w", err)
	}

	// Initialize plugin
	config := PluginConfig{
		Settings:    make(map[string]interface{}),
		Environment: make(map[string]string),
		Permissions: PermissionSet{
			FileSystem: FileSystemPermissions{
				ReadPaths:  []string{testPlugin.TempDir},
				WritePaths: []string{testPlugin.TempDir},
			},
		},
		Limits: ResourceLimits{
			MaxMemoryMB:      100,
			MaxExecutionTime: 30 * time.Second,
		},
	}

	ctx := context.WithValue(context.Background(), "plugin_api", tf.api)
	if err := pluginObj.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	tf.plugins[name] = testPlugin
	return nil
}

// ExecutePlugin executes a plugin method and records the call
func (tf *TestFramework) ExecutePlugin(name string, args map[string]interface{}) (interface{}, error) {
	tf.mu.RLock()
	testPlugin, exists := tf.plugins[name]
	tf.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("test plugin %s not found", name)
	}

	testPlugin.mu.Lock()
	defer testPlugin.mu.Unlock()

	call := TestCall{
		Method:    "Execute",
		Args:      args,
		Timestamp: time.Now(),
	}

	result, err := testPlugin.Plugin.Execute(context.Background(), args)
	call.Result = result
	call.Error = err

	testPlugin.CallHistory = append(testPlugin.CallHistory, call)

	return result, err
}

// GetPluginCallHistory returns the call history for a plugin
func (tf *TestFramework) GetPluginCallHistory(name string) ([]TestCall, error) {
	tf.mu.RLock()
	testPlugin, exists := tf.plugins[name]
	tf.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("test plugin %s not found", name)
	}

	testPlugin.mu.RLock()
	defer testPlugin.mu.RUnlock()

	// Return a copy
	history := make([]TestCall, len(testPlugin.CallHistory))
	copy(history, testPlugin.CallHistory)

	return history, nil
}

// HealthCheckPlugin runs health check on a plugin
func (tf *TestFramework) HealthCheckPlugin(name string) error {
	tf.mu.RLock()
	testPlugin, exists := tf.plugins[name]
	tf.mu.RUnlock()

	if !exists {
		return fmt.Errorf("test plugin %s not found", name)
	}

	return testPlugin.Plugin.HealthCheck(context.Background())
}

// CleanupPlugin cleans up a plugin
func (tf *TestFramework) CleanupPlugin(name string) error {
	tf.mu.RLock()
	testPlugin, exists := tf.plugins[name]
	tf.mu.RUnlock()

	if !exists {
		return fmt.Errorf("test plugin %s not found", name)
	}

	return testPlugin.Plugin.Cleanup(context.Background())
}

// CleanupAll cleans up all plugins
func (tf *TestFramework) CleanupAll() error {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	var errors []error
	for name, testPlugin := range tf.plugins {
		if err := testPlugin.Plugin.Cleanup(context.Background()); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// SetupMockFile sets up a mock file for testing
func (tf *TestFramework) SetupMockFile(path string, content []byte) {
	tf.api.mu.Lock()
	defer tf.api.mu.Unlock()
	tf.api.files[path] = content
}

// SetupHTTPMock sets up an HTTP mock response
func (tf *TestFramework) SetupHTTPMock(url string, response *HTTPResponse) {
	tf.api.mu.Lock()
	defer tf.api.mu.Unlock()
	tf.api.httpMocks[url] = response
}

// SetConfig sets a configuration value for testing
func (tf *TestFramework) SetConfig(key string, value interface{}) {
	tf.api.mu.Lock()
	defer tf.api.mu.Unlock()
	tf.api.config[key] = value
}

// EmitTestEvent emits an event for testing
func (tf *TestFramework) EmitTestEvent(event Event) error {
	return tf.eventBus.Emit(event)
}

// GetLogs returns all logs from the test logger
func (tf *TestFramework) GetLogs() []TestLogEntry {
	tf.api.logger.mu.RLock()
	defer tf.api.logger.mu.RUnlock()

	logs := make([]TestLogEntry, len(tf.api.logger.logs))
	copy(logs, tf.api.logger.logs)

	return logs
}

// GetLogsByLevel returns logs filtered by level
func (tf *TestFramework) GetLogsByLevel(level string) []TestLogEntry {
	logs := tf.GetLogs()
	var filtered []TestLogEntry

	for _, log := range logs {
		if log.Level == level {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

// TestPluginAPI implementation

// GetLogger returns a test logger
func (api *TestPluginAPI) GetLogger(pluginName string) Logger {
	api.mu.Lock()
	defer api.mu.Unlock()

	return &TestLogger{
		pluginName: pluginName,
		logs:       api.logger.logs,
	}
}

// GetConfig retrieves a configuration value
func (api *TestPluginAPI) GetConfig(key string) (interface{}, error) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	if value, exists := api.config[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("configuration key %s not found", key)
}

// SetConfig sets a configuration value
func (api *TestPluginAPI) SetConfig(key string, value interface{}) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.config[key] = value
	return nil
}

// EmitEvent emits an event
func (api *TestPluginAPI) EmitEvent(event Event) error {
	return api.eventBus.Emit(event)
}

// SubscribeToEvent subscribes to an event
func (api *TestPluginAPI) SubscribeToEvent(eventType string, handler EventHandler) error {
	api.eventBus.Subscribe(eventType, handler)
	return nil
}

// ReadFile reads a mock file
func (api *TestPluginAPI) ReadFile(path string) ([]byte, error) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	if content, exists := api.files[path]; exists {
		return content, nil
	}

	return nil, fmt.Errorf("mock file %s not found", path)
}

// WriteFile writes to a mock file
func (api *TestPluginAPI) WriteFile(path string, data []byte) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.files[path] = data
	return nil
}

// HTTPRequest performs a mock HTTP request
func (api *TestPluginAPI) HTTPRequest(method, url string, headers map[string]string, body []byte) (*HTTPResponse, error) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	if response, exists := api.httpMocks[url]; exists {
		return response, nil
	}

	return nil, fmt.Errorf("no mock response configured for %s", url)
}

// CallPlugin calls another plugin (not implemented in test)
func (api *TestPluginAPI) CallPlugin(pluginName string, method string, args map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("inter-plugin calls not supported in test framework")
}

// GetHostInfo returns test host information
func (api *TestPluginAPI) GetHostInfo() HostInfo {
	api.mu.RLock()
	defer api.mu.RUnlock()
	return api.hostInfo
}

// TestLogger implementation

// Debug logs a debug message
func (l *TestLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log("debug", msg, nil, fields...)
}

// Info logs an info message
func (l *TestLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log("info", msg, nil, fields...)
}

// Warn logs a warning message
func (l *TestLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log("warn", msg, nil, fields...)
}

// Error logs an error message
func (l *TestLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log("error", msg, err, fields...)
}

// log is the internal logging function
func (l *TestLogger) log(level string, msg string, err error, fields ...map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := TestLogEntry{
		Level:     level,
		Message:   msg,
		Error:     err,
		Fields:    make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Merge fields
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			entry.Fields[key] = value
		}
	}

	l.logs = append(l.logs, entry)
}

// Helper functions for testing

// AssertPluginCalled checks if a plugin was called with specific args
func AssertPluginCalled(t *testing.T, tf *TestFramework, pluginName string, expectedArgs map[string]interface{}) {
	history, err := tf.GetPluginCallHistory(pluginName)
	if err != nil {
		t.Fatalf("Failed to get call history: %v", err)
	}

	if len(history) == 0 {
		t.Fatalf("Plugin %s was not called", pluginName)
	}

	// Check if any call matches expected args
	for _, call := range history {
		if mapsEqual(call.Args, expectedArgs) {
			return
		}
	}

	t.Fatalf("Plugin %s was not called with expected args %v", pluginName, expectedArgs)
}

// AssertLogContains checks if logs contain a specific message
func AssertLogContains(t *testing.T, tf *TestFramework, level, message string) {
	logs := tf.GetLogsByLevel(level)

	for _, log := range logs {
		if log.Message == message {
			return
		}
	}

	t.Fatalf("Log level %s does not contain message: %s", level, message)
}

// mapsEqual compares two maps for equality
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists || valueA != valueB {
			return false
		}
	}

	return true
}

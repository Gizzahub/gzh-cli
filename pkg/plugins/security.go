package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// SecurityManager handles plugin security and sandboxing
type SecurityManager struct {
	enabled    bool
	executions map[string]*ExecutionContext
	mu         sync.RWMutex
}

// ExecutionContext tracks plugin execution state for security
type ExecutionContext struct {
	PluginName    string
	StartTime     time.Time
	ResourceUsage ResourceUsage
	PermissionSet PermissionSet
	Active        bool
}

// ResourceUsage tracks current resource consumption
type ResourceUsage struct {
	MemoryMB    float64
	CPUPercent  float64
	FileHandles int
	NetworkReqs int
	LastReqTime time.Time
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(enabled bool) *SecurityManager {
	return &SecurityManager{
		enabled:    enabled,
		executions: make(map[string]*ExecutionContext),
	}
}

// CheckExecution validates if a plugin execution should be allowed
func (sm *SecurityManager) CheckExecution(instance *PluginInstance, args map[string]interface{}) error {
	if !sm.enabled {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	pluginName := instance.Metadata.Name

	// Check if plugin is already running and limits concurrent executions
	if ctx, exists := sm.executions[pluginName]; exists && ctx.Active {
		return fmt.Errorf("plugin %s is already executing", pluginName)
	}

	// Validate permissions
	if err := sm.validatePermissions(instance, args); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Check resource limits
	if err := sm.checkResourceLimits(instance); err != nil {
		return fmt.Errorf("resource limit exceeded: %w", err)
	}

	// Create execution context
	sm.executions[pluginName] = &ExecutionContext{
		PluginName:    pluginName,
		StartTime:     time.Now(),
		PermissionSet: instance.Config.Permissions,
		Active:        true,
		ResourceUsage: ResourceUsage{
			LastReqTime: time.Now(),
		},
	}

	return nil
}

// CompleteExecution marks an execution as complete
func (sm *SecurityManager) CompleteExecution(pluginName string) {
	if !sm.enabled {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if ctx, exists := sm.executions[pluginName]; exists {
		ctx.Active = false
	}
}

// validatePermissions checks if the plugin has required permissions
func (sm *SecurityManager) validatePermissions(instance *PluginInstance, args map[string]interface{}) error {
	permissions := instance.Config.Permissions

	// Check file system permissions if file operations are requested
	if filePath, exists := args["file_path"]; exists {
		if path, ok := filePath.(string); ok {
			if err := sm.checkFileAccess(path, permissions.FileSystem); err != nil {
				return err
			}
		}
	}

	// Check network permissions if network operations are requested
	if host, exists := args["host"]; exists {
		if hostStr, ok := host.(string); ok {
			if err := sm.checkNetworkAccess(hostStr, permissions.Network); err != nil {
				return err
			}
		}
	}

	// Check environment variable access
	if envVar, exists := args["env_var"]; exists {
		if envVarStr, ok := envVar.(string); ok {
			if err := sm.checkEnvAccess(envVarStr, permissions.Environment); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkFileAccess validates file system access
func (sm *SecurityManager) checkFileAccess(path string, fs FileSystemPermissions) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check deny list first
	for _, denyPath := range fs.DenyPaths {
		if matched, _ := filepath.Match(denyPath, absPath); matched {
			return fmt.Errorf("access to %s is explicitly denied", absPath)
		}
	}

	// Check if path is in allowed read paths
	for _, allowedPath := range fs.ReadPaths {
		if strings.HasPrefix(absPath, allowedPath) {
			return nil
		}
	}

	// Check if path is in allowed write paths
	for _, allowedPath := range fs.WritePaths {
		if strings.HasPrefix(absPath, allowedPath) {
			return nil
		}
	}

	return fmt.Errorf("access to %s not permitted", absPath)
}

// checkNetworkAccess validates network access
func (sm *SecurityManager) checkNetworkAccess(host string, net NetworkPermissions) error {
	// Check blocked hosts first
	for _, blockedHost := range net.BlockedHosts {
		if matched, _ := filepath.Match(blockedHost, host); matched {
			return fmt.Errorf("access to %s is blocked", host)
		}
	}

	// Check allowed hosts
	if len(net.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range net.AllowedHosts {
			if matched, _ := filepath.Match(allowedHost, host); matched {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access to %s not permitted", host)
		}
	}

	// Check rate limiting
	sm.mu.RLock()
	// Rate limiting logic would go here
	sm.mu.RUnlock()

	return nil
}

// checkEnvAccess validates environment variable access
func (sm *SecurityManager) checkEnvAccess(envVar string, allowedVars []string) error {
	for _, allowed := range allowedVars {
		if allowed == envVar || allowed == "*" {
			return nil
		}
	}
	return fmt.Errorf("access to environment variable %s not permitted", envVar)
}

// checkResourceLimits validates current resource usage
func (sm *SecurityManager) checkResourceLimits(instance *PluginInstance) error {
	limits := instance.Config.Limits

	// Check memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	currentMemMB := float64(memStats.Alloc) / 1024 / 1024

	if limits.MaxMemoryMB > 0 && currentMemMB > float64(limits.MaxMemoryMB) {
		return fmt.Errorf("memory usage %.2fMB exceeds limit %dMB", currentMemMB, limits.MaxMemoryMB)
	}

	// Additional resource checks would go here
	return nil
}

// MonitorExecution starts monitoring a plugin execution
func (sm *SecurityManager) MonitorExecution(ctx context.Context, pluginName string) {
	if !sm.enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				sm.CompleteExecution(pluginName)
				return
			case <-ticker.C:
				sm.updateResourceUsage(pluginName)
			}
		}
	}()
}

// updateResourceUsage updates resource usage statistics
func (sm *SecurityManager) updateResourceUsage(pluginName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ctx, exists := sm.executions[pluginName]
	if !exists || !ctx.Active {
		return
	}

	// Update memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ctx.ResourceUsage.MemoryMB = float64(memStats.Alloc) / 1024 / 1024

	// Check execution timeout
	limits := sm.getPluginLimits(pluginName)
	if limits.MaxExecutionTime > 0 && time.Since(ctx.StartTime) > limits.MaxExecutionTime {
		// Execution timeout - this would trigger cleanup
		ctx.Active = false
	}
}

// getPluginLimits retrieves resource limits for a plugin
func (sm *SecurityManager) getPluginLimits(pluginName string) ResourceLimits {
	// This would typically retrieve limits from configuration
	return ResourceLimits{
		MaxMemoryMB:      512,
		MaxCPUPercent:    50.0,
		MaxExecutionTime: 5 * time.Minute,
		MaxFileHandles:   100,
	}
}

// CreateSandbox creates an isolated execution environment
func (sm *SecurityManager) CreateSandbox(pluginName string, permissions PermissionSet) (*Sandbox, error) {
	if !sm.enabled {
		return nil, nil
	}

	sandbox := &Sandbox{
		PluginName:  pluginName,
		Permissions: permissions,
		TempDir:     filepath.Join(os.TempDir(), "gz-plugin-"+pluginName),
		Active:      true,
	}

	// Create temporary directory for the plugin
	if err := os.MkdirAll(sandbox.TempDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	return sandbox, nil
}

// Sandbox represents an isolated execution environment
type Sandbox struct {
	PluginName  string
	Permissions PermissionSet
	TempDir     string
	Active      bool
}

// Cleanup removes sandbox resources
func (s *Sandbox) Cleanup() error {
	if s.TempDir != "" {
		return os.RemoveAll(s.TempDir)
	}
	return nil
}

// ValidateFileOperation checks if a file operation is allowed in the sandbox
func (s *Sandbox) ValidateFileOperation(operation string, path string) error {
	if !s.Active {
		return fmt.Errorf("sandbox is not active")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	switch operation {
	case "read":
		for _, allowedPath := range s.Permissions.FileSystem.ReadPaths {
			if strings.HasPrefix(absPath, allowedPath) {
				return nil
			}
		}
		// Allow access to sandbox temp directory
		if strings.HasPrefix(absPath, s.TempDir) {
			return nil
		}
		return fmt.Errorf("read access to %s not permitted", absPath)

	case "write":
		for _, allowedPath := range s.Permissions.FileSystem.WritePaths {
			if strings.HasPrefix(absPath, allowedPath) {
				return nil
			}
		}
		// Allow access to sandbox temp directory
		if strings.HasPrefix(absPath, s.TempDir) {
			return nil
		}
		return fmt.Errorf("write access to %s not permitted", absPath)

	default:
		return fmt.Errorf("unknown file operation: %s", operation)
	}
}

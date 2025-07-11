package netenv

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CommandPool provides optimized command execution with caching and parallel processing
type CommandPool struct {
	maxWorkers int
	cache      map[string]*CachedResult
	cacheMutex sync.RWMutex
	workerPool chan struct{}
	resultChan chan *CommandResult
	ctx        context.Context
	cancel     context.CancelFunc
}

// CachedResult stores command execution results with TTL
type CachedResult struct {
	Output    []byte
	Error     error
	Timestamp time.Time
	TTL       time.Duration
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Command   string
	Output    []byte
	Error     error
	Duration  time.Duration
	FromCache bool
}

// NewCommandPool creates a new optimized command pool
func NewCommandPool(maxWorkers int) *CommandPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &CommandPool{
		maxWorkers: maxWorkers,
		cache:      make(map[string]*CachedResult),
		workerPool: make(chan struct{}, maxWorkers),
		resultChan: make(chan *CommandResult, maxWorkers*2),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Close gracefully shuts down the command pool
func (cp *CommandPool) Close() {
	cp.cancel()
}

// ExecuteCommand executes a single command with caching
func (cp *CommandPool) ExecuteCommand(name string, args ...string) *CommandResult {
	cmdStr := fmt.Sprintf("%s %v", name, args)

	// Check cache first
	if result := cp.getCachedResult(cmdStr); result != nil {
		return &CommandResult{
			Command:   cmdStr,
			Output:    result.Output,
			Error:     result.Error,
			FromCache: true,
		}
	}

	start := time.Now()
	cmd := exec.CommandContext(cp.ctx, name, args...)
	output, err := cmd.Output()
	duration := time.Since(start)

	result := &CommandResult{
		Command:  cmdStr,
		Output:   output,
		Error:    err,
		Duration: duration,
	}

	// Cache the result with default TTL of 30 seconds
	cp.setCachedResult(cmdStr, &CachedResult{
		Output:    output,
		Error:     err,
		Timestamp: time.Now(),
		TTL:       30 * time.Second,
	})

	return result
}

// ExecuteBatch executes multiple commands in parallel
func (cp *CommandPool) ExecuteBatch(commands []Command) []*CommandResult {
	results := make([]*CommandResult, len(commands))
	var wg sync.WaitGroup

	for i, cmd := range commands {
		wg.Add(1)
		go func(index int, command Command) {
			defer wg.Done()

			// Acquire worker slot
			cp.workerPool <- struct{}{}
			defer func() { <-cp.workerPool }()

			results[index] = cp.ExecuteCommand(command.Name, command.Args...)
		}(i, cmd)
	}

	wg.Wait()
	return results
}

// Command represents a command to be executed
type Command struct {
	Name string
	Args []string
	TTL  time.Duration
}

// getCachedResult retrieves a cached result if valid
func (cp *CommandPool) getCachedResult(cmdStr string) *CachedResult {
	cp.cacheMutex.RLock()
	defer cp.cacheMutex.RUnlock()

	result, exists := cp.cache[cmdStr]
	if !exists {
		return nil
	}

	// Check if cache is expired
	if time.Since(result.Timestamp) > result.TTL {
		return nil
	}

	return result
}

// setCachedResult stores a result in cache
func (cp *CommandPool) setCachedResult(cmdStr string, result *CachedResult) {
	cp.cacheMutex.Lock()
	defer cp.cacheMutex.Unlock()

	cp.cache[cmdStr] = result
}

// ClearCache removes all cached results
func (cp *CommandPool) ClearCache() {
	cp.cacheMutex.Lock()
	defer cp.cacheMutex.Unlock()

	cp.cache = make(map[string]*CachedResult)
}

// ClearExpiredCache removes expired cache entries
func (cp *CommandPool) ClearExpiredCache() {
	cp.cacheMutex.Lock()
	defer cp.cacheMutex.Unlock()

	now := time.Now()
	for key, result := range cp.cache {
		if now.Sub(result.Timestamp) > result.TTL {
			delete(cp.cache, key)
		}
	}
}

// GetCacheStats returns cache performance statistics
func (cp *CommandPool) GetCacheStats() map[string]interface{} {
	cp.cacheMutex.RLock()
	defer cp.cacheMutex.RUnlock()

	total := len(cp.cache)
	expired := 0
	now := time.Now()

	for _, result := range cp.cache {
		if now.Sub(result.Timestamp) > result.TTL {
			expired++
		}
	}

	return map[string]interface{}{
		"total_entries":   total,
		"expired_entries": expired,
		"valid_entries":   total - expired,
		"cache_size":      len(cp.cache),
	}
}

// OptimizedVPNManager provides performance-optimized VPN operations
type OptimizedVPNManager struct {
	commandPool *CommandPool
	connections map[string]*VPNConnectionState
	mutex       sync.RWMutex
}

// VPNConnectionState tracks the state of VPN connections
type VPNConnectionState struct {
	Name      string
	Type      string
	Status    string
	LastCheck time.Time
	TTL       time.Duration
}

// NewOptimizedVPNManager creates a new optimized VPN manager
func NewOptimizedVPNManager() *OptimizedVPNManager {
	return &OptimizedVPNManager{
		commandPool: NewCommandPool(5),
		connections: make(map[string]*VPNConnectionState),
	}
}

// Close shuts down the VPN manager
func (ovm *OptimizedVPNManager) Close() {
	ovm.commandPool.Close()
}

// ConnectVPNBatch connects multiple VPN connections in parallel
func (ovm *OptimizedVPNManager) ConnectVPNBatch(configs []vpnConfig) error {
	commands := make([]Command, 0, len(configs))

	for _, config := range configs {
		switch config.Type {
		case "networkmanager":
			commands = append(commands, Command{
				Name: "nmcli",
				Args: []string{"connection", "up", config.Name},
				TTL:  10 * time.Second,
			})
		case "openvpn":
			commands = append(commands, Command{
				Name: "systemctl",
				Args: []string{"start", fmt.Sprintf("openvpn@%s", config.Name)},
				TTL:  15 * time.Second,
			})
		case "wireguard":
			commands = append(commands, Command{
				Name: "wg-quick",
				Args: []string{"up", config.Name},
				TTL:  10 * time.Second,
			})
		}
	}

	results := ovm.commandPool.ExecuteBatch(commands)

	// Process results and update connection states
	for i, result := range results {
		config := configs[i]
		status := "disconnected"
		if result.Error == nil {
			status = "connected"
		}

		ovm.updateConnectionState(config.Name, config.Type, status)

		if result.Error != nil {
			return fmt.Errorf("failed to connect VPN '%s': %w", config.Name, result.Error)
		}
	}

	return nil
}

// GetVPNStatusBatch retrieves status for multiple VPN connections efficiently
func (ovm *OptimizedVPNManager) GetVPNStatusBatch(names []string) (map[string]string, error) {
	results := make(map[string]string)

	// Check cached states first
	freshNames := make([]string, 0, len(names))
	for _, name := range names {
		if state := ovm.getCachedConnectionState(name); state != nil {
			results[name] = state.Status
		} else {
			freshNames = append(freshNames, name)
		}
	}

	if len(freshNames) == 0 {
		return results, nil
	}

	// Batch query for remaining connections
	commands := []Command{
		{
			Name: "nmcli",
			Args: []string{"-t", "-f", "NAME,STATE", "connection", "show"},
			TTL:  5 * time.Second,
		},
		{
			Name: "systemctl",
			Args: []string{"list-units", "--type=service", "--state=active", "openvpn@*"},
			TTL:  5 * time.Second,
		},
	}

	commandResults := ovm.commandPool.ExecuteBatch(commands)

	// Parse NetworkManager results
	if len(commandResults) > 0 && commandResults[0].Error == nil {
		ovm.parseNetworkManagerStatus(string(commandResults[0].Output), freshNames, results)
	}

	// Parse OpenVPN results
	if len(commandResults) > 1 && commandResults[1].Error == nil {
		ovm.parseOpenVPNStatus(string(commandResults[1].Output), freshNames, results)
	}

	return results, nil
}

// updateConnectionState updates the cached connection state
func (ovm *OptimizedVPNManager) updateConnectionState(name, vpnType, status string) {
	ovm.mutex.Lock()
	defer ovm.mutex.Unlock()

	ovm.connections[name] = &VPNConnectionState{
		Name:      name,
		Type:      vpnType,
		Status:    status,
		LastCheck: time.Now(),
		TTL:       30 * time.Second,
	}
}

// getCachedConnectionState retrieves cached connection state if valid
func (ovm *OptimizedVPNManager) getCachedConnectionState(name string) *VPNConnectionState {
	ovm.mutex.RLock()
	defer ovm.mutex.RUnlock()

	state, exists := ovm.connections[name]
	if !exists {
		return nil
	}

	if time.Since(state.LastCheck) > state.TTL {
		return nil
	}

	return state
}

// parseNetworkManagerStatus parses nmcli output and updates results
func (ovm *OptimizedVPNManager) parseNetworkManagerStatus(output string, names []string, results map[string]string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) >= 2 {
			connName := fields[0]
			connState := fields[1]

			for _, name := range names {
				if connName == name {
					status := "disconnected"
					if strings.Contains(connState, "activated") {
						status = "connected"
					}
					results[name] = status
					ovm.updateConnectionState(name, "networkmanager", status)
				}
			}
		}
	}
}

// parseOpenVPNStatus parses systemctl output for OpenVPN services
func (ovm *OptimizedVPNManager) parseOpenVPNStatus(output string, names []string, results map[string]string) {
	for _, name := range names {
		serviceName := fmt.Sprintf("openvpn@%s.service", name)
		if strings.Contains(output, serviceName) {
			results[name] = "connected"
			ovm.updateConnectionState(name, "openvpn", "connected")
		} else if _, exists := results[name]; !exists {
			results[name] = "disconnected"
			ovm.updateConnectionState(name, "openvpn", "disconnected")
		}
	}
}

// OptimizedDNSManager provides performance-optimized DNS operations
type OptimizedDNSManager struct {
	commandPool    *CommandPool
	lastConfig     *DNSConfig
	lastConfigTime time.Time
	configTTL      time.Duration
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Servers   []string
	Interface string
	Method    string
}

// NewOptimizedDNSManager creates a new optimized DNS manager
func NewOptimizedDNSManager() *OptimizedDNSManager {
	return &OptimizedDNSManager{
		commandPool: NewCommandPool(3),
		configTTL:   15 * time.Second,
	}
}

// Close shuts down the DNS manager
func (odm *OptimizedDNSManager) Close() {
	odm.commandPool.Close()
}

// SetDNSServersBatch sets DNS servers efficiently with minimal system calls
func (odm *OptimizedDNSManager) SetDNSServersBatch(configs []DNSConfig) error {
	commands := make([]Command, 0, len(configs))

	for _, config := range configs {
		if config.Interface == "" {
			// Auto-detect interface efficiently using cached route info
			if iface := odm.getCachedDefaultInterface(); iface != "" {
				config.Interface = iface
			}
		}

		args := append([]string{"dns", config.Interface}, config.Servers...)
		commands = append(commands, Command{
			Name: "resolvectl",
			Args: args,
			TTL:  10 * time.Second,
		})
	}

	results := odm.commandPool.ExecuteBatch(commands)

	for i, result := range results {
		if result.Error != nil {
			return fmt.Errorf("failed to set DNS for interface '%s': %w", configs[i].Interface, result.Error)
		}
	}

	return nil
}

// getCachedDefaultInterface returns the cached default network interface
func (odm *OptimizedDNSManager) getCachedDefaultInterface() string {
	result := odm.commandPool.ExecuteCommand("ip", "route", "get", "1.1.1.1")
	if result.Error != nil {
		return ""
	}

	fields := strings.Fields(string(result.Output))
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}

	return ""
}

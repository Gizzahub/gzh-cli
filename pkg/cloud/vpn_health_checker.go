package cloud

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// VPNHealthChecker performs health checks on VPN connections
type VPNHealthChecker struct {
	connection   *VPNConnection
	isHealthy    bool
	successCount int
	failureCount int
	lastCheck    *HealthCheckResult
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
}

// NewVPNHealthChecker creates a new VPN health checker
func NewVPNHealthChecker(connection *VPNConnection) *VPNHealthChecker {
	return &VPNHealthChecker{
		connection: connection,
		isHealthy:  true, // Assume healthy initially
	}
}

// Start starts the health checker
func (hc *VPNHealthChecker) Start(ctx context.Context) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.running {
		return
	}

	hc.ctx, hc.cancel = context.WithCancel(ctx)
	hc.running = true

	go hc.runHealthChecks()
}

// Stop stops the health checker
func (hc *VPNHealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if !hc.running {
		return
	}

	if hc.cancel != nil {
		hc.cancel()
	}

	hc.running = false
}

// IsHealthy returns the current health status
func (hc *VPNHealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.isHealthy
}

// GetLastCheckResult returns the last health check result
func (hc *VPNHealthChecker) GetLastCheckResult() *HealthCheckResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.lastCheck
}

// runHealthChecks runs the health check loop
func (hc *VPNHealthChecker) runHealthChecks() {
	if hc.connection.HealthCheck == nil || !hc.connection.HealthCheck.Enabled {
		return
	}

	ticker := time.NewTicker(hc.connection.HealthCheck.Interval)
	defer ticker.Stop()

	// Perform initial health check
	hc.performHealthCheck()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performHealthCheck()
		}
	}
}

// performHealthCheck performs a single health check
func (hc *VPNHealthChecker) performHealthCheck() {
	config := hc.connection.HealthCheck
	if config == nil || !config.Enabled {
		return
	}

	// Test connectivity to all targets
	var results []*HealthCheckResult
	for _, target := range config.Targets {
		result := hc.checkTarget(target, config.Timeout)
		results = append(results, result)
	}

	// Determine overall health based on results
	successCount := 0
	var lastResult *HealthCheckResult
	for _, result := range results {
		if result.Success {
			successCount++
		}
		lastResult = result // Use last result for reporting
	}

	success := successCount > 0 // At least one target is reachable

	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.lastCheck = lastResult

	if success {
		hc.successCount++
		hc.failureCount = 0

		// Consider healthy if we have enough successful checks
		if hc.successCount >= config.SuccessThreshold {
			hc.isHealthy = true
		}
	} else {
		hc.failureCount++
		hc.successCount = 0

		// Consider unhealthy if we have too many failures
		if hc.failureCount >= config.FailureThreshold {
			hc.isHealthy = false
		}
	}
}

// checkTarget checks connectivity to a specific target
func (hc *VPNHealthChecker) checkTarget(target string, timeout time.Duration) *HealthCheckResult {
	start := time.Now()

	result := &HealthCheckResult{
		Timestamp: start,
		Target:    target,
	}

	// Try to establish a connection
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		// If TCP fails, try ICMP ping
		if err := hc.pingTarget(target, timeout); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to reach %s: %v", target, err)
			return result
		}
	} else {
		conn.Close()
	}

	result.Success = true
	result.Latency = time.Since(start)
	return result
}

// pingTarget performs an ICMP ping to the target
func (hc *VPNHealthChecker) pingTarget(target string, timeout time.Duration) error {
	// Parse target to get host (remove port if present)
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		host = target // Target doesn't have port
	}

	// Resolve the address
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", host, err)
	}

	// Create ICMP connection
	conn, err := net.DialTimeout("ip4:icmp", addr.String(), timeout)
	if err != nil {
		return fmt.Errorf("failed to create ICMP connection: %w", err)
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(timeout))

	// Send ping packet
	ping := []byte{8, 0, 0, 0, 0, 0, 0, 0}

	// Calculate checksum
	checksum := hc.calculateChecksum(ping)
	ping[2] = byte(checksum >> 8)
	ping[3] = byte(checksum)

	// Send ping
	_, err = conn.Write(ping)
	if err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// Read response
	response := make([]byte, 1500)
	_, err = conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read ping response: %w", err)
	}

	return nil
}

// calculateChecksum calculates ICMP checksum
func (hc *VPNHealthChecker) calculateChecksum(data []byte) uint16 {
	var sum uint32

	// Sum all 16-bit words
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 + uint32(data[i+1])
	}

	// Add left-over byte, if any
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	// Add the carry
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	// One's complement
	return uint16(^sum)
}

// GetHealthCheckStats returns health check statistics
func (hc *VPNHealthChecker) GetHealthCheckStats() map[string]interface{} {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	stats := map[string]interface{}{
		"is_healthy":    hc.isHealthy,
		"success_count": hc.successCount,
		"failure_count": hc.failureCount,
		"running":       hc.running,
	}

	if hc.lastCheck != nil {
		stats["last_check"] = map[string]interface{}{
			"timestamp": hc.lastCheck.Timestamp,
			"success":   hc.lastCheck.Success,
			"latency":   hc.lastCheck.Latency,
			"target":    hc.lastCheck.Target,
			"error":     hc.lastCheck.Error,
		}
	}

	return stats
}

// VPNHealthMonitor monitors health of multiple VPN connections
type VPNHealthMonitor struct {
	checkers map[string]*VPNHealthChecker
	mu       sync.RWMutex
}

// NewVPNHealthMonitor creates a new VPN health monitor
func NewVPNHealthMonitor() *VPNHealthMonitor {
	return &VPNHealthMonitor{
		checkers: make(map[string]*VPNHealthChecker),
	}
}

// AddConnection adds a connection to monitor
func (monitor *VPNHealthMonitor) AddConnection(connection *VPNConnection) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	if connection.HealthCheck != nil && connection.HealthCheck.Enabled {
		monitor.checkers[connection.Name] = NewVPNHealthChecker(connection)
	}
}

// RemoveConnection removes a connection from monitoring
func (monitor *VPNHealthMonitor) RemoveConnection(name string) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	if checker, exists := monitor.checkers[name]; exists {
		checker.Stop()
		delete(monitor.checkers, name)
	}
}

// StartMonitoring starts monitoring for a specific connection
func (monitor *VPNHealthMonitor) StartMonitoring(ctx context.Context, name string) error {
	monitor.mu.RLock()
	checker, exists := monitor.checkers[name]
	monitor.mu.RUnlock()

	if !exists {
		return fmt.Errorf("health checker not found for connection: %s", name)
	}

	checker.Start(ctx)
	return nil
}

// StopMonitoring stops monitoring for a specific connection
func (monitor *VPNHealthMonitor) StopMonitoring(name string) {
	monitor.mu.RLock()
	checker, exists := monitor.checkers[name]
	monitor.mu.RUnlock()

	if exists {
		checker.Stop()
	}
}

// GetHealthStatus returns health status for all monitored connections
func (monitor *VPNHealthMonitor) GetHealthStatus() map[string]bool {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	status := make(map[string]bool, len(monitor.checkers))
	for name, checker := range monitor.checkers {
		status[name] = checker.IsHealthy()
	}

	return status
}

// GetDetailedHealthStatus returns detailed health status for all monitored connections
func (monitor *VPNHealthMonitor) GetDetailedHealthStatus() map[string]map[string]interface{} {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	status := make(map[string]map[string]interface{}, len(monitor.checkers))
	for name, checker := range monitor.checkers {
		status[name] = checker.GetHealthCheckStats()
	}

	return status
}

// StartAllMonitoring starts monitoring for all registered connections
func (monitor *VPNHealthMonitor) StartAllMonitoring(ctx context.Context) {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	for _, checker := range monitor.checkers {
		checker.Start(ctx)
	}
}

// StopAllMonitoring stops monitoring for all connections
func (monitor *VPNHealthMonitor) StopAllMonitoring() {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	for _, checker := range monitor.checkers {
		checker.Stop()
	}
}

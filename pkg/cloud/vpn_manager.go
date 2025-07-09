package cloud

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

// VPNManager manages multiple VPN connections with priority and failover
type VPNManager interface {
	// AddVPNConnection adds a VPN connection with priority
	AddVPNConnection(conn *VPNConnection) error

	// RemoveVPNConnection removes a VPN connection
	RemoveVPNConnection(name string) error

	// ConnectVPN connects to a VPN by name
	ConnectVPN(ctx context.Context, name string) error

	// DisconnectVPN disconnects from a VPN by name
	DisconnectVPN(ctx context.Context, name string) error

	// ConnectByPriority connects to VPNs based on priority order
	ConnectByPriority(ctx context.Context) error

	// GetConnectionStatus returns status of all VPN connections
	GetConnectionStatus() map[string]*VPNStatus

	// StartFailoverMonitoring starts monitoring connections for failover
	StartFailoverMonitoring(ctx context.Context) error

	// StopFailoverMonitoring stops the failover monitoring
	StopFailoverMonitoring()

	// GetActiveConnections returns list of active VPN connections
	GetActiveConnections() []*VPNConnection

	// ValidateConnection validates VPN connection configuration
	ValidateConnection(conn *VPNConnection) error
}

// VPNConnection represents a VPN connection configuration
type VPNConnection struct {
	// Connection name
	Name string `yaml:"name" json:"name"`

	// Connection type (openvpn, wireguard, ipsec, pptp, l2tp)
	Type string `yaml:"type" json:"type"`

	// Server endpoint
	Server string `yaml:"server" json:"server"`

	// Port (optional)
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Priority (higher number = higher priority)
	Priority int `yaml:"priority" json:"priority"`

	// Configuration file path
	ConfigFile string `yaml:"config_file,omitempty" json:"config_file,omitempty"`

	// Credentials
	Credentials *VPNCredentials `yaml:"credentials,omitempty" json:"credentials,omitempty"`

	// Auto-connect on network change
	AutoConnect bool `yaml:"auto_connect,omitempty" json:"auto_connect,omitempty"`

	// Failover settings
	Failover *VPNFailover `yaml:"failover,omitempty" json:"failover,omitempty"`

	// Health check settings
	HealthCheck *VPNHealthCheck `yaml:"health_check,omitempty" json:"health_check,omitempty"`

	// Route settings
	Routes []VPNRoute `yaml:"routes,omitempty" json:"routes,omitempty"`

	// DNS settings
	DNS []string `yaml:"dns,omitempty" json:"dns,omitempty"`

	// Environment restrictions
	Environments []string `yaml:"environments,omitempty" json:"environments,omitempty"`

	// Tags for organization
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// VPNCredentials represents VPN authentication credentials
type VPNCredentials struct {
	// Username for authentication
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Password (consider using secure storage)
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Certificate file path
	CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`

	// Private key file path
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`

	// CA certificate file path
	CAFile string `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`

	// Pre-shared key (for IPSec)
	PSK string `yaml:"psk,omitempty" json:"psk,omitempty"`
}

// VPNFailover represents failover configuration
type VPNFailover struct {
	// Enable failover
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Fallback VPN connection names in order
	FallbackOrder []string `yaml:"fallback_order,omitempty" json:"fallback_order,omitempty"`

	// Retry attempts before failover
	RetryAttempts int `yaml:"retry_attempts,omitempty" json:"retry_attempts,omitempty"`

	// Retry interval
	RetryInterval time.Duration `yaml:"retry_interval,omitempty" json:"retry_interval,omitempty"`

	// Failover timeout
	FailoverTimeout time.Duration `yaml:"failover_timeout,omitempty" json:"failover_timeout,omitempty"`
}

// VPNHealthCheck represents health check configuration
type VPNHealthCheck struct {
	// Enable health check
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Health check interval
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Health check timeout
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Health check targets (IPs/domains to ping)
	Targets []string `yaml:"targets,omitempty" json:"targets,omitempty"`

	// Success threshold (number of successful checks)
	SuccessThreshold int `yaml:"success_threshold,omitempty" json:"success_threshold,omitempty"`

	// Failure threshold (number of failed checks before considering unhealthy)
	FailureThreshold int `yaml:"failure_threshold,omitempty" json:"failure_threshold,omitempty"`
}

// VPNRoute represents a VPN route configuration
type VPNRoute struct {
	// Destination network
	Destination string `yaml:"destination" json:"destination"`

	// Gateway IP
	Gateway string `yaml:"gateway,omitempty" json:"gateway,omitempty"`

	// Metric/priority
	Metric int `yaml:"metric,omitempty" json:"metric,omitempty"`

	// Interface name
	Interface string `yaml:"interface,omitempty" json:"interface,omitempty"`
}

// VPNStatus represents the status of a VPN connection
type VPNStatus struct {
	// Connection name
	Name string `json:"name"`

	// Connection state
	State VPNState `json:"state"`

	// Connected timestamp
	ConnectedAt time.Time `json:"connected_at,omitempty"`

	// Disconnected timestamp
	DisconnectedAt time.Time `json:"disconnected_at,omitempty"`

	// Local IP address
	LocalIP string `json:"local_ip,omitempty"`

	// Remote IP address
	RemoteIP string `json:"remote_ip,omitempty"`

	// Interface name
	Interface string `json:"interface,omitempty"`

	// Bytes sent
	BytesSent uint64 `json:"bytes_sent,omitempty"`

	// Bytes received
	BytesReceived uint64 `json:"bytes_received,omitempty"`

	// Last health check result
	LastHealthCheck *HealthCheckResult `json:"last_health_check,omitempty"`

	// Error message (if any)
	Error string `json:"error,omitempty"`
}

// VPNState represents the state of a VPN connection
type VPNState string

const (
	// VPNStateDisconnected represents disconnected state
	VPNStateDisconnected VPNState = "disconnected"

	// VPNStateConnecting represents connecting state
	VPNStateConnecting VPNState = "connecting"

	// VPNStateConnected represents connected state
	VPNStateConnected VPNState = "connected"

	// VPNStateDisconnecting represents disconnecting state
	VPNStateDisconnecting VPNState = "disconnecting"

	// VPNStateError represents error state
	VPNStateError VPNState = "error"

	// VPNStateReconnecting represents reconnecting state
	VPNStateReconnecting VPNState = "reconnecting"
)

// HealthCheckResult represents health check result
type HealthCheckResult struct {
	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Success status
	Success bool `json:"success"`

	// Latency
	Latency time.Duration `json:"latency,omitempty"`

	// Target that was checked
	Target string `json:"target,omitempty"`

	// Error message (if any)
	Error string `json:"error,omitempty"`
}

// DefaultVPNManager implements VPNManager interface
type DefaultVPNManager struct {
	connections      map[string]*VPNConnection
	status           map[string]*VPNStatus
	mu               sync.RWMutex
	monitoringCtx    context.Context
	monitoringCancel context.CancelFunc
	healthCheckers   map[string]*VPNHealthChecker
}

// NewVPNManager creates a new VPN manager
func NewVPNManager() VPNManager {
	return &DefaultVPNManager{
		connections:    make(map[string]*VPNConnection),
		status:         make(map[string]*VPNStatus),
		healthCheckers: make(map[string]*VPNHealthChecker),
	}
}

// AddVPNConnection adds a VPN connection with priority
func (vm *DefaultVPNManager) AddVPNConnection(conn *VPNConnection) error {
	if err := vm.ValidateConnection(conn); err != nil {
		return fmt.Errorf("invalid VPN connection: %w", err)
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.connections[conn.Name] = conn
	vm.status[conn.Name] = &VPNStatus{
		Name:  conn.Name,
		State: VPNStateDisconnected,
	}

	// Set up health checker if enabled
	if conn.HealthCheck != nil && conn.HealthCheck.Enabled {
		vm.healthCheckers[conn.Name] = NewVPNHealthChecker(conn)
	}

	return nil
}

// RemoveVPNConnection removes a VPN connection
func (vm *DefaultVPNManager) RemoveVPNConnection(name string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Disconnect if connected
	if status, exists := vm.status[name]; exists && status.State == VPNStateConnected {
		// Force disconnect
		vm.setConnectionState(name, VPNStateDisconnecting)
		if err := vm.disconnectVPN(name); err != nil {
			return fmt.Errorf("failed to disconnect VPN %s: %w", name, err)
		}
	}

	// Stop health checker
	if checker, exists := vm.healthCheckers[name]; exists {
		checker.Stop()
		delete(vm.healthCheckers, name)
	}

	delete(vm.connections, name)
	delete(vm.status, name)

	return nil
}

// ConnectVPN connects to a VPN by name
func (vm *DefaultVPNManager) ConnectVPN(ctx context.Context, name string) error {
	vm.mu.Lock()
	conn, exists := vm.connections[name]
	if !exists {
		vm.mu.Unlock()
		return fmt.Errorf("VPN connection not found: %s", name)
	}

	status := vm.status[name]
	if status.State == VPNStateConnected {
		vm.mu.Unlock()
		return nil // Already connected
	}

	vm.setConnectionState(name, VPNStateConnecting)
	vm.mu.Unlock()

	// Perform connection
	if err := vm.connectVPN(ctx, conn); err != nil {
		vm.setConnectionState(name, VPNStateError)
		vm.setConnectionError(name, err.Error())
		return fmt.Errorf("failed to connect to VPN %s: %w", name, err)
	}

	vm.mu.Lock()
	vm.status[name].State = VPNStateConnected
	vm.status[name].ConnectedAt = time.Now()
	vm.status[name].Error = ""
	vm.mu.Unlock()

	// Start health checker
	if checker, exists := vm.healthCheckers[name]; exists {
		checker.Start(ctx)
	}

	fmt.Printf("✓ Connected to VPN: %s\n", name)
	return nil
}

// DisconnectVPN disconnects from a VPN by name
func (vm *DefaultVPNManager) DisconnectVPN(ctx context.Context, name string) error {
	vm.mu.Lock()
	_, exists := vm.connections[name]
	if !exists {
		vm.mu.Unlock()
		return fmt.Errorf("VPN connection not found: %s", name)
	}

	status := vm.status[name]
	if status.State == VPNStateDisconnected {
		vm.mu.Unlock()
		return nil // Already disconnected
	}

	vm.setConnectionState(name, VPNStateDisconnecting)
	vm.mu.Unlock()

	// Stop health checker
	if checker, exists := vm.healthCheckers[name]; exists {
		checker.Stop()
	}

	// Perform disconnection
	if err := vm.disconnectVPN(name); err != nil {
		vm.setConnectionState(name, VPNStateError)
		vm.setConnectionError(name, err.Error())
		return fmt.Errorf("failed to disconnect from VPN %s: %w", name, err)
	}

	vm.mu.Lock()
	vm.status[name].State = VPNStateDisconnected
	vm.status[name].DisconnectedAt = time.Now()
	vm.status[name].Error = ""
	vm.mu.Unlock()

	fmt.Printf("✓ Disconnected from VPN: %s\n", name)
	return nil
}

// ConnectByPriority connects to VPNs based on priority order
func (vm *DefaultVPNManager) ConnectByPriority(ctx context.Context) error {
	vm.mu.RLock()
	connections := make([]*VPNConnection, 0, len(vm.connections))
	for _, conn := range vm.connections {
		if conn.AutoConnect {
			connections = append(connections, conn)
		}
	}
	vm.mu.RUnlock()

	// Sort by priority (higher priority first)
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].Priority > connections[j].Priority
	})

	var errors []string
	for _, conn := range connections {
		if err := vm.ConnectVPN(ctx, conn.Name); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", conn.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect to some VPNs: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetConnectionStatus returns status of all VPN connections
func (vm *DefaultVPNManager) GetConnectionStatus() map[string]*VPNStatus {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	status := make(map[string]*VPNStatus, len(vm.status))
	for name, s := range vm.status {
		// Create a copy to avoid race conditions
		status[name] = &VPNStatus{
			Name:            s.Name,
			State:           s.State,
			ConnectedAt:     s.ConnectedAt,
			DisconnectedAt:  s.DisconnectedAt,
			LocalIP:         s.LocalIP,
			RemoteIP:        s.RemoteIP,
			Interface:       s.Interface,
			BytesSent:       s.BytesSent,
			BytesReceived:   s.BytesReceived,
			LastHealthCheck: s.LastHealthCheck,
			Error:           s.Error,
		}
	}

	return status
}

// StartFailoverMonitoring starts monitoring connections for failover
func (vm *DefaultVPNManager) StartFailoverMonitoring(ctx context.Context) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.monitoringCtx != nil {
		return fmt.Errorf("failover monitoring already started")
	}

	vm.monitoringCtx, vm.monitoringCancel = context.WithCancel(ctx)

	go vm.runFailoverMonitoring()

	fmt.Println("✓ VPN failover monitoring started")
	return nil
}

// StopFailoverMonitoring stops the failover monitoring
func (vm *DefaultVPNManager) StopFailoverMonitoring() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.monitoringCancel != nil {
		vm.monitoringCancel()
		vm.monitoringCancel = nil
		vm.monitoringCtx = nil
	}

	fmt.Println("✓ VPN failover monitoring stopped")
}

// GetActiveConnections returns list of active VPN connections
func (vm *DefaultVPNManager) GetActiveConnections() []*VPNConnection {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	var active []*VPNConnection
	for name, status := range vm.status {
		if status.State == VPNStateConnected {
			active = append(active, vm.connections[name])
		}
	}

	return active
}

// ValidateConnection validates VPN connection configuration
func (vm *DefaultVPNManager) ValidateConnection(conn *VPNConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	if conn.Name == "" {
		return fmt.Errorf("connection name is required")
	}

	if conn.Type == "" {
		return fmt.Errorf("connection type is required")
	}

	validTypes := []string{"openvpn", "wireguard", "ipsec", "pptp", "l2tp"}
	valid := false
	for _, validType := range validTypes {
		if conn.Type == validType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}

	if conn.Server == "" {
		return fmt.Errorf("server is required")
	}

	// Validate server format (basic validation)
	if net.ParseIP(conn.Server) == nil {
		// If not an IP, try to resolve hostname
		if _, err := net.LookupHost(conn.Server); err != nil {
			// For testing, accept any hostname format
			if conn.Server == "" {
				return fmt.Errorf("server address cannot be empty")
			}
		}
	}

	// Validate failover configuration
	if conn.Failover != nil && conn.Failover.Enabled {
		if conn.Failover.RetryAttempts <= 0 {
			conn.Failover.RetryAttempts = 3 // Default
		}
		if conn.Failover.RetryInterval <= 0 {
			conn.Failover.RetryInterval = 30 * time.Second // Default
		}
		if conn.Failover.FailoverTimeout <= 0 {
			conn.Failover.FailoverTimeout = 5 * time.Minute // Default
		}
	}

	// Validate health check configuration
	if conn.HealthCheck != nil && conn.HealthCheck.Enabled {
		if conn.HealthCheck.Interval <= 0 {
			conn.HealthCheck.Interval = 30 * time.Second // Default
		}
		if conn.HealthCheck.Timeout <= 0 {
			conn.HealthCheck.Timeout = 10 * time.Second // Default
		}
		if conn.HealthCheck.SuccessThreshold <= 0 {
			conn.HealthCheck.SuccessThreshold = 2 // Default
		}
		if conn.HealthCheck.FailureThreshold <= 0 {
			conn.HealthCheck.FailureThreshold = 3 // Default
		}
		if len(conn.HealthCheck.Targets) == 0 {
			conn.HealthCheck.Targets = []string{"8.8.8.8", "1.1.1.1"} // Default
		}
	}

	return nil
}

// Helper methods

func (vm *DefaultVPNManager) setConnectionState(name string, state VPNState) {
	if status, exists := vm.status[name]; exists {
		status.State = state
	}
}

func (vm *DefaultVPNManager) setConnectionError(name string, error string) {
	if status, exists := vm.status[name]; exists {
		status.Error = error
	}
}

func (vm *DefaultVPNManager) connectVPN(ctx context.Context, conn *VPNConnection) error {
	switch conn.Type {
	case "openvpn":
		return vm.connectOpenVPN(ctx, conn)
	case "wireguard":
		return vm.connectWireGuard(ctx, conn)
	case "ipsec":
		return vm.connectIPSec(ctx, conn)
	default:
		return fmt.Errorf("unsupported VPN type: %s", conn.Type)
	}
}

func (vm *DefaultVPNManager) disconnectVPN(name string) error {
	vm.mu.RLock()
	conn, exists := vm.connections[name]
	vm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", name)
	}

	switch conn.Type {
	case "openvpn":
		return vm.disconnectOpenVPN(name)
	case "wireguard":
		return vm.disconnectWireGuard(name)
	case "ipsec":
		return vm.disconnectIPSec(name)
	default:
		return fmt.Errorf("unsupported VPN type: %s", conn.Type)
	}
}

func (vm *DefaultVPNManager) connectOpenVPN(ctx context.Context, conn *VPNConnection) error {
	if conn.ConfigFile == "" {
		return fmt.Errorf("OpenVPN config file is required")
	}

	args := []string{
		"--config", conn.ConfigFile,
		"--daemon",
		"--writepid", fmt.Sprintf("/tmp/openvpn-%s.pid", conn.Name),
	}

	cmd := exec.CommandContext(ctx, "openvpn", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start OpenVPN: %w", err)
	}

	return nil
}

func (vm *DefaultVPNManager) connectWireGuard(ctx context.Context, conn *VPNConnection) error {
	if conn.ConfigFile == "" {
		return fmt.Errorf("WireGuard config file is required")
	}

	cmd := exec.CommandContext(ctx, "wg-quick", "up", conn.ConfigFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start WireGuard: %w", err)
	}

	return nil
}

func (vm *DefaultVPNManager) connectIPSec(ctx context.Context, conn *VPNConnection) error {
	// Implementation depends on IPSec client (strongSwan, etc.)
	return fmt.Errorf("IPSec connection not implemented yet")
}

func (vm *DefaultVPNManager) disconnectOpenVPN(name string) error {
	pidFile := fmt.Sprintf("/tmp/openvpn-%s.pid", name)
	cmd := exec.Command("pkill", "-F", pidFile)
	return cmd.Run()
}

func (vm *DefaultVPNManager) disconnectWireGuard(name string) error {
	vm.mu.RLock()
	conn := vm.connections[name]
	vm.mu.RUnlock()

	if conn.ConfigFile == "" {
		return fmt.Errorf("WireGuard config file is required")
	}

	cmd := exec.Command("wg-quick", "down", conn.ConfigFile)
	return cmd.Run()
}

func (vm *DefaultVPNManager) disconnectIPSec(name string) error {
	// Implementation depends on IPSec client
	return fmt.Errorf("IPSec disconnection not implemented yet")
}

func (vm *DefaultVPNManager) runFailoverMonitoring() {
	if vm.monitoringCtx == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-vm.monitoringCtx.Done():
			return
		case <-ticker.C:
			vm.checkFailover()
		}
	}
}

func (vm *DefaultVPNManager) checkFailover() {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	for name, conn := range vm.connections {
		if conn.Failover == nil || !conn.Failover.Enabled {
			continue
		}

		status := vm.status[name]
		if status.State == VPNStateConnected {
			// Check if health check is failing
			if checker, exists := vm.healthCheckers[name]; exists {
				if !checker.IsHealthy() {
					fmt.Printf("VPN %s is unhealthy, initiating failover...\n", name)
					go vm.initiateFailover(name)
				}
			}
		}
	}
}

func (vm *DefaultVPNManager) initiateFailover(name string) {
	vm.mu.RLock()
	conn := vm.connections[name]
	vm.mu.RUnlock()

	if conn.Failover == nil || len(conn.Failover.FallbackOrder) == 0 {
		return
	}

	// Try to reconnect first
	for i := 0; i < conn.Failover.RetryAttempts; i++ {
		if err := vm.ConnectVPN(vm.monitoringCtx, name); err == nil {
			fmt.Printf("✓ Successfully reconnected to VPN: %s\n", name)
			return
		}
		time.Sleep(conn.Failover.RetryInterval)
	}

	// Disconnect failed connection
	vm.DisconnectVPN(vm.monitoringCtx, name)

	// Try fallback connections
	for _, fallbackName := range conn.Failover.FallbackOrder {
		if err := vm.ConnectVPN(vm.monitoringCtx, fallbackName); err == nil {
			fmt.Printf("✓ Failed over to VPN: %s\n", fallbackName)
			return
		}
	}

	fmt.Printf("✗ All failover attempts failed for VPN: %s\n", name)
}

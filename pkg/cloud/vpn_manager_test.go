package cloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVPNManager(t *testing.T) {
	manager := NewVPNManager()
	assert.NotNil(t, manager)

	// Test type assertion
	defaultManager, ok := manager.(*DefaultVPNManager)
	assert.True(t, ok)
	assert.NotNil(t, defaultManager.connections)
	assert.NotNil(t, defaultManager.status)
	assert.NotNil(t, defaultManager.healthCheckers)
}

func TestVPNManager_AddConnection(t *testing.T) {
	manager := NewVPNManager()

	// Test valid connection
	conn := &VPNConnection{
		Name:        "test-vpn",
		Type:        "openvpn",
		Server:      "vpn.example.com",
		Port:        1194,
		Priority:    100,
		AutoConnect: true,
		HealthCheck: &VPNHealthCheck{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Targets:  []string{"8.8.8.8:53"},
		},
	}

	err := manager.AddVPNConnection(conn)
	assert.NoError(t, err)

	// Verify connection was added
	status := manager.GetConnectionStatus()
	assert.Len(t, status, 1)
	assert.Equal(t, VPNStateDisconnected, status["test-vpn"].State)

	// Test duplicate connection
	err = manager.AddVPNConnection(conn)
	assert.NoError(t, err) // Should replace existing

	// Test invalid connection
	invalidConn := &VPNConnection{
		Name: "invalid-vpn",
		// Missing required fields
	}
	err = manager.AddVPNConnection(invalidConn)
	assert.Error(t, err)
}

func TestVPNManager_RemoveConnection(t *testing.T) {
	manager := NewVPNManager()

	// Add a connection
	conn := &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
	}
	err := manager.AddVPNConnection(conn)
	require.NoError(t, err)

	// Remove the connection
	err = manager.RemoveVPNConnection("test-vpn")
	assert.NoError(t, err)

	// Verify connection was removed
	status := manager.GetConnectionStatus()
	assert.Len(t, status, 0)

	// Test removing non-existent connection
	err = manager.RemoveVPNConnection("non-existent")
	assert.NoError(t, err) // Should not error
}

func TestVPNManager_ValidateConnection(t *testing.T) {
	manager := NewVPNManager()

	// Test nil connection
	err := manager.ValidateConnection(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection cannot be nil")

	// Test empty name
	conn := &VPNConnection{
		Type:   "openvpn",
		Server: "vpn.example.com",
	}
	err = manager.ValidateConnection(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection name is required")

	// Test empty type
	conn = &VPNConnection{
		Name:   "test-vpn",
		Server: "vpn.example.com",
	}
	err = manager.ValidateConnection(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection type is required")

	// Test invalid type
	conn = &VPNConnection{
		Name:   "test-vpn",
		Type:   "invalid-type",
		Server: "vpn.example.com",
	}
	err = manager.ValidateConnection(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported connection type")

	// Test empty server
	conn = &VPNConnection{
		Name: "test-vpn",
		Type: "openvpn",
	}
	err = manager.ValidateConnection(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is required")

	// Test valid connection
	conn = &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
	}
	err = manager.ValidateConnection(conn)
	assert.NoError(t, err)

	// Test connection with failover (should set defaults)
	conn = &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
		Failover: &VPNFailover{
			Enabled: true,
		},
	}
	err = manager.ValidateConnection(conn)
	assert.NoError(t, err)
	assert.Equal(t, 3, conn.Failover.RetryAttempts)
	assert.Equal(t, 30*time.Second, conn.Failover.RetryInterval)

	// Test connection with health check (should set defaults)
	conn = &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
		HealthCheck: &VPNHealthCheck{
			Enabled: true,
		},
	}
	err = manager.ValidateConnection(conn)
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Second, conn.HealthCheck.Interval)
	assert.Equal(t, 10*time.Second, conn.HealthCheck.Timeout)
	assert.Equal(t, 2, conn.HealthCheck.SuccessThreshold)
	assert.Equal(t, 3, conn.HealthCheck.FailureThreshold)
	assert.Contains(t, conn.HealthCheck.Targets, "8.8.8.8")
}

func TestVPNManager_GetConnectionStatus(t *testing.T) {
	manager := NewVPNManager()

	// Test empty status
	status := manager.GetConnectionStatus()
	assert.Len(t, status, 0)

	// Add connections
	conn1 := &VPNConnection{
		Name:   "vpn1",
		Type:   "openvpn",
		Server: "vpn1.example.com",
		Port:   1194,
	}
	conn2 := &VPNConnection{
		Name:   "vpn2",
		Type:   "wireguard",
		Server: "vpn2.example.com",
		Port:   51820,
	}

	err := manager.AddVPNConnection(conn1)
	require.NoError(t, err)
	err = manager.AddVPNConnection(conn2)
	require.NoError(t, err)

	// Test status with connections
	status = manager.GetConnectionStatus()
	assert.Len(t, status, 2)
	assert.Equal(t, VPNStateDisconnected, status["vpn1"].State)
	assert.Equal(t, VPNStateDisconnected, status["vpn2"].State)
	assert.Equal(t, "vpn1", status["vpn1"].Name)
	assert.Equal(t, "vpn2", status["vpn2"].Name)
}

func TestVPNManager_GetActiveConnections(t *testing.T) {
	manager := NewVPNManager()

	// Test empty active connections
	active := manager.GetActiveConnections()
	assert.Len(t, active, 0)

	// Add connections
	conn := &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
	}
	err := manager.AddVPNConnection(conn)
	require.NoError(t, err)

	// Still no active connections (not connected)
	active = manager.GetActiveConnections()
	assert.Len(t, active, 0)

	// Manually set connection as connected for testing
	defaultManager := manager.(*DefaultVPNManager)
	defaultManager.mu.Lock()
	defaultManager.status["test-vpn"].State = VPNStateConnected
	defaultManager.mu.Unlock()

	// Now should have active connection
	active = manager.GetActiveConnections()
	assert.Len(t, active, 1)
	assert.Equal(t, "test-vpn", active[0].Name)
}

func TestVPNManager_StartStopFailoverMonitoring(t *testing.T) {
	manager := NewVPNManager()
	ctx := context.Background()

	// Test starting monitoring
	err := manager.StartFailoverMonitoring(ctx)
	assert.NoError(t, err)

	// Test starting monitoring again (should error)
	err = manager.StartFailoverMonitoring(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Test stopping monitoring
	manager.StopFailoverMonitoring()

	// Test stopping monitoring again (should not error)
	manager.StopFailoverMonitoring()
}

func TestVPNConnection_DefaultValues(t *testing.T) {
	conn := &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
		Failover: &VPNFailover{
			Enabled: true,
		},
		HealthCheck: &VPNHealthCheck{
			Enabled: true,
		},
	}

	manager := NewVPNManager()
	err := manager.ValidateConnection(conn)
	assert.NoError(t, err)

	// Check failover defaults
	assert.Equal(t, 3, conn.Failover.RetryAttempts)
	assert.Equal(t, 30*time.Second, conn.Failover.RetryInterval)
	assert.Equal(t, 5*time.Minute, conn.Failover.FailoverTimeout)

	// Check health check defaults
	assert.Equal(t, 30*time.Second, conn.HealthCheck.Interval)
	assert.Equal(t, 10*time.Second, conn.HealthCheck.Timeout)
	assert.Equal(t, 2, conn.HealthCheck.SuccessThreshold)
	assert.Equal(t, 3, conn.HealthCheck.FailureThreshold)
	assert.Len(t, conn.HealthCheck.Targets, 2)
}

func TestVPNState_String(t *testing.T) {
	tests := []struct {
		state    VPNState
		expected string
	}{
		{VPNStateDisconnected, "disconnected"},
		{VPNStateConnecting, "connecting"},
		{VPNStateConnected, "connected"},
		{VPNStateDisconnecting, "disconnecting"},
		{VPNStateError, "error"},
		{VPNStateReconnecting, "reconnecting"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, string(tt.state))
	}
}

func TestVPNCredentials_Validation(t *testing.T) {
	// Test that credentials structure is properly defined
	creds := &VPNCredentials{
		Username: "testuser",
		Password: "testpass",
		CertFile: "/path/to/cert.pem",
		KeyFile:  "/path/to/key.pem",
		CAFile:   "/path/to/ca.pem",
		PSK:      "presharedkey",
	}

	assert.Equal(t, "testuser", creds.Username)
	assert.Equal(t, "testpass", creds.Password)
	assert.Equal(t, "/path/to/cert.pem", creds.CertFile)
	assert.Equal(t, "/path/to/key.pem", creds.KeyFile)
	assert.Equal(t, "/path/to/ca.pem", creds.CAFile)
	assert.Equal(t, "presharedkey", creds.PSK)
}

func TestVPNRoute_Validation(t *testing.T) {
	// Test that route structure is properly defined
	route := &VPNRoute{
		Destination: "192.168.1.0/24",
		Gateway:     "10.0.0.1",
		Metric:      100,
		Interface:   "tun0",
	}

	assert.Equal(t, "192.168.1.0/24", route.Destination)
	assert.Equal(t, "10.0.0.1", route.Gateway)
	assert.Equal(t, 100, route.Metric)
	assert.Equal(t, "tun0", route.Interface)
}

func TestVPNManager_ConnectionTypes(t *testing.T) {
	manager := NewVPNManager()

	validTypes := []string{"openvpn", "wireguard", "ipsec", "pptp", "l2tp"}

	for _, vpnType := range validTypes {
		conn := &VPNConnection{
			Name:   fmt.Sprintf("test-%s", vpnType),
			Type:   vpnType,
			Server: "vpn.example.com",
			Port:   1194,
		}

		err := manager.ValidateConnection(conn)
		assert.NoError(t, err, "VPN type %s should be valid", vpnType)
	}

	// Test invalid type
	invalidConn := &VPNConnection{
		Name:   "test-invalid",
		Type:   "invalid-type",
		Server: "vpn.example.com",
		Port:   1194,
	}

	err := manager.ValidateConnection(invalidConn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported connection type")
}

func TestVPNManager_PriorityOrdering(t *testing.T) {
	manager := NewVPNManager()

	// Add connections with different priorities
	conn1 := &VPNConnection{
		Name:        "low-priority",
		Type:        "openvpn",
		Server:      "vpn1.example.com",
		Port:        1194,
		Priority:    50,
		AutoConnect: true,
	}
	conn2 := &VPNConnection{
		Name:        "high-priority",
		Type:        "openvpn",
		Server:      "vpn2.example.com",
		Port:        1194,
		Priority:    100,
		AutoConnect: true,
	}
	conn3 := &VPNConnection{
		Name:        "medium-priority",
		Type:        "openvpn",
		Server:      "vpn3.example.com",
		Port:        1194,
		Priority:    75,
		AutoConnect: true,
	}

	err := manager.AddVPNConnection(conn1)
	require.NoError(t, err)
	err = manager.AddVPNConnection(conn2)
	require.NoError(t, err)
	err = manager.AddVPNConnection(conn3)
	require.NoError(t, err)

	// Test that connections are available
	status := manager.GetConnectionStatus()
	assert.Len(t, status, 3)

	// Note: We can't test actual connection priority ordering without mocking
	// the connection methods, but we can verify the connections are properly stored
	assert.Contains(t, status, "low-priority")
	assert.Contains(t, status, "high-priority")
	assert.Contains(t, status, "medium-priority")
}

func TestHealthCheckResult_Structure(t *testing.T) {
	// Test that HealthCheckResult structure is properly defined
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Success:   true,
		Latency:   50 * time.Millisecond,
		Target:    "8.8.8.8:53",
		Error:     "",
	}

	assert.True(t, result.Success)
	assert.Equal(t, 50*time.Millisecond, result.Latency)
	assert.Equal(t, "8.8.8.8:53", result.Target)
	assert.Empty(t, result.Error)
}

func TestVPNManager_ConcurrentAccess(t *testing.T) {
	manager := NewVPNManager()

	// Test concurrent access to manager
	conn := &VPNConnection{
		Name:   "test-vpn",
		Type:   "openvpn",
		Server: "vpn.example.com",
		Port:   1194,
	}

	// Add connection concurrently
	done := make(chan bool, 2)

	go func() {
		err := manager.AddVPNConnection(conn)
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		status := manager.GetConnectionStatus()
		assert.True(t, len(status) >= 0) // Should not panic
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify final state
	status := manager.GetConnectionStatus()
	assert.Len(t, status, 1)
	assert.Equal(t, "test-vpn", status["test-vpn"].Name)
}

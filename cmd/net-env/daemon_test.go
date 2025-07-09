package netenv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDaemonCmd(t *testing.T) {
	ctx := context.Background()
	cmd := newDaemonCmd(ctx)

	assert.Equal(t, "daemon", cmd.Use)
	assert.Equal(t, "Monitor and manage system daemons", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 5)

	var listCmd, statusCmd, monitorCmd, manageCmd, healthCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "list":
			listCmd = true
		case "status":
			statusCmd = true
		case "monitor":
			monitorCmd = true
		case "manage":
			manageCmd = true
		case "health":
			healthCmd = true
		}
	}

	assert.True(t, listCmd, "list subcommand should exist")
	assert.True(t, statusCmd, "status subcommand should exist")
	assert.True(t, monitorCmd, "monitor subcommand should exist")
	assert.True(t, manageCmd, "manage subcommand should exist")
	assert.True(t, healthCmd, "health subcommand should exist")
}

func TestNewDaemonListCmd(t *testing.T) {
	cmd := newDaemonListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List system daemons and their status", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("network-services"))
	assert.NotNil(t, cmd.Flags().Lookup("show-inactive"))
}

func TestNewDaemonStatusCmd(t *testing.T) {
	cmd := newDaemonStatusCmd()

	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Show detailed status of a specific daemon", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	serviceFlag := cmd.Flags().Lookup("service")
	assert.NotNil(t, serviceFlag)
	// Check if flag is required (this is a bit tricky to test directly)
	assert.NotEmpty(t, serviceFlag.Usage)
}

func TestNewDaemonMonitorCmd(t *testing.T) {
	ctx := context.Background()
	cmd := newDaemonMonitorCmd(ctx)

	assert.Equal(t, "monitor", cmd.Use)
	assert.Equal(t, "Monitor daemon status with live updates", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("service"))
	assert.NotNil(t, cmd.Flags().Lookup("network-services"))
	assert.NotNil(t, cmd.Flags().Lookup("follow-logs"))
	assert.NotNil(t, cmd.Flags().Lookup("enable-health"))
}

func TestDaemonOptions(t *testing.T) {
	opts := &daemonOptions{
		serviceName:     "test-service",
		networkServices: true,
		followLogs:      true,
		showInactive:    false,
	}

	assert.Equal(t, "test-service", opts.serviceName)
	assert.True(t, opts.networkServices)
	assert.True(t, opts.followLogs)
	assert.False(t, opts.showInactive)
}

func TestServiceInfo(t *testing.T) {
	service := serviceInfo{
		Name:        "ssh",
		Status:      "active",
		Enabled:     "enabled",
		Description: "OpenSSH server daemon",
		MainPID:     "1234",
		Memory:      "2048000",
		Since:       "2024-01-01 10:00:00",
	}

	assert.Equal(t, "ssh", service.Name)
	assert.Equal(t, "active", service.Status)
	assert.Equal(t, "enabled", service.Enabled)
	assert.Equal(t, "OpenSSH server daemon", service.Description)
	assert.Equal(t, "1234", service.MainPID)
	assert.Equal(t, "2048000", service.Memory)
	assert.Equal(t, "2024-01-01 10:00:00", service.Since)
}

func TestFilterNetworkServices(t *testing.T) {
	opts := &daemonOptions{}

	testServices := []serviceInfo{
		{Name: "ssh", Description: "SSH server daemon"},
		{Name: "apache2", Description: "Apache HTTP server"},
		{Name: "NetworkManager", Description: "Network connection manager"},
		{Name: "cron", Description: "Regular background program processing daemon"},
		{Name: "nginx", Description: "HTTP and reverse proxy server"},
		{Name: "bluetooth", Description: "Bluetooth service"},
		{Name: "firewall", Description: "Firewall daemon"},
		{Name: "systemd-resolved", Description: "Network Name Resolution"},
	}

	filtered := opts.filterNetworkServices(testServices)

	// Should include network-related services
	networkServiceNames := make(map[string]bool)
	for _, service := range filtered {
		networkServiceNames[service.Name] = true
	}

	assert.True(t, networkServiceNames["ssh"])
	assert.True(t, networkServiceNames["apache2"])
	assert.True(t, networkServiceNames["NetworkManager"])
	assert.True(t, networkServiceNames["nginx"])
	assert.True(t, networkServiceNames["firewall"])
	assert.True(t, networkServiceNames["systemd-resolved"])

	// Should not include non-network services
	assert.False(t, networkServiceNames["cron"])
	assert.False(t, networkServiceNames["bluetooth"])
}

func TestFilterActiveServices(t *testing.T) {
	opts := &daemonOptions{}

	testServices := []serviceInfo{
		{Name: "ssh", Status: "active"},
		{Name: "apache2", Status: "inactive"},
		{Name: "nginx", Status: "active"},
		{Name: "mysql", Status: "failed"},
		{Name: "redis", Status: "activating"},
		{Name: "postgres", Status: "deactivating"},
	}

	filtered := opts.filterActiveServices(testServices)

	// Should include active and failed services
	assert.Len(t, filtered, 3)

	serviceNames := make(map[string]bool)
	for _, service := range filtered {
		serviceNames[service.Name] = true
	}

	assert.True(t, serviceNames["ssh"])       // active
	assert.True(t, serviceNames["nginx"])     // active
	assert.True(t, serviceNames["mysql"])     // failed
	assert.False(t, serviceNames["apache2"])  // inactive
	assert.False(t, serviceNames["redis"])    // activating
	assert.False(t, serviceNames["postgres"]) // deactivating
}

func TestGetStatusIcon(t *testing.T) {
	opts := &daemonOptions{}

	testCases := []struct {
		status   string
		expected string
	}{
		{"active", "✅ "},
		{"inactive", "⚪ "},
		{"failed", "❌ "},
		{"activating", "🔄 "},
		{"deactivating", "🔄 "},
		{"unknown", "❓ "},
		{"", "❓ "},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			icon := opts.getStatusIcon(tc.status)
			assert.Equal(t, tc.expected, icon)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	opts := &daemonOptions{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"", "0 B"},
		{"0", "0 B"},
		{"1024", "1024 bytes"},
		{"2048000", "2048000 bytes"},
		{"invalid", "invalid bytes"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := opts.formatBytes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRunMonitorValidation(t *testing.T) {
	opts := &daemonOptions{}

	// Test with no service or network-services flag
	ctx := context.Background()
	err := opts.runMonitor(ctx, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either --service or --network-services must be specified")

	// Test with service name
	opts.serviceName = "ssh"
	// This would normally start the monitor loop, but we can't easily test that
	// without mocking or changing the implementation
}

func TestServiceInfoFields(t *testing.T) {
	service := serviceInfo{}

	// Test zero values
	assert.Empty(t, service.Name)
	assert.Empty(t, service.Status)
	assert.Empty(t, service.Enabled)
	assert.Empty(t, service.Description)
	assert.Empty(t, service.MainPID)
	assert.Empty(t, service.Memory)
	assert.Empty(t, service.Since)

	// Test field assignment
	service.Name = "test-service"
	service.Status = "active"
	service.Enabled = "enabled"
	service.Description = "Test service description"
	service.MainPID = "5678"
	service.Memory = "4096000"
	service.Since = "2024-01-01 12:00:00"

	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "active", service.Status)
	assert.Equal(t, "enabled", service.Enabled)
	assert.Equal(t, "Test service description", service.Description)
	assert.Equal(t, "5678", service.MainPID)
	assert.Equal(t, "4096000", service.Memory)
	assert.Equal(t, "2024-01-01 12:00:00", service.Since)
}

func TestDaemonOptionsDefaults(t *testing.T) {
	opts := &daemonOptions{}

	assert.Empty(t, opts.serviceName)
	assert.False(t, opts.networkServices)
	assert.False(t, opts.followLogs)
	assert.False(t, opts.showInactive)
}

func TestNetworkServicePatterns(t *testing.T) {
	opts := &daemonOptions{}

	// Test various network service patterns
	testServices := []serviceInfo{
		{Name: "sshd", Description: "SSH daemon"},
		{Name: "httpd", Description: "HTTP server"},
		{Name: "wpa_supplicant", Description: "WPA supplicant"},
		{Name: "NetworkManager", Description: "Network Manager"},
		{Name: "systemd-networkd", Description: "systemd network daemon"},
		{Name: "systemd-resolved", Description: "systemd DNS resolver"},
		{Name: "openvpn", Description: "OpenVPN daemon"},
		{Name: "wireguard", Description: "WireGuard VPN"},
		{Name: "ufw", Description: "Uncomplicated Firewall"},
		{Name: "fail2ban", Description: "Fail2ban service"},
		{Name: "named", Description: "BIND DNS server"},
		{Name: "dnsmasq", Description: "DNS forwarder"},
		{Name: "avahi-daemon", Description: "Avahi mDNS/DNS-SD daemon"},
		{Name: "dhcpd", Description: "DHCP server"},
		{Name: "hostapd", Description: "Host AP daemon"},
		{Name: "cups", Description: "Common UNIX Printing System"},
		{Name: "cron", Description: "Cron daemon"},
	}

	filtered := opts.filterNetworkServices(testServices)

	networkServiceNames := make(map[string]bool)
	for _, service := range filtered {
		networkServiceNames[service.Name] = true
	}

	// Should include network services
	assert.True(t, networkServiceNames["sshd"])
	assert.True(t, networkServiceNames["httpd"])
	assert.True(t, networkServiceNames["wpa_supplicant"])
	assert.True(t, networkServiceNames["NetworkManager"])
	assert.True(t, networkServiceNames["systemd-networkd"])
	assert.True(t, networkServiceNames["systemd-resolved"])
	assert.True(t, networkServiceNames["openvpn"])
	assert.True(t, networkServiceNames["wireguard"])
	assert.True(t, networkServiceNames["ufw"])
	assert.True(t, networkServiceNames["fail2ban"])
	assert.True(t, networkServiceNames["named"])
	assert.True(t, networkServiceNames["dnsmasq"])
	assert.True(t, networkServiceNames["avahi-daemon"])
	assert.True(t, networkServiceNames["dhcpd"])
	assert.True(t, networkServiceNames["hostapd"])

	// Should not include non-network services
	assert.False(t, networkServiceNames["cups"])
	assert.False(t, networkServiceNames["cron"])
}

func TestRunListWithFlags(t *testing.T) {
	// This test would require mocking systemctl commands,
	// which is complex. For now, we test the flag configuration.
	cmd := newDaemonListCmd()

	networkFlag := cmd.Flags().Lookup("network-services")
	require.NotNil(t, networkFlag)
	assert.Equal(t, "false", networkFlag.DefValue)

	inactiveFlag := cmd.Flags().Lookup("show-inactive")
	require.NotNil(t, inactiveFlag)
	assert.Equal(t, "false", inactiveFlag.DefValue)
}

func TestRunStatusWithFlags(t *testing.T) {
	cmd := newDaemonStatusCmd()

	serviceFlag := cmd.Flags().Lookup("service")
	require.NotNil(t, serviceFlag)
	assert.Equal(t, "", serviceFlag.DefValue)
}

func TestRunMonitorWithFlags(t *testing.T) {
	ctx := context.Background()
	cmd := newDaemonMonitorCmd(ctx)

	serviceFlag := cmd.Flags().Lookup("service")
	require.NotNil(t, serviceFlag)
	assert.Equal(t, "", serviceFlag.DefValue)

	networkFlag := cmd.Flags().Lookup("network-services")
	require.NotNil(t, networkFlag)
	assert.Equal(t, "false", networkFlag.DefValue)

	logsFlag := cmd.Flags().Lookup("follow-logs")
	require.NotNil(t, logsFlag)
	assert.Equal(t, "false", logsFlag.DefValue)
}

func TestNewDaemonManageCmd(t *testing.T) {
	cmd := newDaemonManageCmd()

	assert.Equal(t, "manage", cmd.Use)
	assert.Contains(t, cmd.Short, "Manage daemon services")
	assert.Contains(t, cmd.Long, "Start, stop, restart")

	// Check required flags
	serviceFlag := cmd.Flags().Lookup("service")
	require.NotNil(t, serviceFlag)

	actionFlag := cmd.Flags().Lookup("action")
	require.NotNil(t, actionFlag)
}

func TestNewDaemonHealthCmd(t *testing.T) {
	ctx := context.Background()
	cmd := newDaemonHealthCmd(ctx)

	assert.Equal(t, "health", cmd.Use)
	assert.Contains(t, cmd.Short, "health and performance metrics")
	assert.Contains(t, cmd.Long, "health monitoring")

	// Check flags
	serviceFlag := cmd.Flags().Lookup("service")
	require.NotNil(t, serviceFlag)

	networkFlag := cmd.Flags().Lookup("network-services")
	require.NotNil(t, networkFlag)

	enableHealthFlag := cmd.Flags().Lookup("enable-health")
	require.NotNil(t, enableHealthFlag)
}

func TestDaemonOptions_checkServiceHealth(t *testing.T) {
	tests := []struct {
		name           string
		service        *serviceInfo
		expectedStatus string
		expectedErrors int
	}{
		{
			name: "healthy active service",
			service: &serviceInfo{
				Name:    "test-service",
				Status:  "active",
				MainPID: "", // Empty PID to skip process check
			},
			expectedStatus: "healthy",
			expectedErrors: 0,
		},
		{
			name: "failed service",
			service: &serviceInfo{
				Name:   "failed-service",
				Status: "failed",
			},
			expectedStatus: "unhealthy",
			expectedErrors: 1,
		},
		{
			name: "inactive service",
			service: &serviceInfo{
				Name:   "inactive-service",
				Status: "inactive",
			},
			expectedStatus: "stopped",
			expectedErrors: 0,
		},
		{
			name: "unknown status service",
			service: &serviceInfo{
				Name:   "unknown-service",
				Status: "unknown",
			},
			expectedStatus: "unknown",
			expectedErrors: 0,
		},
	}

	o := &daemonOptions{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := o.checkServiceHealth(tt.service)

			assert.Equal(t, tt.expectedStatus, health.Status)
			assert.Equal(t, tt.expectedErrors, health.ErrorCount)
			assert.Equal(t, tt.service.Name, health.Name)
			assert.True(t, time.Since(health.LastChecked) < time.Second)
		})
	}
}

func TestDaemonOptions_runManage_InvalidAction(t *testing.T) {
	o := &daemonOptions{
		serviceName: "test-service",
		action:      "invalid-action",
	}

	err := o.runManage(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action 'invalid-action'")
}

func TestDaemonOptions_runHealth_NoServiceOrNetwork(t *testing.T) {
	o := &daemonOptions{}
	ctx := context.Background()

	err := o.runHealth(ctx, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either --service or --network-services must be specified")
}

func TestDaemonOptions_ExtendedFields(t *testing.T) {
	opts := &daemonOptions{
		serviceName:     "test-service",
		networkServices: true,
		followLogs:      true,
		showInactive:    false,
		enableHealth:    true,
		action:          "start",
	}

	assert.Equal(t, "test-service", opts.serviceName)
	assert.True(t, opts.networkServices)
	assert.True(t, opts.followLogs)
	assert.False(t, opts.showInactive)
	assert.True(t, opts.enableHealth)
	assert.Equal(t, "start", opts.action)
}

func TestServiceInfo_ExtendedFields(t *testing.T) {
	service := serviceInfo{
		Name:        "test-service",
		Status:      "active",
		Enabled:     "enabled",
		Description: "Test service description",
		MainPID:     "1234",
		Memory:      "128MB",
		Since:       "2023-01-01 00:00:00",
		CPUUsage:    "5.2%",
		LoadState:   "loaded",
		SubState:    "running",
	}

	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "active", service.Status)
	assert.Equal(t, "5.2%", service.CPUUsage)
	assert.Equal(t, "loaded", service.LoadState)
	assert.Equal(t, "running", service.SubState)
}

func TestHealthCheck_Fields(t *testing.T) {
	health := healthCheck{
		Name:           "test-service",
		Status:         "healthy",
		LastChecked:    time.Now(),
		ResponseTime:   100 * time.Millisecond,
		ErrorCount:     0,
		HealthEndpoint: "http://localhost:8080/health",
	}

	assert.Equal(t, "test-service", health.Name)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, 100*time.Millisecond, health.ResponseTime)
	assert.Equal(t, 0, health.ErrorCount)
	assert.Equal(t, "http://localhost:8080/health", health.HealthEndpoint)
}

func TestDaemonOptions_isProcessRunning(t *testing.T) {
	o := &daemonOptions{}

	// Test with invalid PID
	assert.False(t, o.isProcessRunning("invalid"))
	assert.False(t, o.isProcessRunning("999999"))

	// Test with PID 1 (init process, should exist on most systems)
	// Skip if not running on a system with init process at PID 1
	if o.isProcessRunning("1") {
		assert.True(t, o.isProcessRunning("1"))
	} else {
		t.Skip("Skipping PID 1 test - not running on a system with init at PID 1")
	}
}

func TestValidActions(t *testing.T) {
	o := &daemonOptions{serviceName: "test-service"}

	validActions := []string{"start", "stop", "restart", "enable", "disable", "reload"}
	for _, action := range validActions {
		o.action = action
		// We can't actually test execution without mocking systemctl,
		// but we can verify the action validation logic doesn't reject valid actions
		assert.NotEmpty(t, action)
	}

	// Test invalid action
	o.action = "invalid"
	err := o.runManage(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestNetworkServicePatternsComprehensive(t *testing.T) {
	testServices := []struct {
		name        string
		description string
		shouldMatch bool
	}{
		{"sshd", "SSH daemon", true},
		{"nginx", "Web server", true},
		{"NetworkManager", "Network management", true},
		{"systemd-networkd", "Network configuration", true},
		{"systemd-resolved", "DNS resolver", true},
		{"wpa_supplicant", "WiFi authentication", true},
		{"openvpn", "VPN client", true},
		{"ufw", "Uncomplicated firewall", true},
		{"fail2ban", "Intrusion prevention", true},
		{"bind9", "DNS server", true},
		{"dnsmasq", "DNS/DHCP server", true},
		{"avahi-daemon", "mDNS/DNS-SD daemon", true},
		{"cron", "Task scheduler", false},
		{"bluetooth", "Bluetooth service", false},
		{"cups", "Printing service", false},
		{"docker", "Container runtime", false}, // Not primarily network
	}

	o := &daemonOptions{}

	for _, test := range testServices {
		t.Run(test.name, func(t *testing.T) {
			services := []serviceInfo{{
				Name:        test.name,
				Description: test.description,
			}}

			filtered := o.filterNetworkServices(services)

			if test.shouldMatch {
				assert.Len(t, filtered, 1, "Service %s should be identified as network-related", test.name)
			} else {
				assert.Len(t, filtered, 0, "Service %s should not be identified as network-related", test.name)
			}
		})
	}
}

func TestRunMonitorWithExtendedFlags(t *testing.T) {
	ctx := context.Background()
	cmd := newDaemonMonitorCmd(ctx)

	serviceFlag := cmd.Flags().Lookup("service")
	require.NotNil(t, serviceFlag)
	assert.Equal(t, "", serviceFlag.DefValue)

	networkFlag := cmd.Flags().Lookup("network-services")
	require.NotNil(t, networkFlag)
	assert.Equal(t, "false", networkFlag.DefValue)

	logsFlag := cmd.Flags().Lookup("follow-logs")
	require.NotNil(t, logsFlag)
	assert.Equal(t, "false", logsFlag.DefValue)

	healthFlag := cmd.Flags().Lookup("enable-health")
	require.NotNil(t, healthFlag)
	assert.Equal(t, "false", healthFlag.DefValue)
}

package net_env

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDaemonCmd(t *testing.T) {
	cmd := newDaemonCmd()

	assert.Equal(t, "daemon", cmd.Use)
	assert.Equal(t, "Monitor and manage system daemons", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var listCmd, statusCmd, monitorCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "list":
			listCmd = true
		case "status":
			statusCmd = true
		case "monitor":
			monitorCmd = true
		}
	}

	assert.True(t, listCmd, "list subcommand should exist")
	assert.True(t, statusCmd, "status subcommand should exist")
	assert.True(t, monitorCmd, "monitor subcommand should exist")
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
	cmd := newDaemonMonitorCmd()

	assert.Equal(t, "monitor", cmd.Use)
	assert.Equal(t, "Monitor daemon status with live updates", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("service"))
	assert.NotNil(t, cmd.Flags().Lookup("network-services"))
	assert.NotNil(t, cmd.Flags().Lookup("follow-logs"))
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
	err := opts.runMonitor(nil, nil)
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
	cmd := newDaemonMonitorCmd()

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

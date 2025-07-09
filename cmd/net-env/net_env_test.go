package netenv

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewNetEnvCmd(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx)

	assert.Equal(t, "net-env", cmd.Use)
	assert.Equal(t, "Manage network environment transitions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	// Verify subcommands exist
	var daemonCmd, wifiCmd, actionsCmd *cobra.Command
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "daemon":
			daemonCmd = subcmd
		case "wifi":
			wifiCmd = subcmd
		case "actions":
			actionsCmd = subcmd
		}
	}

	assert.NotNil(t, daemonCmd)
	assert.Equal(t, "daemon", daemonCmd.Use)
	assert.Equal(t, "Monitor and manage system daemons", daemonCmd.Short)

	assert.NotNil(t, wifiCmd)
	assert.Equal(t, "wifi", wifiCmd.Use)
	assert.Equal(t, "Monitor WiFi changes and trigger actions", wifiCmd.Short)

	assert.NotNil(t, actionsCmd)
	assert.Equal(t, "actions", actionsCmd.Use)
	assert.Equal(t, "Execute network configuration actions", actionsCmd.Short)
}

func TestNetEnvCmdStructure(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx)

	// Test that the command has proper structure
	assert.NotNil(t, cmd.Use)
	assert.NotNil(t, cmd.Short)
	assert.NotNil(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Test that examples are included in Long description
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz net-env daemon list")
	assert.Contains(t, cmd.Long, "gz net-env daemon status --service ssh")
	assert.Contains(t, cmd.Long, "gz net-env daemon monitor --network-services")
}

func TestNetEnvCmdHelpContent(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx)

	// Verify help content mentions key features
	longDesc := cmd.Long
	assert.Contains(t, longDesc, "network environment transitions")
	assert.Contains(t, longDesc, "Daemon/service status monitoring")
	assert.Contains(t, longDesc, "Service dependency tracking")
	assert.Contains(t, longDesc, "Network environment transition management")
	assert.Contains(t, longDesc, "WiFi change event monitoring and action triggers")
	assert.Contains(t, longDesc, "Network configuration actions (VPN, DNS, proxy, hosts)")
	assert.Contains(t, longDesc, "System state verification")

	// Verify use cases are mentioned
	assert.Contains(t, longDesc, "Moving between different network environments")
	assert.Contains(t, longDesc, "Switching VPN connections")
	assert.Contains(t, longDesc, "Managing services that depend on network connectivity")
	assert.Contains(t, longDesc, "Verifying system state after network changes")

	// Verify WiFi examples are included
	assert.Contains(t, longDesc, "gz net-env wifi monitor")
	assert.Contains(t, longDesc, "gz net-env wifi status")

	// Verify actions examples are included
	assert.Contains(t, longDesc, "gz net-env actions run")
	assert.Contains(t, longDesc, "gz net-env actions vpn connect")
	assert.Contains(t, longDesc, "gz net-env actions dns set")
}

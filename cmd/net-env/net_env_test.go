package net_env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNetEnvCmd(t *testing.T) {
	cmd := NewNetEnvCmd()

	assert.Equal(t, "net-env", cmd.Use)
	assert.Equal(t, "Manage network environment transitions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1)

	// Verify daemon subcommand exists
	daemonCmd := subcommands[0]
	assert.Equal(t, "daemon", daemonCmd.Use)
	assert.Equal(t, "Monitor and manage system daemons", daemonCmd.Short)
}

func TestNetEnvCmdStructure(t *testing.T) {
	cmd := NewNetEnvCmd()

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
	cmd := NewNetEnvCmd()

	// Verify help content mentions key features
	longDesc := cmd.Long
	assert.Contains(t, longDesc, "network environment transitions")
	assert.Contains(t, longDesc, "Daemon/service status monitoring")
	assert.Contains(t, longDesc, "Service dependency tracking")
	assert.Contains(t, longDesc, "Network environment transition management")
	assert.Contains(t, longDesc, "System state verification")

	// Verify use cases are mentioned
	assert.Contains(t, longDesc, "Moving between different network environments")
	assert.Contains(t, longDesc, "Switching VPN connections")
	assert.Contains(t, longDesc, "Managing services that depend on network connectivity")
	assert.Contains(t, longDesc, "Verifying system state after network changes")
}

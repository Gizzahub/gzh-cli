//nolint:testpackage // White-box testing needed for internal function access
package netenv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

func TestNewNetEnvCmd(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx, app.NewTestAppContext())

	assert.Equal(t, "net-env", cmd.Use)
	assert.Equal(t, "Manage network environment configurations", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Library-backed tree: status, watch, profile (no legacy switch/daemon/wifi)
	subcommands := cmd.Commands()
	require.GreaterOrEqual(t, len(subcommands), 3)

	names := make(map[string]string, len(subcommands))
	for _, subcmd := range subcommands {
		names[subcmd.Name()] = subcmd.Short
	}

	assert.Equal(t, "Show network environment status", names["status"])
	assert.Equal(t, "Continuously monitor network changes", names["watch"])
	assert.Equal(t, "Manage network environment profiles", names["profile"])
	assert.NotContains(t, names, "switch")
}

func TestNetEnvCmdStructure(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx, app.NewTestAppContext())

	assert.NotEmpty(t, cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.HasSubCommands())

	// Library Long includes examples for status/watch/profile
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "net-env status")
	assert.Contains(t, cmd.Long, "net-env watch")
	assert.Contains(t, cmd.Long, "net-env profile list")
}

func TestNetEnvCmdHelpContent(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx, app.NewTestAppContext())

	longDesc := cmd.Long
	assert.Contains(t, longDesc, "network environment transitions")
	assert.Contains(t, longDesc, "Network status checking")
	assert.Contains(t, longDesc, "Real-time network monitoring")
	assert.Contains(t, longDesc, "Network profile management")
	assert.Contains(t, longDesc, "Cross-platform support")
}

func TestNetEnvCmdHelpExecute(t *testing.T) {
	ctx := context.Background()
	cmd := NewNetEnvCmd(ctx, app.NewTestAppContext())
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	require.NoError(t, err)
}

//nolint:testpackage // White-box testing needed for internal function access
package ide

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestNewIDECmd(t *testing.T) {
	cmd := NewIDECmd(context.Background(), app.NewTestAppContext())

	assert.Equal(t, "ide", cmd.Use)
	assert.Equal(t, "Monitor and manage IDE configuration changes", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 6)

	var monitorCmd, listCmd, fixSyncCmd, scanCmd, statusCmd, openCmd *cobra.Command

	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "monitor":
			monitorCmd = subcmd
		case "list":
			listCmd = subcmd
		case "fix-sync":
			fixSyncCmd = subcmd
		case "scan":
			scanCmd = subcmd
		case "status":
			statusCmd = subcmd
		case "open <ide-name> [path]":
			openCmd = subcmd
		}
	}

	assert.NotNil(t, monitorCmd)
	assert.Equal(t, "monitor", monitorCmd.Use)
	assert.Equal(t, "Monitor JetBrains settings for changes", monitorCmd.Short)

	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Equal(t, "List detected JetBrains IDE installations", listCmd.Short)

	assert.NotNil(t, fixSyncCmd)
	assert.Equal(t, "fix-sync", fixSyncCmd.Use)
	assert.Equal(t, "Fix JetBrains settings synchronization issues", fixSyncCmd.Short)

	assert.NotNil(t, scanCmd)
	assert.Equal(t, "scan", scanCmd.Use)
	assert.Equal(t, "Scan for installed IDEs", scanCmd.Short)

	assert.NotNil(t, statusCmd)
	assert.Equal(t, "status", statusCmd.Use)
	assert.Equal(t, "Show status of installed IDEs", statusCmd.Short)

	assert.NotNil(t, openCmd)
	assert.Equal(t, "open <ide-name> [path]", openCmd.Use)
	assert.Equal(t, "Open an IDE with specified project path", openCmd.Short)
}

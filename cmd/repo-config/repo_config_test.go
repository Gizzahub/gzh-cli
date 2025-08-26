//nolint:testpackage // White-box testing needed for internal function access
package repoconfig

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestNewRepoConfigCmd(t *testing.T) {
	cmd := NewRepoConfigCmd(app.NewTestAppContext())

	assert.Equal(t, "repo-config", cmd.Use)
	assert.Equal(t, "GitHub repository configuration management", cmd.Short)
	assert.Contains(t, cmd.Long, "infrastructure-as-code")

	expectedSubcommands := []string{"list", "apply", "validate", "diff", "audit", "template"}
	actualSubcommands := make([]string, 0, len(cmd.Commands()))
	for _, subcmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subcmd.Use)
	}
	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected)
	}
}

func TestGlobalFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var flags GlobalFlags
	addGlobalFlags(cmd, &flags)

	expectedFlags := []string{"org", "config", "token", "dry-run", "verbose", "parallel", "timeout"}
	for _, name := range expectedFlags {
		assert.NotNil(t, cmd.Flags().Lookup(name))
	}

	assert.Equal(t, 5, flags.Parallel)
	assert.Equal(t, "30s", flags.Timeout)
	assert.False(t, flags.DryRun)
	assert.False(t, flags.Verbose)
}

func TestListCommand(t *testing.T) {
	cmd := newListCmd()
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List repositories with current configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "repository details")
}

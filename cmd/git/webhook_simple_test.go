//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewWebhookCmd(t *testing.T) {
	cmd := NewWebhookCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "webhook", cmd.Use)
	assert.Contains(t, cmd.Short, "웹훅")
	assert.Contains(t, cmd.Long, "GitHub 웹훅 CRUD API")

	// Check that subcommands are added
	subcommands := cmd.Commands()
	assert.True(t, len(subcommands) > 0)

	// Find specific subcommands
	var repoCmd, orgCmd, bulkCmd, monitorCmd *cobra.Command

	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "repo":
			repoCmd = subcmd
		case "org":
			orgCmd = subcmd
		case "bulk":
			bulkCmd = subcmd
		case "monitor":
			monitorCmd = subcmd
		}
	}

	assert.NotNil(t, repoCmd, "repo subcommand should exist")
	assert.NotNil(t, orgCmd, "org subcommand should exist")
	assert.NotNil(t, bulkCmd, "bulk subcommand should exist")
	assert.NotNil(t, monitorCmd, "monitor subcommand should exist")
}

func TestRepositoryWebhookCommands(t *testing.T) {
	cmd := NewWebhookCmd()
	repoCmd := findSubcommand(cmd, "repo")

	assert.NotNil(t, repoCmd)
	assert.Equal(t, "repo", repoCmd.Use)
	assert.Contains(t, repoCmd.Short, "리포지토리")

	// Check repo subcommands
	subcommands := repoCmd.Commands()
	expectedSubcommands := []string{"create", "list", "get", "update", "delete"}

	for _, expected := range expectedSubcommands {
		found := false

		for _, subcmd := range subcommands {
			if subcmd.Use == expected || subcmd.Use == expected+" <owner> <repo>" ||
				subcmd.Use == expected+" <owner> <repo> <webhook-id>" {
				found = true
				break
			}
		}

		assert.True(t, found, "Expected subcommand %s not found", expected)
	}
}

func TestOrganizationWebhookCommands(t *testing.T) {
	cmd := NewWebhookCmd()
	orgCmd := findSubcommand(cmd, "org")

	assert.NotNil(t, orgCmd)
	assert.Equal(t, "org", orgCmd.Use)
	assert.Contains(t, orgCmd.Short, "조직")

	// Check org subcommands
	subcommands := orgCmd.Commands()
	assert.True(t, len(subcommands) >= 2) // at least create and list
}

func TestBulkWebhookCommands(t *testing.T) {
	cmd := NewWebhookCmd()
	bulkCmd := findSubcommand(cmd, "bulk")

	assert.NotNil(t, bulkCmd)
	assert.Equal(t, "bulk", bulkCmd.Use)
	assert.Contains(t, bulkCmd.Short, "대량")

	// Check bulk subcommands
	subcommands := bulkCmd.Commands()
	assert.True(t, len(subcommands) >= 1) // at least create
}

func TestMonitorWebhookCommands(t *testing.T) {
	cmd := NewWebhookCmd()
	monitorCmd := findSubcommand(cmd, "monitor")

	assert.NotNil(t, monitorCmd)
	assert.Equal(t, "monitor", monitorCmd.Use)
	assert.Contains(t, monitorCmd.Short, "모니터링")

	// Check monitor subcommands
	subcommands := monitorCmd.Commands()
	assert.True(t, len(subcommands) >= 2) // at least test and deliveries
}

func TestCreateRepositoryWebhookFlags(t *testing.T) {
	cmd := NewWebhookCmd()
	repoCmd := findSubcommand(cmd, "repo")
	createCmd := findSubcommand(repoCmd, "create <owner> <repo>")

	assert.NotNil(t, createCmd)

	// Check required flags
	nameFlag := createCmd.Flag("name")
	assert.NotNil(t, nameFlag)

	urlFlag := createCmd.Flag("url")
	assert.NotNil(t, urlFlag)

	eventsFlag := createCmd.Flag("events")
	assert.NotNil(t, eventsFlag)

	activeFlag := createCmd.Flag("active")
	assert.NotNil(t, activeFlag)

	contentTypeFlag := createCmd.Flag("content-type")
	assert.NotNil(t, contentTypeFlag)

	secretFlag := createCmd.Flag("secret")
	assert.NotNil(t, secretFlag)
}

// Helper function to find a subcommand by name or use pattern.
func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == name || cmd.Name() == name {
			return cmd
		}
		// Also check if the use pattern starts with the name
		if len(cmd.Use) > len(name) && cmd.Use[:len(name)] == name {
			return cmd
		}
	}

	return nil
}

//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGitCommands(t *testing.T) {
	tests := []struct {
		name          string
		newCmd        func() *cobra.Command
		use           string
		shortContains string
	}{
		{"root", NewGitCmd, "git", "Git 플랫폼 관리"},
		{"config", newGitConfigCmd, "config", "Repository configuration"},
		{"event", newGitEventCmd, "event", "Event processing"},
		{"repo", NewGitRepoCmd, "repo", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.newCmd()
			assert.Equal(t, tt.use, cmd.Use)
			if tt.shortContains != "" {
				assert.Contains(t, cmd.Short, tt.shortContains)
			}
		})
	}
}

func TestNewGitCmd_Subcommands(t *testing.T) {
	cmd := NewGitCmd()
	subcommands := cmd.Commands()
	got := make(map[string]bool)
	for _, sub := range subcommands {
		got[sub.Use] = true
	}

	tests := []struct{ name string }{
		{"repo"}, {"config"}, {"webhook"}, {"event"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, got[tt.name], "%s subcommand should exist", tt.name)
		})
	}
}

//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestNewGitCmd(t *testing.T) {
	cmd := NewGitCmd(app.NewTestAppContext())

	require.Equal(t, "git", cmd.Use)
	require.Contains(t, cmd.Short, "Git 플랫폼 관리")
	require.NotEmpty(t, cmd.Long)

	// 서브커맨드 확인
	subcommands := cmd.Commands()
	require.GreaterOrEqual(t, len(subcommands), 4) // repo, config, webhook, event

	// 서브커맨드 존재 확인
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Use] = true
	}

	require.True(t, subcommandNames["repo"], "repo subcommand should exist")
	require.True(t, subcommandNames["config"], "config subcommand should exist")
	require.True(t, subcommandNames["webhook"], "webhook subcommand should exist")
	require.True(t, subcommandNames["event"], "event subcommand should exist")
}

func TestNewGitConfigCmd(t *testing.T) {
	cmd := newGitConfigCmd(app.NewTestAppContext())

	require.Equal(t, "config", cmd.Use)
	require.Contains(t, cmd.Short, "Repository configuration")
	require.NotEmpty(t, cmd.Long)
}

func TestNewGitEventCmd(t *testing.T) {
	cmd := newGitEventCmd()

	require.Equal(t, "event", cmd.Use)
	require.Contains(t, cmd.Short, "Event processing")
	require.NotEmpty(t, cmd.Long)
}

func TestNewGitRepoCmd(t *testing.T) {
	cmd := NewGitRepoCmd()

	require.NotNil(t, cmd)
	require.Equal(t, "repo", cmd.Use)
	require.NotEmpty(t, cmd.Short)
}

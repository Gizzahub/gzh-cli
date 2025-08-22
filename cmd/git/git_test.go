//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitCmd(t *testing.T) {
	cmd := NewGitCmd()

	assert.Equal(t, "git", cmd.Use)
	assert.Contains(t, cmd.Short, "Git 플랫폼 관리")
	assert.NotEmpty(t, cmd.Long)

	// 서브커맨드 확인
	subcommands := cmd.Commands()
	assert.True(t, len(subcommands) >= 4) // repo, config, webhook, event

	// 서브커맨드 존재 확인
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Use] = true
	}

	assert.True(t, subcommandNames["repo"], "repo subcommand should exist")
	assert.True(t, subcommandNames["config"], "config subcommand should exist")
	assert.True(t, subcommandNames["webhook"], "webhook subcommand should exist")
	assert.True(t, subcommandNames["event"], "event subcommand should exist")
}

func TestNewGitConfigCmd(t *testing.T) {
	cmd := newGitConfigCmd()

	assert.Equal(t, "config", cmd.Use)
	assert.Contains(t, cmd.Short, "Repository configuration")
	assert.NotEmpty(t, cmd.Long)
}

func TestNewGitEventCmd(t *testing.T) {
	cmd := newGitEventCmd()

	assert.Equal(t, "event", cmd.Use)
	assert.Contains(t, cmd.Short, "Event processing")
	assert.NotEmpty(t, cmd.Long)
}

func TestNewGitRepoCmd(t *testing.T) {
	cmd := NewGitRepoCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "repo", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

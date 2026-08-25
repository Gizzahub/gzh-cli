//nolint:testpackage // White-box testing needed for internal function access
package synclone

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
	gerrors "github.com/gizzahub/gzh-cli/internal/errors"
)

func TestDefaultSyncCloneOptions(t *testing.T) {
	t.Run("github default options", func(t *testing.T) {
		opts := defaultSyncCloneGithubOptions()
		assert.Equal(t, "reset", opts.strategy)
	})

	t.Run("gitlab default options", func(t *testing.T) {
		opts := defaultSyncCloneGitlabOptions()
		assert.Equal(t, "reset", opts.strategy)
	})

	t.Run("gitea default options", func(t *testing.T) {
		opts := defaultSyncCloneGiteaOptions()
		assert.Equal(t, "reset", opts.strategy)
	})
}

func TestBulkCloneConfigSupport(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test config file
	configContent := `version: "1.0"
default:
  protocol: https
  github:
    root_path: '%s/github-repos'
    org_name: "test-default-org"
repo_roots:
  - root_path: '%s/my-repos'
    provider: "github"
    protocol: "https"
    org_name: "my-test-org"
`
	configPath := filepath.Join(tempDir, "test-config.yaml")
	// YAML double quotes interpret backslashes as escapes. Single quotes preserve
	// native Windows paths exactly as users provide them.
	formattedConfig := fmt.Sprintf(configContent, tempDir, tempDir)
	err := os.WriteFile(configPath, []byte(formattedConfig), 0o600)
	require.NoError(t, err)

	t.Run("github with config file", func(t *testing.T) {
		opts := &syncCloneGithubOptions{
			configFile: configPath,
			orgName:    "my-test-org",
		}

		err := opts.loadFromConfig()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, "my-repos"), opts.targetPath)
	})

	t.Run("github with config use default org", func(t *testing.T) {
		opts := &syncCloneGithubOptions{
			configFile: configPath,
			orgName:    "test-default-org",
		}

		err := opts.loadFromConfig()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, "github-repos"), opts.targetPath)
	})

	t.Run("gitlab with config file", func(t *testing.T) {
		// Create GitLab config
		gitlabConfig := `version: "1.0"
default:
  protocol: https
  gitlab:
    root_path: '%s/gitlab-repos'
    group_name: "test-group"
    recursive: true
`
		gitlabConfigPath := filepath.Join(tempDir, "gitlab-config.yaml")
		formattedGitlabConfig := fmt.Sprintf(gitlabConfig, tempDir)
		err := os.WriteFile(gitlabConfigPath, []byte(formattedGitlabConfig), 0o600)
		require.NoError(t, err)

		opts := &syncCloneGitlabOptions{
			configFile: gitlabConfigPath,
			groupName:  "test-group",
		}

		err = opts.loadFromConfig()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, "gitlab-repos"), opts.targetPath)
		assert.True(t, opts.recursively)
	})

	t.Run("cli flags override config", func(t *testing.T) {
		opts := &syncCloneGithubOptions{
			configFile: configPath,
			orgName:    "my-test-org",
			targetPath: "/override/path",
		}

		err := opts.loadFromConfig()
		assert.NoError(t, err)
		// CLI flag should take precedence
		assert.Equal(t, "/override/path", opts.targetPath)
	})
}

func TestMainSyncCloneCommand(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("default synclone options", func(t *testing.T) {
		opts := defaultSyncCloneOptions()
		assert.Equal(t, "reset", opts.strategy)
		assert.Equal(t, "", opts.configFile)
		assert.False(t, opts.useConfig)
	})

	t.Run("strategy validation", func(t *testing.T) {
		validStrategies := []string{"reset", "pull", "fetch"}
		invalidStrategies := []string{"invalid", "merge", "rebase"}

		// Create a minimal config for testing
		configContent := `version: "1.0"
default:
  protocol: https
repo_roots: []
`
		configPath := filepath.Join(tempDir, "minimal-config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		require.NoError(t, err)

		for _, strategy := range validStrategies {
			t.Run("valid strategy: "+strategy, func(t *testing.T) {
				opts := &syncCloneOptions{
					configFile: configPath,
					strategy:   strategy,
				}

				// The command should not fail on strategy validation
				// It might fail on network operations, but not on validation
				err := opts.run(context.Background(), nil, []string{})
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid strategy")
				}
			})
		}

		for _, strategy := range invalidStrategies {
			t.Run("invalid strategy: "+strategy, func(t *testing.T) {
				opts := &syncCloneOptions{
					configFile: configPath,
					strategy:   strategy,
				}

				err := opts.run(context.Background(), nil, []string{})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid strategy")
			})
		}
	})

	t.Run("config loading", func(t *testing.T) {
		// Create a comprehensive config
		configContent := `version: "1.0"
default:
  protocol: https
  github:
    root_path: "%s/default-github"
    org_name: "default-github-org"
  gitlab:
    root_path: "%s/default-gitlab"
    group_name: "default-gitlab-group"
repo_roots:
  - root_path: "%s/github-org1"
    provider: "github"
    protocol: "ssh"
    org_name: "github-org1"
  - root_path: "%s/github-org2"
    provider: "github"
    protocol: "https"
    org_name: "github-org2"
  - root_path: "%s/gitlab-group1"
    provider: "gitlab"
    protocol: "https"
    org_name: "gitlab-group1"
`
		configPath := filepath.Join(tempDir, "comprehensive-config.yaml")
		formattedConfig := fmt.Sprintf(configContent, tempDir, tempDir, tempDir, tempDir, tempDir)
		err := os.WriteFile(configPath, []byte(formattedConfig), 0o600)
		require.NoError(t, err)

		opts := &syncCloneOptions{
			configFile: configPath,
			strategy:   "fetch",
		}

		// Since we don't have actual git repositories, this will fail
		// but we can verify that config loading and processing works
		err = opts.run(context.Background(), nil, []string{})
		// The error should come from git operations, not from config processing
		if err != nil {
			assert.NotContains(t, err.Error(), "failed to load config")
			assert.NotContains(t, err.Error(), "invalid strategy")
		}
	})

	t.Run("missing config", func(t *testing.T) {
		opts := &syncCloneOptions{
			configFile: "/non/existent/config.yaml",
			strategy:   "reset",
		}

		err := opts.run(context.Background(), nil, []string{})
		require.Error(t, err)

		// 표지로 확인한다. 예전에는 "failed to load config"라는 글자를 찾았는데
		// 그 말은 어느 계층도 쓰지 않아서 통과할 수 없는 확인이었다.
		assert.ErrorIs(t, err, gerrors.ErrConfigNotFound)
		assert.Contains(t, err.Error(), "/non/existent/config.yaml")

		// 표지가 겹쳐 나오지 않아야 한다. 파사드와 서비스와 명령이 차례로
		// 감싸면서 같은 말을 세 번 붙이던 적이 있다.
		assert.Equal(t, 1, strings.Count(err.Error(), "config not found"))
	})

	t.Run("empty config", func(t *testing.T) {
		// Create an empty config
		configContent := `version: "1.0"
default:
  protocol: https
repo_roots: []
`
		configPath := filepath.Join(tempDir, "empty-config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		require.NoError(t, err)

		opts := &syncCloneOptions{
			configFile: configPath,
			strategy:   "reset",
		}

		// 아무것도 설정하지 않은 파일은 오류로 답한다. 예전 기대값은 조용히
		// 성공하는 것이었지만 그 확인은 돌아본 적이 없다 -- 설정을 읽는 쪽이
		// 교착에 빠져 이 하위 시험까지 오지 못했다.
		//
		// 조용히 0으로 끝나면 사용자는 왜 아무것도 받아지지 않았는지 알 길이
		// 없다. 거를 것이 있어서 대상이 0이 된 경우는 여전히 "No targets found
		// to process"로 성공한다. 이건 그것과 다른, 설정 자체가 빈 경우다.
		err = opts.run(context.Background(), nil, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one provider must be configured")
	})
}

func TestLegacySyncloneDeprecation(t *testing.T) {
	cmd := NewSyncCloneCmd(app.NewTestAppContext())
	require.NotNil(t, cmd)

	for _, provider := range []string{"github", "gitlab", "gitea"} {
		t.Run(provider, func(t *testing.T) {
			sub, _, err := cmd.Find([]string{provider})
			require.NoError(t, err)
			require.NotNil(t, sub)
			require.NotEmpty(t, sub.Deprecated, "legacy %s command must set cobra Deprecated", provider)
			assert.Contains(t, sub.Deprecated, "forge",
				"deprecation message for %s must mention forge replacement path", provider)
			assert.Contains(t, sub.Deprecated, provider,
				"deprecation message should include the provider name %s", provider)
			assert.Contains(t, sub.Deprecated, "gz synclone forge --provider "+provider)
		})
	}
}

func TestMainSyncCloneCommandFlags(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := NewSyncCloneCmd(app.NewTestAppContext())
		assert.NotNil(t, cmd)
		assert.Equal(t, "synclone", cmd.Use)
		// bulk-clone에서 synclone으로 이름이 바뀌면서 설명도 "Synchronize and
		// clone"으로 바뀌었는데 여기 기대값만 옛 문구로 남아 있었다.
		assert.Contains(t, cmd.Short, "Synchronize and clone repositories")

		// Check that it has the right flags
		configFlag := cmd.Flags().Lookup("config")
		assert.NotNil(t, configFlag)
		assert.Equal(t, "c", configFlag.Shorthand)

		useConfigFlag := cmd.Flags().Lookup("use-config")
		assert.NotNil(t, useConfigFlag)

		strategyFlag := cmd.Flags().Lookup("strategy")
		assert.NotNil(t, strategyFlag)
		assert.Equal(t, "s", strategyFlag.Shorthand)
		assert.Equal(t, "reset", strategyFlag.DefValue)

		// Check that it has subcommands
		subcommands := cmd.Commands()

		subcommandNames := make([]string, len(subcommands))
		for i, subcmd := range subcommands {
			subcommandNames[i] = subcmd.Use
		}

		assert.Contains(t, subcommandNames, "github")
		assert.Contains(t, subcommandNames, "gitlab")
		assert.Contains(t, subcommandNames, "gitea")
		assert.Contains(t, subcommandNames, "validate")
	})
}

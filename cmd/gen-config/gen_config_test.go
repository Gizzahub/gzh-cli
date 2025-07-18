package genconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenConfigTemplate(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("generate simple template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "simple.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"simple"})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "version: \"0.1\"")
		assert.Contains(t, configText, "provider: \"github\"")
		assert.Contains(t, configText, "org_name: \"myorg\"")
		assert.Contains(t, configText, "protocol: \"https\"")
	})

	t.Run("generate comprehensive template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "comprehensive.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"comprehensive"})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "mycompany")
		assert.Contains(t, configText, "kubernetes")
		assert.Contains(t, configText, "gitlab")
		assert.Contains(t, configText, "gitea")
		assert.Contains(t, configText, "ssh")
		assert.Contains(t, configText, "https")
	})

	t.Run("generate work template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "work.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"work"})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "company")
		assert.Contains(t, configText, "devops")
		assert.Contains(t, configText, "ssh")
		assert.Contains(t, configText, "poc-.*")
	})

	t.Run("generate personal template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "personal.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"personal"})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "yourusername")
		assert.Contains(t, configText, "kubernetes")
		assert.Contains(t, configText, "golang")
		assert.Contains(t, configText, "fork-.*")
	})

	t.Run("generate multi-org template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "multi-org.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"multi-org"})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "company")
		assert.Contains(t, configText, "client1")
		assert.Contains(t, configText, "client2")
		assert.Contains(t, configText, "hashicorp")
	})

	t.Run("unknown template returns error", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "unknown.yaml")
		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err := opts.run(nil, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown template")
	})

	t.Run("existing file without force returns error", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "existing.yaml")
		err := os.WriteFile(outputFile, []byte("existing content"), 0o644)
		require.NoError(t, err)

		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      false,
		}

		err = opts.run(nil, []string{"simple"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("existing file with force overwrites", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "force-overwrite.yaml")
		err := os.WriteFile(outputFile, []byte("existing content"), 0o644)
		require.NoError(t, err)

		opts := &genConfigTemplateOptions{
			outputFile: outputFile,
			force:      true,
		}

		err = opts.run(nil, []string{"simple"})
		assert.NoError(t, err)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "version: \"0.1\"")
		assert.NotContains(t, string(content), "existing content")
	})
}

func TestGenConfigDiscover(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock git repositories
	createMockGitRepo(t, tempDir, "github-org1", "repo1", "https://github.com/org1/repo1.git")
	createMockGitRepo(t, tempDir, "github-org1", "repo2", "git@github.com:org1/repo2.git")
	createMockGitRepo(t, tempDir, "gitlab-group1", "project1", "https://gitlab.com/group1/project1.git")

	t.Run("discover repositories in directory", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "discovered.yaml")
		opts := &genConfigDiscoverOptions{
			directory:  tempDir,
			outputFile: outputFile,
			recursive:  true,
			force:      false,
		}

		err := opts.run(nil, []string{})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "version: \"0.1\"")
		assert.Contains(t, configText, "provider: \"github\"")
		assert.Contains(t, configText, "org_name: \"org1\"")
		assert.Contains(t, configText, "provider: \"gitlab\"")
		assert.Contains(t, configText, "org_name: \"group1\"")
	})

	t.Run("discover with custom directory argument", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "custom-dir.yaml")
		opts := &genConfigDiscoverOptions{
			outputFile: outputFile,
			recursive:  true,
			force:      false,
		}

		err := opts.run(nil, []string{tempDir})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)
	})

	t.Run("no repositories found", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		err := os.MkdirAll(emptyDir, 0o755)
		require.NoError(t, err)

		outputFile := filepath.Join(tempDir, "empty-result.yaml")
		opts := &genConfigDiscoverOptions{
			directory:  emptyDir,
			outputFile: outputFile,
			recursive:  true,
			force:      false,
		}

		err = opts.run(nil, []string{})
		assert.NoError(t, err)
		assert.NoFileExists(t, outputFile) // Should not create file when no repos found
	})
}

func TestGenConfigInit(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("non-interactive with simple template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "init-simple.yaml")
		opts := &genConfigInitOptions{
			outputFile:  outputFile,
			interactive: false,
			template:    "simple",
		}

		err := opts.run(nil, []string{})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "version: \"0.1\"")
		assert.Contains(t, configText, "org_name: \"myorg\"")
	})

	t.Run("non-interactive with comprehensive template", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "init-comprehensive.yaml")
		opts := &genConfigInitOptions{
			outputFile:  outputFile,
			interactive: false,
			template:    "comprehensive",
		}

		err := opts.run(nil, []string{})
		assert.NoError(t, err)
		assert.FileExists(t, outputFile)

		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		configText := string(content)
		assert.Contains(t, configText, "mycompany")
		assert.Contains(t, configText, "kubernetes")
		assert.Contains(t, configText, "gitlab")
	})
}

func TestParseRemoteURL(t *testing.T) {
	opts := &genConfigDiscoverOptions{}

	tests := []struct {
		name         string
		remoteURL    string
		wantProvider string
		wantOrg      string
		wantRepo     string
		wantProtocol string
	}{
		{
			name:         "GitHub SSH URL",
			remoteURL:    "git@github.com:org/repo.git",
			wantProvider: "github",
			wantOrg:      "org",
			wantRepo:     "repo",
			wantProtocol: "ssh",
		},
		{
			name:         "GitHub HTTPS URL",
			remoteURL:    "https://github.com/org/repo.git",
			wantProvider: "github",
			wantOrg:      "org",
			wantRepo:     "repo",
			wantProtocol: "https",
		},
		{
			name:         "GitLab SSH URL",
			remoteURL:    "git@gitlab.com:group/project.git",
			wantProvider: "gitlab",
			wantOrg:      "group",
			wantRepo:     "project",
			wantProtocol: "ssh",
		},
		{
			name:         "GitLab HTTPS URL",
			remoteURL:    "https://gitlab.com/group/project.git",
			wantProvider: "gitlab",
			wantOrg:      "group",
			wantRepo:     "project",
			wantProtocol: "https",
		},
		{
			name:         "GitHub URL without .git suffix",
			remoteURL:    "https://github.com/org/repo",
			wantProvider: "github",
			wantOrg:      "org",
			wantRepo:     "repo",
			wantProtocol: "https",
		},
		{
			name:         "Invalid URL",
			remoteURL:    "invalid-url",
			wantProvider: "",
			wantOrg:      "",
			wantRepo:     "",
			wantProtocol: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, org, repo, protocol := opts.parseRemoteURL(tt.remoteURL)
			assert.Equal(t, tt.wantProvider, provider)
			assert.Equal(t, tt.wantOrg, org)
			assert.Equal(t, tt.wantRepo, repo)
			assert.Equal(t, tt.wantProtocol, protocol)
		})
	}
}

func TestGenConfigCommand(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := NewGenConfigCmd(context.Background())
		assert.NotNil(t, cmd)
		assert.Equal(t, "gen-config", cmd.Use)
		assert.Contains(t, cmd.Short, "Generate bulk-clone configuration files")

		// Check that it has subcommands
		subcommands := cmd.Commands()

		subcommandNames := make([]string, len(subcommands))
		for i, subcmd := range subcommands {
			subcommandNames[i] = subcmd.Use
		}

		assert.Contains(t, subcommandNames, "init")
		assert.Contains(t, subcommandNames, "template <template-name>")
		assert.Contains(t, subcommandNames, "discover [directory]")
		assert.Contains(t, subcommandNames, "github") // Legacy command
	})
}

// Helper function to create mock Git repositories for testing.
func createMockGitRepo(t *testing.T, baseDir, orgDir, repoName, remoteURL string) {
	repoPath := filepath.Join(baseDir, orgDir, repoName)
	gitDir := filepath.Join(repoPath, ".git")

	err := os.MkdirAll(gitDir, 0o755)
	require.NoError(t, err)

	// Create mock .git/config file with remote URL
	configContent := fmt.Sprintf(`[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`, remoteURL)

	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Create HEAD file
	headPath := filepath.Join(gitDir, "HEAD")
	err = os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0o644)
	require.NoError(t, err)
}

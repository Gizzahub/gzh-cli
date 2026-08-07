// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // White-box testing needed for internal function access
package scenarios

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli/test/e2e/helpers"
)

// Valid modern synclone config (repo_roots schema used by `synclone validate`).
const modernSyncloneConfig = `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./repos"
    provider: "github"
    protocol: "https"
    org_name: "test-org"
ignore_names:
  - "^temp.*"
`

// Providers/organizations format accepted by `synclone config validate` (YAML syntax only).
const organizationsConfig = `
version: "1.0.0"
global:
  clone_base_dir: ${HOME}/repos
  default_strategy: pull
providers:
  github:
    organizations:
      - name: test-org
        clone_dir: ${HOME}/repos/test-org
`

func TestSyncloneConfig_ValidateYAML_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("valid.yaml", organizationsConfig)

	result := env.RunCommand("synclone", "config", "validate", "--file", "valid.yaml")

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("valid")
}

func TestSyncloneConfig_ValidateSchema_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("schema-valid.yaml", modernSyncloneConfig)

	// Full schema validation lives on `synclone validate` (not bare `config`).
	result := env.RunCommand("synclone", "validate", "--config", "schema-valid.yaml")

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("Configuration is valid")
}

func TestSyncloneConfig_ValidateInvalidYAML_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("broken.yaml", "version: [\n")

	result := env.RunCommand("synclone", "config", "validate", "--file", "broken.yaml")

	helpers.NewCLIAssertions(t, result).
		Failure().
		OutputContains("invalid YAML")
}

func TestSyncloneConfig_ValidateMissingFile_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("synclone", "config", "validate", "--file", "does-not-exist.yaml")

	helpers.NewCLIAssertions(t, result).Failure()
}

func TestSyncloneConfig_ValidateSchemaRejectsBadProtocol_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("bad-protocol.yaml", `
version: "1.0"
default:
  protocol: not-a-protocol
repo_roots:
  - root_path: "./repos"
    provider: "github"
    protocol: "https"
    org_name: "test-org"
`)

	result := env.RunCommand("synclone", "validate", "--config", "bad-protocol.yaml")

	helpers.NewCLIAssertions(t, result).Failure()
}

func TestSyncloneConfig_GenerateTemplateList_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("synclone", "config", "generate", "template", "--list-templates")

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("minimal")
}

func TestSyncloneConfig_GenerateTemplateMinimal_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand(
		"synclone", "config", "generate", "template",
		"--template", "minimal",
		"--var", "GitHubOrg=e2e-org",
		"--output", "from-template.yaml",
	)

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("generated")

	env.AssertFileExists("from-template.yaml")
	config := helpers.NewConfigAssertions(t, env, "from-template.yaml")
	config.ValidYAML().HasField("version").HasField("providers")
}

func TestSyncloneConfig_GenerateDiscover_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Discover needs a real git repo with a remote, not a stub .git dir.
	createRealGitRepo(t, env.WorkDir, "org1/repo1", "https://github.com/e2e-org/repo1.git")

	result := env.RunCommand(
		"synclone", "config", "generate", "discover",
		"--path", ".",
		"--output", "discovered.yaml",
	)

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("Found")

	env.AssertFileExists("discovered.yaml")
	config := helpers.NewConfigAssertions(t, env, "discovered.yaml")
	config.ValidYAML().HasField("version").HasField("providers")
}

func TestSyncloneConfig_GenerateHelp_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("synclone", "config", "--help")

	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("generate").
		OutputContains("validate").
		OutputContains("convert")
}

func TestSyncloneConfig_TopLevelConfigAbsent_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Bare `gz config` was removed; document current surface: unknown command.
	result := env.RunCommand("config", "--help")

	helpers.NewCLIAssertions(t, result).
		Failure().
		OutputContains("unknown command")
}

// createRealGitRepo initializes a git repository with a remote for discovery tests.
func createRealGitRepo(t *testing.T, workDir, relativePath, remoteURL string) {
	t.Helper()

	repoPath := filepath.Join(workDir, relativePath)
	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=e2e",
			"GIT_AUTHOR_EMAIL=e2e@example.com",
			"GIT_COMMITTER_NAME=e2e",
			"GIT_COMMITTER_EMAIL=e2e@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	runGit("remote", "add", "origin", remoteURL)
}

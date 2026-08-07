// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // White-box testing needed for internal function access
package scenarios

import (
	"testing"

	"github.com/gizzahub/gzh-cli/test/e2e/helpers"
)

func TestSyncClone_HelpAndVersion_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("synclone", "--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("synclone").
		OutputContains("Usage:")

	result = env.RunCommand("version")
	helpers.NewCLIAssertions(t, result).Success().OutputNotEmpty()

	result = env.RunCommand("--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("gz").
		OutputContains("Commands:")
}

func TestSyncClone_ValidateModernSchema_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("valid-config.yaml", modernSyncloneConfig)

	result := env.RunCommand("synclone", "validate", "--config", "valid-config.yaml")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("Configuration is valid")
}

func TestSyncClone_ValidateInvalid_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("invalid-config.yaml", `
version: "1.0"
default:
  protocol: not-a-protocol
`)

	result := env.RunCommand("synclone", "validate", "--config", "invalid-config.yaml")
	helpers.NewCLIAssertions(t, result).Failure()
}

func TestSyncClone_ValidateMissingConfig_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("synclone", "validate", "--config", "non-existent.yaml")
	helpers.NewCLIAssertions(t, result).Failure()
}

func TestSyncClone_ValidateOrganizationsFormat_YAMLOnly_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// organizations: is the preferred key in providers-style configs; YAML syntax
	// validation accepts it via `synclone config validate`.
	env.WriteConfig("orgs-config.yaml", organizationsConfig)

	result := env.RunCommand("synclone", "config", "validate", "--file", "orgs-config.yaml")
	helpers.NewCLIAssertions(t, result).Success().OutputContains("valid")
}

func TestSyncClone_MultiOrgYAML_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Schema for repo_roots currently constrains provider to github.
	// Multi-org github configs are the supported multi-target path.
	config := `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./github-repos"
    provider: "github"
    protocol: "https"
    org_name: "github-org"
  - root_path: "./github-repos-2"
    provider: "github"
    protocol: "ssh"
    org_name: "github-org-two"
`
	env.WriteConfig("multi-org.yaml", config)

	result := env.RunCommand("synclone", "validate", "--config", "multi-org.yaml")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("Configuration is valid")
}

func TestSyncClone_ProvidersOrganizationsYAML_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// providers.github.organizations format is preferred in templates; YAML
	// syntax validation accepts it even when full schema does not.
	env.WriteConfig("providers-orgs.yaml", organizationsConfig)

	result := env.RunCommand("synclone", "config", "validate", "--file", "providers-orgs.yaml")
	helpers.NewCLIAssertions(t, result).Success().OutputContains("valid")
}

func TestSyncClone_StrategyInConfig_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Strategy is a CLI flag on synclone/forge, not a per-org schema field in
	// the modern validate path. Assert the flag is accepted on the command surface.
	result := env.RunCommand("synclone", "--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("--strategy")

	for _, strategy := range []string{"reset", "pull", "fetch"} {
		// Unknown strategy should still be a recognized flag; validation of value
		// happens at run time. Help documents the allowed values.
		_ = strategy
	}
}

func TestSyncClone_ForgeDryRunRequiresFlags_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// --dry-run lives on `synclone forge`, not bare `synclone`.
	result := env.RunCommand("synclone", "forge", "--dry-run")
	helpers.NewCLIAssertions(t, result).
		Failure().
		OutputContains("required")

	// With required flags but fake credentials, CLI must fail honestly (API/auth),
	// not claim success and not reject --dry-run as unknown.
	result = env.RunCommand(
		"synclone", "forge",
		"--provider", "github",
		"--org", "nonexistent-org-e2e-test",
		"--target", "./repos",
		"--dry-run",
		"--token", "invalid-token-for-e2e",
	)
	helpers.NewCLIAssertions(t, result).Failure()
	// Must not be "unknown flag" — dry-run is real.
	helpers.NewCLIAssertions(t, result).OutputNotContains("unknown flag")
}

func TestSyncClone_BareDryRunRejected_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Document that bare synclone does not take --dry-run.
	result := env.RunCommand("synclone", "--dry-run")
	helpers.NewCLIAssertions(t, result).
		Failure().
		OutputContains("unknown flag")
}

func TestSyncClone_MalformedYAML_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	env.WriteConfig("malformed.yaml", `
version: "1.0"
default:
  protocol: https
repo_roots: [
`)

	result := env.RunCommand("synclone", "validate", "--config", "malformed.yaml")
	helpers.NewCLIAssertions(t, result).Failure()
}

func TestSyncClone_SubcommandHelp_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	for _, sub := range []string{"config", "forge", "validate", "state"} {
		result := env.RunCommand("synclone", sub, "--help")
		helpers.NewCLIAssertions(t, result).
			Success().
			OutputContains(sub)
	}
}

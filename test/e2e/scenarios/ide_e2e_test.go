// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // White-box testing needed for internal function access
package scenarios

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gizzahub/gzh-cli/test/e2e/helpers"
)

func TestIDE_HelpAndUsage_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("ide", "--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("ide").
		OutputContains("monitor").
		OutputContains("list").
		OutputContains("fix-sync").
		OutputContains("scan").
		OutputContains("status")

	for _, subcmd := range []string{"monitor", "list", "fix-sync", "scan", "status", "open"} {
		result = env.RunCommand("ide", subcmd, "--help")
		helpers.NewCLIAssertions(t, result).
			Success().
			OutputContains(subcmd)
	}
}

func TestIDE_List_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create a mock JetBrains product dir under the platform-specific path
	// that `ide list` actually scans (HOME is isolated by TestEnvironment).
	createMockJetBrainsProduct(t, env)

	result := env.RunCommand("ide", "list")
	helpers.NewCLIAssertions(t, result).Success()

	// On isolated HOME with mock product, should report the product or at least
	// a deterministic empty-state message — never crash.
	if result.Output == "" {
		t.Fatal("ide list produced empty output")
	}
}

func TestIDE_Scan_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("ide", "scan")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("IDE")
}

func TestIDE_Status_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	result := env.RunCommand("ide", "status")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("IDE")
}

func TestIDE_FixSync_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// fix-sync takes --product, not --target.
	result := env.RunCommand("ide", "fix-sync", "--product", "NonexistentProduct2024.1")
	// Empty install set is success with a clear message; must not be unknown-flag.
	helpers.NewCLIAssertions(t, result).OutputNotContains("unknown flag")
	if result.ExitCode != 0 && result.ExitCode != 1 {
		t.Fatalf("unexpected exit code %d\nOutput: %s", result.ExitCode, result.Output)
	}
}

func TestIDE_MonitorFlags_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Document current flag surface: --watch-dir exists, --target does not.
	result := env.RunCommand("ide", "monitor", "--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("--watch-dir").
		OutputNotContains("--target")

	// Reject legacy --target so regressions of the flag name are visible.
	result = env.RunCommand("ide", "monitor", "--target", "/tmp")
	helpers.NewCLIAssertions(t, result).
		Failure().
		OutputContains("unknown flag")
}

func TestIDE_RemovedCommandsAbsent_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// backup/validate/restore are not shipped. Parent `ide` has no RunE, so
	// cobra prints help for unknown args with exit 0 — assert the surface
	// does not advertise those names as Available Commands.
	result := env.RunCommand("ide", "--help")
	helpers.NewCLIAssertions(t, result).
		Success().
		OutputContains("Available Commands:").
		OutputNotContains("backup").
		OutputNotContains("validate").
		OutputNotContains("restore")
}

func createMockJetBrainsProduct(t *testing.T, env *helpers.TestEnvironment) {
	t.Helper()

	var rel string
	switch runtime.GOOS {
	case "darwin":
		rel = "Library/Application Support/JetBrains/IntelliJIdea2024.1/options"
	case "linux":
		rel = ".config/JetBrains/IntelliJIdea2024.1/options"
	case "windows":
		rel = "AppData/Roaming/JetBrains/IntelliJIdea2024.1/options"
	default:
		t.Skipf("unsupported GOOS for JetBrains mock path: %s", runtime.GOOS)
	}

	content := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="TestComponent">
    <option name="test" value="true" />
  </component>
</application>`

	path := env.GetHomePath(filepath.Join(rel, "ide.xml"))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir jetbrains mock: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write jetbrains mock: %v", err)
	}
}

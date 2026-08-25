// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // White-box testing needed for internal function access
package netenv_integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli/test/integration/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBinary holds the path to a freshly built gz binary for this package.
var testBinary string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "gz-netenv-it-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		os.Exit(1)
	}

	binaryPath := testutil.ExecutablePath(tmpDir, "gz")
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find project root: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/gz")
	cmd.Dir = projectRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "go build ./cmd/gz failed: %v: %s\n", err, strings.TrimSpace(stderr.String()))
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	testBinary = binaryPath
	code := m.Run()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Prefer the gzh-cli module root (this package lives under test/integration).
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// isolatedHome returns a temp HOME so profile commands do not touch the real home.
func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Library may honor CONFIG_DIR (gzh-cli-core) or GZH_CONFIG_DIR.
	cfg := filepath.Join(home, ".config", "gzh-manager")
	require.NoError(t, os.MkdirAll(cfg, 0o750))
	t.Setenv("GZH_CONFIG_DIR", cfg)
	t.Setenv("CONFIG_DIR", cfg)
	return home
}

func runGZ(t *testing.T, args ...string) (string, error) {
	t.Helper()
	require.NotEmpty(t, testBinary, "test binary must be built in TestMain")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, testBinary, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustRunGZ(t *testing.T, args ...string) string {
	t.Helper()
	out, err := runGZ(t, args...)
	require.NoError(t, err, "gz %s failed:\n%s", strings.Join(args, " "), out)
	return out
}

func availableNetEnvCommands(t *testing.T) string {
	t.Helper()
	out, err := runGZ(t, "net-env", "--help")
	if err != nil {
		return out
	}
	return out
}

func failIfUnknownCommand(t *testing.T, out string, err error, want string) {
	t.Helper()
	combined := out
	if err != nil {
		combined += "\n" + err.Error()
	}
	if strings.Contains(combined, "unknown command") {
		t.Fatalf("command missing from binary (wanted %q).\nOutput: %s\nAvailable net-env help:\n%s",
			want, out, availableNetEnvCommands(t))
	}
}

// TestNetEnvCLIIntegration exercises the commands that exist today:
// status, watch (help only), profile list/init/show.
func TestNetEnvCLIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	t.Run("HelpListsCurrentSurface", func(t *testing.T) {
		out := mustRunGZ(t, "net-env", "--help")
		assert.Contains(t, out, "status")
		assert.Contains(t, out, "profile")
		assert.Contains(t, out, "watch")
		// switch was removed; tests must not revive it silently.
		assert.NotContains(t, out, "  switch")
	})

	t.Run("NetworkStatusCheck", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "status")
		failIfUnknownCommand(t, out, err, "net-env status")
		require.NoError(t, err, "net-env status failed:\n%s", out)
		assert.Contains(t, out, "Network Environment Status")
	})

	t.Run("StatusJSONFormat", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "status", "--format", "json")
		failIfUnknownCommand(t, out, err, "net-env status --format json")
		require.NoError(t, err, "status --format json failed:\n%s", out)
		assert.Contains(t, out, "{")
	})

	t.Run("InitializeNetworkProfiles", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "profile", "init")
		failIfUnknownCommand(t, out, err, "net-env profile init")
		require.NoError(t, err, "profile init failed:\n%s", out)
		assert.Contains(t, out, "home")
		assert.Contains(t, out, "office")
	})

	t.Run("ListNetworkProfiles", func(t *testing.T) {
		// Ensure defaults exist.
		_, _ = runGZ(t, "net-env", "profile", "init")

		out, err := runGZ(t, "net-env", "profile", "list")
		failIfUnknownCommand(t, out, err, "net-env profile list")
		require.NoError(t, err, "profile list failed:\n%s", out)
		assert.Contains(t, out, "home")
		assert.Contains(t, out, "office")
	})

	t.Run("ShowNetworkProfile", func(t *testing.T) {
		_, _ = runGZ(t, "net-env", "profile", "init")

		out, err := runGZ(t, "net-env", "profile", "show", "home")
		failIfUnknownCommand(t, out, err, "net-env profile show")
		require.NoError(t, err, "profile show failed:\n%s", out)
		assert.Contains(t, out, "home")
	})
}

func TestNetworkProfileManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	t.Run("InitCreatesDefaults", func(t *testing.T) {
		out := mustRunGZ(t, "net-env", "profile", "init")
		assert.Contains(t, out, "Default profiles")

		list := mustRunGZ(t, "net-env", "profile", "list")
		for _, name := range []string{"home", "office", "cafe"} {
			assert.Contains(t, list, name)
		}
	})

	t.Run("ShowMissingProfileFails", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "profile", "show", "no-such-profile")
		failIfUnknownCommand(t, out, err, "net-env profile show")
		require.Error(t, err, "expected failure for missing profile, got output:\n%s", out)
		assert.Contains(t, strings.ToLower(out), "not found")
	})

	t.Run("DeleteMissingProfileFails", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "profile", "delete", "no-such-profile")
		failIfUnknownCommand(t, out, err, "net-env profile delete")
		require.Error(t, err, "expected failure for missing profile delete, got:\n%s", out)
		assert.Contains(t, strings.ToLower(out), "not found")
	})
}

func TestCliErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	t.Run("SwitchCommandRemoved", func(t *testing.T) {
		// switch was removed. Parent net-env has no RunE, so unknown args may
		// print help with exit 0 — document absence from Available Commands.
		out := mustRunGZ(t, "net-env", "--help")
		assert.Contains(t, out, "Available Commands:")
		// Command table uses leading spaces + name; avoid matching prose.
		for _, line := range strings.Split(out, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "switch" {
				t.Fatalf("net-env still lists switch (removed surface):\n%s", line)
			}
		}
	})

	t.Run("NonexistentProfile", func(t *testing.T) {
		out, err := runGZ(t, "net-env", "profile", "show", "nonexistent-profile")
		require.Error(t, err, "missing profile must fail:\n%s", out)
		assert.Contains(t, strings.ToLower(out), "not found")
	})
}

func TestConcurrentStatusChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	const (
		numGoroutines = 5
		numIterations = 3
	)

	results := make(chan error, numGoroutines*numIterations)

	for range numGoroutines {
		go func() {
			for range numIterations {
				_, err := runGZ(t, "net-env", "status")
				results <- err
			}
		}()
	}

	successCount := 0
	for range numGoroutines * numIterations {
		err := <-results
		if err == nil {
			successCount++
		} else {
			t.Logf("status failed: %v", err)
		}
	}

	// Concurrent status must actually succeed — zero successes is a failure.
	require.Equal(t, numGoroutines*numIterations, successCount,
		"expected all concurrent status checks to succeed, got %d/%d",
		successCount, numGoroutines*numIterations)
}

func TestStatusCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	start := time.Now()
	out, err := runGZ(t, "net-env", "status")
	elapsed := time.Since(start)
	failIfUnknownCommand(t, out, err, "net-env status")
	require.NoError(t, err, "status failed:\n%s", out)
	assert.Less(t, elapsed, 10*time.Second, "status should complete quickly")
	t.Logf("status completed in %v", elapsed)
}

func TestWatchHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	isolatedHome(t)

	// Do not start a long-running watch; only verify the command exists.
	out, err := runGZ(t, "net-env", "watch", "--help")
	failIfUnknownCommand(t, out, err, "net-env watch")
	require.NoError(t, err, "watch --help failed:\n%s", out)
	assert.Contains(t, out, "interval")
}

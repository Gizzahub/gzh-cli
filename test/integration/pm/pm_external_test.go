// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build pm_external

package pm_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/test/integration/internal/testutil"
)

// TestPMExternalCLI verifies the PM command surface when built with -tags pm_external.
// Run: go test -tags pm_external ./test/integration/pm/ -count=1
func TestPMExternalCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	bin := buildPMExternalBinary(t)

	t.Run("HelpListsPM", func(t *testing.T) {
		out := mustRunBin(t, bin, "--help")
		found := false
		for _, line := range strings.Split(out, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "pm" {
				found = true
				break
			}
		}
		assert.True(t, found, "pm_external binary should list pm in help:\n%s", out)
	})

	t.Run("PMHelp", func(t *testing.T) {
		out := mustRunBin(t, bin, "pm", "--help")
		assert.Contains(t, out, "pm")
		// Subcommands from gzh-cli-package-manager (stable subset).
		for _, want := range []string{"status", "bootstrap"} {
			assert.Contains(t, out, want, "pm help should mention %s", want)
		}
	})

	t.Run("PMStatus", func(t *testing.T) {
		// status may return non-zero if managers are missing; it must still run.
		out, err := runBin(t, bin, "pm", "status")
		if err != nil {
			// Fail only if the command itself is missing.
			if strings.Contains(out, "unknown command") {
				t.Fatalf("pm status missing under pm_external:\n%s", out)
			}
			t.Logf("pm status exited non-zero (managers may be missing): %v\n%s", err, out)
			return
		}
		assert.NotEmpty(t, strings.TrimSpace(out))
	})
}

func buildPMExternalBinary(t *testing.T) string {
	t.Helper()

	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	binaryPath := testutil.ExecutablePath(tmpDir, "gz-pm")

	cmd := exec.Command("go", "build", "-tags", "pm_external", "-o", binaryPath, "./cmd/gz")
	cmd.Dir = projectRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	require.NoError(t, err, "go build -tags pm_external ./cmd/gz failed: %s", strings.TrimSpace(stderr.String()))

	return binaryPath
}

func runBin(t *testing.T, bin string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustRunBin(t *testing.T, bin string, args ...string) string {
	t.Helper()
	out, err := runBin(t, bin, args...)
	require.NoError(t, err, "gz %s failed:\n%s", strings.Join(args, " "), out)
	return out
}

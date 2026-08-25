// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package pm_test

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

// testBinary is a default-tagged gz binary (no pm_external). Built once per package.
var testBinary string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "gz-pm-it-*")
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

	// Default ship contract: no pm_external tag.
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

func runGZ(t *testing.T, args ...string) (string, error) {
	t.Helper()
	require.NotEmpty(t, testBinary, "test binary must be built in TestMain")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, testBinary, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestPMNotShipped documents the current product contract: default builds omit `pm`.
// Full PM CLI coverage lives in pm_external_test.go (//go:build pm_external).
func TestPMNotShipped(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("HelpDoesNotListPM", func(t *testing.T) {
		out, err := runGZ(t, "--help")
		require.NoError(t, err, "gz --help failed:\n%s", out)

		// Available Commands lines look like "  pm   ...". Avoid matching words in prose.
		for _, line := range strings.Split(out, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "pm ") || trimmed == "pm" {
				t.Fatalf("default binary lists pm command (ship contract broken):\n%s", line)
			}
			// cobra command table: "  pm         Package manager..."
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "pm" {
				t.Fatalf("default binary lists pm command (ship contract broken):\n%s", line)
			}
		}
	})

	t.Run("PMCommandUnknown", func(t *testing.T) {
		out, err := runGZ(t, "pm", "--help")
		require.Error(t, err, "default binary must not ship pm; got success:\n%s", out)
		assert.Contains(t, out, "unknown command")
		assert.Contains(t, out, "pm")
	})

	t.Run("PMSubcommandUnknown", func(t *testing.T) {
		out, err := runGZ(t, "pm", "status")
		require.Error(t, err, "default binary must not ship pm status; got:\n%s", out)
		assert.Contains(t, out, "unknown command")
	})
}

// TestPMBuildRequiresTag records how to enable the PM surface for optional builds.
func TestPMBuildRequiresTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Sanity: the default binary path we built is executable and reports version.
	out, err := runGZ(t, "version")
	require.NoError(t, err, "version failed:\n%s", out)
	assert.NotEmpty(t, strings.TrimSpace(out))
}

// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const toolCheckerHelperEnv = "GZH_TOOL_CHECKER_HELPER"

type toolCheckerFixture struct {
	command string
	path    string
	trace   string
	ready   string
}

// TestMain은 복사된 테스트 실행 파일만 결정적 CLI fixture로 분기한다.
func TestMain(m *testing.M) {
	if os.Getenv(toolCheckerHelperEnv) == "1" {
		os.Exit(runToolCheckerFixture())
	}
	os.Exit(m.Run())
}

func runToolCheckerFixture() int {
	tracePath := os.Getenv("GZH_TOOL_CHECKER_TRACE")
	if tracePath != "" {
		args, err := json.Marshal(os.Args[1:])
		if err != nil {
			return 90
		}
		if err := os.WriteFile(tracePath, args, 0o600); err != nil {
			return 91
		}
	}

	switch os.Getenv("GZH_TOOL_CHECKER_MODE") {
	case "version":
		_, _ = io.WriteString(os.Stdout, "  fixture v1.2.3  \n")
		return 0
	case "failure":
		_, _ = io.WriteString(os.Stderr, "fixture version command failed\n")
		return 23
	case "wait":
		if err := os.WriteFile(os.Getenv("GZH_TOOL_CHECKER_READY"), []byte("ready"), 0o600); err != nil {
			return 92
		}
		time.Sleep(10 * time.Second)
		return 0
	default:
		_, _ = io.WriteString(os.Stderr, "unexpected fixture mode\n")
		return 93
	}
}

// 테스트는 실제 도구 대신 현재 테스트 실행 파일의 복사본만 PATH에 둔다.
func installToolCheckerFixture(t *testing.T) toolCheckerFixture {
	t.Helper()

	binDir := t.TempDir()
	name := "tool-checker-fixture"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable, err := os.Executable()
	require.NoError(t, err)

	source, err := os.Open(executable)
	require.NoError(t, err)
	t.Cleanup(func() { _ = source.Close() })
	command := filepath.Join(binDir, name)
	target, err := os.OpenFile(command, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	require.NoError(t, err)
	_, copyErr := io.Copy(target, source)
	closeErr := target.Close()
	require.NoError(t, errors.Join(copyErr, closeErr))

	trace := filepath.Join(t.TempDir(), "invocation.json")
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("PATH", binDir)
	if runtime.GOOS == "windows" {
		t.Setenv("PATHEXT", ".EXE")
	}
	t.Setenv(toolCheckerHelperEnv, "1")
	t.Setenv("GZH_TOOL_CHECKER_TRACE", trace)
	t.Setenv("GZH_TOOL_CHECKER_READY", ready)

	return toolCheckerFixture{command: strings.TrimSuffix(name, ".exe"), path: command, trace: trace, ready: ready}
}

func TestToolCheckerUsesOwnedExecutableFixtures(t *testing.T) {
	t.Run("missing tool with empty owned PATH", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if runtime.GOOS == "windows" {
			t.Setenv("PATHEXT", ".EXE")
		}

		result := NewToolChecker(ToolConfig{
			Name:              "missing-fixture",
			Command:           "tool-checker-absent",
			InstallSuggestion: "install fixture",
		}).Check(context.Background())
		require.Equal(t, "missing", result.Status)
		require.Equal(t, "install fixture", result.Suggestion)
	})

	t.Run("present without version args is not executed", func(t *testing.T) {
		fixture := installToolCheckerFixture(t)
		t.Setenv("GZH_TOOL_CHECKER_MODE", "version")

		result := NewToolChecker(ToolConfig{Name: "fixture", Command: fixture.command}).Check(context.Background())
		require.Equal(t, "ok", result.Status)
		require.Equal(t, fixture.path, result.Path)
		_, err := os.Stat(fixture.trace)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("version args are forwarded and output is trimmed", func(t *testing.T) {
		fixture := installToolCheckerFixture(t)
		t.Setenv("GZH_TOOL_CHECKER_MODE", "version")
		args := []string{"--version", "--format", "short"}

		result := NewToolChecker(ToolConfig{Name: "fixture", Command: fixture.command, VersionArgs: args}).Check(context.Background())
		require.Equal(t, "ok", result.Status)
		require.Equal(t, "fixture v1.2.3", result.Version)

		data, err := os.ReadFile(fixture.trace)
		require.NoError(t, err)
		var got []string
		require.NoError(t, json.Unmarshal(data, &got))
		require.Equal(t, args, got)
	})

	t.Run("nonzero version command returns error details", func(t *testing.T) {
		fixture := installToolCheckerFixture(t)
		t.Setenv("GZH_TOOL_CHECKER_MODE", "failure")

		result := NewToolChecker(ToolConfig{Name: "fixture", Command: fixture.command, VersionArgs: []string{"version"}}).Check(context.Background())
		require.Equal(t, "error", result.Status)
		errText, ok := result.Details["error"].(string)
		require.True(t, ok)
		require.Contains(t, errText, "exit status 23")
	})

	t.Run("pre-canceled context does not start the command", func(t *testing.T) {
		fixture := installToolCheckerFixture(t)
		t.Setenv("GZH_TOOL_CHECKER_MODE", "version")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancel()

		result := NewToolChecker(ToolConfig{Name: "fixture", Command: fixture.command, VersionArgs: []string{"version"}}).Check(ctx)
		require.Equal(t, "error", result.Status)
		_, err := os.Stat(fixture.trace)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("in-flight command is canceled after readiness", func(t *testing.T) {
		fixture := installToolCheckerFixture(t)
		t.Setenv("GZH_TOOL_CHECKER_MODE", "wait")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		resultCh := make(chan DevEnvResult, 1)
		go func() {
			resultCh <- NewToolChecker(ToolConfig{Name: "fixture", Command: fixture.command, VersionArgs: []string{"wait"}}).Check(ctx)
		}()

		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			if _, err := os.Stat(fixture.ready); err == nil {
				break
			} else if !os.IsNotExist(err) {
				require.NoError(t, err)
			}
			select {
			case <-deadline.C:
				t.Fatal("fixture did not report readiness before the deadline")
			case <-ticker.C:
			}
		}

		select {
		case result := <-resultCh:
			t.Fatalf("fixture exited before cancellation: %+v", result)
		default:
		}
		cancel()
		select {
		case result := <-resultCh:
			require.Equal(t, "error", result.Status)
			require.NotEmpty(t, result.Details["error"])
		case <-time.After(time.Second):
			t.Fatal("tool check did not return after context cancellation")
		}
	})
}

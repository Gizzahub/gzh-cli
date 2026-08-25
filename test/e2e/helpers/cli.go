// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package helpers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CLIResult represents the result of a CLI command execution.
type CLIResult struct {
	ExitCode int
	Output   string
	Error    error
	Duration time.Duration
}

// CLIExecutor handles CLI command execution for E2E tests.
type CLIExecutor struct {
	BinaryPath string
	WorkDir    string
	Env        []string
	Timeout    time.Duration
}

// NewCLIExecutor creates a new CLI executor.
func NewCLIExecutor(binaryPath, workDir string) *CLIExecutor {
	return &CLIExecutor{
		BinaryPath: binaryPath,
		WorkDir:    workDir,
		Env:        os.Environ(),
		Timeout:    30 * time.Second,
	}
}

// SetEnv sets an environment variable for command execution.
func (c *CLIExecutor) SetEnv(key, value string) {
	// Remove existing env var if present
	for i, env := range c.Env {
		if strings.HasPrefix(env, key+"=") {
			c.Env = append(c.Env[:i], c.Env[i+1:]...)
			break
		}
	}
	// Add new env var
	c.Env = append(c.Env, key+"="+value)
}

// SetTimeout sets the command execution timeout.
func (c *CLIExecutor) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// Run executes a CLI command with the given arguments.
func (c *CLIExecutor) Run(args ...string) *CLIResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...) //nolint:gosec // E2E test helper with controlled binary path
	cmd.Dir = c.WorkDir
	cmd.Env = c.Env

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	// Combine stdout and stderr for complete output
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}

		output += stderr.String()
	}

	result := &CLIResult{
		Output:   output,
		Duration: duration,
	}

	if err != nil {
		result.Error = err

		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	return result
}

// RunWithInput executes a CLI command with stdin input.
func (c *CLIExecutor) RunWithInput(input string, args ...string) *CLIResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...) //nolint:gosec // E2E test helper with controlled binary path
	cmd.Dir = c.WorkDir
	cmd.Env = c.Env

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(input)

	err := cmd.Run()
	duration := time.Since(start)

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}

		output += stderr.String()
	}

	result := &CLIResult{
		Output:   output,
		Duration: duration,
	}

	if err != nil {
		result.Error = err

		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	return result
}

// RunAsync executes a command asynchronously (for daemon processes).
func (c *CLIExecutor) RunAsync(ctx context.Context, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, c.BinaryPath, args...) //nolint:gosec // E2E test helper with controlled binary path
	cmd.Dir = c.WorkDir
	cmd.Env = c.Env

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// BuildBinary builds the gz binary for testing.
func BuildBinary(ctx context.Context, projectRoot string) (string, error) {
	binaryPath := executablePath(projectRoot, "gz")

	// 만들 꾸러미를 짚어 준다. 예전에는 인자가 없어서 go build가 저장소
	// 뿌리를 만들려 했는데 거기에는 .go 파일이 하나도 없다. main은
	// cmd/gz에 있다(Makefile도 ./cmd/gz를 만든다). 그래서 e2e 시험 36개가
	// 하나도 빠짐없이 첫 걸음에서 죽었다.
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/gz") //nolint:gosec // E2E test helper building known binary
	cmd.Dir = projectRoot

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// go가 stderr에 적어 준 이유를 함께 돌려준다. 예전에는 stderr를
		// 모아 두고는 쓰지 않고 버려서, 부르는 쪽은 "exit status 1"만
		// 받았다. 위의 잘못된 꾸러미 인자가 오래 안 보인 이유가 이것이다.
		return "", fmt.Errorf("go build ./cmd/gz failed: %w: %s",
			err, strings.TrimSpace(stderr.String()))
	}

	return binaryPath, nil
}

func executablePath(dir, name string) string {
	return executablePathForGOOS(dir, name, runtime.GOOS)
}

func executablePathForGOOS(dir, name, goos string) string {
	if goos == "windows" {
		name += ".exe"
	}

	return filepath.Join(dir, name)
}

// FindProjectRoot finds the project root directory.
func FindProjectRoot() (string, error) {
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

// WaitForOutput waits for specific output from a running command.
func WaitForOutput(cmd *exec.Cmd, _ string, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		if err := cmd.Wait(); err != nil {
			done <- err
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			// Log error but don't fail the test
			_ = err
		}
		return context.DeadlineExceeded
	}
}

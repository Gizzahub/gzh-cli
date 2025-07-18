package helpers

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
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

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
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
func (c *CLIExecutor) RunAsync(args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Dir = c.WorkDir
	cmd.Env = c.Env

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// BuildBinary builds the gz binary for testing.
func BuildBinary(projectRoot string) (string, error) {
	binaryPath := filepath.Join(projectRoot, "gz")

	cmd := exec.Command("go", "build", "-o", binaryPath)
	cmd.Dir = projectRoot

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return binaryPath, nil
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
func WaitForOutput(cmd *exec.Cmd, expectedOutput string, timeout time.Duration) error {
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
		cmd.Process.Kill()
		return context.DeadlineExceeded
	}
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// AllowedGitCommands defines the whitelist of safe git commands
var AllowedGitCommands = map[string]bool{
	"clone":    true,
	"pull":     true,
	"fetch":    true,
	"reset":    true,
	"status":   true,
	"log":      true,
	"remote":   true,
	"config":   true,
	"branch":   true,
	"checkout": true,
}

// AllowedGitOptions defines safe git options
var AllowedGitOptions = map[string]bool{
	"--hard":           true,
	"--force":          true,
	"--quiet":          true,
	"--verbose":        true,
	"--progress":       true,
	"--prune":          true,
	"--all":            true,
	"--tags":           true,
	"--origin":         true,
	"--upstream":       true,
	"--set-upstream":   true,
	"--unset-upstream": true,
	"--depth":          true,
	"--shallow-depth":  true,
	"--single-branch":  true,
}

// SecureGitExecutor provides safe git command execution with input validation
type SecureGitExecutor struct {
	gitPath string
}

// NewSecureGitExecutor creates a new secure git executor
func NewSecureGitExecutor() (*SecureGitExecutor, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git command not found: %w", err)
	}

	return &SecureGitExecutor{
		gitPath: gitPath,
	}, nil
}

// ValidatedGitCommand represents a validated git command
type ValidatedGitCommand struct {
	Command  string
	Args     []string
	RepoPath string
	Options  []string
}

// ValidateCommand validates git command arguments against allowlists
func (e *SecureGitExecutor) ValidateCommand(repoPath string, args ...string) (*ValidatedGitCommand, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no git command specified")
	}

	// Validate and sanitize repository path
	cleanRepoPath, err := e.validateRepoPath(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid repository path: %w", err)
	}

	// Validate git command
	command := args[0]
	if !AllowedGitCommands[command] {
		return nil, fmt.Errorf("git command '%s' is not allowed", command)
	}

	// Validate arguments
	validArgs, validOptions, err := e.validateArgs(args[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid git arguments: %w", err)
	}

	return &ValidatedGitCommand{
		Command:  command,
		Args:     validArgs,
		RepoPath: cleanRepoPath,
		Options:  validOptions,
	}, nil
}

// Execute executes a validated git command
func (e *SecureGitExecutor) Execute(ctx context.Context, cmd *ValidatedGitCommand) error {
	// Build git arguments
	gitArgs := []string{"-C", cmd.RepoPath, cmd.Command}
	gitArgs = append(gitArgs, cmd.Options...)
	gitArgs = append(gitArgs, cmd.Args...)

	// Execute command
	execCmd := exec.CommandContext(ctx, e.gitPath, gitArgs...)

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("git %s failed in %s: %w", cmd.Command, cmd.RepoPath, err)
	}

	return nil
}

// ExecuteSecure validates and executes a git command in one call
func (e *SecureGitExecutor) ExecuteSecure(ctx context.Context, repoPath string, args ...string) error {
	validatedCmd, err := e.ValidateCommand(repoPath, args...)
	if err != nil {
		return err
	}

	return e.Execute(ctx, validatedCmd)
}

// validateRepoPath validates and cleans the repository path
func (e *SecureGitExecutor) validateRepoPath(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repository path cannot be empty")
	}

	// Clean the path to prevent path traversal
	cleanPath := filepath.Clean(repoPath)

	// Ensure path doesn't contain dangerous sequences
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected in repository path")
	}

	// Additional security checks
	if strings.HasPrefix(cleanPath, "/etc") || strings.HasPrefix(cleanPath, "/usr") {
		return "", fmt.Errorf("access to system directories not allowed")
	}

	return cleanPath, nil
}

// validateArgs validates git command arguments and separates options from args
func (e *SecureGitExecutor) validateArgs(args []string) (validArgs, validOptions []string, err error) {
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+/[a-zA-Z0-9._/-]+\.git$|^git@[a-zA-Z0-9.-]+:[a-zA-Z0-9._/-]+\.git$`)
	branchRegex := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

	for _, arg := range args {
		// Skip empty arguments
		if arg == "" {
			continue
		}

		// Check for option flags
		if strings.HasPrefix(arg, "-") {
			if AllowedGitOptions[arg] {
				validOptions = append(validOptions, arg)
			} else {
				return nil, nil, fmt.Errorf("git option '%s' is not allowed", arg)
			}
			continue
		}

		// Validate specific argument types
		if urlRegex.MatchString(arg) {
			// Valid git URL
			validArgs = append(validArgs, arg)
		} else if branchRegex.MatchString(arg) {
			// Valid branch/ref name
			validArgs = append(validArgs, arg)
		} else if isValidPathArg(arg) {
			// Valid file path
			validArgs = append(validArgs, arg)
		} else {
			return nil, nil, fmt.Errorf("argument '%s' contains invalid characters", arg)
		}
	}

	return validArgs, validOptions, nil
}

// isValidPathArg validates file path arguments
func isValidPathArg(arg string) bool {
	// Allow only safe characters in paths
	pathRegex := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	return pathRegex.MatchString(arg) && !strings.Contains(arg, "..")
}

// GetGitVersion returns the git version for diagnostics
func (e *SecureGitExecutor) GetGitVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, e.gitPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package upgrade

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// AsdfUpgrader implements PackageManagerUpgrader for asdf.
type AsdfUpgrader struct {
	logger logger.CommonLogger
}

// NewAsdfUpgrader creates a new asdf upgrader.
func NewAsdfUpgrader(logger logger.CommonLogger) *AsdfUpgrader {
	return &AsdfUpgrader{
		logger: logger,
	}
}

// CheckUpdate checks if asdf has updates available.
func (a *AsdfUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	a.logger.Debug("Checking asdf update status")

	// Get current version
	currentVersion, err := a.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current asdf version: %w", err)
	}

	// Check for updates in the git repository
	updateAvailable, err := a.checkGitUpdates(ctx)
	if err != nil {
		a.logger.Warn("Failed to check for asdf updates: %v", err)
		updateAvailable = false
	}

	return &UpgradeStatus{
		Manager:         "asdf",
		CurrentVersion:  currentVersion,
		LatestVersion:   "latest", // asdf doesn't use traditional versioning
		UpdateAvailable: updateAvailable,
		UpdateMethod:    a.GetUpdateMethod(),
		ChangelogURL:    "https://github.com/asdf-vm/asdf/releases",
	}, nil
}

// Upgrade performs the actual upgrade of asdf.
func (a *AsdfUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	a.logger.Info("Starting asdf upgrade")

	// Create backup if requested
	var backupPath string
	if options.BackupEnabled {
		var err error
		backupPath, err = a.Backup(ctx)
		if err != nil {
			a.logger.Warn("Failed to create backup: %v", err)
		} else {
			a.logger.Info("Backup created at: %s", backupPath)
		}
	}

	// Perform the upgrade
	if err := a.executeUpdate(ctx, options); err != nil {
		if backupPath != "" && !options.SkipValidation {
			a.logger.Info("Attempting rollback due to upgrade failure")
			if rollbackErr := a.Rollback(ctx, backupPath); rollbackErr != nil {
				a.logger.Error("Failed to rollback: %v", rollbackErr)
			}
		}
		return fmt.Errorf("asdf upgrade failed: %w", err)
	}

	// Validate upgrade if requested
	if !options.SkipValidation {
		if err := a.ValidateUpgrade(ctx); err != nil {
			a.logger.Warn("Upgrade validation failed: %v", err)
			return fmt.Errorf("asdf upgrade validation failed: %w", err)
		}
	}

	a.logger.Info("asdf upgrade completed successfully")
	return nil
}

// Backup creates a backup of current asdf state.
func (a *AsdfUpgrader) Backup(ctx context.Context) (string, error) {
	a.logger.Debug("Creating asdf backup")

	asdfDir := a.getAsdfDir()
	if asdfDir == "" {
		return "", fmt.Errorf("asdf directory not found")
	}

	// Backup the current git commit hash
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = asdfDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}

	backupPath := fmt.Sprintf("/tmp/asdf-backup-%d.txt", time.Now().Unix())
	commitHash := strings.TrimSpace(string(output))

	if err := writeFile(backupPath, commitHash); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	a.logger.Debug("Backup created at: %s (commit: %s)", backupPath, commitHash)
	return backupPath, nil
}

// Rollback restores from a backup.
func (a *AsdfUpgrader) Rollback(ctx context.Context, backupPath string) error {
	a.logger.Info("Rolling back asdf to previous state")

	asdfDir := a.getAsdfDir()
	if asdfDir == "" {
		return fmt.Errorf("asdf directory not found")
	}

	// Read backup commit hash
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	commitHash := strings.TrimSpace(string(content))

	// Reset to previous commit
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard", commitHash)
	cmd.Dir = asdfDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to previous commit: %w", err)
	}

	a.logger.Info("Successfully rolled back asdf to commit: %s", commitHash)
	return nil
}

// GetUpdateMethod returns the update method used.
func (a *AsdfUpgrader) GetUpdateMethod() string {
	return "asdf update"
}

// ValidateUpgrade validates that the upgrade was successful.
func (a *AsdfUpgrader) ValidateUpgrade(ctx context.Context) error {
	// Check if asdf command still works
	cmd := exec.CommandContext(ctx, "asdf", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf command validation failed: %w", err)
	}

	// Check if we can list plugins
	cmd = exec.CommandContext(ctx, "asdf", "plugin", "list")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf plugin list command validation failed: %w", err)
	}

	return nil
}

// getCurrentVersion gets the current asdf version.
func (a *AsdfUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "asdf", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute asdf version: %w", err)
	}

	// Parse version from output like "v0.10.2"
	versionRegex := regexp.MustCompile(`v(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		// If we can't parse version, return the full output
		return strings.TrimSpace(string(output)), nil
	}

	return matches[1], nil
}

// checkGitUpdates checks if there are updates available in the git repository.
func (a *AsdfUpgrader) checkGitUpdates(ctx context.Context) (bool, error) {
	asdfDir := a.getAsdfDir()
	if asdfDir == "" {
		return false, fmt.Errorf("asdf directory not found")
	}

	// Fetch latest from origin
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin")
	cmd.Dir = asdfDir
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from origin: %w", err)
	}

	// Check if local is behind remote
	cmd = exec.CommandContext(ctx, "git", "rev-list", "--count", "HEAD..origin/master")
	cmd.Dir = asdfDir
	output, err := cmd.Output()
	if err != nil {
		// Try with main branch
		cmd = exec.CommandContext(ctx, "git", "rev-list", "--count", "HEAD..origin/main")
		cmd.Dir = asdfDir
		output, err = cmd.Output()
		if err != nil {
			return false, fmt.Errorf("failed to check for updates: %w", err)
		}
	}

	commitsBehind := strings.TrimSpace(string(output))
	return commitsBehind != "0", nil
}

// executeUpdate performs the actual update commands.
func (a *AsdfUpgrader) executeUpdate(ctx context.Context, options UpgradeOptions) error {
	// Set timeout if specified
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Use asdf's built-in update command
	a.logger.Info("Updating asdf...")
	cmd := exec.CommandContext(ctx, "asdf", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf update failed: %w", err)
	}

	return nil
}

// getAsdfDir returns the asdf installation directory.
func (a *AsdfUpgrader) getAsdfDir() string {
	// Check common asdf installation locations
	possiblePaths := []string{
		os.Getenv("ASDF_DIR"),
		filepath.Join(os.Getenv("HOME"), ".asdf"),
		"/opt/asdf-vm",
		"/usr/local/opt/asdf",
	}

	for _, path := range possiblePaths {
		if path != "" {
			if info, err := os.Stat(filepath.Join(path, ".git")); err == nil && info.IsDir() {
				return path
			}
		}
	}

	return ""
}

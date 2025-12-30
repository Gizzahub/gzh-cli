// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

// RbenvGemSynchronizer handles synchronization between rbenv and gem.
type RbenvGemSynchronizer struct {
	logger logger.CommonLogger
}

// NewRbenvGemSynchronizer creates a new rbenv-gem synchronizer.
func NewRbenvGemSynchronizer(logger logger.CommonLogger) *RbenvGemSynchronizer {
	return &RbenvGemSynchronizer{
		logger: logger,
	}
}

// GetManagerPair returns the manager pair names.
func (rgs *RbenvGemSynchronizer) GetManagerPair() (string, string) {
	return "rbenv", "gem"
}

// CheckSync checks the synchronization status between rbenv and gem.
func (rgs *RbenvGemSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
	rgs.logger.Debug("Checking rbenv-gem synchronization status")

	// Check if rbenv is available
	if !rgs.isRbenvAvailable(ctx) {
		return &VersionSyncStatus{
			VersionManager:    "rbenv",
			PackageManager:    "gem",
			VMVersion:         "not_installed",
			PMVersion:         statusUnknown,
			ExpectedPMVersion: statusUnknown,
			InSync:            false,
			SyncAction:        "install_rbenv",
			Issues:            []string{"rbenv is not installed or not in PATH"},
		}, nil
	}

	// Get current Ruby version from rbenv
	rubyVersion, err := rgs.getCurrentRubyVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Ruby version: %w", err)
	}

	// Get current gem version
	gemVersion, err := rgs.getCurrentGemVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current gem version: %w", err)
	}

	// Get expected gem version for the current Ruby version
	expectedGemVersion, err := rgs.getExpectedGemVersion(ctx, rubyVersion)
	if err != nil {
		rgs.logger.Warn("Failed to get expected gem version for Ruby %s: %v", rubyVersion, err)
		expectedGemVersion = statusUnknown
	}

	// Compare versions
	inSync := rgs.compareVersions(gemVersion, expectedGemVersion)
	syncAction := rgs.determineSyncAction(gemVersion, expectedGemVersion, inSync)

	return &VersionSyncStatus{
		VersionManager:    "rbenv",
		PackageManager:    "gem",
		VMVersion:         rubyVersion,
		PMVersion:         gemVersion,
		ExpectedPMVersion: expectedGemVersion,
		InSync:            inSync,
		SyncAction:        syncAction,
	}, nil
}

// Synchronize performs synchronization between rbenv and gem.
func (rgs *RbenvGemSynchronizer) Synchronize(ctx context.Context, policy SyncPolicy) error {
	rgs.logger.Info("Starting rbenv-gem synchronization")

	status, err := rgs.CheckSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if status.InSync {
		rgs.logger.Info("rbenv and gem are already synchronized")
		return nil
	}

	switch policy.Strategy {
	case "vm_priority":
		return rgs.syncToRubyVersion(ctx, status.VMVersion, policy)
	case "pm_priority":
		return rgs.syncToGemVersion(ctx, status.PMVersion, policy)
	case "latest":
		return rgs.upgradeToLatest(ctx, policy)
	default:
		return fmt.Errorf("unknown synchronization strategy: %s", policy.Strategy)
	}
}

// GetExpectedVersion returns the expected gem version for a given Ruby version.
func (rgs *RbenvGemSynchronizer) GetExpectedVersion(ctx context.Context, vmVersion string) (string, error) {
	return rgs.getExpectedGemVersion(ctx, vmVersion)
}

// ValidateSync validates the synchronization status.
func (rgs *RbenvGemSynchronizer) ValidateSync(ctx context.Context) error {
	status, err := rgs.CheckSync(ctx)
	if err != nil {
		return err
	}

	if !status.InSync {
		return fmt.Errorf("rbenv and gem are out of sync: Ruby %s, gem %s (expected %s)",
			status.VMVersion, status.PMVersion, status.ExpectedPMVersion)
	}

	return nil
}

// isRbenvAvailable checks if rbenv is available in the current environment.
func (rgs *RbenvGemSynchronizer) isRbenvAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "which", "rbenv")
	err := cmd.Run()
	return err == nil
}

// getCurrentRubyVersion gets the current Ruby version from rbenv.
func (rgs *RbenvGemSynchronizer) getCurrentRubyVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "rbenv", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current Ruby version: %w", err)
	}

	// Parse output like "3.1.0 (set by /home/user/.rbenv/version)"
	versionLine := strings.TrimSpace(string(output))
	parts := strings.Fields(versionLine)
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected rbenv version output: %s", versionLine)
	}

	return parts[0], nil
}

// getCurrentGemVersion gets the current gem version.
func (rgs *RbenvGemSynchronizer) getCurrentGemVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gem", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get gem version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// getExpectedGemVersion gets the expected gem version for a given Ruby version.
func (rgs *RbenvGemSynchronizer) getExpectedGemVersion(ctx context.Context, rubyVersion string) (string, error) {
	// Ruby version to bundled gem version mapping
	rubyGemMap := map[string]string{
		"3.2.0":  "3.4.1",
		"3.1.0":  "3.3.7",
		"3.0.0":  "3.2.3",
		"2.7.6":  "3.1.6",
		"2.7.0":  "3.1.2",
		"2.6.10": "3.0.3",
	}

	if expectedVersion, exists := rubyGemMap[rubyVersion]; exists {
		return expectedVersion, nil
	}

	// Try to get gem version bundled with Ruby installation
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("RBENV_VERSION=%s gem --version", rubyVersion))
	output, err := cmd.Output()
	if err != nil {
		rgs.logger.Debug("Failed to get bundled gem version for Ruby %s: %v", rubyVersion, err)
		return statusUnknown, nil
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// compareVersions compares two version strings.
func (rgs *RbenvGemSynchronizer) compareVersions(current, expected string) bool {
	if expected == statusUnknown || current == statusUnknown {
		return false
	}

	// Simple version comparison - in practice, you'd use a proper semver library
	return current == expected
}

// determineSyncAction determines what sync action is needed.
func (rgs *RbenvGemSynchronizer) determineSyncAction(current, expected string, inSync bool) string {
	if inSync {
		return "none"
	}

	if expected == statusUnknown {
		return "check_compatibility"
	}

	return fmt.Sprintf("update gem to %s", expected)
}

// syncToRubyVersion synchronizes gem to match the current Ruby version.
func (rgs *RbenvGemSynchronizer) syncToRubyVersion(ctx context.Context, rubyVersion string, policy SyncPolicy) error {
	rgs.logger.Info("Synchronizing gem to match Ruby version %s", rubyVersion)

	expectedGemVersion, err := rgs.getExpectedGemVersion(ctx, rubyVersion)
	if err != nil || expectedGemVersion == statusUnknown {
		return fmt.Errorf("cannot determine expected gem version for Ruby %s", rubyVersion)
	}

	// Install the specific gem version
	cmd := exec.CommandContext(ctx, "gem", "update", "--system", expectedGemVersion)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update gem to %s: %w", expectedGemVersion, err)
	}

	rgs.logger.Info("Successfully synchronized gem to version %s", expectedGemVersion)
	return nil
}

// syncToGemVersion synchronizes Ruby to match the current gem version.
func (rgs *RbenvGemSynchronizer) syncToGemVersion(ctx context.Context, gemVersion string, policy SyncPolicy) error {
	rgs.logger.Info("Synchronizing Ruby to match gem version %s", gemVersion)

	// This is complex as gem versions can work with multiple Ruby versions
	// For now, we'll suggest using the latest stable Ruby
	cmd := exec.CommandContext(ctx, "bash", "-c", "rbenv install $(rbenv install -l | grep -v - | tail -1) && rbenv global $(rbenv install -l | grep -v - | tail -1)")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install/use latest stable Ruby: %w", err)
	}

	rgs.logger.Info("Synchronized to latest stable Ruby version")
	return nil
}

// upgradeToLatest upgrades both Ruby and gem to their latest versions.
func (rgs *RbenvGemSynchronizer) upgradeToLatest(ctx context.Context, policy SyncPolicy) error {
	rgs.logger.Info("Upgrading both Ruby and gem to latest versions")

	// Install latest stable Ruby
	cmd := exec.CommandContext(ctx, "bash", "-c", "rbenv install $(rbenv install -l | grep -v - | tail -1) && rbenv global $(rbenv install -l | grep -v - | tail -1)")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install latest stable Ruby: %w", err)
	}

	// Update gem to latest
	cmd = exec.CommandContext(ctx, "gem", "update", "--system")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update gem to latest: %w", err)
	}

	rgs.logger.Info("Successfully upgraded to latest Ruby and gem versions")
	return nil
}

package sync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// 동기화 전략 상수들
const (
	syncStrategyVMPriority = "vm_priority"
	syncStrategyPMPriority = "pm_priority"
	syncActionCheckCompat  = "check_compatibility"
)

// NvmNpmSynchronizer handles synchronization between nvm (Node Version Manager) and npm
type NvmNpmSynchronizer struct {
	logger logger.CommonLogger
}

// NewNvmNpmSynchronizer creates a new nvm-npm synchronizer
func NewNvmNpmSynchronizer(logger logger.CommonLogger) *NvmNpmSynchronizer {
	return &NvmNpmSynchronizer{
		logger: logger,
	}
}

// GetManagerPair returns the manager pair names
func (nns *NvmNpmSynchronizer) GetManagerPair() (string, string) {
	return "nvm", "npm"
}

// CheckSync checks the synchronization status between nvm and npm
func (nns *NvmNpmSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
	nns.logger.Debug("Checking nvm-npm synchronization status")

	// Check if nvm is available
	if !nns.isNvmAvailable(ctx) {
		return &VersionSyncStatus{
			VersionManager:    "nvm",
			PackageManager:    "npm",
			VMVersion:         "not_installed",
			PMVersion:         "unknown",
			ExpectedPMVersion: "unknown",
			InSync:            false,
			SyncAction:        "install_nvm",
			Issues:            []string{"nvm is not installed or not in PATH"},
		}, nil
	}

	// Get current Node.js version from nvm
	nodeVersion, err := nns.getCurrentNodeVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Node.js version: %w", err)
	}

	// Get current npm version
	npmVersion, err := nns.getCurrentNpmVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current npm version: %w", err)
	}

	// Get expected npm version for the current Node.js version
	expectedNpmVersion, err := nns.getExpectedNpmVersion(ctx, nodeVersion)
	if err != nil {
		nns.logger.Warn("Failed to get expected npm version for Node.js %s: %v", nodeVersion, err)
		expectedNpmVersion = "unknown"
	}

	// Compare versions
	inSync := nns.compareVersions(npmVersion, expectedNpmVersion)
	syncAction := nns.determineSyncAction(npmVersion, expectedNpmVersion, inSync)

	return &VersionSyncStatus{
		VersionManager:    "nvm",
		PackageManager:    "npm",
		VMVersion:         nodeVersion,
		PMVersion:         npmVersion,
		ExpectedPMVersion: expectedNpmVersion,
		InSync:            inSync,
		SyncAction:        syncAction,
	}, nil
}

// Synchronize performs synchronization between nvm and npm
func (nns *NvmNpmSynchronizer) Synchronize(ctx context.Context, policy SyncPolicy) error {
	nns.logger.Info("Starting nvm-npm synchronization")

	status, err := nns.CheckSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if status.InSync {
		nns.logger.Info("nvm and npm are already synchronized")
		return nil
	}

	switch policy.Strategy {
	case syncStrategyVMPriority:
		return nns.syncToNodeVersion(ctx, status.VMVersion, policy)
	case syncStrategyPMPriority:
		return nns.syncToNpmVersion(ctx, status.PMVersion, policy)
	case "latest":
		return nns.upgradeToLatest(ctx, policy)
	default:
		return fmt.Errorf("unknown synchronization strategy: %s", policy.Strategy)
	}
}

// GetExpectedVersion returns the expected npm version for a given Node.js version
func (nns *NvmNpmSynchronizer) GetExpectedVersion(ctx context.Context, vmVersion string) (string, error) {
	return nns.getExpectedNpmVersion(ctx, vmVersion)
}

// ValidateSync validates the synchronization status
func (nns *NvmNpmSynchronizer) ValidateSync(ctx context.Context) error {
	status, err := nns.CheckSync(ctx)
	if err != nil {
		return err
	}

	if !status.InSync {
		return fmt.Errorf("nvm and npm are out of sync: Node.js %s, npm %s (expected %s)",
			status.VMVersion, status.PMVersion, status.ExpectedPMVersion)
	}

	return nil
}

// isNvmAvailable checks if nvm is available in the current environment
func (nns *NvmNpmSynchronizer) isNvmAvailable(ctx context.Context) bool {
	// Check if nvm command is available (nvm is usually a shell function)
	cmd := exec.CommandContext(ctx, "bash", "-c", "command -v nvm")
	err := cmd.Run()
	return err == nil
}

// getCurrentNodeVersion gets the current Node.js version from nvm
func (nns *NvmNpmSynchronizer) getCurrentNodeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", "nvm current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current Node.js version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	// Remove 'v' prefix if present
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	return version, nil
}

// getCurrentNpmVersion gets the current npm version
func (nns *NvmNpmSynchronizer) getCurrentNpmVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get npm version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// getExpectedNpmVersion gets the expected npm version for a given Node.js version
func (nns *NvmNpmSynchronizer) getExpectedNpmVersion(ctx context.Context, nodeVersion string) (string, error) {
	// This is a simplified implementation. In practice, you'd want to maintain
	// a compatibility matrix or query from a reliable source
	nodeVersionMap := map[string]string{
		"18.17.0": "9.6.7",
		"18.16.1": "9.5.1",
		"18.16.0": "9.5.1",
		"16.20.1": "8.19.4",
		"16.20.0": "8.19.4",
		"14.21.3": "6.14.18",
		"14.21.2": "6.14.17",
	}

	if expectedVersion, exists := nodeVersionMap[nodeVersion]; exists {
		return expectedVersion, nil
	}

	// Try to get npm version bundled with Node.js
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("nvm use %s && npm --version", nodeVersion)) // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		nns.logger.Debug("Failed to get bundled npm version for Node.js %s: %v", nodeVersion, err)
		return "unknown", nil
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// compareVersions compares two version strings
func (nns *NvmNpmSynchronizer) compareVersions(current, expected string) bool {
	if expected == "unknown" || current == "unknown" {
		return false
	}

	// Simple version comparison - in practice, you'd use a proper semver library
	return current == expected
}

// determineSyncAction determines what sync action is needed
func (nns *NvmNpmSynchronizer) determineSyncAction(_, expected string, inSync bool) string {
	if inSync {
		return "none"
	}

	if expected == "unknown" {
		return syncActionCheckCompat
	}

	// Simple comparison - in practice, you'd use proper semver comparison
	return fmt.Sprintf("update npm to %s", expected)
}

// syncToNodeVersion synchronizes npm to match the current Node.js version
func (nns *NvmNpmSynchronizer) syncToNodeVersion(ctx context.Context, nodeVersion string, _ SyncPolicy) error {
	nns.logger.Info("Synchronizing npm to match Node.js version %s", nodeVersion)

	expectedNpmVersion, err := nns.getExpectedNpmVersion(ctx, nodeVersion)
	if err != nil || expectedNpmVersion == "unknown" {
		return fmt.Errorf("cannot determine expected npm version for Node.js %s", nodeVersion)
	}

	// Install the specific npm version
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", fmt.Sprintf("npm@%s", expectedNpmVersion)) // #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install npm@%s: %w", expectedNpmVersion, err)
	}

	nns.logger.Info("Successfully synchronized npm to version %s", expectedNpmVersion)
	return nil
}

// syncToNpmVersion synchronizes Node.js to match the current npm version
func (nns *NvmNpmSynchronizer) syncToNpmVersion(ctx context.Context, npmVersion string, policy SyncPolicy) error {
	nns.logger.Info("Synchronizing Node.js to match npm version %s", npmVersion)

	// This is complex as npm versions can work with multiple Node.js versions
	// For now, we'll suggest using the latest LTS Node.js
	cmd := exec.CommandContext(ctx, "bash", "-c", "nvm install --lts && nvm use --lts")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install/use LTS Node.js: %w", err)
	}

	nns.logger.Info("Synchronized to LTS Node.js version")
	return nil
}

// upgradeToLatest upgrades both Node.js and npm to their latest versions
func (nns *NvmNpmSynchronizer) upgradeToLatest(ctx context.Context, policy SyncPolicy) error {
	nns.logger.Info("Upgrading both Node.js and npm to latest versions")

	// Install latest Node.js LTS
	cmd := exec.CommandContext(ctx, "bash", "-c", "nvm install --lts && nvm use --lts")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install latest LTS Node.js: %w", err)
	}

	// Install latest npm
	cmd = exec.CommandContext(ctx, "npm", "install", "-g", "npm@latest")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install latest npm: %w", err)
	}

	nns.logger.Info("Successfully upgraded to latest Node.js and npm versions")
	return nil
}

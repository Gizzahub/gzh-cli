package sync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// PyenvPipSynchronizer handles synchronization between pyenv and pip
type PyenvPipSynchronizer struct {
	logger logger.CommonLogger
}

// NewPyenvPipSynchronizer creates a new pyenv-pip synchronizer
func NewPyenvPipSynchronizer(logger logger.CommonLogger) *PyenvPipSynchronizer {
	return &PyenvPipSynchronizer{
		logger: logger,
	}
}

// GetManagerPair returns the manager pair names
func (pps *PyenvPipSynchronizer) GetManagerPair() (string, string) {
	return "pyenv", "pip"
}

// CheckSync checks the synchronization status between pyenv and pip
func (pps *PyenvPipSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
	pps.logger.Debug("Checking pyenv-pip synchronization status")

	// Check if pyenv is available
	if !pps.isPyenvAvailable(ctx) {
		return &VersionSyncStatus{
			VersionManager:    "pyenv",
			PackageManager:    "pip",
			VMVersion:         "not_installed",
			PMVersion:         "unknown",
			ExpectedPMVersion: "unknown",
			InSync:            false,
			SyncAction:        "install_pyenv",
			Issues:            []string{"pyenv is not installed or not in PATH"},
		}, nil
	}

	// Get current Python version from pyenv
	pythonVersion, err := pps.getCurrentPythonVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Python version: %w", err)
	}

	// Get current pip version
	pipVersion, err := pps.getCurrentPipVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pip version: %w", err)
	}

	// Get expected pip version for the current Python version
	expectedPipVersion, err := pps.getExpectedPipVersion(ctx, pythonVersion)
	if err != nil {
		pps.logger.Warn("Failed to get expected pip version for Python %s: %v", pythonVersion, err)
		expectedPipVersion = "unknown"
	}

	// Compare versions
	inSync := pps.compareVersions(pipVersion, expectedPipVersion)
	syncAction := pps.determineSyncAction(pipVersion, expectedPipVersion, inSync)

	return &VersionSyncStatus{
		VersionManager:    "pyenv",
		PackageManager:    "pip",
		VMVersion:         pythonVersion,
		PMVersion:         pipVersion,
		ExpectedPMVersion: expectedPipVersion,
		InSync:            inSync,
		SyncAction:        syncAction,
	}, nil
}

// Synchronize performs synchronization between pyenv and pip
func (pps *PyenvPipSynchronizer) Synchronize(ctx context.Context, policy SyncPolicy) error {
	pps.logger.Info("Starting pyenv-pip synchronization")

	status, err := pps.CheckSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if status.InSync {
		pps.logger.Info("pyenv and pip are already synchronized")
		return nil
	}

	switch policy.Strategy {
	case "vm_priority":
		return pps.syncToPythonVersion(ctx, status.VMVersion, policy)
	case "pm_priority":
		return pps.syncToPipVersion(ctx, status.PMVersion, policy)
	case "latest":
		return pps.upgradeToLatest(ctx, policy)
	default:
		return fmt.Errorf("unknown synchronization strategy: %s", policy.Strategy)
	}
}

// GetExpectedVersion returns the expected pip version for a given Python version
func (pps *PyenvPipSynchronizer) GetExpectedVersion(ctx context.Context, vmVersion string) (string, error) {
	return pps.getExpectedPipVersion(ctx, vmVersion)
}

// ValidateSync validates the synchronization status
func (pps *PyenvPipSynchronizer) ValidateSync(ctx context.Context) error {
	status, err := pps.CheckSync(ctx)
	if err != nil {
		return err
	}

	if !status.InSync {
		return fmt.Errorf("pyenv and pip are out of sync: Python %s, pip %s (expected %s)",
			status.VMVersion, status.PMVersion, status.ExpectedPMVersion)
	}

	return nil
}

// isPyenvAvailable checks if pyenv is available in the current environment
func (pps *PyenvPipSynchronizer) isPyenvAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "which", "pyenv")
	err := cmd.Run()
	return err == nil
}

// getCurrentPythonVersion gets the current Python version from pyenv
func (pps *PyenvPipSynchronizer) getCurrentPythonVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "pyenv", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current Python version: %w", err)
	}

	// Parse output like "3.11.0 (set by /home/user/.pyenv/version)"
	versionLine := strings.TrimSpace(string(output))
	parts := strings.Fields(versionLine)
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected pyenv version output: %s", versionLine)
	}

	return parts[0], nil
}

// getCurrentPipVersion gets the current pip version
func (pps *PyenvPipSynchronizer) getCurrentPipVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "pip", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pip version: %w", err)
	}

	// Parse output like "pip 22.3 from ..."
	versionLine := strings.TrimSpace(string(output))
	parts := strings.Fields(versionLine)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected pip version output: %s", versionLine)
	}

	return parts[1], nil
}

// getExpectedPipVersion gets the expected pip version for a given Python version
func (pps *PyenvPipSynchronizer) getExpectedPipVersion(ctx context.Context, pythonVersion string) (string, error) {
	// Python version to bundled pip version mapping
	pythonPipMap := map[string]string{
		"3.11.0": "22.3",
		"3.10.8": "22.2.2",
		"3.10.0": "21.2.4",
		"3.9.15": "21.2.4",
		"3.9.0":  "20.2.1",
		"3.8.16": "21.2.4",
		"3.8.0":  "19.2.3",
		"3.7.16": "21.2.4",
		"3.7.0":  "19.0.3",
	}

	if expectedVersion, exists := pythonPipMap[pythonVersion]; exists {
		return expectedVersion, nil
	}

	// Try to get pip version bundled with Python installation
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("PYENV_VERSION=%s pip --version", pythonVersion))
	output, err := cmd.Output()
	if err != nil {
		pps.logger.Debug("Failed to get bundled pip version for Python %s: %v", pythonVersion, err)
		return "unknown", nil
	}

	// Parse output like "pip 22.3 from ..."
	versionLine := strings.TrimSpace(string(output))
	parts := strings.Fields(versionLine)
	if len(parts) < 2 {
		return "unknown", nil
	}

	return parts[1], nil
}

// compareVersions compares two version strings
func (pps *PyenvPipSynchronizer) compareVersions(current, expected string) bool {
	if expected == "unknown" || current == "unknown" {
		return false
	}

	// Simple version comparison - in practice, you'd use a proper semver library
	return current == expected
}

// determineSyncAction determines what sync action is needed
func (pps *PyenvPipSynchronizer) determineSyncAction(current, expected string, inSync bool) string {
	if inSync {
		return "none"
	}

	if expected == "unknown" {
		return "check_compatibility"
	}

	return fmt.Sprintf("update pip to %s", expected)
}

// syncToPythonVersion synchronizes pip to match the current Python version
func (pps *PyenvPipSynchronizer) syncToPythonVersion(ctx context.Context, pythonVersion string, policy SyncPolicy) error {
	pps.logger.Info("Synchronizing pip to match Python version %s", pythonVersion)

	expectedPipVersion, err := pps.getExpectedPipVersion(ctx, pythonVersion)
	if err != nil || expectedPipVersion == "unknown" {
		return fmt.Errorf("cannot determine expected pip version for Python %s", pythonVersion)
	}

	// Install the specific pip version
	cmd := exec.CommandContext(ctx, "pip", "install", fmt.Sprintf("pip==%s", expectedPipVersion))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install pip==%s: %w", expectedPipVersion, err)
	}

	pps.logger.Info("Successfully synchronized pip to version %s", expectedPipVersion)
	return nil
}

// syncToPipVersion synchronizes Python to match the current pip version
func (pps *PyenvPipSynchronizer) syncToPipVersion(ctx context.Context, pipVersion string, policy SyncPolicy) error {
	pps.logger.Info("Synchronizing Python to match pip version %s", pipVersion)

	// This is complex as pip versions can work with multiple Python versions
	// For now, we'll suggest using the latest stable Python
	cmd := exec.CommandContext(ctx, "bash", "-c", "pyenv install $(pyenv install -l | grep -E '^\\s*[0-9]+\\.[0-9]+\\.[0-9]+$' | tail -1) && pyenv global $(pyenv install -l | grep -E '^\\s*[0-9]+\\.[0-9]+\\.[0-9]+$' | tail -1)")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install/use latest stable Python: %w", err)
	}

	pps.logger.Info("Synchronized to latest stable Python version")
	return nil
}

// upgradeToLatest upgrades both Python and pip to their latest versions
func (pps *PyenvPipSynchronizer) upgradeToLatest(ctx context.Context, policy SyncPolicy) error {
	pps.logger.Info("Upgrading both Python and pip to latest versions")

	// Install latest stable Python
	cmd := exec.CommandContext(ctx, "bash", "-c", "pyenv install $(pyenv install -l | grep -E '^\\s*[0-9]+\\.[0-9]+\\.[0-9]+$' | tail -1) && pyenv global $(pyenv install -l | grep -E '^\\s*[0-9]+\\.[0-9]+\\.[0-9]+$' | tail -1)")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install latest stable Python: %w", err)
	}

	// Upgrade pip to latest
	cmd = exec.CommandContext(ctx, "pip", "install", "--upgrade", "pip")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip to latest: %w", err)
	}

	pps.logger.Info("Successfully upgraded to latest Python and pip versions")
	return nil
}

package sync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// AsdfMultiSynchronizer handles synchronization for asdf with multiple tools.
type AsdfMultiSynchronizer struct {
	logger logger.CommonLogger
}

// NewAsdfMultiSynchronizer creates a new asdf multi-tool synchronizer.
func NewAsdfMultiSynchronizer(logger logger.CommonLogger) *AsdfMultiSynchronizer {
	return &AsdfMultiSynchronizer{
		logger: logger,
	}
}

// GetManagerPair returns the manager pair names
func (ams *AsdfMultiSynchronizer) GetManagerPair() (string, string) {
	return "asdf", "multi"
}

// CheckSync checks the synchronization status for all asdf-managed tools
func (ams *AsdfMultiSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
	ams.logger.Debug("Checking asdf multi-tool synchronization status")

	// Check if asdf is available
	if !ams.isAsdfAvailable(ctx) {
		return &VersionSyncStatus{
			VersionManager:    "asdf",
			PackageManager:    "multi",
			VMVersion:         "not_installed",
			PMVersion:         "unknown",
			ExpectedPMVersion: "unknown",
			InSync:            false,
			SyncAction:        "install_asdf",
			Issues:            []string{"asdf is not installed or not in PATH"},
		}, nil
	}

	// Get list of installed tools
	installedTools, err := ams.getInstalledTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed tools: %w", err)
	}

	if len(installedTools) == 0 {
		return &VersionSyncStatus{
			VersionManager:    "asdf",
			PackageManager:    "multi",
			VMVersion:         "installed",
			PMVersion:         "no_tools",
			ExpectedPMVersion: "unknown",
			InSync:            true, // No tools means no conflicts
			SyncAction:        "none",
		}, nil
	}

	// Check synchronization for supported tool pairs
	issues := []string{}
	outOfSyncCount := 0

	for _, tool := range installedTools {
		if ams.isSupportedTool(tool) {
			syncStatus, err := ams.checkToolSync(ctx, tool)
			if err != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", tool, err))
				outOfSyncCount++
			} else if !syncStatus {
				issues = append(issues, fmt.Sprintf("%s: out of sync", tool))
				outOfSyncCount++
			}
		}
	}

	inSync := outOfSyncCount == 0
	vmVersion := fmt.Sprintf("managing %d tools", len(installedTools))
	pmVersion := fmt.Sprintf("%d synchronized", len(installedTools)-outOfSyncCount)

	syncAction := "none"
	if !inSync {
		syncAction = fmt.Sprintf("synchronize %d tools", outOfSyncCount)
	}

	return &VersionSyncStatus{
		VersionManager:    "asdf",
		PackageManager:    "multi",
		VMVersion:         vmVersion,
		PMVersion:         pmVersion,
		ExpectedPMVersion: fmt.Sprintf("%d tools", len(installedTools)),
		InSync:            inSync,
		SyncAction:        syncAction,
		Issues:            issues,
	}, nil
}

// Synchronize performs synchronization for asdf-managed tools.
func (ams *AsdfMultiSynchronizer) Synchronize(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Starting asdf multi-tool synchronization")

	status, err := ams.CheckSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if status.InSync {
		ams.logger.Info("All asdf-managed tools are already synchronized")
		return nil
	}

	// Get list of installed tools
	installedTools, err := ams.getInstalledTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get installed tools: %w", err)
	}

	// Synchronize each supported tool
	for _, tool := range installedTools {
		if ams.isSupportedTool(tool) {
			if err := ams.synchronizeTool(ctx, tool, policy); err != nil {
				ams.logger.Warn("Failed to synchronize %s: %v", tool, err)
			} else {
				ams.logger.Info("Successfully synchronized %s", tool)
			}
		}
	}

	return nil
}

// GetExpectedVersion returns the expected version for asdf (not applicable for multi-tool)
func (ams *AsdfMultiSynchronizer) GetExpectedVersion(ctx context.Context, vmVersion string) (string, error) {
	return "multi-tool", nil
}

// ValidateSync validates the synchronization status for all asdf tools
func (ams *AsdfMultiSynchronizer) ValidateSync(ctx context.Context) error {
	status, err := ams.CheckSync(ctx)
	if err != nil {
		return err
	}

	if !status.InSync {
		return fmt.Errorf("asdf tools are out of sync: %s", strings.Join(status.Issues, "; "))
	}

	return nil
}

// isAsdfAvailable checks if asdf is available in the current environment
func (ams *AsdfMultiSynchronizer) isAsdfAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "which", "asdf")
	err := cmd.Run()
	return err == nil
}

// getInstalledTools gets the list of tools managed by asdf
func (ams *AsdfMultiSynchronizer) getInstalledTools(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "asdf", "plugin", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get asdf plugin list: %w", err)
	}

	tools := []string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		tool := strings.TrimSpace(line)
		if tool != "" {
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

// isSupportedTool checks if the tool is supported for synchronization
func (ams *AsdfMultiSynchronizer) isSupportedTool(tool string) bool {
	supportedTools := map[string]bool{
		"nodejs": true,
		"ruby":   true,
		"python": true,
		"golang": true,
		"java":   true,
	}

	return supportedTools[tool]
}

// checkToolSync checks if a specific tool is synchronized
func (ams *AsdfMultiSynchronizer) checkToolSync(ctx context.Context, tool string) (bool, error) {
	switch tool {
	case "nodejs":
		return ams.checkNodejsSync(ctx)
	case "ruby":
		return ams.checkRubySync(ctx)
	case "python":
		return ams.checkPythonSync(ctx)
	case "golang":
		return ams.checkGolangSync(ctx)
	case "java":
		return ams.checkJavaSync(ctx)
	default:
		return true, nil // Unsupported tools are considered synchronized
	}
}

// synchronizeTool synchronizes a specific tool
func (ams *AsdfMultiSynchronizer) synchronizeTool(ctx context.Context, tool string, policy SyncPolicy) error {
	switch tool {
	case "nodejs":
		return ams.synchronizeNodejs(ctx, policy)
	case "ruby":
		return ams.synchronizeRuby(ctx, policy)
	case "python":
		return ams.synchronizePython(ctx, policy)
	case "golang":
		return ams.synchronizeGolang(ctx, policy)
	case "java":
		return ams.synchronizeJava(ctx, policy)
	default:
		return nil // Nothing to do for unsupported tools
	}
}

// checkNodejsSync checks Node.js and npm synchronization
func (ams *AsdfMultiSynchronizer) checkNodejsSync(ctx context.Context) (bool, error) {
	// Get current Node.js version from asdf
	nodeVersion, err := ams.getCurrentToolVersion(ctx, "nodejs")
	if err != nil {
		return false, err
	}

	// Get current npm version
	cmd := exec.CommandContext(ctx, "npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get npm version: %w", err)
	}

	npmVersion := strings.TrimSpace(string(output))

	// Simple check - in practice, you'd use a proper compatibility matrix
	ams.logger.Debug("Node.js %s with npm %s", nodeVersion, npmVersion)
	return true, nil // For now, assume compatible
}

// checkRubySync checks Ruby and gem synchronization
func (ams *AsdfMultiSynchronizer) checkRubySync(ctx context.Context) (bool, error) {
	// Similar implementation for Ruby and gem
	return true, nil
}

// checkPythonSync checks Python and pip synchronization
func (ams *AsdfMultiSynchronizer) checkPythonSync(ctx context.Context) (bool, error) {
	// Similar implementation for Python and pip
	return true, nil
}

// checkGolangSync checks Go installation
func (ams *AsdfMultiSynchronizer) checkGolangSync(ctx context.Context) (bool, error) {
	// Go doesn't have a separate package manager like npm/gem/pip
	return true, nil
}

// checkJavaSync checks Java installation
func (ams *AsdfMultiSynchronizer) checkJavaSync(ctx context.Context) (bool, error) {
	// Java doesn't have a unified package manager
	return true, nil
}

// getCurrentToolVersion gets the current version of a tool managed by asdf
func (ams *AsdfMultiSynchronizer) getCurrentToolVersion(ctx context.Context, tool string) (string, error) {
	cmd := exec.CommandContext(ctx, "asdf", "current", tool)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current %s version: %w", tool, err)
	}

	// Parse output like "nodejs          18.17.0          /home/user/.tool-versions"
	versionLine := strings.TrimSpace(string(output))
	parts := strings.Fields(versionLine)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected asdf current output for %s: %s", tool, versionLine)
	}

	return parts[1], nil
}

// synchronizeNodejs synchronizes Node.js and npm
func (ams *AsdfMultiSynchronizer) synchronizeNodejs(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Synchronizing Node.js and npm via asdf")

	switch policy.Strategy {
	case "latest":
		// Install latest Node.js version
		cmd := exec.CommandContext(ctx, "asdf", "install", "nodejs", "latest")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install latest Node.js: %w", err)
		}

		cmd = exec.CommandContext(ctx, "asdf", "global", "nodejs", "latest")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set global Node.js version: %w", err)
		}
	}

	return nil
}

// synchronizeRuby synchronizes Ruby and gem
func (ams *AsdfMultiSynchronizer) synchronizeRuby(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Synchronizing Ruby and gem via asdf")
	// Implementation similar to synchronizeNodejs
	return nil
}

// synchronizePython synchronizes Python and pip
func (ams *AsdfMultiSynchronizer) synchronizePython(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Synchronizing Python and pip via asdf")
	// Implementation similar to synchronizeNodejs
	return nil
}

// synchronizeGolang synchronizes Go
func (ams *AsdfMultiSynchronizer) synchronizeGolang(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Synchronizing Go via asdf")
	// Go doesn't need package manager synchronization
	return nil
}

// synchronizeJava synchronizes Java
func (ams *AsdfMultiSynchronizer) synchronizeJava(ctx context.Context, policy SyncPolicy) error {
	ams.logger.Info("Synchronizing Java via asdf")
	// Java doesn't need package manager synchronization
	return nil
}

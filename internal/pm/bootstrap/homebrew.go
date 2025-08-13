// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Gizzahub/gzh-manager-go/internal/logger"
)

// HomebrewBootstrapper handles Homebrew installation and configuration.
type HomebrewBootstrapper struct {
	logger logger.CommonLogger
}

// NewHomebrewBootstrapper creates a new Homebrew bootstrapper.
func NewHomebrewBootstrapper(logger logger.CommonLogger) *HomebrewBootstrapper {
	return &HomebrewBootstrapper{
		logger: logger,
	}
}

// GetName returns the name of this package manager.
func (h *HomebrewBootstrapper) GetName() string {
	return "brew"
}

// IsSupported checks if Homebrew is supported on the current platform.
func (h *HomebrewBootstrapper) IsSupported() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

// GetDependencies returns the dependencies for Homebrew (none).
func (h *HomebrewBootstrapper) GetDependencies() []string {
	return []string{} // Homebrew has no dependencies
}

// CheckInstallation checks if Homebrew is installed and configured.
func (h *HomebrewBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:      h.GetName(),
		Installed:    false,
		Dependencies: h.GetDependencies(),
		Details:      make(map[string]string),
	}

	// Check if brew command exists
	brewPath, err := exec.LookPath("brew")
	if err != nil {
		status.Issues = append(status.Issues, "Homebrew not found in PATH")
		return status, nil
	}

	status.ConfigPath = brewPath
	status.Details["path"] = brewPath

	// Get version
	cmd := exec.CommandContext(ctx, "brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		status.Issues = append(status.Issues, fmt.Sprintf("Failed to get version: %v", err))
		return status, nil
	}

	// Parse version from output like "Homebrew 4.1.14"
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		versionLine := lines[0]
		parts := strings.Fields(versionLine)
		if len(parts) >= 2 {
			status.Version = parts[1]
		}
	}

	status.Installed = true

	// Check if properly configured (PATH, etc.)
	if err := h.checkConfiguration(ctx, status); err != nil {
		status.Issues = append(status.Issues, fmt.Sprintf("Configuration issue: %v", err))
	}

	return status, nil
}

// Install installs Homebrew.
func (h *HomebrewBootstrapper) Install(ctx context.Context, force bool) error {
	h.logger.Info("Installing Homebrew", "platform", runtime.GOOS, "force", force)

	if !h.IsSupported() {
		return fmt.Errorf("Homebrew is not supported on %s", runtime.GOOS)
	}

	// Check if already installed and not forcing
	if !force {
		if status, err := h.CheckInstallation(ctx); err == nil && status.Installed {
			h.logger.Info("Homebrew already installed, skipping")
			return nil
		}
	}

	// Get installation script
	script, err := h.GetInstallScript()
	if err != nil {
		return fmt.Errorf("failed to get install script: %w", err)
	}

	h.logger.Info("Running Homebrew installation script")

	// Execute installation script
	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Allow interactive prompts

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation script failed: %w", err)
	}

	h.logger.Info("Homebrew installation completed")
	return nil
}

// Configure configures Homebrew after installation.
func (h *HomebrewBootstrapper) Configure(ctx context.Context) error {
	h.logger.Info("Configuring Homebrew")

	// Update Homebrew
	h.logger.Info("Updating Homebrew")
	cmd := exec.CommandContext(ctx, "brew", "update")
	if err := cmd.Run(); err != nil {
		h.logger.Warn("Failed to update Homebrew", "error", err)
		// Don't fail configuration for update issues
	}

	return h.updateShellProfile()
}

// GetInstallScript returns the Homebrew installation script.
func (h *HomebrewBootstrapper) GetInstallScript() (string, error) {
	// Official Homebrew installation script
	return `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`, nil
}

// Validate ensures Homebrew installation is working correctly.
func (h *HomebrewBootstrapper) Validate(ctx context.Context) error {
	h.logger.Info("Validating Homebrew installation")

	// Check if brew command works
	cmd := exec.CommandContext(ctx, "brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew command validation failed: %w", err)
	}

	// Run brew doctor to check for issues
	cmd = exec.CommandContext(ctx, "brew", "doctor")
	output, err := cmd.Output()
	if err != nil {
		h.logger.Warn("brew doctor found issues", "output", string(output))
		// Don't fail validation for doctor warnings
	}

	return nil
}

// checkConfiguration checks if Homebrew is properly configured.
func (h *HomebrewBootstrapper) checkConfiguration(ctx context.Context, status *BootstrapStatus) error {
	// Check if Homebrew directories are in PATH
	path := os.Getenv("PATH")

	var expectedPaths []string
	switch runtime.GOOS {
	case "darwin":
		// Check for both Intel and Apple Silicon paths
		expectedPaths = []string{"/opt/homebrew/bin", "/usr/local/bin"}
	case "linux":
		expectedPaths = []string{"/home/linuxbrew/.linuxbrew/bin"}
	}

	pathIssues := make([]string, 0)
	for _, expectedPath := range expectedPaths {
		if !strings.Contains(path, expectedPath) {
			pathIssues = append(pathIssues, expectedPath)
		}
	}

	if len(pathIssues) > 0 {
		status.Details["missing_paths"] = strings.Join(pathIssues, ",")
	}

	// Check HOMEBREW_PREFIX
	cmd := exec.CommandContext(ctx, "brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Homebrew prefix: %w", err)
	}

	prefix := strings.TrimSpace(string(output))
	status.Details["prefix"] = prefix

	return nil
}

// updateShellProfile updates shell configuration to include Homebrew in PATH.
func (h *HomebrewBootstrapper) updateShellProfile() error {
	h.logger.Info("Updating shell profile for Homebrew")

	shell := os.Getenv("SHELL")
	if shell == "" {
		h.logger.Warn("SHELL environment variable not set, skipping shell profile update")
		return nil
	}

	var profilePath string
	switch {
	case strings.Contains(shell, "bash"):
		profilePath = os.ExpandEnv("$HOME/.bash_profile")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			profilePath = os.ExpandEnv("$HOME/.bashrc")
		}
	case strings.Contains(shell, "zsh"):
		profilePath = os.ExpandEnv("$HOME/.zshrc")
	case strings.Contains(shell, "fish"):
		profilePath = os.ExpandEnv("$HOME/.config/fish/config.fish")
	default:
		h.logger.Warn("Unknown shell, skipping profile update", "shell", shell)
		return nil
	}

	// Check if Homebrew is already configured
	content, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read shell profile: %w", err)
	}

	homebrewInit := ""
	switch runtime.GOOS {
	case "darwin":
		// Add both Intel and Apple Silicon Homebrew paths
		homebrewInit = `
# Add Homebrew to PATH
if [[ -d "/opt/homebrew/bin" ]]; then
    export PATH="/opt/homebrew/bin:$PATH"
elif [[ -d "/usr/local/bin" ]]; then
    export PATH="/usr/local/bin:$PATH"
fi`
	case "linux":
		homebrewInit = `
# Add Homebrew to PATH
if [[ -d "/home/linuxbrew/.linuxbrew/bin" ]]; then
    export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
fi`
	}

	if strings.Contains(string(content), "Homebrew") || strings.Contains(string(content), "/opt/homebrew") {
		h.logger.Info("Homebrew already configured in shell profile")
		return nil
	}

	// Append Homebrew configuration
	file, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open shell profile for writing: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(homebrewInit + "\n"); err != nil {
		return fmt.Errorf("failed to write to shell profile: %w", err)
	}

	h.logger.Info("Updated shell profile", "profile", profilePath)
	h.logger.Info("Please restart your shell or run 'source " + profilePath + "' to activate Homebrew")

	return nil
}

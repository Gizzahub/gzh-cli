// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

const (
	asdfName       = "asdf"
	linuxPlatform  = "linux"
	darwinPlatform = "darwin"
)

// AsdfBootstrapper handles asdf installation and configuration.
type AsdfBootstrapper struct {
	logger logger.CommonLogger
}

// NewAsdfBootstrapper creates a new asdf bootstrapper.
func NewAsdfBootstrapper(logger logger.CommonLogger) *AsdfBootstrapper {
	return &AsdfBootstrapper{
		logger: logger,
	}
}

// GetName returns the name of this package manager.
func (a *AsdfBootstrapper) GetName() string {
	return asdfName
}

// IsSupported checks if asdf is supported on the current platform.
func (a *AsdfBootstrapper) IsSupported() bool {
	return runtime.GOOS == darwinPlatform || runtime.GOOS == linuxPlatform
}

// GetDependencies returns the dependencies for asdf.
func (a *AsdfBootstrapper) GetDependencies() []string {
	// asdf can be installed via Homebrew on macOS, or directly via Git
	if runtime.GOOS == darwinPlatform {
		return []string{"brew"}
	}
	return []string{} // Git is assumed to be available
}

// CheckInstallation checks if asdf is installed and configured.
func (a *AsdfBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:      a.GetName(),
		Installed:    false,
		Dependencies: a.GetDependencies(),
		Details:      make(map[string]string),
	}

	// Check if asdf command exists
	asdfPath, err := exec.LookPath("asdf")
	if err != nil {
		// Also check in common installation locations
		commonPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".asdf", "bin", "asdf"),
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				asdfPath = path
				break
			}
		}

		if asdfPath == "" {
			status.Issues = append(status.Issues, "asdf not found in PATH or common locations")
			return status, nil
		}
	}

	status.ConfigPath = asdfPath
	status.Details["path"] = asdfPath

	// Get version
	cmd := exec.CommandContext(ctx, asdfPath, "version")
	output, err := cmd.Output()
	if err != nil {
		status.Issues = append(status.Issues, fmt.Sprintf("Failed to get version: %v", err))
		return status, nil
	}

	// Parse version from output like "v0.10.2"
	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "v") // S1017 수정: 조건 없이 TrimPrefix 사용
	status.Version = version
	status.Installed = true

	// Check configuration
	if err := a.checkConfiguration(ctx, status); err != nil {
		status.Issues = append(status.Issues, fmt.Sprintf("Configuration issue: %v", err))
	}

	return status, nil
}

// Install installs asdf.
func (a *AsdfBootstrapper) Install(ctx context.Context, force bool) error {
	a.logger.Info("Installing asdf", "platform", runtime.GOOS, "force", force)

	if !a.IsSupported() {
		return fmt.Errorf("asdf is not supported on %s", runtime.GOOS)
	}

	// Check if already installed and not forcing
	if !force {
		if status, err := a.CheckInstallation(ctx); err == nil && status.Installed {
			a.logger.Info("asdf already installed, skipping")
			return nil
		}
	}

	switch runtime.GOOS {
	case darwinPlatform:
		return a.installViaBrew(ctx)
	case linuxPlatform:
		return a.installViaGit(ctx)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Configure configures asdf after installation.
func (a *AsdfBootstrapper) Configure(_ context.Context) error {
	a.logger.Info("Configuring asdf")

	return a.updateShellProfile()
}

// GetInstallScript returns the asdf installation script.
func (a *AsdfBootstrapper) GetInstallScript() (string, error) {
	switch runtime.GOOS {
	case darwinPlatform:
		return "brew install asdf", nil
	case linuxPlatform:
		return `git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.10.2`, nil
	default:
		return "", fmt.Errorf("no install script for platform %s", runtime.GOOS)
	}
}

// Validate ensures asdf installation is working correctly.
func (a *AsdfBootstrapper) Validate(ctx context.Context) error {
	a.logger.Info("Validating asdf installation")

	// Check if asdf command works
	cmd := exec.CommandContext(ctx, "asdf", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf command validation failed: %w", err)
	}

	// Check if basic plugins can be listed
	cmd = exec.CommandContext(ctx, "asdf", "plugin", "list", "all")
	if err := cmd.Run(); err != nil {
		a.logger.Warn("asdf plugin list failed", "error", err)
		// Don't fail validation for plugin issues
	}

	return nil
}

// installViaBrew installs asdf using Homebrew on macOS.
func (a *AsdfBootstrapper) installViaBrew(ctx context.Context) error {
	a.logger.Info("Installing asdf via Homebrew")

	cmd := exec.CommandContext(ctx, "brew", "install", "asdf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install asdf via brew: %w", err)
	}

	return nil
}

// installViaGit installs asdf by cloning from Git.
func (a *AsdfBootstrapper) installViaGit(ctx context.Context) error {
	a.logger.Info("Installing asdf via Git clone")

	asdfDir := filepath.Join(os.Getenv("HOME"), ".asdf")

	// Remove existing directory if forcing
	if _, err := os.Stat(asdfDir); err == nil {
		a.logger.Info("Removing existing asdf directory", "dir", asdfDir)
		if err := os.RemoveAll(asdfDir); err != nil {
			return fmt.Errorf("failed to remove existing asdf directory: %w", err)
		}
	}

	cmd := exec.CommandContext(ctx, "git", "clone",
		"https://github.com/asdf-vm/asdf.git", asdfDir, "--branch", "v0.10.2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone asdf repository: %w", err)
	}

	return nil
}

// checkConfiguration checks if asdf is properly configured.
func (a *AsdfBootstrapper) checkConfiguration(ctx context.Context, status *BootstrapStatus) error {
	// Check if asdf is in PATH
	path := os.Getenv("PATH")
	asdfBinPath := filepath.Join(os.Getenv("HOME"), ".asdf", "bin")

	if !strings.Contains(path, asdfBinPath) {
		status.Details["missing_path"] = asdfBinPath
	}

	// Check if asdf directory exists
	asdfDir := filepath.Join(os.Getenv("HOME"), ".asdf")
	if _, err := os.Stat(asdfDir); os.IsNotExist(err) {
		return fmt.Errorf("asdf directory not found: %s", asdfDir)
	}

	status.Details["asdf_dir"] = asdfDir

	// Check if any plugins are installed
	cmd := exec.CommandContext(ctx, "asdf", "plugin", "list")
	output, err := cmd.Output()
	if err != nil {
		status.Details["plugins"] = "error getting plugin list"
	} else {
		plugins := strings.TrimSpace(string(output))
		if plugins == "" {
			plugins = "none"
		}
		status.Details["plugins"] = plugins
	}

	return nil
}

// updateShellProfile updates shell configuration to include asdf.
func (a *AsdfBootstrapper) updateShellProfile() error {
	a.logger.Info("Updating shell profile for asdf")

	shell := os.Getenv("SHELL")
	if shell == "" {
		a.logger.Warn("SHELL environment variable not set, skipping shell profile update")
		return nil
	}

	var profilePath string
	var asdfInit string

	asdfDir := filepath.Join(os.Getenv("HOME"), ".asdf")

	switch {
	case strings.Contains(shell, "bash"):
		profilePath = os.ExpandEnv("$HOME/.bash_profile")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			profilePath = os.ExpandEnv("$HOME/.bashrc")
		}
		asdfInit = fmt.Sprintf(`
# Add asdf to PATH and initialize
if [[ -d "%s" ]]; then
    export PATH="%s/bin:$PATH"
    source "%s/asdf.sh"
fi`, asdfDir, asdfDir, asdfDir)

	case strings.Contains(shell, "zsh"):
		profilePath = os.ExpandEnv("$HOME/.zshrc")
		asdfInit = fmt.Sprintf(`
# Add asdf to PATH and initialize
if [[ -d "%s" ]]; then
    export PATH="%s/bin:$PATH"
    source "%s/asdf.sh"
fi`, asdfDir, asdfDir, asdfDir)

	case strings.Contains(shell, "fish"):
		profilePath = os.ExpandEnv("$HOME/.config/fish/config.fish")
		asdfInit = fmt.Sprintf(`
# Add asdf to PATH and initialize
if test -d "%s"
    set -gx PATH "%s/bin" $PATH
    source "%s/asdf.fish"
end`, asdfDir, asdfDir, asdfDir)

	default:
		a.logger.Warn("Unknown shell, skipping profile update", "shell", shell)
		return nil
	}

	// Check if asdf is already configured
	content, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read shell profile: %w", err)
	}

	if strings.Contains(string(content), "asdf") || strings.Contains(string(content), ".asdf") {
		a.logger.Info("asdf already configured in shell profile")
		return nil
	}

	// Create directory if it doesn't exist (for fish config)
	if strings.Contains(profilePath, "config/fish") {
		if err := os.MkdirAll(filepath.Dir(profilePath), 0o750); err != nil {
			return fmt.Errorf("failed to create fish config directory: %w", err)
		}
	}

	// Append asdf configuration
	file, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open shell profile for writing: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(asdfInit + "\n"); err != nil {
		return fmt.Errorf("failed to write to shell profile: %w", err)
	}

	a.logger.Info("Updated shell profile", "profile", profilePath)
	a.logger.Info("Please restart your shell or run 'source " + profilePath + "' to activate asdf")

	return nil
}

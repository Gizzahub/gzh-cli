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

	"github.com/Gizzahub/gzh-manager-go/internal/logger"
)

// NvmBootstrapper handles Node Version Manager installation.
type NvmBootstrapper struct {
	logger logger.CommonLogger
}

// NewNvmBootstrapper creates a new NVM bootstrapper.
func NewNvmBootstrapper(logger logger.CommonLogger) *NvmBootstrapper {
	return &NvmBootstrapper{logger: logger}
}

func (n *NvmBootstrapper) GetName() string { return "nvm" }

func (n *NvmBootstrapper) IsSupported() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func (n *NvmBootstrapper) GetDependencies() []string { return []string{} }

func (n *NvmBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:   n.GetName(),
		Installed: false,
		Details:   make(map[string]string),
	}

	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(os.Getenv("HOME"), ".nvm")
	}

	nvmScript := filepath.Join(nvmDir, "nvm.sh")
	if _, err := os.Stat(nvmScript); os.IsNotExist(err) {
		status.Issues = append(status.Issues, "nvm.sh not found")
		return status, nil
	}

	status.ConfigPath = nvmScript
	status.Installed = true

	// Try to get version by sourcing nvm.sh
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("source %s && nvm --version", nvmScript))
	if output, err := cmd.Output(); err == nil {
		status.Version = strings.TrimSpace(string(output))
	}

	status.Details["nvm_dir"] = nvmDir
	return status, nil
}

func (n *NvmBootstrapper) Install(ctx context.Context, force bool) error {
	n.logger.Info("Installing nvm")

	script, _ := n.GetInstallScript()
	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NvmBootstrapper) Configure(ctx context.Context) error {
	return n.updateShellProfile()
}

func (n *NvmBootstrapper) GetInstallScript() (string, error) {
	return `curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash`, nil
}

func (n *NvmBootstrapper) Validate(ctx context.Context) error {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(os.Getenv("HOME"), ".nvm")
	}

	nvmScript := filepath.Join(nvmDir, "nvm.sh")
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("source %s && nvm --version", nvmScript))
	return cmd.Run()
}

func (n *NvmBootstrapper) updateShellProfile() error {
	// NVM installer usually handles this automatically
	n.logger.Info("nvm shell profile should be updated automatically by installer")
	return nil
}

// RbenvBootstrapper handles Ruby Version Manager installation.
type RbenvBootstrapper struct {
	logger logger.CommonLogger
}

// NewRbenvBootstrapper creates a new rbenv bootstrapper.
func NewRbenvBootstrapper(logger logger.CommonLogger) *RbenvBootstrapper {
	return &RbenvBootstrapper{logger: logger}
}

func (r *RbenvBootstrapper) GetName() string { return "rbenv" }

func (r *RbenvBootstrapper) IsSupported() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func (r *RbenvBootstrapper) GetDependencies() []string {
	if runtime.GOOS == "darwin" {
		return []string{"brew"}
	}
	return []string{}
}

func (r *RbenvBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:      r.GetName(),
		Installed:    false,
		Dependencies: r.GetDependencies(),
		Details:      make(map[string]string),
	}

	rbenvPath, err := exec.LookPath("rbenv")
	if err != nil {
		status.Issues = append(status.Issues, "rbenv not found in PATH")
		return status, nil
	}

	status.ConfigPath = rbenvPath
	status.Installed = true

	// Get version
	cmd := exec.CommandContext(ctx, "rbenv", "--version")
	if output, err := cmd.Output(); err == nil {
		version := strings.TrimSpace(string(output))
		parts := strings.Fields(version)
		if len(parts) >= 2 {
			status.Version = parts[1]
		}
	}

	return status, nil
}

func (r *RbenvBootstrapper) Install(ctx context.Context, force bool) error {
	r.logger.Info("Installing rbenv")

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "brew", "install", "rbenv")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "linux":
		// Install via Git
		rbenvDir := filepath.Join(os.Getenv("HOME"), ".rbenv")
		cmd := exec.CommandContext(ctx, "git", "clone", "https://github.com/rbenv/rbenv.git", rbenvDir)
		if err := cmd.Run(); err != nil {
			return err
		}

		// Install ruby-build plugin
		rubyBuildDir := filepath.Join(rbenvDir, "plugins", "ruby-build")
		cmd = exec.CommandContext(ctx, "git", "clone", "https://github.com/rbenv/ruby-build.git", rubyBuildDir)
		return cmd.Run()
	}

	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func (r *RbenvBootstrapper) Configure(ctx context.Context) error {
	return r.updateShellProfile()
}

func (r *RbenvBootstrapper) GetInstallScript() (string, error) {
	if runtime.GOOS == "darwin" {
		return "brew install rbenv", nil
	}
	return "git clone https://github.com/rbenv/rbenv.git ~/.rbenv", nil
}

func (r *RbenvBootstrapper) Validate(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "rbenv", "--version")
	return cmd.Run()
}

func (r *RbenvBootstrapper) updateShellProfile() error {
	r.logger.Info("Please add 'eval \"$(rbenv init -)\"' to your shell profile")
	return nil
}

// PyenvBootstrapper handles Python Version Manager installation.
type PyenvBootstrapper struct {
	logger logger.CommonLogger
}

// NewPyenvBootstrapper creates a new pyenv bootstrapper.
func NewPyenvBootstrapper(logger logger.CommonLogger) *PyenvBootstrapper {
	return &PyenvBootstrapper{logger: logger}
}

func (p *PyenvBootstrapper) GetName() string { return "pyenv" }

func (p *PyenvBootstrapper) IsSupported() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func (p *PyenvBootstrapper) GetDependencies() []string {
	if runtime.GOOS == "darwin" {
		return []string{"brew"}
	}
	return []string{}
}

func (p *PyenvBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:      p.GetName(),
		Installed:    false,
		Dependencies: p.GetDependencies(),
		Details:      make(map[string]string),
	}

	pyenvPath, err := exec.LookPath("pyenv")
	if err != nil {
		status.Issues = append(status.Issues, "pyenv not found in PATH")
		return status, nil
	}

	status.ConfigPath = pyenvPath
	status.Installed = true

	// Get version
	cmd := exec.CommandContext(ctx, "pyenv", "--version")
	if output, err := cmd.Output(); err == nil {
		version := strings.TrimSpace(string(output))
		parts := strings.Fields(version)
		if len(parts) >= 2 {
			status.Version = parts[1]
		}
	}

	return status, nil
}

func (p *PyenvBootstrapper) Install(ctx context.Context, force bool) error {
	p.logger.Info("Installing pyenv")

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "brew", "install", "pyenv")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "linux":
		script := `curl https://pyenv.run | bash`
		cmd := exec.CommandContext(ctx, "bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func (p *PyenvBootstrapper) Configure(ctx context.Context) error {
	return p.updateShellProfile()
}

func (p *PyenvBootstrapper) GetInstallScript() (string, error) {
	if runtime.GOOS == "darwin" {
		return "brew install pyenv", nil
	}
	return "curl https://pyenv.run | bash", nil
}

func (p *PyenvBootstrapper) Validate(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "pyenv", "--version")
	return cmd.Run()
}

func (p *PyenvBootstrapper) updateShellProfile() error {
	p.logger.Info("Please add pyenv to your shell profile with 'pyenv init'")
	return nil
}

// SdkmanBootstrapper handles SDKMAN installation.
type SdkmanBootstrapper struct {
	logger logger.CommonLogger
}

// NewSdkmanBootstrapper creates a new SDKMAN bootstrapper.
func NewSdkmanBootstrapper(logger logger.CommonLogger) *SdkmanBootstrapper {
	return &SdkmanBootstrapper{logger: logger}
}

func (s *SdkmanBootstrapper) GetName() string { return "sdkman" }

func (s *SdkmanBootstrapper) IsSupported() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func (s *SdkmanBootstrapper) GetDependencies() []string { return []string{} }

func (s *SdkmanBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	status := &BootstrapStatus{
		Manager:   s.GetName(),
		Installed: false,
		Details:   make(map[string]string),
	}

	sdkmanDir := filepath.Join(os.Getenv("HOME"), ".sdkman")
	sdkmanInit := filepath.Join(sdkmanDir, "bin", "sdkman-init.sh")

	if _, err := os.Stat(sdkmanInit); os.IsNotExist(err) {
		status.Issues = append(status.Issues, "SDKMAN not found")
		return status, nil
	}

	status.ConfigPath = sdkmanInit
	status.Installed = true

	// Try to get version
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("source %s && sdk version", sdkmanInit))
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.Contains(line, "SDKMAN") && strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					status.Version = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	status.Details["sdkman_dir"] = sdkmanDir
	return status, nil
}

func (s *SdkmanBootstrapper) Install(ctx context.Context, force bool) error {
	s.logger.Info("Installing SDKMAN")

	script, _ := s.GetInstallScript()
	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *SdkmanBootstrapper) Configure(ctx context.Context) error {
	// SDKMAN installer handles shell profile automatically
	return nil
}

func (s *SdkmanBootstrapper) GetInstallScript() (string, error) {
	return `curl -s "https://get.sdkman.io" | bash`, nil
}

func (s *SdkmanBootstrapper) Validate(ctx context.Context) error {
	sdkmanInit := filepath.Join(os.Getenv("HOME"), ".sdkman", "bin", "sdkman-init.sh")
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("source %s && sdk version", sdkmanInit))
	return cmd.Run()
}

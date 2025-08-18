package upgrade

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// NvmUpgrader implements PackageManagerUpgrader for Node Version Manager
type NvmUpgrader struct {
	logger logger.CommonLogger
}

// NewNvmUpgrader creates a new nvm upgrader
func NewNvmUpgrader(logger logger.CommonLogger) *NvmUpgrader {
	return &NvmUpgrader{logger: logger}
}

func (n *NvmUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	currentVersion, err := n.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current nvm version: %w", err)
	}

	return &UpgradeStatus{
		Manager:         "nvm",
		CurrentVersion:  currentVersion,
		LatestVersion:   "latest",
		UpdateAvailable: true, // Always suggest updating nvm
		UpdateMethod:    n.GetUpdateMethod(),
		ChangelogURL:    "https://github.com/nvm-sh/nvm/releases",
	}, nil
}

func (n *NvmUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	n.logger.Info("Starting nvm upgrade")

	// Download and run the latest install script
	cmd := exec.CommandContext(ctx, "curl", "-o-", "https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh")
	installScript, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to download nvm install script: %w", err)
	}

	cmd = exec.CommandContext(ctx, "bash", "-c", string(installScript))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nvm upgrade failed: %w", err)
	}

	n.logger.Info("nvm upgrade completed successfully")
	return nil
}

func (n *NvmUpgrader) Backup(ctx context.Context) (string, error) {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(os.Getenv("HOME"), ".nvm")
	}

	backupPath := fmt.Sprintf("/tmp/nvm-backup-%d.txt", time.Now().Unix())
	if err := writeFile(backupPath, fmt.Sprintf("NVM_DIR=%s", nvmDir)); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

func (n *NvmUpgrader) Rollback(ctx context.Context, backupPath string) error {
	n.logger.Info("nvm rollback not implemented - manual restoration required")
	return nil
}

func (n *NvmUpgrader) GetUpdateMethod() string {
	return "curl install script"
}

func (n *NvmUpgrader) ValidateUpgrade(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "bash", "-c", "source ~/.nvm/nvm.sh && nvm --version")
	return cmd.Run()
}

func (n *NvmUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", "source ~/.nvm/nvm.sh && nvm --version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RbenvUpgrader implements PackageManagerUpgrader for rbenv
type RbenvUpgrader struct {
	logger logger.CommonLogger
}

func NewRbenvUpgrader(logger logger.CommonLogger) *RbenvUpgrader {
	return &RbenvUpgrader{logger: logger}
}

func (r *RbenvUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	currentVersion, err := r.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current rbenv version: %w", err)
	}

	return &UpgradeStatus{
		Manager:         "rbenv",
		CurrentVersion:  currentVersion,
		LatestVersion:   "latest",
		UpdateAvailable: true,
		UpdateMethod:    r.GetUpdateMethod(),
		ChangelogURL:    "https://github.com/rbenv/rbenv/releases",
	}, nil
}

func (r *RbenvUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	r.logger.Info("Starting rbenv upgrade")

	if runtime.GOOS == "darwin" {
		// macOS: Use Homebrew
		cmd := exec.CommandContext(ctx, "brew", "upgrade", "rbenv")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("rbenv upgrade via brew failed: %w", err)
		}
	} else {
		// Linux: Git pull in rbenv directory
		rbenvDir := r.getRbenvDir()
		if rbenvDir == "" {
			return fmt.Errorf("rbenv directory not found")
		}

		cmd := exec.CommandContext(ctx, "git", "pull")
		cmd.Dir = rbenvDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("rbenv git pull failed: %w", err)
		}
	}

	r.logger.Info("rbenv upgrade completed successfully")
	return nil
}

func (r *RbenvUpgrader) Backup(ctx context.Context) (string, error) {
	rbenvDir := r.getRbenvDir()
	backupPath := fmt.Sprintf("/tmp/rbenv-backup-%d.txt", time.Now().Unix())

	if rbenvDir != "" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmd.Dir = rbenvDir
		output, err := cmd.Output()
		if err == nil {
			if err := writeFile(backupPath, strings.TrimSpace(string(output))); err != nil {
				return "", err
			}
			return backupPath, nil
		}
	}

	if err := writeFile(backupPath, "rbenv-backup"); err != nil {
		return "", err
	}
	return backupPath, nil
}

func (r *RbenvUpgrader) Rollback(ctx context.Context, backupPath string) error {
	r.logger.Info("rbenv rollback limited - backup available at: %s", backupPath)
	return nil
}

func (r *RbenvUpgrader) GetUpdateMethod() string {
	if runtime.GOOS == "darwin" {
		return "brew upgrade rbenv"
	}
	return "git pull"
}

func (r *RbenvUpgrader) ValidateUpgrade(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "rbenv", "--version")
	return cmd.Run()
}

func (r *RbenvUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "rbenv", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionRegex := regexp.MustCompile(`rbenv\s+(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return strings.TrimSpace(string(output)), nil
	}
	return matches[1], nil
}

func (r *RbenvUpgrader) getRbenvDir() string {
	possiblePaths := []string{
		os.Getenv("RBENV_ROOT"),
		filepath.Join(os.Getenv("HOME"), ".rbenv"),
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

// PyenvUpgrader implements PackageManagerUpgrader for pyenv
type PyenvUpgrader struct {
	logger logger.CommonLogger
}

func NewPyenvUpgrader(logger logger.CommonLogger) *PyenvUpgrader {
	return &PyenvUpgrader{logger: logger}
}

func (p *PyenvUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	currentVersion, err := p.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pyenv version: %w", err)
	}

	return &UpgradeStatus{
		Manager:         "pyenv",
		CurrentVersion:  currentVersion,
		LatestVersion:   "latest",
		UpdateAvailable: true,
		UpdateMethod:    p.GetUpdateMethod(),
		ChangelogURL:    "https://github.com/pyenv/pyenv/releases",
	}, nil
}

func (p *PyenvUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	p.logger.Info("Starting pyenv upgrade")

	if runtime.GOOS == "darwin" {
		// macOS: Use Homebrew
		cmd := exec.CommandContext(ctx, "brew", "upgrade", "pyenv")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pyenv upgrade via brew failed: %w", err)
		}
	} else {
		// Linux: Use pyenv-installer or git pull
		pyenvDir := p.getPyenvDir()
		if pyenvDir != "" {
			cmd := exec.CommandContext(ctx, "git", "pull")
			cmd.Dir = pyenvDir
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("pyenv git pull failed: %w", err)
			}
		} else {
			// Try pyenv-installer
			cmd := exec.CommandContext(ctx, "curl", "-L", "https://github.com/pyenv/pyenv-installer/raw/master/bin/pyenv-installer")
			installScript, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("failed to download pyenv installer: %w", err)
			}

			cmd = exec.CommandContext(ctx, "bash", "-c", string(installScript))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("pyenv installation failed: %w", err)
			}
		}
	}

	p.logger.Info("pyenv upgrade completed successfully")
	return nil
}

func (p *PyenvUpgrader) Backup(ctx context.Context) (string, error) {
	pyenvDir := p.getPyenvDir()
	backupPath := fmt.Sprintf("/tmp/pyenv-backup-%d.txt", time.Now().Unix())

	if pyenvDir != "" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmd.Dir = pyenvDir
		output, err := cmd.Output()
		if err == nil {
			if err := writeFile(backupPath, strings.TrimSpace(string(output))); err != nil {
				return "", err
			}
			return backupPath, nil
		}
	}

	if err := writeFile(backupPath, "pyenv-backup"); err != nil {
		return "", err
	}
	return backupPath, nil
}

func (p *PyenvUpgrader) Rollback(ctx context.Context, backupPath string) error {
	p.logger.Info("pyenv rollback limited - backup available at: %s", backupPath)
	return nil
}

func (p *PyenvUpgrader) GetUpdateMethod() string {
	if runtime.GOOS == "darwin" {
		return "brew upgrade pyenv"
	}
	return "git pull or pyenv-installer"
}

func (p *PyenvUpgrader) ValidateUpgrade(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "pyenv", "--version")
	return cmd.Run()
}

func (p *PyenvUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "pyenv", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionRegex := regexp.MustCompile(`pyenv\s+(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return strings.TrimSpace(string(output)), nil
	}
	return matches[1], nil
}

func (p *PyenvUpgrader) getPyenvDir() string {
	possiblePaths := []string{
		os.Getenv("PYENV_ROOT"),
		filepath.Join(os.Getenv("HOME"), ".pyenv"),
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

// SdkmanUpgrader implements PackageManagerUpgrader for SDKMAN!
type SdkmanUpgrader struct {
	logger logger.CommonLogger
}

func NewSdkmanUpgrader(logger logger.CommonLogger) *SdkmanUpgrader {
	return &SdkmanUpgrader{logger: logger}
}

func (s *SdkmanUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	currentVersion, err := s.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current SDKMAN version: %w", err)
	}

	return &UpgradeStatus{
		Manager:         "sdkman",
		CurrentVersion:  currentVersion,
		LatestVersion:   "latest",
		UpdateAvailable: true,
		UpdateMethod:    s.GetUpdateMethod(),
		ChangelogURL:    "https://github.com/sdkman/sdkman-cli/releases",
	}, nil
}

func (s *SdkmanUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	s.logger.Info("Starting SDKMAN upgrade")

	// Use SDKMAN's self-update command
	cmd := exec.CommandContext(ctx, "bash", "-c", "source ~/.sdkman/bin/sdkman-init.sh && sdk selfupdate")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SDKMAN selfupdate failed: %w", err)
	}

	s.logger.Info("SDKMAN upgrade completed successfully")
	return nil
}

func (s *SdkmanUpgrader) Backup(ctx context.Context) (string, error) {
	sdkmanDir := filepath.Join(os.Getenv("HOME"), ".sdkman")
	backupPath := fmt.Sprintf("/tmp/sdkman-backup-%d.txt", time.Now().Unix())

	if err := writeFile(backupPath, fmt.Sprintf("SDKMAN_DIR=%s", sdkmanDir)); err != nil {
		return "", err
	}
	return backupPath, nil
}

func (s *SdkmanUpgrader) Rollback(ctx context.Context, backupPath string) error {
	s.logger.Info("SDKMAN rollback not implemented - backup available at: %s", backupPath)
	return nil
}

func (s *SdkmanUpgrader) GetUpdateMethod() string {
	return "sdk selfupdate"
}

func (s *SdkmanUpgrader) ValidateUpgrade(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "bash", "-c", "source ~/.sdkman/bin/sdkman-init.sh && sdk version")
	return cmd.Run()
}

func (s *SdkmanUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", "source ~/.sdkman/bin/sdkman-init.sh && sdk version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionRegex := regexp.MustCompile(`SDKMAN!?\s+(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return strings.TrimSpace(string(output)), nil
	}
	return matches[1], nil
}

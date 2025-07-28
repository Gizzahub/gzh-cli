// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// doctorOptions contains options for the doctor command.
type doctorOptions struct {
	verbose bool
	fix     bool
}

// newDoctorCmd creates a new doctor command for checking git-synclone installation.
func newDoctorCmd() *cobra.Command {
	opts := &doctorOptions{}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check git-synclone installation and configuration",
		Long: `Diagnose git-synclone installation and configuration issues.

This command checks:
- Git installation and version
- git-synclone binary availability
- PATH configuration
- Configuration file accessibility
- Git integration functionality

Examples:
  git synclone doctor              # Basic health check
  git synclone doctor --verbose    # Detailed diagnostics
  git synclone doctor --fix        # Attempt to fix common issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Show detailed diagnostic information")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Attempt to fix common configuration issues")

	return cmd
}

// runDoctor performs the installation and configuration checks.
func runDoctor(opts *doctorOptions) error {
	fmt.Println("🔍 git-synclone Doctor - Installation Diagnostics")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	issues := 0

	// Check system information
	if opts.verbose {
		printSystemInfo()
	}

	// Check Git installation
	issues += checkGitInstallation(opts)

	// Check git-synclone binary
	issues += checkBinaryInstallation(opts)

	// Check PATH configuration
	issues += checkPathConfiguration(opts)

	// Check Git integration
	issues += checkGitIntegration(opts)

	// Check configuration files
	issues += checkConfiguration(opts)

	// Check dependencies
	issues += checkDependencies(opts)

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))

	if issues == 0 {
		fmt.Println("✅ All checks passed! git-synclone is properly installed and configured.")
	} else {
		fmt.Printf("⚠️  Found %d issue(s). See details above.\n", issues)
		if !opts.fix {
			fmt.Println("💡 Run with --fix to attempt automatic fixes.")
		}
		return fmt.Errorf("found %d installation issues", issues)
	}

	return nil
}

// printSystemInfo displays system information.
func printSystemInfo() {
	fmt.Println("📋 System Information:")
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	if homeDir, err := os.UserHomeDir(); err == nil {
		fmt.Printf("  Home Directory: %s\n", homeDir)
	}
	fmt.Println()
}

// checkGitInstallation verifies Git is properly installed.
func checkGitInstallation(opts *doctorOptions) int {
	fmt.Println("🔧 Checking Git Installation:")

	// Check if git command exists
	gitPath, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("  ❌ Git is not installed or not in PATH")
		fmt.Println("     Please install Git: https://git-scm.com/downloads")
		return 1
	}

	fmt.Printf("  ✅ Git found at: %s\n", gitPath)

	// Check Git version
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  ❌ Failed to get Git version")
		return 1
	}

	version := strings.TrimSpace(string(output))
	fmt.Printf("  ✅ %s\n", version)

	if opts.verbose {
		// Check Git configuration
		if email := getGitConfig("user.email"); email != "" {
			fmt.Printf("  ✅ Git user.email: %s\n", email)
		} else {
			fmt.Println("  ⚠️  Git user.email not configured")
		}

		if name := getGitConfig("user.name"); name != "" {
			fmt.Printf("  ✅ Git user.name: %s\n", name)
		} else {
			fmt.Println("  ⚠️  Git user.name not configured")
		}
	}

	fmt.Println()
	return 0
}

// checkBinaryInstallation verifies git-synclone binary is installed.
func checkBinaryInstallation(opts *doctorOptions) int {
	fmt.Println("📦 Checking git-synclone Binary:")

	// Check if git-synclone command exists
	binaryPath, err := exec.LookPath("git-synclone")
	if err != nil {
		fmt.Println("  ❌ git-synclone binary not found in PATH")
		fmt.Println("     Run installation script: ./scripts/install-git-extensions.sh")
		return 1
	}

	fmt.Printf("  ✅ Binary found at: %s\n", binaryPath)

	// Check if binary is executable
	if info, err := os.Stat(binaryPath); err == nil {
		if info.Mode()&0o111 == 0 {
			fmt.Println("  ❌ Binary is not executable")
			if opts.fix {
				if err := os.Chmod(binaryPath, 0o755); err == nil {
					fmt.Println("  🔧 Fixed: Made binary executable")
				} else {
					fmt.Printf("  ❌ Failed to fix permissions: %v\n", err)
					return 1
				}
			}
		} else {
			fmt.Println("  ✅ Binary is executable")
		}
	}

	// Check binary version
	cmd := exec.Command("git-synclone", "--version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  ❌ Failed to get binary version")
		if opts.verbose {
			fmt.Printf("     Error: %v\n", err)
		}
		return 1
	}

	version := strings.TrimSpace(string(output))
	fmt.Printf("  ✅ Version: %s\n", version)

	fmt.Println()
	return 0
}

// checkPathConfiguration verifies PATH is properly configured.
func checkPathConfiguration(opts *doctorOptions) int {
	fmt.Println("🛣️  Checking PATH Configuration:")

	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, string(os.PathListSeparator))

	// Common installation directories
	commonDirs := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "bin"),
		filepath.Join(os.Getenv("GOPATH"), "bin"),
		"/usr/local/bin",
		"/usr/bin",
	}

	found := false
	for _, dir := range commonDirs {
		if dir == "" {
			continue
		}

		binaryPath := filepath.Join(dir, "git-synclone")
		if _, err := os.Stat(binaryPath); err == nil {
			fmt.Printf("  ✅ Found git-synclone in: %s\n", dir)

			// Check if this directory is in PATH
			inPath := false
			for _, pathDir := range pathDirs {
				if pathDir == dir {
					inPath = true
					break
				}
			}

			if inPath {
				fmt.Printf("  ✅ %s is in PATH\n", dir)
				found = true
			} else {
				fmt.Printf("  ⚠️  %s is NOT in PATH\n", dir)
			}
		}
	}

	if !found {
		fmt.Println("  ❌ git-synclone not found in any standard location")
		fmt.Println("     Add installation directory to PATH:")
		fmt.Println("     export PATH=\"$PATH:~/.local/bin\"")
		return 1
	}

	fmt.Println()
	return 0
}

// checkGitIntegration verifies Git integration works.
func checkGitIntegration(opts *doctorOptions) int {
	fmt.Println("🔗 Checking Git Integration:")

	// Test git synclone command
	cmd := exec.Command("git", "synclone", "--help")
	err := cmd.Run()
	if err != nil {
		fmt.Println("  ❌ 'git synclone' command not working")
		fmt.Println("     Ensure git-synclone is in PATH and executable")
		if opts.verbose {
			fmt.Printf("     Error: %v\n", err)
		}
		return 1
	}

	fmt.Println("  ✅ 'git synclone' command works")

	// Test specific subcommands
	subcommands := []string{"github", "gitlab", "gitea"}
	for _, subcmd := range subcommands {
		cmd := exec.Command("git", "synclone", subcmd, "--help")
		if err := cmd.Run(); err == nil {
			fmt.Printf("  ✅ 'git synclone %s' available\n", subcmd)
		} else {
			fmt.Printf("  ⚠️  'git synclone %s' not available\n", subcmd)
		}
	}

	fmt.Println()
	return 0
}

// checkConfiguration verifies configuration files and settings.
func checkConfiguration(opts *doctorOptions) int {
	fmt.Println("⚙️  Checking Configuration:")

	// Check for configuration files
	configPaths := []string{
		"./synclone.yaml",
		"./synclone.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "gzh-manager", "synclone.yaml"),
		"/etc/gzh-manager/synclone.yaml",
	}

	foundConfig := false
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("  ✅ Found config: %s\n", configPath)
			foundConfig = true
		} else if opts.verbose {
			fmt.Printf("  ⚪ No config at: %s\n", configPath)
		}
	}

	if !foundConfig {
		fmt.Println("  ⚠️  No configuration files found")
		fmt.Println("     Configuration is optional but recommended")
		fmt.Println("     See examples/ directory for sample configurations")
	}

	// Check config directory permissions
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "gzh-manager")
	if info, err := os.Stat(configDir); err == nil {
		if info.IsDir() {
			fmt.Printf("  ✅ Config directory exists: %s\n", configDir)
		}
	} else if opts.fix {
		if err := os.MkdirAll(configDir, 0o755); err == nil {
			fmt.Printf("  🔧 Created config directory: %s\n", configDir)
		}
	}

	fmt.Println()
	return 0
}

// checkDependencies verifies required dependencies.
func checkDependencies(opts *doctorOptions) int {
	fmt.Println("📚 Checking Dependencies:")

	issues := 0

	// Check for required environment variables (optional)
	envVars := []string{"GITHUB_TOKEN", "GITLAB_TOKEN", "GITEA_TOKEN"}
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			if opts.verbose {
				fmt.Printf("  ✅ %s is set\n", envVar)
			}
		} else {
			fmt.Printf("  ⚪ %s not set (optional for public repos)\n", envVar)
		}
	}

	// Check disk space in common locations
	if opts.verbose {
		checkDiskSpace("Home directory", os.Getenv("HOME"))
		checkDiskSpace("Temp directory", os.TempDir())
	}

	fmt.Println()
	return issues
}

// getGitConfig gets a Git configuration value.
func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// checkDiskSpace checks available disk space.
func checkDiskSpace(name, path string) {
	// This is a simplified check - in production, you might want to use syscalls
	if path == "" {
		return
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fmt.Printf("  ✅ %s accessible: %s\n", name, path)
	} else {
		fmt.Printf("  ⚠️  %s not accessible: %s\n", name, path)
	}
}

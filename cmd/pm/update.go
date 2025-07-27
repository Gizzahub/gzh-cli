// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newUpdateCmd(ctx context.Context) *cobra.Command {
	var (
		allManagers bool
		manager     string
		strategy    string
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update packages based on version strategy",
		Long: `Update packages for specified package managers based on configured version strategy.

Supports update strategies:
- latest: Update to the absolute latest version
- stable: Update to the latest stable version (default)
- minor: Update to latest patch/minor version only
- fixed: Keep exact versions (no updates)

Examples:
  # Update all package managers
  gz pm update --all

  # Update specific package manager
  gz pm update --manager brew

  # Update with specific strategy
  gz pm update --manager asdf --strategy latest

  # Dry run to see what would be updated
  gz pm update --all --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, redirect to existing commands
			// In future, implement unified update logic
			if manager != "" {
				return runUpdateManager(ctx, manager, strategy, dryRun)
			}
			if allManagers {
				return runUpdateAll(ctx, strategy, dryRun)
			}
			return fmt.Errorf("specify --manager or --all")
		},
	}

	cmd.Flags().BoolVar(&allManagers, "all", false, "Update all package managers")
	cmd.Flags().StringVar(&manager, "manager", "", "Package manager to update")
	cmd.Flags().StringVar(&strategy, "strategy", "stable", "Update strategy: latest, stable, minor, fixed")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be updated without making changes")

	return cmd
}

func runUpdateManager(ctx context.Context, manager, strategy string, dryRun bool) error {
	fmt.Printf("Updating %s packages with strategy: %s\n", manager, strategy)
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}

	// TODO: Implement unified update logic
	// For now, this is a placeholder
	switch manager {
	case "brew":
		return updateBrew(ctx, strategy, dryRun)
	case "asdf":
		return updateAsdf(ctx, strategy, dryRun)
	case "sdkman":
		return updateSdkman(ctx, strategy, dryRun)
	case "apt":
		return updateApt(ctx, strategy, dryRun)
	case "pip":
		return updatePip(ctx, strategy, dryRun)
	case "npm":
		return updateNpm(ctx, strategy, dryRun)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func runUpdateAll(ctx context.Context, strategy string, dryRun bool) error {
	managers := []string{"brew", "asdf", "sdkman", "apt", "pip", "npm"}

	fmt.Println("Updating all package managers...")
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}
	fmt.Println()

	for _, manager := range managers {
		fmt.Printf("=== Updating %s ===\n", manager)
		if err := runUpdateManager(ctx, manager, strategy, dryRun); err != nil {
			fmt.Printf("Warning: Failed to update %s: %v\n", manager, err)
			continue
		}
		fmt.Println()
	}

	return nil
}

// updateBrew updates Homebrew packages.
func updateBrew(ctx context.Context, strategy string, dryRun bool) error {
	// Check if brew is installed
	cmd := exec.CommandContext(ctx, "brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew is not installed or not in PATH")
	}

	fmt.Println("🍺 Updating Homebrew...")

	// Update brew itself
	if !dryRun {
		cmd = exec.CommandContext(ctx, "brew", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update brew: %w", err)
		}
	} else {
		fmt.Println("Would run: brew update")
	}

	// Upgrade packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd = exec.CommandContext(ctx, "brew", "upgrade")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to upgrade brew packages: %w", err)
			}
		} else {
			fmt.Println("Would run: brew upgrade")
		}
	}

	// Cleanup old versions
	if !dryRun {
		cmd = exec.CommandContext(ctx, "brew", "cleanup")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: cleanup failed: %v\n", err)
		}
	} else {
		fmt.Println("Would run: brew cleanup")
	}

	return nil
}

func updateAsdf(ctx context.Context, strategy string, dryRun bool) error {
	// Check if asdf is installed
	cmd := exec.CommandContext(ctx, "asdf", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf is not installed or not in PATH")
	}

	fmt.Println("🔄 Updating asdf plugins...")

	// Update asdf plugins
	if !dryRun {
		cmd = exec.CommandContext(ctx, "asdf", "plugin", "update", "--all")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update asdf plugins: %w", err)
		}
	} else {
		fmt.Println("Would run: asdf plugin update --all")
	}

	// Get list of installed plugins
	cmd = exec.CommandContext(ctx, "asdf", "plugin", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list asdf plugins: %w", err)
	}

	plugins := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}
		fmt.Printf("Checking %s for updates...\n", plugin)

		// Install latest version based on strategy
		if strategy == "latest" || strategy == "stable" {
			if !dryRun {
				cmd = exec.CommandContext(ctx, "asdf", "install", plugin, "latest")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("Warning: failed to install latest %s: %v\n", plugin, err)
				}
			} else {
				fmt.Printf("Would run: asdf install %s latest\n", plugin)
			}
		}
	}

	return nil
}

func updateSdkman(ctx context.Context, strategy string, dryRun bool) error {
	// Check if SDKMAN is installed
	sdkmanDir := os.Getenv("SDKMAN_DIR")
	if sdkmanDir == "" {
		sdkmanDir = os.Getenv("HOME") + "/.sdkman"
	}

	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
		return fmt.Errorf("SDKMAN is not installed")
	}

	fmt.Println("☕ Updating SDKMAN...")

	// Update SDKMAN itself
	if !dryRun {
		cmd := exec.CommandContext(ctx, "bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk selfupdate")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to update SDKMAN: %v\n", err)
		}
	} else {
		fmt.Println("Would run: sdk selfupdate")
	}

	// Update candidates based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd := exec.CommandContext(ctx, "bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk update")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: failed to update SDKMAN candidates: %v\n", err)
			}
		} else {
			fmt.Println("Would run: sdk update")
		}
	}

	return nil
}

func updateApt(ctx context.Context, strategy string, dryRun bool) error {
	// Check if apt is available
	cmd := exec.CommandContext(ctx, "apt", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt is not available on this system")
	}

	fmt.Println("📦 Updating APT packages...")

	// Update package lists
	if !dryRun {
		cmd = exec.CommandContext(ctx, "sudo", "apt", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update apt package lists: %w", err)
		}
	} else {
		fmt.Println("Would run: sudo apt update")
	}

	// Upgrade packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd = exec.CommandContext(ctx, "sudo", "apt", "upgrade", "-y")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to upgrade apt packages: %w", err)
			}
		} else {
			fmt.Println("Would run: sudo apt upgrade -y")
		}
	}

	return nil
}

func updatePip(ctx context.Context, strategy string, dryRun bool) error {
	// Check if pip is installed
	pipCmd := findPipCommand(ctx)
	if pipCmd == "" {
		return fmt.Errorf("pip is not installed or not in PATH")
	}

	fmt.Println("🐍 Updating Python packages...")

	// Update pip itself
	if err := upgradePip(ctx, pipCmd, dryRun); err != nil {
		return err
	}

	// Update packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		return updateOutdatedPackages(ctx, pipCmd, dryRun)
	}

	return nil
}

// findPipCommand finds available pip command (python -m pip or pip3).
func findPipCommand(ctx context.Context) string {
	// Try python -m pip first
	cmd := exec.CommandContext(ctx, "python", "-m", "pip", "--version")
	if err := cmd.Run(); err == nil {
		return "python -m pip"
	}

	// Try pip3
	cmd = exec.CommandContext(ctx, "pip3", "--version")
	if err := cmd.Run(); err == nil {
		return "pip3"
	}

	return ""
}

// upgradePip upgrades pip itself.
func upgradePip(ctx context.Context, pipCmd string, dryRun bool) error {
	if !dryRun {
		args := strings.Split(pipCmd, " ")
		args = append(args, "install", "--upgrade", "pip")
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to upgrade pip: %w", err)
		}
	} else {
		fmt.Printf("Would run: %s install --upgrade pip\n", pipCmd)
	}
	return nil
}

// updateOutdatedPackages updates all outdated packages.
func updateOutdatedPackages(ctx context.Context, pipCmd string, dryRun bool) error {
	fmt.Println("Checking for outdated packages...")

	// Get list of outdated packages
	args := strings.Split(pipCmd, " ")
	args = append(args, "list", "--outdated", "--format=freeze")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Warning: failed to list outdated packages: %v\n", err)
		return nil
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(packages) == 0 || (len(packages) == 1 && packages[0] == "") {
		fmt.Println("All packages are up to date")
		return nil
	}

	fmt.Println("Found outdated packages")
	if dryRun {
		fmt.Println("Would update all outdated packages")
		return nil
	}

	// Update each package
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}
		// Extract package name (before ==)
		parts := strings.Split(pkg, "==")
		if len(parts) > 0 {
			pkgName := parts[0]
			fmt.Printf("Updating %s...\n", pkgName)
			args := strings.Split(pipCmd, " ")
			args = append(args, "install", "--upgrade", pkgName)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: failed to update %s: %v\n", pkgName, err)
			}
		}
	}

	return nil
}

func updateNpm(ctx context.Context, strategy string, dryRun bool) error {
	// Check if npm is installed
	cmd := exec.CommandContext(ctx, "npm", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm is not installed or not in PATH")
	}

	fmt.Println("📦 Updating Node.js packages...")

	// Update npm itself
	if !dryRun {
		cmd = exec.CommandContext(ctx, "npm", "install", "-g", "npm@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to upgrade npm: %w", err)
		}
	} else {
		fmt.Println("Would run: npm install -g npm@latest")
	}

	// Update global packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		return updateGlobalNpmPackages(ctx, dryRun)
	}

	return nil
}

// updateGlobalNpmPackages updates globally installed npm packages.
func updateGlobalNpmPackages(ctx context.Context, dryRun bool) error {
	fmt.Println("Checking for outdated global packages...")

	// Get list of outdated global packages
	cmd := exec.CommandContext(ctx, "npm", "outdated", "-g", "--json")
	output, err := cmd.Output()
	
	// npm outdated returns exit code 1 when packages are outdated
	if err != nil && len(output) == 0 {
		fmt.Printf("Warning: failed to list outdated packages: %v\n", err)
		return nil
	}

	// Check if there are outdated packages
	if len(output) == 0 || string(output) == "{}\n" || string(output) == "{}" {
		fmt.Println("All global packages are up to date")
		return nil
	}

	fmt.Println("Found outdated global packages")
	if dryRun {
		// Show what would be updated
		cmd = exec.CommandContext(ctx, "npm", "outdated", "-g")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // Ignore error as npm outdated returns 1 when packages are outdated
		fmt.Println("\nWould update all outdated global packages")
		return nil
	}

	// Update all global packages
	fmt.Println("Updating global packages...")
	cmd = exec.CommandContext(ctx, "npm", "update", "-g")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: some packages may have failed to update: %v\n", err)
	}

	return nil
}

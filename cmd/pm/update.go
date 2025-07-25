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

// updateBrew updates Homebrew packages
func updateBrew(ctx context.Context, strategy string, dryRun bool) error {
	// Check if brew is installed
	cmd := exec.Command("brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew is not installed or not in PATH")
	}

	fmt.Println("🍺 Updating Homebrew...")

	// Update brew itself
	if !dryRun {
		cmd = exec.Command("brew", "update")
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
			cmd = exec.Command("brew", "upgrade")
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
		cmd = exec.Command("brew", "cleanup")
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
	cmd := exec.Command("asdf", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf is not installed or not in PATH")
	}

	fmt.Println("🔄 Updating asdf plugins...")

	// Update asdf plugins
	if !dryRun {
		cmd = exec.Command("asdf", "plugin", "update", "--all")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update asdf plugins: %w", err)
		}
	} else {
		fmt.Println("Would run: asdf plugin update --all")
	}

	// Get list of installed plugins
	cmd = exec.Command("asdf", "plugin", "list")
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
				cmd = exec.Command("asdf", "install", plugin, "latest")
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
		cmd := exec.Command("bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk selfupdate")
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
			cmd := exec.Command("bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk update")
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
	cmd := exec.Command("apt", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt is not available on this system")
	}

	fmt.Println("📦 Updating APT packages...")

	// Update package lists
	if !dryRun {
		cmd = exec.Command("sudo", "apt", "update")
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
			cmd = exec.Command("sudo", "apt", "upgrade", "-y")
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
	fmt.Println("Would update Python packages...")
	// TODO: Implement based on configuration
	return nil
}

func updateNpm(ctx context.Context, strategy string, dryRun bool) error {
	fmt.Println("Would update npm packages...")
	// TODO: Implement based on configuration
	return nil
}

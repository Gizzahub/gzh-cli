// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-manager-go/internal/logger"
	"github.com/Gizzahub/gzh-manager-go/internal/pm/bootstrap"
)

func newBootstrapCmd(ctx context.Context) *cobra.Command {
	var (
		check      bool
		install    string
		force      bool
		jsonOutput bool
		verbose    bool
		dryRun     bool
		skipConfig bool
		timeout    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Install and configure package managers",
		Long: `Install missing package managers and ensure they are properly configured.

This command can automatically install and configure package managers like
brew, asdf, nvm, rbenv, pyenv, and sdkman with proper dependency resolution.

Examples:
  # Check which package managers need installation
  gz pm bootstrap --check

  # Check with JSON output
  gz pm bootstrap --check --json

  # Install all missing package managers
  gz pm bootstrap --install

  # Install specific package managers
  gz pm bootstrap --install brew,nvm,rbenv

  # Force reinstall with verbose output
  gz pm bootstrap --install --force --verbose

  # Dry run to see what would be installed
  gz pm bootstrap --install --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBootstrapCommand(ctx, bootstrapOptions{
				check:      check,
				install:    install,
				force:      force,
				jsonOutput: jsonOutput,
				verbose:    verbose,
				dryRun:     dryRun,
				skipConfig: skipConfig,
				timeout:    timeout,
			})
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Check which managers need installation")
	cmd.Flags().StringVar(&install, "install", "", "Package managers to install (comma-separated, empty = all missing)")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall even if already installed")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without actually installing")
	cmd.Flags().BoolVar(&skipConfig, "skip-config", false, "Skip post-install configuration")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "Timeout for installation operations")

	return cmd
}

type bootstrapOptions struct {
	check      bool
	install    string
	force      bool
	jsonOutput bool
	verbose    bool
	dryRun     bool
	skipConfig bool
	timeout    time.Duration
}

func runBootstrapCommand(ctx context.Context, opts bootstrapOptions) error {
	// Create logger
	logger := logger.NewSimpleLogger("bootstrap")

	// Create bootstrap manager
	manager := bootstrap.NewBootstrapManager(logger)

	if opts.check {
		return runBootstrapCheck(ctx, manager, opts)
	}

	return runBootstrapInstall(ctx, manager, opts)
}

func runBootstrapCheck(ctx context.Context, manager *bootstrap.BootstrapManager, opts bootstrapOptions) error {
	fmt.Println("🔍 Checking package manager installations...")

	report, err := manager.CheckAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to check package managers: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Print(manager.FormatReport(report, opts.verbose))
	return nil
}

func runBootstrapInstall(ctx context.Context, manager *bootstrap.BootstrapManager, opts bootstrapOptions) error {
	var managerNames []string

	if opts.install != "" {
		// Parse specified managers
		managerNames = strings.Split(opts.install, ",")
		for i, name := range managerNames {
			managerNames[i] = strings.TrimSpace(name)
		}
	}

	// Validate manager names
	availableManagers := manager.GetAvailableManagers()
	if len(managerNames) > 0 {
		for _, name := range managerNames {
			found := false
			for _, available := range availableManagers {
				if name == available {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("unknown package manager: %s. Available: %s",
					name, strings.Join(availableManagers, ", "))
			}
		}
	}

	// Show installation plan
	if len(managerNames) == 0 {
		fmt.Println("📦 Installing all missing package managers...")
	} else {
		fmt.Printf("📦 Installing package managers: %s\n", strings.Join(managerNames, ", "))
	}

	if opts.dryRun {
		fmt.Println("🧪 DRY RUN MODE - No actual installation will be performed")
	}

	// Show installation order
	installOrder, err := manager.GetInstallationOrder(managerNames)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if len(installOrder) > 0 {
		fmt.Printf("📋 Installation order: %s\n", strings.Join(installOrder, " → "))
	}

	// Create bootstrap options
	bootstrapOpts := bootstrap.BootstrapOptions{
		Managers:          managerNames,
		Force:             opts.force,
		SkipConfiguration: opts.skipConfig,
		DryRun:            opts.dryRun,
		Timeout:           bootstrap.Duration{Duration: opts.timeout},
		Verbose:           opts.verbose,
	}

	// Install managers
	fmt.Println("\n🚀 Starting installation...")
	report, err := manager.InstallManagers(ctx, managerNames, bootstrapOpts)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Show results
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Println("\n" + manager.FormatReport(report, opts.verbose))

	// Summary
	if report.Summary.Failed > 0 {
		fmt.Printf("\n⚠️  %d package managers had issues during installation\n", report.Summary.Failed)
		return fmt.Errorf("some installations failed")
	}

	if opts.dryRun {
		fmt.Println("\n✅ Dry run completed successfully")
	} else {
		fmt.Printf("\n✅ Bootstrap completed in %v\n", report.Duration)
		fmt.Println("\n💡 You may need to restart your shell or source your profile to use the new package managers")
	}

	return nil
}

func newUpgradeManagersCmd(ctx context.Context) *cobra.Command {
	var (
		all     bool
		manager string
		check   bool
	)

	cmd := &cobra.Command{
		Use:   "upgrade-managers",
		Short: "Upgrade package managers themselves",
		Long:  `Upgrade the package manager tools to their latest versions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case check:
				fmt.Println("Checking for package manager updates...")
			case all:
				fmt.Println("Upgrading all package managers...")
			case manager != "":
				fmt.Printf("Upgrading %s...\n", manager)
			}
			return fmt.Errorf("upgrade-managers command not yet implemented")
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Upgrade all package managers")
	cmd.Flags().StringVar(&manager, "manager", "", "Specific manager to upgrade")
	cmd.Flags().BoolVar(&check, "check", false, "Check available upgrades")

	return cmd
}

func newSyncVersionsCmd(ctx context.Context) *cobra.Command {
	var (
		check bool
		fix   bool
	)

	cmd := &cobra.Command{
		Use:   "sync-versions",
		Short: "Synchronize version manager and package manager versions",
		Long: `Ensure version managers (like nvm, rbenv) are synchronized with their
package managers (npm, gem).

Examples:
  # Check for version mismatches
  gz pm sync-versions --check

  # Fix version mismatches
  gz pm sync-versions --fix`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if check {
				fmt.Println("Checking version synchronization...")
			} else if fix {
				fmt.Println("Fixing version mismatches...")
			}
			return fmt.Errorf("sync-versions command not yet implemented")
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Check for version mismatches")
	cmd.Flags().BoolVar(&fix, "fix", false, "Fix version mismatches")

	return cmd
}

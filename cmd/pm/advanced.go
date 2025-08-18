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

	"github.com/Gizzahub/gzh-cli/internal/logger"
	"github.com/Gizzahub/gzh-cli/internal/pm/bootstrap"
	"github.com/Gizzahub/gzh-cli/internal/pm/sync"
	"github.com/Gizzahub/gzh-cli/internal/pm/upgrade"
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
		all        bool
		manager    string
		check      bool
		backup     bool
		force      bool
		jsonOutput bool
		timeout    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "upgrade-managers",
		Short: "Upgrade package managers themselves",
		Long: `Upgrade the package manager tools to their latest versions.

This command can check for updates and upgrade package managers like
brew, asdf, nvm, rbenv, pyenv, and sdkman to their latest versions.

Examples:
  # Check which package managers have updates available
  gz pm upgrade-managers --check

  # Check with JSON output
  gz pm upgrade-managers --check --json

  # Upgrade all package managers
  gz pm upgrade-managers --all

  # Upgrade specific package managers
  gz pm upgrade-managers --manager brew,nvm

  # Upgrade with backup enabled
  gz pm upgrade-managers --all --backup

  # Force upgrade even if no updates detected
  gz pm upgrade-managers --all --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgradeManagersCommand(ctx, upgradeManagersOptions{
				all:        all,
				manager:    manager,
				check:      check,
				backup:     backup,
				force:      force,
				jsonOutput: jsonOutput,
				timeout:    timeout,
			})
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Upgrade all package managers")
	cmd.Flags().StringVar(&manager, "manager", "", "Specific managers to upgrade (comma-separated)")
	cmd.Flags().BoolVar(&check, "check", false, "Check available upgrades without installing")
	cmd.Flags().BoolVar(&backup, "backup", false, "Create backup before upgrading")
	cmd.Flags().BoolVar(&force, "force", false, "Force upgrade even if no updates detected")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Timeout for upgrade operations")

	return cmd
}

func newSyncVersionsCmd(ctx context.Context) *cobra.Command {
	var (
		check    bool
		fix      bool
		pair     string
		strategy string
		backup   bool
		jsonOut  bool
		verbose  bool
	)

	cmd := &cobra.Command{
		Use:   "sync-versions",
		Short: "Synchronize version manager and package manager versions",
		Long: `Ensure version managers (like nvm, rbenv) are synchronized with their
package managers (npm, gem).

This command checks for version mismatches between version managers and their
corresponding package managers, and can automatically fix synchronization issues.

Examples:
  # Check for version mismatches
  gz pm sync-versions --check

  # Check with JSON output
  gz pm sync-versions --check --json

  # Fix version mismatches with backup
  gz pm sync-versions --fix --backup

  # Check specific manager pair
  gz pm sync-versions --check --pair nvm-npm

  # Fix with specific strategy
  gz pm sync-versions --fix --strategy vm_priority
  gz pm sync-versions --fix --strategy pm_priority
  gz pm sync-versions --fix --strategy latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncVersionsCommand(ctx, syncVersionsOptions{
				check:    check,
				fix:      fix,
				pair:     pair,
				strategy: strategy,
				backup:   backup,
				jsonOut:  jsonOut,
				verbose:  verbose,
			})
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Check for version mismatches")
	cmd.Flags().BoolVar(&fix, "fix", false, "Fix version mismatches")
	cmd.Flags().StringVar(&pair, "pair", "", "Specific manager pair to check/fix (e.g., nvm-npm, rbenv-gem)")
	cmd.Flags().StringVar(&strategy, "strategy", "vm_priority", "Synchronization strategy (vm_priority, pm_priority, latest)")
	cmd.Flags().BoolVar(&backup, "backup", false, "Create backup before fixing")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	return cmd
}

type upgradeManagersOptions struct {
	all        bool
	manager    string
	check      bool
	backup     bool
	force      bool
	jsonOutput bool
	timeout    time.Duration
}

func runUpgradeManagersCommand(ctx context.Context, opts upgradeManagersOptions) error {
	// Create logger
	logger := logger.NewSimpleLogger("upgrade-managers")

	// Create upgrade coordinator
	coordinator := upgrade.NewUpgradeCoordinator(logger, "/tmp/gzh-upgrades")

	if opts.check {
		return runUpgradeCheck(ctx, coordinator, opts)
	}

	return runUpgradeInstall(ctx, coordinator, opts)
}

func runUpgradeCheck(ctx context.Context, coordinator *upgrade.UpgradeCoordinator, opts upgradeManagersOptions) error {
	fmt.Println("🔍 Checking package manager upgrades...")

	var report *upgrade.UpgradeReport
	var err error

	if opts.all || opts.manager == "" {
		report, err = coordinator.CheckAll(ctx)
	} else {
		// Parse specified managers
		managerNames := strings.Split(opts.manager, ",")
		for i, name := range managerNames {
			managerNames[i] = strings.TrimSpace(name)
		}
		report, err = coordinator.CheckManagers(ctx, managerNames)
	}

	if err != nil {
		return fmt.Errorf("failed to check package manager upgrades: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Print(coordinator.FormatReport(report, true))

	// Summary message
	if report.UpdatesNeeded > 0 {
		fmt.Printf("\n💡 %d package managers have updates available\n", report.UpdatesNeeded)
		fmt.Println("Run with --all or --manager <names> to upgrade them")
	} else {
		fmt.Println("\n✅ All package managers are up to date")
	}

	return nil
}

func runUpgradeInstall(ctx context.Context, coordinator *upgrade.UpgradeCoordinator, opts upgradeManagersOptions) error {
	var managerNames []string

	if opts.all {
		fmt.Println("📦 Upgrading all package managers...")
	} else if opts.manager != "" {
		// Parse specified managers
		managerNames = strings.Split(opts.manager, ",")
		for i, name := range managerNames {
			managerNames[i] = strings.TrimSpace(name)
		}
		fmt.Printf("📦 Upgrading package managers: %s\n", strings.Join(managerNames, ", "))
	} else {
		return fmt.Errorf("must specify --all or --manager <names> to upgrade")
	}

	// Validate manager names
	if len(managerNames) > 0 {
		availableManagers := coordinator.GetAvailableManagers()
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

	// Create upgrade options
	upgradeOpts := upgrade.UpgradeOptions{
		Force:          opts.force,
		BackupEnabled:  opts.backup,
		SkipValidation: false,
		Timeout:        opts.timeout,
	}

	// Perform upgrades
	fmt.Println("\n🚀 Starting upgrades...")
	var report *upgrade.UpgradeReport
	var err error

	if opts.all {
		report, err = coordinator.UpgradeAll(ctx, upgradeOpts)
	} else {
		report, err = coordinator.UpgradeManagers(ctx, managerNames, upgradeOpts)
	}

	if err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	// Show results
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Println("\n" + coordinator.FormatReport(report, true))

	// Summary
	failureCount := 0
	successCount := 0
	for _, status := range report.Managers {
		if status.UpdateAvailable {
			failureCount++
		} else {
			successCount++
		}
	}

	if failureCount > 0 {
		fmt.Printf("\n⚠️  %d package managers had issues during upgrade\n", failureCount)
	}

	if successCount > 0 {
		fmt.Printf("\n✅ Successfully upgraded %d package managers\n", successCount)
		fmt.Println("\n💡 You may need to restart your shell or source your profile to use the updated tools")
	}

	return nil
}

type syncVersionsOptions struct {
	check    bool
	fix      bool
	pair     string
	strategy string
	backup   bool
	jsonOut  bool
	verbose  bool
}

func runSyncVersionsCommand(ctx context.Context, opts syncVersionsOptions) error {
	// Create logger
	logger := logger.NewSimpleLogger("sync-versions")

	// Create sync manager
	syncManager := sync.NewSyncManager(logger)

	if opts.check {
		return runSyncVersionsCheck(ctx, syncManager, opts)
	}

	if opts.fix {
		return runSyncVersionsFix(ctx, syncManager, opts)
	}

	return fmt.Errorf("must specify --check or --fix")
}

func runSyncVersionsCheck(ctx context.Context, syncManager *sync.SyncManager, opts syncVersionsOptions) error {
	fmt.Println("🔍 Checking version synchronization...")

	var report *sync.SyncReport
	var err error

	if opts.pair != "" {
		// Check specific pair
		pairs := strings.Split(opts.pair, ",")
		for i, pair := range pairs {
			pairs[i] = strings.TrimSpace(pair)
		}
		report, err = syncManager.CheckPairs(ctx, pairs)
	} else {
		// Check all pairs
		report, err = syncManager.CheckAll(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to check version synchronization: %w", err)
	}

	if opts.jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Print(syncManager.FormatReport(report, opts.verbose))

	// Summary message
	if report.OutOfSyncCount > 0 {
		fmt.Printf("💡 %d manager pairs are out of sync\n", report.OutOfSyncCount)
		fmt.Println("Run with --fix to synchronize them")

		if !opts.verbose {
			fmt.Println("\nSynchronization strategies:")
			fmt.Println("  --strategy vm_priority    Update package managers to match version managers")
			fmt.Println("  --strategy pm_priority    Update version managers to match package managers")
			fmt.Println("  --strategy latest         Update both to latest compatible versions")
		}
	} else {
		fmt.Println("✅ All version manager pairs are synchronized")
	}

	return nil
}

func runSyncVersionsFix(ctx context.Context, syncManager *sync.SyncManager, opts syncVersionsOptions) error {
	var pairs []string

	if opts.pair != "" {
		// Parse specified pairs
		pairs = strings.Split(opts.pair, ",")
		for i, pair := range pairs {
			pairs[i] = strings.TrimSpace(pair)
		}
		fmt.Printf("🔧 Fixing synchronization for pairs: %s\n", strings.Join(pairs, ", "))
	} else {
		// Get all available pairs
		pairs = syncManager.GetAvailablePairs()
		fmt.Println("🔧 Fixing synchronization for all manager pairs...")
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"vm_priority": true,
		"pm_priority": true,
		"latest":      true,
	}

	if !validStrategies[opts.strategy] {
		return fmt.Errorf("invalid strategy: %s. Valid strategies: vm_priority, pm_priority, latest", opts.strategy)
	}

	// Create sync policy
	policy := sync.SyncPolicy{
		Strategy:      opts.strategy,
		AutoFix:       true,
		BackupEnabled: opts.backup,
		PromptUser:    false, // CLI mode, no prompting
	}

	fmt.Printf("Using strategy: %s\n", opts.strategy)
	if opts.backup {
		fmt.Println("Backup enabled")
	}

	// Perform synchronization
	fmt.Println("\n🚀 Starting synchronization...")
	report, err := syncManager.FixSynchronization(ctx, pairs, policy)
	if err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	// Show results
	if opts.jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Println("\n" + syncManager.FormatReport(report, opts.verbose))

	// Summary
	if report.OutOfSyncCount > 0 {
		fmt.Printf("⚠️  %d manager pairs had issues during synchronization\n", report.OutOfSyncCount)
	}

	if report.InSyncCount > 0 {
		fmt.Printf("✅ Successfully synchronized %d manager pairs\n", report.InSyncCount)
		fmt.Println("\n💡 You may need to restart your shell or source your profile to use the synchronized tools")
	}

	return nil
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package alwayslatest

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type alwaysLatestBrewOptions struct {
	strategy    string
	packages    []string
	dryRun      bool
	updateBrew  bool
	configFile  string
	casks       bool
	taps        bool
	interactive bool
	cleanup     bool
	upgradeAll  bool
}

func defaultAlwaysLatestBrewOptions() *alwaysLatestBrewOptions {
	return &alwaysLatestBrewOptions{
		strategy:    "minor",
		packages:    []string{},
		dryRun:      false,
		updateBrew:  true,
		casks:       false,
		taps:        false,
		interactive: true,
		cleanup:     false,
		upgradeAll:  false,
	}
}

func newAlwaysLatestBrewCmd(_ context.Context) *cobra.Command {
	o := defaultAlwaysLatestBrewOptions()

	cmd := &cobra.Command{
		Use:   "brew",
		Short: "Update Homebrew and its managed packages to latest versions",
		Long: `Update Homebrew package manager and its managed software packages.

This command helps keep your development environment current by:
- Updating Homebrew itself to the latest version
- Upgrading outdated packages to their latest versions
- Supporting both formulae (packages) and casks (applications)
- Managing Homebrew taps (third-party repositories)
- Optionally cleaning up old package versions

Update strategies:
  minor: Update to latest available version (Homebrew handles compatibility)
  major: Same as minor (Homebrew doesn't use semantic versioning like other tools)

Examples:
  # Update all outdated packages (default)
  gz always-latest brew

  # Update specific packages only
  gz always-latest brew --packages node,python,git

  # Include cask applications in updates
  gz always-latest brew --casks

  # Update and cleanup old versions
  gz always-latest brew --cleanup

  # Dry run to see what would be updated
  gz always-latest brew --dry-run

  # Update all packages without prompting
  gz always-latest brew --upgrade-all`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major (both work the same for brew)")
	cmd.Flags().StringSliceVar(&o.packages, "packages", o.packages, "Specific packages to update (comma-separated)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updateBrew, "update-brew", o.updateBrew, "Update Homebrew itself")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.casks, "casks", o.casks, "Include cask applications in updates")
	cmd.Flags().BoolVar(&o.taps, "taps", o.taps, "Update taps (third-party repositories)")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")
	cmd.Flags().BoolVar(&o.cleanup, "cleanup", o.cleanup, "Clean up old versions after update")
	cmd.Flags().BoolVar(&o.upgradeAll, "upgrade-all", o.upgradeAll, "Upgrade all packages without individual confirmation")

	return cmd
}

func (o *alwaysLatestBrewOptions) run(_ *cobra.Command, _ []string) error {
	// Check if brew is installed
	if !o.isBrewInstalled() {
		return fmt.Errorf("homebrew is not installed or not in PATH")
	}

	fmt.Println("🍺 Starting Homebrew update process...")

	ctx := context.Background()

	// Update Homebrew and taps
	if err := o.performUpdates(ctx); err != nil {
		return err
	}

	// Get and filter outdated packages
	outdatedPackages, outdatedCasks, err := o.getFilteredOutdatedPackages(ctx)
	if err != nil {
		return err
	}

	totalOutdated := len(outdatedPackages) + len(outdatedCasks)
	if totalOutdated == 0 {
		fmt.Println("✅ All packages are up to date")
		return nil
	}

	fmt.Printf("📦 Found %d outdated package(s)\n", totalOutdated)

	// Update packages and casks
	updatedCount := o.updateAllPackages(ctx, outdatedPackages, outdatedCasks)

	// Cleanup if requested
	if o.cleanup && !o.dryRun {
		if err := o.cleanupBrew(); err != nil {
			fmt.Printf("⚠️  Failed to cleanup: %v\n", err)
		}
	}

	// Summary
	o.printSummary(updatedCount)

	return nil
}

// performUpdates updates Homebrew itself and taps if requested.
func (o *alwaysLatestBrewOptions) performUpdates(ctx context.Context) error {
	// Update Homebrew itself
	if o.updateBrew {
		if err := o.updateHomebrew(ctx); err != nil {
			return fmt.Errorf("failed to update Homebrew: %w", err)
		}
	}

	// Update taps if requested
	if o.taps {
		if err := o.updateTaps(ctx); err != nil {
			fmt.Printf("⚠️  Failed to update taps: %v\n", err)
		}
	}

	return nil
}

// getFilteredOutdatedPackages gets outdated packages and casks, then filters them.
func (o *alwaysLatestBrewOptions) getFilteredOutdatedPackages(ctx context.Context) (packages, casks []string, err error) {
	// Get outdated packages
	outdatedPackages, err := o.getOutdatedPackages(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get outdated packages: %w", err)
	}

	// Get outdated casks if requested
	var outdatedCasks []string
	if o.casks {
		outdatedCasks, err = o.getOutdatedCasks(ctx)
		if err != nil {
			fmt.Printf("⚠️  Failed to get outdated casks: %v\n", err)
		}
	}

	// Filter packages if specific packages are requested
	if len(o.packages) > 0 {
		outdatedPackages = o.filterPackages(outdatedPackages, o.packages)
		if o.casks {
			outdatedCasks = o.filterPackages(outdatedCasks, o.packages)
		}
	}

	return outdatedPackages, outdatedCasks, nil
}

// updateAllPackages updates both formulae and casks.
func (o *alwaysLatestBrewOptions) updateAllPackages(ctx context.Context, outdatedPackages, outdatedCasks []string) int {
	updatedCount := 0

	if len(outdatedPackages) > 0 {
		fmt.Printf("📦 Processing %d outdated formulae...\n", len(outdatedPackages))
		updatedCount += o.updatePackageList(ctx, outdatedPackages, false)
	}

	if len(outdatedCasks) > 0 {
		fmt.Printf("🍺 Processing %d outdated casks...\n", len(outdatedCasks))
		updatedCount += o.updatePackageList(ctx, outdatedCasks, true)
	}

	return updatedCount
}

// updatePackageList updates a list of packages (formulae or casks).
func (o *alwaysLatestBrewOptions) updatePackageList(ctx context.Context, packages []string, isCask bool) int {
	updatedCount := 0
	for _, pkg := range packages {
		updated, err := o.updatePackage(ctx, pkg, isCask)
		if err != nil {
			fmt.Printf("⚠️  Failed to update %s: %v\n", pkg, err)
			continue
		}

		if updated {
			updatedCount++
		}
	}
	return updatedCount
}

// printSummary prints the final summary of the update process.
func (o *alwaysLatestBrewOptions) printSummary(updatedCount int) {
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d packages would be updated\n", updatedCount)
	} else {
		fmt.Printf("✅ Update completed. %d packages were updated\n", updatedCount)
	}
}

func (o *alwaysLatestBrewOptions) isBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func (o *alwaysLatestBrewOptions) updateHomebrew(ctx context.Context) error {
	fmt.Println("🔄 Updating Homebrew...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: brew update")
		return nil
	}

	cmd := exec.CommandContext(ctx, "brew", "update")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew update failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Homebrew updated successfully")

	return nil
}

func (o *alwaysLatestBrewOptions) updateTaps(ctx context.Context) error {
	fmt.Println("🔄 Updating taps...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would update all taps")
		return nil
	}

	// Get list of taps
	cmd := exec.CommandContext(ctx, "brew", "tap")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list taps: %w", err)
	}

	taps := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, tap := range taps {
		tap = strings.TrimSpace(tap)
		if tap == "" {
			continue
		}

		fmt.Printf("   Updating tap: %s\n", tap)

		updateCmd := exec.CommandContext(ctx, "brew", "tap", tap)
		if updateOutput, updateErr := updateCmd.CombinedOutput(); updateErr != nil {
			fmt.Printf("   ⚠️  Failed to update tap %s: %v\n%s", tap, updateErr, string(updateOutput))
		}
	}

	fmt.Println("✅ Taps updated")

	return nil
}

func (o *alwaysLatestBrewOptions) getOutdatedPackages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "brew", "outdated", "--formula")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated packages: %w", err)
	}

	var packages []string

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse package name from output (format: "package (current) < latest")
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}

	return packages, nil
}

func (o *alwaysLatestBrewOptions) getOutdatedCasks(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "brew", "outdated", "--cask")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated casks: %w", err)
	}

	var casks []string

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse cask name from output
		parts := strings.Fields(line)
		if len(parts) > 0 {
			casks = append(casks, parts[0])
		}
	}

	return casks, nil
}

func (o *alwaysLatestBrewOptions) filterPackages(allPackages, requestedPackages []string) []string {
	packageSet := make(map[string]bool)
	for _, pkg := range requestedPackages {
		packageSet[strings.TrimSpace(pkg)] = true
	}

	var filtered []string

	for _, pkg := range allPackages {
		if packageSet[pkg] {
			filtered = append(filtered, pkg)
		}
	}

	return filtered
}

func (o *alwaysLatestBrewOptions) updatePackage(ctx context.Context, packageName string, isCask bool) (bool, error) {
	packageType := "formula"
	if isCask {
		packageType = "cask"
	}

	fmt.Printf("🔍 Checking %s (%s)...\n", packageName, packageType)

	if o.dryRun {
		if isCask {
			fmt.Printf("   [DRY RUN] Would run: brew upgrade --cask %s\n", packageName)
		} else {
			fmt.Printf("   [DRY RUN] Would run: brew upgrade %s\n", packageName)
		}

		return true, nil
	}

	// Confirm update in interactive mode
	if o.interactive && !o.upgradeAll && !o.confirmUpdate(packageName, packageType) {
		fmt.Printf("   Skipping %s update\n", packageName)
		return false, nil
	}

	// Upgrade the package
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.CommandContext(ctx, "brew", "upgrade", "--cask", packageName)
	} else {
		cmd = exec.CommandContext(ctx, "brew", "upgrade", packageName)
	}

	fmt.Printf("   Upgrading %s...\n", packageName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("upgrade failed: %w\n%s", err, string(output))
	}

	fmt.Printf("✅ Updated %s\n", packageName)

	return true, nil
}

func (o *alwaysLatestBrewOptions) confirmUpdate(packageName, packageType string) bool {
	fmt.Printf("   Update %s (%s)? (y/N): ", packageName, packageType)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// If reading fails, default to no
		return false
	}

	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestBrewOptions) cleanupBrew() error {
	fmt.Println("🧹 Cleaning up old versions...")

	cmd := exec.Command("brew", "cleanup")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Cleanup completed")

	return nil
}

func (o *alwaysLatestBrewOptions) extractVersionNumber(versionString string) string {
	// Extract version number using regex
	re := regexp.MustCompile(`(\d+\.[\d.]+)`)

	matches := re.FindStringSubmatch(versionString)
	if len(matches) > 1 {
		return matches[1]
	}

	return versionString
}

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

type alwaysLatestAptOptions struct {
	strategy    string
	packages    []string
	dryRun      bool
	updateApt   bool
	configFile  string
	interactive bool
	cleanup     bool
	upgradeAll  bool
	autoRemove  bool
	fullUpgrade bool
	fixBroken   bool
}

func defaultAlwaysLatestAptOptions() *alwaysLatestAptOptions {
	return &alwaysLatestAptOptions{
		strategy:    "minor",
		packages:    []string{},
		dryRun:      false,
		updateApt:   true,
		interactive: true,
		cleanup:     false,
		upgradeAll:  false,
		autoRemove:  false,
		fullUpgrade: false,
		fixBroken:   false,
	}
}

func newAlwaysLatestAptCmd(ctx context.Context) *cobra.Command {
	o := defaultAlwaysLatestAptOptions()

	cmd := &cobra.Command{
		Use:   "apt",
		Short: "Update APT and its managed packages to latest versions",
		Long: `Update APT package manager and its managed software packages on Debian/Ubuntu systems.

This command helps keep your development environment current by:
- Updating APT package lists to get information on newest versions
- Upgrading outdated packages to their latest versions
- Supporting both individual package updates and full system upgrades
- Optionally cleaning up obsolete packages and dependencies
- Fixing broken package dependencies when needed

APT (Advanced Package Tool) is the package management system used by
Debian, Ubuntu, and other Debian-based Linux distributions.

Update strategies:
  minor: Update to latest available version (APT handles compatibility)
  major: Same as minor (APT doesn't use semantic versioning like other tools)

Examples:
  # Update all upgradeable packages (default)
  gz always-latest apt
  
  # Update specific packages only
  gz always-latest apt --packages curl,git,vim
  
  # Perform full system upgrade
  gz always-latest apt --full-upgrade
  
  # Update and auto-remove unused packages
  gz always-latest apt --auto-remove
  
  # Dry run to see what would be updated
  gz always-latest apt --dry-run
  
  # Update all packages without prompting
  gz always-latest apt --upgrade-all
  
  # Fix broken packages before updating
  gz always-latest apt --fix-broken`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major (both work the same for apt)")
	cmd.Flags().StringSliceVar(&o.packages, "packages", o.packages, "Specific packages to update (comma-separated)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updateApt, "update-apt", o.updateApt, "Update APT package lists")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")
	cmd.Flags().BoolVar(&o.cleanup, "cleanup", o.cleanup, "Clean up obsolete packages after update")
	cmd.Flags().BoolVar(&o.upgradeAll, "upgrade-all", o.upgradeAll, "Upgrade all packages without individual confirmation")
	cmd.Flags().BoolVar(&o.autoRemove, "auto-remove", o.autoRemove, "Remove packages that were automatically installed to satisfy dependencies")
	cmd.Flags().BoolVar(&o.fullUpgrade, "full-upgrade", o.fullUpgrade, "Perform full upgrade (may remove packages)")
	cmd.Flags().BoolVar(&o.fixBroken, "fix-broken", o.fixBroken, "Fix broken package dependencies")

	return cmd
}

func (o *alwaysLatestAptOptions) run(_ *cobra.Command, args []string) error {
	// Check if APT is available
	if !o.isAptInstalled() {
		return fmt.Errorf("APT is not available on this system")
	}

	fmt.Println("📦 Starting APT update process...")

	// Fix broken packages if requested
	if o.fixBroken {
		if err := o.fixBrokenPackages(); err != nil {
			fmt.Printf("⚠️  Failed to fix broken packages: %v\n", err)
		}
	}

	// Update APT package lists
	if o.updateApt {
		if err := o.updatePackageLists(); err != nil {
			return fmt.Errorf("failed to update APT package lists: %w", err)
		}
	}

	// Handle full upgrade mode
	if o.fullUpgrade {
		return o.performFullUpgrade()
	}

	// Get upgradeable packages
	upgradeablePackages, err := o.getUpgradeablePackages()
	if err != nil {
		return fmt.Errorf("failed to get upgradeable packages: %w", err)
	}

	// Filter packages if specific packages are requested
	if len(o.packages) > 0 {
		upgradeablePackages = o.filterPackages(upgradeablePackages, o.packages)
	}

	if len(upgradeablePackages) == 0 {
		fmt.Println("✅ All packages are up to date")
		return o.performCleanupIfRequested()
	}

	fmt.Printf("📦 Found %d upgradeable package(s): %v\n", len(upgradeablePackages), upgradeablePackages)

	// Update packages
	updatedCount := 0
	for _, pkg := range upgradeablePackages {
		updated, err := o.updatePackage(pkg)
		if err != nil {
			fmt.Printf("⚠️  Failed to update %s: %v\n", pkg, err)
			continue
		}
		if updated {
			updatedCount++
		}
	}

	// Perform cleanup if requested
	if err := o.performCleanupIfRequested(); err != nil {
		fmt.Printf("⚠️  Failed to cleanup: %v\n", err)
	}

	// Summary
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d packages would be updated\n", updatedCount)
	} else {
		fmt.Printf("✅ Update completed. %d packages were updated\n", updatedCount)
	}

	return nil
}

func (o *alwaysLatestAptOptions) isAptInstalled() bool {
	_, err := exec.LookPath("apt-get")
	return err == nil
}

func (o *alwaysLatestAptOptions) fixBrokenPackages() error {
	fmt.Println("🔧 Fixing broken packages...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sudo apt-get install -f")
		return nil
	}

	cmd := exec.Command("sudo", "apt-get", "install", "-f", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fix broken packages failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Fixed broken packages")
	return nil
}

func (o *alwaysLatestAptOptions) updatePackageLists() error {
	fmt.Println("🔄 Updating APT package lists...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sudo apt-get update")
		return nil
	}

	cmd := exec.Command("sudo", "apt-get", "update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt update failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Package lists updated successfully")
	return nil
}

func (o *alwaysLatestAptOptions) performFullUpgrade() error {
	fmt.Println("🔄 Performing full system upgrade...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sudo apt-get full-upgrade")
		return nil
	}

	if o.interactive && !o.confirmFullUpgrade() {
		fmt.Println("   Full upgrade cancelled")
		return nil
	}

	cmd := exec.Command("sudo", "apt-get", "full-upgrade", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("full upgrade failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Full upgrade completed")
	return o.performCleanupIfRequested()
}

func (o *alwaysLatestAptOptions) getUpgradeablePackages() ([]string, error) {
	cmd := exec.Command("apt", "list", "--upgradable")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get upgradeable packages: %w", err)
	}

	var packages []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Listing...") || strings.HasPrefix(line, "WARNING") {
			continue
		}

		// Parse package name from output (format: "package/repo version [upgradable from: old_version]")
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packageParts := strings.Split(parts[0], "/")
			if len(packageParts) > 0 {
				packages = append(packages, packageParts[0])
			}
		}
	}

	return packages, nil
}

func (o *alwaysLatestAptOptions) filterPackages(allPackages, requestedPackages []string) []string {
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

func (o *alwaysLatestAptOptions) updatePackage(packageName string) (bool, error) {
	fmt.Printf("🔍 Checking %s...\n", packageName)

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would run: sudo apt-get install --only-upgrade %s\n", packageName)
		return true, nil
	}

	// Confirm update in interactive mode
	if o.interactive && !o.upgradeAll && !o.confirmUpdate(packageName) {
		fmt.Printf("   Skipping %s update\n", packageName)
		return false, nil
	}

	// Upgrade the package
	fmt.Printf("   Upgrading %s...\n", packageName)
	cmd := exec.Command("sudo", "apt-get", "install", "--only-upgrade", "-y", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("upgrade failed: %w\n%s", err, string(output))
	}

	fmt.Printf("✅ Updated %s\n", packageName)
	return true, nil
}

func (o *alwaysLatestAptOptions) confirmUpdate(packageName string) bool {
	fmt.Printf("   Update %s? (y/N): ", packageName)
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestAptOptions) confirmFullUpgrade() bool {
	fmt.Print("   Perform full system upgrade? This may remove packages (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestAptOptions) performCleanupIfRequested() error {
	if o.autoRemove && !o.dryRun {
		if err := o.autoRemovePackages(); err != nil {
			return err
		}
	}

	if o.cleanup && !o.dryRun {
		if err := o.cleanupPackages(); err != nil {
			return err
		}
	}

	return nil
}

func (o *alwaysLatestAptOptions) autoRemovePackages() error {
	fmt.Println("🧹 Auto-removing unused packages...")

	cmd := exec.Command("sudo", "apt-get", "autoremove", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("autoremove failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Auto-remove completed")
	return nil
}

func (o *alwaysLatestAptOptions) cleanupPackages() error {
	fmt.Println("🧹 Cleaning up package cache...")

	cmd := exec.Command("sudo", "apt-get", "autoclean")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("autoclean failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Package cache cleaned")
	return nil
}

func (o *alwaysLatestAptOptions) getPackageVersion(packageName string) (string, error) {
	cmd := exec.Command("dpkg", "-l", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from dpkg output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ii") { // installed package
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}

	return "", fmt.Errorf("package %s not found", packageName)
}

func (o *alwaysLatestAptOptions) extractVersionNumber(versionString string) string {
	// Extract version number using regex
	re := regexp.MustCompile(`(\d+\.[\d.]+)`)
	matches := re.FindStringSubmatch(versionString)
	if len(matches) > 1 {
		return matches[1]
	}
	return versionString
}

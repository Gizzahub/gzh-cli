package alwayslatest

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type alwaysLatestRbenvOptions struct {
	strategy      string
	versions      []string
	dryRun        bool
	updateRbenv   bool
	configFile    string
	global        bool
	interactive   bool
	updatePlugins bool
	rehash        bool
}

func defaultAlwaysLatestRbenvOptions() *alwaysLatestRbenvOptions {
	return &alwaysLatestRbenvOptions{
		strategy:      "minor",
		versions:      []string{},
		dryRun:        false,
		updateRbenv:   true,
		global:        false,
		interactive:   true,
		updatePlugins: true,
		rehash:        true,
	}
}

func newAlwaysLatestRbenvCmd(ctx context.Context) *cobra.Command {
	o := defaultAlwaysLatestRbenvOptions()

	cmd := &cobra.Command{
		Use:   "rbenv",
		Short: "Update rbenv and install latest Ruby versions",
		Long: `Update rbenv Ruby version manager and install latest Ruby versions.

This command helps keep your Ruby development environment current by:
- Updating rbenv itself to the latest version (via git or package manager)
- Installing latest Ruby versions available through ruby-build
- Optionally updating global or local Ruby versions
- Supporting both minor and major update strategies
- Rehashing rbenv shims after installations

rbenv is a Ruby version manager that lets you easily switch between
multiple versions of Ruby on the same machine.

Update strategies:
  minor: Update to latest patch/minor version within same major version (safer)
  major: Update to absolute latest version including major version changes

Examples:
  # Update rbenv and install latest Ruby versions
  gz always-latest rbenv
  
  # Install specific Ruby versions only
  gz always-latest rbenv --versions 3.1,3.2
  
  # Update with major version strategy (includes breaking changes)
  gz always-latest rbenv --strategy major
  
  # Dry run to see what would be updated
  gz always-latest rbenv --dry-run
  
  # Update and set as global version
  gz always-latest rbenv --global
  
  # Skip rbenv update, only install Ruby versions
  gz always-latest rbenv --update-rbenv=false`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major")
	cmd.Flags().StringSliceVar(&o.versions, "versions", o.versions, "Specific Ruby versions to install (comma-separated, e.g., 3.1,3.2)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updateRbenv, "update-rbenv", o.updateRbenv, "Update rbenv itself")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.global, "global", o.global, "Set latest version as global Ruby version")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")
	cmd.Flags().BoolVar(&o.updatePlugins, "update-plugins", o.updatePlugins, "Update rbenv plugins (like ruby-build)")
	cmd.Flags().BoolVar(&o.rehash, "rehash", o.rehash, "Rehash rbenv shims after installations")

	return cmd
}

func (o *alwaysLatestRbenvOptions) run(_ *cobra.Command, args []string) error {
	// Check if rbenv is installed
	if !o.isRbenvInstalled() {
		return fmt.Errorf("rbenv is not installed or not in PATH")
	}

	fmt.Println("💎 Starting rbenv update process...")

	// Update rbenv itself
	if o.updateRbenv {
		if err := o.updateRbenvInstallation(); err != nil {
			fmt.Printf("⚠️  Failed to update rbenv: %v\n", err)
		}
	}

	// Update rbenv plugins
	if o.updatePlugins {
		if err := o.updateRbenvPlugins(); err != nil {
			fmt.Printf("⚠️  Failed to update rbenv plugins: %v\n", err)
		}
	}

	// Get available Ruby versions
	availableVersions, err := o.getAvailableRubyVersions()
	if err != nil {
		return fmt.Errorf("failed to get available Ruby versions: %w", err)
	}

	if len(availableVersions) == 0 {
		fmt.Println("📝 No Ruby versions are available for installation")
		return nil
	}

	// Determine target versions based on strategy and user input
	targetVersions, err := o.getTargetVersions(availableVersions)
	if err != nil {
		return fmt.Errorf("failed to determine target versions: %w", err)
	}

	if len(targetVersions) == 0 {
		fmt.Println("✅ No new Ruby versions to install")
		return nil
	}

	fmt.Printf("📦 Found %d Ruby version(s) to install: %v\n", len(targetVersions), targetVersions)

	// Install each version
	installedCount := 0
	var latestInstalled string
	for _, version := range targetVersions {
		installed, err := o.installRubyVersion(version)
		if err != nil {
			fmt.Printf("⚠️  Failed to install Ruby %s: %v\n", version, err)
			continue
		}
		if installed {
			installedCount++
			latestInstalled = version
		}
	}

	// Set global version if requested and we installed something
	if o.global && latestInstalled != "" {
		if err := o.setGlobalVersion(latestInstalled); err != nil {
			fmt.Printf("⚠️  Failed to set global version: %v\n", err)
		}
	}

	// Rehash if requested and we installed something
	if o.rehash && installedCount > 0 && !o.dryRun {
		if err := o.rehashRbenv(); err != nil {
			fmt.Printf("⚠️  Failed to rehash rbenv: %v\n", err)
		}
	}

	// Summary
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d Ruby versions would be installed\n", installedCount)
	} else {
		fmt.Printf("✅ Installation completed. %d Ruby versions were installed\n", installedCount)
	}

	return nil
}

func (o *alwaysLatestRbenvOptions) isRbenvInstalled() bool {
	_, err := exec.LookPath("rbenv")
	return err == nil
}

func (o *alwaysLatestRbenvOptions) updateRbenvInstallation() error {
	fmt.Println("🔄 Updating rbenv...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would update rbenv installation")
		return nil
	}

	// Try git pull if rbenv was installed via git
	gitCmd := exec.Command("git", "-C", "$(rbenv root)", "pull")
	if err := gitCmd.Run(); err != nil {
		// If git update fails, try package manager update
		fmt.Println("   Git update failed, trying package manager...")

		// Try Homebrew if available
		if _, err := exec.LookPath("brew"); err == nil {
			brewCmd := exec.Command("brew", "upgrade", "rbenv")
			if brewErr := brewCmd.Run(); brewErr == nil {
				fmt.Println("✅ rbenv updated via Homebrew")
				return nil
			}
		}

		fmt.Println("⚠️  Could not update rbenv automatically")
		return nil
	}

	fmt.Println("✅ rbenv updated via git")
	return nil
}

func (o *alwaysLatestRbenvOptions) updateRbenvPlugins() error {
	fmt.Println("🔌 Updating rbenv plugins...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would update rbenv plugins")
		return nil
	}

	// Update ruby-build plugin specifically
	rubyBuildCmd := exec.Command("git", "-C", "$(rbenv root)/plugins/ruby-build", "pull")
	if err := rubyBuildCmd.Run(); err == nil {
		fmt.Println("✅ ruby-build plugin updated")
	} else {
		// Try Homebrew ruby-build if git fails
		if _, err := exec.LookPath("brew"); err == nil {
			brewCmd := exec.Command("brew", "upgrade", "ruby-build")
			if brewErr := brewCmd.Run(); brewErr == nil {
				fmt.Println("✅ ruby-build updated via Homebrew")
			}
		}
	}

	return nil
}

func (o *alwaysLatestRbenvOptions) getAvailableRubyVersions() ([]string, error) {
	cmd := exec.Command("rbenv", "install", "--list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list available Ruby versions: %w", err)
	}

	var versions []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Filter for stable Ruby versions (e.g., 3.1.0, 3.2.1)
		if o.isStableRubyVersion(line) {
			versions = append(versions, line)
		}
	}

	return versions, nil
}

func (o *alwaysLatestRbenvOptions) isStableRubyVersion(version string) bool {
	// Match stable Ruby versions (e.g., 3.1.0, 3.2.1)
	re := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !re.MatchString(version) {
		return false
	}

	// Skip versions with unstable keywords
	unstablePatterns := []string{
		"preview", "rc", "dev", "snapshot", "alpha", "beta",
	}

	lowerVersion := strings.ToLower(version)
	for _, pattern := range unstablePatterns {
		if strings.Contains(lowerVersion, pattern) {
			return false
		}
	}

	return true
}

func (o *alwaysLatestRbenvOptions) getTargetVersions(availableVersions []string) ([]string, error) {
	// If specific versions are requested, filter for those
	if len(o.versions) > 0 {
		return o.filterRequestedVersions(availableVersions), nil
	}

	// Get installed versions
	installedVersions, err := o.getInstalledRubyVersions()
	if err != nil {
		return nil, err
	}

	// Find latest versions to install based on strategy
	return o.findVersionsToInstall(availableVersions, installedVersions), nil
}

func (o *alwaysLatestRbenvOptions) filterRequestedVersions(availableVersions []string) []string {
	versionSet := make(map[string]bool)
	for _, v := range o.versions {
		versionSet[strings.TrimSpace(v)] = true
	}

	var filtered []string
	for _, version := range availableVersions {
		// Check for exact match or major.minor match
		for requestedVersion := range versionSet {
			if version == requestedVersion || strings.HasPrefix(version, requestedVersion+".") {
				filtered = append(filtered, version)
				break
			}
		}
	}

	return filtered
}

func (o *alwaysLatestRbenvOptions) getInstalledRubyVersions() ([]string, error) {
	cmd := exec.Command("rbenv", "versions", "--bare")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed Ruby versions: %w", err)
	}

	var versions []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			versions = append(versions, line)
		}
	}

	return versions, nil
}

func (o *alwaysLatestRbenvOptions) findVersionsToInstall(availableVersions, installedVersions []string) []string {
	installedSet := make(map[string]bool)
	for _, v := range installedVersions {
		installedSet[v] = true
	}

	// Group versions by major.minor
	versionGroups := make(map[string][]string)
	for _, version := range availableVersions {
		if installedSet[version] {
			continue // Skip already installed versions
		}

		majorMinor, err := o.extractMajorMinor(version)
		if err != nil {
			continue
		}

		versionGroups[majorMinor] = append(versionGroups[majorMinor], version)
	}

	var targetVersions []string

	if o.strategy == "major" {
		// Install latest version from each major.minor series
		for _, versions := range versionGroups {
			sort.Strings(versions)
			if len(versions) > 0 {
				targetVersions = append(targetVersions, versions[len(versions)-1])
			}
		}
	} else {
		// Minor strategy: only update within existing major.minor series
		installedMajorMinors := make(map[string]bool)
		for _, installed := range installedVersions {
			if majorMinor, err := o.extractMajorMinor(installed); err == nil {
				installedMajorMinors[majorMinor] = true
			}
		}

		for majorMinor, versions := range versionGroups {
			if installedMajorMinors[majorMinor] {
				sort.Strings(versions)
				if len(versions) > 0 {
					targetVersions = append(targetVersions, versions[len(versions)-1])
				}
			}
		}
	}

	return targetVersions
}

func (o *alwaysLatestRbenvOptions) extractMajorMinor(version string) (string, error) {
	re := regexp.MustCompile(`^(\d+\.\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 2 {
		return "", fmt.Errorf("cannot extract major.minor from %s", version)
	}
	return matches[1], nil
}

func (o *alwaysLatestRbenvOptions) installRubyVersion(version string) (bool, error) {
	fmt.Printf("🔍 Checking Ruby %s...\n", version)

	// Check if already installed
	installed, err := o.isVersionInstalled(version)
	if err != nil {
		return false, err
	}

	if installed {
		fmt.Printf("   Ruby %s is already installed\n", version)
		return false, nil
	}

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would install Ruby %s\n", version)
		return true, nil
	}

	// Confirm installation in interactive mode
	if o.interactive && !o.confirmInstallation(version) {
		fmt.Printf("   Skipping Ruby %s installation\n", version)
		return false, nil
	}

	// Install the version
	fmt.Printf("   Installing Ruby %s...\n", version)
	cmd := exec.Command("rbenv", "install", version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("installation failed: %w\n%s", err, string(output))
	}

	fmt.Printf("✅ Installed Ruby %s\n", version)
	return true, nil
}

func (o *alwaysLatestRbenvOptions) isVersionInstalled(version string) (bool, error) {
	installedVersions, err := o.getInstalledRubyVersions()
	if err != nil {
		return false, err
	}

	for _, installed := range installedVersions {
		if installed == version {
			return true, nil
		}
	}

	return false, nil
}

func (o *alwaysLatestRbenvOptions) confirmInstallation(version string) bool {
	fmt.Printf("   Install Ruby %s? (y/N): ", version)
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestRbenvOptions) setGlobalVersion(version string) error {
	fmt.Printf("🌍 Setting Ruby %s as global version...\n", version)

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would set Ruby %s as global\n", version)
		return nil
	}

	cmd := exec.Command("rbenv", "global", version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set global version: %w\n%s", err, string(output))
	}

	fmt.Printf("✅ Ruby %s set as global version\n", version)
	return nil
}

func (o *alwaysLatestRbenvOptions) rehashRbenv() error {
	fmt.Println("🔄 Rehashing rbenv shims...")

	cmd := exec.Command("rbenv", "rehash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rehash failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ rbenv rehashed successfully")
	return nil
}

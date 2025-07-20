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

const (
	strategyMajor = "major"
	strategyMinor = "minor"
)

type alwaysLatestAsdfOptions struct {
	strategy    string
	tools       []string
	dryRun      bool
	updateAsdf  bool
	configFile  string
	global      bool
	interactive bool
}

func defaultAlwaysLatestAsdfOptions() *alwaysLatestAsdfOptions {
	return &alwaysLatestAsdfOptions{
		strategy:    "minor",
		tools:       []string{},
		dryRun:      false,
		updateAsdf:  true,
		global:      false,
		interactive: true,
	}
}

func newAlwaysLatestAsdfCmd(_ context.Context) *cobra.Command {
	o := defaultAlwaysLatestAsdfOptions()

	cmd := &cobra.Command{
		Use:   "asdf",
		Short: "Update asdf and its managed tools to latest versions",
		Long: `Update asdf version manager and its managed programming language tools.

This command helps keep your development environment current by:
- Updating asdf plugins to their latest versions
- Installing latest versions of programming language runtimes
- Optionally updating global or local tool versions
- Supporting both minor and major update strategies

Update strategies:
  minor: Update to latest patch/minor version within same major version (safer)
  major: Update to absolute latest version including major version changes

Examples:
  # Update all asdf tools using minor strategy (default)
  gz always-latest asdf
  
  # Update specific tools only
  gz always-latest asdf --tools nodejs,python,ruby
  
  # Update with major version strategy (includes breaking changes)
  gz always-latest asdf --strategy major
  
  # Dry run to see what would be updated
  gz always-latest asdf --dry-run
  
  # Update global versions (affects all projects)
  gz always-latest asdf --global`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major")
	cmd.Flags().StringSliceVar(&o.tools, "tools", o.tools, "Specific tools to update (comma-separated)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updateAsdf, "update-asdf", o.updateAsdf, "Update asdf plugins")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.global, "global", o.global, "Update global tool versions")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")

	return cmd
}

func (o *alwaysLatestAsdfOptions) run(_ *cobra.Command, _ []string) error {
	// Check if asdf is installed
	if !o.isAsdfInstalled() {
		return fmt.Errorf("asdf is not installed or not in PATH")
	}

	fmt.Println("🔄 Starting asdf update process...")

	// Update asdf plugins first
	if o.updateAsdf {
		if err := o.updateAsdfPlugins(); err != nil {
			return fmt.Errorf("failed to update asdf plugins: %w", err)
		}
	}

	// Get installed tools
	tools, err := o.getInstalledTools()
	if err != nil {
		return fmt.Errorf("failed to get installed tools: %w", err)
	}

	if len(tools) == 0 {
		fmt.Println("📝 No asdf tools are currently installed")
		return nil
	}

	// Filter tools if specific tools are requested
	if len(o.tools) > 0 {
		tools = o.filterTools(tools, o.tools)
	}

	fmt.Printf("📦 Found %d tools to potentially update: %v\n", len(tools), tools)

	// Update each tool
	updatedCount := 0

	for _, tool := range tools {
		updated, err := o.updateTool(tool)
		if err != nil {
			fmt.Printf("⚠️  Failed to update %s: %v\n", tool, err)
			continue
		}

		if updated {
			updatedCount++
		}
	}

	// Summary
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d tools would be updated\n", updatedCount)
	} else {
		fmt.Printf("✅ Update completed. %d tools were updated\n", updatedCount)
	}

	return nil
}

func (o *alwaysLatestAsdfOptions) isAsdfInstalled() bool {
	_, err := exec.LookPath("asdf")
	return err == nil
}

func (o *alwaysLatestAsdfOptions) updateAsdfPlugins() error {
	fmt.Println("🔌 Updating asdf plugins...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: asdf plugin update --all")
		return nil
	}

	cmd := exec.Command("asdf", "plugin", "update", "--all")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("plugin update failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Plugins updated successfully")

	return nil
}

func (o *alwaysLatestAsdfOptions) getInstalledTools() ([]string, error) {
	cmd := exec.Command("asdf", "plugin", "list")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	var tools []string

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		tool := strings.TrimSpace(scanner.Text())
		if tool != "" {
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

func (o *alwaysLatestAsdfOptions) filterTools(allTools, requestedTools []string) []string {
	toolSet := make(map[string]bool)
	for _, tool := range requestedTools {
		toolSet[strings.TrimSpace(tool)] = true
	}

	var filtered []string

	for _, tool := range allTools {
		if toolSet[tool] {
			filtered = append(filtered, tool)
		}
	}

	return filtered
}

func (o *alwaysLatestAsdfOptions) updateTool(tool string) (bool, error) {
	fmt.Printf("🔍 Checking %s...\n", tool)

	// Get current version
	currentVersion, err := o.getCurrentVersion(tool)
	if err != nil {
		return false, fmt.Errorf("failed to get current version: %w", err)
	}

	// Get latest version
	latestVersion, err := o.getLatestVersion(tool)
	if err != nil {
		return false, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Determine target version based on strategy
	targetVersion, err := o.getTargetVersion(tool, currentVersion, latestVersion)
	if err != nil {
		return false, fmt.Errorf("failed to determine target version: %w", err)
	}

	if targetVersion == "" || (currentVersion != "" && targetVersion == currentVersion) {
		fmt.Printf("   %s is already up to date (%s)\n", tool, currentVersion)
		return false, nil
	}

	fmt.Printf("   %s: %s → %s\n", tool, currentVersion, targetVersion)

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would install %s %s\n", tool, targetVersion)
		return true, nil
	}

	// Confirm update in interactive mode
	if o.interactive && !o.confirmUpdate(tool, currentVersion, targetVersion) {
		fmt.Printf("   Skipping %s update\n", tool)
		return false, nil
	}

	// Install the new version
	if err := o.installVersion(tool, targetVersion); err != nil {
		return false, err
	}

	// Set as global version if requested
	if o.global {
		if err := o.setGlobalVersion(tool, targetVersion); err != nil {
			return false, err
		}
	}

	fmt.Printf("✅ Updated %s to %s\n", tool, targetVersion)

	return true, nil
}

func (o *alwaysLatestAsdfOptions) getCurrentVersion(tool string) (string, error) {
	cmd := exec.Command("asdf", "current", tool)

	output, err := cmd.Output()
	if err != nil {
		return "", err // Tool might not have a version set
	}

	// Parse output: "nodejs          18.17.0         /path/to/.tool-versions"
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1]), nil
	}

	return "", nil
}

func (o *alwaysLatestAsdfOptions) getLatestVersion(tool string) (string, error) {
	cmd := exec.Command("asdf", "list", "all", tool)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list versions: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	// Filter out non-release versions and get the latest
	var versions []string

	for _, line := range lines {
		version := strings.TrimSpace(line)
		if version == "" {
			continue
		}

		// Skip pre-release versions (containing alpha, beta, rc, dev, etc.)
		if o.isStableVersion(version) {
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no stable versions found for %s", tool)
	}

	// Return the last (latest) version
	return versions[len(versions)-1], nil
}

func (o *alwaysLatestAsdfOptions) isStableVersion(version string) bool {
	unstablePatterns := []string{
		"alpha", "beta", "rc", "dev", "snapshot", "preview", "pre",
		"nightly", "canary", "experimental", "test",
	}

	lowerVersion := strings.ToLower(version)
	for _, pattern := range unstablePatterns {
		if strings.Contains(lowerVersion, pattern) {
			return false
		}
	}

	return true
}

func (o *alwaysLatestAsdfOptions) getTargetVersion(tool, currentVersion, latestVersion string) (string, error) {
	if o.strategy == strategyMajor {
		return latestVersion, nil
	}

	// For minor strategy, find the latest version within the same major version
	if currentVersion == "" {
		return latestVersion, nil
	}

	currentMajor, err := o.extractMajorVersion(currentVersion)
	if err != nil {
		return latestVersion, err // Fallback to latest if we can't parse
	}

	// Get all versions and find the latest within the same major version
	cmd := exec.Command("asdf", "list", "all", tool)

	output, err := cmd.Output()
	if err != nil {
		return latestVersion, err
	}

	lines := strings.Split(string(output), "\n")

	var candidateVersions []string

	for _, line := range lines {
		version := strings.TrimSpace(line)
		if version == "" || !o.isStableVersion(version) {
			continue
		}

		major, err := o.extractMajorVersion(version)
		if err != nil {
			continue
		}

		if major == currentMajor {
			candidateVersions = append(candidateVersions, version)
		}
	}

	if len(candidateVersions) == 0 {
		return currentVersion, nil // No updates within same major version
	}

	// Return the latest version within the same major version
	return candidateVersions[len(candidateVersions)-1], nil
}

func (o *alwaysLatestAsdfOptions) extractMajorVersion(version string) (string, error) {
	// Extract major version number from version string
	re := regexp.MustCompile(`^(\d+)`)

	matches := re.FindStringSubmatch(version)
	if len(matches) < 2 {
		return "", fmt.Errorf("cannot extract major version from %s", version)
	}

	return matches[1], nil
}

func (o *alwaysLatestAsdfOptions) confirmUpdate(tool, currentVersion, targetVersion string) bool {
	fmt.Printf("   Update %s from %s to %s? (y/N): ", tool, currentVersion, targetVersion)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// If reading fails, default to no
		return false
	}

	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestAsdfOptions) installVersion(tool, version string) error {
	fmt.Printf("   Installing %s %s...\n", tool, version)

	cmd := exec.Command("asdf", "install", tool, version)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w\n%s", err, string(output))
	}

	return nil
}

func (o *alwaysLatestAsdfOptions) setGlobalVersion(tool, version string) error {
	cmd := exec.Command("asdf", "global", tool, version)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set global version: %w\n%s", err, string(output))
	}

	return nil
}

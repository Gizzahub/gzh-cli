package alwayslatest

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type alwaysLatestSdkmanOptions struct {
	strategy    string
	candidates  []string
	dryRun      bool
	updateSdk   bool
	configFile  string
	global      bool
	interactive bool
	flushBefore bool
	cleanupOld  bool
}

func defaultAlwaysLatestSdkmanOptions() *alwaysLatestSdkmanOptions {
	return &alwaysLatestSdkmanOptions{
		strategy:    "minor",
		candidates:  []string{},
		dryRun:      false,
		updateSdk:   true,
		global:      false,
		interactive: true,
		flushBefore: false,
		cleanupOld:  false,
	}
}

func newAlwaysLatestSdkmanCmd(_ context.Context) *cobra.Command {
	o := defaultAlwaysLatestSdkmanOptions()

	cmd := &cobra.Command{
		Use:   "sdkman",
		Short: "Update SDKMAN and its managed SDKs to latest versions",
		Long: `Update SDKMAN Software Development Kit Manager and its managed Java ecosystem tools.

This command helps keep your development environment current by:
- Updating SDKMAN itself to the latest version
- Installing latest versions of Java SDKs and tools
- Supporting both minor and major update strategies
- Managing multiple versions of the same candidate

Supported candidates include:
  Java, Kotlin, Scala, Groovy, Maven, Gradle, SBT, Spring Boot CLI,
  Micronaut, Quarkus, and many other JVM ecosystem tools.

Update strategies:
  minor: Update to latest patch/minor version within same major version (safer)
  major: Update to absolute latest version including major version changes

Examples:
  # Update all installed SDKMAN candidates
  gz always-latest sdkman
  
  # Update specific candidates only
  gz always-latest sdkman --candidates java,gradle,maven
  
  # Update with major version strategy (includes breaking changes)
  gz always-latest sdkman --strategy major
  
  # Dry run to see what would be updated
  gz always-latest sdkman --dry-run
  
  # Update and set as default versions
  gz always-latest sdkman --global
  
  # Flush archives before updating (clean install)
  gz always-latest sdkman --flush-before`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major")
	cmd.Flags().StringSliceVar(&o.candidates, "candidates", o.candidates, "Specific candidates to update (comma-separated)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updateSdk, "update-sdk", o.updateSdk, "Update SDKMAN itself")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.global, "global", o.global, "Set updated versions as default")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")
	cmd.Flags().BoolVar(&o.flushBefore, "flush-before", o.flushBefore, "Flush archives before installing")
	cmd.Flags().BoolVar(&o.cleanupOld, "cleanup-old", o.cleanupOld, "Remove old versions after update")

	return cmd
}

func (o *alwaysLatestSdkmanOptions) run(_ *cobra.Command, _ []string) error {
	// Check if SDKMAN is installed
	if !o.isSdkmanInstalled() {
		return fmt.Errorf("SDKMAN is not installed or not properly configured")
	}

	fmt.Println("☕ Starting SDKMAN update process...")

	// Update SDKMAN itself first
	if o.updateSdk {
		if err := o.updateSdkman(); err != nil {
			return fmt.Errorf("failed to update SDKMAN: %w", err)
		}
	}

	// Flush archives if requested
	if o.flushBefore && !o.dryRun {
		if err := o.flushArchives(); err != nil {
			fmt.Printf("⚠️  Failed to flush archives: %v\n", err)
		}
	}

	// Get installed candidates
	candidates, err := o.getInstalledCandidates()
	if err != nil {
		return fmt.Errorf("failed to get installed candidates: %w", err)
	}

	if len(candidates) == 0 {
		fmt.Println("📝 No SDKMAN candidates are currently installed")
		return nil
	}

	// Filter candidates if specific candidates are requested
	if len(o.candidates) > 0 {
		candidates = o.filterCandidates(candidates, o.candidates)
	}

	fmt.Printf("📦 Found %d candidates to potentially update: %v\n", len(candidates), candidates)

	// Update each candidate
	updatedCount := 0

	for _, candidate := range candidates {
		updated, err := o.updateCandidate(candidate)
		if err != nil {
			fmt.Printf("⚠️  Failed to update %s: %v\n", candidate, err)
			continue
		}

		if updated {
			updatedCount++
		}
	}

	// Summary
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d candidates would be updated\n", updatedCount)
	} else {
		fmt.Printf("✅ Update completed. %d candidates were updated\n", updatedCount)
	}

	return nil
}

func (o *alwaysLatestSdkmanOptions) isSdkmanInstalled() bool {
	// Check if SDKMAN directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sdkmanDir := filepath.Join(homeDir, ".sdkman")
	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
		return false
	}

	// Check if sdk command is available
	_, err = exec.LookPath("sdk")

	return err == nil
}

func (o *alwaysLatestSdkmanOptions) updateSdkman() error {
	fmt.Println("🔄 Updating SDKMAN...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sdk selfupdate")
		return nil
	}

	cmd := exec.Command("bash", "-c", "source ~/.sdkman/bin/sdkman-init.sh && sdk selfupdate")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SDKMAN update failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ SDKMAN updated successfully")

	return nil
}

func (o *alwaysLatestSdkmanOptions) flushArchives() error {
	fmt.Println("🧹 Flushing SDKMAN archives...")

	cmd := exec.Command("bash", "-c", "source ~/.sdkman/bin/sdkman-init.sh && sdk flush archives")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flush archives failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Archives flushed successfully")

	return nil
}

func (o *alwaysLatestSdkmanOptions) getInstalledCandidates() ([]string, error) {
	// Parse installed candidates from the output
	var candidates []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	candidatesDir := filepath.Join(homeDir, ".sdkman", "candidates")

	entries, err := os.ReadDir(candidatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read candidates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			candidate := entry.Name()
			// Check if candidate has any versions installed
			versionsDir := filepath.Join(candidatesDir, candidate)

			versions, err := os.ReadDir(versionsDir)
			if err != nil {
				continue
			}

			// Skip if no versions are installed (except 'current' symlink)
			hasVersions := false

			for _, version := range versions {
				if version.Name() != "current" {
					hasVersions = true
					break
				}
			}

			if hasVersions {
				candidates = append(candidates, candidate)
			}
		}
	}

	return candidates, nil
}

func (o *alwaysLatestSdkmanOptions) filterCandidates(allCandidates, requestedCandidates []string) []string {
	candidateSet := make(map[string]bool)
	for _, candidate := range requestedCandidates {
		candidateSet[strings.TrimSpace(candidate)] = true
	}

	var filtered []string

	for _, candidate := range allCandidates {
		if candidateSet[candidate] {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

func (o *alwaysLatestSdkmanOptions) updateCandidate(candidate string) (bool, error) {
	fmt.Printf("🔍 Checking %s...\n", candidate)

	// Get current version
	currentVersion, err := o.getCurrentVersion(candidate)
	if err != nil {
		return false, fmt.Errorf("failed to get current version: %w", err)
	}

	// Get latest version
	latestVersion, err := o.getLatestVersion(candidate)
	if err != nil {
		return false, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Determine target version based on strategy
	targetVersion, err := o.getTargetVersion(candidate, currentVersion, latestVersion)
	if err != nil {
		return false, fmt.Errorf("failed to determine target version: %w", err)
	}

	if targetVersion == "" || (currentVersion != "" && targetVersion == currentVersion) {
		fmt.Printf("   %s is already up to date (%s)\n", candidate, currentVersion)
		return false, nil
	}

	fmt.Printf("   %s: %s → %s\n", candidate, currentVersion, targetVersion)

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would install %s %s\n", candidate, targetVersion)
		return true, nil
	}

	// Confirm update in interactive mode
	if o.interactive && !o.confirmUpdate(candidate, currentVersion, targetVersion) {
		fmt.Printf("   Skipping %s update\n", candidate)
		return false, nil
	}

	// Install the new version
	if err := o.installVersion(candidate, targetVersion); err != nil {
		return false, err
	}

	// Set as default version if requested
	if o.global {
		if err := o.setDefaultVersion(candidate, targetVersion); err != nil {
			return false, err
		}
	}

	// Cleanup old versions if requested
	if o.cleanupOld {
		if err := o.cleanupOldVersions(candidate, targetVersion); err != nil {
			fmt.Printf("⚠️  Failed to cleanup old versions for %s: %v\n", candidate, err)
		}
	}

	fmt.Printf("✅ Updated %s to %s\n", candidate, targetVersion)

	return true, nil
}

func (o *alwaysLatestSdkmanOptions) getCurrentVersion(candidate string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source ~/.sdkman/bin/sdkman-init.sh && sdk current %s", candidate))

	output, err := cmd.Output()
	if err != nil {
		return "", err // Candidate might not have a current version set
	}

	// Parse output: "Using java version 11.0.19-tem"
	outputStr := strings.TrimSpace(string(output))
	if strings.Contains(outputStr, "version") {
		parts := strings.Split(outputStr, " ")
		if len(parts) >= 3 {
			return parts[len(parts)-1], nil
		}
	}

	return "", nil
}

func (o *alwaysLatestSdkmanOptions) getLatestVersion(candidate string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source ~/.sdkman/bin/sdkman-init.sh && sdk list %s", candidate))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list versions for %s: %w", candidate, err)
	}

	// Parse the output to find available versions
	var versions []string

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	inVersionSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for version lines (typically contain version numbers)
		if strings.Contains(line, "|") && inVersionSection {
			// Split by | and extract version numbers
			parts := strings.Split(line, "|")
			for _, part := range parts {
				version := strings.TrimSpace(part)
				if version != "" && o.isValidVersion(version) {
					versions = append(versions, version)
				}
			}
		}

		// Start looking for versions after header
		if strings.Contains(line, "Available") || strings.Contains(line, "Vendor") {
			inVersionSection = true
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for %s", candidate)
	}

	// Sort versions and return the latest stable version
	stableVersions := o.filterStableVersions(versions)
	if len(stableVersions) == 0 {
		return "", fmt.Errorf("no stable versions found for %s", candidate)
	}

	// Return the first stable version (SDKMAN typically lists newest first)
	return stableVersions[0], nil
}

func (o *alwaysLatestSdkmanOptions) isValidVersion(version string) bool {
	// Check if string looks like a version number
	re := regexp.MustCompile(`\d+\.[\d.\w]+`)
	return re.MatchString(version)
}

func (o *alwaysLatestSdkmanOptions) filterStableVersions(versions []string) []string {
	var stable []string

	unstablePatterns := []string{
		"ea", "beta", "alpha", "rc", "snapshot", "preview", "dev",
		"nightly", "canary", "experimental", "test", "milestone",
	}

	for _, version := range versions {
		isStable := true
		lowerVersion := strings.ToLower(version)

		for _, pattern := range unstablePatterns {
			if strings.Contains(lowerVersion, pattern) {
				isStable = false
				break
			}
		}

		if isStable {
			stable = append(stable, version)
		}
	}

	return stable
}

func (o *alwaysLatestSdkmanOptions) getTargetVersion(candidate, currentVersion, latestVersion string) (string, error) {
	if o.strategy == "major" {
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
	candidateVersions, err := o.getCompatibleVersions(candidate, currentMajor)
	if err != nil {
		return latestVersion, err
	}

	if len(candidateVersions) == 0 {
		return currentVersion, nil // No updates within same major version
	}

	// Sort and return the latest version within the same major version
	sort.Strings(candidateVersions)

	return candidateVersions[len(candidateVersions)-1], nil
}

// getCompatibleVersions retrieves and filters versions compatible with the current major version.
func (o *alwaysLatestSdkmanOptions) getCompatibleVersions(candidate, currentMajor string) ([]string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source ~/.sdkman/bin/sdkman-init.sh && sdk list %s", candidate))

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var candidateVersions []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	inVersionSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "|") && inVersionSection {
			candidateVersions = append(candidateVersions, o.parseVersionsFromLine(line, currentMajor)...)
		}

		if strings.Contains(line, "Available") || strings.Contains(line, "Vendor") {
			inVersionSection = true
		}
	}

	return candidateVersions, nil
}

// parseVersionsFromLine extracts compatible versions from a single output line.
func (o *alwaysLatestSdkmanOptions) parseVersionsFromLine(line, currentMajor string) []string {
	var versions []string
	parts := strings.Split(line, "|")

	for _, part := range parts {
		version := strings.TrimSpace(part)
		if version != "" && o.isValidVersion(version) {
			major, err := o.extractMajorVersion(version)
			if err != nil {
				continue
			}

			if major == currentMajor && o.isStableVersion(version) {
				versions = append(versions, version)
			}
		}
	}

	return versions
}

func (o *alwaysLatestSdkmanOptions) extractMajorVersion(version string) (string, error) {
	// Extract major version number from version string
	re := regexp.MustCompile(`^(\d+)`)

	matches := re.FindStringSubmatch(version)
	if len(matches) < 2 {
		return "", fmt.Errorf("cannot extract major version from %s", version)
	}

	return matches[1], nil
}

func (o *alwaysLatestSdkmanOptions) isStableVersion(version string) bool {
	unstablePatterns := []string{
		"ea", "beta", "alpha", "rc", "snapshot", "preview", "dev",
		"nightly", "canary", "experimental", "test", "milestone",
	}

	lowerVersion := strings.ToLower(version)
	for _, pattern := range unstablePatterns {
		if strings.Contains(lowerVersion, pattern) {
			return false
		}
	}

	return true
}

func (o *alwaysLatestSdkmanOptions) confirmUpdate(candidate, currentVersion, targetVersion string) bool {
	fmt.Printf("   Update %s from %s to %s? (y/N): ", candidate, currentVersion, targetVersion)

	var response string

	_, _ = fmt.Scanln(&response)

	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestSdkmanOptions) installVersion(candidate, version string) error {
	fmt.Printf("   Installing %s %s...\n", candidate, version)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("source ~/.sdkman/bin/sdkman-init.sh && sdk install %s %s", candidate, version))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w\n%s", err, string(output))
	}

	return nil
}

func (o *alwaysLatestSdkmanOptions) setDefaultVersion(candidate, version string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source ~/.sdkman/bin/sdkman-init.sh && sdk default %s %s", candidate, version))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set default version: %w\n%s", err, string(output))
	}

	return nil
}

func (o *alwaysLatestSdkmanOptions) cleanupOldVersions(candidate, keepVersion string) error {
	fmt.Printf("   Cleaning up old versions of %s...\n", candidate)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	candidateDir := filepath.Join(homeDir, ".sdkman", "candidates", candidate)

	entries, err := os.ReadDir(candidateDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" && entry.Name() != keepVersion {
			versionDir := filepath.Join(candidateDir, entry.Name())
			if err := os.RemoveAll(versionDir); err != nil {
				fmt.Printf("     ⚠️  Failed to remove %s: %v\n", entry.Name(), err)
			} else {
				fmt.Printf("     🗑️  Removed %s %s\n", candidate, entry.Name())
			}
		}
	}

	return nil
}

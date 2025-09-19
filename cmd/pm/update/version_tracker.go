// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package update

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// VersionTracker tracks package version changes across different package managers
type VersionTracker struct {
	formatter *OutputFormatter
	changes   map[string][]PackageChange
}

// NewVersionTracker creates a new version tracker
func NewVersionTracker(formatter *OutputFormatter) *VersionTracker {
	return &VersionTracker{
		formatter: formatter,
		changes:   make(map[string][]PackageChange),
	}
}

// TrackBrewUpdates extracts version changes from brew upgrade output
func (vt *VersionTracker) TrackBrewUpdates(ctx context.Context, dryRun bool) error {
	var cmd *exec.Cmd

	if dryRun {
		// For dry run, show what would be upgraded
		cmd = exec.CommandContext(ctx, "brew", "outdated", "--verbose")
	} else {
		// For real update, we need to capture before state first
		return vt.trackBrewRealtime(ctx)
	}

	output, err := cmd.Output()
	if err != nil {
		// No outdated packages or error - this is OK for dry run
		return nil
	}

	changes := vt.parseBrewOutdated(string(output))
	if len(changes) > 0 {
		vt.changes["brew"] = changes
		vt.printBrewChanges(changes, dryRun)
	}

	return nil
}

// trackBrewRealtime tracks changes during actual brew upgrade
func (vt *VersionTracker) trackBrewRealtime(ctx context.Context) error {
	// First, get current versions
	beforeCmd := exec.CommandContext(ctx, "brew", "list", "--versions")
	beforeOutput, err := beforeCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current brew versions: %w", err)
	}

	beforeVersions := vt.parseBrewVersions(string(beforeOutput))

	// Run the upgrade command
	upgradeCmd := exec.CommandContext(ctx, "brew", "upgrade")
	upgradeOutput, err := upgradeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew upgrade failed: %w", err)
	}

	// Print the upgrade output in real-time format
	fmt.Print(string(upgradeOutput))

	// Get versions after upgrade
	afterCmd := exec.CommandContext(ctx, "brew", "list", "--versions")
	afterOutput, err := afterCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get updated brew versions: %w", err)
	}

	afterVersions := vt.parseBrewVersions(string(afterOutput))
	changes := vt.compareVersions("brew", beforeVersions, afterVersions)

	if len(changes) > 0 {
		vt.changes["brew"] = changes
		vt.printBrewChanges(changes, false)
	}

	return nil
}

// parseBrewOutdated parses 'brew outdated --verbose' output
func (vt *VersionTracker) parseBrewOutdated(output string) []PackageChange {
	var changes []PackageChange

	// Example line: "node (20.11.0) < 20.11.1 [pinned at 20.11.0]"
	re := regexp.MustCompile(`^(\S+)\s+\(([^)]+)\)\s+<\s+([^\s\[]+)`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			name := matches[1]
			oldVersion := matches[2]
			newVersion := matches[3]

			change := PackageChange{
				Name:       name,
				OldVersion: oldVersion,
				NewVersion: newVersion,
				DownloadMB: vt.estimatePackageSize(name, "brew"),
				UpdateType: vt.determineUpdateType(oldVersion, newVersion),
				Manager:    "brew",
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// parseBrewVersions parses 'brew list --versions' output
func (vt *VersionTracker) parseBrewVersions(output string) map[string]string {
	versions := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			version := parts[1] // Take first version if multiple exist
			versions[name] = version
		}
	}

	return versions
}

// compareVersions compares before and after version maps to find changes
func (vt *VersionTracker) compareVersions(manager string, before, after map[string]string) []PackageChange {
	var changes []PackageChange

	for name, afterVersion := range after {
		if beforeVersion, exists := before[name]; exists {
			if beforeVersion != afterVersion {
				change := PackageChange{
					Name:       name,
					OldVersion: beforeVersion,
					NewVersion: afterVersion,
					DownloadMB: vt.estimatePackageSize(name, manager),
					UpdateType: vt.determineUpdateType(beforeVersion, afterVersion),
					Manager:    manager,
				}
				changes = append(changes, change)
			}
		}
	}

	return changes
}

// TrackAsdfUpdates extracts version changes from asdf plugin updates
func (vt *VersionTracker) TrackAsdfUpdates(ctx context.Context, plugin string, dryRun bool) error {
	// Get current version
	currentCmd := exec.CommandContext(ctx, "asdf", "current", plugin)
	currentOutput, err := currentCmd.Output()
	if err != nil {
		return nil // Plugin might not be installed
	}

	currentVersion := vt.parseAsdfCurrentVersion(string(currentOutput))
	if currentVersion == "" {
		return nil
	}

	// Get latest available version
	latestCmd := exec.CommandContext(ctx, "asdf", "latest", plugin)
	latestOutput, err := latestCmd.Output()
	if err != nil {
		return err
	}

	latestVersion := strings.TrimSpace(string(latestOutput))
	if latestVersion == currentVersion {
		fmt.Printf("💡 %s: %s already latest, skipping\n", plugin, currentVersion)
		return nil
	}

	// Create change record
	change := PackageChange{
		Name:       plugin,
		OldVersion: currentVersion,
		NewVersion: latestVersion,
		DownloadMB: vt.estimatePackageSize(plugin, "asdf"),
		UpdateType: vt.determineUpdateType(currentVersion, latestVersion),
		Manager:    "asdf",
	}

	if dryRun {
		fmt.Printf("Would update %s: %s → %s (%.1fMB)\n",
			plugin, currentVersion, latestVersion, change.DownloadMB)
	} else {
		vt.formatter.PrintPackageChange(change)
	}

	// Store change
	if vt.changes["asdf"] == nil {
		vt.changes["asdf"] = make([]PackageChange, 0)
	}
	vt.changes["asdf"] = append(vt.changes["asdf"], change)

	return nil
}

// parseAsdfCurrentVersion parses 'asdf current <plugin>' output
func (vt *VersionTracker) parseAsdfCurrentVersion(output string) string {
	// Example: "nodejs 20.11.0 (set by /Users/user/.tool-versions)"
	fields := strings.Fields(strings.TrimSpace(output))
	if len(fields) >= 2 {
		return fields[1]
	}
	return ""
}

// TrackNpmUpdates tracks npm global package updates
func (vt *VersionTracker) TrackNpmUpdates(ctx context.Context, dryRun bool) error {
	// Get outdated packages
	cmd := exec.CommandContext(ctx, "npm", "outdated", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil // No outdated packages
	}

	changes := vt.parseNpmOutdated(string(output))
	if len(changes) > 0 {
		vt.changes["npm"] = changes
		for _, change := range changes {
			if dryRun {
				fmt.Printf("Would update %s: %s → %s (%.1fMB)\n",
					change.Name, change.OldVersion, change.NewVersion, change.DownloadMB)
			} else {
				vt.formatter.PrintPackageChange(change)
			}
		}
	}

	return nil
}

// parseNpmOutdated parses npm outdated JSON output
func (vt *VersionTracker) parseNpmOutdated(output string) []PackageChange {
	// This would parse JSON output from npm outdated
	// For now, return empty slice as placeholder
	return []PackageChange{}
}

// TrackPipUpdates tracks pip package updates
func (vt *VersionTracker) TrackPipUpdates(ctx context.Context, pipCmd string, dryRun bool) error {
	// Get outdated packages
	cmd := exec.CommandContext(ctx, pipCmd, "list", "--outdated", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to freeze format
		return vt.trackPipFreezeFormat(ctx, pipCmd, dryRun)
	}

	changes := vt.parsePipOutdatedJSON(string(output))
	if len(changes) > 0 {
		vt.changes["pip"] = changes
		for _, change := range changes {
			if dryRun {
				fmt.Printf("Would update %s: %s → %s (%.1fMB)\n",
					change.Name, change.OldVersion, change.NewVersion, change.DownloadMB)
			} else {
				vt.formatter.PrintPackageChange(change)
			}
		}
	}

	return nil
}

// trackPipFreezeFormat tracks pip updates using freeze format (fallback)
func (vt *VersionTracker) trackPipFreezeFormat(ctx context.Context, pipCmd string, dryRun bool) error {
	parts := strings.Fields(pipCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid pip command")
	}

	args := parts[1:]
	args = append(args, "list", "--outdated", "--format=freeze")
	cmd := exec.CommandContext(ctx, parts[0], args...)

	output, err := cmd.Output()
	if err != nil {
		return nil // No outdated packages
	}

	changes := vt.parsePipFreezeFormat(string(output))
	if len(changes) > 0 {
		vt.changes["pip"] = changes
		for _, change := range changes {
			if dryRun {
				fmt.Printf("Would update %s: %s → %s (%.1fMB)\n",
					change.Name, change.OldVersion, change.NewVersion, change.DownloadMB)
			} else {
				vt.formatter.PrintPackageChange(change)
			}
		}
	}

	return nil
}

// parsePipOutdatedJSON parses pip list --outdated --format=json output
func (vt *VersionTracker) parsePipOutdatedJSON(output string) []PackageChange {
	// This would parse JSON output - placeholder for now
	return []PackageChange{}
}

// parsePipFreezeFormat parses pip list --outdated --format=freeze output
func (vt *VersionTracker) parsePipFreezeFormat(output string) []PackageChange {
	var changes []PackageChange

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Format: "package==current_version"
		parts := strings.Split(line, "==")
		if len(parts) == 2 {
			name := parts[0]
			currentVersion := parts[1]

			// For freeze format, we don't have latest version info
			// This would need additional API call or parsing
			change := PackageChange{
				Name:       name,
				OldVersion: currentVersion,
				NewVersion: "latest", // Placeholder
				DownloadMB: vt.estimatePackageSize(name, "pip"),
				UpdateType: "unknown",
				Manager:    "pip",
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// printBrewChanges prints brew package changes with enhanced formatting
func (vt *VersionTracker) printBrewChanges(changes []PackageChange, dryRun bool) {
	for _, change := range changes {
		if dryRun {
			fmt.Printf("Would upgrade %s: %s → %s (%.1fMB)\n",
				change.Name, change.OldVersion, change.NewVersion, change.DownloadMB)
		} else {
			vt.formatter.PrintPackageChange(change)
		}
	}
}

// estimatePackageSize estimates download size for a package
func (vt *VersionTracker) estimatePackageSize(name, manager string) float64 {
	// Package size estimates in MB based on common packages
	estimates := map[string]map[string]float64{
		"brew": {
			"node":           24.8,
			"git":            8.4,
			"python":         15.2,
			"jq":             1.1,
			"tree":           0.156,
			"go":             45.0,
			"rust":           85.0,
			"docker":         120.0,
			"kubernetes-cli": 12.5,
		},
		"asdf": {
			"nodejs": 25.0,
			"python": 20.0,
			"golang": 50.0,
			"rust":   90.0,
			"ruby":   15.0,
			"java":   180.0,
		},
		"npm": {
			"@angular/cli": 45.0,
			"typescript":   8.2,
			"prettier":     2.1,
			"eslint":       5.8,
			"webpack":      12.5,
		},
		"pip": {
			"requests":   2.1,
			"numpy":      15.2,
			"pandas":     25.8,
			"matplotlib": 18.5,
			"django":     8.9,
			"flask":      2.3,
		},
	}

	if managerEstimates, exists := estimates[manager]; exists {
		if size, exists := managerEstimates[name]; exists {
			return size
		}
	}

	// Default estimates by manager type
	defaults := map[string]float64{
		"brew":   5.0,
		"asdf":   20.0,
		"npm":    3.5,
		"pip":    2.8,
		"apt":    1.5,
		"pacman": 2.0,
		"yay":    3.0,
		"sdkman": 25.0,
	}

	if defaultSize, exists := defaults[manager]; exists {
		return defaultSize
	}

	return 2.0 // Generic fallback
}

// determineUpdateType determines if this is a major, minor, or patch update
func (vt *VersionTracker) determineUpdateType(oldVer, newVer string) string {
	oldParts := vt.parseSemanticVersion(oldVer)
	newParts := vt.parseSemanticVersion(newVer)

	if len(oldParts) != 3 || len(newParts) != 3 {
		return "unknown"
	}

	if newParts[0] > oldParts[0] {
		return "major"
	}
	if newParts[1] > oldParts[1] {
		return "minor"
	}
	if newParts[2] > oldParts[2] {
		return "patch"
	}

	return "unknown"
}

// parseSemanticVersion parses semantic version string into major.minor.patch
func (vt *VersionTracker) parseSemanticVersion(version string) []int {
	// Remove common prefixes
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	// Split on dots and parse
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil
	}

	result := make([]int, 3)
	for i := 0; i < 3; i++ {
		// Remove any non-numeric suffixes (e.g., "1.2.3-beta")
		numStr := strings.Split(parts[i], "-")[0]
		if num, err := strconv.Atoi(numStr); err == nil {
			result[i] = num
		} else {
			return nil
		}
	}

	return result
}

// GetAllChanges returns all tracked changes
func (vt *VersionTracker) GetAllChanges() map[string][]PackageChange {
	return vt.changes
}

// GetTotalPackageCount returns total number of packages changed
func (vt *VersionTracker) GetTotalPackageCount() int {
	total := 0
	for _, changes := range vt.changes {
		total += len(changes)
	}
	return total
}

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

type alwaysLatestPortOptions struct {
	strategy    string
	ports       []string
	dryRun      bool
	updatePorts bool
	configFile  string
	interactive bool
	cleanup     bool
	upgradeAll  bool
	selfUpdate  bool
}

func defaultAlwaysLatestPortOptions() *alwaysLatestPortOptions {
	return &alwaysLatestPortOptions{
		strategy:    "minor",
		ports:       []string{},
		dryRun:      false,
		updatePorts: true,
		interactive: true,
		cleanup:     false,
		upgradeAll:  false,
		selfUpdate:  true,
	}
}

func newAlwaysLatestPortCmd(_ context.Context) *cobra.Command {
	o := defaultAlwaysLatestPortOptions()

	cmd := &cobra.Command{
		Use:   "port",
		Short: "Update MacPorts and its managed ports to latest versions",
		Long: `Update MacPorts package manager and its managed software ports.

This command helps keep your development environment current by:
- Updating MacPorts itself to the latest version
- Synchronizing the ports tree with the latest port definitions
- Upgrading outdated ports to their latest versions
- Optionally cleaning up old port versions

MacPorts is a package management system for macOS that provides easy installation
and management of open source software.

Update strategies:
  minor: Update to latest available version (MacPorts handles compatibility)
  major: Same as minor (MacPorts doesn't use semantic versioning like other tools)

Examples:
  # Update all outdated ports (default)
  gz always-latest port
  
  # Update specific ports only
  gz always-latest port --ports python39,git,wget
  
  # Update and cleanup old versions
  gz always-latest port --cleanup
  
  # Dry run to see what would be updated
  gz always-latest port --dry-run
  
  # Update all ports without prompting
  gz always-latest port --upgrade-all
  
  # Skip MacPorts self-update
  gz always-latest port --self-update=false`,
		RunE: o.run,
	}

	cmd.Flags().StringVar(&o.strategy, "strategy", o.strategy, "Update strategy: minor or major (both work the same for port)")
	cmd.Flags().StringSliceVar(&o.ports, "ports", o.ports, "Specific ports to update (comma-separated)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", o.dryRun, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&o.updatePorts, "update-ports", o.updatePorts, "Update MacPorts and sync ports tree")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Configuration file path")
	cmd.Flags().BoolVar(&o.interactive, "interactive", o.interactive, "Interactive mode for confirmation")
	cmd.Flags().BoolVar(&o.cleanup, "cleanup", o.cleanup, "Clean up old versions after update")
	cmd.Flags().BoolVar(&o.upgradeAll, "upgrade-all", o.upgradeAll, "Upgrade all ports without individual confirmation")
	cmd.Flags().BoolVar(&o.selfUpdate, "self-update", o.selfUpdate, "Update MacPorts itself")

	return cmd
}

func (o *alwaysLatestPortOptions) run(_ *cobra.Command, _ []string) error {
	// Check if MacPorts is installed
	if !o.isPortInstalled() {
		return fmt.Errorf("MacPorts is not installed or not in PATH")
	}

	fmt.Println("🚢 Starting MacPorts update process...")

	// Update MacPorts itself and sync ports tree
	if o.updatePorts {
		if o.selfUpdate {
			if err := o.updateMacPorts(); err != nil {
				return fmt.Errorf("failed to update MacPorts: %w", err)
			}
		}

		if err := o.syncPortsTree(); err != nil {
			return fmt.Errorf("failed to sync ports tree: %w", err)
		}
	}

	// Get outdated ports
	outdatedPorts, err := o.getOutdatedPorts()
	if err != nil {
		return fmt.Errorf("failed to get outdated ports: %w", err)
	}

	// Filter ports if specific ports are requested
	if len(o.ports) > 0 {
		outdatedPorts = o.filterPorts(outdatedPorts, o.ports)
	}

	if len(outdatedPorts) == 0 {
		fmt.Println("✅ All ports are up to date")
		return nil
	}

	fmt.Printf("📦 Found %d outdated port(s): %v\n", len(outdatedPorts), outdatedPorts)

	// Update ports
	updatedCount := 0

	for _, port := range outdatedPorts {
		updated, err := o.updatePort(port)
		if err != nil {
			fmt.Printf("⚠️  Failed to update %s: %v\n", port, err)
			continue
		}

		if updated {
			updatedCount++
		}
	}

	// Cleanup if requested
	if o.cleanup && !o.dryRun {
		if err := o.cleanupPorts(); err != nil {
			fmt.Printf("⚠️  Failed to cleanup: %v\n", err)
		}
	}

	// Summary
	if o.dryRun {
		fmt.Printf("🔍 Dry run completed. %d ports would be updated\n", updatedCount)
	} else {
		fmt.Printf("✅ Update completed. %d ports were updated\n", updatedCount)
	}

	return nil
}

func (o *alwaysLatestPortOptions) isPortInstalled() bool {
	_, err := exec.LookPath("port")
	return err == nil
}

func (o *alwaysLatestPortOptions) updateMacPorts() error {
	fmt.Println("🔄 Updating MacPorts...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sudo port selfupdate")
		return nil
	}

	cmd := exec.Command("sudo", "port", "selfupdate")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("MacPorts selfupdate failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ MacPorts updated successfully")

	return nil
}

func (o *alwaysLatestPortOptions) syncPortsTree() error {
	fmt.Println("🔄 Syncing ports tree...")

	if o.dryRun {
		fmt.Println("   [DRY RUN] Would run: sudo port sync")
		return nil
	}

	cmd := exec.Command("sudo", "port", "sync")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("port sync failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Ports tree synced successfully")

	return nil
}

func (o *alwaysLatestPortOptions) getOutdatedPorts() ([]string, error) {
	cmd := exec.Command("port", "outdated")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated ports: %w", err)
	}

	var ports []string

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "The following installed ports are outdated:") {
			continue
		}

		// Parse port name from output (format: "portname @version (active) < @newversion")
		parts := strings.Fields(line)
		if len(parts) > 0 {
			ports = append(ports, parts[0])
		}
	}

	return ports, nil
}

func (o *alwaysLatestPortOptions) filterPorts(allPorts, requestedPorts []string) []string {
	portSet := make(map[string]bool)
	for _, port := range requestedPorts {
		portSet[strings.TrimSpace(port)] = true
	}

	var filtered []string

	for _, port := range allPorts {
		if portSet[port] {
			filtered = append(filtered, port)
		}
	}

	return filtered
}

func (o *alwaysLatestPortOptions) updatePort(portName string) (bool, error) {
	fmt.Printf("🔍 Checking %s...\n", portName)

	if o.dryRun {
		fmt.Printf("   [DRY RUN] Would run: sudo port upgrade %s\n", portName)
		return true, nil
	}

	// Confirm update in interactive mode
	if o.interactive && !o.upgradeAll && !o.confirmUpdate(portName) {
		fmt.Printf("   Skipping %s update\n", portName)
		return false, nil
	}

	// Upgrade the port
	fmt.Printf("   Upgrading %s...\n", portName)
	cmd := exec.Command("sudo", "port", "upgrade", portName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("upgrade failed: %w\n%s", err, string(output))
	}

	fmt.Printf("✅ Updated %s\n", portName)

	return true, nil
}

func (o *alwaysLatestPortOptions) confirmUpdate(portName string) bool {
	fmt.Printf("   Update %s? (y/N): ", portName)

	var response string

	_, _ = fmt.Scanln(&response)

	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (o *alwaysLatestPortOptions) cleanupPorts() error {
	fmt.Println("🧹 Cleaning up inactive ports...")

	cmd := exec.Command("sudo", "port", "uninstall", "inactive")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Cleanup completed")

	return nil
}

func (o *alwaysLatestPortOptions) getPortVersion(portName string) (string, error) {
	cmd := exec.Command("port", "list", "installed", portName)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output (format: "portname @version (active)")
	line := strings.TrimSpace(string(output))
	if line == "" {
		return "", fmt.Errorf("no version found for %s", portName)
	}

	// Extract version using regex
	re := regexp.MustCompile(`@([\d.]+)`)

	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse version for %s", portName)
	}

	return matches[1], nil
}

func (o *alwaysLatestPortOptions) extractVersionNumber(versionString string) string {
	// Extract version number using regex
	re := regexp.MustCompile(`(\d+\.[\d.]+)`)

	matches := re.FindStringSubmatch(versionString)
	if len(matches) > 1 {
		return matches[1]
	}

	return versionString
}

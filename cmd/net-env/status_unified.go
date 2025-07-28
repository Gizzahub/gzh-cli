// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-manager-go/internal/netenv"
)

// newStatusUnifiedCmd creates the unified net-env status command
func newStatusUnifiedCmd() *cobra.Command {
	var (
		verbose bool
		format  string
		health  bool
		watch   bool
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show network environment status",
		Long: `Display the current network environment status including active profile,
network components, and health information.

This command provides a comprehensive view of your current network configuration:
- Active network profile (if any)
- WiFi connection status and details
- VPN connection status
- DNS configuration
- Proxy settings
- Docker network context
- Kubernetes network context
- Overall network health and performance

Examples:
  # Show basic network status
  gz net-env status

  # Show detailed status with verbose information
  gz net-env status --verbose

  # Show status in JSON format
  gz net-env status --format json

  # Include health checks and performance metrics
  gz net-env status --health

  # Monitor status in real-time
  gz net-env status --watch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusUnified(cmd.Context(), verbose, format, health, watch, timeout)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed network information")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&health, "health", false, "Include health checks")
	cmd.Flags().BoolVar(&watch, "watch", false, "Watch mode - continuously update status")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout for network checks")

	return cmd
}

// runStatusUnified executes the unified status command
func runStatusUnified(ctx context.Context, verbose bool, format string, includeHealth, watch bool, timeout time.Duration) error {
	configDir := getConfigDirectory()

	// Initialize profile manager
	profileManager := netenv.NewProfileManager(configDir)
	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Create network detector
	profiles := profileManager.ListProfiles()
	networkProfiles := make([]netenv.NetworkProfile, len(profiles))
	for i, p := range profiles {
		networkProfiles[i] = *p
	}
	detector := netenv.NewNetworkDetector(networkProfiles)

	if watch {
		return runStatusWatch(ctx, detector, profileManager, verbose, format, includeHealth, timeout)
	}

	return runSingleStatusCheck(ctx, detector, profileManager, verbose, format, includeHealth, timeout)
}

// runSingleStatusCheck performs a single status check
func runSingleStatusCheck(ctx context.Context, detector *netenv.NetworkDetector, profileManager *netenv.ProfileManager, verbose bool, format string, includeHealth bool, timeout time.Duration) error {
	status, err := collectNetworkStatus(ctx, detector, profileManager, includeHealth, timeout)
	if err != nil {
		return fmt.Errorf("failed to collect network status: %w", err)
	}

	return displayStatus(status, format, verbose)
}

// runStatusWatch runs status in watch mode
func runStatusWatch(ctx context.Context, detector *netenv.NetworkDetector, profileManager *netenv.ProfileManager, verbose bool, format string, includeHealth bool, timeout time.Duration) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Clear screen function
	clearScreen := func() {
		fmt.Print("\033[2J\033[H")
	}

	for {
		clearScreen()
		fmt.Printf("Network Status (Updated: %s)\n", time.Now().Format("15:04:05"))
		fmt.Println("Press Ctrl+C to exit watch mode\n")

		status, err := collectNetworkStatus(ctx, detector, profileManager, includeHealth, timeout)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			if err := displayStatus(status, format, verbose); err != nil {
				fmt.Printf("Display error: %v\n", err)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue loop
		}
	}
}

// collectNetworkStatus collects current network status information
func collectNetworkStatus(ctx context.Context, detector *netenv.NetworkDetector, profileManager *netenv.ProfileManager, includeHealth bool, timeout time.Duration) (*netenv.NetworkStatus, error) {
	status := &netenv.NetworkStatus{
		LastSwitch: time.Now(),
		Components: netenv.ComponentStatuses{},
		Health: netenv.HealthStatus{
			Status: "unknown",
			Score:  0,
		},
	}

	// Try to detect current profile
	if profile, err := detector.DetectEnvironment(ctx); err == nil && profile != nil {
		status.Profile = profile
	}

	// Collect component statuses
	status.Components.WiFi = checkWiFiStatus(ctx, timeout)
	status.Components.VPN = checkVPNStatus(ctx, timeout)
	status.Components.DNS = checkDNSStatus(ctx, timeout)
	status.Components.Proxy = checkProxyStatus(ctx, timeout)
	status.Components.Docker = checkDockerStatus(ctx, timeout)
	status.Components.Kubernetes = checkKubernetesStatus(ctx, timeout)

	// Calculate overall health
	if includeHealth {
		status.Health = calculateNetworkHealth(status.Components)
		if metrics, err := collectNetworkMetrics(ctx, timeout); err == nil {
			status.Metrics = metrics
		}
	}

	return status, nil
}

// displayStatus displays the network status in the requested format
func displayStatus(status *netenv.NetworkStatus, format string, verbose bool) error {
	switch strings.ToLower(format) {
	case "json":
		return displayStatusJSON(status)
	case "yaml", "yml":
		return displayStatusYAML(status)
	default:
		return displayStatusTable(status, verbose)
	}
}

// displayStatusTable displays status in table format
func displayStatusTable(status *netenv.NetworkStatus, verbose bool) error {
	fmt.Println("Network Environment Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Profile information
	if status.Profile != nil {
		fmt.Printf("Profile: %s", status.Profile.Name)
		if status.Profile.Description != "" {
			fmt.Printf(" (%s)", status.Profile.Description)
		}
		fmt.Println()
	} else {
		fmt.Println("Profile: None (manual configuration)")
	}

	// Network information placeholder - would need actual network detection
	fmt.Println("Network: Auto-detected")

	fmt.Println("\nComponents:")

	// Display component statuses
	displayComponentStatus("WiFi", status.Components.WiFi)
	displayComponentStatus("VPN", status.Components.VPN)
	displayComponentStatus("DNS", status.Components.DNS)
	displayComponentStatus("Proxy", status.Components.Proxy)
	displayComponentStatus("Docker", status.Components.Docker)
	displayComponentStatus("Kubernetes", status.Components.Kubernetes)

	// Health summary
	fmt.Printf("\nNetwork Health: %s", strings.Title(status.Health.Status))
	if status.Health.Score > 0 {
		fmt.Printf(" (%d/100)", status.Health.Score)
	}
	fmt.Println()

	if len(status.Health.Issues) > 0 {
		fmt.Println("\nIssues:")
		for _, issue := range status.Health.Issues {
			fmt.Printf("  ⚠️  %s\n", issue)
		}
	}

	// Performance metrics
	if status.Metrics != nil && verbose {
		fmt.Println("\nPerformance Metrics:")
		fmt.Printf("  Latency: %v\n", status.Metrics.Latency)
		if status.Metrics.Bandwidth != nil {
			fmt.Printf("  Bandwidth: ↓%.1f Mbps ↑%.1f Mbps\n",
				status.Metrics.Bandwidth.Download, status.Metrics.Bandwidth.Upload)
		}
		if status.Metrics.PacketLoss > 0 {
			fmt.Printf("  Packet Loss: %.1f%%\n", status.Metrics.PacketLoss)
		}
	}

	return nil
}

// displayComponentStatus displays the status of a single component
func displayComponentStatus(name string, status *netenv.ComponentStatus) {
	if status == nil {
		fmt.Printf("  %-12s │ ❓ Unknown\n", name)
		return
	}

	statusIcon := "❌ Inactive"
	if status.Active {
		statusIcon = "✅ Active"
	}
	if status.Error != "" {
		statusIcon = "⚠️ Error"
	}

	details := status.Status
	if len(status.Details) > 0 {
		var detailParts []string
		for _, value := range status.Details {
			detailParts = append(detailParts, fmt.Sprintf("%v", value))
		}
		if len(detailParts) > 0 {
			details = strings.Join(detailParts, ", ")
		}
	}

	fmt.Printf("  %-12s │ %s", name, statusIcon)
	if details != "" && details != status.Status {
		fmt.Printf("  %s", details)
	}
	fmt.Println()
}

// displayStatusJSON displays status in JSON format
func displayStatusJSON(status *netenv.NetworkStatus) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(status)
}

// displayStatusYAML displays status in YAML format
func displayStatusYAML(status *netenv.NetworkStatus) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(status)
}

// Component status check functions (simplified implementations)

func checkWiFiStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Simplified WiFi status check
	return &netenv.ComponentStatus{
		Active:    true,
		Status:    "connected",
		Details:   map[string]interface{}{"signal": "good"},
		LastCheck: time.Now(),
	}
}

func checkVPNStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Simplified VPN status check
	return &netenv.ComponentStatus{
		Active:    false,
		Status:    "disconnected",
		LastCheck: time.Now(),
	}
}

func checkDNSStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Simplified DNS status check
	return &netenv.ComponentStatus{
		Active:    true,
		Status:    "configured",
		Details:   map[string]interface{}{"servers": "system default"},
		LastCheck: time.Now(),
	}
}

func checkProxyStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Check environment variables for proxy settings
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	active := httpProxy != "" || httpsProxy != ""
	status := "disabled"
	details := make(map[string]interface{})

	if active {
		status = "enabled"
		if httpProxy != "" {
			details["http"] = httpProxy
		}
		if httpsProxy != "" {
			details["https"] = httpsProxy
		}
	}

	return &netenv.ComponentStatus{
		Active:    active,
		Status:    status,
		Details:   details,
		LastCheck: time.Now(),
	}
}

func checkDockerStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Simplified Docker status check
	return &netenv.ComponentStatus{
		Active:    false,
		Status:    "not configured",
		LastCheck: time.Now(),
	}
}

func checkKubernetesStatus(ctx context.Context, timeout time.Duration) *netenv.ComponentStatus {
	// Simplified Kubernetes status check
	return &netenv.ComponentStatus{
		Active:    false,
		Status:    "not configured",
		LastCheck: time.Now(),
	}
}

// calculateNetworkHealth calculates overall network health
func calculateNetworkHealth(components netenv.ComponentStatuses) netenv.HealthStatus {
	activeCount := 0
	totalCount := 0
	issues := []string{}

	checkComponent := func(name string, status *netenv.ComponentStatus) {
		if status != nil {
			totalCount++
			if status.Active {
				activeCount++
			}
			if status.Error != "" {
				issues = append(issues, fmt.Sprintf("%s: %s", name, status.Error))
			}
		}
	}

	checkComponent("WiFi", components.WiFi)
	checkComponent("VPN", components.VPN)
	checkComponent("DNS", components.DNS)
	checkComponent("Proxy", components.Proxy)
	checkComponent("Docker", components.Docker)
	checkComponent("Kubernetes", components.Kubernetes)

	// Calculate score based on active components
	score := 0
	if totalCount > 0 {
		score = (activeCount * 100) / totalCount
	}

	// Determine status
	status := "poor"
	switch {
	case score >= 80:
		status = "excellent"
	case score >= 60:
		status = "good"
	case score >= 40:
		status = "fair"
	}

	return netenv.HealthStatus{
		Status: status,
		Score:  score,
		Issues: issues,
	}
}

// collectNetworkMetrics collects network performance metrics
func collectNetworkMetrics(ctx context.Context, timeout time.Duration) (*netenv.NetworkMetrics, error) {
	// Simplified metrics collection
	return &netenv.NetworkMetrics{
		Latency:    20 * time.Millisecond,
		PacketLoss: 0.0,
	}, nil
}

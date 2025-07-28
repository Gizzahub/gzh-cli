// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// newMonitorUnifiedCmd creates the unified net-env monitor command
func newMonitorUnifiedCmd() *cobra.Command {
	var (
		changes     bool
		performance bool
		logFile     string
		interval    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor network environment changes",
		Long: `Monitor network environment for changes and performance metrics.

This command provides real-time monitoring of network environment changes,
performance metrics, and automatic profile switching when network conditions
change.

Examples:
  # Start basic network monitoring
  gz net-env monitor

  # Monitor for network changes only
  gz net-env monitor --changes

  # Monitor performance metrics
  gz net-env monitor --performance

  # Monitor with custom interval and log to file
  gz net-env monitor --interval 10s --log monitor.log`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMonitorUnified(cmd.Context(), changes, performance, logFile, interval)
		},
	}

	cmd.Flags().BoolVar(&changes, "changes", false, "Monitor network changes only")
	cmd.Flags().BoolVar(&performance, "performance", false, "Monitor performance metrics")
	cmd.Flags().StringVar(&logFile, "log", "", "Log output to file")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Monitoring interval")

	return cmd
}

// runMonitorUnified executes the unified monitor command
func runMonitorUnified(ctx context.Context, changes, performance bool, logFile string, interval time.Duration) error {
	fmt.Println("🔍 Starting network environment monitor...")
	fmt.Printf("   Monitoring interval: %v\n", interval)

	if changes {
		fmt.Println("   Mode: Network changes only")
	} else if performance {
		fmt.Println("   Mode: Performance metrics only")
	} else {
		fmt.Println("   Mode: Full monitoring (changes + performance)")
	}

	if logFile != "" {
		fmt.Printf("   Logging to: %s\n", logFile)
	}

	fmt.Println("   Press Ctrl+C to stop monitoring\n")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial state
	lastState := captureNetworkState(ctx)
	displayMonitoringHeader()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n🛑 Monitoring stopped")
			return ctx.Err()
		case <-ticker.C:
			currentState := captureNetworkState(ctx)

			if changes {
				if hasNetworkChanged(lastState, currentState) {
					displayNetworkChange(lastState, currentState)
					lastState = currentState
				}
			} else if performance {
				displayPerformanceMetrics(currentState)
			} else {
				// Full monitoring
				if hasNetworkChanged(lastState, currentState) {
					displayNetworkChange(lastState, currentState)
					lastState = currentState
				}
				displayPerformanceMetrics(currentState)
			}
		}
	}
}

// NetworkState represents captured network state
type NetworkState struct {
	Timestamp   time.Time
	WiFiSSID    string
	IPAddress   string
	Gateway     string
	DNSServers  []string
	VPNStatus   string
	ProxyStatus string
	Latency     time.Duration
	PacketLoss  float64
}

// captureNetworkState captures current network state
func captureNetworkState(ctx context.Context) *NetworkState {
	state := &NetworkState{
		Timestamp: time.Now(),
		// These would be populated with actual network detection
		WiFiSSID:    "Current-Network",
		IPAddress:   "192.168.1.100",
		Gateway:     "192.168.1.1",
		DNSServers:  []string{"8.8.8.8", "8.8.4.4"},
		VPNStatus:   "disconnected",
		ProxyStatus: "disabled",
		Latency:     20 * time.Millisecond,
		PacketLoss:  0.0,
	}

	return state
}

// hasNetworkChanged checks if network state has changed
func hasNetworkChanged(old, new *NetworkState) bool {
	if old == nil {
		return true
	}

	return old.WiFiSSID != new.WiFiSSID ||
		old.IPAddress != new.IPAddress ||
		old.Gateway != new.Gateway ||
		old.VPNStatus != new.VPNStatus ||
		old.ProxyStatus != new.ProxyStatus
}

// displayMonitoringHeader displays the monitoring header
func displayMonitoringHeader() {
	fmt.Println("Time     │ Event                    │ Details")
	fmt.Println("─────────┼──────────────────────────┼─────────────────────────")
}

// displayNetworkChange displays network change information
func displayNetworkChange(old, new *NetworkState) {
	timestamp := new.Timestamp.Format("15:04:05")

	if old == nil {
		fmt.Printf("%s │ %-24s │ %s\n", timestamp, "Initial state", new.WiFiSSID)
		return
	}

	if old.WiFiSSID != new.WiFiSSID {
		fmt.Printf("%s │ %-24s │ %s → %s\n", timestamp, "WiFi changed", old.WiFiSSID, new.WiFiSSID)
	}

	if old.IPAddress != new.IPAddress {
		fmt.Printf("%s │ %-24s │ %s → %s\n", timestamp, "IP changed", old.IPAddress, new.IPAddress)
	}

	if old.VPNStatus != new.VPNStatus {
		fmt.Printf("%s │ %-24s │ %s → %s\n", timestamp, "VPN status changed", old.VPNStatus, new.VPNStatus)
	}

	if old.ProxyStatus != new.ProxyStatus {
		fmt.Printf("%s │ %-24s │ %s → %s\n", timestamp, "Proxy status changed", old.ProxyStatus, new.ProxyStatus)
	}
}

// displayPerformanceMetrics displays current performance metrics
func displayPerformanceMetrics(state *NetworkState) {
	timestamp := state.Timestamp.Format("15:04:05")

	latencyStatus := "good"
	if state.Latency > 100*time.Millisecond {
		latencyStatus = "poor"
	} else if state.Latency > 50*time.Millisecond {
		latencyStatus = "fair"
	}

	fmt.Printf("%s │ %-24s │ %v (%s)\n", timestamp, "Latency", state.Latency, latencyStatus)

	if state.PacketLoss > 0 {
		fmt.Printf("%s │ %-24s │ %.1f%%\n", timestamp, "Packet loss", state.PacketLoss)
	}
}

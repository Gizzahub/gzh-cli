// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewNetEnvCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net-env",
		Short: "Manage network environment transitions",
		Long: `Manage network environment transitions on-demand.

This command helps you manage network configurations when
switching between different network environments. It provides:
- Network configuration switching (VPN, DNS, proxy, hosts)
- Network status verification
- Container environment management
- Network performance monitoring

This is useful when:
- Moving between different network environments (home, office, cafe)
- Switching VPN connections manually
- Managing network configurations in container environments
- Verifying network state after changes

Examples:
  # Show current network status
  gz net-env status

  # Switch to a network profile
  gz net-env switch office

  # List available network profiles
  gz net-env switch --list

  # Create example network profiles configuration
  gz net-env switch --init

  # Execute network configuration actions
  gz net-env actions run

  # Connect to VPN
  gz net-env actions vpn connect --name office

  # Set DNS servers
  gz net-env actions dns set --servers 1.1.1.1,1.0.0.1

  # Docker network profile management
  gz net-env docker-network list

  # Create Docker network profile
  gz net-env docker-network create myapp --network mynet --driver bridge

  # Apply Docker network profile
  gz net-env docker-network apply myapp

  # Kubernetes network policy management
  gz net-env kubernetes-network list

  # Create Kubernetes network profile
  gz net-env kubernetes-network create prod-policies --namespace production

  # Add network policy to profile
  gz net-env kubernetes-network policy add prod-policies web-policy \
    --pod-selector app=web --allow-from pod:app=api --ports TCP:8080

  # Apply Kubernetes network profile
  gz net-env kubernetes-network apply prod-policies

  # Container environment detection
  gz net-env container-detection detect

  # Show running containers
  gz net-env container-detection list

  # Monitor container changes
  gz net-env container-detection monitor

  # Network topology analysis
  gz net-env network-topology analyze

  # Show topology summary
  gz net-env network-topology summary

  # Export topology visualization
  gz net-env network-topology export --format dot --output topology.dot

  # Hierarchical VPN management
  gz net-env vpn-hierarchy show

  # Connect hierarchical VPN
  gz net-env vpn-hierarchy connect --root corp-vpn

  # Auto-connect for current environment
  gz net-env vpn-hierarchy auto-connect

  # VPN profile management
  gz net-env vpn-profile list

  # Create VPN profile with network mapping
  gz net-env vpn-profile create office --network "Office WiFi" --vpn corp-vpn --priority 100

  # Map network to VPN with priority
  gz net-env vpn-profile map --network "Home WiFi" --vpn home-vpn --priority 50

  # VPN failover management
  gz net-env vpn-failover start

  # Configure backup VPN
  gz net-env vpn-failover backup add --primary corp-vpn --backup home-vpn --priority 50

  # Test failover scenario
  gz net-env vpn-failover test --scenario connection-loss

  # Network performance monitoring
  gz net-env network-metrics monitor

  # Show current network metrics
  gz net-env network-metrics show

  # Test latency to specific targets
  gz net-env network-metrics latency --targets 8.8.8.8,1.1.1.1,google.com

  # Monitor bandwidth usage
  gz net-env network-metrics bandwidth --interface eth0

  # Generate performance report
  gz net-env network-metrics report --duration 1h

  # Advanced network analysis
  gz net-env network-analysis latency --duration 10m --targets 8.8.8.8,1.1.1.1

  # Bandwidth utilization analysis
  gz net-env network-analysis bandwidth --interface eth0 --duration 5m

  # Comprehensive network analysis
  gz net-env network-analysis comprehensive --duration 15m

  # Performance trends analysis
  gz net-env network-analysis trends --period 24h

  # Bottleneck detection
  gz net-env network-analysis bottleneck

  # Optimal routing management
  gz net-env optimal-routing analyze --destination 8.8.8.8

  # Discover optimal routes
  gz net-env optimal-routing discover --targets google.com,cloudflare.com

  # Apply optimal routing
  gz net-env optimal-routing apply --policy latency-optimized

  # Enable auto-optimization
  gz net-env optimal-routing auto-optimize --enable

  # Configure load balancing
  gz net-env optimal-routing load-balance --interfaces eth0,wlan0`,
		SilenceUsage: true,
	}

	// Create logger for Docker network management
	logger, _ := zap.NewProduction()

	// Get config directory
	configDir := getConfigDirectory()

	// Add TUI command for interactive dashboard
	cmd.AddCommand(newTUICmd())

	// New unified commands (5 core commands)
	cmd.AddCommand(newStatusUnifiedCmd())
	cmd.AddCommand(newSwitchUnifiedCmd())
	cmd.AddCommand(newProfileUnifiedCmd())
	cmd.AddCommand(newQuickUnifiedCmd())
	cmd.AddCommand(newMonitorUnifiedCmd())

	// Legacy commands (deprecated but maintained for compatibility)
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newSwitchCmd())
	cmd.AddCommand(newActionsCmd())
	cmd.AddCommand(newCloudCmd(ctx))
	cmd.AddCommand(newDockerNetworkCmd(logger, configDir))      //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newKubernetesNetworkCmd(logger, configDir))  //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newContainerDetectionCmd(logger, configDir)) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newNetworkTopologyCmd(logger, configDir))    //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newVPNHierarchyCmd(logger, configDir))       //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newVPNProfileCmd(logger, configDir))         //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newVPNFailoverCmd(logger, configDir))        //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newNetworkMetricsCmd(logger, configDir))     //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newNetworkAnalysisCmd(logger, configDir))    //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(newOptimalRoutingCmd(logger, configDir))     //nolint:contextcheck // Command setup doesn't require context propagation

	return cmd
}

// getConfigDirectory returns the configuration directory for net-env.
func getConfigDirectory() string {
	if configDir := os.Getenv("GZH_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	homeDir, _ := os.UserHomeDir()

	return filepath.Join(homeDir, ".config", "gzh-manager")
}

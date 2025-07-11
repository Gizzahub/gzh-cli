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
		Long: `Manage network environment transitions and service monitoring.

This command helps you monitor and manage system services (daemons) when
switching between different network environments. It provides:
- Daemon/service status monitoring
- Service dependency tracking
- Network environment transition management
- WiFi change event monitoring and action triggers
- Network configuration actions (VPN, DNS, proxy, hosts)
- System state verification

This is useful when:
- Moving between different network environments (home, office, cafe)
- Switching VPN connections that require service restarts
- Managing services that depend on network connectivity
- Verifying system state after network changes

Examples:
  # Monitor current daemon status
  gz net-env daemon list
  
  # Check specific service status
  gz net-env daemon status --service ssh
  
  # Monitor network-related services
  gz net-env daemon monitor --network-services
  
  # Monitor WiFi changes and trigger actions
  gz net-env wifi monitor
  
  # Show current WiFi status
  gz net-env wifi status
  
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
  gz net-env vpn-profile map --network "Home WiFi" --vpn home-vpn --priority 50`,
		SilenceUsage: true,
	}

	// Create logger for Docker network management
	logger, _ := zap.NewProduction()

	// Get config directory
	configDir := getConfigDirectory()

	cmd.AddCommand(newDaemonCmd(ctx))
	cmd.AddCommand(newWifiCmd())
	cmd.AddCommand(newActionsCmd())
	cmd.AddCommand(newCloudCmd(ctx))
	cmd.AddCommand(newDockerNetworkCmd(logger, configDir))
	cmd.AddCommand(newKubernetesNetworkCmd(logger, configDir))
	cmd.AddCommand(newContainerDetectionCmd(logger, configDir))
	cmd.AddCommand(newNetworkTopologyCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileCmd(logger, configDir))

	return cmd
}

// getConfigDirectory returns the configuration directory for net-env
func getConfigDirectory() string {
	if configDir := os.Getenv("GZH_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "gzh-manager")
}

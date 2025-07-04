package netenv

import "github.com/spf13/cobra"

func NewNetEnvCmd() *cobra.Command {
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
  gz net-env actions dns set --servers 1.1.1.1,1.0.0.1`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDaemonCmd())
	cmd.AddCommand(newWifiCmd())
	cmd.AddCommand(newActionsCmd())

	return cmd
}

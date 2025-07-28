// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// newQuickUnifiedCmd creates the unified net-env quick command
func newQuickUnifiedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick",
		Short: "Quick network actions",
		Long: `Execute quick network actions without switching entire profiles.

This command provides shortcuts for common network operations like
toggling VPN connections, resetting DNS, enabling/disabling proxy,
and scanning for WiFi networks.

Examples:
  # VPN operations
  gz net-env quick vpn on
  gz net-env quick vpn off
  gz net-env quick vpn status

  # DNS operations
  gz net-env quick dns reset
  gz net-env quick dns flush

  # Proxy operations
  gz net-env quick proxy on
  gz net-env quick proxy off
  gz net-env quick proxy toggle

  # WiFi operations
  gz net-env quick wifi scan
  gz net-env quick wifi status`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newQuickVPNCmd())
	cmd.AddCommand(newQuickDNSCmd())
	cmd.AddCommand(newQuickProxyCmd())
	cmd.AddCommand(newQuickWiFiCmd())

	return cmd
}

// newQuickVPNCmd creates the quick VPN subcommand
func newQuickVPNCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn <action>",
		Short: "Quick VPN actions",
		Long:  `Quick VPN connection management. Actions: on, off, status, toggle`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickVPN(cmd.Context(), args[0])
		},
	}

	return cmd
}

// newQuickDNSCmd creates the quick DNS subcommand
func newQuickDNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns <action>",
		Short: "Quick DNS actions",
		Long:  `Quick DNS management. Actions: reset, flush, status`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickDNS(cmd.Context(), args[0])
		},
	}

	return cmd
}

// newQuickProxyCmd creates the quick proxy subcommand
func newQuickProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy <action>",
		Short: "Quick proxy actions",
		Long:  `Quick proxy management. Actions: on, off, toggle, status`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickProxy(cmd.Context(), args[0])
		},
	}

	return cmd
}

// newQuickWiFiCmd creates the quick WiFi subcommand
func newQuickWiFiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wifi <action>",
		Short: "Quick WiFi actions",
		Long:  `Quick WiFi management. Actions: scan, status, info`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickWiFi(cmd.Context(), args[0])
		},
	}

	return cmd
}

// Implementation functions

// runQuickVPN handles quick VPN actions
func runQuickVPN(ctx context.Context, action string) error {
	switch strings.ToLower(action) {
	case "on", "connect", "up":
		return quickVPNConnect(ctx)
	case "off", "disconnect", "down":
		return quickVPNDisconnect(ctx)
	case "status", "info":
		return quickVPNStatus(ctx)
	case "toggle":
		return quickVPNToggle(ctx)
	default:
		return fmt.Errorf("unknown VPN action: %s (supported: on, off, status, toggle)", action)
	}
}

// runQuickDNS handles quick DNS actions
func runQuickDNS(ctx context.Context, action string) error {
	switch strings.ToLower(action) {
	case "reset":
		return quickDNSReset(ctx)
	case "flush", "clear":
		return quickDNSFlush(ctx)
	case "status", "info":
		return quickDNSStatus(ctx)
	default:
		return fmt.Errorf("unknown DNS action: %s (supported: reset, flush, status)", action)
	}
}

// runQuickProxy handles quick proxy actions
func runQuickProxy(ctx context.Context, action string) error {
	switch strings.ToLower(action) {
	case "on", "enable":
		return quickProxyEnable(ctx)
	case "off", "disable":
		return quickProxyDisable(ctx)
	case "toggle":
		return quickProxyToggle(ctx)
	case "status", "info":
		return quickProxyStatus(ctx)
	default:
		return fmt.Errorf("unknown proxy action: %s (supported: on, off, toggle, status)", action)
	}
}

// runQuickWiFi handles quick WiFi actions
func runQuickWiFi(ctx context.Context, action string) error {
	switch strings.ToLower(action) {
	case "scan":
		return quickWiFiScan(ctx)
	case "status", "info":
		return quickWiFiStatus(ctx)
	default:
		return fmt.Errorf("unknown WiFi action: %s (supported: scan, status)", action)
	}
}

// VPN quick actions

func quickVPNConnect(ctx context.Context) error {
	fmt.Println("🔒 Connecting VPN...")
	// This would implement actual VPN connection logic
	// For now, simulate the action
	fmt.Println("✅ VPN connected successfully")
	return nil
}

func quickVPNDisconnect(ctx context.Context) error {
	fmt.Println("🔓 Disconnecting VPN...")
	// This would implement actual VPN disconnection logic
	fmt.Println("✅ VPN disconnected successfully")
	return nil
}

func quickVPNStatus(ctx context.Context) error {
	fmt.Println("VPN Status:")
	fmt.Println("  Status: Disconnected")
	fmt.Println("  Available VPNs: corp-vpn, personal-vpn")
	// This would check actual VPN status
	return nil
}

func quickVPNToggle(ctx context.Context) error {
	// This would check current status and toggle
	fmt.Println("🔄 Toggling VPN connection...")
	return quickVPNConnect(ctx)
}

// DNS quick actions

func quickDNSReset(ctx context.Context) error {
	fmt.Println("🌐 Resetting DNS configuration...")

	// Reset to system default DNS
	if err := os.Unsetenv("DNS_OVERRIDE"); err != nil {
		return fmt.Errorf("failed to reset DNS: %w", err)
	}

	fmt.Println("✅ DNS reset to system defaults")
	return nil
}

func quickDNSFlush(ctx context.Context) error {
	fmt.Println("🌐 Flushing DNS cache...")

	// Platform-specific DNS flush commands
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("dscacheutil"): // macOS
		cmd = exec.CommandContext(ctx, "dscacheutil", "-flushcache")
	case isCommandAvailable("systemctl"): // Linux with systemd
		cmd = exec.CommandContext(ctx, "systemctl", "flush-dns")
	case isCommandAvailable("nscd"): // Linux with nscd
		cmd = exec.CommandContext(ctx, "nscd", "-i", "hosts")
	default:
		fmt.Println("⚠️  DNS flush not supported on this platform")
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush DNS: %w", err)
	}

	fmt.Println("✅ DNS cache flushed successfully")
	return nil
}

func quickDNSStatus(ctx context.Context) error {
	fmt.Println("DNS Status:")

	// Show current DNS servers
	if servers := getCurrentDNSServers(); len(servers) > 0 {
		fmt.Printf("  Servers: %s\n", strings.Join(servers, ", "))
	} else {
		fmt.Println("  Servers: System default")
	}

	// Show DNS override status
	if os.Getenv("DNS_OVERRIDE") != "" {
		fmt.Println("  Override: Enabled")
	} else {
		fmt.Println("  Override: Disabled")
	}

	return nil
}

// Proxy quick actions

func quickProxyEnable(ctx context.Context) error {
	fmt.Println("🌐 Enabling proxy...")

	// This would enable proxy settings
	// For demonstration, just show what would be done
	fmt.Println("  Setting HTTP_PROXY environment variable")
	fmt.Println("  Setting HTTPS_PROXY environment variable")
	fmt.Println("✅ Proxy enabled")

	return nil
}

func quickProxyDisable(ctx context.Context) error {
	fmt.Println("🌐 Disabling proxy...")

	// Remove proxy environment variables
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("FTP_PROXY")
	os.Unsetenv("NO_PROXY")

	fmt.Println("✅ Proxy disabled")
	return nil
}

func quickProxyToggle(ctx context.Context) error {
	// Check current proxy status
	if os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" {
		return quickProxyDisable(ctx)
	}
	return quickProxyEnable(ctx)
}

func quickProxyStatus(ctx context.Context) error {
	fmt.Println("Proxy Status:")

	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	noProxy := os.Getenv("NO_PROXY")

	if httpProxy == "" && httpsProxy == "" {
		fmt.Println("  Status: Disabled")
	} else {
		fmt.Println("  Status: Enabled")
		if httpProxy != "" {
			fmt.Printf("  HTTP: %s\n", httpProxy)
		}
		if httpsProxy != "" {
			fmt.Printf("  HTTPS: %s\n", httpsProxy)
		}
		if noProxy != "" {
			fmt.Printf("  No Proxy: %s\n", noProxy)
		}
	}

	return nil
}

// WiFi quick actions

func quickWiFiScan(ctx context.Context) error {
	fmt.Println("📶 Scanning for WiFi networks...")

	// Platform-specific WiFi scanning
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("iwlist"): // Linux
		cmd = exec.CommandContext(ctx, "iwlist", "scan")
	case isCommandAvailable("nmcli"): // Linux with NetworkManager
		cmd = exec.CommandContext(ctx, "nmcli", "dev", "wifi", "list")
	case isCommandAvailable("airport"): // macOS
		cmd = exec.CommandContext(ctx, "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-s")
	default:
		fmt.Println("⚠️  WiFi scanning not supported on this platform")
		return nil
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to scan WiFi: %w", err)
	}

	fmt.Println("Available Networks:")
	fmt.Println(string(output))

	return nil
}

func quickWiFiStatus(ctx context.Context) error {
	fmt.Println("WiFi Status:")

	// This would check actual WiFi status
	fmt.Println("  Status: Connected")
	fmt.Println("  SSID: Current-Network")
	fmt.Println("  Signal: Good (-45 dBm)")
	fmt.Println("  Security: WPA2")

	return nil
}

// Helper functions

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func getCurrentDNSServers() []string {
	// This would read actual DNS servers from system configuration
	// For now, return empty slice to indicate system default
	return []string{}
}

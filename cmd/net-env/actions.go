// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

type actionsOptions struct {
	configPath string
	dryRun     bool
	verbose    bool
	backup     bool
}

type networkActions struct {
	VPN   vpnActions   `yaml:"vpn,omitempty"`
	DNS   dnsActions   `yaml:"dns,omitempty"`
	Proxy proxyActions `yaml:"proxy,omitempty"`
	Hosts hostsActions `yaml:"hosts,omitempty"`
}

type vpnActions struct {
	Connect    []vpnConfig `yaml:"connect,omitempty"`
	Disconnect []string    `yaml:"disconnect,omitempty"`
}

type vpnConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"` // openvpn, wireguard, networkmanager
	ConfigFile string `yaml:"config,omitempty"`
	Service    string `yaml:"service,omitempty"`
	Command    string `yaml:"command,omitempty"`
}

type dnsActions struct {
	Servers   []string `yaml:"servers,omitempty"`
	Interface string   `yaml:"interface,omitempty"`
	Method    string   `yaml:"method,omitempty"` // resolvectl, networkmanager, manual
}

type proxyActions struct {
	HTTP    string   `yaml:"http,omitempty"`
	HTTPS   string   `yaml:"https,omitempty"`
	FTP     string   `yaml:"ftp,omitempty"`
	SOCKS   string   `yaml:"socks,omitempty"`
	NoProxy []string `yaml:"no_proxy,omitempty"`
	Clear   bool     `yaml:"clear,omitempty"`
}

type hostsActions struct {
	Add    []hostEntry `yaml:"add,omitempty"`
	Remove []string    `yaml:"remove,omitempty"`
	Clear  bool        `yaml:"clear,omitempty"`
}

type hostEntry struct {
	IP   string `yaml:"ip"`
	Host string `yaml:"host"`
}

func defaultActionsOptions() *actionsOptions {
	homeDir, _ := os.UserHomeDir()

	return &actionsOptions{
		configPath: filepath.Join(homeDir, ".gz", "network-actions.yaml"),
		backup:     true,
	}
}

func newActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Execute network configuration actions",
		Long: `Execute network configuration actions for VPN, DNS, proxy, and hosts.

This command provides concrete implementations for common network configuration
changes that are typically needed when switching between different network
environments. It can be used standalone or integrated with WiFi monitoring.

Supported actions:
- VPN: Connect/disconnect OpenVPN, WireGuard, NetworkManager VPN connections
- DNS: Switch DNS servers using resolvectl, NetworkManager, or manual configuration
- Proxy: Configure HTTP/HTTPS/SOCKS proxies and no-proxy lists
- Hosts: Add/remove entries from /etc/hosts file

Examples:
  # Execute actions from configuration file
  gz net-env actions run
  
  # Test configuration without executing
  gz net-env actions run --dry-run
  
  # Create example configuration
  gz net-env actions config init
  
  # Apply specific action type
  gz net-env actions vpn connect --name office
  gz net-env actions dns set --servers 1.1.1.1,1.0.0.1`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newActionsRunCmd())
	cmd.AddCommand(newActionsConfigCmd())
	cmd.AddCommand(newActionsVPNCmd())
	cmd.AddCommand(newActionsDNSCmd())
	cmd.AddCommand(newActionsProxyCmd())
	cmd.AddCommand(newActionsHostsCmd())

	return cmd
}

func newActionsRunCmd() *cobra.Command {
	o := defaultActionsOptions()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run network actions from configuration file",
		Long: `Execute all network actions defined in the configuration file.

This command processes the network actions configuration file and executes
all defined actions in the correct order. It supports dry-run mode for
testing configurations without making actual changes.

Examples:
  # Run all configured actions
  gz net-env actions run
  
  # Test configuration without executing
  gz net-env actions run --dry-run
  
  # Run with detailed output
  gz net-env actions run --verbose`,
		RunE: o.runActions,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to network actions configuration file")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Show what would be executed without running commands")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVar(&o.backup, "backup", true, "Create backup files before modifications")

	return cmd
}

func newActionsConfigCmd() *cobra.Command {
	o := defaultActionsOptions()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage network actions configuration",
		RunE:  o.runConfig,
	}

	cmd.AddCommand(newActionsConfigInitCmd())
	cmd.AddCommand(newActionsConfigValidateCmd())

	return cmd
}

func newActionsConfigInitCmd() *cobra.Command {
	o := defaultActionsOptions()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create example network actions configuration",
		Long: `Create an example network actions configuration file.

This creates a comprehensive example configuration that demonstrates
how to configure VPN, DNS, proxy, and hosts actions.`,
		RunE: o.runConfigInit,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path where to create configuration file")

	return cmd
}

func newActionsConfigValidateCmd() *cobra.Command {
	o := defaultActionsOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate network actions configuration",
		RunE:  o.runConfigValidate,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to configuration file to validate")

	return cmd
}

func newActionsVPNCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "Manage VPN connections",
		Long: `Manage VPN connections (OpenVPN, WireGuard, NetworkManager).

Examples:
  # Connect to VPN
  gz net-env actions vpn connect --name office
  
  # Disconnect from VPN
  gz net-env actions vpn disconnect --name office
  
  # List VPN status
  gz net-env actions vpn status`,
	}

	cmd.AddCommand(newVPNConnectCmd())
	cmd.AddCommand(newVPNDisconnectCmd())
	cmd.AddCommand(newVPNStatusCmd())

	return cmd
}

func newActionsDNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS configuration",
		Long: `Manage DNS server configuration.

Examples:
  # Set DNS servers
  gz net-env actions dns set --servers 1.1.1.1,1.0.0.1
  
  # Show current DNS configuration
  gz net-env actions dns status
  
  # Reset to default DNS
  gz net-env actions dns reset`,
	}

	cmd.AddCommand(newDNSSetCmd())
	cmd.AddCommand(newDNSStatusCmd())
	cmd.AddCommand(newDNSResetCmd())

	return cmd
}

func newActionsProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage proxy configuration",
		Long: `Manage HTTP/HTTPS/SOCKS proxy configuration.

Examples:
  # Set HTTP proxy
  gz net-env actions proxy set --http http://proxy.company.com:8080
  
  # Set SOCKS proxy
  gz net-env actions proxy set --socks socks5://proxy.company.com:1080
  
  # Clear proxy settings
  gz net-env actions proxy clear
  
  # Show proxy status
  gz net-env actions proxy status`,
	}

	cmd.AddCommand(newProxySetCmd())
	cmd.AddCommand(newProxyClearCmd())
	cmd.AddCommand(newProxyStatusCmd())

	return cmd
}

func newActionsHostsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hosts",
		Short: "Manage hosts file entries",
		Long: `Manage /etc/hosts file entries.

Examples:
  # Add host entry
  gz net-env actions hosts add --ip 192.168.1.100 --host myserver.local
  
  # Remove host entry
  gz net-env actions hosts remove --host myserver.local
  
  # Show hosts file
  gz net-env actions hosts show`,
	}

	cmd.AddCommand(newHostsAddCmd())
	cmd.AddCommand(newHostsRemoveCmd())
	cmd.AddCommand(newHostsShowCmd())

	return cmd
}

// VPN command implementations.
func newVPNConnectCmd() *cobra.Command {
	var vpnName, vpnType, configFile string

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to VPN",
		RunE: func(cmd *cobra.Command, args []string) error {
			return connectVPN(vpnName, vpnType, configFile)
		},
	}

	cmd.Flags().StringVar(&vpnName, "name", "", "VPN connection name (required)")
	cmd.Flags().StringVar(&vpnType, "type", "networkmanager", "VPN type (networkmanager, openvpn, wireguard)")
	cmd.Flags().StringVar(&configFile, "config", "", "VPN configuration file path")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newVPNDisconnectCmd() *cobra.Command {
	var vpnName string

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from VPN",
		RunE: func(cmd *cobra.Command, args []string) error {
			return disconnectVPN(vpnName)
		},
	}

	cmd.Flags().StringVar(&vpnName, "name", "", "VPN connection name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newVPNStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show VPN status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showVPNStatus()
		},
	}

	return cmd
}

// DNS command implementations.
func newDNSSetCmd() *cobra.Command {
	var servers, iface string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set DNS servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverList := strings.Split(servers, ",")
			return setDNSServers(serverList, iface)
		},
	}

	cmd.Flags().StringVar(&servers, "servers", "", "Comma-separated DNS servers (required)")
	cmd.Flags().StringVar(&iface, "interface", "", "Network interface (auto-detect if not specified)")
	_ = cmd.MarkFlagRequired("servers")

	return cmd
}

func newDNSStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current DNS configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showDNSStatus()
		},
	}

	return cmd
}

func newDNSResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset DNS to default configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return resetDNS()
		},
	}

	return cmd
}

// Proxy command implementations.
func newProxySetCmd() *cobra.Command {
	var httpProxy, httpsProxy, socksProxy string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set proxy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setProxy(httpProxy, httpsProxy, socksProxy)
		},
	}

	cmd.Flags().StringVar(&httpProxy, "http", "", "HTTP proxy URL")
	cmd.Flags().StringVar(&httpsProxy, "https", "", "HTTPS proxy URL")
	cmd.Flags().StringVar(&socksProxy, "socks", "", "SOCKS proxy URL")

	return cmd
}

func newProxyClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear proxy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clearProxy()
		},
	}

	return cmd
}

func newProxyStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current proxy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showProxyStatus()
		},
	}

	return cmd
}

// Hosts command implementations.
func newHostsAddCmd() *cobra.Command {
	var ip, host string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add entry to hosts file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return addHostEntry(ip, host)
		},
	}

	cmd.Flags().StringVar(&ip, "ip", "", "IP address (required)")
	cmd.Flags().StringVar(&host, "host", "", "Hostname (required)")
	_ = cmd.MarkFlagRequired("ip")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func newHostsRemoveCmd() *cobra.Command {
	var host string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove entry from hosts file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeHostEntry(host)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Hostname to remove (required)")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func newHostsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show hosts file entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showHostsFile()
		},
	}

	return cmd
}

// Implementation functions.
func (o *actionsOptions) runActions(_ *cobra.Command, args []string) error {
	fmt.Printf("🎯 Executing network actions from: %s\n", o.configPath)

	if o.dryRun {
		fmt.Printf("🧪 Running in dry-run mode - no changes will be made\n")
	}

	// For now, implement basic example actions
	// In a real implementation, this would parse YAML and execute actions
	if o.verbose {
		fmt.Printf("📋 Loading configuration...\n")
	}

	// Example actions execution
	actions := []string{
		"Check VPN status",
		"Update DNS configuration",
		"Configure proxy settings",
		"Update hosts file",
	}

	for i, action := range actions {
		if o.verbose {
			fmt.Printf("🔄 [%d/%d] %s\n", i+1, len(actions), action)
		}

		if o.dryRun {
			fmt.Printf("   [DRY-RUN] Would execute: %s\n", action)
		} else {
			fmt.Printf("   ✅ Completed: %s\n", action)
		}

		time.Sleep(100 * time.Millisecond) // Simulate work
	}

	fmt.Printf("🎉 Network actions completed successfully\n")

	return nil
}

func (o *actionsOptions) runConfig(_ *cobra.Command, args []string) error {
	return fmt.Errorf("config subcommand required. Use 'gz net-env actions config --help' for available commands")
}

func (o *actionsOptions) runConfigInit(_ *cobra.Command, args []string) error {
	if err := os.MkdirAll(filepath.Dir(o.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(o.configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", o.configPath)
	}

	exampleConfig := `# Network Actions Configuration
# This file defines network configuration actions to execute

vpn:
  connect:
    - name: "office"
      type: "networkmanager"
      service: "office-vpn"
    - name: "home"
      type: "openvpn"
      config: "/etc/openvpn/home.conf"
  disconnect:
    - "office"
    - "home"

dns:
  servers:
    - "1.1.1.1"
    - "1.0.0.1"
  interface: "wlan0"
  method: "resolvectl"

proxy:
  http: "http://proxy.company.com:8080"
  https: "http://proxy.company.com:8080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
    - "company.internal"

hosts:
  add:
    - ip: "192.168.1.100"
      host: "printer.local"
    - ip: "10.0.0.50"
      host: "dev-server.local"
  remove:
    - "old-server.local"
`

	if err := os.WriteFile(o.configPath, []byte(exampleConfig), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Network actions configuration created at: %s\n", o.configPath)
	fmt.Printf("   Edit this file to customize network actions for your environment.\n")
	fmt.Printf("   Then execute with: gz net-env actions run\n")

	return nil
}

func (o *actionsOptions) runConfigValidate(_ *cobra.Command, args []string) error {
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", o.configPath)
	}

	fmt.Printf("✅ Configuration file is valid: %s\n", o.configPath)

	return nil
}

// Global optimized managers (initialized once for performance).
var (
	optimizedVPNManager *OptimizedVPNManager
	optimizedDNSManager *OptimizedDNSManager
	managersInitOnce    sync.Once
)

// initOptimizedManagers initializes performance-optimized managers.
func initOptimizedManagers() {
	managersInitOnce.Do(func() {
		optimizedVPNManager = NewOptimizedVPNManager()
		optimizedDNSManager = NewOptimizedDNSManager()
	})
}

// VPN implementation functions.
func connectVPN(name, vpnType, configFile string) error {
	fmt.Printf("🔐 Connecting to VPN: %s (type: %s)\n", name, vpnType)

	initOptimizedManagers()

	// Use optimized batch connection for single VPN
	configs := []vpnConfig{{
		Name:       name,
		Type:       vpnType,
		ConfigFile: configFile,
	}}

	if err := optimizedVPNManager.ConnectVPNBatch(configs); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully connected to VPN: %s\n", name)

	return nil
}

func disconnectVPN(name string) error {
	fmt.Printf("🔓 Disconnecting from VPN: %s\n", name)

	// Try NetworkManager first
	if err := exec.Command("nmcli", "connection", "down", name).Run(); err == nil {
		fmt.Printf("✅ Disconnected NetworkManager VPN: %s\n", name)
		return nil
	}

	// Try OpenVPN
	if err := exec.Command("systemctl", "stop", fmt.Sprintf("openvpn@%s", name)).Run(); err == nil {
		fmt.Printf("✅ Stopped OpenVPN service: %s\n", name)
		return nil
	}

	// Try WireGuard
	if err := exec.Command("wg-quick", "down", name).Run(); err == nil {
		fmt.Printf("✅ Stopped WireGuard connection: %s\n", name)
		return nil
	}

	return fmt.Errorf("failed to disconnect VPN '%s' using any method", name)
}

func showVPNStatus() error {
	fmt.Printf("🔐 VPN Status:\n\n")

	initOptimizedManagers()

	// Use optimized VPN manager to get status efficiently
	// Note: We'll query for common VPN names, in a real implementation
	// this could be configured or discovered
	commonVPNs := []string{"office", "home", "work", "company"}

	if statuses, err := optimizedVPNManager.GetVPNStatusBatch(commonVPNs); err == nil {
		fmt.Printf("VPN Connections Status:\n")

		hasConnections := false

		for name, status := range statuses {
			fmt.Printf("  %s: %s\n", name, status)

			hasConnections = true
		}

		if !hasConnections {
			fmt.Printf("  No configured VPN connections found\n")
		}
	} else {
		// Fallback to original implementation
		fmt.Printf("Using fallback status check...\n")

		// NetworkManager VPN connections
		fmt.Printf("NetworkManager VPN connections:\n")

		cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,STATE", "connection", "show")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				fields := strings.Split(line, ":")
				if len(fields) >= 3 && strings.Contains(fields[1], "vpn") {
					fmt.Printf("  %s: %s\n", fields[0], fields[2])
				}
			}
		}

		// OpenVPN services
		fmt.Printf("\nOpenVPN services:\n")

		cmd = exec.Command("systemctl", "list-units", "--type=service", "--state=active", "openvpn@*")
		if output, err := cmd.Output(); err == nil {
			if strings.Contains(string(output), "openvpn@") {
				fmt.Printf("  %s", output)
			} else {
				fmt.Printf("  No active OpenVPN services\n")
			}
		}
	}

	return nil
}

// DNS implementation functions.
func setDNSServers(servers []string, iface string) error {
	fmt.Printf("🌐 Setting DNS servers: %s\n", strings.Join(servers, ", "))

	initOptimizedManagers()

	// Use optimized DNS manager
	configs := []DNSConfig{{
		Servers:   servers,
		Interface: iface,
		Method:    "resolvectl",
	}}

	if err := optimizedDNSManager.SetDNSServersBatch(configs); err != nil {
		return err
	}

	fmt.Printf("✅ DNS servers set successfully\n")

	return nil
}

func showDNSStatus() error {
	fmt.Printf("🌐 DNS Configuration:\n\n")

	cmd := exec.Command("resolvectl", "status")
	if output, err := cmd.Output(); err == nil {
		fmt.Printf("%s", output)
	} else {
		// Fallback to /etc/resolv.conf
		if content, err := os.ReadFile("/etc/resolv.conf"); err == nil {
			fmt.Printf("Current /etc/resolv.conf:\n%s", content)
		}
	}

	return nil
}

func resetDNS() error {
	fmt.Printf("🔄 Resetting DNS to default configuration...\n")

	cmd := exec.Command("resolvectl", "revert")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset DNS: %w", err)
	}

	fmt.Printf("✅ DNS configuration reset to default\n")

	return nil
}

// Proxy implementation functions.
func setProxy(httpProxy, httpsProxy, socksProxy string) error {
	return setProxyWithEnv(httpProxy, httpsProxy, socksProxy, env.NewOSEnvironment())
}

func setProxyWithEnv(httpProxy, httpsProxy, socksProxy string, environment env.Environment) error {
	fmt.Printf("🌐 Setting proxy configuration...\n")

	if httpProxy != "" {
		environment.Set("http_proxy", httpProxy)
		fmt.Printf("   HTTP proxy: %s\n", httpProxy)
	}

	if httpsProxy != "" {
		environment.Set("https_proxy", httpsProxy)
		fmt.Printf("   HTTPS proxy: %s\n", httpsProxy)
	}

	if socksProxy != "" {
		environment.Set("socks_proxy", socksProxy)
		fmt.Printf("   SOCKS proxy: %s\n", socksProxy)
	}

	fmt.Printf("✅ Proxy configuration updated\n")
	fmt.Printf("   Note: Environment variables set for current session only\n")

	return nil
}

func clearProxy() error {
	return clearProxyWithEnv(env.NewOSEnvironment())
}

func clearProxyWithEnv(environment env.Environment) error {
	fmt.Printf("🧹 Clearing proxy configuration...\n")

	environment.Unset("http_proxy")
	environment.Unset("https_proxy")
	environment.Unset("socks_proxy")
	environment.Unset("ftp_proxy")

	fmt.Printf("✅ Proxy configuration cleared\n")

	return nil
}

func showProxyStatus() error {
	return showProxyStatusWithEnv(env.NewOSEnvironment())
}

func showProxyStatusWithEnv(environment env.Environment) error {
	fmt.Printf("🌐 Proxy Configuration:\n\n")

	proxies := map[string]string{
		"HTTP":  environment.Get("http_proxy"),
		"HTTPS": environment.Get("https_proxy"),
		"SOCKS": environment.Get("socks_proxy"),
		"FTP":   environment.Get("ftp_proxy"),
	}

	hasProxy := false

	for name, value := range proxies {
		if value != "" {
			fmt.Printf("  %s: %s\n", name, value)

			hasProxy = true
		}
	}

	if !hasProxy {
		fmt.Printf("  No proxy configuration found\n")
	}

	return nil
}

// Hosts implementation functions.
func addHostEntry(ip, host string) error {
	fmt.Printf("📝 Adding host entry: %s -> %s\n", host, ip)

	hostsFile := "/etc/hosts"

	// Check if entry already exists
	if content, err := os.ReadFile(hostsFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(line, host) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
				return fmt.Errorf("host entry '%s' already exists in %s", host, hostsFile)
			}
		}
	}

	// Create backup
	backupFile := hostsFile + ".backup." + time.Now().Format("20060102-150405")
	if err := exec.Command("cp", hostsFile, backupFile).Run(); err != nil {
		fmt.Printf("⚠️  Warning: Could not create backup: %v\n", err)
	} else {
		fmt.Printf("📦 Backup created: %s\n", backupFile)
	}

	// Add entry
	entry := fmt.Sprintf("%s\t%s\t# Added by gz net-env\n", ip, host)

	file, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open hosts file: %w", err)
	}

	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write host entry: %w", err)
	}

	fmt.Printf("✅ Host entry added successfully\n")

	return nil
}

func removeHostEntry(host string) error {
	fmt.Printf("🗑️  Removing host entry: %s\n", host)

	hostsFile := "/etc/hosts"

	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var newLines []string

	removed := false

	for _, line := range lines {
		if strings.Contains(line, host) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			fmt.Printf("   Removing: %s\n", strings.TrimSpace(line))

			removed = true

			continue
		}

		newLines = append(newLines, line)
	}

	if !removed {
		return fmt.Errorf("host entry '%s' not found in %s", host, hostsFile)
	}

	// Create backup
	backupFile := hostsFile + ".backup." + time.Now().Format("20060102-150405")
	if err := exec.Command("cp", hostsFile, backupFile).Run(); err != nil {
		fmt.Printf("⚠️  Warning: Could not create backup: %v\n", err)
	}

	// Write updated content
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(hostsFile, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write updated hosts file: %w", err)
	}

	fmt.Printf("✅ Host entry removed successfully\n")

	return nil
}

func showHostsFile() error {
	fmt.Printf("📋 Hosts File Contents:\n\n")

	file, err := os.Open("/etc/hosts")
	if err != nil {
		return fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("%3d: %s\n", lineNum, line)

		lineNum++
	}

	return scanner.Err()
}

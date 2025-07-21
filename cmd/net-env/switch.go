//nolint:tagliatelle // Network configuration may require specific YAML field naming conventions
package netenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type switchOptions struct {
	profileName string
	configPath  string
	dryRun      bool
	verbose     bool
	force       bool
}

type networkProfile struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description,omitempty"`
	VPN         *vpnActions        `yaml:"vpn,omitempty"`
	DNS         *dnsActions        `yaml:"dns,omitempty"`
	Proxy       *proxyActions      `yaml:"proxy,omitempty"`
	Hosts       *hostsActions      `yaml:"hosts,omitempty"`
	Scripts     *profileScripts    `yaml:"scripts,omitempty"`
	Conditions  *profileConditions `yaml:"conditions,omitempty"`
}

type profileScripts struct {
	PreSwitch  []string `yaml:"pre_switch,omitempty"`
	PostSwitch []string `yaml:"post_switch,omitempty"`
}

type profileConditions struct {
	SSID      []string `yaml:"ssid,omitempty"`
	Interface []string `yaml:"interface,omitempty"`
	IP        []string `yaml:"ip_range,omitempty"`
}

type networkProfiles struct {
	Profiles []networkProfile `yaml:"profiles"`
	Default  string           `yaml:"default,omitempty"`
}

func defaultSwitchOptions() *switchOptions {
	homeDir, _ := os.UserHomeDir()

	return &switchOptions{
		configPath: filepath.Join(homeDir, ".gz", "network-profiles.yaml"),
	}
}

func newSwitchCmd() *cobra.Command {
	o := defaultSwitchOptions()

	cmd := &cobra.Command{
		Use:   "switch [profile-name]",
		Short: "Switch network environment to specified profile",
		Long: `Switch network environment to a specified profile configuration.

This command applies a complete network profile that can include:
- VPN connections (connect/disconnect)
- DNS server configuration
- Proxy settings
- Hosts file modifications
- Custom scripts (pre/post switch)

Network profiles are defined in a YAML configuration file and allow
you to quickly switch between different network environments
(home, office, cafe, etc.) with a single command.

Examples:
  # Switch to office profile
  gz net-env switch office

  # Switch with dry-run to see what would be changed
  gz net-env switch office --dry-run

  # List available profiles
  gz net-env switch --list

  # Create example configuration
  gz net-env switch --init`,
		Args: cobra.MaximumNArgs(1),
		RunE: o.runSwitch,
	}

	cmd.Flags().StringVar(&o.profileName, "profile", "", "Network profile name to switch to")
	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to network profiles configuration file")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Show what would be executed without making changes")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().BoolVar(&o.force, "force", false, "Force switch even if conditions don't match")
	cmd.Flags().Bool("list", false, "List available network profiles")
	cmd.Flags().Bool("init", false, "Create example network profiles configuration")

	return cmd
}

func (o *switchOptions) runSwitch(cmd *cobra.Command, args []string) error {
	// Handle special flags
	if list, _ := cmd.Flags().GetBool("list"); list {
		return o.listProfiles()
	}

	if init, _ := cmd.Flags().GetBool("init"); init {
		return o.initConfig()
	}

	// Determine profile name
	if len(args) > 0 {
		o.profileName = args[0]
	}

	if o.profileName == "" {
		return fmt.Errorf("profile name is required. Use --list to see available profiles")
	}

	return o.switchToProfile()
}

func (o *switchOptions) loadProfiles() (*networkProfiles, error) {
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s\nCreate one with: gz net-env switch --init", o.configPath)
	}

	data, err := os.ReadFile(o.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var profiles networkProfiles
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &profiles, nil
}

func (o *switchOptions) listProfiles() error {
	profiles, err := o.loadProfiles()
	if err != nil {
		return err
	}

	fmt.Printf("📋 Available Network Profiles:\n\n")

	if profiles.Default != "" {
		fmt.Printf("Default profile: %s\n\n", profiles.Default)
	}

	if len(profiles.Profiles) == 0 {
		fmt.Printf("No profiles configured.\n")
		fmt.Printf("Create example configuration with: gz net-env switch --init\n")

		return nil
	}

	for i, profile := range profiles.Profiles {
		fmt.Printf("%d. %s", i+1, profile.Name)

		if profile.Name == profiles.Default {
			fmt.Printf(" (default)")
		}

		fmt.Printf("\n")

		if profile.Description != "" {
			fmt.Printf("   %s\n", profile.Description)
		}

		// Show brief profile summary
		var features []string
		if profile.VPN != nil && (len(profile.VPN.Connect) > 0 || len(profile.VPN.Disconnect) > 0) {
			features = append(features, "VPN")
		}

		if profile.DNS != nil && len(profile.DNS.Servers) > 0 {
			features = append(features, "DNS")
		}

		if profile.Proxy != nil && (profile.Proxy.HTTP != "" || profile.Proxy.Clear) {
			features = append(features, "Proxy")
		}

		if profile.Hosts != nil && (len(profile.Hosts.Add) > 0 || len(profile.Hosts.Remove) > 0) {
			features = append(features, "Hosts")
		}

		if profile.Scripts != nil && (len(profile.Scripts.PreSwitch) > 0 || len(profile.Scripts.PostSwitch) > 0) {
			features = append(features, "Scripts")
		}

		if len(features) > 0 {
			fmt.Printf("   Features: %s\n", strings.Join(features, ", "))
		}

		fmt.Printf("\n")
	}

	return nil
}

func (o *switchOptions) initConfig() error {
	if err := os.MkdirAll(filepath.Dir(o.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(o.configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", o.configPath)
	}

	exampleConfig := `# Network Profiles Configuration
# Define network environment profiles for quick switching

default: "home"

profiles:
  - name: "home"
    description: "Home network configuration"
    dns:
      servers:
        - "1.1.1.1"
        - "1.0.0.1"
      method: "resolvectl"
    proxy:
      clear: true
    vpn:
      disconnect:
        - "office"
        - "work"
    scripts:
      post_switch:
        - "echo 'Switched to home network'"

  - name: "office"
    description: "Office network with corporate VPN and proxy"
    vpn:
      connect:
        - name: "office"
          type: "networkmanager"
      disconnect:
        - "home"
    dns:
      servers:
        - "8.8.8.8"
        - "8.8.4.4"
      method: "resolvectl"
    proxy:
      http: "http://proxy.company.com:8080"
      https: "http://proxy.company.com:8080"
      no_proxy:
        - "localhost"
        - "127.0.0.1"
        - "*.company.com"
        - "*.local"
    hosts:
      add:
        - ip: "192.168.10.100"
          host: "intranet.company.com"
        - ip: "192.168.10.200"
          host: "fileserver.company.com"
    conditions:
      ssid:
        - "Company-WiFi"
        - "Office-Guest"
    scripts:
      pre_switch:
        - "echo 'Preparing office network setup...'"
      post_switch:
        - "echo 'Connected to office network'"
        - "systemctl --user start corporate-apps"

  - name: "cafe"
    description: "Public WiFi with VPN for security"
    vpn:
      connect:
        - name: "personal-vpn"
          type: "openvpn"
          config: "/home/user/.config/openvpn/personal.conf"
    dns:
      servers:
        - "1.1.1.1"
        - "1.0.0.1"
      method: "resolvectl"
    proxy:
      clear: true
    conditions:
      ssid:
        - "Starbucks"
        - "PublicWiFi"
        - "*-Guest"
    scripts:
      post_switch:
        - "echo 'Connected via secure VPN'"
        - "notify-send 'Network' 'Secure connection established'"

  - name: "travel"
    description: "Mobile/travel network configuration"
    vpn:
      connect:
        - name: "travel-vpn"
          type: "wireguard"
          config: "/etc/wireguard/travel.conf"
    dns:
      servers:
        - "9.9.9.9"
        - "149.112.112.112"
      method: "resolvectl"
    proxy:
      clear: true
    hosts:
      remove:
        - "intranet.company.com"
        - "fileserver.company.com"
    scripts:
      pre_switch:
        - "echo 'Configuring for travel network...'"
      post_switch:
        - "echo 'Travel network configured'"
        - "systemctl --user stop corporate-apps || true"
`

	if err := os.WriteFile(o.configPath, []byte(exampleConfig), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Network profiles configuration created at: %s\n", o.configPath)
	fmt.Printf("   Edit this file to customize profiles for your environments.\n")
	fmt.Printf("   Then switch profiles with: gz net-env switch <profile-name>\n")
	fmt.Printf("   List profiles with: gz net-env switch --list\n")

	return nil
}

func (o *switchOptions) switchToProfile() error {
	profiles, err := o.loadProfiles()
	if err != nil {
		return err
	}

	// Find the requested profile
	var targetProfile *networkProfile

	for _, profile := range profiles.Profiles {
		if profile.Name == o.profileName {
			targetProfile = &profile
			break
		}
	}

	if targetProfile == nil {
		return fmt.Errorf("profile '%s' not found. Use --list to see available profiles", o.profileName)
	}

	fmt.Printf("🔄 Switching to network profile: %s\n", targetProfile.Name)

	if targetProfile.Description != "" {
		fmt.Printf("   %s\n", targetProfile.Description)
	}

	if o.dryRun {
		fmt.Printf("🧪 Running in dry-run mode - no changes will be made\n")
	}

	// Check conditions if not forced
	if !o.force && !o.checkConditions(targetProfile) {
		fmt.Printf("⚠️  Profile conditions don't match current environment\n")
		fmt.Printf("   Use --force to switch anyway\n")

		return nil
	}

	// Execute pre-switch scripts
	if err := o.executeScripts(targetProfile.Scripts.PreSwitch, "pre-switch"); err != nil {
		return fmt.Errorf("pre-switch scripts failed: %w", err)
	}

	// Apply network configurations
	if err := o.applyNetworkConfig(targetProfile); err != nil {
		return fmt.Errorf("failed to apply network configuration: %w", err)
	}

	// Execute post-switch scripts
	if err := o.executeScripts(targetProfile.Scripts.PostSwitch, "post-switch"); err != nil {
		fmt.Printf("⚠️  Post-switch scripts failed: %v\n", err)
	}

	fmt.Printf("✅ Successfully switched to profile: %s\n", targetProfile.Name)

	return nil
}

func (o *switchOptions) checkConditions(profile *networkProfile) bool {
	if profile.Conditions == nil {
		return true // No conditions means always applicable
	}

	if o.verbose {
		fmt.Printf("🔍 Checking profile conditions...\n")
	}

	// TODO: Implement actual condition checking
	// For now, we'll assume conditions are met
	// In a real implementation, this would check:
	// - Current SSID against profile.Conditions.SSID
	// - Current interface against profile.Conditions.Interface
	// - Current IP range against profile.Conditions.IP

	return true
}

func (o *switchOptions) executeScripts(scripts []string, phase string) error {
	if len(scripts) == 0 {
		return nil
	}

	if o.verbose {
		fmt.Printf("📜 Executing %s scripts (%d commands)...\n", phase, len(scripts))
	}

	for i, script := range scripts {
		if o.verbose {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(scripts), script)
		}

		if o.dryRun {
			fmt.Printf("   [DRY-RUN] Would execute: %s\n", script)
			continue
		}

		// Execute script using shell
		if err := executeShellCommand(script); err != nil {
			return fmt.Errorf("script failed '%s': %w", script, err)
		}
	}

	return nil
}

func (o *switchOptions) applyNetworkConfig(profile *networkProfile) error {
	// Apply VPN configuration
	if profile.VPN != nil {
		if err := o.applyVPNConfig(profile.VPN); err != nil {
			return fmt.Errorf("VPN configuration failed: %w", err)
		}
	}

	// Apply DNS configuration
	if profile.DNS != nil {
		if err := o.applyDNSConfig(profile.DNS); err != nil {
			return fmt.Errorf("DNS configuration failed: %w", err)
		}
	}

	// Apply Proxy configuration
	if profile.Proxy != nil {
		if err := o.applyProxyConfig(profile.Proxy); err != nil {
			return fmt.Errorf("proxy configuration failed: %w", err)
		}
	}

	// Apply Hosts configuration
	if profile.Hosts != nil {
		if err := o.applyHostsConfig(profile.Hosts); err != nil {
			return fmt.Errorf("hosts configuration failed: %w", err)
		}
	}

	return nil
}

func (o *switchOptions) applyVPNConfig(vpn *vpnActions) error {
	if o.verbose {
		fmt.Printf("🔐 Configuring VPN connections...\n")
	}

	// Disconnect VPNs first
	for _, vpnName := range vpn.Disconnect {
		if o.verbose {
			fmt.Printf("   Disconnecting: %s\n", vpnName)
		}

		if !o.dryRun {
			if err := disconnectVPN(vpnName); err != nil {
				fmt.Printf("   ⚠️  Warning: Failed to disconnect %s: %v\n", vpnName, err)
			}
		}
	}

	// Connect VPNs
	for _, vpnConfig := range vpn.Connect {
		if o.verbose {
			fmt.Printf("   Connecting: %s (type: %s)\n", vpnConfig.Name, vpnConfig.Type)
		}

		if !o.dryRun {
			if err := connectVPN(vpnConfig.Name, vpnConfig.Type, vpnConfig.ConfigFile); err != nil {
				return fmt.Errorf("failed to connect VPN %s: %w", vpnConfig.Name, err)
			}
		}
	}

	return nil
}

func (o *switchOptions) applyDNSConfig(dns *dnsActions) error {
	if o.verbose {
		fmt.Printf("🌐 Configuring DNS servers...\n")
	}

	if len(dns.Servers) > 0 {
		if o.verbose {
			fmt.Printf("   Setting DNS servers: %s\n", strings.Join(dns.Servers, ", "))
		}

		if !o.dryRun {
			if err := setDNSServers(dns.Servers, dns.Interface); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *switchOptions) applyProxyConfig(proxy *proxyActions) error {
	if o.verbose {
		fmt.Printf("🌐 Configuring proxy settings...\n")
	}

	if proxy.Clear {
		if o.verbose {
			fmt.Printf("   Clearing proxy configuration\n")
		}

		if !o.dryRun {
			if err := clearProxy(); err != nil {
				return err
			}
		}
	} else {
		if o.verbose {
			fmt.Printf("   Setting proxy configuration\n")
		}

		if !o.dryRun {
			if err := setProxy(proxy.HTTP, proxy.HTTPS, proxy.SOCKS); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *switchOptions) applyHostsConfig(hosts *hostsActions) error {
	if o.verbose {
		fmt.Printf("📝 Configuring hosts file...\n")
	}

	// Remove entries first
	for _, host := range hosts.Remove {
		if o.verbose {
			fmt.Printf("   Removing host: %s\n", host)
		}

		if !o.dryRun {
			if err := removeHostEntry(host); err != nil {
				fmt.Printf("   ⚠️  Warning: Failed to remove %s: %v\n", host, err)
			}
		}
	}

	// Add entries
	for _, entry := range hosts.Add {
		if o.verbose {
			fmt.Printf("   Adding host: %s -> %s\n", entry.Host, entry.IP)
		}

		if !o.dryRun {
			if err := addHostEntry(entry.IP, entry.Host); err != nil {
				fmt.Printf("   ⚠️  Warning: Failed to add %s: %v\n", entry.Host, err)
			}
		}
	}

	return nil
}

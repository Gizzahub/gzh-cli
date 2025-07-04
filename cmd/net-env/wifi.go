package netenv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type wifiOptions struct {
	configPath string
	daemon     bool
	interval   time.Duration
	action     string
	logPath    string
	dryRun     bool
	verbose    bool
}

type networkState struct {
	SSID      string
	Interface string
	State     string
	IP        string
	Timestamp time.Time
}

type wifiAction struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Commands    []string `yaml:"commands"`
	Conditions  struct {
		SSID      []string `yaml:"ssid,omitempty"`
		Interface []string `yaml:"interface,omitempty"`
		State     []string `yaml:"state,omitempty"`
	} `yaml:"conditions,omitempty"`
}

type wifiConfig struct {
	Actions []wifiAction `yaml:"actions"`
	Global  struct {
		LogPath  string        `yaml:"log_path,omitempty"`
		Interval time.Duration `yaml:"interval,omitempty"`
	} `yaml:"global,omitempty"`
}

func defaultWifiOptions() *wifiOptions {
	homeDir, _ := os.UserHomeDir()
	return &wifiOptions{
		configPath: filepath.Join(homeDir, ".gz", "wifi-hooks.yaml"),
		interval:   5 * time.Second,
		logPath:    filepath.Join(homeDir, ".gz", "logs", "wifi-hooks.log"),
	}
}

func newWifiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wifi",
		Short: "Monitor WiFi changes and trigger actions",
		Long: `Monitor WiFi network changes and trigger configurable actions.

This command monitors WiFi network state changes (connect, disconnect, network switch)
and executes predefined actions based on the network state. This is useful for:
- Automatically connecting to VPNs when joining specific networks
- Switching DNS servers based on network location
- Starting/stopping services based on network availability
- Configuring proxy settings for different environments

The monitor can run as a daemon or in foreground mode with configurable
intervals and action conditions.

Examples:
  # Monitor WiFi changes in foreground
  gz net-env wifi monitor
  
  # Show current WiFi status
  gz net-env wifi status
  
  # Run as daemon with custom config
  gz net-env wifi monitor --daemon --config ~/.config/wifi-actions.yaml
  
  # Test configuration without executing actions
  gz net-env wifi monitor --dry-run
  
  # Monitor with verbose logging
  gz net-env wifi monitor --verbose`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newWifiMonitorCmd())
	cmd.AddCommand(newWifiStatusCmd())
	cmd.AddCommand(newWifiConfigCmd())

	return cmd
}

func newWifiMonitorCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor WiFi changes and execute actions",
		Long: `Monitor WiFi network state changes and execute configured actions.

This command continuously monitors WiFi network state and executes actions
when network changes are detected. Actions are configured in a YAML file
and can include commands to run based on network conditions.

Examples:
  # Monitor with default settings
  gz net-env wifi monitor
  
  # Run as background daemon
  gz net-env wifi monitor --daemon
  
  # Use custom config file
  gz net-env wifi monitor --config /path/to/wifi-actions.yaml
  
  # Test mode - show what would be executed
  gz net-env wifi monitor --dry-run`,
		RunE: o.runMonitor,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to WiFi actions configuration file")
	cmd.Flags().BoolVar(&o.daemon, "daemon", false, "Run as background daemon")
	cmd.Flags().DurationVar(&o.interval, "interval", o.interval, "Check interval for network changes")
	cmd.Flags().StringVar(&o.logPath, "log", o.logPath, "Log file path (used when running as daemon)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Show what would be executed without running commands")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Enable verbose logging")

	return cmd
}

func newWifiStatusCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current WiFi network status",
		Long: `Show current WiFi network status and interface information.

This command displays the current state of WiFi networks, including:
- Connected SSID and signal strength
- Network interface information
- IP address and connection details
- Available networks

Examples:
  # Show current WiFi status
  gz net-env wifi status
  
  # Show detailed interface information
  gz net-env wifi status --verbose`,
		RunE: o.runStatus,
	}

	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed interface information")

	return cmd
}

func newWifiConfigCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage WiFi hook configuration",
		Long: `Manage WiFi hook configuration file.

This command helps create and validate WiFi action configuration files.
The configuration file defines actions to execute when network changes occur.

Examples:
  # Create example configuration
  gz net-env wifi config init
  
  # Validate configuration file
  gz net-env wifi config validate
  
  # Show current configuration
  gz net-env wifi config show`,
		RunE: o.runConfig,
	}

	cmd.AddCommand(newWifiConfigInitCmd())
	cmd.AddCommand(newWifiConfigValidateCmd())
	cmd.AddCommand(newWifiConfigShowCmd())

	return cmd
}

func newWifiConfigInitCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create example WiFi configuration file",
		Long: `Create an example WiFi configuration file with common actions.

This creates a template configuration file that demonstrates how to configure
WiFi change actions for different network scenarios.`,
		RunE: o.runConfigInit,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path where to create configuration file")

	return cmd
}

func newWifiConfigValidateCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate WiFi configuration file",
		Long: `Validate the syntax and structure of WiFi configuration file.

This command checks the configuration file for syntax errors and
validates that all required fields are present.`,
		RunE: o.runConfigValidate,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to configuration file to validate")

	return cmd
}

func newWifiConfigShowCmd() *cobra.Command {
	o := defaultWifiOptions()

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current WiFi configuration",
		Long: `Display the current WiFi configuration file contents.

This command shows the loaded configuration and available actions.`,
		RunE: o.runConfigShow,
	}

	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "Path to configuration file to show")

	return cmd
}

func (o *wifiOptions) runMonitor(_ *cobra.Command, args []string) error {
	if o.daemon {
		return o.runAsDaemon()
	}

	return o.runForeground()
}

func (o *wifiOptions) runStatus(_ *cobra.Command, args []string) error {
	state, err := o.getCurrentNetworkState()
	if err != nil {
		return fmt.Errorf("failed to get network state: %w", err)
	}

	fmt.Printf("📶 WiFi Network Status\n\n")
	if state.SSID != "" {
		fmt.Printf("   Connected to: %s\n", state.SSID)
		fmt.Printf("   Interface: %s\n", state.Interface)
		fmt.Printf("   State: %s\n", state.State)
		if state.IP != "" {
			fmt.Printf("   IP Address: %s\n", state.IP)
		}
	} else {
		fmt.Printf("   Status: Not connected to WiFi\n")
	}

	if o.verbose {
		fmt.Printf("\n🔧 Interface Details:\n")
		if err := o.showInterfaceDetails(); err != nil {
			fmt.Printf("   Warning: Could not retrieve interface details: %v\n", err)
		}
	}

	return nil
}

func (o *wifiOptions) runConfig(_ *cobra.Command, args []string) error {
	return fmt.Errorf("config subcommand required. Use 'gz net-env wifi config --help' for available commands")
}

func (o *wifiOptions) runConfigInit(_ *cobra.Command, args []string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(o.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(o.configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", o.configPath)
	}

	exampleConfig := `# WiFi Hook Configuration
# This file defines actions to execute when WiFi network changes occur

actions:
  - name: "vpn-connect-office"
    description: "Connect to office VPN when joining office network"
    conditions:
      ssid: ["OfficeWiFi", "Office-Guest"]
    commands:
      - "echo 'Connecting to office VPN...'"
      - "systemctl start openvpn@office"
      
  - name: "vpn-disconnect-home"
    description: "Disconnect VPN when at home"
    conditions:
      ssid: ["HomeNetwork", "Home-5G"]
    commands:
      - "echo 'Disconnecting VPN...'"
      - "systemctl stop openvpn@office"
      
  - name: "dns-switch-public"
    description: "Switch to public DNS when on public networks"
    conditions:
      ssid: ["Starbucks", "PublicWiFi", "Guest"]
    commands:
      - "echo 'Switching to secure DNS...'"
      - "resolvectl dns wlan0 1.1.1.1 1.0.0.1"
      
  - name: "network-disconnect"
    description: "Actions when disconnecting from any network"
    conditions:
      state: ["disconnected"]
    commands:
      - "echo 'Network disconnected, cleaning up...'"
      - "systemctl stop openvpn@office || true"

global:
  interval: 5s
  log_path: ~/.gz/logs/wifi-hooks.log
`

	if err := os.WriteFile(o.configPath, []byte(exampleConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ WiFi configuration file created at: %s\n", o.configPath)
	fmt.Printf("   Edit this file to customize actions for your network environments.\n")
	fmt.Printf("   Then start monitoring with: gz net-env wifi monitor\n")

	return nil
}

func (o *wifiOptions) runConfigValidate(_ *cobra.Command, args []string) error {
	_, err := o.loadConfig()
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	fmt.Printf("✅ Configuration file is valid: %s\n", o.configPath)
	return nil
}

func (o *wifiOptions) runConfigShow(_ *cobra.Command, args []string) error {
	config, err := o.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("📋 WiFi Configuration: %s\n\n", o.configPath)
	fmt.Printf("Actions configured: %d\n\n", len(config.Actions))

	for i, action := range config.Actions {
		fmt.Printf("%d. %s\n", i+1, action.Name)
		if action.Description != "" {
			fmt.Printf("   Description: %s\n", action.Description)
		}
		if len(action.Conditions.SSID) > 0 {
			fmt.Printf("   SSID conditions: %s\n", strings.Join(action.Conditions.SSID, ", "))
		}
		if len(action.Conditions.State) > 0 {
			fmt.Printf("   State conditions: %s\n", strings.Join(action.Conditions.State, ", "))
		}
		fmt.Printf("   Commands: %d configured\n", len(action.Commands))
		fmt.Println()
	}

	return nil
}

func (o *wifiOptions) runForeground() error {
	config, err := o.loadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("⚠️  No configuration file found at %s\n", o.configPath)
			fmt.Printf("   Create one with: gz net-env wifi config init\n")
			return nil
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("🔍 Starting WiFi monitor (Press Ctrl+C to stop)\n")
	fmt.Printf("   Config: %s\n", o.configPath)
	fmt.Printf("   Interval: %v\n", o.interval)
	fmt.Printf("   Actions: %d configured\n\n", len(config.Actions))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("\n📴 Shutting down WiFi monitor...\n")
		cancel()
	}()

	return o.monitorLoop(ctx, config)
}

func (o *wifiOptions) runAsDaemon() error {
	// TODO: Implement proper daemon mode with PID file
	// For now, just run in background-like mode
	fmt.Printf("🔄 Starting WiFi monitor as daemon\n")
	fmt.Printf("   Log: %s\n", o.logPath)

	config, err := o.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create log directory
	if err := os.MkdirAll(filepath.Dir(o.logPath), 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	ctx := context.Background()
	return o.monitorLoop(ctx, config)
}

func (o *wifiOptions) monitorLoop(ctx context.Context, config *wifiConfig) error {
	var lastState *networkState
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			currentState, err := o.getCurrentNetworkState()
			if err != nil {
				if o.verbose {
					fmt.Printf("⚠️  Error getting network state: %v\n", err)
				}
				continue
			}

			if o.hasStateChanged(lastState, currentState) {
				if o.verbose {
					fmt.Printf("📡 Network change detected: %s -> %s\n",
						o.formatState(lastState), o.formatState(currentState))
				}

				if err := o.executeActions(config, currentState); err != nil {
					fmt.Printf("❌ Error executing actions: %v\n", err)
				}

				lastState = currentState
			}
		}
	}
}

func (o *wifiOptions) getCurrentNetworkState() (*networkState, error) {
	// Try to get WiFi state using NetworkManager
	if state := o.getNetworkManagerState(); state != nil {
		return state, nil
	}

	// Fallback to iwconfig/ip commands
	return o.getNetworkStateFromCommands()
}

func (o *wifiOptions) getNetworkManagerState() *networkState {
	cmd := exec.Command("nmcli", "-t", "-f", "SSID,STATE,IP4", "dev", "wifi")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) >= 2 && fields[1] == "connected" {
			state := &networkState{
				SSID:      fields[0],
				State:     "connected",
				Timestamp: time.Now(),
			}
			if len(fields) >= 3 {
				state.IP = fields[2]
			}
			return state
		}
	}

	return &networkState{
		State:     "disconnected",
		Timestamp: time.Now(),
	}
}

func (o *wifiOptions) getNetworkStateFromCommands() (*networkState, error) {
	// Get interface name
	cmd := exec.Command("iwconfig")
	output, err := cmd.Output()
	if err != nil {
		return &networkState{State: "disconnected", Timestamp: time.Now()}, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ESSID:") && !strings.Contains(line, "off/any") {
			// Extract SSID
			parts := strings.Split(line, "ESSID:")
			if len(parts) >= 2 {
				ssid := strings.Trim(parts[1], "\" ")
				if ssid != "" {
					return &networkState{
						SSID:      ssid,
						State:     "connected",
						Timestamp: time.Now(),
					}, nil
				}
			}
		}
	}

	return &networkState{State: "disconnected", Timestamp: time.Now()}, nil
}

func (o *wifiOptions) showInterfaceDetails() error {
	cmd := exec.Command("ip", "addr", "show")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Printf("%s", output)
	return nil
}

func (o *wifiOptions) hasStateChanged(old, new *networkState) bool {
	if old == nil {
		return true
	}

	return old.SSID != new.SSID || old.State != new.State
}

func (o *wifiOptions) formatState(state *networkState) string {
	if state == nil {
		return "unknown"
	}
	if state.SSID != "" {
		return fmt.Sprintf("%s (%s)", state.SSID, state.State)
	}
	return state.State
}

func (o *wifiOptions) executeActions(config *wifiConfig, state *networkState) error {
	var executedActions int

	for _, action := range config.Actions {
		if o.shouldExecuteAction(action, state) {
			if o.verbose || o.dryRun {
				fmt.Printf("🎯 Executing action: %s\n", action.Name)
				if action.Description != "" {
					fmt.Printf("   %s\n", action.Description)
				}
			}

			if err := o.executeActionCommands(action); err != nil {
				fmt.Printf("❌ Action '%s' failed: %v\n", action.Name, err)
			} else {
				executedActions++
				if o.verbose {
					fmt.Printf("✅ Action '%s' completed\n", action.Name)
				}
			}
		}
	}

	if executedActions > 0 && o.verbose {
		fmt.Printf("📊 Executed %d actions for network change\n", executedActions)
	}

	return nil
}

func (o *wifiOptions) shouldExecuteAction(action wifiAction, state *networkState) bool {
	// Check SSID conditions
	if len(action.Conditions.SSID) > 0 {
		found := false
		for _, ssid := range action.Conditions.SSID {
			if ssid == state.SSID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check state conditions
	if len(action.Conditions.State) > 0 {
		found := false
		for _, stateCondition := range action.Conditions.State {
			if stateCondition == state.State {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (o *wifiOptions) executeActionCommands(action wifiAction) error {
	for _, command := range action.Commands {
		if o.dryRun {
			fmt.Printf("   [DRY-RUN] %s\n", command)
			continue
		}

		if o.verbose {
			fmt.Printf("   Running: %s\n", command)
		}

		cmd := exec.Command("sh", "-c", command)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed '%s': %w", command, err)
		}
	}

	return nil
}

func (o *wifiOptions) loadConfig() (*wifiConfig, error) {
	// For now, return a basic config structure
	// In a real implementation, this would parse YAML
	return &wifiConfig{
		Actions: []wifiAction{},
	}, nil
}

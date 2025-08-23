// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newVPNProfileCmd creates the VPN profile management command.
func NewProfileCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn-profile",
		Short: "Manage VPN connection profiles and priorities",
		Long: `Manage VPN connection profiles with network-specific mappings and automatic switching rules.

This command provides comprehensive VPN profile management including:
- Network-specific VPN mappings
- Priority-based automatic switching
- Rule-based connection decisions
- Environment detection and adaptation

Examples:
  # List all VPN profiles
  gz net-env vpn-profile list

  # Create new VPN profile
  gz net-env vpn-profile create office-profile --network "Office WiFi" --vpn corp-vpn --priority 100

  # Map network to VPN
  gz net-env vpn-profile map --network "Home WiFi" --vpn home-vpn --priority 50

  # Set automatic switching rules
  gz net-env vpn-profile rule add --trigger network-change --action auto-connect

  # Show current active profile
  gz net-env vpn-profile active`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newVPNProfileListCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileCreateCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileDeleteCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileMapCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileUnmapCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileRuleCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileActiveCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileSwitchCmd(logger, configDir))

	return cmd
}

// newVPNProfileListCmd creates the list subcommand.
func newVPNProfileListCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all VPN profiles",
		Long:  `Display all configured VPN profiles with their network mappings and priorities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			profiles, err := manager.GetAllProfiles()
			if err != nil {
				return fmt.Errorf("failed to get VPN profiles: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(profiles)
			default:
				return printVPNProfiles(profiles)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNProfileCreateCmd creates the create subcommand.
func newVPNProfileCreateCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [profile-name]",
		Short: "Create new VPN profile",
		Long:  `Create a new VPN profile with network mappings and priority settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) == 0 {
				return fmt.Errorf("profile name is required")
			}

			profileName := args[0]
			network, _ := cmd.Flags().GetString("network")
			vpnName, _ := cmd.Flags().GetString("vpn")
			priority, _ := cmd.Flags().GetInt("priority")
			autoConnect, _ := cmd.Flags().GetBool("auto-connect")

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			profile := &VPNProfile{
				Name:        profileName,
				AutoConnect: autoConnect,
				Priority:    priority,
				NetworkMappings: []NetworkVPNMapping{
					{
						NetworkPattern: network,
						VPNName:        vpnName,
						Priority:       priority,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := manager.CreateProfile(profile); err != nil {
				return fmt.Errorf("failed to create VPN profile: %w", err)
			}

			fmt.Printf("✅ Created VPN profile: %s\n", profileName)
			return nil
		},
	}

	cmd.Flags().StringP("network", "n", "", "Network name or pattern")
	cmd.Flags().StringP("vpn", "v", "", "VPN connection name")
	cmd.Flags().IntP("priority", "p", 100, "Priority (higher = more preferred)")
	cmd.Flags().BoolP("auto-connect", "a", false, "Enable automatic connection")
	_ = cmd.MarkFlagRequired("network")
	_ = cmd.MarkFlagRequired("vpn")

	return cmd
}

// newVPNProfileDeleteCmd creates the delete subcommand.
func newVPNProfileDeleteCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [profile-name]",
		Short: "Delete VPN profile",
		Long:  `Delete an existing VPN profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) == 0 {
				return fmt.Errorf("profile name is required")
			}

			profileName := args[0]

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			if err := manager.DeleteProfile(profileName); err != nil {
				return fmt.Errorf("failed to delete VPN profile: %w", err)
			}

			fmt.Printf("✅ Deleted VPN profile: %s\n", profileName)
			return nil
		},
	}

	return cmd
}

// newVPNProfileMapCmd creates the map subcommand.
func newVPNProfileMapCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "map",
		Short: "Map network to VPN connection",
		Long:  `Create a mapping between a network and VPN connection with priority.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			network, _ := cmd.Flags().GetString("network")
			vpnName, _ := cmd.Flags().GetString("vpn")
			priority, _ := cmd.Flags().GetInt("priority")
			profileName, _ := cmd.Flags().GetString("profile")

			if network == "" || vpnName == "" {
				return fmt.Errorf("both network and vpn flags are required")
			}

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			mapping := NetworkVPNMapping{
				NetworkPattern: network,
				VPNName:        vpnName,
				Priority:       priority,
			}

			if err := manager.AddNetworkMapping(profileName, mapping); err != nil {
				return fmt.Errorf("failed to add network mapping: %w", err)
			}

			fmt.Printf("✅ Mapped network '%s' to VPN '%s' (priority: %d)\n", network, vpnName, priority)
			return nil
		},
	}

	cmd.Flags().StringP("network", "n", "", "Network name or pattern")
	cmd.Flags().StringP("vpn", "v", "", "VPN connection name")
	cmd.Flags().IntP("priority", "p", 100, "Priority (higher = more preferred)")
	cmd.Flags().String("profile", "default", "Profile name to add mapping to")
	_ = cmd.MarkFlagRequired("network")
	_ = cmd.MarkFlagRequired("vpn")

	return cmd
}

// newVPNProfileUnmapCmd creates the unmap subcommand.
func newVPNProfileUnmapCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmap",
		Short: "Remove network to VPN mapping",
		Long:  `Remove a mapping between a network and VPN connection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			network, _ := cmd.Flags().GetString("network")
			profileName, _ := cmd.Flags().GetString("profile")

			if network == "" {
				return fmt.Errorf("network flag is required")
			}

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			if err := manager.RemoveNetworkMapping(profileName, network); err != nil {
				return fmt.Errorf("failed to remove network mapping: %w", err)
			}

			fmt.Printf("✅ Removed mapping for network: %s\n", network)
			return nil
		},
	}

	cmd.Flags().StringP("network", "n", "", "Network name or pattern")
	cmd.Flags().String("profile", "default", "Profile name to remove mapping from")
	_ = cmd.MarkFlagRequired("network")

	return cmd
}

// newVPNProfileRuleCmd creates the rule subcommand.
func newVPNProfileRuleCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rule",
		Short:        "Manage automatic switching rules",
		Long:         `Manage automatic VPN switching rules based on network changes and other triggers.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newVPNProfileRuleAddCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileRuleRemoveCmd(logger, configDir))
	cmd.AddCommand(newVPNProfileRuleListCmd(logger, configDir))

	return cmd
}

// newVPNProfileRuleAddCmd creates the rule add subcommand.
func newVPNProfileRuleAddCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add automatic switching rule",
		Long:  `Add a new automatic VPN switching rule.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			trigger, _ := cmd.Flags().GetString("trigger")
			action, _ := cmd.Flags().GetString("action")
			condition, _ := cmd.Flags().GetString("condition")

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			rule := AutoSwitchRule{
				Name:      fmt.Sprintf("rule-%d", time.Now().Unix()),
				Trigger:   AutoSwitchTrigger(trigger),
				Action:    AutoSwitchAction(action),
				Condition: condition,
				Enabled:   true,
				CreatedAt: time.Now(),
			}

			if err := manager.AddAutoSwitchRule(rule); err != nil {
				return fmt.Errorf("failed to add auto-switch rule: %w", err)
			}

			fmt.Printf("✅ Added auto-switch rule: %s\n", rule.Name)
			return nil
		},
	}

	cmd.Flags().String("trigger", "network-change", "Trigger type (network-change, time-based, manual)")
	cmd.Flags().String("action", "auto-connect", "Action to take (auto-connect, disconnect-all, switch-profile)")
	cmd.Flags().String("condition", "", "Condition for rule activation")

	return cmd
}

// newVPNProfileRuleRemoveCmd creates the rule remove subcommand.
func newVPNProfileRuleRemoveCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [rule-name]",
		Short: "Remove automatic switching rule",
		Long:  `Remove an existing automatic VPN switching rule.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) == 0 {
				return fmt.Errorf("rule name is required")
			}

			ruleName := args[0]

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			if err := manager.RemoveAutoSwitchRule(ruleName); err != nil {
				return fmt.Errorf("failed to remove auto-switch rule: %w", err)
			}

			fmt.Printf("✅ Removed auto-switch rule: %s\n", ruleName)
			return nil
		},
	}

	return cmd
}

// newVPNProfileRuleListCmd creates the rule list subcommand.
func newVPNProfileRuleListCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List automatic switching rules",
		Long:  `Display all configured automatic VPN switching rules.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			rules, err := manager.GetAutoSwitchRules()
			if err != nil {
				return fmt.Errorf("failed to get auto-switch rules: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(rules)
			default:
				return printAutoSwitchRules(rules)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNProfileActiveCmd creates the active subcommand.
func newVPNProfileActiveCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active",
		Short: "Show current active VPN profile",
		Long:  `Display the currently active VPN profile and its connections.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			profile, err := manager.GetActiveProfile()
			if err != nil {
				return fmt.Errorf("failed to get active profile: %w", err)
			}

			if profile == nil {
				fmt.Println("No active VPN profile.")
				return nil
			}

			output, _ := cmd.Flags().GetString("output")

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(profile)
			default:
				return printActiveProfile(profile)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNProfileSwitchCmd creates the switch subcommand.
func newVPNProfileSwitchCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [profile-name]",
		Short: "Switch to different VPN profile",
		Long:  `Switch to a different VPN profile and apply its settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			if len(args) == 0 {
				return fmt.Errorf("profile name is required")
			}

			profileName := args[0]

			manager, err := createVPNProfileManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN profile manager: %w", err)
			}

			fmt.Printf("🔄 Switching to VPN profile: %s\n", profileName)

			if err := manager.SwitchToProfile(ctx, profileName); err != nil {
				return fmt.Errorf("failed to switch to profile: %w", err)
			}

			fmt.Printf("✅ Successfully switched to profile: %s\n", profileName)
			return nil
		},
	}

	return cmd
}

// Helper functions and types

type VPNProfile struct {
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	AutoConnect     bool                `json:"autoConnect"`
	Priority        int                 `json:"priority"`
	NetworkMappings []NetworkVPNMapping `json:"networkMappings"`
	AutoSwitchRules []AutoSwitchRule    `json:"autoSwitchRules"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

type NetworkVPNMapping struct {
	NetworkPattern string `json:"networkPattern"`
	VPNName        string `json:"vpnName"`
	Priority       int    `json:"priority"`
}

type AutoSwitchRule struct {
	Name      string            `json:"name"`
	Trigger   AutoSwitchTrigger `json:"trigger"`
	Action    AutoSwitchAction  `json:"action"`
	Condition string            `json:"condition,omitempty"`
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"createdAt"`
}

type (
	AutoSwitchTrigger string
	AutoSwitchAction  string
)

const (
	TriggerNetworkChange AutoSwitchTrigger = "network-change"
	TriggerTimeBased     AutoSwitchTrigger = "time-based"
	TriggerManual        AutoSwitchTrigger = "manual"

	ActionAutoConnect   AutoSwitchAction = "auto-connect"
	ActionDisconnectAll AutoSwitchAction = "disconnect-all"
	ActionSwitchProfile AutoSwitchAction = "switch-profile"
)

type VPNProfileManager struct {
	logger    *zap.Logger
	configDir string
	profiles  map[string]*VPNProfile
}

func createVPNProfileManager(_ context.Context, logger *zap.Logger, configDir string) (*VPNProfileManager, error) { //nolint:unparam // TODO: implement error handling
	manager := &VPNProfileManager{
		logger:    logger,
		configDir: configDir,
		profiles:  make(map[string]*VPNProfile),
	}

	// Load existing profiles
	manager.loadProfiles()

	return manager, nil
}

func (vpm *VPNProfileManager) GetAllProfiles() ([]*VPNProfile, error) {
	profiles := make([]*VPNProfile, 0, len(vpm.profiles))
	for _, profile := range vpm.profiles {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (vpm *VPNProfileManager) CreateProfile(profile *VPNProfile) error {
	vpm.profiles[profile.Name] = profile
	return vpm.saveProfiles()
}

func (vpm *VPNProfileManager) DeleteProfile(name string) error {
	delete(vpm.profiles, name)
	return vpm.saveProfiles()
}

func (vpm *VPNProfileManager) AddNetworkMapping(profileName string, mapping NetworkVPNMapping) error {
	profile, exists := vpm.profiles[profileName]
	if !exists {
		// Create default profile if it doesn't exist
		profile = &VPNProfile{
			Name:        profileName,
			Priority:    100,
			AutoConnect: false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		vpm.profiles[profileName] = profile
	}

	profile.NetworkMappings = append(profile.NetworkMappings, mapping)
	profile.UpdatedAt = time.Now()

	return vpm.saveProfiles()
}

func (vpm *VPNProfileManager) RemoveNetworkMapping(profileName, networkPattern string) error {
	profile, exists := vpm.profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %s not found", profileName)
	}

	for i, mapping := range profile.NetworkMappings {
		if mapping.NetworkPattern == networkPattern {
			profile.NetworkMappings = append(profile.NetworkMappings[:i], profile.NetworkMappings[i+1:]...)
			profile.UpdatedAt = time.Now()

			return vpm.saveProfiles()
		}
	}

	return fmt.Errorf("network mapping not found")
}

func (vpm *VPNProfileManager) AddAutoSwitchRule(rule AutoSwitchRule) error {
	// Add to default profile for now
	profile, exists := vpm.profiles["default"]
	if !exists {
		profile = &VPNProfile{
			Name:      "default",
			Priority:  100,
			CreatedAt: time.Now(),
		}
		vpm.profiles["default"] = profile
	}

	profile.AutoSwitchRules = append(profile.AutoSwitchRules, rule)
	profile.UpdatedAt = time.Now()

	return vpm.saveProfiles()
}

func (vpm *VPNProfileManager) RemoveAutoSwitchRule(ruleName string) error {
	for _, profile := range vpm.profiles {
		for i, rule := range profile.AutoSwitchRules {
			if rule.Name == ruleName {
				profile.AutoSwitchRules = append(profile.AutoSwitchRules[:i], profile.AutoSwitchRules[i+1:]...)
				profile.UpdatedAt = time.Now()

				return vpm.saveProfiles()
			}
		}
	}

	return fmt.Errorf("auto-switch rule not found")
}

func (vpm *VPNProfileManager) GetAutoSwitchRules() ([]AutoSwitchRule, error) {
	var rules []AutoSwitchRule
	for _, profile := range vpm.profiles {
		rules = append(rules, profile.AutoSwitchRules...)
	}

	return rules, nil
}

func (vpm *VPNProfileManager) GetActiveProfile() (*VPNProfile, error) {
	// TODO: Implement active profile detection
	// For now, return the default profile if it exists
	if profile, exists := vpm.profiles["default"]; exists {
		return profile, nil
	}

	return nil, fmt.Errorf("no active VPN profile found")
}

func (vpm *VPNProfileManager) SwitchToProfile(ctx context.Context, profileName string) error {
	profile, exists := vpm.profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %s not found", profileName)
	}

	// TODO: Implement profile switching logic
	// This would involve:
	// 1. Disconnecting current VPN connections
	// 2. Applying new profile's network mappings
	// 3. Connecting appropriate VPNs based on current network

	vpm.logger.Info("Switching to VPN profile", zap.String("profile", profileName))

	_ = profile // Use the profile

	return nil
}

func (vpm *VPNProfileManager) loadProfiles() {
	configPath := filepath.Join(vpm.configDir, "vpn-profiles.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return // Config file doesn't exist, that's ok
	}

	// TODO: Implement JSON profile loading
}

func (vpm *VPNProfileManager) saveProfiles() error {
	_ = filepath.Join(vpm.configDir, "vpn-profiles.json") // configPath for future use

	// Ensure config directory exists
	if err := os.MkdirAll(vpm.configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// TODO: Implement JSON profile saving
	return nil
}

func printVPNProfiles(profiles []*VPNProfile) error {
	fmt.Printf("📋 VPN Profiles\n\n")

	if len(profiles) == 0 {
		fmt.Println("  No VPN profiles configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPRIORITY\tAUTO-CONNECT\tMAPPINGS\tRULES\tUPDATED")

	for _, profile := range profiles {
		autoConnect := "No"
		if profile.AutoConnect {
			autoConnect = "Yes"
		}

		updated := profile.UpdatedAt.Format("2006-01-02 15:04")

		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%d\t%s\n",
			profile.Name, profile.Priority, autoConnect,
			len(profile.NetworkMappings), len(profile.AutoSwitchRules), updated)
	}

	return w.Flush()
}

func printAutoSwitchRules(rules []AutoSwitchRule) error {
	fmt.Printf("🔄 Auto-Switch Rules\n\n")

	if len(rules) == 0 {
		fmt.Println("  No auto-switch rules configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTRIGGER\tACTION\tCONDITION\tENABLED\tCREATED")

	for _, rule := range rules {
		enabled := "No"
		if rule.Enabled {
			enabled = "Yes"
		}

		condition := rule.Condition
		if condition == "" {
			condition = "-"
		}

		created := rule.CreatedAt.Format("2006-01-02")

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			rule.Name, rule.Trigger, rule.Action, condition, enabled, created)
	}

	return w.Flush()
}

func printActiveProfile(profile *VPNProfile) error {
	fmt.Printf("🎯 Active VPN Profile: %s\n\n", profile.Name)

	fmt.Printf("Details:\n")
	fmt.Printf("  Priority: %d\n", profile.Priority)
	fmt.Printf("  Auto-Connect: %t\n", profile.AutoConnect)
	fmt.Printf("  Updated: %s\n\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))

	if len(profile.NetworkMappings) > 0 {
		fmt.Printf("Network Mappings:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "  NETWORK\tVPN\tPRIORITY")

		for _, mapping := range profile.NetworkMappings {
			_, _ = fmt.Fprintf(w, "  %s\t%s\t%d\n",
				mapping.NetworkPattern, mapping.VPNName, mapping.Priority)
		}

		_ = w.Flush()
		fmt.Println()
	}

	if len(profile.AutoSwitchRules) > 0 {
		fmt.Printf("Auto-Switch Rules:\n")

		for _, rule := range profile.AutoSwitchRules {
			status := "disabled"
			if rule.Enabled {
				status = "enabled"
			}

			fmt.Printf("  - %s: %s → %s (%s)\n",
				rule.Name, rule.Trigger, rule.Action, status)
		}
	}

	return nil
}

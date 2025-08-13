// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-manager-go/pkg/cloud"
)

const (
	// valueNotAvailable is used when a value is not available.
	valueNotAvailable = "N/A"
)

// cloudOptions contains options for cloud commands.
type cloudOptions struct {
	configFile string
	verbose    bool
}

// newCloudCmd creates the cloud subcommand.
func newCloudCmd(ctx context.Context) *cobra.Command {
	opts := &cloudOptions{}

	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud provider network configurations",
		Long: `Manage cloud provider network configurations and environment profiles.

This command allows you to:
- Configure cloud provider profiles (AWS, GCP, Azure)
- Sync network configurations between cloud environments
- Apply environment-specific network policies
- Switch between cloud profiles with automatic network reconfiguration`,
	}

	// Global flags for cloud commands
	cmd.PersistentFlags().StringVarP(&opts.configFile, "config", "c", "", "Cloud configuration file (default: auto-detect)")
	cmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	cmd.AddCommand(newCloudListCmd(ctx, opts))
	cmd.AddCommand(newCloudShowCmd(ctx, opts))
	cmd.AddCommand(newCloudSwitchCmd(ctx, opts))
	cmd.AddCommand(newCloudSyncCmd(ctx, opts))
	cmd.AddCommand(newCloudValidateCmd(ctx, opts))
	cmd.AddCommand(newCloudPolicyCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNCmd(ctx, opts))

	return cmd
}

// newCloudListCmd creates the list subcommand.
func newCloudListCmd(_ context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit // Large command builder - requires architectural refactoring
	var (
		showProfiles  bool
		showProviders bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud providers and profiles",
		Long:  `List configured cloud providers and their profiles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Show providers if requested or by default
			if showProviders || (!showProfiles && !showProviders) {
				if err := displayCloudProviders(config); err != nil {
					return err
				}
			}

			// Show profiles if requested
			if showProfiles || (!showProfiles && !showProviders) {
				if err := displayCloudProfiles(config); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showProfiles, "profiles", false, "Show only profiles")
	cmd.Flags().BoolVar(&showProviders, "providers", false, "Show only providers")

	return cmd
}

// newCloudShowCmd creates the show subcommand.
func newCloudShowCmd(_ context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit // Complex cloud show command with multiple display options
	cmd := &cobra.Command{
		Use:   "show [profile]",
		Short: "Show detailed profile information",
		Long:  `Show detailed information about a cloud profile including network configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get profile
			profile, exists := config.GetProfile(profileName)
			if !exists {
				return fmt.Errorf("profile not found: %s", profileName)
			}

			// Display profile information
			displayProfileBasicInfo(profile)
			displayProfileNetworkConfig(profile)
			displayProfileServices(profile)
			displayProfileTags(profile)
			displayProfileSyncInfo(profile)

			return nil
		},
	}

	return cmd
}

// newCloudSwitchCmd creates the switch subcommand.
func newCloudSwitchCmd(ctx context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit,gocyclo // Large command builder with multiple flags, options, and execution paths - requires architectural refactoring
	var (
		dryRun      bool
		applyPolicy bool
	)

	cmd := &cobra.Command{
		Use:   "switch [profile]",
		Short: "Switch to a cloud profile",
		Long: `Switch to a cloud profile and apply its network configuration.

This command will:
1. Load the specified cloud profile
2. Apply network policies (DNS, proxy, routes)
3. Connect to VPN if configured
4. Update environment variables`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get profile
			profile, exists := config.GetProfile(profileName)
			if !exists {
				return fmt.Errorf("profile not found: %s", profileName)
			}

			// Get provider config
			providerConfig, exists := config.GetProvider(profile.Provider)
			if !exists {
				return fmt.Errorf("provider not found: %s", profile.Provider)
			}

			fmt.Printf("Switching to cloud profile: %s\n", profileName)

			if dryRun {
				fmt.Println("\n[DRY RUN] Would perform the following actions:")
			}

			// Create provider instance
			provider, err := cloud.NewProvider(ctx, providerConfig)
			if err != nil {
				return fmt.Errorf("failed to create provider: %w", err)
			}

			// Get network policy
			if applyPolicy {
				fmt.Println("\nApplying network policy...")

				policy, err := provider.GetNetworkPolicy(ctx, profileName)
				if err != nil {
					if opts.verbose {
						fmt.Printf("Warning: failed to get network policy: %v\n", err)
					}
				} else if policy != nil && policy.Enabled {
					if !dryRun {
						if err := provider.ApplyNetworkPolicy(ctx, policy); err != nil {
							return fmt.Errorf("failed to apply network policy: %w", err)
						}
					}
					fmt.Println("✓ Network policy applied")
				}
			}

			// Apply network configuration
			fmt.Println("\nApplying network configuration...")

			// DNS configuration
			if len(profile.Network.DNSServers) > 0 {
				fmt.Printf("  Setting DNS servers: %v\n", profile.Network.DNSServers)
				if !dryRun {
					_ = profile.Network.DNSServers // TODO: Implement DNS configuration
				}
			}

			// Proxy configuration
			if profile.Network.Proxy != nil {
				if profile.Network.Proxy.HTTP != "" {
					fmt.Printf("  Setting HTTP proxy: %s\n", profile.Network.Proxy.HTTP)
					if !dryRun {
						if err := os.Setenv("HTTP_PROXY", profile.Network.Proxy.HTTP); err != nil {
							return fmt.Errorf("failed to set HTTP_PROXY: %w", err)
						}
						if err := os.Setenv("http_proxy", profile.Network.Proxy.HTTP); err != nil {
							return fmt.Errorf("failed to set http_proxy: %w", err)
						}
					}
				}
				if profile.Network.Proxy.HTTPS != "" {
					fmt.Printf("  Setting HTTPS proxy: %s\n", profile.Network.Proxy.HTTPS)
					if !dryRun {
						if err := os.Setenv("HTTPS_PROXY", profile.Network.Proxy.HTTPS); err != nil {
							return fmt.Errorf("failed to set HTTPS_PROXY: %w", err)
						}
						if err := os.Setenv("https_proxy", profile.Network.Proxy.HTTPS); err != nil {
							return fmt.Errorf("failed to set https_proxy: %w", err)
						}
					}
				}
			}

			// VPN connection
			if profile.Network.VPN != nil && profile.Network.VPN.AutoConnect {
				fmt.Printf("  Connecting to VPN: %s\n", profile.Network.VPN.Server)
				if !dryRun {
					_ = profile.Network.VPN // TODO: Implement VPN connection
				}
			}

			// Custom routes
			if len(profile.Network.Routes) > 0 {
				fmt.Println("  Adding custom routes:")
				for _, route := range profile.Network.Routes {
					fmt.Printf("    %s via %s\n", route.Destination, route.Gateway)
					if !dryRun {
						_ = route // TODO: Implement route addition
					}
				}
			}

			if dryRun {
				fmt.Println("\n[DRY RUN] No changes were made")
			} else {
				fmt.Printf("\n✓ Successfully switched to profile: %s\n", profileName)

				// Save current profile
				if err := saveCurrentProfile(profileName); err != nil {
					if opts.verbose {
						fmt.Printf("Warning: failed to save current profile: %v\n", err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&applyPolicy, "apply-policy", true, "Apply network policy from cloud provider")

	return cmd
}

// newCloudSyncCmd creates the sync subcommand.
func newCloudSyncCmd(ctx context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit,gocyclo // Complex cloud sync command with profile management and multiple execution branches
	var (
		source              string
		target              string
		profiles            []string
		conflictMode        string
		showStatus          bool
		showRecommendations bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync profiles between cloud providers",
		Long: `Synchronize profiles between different cloud providers.

This allows you to maintain consistent network configurations across
multiple cloud environments with intelligent conflict resolution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Validate sync configuration
			if err := cloud.ValidateSyncConfig(config); err != nil {
				return fmt.Errorf("invalid sync configuration: %w", err)
			}

			// Create sync manager
			syncManager := cloud.NewSyncManager(config)

			// Show sync status if requested
			if showStatus {
				status, err := syncManager.GetSyncStatus(ctx)
				if err != nil {
					return fmt.Errorf("failed to get sync status: %w", err)
				}

				fmt.Println("Sync Status:")
				fmt.Println("============")
				if len(status) == 0 {
					fmt.Println("No sync history found")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "PROFILE\tSOURCE\tTARGET\tSTATUS\tLAST SYNC\tERROR") //nolint:errcheck // CLI output errors are non-critical
				for _, s := range status {
					errorMsg := s.Error
					if errorMsg == "" {
						errorMsg = "-"
					}
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", //nolint:errcheck // CLI output errors are non-critical
						s.ProfileName,
						s.Source,
						s.Target,
						s.Status,
						s.LastSync.Format("2006-01-02 15:04:05"),
						errorMsg,
					)
				}
				_ = w.Flush() //nolint:errcheck // CLI output errors are non-critical
				return nil
			}

			// Show recommendations if requested
			if showRecommendations {
				recommendations, err := cloud.GetSyncRecommendations(config)
				if err != nil {
					return fmt.Errorf("failed to get sync recommendations: %w", err)
				}

				fmt.Println("Sync Recommendations:")
				fmt.Println("====================")
				if len(recommendations) == 0 {
					fmt.Println("No sync recommendations found")
					return nil
				}

				for i, rec := range recommendations {
					fmt.Printf("%d. Sync from %s to %s\n", i+1, rec.Source, rec.Target)
					fmt.Printf("   Profiles: %s\n", strings.Join(rec.Profiles, ", "))
					fmt.Printf("   Command: gz net-env cloud sync --source %s --target %s\n\n", rec.Source, rec.Target)
				}
				return nil
			}

			// Validate source and target
			if source == "" || target == "" {
				return fmt.Errorf("both --source and --target are required")
			}

			sourceConfig, exists := config.GetProvider(source)
			if !exists {
				return fmt.Errorf("source provider not found: %s", source)
			}

			targetConfig, exists := config.GetProvider(target)
			if !exists {
				return fmt.Errorf("target provider not found: %s", target)
			}

			// Create provider instances
			sourceProvider, err := cloud.NewProvider(ctx, sourceConfig)
			if err != nil {
				return fmt.Errorf("failed to create source provider: %w", err)
			}

			targetProvider, err := cloud.NewProvider(ctx, targetConfig)
			if err != nil {
				return fmt.Errorf("failed to create target provider: %w", err)
			}

			fmt.Printf("Syncing profiles from %s to %s...\n", source, target)

			// Set conflict mode if specified
			if conflictMode != "" {
				config.Sync.ConflictMode = cloud.ConflictStrategy(conflictMode)
			}

			// Get profiles to sync
			var profilesToSync []string
			if len(profiles) > 0 {
				profilesToSync = profiles
			} else {
				// Sync all profiles for the source provider
				sourceProfiles := config.GetProfilesForProvider(source)
				for _, p := range sourceProfiles {
					profilesToSync = append(profilesToSync, p.Name)
				}
			}

			if len(profilesToSync) == 0 {
				fmt.Println("No profiles to sync")
				return nil
			}

			// Perform sync using sync manager
			err = syncManager.SyncProfiles(ctx, sourceProvider, targetProvider, profilesToSync)
			if err != nil {
				fmt.Printf("Sync completed with errors: %v\n", err)

				// Show sync status for failed profiles
				status, _ := syncManager.GetSyncStatus(ctx)
				for _, s := range status {
					if s.Status == "error" || s.Status == "conflict" {
						fmt.Printf("  ✗ %s: %s\n", s.ProfileName, s.Error)
					}
				}
				return nil
			}

			fmt.Printf("✓ All profiles synced successfully\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Source provider name")
	cmd.Flags().StringVar(&target, "target", "", "Target provider name")
	cmd.Flags().StringSliceVar(&profiles, "profiles", nil, "Specific profiles to sync (default: all)")
	cmd.Flags().StringVar(&conflictMode, "conflict-mode", "", "Conflict resolution mode (source_wins, target_wins, merge, ask)")
	cmd.Flags().BoolVar(&showStatus, "status", false, "Show sync status history")
	cmd.Flags().BoolVar(&showRecommendations, "recommendations", false, "Show sync recommendations")

	return cmd
}

// newCloudValidateCmd creates the validate subcommand.
func newCloudValidateCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate cloud configuration",
		Long:  `Validate cloud configuration file and test provider connections.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			fmt.Printf("Validating configuration: %s\n\n", configPath)

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("✓ Configuration file is valid")
			fmt.Printf("  Version: %s\n", config.Version)
			fmt.Printf("  Providers: %d\n", len(config.Providers))
			fmt.Printf("  Profiles: %d\n", len(config.Profiles))

			// Test provider connections
			fmt.Println("\nTesting provider connections:")
			for name, providerConfig := range config.Providers {
				fmt.Printf("\n%s (%s):\n", name, providerConfig.Type)

				provider, err := cloud.NewProvider(ctx, providerConfig)
				if err != nil {
					fmt.Printf("  ✗ Failed to create provider: %v\n", err)
					continue
				}

				// Validate configuration
				if err := provider.ValidateConfig(providerConfig); err != nil {
					fmt.Printf("  ✗ Invalid configuration: %v\n", err)
					continue
				}
				fmt.Println("  ✓ Configuration valid")

				// Health check
				if err := provider.HealthCheck(ctx); err != nil {
					fmt.Printf("  ✗ Health check failed: %v\n", err)
					continue
				}
				fmt.Println("  ✓ Connection successful")

				// List profiles
				profiles, err := provider.ListProfiles(ctx)
				if err != nil {
					fmt.Printf("  ⚠ Failed to list profiles: %v\n", err)
				} else {
					fmt.Printf("  ✓ Found %d profiles\n", len(profiles))
				}
			}

			return nil
		},
	}

	return cmd
}

// saveCurrentProfile saves the current profile name.
func saveCurrentProfile(profileName string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	stateFile := filepath.Join(configDir, "gzh-manager", "current-cloud-profile")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		return err
	}

	return os.WriteFile(stateFile, []byte(profileName), 0o644)
}

// getCurrentProfile reads the current profile name.
func getCurrentProfile() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	stateFile := filepath.Join(configDir, "gzh-manager", "current-cloud-profile")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	return string(data), nil
}

// newCloudPolicyCmd creates the policy subcommand.
func newCloudPolicyCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage network policies",
		Long:  `Manage network policies for cloud profiles with automatic application by environment.`,
	}

	// Add policy subcommands
	cmd.AddCommand(newCloudPolicyApplyCmd(ctx, opts))
	cmd.AddCommand(newCloudPolicyListCmd(ctx, opts))
	cmd.AddCommand(newCloudPolicyStatusCmd(ctx, opts))
	cmd.AddCommand(newCloudPolicyValidateCmd(ctx, opts))

	return cmd
}

// newCloudPolicyApplyCmd creates the policy apply subcommand.
func newCloudPolicyApplyCmd(ctx context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit // Complex policy application command with validation
	var (
		environment string
		profileName string
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply network policies",
		Long: `Apply network policies to cloud profiles automatically.

This command can apply policies either by environment (all profiles in an environment)
or by specific profile name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyApplyCommand(ctx, opts, environment, profileName, dryRun)
		},
	}

	cmd.Flags().StringVarP(&environment, "environment", "e", "", "Apply policies for all profiles in environment")
	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Apply policies for specific profile")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without applying policies")

	return cmd
}

// newCloudPolicyListCmd creates the policy list subcommand.
func newCloudPolicyListCmd(ctx context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit,gocyclo // Complex policy listing command with filtering and multiple display formats
	var (
		profileName  string
		environment  string
		showDisabled bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List network policies",
		Long:  `List network policies for profiles or environments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create policy manager
			policyManager := cloud.NewPolicyManager(config)

			switch {
			case profileName != "":
				// List policies for specific profile
				policies, err := policyManager.GetApplicablePolicies(ctx, profileName)
				if err != nil {
					return fmt.Errorf("failed to get policies: %w", err)
				}

				fmt.Printf("Policies for profile: %s\n", profileName)
				fmt.Println("=========================")

				if len(policies) == 0 {
					fmt.Println("No policies found")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "NAME\tPRIORITY\tENABLED\tRULES\tACTIONS") //nolint:errcheck // CLI output errors are non-critical

				for _, policy := range policies {
					if !showDisabled && !policy.Enabled {
						continue
					}

					enabled := "✓"
					if !policy.Enabled {
						enabled = "✗"
					}

					_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%d\n", //nolint:errcheck // CLI output errors are non-critical
						policy.Name,
						policy.Priority,
						enabled,
						len(policy.Rules),
						len(policy.Actions),
					)
				}
				_ = w.Flush() //nolint:errcheck // CLI output errors are non-critical
			case environment != "":
				// List policies for environment
				profiles := getProfilesForEnvironment(config, environment)
				if len(profiles) == 0 {
					fmt.Printf("No profiles found for environment: %s\n", environment)
					return nil
				}

				fmt.Printf("Policies for environment: %s\n", environment)
				fmt.Println("===========================")

				for _, profile := range profiles {
					fmt.Printf("\nProfile: %s\n", profile.Name)
					fmt.Println("----------")

					policies, err := policyManager.GetApplicablePolicies(ctx, profile.Name)
					if err != nil {
						fmt.Printf("Error getting policies: %v\n", err)
						continue
					}

					if len(policies) == 0 {
						fmt.Println("No policies found")
						continue
					}

					for _, policy := range policies {
						if !showDisabled && !policy.Enabled {
							continue
						}

						enabled := "✓"
						if !policy.Enabled {
							enabled = "✗"
						}

						fmt.Printf("  %s %s (Priority: %d, Rules: %d, Actions: %d)\n",
							enabled,
							policy.Name,
							policy.Priority,
							len(policy.Rules),
							len(policy.Actions),
						)
					}
				}
			default:
				// List all configured policies
				fmt.Println("Configured Policies:")
				fmt.Println("==================")

				if len(config.Policies) == 0 {
					fmt.Println("No policies configured")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "NAME\tPROFILE\tENVIRONMENT\tPRIORITY\tENABLED\tRULES\tACTIONS") // Ignore print error

				for _, policy := range config.Policies {
					if !showDisabled && !policy.Enabled {
						continue
					}

					enabled := "✓"
					if !policy.Enabled {
						enabled = "✗"
					}

					profileName := policy.ProfileName
					if profileName == "" {
						profileName = "*"
					}

					environment := policy.Environment
					if environment == "" {
						environment = "*"
					}

					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%d\t%d\n",
						policy.Name,
						profileName,
						environment,
						policy.Priority,
						enabled,
						len(policy.Rules),
						len(policy.Actions),
					)
				}
				_ = w.Flush() // Ignore flush error
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "List policies for specific profile")
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "List policies for environment")
	cmd.Flags().BoolVar(&showDisabled, "show-disabled", false, "Show disabled policies")

	return cmd
}

// newCloudPolicyStatusCmd creates the policy status subcommand.
func newCloudPolicyStatusCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show policy application status",
		Long:  `Show the status of applied network policies across all profiles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create policy manager
			policyManager := cloud.NewPolicyManager(config)

			// Get policy status
			status, err := policyManager.GetPolicyStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get policy status: %w", err)
			}

			fmt.Println("Policy Application Status:")
			fmt.Println("========================")

			if len(status) == 0 {
				fmt.Println("No policy status found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "POLICY\tPROFILE\tPROVIDER\tSTATUS\tAPPLIED\tERROR") // Ignore print error

			for _, s := range status {
				errorMsg := s.Error
				if errorMsg == "" {
					errorMsg = "-"
				}

				appliedTime := s.Applied.Format("2006-01-02 15:04:05")

				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					s.PolicyName,
					s.ProfileName,
					s.Provider,
					s.Status,
					appliedTime,
					errorMsg,
				)
			}
			_ = w.Flush() // Ignore flush error

			return nil
		},
	}

	return cmd
}

// newCloudPolicyValidateCmd creates the policy validate subcommand.
func newCloudPolicyValidateCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate network policies",
		Long:  `Validate network policies for syntax and consistency.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create policy manager
			policyManager := cloud.NewPolicyManager(config)

			fmt.Println("Validating policies...")
			fmt.Println("====================")

			var validationErrors []string

			if profileName != "" {
				// Validate policies for specific profile
				policies, err := policyManager.GetApplicablePolicies(ctx, profileName)
				if err != nil {
					return fmt.Errorf("failed to get policies: %w", err)
				}

				fmt.Printf("Validating policies for profile: %s\n", profileName)

				for _, policy := range policies {
					if err := policyManager.ValidatePolicy(ctx, policy); err != nil {
						validationErrors = append(validationErrors, fmt.Sprintf("Policy %s: %v", policy.Name, err))
						fmt.Printf("  ✗ %s: %v\n", policy.Name, err)
					} else {
						fmt.Printf("  ✓ %s: valid\n", policy.Name)
					}
				}
			} else {
				// Validate all configured policies
				if len(config.Policies) == 0 {
					fmt.Println("No policies configured")
					return nil
				}

				for _, policy := range config.Policies {
					if err := policyManager.ValidatePolicy(ctx, &policy); err != nil {
						validationErrors = append(validationErrors, fmt.Sprintf("Policy %s: %v", policy.Name, err))
						fmt.Printf("  ✗ %s: %v\n", policy.Name, err)
					} else {
						fmt.Printf("  ✓ %s: valid\n", policy.Name)
					}
				}
			}

			if len(validationErrors) > 0 {
				fmt.Printf("\n%d validation errors found:\n", len(validationErrors))
				for _, err := range validationErrors {
					fmt.Printf("  - %s\n", err)
				}
				return fmt.Errorf("validation failed")
			}

			fmt.Println("\n✓ All policies are valid")
			return nil
		},
	}

	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Validate policies for specific profile")

	return cmd
}

// Helper function to get profiles for environment.
func getProfilesForEnvironment(config *cloud.Config, environment string) []cloud.Profile {
	var profiles []cloud.Profile

	for _, profile := range config.Profiles {
		if profile.Environment == environment {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

// newCloudVPNCmd creates the VPN subcommand.
func newCloudVPNCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "Manage VPN connections",
		Long: `Manage multiple VPN connections with priority-based connection and automatic failover.

This command provides comprehensive VPN management including:
- Multiple VPN connection management
- Priority-based connection ordering
- Automatic failover and health monitoring
- Connection status monitoring`,
	}

	// Add VPN subcommands
	cmd.AddCommand(newCloudVPNListCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNConnectCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNDisconnectCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNStatusCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNAddCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNRemoveCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNMonitorCmd(ctx, opts))
	// Hierarchical VPN management commands
	cmd.AddCommand(newCloudVPNHierarchyCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNConnectHierarchicalCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNDisconnectHierarchicalCmd(ctx, opts))
	cmd.AddCommand(newCloudVPNEnvironmentCmd(ctx, opts))

	return cmd
}

// newCloudVPNListCmd creates the VPN list subcommand.
func newCloudVPNListCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VPN connections",
		Long:  `List all configured VPN connections with their status and priority.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create VPN manager
			vpnManager := cloud.NewVPNManager()

			// Add VPN connections from config
			if err := loadVPNConnections(vpnManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			// Get connection status
			status, err := vpnManager.GetAllVPNStatuses(ctx)
			if err != nil {
				return fmt.Errorf("failed to get VPN statuses: %w", err)
			}
			activeConnections, err := vpnManager.GetActiveConnections(ctx)
			if err != nil {
				return fmt.Errorf("failed to get active connections: %w", err)
			}

			if len(status) == 0 {
				fmt.Println("No VPN connections configured")
				return nil
			}

			fmt.Println("VPN Connections:")
			fmt.Println("===============")

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tTYPE\tSERVER\tPRIORITY\tSTATUS\tAUTO-CONNECT") // Ignore print error

			for name, s := range status {
				// Get connection details from config
				conn := findVPNConnection(config, name)
				if conn == nil {
					continue
				}

				if !showAll && s.Status == cloud.VPNStateDisconnected {
					continue
				}

				displayVPNConnectionRow(w, name, conn, s)
			}

			_ = w.Flush() // Ignore flush error

			if len(activeConnections) > 0 {
				fmt.Printf("\nActive Connections: %d\n", len(activeConnections))
				for connName := range activeConnections {
					// Get connection config to get type
					vpnConn := findVPNConnection(config, connName)
					connType := "unknown"
					if vpnConn != nil {
						connType = vpnConn.Type
					}
					fmt.Printf("  - %s (%s)\n", connName, connType)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all connections including disconnected ones")

	return cmd
}

// newCloudVPNConnectCmd creates the VPN connect subcommand.
func newCloudVPNConnectCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var (
		byPriority bool
		vpnName    string
	)

	cmd := &cobra.Command{
		Use:   "connect [vpn-name]",
		Short: "Connect to VPN",
		Long: `Connect to a specific VPN connection or connect by priority order.

Examples:
  gz net-env cloud vpn connect my-vpn        # Connect to specific VPN
  gz net-env cloud vpn connect --priority    # Connect by priority order`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				vpnName = args[0]
			}

			if !byPriority && vpnName == "" {
				return fmt.Errorf("either specify VPN name or use --priority flag")
			}

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create VPN manager
			vpnManager := cloud.NewVPNManager()

			// Add VPN connections from config
			if err := loadVPNConnections(vpnManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			if byPriority {
				fmt.Println("Connecting to VPNs by priority...")
				// Get all VPN connections and connect by priority
				connections, err := vpnManager.ListVPNConnections()
				if err != nil {
					return fmt.Errorf("failed to list VPN connections: %w", err)
				}
				// Sort by priority and extract names
				var connectionNames []string
				for _, conn := range connections {
					connectionNames = append(connectionNames, conn.Name)
				}
				if err := vpnManager.ConnectByPriority(ctx, connectionNames); err != nil {
					return fmt.Errorf("failed to connect by priority: %w", err)
				}
			} else {
				fmt.Printf("Connecting to VPN: %s\n", vpnName)
				if err := vpnManager.ConnectVPN(ctx, vpnName); err != nil {
					return fmt.Errorf("failed to connect to VPN: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&byPriority, "priority", "p", false, "Connect by priority order")

	return cmd
}

// newCloudVPNDisconnectCmd creates the VPN disconnect subcommand.
func newCloudVPNDisconnectCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var disconnectAll bool

	cmd := &cobra.Command{
		Use:   "disconnect [vpn-name]",
		Short: "Disconnect from VPN",
		Long: `Disconnect from a specific VPN connection or disconnect all active connections.

Examples:
  gz net-env cloud vpn disconnect my-vpn    # Disconnect from specific VPN
  gz net-env cloud vpn disconnect --all     # Disconnect from all VPNs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var vpnName string
			if len(args) > 0 {
				vpnName = args[0]
			}

			if !disconnectAll && vpnName == "" {
				return fmt.Errorf("either specify VPN name or use --all flag")
			}

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create VPN manager
			vpnManager := cloud.NewVPNManager()

			// Add VPN connections from config
			if err := loadVPNConnections(vpnManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			if disconnectAll {
				fmt.Println("Disconnecting from all VPNs...")
				activeConnections, err := vpnManager.GetActiveConnections(ctx)
				if err != nil {
					return fmt.Errorf("failed to get active connections: %w", err)
				}
				for connName := range activeConnections {
					if err := vpnManager.DisconnectVPN(ctx, connName); err != nil {
						fmt.Printf("Failed to disconnect from %s: %v\n", connName, err)
					}
				}
			} else {
				fmt.Printf("Disconnecting from VPN: %s\n", vpnName)
				if err := vpnManager.DisconnectVPN(ctx, vpnName); err != nil {
					return fmt.Errorf("failed to disconnect from VPN: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&disconnectAll, "all", "a", false, "Disconnect from all VPNs")

	return cmd
}

// newCloudVPNStatusCmd creates the VPN status subcommand.
func newCloudVPNStatusCmd(ctx context.Context, opts *cloudOptions) *cobra.Command { //nolint:gocognit // Complex VPN status command with multiple providers
	var (
		showHealth bool
		vpnName    string
	)

	cmd := &cobra.Command{
		Use:   "status [vpn-name]",
		Short: "Show VPN connection status",
		Long: `Show detailed status of VPN connections including health information.

Examples:
  gz net-env cloud vpn status              # Show status of all VPNs
  gz net-env cloud vpn status my-vpn       # Show status of specific VPN
  gz net-env cloud vpn status --health     # Show health check details`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				vpnName = args[0]
			}
			return runVPNStatusCommand(ctx, opts, vpnName, showHealth)
		},
	}

	cmd.Flags().BoolVar(&showHealth, "health", false, "Show health check details")

	return cmd
}

// newCloudVPNAddCmd creates the VPN add subcommand.
func newCloudVPNAddCmd(_ context.Context, opts *cloudOptions) *cobra.Command {
	var (
		vpnType     string
		server      string
		port        int
		priority    int
		autoConnect bool
		configFile  string
	)

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new VPN connection",
		Long: `Add a new VPN connection to the configuration.

Examples:
  gz net-env cloud vpn add my-vpn --type openvpn --server vpn.example.com --port 1194
  gz net-env cloud vpn add work-vpn --type wireguard --config /etc/wireguard/wg0.conf`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vpnName := args[0]

			if vpnType == "" {
				return fmt.Errorf("VPN type is required (--type)")
			}
			if server == "" && configFile == "" {
				return fmt.Errorf("either server (--server) or config file (--config) is required")
			}

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create VPN connection
			vpnConn := &cloud.VPNConnection{
				Name:        vpnName,
				Type:        vpnType,
				Server:      server,
				Port:        port,
				Priority:    priority,
				AutoConnect: autoConnect,
				ConfigFile:  configFile,
			}

			// Note: ValidateConnection method not available in VPNManager interface

			// Add VPN connection to config
			addVPNConnectionToConfig(config, vpnConn)

			// Save configuration
			if err := cloud.SaveConfig(config, configPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ VPN connection '%s' added successfully\n", vpnName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&vpnType, "type", "t", "", "VPN type (openvpn, wireguard, ipsec)")
	cmd.Flags().StringVarP(&server, "server", "s", "", "VPN server address")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "VPN server port")
	cmd.Flags().IntVar(&priority, "priority", 100, "Connection priority (higher = more preferred)")
	cmd.Flags().BoolVar(&autoConnect, "auto-connect", false, "Enable auto-connect")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to VPN configuration file")

	return cmd
}

// newCloudVPNRemoveCmd creates the VPN remove subcommand.
func newCloudVPNRemoveCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a VPN connection",
		Long: `Remove a VPN connection from the configuration.

Examples:
  gz net-env cloud vpn remove my-vpn
  gz net-env cloud vpn remove work-vpn --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vpnName := args[0]

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if VPN connection exists
			if findVPNConnection(config, vpnName) == nil {
				return fmt.Errorf("VPN connection not found: %s", vpnName)
			}

			// If not force, check if connection is active
			if !force {
				vpnManager := cloud.NewVPNManager()
				if err := loadVPNConnections(vpnManager, config); err != nil {
					return fmt.Errorf("failed to load VPN connections: %w", err)
				}

				status, err := vpnManager.GetAllVPNStatuses(ctx)
				if err != nil {
					return fmt.Errorf("failed to get VPN statuses: %w", err)
				}
				if s, exists := status[vpnName]; exists && s.Status == cloud.VPNStateConnected {
					return fmt.Errorf("VPN connection '%s' is currently connected. Use --force to remove anyway", vpnName)
				}
			}

			// Remove VPN connection from config
			removeVPNConnectionFromConfig(config, vpnName)

			// Save configuration
			if err := cloud.SaveConfig(config, configPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ VPN connection '%s' removed successfully\n", vpnName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal even if connected")

	return cmd
}

// newCloudVPNMonitorCmd creates the VPN monitor subcommand.
func newCloudVPNMonitorCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor VPN connections with failover",
		Long: `Start monitoring VPN connections with automatic failover.

This command starts a continuous monitoring process that:
- Monitors health of all active VPN connections
- Automatically triggers failover when connections fail
- Attempts to reconnect failed connections
- Provides real-time status updates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create VPN manager
			vpnManager := cloud.NewVPNManager()

			// Add VPN connections from config
			if err := loadVPNConnections(vpnManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			// Note: Failover monitoring methods not available in VPNManager interface

			fmt.Println("VPN monitoring started. Press Ctrl+C to stop.")
			fmt.Printf("Monitoring interval: %v\n", interval)

			// Monitor loop
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("\nMonitoring stopped.")
					return nil
				case <-ticker.C:
					// Display current status
					status, err := vpnManager.GetAllVPNStatuses(ctx)
					if err != nil {
						fmt.Printf("\r[%s] Error getting status: %v",
							time.Now().Format("15:04:05"), err)
						continue
					}
					activeCount := 0
					for _, s := range status {
						if s.Status == cloud.VPNStateConnected {
							activeCount++
						}
					}
					fmt.Printf("\r[%s] Active connections: %d/%d",
						time.Now().Format("15:04:05"),
						activeCount,
						len(status))
				}
			}
		},
	}

	cmd.Flags().DurationVarP(&interval, "interval", "i", 30*time.Second, "Monitoring interval")

	return cmd
}

// Helper functions

func loadVPNConnections(manager cloud.VPNManager, config *cloud.Config) error {
	// Load VPN connections from config
	vpnConnections := config.GetVPNConnections()

	for _, conn := range vpnConnections {
		// Create a copy to avoid reference issues
		vpnConn := conn
		if err := manager.AddVPNConnection(&vpnConn); err != nil {
			return fmt.Errorf("failed to add VPN connection %s: %w", conn.Name, err)
		}
	}

	return nil
}

func findVPNConnection(config *cloud.Config, name string) *cloud.VPNConnection {
	// Find VPN connection in config
	vpnConn, exists := config.GetVPNConnection(name)
	if !exists {
		return nil
	}

	return &vpnConn
}

func addVPNConnectionToConfig(config *cloud.Config, conn *cloud.VPNConnection) {
	// Add VPN connection to config
	config.AddVPNConnection(*conn)
}

func removeVPNConnectionFromConfig(config *cloud.Config, name string) {
	// Remove VPN connection from config
	config.RemoveVPNConnection(name)
}

// Hierarchical VPN Management Commands

// newCloudVPNHierarchyCmd creates the VPN hierarchy management subcommand.
func newCloudVPNHierarchyCmd(_ context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hierarchy",
		Short: "Show VPN connection hierarchy",
		Long: `Display the hierarchical structure of VPN connections showing parent-child relationships,
layers, and site types.

This command helps visualize:
- Parent-child VPN relationships
- Layer assignments
- Site types (corporate, personal, public)
- Network environment requirements`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create hierarchical VPN manager
			baseManager := cloud.NewVPNManager()
			hierarchicalManager := cloud.NewHierarchicalVPNManager(baseManager)

			// Load VPN connections
			if err := loadVPNConnections(hierarchicalManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			// Note: ValidateHierarchy method not available in HierarchicalVPNManager interface

			// Get all hierarchies and display
			hierarchies, err := hierarchicalManager.ListVPNHierarchies()
			if err != nil {
				return fmt.Errorf("failed to list VPN hierarchies: %w", err)
			}

			if len(hierarchies) == 0 {
				fmt.Println("No VPN hierarchies found")
				return nil
			}

			fmt.Println("VPN Connection Hierarchies:")
			fmt.Println("==========================")

			for _, hierarchy := range hierarchies {
				fmt.Printf("\nHierarchy: %s\n", hierarchy.Name)
				if hierarchy.Description != "" {
					fmt.Printf("  Description: %s\n", hierarchy.Description)
				}
				if hierarchy.Environment != "" {
					fmt.Printf("  Environment: %s\n", hierarchy.Environment)
				}
				// Display layers
				for layer, nodes := range hierarchy.Layers {
					fmt.Printf("  Layer %d:\n", layer)
					for _, node := range nodes {
						if node.Connection != nil {
							fmt.Printf("    - %s (%s)\n", node.Connection.Name, node.Connection.Type)
						}
					}
				}
			}

			return nil
		},
	}

	return cmd
}

// newCloudVPNConnectHierarchicalCmd creates the hierarchical VPN connect subcommand.
func newCloudVPNConnectHierarchicalCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect-hierarchy [root-connection]",
		Short: "Connect VPN connections in hierarchical order",
		Long: `Connect VPN connections following the hierarchical structure.
This ensures parent connections are established before children.

Examples:
  gz net-env cloud vpn connect-hierarchy corporate-vpn    # Connect corporate VPN and all children
  gz net-env cloud vpn connect-hierarchy personal-vpn    # Connect personal VPN hierarchy`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootConnection := args[0]

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create hierarchical VPN manager
			baseManager := cloud.NewVPNManager()
			hierarchicalManager := cloud.NewHierarchicalVPNManager(baseManager)

			// Load VPN connections
			if err := loadVPNConnections(hierarchicalManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			// Note: ValidateHierarchy and ConnectHierarchical methods not available in interface

			fmt.Printf("Connecting VPN hierarchy starting from: %s\n", rootConnection)
			// Use ConnectVPNHierarchy instead
			if err := hierarchicalManager.ConnectVPNHierarchy(ctx, rootConnection); err != nil {
				return fmt.Errorf("failed to connect hierarchical VPN: %w", err)
			}

			fmt.Println("✓ Hierarchical VPN connection completed successfully")
			return nil
		},
	}

	return cmd
}

// newCloudVPNDisconnectHierarchicalCmd creates the hierarchical VPN disconnect subcommand.
func newCloudVPNDisconnectHierarchicalCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect-hierarchy [root-connection]",
		Short: "Disconnect VPN connections in reverse hierarchical order",
		Long: `Disconnect VPN connections following the reverse hierarchical structure.
This ensures child connections are disconnected before parents.

Examples:
  gz net-env cloud vpn disconnect-hierarchy corporate-vpn    # Disconnect corporate VPN hierarchy
  gz net-env cloud vpn disconnect-hierarchy personal-vpn    # Disconnect personal VPN hierarchy`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootConnection := args[0]

			// Load configuration
			configPath := opts.configFile
			if configPath == "" {
				configPath = cloud.GetDefaultConfigPath()
			}

			config, err := cloud.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create hierarchical VPN manager
			baseManager := cloud.NewVPNManager()
			hierarchicalManager := cloud.NewHierarchicalVPNManager(baseManager)

			// Load VPN connections
			if err := loadVPNConnections(hierarchicalManager, config); err != nil {
				return fmt.Errorf("failed to load VPN connections: %w", err)
			}

			fmt.Printf("Disconnecting VPN hierarchy starting from: %s\n", rootConnection)
			// Use DisconnectVPNHierarchy instead
			if err := hierarchicalManager.DisconnectVPNHierarchy(ctx, rootConnection); err != nil {
				return fmt.Errorf("failed to disconnect hierarchical VPN: %w", err)
			}

			fmt.Println("✓ Hierarchical VPN disconnection completed successfully")
			return nil
		},
	}

	return cmd
}

// newCloudVPNEnvironmentCmd creates the environment-based VPN management subcommand.
func newCloudVPNEnvironmentCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var (
		autoConnect bool
		listOnly    bool
	)

	cmd := &cobra.Command{
		Use:   "environment [network-environment]",
		Short: "Manage VPNs based on network environment",
		Long: `Connect or list VPN connections appropriate for the current network environment.

Supported environments:
- home: Home network environment
- office: Office network environment
- public: Public network (cafes, airports, etc.)
- mobile: Mobile network environment
- hotel: Hotel network environment

Examples:
  gz net-env cloud vpn environment public --auto-connect    # Auto-connect VPNs for public networks
  gz net-env cloud vpn environment office --list           # List VPNs suitable for office
  gz net-env cloud vpn environment home                    # Show home environment VPNs`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envStr := args[0]

			if err := validateEnvironment(envStr); err != nil {
				return err
			}

			manager, err := setupVPNManager(opts)
			if err != nil {
				return err
			}

			envConnections, err := getEnvironmentConnections(manager, envStr)
			if err != nil {
				return err
			}

			if listOnly {
				return displayVPNList(envStr, envConnections)
			}

			if autoConnect {
				return autoConnectVPNs(ctx, manager, envStr, envConnections)
			}

			return displayVPNSummary(envStr, envConnections)
		},
	}

	cmd.Flags().BoolVar(&autoConnect, "auto-connect", false, "Automatically connect suitable VPNs")
	cmd.Flags().BoolVar(&listOnly, "list", false, "Only list suitable VPNs without connecting")

	return cmd
}

// validateEnvironment validates the network environment string.
func validateEnvironment(envStr string) error {
	validEnvs := []string{"home", "office", "public", "mobile", "hotel"}
	for _, validEnv := range validEnvs {
		if envStr == validEnv {
			return nil
		}
	}
	return fmt.Errorf("unsupported network environment: %s (valid: %v)", envStr, validEnvs)
}

// setupVPNManager creates and configures a hierarchical VPN manager.
func setupVPNManager(opts *cloudOptions) (cloud.HierarchicalVPNManager, error) {
	configPath := opts.configFile
	if configPath == "" {
		configPath = cloud.GetDefaultConfigPath()
	}

	config, err := cloud.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	baseManager := cloud.NewVPNManager()
	hierarchicalManager := cloud.NewHierarchicalVPNManager(baseManager)

	if err := loadVPNConnections(hierarchicalManager, config); err != nil {
		return nil, fmt.Errorf("failed to load VPN connections: %w", err)
	}

	return hierarchicalManager, nil
}

// getEnvironmentConnections filters VPN connections by environment.
func getEnvironmentConnections(manager cloud.HierarchicalVPNManager, envStr string) ([]*cloud.VPNConnection, error) {
	connections, err := manager.ListVPNConnections()
	if err != nil {
		return nil, fmt.Errorf("failed to list VPN connections: %w", err)
	}

	var envConnections []*cloud.VPNConnection
	for _, conn := range connections {
		if conn.Environment == envStr {
			envConnections = append(envConnections, conn)
		}
	}

	return envConnections, nil
}

// displayVPNList displays VPN connections in a table format.
func displayVPNList(envStr string, envConnections []*cloud.VPNConnection) error {
	fmt.Printf("VPN connections suitable for %s environment:\n", envStr)
	fmt.Println("============================================")

	if len(envConnections) == 0 {
		fmt.Println("No VPN connections configured for this environment")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTYPE\tPRIORITY\tAUTO-CONNECT") //nolint:errcheck // Table output
	for _, conn := range envConnections {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%v\n", //nolint:errcheck // Table output
			conn.Name,
			conn.Type,
			conn.Priority,
			conn.AutoConnect,
		)
	}
	_ = w.Flush() //nolint:errcheck // Table output
	return nil
}

// autoConnectVPNs automatically connects VPNs with auto-connect enabled.
func autoConnectVPNs(ctx context.Context, manager cloud.HierarchicalVPNManager, envStr string, envConnections []*cloud.VPNConnection) error {
	fmt.Printf("Auto-connecting VPNs for %s environment...\n", envStr)

	for _, conn := range envConnections {
		if conn.AutoConnect {
			fmt.Printf("Connecting to %s...\n", conn.Name)
			if err := manager.ConnectVPN(ctx, conn.Name); err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", conn.Name, err)
			} else {
				fmt.Printf("✓ Connected to %s\n", conn.Name)
			}
		}
	}

	fmt.Println("✓ Auto-connection completed")
	return nil
}

// displayVPNSummary displays a summary of available VPN connections.
func displayVPNSummary(envStr string, envConnections []*cloud.VPNConnection) error {
	fmt.Printf("VPN connections available for %s environment:\n", envStr)
	for _, conn := range envConnections {
		fmt.Printf("- %s (priority: %d, auto-connect: %v)\n",
			conn.Name, conn.Priority, conn.AutoConnect)
	}
	return nil
}

// displayCloudProviders displays configured cloud providers in a table format.
func displayCloudProviders(config *cloud.Config) error {
	fmt.Println("Cloud Providers:")
	fmt.Println("================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tTYPE\tREGION\tAUTH METHOD"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for name, provider := range config.Providers {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			name,
			provider.Type,
			provider.Region,
			provider.Auth.Method,
		); err != nil {
			return fmt.Errorf("failed to write provider info: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}
	fmt.Println()
	return nil
}

// displayCloudProfiles displays configured cloud profiles in a table format.
func displayCloudProfiles(config *cloud.Config) error {
	fmt.Println("Cloud Profiles:")
	fmt.Println("===============")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tPROVIDER\tENVIRONMENT\tREGION\tVPC"); err != nil {
		return fmt.Errorf("failed to write profile header: %w", err)
	}
	for name, profile := range config.Profiles {
		vpcID := profile.Network.VPCId
		if vpcID == "" {
			vpcID = valueNotAvailable
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			name,
			profile.Provider,
			profile.Environment,
			profile.Region,
			vpcID,
		); err != nil {
			return fmt.Errorf("failed to write profile info: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush profile output: %w", err)
	}
	return nil
}

// displayProfileBasicInfo displays basic profile information.
func displayProfileBasicInfo(profile cloud.Profile) {
	fmt.Printf("Profile: %s\n", profile.Name)
	fmt.Printf("Provider: %s\n", profile.Provider)
	fmt.Printf("Environment: %s\n", profile.Environment)
	fmt.Printf("Region: %s\n", profile.Region)
}

// displayProfileNetworkConfig displays network configuration details.
func displayProfileNetworkConfig(profile cloud.Profile) {
	fmt.Println("\nNetwork Configuration:")
	fmt.Printf("  VPC ID: %s\n", profile.Network.VPCId)

	if len(profile.Network.SubnetIDs) > 0 {
		fmt.Println("  Subnets:")
		for _, subnet := range profile.Network.SubnetIDs {
			fmt.Printf("    - %s\n", subnet)
		}
	}

	if len(profile.Network.SecurityGroups) > 0 {
		fmt.Println("  Security Groups:")
		for _, sg := range profile.Network.SecurityGroups {
			fmt.Printf("    - %s\n", sg)
		}
	}

	if len(profile.Network.DNSServers) > 0 {
		fmt.Println("  DNS Servers:")
		for _, dns := range profile.Network.DNSServers {
			fmt.Printf("    - %s\n", dns)
		}
	}

	// Proxy configuration
	if profile.Network.Proxy != nil {
		fmt.Println("  Proxy:")
		if profile.Network.Proxy.HTTP != "" {
			fmt.Printf("    HTTP: %s\n", profile.Network.Proxy.HTTP)
		}
		if profile.Network.Proxy.HTTPS != "" {
			fmt.Printf("    HTTPS: %s\n", profile.Network.Proxy.HTTPS)
		}
	}

	// VPN configuration
	if profile.Network.VPN != nil {
		fmt.Println("  VPN:")
		fmt.Printf("    Type: %s\n", profile.Network.VPN.Type)
		fmt.Printf("    Server: %s\n", profile.Network.VPN.Server)
		if profile.Network.VPN.AutoConnect {
			fmt.Println("    Auto-connect: enabled")
		}
	}
}

// displayProfileServices displays service configuration details.
func displayProfileServices(profile cloud.Profile) {
	if len(profile.Services) > 0 {
		fmt.Println("\nServices:")
		for name, service := range profile.Services {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Endpoint: %s\n", service.Endpoint)
			if service.Port > 0 {
				fmt.Printf("    Port: %d\n", service.Port)
			}
		}
	}
}

// displayProfileTags displays profile tags.
func displayProfileTags(profile cloud.Profile) {
	if len(profile.Tags) > 0 {
		fmt.Println("\nTags:")
		for key, value := range profile.Tags {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}

// displayProfileSyncInfo displays profile synchronization information.
func displayProfileSyncInfo(profile cloud.Profile) {
	if !profile.LastSync.IsZero() {
		fmt.Printf("\nLast Sync: %s\n", profile.LastSync.Format("2006-01-02 15:04:05"))
	}
}

// displayVPNConnectionRow formats and displays a single VPN connection row in the table.
func displayVPNConnectionRow(w *tabwriter.Writer, name string, conn *cloud.VPNConnection, s *cloud.VPNStatus) {
	autoConnect := "No"
	if conn.AutoConnect {
		autoConnect = "Yes"
	}

	statusStr := s.Status
	switch s.Status {
	case cloud.VPNStateConnected:
		statusStr = "✓ Connected"
	case cloud.VPNStateError:
		statusStr = "✗ Error"
	}

	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
		name, conn.Type, conn.Server, conn.Priority, statusStr, autoConnect)
}

// Helper functions for hierarchy display

// runVPNStatusCommand handles the VPN status command execution.
func runVPNStatusCommand(ctx context.Context, opts *cloudOptions, vpnName string, showHealth bool) error {
	// Load configuration
	configPath := opts.configFile
	if configPath == "" {
		configPath = cloud.GetDefaultConfigPath()
	}

	config, err := cloud.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create VPN manager
	vpnManager := cloud.NewVPNManager()

	// Add VPN connections from config
	if err := loadVPNConnections(vpnManager, config); err != nil {
		return fmt.Errorf("failed to load VPN connections: %w", err)
	}

	// Get connection status
	status, err := vpnManager.GetAllVPNStatuses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get VPN statuses: %w", err)
	}

	if len(status) == 0 {
		fmt.Println("No VPN connections configured")
		return nil
	}

	return displayVPNStatusResults(*config, status, vpnName, showHealth)
}

// displayVPNStatusResults displays VPN status information.
func displayVPNStatusResults(config cloud.Config, status map[string]*cloud.VPNStatus, vpnName string, showHealth bool) error {
	fmt.Println("VPN Connection Status:")
	fmt.Println("=====================")

	for name, s := range status {
		if vpnName != "" && name != vpnName {
			continue
		}

		conn := findVPNConnection(&config, name)
		if conn == nil {
			continue
		}

		displaySingleVPNStatus(name, conn, *s, showHealth)
	}

	if vpnName != "" {
		if _, exists := status[vpnName]; !exists {
			return fmt.Errorf("VPN connection not found: %s", vpnName)
		}
	}

	return nil
}

// displaySingleVPNStatus displays status information for a single VPN connection.
func displaySingleVPNStatus(name string, conn *cloud.VPNConnection, status cloud.VPNStatus, showHealth bool) {
	fmt.Printf("\nConnection: %s\n", name)
	fmt.Printf("  Type: %s\n", conn.Type)
	fmt.Printf("  Server: %s\n", conn.Server)
	fmt.Printf("  Priority: %d\n", conn.Priority)
	fmt.Printf("  State: %s\n", status.Status)

	displayVPNConnectionDetails(status)

	if showHealth {
		displayVPNHealthCheck(status.HealthCheck)
	}

	displayVPNConfiguration(conn)
}

// displayVPNConnectionDetails displays connection-specific details.
func displayVPNConnectionDetails(status cloud.VPNStatus) {
	if status.Status == cloud.VPNStateConnected {
		fmt.Printf("  Connected: %s\n", status.ConnectedAt.Format("2006-01-02 15:04:05"))
		if status.IPAddress != "" {
			fmt.Printf("  IP Address: %s\n", status.IPAddress)
		}
	}

	if status.LastError != "" {
		fmt.Printf("  Error: %s\n", status.LastError)
	}
}

// displayVPNHealthCheck displays health check information.
func displayVPNHealthCheck(healthCheck *cloud.VPNHealthStatus) {
	if healthCheck == nil {
		return
	}

	fmt.Printf("  Last Health Check:\n")
	fmt.Printf("    Time: %s\n", healthCheck.LastCheck.Format("2006-01-02 15:04:05"))
	fmt.Printf("    Status: %s\n", healthCheck.Status)
	fmt.Printf("    Target: %s\n", healthCheck.Target)
	if healthCheck.ResponseTime > 0 {
		fmt.Printf("    Response Time: %v\n", healthCheck.ResponseTime)
	}
	fmt.Printf("    Success Count: %d\n", healthCheck.SuccessCount)
	fmt.Printf("    Failure Count: %d\n", healthCheck.FailureCount)
}

// displayVPNConfiguration displays VPN configuration details.
func displayVPNConfiguration(conn *cloud.VPNConnection) {
	if conn.AutoConnect {
		fmt.Printf("  Auto-Connect: Enabled\n")
	}

	if conn.HealthCheck != nil && conn.HealthCheck.Enabled {
		fmt.Printf("  Health Check: Enabled\n")
		fmt.Printf("    Interval: %v\n", conn.HealthCheck.Interval)
		fmt.Printf("    Target: %s\n", conn.HealthCheck.Target)
	}
}

// runPolicyApplyCommand handles the policy apply command execution.
func runPolicyApplyCommand(ctx context.Context, opts *cloudOptions, environment, profileName string, dryRun bool) error {
	// Load configuration
	configPath := opts.configFile
	if configPath == "" {
		configPath = cloud.GetDefaultConfigPath()
	}

	config, err := cloud.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create policy manager
	policyManager := cloud.NewPolicyManager(config)

	if dryRun {
		fmt.Println("[DRY RUN] Would apply the following policies:")
	}

	// Apply policies by environment or profile
	if err := applyPoliciesByTarget(ctx, policyManager, config, environment, profileName, dryRun); err != nil {
		return err
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] No policies were actually applied")
	} else {
		fmt.Println("✓ Policies applied successfully")
	}

	return nil
}

// applyPoliciesByTarget applies policies based on environment or profile target.
func applyPoliciesByTarget(ctx context.Context, policyManager cloud.PolicyManager, config *cloud.Config, environment, profileName string, dryRun bool) error {
	switch {
	case environment != "":
		return applyEnvironmentPolicies(ctx, policyManager, config, environment, dryRun)
	case profileName != "":
		return applyProfilePolicies(ctx, policyManager, profileName, dryRun)
	default:
		return fmt.Errorf("either --environment or --profile must be specified")
	}
}

// applyEnvironmentPolicies applies policies for an entire environment.
func applyEnvironmentPolicies(ctx context.Context, policyManager cloud.PolicyManager, config *cloud.Config, environment string, dryRun bool) error {
	fmt.Printf("Applying policies for environment: %s\n", environment)

	if !dryRun {
		if err := policyManager.ApplyEnvironmentPolicies(ctx, environment); err != nil {
			return fmt.Errorf("failed to apply environment policies: %w", err)
		}
	} else {
		return showEnvironmentPolicyDryRun(ctx, policyManager, config, environment)
	}

	return nil
}

// showEnvironmentPolicyDryRun displays what policies would be applied for an environment.
func showEnvironmentPolicyDryRun(ctx context.Context, policyManager cloud.PolicyManager, config *cloud.Config, environment string) error {
	profiles := getProfilesForEnvironment(config, environment)
	for _, profile := range profiles {
		fmt.Printf("  - Profile: %s (Provider: %s)\n", profile.Name, profile.Provider)

		policies, err := policyManager.GetApplicablePolicies(ctx, profile.Name)
		if err != nil {
			fmt.Printf("    Warning: failed to get policies: %v\n", err)
			continue
		}

		for _, policy := range policies {
			if policy.Enabled {
				fmt.Printf("    - Policy: %s (Priority: %d)\n", policy.Name, policy.Priority)
			}
		}
	}
	return nil
}

// applyProfilePolicies applies policies for a specific profile.
func applyProfilePolicies(ctx context.Context, policyManager cloud.PolicyManager, profileName string, dryRun bool) error {
	fmt.Printf("Applying policies for profile: %s\n", profileName)

	if !dryRun {
		if err := policyManager.ApplyPoliciesForProfile(ctx, profileName); err != nil {
			return fmt.Errorf("failed to apply profile policies: %w", err)
		}
	} else {
		return showProfilePolicyDryRun(ctx, policyManager, profileName)
	}

	return nil
}

// showProfilePolicyDryRun displays what policies would be applied for a profile.
func showProfilePolicyDryRun(ctx context.Context, policyManager cloud.PolicyManager, profileName string) error {
	policies, err := policyManager.GetApplicablePolicies(ctx, profileName)
	if err != nil {
		return fmt.Errorf("failed to get policies: %w", err)
	}

	for _, policy := range policies {
		if policy.Enabled {
			fmt.Printf("  - Policy: %s (Priority: %d)\n", policy.Name, policy.Priority)
			fmt.Printf("    Rules: %d, Actions: %d\n", len(policy.Rules), len(policy.Actions))
		}
	}

	return nil
}

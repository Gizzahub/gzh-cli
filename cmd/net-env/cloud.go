package netenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	// Import providers to register them
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/aws"
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/azure"
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/gcp"
	"github.com/spf13/cobra"
)

// cloudOptions contains options for cloud commands
type cloudOptions struct {
	configFile string
	provider   string
	profile    string
	verbose    bool
}

// newCloudCmd creates the cloud subcommand
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

	return cmd
}

// newCloudListCmd creates the list subcommand
func newCloudListCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var showProfiles bool
	var showProviders bool

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
				fmt.Println("Cloud Providers:")
				fmt.Println("================")
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tTYPE\tREGION\tAUTH METHOD")
				for name, provider := range config.Providers {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						name,
						provider.Type,
						provider.Region,
						provider.Auth.Method,
					)
				}
				w.Flush()
				fmt.Println()
			}

			// Show profiles if requested
			if showProfiles || (!showProfiles && !showProviders) {
				fmt.Println("Cloud Profiles:")
				fmt.Println("===============")
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tPROVIDER\tENVIRONMENT\tREGION\tVPC")
				for name, profile := range config.Profiles {
					vpcId := profile.Network.VPCId
					if vpcId == "" {
						vpcId = "N/A"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						name,
						profile.Provider,
						profile.Environment,
						profile.Region,
						vpcId,
					)
				}
				w.Flush()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showProfiles, "profiles", false, "Show only profiles")
	cmd.Flags().BoolVar(&showProviders, "providers", false, "Show only providers")

	return cmd
}

// newCloudShowCmd creates the show subcommand
func newCloudShowCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
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
			fmt.Printf("Profile: %s\n", profile.Name)
			fmt.Printf("Provider: %s\n", profile.Provider)
			fmt.Printf("Environment: %s\n", profile.Environment)
			fmt.Printf("Region: %s\n", profile.Region)

			// Network configuration
			fmt.Println("\nNetwork Configuration:")
			fmt.Printf("  VPC ID: %s\n", profile.Network.VPCId)

			if len(profile.Network.SubnetIds) > 0 {
				fmt.Println("  Subnets:")
				for _, subnet := range profile.Network.SubnetIds {
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

			// Services
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

			// Tags
			if len(profile.Tags) > 0 {
				fmt.Println("\nTags:")
				for key, value := range profile.Tags {
					fmt.Printf("  %s: %s\n", key, value)
				}
			}

			if !profile.LastSync.IsZero() {
				fmt.Printf("\nLast Sync: %s\n", profile.LastSync.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	return cmd
}

// newCloudSwitchCmd creates the switch subcommand
func newCloudSwitchCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var dryRun bool
	var applyPolicy bool

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
					// TODO: Implement DNS configuration
				}
			}

			// Proxy configuration
			if profile.Network.Proxy != nil {
				if profile.Network.Proxy.HTTP != "" {
					fmt.Printf("  Setting HTTP proxy: %s\n", profile.Network.Proxy.HTTP)
					if !dryRun {
						os.Setenv("HTTP_PROXY", profile.Network.Proxy.HTTP)
						os.Setenv("http_proxy", profile.Network.Proxy.HTTP)
					}
				}
				if profile.Network.Proxy.HTTPS != "" {
					fmt.Printf("  Setting HTTPS proxy: %s\n", profile.Network.Proxy.HTTPS)
					if !dryRun {
						os.Setenv("HTTPS_PROXY", profile.Network.Proxy.HTTPS)
						os.Setenv("https_proxy", profile.Network.Proxy.HTTPS)
					}
				}
			}

			// VPN connection
			if profile.Network.VPN != nil && profile.Network.VPN.AutoConnect {
				fmt.Printf("  Connecting to VPN: %s\n", profile.Network.VPN.Server)
				if !dryRun {
					// TODO: Implement VPN connection
				}
			}

			// Custom routes
			if len(profile.Network.Routes) > 0 {
				fmt.Println("  Adding custom routes:")
				for _, route := range profile.Network.Routes {
					fmt.Printf("    %s via %s\n", route.Destination, route.Gateway)
					if !dryRun {
						// TODO: Implement route addition
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

// newCloudSyncCmd creates the sync subcommand
func newCloudSyncCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var source string
	var target string
	var profiles []string
	var conflictMode string
	var showStatus bool
	var showRecommendations bool

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
				fmt.Fprintln(w, "PROFILE\tSOURCE\tTARGET\tSTATUS\tLAST SYNC\tERROR")
				for _, s := range status {
					errorMsg := s.Error
					if errorMsg == "" {
						errorMsg = "-"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
						s.ProfileName,
						s.Source,
						s.Target,
						s.Status,
						s.LastSync.Format("2006-01-02 15:04:05"),
						errorMsg,
					)
				}
				w.Flush()
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

// newCloudValidateCmd creates the validate subcommand
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

// saveCurrentProfile saves the current profile name
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

// getCurrentProfile reads the current profile name
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

// newCloudPolicyCmd creates the policy subcommand
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

// newCloudPolicyApplyCmd creates the policy apply subcommand
func newCloudPolicyApplyCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var environment string
	var profileName string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply network policies",
		Long: `Apply network policies to cloud profiles automatically.

This command can apply policies either by environment (all profiles in an environment)
or by specific profile name.`,
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

			if dryRun {
				fmt.Println("[DRY RUN] Would apply the following policies:")
			}

			// Apply policies by environment or profile
			if environment != "" {
				fmt.Printf("Applying policies for environment: %s\n", environment)
				if !dryRun {
					if err := policyManager.ApplyEnvironmentPolicies(ctx, environment); err != nil {
						return fmt.Errorf("failed to apply environment policies: %w", err)
					}
				} else {
					// Show what would be applied
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
				}
			} else if profileName != "" {
				fmt.Printf("Applying policies for profile: %s\n", profileName)
				if !dryRun {
					if err := policyManager.ApplyPoliciesForProfile(ctx, profileName); err != nil {
						return fmt.Errorf("failed to apply profile policies: %w", err)
					}
				} else {
					// Show what would be applied
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
				}
			} else {
				return fmt.Errorf("either --environment or --profile must be specified")
			}

			if dryRun {
				fmt.Println("\n[DRY RUN] No policies were actually applied")
			} else {
				fmt.Println("✓ Policies applied successfully")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "environment", "e", "", "Apply policies for all profiles in environment")
	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Apply policies for specific profile")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without applying policies")

	return cmd
}

// newCloudPolicyListCmd creates the policy list subcommand
func newCloudPolicyListCmd(ctx context.Context, opts *cloudOptions) *cobra.Command {
	var profileName string
	var environment string
	var showDisabled bool

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

			if profileName != "" {
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
				fmt.Fprintln(w, "NAME\tPRIORITY\tENABLED\tRULES\tACTIONS")

				for _, policy := range policies {
					if !showDisabled && !policy.Enabled {
						continue
					}

					enabled := "✓"
					if !policy.Enabled {
						enabled = "✗"
					}

					fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%d\n",
						policy.Name,
						policy.Priority,
						enabled,
						len(policy.Rules),
						len(policy.Actions),
					)
				}
				w.Flush()
			} else if environment != "" {
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
			} else {
				// List all configured policies
				fmt.Println("Configured Policies:")
				fmt.Println("==================")

				if config.Policies == nil || len(config.Policies) == 0 {
					fmt.Println("No policies configured")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tPROFILE\tENVIRONMENT\tPRIORITY\tENABLED\tRULES\tACTIONS")

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

					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%d\t%d\n",
						policy.Name,
						profileName,
						environment,
						policy.Priority,
						enabled,
						len(policy.Rules),
						len(policy.Actions),
					)
				}
				w.Flush()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "List policies for specific profile")
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "List policies for environment")
	cmd.Flags().BoolVar(&showDisabled, "show-disabled", false, "Show disabled policies")

	return cmd
}

// newCloudPolicyStatusCmd creates the policy status subcommand
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
			fmt.Fprintln(w, "POLICY\tPROFILE\tPROVIDER\tSTATUS\tAPPLIED\tERROR")

			for _, s := range status {
				errorMsg := s.Error
				if errorMsg == "" {
					errorMsg = "-"
				}

				appliedTime := s.Applied.Format("2006-01-02 15:04:05")

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					s.PolicyName,
					s.ProfileName,
					s.Provider,
					s.Status,
					appliedTime,
					errorMsg,
				)
			}
			w.Flush()

			return nil
		},
	}

	return cmd
}

// newCloudPolicyValidateCmd creates the policy validate subcommand
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
					if err := policyManager.ValidatePolicy(policy); err != nil {
						validationErrors = append(validationErrors, fmt.Sprintf("Policy %s: %v", policy.Name, err))
						fmt.Printf("  ✗ %s: %v\n", policy.Name, err)
					} else {
						fmt.Printf("  ✓ %s: valid\n", policy.Name)
					}
				}
			} else {
				// Validate all configured policies
				if config.Policies == nil || len(config.Policies) == 0 {
					fmt.Println("No policies configured")
					return nil
				}

				for _, policy := range config.Policies {
					if err := policyManager.ValidatePolicy(&policy); err != nil {
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

// Helper function to get profiles for environment
func getProfilesForEnvironment(config *cloud.Config, environment string) []cloud.Profile {
	var profiles []cloud.Profile
	for _, profile := range config.Profiles {
		if profile.Environment == environment {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

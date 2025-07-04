package bulkclone

import (
	"fmt"

	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	"github.com/spf13/cobra"
)

type bulkCloneOptions struct {
	configFile     string
	useConfig      bool
	useGZHConfig   bool
	strategy       string
	providerFilter string
}

func defaultBulkCloneOptions() *bulkCloneOptions {
	return &bulkCloneOptions{
		strategy: "reset",
	}
}

func NewBulkCloneCmd() *cobra.Command {
	o := defaultBulkCloneOptions()

	cmd := &cobra.Command{
		Use:          "bulk-clone",
		Short:        "Clone repositories from multiple Git hosting services",
		SilenceUsage: true,
		Long: `Clone multiple repositories from various Git hosting services.
		
You can use a configuration file (bulk-clone.yaml) to define multiple organizations
and their settings. This command will process all repository roots defined in the
configuration file regardless of the provider (GitHub, GitLab, Gitea).

For provider-specific operations, use the subcommands (github, gitlab, etc.).`,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().BoolVar(&o.useGZHConfig, "use-gzh-config", false, "Use gzh.yaml configuration format")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVar(&o.providerFilter, "provider", "", "Filter by provider: github, gitlab, gitea")

	// Mark flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("config", "use-config", "use-gzh-config")

	cmd.AddCommand(newBulkCloneGiteaCmd())
	cmd.AddCommand(newBulkCloneGithubCmd())
	cmd.AddCommand(newBulkCloneGitlabCmd())
	cmd.AddCommand(newBulkCloneValidateCmd())

	return cmd
}

func (o *bulkCloneOptions) run(_ *cobra.Command, args []string) error {
	// Check if using gzh.yaml format
	if o.useGZHConfig {
		return o.runWithGZHConfig()
	}

	// Original bulk-clone.yaml format handling
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := bulkclonepkg.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// Process all repository roots
	for _, repoRoot := range cfg.RepoRoots {
		// Filter by provider if specified
		if o.providerFilter != "" && repoRoot.Provider != o.providerFilter {
			continue
		}

		fmt.Printf("Processing %s organization: %s -> %s\n", repoRoot.Provider, repoRoot.OrgName, repoRoot.RootPath)

		// Expand the root path
		targetPath := bulkclonepkg.ExpandPath(repoRoot.RootPath)

		switch repoRoot.Provider {
		case "github":
			err = github.RefreshAll(targetPath, repoRoot.OrgName, o.strategy)
			if err != nil {
				fmt.Printf("Error processing GitHub org %s: %v\n", repoRoot.OrgName, err)
				continue
			}
		case "gitlab":
			// For GitLab, we need to use the group name from the repoRoot
			// Since RepoRoots currently only has GitHub structure, we'll use OrgName as GroupName
			err = gitlab.RefreshAll(targetPath, repoRoot.OrgName, o.strategy)
			if err != nil {
				fmt.Printf("Error processing GitLab group %s: %v\n", repoRoot.OrgName, err)
				continue
			}
		default:
			fmt.Printf("Unsupported provider: %s (skipping %s)\n", repoRoot.Provider, repoRoot.OrgName)
			continue
		}

		fmt.Printf("✓ Successfully processed %s/%s\n", repoRoot.Provider, repoRoot.OrgName)
	}

	// Also process default GitHub and GitLab configurations if they have org/group names
	if cfg.Default.Github.OrgName != "" {
		fmt.Printf("Processing default GitHub organization: %s\n", cfg.Default.Github.OrgName)
		targetPath := bulkclonepkg.ExpandPath(cfg.Default.Github.RootPath)
		err = github.RefreshAll(targetPath, cfg.Default.Github.OrgName, o.strategy)
		if err != nil {
			fmt.Printf("Error processing default GitHub org: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully processed default GitHub org: %s\n", cfg.Default.Github.OrgName)
		}
	}

	if cfg.Default.Gitlab.GroupName != "" {
		fmt.Printf("Processing default GitLab group: %s\n", cfg.Default.Gitlab.GroupName)
		targetPath := bulkclonepkg.ExpandPath(cfg.Default.Gitlab.RootPath)
		err = gitlab.RefreshAll(targetPath, cfg.Default.Gitlab.GroupName, o.strategy)
		if err != nil {
			fmt.Printf("Error processing default GitLab group: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully processed default GitLab group: %s\n", cfg.Default.Gitlab.GroupName)
		}
	}

	fmt.Println("Bulk clone operation completed!")
	return nil
}

// runWithGZHConfig handles bulk cloning using gzh.yaml configuration format
func (o *bulkCloneOptions) runWithGZHConfig() error {
	// Load gzh.yaml configuration
	gzhConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load gzh.yaml config: %w", err)
	}

	// Create integration wrapper
	integration := config.NewBulkCloneIntegration(gzhConfig)

	// Get all targets or filter by provider
	var targets []config.BulkCloneTarget
	if o.providerFilter != "" {
		if err := integration.ValidateProvider(o.providerFilter); err != nil {
			return fmt.Errorf("invalid provider filter: %w", err)
		}
		targets, err = integration.GetTargetsByProvider(o.providerFilter)
		if err != nil {
			return fmt.Errorf("failed to get targets for provider %s: %w", o.providerFilter, err)
		}
	} else {
		targets, err = integration.GetAllTargets()
		if err != nil {
			return fmt.Errorf("failed to get all targets: %w", err)
		}
	}

	if len(targets) == 0 {
		fmt.Println("No targets found to process")
		return nil
	}

	fmt.Printf("🚀 Starting bulk clone operation with %d targets\n", len(targets))

	// Process each target
	successCount := 0
	for i, target := range targets {
		fmt.Printf("\n[%d/%d] Processing %s: %s\n", i+1, len(targets), target.Provider, target.Name)

		// Override strategy if specified via command line
		if o.strategy != "reset" {
			target.Strategy = o.strategy
		}

		// Validate strategy
		if target.Strategy != "reset" && target.Strategy != "pull" && target.Strategy != "fetch" {
			fmt.Printf("⚠️  Invalid strategy '%s' for %s/%s, using 'reset'\n", target.Strategy, target.Provider, target.Name)
			target.Strategy = "reset"
		}

		// Expand clone directory
		targetPath := config.ExpandEnvironmentVariables(target.CloneDir)
		fmt.Printf("   📁 Target directory: %s\n", targetPath)
		fmt.Printf("   🔧 Strategy: %s\n", target.Strategy)

		if target.Token == "" {
			fmt.Printf("⚠️  Warning: No token configured for %s/%s\n", target.Provider, target.Name)
		}

		// Execute cloning based on provider
		err := o.executeProviderCloning(target, targetPath)
		if err != nil {
			fmt.Printf("❌ Error processing %s/%s: %v\n", target.Provider, target.Name, err)
			continue
		}

		fmt.Printf("✅ Successfully processed %s/%s\n", target.Provider, target.Name)
		successCount++
	}

	fmt.Printf("\n🎉 Bulk clone operation completed! (%d/%d successful)\n", successCount, len(targets))
	return nil
}

// executeProviderCloning executes the cloning operation for a specific provider
func (o *bulkCloneOptions) executeProviderCloning(target config.BulkCloneTarget, targetPath string) error {
	switch target.Provider {
	case config.ProviderGitHub:
		return github.RefreshAll(targetPath, target.Name, target.Strategy)
	case config.ProviderGitLab:
		return gitlab.RefreshAll(targetPath, target.Name, target.Strategy)
	case config.ProviderGitea:
		// Gitea support would go here
		return fmt.Errorf("gitea provider not yet implemented for gzh.yaml format")
	default:
		return fmt.Errorf("unsupported provider: %s", target.Provider)
	}
}

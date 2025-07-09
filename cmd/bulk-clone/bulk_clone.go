package bulkclone

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/config"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	pkgconfig "github.com/gizzahub/gzh-manager-go/pkg/config"
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
	parallel       int
	maxRetries     int
	resume         bool
	progressMode   string
}

func defaultBulkCloneOptions() *bulkCloneOptions {
	return &bulkCloneOptions{
		strategy:     "reset",
		parallel:     10,
		maxRetries:   3,
		progressMode: "compact",
	}
}

func NewBulkCloneCmd(ctx context.Context) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ctx, cmd, args)
		},
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().BoolVar(&o.useGZHConfig, "use-gzh-config", false, "Use gzh.yaml configuration format")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVar(&o.providerFilter, "provider", "", "Filter by provider: github, gitlab, gitea")
	cmd.Flags().IntVarP(&o.parallel, "parallel", "p", o.parallel, "Number of parallel workers for cloning")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", o.maxRetries, "Maximum retry attempts for failed operations")
	cmd.Flags().BoolVar(&o.resume, "resume", false, "Resume interrupted clone operation from saved state")
	cmd.Flags().StringVar(&o.progressMode, "progress-mode", o.progressMode, "Progress display mode: compact, detailed, quiet")

	// Mark flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("config", "use-config", "use-gzh-config")

	cmd.AddCommand(newBulkCloneGiteaCmd())
	cmd.AddCommand(newBulkCloneGithubCmd())
	cmd.AddCommand(newBulkCloneGitlabCmd())
	cmd.AddCommand(newBulkCloneValidateCmd())
	cmd.AddCommand(newBulkCloneStateCmd())

	return cmd
}

func (o *bulkCloneOptions) run(ctx context.Context, _ *cobra.Command, args []string) error {
	// Use central configuration service for unified configuration management
	return o.runWithCentralConfigService(ctx)
}

// runWithCentralConfigService uses the central configuration service for unified config management
func (o *bulkCloneOptions) runWithCentralConfigService(ctx context.Context) error {
	// Create configuration service
	configService, err := config.CreateDefaultConfigService()
	if err != nil {
		return fmt.Errorf("failed to create configuration service: %w", err)
	}

	// Load configuration (supports both unified and legacy formats with auto-migration)
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := configService.LoadConfiguration(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Show migration warnings if any
	warnings := configService.GetWarnings()
	for _, warning := range warnings {
		fmt.Printf("⚠ Warning: %s\n", warning)
	}

	// Show required actions if any
	actions := configService.GetRequiredActions()
	for _, action := range actions {
		fmt.Printf("📋 Action required: %s\n", action)
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// Get bulk clone targets using the service
	targets, err := configService.GetBulkCloneTargets(ctx, o.providerFilter)
	if err != nil {
		return fmt.Errorf("failed to get bulk clone targets: %w", err)
	}

	if len(targets) == 0 {
		fmt.Println("No targets found to process")
		return nil
	}

	fmt.Printf("Found %d targets to process\n", len(targets))

	// Process each target
	for _, target := range targets {
		// Check for cancellation before starting each target
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		fmt.Printf("Processing %s organization: %s -> %s\n", target.Provider, target.Name, target.CloneDir)

		// Use the strategy from target or override from command line
		strategy := target.Strategy
		if o.strategy != "reset" { // If user specified a different strategy
			strategy = o.strategy
		}

		err := o.executeProviderCloning(ctx, target, target.CloneDir)
		if err != nil {
			fmt.Printf("❌ Error processing %s/%s: %v\n", target.Provider, target.Name, err)
			continue
		}

		fmt.Printf("✅ Successfully processed %s/%s\n", target.Provider, target.Name)
	}

	return nil
}

// runWithGZHConfig handles bulk cloning using gzh.yaml configuration format
func (o *bulkCloneOptions) runWithGZHConfig(ctx context.Context) error {
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
		// Check for cancellation before processing each target
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled after processing %d/%d targets: %w", i, len(targets), ctx.Err())
		default:
		}

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
		err := o.executeProviderCloning(ctx, target, targetPath)
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
func (o *bulkCloneOptions) executeProviderCloning(ctx context.Context, target config.BulkCloneTarget, targetPath string) error {
	switch target.Provider {
	case config.ProviderGitHub:
		// Use resumable clone if requested or if parallel/worker pool is enabled
		if o.resume || o.parallel > 1 {
			return github.RefreshAllResumable(ctx, targetPath, target.Name, target.Strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		}
		return github.RefreshAll(ctx, targetPath, target.Name, target.Strategy)
	case config.ProviderGitLab:
		// Use resumable clone if requested or if parallel/worker pool is enabled
		if o.resume || o.parallel > 1 {
			return gitlab.RefreshAllResumable(ctx, targetPath, target.Name, target.Strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		}
		return gitlab.RefreshAll(ctx, targetPath, target.Name, target.Strategy)
	case config.ProviderGitea:
		// Gitea support would go here
		return fmt.Errorf("gitea provider not yet implemented for gzh.yaml format")
	default:
		return fmt.Errorf("unsupported provider: %s", target.Provider)
	}
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/config"
	gerrors "github.com/Gizzahub/gzh-cli/internal/errors"
	pkgconfig "github.com/Gizzahub/gzh-cli/pkg/config"
	"github.com/Gizzahub/gzh-cli/pkg/github"
	"github.com/Gizzahub/gzh-cli/pkg/gitlab"
)

type syncCloneOptions struct {
	configFile     string
	useConfig      bool
	useGZHConfig   bool
	strategy       string
	providerFilter string
	parallel       int
	maxRetries     int
	resume         bool
	progressMode   string
	cleanupOrphans bool
}

func defaultSyncCloneOptions() *syncCloneOptions {
	return &syncCloneOptions{
		strategy:       "reset",
		parallel:       10,
		maxRetries:     3,
		progressMode:   "bar",
		cleanupOrphans: false,
	}
}

// NewSyncCloneCmd creates a new cobra command for synchronous repository cloning.
// This command enables cloning multiple repositories from various Git hosting
// services including GitHub, GitLab, Gitea, and Gogs using configuration files
// or command-line flags.
//
// Parameters:
//   - ctx: Context for operation cancellation and timeout control
//
// Returns a configured cobra.Command ready for execution.
func NewSyncCloneCmd(ctx context.Context) *cobra.Command {
	o := defaultSyncCloneOptions()

	cmd := &cobra.Command{
		Use:          "synclone",
		Short:        "Synchronize and clone repositories from multiple Git hosting services",
		SilenceUsage: true,
		Long: `Synchronize and clone multiple repositories from various Git hosting services.

You can use a configuration file (synclone.yaml) to define multiple organizations
and their settings. This command will process all repository roots defined in the
configuration file regardless of the provider (GitHub, GitLab, Gitea).

When targeting an organization, a gzh.yaml file will be created in the target directory
containing the repository list for future reference and synchronization.

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
	cmd.Flags().StringVar(&o.progressMode, "progress-mode", o.progressMode, "Progress display mode: bar, dots, spinner, quiet")
	cmd.Flags().BoolVar(&o.cleanupOrphans, "cleanup-orphans", o.cleanupOrphans, "Remove directories not present in the organization's repositories")

	// Mark flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("config", "use-config", "use-gzh-config")

	cmd.AddCommand(newSyncCloneConfigCmd())
	cmd.AddCommand(newSyncCloneGiteaCmd())
	cmd.AddCommand(newSyncCloneGithubCmd())
	cmd.AddCommand(newSyncCloneGitlabCmd())
	cmd.AddCommand(newSyncCloneValidateCmd())
	cmd.AddCommand(newSyncCloneStateCmd())

	return cmd
}

func (o *syncCloneOptions) run(ctx context.Context, _ *cobra.Command, _ []string) error {
	// Use central configuration service for unified configuration management
	return o.runWithCentralConfigService(ctx)
}

// runWithCentralConfigService uses the central configuration service for unified config management.
func (o *syncCloneOptions) runWithCentralConfigService(ctx context.Context) error {
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

	_, err = configService.LoadConfiguration(ctx, configPath)
	if err != nil {
		// If no configuration file is found and no specific config file was provided,
		// show help instead of erroring out
		if configPath == "" && isConfigNotFoundError(err) {
			fmt.Println("No configuration file found. Please provide a configuration file or use one of the subcommands.")
			fmt.Println("\nAvailable subcommands:")
			fmt.Println("  github   - Clone from GitHub organizations")
			fmt.Println("  gitlab   - Clone from GitLab groups")
			fmt.Println("  gitea    - Clone from Gitea organizations")
			fmt.Println("  validate - Validate configuration file")
			fmt.Println("  state    - Manage operation state")
			fmt.Println("\nExample configuration files can be found in the examples/ directory.")
			fmt.Println("Use --help or -h for more detailed usage information.")
			return nil
		}
		return gerrors.Wrap(err, gerrors.ErrConfigNotFound)
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
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		default:
		}

		fmt.Printf("Processing %s organization: %s -> %s\n", target.Provider, target.Name, target.CloneDir)

		err := o.executeProviderCloning(ctx, target, target.CloneDir)
		if err != nil {
			fmt.Printf("❌ Error processing %s/%s: %v\n", target.Provider, target.Name, err)
			continue
		}

		fmt.Printf("✅ Successfully processed %s/%s\n", target.Provider, target.Name)
	}

	return nil
}

// executeProviderCloning executes the cloning operation for a specific provider.
func (o *syncCloneOptions) executeProviderCloning(ctx context.Context, target pkgconfig.BulkCloneTarget, targetPath string) error {
	switch target.Provider {
	case pkgconfig.ProviderGitHub:
		// Use resumable clone if requested or if parallel/worker pool is enabled
		if o.resume || o.parallel > 1 {
			return github.RefreshAllResumable(ctx, targetPath, target.Name, target.Strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		}

		return github.RefreshAll(ctx, targetPath, target.Name, target.Strategy)
	case pkgconfig.ProviderGitLab:
		// Use resumable clone if requested or if parallel/worker pool is enabled
		if o.resume || o.parallel > 1 {
			return gitlab.RefreshAllResumable(ctx, targetPath, target.Name, target.Strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		}

		return gitlab.RefreshAll(ctx, targetPath, target.Name, target.Strategy)
	case pkgconfig.ProviderGitea:
		// Gitea support would go here
		return fmt.Errorf("gitea provider not yet implemented for gzh.yaml format")
	default:
		return fmt.Errorf("unsupported provider: %s", target.Provider)
	}
}

// isConfigNotFoundError checks if the error indicates a configuration file was not found.
func isConfigNotFoundError(err error) bool {
	return errors.Is(err, gerrors.ErrConfigNotFound)
}

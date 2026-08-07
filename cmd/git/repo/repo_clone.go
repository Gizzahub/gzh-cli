// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	"github.com/gizzahub/gzh-cli/internal/git/clone"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
	"github.com/gizzahub/gzh-cli/pkg/gitea"
	"github.com/gizzahub/gzh-cli/pkg/github"
	"github.com/gizzahub/gzh-cli/pkg/gitlab"
)

// cloneGitRunner is a test seam. When non-nil, runRepoClone injects it into the
// CloneExecutor so tests never shell out to real git / the network.
var cloneGitRunner gitinternal.GitRunner

// newRepoCloneCmd creates the git repo clone command.
func newRepoCloneCmd() *cobra.Command {
	opts := clone.DefaultCloneOptions()

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone repositories from Git hosting platforms",
		Long: `Clone repositories with advanced features:

- Bulk operations for entire organizations/groups
- Parallel execution with configurable workers
- Resume capability for interrupted operations
- Multiple clone strategies (reset, pull, fetch)
- Advanced filtering and matching
- Multiple output formats

This command uses the modern provider abstraction layer to support
GitHub and GitLab platforms through a unified interface.
Gitea and Gogs providers are planned for future implementation.

Note: This is an experimental command. For production use, consider
using the stable 'gz synclone' command.`,
		Example: `  # Clone all repos from GitHub organization
  gz git repo clone --provider github --org myorg

  # Clone with filters and custom target
  gz git repo clone --provider gitlab --org mygroup --match "api-*" --target ./projects

  # Clone with parallel workers and specific strategy
  gz git repo clone --provider github --org myorg --parallel 10 --strategy pull

  # Resume interrupted operation
  gz git repo clone --resume abc12345

  # Dry run to preview what would be cloned
  gz git repo clone --provider github --org myorg --dry-run

  # Clone private repos only with SSH protocol
  gz git repo clone --provider github --org myorg --visibility private --protocol ssh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoClone(cmd.Context(), opts)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")

	// Target and configuration
	cmd.Flags().StringVar(&opts.Target, "target", ".", "Target directory for cloned repositories")
	cmd.Flags().StringVar(&opts.Config, "config", "", "Path to configuration file")

	// Execution options
	cmd.Flags().IntVar(&opts.Parallel, "parallel", 5, "Number of parallel workers (1-50)")
	cmd.Flags().StringVar((*string)(&opts.Strategy), "strategy", string(clone.StrategyReset),
		fmt.Sprintf("Clone strategy (%s)", formatStrategies()))
	cmd.Flags().StringVar(&opts.Resume, "resume", "", "Resume session ID")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 30*time.Minute, "Operation timeout")
	cmd.Flags().IntVar(&opts.MaxRetries, "max-retries", 3, "Maximum retry attempts")
	cmd.Flags().DurationVar(&opts.RetryDelay, "retry-delay", 1*time.Second, "Delay between retries")

	// Filtering options
	cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern (regex)")
	cmd.Flags().StringVar(&opts.Exclude, "exclude", "", "Repository exclusion pattern (regex)")
	cmd.Flags().StringVar(&opts.Visibility, "visibility", "all", "Repository visibility (all, public, private)")
	cmd.Flags().BoolVar(&opts.IncludeArchived, "include-archived", false, "Include archived repositories")
	cmd.Flags().BoolVar(&opts.IncludeForks, "include-forks", true, "Include forked repositories")
	cmd.Flags().StringVar(&opts.Language, "language", "", "Filter by primary language")
	cmd.Flags().StringSliceVar(&opts.Topics, "topics", nil, "Filter by topics (comma-separated)")
	cmd.Flags().IntVar(&opts.MinStars, "min-stars", 0, "Minimum star count")
	cmd.Flags().IntVar(&opts.MaxStars, "max-stars", 0, "Maximum star count (0 = unlimited)")
	cmd.Flags().StringVar(&opts.UpdatedSince, "updated-since", "", "Only repos updated since date (YYYY-MM-DD)")

	// Output and behavior
	cmd.Flags().StringVar(&opts.Format, "format", string(clone.FormatProgress),
		fmt.Sprintf("Output format (%s)", formatOutputFormats()))
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview repositories without cloning")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", false, "Suppress progress output")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Verbose output")
	cmd.Flags().BoolVar(&opts.CleanupOrphans, "cleanup-orphans", false, "Remove directories not in organization")
	cmd.Flags().BoolVar(&opts.CreateGZHFile, "create-gzh-file", true, "Create .gzh metadata files")

	// Authentication
	cmd.Flags().StringVar(&opts.Token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&opts.Username, "username", "", "Username for authentication")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Password for authentication")

	// Git options
	cmd.Flags().StringVar(&opts.Protocol, "protocol", "https", "Git protocol (https, ssh)")
	cmd.Flags().IntVar(&opts.Depth, "depth", 0, "Clone depth (0 = full clone)")
	cmd.Flags().BoolVar(&opts.SingleBranch, "single-branch", false, "Clone single branch only")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Specific branch to clone")

	// Flag validations and relationships
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagsMutuallyExclusive("quiet", "verbose")
	cmd.MarkFlagsMutuallyExclusive("resume", "provider")
	cmd.MarkFlagsMutuallyExclusive("resume", "org")

	return cmd
}

// runRepoClone executes the repository clone operation.
func runRepoClone(ctx context.Context, opts *clone.CloneOptions) error {
	// Handle resume mode differently
	if opts.Resume != "" {
		return runResumeClone(ctx, opts)
	}

	gitProvider, err := getCloneProvider(opts)
	if err != nil {
		return err
	}

	// Create clone executor; inject test runner when the seam is set.
	executor, err := clone.NewCloneExecutorWithRunner(gitProvider, opts, cloneGitRunner)
	if err != nil {
		return fmt.Errorf("failed to create clone executor: %w", err)
	}

	// Execute clone operation
	return executor.Execute(ctx)
}

// runResumeClone handles resuming a clone operation.
func runResumeClone(ctx context.Context, opts *clone.CloneOptions) error {
	// Load session to get original options
	session := clone.NewSession(opts)
	if err := session.Load(opts.Resume); err != nil {
		return fmt.Errorf("failed to load session %s: %w", opts.Resume, err)
	}

	// Use original options but allow certain overrides
	originalOpts := session.Options

	// Allow overriding certain options for resume
	if opts.Parallel > 0 {
		originalOpts.Parallel = opts.Parallel
	}
	if opts.MaxRetries > 0 {
		originalOpts.MaxRetries = opts.MaxRetries
	}
	if opts.Format != "" {
		originalOpts.Format = opts.Format
	}
	if opts.Quiet {
		originalOpts.Quiet = opts.Quiet
	}
	if opts.Verbose {
		originalOpts.Verbose = opts.Verbose
	}

	// Set resume flag
	originalOpts.Resume = opts.Resume

	// Run with original options
	return runRepoClone(ctx, originalOpts)
}

// getCloneProvider resolves the provider the clone executor runs against. It is
// a separate seam from getGitProvider because clone takes its credentials from
// its own flags rather than from the environment; both exist so tests can
// substitute a double instead of reaching the live API.
var getCloneProvider = newCloneProvider

// newCloneProvider builds a real provider from the clone command's options.
func newCloneProvider(opts *clone.CloneOptions) (provider.GitProvider, error) {
	factory := provider.NewProviderFactory()

	if err := registerProviderConstructors(factory); err != nil {
		return nil, fmt.Errorf("failed to register providers: %w\n\nNote: For production use, consider using the 'gz synclone' command which is fully stable:\n  gz synclone --config examples/synclone/synclone-example.yaml", err)
	}

	providerConfig := &provider.ProviderConfig{
		Type:     opts.Provider,
		Name:     fmt.Sprintf("%s-clone", opts.Provider),
		Token:    opts.Token,
		Username: opts.Username,
		Password: opts.Password,
		Enabled:  true,
		Extra:    make(map[string]any),
	}

	if err := factory.RegisterConfig(providerConfig.Name, providerConfig); err != nil {
		return nil, fmt.Errorf("failed to register provider config: %w", err)
	}

	registry := provider.NewProviderRegistry(factory, provider.RegistryConfig{
		EnableCaching:      true,
		EnableHealthChecks: false, // Skip health checks for one-time operations
		CacheTimeout:       5 * time.Minute,
	})

	gitProvider, err := registry.GetProvider(providerConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return gitProvider, nil
}

// registerProviderConstructors registers provider constructors with the factory.
func registerProviderConstructors(factory *provider.ProviderFactory) error {
	constructors := []struct {
		name string
		ctor provider.ProviderConstructor
	}{
		{"github", github.CreateGitHubProvider},
		{"gitlab", gitlab.CreateGitLabProvider},
		{"gitea", gitea.CreateGiteaProvider},
	}

	for _, c := range constructors {
		if err := factory.RegisterProvider(c.name, c.ctor); err != nil {
			return fmt.Errorf("failed to register %s provider: %w", c.name, err)
		}
	}

	// Gogs has no client package yet, so it stays a stub. The error names the
	// endpoint it would have used, which is the only useful thing it can say.
	if err := factory.RegisterProvider("gogs", func(config *provider.ProviderConfig) (provider.GitProvider, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://try.gogs.io/api/v1"
		}

		return nil, fmt.Errorf("gogs provider is not implemented yet (would connect to %s)", baseURL)
	}); err != nil {
		return fmt.Errorf("failed to register gogs provider: %w", err)
	}

	return nil
}

// formatStrategies returns a formatted string of valid strategies.
func formatStrategies() string {
	strategies := clone.GetValidStrategies()
	var result strings.Builder
	for i, strategy := range strategies {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(strategy)
	}
	return result.String()
}

// formatOutputFormats returns a formatted string of valid output formats.
func formatOutputFormats() string {
	formats := clone.GetValidFormats()
	var result strings.Builder
	for i, format := range formats {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(format)
	}
	return result.String()
}

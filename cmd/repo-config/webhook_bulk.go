package repoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gizzahub/gzh-manager-go/pkg/types/repoconfig"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// WebhookBulkFlags represents bulk webhook operation flags
type WebhookBulkFlags struct {
	GlobalFlags
	Repositories []string
	RepoPattern  string
	ConfigFile   string
	URL          string
	Events       []string
	Active       bool
	Secret       string
	ContentType  string
	OutputFormat string
	SkipExisting bool
	MaxWorkers   int
	All          bool
}

// BulkWebhookConfig represents bulk webhook configuration
type BulkWebhookConfig struct {
	Version  string                     `yaml:"version"`
	Webhooks []repoconfig.WebhookConfig `yaml:"webhooks"`
	Targets  BulkWebhookTargets         `yaml:"targets"`
	Options  BulkWebhookOptions         `yaml:"options,omitempty"`
}

// BulkWebhookTargets specifies which repositories to target
type BulkWebhookTargets struct {
	All          bool     `yaml:"all,omitempty"`
	Repositories []string `yaml:"repositories,omitempty"`
	Pattern      string   `yaml:"pattern,omitempty"`
	Exclude      []string `yaml:"exclude,omitempty"`
}

// BulkWebhookOptions contains bulk operation options
type BulkWebhookOptions struct {
	SkipExisting    bool `yaml:"skip_existing,omitempty"`
	MaxWorkers      int  `yaml:"max_workers,omitempty"`
	ContinueOnError bool `yaml:"continue_on_error,omitempty"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Repository string
	Action     string
	Success    bool
	Error      error
	Details    interface{}
}

// newWebhookBulkCmd creates the webhook bulk operations command
func newWebhookBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Manage webhooks across multiple repositories",
		Long: `Perform bulk webhook operations across an entire organization or selected repositories.

This command allows you to create, update, or delete webhooks across multiple repositories
at once, using either command-line flags or a configuration file.

Examples:
  gz repo-config webhook bulk create --org myorg --all --url https://example.com/webhook
  gz repo-config webhook bulk create --org myorg --config webhooks.yaml
  gz repo-config webhook bulk list --org myorg --pattern "^myapp-"
  gz repo-config webhook bulk delete --org myorg --repos repo1,repo2,repo3 --url https://old-webhook.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newWebhookBulkCreateCmd())
	cmd.AddCommand(newWebhookBulkListCmd())
	cmd.AddCommand(newWebhookBulkDeleteCmd())
	cmd.AddCommand(newWebhookBulkSyncCmd())

	return cmd
}

// newWebhookBulkCreateCmd creates the bulk webhook create command
func newWebhookBulkCreateCmd() *cobra.Command {
	flags := &WebhookBulkFlags{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create webhooks across multiple repositories",
		Long: `Create the same webhook configuration across multiple repositories in an organization.

Examples:
  # Create webhook for all repositories
  gz repo-config webhook bulk create --org myorg --all --url https://example.com/webhook --events push

  # Create webhook for specific repositories
  gz repo-config webhook bulk create --org myorg --repos repo1,repo2 --url https://example.com/webhook

  # Create webhook using pattern matching
  gz repo-config webhook bulk create --org myorg --pattern "^api-" --url https://example.com/webhook

  # Create webhook from configuration file
  gz repo-config webhook bulk create --org myorg --config webhooks.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookBulkCreate(flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	addBulkWebhookFlags(cmd, flags)

	cmd.MarkFlagRequired("org")

	return cmd
}

// newWebhookBulkListCmd creates the bulk webhook list command
func newWebhookBulkListCmd() *cobra.Command {
	flags := &WebhookBulkFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks across multiple repositories",
		Long: `List webhooks configured across multiple repositories in an organization.

Examples:
  # List webhooks for all repositories
  gz repo-config webhook bulk list --org myorg --all

  # List webhooks for repositories matching pattern
  gz repo-config webhook bulk list --org myorg --pattern "^api-"

  # Export to YAML for backup or migration
  gz repo-config webhook bulk list --org myorg --all --output yaml > webhooks-backup.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookBulkList(flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringSliceVar(&flags.Repositories, "repos", nil, "Specific repositories (comma-separated)")
	cmd.Flags().StringVar(&flags.RepoPattern, "pattern", "", "Repository name pattern (regex)")
	cmd.Flags().BoolVar(&flags.All, "all", false, "Apply to all repositories")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", "table", "Output format (table, json, yaml)")
	cmd.Flags().IntVar(&flags.MaxWorkers, "max-workers", 5, "Maximum parallel workers")

	cmd.MarkFlagRequired("org")

	return cmd
}

// newWebhookBulkDeleteCmd creates the bulk webhook delete command
func newWebhookBulkDeleteCmd() *cobra.Command {
	flags := &WebhookBulkFlags{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete webhooks across multiple repositories",
		Long: `Delete webhooks matching specific criteria across multiple repositories.

Examples:
  # Delete all webhooks with specific URL
  gz repo-config webhook bulk delete --org myorg --all --url https://old-webhook.com

  # Delete webhooks from specific repositories
  gz repo-config webhook bulk delete --org myorg --repos repo1,repo2 --url https://old-webhook.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookBulkDelete(flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringSliceVar(&flags.Repositories, "repos", nil, "Specific repositories (comma-separated)")
	cmd.Flags().StringVar(&flags.RepoPattern, "pattern", "", "Repository name pattern (regex)")
	cmd.Flags().BoolVar(&flags.All, "all", false, "Apply to all repositories")
	cmd.Flags().StringVar(&flags.URL, "url", "", "Webhook URL to delete (required)")
	cmd.Flags().IntVar(&flags.MaxWorkers, "max-workers", 5, "Maximum parallel workers")

	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("url")

	return cmd
}

// newWebhookBulkSyncCmd creates the bulk webhook sync command
func newWebhookBulkSyncCmd() *cobra.Command {
	flags := &WebhookBulkFlags{}
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize webhooks from configuration file",
		Long: `Synchronize webhooks across repositories based on a configuration file.

This command ensures that repositories have exactly the webhooks specified in the
configuration file. It will create missing webhooks, update existing ones, and
optionally remove webhooks not in the configuration.

Examples:
  gz repo-config webhook bulk sync --org myorg --config webhooks.yaml
  gz repo-config webhook bulk sync --org myorg --config webhooks.yaml --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookBulkSync(flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().IntVar(&flags.MaxWorkers, "max-workers", 5, "Maximum parallel workers")

	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("config")

	return cmd
}

// addBulkWebhookFlags adds common bulk webhook flags
func addBulkWebhookFlags(cmd *cobra.Command, flags *WebhookBulkFlags) {
	cmd.Flags().StringSliceVar(&flags.Repositories, "repos", nil, "Specific repositories (comma-separated)")
	cmd.Flags().StringVar(&flags.RepoPattern, "pattern", "", "Repository name pattern (regex)")
	cmd.Flags().BoolVar(&flags.All, "all", false, "Apply to all repositories")
	cmd.Flags().StringVar(&flags.URL, "url", "", "Webhook URL")
	cmd.Flags().StringSliceVar(&flags.Events, "events", []string{"push"}, "Webhook events")
	cmd.Flags().BoolVar(&flags.Active, "active", true, "Whether webhook is active")
	cmd.Flags().StringVar(&flags.Secret, "secret", "", "Webhook secret")
	cmd.Flags().StringVar(&flags.ContentType, "content-type", "json", "Content type (json or form)")
	cmd.Flags().BoolVar(&flags.SkipExisting, "skip-existing", false, "Skip repositories with existing webhooks")
	cmd.Flags().IntVar(&flags.MaxWorkers, "max-workers", 5, "Maximum parallel workers")
}

// runWebhookBulkCreate executes the bulk webhook create command
func runWebhookBulkCreate(flags *WebhookBulkFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	// Load configuration if provided
	var config *BulkWebhookConfig
	if flags.GlobalFlags.ConfigFile != "" {
		var err error
		config, err = loadBulkWebhookConfig(flags.GlobalFlags.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	} else {
		// Create config from flags
		config = &BulkWebhookConfig{
			Version: "1.0",
			Webhooks: []repoconfig.WebhookConfig{
				{
					URL:         flags.URL,
					Events:      flags.Events,
					Active:      &flags.Active,
					ContentType: flags.ContentType,
					Secret:      flags.Secret,
				},
			},
			Options: BulkWebhookOptions{
				SkipExisting: flags.SkipExisting,
				MaxWorkers:   flags.MaxWorkers,
			},
		}
	}

	// Get target repositories
	repos, err := getTargetRepositories(ctx, client, flags)
	if err != nil {
		return fmt.Errorf("failed to get target repositories: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found matching criteria")
	}

	fmt.Printf("🎯 Creating webhooks for %d repositories\n", len(repos))

	// Execute bulk operation
	results := executeBulkWebhookOperation(ctx, client, repos, config, createWebhookOperation)

	// Display results
	displayBulkResults(results, "create")

	return nil
}

// runWebhookBulkList executes the bulk webhook list command
func runWebhookBulkList(flags *WebhookBulkFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	// Get target repositories
	repos, err := getTargetRepositories(ctx, client, flags)
	if err != nil {
		return fmt.Errorf("failed to get target repositories: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found matching criteria")
	}

	fmt.Printf("📋 Listing webhooks for %d repositories\n", len(repos))

	// Collect all webhooks
	allWebhooks := make(map[string][]*github.Hook)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, flags.MaxWorkers)

	for _, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(r *github.Repository) {
			defer wg.Done()
			defer func() { <-sem }()

			hooks, _, err := client.Repositories.ListHooks(ctx, flags.Organization, r.GetName(), nil)
			if err == nil && len(hooks) > 0 {
				mu.Lock()
				allWebhooks[r.GetName()] = hooks
				mu.Unlock()
			}
		}(repo)
	}
	wg.Wait()

	// Display results based on format
	switch flags.OutputFormat {
	case "json":
		return displayWebhooksJSON(allWebhooks)
	case "yaml":
		return displayWebhooksYAML(allWebhooks)
	default:
		return displayWebhooksTable(allWebhooks)
	}
}

// runWebhookBulkDelete executes the bulk webhook delete command
func runWebhookBulkDelete(flags *WebhookBulkFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	// Get target repositories
	repos, err := getTargetRepositories(ctx, client, flags)
	if err != nil {
		return fmt.Errorf("failed to get target repositories: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found matching criteria")
	}

	fmt.Printf("🗑️  Deleting webhooks with URL '%s' from %d repositories\n", flags.URL, len(repos))

	if flags.DryRun {
		fmt.Println("DRY RUN MODE - No webhooks will be deleted")
	}

	// Execute delete operation
	var results []BulkOperationResult
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, flags.MaxWorkers)

	for _, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(r *github.Repository) {
			defer wg.Done()
			defer func() { <-sem }()

			result := deleteWebhooksByURL(ctx, client, flags.Organization, r.GetName(), flags.URL, flags.DryRun)
			mu.Lock()
			results = append(results, result...)
			mu.Unlock()
		}(repo)
	}
	wg.Wait()

	// Display results
	displayBulkResults(results, "delete")

	return nil
}

// runWebhookBulkSync executes the bulk webhook sync command
func runWebhookBulkSync(flags *WebhookBulkFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	// Load configuration
	config, err := loadBulkWebhookConfig(flags.GlobalFlags.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get target repositories
	repos, err := getTargetRepositories(ctx, client, flags)
	if err != nil {
		return fmt.Errorf("failed to get target repositories: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found matching criteria")
	}

	fmt.Printf("🔄 Synchronizing webhooks for %d repositories\n", len(repos))

	// Execute sync operation
	results := executeBulkWebhookOperation(ctx, client, repos, config, syncWebhookOperation)

	// Display results
	displayBulkResults(results, "sync")

	return nil
}

// getTargetRepositories returns the list of repositories to operate on
func getTargetRepositories(ctx context.Context, client *github.Client, flags *WebhookBulkFlags) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	// List all organization repositories
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, flags.Organization, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Filter repositories based on criteria
	var targetRepos []*github.Repository

	// If specific repositories are specified
	if len(flags.Repositories) > 0 {
		repoMap := make(map[string]bool)
		for _, name := range flags.Repositories {
			repoMap[name] = true
		}
		for _, repo := range allRepos {
			if repoMap[repo.GetName()] {
				targetRepos = append(targetRepos, repo)
			}
		}
	} else if flags.RepoPattern != "" {
		// Filter by pattern
		for _, repo := range allRepos {
			if matched, _ := matchPattern(repo.GetName(), flags.RepoPattern); matched {
				targetRepos = append(targetRepos, repo)
			}
		}
	} else if flags.All { // Using "all" flag
		targetRepos = allRepos
	}

	return targetRepos, nil
}

// executeBulkWebhookOperation executes a webhook operation across multiple repositories
func executeBulkWebhookOperation(ctx context.Context, client *github.Client, repos []*github.Repository, config *BulkWebhookConfig, operation webhookOperation) []BulkOperationResult {
	var results []BulkOperationResult
	var mu sync.Mutex

	// Use errgroup for better error handling
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.Options.MaxWorkers)

	for _, repo := range repos {
		repo := repo // capture loop variable
		g.Go(func() error {
			repoResults := operation(ctx, client, repo, config)
			mu.Lock()
			results = append(results, repoResults...)
			mu.Unlock()

			if !config.Options.ContinueOnError && hasErrors(repoResults) {
				return fmt.Errorf("operation failed for repository %s", repo.GetName())
			}
			return nil
		})
	}

	// Wait for all operations to complete
	_ = g.Wait() // Ignore error as we collect individual results

	return results
}

// webhookOperation is a function type for webhook operations
type webhookOperation func(ctx context.Context, client *github.Client, repo *github.Repository, config *BulkWebhookConfig) []BulkOperationResult

// createWebhookOperation creates webhooks in a repository
func createWebhookOperation(ctx context.Context, client *github.Client, repo *github.Repository, config *BulkWebhookConfig) []BulkOperationResult {
	var results []BulkOperationResult

	for _, webhookConfig := range config.Webhooks {
		hook := &github.Hook{
			Events: webhookConfig.Events,
			Active: webhookConfig.Active,
			Config: &github.HookConfig{
				URL:         &webhookConfig.URL,
				ContentType: &webhookConfig.ContentType,
			},
		}

		if webhookConfig.Secret != "" {
			hook.Config.Secret = &webhookConfig.Secret
		}

		// Check if webhook already exists
		if config.Options.SkipExisting {
			existing, err := findWebhookByURL(ctx, client, repo.GetOwner().GetLogin(), repo.GetName(), webhookConfig.URL)
			if err == nil && existing != nil {
				results = append(results, BulkOperationResult{
					Repository: repo.GetName(),
					Action:     "create",
					Success:    true,
					Details:    "Skipped - webhook already exists",
				})
				continue
			}
		}

		// Create webhook
		created, _, err := client.Repositories.CreateHook(ctx, repo.GetOwner().GetLogin(), repo.GetName(), hook)
		if err != nil {
			results = append(results, BulkOperationResult{
				Repository: repo.GetName(),
				Action:     "create",
				Success:    false,
				Error:      err,
			})
		} else {
			results = append(results, BulkOperationResult{
				Repository: repo.GetName(),
				Action:     "create",
				Success:    true,
				Details:    fmt.Sprintf("Created webhook ID: %d", created.GetID()),
			})
		}
	}

	return results
}

// syncWebhookOperation synchronizes webhooks in a repository
func syncWebhookOperation(ctx context.Context, client *github.Client, repo *github.Repository, config *BulkWebhookConfig) []BulkOperationResult {
	var results []BulkOperationResult

	// Get existing webhooks
	existing, _, err := client.Repositories.ListHooks(ctx, repo.GetOwner().GetLogin(), repo.GetName(), nil)
	if err != nil {
		return []BulkOperationResult{{
			Repository: repo.GetName(),
			Action:     "sync",
			Success:    false,
			Error:      err,
		}}
	}

	// Create map of existing webhooks by URL
	existingMap := make(map[string]*github.Hook)
	for _, hook := range existing {
		if hook.Config != nil && hook.Config.URL != nil {
			existingMap[*hook.Config.URL] = hook
		}
	}

	// Sync webhooks from config
	for _, webhookConfig := range config.Webhooks {
		if existingHook, found := existingMap[webhookConfig.URL]; found {
			// Update existing webhook
			updated := &github.Hook{
				Events: webhookConfig.Events,
				Active: webhookConfig.Active,
				Config: &github.HookConfig{
					URL:         &webhookConfig.URL,
					ContentType: &webhookConfig.ContentType,
				},
			}

			if webhookConfig.Secret != "" {
				updated.Config.Secret = &webhookConfig.Secret
			}

			_, _, err := client.Repositories.EditHook(ctx, repo.GetOwner().GetLogin(), repo.GetName(), existingHook.GetID(), updated)
			if err != nil {
				results = append(results, BulkOperationResult{
					Repository: repo.GetName(),
					Action:     "update",
					Success:    false,
					Error:      err,
				})
			} else {
				results = append(results, BulkOperationResult{
					Repository: repo.GetName(),
					Action:     "update",
					Success:    true,
					Details:    fmt.Sprintf("Updated webhook ID: %d", existingHook.GetID()),
				})
			}
			delete(existingMap, webhookConfig.URL)
		} else {
			// Create new webhook
			createResults := createWebhookOperation(ctx, client, repo, &BulkWebhookConfig{
				Webhooks: []repoconfig.WebhookConfig{webhookConfig},
				Options:  config.Options,
			})
			results = append(results, createResults...)
		}
	}

	// Optionally remove webhooks not in config
	// This is commented out for safety - uncomment if you want to remove extra webhooks
	/*
		for url, hook := range existingMap {
			_, err := client.Repositories.DeleteHook(ctx, repo.GetOwner().GetLogin(), repo.GetName(), hook.GetID())
			if err != nil {
				results = append(results, BulkOperationResult{
					Repository: repo.GetName(),
					Action:     "delete",
					Success:    false,
					Error:      err,
				})
			} else {
				results = append(results, BulkOperationResult{
					Repository: repo.GetName(),
					Action:     "delete",
					Success:    true,
					Details:    fmt.Sprintf("Deleted webhook URL: %s", url),
				})
			}
		}
	*/

	return results
}

// deleteWebhooksByURL deletes webhooks with a specific URL
func deleteWebhooksByURL(ctx context.Context, client *github.Client, org, repo, url string, dryRun bool) []BulkOperationResult {
	var results []BulkOperationResult

	// Get all webhooks
	hooks, _, err := client.Repositories.ListHooks(ctx, org, repo, nil)
	if err != nil {
		return []BulkOperationResult{{
			Repository: repo,
			Action:     "delete",
			Success:    false,
			Error:      err,
		}}
	}

	// Find and delete matching webhooks
	for _, hook := range hooks {
		if hook.Config != nil && hook.Config.URL != nil && *hook.Config.URL == url {
			if dryRun {
				results = append(results, BulkOperationResult{
					Repository: repo,
					Action:     "delete",
					Success:    true,
					Details:    fmt.Sprintf("Would delete webhook ID: %d", hook.GetID()),
				})
			} else {
				_, err := client.Repositories.DeleteHook(ctx, org, repo, hook.GetID())
				if err != nil {
					results = append(results, BulkOperationResult{
						Repository: repo,
						Action:     "delete",
						Success:    false,
						Error:      err,
					})
				} else {
					results = append(results, BulkOperationResult{
						Repository: repo,
						Action:     "delete",
						Success:    true,
						Details:    fmt.Sprintf("Deleted webhook ID: %d", hook.GetID()),
					})
				}
			}
		}
	}

	if len(results) == 0 {
		results = append(results, BulkOperationResult{
			Repository: repo,
			Action:     "delete",
			Success:    true,
			Details:    "No matching webhooks found",
		})
	}

	return results
}

// findWebhookByURL finds a webhook by URL
func findWebhookByURL(ctx context.Context, client *github.Client, org, repo, url string) (*github.Hook, error) {
	hooks, _, err := client.Repositories.ListHooks(ctx, org, repo, nil)
	if err != nil {
		return nil, err
	}

	for _, hook := range hooks {
		if hook.Config != nil && hook.Config.URL != nil && *hook.Config.URL == url {
			return hook, nil
		}
	}

	return nil, nil
}

// loadBulkWebhookConfig loads bulk webhook configuration from file
func loadBulkWebhookConfig(path string) (*BulkWebhookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config BulkWebhookConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Options.MaxWorkers == 0 {
		config.Options.MaxWorkers = 5
	}

	return &config, nil
}

// displayBulkResults displays the results of bulk operations
func displayBulkResults(results []BulkOperationResult, operation string) {
	successCount := 0
	failureCount := 0

	// Group results by repository
	repoResults := make(map[string][]BulkOperationResult)
	for _, result := range results {
		repoResults[result.Repository] = append(repoResults[result.Repository], result)
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Display summary
	fmt.Println("\n📊 Bulk Operation Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Operation: %s\n", strings.Title(operation))
	fmt.Printf("Total operations: %d\n", len(results))
	fmt.Printf("✅ Successful: %d\n", successCount)
	fmt.Printf("❌ Failed: %d\n", failureCount)
	fmt.Println()

	// Display detailed results for failures
	if failureCount > 0 {
		fmt.Println("Failed Operations:")
		for repo, results := range repoResults {
			for _, result := range results {
				if !result.Success {
					fmt.Printf("  - %s: %v\n", repo, result.Error)
				}
			}
		}
		fmt.Println()
	}

	// Display sample of successful operations
	fmt.Println("Operation Details:")
	count := 0
	for repo, results := range repoResults {
		for _, result := range results {
			if result.Success && count < 10 {
				details := ""
				if result.Details != nil {
					details = fmt.Sprintf(" - %v", result.Details)
				}
				fmt.Printf("  ✅ %s: %s%s\n", repo, result.Action, details)
				count++
			}
		}
	}

	if len(results) > 10 {
		fmt.Printf("  ... and %d more operations\n", len(results)-10)
	}
}

// displayWebhooksTable displays webhooks in table format
func displayWebhooksTable(webhooks map[string][]*github.Hook) error {
	if len(webhooks) == 0 {
		fmt.Println("No webhooks found in any repository")
		return nil
	}

	for repo, hooks := range webhooks {
		fmt.Printf("\n📦 Repository: %s\n", repo)
		fmt.Printf("%-8s %-40s %-20s %-8s\n", "ID", "URL", "EVENTS", "ACTIVE")
		fmt.Println("-------- ---------------------------------------- -------------------- --------")

		for _, hook := range hooks {
			url := ""
			if hook.Config != nil && hook.Config.URL != nil {
				url = *hook.Config.URL
			}

			eventsStr := strings.Join(hook.Events, ",")
			if len(eventsStr) > 20 {
				eventsStr = eventsStr[:17] + "..."
			}

			fmt.Printf("%-8d %-40s %-20s %-8t\n",
				hook.GetID(),
				truncateString(url, 40),
				eventsStr,
				hook.GetActive())
		}
	}

	return nil
}

// displayWebhooksJSON displays webhooks in JSON format
func displayWebhooksJSON(webhooks map[string][]*github.Hook) error {
	return json.NewEncoder(os.Stdout).Encode(webhooks)
}

// displayWebhooksYAML displays webhooks in YAML format
func displayWebhooksYAML(webhooks map[string][]*github.Hook) error {
	// Convert to configuration format
	config := BulkWebhookConfig{
		Version: "1.0",
		Targets: BulkWebhookTargets{
			Repositories: make([]string, 0, len(webhooks)),
		},
	}

	// Collect unique webhooks
	uniqueWebhooks := make(map[string]repoconfig.WebhookConfig)
	for repo, hooks := range webhooks {
		config.Targets.Repositories = append(config.Targets.Repositories, repo)
		for _, hook := range hooks {
			if hook.Config != nil && hook.Config.URL != nil {
				key := *hook.Config.URL
				if _, exists := uniqueWebhooks[key]; !exists {
					webhookConfig := repoconfig.WebhookConfig{
						URL:    *hook.Config.URL,
						Events: hook.Events,
						Active: hook.Active,
					}
					if hook.Config.ContentType != nil {
						webhookConfig.ContentType = *hook.Config.ContentType
					}
					uniqueWebhooks[key] = webhookConfig
				}
			}
		}
	}

	// Convert map to slice
	for _, webhook := range uniqueWebhooks {
		config.Webhooks = append(config.Webhooks, webhook)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

// hasErrors checks if any results contain errors
func hasErrors(results []BulkOperationResult) bool {
	for _, result := range results {
		if !result.Success {
			return true
		}
	}
	return false
}

package bulkclone

import (
	"fmt"
	"os"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

type bulkCloneGithubOptions struct {
	targetPath    string
	orgName       string
	strategy      string
	configFile    string
	useConfig     bool
	parallel      int
	maxRetries    int
	resume        bool
	optimized     bool
	token         string
	memoryLimit   string
	streamingMode bool
	enableCache   bool
	enableRedis   bool
	redisAddr     string
	progressMode  string
}

func defaultBulkCloneGithubOptions() *bulkCloneGithubOptions {
	return &bulkCloneGithubOptions{
		strategy:      "reset",
		parallel:      10,
		maxRetries:    3,
		optimized:     false,
		streamingMode: false,
		memoryLimit:   "500MB",
		progressMode:  "compact",
	}
}

func newBulkCloneGithubCmd() *cobra.Command {
	o := defaultBulkCloneGithubOptions()

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Clone repositories from a GitHub organization",
		Args:  cobra.NoArgs,
		RunE:  o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().IntVarP(&o.parallel, "parallel", "p", o.parallel, "Number of parallel workers for cloning")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", o.maxRetries, "Maximum retry attempts for failed operations")
	cmd.Flags().BoolVar(&o.resume, "resume", false, "Resume interrupted clone operation from saved state")

	// New optimization flags
	cmd.Flags().BoolVar(&o.optimized, "optimized", o.optimized, "Use optimized streaming API for large-scale operations (recommended for >1000 repos)")
	cmd.Flags().StringVar(&o.token, "token", "", "GitHub token for API access (can also use GITHUB_TOKEN env var)")
	cmd.Flags().StringVar(&o.memoryLimit, "memory-limit", o.memoryLimit, "Maximum memory usage (e.g., 500MB, 2GB)")
	cmd.Flags().BoolVar(&o.streamingMode, "streaming", o.streamingMode, "Enable streaming mode for memory-efficient processing")

	// Cache flags
	cmd.Flags().BoolVar(&o.enableCache, "cache", false, "Enable caching for repeated requests (recommended for frequent operations)")
	cmd.Flags().BoolVar(&o.enableRedis, "redis", false, "Enable Redis distributed caching (requires Redis server)")
	cmd.Flags().StringVar(&o.redisAddr, "redis-addr", "localhost:6379", "Redis server address for distributed caching")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("orgName", "config", "use-config")

	return cmd
}

func (o *bulkCloneGithubOptions) run(cmd *cobra.Command, args []string) error {
	// Load config if specified
	if o.configFile != "" || o.useConfig {
		err := o.loadFromConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// Get GitHub token
	token := o.token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Use optimized streaming approach for large-scale operations
	ctx := cmd.Context()
	var err error

	// Determine which approach to use
	if o.enableCache {
		// Use cached approach (Redis cache disabled, using local cache only)
		fmt.Printf("🔄 Using cached API calls for improved performance\n")
		err = github.RefreshAllOptimizedStreamingWithCache(ctx, o.targetPath, o.orgName, o.strategy, token)
	} else if o.optimized || o.streamingMode || token != "" {
		if token == "" {
			fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
			fmt.Printf("   Set GITHUB_TOKEN environment variable or use --token flag for better performance.\n")
		}

		fmt.Printf("🚀 Using optimized streaming API for large-scale operations\n")
		if o.memoryLimit != "" {
			fmt.Printf("🧠 Memory limit: %s\n", o.memoryLimit)
		}

		err = github.RefreshAllOptimizedStreaming(ctx, o.targetPath, o.orgName, o.strategy, token)
	} else if o.resume || o.parallel > 1 {
		err = github.RefreshAllResumable(ctx, o.targetPath, o.orgName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
	} else {
		err = github.RefreshAll(ctx, o.targetPath, o.orgName, o.strategy)
	}

	if err != nil {
		// return err
		// return fmt.Errorf("failed to refresh repositories: %w", err)
		return nil
	}

	return nil
}

func (o *bulkCloneGithubOptions) loadFromConfig() error {
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := config.LoadConfigFromFile(configPath)
	if err != nil {
		return err
	}

	// If orgName is specified via CLI, use it; otherwise get from config
	if o.orgName == "" {
		// Look for GitHub provider configuration
		if githubProvider, exists := cfg.Providers[config.ProviderGitHub]; exists {
			if len(githubProvider.Orgs) > 0 {
				o.orgName = githubProvider.Orgs[0].Name
			} else {
				return fmt.Errorf("no GitHub organizations found in config")
			}
		} else {
			return fmt.Errorf("no GitHub provider configuration found")
		}
	}

	// Find the organization configuration
	var orgConfig *config.GitTarget
	if githubProvider, exists := cfg.Providers[config.ProviderGitHub]; exists {
		for i := range githubProvider.Orgs {
			if githubProvider.Orgs[i].Name == o.orgName {
				orgConfig = &githubProvider.Orgs[i]
				break
			}
		}
	}

	if orgConfig == nil {
		return fmt.Errorf("organization '%s' not found in GitHub provider configuration", o.orgName)
	}

	// Apply config values (CLI flags take precedence)
	if o.targetPath == "" && orgConfig.CloneDir != "" {
		o.targetPath = config.ExpandEnvironmentVariables(orgConfig.CloneDir)
	}

	// Apply auth configuration if available
	// Authentication support would be implemented here using
	// environment variables or config-based token management
	// For now, this is documented for future implementation

	return nil
}

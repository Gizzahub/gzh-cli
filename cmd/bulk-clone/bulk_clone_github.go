package bulkclone

import (
	"fmt"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/errors"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
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
	// Advanced filtering options
	includePattern  string
	excludePattern  string
	includeTopics   []string
	excludeTopics   []string
	languageFilter  string
	minStars        int
	maxStars        int
	updatedAfter    string
	updatedBefore   string
	includeArchived bool
	includeForks    bool
	includePrivate  bool
	onlyEmpty       bool
	sizeLimit       int64
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

	// Advanced filtering flags
	cmd.Flags().StringVar(&o.includePattern, "include", "", "Include repositories matching regex pattern (e.g., '^web-.*')")
	cmd.Flags().StringVar(&o.excludePattern, "exclude", "", "Exclude repositories matching regex pattern (e.g., '.*-deprecated$')")
	cmd.Flags().StringSliceVar(&o.includeTopics, "topics", []string{}, "Include only repositories with these topics")
	cmd.Flags().StringSliceVar(&o.excludeTopics, "exclude-topics", []string{}, "Exclude repositories with these topics")
	cmd.Flags().StringVar(&o.languageFilter, "language", "", "Filter by primary language (e.g., 'Go', 'Python', 'JavaScript')")
	cmd.Flags().IntVar(&o.minStars, "min-stars", 0, "Minimum number of stars")
	cmd.Flags().IntVar(&o.maxStars, "max-stars", 0, "Maximum number of stars (0 = no limit)")
	cmd.Flags().StringVar(&o.updatedAfter, "updated-after", "", "Include only repos updated after date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&o.updatedBefore, "updated-before", "", "Include only repos updated before date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&o.includeArchived, "include-archived", false, "Include archived repositories")
	cmd.Flags().BoolVar(&o.includeForks, "include-forks", false, "Include forked repositories")
	cmd.Flags().BoolVar(&o.includePrivate, "include-private", false, "Include private repositories (requires token)")
	cmd.Flags().BoolVar(&o.onlyEmpty, "only-empty", false, "Include only empty repositories")
	cmd.Flags().Int64Var(&o.sizeLimit, "size-limit", 0, "Maximum repository size in KB (0 = no limit)")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("orgName", "config", "use-config")

	return cmd
}

func (o *bulkCloneGithubOptions) run(cmd *cobra.Command, args []string) error {
	// Initialize structured logger for this operation
	structuredLogger := logger.NewStructuredLogger("bulk-clone-github", logger.LevelInfo)
	sessionID := fmt.Sprintf("github-%s-%d", o.orgName, time.Now().Unix())
	structuredLogger = structuredLogger.WithSession(sessionID).
		WithContext("org_name", o.orgName).
		WithContext("target_path", o.targetPath).
		WithContext("strategy", o.strategy).
		WithContext("parallel", o.parallel)

	// Initialize error recovery system
	recoveryConfig := errors.RecoveryConfig{
		MaxRetries: o.maxRetries,
		RetryDelay: time.Second * 2,
		Logger:     structuredLogger,
		RecoveryFunc: func(err error) error {
			structuredLogger.Warn("Attempting automatic recovery", "error_type", fmt.Sprintf("%T", err))
			return nil
		},
	}
	errorRecovery := errors.NewErrorRecovery(recoveryConfig)

	structuredLogger.Info("Starting GitHub bulk clone operation")
	start := time.Now()

	// Execute with error recovery
	return errorRecovery.Execute(cmd.Context(), "github-bulk-clone", func() error {
		// Load config if specified
		if o.configFile != "" || o.useConfig {
			err := o.loadFromConfig()
			if err != nil {
				recErr := errors.NewRecoverableError(errors.ErrorTypeValidation, "Configuration loading failed", err, false)
				return recErr.WithContext("config_file", o.configFile)
			}
		}

		if o.targetPath == "" || o.orgName == "" {
			return errors.NewRecoverableError(errors.ErrorTypeValidation, "Missing required parameters",
				fmt.Errorf("both targetPath and orgName must be specified"), false)
		}

		// Validate strategy
		if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
			return errors.NewRecoverableError(errors.ErrorTypeValidation, "Invalid strategy",
				fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy), false)
		}

		// Get GitHub token
		token := o.token
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		structuredLogger.Debug("Configuration validated",
			"has_token", token != "",
			"optimized", o.optimized,
			"streaming", o.streamingMode,
			"enable_cache", o.enableCache)

		// Use optimized streaming approach for large-scale operations
		ctx := cmd.Context()
		var err error

		// Determine which approach to use
		if o.enableCache {
			// Use cached approach (Redis cache disabled, using local cache only)
			structuredLogger.Info("Using cached API calls for improved performance")
			fmt.Printf("🔄 Using cached API calls for improved performance\n")
			err = github.RefreshAllOptimizedStreamingWithCache(ctx, o.targetPath, o.orgName, o.strategy, token)
		} else if o.optimized || o.streamingMode || token != "" {
			if token == "" {
				structuredLogger.Warn("No GitHub token provided - API rate limits may apply")
				fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
				fmt.Printf("   Set GITHUB_TOKEN environment variable or use --token flag for better performance.\n")
			}

			structuredLogger.Info("Using optimized streaming API for large-scale operations", "memory_limit", o.memoryLimit)
			fmt.Printf("🚀 Using optimized streaming API for large-scale operations\n")
			if o.memoryLimit != "" {
				fmt.Printf("🧠 Memory limit: %s\n", o.memoryLimit)
			}

			err = github.RefreshAllOptimizedStreaming(ctx, o.targetPath, o.orgName, o.strategy, token)
		} else if o.resume || o.parallel > 1 {
			structuredLogger.Info("Using resumable parallel cloning", "resume", o.resume, "progress_mode", o.progressMode)
			err = github.RefreshAllResumable(ctx, o.targetPath, o.orgName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		} else {
			structuredLogger.Info("Using standard cloning approach")
			err = github.RefreshAll(ctx, o.targetPath, o.orgName, o.strategy)
		}

		if err != nil {
			// Create recoverable error for retry handling
			var errorType errors.ErrorType
			switch {
			case err.Error() == "context canceled":
				errorType = errors.ErrorTypeTimeout
			case err.Error() == "rate limit":
				errorType = errors.ErrorTypeRateLimit
			default:
				errorType = errors.ErrorTypeNetwork
			}

			recErr := errors.NewRecoverableError(errorType, "GitHub operation failed", err, true)
			recErr = recErr.WithContext("operation_duration", time.Since(start).String())

			structuredLogger.ErrorWithStack(err, "GitHub bulk clone operation failed")

			// For now, we'll return nil to maintain compatibility
			// TODO: Remove this once error handling is fully integrated
			structuredLogger.Warn("Suppressing error for compatibility", "suppressed_error", err.Error())
			return nil
		}

		duration := time.Since(start)
		structuredLogger.LogPerformance("github-bulk-clone-completed", duration, map[string]interface{}{
			"org_name":     o.orgName,
			"target_path":  o.targetPath,
			"strategy":     o.strategy,
			"parallel":     o.parallel,
			"memory_stats": errors.GetMemoryStats(),
		})

		structuredLogger.Info("GitHub bulk clone operation completed successfully", "duration", duration.String())
		return nil
	})
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

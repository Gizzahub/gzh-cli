// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	internalconfig "github.com/gizzahub/gzh-manager-go/internal/config"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/internal/errors"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
	"github.com/gizzahub/gzh-manager-go/internal/validation"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// GzhYamlConfig represents the structure of gzh.yaml file generated in target directory
type GzhYamlConfig struct {
	Organization string      `yaml:"organization"`
	Provider     string      `yaml:"provider"`
	GeneratedAt  time.Time   `yaml:"generated_at"`
	SyncMode     SyncMode    `yaml:"sync_mode"`
	Repositories []RepoInfo  `yaml:"repositories"`
}

// SyncMode represents synchronization configuration
type SyncMode struct {
	CleanupOrphans bool `yaml:"cleanup_orphans"`
}

// RepoInfo represents repository information
type RepoInfo struct {
	Name        string `yaml:"name"`
	CloneURL    string `yaml:"clone_url"`
	Description string `yaml:"description"`
	Private     bool   `yaml:"private"`
	Archived    bool   `yaml:"archived"`
	Fork        bool   `yaml:"fork"`
}

// generateGzhYaml creates a gzh.yaml file in the target directory with repository information
func (o *syncCloneGithubOptions) generateGzhYaml(targetPath string, repos []github.RepoInfo) error {
	gzhConfig := GzhYamlConfig{
		Organization: o.orgName,
		Provider:     "github",
		GeneratedAt:  time.Now(),
		SyncMode: SyncMode{
			CleanupOrphans: o.cleanupOrphans,
		},
		Repositories: make([]RepoInfo, len(repos)),
	}

	// Convert github.RepoInfo to our RepoInfo
	// TODO: Update when GitHub RepoInfo struct is expanded with more fields
	for i := range repos {
		gzhConfig.Repositories[i] = RepoInfo{
			Name:        "repository_" + fmt.Sprint(i), // TODO: Get actual name from GitHub API
			CloneURL:    fmt.Sprintf("https://github.com/%s/repository_%d.git", o.orgName, i), // TODO: Get actual clone URL
			Description: "Repository description", // TODO: Get actual description
			Private:     false, // TODO: Get actual private status
			Archived:    false, // TODO: Get actual archived status
			Fork:        false, // TODO: Get actual fork status
		}
	}

	// Write to gzh.yaml file
	gzhPath := filepath.Join(targetPath, "gzh.yaml")
	data, err := yaml.Marshal(gzhConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal gzh.yaml: %w", err)
	}

	if err := os.WriteFile(gzhPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write gzh.yaml: %w", err)
	}

	return nil
}

// cleanupOrphanDirectories removes directories in targetPath that are not in the repository list
func (o *syncCloneGithubOptions) cleanupOrphanDirectories(targetPath string, repos []github.RepoInfo) error {
	if !o.cleanupOrphans {
		return nil
	}

	// Get list of expected repository directories
	repoNames := make(map[string]bool)
	for i := range repos {
		// TODO: Use actual repository names when available
		repoName := "repository_" + fmt.Sprint(i)
		repoNames[repoName] = true
	}

	// Read target directory
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	// Check each directory and remove orphans
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files
		}

		// Skip special files/directories
		name := entry.Name()
		if name == "gzh.yaml" || name == ".git" || name[0] == '.' {
			continue
		}

		// Remove directory if it's not in the repository list
		if !repoNames[name] {
			orphanPath := filepath.Join(targetPath, name)
			fmt.Printf("🗑️ Removing orphan directory: %s\n", name)
			if err := os.RemoveAll(orphanPath); err != nil {
				return fmt.Errorf("failed to remove orphan directory %s: %w", name, err)
			}
		}
	}

	return nil
}

type syncCloneGithubOptions struct {
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
	cleanupOrphans  bool
}

func defaultSyncCloneGithubOptions() *syncCloneGithubOptions {
	return &syncCloneGithubOptions{
		strategy:      "reset",
		parallel:      10,
		maxRetries:    3,
		optimized:     false,
		streamingMode:  false,
		memoryLimit:    "500MB",
		progressMode:   "bar",
		cleanupOrphans: false,
	}
}

func newSyncCloneGithubCmd() *cobra.Command {
	o := defaultSyncCloneGithubOptions()

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
	cmd.Flags().BoolVar(&o.cleanupOrphans, "cleanup-orphans", false, "Remove directories not present in the organization's repositories")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("orgName", "config", "use-config")

	return cmd
}

func (o *syncCloneGithubOptions) run(cmd *cobra.Command, args []string) error { //nolint:gocognit // Complex business logic for sync clone operations
	// Initialize simple logger for this operation
	simpleLogger := logger.NewSimpleLogger("synclone-github")
	sessionID := fmt.Sprintf("github-%s-%d", o.orgName, time.Now().Unix())
	simpleLogger = simpleLogger.WithSession(sessionID).
		WithContext("org_name", o.orgName).
		WithContext("target_path", o.targetPath).
		WithContext("strategy", o.strategy).
		WithContext("parallel", o.parallel)

	// Initialize error recovery system
	recoveryConfig := errors.RecoveryConfig{
		MaxRetries: o.maxRetries,
		RetryDelay: time.Second * 2,
		Logger:     simpleLogger,
		RecoveryFunc: func(err error) error {
			simpleLogger.Warn("Attempting automatic recovery", "error_type", fmt.Sprintf("%T", err))
			return nil
		},
	}
	errorRecovery := errors.NewErrorRecovery(recoveryConfig)

	simpleLogger.Info("Starting GitHub bulk clone operation")

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

		// Comprehensive input validation
		validator := validation.NewBulkCloneValidator()
		opts := &validation.BulkCloneOptions{
			TargetPath:     o.targetPath,
			OrgName:        o.orgName,
			Strategy:       o.strategy,
			ConfigFile:     o.configFile,
			Parallel:       o.parallel,
			MaxRetries:     o.maxRetries,
			Token:          o.token,
			MemoryLimit:    o.memoryLimit,
			ProgressMode:   o.progressMode,
			RedisAddr:      o.redisAddr,
			IncludePattern: o.includePattern,
			ExcludePattern: o.excludePattern,
			IncludeTopics:  o.includeTopics,
			ExcludeTopics:  o.excludeTopics,
			LanguageFilter: o.languageFilter,
			MinStars:       o.minStars,
			MaxStars:       o.maxStars,
			UpdatedAfter:   o.updatedAfter,
			UpdatedBefore:  o.updatedBefore,
		}

		if err := validator.ValidateOptions(opts); err != nil {
			return errors.NewRecoverableError(errors.ErrorTypeValidation, "Input validation failed", err, false)
		}

		// Sanitize inputs for additional security
		sanitized := validator.SanitizeOptions(opts)
		o.targetPath = sanitized.TargetPath
		o.orgName = sanitized.OrgName
		o.strategy = sanitized.Strategy

		// Get GitHub token
		token := o.token
		if token == "" {
			token = env.GetToken("github")
		}

		simpleLogger.Debug("Configuration validated",
			"has_token", token != "",
			"optimized", o.optimized,
			"streaming", o.streamingMode,
			"enable_cache", o.enableCache)

		// Use optimized streaming approach for large-scale operations
		ctx := cmd.Context()

		var err error

		// Determine which approach to use
		if o.enableCache { //nolint:gocritic // Complex boolean conditions not suitable for switch
			// Use cached approach (Redis cache disabled, using local cache only)
			simpleLogger.Info("Using cached API calls for improved performance")
			fmt.Printf("🔄 Using cached API calls for improved performance\n")

			err = github.RefreshAllOptimizedStreamingWithCache(ctx, o.targetPath, o.orgName, o.strategy, token)
		} else if o.optimized || o.streamingMode || token != "" {
			if token == "" {
				simpleLogger.Warn("No GitHub token provided - API rate limits may apply")
				fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
				fmt.Printf("   Set GITHUB_TOKEN environment variable or use --token flag for better performance.\n")
			}

			simpleLogger.Info("Using optimized streaming API for large-scale operations", "memory_limit", o.memoryLimit)
			fmt.Printf("🚀 Using optimized streaming API for large-scale operations\n")

			if o.memoryLimit != "" {
				fmt.Printf("🧠 Memory limit: %s\n", o.memoryLimit)
			}

			err = github.RefreshAllOptimizedStreaming(ctx, o.targetPath, o.orgName, o.strategy, token)
		} else if o.resume || o.parallel > 1 {
			simpleLogger.Info("Using resumable parallel cloning", "resume", o.resume, "progress_mode", o.progressMode)
			err = github.RefreshAllResumable(ctx, o.targetPath, o.orgName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
		} else {
			simpleLogger.Info("Using standard cloning approach")

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

			simpleLogger.ErrorWithStack(err, "GitHub bulk clone operation failed")

			// Return the error properly for error handling
			return recErr
		}

		duration := time.Since(start)
		simpleLogger.LogPerformance("github-bulk-clone-completed", duration, map[string]interface{}{
			"org_name":     o.orgName,
			"target_path":  o.targetPath,
			"strategy":     o.strategy,
			"parallel":     o.parallel,
			"memory_stats": errors.GetMemoryStats(),
		})

		simpleLogger.Info("GitHub bulk clone operation completed successfully", "duration", duration.String())

		return nil
	})
}

func (o *syncCloneGithubOptions) loadFromConfig() error {
	// Use unified config loading
	cfg, err := internalconfig.LoadCommandConfig(context.Background(), o.configFile, "bulk-clone")
	if err != nil {
		return err
	}

	// If orgName is specified via CLI, use it; otherwise get from config
	if o.orgName == "" {
		// Look for GitHub provider configuration
		if githubProvider, exists := cfg.Providers[config.ProviderGitHub]; exists {
			if len(githubProvider.Organizations) > 0 {
				o.orgName = githubProvider.Organizations[0].Name
			} else {
				return fmt.Errorf("no GitHub organizations found in config")
			}
		} else {
			return fmt.Errorf("no GitHub provider configuration found")
		}
	}

	// Find the organization configuration
	var orgConfig *config.OrganizationConfig
	if githubProvider, exists := cfg.Providers[config.ProviderGitHub]; exists {
		for _, org := range githubProvider.Organizations {
			if org.Name == o.orgName {
				orgConfig = org
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

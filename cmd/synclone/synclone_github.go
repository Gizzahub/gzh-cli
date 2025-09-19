// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Gizzahub/gzh-cli/internal/app"
	internalconfig "github.com/Gizzahub/gzh-cli/internal/config"
	"github.com/Gizzahub/gzh-cli/internal/env"
	"github.com/Gizzahub/gzh-cli/internal/errors"
	"github.com/Gizzahub/gzh-cli/internal/validation"
	"github.com/Gizzahub/gzh-cli/pkg/config"
	"github.com/Gizzahub/gzh-cli/pkg/github"
)

// GzhYamlConfig represents the structure of gzh.yaml file generated in target directory
type GzhYamlConfig struct {
	Organization string            `yaml:"organization"`
	Provider     string            `yaml:"provider"`
	GeneratedAt  time.Time         `yaml:"generated_at"`
	SyncMode     SyncMode          `yaml:"sync_mode"`
	Repositories []github.RepoInfo `yaml:"repositories"`
}

// SyncMode represents synchronization configuration
type SyncMode struct {
	CleanupOrphans bool `yaml:"cleanup_orphans"`
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
		Repositories: make([]github.RepoInfo, len(repos)),
	}

	// Convert github.RepoInfo to our RepoInfo using actual data
	for i, repo := range repos {
		gzhConfig.Repositories[i] = github.RepoInfo{
			Name:        repo.Name,
			CloneURL:    repo.CloneURL,
			Description: repo.Description,
			Private:     repo.Private,
			Archived:    repo.Archived,
			Fork:        repo.Fork,
		}
	}

	// Write to gzh.yaml file
	gzhPath := filepath.Join(targetPath, "gzh.yaml")
	data, err := yaml.Marshal(gzhConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal gzh.yaml: %w", err)
	}

	if err := os.WriteFile(gzhPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write gzh.yaml: %w", err)
	}

	fmt.Printf("📝 Generated gzh.yaml with %d repositories\n", len(repos))
	return nil
}

// cleanupOrphanDirectories removes directories in targetPath that are not in the repository list
func (o *syncCloneGithubOptions) cleanupOrphanDirectories(targetPath string, repos []github.RepoInfo) error {
	if !o.cleanupOrphans {
		return nil
	}

	// Get list of expected repository directories using actual repository names
	repoNames := make(map[string]bool)
	for _, repo := range repos {
		repoNames[repo.Name] = true
	}

	// Read target directory
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	var orphansRemoved int
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
			orphansRemoved++
		}
	}

	if orphansRemoved > 0 {
		fmt.Printf("✅ Removed %d orphan directories\n", orphansRemoved)
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
		strategy:       "reset",
		parallel:       10,
		maxRetries:     3,
		optimized:      false,
		streamingMode:  false,
		memoryLimit:    "500MB",
		progressMode:   "bar",
		cleanupOrphans: false,
	}
}

func newSyncCloneGithubCmd(appCtx *app.AppContext) *cobra.Command {
	o := defaultSyncCloneGithubOptions()

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Clone repositories from a GitHub organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd, args, appCtx)
		},
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "Target directory (defaults to ./org_name if not specified)")
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

	// Aliases for simpler flags
	cmd.Flags().StringVar(&o.targetPath, "target", o.targetPath, "Target directory; defaults to current directory + org name (e.g., ./ScriptonBasestar) if not set")
	cmd.Flags().StringVar(&o.orgName, "org", o.orgName, "GitHub organization name")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")

	// Custom validation to handle both --org and --orgName aliases
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Check if any of the required flags are provided
		hasOrg := cmd.Flags().Changed("orgName") || cmd.Flags().Changed("org")
		hasConfig := cmd.Flags().Changed("config") || cmd.Flags().Changed("use-config")

		if !hasOrg && !hasConfig {
			return fmt.Errorf("at least one of the flags in the group [orgName org config use-config] is required")
		}
		return nil
	}

	return cmd
}

func (o *syncCloneGithubOptions) run(cmd *cobra.Command, args []string, appCtx *app.AppContext) error { //nolint:gocognit // Complex business logic for sync clone operations
	log := appCtx.Logger.WithSession(fmt.Sprintf("github-%s-%d", o.orgName, time.Now().Unix())).
		WithContext("org_name", o.orgName).
		WithContext("target_path", o.targetPath).
		WithContext("strategy", o.strategy).
		WithContext("parallel", o.parallel)

	log.Info("Starting GitHub synclone operation")

	start := time.Now()
	// 기본 targetPath가 비어있으면 현재 작업 디렉터리의 org_name 하위 디렉터리로 설정
	if o.targetPath == "" {
		if wd, err := os.Getwd(); err == nil && wd != "" {
			o.targetPath = filepath.Join(wd, o.orgName)
		} else {
			o.targetPath = filepath.Join(".", o.orgName)
		}
	}

	// Load config if specified
	if o.configFile != "" || o.useConfig {
		err := o.loadFromConfig()
		if err != nil {
			recErr := errors.NewRecoverableError(errors.ErrorTypeValidation, "Configuration loading failed", err, false)
			return recErr.WithContext("config_file", o.configFile)
		}
	}

	// Comprehensive input validation
	validator := validation.NewSyncCloneValidator()
	opts := &validation.SyncCloneOptions{
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

	log.Debug("Configuration validated",
		"has_token", token != "",
		"optimized", o.optimized,
		"streaming", o.streamingMode,
		"enable_cache", o.enableCache)

	// Use optimized streaming approach for large-scale operations
	ctx := cmd.Context()

	var err error

	// New synclone workflow: 1. Get repo list -> 2. Generate gzh.yaml -> 3. Cleanup orphans -> 4. Clone repos
	log.Info("Starting synclone workflow: fetching repository list from GitHub")

	// Check if gzh.yaml already exists
	gzhYamlPath := filepath.Join(o.targetPath, "gzh.yaml")
	var repos []github.RepoInfo
	existingYaml := false

	if _, err := os.Stat(gzhYamlPath); err == nil {
		// gzh.yaml exists, try to load it
		log.Info("Found existing gzh.yaml, loading repository list from file")
		fmt.Printf("📄 Found existing gzh.yaml, loading repository list...\n")

		data, err := os.ReadFile(gzhYamlPath)
		if err == nil {
			var gzhConfig GzhYamlConfig
			if err := yaml.Unmarshal(data, &gzhConfig); err == nil && gzhConfig.Organization == o.orgName {
				repos = gzhConfig.Repositories
				existingYaml = true
				fmt.Printf("✅ Loaded %d repositories from existing gzh.yaml\n", len(repos))
			}
		}
	}

	// If no existing yaml or failed to load, fetch from API
	if !existingYaml {
		fmt.Printf("🔍 Fetching repository list from GitHub organization: %s\n", o.orgName)

		// Step 1: Get repository list from GitHub API
		repos, err = github.ListRepos(ctx, o.orgName)
		if err != nil {
			// Check if it's a rate limit error
			if strings.Contains(err.Error(), "rate limit") || strings.Contains(err.Error(), "403") {
				// For rate limit errors, suppress usage display
				cmd.SilenceUsage = true
				recErr := errors.NewRecoverableError(errors.ErrorTypeRateLimit, "GitHub API rate limit exceeded", err, false)
				return recErr.WithContext("organization", o.orgName).WithContext("action", "list_repositories")
			}
			// Check for authentication errors
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
				// For auth errors, suppress usage display
				cmd.SilenceUsage = true
				recErr := errors.NewRecoverableError(errors.ErrorTypeAuth, "GitHub authentication failed", err, false)
				return recErr.WithContext("organization", o.orgName).WithContext("hint", "Check your GitHub token")
			}
			return fmt.Errorf("failed to fetch repository list: %w", err)
		}

		fmt.Printf("📋 Found %d repositories in organization %s\n", len(repos), o.orgName)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(o.targetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Step 2: Generate/Update gzh.yaml file with repository information
	if !existingYaml {
		if err := o.generateGzhYaml(o.targetPath, repos); err != nil {
			return fmt.Errorf("failed to generate gzh.yaml: %w", err)
		}
	}

	// Step 3: Verify existing clones if gzh.yaml existed
	if existingYaml {
		validClones, invalidClones := o.verifyExistingClones(repos)
		if len(validClones) > 0 {
			fmt.Printf("✅ Found %d valid existing clones\n", len(validClones))
		}
		if len(invalidClones) > 0 {
			fmt.Printf("⚠️  Found %d invalid/incomplete clones that will be re-cloned\n", len(invalidClones))
		}
	}

	// Step 4: Cleanup orphan directories if requested
	if err := o.cleanupOrphanDirectories(o.targetPath, repos); err != nil {
		return fmt.Errorf("failed to cleanup orphan directories: %w", err)
	}

	// Step 4: Clone/sync repositories using appropriate method
	if o.enableCache { //nolint:gocritic // Complex boolean conditions not suitable for switch
		// Use cached approach (Redis cache disabled, using local cache only)
		log.Info("Using cached API calls for improved performance")
		fmt.Printf("🔄 Using cached API calls for improved performance\n")

		err = github.RefreshAllOptimizedStreamingWithCache(ctx, o.targetPath, o.orgName, o.strategy, token)
	} else if o.optimized || o.streamingMode || token != "" {
		if token == "" {
			log.Warn("No GitHub token provided - API rate limits may apply")
			fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
			fmt.Printf("   Set GITHUB_TOKEN environment variable or use --token flag for better performance.\n")
		}

		log.Info("Using optimized streaming API for large-scale operations", "memory_limit", o.memoryLimit)
		fmt.Printf("🚀 Using optimized streaming API for large-scale operations\n")

		if o.memoryLimit != "" {
			fmt.Printf("🧠 Memory limit: %s\n", o.memoryLimit)
		}

		err = github.RefreshAllOptimizedStreaming(ctx, o.targetPath, o.orgName, o.strategy, token)
	} else if o.resume || o.parallel > 1 {
		log.Info("Using resumable parallel cloning", "resume", o.resume, "progress_mode", o.progressMode)
		err = github.RefreshAllResumable(ctx, o.targetPath, o.orgName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
	} else {
		log.Info("Using standard cloning approach")
		fmt.Printf("⚙️ Starting repository synchronization with strategy: %s\n", o.strategy)

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

		log.ErrorWithStack(err, "GitHub synclone operation failed")

		// Return the error properly for error handling
		return recErr
	}

	duration := time.Since(start)
	log.LogPerformance("github-synclone-completed", duration, map[string]interface{}{
		"org_name":     o.orgName,
		"target_path":  o.targetPath,
		"strategy":     o.strategy,
		"parallel":     o.parallel,
		"memory_stats": errors.GetMemoryStats(),
	})

	log.Info("GitHub synclone operation completed successfully", "duration", duration.String())

	return nil
}

func (o *syncCloneGithubOptions) loadFromConfig() error {
	// Use unified config loading
	cfg, err := internalconfig.LoadCommandConfig(context.Background(), o.configFile, "synclone")
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

// verifyExistingClones checks if existing clone directories are valid git repositories
func (o *syncCloneGithubOptions) verifyExistingClones(repos []github.RepoInfo) ([]string, []string) {
	validClones := []string{}
	invalidClones := []string{}

	for _, repo := range repos {
		repoPath := filepath.Join(o.targetPath, repo.Name)

		// Check if directory exists
		if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
			// Check if it's a valid git repository
			gitPath := filepath.Join(repoPath, ".git")
			if gitInfo, err := os.Stat(gitPath); err == nil && gitInfo.IsDir() {
				// Verify remote URL matches
				cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
				if output, err := cmd.Output(); err == nil {
					remoteURL := strings.TrimSpace(string(output))
					if remoteURL == repo.CloneURL || remoteURL == strings.Replace(repo.CloneURL, "https://", "git@", 1) {
						validClones = append(validClones, repo.Name)
						continue
					}
				}
			}
			// Directory exists but is not a valid git repo or has wrong remote
			invalidClones = append(invalidClones, repo.Name)
		}
	}

	return validClones, invalidClones
}

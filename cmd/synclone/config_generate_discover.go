// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-manager-go/internal/synclone/discovery"
)

// newConfigGenerateDiscoverCmd creates the config generate discover command.
func newConfigGenerateDiscoverCmd() *cobra.Command {
	var (
		basePath       string
		outputFile     string
		mergeExisting  bool
		recursive      bool
		maxDepth       int
		ignorePatterns []string
		followSymlinks bool
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover repositories and generate synclone configuration",
		Long: `Discover Git repositories in a directory tree and automatically generate
a synclone configuration file based on the discovered repositories.

This command scans the specified directory (and subdirectories if --recursive is enabled)
to find Git repositories, extracts their remote URLs, and generates a comprehensive
synclone configuration file.

Examples:
  # Discover repositories in current directory
  gz synclone config generate discover

  # Discover repositories in specific path
  gz synclone config generate discover --path ~/projects

  # Generate configuration with custom output file
  gz synclone config generate discover --path ~/repos --output my-config.yaml

  # Recursive discovery with custom depth
  gz synclone config generate discover --path ~/workspace --recursive --depth 3

  # Merge with existing configuration
  gz synclone config generate discover --path ~/repos --merge-existing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGenerateDiscover(basePath, outputFile, mergeExisting, recursive, maxDepth, ignorePatterns, followSymlinks)
		},
	}

	cmd.Flags().StringVar(&basePath, "path", ".", "Base path to scan for repositories")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone-discovered.yaml", "Output configuration file")
	cmd.Flags().BoolVar(&mergeExisting, "merge-existing", false, "Merge with existing configuration file")
	cmd.Flags().BoolVar(&recursive, "recursive", true, "Recursively scan subdirectories")
	cmd.Flags().IntVar(&maxDepth, "depth", 3, "Maximum directory depth for recursive scan")
	cmd.Flags().StringSliceVar(&ignorePatterns, "ignore", []string{".git", "node_modules", ".venv", "target", "build"}, "Patterns to ignore during discovery")
	cmd.Flags().BoolVar(&followSymlinks, "follow-symlinks", false, "Follow symbolic links during discovery")

	return cmd
}

// runConfigGenerateDiscover executes the discover command.
func runConfigGenerateDiscover(basePath, outputFile string, mergeExisting, recursive bool, maxDepth int, ignorePatterns []string, followSymlinks bool) error {
	fmt.Printf("🔍 Discovering repositories in %s...\n", basePath)

	// Create repository discoverer
	discoverer := discovery.NewRepoDiscoverer(basePath)
	discoverer.SetMaxDepth(maxDepth)
	discoverer.SetIgnorePatterns(ignorePatterns)
	discoverer.SetFollowSymlinks(followSymlinks)

	// Discover repositories
	repos, err := discoverer.DiscoverRepos()
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("❌ No Git repositories found in the specified path")
		return nil
	}

	fmt.Printf("✅ Found %d repositories\n", len(repos))

	// Group repositories by provider and organization
	groupedRepos := groupRepositoriesByProviderOrg(repos)

	// Generate configuration
	config := generateSyncloneConfig(groupedRepos, basePath)

	// Handle merging with existing configuration if requested
	if mergeExisting {
		if existingConfig, err := loadExistingConfig(outputFile); err == nil {
			config = mergeConfigurations(existingConfig, config)
			fmt.Printf("📄 Merged with existing configuration\n")
		}
	}

	// Save configuration to file
	if err := saveConfiguration(config, outputFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("📁 Configuration saved to %s\n", outputFile)

	// Display summary
	displayDiscoverySummary(groupedRepos, repos)

	return nil
}

// groupRepositoriesByProviderOrg groups repositories by provider and organization.
func groupRepositoriesByProviderOrg(repos []discovery.DiscoveredRepo) map[string]map[string][]discovery.DiscoveredRepo {
	grouped := make(map[string]map[string][]discovery.DiscoveredRepo)

	for _, repo := range repos {
		provider := repo.Provider
		if provider == "" {
			provider = "unknown"
		}

		org := repo.Org
		if org == "" {
			org = "personal"
		}

		if grouped[provider] == nil {
			grouped[provider] = make(map[string][]discovery.DiscoveredRepo)
		}

		grouped[provider][org] = append(grouped[provider][org], repo)
	}

	return grouped
}

// generateSyncloneConfig generates a synclone configuration from discovered repositories.
func generateSyncloneConfig(groupedRepos map[string]map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	config := map[string]interface{}{
		"version": "1.0.0",
		"global": map[string]interface{}{
			"clone_base_dir":   basePath,
			"default_strategy": "pull",
			"concurrency": map[string]interface{}{
				"clone_workers":  4,
				"update_workers": 8,
			},
		},
		"providers": make(map[string]interface{}),
		"sync_mode": map[string]interface{}{
			"cleanup_orphans":     false,
			"conflict_resolution": "local-keep",
		},
	}

	providers, ok := config["providers"].(map[string]interface{})
	if !ok {
		return nil // Invalid providers structure
	}

	// Generate provider configurations
	for provider, orgs := range groupedRepos {
		if provider == "unknown" {
			continue // Skip unknown providers
		}

		providerConfig := generateProviderConfig(provider, orgs, basePath)
		if providerConfig != nil {
			providers[provider] = providerConfig
		}
	}

	return config
}

// generateProviderConfig generates configuration for a specific provider.
func generateProviderConfig(provider string, orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	switch provider {
	case "github":
		return generateGitHubConfig(orgs, basePath)
	case "gitlab":
		return generateGitLabConfig(orgs, basePath)
	case "bitbucket":
		return generateBitbucketConfig(orgs, basePath)
	case "gitea":
		return generateGiteaConfig(orgs, basePath)
	default:
		return generateGenericProviderConfig(provider, orgs, basePath)
	}
}

// generateGitHubConfig generates GitHub-specific configuration.
func generateGitHubConfig(orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	var organizations []map[string]interface{}

	for orgName, repos := range orgs {
		orgConfig := map[string]interface{}{
			"name":      orgName,
			"clone_dir": filepath.Join(basePath, "github", orgName),
		}

		// Add repository-specific configuration if needed
		if hasPrivateRepos(repos) {
			orgConfig["visibility"] = "private"
			orgConfig["auth"] = map[string]interface{}{
				"token": "${GITHUB_TOKEN}",
			}
		}

		// Add exclude patterns if there are archived/deprecated repos
		excludePatterns := generateExcludePatterns(repos)
		if len(excludePatterns) > 0 {
			orgConfig["exclude"] = excludePatterns
		}

		organizations = append(organizations, orgConfig)
	}

	return map[string]interface{}{
		"organizations": organizations,
	}
}

// generateGitLabConfig generates GitLab-specific configuration.
func generateGitLabConfig(orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	var groups []map[string]interface{}

	for groupName, repos := range orgs {
		groupConfig := map[string]interface{}{
			"name":      groupName,
			"clone_dir": filepath.Join(basePath, "gitlab", groupName),
		}

		if hasPrivateRepos(repos) {
			groupConfig["visibility"] = "private"
			groupConfig["auth"] = map[string]interface{}{
				"token": "${GITLAB_TOKEN}",
			}
		}

		groups = append(groups, groupConfig)
	}

	return map[string]interface{}{
		"groups": groups,
	}
}

// generateBitbucketConfig generates Bitbucket-specific configuration.
func generateBitbucketConfig(orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	var workspaces []map[string]interface{}

	for workspaceName, repos := range orgs {
		workspaceConfig := map[string]interface{}{
			"name":      workspaceName,
			"clone_dir": filepath.Join(basePath, "bitbucket", workspaceName),
		}

		if hasPrivateRepos(repos) {
			workspaceConfig["auth"] = map[string]interface{}{
				"token": "${BITBUCKET_TOKEN}",
			}
		}

		workspaces = append(workspaces, workspaceConfig)
	}

	return map[string]interface{}{
		"workspaces": workspaces,
	}
}

// generateGiteaConfig generates Gitea-specific configuration.
func generateGiteaConfig(orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	var organizations []map[string]interface{}

	for orgName, repos := range orgs {
		orgConfig := map[string]interface{}{
			"name":      orgName,
			"clone_dir": filepath.Join(basePath, "gitea", orgName),
			"base_url":  "https://gitea.com", // Default, should be customized
		}

		if hasPrivateRepos(repos) {
			orgConfig["auth"] = map[string]interface{}{
				"token": "${GITEA_TOKEN}",
			}
		}

		organizations = append(organizations, orgConfig)
	}

	return map[string]interface{}{
		"organizations": organizations,
	}
}

// generateGenericProviderConfig generates configuration for unknown providers.
func generateGenericProviderConfig(provider string, orgs map[string][]discovery.DiscoveredRepo, basePath string) map[string]interface{} {
	var repositories []map[string]interface{}

	for orgName, repos := range orgs {
		for _, repo := range repos {
			repoConfig := map[string]interface{}{
				"name":       repo.RepoName,
				"url":        repo.RemoteURL,
				"clone_dir":  filepath.Join(basePath, provider, orgName),
				"local_path": repo.Path,
			}

			repositories = append(repositories, repoConfig)
		}
	}

	return map[string]interface{}{
		"repositories": repositories,
	}
}

// hasPrivateRepos checks if any repositories in the list might be private.
func hasPrivateRepos(repos []discovery.DiscoveredRepo) bool {
	for _, repo := range repos {
		// Heuristic: if URL contains authentication or is SSH, likely private
		if containsAuth(repo.RemoteURL) {
			return true
		}
	}
	return false
}

// containsAuth checks if a URL contains authentication information.
func containsAuth(url string) bool {
	return url != "" && (url[0:4] == "git@" || containsAtSymbol(url))
}

// containsAtSymbol checks if URL contains @ symbol (indicating auth).
func containsAtSymbol(url string) bool {
	for _, char := range url {
		if char == '@' {
			return true
		}
	}
	return false
}

// generateExcludePatterns generates exclude patterns based on repository names.
func generateExcludePatterns(repos []discovery.DiscoveredRepo) []string {
	var patterns []string
	patternMap := make(map[string]bool)

	for _, repo := range repos {
		name := repo.RepoName

		// Check for common patterns
		if containsPattern(name, "archive") {
			pattern := ".*-archive$"
			if !patternMap[pattern] {
				patterns = append(patterns, pattern)
				patternMap[pattern] = true
			}
		}

		if containsPattern(name, "deprecated") {
			pattern := ".*-deprecated$"
			if !patternMap[pattern] {
				patterns = append(patterns, pattern)
				patternMap[pattern] = true
			}
		}

		if containsPattern(name, "legacy") {
			pattern := ".*-legacy$"
			if !patternMap[pattern] {
				patterns = append(patterns, pattern)
				patternMap[pattern] = true
			}
		}
	}

	return patterns
}

// containsPattern checks if a string contains a pattern.
func containsPattern(s, pattern string) bool {
	return len(s) >= len(pattern) && findPattern(s, pattern)
}

// findPattern finds a pattern in a string.
func findPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}

// loadExistingConfig loads an existing configuration file.
func loadExistingConfig(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// mergeConfigurations merges two configurations.
func mergeConfigurations(existing, newConfig map[string]interface{}) map[string]interface{} {
	// Simple merge strategy - prefer new configuration but preserve existing structure
	merged := make(map[string]interface{})

	// Copy existing configuration
	for key, value := range existing {
		merged[key] = value
	}

	// Merge providers
	if existingProviders, ok := existing["providers"].(map[string]interface{}); ok {
		if newProviders, ok := newConfig["providers"].(map[string]interface{}); ok {
			mergedProviders := make(map[string]interface{})

			// Copy existing providers
			for provider, config := range existingProviders {
				mergedProviders[provider] = config
			}

			// Add new providers
			for provider, config := range newProviders {
				mergedProviders[provider] = config
			}

			merged["providers"] = mergedProviders
		}
	} else {
		merged["providers"] = newConfig["providers"]
	}

	return merged
}

// saveConfiguration saves the configuration to a file.
func saveConfiguration(config map[string]interface{}, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// displayDiscoverySummary displays a summary of the discovery results.
func displayDiscoverySummary(groupedRepos map[string]map[string][]discovery.DiscoveredRepo, allRepos []discovery.DiscoveredRepo) {
	fmt.Printf("\n📊 Discovery Summary:\n")
	fmt.Printf("   Total repositories: %d\n", len(allRepos))

	// Sort providers for consistent output
	var providers []string
	for provider := range groupedRepos {
		providers = append(providers, provider)
	}
	sort.Strings(providers)

	for _, provider := range providers {
		orgs := groupedRepos[provider]
		orgCount := len(orgs)
		repoCount := 0

		for _, repos := range orgs {
			repoCount += len(repos)
		}

		fmt.Printf("   %s: %d organizations, %d repositories\n", provider, orgCount, repoCount)
	}

	// Calculate total size
	var totalSize int64
	for _, repo := range allRepos {
		totalSize += repo.Size
	}

	fmt.Printf("   Total size: %.2f MB\n", float64(totalSize)/(1024*1024))
}

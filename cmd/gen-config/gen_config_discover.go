// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package genconfig

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gizzahub/gzh-manager-go/internal/helpers"
	"github.com/spf13/cobra"
)

const (
	protocolHTTPS  = "https"
	providerGitlab = "gitlab"
)

type genConfigDiscoverOptions struct {
	directory  string
	outputFile string
	recursive  bool
	force      bool
}

func defaultGenConfigDiscoverOptions() *genConfigDiscoverOptions {
	return &genConfigDiscoverOptions{
		directory:  ".",
		outputFile: "bulk-clone.yaml",
		recursive:  false,
		force:      false,
	}
}

type ConfigData struct {
	Version   string
	Protocol  string
	RepoRoots []RepoRootConfig
	Ignores   []string
}

type RepoRootConfig struct {
	Provider string
	OrgName  string
	RootPath string
	Protocol string
}

func newGenConfigDiscoverCmd(_ context.Context) *cobra.Command {
	o := defaultGenConfigDiscoverOptions()

	cmd := &cobra.Command{
		Use:   "discover [directory]",
		Short: "Discover existing Git repositories and generate configuration",
		Long: `Discover existing Git repositories in a directory and generate a bulk-clone.yaml
configuration file based on the found repositories.

This command will scan the specified directory for Git repositories, extract
their remote URLs, and generate a configuration file that can be used to
recreate the same repository structure.

Features:
- Detects Git repositories automatically
- Extracts provider information from remote URLs (GitHub, GitLab, etc.)
- Groups repositories by organization/group
- Suggests appropriate directory structures
- Generates protocol configurations based on remote URLs

Examples:
  # Discover repositories in current directory
  gz gen-config discover

  # Discover repositories recursively in ~/work
  gz gen-config discover ~/work --recursive

  # Generate to custom output file
  gz gen-config discover ~/projects --output my-discovered-config.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.outputFile, "output", "o", o.outputFile, "Output configuration file")
	cmd.Flags().BoolVarP(&o.recursive, "recursive", "r", o.recursive, "Search recursively for Git repositories")
	cmd.Flags().BoolVarP(&o.force, "force", "f", o.force, "Force overwrite existing configuration file")

	return cmd
}

type DiscoveredRepo struct {
	Path      string
	Provider  string
	OrgName   string
	RepoName  string
	Protocol  string
	RemoteURL string
}

func (o *genConfigDiscoverOptions) run(_ *cobra.Command, args []string) error {
	// Set directory from args if provided
	if len(args) > 0 {
		o.directory = args[0]
	}

	// Check if output file exists
	if !o.force {
		if _, err := os.Stat(o.outputFile); err == nil {
			return fmt.Errorf("configuration file already exists: %s (use --force to overwrite)", o.outputFile)
		}
	}

	// Expand directory path
	directory, err := filepath.Abs(o.directory)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	fmt.Printf("🔍 Discovering Git repositories in: %s\n", directory)

	if o.recursive {
		fmt.Println("   Searching recursively...")
	}

	// Discover repositories
	repos, err := o.discoverRepositories(directory)
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("❌ No Git repositories found in the specified directory")
		return nil
	}

	fmt.Printf("✅ Found %d Git repositories\n\n", len(repos))

	// Generate configuration
	config := o.generateConfigFromRepos(repos, directory)

	// Write configuration file
	yamlContent := o.generateYAMLFromConfig(config)

	err = os.WriteFile(o.outputFile, []byte(yamlContent), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("📝 Configuration file generated: %s\n", o.outputFile)
	fmt.Println("\nDiscovered repositories:")

	for provider, orgs := range o.groupReposByProvider(repos) {
		fmt.Printf("  %s:\n", strings.ToUpper(provider[:1])+provider[1:])

		for orgName, repoCount := range orgs {
			fmt.Printf("    %s: %d repositories\n", orgName, repoCount)
		}
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Review and edit the generated configuration file")
	fmt.Println("2. Adjust target paths as needed")
	fmt.Println("3. Configure authentication tokens")
	fmt.Printf("4. Test the configuration: gz bulk-clone --config %s --dry-run\n", o.outputFile)

	return nil
}

func (o *genConfigDiscoverOptions) discoverRepositories(rootDir string) ([]DiscoveredRepo, error) {
	var repos []DiscoveredRepo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Skip inaccessible directories
		}

		// Skip if not recursive and not in root directory
		if !o.recursive && filepath.Dir(path) != rootDir {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Check if this is a Git repository
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)

			repo, err := o.analyzeRepository(repoPath)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to analyze repository at %s: %v\n", repoPath, err)
				return nil
			}

			if repo != nil {
				repos = append(repos, *repo)
			}

			return filepath.SkipDir // Don't descend into .git directory
		}

		return nil
	})

	return repos, err
}

func (o *genConfigDiscoverOptions) analyzeRepository(repoPath string) (*DiscoveredRepo, error) {
	// Check if it's actually a Git repository
	repoType, _ := helpers.CheckGitRepoType(repoPath)
	if repoType == helpers.RepoTypeNone {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Get remote URL
	remoteURL, err := o.getRemoteURL(repoPath)
	if err != nil || remoteURL == "" {
		return nil, fmt.Errorf("no remote URL found")
	}

	// Parse the remote URL
	provider, orgName, repoName, protocol := o.parseRemoteURL(remoteURL)
	if provider == "" {
		return nil, fmt.Errorf("unknown provider for URL: %s", remoteURL)
	}

	return &DiscoveredRepo{
		Path:      repoPath,
		Provider:  provider,
		OrgName:   orgName,
		RepoName:  repoName,
		Protocol:  protocol,
		RemoteURL: remoteURL,
	}, nil
}

func (o *genConfigDiscoverOptions) getRemoteURL(repoPath string) (string, error) {
	// Try to read .git/config file
	configPath := filepath.Join(repoPath, ".git", "config")

	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Look for remote origin URL
	lines := strings.Split(string(content), "\n")
	inOriginSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[remote \"origin\"]" {
			inOriginSection = true
			continue
		}

		if strings.HasPrefix(line, "[") && inOriginSection {
			break // End of origin section
		}

		if inOriginSection && strings.HasPrefix(line, "url = ") {
			return strings.TrimPrefix(line, "url = "), nil
		}
	}

	return "", fmt.Errorf("no remote origin URL found")
}

func (o *genConfigDiscoverOptions) parseRemoteURL(remoteURL string) (provider, orgName, repoName, protocol string) {
	// SSH URL patterns
	sshPatterns := map[string]*regexp.Regexp{
		"github":       regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`),
		providerGitlab: regexp.MustCompile(`^git@gitlab\.com:([^/]+)/(.+?)(?:\.git)?$`),
		"gitea":        regexp.MustCompile(`^git@gitea\.com:([^/]+)/(.+?)(?:\.git)?$`),
	}

	// Check SSH patterns
	for prov, pattern := range sshPatterns {
		if matches := pattern.FindStringSubmatch(remoteURL); matches != nil {
			return prov, matches[1], matches[2], "ssh"
		}
	}

	// HTTPS URL patterns
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", "", ""
	}

	host := parsedURL.Host
	path := strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return "", "", "", ""
	}

	protocol = parsedURL.Scheme
	if protocol == "" {
		protocol = protocolHTTPS
	}

	switch host {
	case "github.com":
		return "github", pathParts[0], pathParts[1], protocol
	case "gitlab.com":
		return providerGitlab, pathParts[0], pathParts[1], protocol
	case "gitea.com":
		return "gitea", pathParts[0], pathParts[1], protocol
	default:
		// Custom GitLab/Gitea instance
		if strings.Contains(host, providerGitlab) {
			return providerGitlab, pathParts[0], pathParts[1], protocol
		}

		if strings.Contains(host, "gitea") {
			return "gitea", pathParts[0], pathParts[1], protocol
		}
	}

	return "", "", "", ""
}

func (o *genConfigDiscoverOptions) generateConfigFromRepos(repos []DiscoveredRepo, baseDir string) ConfigData {
	config := ConfigData{
		Version:   "0.1",
		Protocol:  protocolHTTPS,
		RepoRoots: []RepoRootConfig{},
		Ignores:   []string{"test-.*", ".*-archive"},
	}

	// Group repositories by provider and organization
	groups := make(map[string]map[string][]DiscoveredRepo)
	for _, repo := range repos {
		if groups[repo.Provider] == nil {
			groups[repo.Provider] = make(map[string][]DiscoveredRepo)
		}

		groups[repo.Provider][repo.OrgName] = append(groups[repo.Provider][repo.OrgName], repo)
	}

	// Generate repository roots
	for provider, orgs := range groups {
		for orgName, orgRepos := range orgs {
			// Find common root path for this organization
			rootPath := o.findCommonRootPath(orgRepos, baseDir)

			// Determine most common protocol for this organization
			protocol := o.determineMostCommonProtocol(orgRepos)

			repoRoot := RepoRootConfig{
				Provider: provider,
				OrgName:  orgName,
				RootPath: rootPath,
				Protocol: protocol,
			}
			config.RepoRoots = append(config.RepoRoots, repoRoot)
		}
	}

	return config
}

func (o *genConfigDiscoverOptions) findCommonRootPath(repos []DiscoveredRepo, baseDir string) string {
	if len(repos) == 0 {
		return ""
	}

	// For single repository, use parent directory
	if len(repos) == 1 {
		return filepath.Dir(repos[0].Path)
	}

	// Find common parent directory
	commonPath := repos[0].Path
	for _, repo := range repos[1:] {
		commonPath = o.findCommonPath(commonPath, repo.Path)
	}

	// If common path is too high up, use organization-specific path
	relPath, err := filepath.Rel(baseDir, commonPath)
	if err != nil || relPath == "." || relPath == ".." || strings.HasPrefix(relPath, "../") {
		// Create organization-specific path
		provider := repos[0].Provider
		orgName := repos[0].OrgName

		return fmt.Sprintf("$HOME/%s/%s", provider, orgName)
	}

	return commonPath
}

func (o *genConfigDiscoverOptions) findCommonPath(path1, path2 string) string {
	parts1 := strings.Split(filepath.Clean(path1), string(filepath.Separator))
	parts2 := strings.Split(filepath.Clean(path2), string(filepath.Separator))

	var common []string

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] == parts2[i] {
			common = append(common, parts1[i])
		} else {
			break
		}
	}

	return strings.Join(common, string(filepath.Separator))
}

func (o *genConfigDiscoverOptions) determineMostCommonProtocol(repos []DiscoveredRepo) string {
	protocolCount := make(map[string]int)
	for _, repo := range repos {
		protocolCount[repo.Protocol]++
	}

	maxCount := 0
	mostCommon := protocolHTTPS

	for protocol, count := range protocolCount {
		if count > maxCount {
			maxCount = count
			mostCommon = protocol
		}
	}

	return mostCommon
}

func (o *genConfigDiscoverOptions) groupReposByProvider(repos []DiscoveredRepo) map[string]map[string]int {
	groups := make(map[string]map[string]int)
	for _, repo := range repos {
		if groups[repo.Provider] == nil {
			groups[repo.Provider] = make(map[string]int)
		}

		groups[repo.Provider][repo.OrgName]++
	}

	return groups
}

func (o *genConfigDiscoverOptions) generateYAMLFromConfig(config ConfigData) string {
	var content strings.Builder

	content.WriteString("# Generated by gzh-manager gen-config discover\n")
	content.WriteString("# Auto-discovered Git repository configuration\n\n")
	content.WriteString(fmt.Sprintf("version: %q\n\n", config.Version))

	// Default section
	content.WriteString("# Global default settings\n")
	content.WriteString("default:\n")
	content.WriteString(fmt.Sprintf("  protocol: %s\n", config.Protocol))
	content.WriteString("  github:\n")
	content.WriteString("    root_path: \"$HOME/github-repos\"\n")
	content.WriteString("  gitlab:\n")
	content.WriteString("    root_path: \"$HOME/gitlab-repos\"\n\n")

	// Repository roots
	content.WriteString("# Discovered repository configurations\n")
	content.WriteString("repo_roots:\n")

	for _, root := range config.RepoRoots {
		content.WriteString(fmt.Sprintf("  - root_path: %q\n", root.RootPath))
		content.WriteString(fmt.Sprintf("    provider: %q\n", root.Provider))
		content.WriteString(fmt.Sprintf("    protocol: %q\n", root.Protocol))
		content.WriteString(fmt.Sprintf("    org_name: %q\n", root.OrgName))
		content.WriteString("\n")
	}

	// Ignore patterns
	content.WriteString("# Common ignore patterns\n")
	content.WriteString("ignore_names:\n")

	for _, pattern := range config.Ignores {
		content.WriteString(fmt.Sprintf("  - %q\n", pattern))
	}

	return content.String()
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package discovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RepoDiscoverer handles repository discovery and configuration generation.
type RepoDiscoverer struct {
	BasePath       string
	MaxDepth       int
	IgnorePatterns []string
	FollowSymlinks bool
}

// DiscoveredRepo represents a discovered repository with metadata.
type DiscoveredRepo struct {
	Path       string `yaml:"path"`
	RemoteURL  string `yaml:"remoteUrl"`
	Provider   string `yaml:"provider"`
	Org        string `yaml:"org"`
	RepoName   string `yaml:"repoName"`
	Branch     string `yaml:"branch"`
	LastCommit string `yaml:"lastCommit"`
	Size       int64  `yaml:"sizeBytes"`
}

// NewRepoDiscoverer creates a new repository discoverer.
func NewRepoDiscoverer(basePath string) *RepoDiscoverer {
	return &RepoDiscoverer{
		BasePath:       basePath,
		MaxDepth:       3,
		IgnorePatterns: []string{".git", "node_modules", ".venv", "target", "build"},
		FollowSymlinks: false,
	}
}

// DiscoverRepos discovers all Git repositories in the base path.
func (rd *RepoDiscoverer) DiscoverRepos() ([]DiscoveredRepo, error) {
	var repos []DiscoveredRepo

	err := rd.walkDirectory(rd.BasePath, 0, &repos)
	if err != nil {
		return nil, fmt.Errorf("failed to discover repositories: %w", err)
	}

	return repos, nil
}

// walkDirectory recursively walks directories to find Git repositories.
func (rd *RepoDiscoverer) walkDirectory(dir string, depth int, repos *[]DiscoveredRepo) error {
	if depth > rd.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Check if current directory is a Git repository
	gitDir := filepath.Join(dir, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		repo, err := rd.analyzeRepository(dir)
		if err == nil {
			*repos = append(*repos, *repo)
		}
		// Don't recurse into .git directories
		return nil
	}

	// Recurse into subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if rd.shouldIgnore(name) {
			continue
		}

		subPath := filepath.Join(dir, name)

		// Handle symlinks
		if !rd.FollowSymlinks {
			if info, err := entry.Info(); err == nil {
				if info.Mode()&os.ModeSymlink != 0 {
					continue
				}
			}
		}

		if err := rd.walkDirectory(subPath, depth+1, repos); err != nil {
			// Log error but continue with other directories
			continue
		}
	}

	return nil
}

// analyzeRepository analyzes a Git repository and extracts metadata.
func (rd *RepoDiscoverer) analyzeRepository(repoPath string) (*DiscoveredRepo, error) {
	// Get remote URL using git command
	remoteURL, err := rd.getRemoteURL(repoPath)
	if err != nil {
		// Continue without remote URL if git command fails
		remoteURL = ""
	}

	// Parse provider, org, and repo name from URL
	provider, org, repoName := rd.parseRemoteURL(remoteURL)

	// Get current branch
	branch, err := rd.getCurrentBranch(repoPath)
	if err != nil {
		branch = ""
	}

	// Get last commit
	lastCommit, err := rd.getLastCommit(repoPath)
	if err != nil {
		lastCommit = ""
	}

	// Calculate repository size
	size := rd.calculateRepoSize(repoPath)

	return &DiscoveredRepo{
		Path:       repoPath,
		RemoteURL:  remoteURL,
		Provider:   provider,
		Org:        org,
		RepoName:   repoName,
		Branch:     branch,
		LastCommit: lastCommit,
		Size:       size,
	}, nil
}

// parseRemoteURL parses a Git remote URL to extract provider, organization, and repository name.
func (rd *RepoDiscoverer) parseRemoteURL(url string) (provider, org, repo string) {
	if url == "" {
		return "", "", ""
	}

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle different URL formats
	if strings.Contains(url, "github.com") {
		provider = "github"
		return rd.parseGitHubURL(url)
	} else if strings.Contains(url, "gitlab.com") {
		provider = "gitlab"
		return rd.parseGitLabURL(url)
	} else if strings.Contains(url, "bitbucket.org") {
		provider = "bitbucket"
		return rd.parseBitbucketURL(url)
	}

	// Generic parsing for other providers
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		repo = parts[len(parts)-1]
		org = parts[len(parts)-2]

		// Try to extract provider from hostname
		for _, part := range parts {
			if strings.Contains(part, ".") && !strings.HasPrefix(part, "git@") {
				provider = strings.Split(part, ".")[0]
				break
			}
		}
	}

	return provider, org, repo
}

// parseGitHubURL parses GitHub-specific URLs.
func (rd *RepoDiscoverer) parseGitHubURL(url string) (provider, org, repo string) {
	provider = "github"

	// Handle SSH format: git@github.com:org/repo
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://github.com/org/repo
	if strings.HasPrefix(url, "https://github.com/") {
		path := strings.TrimPrefix(url, "https://github.com/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	return provider, org, repo
}

// parseGitLabURL parses GitLab-specific URLs.
func (rd *RepoDiscoverer) parseGitLabURL(url string) (provider, org, repo string) {
	provider = "gitlab"

	// Handle SSH format: git@gitlab.com:org/repo
	if strings.HasPrefix(url, "git@gitlab.com:") {
		path := strings.TrimPrefix(url, "git@gitlab.com:")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://gitlab.com/org/repo
	if strings.HasPrefix(url, "https://gitlab.com/") {
		path := strings.TrimPrefix(url, "https://gitlab.com/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	return provider, org, repo
}

// parseBitbucketURL parses Bitbucket-specific URLs.
func (rd *RepoDiscoverer) parseBitbucketURL(url string) (provider, org, repo string) {
	provider = "bitbucket"

	// Handle SSH format: git@bitbucket.org:org/repo
	if strings.HasPrefix(url, "git@bitbucket.org:") {
		path := strings.TrimPrefix(url, "git@bitbucket.org:")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://bitbucket.org/org/repo
	if strings.HasPrefix(url, "https://bitbucket.org/") {
		path := strings.TrimPrefix(url, "https://bitbucket.org/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	return provider, org, repo
}

// shouldIgnore checks if a directory should be ignored.
func (rd *RepoDiscoverer) shouldIgnore(name string) bool {
	for _, pattern := range rd.IgnorePatterns {
		if name == pattern || strings.HasPrefix(name, pattern) {
			return true
		}
	}
	return false
}

// calculateRepoSize calculates the approximate size of a repository.
func (rd *RepoDiscoverer) calculateRepoSize(repoPath string) int64 {
	var size int64

	err := filepath.Walk(repoPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // 접근 불가능한 파일은 건너뜀
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}

	return size
}

// SetMaxDepth sets the maximum directory depth for discovery.
func (rd *RepoDiscoverer) SetMaxDepth(depth int) {
	rd.MaxDepth = depth
}

// SetIgnorePatterns sets custom ignore patterns.
func (rd *RepoDiscoverer) SetIgnorePatterns(patterns []string) {
	rd.IgnorePatterns = patterns
}

// SetFollowSymlinks enables or disables symlink following.
func (rd *RepoDiscoverer) SetFollowSymlinks(follow bool) {
	rd.FollowSymlinks = follow
}

// getRemoteURL gets the remote URL for a Git repository.
func (rd *RepoDiscoverer) getRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		// Try to get any remote if origin doesn't exist
		cmd = exec.Command("git", "-C", repoPath, "remote")
		remoteOutput, err2 := cmd.Output()
		if err2 != nil {
			return "", fmt.Errorf("failed to get remotes: %w", err2)
		}

		remotes := strings.Fields(strings.TrimSpace(string(remoteOutput)))
		if len(remotes) > 0 {
			// Validate remote name to prevent injection
			remoteName := remotes[0]
			if !isValidRemoteName(remoteName) {
				return "", fmt.Errorf("invalid remote name: %s", remoteName)
			}
			// Get URL for first remote
			// #nosec G204 - remoteName is validated for safety
			cmd = exec.Command("git", "-C", repoPath, "remote", "get-url", remoteName)
			output, err = cmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to get remote URL: %w", err)
			}
		} else {
			return "", fmt.Errorf("no remotes found")
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// getCurrentBranch gets the current branch of a Git repository.
func (rd *RepoDiscoverer) getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getLastCommit gets the last commit hash of a Git repository.
func (rd *RepoDiscoverer) getLastCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// isValidRemoteName validates a git remote name to prevent command injection.
func isValidRemoteName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}

	// Git remote names should only contain alphanumeric characters, hyphens, underscores, and dots
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validPattern.MatchString(name)
}

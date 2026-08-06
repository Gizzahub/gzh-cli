// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package discovery

import (
	"context"
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
//
// ctx는 훑기 전체를 감싼다. 저장소 하나마다 git을 세 번 부르기 때문에 나무가
// 크면 바깥 프로세스가 수백 개 뜬다. 맥락이 없으면 그 중 어느 것도 끊을 수
// 없고 Ctrl+C가 먹지 않는다.
func (rd *RepoDiscoverer) DiscoverRepos(ctx context.Context) ([]DiscoveredRepo, error) {
	var repos []DiscoveredRepo

	err := rd.walkDirectory(ctx, rd.BasePath, 0, &repos)
	if err != nil {
		// 부르는 쪽이 "failed to discover repositories"로 한 번 더 감싼다.
		// 같은 말을 두 번 적지 않도록 여기서는 어디를 훑다 걸렸는지만 밝힌다.
		return nil, fmt.Errorf("failed to walk %s: %w", rd.BasePath, err)
	}

	return repos, nil
}

// walkDirectory recursively walks directories to find Git repositories.
func (rd *RepoDiscoverer) walkDirectory(ctx context.Context, dir string, depth int, repos *[]DiscoveredRepo) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
		repo, err := rd.analyzeRepository(ctx, dir)

		switch {
		case err == nil:
			*repos = append(*repos, *repo)
		case ctx.Err() != nil:
			// 맥락이 끊겨서 실패한 것이면 위로 올린다. 그 밖의 실패는
			// 예전처럼 이 저장소만 건너뛴다.
			return err
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

		if err := rd.walkDirectory(ctx, subPath, depth+1, repos); err != nil {
			// 맥락이 끊겼으면 남은 디렉토리를 마저 돌 이유가 없다.
			if ctx.Err() != nil {
				return err
			}

			// Log error but continue with other directories
			continue
		}
	}

	return nil
}

// analyzeRepository analyzes a Git repository and extracts metadata.
func (rd *RepoDiscoverer) analyzeRepository(ctx context.Context, repoPath string) (*DiscoveredRepo, error) {
	// Get remote URL using git command
	remoteURL, err := rd.getRemoteURL(ctx, repoPath)
	if err != nil {
		// Continue without remote URL if git command fails
		remoteURL = ""
	}

	// Parse provider, org, and repo name from URL
	provider, org, repoName := rd.parseRemoteURL(remoteURL)

	// Get current branch
	branch, err := rd.getCurrentBranch(ctx, repoPath)
	if err != nil {
		branch = ""
	}

	// Get last commit
	lastCommit, err := rd.getLastCommit(ctx, repoPath)
	if err != nil {
		lastCommit = ""
	}

	// 위 세 번의 git 호출은 실패를 삼킨다 -- 원격이 없는 저장소도 결과에
	// 넣기 위해서다. 맥락이 끊겨서 실패한 것이라면 이야기가 다르다. 빈
	// 칸투성이 항목을 설정 파일에 적어 넣지 않도록 여기서 가른다.
	if err := ctx.Err(); err != nil {
		return nil, err
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
		return rd.parseGitHubURL(url)
	} else if strings.Contains(url, "gitlab.com") {
		return rd.parseGitLabURL(url)
	} else if strings.Contains(url, "bitbucket.org") {
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
	if after, ok := strings.CutPrefix(url, "git@github.com:"); ok {
		path := after
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://github.com/org/repo
	if after, ok := strings.CutPrefix(url, "https://github.com/"); ok {
		path := after
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
	if after, ok := strings.CutPrefix(url, "git@gitlab.com:"); ok {
		path := after
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://gitlab.com/org/repo
	if after, ok := strings.CutPrefix(url, "https://gitlab.com/"); ok {
		path := after
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
	if after, ok := strings.CutPrefix(url, "git@bitbucket.org:"); ok {
		path := after
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			org = parts[0]
			repo = parts[1]
		}
		return provider, org, repo
	}

	// Handle HTTPS format: https://bitbucket.org/org/repo
	if after, ok := strings.CutPrefix(url, "https://bitbucket.org/"); ok {
		path := after
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
func (rd *RepoDiscoverer) getRemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		// Try to get any remote if origin doesn't exist
		cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "remote")
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
			cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "get-url", remoteName)
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
func (rd *RepoDiscoverer) getCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getLastCommit gets the last commit hash of a Git repository.
func (rd *RepoDiscoverer) getLastCommit(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "HEAD")
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

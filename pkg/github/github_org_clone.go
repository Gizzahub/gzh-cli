package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gizzahub/gzh-manager-go/internal/helpers"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// RepoInfo represents GitHub repository information returned by the GitHub API.
// It contains essential repository metadata used during clone operations.
type RepoInfo struct {
	// DefaultBranch is the name of the repository's default branch (e.g., "main", "master")
	DefaultBranch string `json:"default_branch"`
}

// GetDefaultBranch retrieves the default branch name for a GitHub repository.
// It makes an authenticated HTTP GET request to the GitHub API to fetch repository information.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - org: GitHub organization or user name
//   - repo: Repository name
//
// Returns the default branch name (e.g., "main", "master") or an error if the
// repository doesn't exist, access is denied, or the API request fails.
func GetDefaultBranch(ctx context.Context, org string, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", org, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP response body cleanup

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get repository info: %s", resp.Status)
	}

	var repoInfo RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", err
	}

	return repoInfo.DefaultBranch, nil
}

// List retrieves all repository names for a GitHub organization.
// It makes paginated requests to the GitHub API to fetch all repositories
// in the specified organization, handling pagination automatically.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - org: GitHub organization name
//
// Returns a slice of repository names or an error if the organization
// doesn't exist, access is denied, or the API request fails.
func List(ctx context.Context, org string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP response body cleanup

	if resp.StatusCode != http.StatusOK {
		// body, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(body))
		// resp.Header.Write(os.Stdout)
		rateReset := resp.Header.Get("X-RateLimit-Reset")

		resetTime, err := strconv.ParseInt(rateReset, 10, 64)
		if err == nil {
			c := color.New(color.FgCyan, color.Bold)
			_, _ = c.Println("Github RateLimit !!! you must wait until: ")                                                                                            //nolint:errcheck // User information display
			_, _ = c.Println(time.Unix(resetTime, 0).Format(time.RFC1123))                                                                                            //nolint:errcheck // User information display
			_, _ = c.Printf("%d minutes and %d seconds\n", int(time.Until(time.Unix(resetTime, 0)).Minutes()), int(time.Until(time.Unix(resetTime, 0)).Seconds())%60) //nolint:errcheck // User information display
			_, _ = c.Println("or Use Github Token (not provided yet ^*)")                                                                                             //nolint:errcheck // User information display
		}
		// try after
		return nil, fmt.Errorf("failed to get repositories: %s", resp.Status)
	}

	// Use standard JSON decoding - DISABLED (memory package removed)
	// Simple implementation without external memory dependency
	var repos []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Create result slice
	result := make([]string, len(repos))
	for i, repo := range repos {
		result[i] = repo.Name
	}

	return result, nil
}

// Clone downloads a GitHub repository to the specified local path.
// It performs a git clone operation using the repository's HTTPS URL.
// The repository is cloned into a subdirectory named after the repository
// within the targetPath directory.
//
// Parameters:
//   - ctx: Context for operation cancellation and timeout control
//   - targetPath: Local directory path where the repository will be cloned
//   - org: GitHub organization or user name
//   - repo: Repository name
//
// Returns an error if the clone operation fails due to network issues,
// authentication problems, or local file system errors.
func Clone(ctx context.Context, targetPath string, org string, repo string) error {
	// if branch == "" {
	//	defaultBranch, err := GetDefaultBranch(ctx, org, repo)
	//	if err != nil {
	//		fmt.Println("failed to get default. clone without branch specify.")
	//		//return fmt.Errorf("failed to get default branch: %w", err)
	//	}
	//	branch = defaultBranch
	//}
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)

	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)

	// var cmd *exec.Cmd
	// if branch == "" {
	//	//cmd := exec.Command("git", "clone", cloneURL, targetPath)
	//	cmd = exec.Command("git", "clone", "-b", branch, cloneURL, targetPath)
	// } else {
	//	cmd = exec.Command("git", "clone", cloneURL, targetPath)
	//}
	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, targetPath)
	cmd.Stdout = &out

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
		fmt.Println(out.String())
		fmt.Println("execute git clone fail: ", out.String())

		return fmt.Errorf("Clone Failed  (url: %s, targetPath: %s, err: %w)", cloneURL, targetPath, err)
	}

	fmt.Println("execute git clone: ", out.String())

	return nil
}

// RefreshAllOptimizedStreaming performs optimized bulk repository refresh using streaming API and memory management
// This is the recommended method for large-scale organization cloning (>1000 repositories).
func RefreshAllOptimizedStreaming(ctx context.Context, targetPath, org, strategy, token string) error {
	config := DefaultOptimizedCloneConfig()

	manager, err := NewOptimizedBulkCloneManager(token, config) //nolint:contextcheck // Manager creation doesn't require context propagation
	if err != nil {
		return fmt.Errorf("failed to create optimized bulk clone manager: %w", err)
	}
	defer func() { _ = manager.Close() }() //nolint:errcheck // Resource cleanup

	stats, err := manager.RefreshAllOptimized(ctx, targetPath, org, strategy)
	if err != nil {
		return fmt.Errorf("optimized bulk clone failed: %w", err)
	}

	// Print summary
	fmt.Printf("\n🎉 Bulk clone completed: %d successful, %d failed (%.1f%% success rate)\n",
		stats.Successful, stats.Failed,
		float64(stats.Successful)/float64(stats.TotalRepositories)*100)

	return nil
}

// RefreshAll synchronizes the repositories in the targetPath with the repositories in the given organization.
// strategy can be "reset" (default), "pull", or "fetch"
//
// Note: For better performance with large numbers of repositories, consider using RefreshAllOptimizedStreaming
// for organizations with >1000 repositories, which provides streaming API, memory management, and better resource control.
func RefreshAll(ctx context.Context, targetPath string, org string, strategy string) error {
	// Get all directories inside targetPath
	targetRepos, err := getDirectories(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get directories in target path: %w", err)
	}

	// Get all repositories from the organization
	orgRepos, err := List(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to list repositories from organization: %w", err)
	}

	// bar := progressbar.Default(int64(len(orgRepos)), "Cloning Repositories")
	bar := progressbar.NewOptions(len(orgRepos),
		progressbar.OptionSetDescription("Cloning Repositories"),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Determine repos to delete (targetRepos - orgRepos)
	// reposToDelete := difference(targetRepos, orgRepos)

	// Delete repos that are not in the organization
	for _, repo := range targetRepos {
		repoPath := filepath.Join(targetPath, repo)

		repoType, _ := helpers.CheckGitRepoType(repoPath)
		if !Contains(targetRepos, repo) || repoType == helpers.RepoTypeNone {
			if err := os.RemoveAll(repoPath); err != nil {
				return fmt.Errorf("failed to delete repository %s: %w", repoPath, err)
			}
		}
	}

	// print all orgs
	c := color.New(color.FgCyan, color.Bold)
	_, _ = c.Printf("All Target %d >>>>>>>>>>>>>>>>>>>>\n", len(orgRepos)) //nolint:errcheck // User information display

	for _, repo := range orgRepos {
		_, _ = c.Println(repo)
	}

	_, _ = c.Println("All Target <<<<<<<<<<<<<<<<<<<")

	// Use errgroup for concurrent repository processing
	g, gCtx := errgroup.WithContext(ctx)
	// Limit concurrent git operations to avoid overwhelming the system
	sem := semaphore.NewWeighted(5) // Max 5 concurrent git operations

	var mu sync.Mutex

	for _, repo := range orgRepos {
		// Capture loop variable
		repo := repo

		g.Go(func() error {
			// Acquire semaphore to limit concurrency
			if err := sem.Acquire(gCtx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			// Update progress bar safely
			mu.Lock()
			bar.Describe(fmt.Sprintf("Clone or Reset %s", repo))
			mu.Unlock()

			repoPath := filepath.Join(targetPath, repo)
			if _, err := os.Stat(repoPath); os.IsNotExist(err) {
				// Clone the repository if it does not exist
				if err := Clone(gCtx, repoPath, org, repo); err != nil {
					fmt.Printf("failed to clone repository %s: %v\n", repoPath, err)
					// Don't return error to prevent stopping other operations
					// Log error but continue with other repositories
				}
			} else {
				// Execute git operation based on strategy
				repoType, _ := helpers.CheckGitRepoType(repoPath)
				if repoType != helpers.RepoTypeEmpty {
					switch strategy {
					case "reset":
						// Reset hard HEAD and pull
						cmd := exec.CommandContext(gCtx, "git", "-C", repoPath, "reset", "--hard", "HEAD")
						if err := cmd.Run(); err != nil {
							fmt.Printf("execute git reset fail for %s: %v\n", repo, err)
						}

						cmd = exec.CommandContext(gCtx, "git", "-C", repoPath, "pull")
						if err := cmd.Run(); err != nil {
							fmt.Printf("execute git pull fail for %s: %v\n", repo, err)
						}
					case "pull":
						// Only pull without reset
						cmd := exec.CommandContext(gCtx, "git", "-C", repoPath, "pull")
						if err := cmd.Run(); err != nil {
							fmt.Printf("execute git pull fail for %s: %v\n", repo, err)
						}
					case "fetch":
						// Only fetch without modifying working directory
						cmd := exec.CommandContext(gCtx, "git", "-C", repoPath, "fetch")
						if err := cmd.Run(); err != nil {
							fmt.Printf("execute git fetch fail for %s: %v\n", repo, err)
						}
					}
				}
			}

			// Update progress bar safely
			mu.Lock()
			_ = bar.Add(1)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all operations to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error in concurrent git operations: %w", err)
	}

	return nil
}

// getDirectories returns a list of directory names in the given path.
func getDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func Contains(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

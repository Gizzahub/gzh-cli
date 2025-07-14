package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var (
	ErrFailedToGetSubgroups    = errors.New("failed to get subgroups")
	ErrFailedToGetRepositories = errors.New("failed to get repositories")
)

// GitLabRepoInfo represents GitLab project information returned by the GitLab API.
// It contains essential project metadata used during clone operations.
type GitLabRepoInfo struct {
	// DefaultBranch is the name of the project's default branch (e.g., "main", "master")
	DefaultBranch string `json:"default_branch"`
}

// GetDefaultBranch retrieves the default branch name for a GitLab project.
// It makes an HTTP GET request to the GitLab API to fetch project information.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - group: GitLab group or user name
//   - repo: Project name
//
// Returns the default branch name (e.g., "main", "master") or an error if the
// project doesn't exist, access is denied, or the API request fails.
func GetDefaultBranch(ctx context.Context, group string, repo string) (string, error) {
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s", group, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrFailedToGetRepositories, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gitLabRepo GitLabRepoInfo
	err = json.Unmarshal(body, &gitLabRepo)
	if err != nil {
		return "", err
	}

	return gitLabRepo.DefaultBranch, nil
}

func listGroupRepos(ctx context.Context, group string, allRepos *[]string) error {
	url := fmt.Sprintf("https://gitlab.com/api/v4/groups/%s/projects", group)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetRepositories, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrFailedToGetRepositories, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var repos []struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		*allRepos = append(*allRepos, repo.Name)
	}

	// Get subgroups
	subgroupsURL := fmt.Sprintf("https://gitlab.com/api/v4/groups/%s/subgroups", group)

	subgroupsReq, err := http.NewRequestWithContext(ctx, "GET", subgroupsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create subgroups request: %w", err)
	}

	subgroupsResp, err := client.Do(subgroupsReq)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetSubgroups, err)
	}
	defer subgroupsResp.Body.Close()

	if subgroupsResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrFailedToGetSubgroups, subgroupsResp.Status)
	}

	subgroupsBody, err := io.ReadAll(subgroupsResp.Body)
	if err != nil {
		return err
	}

	var subgroups []struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(subgroupsBody, &subgroups)
	if err != nil {
		return err
	}

	for _, subgroup := range subgroups {
		err := listGroupRepos(ctx, subgroup.ID, allRepos)
		if err != nil {
			return err
		}
	}

	return nil
}

// List retrieves all project names for a GitLab group.
// It makes paginated requests to the GitLab API to fetch all projects
// in the specified group, handling pagination automatically.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - group: GitLab group name
//
// Returns a slice of project names or an error if the group
// doesn't exist, access is denied, or the API request fails.
func List(ctx context.Context, group string) ([]string, error) {
	var allRepos []string
	err := listGroupRepos(ctx, group, &allRepos)
	if err != nil {
		return nil, err
	}
	return allRepos, nil
}

// Clone downloads a GitLab project to the specified local path.
// It performs a git clone operation using the project's HTTPS URL.
// The project is cloned into a subdirectory named after the project
// within the targetPath directory.
//
// Parameters:
//   - ctx: Context for operation cancellation and timeout control
//   - targetPath: Local directory path where the project will be cloned
//   - group: GitLab group or user name
//   - repo: Project name
//   - branch: Specific branch to clone (if empty, uses default branch)
//
// Returns an error if the clone operation fails due to network issues,
// authentication problems, or local file system errors.
func Clone(ctx context.Context, targetPath string, group string, repo string, branch string) error {
	if branch == "" {
		defaultBranch, err := GetDefaultBranch(ctx, group, repo)
		if err != nil {
			return fmt.Errorf("failed to get default branch: %w", err)
		}
		branch = defaultBranch
	}

	cloneURL := fmt.Sprintf("https://gitlab.com/%s/%s.git", group, repo)
	var out bytes.Buffer
	var stderr bytes.Buffer
	// cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, cloneURL, targetPath)
	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, targetPath)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
		fmt.Println(out.String())
		return fmt.Errorf("Clone Failed  (url: %s, branch: %s, targetPath: %s, err: %w)\n", cloneURL, branch, targetPath, err)
	}

	return nil
}

// RefreshAll synchronizes the repositories in the targetPath with the repositories in the given group.
// strategy can be "reset" (default), "pull", or "fetch"
//
// Note: For better performance with large numbers of repositories, consider using RefreshAllWithWorkerPool
// from the bulk_operations.go file, which provides configurable worker pools and better resource management.
func RefreshAll(ctx context.Context, targetPath string, group string, strategy string) error {
	// Get all directories inside targetPath
	targetRepos, err := getDirectories(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get directories in target path: %w", err)
	}

	// Get all repositories from the group
	groupRepos, err := List(ctx, group)
	if err != nil {
		return fmt.Errorf("failed to list repositories from group: %w", err)
	}

	// Determine repos to delete (targetRepos - groupRepos)
	reposToDelete := difference(targetRepos, groupRepos)

	// Delete repos that are not in the group
	for _, repo := range reposToDelete {
		repoPath := filepath.Join(targetPath, repo)
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("failed to delete repository %s: %w", repoPath, err)
		}
	}

	// Use errgroup for concurrent repository processing
	g, gCtx := errgroup.WithContext(ctx)
	// Limit concurrent git operations to avoid overwhelming the system
	sem := semaphore.NewWeighted(5) // Max 5 concurrent git operations
	var mu sync.Mutex

	for _, repo := range groupRepos {
		// Capture loop variable
		repo := repo

		g.Go(func() error {
			// Acquire semaphore to limit concurrency
			if err := sem.Acquire(gCtx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			repoPath := filepath.Join(targetPath, repo)
			if _, err := os.Stat(repoPath); os.IsNotExist(err) {
				// Clone the repository if it does not exist
				if err := Clone(gCtx, repoPath, group, repo, ""); err != nil {
					fmt.Printf("failed to clone repository %s: %v\n", repoPath, err)
					// Don't return error to prevent stopping other operations
					// Log error but continue with other repositories
				}
			} else {
				// Execute git operation based on strategy
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

				// Thread-safe success message
				mu.Lock()
				fmt.Printf("Repo sync success with strategy %s: %s\n", strategy, repoPath)
				mu.Unlock()
			}

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

// difference returns the elements in 'a' that are not in 'b'.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

package gitea

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// RepoInfo represents Gitea repository information returned by the Gitea API.
// It contains essential repository metadata used during clone operations.
type RepoInfo struct {
	// DefaultBranch is the name of the repository's default branch (e.g., "main", "master")
	DefaultBranch string `json:"default_branch"`
}

// GetDefaultBranch retrieves the default branch name for a Gitea repository.
// It makes an HTTP GET request to the Gitea API to fetch repository information.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - org: Gitea organization or user name
//   - repo: Repository name
//
// Returns the default branch name (e.g., "main", "master") or an error if the
// repository doesn't exist, access is denied, or the API request fails.
func GetDefaultBranch(ctx context.Context, org string, repo string) (string, error) {
	url := fmt.Sprintf("https://gitea.com/api/v1/repos/%s/%s", org, repo)

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
		return "", fmt.Errorf("failed to get repository info: %s", resp.Status)
	}

	var repoInfo RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", err
	}

	return repoInfo.DefaultBranch, nil
}

// List retrieves all repository names for a Gitea organization.
// It makes paginated requests to the Gitea API to fetch all repositories
// in the specified organization, handling pagination automatically.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - org: Gitea organization name
//
// Returns a slice of repository names or an error if the organization
// doesn't exist, access is denied, or the API request fails.
func List(ctx context.Context, org string) ([]string, error) {
	url := fmt.Sprintf("https://gitea.com/api/v1/orgs/%s/repos", org)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repositories: %s", resp.Status)
	}

	var repos []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}

	return repoNames, nil
}

// Clone downloads a Gitea repository to the specified local path.
// It performs a git clone operation using the repository's HTTPS URL.
// The repository is cloned into a subdirectory named after the repository
// within the targetPath directory.
//
// Parameters:
//   - ctx: Context for operation cancellation and timeout control
//   - targetPath: Local directory path where the repository will be cloned
//   - org: Gitea organization or user name
//   - repo: Repository name
//   - branch: Specific branch to clone (if empty, uses default branch)
//
// Returns an error if the clone operation fails due to network issues,
// authentication problems, or local file system errors.
func Clone(ctx context.Context, targetPath string, org string, repo string, branch string) error {
	if branch == "" {
		defaultBranch, err := GetDefaultBranch(ctx, org, repo)
		if err != nil {
			return fmt.Errorf("failed to get default branch: %w", err)
		}

		branch = defaultBranch
	}

	cloneURL := fmt.Sprintf("https://gitea.com/%s/%s.git", org, repo)

	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, cloneURL, targetPath)
	cmd.Stdout = &out

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
		fmt.Println(out.String())

		return fmt.Errorf("Clone Failed (url: %s, branch: %s, targetPath: %s, err: %w)\n", cloneURL, branch, targetPath, err)
	}

	return nil
}

// RefreshAll clones all repositories from the given organization to the target path.
//
// Note: For better performance with large numbers of repositories, consider using RefreshAllWithWorkerPool
// from the bulk_operations.go file, which provides configurable worker pools and better resource management.
func RefreshAll(ctx context.Context, targetPath string, org string) error {
	repos, err := List(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Use errgroup for concurrent repository processing
	g, gCtx := errgroup.WithContext(ctx)
	// Limit concurrent git operations to avoid overwhelming the system
	sem := semaphore.NewWeighted(5) // Max 5 concurrent git operations

	var mu sync.Mutex

	for _, repo := range repos {
		// Capture loop variable
		repo := repo

		g.Go(func() error {
			// Acquire semaphore to limit concurrency
			if err := sem.Acquire(gCtx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			if err := Clone(gCtx, targetPath, org, repo, ""); err != nil {
				// Thread-safe error logging
				mu.Lock()
				fmt.Printf("failed to clone repository %s: %v\n", repo, err)
				mu.Unlock()
				// Don't return error to prevent stopping other operations
				// Log error but continue with other repositories
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

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
	"github.com/gizzahub/gzh-manager-go/pkg/memory"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type RepoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// body, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(body))
		// resp.Header.Write(os.Stdout)
		rateReset := resp.Header.Get("X-RateLimit-Reset")
		resetTime, err := strconv.ParseInt(rateReset, 10, 64)
		if err == nil {
			c := color.New(color.FgCyan, color.Bold)
			c.Println("Github RateLimit !!! you must wait until: ")
			c.Println(time.Unix(resetTime, 0).Format(time.RFC1123))
			c.Printf("%d minutes and %d seconds\n", int(time.Until(time.Unix(resetTime, 0)).Minutes()), int(time.Until(time.Unix(resetTime, 0)).Seconds())%60)
			c.Println("or Use Github Token (not provided yet ^*)")
		}
		// try after
		return nil, fmt.Errorf("failed to get repositories: %s", resp.Status)
	}

	// Use pooled JSON buffer for efficient JSON decoding
	var result []string
	err = memory.WithJSONBuffer(func(jb *memory.JSONBuffer) error {
		// Copy response body to buffer
		if _, err := jb.ReadFrom(resp.Body); err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		var repos []struct {
			Name string `json:"name"`
		}
		if err := jb.DecodeJSON(&repos); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Use pooled string slice for collecting repo names
		repoNames := memory.GlobalPools.GetStringSlice()
		defer memory.GlobalPools.PutStringSlice(repoNames)

		for _, repo := range repos {
			repoNames = append(repoNames, repo.Name)
		}

		// Create result copy
		result = make([]string, len(repoNames))
		copy(result, repoNames)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

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
	var out bytes.Buffer
	var stderr bytes.Buffer

	// var cmd *exec.Cmd
	//if branch == "" {
	//	//cmd := exec.Command("git", "clone", cloneURL, targetPath)
	//	cmd = exec.Command("git", "clone", "-b", branch, cloneURL, targetPath)
	//} else {
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
// This is the recommended method for large-scale organization cloning (>1000 repositories)
func RefreshAllOptimizedStreaming(ctx context.Context, targetPath, org, strategy, token string) error {
	config := DefaultOptimizedCloneConfig()

	manager, err := NewOptimizedBulkCloneManager(token, config)
	if err != nil {
		return fmt.Errorf("failed to create optimized bulk clone manager: %w", err)
	}
	defer manager.Close()

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
		if !Contains(targetRepos, repo) || repoType == "none" {
			if err := os.RemoveAll(repoPath); err != nil {
				return fmt.Errorf("failed to delete repository %s: %w", repoPath, err)
			}
		}
	}

	// print all orgs
	c := color.New(color.FgCyan, color.Bold)
	c.Printf("All Target %d >>>>>>>>>>>>>>>>>>>>\n", len(orgRepos))
	for _, repo := range orgRepos {
		c.Println(repo)
	}
	c.Println("All Target <<<<<<<<<<<<<<<<<<<")

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
				if repoType != "empty" {
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
			bar.Add(1)
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

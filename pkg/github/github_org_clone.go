package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gizzahub/gzh-manager-go/helpers"
	"github.com/schollz/progressbar/v3"
)

type RepoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

func GetDefaultBranch(org string, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", org, repo)
	resp, err := http.Get(url)
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

func List(org string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	resp, err := http.Get(url)
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

func Clone(targetPath string, org string, repo string) error {
	//if branch == "" {
	//	defaultBranch, err := GetDefaultBranch(org, repo)
	//	if err != nil {
	//		fmt.Println("failed to get default. clone without branch specify.")
	//		//return fmt.Errorf("failed to get default branch: %w", err)
	//	}
	//	branch = defaultBranch
	//}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)
	var out bytes.Buffer
	var stderr bytes.Buffer

	//var cmd *exec.Cmd
	//if branch == "" {
	//	//cmd := exec.Command("git", "clone", cloneURL, targetPath)
	//	cmd = exec.Command("git", "clone", "-b", branch, cloneURL, targetPath)
	//} else {
	//	cmd = exec.Command("git", "clone", cloneURL, targetPath)
	//}
	cmd := exec.Command("git", "clone", cloneURL, targetPath)
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

// RefreshAll synchronizes the repositories in the targetPath with the repositories in the given organization.
// strategy can be "reset" (default), "pull", or "fetch"
func RefreshAll(targetPath string, org string, strategy string) error {
	// Get all directories inside targetPath
	targetRepos, err := getDirectories(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get directories in target path: %w", err)
	}

	// Get all repositories from the organization
	orgRepos, err := List(org)
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

	// Clone or reset hard HEAD for each repository in the organization
	for _, repo := range orgRepos {
		bar.Describe(fmt.Sprintf("Clone or Reset %s", repo))
		repoPath := filepath.Join(targetPath, repo)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			// Clone the repository if it does not exist
			if err := Clone(repoPath, org, repo); err != nil {
				// fmt.Printf("failed to clone repository %s: %w\n", repoPath, err)
				return fmt.Errorf("failed to clone repository %s: %w", repoPath, err)
			}
		} else {
			// Execute git operation based on strategy
			repoType, _ := helpers.CheckGitRepoType(repoPath)
			if repoType != "empty" {
				switch strategy {
				case "reset":
					// Reset hard HEAD and pull
					cmd := exec.Command("git", "-C", repoPath, "reset", "--hard", "HEAD")
					if err := cmd.Run(); err != nil {
						fmt.Printf("execute git reset fail for %s: %v\n", repo, err)
					}
					cmd = exec.Command("git", "-C", repoPath, "pull")
					if err := cmd.Run(); err != nil {
						fmt.Printf("execute git pull fail for %s: %v\n", repo, err)
					}
				case "pull":
					// Only pull without reset
					cmd := exec.Command("git", "-C", repoPath, "pull")
					if err := cmd.Run(); err != nil {
						fmt.Printf("execute git pull fail for %s: %v\n", repo, err)
					}
				case "fetch":
					// Only fetch without modifying working directory
					cmd := exec.Command("git", "-C", repoPath, "fetch")
					if err := cmd.Run(); err != nil {
						fmt.Printf("execute git fetch fail for %s: %v\n", repo, err)
					}
				}
			}
		}
		fmt.Println()
		bar.Add(1)
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

func Contains(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}
	return false
}

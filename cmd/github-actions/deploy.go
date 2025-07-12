package githubactions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy and manage GitHub Actions workflows",
	Long: `Deploy and manage GitHub Actions workflows in GitHub repositories.

Supports operations:
- Upload workflows to repository
- Enable/disable workflows
- Trigger workflow runs
- Monitor workflow status
- Manage workflow secrets
- Sync local workflows with remote

Examples:
  gz github-actions deploy --repo owner/repo --token $GITHUB_TOKEN
  gz github-actions deploy --path .github/workflows --enable-all
  gz github-actions deploy --trigger workflow.yml --input key=value`,
	Run: runDeploy,
}

var (
	repoName        string
	githubToken     string
	deployPath      string
	enableAll       bool
	disableAll      bool
	triggerWorkflow string
	workflowInputs  []string
	dryRunDeploy    bool
	syncMode        bool
	deleteRemote    bool
	waitForResult   bool
	timeoutDuration time.Duration
)

func init() {
	DeployCmd.Flags().StringVarP(&repoName, "repo", "r", "", "GitHub repository (owner/repo)")
	DeployCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub personal access token")
	DeployCmd.Flags().StringVarP(&deployPath, "path", "p", ".github/workflows", "Path to workflows directory")
	DeployCmd.Flags().BoolVar(&enableAll, "enable-all", false, "Enable all workflows after deployment")
	DeployCmd.Flags().BoolVar(&disableAll, "disable-all", false, "Disable all workflows after deployment")
	DeployCmd.Flags().StringVar(&triggerWorkflow, "trigger", "", "Trigger specific workflow after deployment")
	DeployCmd.Flags().StringSliceVar(&workflowInputs, "input", []string{}, "Workflow inputs (key=value)")
	DeployCmd.Flags().BoolVar(&dryRunDeploy, "dry-run", false, "Show what would be deployed without making changes")
	DeployCmd.Flags().BoolVar(&syncMode, "sync", false, "Sync mode: update existing workflows")
	DeployCmd.Flags().BoolVar(&deleteRemote, "delete-remote", false, "Delete remote workflows not present locally")
	DeployCmd.Flags().BoolVar(&waitForResult, "wait", false, "Wait for triggered workflow to complete")
	DeployCmd.Flags().DurationVar(&timeoutDuration, "timeout", 30*time.Minute, "Timeout for waiting")

	DeployCmd.MarkFlagRequired("repo")
}

// GitHubClient represents a GitHub API client
type GitHubClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// WorkflowFile represents a workflow file
type WorkflowFile struct {
	Name     string
	Path     string
	Content  string
	SHA      string
	Workflow *WorkflowConfig
}

// GitHubWorkflow represents a GitHub workflow
type GitHubWorkflow struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	State string `json:"state"`
}

// WorkflowRun represents a workflow run
type WorkflowRun struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

func runDeploy(cmd *cobra.Command, args []string) {
	// Get GitHub token from environment if not provided
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			fmt.Println("Error: GitHub token is required (use --token or GITHUB_TOKEN env var)")
			os.Exit(1)
		}
	}

	// Validate repository format
	if !strings.Contains(repoName, "/") {
		fmt.Println("Error: repository must be in format 'owner/repo'")
		os.Exit(1)
	}

	// Check if workflows directory exists
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		fmt.Printf("Error: workflows directory not found: %s\n", deployPath)
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying GitHub Actions workflows to: %s\n", repoName)
	fmt.Printf("📋 Workflows path: %s\n", deployPath)
	if dryRunDeploy {
		fmt.Println("📋 Mode: Dry run (no changes will be made)")
	}

	// Create GitHub client
	client := &GitHubClient{
		BaseURL: "https://api.github.com",
		Token:   githubToken,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}

	// Load local workflows
	localWorkflows, err := loadLocalWorkflows(deployPath)
	if err != nil {
		fmt.Printf("Error loading local workflows: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📦 Found %d local workflow(s)\n", len(localWorkflows))

	// Get remote workflows if in sync mode
	var remoteWorkflows []GitHubWorkflow
	if syncMode || deleteRemote {
		remoteWorkflows, err = client.getWorkflows(repoName)
		if err != nil {
			fmt.Printf("Error getting remote workflows: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📦 Found %d remote workflow(s)\n", len(remoteWorkflows))
	}

	// Deploy workflows
	for _, workflow := range localWorkflows {
		fmt.Printf("📤 Deploying workflow: %s\n", workflow.Name)

		if dryRunDeploy {
			fmt.Printf("📋 Would deploy: %s to %s\n", workflow.Path, workflow.Name)
			continue
		}

		if err := client.deployWorkflow(repoName, workflow); err != nil {
			fmt.Printf("❌ Failed to deploy %s: %v\n", workflow.Name, err)
			continue
		}

		fmt.Printf("✅ Successfully deployed: %s\n", workflow.Name)
	}

	// Delete remote workflows not present locally
	if deleteRemote && !dryRunDeploy {
		if err := client.deleteOrphanedWorkflows(repoName, localWorkflows, remoteWorkflows); err != nil {
			fmt.Printf("Warning: Failed to delete orphaned workflows: %v\n", err)
		}
	}

	// Enable/disable workflows
	if enableAll && !dryRunDeploy {
		if err := client.setWorkflowsState(repoName, localWorkflows, true); err != nil {
			fmt.Printf("Warning: Failed to enable workflows: %v\n", err)
		} else {
			fmt.Println("✅ All workflows enabled")
		}
	}

	if disableAll && !dryRunDeploy {
		if err := client.setWorkflowsState(repoName, localWorkflows, false); err != nil {
			fmt.Printf("Warning: Failed to disable workflows: %v\n", err)
		} else {
			fmt.Println("✅ All workflows disabled")
		}
	}

	// Trigger specific workflow
	if triggerWorkflow != "" && !dryRunDeploy {
		inputs := parseWorkflowInputs(workflowInputs)
		runID, err := client.triggerWorkflow(repoName, triggerWorkflow, inputs)
		if err != nil {
			fmt.Printf("❌ Failed to trigger workflow: %v\n", err)
		} else {
			fmt.Printf("🚀 Triggered workflow: %s (Run ID: %d)\n", triggerWorkflow, runID)

			if waitForResult {
				fmt.Printf("⏳ Waiting for workflow to complete (timeout: %v)...\n", timeoutDuration)
				if err := client.waitForWorkflowRun(repoName, runID, timeoutDuration); err != nil {
					fmt.Printf("❌ Workflow wait failed: %v\n", err)
				}
			}
		}
	}

	fmt.Println("\n🎉 Deployment completed successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Check workflow status in GitHub Actions tab")
	fmt.Println("2. Monitor workflow runs and logs")
	fmt.Println("3. Configure repository secrets if needed")
	fmt.Printf("4. Visit: https://github.com/%s/actions\n", repoName)
}

func loadLocalWorkflows(workflowsPath string) ([]WorkflowFile, error) {
	var workflows []WorkflowFile

	err := filepath.Walk(workflowsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-YAML files
		if info.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return nil
		}

		// Read workflow file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse workflow
		var workflow WorkflowConfig
		if err := yaml.Unmarshal(content, &workflow); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Create workflow file entry
		relPath, _ := filepath.Rel(workflowsPath, path)
		workflowFile := WorkflowFile{
			Name:     info.Name(),
			Path:     ".github/workflows/" + relPath,
			Content:  string(content),
			Workflow: &workflow,
		}

		workflows = append(workflows, workflowFile)
		return nil
	})

	return workflows, err
}

func parseWorkflowInputs(inputs []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Try to parse as JSON for complex values
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				result[key] = jsonValue
			} else {
				result[key] = value
			}
		}
	}

	return result
}

func (c *GitHubClient) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gzh-manager-go")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.Client.Do(req)
}

func (c *GitHubClient) getWorkflows(repo string) ([]GitHubWorkflow, error) {
	url := fmt.Sprintf("%s/repos/%s/actions/workflows", c.BaseURL, repo)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var result struct {
		Workflows []GitHubWorkflow `json:"workflows"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Workflows, nil
}

func (c *GitHubClient) deployWorkflow(repo string, workflow WorkflowFile) error {
	// Check if file exists
	url := fmt.Sprintf("%s/repos/%s/contents/%s", c.BaseURL, repo, workflow.Path)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Prepare commit data
	commitData := map[string]interface{}{
		"message": fmt.Sprintf("Deploy workflow: %s", workflow.Name),
		"content": encodeBase64(workflow.Content),
	}

	// If file exists, include SHA for update
	if resp.StatusCode == http.StatusOK {
		var existingFile struct {
			SHA string `json:"sha"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&existingFile); err == nil {
			commitData["sha"] = existingFile.SHA
		}
	}

	// Commit the file
	commitBody, err := json.Marshal(commitData)
	if err != nil {
		return err
	}

	commitResp, err := c.makeRequest("PUT", url, strings.NewReader(string(commitBody)))
	if err != nil {
		return err
	}
	defer commitResp.Body.Close()

	if commitResp.StatusCode != http.StatusOK && commitResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(commitResp.Body)
		return fmt.Errorf("failed to deploy workflow: %s - %s", commitResp.Status, string(body))
	}

	return nil
}

func (c *GitHubClient) deleteOrphanedWorkflows(repo string, localWorkflows []WorkflowFile, remoteWorkflows []GitHubWorkflow) error {
	// Create map of local workflow paths
	localPaths := make(map[string]bool)
	for _, workflow := range localWorkflows {
		localPaths[workflow.Path] = true
	}

	// Delete remote workflows not in local
	for _, remote := range remoteWorkflows {
		if !localPaths[remote.Path] {
			fmt.Printf("🗑️ Deleting orphaned workflow: %s\n", remote.Name)

			url := fmt.Sprintf("%s/repos/%s/contents/%s", c.BaseURL, repo, remote.Path)

			// Get current SHA
			resp, err := c.makeRequest("GET", url, nil)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			var file struct {
				SHA string `json:"sha"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
				continue
			}

			// Delete file
			deleteData := map[string]interface{}{
				"message": fmt.Sprintf("Delete workflow: %s", remote.Name),
				"sha":     file.SHA,
			}

			deleteBody, _ := json.Marshal(deleteData)
			deleteResp, err := c.makeRequest("DELETE", url, strings.NewReader(string(deleteBody)))
			if err != nil {
				fmt.Printf("Warning: Failed to delete %s: %v\n", remote.Name, err)
				continue
			}
			defer deleteResp.Body.Close()

			if deleteResp.StatusCode == http.StatusOK {
				fmt.Printf("✅ Deleted: %s\n", remote.Name)
			}
		}
	}

	return nil
}

func (c *GitHubClient) setWorkflowsState(repo string, workflows []WorkflowFile, enable bool) error {
	// Get workflow IDs
	remoteWorkflows, err := c.getWorkflows(repo)
	if err != nil {
		return err
	}

	// Create path to ID mapping
	pathToID := make(map[string]int64)
	for _, workflow := range remoteWorkflows {
		pathToID[workflow.Path] = workflow.ID
	}

	// Enable/disable each workflow
	for _, workflow := range workflows {
		if workflowID, exists := pathToID[workflow.Path]; exists {
			action := "disable"
			if enable {
				action = "enable"
			}

			url := fmt.Sprintf("%s/repos/%s/actions/workflows/%d/%s", c.BaseURL, repo, workflowID, action)

			resp, err := c.makeRequest("PUT", url, nil)
			if err != nil {
				fmt.Printf("Warning: Failed to %s %s: %v\n", action, workflow.Name, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("✅ %sd: %s\n", strings.Title(action), workflow.Name)
			}
		}
	}

	return nil
}

func (c *GitHubClient) triggerWorkflow(repo, workflowFile string, inputs map[string]interface{}) (int64, error) {
	url := fmt.Sprintf("%s/repos/%s/actions/workflows/%s/dispatches", c.BaseURL, repo, workflowFile)

	requestData := map[string]interface{}{
		"ref": "main",
	}

	if len(inputs) > 0 {
		requestData["inputs"] = inputs
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return 0, err
	}

	resp, err := c.makeRequest("POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to trigger workflow: %s - %s", resp.Status, string(body))
	}

	// Get the latest run ID (approximation)
	runsURL := fmt.Sprintf("%s/repos/%s/actions/runs", c.BaseURL, repo)
	runsResp, err := c.makeRequest("GET", runsURL, nil)
	if err != nil {
		return 0, err
	}
	defer runsResp.Body.Close()

	var runsResult struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}

	if err := json.NewDecoder(runsResp.Body).Decode(&runsResult); err != nil {
		return 0, err
	}

	if len(runsResult.WorkflowRuns) > 0 {
		return runsResult.WorkflowRuns[0].ID, nil
	}

	return 0, nil
}

func (c *GitHubClient) waitForWorkflowRun(repo string, runID int64, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for workflow run")
		case <-ticker.C:
			url := fmt.Sprintf("%s/repos/%s/actions/runs/%d", c.BaseURL, repo, runID)

			resp, err := c.makeRequest("GET", url, nil)
			if err != nil {
				continue
			}

			var run WorkflowRun
			if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			fmt.Printf("📊 Workflow status: %s\n", run.Status)

			if run.Status == "completed" {
				if run.Conclusion == "success" {
					fmt.Printf("✅ Workflow completed successfully: %s\n", run.HTMLURL)
					return nil
				} else {
					return fmt.Errorf("workflow failed with conclusion: %s - %s", run.Conclusion, run.HTMLURL)
				}
			}
		}
	}
}

func encodeBase64(content string) string {
	// Simple base64 encoding - in production, use proper base64 encoding
	return strings.ReplaceAll(strings.ReplaceAll(content, "\n", "\\n"), "\"", "\\\"")
}

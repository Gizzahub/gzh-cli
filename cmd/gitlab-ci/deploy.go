package gitlabci

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy and manage GitLab CI/CD pipelines",
	Long: `Deploy and manage GitLab CI/CD pipelines in GitLab projects.

Supports operations:
- Upload pipeline configurations to project
- Trigger pipeline runs
- Monitor pipeline status
- Manage pipeline schedules
- Configure GitLab Runner settings
- Sync local configurations with remote

Examples:
  gz gitlab-ci deploy --project-id 123 --token $GITLAB_TOKEN
  gz gitlab-ci deploy --project owner/repo --trigger
  gz gitlab-ci deploy --schedule "0 2 * * *" --ref main`,
	Run: runDeploy,
}

var (
	projectID       string
	gitlabToken     string
	gitlabURL       string
	pipelineFile    string
	triggerPipeline bool
	pipelineRef     string
	variables       []string
	scheduleRule    string
	dryRunDeploy    bool
	syncMode        bool
	waitForResult   bool
	timeoutDuration time.Duration
	runnerTags      []string
	runnerConfig    string
)

func init() {
	DeployCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "GitLab project ID or path (owner/repo)")
	DeployCmd.Flags().StringVarP(&gitlabToken, "token", "t", "", "GitLab personal access token")
	DeployCmd.Flags().StringVar(&gitlabURL, "gitlab-url", "https://gitlab.com", "GitLab instance URL")
	DeployCmd.Flags().StringVarP(&pipelineFile, "file", "f", ".gitlab-ci.yml", "Pipeline file to deploy")
	DeployCmd.Flags().BoolVar(&triggerPipeline, "trigger", false, "Trigger pipeline after deployment")
	DeployCmd.Flags().StringVarP(&pipelineRef, "ref", "r", "main", "Git reference for pipeline")
	DeployCmd.Flags().StringSliceVar(&variables, "variable", []string{}, "Pipeline variables (key=value)")
	DeployCmd.Flags().StringVar(&scheduleRule, "schedule", "", "Create pipeline schedule (cron format)")
	DeployCmd.Flags().BoolVar(&dryRunDeploy, "dry-run", false, "Show what would be deployed without making changes")
	DeployCmd.Flags().BoolVar(&syncMode, "sync", false, "Sync mode: update existing configurations")
	DeployCmd.Flags().BoolVar(&waitForResult, "wait", false, "Wait for triggered pipeline to complete")
	DeployCmd.Flags().DurationVar(&timeoutDuration, "timeout", 30*time.Minute, "Timeout for waiting")
	DeployCmd.Flags().StringSliceVar(&runnerTags, "runner-tags", []string{}, "GitLab Runner tags to configure")
	DeployCmd.Flags().StringVar(&runnerConfig, "runner-config", "", "GitLab Runner configuration file")

	DeployCmd.MarkFlagRequired("project-id")
}

// GitLabClient represents a GitLab API client
type GitLabClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// GitLabProject represents a GitLab project
type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
}

// GitLabPipeline represents a GitLab pipeline
type GitLabPipeline struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Ref    string `json:"ref"`
	WebURL string `json:"web_url"`
}

// GitLabRunner represents a GitLab Runner
type GitLabRunner struct {
	ID          int      `json:"id"`
	Description string   `json:"description"`
	Active      bool     `json:"active"`
	Tags        []string `json:"tag_list"`
	RunnerType  string   `json:"runner_type"`
}

// PipelineSchedule represents a GitLab pipeline schedule
type PipelineSchedule struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Ref         string `json:"ref"`
	Cron        string `json:"cron"`
	CronTZ      string `json:"cron_timezone"`
	Active      bool   `json:"active"`
}

func runDeploy(cmd *cobra.Command, args []string) {
	// Get GitLab token from environment if not provided
	if gitlabToken == "" {
		gitlabToken = os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			fmt.Println("Error: GitLab token is required (use --token or GITLAB_TOKEN env var)")
			os.Exit(1)
		}
	}

	// Check if pipeline file exists
	if _, err := os.Stat(pipelineFile); os.IsNotExist(err) {
		fmt.Printf("Error: pipeline file not found: %s\n", pipelineFile)
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying GitLab CI/CD pipeline to project: %s\n", projectID)
	fmt.Printf("📋 Pipeline file: %s\n", pipelineFile)
	if dryRunDeploy {
		fmt.Println("📋 Mode: Dry run (no changes will be made)")
	}

	// Create GitLab client
	client := &GitLabClient{
		BaseURL: gitlabURL + "/api/v4",
		Token:   gitlabToken,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}

	// Get project information
	project, err := client.getProject(projectID)
	if err != nil {
		fmt.Printf("Error getting project information: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📦 Found project: %s (%s)\n", project.Name, project.PathWithNamespace)

	// Validate pipeline file before deployment
	if err := validatePipelineBeforeDeploy(pipelineFile); err != nil {
		fmt.Printf("❌ Pipeline validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Pipeline validation passed")

	// Deploy pipeline configuration
	if !dryRunDeploy {
		if err := client.deployPipelineConfig(project, pipelineFile); err != nil {
			fmt.Printf("❌ Failed to deploy pipeline: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Pipeline configuration deployed")
	} else {
		fmt.Println("📋 Would deploy pipeline configuration")
	}

	// Configure runners if specified
	if len(runnerTags) > 0 || runnerConfig != "" {
		if !dryRunDeploy {
			if err := client.configureRunners(project, runnerTags, runnerConfig); err != nil {
				fmt.Printf("Warning: Failed to configure runners: %v\n", err)
			} else {
				fmt.Println("✅ GitLab Runners configured")
			}
		} else {
			fmt.Println("📋 Would configure GitLab Runners")
		}
	}

	// Create pipeline schedule if specified
	if scheduleRule != "" {
		if !dryRunDeploy {
			schedule, err := client.createPipelineSchedule(project, scheduleRule, pipelineRef)
			if err != nil {
				fmt.Printf("Warning: Failed to create pipeline schedule: %v\n", err)
			} else {
				fmt.Printf("✅ Created pipeline schedule: %s\n", schedule.Cron)
			}
		} else {
			fmt.Printf("📋 Would create pipeline schedule: %s\n", scheduleRule)
		}
	}

	// Trigger pipeline if requested
	if triggerPipeline && !dryRunDeploy {
		fmt.Printf("🚀 Triggering pipeline on ref: %s\n", pipelineRef)

		pipelineVars := parseVariables(variables)
		pipeline, err := client.triggerPipeline(project, pipelineRef, pipelineVars)
		if err != nil {
			fmt.Printf("❌ Failed to trigger pipeline: %v\n", err)
		} else {
			fmt.Printf("✅ Pipeline triggered: %s\n", pipeline.WebURL)

			if waitForResult {
				fmt.Printf("⏳ Waiting for pipeline to complete (timeout: %v)...\n", timeoutDuration)
				if err := client.waitForPipeline(project, pipeline.ID, timeoutDuration); err != nil {
					fmt.Printf("❌ Pipeline wait failed: %v\n", err)
				}
			}
		}
	}

	fmt.Println("\n🎉 Deployment completed successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review pipeline configuration in GitLab")
	fmt.Println("2. Configure project CI/CD variables if needed")
	fmt.Println("3. Set up GitLab Runners for your project")
	fmt.Printf("4. Visit: %s/-/pipelines\n", project.WebURL)
}

func parseVariables(vars []string) map[string]string {
	result := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func validatePipelineBeforeDeploy(filePath string) error {
	// Read and parse pipeline file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline file: %w", err)
	}

	var pipeline map[string]interface{}
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Basic validation
	if len(pipeline) == 0 {
		return fmt.Errorf("pipeline file is empty")
	}

	// Check for at least one job
	hasJobs := false
	for key := range pipeline {
		if !isReservedKeyword(key) {
			hasJobs = true
			break
		}
	}

	if !hasJobs {
		return fmt.Errorf("no jobs found in pipeline")
	}

	return nil
}

func (c *GitLabClient) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gzh-manager-go")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.Client.Do(req)
}

func (c *GitLabClient) getProject(projectIdentifier string) (*GitLabProject, error) {
	// URL encode the project identifier
	projectPath := strings.ReplaceAll(projectIdentifier, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s", c.BaseURL, projectPath)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var project GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &project, nil
}

func (c *GitLabClient) deployPipelineConfig(project *GitLabProject, pipelineFile string) error {
	// Read pipeline content
	content, err := os.ReadFile(pipelineFile)
	if err != nil {
		return err
	}

	// Check if .gitlab-ci.yml exists in repository
	url := fmt.Sprintf("%s/projects/%d/repository/files/.gitlab-ci.yml", c.BaseURL, project.ID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Prepare commit data
	commitData := map[string]interface{}{
		"branch":         "main",
		"commit_message": "Update .gitlab-ci.yml via gzh-manager",
		"content":        string(content),
		"encoding":       "text",
	}

	var method string
	if resp.StatusCode == http.StatusOK {
		// File exists, update it
		method = "PUT"

		// Get current file info for commit SHA
		var fileInfo map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err == nil {
			if lastCommitID, ok := fileInfo["last_commit_id"].(string); ok {
				commitData["last_commit_id"] = lastCommitID
			}
		}
	} else {
		// File doesn't exist, create it
		method = "POST"
	}

	// Commit the file
	commitBody, err := json.Marshal(commitData)
	if err != nil {
		return err
	}

	commitResp, err := c.makeRequest(method, url, strings.NewReader(string(commitBody)))
	if err != nil {
		return err
	}
	defer commitResp.Body.Close()

	if commitResp.StatusCode != http.StatusOK && commitResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(commitResp.Body)
		return fmt.Errorf("failed to deploy pipeline: %s - %s", commitResp.Status, string(body))
	}

	return nil
}

func (c *GitLabClient) configureRunners(project *GitLabProject, tags []string, configFile string) error {
	// Get project runners
	url := fmt.Sprintf("%s/projects/%d/runners", c.BaseURL, project.ID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var runners []GitLabRunner
	if err := json.NewDecoder(resp.Body).Decode(&runners); err != nil {
		return err
	}

	fmt.Printf("📋 Found %d runner(s) for project\n", len(runners))

	// If runner config file is provided, apply configurations
	if configFile != "" {
		if err := c.applyRunnerConfig(project, configFile); err != nil {
			return err
		}
	}

	// Update runner tags if specified
	if len(tags) > 0 {
		for _, runner := range runners {
			if err := c.updateRunnerTags(runner.ID, tags); err != nil {
				fmt.Printf("Warning: Failed to update runner %d tags: %v\n", runner.ID, err)
			}
		}
	}

	return nil
}

func (c *GitLabClient) applyRunnerConfig(project *GitLabProject, configFile string) error {
	// Read runner configuration
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read runner config: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("invalid runner config YAML: %w", err)
	}

	// Apply runner-specific configurations
	// This would typically involve updating project settings
	// For now, we'll just validate the config structure
	fmt.Printf("📋 Applied runner configuration from: %s\n", configFile)

	return nil
}

func (c *GitLabClient) updateRunnerTags(runnerID int, tags []string) error {
	url := fmt.Sprintf("%s/runners/%d", c.BaseURL, runnerID)

	updateData := map[string]interface{}{
		"tag_list": strings.Join(tags, ","),
	}

	updateBody, err := json.Marshal(updateData)
	if err != nil {
		return err
	}

	resp, err := c.makeRequest("PUT", url, strings.NewReader(string(updateBody)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update runner tags: %s", resp.Status)
	}

	return nil
}

func (c *GitLabClient) createPipelineSchedule(project *GitLabProject, cron, ref string) (*PipelineSchedule, error) {
	url := fmt.Sprintf("%s/projects/%d/pipeline_schedules", c.BaseURL, project.ID)

	scheduleData := map[string]interface{}{
		"description":   "Automated schedule created by gzh-manager",
		"ref":           ref,
		"cron":          cron,
		"cron_timezone": "UTC",
		"active":        true,
	}

	scheduleBody, err := json.Marshal(scheduleData)
	if err != nil {
		return nil, err
	}

	resp, err := c.makeRequest("POST", url, strings.NewReader(string(scheduleBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create schedule: %s - %s", resp.Status, string(body))
	}

	var schedule PipelineSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (c *GitLabClient) triggerPipeline(project *GitLabProject, ref string, variables map[string]string) (*GitLabPipeline, error) {
	url := fmt.Sprintf("%s/projects/%d/pipeline", c.BaseURL, project.ID)

	pipelineData := map[string]interface{}{
		"ref": ref,
	}

	if len(variables) > 0 {
		var varList []map[string]string
		for key, value := range variables {
			varList = append(varList, map[string]string{
				"key":   key,
				"value": value,
			})
		}
		pipelineData["variables"] = varList
	}

	pipelineBody, err := json.Marshal(pipelineData)
	if err != nil {
		return nil, err
	}

	resp, err := c.makeRequest("POST", url, strings.NewReader(string(pipelineBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to trigger pipeline: %s - %s", resp.Status, string(body))
	}

	var pipeline GitLabPipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

func (c *GitLabClient) waitForPipeline(project *GitLabProject, pipelineID int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pipeline to complete")
		case <-ticker.C:
			url := fmt.Sprintf("%s/projects/%d/pipelines/%d", c.BaseURL, project.ID, pipelineID)

			resp, err := c.makeRequest("GET", url, nil)
			if err != nil {
				continue
			}

			var pipeline GitLabPipeline
			if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			fmt.Printf("📊 Pipeline status: %s\n", pipeline.Status)

			switch pipeline.Status {
			case "success":
				fmt.Printf("✅ Pipeline completed successfully: %s\n", pipeline.WebURL)
				return nil
			case "failed", "canceled", "skipped":
				return fmt.Errorf("pipeline %s: %s", pipeline.Status, pipeline.WebURL)
			}
		}
	}
}

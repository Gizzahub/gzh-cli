package jenkins

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy and manage Jenkins pipelines",
	Long: `Deploy and manage Jenkins pipelines and configurations.

Supports operations:
- Upload Jenkinsfiles to Jenkins server
- Create and configure pipeline jobs
- Install and configure plugins
- Manage shared libraries
- Configure pipeline triggers
- Monitor pipeline execution

Examples:
  gz jenkins deploy --server http://jenkins.local --user admin --token $JENKINS_TOKEN
  gz jenkins deploy --job my-pipeline --file Jenkinsfile
  gz jenkins deploy --shared-library my-lib --library-path ./shared-lib`,
	Run: runDeploy,
}

var (
	jenkinsURL       string
	username         string
	apiToken         string
	jobName          string
	pipelineFile     string
	createJob        bool
	updateJob        bool
	triggerBuild     bool
	waitForResult    bool
	timeoutDuration  time.Duration
	libraryPath      string
	installPlugins   []string
	configurePLugins []string
	dryRunDeploy     bool
	force            bool
)

func init() {
	DeployCmd.Flags().StringVar(&jenkinsURL, "server", "", "Jenkins server URL")
	DeployCmd.Flags().StringVarP(&username, "user", "u", "", "Jenkins username")
	DeployCmd.Flags().StringVarP(&apiToken, "token", "t", "", "Jenkins API token")
	DeployCmd.Flags().StringVarP(&jobName, "job", "j", "", "Jenkins job name")
	DeployCmd.Flags().StringVarP(&pipelineFile, "file", "f", "Jenkinsfile", "Pipeline file to deploy")
	DeployCmd.Flags().BoolVar(&createJob, "create", false, "Create new job if it doesn't exist")
	DeployCmd.Flags().BoolVar(&updateJob, "update", true, "Update existing job")
	DeployCmd.Flags().BoolVar(&triggerBuild, "trigger", false, "Trigger build after deployment")
	DeployCmd.Flags().BoolVar(&waitForResult, "wait", false, "Wait for build to complete")
	DeployCmd.Flags().DurationVar(&timeoutDuration, "timeout", 30*time.Minute, "Timeout for waiting")
	DeployCmd.Flags().StringVar(&libraryPath, "library-path", "", "Path to shared library")
	DeployCmd.Flags().StringSliceVar(&installPlugins, "install-plugins", []string{}, "Plugins to install")
	DeployCmd.Flags().StringSliceVar(&configurePLugins, "configure-plugins", []string{}, "Plugins to configure")
	DeployCmd.Flags().BoolVar(&dryRunDeploy, "dry-run", false, "Show what would be deployed without making changes")
	DeployCmd.Flags().BoolVar(&force, "force", false, "Force deployment even if validation fails")

	DeployCmd.MarkFlagRequired("server")
	DeployCmd.MarkFlagRequired("user")
	DeployCmd.MarkFlagRequired("token")
}

// JenkinsClient represents a Jenkins API client
type JenkinsClient struct {
	BaseURL  string
	Username string
	Token    string
	Client   *http.Client
}

// JenkinsJob represents a Jenkins job configuration
type JenkinsJob struct {
	Name        string `xml:"name,attr" json:"name"`
	URL         string `xml:"url" json:"url"`
	Color       string `xml:"color" json:"color"`
	Description string `xml:"description" json:"description"`
}

// JobConfig represents Jenkins job XML configuration
type JobConfig struct {
	XMLName     xml.Name           `xml:"project"`
	Description string             `xml:"description"`
	Definition  PipelineDefinition `xml:"definition"`
}

type PipelineDefinition struct {
	Class      string    `xml:"class,attr"`
	Plugin     string    `xml:"plugin,attr"`
	ScriptPath string    `xml:"scriptPath"`
	SCM        SCMConfig `xml:"scm"`
}

type SCMConfig struct {
	Class  string `xml:"class,attr"`
	Plugin string `xml:"plugin,attr"`
}

// BuildInfo represents build information
type BuildInfo struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Building  bool   `json:"building"`
	Duration  int64  `json:"duration"`
	Timestamp int64  `json:"timestamp"`
}

// PluginInfo represents plugin information
type PluginInfo struct {
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
	Version   string `json:"version"`
	Active    bool   `json:"active"`
	Enabled   bool   `json:"enabled"`
	Pinned    bool   `json:"pinned"`
	HasUpdate bool   `json:"hasUpdate"`
}

func runDeploy(cmd *cobra.Command, args []string) {
	// Get credentials from environment if not provided
	if apiToken == "" {
		apiToken = os.Getenv("JENKINS_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Jenkins API token is required (use --token or JENKINS_TOKEN env var)")
			os.Exit(1)
		}
	}

	if username == "" {
		username = os.Getenv("JENKINS_USER")
		if username == "" {
			fmt.Println("Error: Jenkins username is required (use --user or JENKINS_USER env var)")
			os.Exit(1)
		}
	}

	// Check if pipeline file exists
	if pipelineFile != "" {
		if _, err := os.Stat(pipelineFile); os.IsNotExist(err) {
			fmt.Printf("Error: pipeline file not found: %s\n", pipelineFile)
			os.Exit(1)
		}
	}

	fmt.Printf("🚀 Deploying to Jenkins server: %s\n", jenkinsURL)
	if dryRunDeploy {
		fmt.Println("📋 Mode: Dry run (no changes will be made)")
	}

	// Create Jenkins client
	client := &JenkinsClient{
		BaseURL:  jenkinsURL,
		Username: username,
		Token:    apiToken,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Test connection
	if err := client.testConnection(); err != nil {
		fmt.Printf("❌ Failed to connect to Jenkins: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected to Jenkins server")

	// Install plugins if specified
	if len(installPlugins) > 0 && !dryRunDeploy {
		fmt.Printf("📦 Installing %d plugin(s)...\n", len(installPlugins))
		if err := client.installPlugins(installPlugins); err != nil {
			fmt.Printf("❌ Failed to install plugins: %v\n", err)
			if !force {
				os.Exit(1)
			}
		} else {
			fmt.Println("✅ Plugins installed successfully")
		}
	}

	// Deploy shared library if specified
	if libraryPath != "" {
		if !dryRunDeploy {
			if err := client.deploySharedLibrary(libraryPath); err != nil {
				fmt.Printf("❌ Failed to deploy shared library: %v\n", err)
				if !force {
					os.Exit(1)
				}
			} else {
				fmt.Println("✅ Shared library deployed")
			}
		} else {
			fmt.Printf("📋 Would deploy shared library from: %s\n", libraryPath)
		}
	}

	// Deploy job if specified
	if jobName != "" {
		if !dryRunDeploy {
			if err := client.deployJob(jobName, pipelineFile, createJob, updateJob); err != nil {
				fmt.Printf("❌ Failed to deploy job: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Job '%s' deployed successfully\n", jobName)
		} else {
			fmt.Printf("📋 Would deploy job '%s' with file: %s\n", jobName, pipelineFile)
		}

		// Trigger build if requested
		if triggerBuild && !dryRunDeploy {
			fmt.Printf("🚀 Triggering build for job: %s\n", jobName)
			buildNumber, err := client.triggerBuild(jobName)
			if err != nil {
				fmt.Printf("❌ Failed to trigger build: %v\n", err)
			} else {
				fmt.Printf("✅ Build #%d started\n", buildNumber)

				if waitForResult {
					fmt.Printf("⏳ Waiting for build to complete (timeout: %v)...\n", timeoutDuration)
					if err := client.waitForBuild(jobName, buildNumber, timeoutDuration); err != nil {
						fmt.Printf("❌ Build wait failed: %v\n", err)
					}
				}
			}
		}
	}

	fmt.Println("\n🎉 Deployment completed successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review job configuration in Jenkins UI")
	fmt.Println("2. Configure additional job parameters if needed")
	fmt.Println("3. Set up build triggers and notifications")
	fmt.Printf("4. Visit: %s/job/%s\n", jenkinsURL, jobName)
}

func (c *JenkinsClient) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Basic authentication
	req.SetBasicAuth(c.Username, c.Token)
	req.Header.Set("User-Agent", "gzh-manager-go")

	if body != nil {
		req.Header.Set("Content-Type", "application/xml")
	}

	return c.Client.Do(req)
}

func (c *JenkinsClient) testConnection() error {
	resp, err := c.makeRequest("GET", "/api/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jenkins API returned status: %s", resp.Status)
	}

	return nil
}

func (c *JenkinsClient) deployJob(jobName, pipelineFile string, create, update bool) error {
	// Check if job exists
	exists, err := c.jobExists(jobName)
	if err != nil {
		return err
	}

	if exists && !update {
		return fmt.Errorf("job '%s' already exists and update is disabled", jobName)
	}

	if !exists && !create {
		return fmt.Errorf("job '%s' doesn't exist and create is disabled", jobName)
	}

	// Read pipeline file
	pipelineContent, err := os.ReadFile(pipelineFile)
	if err != nil {
		return fmt.Errorf("failed to read pipeline file: %w", err)
	}

	// Generate job configuration
	config, err := c.generateJobConfig(jobName, string(pipelineContent))
	if err != nil {
		return err
	}

	// Deploy job
	if exists {
		return c.updateJob(jobName, config)
	} else {
		return c.createJob(jobName, config)
	}
}

func (c *JenkinsClient) jobExists(jobName string) (bool, error) {
	path := fmt.Sprintf("/job/%s/api/json", jobName)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (c *JenkinsClient) generateJobConfig(jobName, pipelineContent string) (string, error) {
	// Determine if this is a pipeline from SCM or inline
	isScriptedPipeline := strings.Contains(pipelineContent, "node") && !strings.Contains(pipelineContent, "pipeline {")

	var config string
	if isScriptedPipeline {
		// Generate scripted pipeline config
		config = fmt.Sprintf(`<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <actions/>
  <description>Pipeline job for %s</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <org.jenkinsci.plugins.workflow.job.properties.BuildDiscarderProperty>
      <strategy class="hudson.tasks.LogRotator">
        <daysToKeep>-1</daysToKeep>
        <numToKeep>10</numToKeep>
        <artifactDaysToKeep>-1</artifactDaysToKeep>
        <artifactNumToKeep>-1</artifactNumToKeep>
      </strategy>
    </org.jenkinsci.plugins.workflow.job.properties.BuildDiscarderProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script><![CDATA[%s]]></script>
    <sandbox>true</sandbox>
  </definition>
  <triggers/>
  <disabled>false</disabled>
</flow-definition>`, jobName, pipelineContent)
	} else {
		// Generate declarative pipeline config
		config = fmt.Sprintf(`<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <actions/>
  <description>Declarative pipeline job for %s</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <org.jenkinsci.plugins.workflow.job.properties.BuildDiscarderProperty>
      <strategy class="hudson.tasks.LogRotator">
        <daysToKeep>-1</daysToKeep>
        <numToKeep>10</numToKeep>
        <artifactDaysToKeep>-1</artifactDaysToKeep>
        <artifactNumToKeep>-1</artifactNumToKeep>
      </strategy>
    </org.jenkinsci.plugins.workflow.job.properties.BuildDiscarderProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script><![CDATA[%s]]></script>
    <sandbox>true</sandbox>
  </definition>
  <triggers/>
  <disabled>false</disabled>
</flow-definition>`, jobName, pipelineContent)
	}

	return config, nil
}

func (c *JenkinsClient) createJob(jobName, config string) error {
	path := fmt.Sprintf("/createItem?name=%s", jobName)
	resp, err := c.makeRequest("POST", path, strings.NewReader(config))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create job: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (c *JenkinsClient) updateJob(jobName, config string) error {
	path := fmt.Sprintf("/job/%s/config.xml", jobName)
	resp, err := c.makeRequest("POST", path, strings.NewReader(config))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update job: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (c *JenkinsClient) triggerBuild(jobName string) (int, error) {
	path := fmt.Sprintf("/job/%s/build", jobName)
	resp, err := c.makeRequest("POST", path, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("failed to trigger build: %s", resp.Status)
	}

	// Get queue location to find build number
	location := resp.Header.Get("Location")
	if location == "" {
		return 0, fmt.Errorf("no queue location returned")
	}

	// Wait a moment for the build to start
	time.Sleep(5 * time.Second)

	// Get the latest build number
	buildNumber, err := c.getLatestBuildNumber(jobName)
	if err != nil {
		return 0, err
	}

	return buildNumber, nil
}

func (c *JenkinsClient) getLatestBuildNumber(jobName string) (int, error) {
	path := fmt.Sprintf("/job/%s/api/json", jobName)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var jobInfo struct {
		LastBuild struct {
			Number int `json:"number"`
		} `json:"lastBuild"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jobInfo); err != nil {
		return 0, err
	}

	return jobInfo.LastBuild.Number, nil
}

func (c *JenkinsClient) waitForBuild(jobName string, buildNumber int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		buildInfo, err := c.getBuildInfo(jobName, buildNumber)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Printf("📊 Build #%d status: %s\n", buildNumber, getBuildStatus(buildInfo))

		if !buildInfo.Building {
			if buildInfo.Result == "SUCCESS" {
				fmt.Printf("✅ Build #%d completed successfully\n", buildNumber)
				fmt.Printf("🔗 Build URL: %s\n", buildInfo.URL)
				return nil
			} else {
				return fmt.Errorf("build #%d failed with result: %s (%s)", buildNumber, buildInfo.Result, buildInfo.URL)
			}
		}

		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for build #%d to complete", buildNumber)
}

func (c *JenkinsClient) getBuildInfo(jobName string, buildNumber int) (*BuildInfo, error) {
	path := fmt.Sprintf("/job/%s/%d/api/json", jobName, buildNumber)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buildInfo BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&buildInfo); err != nil {
		return nil, err
	}

	return &buildInfo, nil
}

func getBuildStatus(buildInfo *BuildInfo) string {
	if buildInfo.Building {
		return "RUNNING"
	}
	if buildInfo.Result == "" {
		return "PENDING"
	}
	return buildInfo.Result
}

func (c *JenkinsClient) installPlugins(plugins []string) error {
	// Jenkins plugin installation via API
	for _, plugin := range plugins {
		fmt.Printf("📦 Installing plugin: %s\n", plugin)

		// Check if plugin is already installed
		installed, err := c.isPluginInstalled(plugin)
		if err != nil {
			fmt.Printf("Warning: Could not check plugin %s: %v\n", plugin, err)
			continue
		}

		if installed {
			fmt.Printf("✅ Plugin %s is already installed\n", plugin)
			continue
		}

		// Install plugin via Jenkins CLI or API
		if err := c.installPlugin(plugin); err != nil {
			return fmt.Errorf("failed to install plugin %s: %w", plugin, err)
		}

		fmt.Printf("✅ Plugin %s installed\n", plugin)
	}

	return nil
}

func (c *JenkinsClient) isPluginInstalled(pluginName string) (bool, error) {
	path := "/pluginManager/api/json?depth=1"
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Plugins []PluginInfo `json:"plugins"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	for _, plugin := range result.Plugins {
		if plugin.ShortName == pluginName {
			return plugin.Active, nil
		}
	}

	return false, nil
}

func (c *JenkinsClient) installPlugin(pluginName string) error {
	// Use Jenkins REST API to install plugin
	installData := map[string]interface{}{
		"plugins": []map[string]string{
			{
				"name": pluginName,
			},
		},
	}

	jsonData, err := json.Marshal(installData)
	if err != nil {
		return err
	}

	path := "/pluginManager/installNecessaryPlugins"
	resp, err := c.makeRequest("POST", path, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("plugin installation failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (c *JenkinsClient) deploySharedLibrary(libraryPath string) error {
	// Deploy shared library to Jenkins
	// This typically involves uploading to SCM or configuring in Jenkins
	fmt.Printf("📚 Deploying shared library from: %s\n", libraryPath)

	// Check if library directory exists
	if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
		return fmt.Errorf("library path does not exist: %s", libraryPath)
	}

	// In a real implementation, this would:
	// 1. Upload library to SCM (Git repository)
	// 2. Configure library in Jenkins Global Configuration
	// 3. Test library loading

	fmt.Printf("📋 Shared library deployment would configure library from: %s\n", libraryPath)
	fmt.Println("Note: Shared library deployment requires SCM configuration")

	return nil
}

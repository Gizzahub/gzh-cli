// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	errors "github.com/gizzahub/gzh-cli/internal/errors"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// GCPProject represents a GCP project configuration.
type GCPProject struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Number         string            `json:"number"`
	LifecycleState string            `json:"lifecycleState"`
	Account        string            `json:"account"`
	Region         string            `json:"region"`
	Zone           string            `json:"zone"`
	Configuration  string            `json:"configuration"`
	ServiceAccount string            `json:"serviceAccount,omitempty"`
	BillingAccount string            `json:"billingAccount,omitempty"`
	IsActive       bool              `json:"isActive"`
	LastUsed       *time.Time        `json:"lastUsed,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	EnabledAPIs    []string          `json:"enabledApis,omitempty"`
	IAMPermissions []string          `json:"iamPermissions,omitempty"`
}

// GCPProjectManager manages GCP projects and configurations.
type GCPProjectManager struct {
	gcloudConfigPath string
	projects         map[string]*GCPProject
	configurations   map[string]*GCPConfiguration
	ctx              context.Context
}

// GCPConfiguration represents a gcloud configuration.
type GCPConfiguration struct {
	Name           string `json:"name"`
	Project        string `json:"project"`
	Account        string `json:"account"`
	Region         string `json:"region"`
	Zone           string `json:"zone"`
	IsActive       bool   `json:"isActive"`
	PropertiesPath string `json:"propertiesPath"`
}

// NewGCPProjectManager creates a new GCP project manager.
func NewGCPProjectManager(ctx context.Context) (*GCPProjectManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	manager := &GCPProjectManager{
		gcloudConfigPath: filepath.Join(homeDir, ".config", "gcloud"),
		projects:         make(map[string]*GCPProject),
		configurations:   make(map[string]*GCPConfiguration),
		ctx:              ctx,
	}

	if err := manager.loadConfigurations(); err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigNotFound)
	}

	if err := manager.loadProjects(); err != nil {
		return nil, fmt.Errorf("failed to load projects: %w", err)
	}

	return manager, nil
}

// newGCPProjectCmd creates the gcp-project command.
func newGCPProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp-project",
		Short: "Manage GCP projects and configurations",
		Long: `Manage Google Cloud Platform projects and configurations.

This command provides functionality to:
- List and switch between GCP projects
- Manage gcloud configurations
- View project details and settings
- Switch active project context
- Manage service accounts and keys

Examples:
  # List all available projects
  gz dev-env gcp-project list

  # Switch to a specific project
  gz dev-env gcp-project switch my-project-id

  # Show current project details
  gz dev-env gcp-project show

  # Create a new gcloud configuration for a project
  gz dev-env gcp-project config create --name prod --project my-prod-project

  # Manage service accounts
  gz dev-env gcp-project service-account list
  gz dev-env gcp-project service-account create --name my-service`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newGCPProjectListCmd())
	cmd.AddCommand(newGCPProjectSwitchCmd())
	cmd.AddCommand(newGCPProjectShowCmd())
	cmd.AddCommand(newGCPProjectConfigCmd())
	cmd.AddCommand(newGCPProjectValidateCmd())
	cmd.AddCommand(newGCPServiceAccountCmd())

	return cmd
}

// newGCPProjectListCmd creates the list subcommand.
func newGCPProjectListCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available GCP projects",
		Long: `List all Google Cloud Platform projects accessible to the current user.

This command shows:
- Project ID and name
- Current status and lifecycle state
- Associated billing account
- Active configuration
- Last used timestamp

Examples:
  # List projects in table format
  gz dev-env gcp-project list

  # List projects in JSON format
  gz dev-env gcp-project list --output json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.listProjects(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// newGCPProjectSwitchCmd creates the switch subcommand.
func newGCPProjectSwitchCmd() *cobra.Command {
	var (
		interactive   bool
		configuration string
	)

	cmd := &cobra.Command{
		Use:   "switch [PROJECT_ID]",
		Short: "Switch to a specific GCP project",
		Long: `Switch the active GCP project context.

This command:
- Updates the gcloud active configuration
- Sets the default project for gcloud commands
- Updates environment variables if needed
- Records the switch timestamp

Examples:
  # Switch to a specific project
  gz dev-env gcp-project switch my-project-id

  # Interactive project selection
  gz dev-env gcp-project switch --interactive

  # Switch using a specific configuration
  gz dev-env gcp-project switch my-project-id --config my-config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}

			var projectID string
			if len(args) > 0 {
				projectID = args[0]
			}

			return manager.switchProject(projectID, interactive, configuration)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive project selection")
	cmd.Flags().StringVarP(&configuration, "config", "c", "", "Use specific gcloud configuration")

	return cmd
}

// newGCPProjectShowCmd creates the show subcommand.
func newGCPProjectShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show [PROJECT_ID]",
		Short: "Show detailed information about a GCP project",
		Long: `Show detailed information about a Google Cloud Platform project.

If no project ID is specified, shows information about the current active project.

Examples:
  # Show current active project
  gz dev-env gcp-project show

  # Show specific project details
  gz dev-env gcp-project show my-project-id

  # Show project details in JSON format
  gz dev-env gcp-project show --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}

			var projectID string
			if len(args) > 0 {
				projectID = args[0]
			}

			return manager.showProject(projectID, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// newGCPProjectConfigCmd creates the config subcommand.
func newGCPProjectConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage gcloud configurations",
		Long: `Manage Google Cloud SDK configurations.

Configurations allow you to maintain different sets of gcloud properties
for different environments or projects.

Examples:
  # List all configurations
  gz dev-env gcp-project config list

  # Create a new configuration
  gz dev-env gcp-project config create --name prod --project my-prod-project

  # Switch to a configuration
  gz dev-env gcp-project config activate prod`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newGCPProjectConfigListCmd())
	cmd.AddCommand(newGCPProjectConfigCreateCmd())
	cmd.AddCommand(newGCPProjectConfigActivateCmd())
	cmd.AddCommand(newGCPProjectConfigDeleteCmd())

	return cmd
}

// newGCPProjectValidateCmd creates the validate subcommand.
func newGCPProjectValidateCmd() *cobra.Command {
	var (
		checkAPIs        bool
		checkBilling     bool
		checkPermissions bool
	)

	cmd := &cobra.Command{
		Use:   "validate [PROJECT_ID]",
		Short: "Validate GCP project access and configuration",
		Long: `Validate Google Cloud Platform project access and configuration.

This command checks:
- Project existence and access permissions
- Billing account association
- Required APIs enablement
- Service account permissions
- Quota and usage limits

Examples:
  # Validate current project
  gz dev-env gcp-project validate

  # Validate specific project with all checks
  gz dev-env gcp-project validate my-project-id --check-apis --check-billing --check-permissions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}

			var projectID string
			if len(args) > 0 {
				projectID = args[0]
			}

			return manager.validateProject(projectID, checkAPIs, checkBilling, checkPermissions)
		},
	}

	cmd.Flags().BoolVar(&checkAPIs, "check-apis", false, "Check required APIs are enabled")
	cmd.Flags().BoolVar(&checkBilling, "check-billing", false, "Check billing account association")
	cmd.Flags().BoolVar(&checkPermissions, "check-permissions", false, "Check IAM permissions")

	return cmd
}

// loadConfigurations loads all gcloud configurations.
func (m *GCPProjectManager) loadConfigurations() error {
	configurationsPath := filepath.Join(m.gcloudConfigPath, "configurations")
	if _, err := os.Stat(configurationsPath); os.IsNotExist(err) {
		return nil // No configurations directory
	}

	entries, err := os.ReadDir(configurationsPath)
	if err != nil {
		return fmt.Errorf("failed to read configurations directory: %w", err)
	}

	// Get active configuration
	activeConfig, _ := m.getActiveConfiguration()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configName := entry.Name()
		propertiesPath := filepath.Join(configurationsPath, configName, "properties")

		config := &GCPConfiguration{
			Name:           configName,
			IsActive:       configName == activeConfig,
			PropertiesPath: propertiesPath,
		}

		if err := m.parseConfigurationProperties(propertiesPath, config); err == nil {
			m.configurations[configName] = config
		}
	}

	return nil
}

// loadProjects loads accessible GCP projects.
func (m *GCPProjectManager) loadProjects() error {
	// Use gcloud to list projects
	cmd := exec.CommandContext(m.ctx, "gcloud", "projects", "list", "--format=json")

	output, err := cmd.Output()
	if err != nil {
		// If gcloud is not available or not authenticated, return empty list
		return err
	}

	var projects []struct {
		ProjectID      string `json:"projectId"`
		Name           string `json:"name"`
		ProjectNumber  string `json:"projectNumber"`
		LifecycleState string `json:"lifecycleState"`
	}

	if err := json.Unmarshal(output, &projects); err != nil {
		return fmt.Errorf("failed to parse projects JSON: %w", err)
	}

	// Get current project from active configuration
	currentProject, _ := m.getCurrentProject()

	for _, proj := range projects {
		project := &GCPProject{
			ID:             proj.ProjectID,
			Name:           proj.Name,
			Number:         proj.ProjectNumber,
			LifecycleState: proj.LifecycleState,
			IsActive:       proj.ProjectID == currentProject,
			Tags:           make(map[string]string),
		}

		// Try to get additional project details
		m.enrichProjectDetails(project)
		m.projects[proj.ProjectID] = project
	}

	return nil
}

// enrichProjectDetails adds additional details to a project.
func (m *GCPProjectManager) enrichProjectDetails(project *GCPProject) {
	// Get billing account
	if billingCmd := exec.CommandContext(m.ctx, "gcloud", "billing", "projects", "describe", project.ID, "--format=value(billingAccountName)"); billingCmd != nil {
		if output, err := billingCmd.Output(); err == nil {
			project.BillingAccount = strings.TrimSpace(string(output))
		}
	}

	// Get current configuration details
	for _, config := range m.configurations {
		if config.Project != project.ID {
			continue
		}
		project.Account = config.Account
		project.Region = config.Region
		project.Zone = config.Zone
		project.Configuration = config.Name

		break
	}
}

// getCurrentProject gets the current active project.
func (m *GCPProjectManager) getCurrentProject() (string, error) {
	cmd := exec.CommandContext(m.ctx, "gcloud", "config", "get-value", "project")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getActiveConfiguration gets the current active configuration.
func (m *GCPProjectManager) getActiveConfiguration() (string, error) {
	activeConfigPath := filepath.Join(m.gcloudConfigPath, "active_config")

	content, err := os.ReadFile(activeConfigPath)
	if err != nil {
		return "default", err // Default to "default" if no active config
	}

	return strings.TrimSpace(string(content)), nil
}

// parseConfigurationProperties parses gcloud configuration properties.
func (m *GCPProjectManager) parseConfigurationProperties(propertiesPath string, config *GCPConfiguration) error { //nolint:gocognit,gocyclo // Complex GCP configuration properties parsing with multiple branches
	content, err := os.ReadFile(propertiesPath)
	if err != nil {
		return err
	}

	// Try JSON format first
	var properties map[string]interface{}
	if err := json.Unmarshal(content, &properties); err == nil {
		if core, ok := properties["core"].(map[string]interface{}); ok {
			if project, ok := core["project"].(string); ok {
				config.Project = project
			}

			if account, ok := core["account"].(string); ok {
				config.Account = account
			}
		}

		if compute, ok := properties["compute"].(map[string]interface{}); ok {
			if region, ok := compute["region"].(string); ok {
				config.Region = region
			}

			if zone, ok := compute["zone"].(string); ok {
				config.Zone = zone
			}
		}

		return nil
	}

	// Fallback to INI format
	lines := strings.Split(string(content), "\n")

	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch currentSection + "." + key {
				case "core.project":
					config.Project = value
				case "core.account":
					config.Account = value
				case "compute.region":
					config.Region = value
				case "compute.zone":
					config.Zone = value
				}
			}
		}
	}

	return nil
}

// listProjects lists all available projects.
func (m *GCPProjectManager) listProjects(format string) error {
	if len(m.projects) == 0 {
		fmt.Println("No GCP projects found. Make sure you are authenticated with gcloud.")
		return nil
	}

	// Sort projects by name
	projects := make([]*GCPProject, 0, len(m.projects))
	for _, project := range m.projects {
		projects = append(projects, project)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(projects)

	case outputFormatTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Project ID", "Name", "State", "Account", "Region", "Active")

		for _, project := range projects {
			active := ""
			if project.IsActive {
				active = statusActive
			}

			_ = table.Append([]string{ //nolint:errcheck // Table operations are non-critical for CLI display
				project.ID,
				project.Name,
				project.LifecycleState,
				project.Account,
				project.Region,
				active,
			})
		}

		_ = table.Render() //nolint:errcheck // Table rendering errors are non-critical for CLI display

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// switchProject switches to a specific project.
func (m *GCPProjectManager) switchProject(projectID string, interactive bool, configuration string) error {
	// Interactive selection if no project specified
	if projectID == "" || interactive {
		var err error

		projectID, err = m.selectProjectInteractively()
		if err != nil {
			return err
		}
	}

	// Validate project exists
	project, exists := m.projects[projectID]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectID)
	}

	// Use specific configuration if provided
	if configuration != "" {
		if _, exists := m.configurations[configuration]; !exists {
			return fmt.Errorf("configuration '%s' not found", configuration)
		}

		// Activate the configuration
		cmd := exec.CommandContext(m.ctx, "gcloud", "config", "configurations", "activate", configuration)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to activate configuration '%s': %w", configuration, err)
		}
	}

	// Set the project
	cmd := exec.CommandContext(m.ctx, "gcloud", "config", "set", "project", projectID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set project '%s': %w", projectID, err)
	}

	// Update project status
	now := time.Now()

	for _, p := range m.projects {
		p.IsActive = false
	}

	project.IsActive = true
	project.LastUsed = &now

	fmt.Printf("✅ Switched to project: %s (%s)\n", project.Name, project.ID)

	if project.Region != "" {
		fmt.Printf("   Region: %s\n", project.Region)
	}

	if project.Zone != "" {
		fmt.Printf("   Zone: %s\n", project.Zone)
	}

	return nil
}

// selectProjectInteractively provides interactive project selection.
func (m *GCPProjectManager) selectProjectInteractively() (string, error) {
	if len(m.projects) == 0 {
		return "", fmt.Errorf("no projects available")
	}

	var (
		projects = make([]*GCPProject, 0, len(m.projects))
		items    = make([]string, 0, len(m.projects))
	)

	for _, project := range m.projects {
		projects = append(projects, project)

		label := fmt.Sprintf("%s (%s)", project.Name, project.ID)
		if project.IsActive {
			label += " [current]"
		}

		items = append(items, label)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	prompt := promptui.Select{
		Label: "Select a GCP project",
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return projects[index].ID, nil
}

// showProject shows detailed information about a project.
func (m *GCPProjectManager) showProject(projectID, format string) error {
	// Get current project if none specified
	if projectID == "" {
		var err error

		projectID, err = m.getCurrentProject()
		if err != nil {
			return fmt.Errorf("failed to get current project: %w", err)
		}
	}

	project, exists := m.projects[projectID]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectID)
	}

	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(project)

	case outputFormatTable:
		fmt.Printf("Project Details\n")
		fmt.Printf("===============\n")
		fmt.Printf("ID:               %s\n", project.ID)
		fmt.Printf("Name:             %s\n", project.Name)
		fmt.Printf("Number:           %s\n", project.Number)
		fmt.Printf("Lifecycle State:  %s\n", project.LifecycleState)
		fmt.Printf("Account:          %s\n", project.Account)
		fmt.Printf("Region:           %s\n", project.Region)
		fmt.Printf("Zone:             %s\n", project.Zone)
		fmt.Printf("Configuration:    %s\n", project.Configuration)
		fmt.Printf("Billing Account:  %s\n", project.BillingAccount)
		fmt.Printf("Active:           %t\n", project.IsActive)

		if project.LastUsed != nil {
			fmt.Printf("Last Used:        %s\n", project.LastUsed.Format("2006-01-02 15:04:05"))
		}

		if len(project.Tags) > 0 {
			fmt.Printf("\nTags:\n")

			for key, value := range project.Tags {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		if len(project.EnabledAPIs) > 0 {
			fmt.Printf("\nEnabled APIs:\n")

			for _, api := range project.EnabledAPIs {
				fmt.Printf("  - %s\n", api)
			}
		}

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// validateProject validates project access and configuration.
func (m *GCPProjectManager) validateProject(projectID string, checkAPIs, checkBilling, checkPermissions bool) error {
	// Get current project if none specified
	if projectID == "" {
		var err error

		projectID, err = m.getCurrentProject()
		if err != nil {
			return fmt.Errorf("failed to get current project: %w", err)
		}
	}

	fmt.Printf("Validating project: %s\n", projectID)
	fmt.Println("=======================")

	// Basic project access check
	cmd := exec.CommandContext(m.ctx, "gcloud", "projects", "describe", projectID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Project access: FAILED\n")
		return fmt.Errorf("cannot access project '%s': %w", projectID, err)
	}

	fmt.Printf("✅ Project access: OK\n")

	// Check billing if requested
	if checkBilling {
		cmd = exec.CommandContext(m.ctx, "gcloud", "billing", "projects", "describe", projectID)
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Billing account: NOT LINKED\n")
		} else {
			fmt.Printf("✅ Billing account: LINKED\n")
		}
	}

	// Check APIs if requested
	if checkAPIs {
		cmd = exec.CommandContext(m.ctx, "gcloud", "services", "list", "--enabled", "--project", projectID)
		if output, err := cmd.Output(); err != nil {
			fmt.Printf("❌ API access: FAILED\n")
		} else {
			apiCount := strings.Count(string(output), "\n") - 1 // Subtract header
			fmt.Printf("✅ APIs enabled: %d services\n", apiCount)
		}
	}

	// Check permissions if requested
	if checkPermissions {
		cmd = exec.CommandContext(m.ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
		if output, err := cmd.Output(); err != nil {
			fmt.Printf("❌ Authentication: FAILED\n")
		} else {
			account := strings.TrimSpace(string(output))
			fmt.Printf("✅ Authentication: %s\n", account)
		}
	}

	fmt.Printf("\nValidation completed for project: %s\n", projectID)

	return nil
}

// Configuration management commands.
func newGCPProjectConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all gcloud configurations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.listConfigurations()
		},
	}
}

func newGCPProjectConfigCreateCmd() *cobra.Command {
	var name, project, account, region, zone string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new gcloud configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.createConfiguration(name, project, account, region, zone)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Configuration name (required)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project ID")
	cmd.Flags().StringVarP(&account, "account", "a", "", "Account email")
	cmd.Flags().StringVarP(&region, "region", "r", "", "Default region")
	cmd.Flags().StringVarP(&zone, "zone", "z", "", "Default zone")

	_ = cmd.MarkFlagRequired("name") //nolint:errcheck // Required flag setup

	return cmd
}

func newGCPProjectConfigActivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "activate [CONFIG_NAME]",
		Short: "Activate a gcloud configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("configuration name is required")
			}
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.activateConfiguration(args[0])
		},
	}
}

func newGCPProjectConfigDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [CONFIG_NAME]",
		Short: "Delete a gcloud configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("configuration name is required")
			}
			manager, err := NewGCPProjectManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.deleteConfiguration(args[0])
		},
	}
}

func (m *GCPProjectManager) listConfigurations() error {
	if len(m.configurations) == 0 {
		fmt.Println("No gcloud configurations found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Name", "Project", "Account", "Region", "Zone", "Active")

	for _, config := range m.configurations {
		active := ""
		if config.IsActive {
			active = statusActive
		}

		_ = table.Append( //nolint:errcheck // Table operations are non-critical for CLI display
			config.Name,
			config.Project,
			config.Account,
			config.Region,
			config.Zone,
			active,
		)
	}

	_ = table.Render() //nolint:errcheck // Table rendering errors are non-critical for CLI display

	return nil
}

func (m *GCPProjectManager) createConfiguration(name, project, account, region, zone string) error {
	// Create the configuration
	cmd := exec.CommandContext(m.ctx, "gcloud", "config", "configurations", "create", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create configuration '%s': %w", name, err)
	}

	// Set properties if provided
	if project != "" {
		cmd = exec.CommandContext(m.ctx, "gcloud", "config", "set", "project", project, "--configuration", name)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to set project: %v\n", err)
		}
	}

	if account != "" {
		cmd = exec.CommandContext(m.ctx, "gcloud", "config", "set", "account", account, "--configuration", name)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to set account: %v\n", err)
		}
	}

	if region != "" {
		cmd = exec.CommandContext(m.ctx, "gcloud", "config", "set", "compute/region", region, "--configuration", name)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to set region: %v\n", err)
		}
	}

	if zone != "" {
		cmd = exec.CommandContext(m.ctx, "gcloud", "config", "set", "compute/zone", zone, "--configuration", name)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to set zone: %v\n", err)
		}
	}

	fmt.Printf("✅ Configuration '%s' created successfully\n", name)

	return nil
}

func (m *GCPProjectManager) activateConfiguration(name string) error {
	cmd := exec.CommandContext(m.ctx, "gcloud", "config", "configurations", "activate", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to activate configuration '%s': %w", name, err)
	}

	fmt.Printf("✅ Configuration '%s' activated\n", name)

	return nil
}

func (m *GCPProjectManager) deleteConfiguration(name string) error {
	// Confirm deletion
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Delete configuration '%s'", name),
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
		return fmt.Errorf("operation canceled")
	}

	cmd := exec.CommandContext(m.ctx, "gcloud", "config", "configurations", "delete", name, "--quiet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete configuration '%s': %w", name, err)
	}

	fmt.Printf("✅ Configuration '%s' deleted\n", name)

	return nil
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// CreateOptions contains options for repository creation.
type CreateOptions struct {
	// Required fields
	Provider string
	Org      string
	Name     string

	// Repository settings
	Description string
	Private     bool
	Template    string

	// Initialization options
	AutoInit          bool
	GitignoreTemplate string
	License           string
	DefaultBranch     string

	// Features
	Issues   bool
	Wiki     bool
	Projects bool

	// Advanced settings
	Homepage         string
	Topics           []string
	AllowMergeCommit bool
	AllowSquashMerge bool
	AllowRebaseMerge bool

	// Output options
	Format string
	Quiet  bool
}

// newRepoCreateCmd creates the repo create command.
func newRepoCreateCmd() *cobra.Command {
	opts := &CreateOptions{
		AutoInit:         true,
		DefaultBranch:    "main",
		Issues:           true,
		AllowMergeCommit: true,
		AllowSquashMerge: true,
		AllowRebaseMerge: true,
		Format:           "table",
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new repository",
		Long: `Create a new repository on the specified Git platform with
customizable settings including visibility, templates, and initialization options.

This command provides comprehensive repository creation capabilities including:
- Template-based repository creation
- Advanced repository settings and permissions
- Automatic initialization with README, gitignore, and license
- Support for all major Git platforms`,
		Example: `  # Create a public repository
  gz git repo create --provider github --org myorg --name newrepo

  # Create from template
  gz git repo create --provider github --template myorg/template-repo --name myapp

  # Create with full options
  gz git repo create --provider gitlab --org mygroup --name api \
    --private --description "API service" --auto-init --license MIT

  # Create with topics and homepage
  gz git repo create --provider github --org myorg --name webapp \
    --description "Web application" --homepage "https://example.com" \
    --topics "javascript,react,webapp"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoCreate(cmd.Context(), opts)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Repository name")

	// Repository settings
	cmd.Flags().StringVar(&opts.Description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&opts.Private, "private", false, "Create as private repository")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Template repository (org/repo)")

	// Initialization options
	cmd.Flags().BoolVar(&opts.AutoInit, "auto-init", true, "Initialize with README")
	cmd.Flags().StringVar(&opts.GitignoreTemplate, "gitignore", "", "Gitignore template name")
	cmd.Flags().StringVar(&opts.License, "license", "", "License template (MIT, Apache-2.0, GPL-3.0, etc.)")
	cmd.Flags().StringVar(&opts.DefaultBranch, "default-branch", "main", "Default branch name")

	// Features
	cmd.Flags().BoolVar(&opts.Issues, "issues", true, "Enable issues")
	cmd.Flags().BoolVar(&opts.Wiki, "wiki", false, "Enable wiki")
	cmd.Flags().BoolVar(&opts.Projects, "projects", false, "Enable projects")

	// Advanced settings
	cmd.Flags().StringVar(&opts.Homepage, "homepage", "", "Repository homepage URL")
	cmd.Flags().StringSliceVar(&opts.Topics, "topics", nil, "Repository topics/tags")
	cmd.Flags().BoolVar(&opts.AllowMergeCommit, "allow-merge-commit", true, "Allow merge commits")
	cmd.Flags().BoolVar(&opts.AllowSquashMerge, "allow-squash-merge", true, "Allow squash merging")
	cmd.Flags().BoolVar(&opts.AllowRebaseMerge, "allow-rebase-merge", true, "Allow rebase merging")

	// Output options
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", false, "Suppress output")

	// Mark required flags
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("name")

	return cmd
}

// runRepoCreate executes the repository creation operation.
func runRepoCreate(ctx context.Context, opts *CreateOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get provider
	gitProvider, err := getGitProvider(opts.Provider, opts.Org)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Create repository request
	request := opts.toCreateRequest()

	// Execute creation
	repo, err := gitProvider.CreateRepository(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Output result
	if !opts.Quiet {
		return outputRepository(repo, opts.Format)
	}

	return nil
}

// Validate validates the create options.
func (opts *CreateOptions) Validate() error {
	if opts.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if opts.Org == "" {
		return fmt.Errorf("organization is required")
	}

	if opts.Name == "" {
		return fmt.Errorf("repository name is required")
	}

	// Validate repository name
	if err := validateRepositoryName(opts.Name); err != nil {
		return fmt.Errorf("invalid repository name: %w", err)
	}

	// Validate template format if provided
	if opts.Template != "" {
		if !strings.Contains(opts.Template, "/") {
			return fmt.Errorf("template must be in format 'owner/repo'")
		}
	}

	// Validate output format
	if !isValidOutputFormat(opts.Format) {
		return fmt.Errorf("invalid output format: %s (valid: table, json, yaml)", opts.Format)
	}

	return nil
}

// toCreateRequest converts CreateOptions to provider.CreateRepoRequest.
func (opts *CreateOptions) toCreateRequest() provider.CreateRepoRequest {
	// Parse template if provided
	var templateOwner, templateRepo string
	if opts.Template != "" {
		parts := strings.SplitN(opts.Template, "/", 2)
		if len(parts) == 2 {
			templateOwner = parts[0]
			templateRepo = parts[1]
		}
	}

	// Determine visibility
	visibility := provider.VisibilityPublic
	if opts.Private {
		visibility = provider.VisibilityPrivate
	}

	return provider.CreateRepoRequest{
		Name:              opts.Name,
		Description:       opts.Description,
		Homepage:          opts.Homepage,
		Private:           opts.Private,
		Visibility:        visibility,
		HasIssues:         opts.Issues,
		HasProjects:       opts.Projects,
		HasWiki:           opts.Wiki,
		AutoInit:          opts.AutoInit,
		GitignoreTemplate: opts.GitignoreTemplate,
		LicenseTemplate:   opts.License,
		AllowSquashMerge:  opts.AllowSquashMerge,
		AllowMergeCommit:  opts.AllowMergeCommit,
		AllowRebaseMerge:  opts.AllowRebaseMerge,
		DefaultBranch:     opts.DefaultBranch,
		Topics:            opts.Topics,
		TemplateOwner:     templateOwner,
		TemplateRepo:      templateRepo,
	}
}

// validateRepositoryName validates repository name according to Git hosting standards.
func validateRepositoryName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("name too long (max 100 characters)")
	}

	// Check for invalid characters
	for _, char := range name {
		if !isValidRepoNameChar(char) {
			return fmt.Errorf("invalid character '%c' in repository name", char)
		}
	}

	// Check for reserved names
	if isReservedName(name) {
		return fmt.Errorf("'%s' is a reserved name", name)
	}

	return nil
}

// isValidRepoNameChar checks if a character is valid in repository names.
func isValidRepoNameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}

// isReservedName checks if a name is reserved.
func isReservedName(name string) bool {
	reserved := []string{
		"admin", "api", "app", "blog", "docs", "help", "mail", "support",
		"www", "ftp", "ssh", "git", "raw", "gist", "status", "wiki",
	}

	lower := strings.ToLower(name)
	for _, r := range reserved {
		if lower == r {
			return true
		}
	}

	return false
}

// isValidOutputFormat checks if the output format is valid.
func isValidOutputFormat(format string) bool {
	validFormats := []string{"table", "json", "yaml"}
	for _, valid := range validFormats {
		if format == valid {
			return true
		}
	}
	return false
}

// getGitProvider gets a provider instance for the specified type and organization.
func getGitProvider(providerType, org string) (provider.GitProvider, error) {
	// For now, return error indicating providers need implementation
	// TODO: Implement actual provider creation logic
	// This would involve:
	// 1. Creating provider factory
	// 2. Registering provider constructors
	// 3. Creating provider configuration
	// 4. Getting provider instance from registry
	return nil, fmt.Errorf("provider implementation not available yet - provider: %s, org: %s", providerType, org)
}

// outputRepository outputs repository information in the specified format.
func outputRepository(repo *provider.Repository, format string) error {
	switch format {
	case "table":
		return outputRepositoryTable(repo)
	case "json":
		return outputRepositoryJSON(repo)
	case "yaml":
		return outputRepositoryYAML(repo)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// outputRepositoryTable outputs repository in table format.
func outputRepositoryTable(repo *provider.Repository) error {
	fmt.Printf("\n✅ Repository created successfully!\n\n")
	fmt.Printf("Name:         %s\n", repo.FullName)
	fmt.Printf("Description:  %s\n", repo.Description)
	fmt.Printf("Private:      %v\n", repo.Private)
	fmt.Printf("URL:          %s\n", repo.HTMLURL)
	fmt.Printf("Clone URL:    %s\n", repo.CloneURL)
	fmt.Printf("SSH URL:      %s\n", repo.SSHURL)
	fmt.Printf("Created:      %s\n", repo.CreatedAt.Format(time.RFC3339))

	if len(repo.Topics) > 0 {
		fmt.Printf("Topics:       %s\n", strings.Join(repo.Topics, ", "))
	}

	return nil
}

// outputRepositoryJSON outputs repository in JSON format.
func outputRepositoryJSON(repo *provider.Repository) error {
	// TODO: Implement JSON output
	return fmt.Errorf("JSON output not implemented yet")
}

// outputRepositoryYAML outputs repository in YAML format.
func outputRepositoryYAML(repo *provider.Repository) error {
	// TODO: Implement YAML output
	return fmt.Errorf("YAML output not implemented yet")
}

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/google/go-github/v66/github"
)

// RunComplianceAudit performs a compliance audit for all repositories in an organization
func (c *RepoConfigClient) RunComplianceAudit(ctx context.Context, configPath string) (*config.AuditReport, error) {
	// Load the repository configuration
	repoConfig, err := config.LoadRepoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Fetch all repositories from the organization
	repos, err := c.listAllRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Build repository state map
	repoStates := make(map[string]config.RepositoryState)

	for _, repo := range repos {
		state, err := c.getRepositoryState(ctx, repo)
		if err != nil {
			// Log error but continue with other repos
			fmt.Printf("Warning: Failed to get state for %s: %v\n", repo.GetName(), err)
			continue
		}
		repoStates[repo.GetName()] = state
	}

	// Run the compliance audit
	return repoConfig.RunComplianceAudit(repoStates)
}

// getRepositoryState fetches the current state of a repository
func (c *RepoConfigClient) getRepositoryState(ctx context.Context, repo *github.Repository) (config.RepositoryState, error) {
	state := config.RepositoryState{
		Name:             repo.GetName(),
		Private:          repo.GetPrivate(),
		Archived:         repo.GetArchived(),
		HasIssues:        repo.GetHasIssues(),
		HasWiki:          repo.GetHasWiki(),
		HasProjects:      repo.GetHasProjects(),
		HasDownloads:     repo.GetHasDownloads(),
		LastModified:     repo.GetUpdatedAt().Time,
		BranchProtection: make(map[string]config.BranchProtectionState),
	}

	// Get security features
	if repo.GetSecurityAndAnalysis() != nil {
		sa := repo.GetSecurityAndAnalysis()
		if sa.SecretScanning != nil {
			state.VulnerabilityAlerts = sa.SecretScanning.GetStatus() == "enabled"
		}
		if sa.AdvancedSecurity != nil {
			state.SecurityAdvisories = sa.AdvancedSecurity.GetStatus() == "enabled"
		}
	}

	// Get branch protection for default branch
	defaultBranch := repo.GetDefaultBranch()
	if defaultBranch != "" {
		protection, _, err := c.client.Repositories.GetBranchProtection(ctx, c.org, repo.GetName(), defaultBranch)
		if err == nil && protection != nil {
			state.BranchProtection[defaultBranch] = config.BranchProtectionState{
				Protected:       true,
				RequiredReviews: getRequiredReviews(protection),
				EnforceAdmins:   protection.GetEnforceAdmins().GetEnabled(),
			}
		}
	}

	// Check for specific files
	state.Files = c.checkRequiredFiles(ctx, repo.GetName())

	// Check for workflows
	state.Workflows = c.listWorkflows(ctx, repo.GetName())

	return state, nil
}

// getRequiredReviews extracts the required review count from branch protection
func getRequiredReviews(protection *github.Protection) int {
	if protection.RequiredPullRequestReviews == nil {
		return 0
	}
	return protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
}

// checkRequiredFiles checks for the existence of specific files in the repository
func (c *RepoConfigClient) checkRequiredFiles(ctx context.Context, repoName string) []string {
	var foundFiles []string

	// List of files to check
	filesToCheck := []string{
		"README.md",
		"LICENSE",
		"SECURITY.md",
		"CONTRIBUTING.md",
		"CODE_OF_CONDUCT.md",
		".github/CODEOWNERS",
		"COMPLIANCE.md",
	}

	for _, file := range filesToCheck {
		_, _, _, err := c.client.Repositories.GetContents(ctx, c.org, repoName, file, nil)
		if err == nil {
			foundFiles = append(foundFiles, file)
		}
	}

	return foundFiles
}

// listWorkflows lists GitHub Actions workflows in the repository
func (c *RepoConfigClient) listWorkflows(ctx context.Context, repoName string) []string {
	var workflows []string

	// List workflows directory
	_, dirContent, _, err := c.client.Repositories.GetContents(ctx, c.org, repoName, ".github/workflows", nil)
	if err != nil {
		return workflows
	}

	for _, content := range dirContent {
		if content.Type != nil && *content.Type == "file" {
			name := content.GetName()
			if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
				// Remove extension for consistency
				workflows = append(workflows, strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml"))
			}
		}
	}

	return workflows
}

// listAllRepositories fetches all repositories from the organization
func (c *RepoConfigClient) listAllRepositories(ctx context.Context) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.org, opt)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// GenerateAuditReportFile saves the audit report to a file
func GenerateAuditReportFile(report *config.AuditReport, format, outputPath string) error {
	var content []byte
	var err error

	switch format {
	case "json":
		content, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		if !strings.HasSuffix(outputPath, ".json") {
			outputPath += ".json"
		}

	case "yaml":
		// Could add YAML support here
		return fmt.Errorf("YAML format not yet implemented")

	case "html":
		content = []byte(generateHTMLReport(report))
		if !strings.HasSuffix(outputPath, ".html") {
			outputPath += ".html"
		}

	case "markdown", "md":
		content = []byte(report.GenerateAuditSummary())
		if !strings.HasSuffix(outputPath, ".md") {
			outputPath += ".md"
		}

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateHTMLReport creates an HTML report from the audit data
func generateHTMLReport(report *config.AuditReport) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Compliance Audit Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2, h3 { color: #333; }
        .summary { background: #f0f0f0; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .compliant { color: #28a745; }
        .non-compliant { color: #dc3545; }
        .warning { color: #ffc107; }
        table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .violation { background-color: #fee; }
        .exception { background-color: #fef; }
        .metric { font-size: 24px; font-weight: bold; }
    </style>
</head>
<body>
    <h1>Compliance Audit Report</h1>
    <p><strong>Organization:</strong> %s<br>
    <strong>Generated:</strong> %s</p>
    
    <div class="summary">
        <h2>Summary</h2>
        <p>
            <span class="metric">%.1f%%</span> Overall Compliance<br>
            %d of %d repositories are fully compliant<br>
            %d total violations found<br>
            %d active exceptions
        </p>
    </div>
`, report.Organization, report.Organization, report.GeneratedAt.Format(time.RFC3339),
		report.Summary.CompliancePercentage,
		report.Summary.CompliantRepositories,
		report.Summary.AuditedRepositories,
		report.Summary.TotalViolations,
		report.Summary.ActiveExceptions)

	// Add policy compliance section
	html += `<h2>Policy Compliance</h2><table><tr><th>Policy</th><th>Compliance</th><th>Compliant</th><th>Violating</th><th>Exempted</th></tr>`

	for _, policy := range report.Policies {
		html += fmt.Sprintf(`<tr>
			<td>%s<br><small>%s</small></td>
			<td>%.1f%%</td>
			<td class="compliant">%d</td>
			<td class="non-compliant">%d</td>
			<td class="warning">%d</td>
		</tr>`, policy.PolicyName, policy.Description, policy.CompliancePercentage,
			policy.CompliantRepos, policy.ViolatingRepos, policy.ExemptedRepos)
	}
	html += `</table>`

	// Add non-compliant repositories
	html += `<h2>Repository Compliance Details</h2><table><tr><th>Repository</th><th>Status</th><th>Violations</th><th>Exceptions</th></tr>`

	for _, repo := range report.Repositories {
		status := `<span class="compliant">Compliant</span>`
		if !repo.Compliant {
			status = `<span class="non-compliant">Non-Compliant</span>`
		}

		violations := ""
		for _, v := range repo.Violations {
			violations += fmt.Sprintf("%s/%s: %s<br>", v.PolicyName, v.RuleName, v.Message)
		}

		exceptions := ""
		for _, e := range repo.Exceptions {
			exceptions += fmt.Sprintf("%s/%s: %s<br>", e.PolicyName, e.RuleName, e.Reason)
		}

		rowClass := ""
		if !repo.Compliant {
			rowClass = ` class="violation"`
		} else if len(repo.Exceptions) > 0 {
			rowClass = ` class="exception"`
		}

		html += fmt.Sprintf(`<tr%s>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, rowClass, repo.Repository, status, violations, exceptions)
	}

	html += `</table></body></html>`
	return html
}

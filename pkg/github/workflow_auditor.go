package github

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowAuditor performs security audits on GitHub Actions workflows
type WorkflowAuditor struct {
	logger    Logger
	apiClient APIClient
}

// WorkflowAuditResult represents the audit result for a repository
type WorkflowAuditResult struct {
	Repository      string                    `json:"repository"`
	Organization    string                    `json:"organization"`
	TotalWorkflows  int                       `json:"total_workflows"`
	AuditedFiles    []WorkflowFileAudit       `json:"audited_files"`
	SecurityIssues  []WorkflowSecurityIssue   `json:"security_issues"`
	PermissionUsage []WorkflowPermissionUsage `json:"permission_usage"`
	ActionUsage     []ActionUsageInfo         `json:"action_usage"`
	Summary         WorkflowAuditSummary      `json:"summary"`
	Timestamp       time.Time                 `json:"timestamp"`
}

// WorkflowFileAudit represents audit information for a single workflow file
type WorkflowFileAudit struct {
	FilePath      string                  `json:"file_path"`
	WorkflowName  string                  `json:"workflow_name"`
	Triggers      []string                `json:"triggers"`
	Jobs          []JobAuditInfo          `json:"jobs"`
	Permissions   map[string]string       `json:"permissions,omitempty"`
	SecurityScore int                     `json:"security_score"`
	Issues        []WorkflowSecurityIssue `json:"issues"`
	LastModified  time.Time               `json:"last_modified"`
}

// JobAuditInfo represents audit information for a job within a workflow
type JobAuditInfo struct {
	JobID         string            `json:"job_id"`
	RunsOn        string            `json:"runs_on"`
	Permissions   map[string]string `json:"permissions,omitempty"`
	Steps         []StepAuditInfo   `json:"steps"`
	Environment   string            `json:"environment,omitempty"`
	SecurityScore int               `json:"security_score"`
	UsesSecrets   []string          `json:"uses_secrets,omitempty"`
	UsesVariables []string          `json:"uses_variables,omitempty"`
}

// StepAuditInfo represents audit information for a step within a job
type StepAuditInfo struct {
	Name          string            `json:"name,omitempty"`
	Uses          string            `json:"uses,omitempty"`
	Run           string            `json:"run,omitempty"`
	ActionVersion string            `json:"action_version,omitempty"`
	SecurityRisk  SecurityRiskLevel `json:"security_risk"`
	RiskReasons   []string          `json:"risk_reasons,omitempty"`
	UsesSecrets   []string          `json:"uses_secrets,omitempty"`
	UsesVariables []string          `json:"uses_variables,omitempty"`
}

// WorkflowSecurityIssue represents a security issue found in a workflow
type WorkflowSecurityIssue struct {
	ID          string                `json:"id"`
	Type        WorkflowIssueType     `json:"type"`
	Severity    SecurityIssueSeverity `json:"severity"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	FilePath    string                `json:"file_path"`
	JobID       string                `json:"job_id,omitempty"`
	StepIndex   int                   `json:"step_index,omitempty"`
	LineNumber  int                   `json:"line_number,omitempty"`
	Suggestion  string                `json:"suggestion"`
	References  []string              `json:"references,omitempty"`
}

// WorkflowPermissionUsage represents permission usage statistics
type WorkflowPermissionUsage struct {
	Scope         string   `json:"scope"`
	Permission    string   `json:"permission"`
	UsageCount    int      `json:"usage_count"`
	WorkflowFiles []string `json:"workflow_files"`
	Recommended   string   `json:"recommended,omitempty"`
}

// ActionUsageInfo represents information about action usage
type ActionUsageInfo struct {
	ActionName    string            `json:"action_name"`
	Version       string            `json:"version"`
	UsageCount    int               `json:"usage_count"`
	WorkflowFiles []string          `json:"workflow_files"`
	SecurityRisk  SecurityRiskLevel `json:"security_risk"`
	IsVerified    bool              `json:"is_verified"`
	IsDeprecated  bool              `json:"is_deprecated"`
}

// WorkflowAuditSummary provides summary statistics
type WorkflowAuditSummary struct {
	TotalFiles             int                       `json:"total_files"`
	FilesWithIssues        int                       `json:"files_with_issues"`
	CriticalIssues         int                       `json:"critical_issues"`
	HighRiskIssues         int                       `json:"high_risk_issues"`
	MediumRiskIssues       int                       `json:"medium_risk_issues"`
	LowRiskIssues          int                       `json:"low_risk_issues"`
	AverageSecurityScore   float64                   `json:"average_security_score"`
	PermissionDistribution map[string]int            `json:"permission_distribution"`
	ActionRiskDistribution map[SecurityRiskLevel]int `json:"action_risk_distribution"`
	ComplianceScore        float64                   `json:"compliance_score"`
}

// Enum types
type WorkflowIssueType string

const (
	IssueTypeExcessivePermissions WorkflowIssueType = "excessive_permissions"
	IssueTypeUnpinnedAction       WorkflowIssueType = "unpinned_action"
	IssueTypeDeprecatedAction     WorkflowIssueType = "deprecated_action"
	IssueTypeUnverifiedAction     WorkflowIssueType = "unverified_action"
	IssueTypeSecretExposure       WorkflowIssueType = "secret_exposure"
	IssueTypeCodeInjection        WorkflowIssueType = "code_injection"
	IssueTypePrivilegeEscalation  WorkflowIssueType = "privilege_escalation"
	IssueTypeInsecureRunner       WorkflowIssueType = "insecure_runner"
	IssueTypeMissingPermissions   WorkflowIssueType = "missing_permissions"
	IssueTypeEnvironmentIssue     WorkflowIssueType = "environment_issue"
)

type SecurityIssueSeverity string

const (
	SeverityCritical SecurityIssueSeverity = "critical"
	SeverityHigh     SecurityIssueSeverity = "high"
	SeverityMedium   SecurityIssueSeverity = "medium"
	SeverityLow      SecurityIssueSeverity = "low"
	SeverityInfo     SecurityIssueSeverity = "info"
)

type SecurityRiskLevel string

const (
	SecurityRiskCritical SecurityRiskLevel = "critical"
	SecurityRiskHigh     SecurityRiskLevel = "high"
	SecurityRiskMedium   SecurityRiskLevel = "medium"
	SecurityRiskLow      SecurityRiskLevel = "low"
	SecurityRiskNone     SecurityRiskLevel = "none"
)

// Workflow structure for parsing YAML
type WorkflowFile struct {
	Name        string                 `yaml:"name"`
	On          interface{}            `yaml:"on"`
	Permissions map[string]interface{} `yaml:"permissions"`
	Jobs        map[string]Job         `yaml:"jobs"`
	Env         map[string]string      `yaml:"env"`
}

type Job struct {
	Name        string                 `yaml:"name"`
	RunsOn      interface{}            `yaml:"runs-on"`
	Permissions map[string]interface{} `yaml:"permissions"`
	Environment interface{}            `yaml:"environment"`
	Steps       []Step                 `yaml:"steps"`
	Env         map[string]string      `yaml:"env"`
}

type Step struct {
	Name string            `yaml:"name"`
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	With map[string]string `yaml:"with"`
	Env  map[string]string `yaml:"env"`
}

// NewWorkflowAuditor creates a new workflow auditor
func NewWorkflowAuditor(logger Logger, apiClient APIClient) *WorkflowAuditor {
	return &WorkflowAuditor{
		logger:    logger,
		apiClient: apiClient,
	}
}

// AuditRepository performs a comprehensive audit of all workflows in a repository
func (wa *WorkflowAuditor) AuditRepository(ctx context.Context, organization, repository string) (*WorkflowAuditResult, error) {
	wa.logger.Info("Starting workflow audit", "organization", organization, "repository", repository)

	result := &WorkflowAuditResult{
		Repository:      repository,
		Organization:    organization,
		AuditedFiles:    make([]WorkflowFileAudit, 0),
		SecurityIssues:  make([]WorkflowSecurityIssue, 0),
		PermissionUsage: make([]WorkflowPermissionUsage, 0),
		ActionUsage:     make([]ActionUsageInfo, 0),
		Timestamp:       time.Now(),
	}

	// Get workflow files from repository
	workflowFiles, err := wa.getWorkflowFiles(ctx, organization, repository)
	if err != nil {
		return result, fmt.Errorf("failed to get workflow files: %w", err)
	}

	result.TotalWorkflows = len(workflowFiles)

	// Audit each workflow file
	for _, file := range workflowFiles {
		fileAudit, err := wa.auditWorkflowFile(ctx, organization, repository, file)
		if err != nil {
			wa.logger.Error("Failed to audit workflow file", "file", file, "error", err)
			continue
		}

		result.AuditedFiles = append(result.AuditedFiles, *fileAudit)
		result.SecurityIssues = append(result.SecurityIssues, fileAudit.Issues...)
	}

	// Generate usage statistics
	result.PermissionUsage = wa.analyzePermissionUsage(result.AuditedFiles)
	result.ActionUsage = wa.analyzeActionUsage(result.AuditedFiles)
	result.Summary = wa.generateSummary(result)

	wa.logger.Info("Workflow audit completed",
		"organization", organization,
		"repository", repository,
		"total_workflows", result.TotalWorkflows,
		"security_issues", len(result.SecurityIssues))

	return result, nil
}

// AuditOrganization performs workflow audit across all repositories in an organization
func (wa *WorkflowAuditor) AuditOrganization(ctx context.Context, organization string) ([]*WorkflowAuditResult, error) {
	wa.logger.Info("Starting organization-wide workflow audit", "organization", organization)

	repos, err := wa.apiClient.ListOrganizationRepositories(ctx, organization)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	results := make([]*WorkflowAuditResult, 0)

	for _, repo := range repos {
		if repo.Archived || repo.Disabled {
			continue
		}

		result, err := wa.AuditRepository(ctx, organization, repo.Name)
		if err != nil {
			wa.logger.Error("Failed to audit repository", "repository", repo.Name, "error", err)
			continue
		}

		if result.TotalWorkflows > 0 {
			results = append(results, result)
		}
	}

	wa.logger.Info("Organization audit completed",
		"organization", organization,
		"audited_repositories", len(results))

	return results, nil
}

// getWorkflowFiles retrieves all workflow files from a repository
func (wa *WorkflowAuditor) getWorkflowFiles(ctx context.Context, organization, repository string) ([]string, error) {
	// In a real implementation, this would use GitHub API to get files from .github/workflows/
	// For now, return mock workflow files
	return []string{
		".github/workflows/ci.yml",
		".github/workflows/release.yml",
		".github/workflows/security.yml",
	}, nil
}

// auditWorkflowFile performs security audit on a single workflow file
func (wa *WorkflowAuditor) auditWorkflowFile(ctx context.Context, organization, repository, filePath string) (*WorkflowFileAudit, error) {
	// Get file content - in real implementation, this would fetch from GitHub API
	content, err := wa.getFileContent(ctx, organization, repository, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	// Parse workflow YAML
	var workflow WorkflowFile
	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	audit := &WorkflowFileAudit{
		FilePath:     filePath,
		WorkflowName: workflow.Name,
		Triggers:     wa.extractTriggers(workflow.On),
		Jobs:         make([]JobAuditInfo, 0),
		Permissions:  wa.normalizePermissions(workflow.Permissions),
		Issues:       make([]WorkflowSecurityIssue, 0),
		LastModified: time.Now(),
	}

	// Audit workflow-level permissions
	audit.Issues = append(audit.Issues, wa.auditWorkflowPermissions(audit.Permissions, filePath)...)

	// Audit each job
	for jobID, job := range workflow.Jobs {
		jobAudit := wa.auditJob(jobID, job, filePath)
		audit.Jobs = append(audit.Jobs, jobAudit)

		// Collect job-level issues
		for _, step := range jobAudit.Steps {
			if step.SecurityRisk == SecurityRiskHigh || step.SecurityRisk == SecurityRiskCritical {
				issue := WorkflowSecurityIssue{
					ID:          fmt.Sprintf("%s-%s-%d", filePath, jobID, len(audit.Issues)),
					Type:        wa.determineIssueType(step.RiskReasons),
					Severity:    wa.mapRiskToSeverity(step.SecurityRisk),
					Title:       fmt.Sprintf("Security risk in job '%s'", jobID),
					Description: strings.Join(step.RiskReasons, "; "),
					FilePath:    filePath,
					JobID:       jobID,
					Suggestion:  wa.generateSecuritySuggestion(step.RiskReasons),
				}
				audit.Issues = append(audit.Issues, issue)
			}
		}
	}

	// Calculate security score
	audit.SecurityScore = wa.calculateSecurityScore(audit)

	return audit, nil
}

// auditJob performs security audit on a job
func (wa *WorkflowAuditor) auditJob(jobID string, job Job, filePath string) JobAuditInfo {
	jobAudit := JobAuditInfo{
		JobID:         jobID,
		RunsOn:        wa.normalizeRunsOn(job.RunsOn),
		Permissions:   wa.normalizePermissions(job.Permissions),
		Steps:         make([]StepAuditInfo, 0),
		UsesSecrets:   make([]string, 0),
		UsesVariables: make([]string, 0),
	}

	// Extract environment
	if env := wa.normalizeEnvironment(job.Environment); env != "" {
		jobAudit.Environment = env
	}

	// Audit each step
	for i, step := range job.Steps {
		stepAudit := wa.auditStep(step, i)
		jobAudit.Steps = append(jobAudit.Steps, stepAudit)

		// Collect secrets and variables usage
		jobAudit.UsesSecrets = append(jobAudit.UsesSecrets, stepAudit.UsesSecrets...)
		jobAudit.UsesVariables = append(jobAudit.UsesVariables, stepAudit.UsesVariables...)
	}

	// Calculate job security score
	jobAudit.SecurityScore = wa.calculateJobSecurityScore(jobAudit)

	return jobAudit
}

// auditStep performs security audit on a step
func (wa *WorkflowAuditor) auditStep(step Step, stepIndex int) StepAuditInfo {
	stepAudit := StepAuditInfo{
		Name:          step.Name,
		Uses:          step.Uses,
		Run:           step.Run,
		SecurityRisk:  SecurityRiskNone,
		RiskReasons:   make([]string, 0),
		UsesSecrets:   make([]string, 0),
		UsesVariables: make([]string, 0),
	}

	// Analyze action usage
	if step.Uses != "" {
		stepAudit.ActionVersion = wa.extractActionVersion(step.Uses)
		risks := wa.analyzeActionSecurity(step.Uses)
		stepAudit.RiskReasons = append(stepAudit.RiskReasons, risks...)
	}

	// Analyze script execution
	if step.Run != "" {
		risks := wa.analyzeScriptSecurity(step.Run)
		stepAudit.RiskReasons = append(stepAudit.RiskReasons, risks...)
	}

	// Extract secrets and variables usage
	stepAudit.UsesSecrets = wa.extractSecretsUsage(step)
	stepAudit.UsesVariables = wa.extractVariablesUsage(step)

	// Determine overall risk level
	stepAudit.SecurityRisk = wa.calculateStepRisk(stepAudit.RiskReasons)

	return stepAudit
}

// Security analysis helper methods

func (wa *WorkflowAuditor) auditWorkflowPermissions(permissions map[string]string, filePath string) []WorkflowSecurityIssue {
	issues := make([]WorkflowSecurityIssue, 0)

	// Check for overly broad permissions
	for scope, permission := range permissions {
		if permission == "write" && wa.isHighRiskScope(scope) {
			issue := WorkflowSecurityIssue{
				ID:          fmt.Sprintf("%s-perm-%s", filePath, scope),
				Type:        IssueTypeExcessivePermissions,
				Severity:    SeverityHigh,
				Title:       fmt.Sprintf("Excessive %s permission", scope),
				Description: fmt.Sprintf("Workflow has write permission for %s, which may be unnecessary", scope),
				FilePath:    filePath,
				Suggestion:  fmt.Sprintf("Consider reducing %s permission to 'read' or 'none' if not needed", scope),
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

func (wa *WorkflowAuditor) analyzeActionSecurity(actionUses string) []string {
	risks := make([]string, 0)

	// Check for unpinned actions
	if !wa.isActionPinned(actionUses) {
		risks = append(risks, "Action is not pinned to a specific commit hash")
	}

	// Check for deprecated actions
	if wa.isActionDeprecated(actionUses) {
		risks = append(risks, "Action is deprecated and should be updated")
	}

	// Check for unverified actions
	if !wa.isActionVerified(actionUses) {
		risks = append(risks, "Action is not from a verified publisher")
	}

	// Check for high-risk actions
	if wa.isHighRiskAction(actionUses) {
		risks = append(risks, "Action has known security concerns")
	}

	return risks
}

func (wa *WorkflowAuditor) analyzeScriptSecurity(script string) []string {
	risks := make([]string, 0)

	// Check for potential code injection
	if wa.hasCodeInjectionRisk(script) {
		risks = append(risks, "Script may be vulnerable to code injection")
	}

	// Check for secret exposure
	if wa.hasSecretExposureRisk(script) {
		risks = append(risks, "Script may expose secrets in logs")
	}

	// Check for privilege escalation
	if wa.hasPrivilegeEscalationRisk(script) {
		risks = append(risks, "Script may attempt privilege escalation")
	}

	return risks
}

// Helper methods for analysis

func (wa *WorkflowAuditor) extractTriggers(on interface{}) []string {
	triggers := make([]string, 0)

	switch v := on.(type) {
	case string:
		triggers = append(triggers, v)
	case []interface{}:
		for _, trigger := range v {
			if str, ok := trigger.(string); ok {
				triggers = append(triggers, str)
			}
		}
	case map[string]interface{}:
		for key := range v {
			triggers = append(triggers, key)
		}
	}

	return triggers
}

func (wa *WorkflowAuditor) normalizePermissions(permissions map[string]interface{}) map[string]string {
	normalized := make(map[string]string)

	for scope, perm := range permissions {
		if str, ok := perm.(string); ok {
			normalized[scope] = str
		}
	}

	return normalized
}

func (wa *WorkflowAuditor) normalizeRunsOn(runsOn interface{}) string {
	switch v := runsOn.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str
			}
		}
	}
	return "unknown"
}

func (wa *WorkflowAuditor) normalizeEnvironment(env interface{}) string {
	switch v := env.(type) {
	case string:
		return v
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			return name
		}
	}
	return ""
}

func (wa *WorkflowAuditor) extractActionVersion(uses string) string {
	parts := strings.Split(uses, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return "latest"
}

func (wa *WorkflowAuditor) extractSecretsUsage(step Step) []string {
	secrets := make([]string, 0)

	// Check in step.Run
	if step.Run != "" {
		matches := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z_]+)\s*\}\}`).FindAllStringSubmatch(step.Run, -1)
		for _, match := range matches {
			if len(match) > 1 {
				secrets = append(secrets, match[1])
			}
		}
	}

	// Check in step.With
	for _, value := range step.With {
		matches := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z_]+)\s*\}\}`).FindAllStringSubmatch(value, -1)
		for _, match := range matches {
			if len(match) > 1 {
				secrets = append(secrets, match[1])
			}
		}
	}

	return wa.removeDuplicates(secrets)
}

func (wa *WorkflowAuditor) extractVariablesUsage(step Step) []string {
	variables := make([]string, 0)

	// Check in step.Run
	if step.Run != "" {
		matches := regexp.MustCompile(`\$\{\{\s*vars\.([A-Z_]+)\s*\}\}`).FindAllStringSubmatch(step.Run, -1)
		for _, match := range matches {
			if len(match) > 1 {
				variables = append(variables, match[1])
			}
		}
	}

	return wa.removeDuplicates(variables)
}

// Security check methods

func (wa *WorkflowAuditor) isHighRiskScope(scope string) bool {
	highRiskScopes := []string{"contents", "packages", "deployments", "security-events"}
	for _, risk := range highRiskScopes {
		if scope == risk {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) isActionPinned(uses string) bool {
	// Check if action is pinned to a commit hash (40 character hex string)
	parts := strings.Split(uses, "@")
	if len(parts) > 1 {
		version := parts[1]
		matched, _ := regexp.MatchString(`^[a-f0-9]{40}$`, version)
		return matched
	}
	return false
}

func (wa *WorkflowAuditor) isActionDeprecated(uses string) bool {
	// Mock implementation - in real world, this would check against a database
	deprecatedActions := []string{
		"actions/setup-node@v1",
		"actions/setup-python@v1",
		"docker/build-push-action@v1",
	}

	for _, deprecated := range deprecatedActions {
		if strings.HasPrefix(uses, strings.Split(deprecated, "@")[0]) {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) isActionVerified(uses string) bool {
	// Mock implementation - in real world, this would check GitHub's verified actions
	verifiedPublishers := []string{"actions/", "github/", "docker/", "azure/"}

	for _, verified := range verifiedPublishers {
		if strings.HasPrefix(uses, verified) {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) isHighRiskAction(uses string) bool {
	// Mock implementation - actions known to have security concerns
	highRiskActions := []string{
		"dangerous/action",
		"untrusted/tool",
	}

	for _, risk := range highRiskActions {
		if strings.HasPrefix(uses, risk) {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) hasCodeInjectionRisk(script string) bool {
	// Check for potential code injection patterns
	patterns := []string{
		`\$\{\{\s*github\.event\..*\}\}`,
		`\$\{\{\s*github\.head_ref\s*\}\}`,
		`eval\s*\(`,
		`exec\s*\(`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, script); matched {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) hasSecretExposureRisk(script string) bool {
	// Check for potential secret exposure
	patterns := []string{
		`echo.*\$\{\{\s*secrets\.`,
		`printf.*\$\{\{\s*secrets\.`,
		`cat.*\$\{\{\s*secrets\.`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, script); matched {
			return true
		}
	}
	return false
}

func (wa *WorkflowAuditor) hasPrivilegeEscalationRisk(script string) bool {
	// Check for privilege escalation attempts
	patterns := []string{
		`sudo\s+chmod`,
		`sudo\s+chown`,
		`sudo\s+su`,
		`chmod\s+777`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, script); matched {
			return true
		}
	}
	return false
}

// Analysis and calculation methods

func (wa *WorkflowAuditor) calculateStepRisk(risks []string) SecurityRiskLevel {
	if len(risks) == 0 {
		return SecurityRiskNone
	}

	criticalCount := 0
	highCount := 0

	for _, risk := range risks {
		if strings.Contains(risk, "code injection") || strings.Contains(risk, "secret exposure") {
			criticalCount++
		} else if strings.Contains(risk, "privilege escalation") || strings.Contains(risk, "not pinned") {
			highCount++
		}
	}

	if criticalCount > 0 {
		return SecurityRiskCritical
	} else if highCount > 0 {
		return SecurityRiskHigh
	} else if len(risks) > 2 {
		return SecurityRiskMedium
	} else {
		return SecurityRiskLow
	}
}

func (wa *WorkflowAuditor) calculateJobSecurityScore(job JobAuditInfo) int {
	score := 100

	for _, step := range job.Steps {
		switch step.SecurityRisk {
		case SecurityRiskCritical:
			score -= 25
		case SecurityRiskHigh:
			score -= 15
		case SecurityRiskMedium:
			score -= 10
		case SecurityRiskLow:
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (wa *WorkflowAuditor) calculateSecurityScore(audit *WorkflowFileAudit) int {
	if len(audit.Jobs) == 0 {
		return 100
	}

	totalScore := 0
	for _, job := range audit.Jobs {
		totalScore += job.SecurityScore
	}

	return totalScore / len(audit.Jobs)
}

func (wa *WorkflowAuditor) analyzePermissionUsage(files []WorkflowFileAudit) []WorkflowPermissionUsage {
	permissionMap := make(map[string]*WorkflowPermissionUsage)

	for _, file := range files {
		// Analyze workflow-level permissions
		for scope, permission := range file.Permissions {
			key := fmt.Sprintf("%s:%s", scope, permission)
			if usage, exists := permissionMap[key]; exists {
				usage.UsageCount++
				usage.WorkflowFiles = append(usage.WorkflowFiles, file.FilePath)
			} else {
				permissionMap[key] = &WorkflowPermissionUsage{
					Scope:         scope,
					Permission:    permission,
					UsageCount:    1,
					WorkflowFiles: []string{file.FilePath},
					Recommended:   wa.getRecommendedPermission(scope, permission),
				}
			}
		}

		// Analyze job-level permissions
		for _, job := range file.Jobs {
			for scope, permission := range job.Permissions {
				key := fmt.Sprintf("%s:%s", scope, permission)
				if usage, exists := permissionMap[key]; exists {
					usage.UsageCount++
					if !wa.contains(usage.WorkflowFiles, file.FilePath) {
						usage.WorkflowFiles = append(usage.WorkflowFiles, file.FilePath)
					}
				} else {
					permissionMap[key] = &WorkflowPermissionUsage{
						Scope:         scope,
						Permission:    permission,
						UsageCount:    1,
						WorkflowFiles: []string{file.FilePath},
						Recommended:   wa.getRecommendedPermission(scope, permission),
					}
				}
			}
		}
	}

	permissions := make([]WorkflowPermissionUsage, 0, len(permissionMap))
	for _, usage := range permissionMap {
		permissions = append(permissions, *usage)
	}

	return permissions
}

func (wa *WorkflowAuditor) analyzeActionUsage(files []WorkflowFileAudit) []ActionUsageInfo {
	actionMap := make(map[string]*ActionUsageInfo)

	for _, file := range files {
		for _, job := range file.Jobs {
			for _, step := range job.Steps {
				if step.Uses != "" {
					actionName := strings.Split(step.Uses, "@")[0]

					if usage, exists := actionMap[actionName]; exists {
						usage.UsageCount++
						if !wa.contains(usage.WorkflowFiles, file.FilePath) {
							usage.WorkflowFiles = append(usage.WorkflowFiles, file.FilePath)
						}
					} else {
						actionMap[actionName] = &ActionUsageInfo{
							ActionName:    actionName,
							Version:       step.ActionVersion,
							UsageCount:    1,
							WorkflowFiles: []string{file.FilePath},
							SecurityRisk:  step.SecurityRisk,
							IsVerified:    wa.isActionVerified(step.Uses),
							IsDeprecated:  wa.isActionDeprecated(step.Uses),
						}
					}
				}
			}
		}
	}

	actions := make([]ActionUsageInfo, 0, len(actionMap))
	for _, usage := range actionMap {
		actions = append(actions, *usage)
	}

	return actions
}

func (wa *WorkflowAuditor) generateSummary(result *WorkflowAuditResult) WorkflowAuditSummary {
	summary := WorkflowAuditSummary{
		TotalFiles:             len(result.AuditedFiles),
		PermissionDistribution: make(map[string]int),
		ActionRiskDistribution: make(map[SecurityRiskLevel]int),
	}

	totalScore := 0
	filesWithIssues := 0

	for _, file := range result.AuditedFiles {
		totalScore += file.SecurityScore

		if len(file.Issues) > 0 {
			filesWithIssues++
		}
	}

	summary.FilesWithIssues = filesWithIssues

	if summary.TotalFiles > 0 {
		summary.AverageSecurityScore = float64(totalScore) / float64(summary.TotalFiles)
	}

	// Count issues by severity
	for _, issue := range result.SecurityIssues {
		switch issue.Severity {
		case SeverityCritical:
			summary.CriticalIssues++
		case SeverityHigh:
			summary.HighRiskIssues++
		case SeverityMedium:
			summary.MediumRiskIssues++
		case SeverityLow:
			summary.LowRiskIssues++
		}
	}

	// Calculate compliance score
	if summary.TotalFiles > 0 {
		maxPossibleIssues := summary.TotalFiles * 10 // Assumption: max 10 issues per file
		actualIssues := len(result.SecurityIssues)
		summary.ComplianceScore = float64(maxPossibleIssues-actualIssues) / float64(maxPossibleIssues) * 100
		if summary.ComplianceScore < 0 {
			summary.ComplianceScore = 0
		}
	}

	return summary
}

// Helper methods

func (wa *WorkflowAuditor) getFileContent(ctx context.Context, organization, repository, filePath string) (string, error) {
	// Mock implementation - in real world, this would fetch from GitHub API
	return wa.getMockWorkflowContent(filepath.Base(filePath)), nil
}

func (wa *WorkflowAuditor) getMockWorkflowContent(filename string) string {
	switch filename {
	case "ci.yml":
		return `
name: CI
on: [push, pull_request]
permissions:
  contents: read
  packages: write
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm test
`
	case "release.yml":
		return `
name: Release
on:
  push:
    tags: ['v*']
permissions:
  contents: write
  packages: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v4
        with:
          push: true
          tags: myapp:latest
`
	default:
		return `
name: Default Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello World"
`
	}
}

func (wa *WorkflowAuditor) getRecommendedPermission(scope, permission string) string {
	if wa.isHighRiskScope(scope) && permission == "write" {
		return "read"
	}
	return permission
}

func (wa *WorkflowAuditor) determineIssueType(reasons []string) WorkflowIssueType {
	for _, reason := range reasons {
		if strings.Contains(reason, "code injection") {
			return IssueTypeCodeInjection
		} else if strings.Contains(reason, "secret") {
			return IssueTypeSecretExposure
		} else if strings.Contains(reason, "privilege") {
			return IssueTypePrivilegeEscalation
		} else if strings.Contains(reason, "not pinned") {
			return IssueTypeUnpinnedAction
		} else if strings.Contains(reason, "deprecated") {
			return IssueTypeDeprecatedAction
		} else if strings.Contains(reason, "verified") {
			return IssueTypeUnverifiedAction
		}
	}
	return IssueTypeExcessivePermissions
}

func (wa *WorkflowAuditor) mapRiskToSeverity(risk SecurityRiskLevel) SecurityIssueSeverity {
	switch risk {
	case SecurityRiskCritical:
		return SeverityCritical
	case SecurityRiskHigh:
		return SeverityHigh
	case SecurityRiskMedium:
		return SeverityMedium
	case SecurityRiskLow:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func (wa *WorkflowAuditor) generateSecuritySuggestion(reasons []string) string {
	suggestions := []string{
		"Pin actions to specific commit hashes",
		"Use minimal permissions required for the job",
		"Avoid exposing secrets in command outputs",
		"Validate all external inputs",
		"Use verified actions when possible",
	}

	if len(reasons) > 0 {
		return suggestions[0] // Return most relevant suggestion
	}
	return "Review workflow for security best practices"
}

func (wa *WorkflowAuditor) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (wa *WorkflowAuditor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

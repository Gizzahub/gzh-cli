package repoconfig

import (
	"time"
)

// RepositoryState represents the actual state of a repository
type RepositoryState struct {
	Name         string
	Private      bool
	Archived     bool
	HasIssues    bool
	HasWiki      bool
	HasProjects  bool
	HasDownloads bool

	// Branch protection
	BranchProtection map[string]BranchProtectionState

	// Security features
	VulnerabilityAlerts bool
	SecurityAdvisories  bool

	// Files present
	Files []string

	// Workflows
	Workflows []string

	// Last modified
	LastModified time.Time
}

// BranchProtectionState represents actual branch protection settings
type BranchProtectionState struct {
	Protected       bool
	RequiredReviews int
	EnforceAdmins   bool
	// Add other relevant fields as needed
}

// AuditReport represents a comprehensive compliance audit report
type AuditReport struct {
	Organization string              `yaml:"organization" json:"organization"`
	GeneratedAt  time.Time           `yaml:"generated_at" json:"generated_at"`
	PolicyFile   string              `yaml:"policy_file" json:"policy_file"`
	Summary      AuditSummary        `yaml:"summary" json:"summary"`
	Policies     []PolicyAuditResult `yaml:"policies" json:"policies"`
	Repositories []RepoAuditResult   `yaml:"repositories" json:"repositories"`
}

// AuditSummary provides high-level compliance metrics
type AuditSummary struct {
	TotalRepositories     int     `yaml:"total_repositories" json:"total_repositories"`
	AuditedRepositories   int     `yaml:"audited_repositories" json:"audited_repositories"`
	CompliantRepositories int     `yaml:"compliant_repositories" json:"compliant_repositories"`
	CompliancePercentage  float64 `yaml:"compliance_percentage" json:"compliance_percentage"`
	TotalPolicies         int     `yaml:"total_policies" json:"total_policies"`
	TotalViolations       int     `yaml:"total_violations" json:"total_violations"`
	TotalExceptions       int     `yaml:"total_exceptions" json:"total_exceptions"`
	ActiveExceptions      int     `yaml:"active_exceptions" json:"active_exceptions"`
}

// PolicyAuditResult represents audit results for a specific policy
type PolicyAuditResult struct {
	PolicyName           string            `yaml:"policy_name" json:"policy_name"`
	Description          string            `yaml:"description" json:"description"`
	Rules                []RuleAuditResult `yaml:"rules" json:"rules"`
	CompliantRepos       int               `yaml:"compliant_repos" json:"compliant_repos"`
	ViolatingRepos       int               `yaml:"violating_repos" json:"violating_repos"`
	ExemptedRepos        int               `yaml:"exempted_repos" json:"exempted_repos"`
	CompliancePercentage float64           `yaml:"compliance_percentage" json:"compliance_percentage"`
}

// RuleAuditResult represents audit results for a specific rule within a policy
type RuleAuditResult struct {
	RuleName       string   `yaml:"rule_name" json:"rule_name"`
	Type           string   `yaml:"type" json:"type"`
	Enforcement    string   `yaml:"enforcement" json:"enforcement"`
	ViolatingRepos []string `yaml:"violating_repos" json:"violating_repos"`
	ExemptedRepos  []string `yaml:"exempted_repos" json:"exempted_repos"`
}

// RepoAuditResult represents audit results for a specific repository
type RepoAuditResult struct {
	Repository   string            `yaml:"repository" json:"repository"`
	Template     string            `yaml:"template,omitempty" json:"template,omitempty"`
	Compliant    bool              `yaml:"compliant" json:"compliant"`
	Violations   []PolicyViolation `yaml:"violations,omitempty" json:"violations,omitempty"`
	Exceptions   []PolicyException `yaml:"exceptions,omitempty" json:"exceptions,omitempty"`
	LastModified time.Time         `yaml:"last_modified,omitempty" json:"last_modified,omitempty"`
}

// PolicyViolation represents a specific policy violation
type PolicyViolation struct {
	PolicyName  string      `yaml:"policy" json:"policy"`
	RuleName    string      `yaml:"rule" json:"rule"`
	Type        string      `yaml:"type" json:"type"`
	Expected    interface{} `yaml:"expected" json:"expected"`
	Actual      interface{} `yaml:"actual,omitempty" json:"actual,omitempty"`
	Severity    string      `yaml:"severity" json:"severity"`
	Message     string      `yaml:"message" json:"message"`
	Remediation string      `yaml:"remediation,omitempty" json:"remediation,omitempty"`
}

// PolicyException represents an exception to a policy rule
type PolicyException struct {
	PolicyName  string     `yaml:"policy" json:"policy"`
	RuleName    string     `yaml:"rule" json:"rule"`
	Reason      string     `yaml:"reason" json:"reason"`
	ApprovedBy  string     `yaml:"approved_by" json:"approved_by"`
	ApprovedAt  time.Time  `yaml:"approved_at" json:"approved_at"`
	ExpiresAt   *time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
	JiraTicket  string     `yaml:"jira_ticket,omitempty" json:"jira_ticket,omitempty"`
	ReviewNotes string     `yaml:"review_notes,omitempty" json:"review_notes,omitempty"`
}

// IsExceptionActive checks if an exception is currently active
func (e PolicyException) IsExceptionActive() bool {
	if e.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(*e.ExpiresAt)
}
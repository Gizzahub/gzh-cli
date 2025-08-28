// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"time"
)

// RepositoryState represents the actual state of a repository.
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

// BranchProtectionState represents actual branch protection settings.
type BranchProtectionState struct {
	Protected       bool
	RequiredReviews int
	EnforceAdmins   bool
	// Add other relevant fields as needed
}

// AuditReport represents a comprehensive compliance audit report.
type AuditReport struct {
	Organization string              `yaml:"organization" json:"organization"`
	GeneratedAt  time.Time           `yaml:"generatedAt" json:"generated_at"`
	PolicyFile   string              `yaml:"policyFile" json:"policy_file"`
	Summary      AuditSummary        `yaml:"summary" json:"summary"`
	Policies     []PolicyAuditResult `yaml:"policies" json:"policies"`
	Repositories []RepoAuditResult   `yaml:"repositories" json:"repositories"`
}

// AuditSummary provides high-level compliance metrics.
type AuditSummary struct {
	TotalRepositories     int     `yaml:"totalRepositories" json:"totalRepositories"`
	AuditedRepositories   int     `yaml:"auditedRepositories" json:"auditedRepositories"`
	CompliantRepositories int     `yaml:"compliantRepositories" json:"compliantRepositories"`
	CompliancePercentage  float64 `yaml:"compliancePercentage" json:"compliancePercentage"`
	TotalPolicies         int     `yaml:"totalPolicies" json:"totalPolicies"`
	TotalViolations       int     `yaml:"totalViolations" json:"totalViolations"`
	TotalExceptions       int     `yaml:"totalExceptions" json:"totalExceptions"`
	ActiveExceptions      int     `yaml:"activeExceptions" json:"activeExceptions"`
}

// PolicyAuditResult represents audit results for a specific policy.
type PolicyAuditResult struct {
	PolicyName           string            `yaml:"policyName" json:"policyName"`
	Description          string            `yaml:"description" json:"description"`
	Rules                []RuleAuditResult `yaml:"rules" json:"rules"`
	CompliantRepos       int               `yaml:"compliantRepos" json:"compliantRepos"`
	ViolatingRepos       int               `yaml:"violatingRepos" json:"violatingRepos"`
	ExemptedRepos        int               `yaml:"exemptedRepos" json:"exemptedRepos"`
	CompliancePercentage float64           `yaml:"compliancePercentage" json:"compliancePercentage"`
}

// RuleAuditResult represents audit results for a specific rule within a policy.
type RuleAuditResult struct {
	RuleName       string   `yaml:"ruleName" json:"ruleName"`
	Type           string   `yaml:"type" json:"type"`
	Enforcement    string   `yaml:"enforcement" json:"enforcement"`
	ViolatingRepos []string `yaml:"violatingRepos" json:"violatingRepos"`
	ExemptedRepos  []string `yaml:"exemptedRepos" json:"exemptedRepos"`
}

// RepoAuditResult represents audit results for a specific repository.
type RepoAuditResult struct {
	Repository   string            `yaml:"repository" json:"repository"`
	Template     string            `yaml:"template,omitempty" json:"template,omitempty"`
	Compliant    bool              `yaml:"compliant" json:"compliant"`
	Violations   []PolicyViolation `yaml:"violations,omitempty" json:"violations,omitempty"`
	Exceptions   []PolicyException `yaml:"exceptions,omitempty" json:"exceptions,omitempty"`
	LastModified time.Time         `yaml:"lastModified,omitempty" json:"lastModified,omitempty"`
}

// PolicyViolation represents a specific policy violation.
type PolicyViolation struct {
	PolicyName  string      `yaml:"policyName" json:"policyName"`
	RuleName    string      `yaml:"ruleName" json:"ruleName"`
	Type        string      `yaml:"type" json:"type"`
	Expected    interface{} `yaml:"expected" json:"expected"`
	Actual      interface{} `yaml:"actual,omitempty" json:"actual,omitempty"`
	Severity    string      `yaml:"severity" json:"severity"`
	Message     string      `yaml:"message" json:"message"`
	Remediation string      `yaml:"remediation,omitempty" json:"remediation,omitempty"`
}

// PolicyException represents an exception to a policy rule.
type PolicyException struct {
	PolicyName  string     `yaml:"policyName" json:"policyName"`
	RuleName    string     `yaml:"ruleName" json:"ruleName"`
	Reason      string     `yaml:"reason" json:"reason"`
	ApprovedBy  string     `yaml:"approvedBy" json:"approvedBy"`
	ApprovedAt  time.Time  `yaml:"approvedAt" json:"approvedAt"`
	ExpiresAt   *time.Time `yaml:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	JiraTicket  string     `yaml:"jiraTicket,omitempty" json:"jiraTicket,omitempty"`
	ReviewNotes string     `yaml:"reviewNotes,omitempty" json:"reviewNotes,omitempty"`
}

// IsExceptionActive checks if an exception is currently active.
func (e PolicyException) IsExceptionActive() bool {
	if e.ExpiresAt == nil {
		return true
	}

	return time.Now().Before(*e.ExpiresAt)
}

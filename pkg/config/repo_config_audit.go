// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"strings"
	"time"
)

// AuditReport represents a comprehensive compliance audit report.
type AuditReport struct {
	Organization string              `yaml:"organization" json:"organization"`
	GeneratedAt  time.Time           `yaml:"generatedAt" json:"generatedAt"`
	PolicyFile   string              `yaml:"policyFile" json:"policyFile"`
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
	PolicyName  string      `yaml:"policy" json:"policy"`
	RuleName    string      `yaml:"rule" json:"rule"`
	Type        string      `yaml:"type" json:"type"`
	Expected    interface{} `yaml:"expected" json:"expected"`
	Actual      interface{} `yaml:"actual,omitempty" json:"actual,omitempty"`
	Severity    string      `yaml:"severity" json:"severity"`
	Message     string      `yaml:"message" json:"message"`
	Remediation string      `yaml:"remediation,omitempty" json:"remediation,omitempty"`
}

// RunComplianceAudit performs a compliance audit against configured policies.
func (rc *RepoConfig) RunComplianceAudit(actualRepos map[string]RepositoryState) (*AuditReport, error) { //nolint:gocognit // Complex compliance audit logic with multiple policy checks
	report := &AuditReport{
		Organization: rc.Organization,
		GeneratedAt:  time.Now(),
		Summary:      AuditSummary{},
		Policies:     []PolicyAuditResult{},
		Repositories: []RepoAuditResult{},
	}

	// Initialize policy results
	policyResults := make(map[string]*PolicyAuditResult)
	for policyName, policy := range rc.Policies {
		policyResult := &PolicyAuditResult{
			PolicyName:  policyName,
			Description: policy.Description,
			Rules:       []RuleAuditResult{},
		}

		// Initialize rule results
		for ruleName, rule := range policy.Rules {
			policyResult.Rules = append(policyResult.Rules, RuleAuditResult{
				RuleName:       ruleName,
				Type:           rule.Type,
				Enforcement:    rule.Enforcement,
				ViolatingRepos: []string{},
				ExemptedRepos:  []string{},
			})
		}

		policyResults[policyName] = policyResult
	}

	// Audit each repository
	for repoName, repoState := range actualRepos {
		repoResult := RepoAuditResult{
			Repository:   repoName,
			Compliant:    true,
			Violations:   []PolicyViolation{},
			Exceptions:   []PolicyException{},
			LastModified: repoState.LastModified,
		}

		// Get effective configuration and exceptions for this repository
		settings, security, permissions, exceptions, err := rc.GetEffectiveConfig(repoName)
		if err != nil {
			continue // Skip this repo if we can't get config
		}

		repoResult.Exceptions = exceptions

		// Check each policy
		for policyName, policy := range rc.Policies {
			for ruleName, rule := range policy.Rules {
				// Check if there's an active exception for this rule
				hasException := false

				for _, exc := range exceptions {
					if exc.PolicyName == policyName && exc.RuleName == ruleName && exc.IsExceptionActive() {
						hasException = true
						// Update policy results
						for i, r := range policyResults[policyName].Rules {
							if r.RuleName == ruleName {
								policyResults[policyName].Rules[i].ExemptedRepos = append(
									policyResults[policyName].Rules[i].ExemptedRepos, repoName)

								break
							}
						}

						break
					}
				}

				if hasException {
					policyResults[policyName].ExemptedRepos++
					continue
				}

				// Check compliance based on rule type
				violation := checkRuleCompliance(rule, settings, security, permissions, repoState)
				if violation != nil {
					violation.PolicyName = policyName
					violation.RuleName = ruleName
					violation.Severity = getSeverity(rule.Enforcement)
					violation.Message = rule.Message

					repoResult.Compliant = false
					repoResult.Violations = append(repoResult.Violations, *violation)

					// Update policy results
					for i, r := range policyResults[policyName].Rules {
						if r.RuleName == ruleName {
							policyResults[policyName].Rules[i].ViolatingRepos = append(
								policyResults[policyName].Rules[i].ViolatingRepos, repoName)

							break
						}
					}
				}
			}
		}

		// Update summary
		report.Summary.TotalRepositories++

		report.Summary.AuditedRepositories++
		if repoResult.Compliant {
			report.Summary.CompliantRepositories++
		}

		report.Summary.TotalViolations += len(repoResult.Violations)

		report.Summary.TotalExceptions += len(repoResult.Exceptions)
		for _, exc := range repoResult.Exceptions {
			if exc.IsExceptionActive() {
				report.Summary.ActiveExceptions++
			}
		}

		report.Repositories = append(report.Repositories, repoResult)
	}

	// Calculate policy compliance percentages
	for _, policyResult := range policyResults {
		total := report.Summary.AuditedRepositories
		compliant := total - policyResult.ViolatingRepos - policyResult.ExemptedRepos
		policyResult.CompliantRepos = compliant

		if total > 0 {
			// Compliance percentage excludes exempted repos
			nonExempted := total - policyResult.ExemptedRepos
			if nonExempted > 0 {
				policyResult.CompliancePercentage = float64(compliant) / float64(nonExempted) * 100
			}
		}

		report.Policies = append(report.Policies, *policyResult)
	}

	// Update summary
	report.Summary.TotalPolicies = len(rc.Policies)
	if report.Summary.AuditedRepositories > 0 {
		report.Summary.CompliancePercentage = float64(report.Summary.CompliantRepositories) /
			float64(report.Summary.AuditedRepositories) * 100
	}

	return report, nil
}

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

// checkRuleCompliance checks if a repository complies with a specific rule.
func checkRuleCompliance(rule PolicyRule, settings *RepoSettings, security *SecuritySettings, //nolint:gocognit // Complex rule compliance checking with multiple policy types
	permissions *PermissionSettings, state RepositoryState,
) *PolicyViolation {
	switch rule.Type {
	case "visibility":
		expected, ok := rule.Value.(string)
		if !ok {
			return nil // Skip invalid rule value
		}

		actual := "public"
		if state.Private {
			actual = "private"
		}

		if expected != actual {
			return &PolicyViolation{
				Type:     rule.Type,
				Expected: expected,
				Actual:   actual,
			}
		}

	case "branch_protection":
		if expectedBool, ok := rule.Value.(bool); ok && expectedBool {
			// Check if main branch is protected
			if mainProtection, exists := state.BranchProtection["main"]; !exists || !mainProtection.Protected {
				return &PolicyViolation{
					Type:        rule.Type,
					Expected:    true,
					Actual:      false,
					Remediation: "Enable branch protection for the main branch",
				}
			}
		}

	case "min_reviews":
		if expectedReviews, ok := getIntValue(rule.Value); ok {
			// Check main branch review requirements
			if mainProtection, exists := state.BranchProtection["main"]; exists {
				if mainProtection.RequiredReviews < expectedReviews {
					return &PolicyViolation{
						Type:        rule.Type,
						Expected:    expectedReviews,
						Actual:      mainProtection.RequiredReviews,
						Remediation: fmt.Sprintf("Increase required reviewers to %d", expectedReviews),
					}
				}
			} else {
				return &PolicyViolation{
					Type:        rule.Type,
					Expected:    expectedReviews,
					Actual:      0,
					Remediation: "Enable branch protection with required reviews",
				}
			}
		}

	case "file_exists":
		if expectedFile, ok := rule.Value.(string); ok {
			found := false

			for _, file := range state.Files {
				if strings.EqualFold(file, expectedFile) {
					found = true
					break
				}
			}

			if !found {
				return &PolicyViolation{
					Type:        rule.Type,
					Expected:    expectedFile,
					Actual:      "not found",
					Remediation: fmt.Sprintf("Add required file: %s", expectedFile),
				}
			}
		}

	case "workflow_exists":
		if expectedWorkflow, ok := rule.Value.(string); ok {
			found := false

			workflowName := strings.TrimPrefix(expectedWorkflow, ".github/workflows/")
			for _, workflow := range state.Workflows {
				if strings.EqualFold(workflow, workflowName) {
					found = true
					break
				}
			}

			if !found {
				return &PolicyViolation{
					Type:        rule.Type,
					Expected:    expectedWorkflow,
					Actual:      "not found",
					Remediation: fmt.Sprintf("Add required workflow: %s", expectedWorkflow),
				}
			}
		}

	case "security_feature":
		if feature, ok := rule.Value.(string); ok {
			enabled := false

			switch feature {
			case "vulnerability_alerts":
				enabled = state.VulnerabilityAlerts
			case "security_advisories":
				enabled = state.SecurityAdvisories
			}

			if !enabled {
				return &PolicyViolation{
					Type:        rule.Type,
					Expected:    fmt.Sprintf("%s enabled", feature),
					Actual:      "disabled",
					Remediation: fmt.Sprintf("Enable %s in repository settings", feature),
				}
			}
		}
	}

	return nil
}

// getIntValue safely extracts an int value from an interface{}.
func getIntValue(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

// getSeverity determines severity level from enforcement.
func getSeverity(enforcement string) string {
	switch strings.ToLower(enforcement) {
	case "required":
		return "critical"
	case "recommended":
		return "medium"
	default:
		return "low"
	}
}

// GenerateAuditSummary creates a human-readable summary of the audit report.
func (ar *AuditReport) GenerateAuditSummary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Compliance Audit Report for %s\n\n", ar.Organization))
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", ar.GeneratedAt.Format(time.RFC3339)))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- Total Repositories: %d\n", ar.Summary.TotalRepositories))
	sb.WriteString(fmt.Sprintf("- Audited Repositories: %d\n", ar.Summary.AuditedRepositories))
	sb.WriteString(fmt.Sprintf("- Compliant Repositories: %d (%.1f%%)\n",
		ar.Summary.CompliantRepositories, ar.Summary.CompliancePercentage))
	sb.WriteString(fmt.Sprintf("- Total Violations: %d\n", ar.Summary.TotalViolations))
	sb.WriteString(fmt.Sprintf("- Active Exceptions: %d\n", ar.Summary.ActiveExceptions))

	sb.WriteString("\n## Policy Compliance\n\n")

	for _, policy := range ar.Policies {
		sb.WriteString(fmt.Sprintf("### %s\n", policy.PolicyName))
		sb.WriteString(fmt.Sprintf("%s\n\n", policy.Description))
		sb.WriteString(fmt.Sprintf("- Compliance: %.1f%%\n", policy.CompliancePercentage))
		sb.WriteString(fmt.Sprintf("- Compliant: %d repos\n", policy.CompliantRepos))
		sb.WriteString(fmt.Sprintf("- Violating: %d repos\n", policy.ViolatingRepos))
		sb.WriteString(fmt.Sprintf("- Exempted: %d repos\n\n", policy.ExemptedRepos))
	}

	// List non-compliant repositories
	nonCompliant := 0

	for _, repo := range ar.Repositories {
		if !repo.Compliant {
			nonCompliant++
		}
	}

	if nonCompliant > 0 {
		sb.WriteString("\n## Non-Compliant Repositories\n\n")

		for _, repo := range ar.Repositories {
			if !repo.Compliant {
				sb.WriteString(fmt.Sprintf("### %s\n", repo.Repository))

				for _, violation := range repo.Violations {
					sb.WriteString(fmt.Sprintf("- **%s/%s**: %s\n",
						violation.PolicyName, violation.RuleName, violation.Message))

					if violation.Remediation != "" {
						sb.WriteString(fmt.Sprintf("  - Remediation: %s\n", violation.Remediation))
					}
				}

				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

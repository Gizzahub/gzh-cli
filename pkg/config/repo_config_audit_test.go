//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunComplianceAudit(t *testing.T) {
	// Create a test configuration with policies
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Policies: map[string]*PolicyTemplate{
			"security": {
				Description: "Security requirements",
				Rules: map[string]PolicyRule{
					"private_repos": {
						Type:        "visibility",
						Value:       "private",
						Enforcement: "required",
						Message:     "All repos must be private",
					},
					"branch_protection": {
						Type:        "branch_protection",
						Value:       true,
						Enforcement: "required",
						Message:     "Branch protection required",
					},
					"min_reviewers": {
						Type:        "min_reviews",
						Value:       2,
						Enforcement: "required",
						Message:     "Minimum 2 reviewers required",
					},
				},
			},
			"compliance": {
				Description: "Compliance requirements",
				Rules: map[string]PolicyRule{
					"license_file": {
						Type:        "file_exists",
						Value:       "LICENSE",
						Enforcement: "required",
						Message:     "LICENSE file required",
					},
					"security_workflow": {
						Type:        "workflow_exists",
						Value:       ".github/workflows/security.yml",
						Enforcement: "recommended",
						Message:     "Security workflow recommended",
					},
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name: "exempt-repo",
					Exceptions: []PolicyException{
						{
							PolicyName: "security",
							RuleName:   "private_repos",
							Reason:     "Public demo repository",
							ApprovedBy: "cto@company.com",
						},
					},
				},
			},
		},
	}

	// Create test repository states
	repoStates := map[string]RepositoryState{
		"compliant-repo": {
			Name:      "compliant-repo",
			Private:   true,
			Files:     []string{"README.md", "LICENSE"},
			Workflows: []string{"security"},
			BranchProtection: map[string]BranchProtectionState{
				"main": {
					Protected:       true,
					RequiredReviews: 2,
					EnforceAdmins:   true,
				},
			},
			LastModified: time.Now(),
		},
		// 각 리포는 "어떤 규칙을 어기는가"를 정해 두고 그 외 규칙은 모두
		// 만족시킨다. 아래 두 리포에 Workflows를 채우는 것도 그래서다 --
		// security_workflow 규칙은 이들의 검증 대상이 아닌데, 비워 두면
		// 의도하지 않은 위반이 하나씩 더 붙어 위반 개수 어설션이 어긋난다.
		"non-compliant-repo": {
			Name:      "non-compliant-repo",
			Private:   false,                 // Violates private_repos rule
			Files:     []string{"README.md"}, // Missing LICENSE
			Workflows: []string{"security"},
			BranchProtection: map[string]BranchProtectionState{
				"main": {
					Protected:       true,
					RequiredReviews: 1, // Violates min_reviewers
				},
			},
			LastModified: time.Now(),
		},
		"exempt-repo": {
			Name:      "exempt-repo",
			Private:   false, // Has exception for this
			Files:     []string{"LICENSE"},
			Workflows: []string{"security"},
			BranchProtection: map[string]BranchProtectionState{
				"main": {
					Protected:       true,
					RequiredReviews: 2,
				},
			},
			LastModified: time.Now(),
		},
		"unprotected-repo": {
			Name:             "unprotected-repo",
			Private:          true,
			Files:            []string{"LICENSE"},
			BranchProtection: map[string]BranchProtectionState{
				// No branch protection
			},
			LastModified: time.Now(),
		},
	}

	// Run the audit
	report, err := config.RunComplianceAudit(repoStates)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Check summary
	assert.Equal(t, "test-org", report.Organization)
	assert.Equal(t, 4, report.Summary.TotalRepositories)
	assert.Equal(t, 4, report.Summary.AuditedRepositories)
	// 승인된 예외가 걸린 위반은 위반으로 기록되지 않으므로 exempt-repo도
	// compliant다 -- 예외는 "이 리포는 이 규칙을 어겨도 된다"는 승인이고,
	// 아래 156줄의 어설션도 같은 것을 요구한다. 예전 기대값 1/25%는 이와
	// 모순됐는데, exempt-repo가 별개 버그(워크플로 표기 불일치)로 non-compliant
	// 판정을 받고 있어서 우연히 맞아떨어졌다. 예외를 쓴 리포는 정책별
	// ExemptedRepos로 따로 집계되므로 보고서에서 구분은 유지된다.
	assert.Equal(t, 2, report.Summary.CompliantRepositories)   // compliant-repo, exempt-repo
	assert.Equal(t, 50.0, report.Summary.CompliancePercentage) // 2/4 = 50%
	assert.Equal(t, 1, report.Summary.TotalExceptions)
	assert.Equal(t, 1, report.Summary.ActiveExceptions)

	// Check repository results
	repoResults := make(map[string]RepoAuditResult)
	for _, r := range report.Repositories {
		repoResults[r.Repository] = r
	}

	// Check compliant repo
	assert.True(t, repoResults["compliant-repo"].Compliant)
	assert.Empty(t, repoResults["compliant-repo"].Violations)

	// Check non-compliant repo
	assert.False(t, repoResults["non-compliant-repo"].Compliant)
	assert.Len(t, repoResults["non-compliant-repo"].Violations, 3) // 3 violations

	// Check exempt repo
	assert.True(t, repoResults["exempt-repo"].Compliant)
	assert.Len(t, repoResults["exempt-repo"].Exceptions, 1)

	// Check unprotected repo
	assert.False(t, repoResults["unprotected-repo"].Compliant)
	assert.NotEmpty(t, repoResults["unprotected-repo"].Violations)
}

func TestCheckRuleCompliance(t *testing.T) {
	tests := []struct {
		name            string
		rule            PolicyRule
		state           RepositoryState
		expectViolation bool
		violationType   string
	}{
		{
			name: "visibility compliant - private",
			rule: PolicyRule{
				Type:  "visibility",
				Value: "private",
			},
			state: RepositoryState{
				Private: true,
			},
			expectViolation: false,
		},
		{
			name: "visibility violation - should be private",
			rule: PolicyRule{
				Type:  "visibility",
				Value: "private",
			},
			state: RepositoryState{
				Private: false,
			},
			expectViolation: true,
			violationType:   "visibility",
		},
		{
			name: "branch protection compliant",
			rule: PolicyRule{
				Type:  "branch_protection",
				Value: true,
			},
			state: RepositoryState{
				BranchProtection: map[string]BranchProtectionState{
					"main": {Protected: true},
				},
			},
			expectViolation: false,
		},
		{
			name: "branch protection violation",
			rule: PolicyRule{
				Type:  "branch_protection",
				Value: true,
			},
			state: RepositoryState{
				BranchProtection: map[string]BranchProtectionState{},
			},
			expectViolation: true,
			violationType:   "branch_protection",
		},
		{
			name: "min reviews compliant",
			rule: PolicyRule{
				Type:  "min_reviews",
				Value: 2,
			},
			state: RepositoryState{
				BranchProtection: map[string]BranchProtectionState{
					"main": {
						Protected:       true,
						RequiredReviews: 2,
					},
				},
			},
			expectViolation: false,
		},
		{
			name: "min reviews violation",
			rule: PolicyRule{
				Type:  "min_reviews",
				Value: 2,
			},
			state: RepositoryState{
				BranchProtection: map[string]BranchProtectionState{
					"main": {
						Protected:       true,
						RequiredReviews: 1,
					},
				},
			},
			expectViolation: true,
			violationType:   "min_reviews",
		},
		{
			name: "file exists compliant",
			rule: PolicyRule{
				Type:  "file_exists",
				Value: "LICENSE",
			},
			state: RepositoryState{
				Files: []string{"README.md", "LICENSE"},
			},
			expectViolation: false,
		},
		{
			name: "file exists violation",
			rule: PolicyRule{
				Type:  "file_exists",
				Value: "LICENSE",
			},
			state: RepositoryState{
				Files: []string{"README.md"},
			},
			expectViolation: true,
			violationType:   "file_exists",
		},
		{
			name: "workflow exists compliant",
			rule: PolicyRule{
				Type:  "workflow_exists",
				Value: ".github/workflows/security.yml",
			},
			state: RepositoryState{
				Workflows: []string{"security", "test"},
			},
			expectViolation: false,
		},
		{
			name: "security feature compliant",
			rule: PolicyRule{
				Type:  "security_feature",
				Value: "vulnerability_alerts",
			},
			state: RepositoryState{
				VulnerabilityAlerts: true,
			},
			expectViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := checkRuleCompliance(tt.rule, nil, nil, nil, tt.state)

			if tt.expectViolation {
				require.NotNil(t, violation)
				assert.Equal(t, tt.violationType, violation.Type)
			} else {
				assert.Nil(t, violation)
			}
		})
	}
}

func TestAuditReportGeneration(t *testing.T) {
	report := &AuditReport{
		Organization: "test-org",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:     10,
			AuditedRepositories:   10,
			CompliantRepositories: 7,
			CompliancePercentage:  70.0,
			TotalPolicies:         2,
			TotalViolations:       5,
			ActiveExceptions:      2,
		},
		Policies: []PolicyAuditResult{
			{
				PolicyName:           "security",
				Description:          "Security requirements",
				CompliantRepos:       7,
				ViolatingRepos:       2,
				ExemptedRepos:        1,
				CompliancePercentage: 77.8,
			},
		},
		Repositories: []RepoAuditResult{
			{
				Repository: "non-compliant-repo",
				Compliant:  false,
				Violations: []PolicyViolation{
					{
						PolicyName:  "security",
						RuleName:    "private_repos",
						Type:        "visibility",
						Expected:    "private",
						Actual:      "public",
						Severity:    "critical",
						Message:     "Repository must be private",
						Remediation: "Change repository visibility to private",
					},
				},
			},
		},
	}

	// Test audit summary generation
	summary := report.GenerateAuditSummary()
	assert.Contains(t, summary, "Compliance Audit Report for test-org")
	assert.Contains(t, summary, "Total Repositories: 10")
	assert.Contains(t, summary, "Compliant Repositories: 7 (70.0%)")
	assert.Contains(t, summary, "security")
	assert.Contains(t, summary, "non-compliant-repo")
	assert.Contains(t, summary, "Repository must be private")
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		enforcement string
		expected    string
	}{
		{"required", "critical"},
		{"REQUIRED", "critical"},
		{"recommended", "medium"},
		{"optional", "low"},
		{"unknown", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.enforcement, func(t *testing.T) {
			assert.Equal(t, tt.expected, getSeverity(tt.enforcement))
		})
	}
}

func TestPolicyExceptionInAudit(t *testing.T) {
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Policies: map[string]*PolicyTemplate{
			"test_policy": {
				Rules: map[string]PolicyRule{
					"test_rule": {
						Type:        "visibility",
						Value:       "private",
						Enforcement: "required",
					},
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name: "public-demo",
					Exceptions: []PolicyException{
						{
							PolicyName: "test_policy",
							RuleName:   "test_rule",
							Reason:     "Approved public demo",
							ApprovedBy: "cto@company.com",
						},
					},
				},
			},
		},
	}

	repoStates := map[string]RepositoryState{
		"public-demo": {
			Name:    "public-demo",
			Private: false, // Would violate without exception
		},
		"other-public": {
			Name:    "other-public",
			Private: false, // Will violate
		},
	}

	report, err := config.RunComplianceAudit(repoStates)
	require.NoError(t, err)

	// Check that public-demo is compliant due to exception
	var publicDemo, otherPublic *RepoAuditResult

	for i := range report.Repositories {
		switch report.Repositories[i].Repository {
		case "public-demo":
			publicDemo = &report.Repositories[i]
		case "other-public":
			otherPublic = &report.Repositories[i]
		}
	}

	require.NotNil(t, publicDemo)
	assert.True(t, publicDemo.Compliant)
	assert.Empty(t, publicDemo.Violations)
	assert.Len(t, publicDemo.Exceptions, 1)

	require.NotNil(t, otherPublic)
	assert.False(t, otherPublic.Compliant)
	assert.NotEmpty(t, otherPublic.Violations)
}

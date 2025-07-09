package repoconfig

import (
	"strings"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyTemplates(t *testing.T) {
	// Test that predefined policy templates are loaded correctly
	templates := config.GetPredefinedPolicyTemplates()

	// Verify we have essential policies
	assert.NotEmpty(t, templates)

	// Check branch protection policy
	branchProtection, exists := templates["branch_protection"]
	require.True(t, exists, "branch_protection policy should exist")
	assert.Equal(t, "security", branchProtection.Group)
	assert.Equal(t, "critical", branchProtection.Severity)
	assert.NotEmpty(t, branchProtection.Rules)

	// Check vulnerability management policy
	vulnMgmt, exists := templates["vulnerability_management"]
	require.True(t, exists, "vulnerability_management policy should exist")
	assert.Equal(t, "security", vulnMgmt.Group)
}

func TestPolicyPresets(t *testing.T) {
	// Test that policy presets are loaded correctly
	presets := config.GetPolicyPresets()

	// Verify we have essential presets
	assert.NotEmpty(t, presets)

	// Check SOC2 preset
	soc2, exists := presets["soc2"]
	require.True(t, exists, "soc2 preset should exist")
	assert.Equal(t, "SOC2", soc2.Framework)
	assert.NotEmpty(t, soc2.Policies)
	assert.Contains(t, soc2.Policies, "branch_protection")
	assert.Contains(t, soc2.Policies, "audit_logging")

	// Check ISO27001 preset
	iso, exists := presets["iso27001"]
	require.True(t, exists, "iso27001 preset should exist")
	assert.Equal(t, "ISO27001", iso.Framework)
	assert.Equal(t, "2022", iso.Version)
}

func TestPolicyGroups(t *testing.T) {
	// Test that policy groups are configured correctly
	groups := config.GetPolicyGroups()

	// Verify we have essential groups
	assert.NotEmpty(t, groups)

	// Check security group
	security, exists := groups["security"]
	require.True(t, exists, "security group should exist")
	assert.Equal(t, 0.4, security.Weight) // 40% weight
	assert.True(t, security.Required)
	assert.Contains(t, security.Policies, "branch_protection")

	// Check compliance group
	compliance, exists := groups["compliance"]
	require.True(t, exists, "compliance group should exist")
	assert.Equal(t, 0.35, compliance.Weight) // 35% weight
	assert.False(t, compliance.Required)

	// Verify total weights add up to 1.0
	totalWeight := 0.0
	for _, group := range groups {
		totalWeight += group.Weight
	}
	assert.Equal(t, 1.0, totalWeight, "Total group weights should sum to 1.0")
}

func TestRepositoryFiltering(t *testing.T) {
	tests := []struct {
		name     string
		repo     config.RepositoryState
		opts     AuditOptions
		expected bool
	}{
		{
			name: "filter by visibility - private",
			repo: config.RepositoryState{
				Name:    "test-repo",
				Private: true,
			},
			opts: AuditOptions{
				FilterVisibility: "private",
			},
			expected: true,
		},
		{
			name: "filter by visibility - public",
			repo: config.RepositoryState{
				Name:    "test-repo",
				Private: true,
			},
			opts: AuditOptions{
				FilterVisibility: "public",
			},
			expected: false,
		},
		{
			name: "filter by pattern - match",
			repo: config.RepositoryState{
				Name: "api-service",
			},
			opts: AuditOptions{
				FilterPattern: "api-.*",
			},
			expected: true,
		},
		{
			name: "filter by pattern - no match",
			repo: config.RepositoryState{
				Name: "web-frontend",
			},
			opts: AuditOptions{
				FilterPattern: "api-.*",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeRepo(tt.repo, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRiskScoring(t *testing.T) {
	// Test repository risk scoring
	auditData := AuditData{
		Organization: "test-org",
		Summary: AuditSummary{
			TotalRepositories:     3,
			CompliantRepositories: 1,
			TotalViolations:       5,
			CriticalViolations:    2,
		},
		Repositories: []RepositoryAudit{
			{
				Name:             "high-risk-repo",
				Visibility:       "public",
				OverallCompliant: false,
				ViolationCount:   3,
				CriticalCount:    2,
			},
			{
				Name:             "medium-risk-repo",
				Visibility:       "private",
				OverallCompliant: false,
				ViolationCount:   2,
				CriticalCount:    0,
			},
			{
				Name:             "low-risk-repo",
				Visibility:       "private",
				OverallCompliant: true,
				ViolationCount:   0,
				CriticalCount:    0,
			},
		},
		Violations: []ViolationDetail{
			{
				Repository: "high-risk-repo",
				Policy:     "branch_protection",
				Severity:   "critical",
			},
			{
				Repository: "high-risk-repo",
				Policy:     "vulnerability_management",
				Severity:   "critical",
			},
			{
				Repository: "high-risk-repo",
				Policy:     "documentation",
				Severity:   "low",
			},
		},
	}

	riskScores := calculateRepositoryRiskScores(auditData)

	// Verify risk scores are calculated
	assert.Len(t, riskScores, 3)

	// Find high-risk repo
	var highRiskScore *RepositoryRiskScore
	for i, score := range riskScores {
		if score.Repository == "high-risk-repo" {
			highRiskScore = &riskScores[i]
			break
		}
	}

	require.NotNil(t, highRiskScore)
	assert.Equal(t, "critical", highRiskScore.RiskLevel)
	assert.Greater(t, highRiskScore.TotalScore, 50.0)
	assert.NotEmpty(t, highRiskScore.RiskFactors)
	assert.NotEmpty(t, highRiskScore.Recommendations)

	// Verify recommendations include critical security fix
	hasUrgentRec := false
	for _, rec := range highRiskScore.Recommendations {
		if strings.Contains(rec, "URGENT") {
			hasUrgentRec = true
			break
		}
	}
	assert.True(t, hasUrgentRec, "High-risk repo should have urgent recommendations")
}

func TestPolicyOverrides(t *testing.T) {
	// Test that policy overrides work correctly
	policy := &config.PolicyTemplate{
		Rules: map[string]config.PolicyRule{
			"require_reviews": {
				Type:        "min_reviews",
				Value:       1,
				Enforcement: "recommended",
			},
			"enforce_admins": {
				Type:        "enforce_admins",
				Value:       false,
				Enforcement: "optional",
			},
		},
	}

	override := config.PolicyOverride{
		Enforcement: "required",
		Rules: map[string]config.RuleOverride{
			"require_reviews": {
				Value: 2,
			},
			"enforce_admins": {
				Disabled: true,
			},
		},
	}

	// Apply override
	applyPolicyOverride(policy, override)

	// Verify changes
	require.Contains(t, policy.Rules, "require_reviews")
	assert.Equal(t, 2, policy.Rules["require_reviews"].Value)
	assert.Equal(t, "required", policy.Rules["require_reviews"].Enforcement)

	// Verify disabled rule is removed
	assert.NotContains(t, policy.Rules, "enforce_admins")
}

func TestAuditOptions(t *testing.T) {
	// Test hasNewFeatures function
	tests := []struct {
		name     string
		opts     AuditOptions
		expected bool
	}{
		{
			name:     "no new features",
			opts:     AuditOptions{},
			expected: false,
		},
		{
			name: "has filter",
			opts: AuditOptions{
				FilterVisibility: "private",
			},
			expected: true,
		},
		{
			name: "has policy preset",
			opts: AuditOptions{
				PolicyPreset: "soc2",
			},
			expected: true,
		},
		{
			name: "has SARIF format",
			opts: AuditOptions{
				Format: "sarif",
			},
			expected: true,
		},
		{
			name: "has CI/CD features",
			opts: AuditOptions{
				ExitOnFail: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasNewFeatures(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkflowAuditor(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}

	auditor := NewWorkflowAuditor(logger, apiClient)

	assert.NotNil(t, auditor)
	assert.Equal(t, logger, auditor.logger)
	assert.Equal(t, apiClient, auditor.apiClient)
}

func TestWorkflowAuditor_AuditRepository(t *testing.T) {
	auditor := createTestAuditor()
	ctx := context.Background()

	result, err := auditor.AuditRepository(ctx, "testorg", "testrepo")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "testrepo", result.Repository)
	assert.Equal(t, "testorg", result.Organization)
	assert.Greater(t, result.TotalWorkflows, 0)
	assert.NotEmpty(t, result.AuditedFiles)
	assert.NotNil(t, result.Summary)
	assert.NotZero(t, result.Timestamp)
}

func TestWorkflowAuditor_ExtractTriggers(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "string trigger",
			input:    "push",
			expected: []string{"push"},
		},
		{
			name:     "array triggers",
			input:    []interface{}{"push", "pull_request"},
			expected: []string{"push", "pull_request"},
		},
		{
			name:     "map triggers",
			input:    map[string]interface{}{"push": nil, "release": map[string]interface{}{"types": []string{"published"}}},
			expected: []string{"push", "release"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.extractTriggers(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_NormalizePermissions(t *testing.T) {
	auditor := createTestAuditor()

	input := map[string]interface{}{
		"contents": "read",
		"packages": "write",
		"metadata": "read",
		"invalid":  123, // Should be ignored
	}

	expected := map[string]string{
		"contents": "read",
		"packages": "write",
		"metadata": "read",
	}

	result := auditor.normalizePermissions(input)
	assert.Equal(t, expected, result)
}

func TestWorkflowAuditor_IsActionPinned(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		uses     string
		expected bool
	}{
		{
			name:     "pinned to commit hash",
			uses:     "actions/checkout@8e5e7e5ab8b370d6c329ec480221332ada57f0ab",
			expected: true,
		},
		{
			name:     "pinned to tag",
			uses:     "actions/checkout@v4",
			expected: false,
		},
		{
			name:     "not pinned",
			uses:     "actions/checkout",
			expected: false,
		},
		{
			name:     "short hash",
			uses:     "actions/checkout@abc123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.isActionPinned(tt.uses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_IsActionVerified(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		uses     string
		expected bool
	}{
		{
			name:     "GitHub verified action",
			uses:     "actions/checkout@v4",
			expected: true,
		},
		{
			name:     "Docker verified action",
			uses:     "docker/build-push-action@v4",
			expected: true,
		},
		{
			name:     "Third party action",
			uses:     "third-party/some-action@v1",
			expected: false,
		},
		{
			name:     "User action",
			uses:     "user123/custom-action@main",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.isActionVerified(tt.uses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_HasCodeInjectionRisk(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name:     "safe script",
			script:   "echo 'Hello World'",
			expected: false,
		},
		{
			name:     "github event injection",
			script:   "echo ${{ github.event.head_commit.message }}",
			expected: true,
		},
		{
			name:     "eval usage",
			script:   "eval $(some-command)",
			expected: true,
		},
		{
			name:     "exec usage",
			script:   "exec(dangerous_code)",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.hasCodeInjectionRisk(tt.script)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_HasSecretExposureRisk(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name:     "safe secret usage",
			script:   "curl -H 'Authorization: Bearer ${{ secrets.TOKEN }}' https://api.example.com",
			expected: false,
		},
		{
			name:     "echo secret exposure",
			script:   "echo ${{ secrets.API_KEY }}",
			expected: true,
		},
		{
			name:     "printf secret exposure",
			script:   "printf 'Token: %s' ${{ secrets.TOKEN }}",
			expected: true,
		},
		{
			name:     "cat secret exposure",
			script:   "cat ${{ secrets.CONFIG_FILE }}",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.hasSecretExposureRisk(tt.script)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_ExtractSecretsUsage(t *testing.T) {
	auditor := createTestAuditor()

	step := Step{
		Run: "echo ${{ secrets.API_KEY }} && curl -H 'Authorization: Bearer ${{ secrets.TOKEN }}' https://api.example.com",
		With: map[string]string{
			"token": "${{ secrets.GITHUB_TOKEN }}",
			"key":   "${{ secrets.SECRET_KEY }}",
		},
	}

	expected := []string{"API_KEY", "TOKEN", "GITHUB_TOKEN", "SECRET_KEY"}
	result := auditor.extractSecretsUsage(step)

	assert.ElementsMatch(t, expected, result)
}

func TestWorkflowAuditor_ExtractVariablesUsage(t *testing.T) {
	auditor := createTestAuditor()

	step := Step{
		Run: "echo ${{ vars.NODE_VERSION }} && echo ${{ vars.BUILD_ENV }}",
	}

	expected := []string{"NODE_VERSION", "BUILD_ENV"}
	result := auditor.extractVariablesUsage(step)

	assert.ElementsMatch(t, expected, result)
}

func TestWorkflowAuditor_CalculateStepRisk(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name     string
		risks    []string
		expected SecurityRiskLevel
	}{
		{
			name:     "no risks",
			risks:    []string{},
			expected: SecurityRiskNone,
		},
		{
			name:     "critical risk",
			risks:    []string{"Script may be vulnerable to code injection"},
			expected: SecurityRiskCritical,
		},
		{
			name:     "high risk",
			risks:    []string{"Action is not pinned to a specific commit hash"},
			expected: SecurityRiskHigh,
		},
		{
			name:     "medium risk",
			risks:    []string{"Action is deprecated", "Action not verified", "Some other issue"},
			expected: SecurityRiskMedium,
		},
		{
			name:     "low risk",
			risks:    []string{"Minor issue"},
			expected: SecurityRiskLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditor.calculateStepRisk(tt.risks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowAuditor_AuditWorkflowPermissions(t *testing.T) {
	auditor := createTestAuditor()

	tests := []struct {
		name         string
		permissions  map[string]string
		expectIssues bool
	}{
		{
			name: "safe permissions",
			permissions: map[string]string{
				"contents": "read",
				"metadata": "read",
			},
			expectIssues: false,
		},
		{
			name: "excessive permissions",
			permissions: map[string]string{
				"contents": "write",
				"packages": "write",
				"metadata": "read",
			},
			expectIssues: true,
		},
		{
			name: "mixed permissions",
			permissions: map[string]string{
				"contents": "write",
				"metadata": "read",
				"actions":  "read",
			},
			expectIssues: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := auditor.auditWorkflowPermissions(tt.permissions, "test.yml")

			if tt.expectIssues {
				assert.NotEmpty(t, issues)
				// Check that issues have proper structure
				for _, issue := range issues {
					assert.NotEmpty(t, issue.ID)
					assert.Equal(t, IssueTypeExcessivePermissions, issue.Type)
					assert.Equal(t, SeverityHigh, issue.Severity)
					assert.NotEmpty(t, issue.Title)
					assert.NotEmpty(t, issue.Description)
					assert.NotEmpty(t, issue.Suggestion)
				}
			} else {
				assert.Empty(t, issues)
			}
		})
	}
}

func TestWorkflowAuditor_AnalyzePermissionUsage(t *testing.T) {
	auditor := createTestAuditor()

	files := []WorkflowFileAudit{
		{
			FilePath: "ci.yml",
			Permissions: map[string]string{
				"contents": "read",
				"packages": "write",
			},
			Jobs: []JobAuditInfo{
				{
					JobID: "test",
					Permissions: map[string]string{
						"contents": "read",
					},
				},
			},
		},
		{
			FilePath: "release.yml",
			Permissions: map[string]string{
				"contents": "write",
				"packages": "write",
			},
		},
	}

	result := auditor.analyzePermissionUsage(files)

	assert.NotEmpty(t, result)

	// Check that we have entries for different permission combinations
	foundContentsRead := false
	foundPackagesWrite := false

	for _, usage := range result {
		if usage.Scope == "contents" && usage.Permission == "read" {
			foundContentsRead = true
			assert.GreaterOrEqual(t, usage.UsageCount, 1)
		}
		if usage.Scope == "packages" && usage.Permission == "write" {
			foundPackagesWrite = true
			assert.GreaterOrEqual(t, usage.UsageCount, 1)
		}
	}

	assert.True(t, foundContentsRead)
	assert.True(t, foundPackagesWrite)
}

func TestWorkflowAuditor_AnalyzeActionUsage(t *testing.T) {
	auditor := createTestAuditor()

	files := []WorkflowFileAudit{
		{
			FilePath: "ci.yml",
			Jobs: []JobAuditInfo{
				{
					Steps: []StepAuditInfo{
						{
							Uses:          "actions/checkout@v4",
							ActionVersion: "v4",
							SecurityRisk:  SecurityRiskLow,
						},
						{
							Uses:          "actions/setup-node@v4",
							ActionVersion: "v4",
							SecurityRisk:  SecurityRiskLow,
						},
					},
				},
			},
		},
		{
			FilePath: "release.yml",
			Jobs: []JobAuditInfo{
				{
					Steps: []StepAuditInfo{
						{
							Uses:          "actions/checkout@v4",
							ActionVersion: "v4",
							SecurityRisk:  SecurityRiskLow,
						},
					},
				},
			},
		},
	}

	result := auditor.analyzeActionUsage(files)

	assert.NotEmpty(t, result)

	// Check for checkout action (should appear in both files)
	foundCheckout := false
	foundSetupNode := false

	for _, usage := range result {
		if usage.ActionName == "actions/checkout" {
			foundCheckout = true
			assert.Equal(t, 2, usage.UsageCount)
			assert.Len(t, usage.WorkflowFiles, 2)
		}
		if usage.ActionName == "actions/setup-node" {
			foundSetupNode = true
			assert.Equal(t, 1, usage.UsageCount)
			assert.Len(t, usage.WorkflowFiles, 1)
		}
	}

	assert.True(t, foundCheckout)
	assert.True(t, foundSetupNode)
}

func TestWorkflowAuditor_GenerateSummary(t *testing.T) {
	auditor := createTestAuditor()

	result := &WorkflowAuditResult{
		AuditedFiles: []WorkflowFileAudit{
			{
				SecurityScore: 85,
				Issues: []WorkflowSecurityIssue{
					{Severity: SeverityHigh},
					{Severity: SeverityMedium},
				},
			},
			{
				SecurityScore: 90,
				Issues:        []WorkflowSecurityIssue{},
			},
			{
				SecurityScore: 70,
				Issues: []WorkflowSecurityIssue{
					{Severity: SeverityCritical},
					{Severity: SeverityLow},
				},
			},
		},
		SecurityIssues: []WorkflowSecurityIssue{
			{Severity: SeverityCritical},
			{Severity: SeverityHigh},
			{Severity: SeverityMedium},
			{Severity: SeverityLow},
		},
	}

	summary := auditor.generateSummary(result)

	assert.Equal(t, 3, summary.TotalFiles)
	assert.Equal(t, 2, summary.FilesWithIssues) // 2 files have issues
	assert.Equal(t, 1, summary.CriticalIssues)
	assert.Equal(t, 1, summary.HighRiskIssues)
	assert.Equal(t, 1, summary.MediumRiskIssues)
	assert.Equal(t, 1, summary.LowRiskIssues)
	assert.Equal(t, float64(245)/3, summary.AverageSecurityScore) // (85+90+70)/3
	assert.GreaterOrEqual(t, summary.ComplianceScore, 0.0)
	assert.LessOrEqual(t, summary.ComplianceScore, 100.0)
}

func TestWorkflowAuditor_SecurityConstants(t *testing.T) {
	// Test enum constants are properly defined
	assert.Equal(t, WorkflowIssueType("excessive_permissions"), IssueTypeExcessivePermissions)
	assert.Equal(t, WorkflowIssueType("unpinned_action"), IssueTypeUnpinnedAction)
	assert.Equal(t, WorkflowIssueType("code_injection"), IssueTypeCodeInjection)

	assert.Equal(t, SecurityIssueSeverity("critical"), SeverityCritical)
	assert.Equal(t, SecurityIssueSeverity("high"), SeverityHigh)
	assert.Equal(t, SecurityIssueSeverity("medium"), SeverityMedium)

	assert.Equal(t, SecurityRiskLevel("critical"), SecurityRiskCritical)
	assert.Equal(t, SecurityRiskLevel("high"), SecurityRiskHigh)
	assert.Equal(t, SecurityRiskLevel("none"), SecurityRiskNone)
}

// Integration test with mock workflow content
func TestWorkflowAuditor_AuditWorkflowFile_Integration(t *testing.T) {
	auditor := createTestAuditor()
	ctx := context.Background()

	// Test CI workflow audit
	audit, err := auditor.auditWorkflowFile(ctx, "testorg", "testrepo", ".github/workflows/ci.yml")
	require.NoError(t, err)
	require.NotNil(t, audit)

	assert.Equal(t, ".github/workflows/ci.yml", audit.FilePath)
	assert.Equal(t, "CI", audit.WorkflowName)
	assert.Contains(t, audit.Triggers, "push")
	assert.Contains(t, audit.Triggers, "pull_request")
	assert.NotEmpty(t, audit.Jobs)
	assert.Greater(t, audit.SecurityScore, 0)

	// Check that we have analyzed permissions
	assert.Contains(t, audit.Permissions, "contents")
	assert.Equal(t, "read", audit.Permissions["contents"])

	// Check jobs
	require.Len(t, audit.Jobs, 1)
	testJob := audit.Jobs[0]
	assert.Equal(t, "test", testJob.JobID)
	assert.Equal(t, "ubuntu-latest", testJob.RunsOn)
	assert.NotEmpty(t, testJob.Steps)

	// Check steps analysis
	require.GreaterOrEqual(t, len(testJob.Steps), 3)

	checkoutStep := testJob.Steps[0]
	assert.Equal(t, "actions/checkout@v4", checkoutStep.Uses)
	assert.Equal(t, "v4", checkoutStep.ActionVersion)

	setupNodeStep := testJob.Steps[1]
	assert.Equal(t, "actions/setup-node@v4", setupNodeStep.Uses)
	assert.Equal(t, "v4", setupNodeStep.ActionVersion)
}

// Benchmark tests
func BenchmarkAuditRepository(b *testing.B) {
	auditor := createTestAuditor()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auditor.AuditRepository(ctx, "testorg", "testrepo")
	}
}

func BenchmarkAnalyzeActionSecurity(b *testing.B) {
	auditor := createTestAuditor()

	testActions := []string{
		"actions/checkout@v4",
		"actions/setup-node@v4",
		"docker/build-push-action@v4",
		"third-party/action@main",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, action := range testActions {
			auditor.analyzeActionSecurity(action)
		}
	}
}

func BenchmarkExtractSecretsUsage(b *testing.B) {
	auditor := createTestAuditor()

	step := Step{
		Run: `
			echo "Starting deployment"
			curl -H "Authorization: Bearer ${{ secrets.API_TOKEN }}" \
				-d '{"key": "${{ secrets.SECRET_KEY }}"}' \
				https://api.example.com/deploy
			echo "Deployment completed"
		`,
		With: map[string]string{
			"token":    "${{ secrets.GITHUB_TOKEN }}",
			"password": "${{ secrets.DB_PASSWORD }}",
			"api_key":  "${{ secrets.API_KEY }}",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auditor.extractSecretsUsage(step)
	}
}

// Helper function to create a test auditor
func createTestAuditor() *WorkflowAuditor {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	return NewWorkflowAuditor(logger, apiClient)
}

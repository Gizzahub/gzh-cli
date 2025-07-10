package repoconfig

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCVSSVector(t *testing.T) {
	tests := []struct {
		name      string
		violation ViolationDetail
		expected  CVSSVector
	}{
		{
			name: "Branch Protection Violation",
			violation: ViolationDetail{
				Policy:   "Branch Protection",
				Severity: "critical",
			},
			expected: CVSSVector{
				AttackVector:       "Network",
				AttackComplexity:   "Low",
				PrivilegesRequired: "Low",
				Integrity:          "High",
				Availability:       "High",
			},
		},
		{
			name: "Security Scanning Violation",
			violation: ViolationDetail{
				Policy:   "Security Scanning",
				Severity: "high",
			},
			expected: CVSSVector{
				AttackVector:       "Network",
				AttackComplexity:   "Low",
				PrivilegesRequired: "None",
				Confidentiality:    "High",
				Integrity:          "High",
			},
		},
		{
			name: "Visibility Violation",
			violation: ViolationDetail{
				Policy:   "Repository Visibility",
				Severity: "medium",
			},
			expected: CVSSVector{
				AttackVector:       "Network",
				AttackComplexity:   "Low",
				PrivilegesRequired: "None",
				Confidentiality:    "Low",
				Integrity:          "Low",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCVSSVector(tt.violation)

			assert.Equal(t, tt.expected.AttackVector, result.AttackVector)
			assert.Equal(t, tt.expected.AttackComplexity, result.AttackComplexity)
			assert.Equal(t, tt.expected.PrivilegesRequired, result.PrivilegesRequired)
			assert.True(t, result.BaseScore >= 0.0 && result.BaseScore <= 10.0)
			assert.True(t, result.TemporalScore >= 0.0 && result.TemporalScore <= 10.0)
			assert.True(t, result.EnvironmentalScore >= 0.0 && result.EnvironmentalScore <= 10.0)
		})
	}
}

func TestCalculateCVSSBaseScore(t *testing.T) {
	tests := []struct {
		name     string
		vector   CVSSVector
		expected float64
	}{
		{
			name: "High Impact Vector",
			vector: CVSSVector{
				AttackVector:       "Network",
				AttackComplexity:   "Low",
				PrivilegesRequired: "None",
				UserInteraction:    "None",
				Scope:              "Unchanged",
				Confidentiality:    "High",
				Integrity:          "High",
				Availability:       "High",
			},
			expected: 9.8,
		},
		{
			name: "Medium Impact Vector",
			vector: CVSSVector{
				AttackVector:       "Network",
				AttackComplexity:   "Low",
				PrivilegesRequired: "Low",
				UserInteraction:    "None",
				Scope:              "Unchanged",
				Confidentiality:    "Low",
				Integrity:          "Low",
				Availability:       "Low",
			},
			expected: 6.3,
		},
		{
			name: "Low Impact Vector",
			vector: CVSSVector{
				AttackVector:       "Local",
				AttackComplexity:   "High",
				PrivilegesRequired: "High",
				UserInteraction:    "Required",
				Scope:              "Unchanged",
				Confidentiality:    "Low",
				Integrity:          "None",
				Availability:       "None",
			},
			expected: 1.8,
		},
		{
			name: "No Impact Vector",
			vector: CVSSVector{
				AttackVector:       "Physical",
				AttackComplexity:   "High",
				PrivilegesRequired: "High",
				UserInteraction:    "Required",
				Scope:              "Unchanged",
				Confidentiality:    "None",
				Integrity:          "None",
				Availability:       "None",
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCVSSBaseScore(tt.vector)
			assert.InDelta(t, tt.expected, result, 0.5, "CVSS base score should be within acceptable range")
		})
	}
}

func TestGetCVSSMetricValue(t *testing.T) {
	tests := []struct {
		metric     string
		metricType string
		expected   float64
	}{
		{"None", "impact", 0.0},
		{"Low", "impact", 0.22},
		{"High", "impact", 0.56},
		{"Network", "av", 0.85},
		{"Adjacent", "av", 0.62},
		{"Local", "av", 0.55},
		{"Physical", "av", 0.2},
		{"Low", "ac", 0.77},
		{"High", "ac", 0.44},
		{"None", "pr", 0.85},
		{"Low", "pr", 0.62},
		{"High", "pr", 0.27},
		{"None", "ui", 0.85},
		{"Required", "ui", 0.62},
	}

	for _, tt := range tests {
		t.Run(tt.metric+"_"+tt.metricType, func(t *testing.T) {
			result := getCVSSMetricValue(tt.metric, tt.metricType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRiskLevel(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{9.5, "Critical"},
		{9.0, "Critical"},
		{8.5, "High"},
		{7.0, "High"},
		{6.5, "Medium"},
		{4.0, "Medium"},
		{3.5, "Low"},
		{0.1, "Low"},
		{0.0, "None"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateBusinessRisk(t *testing.T) {
	tests := []struct {
		name      string
		violation ViolationDetail
		cvssScore float64
		expected  BusinessRiskFactor
	}{
		{
			name: "API Service Critical Violation",
			violation: ViolationDetail{
				Repository: "api-service",
				Severity:   "critical",
			},
			cvssScore: 9.5,
			expected: BusinessRiskFactor{
				DataSensitivity:     "Confidential",
				BusinessCriticality: "Critical",
				ComplianceImpact:    "High",
				ReputationRisk:      "High",
				CustomerImpact:      "High",
				FinancialImpact:     100000,
			},
		},
		{
			name: "Documentation Low Violation",
			violation: ViolationDetail{
				Repository: "public-docs",
				Severity:   "low",
			},
			cvssScore: 2.0,
			expected: BusinessRiskFactor{
				DataSensitivity:     "Public",
				BusinessCriticality: "Low",
				ComplianceImpact:    "None",
				ReputationRisk:      "Medium",
				CustomerImpact:      "None",
				FinancialImpact:     1000,
			},
		},
		{
			name: "Regular Service Medium Violation",
			violation: ViolationDetail{
				Repository: "regular-service",
				Severity:   "medium",
			},
			cvssScore: 5.5,
			expected: BusinessRiskFactor{
				DataSensitivity:     "Confidential",
				BusinessCriticality: "Medium",
				ComplianceImpact:    "Low",
				ReputationRisk:      "Low",
				CustomerImpact:      "Low",
				FinancialImpact:     10000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBusinessRisk(tt.violation, tt.cvssScore)

			assert.Equal(t, tt.expected.DataSensitivity, result.DataSensitivity)
			assert.Equal(t, tt.expected.BusinessCriticality, result.BusinessCriticality)
			assert.Equal(t, tt.expected.ComplianceImpact, result.ComplianceImpact)
			assert.Equal(t, tt.expected.ReputationRisk, result.ReputationRisk)
			assert.Equal(t, tt.expected.CustomerImpact, result.CustomerImpact)
			assert.Equal(t, tt.expected.FinancialImpact, result.FinancialImpact)
		})
	}
}

func TestAssessImpact(t *testing.T) {
	tests := []struct {
		name      string
		violation ViolationDetail
		cvssScore float64
		expected  ImpactAssessment
	}{
		{
			name: "Critical CVSS Score",
			violation: ViolationDetail{
				Repository: "test-repo",
			},
			cvssScore: 9.5,
			expected: ImpactAssessment{
				SecurityImpact:    "Critical",
				ComplianceImpact:  "High",
				OperationalImpact: "High",
				ExposureLevel:     "High",
				LikelihoodExploit: "High",
			},
		},
		{
			name: "High CVSS Score",
			violation: ViolationDetail{
				Repository: "test-repo",
			},
			cvssScore: 8.0,
			expected: ImpactAssessment{
				SecurityImpact:    "High",
				ComplianceImpact:  "Medium",
				OperationalImpact: "Medium",
				ExposureLevel:     "Medium",
				LikelihoodExploit: "Medium",
			},
		},
		{
			name: "Medium CVSS Score",
			violation: ViolationDetail{
				Repository: "test-repo",
			},
			cvssScore: 5.0,
			expected: ImpactAssessment{
				SecurityImpact:    "Medium",
				ComplianceImpact:  "Low",
				OperationalImpact: "Low",
				ExposureLevel:     "Low",
				LikelihoodExploit: "Low",
			},
		},
		{
			name: "Low CVSS Score",
			violation: ViolationDetail{
				Repository: "test-repo",
			},
			cvssScore: 2.0,
			expected: ImpactAssessment{
				SecurityImpact:    "Low",
				ComplianceImpact:  "None",
				OperationalImpact: "None",
				ExposureLevel:     "Low",
				LikelihoodExploit: "Low",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := assessImpact(tt.violation, tt.cvssScore)

			assert.Equal(t, tt.expected.SecurityImpact, result.SecurityImpact)
			assert.Equal(t, tt.expected.ComplianceImpact, result.ComplianceImpact)
			assert.Equal(t, tt.expected.OperationalImpact, result.OperationalImpact)
			assert.Equal(t, tt.expected.ExposureLevel, result.ExposureLevel)
			assert.Equal(t, tt.expected.LikelihoodExploit, result.LikelihoodExploit)
			assert.Equal(t, tt.violation.Repository, result.AffectedSystems[0])
		})
	}
}

func TestGenerateRemediationGuidance(t *testing.T) {
	tests := []struct {
		name      string
		violation ViolationDetail
		cvssScore float64
		expected  RemediationGuidance
	}{
		{
			name: "Branch Protection Violation",
			violation: ViolationDetail{
				Policy:      "Branch Protection",
				Remediation: "Enable branch protection",
			},
			cvssScore: 8.5,
			expected: RemediationGuidance{
				Recommendation:  "Enable branch protection",
				EstimatedEffort: "Medium",
				RequiredSkills:  []string{"GitHub Administration", "Branch Protection Configuration"},
				RiskReduction:   7.65,          // 8.5 * 0.9
				Timeline:        time.Hour * 3, // 2 hours * 1.5
				Cost:            2000,
			},
		},
		{
			name: "Security Scanning Violation",
			violation: ViolationDetail{
				Policy:      "Security Scanning",
				Remediation: "Enable security scanning",
			},
			cvssScore: 9.0,
			expected: RemediationGuidance{
				Recommendation:  "Enable security scanning",
				EstimatedEffort: "High",
				RequiredSkills:  []string{"GitHub Administration", "Security Configuration"},
				RiskReduction:   8.1,           // 9.0 * 0.9
				Timeline:        time.Hour * 8, // 4 hours * 2
				Cost:            5000,
			},
		},
		{
			name: "Access Control Violation",
			violation: ViolationDetail{
				Policy:      "Access Control",
				Remediation: "Review access permissions",
			},
			cvssScore: 6.0,
			expected: RemediationGuidance{
				Recommendation:  "Review access permissions",
				EstimatedEffort: "Medium",
				RequiredSkills:  []string{"GitHub Administration", "Access Management"},
				RiskReduction:   5.4,            // 6.0 * 0.9
				Timeline:        time.Hour * 12, // 8 hours * 1.5
				Cost:            2000,
			},
		},
		{
			name: "Low Severity Violation",
			violation: ViolationDetail{
				Policy:      "Documentation",
				Remediation: "Add documentation",
			},
			cvssScore: 3.0,
			expected: RemediationGuidance{
				Recommendation:  "Add documentation",
				EstimatedEffort: "Low",
				RequiredSkills:  []string{"GitHub Administration"},
				RiskReduction:   2.7, // 3.0 * 0.9
				Timeline:        time.Hour * 24,
				Cost:            500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRemediationGuidance(tt.violation, tt.cvssScore)

			assert.Equal(t, tt.expected.Recommendation, result.Recommendation)
			assert.Equal(t, tt.expected.EstimatedEffort, result.EstimatedEffort)
			assert.Equal(t, tt.expected.RequiredSkills, result.RequiredSkills)
			assert.InDelta(t, tt.expected.RiskReduction, result.RiskReduction, 0.1)
			assert.Equal(t, tt.expected.Timeline, result.Timeline)
			assert.Equal(t, tt.expected.Cost, result.Cost)
			assert.True(t, len(result.Steps) > 0)
		})
	}
}

func TestGenerateRiskTimeline(t *testing.T) {
	tests := []struct {
		name      string
		violation ViolationDetail
		expected  RiskTimeline
	}{
		{
			name: "Critical Violation",
			violation: ViolationDetail{
				Severity: "critical",
			},
			expected: RiskTimeline{
				TimeToFix:   "1 hour",
				SLADeadline: "4 hours",
			},
		},
		{
			name: "High Violation",
			violation: ViolationDetail{
				Severity: "high",
			},
			expected: RiskTimeline{
				TimeToFix:   "4 hours",
				SLADeadline: "24 hours",
			},
		},
		{
			name: "Medium Violation",
			violation: ViolationDetail{
				Severity: "medium",
			},
			expected: RiskTimeline{
				TimeToFix:   "1 day",
				SLADeadline: "7 days",
			},
		},
		{
			name: "Low Violation",
			violation: ViolationDetail{
				Severity: "low",
			},
			expected: RiskTimeline{
				TimeToFix:   "1 week",
				SLADeadline: "30 days",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRiskTimeline(tt.violation)

			assert.Equal(t, tt.expected.TimeToFix, result.TimeToFix)
			assert.Equal(t, tt.expected.SLADeadline, result.SLADeadline)
			assert.True(t, !result.FirstDetected.IsZero())
			assert.True(t, !result.LastAssessed.IsZero())
			assert.True(t, len(result.ExposureDuration) > 0)
		})
	}
}

func TestCalculatePriority(t *testing.T) {
	tests := []struct {
		name         string
		cvssScore    float64
		businessRisk BusinessRiskFactor
		expected     int
	}{
		{
			name:      "Critical Business Risk",
			cvssScore: 8.5,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Critical",
				ComplianceImpact:    "High",
			},
			expected: 10, // 8 + 3 + 2 = 13, capped at 10
		},
		{
			name:      "High Business Risk",
			cvssScore: 6.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "High",
				ComplianceImpact:    "Medium",
			},
			expected: 9, // 6 + 2 + 1 = 9
		},
		{
			name:      "Medium Business Risk",
			cvssScore: 4.5,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Medium",
				ComplianceImpact:    "Low",
			},
			expected: 5, // 4 + 1 + 0 = 5
		},
		{
			name:      "Low Business Risk",
			cvssScore: 2.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Low",
				ComplianceImpact:    "None",
			},
			expected: 2, // 2 + 0 + 0 = 2
		},
		{
			name:      "Minimal Risk",
			cvssScore: 0.5,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Low",
				ComplianceImpact:    "None",
			},
			expected: 1, // 0 + 0 + 0 = 0, minimum is 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePriority(tt.cvssScore, tt.businessRisk)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineEscalation(t *testing.T) {
	tests := []struct {
		name         string
		cvssScore    float64
		businessRisk BusinessRiskFactor
		expected     EscalationLevel
	}{
		{
			name:      "Critical CVSS Score",
			cvssScore: 9.5,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "High",
			},
			expected: EscalationLevel{
				Level:            "Executive",
				EscalationReason: "Critical security risk requires immediate executive attention",
				Stakeholders:     []string{"CISO", "CTO", "CEO"},
			},
		},
		{
			name:      "Critical Business Risk",
			cvssScore: 7.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Critical",
			},
			expected: EscalationLevel{
				Level:            "Executive",
				EscalationReason: "Critical security risk requires immediate executive attention",
				Stakeholders:     []string{"CISO", "CTO", "CEO"},
			},
		},
		{
			name:      "High CVSS Score",
			cvssScore: 8.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Medium",
			},
			expected: EscalationLevel{
				Level:            "Management",
				EscalationReason: "High-priority security issue requires management oversight",
				Stakeholders:     []string{"Security Manager", "Engineering Manager"},
			},
		},
		{
			name:      "Medium CVSS with High Compliance Impact",
			cvssScore: 5.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Medium",
				ComplianceImpact:    "High",
			},
			expected: EscalationLevel{
				Level:            "Management",
				EscalationReason: "Compliance-related security issue requires management review",
				Stakeholders:     []string{"Compliance Officer", "Security Team Lead"},
			},
		},
		{
			name:      "Low Risk - No Escalation",
			cvssScore: 3.0,
			businessRisk: BusinessRiskFactor{
				BusinessCriticality: "Low",
				ComplianceImpact:    "Low",
			},
			expected: EscalationLevel{
				Level:            "None",
				EscalationReason: "No escalation required",
				Stakeholders:     []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineEscalation(tt.cvssScore, tt.businessRisk)

			assert.Equal(t, tt.expected.Level, result.Level)
			assert.Equal(t, tt.expected.EscalationReason, result.EscalationReason)
			assert.Equal(t, tt.expected.Stakeholders, result.Stakeholders)
			assert.True(t, !result.RequiredBy.IsZero())
		})
	}
}

func TestFilterByThreshold(t *testing.T) {
	assessments := []RiskAssessment{
		{RiskLevel: "Critical", CVSSScore: 9.5},
		{RiskLevel: "High", CVSSScore: 8.0},
		{RiskLevel: "Medium", CVSSScore: 5.0},
		{RiskLevel: "Low", CVSSScore: 2.0},
	}

	tests := []struct {
		name      string
		threshold string
		expected  int
	}{
		{"All risks", "all", 4},
		{"Critical only", "critical", 1},
		{"High only", "high", 1},
		{"Medium only", "medium", 1},
		{"Low only", "low", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByThreshold(assessments, tt.threshold)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestSortRiskAssessments(t *testing.T) {
	assessments := []RiskAssessment{
		{
			Repository: "repo-c",
			Policy:     "Policy Z",
			CVSSScore:  5.0,
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact: 10000,
			},
		},
		{
			Repository: "repo-a",
			Policy:     "Policy A",
			CVSSScore:  8.0,
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact: 50000,
			},
		},
		{
			Repository: "repo-b",
			Policy:     "Policy B",
			CVSSScore:  7.0,
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact: 30000,
			},
		},
	}

	tests := []struct {
		name     string
		sortBy   string
		expected []string
	}{
		{"Sort by score", "score", []string{"repo-a", "repo-b", "repo-c"}},
		{"Sort by repository", "repository", []string{"repo-a", "repo-b", "repo-c"}},
		{"Sort by policy", "policy", []string{"repo-a", "repo-b", "repo-c"}},
		{"Sort by impact", "impact", []string{"repo-a", "repo-b", "repo-c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			testAssessments := make([]RiskAssessment, len(assessments))
			copy(testAssessments, assessments)

			sortRiskAssessments(testAssessments, tt.sortBy)

			var result []string
			for _, assessment := range testAssessments {
				result = append(result, assessment.Repository)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateRiskAssessments(t *testing.T) {
	violations := []ViolationDetail{
		{
			Repository:  "test-repo-1",
			Policy:      "Branch Protection",
			Setting:     "main.protection",
			Description: "Branch protection disabled",
			Severity:    "critical",
			Remediation: "Enable branch protection",
		},
		{
			Repository:  "test-repo-2",
			Policy:      "Security Scanning",
			Setting:     "vulnerability_alerts",
			Description: "Security scanning disabled",
			Severity:    "high",
			Remediation: "Enable security scanning",
		},
	}

	assessments := calculateRiskAssessments(violations)

	assert.Equal(t, 2, len(assessments))

	// Test first assessment
	assert.Equal(t, "test-repo-1", assessments[0].Repository)
	assert.Equal(t, "Branch Protection", assessments[0].Policy)
	assert.Equal(t, "Branch protection disabled", assessments[0].Violation)
	assert.True(t, assessments[0].CVSSScore > 0)
	assert.True(t, assessments[0].Priority >= 1 && assessments[0].Priority <= 10)
	assert.True(t, len(assessments[0].ID) > 0)

	// Test second assessment
	assert.Equal(t, "test-repo-2", assessments[1].Repository)
	assert.Equal(t, "Security Scanning", assessments[1].Policy)
	assert.Equal(t, "Security scanning disabled", assessments[1].Violation)
	assert.True(t, assessments[1].CVSSScore > 0)
	assert.True(t, assessments[1].Priority >= 1 && assessments[1].Priority <= 10)
	assert.True(t, len(assessments[1].ID) > 0)
}

func TestCalculateBusinessRiskMetrics(t *testing.T) {
	assessments := []RiskAssessment{
		{
			CVSSScore: 9.0,
			RiskLevel: "Critical",
			Policy:    "Branch Protection",
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact:     100000,
				ComplianceImpact:    "High",
				BusinessCriticality: "Critical",
			},
			Escalation: EscalationLevel{
				Level: "Executive",
			},
		},
		{
			CVSSScore: 7.0,
			RiskLevel: "High",
			Policy:    "Security Scanning",
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact:     50000,
				ComplianceImpact:    "Medium",
				BusinessCriticality: "High",
			},
			Escalation: EscalationLevel{
				Level: "Management",
			},
		},
		{
			CVSSScore: 4.0,
			RiskLevel: "Medium",
			Policy:    "Access Control",
			BusinessRisk: BusinessRiskFactor{
				FinancialImpact:     10000,
				ComplianceImpact:    "Low",
				BusinessCriticality: "Medium",
			},
			Escalation: EscalationLevel{
				Level: "None",
			},
		},
	}

	auditData := AuditData{
		Organization: "test-org",
		Summary: AuditSummary{
			TotalRepositories: 10,
			TotalViolations:   3,
		},
	}

	metrics := calculateBusinessRiskMetrics(assessments, auditData)

	assert.Equal(t, 20.0, metrics.TotalRiskScore)                       // 9.0 + 7.0 + 4.0
	assert.Equal(t, 6.67, math.Round(metrics.AverageRiskScore*100)/100) // 20.0 / 3
	assert.Equal(t, 160000.0, metrics.EstimatedCost)                    // 100000 + 50000 + 10000
	assert.Equal(t, 1, metrics.CriticalRiskCount)
	assert.True(t, metrics.EscalationRequired)
	assert.Equal(t, 3, len(metrics.RiskDistribution))
	assert.Equal(t, 1, metrics.RiskDistribution["Critical"])
	assert.Equal(t, 1, metrics.RiskDistribution["High"])
	assert.Equal(t, 1, metrics.RiskDistribution["Medium"])
	assert.True(t, metrics.SecurityRiskScore > 0)
	assert.True(t, metrics.ComplianceRiskScore > 0)
	assert.True(t, metrics.BusinessImpactScore > 0)
}

func TestMapPolicyToStandard(t *testing.T) {
	tests := []struct {
		policy   string
		expected string
	}{
		{"Security Scanning", "ISO 27001"},
		{"Branch Protection", "SOC 2"},
		{"Access Control", "GDPR"},
		{"Audit Logging", "SOX"},
		{"Unknown Policy", "General"},
	}

	for _, tt := range tests {
		t.Run(tt.policy, func(t *testing.T) {
			result := mapPolicyToStandard(tt.policy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Minute, "30 minutes"},
		{2 * time.Hour, "2 hours"},
		{25 * time.Hour, "1 days"},
		{48 * time.Hour, "2 days"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRiskAssessmentTableOutput(t *testing.T) {
	assessments := []RiskAssessment{
		{
			Repository: "test-repo",
			Policy:     "Branch Protection",
			Violation:  "Branch protection disabled",
			CVSSScore:  8.5,
			RiskLevel:  "High",
			Priority:   8,
			BusinessRisk: BusinessRiskFactor{
				BusinessCriticality: "High",
				FinancialImpact:     50000,
			},
			Remediation: RemediationGuidance{
				Recommendation:  "Enable branch protection",
				EstimatedEffort: "Medium",
			},
			Escalation: EscalationLevel{
				Level: "Management",
			},
		},
	}

	metrics := &BusinessRiskMetrics{
		TotalRiskScore:     8.5,
		AverageRiskScore:   8.5,
		CriticalRiskCount:  0,
		EstimatedCost:      50000,
		RiskTrend:          "stable",
		EscalationRequired: true,
		RiskDistribution: map[string]int{
			"High": 1,
		},
	}

	// Test table output (should not error)
	err := outputRiskAssessmentTable(assessments, metrics, true)
	assert.NoError(t, err)

	// Test without metrics
	err = outputRiskAssessmentTable(assessments, nil, false)
	assert.NoError(t, err)

	// Test with empty assessments
	err = outputRiskAssessmentTable([]RiskAssessment{}, metrics, false)
	assert.NoError(t, err)
}

func TestOutputRiskAssessmentJSON(t *testing.T) {
	assessments := []RiskAssessment{
		{
			Repository: "test-repo",
			Policy:     "Branch Protection",
			CVSSScore:  8.5,
			RiskLevel:  "High",
		},
	}

	metrics := &BusinessRiskMetrics{
		TotalRiskScore:   8.5,
		AverageRiskScore: 8.5,
	}

	// Test JSON output (should not error)
	err := outputRiskAssessmentJSON(assessments, metrics, "")
	assert.NoError(t, err)
}

func TestOutputRiskAssessmentCSV(t *testing.T) {
	assessments := []RiskAssessment{
		{
			Repository: "test-repo",
			Policy:     "Branch Protection",
			Setting:    "main.protection",
			Violation:  "Branch protection disabled",
			CVSSScore:  8.5,
			RiskLevel:  "High",
			Priority:   8,
			BusinessRisk: BusinessRiskFactor{
				BusinessCriticality: "High",
				FinancialImpact:     50000,
			},
			Remediation: RemediationGuidance{
				Recommendation:  "Enable branch protection",
				EstimatedEffort: "Medium",
			},
			Escalation: EscalationLevel{
				Level: "Management",
			},
		},
	}

	// Test CSV output (should not error)
	err := outputRiskAssessmentCSV(assessments, "")
	assert.NoError(t, err)
}

// Integration test
func TestRiskAssessmentIntegration(t *testing.T) {
	// Create sample violations
	violations := []ViolationDetail{
		{
			Repository:  "critical-service",
			Policy:      "Branch Protection",
			Setting:     "main.protection",
			Description: "Branch protection disabled",
			Severity:    "critical",
			Remediation: "Enable branch protection",
		},
		{
			Repository:  "api-service",
			Policy:      "Security Scanning",
			Setting:     "vulnerability_alerts",
			Description: "Security scanning disabled",
			Severity:    "high",
			Remediation: "Enable security scanning",
		},
	}

	// Calculate risk assessments
	assessments := calculateRiskAssessments(violations)
	assert.Equal(t, 2, len(assessments))

	// Test filtering
	criticalAssessments := filterByThreshold(assessments, "critical")
	assert.True(t, len(criticalAssessments) <= len(assessments))

	// Test sorting
	sortRiskAssessments(assessments, "score")
	assert.True(t, assessments[0].CVSSScore >= assessments[1].CVSSScore)

	// Test business metrics calculation
	auditData := AuditData{
		Organization: "test-org",
		Summary: AuditSummary{
			TotalRepositories: 10,
			TotalViolations:   2,
		},
	}
	metrics := calculateBusinessRiskMetrics(assessments, auditData)
	assert.True(t, metrics.TotalRiskScore > 0)
	assert.True(t, metrics.AverageRiskScore > 0)
	assert.True(t, metrics.EstimatedCost > 0)
}

func TestNewRiskAssessmentCmd(t *testing.T) {
	cmd := newRiskAssessmentCmd()

	assert.Equal(t, "risk-assessment", cmd.Use)
	assert.Equal(t, "Perform CVSS-based risk assessment of policy violations", cmd.Short)
	assert.True(t, strings.Contains(cmd.Long, "CVSS"))
	assert.True(t, strings.Contains(cmd.Long, "risk assessment"))
	assert.True(t, strings.Contains(cmd.Long, "Critical: 9.0-10.0"))
	assert.True(t, cmd.HasFlags())

	// Test flags
	assert.True(t, cmd.Flags().HasFlag("format"))
	assert.True(t, cmd.Flags().HasFlag("output"))
	assert.True(t, cmd.Flags().HasFlag("threshold"))
	assert.True(t, cmd.Flags().HasFlag("show-details"))
	assert.True(t, cmd.Flags().HasFlag("sort-by"))
	assert.True(t, cmd.Flags().HasFlag("include-metrics"))
}

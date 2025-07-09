package compliance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreCalculator(t *testing.T) {
	calculator := NewScoreCalculator()

	t.Run("NewScoreCalculator", func(t *testing.T) {
		assert.NotNil(t, calculator)
		assert.NotEmpty(t, calculator.policyWeights)
		assert.NotEmpty(t, calculator.gradeThresholds)

		// Check default policy weights
		branchProtection, exists := calculator.policyWeights["Branch Protection"]
		assert.True(t, exists)
		assert.Equal(t, 0.25, branchProtection.Weight)
		assert.Equal(t, "security", branchProtection.Category)
	})

	t.Run("CalculateScore_HighCompliance", func(t *testing.T) {
		// High compliance scenario
		auditSummary := AuditSummary{
			TotalRepositories:     10,
			CompliantRepositories: 9,
			CompliancePercentage:  90.0,
			TotalViolations:       5,
			CriticalViolations:    0,
			PolicyCount:           3,
			CompliantCount:        9,
			NonCompliantCount:     1,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Description:          "Require branch protection on main branches",
				Severity:             "critical",
				CompliantRepos:       9,
				ViolatingRepos:       1,
				CompliancePercentage: 90.0,
			},
			{
				PolicyName:           "Required Reviews",
				Description:          "Minimum 2 reviews required for PRs",
				Severity:             "high",
				CompliantRepos:       8,
				ViolatingRepos:       2,
				CompliancePercentage: 80.0,
			},
			{
				PolicyName:           "Security Scanning",
				Description:          "Enable security scanning features",
				Severity:             "medium",
				CompliantRepos:       10,
				ViolatingRepos:       0,
				CompliancePercentage: 100.0,
			},
		}

		score, err := calculator.CalculateScore(auditSummary, policyCompliance, nil)
		require.NoError(t, err)
		require.NotNil(t, score)

		// Verify basic score properties
		assert.True(t, score.TotalScore >= 70.0, "High compliance should result in good score")
		assert.True(t, score.Grade == GradeA || score.Grade == GradeB, "High compliance should get Grade A or B")
		assert.NotEmpty(t, score.PolicyScores)
		assert.Len(t, score.PolicyScores, 3)

		// Verify policy scores
		branchScore, exists := score.PolicyScores["Branch Protection"]
		assert.True(t, exists)
		assert.True(t, branchScore.Score > 80.0)
		assert.Equal(t, 0.25, branchScore.Weight)

		// Verify metrics
		assert.True(t, score.ScoreMetrics.SecurityIndex >= 80.0)
		assert.True(t, score.ScoreMetrics.MaturityIndex >= 80.0)
	})

	t.Run("CalculateScore_LowCompliance", func(t *testing.T) {
		// Low compliance scenario
		auditSummary := AuditSummary{
			TotalRepositories:     10,
			CompliantRepositories: 4,
			CompliancePercentage:  40.0,
			TotalViolations:       15,
			CriticalViolations:    8,
			PolicyCount:           3,
			CompliantCount:        4,
			NonCompliantCount:     6,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Description:          "Require branch protection on main branches",
				Severity:             "critical",
				CompliantRepos:       3,
				ViolatingRepos:       7,
				CompliancePercentage: 30.0,
			},
			{
				PolicyName:           "Required Reviews",
				Description:          "Minimum 2 reviews required for PRs",
				Severity:             "critical",
				CompliantRepos:       2,
				ViolatingRepos:       8,
				CompliancePercentage: 20.0,
			},
			{
				PolicyName:           "Security Scanning",
				Description:          "Enable security scanning features",
				Severity:             "high",
				CompliantRepos:       7,
				ViolatingRepos:       3,
				CompliancePercentage: 70.0,
			},
		}

		score, err := calculator.CalculateScore(auditSummary, policyCompliance, nil)
		require.NoError(t, err)
		require.NotNil(t, score)

		// Verify low compliance results
		assert.True(t, score.TotalScore < 60.0, "Low compliance should result in low score")
		assert.Equal(t, GradeF, score.Grade, "Low compliance should get Grade F")
		assert.NotEmpty(t, score.Recommendations)

		// Should have critical recommendations
		hasUrgentRecommendation := false
		for _, rec := range score.Recommendations {
			t.Logf("Recommendation: %s", rec)
			if strings.Contains(rec, "긴급") {
				hasUrgentRecommendation = true
				break
			}
		}
		assert.True(t, hasUrgentRecommendation, "Should have urgent recommendations for critical violations")
	})

	t.Run("CalculatePolicyScore", func(t *testing.T) {
		policy := PolicyCompliance{
			PolicyName:           "Branch Protection",
			Severity:             "critical",
			CompliantRepos:       8,
			ViolatingRepos:       2,
			CompliancePercentage: 80.0,
		}

		policyScore := calculator.calculatePolicyScore(policy)

		assert.Equal(t, "Branch Protection", policyScore.PolicyName)
		assert.Equal(t, 0.25, policyScore.Weight)
		assert.True(t, policyScore.Score <= 80.0, "Score should be penalized for violations")
		assert.True(t, policyScore.Penalty > 0, "Should have penalty for violations")
		assert.Equal(t, 2, policyScore.ViolationCount)
	})

	t.Run("CalculateGrade", func(t *testing.T) {
		testCases := []struct {
			score         float64
			expectedGrade Grade
		}{
			{95.0, GradeA},
			{90.0, GradeA},
			{85.0, GradeB},
			{75.0, GradeC},
			{65.0, GradeD},
			{45.0, GradeF},
		}

		for _, tc := range testCases {
			grade := calculator.calculateGrade(tc.score)
			assert.Equal(t, tc.expectedGrade, grade,
				"Score %.1f should result in grade %s", tc.score, tc.expectedGrade)
		}
	})

	t.Run("SeverityMultiplier", func(t *testing.T) {
		assert.Equal(t, 2.0, getSeverityMultiplier("critical"))
		assert.Equal(t, 1.5, getSeverityMultiplier("high"))
		assert.Equal(t, 1.0, getSeverityMultiplier("medium"))
		assert.Equal(t, 0.5, getSeverityMultiplier("low"))
		assert.Equal(t, 1.0, getSeverityMultiplier("unknown"))
	})

	t.Run("ScoreBreakdown", func(t *testing.T) {
		auditSummary := AuditSummary{
			CompliancePercentage: 75.0,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Severity:             "critical",
				CompliantRepos:       7,
				ViolatingRepos:       3,
				CompliancePercentage: 70.0,
			},
			{
				PolicyName:           "Required Files",
				Severity:             "low",
				CompliantRepos:       10,
				ViolatingRepos:       0,
				CompliancePercentage: 100.0,
			},
		}

		breakdown := calculator.calculateScoreBreakdown(auditSummary, policyCompliance, 75.0)

		assert.Equal(t, 75.0, breakdown.BaseScore)
		assert.True(t, breakdown.SecurityPenalty >= 0)
		assert.True(t, breakdown.BestPracticeBonus >= 0)
		assert.True(t, breakdown.FinalScore >= 0 && breakdown.FinalScore <= 100)
	})

	t.Run("ConsistencyIndex", func(t *testing.T) {
		// High consistency (all policies at similar compliance levels)
		consistentPolicies := []PolicyCompliance{
			{CompliancePercentage: 85.0},
			{CompliancePercentage: 87.0},
			{CompliancePercentage: 83.0},
		}

		consistentIndex := calculator.calculateConsistencyIndex(consistentPolicies)

		// Low consistency (policies at very different compliance levels)
		inconsistentPolicies := []PolicyCompliance{
			{CompliancePercentage: 95.0},
			{CompliancePercentage: 30.0},
			{CompliancePercentage: 70.0},
		}

		inconsistentIndex := calculator.calculateConsistencyIndex(inconsistentPolicies)

		assert.True(t, consistentIndex > inconsistentIndex,
			"Consistent policies should have higher consistency index")
		assert.True(t, consistentIndex >= 80.0, "High consistency should result in high index")
		assert.True(t, inconsistentIndex < 80.0, "Low consistency should result in low index")
	})

	t.Run("CategoryScore", func(t *testing.T) {
		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				CompliancePercentage: 80.0,
			},
			{
				PolicyName:           "Required Reviews",
				CompliancePercentage: 90.0,
			},
			{
				PolicyName:           "Required Files",
				CompliancePercentage: 95.0,
			},
		}

		securityScore := calculator.calculateCategoryScore(policyCompliance, "security")
		bestPracticeScore := calculator.calculateCategoryScore(policyCompliance, "best-practice")

		// Security category has Branch Protection and Required Reviews
		assert.Equal(t, 85.0, securityScore) // (80 + 90) / 2

		// Best practice category has Required Files
		assert.Equal(t, 95.0, bestPracticeScore)

		// Non-existent category should return 0
		otherScore := calculator.calculateCategoryScore(policyCompliance, "non-existent")
		assert.Equal(t, 0.0, otherScore)
	})

	t.Run("GenerateRecommendations", func(t *testing.T) {
		auditSummary := AuditSummary{
			CompliancePercentage: 60.0,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Severity:             "critical",
				CompliantRepos:       5,
				ViolatingRepos:       5,
				CompliancePercentage: 50.0,
			},
		}

		recommendations := calculator.generateRecommendations(auditSummary, policyCompliance, 60.0)

		assert.NotEmpty(t, recommendations)

		// Should have urgent recommendation for critical violations
		hasUrgent := false
		for _, rec := range recommendations {
			t.Logf("Recommendation: %s", rec)
			if strings.Contains(rec, "긴급") {
				hasUrgent = true
				break
			}
		}
		assert.True(t, hasUrgent, "Should have urgent recommendations for critical violations")
	})

	t.Run("PreviousScoreComparison", func(t *testing.T) {
		// Current score
		auditSummary := AuditSummary{
			CompliancePercentage: 85.0,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Severity:             "critical",
				CompliantRepos:       8,
				ViolatingRepos:       2,
				CompliancePercentage: 80.0,
			},
		}

		// Previous score (lower)
		previousScore := &ComplianceScore{
			TotalScore:   70.0,
			Grade:        GradeC,
			CalculatedAt: time.Now().Add(-24 * time.Hour),
		}

		score, err := calculator.CalculateScore(auditSummary, policyCompliance, previousScore)
		require.NoError(t, err)

		assert.NotNil(t, score.PreviousScore)
		assert.Equal(t, 70.0, score.PreviousScore.TotalScore)
		assert.True(t, score.TotalScore > score.PreviousScore.TotalScore,
			"Current score should be higher than previous")
	})
}

func TestGradeString(t *testing.T) {
	assert.Equal(t, "A", string(GradeA))
	assert.Equal(t, "B", string(GradeB))
	assert.Equal(t, "C", string(GradeC))
	assert.Equal(t, "D", string(GradeD))
	assert.Equal(t, "F", string(GradeF))
}

func TestPolicyWeightValidation(t *testing.T) {
	weights := getDefaultPolicyWeights()

	// Check that all weights sum to reasonable total
	var totalWeight float64
	for _, weight := range weights {
		totalWeight += weight.Weight

		// Validate individual weights
		assert.True(t, weight.Weight > 0 && weight.Weight <= 1.0,
			"Policy weight should be between 0 and 1: %s", weight.PolicyName)
		assert.True(t, weight.MaxPenalty > 0 && weight.MaxPenalty <= 50.0,
			"Max penalty should be reasonable: %s", weight.PolicyName)
		assert.Contains(t, []string{"security", "compliance", "best-practice"}, weight.Category,
			"Category should be valid: %s", weight.PolicyName)
	}

	// Total weight should be 1.0 (100%)
	assert.InDelta(t, 1.0, totalWeight, 0.01, "Total policy weights should sum to 1.0")
}

func TestEdgeCases(t *testing.T) {
	calculator := NewScoreCalculator()

	t.Run("EmptyPolicies", func(t *testing.T) {
		auditSummary := AuditSummary{
			CompliancePercentage: 100.0,
		}

		score, err := calculator.CalculateScore(auditSummary, []PolicyCompliance{}, nil)
		require.NoError(t, err)

		assert.Equal(t, 100.0, score.TotalScore)
		assert.Equal(t, GradeA, score.Grade)
		assert.Empty(t, score.PolicyScores)
	})

	t.Run("UnknownPolicy", func(t *testing.T) {
		auditSummary := AuditSummary{
			CompliancePercentage: 80.0,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Unknown Policy",
				Severity:             "medium",
				CompliantRepos:       8,
				ViolatingRepos:       2,
				CompliancePercentage: 80.0,
			},
		}

		score, err := calculator.CalculateScore(auditSummary, policyCompliance, nil)
		require.NoError(t, err)

		// Should handle unknown policies gracefully
		unknownScore, exists := score.PolicyScores["Unknown Policy"]
		assert.True(t, exists)
		assert.Equal(t, 0.01, unknownScore.Weight) // Default weight
	})

	t.Run("PerfectCompliance", func(t *testing.T) {
		auditSummary := AuditSummary{
			CompliancePercentage: 100.0,
		}

		policyCompliance := []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Severity:             "critical",
				CompliantRepos:       10,
				ViolatingRepos:       0,
				CompliancePercentage: 100.0,
			},
			{
				PolicyName:           "Required Files",
				Severity:             "low",
				CompliantRepos:       10,
				ViolatingRepos:       0,
				CompliancePercentage: 100.0,
			},
		}

		score, err := calculator.CalculateScore(auditSummary, policyCompliance, nil)
		require.NoError(t, err)

		assert.Equal(t, GradeA, score.Grade)
		assert.True(t, score.TotalScore >= 95.0, "Perfect compliance should get near-perfect score")
		assert.True(t, score.ScoreBreakdown.BestPracticeBonus > 0, "Should get best practice bonus")
	})
}

package compliance

import (
	"fmt"
)

// DemoScoreCalculation demonstrates the compliance scoring system
func DemoScoreCalculation() {
	fmt.Println("🎯 Compliance Score Calculator Demo")
	fmt.Println("=====================================")

	calculator := NewScoreCalculator()

	// Demo scenario 1: High compliance organization
	fmt.Println("\n📊 Scenario 1: High Compliance Organization")
	highComplianceAudit := AuditSummary{
		TotalRepositories:     50,
		CompliantRepositories: 45,
		CompliancePercentage:  90.0,
		TotalViolations:       8,
		CriticalViolations:    1,
		PolicyCount:           5,
		CompliantCount:        45,
		NonCompliantCount:     5,
	}

	highCompliancePolicies := []PolicyCompliance{
		{
			PolicyName:           "Branch Protection",
			Description:          "Require branch protection on main branches",
			Severity:             "critical",
			CompliantRepos:       47,
			ViolatingRepos:       3,
			CompliancePercentage: 94.0,
		},
		{
			PolicyName:           "Required Reviews",
			Description:          "Minimum 2 reviews required for PRs",
			Severity:             "high",
			CompliantRepos:       45,
			ViolatingRepos:       5,
			CompliancePercentage: 90.0,
		},
		{
			PolicyName:           "Security Scanning",
			Description:          "Enable security scanning features",
			Severity:             "medium",
			CompliantRepos:       48,
			ViolatingRepos:       2,
			CompliancePercentage: 96.0,
		},
		{
			PolicyName:           "Required Files",
			Description:          "Require README, LICENSE files",
			Severity:             "low",
			CompliantRepos:       40,
			ViolatingRepos:       10,
			CompliancePercentage: 80.0,
		},
	}

	score1, err := calculator.CalculateScore(highComplianceAudit, highCompliancePolicies, nil)
	if err != nil {
		fmt.Printf("Error calculating score: %v\n", err)
		return
	}

	printScoreReport(score1)

	// Demo scenario 2: Low compliance organization
	fmt.Println("\n📊 Scenario 2: Low Compliance Organization")
	lowComplianceAudit := AuditSummary{
		TotalRepositories:     30,
		CompliantRepositories: 10,
		CompliancePercentage:  33.3,
		TotalViolations:       25,
		CriticalViolations:    12,
		PolicyCount:           5,
		CompliantCount:        10,
		NonCompliantCount:     20,
	}

	lowCompliancePolicies := []PolicyCompliance{
		{
			PolicyName:           "Branch Protection",
			Description:          "Require branch protection on main branches",
			Severity:             "critical",
			CompliantRepos:       8,
			ViolatingRepos:       22,
			CompliancePercentage: 26.7,
		},
		{
			PolicyName:           "Required Reviews",
			Description:          "Minimum 2 reviews required for PRs",
			Severity:             "critical",
			CompliantRepos:       5,
			ViolatingRepos:       25,
			CompliancePercentage: 16.7,
		},
		{
			PolicyName:           "Security Scanning",
			Description:          "Enable security scanning features",
			Severity:             "high",
			CompliantRepos:       15,
			ViolatingRepos:       15,
			CompliancePercentage: 50.0,
		},
	}

	score2, err := calculator.CalculateScore(lowComplianceAudit, lowCompliancePolicies, score1)
	if err != nil {
		fmt.Printf("Error calculating score: %v\n", err)
		return
	}

	printScoreReport(score2)

	// Compare scores
	fmt.Println("\n📈 Score Comparison")
	fmt.Println("==================")
	if score2.PreviousScore != nil {
		change := score2.TotalScore - score2.PreviousScore.TotalScore
		if change > 0 {
			fmt.Printf("📈 Score improved by %.1f points\n", change)
		} else {
			fmt.Printf("📉 Score decreased by %.1f points\n", -change)
		}
	}
}

// printScoreReport prints a detailed score report
func printScoreReport(score *ComplianceScore) {
	fmt.Printf("🏆 Total Score: %.1f/100 (%s)\n", score.TotalScore, score.Grade)
	fmt.Printf("⚖️  Weighted Score: %.1f\n", score.WeightedScore)

	fmt.Println("\n📋 Score Breakdown:")
	fmt.Printf("  Base Score: %.1f\n", score.ScoreBreakdown.BaseScore)
	fmt.Printf("  Security Penalty: -%.1f\n", score.ScoreBreakdown.SecurityPenalty)
	fmt.Printf("  Compliance Penalty: -%.1f\n", score.ScoreBreakdown.CompliancePenalty)
	fmt.Printf("  Best Practice Bonus: +%.1f\n", score.ScoreBreakdown.BestPracticeBonus)

	fmt.Println("\n📊 Metrics:")
	fmt.Printf("  Security Index: %.1f\n", score.ScoreMetrics.SecurityIndex)
	fmt.Printf("  Compliance Index: %.1f\n", score.ScoreMetrics.ComplianceIndex)
	fmt.Printf("  Maturity Index: %.1f\n", score.ScoreMetrics.MaturityIndex)
	fmt.Printf("  Consistency Index: %.1f\n", score.ScoreMetrics.ConsistencyIndex)

	fmt.Println("\n🎯 Policy Scores:")
	for policyName, policyScore := range score.PolicyScores {
		fmt.Printf("  %s: %.1f (weight: %.2f, penalty: %.1f)\n",
			policyName, policyScore.Score, policyScore.Weight, policyScore.Penalty)
	}

	if len(score.Recommendations) > 0 {
		fmt.Println("\n💡 Recommendations:")
		for _, rec := range score.Recommendations {
			fmt.Printf("  %s\n", rec)
		}
	}
}

package repoconfig

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"
)

// SARIF structures for Static Analysis Results Interchange Format
type SARIFReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool              SARIFTool         `json:"tool"`
	Results           []SARIFResult     `json:"results"`
	ArtifactsLocation SARIFLocation     `json:"artifactsLocation"`
	ColumnKind        string            `json:"columnKind"`
	LogicalLocations  []LogicalLocation `json:"logicalLocations,omitempty"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name            string      `json:"name"`
	SemanticVersion string      `json:"semanticVersion"`
	FullName        string      `json:"fullName"`
	Rules           []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ShortDescription SARIFText         `json:"shortDescription"`
	FullDescription  SARIFText         `json:"fullDescription"`
	Help             SARIFText         `json:"help"`
	Properties       map[string]string `json:"properties"`
}

type SARIFText struct {
	Text string `json:"text"`
}

type SARIFResult struct {
	RuleID     string                 `json:"ruleId"`
	RuleIndex  int                    `json:"ruleIndex"`
	Level      string                 `json:"level"`
	Message    SARIFText              `json:"message"`
	Locations  []SARIFLocation        `json:"locations"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type SARIFLocation struct {
	URI             string           `json:"uri,omitempty"`
	LogicalLocation *LogicalLocation `json:"logicalLocation,omitempty"`
}

type LogicalLocation struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// JUnit XML structures
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       float64          `xml:"time,attr"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Name       string          `xml:"name,attr"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Time       float64         `xml:"time,attr"`
	Timestamp  string          `xml:"timestamp,attr"`
	Properties JUnitProperties `xml:"properties,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
	SystemOut  string          `xml:"system-out,omitempty"`
	SystemErr  string          `xml:"system-err,omitempty"`
}

type JUnitProperties struct {
	Properties []JUnitProperty `xml:"property"`
}

type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	ClassName string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
	Skipped   *JUnitSkipped `xml:"skipped,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
	SystemErr string        `xml:"system-err,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

type JUnitSkipped struct {
	Message string `xml:"message,attr"`
}

// displayAuditSARIF generates SARIF format output
func displayAuditSARIF(data AuditData, outputFile string) error {
	report := SARIFReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:            "gzh-manager-audit",
						SemanticVersion: "1.0.0",
						FullName:        "GZH Manager Repository Compliance Audit",
						Rules:           []SARIFRule{},
					},
				},
				Results:          []SARIFResult{},
				ColumnKind:       "utf16CodeUnits",
				LogicalLocations: []LogicalLocation{},
			},
		},
	}

	// Create rules and results from violations
	ruleMap := make(map[string]int)
	for i, violation := range data.Violations {
		ruleID := fmt.Sprintf("%s.%s", violation.Policy, violation.Setting)

		// Add rule if not already added
		if _, exists := ruleMap[ruleID]; !exists {
			rule := SARIFRule{
				ID:   ruleID,
				Name: violation.Setting,
				ShortDescription: SARIFText{
					Text: violation.Description,
				},
				FullDescription: SARIFText{
					Text: fmt.Sprintf("Policy: %s - %s", violation.Policy, violation.Description),
				},
				Help: SARIFText{
					Text: violation.Remediation,
				},
				Properties: map[string]string{
					"severity": violation.Severity,
					"category": "compliance",
				},
			}
			report.Runs[0].Tool.Driver.Rules = append(report.Runs[0].Tool.Driver.Rules, rule)
			ruleMap[ruleID] = len(report.Runs[0].Tool.Driver.Rules) - 1
		}

		// Add result
		level := "error"
		if violation.Severity == "medium" || violation.Severity == "low" {
			level = "warning"
		}

		result := SARIFResult{
			RuleID:    ruleID,
			RuleIndex: ruleMap[ruleID],
			Level:     level,
			Message: SARIFText{
				Text: fmt.Sprintf("%s: Expected %s, but found %s",
					violation.Description, violation.Expected, violation.Actual),
			},
			Locations: []SARIFLocation{
				{
					LogicalLocation: &LogicalLocation{
						Name: violation.Repository,
						Kind: "repository",
					},
				},
			},
			Properties: map[string]interface{}{
				"policy":      violation.Policy,
				"expected":    violation.Expected,
				"actual":      violation.Actual,
				"remediation": violation.Remediation,
			},
		}
		report.Runs[0].Results = append(report.Runs[0].Results, result)

		// Add logical location
		if i == 0 || report.Runs[0].LogicalLocations[len(report.Runs[0].LogicalLocations)-1].Name != violation.Repository {
			report.Runs[0].LogicalLocations = append(report.Runs[0].LogicalLocations, LogicalLocation{
				Name: violation.Repository,
				Kind: "repository",
			})
		}
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SARIF report: %w", err)
	}

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0o644); err != nil {
			return fmt.Errorf("failed to write SARIF report: %w", err)
		}
		fmt.Printf("✅ SARIF report generated: %s\n", outputFile)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// displayAuditJUnit generates JUnit XML format output
func displayAuditJUnit(data AuditData, outputFile string) error {
	testSuites := JUnitTestSuites{
		Name:     fmt.Sprintf("Repository Compliance Audit - %s", data.Organization),
		Tests:    0,
		Failures: 0,
		Errors:   0,
		Time:     0.1, // Mock execution time
	}

	// Create test suite for each policy
	for _, policy := range data.PolicyCompliance {
		testSuite := JUnitTestSuite{
			Name:      policy.PolicyName,
			Tests:     0,
			Failures:  0,
			Errors:    0,
			Time:      0.01,
			Timestamp: data.GeneratedAt.Format(time.RFC3339),
			Properties: JUnitProperties{
				Properties: []JUnitProperty{
					{Name: "description", Value: policy.Description},
					{Name: "severity", Value: policy.Severity},
					{Name: "compliance_percentage", Value: fmt.Sprintf("%.1f", policy.CompliancePercentage)},
				},
			},
		}

		// Create test cases for each repository
		repoViolations := make(map[string][]ViolationDetail)
		for _, violation := range data.Violations {
			if violation.Policy == policy.PolicyName {
				repoViolations[violation.Repository] = append(repoViolations[violation.Repository], violation)
			}
		}

		// Add test case for each repository
		for _, repo := range data.Repositories {
			testCase := JUnitTestCase{
				ClassName: policy.PolicyName,
				Name:      fmt.Sprintf("%s compliance", repo.Name),
				Time:      0.001,
			}

			violations, hasViolations := repoViolations[repo.Name]
			if hasViolations {
				// Repository has violations for this policy
				var messages []string
				for _, v := range violations {
					messages = append(messages, fmt.Sprintf("%s: Expected %s, got %s",
						v.Setting, v.Expected, v.Actual))
				}

				testCase.Failure = &JUnitFailure{
					Message: fmt.Sprintf("%d violations found", len(violations)),
					Type:    "ComplianceViolation",
					Text:    strings.Join(messages, "\n"),
				}
				testSuite.Failures++
			}

			testSuite.TestCases = append(testSuite.TestCases, testCase)
			testSuite.Tests++
		}

		testSuites.TestSuites = append(testSuites.TestSuites, testSuite)
		testSuites.Tests += testSuite.Tests
		testSuites.Failures += testSuite.Failures
	}

	// Add summary information
	if len(testSuites.TestSuites) > 0 {
		testSuites.TestSuites[0].SystemOut = fmt.Sprintf(
			"Compliance Summary:\n"+
				"Total Repositories: %d\n"+
				"Compliant: %d (%.1f%%)\n"+
				"Total Violations: %d\n"+
				"Critical Violations: %d\n",
			data.Summary.TotalRepositories,
			data.Summary.CompliantRepositories,
			data.Summary.CompliancePercentage,
			data.Summary.TotalViolations,
			data.Summary.CriticalViolations,
		)
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(testSuites, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JUnit report: %w", err)
	}

	// Add XML declaration
	xmlOutput := []byte(xml.Header)
	xmlOutput = append(xmlOutput, xmlData...)

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, xmlOutput, 0o644); err != nil {
			return fmt.Errorf("failed to write JUnit report: %w", err)
		}
		fmt.Printf("✅ JUnit report generated: %s\n", outputFile)
	} else {
		fmt.Println(string(xmlOutput))
	}

	return nil
}

// Risk scoring structures
type RepositoryRiskScore struct {
	Repository      string       `json:"repository"`
	TotalScore      float64      `json:"total_score"` // 0-100 (100 = highest risk)
	RiskLevel       string       `json:"risk_level"`  // critical, high, medium, low
	RiskFactors     []RiskFactor `json:"risk_factors"`
	Recommendations []string     `json:"recommendations"`
}

type RiskFactor struct {
	Category    string  `json:"category"` // security, compliance, operational, exposure
	Name        string  `json:"name"`
	Score       float64 `json:"score"`  // Contribution to total risk
	Weight      float64 `json:"weight"` // Weight of this factor
	Description string  `json:"description"`
}

// calculateRepositoryRiskScores calculates risk scores for all repositories
func calculateRepositoryRiskScores(data AuditData) []RepositoryRiskScore {
	riskScores := []RepositoryRiskScore{}

	// Create violation map by repository
	repoViolations := make(map[string][]ViolationDetail)
	for _, violation := range data.Violations {
		repoViolations[violation.Repository] = append(repoViolations[violation.Repository], violation)
	}

	for _, repo := range data.Repositories {
		riskScore := RepositoryRiskScore{
			Repository:      repo.Name,
			TotalScore:      0,
			RiskFactors:     []RiskFactor{},
			Recommendations: []string{},
		}

		// Factor 1: Security violations (40% weight)
		securityScore := calculateSecurityRiskScore(repo, repoViolations[repo.Name])
		if securityScore > 0 {
			riskScore.RiskFactors = append(riskScore.RiskFactors, RiskFactor{
				Category:    "security",
				Name:        "Security Policy Violations",
				Score:       securityScore * 0.4,
				Weight:      0.4,
				Description: fmt.Sprintf("%d critical security violations", repo.CriticalCount),
			})
		}

		// Factor 2: Compliance violations (30% weight)
		complianceScore := calculateComplianceRiskScore(repo, repoViolations[repo.Name])
		if complianceScore > 0 {
			riskScore.RiskFactors = append(riskScore.RiskFactors, RiskFactor{
				Category:    "compliance",
				Name:        "Compliance Policy Violations",
				Score:       complianceScore * 0.3,
				Weight:      0.3,
				Description: fmt.Sprintf("%d total policy violations", repo.ViolationCount),
			})
		}

		// Factor 3: Exposure risk (20% weight) - public repos with violations
		exposureScore := calculateExposureRiskScore(repo)
		if exposureScore > 0 {
			riskScore.RiskFactors = append(riskScore.RiskFactors, RiskFactor{
				Category:    "exposure",
				Name:        "Public Exposure Risk",
				Score:       exposureScore * 0.2,
				Weight:      0.2,
				Description: "Public repository with security issues",
			})
		}

		// Factor 4: Operational risk (10% weight) - missing CI/CD, docs, etc.
		operationalScore := calculateOperationalRiskScore(repo, repoViolations[repo.Name])
		if operationalScore > 0 {
			riskScore.RiskFactors = append(riskScore.RiskFactors, RiskFactor{
				Category:    "operational",
				Name:        "Operational Risk",
				Score:       operationalScore * 0.1,
				Weight:      0.1,
				Description: "Missing operational best practices",
			})
		}

		// Calculate total risk score
		for _, factor := range riskScore.RiskFactors {
			riskScore.TotalScore += factor.Score
		}

		// Determine risk level
		riskScore.RiskLevel = determineRiskLevel(riskScore.TotalScore)

		// Generate recommendations
		riskScore.Recommendations = generateRiskRecommendations(repo, repoViolations[repo.Name], riskScore)

		riskScores = append(riskScores, riskScore)
	}

	return riskScores
}

// calculateSecurityRiskScore calculates security-related risk
func calculateSecurityRiskScore(repo RepositoryAudit, violations []ViolationDetail) float64 {
	score := 0.0

	// Critical security violations have highest impact
	criticalViolations := 0
	highViolations := 0

	for _, v := range violations {
		if strings.Contains(strings.ToLower(v.Policy), "security") ||
			strings.Contains(strings.ToLower(v.Policy), "protection") ||
			strings.Contains(strings.ToLower(v.Policy), "vulnerability") {
			if v.Severity == "critical" {
				criticalViolations++
			} else if v.Severity == "high" {
				highViolations++
			}
		}
	}

	// Score based on violation count and severity
	score = float64(criticalViolations)*40.0 + float64(highViolations)*20.0

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateComplianceRiskScore calculates compliance-related risk
func calculateComplianceRiskScore(repo RepositoryAudit, violations []ViolationDetail) float64 {
	if len(violations) == 0 {
		return 0
	}

	// Base score on violation count relative to policy count
	violationRatio := float64(len(violations)) / 10.0 // Assume 10 policies as baseline
	score := violationRatio * 100

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateExposureRiskScore calculates exposure-related risk
func calculateExposureRiskScore(repo RepositoryAudit) float64 {
	score := 0.0

	// Public repositories with violations are higher risk
	if repo.Visibility == "public" && repo.ViolationCount > 0 {
		score = 50.0

		// Additional risk for critical violations in public repos
		if repo.CriticalCount > 0 {
			score += float64(repo.CriticalCount) * 10.0
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateOperationalRiskScore calculates operational risk
func calculateOperationalRiskScore(repo RepositoryAudit, violations []ViolationDetail) float64 {
	score := 0.0

	// Check for missing operational controls
	for _, v := range violations {
		if strings.Contains(strings.ToLower(v.Policy), "ci") ||
			strings.Contains(strings.ToLower(v.Policy), "documentation") ||
			strings.Contains(strings.ToLower(v.Policy), "workflow") {
			score += 20.0
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// determineRiskLevel determines risk level based on score
func determineRiskLevel(score float64) string {
	if score >= 75 {
		return "critical"
	} else if score >= 50 {
		return "high"
	} else if score >= 25 {
		return "medium"
	}
	return "low"
}

// generateRiskRecommendations generates recommendations based on risk factors
func generateRiskRecommendations(repo RepositoryAudit, violations []ViolationDetail, riskScore RepositoryRiskScore) []string {
	recommendations := []string{}

	// Critical security recommendations
	if repo.CriticalCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🚨 URGENT: Fix %d critical security violations immediately", repo.CriticalCount))
	}

	// Exposure recommendations
	if repo.Visibility == "public" && repo.ViolationCount > 0 {
		recommendations = append(recommendations,
			"⚠️ Consider making repository private until violations are resolved")
	}

	// Factor-specific recommendations
	for _, factor := range riskScore.RiskFactors {
		switch factor.Category {
		case "security":
			if factor.Score > 30 {
				recommendations = append(recommendations,
					"🛡️ Enable branch protection and require code reviews",
					"🔒 Enable security scanning and vulnerability alerts")
			}
		case "compliance":
			if factor.Score > 20 {
				recommendations = append(recommendations,
					"📋 Review and implement required compliance policies",
					"📊 Schedule regular compliance audits")
			}
		case "operational":
			if factor.Score > 5 {
				recommendations = append(recommendations,
					"🔧 Implement CI/CD pipelines for automated testing",
					"📚 Add required documentation (README, LICENSE, SECURITY.md)")
			}
		}
	}

	// General recommendations based on risk level
	switch riskScore.RiskLevel {
	case "critical":
		recommendations = append(recommendations,
			"🚫 Block all deployments until critical issues are resolved",
			"👥 Assign security team to review immediately")
	case "high":
		recommendations = append(recommendations,
			"📅 Schedule immediate remediation (within 7 days)",
			"🔍 Conduct security review before next release")
	case "medium":
		recommendations = append(recommendations,
			"📋 Add to remediation backlog (resolve within 30 days)",
			"📊 Monitor for any increase in violations")
	}

	return recommendations
}

// Enhanced audit data with risk scores
type EnhancedAuditData struct {
	AuditData
	RiskAnalysis RiskAnalysis `json:"risk_analysis"`
}

type RiskAnalysis struct {
	OverallRiskLevel  string                `json:"overall_risk_level"`
	HighRiskRepos     int                   `json:"high_risk_repos"`
	CriticalRiskRepos int                   `json:"critical_risk_repos"`
	RiskDistribution  map[string]int        `json:"risk_distribution"`
	TopRisks          []RepositoryRiskScore `json:"top_risks"`
	RiskTrend         string                `json:"risk_trend"` // increasing, decreasing, stable
}

// enhanceAuditDataWithRiskScores adds risk analysis to audit data
func enhanceAuditDataWithRiskScores(data AuditData) EnhancedAuditData {
	enhanced := EnhancedAuditData{
		AuditData: data,
	}

	// Calculate risk scores
	riskScores := calculateRepositoryRiskScores(data)

	// Analyze risk distribution
	riskDist := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	criticalCount := 0
	highCount := 0
	var topRisks []RepositoryRiskScore

	for _, risk := range riskScores {
		riskDist[risk.RiskLevel]++

		if risk.RiskLevel == "critical" {
			criticalCount++
			topRisks = append(topRisks, risk)
		} else if risk.RiskLevel == "high" {
			highCount++
			if len(topRisks) < 10 {
				topRisks = append(topRisks, risk)
			}
		}
	}

	// Determine overall risk level
	overallRisk := "low"
	if criticalCount > 0 {
		overallRisk = "critical"
	} else if highCount > 2 || (highCount > 0 && data.Summary.CompliancePercentage < 70) {
		overallRisk = "high"
	} else if riskDist["medium"] > len(riskScores)/2 {
		overallRisk = "medium"
	}

	enhanced.RiskAnalysis = RiskAnalysis{
		OverallRiskLevel:  overallRisk,
		HighRiskRepos:     highCount,
		CriticalRiskRepos: criticalCount,
		RiskDistribution:  riskDist,
		TopRisks:          topRisks,
		RiskTrend:         "stable", // Would need historical data for real trend
	}

	return enhanced
}

package repoconfig

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// newRiskAssessmentCmd creates the risk assessment subcommand
func newRiskAssessmentCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		format         string
		outputFile     string
		threshold      string
		showDetails    bool
		sortBy         string
		includeMetrics bool
	)

	cmd := &cobra.Command{
		Use:   "risk-assessment",
		Short: "Perform CVSS-based risk assessment of policy violations",
		Long: `Perform comprehensive risk assessment of policy violations using CVSS methodology.

This command analyzes policy violations and calculates business risk scores based on
CVSS (Common Vulnerability Scoring System) principles adapted for repository configurations.

Risk Assessment Features:
- CVSS-based scoring methodology
- Business impact analysis
- Priority-based fix recommendations
- Risk trend analysis
- Escalation threshold management
- Custom risk metrics calculation

Risk Levels:
- Critical: 9.0-10.0 (Immediate action required)
- High: 7.0-8.9 (High priority fixes)
- Medium: 4.0-6.9 (Medium priority fixes)
- Low: 0.1-3.9 (Low priority fixes)
- None: 0.0 (No risk)

Examples:
  gz repo-config risk-assessment --org myorg                     # Full risk assessment
  gz repo-config risk-assessment --org myorg --threshold critical # Critical risks only
  gz repo-config risk-assessment --org myorg --sort-by score     # Sort by risk score
  gz repo-config risk-assessment --org myorg --show-details      # Detailed risk breakdown
  gz repo-config risk-assessment --org myorg --include-metrics   # Include business metrics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRiskAssessmentCommand(flags, format, outputFile, threshold, showDetails, sortBy, includeMetrics)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add risk assessment specific flags
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, csv)")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().StringVar(&threshold, "threshold", "all", "Risk threshold filter (critical, high, medium, low, all)")
	cmd.Flags().BoolVar(&showDetails, "show-details", false, "Show detailed risk breakdown")
	cmd.Flags().StringVar(&sortBy, "sort-by", "score", "Sort by (score, repository, policy, impact)")
	cmd.Flags().BoolVar(&includeMetrics, "include-metrics", false, "Include business risk metrics")

	return cmd
}

// runRiskAssessmentCommand executes the risk assessment command
func runRiskAssessmentCommand(flags GlobalFlags, format, outputFile, threshold string, showDetails bool, sortBy string, includeMetrics bool) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("🎯 Performing risk assessment for organization: %s\n", flags.Organization)
		fmt.Printf("Risk threshold: %s\n", threshold)
		fmt.Printf("Sort by: %s\n", sortBy)
		fmt.Printf("Format: %s\n", format)
		fmt.Println()
	}

	// Perform compliance audit to get violation data
	auditData, err := performComplianceAudit(flags.Organization, "")
	if err != nil {
		return fmt.Errorf("failed to perform compliance audit: %w", err)
	}

	// Calculate risk assessments
	riskAssessments := calculateRiskAssessments(auditData.Violations)

	// Filter by threshold
	filteredAssessments := filterByThreshold(riskAssessments, threshold)

	// Sort assessments
	sortRiskAssessments(filteredAssessments, sortBy)

	// Generate business metrics if requested
	var businessMetrics *BusinessRiskMetrics
	if includeMetrics {
		businessMetrics = calculateBusinessRiskMetrics(riskAssessments, auditData)
	}

	// Generate output
	switch format {
	case "json":
		return outputRiskAssessmentJSON(filteredAssessments, businessMetrics, outputFile)
	case "csv":
		return outputRiskAssessmentCSV(filteredAssessments, outputFile)
	case "table":
		return outputRiskAssessmentTable(filteredAssessments, businessMetrics, showDetails)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// RiskAssessment represents a comprehensive risk assessment for a policy violation
type RiskAssessment struct {
	ID           string              `json:"id"`
	Repository   string              `json:"repository"`
	Policy       string              `json:"policy"`
	Setting      string              `json:"setting"`
	Violation    string              `json:"violation"`
	CVSSScore    float64             `json:"cvss_score"`
	RiskLevel    string              `json:"risk_level"`
	BusinessRisk BusinessRiskFactor  `json:"business_risk"`
	CVSSVector   CVSSVector          `json:"cvss_vector"`
	Impact       ImpactAssessment    `json:"impact"`
	Remediation  RemediationGuidance `json:"remediation"`
	Timeline     RiskTimeline        `json:"timeline"`
	Priority     int                 `json:"priority"`
	Escalation   EscalationLevel     `json:"escalation"`
}

// CVSSVector represents CVSS scoring components
type CVSSVector struct {
	AttackVector       string  `json:"attack_vector"`       // Network, Adjacent, Local, Physical
	AttackComplexity   string  `json:"attack_complexity"`   // Low, High
	PrivilegesRequired string  `json:"privileges_required"` // None, Low, High
	UserInteraction    string  `json:"user_interaction"`    // None, Required
	Scope              string  `json:"scope"`               // Unchanged, Changed
	Confidentiality    string  `json:"confidentiality"`     // None, Low, High
	Integrity          string  `json:"integrity"`           // None, Low, High
	Availability       string  `json:"availability"`        // None, Low, High
	BaseScore          float64 `json:"base_score"`
	TemporalScore      float64 `json:"temporal_score"`
	EnvironmentalScore float64 `json:"environmental_score"`
}

// BusinessRiskFactor represents business-specific risk factors
type BusinessRiskFactor struct {
	DataSensitivity     string  `json:"data_sensitivity"`     // Public, Internal, Confidential, Restricted
	BusinessCriticality string  `json:"business_criticality"` // Low, Medium, High, Critical
	ComplianceImpact    string  `json:"compliance_impact"`    // None, Low, Medium, High
	ReputationRisk      string  `json:"reputation_risk"`      // Low, Medium, High
	FinancialImpact     float64 `json:"financial_impact"`     // Estimated cost in USD
	CustomerImpact      string  `json:"customer_impact"`      // None, Low, Medium, High
}

// ImpactAssessment represents the potential impact of a violation
type ImpactAssessment struct {
	SecurityImpact    string   `json:"security_impact"`
	ComplianceImpact  string   `json:"compliance_impact"`
	OperationalImpact string   `json:"operational_impact"`
	AffectedSystems   []string `json:"affected_systems"`
	ExposureLevel     string   `json:"exposure_level"`
	LikelihoodExploit string   `json:"likelihood_exploit"`
}

// RemediationGuidance provides detailed remediation guidance
type RemediationGuidance struct {
	Recommendation  string        `json:"recommendation"`
	Steps           []string      `json:"steps"`
	EstimatedEffort string        `json:"estimated_effort"`
	RequiredSkills  []string      `json:"required_skills"`
	Dependencies    []string      `json:"dependencies"`
	RiskReduction   float64       `json:"risk_reduction"`
	Cost            float64       `json:"cost"`
	Timeline        time.Duration `json:"timeline"`
}

// RiskTimeline represents risk timeline information
type RiskTimeline struct {
	FirstDetected    time.Time `json:"first_detected"`
	LastAssessed     time.Time `json:"last_assessed"`
	ExposureDuration string    `json:"exposure_duration"`
	TimeToFix        string    `json:"time_to_fix"`
	SLADeadline      string    `json:"sla_deadline"`
	DaysOverdue      int       `json:"days_overdue"`
}

// EscalationLevel represents escalation requirements
type EscalationLevel struct {
	Level             string    `json:"level"` // None, Management, Executive, Board
	RequiredBy        time.Time `json:"required_by"`
	NotificationsSent int       `json:"notifications_sent"`
	Stakeholders      []string  `json:"stakeholders"`
	EscalationReason  string    `json:"escalation_reason"`
}

// BusinessRiskMetrics provides organization-wide risk metrics
type BusinessRiskMetrics struct {
	TotalRiskScore       float64               `json:"total_risk_score"`
	AverageRiskScore     float64               `json:"average_risk_score"`
	RiskDistribution     map[string]int        `json:"risk_distribution"`
	BusinessImpactScore  float64               `json:"business_impact_score"`
	ComplianceRiskScore  float64               `json:"compliance_risk_score"`
	SecurityRiskScore    float64               `json:"security_risk_score"`
	EstimatedCost        float64               `json:"estimated_cost"`
	RiskTrend            string                `json:"risk_trend"`
	CriticalRiskCount    int                   `json:"critical_risk_count"`
	TopRiskCategories    []RiskCategoryRank    `json:"top_risk_categories"`
	EscalationRequired   bool                  `json:"escalation_required"`
	ComplianceViolations []ComplianceViolation `json:"compliance_violations"`
}

// RiskCategoryRank represents risk ranking by category
type RiskCategoryRank struct {
	Category       string  `json:"category"`
	RiskScore      float64 `json:"risk_score"`
	ViolationCount int     `json:"violation_count"`
	AverageScore   float64 `json:"average_score"`
}

// ComplianceViolation represents a compliance-related violation
type ComplianceViolation struct {
	Standard       string  `json:"standard"` // SOC2, ISO27001, GDPR, etc.
	Requirement    string  `json:"requirement"`
	ViolationCount int     `json:"violation_count"`
	RiskScore      float64 `json:"risk_score"`
	Severity       string  `json:"severity"`
}

// calculateRiskAssessments calculates risk assessments for all violations
func calculateRiskAssessments(violations []ViolationDetail) []RiskAssessment {
	var assessments []RiskAssessment

	for i, violation := range violations {
		assessment := RiskAssessment{
			ID:         fmt.Sprintf("risk-%d", i+1),
			Repository: violation.Repository,
			Policy:     violation.Policy,
			Setting:    violation.Setting,
			Violation:  violation.Description,
		}

		// Calculate CVSS score
		assessment.CVSSVector = calculateCVSSVector(violation)
		assessment.CVSSScore = assessment.CVSSVector.BaseScore

		// Determine risk level
		assessment.RiskLevel = getRiskLevel(assessment.CVSSScore)

		// Calculate business risk factors
		assessment.BusinessRisk = calculateBusinessRisk(violation, assessment.CVSSScore)

		// Perform impact assessment
		assessment.Impact = assessImpact(violation, assessment.CVSSScore)

		// Generate remediation guidance
		assessment.Remediation = generateRemediationGuidance(violation, assessment.CVSSScore)

		// Set timeline information
		assessment.Timeline = generateRiskTimeline(violation)

		// Calculate priority
		assessment.Priority = calculatePriority(assessment.CVSSScore, assessment.BusinessRisk)

		// Determine escalation requirements
		assessment.Escalation = determineEscalation(assessment.CVSSScore, assessment.BusinessRisk)

		assessments = append(assessments, assessment)
	}

	return assessments
}

// calculateCVSSVector calculates CVSS vector based on violation details
func calculateCVSSVector(violation ViolationDetail) CVSSVector {
	vector := CVSSVector{
		AttackVector:       "Network",
		AttackComplexity:   "Low",
		PrivilegesRequired: "None",
		UserInteraction:    "None",
		Scope:              "Unchanged",
		Confidentiality:    "Low",
		Integrity:          "Low",
		Availability:       "None",
	}

	// Adjust based on violation type
	switch {
	case strings.Contains(strings.ToLower(violation.Policy), "branch protection"):
		vector.AttackVector = "Network"
		vector.AttackComplexity = "Low"
		vector.PrivilegesRequired = "Low"
		vector.Integrity = "High"
		vector.Availability = "Low"
	case strings.Contains(strings.ToLower(violation.Policy), "security"):
		vector.AttackVector = "Network"
		vector.AttackComplexity = "Low"
		vector.PrivilegesRequired = "None"
		vector.Confidentiality = "High"
		vector.Integrity = "High"
		vector.Availability = "Low"
	case strings.Contains(strings.ToLower(violation.Policy), "access"):
		vector.AttackVector = "Network"
		vector.AttackComplexity = "Low"
		vector.PrivilegesRequired = "None"
		vector.Confidentiality = "High"
		vector.Integrity = "Medium"
		vector.Availability = "None"
	case strings.Contains(strings.ToLower(violation.Policy), "visibility"):
		vector.AttackVector = "Network"
		vector.AttackComplexity = "Low"
		vector.PrivilegesRequired = "None"
		vector.Confidentiality = "High"
		vector.Integrity = "Low"
		vector.Availability = "None"
	}

	// Adjust based on severity
	switch violation.Severity {
	case "critical":
		vector.Confidentiality = "High"
		vector.Integrity = "High"
		vector.Availability = "High"
	case "high":
		vector.Confidentiality = "High"
		vector.Integrity = "High"
		vector.Availability = "Low"
	case "medium":
		vector.Confidentiality = "Low"
		vector.Integrity = "Low"
		vector.Availability = "Low"
	case "low":
		vector.Confidentiality = "None"
		vector.Integrity = "Low"
		vector.Availability = "None"
	}

	// Calculate base score
	vector.BaseScore = calculateCVSSBaseScore(vector)
	vector.TemporalScore = vector.BaseScore * 0.95     // Assume exploit code is available
	vector.EnvironmentalScore = vector.BaseScore * 1.1 // Assume high business impact

	return vector
}

// calculateCVSSBaseScore calculates CVSS base score from vector components
func calculateCVSSBaseScore(vector CVSSVector) float64 {
	// CVSS 3.1 Base Score calculation
	// This is a simplified version - actual CVSS calculation is more complex

	// Impact Sub-Score
	confImpact := getCVSSMetricValue(vector.Confidentiality, "impact")
	intImpact := getCVSSMetricValue(vector.Integrity, "impact")
	availImpact := getCVSSMetricValue(vector.Availability, "impact")

	impactSubScore := 1 - ((1 - confImpact) * (1 - intImpact) * (1 - availImpact))

	// Exploitability Sub-Score
	attackVector := getCVSSMetricValue(vector.AttackVector, "av")
	attackComplexity := getCVSSMetricValue(vector.AttackComplexity, "ac")
	privilegesRequired := getCVSSMetricValue(vector.PrivilegesRequired, "pr")
	userInteraction := getCVSSMetricValue(vector.UserInteraction, "ui")

	exploitabilitySubScore := 8.22 * attackVector * attackComplexity * privilegesRequired * userInteraction

	// Base Score
	if impactSubScore <= 0 {
		return 0
	}

	var baseScore float64
	if vector.Scope == "Unchanged" {
		baseScore = math.Min(impactSubScore+exploitabilitySubScore, 10)
	} else {
		baseScore = math.Min(1.08*(impactSubScore+exploitabilitySubScore), 10)
	}

	return math.Round(baseScore*10) / 10
}

// getCVSSMetricValue returns the numeric value for CVSS metrics
func getCVSSMetricValue(metric, metricType string) float64 {
	switch metricType {
	case "impact":
		switch metric {
		case "None":
			return 0
		case "Low":
			return 0.22
		case "High":
			return 0.56
		default:
			return 0
		}
	case "av":
		switch metric {
		case "Network":
			return 0.85
		case "Adjacent":
			return 0.62
		case "Local":
			return 0.55
		case "Physical":
			return 0.2
		default:
			return 0.85
		}
	case "ac":
		switch metric {
		case "Low":
			return 0.77
		case "High":
			return 0.44
		default:
			return 0.77
		}
	case "pr":
		switch metric {
		case "None":
			return 0.85
		case "Low":
			return 0.62
		case "High":
			return 0.27
		default:
			return 0.85
		}
	case "ui":
		switch metric {
		case "None":
			return 0.85
		case "Required":
			return 0.62
		default:
			return 0.85
		}
	}
	return 0
}

// getRiskLevel determines risk level based on CVSS score
func getRiskLevel(score float64) string {
	switch {
	case score >= 9.0:
		return "Critical"
	case score >= 7.0:
		return "High"
	case score >= 4.0:
		return "Medium"
	case score >= 0.1:
		return "Low"
	default:
		return "None"
	}
}

// calculateBusinessRisk calculates business-specific risk factors
func calculateBusinessRisk(violation ViolationDetail, cvssScore float64) BusinessRiskFactor {
	risk := BusinessRiskFactor{
		DataSensitivity:     "Internal",
		BusinessCriticality: "Medium",
		ComplianceImpact:    "Medium",
		ReputationRisk:      "Low",
		CustomerImpact:      "Low",
	}

	// Adjust based on repository characteristics
	if strings.Contains(strings.ToLower(violation.Repository), "api") ||
		strings.Contains(strings.ToLower(violation.Repository), "service") {
		risk.DataSensitivity = "Confidential"
		risk.BusinessCriticality = "High"
		risk.CustomerImpact = "Medium"
	}

	if strings.Contains(strings.ToLower(violation.Repository), "public") ||
		strings.Contains(strings.ToLower(violation.Repository), "doc") {
		risk.DataSensitivity = "Public"
		risk.BusinessCriticality = "Low"
		risk.ReputationRisk = "Medium"
	}

	// Adjust based on violation severity
	switch violation.Severity {
	case "critical":
		risk.BusinessCriticality = "Critical"
		risk.ComplianceImpact = "High"
		risk.ReputationRisk = "High"
		risk.CustomerImpact = "High"
		risk.FinancialImpact = 100000 // $100K estimated impact
	case "high":
		risk.BusinessCriticality = "High"
		risk.ComplianceImpact = "Medium"
		risk.ReputationRisk = "Medium"
		risk.CustomerImpact = "Medium"
		risk.FinancialImpact = 50000 // $50K estimated impact
	case "medium":
		risk.BusinessCriticality = "Medium"
		risk.ComplianceImpact = "Low"
		risk.ReputationRisk = "Low"
		risk.CustomerImpact = "Low"
		risk.FinancialImpact = 10000 // $10K estimated impact
	case "low":
		risk.BusinessCriticality = "Low"
		risk.ComplianceImpact = "None"
		risk.ReputationRisk = "Low"
		risk.CustomerImpact = "None"
		risk.FinancialImpact = 1000 // $1K estimated impact
	}

	return risk
}

// assessImpact performs impact assessment
func assessImpact(violation ViolationDetail, cvssScore float64) ImpactAssessment {
	impact := ImpactAssessment{
		SecurityImpact:    "Medium",
		ComplianceImpact:  "Medium",
		OperationalImpact: "Low",
		AffectedSystems:   []string{violation.Repository},
		ExposureLevel:     "Medium",
		LikelihoodExploit: "Medium",
	}

	// Adjust based on CVSS score
	if cvssScore >= 9.0 {
		impact.SecurityImpact = "Critical"
		impact.ComplianceImpact = "High"
		impact.OperationalImpact = "High"
		impact.ExposureLevel = "High"
		impact.LikelihoodExploit = "High"
	} else if cvssScore >= 7.0 {
		impact.SecurityImpact = "High"
		impact.ComplianceImpact = "Medium"
		impact.OperationalImpact = "Medium"
		impact.ExposureLevel = "Medium"
		impact.LikelihoodExploit = "Medium"
	} else if cvssScore >= 4.0 {
		impact.SecurityImpact = "Medium"
		impact.ComplianceImpact = "Low"
		impact.OperationalImpact = "Low"
		impact.ExposureLevel = "Low"
		impact.LikelihoodExploit = "Low"
	} else {
		impact.SecurityImpact = "Low"
		impact.ComplianceImpact = "None"
		impact.OperationalImpact = "None"
		impact.ExposureLevel = "Low"
		impact.LikelihoodExploit = "Low"
	}

	return impact
}

// generateRemediationGuidance generates detailed remediation guidance
func generateRemediationGuidance(violation ViolationDetail, cvssScore float64) RemediationGuidance {
	guidance := RemediationGuidance{
		Recommendation: violation.Remediation,
		Steps:          []string{},
		RequiredSkills: []string{"GitHub Administration"},
		Dependencies:   []string{},
		Cost:           0,
		Timeline:       time.Hour * 24,
	}

	// Generate specific steps based on violation type
	switch {
	case strings.Contains(strings.ToLower(violation.Policy), "branch protection"):
		guidance.Steps = []string{
			"Navigate to repository settings",
			"Select 'Branches' tab",
			"Add branch protection rule for main branch",
			"Configure required status checks",
			"Enable 'Require pull request reviews'",
			"Test the protection rule",
		}
		guidance.RequiredSkills = append(guidance.RequiredSkills, "Branch Protection Configuration")
		guidance.Timeline = time.Hour * 2
	case strings.Contains(strings.ToLower(violation.Policy), "security"):
		guidance.Steps = []string{
			"Navigate to repository settings",
			"Select 'Security & analysis' tab",
			"Enable vulnerability alerts",
			"Configure security scanning",
			"Set up dependency scanning",
			"Review security policies",
		}
		guidance.RequiredSkills = append(guidance.RequiredSkills, "Security Configuration")
		guidance.Timeline = time.Hour * 4
	case strings.Contains(strings.ToLower(violation.Policy), "access"):
		guidance.Steps = []string{
			"Review current access permissions",
			"Identify unauthorized users",
			"Remove unnecessary access",
			"Configure team-based access",
			"Set up access reviews",
			"Document access changes",
		}
		guidance.RequiredSkills = append(guidance.RequiredSkills, "Access Management")
		guidance.Timeline = time.Hour * 8
	}

	// Calculate risk reduction
	guidance.RiskReduction = cvssScore * 0.9 // Assume 90% risk reduction

	// Adjust effort based on CVSS score
	if cvssScore >= 9.0 {
		guidance.EstimatedEffort = "High"
		guidance.Timeline = guidance.Timeline * 2
		guidance.Cost = 5000
	} else if cvssScore >= 7.0 {
		guidance.EstimatedEffort = "Medium"
		guidance.Timeline = guidance.Timeline * 1.5
		guidance.Cost = 2000
	} else if cvssScore >= 4.0 {
		guidance.EstimatedEffort = "Low"
		guidance.Cost = 500
	} else {
		guidance.EstimatedEffort = "Minimal"
		guidance.Cost = 100
	}

	return guidance
}

// generateRiskTimeline generates risk timeline information
func generateRiskTimeline(violation ViolationDetail) RiskTimeline {
	now := time.Now()
	firstDetected := now.Add(-time.Hour * 24 * 7) // Assume detected 7 days ago

	timeline := RiskTimeline{
		FirstDetected:    firstDetected,
		LastAssessed:     now,
		ExposureDuration: formatDuration(now.Sub(firstDetected)),
		TimeToFix:        "2 hours",
		SLADeadline:      "24 hours",
		DaysOverdue:      0,
	}

	// Adjust based on severity
	switch violation.Severity {
	case "critical":
		timeline.TimeToFix = "1 hour"
		timeline.SLADeadline = "4 hours"
		if now.Sub(firstDetected) > time.Hour*4 {
			timeline.DaysOverdue = int(now.Sub(firstDetected).Hours() / 24)
		}
	case "high":
		timeline.TimeToFix = "4 hours"
		timeline.SLADeadline = "24 hours"
		if now.Sub(firstDetected) > time.Hour*24 {
			timeline.DaysOverdue = int(now.Sub(firstDetected).Hours() / 24)
		}
	case "medium":
		timeline.TimeToFix = "1 day"
		timeline.SLADeadline = "7 days"
		if now.Sub(firstDetected) > time.Hour*24*7 {
			timeline.DaysOverdue = int(now.Sub(firstDetected).Hours() / 24)
		}
	case "low":
		timeline.TimeToFix = "1 week"
		timeline.SLADeadline = "30 days"
		if now.Sub(firstDetected) > time.Hour*24*30 {
			timeline.DaysOverdue = int(now.Sub(firstDetected).Hours() / 24)
		}
	}

	return timeline
}

// calculatePriority calculates priority score based on CVSS and business factors
func calculatePriority(cvssScore float64, businessRisk BusinessRiskFactor) int {
	priority := int(cvssScore)

	// Adjust based on business criticality
	switch businessRisk.BusinessCriticality {
	case "Critical":
		priority += 3
	case "High":
		priority += 2
	case "Medium":
		priority += 1
	}

	// Adjust based on compliance impact
	switch businessRisk.ComplianceImpact {
	case "High":
		priority += 2
	case "Medium":
		priority += 1
	}

	// Ensure priority is within valid range
	if priority < 1 {
		priority = 1
	}
	if priority > 10 {
		priority = 10
	}

	return priority
}

// determineEscalation determines escalation requirements
func determineEscalation(cvssScore float64, businessRisk BusinessRiskFactor) EscalationLevel {
	escalation := EscalationLevel{
		Level:             "None",
		RequiredBy:        time.Now().Add(time.Hour * 24 * 30),
		NotificationsSent: 0,
		Stakeholders:      []string{},
		EscalationReason:  "No escalation required",
	}

	// Determine escalation based on CVSS score and business risk
	if cvssScore >= 9.0 || businessRisk.BusinessCriticality == "Critical" {
		escalation.Level = "Executive"
		escalation.RequiredBy = time.Now().Add(time.Hour * 2)
		escalation.Stakeholders = []string{"CISO", "CTO", "CEO"}
		escalation.EscalationReason = "Critical security risk requires immediate executive attention"
	} else if cvssScore >= 7.0 || businessRisk.BusinessCriticality == "High" {
		escalation.Level = "Management"
		escalation.RequiredBy = time.Now().Add(time.Hour * 8)
		escalation.Stakeholders = []string{"Security Manager", "Engineering Manager"}
		escalation.EscalationReason = "High-priority security issue requires management oversight"
	} else if cvssScore >= 4.0 && businessRisk.ComplianceImpact == "High" {
		escalation.Level = "Management"
		escalation.RequiredBy = time.Now().Add(time.Hour * 24)
		escalation.Stakeholders = []string{"Compliance Officer", "Security Team Lead"}
		escalation.EscalationReason = "Compliance-related security issue requires management review"
	}

	return escalation
}

// Helper functions

// filterByThreshold filters risk assessments by threshold
func filterByThreshold(assessments []RiskAssessment, threshold string) []RiskAssessment {
	if threshold == "all" {
		return assessments
	}

	var filtered []RiskAssessment
	for _, assessment := range assessments {
		if strings.EqualFold(assessment.RiskLevel, threshold) {
			filtered = append(filtered, assessment)
		}
	}
	return filtered
}

// sortRiskAssessments sorts risk assessments by specified criteria
func sortRiskAssessments(assessments []RiskAssessment, sortBy string) {
	switch sortBy {
	case "score":
		sort.Slice(assessments, func(i, j int) bool {
			return assessments[i].CVSSScore > assessments[j].CVSSScore
		})
	case "repository":
		sort.Slice(assessments, func(i, j int) bool {
			return assessments[i].Repository < assessments[j].Repository
		})
	case "policy":
		sort.Slice(assessments, func(i, j int) bool {
			return assessments[i].Policy < assessments[j].Policy
		})
	case "impact":
		sort.Slice(assessments, func(i, j int) bool {
			return assessments[i].BusinessRisk.FinancialImpact > assessments[j].BusinessRisk.FinancialImpact
		})
	default:
		// Default to score-based sorting
		sort.Slice(assessments, func(i, j int) bool {
			return assessments[i].CVSSScore > assessments[j].CVSSScore
		})
	}
}

// formatDuration formats time duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < time.Hour*24 {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}

// calculateBusinessRiskMetrics calculates organization-wide business risk metrics
func calculateBusinessRiskMetrics(assessments []RiskAssessment, auditData AuditData) *BusinessRiskMetrics {
	metrics := &BusinessRiskMetrics{
		RiskDistribution:     make(map[string]int),
		TopRiskCategories:    []RiskCategoryRank{},
		ComplianceViolations: []ComplianceViolation{},
	}

	totalScore := 0.0
	totalCost := 0.0
	criticalCount := 0
	escalationRequired := false

	// Process each assessment
	for _, assessment := range assessments {
		// Add to total score
		totalScore += assessment.CVSSScore

		// Add to total cost
		totalCost += assessment.BusinessRisk.FinancialImpact

		// Count risk distribution
		metrics.RiskDistribution[assessment.RiskLevel]++

		// Count critical risks
		if assessment.RiskLevel == "Critical" {
			criticalCount++
		}

		// Check for escalation requirements
		if assessment.Escalation.Level != "None" {
			escalationRequired = true
		}
	}

	// Calculate averages
	if len(assessments) > 0 {
		metrics.AverageRiskScore = totalScore / float64(len(assessments))
	}
	metrics.TotalRiskScore = totalScore
	metrics.EstimatedCost = totalCost
	metrics.CriticalRiskCount = criticalCount
	metrics.EscalationRequired = escalationRequired

	// Calculate component scores
	metrics.SecurityRiskScore = calculateSecurityRiskScore(assessments)
	metrics.ComplianceRiskScore = calculateComplianceRiskScore(assessments)
	metrics.BusinessImpactScore = calculateBusinessImpactScore(assessments)

	// Determine risk trend (simplified)
	if metrics.AverageRiskScore > 7.0 {
		metrics.RiskTrend = "increasing"
	} else if metrics.AverageRiskScore < 3.0 {
		metrics.RiskTrend = "decreasing"
	} else {
		metrics.RiskTrend = "stable"
	}

	// Generate top risk categories
	metrics.TopRiskCategories = generateTopRiskCategories(assessments)

	// Generate compliance violations
	metrics.ComplianceViolations = generateComplianceViolations(assessments)

	return metrics
}

// calculateSecurityRiskScore calculates security-specific risk score
func calculateSecurityRiskScore(assessments []RiskAssessment) float64 {
	totalScore := 0.0
	securityCount := 0

	for _, assessment := range assessments {
		if strings.Contains(strings.ToLower(assessment.Policy), "security") ||
			strings.Contains(strings.ToLower(assessment.Policy), "branch protection") {
			totalScore += assessment.CVSSScore
			securityCount++
		}
	}

	if securityCount > 0 {
		return totalScore / float64(securityCount)
	}
	return 0.0
}

// calculateComplianceRiskScore calculates compliance-specific risk score
func calculateComplianceRiskScore(assessments []RiskAssessment) float64 {
	totalScore := 0.0
	complianceCount := 0

	for _, assessment := range assessments {
		if assessment.BusinessRisk.ComplianceImpact != "None" {
			totalScore += assessment.CVSSScore
			complianceCount++
		}
	}

	if complianceCount > 0 {
		return totalScore / float64(complianceCount)
	}
	return 0.0
}

// calculateBusinessImpactScore calculates business impact score
func calculateBusinessImpactScore(assessments []RiskAssessment) float64 {
	totalImpact := 0.0

	for _, assessment := range assessments {
		businessMultiplier := 1.0
		switch assessment.BusinessRisk.BusinessCriticality {
		case "Critical":
			businessMultiplier = 2.0
		case "High":
			businessMultiplier = 1.5
		case "Medium":
			businessMultiplier = 1.0
		case "Low":
			businessMultiplier = 0.5
		}
		totalImpact += assessment.CVSSScore * businessMultiplier
	}

	return totalImpact
}

// generateTopRiskCategories generates top risk categories ranking
func generateTopRiskCategories(assessments []RiskAssessment) []RiskCategoryRank {
	categoryMap := make(map[string]*RiskCategoryRank)

	for _, assessment := range assessments {
		category := assessment.Policy
		if rank, exists := categoryMap[category]; exists {
			rank.RiskScore += assessment.CVSSScore
			rank.ViolationCount++
		} else {
			categoryMap[category] = &RiskCategoryRank{
				Category:       category,
				RiskScore:      assessment.CVSSScore,
				ViolationCount: 1,
			}
		}
	}

	// Calculate average scores and convert to slice
	var ranks []RiskCategoryRank
	for _, rank := range categoryMap {
		rank.AverageScore = rank.RiskScore / float64(rank.ViolationCount)
		ranks = append(ranks, *rank)
	}

	// Sort by total risk score
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].RiskScore > ranks[j].RiskScore
	})

	// Return top 5
	if len(ranks) > 5 {
		return ranks[:5]
	}
	return ranks
}

// generateComplianceViolations generates compliance violations summary
func generateComplianceViolations(assessments []RiskAssessment) []ComplianceViolation {
	var violations []ComplianceViolation

	// Group by compliance standards
	standardMap := make(map[string]*ComplianceViolation)

	for _, assessment := range assessments {
		if assessment.BusinessRisk.ComplianceImpact != "None" {
			// Map policy to compliance standard
			standard := mapPolicyToStandard(assessment.Policy)
			if violation, exists := standardMap[standard]; exists {
				violation.ViolationCount++
				violation.RiskScore += assessment.CVSSScore
			} else {
				standardMap[standard] = &ComplianceViolation{
					Standard:       standard,
					Requirement:    assessment.Policy,
					ViolationCount: 1,
					RiskScore:      assessment.CVSSScore,
					Severity:       assessment.RiskLevel,
				}
			}
		}
	}

	// Convert to slice
	for _, violation := range standardMap {
		violations = append(violations, *violation)
	}

	return violations
}

// mapPolicyToStandard maps policy to compliance standard
func mapPolicyToStandard(policy string) string {
	switch {
	case strings.Contains(strings.ToLower(policy), "security"):
		return "ISO 27001"
	case strings.Contains(strings.ToLower(policy), "branch protection"):
		return "SOC 2"
	case strings.Contains(strings.ToLower(policy), "access"):
		return "GDPR"
	case strings.Contains(strings.ToLower(policy), "audit"):
		return "SOX"
	default:
		return "General"
	}
}

// outputRiskAssessmentTable outputs risk assessment in table format
func outputRiskAssessmentTable(assessments []RiskAssessment, metrics *BusinessRiskMetrics, showDetails bool) error {
	fmt.Println("🎯 Risk Assessment Report")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Show business metrics if provided
	if metrics != nil {
		fmt.Printf("📊 Organization Risk Overview\n")
		fmt.Printf("Total Risk Score: %.1f\n", metrics.TotalRiskScore)
		fmt.Printf("Average Risk Score: %.1f\n", metrics.AverageRiskScore)
		fmt.Printf("Critical Risks: %d\n", metrics.CriticalRiskCount)
		fmt.Printf("Estimated Cost: $%.0f\n", metrics.EstimatedCost)
		fmt.Printf("Risk Trend: %s\n", metrics.RiskTrend)
		fmt.Printf("Escalation Required: %t\n", metrics.EscalationRequired)
		fmt.Println()

		// Show risk distribution
		fmt.Printf("📈 Risk Distribution\n")
		for level, count := range metrics.RiskDistribution {
			fmt.Printf("  %s: %d\n", level, count)
		}
		fmt.Println()
	}

	// Show individual assessments
	fmt.Printf("📋 Individual Risk Assessments\n")
	fmt.Printf("%-20s %-15s %-20s %-8s %-10s %-10s\n", "Repository", "Policy", "Violation", "Score", "Risk Level", "Priority")
	fmt.Println(strings.Repeat("-", 100))

	for _, assessment := range assessments {
		fmt.Printf("%-20s %-15s %-20s %-8.1f %-10s %-10d\n",
			truncateString(assessment.Repository, 20),
			truncateString(assessment.Policy, 15),
			truncateString(assessment.Violation, 20),
			assessment.CVSSScore,
			assessment.RiskLevel,
			assessment.Priority)

		if showDetails {
			fmt.Printf("  💼 Business Impact: %s | 💰 Cost: $%.0f | ⏱️ Timeline: %s\n",
				assessment.BusinessRisk.BusinessCriticality,
				assessment.BusinessRisk.FinancialImpact,
				assessment.Remediation.EstimatedEffort)
			fmt.Printf("  🔧 Remediation: %s\n", assessment.Remediation.Recommendation)
			if assessment.Escalation.Level != "None" {
				fmt.Printf("  🚨 Escalation: %s (Required by: %s)\n",
					assessment.Escalation.Level,
					assessment.Escalation.RequiredBy.Format("2006-01-02 15:04"))
			}
			fmt.Println()
		}
	}

	return nil
}

// outputRiskAssessmentJSON outputs risk assessment in JSON format
func outputRiskAssessmentJSON(assessments []RiskAssessment, metrics *BusinessRiskMetrics, outputFile string) error {
	output := map[string]interface{}{
		"risk_assessments": assessments,
		"business_metrics": metrics,
		"generated_at":     time.Now(),
	}

	if outputFile != "" {
		return writeJSONToFile(output, outputFile)
	}

	return printJSON(output)
}

// outputRiskAssessmentCSV outputs risk assessment in CSV format
func outputRiskAssessmentCSV(assessments []RiskAssessment, outputFile string) error {
	headers := []string{
		"Repository", "Policy", "Setting", "Violation", "CVSS Score", "Risk Level",
		"Business Criticality", "Financial Impact", "Estimated Effort", "Priority",
		"Escalation Level", "Remediation",
	}

	var rows [][]string
	for _, assessment := range assessments {
		row := []string{
			assessment.Repository,
			assessment.Policy,
			assessment.Setting,
			assessment.Violation,
			strconv.FormatFloat(assessment.CVSSScore, 'f', 1, 64),
			assessment.RiskLevel,
			assessment.BusinessRisk.BusinessCriticality,
			strconv.FormatFloat(assessment.BusinessRisk.FinancialImpact, 'f', 0, 64),
			assessment.Remediation.EstimatedEffort,
			strconv.Itoa(assessment.Priority),
			assessment.Escalation.Level,
			assessment.Remediation.Recommendation,
		}
		rows = append(rows, row)
	}

	return writeCSVToFile(headers, rows, outputFile)
}

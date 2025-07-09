package compliance

import (
	"fmt"
	"math"
	"time"
)

// ComplianceScore represents a calculated compliance score
type ComplianceScore struct {
	TotalScore      float64                `json:"total_score"`              // 0-100 점수
	Grade           Grade                  `json:"grade"`                    // A, B, C, D, F 등급
	WeightedScore   float64                `json:"weighted_score"`           // 가중치 적용 점수
	PolicyScores    map[string]PolicyScore `json:"policy_scores"`            // 정책별 점수
	ScoreBreakdown  ScoreBreakdown         `json:"score_breakdown"`          // 점수 세부사항
	Recommendations []string               `json:"recommendations"`          // 개선 권장사항
	CalculatedAt    time.Time              `json:"calculated_at"`            // 계산 시간
	PreviousScore   *ComplianceScore       `json:"previous_score,omitempty"` // 이전 점수 (변화 추적용)
	ScoreMetrics    ScoreMetrics           `json:"score_metrics"`            // 점수 지표
}

// Grade represents compliance grade levels
type Grade string

const (
	GradeA Grade = "A" // 90-100: 우수
	GradeB Grade = "B" // 80-89: 양호
	GradeC Grade = "C" // 70-79: 보통
	GradeD Grade = "D" // 60-69: 미흡
	GradeF Grade = "F" // 0-59: 불량
)

// PolicyScore represents score for a specific policy
type PolicyScore struct {
	PolicyName        string  `json:"policy_name"`
	Score             float64 `json:"score"`              // 0-100 정책 점수
	Weight            float64 `json:"weight"`             // 정책 가중치 (0-1)
	WeightedScore     float64 `json:"weighted_score"`     // 가중치 적용된 점수
	ViolationCount    int     `json:"violation_count"`    // 위반 개수
	ViolationSeverity string  `json:"violation_severity"` // 위반 심각도
	MaxScore          float64 `json:"max_score"`          // 최대 가능 점수
	Penalty           float64 `json:"penalty"`            // 위반으로 인한 감점
}

// ScoreBreakdown provides detailed score breakdown
type ScoreBreakdown struct {
	BaseScore         float64 `json:"base_score"`          // 기본 점수
	SecurityPenalty   float64 `json:"security_penalty"`    // 보안 위반 감점
	CompliancePenalty float64 `json:"compliance_penalty"`  // 규정 위반 감점
	BestPracticeBonus float64 `json:"best_practice_bonus"` // 모범 사례 보너스
	TrendAdjustment   float64 `json:"trend_adjustment"`    // 트렌드 조정
	FinalScore        float64 `json:"final_score"`         // 최종 점수
}

// ScoreMetrics provides additional scoring metrics
type ScoreMetrics struct {
	SecurityIndex    float64 `json:"security_index"`    // 보안 지수
	ComplianceIndex  float64 `json:"compliance_index"`  // 규정 준수 지수
	MaturityIndex    float64 `json:"maturity_index"`    // 성숙도 지수
	ConsistencyIndex float64 `json:"consistency_index"` // 일관성 지수
	TrendIndex       float64 `json:"trend_index"`       // 트렌드 지수
}

// PolicyWeight defines importance weights for different policies
type PolicyWeight struct {
	PolicyName string  `json:"policy_name"`
	Weight     float64 `json:"weight"`      // 0.0 - 1.0
	MaxPenalty float64 `json:"max_penalty"` // 최대 감점
	Category   string  `json:"category"`    // security, compliance, best-practice
}

// ScoreCalculator calculates compliance scores
type ScoreCalculator struct {
	policyWeights   map[string]PolicyWeight
	gradeThresholds map[Grade]float64
}

// NewScoreCalculator creates a new score calculator with default weights
func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{
		policyWeights: getDefaultPolicyWeights(),
		gradeThresholds: map[Grade]float64{
			GradeA: 90.0,
			GradeB: 80.0,
			GradeC: 70.0,
			GradeD: 60.0,
			GradeF: 0.0,
		},
	}
}

// getDefaultPolicyWeights returns default policy weights
func getDefaultPolicyWeights() map[string]PolicyWeight {
	return map[string]PolicyWeight{
		"Branch Protection": {
			PolicyName: "Branch Protection",
			Weight:     0.25, // 25% 가중치
			MaxPenalty: 30.0,
			Category:   "security",
		},
		"Required Reviews": {
			PolicyName: "Required Reviews",
			Weight:     0.20,
			MaxPenalty: 25.0,
			Category:   "security",
		},
		"Security Scanning": {
			PolicyName: "Security Scanning",
			Weight:     0.20,
			MaxPenalty: 25.0,
			Category:   "security",
		},
		"Vulnerability Alerts": {
			PolicyName: "Vulnerability Alerts",
			Weight:     0.15,
			MaxPenalty: 20.0,
			Category:   "security",
		},
		"Admin Enforcement": {
			PolicyName: "Admin Enforcement",
			Weight:     0.10,
			MaxPenalty: 15.0,
			Category:   "compliance",
		},
		"Required Files": {
			PolicyName: "Required Files",
			Weight:     0.05,
			MaxPenalty: 10.0,
			Category:   "best-practice",
		},
		"Workflow Requirements": {
			PolicyName: "Workflow Requirements",
			Weight:     0.05,
			MaxPenalty: 10.0,
			Category:   "best-practice",
		},
	}
}

// CalculateScore calculates comprehensive compliance score
func (sc *ScoreCalculator) CalculateScore(auditSummary AuditSummary, policyCompliance []PolicyCompliance, previousScore *ComplianceScore) (*ComplianceScore, error) {
	score := &ComplianceScore{
		PolicyScores:    make(map[string]PolicyScore),
		CalculatedAt:    time.Now(),
		PreviousScore:   previousScore,
		Recommendations: []string{},
	}

	// Calculate base score from compliance percentage
	baseScore := auditSummary.CompliancePercentage

	// Calculate policy-specific scores
	var totalWeightedScore float64
	var totalWeight float64

	for _, policy := range policyCompliance {
		policyScore := sc.calculatePolicyScore(policy)
		score.PolicyScores[policy.PolicyName] = policyScore

		totalWeightedScore += policyScore.WeightedScore
		if weight, exists := sc.policyWeights[policy.PolicyName]; exists {
			totalWeight += weight.Weight
		}
	}

	// Calculate weighted score
	if totalWeight > 0 {
		score.WeightedScore = totalWeightedScore / totalWeight * 100
	} else {
		score.WeightedScore = baseScore
	}

	// Calculate score breakdown
	score.ScoreBreakdown = sc.calculateScoreBreakdown(auditSummary, policyCompliance, baseScore)

	// Calculate final score
	score.TotalScore = sc.calculateFinalScore(score.ScoreBreakdown)

	// Determine grade
	score.Grade = sc.calculateGrade(score.TotalScore)

	// Calculate metrics
	score.ScoreMetrics = sc.calculateMetrics(auditSummary, policyCompliance, score.TotalScore)

	// Generate recommendations
	score.Recommendations = sc.generateRecommendations(auditSummary, policyCompliance, score.TotalScore)

	return score, nil
}

// calculatePolicyScore calculates score for a specific policy
func (sc *ScoreCalculator) calculatePolicyScore(policy PolicyCompliance) PolicyScore {
	weight, exists := sc.policyWeights[policy.PolicyName]
	if !exists {
		weight = PolicyWeight{
			PolicyName: policy.PolicyName,
			Weight:     0.01, // 기본 1% 가중치
			MaxPenalty: 5.0,
			Category:   "other",
		}
	}

	// Calculate base score from compliance percentage
	baseScore := policy.CompliancePercentage

	// Calculate penalty based on violations and severity
	penalty := sc.calculatePenalty(policy, weight)

	// Calculate final policy score
	policyScore := math.Max(0, baseScore-penalty)

	// Apply weight
	weightedScore := policyScore * weight.Weight

	return PolicyScore{
		PolicyName:        policy.PolicyName,
		Score:             policyScore,
		Weight:            weight.Weight,
		WeightedScore:     weightedScore,
		ViolationCount:    policy.ViolatingRepos,
		ViolationSeverity: policy.Severity,
		MaxScore:          100.0,
		Penalty:           penalty,
	}
}

// calculatePenalty calculates penalty based on violations
func (sc *ScoreCalculator) calculatePenalty(policy PolicyCompliance, weight PolicyWeight) float64 {
	if policy.ViolatingRepos == 0 {
		return 0.0
	}

	// Base penalty calculation
	violationRate := float64(policy.ViolatingRepos) / float64(policy.CompliantRepos+policy.ViolatingRepos)
	basePenalty := violationRate * weight.MaxPenalty

	// Severity multiplier
	severityMultiplier := getSeverityMultiplier(policy.Severity)

	penalty := basePenalty * severityMultiplier

	// Cap at max penalty
	return math.Min(penalty, weight.MaxPenalty)
}

// getSeverityMultiplier returns penalty multiplier based on severity
func getSeverityMultiplier(severity string) float64 {
	switch severity {
	case "critical":
		return 2.0
	case "high":
		return 1.5
	case "medium":
		return 1.0
	case "low":
		return 0.5
	default:
		return 1.0
	}
}

// calculateScoreBreakdown calculates detailed score breakdown
func (sc *ScoreCalculator) calculateScoreBreakdown(auditSummary AuditSummary, policyCompliance []PolicyCompliance, baseScore float64) ScoreBreakdown {
	breakdown := ScoreBreakdown{
		BaseScore: baseScore,
	}

	// Calculate security penalty
	breakdown.SecurityPenalty = sc.calculateSecurityPenalty(policyCompliance)

	// Calculate compliance penalty
	breakdown.CompliancePenalty = sc.calculateCompliancePenalty(policyCompliance)

	// Calculate best practice bonus
	breakdown.BestPracticeBonus = sc.calculateBestPracticeBonus(policyCompliance)

	// Trend adjustment (나중에 트렌드 데이터와 연계)
	breakdown.TrendAdjustment = 0.0

	// Calculate final score
	breakdown.FinalScore = math.Max(0, math.Min(100,
		breakdown.BaseScore-
			breakdown.SecurityPenalty-
			breakdown.CompliancePenalty+
			breakdown.BestPracticeBonus+
			breakdown.TrendAdjustment))

	return breakdown
}

// calculateSecurityPenalty calculates penalty from security violations
func (sc *ScoreCalculator) calculateSecurityPenalty(policyCompliance []PolicyCompliance) float64 {
	var penalty float64

	for _, policy := range policyCompliance {
		if weight, exists := sc.policyWeights[policy.PolicyName]; exists && weight.Category == "security" {
			if policy.ViolatingRepos > 0 {
				violationRate := float64(policy.ViolatingRepos) / float64(policy.CompliantRepos+policy.ViolatingRepos)
				severityMultiplier := getSeverityMultiplier(policy.Severity)
				penalty += violationRate * weight.MaxPenalty * severityMultiplier * 0.5 // 50% 가중치
			}
		}
	}

	return math.Min(penalty, 40.0) // 최대 40점 감점
}

// calculateCompliancePenalty calculates penalty from compliance violations
func (sc *ScoreCalculator) calculateCompliancePenalty(policyCompliance []PolicyCompliance) float64 {
	var penalty float64

	for _, policy := range policyCompliance {
		if weight, exists := sc.policyWeights[policy.PolicyName]; exists && weight.Category == "compliance" {
			if policy.ViolatingRepos > 0 {
				violationRate := float64(policy.ViolatingRepos) / float64(policy.CompliantRepos+policy.ViolatingRepos)
				penalty += violationRate * weight.MaxPenalty * 0.3 // 30% 가중치
			}
		}
	}

	return math.Min(penalty, 20.0) // 최대 20점 감점
}

// calculateBestPracticeBonus calculates bonus from best practices
func (sc *ScoreCalculator) calculateBestPracticeBonus(policyCompliance []PolicyCompliance) float64 {
	var bonus float64
	bestPracticeCount := 0
	perfectCount := 0

	for _, policy := range policyCompliance {
		if weight, exists := sc.policyWeights[policy.PolicyName]; exists && weight.Category == "best-practice" {
			bestPracticeCount++
			if policy.CompliancePercentage == 100.0 {
				perfectCount++
				bonus += 2.0 // 완벽한 모범 사례당 2점 보너스
			} else if policy.CompliancePercentage >= 90.0 {
				bonus += 1.0 // 우수한 모범 사례당 1점 보너스
			}
		}
	}

	// 모든 모범 사례가 완벽할 때 추가 보너스
	if bestPracticeCount > 0 && perfectCount == bestPracticeCount {
		bonus += 5.0
	}

	return math.Min(bonus, 10.0) // 최대 10점 보너스
}

// calculateFinalScore calculates the final compliance score
func (sc *ScoreCalculator) calculateFinalScore(breakdown ScoreBreakdown) float64 {
	return breakdown.FinalScore
}

// calculateGrade determines grade based on score
func (sc *ScoreCalculator) calculateGrade(score float64) Grade {
	if score >= sc.gradeThresholds[GradeA] {
		return GradeA
	} else if score >= sc.gradeThresholds[GradeB] {
		return GradeB
	} else if score >= sc.gradeThresholds[GradeC] {
		return GradeC
	} else if score >= sc.gradeThresholds[GradeD] {
		return GradeD
	}
	return GradeF
}

// calculateMetrics calculates additional scoring metrics
func (sc *ScoreCalculator) calculateMetrics(auditSummary AuditSummary, policyCompliance []PolicyCompliance, totalScore float64) ScoreMetrics {
	// Security Index: 보안 정책들의 평균 준수율
	securityScore := sc.calculateCategoryScore(policyCompliance, "security")

	// Compliance Index: 규정 준수 정책들의 평균 준수율
	complianceScore := sc.calculateCategoryScore(policyCompliance, "compliance")

	// Maturity Index: 전체적인 성숙도 (총 점수 기반)
	maturityIndex := totalScore

	// Consistency Index: 정책간 점수 편차의 역수
	consistencyIndex := sc.calculateConsistencyIndex(policyCompliance)

	// Trend Index: 나중에 트렌드 데이터와 연계 (현재는 고정값)
	trendIndex := 50.0

	return ScoreMetrics{
		SecurityIndex:    securityScore,
		ComplianceIndex:  complianceScore,
		MaturityIndex:    maturityIndex,
		ConsistencyIndex: consistencyIndex,
		TrendIndex:       trendIndex,
	}
}

// calculateCategoryScore calculates average score for a category
func (sc *ScoreCalculator) calculateCategoryScore(policyCompliance []PolicyCompliance, category string) float64 {
	var totalScore float64
	var count int

	for _, policy := range policyCompliance {
		if weight, exists := sc.policyWeights[policy.PolicyName]; exists && weight.Category == category {
			totalScore += policy.CompliancePercentage
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalScore / float64(count)
}

// calculateConsistencyIndex calculates consistency between policies
func (sc *ScoreCalculator) calculateConsistencyIndex(policyCompliance []PolicyCompliance) float64 {
	if len(policyCompliance) < 2 {
		return 100.0
	}

	// Calculate variance in compliance percentages
	var sum, sumSq float64
	count := float64(len(policyCompliance))

	for _, policy := range policyCompliance {
		sum += policy.CompliancePercentage
		sumSq += policy.CompliancePercentage * policy.CompliancePercentage
	}

	mean := sum / count
	variance := (sumSq / count) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// Convert to consistency index (lower variance = higher consistency)
	consistencyIndex := math.Max(0, 100-stdDev)

	return consistencyIndex
}

// generateRecommendations generates improvement recommendations
func (sc *ScoreCalculator) generateRecommendations(auditSummary AuditSummary, policyCompliance []PolicyCompliance, totalScore float64) []string {
	var recommendations []string

	// Priority-based recommendations
	criticalPolicies := sc.findCriticalViolations(policyCompliance)
	if len(criticalPolicies) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔴 긴급: %d개의 치명적 보안 정책 위반을 즉시 해결하세요", len(criticalPolicies)))
	}

	// Security recommendations
	securityScore := sc.calculateCategoryScore(policyCompliance, "security")
	if securityScore < 80.0 {
		recommendations = append(recommendations,
			"🛡️ 보안 정책 준수율이 낮습니다. 브랜치 보호 및 코드 리뷰 정책을 강화하세요")
	}

	// Best practice recommendations
	bestPracticeScore := sc.calculateCategoryScore(policyCompliance, "best-practice")
	if bestPracticeScore < 90.0 {
		recommendations = append(recommendations,
			"📋 모범 사례 적용률을 높이기 위해 필수 파일 및 워크플로우를 추가하세요")
	}

	// Grade-based recommendations
	grade := sc.calculateGrade(totalScore)
	switch grade {
	case GradeF:
		recommendations = append(recommendations,
			"⚠️ 규정 준수 점수가 매우 낮습니다. 전체적인 정책 검토가 필요합니다")
	case GradeD:
		recommendations = append(recommendations,
			"📈 보안 정책부터 우선적으로 개선하여 점수를 향상시키세요")
	case GradeC:
		recommendations = append(recommendations,
			"✨ 양호한 수준입니다. 추가 보안 강화로 더 높은 등급을 달성하세요")
	}

	return recommendations
}

// findCriticalViolations finds policies with critical violations
func (sc *ScoreCalculator) findCriticalViolations(policyCompliance []PolicyCompliance) []PolicyCompliance {
	var critical []PolicyCompliance

	for _, policy := range policyCompliance {
		if policy.Severity == "critical" && policy.ViolatingRepos > 0 {
			critical = append(critical, policy)
		}
	}

	return critical
}

// AuditSummary represents audit summary (imported from existing code)
type AuditSummary struct {
	TotalRepositories     int     `json:"total_repositories"`
	CompliantRepositories int     `json:"compliant_repositories"`
	CompliancePercentage  float64 `json:"compliance_percentage"`
	TotalViolations       int     `json:"total_violations"`
	CriticalViolations    int     `json:"critical_violations"`
	PolicyCount           int     `json:"policy_count"`
	CompliantCount        int     `json:"compliant_count"`
	NonCompliantCount     int     `json:"non_compliant_count"`
}

// PolicyCompliance represents policy compliance (imported from existing code)
type PolicyCompliance struct {
	PolicyName           string  `json:"policy_name"`
	Description          string  `json:"description"`
	Severity             string  `json:"severity"`
	CompliantRepos       int     `json:"compliant_repos"`
	ViolatingRepos       int     `json:"violating_repos"`
	CompliancePercentage float64 `json:"compliance_percentage"`
}

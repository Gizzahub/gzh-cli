package reposync

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ImpactAnalyzer analyzes the impact of changes across the dependency graph
type ImpactAnalyzer struct {
	logger *zap.Logger
	config *ImpactAnalysisConfig
}

// ImpactAnalysisConfig represents configuration for impact analysis
type ImpactAnalysisConfig struct {
	MaxDepth            int        `json:"max_depth"`             // Maximum depth for impact analysis
	IncludeExternalDeps bool       `json:"include_external_deps"` // Include external dependencies in analysis
	AnalyzeTestImpact   bool       `json:"analyze_test_impact"`   // Analyze impact on test files
	ConsiderWeakDeps    bool       `json:"consider_weak_deps"`    // Consider weak dependencies in analysis
	ImpactThreshold     float64    `json:"impact_threshold"`      // Minimum impact score to report
	ExcludePatterns     []string   `json:"exclude_patterns"`      // Patterns to exclude from analysis
	RiskLevels          RiskLevels `json:"risk_levels"`           // Risk level thresholds
}

// RiskLevels defines thresholds for risk classification
type RiskLevels struct {
	HighRiskThreshold   float64 `json:"high_risk_threshold"`   // >= this score = high risk
	MediumRiskThreshold float64 `json:"medium_risk_threshold"` // >= this score = medium risk
	LowRiskThreshold    float64 `json:"low_risk_threshold"`    // >= this score = low risk
}

// ImpactAnalysisReport represents the complete impact analysis result
type ImpactAnalysisReport struct {
	ChangeSet            *ChangeSet            `json:"change_set"`
	Summary              *ImpactSummary        `json:"summary"`
	AffectedModules      []*AffectedModule     `json:"affected_modules"`
	ImpactPaths          []*ImpactPath         `json:"impact_paths"`
	RiskAssessment       *RiskAssessment       `json:"risk_assessment"`
	TestImpact           *TestImpact           `json:"test_impact,omitempty"`
	PerformanceImpact    *PerformanceImpact    `json:"performance_impact"`
	Recommendations      []string              `json:"recommendations"`
	MitigationStrategies []*MitigationStrategy `json:"mitigation_strategies"`
	GeneratedAt          time.Time             `json:"generated_at"`
}

// ChangeSet represents a set of changes to analyze
type ChangeSet struct {
	ID             string    `json:"id"`
	Description    string    `json:"description"`
	ChangedModules []string  `json:"changed_modules"`
	ChangedFiles   []string  `json:"changed_files"`
	ChangeType     string    `json:"change_type"` // addition, modification, deletion, refactor
	Language       string    `json:"language"`
	Author         string    `json:"author,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	CommitHash     string    `json:"commit_hash,omitempty"`
	PullRequestID  string    `json:"pull_request_id,omitempty"`
}

// ImpactSummary provides high-level impact statistics
type ImpactSummary struct {
	TotalAffectedModules int            `json:"total_affected_modules"`
	TotalImpactPaths     int            `json:"total_impact_paths"`
	MaxImpactDepth       int            `json:"max_impact_depth"`
	HighRiskModules      int            `json:"high_risk_modules"`
	MediumRiskModules    int            `json:"medium_risk_modules"`
	LowRiskModules       int            `json:"low_risk_modules"`
	CrossLanguageImpact  bool           `json:"cross_language_impact"`
	LanguageBreakdown    map[string]int `json:"language_breakdown"`
	ImpactByDepth        map[int]int    `json:"impact_by_depth"`
	EstimatedEffort      string         `json:"estimated_effort"`   // low, medium, high
	OverallRiskLevel     string         `json:"overall_risk_level"` // low, medium, high, critical
}

// AffectedModule represents a module affected by the changes
type AffectedModule struct {
	ModulePath         string             `json:"module_path"`
	Language           string             `json:"language"`
	ImpactScore        float64            `json:"impact_score"` // 0-10 scale
	RiskLevel          string             `json:"risk_level"`   // low, medium, high
	ImpactType         []string           `json:"impact_type"`  // compile, runtime, interface, behavior
	DistanceFromChange int                `json:"distance_from_change"`
	DependencyType     DependencyType     `json:"dependency_type"`
	DependencyStrength DependencyStrength `json:"dependency_strength"`
	ReasonForImpact    string             `json:"reason_for_impact"`
	AffectedFeatures   []string           `json:"affected_features"`
	TestCoverage       float64            `json:"test_coverage"` // 0-1 scale
	LastModified       time.Time          `json:"last_modified"`
	Maintainers        []string           `json:"maintainers,omitempty"`
	Related            []string           `json:"related"` // Related module IDs
}

// ImpactPath represents a path of impact through the dependency graph
type ImpactPath struct {
	ID             string   `json:"id"`
	SourceModule   string   `json:"source_module"`
	TargetModule   string   `json:"target_module"`
	Path           []string `json:"path"`
	PathLength     int      `json:"path_length"`
	TotalWeight    float64  `json:"total_weight"`
	RiskScore      float64  `json:"risk_score"`
	PathType       string   `json:"path_type"`       // direct, transitive, circular
	CriticalEdges  []string `json:"critical_edges"`  // Edges that are critical for this path
	BreakingPoints []string `json:"breaking_points"` // Points where the path could be broken
}

// RiskAssessment provides risk analysis of the changes
type RiskAssessment struct {
	OverallRisk       string            `json:"overall_risk"`
	RiskFactors       []string          `json:"risk_factors"`
	MitigatingFactors []string          `json:"mitigating_factors"`
	HighRiskModules   []string          `json:"high_risk_modules"`
	CriticalPaths     []string          `json:"critical_paths"`
	RiskMetrics       *RiskMetrics      `json:"risk_metrics"`
	ComplianceImpact  *ComplianceImpact `json:"compliance_impact,omitempty"`
	SecurityImpact    *SecurityImpact   `json:"security_impact,omitempty"`
}

// RiskMetrics contains quantitative risk metrics
type RiskMetrics struct {
	ChangeComplexity float64 `json:"change_complexity"`  // Complexity of the change itself
	ImpactScope      float64 `json:"impact_scope"`       // How widely the change affects the system
	TestCoverageRisk float64 `json:"test_coverage_risk"` // Risk due to insufficient test coverage
	DependencyRisk   float64 `json:"dependency_risk"`    // Risk from dependency complexity
	HistoricalRisk   float64 `json:"historical_risk"`    // Risk based on historical failure patterns
	OverallRiskScore float64 `json:"overall_risk_score"` // Combined risk score (0-10)
}

// ComplianceImpact represents impact on compliance requirements
type ComplianceImpact struct {
	AffectedStandards []string `json:"affected_standards"` // e.g., SOX, PCI-DSS, GDPR
	RequiredActions   []string `json:"required_actions"`
	ComplianceRisk    string   `json:"compliance_risk"` // low, medium, high
}

// SecurityImpact represents security implications of changes
type SecurityImpact struct {
	SecurityDomains   []string `json:"security_domains"`   // authentication, authorization, encryption, etc.
	VulnerabilityRisk string   `json:"vulnerability_risk"` // low, medium, high
	RequiredReviews   []string `json:"required_reviews"`   // security review, penetration test, etc.
}

// TestImpact represents impact on testing
type TestImpact struct {
	AffectedTestSuites  []string `json:"affected_test_suites"`
	RequiredTestUpdates []string `json:"required_test_updates"`
	EstimatedTestEffort string   `json:"estimated_test_effort"` // low, medium, high
	TestCoverageGaps    []string `json:"test_coverage_gaps"`
	RecommendedTests    []string `json:"recommended_tests"`
}

// PerformanceImpact represents potential performance implications
type PerformanceImpact struct {
	AffectedComponents    []string `json:"affected_components"`
	PerformanceRisk       string   `json:"performance_risk"` // low, medium, high
	RecommendedBenchmarks []string `json:"recommended_benchmarks"`
	PotentialBottlenecks  []string `json:"potential_bottlenecks"`
}

// MitigationStrategy represents a strategy to mitigate risks
type MitigationStrategy struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Priority        int      `json:"priority"`      // 1-10, higher = more important
	Effort          string   `json:"effort"`        // low, medium, high
	Effectiveness   string   `json:"effectiveness"` // low, medium, high
	ApplicableRisks []string `json:"applicable_risks"`
	Steps           []string `json:"steps"`
	Owner           string   `json:"owner,omitempty"`
	Deadline        string   `json:"deadline,omitempty"`
}

// NewImpactAnalyzer creates a new impact analyzer
func NewImpactAnalyzer(logger *zap.Logger, config *ImpactAnalysisConfig) *ImpactAnalyzer {
	if config == nil {
		config = &ImpactAnalysisConfig{
			MaxDepth:            10,
			IncludeExternalDeps: false,
			AnalyzeTestImpact:   true,
			ConsiderWeakDeps:    true,
			ImpactThreshold:     0.1,
			ExcludePatterns:     []string{"test_*", "*_test.*", "mock_*"},
			RiskLevels: RiskLevels{
				HighRiskThreshold:   7.0,
				MediumRiskThreshold: 4.0,
				LowRiskThreshold:    1.0,
			},
		}
	}

	return &ImpactAnalyzer{
		logger: logger,
		config: config,
	}
}

// AnalyzeImpact performs comprehensive impact analysis for a change set
func (ia *ImpactAnalyzer) AnalyzeImpact(changeSet *ChangeSet, depResult *DependencyResult) (*ImpactAnalysisReport, error) {
	ia.logger.Info("Starting impact analysis",
		zap.String("change_id", changeSet.ID),
		zap.Int("changed_modules", len(changeSet.ChangedModules)))

	// Build dependency graph for traversal
	graph := ia.buildDependencyGraph(depResult)
	reverseGraph := ia.buildReverseDependencyGraph(depResult)

	// Find affected modules
	affectedModules := ia.findAffectedModules(changeSet, graph, reverseGraph, depResult)

	// Trace impact paths
	impactPaths := ia.traceImpactPaths(changeSet, graph, affectedModules)

	// Assess risks
	riskAssessment := ia.assessRisks(changeSet, affectedModules, impactPaths, depResult)

	// Analyze test impact
	var testImpact *TestImpact
	if ia.config.AnalyzeTestImpact {
		testImpact = ia.analyzeTestImpact(changeSet, affectedModules, depResult)
	}

	// Analyze performance impact
	performanceImpact := ia.analyzePerformanceImpact(changeSet, affectedModules, depResult)

	// Generate summary
	summary := ia.generateImpactSummary(affectedModules, impactPaths, riskAssessment)

	// Generate recommendations
	recommendations := ia.generateRecommendations(changeSet, affectedModules, riskAssessment)

	// Generate mitigation strategies
	mitigationStrategies := ia.generateMitigationStrategies(riskAssessment, affectedModules)

	report := &ImpactAnalysisReport{
		ChangeSet:            changeSet,
		Summary:              summary,
		AffectedModules:      affectedModules,
		ImpactPaths:          impactPaths,
		RiskAssessment:       riskAssessment,
		TestImpact:           testImpact,
		PerformanceImpact:    performanceImpact,
		Recommendations:      recommendations,
		MitigationStrategies: mitigationStrategies,
		GeneratedAt:          time.Now(),
	}

	ia.logger.Info("Impact analysis completed",
		zap.Int("affected_modules", len(affectedModules)),
		zap.Int("impact_paths", len(impactPaths)),
		zap.String("overall_risk", riskAssessment.OverallRisk))

	return report, nil
}

// buildDependencyGraph builds a forward dependency graph
func (ia *ImpactAnalyzer) buildDependencyGraph(depResult *DependencyResult) map[string][]*Dependency {
	graph := make(map[string][]*Dependency)

	for _, dep := range depResult.Dependencies {
		// Skip external dependencies if not configured to include them
		if dep.External && !ia.config.IncludeExternalDeps {
			continue
		}

		// Skip weak dependencies if not configured to consider them
		if dep.Strength == DependencyStrengthWeak && !ia.config.ConsiderWeakDeps {
			continue
		}

		graph[dep.From] = append(graph[dep.From], dep)
	}

	return graph
}

// buildReverseDependencyGraph builds a reverse dependency graph (who depends on what)
func (ia *ImpactAnalyzer) buildReverseDependencyGraph(depResult *DependencyResult) map[string][]*Dependency {
	reverseGraph := make(map[string][]*Dependency)

	for _, dep := range depResult.Dependencies {
		if dep.External && !ia.config.IncludeExternalDeps {
			continue
		}

		if dep.Strength == DependencyStrengthWeak && !ia.config.ConsiderWeakDeps {
			continue
		}

		reverseGraph[dep.To] = append(reverseGraph[dep.To], dep)
	}

	return reverseGraph
}

// findAffectedModules finds all modules affected by the changes
func (ia *ImpactAnalyzer) findAffectedModules(changeSet *ChangeSet, graph, reverseGraph map[string][]*Dependency, depResult *DependencyResult) []*AffectedModule {
	var affected []*AffectedModule
	visited := make(map[string]bool)
	distanceMap := make(map[string]int)

	// Initialize with directly changed modules
	for _, module := range changeSet.ChangedModules {
		if ia.shouldExcludeModule(module) {
			continue
		}

		distanceMap[module] = 0
		affectedModule := ia.createAffectedModule(module, 0, changeSet, depResult)
		affected = append(affected, affectedModule)
		visited[module] = true
	}

	// Traverse forward dependencies (modules that depend on changed modules)
	for _, changedModule := range changeSet.ChangedModules {
		ia.traverseDependencies(changedModule, reverseGraph, visited, distanceMap, &affected, changeSet, depResult, "forward")
	}

	// Traverse backward dependencies (modules that changed modules depend on)
	for _, changedModule := range changeSet.ChangedModules {
		ia.traverseDependencies(changedModule, graph, visited, distanceMap, &affected, changeSet, depResult, "backward")
	}

	// Sort by impact score
	sort.Slice(affected, func(i, j int) bool {
		return affected[i].ImpactScore > affected[j].ImpactScore
	})

	return affected
}

// traverseDependencies traverses dependencies using BFS
func (ia *ImpactAnalyzer) traverseDependencies(startModule string, graph map[string][]*Dependency, visited map[string]bool, distanceMap map[string]int, affected *[]*AffectedModule, changeSet *ChangeSet, depResult *DependencyResult, direction string) {
	queue := []string{startModule}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentDistance := distanceMap[current]
		if currentDistance >= ia.config.MaxDepth {
			continue
		}

		for _, dep := range graph[current] {
			target := dep.To
			if direction == "forward" {
				target = dep.From // For reverse graph, the "From" is the dependent
			}

			if ia.shouldExcludeModule(target) {
				continue
			}

			newDistance := currentDistance + 1

			// Check if we should add this module
			if !visited[target] || distanceMap[target] > newDistance {
				if !visited[target] {
					affectedModule := ia.createAffectedModule(target, newDistance, changeSet, depResult)
					affectedModule.DependencyType = dep.Type
					affectedModule.DependencyStrength = dep.Strength
					*affected = append(*affected, affectedModule)
					visited[target] = true
				}

				distanceMap[target] = newDistance

				if newDistance < ia.config.MaxDepth {
					queue = append(queue, target)
				}
			}
		}
	}
}

// createAffectedModule creates an AffectedModule from module information
func (ia *ImpactAnalyzer) createAffectedModule(modulePath string, distance int, changeSet *ChangeSet, depResult *DependencyResult) *AffectedModule {
	module := &AffectedModule{
		ModulePath:         modulePath,
		DistanceFromChange: distance,
		TestCoverage:       0.8, // Default assumption
	}

	// Get module information from dependency result
	if moduleInfo, exists := depResult.Modules[modulePath]; exists {
		module.Language = moduleInfo.Language
		module.LastModified = time.Now() // Would need actual file info
	}

	// Calculate impact score based on distance and other factors
	module.ImpactScore = ia.calculateImpactScore(module, changeSet)
	module.RiskLevel = ia.determineRiskLevel(module.ImpactScore)
	module.ImpactType = ia.determineImpactType(module, changeSet)
	module.ReasonForImpact = ia.generateReasonForImpact(module, changeSet)
	module.AffectedFeatures = ia.identifyAffectedFeatures(module, changeSet)

	return module
}

// calculateImpactScore calculates the impact score for a module
func (ia *ImpactAnalyzer) calculateImpactScore(module *AffectedModule, changeSet *ChangeSet) float64 {
	baseScore := 10.0 // Start with maximum impact

	// Reduce score based on distance
	distancePenalty := float64(module.DistanceFromChange) * 1.5
	baseScore -= distancePenalty

	// Adjust based on dependency strength
	switch module.DependencyStrength {
	case DependencyStrengthStrong:
		baseScore *= 1.0
	case DependencyStrengthWeak:
		baseScore *= 0.7
	case DependencyStrengthOptional:
		baseScore *= 0.4
	}

	// Adjust based on change type
	switch changeSet.ChangeType {
	case "deletion":
		baseScore *= 1.5 // Deletions have higher impact
	case "addition":
		baseScore *= 0.8 // Additions have lower impact
	case "modification":
		baseScore *= 1.0 // Normal impact
	case "refactor":
		baseScore *= 0.9 // Refactors slightly lower impact
	}

	// Adjust based on test coverage
	if module.TestCoverage < 0.5 {
		baseScore *= 1.3 // Poor test coverage increases risk
	} else if module.TestCoverage > 0.8 {
		baseScore *= 0.9 // Good test coverage reduces risk
	}

	// Ensure score is within bounds
	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 10 {
		baseScore = 10
	}

	return baseScore
}

// shouldExcludeModule checks if a module should be excluded from analysis
func (ia *ImpactAnalyzer) shouldExcludeModule(modulePath string) bool {
	for _, pattern := range ia.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, modulePath); matched {
			return true
		}
	}
	return false
}

// determineRiskLevel determines the risk level based on impact score
func (ia *ImpactAnalyzer) determineRiskLevel(impactScore float64) string {
	thresholds := ia.config.RiskLevels

	if impactScore >= thresholds.HighRiskThreshold {
		return "high"
	}
	if impactScore >= thresholds.MediumRiskThreshold {
		return "medium"
	}
	if impactScore >= thresholds.LowRiskThreshold {
		return "low"
	}

	return "minimal"
}

// Additional methods would continue here...
// (The file is getting quite long, so I'll implement the remaining methods in separate files)

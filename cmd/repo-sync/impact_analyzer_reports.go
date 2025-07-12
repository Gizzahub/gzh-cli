package reposync

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Report generation methods for ImpactAnalyzer

// analyzeTestImpact analyzes the impact on testing
func (ia *ImpactAnalyzer) analyzeTestImpact(changeSet *ChangeSet, affectedModules []*AffectedModule, depResult *DependencyResult) *TestImpact {
	testImpact := &TestImpact{
		AffectedTestSuites:  []string{},
		RequiredTestUpdates: []string{},
		TestCoverageGaps:    []string{},
		RecommendedTests:    []string{},
	}

	// Identify affected test suites
	for _, module := range affectedModules {
		// Infer test suite names from module paths
		testSuite := ia.inferTestSuiteName(module.ModulePath, module.Language)
		if testSuite != "" {
			testImpact.AffectedTestSuites = append(testImpact.AffectedTestSuites, testSuite)
		}

		// Identify modules with poor test coverage
		if module.TestCoverage < 0.7 {
			testImpact.TestCoverageGaps = append(testImpact.TestCoverageGaps, module.ModulePath)
		}

		// Generate test recommendations based on impact type
		for _, impactType := range module.ImpactType {
			recommendations := ia.generateTestRecommendations(module, impactType)
			testImpact.RecommendedTests = append(testImpact.RecommendedTests, recommendations...)
		}
	}

	// Estimate test effort
	testImpact.EstimatedTestEffort = ia.estimateTestEffort(affectedModules, changeSet)

	// Remove duplicates
	testImpact.AffectedTestSuites = ia.removeDuplicates(testImpact.AffectedTestSuites)
	testImpact.RecommendedTests = ia.removeDuplicates(testImpact.RecommendedTests)
	testImpact.TestCoverageGaps = ia.removeDuplicates(testImpact.TestCoverageGaps)

	return testImpact
}

// inferTestSuiteName infers test suite names from module paths
func (ia *ImpactAnalyzer) inferTestSuiteName(modulePath, language string) string {
	switch language {
	case "go":
		// Go test conventions
		if strings.Contains(modulePath, "/") {
			parts := strings.Split(modulePath, "/")
			return fmt.Sprintf("%s_test", parts[len(parts)-1])
		}
		return fmt.Sprintf("%s_test", modulePath)
	case "javascript", "typescript":
		// Jest/Mocha conventions
		return fmt.Sprintf("%s.test.js", strings.ReplaceAll(modulePath, "/", "."))
	case "python":
		// pytest conventions
		return fmt.Sprintf("test_%s.py", strings.ReplaceAll(modulePath, "/", "_"))
	case "java":
		// JUnit conventions
		return fmt.Sprintf("%sTest.java", strings.Title(strings.ReplaceAll(modulePath, "/", "")))
	default:
		return fmt.Sprintf("%s_test", modulePath)
	}
}

// generateTestRecommendations generates test recommendations based on impact type
func (ia *ImpactAnalyzer) generateTestRecommendations(module *AffectedModule, impactType string) []string {
	var recommendations []string

	switch impactType {
	case "compile":
		recommendations = append(recommendations, fmt.Sprintf("Run compilation tests for %s", module.ModulePath))
	case "runtime":
		recommendations = append(recommendations, fmt.Sprintf("Run integration tests for %s", module.ModulePath))
	case "interface":
		recommendations = append(recommendations, fmt.Sprintf("Run contract tests for %s", module.ModulePath))
	case "behavior":
		recommendations = append(recommendations, fmt.Sprintf("Run behavioral tests for %s", module.ModulePath))
	case "breaking":
		recommendations = append(recommendations, fmt.Sprintf("Run full test suite for %s - breaking changes detected", module.ModulePath))
	}

	return recommendations
}

// estimateTestEffort estimates the effort required for testing
func (ia *ImpactAnalyzer) estimateTestEffort(affectedModules []*AffectedModule, changeSet *ChangeSet) string {
	totalModules := len(affectedModules)
	highRiskModules := 0

	for _, module := range affectedModules {
		if module.RiskLevel == "high" {
			highRiskModules++
		}
	}

	// Factor in change type
	effortMultiplier := 1.0
	switch changeSet.ChangeType {
	case "deletion":
		effortMultiplier = 1.5
	case "modification":
		effortMultiplier = 1.0
	case "addition":
		effortMultiplier = 0.8
	case "refactor":
		effortMultiplier = 1.2
	}

	adjustedModules := float64(totalModules) * effortMultiplier

	switch {
	case adjustedModules > 20 || highRiskModules > 5:
		return "high"
	case adjustedModules > 10 || highRiskModules > 2:
		return "medium"
	default:
		return "low"
	}
}

// analyzePerformanceImpact analyzes potential performance implications
func (ia *ImpactAnalyzer) analyzePerformanceImpact(changeSet *ChangeSet, affectedModules []*AffectedModule, depResult *DependencyResult) *PerformanceImpact {
	impact := &PerformanceImpact{
		AffectedComponents:    []string{},
		RecommendedBenchmarks: []string{},
		PotentialBottlenecks:  []string{},
	}

	performanceKeywords := []string{"cache", "database", "network", "algorithm", "loop", "query", "index", "memory"}

	for _, module := range affectedModules {
		// Check if module path suggests performance-critical components
		for _, keyword := range performanceKeywords {
			if strings.Contains(strings.ToLower(module.ModulePath), keyword) {
				impact.AffectedComponents = append(impact.AffectedComponents, module.ModulePath)
				impact.RecommendedBenchmarks = append(impact.RecommendedBenchmarks,
					fmt.Sprintf("Benchmark %s performance", module.ModulePath))
				break
			}
		}

		// High-risk modules in core paths are potential bottlenecks
		if module.RiskLevel == "high" && module.DistanceFromChange <= 2 {
			impact.PotentialBottlenecks = append(impact.PotentialBottlenecks, module.ModulePath)
		}
	}

	// Determine performance risk
	if len(impact.AffectedComponents) > 5 || len(impact.PotentialBottlenecks) > 3 {
		impact.PerformanceRisk = "high"
	} else if len(impact.AffectedComponents) > 2 || len(impact.PotentialBottlenecks) > 1 {
		impact.PerformanceRisk = "medium"
	} else {
		impact.PerformanceRisk = "low"
	}

	// Remove duplicates
	impact.AffectedComponents = ia.removeDuplicates(impact.AffectedComponents)
	impact.RecommendedBenchmarks = ia.removeDuplicates(impact.RecommendedBenchmarks)
	impact.PotentialBottlenecks = ia.removeDuplicates(impact.PotentialBottlenecks)

	return impact
}

// generateImpactSummary generates a summary of the impact analysis
func (ia *ImpactAnalyzer) generateImpactSummary(affectedModules []*AffectedModule, impactPaths []*ImpactPath, riskAssessment *RiskAssessment) *ImpactSummary {
	summary := &ImpactSummary{
		TotalAffectedModules: len(affectedModules),
		TotalImpactPaths:     len(impactPaths),
		LanguageBreakdown:    make(map[string]int),
		ImpactByDepth:        make(map[int]int),
	}

	// Count by risk level
	for _, module := range affectedModules {
		switch module.RiskLevel {
		case "high":
			summary.HighRiskModules++
		case "medium":
			summary.MediumRiskModules++
		case "low":
			summary.LowRiskModules++
		}

		// Count by language
		summary.LanguageBreakdown[module.Language]++

		// Count by depth
		summary.ImpactByDepth[module.DistanceFromChange]++

		// Track max depth
		if module.DistanceFromChange > summary.MaxImpactDepth {
			summary.MaxImpactDepth = module.DistanceFromChange
		}
	}

	// Check for cross-language impact
	summary.CrossLanguageImpact = len(summary.LanguageBreakdown) > 1

	// Estimate effort
	summary.EstimatedEffort = ia.estimateOverallEffort(affectedModules, riskAssessment)

	// Set overall risk level
	summary.OverallRiskLevel = riskAssessment.OverallRisk

	return summary
}

// estimateOverallEffort estimates the overall effort required
func (ia *ImpactAnalyzer) estimateOverallEffort(affectedModules []*AffectedModule, riskAssessment *RiskAssessment) string {
	effortScore := 0.0

	// Base effort from number of affected modules
	effortScore += float64(len(affectedModules)) * 0.5

	// Additional effort from high-risk modules
	for _, module := range affectedModules {
		switch module.RiskLevel {
		case "high":
			effortScore += 3.0
		case "medium":
			effortScore += 1.5
		case "low":
			effortScore += 0.5
		}
	}

	// Factor in overall risk
	switch riskAssessment.OverallRisk {
	case "critical":
		effortScore *= 2.0
	case "high":
		effortScore *= 1.5
	case "medium":
		effortScore *= 1.2
	}

	// Determine effort level
	switch {
	case effortScore > 50:
		return "high"
	case effortScore > 20:
		return "medium"
	default:
		return "low"
	}
}

// generateRecommendations generates recommendations based on the analysis
func (ia *ImpactAnalyzer) generateRecommendations(changeSet *ChangeSet, affectedModules []*AffectedModule, riskAssessment *RiskAssessment) []string {
	var recommendations []string

	// Overall recommendations based on risk level
	switch riskAssessment.OverallRisk {
	case "critical":
		recommendations = append(recommendations, "🚨 Critical risk detected - Consider breaking changes into smaller parts")
		recommendations = append(recommendations, "Require senior engineer review before deployment")
		recommendations = append(recommendations, "Implement comprehensive rollback plan")
	case "high":
		recommendations = append(recommendations, "⚠️ High risk - Thorough testing and staged rollout recommended")
		recommendations = append(recommendations, "Consider feature flags for safer deployment")
	case "medium":
		recommendations = append(recommendations, "Moderate risk - Standard testing and review processes apply")
	case "low":
		recommendations = append(recommendations, "✅ Low risk - Standard deployment procedures sufficient")
	}

	// Specific recommendations based on affected modules
	if len(affectedModules) > 20 {
		recommendations = append(recommendations, "Large impact scope - Consider phased deployment")
	}

	// Test-related recommendations
	poorCoverageCount := 0
	for _, module := range affectedModules {
		if module.TestCoverage < 0.5 {
			poorCoverageCount++
		}
	}

	if poorCoverageCount > 3 {
		recommendations = append(recommendations,
			fmt.Sprintf("🧪 %d modules have poor test coverage - Prioritize test improvements", poorCoverageCount))
	}

	// Language-specific recommendations
	languages := make(map[string]bool)
	for _, module := range affectedModules {
		languages[module.Language] = true
	}

	if len(languages) > 1 {
		recommendations = append(recommendations, "Cross-language impact detected - Ensure integration testing")
	}

	// Change type specific recommendations
	switch changeSet.ChangeType {
	case "deletion":
		recommendations = append(recommendations, "Deletion changes - Verify no breaking dependencies remain")
	case "refactor":
		recommendations = append(recommendations, "Refactoring changes - Ensure behavior equivalence through testing")
	}

	return recommendations
}

// generateMitigationStrategies generates strategies to mitigate identified risks
func (ia *ImpactAnalyzer) generateMitigationStrategies(riskAssessment *RiskAssessment, affectedModules []*AffectedModule) []*MitigationStrategy {
	var strategies []*MitigationStrategy
	strategyID := 1

	// Strategy for high-risk modules
	if len(riskAssessment.HighRiskModules) > 0 {
		strategies = append(strategies, &MitigationStrategy{
			ID:              fmt.Sprintf("strategy_%d", strategyID),
			Name:            "High-Risk Module Monitoring",
			Description:     "Implement enhanced monitoring for high-risk modules",
			Priority:        9,
			Effort:          "medium",
			Effectiveness:   "high",
			ApplicableRisks: []string{"high_risk_modules"},
			Steps: []string{
				"Set up detailed logging for high-risk modules",
				"Implement health checks and alerts",
				"Create rollback procedures",
				"Schedule post-deployment monitoring",
			},
		})
		strategyID++
	}

	// Strategy for poor test coverage
	poorCoverageModules := []string{}
	for _, module := range affectedModules {
		if module.TestCoverage < 0.5 {
			poorCoverageModules = append(poorCoverageModules, module.ModulePath)
		}
	}

	if len(poorCoverageModules) > 0 {
		strategies = append(strategies, &MitigationStrategy{
			ID:              fmt.Sprintf("strategy_%d", strategyID),
			Name:            "Test Coverage Improvement",
			Description:     "Improve test coverage for affected modules",
			Priority:        7,
			Effort:          "high",
			Effectiveness:   "high",
			ApplicableRisks: []string{"poor_test_coverage"},
			Steps: []string{
				"Identify critical paths in affected modules",
				"Write unit tests for core functionality",
				"Add integration tests for module interactions",
				"Implement mutation testing to verify test quality",
			},
		})
		strategyID++
	}

	// Strategy for critical paths
	if len(riskAssessment.CriticalPaths) > 0 {
		strategies = append(strategies, &MitigationStrategy{
			ID:              fmt.Sprintf("strategy_%d", strategyID),
			Name:            "Critical Path Protection",
			Description:     "Protect critical dependency paths from failures",
			Priority:        8,
			Effort:          "medium",
			Effectiveness:   "high",
			ApplicableRisks: []string{"critical_paths"},
			Steps: []string{
				"Implement circuit breakers for critical paths",
				"Add redundancy where possible",
				"Create fallback mechanisms",
				"Monitor path health continuously",
			},
		})
		strategyID++
	}

	// Strategy for cross-language risks
	languages := make(map[string]bool)
	for _, module := range affectedModules {
		languages[module.Language] = true
	}

	if len(languages) > 1 {
		strategies = append(strategies, &MitigationStrategy{
			ID:              fmt.Sprintf("strategy_%d", strategyID),
			Name:            "Cross-Language Integration Testing",
			Description:     "Ensure proper integration across different programming languages",
			Priority:        6,
			Effort:          "medium",
			Effectiveness:   "medium",
			ApplicableRisks: []string{"cross_language_complexity"},
			Steps: []string{
				"Set up end-to-end test environments",
				"Test API contracts between languages",
				"Verify data serialization/deserialization",
				"Monitor inter-service communication",
			},
		})
		strategyID++
	}

	// Strategy for overall risk mitigation
	if riskAssessment.OverallRisk == "critical" || riskAssessment.OverallRisk == "high" {
		strategies = append(strategies, &MitigationStrategy{
			ID:              fmt.Sprintf("strategy_%d", strategyID),
			Name:            "Staged Deployment",
			Description:     "Deploy changes in stages to minimize risk",
			Priority:        10,
			Effort:          "low",
			Effectiveness:   "high",
			ApplicableRisks: []string{"overall_high_risk"},
			Steps: []string{
				"Deploy to development environment first",
				"Run comprehensive test suite",
				"Deploy to staging with production-like data",
				"Monitor for 24 hours before production deployment",
				"Deploy to production during low-traffic periods",
			},
		})
	}

	// Sort strategies by priority
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].Priority > strategies[j].Priority
	})

	return strategies
}

// removeDuplicates removes duplicate strings from a slice
func (ia *ImpactAnalyzer) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// CreateChangeSetFromGitDiff creates a ChangeSet from git diff information
func (ia *ImpactAnalyzer) CreateChangeSetFromGitDiff(commitHash, author string, changedFiles []string) *ChangeSet {
	changeSet := &ChangeSet{
		ID:             fmt.Sprintf("changeset_%s", commitHash[:8]),
		CommitHash:     commitHash,
		Author:         author,
		ChangedFiles:   changedFiles,
		ChangedModules: ia.extractModulesFromFiles(changedFiles),
		Timestamp:      time.Now(),
		ChangeType:     ia.inferChangeType(changedFiles),
		Language:       ia.inferPrimaryLanguage(changedFiles),
		Description:    fmt.Sprintf("Change set from commit %s", commitHash[:8]),
	}

	return changeSet
}

// extractModulesFromFiles extracts module names from file paths
func (ia *ImpactAnalyzer) extractModulesFromFiles(files []string) []string {
	moduleSet := make(map[string]bool)

	for _, file := range files {
		// Extract module path from file path
		dir := filepath.Dir(file)

		// Normalize module path (remove common prefixes)
		module := strings.TrimPrefix(dir, "./")
		module = strings.TrimPrefix(module, "src/")
		module = strings.TrimPrefix(module, "lib/")

		if module != "" && module != "." {
			moduleSet[module] = true
		}
	}

	modules := make([]string, 0, len(moduleSet))
	for module := range moduleSet {
		modules = append(modules, module)
	}

	return modules
}

// inferChangeType infers the type of change from file list
func (ia *ImpactAnalyzer) inferChangeType(files []string) string {
	// This is a simplified implementation
	// In practice, you would analyze git status to determine if files were added, modified, or deleted

	if len(files) > 10 {
		return "refactor"
	}

	// Check for common patterns
	for _, file := range files {
		if strings.Contains(file, "test") {
			continue // Skip test files for change type detection
		}

		if strings.Contains(file, "new") || strings.Contains(file, "add") {
			return "addition"
		}
	}

	return "modification"
}

// inferPrimaryLanguage infers the primary language from file extensions
func (ia *ImpactAnalyzer) inferPrimaryLanguage(files []string) string {
	langCount := make(map[string]int)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".go":
			langCount["go"]++
		case ".js", ".jsx":
			langCount["javascript"]++
		case ".ts", ".tsx":
			langCount["typescript"]++
		case ".py":
			langCount["python"]++
		case ".java":
			langCount["java"]++
		case ".rb":
			langCount["ruby"]++
		case ".php":
			langCount["php"]++
		case ".cs":
			langCount["csharp"]++
		}
	}

	// Find the language with the most files
	maxCount := 0
	primaryLang := "unknown"
	for lang, count := range langCount {
		if count > maxCount {
			maxCount = count
			primaryLang = lang
		}
	}

	return primaryLang
}

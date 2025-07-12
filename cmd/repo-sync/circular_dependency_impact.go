package reposync

import (
	"fmt"
	"sort"
	"strings"
)

// Impact analysis methods for CircularDependencyDetector

// analyzeImpact performs comprehensive impact analysis
func (cdd *CircularDependencyDetector) analyzeImpact(cycles []*EnhancedCycle, graph map[string][]*CycleEdge, result *DependencyResult) *ImpactAnalysis {
	analysis := &ImpactAnalysis{
		LanguageImpact: make(map[string]*LanguageImpact),
	}

	// Find most affected nodes
	analysis.MostAffectedNodes = cdd.findMostAffectedNodes(cycles)

	// Identify critical paths
	analysis.CriticalPaths = cdd.identifyCriticalPaths(cycles, graph)

	// Analyze language-specific impact
	analysis.LanguageImpact = cdd.analyzeLanguageImpact(cycles, result)

	// Calculate system metrics
	analysis.SystemComplexity = cdd.calculateSystemComplexity(cycles, result)
	analysis.TestabilityScore = cdd.calculateTestabilityScore(cycles, result)
	analysis.MaintainabilityScore = cdd.calculateMaintainabilityScore(cycles, result)

	return analysis
}

// findMostAffectedNodes finds nodes that appear in multiple cycles
func (cdd *CircularDependencyDetector) findMostAffectedNodes(cycles []*EnhancedCycle) []string {
	nodeCount := make(map[string]int)

	// Count how many cycles each node appears in
	for _, cycle := range cycles {
		for _, node := range cycle.Cycle {
			nodeCount[node]++
		}
	}

	// Convert to sorted slice
	type nodeScore struct {
		node  string
		count int
	}

	var scores []nodeScore
	for node, count := range nodeCount {
		if count > 1 { // Only nodes in multiple cycles
			scores = append(scores, nodeScore{node, count})
		}
	}

	// Sort by count (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].count > scores[j].count
	})

	// Extract top nodes
	result := make([]string, 0, len(scores))
	for i, score := range scores {
		if i >= 10 { // Limit to top 10
			break
		}
		result = append(result, score.node)
	}

	return result
}

// identifyCriticalPaths identifies paths that are part of multiple cycles
func (cdd *CircularDependencyDetector) identifyCriticalPaths(cycles []*EnhancedCycle, graph map[string][]*CycleEdge) [][]string {
	var criticalPaths [][]string

	// Find paths that appear in multiple critical cycles
	pathCount := make(map[string]int)
	pathCycles := make(map[string][]*EnhancedCycle)

	for _, cycle := range cycles {
		if cycle.Severity == "critical" || cycle.Severity == "high" {
			// Extract all paths of length 2 from the cycle
			for i := 0; i < len(cycle.Cycle)-1; i++ {
				for j := i + 1; j < len(cycle.Cycle)-1 && j-i <= 3; j++ {
					path := cycle.Cycle[i:j+1]
					pathKey := cdd.pathKey(path)
					pathCount[pathKey]++
					pathCycles[pathKey] = append(pathCycles[pathKey], cycle)
				}
			}
		}
	}

	// Select paths that appear in multiple cycles
	for pathKey, count := range pathCount {
		if count > 1 {
			path := cdd.pathFromKey(pathKey)
			criticalPaths = append(criticalPaths, path)
		}
	}

	// Sort by frequency
	sort.Slice(criticalPaths, func(i, j int) bool {
		keyI := cdd.pathKey(criticalPaths[i])
		keyJ := cdd.pathKey(criticalPaths[j])
		return pathCount[keyI] > pathCount[keyJ]
	})

	// Limit to top 10 critical paths
	if len(criticalPaths) > 10 {
		criticalPaths = criticalPaths[:10]
	}

	return criticalPaths
}

// pathKey generates a key for a path
func (cdd *CircularDependencyDetector) pathKey(path []string) string {
	return strings.Join(path, "->")
}

// pathFromKey recreates a path from a key
func (cdd *CircularDependencyDetector) pathFromKey(key string) []string {
	return strings.Split(key, "->")
}

// analyzeLanguageImpact analyzes impact by programming language
func (cdd *CircularDependencyDetector) analyzeLanguageImpact(cycles []*EnhancedCycle, result *DependencyResult) map[string]*LanguageImpact {
	languageImpact := make(map[string]*LanguageImpact)

	// Initialize for all languages in the project
	for _, dep := range result.Dependencies {
		if _, exists := languageImpact[dep.Language]; !exists {
			languageImpact[dep.Language] = &LanguageImpact{
				Language:        dep.Language,
				CycleCount:      0,
				AffectedModules: 0,
				ComplexityScore: 0.0,
				Recommendations: []string{},
			}
		}
	}

	// Count cycles and affected modules per language
	modulesByLanguage := make(map[string]map[string]bool)
	for lang := range languageImpact {
		modulesByLanguage[lang] = make(map[string]bool)
	}

	for _, cycle := range cycles {
		for _, lang := range cycle.Languages {
			if impact, exists := languageImpact[lang]; exists {
				impact.CycleCount++
				impact.ComplexityScore += cycle.Metrics.Complexity

				// Count affected modules
				for _, node := range cycle.Cycle {
					modulesByLanguage[lang][node] = true
				}
			}
		}
	}

	// Calculate final metrics and recommendations
	for lang, impact := range languageImpact {
		impact.AffectedModules = len(modulesByLanguage[lang])
		
		if impact.CycleCount > 0 {
			impact.ComplexityScore /= float64(impact.CycleCount) // Average complexity
		}

		impact.Recommendations = cdd.generateLanguageRecommendations(lang, impact)
	}

	return languageImpact
}

// generateLanguageRecommendations generates language-specific recommendations
func (cdd *CircularDependencyDetector) generateLanguageRecommendations(language string, impact *LanguageImpact) []string {
	var recommendations []string

	switch language {
	case "go":
		if impact.CycleCount > 3 {
			recommendations = append(recommendations, "Use Go interfaces to break dependencies")
			recommendations = append(recommendations, "Consider using dependency injection with wire or similar tools")
		}
		if impact.ComplexityScore > 5.0 {
			recommendations = append(recommendations, "Break packages into smaller, focused modules")
		}

	case "javascript", "typescript":
		if impact.CycleCount > 2 {
			recommendations = append(recommendations, "Use dependency injection containers like inversify")
			recommendations = append(recommendations, "Implement barrel exports to control dependencies")
		}
		if impact.ComplexityScore > 4.0 {
			recommendations = append(recommendations, "Consider using micro-frontends architecture")
		}

	case "python":
		if impact.CycleCount > 2 {
			recommendations = append(recommendations, "Use dependency injection with dependency-injector")
			recommendations = append(recommendations, "Implement abstract base classes to define interfaces")
		}
		if impact.ComplexityScore > 4.0 {
			recommendations = append(recommendations, "Break large modules into smaller packages")
		}

	case "java":
		if impact.CycleCount > 3 {
			recommendations = append(recommendations, "Use Spring's dependency injection")
			recommendations = append(recommendations, "Apply SOLID principles more strictly")
		}

	default:
		recommendations = append(recommendations, "Apply language-specific dependency inversion patterns")
	}

	// General recommendations based on impact
	if impact.CycleCount > 5 {
		recommendations = append(recommendations, "Consider architectural refactoring")
	}

	if impact.AffectedModules > 10 {
		recommendations = append(recommendations, "Implement module boundaries more clearly")
	}

	return recommendations
}

// calculateSystemComplexity calculates overall system complexity
func (cdd *CircularDependencyDetector) calculateSystemComplexity(cycles []*EnhancedCycle, result *DependencyResult) float64 {
	if len(result.Modules) == 0 {
		return 0.0
	}

	// Base complexity from cycles
	cycleComplexity := 0.0
	for _, cycle := range cycles {
		weight := 1.0
		switch cycle.Severity {
		case "critical":
			weight = 4.0
		case "high":
			weight = 3.0
		case "medium":
			weight = 2.0
		case "low":
			weight = 1.0
		}
		cycleComplexity += cycle.Metrics.Complexity * weight
	}

	// Normalize by total modules
	normalizedComplexity := cycleComplexity / float64(len(result.Modules))

	// Scale to 0-10 range
	if normalizedComplexity > 10.0 {
		normalizedComplexity = 10.0
	}

	return normalizedComplexity
}

// calculateTestabilityScore calculates how testable the system is
func (cdd *CircularDependencyDetector) calculateTestabilityScore(cycles []*EnhancedCycle, result *DependencyResult) float64 {
	if len(result.Modules) == 0 {
		return 10.0 // Perfect score for empty system
	}

	// Start with perfect score and deduct for issues
	score := 10.0

	// Deduct points for circular dependencies
	for _, cycle := range cycles {
		penalty := 0.5
		switch cycle.Severity {
		case "critical":
			penalty = 2.0
		case "high":
			penalty = 1.5
		case "medium":
			penalty = 1.0
		case "low":
			penalty = 0.5
		}
		score -= penalty
	}

	// Additional penalty for cross-language cycles (harder to test)
	for _, cycle := range cycles {
		if len(cycle.Languages) > 1 {
			score -= 0.3
		}
	}

	// Ensure score is within bounds
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// calculateMaintainabilityScore calculates how maintainable the system is
func (cdd *CircularDependencyDetector) calculateMaintainabilityScore(cycles []*EnhancedCycle, result *DependencyResult) float64 {
	if len(result.Modules) == 0 {
		return 10.0
	}

	// Start with perfect score
	score := 10.0

	// Deduct for circular dependencies
	for _, cycle := range cycles {
		penalty := 0.3
		switch cycle.Severity {
		case "critical":
			penalty = 1.5
		case "high":
			penalty = 1.0
		case "medium":
			penalty = 0.7
		case "low":
			penalty = 0.3
		}
		score -= penalty
	}

	// Additional penalties
	for _, cycle := range cycles {
		// Long cycles are harder to maintain
		if cycle.Length > 5 {
			score -= 0.2
		}

		// Cross-language cycles add complexity
		if len(cycle.Languages) > 1 {
			score -= 0.2
		}

		// Strong coupling reduces maintainability
		if cycle.Metrics.StrongEdges > cycle.Metrics.WeakEdges {
			score -= 0.1
		}
	}

	if score < 0.0 {
		score = 0.0
	}

	return score
}

// generateSummary generates a summary of the analysis
func (cdd *CircularDependencyDetector) generateSummary(cycles []*EnhancedCycle) *CircularSummary {
	summary := &CircularSummary{
		TotalCycles:          len(cycles),
		LanguageBreakdown:    make(map[string]int),
		SeverityDistribution: make(map[string]int),
	}

	nodeSet := make(map[string]bool)
	totalLength := 0

	for _, cycle := range cycles {
		// Count by severity
		summary.SeverityDistribution[cycle.Severity]++

		// Count by language
		for _, lang := range cycle.Languages {
			summary.LanguageBreakdown[lang]++
		}

		// Track unique nodes
		for _, node := range cycle.Cycle {
			nodeSet[node] = true
		}

		// Track lengths
		totalLength += cycle.Length
		if cycle.Length > summary.MaxCycleLength {
			summary.MaxCycleLength = cycle.Length
		}

		// Count severe cycles
		switch cycle.Severity {
		case "critical":
			summary.CriticalCycles++
		case "high":
			summary.HighSeverityCycles++
		}
	}

	summary.TotalNodes = len(nodeSet)
	summary.AffectedNodes = len(nodeSet) // All tracked nodes are affected

	if len(cycles) > 0 {
		summary.AverageCycleLength = float64(totalLength) / float64(len(cycles))
	}

	return summary
}

// generateRecommendations generates high-level recommendations
func (cdd *CircularDependencyDetector) generateRecommendations(cycles []*EnhancedCycle, impact *ImpactAnalysis) []string {
	var recommendations []string

	// Recommendations based on number of cycles
	switch {
	case len(cycles) == 0:
		recommendations = append(recommendations, "✅ No circular dependencies detected - excellent architecture!")
	case len(cycles) <= 3:
		recommendations = append(recommendations, "Minor circular dependencies detected - address when convenient")
	case len(cycles) <= 10:
		recommendations = append(recommendations, "⚠️ Moderate number of circular dependencies - plan refactoring")
	default:
		recommendations = append(recommendations, "🚨 High number of circular dependencies - immediate attention required")
	}

	// Critical cycle recommendations
	criticalCount := 0
	for _, cycle := range cycles {
		if cycle.Severity == "critical" {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔥 %d critical cycles require immediate resolution", criticalCount))
	}

	// Complexity recommendations
	if impact.SystemComplexity > 7.0 {
		recommendations = append(recommendations,
			"High system complexity detected - consider architectural restructuring")
	}

	// Testability recommendations
	if impact.TestabilityScore < 6.0 {
		recommendations = append(recommendations,
			"Low testability score - circular dependencies are making testing difficult")
	}

	// Maintainability recommendations
	if impact.MaintainabilityScore < 6.0 {
		recommendations = append(recommendations,
			"Low maintainability score - refactoring recommended to improve code quality")
	}

	// Language-specific recommendations
	for lang, langImpact := range impact.LanguageImpact {
		if langImpact.CycleCount > 3 {
			recommendations = append(recommendations,
				fmt.Sprintf("High number of cycles in %s - review module structure", lang))
		}
	}

	// Multi-language recommendations
	crossLangCycles := 0
	for _, cycle := range cycles {
		if len(cycle.Languages) > 1 {
			crossLangCycles++
		}
	}

	if crossLangCycles > 0 {
		recommendations = append(recommendations,
			"Cross-language cycles detected - consider API boundaries")
	}

	return recommendations
}

// groupCyclesByLanguage groups cycles by programming language
func (cdd *CircularDependencyDetector) groupCyclesByLanguage(cycles []*EnhancedCycle) map[string][]*EnhancedCycle {
	groups := make(map[string][]*EnhancedCycle)

	for _, cycle := range cycles {
		for _, lang := range cycle.Languages {
			groups[lang] = append(groups[lang], cycle)
		}
	}

	return groups
}

// groupCyclesByLength groups cycles by their length
func (cdd *CircularDependencyDetector) groupCyclesByLength(cycles []*EnhancedCycle) map[int][]*EnhancedCycle {
	groups := make(map[int][]*EnhancedCycle)

	for _, cycle := range cycles {
		groups[cycle.Length] = append(groups[cycle.Length], cycle)
	}

	return groups
}

// groupCyclesBySeverity groups cycles by severity level
func (cdd *CircularDependencyDetector) groupCyclesBySeverity(cycles []*EnhancedCycle) map[string][]*EnhancedCycle {
	groups := make(map[string][]*EnhancedCycle)

	for _, cycle := range cycles {
		groups[cycle.Severity] = append(groups[cycle.Severity], cycle)
	}

	return groups
}
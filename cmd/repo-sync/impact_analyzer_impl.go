package reposync

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Additional implementation methods for ImpactAnalyzer

// determineImpactType determines the type of impact for a module
func (ia *ImpactAnalyzer) determineImpactType(module *AffectedModule, changeSet *ChangeSet) []string {
	var impactTypes []string

	// Determine impact type based on distance and dependency type
	switch module.DistanceFromChange {
	case 0:
		impactTypes = append(impactTypes, "direct")
	case 1:
		impactTypes = append(impactTypes, "immediate")
	default:
		impactTypes = append(impactTypes, "transitive")
	}

	// Impact type based on dependency type
	switch module.DependencyType {
	case DependencyTypeImport:
		impactTypes = append(impactTypes, "compile")
	case DependencyTypeRequire:
		impactTypes = append(impactTypes, "runtime")
	case DependencyTypeInherit:
		impactTypes = append(impactTypes, "interface")
	case DependencyTypeCall:
		impactTypes = append(impactTypes, "behavior")
	}

	// Impact type based on change type
	switch changeSet.ChangeType {
	case "deletion":
		impactTypes = append(impactTypes, "breaking")
	case "addition":
		impactTypes = append(impactTypes, "additive")
	case "modification":
		impactTypes = append(impactTypes, "behavioral")
	case "refactor":
		impactTypes = append(impactTypes, "structural")
	}

	return impactTypes
}

// generateReasonForImpact generates a human-readable reason for impact
func (ia *ImpactAnalyzer) generateReasonForImpact(module *AffectedModule, changeSet *ChangeSet) string {
	if module.DistanceFromChange == 0 {
		return "Module is directly modified"
	}

	var reason strings.Builder
	
	if module.DistanceFromChange == 1 {
		reason.WriteString("Module directly depends on modified code")
	} else {
		reason.WriteString(fmt.Sprintf("Module transitively affected through %d-degree dependency", module.DistanceFromChange))
	}

	// Add dependency strength context
	switch module.DependencyStrength {
	case DependencyStrengthStrong:
		reason.WriteString(" (strong coupling)")
	case DependencyStrengthWeak:
		reason.WriteString(" (weak coupling)")
	case DependencyStrengthOptional:
		reason.WriteString(" (optional dependency)")
	}

	// Add change type context
	switch changeSet.ChangeType {
	case "deletion":
		reason.WriteString(" - may break functionality")
	case "modification":
		reason.WriteString(" - may change behavior")
	case "addition":
		reason.WriteString(" - may affect interfaces")
	}

	return reason.String()
}

// identifyAffectedFeatures identifies features that might be affected
func (ia *ImpactAnalyzer) identifyAffectedFeatures(module *AffectedModule, changeSet *ChangeSet) []string {
	var features []string

	// Extract features from module path
	pathParts := strings.Split(module.ModulePath, "/")
	for _, part := range pathParts {
		if strings.Contains(part, "service") || strings.Contains(part, "handler") || 
		   strings.Contains(part, "controller") || strings.Contains(part, "api") {
			features = append(features, strings.Title(part))
		}
	}

	// Infer features from language-specific patterns
	switch module.Language {
	case "go":
		if strings.Contains(module.ModulePath, "cmd/") {
			features = append(features, "CLI Commands")
		}
		if strings.Contains(module.ModulePath, "pkg/") {
			features = append(features, "Core Libraries")
		}
		if strings.Contains(module.ModulePath, "internal/") {
			features = append(features, "Internal APIs")
		}
	case "javascript", "typescript":
		if strings.Contains(module.ModulePath, "components") {
			features = append(features, "UI Components")
		}
		if strings.Contains(module.ModulePath, "services") {
			features = append(features, "Business Services")
		}
		if strings.Contains(module.ModulePath, "utils") {
			features = append(features, "Utility Functions")
		}
	case "python":
		if strings.Contains(module.ModulePath, "models") {
			features = append(features, "Data Models")
		}
		if strings.Contains(module.ModulePath, "views") {
			features = append(features, "View Layer")
		}
		if strings.Contains(module.ModulePath, "serializers") {
			features = append(features, "Data Serialization")
		}
	}

	// Default feature if none found
	if len(features) == 0 {
		features = append(features, "Core Functionality")
	}

	return features
}

// traceImpactPaths traces paths of impact through the dependency graph
func (ia *ImpactAnalyzer) traceImpactPaths(changeSet *ChangeSet, graph map[string][]*Dependency, affectedModules []*AffectedModule) []*ImpactPath {
	var paths []*ImpactPath
	pathID := 1

	// Create map of affected modules for quick lookup
	affectedMap := make(map[string]*AffectedModule)
	for _, module := range affectedModules {
		affectedMap[module.ModulePath] = module
	}

	// Trace paths from each changed module to affected modules
	for _, sourceModule := range changeSet.ChangedModules {
		for _, affectedModule := range affectedModules {
			if affectedModule.ModulePath == sourceModule || affectedModule.DistanceFromChange == 0 {
				continue
			}

			path := ia.findPath(sourceModule, affectedModule.ModulePath, graph)
			if path != nil && len(path) > 1 {
				impactPath := &ImpactPath{
					ID:           fmt.Sprintf("path_%d", pathID),
					SourceModule: sourceModule,
					TargetModule: affectedModule.ModulePath,
					Path:         path,
					PathLength:   len(path) - 1,
					PathType:     ia.determinePathType(path, affectedMap),
				}

				impactPath.TotalWeight = ia.calculatePathWeight(path, graph)
				impactPath.RiskScore = ia.calculatePathRiskScore(impactPath, affectedMap)
				impactPath.CriticalEdges = ia.identifyCriticalEdges(path, graph)
				impactPath.BreakingPoints = ia.identifyBreakingPoints(path, graph)

				paths = append(paths, impactPath)
				pathID++
			}
		}
	}

	// Sort paths by risk score
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].RiskScore > paths[j].RiskScore
	})

	return paths
}

// findPath finds a path between two modules using BFS
func (ia *ImpactAnalyzer) findPath(source, target string, graph map[string][]*Dependency) []string {
	if source == target {
		return []string{source}
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{source}
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dep := range graph[current] {
			next := dep.To
			if !visited[next] {
				visited[next] = true
				parent[next] = current
				queue = append(queue, next)

				if next == target {
					// Reconstruct path
					path := []string{target}
					for p := parent[target]; p != ""; p = parent[p] {
						path = append([]string{p}, path...)
						if p == source {
							break
						}
					}
					return path
				}
			}
		}
	}

	return nil // No path found
}

// determinePathType determines the type of an impact path
func (ia *ImpactAnalyzer) determinePathType(path []string, affectedMap map[string]*AffectedModule) string {
	if len(path) == 2 {
		return "direct"
	}

	// Check if it's a circular path
	nodeSet := make(map[string]bool)
	for _, node := range path {
		if nodeSet[node] {
			return "circular"
		}
		nodeSet[node] = true
	}

	return "transitive"
}

// calculatePathWeight calculates the total weight of a path
func (ia *ImpactAnalyzer) calculatePathWeight(path []string, graph map[string][]*Dependency) float64 {
	totalWeight := 0.0

	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]

		// Find the dependency between these nodes
		for _, dep := range graph[from] {
			if dep.To == to {
				weight := 1.0
				switch dep.Strength {
				case DependencyStrengthStrong:
					weight = 3.0
				case DependencyStrengthWeak:
					weight = 1.0
				case DependencyStrengthOptional:
					weight = 0.3
				}
				totalWeight += weight
				break
			}
		}
	}

	return totalWeight
}

// calculatePathRiskScore calculates the risk score for a path
func (ia *ImpactAnalyzer) calculatePathRiskScore(path *ImpactPath, affectedMap map[string]*AffectedModule) float64 {
	baseScore := path.TotalWeight

	// Adjust based on path length
	lengthPenalty := float64(path.PathLength) * 0.5
	baseScore -= lengthPenalty

	// Adjust based on target module risk
	if targetModule, exists := affectedMap[path.TargetModule]; exists {
		baseScore += targetModule.ImpactScore * 0.3
	}

	// Circular paths are riskier
	if path.PathType == "circular" {
		baseScore *= 1.5
	}

	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 10 {
		baseScore = 10
	}

	return baseScore
}

// identifyCriticalEdges identifies critical edges in a path
func (ia *ImpactAnalyzer) identifyCriticalEdges(path []string, graph map[string][]*Dependency) []string {
	var criticalEdges []string

	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]

		for _, dep := range graph[from] {
			if dep.To == to {
				// Edge is critical if it's a strong dependency or inheritance
				if dep.Strength == DependencyStrengthStrong || dep.Type == DependencyTypeInherit {
					criticalEdges = append(criticalEdges, fmt.Sprintf("%s -> %s", from, to))
				}
				break
			}
		}
	}

	return criticalEdges
}

// identifyBreakingPoints identifies potential breaking points in a path
func (ia *ImpactAnalyzer) identifyBreakingPoints(path []string, graph map[string][]*Dependency) []string {
	var breakingPoints []string

	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]

		for _, dep := range graph[from] {
			if dep.To == to {
				// Weak or optional dependencies are potential breaking points
				if dep.Strength == DependencyStrengthWeak || dep.Strength == DependencyStrengthOptional {
					breakingPoints = append(breakingPoints, fmt.Sprintf("%s -> %s", from, to))
				}
				break
			}
		}
	}

	return breakingPoints
}

// assessRisks performs comprehensive risk assessment
func (ia *ImpactAnalyzer) assessRisks(changeSet *ChangeSet, affectedModules []*AffectedModule, impactPaths []*ImpactPath, depResult *DependencyResult) *RiskAssessment {
	assessment := &RiskAssessment{
		RiskFactors:       []string{},
		MitigatingFactors: []string{},
		HighRiskModules:   []string{},
		CriticalPaths:     []string{},
		RiskMetrics:       &RiskMetrics{},
	}

	// Count risk levels
	highRiskCount := 0
	for _, module := range affectedModules {
		if module.RiskLevel == "high" {
			highRiskCount++
			assessment.HighRiskModules = append(assessment.HighRiskModules, module.ModulePath)
		}
	}

	// Identify critical paths
	for _, path := range impactPaths {
		if path.RiskScore >= ia.config.RiskLevels.HighRiskThreshold {
			assessment.CriticalPaths = append(assessment.CriticalPaths, path.ID)
		}
	}

	// Calculate risk metrics
	assessment.RiskMetrics = ia.calculateRiskMetrics(changeSet, affectedModules, impactPaths)

	// Determine overall risk
	assessment.OverallRisk = ia.determineOverallRisk(assessment.RiskMetrics)

	// Identify risk factors
	assessment.RiskFactors = ia.identifyRiskFactors(changeSet, affectedModules, impactPaths)

	// Identify mitigating factors
	assessment.MitigatingFactors = ia.identifyMitigatingFactors(changeSet, affectedModules)

	return assessment
}

// calculateRiskMetrics calculates detailed risk metrics
func (ia *ImpactAnalyzer) calculateRiskMetrics(changeSet *ChangeSet, affectedModules []*AffectedModule, impactPaths []*ImpactPath) *RiskMetrics {
	metrics := &RiskMetrics{}

	// Change complexity based on number of changed modules
	metrics.ChangeComplexity = float64(len(changeSet.ChangedModules)) * 1.5
	if metrics.ChangeComplexity > 10 {
		metrics.ChangeComplexity = 10
	}

	// Impact scope based on number of affected modules
	metrics.ImpactScope = float64(len(affectedModules)) * 0.5
	if metrics.ImpactScope > 10 {
		metrics.ImpactScope = 10
	}

	// Test coverage risk
	totalCoverage := 0.0
	for _, module := range affectedModules {
		totalCoverage += module.TestCoverage
	}
	if len(affectedModules) > 0 {
		avgCoverage := totalCoverage / float64(len(affectedModules))
		metrics.TestCoverageRisk = (1.0 - avgCoverage) * 10
	}

	// Dependency risk based on path complexity
	maxPathLength := 0
	totalPathWeight := 0.0
	for _, path := range impactPaths {
		if path.PathLength > maxPathLength {
			maxPathLength = path.PathLength
		}
		totalPathWeight += path.TotalWeight
	}
	metrics.DependencyRisk = float64(maxPathLength) * 1.0
	if len(impactPaths) > 0 {
		metrics.DependencyRisk += (totalPathWeight / float64(len(impactPaths))) * 0.5
	}

	// Historical risk (simplified - would need actual historical data)
	switch changeSet.ChangeType {
	case "deletion":
		metrics.HistoricalRisk = 8.0
	case "modification":
		metrics.HistoricalRisk = 5.0
	case "addition":
		metrics.HistoricalRisk = 3.0
	case "refactor":
		metrics.HistoricalRisk = 4.0
	default:
		metrics.HistoricalRisk = 5.0
	}

	// Overall risk score (weighted average)
	metrics.OverallRiskScore = (metrics.ChangeComplexity*0.2 +
		metrics.ImpactScope*0.3 +
		metrics.TestCoverageRisk*0.2 +
		metrics.DependencyRisk*0.2 +
		metrics.HistoricalRisk*0.1)

	return metrics
}

// determineOverallRisk determines the overall risk level
func (ia *ImpactAnalyzer) determineOverallRisk(metrics *RiskMetrics) string {
	score := metrics.OverallRiskScore

	if score >= 8.0 {
		return "critical"
	}
	if score >= 6.0 {
		return "high"
	}
	if score >= 3.0 {
		return "medium"
	}
	return "low"
}

// identifyRiskFactors identifies factors that increase risk
func (ia *ImpactAnalyzer) identifyRiskFactors(changeSet *ChangeSet, affectedModules []*AffectedModule, impactPaths []*ImpactPath) []string {
	var factors []string

	// Change-related factors
	if len(changeSet.ChangedModules) > 5 {
		factors = append(factors, "Large number of changed modules")
	}

	if changeSet.ChangeType == "deletion" {
		factors = append(factors, "Deletion changes have higher risk of breaking dependencies")
	}

	// Module-related factors
	highRiskCount := 0
	poorCoverageCount := 0
	for _, module := range affectedModules {
		if module.RiskLevel == "high" {
			highRiskCount++
		}
		if module.TestCoverage < 0.5 {
			poorCoverageCount++
		}
	}

	if highRiskCount > 3 {
		factors = append(factors, fmt.Sprintf("%d modules at high risk", highRiskCount))
	}

	if poorCoverageCount > 2 {
		factors = append(factors, fmt.Sprintf("%d modules with poor test coverage", poorCoverageCount))
	}

	// Path-related factors
	complexPaths := 0
	for _, path := range impactPaths {
		if path.PathLength > 4 {
			complexPaths++
		}
	}

	if complexPaths > 0 {
		factors = append(factors, fmt.Sprintf("%d complex dependency paths", complexPaths))
	}

	// Cross-language factors
	languages := make(map[string]bool)
	for _, module := range affectedModules {
		languages[module.Language] = true
	}

	if len(languages) > 1 {
		factors = append(factors, "Cross-language dependencies increase complexity")
	}

	return factors
}

// identifyMitigatingFactors identifies factors that reduce risk
func (ia *ImpactAnalyzer) identifyMitigatingFactors(changeSet *ChangeSet, affectedModules []*AffectedModule) []string {
	var factors []string

	// Test coverage factors
	goodCoverageCount := 0
	for _, module := range affectedModules {
		if module.TestCoverage > 0.8 {
			goodCoverageCount++
		}
	}

	if goodCoverageCount > 0 {
		factors = append(factors, fmt.Sprintf("%d modules have good test coverage", goodCoverageCount))
	}

	// Change type factors
	if changeSet.ChangeType == "addition" {
		factors = append(factors, "Addition changes are generally safer")
	}

	// Recent modifications
	recentlyModified := 0
	for _, module := range affectedModules {
		if time.Since(module.LastModified) < 30*24*time.Hour { // 30 days
			recentlyModified++
		}
	}

	if recentlyModified > len(affectedModules)/2 {
		factors = append(factors, "Many affected modules were recently modified (code is fresh)")
	}

	return factors
}
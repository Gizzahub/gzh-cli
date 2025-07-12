package reposync

import (
	"fmt"
	"sort"
	"strings"
)

// Additional methods for CircularDependencyDetector

// determineCycleType determines the type of a cycle
func (cdd *CircularDependencyDetector) determineCycleType(cycle *EnhancedCycle) string {
	if len(cycle.Languages) > 1 {
		return "cross-language"
	}

	if cycle.Length == 2 {
		return "direct"
	}

	// Check for indirect cycles (more than 2 nodes)
	if cycle.Length > 2 {
		return "indirect"
	}

	return "unknown"
}

// generateCycleDescription generates a human-readable description
func (cdd *CircularDependencyDetector) generateCycleDescription(cycle *EnhancedCycle) string {
	if len(cycle.Cycle) < 2 {
		return "Invalid cycle"
	}

	cycleStr := strings.Join(cycle.Cycle[:len(cycle.Cycle)-1], " → ")
	cycleStr += " → " + cycle.Cycle[0] // Close the cycle

	return fmt.Sprintf("%s cycle of length %d: %s",
		strings.Title(cycle.CycleType), cycle.Length, cycleStr)
}

// assessCycleImpact assesses the impact of a cycle
func (cdd *CircularDependencyDetector) assessCycleImpact(cycle *EnhancedCycle) string {
	switch cycle.Severity {
	case "critical":
		return "Very High - May cause build failures, runtime errors, or deadlocks"
	case "high":
		return "High - Significantly increases complexity and reduces maintainability"
	case "medium":
		return "Medium - Makes code harder to understand and test"
	case "low":
		return "Low - Minor impact on code organization"
	default:
		return "Unknown impact"
	}
}

// generateCycleSuggestions generates suggestions for resolving a cycle
func (cdd *CircularDependencyDetector) generateCycleSuggestions(cycle *EnhancedCycle) []string {
	var suggestions []string

	// Generic suggestions based on cycle type
	switch cycle.CycleType {
	case "direct":
		suggestions = append(suggestions, []string{
			"Extract common functionality into a separate module",
			"Use dependency injection to invert the dependency",
			"Consider merging the modules if they are tightly coupled",
			"Introduce an interface to break the direct dependency",
		}...)
	case "indirect":
		suggestions = append(suggestions, []string{
			"Analyze the dependency chain to find the weakest link",
			"Extract shared functionality into a common module",
			"Use events or messaging to decouple modules",
			"Apply the mediator pattern to coordinate interactions",
		}...)
	case "cross-language":
		suggestions = append(suggestions, []string{
			"Use language-specific dependency injection frameworks",
			"Implement service interfaces for cross-language communication",
			"Consider API-based communication instead of direct dependencies",
			"Use shared libraries or common interfaces",
		}...)
	}

	// Suggestions based on severity
	if cycle.Severity == "critical" || cycle.Severity == "high" {
		suggestions = append(suggestions, "Priority: Resolve immediately to prevent system issues")
	}

	// Suggestions based on complexity
	if cycle.Metrics.CrossLanguageEdges > 0 {
		suggestions = append(suggestions, "Consider standardizing on fewer programming languages")
	}

	if cycle.Metrics.StrongEdges > cycle.Metrics.WeakEdges {
		suggestions = append(suggestions, "Look for opportunities to make some dependencies optional or weak")
	}

	return suggestions
}

// identifyBreakingPoints identifies potential points to break the cycle
func (cdd *CircularDependencyDetector) identifyBreakingPoints(cycle *EnhancedCycle, graph map[string][]*CycleEdge) []*BreakingPoint {
	var points []*BreakingPoint

	for _, edge := range cycle.Edges {
		confidence := cdd.calculateBreakingConfidence(edge, cycle)
		impact := cdd.assessBreakingImpact(edge, cycle)
		strategy := cdd.suggestBreakingStrategy(edge, cycle)

		point := &BreakingPoint{
			FromNode:    edge.From,
			ToNode:      edge.To,
			Confidence:  confidence,
			Impact:      impact,
			Strategy:    strategy,
			Description: cdd.generateBreakingDescription(edge, strategy),
			Effort:      cdd.estimateBreakingEffort(edge, strategy),
			Rationale:   cdd.generateBreakingRationale(edge, cycle),
		}

		points = append(points, point)
	}

	// Sort by confidence (highest first)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Confidence > points[j].Confidence
	})

	return points
}

// calculateBreakingConfidence calculates confidence for breaking an edge
func (cdd *CircularDependencyDetector) calculateBreakingConfidence(edge *CycleEdge, cycle *EnhancedCycle) float64 {
	confidence := 0.5 // Base confidence

	// Weak dependencies are easier to break
	switch edge.Strength {
	case DependencyStrengthOptional:
		confidence += 0.4
	case DependencyStrengthWeak:
		confidence += 0.2
	case DependencyStrengthStrong:
		confidence -= 0.1
	}

	// Some dependency types are easier to break
	switch edge.Type {
	case DependencyTypeInclude:
		confidence += 0.2
	case DependencyTypeCall:
		confidence += 0.1
	case DependencyTypeInherit:
		confidence -= 0.2
	}

	// Cross-language dependencies might be easier to break
	if len(cycle.Languages) > 1 {
		confidence += 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// assessBreakingImpact assesses the impact of breaking an edge
func (cdd *CircularDependencyDetector) assessBreakingImpact(edge *CycleEdge, cycle *EnhancedCycle) string {
	if edge.Strength == DependencyStrengthOptional {
		return "low"
	}

	if edge.Strength == DependencyStrengthWeak {
		return "medium"
	}

	if cycle.Length <= 3 {
		return "high"
	}

	return "medium"
}

// suggestBreakingStrategy suggests a strategy for breaking an edge
func (cdd *CircularDependencyDetector) suggestBreakingStrategy(edge *CycleEdge, cycle *EnhancedCycle) string {
	switch edge.Type {
	case DependencyTypeInherit:
		return "interface"
	case DependencyTypeRequire, DependencyTypeImport:
		if edge.Strength == DependencyStrengthWeak {
			return "injection"
		}
		return "extract"
	case DependencyTypeInclude:
		return "merge"
	case DependencyTypeCall:
		return "event"
	default:
		return "refactor"
	}
}

// generateBreakingDescription generates a description for a breaking point
func (cdd *CircularDependencyDetector) generateBreakingDescription(edge *CycleEdge, strategy string) string {
	switch strategy {
	case "interface":
		return fmt.Sprintf("Introduce an interface between %s and %s", edge.From, edge.To)
	case "injection":
		return fmt.Sprintf("Use dependency injection to inject %s into %s", edge.To, edge.From)
	case "extract":
		return fmt.Sprintf("Extract common functionality from %s and %s", edge.From, edge.To)
	case "merge":
		return fmt.Sprintf("Consider merging %s and %s if closely related", edge.From, edge.To)
	case "event":
		return fmt.Sprintf("Replace direct call with event-based communication")
	default:
		return fmt.Sprintf("Refactor the dependency between %s and %s", edge.From, edge.To)
	}
}

// estimateBreakingEffort estimates the effort required to break an edge
func (cdd *CircularDependencyDetector) estimateBreakingEffort(edge *CycleEdge, strategy string) string {
	switch strategy {
	case "injection":
		return "medium"
	case "interface":
		return "medium"
	case "extract":
		return "high"
	case "merge":
		return "low"
	case "event":
		return "high"
	default:
		return "medium"
	}
}

// generateBreakingRationale generates rationale for breaking an edge
func (cdd *CircularDependencyDetector) generateBreakingRationale(edge *CycleEdge, cycle *EnhancedCycle) string {
	reasons := []string{}

	if edge.Strength == DependencyStrengthWeak || edge.Strength == DependencyStrengthOptional {
		reasons = append(reasons, "weak dependency makes breaking easier")
	}

	if edge.Type == DependencyTypeInclude {
		reasons = append(reasons, "include dependencies are often refactorable")
	}

	if len(cycle.Languages) > 1 {
		reasons = append(reasons, "cross-language boundary provides natural breaking point")
	}

	if len(reasons) == 0 {
		return "Breaking this dependency would help resolve the circular dependency"
	}

	return strings.Join(reasons, "; ")
}

// findRelatedCycles finds cycles that share nodes with the given cycle
func (cdd *CircularDependencyDetector) findRelatedCycles(cycle *EnhancedCycle, allCycles []*EnhancedCycle) []string {
	var related []string
	cycleNodes := make(map[string]bool)

	for _, node := range cycle.Cycle {
		cycleNodes[node] = true
	}

	for _, other := range allCycles {
		if other.ID == cycle.ID {
			continue
		}

		// Check if cycles share any nodes
		for _, node := range other.Cycle {
			if cycleNodes[node] {
				related = append(related, other.ID)
				break
			}
		}
	}

	return related
}

// filterCycles filters cycles based on configuration
func (cdd *CircularDependencyDetector) filterCycles(cycles []*EnhancedCycle) []*EnhancedCycle {
	var filtered []*EnhancedCycle

	for _, cycle := range cycles {
		// Filter by severity
		if !cdd.shouldIncludeSeverity(cycle.Severity) {
			continue
		}

		// Filter by length
		if cycle.Length > cdd.config.MaxCycleLength {
			continue
		}

		filtered = append(filtered, cycle)
	}

	return filtered
}

// shouldIncludeSeverity checks if a severity level should be included
func (cdd *CircularDependencyDetector) shouldIncludeSeverity(severity string) bool {
	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	minLevel := severityLevels[cdd.config.MinSeverityLevel]
	currentLevel := severityLevels[severity]

	return currentLevel >= minLevel
}

// generateBreakingStrategies generates comprehensive breaking strategies
func (cdd *CircularDependencyDetector) generateBreakingStrategies(cycles []*EnhancedCycle, graph map[string][]*CycleEdge) []*BreakingStrategy {
	var strategies []*BreakingStrategy

	// Strategy 1: Interface Segregation
	if cdd.hasInheritanceCycles(cycles) {
		strategies = append(strategies, &BreakingStrategy{
			ID:          "interface_segregation",
			Name:        "Interface Segregation",
			Description: "Break inheritance cycles by introducing interfaces",
			Priority:    8,
			Effort:      "medium",
			Impact:      "high",
			Steps: []string{
				"Identify common behaviors in the cycle",
				"Extract these behaviors into interfaces",
				"Have classes implement interfaces instead of inheriting",
				"Use composition over inheritance where possible",
			},
			Examples: []string{
				"Replace class inheritance with interface implementation",
				"Use dependency injection with interfaces",
			},
			Risks: []string{
				"May require significant refactoring",
				"Could increase code complexity initially",
			},
		})
	}

	// Strategy 2: Dependency Injection
	if cdd.hasStrongCouplingCycles(cycles) {
		strategies = append(strategies, &BreakingStrategy{
			ID:          "dependency_injection",
			Name:        "Dependency Injection",
			Description: "Use dependency injection to invert dependencies",
			Priority:    7,
			Effort:      "medium",
			Impact:      "high",
			Steps: []string{
				"Identify the dependency direction to invert",
				"Create interfaces for the dependencies",
				"Inject dependencies through constructors or setters",
				"Use a dependency injection container if needed",
			},
			Examples: []string{
				"Inject service dependencies through constructors",
				"Use factory patterns for complex dependencies",
			},
			Risks: []string{
				"May increase setup complexity",
				"Requires team understanding of DI patterns",
			},
		})
	}

	// Strategy 3: Extract Common Module
	strategies = append(strategies, &BreakingStrategy{
		ID:          "extract_common",
		Name:        "Extract Common Module",
		Description: "Extract shared functionality into a separate module",
		Priority:    6,
		Effort:      "high",
		Impact:      "medium",
		Steps: []string{
			"Identify common functionality in the cycle",
			"Create a new module for shared code",
			"Move common code to the new module",
			"Update imports to use the new module",
		},
		Examples: []string{
			"Extract utility functions to a common package",
			"Move shared data structures to a separate module",
		},
		Risks: []string{
			"May create a god module if not done carefully",
			"Could introduce new dependencies",
		},
	})

	// Strategy 4: Event-Driven Architecture
	if cdd.hasCallCycles(cycles) {
		strategies = append(strategies, &BreakingStrategy{
			ID:          "event_driven",
			Name:        "Event-Driven Architecture",
			Description: "Replace direct calls with event-based communication",
			Priority:    5,
			Effort:      "high",
			Impact:      "high",
			Steps: []string{
				"Identify call patterns in the cycle",
				"Design appropriate events for these interactions",
				"Implement event publishing and subscription",
				"Replace direct calls with event publishing",
			},
			Examples: []string{
				"Use observer pattern for notifications",
				"Implement message queues for async communication",
			},
			Risks: []string{
				"Increases system complexity",
				"May impact performance due to async nature",
			},
		})
	}

	// Set applicable cycles for each strategy
	for _, strategy := range strategies {
		strategy.ApplicableTo = cdd.findApplicableCycles(strategy, cycles)
	}

	// Sort by priority
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].Priority > strategies[j].Priority
	})

	return strategies
}

// hasInheritanceCycles checks if there are inheritance-based cycles
func (cdd *CircularDependencyDetector) hasInheritanceCycles(cycles []*EnhancedCycle) bool {
	for _, cycle := range cycles {
		for _, edge := range cycle.Edges {
			if edge.Type == DependencyTypeInherit {
				return true
			}
		}
	}
	return false
}

// hasStrongCouplingCycles checks if there are strongly coupled cycles
func (cdd *CircularDependencyDetector) hasStrongCouplingCycles(cycles []*EnhancedCycle) bool {
	for _, cycle := range cycles {
		if cycle.Metrics.StrongEdges >= 2 {
			return true
		}
	}
	return false
}

// hasCallCycles checks if there are call-based cycles
func (cdd *CircularDependencyDetector) hasCallCycles(cycles []*EnhancedCycle) bool {
	for _, cycle := range cycles {
		for _, edge := range cycle.Edges {
			if edge.Type == DependencyTypeCall {
				return true
			}
		}
	}
	return false
}

// findApplicableCycles finds cycles that a strategy applies to
func (cdd *CircularDependencyDetector) findApplicableCycles(strategy *BreakingStrategy, cycles []*EnhancedCycle) []string {
	var applicable []string

	for _, cycle := range cycles {
		switch strategy.ID {
		case "interface_segregation":
			if cdd.hasInheritanceEdges(cycle) {
				applicable = append(applicable, cycle.ID)
			}
		case "dependency_injection":
			if cycle.Metrics.StrongEdges > 0 {
				applicable = append(applicable, cycle.ID)
			}
		case "extract_common":
			applicable = append(applicable, cycle.ID) // Always applicable
		case "event_driven":
			if cdd.hasCallEdges(cycle) {
				applicable = append(applicable, cycle.ID)
			}
		}
	}

	return applicable
}

// hasInheritanceEdges checks if a cycle has inheritance edges
func (cdd *CircularDependencyDetector) hasInheritanceEdges(cycle *EnhancedCycle) bool {
	for _, edge := range cycle.Edges {
		if edge.Type == DependencyTypeInherit {
			return true
		}
	}
	return false
}

// hasCallEdges checks if a cycle has call edges
func (cdd *CircularDependencyDetector) hasCallEdges(cycle *EnhancedCycle) bool {
	for _, edge := range cycle.Edges {
		if edge.Type == DependencyTypeCall {
			return true
		}
	}
	return false
}

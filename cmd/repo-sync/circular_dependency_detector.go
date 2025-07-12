package reposync

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CircularDependencyDetector provides advanced circular dependency detection
type CircularDependencyDetector struct {
	logger *zap.Logger
	config *CircularDetectionConfig
}

// CircularDetectionConfig represents configuration for circular dependency detection
type CircularDetectionConfig struct {
	MaxCycleLength     int     `json:"max_cycle_length"`     // Maximum cycle length to report
	MinSeverityLevel   string  `json:"min_severity_level"`   // Minimum severity to report (low, medium, high, critical)
	IncludeExternal    bool    `json:"include_external"`     // Include external dependencies in analysis
	DetectionDepth     int     `json:"detection_depth"`      // Maximum depth for cycle detection
	AnalyzeWeakCycles  bool    `json:"analyze_weak_cycles"`  // Include weak dependency cycles
	GroupByLanguage    bool    `json:"group_by_language"`    // Group results by programming language
	SeverityThresholds SeverityThresholds `json:"severity_thresholds"`
}

// SeverityThresholds defines thresholds for cycle severity classification
type SeverityThresholds struct {
	CriticalCycleLength int     `json:"critical_cycle_length"` // <= this length = critical
	HighCycleLength     int     `json:"high_cycle_length"`     // <= this length = high
	MediumCycleLength   int     `json:"medium_cycle_length"`   // <= this length = medium
	WeakCycleWeight     float64 `json:"weak_cycle_weight"`     // Weight threshold for weak cycles
}

// CircularDependencyReport represents a comprehensive analysis of circular dependencies
type CircularDependencyReport struct {
	Summary           *CircularSummary           `json:"summary"`
	CyclesByLanguage  map[string][]*EnhancedCycle `json:"cycles_by_language"`
	CyclesByLength    map[int][]*EnhancedCycle   `json:"cycles_by_length"`
	CyclesBySeverity  map[string][]*EnhancedCycle `json:"cycles_by_severity"`
	ImpactAnalysis    *ImpactAnalysis            `json:"impact_analysis"`
	BreakingStrategies []*BreakingStrategy       `json:"breaking_strategies"`
	Recommendations   []string                   `json:"recommendations"`
	GeneratedAt       time.Time                  `json:"generated_at"`
}

// CircularSummary provides high-level statistics about circular dependencies
type CircularSummary struct {
	TotalCycles        int                        `json:"total_cycles"`
	TotalNodes         int                        `json:"total_nodes"`
	AffectedNodes      int                        `json:"affected_nodes"`
	CriticalCycles     int                        `json:"critical_cycles"`
	HighSeverityCycles int                        `json:"high_severity_cycles"`
	AverageCycleLength float64                    `json:"average_cycle_length"`
	MaxCycleLength     int                        `json:"max_cycle_length"`
	LanguageBreakdown  map[string]int             `json:"language_breakdown"`
	SeverityDistribution map[string]int           `json:"severity_distribution"`
}

// EnhancedCycle represents a circular dependency with enhanced analysis
type EnhancedCycle struct {
	ID              string                 `json:"id"`
	Cycle           []string               `json:"cycle"`
	Length          int                    `json:"length"`
	Severity        string                 `json:"severity"`
	Weight          float64                `json:"weight"`
	Languages       []string               `json:"languages"`
	CycleType       string                 `json:"cycle_type"`       // direct, indirect, cross-language
	Description     string                 `json:"description"`
	Impact          string                 `json:"impact"`
	Suggestions     []string               `json:"suggestions"`
	BreakingPoints  []*BreakingPoint       `json:"breaking_points"`
	RelatedCycles   []string               `json:"related_cycles"`
	DetectedAt      time.Time              `json:"detected_at"`
	Edges           []*CycleEdge           `json:"edges"`
	Metrics         *CycleMetrics          `json:"metrics"`
}

// BreakingPoint represents a potential point to break a circular dependency
type BreakingPoint struct {
	FromNode      string  `json:"from_node"`
	ToNode        string  `json:"to_node"`
	Confidence    float64 `json:"confidence"`    // 0-1, how confident we are this is a good breaking point
	Impact        string  `json:"impact"`        // low, medium, high
	Strategy      string  `json:"strategy"`      // interface, injection, merge, extract, etc.
	Description   string  `json:"description"`
	Effort        string  `json:"effort"`        // low, medium, high
	Rationale     string  `json:"rationale"`
}

// CycleEdge represents an edge within a circular dependency
type CycleEdge struct {
	From     string             `json:"from"`
	To       string             `json:"to"`
	Type     DependencyType     `json:"type"`
	Strength DependencyStrength `json:"strength"`
	Weight   float64            `json:"weight"`
	Language string             `json:"language"`
	Location *SourceLocation    `json:"location,omitempty"`
}

// CycleMetrics provides metrics about a cycle
type CycleMetrics struct {
	TotalWeight       float64 `json:"total_weight"`
	AverageWeight     float64 `json:"average_weight"`
	StrongEdges       int     `json:"strong_edges"`
	WeakEdges         int     `json:"weak_edges"`
	OptionalEdges     int     `json:"optional_edges"`
	CrossLanguageEdges int    `json:"cross_language_edges"`
	Complexity        float64 `json:"complexity"`
}

// ImpactAnalysis provides analysis of the impact of circular dependencies
type ImpactAnalysis struct {
	MostAffectedNodes []string                   `json:"most_affected_nodes"`
	CriticalPaths     [][]string                 `json:"critical_paths"`
	LanguageImpact    map[string]*LanguageImpact `json:"language_impact"`
	SystemComplexity  float64                    `json:"system_complexity"`
	TestabilityScore  float64                    `json:"testability_score"`
	MaintainabilityScore float64                 `json:"maintainability_score"`
}

// LanguageImpact represents the impact on a specific programming language
type LanguageImpact struct {
	Language       string  `json:"language"`
	CycleCount     int     `json:"cycle_count"`
	AffectedModules int    `json:"affected_modules"`
	ComplexityScore float64 `json:"complexity_score"`
	Recommendations []string `json:"recommendations"`
}

// BreakingStrategy represents a strategy for breaking circular dependencies
type BreakingStrategy struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ApplicableTo []string `json:"applicable_to"` // cycle IDs this strategy applies to
	Priority     int      `json:"priority"`      // 1-10, higher = more important
	Effort       string   `json:"effort"`        // low, medium, high
	Impact       string   `json:"impact"`        // low, medium, high
	Steps        []string `json:"steps"`
	Examples     []string `json:"examples"`
	Risks        []string `json:"risks"`
}

// NewCircularDependencyDetector creates a new circular dependency detector
func NewCircularDependencyDetector(logger *zap.Logger, config *CircularDetectionConfig) *CircularDependencyDetector {
	if config == nil {
		config = &CircularDetectionConfig{
			MaxCycleLength:   10,
			MinSeverityLevel: "low",
			IncludeExternal:  false,
			DetectionDepth:   20,
			AnalyzeWeakCycles: true,
			GroupByLanguage:  true,
			SeverityThresholds: SeverityThresholds{
				CriticalCycleLength: 2,
				HighCycleLength:     3,
				MediumCycleLength:   5,
				WeakCycleWeight:     0.5,
			},
		}
	}

	return &CircularDependencyDetector{
		logger: logger,
		config: config,
	}
}

// DetectCircularDependencies performs advanced circular dependency detection
func (cdd *CircularDependencyDetector) DetectCircularDependencies(result *DependencyResult) (*CircularDependencyReport, error) {
	cdd.logger.Info("Starting advanced circular dependency detection")

	// Build enhanced dependency graph
	graph := cdd.buildEnhancedGraph(result)

	// Detect all cycles using multiple algorithms
	allCycles := cdd.detectAllCycles(graph)

	// Enhance cycles with detailed analysis
	enhancedCycles := cdd.enhanceCycles(allCycles, graph, result)

	// Filter cycles based on configuration
	filteredCycles := cdd.filterCycles(enhancedCycles)

	// Generate breaking strategies
	strategies := cdd.generateBreakingStrategies(filteredCycles, graph)

	// Perform impact analysis
	impact := cdd.analyzeImpact(filteredCycles, graph, result)

	// Generate summary and recommendations
	summary := cdd.generateSummary(filteredCycles)
	recommendations := cdd.generateRecommendations(filteredCycles, impact)

	report := &CircularDependencyReport{
		Summary:            summary,
		CyclesByLanguage:   cdd.groupCyclesByLanguage(filteredCycles),
		CyclesByLength:     cdd.groupCyclesByLength(filteredCycles),
		CyclesBySeverity:   cdd.groupCyclesBySeverity(filteredCycles),
		ImpactAnalysis:     impact,
		BreakingStrategies: strategies,
		Recommendations:    recommendations,
		GeneratedAt:        time.Now(),
	}

	cdd.logger.Info("Circular dependency detection completed",
		zap.Int("total_cycles", len(filteredCycles)),
		zap.Int("critical_cycles", summary.CriticalCycles))

	return report, nil
}

// buildEnhancedGraph builds an enhanced graph for cycle detection
func (cdd *CircularDependencyDetector) buildEnhancedGraph(result *DependencyResult) map[string][]*CycleEdge {
	graph := make(map[string][]*CycleEdge)

	for _, dep := range result.Dependencies {
		// Skip external dependencies if not configured to include them
		if dep.External && !cdd.config.IncludeExternal {
			continue
		}

		// Skip weak cycles if not configured to analyze them
		if !cdd.config.AnalyzeWeakCycles && dep.Strength == DependencyStrengthWeak {
			continue
		}

		edge := &CycleEdge{
			From:     dep.From,
			To:       dep.To,
			Type:     dep.Type,
			Strength: dep.Strength,
			Weight:   cdd.calculateEdgeWeight(dep),
			Language: dep.Language,
			Location: &dep.Location,
		}

		graph[dep.From] = append(graph[dep.From], edge)
	}

	return graph
}

// detectAllCycles detects all cycles using enhanced algorithms
func (cdd *CircularDependencyDetector) detectAllCycles(graph map[string][]*CycleEdge) [][]*CycleEdge {
	var allCycles [][]*CycleEdge

	// Use Tarjan's strongly connected components algorithm for better cycle detection
	cycles := cdd.findStronglyConnectedComponents(graph)
	
	// Convert SCCs to cycle paths
	for _, scc := range cycles {
		if len(scc) > 1 {
			cyclePaths := cdd.findCyclePathsInSCC(scc, graph)
			allCycles = append(allCycles, cyclePaths...)
		}
	}

	// Also use Johnson's algorithm for finding all elementary cycles
	johnsonCycles := cdd.johnsonCycleDetection(graph)
	allCycles = append(allCycles, johnsonCycles...)

	// Remove duplicates and filter by length
	allCycles = cdd.removeDuplicateCycles(allCycles)
	allCycles = cdd.filterCyclesByLength(allCycles)

	return allCycles
}

// findStronglyConnectedComponents finds SCCs using Tarjan's algorithm
func (cdd *CircularDependencyDetector) findStronglyConnectedComponents(graph map[string][]*CycleEdge) [][]string {
	index := 0
	stack := make([]string, 0)
	indices := make(map[string]int)
	lowlinks := make(map[string]int)
	onStack := make(map[string]bool)
	var sccs [][]string

	var strongConnect func(string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, edge := range graph[v] {
			w := edge.To
			if _, exists := indices[w]; !exists {
				strongConnect(w)
				if lowlinks[w] < lowlinks[v] {
					lowlinks[v] = lowlinks[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlinks[v] {
					lowlinks[v] = indices[w]
				}
			}
		}

		if lowlinks[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			if len(scc) > 1 {
				sccs = append(sccs, scc)
			}
		}
	}

	for node := range graph {
		if _, exists := indices[node]; !exists {
			strongConnect(node)
		}
	}

	return sccs
}

// findCyclePathsInSCC finds actual cycle paths within an SCC
func (cdd *CircularDependencyDetector) findCyclePathsInSCC(scc []string, graph map[string][]*CycleEdge) [][]*CycleEdge {
	var cycles [][]*CycleEdge
	
	// For each node in SCC, try to find cycles starting from it
	for _, startNode := range scc {
		visited := make(map[string]bool)
		path := make([]*CycleEdge, 0)
		
		var dfs func(string, string) bool
		dfs = func(current, target string) bool {
			if current == target && len(path) > 0 {
				// Found a cycle
				cycleCopy := make([]*CycleEdge, len(path))
				copy(cycleCopy, path)
				cycles = append(cycles, cycleCopy)
				return true
			}
			
			if visited[current] {
				return false
			}
			
			visited[current] = true
			
			for _, edge := range graph[current] {
				if edge.To == target || (!visited[edge.To] && cdd.contains(scc, edge.To)) {
					path = append(path, edge)
					if dfs(edge.To, target) && len(cycles) < 10 { // Limit cycles per SCC
						return true
					}
					path = path[:len(path)-1]
				}
			}
			
			visited[current] = false
			return false
		}
		
		dfs(startNode, startNode)
	}
	
	return cycles
}

// johnsonCycleDetection implements Johnson's algorithm for finding all elementary cycles
func (cdd *CircularDependencyDetector) johnsonCycleDetection(graph map[string][]*CycleEdge) [][]*CycleEdge {
	var cycles [][]*CycleEdge
	blocked := make(map[string]bool)
	blockedMap := make(map[string]map[string]bool)
	stack := make([]*CycleEdge, 0)
	
	nodes := make([]string, 0, len(graph))
	for node := range graph {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	
	var circuit func(string, string) bool
	circuit = func(v, s string) bool {
		found := false
		stack = append(stack, &CycleEdge{From: v}) // Placeholder edge
		blocked[v] = true
		
		for _, edge := range graph[v] {
			w := edge.To
			if w == s {
				// Found cycle
				cycleCopy := make([]*CycleEdge, len(stack))
				copy(cycleCopy, stack)
				cycleCopy[len(cycleCopy)-1] = edge // Replace placeholder
				cycles = append(cycles, cycleCopy)
				found = true
			} else if !blocked[w] {
				if circuit(w, s) {
					found = true
				}
			}
		}
		
		if found {
			cdd.unblock(v, blocked, blockedMap)
		} else {
			for _, edge := range graph[v] {
				w := edge.To
				if blockedMap[w] == nil {
					blockedMap[w] = make(map[string]bool)
				}
				blockedMap[w][v] = true
			}
		}
		
		stack = stack[:len(stack)-1]
		return found
	}
	
	for i, s := range nodes {
		// Limit detection to prevent excessive computation
		if len(cycles) > 100 {
			break
		}
		
		// Reset for this starting node
		for _, node := range nodes[i:] {
			blocked[node] = false
			blockedMap[node] = make(map[string]bool)
		}
		
		circuit(s, s)
	}
	
	return cycles
}

// unblock unblocks a node in Johnson's algorithm
func (cdd *CircularDependencyDetector) unblock(u string, blocked map[string]bool, blockedMap map[string]map[string]bool) {
	blocked[u] = false
	for w := range blockedMap[u] {
		delete(blockedMap[u], w)
		if blocked[w] {
			cdd.unblock(w, blocked, blockedMap)
		}
	}
}

// contains checks if a slice contains a string
func (cdd *CircularDependencyDetector) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// enhanceCycles enhances raw cycles with detailed analysis
func (cdd *CircularDependencyDetector) enhanceCycles(rawCycles [][]*CycleEdge, graph map[string][]*CycleEdge, result *DependencyResult) []*EnhancedCycle {
	var enhanced []*EnhancedCycle

	for i, cycle := range rawCycles {
		enhancedCycle := &EnhancedCycle{
			ID:          fmt.Sprintf("cycle_%d", i+1),
			Cycle:       cdd.extractNodePath(cycle),
			Length:      len(cycle),
			Edges:       cycle,
			DetectedAt:  time.Now(),
			Languages:   cdd.extractLanguages(cycle),
			Metrics:     cdd.calculateCycleMetrics(cycle),
		}

		enhancedCycle.Weight = enhancedCycle.Metrics.TotalWeight
		enhancedCycle.Severity = cdd.calculateSeverity(enhancedCycle)
		enhancedCycle.CycleType = cdd.determineCycleType(enhancedCycle)
		enhancedCycle.Description = cdd.generateCycleDescription(enhancedCycle)
		enhancedCycle.Impact = cdd.assessCycleImpact(enhancedCycle)
		enhancedCycle.Suggestions = cdd.generateCycleSuggestions(enhancedCycle)
		enhancedCycle.BreakingPoints = cdd.identifyBreakingPoints(enhancedCycle, graph)

		enhanced = append(enhanced, enhancedCycle)
	}

	// Find related cycles
	for _, cycle := range enhanced {
		cycle.RelatedCycles = cdd.findRelatedCycles(cycle, enhanced)
	}

	return enhanced
}

// calculateEdgeWeight calculates the weight of an edge
func (cdd *CircularDependencyDetector) calculateEdgeWeight(dep *Dependency) float64 {
	weight := 1.0

	switch dep.Strength {
	case DependencyStrengthStrong:
		weight = 3.0
	case DependencyStrengthWeak:
		weight = 1.0
	case DependencyStrengthOptional:
		weight = 0.3
	}

	switch dep.Type {
	case DependencyTypeImport:
		weight *= 1.0
	case DependencyTypeRequire:
		weight *= 1.2
	case DependencyTypeInclude:
		weight *= 0.8
	case DependencyTypeInherit:
		weight *= 1.5
	}

	if dep.External {
		weight *= 0.5
	}

	return weight
}

// extractNodePath extracts the node path from a cycle of edges
func (cdd *CircularDependencyDetector) extractNodePath(cycle []*CycleEdge) []string {
	if len(cycle) == 0 {
		return []string{}
	}

	path := make([]string, len(cycle)+1)
	for i, edge := range cycle {
		path[i] = edge.From
		if i == len(cycle)-1 {
			path[i+1] = edge.To
		}
	}

	return path
}

// extractLanguages extracts unique languages from a cycle
func (cdd *CircularDependencyDetector) extractLanguages(cycle []*CycleEdge) []string {
	languageSet := make(map[string]bool)
	for _, edge := range cycle {
		languageSet[edge.Language] = true
	}

	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	sort.Strings(languages)
	return languages
}

// calculateCycleMetrics calculates metrics for a cycle
func (cdd *CircularDependencyDetector) calculateCycleMetrics(cycle []*CycleEdge) *CycleMetrics {
	metrics := &CycleMetrics{}

	for _, edge := range cycle {
		metrics.TotalWeight += edge.Weight

		switch edge.Strength {
		case DependencyStrengthStrong:
			metrics.StrongEdges++
		case DependencyStrengthWeak:
			metrics.WeakEdges++
		case DependencyStrengthOptional:
			metrics.OptionalEdges++
		}
	}

	if len(cycle) > 0 {
		metrics.AverageWeight = metrics.TotalWeight / float64(len(cycle))
	}

	// Calculate complexity based on cycle length and weight
	metrics.Complexity = float64(len(cycle)) * (1.0 + metrics.AverageWeight)

	// Count cross-language edges
	languages := make(map[string]bool)
	for _, edge := range cycle {
		languages[edge.Language] = true
	}
	if len(languages) > 1 {
		metrics.CrossLanguageEdges = len(cycle) // Approximate
	}

	return metrics
}

// calculateSeverity calculates the severity of a cycle
func (cdd *CircularDependencyDetector) calculateSeverity(cycle *EnhancedCycle) string {
	thresholds := cdd.config.SeverityThresholds

	// Base severity on cycle length
	if cycle.Length <= thresholds.CriticalCycleLength {
		return "critical"
	}
	if cycle.Length <= thresholds.HighCycleLength {
		return "high"
	}
	if cycle.Length <= thresholds.MediumCycleLength {
		return "medium"
	}

	// Consider weight for weak cycles
	if cycle.Weight < thresholds.WeakCycleWeight {
		return "low"
	}

	// Consider cross-language complexity
	if len(cycle.Languages) > 1 {
		return "medium"
	}

	return "low"
}

// Additional helper methods would continue here...
// (The file is getting quite long, so I'll implement the remaining methods in a separate file)

// removeDuplicateCycles removes duplicate cycles
func (cdd *CircularDependencyDetector) removeDuplicateCycles(cycles [][]*CycleEdge) [][]*CycleEdge {
	seen := make(map[string]bool)
	var unique [][]*CycleEdge

	for _, cycle := range cycles {
		key := cdd.cycleKey(cycle)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
}

// cycleKey generates a unique key for a cycle
func (cdd *CircularDependencyDetector) cycleKey(cycle []*CycleEdge) string {
	nodes := cdd.extractNodePath(cycle)
	if len(nodes) == 0 {
		return ""
	}

	// Normalize the cycle by starting from the lexicographically smallest node
	minIdx := 0
	for i, node := range nodes[:len(nodes)-1] { // Exclude last duplicate node
		if node < nodes[minIdx] {
			minIdx = i
		}
	}

	normalized := make([]string, len(nodes)-1)
	for i := 0; i < len(nodes)-1; i++ {
		normalized[i] = nodes[(minIdx+i)%(len(nodes)-1)]
	}

	return strings.Join(normalized, "->")
}

// filterCyclesByLength filters cycles by maximum length
func (cdd *CircularDependencyDetector) filterCyclesByLength(cycles [][]*CycleEdge) [][]*CycleEdge {
	var filtered [][]*CycleEdge

	for _, cycle := range cycles {
		if len(cycle) <= cdd.config.MaxCycleLength {
			filtered = append(filtered, cycle)
		}
	}

	return filtered
}
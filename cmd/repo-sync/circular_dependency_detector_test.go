package reposync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewCircularDependencyDetector(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with nil config (should use defaults)
	detector1 := NewCircularDependencyDetector(logger, nil)
	assert.NotNil(t, detector1)
	assert.Equal(t, 10, detector1.config.MaxCycleLength)
	assert.Equal(t, "low", detector1.config.MinSeverityLevel)

	// Test with custom config
	config := &CircularDetectionConfig{
		MaxCycleLength:   5,
		MinSeverityLevel: "medium",
		IncludeExternal:  true,
		DetectionDepth:   15,
	}
	detector2 := NewCircularDependencyDetector(logger, config)
	assert.NotNil(t, detector2)
	assert.Equal(t, config, detector2.config)
}

func TestCircularDependencyDetector_DetectCircularDependencies(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &CircularDetectionConfig{
		MaxCycleLength:    10,
		MinSeverityLevel:  "low",
		IncludeExternal:   false,
		AnalyzeWeakCycles: true,
		GroupByLanguage:   true,
	}
	detector := NewCircularDependencyDetector(logger, config)

	// Create test dependency result with circular dependencies
	result := &DependencyResult{
		Repository: "/test/repo",
		Dependencies: []*Dependency{
			// Simple 2-node cycle: A -> B -> A
			{From: "moduleA", To: "moduleB", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: false},
			{From: "moduleB", To: "moduleA", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: false},
			
			// 3-node cycle: C -> D -> E -> C
			{From: "moduleC", To: "moduleD", Type: DependencyTypeRequire, Language: "javascript", Strength: DependencyStrengthWeak, External: false},
			{From: "moduleD", To: "moduleE", Type: DependencyTypeRequire, Language: "javascript", Strength: DependencyStrengthStrong, External: false},
			{From: "moduleE", To: "moduleC", Type: DependencyTypeRequire, Language: "javascript", Strength: DependencyStrengthStrong, External: false},
			
			// External dependency (should be excluded)
			{From: "moduleA", To: "external_lib", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: true},
			
			// Non-circular dependencies
			{From: "moduleF", To: "moduleG", Type: DependencyTypeImport, Language: "python", Strength: DependencyStrengthStrong, External: false},
		},
		Modules: map[string]*ModuleDependencies{
			"moduleA": {ModulePath: "moduleA", Language: "go"},
			"moduleB": {ModulePath: "moduleB", Language: "go"},
			"moduleC": {ModulePath: "moduleC", Language: "javascript"},
			"moduleD": {ModulePath: "moduleD", Language: "javascript"},
			"moduleE": {ModulePath: "moduleE", Language: "javascript"},
			"moduleF": {ModulePath: "moduleF", Language: "python"},
			"moduleG": {ModulePath: "moduleG", Language: "python"},
		},
	}

	report, err := detector.DetectCircularDependencies(result)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Should detect at least 2 cycles
	assert.True(t, report.Summary.TotalCycles >= 2)

	// Check summary
	assert.True(t, report.Summary.TotalNodes > 0)
	assert.True(t, report.Summary.AffectedNodes > 0)
	assert.Contains(t, report.Summary.LanguageBreakdown, "go")
	assert.Contains(t, report.Summary.LanguageBreakdown, "javascript")

	// Check cycles by language grouping
	assert.NotEmpty(t, report.CyclesByLanguage)
	assert.Contains(t, report.CyclesByLanguage, "go")
	assert.Contains(t, report.CyclesByLanguage, "javascript")

	// Check cycles by length grouping
	assert.NotEmpty(t, report.CyclesByLength)

	// Check cycles by severity grouping
	assert.NotEmpty(t, report.CyclesBySeverity)

	// Verify impact analysis
	assert.NotNil(t, report.ImpactAnalysis)
	assert.NotEmpty(t, report.ImpactAnalysis.LanguageImpact)

	// Verify breaking strategies
	assert.NotEmpty(t, report.BreakingStrategies)

	// Verify recommendations
	assert.NotEmpty(t, report.Recommendations)

	// Check timestamp
	assert.True(t, time.Since(report.GeneratedAt) < time.Minute)
}

func TestCircularDependencyDetector_CalculateEdgeWeight(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	tests := []struct {
		name     string
		dep      *Dependency
		expected float64
	}{
		{
			name: "strong import",
			dep: &Dependency{
				Type:     DependencyTypeImport,
				Strength: DependencyStrengthStrong,
				External: false,
			},
			expected: 3.0, // strong(3.0) * import(1.0) * internal(1.0)
		},
		{
			name: "weak require",
			dep: &Dependency{
				Type:     DependencyTypeRequire,
				Strength: DependencyStrengthWeak,
				External: false,
			},
			expected: 1.2, // weak(1.0) * require(1.2) * internal(1.0)
		},
		{
			name: "optional external",
			dep: &Dependency{
				Type:     DependencyTypeInclude,
				Strength: DependencyStrengthOptional,
				External: true,
			},
			expected: 0.12, // optional(0.3) * include(0.8) * external(0.5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := detector.calculateEdgeWeight(tt.dep)
			assert.InDelta(t, tt.expected, weight, 0.01)
		})
	}
}

func TestCircularDependencyDetector_CalculateSeverity(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &CircularDetectionConfig{
		SeverityThresholds: SeverityThresholds{
			CriticalCycleLength: 2,
			HighCycleLength:     3,
			MediumCycleLength:   5,
			WeakCycleWeight:     0.5,
		},
	}
	detector := NewCircularDependencyDetector(logger, config)

	tests := []struct {
		name     string
		cycle    *EnhancedCycle
		expected string
	}{
		{
			name: "critical short cycle",
			cycle: &EnhancedCycle{
				Length:    2,
				Weight:    3.0,
				Languages: []string{"go"},
			},
			expected: "critical",
		},
		{
			name: "high medium cycle",
			cycle: &EnhancedCycle{
				Length:    3,
				Weight:    2.0,
				Languages: []string{"javascript"},
			},
			expected: "high",
		},
		{
			name: "medium length cycle",
			cycle: &EnhancedCycle{
				Length:    4,
				Weight:    1.5,
				Languages: []string{"python"},
			},
			expected: "medium",
		},
		{
			name: "low weak cycle",
			cycle: &EnhancedCycle{
				Length:    6,
				Weight:    0.3,
				Languages: []string{"go"},
			},
			expected: "low",
		},
		{
			name: "medium cross-language",
			cycle: &EnhancedCycle{
				Length:    6,
				Weight:    1.0,
				Languages: []string{"go", "javascript"},
			},
			expected: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := detector.calculateSeverity(tt.cycle)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestCircularDependencyDetector_DetermineCycleType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	tests := []struct {
		name     string
		cycle    *EnhancedCycle
		expected string
	}{
		{
			name: "cross-language cycle",
			cycle: &EnhancedCycle{
				Length:    3,
				Languages: []string{"go", "javascript"},
			},
			expected: "cross-language",
		},
		{
			name: "direct cycle",
			cycle: &EnhancedCycle{
				Length:    2,
				Languages: []string{"go"},
			},
			expected: "direct",
		},
		{
			name: "indirect cycle",
			cycle: &EnhancedCycle{
				Length:    4,
				Languages: []string{"python"},
			},
			expected: "indirect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycleType := detector.determineCycleType(tt.cycle)
			assert.Equal(t, tt.expected, cycleType)
		})
	}
}

func TestCircularDependencyDetector_ExtractNodePath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	edges := []*CycleEdge{
		{From: "A", To: "B"},
		{From: "B", To: "C"},
		{From: "C", To: "A"},
	}

	path := detector.extractNodePath(edges)
	expected := []string{"A", "B", "C", "A"}
	assert.Equal(t, expected, path)

	// Test empty cycle
	emptyPath := detector.extractNodePath([]*CycleEdge{})
	assert.Empty(t, emptyPath)
}

func TestCircularDependencyDetector_CalculateCycleMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	edges := []*CycleEdge{
		{Weight: 2.0, Strength: DependencyStrengthStrong, Language: "go"},
		{Weight: 1.0, Strength: DependencyStrengthWeak, Language: "go"},
		{Weight: 0.5, Strength: DependencyStrengthOptional, Language: "javascript"},
	}

	metrics := detector.calculateCycleMetrics(edges)

	assert.Equal(t, 3.5, metrics.TotalWeight)
	assert.InDelta(t, 1.167, metrics.AverageWeight, 0.01)
	assert.Equal(t, 1, metrics.StrongEdges)
	assert.Equal(t, 1, metrics.WeakEdges)
	assert.Equal(t, 1, metrics.OptionalEdges)
	assert.True(t, metrics.Complexity > 0)
}

func TestCircularDependencyDetector_FilterCycles(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &CircularDetectionConfig{
		MaxCycleLength:   3,
		MinSeverityLevel: "medium",
	}
	detector := NewCircularDependencyDetector(logger, config)

	cycles := []*EnhancedCycle{
		{ID: "1", Length: 2, Severity: "critical"},
		{ID: "2", Length: 3, Severity: "medium"},
		{ID: "3", Length: 4, Severity: "high"},      // Should be filtered (too long)
		{ID: "4", Length: 2, Severity: "low"},       // Should be filtered (low severity)
		{ID: "5", Length: 3, Severity: "high"},
	}

	filtered := detector.filterCycles(cycles)

	assert.Len(t, filtered, 3) // Should have cycles 1, 2, and 5
	
	ids := make([]string, len(filtered))
	for i, cycle := range filtered {
		ids[i] = cycle.ID
	}
	
	assert.Contains(t, ids, "1")
	assert.Contains(t, ids, "2")
	assert.Contains(t, ids, "5")
	assert.NotContains(t, ids, "3")
	assert.NotContains(t, ids, "4")
}

func TestCircularDependencyDetector_GenerateRecommendations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	// Test with no cycles
	noCycles := []*EnhancedCycle{}
	noImpact := &ImpactAnalysis{
		SystemComplexity:     2.0,
		TestabilityScore:     8.0,
		MaintainabilityScore: 8.0,
		LanguageImpact:       make(map[string]*LanguageImpact),
	}
	
	recommendations := detector.generateRecommendations(noCycles, noImpact)
	assert.NotEmpty(t, recommendations)
	assert.Contains(t, recommendations[0], "No circular dependencies")

	// Test with critical cycles
	criticalCycles := []*EnhancedCycle{
		{Severity: "critical"},
		{Severity: "high"},
		{Severity: "medium"},
	}
	
	highImpact := &ImpactAnalysis{
		SystemComplexity:     8.0,
		TestabilityScore:     4.0,
		MaintainabilityScore: 3.0,
		LanguageImpact: map[string]*LanguageImpact{
			"go": {CycleCount: 5},
		},
	}
	
	criticalRecommendations := detector.generateRecommendations(criticalCycles, highImpact)
	assert.NotEmpty(t, criticalRecommendations)
	
	// Should contain recommendations about critical cycles, complexity, etc.
	recommendationText := strings.Join(criticalRecommendations, " ")
	assert.Contains(t, recommendationText, "critical")
	assert.Contains(t, recommendationText, "complexity")
}

func TestCircularDependencyDetector_GroupMethods(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	cycles := []*EnhancedCycle{
		{ID: "1", Length: 2, Severity: "critical", Languages: []string{"go"}},
		{ID: "2", Length: 3, Severity: "high", Languages: []string{"javascript"}},
		{ID: "3", Length: 2, Severity: "critical", Languages: []string{"go", "python"}},
		{ID: "4", Length: 4, Severity: "medium", Languages: []string{"python"}},
	}

	// Test grouping by language
	byLanguage := detector.groupCyclesByLanguage(cycles)
	assert.Len(t, byLanguage["go"], 2)      // cycles 1 and 3
	assert.Len(t, byLanguage["javascript"], 1) // cycle 2
	assert.Len(t, byLanguage["python"], 2)     // cycles 3 and 4

	// Test grouping by length
	byLength := detector.groupCyclesByLength(cycles)
	assert.Len(t, byLength[2], 2) // cycles 1 and 3
	assert.Len(t, byLength[3], 1) // cycle 2
	assert.Len(t, byLength[4], 1) // cycle 4

	// Test grouping by severity
	bySeverity := detector.groupCyclesBySeverity(cycles)
	assert.Len(t, bySeverity["critical"], 2) // cycles 1 and 3
	assert.Len(t, bySeverity["high"], 1)     // cycle 2
	assert.Len(t, bySeverity["medium"], 1)   // cycle 4
}

func TestCircularDependencyDetector_CycleKey(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	// Test cycle key generation and normalization
	cycle1 := []*CycleEdge{
		{From: "A", To: "B"},
		{From: "B", To: "C"},
		{From: "C", To: "A"},
	}

	cycle2 := []*CycleEdge{
		{From: "B", To: "C"},
		{From: "C", To: "A"},
		{From: "A", To: "B"},
	}

	key1 := detector.cycleKey(cycle1)
	key2 := detector.cycleKey(cycle2)

	// Both cycles should have the same normalized key
	assert.Equal(t, key1, key2)
	assert.NotEmpty(t, key1)
}

func TestCircularDependencyDetector_BreakingPointAnalysis(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewCircularDependencyDetector(logger, nil)

	cycle := &EnhancedCycle{
		Edges: []*CycleEdge{
			{From: "A", To: "B", Type: DependencyTypeImport, Strength: DependencyStrengthStrong},
			{From: "B", To: "A", Type: DependencyTypeInclude, Strength: DependencyStrengthWeak},
		},
		Languages: []string{"go"},
		Severity:  "high",
	}

	graph := make(map[string][]*CycleEdge)
	breakingPoints := detector.identifyBreakingPoints(cycle, graph)

	assert.Len(t, breakingPoints, 2)

	// Verify breaking points are sorted by confidence
	for i := 1; i < len(breakingPoints); i++ {
		assert.True(t, breakingPoints[i-1].Confidence >= breakingPoints[i].Confidence)
	}

	// Verify each breaking point has required fields
	for _, point := range breakingPoints {
		assert.NotEmpty(t, point.FromNode)
		assert.NotEmpty(t, point.ToNode)
		assert.True(t, point.Confidence >= 0 && point.Confidence <= 1)
		assert.NotEmpty(t, point.Impact)
		assert.NotEmpty(t, point.Strategy)
		assert.NotEmpty(t, point.Description)
		assert.NotEmpty(t, point.Effort)
		assert.NotEmpty(t, point.Rationale)
	}
}
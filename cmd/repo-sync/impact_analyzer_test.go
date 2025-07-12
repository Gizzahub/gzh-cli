package reposync

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewImpactAnalyzer(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with nil config (should use defaults)
	analyzer1 := NewImpactAnalyzer(logger, nil)
	assert.NotNil(t, analyzer1)
	assert.Equal(t, 10, analyzer1.config.MaxDepth)
	assert.Equal(t, false, analyzer1.config.IncludeExternalDeps)
	assert.Equal(t, true, analyzer1.config.AnalyzeTestImpact)

	// Test with custom config
	config := &ImpactAnalysisConfig{
		MaxDepth:            5,
		IncludeExternalDeps: true,
		AnalyzeTestImpact:   false,
		ImpactThreshold:     0.5,
	}
	analyzer2 := NewImpactAnalyzer(logger, config)
	assert.NotNil(t, analyzer2)
	assert.Equal(t, config, analyzer2.config)
}

func TestImpactAnalyzer_AnalyzeImpact(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &ImpactAnalysisConfig{
		MaxDepth:            5,
		IncludeExternalDeps: false,
		AnalyzeTestImpact:   true,
		ConsiderWeakDeps:    true,
		ImpactThreshold:     0.1,
		RiskLevels: RiskLevels{
			HighRiskThreshold:   7.0,
			MediumRiskThreshold: 4.0,
			LowRiskThreshold:    1.0,
		},
	}
	analyzer := NewImpactAnalyzer(logger, config)

	// Create test change set
	changeSet := &ChangeSet{
		ID:             "test_change_1",
		Description:    "Test modification",
		ChangedModules: []string{"core/auth", "core/user"},
		ChangedFiles:   []string{"core/auth/service.go", "core/user/model.go"},
		ChangeType:     "modification",
		Language:       "go",
		Author:         "test-user",
		Timestamp:      time.Now(),
	}

	// Create test dependency result
	depResult := &DependencyResult{
		Repository: "/test/repo",
		Dependencies: []*Dependency{
			// core/auth depends on core/user
			{From: "core/auth", To: "core/user", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: false},
			// api/handlers depends on core/auth
			{From: "api/handlers", To: "core/auth", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: false},
			// web/controllers depends on api/handlers
			{From: "web/controllers", To: "api/handlers", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthWeak, External: false},
			// external dependency (should be excluded)
			{From: "core/auth", To: "github.com/external/lib", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: true},
		},
		Modules: map[string]*ModuleDependencies{
			"core/auth":      {ModulePath: "core/auth", Language: "go"},
			"core/user":      {ModulePath: "core/user", Language: "go"},
			"api/handlers":   {ModulePath: "api/handlers", Language: "go"},
			"web/controllers": {ModulePath: "web/controllers", Language: "go"},
		},
	}

	report, err := analyzer.AnalyzeImpact(changeSet, depResult)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Verify basic report structure
	assert.Equal(t, changeSet, report.ChangeSet)
	assert.NotNil(t, report.Summary)
	assert.NotNil(t, report.RiskAssessment)
	assert.NotNil(t, report.TestImpact)
	assert.NotNil(t, report.PerformanceImpact)
	assert.NotEmpty(t, report.Recommendations)
	assert.NotEmpty(t, report.MitigationStrategies)

	// Verify affected modules
	assert.True(t, len(report.AffectedModules) >= 2) // At least the changed modules
	
	// Should include the changed modules themselves
	changedModuleNames := make([]string, len(report.AffectedModules))
	for i, module := range report.AffectedModules {
		changedModuleNames[i] = module.ModulePath
	}
	assert.Contains(t, changedModuleNames, "core/auth")
	assert.Contains(t, changedModuleNames, "core/user")

	// Verify impact paths
	assert.True(t, len(report.ImpactPaths) >= 0) // May be zero if no transitive impacts

	// Verify summary statistics
	assert.Equal(t, len(report.AffectedModules), report.Summary.TotalAffectedModules)
	assert.Equal(t, len(report.ImpactPaths), report.Summary.TotalImpactPaths)
	assert.Contains(t, []string{"low", "medium", "high", "critical"}, report.Summary.OverallRiskLevel)

	// Verify risk assessment
	assert.NotNil(t, report.RiskAssessment.RiskMetrics)
	assert.True(t, report.RiskAssessment.RiskMetrics.OverallRiskScore >= 0)
	assert.True(t, report.RiskAssessment.RiskMetrics.OverallRiskScore <= 10)
}

func TestImpactAnalyzer_CalculateImpactScore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	changeSet := &ChangeSet{
		ChangeType: "modification",
	}

	tests := []struct {
		name           string
		module         *AffectedModule
		expectedRange  [2]float64 // min, max expected score
	}{
		{
			name: "directly changed module",
			module: &AffectedModule{
				DistanceFromChange: 0,
				DependencyStrength: DependencyStrengthStrong,
				TestCoverage:      0.8,
			},
			expectedRange: [2]float64{9.0, 10.0},
		},
		{
			name: "immediate dependency",
			module: &AffectedModule{
				DistanceFromChange: 1,
				DependencyStrength: DependencyStrengthStrong,
				TestCoverage:      0.8,
			},
			expectedRange: [2]float64{7.0, 9.0},
		},
		{
			name: "weak transitive dependency",
			module: &AffectedModule{
				DistanceFromChange: 3,
				DependencyStrength: DependencyStrengthWeak,
				TestCoverage:      0.3,
			},
			expectedRange: [2]float64{1.0, 4.0},
		},
		{
			name: "optional distant dependency",
			module: &AffectedModule{
				DistanceFromChange: 5,
				DependencyStrength: DependencyStrengthOptional,
				TestCoverage:      0.9,
			},
			expectedRange: [2]float64{0.0, 2.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.calculateImpactScore(tt.module, changeSet)
			assert.True(t, score >= tt.expectedRange[0] && score <= tt.expectedRange[1],
				"Score %f not in expected range [%f, %f]", score, tt.expectedRange[0], tt.expectedRange[1])
		})
	}
}

func TestImpactAnalyzer_DetermineRiskLevel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &ImpactAnalysisConfig{
		RiskLevels: RiskLevels{
			HighRiskThreshold:   7.0,
			MediumRiskThreshold: 4.0,
			LowRiskThreshold:    1.0,
		},
	}
	analyzer := NewImpactAnalyzer(logger, config)

	tests := []struct {
		score    float64
		expected string
	}{
		{9.0, "high"},
		{7.0, "high"},
		{6.5, "medium"},
		{4.0, "medium"},
		{2.5, "low"},
		{1.0, "low"},
		{0.5, "minimal"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := analyzer.determineRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImpactAnalyzer_DetermineImpactType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	changeSet := &ChangeSet{
		ChangeType: "modification",
	}

	module := &AffectedModule{
		DistanceFromChange: 1,
		DependencyType:     DependencyTypeImport,
	}

	impactTypes := analyzer.determineImpactType(module, changeSet)
	
	assert.NotEmpty(t, impactTypes)
	assert.Contains(t, impactTypes, "immediate") // Distance 1
	assert.Contains(t, impactTypes, "compile")   // Import dependency
	assert.Contains(t, impactTypes, "behavioral") // Modification change
}

func TestImpactAnalyzer_FindPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	// Create test graph: A -> B -> C -> D
	graph := map[string][]*Dependency{
		"A": {{From: "A", To: "B"}},
		"B": {{From: "B", To: "C"}},
		"C": {{From: "C", To: "D"}},
	}

	// Test direct path
	path1 := analyzer.findPath("A", "B", graph)
	assert.Equal(t, []string{"A", "B"}, path1)

	// Test transitive path
	path2 := analyzer.findPath("A", "D", graph)
	assert.Equal(t, []string{"A", "B", "C", "D"}, path2)

	// Test non-existent path
	path3 := analyzer.findPath("D", "A", graph)
	assert.Nil(t, path3)

	// Test same source and target
	path4 := analyzer.findPath("A", "A", graph)
	assert.Equal(t, []string{"A"}, path4)
}

func TestImpactAnalyzer_CalculatePathWeight(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	graph := map[string][]*Dependency{
		"A": {{From: "A", To: "B", Strength: DependencyStrengthStrong}},   // Weight 3.0
		"B": {{From: "B", To: "C", Strength: DependencyStrengthWeak}},     // Weight 1.0
		"C": {{From: "C", To: "D", Strength: DependencyStrengthOptional}}, // Weight 0.3
	}

	path := []string{"A", "B", "C", "D"}
	weight := analyzer.calculatePathWeight(path, graph)
	
	expectedWeight := 3.0 + 1.0 + 0.3 // Sum of edge weights
	assert.InDelta(t, expectedWeight, weight, 0.1)
}

func TestImpactAnalyzer_EstimateTestEffort(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	changeSet := &ChangeSet{
		ChangeType: "modification",
	}

	tests := []struct {
		name            string
		moduleCount     int
		highRiskCount   int
		changeType      string
		expectedEffort  string
	}{
		{
			name:           "small change",
			moduleCount:    3,
			highRiskCount:  0,
			changeType:     "addition",
			expectedEffort: "low",
		},
		{
			name:           "medium change",
			moduleCount:    12,
			highRiskCount:  1,
			changeType:     "modification",
			expectedEffort: "medium",
		},
		{
			name:           "large change",
			moduleCount:    25,
			highRiskCount:  6,
			changeType:     "deletion",
			expectedEffort: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modules := make([]*AffectedModule, tt.moduleCount)
			for i := 0; i < tt.moduleCount; i++ {
				riskLevel := "low"
				if i < tt.highRiskCount {
					riskLevel = "high"
				}
				modules[i] = &AffectedModule{RiskLevel: riskLevel}
			}

			changeSet.ChangeType = tt.changeType
			effort := analyzer.estimateTestEffort(modules, changeSet)
			assert.Equal(t, tt.expectedEffort, effort)
		})
	}
}

func TestImpactAnalyzer_InferTestSuiteName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	tests := []struct {
		modulePath string
		language   string
		expected   string
	}{
		{"core/auth", "go", "auth_test"},
		{"services/user", "javascript", "services.user.test.js"},
		{"models/product", "python", "test_models_product.py"},
		{"com/example/service", "java", "ComExampleServiceTest.java"},
		{"simple", "go", "simple_test"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.language, tt.modulePath), func(t *testing.T) {
			result := analyzer.inferTestSuiteName(tt.modulePath, tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImpactAnalyzer_CreateChangeSetFromGitDiff(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	changedFiles := []string{
		"src/core/auth/service.go",
		"src/core/auth/handler.go",
		"src/api/routes.js",
		"test/core/auth_test.go",
	}

	changeSet := analyzer.CreateChangeSetFromGitDiff("abc123def456", "test-author", changedFiles)

	assert.NotNil(t, changeSet)
	assert.Equal(t, "changeset_abc123de", changeSet.ID)
	assert.Equal(t, "abc123def456", changeSet.CommitHash)
	assert.Equal(t, "test-author", changeSet.Author)
	assert.Equal(t, changedFiles, changeSet.ChangedFiles)
	assert.NotEmpty(t, changeSet.ChangedModules)
	assert.Contains(t, changeSet.ChangedModules, "core/auth")
	assert.Contains(t, changeSet.ChangedModules, "api")
	assert.Equal(t, "go", changeSet.Language) // Go files dominate
}

func TestImpactAnalyzer_RemoveDuplicates(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	input := []string{"a", "b", "a", "c", "b", "d"}
	result := analyzer.removeDuplicates(input)

	expected := []string{"a", "b", "c", "d"}
	assert.ElementsMatch(t, expected, result)
}

func TestImpactAnalyzer_GenerateRecommendations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	analyzer := NewImpactAnalyzer(logger, nil)

	changeSet := &ChangeSet{
		ChangeType: "deletion",
	}

	affectedModules := []*AffectedModule{
		{ModulePath: "module1", TestCoverage: 0.3},
		{ModulePath: "module2", TestCoverage: 0.9},
		{ModulePath: "module3", TestCoverage: 0.4},
	}

	riskAssessment := &RiskAssessment{
		OverallRisk: "high",
	}

	recommendations := analyzer.generateRecommendations(changeSet, affectedModules, riskAssessment)

	assert.NotEmpty(t, recommendations)
	
	// Should contain high risk recommendation
	recommendationText := strings.Join(recommendations, " ")
	assert.Contains(t, recommendationText, "High risk")
	
	// Should contain deletion-specific recommendation
	assert.Contains(t, recommendationText, "Deletion changes")
}

func TestImpactAnalyzer_ShouldExcludeModule(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &ImpactAnalysisConfig{
		ExcludePatterns: []string{"test_*", "*_test.*", "mock_*"},
	}
	analyzer := NewImpactAnalyzer(logger, config)

	tests := []struct {
		modulePath string
		expected   bool
	}{
		{"core/auth", false},
		{"test_auth", true},
		{"auth_test.go", true},
		{"mock_service", true},
		{"normal/module", false},
	}

	for _, tt := range tests {
		t.Run(tt.modulePath, func(t *testing.T) {
			result := analyzer.shouldExcludeModule(tt.modulePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}
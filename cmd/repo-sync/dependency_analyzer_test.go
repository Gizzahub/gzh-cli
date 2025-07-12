package reposync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDependencyAnalyzer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{
		RepositoryPath:   "/test/repo",
		IncludeLanguages: []string{"go", "javascript"},
		OutputFormat:     "json",
	}

	analyzer := NewDependencyAnalyzer(logger, config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config, analyzer.config)
	assert.Contains(t, analyzer.parsers, "go")
	assert.Contains(t, analyzer.parsers, "javascript")
	assert.Contains(t, analyzer.parsers, "typescript")
	assert.Contains(t, analyzer.parsers, "python")
}

func TestDependencyAnalyzer_GetLanguageFromExtension(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "go"},
		{".js", "javascript"},
		{".jsx", "javascript"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".py", "python"},
		{".java", "java"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := analyzer.getLanguageFromExtension(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyAnalyzer_ShouldSkipDir(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{
		ExcludePatterns: []string{"custom_*"},
	}
	analyzer := NewDependencyAnalyzer(logger, config)

	tests := []struct {
		dirname  string
		expected bool
	}{
		{"node_modules", true},
		{".git", true},
		{"vendor", true},
		{"__pycache__", true},
		{"src", false},
		{"custom_folder", true},
		{"normal_folder", false},
	}

	for _, tt := range tests {
		t.Run(tt.dirname, func(t *testing.T) {
			result := analyzer.shouldSkipDir(tt.dirname)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyAnalyzer_ShouldIncludeFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{
		IncludeLanguages: []string{"go", "javascript"},
		ExcludePatterns:  []string{"*_test.go"},
	}
	analyzer := NewDependencyAnalyzer(logger, config)

	tests := []struct {
		filePath string
		expected bool
	}{
		{"/path/to/file.go", true},
		{"/path/to/file.js", true},
		{"/path/to/file.py", false},      // not in include list
		{"/path/to/file_test.go", false}, // excluded pattern
		{"/path/to/file.txt", false},     // unsupported extension
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := analyzer.shouldIncludeFile(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyAnalyzer_DetectCircularDependencies(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	// Create test result with circular dependencies
	result := &DependencyResult{
		Dependencies: []*Dependency{
			{From: "module_a", To: "module_b", External: false},
			{From: "module_b", To: "module_c", External: false},
			{From: "module_c", To: "module_a", External: false},
			{From: "module_d", To: "external_lib", External: true}, // external, should be ignored
		},
	}

	cyclicDeps := analyzer.detectCircularDependencies(result)

	assert.Len(t, cyclicDeps, 1)
	assert.Equal(t, 3, cyclicDeps[0].Length)
	assert.Contains(t, cyclicDeps[0].Cycle, "module_a")
	assert.Contains(t, cyclicDeps[0].Cycle, "module_b")
	assert.Contains(t, cyclicDeps[0].Cycle, "module_c")
}

func TestDependencyAnalyzer_CalculateCycleSeverity(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	tests := []struct {
		cycle    []string
		expected string
	}{
		{[]string{"a", "b", "a"}, "high"},
		{[]string{"a", "b", "c", "a"}, "high"},
		{[]string{"a", "b", "c", "d", "a"}, "medium"},
		{[]string{"a", "b", "c", "d", "e", "a"}, "medium"},
		{[]string{"a", "b", "c", "d", "e", "f", "a"}, "low"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := analyzer.calculateCycleSeverity(tt.cycle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyAnalyzer_AnalyzeExternalDependencies(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	result := &DependencyResult{
		Dependencies: []*Dependency{
			{To: "react", Language: "javascript", External: true, Version: "18.0.0"},
			{To: "react", Language: "javascript", External: true, Version: "18.0.0"},
			{To: "lodash", Language: "javascript", External: true, Version: "4.17.21"},
			{To: "internal_module", Language: "javascript", External: false},
		},
	}

	externalDeps := analyzer.analyzeExternalDependencies(result)

	assert.Len(t, externalDeps, 2)

	// Should be sorted by usage count
	assert.Equal(t, "react", externalDeps[0].Name)
	assert.Equal(t, 2, externalDeps[0].UsageCount)
	assert.Equal(t, "lodash", externalDeps[1].Name)
	assert.Equal(t, 1, externalDeps[1].UsageCount)
}

func TestDependencyAnalyzer_CalculateStatistics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	result := &DependencyResult{
		Dependencies: []*Dependency{
			{Language: "go", Type: DependencyTypeImport, Strength: DependencyStrengthStrong, External: true},
			{Language: "go", Type: DependencyTypeImport, Strength: DependencyStrengthWeak, External: false},
			{Language: "javascript", Type: DependencyTypeRequire, Strength: DependencyStrengthStrong, External: true},
		},
		CyclicDeps: []CyclicDependency{{}},
		Modules: map[string]*ModuleDependencies{
			"module1": {Dependencies: []*Dependency{{}, {}}},
			"module2": {Dependencies: []*Dependency{{}}},
		},
	}

	stats := analyzer.calculateStatistics(result)

	assert.Equal(t, 3, stats.TotalDependencies)
	assert.Equal(t, 1, stats.InternalDependencies)
	assert.Equal(t, 2, stats.ExternalDependencies)
	assert.Equal(t, 1, stats.CircularDependencies)

	assert.Equal(t, 2, stats.LanguageBreakdown["go"])
	assert.Equal(t, 1, stats.LanguageBreakdown["javascript"])

	assert.Equal(t, 2, stats.TypeBreakdown[DependencyTypeImport])
	assert.Equal(t, 1, stats.TypeBreakdown[DependencyTypeRequire])

	assert.Equal(t, 2, stats.StrengthBreakdown[DependencyStrengthStrong])
	assert.Equal(t, 1, stats.StrengthBreakdown[DependencyStrengthWeak])
}

func TestDependencyAnalyzer_InferModuleFromPath(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "dependency_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create go.mod file
	goModDir := filepath.Join(tempDir, "go_project")
	err = os.MkdirAll(goModDir, 0755)
	require.NoError(t, err)

	goModContent := "module github.com/test/project\n\ngo 1.19\n"
	err = os.WriteFile(filepath.Join(goModDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{
		RepositoryPath: tempDir,
	}
	analyzer := NewDependencyAnalyzer(logger, config)

	testFile := filepath.Join(goModDir, "cmd", "main.go")
	result := analyzer.inferModuleFromPath(testFile)

	expected := filepath.Join("go_project", "cmd")
	assert.Equal(t, expected, result)
}

func TestDependency_Types(t *testing.T) {
	// Test dependency types
	assert.Equal(t, DependencyType("import"), DependencyTypeImport)
	assert.Equal(t, DependencyType("require"), DependencyTypeRequire)
	assert.Equal(t, DependencyType("include"), DependencyTypeInclude)

	// Test dependency strengths
	assert.Equal(t, DependencyStrength("strong"), DependencyStrengthStrong)
	assert.Equal(t, DependencyStrength("weak"), DependencyStrengthWeak)
	assert.Equal(t, DependencyStrength("optional"), DependencyStrengthOptional)
}

func TestDependencyAnalyzer_AddParser(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	// Create a mock parser
	mockParser := &GoDependencyParser{logger: logger}

	analyzer.AddParser(mockParser)

	assert.Contains(t, analyzer.parsers, "go")
	assert.Equal(t, mockParser, analyzer.parsers["go"])
}

func TestDependencyAnalyzer_SaveResult(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DependencyConfig{}
	analyzer := NewDependencyAnalyzer(logger, config)

	tempFile := filepath.Join(os.TempDir(), "test_result.json")
	defer os.Remove(tempFile)

	result := &DependencyResult{
		Repository:   "/test/repo",
		AnalysisTime: time.Now(),
		TotalFiles:   5,
		TotalModules: 2,
		Dependencies: []*Dependency{
			{From: "module_a", To: "module_b", Type: DependencyTypeImport},
		},
	}

	err := analyzer.SaveResult(result, tempFile)
	assert.NoError(t, err)

	// Verify file was created and has content
	data, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "module_a")
	assert.Contains(t, string(data), "module_b")
}

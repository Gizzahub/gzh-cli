//nolint:testpackage // White-box testing needed for internal function access
package profile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfileAnalyzer(t *testing.T) {
	outputDir := "tmp/test-profiles"
	analyzer := NewProfileAnalyzer(outputDir)

	assert.NotNil(t, analyzer)
	assert.Equal(t, outputDir, analyzer.outputDir)
	assert.NotNil(t, analyzer.profiler)
}

func TestNewCompareCmd(t *testing.T) {
	cmd := newCompareCmd()

	assert.Contains(t, cmd.Use, "compare")
	assert.Equal(t, "Compare two profiles to identify performance differences", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	thresholdFlag := cmd.Flags().Lookup("threshold")
	assert.NotNil(t, thresholdFlag)
	assert.Equal(t, "5", thresholdFlag.DefValue)

	// Check that it requires exactly 2 arguments (Args field should be set)
	assert.NotNil(t, cmd.Args)
}

func TestNewContinuousCmd(t *testing.T) {
	cmd := newContinuousCmd()

	assert.Equal(t, "continuous", cmd.Use)
	assert.Equal(t, "Run continuous profiling over time", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	intervalFlag := cmd.Flags().Lookup("interval")
	assert.NotNil(t, intervalFlag)
	assert.Equal(t, "5m0s", intervalFlag.DefValue)

	durationFlag := cmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)
	assert.Equal(t, "1h0m0s", durationFlag.DefValue)

	typeFlag := cmd.Flags().Lookup("type")
	assert.NotNil(t, typeFlag)
	assert.Equal(t, "cpu", typeFlag.DefValue)

	autoAnalyzeFlag := cmd.Flags().Lookup("auto-analyze")
	assert.NotNil(t, autoAnalyzeFlag)
	assert.Equal(t, "false", autoAnalyzeFlag.DefValue)
}

func TestNewAnalyzeCmd(t *testing.T) {
	cmd := newAnalyzeCmd()

	assert.Contains(t, cmd.Use, "analyze")
	assert.Equal(t, "Analyze profile for performance issues", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	thresholdFlag := cmd.Flags().Lookup("threshold")
	assert.NotNil(t, thresholdFlag)
	assert.Equal(t, "5", thresholdFlag.DefValue)

	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	autoSuggestFlag := cmd.Flags().Lookup("auto-suggest")
	assert.NotNil(t, autoSuggestFlag)
	assert.Equal(t, "true", autoSuggestFlag.DefValue)

	// Check that it requires exactly 1 argument (Args field should be set)
	assert.NotNil(t, cmd.Args)
}

func TestCompareProfiles(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Test with baseline and current files
	comparison, err := analyzer.CompareProfiles("baseline.prof", "current.prof", 5.0)

	require.NoError(t, err)
	assert.NotNil(t, comparison)
	assert.Equal(t, "baseline.prof", comparison.BaselineFile)
	assert.Equal(t, "current.prof", comparison.CurrentFile)

	// Should have some simulated data
	assert.True(t, len(comparison.Improvements) > 0 || len(comparison.Regressions) > 0)
	assert.NotNil(t, comparison.Summary)
}

func TestCompareProfilesNonBaseline(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Test with files that don't match the simulation pattern
	comparison, err := analyzer.CompareProfiles("test1.prof", "test2.prof", 5.0)

	require.NoError(t, err)
	assert.NotNil(t, comparison)
	assert.Equal(t, "test1.prof", comparison.BaselineFile)
	assert.Equal(t, "test2.prof", comparison.CurrentFile)

	// Should have empty results for non-matching files
	assert.Equal(t, 0, len(comparison.Improvements))
	assert.Equal(t, 0, len(comparison.Regressions))
	assert.Equal(t, 0, len(comparison.Issues))
}

func TestAnalyzeProfile(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create temporary profile file
	tempDir := t.TempDir()
	cpuProfile := filepath.Join(tempDir, "cpu_profile.prof")
	err := os.WriteFile(cpuProfile, []byte("fake cpu profile data"), 0o644)
	require.NoError(t, err)

	// Analyze CPU profile
	issues, err := analyzer.AnalyzeProfile(cpuProfile, 5.0)

	require.NoError(t, err)
	assert.NotNil(t, issues)

	// Should detect simulated CPU issues
	assert.True(t, len(issues) >= 1)

	// Check that issues have required fields
	for _, issue := range issues {
		assert.NotEmpty(t, issue.Type)
		assert.NotEmpty(t, issue.Severity)
		assert.NotEmpty(t, issue.Description)
		assert.True(t, issue.Impact >= 5.0) // Should respect threshold
	}
}

func TestAnalyzeProfileMemory(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create temporary memory profile file
	tempDir := t.TempDir()
	memProfile := filepath.Join(tempDir, "memory_profile.prof")
	err := os.WriteFile(memProfile, []byte("fake memory profile data"), 0o644)
	require.NoError(t, err)

	// Analyze memory profile
	issues, err := analyzer.AnalyzeProfile(memProfile, 10.0)

	require.NoError(t, err)
	assert.NotNil(t, issues)

	// Should detect memory issues above threshold
	for _, issue := range issues {
		assert.True(t, issue.Impact >= 10.0) // Should respect threshold
		assert.Contains(t, []string{"memory_leak", "excessive_allocation"}, issue.Type)
	}
}

func TestAnalyzeProfileGoroutine(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create temporary goroutine profile file
	tempDir := t.TempDir()
	goroutineProfile := filepath.Join(tempDir, "goroutine_profile.prof")
	err := os.WriteFile(goroutineProfile, []byte("fake goroutine profile data"), 0o644)
	require.NoError(t, err)

	// Analyze goroutine profile
	issues, err := analyzer.AnalyzeProfile(goroutineProfile, 1.0)

	require.NoError(t, err)
	assert.NotNil(t, issues)

	// Should detect goroutine leaks
	for _, issue := range issues {
		if issue.Type == "goroutine_leak" {
			assert.Equal(t, "critical", issue.Severity)
			assert.Contains(t, issue.Description, "goroutines not terminating")
		}
	}
}

func TestAnalyzeProfileNonExistent(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Try to analyze non-existent file
	issues, err := analyzer.AnalyzeProfile("nonexistent.prof", 5.0)

	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Contains(t, err.Error(), "profile file not found")
}

func TestRunContinuousProfiling(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create a context with timeout for testing
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run continuous profiling for a very short duration
	err := analyzer.RunContinuousProfiling(ctx, "cpu", 50*time.Millisecond, 200*time.Millisecond, false)

	// Should not return an error when context is cancelled
	assert.NoError(t, err)
}

func TestGetSeverityEmoji(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "🔴"},
		{"warning", "🟡"},
		{"info", "🔵"},
		{"unknown", "ℹ️"},
		{"", "ℹ️"},
	}

	for _, tt := range tests {
		t.Run("severity "+tt.severity, func(t *testing.T) {
			result := getSeverityEmoji(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnhancedStats(t *testing.T) {
	stats := GetEnhancedStats()

	assert.NotNil(t, stats)

	// Check required top-level keys
	assert.Contains(t, stats, "timestamp")
	assert.Contains(t, stats, "memory")
	assert.Contains(t, stats, "gc")
	assert.Contains(t, stats, "runtime")

	// Check memory section
	memory := stats["memory"].(map[string]interface{})
	assert.Contains(t, memory, "heap_alloc")
	assert.Contains(t, memory, "heap_sys")
	assert.Contains(t, memory, "total_alloc")

	// Check GC section
	gc := stats["gc"].(map[string]interface{})
	assert.Contains(t, gc, "num_gc")
	assert.Contains(t, gc, "last_gc")

	// Check runtime section
	runtimeStats := stats["runtime"].(map[string]interface{})
	assert.Contains(t, runtimeStats, "goroutines")
	assert.Contains(t, runtimeStats, "num_cpu")
	assert.Contains(t, runtimeStats, "version")
}

func TestPrintComparisonText(t *testing.T) {
	comparison := &ProfileComparison{
		BaselineFile: "baseline.prof",
		CurrentFile:  "current.prof",
		Improvements: []ProfileDifference{
			{
				Function:      "test.function",
				Metric:        "CPU Time",
				BaseValue:     10.0,
				CurrentValue:  8.0,
				PercentChange: -20.0,
			},
		},
		Regressions: []ProfileDifference{
			{
				Function:      "slow.function",
				Metric:        "Memory",
				BaseValue:     5.0,
				CurrentValue:  7.0,
				PercentChange: 40.0,
			},
		},
		Issues: []PerformanceIssue{
			{
				Type:        "memory_leak",
				Severity:    "warning",
				Description: "Memory leak detected",
				Impact:      15.5,
				Suggestion:  "Fix the leak",
			},
		},
		Summary: ProfileComparisonSummary{
			TotalFunctions: 50,
			ImprovedCount:  1,
			RegressedCount: 1,
			OverallChange:  -2.5,
			Recommendation: "Optimize slow functions",
		},
	}

	// This should not panic or error
	err := printComparisonText(comparison)
	assert.NoError(t, err)
}

func TestPrintAnalysisText(t *testing.T) {
	issues := []PerformanceIssue{
		{
			Type:        "high_cpu_usage",
			Severity:    "critical",
			Description: "High CPU usage detected",
			Location:    "main.go:42",
			Impact:      25.5,
			Suggestion:  "Optimize the algorithm",
		},
		{
			Type:        "memory_leak",
			Severity:    "warning",
			Description: "Potential memory leak",
			Impact:      8.2,
			Suggestion:  "Check for proper cleanup",
		},
	}

	// Test with suggestions
	err := printAnalysisText(issues, true)
	assert.NoError(t, err)

	// Test without suggestions
	err = printAnalysisText(issues, false)
	assert.NoError(t, err)

	// Test with empty issues
	err = printAnalysisText([]PerformanceIssue{}, true)
	assert.NoError(t, err)
}

func TestCommandLongDescriptions(t *testing.T) {
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"compare", newCompareCmd},
		{"continuous", newContinuousCmd},
		{"analyze", newAnalyzeCmd},
	}

	for _, tc := range commands {
		t.Run("command long description "+tc.name, func(t *testing.T) {
			cmd := tc.cmd()

			// Long description should be detailed
			assert.True(t, len(cmd.Long) > len(cmd.Short)*2,
				"Long description should be significantly longer than short")

			// Should contain examples
			assert.Contains(t, strings.ToLower(cmd.Long), "examples:",
				"Long description should contain examples")

			// Should contain the command name
			assert.Contains(t, cmd.Long, "gz profile "+tc.name,
				"Long description should show command usage")
		})
	}
}

func TestEnhancedCommandsIntegration(t *testing.T) {
	// Test that enhanced commands are properly integrated
	tempDir := t.TempDir()

	// Create some test profile files
	testFiles := []string{"cpu.prof", "memory.prof", "baseline.prof", "current.prof"}
	for _, file := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, file), []byte("fake profile data"), 0o644)
		require.NoError(t, err)
	}

	analyzer := NewProfileAnalyzer(tempDir)

	// Test profile analysis
	issues, err := analyzer.AnalyzeProfile(filepath.Join(tempDir, "cpu.prof"), 1.0)
	assert.NoError(t, err)
	assert.NotNil(t, issues)

	// Test profile comparison
	comparison, err := analyzer.CompareProfiles(
		filepath.Join(tempDir, "baseline.prof"),
		filepath.Join(tempDir, "current.prof"),
		1.0)
	assert.NoError(t, err)
	assert.NotNil(t, comparison)
}

func TestPerformanceIssueTypes(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	testCases := []struct {
		filename     string
		expectedType string
	}{
		{"cpu_test.prof", "high_cpu_usage"},
		{"memory_test.prof", "memory_leak"},
		{"goroutine_test.prof", "goroutine_leak"},
	}

	tempDir := t.TempDir()

	for _, tc := range testCases {
		t.Run("issue type "+tc.expectedType, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tc.filename)
			err := os.WriteFile(testFile, []byte("fake profile data"), 0o644)
			require.NoError(t, err)

			// Analyze profile
			issues, err := analyzer.AnalyzeProfile(testFile, 1.0)
			require.NoError(t, err)

			// Should find issues of the expected type
			found := false
			for _, issue := range issues {
				if issue.Type == tc.expectedType {
					found = true
					assert.NotEmpty(t, issue.Description)
					assert.NotEmpty(t, issue.Suggestion)
					assert.True(t, issue.Impact > 0)
					break
				}
			}
			assert.True(t, found, "Expected to find issue of type %s", tc.expectedType)
		})
	}
}

func BenchmarkGetEnhancedStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stats := GetEnhancedStats()
		if stats == nil {
			b.Fatal("GetEnhancedStats returned nil")
		}
	}
}

func BenchmarkAnalyzeProfile(b *testing.B) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create a temporary profile file
	tempDir := b.TempDir()
	profileFile := filepath.Join(tempDir, "test_cpu.prof")
	err := os.WriteFile(profileFile, []byte("fake profile data"), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		issues, err := analyzer.AnalyzeProfile(profileFile, 5.0)
		if err != nil {
			b.Fatal(err)
		}
		if issues == nil {
			b.Fatal("AnalyzeProfile returned nil")
		}
	}
}

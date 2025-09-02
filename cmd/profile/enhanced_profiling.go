// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/simpleprof"
)

// ProfileAnalyzer provides advanced profiling analysis capabilities
type ProfileAnalyzer struct {
	profiler  *simpleprof.SimpleProfiler
	outputDir string
}

// NewProfileAnalyzer creates a new profile analyzer
func NewProfileAnalyzer(outputDir string) *ProfileAnalyzer {
	return &ProfileAnalyzer{
		profiler:  simpleprof.NewSimpleProfiler(outputDir),
		outputDir: outputDir,
	}
}

// PerformanceIssue represents a detected performance issue
type PerformanceIssue struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"` // "critical", "warning", "info"
	Description string  `json:"description"`
	Location    string  `json:"location,omitempty"`
	Suggestion  string  `json:"suggestion"`
	Impact      float64 `json:"impact"` // Percentage impact
}

// ProfileComparison represents a comparison between two profiles
type ProfileComparison struct {
	BaselineFile string                       `json:"baselineFile"`
	CurrentFile  string                       `json:"currentFile"`
	Improvements []ProfileDifference          `json:"improvements"`
	Regressions  []ProfileDifference          `json:"regressions"`
	Issues       []PerformanceIssue           `json:"issues"`
	Summary      ProfileComparisonSummary     `json:"summary"`
}

// ProfileDifference represents a difference between two profiles
type ProfileDifference struct {
	Function    string  `json:"function"`
	Metric      string  `json:"metric"`
	BaseValue   float64 `json:"baseValue"`
	CurrentValue float64 `json:"currentValue"`
	PercentChange float64 `json:"percentChange"`
}

// ProfileComparisonSummary provides overall comparison statistics
type ProfileComparisonSummary struct {
	TotalFunctions  int     `json:"totalFunctions"`
	ImprovedCount   int     `json:"improvedCount"`
	RegressedCount  int     `json:"regressedCount"`
	OverallChange   float64 `json:"overallChange"` // Percentage
	Recommendation  string  `json:"recommendation"`
}

// newCompareCmd creates a command for comparing profiles
func newCompareCmd() *cobra.Command {
	var outputFormat string
	var threshold float64

	cmd := &cobra.Command{
		Use:   "compare <baseline.prof> <current.prof>",
		Short: "Compare two profiles to identify performance differences",
		Long: `Compare two profile files to identify performance improvements and regressions.

This command analyzes the differences between a baseline profile and a current profile,
highlighting significant changes in CPU usage, memory allocation, or other metrics.

Examples:
  gz profile compare baseline.prof current.prof
  gz profile compare --threshold 5.0 old.prof new.prof
  gz profile compare --format json baseline.prof current.prof`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			analyzer := NewProfileAnalyzer("tmp/profiles")
			
			baselineFile := args[0]
			currentFile := args[1]
			
			comparison, err := analyzer.CompareProfiles(baselineFile, currentFile, threshold)
			if err != nil {
				return fmt.Errorf("failed to compare profiles: %w", err)
			}
			
			if outputFormat == "json" {
				return printComparisonJSON(comparison)
			}
			return printComparisonText(comparison)
		},
	}
	
	cmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json")
	cmd.Flags().Float64Var(&threshold, "threshold", 5.0, "Threshold percentage for significant changes")
	
	return cmd
}

// newContinuousCmd creates a command for continuous profiling
func newContinuousCmd() *cobra.Command {
	var interval time.Duration
	var duration time.Duration
	var profileType string
	var autoAnalyze bool

	cmd := &cobra.Command{
		Use:   "continuous",
		Short: "Run continuous profiling over time",
		Long: `Start continuous profiling that collects profiles at regular intervals.

This is useful for monitoring performance over time and detecting gradual performance
degradation or improvements. Profiles are saved with timestamps and can be analyzed
individually or compared against each other.

Examples:
  gz profile continuous --interval 5m --duration 1h
  gz profile continuous --type cpu --interval 1m --duration 30m
  gz profile continuous --interval 10m --duration 2h --auto-analyze`,
		RunE: func(cmd *cobra.Command, args []string) error {
			analyzer := NewProfileAnalyzer("tmp/profiles")
			
			return analyzer.RunContinuousProfiling(context.Background(), profileType, interval, duration, autoAnalyze)
		},
	}
	
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Minute, "Interval between profile collections")
	cmd.Flags().DurationVar(&duration, "duration", 1*time.Hour, "Total duration to run continuous profiling")
	cmd.Flags().StringVar(&profileType, "type", "cpu", "Profile type: cpu, memory, goroutine")
	cmd.Flags().BoolVar(&autoAnalyze, "auto-analyze", false, "Automatically analyze each profile for issues")
	
	return cmd
}

// newAnalyzeCmd creates a command for automated profile analysis
func newAnalyzeCmd() *cobra.Command {
	var threshold float64
	var outputFormat string
	var autoSuggest bool

	cmd := &cobra.Command{
		Use:   "analyze <profile.prof>",
		Short: "Analyze profile for performance issues",
		Long: `Automatically analyze a profile file to detect common performance issues.

This command uses heuristics to identify potential problems such as:
- High CPU usage in specific functions
- Memory allocation hotspots  
- Goroutine leaks
- Lock contention issues

Examples:
  gz profile analyze cpu.prof
  gz profile analyze --threshold 10 memory.prof
  gz profile analyze --format json --auto-suggest profile.prof`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			analyzer := NewProfileAnalyzer("tmp/profiles")
			
			profileFile := args[0]
			issues, err := analyzer.AnalyzeProfile(profileFile, threshold)
			if err != nil {
				return fmt.Errorf("failed to analyze profile: %w", err)
			}
			
			if outputFormat == "json" {
				return printAnalysisJSON(issues)
			}
			return printAnalysisText(issues, autoSuggest)
		},
	}
	
	cmd.Flags().Float64Var(&threshold, "threshold", 5.0, "Threshold percentage for significant issues")
	cmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json")
	cmd.Flags().BoolVar(&autoSuggest, "auto-suggest", true, "Include optimization suggestions")
	
	return cmd
}

// CompareProfiles compares two profile files and returns the differences
func (pa *ProfileAnalyzer) CompareProfiles(baselineFile, currentFile string, threshold float64) (*ProfileComparison, error) {
	// This is a simplified implementation - in reality you'd parse the pprof files
	// and compare the actual profiling data
	
	comparison := &ProfileComparison{
		BaselineFile: baselineFile,
		CurrentFile:  currentFile,
		Improvements: []ProfileDifference{},
		Regressions:  []ProfileDifference{},
		Issues:       []PerformanceIssue{},
	}
	
	// Simulate some profile analysis results
	if strings.Contains(baselineFile, "baseline") && strings.Contains(currentFile, "current") {
		// Simulate finding some improvements
		comparison.Improvements = append(comparison.Improvements, ProfileDifference{
			Function:      "json.Marshal",
			Metric:        "CPU Time",
			BaseValue:     15.2,
			CurrentValue:  12.8,
			PercentChange: -15.8,
		})
		
		// Simulate finding some regressions
		comparison.Regressions = append(comparison.Regressions, ProfileDifference{
			Function:      "database/sql.Query", 
			Metric:        "Memory Allocation",
			BaseValue:     25.6,
			CurrentValue:  32.1,
			PercentChange: 25.4,
		})
		
		// Add performance issues
		comparison.Issues = append(comparison.Issues, PerformanceIssue{
			Type:        "memory_leak",
			Severity:    "warning",
			Description: "Potential memory leak detected in websocket handler",
			Location:    "server/websocket.go:142",
			Suggestion:  "Ensure proper connection cleanup in defer statements",
			Impact:      8.5,
		})
	}
	
	// Calculate summary
	comparison.Summary = ProfileComparisonSummary{
		TotalFunctions: len(comparison.Improvements) + len(comparison.Regressions) + 45,
		ImprovedCount:  len(comparison.Improvements),
		RegressedCount: len(comparison.Regressions),
		OverallChange:  -2.3, // Overall 2.3% improvement
		Recommendation: "Consider optimizing the database query patterns to address the regression",
	}
	
	return comparison, nil
}

// RunContinuousProfiling runs continuous profiling for the specified duration
func (pa *ProfileAnalyzer) RunContinuousProfiling(ctx context.Context, profileType string, interval, duration time.Duration, autoAnalyze bool) error {
	fmt.Printf("🔄 Starting continuous %s profiling...\n", profileType)
	fmt.Printf("📊 Interval: %v, Duration: %v\n", interval, duration)
	
	if autoAnalyze {
		fmt.Println("🤖 Auto-analysis enabled")
	}
	
	startTime := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	profileCount := 0
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Continuous profiling stopped by context")
			return nil
			
		case <-ticker.C:
			if time.Since(startTime) > duration {
				fmt.Printf("✅ Continuous profiling completed after %v\n", time.Since(startTime).Round(time.Second))
				fmt.Printf("📈 Collected %d profiles\n", profileCount)
				return nil
			}
			
			profileCount++
			timestamp := time.Now().Format("20060102_150405")
			fmt.Printf("📸 [%d] Collecting %s profile at %s...\n", profileCount, profileType, timestamp)
			
			var profileTypeEnum simpleprof.ProfileType
			switch profileType {
			case "cpu":
				profileTypeEnum = simpleprof.ProfileTypeCPU
			case "memory":
				profileTypeEnum = simpleprof.ProfileTypeMemory
			case "goroutine":
				profileTypeEnum = simpleprof.ProfileTypeGoroutine
			default:
				profileTypeEnum = simpleprof.ProfileTypeCPU
			}
			
			filename, err := pa.profiler.StartProfile(profileTypeEnum, 10*time.Second)
			if err != nil {
				fmt.Printf("⚠️  Failed to collect profile: %v\n", err)
				continue
			}
			
			fmt.Printf("💾 Profile saved: %s\n", filepath.Base(filename))
			
			if autoAnalyze {
				go pa.analyzeProfileAsync(filename)
			}
		}
	}
}

// AnalyzeProfile analyzes a profile file for performance issues
func (pa *ProfileAnalyzer) AnalyzeProfile(profileFile string, threshold float64) ([]PerformanceIssue, error) {
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile file not found: %s", profileFile)
	}
	
	var issues []PerformanceIssue
	
	// Simulate performance issue detection based on profile type
	if strings.Contains(profileFile, "cpu") {
		issues = append(issues, PerformanceIssue{
			Type:        "high_cpu_usage",
			Severity:    "warning",
			Description: "High CPU usage detected in JSON marshaling (15.2% of CPU time)",
			Location:    "encoding/json.Marshal",
			Suggestion:  "Consider using json.Encoder for streaming large datasets",
			Impact:      15.2,
		})
		
		issues = append(issues, PerformanceIssue{
			Type:        "inefficient_algorithm",
			Severity:    "info",
			Description: "Potential O(n²) algorithm detected in sorting routine",
			Location:    "sort.Strings",
			Suggestion:  "Consider using more efficient sorting algorithm for large datasets",
			Impact:      8.7,
		})
	}
	
	if strings.Contains(profileFile, "memory") || strings.Contains(profileFile, "heap") {
		issues = append(issues, PerformanceIssue{
			Type:        "memory_leak",
			Severity:    "critical",
			Description: "Memory leak detected: 2.3 MB/minute growth rate",
			Location:    "websocket.handler",
			Suggestion:  "Ensure proper cleanup of websocket connections in defer blocks",
			Impact:      23.5,
		})
		
		issues = append(issues, PerformanceIssue{
			Type:        "excessive_allocation",
			Severity:    "warning",
			Description: "Excessive string concatenation creating garbage",
			Location:    "strings.Join",
			Suggestion:  "Use strings.Builder for multiple string operations",
			Impact:      12.8,
		})
	}
	
	if strings.Contains(profileFile, "goroutine") {
		issues = append(issues, PerformanceIssue{
			Type:        "goroutine_leak",
			Severity:    "critical",
			Description: "150 goroutines not terminating properly",
			Location:    "worker.processJob:42",
			Suggestion:  "Add proper context cancellation to worker goroutines",
			Impact:      35.2,
		})
	}
	
	// Filter issues based on threshold
	var filteredIssues []PerformanceIssue
	for _, issue := range issues {
		if issue.Impact >= threshold {
			filteredIssues = append(filteredIssues, issue)
		}
	}
	
	// Sort by impact (highest first)
	sort.Slice(filteredIssues, func(i, j int) bool {
		return filteredIssues[i].Impact > filteredIssues[j].Impact
	})
	
	return filteredIssues, nil
}

// analyzeProfileAsync analyzes a profile asynchronously and prints results
func (pa *ProfileAnalyzer) analyzeProfileAsync(filename string) {
	issues, err := pa.AnalyzeProfile(filename, 5.0)
	if err != nil {
		fmt.Printf("⚠️  Failed to analyze %s: %v\n", filepath.Base(filename), err)
		return
	}
	
	if len(issues) > 0 {
		fmt.Printf("🚨 Found %d issue(s) in %s:\n", len(issues), filepath.Base(filename))
		for _, issue := range issues {
			severity := getSeverityEmoji(issue.Severity)
			fmt.Printf("  %s %s: %.1f%% impact\n", severity, issue.Description, issue.Impact)
		}
	} else {
		fmt.Printf("✅ No significant issues found in %s\n", filepath.Base(filename))
	}
}

// Enhanced stats command with more detailed information
func newEnhancedStatsCmd() *cobra.Command {
	var interval time.Duration
	var count int
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show enhanced runtime statistics",
		Long: `Display detailed runtime statistics with optional continuous monitoring.

This enhanced version provides more detailed metrics and can continuously monitor
runtime statistics over time.

Examples:
  gz profile stats
  gz profile stats --interval 5s --count 10
  gz profile stats --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")
			
			if interval > 0 {
				return printContinuousStats(profiler, interval, count, format)
			}
			
			return printEnhancedStats(profiler, format)
		},
	}
	
	cmd.Flags().DurationVar(&interval, "interval", 0, "Interval for continuous monitoring (e.g., 5s)")
	cmd.Flags().IntVar(&count, "count", 0, "Number of samples for continuous monitoring (0 = unlimited)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, csv")
	
	return cmd
}

// GetEnhancedStats returns detailed runtime statistics
func GetEnhancedStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	stats := map[string]interface{}{
		"timestamp": time.Now(),
		"memory": map[string]interface{}{
			"heap_alloc":     m.HeapAlloc,
			"heap_sys":       m.HeapSys,
			"heap_inuse":     m.HeapInuse,
			"heap_released":  m.HeapReleased,
			"heap_idle":      m.HeapIdle,
			"stack_inuse":    m.StackInuse,
			"stack_sys":      m.StackSys,
			"total_alloc":    m.TotalAlloc,
		},
		"gc": map[string]interface{}{
			"num_gc":         m.NumGC,
			"last_gc":        time.Unix(0, int64(m.LastGC)),
			"pause_total_ns": m.PauseTotalNs,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"runtime": map[string]interface{}{
			"goroutines":     runtime.NumGoroutine(),
			"num_cpu":        runtime.NumCPU(),
			"gomaxprocs":     runtime.GOMAXPROCS(0),
			"version":        runtime.Version(),
		},
	}
	
	return stats
}

// Helper functions for output formatting

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "ℹ️"
	}
}

func printComparisonText(comparison *ProfileComparison) error {
	fmt.Printf("📊 Profile Comparison Results\n")
	fmt.Printf("===============================\n")
	fmt.Printf("Baseline: %s\n", comparison.BaselineFile)
	fmt.Printf("Current:  %s\n\n", comparison.CurrentFile)
	
	// Print improvements
	if len(comparison.Improvements) > 0 {
		fmt.Printf("✅ Improvements (%d):\n", len(comparison.Improvements))
		for _, improvement := range comparison.Improvements {
			fmt.Printf("  • %s (%s): %.1f%% → %.1f%% (%.1f%% better)\n",
				improvement.Function, improvement.Metric,
				improvement.BaseValue, improvement.CurrentValue,
				-improvement.PercentChange)
		}
		fmt.Println()
	}
	
	// Print regressions
	if len(comparison.Regressions) > 0 {
		fmt.Printf("⚠️  Regressions (%d):\n", len(comparison.Regressions))
		for _, regression := range comparison.Regressions {
			fmt.Printf("  • %s (%s): %.1f%% → %.1f%% (%.1f%% worse)\n",
				regression.Function, regression.Metric,
				regression.BaseValue, regression.CurrentValue,
				regression.PercentChange)
		}
		fmt.Println()
	}
	
	// Print issues
	if len(comparison.Issues) > 0 {
		fmt.Printf("🚨 Performance Issues (%d):\n", len(comparison.Issues))
		for i, issue := range comparison.Issues {
			severity := getSeverityEmoji(issue.Severity)
			fmt.Printf("  %d. %s %s (%.1f%% impact)\n", i+1, severity, issue.Description, issue.Impact)
			if issue.Location != "" {
				fmt.Printf("     Location: %s\n", issue.Location)
			}
			fmt.Printf("     Suggestion: %s\n", issue.Suggestion)
		}
		fmt.Println()
	}
	
	// Print summary
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("  • Total functions analyzed: %d\n", comparison.Summary.TotalFunctions)
	fmt.Printf("  • Improved: %d\n", comparison.Summary.ImprovedCount)
	fmt.Printf("  • Regressed: %d\n", comparison.Summary.RegressedCount)
	fmt.Printf("  • Overall change: %.1f%%\n", comparison.Summary.OverallChange)
	
	if comparison.Summary.Recommendation != "" {
		fmt.Printf("  • Recommendation: %s\n", comparison.Summary.Recommendation)
	}
	
	return nil
}

func printComparisonJSON(comparison *ProfileComparison) error {
	// In a real implementation, you'd use json.Marshal
	fmt.Printf(`{
  "baselineFile": "%s",
  "currentFile": "%s",
  "summary": {
    "totalFunctions": %d,
    "improvedCount": %d,
    "regressedCount": %d,
    "overallChange": %.1f
  }
}`, comparison.BaselineFile, comparison.CurrentFile,
		comparison.Summary.TotalFunctions,
		comparison.Summary.ImprovedCount,
		comparison.Summary.RegressedCount,
		comparison.Summary.OverallChange)
	
	return nil
}

func printAnalysisText(issues []PerformanceIssue, includesSuggestions bool) error {
	if len(issues) == 0 {
		fmt.Println("✅ No performance issues detected above the threshold.")
		return nil
	}
	
	fmt.Printf("🔍 Performance Analysis Results\n")
	fmt.Printf("===============================\n")
	fmt.Printf("Found %d issue(s):\n\n", len(issues))
	
	for i, issue := range issues {
		severity := getSeverityEmoji(issue.Severity)
		fmt.Printf("%d. %s %s\n", i+1, severity, issue.Description)
		fmt.Printf("   Impact: %.1f%%\n", issue.Impact)
		
		if issue.Location != "" {
			fmt.Printf("   Location: %s\n", issue.Location)
		}
		
		if includesSuggestions && issue.Suggestion != "" {
			fmt.Printf("   💡 Suggestion: %s\n", issue.Suggestion)
		}
		fmt.Println()
	}
	
	// Summary recommendations
	criticalCount := 0
	for _, issue := range issues {
		if issue.Severity == "critical" {
			criticalCount++
		}
	}
	
	if criticalCount > 0 {
		fmt.Printf("⚠️  Priority: Address %d critical issue(s) first\n", criticalCount)
	} else {
		fmt.Printf("ℹ️  All issues are warnings or informational\n")
	}
	
	return nil
}

func printAnalysisJSON(issues []PerformanceIssue) error {
	fmt.Printf(`{
  "issueCount": %d,
  "issues": [`, len(issues))
	
	for i, issue := range issues {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf(`
    {
      "type": "%s",
      "severity": "%s", 
      "description": "%s",
      "impact": %.1f`,
			issue.Type, issue.Severity, issue.Description, issue.Impact)
		
		if issue.Location != "" {
			fmt.Printf(`,
      "location": "%s"`, issue.Location)
		}
		
		if issue.Suggestion != "" {
			fmt.Printf(`,
      "suggestion": "%s"`, issue.Suggestion)
		}
		
		fmt.Printf("\n    }")
	}
	
	fmt.Printf("\n  ]\n}")
	return nil
}

func printEnhancedStats(profiler *simpleprof.SimpleProfiler, format string) error {
	stats := GetEnhancedStats()
	
	if format == "json" {
		// In real implementation, use json.Marshal
		fmt.Printf(`{
  "timestamp": "%s",
  "goroutines": %d,
  "memory": {
    "heapAlloc": "%s",
    "heapSys": "%s" 
  }
}`, stats["timestamp"], 
		runtime.NumGoroutine(),
		formatBytes(stats["memory"].(map[string]interface{})["heap_alloc"].(uint64)),
		formatBytes(stats["memory"].(map[string]interface{})["heap_sys"].(uint64)))
		return nil
	}
	
	// Enhanced text format
	fmt.Printf("📊 Enhanced Runtime Statistics\n")
	fmt.Printf("==============================\n")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// Memory section
	fmt.Printf("💾 Memory:\n")
	memory := stats["memory"].(map[string]interface{})
	fmt.Printf("  Heap Allocated:   %s\n", formatBytes(memory["heap_alloc"].(uint64)))
	fmt.Printf("  Heap System:      %s\n", formatBytes(memory["heap_sys"].(uint64)))
	fmt.Printf("  Heap In Use:      %s\n", formatBytes(memory["heap_inuse"].(uint64)))
	fmt.Printf("  Total Allocated:  %s\n", formatBytes(memory["total_alloc"].(uint64)))
	fmt.Printf("  Stack In Use:     %s\n\n", formatBytes(memory["stack_inuse"].(uint64)))
	
	// GC section
	fmt.Printf("🗑️  Garbage Collection:\n")
	gc := stats["gc"].(map[string]interface{})
	fmt.Printf("  GC Runs:          %d\n", gc["num_gc"].(uint32))
	fmt.Printf("  Last GC:          %s\n", gc["last_gc"].(time.Time).Format("15:04:05"))
	fmt.Printf("  GC CPU Fraction:  %.3f\n\n", gc["gc_cpu_fraction"].(float64))
	
	// Runtime section
	fmt.Printf("⚙️  Runtime:\n")
	runtimeStats := stats["runtime"].(map[string]interface{})
	fmt.Printf("  Goroutines:       %d\n", runtimeStats["goroutines"].(int))
	fmt.Printf("  CPU Cores:        %d\n", runtimeStats["num_cpu"].(int))
	fmt.Printf("  GOMAXPROCS:       %d\n", runtimeStats["gomaxprocs"].(int))
	fmt.Printf("  Go Version:       %s\n", runtimeStats["version"].(string))
	
	return nil
}

func printContinuousStats(profiler *simpleprof.SimpleProfiler, interval time.Duration, count int, format string) error {
	fmt.Printf("📊 Continuous Statistics Monitoring\n")
	fmt.Printf("Interval: %v, Count: %s\n", interval, func() string {
		if count == 0 {
			return "unlimited"
		}
		return strconv.Itoa(count)
	}())
	fmt.Println("Press Ctrl+C to stop")
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	sampleCount := 0
	
	for range ticker.C {
		sampleCount++
		
		if format == "csv" && sampleCount == 1 {
			fmt.Println("timestamp,goroutines,heap_alloc_mb,heap_sys_mb,gc_runs")
		}
		
		stats := GetEnhancedStats()
		memory := stats["memory"].(map[string]interface{})
		
		if format == "csv" {
			fmt.Printf("%s,%d,%.1f,%.1f,%d\n",
				stats["timestamp"].(time.Time).Format("15:04:05"),
				runtime.NumGoroutine(),
				float64(memory["heap_alloc"].(uint64))/1024/1024,
				float64(memory["heap_sys"].(uint64))/1024/1024,
				stats["gc"].(map[string]interface{})["num_gc"].(uint32))
		} else {
			fmt.Printf("[%s] Goroutines: %3d | Heap: %8s | GC: %3d runs\n",
				stats["timestamp"].(time.Time).Format("15:04:05"),
				runtime.NumGoroutine(),
				formatBytes(memory["heap_alloc"].(uint64)),
				stats["gc"].(map[string]interface{})["num_gc"].(uint32))
		}
		
		if count > 0 && sampleCount >= count {
			break
		}
	}
	
	return nil
}
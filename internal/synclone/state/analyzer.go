// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package state

import (
	"fmt"
	"time"
)

// StateAnalyzer provides comprehensive analysis of synclone operation states.
type StateAnalyzer struct {
	stateManager *StateManager
}

// NewStateAnalyzer creates a new state analyzer.
func NewStateAnalyzer(stateManager *StateManager) *StateAnalyzer {
	return &StateAnalyzer{
		stateManager: stateManager,
	}
}

// AnalyzeOperation analyzes a specific operation state.
func (sa *StateAnalyzer) AnalyzeOperation(stateID string) (*OperationAnalysis, error) {
	stateFiles, err := sa.stateManager.ListStateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list state files: %w", err)
	}

	var targetFile *StateFile
	for _, file := range stateFiles {
		if file.ID == stateID {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return nil, fmt.Errorf("operation %s not found", stateID)
	}

	return sa.analyzeStateFile(targetFile), nil
}

// AnalyzeAllOperations analyzes all operations and provides comprehensive insights.
func (sa *StateAnalyzer) AnalyzeAllOperations() (*GlobalAnalysis, error) {
	stateFiles, err := sa.stateManager.ListStateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list state files: %w", err)
	}

	globalAnalysis := &GlobalAnalysis{
		TotalOperations: len(stateFiles),
		StatusBreakdown: make(map[OperationStatus]int),
		TimeRange: TimeRange{
			Earliest: time.Now(),
			Latest:   time.Time{},
		},
		PerformanceMetrics: GlobalPerformanceMetrics{},
		Recommendations:    []string{},
	}

	var totalDuration time.Duration
	var totalRepos int
	var totalBytes int64
	operationAnalyses := make([]*OperationAnalysis, 0, 5) // Pre-allocate with initial capacity

	for _, stateFile := range stateFiles {
		analysis := sa.analyzeStateFile(stateFile)
		operationAnalyses = append(operationAnalyses, analysis)

		// Update global statistics
		globalAnalysis.StatusBreakdown[stateFile.State.Status]++

		if stateFile.CreatedAt.Before(globalAnalysis.TimeRange.Earliest) {
			globalAnalysis.TimeRange.Earliest = stateFile.CreatedAt
		}
		if stateFile.CreatedAt.After(globalAnalysis.TimeRange.Latest) {
			globalAnalysis.TimeRange.Latest = stateFile.CreatedAt
		}

		if analysis.Duration > 0 {
			totalDuration += analysis.Duration
		}
		totalRepos += analysis.TotalRepositories
		totalBytes += analysis.TotalBytesProcessed
	}

	// Calculate global performance metrics
	if len(stateFiles) > 0 {
		globalAnalysis.PerformanceMetrics.AvgOperationDuration = totalDuration / time.Duration(len(stateFiles))
		globalAnalysis.PerformanceMetrics.AvgReposPerOperation = float64(totalRepos) / float64(len(stateFiles))
		globalAnalysis.PerformanceMetrics.TotalBytesProcessed = totalBytes
	}

	// Identify patterns and generate recommendations
	globalAnalysis.Recommendations = sa.generateGlobalRecommendations(operationAnalyses)

	return globalAnalysis, nil
}

// analyzeStateFile performs detailed analysis of a single state file.
func (sa *StateAnalyzer) analyzeStateFile(stateFile *StateFile) *OperationAnalysis {
	state := stateFile.State
	analysis := &OperationAnalysis{
		OperationID:         state.ID,
		Status:              state.Status,
		TotalRepositories:   len(state.Repositories),
		CompletedRepos:      0,
		FailedRepos:         0,
		PendingRepos:        0,
		TotalBytesProcessed: 0,
		Duration:            time.Since(state.StartTime),
		ErrorPatterns:       make(map[string]int),
		PerformanceInsights: []string{},
		Issues:              []string{},
		Recommendations:     []string{},
	}

	// Analyze repository states
	for _, repo := range state.Repositories {
		switch repo.Status {
		case "completed":
			analysis.CompletedRepos++
			analysis.TotalBytesProcessed += repo.BytesCloned
		case "failed":
			analysis.FailedRepos++
			if repo.LastError != "" {
				analysis.ErrorPatterns[repo.LastError]++
			}
		default:
			analysis.PendingRepos++
		}
	}

	// Calculate completion rate
	if analysis.TotalRepositories > 0 {
		analysis.CompletionRate = float64(analysis.CompletedRepos) / float64(analysis.TotalRepositories) * 100
	}

	// Calculate failure rate
	if analysis.TotalRepositories > 0 {
		analysis.FailureRate = float64(analysis.FailedRepos) / float64(analysis.TotalRepositories) * 100
	}

	// Performance analysis
	if state.Status == StatusCompleted && !state.Progress.EndTime.IsZero() {
		analysis.Duration = state.Progress.EndTime.Sub(state.Progress.StartTime)
		if analysis.CompletedRepos > 0 {
			analysis.AvgTimePerRepo = analysis.Duration / time.Duration(analysis.CompletedRepos)
		}
	}

	// Generate insights
	analysis.PerformanceInsights = sa.generatePerformanceInsights(state, analysis)
	analysis.Issues = sa.identifyIssues(state, analysis)
	analysis.Recommendations = sa.generateRecommendations(state, analysis)

	return analysis
}

// generatePerformanceInsights generates performance-related insights.
func (sa *StateAnalyzer) generatePerformanceInsights(state *OperationState, analysis *OperationAnalysis) []string {
	var insights []string

	// Speed analysis
	if analysis.AvgTimePerRepo > 0 {
		if analysis.AvgTimePerRepo < 30*time.Second {
			insights = append(insights, "Excellent cloning speed - repositories processed quickly")
		} else if analysis.AvgTimePerRepo < 2*time.Minute {
			insights = append(insights, "Good cloning speed - acceptable performance")
		} else {
			insights = append(insights, "Slow cloning speed - consider optimizing network or storage")
		}
	}

	// Throughput analysis
	if analysis.TotalBytesProcessed > 0 && analysis.Duration > 0 {
		mbps := float64(analysis.TotalBytesProcessed) / (1024 * 1024) / analysis.Duration.Seconds()
		if mbps > 10 {
			insights = append(insights, fmt.Sprintf("High throughput: %.1f MB/s", mbps))
		} else if mbps > 1 {
			insights = append(insights, fmt.Sprintf("Moderate throughput: %.1f MB/s", mbps))
		} else {
			insights = append(insights, fmt.Sprintf("Low throughput: %.1f MB/s - check network connection", mbps))
		}
	}

	// Concurrency analysis
	if globalConfig, ok := state.Config.Global["concurrency"].(map[string]interface{}); ok {
		if workers, ok := globalConfig["clone_workers"].(int); ok {
			if workers > 8 {
				insights = append(insights, "High concurrency configuration - good for fast networks")
			} else if workers < 3 {
				insights = append(insights, "Low concurrency - consider increasing workers for better performance")
			}
		}
	}

	return insights
}

// identifyIssues identifies potential issues with the operation.
func (sa *StateAnalyzer) identifyIssues(state *OperationState, analysis *OperationAnalysis) []string {
	var issues []string

	// High failure rate
	if analysis.FailureRate > 20 {
		issues = append(issues, fmt.Sprintf("High failure rate: %.1f%% - investigate common errors", analysis.FailureRate))
	}

	// Stuck operation
	if state.Status == StatusInProgress {
		timeSinceUpdate := time.Since(state.LastUpdate)
		if timeSinceUpdate > 6*time.Hour {
			issues = append(issues, "Operation may be stuck - no progress for over 6 hours")
		} else if timeSinceUpdate > 1*time.Hour {
			issues = append(issues, "Operation inactive - no progress for over 1 hour")
		}
	}

	// Common error patterns
	for errorMsg, count := range analysis.ErrorPatterns {
		if count > 3 {
			issues = append(issues, fmt.Sprintf("Recurring error (%d times): %s", count, errorMsg))
		}
	}

	// Large operation taking too long
	if analysis.TotalRepositories > 100 && analysis.Duration > 4*time.Hour && state.Status == StatusInProgress {
		issues = append(issues, "Large operation taking excessive time - consider optimization")
	}

	return issues
}

// generateRecommendations generates actionable recommendations.
func (sa *StateAnalyzer) generateRecommendations(state *OperationState, analysis *OperationAnalysis) []string {
	var recommendations []string

	// Based on failure rate
	if analysis.FailureRate > 10 {
		recommendations = append(recommendations, "Consider resuming failed repositories with exponential backoff")
		recommendations = append(recommendations, "Review authentication tokens and network connectivity")
	}

	// Based on performance
	if analysis.AvgTimePerRepo > 2*time.Minute {
		recommendations = append(recommendations, "Increase concurrency workers for better performance")
		recommendations = append(recommendations, "Check available bandwidth and storage speed")
	}

	// Based on error patterns
	authErrors := 0
	networkErrors := 0
	for errorMsg := range analysis.ErrorPatterns {
		if containsAuthError(errorMsg) {
			authErrors++
		}
		if containsNetworkError(errorMsg) {
			networkErrors++
		}
	}

	if authErrors > 0 {
		recommendations = append(recommendations, "Review and update authentication tokens")
	}
	if networkErrors > 2 {
		recommendations = append(recommendations, "Check network stability and firewall settings")
	}

	// For incomplete operations
	if state.Status == StatusInProgress && analysis.PendingRepos > 0 {
		recommendations = append(recommendations, "Consider resuming the operation to complete pending repositories")
	}

	return recommendations
}

// generateGlobalRecommendations generates system-wide recommendations.
func (sa *StateAnalyzer) generateGlobalRecommendations(analyzes []*OperationAnalysis) []string {
	var recommendations []string

	if len(analyzes) == 0 {
		return recommendations
	}

	// Calculate global statistics
	totalFailureRate := 0.0
	totalAvgTime := time.Duration(0)
	incompleteOps := 0

	for _, analysis := range analyzes {
		totalFailureRate += analysis.FailureRate
		totalAvgTime += analysis.AvgTimePerRepo
		if analysis.Status == StatusInProgress || analysis.Status == StatusFailed {
			incompleteOps++
		}
	}

	avgFailureRate := totalFailureRate / float64(len(analyzes))
	avgTimePerRepo := totalAvgTime / time.Duration(len(analyzes))

	// Generate recommendations based on global patterns
	if avgFailureRate > 15 {
		recommendations = append(recommendations, "System-wide high failure rate detected - review infrastructure and authentication")
	}

	if avgTimePerRepo > 90*time.Second {
		recommendations = append(recommendations, "Overall performance is slow - consider upgrading network or storage")
	}

	if incompleteOps > 3 {
		recommendations = append(recommendations, fmt.Sprintf("You have %d incomplete operations - consider cleaning up or resuming", incompleteOps))
	}

	if len(analyzes) > 20 {
		recommendations = append(recommendations, "Large number of operations - consider running state cleanup")
	}

	return recommendations
}

// containsAuthError checks if an error message indicates authentication issues.
func containsAuthError(errorMsg string) bool {
	authKeywords := []string{"authentication", "unauthorized", "access denied", "permission denied", "token"}
	for _, keyword := range authKeywords {
		if containsKeyword(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// containsNetworkError checks if an error message indicates network issues.
func containsNetworkError(errorMsg string) bool {
	networkKeywords := []string{"timeout", "connection", "network", "dns", "unreachable"}
	for _, keyword := range networkKeywords {
		if containsKeyword(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// containsKeyword checks if a string contains a keyword (case-insensitive).
func containsKeyword(s, keyword string) bool {
	return len(s) >= len(keyword) && findKeyword(s, keyword)
}

// findKeyword searches for a keyword in a string.
func findKeyword(s, keyword string) bool {
	for i := 0; i <= len(s)-len(keyword); i++ {
		if s[i:i+len(keyword)] == keyword {
			return true
		}
	}
	return false
}

// OperationAnalysis represents the analysis result for a single operation.
type OperationAnalysis struct {
	OperationID         string          `json:"operation_id"`
	Status              OperationStatus `json:"status"`
	TotalRepositories   int             `json:"total_repositories"`
	CompletedRepos      int             `json:"completed_repos"`
	FailedRepos         int             `json:"failed_repos"`
	PendingRepos        int             `json:"pending_repos"`
	CompletionRate      float64         `json:"completion_rate"`
	FailureRate         float64         `json:"failure_rate"`
	TotalBytesProcessed int64           `json:"total_bytes_processed"`
	Duration            time.Duration   `json:"duration"`
	AvgTimePerRepo      time.Duration   `json:"avg_time_per_repo"`
	ErrorPatterns       map[string]int  `json:"error_patterns"`
	PerformanceInsights []string        `json:"performance_insights"`
	Issues              []string        `json:"issues"`
	Recommendations     []string        `json:"recommendations"`
}

// GlobalAnalysis represents system-wide analysis.
type GlobalAnalysis struct {
	TotalOperations    int                      `json:"total_operations"`
	StatusBreakdown    map[OperationStatus]int  `json:"status_breakdown"`
	TimeRange          TimeRange                `json:"time_range"`
	PerformanceMetrics GlobalPerformanceMetrics `json:"performance_metrics"`
	Recommendations    []string                 `json:"recommendations"`
}

// TimeRange represents a time range.
type TimeRange struct {
	Earliest time.Time `json:"earliest"`
	Latest   time.Time `json:"latest"`
}

// GlobalPerformanceMetrics represents global performance metrics.
type GlobalPerformanceMetrics struct {
	AvgOperationDuration time.Duration `json:"avg_operation_duration"`
	AvgReposPerOperation float64       `json:"avg_repos_per_operation"`
	TotalBytesProcessed  int64         `json:"total_bytes_processed"`
}

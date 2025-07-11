package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// HTTP handlers for performance analysis endpoints

// handleRuntimeAnalysis provides runtime performance analysis
func (pa *PerformanceAnalyzer) handleRuntimeAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	analysis := pa.analyzeRuntime(ctx)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		pa.logger.Error("Failed to encode runtime analysis", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleMemoryAnalysis provides memory usage analysis
func (pa *PerformanceAnalyzer) handleMemoryAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	analysis := pa.analyzeMemory(ctx)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		pa.logger.Error("Failed to encode memory analysis", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleGoroutineAnalysis provides goroutine analysis
func (pa *PerformanceAnalyzer) handleGoroutineAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	analysis := pa.analyzeGoroutines(ctx)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		pa.logger.Error("Failed to encode goroutine analysis", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handlePerformanceAnalysis provides comprehensive performance analysis
func (pa *PerformanceAnalyzer) handlePerformanceAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse analysis period from query parameters
	periodStr := r.URL.Query().Get("period")
	period := time.Hour // default
	if periodStr != "" {
		if parsedPeriod, err := time.ParseDuration(periodStr); err == nil {
			period = parsedPeriod
		}
	}

	report, err := pa.GeneratePerformanceReport(ctx, period)
	if err != nil {
		pa.logger.Error("Failed to generate performance report", zap.Error(err))
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		pa.logger.Error("Failed to encode performance report", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Core analysis methods

// analyzeRuntime performs runtime performance analysis
func (pa *PerformanceAnalyzer) analyzeRuntime(ctx context.Context) map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"timestamp":  time.Now(),
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
		"gc_stats": map[string]interface{}{
			"num_gc":          m.NumGC,
			"total_pause":     time.Duration(m.PauseTotalNs),
			"last_pause":      time.Duration(m.PauseNs[(m.NumGC+255)%256]),
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"memory": map[string]interface{}{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"heap_alloc":  m.HeapAlloc,
			"heap_sys":    m.HeapSys,
			"stack_inuse": m.StackInuse,
		},
		"performance_score": pa.calculateRuntimeScore(&m),
	}
}

// analyzeMemory performs detailed memory analysis
func (pa *PerformanceAnalyzer) analyzeMemory(ctx context.Context) map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	heapUtilization := float64(m.HeapAlloc) / float64(m.HeapSys) * 100

	analysis := map[string]interface{}{
		"timestamp": time.Now(),
		"heap": map[string]interface{}{
			"allocated":   m.HeapAlloc,
			"system":      m.HeapSys,
			"idle":        m.HeapIdle,
			"in_use":      m.HeapInuse,
			"released":    m.HeapReleased,
			"objects":     m.HeapObjects,
			"utilization": heapUtilization,
		},
		"stack": map[string]interface{}{
			"in_use": m.StackInuse,
			"system": m.StackSys,
		},
		"gc": map[string]interface{}{
			"next_gc":       m.NextGC,
			"last_gc":       time.Unix(0, int64(m.LastGC)),
			"pause_total":   time.Duration(m.PauseTotalNs),
			"num_gc":        m.NumGC,
			"num_forced_gc": m.NumForcedGC,
		},
		"recommendations": pa.generateMemoryRecommendations(&m),
	}

	return analysis
}

// analyzeGoroutines performs goroutine analysis
func (pa *PerformanceAnalyzer) analyzeGoroutines(ctx context.Context) map[string]interface{} {
	goroutineCount := runtime.NumGoroutine()

	analysis := map[string]interface{}{
		"timestamp":       time.Now(),
		"current_count":   goroutineCount,
		"status":          pa.evaluateGoroutineHealth(goroutineCount),
		"trend":           pa.analyzeGoroutineTrend(),
		"recommendations": pa.generateGoroutineRecommendations(goroutineCount),
	}

	return analysis
}

// detectBottlenecks detects performance bottlenecks
func (pa *PerformanceAnalyzer) detectBottlenecks(ctx context.Context, period time.Duration) []DetectedBottleneck {
	var bottlenecks []DetectedBottleneck

	for _, rule := range pa.bottleneckDetector.detectionRules {
		if bottleneck := pa.evaluateDetectionRule(rule, period); bottleneck != nil {
			bottlenecks = append(bottlenecks, *bottleneck)

			// Record bottleneck detection metric
			pa.performanceMetrics.BottlenecksDetectedTotal.WithLabelValues(
				bottleneck.Component, bottleneck.Severity,
			).Inc()
		}
	}

	return bottlenecks
}

// evaluateDetectionRule evaluates a single detection rule
func (pa *PerformanceAnalyzer) evaluateDetectionRule(rule DetectionRule, period time.Duration) *DetectedBottleneck {
	pa.mutex.RLock()
	dataPoints, exists := pa.historicalData[rule.Metric]
	pa.mutex.RUnlock()

	if !exists || len(dataPoints) == 0 {
		return nil
	}

	// Filter data points within the analysis period
	cutoff := time.Now().Add(-period)
	var recentPoints []DataPoint
	for _, point := range dataPoints {
		if point.Timestamp.After(cutoff) {
			recentPoints = append(recentPoints, point)
		}
	}

	if len(recentPoints) == 0 {
		return nil
	}

	// Evaluate condition
	currentValue := recentPoints[len(recentPoints)-1].Value
	violated := false

	switch rule.Condition {
	case "greater_than":
		violated = currentValue > rule.Threshold
	case "less_than":
		violated = currentValue < rule.Threshold
	case "trend_up":
		violated = pa.isTrendIncreasing(recentPoints) && currentValue > rule.Threshold
	case "trend_down":
		violated = pa.isTrendDecreasing(recentPoints) && currentValue < rule.Threshold
	}

	if !violated {
		return nil
	}

	// Create bottleneck detection
	bottleneck := &DetectedBottleneck{
		ID:           fmt.Sprintf("bottleneck_%s_%d", rule.Name, time.Now().Unix()),
		Name:         rule.Name,
		Severity:     rule.Severity,
		Component:    rule.Metric,
		Metric:       rule.Metric,
		CurrentValue: currentValue,
		Threshold:    rule.Threshold,
		Impact:       pa.calculateBottleneckImpact(rule, currentValue),
		DetectedAt:   time.Now(),
		Duration:     pa.calculateViolationDuration(recentPoints, rule),
		Suggestions:  pa.generateBottleneckSuggestions(rule),
	}

	return bottleneck
}

// generateOptimizationSuggestions generates optimization suggestions
func (pa *PerformanceAnalyzer) generateOptimizationSuggestions(systemMetrics SystemMetrics, bottlenecks []DetectedBottleneck) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	for _, rule := range pa.optimizationEngine.optimizationRules {
		if pa.evaluateOptimizationConditions(rule.Conditions, systemMetrics, bottlenecks) {
			for i, suggestionText := range rule.Suggestions {
				suggestion := OptimizationSuggestion{
					ID:            fmt.Sprintf("opt_%s_%d_%d", rule.Name, time.Now().Unix(), i),
					Title:         rule.Name,
					Description:   suggestionText,
					Impact:        rule.Impact,
					Effort:        rule.Effort,
					Priority:      rule.Priority,
					Category:      pa.categorizeOptimization(rule.Name),
					Component:     pa.identifyComponent(rule.Name),
					EstimatedGain: pa.estimateOptimizationGain(rule),
					CreatedAt:     time.Now(),
					Status:        "pending",
				}
				suggestions = append(suggestions, suggestion)
			}

			// Record optimization suggestion metric
			pa.performanceMetrics.OptimizationSuggestionsTotal.WithLabelValues(
				rule.Name, rule.Impact,
			).Inc()
		}
	}

	// Sort by priority
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Priority < suggestions[j].Priority
	})

	return suggestions
}

// performTrendAnalysis performs trend analysis on metrics
func (pa *PerformanceAnalyzer) performTrendAnalysis(period time.Duration) map[string]TrendAnalysis {
	trends := make(map[string]TrendAnalysis)

	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	for metric, dataPoints := range pa.historicalData {
		if len(dataPoints) < 10 { // Need minimum data points for trend analysis
			continue
		}

		// Filter data points within the analysis period
		cutoff := time.Now().Add(-period)
		var recentPoints []DataPoint
		for _, point := range dataPoints {
			if point.Timestamp.After(cutoff) {
				recentPoints = append(recentPoints, point)
			}
		}

		if len(recentPoints) < 5 {
			continue
		}

		trend := pa.calculateTrend(recentPoints)
		trends[metric] = trend
	}

	return trends
}

// calculateTrend calculates trend analysis for data points
func (pa *PerformanceAnalyzer) calculateTrend(dataPoints []DataPoint) TrendAnalysis {
	n := len(dataPoints)
	if n < 2 {
		return TrendAnalysis{Trend: "stable", TrendStrength: 0, Confidence: 0}
	}

	// Calculate linear regression
	var sumX, sumY, sumXY, sumX2 float64
	for i, point := range dataPoints {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (trend)
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)

	// Calculate correlation coefficient for confidence
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	var sumXDiff2, sumYDiff2, sumXYDiff float64
	for i, point := range dataPoints {
		xDiff := float64(i) - meanX
		yDiff := point.Value - meanY
		sumXDiff2 += xDiff * xDiff
		sumYDiff2 += yDiff * yDiff
		sumXYDiff += xDiff * yDiff
	}

	correlation := sumXYDiff / math.Sqrt(sumXDiff2*sumYDiff2)

	// Determine trend direction and strength
	var trendDirection string
	trendStrength := math.Abs(slope)

	if math.Abs(slope) < 0.01 {
		trendDirection = "stable"
	} else if slope > 0 {
		trendDirection = "increasing"
	} else {
		trendDirection = "decreasing"
	}

	// Check for volatility
	variance := pa.calculateVariance(dataPoints)
	if variance > pa.calculateMean(dataPoints)*0.5 {
		trendDirection = "volatile"
	}

	// Predict next value
	nextValue := meanY + slope*float64(n)

	return TrendAnalysis{
		Trend:         trendDirection,
		TrendStrength: trendStrength,
		Prediction:    nextValue,
		Confidence:    math.Abs(correlation),
		DataPoints:    n,
		AnalyzedAt:    time.Now(),
	}
}

// generateRecommendations generates high-level recommendations
func (pa *PerformanceAnalyzer) generateRecommendations(systemMetrics SystemMetrics, bottlenecks []DetectedBottleneck, optimizations []OptimizationSuggestion) []string {
	var recommendations []string

	// System-level recommendations
	if systemMetrics.CPUUtilization > 80 {
		recommendations = append(recommendations, "Consider horizontal scaling or CPU optimization")
	}

	if systemMetrics.MemoryUtilization > 85 {
		recommendations = append(recommendations, "Implement memory optimization strategies")
	}

	if systemMetrics.GoroutineCount > 10000 {
		recommendations = append(recommendations, "Review goroutine lifecycle and implement pooling")
	}

	if systemMetrics.ErrorRate > 5 {
		recommendations = append(recommendations, "Investigate and fix high error rate")
	}

	// Bottleneck-based recommendations
	criticalBottlenecks := 0
	for _, bottleneck := range bottlenecks {
		if bottleneck.Severity == "critical" {
			criticalBottlenecks++
		}
	}

	if criticalBottlenecks > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d critical performance bottlenecks immediately", criticalBottlenecks))
	}

	// Optimization-based recommendations
	highImpactOptimizations := 0
	for _, opt := range optimizations {
		if opt.Impact == "high" {
			highImpactOptimizations++
		}
	}

	if highImpactOptimizations > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Prioritize %d high-impact optimization suggestions", highImpactOptimizations))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System performance is within acceptable ranges")
	}

	return recommendations
}

// calculatePerformanceScore calculates overall performance score (0-100)
func (pa *PerformanceAnalyzer) calculatePerformanceScore(systemMetrics SystemMetrics, bottlenecks []DetectedBottleneck) float64 {
	score := 100.0

	// Deduct points based on system metrics
	if systemMetrics.CPUUtilization > 80 {
		score -= (systemMetrics.CPUUtilization - 80) * 0.5
	}

	if systemMetrics.MemoryUtilization > 85 {
		score -= (systemMetrics.MemoryUtilization - 85) * 0.7
	}

	if systemMetrics.ErrorRate > 1 {
		score -= systemMetrics.ErrorRate * 5
	}

	// Deduct points based on bottlenecks
	for _, bottleneck := range bottlenecks {
		switch bottleneck.Severity {
		case "critical":
			score -= 20
		case "high":
			score -= 10
		case "medium":
			score -= 5
		case "low":
			score -= 2
		}
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}

	return score
}

// Helper methods

func (pa *PerformanceAnalyzer) getCPUUtilization() float64 {
	// This would integrate with system monitoring to get actual CPU utilization
	// For now, return a placeholder value
	return 45.0
}

func (pa *PerformanceAnalyzer) getAverageResponseTime() time.Duration {
	// This would integrate with HTTP metrics to get actual response time
	return time.Millisecond * 150
}

func (pa *PerformanceAnalyzer) getCurrentThroughput() float64 {
	// This would calculate current request throughput
	return 250.0
}

func (pa *PerformanceAnalyzer) getCurrentErrorRate() float64 {
	// This would calculate current error rate
	return 1.5
}

func (pa *PerformanceAnalyzer) calculateRuntimeScore(m *runtime.MemStats) float64 {
	score := 100.0

	// Factor in GC pressure
	if m.GCCPUFraction > 0.05 {
		score -= m.GCCPUFraction * 100
	}

	// Factor in memory utilization
	heapUtilization := float64(m.HeapAlloc) / float64(m.HeapSys)
	if heapUtilization > 0.8 {
		score -= (heapUtilization - 0.8) * 100
	}

	return math.Max(0, score)
}

func (pa *PerformanceAnalyzer) generateMemoryRecommendations(m *runtime.MemStats) []string {
	var recommendations []string

	heapUtilization := float64(m.HeapAlloc) / float64(m.HeapSys)
	if heapUtilization > 0.85 {
		recommendations = append(recommendations, "High heap utilization - consider memory optimization")
	}

	if m.GCCPUFraction > 0.05 {
		recommendations = append(recommendations, "High GC overhead - optimize allocation patterns")
	}

	return recommendations
}

func (pa *PerformanceAnalyzer) evaluateGoroutineHealth(count int) string {
	if count > 10000 {
		return "critical"
	} else if count > 5000 {
		return "high"
	} else if count > 1000 {
		return "medium"
	}
	return "healthy"
}

func (pa *PerformanceAnalyzer) analyzeGoroutineTrend() string {
	pa.mutex.RLock()
	dataPoints, exists := pa.historicalData["goroutine_count"]
	pa.mutex.RUnlock()

	if !exists || len(dataPoints) < 10 {
		return "insufficient_data"
	}

	if pa.isTrendIncreasing(dataPoints) {
		return "increasing"
	} else if pa.isTrendDecreasing(dataPoints) {
		return "decreasing"
	}
	return "stable"
}

func (pa *PerformanceAnalyzer) generateGoroutineRecommendations(count int) []string {
	var recommendations []string

	if count > 5000 {
		recommendations = append(recommendations,
			"Implement goroutine pooling to control resource usage")
	}

	if count > 10000 {
		recommendations = append(recommendations,
			"Investigate potential goroutine leaks")
	}

	return recommendations
}

func (pa *PerformanceAnalyzer) isTrendIncreasing(dataPoints []DataPoint) bool {
	if len(dataPoints) < 3 {
		return false
	}

	recentAvg := pa.calculateMean(dataPoints[len(dataPoints)-3:])
	olderAvg := pa.calculateMean(dataPoints[:len(dataPoints)-3])

	return recentAvg > olderAvg*1.1 // 10% increase threshold
}

func (pa *PerformanceAnalyzer) isTrendDecreasing(dataPoints []DataPoint) bool {
	if len(dataPoints) < 3 {
		return false
	}

	recentAvg := pa.calculateMean(dataPoints[len(dataPoints)-3:])
	olderAvg := pa.calculateMean(dataPoints[:len(dataPoints)-3])

	return recentAvg < olderAvg*0.9 // 10% decrease threshold
}

func (pa *PerformanceAnalyzer) calculateMean(dataPoints []DataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0
	}

	sum := 0.0
	for _, point := range dataPoints {
		sum += point.Value
	}
	return sum / float64(len(dataPoints))
}

func (pa *PerformanceAnalyzer) calculateVariance(dataPoints []DataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	mean := pa.calculateMean(dataPoints)
	sumSquaredDiff := 0.0

	for _, point := range dataPoints {
		diff := point.Value - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(dataPoints)-1)
}

func (pa *PerformanceAnalyzer) calculateBottleneckImpact(rule DetectionRule, currentValue float64) string {
	ratio := currentValue / rule.Threshold

	if ratio > 2.0 {
		return "severe"
	} else if ratio > 1.5 {
		return "moderate"
	}
	return "minor"
}

func (pa *PerformanceAnalyzer) calculateViolationDuration(dataPoints []DataPoint, rule DetectionRule) time.Duration {
	if len(dataPoints) < 2 {
		return 0
	}

	var violationStart *time.Time
	for _, point := range dataPoints {
		isViolation := false
		switch rule.Condition {
		case "greater_than":
			isViolation = point.Value > rule.Threshold
		case "less_than":
			isViolation = point.Value < rule.Threshold
		}

		if isViolation && violationStart == nil {
			violationStart = &point.Timestamp
		} else if !isViolation {
			violationStart = nil
		}
	}

	if violationStart != nil {
		return time.Since(*violationStart)
	}
	return 0
}

func (pa *PerformanceAnalyzer) generateBottleneckSuggestions(rule DetectionRule) []string {
	suggestions := make([]string, len(rule.Actions))
	for i, action := range rule.Actions {
		suggestions[i] = fmt.Sprintf("Execute %s action for %s", action.Type, rule.Name)
	}
	return suggestions
}

func (pa *PerformanceAnalyzer) evaluateOptimizationConditions(conditions []string, systemMetrics SystemMetrics, bottlenecks []DetectedBottleneck) bool {
	for _, condition := range conditions {
		if !pa.evaluateCondition(condition, systemMetrics, bottlenecks) {
			return false
		}
	}
	return true
}

func (pa *PerformanceAnalyzer) evaluateCondition(condition string, systemMetrics SystemMetrics, bottlenecks []DetectedBottleneck) bool {
	// Parse simple conditions like "goroutine_count > 5000"
	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return false
	}

	metric := parts[0]
	operator := parts[1]
	thresholdStr := parts[2]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return false
	}

	var value float64
	switch metric {
	case "goroutine_count":
		value = float64(systemMetrics.GoroutineCount)
	case "cpu_utilization":
		value = systemMetrics.CPUUtilization
	case "memory_utilization":
		value = systemMetrics.MemoryUtilization
	case "response_time":
		value = float64(systemMetrics.ResponseTime.Milliseconds())
	case "error_rate":
		value = systemMetrics.ErrorRate
	default:
		return false
	}

	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

func (pa *PerformanceAnalyzer) categorizeOptimization(name string) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "memory") {
		return "memory"
	} else if strings.Contains(name, "cpu") {
		return "cpu"
	} else if strings.Contains(name, "goroutine") {
		return "concurrency"
	} else if strings.Contains(name, "response") {
		return "latency"
	}
	return "general"
}

func (pa *PerformanceAnalyzer) identifyComponent(name string) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "runtime") {
		return "runtime"
	} else if strings.Contains(name, "database") {
		return "database"
	} else if strings.Contains(name, "network") {
		return "network"
	}
	return "application"
}

func (pa *PerformanceAnalyzer) estimateOptimizationGain(rule OptimizationRule) float64 {
	// Simple estimation based on impact and priority
	switch rule.Impact {
	case "high":
		return 25.0 + float64(10-rule.Priority)*2.5
	case "medium":
		return 15.0 + float64(10-rule.Priority)*1.5
	case "low":
		return 5.0 + float64(10-rule.Priority)*0.5
	default:
		return 10.0
	}
}

// StartPerformanceMonitoring starts the continuous performance monitoring
func (pa *PerformanceAnalyzer) StartPerformanceMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pa.collectAndRecordMetrics()
		}
	}
}

// collectAndRecordMetrics collects current metrics and records them for analysis
func (pa *PerformanceAnalyzer) collectAndRecordMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Record key metrics
	pa.recordDataPoint("goroutine_count", float64(runtime.NumGoroutine()), nil)
	pa.recordDataPoint("memory_utilization", float64(m.Alloc)/float64(m.Sys)*100, nil)
	pa.recordDataPoint("cpu_utilization", pa.getCPUUtilization(), nil)
	pa.recordDataPoint("response_time", float64(pa.getAverageResponseTime().Milliseconds()), nil)
	pa.recordDataPoint("error_rate", pa.getCurrentErrorRate(), nil)
	pa.recordDataPoint("gc_pause_time", float64(m.PauseNs[(m.NumGC+255)%256]), nil)
}

// Stop stops the performance analyzer
func (pa *PerformanceAnalyzer) Stop(ctx context.Context) error {
	if pa.profilingServer != nil {
		return pa.profilingServer.Shutdown(ctx)
	}
	return nil
}

package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPerformanceAnalyzer(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	metricsCollector := &MetricsCollector{} // Mock metrics collector
	config := &ProfilingConfig{
		Enabled:         true,
		ListenAddress:   ":0", // Use any available port
		CPUProfiling:    true,
		MemoryProfiling: true,
		SampleRate:      time.Second,
		ProfileDuration: time.Minute,
	}

	pa := NewPerformanceAnalyzer(logger, registry, metricsCollector, config)

	assert.NotNil(t, pa)
	assert.NotNil(t, pa.performanceMetrics)
	assert.NotNil(t, pa.bottleneckDetector)
	assert.NotNil(t, pa.optimizationEngine)
	assert.Equal(t, 1000, pa.maxDataPoints)
}

func TestPerformanceAnalyzer_RecordDataPoint(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Record some data points
	pa.recordDataPoint("test_metric", 10.0, map[string]string{"test": "data"})
	pa.recordDataPoint("test_metric", 15.0, nil)
	pa.recordDataPoint("test_metric", 20.0, nil)

	// Verify data points are recorded
	pa.mutex.RLock()
	dataPoints := pa.historicalData["test_metric"]
	pa.mutex.RUnlock()

	assert.Len(t, dataPoints, 3)
	assert.Equal(t, 10.0, dataPoints[0].Value)
	assert.Equal(t, 15.0, dataPoints[1].Value)
	assert.Equal(t, 20.0, dataPoints[2].Value)
	assert.NotNil(t, dataPoints[0].Metadata)
}

func TestPerformanceAnalyzer_RecordDataPointMaxLimit(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)
	pa.maxDataPoints = 5 // Set small limit for testing

	// Record more data points than the limit
	for i := 0; i < 10; i++ {
		pa.recordDataPoint("test_metric", float64(i), nil)
	}

	// Verify only the last 5 data points are kept
	pa.mutex.RLock()
	dataPoints := pa.historicalData["test_metric"]
	pa.mutex.RUnlock()

	assert.Len(t, dataPoints, 5)
	assert.Equal(t, 5.0, dataPoints[0].Value) // First should be value 5
	assert.Equal(t, 9.0, dataPoints[4].Value) // Last should be value 9
}

func TestPerformanceAnalyzer_CollectSystemMetrics(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	metrics := pa.collectSystemMetrics()

	assert.Greater(t, metrics.GoroutineCount, 0)
	assert.GreaterOrEqual(t, metrics.MemoryUtilization, 0.0)
	assert.GreaterOrEqual(t, metrics.CPUUtilization, 0.0)
	assert.GreaterOrEqual(t, metrics.HeapSize, int64(0))
}

func TestPerformanceAnalyzer_AnalyzeRuntime(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)
	ctx := context.Background()

	analysis := pa.analyzeRuntime(ctx)

	assert.Contains(t, analysis, "timestamp")
	assert.Contains(t, analysis, "goroutines")
	assert.Contains(t, analysis, "cpu_cores")
	assert.Contains(t, analysis, "gc_stats")
	assert.Contains(t, analysis, "memory")
	assert.Contains(t, analysis, "performance_score")

	goroutines, ok := analysis["goroutines"].(int)
	assert.True(t, ok)
	assert.Greater(t, goroutines, 0)
}

func TestPerformanceAnalyzer_AnalyzeMemory(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)
	ctx := context.Background()

	analysis := pa.analyzeMemory(ctx)

	assert.Contains(t, analysis, "timestamp")
	assert.Contains(t, analysis, "heap")
	assert.Contains(t, analysis, "stack")
	assert.Contains(t, analysis, "gc")
	assert.Contains(t, analysis, "recommendations")

	heap, ok := analysis["heap"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, heap, "allocated")
	assert.Contains(t, heap, "utilization")
}

func TestPerformanceAnalyzer_AnalyzeGoroutines(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)
	ctx := context.Background()

	analysis := pa.analyzeGoroutines(ctx)

	assert.Contains(t, analysis, "timestamp")
	assert.Contains(t, analysis, "current_count")
	assert.Contains(t, analysis, "status")
	assert.Contains(t, analysis, "trend")
	assert.Contains(t, analysis, "recommendations")

	count, ok := analysis["current_count"].(int)
	assert.True(t, ok)
	assert.Greater(t, count, 0)
}

func TestPerformanceAnalyzer_EvaluateDetectionRule(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Setup test data
	for i := 0; i < 10; i++ {
		pa.recordDataPoint("cpu_utilization", 85.0, nil) // Above threshold
	}

	rule := DetectionRule{
		Name:      "High CPU",
		Metric:    "cpu_utilization",
		Condition: "greater_than",
		Threshold: 80.0,
		Severity:  "high",
	}

	bottleneck := pa.evaluateDetectionRule(rule, time.Hour)

	assert.NotNil(t, bottleneck)
	assert.Equal(t, "High CPU", bottleneck.Name)
	assert.Equal(t, "high", bottleneck.Severity)
	assert.Equal(t, 85.0, bottleneck.CurrentValue)
	assert.Equal(t, 80.0, bottleneck.Threshold)
}

func TestPerformanceAnalyzer_EvaluateDetectionRuleNoViolation(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Setup test data with values below threshold
	for i := 0; i < 10; i++ {
		pa.recordDataPoint("cpu_utilization", 70.0, nil) // Below threshold
	}

	rule := DetectionRule{
		Name:      "High CPU",
		Metric:    "cpu_utilization",
		Condition: "greater_than",
		Threshold: 80.0,
		Severity:  "high",
	}

	bottleneck := pa.evaluateDetectionRule(rule, time.Hour)

	assert.Nil(t, bottleneck)
}

func TestPerformanceAnalyzer_CalculateTrend(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Create increasing trend
	var dataPoints []DataPoint
	for i := 0; i < 10; i++ {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Value:     float64(i * 10), // Increasing values
		})
	}

	trend := pa.calculateTrend(dataPoints)

	assert.Equal(t, "increasing", trend.Trend)
	assert.Greater(t, trend.TrendStrength, 0.0)
	assert.Greater(t, trend.Confidence, 0.8) // Should be high confidence for linear trend
}

func TestPerformanceAnalyzer_CalculateTrendStable(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Create stable trend
	var dataPoints []DataPoint
	for i := 0; i < 10; i++ {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Value:     50.0, // Constant values
		})
	}

	trend := pa.calculateTrend(dataPoints)

	assert.Equal(t, "stable", trend.Trend)
	assert.Equal(t, 10, trend.DataPoints)
}

func TestPerformanceAnalyzer_GenerateOptimizationSuggestions(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	systemMetrics := SystemMetrics{
		GoroutineCount:    6000, // Above threshold in rule
		CPUUtilization:    75.0,
		MemoryUtilization: 80.0,
	}

	var bottlenecks []DetectedBottleneck

	suggestions := pa.generateOptimizationSuggestions(systemMetrics, bottlenecks)

	assert.NotEmpty(t, suggestions)

	// Find the goroutine optimization suggestion
	var found bool
	for _, suggestion := range suggestions {
		if suggestion.Category == "concurrency" {
			found = true
			assert.Contains(t, suggestion.Description, "goroutine")
			break
		}
	}
	assert.True(t, found, "Should find goroutine optimization suggestion")
}

func TestPerformanceAnalyzer_CalculatePerformanceScore(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Test good performance metrics
	goodMetrics := SystemMetrics{
		CPUUtilization:    50.0,
		MemoryUtilization: 60.0,
		ErrorRate:         0.5,
	}

	var noBottlenecks []DetectedBottleneck
	score := pa.calculatePerformanceScore(goodMetrics, noBottlenecks)
	assert.Equal(t, 100.0, score)

	// Test poor performance metrics
	poorMetrics := SystemMetrics{
		CPUUtilization:    90.0, // High CPU
		MemoryUtilization: 95.0, // High memory
		ErrorRate:         10.0, // High error rate
	}

	criticalBottlenecks := []DetectedBottleneck{
		{Severity: "critical"},
		{Severity: "high"},
	}

	score = pa.calculatePerformanceScore(poorMetrics, criticalBottlenecks)
	assert.Less(t, score, 50.0) // Should be significantly reduced
}

func TestPerformanceAnalyzer_GeneratePerformanceReport(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)
	ctx := context.Background()

	// Setup some test data
	pa.recordDataPoint("cpu_utilization", 85.0, nil)
	pa.recordDataPoint("memory_utilization", 75.0, nil)

	report, err := pa.GeneratePerformanceReport(ctx, time.Hour)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotZero(t, report.GeneratedAt)
	assert.Equal(t, time.Hour, report.AnalysisPeriod)
	assert.GreaterOrEqual(t, report.OverallScore, 0.0)
	assert.LessOrEqual(t, report.OverallScore, 100.0)
	assert.NotNil(t, report.SystemMetrics)
	assert.NotNil(t, report.DetectedBottlenecks)
	assert.NotNil(t, report.Optimizations)
	assert.NotNil(t, report.Trends)
	assert.NotEmpty(t, report.Recommendations)
}

func TestPerformanceAnalyzer_HTTPHandlers(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	tests := []struct {
		name     string
		endpoint string
		handler  http.HandlerFunc
	}{
		{
			name:     "Runtime Analysis",
			endpoint: "/debug/analysis/runtime",
			handler:  pa.handleRuntimeAnalysis,
		},
		{
			name:     "Memory Analysis",
			endpoint: "/debug/analysis/memory",
			handler:  pa.handleMemoryAnalysis,
		},
		{
			name:     "Goroutine Analysis",
			endpoint: "/debug/analysis/goroutines",
			handler:  pa.handleGoroutineAnalysis,
		},
		{
			name:     "Performance Analysis",
			endpoint: "/debug/analysis/performance",
			handler:  pa.handlePerformanceAnalysis,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Verify JSON response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response)
		})
	}
}

func TestPerformanceAnalyzer_EvaluateOptimizationConditions(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	systemMetrics := SystemMetrics{
		GoroutineCount:    6000,
		CPUUtilization:    75.0,
		MemoryUtilization: 80.0,
		ErrorRate:         2.0,
	}

	var bottlenecks []DetectedBottleneck

	tests := []struct {
		name       string
		conditions []string
		expected   bool
	}{
		{
			name:       "Goroutine count condition met",
			conditions: []string{"goroutine_count > 5000"},
			expected:   true,
		},
		{
			name:       "CPU condition not met",
			conditions: []string{"cpu_utilization > 80"},
			expected:   false,
		},
		{
			name:       "Multiple conditions - all met",
			conditions: []string{"goroutine_count > 5000", "memory_utilization > 70"},
			expected:   true,
		},
		{
			name:       "Multiple conditions - not all met",
			conditions: []string{"goroutine_count > 5000", "cpu_utilization > 80"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pa.evaluateOptimizationConditions(tt.conditions, systemMetrics, bottlenecks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPerformanceAnalyzer_TrendAnalysis(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Setup different trend patterns
	testCases := []struct {
		name     string
		values   []float64
		expected string
	}{
		{
			name:     "Increasing trend",
			values:   []float64{10, 15, 20, 25, 30, 35, 40},
			expected: "increasing",
		},
		{
			name:     "Decreasing trend",
			values:   []float64{40, 35, 30, 25, 20, 15, 10},
			expected: "decreasing",
		},
		{
			name:     "Stable trend",
			values:   []float64{25, 25, 26, 24, 25, 25, 25},
			expected: "stable",
		},
		{
			name:     "Volatile trend",
			values:   []float64{10, 50, 5, 45, 15, 40, 20},
			expected: "volatile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metric := "test_" + tc.name

			// Record test data
			for i, value := range tc.values {
				timestamp := time.Now().Add(time.Duration(i) * time.Minute)
				pa.recordDataPoint(metric, value, nil)
				// Update timestamp manually for testing
				pa.mutex.Lock()
				pa.historicalData[metric][i].Timestamp = timestamp
				pa.mutex.Unlock()
			}

			// Perform trend analysis
			trends := pa.performTrendAnalysis(time.Hour)

			trend, exists := trends[metric]
			assert.True(t, exists, "Trend should exist for metric %s", metric)
			assert.Equal(t, tc.expected, trend.Trend)
			assert.Equal(t, len(tc.values), trend.DataPoints)
		})
	}
}

func TestPerformanceAnalyzer_StartStopMonitoring(t *testing.T) {
	pa := createTestPerformanceAnalyzer(t)

	// Test starting monitoring
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// This should run briefly and then stop when context is cancelled
	pa.StartPerformanceMonitoring(ctx, time.Millisecond*10)

	// Verify some data was collected
	pa.mutex.RLock()
	hasData := len(pa.historicalData) > 0
	pa.mutex.RUnlock()

	assert.True(t, hasData, "Should have collected some performance data")
}

func TestPerformanceAnalyzer_ProfilingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling integration test in short mode")
	}

	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	metricsCollector := &MetricsCollector{}
	config := &ProfilingConfig{
		Enabled:         true,
		ListenAddress:   "127.0.0.1:0", // Use any available port
		CPUProfiling:    true,
		MemoryProfiling: true,
		SampleRate:      time.Second,
		ProfileDuration: time.Second * 10,
	}

	pa := NewPerformanceAnalyzer(logger, registry, metricsCollector, config)
	defer func() {
		if pa.profilingServer != nil {
			pa.Stop(context.Background())
		}
	}()

	// Wait a moment for server to start
	time.Sleep(time.Millisecond * 100)

	assert.NotNil(t, pa.profilingServer)
}

// Benchmark tests
func BenchmarkPerformanceAnalyzer_RecordDataPoint(b *testing.B) {
	pa := createTestPerformanceAnalyzer(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pa.recordDataPoint("benchmark_metric", float64(i), nil)
	}
}

func BenchmarkPerformanceAnalyzer_CalculateTrend(b *testing.B) {
	pa := createTestPerformanceAnalyzer(&testing.T{})

	// Setup test data
	var dataPoints []DataPoint
	for i := 0; i < 100; i++ {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Value:     float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pa.calculateTrend(dataPoints)
	}
}

func BenchmarkPerformanceAnalyzer_GenerateReport(b *testing.B) {
	pa := createTestPerformanceAnalyzer(&testing.T{})
	ctx := context.Background()

	// Setup some test data
	for i := 0; i < 50; i++ {
		pa.recordDataPoint("cpu_utilization", float64(50+i), nil)
		pa.recordDataPoint("memory_utilization", float64(60+i), nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pa.GeneratePerformanceReport(ctx, time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to create test performance analyzer
func createTestPerformanceAnalyzer(t *testing.T) *PerformanceAnalyzer {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	metricsCollector := &MetricsCollector{} // Mock metrics collector
	config := &ProfilingConfig{
		Enabled:         false, // Disable for most tests
		ListenAddress:   ":0",
		CPUProfiling:    true,
		MemoryProfiling: true,
		SampleRate:      time.Second,
		ProfileDuration: time.Minute,
	}

	return NewPerformanceAnalyzer(logger, registry, metricsCollector, config)
}

// Mock MetricsCollector for testing
type mockMetricsCollector struct{}

func (m *mockMetricsCollector) Start() error                            { return nil }
func (m *mockMetricsCollector) Stop() error                             { return nil }
func (m *mockMetricsCollector) GetMetrics() map[string]float64          { return make(map[string]float64) }
func (m *mockMetricsCollector) RecordMetric(name string, value float64) {}

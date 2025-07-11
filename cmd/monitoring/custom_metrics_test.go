package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCustomMetricsManager_Creation(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()

	cmm := NewCustomMetricsManager(logger, registry)

	assert.NotNil(t, cmm)
	assert.NotNil(t, cmm.businessMetrics)
	assert.NotNil(t, cmm.performanceMetrics)
	assert.NotNil(t, cmm.usageMetrics)
	assert.NotNil(t, cmm.customCounters)
	assert.NotNil(t, cmm.customGauges)
	assert.NotNil(t, cmm.customHistograms)
	assert.NotNil(t, cmm.customSummaries)
}

func TestCustomMetricsManager_BusinessMetrics(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	t.Run("RecordRepoClone", func(t *testing.T) {
		cmm.RecordRepoClone("test-org", "github", "success", 30*time.Second, "large")
		cmm.RecordRepoClone("test-org", "github", "failed", 5*time.Second, "small")

		// Verify metrics were recorded (we can't easily check values, but ensure no panic)
		assert.NotPanics(t, func() {
			cmm.RecordRepoClone("another-org", "gitlab", "success", 45*time.Second, "medium")
		})
	})

	t.Run("RecordRepoSync", func(t *testing.T) {
		cmm.RecordRepoSync("pull", "success", "test-org")
		cmm.RecordRepoSync("push", "failed", "test-org")

		assert.NotPanics(t, func() {
			cmm.RecordRepoSync("fetch", "success", "another-org")
		})
	})

	t.Run("SetRepoSize", func(t *testing.T) {
		cmm.SetRepoSize("repo1", "test-org", 1024*1024*50)  // 50MB
		cmm.SetRepoSize("repo2", "test-org", 1024*1024*100) // 100MB

		assert.NotPanics(t, func() {
			cmm.SetRepoSize("repo3", "another-org", 1024*1024*200)
		})
	})

	t.Run("SetOrganizationsTotal", func(t *testing.T) {
		cmm.SetOrganizationsTotal(5)
		cmm.SetOrganizationsTotal(10)

		assert.NotPanics(t, func() {
			cmm.SetOrganizationsTotal(15)
		})
	})

	t.Run("SetProjectsActiveTotal", func(t *testing.T) {
		cmm.SetProjectsActiveTotal(25)
		cmm.SetProjectsActiveTotal(30)

		assert.NotPanics(t, func() {
			cmm.SetProjectsActiveTotal(35)
		})
	})

	t.Run("SetUsersActiveTotal", func(t *testing.T) {
		cmm.SetUsersActiveTotal(100)
		cmm.SetUsersActiveTotal(150)

		assert.NotPanics(t, func() {
			cmm.SetUsersActiveTotal(200)
		})
	})

	t.Run("RecordTaskCompletion", func(t *testing.T) {
		cmm.RecordTaskCompletion("clone", "test-org", "success")
		cmm.RecordTaskCompletion("sync", "test-org", "failed")

		assert.NotPanics(t, func() {
			cmm.RecordTaskCompletion("analyze", "another-org", "success")
		})
	})

	t.Run("SetTaskFailureRate", func(t *testing.T) {
		cmm.SetTaskFailureRate("clone", "1h", 5.5)
		cmm.SetTaskFailureRate("sync", "24h", 2.1)

		assert.NotPanics(t, func() {
			cmm.SetTaskFailureRate("analyze", "1h", 0.8)
		})
	})

	t.Run("SetTaskThroughput", func(t *testing.T) {
		cmm.SetTaskThroughput("clone", "pool1", 10.5)
		cmm.SetTaskThroughput("sync", "pool2", 25.0)

		assert.NotPanics(t, func() {
			cmm.SetTaskThroughput("analyze", "pool1", 5.2)
		})
	})

	t.Run("SetIntegrationStatus", func(t *testing.T) {
		cmm.SetIntegrationStatus("github", "api.github.com", true)
		cmm.SetIntegrationStatus("gitlab", "gitlab.com", false)

		assert.NotPanics(t, func() {
			cmm.SetIntegrationStatus("slack", "hooks.slack.com", true)
		})
	})

	t.Run("RecordIntegrationAPICall", func(t *testing.T) {
		cmm.RecordIntegrationAPICall("github", "GET", "/repos", 150*time.Millisecond)
		cmm.RecordIntegrationAPICall("gitlab", "POST", "/projects", 300*time.Millisecond)

		assert.NotPanics(t, func() {
			cmm.RecordIntegrationAPICall("slack", "POST", "/webhooks", 100*time.Millisecond)
		})
	})

	t.Run("SetIntegrationRateLimit", func(t *testing.T) {
		cmm.SetIntegrationRateLimit("github", "oauth", 4500)
		cmm.SetIntegrationRateLimit("gitlab", "personal", 1800)

		assert.NotPanics(t, func() {
			cmm.SetIntegrationRateLimit("slack", "bot", 900)
		})
	})
}

func TestCustomMetricsManager_PerformanceMetrics(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	t.Run("SetCPUUtilization", func(t *testing.T) {
		cmm.SetCPUUtilization("0", "user", 25.5)
		cmm.SetCPUUtilization("0", "system", 8.2)
		cmm.SetCPUUtilization("1", "user", 30.1)

		assert.NotPanics(t, func() {
			cmm.SetCPUUtilization("all", "idle", 65.0)
		})
	})

	t.Run("SetMemoryUtilization", func(t *testing.T) {
		cmm.SetMemoryUtilization("heap", 45.7)
		cmm.SetMemoryUtilization("stack", 12.3)
		cmm.SetMemoryUtilization("cache", 22.1)

		assert.NotPanics(t, func() {
			cmm.SetMemoryUtilization("buffer", 8.9)
		})
	})

	t.Run("RecordDiskIO", func(t *testing.T) {
		cmm.RecordDiskIO("sda", "read", 100, 4096*100, "in")
		cmm.RecordDiskIO("sda", "write", 50, 4096*50, "out")

		assert.NotPanics(t, func() {
			cmm.RecordDiskIO("sdb", "read", 75, 4096*75, "in")
		})
	})

	t.Run("RecordNetworkIO", func(t *testing.T) {
		cmm.RecordNetworkIO("eth0", "in", 1024*1024)
		cmm.RecordNetworkIO("eth0", "out", 512*1024)

		assert.NotPanics(t, func() {
			cmm.RecordNetworkIO("lo", "in", 256*1024)
		})
	})

	t.Run("SetGoroutineCount", func(t *testing.T) {
		cmm.SetGoroutineCount(150)
		cmm.SetGoroutineCount(200)

		assert.NotPanics(t, func() {
			cmm.SetGoroutineCount(175)
		})
	})

	t.Run("RecordGCDuration", func(t *testing.T) {
		cmm.RecordGCDuration("mark", 5*time.Millisecond)
		cmm.RecordGCDuration("sweep", 3*time.Millisecond)

		assert.NotPanics(t, func() {
			cmm.RecordGCDuration("concurrent", 8*time.Millisecond)
		})
	})

	t.Run("SetHeapAlloc", func(t *testing.T) {
		cmm.SetHeapAlloc(1024 * 1024 * 50) // 50MB
		cmm.SetHeapAlloc(1024 * 1024 * 75) // 75MB

		assert.NotPanics(t, func() {
			cmm.SetHeapAlloc(1024 * 1024 * 60) // 60MB
		})
	})
}

func TestCustomMetricsManager_UsageMetrics(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	t.Run("SetActiveUsers", func(t *testing.T) {
		cmm.SetActiveUsers("test-org", "admin", 5, 25, 100)
		cmm.SetActiveUsers("test-org", "user", 15, 75, 300)

		assert.NotPanics(t, func() {
			cmm.SetActiveUsers("another-org", "viewer", 3, 12, 45)
		})
	})

	t.Run("RecordUserSession", func(t *testing.T) {
		cmm.RecordUserSession("test-org", "admin", "web", 2*time.Hour)
		cmm.RecordUserSession("test-org", "user", "api", 30*time.Minute)

		assert.NotPanics(t, func() {
			cmm.RecordUserSession("another-org", "viewer", "mobile", 45*time.Minute)
		})
	})

	t.Run("RecordFeatureUsage", func(t *testing.T) {
		cmm.RecordFeatureUsage("bulk-clone", "test-org", "admin", 5*time.Second, "high")
		cmm.RecordFeatureUsage("sync", "test-org", "user", 2*time.Second, "medium")

		assert.NotPanics(t, func() {
			cmm.RecordFeatureUsage("search", "another-org", "viewer", 500*time.Millisecond, "low")
		})
	})

	t.Run("SetFeatureErrorRate", func(t *testing.T) {
		cmm.SetFeatureErrorRate("bulk-clone", "timeout", 2.5)
		cmm.SetFeatureErrorRate("sync", "network", 1.8)

		assert.NotPanics(t, func() {
			cmm.SetFeatureErrorRate("search", "validation", 0.5)
		})
	})

	t.Run("RecordResourceUsage", func(t *testing.T) {
		cmm.RecordResourceUsage("test-org", "clone", "in", 1024*1024*100, "local", 1024*1024*1024*5, "cpu", 2.5)
		cmm.RecordResourceUsage("test-org", "sync", "out", 1024*1024*50, "cache", 1024*1024*500, "memory", 1.0)

		assert.NotPanics(t, func() {
			cmm.RecordResourceUsage("another-org", "search", "in", 1024*1024*10, "remote", 1024*1024*100, "io", 0.1)
		})
	})

	t.Run("RecordAPICall", func(t *testing.T) {
		cmm.RecordAPICall("v1", "/repos", "GET", "200", 150*time.Millisecond)
		cmm.RecordAPICall("v1", "/orgs", "POST", "201", 300*time.Millisecond)

		assert.NotPanics(t, func() {
			cmm.RecordAPICall("v2", "/search", "GET", "200", 80*time.Millisecond)
		})
	})

	t.Run("SetAPIQuotaUsage", func(t *testing.T) {
		cmm.SetAPIQuotaUsage("test-org", "key1", "requests", 75.5)
		cmm.SetAPIQuotaUsage("test-org", "key2", "bandwidth", 45.2)

		assert.NotPanics(t, func() {
			cmm.SetAPIQuotaUsage("another-org", "key3", "storage", 20.1)
		})
	})

	t.Run("RecordAPIRetry", func(t *testing.T) {
		cmm.RecordAPIRetry("/repos", "timeout", "200")
		cmm.RecordAPIRetry("/orgs", "rate_limit", "429")

		assert.NotPanics(t, func() {
			cmm.RecordAPIRetry("/search", "network", "500")
		})
	})

	t.Run("RecordGeoRequest", func(t *testing.T) {
		cmm.RecordGeoRequest("US", "West", "America/Los_Angeles", "14")
		cmm.RecordGeoRequest("KR", "Seoul", "Asia/Seoul", "22")

		assert.NotPanics(t, func() {
			cmm.RecordGeoRequest("JP", "Tokyo", "Asia/Tokyo", "10")
		})
	})
}

func TestCustomMetricsManager_CustomMetrics(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	t.Run("CreateCustomCounter", func(t *testing.T) {
		err := cmm.CreateCustomCounter("test_counter", "Test counter metric", []string{"label1", "label2"}, nil)
		assert.NoError(t, err)

		// Try to create duplicate
		err = cmm.CreateCustomCounter("test_counter", "Duplicate counter", []string{"label1"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		// Get the counter and use it
		counter, err := cmm.GetCustomCounter("test_counter")
		assert.NoError(t, err)
		assert.NotNil(t, counter)

		counter.WithLabelValues("value1", "value2").Inc()
	})

	t.Run("CreateCustomGauge", func(t *testing.T) {
		err := cmm.CreateCustomGauge("test_gauge", "Test gauge metric", []string{"env"}, map[string]string{"service": "gzh-manager"})
		assert.NoError(t, err)

		// Try to create duplicate
		err = cmm.CreateCustomGauge("test_gauge", "Duplicate gauge", []string{"env"}, nil)
		assert.Error(t, err)

		// Get the gauge and use it
		gauge, err := cmm.GetCustomGauge("test_gauge")
		assert.NoError(t, err)
		assert.NotNil(t, gauge)

		gauge.WithLabelValues("production").Set(42.5)
	})

	t.Run("CreateCustomHistogram", func(t *testing.T) {
		buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}
		err := cmm.CreateCustomHistogram("test_histogram", "Test histogram metric", []string{"operation"}, buckets, nil)
		assert.NoError(t, err)

		// Try to create duplicate
		err = cmm.CreateCustomHistogram("test_histogram", "Duplicate histogram", []string{"operation"}, nil, nil)
		assert.Error(t, err)

		// Get the histogram and use it
		histogram, err := cmm.GetCustomHistogram("test_histogram")
		assert.NoError(t, err)
		assert.NotNil(t, histogram)

		histogram.WithLabelValues("clone").Observe(2.3)
	})

	t.Run("CreateCustomSummary", func(t *testing.T) {
		objectives := map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
		err := cmm.CreateCustomSummary("test_summary", "Test summary metric", []string{"api"}, objectives, nil)
		assert.NoError(t, err)

		// Try to create duplicate
		err = cmm.CreateCustomSummary("test_summary", "Duplicate summary", []string{"api"}, nil, nil)
		assert.Error(t, err)

		// Get the summary and use it
		summary, err := cmm.GetCustomSummary("test_summary")
		assert.NoError(t, err)
		assert.NotNil(t, summary)

		summary.WithLabelValues("v1").Observe(0.123)
	})

	t.Run("GetNonexistentMetric", func(t *testing.T) {
		_, err := cmm.GetCustomCounter("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		_, err = cmm.GetCustomGauge("nonexistent")
		assert.Error(t, err)

		_, err = cmm.GetCustomHistogram("nonexistent")
		assert.Error(t, err)

		_, err = cmm.GetCustomSummary("nonexistent")
		assert.Error(t, err)
	})

	t.Run("ListCustomMetrics", func(t *testing.T) {
		metrics := cmm.ListCustomMetrics()

		assert.Contains(t, metrics, "test_counter")
		assert.Equal(t, "counter", metrics["test_counter"])

		assert.Contains(t, metrics, "test_gauge")
		assert.Equal(t, "gauge", metrics["test_gauge"])

		assert.Contains(t, metrics, "test_histogram")
		assert.Equal(t, "histogram", metrics["test_histogram"])

		assert.Contains(t, metrics, "test_summary")
		assert.Equal(t, "summary", metrics["test_summary"])
	})

	t.Run("DeleteCustomMetric", func(t *testing.T) {
		// Create a metric to delete
		err := cmm.CreateCustomCounter("temp_counter", "Temporary counter", []string{"temp"}, nil)
		assert.NoError(t, err)

		// Verify it exists
		_, err = cmm.GetCustomCounter("temp_counter")
		assert.NoError(t, err)

		// Delete it
		err = cmm.DeleteCustomMetric("temp_counter")
		assert.NoError(t, err)

		// Verify it's gone
		_, err = cmm.GetCustomCounter("temp_counter")
		assert.Error(t, err)

		// Try to delete nonexistent metric
		err = cmm.DeleteCustomMetric("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCustomMetricsManager_MetricsSummary(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	summary := cmm.GetMetricsSummary()

	assert.Contains(t, summary, "business_metrics")
	assert.Contains(t, summary, "performance_metrics")
	assert.Contains(t, summary, "usage_metrics")
	assert.Contains(t, summary, "custom_metrics")

	businessMetrics := summary["business_metrics"].(map[string]interface{})
	assert.Contains(t, businessMetrics, "repo_operations")
	assert.Contains(t, businessMetrics, "organizations")
	assert.Contains(t, businessMetrics, "task_execution")
	assert.Contains(t, businessMetrics, "integrations")

	performanceMetrics := summary["performance_metrics"].(map[string]interface{})
	assert.Contains(t, performanceMetrics, "system_resources")
	assert.Contains(t, performanceMetrics, "application")
	assert.Contains(t, performanceMetrics, "database")
	assert.Contains(t, performanceMetrics, "cache")
	assert.Contains(t, performanceMetrics, "queue")

	usageMetrics := summary["usage_metrics"].(map[string]interface{})
	assert.Contains(t, usageMetrics, "user_activity")
	assert.Contains(t, usageMetrics, "feature_usage")
	assert.Contains(t, usageMetrics, "resource_usage")
	assert.Contains(t, usageMetrics, "api_usage")

	customMetrics := summary["custom_metrics"].(map[string]interface{})
	assert.Contains(t, customMetrics, "counters")
	assert.Contains(t, customMetrics, "gauges")
	assert.Contains(t, customMetrics, "histograms")
	assert.Contains(t, customMetrics, "summaries")
}

func TestCustomMetricsManager_StartStop(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start", func(t *testing.T) {
		err := cmm.Start(ctx)
		assert.NoError(t, err)

		// Give it a moment to start collecting metrics
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Stop", func(t *testing.T) {
		err := cmm.Stop()
		assert.NoError(t, err)
	})
}

func TestCustomMetricsManager_BusinessMetricsIntegration(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	// Simulate a complete business workflow
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Set initial state
		cmm.SetOrganizationsTotal(3)
		cmm.SetProjectsActiveTotal(15)
		cmm.SetUsersActiveTotal(50)

		// Record repository operations
		cmm.RecordRepoClone("acme-corp", "github", "success", 45*time.Second, "large")
		cmm.SetRepoSize("acme-repo", "acme-corp", 1024*1024*250) // 250MB
		cmm.RecordRepoSync("pull", "success", "acme-corp")

		// Record task execution
		cmm.RecordTaskCompletion("clone", "acme-corp", "success")
		cmm.SetTaskFailureRate("clone", "1h", 2.1)
		cmm.SetTaskThroughput("clone", "pool1", 8.5)

		// Record integration health
		cmm.SetIntegrationStatus("github", "api.github.com", true)
		cmm.RecordIntegrationAPICall("github", "GET", "/repos", 120*time.Millisecond)
		cmm.SetIntegrationRateLimit("github", "oauth", 4800)

		// Verify no panics occurred
		assert.NotPanics(t, func() {
			summary := cmm.GetMetricsSummary()
			assert.NotNil(t, summary)
		})
	})
}

func TestCustomMetricsManager_PerformanceMetricsIntegration(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	// Simulate performance monitoring
	t.Run("PerformanceMonitoring", func(t *testing.T) {
		// System metrics
		cmm.SetCPUUtilization("0", "user", 35.2)
		cmm.SetCPUUtilization("0", "system", 12.8)
		cmm.SetMemoryUtilization("heap", 65.5)

		// I/O metrics
		cmm.RecordDiskIO("sda", "read", 150, 4096*150, "in")
		cmm.RecordNetworkIO("eth0", "out", 1024*1024*2)

		// Runtime metrics
		cmm.SetGoroutineCount(245)
		cmm.RecordGCDuration("mark", 8*time.Millisecond)
		cmm.SetHeapAlloc(1024 * 1024 * 85) // 85MB

		// Verify no panics occurred
		assert.NotPanics(t, func() {
			summary := cmm.GetMetricsSummary()
			assert.NotNil(t, summary)
		})
	})
}

func TestCustomMetricsManager_UsageMetricsIntegration(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	// Simulate usage tracking
	t.Run("UsageTracking", func(t *testing.T) {
		// User activity
		cmm.SetActiveUsers("acme-corp", "admin", 3, 8, 25)
		cmm.SetActiveUsers("acme-corp", "user", 12, 35, 120)
		cmm.RecordUserSession("acme-corp", "admin", "web", 3*time.Hour)

		// Feature usage
		cmm.RecordFeatureUsage("bulk-clone", "acme-corp", "admin", 8*time.Second, "high")
		cmm.SetFeatureErrorRate("bulk-clone", "timeout", 1.5)

		// Resource usage
		cmm.RecordResourceUsage("acme-corp", "clone", "in", 1024*1024*200, "local", 1024*1024*1024*10, "cpu", 5.2)

		// API usage
		cmm.RecordAPICall("v1", "/repos", "GET", "200", 95*time.Millisecond)
		cmm.SetAPIQuotaUsage("acme-corp", "api-key-1", "requests", 65.8)
		cmm.RecordAPIRetry("/repos", "rate_limit", "200")

		// Geographical
		cmm.RecordGeoRequest("US", "West", "America/Los_Angeles", "15")

		// Verify no panics occurred
		assert.NotPanics(t, func() {
			summary := cmm.GetMetricsSummary()
			assert.NotNil(t, summary)
		})
	})
}

func TestCustomMetricsManager_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	// Test concurrent metric operations
	t.Run("ConcurrentOperations", func(t *testing.T) {
		done := make(chan bool, 10)

		// Launch concurrent goroutines
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Business metrics
				cmm.RecordRepoClone("org", "github", "success", time.Duration(id)*time.Second, "medium")
				cmm.SetOrganizationsTotal(float64(id))

				// Performance metrics
				cmm.SetCPUUtilization("all", "user", float64(id)*10)
				cmm.SetGoroutineCount(float64(100 + id*10))

				// Usage metrics
				cmm.SetActiveUsers("org", "user", float64(id), float64(id*2), float64(id*5))
				cmm.RecordAPICall("v1", "/test", "GET", "200", time.Duration(id)*time.Millisecond)

				// Custom metrics
				metricName := fmt.Sprintf("test_metric_%d", id)
				err := cmm.CreateCustomGauge(metricName, "Test metric", []string{"id"}, nil)
				if err == nil {
					gauge, _ := cmm.GetCustomGauge(metricName)
					gauge.WithLabelValues(fmt.Sprintf("%d", id)).Set(float64(id))
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify metrics summary still works
		summary := cmm.GetMetricsSummary()
		assert.NotNil(t, summary)
	})
}

func BenchmarkCustomMetricsManager_RecordMetrics(b *testing.B) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	cmm := NewCustomMetricsManager(logger, registry)

	b.Run("BusinessMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmm.RecordRepoClone("test-org", "github", "success", 30*time.Second, "medium")
			cmm.RecordTaskCompletion("clone", "test-org", "success")
			cmm.SetOrganizationsTotal(float64(i % 100))
		}
	})

	b.Run("PerformanceMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmm.SetCPUUtilization("all", "user", float64(i%100))
			cmm.SetGoroutineCount(float64(100 + i%50))
			cmm.RecordGCDuration("mark", time.Duration(i%10)*time.Millisecond)
		}
	})

	b.Run("UsageMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmm.RecordAPICall("v1", "/test", "GET", "200", time.Duration(i%100)*time.Millisecond)
			cmm.RecordFeatureUsage("test", "org", "user", time.Duration(i%1000)*time.Millisecond, "low")
			cmm.SetActiveUsers("org", "user", float64(i%10), float64(i%50), float64(i%200))
		}
	})
}

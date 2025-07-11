package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// CustomMetricsManager manages custom business and performance metrics
type CustomMetricsManager struct {
	logger   *zap.Logger
	registry *prometheus.Registry
	mutex    sync.RWMutex

	// Business Metrics
	businessMetrics *BusinessMetrics

	// Performance Indicators
	performanceMetrics *PerformanceMetrics

	// Usage Statistics
	usageMetrics *UsageMetrics

	// Custom metric definitions
	customCounters   map[string]*prometheus.CounterVec
	customGauges     map[string]*prometheus.GaugeVec
	customHistograms map[string]*prometheus.HistogramVec
	customSummaries  map[string]*prometheus.SummaryVec
}

// BusinessMetrics represents business-specific metrics
type BusinessMetrics struct {
	// Repository operations
	RepoCloneTotal     *prometheus.CounterVec
	RepoCloneDuration  *prometheus.HistogramVec
	RepoSyncOperations *prometheus.CounterVec
	RepoSizeBytes      *prometheus.GaugeVec

	// Organization and project metrics
	OrganizationsTotal  prometheus.Gauge
	ProjectsActiveTotal prometheus.Gauge
	UsersActiveTotal    prometheus.Gauge

	// Task execution business metrics
	TasksCompletedTotal     *prometheus.CounterVec
	TaskFailureRatePercent  *prometheus.GaugeVec
	TaskThroughputPerSecond *prometheus.GaugeVec

	// Integration health
	IntegrationUpStatus   *prometheus.GaugeVec
	IntegrationAPILatency *prometheus.HistogramVec
	IntegrationRateLimit  *prometheus.GaugeVec
}

// PerformanceMetrics represents performance indicators
type PerformanceMetrics struct {
	// System performance
	CPUUtilizationPercent    *prometheus.GaugeVec
	MemoryUtilizationPercent *prometheus.GaugeVec
	DiskIOOperationsTotal    *prometheus.CounterVec
	DiskIOBytesTotal         *prometheus.CounterVec
	NetworkIOBytesTotal      *prometheus.CounterVec

	// Application performance
	GoroutineCount prometheus.Gauge
	GCDuration     *prometheus.HistogramVec
	HeapAllocBytes prometheus.Gauge

	// Database performance (if applicable)
	DatabaseConnectionsActive  *prometheus.GaugeVec
	DatabaseQueryDuration      *prometheus.HistogramVec
	DatabaseConnectionPoolSize *prometheus.GaugeVec

	// Cache performance
	CacheHitRatio        *prometheus.GaugeVec
	CacheOperationsTotal *prometheus.CounterVec
	CacheEvictionsTotal  *prometheus.CounterVec

	// Queue performance
	QueueDepth              *prometheus.GaugeVec
	QueueProcessingDuration *prometheus.HistogramVec
	QueueThroughput         *prometheus.GaugeVec
}

// UsageMetrics represents usage statistics
type UsageMetrics struct {
	// User activity
	ActiveUsers5Min     *prometheus.GaugeVec
	ActiveUsers1Hour    *prometheus.GaugeVec
	ActiveUsers24Hour   *prometheus.GaugeVec
	UserSessionDuration *prometheus.HistogramVec

	// Feature usage
	FeatureUsageTotal *prometheus.CounterVec
	FeatureLatency    *prometheus.HistogramVec
	FeatureErrorRate  *prometheus.GaugeVec

	// Resource consumption
	BandwidthUsageBytes *prometheus.CounterVec
	StorageUsageBytes   *prometheus.GaugeVec
	ComputeHoursTotal   *prometheus.CounterVec

	// API usage
	APICallsTotal         *prometheus.CounterVec
	APILatencyPercentiles *prometheus.SummaryVec
	APIQuotaUsagePercent  *prometheus.GaugeVec
	APIRetryAttemptsTotal *prometheus.CounterVec

	// Geographical usage
	RequestsByCountry  *prometheus.CounterVec
	RequestsByTimezone *prometheus.CounterVec
}

// NewCustomMetricsManager creates a new custom metrics manager
func NewCustomMetricsManager(logger *zap.Logger, registry *prometheus.Registry) *CustomMetricsManager {
	cmm := &CustomMetricsManager{
		logger:           logger,
		registry:         registry,
		customCounters:   make(map[string]*prometheus.CounterVec),
		customGauges:     make(map[string]*prometheus.GaugeVec),
		customHistograms: make(map[string]*prometheus.HistogramVec),
		customSummaries:  make(map[string]*prometheus.SummaryVec),
	}

	cmm.initializeBusinessMetrics()
	cmm.initializePerformanceMetrics()
	cmm.initializeUsageMetrics()

	return cmm
}

// initializeBusinessMetrics initializes business metrics
func (cmm *CustomMetricsManager) initializeBusinessMetrics() {
	cmm.businessMetrics = &BusinessMetrics{
		RepoCloneTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "repo_clone_total",
				Help:      "Total number of repository clones",
			},
			[]string{"organization", "platform", "status"},
		),

		RepoCloneDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "repo_clone_duration_seconds",
				Help:      "Duration of repository clone operations",
				Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10),
			},
			[]string{"organization", "platform", "size_category"},
		),

		RepoSyncOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "repo_sync_operations_total",
				Help:      "Total number of repository sync operations",
			},
			[]string{"operation", "status", "organization"},
		),

		RepoSizeBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "repo_size_bytes",
				Help:      "Repository size in bytes",
			},
			[]string{"repository", "organization"},
		),

		OrganizationsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "organizations_total",
				Help:      "Total number of organizations managed",
			},
		),

		ProjectsActiveTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "projects_active_total",
				Help:      "Total number of active projects",
			},
		),

		UsersActiveTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "users_active_total",
				Help:      "Total number of active users",
			},
		),

		TasksCompletedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "tasks_completed_total",
				Help:      "Total number of completed tasks",
			},
			[]string{"task_type", "organization", "outcome"},
		),

		TaskFailureRatePercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "task_failure_rate_percent",
				Help:      "Task failure rate percentage",
			},
			[]string{"task_type", "time_window"},
		),

		TaskThroughputPerSecond: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "task_throughput_per_second",
				Help:      "Task processing throughput per second",
			},
			[]string{"task_type", "worker_pool"},
		),

		IntegrationUpStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "integration_up_status",
				Help:      "Integration service status (1 = up, 0 = down)",
			},
			[]string{"service", "endpoint"},
		),

		IntegrationAPILatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "integration_api_latency_seconds",
				Help:      "Integration API latency",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 12),
			},
			[]string{"service", "method", "endpoint"},
		),

		IntegrationRateLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "business",
				Name:      "integration_rate_limit_remaining",
				Help:      "Remaining API rate limit for integrations",
			},
			[]string{"service", "token_type"},
		),
	}

	// Register all business metrics
	cmm.registry.MustRegister(
		cmm.businessMetrics.RepoCloneTotal,
		cmm.businessMetrics.RepoCloneDuration,
		cmm.businessMetrics.RepoSyncOperations,
		cmm.businessMetrics.RepoSizeBytes,
		cmm.businessMetrics.OrganizationsTotal,
		cmm.businessMetrics.ProjectsActiveTotal,
		cmm.businessMetrics.UsersActiveTotal,
		cmm.businessMetrics.TasksCompletedTotal,
		cmm.businessMetrics.TaskFailureRatePercent,
		cmm.businessMetrics.TaskThroughputPerSecond,
		cmm.businessMetrics.IntegrationUpStatus,
		cmm.businessMetrics.IntegrationAPILatency,
		cmm.businessMetrics.IntegrationRateLimit,
	)
}

// initializePerformanceMetrics initializes performance metrics
func (cmm *CustomMetricsManager) initializePerformanceMetrics() {
	cmm.performanceMetrics = &PerformanceMetrics{
		CPUUtilizationPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "cpu_utilization_percent",
				Help:      "CPU utilization percentage",
			},
			[]string{"core", "type"},
		),

		MemoryUtilizationPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "memory_utilization_percent",
				Help:      "Memory utilization percentage",
			},
			[]string{"type"},
		),

		DiskIOOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "disk_io_operations_total",
				Help:      "Total disk I/O operations",
			},
			[]string{"device", "operation"},
		),

		DiskIOBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "disk_io_bytes_total",
				Help:      "Total disk I/O bytes",
			},
			[]string{"device", "direction"},
		),

		NetworkIOBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "network_io_bytes_total",
				Help:      "Total network I/O bytes",
			},
			[]string{"interface", "direction"},
		),

		GoroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "goroutine_count",
				Help:      "Number of goroutines",
			},
		),

		GCDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "gc_duration_seconds",
				Help:      "Garbage collection duration",
				Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
			},
			[]string{"gc_type"},
		),

		HeapAllocBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "heap_alloc_bytes",
				Help:      "Heap allocated bytes",
			},
		),

		DatabaseConnectionsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "database_connections_active",
				Help:      "Active database connections",
			},
			[]string{"database", "pool"},
		),

		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "database_query_duration_seconds",
				Help:      "Database query duration",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
			},
			[]string{"database", "operation", "table"},
		),

		DatabaseConnectionPoolSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "database_connection_pool_size",
				Help:      "Database connection pool size",
			},
			[]string{"database", "pool", "status"},
		),

		CacheHitRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "cache_hit_ratio",
				Help:      "Cache hit ratio",
			},
			[]string{"cache_type", "instance"},
		),

		CacheOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "cache_operations_total",
				Help:      "Total cache operations",
			},
			[]string{"cache_type", "operation", "status"},
		),

		CacheEvictionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "cache_evictions_total",
				Help:      "Total cache evictions",
			},
			[]string{"cache_type", "reason"},
		),

		QueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "queue_depth",
				Help:      "Queue depth",
			},
			[]string{"queue", "priority"},
		),

		QueueProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "queue_processing_duration_seconds",
				Help:      "Queue processing duration",
				Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
			},
			[]string{"queue", "job_type"},
		),

		QueueThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "performance",
				Name:      "queue_throughput",
				Help:      "Queue throughput (items per second)",
			},
			[]string{"queue", "worker"},
		),
	}

	// Register all performance metrics
	cmm.registry.MustRegister(
		cmm.performanceMetrics.CPUUtilizationPercent,
		cmm.performanceMetrics.MemoryUtilizationPercent,
		cmm.performanceMetrics.DiskIOOperationsTotal,
		cmm.performanceMetrics.DiskIOBytesTotal,
		cmm.performanceMetrics.NetworkIOBytesTotal,
		cmm.performanceMetrics.GoroutineCount,
		cmm.performanceMetrics.GCDuration,
		cmm.performanceMetrics.HeapAllocBytes,
		cmm.performanceMetrics.DatabaseConnectionsActive,
		cmm.performanceMetrics.DatabaseQueryDuration,
		cmm.performanceMetrics.DatabaseConnectionPoolSize,
		cmm.performanceMetrics.CacheHitRatio,
		cmm.performanceMetrics.CacheOperationsTotal,
		cmm.performanceMetrics.CacheEvictionsTotal,
		cmm.performanceMetrics.QueueDepth,
		cmm.performanceMetrics.QueueProcessingDuration,
		cmm.performanceMetrics.QueueThroughput,
	)
}

// initializeUsageMetrics initializes usage metrics
func (cmm *CustomMetricsManager) initializeUsageMetrics() {
	cmm.usageMetrics = &UsageMetrics{
		ActiveUsers5Min: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "active_users_5min",
				Help:      "Number of active users in last 5 minutes",
			},
			[]string{"organization", "role"},
		),

		ActiveUsers1Hour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "active_users_1hour",
				Help:      "Number of active users in last hour",
			},
			[]string{"organization", "role"},
		),

		ActiveUsers24Hour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "active_users_24hour",
				Help:      "Number of active users in last 24 hours",
			},
			[]string{"organization", "role"},
		),

		UserSessionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "user_session_duration_seconds",
				Help:      "User session duration",
				Buckets:   prometheus.ExponentialBuckets(60, 2, 10), // Start at 1 minute
			},
			[]string{"organization", "role", "session_type"},
		),

		FeatureUsageTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "feature_usage_total",
				Help:      "Total feature usage count",
			},
			[]string{"feature", "organization", "user_role"},
		),

		FeatureLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "feature_latency_seconds",
				Help:      "Feature execution latency",
				Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
			},
			[]string{"feature", "complexity"},
		),

		FeatureErrorRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "feature_error_rate_percent",
				Help:      "Feature error rate percentage",
			},
			[]string{"feature", "error_type"},
		),

		BandwidthUsageBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "bandwidth_usage_bytes_total",
				Help:      "Total bandwidth usage in bytes",
			},
			[]string{"organization", "operation", "direction"},
		),

		StorageUsageBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "storage_usage_bytes",
				Help:      "Storage usage in bytes",
			},
			[]string{"organization", "storage_type"},
		),

		ComputeHoursTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "compute_hours_total",
				Help:      "Total compute hours consumed",
			},
			[]string{"organization", "resource_type"},
		),

		APICallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "api_calls_total",
				Help:      "Total API calls",
			},
			[]string{"api_version", "endpoint", "method", "status"},
		),

		APILatencyPercentiles: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "api_latency_seconds",
				Help:      "API latency percentiles",
				Objectives: map[float64]float64{
					0.5:  0.01,
					0.9:  0.01,
					0.95: 0.005,
					0.99: 0.001,
				},
			},
			[]string{"api_version", "endpoint", "method"},
		),

		APIQuotaUsagePercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "api_quota_usage_percent",
				Help:      "API quota usage percentage",
			},
			[]string{"organization", "api_key", "quota_type"},
		),

		APIRetryAttemptsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "api_retry_attempts_total",
				Help:      "Total API retry attempts",
			},
			[]string{"endpoint", "reason", "final_status"},
		),

		RequestsByCountry: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "requests_by_country_total",
				Help:      "Total requests by country",
			},
			[]string{"country", "region"},
		),

		RequestsByTimezone: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gzh_manager",
				Subsystem: "usage",
				Name:      "requests_by_timezone_total",
				Help:      "Total requests by timezone",
			},
			[]string{"timezone", "hour_of_day"},
		),
	}

	// Register all usage metrics
	cmm.registry.MustRegister(
		cmm.usageMetrics.ActiveUsers5Min,
		cmm.usageMetrics.ActiveUsers1Hour,
		cmm.usageMetrics.ActiveUsers24Hour,
		cmm.usageMetrics.UserSessionDuration,
		cmm.usageMetrics.FeatureUsageTotal,
		cmm.usageMetrics.FeatureLatency,
		cmm.usageMetrics.FeatureErrorRate,
		cmm.usageMetrics.BandwidthUsageBytes,
		cmm.usageMetrics.StorageUsageBytes,
		cmm.usageMetrics.ComputeHoursTotal,
		cmm.usageMetrics.APICallsTotal,
		cmm.usageMetrics.APILatencyPercentiles,
		cmm.usageMetrics.APIQuotaUsagePercent,
		cmm.usageMetrics.APIRetryAttemptsTotal,
		cmm.usageMetrics.RequestsByCountry,
		cmm.usageMetrics.RequestsByTimezone,
	)
}

// Business metric recording methods

// RecordRepoClone records a repository clone operation
func (cmm *CustomMetricsManager) RecordRepoClone(organization, platform, status string, duration time.Duration, sizeCategory string) {
	cmm.businessMetrics.RepoCloneTotal.WithLabelValues(organization, platform, status).Inc()
	cmm.businessMetrics.RepoCloneDuration.WithLabelValues(organization, platform, sizeCategory).Observe(duration.Seconds())
}

// RecordRepoSync records a repository sync operation
func (cmm *CustomMetricsManager) RecordRepoSync(operation, status, organization string) {
	cmm.businessMetrics.RepoSyncOperations.WithLabelValues(operation, status, organization).Inc()
}

// SetRepoSize sets repository size metric
func (cmm *CustomMetricsManager) SetRepoSize(repository, organization string, sizeBytes float64) {
	cmm.businessMetrics.RepoSizeBytes.WithLabelValues(repository, organization).Set(sizeBytes)
}

// SetOrganizationsTotal sets total organizations metric
func (cmm *CustomMetricsManager) SetOrganizationsTotal(count float64) {
	cmm.businessMetrics.OrganizationsTotal.Set(count)
}

// SetProjectsActiveTotal sets active projects metric
func (cmm *CustomMetricsManager) SetProjectsActiveTotal(count float64) {
	cmm.businessMetrics.ProjectsActiveTotal.Set(count)
}

// SetUsersActiveTotal sets active users metric
func (cmm *CustomMetricsManager) SetUsersActiveTotal(count float64) {
	cmm.businessMetrics.UsersActiveTotal.Set(count)
}

// RecordTaskCompletion records task completion
func (cmm *CustomMetricsManager) RecordTaskCompletion(taskType, organization, outcome string) {
	cmm.businessMetrics.TasksCompletedTotal.WithLabelValues(taskType, organization, outcome).Inc()
}

// SetTaskFailureRate sets task failure rate
func (cmm *CustomMetricsManager) SetTaskFailureRate(taskType, timeWindow string, ratePercent float64) {
	cmm.businessMetrics.TaskFailureRatePercent.WithLabelValues(taskType, timeWindow).Set(ratePercent)
}

// SetTaskThroughput sets task throughput
func (cmm *CustomMetricsManager) SetTaskThroughput(taskType, workerPool string, throughput float64) {
	cmm.businessMetrics.TaskThroughputPerSecond.WithLabelValues(taskType, workerPool).Set(throughput)
}

// SetIntegrationStatus sets integration status
func (cmm *CustomMetricsManager) SetIntegrationStatus(service, endpoint string, isUp bool) {
	var status float64
	if isUp {
		status = 1
	}
	cmm.businessMetrics.IntegrationUpStatus.WithLabelValues(service, endpoint).Set(status)
}

// RecordIntegrationAPICall records integration API call latency
func (cmm *CustomMetricsManager) RecordIntegrationAPICall(service, method, endpoint string, duration time.Duration) {
	cmm.businessMetrics.IntegrationAPILatency.WithLabelValues(service, method, endpoint).Observe(duration.Seconds())
}

// SetIntegrationRateLimit sets integration rate limit remaining
func (cmm *CustomMetricsManager) SetIntegrationRateLimit(service, tokenType string, remaining float64) {
	cmm.businessMetrics.IntegrationRateLimit.WithLabelValues(service, tokenType).Set(remaining)
}

// Performance metric recording methods

// SetCPUUtilization sets CPU utilization
func (cmm *CustomMetricsManager) SetCPUUtilization(core, cpuType string, percent float64) {
	cmm.performanceMetrics.CPUUtilizationPercent.WithLabelValues(core, cpuType).Set(percent)
}

// SetMemoryUtilization sets memory utilization
func (cmm *CustomMetricsManager) SetMemoryUtilization(memType string, percent float64) {
	cmm.performanceMetrics.MemoryUtilizationPercent.WithLabelValues(memType).Set(percent)
}

// RecordDiskIO records disk I/O operations
func (cmm *CustomMetricsManager) RecordDiskIO(device, operation string, count float64, bytes float64, direction string) {
	cmm.performanceMetrics.DiskIOOperationsTotal.WithLabelValues(device, operation).Add(count)
	cmm.performanceMetrics.DiskIOBytesTotal.WithLabelValues(device, direction).Add(bytes)
}

// RecordNetworkIO records network I/O
func (cmm *CustomMetricsManager) RecordNetworkIO(iface, direction string, bytes float64) {
	cmm.performanceMetrics.NetworkIOBytesTotal.WithLabelValues(iface, direction).Add(bytes)
}

// SetGoroutineCount sets goroutine count
func (cmm *CustomMetricsManager) SetGoroutineCount(count float64) {
	cmm.performanceMetrics.GoroutineCount.Set(count)
}

// RecordGCDuration records garbage collection duration
func (cmm *CustomMetricsManager) RecordGCDuration(gcType string, duration time.Duration) {
	cmm.performanceMetrics.GCDuration.WithLabelValues(gcType).Observe(duration.Seconds())
}

// SetHeapAlloc sets heap allocated bytes
func (cmm *CustomMetricsManager) SetHeapAlloc(bytes float64) {
	cmm.performanceMetrics.HeapAllocBytes.Set(bytes)
}

// Usage metric recording methods

// SetActiveUsers sets active user counts
func (cmm *CustomMetricsManager) SetActiveUsers(organization, role string, users5min, users1hour, users24hour float64) {
	cmm.usageMetrics.ActiveUsers5Min.WithLabelValues(organization, role).Set(users5min)
	cmm.usageMetrics.ActiveUsers1Hour.WithLabelValues(organization, role).Set(users1hour)
	cmm.usageMetrics.ActiveUsers24Hour.WithLabelValues(organization, role).Set(users24hour)
}

// RecordUserSession records user session duration
func (cmm *CustomMetricsManager) RecordUserSession(organization, role, sessionType string, duration time.Duration) {
	cmm.usageMetrics.UserSessionDuration.WithLabelValues(organization, role, sessionType).Observe(duration.Seconds())
}

// RecordFeatureUsage records feature usage
func (cmm *CustomMetricsManager) RecordFeatureUsage(feature, organization, userRole string, duration time.Duration, complexity string) {
	cmm.usageMetrics.FeatureUsageTotal.WithLabelValues(feature, organization, userRole).Inc()
	cmm.usageMetrics.FeatureLatency.WithLabelValues(feature, complexity).Observe(duration.Seconds())
}

// SetFeatureErrorRate sets feature error rate
func (cmm *CustomMetricsManager) SetFeatureErrorRate(feature, errorType string, ratePercent float64) {
	cmm.usageMetrics.FeatureErrorRate.WithLabelValues(feature, errorType).Set(ratePercent)
}

// RecordResourceUsage records resource usage
func (cmm *CustomMetricsManager) RecordResourceUsage(organization, operation, direction string, bandwidthBytes float64, storageType string, storageBytes float64, resourceType string, computeHours float64) {
	cmm.usageMetrics.BandwidthUsageBytes.WithLabelValues(organization, operation, direction).Add(bandwidthBytes)
	cmm.usageMetrics.StorageUsageBytes.WithLabelValues(organization, storageType).Set(storageBytes)
	cmm.usageMetrics.ComputeHoursTotal.WithLabelValues(organization, resourceType).Add(computeHours)
}

// RecordAPICall records API call
func (cmm *CustomMetricsManager) RecordAPICall(apiVersion, endpoint, method, status string, duration time.Duration) {
	cmm.usageMetrics.APICallsTotal.WithLabelValues(apiVersion, endpoint, method, status).Inc()
	cmm.usageMetrics.APILatencyPercentiles.WithLabelValues(apiVersion, endpoint, method).Observe(duration.Seconds())
}

// SetAPIQuotaUsage sets API quota usage
func (cmm *CustomMetricsManager) SetAPIQuotaUsage(organization, apiKey, quotaType string, usagePercent float64) {
	cmm.usageMetrics.APIQuotaUsagePercent.WithLabelValues(organization, apiKey, quotaType).Set(usagePercent)
}

// RecordAPIRetry records API retry attempt
func (cmm *CustomMetricsManager) RecordAPIRetry(endpoint, reason, finalStatus string) {
	cmm.usageMetrics.APIRetryAttemptsTotal.WithLabelValues(endpoint, reason, finalStatus).Inc()
}

// RecordGeoRequest records geographical request
func (cmm *CustomMetricsManager) RecordGeoRequest(country, region, timezone, hourOfDay string) {
	cmm.usageMetrics.RequestsByCountry.WithLabelValues(country, region).Inc()
	cmm.usageMetrics.RequestsByTimezone.WithLabelValues(timezone, hourOfDay).Inc()
}

// Custom metric management methods

// CreateCustomCounter creates a custom counter metric
func (cmm *CustomMetricsManager) CreateCustomCounter(name, help string, labels []string, constLabels map[string]string) error {
	cmm.mutex.Lock()
	defer cmm.mutex.Unlock()

	if _, exists := cmm.customCounters[name]; exists {
		return fmt.Errorf("counter metric %s already exists", name)
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "gzh_manager",
			Subsystem:   "custom",
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		labels,
	)

	if err := cmm.registry.Register(counter); err != nil {
		return fmt.Errorf("failed to register counter %s: %w", name, err)
	}

	cmm.customCounters[name] = counter
	return nil
}

// CreateCustomGauge creates a custom gauge metric
func (cmm *CustomMetricsManager) CreateCustomGauge(name, help string, labels []string, constLabels map[string]string) error {
	cmm.mutex.Lock()
	defer cmm.mutex.Unlock()

	if _, exists := cmm.customGauges[name]; exists {
		return fmt.Errorf("gauge metric %s already exists", name)
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "gzh_manager",
			Subsystem:   "custom",
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		labels,
	)

	if err := cmm.registry.Register(gauge); err != nil {
		return fmt.Errorf("failed to register gauge %s: %w", name, err)
	}

	cmm.customGauges[name] = gauge
	return nil
}

// CreateCustomHistogram creates a custom histogram metric
func (cmm *CustomMetricsManager) CreateCustomHistogram(name, help string, labels []string, buckets []float64, constLabels map[string]string) error {
	cmm.mutex.Lock()
	defer cmm.mutex.Unlock()

	if _, exists := cmm.customHistograms[name]; exists {
		return fmt.Errorf("histogram metric %s already exists", name)
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   "gzh_manager",
			Subsystem:   "custom",
			Name:        name,
			Help:        help,
			Buckets:     buckets,
			ConstLabels: constLabels,
		},
		labels,
	)

	if err := cmm.registry.Register(histogram); err != nil {
		return fmt.Errorf("failed to register histogram %s: %w", name, err)
	}

	cmm.customHistograms[name] = histogram
	return nil
}

// CreateCustomSummary creates a custom summary metric
func (cmm *CustomMetricsManager) CreateCustomSummary(name, help string, labels []string, objectives map[float64]float64, constLabels map[string]string) error {
	cmm.mutex.Lock()
	defer cmm.mutex.Unlock()

	if _, exists := cmm.customSummaries[name]; exists {
		return fmt.Errorf("summary metric %s already exists", name)
	}

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   "gzh_manager",
			Subsystem:   "custom",
			Name:        name,
			Help:        help,
			Objectives:  objectives,
			ConstLabels: constLabels,
		},
		labels,
	)

	if err := cmm.registry.Register(summary); err != nil {
		return fmt.Errorf("failed to register summary %s: %w", name, err)
	}

	cmm.customSummaries[name] = summary
	return nil
}

// GetCustomCounter gets a custom counter metric
func (cmm *CustomMetricsManager) GetCustomCounter(name string) (*prometheus.CounterVec, error) {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	counter, exists := cmm.customCounters[name]
	if !exists {
		return nil, fmt.Errorf("counter metric %s not found", name)
	}

	return counter, nil
}

// GetCustomGauge gets a custom gauge metric
func (cmm *CustomMetricsManager) GetCustomGauge(name string) (*prometheus.GaugeVec, error) {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	gauge, exists := cmm.customGauges[name]
	if !exists {
		return nil, fmt.Errorf("gauge metric %s not found", name)
	}

	return gauge, nil
}

// GetCustomHistogram gets a custom histogram metric
func (cmm *CustomMetricsManager) GetCustomHistogram(name string) (*prometheus.HistogramVec, error) {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	histogram, exists := cmm.customHistograms[name]
	if !exists {
		return nil, fmt.Errorf("histogram metric %s not found", name)
	}

	return histogram, nil
}

// GetCustomSummary gets a custom summary metric
func (cmm *CustomMetricsManager) GetCustomSummary(name string) (*prometheus.SummaryVec, error) {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	summary, exists := cmm.customSummaries[name]
	if !exists {
		return nil, fmt.Errorf("summary metric %s not found", name)
	}

	return summary, nil
}

// DeleteCustomMetric deletes a custom metric
func (cmm *CustomMetricsManager) DeleteCustomMetric(name string) error {
	cmm.mutex.Lock()
	defer cmm.mutex.Unlock()

	// Try to unregister from all metric types
	if counter, exists := cmm.customCounters[name]; exists {
		cmm.registry.Unregister(counter)
		delete(cmm.customCounters, name)
		return nil
	}

	if gauge, exists := cmm.customGauges[name]; exists {
		cmm.registry.Unregister(gauge)
		delete(cmm.customGauges, name)
		return nil
	}

	if histogram, exists := cmm.customHistograms[name]; exists {
		cmm.registry.Unregister(histogram)
		delete(cmm.customHistograms, name)
		return nil
	}

	if summary, exists := cmm.customSummaries[name]; exists {
		cmm.registry.Unregister(summary)
		delete(cmm.customSummaries, name)
		return nil
	}

	return fmt.Errorf("metric %s not found", name)
}

// ListCustomMetrics lists all custom metrics
func (cmm *CustomMetricsManager) ListCustomMetrics() map[string]string {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	metrics := make(map[string]string)

	for name := range cmm.customCounters {
		metrics[name] = "counter"
	}

	for name := range cmm.customGauges {
		metrics[name] = "gauge"
	}

	for name := range cmm.customHistograms {
		metrics[name] = "histogram"
	}

	for name := range cmm.customSummaries {
		metrics[name] = "summary"
	}

	return metrics
}

// GetMetricsSummary returns a summary of all custom metrics
func (cmm *CustomMetricsManager) GetMetricsSummary() map[string]interface{} {
	cmm.mutex.RLock()
	defer cmm.mutex.RUnlock()

	return map[string]interface{}{
		"business_metrics": map[string]interface{}{
			"repo_operations": "Repository clone and sync operations",
			"organizations":   "Organization and project management",
			"task_execution":  "Task completion and throughput metrics",
			"integrations":    "External service integration health",
		},
		"performance_metrics": map[string]interface{}{
			"system_resources": "CPU, memory, disk, and network utilization",
			"application":      "Goroutines, GC, and heap allocation",
			"database":         "Connection pools and query performance",
			"cache":            "Cache hit ratios and operations",
			"queue":            "Queue depth and processing metrics",
		},
		"usage_metrics": map[string]interface{}{
			"user_activity":  "Active user counts and session tracking",
			"feature_usage":  "Feature utilization and performance",
			"resource_usage": "Bandwidth, storage, and compute consumption",
			"api_usage":      "API calls, quotas, and geographical distribution",
		},
		"custom_metrics": map[string]interface{}{
			"counters":   len(cmm.customCounters),
			"gauges":     len(cmm.customGauges),
			"histograms": len(cmm.customHistograms),
			"summaries":  len(cmm.customSummaries),
		},
	}
}

// Start starts background metric collection routines
func (cmm *CustomMetricsManager) Start(ctx context.Context) error {
	cmm.logger.Info("Starting custom metrics manager")

	// Start metric collection goroutines
	go cmm.collectSystemMetrics(ctx)
	go cmm.collectRuntimeMetrics(ctx)

	return nil
}

// Stop stops metric collection
func (cmm *CustomMetricsManager) Stop() error {
	cmm.logger.Info("Stopping custom metrics manager")
	return nil
}

// collectSystemMetrics collects system-level metrics
func (cmm *CustomMetricsManager) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// This is a placeholder - in production you would integrate with actual system monitoring
			cmm.SetCPUUtilization("all", "user", 25.5)
			cmm.SetCPUUtilization("all", "system", 8.2)
			cmm.SetMemoryUtilization("heap", 45.7)
			cmm.SetMemoryUtilization("stack", 12.3)
		}
	}
}

// collectRuntimeMetrics collects Go runtime metrics
func (cmm *CustomMetricsManager) collectRuntimeMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			cmm.SetGoroutineCount(float64(runtime.NumGoroutine()))
			cmm.SetHeapAlloc(float64(memStats.Alloc))
		}
	}
}

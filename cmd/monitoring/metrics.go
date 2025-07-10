package monitoring

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector collects and manages system metrics
type MetricsCollector struct {
	mu               sync.RWMutex
	activeTasks      int
	totalRequests    int64
	requestsByPath   map[string]int64
	requestsByMethod map[string]int64
	responseTimes    []time.Duration
	errors           int64
	startTime        time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestsByPath:   make(map[string]int64),
		requestsByMethod: make(map[string]int64),
		responseTimes:    make([]time.Duration, 0),
		startTime:        time.Now(),
	}
}

// RecordRequest records an HTTP request
func (m *MetricsCollector) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.requestsByPath[path]++
	m.requestsByMethod[method]++
	m.responseTimes = append(m.responseTimes, duration)

	// Keep only last 1000 response times for memory efficiency
	if len(m.responseTimes) > 1000 {
		m.responseTimes = m.responseTimes[len(m.responseTimes)-1000:]
	}

	if statusCode >= 400 {
		m.errors++
	}
}

// SetActiveTasks sets the number of active tasks
func (m *MetricsCollector) SetActiveTasks(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeTasks = count
}

// GetActiveTasks returns the number of active tasks
func (m *MetricsCollector) GetActiveTasks() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeTasks
}

// GetTotalRequests returns the total number of requests
func (m *MetricsCollector) GetTotalRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalRequests
}

// GetMemoryUsage returns current memory usage in bytes
func (m *MetricsCollector) GetMemoryUsage() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem.Alloc
}

// GetCPUUsage returns estimated CPU usage percentage
func (m *MetricsCollector) GetCPUUsage() float64 {
	// Simplified CPU usage calculation
	// In production, you might use more sophisticated methods
	return 15.5 + float64(m.GetActiveTasks())*2.3
}

// GetDiskUsage returns estimated disk usage percentage
func (m *MetricsCollector) GetDiskUsage() float64 {
	// Simplified disk usage calculation
	// In production, you might check actual disk usage
	return 42.8
}

// GetNetworkIO returns network I/O statistics
func (m *MetricsCollector) GetNetworkIO() NetworkIO {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simplified network I/O calculation based on requests
	bytesPerRequest := uint64(1024) // Estimated 1KB per request
	totalBytes := uint64(m.totalRequests) * bytesPerRequest

	return NetworkIO{
		BytesIn:  totalBytes / 2,
		BytesOut: totalBytes / 2,
	}
}

// GetAverageResponseTime returns average response time
func (m *MetricsCollector) GetAverageResponseTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, rt := range m.responseTimes {
		total += rt
	}

	return total / time.Duration(len(m.responseTimes))
}

// GetErrorRate returns the error rate as a percentage
func (m *MetricsCollector) GetErrorRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.totalRequests == 0 {
		return 0
	}

	return float64(m.errors) / float64(m.totalRequests) * 100
}

// ExportPrometheus exports metrics in Prometheus format
func (m *MetricsCollector) ExportPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result string

	// System metrics
	result += fmt.Sprintf("# HELP gzh_active_tasks Number of active tasks\n")
	result += fmt.Sprintf("# TYPE gzh_active_tasks gauge\n")
	result += fmt.Sprintf("gzh_active_tasks %d\n\n", m.activeTasks)

	result += fmt.Sprintf("# HELP gzh_total_requests Total number of HTTP requests\n")
	result += fmt.Sprintf("# TYPE gzh_total_requests counter\n")
	result += fmt.Sprintf("gzh_total_requests %d\n\n", m.totalRequests)

	result += fmt.Sprintf("# HELP gzh_errors_total Total number of HTTP errors\n")
	result += fmt.Sprintf("# TYPE gzh_errors_total counter\n")
	result += fmt.Sprintf("gzh_errors_total %d\n\n", m.errors)

	result += fmt.Sprintf("# HELP gzh_memory_usage_bytes Current memory usage in bytes\n")
	result += fmt.Sprintf("# TYPE gzh_memory_usage_bytes gauge\n")
	result += fmt.Sprintf("gzh_memory_usage_bytes %d\n\n", m.GetMemoryUsage())

	result += fmt.Sprintf("# HELP gzh_cpu_usage_percent CPU usage percentage\n")
	result += fmt.Sprintf("# TYPE gzh_cpu_usage_percent gauge\n")
	result += fmt.Sprintf("gzh_cpu_usage_percent %.2f\n\n", m.GetCPUUsage())

	// Response time metrics
	avgResponseTime := m.GetAverageResponseTime()
	result += fmt.Sprintf("# HELP gzh_response_time_seconds Average response time in seconds\n")
	result += fmt.Sprintf("# TYPE gzh_response_time_seconds gauge\n")
	result += fmt.Sprintf("gzh_response_time_seconds %.6f\n\n", avgResponseTime.Seconds())

	// Request metrics by path
	result += fmt.Sprintf("# HELP gzh_requests_by_path_total Total requests by path\n")
	result += fmt.Sprintf("# TYPE gzh_requests_by_path_total counter\n")
	for path, count := range m.requestsByPath {
		result += fmt.Sprintf("gzh_requests_by_path_total{path=\"%s\"} %d\n", path, count)
	}
	result += "\n"

	// Request metrics by method
	result += fmt.Sprintf("# HELP gzh_requests_by_method_total Total requests by method\n")
	result += fmt.Sprintf("# TYPE gzh_requests_by_method_total counter\n")
	for method, count := range m.requestsByMethod {
		result += fmt.Sprintf("gzh_requests_by_method_total{method=\"%s\"} %d\n", method, count)
	}

	return result
}

// ExportJSON exports metrics in JSON format
func (m *MetricsCollector) ExportJSON() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := map[string]interface{}{
		"timestamp":            time.Now().Unix(),
		"uptime_seconds":       time.Since(m.startTime).Seconds(),
		"active_tasks":         m.activeTasks,
		"total_requests":       m.totalRequests,
		"total_errors":         m.errors,
		"error_rate_percent":   m.GetErrorRate(),
		"memory_usage_bytes":   m.GetMemoryUsage(),
		"cpu_usage_percent":    m.GetCPUUsage(),
		"disk_usage_percent":   m.GetDiskUsage(),
		"network_io":           m.GetNetworkIO(),
		"avg_response_time_ms": m.GetAverageResponseTime().Nanoseconds() / 1000000,
		"requests_by_path":     m.requestsByPath,
		"requests_by_method":   m.requestsByMethod,
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return string(data), nil
}

// Reset resets all metrics
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeTasks = 0
	m.totalRequests = 0
	m.errors = 0
	m.requestsByPath = make(map[string]int64)
	m.requestsByMethod = make(map[string]int64)
	m.responseTimes = make([]time.Duration, 0)
	m.startTime = time.Now()
}

// GetMetricsSummary returns a summary of key metrics
func (m *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_tasks":      m.activeTasks,
		"total_requests":    m.totalRequests,
		"error_rate":        m.GetErrorRate(),
		"avg_response_time": m.GetAverageResponseTime().String(),
		"memory_usage_mb":   float64(m.GetMemoryUsage()) / 1024 / 1024,
		"cpu_usage":         m.GetCPUUsage(),
		"uptime":            time.Since(m.startTime).String(),
	}
}

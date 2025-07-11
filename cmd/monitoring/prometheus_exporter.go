package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusExporter provides metrics export to Prometheus
type PrometheusExporter struct {
	logger           *zap.Logger
	registry         *prometheus.Registry
	metricsCollector *MetricsCollector
	server           *http.Server

	// Core metrics
	taskCounter     *prometheus.CounterVec
	taskDuration    *prometheus.HistogramVec
	systemCPU       prometheus.Gauge
	systemMemory    prometheus.Gauge
	alertCounter    *prometheus.CounterVec
	ruleEvaluations *prometheus.CounterVec
	httpRequests    *prometheus.CounterVec
	httpDuration    *prometheus.HistogramVec

	// Custom business metrics
	customMetrics    map[string]prometheus.Metric
	customCounters   map[string]*prometheus.CounterVec
	customGauges     map[string]*prometheus.GaugeVec
	customHistograms map[string]*prometheus.HistogramVec

	mutex sync.RWMutex
}

// PrometheusConfig represents Prometheus exporter configuration
type PrometheusConfig struct {
	Enabled          bool                    `yaml:"enabled" json:"enabled"`
	ListenAddress    string                  `yaml:"listen_address" json:"listen_address"`
	MetricsPath      string                  `yaml:"metrics_path" json:"metrics_path"`
	Namespace        string                  `yaml:"namespace" json:"namespace"`
	Subsystem        string                  `yaml:"subsystem" json:"subsystem"`
	Labels           map[string]string       `yaml:"labels" json:"labels"`
	ServiceDiscovery *ServiceDiscoveryConfig `yaml:"service_discovery" json:"service_discovery"`
}

// ServiceDiscoveryConfig represents service discovery configuration
type ServiceDiscoveryConfig struct {
	Enabled        bool                   `yaml:"enabled" json:"enabled"`
	Type           string                 `yaml:"type" json:"type"` // "static", "kubernetes", "consul", "dns"
	Config         map[string]interface{} `yaml:"config" json:"config"`
	ScrapeInterval time.Duration          `yaml:"scrape_interval" json:"scrape_interval"`
	Labels         map[string]string      `yaml:"labels" json:"labels"`
}

// CustomMetricDefinition represents a custom metric definition
type CustomMetricDefinition struct {
	Name        string              `yaml:"name" json:"name"`
	Type        string              `yaml:"type" json:"type"` // "counter", "gauge", "histogram", "summary"
	Help        string              `yaml:"help" json:"help"`
	Labels      []string            `yaml:"labels" json:"labels"`
	Buckets     []float64           `yaml:"buckets,omitempty" json:"buckets,omitempty"`       // For histograms
	Objectives  map[float64]float64 `yaml:"objectives,omitempty" json:"objectives,omitempty"` // For summaries
	ConstLabels map[string]string   `yaml:"const_labels,omitempty" json:"const_labels,omitempty"`
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(logger *zap.Logger, config *PrometheusConfig, metricsCollector *MetricsCollector) *PrometheusExporter {
	registry := prometheus.NewRegistry()

	exporter := &PrometheusExporter{
		logger:           logger,
		registry:         registry,
		metricsCollector: metricsCollector,
		customMetrics:    make(map[string]prometheus.Metric),
		customCounters:   make(map[string]*prometheus.CounterVec),
		customGauges:     make(map[string]*prometheus.GaugeVec),
		customHistograms: make(map[string]*prometheus.HistogramVec),
	}

	exporter.initializeMetrics(config)
	exporter.registerMetrics()

	if config.Enabled {
		exporter.setupHTTPServer(config)
	}

	return exporter
}

// initializeMetrics initializes all core Prometheus metrics
func (pe *PrometheusExporter) initializeMetrics(config *PrometheusConfig) {
	namespace := config.Namespace
	if namespace == "" {
		namespace = "gzh_manager"
	}

	subsystem := config.Subsystem
	if subsystem == "" {
		subsystem = "monitoring"
	}

	// Task metrics
	pe.taskCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "tasks_total",
			Help:      "Total number of tasks executed",
		},
		[]string{"type", "status", "organization"},
	)

	pe.taskDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "task_duration_seconds",
			Help:      "Time spent executing tasks",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"type", "organization"},
	)

	// System metrics
	pe.systemCPU = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cpu_usage_percent",
			Help:      "Current CPU usage percentage",
		},
	)

	pe.systemMemory = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		},
	)

	// Alert metrics
	pe.alertCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "alerts_total",
			Help:      "Total number of alerts fired",
		},
		[]string{"severity", "status", "rule_id"},
	)

	pe.ruleEvaluations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "rule_evaluations_total",
			Help:      "Total number of rule evaluations",
		},
		[]string{"rule_id", "result"},
	)

	// HTTP metrics
	pe.httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	pe.httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "Time spent processing HTTP requests",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "endpoint"},
	)
}

// registerMetrics registers all metrics with the Prometheus registry
func (pe *PrometheusExporter) registerMetrics() {
	pe.registry.MustRegister(
		pe.taskCounter,
		pe.taskDuration,
		pe.systemCPU,
		pe.systemMemory,
		pe.alertCounter,
		pe.ruleEvaluations,
		pe.httpRequests,
		pe.httpDuration,
	)

	// Add Go runtime metrics
	pe.registry.MustRegister(prometheus.NewGoCollector())
	pe.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// setupHTTPServer sets up the HTTP server for metrics exposition
func (pe *PrometheusExporter) setupHTTPServer(config *PrometheusConfig) {
	metricsPath := config.MetricsPath
	if metricsPath == "" {
		metricsPath = "/metrics"
	}

	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(pe.registry, promhttp.HandlerOpts{}))

	pe.server = &http.Server{
		Addr:    config.ListenAddress,
		Handler: mux,
	}
}

// Start starts the Prometheus metrics server
func (pe *PrometheusExporter) Start(ctx context.Context) error {
	if pe.server == nil {
		return nil // Not enabled
	}

	pe.logger.Info("Starting Prometheus metrics server",
		zap.String("address", pe.server.Addr))

	go func() {
		if err := pe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pe.logger.Error("Prometheus server failed", zap.Error(err))
		}
	}()

	// Start metrics collection
	go pe.collectMetrics(ctx)

	return nil
}

// Stop stops the Prometheus metrics server
func (pe *PrometheusExporter) Stop(ctx context.Context) error {
	if pe.server == nil {
		return nil
	}

	pe.logger.Info("Stopping Prometheus metrics server")
	return pe.server.Shutdown(ctx)
}

// collectMetrics continuously collects and updates metrics
func (pe *PrometheusExporter) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pe.updateSystemMetrics()
		}
	}
}

// updateSystemMetrics updates system-level metrics
func (pe *PrometheusExporter) updateSystemMetrics() {
	if pe.metricsCollector == nil {
		return
	}

	pe.systemCPU.Set(pe.metricsCollector.GetCPUUsage())
	pe.systemMemory.Set(float64(pe.metricsCollector.GetMemoryUsage()))
}

// RecordTaskExecution records task execution metrics
func (pe *PrometheusExporter) RecordTaskExecution(taskType, status, organization string, duration time.Duration) {
	pe.taskCounter.WithLabelValues(taskType, status, organization).Inc()
	pe.taskDuration.WithLabelValues(taskType, organization).Observe(duration.Seconds())
}

// RecordAlert records alert metrics
func (pe *PrometheusExporter) RecordAlert(severity, status, ruleID string) {
	pe.alertCounter.WithLabelValues(severity, status, ruleID).Inc()
}

// RecordRuleEvaluation records rule evaluation metrics
func (pe *PrometheusExporter) RecordRuleEvaluation(ruleID string, matched bool) {
	result := "false"
	if matched {
		result = "true"
	}
	pe.ruleEvaluations.WithLabelValues(ruleID, result).Inc()
}

// RecordHTTPRequest records HTTP request metrics
func (pe *PrometheusExporter) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	pe.httpRequests.WithLabelValues(method, endpoint, status).Inc()
	pe.httpDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RegisterCustomMetric registers a custom metric
func (pe *PrometheusExporter) RegisterCustomMetric(definition *CustomMetricDefinition) error {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	switch definition.Type {
	case "counter":
		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        definition.Name,
				Help:        definition.Help,
				ConstLabels: definition.ConstLabels,
			},
			definition.Labels,
		)
		pe.customCounters[definition.Name] = counter
		pe.registry.MustRegister(counter)

	case "gauge":
		gauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        definition.Name,
				Help:        definition.Help,
				ConstLabels: definition.ConstLabels,
			},
			definition.Labels,
		)
		pe.customGauges[definition.Name] = gauge
		pe.registry.MustRegister(gauge)

	case "histogram":
		buckets := definition.Buckets
		if len(buckets) == 0 {
			buckets = prometheus.DefBuckets
		}
		histogram := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        definition.Name,
				Help:        definition.Help,
				Buckets:     buckets,
				ConstLabels: definition.ConstLabels,
			},
			definition.Labels,
		)
		pe.customHistograms[definition.Name] = histogram
		pe.registry.MustRegister(histogram)

	default:
		return fmt.Errorf("invalid metric type: %s", definition.Type)
	}

	pe.logger.Info("Registered custom metric",
		zap.String("name", definition.Name),
		zap.String("type", definition.Type))

	return nil
}

// IncrementCustomCounter increments a custom counter metric
func (pe *PrometheusExporter) IncrementCustomCounter(name string, labelValues ...string) error {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	counter, exists := pe.customCounters[name]
	if !exists {
		return fmt.Errorf("custom counter metric not found: %s", name)
	}

	counter.WithLabelValues(labelValues...).Inc()
	return nil
}

// SetCustomGauge sets a custom gauge metric value
func (pe *PrometheusExporter) SetCustomGauge(name string, value float64, labelValues ...string) error {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	gauge, exists := pe.customGauges[name]
	if !exists {
		return fmt.Errorf("custom gauge metric not found: %s", name)
	}

	gauge.WithLabelValues(labelValues...).Set(value)
	return nil
}

// ObserveCustomHistogram observes a value in a custom histogram metric
func (pe *PrometheusExporter) ObserveCustomHistogram(name string, value float64, labelValues ...string) error {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	histogram, exists := pe.customHistograms[name]
	if !exists {
		return fmt.Errorf("custom histogram metric not found: %s", name)
	}

	histogram.WithLabelValues(labelValues...).Observe(value)
	return nil
}

// GetMetrics returns the current metrics registry
func (pe *PrometheusExporter) GetMetrics() *prometheus.Registry {
	return pe.registry
}

// HealthCheck performs a health check of the exporter
func (pe *PrometheusExporter) HealthCheck() error {
	if pe.server == nil {
		return nil // Not enabled, always healthy
	}

	// Try to gather metrics to ensure registry is working
	_, err := pe.registry.Gather()
	return err
}

package monitoring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// CentralizedLogger provides centralized logging capabilities
type CentralizedLogger struct {
	logger     *zap.Logger
	config     *CentralizedLoggingConfig
	outputs    map[string]LogOutput
	processors map[string]LogProcessor
	shippers   map[string]LogShipper
	indexer    LogIndexer
	metrics    *LoggingMetrics
	mutex      sync.RWMutex

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// CentralizedLoggingConfig represents the centralized logging configuration
type CentralizedLoggingConfig struct {
	// Global logging settings
	Level        string            `yaml:"level" json:"level"`
	Format       string            `yaml:"format" json:"format"` // "json", "console", "structured"
	Directory    string            `yaml:"directory" json:"directory"`
	BaseFilename string            `yaml:"base_filename" json:"base_filename"`
	Labels       map[string]string `yaml:"labels" json:"labels"`

	// File rotation settings
	Rotation *RotationConfig `yaml:"rotation" json:"rotation"`

	// Output configurations
	Outputs map[string]*OutputConfig `yaml:"outputs" json:"outputs"`

	// Processing pipelines
	Processors map[string]*ProcessorConfig `yaml:"processors" json:"processors"`

	// Log shipping configurations
	Shippers map[string]*ShipperConfig `yaml:"shippers" json:"shippers"`

	// Search and indexing
	Indexing *IndexingConfig `yaml:"indexing" json:"indexing"`

	// Filtering and sampling
	Filters *FilterConfig `yaml:"filters" json:"filters"`

	// Performance settings
	BufferSize    int           `yaml:"buffer_size" json:"buffer_size"`
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`
	AsyncMode     bool          `yaml:"async_mode" json:"async_mode"`
}

// RotationConfig represents log rotation configuration
type RotationConfig struct {
	MaxSizeMB  int  `yaml:"max_size_mb" json:"max_size_mb"`
	MaxFiles   int  `yaml:"max_files" json:"max_files"`
	MaxAgeDays int  `yaml:"max_age_days" json:"max_age_days"`
	Compress   bool `yaml:"compress" json:"compress"`
	LocalTime  bool `yaml:"local_time" json:"local_time"`
}

// OutputConfig represents a log output configuration
type OutputConfig struct {
	Type     string                 `yaml:"type" json:"type"` // "file", "console", "syslog", "http"
	Format   string                 `yaml:"format" json:"format"`
	Level    string                 `yaml:"level" json:"level"`
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
	Settings map[string]interface{} `yaml:"settings" json:"settings"`
}

// ProcessorConfig represents a log processor configuration
type ProcessorConfig struct {
	Type     string                 `yaml:"type" json:"type"` // "filter", "transform", "enrich", "sample"
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
	Settings map[string]interface{} `yaml:"settings" json:"settings"`
}

// ShipperConfig represents a log shipper configuration
type ShipperConfig struct {
	Type     string                 `yaml:"type" json:"type"` // "elasticsearch", "loki", "fluentd", "http"
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
	Endpoint string                 `yaml:"endpoint" json:"endpoint"`
	Settings map[string]interface{} `yaml:"settings" json:"settings"`
}

// FilterConfig represents filtering and sampling configuration
type FilterConfig struct {
	MinLevel    string   `yaml:"min_level" json:"min_level"`
	ExcludeKeys []string `yaml:"exclude_keys" json:"exclude_keys"`
	IncludeKeys []string `yaml:"include_keys" json:"include_keys"`
	SampleRate  float64  `yaml:"sample_rate" json:"sample_rate"`
}

// IndexingConfig represents search indexing configuration
type IndexingConfig struct {
	Enabled   bool                   `yaml:"enabled" json:"enabled"`
	Type      string                 `yaml:"type" json:"type"` // "memory", "elasticsearch", "opensearch"
	IndexName string                 `yaml:"index_name" json:"index_name"`
	Settings  map[string]interface{} `yaml:"settings" json:"settings"`
	Mappings  *IndexMappings         `yaml:"mappings" json:"mappings"`
	Retention *RetentionPolicy       `yaml:"retention" json:"retention"`
	SearchAPI *SearchAPIConfig       `yaml:"search_api" json:"search_api"`
}

// SearchAPIConfig represents search API configuration
type SearchAPIConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	Address        string `yaml:"address" json:"address"`
	Port           int    `yaml:"port" json:"port"`
	MaxResultLimit int    `yaml:"max_result_limit" json:"max_result_limit"`
	TimeoutSeconds int    `yaml:"timeout_seconds" json:"timeout_seconds"`
	AuthEnabled    bool   `yaml:"auth_enabled" json:"auth_enabled"`
	CorsEnabled    bool   `yaml:"cors_enabled" json:"cors_enabled"`
}

// LogOutput represents a log output destination
type LogOutput interface {
	Write(entry *LogEntry) error
	Flush() error
	Close() error
	Name() string
}

// LogProcessor represents a log processing pipeline stage
type LogProcessor interface {
	Process(entry *LogEntry) (*LogEntry, error)
	Name() string
}

// LogShipper represents a log shipping mechanism
type LogShipper interface {
	Ship(entries []*LogEntry) error
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Logger    string                 `json:"logger"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Labels    map[string]string      `json:"labels,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Source    *LogSource             `json:"source,omitempty"`
}

// LogSource represents the source of a log entry
type LogSource struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
	Package  string `json:"package,omitempty"`
}

// LoggingMetrics represents metrics for the logging system
type LoggingMetrics struct {
	EntriesTotal       *prometheus.CounterVec
	EntriesProcessed   *prometheus.CounterVec
	EntriesShipped     *prometheus.CounterVec
	EntriesDropped     *prometheus.CounterVec
	ProcessingDuration *prometheus.HistogramVec
	ShippingDuration   *prometheus.HistogramVec
	BufferUtilization  *prometheus.GaugeVec
	ActiveOutputs      prometheus.Gauge
	ActiveShippers     prometheus.Gauge
	ErrorsTotal        *prometheus.CounterVec
}

// NewCentralizedLogger creates a new centralized logger
func NewCentralizedLogger(config *CentralizedLoggingConfig, registry *prometheus.Registry) (*CentralizedLogger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cl := &CentralizedLogger{
		config:     config,
		outputs:    make(map[string]LogOutput),
		processors: make(map[string]LogProcessor),
		shippers:   make(map[string]LogShipper),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize metrics
	cl.initializeMetrics(registry)

	// Initialize logger with centralized configuration
	if err := cl.initializeLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize outputs
	if err := cl.initializeOutputs(); err != nil {
		return nil, fmt.Errorf("failed to initialize outputs: %w", err)
	}

	// Initialize processors
	if err := cl.initializeProcessors(); err != nil {
		return nil, fmt.Errorf("failed to initialize processors: %w", err)
	}

	// Initialize shippers
	if err := cl.initializeShippers(); err != nil {
		return nil, fmt.Errorf("failed to initialize shippers: %w", err)
	}

	// Initialize indexer
	if err := cl.initializeIndexer(); err != nil {
		return nil, fmt.Errorf("failed to initialize indexer: %w", err)
	}

	// Start background processing
	cl.startBackgroundProcessing()

	return cl, nil
}

// initializeMetrics initializes logging metrics
func (cl *CentralizedLogger) initializeMetrics(registry *prometheus.Registry) {
	cl.metrics = &LoggingMetrics{
		EntriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gzh_log_entries_total",
				Help: "Total number of log entries processed",
			},
			[]string{"level", "logger", "output"},
		),

		EntriesProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gzh_log_entries_processed_total",
				Help: "Total number of log entries processed by processors",
			},
			[]string{"processor", "status"},
		),

		EntriesShipped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gzh_log_entries_shipped_total",
				Help: "Total number of log entries shipped to external systems",
			},
			[]string{"shipper", "destination"},
		),

		EntriesDropped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gzh_log_entries_dropped_total",
				Help: "Total number of log entries dropped",
			},
			[]string{"reason", "output"},
		),

		ProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gzh_log_processing_duration_seconds",
				Help:    "Duration of log processing operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "processor"},
		),

		ShippingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gzh_log_shipping_duration_seconds",
				Help:    "Duration of log shipping operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"shipper", "destination"},
		),

		BufferUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gzh_log_buffer_utilization_percent",
				Help: "Utilization percentage of log buffers",
			},
			[]string{"buffer_type"},
		),

		ActiveOutputs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gzh_log_active_outputs",
			Help: "Number of active log outputs",
		}),

		ActiveShippers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gzh_log_active_shippers",
			Help: "Number of active log shippers",
		}),

		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gzh_log_errors_total",
				Help: "Total number of logging errors",
			},
			[]string{"component", "error_type"},
		),
	}

	// Register metrics
	registry.MustRegister(
		cl.metrics.EntriesTotal,
		cl.metrics.EntriesProcessed,
		cl.metrics.EntriesShipped,
		cl.metrics.EntriesDropped,
		cl.metrics.ProcessingDuration,
		cl.metrics.ShippingDuration,
		cl.metrics.BufferUtilization,
		cl.metrics.ActiveOutputs,
		cl.metrics.ActiveShippers,
		cl.metrics.ErrorsTotal,
	)
}

// initializeLogger initializes the underlying zap logger
func (cl *CentralizedLogger) initializeLogger() error {
	level, err := zapcore.ParseLevel(cl.config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoders based on format
	var encoder zapcore.Encoder
	switch cl.config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	case "console":
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	default:
		// Structured format with custom encoding
		config := zap.NewProductionEncoderConfig()
		config.TimeKey = "timestamp"
		config.MessageKey = "message"
		config.LevelKey = "level"
		config.CallerKey = "caller"
		config.StacktraceKey = "stacktrace"
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(config)
	}

	// Create core with multiple outputs
	var cores []zapcore.Core

	// Add file output
	if cl.config.Directory != "" {
		fileWriter, err := cl.createFileWriter()
		if err != nil {
			return fmt.Errorf("failed to create file writer: %w", err)
		}
		cores = append(cores, zapcore.NewCore(encoder, fileWriter, level))
	}

	// Add console output (always enabled for development)
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))

	core := zapcore.NewTee(cores...)

	// Create logger with caller information
	cl.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// createFileWriter creates a file writer with rotation
func (cl *CentralizedLogger) createFileWriter() (zapcore.WriteSyncer, error) {
	if err := os.MkdirAll(cl.config.Directory, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	filename := filepath.Join(cl.config.Directory, cl.config.BaseFilename)

	rotation := cl.config.Rotation
	if rotation == nil {
		rotation = &RotationConfig{
			MaxSizeMB:  10,
			MaxFiles:   5,
			MaxAgeDays: 30,
			Compress:   true,
		}
	}

	lumberJack := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    rotation.MaxSizeMB,
		MaxBackups: rotation.MaxFiles,
		MaxAge:     rotation.MaxAgeDays,
		Compress:   rotation.Compress,
		LocalTime:  rotation.LocalTime,
	}

	return zapcore.AddSync(lumberJack), nil
}

// initializeOutputs initializes configured log outputs
func (cl *CentralizedLogger) initializeOutputs() error {
	for name, config := range cl.config.Outputs {
		if !config.Enabled {
			continue
		}

		output, err := cl.createOutput(name, config)
		if err != nil {
			cl.logger.Error("Failed to create output", zap.String("output", name), zap.Error(err))
			continue
		}

		cl.outputs[name] = output
		cl.metrics.ActiveOutputs.Inc()
	}

	return nil
}

// createOutput creates a specific log output
func (cl *CentralizedLogger) createOutput(name string, config *OutputConfig) (LogOutput, error) {
	switch config.Type {
	case "file":
		return NewFileOutput(name, config)
	case "console":
		return NewConsoleOutput(name, config)
	case "syslog":
		return NewSyslogOutput(name, config)
	case "http":
		return NewHTTPOutput(name, config)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", config.Type)
	}
}

// initializeProcessors initializes configured log processors
func (cl *CentralizedLogger) initializeProcessors() error {
	for name, config := range cl.config.Processors {
		if !config.Enabled {
			continue
		}

		processor, err := cl.createProcessor(name, config)
		if err != nil {
			cl.logger.Error("Failed to create processor", zap.String("processor", name), zap.Error(err))
			continue
		}

		cl.processors[name] = processor
	}

	return nil
}

// createProcessor creates a specific log processor
func (cl *CentralizedLogger) createProcessor(name string, config *ProcessorConfig) (LogProcessor, error) {
	switch config.Type {
	case "filter":
		return NewFilterProcessor(name, config)
	case "transform":
		return NewTransformProcessor(name, config)
	case "enrich":
		return NewEnrichProcessor(name, config)
	case "sample":
		return NewSampleProcessor(name, config)
	case "parse":
		return NewParseProcessor(name, config)
	default:
		return nil, fmt.Errorf("unsupported processor type: %s", config.Type)
	}
}

// initializeShippers initializes configured log shippers
func (cl *CentralizedLogger) initializeShippers() error {
	for name, config := range cl.config.Shippers {
		if !config.Enabled {
			continue
		}

		shipper, err := cl.createShipper(name, config)
		if err != nil {
			cl.logger.Error("Failed to create shipper", zap.String("shipper", name), zap.Error(err))
			continue
		}

		cl.shippers[name] = shipper
		cl.metrics.ActiveShippers.Inc()

		// Start shipper
		if err := shipper.Start(cl.ctx); err != nil {
			cl.logger.Error("Failed to start shipper", zap.String("shipper", name), zap.Error(err))
		}
	}

	return nil
}

// initializeIndexer initializes the log indexer
func (cl *CentralizedLogger) initializeIndexer() error {
	if cl.config.Indexing == nil || !cl.config.Indexing.Enabled {
		cl.logger.Info("Log indexing disabled")
		return nil
	}

	indexConfig := &IndexConfig{
		Name: cl.config.Indexing.IndexName,
		Settings: &IndexSettings{
			Shards:          1,
			Replicas:        0,
			RefreshInterval: "1s",
			MaxResultWindow: 10000,
		},
		Mappings:        cl.config.Indexing.Mappings,
		RetentionPolicy: cl.config.Indexing.Retention,
	}

	// Override with custom settings if provided
	if settings := cl.config.Indexing.Settings; settings != nil {
		if shards, ok := settings["shards"].(int); ok {
			indexConfig.Settings.Shards = shards
		}
		if replicas, ok := settings["replicas"].(int); ok {
			indexConfig.Settings.Replicas = replicas
		}
		if refresh, ok := settings["refresh_interval"].(string); ok {
			indexConfig.Settings.RefreshInterval = refresh
		}
	}

	switch cl.config.Indexing.Type {
	case "memory", "":
		cl.indexer = NewMemoryIndexer(cl.config.Indexing.IndexName, indexConfig)
	default:
		return fmt.Errorf("unsupported indexer type: %s", cl.config.Indexing.Type)
	}

	if err := cl.indexer.CreateIndex(cl.config.Indexing.IndexName, indexConfig); err != nil {
		return fmt.Errorf("failed to create search index: %w", err)
	}

	cl.logger.Info("Search indexer initialized",
		zap.String("type", cl.config.Indexing.Type),
		zap.String("index", cl.config.Indexing.IndexName))

	return nil
}

// createShipper creates a specific log shipper
func (cl *CentralizedLogger) createShipper(name string, config *ShipperConfig) (LogShipper, error) {
	switch config.Type {
	case "elasticsearch":
		return NewElasticsearchShipper(name, config)
	case "loki":
		return NewLokiShipper(name, config)
	case "fluentd":
		return NewFluentdShipper(name, config)
	case "http":
		return NewHTTPShipper(name, config)
	default:
		return nil, fmt.Errorf("unsupported shipper type: %s", config.Type)
	}
}

// startBackgroundProcessing starts background processing routines
func (cl *CentralizedLogger) startBackgroundProcessing() {
	// Start buffer flushing routine
	cl.wg.Add(1)
	go cl.flushBuffersRoutine()

	// Start metrics collection routine
	cl.wg.Add(1)
	go cl.metricsCollectionRoutine()
}

// flushBuffersRoutine periodically flushes log buffers
func (cl *CentralizedLogger) flushBuffersRoutine() {
	defer cl.wg.Done()

	ticker := time.NewTicker(cl.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cl.ctx.Done():
			return
		case <-ticker.C:
			cl.flushAllBuffers()
		}
	}
}

// metricsCollectionRoutine collects logging metrics
func (cl *CentralizedLogger) metricsCollectionRoutine() {
	defer cl.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cl.ctx.Done():
			return
		case <-ticker.C:
			cl.collectMetrics()
		}
	}
}

// flushAllBuffers flushes all output buffers
func (cl *CentralizedLogger) flushAllBuffers() {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	for name, output := range cl.outputs {
		if err := output.Flush(); err != nil {
			cl.logger.Error("Failed to flush output buffer",
				zap.String("output", name),
				zap.Error(err))
		}
	}
}

// collectMetrics collects logging system metrics
func (cl *CentralizedLogger) collectMetrics() {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	// Update active components count
	cl.metrics.ActiveOutputs.Set(float64(len(cl.outputs)))
	cl.metrics.ActiveShippers.Set(float64(len(cl.shippers)))
}

// Log logs an entry through the centralized logging system
func (cl *CentralizedLogger) Log(entry *LogEntry) error {
	start := time.Now()

	// Record entry metric
	cl.metrics.EntriesTotal.WithLabelValues(entry.Level, entry.Logger, "centralized").Inc()

	// Process through processors
	processedEntry := entry
	for name, processor := range cl.processors {
		var err error
		processingStart := time.Now()

		processedEntry, err = processor.Process(processedEntry)
		if err != nil {
			cl.metrics.EntriesProcessed.WithLabelValues(name, "error").Inc()
			cl.logger.Error("Processor error", zap.String("processor", name), zap.Error(err))
			continue
		}

		cl.metrics.ProcessingDuration.WithLabelValues("process", name).Observe(time.Since(processingStart).Seconds())
		cl.metrics.EntriesProcessed.WithLabelValues(name, "success").Inc()

		// If processor returns nil, entry is filtered out
		if processedEntry == nil {
			cl.metrics.EntriesDropped.WithLabelValues("filtered", name).Inc()
			return nil
		}
	}

	// Send to outputs
	for name, output := range cl.outputs {
		if err := output.Write(processedEntry); err != nil {
			cl.metrics.ErrorsTotal.WithLabelValues("output", "write").Inc()
			cl.logger.Error("Output write error", zap.String("output", name), zap.Error(err))
		}
	}

	// Index the entry for search
	if cl.indexer != nil {
		if err := cl.indexer.Index(processedEntry); err != nil {
			cl.metrics.ErrorsTotal.WithLabelValues("indexer", "index").Inc()
			cl.logger.Error("Indexing error", zap.Error(err))
		}
	}

	// Record processing duration
	cl.metrics.ProcessingDuration.WithLabelValues("total", "centralized").Observe(time.Since(start).Seconds())

	return nil
}

// GetLogger returns the underlying zap logger
func (cl *CentralizedLogger) GetLogger() *zap.Logger {
	return cl.logger
}

// AddLabels adds labels to the centralized logger context
func (cl *CentralizedLogger) AddLabels(labels map[string]string) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	if cl.config.Labels == nil {
		cl.config.Labels = make(map[string]string)
	}

	for k, v := range labels {
		cl.config.Labels[k] = v
	}
}

// CreateContextualLogger creates a contextual logger with additional fields
func (cl *CentralizedLogger) CreateContextualLogger(component string, fields map[string]interface{}) *zap.Logger {
	logger := cl.logger.Named(component)

	var zapFields []zap.Field
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return logger.With(zapFields...)
}

// Shutdown gracefully shuts down the centralized logging system
func (cl *CentralizedLogger) Shutdown(ctx context.Context) error {
	cl.logger.Info("Shutting down centralized logging system")

	// Cancel background processing
	cl.cancel()

	// Wait for background routines to finish
	done := make(chan struct{})
	go func() {
		cl.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		cl.logger.Info("Background routines stopped")
	case <-ctx.Done():
		cl.logger.Warn("Shutdown timeout, forcing stop")
	}

	// Flush all buffers
	cl.flushAllBuffers()

	// Stop all shippers
	for name, shipper := range cl.shippers {
		if err := shipper.Stop(); err != nil {
			cl.logger.Error("Failed to stop shipper", zap.String("shipper", name), zap.Error(err))
		}
	}

	// Close all outputs
	for name, output := range cl.outputs {
		if err := output.Close(); err != nil {
			cl.logger.Error("Failed to close output", zap.String("output", name), zap.Error(err))
		}
	}

	// Close indexer
	if cl.indexer != nil {
		if err := cl.indexer.Close(); err != nil {
			cl.logger.Error("Failed to close indexer", zap.Error(err))
		}
	}

	// Sync logger
	if err := cl.logger.Sync(); err != nil {
		// Ignore sync errors on stderr/stdout
		if err.Error() != "sync /dev/stderr: invalid argument" &&
			err.Error() != "sync /dev/stdout: inappropriate ioctl for device" {
			return err
		}
	}

	return nil
}

// GetStats returns current logging system statistics
func (cl *CentralizedLogger) GetStats() map[string]interface{} {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	stats := map[string]interface{}{
		"outputs":        len(cl.outputs),
		"processors":     len(cl.processors),
		"shippers":       len(cl.shippers),
		"config":         cl.config,
		"uptime":         time.Since(time.Now()), // This would be tracked properly
		"buffer_size":    cl.config.BufferSize,
		"flush_interval": cl.config.FlushInterval.String(),
		"async_mode":     cl.config.AsyncMode,
	}

	// Add indexer stats if available
	if cl.indexer != nil {
		stats["indexer"] = cl.indexer.GetStats()
	}

	return stats
}

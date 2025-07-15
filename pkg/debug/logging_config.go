package debug

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-manager-go/cmd/monitoring"
	"github.com/prometheus/client_golang/prometheus"
)

// IntegratedLoggingConfig provides unified configuration for both structured and centralized
// logging systems, enabling seamless integration between local logging and remote log shipping.
//
// This configuration structure supports:
//   - Structured logging with RFC 5424 compliance
//   - Centralized log collection and processing
//   - Remote log shipping to multiple destinations
//   - Bridge configuration for seamless integration
//   - Failover and backup mechanisms
//
// The configuration can be used to set up different deployment scenarios:
//   - Development: Local logging with console output
//   - Staging: File logging with basic shipping
//   - Production: Full centralized logging with multiple shippers and failover
//
// Example usage:
//
//	config := DefaultIntegratedLoggingConfig()
//	config.AddElasticsearchShipper("es", "http://localhost:9200", "logs")
//	setup, err := NewIntegratedLoggingSetup(config)
type IntegratedLoggingConfig struct {
	// Application settings
	AppName     string `json:"app_name" yaml:"app_name"`
	Environment string `json:"environment" yaml:"environment"`
	Version     string `json:"version" yaml:"version"`

	// Log levels and output
	Level     string `json:"level" yaml:"level"`         // "debug", "info", "warn", "error"
	Format    string `json:"format" yaml:"format"`       // "json", "console", "structured"
	Directory string `json:"directory" yaml:"directory"` // Log file directory

	// Structured logging specific
	StructuredConfig *StructuredLoggerConfig `json:"structured" yaml:"structured"`

	// Centralized logging specific
	CentralizedConfig *monitoring.CentralizedLoggingConfig `json:"centralized" yaml:"centralized"`

	// Integration settings
	EnableCentralizedForwarding bool                     `json:"enable_centralized_forwarding" yaml:"enable_centralized_forwarding"`
	BridgeConfig                *CentralizedBridgeConfig `json:"bridge" yaml:"bridge"`
	RemoteShipping              *RemoteShippingConfig    `json:"remote_shipping" yaml:"remote_shipping"`
}

// RemoteShippingConfig configures remote log shipping
type RemoteShippingConfig struct {
	Enabled     bool                                 `json:"enabled" yaml:"enabled"`
	Endpoints   map[string]*monitoring.ShipperConfig `json:"endpoints" yaml:"endpoints"`
	Failover    *FailoverConfig                      `json:"failover" yaml:"failover"`
	Compression bool                                 `json:"compression" yaml:"compression"`
	BatchSize   int                                  `json:"batch_size" yaml:"batch_size"`
	Timeout     time.Duration                        `json:"timeout" yaml:"timeout"`
}

// FailoverConfig configures failover behavior for remote shipping
type FailoverConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	BackupOutput  string        `json:"backup_output" yaml:"backup_output"` // "file", "console"
}

// LoggingSetup manages the integrated logging system
type LoggingSetup struct {
	config             *IntegratedLoggingConfig
	enhancedLogger     *EnhancedStructuredLogger
	centralizedLogger  *monitoring.CentralizedLogger
	prometheusRegistry *prometheus.Registry
}

// NewIntegratedLoggingSetup creates a complete integrated logging system
func NewIntegratedLoggingSetup(config *IntegratedLoggingConfig) (*LoggingSetup, error) {
	if config == nil {
		config = DefaultIntegratedLoggingConfig()
	}

	setup := &LoggingSetup{
		config:             config,
		prometheusRegistry: prometheus.NewRegistry(),
	}

	// Initialize centralized logging if enabled
	if config.EnableCentralizedForwarding && config.CentralizedConfig != nil {
		centralizedLogger, err := monitoring.NewCentralizedLogger(config.CentralizedConfig, setup.prometheusRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to create centralized logger: %w", err)
		}
		setup.centralizedLogger = centralizedLogger
	}

	// Initialize enhanced structured logger
	enhancedLogger, err := NewEnhancedStructuredLogger(
		config.StructuredConfig,
		setup.centralizedLogger,
		config.BridgeConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced structured logger: %w", err)
	}
	setup.enhancedLogger = enhancedLogger

	return setup, nil
}

// DefaultIntegratedLoggingConfig returns a default configuration
func DefaultIntegratedLoggingConfig() *IntegratedLoggingConfig {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".local", "share", "gzh-manager", "logs")

	return &IntegratedLoggingConfig{
		AppName:     "gzh-manager",
		Environment: "development",
		Version:     "1.0.0",
		Level:       "info",
		Format:      "json",
		Directory:   logDir,

		StructuredConfig: &StructuredLoggerConfig{
			Level:          SeverityInfo,
			Format:         "json",
			Output:         "stderr",
			AppName:        "gzh-manager",
			Version:        "1.0.0",
			Environment:    "development",
			EnableTracing:  true,
			EnableCaller:   true,
			CallerSkip:     3, // Adjusted for enhanced logger
			EnableSampling: false,
			SampleRate:     1.0,
			AsyncLogging:   false,
			BufferSize:     1000,
			FlushInterval:  time.Second,
			MaxFileSize:    100 * 1024 * 1024, // 100MB
			MaxBackups:     5,
			Compress:       true,
			ModuleLevels:   make(map[string]RFC5424Severity),
		},

		CentralizedConfig: &monitoring.CentralizedLoggingConfig{
			Level:         "info",
			Format:        "json",
			Directory:     logDir,
			BaseFilename:  "centralized.log",
			BufferSize:    1000,
			FlushInterval: 5 * time.Second,
			AsyncMode:     true,
			Labels: map[string]string{
				"service": "gzh-manager",
				"env":     "development",
			},
			Rotation: &monitoring.RotationConfig{
				MaxSizeMB:  10,
				MaxFiles:   5,
				MaxAgeDays: 30,
				Compress:   true,
				LocalTime:  true,
			},
			Outputs: map[string]*monitoring.OutputConfig{
				"console": {
					Type:    "console",
					Format:  "console",
					Level:   "info",
					Enabled: true,
					Settings: map[string]interface{}{
						"target": "stdout",
					},
				},
				"file": {
					Type:    "file",
					Format:  "json",
					Level:   "debug",
					Enabled: true,
					Settings: map[string]interface{}{
						"filename":     filepath.Join(logDir, "application.log"),
						"max_size_mb":  10,
						"max_files":    5,
						"max_age_days": 30,
						"compress":     true,
					},
				},
			},
			Processors: map[string]*monitoring.ProcessorConfig{
				"enrich": {
					Type:    "enrich",
					Enabled: true,
					Settings: map[string]interface{}{
						"static_fields": map[string]interface{}{
							"service": "gzh-manager",
							"version": "1.0.0",
						},
						"host_field":    "host",
						"process_field": "process",
					},
				},
			},
			Shippers: map[string]*monitoring.ShipperConfig{},
		},

		EnableCentralizedForwarding: true,

		BridgeConfig: &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        1000,
			AddStructuredData: true,
		},

		RemoteShipping: &RemoteShippingConfig{
			Enabled:     false,
			Endpoints:   make(map[string]*monitoring.ShipperConfig),
			Compression: true,
			BatchSize:   100,
			Timeout:     30 * time.Second,
			Failover: &FailoverConfig{
				Enabled:       true,
				RetryInterval: 5 * time.Minute,
				MaxRetries:    3,
				BackupOutput:  "file",
			},
		},
	}
}

// ProductionIntegratedLoggingConfig returns a production-ready configuration
func ProductionIntegratedLoggingConfig(appName, version string) *IntegratedLoggingConfig {
	config := DefaultIntegratedLoggingConfig()

	// Production-specific settings
	config.AppName = appName
	config.Version = version
	config.Environment = "production"
	config.Level = "info"
	config.Directory = "/var/log/gzh-manager"

	// Structured logger for production
	config.StructuredConfig.Level = SeverityInfo
	config.StructuredConfig.Environment = "production"
	config.StructuredConfig.AsyncLogging = true
	config.StructuredConfig.EnableSampling = true
	config.StructuredConfig.SampleRate = 0.1 // Sample 10% of debug/info logs
	config.StructuredConfig.Output = filepath.Join(config.Directory, "structured.log")

	// Centralized logger for production
	config.CentralizedConfig.Level = "info"
	config.CentralizedConfig.AsyncMode = true
	config.CentralizedConfig.Directory = config.Directory
	config.CentralizedConfig.BufferSize = 5000
	config.CentralizedConfig.FlushInterval = 10 * time.Second

	// Enable syslog output for production
	config.CentralizedConfig.Outputs["syslog"] = &monitoring.OutputConfig{
		Type:    "syslog",
		Format:  "structured",
		Level:   "warn",
		Enabled: true,
		Settings: map[string]interface{}{
			"network":  "tcp",
			"address":  "localhost:514",
			"tag":      appName,
			"priority": 16, // LOG_INFO | LOG_DAEMON
		},
	}

	// Enable remote shipping for production
	config.RemoteShipping.Enabled = true

	return config
}

// AddElasticsearchShipper adds Elasticsearch shipping to the configuration
func (config *IntegratedLoggingConfig) AddElasticsearchShipper(name, endpoint, index string) {
	if config.CentralizedConfig.Shippers == nil {
		config.CentralizedConfig.Shippers = make(map[string]*monitoring.ShipperConfig)
	}

	config.CentralizedConfig.Shippers[name] = &monitoring.ShipperConfig{
		Type:     "elasticsearch",
		Enabled:  true,
		Endpoint: endpoint,
		Settings: map[string]interface{}{
			"index":       index,
			"doc_type":    "_doc",
			"buffer_size": 100,
			"timeout":     "30s",
		},
	}

	config.RemoteShipping.Enabled = true
	if config.RemoteShipping.Endpoints == nil {
		config.RemoteShipping.Endpoints = make(map[string]*monitoring.ShipperConfig)
	}
	config.RemoteShipping.Endpoints[name] = config.CentralizedConfig.Shippers[name]
}

// AddLokiShipper adds Grafana Loki shipping to the configuration
func (config *IntegratedLoggingConfig) AddLokiShipper(name, endpoint string, labels map[string]string) {
	if config.CentralizedConfig.Shippers == nil {
		config.CentralizedConfig.Shippers = make(map[string]*monitoring.ShipperConfig)
	}

	settings := map[string]interface{}{
		"buffer_size": 100,
		"timeout":     "30s",
	}
	if labels != nil {
		settings["labels"] = labels
	}

	config.CentralizedConfig.Shippers[name] = &monitoring.ShipperConfig{
		Type:     "loki",
		Enabled:  true,
		Endpoint: endpoint,
		Settings: settings,
	}

	config.RemoteShipping.Enabled = true
	if config.RemoteShipping.Endpoints == nil {
		config.RemoteShipping.Endpoints = make(map[string]*monitoring.ShipperConfig)
	}
	config.RemoteShipping.Endpoints[name] = config.CentralizedConfig.Shippers[name]
}

// AddFluentdShipper adds Fluentd shipping to the configuration
func (config *IntegratedLoggingConfig) AddFluentdShipper(name, endpoint, tag string) {
	if config.CentralizedConfig.Shippers == nil {
		config.CentralizedConfig.Shippers = make(map[string]*monitoring.ShipperConfig)
	}

	config.CentralizedConfig.Shippers[name] = &monitoring.ShipperConfig{
		Type:     "fluentd",
		Enabled:  true,
		Endpoint: endpoint,
		Settings: map[string]interface{}{
			"tag":         tag,
			"buffer_size": 100,
			"timeout":     "30s",
		},
	}

	config.RemoteShipping.Enabled = true
	if config.RemoteShipping.Endpoints == nil {
		config.RemoteShipping.Endpoints = make(map[string]*monitoring.ShipperConfig)
	}
	config.RemoteShipping.Endpoints[name] = config.CentralizedConfig.Shippers[name]
}

// AddHTTPShipper adds generic HTTP shipping to the configuration
func (config *IntegratedLoggingConfig) AddHTTPShipper(name, endpoint string, headers map[string]string) {
	if config.CentralizedConfig.Shippers == nil {
		config.CentralizedConfig.Shippers = make(map[string]*monitoring.ShipperConfig)
	}

	settings := map[string]interface{}{
		"method":      "POST",
		"buffer_size": 100,
		"batch_size":  50,
		"timeout":     "30s",
	}
	if headers != nil {
		settings["headers"] = headers
	}

	config.CentralizedConfig.Shippers[name] = &monitoring.ShipperConfig{
		Type:     "http",
		Enabled:  true,
		Endpoint: endpoint,
		Settings: settings,
	}

	config.RemoteShipping.Enabled = true
	if config.RemoteShipping.Endpoints == nil {
		config.RemoteShipping.Endpoints = make(map[string]*monitoring.ShipperConfig)
	}
	config.RemoteShipping.Endpoints[name] = config.CentralizedConfig.Shippers[name]
}

// GetLogger returns the enhanced structured logger
func (setup *LoggingSetup) GetLogger() *EnhancedStructuredLogger {
	return setup.enhancedLogger
}

// GetCentralizedLogger returns the centralized logger
func (setup *LoggingSetup) GetCentralizedLogger() *monitoring.CentralizedLogger {
	return setup.centralizedLogger
}

// GetPrometheusRegistry returns the Prometheus registry for metrics
func (setup *LoggingSetup) GetPrometheusRegistry() *prometheus.Registry {
	return setup.prometheusRegistry
}

// GetConfig returns the integrated logging configuration
func (setup *LoggingSetup) GetConfig() *IntegratedLoggingConfig {
	return setup.config
}

// Shutdown gracefully shuts down the integrated logging system
func (setup *LoggingSetup) Shutdown() error {
	var lastErr error

	if setup.enhancedLogger != nil {
		if err := setup.enhancedLogger.Close(); err != nil {
			lastErr = err
		}
	}

	if setup.centralizedLogger != nil {
		if err := setup.centralizedLogger.Shutdown(context.Background()); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// UpdateLogLevel dynamically updates the log level for both loggers
func (setup *LoggingSetup) UpdateLogLevel(level string) error {
	// Update structured logger level
	severity, err := ParseRFC5424Severity(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	if setup.enhancedLogger != nil && setup.enhancedLogger.StructuredLogger != nil {
		setup.enhancedLogger.StructuredLogger.SetLevel(severity)
	}

	// Update centralized logger level (this would need to be implemented in monitoring package)
	if setup.centralizedLogger != nil {
		setup.config.CentralizedConfig.Level = level
	}

	return nil
}

// GetStats returns comprehensive logging system statistics
func (setup *LoggingSetup) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"config": setup.config,
	}

	if setup.enhancedLogger != nil {
		if bridge := setup.enhancedLogger.GetBridge(); bridge != nil {
			stats["bridge"] = bridge.GetStats()
		}
	}

	if setup.centralizedLogger != nil {
		stats["centralized"] = setup.centralizedLogger.GetStats()
	}

	return stats
}

// Global integrated logging instance
var globalIntegratedLogging *LoggingSetup

// InitGlobalIntegratedLogging initializes the global integrated logging system
func InitGlobalIntegratedLogging(config *IntegratedLoggingConfig) error {
	setup, err := NewIntegratedLoggingSetup(config)
	if err != nil {
		return fmt.Errorf("failed to create integrated logging setup: %w", err)
	}

	if globalIntegratedLogging != nil {
		globalIntegratedLogging.Shutdown()
	}

	globalIntegratedLogging = setup
	return nil
}

// GetGlobalIntegratedLogging returns the global integrated logging system
func GetGlobalIntegratedLogging() *LoggingSetup {
	return globalIntegratedLogging
}

// GetGlobalEnhancedLogger returns the global enhanced structured logger
func GetGlobalEnhancedLogger() *EnhancedStructuredLogger {
	if globalIntegratedLogging != nil {
		return globalIntegratedLogging.GetLogger()
	}
	return nil
}

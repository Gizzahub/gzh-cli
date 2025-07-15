package debug

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/cmd/monitoring"
)

// Example demonstrates comprehensive usage of the integrated logging system
func ExampleIntegratedLogging() {
	// 1. Create integrated logging configuration
	config := DefaultIntegratedLoggingConfig()
	config.AppName = "example-service"
	config.Environment = "development"
	config.Level = "debug"

	// 2. Add remote shipping endpoints
	config.AddElasticsearchShipper("elasticsearch", "http://localhost:9200", "example-logs")
	config.AddLokiShipper("loki", "http://localhost:3100", map[string]string{
		"environment": "development",
		"service":     "example-service",
	})

	// 3. Configure additional outputs
	config.CentralizedConfig.Outputs["file-debug"] = &monitoring.OutputConfig{
		Type:    "file",
		Format:  "json",
		Level:   "debug",
		Enabled: true,
		Settings: map[string]interface{}{
			"filename":     "/tmp/example-debug.log",
			"max_size_mb":  5,
			"max_files":    3,
			"max_age_days": 7,
			"compress":     true,
		},
	}

	// 4. Initialize the integrated logging system
	loggingSetup, err := NewIntegratedLoggingSetup(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logging: %w", err))
	}
	defer loggingSetup.Shutdown()

	// 5. Get the enhanced logger
	logger := loggingSetup.GetLogger()

	// 6. Use the logger with automatic forwarding to centralized system
	ctx := context.Background()

	logger.InfoLevel(ctx, "Application started", map[string]interface{}{
		"version":   "1.0.0",
		"component": "main",
		"startup_time": time.Now().Format(time.RFC3339),
	})

	// 7. Log different severity levels
	logger.DebugLevel(ctx, "Debugging configuration", map[string]interface{}{
		"config_loaded": true,
		"endpoints":     []string{"elasticsearch", "loki"},
	})

	logger.Warning(ctx, "Performance threshold exceeded", map[string]interface{}{
		"response_time_ms": 500,
		"threshold_ms":     300,
		"endpoint":         "/api/users",
	})

	logger.ErrorLevel(ctx, "Database connection failed", map[string]interface{}{
		"error":       "connection timeout",
		"database":    "users_db",
		"retry_count": 3,
	})

	// 8. Use module-specific logging
	dbLogger := logger.WithModule("database")
	dbLogger.InfoLevel(ctx, "Query executed", map[string]interface{}{
		"query":        "SELECT * FROM users WHERE id = ?",
		"execution_ms": 25,
		"rows_returned": 1,
	})

	// 9. Use pre-set fields logging
	httpLogger := logger.WithFields(map[string]interface{}{
		"component": "http_server",
		"version":   "2.1.0",
	})
	httpLogger.InfoLevel(ctx, "Request processed", map[string]interface{}{
		"method":     "GET",
		"path":       "/api/users/123",
		"status":     200,
		"latency_ms": 45,
	})

	// 10. Log with performance metrics
	httpLogger.InfoLevel(ctx, "Request with metrics", map[string]interface{}{
		"method":      "POST",
		"path":        "/api/users",
		"status":      201,
		"bytes_read":  1024,
		"bytes_written": 256,
	})

	// 11. Dynamic log level management
	if err := loggingSetup.UpdateLogLevel("warn"); err != nil {
		logger.ErrorLevel(ctx, "Failed to update log level", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 12. Get logging statistics
	stats := loggingSetup.GetStats()
	logger.InfoLevel(ctx, "Logging system statistics", map[string]interface{}{
		"stats": stats,
	})

	fmt.Printf("Example completed successfully. Logs sent to:\n")
	fmt.Printf("- Local files\n")
	fmt.Printf("- Console output\n")
	fmt.Printf("- Elasticsearch (if running)\n")
	fmt.Printf("- Grafana Loki (if running)\n")
}

// ExampleProductionLogging demonstrates production-ready logging setup
func ExampleProductionLogging() {
	// 1. Use production configuration
	config := ProductionIntegratedLoggingConfig("production-service", "2.1.0")

	// 2. Add production-grade remote shipping
	config.AddElasticsearchShipper("prod-elasticsearch", "https://elasticsearch.company.com:9200", "production-logs")
	config.AddLokiShipper("prod-loki", "https://loki.company.com:3100", map[string]string{
		"environment": "production",
		"service":     "production-service",
		"datacenter":  "us-east-1",
	})

	// 3. Add monitoring webhook
	config.AddHTTPShipper("monitoring-webhook", "https://monitoring.company.com/webhook", map[string]string{
		"Authorization": "Bearer YOUR_TOKEN_HERE",
		"Content-Type":  "application/json",
	})

	// 4. Initialize logging
	loggingSetup, err := NewIntegratedLoggingSetup(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize production logging: %w", err))
	}
	defer loggingSetup.Shutdown()

	logger := loggingSetup.GetLogger()
	ctx := context.Background()

	// 5. Production logging patterns
	logger.InfoLevel(ctx, "Service health check", map[string]interface{}{
		"status":         "healthy",
		"uptime_seconds": 3600,
		"memory_usage":   "256MB",
		"cpu_usage":      "15%",
	})

	// Error with structured context
	logger.ErrorLevel(ctx, "Payment processing failed", map[string]interface{}{
		"error_code":    "PAYMENT_DECLINED",
		"transaction_id": "txn_abc123",
		"user_id":       "user_456",
		"amount":        99.99,
		"currency":      "USD",
		"retry_count":   2,
	})

	// Business metrics
	logger.InfoLevel(ctx, "Order completed", map[string]interface{}{
		"order_id":      "order_789",
		"user_id":       "user_456",
		"total_amount":  149.99,
		"items_count":   3,
		"processing_ms": 1250,
		"payment_method": "credit_card",
	})

	fmt.Printf("Production logging example completed.\n")
}

// ExampleCustomConfiguration demonstrates advanced configuration
func ExampleCustomConfiguration() {
	// 1. Create custom configuration from scratch
	config := &IntegratedLoggingConfig{
		AppName:     "custom-service",
		Environment: "staging",
		Version:     "1.5.0",
		Level:       "info",
		Format:      "json",
		Directory:   "/var/log/custom-service",

		// Custom structured logger configuration
		StructuredConfig: &StructuredLoggerConfig{
			Level:           SeverityInfo,
			Format:          "json",
			Output:          "/var/log/custom-service/structured.log",
			AppName:         "custom-service",
			Version:         "1.5.0",
			Environment:     "staging",
			EnableTracing:   true,
			EnableCaller:    true,
			CallerSkip:      3,
			EnableSampling:  true,
			SampleRate:      0.2, // Sample 20% of logs
			AsyncLogging:    true,
			BufferSize:      2000,
			FlushInterval:   5 * time.Second,
			MaxFileSize:     50 * 1024 * 1024, // 50MB
			MaxBackups:      10,
			Compress:        true,
			ModuleLevels: map[string]RFC5424Severity{
				"database": SeverityDebug,
				"cache":    SeverityWarning,
				"http":     SeverityInfo,
			},
		},

		// Custom centralized logger configuration
		CentralizedConfig: &monitoring.CentralizedLoggingConfig{
			Level:         "info",
			Format:        "json",
			Directory:     "/var/log/custom-service",
			BaseFilename:  "centralized.log",
			BufferSize:    2000,
			FlushInterval: 10 * time.Second,
			AsyncMode:     true,
			Labels: map[string]string{
				"service":     "custom-service",
				"environment": "staging",
				"version":     "1.5.0",
			},
			Outputs: map[string]*monitoring.OutputConfig{
				"file-info": {
					Type:    "file",
					Format:  "json",
					Level:   "info",
					Enabled: true,
					Settings: map[string]interface{}{
						"filename":     "/var/log/custom-service/info.log",
						"max_size_mb":  20,
						"max_files":    5,
						"max_age_days": 14,
						"compress":     true,
					},
				},
				"file-error": {
					Type:    "file",
					Format:  "structured",
					Level:   "error",
					Enabled: true,
					Settings: map[string]interface{}{
						"filename":     "/var/log/custom-service/error.log",
						"max_size_mb":  10,
						"max_files":    10,
						"max_age_days": 30,
						"compress":     true,
					},
				},
				"syslog": {
					Type:    "syslog",
					Format:  "structured",
					Level:   "warn",
					Enabled: true,
					Settings: map[string]interface{}{
						"network":  "tcp",
						"address":  "log-server.company.com:514",
						"tag":      "custom-service",
						"priority": 16,
					},
				},
			},
			Processors: map[string]*monitoring.ProcessorConfig{
				"enrich": {
					Type:    "enrich",
					Enabled: true,
					Settings: map[string]interface{}{
						"static_fields": map[string]interface{}{
							"service":     "custom-service",
							"version":     "1.5.0",
							"environment": "staging",
							"datacenter":  "us-west-2",
						},
					},
				},
				"filter": {
					Type:    "filter",
					Enabled: true,
					Settings: map[string]interface{}{
						"allowed_levels": []string{"info", "warn", "error"},
						"message_patterns": []string{
							".*error.*",
							".*failed.*",
							".*exception.*",
						},
					},
				},
			},
			Shippers: map[string]*monitoring.ShipperConfig{
				"staging-elasticsearch": {
					Type:     "elasticsearch",
					Enabled:  true,
					Endpoint: "https://staging-elasticsearch.company.com:9200",
					Settings: map[string]interface{}{
						"index":       "custom-service-staging",
						"doc_type":    "_doc",
						"buffer_size": 200,
						"timeout":     "45s",
					},
				},
			},
		},

		EnableCentralizedForwarding: true,

		BridgeConfig: &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        2000,
			AddStructuredData: true,
		},

		RemoteShipping: &RemoteShippingConfig{
			Enabled:     true,
			Compression: true,
			BatchSize:   150,
			Timeout:     45 * time.Second,
			Failover: &FailoverConfig{
				Enabled:       true,
				RetryInterval: 2 * time.Minute,
				MaxRetries:    5,
				BackupOutput:  "file",
			},
		},
	}

	// 2. Initialize and use the custom configuration
	loggingSetup, err := NewIntegratedLoggingSetup(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize custom logging: %w", err))
	}
	defer loggingSetup.Shutdown()

	logger := loggingSetup.GetLogger()
	ctx := context.Background()

	// 3. Test the custom configuration
	logger.InfoLevel(ctx, "Custom logging system initialized", map[string]interface{}{
		"config_type":        "custom",
		"sampling_enabled":   true,
		"sampling_rate":      0.2,
		"async_enabled":      true,
		"centralized_enabled": true,
	})

	fmt.Printf("Custom configuration example completed.\n")
}

// ExampleMinimalSetup demonstrates the simplest possible setup
func ExampleMinimalSetup() {
	// 1. Use default configuration with minimal customization
	config := DefaultIntegratedLoggingConfig()
	config.AppName = "minimal-app"

	// 2. Initialize logging
	loggingSetup, err := NewIntegratedLoggingSetup(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize minimal logging: %w", err))
	}
	defer loggingSetup.Shutdown()

	// 3. Get logger and use it
	logger := loggingSetup.GetLogger()
	ctx := context.Background()

	logger.InfoLevel(ctx, "Hello from minimal logging setup!")

	fmt.Printf("Minimal setup example completed.\n")
}

// ExampleGlobalLogging demonstrates using the global logging instance
func ExampleGlobalLogging() {
	// 1. Initialize global logging
	config := DefaultIntegratedLoggingConfig()
	config.AppName = "global-app"

	if err := InitGlobalIntegratedLogging(config); err != nil {
		panic(fmt.Errorf("failed to initialize global logging: %w", err))
	}

	// 2. Use global logger from anywhere in the application
	logger := GetGlobalEnhancedLogger()
	if logger == nil {
		panic("global logger not initialized")
	}

	ctx := context.Background()
	logger.InfoLevel(ctx, "Using global logging instance", map[string]interface{}{
		"feature": "global_logging",
		"easy":    true,
	})

	// 3. Get global logging setup for advanced operations
	setup := GetGlobalIntegratedLogging()
	if setup != nil {
		stats := setup.GetStats()
		logger.InfoLevel(ctx, "Global logging stats", map[string]interface{}{
			"stats": stats,
		})
	}

	fmt.Printf("Global logging example completed.\n")
}
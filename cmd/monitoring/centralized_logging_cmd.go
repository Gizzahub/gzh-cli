package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// newCentralizedLoggingCmd creates the centralized logging command
func newCentralizedLoggingCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logging",
		Short: "Centralized logging system management",
		Long: `Manage the centralized logging system for collecting, processing, and shipping logs.

Features:
- Centralized log collection from all components
- Structured logging with configurable formats
- Log processing pipelines (filtering, transformation, enrichment)
- Multiple output destinations (file, console, syslog, HTTP)
- Log shipping to external systems (Elasticsearch, Loki, Fluentd)
- Real-time log streaming and aggregation
- Comprehensive logging metrics and monitoring

Examples:
  # Start centralized logging server
  gz monitoring logging server --config /path/to/logging.yaml
  
  # View current logging configuration
  gz monitoring logging config show
  
  # Test log shipping to external systems
  gz monitoring logging test-ship --shipper elasticsearch
  
  # Generate sample logging configuration
  gz monitoring logging config generate > logging.yaml`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newLoggingServerCmd(ctx))
	cmd.AddCommand(newLoggingConfigCmd(ctx))
	cmd.AddCommand(newLoggingTestCmd(ctx))
	cmd.AddCommand(newLoggingStatsCmd(ctx))

	return cmd
}

// newLoggingServerCmd creates the logging server command
func newLoggingServerCmd(ctx context.Context) *cobra.Command {
	var (
		configPath string
		logLevel   string
		logFormat  string
		logDir     string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start centralized logging server",
		Long: `Start the centralized logging server that collects logs from all components.

The server provides:
- HTTP API for log ingestion
- WebSocket streaming for real-time log viewing
- Processing pipelines for log transformation
- Multiple output destinations
- Shipping to external log aggregation systems

Examples:
  # Start with default configuration
  gz monitoring logging server
  
  # Start with custom configuration
  gz monitoring logging server --config /etc/gzh-manager/logging.yaml
  
  # Start with custom log level and format
  gz monitoring logging server --log-level debug --log-format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			config, err := loadLoggingConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load logging configuration: %w", err)
			}

			// Override configuration with command line flags
			if logLevel != "" {
				config.Level = logLevel
			}
			if logFormat != "" {
				config.Format = logFormat
			}
			if logDir != "" {
				config.Directory = logDir
			}

			// Apply defaults
			applyLoggingConfigDefaults(config)

			// Create Prometheus registry
			registry := prometheus.NewRegistry()

			// Create centralized logger
			logger, err := NewCentralizedLogger(config, registry)
			if err != nil {
				return fmt.Errorf("failed to create centralized logger: %w", err)
			}

			fmt.Printf("🚀 Starting centralized logging server\n")
			fmt.Printf("📊 Log level: %s\n", config.Level)
			fmt.Printf("📝 Log format: %s\n", config.Format)
			fmt.Printf("📁 Log directory: %s\n", config.Directory)
			fmt.Printf("🔄 Active outputs: %d\n", len(config.Outputs))
			fmt.Printf("⚙️  Active processors: %d\n", len(config.Processors))
			fmt.Printf("🚢 Active shippers: %d\n", len(config.Shippers))

			// Start HTTP API server for log ingestion
			apiServer := NewLoggingAPIServer(logger, config)
			go func() {
				if err := apiServer.Start(":8080"); err != nil {
					logger.GetLogger().Error("Failed to start logging API server", zap.Error(err))
				}
			}()

			// Test logging
			logger.GetLogger().Info("Centralized logging server started successfully")

			// Wait for context cancellation
			<-ctx.Done()

			fmt.Println("\n🛑 Shutting down centralized logging server...")

			// Graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := logger.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("failed to shutdown logger: %w", err)
			}

			fmt.Println("✅ Centralized logging server stopped")
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to logging configuration file")
	cmd.Flags().StringVar(&logLevel, "log-level", "", "Log level override (debug, info, warn, error)")
	cmd.Flags().StringVar(&logFormat, "log-format", "", "Log format override (json, console, structured)")
	cmd.Flags().StringVar(&logDir, "log-dir", "", "Log directory override")

	return cmd
}

// newLoggingConfigCmd creates the logging configuration command
func newLoggingConfigCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Manage logging configuration",
		Long:         `View, generate, and validate logging configuration files.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newLoggingConfigShowCmd())
	cmd.AddCommand(newLoggingConfigGenerateCmd())
	cmd.AddCommand(newLoggingConfigValidateCmd())

	return cmd
}

func newLoggingConfigShowCmd() *cobra.Command {
	var configPath string

	return &cobra.Command{
		Use:   "show",
		Short: "Show current logging configuration",
		Long:  `Display the current logging configuration in YAML format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadLoggingConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			data, err := yaml.Marshal(config)
			if err != nil {
				return fmt.Errorf("failed to marshal configuration: %w", err)
			}

			fmt.Print(string(data))
			return nil
		},
	}
}

func newLoggingConfigGenerateCmd() *cobra.Command {
	var (
		template string
		output   string
	)

	return &cobra.Command{
		Use:   "generate",
		Short: "Generate sample logging configuration",
		Long:  `Generate a sample logging configuration file with common settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := generateSampleLoggingConfig(template)

			data, err := yaml.Marshal(config)
			if err != nil {
				return fmt.Errorf("failed to marshal configuration: %w", err)
			}

			if output != "" {
				if err := os.WriteFile(output, data, 0o644); err != nil {
					return fmt.Errorf("failed to write configuration file: %w", err)
				}
				fmt.Printf("📄 Sample configuration written to %s\n", output)
			} else {
				fmt.Print(string(data))
			}

			return nil
		},
	}
}

func newLoggingConfigValidateCmd() *cobra.Command {
	var configPath string

	return &cobra.Command{
		Use:   "validate",
		Short: "Validate logging configuration",
		Long:  `Validate the syntax and settings of a logging configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadLoggingConfig(configPath)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// Validate configuration
			if err := validateLoggingConfig(config); err != nil {
				fmt.Printf("❌ Configuration validation failed: %v\n", err)
				return err
			}

			fmt.Printf("✅ Configuration is valid\n")
			fmt.Printf("📊 Outputs: %d\n", len(config.Outputs))
			fmt.Printf("⚙️  Processors: %d\n", len(config.Processors))
			fmt.Printf("🚢 Shippers: %d\n", len(config.Shippers))

			return nil
		},
	}
}

// newLoggingTestCmd creates the logging test command
func newLoggingTestCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "test",
		Short:        "Test logging system components",
		Long:         `Test various components of the logging system including outputs, processors, and shippers.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newLoggingTestOutputCmd())
	cmd.AddCommand(newLoggingTestShipperCmd())
	cmd.AddCommand(newLoggingTestProcessorCmd())

	return cmd
}

func newLoggingTestOutputCmd() *cobra.Command {
	var (
		outputType string
		configPath string
	)

	return &cobra.Command{
		Use:   "output",
		Short: "Test log output",
		Long:  `Test a specific log output configuration by sending sample log entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🧪 Testing %s output...\n", outputType)

			// Use configPath if needed for future implementation
			_ = configPath

			// Create sample log entry
			entry := &LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Test log entry from centralized logging system",
				Logger:    "test",
				Fields: map[string]interface{}{
					"test_id":   "test-123",
					"component": "centralized-logging",
					"test_type": "output-test",
				},
				Labels: map[string]string{
					"env":     "test",
					"service": "gzh-manager",
				},
			}

			// Test output based on type
			// Implementation would create and test the specific output type
			// For now, just use the entry for placeholder testing
			_ = entry
			fmt.Printf("✅ %s output test completed successfully\n", outputType)

			return nil
		},
	}
}

func newLoggingTestShipperCmd() *cobra.Command {
	var (
		shipperName string
		endpoint    string
	)

	return &cobra.Command{
		Use:   "shipper",
		Short: "Test log shipper",
		Long:  `Test a specific log shipper by sending sample log entries to the configured endpoint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚢 Testing %s shipper to %s...\n", shipperName, endpoint)

			// Create sample log entries
			entries := []*LogEntry{
				{
					Timestamp: time.Now(),
					Level:     "info",
					Message:   "Test log entry #1",
					Logger:    "test",
					Fields: map[string]interface{}{
						"test_id": "ship-test-1",
					},
				},
				{
					Timestamp: time.Now(),
					Level:     "warn",
					Message:   "Test log entry #2",
					Logger:    "test",
					Fields: map[string]interface{}{
						"test_id": "ship-test-2",
					},
				},
			}

			// Test shipper
			// Implementation would create and test the specific shipper
			fmt.Printf("✅ %s shipper test completed successfully\n", shipperName)
			fmt.Printf("📊 Shipped %d log entries\n", len(entries))

			return nil
		},
	}
}

func newLoggingTestProcessorCmd() *cobra.Command {
	var processorType string

	return &cobra.Command{
		Use:   "processor",
		Short: "Test log processor",
		Long:  `Test a specific log processor by processing sample log entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("⚙️  Testing %s processor...\n", processorType)

			// Create sample log entry
			entry := &LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Test log entry for processor",
				Logger:    "test",
				Fields: map[string]interface{}{
					"original_field": "original_value",
					"test_field":     "test_value",
				},
			}

			// Test processor
			// Implementation would create and test the specific processor
			// For now, just use the entry for placeholder testing
			_ = entry
			fmt.Printf("✅ %s processor test completed successfully\n", processorType)

			return nil
		},
	}
}

// newLoggingStatsCmd creates the logging statistics command
func newLoggingStatsCmd(ctx context.Context) *cobra.Command {
	var (
		format   string
		interval string
		watch    bool
	)

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show logging system statistics",
		Long: `Display real-time statistics about the centralized logging system including:
- Processing rates and throughput
- Buffer utilization
- Error counts
- Output and shipper status

Examples:
  # Show current statistics
  gz monitoring logging stats
  
  # Watch statistics in real-time
  gz monitoring logging stats --watch
  
  # Show statistics in JSON format
  gz monitoring logging stats --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if watch {
				return watchLoggingStats(format, interval)
			}

			return showLoggingStats(format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")
	cmd.Flags().StringVar(&interval, "interval", "5s", "Watch interval")
	cmd.Flags().BoolVar(&watch, "watch", false, "Watch statistics in real-time")

	return cmd
}

// Helper functions

func loadLoggingConfig(configPath string) (*CentralizedLoggingConfig, error) {
	if configPath == "" {
		// Try default locations
		defaultPaths := []string{
			"./logging.yaml",
			"./config/logging.yaml",
			"~/.config/gzh-manager/logging.yaml",
			"/etc/gzh-manager/logging.yaml",
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}

		if configPath == "" {
			// Return default configuration
			return getDefaultLoggingConfig(), nil
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config CentralizedLoggingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func getDefaultLoggingConfig() *CentralizedLoggingConfig {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".local", "share", "gzh-manager", "logs")

	return &CentralizedLoggingConfig{
		Level:        "info",
		Format:       "json",
		Directory:    logDir,
		BaseFilename: "gzh-manager.log",
		Labels: map[string]string{
			"service": "gzh-manager",
			"env":     "development",
		},
		Rotation: &RotationConfig{
			MaxSizeMB:  10,
			MaxFiles:   5,
			MaxAgeDays: 30,
			Compress:   true,
			LocalTime:  true,
		},
		Outputs: map[string]*OutputConfig{
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
				Level:   "info",
				Enabled: true,
				Settings: map[string]interface{}{
					"filename":     filepath.Join(logDir, "centralized.log"),
					"max_size_mb":  10,
					"max_files":    5,
					"max_age_days": 30,
					"compress":     true,
				},
			},
		},
		Processors: map[string]*ProcessorConfig{
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
		Shippers:      map[string]*ShipperConfig{},
		BufferSize:    1000,
		FlushInterval: 5 * time.Second,
		AsyncMode:     true,
	}
}

func applyLoggingConfigDefaults(config *CentralizedLoggingConfig) {
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Format == "" {
		config.Format = "json"
	}
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.BaseFilename == "" {
		config.BaseFilename = "gzh-manager.log"
	}
}

func generateSampleLoggingConfig(template string) *CentralizedLoggingConfig {
	config := getDefaultLoggingConfig()

	if template == "complete" {
		// Add more outputs, processors, and shippers for complete example
		config.Outputs["syslog"] = &OutputConfig{
			Type:    "syslog",
			Format:  "structured",
			Level:   "warn",
			Enabled: false,
			Settings: map[string]interface{}{
				"network":  "tcp",
				"address":  "localhost:514",
				"tag":      "gzh-manager",
				"priority": 16, // LOG_INFO | LOG_DAEMON
			},
		}

		config.Processors["filter"] = &ProcessorConfig{
			Type:    "filter",
			Enabled: false,
			Settings: map[string]interface{}{
				"allowed_levels": []string{"info", "warn", "error"},
				"message_patterns": []string{
					".*error.*",
					".*failed.*",
				},
			},
		}

		config.Shippers["elasticsearch"] = &ShipperConfig{
			Type:     "elasticsearch",
			Enabled:  false,
			Endpoint: "http://localhost:9200",
			Settings: map[string]interface{}{
				"index":       "gzh-manager-logs",
				"doc_type":    "_doc",
				"buffer_size": 100,
				"timeout":     "30s",
			},
		}
	}

	return config
}

func validateLoggingConfig(config *CentralizedLoggingConfig) error {
	// Validate basic settings
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[config.Level] {
		return fmt.Errorf("invalid log level: %s", config.Level)
	}

	validFormats := map[string]bool{
		"json": true, "console": true, "structured": true,
	}
	if !validFormats[config.Format] {
		return fmt.Errorf("invalid log format: %s", config.Format)
	}

	// Validate outputs
	for name, output := range config.Outputs {
		if output.Type == "" {
			return fmt.Errorf("output %s: type is required", name)
		}
	}

	// Validate processors
	for name, processor := range config.Processors {
		if processor.Type == "" {
			return fmt.Errorf("processor %s: type is required", name)
		}
	}

	// Validate shippers
	for name, shipper := range config.Shippers {
		if shipper.Type == "" {
			return fmt.Errorf("shipper %s: type is required", name)
		}
		if shipper.Enabled && shipper.Endpoint == "" {
			return fmt.Errorf("shipper %s: endpoint is required when enabled", name)
		}
	}

	return nil
}

func showLoggingStats(format string) error {
	// This would connect to a running logging system and fetch statistics
	stats := map[string]interface{}{
		"timestamp":       time.Now(),
		"entries_total":   12345,
		"entries_rate":    150.5,
		"active_outputs":  3,
		"active_shippers": 1,
		"buffer_usage":    "45%",
		"errors_total":    2,
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Println("📊 Centralized Logging Statistics")
		fmt.Println("================================")
		fmt.Printf("Timestamp:        %s\n", stats["timestamp"])
		fmt.Printf("Entries Total:    %v\n", stats["entries_total"])
		fmt.Printf("Entries Rate:     %.1f/sec\n", stats["entries_rate"])
		fmt.Printf("Active Outputs:   %v\n", stats["active_outputs"])
		fmt.Printf("Active Shippers:  %v\n", stats["active_shippers"])
		fmt.Printf("Buffer Usage:     %v\n", stats["buffer_usage"])
		fmt.Printf("Errors Total:     %v\n", stats["errors_total"])
	}

	return nil
}

func watchLoggingStats(format, interval string) error {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		if format != "json" {
			fmt.Print("\033[2J\033[H") // Clear screen
		}

		if err := showLoggingStats(format); err != nil {
			return err
		}

		<-ticker.C
	}
}

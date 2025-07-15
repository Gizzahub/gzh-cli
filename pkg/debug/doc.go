// Package debug provides comprehensive debugging, logging, profiling, and tracing capabilities
// for the gzh-manager CLI tool. It implements RFC 5424 compliant structured logging with
// centralized log management, dynamic log level control, and performance profiling tools.
//
// # Key Features
//
// ## Structured Logging
//
// The package implements RFC 5424 compliant structured logging with support for:
//   - JSON, console, and logfmt output formats
//   - Distributed tracing integration (OpenTelemetry)
//   - Async logging with configurable buffering
//   - Sampling and filtering capabilities
//   - Module-specific log level control
//
// ## Centralized Logging Integration
//
// Provides seamless integration between structured logging and centralized log management:
//   - Bridge between StructuredLogger and CentralizedLogger
//   - Support for multiple log shipping destinations (Elasticsearch, Loki, Fluentd, HTTP)
//   - Real-time log streaming via WebSocket
//   - Prometheus metrics integration
//   - Automatic failover and backup mechanisms
//
// ## Dynamic Log Level Management
//
// Advanced log level control system featuring:
//   - Rule-based conditional logging
//   - HTTP API for runtime configuration
//   - Signal-based control (SIGUSR1, SIGUSR2, SIGHUP)
//   - Profile-based configuration (development, testing, production)
//   - Performance-aware adaptive sampling
//
// ## Performance Profiling
//
// Comprehensive profiling capabilities including:
//   - CPU, memory, goroutine, and mutex profiling
//   - Custom performance tracing
//   - Runtime statistics collection
//   - HTTP endpoints for pprof integration
//   - Configurable profiling duration and intervals
//
// # Usage Examples
//
// ## Basic Structured Logging
//
//	config := DefaultStructuredLoggerConfig()
//	logger, err := NewStructuredLogger(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer logger.Close()
//
//	ctx := context.Background()
//	logger.InfoLevel(ctx, "Application started", map[string]interface{}{
//		"version": "1.0.0",
//		"env":     "production",
//	})
//
// ## Integrated Logging with Remote Shipping
//
//	config := DefaultIntegratedLoggingConfig()
//	config.AddElasticsearchShipper("es", "http://localhost:9200", "app-logs")
//	config.AddLokiShipper("loki", "http://localhost:3100", map[string]string{
//		"service": "gzh-manager",
//		"env":     "production",
//	})
//
//	setup, err := NewIntegratedLoggingSetup(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer setup.Shutdown()
//
//	logger := setup.GetLogger()
//	logger.InfoLevel(ctx, "Processing request", map[string]interface{}{
//		"request_id": "req-123",
//		"user_id":    "user-456",
//	})
//
// ## Dynamic Log Level Control
//
//	manager, err := NewLogLevelManager(DefaultLogLevelManagerConfig(), logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Apply production profile
//	manager.ApplyProfile("production")
//
//	// Start HTTP API for runtime control
//	manager.StartHTTPServer(8080)
//
//	// Add custom rule
//	rule := LogLevelRule{
//		ID:      "debug-module",
//		Enabled: true,
//		Conditions: []LogCondition{{
//			Field:    "module",
//			Operator: "eq",
//			Value:    "database",
//		}},
//		Actions: []LogAction{{
//			Type:  "set_level",
//			Value: SeverityDebug,
//		}},
//	}
//	manager.AddRule(rule)
//
// ## Performance Profiling
//
//	config := DefaultProfilerConfig()
//	config.CPUProfile = true
//	config.MemoryProfile = true
//	config.Duration = 30 * time.Second
//
//	profiler, err := NewProfiler(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	profiler.Start()
//	// ... application logic ...
//	profiler.Stop()
//
// # Configuration
//
// The package supports extensive configuration through various config structures:
//   - StructuredLoggerConfig: Core logging configuration
//   - IntegratedLoggingConfig: Unified logging and shipping configuration
//   - LogLevelManagerConfig: Dynamic log level management
//   - ProfilerConfig: Performance profiling settings
//   - TracerConfig: Custom tracing configuration
//
// # Architecture Integration
//
// The debug package integrates seamlessly with the gzh-manager architecture:
//   - CLI commands can use the global logging instance
//   - Service packages can create module-specific loggers
//   - Configuration is managed through the unified config system
//   - Metrics are exposed via Prometheus integration
//   - HTTP APIs provide runtime observability and control
//
// # Thread Safety
//
// All components in this package are designed to be thread-safe:
//   - Concurrent logging from multiple goroutines
//   - Safe configuration updates during runtime
//   - Protected access to shared state with appropriate locking
//   - Atomic operations for performance-critical paths
//
// # Error Handling
//
// The package follows Go best practices for error handling:
//   - Wrapped errors with context information
//   - Fallback mechanisms for external dependencies
//   - Graceful degradation when remote services are unavailable
//   - Comprehensive error logging and debugging information
//
// For more detailed information, see the individual type and function documentation.
package debug

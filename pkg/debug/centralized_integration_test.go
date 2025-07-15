package debug

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/cmd/monitoring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCentralizedLoggerBridge(t *testing.T) {
	// Create test centralized logging config
	centralizedConfig := &monitoring.CentralizedLoggingConfig{
		Level:         "debug",
		Format:        "json",
		Directory:     "/tmp/gzh-manager-test",
		BaseFilename:  "test.log",
		BufferSize:    100,
		FlushInterval: 1 * time.Second,
		AsyncMode:     false,
		Labels: map[string]string{
			"service": "gzh-manager-test",
			"env":     "test",
		},
		Outputs: map[string]*monitoring.OutputConfig{
			"console": {
				Type:    "console",
				Format:  "json",
				Level:   "debug",
				Enabled: true,
				Settings: map[string]interface{}{
					"target": "stdout",
				},
			},
		},
		Processors: map[string]*monitoring.ProcessorConfig{
			"enrich": {
				Type:    "enrich",
				Enabled: true,
				Settings: map[string]interface{}{
					"static_fields": map[string]interface{}{
						"test_run": true,
					},
				},
			},
		},
		Shippers: map[string]*monitoring.ShipperConfig{},
	}

	// Create Prometheus registry for metrics
	registry := prometheus.NewRegistry()

	// Create centralized logger
	centralizedLogger, err := monitoring.NewCentralizedLogger(centralizedConfig, registry)
	require.NoError(t, err)
	defer centralizedLogger.Shutdown(context.Background())

	// Create structured logger config
	structuredConfig := DefaultStructuredLoggerConfig()
	structuredConfig.Level = SeverityDebug
	structuredConfig.AppName = "test-app"
	structuredConfig.Environment = "test"

	// Create structured logger
	structuredLogger, err := NewStructuredLogger(structuredConfig)
	require.NoError(t, err)
	defer structuredLogger.Close()

	t.Run("Bridge Creation", func(t *testing.T) {
		bridgeConfig := &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        100,
			AddStructuredData: true,
		}

		bridge, err := NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, bridgeConfig)
		require.NoError(t, err)
		defer bridge.Shutdown()

		assert.True(t, bridge.IsForwardingEnabled())

		stats := bridge.GetStats()
		assert.Equal(t, true, stats["enabled"])
		assert.Equal(t, 100, stats["buffer_size"])
	})

	t.Run("Log Entry Forwarding", func(t *testing.T) {
		bridgeConfig := &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        10,
			AddStructuredData: true,
		}

		bridge, err := NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, bridgeConfig)
		require.NoError(t, err)
		defer bridge.Shutdown()

		// Create a test log entry
		entry := &StructuredLogEntry{
			Timestamp: time.Now().UTC(),
			Version:   1,
			Level:     "info",
			Severity:  SeverityInfo,
			Hostname:  "test-host",
			AppName:   "test-app",
			ProcID:    "12345",
			Message:   "Test message for forwarding",
			Fields: map[string]interface{}{
				"test_field": "test_value",
				"number":     42,
			},
			TraceID: "trace-123",
			SpanID:  "span-456",
		}
		entry.Caller.File = "test.go"
		entry.Caller.Line = 100
		entry.Caller.Function = "TestFunction"

		// Forward the entry
		err = bridge.ForwardLogEntry(entry)
		assert.NoError(t, err)

		// Give some time for async processing
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Enhanced Logger Integration", func(t *testing.T) {
		bridgeConfig := &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        50,
			AddStructuredData: true,
		}

		enhancedLogger, err := NewEnhancedStructuredLogger(structuredConfig, centralizedLogger, bridgeConfig)
		require.NoError(t, err)
		defer enhancedLogger.Close()

		ctx := context.Background()

		// Test all log levels
		enhancedLogger.Emergency(ctx, "Emergency test message", map[string]interface{}{"type": "emergency"})
		enhancedLogger.Alert(ctx, "Alert test message", map[string]interface{}{"type": "alert"})
		enhancedLogger.Critical(ctx, "Critical test message", map[string]interface{}{"type": "critical"})
		enhancedLogger.ErrorLevel(ctx, "Error test message", map[string]interface{}{"type": "error"})
		enhancedLogger.Warning(ctx, "Warning test message", map[string]interface{}{"type": "warning"})
		enhancedLogger.Notice(ctx, "Notice test message", map[string]interface{}{"type": "notice"})
		enhancedLogger.InfoLevel(ctx, "Info test message", map[string]interface{}{"type": "info"})
		enhancedLogger.DebugLevel(ctx, "Debug test message", map[string]interface{}{"type": "debug"})

		// Verify bridge is accessible
		bridge := enhancedLogger.GetBridge()
		assert.NotNil(t, bridge)
		assert.True(t, bridge.IsForwardingEnabled())

		// Verify centralized logger is accessible
		centralizedFromBridge := enhancedLogger.GetCentralizedLogger()
		assert.NotNil(t, centralizedFromBridge)

		// Allow time for async processing
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("Bridge Disable/Enable", func(t *testing.T) {
		bridgeConfig := &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        10,
			AddStructuredData: true,
		}

		bridge, err := NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, bridgeConfig)
		require.NoError(t, err)
		defer bridge.Shutdown()

		// Test enabling/disabling
		assert.True(t, bridge.IsForwardingEnabled())

		bridge.DisableForwarding()
		assert.False(t, bridge.IsForwardingEnabled())

		bridge.EnableForwarding()
		assert.True(t, bridge.IsForwardingEnabled())
	})

	t.Run("Buffer Overflow Handling", func(t *testing.T) {
		// Create bridge with small buffer
		bridgeConfig := &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        2,
			AddStructuredData: true,
		}

		bridge, err := NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, bridgeConfig)
		require.NoError(t, err)
		defer bridge.Shutdown()

		// Send more entries than buffer can hold
		for i := 0; i < 5; i++ {
			entry := &StructuredLogEntry{
				Timestamp: time.Now().UTC(),
				Version:   1,
				Level:     "info",
				Severity:  SeverityInfo,
				Hostname:  "test-host",
				AppName:   "test-app",
				Message:   fmt.Sprintf("Test message %d", i),
			}

			err := bridge.ForwardLogEntry(entry)
			assert.NoError(t, err)
		}

		// Check buffer stats
		stats := bridge.GetStats()
		bufferUtilization := stats["buffer_utilization"].(float64)
		assert.GreaterOrEqual(t, bufferUtilization, 0.0)
		assert.LessOrEqual(t, bufferUtilization, 100.0)
	})
}

func TestCentralizedIntegrationWithNilLogger(t *testing.T) {
	// Test bridge creation with nil centralized logger
	structuredConfig := DefaultStructuredLoggerConfig()
	structuredLogger, err := NewStructuredLogger(structuredConfig)
	require.NoError(t, err)
	defer structuredLogger.Close()

	enhancedLogger, err := NewEnhancedStructuredLogger(structuredConfig, nil, nil)
	require.NoError(t, err)
	defer enhancedLogger.Close()

	// Should work without centralized logger
	ctx := context.Background()
	enhancedLogger.InfoLevel(ctx, "Test message without centralized logger")

	// Bridge should be nil
	bridge := enhancedLogger.GetBridge()
	assert.Nil(t, bridge)

	centralizedLogger := enhancedLogger.GetCentralizedLogger()
	assert.Nil(t, centralizedLogger)
}

func TestBridgeConfigDefaults(t *testing.T) {
	structuredConfig := DefaultStructuredLoggerConfig()
	structuredLogger, err := NewStructuredLogger(structuredConfig)
	require.NoError(t, err)
	defer structuredLogger.Close()

	centralizedConfig := &monitoring.CentralizedLoggingConfig{
		Level:         "info",
		Format:        "json",
		BufferSize:    100,
		FlushInterval: 1 * time.Second,
		AsyncMode:     false,
		Outputs: map[string]*monitoring.OutputConfig{
			"console": {
				Type:    "console",
				Format:  "json",
				Level:   "info",
				Enabled: true,
				Settings: map[string]interface{}{
					"target": "stdout",
				},
			},
		},
	}

	registry := prometheus.NewRegistry()
	centralizedLogger, err := monitoring.NewCentralizedLogger(centralizedConfig, registry)
	require.NoError(t, err)
	defer centralizedLogger.Shutdown(context.Background())

	// Test with nil config (should use defaults)
	bridge, err := NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, nil)
	require.NoError(t, err)
	defer bridge.Shutdown()

	stats := bridge.GetStats()
	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, 1000, stats["buffer_size"]) // Default buffer size
}

// Benchmark tests for performance validation
func BenchmarkCentralizedBridgeForwarding(b *testing.B) {
	// Setup
	centralizedConfig := &monitoring.CentralizedLoggingConfig{
		Level:         "info",
		Format:        "json",
		BufferSize:    1000,
		FlushInterval: 1 * time.Second,
		AsyncMode:     true,
		Outputs: map[string]*monitoring.OutputConfig{
			"console": {
				Type:    "console",
				Format:  "json",
				Level:   "info",
				Enabled: true,
				Settings: map[string]interface{}{
					"target": "stdout",
				},
			},
		},
	}

	registry := prometheus.NewRegistry()
	centralizedLogger, err := monitoring.NewCentralizedLogger(centralizedConfig, registry)
	require.NoError(b, err)
	defer centralizedLogger.Shutdown(context.Background())

	structuredConfig := DefaultStructuredLoggerConfig()
	structuredConfig.Level = SeverityInfo
	structuredConfig.AsyncLogging = true

	enhancedLogger, err := NewEnhancedStructuredLogger(structuredConfig, centralizedLogger, &CentralizedBridgeConfig{
		Enabled:           true,
		BufferSize:        1000,
		AddStructuredData: true,
	})
	require.NoError(b, err)
	defer enhancedLogger.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enhancedLogger.InfoLevel(ctx, "Benchmark test message", map[string]interface{}{
				"benchmark": true,
				"timestamp": time.Now().Unix(),
			})
		}
	})
}

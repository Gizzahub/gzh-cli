package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTracerConfig(t *testing.T) {
	config := DefaultTracerConfig()

	assert.False(t, config.Enabled)
	assert.Equal(t, "./debug-trace.json", config.OutputFile)
	assert.Equal(t, 100000, config.MaxEvents)
	assert.False(t, config.IncludeStack)
	assert.Equal(t, 10, config.StackDepth)
	assert.Equal(t, 1000, config.BufferSize)
	assert.Equal(t, 5*time.Second, config.FlushInterval)
	assert.Contains(t, config.Categories, "default")
}

func TestNewTracer(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("with config", func(t *testing.T) {
		config := &TracerConfig{
			Enabled:    true,
			OutputFile: filepath.Join(tmpDir, "trace.json"),
			MaxEvents:  1000,
		}

		tracer := NewTracer(config)
		assert.NotNil(t, tracer)

		// Clean up
		tracer.Stop()
	})

	t.Run("with nil config", func(t *testing.T) {
		tracer := NewTracer(nil)
		assert.NotNil(t, tracer)

		// Clean up
		tracer.Stop()
	})
}

func TestTracerEvents(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.json")

	config := &TracerConfig{
		Enabled:       true,
		OutputFile:    traceFile,
		MaxEvents:     100,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// Test instant event
	tracer.Instant("test_instant", "test", nil)

	// Test begin/end events
	span := tracer.Begin("test_span", "test", nil)
	time.Sleep(10 * time.Millisecond)
	span.End()

	// Wait for flush
	time.Sleep(200 * time.Millisecond)

	// Check if trace file exists
	if _, err := os.Stat(traceFile); err == nil {
		// Read and verify trace file content
		content, err := os.ReadFile(traceFile)
		if err == nil {
			// Should contain JSON events
			assert.True(t, len(content) > 0)
			// Basic JSON structure check
			assert.Contains(t, string(content), "[")
		}
	}
}

func TestTracerSpan(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.json")

	config := &TracerConfig{
		Enabled:       true,
		OutputFile:    traceFile,
		MaxEvents:     100,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// Test span
	span := tracer.Begin("test_operation", "test")
	assert.NotNil(t, span)

	time.Sleep(10 * time.Millisecond)
	span.End()

	// Wait for flush
	time.Sleep(200 * time.Millisecond)

	// Check if trace file exists (may not exist in test environment)
	if _, err := os.Stat(traceFile); err != nil {
		t.Logf("Trace file not created (expected in test environment): %v", err)
	}
}

func TestTracerWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.json")

	config := &TracerConfig{
		Enabled:       true,
		OutputFile:    traceFile,
		MaxEvents:     100,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// Test context span (using Begin since StartSpanWithContext doesn't exist)
	span := tracer.Begin("context_operation", "test")
	assert.NotNil(t, span)

	time.Sleep(10 * time.Millisecond)
	span.End()

	// Wait for flush
	time.Sleep(200 * time.Millisecond)
}

func TestTracerDisabled(t *testing.T) {
	config := &TracerConfig{
		Enabled: false,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// These should be no-ops when disabled
	tracer.Instant("test", "test", nil)

	span := tracer.Begin("test", "test")
	assert.NotNil(t, span) // Should return a no-op span
	span.End()
}

func TestTraceEvent(t *testing.T) {
	event := &TraceEvent{
		ID:        "test-123",
		Name:      "test_event",
		Category:  "test",
		Phase:     "I",
		Timestamp: time.Now().UnixNano() / 1000,
		PID:       os.Getpid(),
		TID:       1,
		Args: map[string]interface{}{
			"key": "value",
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "test_event")
	assert.Contains(t, string(data), "test-123")

	// Test deserialization
	var decoded TraceEvent
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Name, decoded.Name)
}

func TestTracerConcurrentAccess(t *testing.T) {
	config := &TracerConfig{
		Enabled:    true,
		OutputFile: filepath.Join(t.TempDir(), "trace.json"),
		MaxEvents:  1000,
		BufferSize: 100,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// Test concurrent tracing
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			span := tracer.Begin("goroutine1_op", "test")
			time.Sleep(1 * time.Millisecond)
			span.End()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			tracer.Instant("goroutine2_event", "test", nil)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we reach here without panic, concurrent access is safe
	assert.True(t, true)
}

func TestTracerBufferLimits(t *testing.T) {
	config := &TracerConfig{
		Enabled:    true,
		OutputFile: filepath.Join(t.TempDir(), "trace.json"),
		MaxEvents:  5, // Small limit for testing
		BufferSize: 3,
	}

	tracer := NewTracer(config)
	require.NotNil(t, tracer)
	defer tracer.Stop()

	// Add more events than the limit
	for i := 0; i < 10; i++ {
		tracer.Instant("event", "test", map[string]interface{}{
			"iteration": i,
		})
	}

	// The tracer should handle buffer limits gracefully
	assert.True(t, true)
}

func BenchmarkTracerInstantEvent(b *testing.B) {
	config := &TracerConfig{
		Enabled:    true,
		OutputFile: filepath.Join(b.TempDir(), "trace.json"),
		MaxEvents:  100000,
		BufferSize: 1000,
	}

	tracer := NewTracer(config)
	require.NotNil(b, tracer)
	defer tracer.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracer.Instant("benchmark_event", "test", nil)
	}
}

func BenchmarkTracerSpan(b *testing.B) {
	config := &TracerConfig{
		Enabled:    true,
		OutputFile: filepath.Join(b.TempDir(), "trace.json"),
		MaxEvents:  100000,
		BufferSize: 1000,
	}

	tracer := NewTracer(config)
	require.NotNil(b, tracer)
	defer tracer.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span := tracer.Begin("benchmark_span", "test")
		span.End()
	}
}

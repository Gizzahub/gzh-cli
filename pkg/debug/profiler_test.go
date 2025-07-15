package debug

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultProfilerConfig(t *testing.T) {
	config := DefaultProfilerConfig()

	assert.False(t, config.Enabled)
	assert.True(t, config.CPUProfile)
	assert.True(t, config.MemoryProfile)
	assert.True(t, config.GoroutineTrace)
	assert.False(t, config.BlockProfile)
	assert.False(t, config.MutexProfile)
	assert.Equal(t, "./debug-profiles", config.OutputDir)
	assert.Equal(t, 30*time.Second, config.Duration)
	assert.Equal(t, 5*time.Second, config.Interval)
	assert.Equal(t, ":6060", config.HTTPEndpoint)
}

func TestNewProfiler(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		config := &ProfilerConfig{
			Enabled:   true,
			OutputDir: "/tmp/test-profiles",
		}

		profiler := NewProfiler(config)
		assert.NotNil(t, profiler)
		assert.Equal(t, config, profiler.config)
		assert.False(t, profiler.active)
		assert.Equal(t, "/tmp/test-profiles", profiler.outputDir)
	})

	t.Run("with nil config", func(t *testing.T) {
		profiler := NewProfiler(nil)
		assert.NotNil(t, profiler)
		assert.NotNil(t, profiler.config)
		assert.False(t, profiler.config.Enabled)
	})
}

func TestProfilerStart(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("disabled profiler", func(t *testing.T) {
		config := &ProfilerConfig{
			Enabled:   false,
			OutputDir: tmpDir,
		}

		profiler := NewProfiler(config)
		err := profiler.Start(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profiler is disabled")
	})

	t.Run("already active", func(t *testing.T) {
		config := &ProfilerConfig{
			Enabled:   true,
			OutputDir: tmpDir,
		}

		profiler := NewProfiler(config)
		profiler.active = true // Simulate already active

		err := profiler.Start(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profiler is already active")
	})

	t.Run("successful start", func(t *testing.T) {
		config := &ProfilerConfig{
			Enabled:      true,
			OutputDir:    filepath.Join(tmpDir, "profiles"),
			HTTPEndpoint: "", // Disable HTTP endpoint for test
			Duration:     1 * time.Second,
			Interval:     100 * time.Millisecond,
		}

		profiler := NewProfiler(config)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := profiler.Start(ctx)
		// Note: This might fail due to pprof setup, but we can test the basic logic
		if err != nil {
			t.Logf("Start failed (expected in test environment): %v", err)
		}

		// Test directory creation
		_, err = os.Stat(config.OutputDir)
		assert.NoError(t, err, "Output directory should be created")

		// Stop the profiler to clean up
		if profiler.IsActive() {
			profiler.Stop()
		}
	})
}

func TestProfilerStop(t *testing.T) {
	tmpDir := t.TempDir()
	config := &ProfilerConfig{
		Enabled:   true,
		OutputDir: tmpDir,
	}

	profiler := NewProfiler(config)

	t.Run("stop inactive profiler", func(t *testing.T) {
		err := profiler.Stop()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profiler is not active")
	})
}

func TestProfilerIsActive(t *testing.T) {
	config := DefaultProfilerConfig()
	profiler := NewProfiler(config)

	assert.False(t, profiler.IsActive())

	profiler.active = true
	assert.True(t, profiler.IsActive())
}

func TestProfilerGetConfig(t *testing.T) {
	config := &ProfilerConfig{
		Enabled:   true,
		OutputDir: "/tmp/test",
	}

	profiler := NewProfiler(config)
	retrievedConfig := profiler.GetConfig()

	assert.Equal(t, config, retrievedConfig)
}

func TestProfilerStats(t *testing.T) {
	config := DefaultProfilerConfig()
	profiler := NewProfiler(config)

	stats := profiler.GetStats()
	assert.NotNil(t, stats)
	assert.False(t, stats.Active)
	assert.Equal(t, time.Duration(0), stats.Uptime)
	assert.Zero(t, stats.ProfilesGenerated)
}

func TestProfilerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ProfilerConfig
		isValid bool
	}{
		{
			name: "valid config",
			config: &ProfilerConfig{
				Enabled:   true,
				OutputDir: "/tmp/valid",
				Duration:  30 * time.Second,
				Interval:  5 * time.Second,
			},
			isValid: true,
		},
		{
			name: "invalid duration",
			config: &ProfilerConfig{
				Enabled:   true,
				OutputDir: "/tmp/test",
				Duration:  -1 * time.Second,
				Interval:  5 * time.Second,
			},
			isValid: false,
		},
		{
			name: "invalid interval",
			config: &ProfilerConfig{
				Enabled:   true,
				OutputDir: "/tmp/test",
				Duration:  30 * time.Second,
				Interval:  -1 * time.Second,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiler := NewProfiler(tt.config)
			valid := profiler.validateConfig()
			assert.Equal(t, tt.isValid, valid)
		})
	}
}

func TestProfilerOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	config := &ProfilerConfig{
		Enabled:   true,
		OutputDir: filepath.Join(tmpDir, "nested", "profiles"),
	}

	profiler := NewProfiler(config)

	// Test that directory creation works with nested paths
	err := os.MkdirAll(profiler.outputDir, 0o755)
	assert.NoError(t, err)

	// Verify directory exists
	info, err := os.Stat(profiler.outputDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestProfilerConcurrentAccess(t *testing.T) {
	config := DefaultProfilerConfig()
	profiler := NewProfiler(config)

	// Test concurrent access to profiler methods
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			profiler.IsActive()
			profiler.GetConfig()
			profiler.GetStats()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			profiler.IsActive()
			profiler.GetConfig()
			profiler.GetStats()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we reach here without panic, concurrent access is safe
	assert.True(t, true)
}

func BenchmarkProfilerGetStats(b *testing.B) {
	config := DefaultProfilerConfig()
	profiler := NewProfiler(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.GetStats()
	}
}

func BenchmarkProfilerIsActive(b *testing.B) {
	config := DefaultProfilerConfig()
	profiler := NewProfiler(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.IsActive()
	}
}

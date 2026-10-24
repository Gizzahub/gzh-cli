//nolint:testpackage // White-box testing needed for internal function access
package profile

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewProfileAnalyzer(t *testing.T) {
	outputDir := "tmp/test-profiles"
	analyzer := NewProfileAnalyzer(outputDir)

	assert.NotNil(t, analyzer)
	assert.Equal(t, outputDir, analyzer.outputDir)
	assert.NotNil(t, analyzer.profiler)
}

func TestNewContinuousCmd(t *testing.T) {
	cmd := newContinuousCmd()

	assert.Equal(t, "continuous", cmd.Use)
	assert.Equal(t, "Run continuous profiling over time", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	intervalFlag := cmd.Flags().Lookup("interval")
	assert.NotNil(t, intervalFlag)
	assert.Equal(t, "5m0s", intervalFlag.DefValue)

	durationFlag := cmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)
	assert.Equal(t, "1h0m0s", durationFlag.DefValue)

	typeFlag := cmd.Flags().Lookup("type")
	assert.NotNil(t, typeFlag)
	assert.Equal(t, "cpu", typeFlag.DefValue)

	// --auto-analyze는 파일 이름만 보고 고정된 결과를 내던 플래그라 제거됐다.
	assert.Nil(t, cmd.Flags().Lookup("auto-analyze"))
}

func TestRunContinuousProfiling(t *testing.T) {
	analyzer := NewProfileAnalyzer("tmp/test-profiles")

	// Create a context with timeout for testing
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run continuous profiling for a very short duration
	err := analyzer.RunContinuousProfiling(ctx, "cpu", 50*time.Millisecond, 200*time.Millisecond)

	// Deadline/cancel surfaces as ctx.Err(); apprunner maps only Canceled → 130.
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

func TestGetEnhancedStats(t *testing.T) {
	stats := GetEnhancedStats()

	assert.NotNil(t, stats)

	// Check required top-level keys
	assert.Contains(t, stats, "timestamp")
	assert.Contains(t, stats, "memory")
	assert.Contains(t, stats, "gc")
	assert.Contains(t, stats, "runtime")

	// Check memory section
	memory := stats["memory"].(map[string]any) //nolint:forcetypeassert // test knows the type
	assert.Contains(t, memory, "heap_alloc")
	assert.Contains(t, memory, "heap_sys")
	assert.Contains(t, memory, "total_alloc")

	// Check GC section
	gc := stats["gc"].(map[string]any) //nolint:forcetypeassert // test knows the type
	assert.Contains(t, gc, "num_gc")
	assert.Contains(t, gc, "last_gc")

	// Check runtime section
	runtimeStats := stats["runtime"].(map[string]any) //nolint:forcetypeassert // test knows the type
	assert.Contains(t, runtimeStats, "goroutines")
	assert.Contains(t, runtimeStats, "num_cpu")
	assert.Contains(t, runtimeStats, "version")
}

func TestCommandLongDescriptions(t *testing.T) {
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"continuous", newContinuousCmd},
	}

	for _, tc := range commands {
		t.Run("command long description "+tc.name, func(t *testing.T) {
			cmd := tc.cmd()

			// Long description should be detailed
			assert.True(t, len(cmd.Long) > len(cmd.Short)*2,
				"Long description should be significantly longer than short")

			// Should contain examples
			assert.Contains(t, strings.ToLower(cmd.Long), "examples:",
				"Long description should contain examples")

			// Should contain the command name
			assert.Contains(t, cmd.Long, "gz profile "+tc.name,
				"Long description should show command usage")
		})
	}
}

func BenchmarkGetEnhancedStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stats := GetEnhancedStats()
		if stats == nil {
			b.Fatal("GetEnhancedStats returned nil")
		}
	}
}

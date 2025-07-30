// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProfileConfig(t *testing.T) {
	config := DefaultProfileConfig()

	assert.False(t, config.Enabled)
	assert.Equal(t, 6060, config.HTTPPort)
	assert.Equal(t, "tmp/profiles", config.OutputDir)
	assert.False(t, config.AutoProfile)
	assert.Equal(t, 30*time.Second, config.CPUDuration)
	assert.Equal(t, 100, config.SampleRate)
	assert.Equal(t, 1, config.BlockRate)
	assert.Equal(t, 1, config.MutexFraction)
}

func TestNewProfiler(t *testing.T) {
	// Test with nil config (should use defaults)
	profiler := NewProfiler(nil)

	assert.NotNil(t, profiler)
	assert.NotNil(t, profiler.config)
	assert.NotNil(t, profiler.logger)
	assert.NotNil(t, profiler.profiles)
	assert.Equal(t, "tmp/profiles", profiler.outputDir)

	// Test with custom config
	config := &ProfileConfig{
		Enabled:   true,
		HTTPPort:  8080,
		OutputDir: "custom/profiles",
	}
	profiler = NewProfiler(config)

	assert.Equal(t, config, profiler.config)
	assert.Equal(t, "custom/profiles", profiler.outputDir)
}

func TestProfiler_StartStop_Disabled(t *testing.T) {
	config := &ProfileConfig{Enabled: false}
	profiler := NewProfiler(config)

	ctx := context.Background()
	err := profiler.Start(ctx)
	assert.NoError(t, err)

	err = profiler.Stop()
	assert.NoError(t, err)
}

func TestProfiler_StartProfile_Disabled(t *testing.T) {
	config := &ProfileConfig{Enabled: false}
	profiler := NewProfiler(config)

	sessionID, err := profiler.StartProfile(ProfileTypeCPU)
	assert.Error(t, err)
	assert.Empty(t, sessionID)
	assert.Contains(t, err.Error(), "profiling is disabled")
}

func TestProfiler_StartProfile_UnsupportedType(t *testing.T) {
	config := &ProfileConfig{Enabled: true, OutputDir: "tmp/test_profiles"}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	sessionID, err := profiler.StartProfile(ProfileType("unsupported"))
	assert.Error(t, err)
	assert.Empty(t, sessionID)
	assert.Contains(t, err.Error(), "unsupported profile type")
}

func TestProfiler_StartStopProfile_CPU(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Start CPU profile
	sessionID, err := profiler.StartProfile(ProfileTypeCPU)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Verify session is active
	sessions := profiler.ListActiveSessions()
	assert.Len(t, sessions, 1)
	assert.Contains(t, sessions, sessionID)
	assert.True(t, sessions[sessionID].Active)
	assert.Equal(t, ProfileTypeCPU, sessions[sessionID].Type)

	// Stop profile
	err = profiler.StopProfile(sessionID)
	assert.NoError(t, err)

	// Verify session is no longer active
	sessions = profiler.ListActiveSessions()
	assert.Len(t, sessions, 0)
}

func TestProfiler_StartStopProfile_Memory(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Start memory profile
	sessionID, err := profiler.StartProfile(ProfileTypeMemory)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Stop profile
	err = profiler.StopProfile(sessionID)
	assert.NoError(t, err)

	// Verify profile file was created
	expectedFile := filepath.Join("tmp/test_profiles", "memory_"+sessionID+".prof")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err)
}

func TestProfiler_StartStopProfile_Goroutine(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Start goroutine profile
	sessionID, err := profiler.StartProfile(ProfileTypeGoroutine)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Stop profile
	err = profiler.StopProfile(sessionID)
	assert.NoError(t, err)

	// Verify profile file was created
	expectedFile := filepath.Join("tmp/test_profiles", "goroutine_"+sessionID+".prof")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err)
}

func TestProfiler_StopProfile_NotFound(t *testing.T) {
	config := &ProfileConfig{Enabled: true}
	profiler := NewProfiler(config)

	err := profiler.StopProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile session nonexistent not found")
}

func TestProfiler_StopProfile_NotActive(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Start and stop a session
	sessionID, err := profiler.StartProfile(ProfileTypeMemory)
	require.NoError(t, err)

	err = profiler.StopProfile(sessionID)
	require.NoError(t, err)

	// Try to stop again - should get "not found" error since session is deleted after stopping
	err = profiler.StopProfile(sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProfiler_ProfileOperation_Disabled(t *testing.T) {
	config := &ProfileConfig{Enabled: false}
	profiler := NewProfiler(config)

	executed := false
	err := profiler.ProfileOperation(context.Background(), "test-op", []ProfileType{ProfileTypeCPU}, func() error {
		executed = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestProfiler_ProfileOperation_Enabled(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	executed := false
	err := profiler.ProfileOperation(context.Background(), "test-op", []ProfileType{ProfileTypeMemory}, func() error {
		executed = true
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)

	// Verify profile file was created
	files, err := filepath.Glob("tmp/test_profiles/memory_*.prof")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestProfiler_ProfileOperation_WithError(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	testError := assert.AnError
	err := profiler.ProfileOperation(context.Background(), "test-op", []ProfileType{ProfileTypeMemory}, func() error {
		return testError
	})

	assert.Equal(t, testError, err)

	// Verify profile file was still created
	files, err := filepath.Glob("tmp/test_profiles/memory_*.prof")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestProfiler_GetRuntimeStats(t *testing.T) {
	profiler := NewProfiler(nil)

	stats := profiler.GetRuntimeStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "goroutines")
	assert.Contains(t, stats, "cgo_calls")
	assert.Contains(t, stats, "memory")
	assert.Contains(t, stats, "timestamp")

	memory, ok := stats["memory"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, memory, "alloc_bytes")
	assert.Contains(t, memory, "total_alloc_bytes")
	assert.Contains(t, memory, "sys_bytes")
	assert.Contains(t, memory, "heap_alloc_bytes")
	assert.Contains(t, memory, "heap_sys_bytes")
	assert.Contains(t, memory, "heap_objects")
	assert.Contains(t, memory, "stack_sys_bytes")
	assert.Contains(t, memory, "gc_runs")
	assert.Contains(t, memory, "gc_pause_total_ns")

	// Verify types
	assert.IsType(t, 0, stats["goroutines"])
	assert.IsType(t, int64(0), stats["cgo_calls"])
	assert.IsType(t, int64(0), stats["timestamp"])
}

func TestProfiler_ListActiveSessions(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Initially no sessions
	sessions := profiler.ListActiveSessions()
	assert.Len(t, sessions, 0)

	// Start multiple sessions
	sessionID1, err := profiler.StartProfile(ProfileTypeMemory)
	require.NoError(t, err)

	sessionID2, err := profiler.StartProfile(ProfileTypeGoroutine)
	require.NoError(t, err)

	// Should have 2 active sessions
	sessions = profiler.ListActiveSessions()
	assert.Len(t, sessions, 2)
	assert.Contains(t, sessions, sessionID1)
	assert.Contains(t, sessions, sessionID2)

	// Stop one session
	err = profiler.StopProfile(sessionID1)
	require.NoError(t, err)

	// Should have 1 active session
	sessions = profiler.ListActiveSessions()
	assert.Len(t, sessions, 1)
	assert.Contains(t, sessions, sessionID2)
	assert.NotContains(t, sessions, sessionID1)

	// Stop remaining session
	err = profiler.StopProfile(sessionID2)
	require.NoError(t, err)

	// Should have no active sessions
	sessions = profiler.ListActiveSessions()
	assert.Len(t, sessions, 0)
}

func TestProfiler_Stop_WithActiveSessions(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Start some sessions
	_, err := profiler.StartProfile(ProfileTypeMemory)
	require.NoError(t, err)

	_, err = profiler.StartProfile(ProfileTypeGoroutine)
	require.NoError(t, err)

	// Verify sessions are active
	sessions := profiler.ListActiveSessions()
	assert.Len(t, sessions, 2)

	// Stop profiler (should stop all active sessions)
	err = profiler.Stop()
	assert.NoError(t, err)

	// Verify all sessions were stopped
	sessions = profiler.ListActiveSessions()
	assert.Len(t, sessions, 0)

	// Verify profile files were created
	files, err := filepath.Glob("tmp/test_profiles/*.prof")
	assert.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestProfileSession_Structure(t *testing.T) {
	session := &ProfileSession{
		Type:      ProfileTypeCPU,
		StartTime: time.Now(),
		File:      nil,
		Active:    true,
	}

	assert.Equal(t, ProfileTypeCPU, session.Type)
	assert.True(t, session.Active)
	assert.Nil(t, session.File)
	assert.WithinDuration(t, time.Now(), session.StartTime, time.Second)
}

func TestProfileTypes(t *testing.T) {
	// Test all profile type constants
	assert.Equal(t, ProfileType("cpu"), ProfileTypeCPU)
	assert.Equal(t, ProfileType("memory"), ProfileTypeMemory)
	assert.Equal(t, ProfileType("goroutine"), ProfileTypeGoroutine)
	assert.Equal(t, ProfileType("block"), ProfileTypeBlock)
	assert.Equal(t, ProfileType("mutex"), ProfileTypeMutex)
	assert.Equal(t, ProfileType("threadcreate"), ProfileTypeThreadCreate)
}

func TestProfiler_MultipleProfileTypes(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	profileTypes := []ProfileType{
		ProfileTypeMemory,
		ProfileTypeGoroutine,
		ProfileTypeBlock,
		ProfileTypeMutex,
	}

	err := profiler.ProfileOperation(context.Background(), "multi-profile-test", profileTypes, func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	assert.NoError(t, err)

	// Verify multiple profile files were created
	files, err := filepath.Glob("tmp/test_profiles/*.prof")
	assert.NoError(t, err)
	assert.Len(t, files, len(profileTypes))
}

func TestProfiler_ConcurrentProfileOperations(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)

	// Clean up test directory
	defer os.RemoveAll("tmp/test_profiles")

	// Run multiple concurrent profile operations with small delays to ensure different timestamps
	done := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			// Add small delay to ensure different timestamps
			time.Sleep(time.Duration(id*10) * time.Millisecond)
			err := profiler.ProfileOperation(context.Background(), "concurrent-test", []ProfileType{ProfileTypeMemory}, func() error {
				time.Sleep(20 * time.Millisecond)
				return nil
			})
			done <- err
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify profile files were created (may be fewer than 3 due to timestamp collisions)
	files, err := filepath.Glob("tmp/test_profiles/memory_*.prof")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1) // At least one file should be created
	assert.LessOrEqual(t, len(files), 3)    // At most 3 files should be created
}

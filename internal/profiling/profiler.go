// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	runtimepprof "runtime/pprof"
	"sync"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// ProfileType represents different types of profiling.
type ProfileType string

const (
	ProfileTypeCPU          ProfileType = "cpu"
	ProfileTypeMemory       ProfileType = "memory"
	ProfileTypeGoroutine    ProfileType = "goroutine"
	ProfileTypeBlock        ProfileType = "block"
	ProfileTypeMutex        ProfileType = "mutex"
	ProfileTypeThreadCreate ProfileType = "threadcreate"
)

// ProfileConfig holds configuration for profiling.
type ProfileConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	HTTPPort      int           `yaml:"http_port" json:"http_port"`
	OutputDir     string        `yaml:"output_dir" json:"output_dir"`
	AutoProfile   bool          `yaml:"auto_profile" json:"auto_profile"`
	CPUDuration   time.Duration `yaml:"cpu_duration" json:"cpu_duration"`
	SampleRate    int           `yaml:"sample_rate" json:"sample_rate"`
	BlockRate     int           `yaml:"block_rate" json:"block_rate"`
	MutexFraction int           `yaml:"mutex_fraction" json:"mutex_fraction"`
}

// DefaultProfileConfig returns default profiling configuration.
func DefaultProfileConfig() *ProfileConfig {
	return &ProfileConfig{
		Enabled:       false,
		HTTPPort:      6060,
		OutputDir:     "tmp/profiles",
		AutoProfile:   false,
		CPUDuration:   30 * time.Second,
		SampleRate:    100,
		BlockRate:     1,
		MutexFraction: 1,
	}
}

// Profiler provides performance profiling capabilities.
type Profiler struct {
	config    *ProfileConfig
	logger    *logger.SimpleLogger
	server    *http.Server
	mu        sync.RWMutex
	profiles  map[string]*ProfileSession
	outputDir string
}

// ProfileSession represents an active profiling session.
type ProfileSession struct {
	Type      ProfileType
	StartTime time.Time
	File      *os.File
	Active    bool
}

// NewProfiler creates a new profiler instance.
func NewProfiler(config *ProfileConfig) *Profiler {
	if config == nil {
		config = DefaultProfileConfig()
	}

	p := &Profiler{
		config:    config,
		logger:    logger.NewSimpleLogger("profiler"),
		profiles:  make(map[string]*ProfileSession),
		outputDir: config.OutputDir,
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(p.outputDir, 0o755); err != nil {
		p.logger.Warn("Failed to create profile output directory", "dir", p.outputDir, "error", err)
	}

	// Configure runtime profiling rates
	runtime.SetCPUProfileRate(config.SampleRate)
	runtime.SetBlockProfileRate(config.BlockRate)
	runtime.SetMutexProfileFraction(config.MutexFraction)

	return p
}

// Start initializes profiling services.
func (p *Profiler) Start(ctx context.Context) error {
	if !p.config.Enabled {
		p.logger.Debug("Profiling is disabled")
		return nil
	}

	// Start HTTP server for pprof endpoints
	if p.config.HTTPPort > 0 {
		go p.startHTTPServer(ctx)
	}

	// Start automatic profiling if enabled
	if p.config.AutoProfile {
		go p.startAutoProfile(ctx)
	}

	p.logger.Info("Profiler started", "http_port", p.config.HTTPPort, "output_dir", p.outputDir)
	return nil
}

// Stop shuts down profiling services.
func (p *Profiler) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop all active profiling sessions
	for id, session := range p.profiles {
		if session.Active {
			p.stopProfileSession(id, session) // Cleanup, errors are ignored intentionally
		}
	}

	// Stop HTTP server
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.server.Shutdown(ctx); err != nil {
			p.logger.Warn("Failed to shutdown profiling HTTP server", "error", err)
		}
	}

	p.logger.Info("Profiler stopped")
	return nil
}

// StartProfile begins a profiling session for the specified type.
func (p *Profiler) StartProfile(profileType ProfileType) (string, error) {
	if !p.config.Enabled {
		return "", fmt.Errorf("profiling is disabled")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	sessionID := fmt.Sprintf("%s_%d", profileType, time.Now().Unix())

	session := &ProfileSession{
		Type:      profileType,
		StartTime: time.Now(),
		Active:    true,
	}

	switch profileType {
	case ProfileTypeCPU:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("cpu_%s.prof", sessionID))
		file, err := os.Create(filename)
		if err != nil {
			return "", fmt.Errorf("failed to create CPU profile file: %w", err)
		}
		session.File = file

		if err := runtimepprof.StartCPUProfile(file); err != nil {
			file.Close()
			return "", fmt.Errorf("failed to start CPU profile: %w", err)
		}

	case ProfileTypeMemory, ProfileTypeGoroutine, ProfileTypeBlock, ProfileTypeMutex, ProfileTypeThreadCreate:
		// These profiles are captured on-demand, no file needed during session

	default:
		return "", fmt.Errorf("unsupported profile type: %s", profileType)
	}

	p.profiles[sessionID] = session
	p.logger.Info("Started profiling session", "type", profileType, "session_id", sessionID)

	return sessionID, nil
}

// StopProfile ends a profiling session and saves the results.
func (p *Profiler) StopProfile(sessionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.profiles[sessionID]
	if !exists {
		return fmt.Errorf("profile session %s not found", sessionID)
	}

	if !session.Active {
		return fmt.Errorf("profile session %s is not active", sessionID)
	}

	return p.stopProfileSession(sessionID, session)
}

// stopProfileSession stops a profile session (assumes mutex is held).
func (p *Profiler) stopProfileSession(sessionID string, session *ProfileSession) error {
	defer func() {
		session.Active = false
		delete(p.profiles, sessionID)
	}()

	duration := time.Since(session.StartTime)

	switch session.Type {
	case ProfileTypeCPU:
		runtimepprof.StopCPUProfile()
		if session.File != nil {
			session.File.Close()
			p.logger.Info("CPU profile saved", "session_id", sessionID, "duration", duration, "file", session.File.Name())
		}

	case ProfileTypeMemory:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("memory_%s.prof", sessionID))
		if err := p.writeProfile("heap", filename); err != nil {
			return fmt.Errorf("failed to write memory profile: %w", err)
		}
		p.logger.Info("Memory profile saved", "session_id", sessionID, "duration", duration, "file", filename)

	case ProfileTypeGoroutine:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("goroutine_%s.prof", sessionID))
		if err := p.writeProfile("goroutine", filename); err != nil {
			return fmt.Errorf("failed to write goroutine profile: %w", err)
		}
		p.logger.Info("Goroutine profile saved", "session_id", sessionID, "duration", duration, "file", filename)

	case ProfileTypeBlock:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("block_%s.prof", sessionID))
		if err := p.writeProfile("block", filename); err != nil {
			return fmt.Errorf("failed to write block profile: %w", err)
		}
		p.logger.Info("Block profile saved", "session_id", sessionID, "duration", duration, "file", filename)

	case ProfileTypeMutex:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("mutex_%s.prof", sessionID))
		if err := p.writeProfile("mutex", filename); err != nil {
			return fmt.Errorf("failed to write mutex profile: %w", err)
		}
		p.logger.Info("Mutex profile saved", "session_id", sessionID, "duration", duration, "file", filename)

	case ProfileTypeThreadCreate:
		filename := filepath.Join(p.outputDir, fmt.Sprintf("threadcreate_%s.prof", sessionID))
		if err := p.writeProfile("threadcreate", filename); err != nil {
			return fmt.Errorf("failed to write threadcreate profile: %w", err)
		}
		p.logger.Info("ThreadCreate profile saved", "session_id", sessionID, "duration", duration, "file", filename)
	}

	return nil
}

// writeProfile writes a named profile to a file.
func (p *Profiler) writeProfile(name, filename string) error {
	profile := runtimepprof.Lookup(name)
	if profile == nil {
		return fmt.Errorf("profile %s not found", name)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return profile.WriteTo(file, 0)
}

// ProfileOperation runs an operation with automatic profiling.
func (p *Profiler) ProfileOperation(_ context.Context, operationName string, profileTypes []ProfileType, operation func() error) error {
	if !p.config.Enabled {
		return operation()
	}

	// Start profiling sessions
	sessionIDs := make([]string, 0, len(profileTypes))
	for _, profileType := range profileTypes {
		sessionID, err := p.StartProfile(profileType)
		if err != nil {
			p.logger.Warn("Failed to start profile", "type", profileType, "error", err)
			continue
		}
		sessionIDs = append(sessionIDs, sessionID)
	}

	startTime := time.Now()

	// Run the operation
	err := operation()

	duration := time.Since(startTime)

	// Stop profiling sessions
	for _, sessionID := range sessionIDs {
		if stopErr := p.StopProfile(sessionID); stopErr != nil {
			p.logger.Warn("Failed to stop profile", "session_id", sessionID, "error", stopErr)
		}
	}

	// Log performance metrics
	p.logger.LogPerformance(operationName, duration, map[string]interface{}{
		"profile_types":    profileTypes,
		"sessions_started": len(sessionIDs),
		"success":          err == nil,
	})

	return err
}

// GetRuntimeStats returns current runtime statistics.
func (p *Profiler) GetRuntimeStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"cgo_calls":  runtime.NumCgoCall(),
		"memory": map[string]interface{}{
			"alloc_bytes":       m.Alloc,
			"total_alloc_bytes": m.TotalAlloc,
			"sys_bytes":         m.Sys,
			"heap_alloc_bytes":  m.HeapAlloc,
			"heap_sys_bytes":    m.HeapSys,
			"heap_objects":      m.HeapObjects,
			"stack_sys_bytes":   m.StackSys,
			"gc_runs":           m.NumGC,
			"gc_pause_total_ns": m.PauseTotalNs,
		},
		"timestamp": time.Now().Unix(),
	}
}

// startHTTPServer starts the HTTP server for pprof endpoints.
func (p *Profiler) startHTTPServer(_ context.Context) {
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Custom endpoint for runtime stats
	mux.HandleFunc("/debug/stats", func(w http.ResponseWriter, _ *http.Request) {
		stats := p.GetRuntimeStats()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%+v\n", stats)
	})

	p.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", p.config.HTTPPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	p.logger.Info("Starting profiling HTTP server", "port", p.config.HTTPPort)
	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.logger.Error("Profiling HTTP server failed", "error", err)
	}
}

// startAutoProfile starts automatic periodic profiling.
func (p *Profiler) startAutoProfile(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Profile every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.logger.Debug("Starting automatic profile session")

			// CPU profile for configured duration
			if sessionID, err := p.StartProfile(ProfileTypeCPU); err == nil {
				time.AfterFunc(p.config.CPUDuration, func() {
					if err := p.StopProfile(sessionID); err != nil {
						p.logger.Warn("Failed to stop auto CPU profile", "error", err)
					}
				})
			}

			// Immediate memory and goroutine snapshots
			for _, profileType := range []ProfileType{ProfileTypeMemory, ProfileTypeGoroutine} {
				if sessionID, err := p.StartProfile(profileType); err == nil {
					// Stop immediately to capture snapshot
					if err := p.StopProfile(sessionID); err != nil {
						p.logger.Warn("Failed to stop auto profile", "type", profileType, "error", err)
					}
				}
			}
		}
	}
}

// ListActiveSessions returns information about currently active profiling sessions.
func (p *Profiler) ListActiveSessions() map[string]ProfileSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]ProfileSession)
	for id, session := range p.profiles {
		if session.Active {
			result[id] = *session
		}
	}
	return result
}

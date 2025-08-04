// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package simpleprof provides simplified profiling using standard Go pprof.
// This replaces the complex internal/profiling package with lightweight alternatives.
package simpleprof

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Import pprof HTTP handlers
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// ProfileType represents the type of profile to collect.
type ProfileType string

const (
	ProfileTypeCPU       ProfileType = "cpu"
	ProfileTypeMemory    ProfileType = "memory"
	ProfileTypeGoroutine ProfileType = "goroutine" 
	ProfileTypeBlock     ProfileType = "block"
	ProfileTypeMutex     ProfileType = "mutex"
)

// SimpleProfiler provides basic profiling functionality using standard Go pprof.
type SimpleProfiler struct {
	outputDir string
	server    *http.Server
}

// NewSimpleProfiler creates a new simple profiler.
func NewSimpleProfiler(outputDir string) *SimpleProfiler {
	if outputDir == "" {
		outputDir = "tmp/profiles"
	}
	
	// Ensure output directory exists
	os.MkdirAll(outputDir, 0755)
	
	return &SimpleProfiler{
		outputDir: outputDir,
	}
}

// StartHTTPServer starts the pprof HTTP server on the specified port.
func (p *SimpleProfiler) StartHTTPServer(port int) error {
	addr := fmt.Sprintf("localhost:%d", port)
	
	p.server = &http.Server{
		Addr: addr,
	}
	
	log.Printf("Starting pprof server at http://%s/debug/pprof/", addr)
	log.Printf("Available profiles:")
	log.Printf("  - CPU: http://%s/debug/pprof/profile", addr)
	log.Printf("  - Heap: http://%s/debug/pprof/heap", addr)
	log.Printf("  - Goroutines: http://%s/debug/pprof/goroutine", addr)
	log.Printf("  - Block: http://%s/debug/pprof/block", addr)
	log.Printf("  - Mutex: http://%s/debug/pprof/mutex", addr)
	
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("pprof server error: %v", err)
		}
	}()
	
	return nil
}

// StopHTTPServer stops the pprof HTTP server.
func (p *SimpleProfiler) StopHTTPServer(ctx context.Context) error {
	if p.server == nil {
		return nil
	}
	
	return p.server.Shutdown(ctx)
}

// StartProfile starts collecting a profile of the specified type.
func (p *SimpleProfiler) StartProfile(profileType ProfileType, duration time.Duration) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s_%s.prof", profileType, timestamp))
	
	switch profileType {
	case ProfileTypeCPU:
		return p.startCPUProfile(filename, duration)
	case ProfileTypeMemory:
		return p.saveMemoryProfile(filename)
	case ProfileTypeGoroutine:
		return p.saveGoroutineProfile(filename)
	case ProfileTypeBlock:
		return p.saveBlockProfile(filename)
	case ProfileTypeMutex:
		return p.saveMutexProfile(filename)
	default:
		return "", fmt.Errorf("unsupported profile type: %s", profileType)
	}
}

func (p *SimpleProfiler) startCPUProfile(filename string, duration time.Duration) (string, error) {
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("could not create CPU profile: %w", err)
	}
	
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return "", fmt.Errorf("could not start CPU profile: %w", err)
	}
	
	// Stop CPU profiling after duration
	go func() {
		time.Sleep(duration)
		pprof.StopCPUProfile()
		f.Close()
		log.Printf("CPU profile saved to %s", filename)
	}()
	
	return filename, nil
}

func (p *SimpleProfiler) saveMemoryProfile(filename string) (string, error) {
	return p.saveProfile("heap", filename)
}

func (p *SimpleProfiler) saveGoroutineProfile(filename string) (string, error) {
	return p.saveProfile("goroutine", filename)
}

func (p *SimpleProfiler) saveBlockProfile(filename string) (string, error) {
	runtime.SetBlockProfileRate(1)
	return p.saveProfile("block", filename)
}

func (p *SimpleProfiler) saveMutexProfile(filename string) (string, error) {
	runtime.SetMutexProfileFraction(1)
	return p.saveProfile("mutex", filename)
}

func (p *SimpleProfiler) saveProfile(profileName, filename string) (string, error) {
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("could not create %s profile: %w", profileName, err)
	}
	defer f.Close()
	
	profile := pprof.Lookup(profileName)
	if profile == nil {
		return "", fmt.Errorf("profile %s not found", profileName)
	}
	
	if err := profile.WriteTo(f, 0); err != nil {
		return "", fmt.Errorf("could not write %s profile: %w", profileName, err)
	}
	
	log.Printf("%s profile saved to %s", profileName, filename)
	return filename, nil
}

// GetStats returns basic runtime statistics.
func (p *SimpleProfiler) GetStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_inuse":     m.HeapInuse,
		"heap_released":  m.HeapReleased,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"gc_runs":        m.NumGC,
		"last_gc":        time.Unix(0, int64(m.LastGC)),
		"pause_total_ns": m.PauseTotalNs,
	}
}
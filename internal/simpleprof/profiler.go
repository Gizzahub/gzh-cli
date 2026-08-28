// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package simpleprof provides simplified profiling using standard Go pprof.
// This replaces the complex internal/profiling package with lightweight alternatives.
package simpleprof

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// LastGCTime converts runtime.MemStats.LastGC to time.Time without narrowing
// the unsigned nanosecond timestamp to a signed integer.
func LastGCTime(lastGC uint64) time.Time {
	const nanosPerSecond = uint64(time.Second)
	seconds := lastGC / nanosPerSecond
	nanoseconds := lastGC % nanosPerSecond
	if seconds > math.MaxInt64 {
		seconds = math.MaxInt64
	}

	return time.Unix(int64(seconds), int64(nanoseconds))
}

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
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		// If we can't create the directory, fall back to current directory
		outputDir = "."
	}

	return &SimpleProfiler{
		outputDir: outputDir,
	}
}

// StartHTTPServer starts the pprof HTTP server on the specified port.
func (p *SimpleProfiler) StartHTTPServer(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	p.server = &http.Server{
		Addr:              addr,
		Handler:           newPprofServeMux(),
		ReadHeaderTimeout: 10 * time.Second,
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

func newPprofServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprofGETHandler(httppprof.Index))
	mux.HandleFunc("/debug/pprof/cmdline", pprofGETHandler(httppprof.Cmdline))
	mux.HandleFunc("/debug/pprof/profile", pprofGETHandler(httppprof.Profile))
	mux.HandleFunc("/debug/pprof/symbol", pprofGETHandler(httppprof.Symbol))
	mux.HandleFunc("/debug/pprof/trace", pprofGETHandler(httppprof.Trace))
	return mux
}

func pprofGETHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
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
	return p.saveProfile("block", filename)
}

func (p *SimpleProfiler) saveMutexProfile(filename string) (string, error) {
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
func (p *SimpleProfiler) GetStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"goroutines":     runtime.NumGoroutine(),
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_inuse":     m.HeapInuse,
		"heap_released":  m.HeapReleased,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"gc_runs":        m.NumGC,
		"last_gc":        LastGCTime(m.LastGC),
		"pause_total_ns": m.PauseTotalNs,
	}
}

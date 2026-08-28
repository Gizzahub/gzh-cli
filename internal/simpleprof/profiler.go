// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package simpleprof provides simplified profiling using standard Go pprof.
// This replaces the complex internal/profiling package with lightweight alternatives.
package simpleprof

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
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
	outputDir     string
	server        *http.Server
	now           func() time.Time
	nextSequence  func() uint64
	openFile      func(string, int, os.FileMode) (*os.File, error)
	closeFile     func(*os.File) error
	removeFile    func(string) error
	lookupProfile func(string) profileWriter
	startCPU      func(io.Writer) error
	stopCPU       func()
	sleep         func(time.Duration)
	cpuDone       func()
}

type profileWriter interface {
	WriteTo(io.Writer, int) error
}

const maxProfileReservationAttempts = 64

// 프로세스 내 예약 순서는 프로파일러 인스턴스가 달라도 공유한다.
var profileSequence atomic.Uint64

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
		now:       time.Now,
		nextSequence: func() uint64 {
			return profileSequence.Add(1)
		},
		openFile:      os.OpenFile,
		closeFile:     (*os.File).Close,
		removeFile:    os.Remove,
		lookupProfile: func(name string) profileWriter { return pprof.Lookup(name) },
		startCPU:      pprof.StartCPUProfile,
		stopCPU:       pprof.StopCPUProfile,
		sleep:         time.Sleep,
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
	profileName, err := profileNameForType(profileType)
	if err != nil {
		return "", err
	}

	filename, file, err := p.reserveProfile(profileType)
	if err != nil {
		return "", err
	}

	if profileType == ProfileTypeCPU {
		return p.startCPUProfile(filename, file, duration)
	}

	return p.saveProfile(profileName, filename, file)
}

func profileNameForType(profileType ProfileType) (string, error) {
	switch profileType {
	case ProfileTypeCPU:
		return "cpu", nil
	case ProfileTypeMemory:
		return "heap", nil
	case ProfileTypeGoroutine:
		return "goroutine", nil
	case ProfileTypeBlock:
		return "block", nil
	case ProfileTypeMutex:
		return "mutex", nil
	default:
		return "", fmt.Errorf("unsupported profile type: %s", profileType)
	}
}

func (p *SimpleProfiler) reserveProfile(profileType ProfileType) (string, *os.File, error) {
	timestamp := p.now().Format("20060102_150405")
	lastCandidate := ""

	for attempt := 1; attempt <= maxProfileReservationAttempts; attempt++ {
		sequence := p.nextSequence()
		filename := filepath.Join(p.outputDir, fmt.Sprintf("%s_%s_%d.prof", profileType, timestamp, sequence))
		lastCandidate = filename
		file, err := p.openFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return filename, file, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", nil, fmt.Errorf("could not reserve %s profile %q: %w", profileType, filename, err)
		}
	}

	return "", nil, fmt.Errorf("could not reserve %s profile after %d attempts (last candidate %q): %w", profileType, maxProfileReservationAttempts, lastCandidate, os.ErrExist)
}

func (p *SimpleProfiler) startCPUProfile(filename string, file *os.File, duration time.Duration) (string, error) {
	if err := p.startCPU(file); err != nil {
		return "", p.discardReservedProfile(file, filename, fmt.Errorf("could not start CPU profile: %w", err))
	}

	// Stop CPU profiling after duration
	go func() {
		p.sleep(duration)
		p.stopCPU()
		if err := p.closeFile(file); err != nil {
			log.Printf("could not close CPU profile %s: %v", filename, err)
		}
		if p.cpuDone != nil {
			p.cpuDone()
		}
		log.Printf("CPU profile saved to %s", filename)
	}()

	return filename, nil
}

func (p *SimpleProfiler) saveProfile(profileName, filename string, file *os.File) (string, error) {
	profile := p.lookupProfile(profileName)
	if profile == nil {
		return "", p.discardReservedProfile(file, filename, fmt.Errorf("profile %s not found", profileName))
	}

	if err := profile.WriteTo(file, 0); err != nil {
		return "", p.discardReservedProfile(file, filename, fmt.Errorf("could not write %s profile: %w", profileName, err))
	}

	if err := p.closeFile(file); err != nil {
		return "", p.removeReservedProfile(filename, fmt.Errorf("could not close %s profile: %w", profileName, err))
	}

	log.Printf("%s profile saved to %s", profileName, filename)
	return filename, nil
}

// 예약한 파일만 닫은 뒤 제거하여 기존 프로파일을 건드리지 않는다.
func (p *SimpleProfiler) discardReservedProfile(file *os.File, filename string, primary error) error {
	closeErr := p.closeFile(file)
	return p.removeReservedProfile(filename, errors.Join(primary, closeErr))
}

func (p *SimpleProfiler) removeReservedProfile(filename string, primary error) error {
	return errors.Join(primary, p.removeFile(filename))
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

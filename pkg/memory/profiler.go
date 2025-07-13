package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
)

// Profiler handles memory and CPU profiling for performance analysis
type Profiler struct {
	config       ProfilerConfig
	active       bool
	mu           sync.RWMutex
	profileFiles map[string]*os.File
	stopCh       chan struct{}
}

// ProfilerConfig configures the profiler behavior
type ProfilerConfig struct {
	// Output directory for profile files
	OutputDir string

	// CPU profiling duration
	CPUProfileDuration time.Duration

	// Memory profile interval
	MemProfileInterval time.Duration

	// Enable trace profiling
	TraceEnabled bool

	// Trace duration
	TraceDuration time.Duration

	// Enable block profiling
	BlockProfileEnabled bool

	// Block profile rate
	BlockProfileRate int

	// Enable mutex profiling
	MutexProfileEnabled bool

	// Mutex profile fraction
	MutexProfileFraction int

	// Enable goroutine profiling
	GoroutineProfileEnabled bool

	// Profile file prefix
	FilePrefix string
}

// DefaultProfilerConfig returns default profiler configuration
func DefaultProfilerConfig() ProfilerConfig {
	return ProfilerConfig{
		OutputDir:               "./profiles",
		CPUProfileDuration:      30 * time.Second,
		MemProfileInterval:      1 * time.Minute,
		TraceEnabled:            false,
		TraceDuration:           10 * time.Second,
		BlockProfileEnabled:     false,
		BlockProfileRate:        1,
		MutexProfileEnabled:     false,
		MutexProfileFraction:    1,
		GoroutineProfileEnabled: true,
		FilePrefix:              "gzh-manager",
	}
}

// ProfileSnapshot represents a snapshot of profiling data
type ProfileSnapshot struct {
	Timestamp    time.Time
	MemStats     runtime.MemStats
	NumGoroutine int
	CPUProfile   string
	MemProfile   string
	TraceFile    string
	Metadata     map[string]interface{}
}

// NewProfiler creates a new profiler with the given configuration
func NewProfiler(config ProfilerConfig) *Profiler {
	return &Profiler{
		config:       config,
		profileFiles: make(map[string]*os.File),
		stopCh:       make(chan struct{}),
	}
}

// Start begins profiling with the configured settings
func (p *Profiler) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.active {
		p.mu.Unlock()
		return fmt.Errorf("profiler is already active")
	}
	p.active = true
	p.mu.Unlock()

	// Create output directory
	if err := os.MkdirAll(p.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Enable runtime profiling
	p.enableRuntimeProfiling()

	// Start periodic profiling
	go p.periodicMemoryProfiling(ctx)

	if p.config.TraceEnabled {
		go p.periodicTraceProfiling(ctx)
	}

	return nil
}

// Stop stops all profiling activities
func (p *Profiler) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return nil
	}

	close(p.stopCh)
	p.active = false

	// Close all open profile files
	for name, file := range p.profileFiles {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing profile file %s: %v\n", name, err)
		}
	}
	p.profileFiles = make(map[string]*os.File)

	// Disable runtime profiling
	p.disableRuntimeProfiling()

	return nil
}

// enableRuntimeProfiling enables various runtime profiling features
func (p *Profiler) enableRuntimeProfiling() {
	if p.config.BlockProfileEnabled {
		runtime.SetBlockProfileRate(p.config.BlockProfileRate)
	}

	if p.config.MutexProfileEnabled {
		runtime.SetMutexProfileFraction(p.config.MutexProfileFraction)
	}
}

// disableRuntimeProfiling disables runtime profiling features
func (p *Profiler) disableRuntimeProfiling() {
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)
}

// periodicMemoryProfiling performs memory profiling at regular intervals
func (p *Profiler) periodicMemoryProfiling(ctx context.Context) {
	ticker := time.NewTicker(p.config.MemProfileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			if err := p.CaptureMemoryProfile(); err != nil {
				fmt.Printf("Error capturing memory profile: %v\n", err)
			}
		}
	}
}

// periodicTraceProfiling performs trace profiling at regular intervals
func (p *Profiler) periodicTraceProfiling(ctx context.Context) {
	ticker := time.NewTicker(p.config.TraceDuration * 2) // Start trace every 2x duration
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			if err := p.CaptureTrace(p.config.TraceDuration); err != nil {
				fmt.Printf("Error capturing trace: %v\n", err)
			}
		}
	}
}

// CaptureCPUProfile captures CPU profile for the specified duration
func (p *Profiler) CaptureCPUProfile(duration time.Duration) error {
	filename := fmt.Sprintf("%s_cpu_%s.prof", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing CPU profile file: %v\n", err)
		}
	}()

	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	time.Sleep(duration)
	pprof.StopCPUProfile()

	fmt.Printf("CPU profile saved to: %s\n", filepath)
	return nil
}

// CaptureMemoryProfile captures a memory profile
func (p *Profiler) CaptureMemoryProfile() error {
	filename := fmt.Sprintf("%s_mem_%s.prof", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing memory profile file: %v\n", err)
		}
	}()

	runtime.GC() // Force GC before capturing memory profile

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	fmt.Printf("Memory profile saved to: %s\n", filepath)
	return nil
}

// CaptureTrace captures execution trace for the specified duration
func (p *Profiler) CaptureTrace(duration time.Duration) error {
	filename := fmt.Sprintf("%s_trace_%s.out", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create trace file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing trace file: %v\n", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		return fmt.Errorf("failed to start trace: %w", err)
	}

	time.Sleep(duration)
	trace.Stop()

	fmt.Printf("Trace saved to: %s\n", filepath)
	return nil
}

// CaptureGoroutineProfile captures goroutine profile
func (p *Profiler) CaptureGoroutineProfile() error {
	filename := fmt.Sprintf("%s_goroutine_%s.prof", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing goroutine profile file: %v\n", err)
		}
	}()

	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	fmt.Printf("Goroutine profile saved to: %s\n", filepath)
	return nil
}

// CaptureBlockProfile captures block profile
func (p *Profiler) CaptureBlockProfile() error {
	if !p.config.BlockProfileEnabled {
		return fmt.Errorf("block profiling is not enabled")
	}

	filename := fmt.Sprintf("%s_block_%s.prof", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create block profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing block profile file: %v\n", err)
		}
	}()

	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	fmt.Printf("Block profile saved to: %s\n", filepath)
	return nil
}

// CaptureMutexProfile captures mutex profile
func (p *Profiler) CaptureMutexProfile() error {
	if !p.config.MutexProfileEnabled {
		return fmt.Errorf("mutex profiling is not enabled")
	}

	filename := fmt.Sprintf("%s_mutex_%s.prof", p.config.FilePrefix, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(p.config.OutputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create mutex profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing mutex profile file: %v\n", err)
		}
	}()

	if err := pprof.Lookup("mutex").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	fmt.Printf("Mutex profile saved to: %s\n", filepath)
	return nil
}

// CaptureAllProfiles captures all enabled profile types
func (p *Profiler) CaptureAllProfiles() error {
	var errors []error

	// CPU profile
	if err := p.CaptureCPUProfile(p.config.CPUProfileDuration); err != nil {
		errors = append(errors, fmt.Errorf("CPU profile error: %w", err))
	}

	// Memory profile
	if err := p.CaptureMemoryProfile(); err != nil {
		errors = append(errors, fmt.Errorf("memory profile error: %w", err))
	}

	// Goroutine profile
	if p.config.GoroutineProfileEnabled {
		if err := p.CaptureGoroutineProfile(); err != nil {
			errors = append(errors, fmt.Errorf("goroutine profile error: %w", err))
		}
	}

	// Block profile
	if p.config.BlockProfileEnabled {
		if err := p.CaptureBlockProfile(); err != nil {
			errors = append(errors, fmt.Errorf("block profile error: %w", err))
		}
	}

	// Mutex profile
	if p.config.MutexProfileEnabled {
		if err := p.CaptureMutexProfile(); err != nil {
			errors = append(errors, fmt.Errorf("mutex profile error: %w", err))
		}
	}

	// Trace
	if p.config.TraceEnabled {
		if err := p.CaptureTrace(p.config.TraceDuration); err != nil {
			errors = append(errors, fmt.Errorf("trace error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("profiling errors: %v", errors)
	}

	return nil
}

// GetSnapshot returns a snapshot of current profiling data
func (p *Profiler) GetSnapshot() ProfileSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ProfileSnapshot{
		Timestamp:    time.Now(),
		MemStats:     m,
		NumGoroutine: runtime.NumGoroutine(),
		Metadata: map[string]interface{}{
			"GOMAXPROCS": runtime.GOMAXPROCS(0),
			"NumCPU":     runtime.NumCPU(),
			"Version":    runtime.Version(),
		},
	}
}

// AnalyzeMemoryUsage provides memory usage analysis
func (p *Profiler) AnalyzeMemoryUsage() MemoryAnalysis {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	analysis := MemoryAnalysis{
		Timestamp: time.Now(),
		HeapUsage: HeapUsage{
			Allocated:   m.HeapAlloc,
			System:      m.HeapSys,
			InUse:       m.HeapInuse,
			Released:    m.HeapReleased,
			Objects:     m.HeapObjects,
			Utilization: float64(m.HeapInuse) / float64(m.HeapSys),
		},
		GCStats: GCAnalysis{
			NumGC:      m.NumGC,
			TotalPause: time.Duration(m.PauseTotalNs),
			AvgPause:   time.Duration(m.PauseTotalNs / uint64(max(m.NumGC, 1))),
			LastGC:     time.Unix(0, int64(m.LastGC)),
		},
		Goroutines: runtime.NumGoroutine(),
	}

	// Calculate memory growth rate if we have previous data
	analysis.MemoryGrowthRate = p.calculateGrowthRate(m.HeapAlloc)

	return analysis
}

// MemoryAnalysis provides detailed memory usage analysis
type MemoryAnalysis struct {
	Timestamp        time.Time
	HeapUsage        HeapUsage
	GCStats          GCAnalysis
	Goroutines       int
	MemoryGrowthRate float64 // bytes per second
}

// HeapUsage provides heap memory usage details
type HeapUsage struct {
	Allocated   uint64  // bytes allocated and in use
	System      uint64  // bytes obtained from system
	InUse       uint64  // bytes in use
	Released    uint64  // bytes released to system
	Objects     uint64  // number of allocated objects
	Utilization float64 // heap utilization ratio
}

// GCAnalysis provides garbage collection analysis
type GCAnalysis struct {
	NumGC      uint32
	TotalPause time.Duration
	AvgPause   time.Duration
	LastGC     time.Time
}

var (
	lastHeapAlloc uint64
	lastHeapTime  time.Time
)

// calculateGrowthRate calculates memory growth rate
func (p *Profiler) calculateGrowthRate(currentHeap uint64) float64 {
	now := time.Now()
	if lastHeapTime.IsZero() {
		lastHeapAlloc = currentHeap
		lastHeapTime = now
		return 0
	}

	timeDiff := now.Sub(lastHeapTime).Seconds()
	if timeDiff == 0 {
		return 0
	}

	heapDiff := int64(currentHeap) - int64(lastHeapAlloc)
	growthRate := float64(heapDiff) / timeDiff

	lastHeapAlloc = currentHeap
	lastHeapTime = now

	return growthRate
}

// max returns the maximum of two uint32 values
func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

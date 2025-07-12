// Package debug provides advanced debugging and profiling capabilities
package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// ProfilerConfig holds profiler configuration
type ProfilerConfig struct {
	Enabled        bool          `json:"enabled"`
	CPUProfile     bool          `json:"cpu_profile"`
	MemoryProfile  bool          `json:"memory_profile"`
	GoroutineTrace bool          `json:"goroutine_trace"`
	BlockProfile   bool          `json:"block_profile"`
	MutexProfile   bool          `json:"mutex_profile"`
	OutputDir      string        `json:"output_dir"`
	Duration       time.Duration `json:"duration"`
	Interval       time.Duration `json:"interval"`
	HTTPEndpoint   string        `json:"http_endpoint"`
}

// DefaultProfilerConfig returns a default profiler configuration
func DefaultProfilerConfig() *ProfilerConfig {
	return &ProfilerConfig{
		Enabled:        false,
		CPUProfile:     true,
		MemoryProfile:  true,
		GoroutineTrace: true,
		BlockProfile:   false,
		MutexProfile:   false,
		OutputDir:      "./debug-profiles",
		Duration:       30 * time.Second,
		Interval:       5 * time.Second,
		HTTPEndpoint:   ":6060",
	}
}

// Profiler provides comprehensive profiling capabilities
type Profiler struct {
	config    *ProfilerConfig
	active    bool
	mu        sync.RWMutex
	cancel    context.CancelFunc
	outputDir string
	startTime time.Time
	server    *http.Server
}

// NewProfiler creates a new profiler instance
func NewProfiler(config *ProfilerConfig) *Profiler {
	if config == nil {
		config = DefaultProfilerConfig()
	}

	return &Profiler{
		config:    config,
		active:    false,
		outputDir: config.OutputDir,
	}
}

// Start starts the profiler
func (p *Profiler) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return fmt.Errorf("profiler is already active")
	}

	if !p.config.Enabled {
		return fmt.Errorf("profiler is disabled")
	}

	// Create output directory
	if err := os.MkdirAll(p.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set up context for cancellation
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.startTime = time.Now()
	p.active = true

	log.Printf("🔍 Starting profiler with config: %+v", p.config)

	// Start HTTP endpoint for pprof
	if p.config.HTTPEndpoint != "" {
		go p.startHTTPEndpoint()
	}

	// Start profiling goroutine
	go p.profileLoop(ctx)

	return nil
}

// Stop stops the profiler
func (p *Profiler) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return fmt.Errorf("profiler is not active")
	}

	if p.cancel != nil {
		p.cancel()
	}

	if p.server != nil {
		if err := p.server.Shutdown(context.Background()); err != nil {
			log.Printf("⚠️ Error shutting down profiler HTTP server: %v", err)
		}
	}

	p.active = false
	duration := time.Since(p.startTime)
	log.Printf("🔍 Profiler stopped after %v", duration)

	return nil
}

// IsActive returns true if the profiler is currently active
func (p *Profiler) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

// GetStats returns profiler statistics
func (p *Profiler) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"active":     p.active,
		"config":     p.config,
		"output_dir": p.outputDir,
	}

	if p.active {
		stats["uptime"] = time.Since(p.startTime).String()
		stats["start_time"] = p.startTime
	}

	return stats
}

// profileLoop runs the main profiling loop
func (p *Profiler) profileLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	// Initial profile
	p.captureProfiles("initial")

	for {
		select {
		case <-ctx.Done():
			// Final profile before stopping
			p.captureProfiles("final")
			return
		case <-ticker.C:
			p.captureProfiles("periodic")
		}
	}
}

// captureProfiles captures various types of profiles
func (p *Profiler) captureProfiles(label string) {
	timestamp := time.Now().Format("20060102-150405")
	prefix := fmt.Sprintf("%s-%s", label, timestamp)

	log.Printf("📊 Capturing profiles: %s", prefix)

	// CPU Profile
	if p.config.CPUProfile {
		go p.captureCPUProfile(prefix)
	}

	// Memory Profile
	if p.config.MemoryProfile {
		go p.captureMemoryProfile(prefix)
	}

	// Goroutine Trace
	if p.config.GoroutineTrace {
		go p.captureGoroutineTrace(prefix)
	}

	// Block Profile
	if p.config.BlockProfile {
		go p.captureBlockProfile(prefix)
	}

	// Mutex Profile
	if p.config.MutexProfile {
		go p.captureMutexProfile(prefix)
	}

	// Runtime stats
	go p.captureRuntimeStats(prefix)
}

// captureCPUProfile captures CPU profiling data
func (p *Profiler) captureCPUProfile(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-cpu.prof", prefix))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Failed to create CPU profile: %v", err)
		return
	}
	defer file.Close()

	if err := pprof.StartCPUProfile(file); err != nil {
		log.Printf("❌ Failed to start CPU profile: %v", err)
		return
	}

	// Profile for a short duration
	time.Sleep(2 * time.Second)
	pprof.StopCPUProfile()

	log.Printf("💾 CPU profile saved: %s", filename)
}

// captureMemoryProfile captures memory profiling data
func (p *Profiler) captureMemoryProfile(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-mem.prof", prefix))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Failed to create memory profile: %v", err)
		return
	}
	defer file.Close()

	runtime.GC() // Force garbage collection for accurate stats
	if err := pprof.WriteHeapProfile(file); err != nil {
		log.Printf("❌ Failed to write memory profile: %v", err)
		return
	}

	log.Printf("💾 Memory profile saved: %s", filename)
}

// captureGoroutineTrace captures goroutine information
func (p *Profiler) captureGoroutineTrace(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-goroutine.prof", prefix))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Failed to create goroutine profile: %v", err)
		return
	}
	defer file.Close()

	if err := pprof.Lookup("goroutine").WriteTo(file, 0); err != nil {
		log.Printf("❌ Failed to write goroutine profile: %v", err)
		return
	}

	log.Printf("💾 Goroutine profile saved: %s", filename)
}

// captureBlockProfile captures blocking profiling data
func (p *Profiler) captureBlockProfile(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-block.prof", prefix))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Failed to create block profile: %v", err)
		return
	}
	defer file.Close()

	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	if err := pprof.Lookup("block").WriteTo(file, 0); err != nil {
		log.Printf("❌ Failed to write block profile: %v", err)
		return
	}

	log.Printf("💾 Block profile saved: %s", filename)
}

// captureMutexProfile captures mutex profiling data
func (p *Profiler) captureMutexProfile(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-mutex.prof", prefix))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Failed to create mutex profile: %v", err)
		return
	}
	defer file.Close()

	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	if err := pprof.Lookup("mutex").WriteTo(file, 0); err != nil {
		log.Printf("❌ Failed to write mutex profile: %v", err)
		return
	}

	log.Printf("💾 Mutex profile saved: %s", filename)
}

// captureRuntimeStats captures runtime statistics
func (p *Profiler) captureRuntimeStats(prefix string) {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("%s-runtime.json", prefix))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		"timestamp": time.Now(),
		"memory": map[string]interface{}{
			"alloc":          m.Alloc,
			"total_alloc":    m.TotalAlloc,
			"sys":            m.Sys,
			"lookups":        m.Lookups,
			"mallocs":        m.Mallocs,
			"frees":          m.Frees,
			"heap_alloc":     m.HeapAlloc,
			"heap_sys":       m.HeapSys,
			"heap_idle":      m.HeapIdle,
			"heap_inuse":     m.HeapInuse,
			"heap_released":  m.HeapReleased,
			"heap_objects":   m.HeapObjects,
			"stack_inuse":    m.StackInuse,
			"stack_sys":      m.StackSys,
			"next_gc":        m.NextGC,
			"last_gc":        m.LastGC,
			"pause_total_ns": m.PauseTotalNs,
			"num_gc":         m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_count":  runtime.NumCPU(),
		"go_version": runtime.Version(),
		"compiler":   runtime.Compiler,
		"goos":       runtime.GOOS,
		"goarch":     runtime.GOARCH,
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal runtime stats: %v", err)
		return
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		log.Printf("❌ Failed to write runtime stats: %v", err)
		return
	}

	log.Printf("💾 Runtime stats saved: %s", filename)
}

// startHTTPEndpoint starts the HTTP endpoint for pprof
func (p *Profiler) startHTTPEndpoint() {
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/trace", http.DefaultServeMux.ServeHTTP)

	// Custom health endpoint
	mux.HandleFunc("/debug/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := p.GetStats()
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	p.server = &http.Server{
		Addr:    p.config.HTTPEndpoint,
		Handler: mux,
	}

	log.Printf("🌐 Profiler HTTP endpoint starting on %s", p.config.HTTPEndpoint)
	log.Printf("📊 Profiles available at: http://localhost%s/debug/pprof/", p.config.HTTPEndpoint)

	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Profiler HTTP server error: %v", err)
	}
}

// GenerateReport generates a comprehensive profiling report
func (p *Profiler) GenerateReport() (string, error) {
	reportPath := filepath.Join(p.outputDir, fmt.Sprintf("profiling-report-%s.html",
		time.Now().Format("20060102-150405")))

	files, err := filepath.Glob(filepath.Join(p.outputDir, "*.prof"))
	if err != nil {
		return "", err
	}

	statsFiles, err := filepath.Glob(filepath.Join(p.outputDir, "*-runtime.json"))
	if err != nil {
		return "", err
	}

	// Generate HTML report
	report := p.generateHTMLReport(files, statsFiles)

	if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
		return "", err
	}

	log.Printf("📊 Profiling report generated: %s", reportPath)
	return reportPath, nil
}

// generateHTMLReport generates an HTML report
func (p *Profiler) generateHTMLReport(profileFiles, statsFiles []string) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>GZH Manager Profiling Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .file-list { background: #f9f9f9; padding: 15px; border-radius: 5px; }
        .file-item { margin: 5px 0; }
        pre { background: #f0f0f0; padding: 10px; border-radius: 3px; overflow-x: auto; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .stat-box { background: #e7f3ff; padding: 15px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>GZH Manager Profiling Report</h1>
        <p><strong>Generated:</strong> ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        <p><strong>Output Directory:</strong> ` + p.outputDir + `</p>
    </div>

    <div class="section">
        <h2>Profile Files</h2>
        <div class="file-list">`

	for _, file := range profileFiles {
		baseFile := filepath.Base(file)
		html += fmt.Sprintf(`
            <div class="file-item">
                <strong>%s</strong>
                <p>Use: <code>go tool pprof %s</code></p>
            </div>`, baseFile, file)
	}

	html += `
        </div>
    </div>

    <div class="section">
        <h2>Runtime Statistics</h2>
        <div class="stats">`

	for _, statsFile := range statsFiles {
		data, err := os.ReadFile(statsFile)
		if err != nil {
			continue
		}

		baseFile := filepath.Base(statsFile)
		html += fmt.Sprintf(`
            <div class="stat-box">
                <h3>%s</h3>
                <pre>%s</pre>
            </div>`, baseFile, string(data))
	}

	html += `
        </div>
    </div>

    <div class="section">
        <h2>Analysis Commands</h2>
        <div class="file-list">
            <h3>CPU Analysis</h3>
            <p><code>go tool pprof -http=:8080 *.cpu.prof</code></p>
            
            <h3>Memory Analysis</h3>
            <p><code>go tool pprof -http=:8081 *.mem.prof</code></p>
            
            <h3>Goroutine Analysis</h3>
            <p><code>go tool pprof -http=:8082 *.goroutine.prof</code></p>
            
            <h3>Compare Profiles</h3>
            <p><code>go tool pprof -http=:8083 -diff_base=initial-*.cpu.prof final-*.cpu.prof</code></p>
        </div>
    </div>
</body>
</html>`

	return html
}

// ProfileMemoryUsage returns current memory usage statistics
func ProfileMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"allocated_mb":   float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"heap_objects":   m.HeapObjects,
		"timestamp":      time.Now(),
	}
}

// FormatMemorySize formats bytes as human readable string
func FormatMemorySize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	Div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		Div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

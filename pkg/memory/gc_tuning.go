package memory

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// GCTuner manages garbage collection optimization and memory profiling
type GCTuner struct {
	config      GCConfig
	stats       GCStats
	mu          sync.RWMutex
	memoryPools map[string]*MemoryPool
	stopCh      chan struct{}
}

// GCConfig configures garbage collection tuning parameters
type GCConfig struct {
	// GC percentage controls how much new allocation triggers GC
	GCPercent int

	// Memory limit for the application (0 = unlimited)
	MemoryLimit int64

	// Enable memory profiling
	ProfilingEnabled bool

	// Profiling interval
	ProfilingInterval time.Duration

	// Force GC interval (0 = disabled)
	ForceGCInterval time.Duration

	// Enable object pooling
	ObjectPoolingEnabled bool

	// Memory pressure threshold (0.0-1.0)
	MemoryPressureThreshold float64
}

// DefaultGCConfig returns optimized GC configuration
func DefaultGCConfig() GCConfig {
	return GCConfig{
		GCPercent:               100, // Default Go setting
		MemoryLimit:             0,   // Unlimited by default
		ProfilingEnabled:        false,
		ProfilingInterval:       30 * time.Second,
		ForceGCInterval:         0, // Disabled by default
		ObjectPoolingEnabled:    true,
		MemoryPressureThreshold: 0.8,
	}
}

// GCStats tracks garbage collection statistics
type GCStats struct {
	// Basic GC stats
	NumGC        uint32
	TotalPauseNs uint64
	LastGC       time.Time

	// Memory stats
	HeapAlloc    uint64
	HeapSys      uint64
	HeapInuse    uint64
	HeapReleased uint64

	// Performance metrics
	AveragePauseNs uint64
	MaxPauseNs     uint64

	// Custom metrics
	PoolHits     int64
	PoolMisses   int64
	ForceGCCount int64

	// Memory pressure
	MemoryPressure float64
	LastUpdate     time.Time
}

// MemoryPool provides object pooling to reduce GC pressure
type MemoryPool struct {
	name      string
	pool      sync.Pool
	newFunc   func() interface{}
	resetFunc func(interface{})
	stats     PoolStats
	mu        sync.RWMutex
}

// PoolStats tracks memory pool statistics
type PoolStats struct {
	Gets    int64
	Puts    int64
	News    int64
	Resets  int64
	MaxSize int64
	HitRate float64
}

// NewGCTuner creates a new GC tuner with the given configuration
func NewGCTuner(config GCConfig) *GCTuner {
	tuner := &GCTuner{
		config:      config,
		memoryPools: make(map[string]*MemoryPool),
		stopCh:      make(chan struct{}),
	}

	// Apply GC configuration
	tuner.applyGCConfig()

	return tuner
}

// Start begins GC monitoring and optimization
func (t *GCTuner) Start(ctx context.Context) error {
	if t.config.ProfilingEnabled {
		go t.profileMemory(ctx)
	}

	if t.config.ForceGCInterval > 0 {
		go t.forceGCPeriodically(ctx)
	}

	// Start memory monitoring
	go t.monitorMemory(ctx)

	return nil
}

// Stop stops GC monitoring
func (t *GCTuner) Stop() {
	close(t.stopCh)
}

// applyGCConfig applies the GC configuration
func (t *GCTuner) applyGCConfig() {
	// Set GC percentage
	debug.SetGCPercent(t.config.GCPercent)

	// Set memory limit if specified
	if t.config.MemoryLimit > 0 {
		debug.SetMemoryLimit(t.config.MemoryLimit)
	}
}

// profileMemory periodically profiles memory usage
func (t *GCTuner) profileMemory(ctx context.Context) {
	ticker := time.NewTicker(t.config.ProfilingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.updateStats()
			t.analyzeMemoryPressure()
		}
	}
}

// forceGCPeriodically forces garbage collection at regular intervals
func (t *GCTuner) forceGCPeriodically(ctx context.Context) {
	ticker := time.NewTicker(t.config.ForceGCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.ForceGC()
		}
	}
}

// monitorMemory continuously monitors memory usage
func (t *GCTuner) monitorMemory(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.updateStats()

			// Check memory pressure and take action if needed
			if t.stats.MemoryPressure > t.config.MemoryPressureThreshold {
				t.handleMemoryPressure()
			}
		}
	}
}

// updateStats updates GC statistics
func (t *GCTuner) updateStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Update basic stats
	t.stats.NumGC = m.NumGC
	t.stats.TotalPauseNs = m.PauseTotalNs
	t.stats.HeapAlloc = m.HeapAlloc
	t.stats.HeapSys = m.HeapSys
	t.stats.HeapInuse = m.HeapInuse
	t.stats.HeapReleased = m.HeapReleased
	t.stats.LastUpdate = time.Now()

	if m.NumGC > 0 {
		t.stats.LastGC = time.Unix(0, int64(m.LastGC))
		t.stats.AveragePauseNs = m.PauseTotalNs / uint64(m.NumGC)

		// Find max pause time
		var maxPause uint64
		for _, pause := range m.PauseNs {
			if pause > maxPause {
				maxPause = pause
			}
		}
		t.stats.MaxPauseNs = maxPause
	}
}

// analyzeMemoryPressure calculates current memory pressure
func (t *GCTuner) analyzeMemoryPressure() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.config.MemoryLimit > 0 {
		// Calculate pressure based on memory limit
		t.stats.MemoryPressure = float64(t.stats.HeapInuse) / float64(t.config.MemoryLimit)
	} else {
		// Calculate pressure based on heap growth
		if t.stats.HeapSys > 0 {
			t.stats.MemoryPressure = float64(t.stats.HeapInuse) / float64(t.stats.HeapSys)
		}
	}
}

// handleMemoryPressure takes action when memory pressure is high
func (t *GCTuner) handleMemoryPressure() {
	// Force garbage collection
	t.ForceGC()

	// Clear memory pools if pressure is very high
	if t.stats.MemoryPressure > 0.9 {
		t.ClearAllPools()
	}
}

// ForceGC forces an immediate garbage collection
func (t *GCTuner) ForceGC() {
	runtime.GC()

	t.mu.Lock()
	t.stats.ForceGCCount++
	t.mu.Unlock()
}

// CreatePool creates a new memory pool for the given type
func (t *GCTuner) CreatePool(name string, newFunc func() interface{}, resetFunc func(interface{})) *MemoryPool {
	pool := &MemoryPool{
		name:      name,
		newFunc:   newFunc,
		resetFunc: resetFunc,
	}

	// Initialize sync.Pool with closure that references the pool correctly
	pool.pool = sync.Pool{
		New: func() interface{} {
			pool.mu.Lock()
			pool.stats.News++
			pool.mu.Unlock()
			return newFunc()
		},
	}

	t.mu.Lock()
	t.memoryPools[name] = pool
	t.mu.Unlock()

	return pool
}

// Get retrieves an object from the pool
func (p *MemoryPool) Get() interface{} {
	p.mu.Lock()
	p.stats.Gets++
	p.mu.Unlock()

	obj := p.pool.Get()

	p.mu.Lock()
	if p.stats.Gets > p.stats.News {
		// This was reused from pool
		hitRate := float64(p.stats.Gets-p.stats.News) / float64(p.stats.Gets)
		p.stats.HitRate = hitRate
	}
	p.mu.Unlock()

	return obj
}

// Put returns an object to the pool
func (p *MemoryPool) Put(obj interface{}) {
	if p.resetFunc != nil {
		p.resetFunc(obj)
		p.mu.Lock()
		p.stats.Resets++
		p.mu.Unlock()
	}

	p.mu.Lock()
	p.stats.Puts++
	p.mu.Unlock()

	p.pool.Put(obj)
}

// GetStats returns pool statistics
func (p *MemoryPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// ClearAllPools clears all memory pools
func (t *GCTuner) ClearAllPools() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, pool := range t.memoryPools {
		// Create new pool to clear the old one
		pool.pool = sync.Pool{
			New: pool.newFunc,
		}
	}
}

// GetStats returns current GC statistics
func (t *GCTuner) GetStats() GCStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// GetPoolStats returns statistics for all memory pools
func (t *GCTuner) GetPoolStats() map[string]PoolStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := make(map[string]PoolStats)
	for name, pool := range t.memoryPools {
		stats[name] = pool.GetStats()
	}
	return stats
}

// OptimizeForWorkload optimizes GC settings for a specific workload
func (t *GCTuner) OptimizeForWorkload(workloadType string) {
	switch workloadType {
	case "low-latency":
		// Optimize for low latency - more frequent but shorter GC pauses
		debug.SetGCPercent(50)
		t.config.GCPercent = 50

	case "high-throughput":
		// Optimize for throughput - less frequent GC
		debug.SetGCPercent(200)
		t.config.GCPercent = 200

	case "memory-constrained":
		// Optimize for memory usage - more aggressive GC
		debug.SetGCPercent(25)
		t.config.GCPercent = 25

	default:
		// Balanced approach
		debug.SetGCPercent(100)
		t.config.GCPercent = 100
	}
}

// PrintStats prints detailed GC and memory statistics
func (t *GCTuner) PrintStats() {
	stats := t.GetStats()
	poolStats := t.GetPoolStats()

	fmt.Printf("=== GC Statistics ===\n")
	fmt.Printf("Number of GC cycles: %d\n", stats.NumGC)
	fmt.Printf("Total pause time: %v\n", time.Duration(stats.TotalPauseNs))
	fmt.Printf("Average pause time: %v\n", time.Duration(stats.AveragePauseNs))
	fmt.Printf("Max pause time: %v\n", time.Duration(stats.MaxPauseNs))
	fmt.Printf("Last GC: %v\n", stats.LastGC)
	fmt.Printf("Heap allocated: %s\n", formatBytes(stats.HeapAlloc))
	fmt.Printf("Heap system: %s\n", formatBytes(stats.HeapSys))
	fmt.Printf("Heap in use: %s\n", formatBytes(stats.HeapInuse))
	fmt.Printf("Heap released: %s\n", formatBytes(stats.HeapReleased))
	fmt.Printf("Memory pressure: %.2f%%\n", stats.MemoryPressure*100)
	fmt.Printf("Force GC count: %d\n", stats.ForceGCCount)

	fmt.Printf("\n=== Pool Statistics ===\n")
	for name, pStats := range poolStats {
		fmt.Printf("Pool '%s':\n", name)
		fmt.Printf("  Gets: %d, Puts: %d, News: %d\n", pStats.Gets, pStats.Puts, pStats.News)
		fmt.Printf("  Hit rate: %.2f%%\n", pStats.HitRate*100)
		fmt.Printf("  Resets: %d\n", pStats.Resets)
	}
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

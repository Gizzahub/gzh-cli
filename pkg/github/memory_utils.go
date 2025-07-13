package github

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// parseMemorySize parses memory size strings (e.g., "500MB", "2GB") into bytes
func parseMemorySize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty memory size string")
	}

	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			numStr += string(char)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid memory size format: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value in memory size: %s", numStr)
	}

	switch unit {
	case "B", "":
		return int64(value), nil
	case "KB", "K":
		return int64(value * 1024), nil
	case "MB", "M":
		return int64(value * 1024 * 1024), nil
	case "GB", "G":
		return int64(value * 1024 * 1024 * 1024), nil
	case "TB", "T":
		return int64(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unsupported memory unit: %s", unit)
	}
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc         uint64      // Currently allocated bytes
	TotalAlloc    uint64      // Total allocated bytes (cumulative)
	Sys           uint64      // System bytes obtained from OS
	Lookups       uint64      // Number of pointer lookups
	Mallocs       uint64      // Number of allocations
	Frees         uint64      // Number of frees
	HeapAlloc     uint64      // Heap allocated bytes
	HeapSys       uint64      // Heap system bytes
	HeapIdle      uint64      // Heap idle bytes
	HeapInuse     uint64      // Heap in-use bytes
	HeapReleased  uint64      // Heap released bytes
	HeapObjects   uint64      // Number of heap objects
	StackInuse    uint64      // Stack in-use bytes
	StackSys      uint64      // Stack system bytes
	MSpanInuse    uint64      // MSpan in-use bytes
	MSpanSys      uint64      // MSpan system bytes
	MCacheInuse   uint64      // MCache in-use bytes
	MCacheSys     uint64      // MCache system bytes
	BuckHashSys   uint64      // Bucket hash system bytes
	GCSys         uint64      // GC system bytes
	OtherSys      uint64      // Other system bytes
	NextGC        uint64      // Next GC threshold
	LastGC        uint64      // Last GC time (nanoseconds since epoch)
	PauseTotalNs  uint64      // Total pause time in nanoseconds
	PauseNs       [256]uint64 // Last 256 GC pause times
	PauseEnd      [256]uint64 // Last 256 GC pause end times
	NumGC         uint32      // Number of GC cycles
	NumForcedGC   uint32      // Number of forced GC cycles
	GCCPUFraction float64     // Fraction of CPU time used by GC
	EnableGC      bool        // GC enabled flag
	DebugGC       bool        // Debug GC flag
	Timestamp     time.Time   // When stats were collected
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		PauseNs:       m.PauseNs,
		PauseEnd:      m.PauseEnd,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		EnableGC:      m.EnableGC,
		DebugGC:       m.DebugGC,
		Timestamp:     time.Now(),
	}
}

// String returns a human-readable string representation of memory stats
func (ms *MemoryStats) String() string {
	return fmt.Sprintf(
		"Memory Stats:\n"+
			"  Allocated: %s\n"+
			"  System: %s\n"+
			"  Heap: %s (in-use: %s, idle: %s)\n"+
			"  GC: %d cycles, %.2f%% CPU, last: %v ago\n"+
			"  Objects: %d (allocs: %d, frees: %d)",
		formatBytes(int64(ms.Alloc)),
		formatBytes(int64(ms.Sys)),
		formatBytes(int64(ms.HeapSys)),
		formatBytes(int64(ms.HeapInuse)),
		formatBytes(int64(ms.HeapIdle)),
		ms.NumGC,
		ms.GCCPUFraction*100,
		time.Since(time.Unix(0, int64(ms.LastGC))),
		ms.HeapObjects,
		ms.Mallocs,
		ms.Frees,
	)
}

// MemoryEfficiency calculates memory usage efficiency metrics
func (ms *MemoryStats) MemoryEfficiency() map[string]float64 {
	efficiency := make(map[string]float64)

	// Heap utilization (allocated / system)
	if ms.HeapSys > 0 {
		efficiency["heap_utilization"] = float64(ms.HeapInuse) / float64(ms.HeapSys)
	}

	// Overall memory efficiency (heap allocated / total system)
	if ms.Sys > 0 {
		efficiency["memory_efficiency"] = float64(ms.HeapAlloc) / float64(ms.Sys)
	}

	// GC pressure (GC time / total time)
	efficiency["gc_pressure"] = ms.GCCPUFraction

	// Fragmentation (heap idle / heap system)
	if ms.HeapSys > 0 {
		efficiency["fragmentation"] = float64(ms.HeapIdle) / float64(ms.HeapSys)
	}

	// Allocation efficiency (active objects / total allocations)
	if ms.Mallocs > 0 {
		efficiency["allocation_efficiency"] = float64(ms.HeapObjects) / float64(ms.Mallocs)
	}

	return efficiency
}

// OptimizeMemoryUsage performs aggressive memory optimization
func OptimizeMemoryUsage() *MemoryStats {
	beforeStats := GetMemoryStats()

	// Force multiple GC cycles to ensure complete cleanup
	for i := 0; i < 3; i++ {
		runtime.GC()
		runtime.GC() // Double GC to ensure sweep phase completes
	}

	// Force finalization of objects with finalizers
	runtime.Gosched()

	afterStats := GetMemoryStats()

	// Return the difference
	return &MemoryStats{
		Alloc:     beforeStats.Alloc - afterStats.Alloc,
		HeapAlloc: beforeStats.HeapAlloc - afterStats.HeapAlloc,
		HeapInuse: beforeStats.HeapInuse - afterStats.HeapInuse,
		NumGC:     afterStats.NumGC - beforeStats.NumGC,
		Timestamp: afterStats.Timestamp,
	}
}

// MemoryPressureLevel represents the current memory pressure
type MemoryPressureLevel int

const (
	MemoryPressureLow MemoryPressureLevel = iota
	MemoryPressureMedium
	MemoryPressureHigh
	MemoryPressureCritical
)

// String returns string representation of memory pressure level
func (mpl MemoryPressureLevel) String() string {
	switch mpl {
	case MemoryPressureLow:
		return "Low"
	case MemoryPressureMedium:
		return "Medium"
	case MemoryPressureHigh:
		return "High"
	case MemoryPressureCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// GetMemoryPressure assesses current memory pressure level
func GetMemoryPressure(maxMemory int64) MemoryPressureLevel {
	stats := GetMemoryStats()

	if maxMemory <= 0 {
		// Use system memory as baseline
		maxMemory = int64(stats.Sys)
	}

	currentUsage := int64(stats.Alloc)
	usagePercent := float64(currentUsage) / float64(maxMemory)

	switch {
	case usagePercent < 0.5:
		return MemoryPressureLow
	case usagePercent < 0.7:
		return MemoryPressureMedium
	case usagePercent < 0.9:
		return MemoryPressureHigh
	default:
		return MemoryPressureCritical
	}
}

// MemoryWatcher monitors memory usage and triggers cleanup when needed
type MemoryWatcher struct {
	maxMemory     int64
	threshold     float64
	checkInterval time.Duration
	onPressure    func(MemoryPressureLevel)
	stopChan      chan struct{}
}

// NewMemoryWatcher creates a new memory watcher
func NewMemoryWatcher(maxMemory int64, threshold float64, checkInterval time.Duration) *MemoryWatcher {
	return &MemoryWatcher{
		maxMemory:     maxMemory,
		threshold:     threshold,
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}
}

// SetPressureHandler sets the callback function for memory pressure events
func (mw *MemoryWatcher) SetPressureHandler(handler func(MemoryPressureLevel)) {
	mw.onPressure = handler
}

// Start begins memory monitoring
func (mw *MemoryWatcher) Start() {
	go func() {
		ticker := time.NewTicker(mw.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pressure := GetMemoryPressure(mw.maxMemory)

				if pressure >= MemoryPressureHigh && mw.onPressure != nil {
					mw.onPressure(pressure)
				}

			case <-mw.stopChan:
				return
			}
		}
	}()
}

// Stop stops memory monitoring
func (mw *MemoryWatcher) Stop() {
	close(mw.stopChan)
}

package reports

import (
	"sync"
	"time"
)

// BandwidthCalculator tracks and calculates network bandwidth rates
type BandwidthCalculator struct {
	previousMeasurements map[string]*InterfaceSnapshot
	measurementInterval  time.Duration
	mutex                sync.RWMutex
	history              map[string][]BandwidthMeasurement
	maxHistorySize       int
}

// InterfaceSnapshot represents a point-in-time measurement of interface metrics
type InterfaceSnapshot struct {
	InterfaceName string
	Timestamp     time.Time
	RxBytes       uint64
	TxBytes       uint64
	RxPackets     uint64
	TxPackets     uint64
	RxDropped     uint64
	TxDropped     uint64
	RxErrors      uint64
	TxErrors      uint64
}

// BandwidthMeasurement represents calculated bandwidth rates
type BandwidthMeasurement struct {
	InterfaceName   string        `json:"interface"`
	Timestamp       time.Time     `json:"timestamp"`
	RxBytesPerSec   uint64        `json:"rx_bytes_per_sec"`
	TxBytesPerSec   uint64        `json:"tx_bytes_per_sec"`
	RxPacketsPerSec uint64        `json:"rx_packets_per_sec"`
	TxPacketsPerSec uint64        `json:"tx_packets_per_sec"`
	Utilization     float64       `json:"utilization_percent"`
	Duration        time.Duration `json:"duration"`
}

// BandwidthTrend represents historical bandwidth usage over time
type BandwidthTrend struct {
	InterfaceName string                 `json:"interface"`
	Measurements  []BandwidthMeasurement `json:"measurements"`
	AverageRxRate uint64                 `json:"avg_rx_rate_bps"`
	AverageTxRate uint64                 `json:"avg_tx_rate_bps"`
	PeakRxRate    uint64                 `json:"peak_rx_rate_bps"`
	PeakTxRate    uint64                 `json:"peak_tx_rate_bps"`
	AverageUtil   float64                `json:"avg_utilization_percent"`
	PeakUtil      float64                `json:"peak_utilization_percent"`
}

// NewBandwidthCalculator creates a new bandwidth calculator
func NewBandwidthCalculator(measurementInterval time.Duration) *BandwidthCalculator {
	return &BandwidthCalculator{
		previousMeasurements: make(map[string]*InterfaceSnapshot),
		measurementInterval:  measurementInterval,
		history:              make(map[string][]BandwidthMeasurement),
		maxHistorySize:       100, // Keep last 100 measurements per interface
	}
}

// AddSnapshot adds a new interface measurement and calculates rates
func (bc *BandwidthCalculator) AddSnapshot(snapshot *InterfaceSnapshot) *BandwidthMeasurement {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	interfaceName := snapshot.InterfaceName

	// Check if we have a previous measurement for this interface
	previous, exists := bc.previousMeasurements[interfaceName]
	if !exists {
		// First measurement, store and return nil
		bc.previousMeasurements[interfaceName] = snapshot
		return nil
	}

	// Calculate time difference
	timeDiff := snapshot.Timestamp.Sub(previous.Timestamp)
	if timeDiff <= 0 {
		// Invalid time difference, skip calculation
		return nil
	}

	seconds := timeDiff.Seconds()

	// Calculate rates (bytes and packets per second)
	measurement := &BandwidthMeasurement{
		InterfaceName:   interfaceName,
		Timestamp:       snapshot.Timestamp,
		Duration:        timeDiff,
		RxBytesPerSec:   bc.calculateRate(snapshot.RxBytes, previous.RxBytes, seconds),
		TxBytesPerSec:   bc.calculateRate(snapshot.TxBytes, previous.TxBytes, seconds),
		RxPacketsPerSec: bc.calculateRate(snapshot.RxPackets, previous.RxPackets, seconds),
		TxPacketsPerSec: bc.calculateRate(snapshot.TxPackets, previous.TxPackets, seconds),
	}

	// Store current snapshot for next calculation
	bc.previousMeasurements[interfaceName] = snapshot

	// Add to history
	bc.addToHistory(interfaceName, *measurement)

	return measurement
}

// calculateRate calculates rate per second, handling counter wraps
func (bc *BandwidthCalculator) calculateRate(current, previous uint64, seconds float64) uint64 {
	if current < previous {
		// Counter wrapped around, skip this calculation
		return 0
	}

	diff := current - previous
	return uint64(float64(diff) / seconds)
}

// addToHistory adds a measurement to the interface history
func (bc *BandwidthCalculator) addToHistory(interfaceName string, measurement BandwidthMeasurement) {
	history := bc.history[interfaceName]
	history = append(history, measurement)

	// Trim history if it exceeds max size
	if len(history) > bc.maxHistorySize {
		history = history[len(history)-bc.maxHistorySize:]
	}

	bc.history[interfaceName] = history
}

// GetBandwidthTrends returns bandwidth trends for all interfaces
func (bc *BandwidthCalculator) GetBandwidthTrends() map[string]BandwidthTrend {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	trends := make(map[string]BandwidthTrend)

	for interfaceName, measurements := range bc.history {
		if len(measurements) == 0 {
			continue
		}

		trend := BandwidthTrend{
			InterfaceName: interfaceName,
			Measurements:  make([]BandwidthMeasurement, len(measurements)),
		}

		// Copy measurements
		copy(trend.Measurements, measurements)

		// Calculate statistics
		var totalRx, totalTx uint64
		var totalUtil float64
		var peakRx, peakTx, peakUtil uint64 = 0, 0, 0

		for _, m := range measurements {
			totalRx += m.RxBytesPerSec
			totalTx += m.TxBytesPerSec
			totalUtil += m.Utilization

			if m.RxBytesPerSec > peakRx {
				peakRx = m.RxBytesPerSec
			}
			if m.TxBytesPerSec > peakTx {
				peakTx = m.TxBytesPerSec
			}
			if m.Utilization > float64(peakUtil) {
				peakUtil = uint64(m.Utilization)
			}
		}

		count := len(measurements)
		trend.AverageRxRate = totalRx / uint64(count)
		trend.AverageTxRate = totalTx / uint64(count)
		trend.AverageUtil = totalUtil / float64(count)
		trend.PeakRxRate = peakRx
		trend.PeakTxRate = peakTx
		trend.PeakUtil = float64(peakUtil)

		trends[interfaceName] = trend
	}

	return trends
}

// GetRecentMeasurements returns the most recent measurements for all interfaces
func (bc *BandwidthCalculator) GetRecentMeasurements() map[string]BandwidthMeasurement {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	recent := make(map[string]BandwidthMeasurement)

	for interfaceName, measurements := range bc.history {
		if len(measurements) > 0 {
			recent[interfaceName] = measurements[len(measurements)-1]
		}
	}

	return recent
}

// Reset clears all historical data
func (bc *BandwidthCalculator) Reset() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.previousMeasurements = make(map[string]*InterfaceSnapshot)
	bc.history = make(map[string][]BandwidthMeasurement)
}

// SetUtilization updates the utilization for a specific measurement
func (bc *BandwidthCalculator) SetUtilization(interfaceName string, utilization float64) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if history, exists := bc.history[interfaceName]; exists && len(history) > 0 {
		// Update the most recent measurement
		history[len(history)-1].Utilization = utilization
		bc.history[interfaceName] = history
	}
}

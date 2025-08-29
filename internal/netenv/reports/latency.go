package reports

import (
	"context"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LatencyTester performs network latency measurements
type LatencyTester struct {
	timeout    time.Duration
	pingCount  int
	concurrent int
	mutex      sync.RWMutex
}

// NewLatencyTester creates a new latency tester
func NewLatencyTester() *LatencyTester {
	return &LatencyTester{
		timeout:    5 * time.Second,
		pingCount:  5,
		concurrent: 4,
	}
}

// RunLatencyTests performs latency tests against multiple targets
func (lt *LatencyTester) RunLatencyTests(targets []string) (*LatencyReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), lt.timeout*time.Duration(len(targets)))
	defer cancel()

	results := make(chan LatencyTarget, len(targets))
	sem := make(chan struct{}, lt.concurrent)

	var wg sync.WaitGroup

	// Launch ping tests concurrently
	for _, target := range targets {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result := lt.pingTarget(ctx, host)
			results <- result
		}(target)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	latencyTargets := make([]LatencyTarget, 0, len(lt.targets))
	for result := range results {
		latencyTargets = append(latencyTargets, result)
	}

	// Calculate aggregate statistics
	report := lt.calculateLatencyStatistics(latencyTargets)

	return report, nil
}

// pingTarget performs ping test against a single target
func (lt *LatencyTester) pingTarget(ctx context.Context, host string) LatencyTarget {
	target := LatencyTarget{
		Host:        host,
		LastChecked: time.Now(),
		Reachable:   false,
	}

	// Resolve hostname to IP
	if ip, err := net.ResolveIPAddr("ip", host); err == nil {
		target.IP = ip.IP.String()
	} else {
		target.IP = host // Assume it's already an IP
	}

	// Perform ping based on OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "ping", "-n", strconv.Itoa(lt.pingCount), host)
	case "darwin", "linux":
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(lt.pingCount), host)
	default:
		// Fallback for other Unix-like systems
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(lt.pingCount), host)
	}

	output, err := cmd.Output()
	if err != nil {
		return target
	}

	// Parse ping results
	return lt.parsePingOutput(target, string(output))
}

// parsePingOutput extracts latency and packet loss from ping command output
func (lt *LatencyTester) parsePingOutput(target LatencyTarget, output string) LatencyTarget {
	lines := strings.Split(output, "\n")

	// Platform-specific parsing
	switch runtime.GOOS {
	case "windows":
		target = lt.parseWindowsPing(target, lines)
	default:
		target = lt.parseUnixPing(target, lines)
	}

	return target
}

// parseUnixPing parses Unix/Linux/macOS ping output
func (lt *LatencyTester) parseUnixPing(target LatencyTarget, lines []string) LatencyTarget {
	var latencies []float64
	transmitted := 0
	received := 0

	// Regex patterns for Unix ping
	timeRegex := regexp.MustCompile(`time[<=]([0-9]+\.?[0-9]*)\s*ms`)
	statsRegex := regexp.MustCompile(`(\d+) packets transmitted, (\d+) (?:packets )?received`)

	for _, line := range lines {
		// Extract individual ping times
		if matches := timeRegex.FindStringSubmatch(line); len(matches) > 1 {
			if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
				latencies = append(latencies, latency)
			}
		}

		// Extract packet statistics
		if matches := statsRegex.FindStringSubmatch(line); len(matches) > 2 {
			transmitted, _ = strconv.Atoi(matches[1])
			received, _ = strconv.Atoi(matches[2])
		}
	}

	if len(latencies) > 0 {
		target.Reachable = true
		sum := 0.0
		for _, lat := range latencies {
			sum += lat
		}
		target.LatencyMs = sum / float64(len(latencies))
	}

	if transmitted > 0 {
		lost := transmitted - received
		target.PacketLoss = (float64(lost) / float64(transmitted)) * 100
	}

	return target
}

// parseWindowsPing parses Windows ping output
func (lt *LatencyTester) parseWindowsPing(target LatencyTarget, lines []string) LatencyTarget {
	var latencies []float64
	transmitted := 0
	received := 0

	// Regex patterns for Windows ping
	timeRegex := regexp.MustCompile(`time[<=]([0-9]+)ms`)
	statsRegex := regexp.MustCompile(`Packets: Sent = (\d+), Received = (\d+), Lost = (\d+)`)

	for _, line := range lines {
		// Extract individual ping times
		if matches := timeRegex.FindStringSubmatch(line); len(matches) > 1 {
			if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
				latencies = append(latencies, latency)
			}
		}

		// Extract packet statistics
		if matches := statsRegex.FindStringSubmatch(line); len(matches) > 3 {
			transmitted, _ = strconv.Atoi(matches[1])
			received, _ = strconv.Atoi(matches[2])
		}
	}

	if len(latencies) > 0 {
		target.Reachable = true
		sum := 0.0
		for _, lat := range latencies {
			sum += lat
		}
		target.LatencyMs = sum / float64(len(latencies))
	}

	if transmitted > 0 {
		lost := transmitted - received
		target.PacketLoss = (float64(lost) / float64(transmitted)) * 100
	}

	return target
}

// calculateLatencyStatistics computes aggregate statistics from individual target results
func (lt *LatencyTester) calculateLatencyStatistics(targets []LatencyTarget) *LatencyReport {
	report := &LatencyReport{
		Targets: targets,
	}

	if len(targets) == 0 {
		return report
	}

	var totalLatency, totalPacketLoss float64
	var latencies []float64
	reachableCount := 0

	for _, target := range targets {
		if target.Reachable {
			reachableCount++
			totalLatency += target.LatencyMs
			latencies = append(latencies, target.LatencyMs)
		}
		totalPacketLoss += target.PacketLoss
	}

	if reachableCount > 0 {
		report.AverageLatency = totalLatency / float64(reachableCount)

		// Calculate min/max latency
		report.MinLatency = latencies[0]
		report.MaxLatency = latencies[0]
		for _, lat := range latencies {
			if lat < report.MinLatency {
				report.MinLatency = lat
			}
			if lat > report.MaxLatency {
				report.MaxLatency = lat
			}
		}

		// Calculate jitter (standard deviation)
		if len(latencies) > 1 {
			variance := 0.0
			for _, lat := range latencies {
				diff := lat - report.AverageLatency
				variance += diff * diff
			}
			variance /= float64(len(latencies) - 1)
			report.JitterAverage = variance // Simplified jitter calculation
		}
	}

	report.PacketLoss = totalPacketLoss / float64(len(targets))
	report.ReachabilityRate = (float64(reachableCount) / float64(len(targets))) * 100

	return report
}

// SetTimeout configures the timeout for ping operations
func (lt *LatencyTester) SetTimeout(timeout time.Duration) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()
	lt.timeout = timeout
}

// SetPingCount configures the number of pings per target
func (lt *LatencyTester) SetPingCount(count int) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()
	lt.pingCount = count
}

// SetConcurrency configures the number of concurrent ping operations
func (lt *LatencyTester) SetConcurrency(concurrent int) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()
	lt.concurrent = concurrent
}

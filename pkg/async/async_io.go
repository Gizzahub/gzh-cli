package async

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// IOResult represents the result of an async I/O operation
type IOResult struct {
	Data  []byte
	Error error
	Path  string
	Meta  map[string]interface{}
}

// HTTPResult represents the result of an async HTTP request
type HTTPResult struct {
	Response *http.Response
	Body     []byte
	Error    error
	URL      string
	Duration time.Duration
}

// AsyncIO provides non-blocking I/O operations
type AsyncIO struct {
	maxConcurrency int
	httpClient     *http.Client
	semaphore      chan struct{}
	stats          IOStats
	mu             sync.RWMutex
}

// IOStats tracks async I/O performance metrics
type IOStats struct {
	TotalOperations int64
	CompletedOps    int64
	FailedOps       int64
	AverageLatency  time.Duration
	MaxConcurrent   int
	ActiveOps       int
}

// NewAsyncIO creates a new async I/O manager
func NewAsyncIO(maxConcurrency int) *AsyncIO {
	return &AsyncIO{
		maxConcurrency: maxConcurrency,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

// ReadFileAsync reads a file asynchronously
func (aio *AsyncIO) ReadFileAsync(ctx context.Context, path string) <-chan IOResult {
	result := make(chan IOResult, 1)

	go func() {
		defer close(result)

		// Acquire semaphore
		select {
		case aio.semaphore <- struct{}{}:
			defer func() { <-aio.semaphore }()
		case <-ctx.Done():
			result <- IOResult{Error: ctx.Err(), Path: path}
			return
		}

		aio.trackOperation(true)
		defer aio.trackOperation(false)

		start := time.Now()
		data, err := os.ReadFile(path)
		duration := time.Since(start)

		aio.updateStats(duration, err == nil)

		result <- IOResult{
			Data:  data,
			Error: err,
			Path:  path,
			Meta:  map[string]interface{}{"duration": duration},
		}
	}()

	return result
}

// WriteFileAsync writes a file asynchronously
func (aio *AsyncIO) WriteFileAsync(ctx context.Context, path string, data []byte, perm os.FileMode) <-chan IOResult {
	result := make(chan IOResult, 1)

	go func() {
		defer close(result)

		// Acquire semaphore
		select {
		case aio.semaphore <- struct{}{}:
			defer func() { <-aio.semaphore }()
		case <-ctx.Done():
			result <- IOResult{Error: ctx.Err(), Path: path}
			return
		}

		aio.trackOperation(true)
		defer aio.trackOperation(false)

		start := time.Now()

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			result <- IOResult{Error: err, Path: path}
			return
		}

		err := os.WriteFile(path, data, perm)
		duration := time.Since(start)

		aio.updateStats(duration, err == nil)

		result <- IOResult{
			Error: err,
			Path:  path,
			Meta:  map[string]interface{}{"duration": duration, "size": len(data)},
		}
	}()

	return result
}

// HTTPRequestAsync performs an HTTP request asynchronously
func (aio *AsyncIO) HTTPRequestAsync(ctx context.Context, req *http.Request) <-chan HTTPResult {
	result := make(chan HTTPResult, 1)

	go func() {
		defer close(result)

		// Acquire semaphore
		select {
		case aio.semaphore <- struct{}{}:
			defer func() { <-aio.semaphore }()
		case <-ctx.Done():
			result <- HTTPResult{Error: ctx.Err(), URL: req.URL.String()}
			return
		}

		aio.trackOperation(true)
		defer aio.trackOperation(false)

		start := time.Now()
		resp, err := aio.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			duration := time.Since(start)
			aio.updateStats(duration, false)
			result <- HTTPResult{Error: err, URL: req.URL.String(), Duration: duration}
			return
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		duration := time.Since(start)

		aio.updateStats(duration, err == nil && resp.StatusCode < 400)

		result <- HTTPResult{
			Response: resp,
			Body:     body,
			Error:    err,
			URL:      req.URL.String(),
			Duration: duration,
		}
	}()

	return result
}

// WalkDirAsync walks a directory tree asynchronously
func (aio *AsyncIO) WalkDirAsync(ctx context.Context, root string, walkFn func(path string, info os.FileInfo) error) <-chan error {
	result := make(chan error, 1)

	go func() {
		defer close(result)

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			return walkFn(path, info)
		})

		result <- err
	}()

	return result
}

// BatchReadFiles reads multiple files concurrently
func (aio *AsyncIO) BatchReadFiles(ctx context.Context, paths []string) <-chan IOResult {
	results := make(chan IOResult, len(paths))

	go func() {
		defer close(results)

		var wg sync.WaitGroup
		for _, path := range paths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()

				resultCh := aio.ReadFileAsync(ctx, p)
				select {
				case result := <-resultCh:
					results <- result
				case <-ctx.Done():
					results <- IOResult{Error: ctx.Err(), Path: p}
				}
			}(path)
		}

		wg.Wait()
	}()

	return results
}

// BatchHTTPRequests performs multiple HTTP requests concurrently
func (aio *AsyncIO) BatchHTTPRequests(ctx context.Context, requests []*http.Request) <-chan HTTPResult {
	results := make(chan HTTPResult, len(requests))

	go func() {
		defer close(results)

		var wg sync.WaitGroup
		for _, req := range requests {
			wg.Add(1)
			go func(r *http.Request) {
				defer wg.Done()

				resultCh := aio.HTTPRequestAsync(ctx, r)
				select {
				case result := <-resultCh:
					results <- result
				case <-ctx.Done():
					results <- HTTPResult{Error: ctx.Err(), URL: r.URL.String()}
				}
			}(req)
		}

		wg.Wait()
	}()

	return results
}

// trackOperation updates active operation count
func (aio *AsyncIO) trackOperation(start bool) {
	aio.mu.Lock()
	defer aio.mu.Unlock()

	if start {
		aio.stats.TotalOperations++
		aio.stats.ActiveOps++
		if aio.stats.ActiveOps > aio.stats.MaxConcurrent {
			aio.stats.MaxConcurrent = aio.stats.ActiveOps
		}
	} else {
		aio.stats.ActiveOps--
	}
}

// updateStats updates performance statistics
func (aio *AsyncIO) updateStats(duration time.Duration, success bool) {
	aio.mu.Lock()
	defer aio.mu.Unlock()

	if success {
		aio.stats.CompletedOps++
	} else {
		aio.stats.FailedOps++
	}

	// Update average latency using exponential moving average
	if aio.stats.AverageLatency == 0 {
		aio.stats.AverageLatency = duration
	} else {
		alpha := 0.1
		aio.stats.AverageLatency = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(aio.stats.AverageLatency),
		)
	}
}

// GetStats returns current I/O statistics
func (aio *AsyncIO) GetStats() IOStats {
	aio.mu.RLock()
	defer aio.mu.RUnlock()
	return aio.stats
}

// PrintStats prints detailed I/O statistics
func (aio *AsyncIO) PrintStats() {
	stats := aio.GetStats()

	fmt.Printf("=== Async I/O Statistics ===\n")
	fmt.Printf("Total Operations: %d\n", stats.TotalOperations)
	fmt.Printf("Completed: %d\n", stats.CompletedOps)
	fmt.Printf("Failed: %d\n", stats.FailedOps)
	fmt.Printf("Active Operations: %d\n", stats.ActiveOps)
	fmt.Printf("Max Concurrent: %d\n", stats.MaxConcurrent)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

	if stats.TotalOperations > 0 {
		successRate := float64(stats.CompletedOps) / float64(stats.TotalOperations) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
	}
}

// Close gracefully shuts down the async I/O manager
func (aio *AsyncIO) Close() error {
	// Wait for all operations to complete or timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for operations to complete")
		case <-ticker.C:
			stats := aio.GetStats()
			if stats.ActiveOps == 0 {
				return nil
			}
		}
	}
}

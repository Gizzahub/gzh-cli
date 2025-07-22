// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchProcessor provides efficient batch processing for API requests.
type BatchProcessor struct {
	maxBatchSize  int
	flushInterval time.Duration
	concurrency   int
	batches       map[string]*pendingBatch
	mu            sync.RWMutex
	stats         BatchStats
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// BatchStats tracks batch processing performance metrics.
type BatchStats struct {
	TotalRequests    int64
	TotalBatches     int64
	AverageBatchSize float64
	TotalSavings     time.Duration
	LastProcessed    time.Time
}

// BatchRequest represents a single request in a batch.
type BatchRequest struct {
	ID       string
	Data     interface{}
	Response chan BatchResponse
}

// BatchResponse contains the result of a batch request.
type BatchResponse struct {
	ID    string
	Data  interface{}
	Error error
}

// BatchProcessor handles batching of similar requests.
type pendingBatch struct {
	requests   []*BatchRequest
	timer      *time.Timer
	processor  BatchFunc
	mu         sync.Mutex
	lastUpdate time.Time
}

// BatchFunc processes a batch of requests and returns responses.
type BatchFunc func(ctx context.Context, requests []*BatchRequest) []BatchResponse

// BatchConfig configures batch processing behavior.
type BatchConfig struct {
	MaxBatchSize  int           // Maximum number of requests per batch
	FlushInterval time.Duration // Maximum time to wait before processing batch
	Concurrency   int           // Number of concurrent batch processors
}

// DefaultBatchConfig returns reasonable default configuration.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxBatchSize:  50,
		FlushInterval: 100 * time.Millisecond,
		Concurrency:   5,
	}
}

// NewBatchProcessor creates a new batch processor with the given configuration.
func NewBatchProcessor(config BatchConfig) *BatchProcessor {
	bp := &BatchProcessor{
		maxBatchSize:  config.MaxBatchSize,
		flushInterval: config.FlushInterval,
		concurrency:   config.Concurrency,
		batches:       make(map[string]*pendingBatch),
		stopCh:        make(chan struct{}),
	}

	// Start batch processing workers
	for i := 0; i < config.Concurrency; i++ {
		bp.wg.Add(1)

		go bp.processBatches()
	}

	return bp
}

// Add adds a request to the appropriate batch.
func (bp *BatchProcessor) Add(_ context.Context, batchKey string, request *BatchRequest, processor BatchFunc) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Update stats
	bp.stats.TotalRequests++
	bp.stats.LastProcessed = time.Now()

	// Get or create batch for this key
	batch, exists := bp.batches[batchKey]
	if !exists {
		batch = &pendingBatch{
			requests:   make([]*BatchRequest, 0, bp.maxBatchSize),
			processor:  processor,
			lastUpdate: time.Now(),
		}
		bp.batches[batchKey] = batch

		// Set timer to flush batch after interval
		batch.timer = time.AfterFunc(bp.flushInterval, func() { //nolint:contextcheck // Timer callback doesn't need context
			bp.flushBatch(batchKey)
		})
	}

	batch.mu.Lock()
	batch.requests = append(batch.requests, request)
	batch.lastUpdate = time.Now()
	shouldFlush := len(batch.requests) >= bp.maxBatchSize
	batch.mu.Unlock()

	// Flush immediately if batch is full
	if shouldFlush {
		if batch.timer != nil {
			batch.timer.Stop()
		}
		// Use a goroutine to avoid deadlock since flushBatch also tries to acquire the lock
		go bp.flushBatch(batchKey) //nolint:contextcheck // Goroutine flushing doesn't need context
	}

	return nil
}

// flushBatch processes a batch of requests.
func (bp *BatchProcessor) flushBatch(batchKey string) {
	bp.mu.Lock()

	batch, exists := bp.batches[batchKey]
	if !exists {
		bp.mu.Unlock()
		return
	}

	delete(bp.batches, batchKey)
	bp.mu.Unlock()

	batch.mu.Lock()
	requests := make([]*BatchRequest, len(batch.requests))
	copy(requests, batch.requests)
	processor := batch.processor
	batch.mu.Unlock()

	if len(requests) == 0 {
		return
	}

	// Update batch stats
	bp.mu.Lock()

	bp.stats.TotalBatches++
	if bp.stats.TotalBatches > 0 {
		bp.stats.AverageBatchSize = float64(bp.stats.TotalRequests) / float64(bp.stats.TotalBatches)
	}

	bp.mu.Unlock()

	// Process the batch synchronously to avoid goroutine leaks in tests
	start := time.Now()
	responses := processor(context.Background(), requests)

	bp.mu.Lock()
	bp.stats.TotalSavings += time.Since(start)
	bp.mu.Unlock()

	// Send responses back to requesters
	responseMap := make(map[string]BatchResponse)
	for _, resp := range responses {
		responseMap[resp.ID] = resp
	}

	for _, req := range requests {
		if resp, found := responseMap[req.ID]; found {
			req.Response <- resp
		} else {
			req.Response <- BatchResponse{
				ID:    req.ID,
				Error: fmt.Errorf("no response found for request ID: %s", req.ID),
			}
		}

		close(req.Response)
	}
}

// processBatches is a worker goroutine that processes batches.
func (bp *BatchProcessor) processBatches() {
	defer bp.wg.Done()

	ticker := time.NewTicker(bp.flushInterval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-bp.stopCh:
			return
		case <-ticker.C:
			bp.flushExpiredBatches()
		}
	}
}

// flushExpiredBatches flushes batches that have been waiting too long.
func (bp *BatchProcessor) flushExpiredBatches() {
	bp.mu.RLock()

	expiredKeys := make([]string, 0)
	now := time.Now()

	for key, batch := range bp.batches {
		batch.mu.Lock()

		if now.Sub(batch.lastUpdate) >= bp.flushInterval {
			expiredKeys = append(expiredKeys, key)
		}

		batch.mu.Unlock()
	}

	bp.mu.RUnlock()

	// Flush expired batches
	for _, key := range expiredKeys {
		bp.flushBatch(key)
	}
}

// GetStats returns current batch processing statistics.
func (bp *BatchProcessor) GetStats() BatchStats {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	return bp.stats
}

// Stop stops the batch processor and flushes all pending batches.
func (bp *BatchProcessor) Stop() {
	close(bp.stopCh)
	bp.wg.Wait()

	// Flush all remaining batches
	bp.mu.Lock()

	for key := range bp.batches {
		bp.flushBatch(key)
	}

	bp.mu.Unlock()
}

// PrintStats prints detailed batch processing statistics.
func (bp *BatchProcessor) PrintStats() {
	stats := bp.GetStats()

	fmt.Printf("=== Batch Processing Statistics ===\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Total Batches: %d\n", stats.TotalBatches)
	fmt.Printf("Average Batch Size: %.2f\n", stats.AverageBatchSize)
	fmt.Printf("Total Time Saved: %v\n", stats.TotalSavings)
	fmt.Printf("Last Processed: %v\n", stats.LastProcessed)
}

// RepositoryBatchProcessor specialized batch processor for repository metadata.
type RepositoryBatchProcessor struct {
	*BatchProcessor
	serviceName string
}

// NewRepositoryBatchProcessor creates a batch processor optimized for repository operations.
func NewRepositoryBatchProcessor(serviceName string) *RepositoryBatchProcessor {
	config := BatchConfig{
		MaxBatchSize:  100, // GitHub GraphQL can handle up to 100 repos per query
		FlushInterval: 200 * time.Millisecond,
		Concurrency:   3,
	}

	return &RepositoryBatchProcessor{
		BatchProcessor: NewBatchProcessor(config),
		serviceName:    serviceName,
	}
}

// BatchRepositoryMetadata batches repository metadata requests.
func (rbp *RepositoryBatchProcessor) BatchRepositoryMetadata(ctx context.Context, org string, repos []string, fetcher BatchFunc) (map[string]interface{}, error) {
	responses := make(map[string]interface{})
	respChans := make(map[string]chan BatchResponse)

	// Create batch requests
	for i, repo := range repos {
		respChan := make(chan BatchResponse, 1)
		respChans[repo] = respChan

		request := &BatchRequest{
			ID:       fmt.Sprintf("%s/%s", org, repo),
			Data:     map[string]string{"org": org, "repo": repo},
			Response: respChan,
		}

		batchKey := fmt.Sprintf("%s:metadata:%s", rbp.serviceName, org)

		err := rbp.Add(ctx, batchKey, request, fetcher)
		if err != nil {
			return nil, fmt.Errorf("failed to add request %d to batch: %w", i, err)
		}
	}

	// Collect responses
	for repo, respChan := range respChans {
		select {
		case resp := <-respChan:
			if resp.Error != nil {
				return nil, fmt.Errorf("batch request failed for %s: %w", repo, resp.Error)
			}

			responses[repo] = resp.Data
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return responses, nil
}

// BatchDefaultBranches batches default branch requests for multiple repositories.
func (rbp *RepositoryBatchProcessor) BatchDefaultBranches(ctx context.Context, org string, repos []string, fetcher BatchFunc) (map[string]string, error) {
	responses := make(map[string]string)
	respChans := make(map[string]chan BatchResponse)

	// Create batch requests
	for _, repo := range repos {
		respChan := make(chan BatchResponse, 1)
		respChans[repo] = respChan

		request := &BatchRequest{
			ID:       fmt.Sprintf("%s/%s", org, repo),
			Data:     map[string]string{"org": org, "repo": repo},
			Response: respChan,
		}

		batchKey := fmt.Sprintf("%s:branches:%s", rbp.serviceName, org)

		err := rbp.Add(ctx, batchKey, request, fetcher)
		if err != nil {
			return nil, fmt.Errorf("failed to add branch request for %s: %w", repo, err)
		}
	}

	// Collect responses
	for repo, respChan := range respChans {
		select {
		case resp := <-respChan:
			if resp.Error != nil {
				return nil, fmt.Errorf("batch request failed for %s: %w", repo, resp.Error)
			}

			if branch, ok := resp.Data.(string); ok {
				responses[repo] = branch
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return responses, nil
}

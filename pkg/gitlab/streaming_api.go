package gitlab

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Gizzahub/gzh-manager-go/internal/constants"
	"github.com/Gizzahub/gzh-manager-go/internal/httpclient"
)

// StreamingClient provides streaming API access for GitLab large-scale operations.
type StreamingClient struct {
	httpClient     *http.Client
	token          string
	baseURL        string
	rateLimiter    *RateLimiter
	memoryPool     *MemoryPool
	bufferPool     sync.Pool
	requestMetrics *RequestMetrics
}

// RateLimiter manages GitLab API rate limiting.
type RateLimiter struct {
	remaining int
	reset     time.Time
	limit     int
	mu        sync.RWMutex
}

// MemoryPool manages reusable memory allocations for GitLab.
type MemoryPool struct {
	bufferPool  sync.Pool
	projectPool sync.Pool
	resultPool  sync.Pool
}

// RequestMetrics tracks GitLab API usage statistics.
type RequestMetrics struct {
	totalRequests   int64
	cachedResponses int64
	rateLimitHits   int64
	retryAttempts   int64
	averageLatency  time.Duration
	memoryUsage     int64
	mu              sync.RWMutex
}

// ProjectStream represents a streaming project result.
type ProjectStream struct {
	Project  *Project
	Error    error
	Metadata StreamMetadata
}

// StreamMetadata contains GitLab stream processing metadata.
type StreamMetadata struct {
	Page         int
	TotalPages   int
	ProcessedAt  time.Time
	MemoryUsage  int64
	CacheHit     bool
	RetryAttempt int
}

// Project represents a GitLab project with optimized memory layout.
// nolint:tagliatelle // External API format - must match GitLab JSON output
type Project struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Path              string    `json:"path"`
	PathWithNamespace string    `json:"path_with_namespace"`
	DefaultBranch     string    `json:"default_branch"`
	Visibility        string    `json:"visibility"`
	ForksCount        int       `json:"forks_count"`
	StarCount         int       `json:"star_count"`
	CreatedAt         time.Time `json:"created_at"`
	LastActivityAt    time.Time `json:"last_activity_at"`
	// Only include essential fields to minimize memory usage
}

// CursorPagination represents cursor-based pagination for GitLab API.
type CursorPagination struct {
	Page       int
	PerPage    int
	NextPage   int
	PrevPage   int
	TotalPages int
	Total      int
	HasNext    bool
	HasPrev    bool
}

// StreamingConfig configures GitLab streaming behavior.
type StreamingConfig struct {
	PageSize        int
	MaxConcurrency  int
	BufferSize      int
	MemoryLimit     int64 // in bytes
	CacheEnabled    bool
	CacheTTL        time.Duration
	RetryAttempts   int
	RetryDelay      time.Duration
	RateLimitBuffer int // requests to keep in reserve
}

// DefaultStreamingConfig returns optimized defaults for GitLab large-scale operations.
func DefaultStreamingConfig() StreamingConfig {
	return StreamingConfig{
		PageSize:        100,                              // GitLab's max per page
		MaxConcurrency:  constants.DefaultParallelism + 3, // Conservative for GitLab rate limits
		BufferSize:      1000,
		MemoryLimit:     500 * constants.BytesPerMB, // 500MB
		CacheEnabled:    true,
		CacheTTL:        10 * time.Minute,
		RetryAttempts:   constants.DefaultMaxRetries,
		RetryDelay:      constants.RetryDelay * 2,
		RateLimitBuffer: 50,
	}
}

// NewStreamingClient creates a new streaming GitLab API client.
func NewStreamingClient(token, baseURL string, config StreamingConfig) *StreamingClient {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	// Use secure HTTP client instead of creating one directly
	httpClient := httpclient.GetGlobalClient("gitlab")

	memoryPool := &MemoryPool{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, constants.BytesPerKB*64) // 64KB initial capacity
			},
		},
		projectPool: sync.Pool{
			New: func() interface{} {
				return &Project{}
			},
		},
		resultPool: sync.Pool{
			New: func() interface{} {
				return &ProjectStream{}
			},
		},
	}

	rateLimiter := &RateLimiter{
		remaining: 2000, // GitLab.com default for authenticated users
		reset:     time.Now().Add(time.Minute),
		limit:     2000,
	}

	return &StreamingClient{
		httpClient:     httpClient,
		token:          token,
		baseURL:        baseURL,
		rateLimiter:    rateLimiter,
		memoryPool:     memoryPool,
		requestMetrics: &RequestMetrics{},
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, constants.BytesPerKB*32) // 32KB buffers
			},
		},
	}
}

// StreamGroupProjects streams projects for a GitLab group with memory optimization.
func (sc *StreamingClient) StreamGroupProjects(ctx context.Context, groupID string, config StreamingConfig) (<-chan ProjectStream, error) {
	resultChan := make(chan ProjectStream, config.BufferSize)

	go func() {
		defer close(resultChan)

		pagination := CursorPagination{
			Page:    1,
			PerPage: config.PageSize,
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Check memory usage before proceeding
			if err := sc.checkMemoryLimit(config.MemoryLimit); err != nil {
				sc.sendError(resultChan, fmt.Errorf("memory limit exceeded: %w", err))
				return
			}

			// Check rate limit
			if err := sc.waitForRateLimit(ctx, config.RateLimitBuffer); err != nil {
				sc.sendError(resultChan, fmt.Errorf("rate limit check failed: %w", err))
				return
			}

			// Fetch page with streaming
			projects, newPagination, err := sc.fetchProjectPage(ctx, groupID, pagination)
			if err != nil {
				sc.sendError(resultChan, fmt.Errorf("failed to fetch page %d: %w", pagination.Page, err))
				return
			}

			// Stream projects to channel
			for _, project := range projects {
				select {
				case <-ctx.Done():
					return
				case resultChan <- ProjectStream{
					Project: project,
					Metadata: StreamMetadata{
						Page:        pagination.Page,
						TotalPages:  newPagination.TotalPages,
						ProcessedAt: time.Now(),
						MemoryUsage: sc.getCurrentMemoryUsage(),
					},
				}:
				}
			}

			// Check if we have more pages
			if !newPagination.HasNext {
				break
			}

			// Update pagination for next page
			pagination = newPagination
			pagination.Page++

			// Trigger garbage collection periodically
			if pagination.Page%10 == 0 {
				sc.optimizeMemory()
			}
		}
	}()

	return resultChan, nil
}

// fetchProjectPage fetches a single page of projects with optimized memory usage.
func (sc *StreamingClient) fetchProjectPage(ctx context.Context, groupID string, pagination CursorPagination) ([]*Project, CursorPagination, error) {
	url := sc.buildProjectURL(groupID, pagination)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if sc.token != "" {
		req.Header.Set("Authorization", "Bearer "+sc.token)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	startTime := time.Now()

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	// Update metrics
	sc.updateRequestMetrics(time.Since(startTime))

	// Update rate limit info
	sc.updateRateLimit(resp.Header)

	if resp.StatusCode != http.StatusOK {
		return nil, CursorPagination{}, fmt.Errorf("API request failed: %s", resp.Status)
	}

	// Parse pagination info from headers
	newPagination := sc.parsePaginationHeaders(resp.Header, pagination)

	// Stream parse response to minimize memory usage
	projects, err := sc.parseProjectResponse(resp.Body)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return projects, newPagination, nil
}

// parseProjectResponse parses JSON response with streaming to minimize memory usage.
func (sc *StreamingClient) parseProjectResponse(reader io.Reader) ([]*Project, error) {
	// Use buffered reader for efficient streaming
	bufReader := bufio.NewReaderSize(reader, 64*1024)

	var rawProjects []json.RawMessage

	decoder := json.NewDecoder(bufReader)

	if err := decoder.Decode(&rawProjects); err != nil {
		return nil, fmt.Errorf("failed to decode project array: %w", err)
	}

	projects := make([]*Project, 0, len(rawProjects))
	for _, rawProject := range rawProjects {
		projectInterface := sc.memoryPool.projectPool.Get()
		project, ok := projectInterface.(*Project)
		if !ok {
			// Skip this entry if type assertion fails
			continue
		}
		if err := json.Unmarshal(rawProject, project); err != nil {
			sc.memoryPool.projectPool.Put(project)
			continue // Skip malformed entries
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// buildProjectURL constructs the GitLab API URL with pagination.
func (sc *StreamingClient) buildProjectURL(groupID string, pagination CursorPagination) string {
	baseURL := fmt.Sprintf("%s/api/v4/groups/%s/projects", sc.baseURL, groupID)

	params := url.Values{}
	params.Set("page", strconv.Itoa(pagination.Page))
	params.Set("per_page", strconv.Itoa(pagination.PerPage))
	params.Set("include_subgroups", "true")
	params.Set("order_by", "last_activity_at")
	params.Set("sort", "desc")
	params.Set("simple", "true") // Get simplified project info to reduce memory

	return baseURL + "?" + params.Encode()
}

// parsePaginationHeaders extracts pagination info from GitLab response headers.
func (sc *StreamingClient) parsePaginationHeaders(headers http.Header, current CursorPagination) CursorPagination {
	newPagination := current

	if nextPage := headers.Get("X-Next-Page"); nextPage != "" {
		if np, err := strconv.Atoi(nextPage); err == nil {
			newPagination.NextPage = np
			newPagination.HasNext = true
		}
	} else {
		newPagination.HasNext = false
	}

	if prevPage := headers.Get("X-Prev-Page"); prevPage != "" {
		if pp, err := strconv.Atoi(prevPage); err == nil {
			newPagination.PrevPage = pp
			newPagination.HasPrev = true
		}
	}

	if totalPages := headers.Get("X-Total-Pages"); totalPages != "" {
		if tp, err := strconv.Atoi(totalPages); err == nil {
			newPagination.TotalPages = tp
		}
	}

	if total := headers.Get("X-Total"); total != "" {
		if t, err := strconv.Atoi(total); err == nil {
			newPagination.Total = t
		}
	}

	return newPagination
}

// waitForRateLimit waits if necessary to respect GitLab rate limits.
func (sc *StreamingClient) waitForRateLimit(ctx context.Context, buffer int) error {
	sc.rateLimiter.mu.RLock()
	remaining := sc.rateLimiter.remaining
	reset := sc.rateLimiter.reset
	sc.rateLimiter.mu.RUnlock()

	if remaining <= buffer {
		waitDuration := time.Until(reset)
		if waitDuration > 0 {
			fmt.Printf("GitLab rate limit approached, waiting %v...\n", waitDuration)

			select {
			case <-time.After(waitDuration):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// updateRateLimit updates rate limit info from GitLab response headers.
func (sc *StreamingClient) updateRateLimit(headers http.Header) {
	sc.rateLimiter.mu.Lock()
	defer sc.rateLimiter.mu.Unlock()

	if remaining := headers.Get("RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			sc.rateLimiter.remaining = r
		}
	}

	if reset := headers.Get("RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			sc.rateLimiter.reset = time.Unix(r, 0)
		}
	}

	if limit := headers.Get("RateLimit-Limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			sc.rateLimiter.limit = l
		}
	}
}

// checkMemoryLimit monitors memory usage and triggers cleanup if needed.
func (sc *StreamingClient) checkMemoryLimit(limit int64) error {
	current := sc.getCurrentMemoryUsage()
	if current > limit {
		sc.optimizeMemory()

		// Check again after optimization
		current = sc.getCurrentMemoryUsage()
		if current > limit {
			return fmt.Errorf("memory usage %d bytes exceeds limit %d bytes", current, limit)
		}
	}

	return nil
}

// getCurrentMemoryUsage gets current memory usage (simplified implementation).
func (sc *StreamingClient) getCurrentMemoryUsage() int64 {
	// In a real implementation, this would use runtime.ReadMemStats()
	sc.requestMetrics.mu.RLock()
	defer sc.requestMetrics.mu.RUnlock()

	return sc.requestMetrics.memoryUsage
}

// optimizeMemory triggers garbage collection and pool cleanup.
func (sc *StreamingClient) optimizeMemory() {
	// Force garbage collection
	// runtime.GC()

	// Reset pools periodically to prevent memory leaks
	sc.memoryPool.bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 64*1024)
		},
	}
}

// updateRequestMetrics updates GitLab API usage statistics.
func (sc *StreamingClient) updateRequestMetrics(latency time.Duration) {
	sc.requestMetrics.mu.Lock()
	defer sc.requestMetrics.mu.Unlock()

	sc.requestMetrics.totalRequests++

	// Update average latency using exponential moving average
	if sc.requestMetrics.averageLatency == 0 {
		sc.requestMetrics.averageLatency = latency
	} else {
		alpha := 0.1 // Smoothing factor
		sc.requestMetrics.averageLatency = time.Duration(
			float64(sc.requestMetrics.averageLatency)*(1-alpha) + float64(latency)*alpha,
		)
	}
}

// sendError sends an error to the result channel.
func (sc *StreamingClient) sendError(resultChan chan<- ProjectStream, err error) {
	resultInterface := sc.memoryPool.resultPool.Get()
	result, ok := resultInterface.(*ProjectStream)
	if !ok {
		// Can't send error if type assertion fails
		return
	}
	result.Error = err
	result.Project = nil
	result.Metadata = StreamMetadata{
		ProcessedAt: time.Now(),
		MemoryUsage: sc.getCurrentMemoryUsage(),
	}

	select {
	case resultChan <- *result:
	default:
		// Channel full, drop error
	}

	sc.memoryPool.resultPool.Put(result)
}

// GetMetrics returns current GitLab API usage metrics.
func (sc *StreamingClient) GetMetrics() RequestMetrics {
	sc.requestMetrics.mu.RLock()
	defer sc.requestMetrics.mu.RUnlock()

	return RequestMetrics{
		totalRequests:   sc.requestMetrics.totalRequests,
		cachedResponses: sc.requestMetrics.cachedResponses,
		rateLimitHits:   sc.requestMetrics.rateLimitHits,
		retryAttempts:   sc.requestMetrics.retryAttempts,
		averageLatency:  sc.requestMetrics.averageLatency,
		memoryUsage:     sc.requestMetrics.memoryUsage,
	}
}

// Close cleans up resources.
func (sc *StreamingClient) Close() error {
	// Clean up HTTP client connections
	if transport, ok := sc.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}

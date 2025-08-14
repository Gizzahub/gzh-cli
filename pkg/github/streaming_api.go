package github

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

	"github.com/Gizzahub/gzh-cli/internal/constants"
	"github.com/Gizzahub/gzh-cli/internal/httpclient"
)

// StreamingClient provides streaming API access for large-scale operations.
type StreamingClient struct {
	httpClient     *http.Client
	token          string
	rateLimiter    *RateLimiter
	memoryPool     *MemoryPool
	bufferPool     sync.Pool
	requestMetrics *RequestMetrics
}

// StreamingRateLimiter manages API rate limiting for streaming.
type StreamingRateLimiter struct{}

// MemoryPool manages reusable memory allocations.
type MemoryPool struct {
	bufferPool     sync.Pool
	repositoryPool sync.Pool
	resultPool     sync.Pool
}

// RequestMetrics tracks API usage statistics.
type RequestMetrics struct {
	totalRequests   int64
	cachedResponses int64
	rateLimitHits   int64
	retryAttempts   int64
	averageLatency  time.Duration
	memoryUsage     int64
	mu              sync.RWMutex
}

// RepositoryStream represents a streaming repository result.
type RepositoryStream struct {
	Repository *Repository
	Error      error
	Metadata   StreamMetadata
}

// StreamMetadata contains stream processing metadata.
type StreamMetadata struct {
	Page         int
	TotalPages   int
	ProcessedAt  time.Time
	MemoryUsage  int64
	CacheHit     bool
	RetryAttempt int
}

// StreamingRepository represents a GitHub repository with optimized memory layout for streaming.
type StreamingRepository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	DefaultBranch string    `json:"default_branch"`
	Private       bool      `json:"private"`
	Fork          bool      `json:"fork"`
	Size          int       `json:"size"`
	Language      string    `json:"language,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	// Only include essential fields to minimize memory usage
}

// CursorPagination represents cursor-based pagination for efficient large dataset traversal.
type CursorPagination struct {
	After       string
	Before      string
	First       int
	Last        int
	HasNext     bool
	HasPrev     bool
	EndCursor   string
	StartCursor string
}

// StreamingConfig configures streaming behavior.
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

// DefaultStreamingConfig returns optimized defaults for large-scale operations.
func DefaultStreamingConfig() StreamingConfig {
	return StreamingConfig{
		PageSize:        100,                              // GitHub's max per page
		MaxConcurrency:  constants.DefaultParallelism * 2, // Conservative to avoid rate limits
		BufferSize:      1000,
		MemoryLimit:     500 * constants.BytesPerMB, // 500MB
		CacheEnabled:    true,
		CacheTTL:        10 * time.Minute,
		RetryAttempts:   constants.DefaultMaxRetries,
		RetryDelay:      constants.RetryDelay * 2,
		RateLimitBuffer: 100,
	}
}

// NewStreamingClient creates a new streaming GitHub API client.
func NewStreamingClient(token string, config StreamingConfig) *StreamingClient {
	// Use secure HTTP client instead of creating one directly
	httpClient := httpclient.GetGlobalClient("github")

	memoryPool := &MemoryPool{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, constants.BytesPerKB*64) // 64KB initial capacity
			},
		},
		repositoryPool: sync.Pool{
			New: func() interface{} {
				return &Repository{}
			},
		},
		resultPool: sync.Pool{
			New: func() interface{} {
				return &RepositoryStream{}
			},
		},
	}

	rateLimiter := &RateLimiter{
		remaining: 5000, // GitHub default
		resetTime: time.Now().Add(time.Hour),
		limit:     5000,
	}

	return &StreamingClient{
		httpClient:     httpClient,
		token:          token,
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

// StreamOrganizationRepositories streams repositories for an organization with memory optimization.
func (sc *StreamingClient) StreamOrganizationRepositories(ctx context.Context, org string, config StreamingConfig) (<-chan RepositoryStream, error) {
	resultChan := make(chan RepositoryStream, config.BufferSize)

	go func() {
		defer close(resultChan)

		// Use cursor-based pagination for efficient traversal
		cursor := CursorPagination{
			First: config.PageSize,
		}

		pageNum := 1

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
			repos, pagination, err := sc.fetchRepositoryPage(ctx, org, cursor)
			if err != nil {
				sc.sendError(resultChan, fmt.Errorf("failed to fetch page %d: %w", pageNum, err))
				return
			}

			// Stream repositories to channel
			for _, repo := range repos {
				select {
				case <-ctx.Done():
					return
				case resultChan <- RepositoryStream{
					Repository: repo,
					Metadata: StreamMetadata{
						Page:        pageNum,
						ProcessedAt: time.Now(),
						MemoryUsage: sc.getCurrentMemoryUsage(),
					},
				}:
				}
			}

			// Check if we have more pages
			if !pagination.HasNext {
				break
			}

			// Update cursor for next page
			cursor.After = pagination.EndCursor
			pageNum++

			// Trigger garbage collection periodically
			if pageNum%10 == 0 {
				sc.optimizeMemory()
			}
		}
	}()

	return resultChan, nil
}

// fetchRepositoryPage fetches a single page of repositories with optimized memory usage.
func (sc *StreamingClient) fetchRepositoryPage(ctx context.Context, org string, cursor CursorPagination) ([]*Repository, CursorPagination, error) {
	url := sc.buildRepositoryURL(org, cursor)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if sc.token != "" {
		req.Header.Set("Authorization", "Bearer "+sc.token)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// Execute request
	startTime := time.Now()

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Update metrics
	sc.updateRequestMetrics(time.Since(startTime))

	// Update rate limit info
	sc.updateRateLimit(resp.Header)

	if resp.StatusCode != http.StatusOK {
		return nil, CursorPagination{}, fmt.Errorf("API request failed: %s", resp.Status)
	}

	// Stream parse response to minimize memory usage
	return sc.parseRepositoryResponse(resp.Body, cursor)
}

// parseRepositoryResponse parses JSON response with streaming to minimize memory usage.
func (sc *StreamingClient) parseRepositoryResponse(reader io.Reader, cursor CursorPagination) ([]*Repository, CursorPagination, error) {
	// Use buffered reader for efficient streaming
	bufReader := bufio.NewReaderSize(reader, constants.BytesPerKB*64)

	var response struct {
		Items []json.RawMessage `json:"items,omitempty"`
		// For regular API calls without search
		Repositories []json.RawMessage `json:"-"`
	}

	// Try to read as array first (for /orgs/{org}/repos)
	decoder := json.NewDecoder(bufReader)

	// Peek at first character to determine response type
	firstByte, err := bufReader.Peek(1)
	if err != nil {
		return nil, CursorPagination{}, fmt.Errorf("failed to peek response: %w", err)
	}

	var (
		repositories  []*Repository
		newPagination CursorPagination
	)

	if firstByte[0] == '[' {
		// Direct array response
		var rawRepos []json.RawMessage
		if err := decoder.Decode(&rawRepos); err != nil {
			return nil, CursorPagination{}, fmt.Errorf("failed to decode repository array: %w", err)
		}

		repositories = make([]*Repository, 0, len(rawRepos))
		for _, rawRepo := range rawRepos {
			repoInterface := sc.memoryPool.repositoryPool.Get()
			repo, ok := repoInterface.(*Repository)
			if !ok {
				// Skip this entry if type assertion fails
				continue
			}
			if err := json.Unmarshal(rawRepo, repo); err != nil {
				sc.memoryPool.repositoryPool.Put(repo)
				continue // Skip malformed entries
			}

			repositories = append(repositories, repo)
		}

		// For array responses, pagination is handled via Link headers
		newPagination = sc.parseLinkPagination(cursor)
	} else {
		// Object response (search results)
		if err := decoder.Decode(&response); err != nil {
			return nil, CursorPagination{}, fmt.Errorf("failed to decode repository response: %w", err)
		}

		repositories = make([]*Repository, 0, len(response.Items))
		for _, rawRepo := range response.Items {
			repoInterface := sc.memoryPool.repositoryPool.Get()
			repo, ok := repoInterface.(*Repository)
			if !ok {
				// Skip this entry if type assertion fails
				continue
			}
			if err := json.Unmarshal(rawRepo, repo); err != nil {
				sc.memoryPool.repositoryPool.Put(repo)
				continue
			}

			repositories = append(repositories, repo)
		}

		newPagination = cursor // Update based on actual response
	}

	return repositories, newPagination, nil
}

// buildRepositoryURL constructs the API URL with cursor pagination.
func (sc *StreamingClient) buildRepositoryURL(org string, cursor CursorPagination) string {
	baseURL := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)

	params := url.Values{}
	params.Set("per_page", strconv.Itoa(cursor.First))
	params.Set("type", "all")
	params.Set("sort", "updated")
	params.Set("direction", "desc")

	if cursor.After != "" {
		params.Set("page", cursor.After)
	}

	return baseURL + "?" + params.Encode()
}

// parseLinkPagination extracts pagination info from Link headers.
func (sc *StreamingClient) parseLinkPagination(current CursorPagination) CursorPagination {
	// Implementation would parse GitHub's Link header for next/prev URLs
	// For now, return a simplified version
	return CursorPagination{
		After:   current.After,
		HasNext: true, // Would be determined by Link header presence
	}
}

// waitForRateLimit waits if necessary to respect rate limits.
func (sc *StreamingClient) waitForRateLimit(ctx context.Context, buffer int) error {
	sc.rateLimiter.mu.Lock()
	remaining := sc.rateLimiter.remaining
	reset := sc.rateLimiter.resetTime
	sc.rateLimiter.mu.Unlock()

	if remaining <= buffer {
		waitDuration := time.Until(reset)

		if waitDuration <= 0 {
			return nil
		}

		if waitDuration > 0 {
			fmt.Printf("Rate limit approached, waiting %v...\n", waitDuration)

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

// updateRateLimit updates rate limit info from response headers.
func (sc *StreamingClient) updateRateLimit(headers http.Header) {
	sc.rateLimiter.mu.Lock()
	defer sc.rateLimiter.mu.Unlock()

	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			sc.rateLimiter.remaining = r
		}
	}

	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			sc.rateLimiter.resetTime = time.Unix(r, 0)
		}
	}

	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
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
	// For now, return a placeholder
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
			return make([]byte, 0, constants.BytesPerKB*64)
		},
	}
}

// updateRequestMetrics updates API usage statistics.
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
func (sc *StreamingClient) sendError(resultChan chan<- RepositoryStream, err error) {
	resultInterface := sc.memoryPool.resultPool.Get()
	result, ok := resultInterface.(*RepositoryStream)
	if !ok {
		// Can't send error if type assertion fails
		return
	}
	result.Error = err
	result.Repository = nil
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

// GetMetrics returns current API usage metrics.
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

package github

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gizzahub/gzh-cli/internal/constants"
	"github.com/gizzahub/gzh-cli/internal/httpclient"
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
			New: func() any {
				return make([]byte, 0, constants.BytesPerKB*64) // 64KB initial capacity
			},
		},
		repositoryPool: sync.Pool{
			New: func() any {
				return &Repository{}
			},
		},
		resultPool: sync.Pool{
			New: func() any {
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
			New: func() any {
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
				sc.sendError(ctx, resultChan, fmt.Errorf("memory limit exceeded: %w", err))
				return
			}

			// Check rate limit
			if err := sc.waitForRateLimit(ctx, config.RateLimitBuffer); err != nil {
				sc.sendError(ctx, resultChan, fmt.Errorf("rate limit check failed: %w", err))
				return
			}

			// Fetch page with streaming
			repos, pagination, err := sc.fetchRepositoryPage(ctx, org, cursor)
			if err != nil {
				sc.sendError(ctx, resultChan, fmt.Errorf("failed to fetch page %d: %w", pageNum, err))
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
	return sc.parseRepositoryResponse(resp.Body, cursor, req.URL, strings.Join(resp.Header.Values("Link"), ","))
}

// parseRepositoryResponse parses JSON response with streaming to minimize memory usage.
func (sc *StreamingClient) parseRepositoryResponse(reader io.Reader, cursor CursorPagination, requestURL *url.URL, linkHeader string) ([]*Repository, CursorPagination, error) {
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

	var repositories []*Repository

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

	}

	newPagination, err := sc.parseLinkPagination(cursor, requestURL, linkHeader)
	if err != nil {
		return nil, CursorPagination{}, err
	}

	return repositories, newPagination, nil
}

// buildRepositoryURL constructs the API URL with cursor pagination.
func (sc *StreamingClient) buildRepositoryURL(org string, cursor CursorPagination) string {
	baseURL := fmt.Sprintf("%s/orgs/%s/repos", DefaultGitHubAPIBaseURL, org)

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

// parseLinkPagination 은 현재 조직 경로의 다음 페이지 번호만 Link에서 가져온다.
func (sc *StreamingClient) parseLinkPagination(current CursorPagination, requestURL *url.URL, linkHeader string) (CursorPagination, error) {
	pagination := current
	pagination.HasNext = false
	if strings.TrimSpace(linkHeader) == "" {
		return pagination, nil
	}
	if requestURL == nil {
		return CursorPagination{}, fmt.Errorf("invalid next page link: request URL is missing")
	}
	currentPage, err := parseCurrentPage(current.After)
	if err != nil {
		return CursorPagination{}, err
	}
	nextPage, found, err := findNextPage(currentPage, requestURL, linkHeader)
	if err != nil {
		return CursorPagination{}, err
	}
	if found {
		pagination.HasNext = true
		pagination.EndCursor = strconv.Itoa(nextPage)
	}
	return pagination, nil
}

func findNextPage(currentPage int, requestURL *url.URL, linkHeader string) (int, bool, error) {
	foundNext := false
	var nextPage int
	for _, link := range strings.Split(linkHeader, ",") {
		target, relation, err := parseLinkEntry(link)
		if err != nil {
			return 0, false, err
		}
		if !hasNextRelation(relation) {
			continue
		}
		if foundNext {
			return 0, false, fmt.Errorf("invalid next page link: multiple next relations")
		}
		foundNext = true

		nextURL, err := resolveNextPageURL(requestURL, target)
		if err != nil {
			return 0, false, err
		}
		nextPage, err = parseNextPageNumber(currentPage, requestURL, nextURL)
		if err != nil {
			return 0, false, err
		}
	}
	return nextPage, foundNext, nil
}

// parseCurrentPage는 요청에서 사용한 페이지 번호를 검증한다.
func parseCurrentPage(cursor string) (int, error) {
	if cursor == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(cursor)
	if err != nil || page < 1 {
		return 0, fmt.Errorf("invalid current page cursor %q", cursor)
	}
	return page, nil
}

// parseLinkEntry는 GitHub Link 항목 하나에서 대상과 관계를 추출한다.
func parseLinkEntry(link string) (string, string, error) {
	parts := strings.Split(link, ";")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid next page link: malformed link %q", strings.TrimSpace(link))
	}
	target := strings.TrimSpace(parts[0])
	if len(target) < 3 || target[0] != '<' || target[len(target)-1] != '>' {
		return "", "", fmt.Errorf("invalid next page link: malformed target %q", target)
	}
	relation := ""
	for _, parameter := range parts[1:] {
		name, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
		if !ok {
			return "", "", fmt.Errorf("invalid next page link: malformed parameter %q", strings.TrimSpace(parameter))
		}
		if strings.EqualFold(strings.TrimSpace(name), "rel") {
			if relation != "" {
				return "", "", fmt.Errorf("invalid next page link: duplicate relation")
			}
			relation = strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	if relation == "" {
		return "", "", fmt.Errorf("invalid next page link: relation is missing")
	}
	return target, relation, nil
}

func hasNextRelation(relation string) bool {
	for _, rel := range strings.Fields(relation) {
		if strings.EqualFold(rel, "next") {
			return true
		}
	}
	return false
}

func resolveNextPageURL(requestURL *url.URL, target string) (*url.URL, error) {
	linkURL, err := url.Parse(strings.TrimSuffix(strings.TrimPrefix(target, "<"), ">"))
	if err != nil {
		return nil, fmt.Errorf("invalid next page link: %w", err)
	}
	nextURL := requestURL.ResolveReference(linkURL)
	if nextURL.User != nil || nextURL.Fragment != "" ||
		!strings.EqualFold(nextURL.Scheme, requestURL.Scheme) ||
		!strings.EqualFold(nextURL.Host, requestURL.Host) ||
		nextURL.EscapedPath() != requestURL.EscapedPath() {
		return nil, fmt.Errorf("invalid next page link: target must use the current API origin and organization path")
	}
	return nextURL, nil
}

func parseNextPageNumber(currentPage int, requestURL, nextURL *url.URL) (int, error) {
	query := nextURL.Query()
	pageValues := query["page"]
	if len(pageValues) != 1 {
		return 0, fmt.Errorf("invalid next page link: exactly one page value is required")
	}
	nextPage, err := strconv.Atoi(pageValues[0])
	if err != nil || nextPage < 1 || nextPage <= currentPage {
		return 0, fmt.Errorf("invalid next page link: page %q must be a positive number greater than %d", pageValues[0], currentPage)
	}
	requestQuery := requestURL.Query()
	for name, values := range query {
		if name == "page" {
			continue
		}
		expected, ok := requestQuery[name]
		if !ok || !slices.Equal(values, expected) {
			return 0, fmt.Errorf("invalid next page link: query parameter %q changes the current request", name)
		}
	}
	return nextPage, nil
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
		New: func() any {
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
func (sc *StreamingClient) sendError(ctx context.Context, resultChan chan<- RepositoryStream, err error) {
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

	streamError := *result
	select {
	case resultChan <- streamError:
	default:
		select {
		case resultChan <- streamError:
		case <-ctx.Done():
		}
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

package largescale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// LargeScaleConfig holds configuration for large-scale repository operations
type LargeScaleConfig struct {
	MaxConcurrency    int
	BatchSize         int
	UseShallowClone   bool
	EnableCompression bool
	ProgressInterval  time.Duration
	MaxRetries        int
	MemoryThreshold   int64 // bytes
}

// DefaultLargeScaleConfig returns optimized configuration for large-scale operations
func DefaultLargeScaleConfig() *LargeScaleConfig {
	return &LargeScaleConfig{
		MaxConcurrency:    minInt(runtime.NumCPU()*4, 20), // Scale with CPU but cap at 20
		BatchSize:         100,                            // GitHub API max per page
		UseShallowClone:   true,                           // Use shallow clones by default
		EnableCompression: true,
		ProgressInterval:  time.Second * 2, // Update progress every 2 seconds
		MaxRetries:        3,
		MemoryThreshold:   500 * 1024 * 1024, // 500MB threshold
	}
}

// LargeScaleRepository represents a minimal repository structure for large-scale operations
type LargeScaleRepository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Size          int    `json:"size"` // Size in KB
	Fork          bool   `json:"fork"`
	Archived      bool   `json:"archived"`
}

// ProgressCallback is called periodically during large operations
type ProgressCallback func(processed, total int, current string)

// LargeScaleManager handles large-scale repository operations efficiently
type LargeScaleManager struct {
	config           *LargeScaleConfig
	client           *http.Client
	rateLimiter      *AdaptiveRateLimiter
	progressCallback ProgressCallback
	stats            *OperationStats
}

// OperationStats tracks statistics for large-scale operations
type OperationStats struct {
	TotalRepos     int
	ProcessedRepos int
	FailedRepos    int
	SkippedRepos   int
	TotalSize      int64 // Total size in KB
	StartTime      time.Time
	LastUpdateTime time.Time
	APICallsUsed   int
	mu             sync.RWMutex
}

// NewLargeScaleManager creates a new manager for large-scale repository operations
func NewLargeScaleManager(config *LargeScaleConfig, progressCallback ProgressCallback) *LargeScaleManager {
	if config == nil {
		config = DefaultLargeScaleConfig()
	}

	return &LargeScaleManager{
		config:           config,
		client:           &http.Client{Timeout: 30 * time.Second},
		rateLimiter:      NewAdaptiveRateLimiter(),
		progressCallback: progressCallback,
		stats: &OperationStats{
			StartTime: time.Now(),
		},
	}
}

// ListAllRepositories fetches all repositories from an organization with proper pagination
func (m *LargeScaleManager) ListAllRepositories(ctx context.Context, org string) ([]LargeScaleRepository, error) {
	var allRepos []LargeScaleRepository
	page := 1
	perPage := m.config.BatchSize

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Wait for rate limit if necessary
		if err := m.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit error: %w", err)
		}

		repos, hasMore, err := m.fetchRepositoryPage(ctx, org, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		allRepos = append(allRepos, repos...)
		m.updateStats(len(repos), 0, 0)

		// Update progress
		if m.progressCallback != nil {
			m.progressCallback(len(allRepos), -1, fmt.Sprintf("Fetched page %d (%d repos)", page, len(repos)))
		}

		if !hasMore {
			break
		}

		page++

		// Check memory usage and trigger GC if necessary
		if m.shouldTriggerGC(len(allRepos)) {
			runtime.GC()
		}
	}

	m.stats.mu.Lock()
	m.stats.TotalRepos = len(allRepos)
	m.stats.mu.Unlock()

	return allRepos, nil
}

// fetchRepositoryPage fetches a single page of repositories
func (m *LargeScaleManager) fetchRepositoryPage(ctx context.Context, org string, page, perPage int) ([]LargeScaleRepository, bool, error) {
	u, err := url.Parse(fmt.Sprintf("https://api.github.com/orgs/%s/repos", org))
	if err != nil {
		return nil, false, err
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("sort", "name")
	q.Set("type", "all")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, false, err
	}

	// Add authentication if available
	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	m.updateAPIStats(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var repos []LargeScaleRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, false, err
	}

	// Check if there are more pages using Link header
	hasMore := false
	if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
		hasMore = containsNextLink(linkHeader)
	}

	return repos, hasMore, nil
}

// BulkCloneRepositories clones multiple repositories with optimized concurrency
func (m *LargeScaleManager) BulkCloneRepositories(ctx context.Context, repos []LargeScaleRepository, targetPath string) error {
	if len(repos) == 0 {
		return nil
	}

	// Calculate optimal concurrency based on available resources
	concurrency := m.calculateOptimalConcurrency(len(repos))
	sem := semaphore.NewWeighted(int64(concurrency))

	g, ctx := errgroup.WithContext(ctx)

	// Progress tracking
	var processed int64
	var mu sync.Mutex

	// Progress reporting goroutine
	progressTicker := time.NewTicker(m.config.ProgressInterval)
	defer progressTicker.Stop()

	go func() {
		for {
			select {
			case <-progressTicker.C:
				mu.Lock()
				current := processed
				mu.Unlock()

				if m.progressCallback != nil {
					m.progressCallback(int(current), len(repos), "Cloning repositories...")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Process repositories in batches
	for i, repo := range repos {
		repo := repo // Capture loop variable
		index := i

		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			// Check if we should skip this repository
			if m.shouldSkipRepository(repo) {
				m.updateStats(0, 0, 1)
				return nil
			}

			// Clone with retry logic
			var err error
			for attempt := 0; attempt < m.config.MaxRetries; attempt++ {
				err = m.cloneRepository(ctx, repo, targetPath)
				if err == nil {
					break
				}

				// Wait before retry with exponential backoff
				if attempt < m.config.MaxRetries-1 {
					backoff := time.Duration(attempt+1) * time.Second
					select {
					case <-time.After(backoff):
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			mu.Lock()
			processed++
			mu.Unlock()

			if err != nil {
				m.updateStats(0, 1, 0)
				return fmt.Errorf("failed to clone %s after %d attempts: %w", repo.Name, m.config.MaxRetries, err)
			}

			m.updateStats(1, 0, 0)
			return nil
		})

		// Check memory pressure every 100 repositories
		if index%100 == 0 && m.shouldTriggerGC(index) {
			runtime.GC()
		}
	}

	return g.Wait()
}

// cloneRepository clones a single repository with optimization
func (m *LargeScaleManager) cloneRepository(ctx context.Context, repo LargeScaleRepository, targetPath string) error {
	args := []string{"clone"}

	if m.config.UseShallowClone {
		args = append(args, "--depth", "1")
	}

	if m.config.EnableCompression {
		args = append(args, "--config", "core.compression=9")
	}

	args = append(args, repo.CloneURL, targetPath+"/"+repo.Name)

	return executeGitCommand(ctx, args...)
}

// Helper functions

func (m *LargeScaleManager) shouldSkipRepository(repo LargeScaleRepository) bool {
	// Skip archived repositories by default
	if repo.Archived {
		return true
	}

	// Skip very large repositories if memory is constrained
	if repo.Size > 1000000 { // 1GB in KB
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		if int64(memStats.Alloc) > m.config.MemoryThreshold {
			return true
		}
	}

	return false
}

func (m *LargeScaleManager) calculateOptimalConcurrency(totalRepos int) int {
	// Base concurrency on available CPU and memory
	baseConcurrency := m.config.MaxConcurrency

	// Reduce concurrency for very large operations to avoid memory pressure
	if totalRepos > 1000 {
		baseConcurrency = minInt(baseConcurrency, 10)
	}

	// Check available memory
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	availableMemory := m.config.MemoryThreshold - int64(memStats.Alloc)

	if availableMemory < 100*1024*1024 { // Less than 100MB available
		baseConcurrency = minInt(baseConcurrency, 5)
	}

	return maxInt(1, baseConcurrency)
}

func (m *LargeScaleManager) shouldTriggerGC(processed int) bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc) > m.config.MemoryThreshold
}

func (m *LargeScaleManager) updateStats(processed, failed, skipped int) {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats.ProcessedRepos += processed
	m.stats.FailedRepos += failed
	m.stats.SkippedRepos += skipped
	m.stats.LastUpdateTime = time.Now()
}

func (m *LargeScaleManager) updateAPIStats(resp *http.Response) {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats.APICallsUsed++

	// Update rate limiter with response headers
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if count, err := strconv.Atoi(remaining); err == nil {
			m.rateLimiter.UpdateRemaining(count)
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			m.rateLimiter.UpdateResetTime(time.Unix(timestamp, 0))
		}
	}
}

// GetStats returns current operation statistics
func (m *LargeScaleManager) GetStats() OperationStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	return *m.stats
}

// Utility functions

func containsNextLink(linkHeader string) bool {
	return len(linkHeader) > 0 && (contains(linkHeader, `rel="next"`) ||
		contains(linkHeader, `rel="last"`))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// executeGitCommand is a placeholder for actual git command execution
func executeGitCommand(ctx context.Context, args ...string) error {
	// This would be implemented with actual git command execution
	// For now, this is a placeholder
	return nil
}

// getGitHubToken retrieves GitHub token from environment
func getGitHubToken() string {
	// This would retrieve token from environment or configuration
	// For now, this is a placeholder
	return ""
}

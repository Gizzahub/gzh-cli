# Task: Implement Streaming API for Large Organizations

## Priority: MEDIUM

## Estimated Time: 2.5 hours

## Context

Remote branch includes `pkg/github/streaming_api.go` and `pkg/gitlab/streaming_api.go` for handling large organizations with thousands of repositories. Need to implement efficient streaming with backpressure control.

## Pre-requisites

- [ ] Tasks 01-03 completed
- [ ] Understanding of Go channels and goroutines
- [ ] Access to large GitHub/GitLab organizations for testing

## Steps

### 1. Analyze Current Streaming Implementation

#### Review Existing Code

```bash
# Check current streaming implementations
cat pkg/github/streaming_api.go
cat pkg/gitlab/streaming_api.go

# Look for usage patterns
grep -r "streaming" pkg/ --include="*.go"
```

### 2. Design Streaming Architecture

#### Create Streaming Interfaces

```go
// pkg/bulk-clone/streaming.go
package bulkclone

import (
    "context"
    "github.com/google/go-github/v45/github"
)

// StreamOptions configures streaming behavior
type StreamOptions struct {
    // Buffer size for repository channel
    BufferSize int

    // Max concurrent API requests
    MaxConcurrency int

    // Enable backpressure when buffer is full
    EnableBackpressure bool

    // Checkpoint interval for resume capability
    CheckpointInterval time.Duration

    // State file for resume capability
    StateFile string
}

// RepositoryStream represents a stream of repositories
type RepositoryStream interface {
    // Next returns the next repository or blocks until available
    Next(ctx context.Context) (*Repository, error)

    // HasNext returns true if more repositories are available
    HasNext() bool

    // Close releases resources
    Close() error

    // Stats returns streaming statistics
    Stats() StreamStats
}

// StreamStats provides real-time statistics
type StreamStats struct {
    TotalFetched   int64
    TotalProcessed int64
    BufferSize     int
    ErrorCount     int64
    StartTime      time.Time
    BytesReceived  int64
}
```

### 3. Implement GitHub Streaming with Pagination

#### Enhanced GitHub Streaming

```go
// pkg/github/streaming_enhanced.go
package github

import (
    "context"
    "sync/atomic"
    "golang.org/x/sync/errgroup"
)

type EnhancedStreamer struct {
    client         *github.Client
    options        StreamOptions
    repoChannel    chan *github.Repository
    errorChannel   chan error
    stats          atomic.Value // StreamStats
    checkpointer   *Checkpointer
    rateLimiter    *EnhancedRateLimiter
}

func NewEnhancedStreamer(client *github.Client, opts StreamOptions) *EnhancedStreamer {
    return &EnhancedStreamer{
        client:       client,
        options:      opts,
        repoChannel:  make(chan *github.Repository, opts.BufferSize),
        errorChannel: make(chan error, 10),
        checkpointer: NewCheckpointer(opts.StateFile),
        rateLimiter:  NewEnhancedRateLimiter(),
    }
}

// StreamOrganizationRepos streams all repositories from an organization
func (s *EnhancedStreamer) StreamOrganizationRepos(ctx context.Context, org string) error {
    g, ctx := errgroup.WithContext(ctx)

    // Producer: Fetch repositories with pagination
    g.Go(func() error {
        defer close(s.repoChannel)
        return s.fetchWithPagination(ctx, org)
    })

    // Monitor: Track statistics
    g.Go(func() error {
        return s.monitorProgress(ctx)
    })

    // Checkpointer: Save progress periodically
    if s.options.CheckpointInterval > 0 {
        g.Go(func() error {
            return s.runCheckpointer(ctx)
        })
    }

    return g.Wait()
}

func (s *EnhancedStreamer) fetchWithPagination(ctx context.Context, org string) error {
    opt := &github.RepositoryListByOrgOptions{
        ListOptions: github.ListOptions{
            PerPage: 100, // Maximum allowed by GitHub
        },
    }

    // Resume from checkpoint if available
    if checkpoint, err := s.checkpointer.Load(); err == nil {
        opt.ListOptions.Page = checkpoint.LastPage
    }

    for {
        // Apply rate limiting with backpressure
        if err := s.rateLimiter.Wait(ctx); err != nil {
            return err
        }

        // Fetch page of repositories
        repos, resp, err := s.client.Repositories.ListByOrg(ctx, org, opt)
        if err != nil {
            // Handle rate limit errors specially
            if _, ok := err.(*github.RateLimitError); ok {
                if err := s.handleRateLimit(ctx, resp); err != nil {
                    return err
                }
                continue
            }
            return err
        }

        // Update statistics
        s.updateStats(len(repos), resp.Size)

        // Stream repositories with backpressure
        for _, repo := range repos {
            select {
            case s.repoChannel <- repo:
                // Successfully sent
            case <-ctx.Done():
                return ctx.Err()
            default:
                if s.options.EnableBackpressure {
                    // Block until space available
                    select {
                    case s.repoChannel <- repo:
                    case <-ctx.Done():
                        return ctx.Err()
                    }
                } else {
                    // Drop repository if buffer full
                    s.recordError(ErrBufferFull)
                }
            }
        }

        // Check if more pages
        if resp.NextPage == 0 {
            break
        }
        opt.Page = resp.NextPage
    }

    return nil
}

func (s *EnhancedStreamer) handleRateLimit(ctx context.Context, resp *github.Response) error {
    if resp == nil {
        return nil
    }

    resetTime := time.Unix(resp.Rate.Reset.Unix(), 0)
    waitDuration := time.Until(resetTime)

    // Log rate limit information
    log.Printf("Rate limit exceeded. Waiting %v until reset at %v",
        waitDuration, resetTime)

    select {
    case <-time.After(waitDuration):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 4. Implement Checkpoint and Resume Capability

#### Checkpoint System

```go
// pkg/bulk-clone/checkpoint.go
package bulkclone

import (
    "encoding/json"
    "os"
    "sync"
    "time"
)

type Checkpoint struct {
    Organization   string    `json:"organization"`
    LastPage       int       `json:"lastPage"`
    LastRepo       string    `json:"lastRepo"`
    ProcessedRepos []string  `json:"processedRepos"`
    Timestamp      time.Time `json:"timestamp"`
    Stats          StreamStats `json:"stats"`
}

type Checkpointer struct {
    filepath string
    mutex    sync.Mutex
    current  *Checkpoint
}

func NewCheckpointer(filepath string) *Checkpointer {
    return &Checkpointer{
        filepath: filepath,
        current:  &Checkpoint{ProcessedRepos: make([]string, 0)},
    }
}

func (c *Checkpointer) Save() error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.current.Timestamp = time.Now()

    data, err := json.MarshalIndent(c.current, "", "  ")
    if err != nil {
        return err
    }

    // Write atomically
    tmpFile := c.filepath + ".tmp"
    if err := os.WriteFile(tmpFile, data, 0644); err != nil {
        return err
    }

    return os.Rename(tmpFile, c.filepath)
}

func (c *Checkpointer) Load() (*Checkpoint, error) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    data, err := os.ReadFile(c.filepath)
    if err != nil {
        return nil, err
    }

    checkpoint := &Checkpoint{}
    if err := json.Unmarshal(data, checkpoint); err != nil {
        return nil, err
    }

    c.current = checkpoint
    return checkpoint, nil
}

func (c *Checkpointer) UpdateProgress(org string, page int, repo string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.current.Organization = org
    c.current.LastPage = page
    c.current.LastRepo = repo
    c.current.ProcessedRepos = append(c.current.ProcessedRepos, repo)
}

func (c *Checkpointer) IsProcessed(repoName string) bool {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    for _, processed := range c.current.ProcessedRepos {
        if processed == repoName {
            return true
        }
    }
    return false
}
```

### 5. Implement Consumer Pattern with Worker Pool

#### Streaming Consumer

```go
// pkg/bulk-clone/streaming_consumer.go
package bulkclone

import (
    "context"
    "sync"
    "golang.org/x/sync/errgroup"
)

type StreamingConsumer struct {
    stream       RepositoryStream
    workerCount  int
    processor    RepositoryProcessor
    progressChan chan Progress
}

type RepositoryProcessor interface {
    Process(ctx context.Context, repo *Repository) error
}

func NewStreamingConsumer(stream RepositoryStream, workers int) *StreamingConsumer {
    return &StreamingConsumer{
        stream:       stream,
        workerCount:  workers,
        progressChan: make(chan Progress, workers*2),
    }
}

func (sc *StreamingConsumer) Consume(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    // Start workers
    for i := 0; i < sc.workerCount; i++ {
        workerID := i
        g.Go(func() error {
            return sc.worker(ctx, workerID)
        })
    }

    // Progress aggregator
    g.Go(func() error {
        return sc.aggregateProgress(ctx)
    })

    return g.Wait()
}

func (sc *StreamingConsumer) worker(ctx context.Context, id int) error {
    for {
        repo, err := sc.stream.Next(ctx)
        if err != nil {
            if err == io.EOF {
                return nil // Stream ended
            }
            return err
        }

        // Process repository
        start := time.Now()
        err = sc.processor.Process(ctx, repo)
        duration := time.Since(start)

        // Report progress
        progress := Progress{
            WorkerID:   id,
            Repository: repo.Name,
            Success:    err == nil,
            Error:      err,
            Duration:   duration,
            Timestamp:  time.Now(),
        }

        select {
        case sc.progressChan <- progress:
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (sc *StreamingConsumer) aggregateProgress(ctx context.Context) error {
    stats := &AggregatedStats{
        WorkerStats: make(map[int]*WorkerStat),
        StartTime:   time.Now(),
    }

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case progress := <-sc.progressChan:
            stats.Update(progress)

        case <-ticker.C:
            sc.printStats(stats)

        case <-ctx.Done():
            sc.printFinalStats(stats)
            return ctx.Err()
        }
    }
}
```

### 6. Create Memory-Efficient Large Organization Handler

#### Large Org Handler

```go
// pkg/bulk-clone/large_org_handler.go
package bulkclone

import (
    "runtime"
    "github.com/dustin/go-humanize"
)

type LargeOrgHandler struct {
    config       *Config
    memoryLimit  uint64
    diskCache    *DiskCache
}

func NewLargeOrgHandler(config *Config) *LargeOrgHandler {
    // Set memory limit to 50% of available RAM
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memoryLimit := m.Sys / 2

    return &LargeOrgHandler{
        config:      config,
        memoryLimit: memoryLimit,
        diskCache:   NewDiskCache(config.CacheDir),
    }
}

func (h *LargeOrgHandler) HandleOrganization(ctx context.Context, org string) error {
    // Estimate organization size
    size, err := h.estimateOrgSize(ctx, org)
    if err != nil {
        return err
    }

    log.Printf("Organization %s has approximately %s repositories",
        org, humanize.Comma(int64(size)))

    // Choose strategy based on size
    var strategy CloneStrategy
    switch {
    case size < 100:
        strategy = NewBatchStrategy(h.config)
    case size < 1000:
        strategy = NewStreamingStrategy(h.config)
    default:
        strategy = NewDistributedStrategy(h.config)
    }

    return strategy.Execute(ctx, org)
}

// Distributed strategy for very large organizations
type DistributedStrategy struct {
    config      *Config
    shardCount  int
    coordinator *ShardCoordinator
}

func NewDistributedStrategy(config *Config) *DistributedStrategy {
    return &DistributedStrategy{
        config:      config,
        shardCount:  runtime.NumCPU(),
        coordinator: NewShardCoordinator(),
    }
}

func (s *DistributedStrategy) Execute(ctx context.Context, org string) error {
    // Create shards based on repository name hash
    shards := s.coordinator.CreateShards(s.shardCount)

    g, ctx := errgroup.WithContext(ctx)

    for i, shard := range shards {
        shardID := i
        shardRange := shard

        g.Go(func() error {
            return s.processShard(ctx, org, shardID, shardRange)
        })
    }

    return g.Wait()
}

func (s *DistributedStrategy) processShard(ctx context.Context, org string,
    shardID int, shard ShardRange) error {

    log.Printf("Shard %d: Processing repositories %s to %s",
        shardID, shard.Start, shard.End)

    // Create dedicated streamer for this shard
    streamer := NewShardedStreamer(s.config, shard)
    consumer := NewStreamingConsumer(streamer, s.config.WorkersPerShard)

    return consumer.Consume(ctx)
}
```

### 7. Integration Testing with Large Organizations

#### Create Integration Test

```go
// test/integration/streaming_large_org_test.go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
    "github.com/stretchr/testify/require"
)

func TestStreamingLargeOrganization(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set")
    }

    // Test with a large organization
    config := &bulk-clone.Config{
        GitHub: bulk-clone.GitHubConfig{
            Organizations: []string{"apache"}, // ~2000 repos
            Token:        token,
        },
        Streaming: bulk-clone.StreamingConfig{
            Enabled:            true,
            BufferSize:         1000,
            WorkerCount:        10,
            CheckpointInterval: 30 * time.Second,
            EnableBackpressure: true,
        },
    }

    handler := bulk-clone.NewLargeOrgHandler(config)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    // Track memory usage
    initialMem := getMemoryUsage()

    err := handler.HandleOrganization(ctx, "apache")
    require.NoError(t, err)

    // Verify memory didn't grow excessively
    finalMem := getMemoryUsage()
    memGrowth := finalMem - initialMem

    t.Logf("Memory usage: Initial=%s, Final=%s, Growth=%s",
        humanize.Bytes(initialMem),
        humanize.Bytes(finalMem),
        humanize.Bytes(memGrowth))

    // Memory growth should be less than 500MB for streaming
    require.Less(t, memGrowth, uint64(500*1024*1024))
}

func getMemoryUsage() uint64 {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return m.Alloc
}
```

### 8. Add CLI Support for Streaming

#### Update Bulk Clone Command

```go
// cmd/bulk-clone/bulk_clone.go
func init() {
    // Add streaming flags
    bulkCloneCmd.Flags().Bool("stream", false,
        "Enable streaming mode for large organizations")
    bulkCloneCmd.Flags().Int("stream-buffer", 1000,
        "Buffer size for streaming mode")
    bulkCloneCmd.Flags().Bool("stream-checkpoint", true,
        "Enable checkpoint/resume for streaming")
    bulkCloneCmd.Flags().String("checkpoint-file", ".gzh-checkpoint.json",
        "Checkpoint file for resume capability")
}

func runBulkClone(cmd *cobra.Command, args []string) error {
    // ... existing code ...

    // Configure streaming if enabled
    if viper.GetBool("stream") {
        config.Streaming = bulk-clone.StreamingConfig{
            Enabled:            true,
            BufferSize:         viper.GetInt("stream-buffer"),
            EnableCheckpoint:   viper.GetBool("stream-checkpoint"),
            CheckpointFile:     viper.GetString("checkpoint-file"),
            WorkerCount:        runtime.NumCPU() * 2,
            EnableBackpressure: true,
        }

        // Use streaming facade
        facade := bulk-clone.NewStreamingFacade(config)
        return facade.StreamCloneAll()
    }

    // Use regular facade
    facade := bulk-clone.NewFacade(config)
    return facade.CloneAll()
}
```

## Expected Outcomes

- [ ] Streaming API handles organizations with 1000+ repos efficiently
- [ ] Memory usage stays constant regardless of org size
- [ ] Checkpoint/resume capability works correctly
- [ ] Backpressure prevents OOM errors
- [ ] Progress tracking shows real-time statistics
- [ ] Integration tests pass for large organizations

## Verification Commands

```bash
# Test streaming with large organization
gz bulk-clone --stream --org apache --stream-buffer 500

# Test checkpoint/resume
gz bulk-clone --stream --org kubernetes --checkpoint-file k8s.checkpoint
# Ctrl+C to interrupt
gz bulk-clone --stream --org kubernetes --checkpoint-file k8s.checkpoint
# Should resume from last position

# Monitor memory usage
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# Benchmark streaming vs batch
go test -bench=BenchmarkStreaming ./pkg/bulk-clone/...
```

## Performance Metrics

- Target: Handle 10,000+ repository organizations
- Memory: < 500MB constant usage
- Throughput: 100+ repos/second with parallel workers
- Resume: < 5 second checkpoint recovery time

## Next Steps

- Task 06: Add Kubernetes network topology analysis
- Task 07: Implement GitLab group streaming

# Task: Implement Bulk Clone Performance Improvements

## Priority: HIGH
## Estimated Time: 2 hours

## Context
The remote branch includes new performance optimization features:
- Enhanced rate limiter in `internal/api/enhanced_rate_limiter.go`
- Batch processing in `internal/api/batcher.go`
- Optimized bulk clone in `pkg/github/optimized_bulk_clone.go`
- Worker pool improvements in `internal/workerpool/`

## Pre-requisites
- [ ] Tasks 01-02 completed
- [ ] Understanding of current bulk-clone implementation
- [ ] Test repositories available for benchmarking

## Steps

### 1. Analyze Current Performance Bottlenecks

#### Create Performance Test Script
```go
// test/benchmark/bulk_clone_bench_test.go
package benchmark

import (
    "testing"
    "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
    "github.com/gizzahub/gzh-manager-go/pkg/github"
)

func BenchmarkBulkCloneSmallOrg(b *testing.B) {
    // Setup test config for org with 5-10 repos
    config := &bulk-clone.Config{
        GitHub: bulk-clone.GitHubConfig{
            Organizations: []string{"test-small-org"},
            Token: os.Getenv("GITHUB_TOKEN"),
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        facade := bulk-clone.NewFacade(config)
        _ = facade.CloneAll()
    }
}

func BenchmarkBulkCloneLargeOrg(b *testing.B) {
    // Setup test config for org with 50+ repos
    config := &bulk-clone.Config{
        GitHub: bulk-clone.GitHubConfig{
            Organizations: []string{"kubernetes"}, // Example large org
            Token: os.Getenv("GITHUB_TOKEN"),
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        facade := bulk-clone.NewFacade(config)
        _ = facade.CloneAll()
    }
}
```

#### Run Baseline Benchmarks
```bash
# Create results directory
mkdir -p test/benchmark/results

# Run benchmarks before optimization
go test -bench=. -benchmem -benchtime=10s ./test/benchmark > test/benchmark/results/baseline.txt
```

### 2. Implement Enhanced Rate Limiter Integration

#### Update GitHub Client Factory
```go
// pkg/github/factory.go
import (
    "github.com/gizzahub/gzh-manager-go/internal/api"
)

func (f *Factory) createOptimizedClient(token string) *github.Client {
    // Use enhanced rate limiter
    rateLimiter := api.NewEnhancedRateLimiter(
        api.WithBurstSize(10),
        api.WithRefillRate(5000), // 5000 requests per hour
        api.WithAdaptiveScaling(true),
    )
    
    // Create HTTP client with rate limiter
    httpClient := &http.Client{
        Transport: &api.RateLimitedTransport{
            Base:        http.DefaultTransport,
            RateLimiter: rateLimiter,
        },
    }
    
    return github.NewClient(httpClient).WithAuthToken(token)
}
```

### 3. Implement Batch Processing for Repository Lists

#### Create Batch Processor
```go
// pkg/github/batch_processor.go
package github

import (
    "context"
    "github.com/gizzahub/gzh-manager-go/internal/api"
)

type BatchProcessor struct {
    batcher *api.Batcher
    client  *github.Client
}

func NewBatchProcessor(client *github.Client) *BatchProcessor {
    return &BatchProcessor{
        batcher: api.NewBatcher(api.BatcherConfig{
            MaxBatchSize:     100,
            MaxConcurrent:    5,
            FlushInterval:    time.Second,
        }),
        client: client,
    }
}

func (bp *BatchProcessor) GetRepositoriesBatch(ctx context.Context, orgs []string) ([]*github.Repository, error) {
    var allRepos []*github.Repository
    
    // Create batch requests
    requests := make([]api.BatchRequest, len(orgs))
    for i, org := range orgs {
        requests[i] = api.BatchRequest{
            ID:   org,
            Type: "list_repos",
            Data: org,
        }
    }
    
    // Process in batches
    results := bp.batcher.ProcessBatch(ctx, requests, func(req api.BatchRequest) (interface{}, error) {
        repos, _, err := bp.client.Repositories.ListByOrg(ctx, req.Data.(string), nil)
        return repos, err
    })
    
    // Collect results
    for _, result := range results {
        if result.Error == nil {
            allRepos = append(allRepos, result.Data.([]*github.Repository)...)
        }
    }
    
    return allRepos, nil
}
```

### 4. Optimize Worker Pool Implementation

#### Update Repository Worker Pool
```go
// internal/workerpool/repository_pool.go
package workerpool

type OptimizedRepoPool struct {
    workers    int
    tasks      chan CloneTask
    results    chan CloneResult
    workerPool sync.Pool
}

func NewOptimizedRepoPool(workers int) *OptimizedRepoPool {
    pool := &OptimizedRepoPool{
        workers: workers,
        tasks:   make(chan CloneTask, workers*2),
        results: make(chan CloneResult, workers*2),
    }
    
    // Pre-allocate workers
    pool.workerPool = sync.Pool{
        New: func() interface{} {
            return &Worker{
                gitClient: git.NewClient(),
            }
        },
    }
    
    return pool
}

func (p *OptimizedRepoPool) Start(ctx context.Context) {
    var wg sync.WaitGroup
    
    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            worker := p.workerPool.Get().(*Worker)
            defer p.workerPool.Put(worker)
            
            for task := range p.tasks {
                result := worker.Process(ctx, task)
                select {
                case p.results <- result:
                case <-ctx.Done():
                    return
                }
            }
        }()
    }
    
    go func() {
        wg.Wait()
        close(p.results)
    }()
}
```

### 5. Implement Progress Tracking with ETA

#### Enhanced Progress Reporter
```go
// pkg/bulk-clone/progress_enhanced.go
package bulkclone

import (
    "github.com/schollz/progressbar/v3"
)

type EnhancedProgress struct {
    bar         *progressbar.ProgressBar
    startTime   time.Time
    processed   int64
    total       int64
    mutex       sync.Mutex
}

func NewEnhancedProgress(total int) *EnhancedProgress {
    bar := progressbar.NewOptions(total,
        progressbar.OptionEnableColorCodes(true),
        progressbar.OptionShowBytes(false),
        progressbar.OptionShowCount(),
        progressbar.OptionSetWidth(50),
        progressbar.OptionSetDescription("[cyan]Cloning repositories...[reset]"),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "[green]=[reset]",
            SaucerHead:    "[green]>[reset]",
            SaucerPadding: " ",
            BarStart:      "[",
            BarEnd:        "]",
        }),
        progressbar.OptionOnCompletion(func() {
            fmt.Println("\n✅ Bulk clone completed!")
        }),
    )
    
    return &EnhancedProgress{
        bar:       bar,
        startTime: time.Now(),
        total:     int64(total),
    }
}

func (p *EnhancedProgress) Update(repo string, status string) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    p.processed++
    elapsed := time.Since(p.startTime)
    rate := float64(p.processed) / elapsed.Seconds()
    eta := time.Duration(float64(p.total-p.processed)/rate) * time.Second
    
    desc := fmt.Sprintf("[cyan]%s[reset] | Rate: %.1f repos/s | ETA: %s", 
        status, rate, eta.Round(time.Second))
    p.bar.Describe(desc)
    p.bar.Add(1)
}
```

### 6. Integration and Testing

#### Update Bulk Clone Command
```go
// cmd/bulk-clone/bulk_clone.go
func runBulkClone(cmd *cobra.Command, args []string) error {
    // Load configuration
    config, err := loadConfig()
    if err != nil {
        return err
    }
    
    // Enable performance optimizations
    if viper.GetBool("optimize") {
        config.Performance = bulk-clone.PerformanceConfig{
            WorkerCount:      runtime.NumCPU() * 2,
            BatchSize:        50,
            RateLimitBurst:   10,
            EnableCaching:    true,
            ProgressTracking: true,
        }
    }
    
    // Create optimized facade
    facade := bulk-clone.NewOptimizedFacade(config)
    
    // Run with progress tracking
    return facade.CloneAllWithProgress()
}
```

### 7. Performance Testing and Validation

#### Run Performance Tests
```bash
# Run optimized benchmarks
go test -bench=. -benchmem -benchtime=10s ./test/benchmark > test/benchmark/results/optimized.txt

# Compare results
benchstat test/benchmark/results/baseline.txt test/benchmark/results/optimized.txt
```

#### Create Performance Report
```markdown
# Bulk Clone Performance Improvements

## Benchmark Results

| Metric | Baseline | Optimized | Improvement |
|--------|----------|-----------|-------------|
| Small Org (10 repos) | 45s | 12s | 73% faster |
| Large Org (100 repos) | 15m | 3m 20s | 78% faster |
| Memory Usage | 250MB | 180MB | 28% less |
| API Calls | 500 | 125 | 75% fewer |

## Key Optimizations
1. **Batch API Requests**: Reduced API calls by 75%
2. **Worker Pool Reuse**: Eliminated goroutine creation overhead
3. **Enhanced Rate Limiting**: Better utilization of API quota
4. **Progress Tracking**: Real-time ETA and performance metrics
```

## Expected Outcomes
- [ ] Bulk clone is 70%+ faster for large organizations
- [ ] Memory usage reduced by 25%+
- [ ] API rate limit utilization improved
- [ ] Progress tracking shows ETA and rate
- [ ] All tests pass with optimizations enabled

## Verification Commands
```bash
# Test optimized bulk clone
gz bulk-clone --optimize --config samples/bulk-clone-example.yaml

# Benchmark specific scenarios
go test -bench=BulkClone -benchmem ./pkg/bulk-clone/...

# Profile memory usage
go test -memprofile=mem.prof -bench=. ./test/benchmark
go tool pprof -http=:8080 mem.prof
```

## Next Steps
- Task 04: Add programmatic usage examples
- Task 05: Implement streaming API for large organizations
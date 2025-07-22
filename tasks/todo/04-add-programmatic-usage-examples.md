# Task: Add Programmatic Usage Examples for Library Users

## Priority: MEDIUM

## Estimated Time: 1.5 hours

## Context

The remote branch added `examples/programmatic-usage.go` showing how to use gzh-manager-go as a library. We need to expand this with comprehensive examples for all major features.

## Pre-requisites

- [ ] Task 01 completed (merged remote changes)
- [ ] Understanding of public API in `pkg/` directory
- [ ] Go modules understanding for import examples

## Steps

### 1. Analyze Current Example Structure

#### Review Existing Example

```bash
# Check current example file
cat examples/programmatic-usage.go

# Identify which packages are demonstrated
grep -E "gzh-manager-go/pkg" examples/programmatic-usage.go
```

### 2. Create Comprehensive Example Structure

#### Create Example Directory Structure

```bash
mkdir -p examples/{basic,advanced,integrations}
touch examples/README.md
```

#### Create Main Examples README

````markdown
# gzh-manager-go Examples

This directory contains examples of using gzh-manager-go as a library in your Go applications.

## Quick Start

```go
import "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
```
````

## Examples by Category

### Basic Usage

- [Bulk Clone Organizations](./basic/bulk_clone_simple.go)
- [GitHub API Client](./basic/github_client.go)
- [Configuration Loading](./basic/config_loader.go)

### Advanced Usage

- [Custom Rate Limiting](./advanced/rate_limiting.go)
- [Progress Tracking](./advanced/progress_tracking.go)
- [Error Handling](./advanced/error_handling.go)

### Integrations

- [CI/CD Integration](./integrations/cicd_example.go)
- [Webhook Handler](./integrations/webhook_server.go)
- [Cloud Provider Sync](./integrations/cloud_sync.go)

````

### 3. Create Basic Usage Examples

#### Basic Bulk Clone Example
```go
// examples/basic/bulk_clone_simple.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
)

func main() {
    // Create configuration
    config := &bulkclone.Config{
        GitHub: bulkclone.GitHubConfig{
            Organizations: []string{"kubernetes", "docker"},
            Token:        os.Getenv("GITHUB_TOKEN"),
            CloneOptions: bulkclone.CloneOptions{
                Strategy:  "pull", // or "reset", "fetch"
                Depth:     1,      // shallow clone
                Directory: "./repos",
            },
        },
    }

    // Create facade
    facade := bulkclone.NewFacade(config)

    // Set progress callback
    facade.OnProgress(func(event bulkclone.ProgressEvent) {
        fmt.Printf("[%s] %s: %s\n", event.Type, event.Repository, event.Message)
    })

    // Execute bulk clone
    ctx := context.Background()
    if err := facade.CloneAllContext(ctx); err != nil {
        log.Fatal(err)
    }

    fmt.Println("✅ Bulk clone completed successfully!")
}
````

#### GitHub Client Example

```go
// examples/basic/github_client.go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gizzahub/gzh-manager-go/pkg/github"
)

func main() {
    // Create factory
    factory := github.NewFactory()

    // Create client with token
    client, err := factory.CreateClient("github_token_here")
    if err != nil {
        log.Fatal(err)
    }

    // List repositories for an organization
    ctx := context.Background()
    repos, err := client.ListOrganizationRepos(ctx, "kubernetes")
    if err != nil {
        log.Fatal(err)
    }

    // Print repository information
    for _, repo := range repos {
        fmt.Printf("- %s: %s (⭐ %d)\n",
            repo.Name,
            repo.Description,
            repo.StargazersCount)
    }
}
```

### 4. Create Advanced Usage Examples

#### Custom Rate Limiting Example

```go
// examples/advanced/rate_limiting.go
package main

import (
    "context"
    "time"

    "github.com/gizzahub/gzh-manager-go/pkg/github"
    "github.com/gizzahub/gzh-manager-go/internal/api"
)

func main() {
    // Create custom rate limiter
    rateLimiter := api.NewEnhancedRateLimiter(
        api.WithBurstSize(30),              // Allow 30 concurrent requests
        api.WithRefillRate(5000),           // 5000 requests per hour
        api.WithAdaptiveScaling(true),      // Auto-adjust based on headers
        api.WithRetryBackoff(time.Second),  // Backoff strategy
    )

    // Create GitHub client with custom rate limiter
    factory := github.NewFactory(
        github.WithRateLimiter(rateLimiter),
        github.WithTimeout(30 * time.Second),
    )

    client, _ := factory.CreateClient("token")

    // Use client with automatic rate limiting
    ctx := context.Background()

    // This will automatically respect rate limits
    for i := 0; i < 100; i++ {
        repo, _, err := client.GetRepository(ctx, "kubernetes", "kubernetes")
        if err != nil {
            // Rate limiter will handle retries automatically
            continue
        }
        fmt.Printf("Fetched: %s\n", repo.FullName)
    }
}
```

#### Progress Tracking Example

```go
// examples/advanced/progress_tracking.go
package main

import (
    "fmt"
    "sync"

    bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
)

type ProgressTracker struct {
    mutex     sync.Mutex
    total     int
    completed int
    failed    int
    inProgress map[string]bool
}

func NewProgressTracker(total int) *ProgressTracker {
    return &ProgressTracker{
        total:      total,
        inProgress: make(map[string]bool),
    }
}

func (pt *ProgressTracker) HandleEvent(event bulkclone.ProgressEvent) {
    pt.mutex.Lock()
    defer pt.mutex.Unlock()

    switch event.Type {
    case bulkclone.EventTypeStart:
        pt.inProgress[event.Repository] = true
        fmt.Printf("🚀 Starting: %s\n", event.Repository)

    case bulkclone.EventTypeComplete:
        delete(pt.inProgress, event.Repository)
        pt.completed++
        percentage := float64(pt.completed) / float64(pt.total) * 100
        fmt.Printf("✅ Completed: %s [%.1f%%]\n", event.Repository, percentage)

    case bulkclone.EventTypeError:
        delete(pt.inProgress, event.Repository)
        pt.failed++
        fmt.Printf("❌ Failed: %s - %v\n", event.Repository, event.Error)

    case bulkclone.EventTypeProgress:
        fmt.Printf("📊 Progress: %s - %s\n", event.Repository, event.Message)
    }

    // Show summary
    if pt.completed+pt.failed == pt.total {
        pt.PrintSummary()
    }
}

func (pt *ProgressTracker) PrintSummary() {
    fmt.Println("\n=== Clone Summary ===")
    fmt.Printf("Total: %d\n", pt.total)
    fmt.Printf("Completed: %d (%.1f%%)\n",
        pt.completed,
        float64(pt.completed)/float64(pt.total)*100)
    fmt.Printf("Failed: %d (%.1f%%)\n",
        pt.failed,
        float64(pt.failed)/float64(pt.total)*100)
}

func main() {
    config := &bulkclone.Config{
        // ... configuration
    }

    facade := bulkclone.NewFacade(config)

    // Get total repository count
    repos, _ := facade.ListRepositories()
    tracker := NewProgressTracker(len(repos))

    // Set progress handler
    facade.OnProgress(tracker.HandleEvent)

    // Execute with tracking
    _ = facade.CloneAll()
}
```

### 5. Create Integration Examples

#### CI/CD Integration Example

```go
// examples/integrations/cicd_example.go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
    "github.com/gizzahub/gzh-manager-go/pkg/config"
)

// CI/CD Integration for automated repository synchronization
func main() {
    // Load configuration from CI environment
    configPath := os.Getenv("GZH_CONFIG_PATH")
    if configPath == "" {
        configPath = ".gzh-manager.yaml"
    }

    // Load and validate configuration
    loader := config.NewLoader()
    cfg, err := loader.LoadFile(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
        os.Exit(1)
    }

    // Create bulk clone facade
    facade := bulk-clone.NewFacade(cfg.BulkClone)

    // Set CI-friendly progress output (JSON format)
    facade.OnProgress(func(event bulk-clone.ProgressEvent) {
        output := map[string]interface{}{
            "timestamp": event.Timestamp,
            "type":      event.Type,
            "repo":      event.Repository,
            "message":   event.Message,
        }
        if event.Error != nil {
            output["error"] = event.Error.Error()
        }

        jsonOutput, _ := json.Marshal(output)
        fmt.Println(string(jsonOutput))
    })

    // Execute clone
    if err := facade.CloneAll(); err != nil {
        fmt.Fprintf(os.Stderr, "Clone failed: %v\n", err)
        os.Exit(1)
    }

    // Output summary for CI
    summary := map[string]interface{}{
        "status":      "success",
        "repos_count": len(facade.GetClonedRepos()),
        "duration":    facade.GetDuration().Seconds(),
    }
    jsonSummary, _ := json.Marshal(summary)
    fmt.Println(string(jsonSummary))
}
```

### 6. Create Test Examples

#### Create Example Tests

```go
// examples/basic/bulk_clone_simple_test.go
package main

import (
    "testing"
    "os"

    bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBulkCloneExample(t *testing.T) {
    // Skip if no token
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set")
    }

    // Create test configuration
    config := &bulkclone.Config{
        GitHub: bulkclone.GitHubConfig{
            Organizations: []string{"gizzahub"}, // Small test org
            Token:        token,
            CloneOptions: bulkclone.CloneOptions{
                Strategy:  "fetch",
                Directory: t.TempDir(),
            },
        },
    }

    // Create facade
    facade := bulkclone.NewFacade(config)

    // Track events
    var events []bulkclone.ProgressEvent
    facade.OnProgress(func(event bulkclone.ProgressEvent) {
        events = append(events, event)
    })

    // Execute
    err := facade.CloneAll()
    require.NoError(t, err)

    // Verify events were received
    assert.NotEmpty(t, events)

    // Verify repositories were cloned
    repos := facade.GetClonedRepos()
    assert.NotEmpty(t, repos)
}
```

### 7. Create go.mod for Examples

```go
// examples/go.mod
module github.com/gizzahub/gzh-manager-go/examples

go 1.23

require (
    github.com/gizzahub/gzh-manager-go v0.0.0-00010101000000-000000000000
    github.com/stretchr/testify v1.8.4
)

replace github.com/gizzahub/gzh-manager-go => ../
```

## Expected Outcomes

- [ ] Comprehensive examples directory structure created
- [ ] Basic usage examples for all major features
- [ ] Advanced examples showing performance optimizations
- [ ] Integration examples for CI/CD and automation
- [ ] All examples are executable and tested
- [ ] Examples have their own go.mod for easy testing

## Verification Commands

```bash
# Test all examples compile
cd examples
go build ./...

# Run example tests
go test ./...

# Run specific example
go run basic/bulk_clone_simple.go

# Verify imports work correctly
go mod tidy
go mod verify
```

## Documentation Updates

Add to main README.md:

````markdown
## Using as a Library

gzh-manager-go can be used as a library in your Go applications. See the [examples](./examples) directory for comprehensive usage examples.

```go
import "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
```
````

### Quick Example

[Include simple example here]

```

## Next Steps
- Task 05: Implement streaming API for large organizations
- Task 06: Add Kubernetes network topology analysis
```

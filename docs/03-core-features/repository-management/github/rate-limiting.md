# GitHub API Rate Limiting

This document describes the rate limiting implementation for the GitHub repository configuration management feature.

## Overview

The rate limiting system is designed to handle GitHub API rate limits gracefully, ensuring reliable operation even under heavy API usage. It implements:

- Primary rate limit tracking (5000 requests/hour for authenticated requests)
- Secondary rate limit handling (abuse detection)
- Automatic retry with exponential backoff
- Context-aware waiting and cancellation

## Architecture

### Components

1. **RateLimiter**: Core rate limiting logic with mutex-protected state
2. **RepoConfigClient**: GitHub API client with integrated rate limiting
3. **Retry Logic**: Exponential backoff with jitter for failed requests

### Rate Limit Headers

The system tracks the following GitHub API headers:

- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets
- `Retry-After`: Seconds to wait before retrying (for secondary limits)

## Implementation Details

### Primary Rate Limiting

```go
// Wait blocks until rate limit allows making a request
func (rl *RateLimiter) Wait(ctx context.Context) error
```

The Wait method:
1. Checks if we need to wait for retry-after (secondary limit)
2. Checks if we've exhausted the primary rate limit
3. Blocks until the rate limit resets or context is cancelled
4. Decrements the remaining counter

### Retry Logic

The system automatically retries requests that fail due to:
- Rate limit errors (429 Too Many Requests)
- Server errors (5xx status codes)
- Secondary rate limits (403 with specific headers)

Retry behavior:
- Maximum 3 retries per request
- Exponential backoff: 1s, 2s, 4s, 8s... (capped at 60s)
- 10% jitter added to prevent thundering herd
- Respects Retry-After header if present

### Example Usage

```go
client := github.NewRepoConfigClient(token)

// List repositories with automatic rate limiting
repos, err := client.ListRepositories(ctx, "myorg", &github.ListOptions{
    PerPage: 100,
})

// Check current rate limit status
remaining, limit, resetTime := client.GetRateLimitStatus()
fmt.Printf("Rate limit: %d/%d (resets at %s)\n", 
    remaining, limit, resetTime.Format(time.RFC3339))
```

## Error Handling

### Rate Limit Exhausted

When the primary rate limit is exhausted:
1. The system waits until the reset time
2. If context is cancelled during wait, returns context error
3. Automatically resumes operation after reset

### Secondary Rate Limits

GitHub may impose secondary rate limits for:
- Rapid creation of content
- Aggressive crawling behavior
- Concurrent requests from same user

The system handles these by:
1. Detecting 403 responses with X-GitHub-Request-Id header
2. Respecting Retry-After header
3. Backing off appropriately

## Best Practices

1. **Use Context**: Always pass a context with timeout to prevent indefinite waiting
2. **Monitor Rate Limits**: Check rate limit status periodically
3. **Batch Operations**: Group related API calls to minimize requests
4. **Cache Results**: Avoid redundant API calls by caching responses

## Configuration

The rate limiter is configured with sensible defaults:
- Initial limit: 5000 (GitHub's default for authenticated requests)
- Max retries: 3
- Max backoff: 60 seconds

These can be adjusted if needed:

```go
rateLimiter := github.NewRateLimiter()
rateLimiter.maxRetries = 5  // Increase retry attempts
```

## Testing

The implementation includes comprehensive tests for:
- Rate limit waiting behavior
- Context cancellation
- Retry logic with backoff
- Header parsing
- Error handling

Run tests with:
```bash
go test ./pkg/github/... -v
```

## Future Enhancements

1. **Predictive Waiting**: Pre-emptively slow down when approaching limits
2. **Multi-Account Support**: Rotate between multiple tokens
3. **Metrics Collection**: Export rate limit metrics for reporting
4. **Conditional Requests**: Use ETags to reduce API calls
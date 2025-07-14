// Package largescale provides efficient large-scale repository operations for GitHub.
//
// This package contains specialized implementations for handling bulk operations
// on thousands of repositories with optimized memory usage, adaptive rate limiting,
// and concurrent processing capabilities.
//
// Key features:
//   - Large-scale repository listing with pagination support
//   - Bulk repository cloning with memory-efficient batching
//   - Adaptive rate limiting based on GitHub API response headers
//   - Progress tracking and statistics for long-running operations
//   - Memory pressure monitoring and garbage collection management
//   - Retry mechanisms with exponential backoff
//
// Example usage:
//
//	config := largescale.DefaultLargeScaleConfig()
//	manager := largescale.NewLargeScaleManager(config, progressCallback)
//
//	repos, err := manager.ListAllRepositories(ctx, "organization")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = manager.BulkCloneRepositories(ctx, repos, "/target/path")
//	if err != nil {
//		log.Fatal(err)
//	}
package largescale

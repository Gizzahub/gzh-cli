// Package workerpool provides worker pool implementations and concurrency
// management utilities for the GZH Manager system.
//
// This package implements various worker pool patterns to efficiently
// manage concurrent operations, control resource usage, and provide
// scalable processing capabilities for different workloads.
//
// Key Components:
//
// Generic Worker Pool:
//   - Configurable number of workers
//   - Job queue management
//   - Worker lifecycle management
//   - Graceful shutdown handling
//
// Repository Pool:
//   - Specialized pool for repository operations
//   - Git operation optimization
//   - Resource-aware scheduling
//   - Repository-specific error handling
//
// Features:
//   - Dynamic worker scaling based on load
//   - Job priority and scheduling
//   - Worker health monitoring
//   - Performance metrics and monitoring
//   - Context-aware cancellation
//   - Memory and CPU usage optimization
//
// Pool Types:
//   - FixedPool: Fixed number of workers
//   - DynamicPool: Auto-scaling worker pool
//   - BoundedPool: Limited resource pool
//   - PriorityPool: Priority-based job processing
//
// Example usage:
//
//	pool := workerpool.NewPool(10)
//	defer pool.Close()
//
//	job := workerpool.NewJob(func() error {
//		return processRepository(repo)
//	})
//
//	err := pool.Submit(ctx, job)
//	result := <-job.Done()
//
// The package provides efficient concurrent processing while maintaining
// system stability and resource constraints.
package workerpool

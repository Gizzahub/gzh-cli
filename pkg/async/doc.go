// Package async provides asynchronous processing capabilities and
// concurrency utilities for the GZH Manager system.
//
// This package implements patterns for handling concurrent operations,
// background processing, and asynchronous I/O operations, enabling
// efficient resource utilization and responsive user experiences.
//
// Key Components:
//
// Work Queue:
//   - Concurrent job processing with worker pools
//   - Priority-based job scheduling
//   - Job retry and error handling
//   - Progress tracking and reporting
//
// Connection Management:
//   - HTTP connection pooling and reuse
//   - Connection health monitoring
//   - Automatic connection recovery
//   - Load balancing and failover
//
// Event Bus:
//   - Asynchronous event publishing and subscription
//   - Event routing and filtering
//   - Event persistence and replay
//   - Cross-component communication
//
// Async I/O:
//   - Non-blocking file operations
//   - Streaming data processing
//   - Buffered I/O with flow control
//   - Timeout and cancellation support
//
// Features:
//   - Context-aware cancellation
//   - Configurable concurrency limits
//   - Memory and resource management
//   - Comprehensive error handling
//   - Performance monitoring and metrics
//
// Example usage:
//
//	queue := async.NewWorkQueue(workers)
//	job := async.NewJob(handler, data)
//	err := queue.Submit(ctx, job)
//
//	bus := async.NewEventBus()
//	bus.Subscribe("repo.cloned", handler)
//	bus.Publish("repo.cloned", event)
//
// The package enables efficient concurrent processing while maintaining
// system stability and providing comprehensive monitoring capabilities.
package async

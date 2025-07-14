// Package recovery provides automated error recovery and resilience
// mechanisms for the GZH Manager system.
//
// This package implements intelligent recovery strategies that automatically
// handle transient failures, network issues, and system errors, ensuring
// robust operation even in challenging environments.
//
// Key Components:
//
// Circuit Breaker:
//   - Automatic failure detection and circuit opening
//   - Configurable failure thresholds and timeouts
//   - Gradual recovery and circuit closing
//   - Health check integration
//
// Retry Logic:
//   - Exponential backoff retry strategies
//   - Jittered retry timing to prevent thundering herd
//   - Configurable retry limits and conditions
//   - Context-aware cancellation
//
// Fallback Providers:
//   - Alternative service endpoints
//   - Cached response fallbacks
//   - Degraded mode operations
//   - Emergency operational procedures
//
// Recovery Orchestration:
//   - Automated recovery workflow execution
//   - Dependency-aware recovery ordering
//   - Recovery progress tracking
//   - Recovery metrics and reporting
//
// Features:
//   - Real-time system health monitoring
//   - Proactive failure detection
//   - Intelligent recovery decision making
//   - Recovery action logging and audit
//   - Performance impact minimization
//
// Example usage:
//
//	cb := recovery.NewCircuitBreaker(config)
//	result, err := cb.Execute(operation)
//
//	retry := recovery.NewRetryManager()
//	err = retry.ExecuteWithRetry(ctx, operation)
//
//	orchestrator := recovery.NewRecoveryOrchestrator()
//	err = orchestrator.ExecuteRecovery(scenario)
//
// The package provides comprehensive error recovery capabilities that
// enhance system reliability and reduce operational overhead.
package recovery

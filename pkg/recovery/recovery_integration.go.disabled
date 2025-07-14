package recovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/async"
	"github.com/gizzahub/gzh-manager-go/pkg/errors"
)

// RecoveryManager integrates recovery capabilities with existing systems
type RecoveryManager struct {
	orchestrator      *RecoveryOrchestrator
	connectionManager *async.ConnectionManager
	workQueue         *async.WorkQueue
}

// NewRecoveryManager creates a new recovery manager with integrated systems
func NewRecoveryManager() *RecoveryManager {
	orchestrator := NewRecoveryOrchestrator()

	// Register fallback providers
	orchestrator.RegisterFallbackProvider(NewNetworkFallbackProvider())
	orchestrator.RegisterFallbackProvider(NewFileFallbackProvider())
	orchestrator.RegisterFallbackProvider(NewAuthFallbackProvider())

	// Create connection manager with recovery integration
	connectionManager := async.NewConnectionManager(async.ConnectionConfig{
		MaxRetries:     3,
		BaseDelay:      1000,
		BackoffFactor:  2.0,
		JitterFactor:   0.1,
		MaxConnections: 100,
		IdleTimeout:    30000,
		RequestTimeout: 30000,
	})

	// Create work queue with recovery integration
	workQueue := async.NewWorkQueue(async.WorkQueueConfig{
		MaxWorkers:      10,
		BufferSize:      1000,
		MaxRetries:      3,
		RetryDelay:      1000,
		ProcessTimeout:  30000,
		EnableMetrics:   true,
		BackoffStrategy: "exponential",
	})

	return &RecoveryManager{
		orchestrator:      orchestrator,
		connectionManager: connectionManager,
		workQueue:         workQueue,
	}
}

// ExecuteWithRecovery executes an operation with full recovery support
func (rm *RecoveryManager) ExecuteWithRecovery(ctx context.Context, operation func() error) error {
	return rm.executeWithRecoveryInternal(ctx, operation, 0)
}

// executeWithRecoveryInternal is the internal implementation with recursion protection
func (rm *RecoveryManager) executeWithRecoveryInternal(ctx context.Context, operation func() error, depth int) error {
	// Prevent infinite recursion
	if depth > 3 {
		return fmt.Errorf("maximum recovery depth exceeded")
	}

	// Try the operation first
	err := operation()
	if err == nil {
		return nil
	}

	// Convert to UserError if not already
	var userErr *errors.UserError
	if ue, ok := err.(*errors.UserError); ok {
		userErr = ue
	} else {
		// Create UserError from generic error
		userErr = rm.createUserErrorFromGeneric(err)
	}

	// Attempt recovery
	recoveryErr := rm.orchestrator.RecoverFromError(ctx, userErr, func() error {
		return rm.executeWithRecoveryInternal(ctx, operation, depth+1)
	})

	if recoveryErr == nil {
		return nil
	}

	// If recovery failed, return original error with recovery context
	return fmt.Errorf("operation failed and recovery unsuccessful: original error: %w, recovery error: %v", err, recoveryErr)
}

// ExecuteHTTPWithRecovery executes HTTP operations with connection manager and recovery
func (rm *RecoveryManager) ExecuteHTTPWithRecovery(ctx context.Context, url string, options async.RequestOptions) (*async.Response, error) {
	var response *async.Response
	var err error

	operation := func() error {
		response, err = rm.connectionManager.DoRequest(ctx, url, options)
		return err
	}

	recoveryErr := rm.ExecuteWithRecovery(ctx, operation)
	if recoveryErr != nil {
		return nil, recoveryErr
	}

	return response, nil
}

// SubmitJobWithRecovery submits a job to work queue with recovery support
func (rm *RecoveryManager) SubmitJobWithRecovery(ctx context.Context, job async.Job) error {
	// Wrap job execution with recovery
	wrappedJob := async.Job{
		ID:       job.ID,
		Type:     job.Type,
		Priority: job.Priority,
		Data:     job.Data,
		Handler: func(ctx context.Context, job async.Job) error {
			return rm.ExecuteWithRecovery(ctx, func() error {
				return job.Handler(ctx, job)
			})
		},
		OnSuccess: job.OnSuccess,
		OnFailure: func(job async.Job, err error) {
			// Enhanced failure handling with recovery context
			if job.OnFailure != nil {
				job.OnFailure(job, err)
			}

			// Log recovery metrics
			rm.logRecoveryMetrics(job.ID, err)
		},
		MaxRetries: job.MaxRetries,
		Context:    job.Context,
	}

	return rm.workQueue.Submit(ctx, wrappedJob)
}

// GetRecoveryStatus returns comprehensive recovery status
func (rm *RecoveryManager) GetRecoveryStatus() RecoveryStatus {
	orchestratorMetrics := rm.orchestrator.GetMetrics()
	circuitBreakerMetrics := rm.orchestrator.GetCircuitBreakerMetrics()
	connectionMetrics := rm.connectionManager.GetMetrics()
	workQueueMetrics := rm.workQueue.GetMetrics()

	return RecoveryStatus{
		OrchestratorMetrics:   orchestratorMetrics,
		CircuitBreakerMetrics: circuitBreakerMetrics,
		ConnectionMetrics:     connectionMetrics,
		WorkQueueMetrics:      workQueueMetrics,
		OverallHealth:         rm.calculateOverallHealth(orchestratorMetrics, circuitBreakerMetrics),
	}
}

// RecoveryStatus provides comprehensive recovery system status
type RecoveryStatus struct {
	OrchestratorMetrics   RecoveryMetrics                  `json:"orchestrator_metrics"`
	CircuitBreakerMetrics map[string]CircuitBreakerMetrics `json:"circuit_breaker_metrics"`
	ConnectionMetrics     async.ConnectionMetrics          `json:"connection_metrics"`
	WorkQueueMetrics      async.WorkQueueMetrics           `json:"work_queue_metrics"`
	OverallHealth         HealthStatus                     `json:"overall_health"`
}

// HealthStatus represents system health
type HealthStatus struct {
	Status      string   `json:"status"` // "healthy", "degraded", "unhealthy"
	Score       float64  `json:"score"`  // 0.0 - 1.0
	Issues      []string `json:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// calculateOverallHealth calculates overall system health
func (rm *RecoveryManager) calculateOverallHealth(orchestratorMetrics RecoveryMetrics, circuitBreakerMetrics map[string]CircuitBreakerMetrics) HealthStatus {
	score := 1.0
	var issues []string
	var suggestions []string

	// Check recovery success rate
	if orchestratorMetrics.TotalAttempts > 0 {
		successRate := float64(orchestratorMetrics.SuccessfulRecoveries) / float64(orchestratorMetrics.TotalAttempts)
		if successRate < 0.7 {
			score -= 0.3
			issues = append(issues, "Low recovery success rate")
			suggestions = append(suggestions, "Review recovery policies and increase timeout values")
		}
	}

	// Check circuit breaker states
	openCircuits := 0
	for name, metrics := range circuitBreakerMetrics {
		if metrics.State == StateOpen {
			openCircuits++
			issues = append(issues, fmt.Sprintf("Circuit breaker %s is OPEN", name))
			suggestions = append(suggestions, fmt.Sprintf("Investigate issues with %s service", name))
		}
	}

	if openCircuits > 0 {
		score -= float64(openCircuits) * 0.2
	}

	// Determine status
	status := "healthy"
	if score < 0.3 {
		status = "unhealthy"
	} else if score < 0.7 {
		status = "degraded"
	}

	return HealthStatus{
		Status:      status,
		Score:       score,
		Issues:      issues,
		Suggestions: suggestions,
	}
}

// createUserErrorFromGeneric converts a generic error to UserError
func (rm *RecoveryManager) createUserErrorFromGeneric(err error) *errors.UserError {
	errStr := err.Error()

	// Determine error category based on error message
	var domain, category, code string

	if containsAny(errStr, []string{"timeout", "deadline exceeded"}) {
		domain, category, code = "network", "timeout", "operation_timeout"
	} else if containsAny(errStr, []string{"connection refused", "no route to host"}) {
		domain, category, code = "network", "connection", "connection_failed"
	} else if containsAny(errStr, []string{"permission denied", "access denied"}) {
		domain, category, code = "file", "permission", "access_denied"
	} else if containsAny(errStr, []string{"not found", "no such file"}) {
		domain, category, code = "file", "not_found", "file_not_found"
	} else if containsAny(errStr, []string{"invalid token", "unauthorized"}) {
		domain, category, code = "auth", "invalid", "invalid_token"
	} else {
		domain, category, code = "generic", "unknown", "unknown_error"
	}

	errorCode := &errors.ErrorCode{
		Domain:   domain,
		Category: category,
		Code:     code,
	}

	return &errors.UserError{
		Code:        *errorCode,
		Message:     errStr,
		Description: "Automatically generated error from generic error",
		Context:     map[string]interface{}{"original_error": errStr},
		RequestID:   fmt.Sprintf("auto_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Suggestions: []string{"Check system logs for more details"},
	}
}

// logRecoveryMetrics logs recovery metrics for analysis
func (rm *RecoveryManager) logRecoveryMetrics(jobID string, err error) {
	// In a real implementation, this would send metrics to monitoring system
	fmt.Printf("Recovery metrics for job %s: error=%v\n", jobID, err)
}

// Shutdown gracefully shuts down all recovery components
func (rm *RecoveryManager) Shutdown() {
	rm.orchestrator.Shutdown()
	rm.workQueue.Shutdown()
}

// containsAny checks if string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	sLower := strings.ToLower(s)
	for _, substring := range substrings {
		if strings.Contains(sLower, strings.ToLower(substring)) {
			return true
		}
	}
	return false
}

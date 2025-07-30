// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// ValidationResult represents the result of a container validation.
type ValidationResult struct {
	IsValid    bool                 `json:"is_valid"`
	Errors     []ValidationError    `json:"errors,omitempty"`
	Warnings   []ValidationWarning  `json:"warnings,omitempty"`
	Statistics ValidationStatistics `json:"statistics"`
	Duration   time.Duration        `json:"duration"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	DependencyName string `json:"dependency_name"`
	ErrorType      string `json:"error_type"`
	Message        string `json:"message"`
	Suggestion     string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	DependencyName string `json:"dependency_name"`
	WarningType    string `json:"warning_type"`
	Message        string `json:"message"`
}

// ValidationStatistics provides statistics about the validation.
type ValidationStatistics struct {
	TotalDependencies    int `json:"total_dependencies"`
	ValidDependencies    int `json:"valid_dependencies"`
	ErrorCount           int `json:"error_count"`
	WarningCount         int `json:"warning_count"`
	CircularDependencies int `json:"circular_dependencies"`
}

// ContainerValidator validates container configuration and dependencies.
type ContainerValidator struct {
	container *Container
	context   context.Context
	timeout   time.Duration
}

// NewContainerValidator creates a new container validator.
func NewContainerValidator(container *Container) *ContainerValidator {
	return &ContainerValidator{
		container: container,
		context:   context.Background(),
		timeout:   30 * time.Second,
	}
}

// WithContext sets the validation context.
func (v *ContainerValidator) WithContext(ctx context.Context) *ContainerValidator {
	v.context = ctx
	return v
}

// WithTimeout sets the validation timeout.
func (v *ContainerValidator) WithTimeout(timeout time.Duration) *ContainerValidator {
	v.timeout = timeout
	return v
}

// Validate performs comprehensive validation of the container.
func (v *ContainerValidator) Validate() *ValidationResult {
	startTime := time.Now()

	result := &ValidationResult{
		IsValid:    true,
		Errors:     make([]ValidationError, 0),
		Warnings:   make([]ValidationWarning, 0),
		Statistics: ValidationStatistics{},
		Duration:   0,
	}

	ctx, cancel := context.WithTimeout(v.context, v.timeout)
	defer cancel()

	// Get list of registered dependencies
	dependencies := v.container.ListRegistered()
	result.Statistics.TotalDependencies = len(dependencies)

	// Validate each dependency
	for _, depName := range dependencies {
		v.validateDependency(ctx, depName, result)
	}

	// Check for circular dependencies
	v.checkCircularDependencies(result)

	// Validate interface compliance
	v.validateInterfaceCompliance(ctx, result)

	// Set final statistics
	result.Statistics.ErrorCount = len(result.Errors)
	result.Statistics.WarningCount = len(result.Warnings)
	result.Statistics.ValidDependencies = result.Statistics.TotalDependencies - result.Statistics.ErrorCount
	result.IsValid = result.Statistics.ErrorCount == 0
	result.Duration = time.Since(startTime)

	return result
}

// validateDependency validates a single dependency.
func (v *ContainerValidator) validateDependency(ctx context.Context, depName string, result *ValidationResult) {
	// Create a timeout for this specific dependency
	depCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Channel to receive the result
	done := make(chan error, 1)

	go func() {
		_, err := v.container.Get(depName)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				DependencyName: depName,
				ErrorType:      "creation_failed",
				Message:        fmt.Sprintf("Failed to create dependency: %v", err),
				Suggestion:     "Check factory function implementation and dependencies",
			})
		}
	case <-depCtx.Done():
		result.Errors = append(result.Errors, ValidationError{
			DependencyName: depName,
			ErrorType:      "timeout",
			Message:        "Dependency creation timed out",
			Suggestion:     "Check for slow initialization or circular dependencies",
		})
	}
}

// checkCircularDependencies detects circular dependency patterns.
func (v *ContainerValidator) checkCircularDependencies(result *ValidationResult) {
	// This is a simplified circular dependency check
	// A more sophisticated implementation would track dependency graphs
	dependencies := v.container.ListRegistered()

	// Check for obvious circular patterns by examining factory functions
	for _, depName := range dependencies {
		if v.hasCircularReference(depName, make(map[string]bool)) {
			result.Statistics.CircularDependencies++
			result.Warnings = append(result.Warnings, ValidationWarning{
				DependencyName: depName,
				WarningType:    "potential_circular_dependency",
				Message:        "Potential circular dependency detected",
			})
		}
	}
}

// hasCircularReference checks if a dependency has circular references.
func (v *ContainerValidator) hasCircularReference(depName string, visited map[string]bool) bool {
	if visited[depName] {
		return true
	}

	visited[depName] = true
	// This is a placeholder - actual implementation would need to parse factory functions
	// to detect dependencies between them
	delete(visited, depName)
	return false
}

// validateInterfaceCompliance validates that dependencies implement expected interfaces.
func (v *ContainerValidator) validateInterfaceCompliance(ctx context.Context, result *ValidationResult) {
	// Define expected interface compliance
	expectedInterfaces := map[string]reflect.Type{
		"logger": reflect.TypeOf((*interface{ Log(string) })(nil)).Elem(),
		"env":    reflect.TypeOf((*interface{ Get(string) string })(nil)).Elem(),
		"providerRegistry": reflect.TypeOf((*interface {
			GetProvider(string) (interface{}, error)
		})(nil)).Elem(),
	}

	for depName, expectedType := range expectedInterfaces {
		if v.container.Has(depName) {
			instance, err := v.container.Get(depName)
			if err != nil {
				continue // Already reported in dependency validation
			}

			instanceType := reflect.TypeOf(instance)
			if !instanceType.Implements(expectedType) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					DependencyName: depName,
					WarningType:    "interface_compliance",
					Message:        fmt.Sprintf("Dependency does not implement expected interface: %v", expectedType),
				})
			}
		}
	}
}

// HealthChecker provides health checking capabilities for the container.
type HealthChecker struct {
	container *Container
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(container *Container) *HealthChecker {
	return &HealthChecker{container: container}
}

// HealthStatus represents the health status of the container.
type HealthStatus struct {
	Status       string                      `json:"status"`
	Timestamp    time.Time                   `json:"timestamp"`
	Dependencies map[string]DependencyHealth `json:"dependencies"`
	Summary      HealthSummary               `json:"summary"`
}

// DependencyHealth represents the health of a single dependency.
type DependencyHealth struct {
	Status      string        `json:"status"`
	Message     string        `json:"message,omitempty"`
	Latency     time.Duration `json:"latency"`
	LastChecked time.Time     `json:"last_checked"`
}

// HealthSummary provides summary statistics.
type HealthSummary struct {
	TotalChecked int `json:"total_checked"`
	Healthy      int `json:"healthy"`
	Unhealthy    int `json:"unhealthy"`
	Unknown      int `json:"unknown"`
}

// CheckHealth performs a health check of all dependencies.
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Timestamp:    time.Now(),
		Dependencies: make(map[string]DependencyHealth),
		Summary:      HealthSummary{},
	}

	dependencies := h.container.ListRegistered()
	status.Summary.TotalChecked = len(dependencies)

	for _, depName := range dependencies {
		depHealth := h.checkDependencyHealth(ctx, depName)
		status.Dependencies[depName] = depHealth

		switch depHealth.Status {
		case "healthy":
			status.Summary.Healthy++
		case "unhealthy":
			status.Summary.Unhealthy++
		default:
			status.Summary.Unknown++
		}
	}

	// Determine overall status
	if status.Summary.Unhealthy > 0 {
		status.Status = "unhealthy"
	} else if status.Summary.Unknown > 0 {
		status.Status = "degraded"
	} else {
		status.Status = "healthy"
	}

	return status
}

// checkDependencyHealth checks the health of a single dependency.
func (h *HealthChecker) checkDependencyHealth(ctx context.Context, depName string) DependencyHealth {
	startTime := time.Now()

	depHealth := DependencyHealth{
		Status:      "unknown",
		LastChecked: startTime,
	}

	// Try to create/get the dependency
	_, err := h.container.Get(depName)
	latency := time.Since(startTime)
	depHealth.Latency = latency

	if err != nil {
		depHealth.Status = "unhealthy"
		depHealth.Message = err.Error()
	} else if latency > 5*time.Second {
		depHealth.Status = "unhealthy"
		depHealth.Message = "slow response time"
	} else {
		depHealth.Status = "healthy"
	}

	return depHealth
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// EnvironmentSwitcher handles switching between different development environments
type EnvironmentSwitcher struct {
	serviceSwitchers map[string]ServiceSwitcher
	progressCallback func(SwitchProgress)
	mu               sync.RWMutex
}

// NewEnvironmentSwitcher creates a new environment switcher
func NewEnvironmentSwitcher() *EnvironmentSwitcher {
	return &EnvironmentSwitcher{
		serviceSwitchers: make(map[string]ServiceSwitcher),
	}
}

// RegisterServiceSwitcher registers a service switcher
func (es *EnvironmentSwitcher) RegisterServiceSwitcher(name string, switcher ServiceSwitcher) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.serviceSwitchers[name] = switcher
}

// SetProgressCallback sets the progress callback function
func (es *EnvironmentSwitcher) SetProgressCallback(callback func(SwitchProgress)) {
	es.progressCallback = callback
}

// SwitchOptions contains options for environment switching
type SwitchOptions struct {
	DryRun          bool
	Force           bool
	Parallel        bool
	RollbackOnError bool
	Timeout         time.Duration
}

// SwitchEnvironment switches to the specified environment
func (es *EnvironmentSwitcher) SwitchEnvironment(ctx context.Context, env *Environment, options SwitchOptions) (*SwitchResult, error) {
	startTime := time.Now()

	// Validate environment
	if err := env.Validate(); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	// Resolve dependencies
	resolver := NewDependencyResolver(env.Services, env.Dependencies)
	groups, err := resolver.GetParallelGroups()
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Initialize result
	result := &SwitchResult{
		Success:          true,
		SwitchedServices: []string{},
		FailedServices:   []string{},
		Errors:           []SwitchError{},
	}

	// Store current states for rollback
	previousStates := make(map[string]interface{})

	// Execute pre-hooks
	if err := es.executeHooks(ctx, env.PreHooks, "pre-hook"); err != nil {
		return &SwitchResult{
			Success:  false,
			Duration: time.Since(startTime),
			Errors:   []SwitchError{{Service: "pre-hook", Error: err.Error(), Time: time.Now()}},
		}, err
	}

	totalServices := len(env.Services)
	completedServices := 0

	// Process each dependency group
	for _, group := range groups {
		if options.Parallel && len(group.Services) > 1 {
			// Parallel execution within the group
			if err := es.switchServicesParallel(ctx, env, group.Services, previousStates, result, options); err != nil {
				if options.RollbackOnError {
					es.rollbackServices(ctx, previousStates, result)
				}
				result.Success = false
				result.Duration = time.Since(startTime)
				return result, err
			}
		} else {
			// Sequential execution within the group
			for _, serviceName := range group.Services {
				if err := es.switchSingleService(ctx, env, serviceName, previousStates, result, options); err != nil {
					if options.RollbackOnError {
						es.rollbackServices(ctx, previousStates, result)
					}
					result.Success = false
					result.Duration = time.Since(startTime)
					return result, err
				}
			}
		}

		completedServices += len(group.Services)

		// Report progress
		if es.progressCallback != nil {
			progress := SwitchProgress{
				TotalServices:     totalServices,
				CompletedServices: completedServices,
				Status:            fmt.Sprintf("Completed group %d", group.Level),
				StartTime:         startTime,
				EstimatedEnd:      startTime.Add(time.Duration(float64(time.Since(startTime)) * float64(totalServices) / float64(completedServices))),
			}
			es.progressCallback(progress)
		}
	}

	// Execute post-hooks
	if err := es.executeHooks(ctx, env.PostHooks, "post-hook"); err != nil {
		// Post-hook failures don't trigger rollback but are recorded
		result.Errors = append(result.Errors, SwitchError{
			Service: "post-hook",
			Error:   err.Error(),
			Time:    time.Now(),
		})
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// switchSingleService switches a single service
func (es *EnvironmentSwitcher) switchSingleService(ctx context.Context, env *Environment, serviceName string, previousStates map[string]interface{}, result *SwitchResult, options SwitchOptions) error {
	es.mu.RLock()
	switcher, exists := es.serviceSwitchers[serviceName]
	es.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no switcher registered for service: %s", serviceName)
	}

	serviceConfig, exists := env.Services[serviceName]
	if !exists {
		return fmt.Errorf("service configuration not found: %s", serviceName)
	}

	// Get current state for rollback
	currentState, err := switcher.GetCurrentState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current state for %s: %w", serviceName, err)
	}
	previousStates[serviceName] = currentState

	// Get service-specific configuration
	var config interface{}
	switch serviceName {
	case "aws":
		config = serviceConfig.AWS
	case "gcp":
		config = serviceConfig.GCP
	case "azure":
		config = serviceConfig.Azure
	case "docker":
		config = serviceConfig.Docker
	case "kubernetes":
		config = serviceConfig.Kubernetes
	case "ssh":
		config = serviceConfig.SSH
	default:
		return fmt.Errorf("unknown service type: %s", serviceName)
	}

	if config == nil {
		return fmt.Errorf("no configuration provided for service: %s", serviceName)
	}

	// Perform the switch (skip in dry-run mode)
	if !options.DryRun {
		if err := switcher.Switch(ctx, config); err != nil {
			result.FailedServices = append(result.FailedServices, serviceName)
			result.Errors = append(result.Errors, SwitchError{
				Service: serviceName,
				Error:   err.Error(),
				Time:    time.Now(),
			})
			return fmt.Errorf("failed to switch %s: %w", serviceName, err)
		}
	}

	result.SwitchedServices = append(result.SwitchedServices, serviceName)
	return nil
}

// switchServicesParallel switches multiple services in parallel
func (es *EnvironmentSwitcher) switchServicesParallel(ctx context.Context, env *Environment, serviceNames []string, previousStates map[string]interface{}, result *SwitchResult, options SwitchOptions) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(serviceNames))

	for _, serviceName := range serviceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := es.switchSingleService(ctx, env, name, previousStates, result, options); err != nil {
				errChan <- err
			}
		}(serviceName)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("parallel switch failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// rollbackServices rolls back services to their previous states
func (es *EnvironmentSwitcher) rollbackServices(ctx context.Context, previousStates map[string]interface{}, result *SwitchResult) {
	var rollbackErrors []string

	for serviceName, previousState := range previousStates {
		es.mu.RLock()
		switcher, exists := es.serviceSwitchers[serviceName]
		es.mu.RUnlock()

		if !exists {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("no switcher for %s", serviceName))
			continue
		}

		if err := switcher.Rollback(ctx, previousState); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	result.RollbackPerformed = true
	if len(rollbackErrors) > 0 {
		result.Errors = append(result.Errors, SwitchError{
			Service: "rollback",
			Error:   strings.Join(rollbackErrors, "; "),
			Time:    time.Now(),
		})
	}
}

// executeHooks executes pre or post hooks
func (es *EnvironmentSwitcher) executeHooks(ctx context.Context, hooks []Hook, hookType string) error {
	for i, hook := range hooks {
		if err := es.executeHook(ctx, hook, fmt.Sprintf("%s-%d", hookType, i)); err != nil {
			if hook.OnError == "continue" {
				continue
			}
			return fmt.Errorf("hook execution failed: %w", err)
		}
	}
	return nil
}

// executeHook executes a single hook
func (es *EnvironmentSwitcher) executeHook(ctx context.Context, hook Hook, hookName string) error {
	timeout := hook.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // default timeout
	}

	hookCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(hookCtx, "sh", "-c", hook.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook '%s' failed: %w (output: %s)", hookName, err, string(output))
	}

	return nil
}

// GetAvailableServices returns a list of available service switchers
func (es *EnvironmentSwitcher) GetAvailableServices() []string {
	es.mu.RLock()
	defer es.mu.RUnlock()

	services := make([]string, 0, len(es.serviceSwitchers))
	for name := range es.serviceSwitchers {
		services = append(services, name)
	}
	return services
}

// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DockerChecker implements ServiceChecker for Docker.
type DockerChecker struct{}

// NewDockerChecker creates a new Docker status checker.
func NewDockerChecker() *DockerChecker {
	return &DockerChecker{}
}

// Name returns the service name.
func (d *DockerChecker) Name() string {
	return "docker"
}

// CheckStatus checks Docker current status.
func (d *DockerChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "docker",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if Docker CLI is available
	if !d.isDockerAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "Docker CLI not found"
		return status, nil
	}

	// Check if Docker daemon is running
	if !d.isDockerDaemonRunning(ctx) {
		status.Status = StatusInactive
		status.Details["error"] = "Docker daemon not running"
		return status, nil
	}

	// Get current context
	context, err := d.getCurrentContext(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["error"] = fmt.Sprintf("Failed to get Docker context: %v", err)
		return status, nil
	}

	status.Current.Context = context
	status.Status = StatusActive

	// Docker doesn't typically have credential expiration like cloud services
	status.Credentials = CredentialStatus{
		Valid: true,
		Type:  "docker-socket",
	}

	return status, nil
}

// CheckHealth performs detailed health check for Docker.
func (d *DockerChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Test Docker connectivity with docker info
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to connect to Docker daemon: %v", err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "Docker daemon is running and accessible"
	health.Details["server_version"] = strings.TrimSpace(string(output))

	// Get additional Docker info
	cmd = exec.CommandContext(ctx, "docker", "system", "df", "--format", "table")
	dfOutput, err := cmd.Output()
	if err == nil {
		health.Details["disk_usage"] = string(dfOutput)
	}

	// Check running containers count
	cmd = exec.CommandContext(ctx, "docker", "ps", "-q")
	psOutput, err := cmd.Output()
	if err == nil {
		containerCount := len(strings.Split(strings.TrimSpace(string(psOutput)), "\n"))
		if strings.TrimSpace(string(psOutput)) == "" {
			containerCount = 0
		}
		health.Details["running_containers"] = containerCount
	}

	return health, nil
}

// isDockerAvailable checks if Docker CLI is installed.
func (d *DockerChecker) isDockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// isDockerDaemonRunning checks if Docker daemon is running.
func (d *DockerChecker) isDockerDaemonRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "info")
	err := cmd.Run()
	return err == nil
}

// getCurrentContext gets the current Docker context.
func (d *DockerChecker) getCurrentContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "context", "show")
	output, err := cmd.Output()
	if err != nil {
		// If context command fails, assume default context
		return awsDefaultProfile, nil
	}
	return strings.TrimSpace(string(output)), nil
}

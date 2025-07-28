// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SSHChecker implements ServiceChecker for SSH
type SSHChecker struct{}

// NewSSHChecker creates a new SSH status checker
func NewSSHChecker() *SSHChecker {
	return &SSHChecker{}
}

// Name returns the service name
func (s *SSHChecker) Name() string {
	return "ssh"
}

// CheckStatus checks SSH current status
func (s *SSHChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "ssh",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if SSH is available
	if !s.isSSHAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "SSH not found"
		return status, nil
	}

	// Check SSH agent status
	agentStatus := s.checkSSHAgent()
	if !agentStatus {
		status.Status = StatusInactive
		status.Details["error"] = "SSH agent not running"
		return status, nil
	}

	// Get loaded keys
	keys, err := s.getLoadedKeys(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["error"] = fmt.Sprintf("Failed to get SSH keys: %v", err)
		return status, nil
	}

	if len(keys) == 0 {
		status.Status = StatusInactive
		status.Details["error"] = "No SSH keys loaded"
		return status, nil
	}

	status.Status = StatusActive
	status.Current.Context = fmt.Sprintf("%d keys loaded", len(keys))

	// Check SSH key validity
	credStatus := s.checkSSHKeys(keys)
	status.Credentials = *credStatus

	return status, nil
}

// CheckHealth performs detailed health check for SSH
func (s *SSHChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Check SSH agent connectivity
	cmd := exec.CommandContext(ctx, "ssh-add", "-l")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to connect to SSH agent: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "SSH agent is running with loaded keys"
	health.Details["loaded_keys"] = string(output)

	// Check SSH config file
	configPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	if _, err := os.Stat(configPath); err == nil {
		health.Details["config_file"] = configPath
	}

	return health, nil
}

// isSSHAvailable checks if SSH is installed
func (s *SSHChecker) isSSHAvailable() bool {
	_, err := exec.LookPath("ssh")
	return err == nil
}

// checkSSHAgent checks if SSH agent is running
func (s *SSHChecker) checkSSHAgent() bool {
	// Check SSH_AUTH_SOCK environment variable
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return false
	}

	// Try to connect to SSH agent
	cmd := exec.Command("ssh-add", "-l")
	err := cmd.Run()
	// ssh-add -l returns 0 if keys are loaded, 1 if no keys, 2 if agent not running
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode() != 2
	}
	return err == nil
}

// getLoadedKeys gets the list of loaded SSH keys
func (s *SSHChecker) getLoadedKeys(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "ssh-add", "-l")
	output, err := cmd.Output()
	if err != nil {
		// Check if it's "no keys loaded" vs actual error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil // No keys loaded, but agent is running
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var keys []string
	for _, line := range lines {
		if line != "" {
			keys = append(keys, line)
		}
	}

	return keys, nil
}

// checkSSHKeys checks the status of SSH keys
func (s *SSHChecker) checkSSHKeys(keys []string) *CredentialStatus {
	credStatus := &CredentialStatus{
		Valid: len(keys) > 0,
		Type:  "ssh-keys",
	}

	if len(keys) == 0 {
		credStatus.Warning = "No SSH keys loaded"
		return credStatus
	}

	// Check for common key types and potential issues
	hasRSA := false
	hasEd25519 := false
	for _, key := range keys {
		if strings.Contains(key, "RSA") {
			hasRSA = true
		}
		if strings.Contains(key, "ED25519") {
			hasEd25519 = true
		}
	}

	if hasRSA && !hasEd25519 {
		credStatus.Warning = "Consider using Ed25519 keys for better security"
	}

	return credStatus
}

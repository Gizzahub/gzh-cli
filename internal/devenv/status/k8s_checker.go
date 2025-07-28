// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// KubernetesChecker implements ServiceChecker for Kubernetes
type KubernetesChecker struct{}

// NewKubernetesChecker creates a new Kubernetes status checker
func NewKubernetesChecker() *KubernetesChecker {
	return &KubernetesChecker{}
}

// Name returns the service name
func (k *KubernetesChecker) Name() string {
	return "kubernetes"
}

// CheckStatus checks Kubernetes current status
func (k *KubernetesChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "kubernetes",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if kubectl is available
	if !k.isKubectlAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "kubectl not found"
		return status, nil
	}

	// Get current context
	context, err := k.getCurrentContext(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["error"] = fmt.Sprintf("Failed to get current context: %v", err)
		return status, nil
	}

	if context == "" {
		status.Status = StatusInactive
		status.Details["error"] = "No Kubernetes context set"
		return status, nil
	}

	status.Current.Context = context

	// Get current namespace
	namespace, err := k.getCurrentNamespace(ctx)
	if err == nil {
		status.Current.Namespace = namespace
	}

	// Check cluster connectivity
	credStatus, err := k.checkClusterAccess(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["connectivity_error"] = err.Error()
		return status, nil
	}

	status.Credentials = *credStatus
	if credStatus.Valid {
		status.Status = StatusActive
	} else {
		status.Status = StatusInactive
	}

	return status, nil
}

// CheckHealth performs detailed health check for Kubernetes
func (k *KubernetesChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Test cluster connectivity with kubectl cluster-info
	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info", "--request-timeout=10s")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to connect to Kubernetes cluster: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "Kubernetes cluster is accessible"
	health.Details["cluster_info"] = string(output)

	// Additional check: get node status
	cmd = exec.CommandContext(ctx, "kubectl", "get", "nodes", "--no-headers", "-o", "custom-columns=NAME:.metadata.name,STATUS:.status.conditions[?(@.type==\"Ready\")].status")
	nodeOutput, err := cmd.Output()
	if err == nil {
		health.Details["node_status"] = string(nodeOutput)
	}

	return health, nil
}

// isKubectlAvailable checks if kubectl is installed
func (k *KubernetesChecker) isKubectlAvailable() bool {
	_, err := exec.LookPath("kubectl")
	return err == nil
}

// getCurrentContext gets the current Kubernetes context
func (k *KubernetesChecker) getCurrentContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentNamespace gets the current Kubernetes namespace
func (k *KubernetesChecker) getCurrentNamespace(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	output, err := cmd.Output()
	if err != nil {
		return "default", nil // Default to "default" namespace
	}

	namespace := strings.TrimSpace(string(output))
	if namespace == "" {
		return "default", nil
	}
	return namespace, nil
}

// checkClusterAccess checks if we can access the Kubernetes cluster
func (k *KubernetesChecker) checkClusterAccess(ctx context.Context) (*CredentialStatus, error) {
	credStatus := &CredentialStatus{
		Valid: false,
		Type:  "kubeconfig",
	}

	// Test cluster access with a simple API call
	cmd := exec.CommandContext(ctx, "kubectl", "auth", "can-i", "get", "pods", "--request-timeout=10s")
	err := cmd.Run()
	if err != nil {
		credStatus.Warning = "Cannot access Kubernetes cluster"
		return credStatus, nil
	}

	credStatus.Valid = true

	// Check if credentials have expiration (for OIDC/cloud providers)
	cmd = exec.CommandContext(ctx, "kubectl", "config", "view", "--raw", "-o", "jsonpath={.users[?(@.name==\""+k.getCurrentUser(ctx)+"\")].user}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "expiry") {
		credStatus.Type = "oidc-token"
		credStatus.Warning = "Token may expire - check manually"
	}

	return credStatus, nil
}

// getCurrentUser gets the current Kubernetes user
func (k *KubernetesChecker) getCurrentUser(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "view", "--minify", "--output", "jsonpath={.contexts[0].context.user}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

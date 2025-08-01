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

// GCPChecker implements ServiceChecker for Google Cloud Platform.
type GCPChecker struct{}

// NewGCPChecker creates a new GCP status checker.
func NewGCPChecker() *GCPChecker {
	return &GCPChecker{}
}

// Name returns the service name.
func (g *GCPChecker) Name() string {
	return "gcp"
}

// CheckStatus checks GCP current status.
func (g *GCPChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "gcp",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if gcloud CLI is available
	if !g.isGcloudAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "gcloud CLI not found"
		return status, nil
	}

	// Get current project
	project, err := g.getCurrentProject(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["error"] = fmt.Sprintf("Failed to get current project: %v", err)
		return status, nil
	}

	if project == "" {
		status.Status = StatusInactive
		status.Details["error"] = "No GCP project configured"
		return status, nil
	}

	status.Current.Project = project

	// Get current account
	account, err := g.getCurrentAccount(ctx)
	if err == nil {
		status.Current.Account = account
	}

	// Get current region
	region, err := g.getCurrentRegion(ctx)
	if err == nil {
		status.Current.Region = region
	}

	// Check credentials validity
	credStatus, err := g.checkCredentials(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["credential_error"] = err.Error()
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

// CheckHealth performs detailed health check for GCP.
func (g *GCPChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Test GCP connectivity with gcloud auth list
	cmd := exec.CommandContext(ctx, "gcloud", "auth", "list", "--format=json")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to check GCP authentication: %v", err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "GCP credentials are valid and accessible"
	health.Details["auth_list"] = string(output)

	return health, nil
}

// isGcloudAvailable checks if gcloud CLI is installed.
func (g *GCPChecker) isGcloudAvailable() bool {
	_, err := exec.LookPath("gcloud")
	return err == nil
}

// getCurrentProject gets the current GCP project.
func (g *GCPChecker) getCurrentProject(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "project")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentAccount gets the current GCP account.
func (g *GCPChecker) getCurrentAccount(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentRegion gets the current GCP region.
func (g *GCPChecker) getCurrentRegion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "compute/region")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkCredentials checks GCP credentials validity.
func (g *GCPChecker) checkCredentials(ctx context.Context) (*CredentialStatus, error) {
	credStatus := &CredentialStatus{
		Valid: false,
		Type:  "gcp-credentials",
	}

	// Test credentials with gcloud auth application-default print-access-token
	cmd := exec.CommandContext(ctx, "gcloud", "auth", "print-access-token")
	err := cmd.Run()
	if err != nil {
		credStatus.Warning = "Credentials invalid or expired"
		return credStatus, nil
	}

	credStatus.Valid = true

	// Check if using service account
	cmd = exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	output, err := cmd.Output()
	if err == nil {
		account := strings.TrimSpace(string(output))
		if strings.Contains(account, ".iam.gserviceaccount.com") {
			credStatus.Type = "service-account"
		} else {
			credStatus.Type = "user-account"
		}
	}

	return credStatus, nil
}

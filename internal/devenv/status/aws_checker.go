// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// AWSChecker implements ServiceChecker for AWS
type AWSChecker struct{}

// NewAWSChecker creates a new AWS status checker
func NewAWSChecker() *AWSChecker {
	return &AWSChecker{}
}

// Name returns the service name
func (a *AWSChecker) Name() string {
	return "aws"
}

// CheckStatus checks AWS current status
func (a *AWSChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "aws",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if AWS CLI is available
	if !a.isAWSCLIAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "AWS CLI not found"
		return status, nil
	}

	// Get current profile
	profile := a.getCurrentProfile()
	if profile == "" {
		status.Status = StatusInactive
		status.Details["error"] = "No AWS profile configured"
		return status, nil
	}

	status.Current.Profile = profile

	// Get current region
	region := a.getCurrentRegion()
	status.Current.Region = region

	// Check credentials validity
	credStatus, err := a.checkCredentials(ctx)
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

// CheckHealth performs detailed health check for AWS
func (a *AWSChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Test STS GetCallerIdentity
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity", "--output", "json")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to call AWS STS: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "AWS credentials are valid and accessible"
	health.Details["caller_identity"] = string(output)

	return health, nil
}

// isAWSCLIAvailable checks if AWS CLI is installed
func (a *AWSChecker) isAWSCLIAvailable() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}

// getCurrentProfile gets the current AWS profile
func (a *AWSChecker) getCurrentProfile() string {
	// Check AWS_PROFILE environment variable
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return profile
	}

	// Check AWS config file for default profile
	cmd := exec.Command("aws", "configure", "list", "--profile", "default")
	if err := cmd.Run(); err == nil {
		return "default"
	}

	return ""
}

// getCurrentRegion gets the current AWS region
func (a *AWSChecker) getCurrentRegion() string {
	// Check AWS_REGION environment variable
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}

	// Check AWS_DEFAULT_REGION environment variable
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}

	// Try to get from AWS config
	cmd := exec.Command("aws", "configure", "get", "region")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return "us-east-1" // Default fallback
}

// checkCredentials checks AWS credentials validity
func (a *AWSChecker) checkCredentials(ctx context.Context) (*CredentialStatus, error) {
	credStatus := &CredentialStatus{
		Valid: false,
		Type:  "aws-credentials",
	}

	// Test credentials with a simple STS call
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity")
	err := cmd.Run()
	if err != nil {
		credStatus.Warning = "Credentials invalid or expired"
		return credStatus, nil
	}

	credStatus.Valid = true

	// Try to get session token expiration (for assumed roles)
	cmd = exec.CommandContext(ctx, "aws", "sts", "get-session-token", "--duration-seconds", "900")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// Parse session token response to get expiration
		// This is a simplified check - in practice you'd parse the JSON
		credStatus.Type = "session-token"
	}

	return credStatus, nil
}

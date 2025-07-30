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

// AzureChecker implements ServiceChecker for Microsoft Azure
type AzureChecker struct{}

// NewAzureChecker creates a new Azure status checker
func NewAzureChecker() *AzureChecker {
	return &AzureChecker{}
}

// Name returns the service name
func (a *AzureChecker) Name() string {
	return "azure"
}

// CheckStatus checks Azure current status
func (a *AzureChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:        "azure",
		Status:      StatusUnknown,
		Current:     CurrentConfig{},
		Credentials: CredentialStatus{},
		LastUsed:    time.Now(),
		Details:     make(map[string]string),
	}

	// Check if Azure CLI is available
	if !a.isAzureCLIAvailable() {
		status.Status = StatusInactive
		status.Details["error"] = "Azure CLI not found"
		return status, nil
	}

	// Get current subscription
	subscription, err := a.getCurrentSubscription(ctx)
	if err != nil {
		status.Status = StatusError
		status.Details["error"] = fmt.Sprintf("Failed to get current subscription: %v", err)
		return status, nil
	}

	if subscription == "" {
		status.Status = StatusInactive
		status.Details["error"] = "No Azure subscription configured"
		return status, nil
	}

	status.Current.Project = subscription

	// Get current account
	account, err := a.getCurrentAccount(ctx)
	if err == nil {
		status.Current.Account = account
	}

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

// CheckHealth performs detailed health check for Azure
func (a *AzureChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	health := &HealthStatus{
		Status:    StatusUnknown,
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Test Azure connectivity with az account show
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	health.Duration = time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Message = fmt.Sprintf("Failed to check Azure authentication: %v", err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			health.Details["stderr"] = string(exitErr.Stderr)
		}
		return health, nil
	}

	health.Status = StatusActive
	health.Message = "Azure credentials are valid and accessible"
	health.Details["account_info"] = string(output)

	return health, nil
}

// isAzureCLIAvailable checks if Azure CLI is installed
func (a *AzureChecker) isAzureCLIAvailable() bool {
	_, err := exec.LookPath("az")
	return err == nil
}

// getCurrentSubscription gets the current Azure subscription
func (a *AzureChecker) getCurrentSubscription(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--query", "name", "--output", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentAccount gets the current Azure account
func (a *AzureChecker) getCurrentAccount(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--query", "user.name", "--output", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkCredentials checks Azure credentials validity
func (a *AzureChecker) checkCredentials(ctx context.Context) (*CredentialStatus, error) {
	credStatus := &CredentialStatus{
		Valid: false,
		Type:  "azure-credentials",
	}

	// Test credentials with az account show
	cmd := exec.CommandContext(ctx, "az", "account", "show")
	err := cmd.Run()
	if err != nil {
		credStatus.Warning = "Credentials invalid or expired"
		return credStatus, nil
	}

	credStatus.Valid = true

	// Check authentication method
	cmd = exec.CommandContext(ctx, "az", "account", "show", "--query", "user.type", "--output", "tsv")
	output, err := cmd.Output()
	if err == nil {
		userType := strings.TrimSpace(string(output))
		switch userType {
		case "user":
			credStatus.Type = "user-account"
		case "servicePrincipal":
			credStatus.Type = "service-principal"
		default:
			credStatus.Type = userType
		}
	}

	return credStatus, nil
}

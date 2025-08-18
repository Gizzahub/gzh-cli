// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CommonHelpers provides shared utility functions for Git providers.
type CommonHelpers struct{}

// NewCommonHelpers creates a new instance of common helpers.
func NewCommonHelpers() *CommonHelpers {
	return &CommonHelpers{}
}

// ValidateRepositoryRequest validates common repository request fields.
func (h *CommonHelpers) ValidateRepositoryRequest(owner, name string) error {
	if strings.TrimSpace(owner) == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	return nil
}

// ValidateWebhookRequest validates common webhook request fields.
func (h *CommonHelpers) ValidateWebhookRequest(owner, repo, url string) error {
	if err := h.ValidateRepositoryRequest(owner, repo); err != nil {
		return err
	}
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must be a valid HTTP/HTTPS URL")
	}
	return nil
}

// ValidateEventRequest validates common event request fields.
func (h *CommonHelpers) ValidateEventRequest(owner, repo string) error {
	return h.ValidateRepositoryRequest(owner, repo)
}

// HandleTimeout wraps context operations with timeout handling.
func (h *CommonHelpers) HandleTimeout(ctx context.Context, timeout time.Duration, operation func(context.Context) error) error {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return operation(ctx)
}

// FormatAPIURL constructs API URLs with consistent format.
func (h *CommonHelpers) FormatAPIURL(baseURL, endpoint string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}

// ParseRepositoryURL extracts owner and repository name from various URL formats.
func (h *CommonHelpers) ParseRepositoryURL(url string) (owner, repo string, err error) {
	// Remove common prefixes
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")

	// Handle SSH format
	if strings.Contains(url, ":") && !strings.Contains(url, "/") {
		parts := strings.Split(url, ":")
		if len(parts) >= 2 {
			url = parts[1]
		}
	}

	// Extract path part
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository URL format")
	}

	// Find owner/repo part
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] != "" && parts[i+1] != "" {
			owner = parts[i]
			repo = strings.TrimSuffix(parts[i+1], ".git")
			break
		}
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not extract owner and repository from URL")
	}

	return owner, repo, nil
}

// StandardizeCapabilities returns a standardized set of capabilities.
func (h *CommonHelpers) StandardizeCapabilities(providerType string) []Capability {
	base := []Capability{
		CapabilityRepositories,
		CapabilityWebhooks,
		CapabilityEvents,
		CapabilityIssues,
		CapabilityWiki,
		CapabilityProjects,
		CapabilityReleases,
	}

	switch providerType {
	case "github":
		return append(base, []Capability{
			CapabilityPullRequests,
			CapabilityActions,
			CapabilityPackages,
		}...)
	case "gitlab":
		return append(base, []Capability{
			CapabilityMergeRequests,
			CapabilityCICD,
			CapabilityPackages,
		}...)
	case "gitea":
		return append(base, CapabilityPullRequests)
	default:
		return base
	}
}

// BuildStandardErrorResponse creates a standardized error response.
func (h *CommonHelpers) BuildStandardErrorResponse(provider, operation string, err error) *ErrorResponse {
	return &ErrorResponse{
		Provider:  provider,
		Operation: operation,
		Message:   err.Error(),
		Timestamp: time.Now().UTC(),
	}
}

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Provider  string    `json:"provider"`
	Operation string    `json:"operation"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Operation, e.Message)
}

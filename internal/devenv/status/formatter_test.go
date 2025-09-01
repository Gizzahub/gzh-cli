// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTableFormatter(t *testing.T) {
	statuses := []ServiceStatus{
		{
			Name:   "aws",
			Status: StatusActive,
			Current: CurrentConfig{
				Profile: "prod-profile",
				Region:  "us-west-2",
			},
			Credentials: CredentialStatus{
				Valid: true,
				Type:  "aws-credentials",
			},
			LastUsed: time.Now().Add(-5 * time.Minute),
		},
		{
			Name:   "gcp",
			Status: StatusInactive,
			Current: CurrentConfig{
				Project: "my-project",
			},
			Credentials: CredentialStatus{
				Valid:   false,
				Warning: "Credentials expired",
				Type:    "gcp-credentials",
			},
			LastUsed: time.Now().Add(-2 * time.Hour),
		},
	}

	tests := []struct {
		name     string
		useColor bool
	}{
		{
			name:     "with color",
			useColor: true,
		},
		{
			name:     "without color",
			useColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewStatusTableFormatter(tt.useColor)
			output, err := formatter.Format(statuses)

			require.NoError(t, err)
			assert.Contains(t, output, "Development Environment Status")
			assert.Contains(t, output, "aws")
			assert.Contains(t, output, "gcp")
			assert.Contains(t, output, "prod-profile")
			assert.Contains(t, output, "my-project")

			if tt.useColor {
				// Should contain ANSI color codes
				assert.Contains(t, output, "\033[")
			} else {
				// Should not contain ANSI color codes
				assert.NotContains(t, output, "\033[")
			}
		})
	}
}

func TestTableFormatter_EmptyStatuses(t *testing.T) {
	formatter := NewStatusTableFormatter(false)
	output, err := formatter.Format([]ServiceStatus{})

	require.NoError(t, err)
	assert.Equal(t, "No services to display", output)
}

func TestJSONFormatter(t *testing.T) {
	statuses := []ServiceStatus{
		{
			Name:   "aws",
			Status: StatusActive,
			Current: CurrentConfig{
				Profile: "test-profile",
			},
			Credentials: CredentialStatus{
				Valid: true,
				Type:  "aws-credentials",
			},
		},
	}

	tests := []struct {
		name   string
		pretty bool
	}{
		{
			name:   "pretty JSON",
			pretty: true,
		},
		{
			name:   "compact JSON",
			pretty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewStatusJSONFormatter(tt.pretty)
			output, err := formatter.Format(statuses)

			require.NoError(t, err)

			// Verify it's valid JSON
			var parsed []ServiceStatus
			err = json.Unmarshal([]byte(output), &parsed)
			require.NoError(t, err)
			assert.Len(t, parsed, 1)
			assert.Equal(t, "aws", parsed[0].Name)
			assert.Equal(t, StatusActive, parsed[0].Status)

			if tt.pretty {
				// Pretty JSON should contain indentation
				assert.Contains(t, output, "\n")
				assert.Contains(t, output, "  ")
			} else {
				// Compact JSON should be on one line (mostly)
				lines := strings.Split(strings.TrimSpace(output), "\n")
				assert.Len(t, lines, 1)
			}
		})
	}
}

func TestYAMLFormatter(t *testing.T) {
	statuses := []ServiceStatus{
		{
			Name:   "kubernetes",
			Status: StatusActive,
			Current: CurrentConfig{
				Context:   "prod-cluster",
				Namespace: "default",
			},
			Credentials: CredentialStatus{
				Valid: true,
				Type:  "kubeconfig",
			},
		},
	}

	formatter := NewStatusYAMLFormatter()
	output, err := formatter.Format(statuses)

	require.NoError(t, err)

	// Verify it's valid YAML
	var parsed []ServiceStatus
	err = yaml.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed, 1)
	assert.Equal(t, "kubernetes", parsed[0].Name)
	assert.Equal(t, StatusActive, parsed[0].Status)
	assert.Equal(t, "prod-cluster", parsed[0].Current.Context)

	// YAML output should contain expected structure
	assert.Contains(t, output, "name: kubernetes")
	assert.Contains(t, output, "status: active")
	assert.Contains(t, output, "context: prod-cluster")
}

func TestFormatDuration(t *testing.T) {
	formatter := NewStatusTableFormatter(false)

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "less than minute",
			duration: 30 * time.Second,
			expected: "< 1 min",
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			expected: "5 min",
		},
		{
			name:     "hours",
			duration: 2 * time.Hour,
			expected: "2 hour",
		},
		{
			name:     "days",
			duration: 3 * 24 * time.Hour,
			expected: "3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

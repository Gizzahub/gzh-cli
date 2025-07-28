// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyResolver(t *testing.T) {
	services := map[string]ServiceConfig{
		"aws": {
			AWS: &AWSConfig{Profile: "test"},
		},
		"kubernetes": {
			Kubernetes: &KubernetesConfig{Context: "test"},
		},
		"docker": {
			Docker: &DockerConfig{Context: "test"},
		},
	}

	tests := []struct {
		name         string
		dependencies []string
		wantOrder    []string
		wantErr      bool
	}{
		{
			name:         "no dependencies",
			dependencies: []string{},
			wantOrder:    []string{"aws", "docker", "kubernetes"}, // alphabetical order
			wantErr:      false,
		},
		{
			name:         "simple dependency",
			dependencies: []string{"aws -> kubernetes"},
			wantOrder:    []string{"aws", "docker", "kubernetes"}, // aws before kubernetes
			wantErr:      false,
		},
		{
			name:         "chain dependency",
			dependencies: []string{"aws -> kubernetes", "docker -> kubernetes"},
			wantOrder:    []string{"aws", "docker", "kubernetes"}, // aws and docker before kubernetes
			wantErr:      false,
		},
		{
			name:         "circular dependency",
			dependencies: []string{"aws -> kubernetes", "kubernetes -> aws"},
			wantErr:      true,
		},
		{
			name:         "invalid service",
			dependencies: []string{"aws -> nonexistent"},
			wantErr:      true,
		},
		{
			name:         "invalid format",
			dependencies: []string{"aws kubernetes"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDependencyResolver(services, tt.dependencies)
			order, err := resolver.GetExecutionOrder()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, order, len(services))

			// Check that all services are included
			serviceSet := make(map[string]bool)
			for _, service := range order {
				serviceSet[service] = true
			}
			for serviceName := range services {
				assert.True(t, serviceSet[serviceName], "service %s should be in execution order", serviceName)
			}

			// Check dependency ordering
			if len(tt.wantOrder) > 0 {
				serviceIndex := make(map[string]int)
				for i, service := range order {
					serviceIndex[service] = i
				}

				for _, dep := range tt.dependencies {
					parts := parseDependency(dep)
					if len(parts) == 2 {
						from, to := parts[0], parts[1]
						assert.True(t, serviceIndex[from] < serviceIndex[to],
							"service %s should come before %s", from, to)
					}
				}
			}
		})
	}
}

func TestParseDependency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple dependency",
			input:    "aws -> kubernetes",
			expected: []string{"aws", "kubernetes"},
		},
		{
			name:     "dependency with spaces",
			input:    "  aws  ->  kubernetes  ",
			expected: []string{"aws", "kubernetes"},
		},
		{
			name:     "invalid format",
			input:    "aws kubernetes",
			expected: []string{"aws kubernetes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDependency(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetParallelGroups(t *testing.T) {
	services := map[string]ServiceConfig{
		"aws": {
			AWS: &AWSConfig{Profile: "test"},
		},
		"gcp": {
			GCP: &GCPConfig{Project: "test"},
		},
		"kubernetes": {
			Kubernetes: &KubernetesConfig{Context: "test"},
		},
	}

	dependencies := []string{"aws -> kubernetes", "gcp -> kubernetes"}
	resolver := NewDependencyResolver(services, dependencies)

	groups, err := resolver.GetParallelGroups()
	require.NoError(t, err)

	// Should have 2 groups: [aws, gcp] and [kubernetes]
	assert.Len(t, groups, 2)

	// First group should have aws and gcp (can be parallel)
	assert.Len(t, groups[0].Services, 2)
	assert.Contains(t, groups[0].Services, "aws")
	assert.Contains(t, groups[0].Services, "gcp")
	assert.Equal(t, 0, groups[0].Level)

	// Second group should have kubernetes
	assert.Len(t, groups[1].Services, 1)
	assert.Contains(t, groups[1].Services, "kubernetes")
	assert.Equal(t, 1, groups[1].Level)
}

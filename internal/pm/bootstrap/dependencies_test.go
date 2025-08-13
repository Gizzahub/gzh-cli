// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyResolver_ResolveDependencies(t *testing.T) {
	resolver := NewDependencyResolver()

	// Setup dependencies: brew -> asdf -> nvm, rbenv, pyenv
	resolver.AddDependency("asdf", []string{"brew"})
	resolver.AddDependency("rbenv", []string{"brew"})
	resolver.AddDependency("pyenv", []string{"brew"})
	resolver.AddDependency("nvm", []string{})    // No dependencies
	resolver.AddDependency("sdkman", []string{}) // No dependencies

	tests := []struct {
		name        string
		managers    []string
		expected    []string
		expectError bool
	}{
		{
			name:     "single manager no deps",
			managers: []string{"nvm"},
			expected: []string{"nvm"},
		},
		{
			name:     "single manager with deps",
			managers: []string{"asdf"},
			expected: []string{"asdf"}, // Only originally requested managers in result
		},
		{
			name:     "multiple managers with shared deps",
			managers: []string{"asdf", "rbenv"},
			expected: []string{"asdf", "rbenv"},
		},
		{
			name:     "all managers",
			managers: []string{"asdf", "rbenv", "pyenv", "nvm", "sdkman"},
			expected: []string{"asdf", "rbenv", "pyenv", "nvm", "sdkman"},
		},
		{
			name:     "empty input",
			managers: []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveDependencies(tt.managers)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestDependencyResolver_CircularDependency(t *testing.T) {
	resolver := NewDependencyResolver()

	// Create circular dependency: A -> B -> C -> A
	resolver.AddDependency("a", []string{"b"})
	resolver.AddDependency("b", []string{"c"})
	resolver.AddDependency("c", []string{"a"})

	_, err := resolver.ResolveDependencies([]string{"a", "b", "c"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDependencyResolver_GetDependencies(t *testing.T) {
	resolver := NewDependencyResolver()
	resolver.AddDependency("asdf", []string{"brew", "git"})

	deps := resolver.GetDependencies("asdf")
	expected := []string{"brew", "git"}
	assert.ElementsMatch(t, expected, deps)

	// Non-existent manager should return empty slice
	deps = resolver.GetDependencies("nonexistent")
	assert.Empty(t, deps)
}

func TestDependencyResolver_HasDependency(t *testing.T) {
	resolver := NewDependencyResolver()
	resolver.AddDependency("asdf", []string{"brew"})

	assert.True(t, resolver.HasDependency("asdf", "brew"))
	assert.False(t, resolver.HasDependency("asdf", "nvm"))
	assert.False(t, resolver.HasDependency("nonexistent", "brew"))
}

func TestDependencyResolver_GetAllDependencies(t *testing.T) {
	resolver := NewDependencyResolver()

	// Setup chain: a -> b -> c -> d
	resolver.AddDependency("a", []string{"b"})
	resolver.AddDependency("b", []string{"c"})
	resolver.AddDependency("c", []string{"d"})

	deps, err := resolver.GetAllDependencies("a")
	require.NoError(t, err)
	expected := []string{"b", "c", "d"}
	assert.ElementsMatch(t, expected, deps)

	// Manager with no dependencies
	deps, err = resolver.GetAllDependencies("d")
	require.NoError(t, err)
	assert.Empty(t, deps)
}

func TestDependencyResolver_ValidateNoCycles(t *testing.T) {
	resolver := NewDependencyResolver()

	// Valid graph
	resolver.AddDependency("a", []string{"b"})
	resolver.AddDependency("b", []string{"c"})

	err := resolver.ValidateNoCycles()
	assert.NoError(t, err)

	// Add cycle
	resolver.AddDependency("c", []string{"a"})

	err = resolver.ValidateNoCycles()
	assert.Error(t, err)
}

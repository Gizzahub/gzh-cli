// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package registry

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLifecycleManager(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("GZ_EXPERIMENTAL")
	defer os.Setenv("GZ_EXPERIMENTAL", originalEnv)

	t.Run("default without env var", func(t *testing.T) {
		os.Unsetenv("GZ_EXPERIMENTAL")
		lm := NewLifecycleManager()
		assert.NotNil(t, lm)
		assert.False(t, lm.IsExperimentalEnabled())
	})

	t.Run("enabled with env var", func(t *testing.T) {
		os.Setenv("GZ_EXPERIMENTAL", "1")
		lm := NewLifecycleManager()
		assert.NotNil(t, lm)
		assert.True(t, lm.IsExperimentalEnabled())
	})

	t.Run("not enabled with wrong env value", func(t *testing.T) {
		os.Setenv("GZ_EXPERIMENTAL", "true")
		lm := NewLifecycleManager()
		assert.False(t, lm.IsExperimentalEnabled())
	})
}

func TestLifecycleManager_CheckCommand(t *testing.T) {
	tests := []struct {
		name              string
		meta              CommandMetadata
		allowExperimental bool
		expectError       bool
		errorContains     string
	}{
		{
			name: "stable command - no error",
			meta: CommandMetadata{
				Name:      "test",
				Lifecycle: LifecycleStable,
			},
			allowExperimental: false,
			expectError:       false,
		},
		{
			name: "experimental command - disabled",
			meta: CommandMetadata{
				Name:      "test-exp",
				Lifecycle: LifecycleExperimental,
			},
			allowExperimental: false,
			expectError:       true,
			errorContains:     "experimental",
		},
		{
			name: "experimental command - enabled",
			meta: CommandMetadata{
				Name:      "test-exp",
				Lifecycle: LifecycleExperimental,
			},
			allowExperimental: true,
			expectError:       false,
		},
		{
			name: "deprecated command - always allowed with warning",
			meta: CommandMetadata{
				Name:      "test-dep",
				Lifecycle: LifecycleDeprecated,
			},
			allowExperimental: false,
			expectError:       false,
		},
		{
			name: "beta command - always allowed with info",
			meta: CommandMetadata{
				Name:      "test-beta",
				Lifecycle: LifecycleBeta,
			},
			allowExperimental: false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLifecycleManager()
			if tt.allowExperimental {
				lm.EnableExperimental()
			}

			err := lm.CheckCommand(tt.meta)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLifecycleManager_EnableDisableExperimental(t *testing.T) {
	lm := NewLifecycleManager()

	// Initially disabled
	assert.False(t, lm.IsExperimentalEnabled())

	// Enable
	lm.EnableExperimental()
	assert.True(t, lm.IsExperimentalEnabled())

	// Disable
	lm.DisableExperimental()
	assert.False(t, lm.IsExperimentalEnabled())
}

func TestLifecycleManager_FilterCommands(t *testing.T) {
	// Create test providers
	stableProvider := &mockProviderWithMeta{
		meta: CommandMetadata{
			Name:      "stable",
			Lifecycle: LifecycleStable,
		},
	}

	expProvider := &mockProviderWithMeta{
		meta: CommandMetadata{
			Name:      "experimental",
			Lifecycle: LifecycleExperimental,
		},
	}

	betaProvider := &mockProviderWithMeta{
		meta: CommandMetadata{
			Name:      "beta",
			Lifecycle: LifecycleBeta,
		},
	}

	providers := []CommandProvider{stableProvider, expProvider, betaProvider}

	t.Run("filter experimental when disabled", func(t *testing.T) {
		lm := NewLifecycleManager()
		filtered := lm.FilterCommands(providers)

		assert.Len(t, filtered, 2) // stable + beta
		names := make([]string, len(filtered))
		for i, p := range filtered {
			names[i] = GetMetadata(p).Name
		}
		assert.Contains(t, names, "stable")
		assert.Contains(t, names, "beta")
		assert.NotContains(t, names, "experimental")
	})

	t.Run("include experimental when enabled", func(t *testing.T) {
		lm := NewLifecycleManager()
		lm.EnableExperimental()
		filtered := lm.FilterCommands(providers)

		assert.Len(t, filtered, 3) // all
		names := make([]string, len(filtered))
		for i, p := range filtered {
			names[i] = GetMetadata(p).Name
		}
		assert.Contains(t, names, "stable")
		assert.Contains(t, names, "beta")
		assert.Contains(t, names, "experimental")
	})
}

func TestLifecycleManager_CheckDependencies(t *testing.T) {
	lm := NewLifecycleManager()

	t.Run("no dependencies", func(t *testing.T) {
		meta := CommandMetadata{
			Name:         "test",
			Dependencies: []string{},
		}
		missing := lm.CheckDependencies(meta)
		assert.Empty(t, missing)
	})

	t.Run("with dependencies", func(t *testing.T) {
		meta := CommandMetadata{
			Name:         "test",
			Dependencies: []string{"git", "nonexistent-tool-xyz"},
		}
		missing := lm.CheckDependencies(meta)
		// Since isCommandAvailable always returns false in test env,
		// all dependencies will be "missing"
		assert.NotEmpty(t, missing)
	})
}

// mockProviderWithMeta for testing.
type mockProviderWithMeta struct {
	meta CommandMetadata
}

func (m *mockProviderWithMeta) Command() *cobra.Command {
	return &cobra.Command{
		Use:   m.meta.Name,
		Short: "Mock command",
	}
}

func (m *mockProviderWithMeta) Metadata() CommandMetadata {
	return m.meta
}

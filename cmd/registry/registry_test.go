// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package registry

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockProvider는 메타데이터가 없는 기본 provider
type mockProvider struct {
	name string
}

func (m mockProvider) Command() *cobra.Command {
	return &cobra.Command{
		Use:   m.name,
		Short: "Mock command",
	}
}

// mockProviderWithMetadata는 메타데이터를 지원하는 provider
type mockProviderWithMetadata struct {
	mockProvider
	metadata CommandMetadata
}

func (m mockProviderWithMetadata) Metadata() CommandMetadata {
	return m.metadata
}

func TestRegister(t *testing.T) {
	// 테스트를 위해 providers 초기화
	providers = []CommandProvider{}

	p1 := mockProvider{name: "test1"}
	Register(p1)

	assert.Len(t, providers, 1, "Provider should be registered")
}

func TestList(t *testing.T) {
	providers = []CommandProvider{}

	p1 := mockProvider{name: "test1"}
	p2 := mockProvider{name: "test2"}

	Register(p1)
	Register(p2)

	list := List()
	assert.Len(t, list, 2, "Should return all registered providers")
}

func TestByCategory(t *testing.T) {
	providers = []CommandProvider{}

	p1 := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "git"},
		metadata: CommandMetadata{
			Name:     "git",
			Category: CategoryGit,
		},
	}

	p2 := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "quality"},
		metadata: CommandMetadata{
			Name:     "quality",
			Category: CategoryQuality,
		},
	}

	p3 := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "synclone"},
		metadata: CommandMetadata{
			Name:     "synclone",
			Category: CategoryGit,
		},
	}

	Register(p1)
	Register(p2)
	Register(p3)

	gitCommands := ByCategory(CategoryGit)
	assert.Len(t, gitCommands, 2, "Should return 2 git commands")

	qualityCommands := ByCategory(CategoryQuality)
	assert.Len(t, qualityCommands, 1, "Should return 1 quality command")
}

func TestStableCommands(t *testing.T) {
	providers = []CommandProvider{}

	stable := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "stable"},
		metadata: CommandMetadata{
			Name:      "stable",
			Lifecycle: LifecycleStable,
		},
	}

	experimental := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "experimental"},
		metadata: CommandMetadata{
			Name:      "experimental",
			Lifecycle: LifecycleExperimental,
		},
	}

	noMetadata := mockProvider{name: "noMeta"}

	Register(stable)
	Register(experimental)
	Register(noMetadata)

	stableCommands := StableCommands()
	// stable + noMetadata (메타데이터 없으면 stable로 간주)
	assert.Len(t, stableCommands, 2, "Should return stable commands and those without metadata")
}

func TestExperimentalCommands(t *testing.T) {
	providers = []CommandProvider{}

	experimental1 := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "exp1"},
		metadata: CommandMetadata{
			Name:      "exp1",
			Lifecycle: LifecycleExperimental,
		},
	}

	experimental2 := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "exp2"},
		metadata: CommandMetadata{
			Name:         "exp2",
			Experimental: true,
			Lifecycle:    LifecycleStable, // lifecycle은 stable이지만 Experimental flag true
		},
	}

	stable := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "stable"},
		metadata: CommandMetadata{
			Name:      "stable",
			Lifecycle: LifecycleStable,
		},
	}

	Register(experimental1)
	Register(experimental2)
	Register(stable)

	expCommands := ExperimentalCommands()
	assert.Len(t, expCommands, 2, "Should return experimental commands")
}

func TestAllByPriority(t *testing.T) {
	providers = []CommandProvider{}

	high := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "high"},
		metadata: CommandMetadata{
			Name:     "high",
			Priority: 10,
		},
	}

	medium := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "medium"},
		metadata: CommandMetadata{
			Name:     "medium",
			Priority: 50,
		},
	}

	low := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "low"},
		metadata: CommandMetadata{
			Name:     "low",
			Priority: 100,
		},
	}

	// 등록 순서는 medium, high, low
	Register(medium)
	Register(high)
	Register(low)

	sorted := AllByPriority()
	assert.Len(t, sorted, 3)

	// 우선순위 순으로 정렬되어야 함 (낮은 숫자가 높은 우선순위)
	assert.Equal(t, "high", sorted[0].Command().Name())
	assert.Equal(t, "medium", sorted[1].Command().Name())
	assert.Equal(t, "low", sorted[2].Command().Name())
}

func TestHasMetadata(t *testing.T) {
	withMeta := mockProviderWithMetadata{
		mockProvider: mockProvider{name: "test"},
		metadata:     CommandMetadata{Name: "test"},
	}

	withoutMeta := mockProvider{name: "test"}

	assert.True(t, HasMetadata(withMeta), "Should detect metadata support")
	assert.False(t, HasMetadata(withoutMeta), "Should detect no metadata support")
}

func TestGetMetadata(t *testing.T) {
	t.Run("Provider with metadata", func(t *testing.T) {
		expected := CommandMetadata{
			Name:      "test",
			Category:  CategoryGit,
			Version:   "1.0.0",
			Priority:  10,
			Lifecycle: LifecycleStable,
		}

		p := mockProviderWithMetadata{
			mockProvider: mockProvider{name: "test"},
			metadata:     expected,
		}

		meta := GetMetadata(p)
		assert.Equal(t, expected, meta)
	})

	t.Run("Provider without metadata", func(t *testing.T) {
		p := mockProvider{name: "test"}

		meta := GetMetadata(p)
		assert.Equal(t, "test", meta.Name)
		assert.Equal(t, CategoryUtility, meta.Category)
		assert.Equal(t, "unknown", meta.Version)
		assert.Equal(t, 999, meta.Priority)
		assert.Equal(t, LifecycleStable, meta.Lifecycle)
	})
}

func TestCommandCategory(t *testing.T) {
	// 카테고리 상수가 올바르게 정의되었는지 확인
	categories := []CommandCategory{
		CategoryGit,
		CategoryDevelopment,
		CategoryQuality,
		CategoryNetwork,
		CategoryUtility,
		CategoryConfig,
	}

	for _, cat := range categories {
		assert.NotEmpty(t, string(cat), "Category should not be empty")
	}
}

func TestLifecycleStage(t *testing.T) {
	// Lifecycle 상수가 올바르게 정의되었는지 확인
	stages := []LifecycleStage{
		LifecycleStable,
		LifecycleBeta,
		LifecycleExperimental,
		LifecycleDeprecated,
	}

	for _, stage := range stages {
		assert.NotEmpty(t, string(stage), "Lifecycle stage should not be empty")
	}
}

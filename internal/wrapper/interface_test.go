// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package wrapper

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

// mockLibrary is a mock external library for testing.
type mockLibrary struct {
	*BaseWrapper
}

func (m *mockLibrary) CreateCommand(appCtx *app.AppContext) (*cobra.Command, error) {
	return &cobra.Command{
		Use:   m.Name(),
		Short: "Mock command",
	}, nil
}

func TestNewBaseWrapper(t *testing.T) {
	wrapper := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{"git"})

	assert.Equal(t, "test", wrapper.Name())
	assert.Equal(t, "1.0.0", wrapper.Version())
	assert.Equal(t, "https://github.com/test/test", wrapper.Repository())
	assert.Equal(t, []string{"git"}, wrapper.Dependencies())
}

func TestBaseWrapper_Validate_Success(t *testing.T) {
	// ls, echo 등은 대부분의 시스템에 존재
	wrapper := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{"ls"})

	err := wrapper.Validate()
	assert.NoError(t, err)
}

func TestBaseWrapper_Validate_MissingDependency(t *testing.T) {
	wrapper := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{"nonexistent-command-12345"})

	err := wrapper.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing dependency")
}

func TestBaseWrapper_Validate_NoDependencies(t *testing.T) {
	wrapper := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{})

	err := wrapper.Validate()
	assert.NoError(t, err, "No dependencies should pass validation")
}

func TestGetMetadata(t *testing.T) {
	base := NewBaseWrapper("quality", "1.2.3", "https://github.com/gizzahub/gzh-cli-quality", []string{"go", "npm"})
	lib := &mockLibrary{BaseWrapper: base}

	meta := GetMetadata(lib)

	assert.Equal(t, "quality", meta.Name)
	assert.Equal(t, "1.2.3", meta.Version)
	assert.Equal(t, "https://github.com/gizzahub/gzh-cli-quality", meta.Repository)
	assert.Equal(t, []string{"go", "npm"}, meta.Dependencies)
	assert.Equal(t, "active", meta.Status)
}

func TestMockLibrary_CreateCommand(t *testing.T) {
	base := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{})
	lib := &mockLibrary{BaseWrapper: base}

	cmd, err := lib.CreateCommand(&app.AppContext{})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "test", cmd.Use)
}

func TestExternalLibraryInterface(t *testing.T) {
	// ExternalLibrary 인터페이스가 올바르게 정의되었는지 확인
	base := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{})
	var lib ExternalLibrary = &mockLibrary{BaseWrapper: base}
	assert.NotNil(t, lib)
}

func TestBaseWrapper_AllMethods(t *testing.T) {
	wrapper := NewBaseWrapper(
		"mylib",
		"2.0.0",
		"https://github.com/myorg/mylib",
		[]string{"dep1", "dep2"},
	)

	// 모든 메서드 테스트
	assert.Equal(t, "mylib", wrapper.Name())
	assert.Equal(t, "2.0.0", wrapper.Version())
	assert.Equal(t, "https://github.com/myorg/mylib", wrapper.Repository())
	assert.Equal(t, []string{"dep1", "dep2"}, wrapper.Dependencies())
}

// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectMetricsUseRootScopedReads(t *testing.T) {
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main_test.go"), []byte("package main\n"), 0o600))

	projectFiles, err := openProjectRootReader(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, projectFiles.Close())
	})

	opts := metricsOptions{projectPath: projectPath, includeTests: false}
	report := &CodeQualityReport{FileAnalysis: make([]FileQualityInfo, 0)}

	require.NoError(t, collectProjectMetrics(report, opts, projectFiles))
	require.NoError(t, collectComplexityMetrics(report, opts, projectFiles))
	require.NoError(t, collectFileAnalysis(report, opts, projectFiles))

	assert.Equal(t, 1, report.Summary.TotalFiles)
	assert.Len(t, report.FileAnalysis, 1)
	assert.Equal(t, filepath.Join(projectPath, "main.go"), report.FileAnalysis[0].Path)
}

func TestProjectMetricsRejectSymlinkEscape(t *testing.T) {
	projectPath := t.TempDir()
	outsidePath := t.TempDir()
	outsideFile := filepath.Join(outsidePath, "outside.go")
	require.NoError(t, os.WriteFile(outsideFile, []byte("package outside\n"), 0o600))

	linkPath := filepath.Join(projectPath, "escape.go")
	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}

	projectFiles, err := openProjectRootReader(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, projectFiles.Close())
	})

	report := &CodeQualityReport{}
	err = collectProjectMetrics(report, metricsOptions{projectPath: projectPath, includeTests: true}, projectFiles)
	require.Error(t, err)
	assert.Zero(t, report.Summary.TotalFiles)
}

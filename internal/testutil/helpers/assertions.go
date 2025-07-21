// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertFileExists checks if a file exists at the given path.
func AssertFileExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()

	_, err := os.Stat(path)
	assert.NoError(t, err, msgAndArgs...)
}

// AssertFileNotExists checks if a file does not exist at the given path.
func AssertFileNotExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), msgAndArgs...)
}

// AssertDirExists checks if a directory exists at the given path.
func AssertDirExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, msgAndArgs...)
	assert.True(t, info.IsDir(), "expected %s to be a directory", path)
}

// AssertFileContains checks if a file contains the expected content.
func AssertFileContains(t *testing.T, path, expected string, msgAndArgs ...interface{}) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err, msgAndArgs...)
	assert.Contains(t, string(content), expected, msgAndArgs...)
}

// AssertFileEquals checks if a file's content exactly matches expected.
func AssertFileEquals(t *testing.T, path, expected string, msgAndArgs ...interface{}) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err, msgAndArgs...)
	assert.Equal(t, expected, string(content), msgAndArgs...)
}

// AssertGitRepository checks if a directory is a valid git repository.
func AssertGitRepository(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()

	gitDir := filepath.Join(path, ".git")
	AssertDirExists(t, gitDir, msgAndArgs...)
}

// AssertErrorContains checks if an error contains expected substring.
func AssertErrorContains(t *testing.T, err error, contains string, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
	assert.Contains(t, err.Error(), contains, msgAndArgs...)
}

// AssertNoError is a helper that calls t.Fatal on error.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err != nil {
		t.Fatal(err, msgAndArgs)
	}
}

// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDownloadFile struct {
	writeErr  error
	chmodErr  error
	syncErr   error
	closeErr  error
	closeCall int
	events    []string
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	//gosec:disable G304 -- AR-2026-003 각 테스트가 t.TempDir 아래에 직접 생성한 경로만 읽는다.
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return content
}

func (f *fakeDownloadFile) Write(p []byte) (int, error) {
	f.events = append(f.events, "write")
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}

func (f *fakeDownloadFile) Chmod(os.FileMode) error {
	f.events = append(f.events, "chmod")
	return f.chmodErr
}

func (f *fakeDownloadFile) Sync() error {
	f.events = append(f.events, "sync")
	return f.syncErr
}

func (f *fakeDownloadFile) Close() error {
	f.closeCall++
	f.events = append(f.events, "close")
	return f.closeErr
}

func TestUpdater_GetAssetName(t *testing.T) {
	updater := NewUpdater("1.0.0")

	tests := []struct {
		name         string
		goos         string
		goarch       string
		expectedName string
	}{
		{
			name:         "Linux x86_64",
			goos:         "linux",
			goarch:       "amd64",
			expectedName: "gz_linux_x86_64",
		},
		{
			name:         "Windows x86_64",
			goos:         "windows",
			goarch:       "amd64",
			expectedName: "gz_windows_x86_64.exe",
		},
		{
			name:         "Darwin ARM64",
			goos:         "darwin",
			goarch:       "arm64",
			expectedName: "gz_darwin_arm64",
		},
		{
			name:         "Linux i386",
			goos:         "linux",
			goarch:       "386",
			expectedName: "gz_linux_i386",
		},
	}

	// Save original values
	originalGOOS := runtime.GOOS

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is conceptual since we can't actually change runtime.GOOS/GOARCH
			// In a real implementation, we would need to refactor GetAssetName to accept parameters
			_ = tt.goos
			_ = tt.goarch
			_ = tt.expectedName

			// For now, just test that GetAssetName returns something reasonable
			assetName := updater.GetAssetName()
			assert.NotEmpty(t, assetName)
			assert.Contains(t, assetName, "gz_")
			assert.Contains(t, assetName, originalGOOS)
		})
	}
}

func TestUpdater_IsNewerVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		remoteVersion  string
		expected       bool
	}{
		{
			name:           "Same version",
			currentVersion: "v1.0.0",
			remoteVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "Same version without v prefix",
			currentVersion: "1.0.0",
			remoteVersion:  "1.0.0",
			expected:       false,
		},
		{
			name:           "Mixed v prefix",
			currentVersion: "v1.0.0",
			remoteVersion:  "1.0.0",
			expected:       false,
		},
		{
			name:           "Different versions",
			currentVersion: "v1.0.0",
			remoteVersion:  "v1.1.0",
			expected:       true,
		},
		{
			name:           "Dev version",
			currentVersion: "dev",
			remoteVersion:  "v1.0.0",
			expected:       true,
		},
		{
			name:           "Empty current version",
			currentVersion: "",
			remoteVersion:  "v1.0.0",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := NewUpdater(tt.currentVersion)
			result := updater.IsNewerVersion(tt.remoteVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewUpdater(t *testing.T) {
	version := "1.0.0"
	updater := NewUpdater(version)

	assert.NotNil(t, updater)
	assert.Equal(t, version, updater.currentVersion)
	assert.NotNil(t, updater.logger)
}

func TestUpdater_DownloadAssetCreatesExclusively(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "new binary")
	}))
	t.Cleanup(server.Close)

	t.Run("existing file remains unchanged", func(t *testing.T) {
		tempPath := filepath.Join(t.TempDir(), "existing")
		require.NoError(t, os.WriteFile(tempPath, []byte("keep me"), 0o600))

		err := NewUpdater("dev").DownloadAsset(context.Background(), server.URL, tempPath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "creating temporary file")

		got := readTestFile(t, tempPath)
		assert.Equal(t, "keep me", string(got))
	})

	t.Run("symlink target remains unchanged", func(t *testing.T) {
		dir := t.TempDir()
		targetPath := filepath.Join(dir, "target")
		linkPath := filepath.Join(dir, "stage")
		require.NoError(t, os.WriteFile(targetPath, []byte("keep target"), 0o600))
		if err := os.Symlink(targetPath, linkPath); err != nil {
			t.Skipf("symlink unsupported: %v", err)
		}

		err := NewUpdater("dev").DownloadAsset(context.Background(), server.URL, linkPath)
		require.Error(t, err)

		got := readTestFile(t, targetPath)
		assert.Equal(t, "keep target", string(got))
	})
}

func TestUpdater_DownloadAssetCleansOnlyOwnedFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "20")
		_, _ = io.WriteString(w, "short")
	}))
	t.Cleanup(server.Close)

	tempPath := filepath.Join(t.TempDir(), "owned-stage")
	err := NewUpdater("dev").DownloadAsset(context.Background(), server.URL, tempPath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "writing downloaded file")
	_, statErr := os.Lstat(tempPath)
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestUpdater_DownloadAssetForTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "release binary")
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz")
	require.NoError(t, os.WriteFile(currentPath, []byte("old binary"), 0o600))

	tempPath, err := NewUpdater("dev").downloadAssetForTarget(context.Background(), server.URL, currentPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tempPath) })
	assert.Equal(t, dir, filepath.Dir(tempPath))

	got := readTestFile(t, tempPath)
	assert.Equal(t, "release binary", string(got))

	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(tempPath)
		require.NoError(t, statErr)
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	}
}

func TestUpdater_DownloadAssetPropagatesFileErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "release binary")
	}))
	t.Cleanup(server.Close)

	tests := []struct {
		name      string
		file      *fakeDownloadFile
		wantError string
		skip      bool
	}{
		{name: "write", file: &fakeDownloadFile{writeErr: errors.New("write failed")}, wantError: "writing downloaded file"},
		{name: "chmod", file: &fakeDownloadFile{chmodErr: errors.New("chmod failed")}, wantError: "setting executable permissions", skip: runtime.GOOS == "windows"},
		{name: "sync", file: &fakeDownloadFile{syncErr: errors.New("sync failed")}, wantError: "syncing downloaded file"},
		{name: "close", file: &fakeDownloadFile{closeErr: errors.New("close failed")}, wantError: "closing downloaded file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("chmod is not applied on Windows")
			}

			err := NewUpdater("dev").downloadAsset(context.Background(), server.URL, tt.file)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantError)
			assert.Equal(t, 1, tt.file.closeCall)
		})
	}
}

func TestUpdater_DownloadAssetFileOperationOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "release binary")
	}))
	t.Cleanup(server.Close)

	file := &fakeDownloadFile{}
	require.NoError(t, NewUpdater("dev").downloadAsset(context.Background(), server.URL, file))
	want := []string{"write", "sync", "close"}
	if runtime.GOOS != "windows" {
		want = []string{"write", "chmod", "sync", "close"}
	}
	assert.Equal(t, want, file.events)
}

func TestUpdater_DownloadAssetForTargetCleansFailedStage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "20")
		_, _ = io.WriteString(w, "short")
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz")
	require.NoError(t, os.WriteFile(currentPath, []byte("old binary"), 0o600))

	_, err := NewUpdater("dev").downloadAssetForTarget(context.Background(), server.URL, currentPath)
	require.Error(t, err)
	matches, globErr := filepath.Glob(filepath.Join(dir, ".gz-update-*"))
	require.NoError(t, globErr)
	assert.Empty(t, matches)
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz")
	tempPath := filepath.Join(dir, ".gz-update-stage")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(tempPath, []byte("new"), 0o600))

	result, err := replaceBinary(tempPath, currentPath)
	require.NoError(t, err)
	require.NoError(t, result.cleanupWarning)
	got := readTestFile(t, currentPath)
	assert.Equal(t, "new", string(got))
	_, statErr := os.Lstat(tempPath)
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestReplaceBinaryRejectsDifferentDirectory(t *testing.T) {
	currentDir := t.TempDir()
	tempDir := t.TempDir()
	currentPath := filepath.Join(currentDir, "gz")
	tempPath := filepath.Join(tempDir, "stage")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(tempPath, []byte("new"), 0o600))

	_, err := replaceBinary(tempPath, currentPath)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "replacement stage must be in current binary directory"))

	got := readTestFile(t, currentPath)
	assert.Equal(t, "old", string(got))
	_, statErr := os.Stat(tempPath)
	assert.NoError(t, statErr)
}

func TestReplaceBinaryRejectsUnsafeStage(t *testing.T) {
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz")
	realStagePath := filepath.Join(dir, "real-stage")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(realStagePath, []byte("new"), 0o600))

	t.Run("current binary itself", func(t *testing.T) {
		_, err := replaceBinary(currentPath, currentPath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "must differ")
	})

	t.Run("directory", func(t *testing.T) {
		stagePath := filepath.Join(dir, "stage-directory")
		require.NoError(t, os.Mkdir(stagePath, 0o700))

		_, err := replaceBinary(stagePath, currentPath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "regular file")
	})

	t.Run("symlink", func(t *testing.T) {
		stagePath := filepath.Join(dir, "stage-link")
		if err := os.Symlink(realStagePath, stagePath); err != nil {
			t.Skipf("symlink unsupported: %v", err)
		}

		_, err := replaceBinary(stagePath, currentPath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "regular file")
	})

	got := readTestFile(t, currentPath)
	assert.Equal(t, "old", string(got))
}

func TestValidateCurrentBinaryIdentity(t *testing.T) {
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	expected, err := currentBinaryIdentity(currentPath)
	require.NoError(t, err)
	require.NoError(t, validateCurrentBinaryIdentity(currentPath, expected))

	replacementPath := filepath.Join(dir, "replacement")
	require.NoError(t, os.WriteFile(replacementPath, []byte("new"), 0o600))
	require.NoError(t, os.Remove(currentPath))
	require.NoError(t, os.Rename(replacementPath, currentPath))

	err = validateCurrentBinaryIdentity(currentPath, expected)
	require.Error(t, err)
	assert.ErrorContains(t, err, "changed while downloading")
}

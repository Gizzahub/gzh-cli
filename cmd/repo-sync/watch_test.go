//nolint:testpackage // White-box testing needed for internal function access
package reposync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewRepositoryWatcher(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.Equal(t, config, watcher.config)
	assert.NotNil(t, watcher.events)
	assert.NotNil(t, watcher.batches)

	err = watcher.Close()
	assert.NoError(t, err)
}

func TestFileChangeEventDeduplication(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	// Create test events with same file path but different timestamps
	events := []FileChangeEvent{
		{
			Path:      "/test/file.go",
			Operation: "write",
			Timestamp: time.Now().Add(-2 * time.Second),
		},
		{
			Path:      "/test/file.go",
			Operation: "write",
			Timestamp: time.Now().Add(-1 * time.Second),
		},
		{
			Path:      "/test/file.go",
			Operation: "write",
			Timestamp: time.Now(),
		},
	}

	deduped := watcher.deduplicateEvents(events)
	assert.Len(t, deduped, 1)
	assert.Equal(t, events[2].Timestamp, deduped[0].Timestamp) // Should keep the latest
}

func TestShouldIgnorePatterns(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	tests := []struct {
		path     string
		expected bool
	}{
		{".git/config", true},
		{"vendor/package", true},
		{"node_modules/module", true},
		{"file.tmp", true},
		{"file.log", true},
		{"src/main.go", false},
		{"README.md", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := watcher.shouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesWatchPatterns(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"README.md", true},
		{"config.yaml", true},
		{"data.json", true},
		{"src/main.go", true},
		{"file.txt", false},
		{"image.png", false},
		{"video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := watcher.matchesWatchPatterns(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapOperation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	tests := []struct {
		op       uint32
		expected string
	}{
		{1, "create"},   // fsnotify.Create
		{2, "write"},    // fsnotify.Write
		{4, "remove"},   // fsnotify.Remove
		{8, "rename"},   // fsnotify.Rename
		{16, "chmod"},   // fsnotify.Chmod
		{32, "unknown"}, // Unknown operation
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Create a mock fsnotify.Op with the given value
			result := watcher.mapOperation(fsnotify.Op(tt.op))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWatcherStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	// Initial stats should be zero
	assert.Equal(t, int64(0), watcher.stats.TotalEvents)
	assert.Equal(t, int64(0), watcher.stats.BatchesProcessed)
	assert.Equal(t, int64(0), watcher.stats.FilesModified)
	assert.Equal(t, int64(0), watcher.stats.ErrorCount)
}

func TestValidateRepositoryPath(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "repo-sync-test")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err = os.MkdirAll(gitDir, 0o755)
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid repository",
			path:    tempDir,
			wantErr: false,
		},
		{
			name:    "non-existent path",
			path:    "/path/that/does/not/exist",
			wantErr: true,
		},
		{
			name:    "not a git repository",
			path:    os.TempDir(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepositoryPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Contains(t, config.WatchPatterns, "**/*.go")
	assert.Contains(t, config.IgnorePatterns, ".git/**")
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 5*time.Second, config.BatchTimeout)
	assert.False(t, config.Bidirectional)
	assert.Equal(t, "manual", config.ConflictStrategy)
	assert.Equal(t, "origin", config.RemoteName)
	assert.False(t, config.AutoCommit)
}

func TestCalculateChecksum(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	// Create a temporary file for checksum testing
	tempFile, err := os.CreateTemp("", "checksum-test")
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile.Name()) }()

	testContent := "Hello, World!"
	_, err = tempFile.WriteString(testContent)
	require.NoError(t, err)
	_ = tempFile.Close()

	checksum, err := watcher.calculateChecksum(tempFile.Name())
	assert.NoError(t, err)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 64) // SHA256 produces 64-character hex string

	// Calculate again to ensure consistency
	checksum2, err := watcher.calculateChecksum(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, checksum, checksum2)
}

func TestMapOperationWithFsnotify(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultConfig()

	watcher, err := NewRepositoryWatcher(logger, config)
	require.NoError(t, err)

	defer func() { _ = watcher.Close() }()

	tests := []struct {
		op       fsnotify.Op
		expected string
	}{
		{fsnotify.Create, "create"},
		{fsnotify.Write, "write"},
		{fsnotify.Remove, "remove"},
		{fsnotify.Rename, "rename"},
		{fsnotify.Chmod, "chmod"},
		{fsnotify.Op(64), "unknown"}, // Unknown operation
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := watcher.mapOperation(tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileChangeEventCreation(t *testing.T) {
	event := FileChangeEvent{
		Path:        "/test/file.go",
		Operation:   "write",
		IsDirectory: false,
		Timestamp:   time.Now(),
		Size:        1024,
		Checksum:    "abc123",
	}

	assert.Equal(t, "/test/file.go", event.Path)
	assert.Equal(t, "write", event.Operation)
	assert.False(t, event.IsDirectory)
	assert.Equal(t, int64(1024), event.Size)
	assert.Equal(t, "abc123", event.Checksum)
}

func TestFileChangeBatchCreation(t *testing.T) {
	events := []FileChangeEvent{
		{Path: "/test/file1.go", Operation: "write", Timestamp: time.Now()},
		{Path: "/test/file2.go", Operation: "create", Timestamp: time.Now()},
	}

	batch := FileChangeBatch{
		Events:      events,
		BatchID:     "test-batch-1",
		StartTime:   time.Now().Add(-5 * time.Second),
		EndTime:     time.Now(),
		TotalEvents: len(events),
	}

	assert.Len(t, batch.Events, 2)
	assert.Equal(t, "test-batch-1", batch.BatchID)
	assert.Equal(t, 2, batch.TotalEvents)
}

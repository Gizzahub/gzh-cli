//nolint:testpackage // White-box testing needed for internal function access
package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemImpl_MkdirAll(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "create directory with default permissions",
			path:    "test-dir",
			perm:    0o755,
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    "test/nested/dir",
			perm:    0o755,
			wantErr: false,
		},
		{
			name:    "create directory with custom permissions",
			path:    "custom-perm-dir",
			perm:    0o700,
			wantErr: false,
		},
		{
			name:    "create directory with empty path",
			path:    "",
			perm:    0o755,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			fs := NewFileSystem(nil, nil)
			targetPath := filepath.Join(tmpDir, tt.path)

			err = fs.MkdirAll(context.Background(), targetPath, tt.perm)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify directory was created
				info, err := os.Stat(targetPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			}
		})
	}
}

func TestFileSystemImpl_ReadFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     string
		createFile  bool
		wantErr     bool
		wantContent string
	}{
		{
			name:        "read existing file",
			filename:    "test.txt",
			content:     "Hello, World!",
			createFile:  true,
			wantErr:     false,
			wantContent: "Hello, World!",
		},
		{
			name:        "read non-existent file",
			filename:    "non-existent.txt",
			content:     "",
			createFile:  false,
			wantErr:     true,
			wantContent: "",
		},
		{
			name:        "read empty file",
			filename:    "empty.txt",
			content:     "",
			createFile:  true,
			wantErr:     false,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			fs := NewFileSystem(nil, nil)
			targetPath := filepath.Join(tmpDir, tt.filename)

			// Create test file if needed
			if tt.createFile {
				err := os.WriteFile(targetPath, []byte(tt.content), 0o644)
				require.NoError(t, err)
			}

			content, err := fs.ReadFile(context.Background(), targetPath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, content)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantContent, string(content))
			}
		})
	}
}

func TestFileSystemImpl_WriteFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content []byte
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "write new file",
			path:    "new-file.txt",
			content: []byte("New content"),
			perm:    0o644,
			wantErr: false,
		},
		{
			name:    "overwrite existing file",
			path:    "existing.txt",
			content: []byte("Updated content"),
			perm:    0o644,
			wantErr: false,
		},
		{
			name:    "write file in nested directory",
			path:    "nested/dir/file.txt",
			content: []byte("Nested content"),
			perm:    0o644,
			wantErr: false,
		},
		{
			name:    "write empty file",
			path:    "empty.txt",
			content: []byte{},
			perm:    0o644,
			wantErr: false,
		},
		{
			name:    "write file with empty path",
			path:    "",
			content: []byte("content"),
			perm:    0o644,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			fs := NewFileSystem(nil, nil)
			targetPath := filepath.Join(tmpDir, tt.path)

			// Create existing file for overwrite test
			if tt.name == "overwrite existing file" {
				dir := filepath.Dir(targetPath)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Errorf("failed to create directory: %v", err)
				}
				if err := os.WriteFile(targetPath, []byte("Original content"), 0o644); err != nil {
					t.Errorf("failed to write test file: %v", err)
				}
			}

			err = fs.WriteFile(context.Background(), targetPath, tt.content, tt.perm)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify file was written
				content, err := os.ReadFile(targetPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.content, content)
			}
		})
	}
}

func TestFileSystemImpl_RemoveAll(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		createType string // "file", "dir", "none"
		wantErr    bool
	}{
		{
			name:       "delete existing file",
			path:       "test.txt",
			createType: "file",
			wantErr:    false,
		},
		{
			name:       "delete existing directory",
			path:       "test-dir",
			createType: "dir",
			wantErr:    false,
		},
		{
			name:       "delete non-existent path",
			path:       "non-existent",
			createType: "none",
			wantErr:    false, // os.RemoveAll doesn't error on non-existent paths
		},
		{
			name:       "delete with empty path",
			path:       "",
			createType: "none",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			fs := NewFileSystem(nil, nil)
			targetPath := filepath.Join(tmpDir, tt.path)

			// Create test file or directory if needed
			switch tt.createType {
			case "file":
				err := os.WriteFile(targetPath, []byte("test content"), 0o644)
				require.NoError(t, err)
			case "dir":
				err := os.MkdirAll(targetPath, 0o755)
				require.NoError(t, err)
			}

			err = fs.RemoveAll(context.Background(), targetPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify path was deleted
				_, err := os.Stat(targetPath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestFileSystemImpl_Exists(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		createType string // "file", "dir", "none"
		want       bool
	}{
		{
			name:       "check existing file",
			path:       "test.txt",
			createType: "file",
			want:       true,
		},
		{
			name:       "check existing directory",
			path:       "test-dir",
			createType: "dir",
			want:       true,
		},
		{
			name:       "check non-existent path",
			path:       "non-existent",
			createType: "none",
			want:       false,
		},
		{
			name:       "check empty path",
			path:       "",
			createType: "none",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			fs := NewFileSystem(nil, nil)
			targetPath := filepath.Join(tmpDir, tt.path)

			// Create test file or directory if needed
			switch tt.createType {
			case "file":
				err := os.WriteFile(targetPath, []byte("test content"), 0o644)
				require.NoError(t, err)
			case "dir":
				err := os.MkdirAll(targetPath, 0o755)
				require.NoError(t, err)
			}

			exists := fs.Exists(context.Background(), targetPath)
			assert.Equal(t, tt.want, exists)
		})
	}
}

func TestFileSystemImpl_ListDir(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		wantEntries []string
		wantErr     bool
	}{
		{
			name: "list directory with files",
			setup: func(dir string) error {
				files := []string{"file1.txt", "file2.txt", "file3.md"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte("content"), 0o644); err != nil {
						return err
					}
				}
				return nil
			},
			wantEntries: []string{"file1.txt", "file2.txt", "file3.md"},
			wantErr:     false,
		},
		{
			name: "list directory with subdirectories",
			setup: func(dir string) error {
				dirs := []string{"subdir1", "subdir2"}
				for _, d := range dirs {
					if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
						return err
					}
				}
				return nil
			},
			wantEntries: []string{"subdir1", "subdir2"},
			wantErr:     false,
		},
		{
			name:        "list empty directory",
			setup:       func(_ string) error { return nil },
			wantEntries: []string{},
			wantErr:     false,
		},
		{
			name:        "list non-existent directory",
			setup:       nil,
			wantEntries: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "fs-test-*")
			require.NoError(t, err)

			defer func() { _ = os.RemoveAll(tmpDir) }()

			testDir := filepath.Join(tmpDir, "test-dir")

			// Setup test directory if not testing non-existent
			if tt.name != "list non-existent directory" {
				err := os.MkdirAll(testDir, 0o755)
				require.NoError(t, err)

				if tt.setup != nil {
					err := tt.setup(testDir)
					require.NoError(t, err)
				}
			}

			fs := NewFileSystem(nil, nil)
			entries, err := fs.ListDir(context.Background(), testDir)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, entries)
			} else {
				assert.NoError(t, err)
				// Extract just the names for comparison
				var names []string
				for _, entry := range entries {
					names = append(names, entry.Name)
				}

				assert.ElementsMatch(t, tt.wantEntries, names)
			}
		})
	}
}

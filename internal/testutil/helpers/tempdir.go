package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory and returns a cleanup function
func TempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to remove temp dir %s: %v", dir, err)
		}
	}

	return dir, cleanup
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, dir, pattern string, content []byte) string {
	t.Helper()

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer file.Close()

	if len(content) > 0 {
		if _, err := file.Write(content); err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
	}

	return file.Name()
}

// CreateTestRepo creates a test git repository structure
func CreateTestRepo(t *testing.T, baseDir string, name string, files map[string]string) string {
	t.Helper()

	repoPath := filepath.Join(baseDir, name)
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	// Create .git directory to simulate a git repo
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create test files
	for path, content := range files {
		fullPath := filepath.Join(repoPath, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", fullPath, err)
		}
	}

	return repoPath
}

// CreateTestConfig creates a test configuration file
func CreateTestConfig(t *testing.T, dir string, content string) string {
	t.Helper()

	configPath := filepath.Join(dir, "test-config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	return configPath
}

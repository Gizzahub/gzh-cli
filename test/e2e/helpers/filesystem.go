package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnvironment represents an isolated test environment.
type TestEnvironment struct {
	t       *testing.T
	TempDir string
	HomeDir string
	WorkDir string
	CLI     *CLIExecutor
}

// NewTestEnvironment creates a new isolated test environment.
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gz-e2e-test-*")
	require.NoError(t, err)

	// Create subdirectories
	homeDir := filepath.Join(tempDir, "home")
	workDir := filepath.Join(tempDir, "work")

	require.NoError(t, os.MkdirAll(homeDir, 0o755))
	require.NoError(t, os.MkdirAll(workDir, 0o755))

	// Find and build binary
	projectRoot, err := FindProjectRoot()
	require.NoError(t, err)

	binaryPath, err := BuildBinary(projectRoot)
	require.NoError(t, err)

	// Create CLI executor
	cli := NewCLIExecutor(binaryPath, workDir)
	cli.SetEnv("HOME", homeDir)
	cli.SetEnv("GZ_CONFIG_DIR", filepath.Join(homeDir, ".config", "gzh-manager"))

	return &TestEnvironment{
		t:       t,
		TempDir: tempDir,
		HomeDir: homeDir,
		WorkDir: workDir,
		CLI:     cli,
	}
}

// Cleanup removes the test environment.
func (env *TestEnvironment) Cleanup() {
	if env.TempDir != "" {
		if err := os.RemoveAll(env.TempDir); err != nil {
			// Log error but don't fail cleanup
			_ = err
		}
	}
}

// CreateFile creates a file with the given content.
func (env *TestEnvironment) CreateFile(relativePath, content string) string {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	dir := filepath.Dir(fullPath)

	require.NoError(env.t, os.MkdirAll(dir, 0o755))
	require.NoError(env.t, os.WriteFile(fullPath, []byte(content), 0o644))

	return fullPath
}

// CreateDir creates a directory.
func (env *TestEnvironment) CreateDir(relativePath string) string {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	require.NoError(env.t, os.MkdirAll(fullPath, 0o755))

	return fullPath
}

// WriteConfig writes a configuration file to the work directory.
func (env *TestEnvironment) WriteConfig(filename, content string) string {
	return env.CreateFile(filename, content)
}

// ReadFile reads the content of a file.
func (env *TestEnvironment) ReadFile(relativePath string) string {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	content, err := os.ReadFile(fullPath)
	require.NoError(env.t, err)

	return string(content)
}

// RunCommand executes a CLI command.
func (env *TestEnvironment) RunCommand(args ...string) *CLIResult {
	return env.CLI.Run(args...)
}

// RunCommandWithInput executes a CLI command with stdin input.
func (env *TestEnvironment) RunCommandWithInput(input string, args ...string) *CLIResult {
	return env.CLI.RunWithInput(input, args...)
}

// SetEnv sets an environment variable for the test.
func (env *TestEnvironment) SetEnv(key, value string) {
	env.CLI.SetEnv(key, value)
}

// SetTimeout sets the command timeout.
func (env *TestEnvironment) SetTimeout(_ string) {
	// Parse timeout string and set on CLI
	// This is a simplified version - you might want to use time.ParseDuration
	env.CLI.SetTimeout(30) // Default to 30 seconds for now
}

// AssertFileExists checks that a file exists.
func (env *TestEnvironment) AssertFileExists(relativePath string) {
	env.t.Helper()
	fullPath := filepath.Join(env.WorkDir, relativePath)
	_, err := os.Stat(fullPath)
	require.NoError(env.t, err, "File should exist: %s", relativePath)
}

// AssertFileNotExists checks that a file does not exist.
func (env *TestEnvironment) AssertFileNotExists(relativePath string) {
	env.t.Helper()
	fullPath := filepath.Join(env.WorkDir, relativePath)
	_, err := os.Stat(fullPath)
	require.True(env.t, os.IsNotExist(err), "File should not exist: %s", relativePath)
}

// AssertFileContains checks that a file contains specific content.
func (env *TestEnvironment) AssertFileContains(relativePath, expectedContent string) {
	env.t.Helper()
	content := env.ReadFile(relativePath)
	require.Contains(env.t, content, expectedContent, "File should contain expected content")
}

// AssertFileNotContains checks that a file does not contain specific content.
func (env *TestEnvironment) AssertFileNotContains(relativePath, unexpectedContent string) {
	env.t.Helper()
	content := env.ReadFile(relativePath)
	require.NotContains(env.t, content, unexpectedContent, "File should not contain unexpected content")
}

// AssertDirectoryExists checks that a directory exists.
func (env *TestEnvironment) AssertDirectoryExists(relativePath string) {
	env.t.Helper()
	fullPath := filepath.Join(env.WorkDir, relativePath)
	info, err := os.Stat(fullPath)
	require.NoError(env.t, err, "Directory should exist: %s", relativePath)
	require.True(env.t, info.IsDir(), "Path should be a directory: %s", relativePath)
}

// AssertDirectoryNotEmpty checks that a directory exists and is not empty.
func (env *TestEnvironment) AssertDirectoryNotEmpty(relativePath string) {
	env.t.Helper()
	env.AssertDirectoryExists(relativePath)

	fullPath := filepath.Join(env.WorkDir, relativePath)
	entries, err := os.ReadDir(fullPath)
	require.NoError(env.t, err)
	require.Greater(env.t, len(entries), 0, "Directory should not be empty: %s", relativePath)
}

// ListFiles lists all files in a directory recursively.
func (env *TestEnvironment) ListFiles(relativePath string) []string {
	fullPath := filepath.Join(env.WorkDir, relativePath)

	var files []string

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return err
			}

			files = append(files, relPath)
		}

		return nil
	})

	require.NoError(env.t, err)

	return files
}

// CreateGitRepo creates a minimal git repository for testing.
func (env *TestEnvironment) CreateGitRepo(relativePath string) string {
	repoPath := env.CreateDir(relativePath)

	// Initialize git repo
	gitDir := filepath.Join(repoPath, ".git")
	require.NoError(env.t, os.MkdirAll(gitDir, 0o755))

	// Create minimal git files
	env.CreateFile(filepath.Join(relativePath, "README.md"), "# Test Repository\n")
	env.CreateFile(filepath.Join(relativePath, ".gitignore"), "*.log\n")

	return repoPath
}

// CreateConfigDir creates the configuration directory structure.
func (env *TestEnvironment) CreateConfigDir() string {
	configDir := filepath.Join(env.HomeDir, ".config", "gzh-manager")
	require.NoError(env.t, os.MkdirAll(configDir, 0o755))

	return configDir
}

// GetWorkPath returns the full path for a relative work directory path.
func (env *TestEnvironment) GetWorkPath(relativePath string) string {
	return filepath.Join(env.WorkDir, relativePath)
}

// GetHomePath returns the full path for a relative home directory path.
func (env *TestEnvironment) GetHomePath(relativePath string) string {
	return filepath.Join(env.HomeDir, relativePath)
}

// CopyFile copies a file from source to destination.
func (env *TestEnvironment) CopyFile(src, dst string) {
	srcPath := filepath.Join(env.WorkDir, src)
	dstPath := filepath.Join(env.WorkDir, dst)

	srcContent, err := os.ReadFile(srcPath)
	require.NoError(env.t, err)

	dstDir := filepath.Dir(dstPath)
	require.NoError(env.t, os.MkdirAll(dstDir, 0o755))
	require.NoError(env.t, os.WriteFile(dstPath, srcContent, 0o644))
}

// CreateSymlink creates a symbolic link.
func (env *TestEnvironment) CreateSymlink(target, link string) {
	targetPath := filepath.Join(env.WorkDir, target)
	linkPath := filepath.Join(env.WorkDir, link)

	linkDir := filepath.Dir(linkPath)
	require.NoError(env.t, os.MkdirAll(linkDir, 0o755))
	require.NoError(env.t, os.Symlink(targetPath, linkPath))
}

// SetFilePermissions sets file permissions.
func (env *TestEnvironment) SetFilePermissions(relativePath string, mode os.FileMode) {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	require.NoError(env.t, os.Chmod(fullPath, mode))
}

// CountFiles counts files in a directory (non-recursive).
func (env *TestEnvironment) CountFiles(relativePath string) int {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	entries, err := os.ReadDir(fullPath)
	require.NoError(env.t, err)

	count := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count
}

// CountDirectories counts directories in a directory (non-recursive).
func (env *TestEnvironment) CountDirectories(relativePath string) int {
	fullPath := filepath.Join(env.WorkDir, relativePath)
	entries, err := os.ReadDir(fullPath)
	require.NoError(env.t, err)

	count := 0

	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}

	return count
}

// HasFileExtension checks if any files in directory have the given extension.
func (env *TestEnvironment) HasFileExtension(relativePath, extension string) bool {
	files := env.ListFiles(relativePath)
	for _, file := range files {
		if strings.HasSuffix(file, extension) {
			return true
		}
	}

	return false
}

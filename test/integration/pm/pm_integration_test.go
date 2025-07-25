package pm_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPMIntegration runs package manager integration tests in Docker containers
func TestPMIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		dockerfile string
		osName     string
		skipTests  []string // Tests to skip for specific OS
	}{
		{
			name:       "Ubuntu 22.04",
			dockerfile: "Dockerfile.ubuntu",
			osName:     "ubuntu",
			skipTests:  []string{},
		},
		{
			name:       "Alpine Linux",
			dockerfile: "Dockerfile.alpine",
			osName:     "alpine",
			skipTests:  []string{"TestBrewInstall", "TestNvmInstall", "TestRbenvInstall"}, // Limited support
		},
		{
			name:       "Fedora 39",
			dockerfile: "Dockerfile.fedora",
			osName:     "fedora",
			skipTests:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Build gz binary
			gzBinaryPath := buildGzBinary(t)

			// Create container
			container := createTestContainer(t, ctx, tt.dockerfile, gzBinaryPath)
			defer func() {
				if err := container.Terminate(ctx); err != nil {
					t.Logf("Failed to terminate container: %v", err)
				}
			}()

			// Run tests in container
			t.Run("Bootstrap", func(t *testing.T) {
				if !contains(tt.skipTests, "TestBootstrap") {
					testBootstrap(t, ctx, container)
				}
			})

			t.Run("BrewInstall", func(t *testing.T) {
				if !contains(tt.skipTests, "TestBrewInstall") {
					testBrewInstall(t, ctx, container)
				}
			})

			t.Run("AsdfInstall", func(t *testing.T) {
				if !contains(tt.skipTests, "TestAsdfInstall") {
					testAsdfInstall(t, ctx, container)
				}
			})

			t.Run("NpmPackages", func(t *testing.T) {
				if !contains(tt.skipTests, "TestNpmPackages") {
					testNpmPackages(t, ctx, container)
				}
			})

			t.Run("VersionCoordination", func(t *testing.T) {
				if !contains(tt.skipTests, "TestVersionCoordination") {
					testVersionCoordination(t, ctx, container)
				}
			})

			t.Run("Export", func(t *testing.T) {
				if !contains(tt.skipTests, "TestExport") {
					testExport(t, ctx, container)
				}
			})
		})
	}
}

// buildGzBinary builds the gz binary for testing
func buildGzBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "gz")

	// Build the binary
	cmd := fmt.Sprintf("go build -o %s ../../..", binaryPath)
	if err := runCommand("bash", "-c", cmd); err != nil {
		t.Fatalf("Failed to build gz binary: %v", err)
	}

	return binaryPath
}

// createTestContainer creates a Docker container for testing
func createTestContainer(t *testing.T, ctx context.Context, dockerfile string, gzBinaryPath string) testcontainers.Container {
	t.Helper()

	// Get current directory
	pwd, err := os.Getwd()
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    pwd,
			Dockerfile: dockerfile,
		},
		Cmd:        []string{"sleep", "infinity"}, // Keep container running
		WaitingFor: wait.ForLog("").WithStartupTimeout(5 * time.Minute),
		Mounts: []testcontainers.ContainerMount{
			{
				Source:   testcontainers.GenericBindMountSource{HostPath: gzBinaryPath},
				Target:   "/usr/local/bin/gz",
				ReadOnly: true,
			},
			{
				Source:   testcontainers.GenericBindMountSource{HostPath: filepath.Join(pwd, "../../../test/fixtures/pm")},
				Target:   "/home/testuser/.gzh/pm",
				ReadOnly: true,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Make gz executable
	_, _, err = container.Exec(ctx, []string{"chmod", "+x", "/usr/local/bin/gz"})
	require.NoError(t, err)

	return container
}

// Test functions

func testBootstrap(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Check which package managers need installation
	code, reader, err := container.Exec(ctx, []string{"sudo", "-u", "testuser", "gz", "pm", "bootstrap", "--check"})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("Bootstrap check output: %s", string(output))

	// The command might return non-zero if managers are missing, which is expected
	if code != 0 {
		t.Logf("Some package managers need installation (exit code: %d)", code)
	}
}

func testBrewInstall(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Test brew package installation
	code, reader, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"gz pm install --manager brew",
	})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code, "Command failed with output: %s", string(output))

	// Verify packages were installed
	code, reader, err = container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"brew list",
	})
	require.NoError(t, err)

	output, err = io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(output), "jq")
	assert.Contains(t, string(output), "tree")
}

func testAsdfInstall(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Install asdf plugins first
	code, _, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"asdf plugin add nodejs || true",
	})
	require.NoError(t, err)

	// Test asdf package installation
	code, reader, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"gz pm install --manager asdf",
	})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("ASDF install output: %s", string(output))

	// Verify Node.js was installed
	code, reader, err = container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"asdf list nodejs",
	})
	require.NoError(t, err)

	output, err = io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	t.Logf("ASDF list output: %s", string(output))
}

func testNpmPackages(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Ensure Node.js is available
	code, _, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"which node || echo 'Node.js not found'",
	})
	require.NoError(t, err)

	if code != 0 {
		t.Skip("Node.js not available, skipping npm test")
	}

	// Test npm package installation
	code, reader, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"gz pm install --manager npm",
	})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("NPM install output: %s", string(output))

	// Verify packages were installed
	code, reader, err = container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"npm list -g --depth=0",
	})
	require.NoError(t, err)

	output, err = io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	t.Logf("NPM list output: %s", string(output))
}

func testVersionCoordination(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Test version synchronization check
	code, reader, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"gz pm sync-versions --check",
	})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("Version sync check output: %s", string(output))

	// The command might return non-zero if versions are mismatched
	if code != 0 {
		t.Logf("Version mismatch detected (exit code: %d)", code)

		// Try to fix version mismatches
		code, reader, err = container.Exec(ctx, []string{
			"sudo", "-u", "testuser", "bash", "-l", "-c",
			"gz pm sync-versions --fix",
		})
		require.NoError(t, err)

		output, err = io.ReadAll(reader)
		require.NoError(t, err)
		t.Logf("Version sync fix output: %s", string(output))
	}
}

func testExport(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Test configuration export
	code, reader, err := container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"gz pm export --all",
	})
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code, "Export failed with output: %s", string(output))

	// Verify export files were created
	code, reader, err = container.Exec(ctx, []string{
		"sudo", "-u", "testuser", "bash", "-l", "-c",
		"ls -la ~/.gzh/pm/",
	})
	require.NoError(t, err)

	output, err = io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	t.Logf("Export directory contents: %s", string(output))
}

// Helper functions

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

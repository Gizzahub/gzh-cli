package helpers

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGitRepoType(t *testing.T) {
	// setup
	var cmd *exec.Cmd

	// rmp tmp
	// Clean up and create directories using Go's built-in functions for cross-platform compatibility
	if err := os.RemoveAll("tmp"); err != nil {
		t.Logf("Warning: failed to remove tmp dir: %v", err)
	}

	if err := os.MkdirAll("tmp/git-commit0", 0o755); err != nil {
		t.Fatalf("Failed to create tmp/git-commit0: %v", err)
	}
	if err := os.MkdirAll("tmp/git-commit2", 0o755); err != nil {
		t.Fatalf("Failed to create tmp/git-commit2: %v", err)
	}
	if err := os.MkdirAll("tmp/nongit", 0o755); err != nil {
		t.Fatalf("Failed to create tmp/nongit: %v", err)
	}

	cmd = exec.Command("git", "-C", "tmp/git-commit0", "init")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git init failed: %v", err)
	}
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "init")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git init failed: %v", err)
	}
	// Create test file using Go's built-in function for cross-platform compatibility
	if err := os.WriteFile("tmp/git-commit2/test", []byte{}, 0o644); err != nil {
		t.Logf("Warning: failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "-C", "tmp/git-commit2", "add", ".")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git add failed: %v", err)
	}
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "commit", "-m", "test1")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git commit failed: %v", err)
	}
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "add", ".")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git add failed: %v", err)
	}
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "commit", "-m", "test2")
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: git commit failed: %v", err)
	}

	println("============")

	res, _ := CheckGitRepoType("tmp/git-commit0")
	println("tmp/git-commit0:", res)
	assert.Equal(t, "empty", res, "they should be equal")

	res, _ = CheckGitRepoType("tmp/git-commit2")
	println("tmp/git-commit2:", res)
	assert.Equal(t, "normal", res, "they should be equal")

	res, _ = CheckGitRepoType("tmp/nongit")
	println("tmp/nongit:", res)
	assert.Equal(t, "none", res, "they should be equal")
}

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
	os.RemoveAll("tmp")

	os.MkdirAll("tmp/git-commit0", 0o755)
	os.MkdirAll("tmp/git-commit2", 0o755)
	os.MkdirAll("tmp/nongit", 0o755)

	cmd = exec.Command("git", "-C", "tmp/git-commit0", "init")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "init")
	cmd.Run()
	// Create test file using Go's built-in function for cross-platform compatibility
	os.WriteFile("tmp/git-commit2/test", []byte{}, 0o644)

	cmd = exec.Command("git", "-C", "tmp/git-commit2", "add", ".")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "commit", "-m", "test1")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "add", ".")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "commit", "-m", "test2")
	cmd.Run()

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

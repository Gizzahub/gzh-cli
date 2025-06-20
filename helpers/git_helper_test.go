package helpers

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGitRepoType(t *testing.T) {
	// setup
	var cmd *exec.Cmd

	// rmp tmp
	cmd = exec.Command("rm", "-rf", "tmp")
	cmd.Run()

	cmd = exec.Command("mkdir", "-p", "tmp/git-commit0")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "tmp/git-commit2")
	cmd.Run()
	cmd = exec.Command("mkdir", "-p", "tmp/nongit")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit0", "init")
	cmd.Run()
	cmd = exec.Command("git", "-C", "tmp/git-commit2", "init")
	cmd.Run()
	cmd = exec.Command("touch", "tmp/git-commit2/test")
	cmd.Run()
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

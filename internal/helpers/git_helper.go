package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CheckGitRepoType(dir string) (string, error) {
	// Check if .git directory exists
	if _, err := os.Stat(fmt.Sprintf("%s/.git", dir)); os.IsNotExist(err) {
		return "none", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to access directory: %v", err)
	}

	// git rev-list --count HEAD 2>/dev/null || echo 0
	// Check if there are any commits in the repository
	cmd := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "empty", nil
	}

	commitCount := strings.TrimSpace(string(output))
	if commitCount == "0" {
		return "empty", nil
	}

	return "normal", nil
}

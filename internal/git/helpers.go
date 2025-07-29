// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Repository type constants.
const (
	RepoTypeNone   = "none"   // Not a Git repository
	RepoTypeEmpty  = "empty"  // Git repository with no commits
	RepoTypeNormal = "normal" // Git repository with commits
)

// CheckGitRepoType checks the type of a Git repository.
// Returns "none" if not a Git repo, "empty" if no commits, "normal" if has commits.
func CheckGitRepoType(dir string) (string, error) {
	// Check if .git directory exists
	if _, err := os.Stat(fmt.Sprintf("%s/.git", dir)); os.IsNotExist(err) {
		return RepoTypeNone, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to access directory: %w", err)
	}

	// git rev-list --count HEAD 2>/dev/null || echo 0
	// Check if there are any commits in the repository
	cmd := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return RepoTypeEmpty, err
	}

	commitCount := strings.TrimSpace(string(output))
	if commitCount == "0" {
		return RepoTypeEmpty, nil
	}

	return RepoTypeNormal, nil
}

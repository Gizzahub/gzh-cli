// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SubmoduleManager manages Git submodules for testing complex repository structures.
type SubmoduleManager struct {
	timeout time.Duration
}

// SubmoduleConfig represents submodule configuration.
type SubmoduleConfig struct {
	Name   string // Submodule name
	Path   string // Path within parent repo
	URL    string // URL or path to submodule repo
	Branch string // Branch to track (optional)
}

// NewSubmoduleManager creates a new SubmoduleManager instance.
func NewSubmoduleManager() *SubmoduleManager {
	return &SubmoduleManager{
		timeout: 60 * time.Second,
	}
}

// CreateRepositoryWithSubmodules creates a repository with submodules.
func (sm *SubmoduleManager) CreateRepositoryWithSubmodules(ctx context.Context, parentPath string, submodules []SubmoduleConfig) error {
	// Initialize parent repository
	if err := sm.initializeRepo(ctx, parentPath); err != nil {
		return fmt.Errorf("failed to initialize parent repository: %w", err)
	}

	// Create each submodule
	for _, sub := range submodules {
		if err := sm.addSubmodule(ctx, parentPath, sub); err != nil {
			return fmt.Errorf("failed to add submodule %s: %w", sub.Name, err)
		}
	}

	return nil
}

// CreateNestedSubmodules creates a repository with nested submodules.
func (sm *SubmoduleManager) CreateNestedSubmodules(ctx context.Context, basePath string) error {
	// Create main repository
	mainPath := filepath.Join(basePath, "main-repo")

	// Create submodule repositories first
	libPath := filepath.Join(basePath, "lib-repo")
	utilsPath := filepath.Join(basePath, "utils-repo")
	nestedLibPath := filepath.Join(basePath, "nested-lib-repo")

	// Create lib repository
	if err := sm.createLibRepository(ctx, libPath); err != nil {
		return fmt.Errorf("failed to create lib repository: %w", err)
	}

	// Create utils repository with nested submodule
	if err := sm.createUtilsRepository(ctx, utilsPath, nestedLibPath); err != nil {
		return fmt.Errorf("failed to create utils repository: %w", err)
	}

	// Create main repository with submodules
	submodules := []SubmoduleConfig{
		{
			Name: "lib",
			Path: "vendor/lib",
			URL:  libPath,
		},
		{
			Name: "utils",
			Path: "vendor/utils",
			URL:  utilsPath,
		},
	}

	return sm.CreateRepositoryWithSubmodules(ctx, mainPath, submodules)
}

// UpdateSubmodules updates all submodules to latest commits.
func (sm *SubmoduleManager) UpdateSubmodules(ctx context.Context, repoPath string) error {
	if err := sm.runGitCommand(ctx, repoPath, "submodule", "update", "--remote"); err != nil {
		return fmt.Errorf("failed to update submodules: %w", err)
	}
	return nil
}

// GetSubmoduleStatus returns the status of all submodules.
func (sm *SubmoduleManager) GetSubmoduleStatus(ctx context.Context, repoPath string) ([]SubmoduleStatus, error) {
	cmd := exec.CommandContext(ctx, "git", "submodule", "status")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get submodule status: %w", err)
	}

	return sm.parseSubmoduleStatus(string(output)), nil
}

// SubmoduleStatus represents the status of a submodule.
type SubmoduleStatus struct {
	Name   string
	Path   string
	Hash   string
	Status string // "", "+", "-", "U" for clean, ahead, behind, conflict
}

// addSubmodule adds a submodule to the repository.
func (sm *SubmoduleManager) addSubmodule(ctx context.Context, parentPath string, config SubmoduleConfig) error {
	// Create the submodule repository if it doesn't exist
	if _, err := os.Stat(config.URL); os.IsNotExist(err) {
		if err := sm.createSubmoduleRepo(ctx, config.URL, config.Name); err != nil {
			return fmt.Errorf("failed to create submodule repository: %w", err)
		}
	}

	// Add submodule to parent repository
	args := []string{"submodule", "add"}
	if config.Branch != "" {
		args = append(args, "-b", config.Branch)
	}
	args = append(args, config.URL, config.Path)

	if err := sm.runGitCommand(ctx, parentPath, args...); err != nil {
		return fmt.Errorf("failed to add submodule: %w", err)
	}

	// Commit the submodule addition
	commitMsg := fmt.Sprintf("Add %s submodule at %s", config.Name, config.Path)
	if err := sm.runGitCommand(ctx, parentPath, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit submodule addition: %w", err)
	}

	return nil
}

// createSubmoduleRepo creates a simple repository for use as a submodule.
func (sm *SubmoduleManager) createSubmoduleRepo(ctx context.Context, repoPath, name string) error {
	if err := sm.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize submodule repo: %w", err)
	}

	// Create some content specific to the submodule
	files := map[string]string{
		fmt.Sprintf("%s.go", name): fmt.Sprintf("package %s\n\n// %s provides functionality\nfunc %sFunction() {\n\t// implementation\n}\n", name, strings.ToUpper(name[:1])+strings.ToLower(name[1:]), strings.ToUpper(name[:1])+strings.ToLower(name[1:])), // SA1019 수정: strings.Title 대신 수동 변환
		"VERSION":                  "1.0.0\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := sm.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	if err := sm.runGitCommand(ctx, repoPath, "commit", "-m", fmt.Sprintf("Initial %s implementation", name)); err != nil {
		return fmt.Errorf("failed to commit initial content: %w", err)
	}

	return nil
}

// createLibRepository creates a library repository.
func (sm *SubmoduleManager) createLibRepository(ctx context.Context, repoPath string) error {
	if err := sm.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize lib repository: %w", err)
	}

	files := map[string]string{
		"lib.go": `package lib

import "fmt"

// LibFunction provides library functionality
func LibFunction() {
	fmt.Println("Library function called")
}

// Version returns the library version
func Version() string {
	return "1.0.0"
}
`,
		"README.md": "# Library\n\nA test library for submodule testing.\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := sm.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	if err := sm.runGitCommand(ctx, repoPath, "commit", "-m", "Add library implementation"); err != nil {
		return fmt.Errorf("failed to commit lib content: %w", err)
	}

	return nil
}

// createUtilsRepository creates a utils repository with its own submodule.
func (sm *SubmoduleManager) createUtilsRepository(ctx context.Context, utilsPath, nestedLibPath string) error {
	// Create the nested lib first
	if err := sm.createNestedLibRepository(ctx, nestedLibPath); err != nil {
		return fmt.Errorf("failed to create nested lib: %w", err)
	}

	// Initialize utils repository
	if err := sm.initializeRepo(ctx, utilsPath); err != nil {
		return fmt.Errorf("failed to initialize utils repository: %w", err)
	}

	// Add some utils content
	files := map[string]string{
		"utils.go": `package utils

import "fmt"

// UtilFunction provides utility functionality
func UtilFunction() {
	fmt.Println("Utility function called")
}

// Helper returns a helper string
func Helper() string {
	return "helper"
}
`,
		"README.md": "# Utils\n\nUtilities with nested dependencies.\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(utilsPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := sm.runGitCommand(ctx, utilsPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	if err := sm.runGitCommand(ctx, utilsPath, "commit", "-m", "Add utils implementation"); err != nil {
		return fmt.Errorf("failed to commit utils content: %w", err)
	}

	// Add nested submodule
	nestedConfig := SubmoduleConfig{
		Name: "nested-lib",
		Path: "deps/nested-lib",
		URL:  nestedLibPath,
	}

	if err := sm.addSubmodule(ctx, utilsPath, nestedConfig); err != nil {
		return fmt.Errorf("failed to add nested submodule: %w", err)
	}

	return nil
}

// createNestedLibRepository creates a nested library repository.
func (sm *SubmoduleManager) createNestedLibRepository(ctx context.Context, repoPath string) error {
	if err := sm.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize nested lib repository: %w", err)
	}

	content := `package nestedlib

// NestedFunction provides nested library functionality
func NestedFunction() string {
	return "nested functionality"
}
`

	filePath := filepath.Join(repoPath, "nestedlib.go")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create nested lib file: %w", err)
	}

	if err := sm.runGitCommand(ctx, repoPath, "add", "nestedlib.go"); err != nil {
		return fmt.Errorf("failed to add nested lib file: %w", err)
	}

	if err := sm.runGitCommand(ctx, repoPath, "commit", "-m", "Add nested library implementation"); err != nil {
		return fmt.Errorf("failed to commit nested lib content: %w", err)
	}

	return nil
}

// parseSubmoduleStatus parses git submodule status output.
func (sm *SubmoduleManager) parseSubmoduleStatus(output string) []SubmoduleStatus {
	var statuses []SubmoduleStatus
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: " hash path (description)" or "+hash path (description)"
		status := ""
		if len(line) > 0 && (line[0] == '+' || line[0] == '-' || line[0] == 'U') {
			status = string(line[0])
			line = line[1:]
		} else if len(line) > 0 && line[0] == ' ' {
			line = line[1:]
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			statuses = append(statuses, SubmoduleStatus{
				Hash:   parts[0],
				Path:   parts[1],
				Name:   filepath.Base(parts[1]),
				Status: status,
			})
		}
	}

	return statuses
}

// Helper methods

// initializeRepo initializes a Git repository.
func (sm *SubmoduleManager) initializeRepo(ctx context.Context, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := sm.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	if err := sm.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := sm.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nRepository with submodule support.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := sm.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := sm.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// runGitCommand executes a git command with timeout.
func (sm *SubmoduleManager) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, sm.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}

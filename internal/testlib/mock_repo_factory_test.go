package testlib

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultMockRepoFactory(t *testing.T) {
	factory := NewMockRepoFactory()
	if factory == nil {
		t.Fatal("NewMockRepoFactory() returned nil")
	}
}

func TestCreateBasicRepos(t *testing.T) {
	factory := NewMockRepoFactory()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-repos-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name string
		opts BasicRepoOptions
		want bool // expect success
	}{
		{
			name: "basic repo with data",
			opts: BasicRepoOptions{
				BaseDir:     tempDir,
				RepoName:    "test-basic-1",
				InitialData: true,
				Branches:    []string{"main", "develop"},
			},
			want: true,
		},
		{
			name: "basic repo without data",
			opts: BasicRepoOptions{
				BaseDir:     tempDir,
				RepoName:    "test-basic-2",
				InitialData: false,
				Branches:    []string{"main"},
			},
			want: true,
		},
		{
			name: "invalid - missing base dir",
			opts: BasicRepoOptions{
				BaseDir:     "",
				RepoName:    "test-basic-3",
				InitialData: true,
			},
			want: false,
		},
		{
			name: "invalid - missing repo name",
			opts: BasicRepoOptions{
				BaseDir:     tempDir,
				RepoName:    "",
				InitialData: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := factory.CreateBasicRepos(ctx, tt.opts)

			if tt.want && err != nil {
				t.Errorf("CreateBasicRepos() error = %v, want success", err)
			}

			if !tt.want && err == nil {
				t.Errorf("CreateBasicRepos() expected error, got success")
			}

			// For successful cases, verify the repository was created
			if tt.want && err == nil {
				repoPath := filepath.Join(tt.opts.BaseDir, tt.opts.RepoName)
				gitDir := filepath.Join(repoPath, ".git")

				if _, err := os.Stat(gitDir); os.IsNotExist(err) {
					t.Errorf("Git repository not created at %s", gitDir)
				}

				// Check if README exists when InitialData is true
				if tt.opts.InitialData {
					readmePath := filepath.Join(repoPath, "README.md")
					if _, err := os.Stat(readmePath); os.IsNotExist(err) {
						t.Errorf("README.md not created at %s", readmePath)
					}
				}
			}
		})
	}
}

func TestCreateConflictRepos(t *testing.T) {
	factory := NewMockRepoFactory()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-conflict-repos-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name string
		opts ConflictRepoOptions
		want bool
	}{
		{
			name: "merge conflict repo",
			opts: ConflictRepoOptions{
				BaseDir:      tempDir,
				RepoName:     "conflict-merge",
				ConflictType: "merge",
				LocalChanges: true,
			},
			want: true,
		},
		{
			name: "diverged conflict repo",
			opts: ConflictRepoOptions{
				BaseDir:      tempDir,
				RepoName:     "conflict-diverged",
				ConflictType: "diverged",
				LocalChanges: false,
			},
			want: true,
		},
		{
			name: "invalid conflict type",
			opts: ConflictRepoOptions{
				BaseDir:      tempDir,
				RepoName:     "conflict-invalid",
				ConflictType: "invalid-type",
				LocalChanges: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := factory.CreateConflictRepos(ctx, tt.opts)

			if tt.want && err != nil {
				t.Errorf("CreateConflictRepos() error = %v, want success", err)
			}

			if !tt.want && err == nil {
				t.Errorf("CreateConflictRepos() expected error, got success")
			}

			// For successful cases, verify the repository was created
			if tt.want && err == nil {
				repoPath := filepath.Join(tt.opts.BaseDir, tt.opts.RepoName)
				gitDir := filepath.Join(repoPath, ".git")

				if _, err := os.Stat(gitDir); os.IsNotExist(err) {
					t.Errorf("Git repository not created at %s", gitDir)
				}
			}
		})
	}
}

func TestCreateSpecialRepos(t *testing.T) {
	factory := NewMockRepoFactory()

	tempDir, err := os.MkdirTemp("", "test-special-repos-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := SpecialRepoOptions{
		BaseDir:     tempDir,
		RepoName:    "special-test",
		SpecialType: "lfs",
	}

	err = factory.CreateSpecialRepos(ctx, opts)

	// Should return error as this is not implemented yet (Phase 1C)
	if err == nil {
		t.Error("CreateSpecialRepos() expected error for unimplemented functionality, got success")
	}
}

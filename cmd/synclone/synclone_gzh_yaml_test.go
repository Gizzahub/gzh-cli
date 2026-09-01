// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposync"
)

func TestWriteAndLoadForgeGzhYaml(t *testing.T) {
	dir := t.TempDir()

	repos := []ForgeGzhRepo{
		{Name: "alpha", CloneURL: "https://github.com/acme/alpha.git", Description: "first"},
		{Name: "beta", CloneURL: "https://github.com/acme/beta.git", Private: true},
	}

	if err := WriteForgeGzhYaml(GzhYamlWriteOptions{
		TargetPath:     dir,
		Organization:   "acme",
		Provider:       "github",
		CleanupOrphans: true,
	}, repos); err != nil {
		t.Fatalf("WriteForgeGzhYaml: %v", err)
	}

	gzhPath := filepath.Join(dir, "gzh.yaml")
	if _, err := os.Stat(gzhPath); err != nil {
		t.Fatalf("expected gzh.yaml at %s: %v", gzhPath, err)
	}

	loaded, err := LoadForgeGzhYaml(dir)
	if err != nil {
		t.Fatalf("LoadForgeGzhYaml: %v", err)
	}

	if loaded.Organization != "acme" {
		t.Errorf("organization = %q, want acme", loaded.Organization)
	}
	if loaded.Provider != "github" {
		t.Errorf("provider = %q, want github", loaded.Provider)
	}
	if !loaded.SyncMode.CleanupOrphans {
		t.Error("expected cleanup_orphans true")
	}
	if len(loaded.Repositories) != 2 {
		t.Fatalf("repositories len = %d, want 2", len(loaded.Repositories))
	}
	if loaded.Repositories[0].Name != "alpha" || loaded.Repositories[0].CloneURL == "" {
		t.Errorf("first repo = %+v", loaded.Repositories[0])
	}
	if loaded.GeneratedAt.IsZero() {
		t.Error("generated_at should be set")
	}
	// GeneratedAt should be recent (within last minute)
	if time.Since(loaded.GeneratedAt) > time.Minute {
		t.Errorf("generated_at too old: %v", loaded.GeneratedAt)
	}
}

func TestWriteForgeGzhYaml_CreatesNestedTarget(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "target")
	if err := WriteForgeGzhYaml(GzhYamlWriteOptions{
		TargetPath:   dir,
		Organization: "org",
		Provider:     "gitlab",
	}, nil); err != nil {
		t.Fatalf("WriteForgeGzhYaml nested: %v", err)
	}

	loaded, err := LoadForgeGzhYaml(dir)
	if err != nil {
		t.Fatalf("LoadForgeGzhYaml: %v", err)
	}
	if loaded.Repositories == nil {
		t.Error("repositories should be non-nil empty slice")
	}
	if len(loaded.Repositories) != 0 {
		t.Errorf("repositories len = %d, want 0", len(loaded.Repositories))
	}
}

func TestTryReuseForgeGzhYaml(t *testing.T) {
	dir := t.TempDir()

	if err := WriteForgeGzhYaml(GzhYamlWriteOptions{
		TargetPath:   dir,
		Organization: "acme",
		Provider:     "github",
	}, []ForgeGzhRepo{{Name: "r1", CloneURL: "https://example.com/r1.git"}}); err != nil {
		t.Fatalf("write: %v", err)
	}

	repos, ok, err := TryReuseForgeGzhYaml(dir, "acme", "github")
	if err != nil {
		t.Fatalf("reuse: %v", err)
	}
	if !ok {
		t.Fatal("expected reuse ok")
	}
	if len(repos) != 1 || repos[0].Name != "r1" {
		t.Errorf("repos = %+v", repos)
	}

	// Mismatched org/provider → no reuse
	_, ok, err = TryReuseForgeGzhYaml(dir, "other", "github")
	if err != nil {
		t.Fatalf("mismatch org: %v", err)
	}
	if ok {
		t.Error("expected no reuse for different org")
	}

	_, ok, err = TryReuseForgeGzhYaml(dir, "acme", "gitlab")
	if err != nil {
		t.Fatalf("mismatch provider: %v", err)
	}
	if ok {
		t.Error("expected no reuse for different provider")
	}

	// Missing file
	_, ok, err = TryReuseForgeGzhYaml(filepath.Join(dir, "missing"), "acme", "github")
	if err != nil {
		t.Fatalf("missing: %v", err)
	}
	if ok {
		t.Error("expected no reuse when file missing")
	}
}

func TestGzhYamlReposFromPlan(t *testing.T) {
	plan := reposync.Plan{
		Actions: []reposync.Action{
			{Type: reposync.ActionClone, Repo: reposync.RepoSpec{Name: "a", CloneURL: "https://x/a.git", Description: "A"}},
			{Type: reposync.ActionUpdate, Repo: reposync.RepoSpec{Name: "b", CloneURL: "https://x/b.git"}},
			{Type: reposync.ActionDelete, Repo: reposync.RepoSpec{Name: "orphan", CloneURL: ""}},
			{Type: reposync.ActionClone, Repo: reposync.RepoSpec{Name: "a", CloneURL: "https://x/a-dup.git"}}, // duplicate name
			{Type: reposync.ActionSkip, Repo: reposync.RepoSpec{Name: ""}},                                    // empty name
		},
	}

	repos := GzhReposFromPlan(plan)
	if len(repos) != 2 {
		t.Fatalf("len = %d, want 2 (delete/empty/dup filtered)", len(repos))
	}
	if repos[0].Name != "a" || repos[0].CloneURL != "https://x/a.git" {
		t.Errorf("first = %+v", repos[0])
	}
	if repos[1].Name != "b" {
		t.Errorf("second = %+v", repos[1])
	}
}

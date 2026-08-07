// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposync"
)

// ForgeGzhRepo is a provider-agnostic repository entry written to gzh.yaml.
// Field names align with common clone metadata (name + clone URL required).
type ForgeGzhRepo struct {
	Name        string `yaml:"name"`
	CloneURL    string `yaml:"clone_url"`
	Description string `yaml:"description,omitempty"`
	Private     bool   `yaml:"private,omitempty"`
	Archived    bool   `yaml:"archived,omitempty"`
	Fork        bool   `yaml:"fork,omitempty"`
}

// ForgeGzhYamlConfig is the multi-provider gzh.yaml written by the forge path.
type ForgeGzhYamlConfig struct {
	Organization string         `yaml:"organization"`
	Provider     string         `yaml:"provider"`
	GeneratedAt  time.Time      `yaml:"generated_at"`
	SyncMode     SyncMode       `yaml:"sync_mode"`
	Repositories []ForgeGzhRepo `yaml:"repositories"`
}

// GzhYamlWriteOptions configures forge gzh.yaml generation.
type GzhYamlWriteOptions struct {
	TargetPath     string
	Organization   string
	Provider       string
	CleanupOrphans bool
}

// WriteForgeGzhYaml writes a gzh.yaml for the given repository list.
// It creates the target directory when missing.
func WriteForgeGzhYaml(opts GzhYamlWriteOptions, repos []ForgeGzhRepo) error {
	if opts.TargetPath == "" {
		return fmt.Errorf("target path is required for gzh.yaml")
	}
	if err := os.MkdirAll(opts.TargetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	cfg := ForgeGzhYamlConfig{
		Organization: opts.Organization,
		Provider:     opts.Provider,
		GeneratedAt:  time.Now().UTC(),
		SyncMode: SyncMode{
			CleanupOrphans: opts.CleanupOrphans,
		},
		Repositories: repos,
	}
	if cfg.Repositories == nil {
		cfg.Repositories = []ForgeGzhRepo{}
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal gzh.yaml: %w", err)
	}

	gzhPath := filepath.Join(opts.TargetPath, "gzh.yaml")
	if err := os.WriteFile(gzhPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write gzh.yaml: %w", err)
	}
	return nil
}

// LoadForgeGzhYaml loads a forge-format gzh.yaml from the target directory.
// Returns (nil, error) when the file is missing or invalid.
func LoadForgeGzhYaml(targetPath string) (*ForgeGzhYamlConfig, error) {
	gzhPath := filepath.Join(targetPath, "gzh.yaml")
	data, err := os.ReadFile(gzhPath)
	if err != nil {
		return nil, err
	}

	var cfg ForgeGzhYamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse gzh.yaml: %w", err)
	}
	return &cfg, nil
}

// TryReuseForgeGzhYaml loads existing gzh.yaml when organization and provider match.
// Returns (repos, true, nil) on reuse, (nil, false, nil) when no reusable file.
func TryReuseForgeGzhYaml(targetPath, organization, provider string) ([]ForgeGzhRepo, bool, error) {
	cfg, err := LoadForgeGzhYaml(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		// Unreadable/invalid: treat as no reuse (caller may rewrite).
		return nil, false, nil
	}
	if cfg.Organization != organization || cfg.Provider != provider {
		return nil, false, nil
	}
	return cfg.Repositories, true, nil
}

// GzhReposFromPlan extracts clone/update repo entries from a forge plan.
// Delete/skip-only actions without a clone URL are omitted.
func GzhReposFromPlan(plan reposync.Plan) []ForgeGzhRepo {
	repos := make([]ForgeGzhRepo, 0, len(plan.Actions))
	seen := make(map[string]struct{}, len(plan.Actions))

	for _, action := range plan.Actions {
		if action.Type == reposync.ActionDelete {
			continue
		}
		name := action.Repo.Name
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		repos = append(repos, ForgeGzhRepo{
			Name:        name,
			CloneURL:    action.Repo.CloneURL,
			Description: action.Repo.Description,
		})
	}
	return repos
}

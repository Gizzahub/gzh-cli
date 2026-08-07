// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"testing"

	"github.com/gizzahub/gzh-cli-gitforge/pkg/provider"
	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposync"
)

func TestBuildForgeMetadataFilter(t *testing.T) {
	f := buildForgeMetadataFilter("Go, Python", 5, 100, []string{"CLI", " tools "}, []string{"deprecated"})

	if len(f.Languages) != 2 || f.Languages[0] != "go" || f.Languages[1] != "python" {
		t.Errorf("languages = %v, want [go python]", f.Languages)
	}
	if f.MinStars != 5 || f.MaxStars != 100 {
		t.Errorf("stars = %d/%d, want 5/100", f.MinStars, f.MaxStars)
	}
	if len(f.IncludeTopics) != 2 || f.IncludeTopics[0] != "cli" || f.IncludeTopics[1] != "tools" {
		t.Errorf("include topics = %v", f.IncludeTopics)
	}
	if len(f.ExcludeTopics) != 1 || f.ExcludeTopics[0] != "deprecated" {
		t.Errorf("exclude topics = %v", f.ExcludeTopics)
	}

	empty := buildForgeMetadataFilter("  ,  ", 0, 0, nil, nil)
	if empty.Languages != nil {
		t.Errorf("empty language should yield nil, got %v", empty.Languages)
	}
}

func TestApplyMetadataToPlannerConfig_FilterFlags(t *testing.T) {
	cfg := reposync.ForgePlannerConfig{}
	meta := buildForgeMetadataFilter("Rust", 10, 50, []string{"x"}, []string{"y"})
	applyMetadataToPlannerConfig(&cfg, meta)

	if len(cfg.FilterLanguages) != 1 || cfg.FilterLanguages[0] != "rust" {
		t.Errorf("FilterLanguages = %v", cfg.FilterLanguages)
	}
	if cfg.FilterMinStars != 10 {
		t.Errorf("FilterMinStars = %d, want 10", cfg.FilterMinStars)
	}
	if cfg.FilterMaxStars != 50 {
		t.Errorf("FilterMaxStars = %d, want 50", cfg.FilterMaxStars)
	}
}

func TestBuildForgePlannerConfig_WiresFilters(t *testing.T) {
	opts := &forgeOptions{
		TargetPath:      "/tmp/repos",
		Organization:    "acme",
		IsUser:          false,
		IncludeArchived: true,
		IncludeForks:    true,
		IncludePrivate:  false,
		UseSSH:          true,
		Language:        "Go,TypeScript",
		MinStars:        3,
		MaxStars:        99,
		IncludeTopics:   []string{"cli"},
		ExcludeTopics:   []string{"wip"},
	}

	cfg := buildForgePlannerConfig(opts)

	if cfg.TargetPath != "/tmp/repos" || cfg.Organization != "acme" {
		t.Errorf("path/org not wired: %+v", cfg)
	}
	if cfg.CloneProto != "ssh" {
		t.Errorf("CloneProto = %q, want ssh", cfg.CloneProto)
	}
	if !cfg.IncludeArchived || !cfg.IncludeForks || cfg.IncludePrivate {
		t.Errorf("include flags not wired: archived=%v forks=%v private=%v",
			cfg.IncludeArchived, cfg.IncludeForks, cfg.IncludePrivate)
	}
	if len(cfg.FilterLanguages) != 2 {
		t.Fatalf("FilterLanguages = %v", cfg.FilterLanguages)
	}
	if cfg.FilterMinStars != 3 || cfg.FilterMaxStars != 99 {
		t.Errorf("stars not wired: %d/%d", cfg.FilterMinStars, cfg.FilterMaxStars)
	}
}

func TestForgeCommand_RegistersFilterFlags(t *testing.T) {
	cmd := newSyncCloneForgeCmd(nil)

	for _, name := range []string{"language", "min-stars", "max-stars", "topics", "exclude-topics"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s on forge command", name)
		}
	}

	// Bind sample values through flag set (unit, no network)
	if err := cmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("set language: %v", err)
	}
	if err := cmd.Flags().Set("min-stars", "7"); err != nil {
		t.Fatalf("set min-stars: %v", err)
	}
	if err := cmd.Flags().Set("max-stars", "42"); err != nil {
		t.Fatalf("set max-stars: %v", err)
	}
	if err := cmd.Flags().Set("topics", "cli,tools"); err != nil {
		t.Fatalf("set topics: %v", err)
	}
	if err := cmd.Flags().Set("exclude-topics", "deprecated"); err != nil {
		t.Fatalf("set exclude-topics: %v", err)
	}

	// Flags are bound to the opts captured in newSyncCloneForgeCmd; re-parse via RunE is heavy.
	// Verify lookup values on the flag set instead.
	lang, _ := cmd.Flags().GetString("language")
	minStars, _ := cmd.Flags().GetInt("min-stars")
	maxStars, _ := cmd.Flags().GetInt("max-stars")
	topics, _ := cmd.Flags().GetStringSlice("topics")
	exclude, _ := cmd.Flags().GetStringSlice("exclude-topics")

	if lang != "Go" || minStars != 7 || maxStars != 42 {
		t.Errorf("language/stars flags: lang=%q min=%d max=%d", lang, minStars, maxStars)
	}
	if len(topics) != 2 || topics[0] != "cli" || topics[1] != "tools" {
		t.Errorf("topics flag = %v", topics)
	}
	if len(exclude) != 1 || exclude[0] != "deprecated" {
		t.Errorf("exclude-topics flag = %v", exclude)
	}
}

func TestFilterReposByTopics(t *testing.T) {
	repos := []*provider.Repository{
		{Name: "a", Topics: []string{"cli", "tools"}},
		{Name: "b", Topics: []string{"cli"}},
		{Name: "c", Topics: []string{"web", "deprecated"}},
		{Name: "d", Topics: nil},
		nil,
	}

	t.Run("include AND semantics", func(t *testing.T) {
		got := filterReposByTopics(repos, []string{"cli", "tools"}, nil)
		if len(got) != 1 || got[0].Name != "a" {
			t.Errorf("got %v names", repoNames(got))
		}
	})

	t.Run("exclude any", func(t *testing.T) {
		got := filterReposByTopics(repos, nil, []string{"deprecated"})
		names := repoNames(got)
		if len(names) != 3 { // a,b,d — not c, not nil
			t.Errorf("got %v, want 3 non-deprecated", names)
		}
		for _, n := range names {
			if n == "c" {
				t.Error("excluded repo c still present")
			}
		}
	})

	t.Run("include and exclude", func(t *testing.T) {
		got := filterReposByTopics(repos, []string{"cli"}, []string{"tools"})
		// a has tools → excluded; b has cli only → kept
		if len(got) != 1 || got[0].Name != "b" {
			t.Errorf("got %v", repoNames(got))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got := filterReposByTopics(repos, []string{"CLI"}, nil)
		if len(got) != 2 {
			t.Errorf("got %v, want a and b", repoNames(got))
		}
	})

	t.Run("no filters passthrough", func(t *testing.T) {
		got := filterReposByTopics(repos, nil, nil)
		if len(got) != len(repos) {
			t.Errorf("passthrough len = %d, want %d", len(got), len(repos))
		}
	})
}

func TestTopicFilterProvider(t *testing.T) {
	inner := &stubForgeProvider{
		repos: []*provider.Repository{
			{Name: "keep", Topics: []string{"go"}},
			{Name: "drop", Topics: []string{"java"}},
		},
	}
	wrapped := wrapTopicFilter(inner, []string{"go"}, nil)

	got, err := wrapped.ListOrganizationRepos(context.Background(), "org")
	if err != nil {
		t.Fatalf("list org: %v", err)
	}
	if len(got) != 1 || got[0].Name != "keep" {
		t.Errorf("org filter = %v", repoNames(got))
	}

	got, err = wrapped.ListUserRepos(context.Background(), "user")
	if err != nil {
		t.Fatalf("list user: %v", err)
	}
	if len(got) != 1 || got[0].Name != "keep" {
		t.Errorf("user filter = %v", repoNames(got))
	}

	// No topics → same instance (no wrap)
	same := wrapTopicFilter(inner, nil, nil)
	if same != inner {
		t.Error("expected unwrap when no topic filters")
	}
}

func repoNames(repos []*provider.Repository) []string {
	out := make([]string, 0, len(repos))
	for _, r := range repos {
		if r != nil {
			out = append(out, r.Name)
		}
	}
	return out
}

type stubForgeProvider struct {
	repos []*provider.Repository
}

func (s *stubForgeProvider) Name() string { return "stub" }

func (s *stubForgeProvider) ListOrganizationRepos(context.Context, string) ([]*provider.Repository, error) {
	return s.repos, nil
}

func (s *stubForgeProvider) ListUserRepos(context.Context, string) ([]*provider.Repository, error) {
	return s.repos, nil
}

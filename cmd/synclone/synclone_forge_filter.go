// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"strings"

	"github.com/gizzahub/gzh-cli-gitforge/pkg/provider"
	"github.com/gizzahub/gzh-cli-gitforge/pkg/reposync"
)

// forgeMetadataFilter holds pure filter inputs mapped onto ForgePlannerConfig.
type forgeMetadataFilter struct {
	Languages     []string
	MinStars      int
	MaxStars      int
	IncludeTopics []string
	ExcludeTopics []string
}

// buildForgeMetadataFilter normalizes CLI filter flags into planner-ready values.
// Language accepts a single value or comma-separated list (case-insensitive).
func buildForgeMetadataFilter(language string, minStars, maxStars int, includeTopics, excludeTopics []string) forgeMetadataFilter {
	return forgeMetadataFilter{
		Languages:     parseLanguageList(language),
		MinStars:      minStars,
		MaxStars:      maxStars,
		IncludeTopics: normalizeTopicList(includeTopics),
		ExcludeTopics: normalizeTopicList(excludeTopics),
	}
}

// applyMetadataToPlannerConfig wires language/stars filters onto planner config.
// Topics are applied client-side via topicFilterProvider (not planner-native).
func applyMetadataToPlannerConfig(cfg *reposync.ForgePlannerConfig, f forgeMetadataFilter) {
	if cfg == nil {
		return
	}
	cfg.FilterLanguages = f.Languages
	cfg.FilterMinStars = f.MinStars
	cfg.FilterMaxStars = f.MaxStars
}

// parseLanguageList splits a language flag into lowercase entries.
func parseLanguageList(language string) []string {
	language = strings.TrimSpace(language)
	if language == "" {
		return nil
	}
	parts := strings.Split(language, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeTopicList(topics []string) []string {
	if len(topics) == 0 {
		return nil
	}
	out := make([]string, 0, len(topics))
	for _, t := range topics {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// filterReposByTopics applies include/exclude topic filters (case-insensitive).
// Include uses AND semantics (repo must have all include topics).
// Exclude removes a repo if it has any exclude topic.
func filterReposByTopics(repos []*provider.Repository, includeTopics, excludeTopics []string) []*provider.Repository {
	include := normalizeTopicList(includeTopics)
	exclude := normalizeTopicList(excludeTopics)
	if len(include) == 0 && len(exclude) == 0 {
		return repos
	}

	filtered := make([]*provider.Repository, 0, len(repos))
	for _, repo := range repos {
		if repo == nil {
			continue
		}
		if repoMatchesTopics(repo.Topics, include, exclude) {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

func repoMatchesTopics(repoTopics, include, exclude []string) bool {
	normalizedRepo := normalizeTopicList(repoTopics)

	for _, et := range exclude {
		if containsStringFold(normalizedRepo, et) {
			return false
		}
	}
	if len(include) == 0 {
		return true
	}
	for _, it := range include {
		if !containsStringFold(normalizedRepo, it) {
			return false
		}
	}
	return true
}

func containsStringFold(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// topicFilterProvider wraps a ForgeProvider and filters listed repos by topics.
// Topics live on provider.Repository but are not yet first-class in ForgePlannerConfig.
type topicFilterProvider struct {
	inner         reposync.ForgeProvider
	includeTopics []string
	excludeTopics []string
}

func wrapTopicFilter(inner reposync.ForgeProvider, includeTopics, excludeTopics []string) reposync.ForgeProvider {
	include := normalizeTopicList(includeTopics)
	exclude := normalizeTopicList(excludeTopics)
	if len(include) == 0 && len(exclude) == 0 {
		return inner
	}
	return &topicFilterProvider{
		inner:         inner,
		includeTopics: include,
		excludeTopics: exclude,
	}
}

func (p *topicFilterProvider) Name() string {
	return p.inner.Name()
}

func (p *topicFilterProvider) ListOrganizationRepos(ctx context.Context, org string) ([]*provider.Repository, error) {
	repos, err := p.inner.ListOrganizationRepos(ctx, org)
	if err != nil {
		return nil, err
	}
	return filterReposByTopics(repos, p.includeTopics, p.excludeTopics), nil
}

func (p *topicFilterProvider) ListUserRepos(ctx context.Context, user string) ([]*provider.Repository, error) {
	repos, err := p.inner.ListUserRepos(ctx, user)
	if err != nil {
		return nil, err
	}
	return filterReposByTopics(repos, p.includeTopics, p.excludeTopics), nil
}

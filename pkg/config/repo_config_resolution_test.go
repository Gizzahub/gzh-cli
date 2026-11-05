// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEffectiveConfigRejectsInvalidTemplateLayers(t *testing.T) {
	// 잘못된 참조가 부분 설정의 성공으로 바뀌지 않아야 한다.
	for _, layer := range []string{"defaults", "specific", "pattern", "repository default"} {
		t.Run(layer, func(t *testing.T) {
			for _, invalid := range []string{"missing", "null", "cycle"} {
				t.Run(invalid, func(t *testing.T) {
					config := &RepoConfig{
						Templates:    map[string]*RepoTemplate{"null": nil, "cycle": {Base: "cycle"}},
						Defaults:     &RepoDefaults{Settings: &RepoSettings{Private: boolPtr(true)}},
						Repositories: &RepoTargets{},
					}
					switch layer {
					case "defaults":
						config.Defaults.Template = invalid
					case "specific":
						config.Repositories.Specific = []RepoSpecificConfig{{Name: "service-api", Template: invalid}}
					case "pattern":
						config.Repositories.Patterns = []RepoPatternConfig{{Match: "service-*", Template: invalid}}
					case "repository default":
						config.Repositories.Default = &RepoDefaultConfig{Template: invalid}
					}
					settings, security, permissions, exceptions, err := config.GetEffectiveConfig("service-api")
					require.ErrorContains(t, err, "service-api")
					require.ErrorContains(t, err, layer)
					require.ErrorContains(t, err, invalid)
					require.Nil(t, settings)
					require.Nil(t, security)
					require.Nil(t, permissions)
					require.Nil(t, exceptions)
				})
			}
		})
	}
}

func TestEffectiveConfigPreservesLayerPrecedence(t *testing.T) {
	// specific은 pattern/default를 건너뛰고, 나머지는 기존 적용 순서를 유지한다.
	config := &RepoConfig{
		Defaults: &RepoDefaults{Template: "base"},
		Templates: map[string]*RepoTemplate{
			"base":  {Settings: &RepoSettings{Private: boolPtr(true), Description: stringPtr("base")}},
			"child": {Base: "base", Settings: &RepoSettings{HasIssues: boolPtr(true)}},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{{Name: "service-api", Template: "child", Settings: &RepoSettings{Description: stringPtr("specific")}}},
			Patterns: []RepoPatternConfig{{Match: "service-*", Settings: &RepoSettings{Description: stringPtr("pattern"), HasWiki: boolPtr(true)}}},
			Default:  &RepoDefaultConfig{Settings: &RepoSettings{Description: stringPtr("repository default")}},
		},
	}
	settings, _, _, _, err := config.GetEffectiveConfig("service-api")
	require.NoError(t, err)
	require.Equal(t, "specific", *settings.Description)
	require.True(t, *settings.Private)
	require.True(t, *settings.HasIssues)
	require.Nil(t, settings.HasWiki)
	settings, _, _, _, err = config.GetEffectiveConfig("service-other")
	require.NoError(t, err)
	require.Equal(t, "repository default", *settings.Description)
	require.True(t, *settings.Private)
	require.True(t, *settings.HasWiki)
	config.Repositories.Patterns = []RepoPatternConfig{{Match: "unrelated-*", Template: "missing"}}
	_, _, _, _, err = config.GetEffectiveConfig("service-api")
	require.NoError(t, err)
}

func TestTemplateHelpersRejectNullEntries(t *testing.T) {
	config := &RepoConfig{Templates: map[string]*RepoTemplate{
		"null": nil, "child": {Base: "null"}, "valid": {},
	}}
	for _, name := range []string{"null", "child"} {
		t.Run(name, func(t *testing.T) {
			// 메모리에서 구성한 설정도 YAML loader와 같은 null 거부 규칙을 따른다.
			chain, err := config.GetTemplateInheritanceChain(name)
			require.ErrorContains(t, err, `template "null" must not be null`)
			require.Nil(t, chain)
			resolved, err := config.resolveTemplate(name)
			require.ErrorContains(t, err, `template "null" must not be null`)
			require.Nil(t, resolved)
		})
	}
	require.Equal(t, map[string][]string{"valid": {"valid"}}, config.GetAllTemplateChains())
	warnings := config.ValidateTemplateOverrides()
	require.Len(t, warnings, 2)
	for _, warning := range warnings {
		require.Contains(t, warning, `template "null" must not be null`)
	}
}

func TestMergeRepoConfigsSkipsNilInputs(t *testing.T) {
	first := &RepoConfig{
		Version: "1", Organization: "first",
		Templates: map[string]*RepoTemplate{"shared": {Description: "first"}, "kept": {}},
		Policies:  map[string]*PolicyTemplate{"first": {}},
		Defaults:  &RepoDefaults{Template: "kept"},
	}
	last := &RepoConfig{
		Version: "2", Organization: "last",
		Templates:    map[string]*RepoTemplate{"shared": {Description: "last"}},
		Policies:     map[string]*PolicyTemplate{"last": {}},
		Repositories: &RepoTargets{Default: &RepoDefaultConfig{Template: "shared"}},
	}
	merged, err := MergeRepoConfigs(nil, first, nil, last, nil)
	require.NoError(t, err)
	require.Equal(t, "1", merged.Version)
	require.Equal(t, "first", merged.Organization)
	require.Len(t, merged.Templates, 2)
	require.Equal(t, "last", merged.Templates["shared"].Description)
	require.Len(t, merged.Policies, 2)
	require.Equal(t, first.Defaults, merged.Defaults)
	require.Equal(t, last.Repositories, merged.Repositories)
	require.Equal(t, "first", first.Templates["shared"].Description)
	for _, inputs := range [][]*RepoConfig{nil, {nil}, {nil, nil}} {
		result, mergeErr := MergeRepoConfigs(inputs...)
		require.ErrorContains(t, mergeErr, "no configurations to merge")
		require.Nil(t, result)
	}
}

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRepoConfigRejectsNullTemplates(t *testing.T) {
	for _, templates := range []string{"  empty: null\n", "  child: {base: empty}\n  empty: null\n", "  empty: null\n  child: {base: empty}\n"} {
		file := filepath.Join(t.TempDir(), "config.yaml")
		require.NoError(t, os.WriteFile(file, []byte("version: '1'\norganization: org\ntemplates:\n"+templates), 0o600))
		config, err := LoadRepoConfig(file)
		require.ErrorContains(t, err, "must not be null")
		require.Nil(t, config)
	}
}

func TestTemplateInheritanceRejectsNullBaseBeforeDereference(t *testing.T) {
	templates := map[string]*RepoTemplate{"child": {Base: "empty"}, "empty": nil}
	require.ErrorContains(t, validateTemplateInheritance("child", templates["child"], templates), `template "empty" must not be null`)
}

func TestRepoConfigEnvironmentExpansionSurvivesNullTemplate(t *testing.T) {
	t.Setenv("CONFIG_CONTRACT_URL", "https://example.invalid/hook")
	t.Setenv("CONFIG_CONTRACT_SECRET", "fixture-not-a-credential")
	config := &RepoConfig{Templates: map[string]*RepoTemplate{
		"empty": nil,
		"valid": {Security: &SecuritySettings{Webhooks: []WebhookConfig{{URL: "${CONFIG_CONTRACT_URL}", Secret: "${CONFIG_CONTRACT_SECRET}"}}}},
	}}
	expandRepoConfigEnvVars(config)
	require.Equal(t, "https://example.invalid/hook", config.Templates["valid"].Security.Webhooks[0].URL)
	require.Equal(t, "fixture-not-a-credential", config.Templates["valid"].Security.Webhooks[0].Secret)
	file := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(file, []byte("version: '1'\norganization: org\ntemplates:\n  valid:\n    security:\n      webhooks:\n      - url: ${CONFIG_CONTRACT_URL}\n        secret: ${CONFIG_CONTRACT_SECRET}\n"), 0o600))
	loaded, err := LoadRepoConfig(file)
	require.NoError(t, err)
	require.Equal(t, "https://example.invalid/hook", loaded.Templates["valid"].Security.Webhooks[0].URL)
	require.Equal(t, "fixture-not-a-credential", loaded.Templates["valid"].Security.Webhooks[0].Secret)
}

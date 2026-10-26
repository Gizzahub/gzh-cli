// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Git 초기화나 원격 접속 없이 읽기 가능한 최소 로컬 저장소 메타데이터를 만든다.
func discoverContractRepo(t *testing.T, dir, org string) {
	t.Helper()
	for _, name := range []string{"objects", "refs/heads"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", name), 0o700))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o600))
	config := "[core]\n repositoryformatversion = 0\n bare = false\n[remote \"origin\"]\n url = https://github.com/" + org + "/fixture.git\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config"), []byte(config), 0o600))
}

func isolateDiscoverContract(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") {
			t.Setenv(key, "")
			require.NoError(t, os.Unsetenv(key))
		}
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_COUNT", "0")
}

func discoverContractOrganizations(t *testing.T, output string) []string {
	t.Helper()
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	var config map[string]any
	require.NoError(t, yaml.Unmarshal(data, &config))
	providers, ok := config["providers"].(map[string]any)
	require.True(t, ok)
	github, ok := providers["github"].(map[string]any)
	require.True(t, ok)
	rows, ok := github["organizations"].([]any)
	require.True(t, ok)
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		require.True(t, ok)
		name, ok := item["name"].(string)
		require.True(t, ok)
		names = append(names, name)
	}
	return names
}

func TestDiscoverContractDepthAndRecursiveFlag(t *testing.T) {
	isolateDiscoverContract(t)
	base := t.TempDir()
	discoverContractRepo(t, filepath.Join(base, "one"), "one")
	discoverContractRepo(t, filepath.Join(base, "nested", "two"), "two")
	for _, tc := range []struct {
		name      string
		recursive bool
		depth     int
		want      []string
	}{
		{name: "nonrecursive", recursive: false, depth: 3},
		{name: "depthZero", recursive: true, depth: 0},
		{name: "depthOne", recursive: true, depth: 1, want: []string{"one"}},
		{name: "depthTwo", recursive: true, depth: 2, want: []string{"one", "two"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "generated.yaml")
			cmd := newConfigGenerateDiscoverCmd()
			args := []string{"--path", base, "--output", output, "--depth", strconv.Itoa(tc.depth)}
			if !tc.recursive {
				args = append(args, "--recursive=false")
			}
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(context.Background()))
			if tc.want == nil {
				require.NoFileExists(t, output)
			} else {
				require.ElementsMatch(t, tc.want, discoverContractOrganizations(t, output))
			}
		})
	}
}

func TestDiscoverContractRepositoryRootStopsTraversal(t *testing.T) {
	isolateDiscoverContract(t)
	base := t.TempDir()
	discoverContractRepo(t, base, "root")
	discoverContractRepo(t, filepath.Join(base, "nested"), "hidden")
	for _, recursive := range []bool{false, true} {
		output := filepath.Join(t.TempDir(), "generated.yaml")
		require.NoError(t, runConfigGenerateDiscover(context.Background(), base, output, false, recursive, 3, nil, false))
		require.Equal(t, []string{"root"}, discoverContractOrganizations(t, output))
	}
}

func TestDiscoverContractIgnoreAndCancel(t *testing.T) {
	isolateDiscoverContract(t)
	base := t.TempDir()
	discoverContractRepo(t, filepath.Join(base, "keep"), "keep")
	discoverContractRepo(t, filepath.Join(base, "ignored"), "ignored")
	output := filepath.Join(t.TempDir(), "generated.yaml")
	require.NoError(t, runConfigGenerateDiscover(context.Background(), base, output, false, true, 3, []string{"ignored"}, false))
	require.Equal(t, []string{"keep"}, discoverContractOrganizations(t, output))
	before, err := os.ReadFile(output)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, runConfigGenerateDiscover(ctx, base, output, true, true, 3, nil, false), context.Canceled)
	after, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestDiscoverContractMerge(t *testing.T) {
	isolateDiscoverContract(t)
	base := t.TempDir()
	discoverContractRepo(t, base, "neworg")
	for _, tc := range []struct {
		name, existing string
		missing        bool
	}{
		{name: "existing", existing: "version: original\ncustom: keep\nproviders:\n  gitlab:\n    custom: preserved\n"},
		{name: "missing", missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "generated.yaml")
			if !tc.missing {
				require.NoError(t, os.WriteFile(output, []byte(tc.existing), 0o600))
			}
			require.NoError(t, runConfigGenerateDiscover(context.Background(), base, output, true, false, 3, nil, false))
			require.Equal(t, []string{"neworg"}, discoverContractOrganizations(t, output))
			if !tc.missing {
				data, err := os.ReadFile(output)
				require.NoError(t, err)
				var config map[string]any
				require.NoError(t, yaml.Unmarshal(data, &config))
				require.Equal(t, "keep", config["custom"])
				require.Equal(t, "original", config["version"])
				require.Contains(t, config["providers"], "gitlab")
			}
		})
	}
}

func TestDiscoverContractMergeErrorsPreserveExisting(t *testing.T) {
	isolateDiscoverContract(t)
	base := t.TempDir()
	discoverContractRepo(t, base, "org")
	for _, directory := range []bool{false, true} {
		t.Run(map[bool]string{false: "malformed", true: "readDirectory"}[directory], func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "generated.yaml")
			if directory {
				require.NoError(t, os.Mkdir(output, 0o700))
			} else {
				require.NoError(t, os.WriteFile(output, []byte("providers: [invalid\n"), 0o600))
			}
			err := runConfigGenerateDiscover(context.Background(), base, output, true, false, 3, nil, false)
			require.ErrorContains(t, err, "failed to load existing configuration")
			if directory {
				entries, readErr := os.ReadDir(output)
				require.NoError(t, readErr)
				require.Empty(t, entries)
			} else {
				data, readErr := os.ReadFile(output)
				require.NoError(t, readErr)
				require.Equal(t, "providers: [invalid\n", string(data))
			}
		})
	}
}

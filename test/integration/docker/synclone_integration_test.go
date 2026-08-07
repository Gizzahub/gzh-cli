// Package docker_test holds integration-style tests that previously pretended
// to use Docker/testcontainers. Those stubs never started containers and
// silently passed on LoadConfig errors — see issue 43.
//
// These tests now assert only what they actually exercise: writing and loading
// synclone configuration YAML. Real forge containers are not in scope until a
// non-stub testcontainers implementation lands.
package docker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/pkg/config"
	synclone "github.com/gizzahub/gzh-cli/pkg/synclone"
)

func TestSyncClone_ConfigRoundTrip_GitLabShape(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Version:         "1.0.0",
		DefaultProvider: "gitlab",
		Providers: map[string]config.Provider{
			"gitlab": {
				Token: "test-token",
				Groups: []config.GitTarget{
					{
						Name:       "test-group",
						Visibility: "public",
						Strategy:   "reset",
						CloneDir:   filepath.Join(tmpDir, "repos"),
					},
				},
			},
		},
	}
	path := writeConfig(t, tmpDir, cfg)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err, "LoadConfig must succeed for valid YAML (no network)")
	require.NotNil(t, loaded)
}

func TestSyncClone_ConfigRoundTrip_GiteaShape(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Version:         "1.0.0",
		DefaultProvider: "gitea",
		Providers: map[string]config.Provider{
			"gitea": {
				Token: "test-token",
				Orgs: []config.GitTarget{
					{
						Name:       "test-org",
						Visibility: "public",
						Strategy:   "reset",
						CloneDir:   filepath.Join(tmpDir, "repos"),
					},
				},
			},
		},
	}
	path := writeConfig(t, tmpDir, cfg)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
}

func TestSyncClone_ConfigRoundTrip_GitHubShape(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Providers: map[string]config.Provider{
			"github": {
				Token: "test-token",
				Orgs: []config.GitTarget{
					{
						Name:       "test-org",
						Visibility: "public",
						Strategy:   "reset",
						CloneDir:   filepath.Join(tmpDir, "repos"),
					},
				},
			},
		},
	}
	path := writeConfig(t, tmpDir, cfg)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
}

func writeConfig(t *testing.T, dir string, cfg *config.Config) string {
	t.Helper()
	path := filepath.Join(dir, "synclone.yaml")
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

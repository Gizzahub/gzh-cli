// Package docker_test holds integration-style tests that previously pretended
// to use Docker/testcontainers. Those stubs never started containers and
// silently passed on LoadConfig errors — see issue 43.
//
// These tests now assert only what they actually exercise: writing and loading
// legacy bulk-clone style synclone configuration YAML. Real forge containers
// are not in scope until a non-stub testcontainers implementation lands.
package docker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	synclone "github.com/gizzahub/gzh-cli/pkg/synclone"
)

func TestSyncClone_ConfigRoundTrip_GitHub(t *testing.T) {
	path := writeBulkYAML(t, `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "/tmp/repos/github"
    provider: "github"
    protocol: "https"
    org_name: "test-org"
`)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err, "LoadConfig must succeed for valid bulk-clone YAML")
	require.NotNil(t, loaded)
}

func TestSyncClone_ConfigRoundTrip_GitLab(t *testing.T) {
	path := writeBulkYAML(t, `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "/tmp/repos/gitlab"
    provider: "gitlab"
    protocol: "https"
    group_name: "test-group"
`)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
}

func TestSyncClone_ConfigRoundTrip_Gitea(t *testing.T) {
	path := writeBulkYAML(t, `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "/tmp/repos/gitea"
    provider: "gitea"
    protocol: "https"
    org_name: "test-org"
`)
	loaded, err := synclone.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
}

func TestSyncClone_ConfigRoundTrip_InvalidFails(t *testing.T) {
	path := writeBulkYAML(t, `
version: "1.0"
default:
  protocol: ftp
repo_roots: []
`)
	_, err := synclone.LoadConfig(path)
	require.Error(t, err, "invalid protocol must fail LoadConfig, not pass silently")
}

func writeBulkYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synclone.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

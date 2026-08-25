//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckGitRepoType(t *testing.T) {
	setupIsolatedGitConfig(t)
	root := t.TempDir()
	emptyRepo := filepath.Join(root, "empty")
	normalRepo := filepath.Join(root, "normal")
	nonGitDir := filepath.Join(root, "not-a-repository")

	require.NoError(t, os.MkdirAll(emptyRepo, 0o750))
	runFixtureGit(t, emptyRepo, "init", "-q")
	initFixtureRepo(t, normalRepo)
	require.NoError(t, os.MkdirAll(nonGitDir, 0o750))

	res, err := CheckGitRepoType(emptyRepo)
	require.Error(t, err)
	assert.Equal(t, RepoTypeEmpty, res)

	res, err = CheckGitRepoType(normalRepo)
	require.NoError(t, err)
	assert.Equal(t, RepoTypeNormal, res)

	res, err = CheckGitRepoType(nonGitDir)
	require.NoError(t, err)
	assert.Equal(t, RepoTypeNone, res)
}

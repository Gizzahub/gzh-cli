//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupIsolatedGitConfig prevents fixture commits from reading a developer or
// CI runner's Git configuration. The fixture identity is deliberately local
// to each test, so these tests do not depend on a globally configured user.
func setupIsolatedGitConfig(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	globalConfig := filepath.Join(home, "gitconfig")
	require.NoError(t, os.WriteFile(globalConfig, []byte("[user]\n\tname = fixture\n\temail = fixture@example.invalid\n"), 0o600))
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func runFixtureGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s failed: %s", strings.Join(args, " "), output)

	return strings.TrimSpace(string(output))
}

func initFixtureRepo(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o750))
	runFixtureGit(t, dir, "init", "-q")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("initial\n"), 0o600))
	runFixtureGit(t, dir, "add", "README.md")
	runFixtureGit(t, dir, "commit", "-qm", "initial")
}

func commitFixtureFile(t *testing.T, dir, name, contents, message string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600))
	runFixtureGit(t, dir, "add", name)
	runFixtureGit(t, dir, "commit", "-qm", message)
}

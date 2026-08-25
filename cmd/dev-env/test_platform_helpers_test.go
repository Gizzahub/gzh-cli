//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setTestHome makes os.UserHomeDir resolve to the temporary directory on every
// supported platform. Windows uses USERPROFILE rather than HOME.
func setTestHome(tb testing.TB, home string) {
	tb.Helper()
	tb.Setenv("HOME", home)
	tb.Setenv("USERPROFILE", home)
}

// assertPrivateMode verifies POSIX permission bits where the operating system
// exposes them. Windows reports synthetic permission bits, so file creation is
// the portable security contract tested there.
func assertPrivateMode(t *testing.T, info os.FileInfo, want os.FileMode) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	assert.Equal(t, want, info.Mode().Perm())
}

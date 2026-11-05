package synclone

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	synclonepkg "github.com/gizzahub/gzh-cli/pkg/synclone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStateClean(t *testing.T) (string, error) {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	cleanErr := runStateClean("", "", true)
	require.NoError(t, writer.Close())
	os.Stdout = original
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	require.NoError(t, err)
	return buf.String(), cleanErr
}

func TestStatePathSafetyCleanAllPreservesPoisonedIdentity(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := filepath.Join(home, ".gzh", "state")
	require.NoError(t, os.MkdirAll(root, 0o750))
	state := synclonepkg.NewCloneState("../../outside", "victim", t.TempDir(), "pull", 1, 0)
	data, err := json.Marshal(state)
	require.NoError(t, err)
	poison := filepath.Join(root, "poison.json")
	victim := filepath.Join(home, "outside_victim.json")
	require.NoError(t, os.WriteFile(poison, data, 0o600))
	require.NoError(t, os.WriteFile(victim, data, 0o600))
	out, cleanErr := captureStateClean(t)
	assert.Error(t, cleanErr)
	assert.NotContains(t, out, "Cleaned")
	assert.NotContains(t, out, "outside")
	for _, path := range []string{poison, victim} {
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, data, got)
	}
}

func TestStatePathSafetyCleanAllPreservesCorruptSnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	sm := synclonepkg.NewStateManager(filepath.Join(home, ".gzh", "state"))
	require.NoError(t, sm.SaveState(synclonepkg.NewCloneState("github", "org", t.TempDir(), "pull", 1, 0)))
	corrupt := sm.GetStateFilePath("github", "corrupt")
	data := []byte("{preserve corrupt recovery data")
	require.NoError(t, os.WriteFile(corrupt, data, 0o600))
	out, cleanErr := captureStateClean(t)
	assert.Error(t, cleanErr)
	assert.NotContains(t, out, "Cleaned")
	assert.NotContains(t, out, "github/org")
	assert.NotContains(t, out, "github/corrupt")
	require.FileExists(t, sm.GetStateFilePath("github", "org"))
	got, err := os.ReadFile(corrupt)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

func TestStatePathSafetyCleanAllRemovesValidatedLegacyAndEncodedEntries(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := filepath.Join(home, ".gzh", "state")
	sm := synclonepkg.NewStateManager(root)
	require.NoError(t, sm.SaveState(synclonepkg.NewCloneState("github", "org", t.TempDir(), "pull", 1, 0)))
	require.NoError(t, sm.SaveState(synclonepkg.NewCloneState("gitlab", "team/platform", t.TempDir(), "pull", 1, 0)))
	sentinel := filepath.Join(root, "notes.txt")
	require.NoError(t, os.WriteFile(sentinel, []byte("preserve notes"), 0o600))
	out, cleanErr := captureStateClean(t)
	require.NoError(t, cleanErr)
	require.Contains(t, out, "Cleaned github/org\n")
	require.Contains(t, out, "Cleaned gitlab/team/platform\n")
	require.Contains(t, out, "Cleaned 2 state(s)\n")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	data, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	require.Equal(t, "preserve notes", string(data))
}

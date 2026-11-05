package bulkclone

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writePathSafetyState(t *testing.T, path, provider, organization string) []byte {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	state := NewCloneState(provider, organization, t.TempDir(), "pull", 1, 0)
	data, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return data
}

func pathSafetySymlink(t *testing.T, oldname, newname string) {
	t.Helper()
	err := os.Symlink(oldname, newname)
	if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
		t.Skip("Windows symlink privilege is unavailable")
	}
	require.NoError(t, err)
}

func assertPathSafetyBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestStatePathSafetyTraversalPreservesOutsideSentinel(t *testing.T) {
	for _, identity := range []struct{ provider, organization, victim string }{
		{"../outside", "victim", "outside_victim.json"},
		{"github", "team/../../victim", "victim.json"},
	} {
		for _, operation := range []string{"load", "save", "delete"} {
			t.Run(identity.provider+"/"+identity.organization+"/"+operation, func(t *testing.T) {
				base := t.TempDir()
				root := filepath.Join(base, "states")
				require.NoError(t, os.Mkdir(root, 0o750))
				victim := filepath.Join(base, identity.victim)
				before := writePathSafetyState(t, victim, identity.provider, identity.organization)
				sm := NewStateManager(root)
				var err error
				switch operation {
				case "load":
					_, err = sm.LoadState(identity.provider, identity.organization)
				case "save":
					err = sm.SaveState(NewCloneState(identity.provider, identity.organization, t.TempDir(), "fetch", 2, 0))
				case "delete":
					err = sm.DeleteState(identity.provider, identity.organization)
				}
				assert.Error(t, err)
				assertPathSafetyBytes(t, victim, before)
			})
		}
	}
}

func TestStatePathSafetyRejectsUntrustedSnapshots(t *testing.T) {
	for _, kind := range []string{"corrupt", "mismatched"} {
		for _, operation := range []string{"load", "save", "delete", "list"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, "github_org.json")
				before := []byte("{corrupt snapshot preserved")
				if kind == "mismatched" {
					before = writePathSafetyState(t, path, "github", "other")
				} else {
					require.NoError(t, os.WriteFile(path, before, 0o600))
				}
				sm := NewStateManager(root)
				require.True(t, sm.HasState("github", "org"))
				var err error
				switch operation {
				case "load":
					_, err = sm.LoadState("github", "org")
				case "save":
					err = sm.SaveState(NewCloneState("github", "org", t.TempDir(), "pull", 1, 0))
				case "delete":
					err = sm.DeleteState("github", "org")
				case "list":
					_, err = sm.ListStates()
				}
				require.Error(t, err)
				assertPathSafetyBytes(t, path, before)
			})
		}
	}
}

func TestStatePathSafetyRejectsSymlinkEscapes(t *testing.T) {
	for _, kind := range []string{"file", "parent"} {
		for _, operation := range []string{"load", "save", "delete", "list"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				base := t.TempDir()
				root := filepath.Join(base, "states")
				require.NoError(t, os.Mkdir(root, 0o750))
				organization := "org"
				if kind == "parent" {
					organization = "team/platform"
				}
				outside := filepath.Join(base, "outside", "platform.json")
				before := writePathSafetyState(t, outside, "gitlab", organization)
				if kind == "file" {
					pathSafetySymlink(t, outside, filepath.Join(root, "gitlab_org.json"))
				} else {
					pathSafetySymlink(t, filepath.Dir(outside), filepath.Join(root, "gitlab_team"))
				}
				sm := NewStateManager(root)
				require.True(t, sm.HasState("gitlab", organization))
				var err error
				switch operation {
				case "load":
					_, err = sm.LoadState("gitlab", organization)
				case "save":
					err = sm.SaveState(NewCloneState("gitlab", organization, t.TempDir(), "pull", 1, 0))
				case "delete":
					err = sm.DeleteState("gitlab", organization)
				case "list":
					_, err = sm.ListStates()
				}
				require.Error(t, err)
				assertPathSafetyBytes(t, outside, before)
			})
		}
	}
}

func TestStatePathSafetyDistinctSubgroupIdentities(t *testing.T) {
	sm := NewStateManager(t.TempDir())
	organizations := []string{"team/platform", "team%2Fplatform", "team_platform"}
	paths := make(map[string]bool)
	for _, organization := range organizations {
		require.NoError(t, sm.SaveState(NewCloneState("gitlab", organization, t.TempDir(), "pull", 1, 0)))
		path := sm.GetStateFilePath("gitlab", organization)
		require.False(t, paths[path])
		paths[path] = true
		require.Equal(t, sm.stateDir, filepath.Dir(path))
		loaded, err := sm.LoadState("gitlab", organization)
		require.NoError(t, err)
		require.Equal(t, organization, loaded.Organization)
	}
	states, err := sm.ListStates()
	require.NoError(t, err)
	require.Len(t, states, len(organizations))
	require.NoError(t, sm.DeleteState("gitlab", organizations[0]))
	for _, organization := range organizations[1:] {
		require.True(t, sm.HasState("gitlab", organization))
	}
}

func TestStatePathSafetyPreservesLegacyNestedSnapshot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "gitlab_team", "platform.json")
	writePathSafetyState(t, path, "gitlab", "team/platform")
	sm := NewStateManager(root)
	require.Equal(t, path, sm.GetStateFilePath("gitlab", "team/platform"))
	state, err := sm.LoadState("gitlab", "team/platform")
	require.NoError(t, err)
	state.MarkFailed()
	require.NoError(t, sm.SaveState(state))
	require.Equal(t, path, sm.GetStateFilePath("gitlab", "team/platform"))
	states, err := sm.ListStates()
	require.NoError(t, err)
	require.Len(t, states, 1)
	require.Equal(t, "failed", states[0].Status)
	require.NoError(t, sm.DeleteState("gitlab", "team/platform"))
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Dir(path))
	require.NoError(t, err)
}

func TestStatePathSafetyPreservesDuplicateSnapshots(t *testing.T) {
	for _, operation := range []string{"load", "save", "delete", "list"} {
		t.Run(operation, func(t *testing.T) {
			root := t.TempDir()
			legacy := filepath.Join(root, "gitlab_team", "platform.json")
			encoded := NewStateManager(root).GetStateFilePath("gitlab", "team/platform")
			beforeLegacy := writePathSafetyState(t, legacy, "gitlab", "team/platform")
			beforeEncoded := writePathSafetyState(t, encoded, "gitlab", "team/platform")
			sm := NewStateManager(root)
			require.Empty(t, sm.GetStateFilePath("gitlab", "team/platform"))
			resolved, resolveErr := sm.ResolveStateFilePath("gitlab", "team/platform")
			require.Empty(t, resolved)
			require.ErrorContains(t, resolveErr, "duplicate state snapshots")
			var err error
			switch operation {
			case "load":
				_, err = sm.LoadState("gitlab", "team/platform")
			case "save":
				err = sm.SaveState(NewCloneState("gitlab", "team/platform", t.TempDir(), "pull", 1, 0))
			case "delete":
				err = sm.DeleteState("gitlab", "team/platform")
			case "list":
				_, err = sm.ListStates()
			}
			require.Error(t, err)
			assertPathSafetyBytes(t, legacy, beforeLegacy)
			assertPathSafetyBytes(t, encoded, beforeEncoded)
		})
	}
}

func TestStatePathSafetyPreservesExplicitRootSymlink(t *testing.T) {
	base := t.TempDir()
	actual := filepath.Join(base, "actual")
	require.NoError(t, os.Mkdir(actual, 0o750))
	alias := filepath.Join(base, "selected-root")
	pathSafetySymlink(t, actual, alias)
	sm := NewStateManager(alias)
	require.NoError(t, sm.SaveState(NewCloneState("github", "org", t.TempDir(), "pull", 1, 0)))
	_, err := sm.LoadState("github", "org")
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(actual, "github_org.json"))
	require.NoError(t, sm.DeleteState("github", "org"))
}

func TestStatePathSafetyRejectsInvalidIdentity(t *testing.T) {
	for _, identity := range []struct{ provider, organization string }{
		{"", "org"},
		{"github", ""},
		{".", "org"},
		{"..", "org"},
		{"C:", "org"},
		{`\\server\share`, "org"},
		{"github", `C:\temp`},
		{"github", "team:stream"},
		{"github", `\\server\share`},
		{"github", "org\x00suffix"},
		{"github", ".."},
		{"gitlab", "team//sub"},
		{"gitlab", "team/../sub"},
	} {
		t.Run(identity.provider+"/"+identity.organization, func(t *testing.T) {
			root := t.TempDir()
			sm := NewStateManager(root)
			require.Error(t, sm.SaveState(NewCloneState(identity.provider, identity.organization, t.TempDir(), "pull", 1, 0)))
			_, err := sm.LoadState(identity.provider, identity.organization)
			require.Error(t, err)
			require.Error(t, sm.DeleteState(identity.provider, identity.organization))
			require.Empty(t, sm.GetStateFilePath(identity.provider, identity.organization))
			require.True(t, sm.HasState(identity.provider, identity.organization))
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestStatePathSafetyPreservesLegacyPercentLiteral(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "gitlab_team%2Fplatform.json")
	before := writePathSafetyState(t, legacy, "gitlab", "team%2Fplatform")
	sm := NewStateManager(root)
	require.NoError(t, sm.SaveState(NewCloneState("gitlab", "team/platform", t.TempDir(), "pull", 1, 0)))
	require.Equal(t, legacy, sm.GetStateFilePath("gitlab", "team%2Fplatform"))
	require.NotEqual(t, legacy, sm.GetStateFilePath("gitlab", "team/platform"))
	assertPathSafetyBytes(t, legacy, before)
	state, err := sm.LoadState("gitlab", "team%2Fplatform")
	require.NoError(t, err)
	state.MarkFailed()
	require.NoError(t, sm.SaveState(state))
	require.Equal(t, legacy, sm.GetStateFilePath("gitlab", "team%2Fplatform"))
	states, err := sm.ListStates()
	require.NoError(t, err)
	require.Len(t, states, 2)
}

func TestStatePathSafetyPreservesLegacyFileDirectoryConflict(t *testing.T) {
	root := t.TempDir()
	leaf := filepath.Join(root, "gitlab_team", "sub.json", "leaf.json")
	before := writePathSafetyState(t, leaf, "gitlab", "team/sub.json/leaf")
	sm := NewStateManager(root)
	_, err := sm.LoadState("gitlab", "team/sub.json/leaf")
	require.NoError(t, err)
	require.Error(t, sm.SaveState(NewCloneState("gitlab", "team/sub", t.TempDir(), "pull", 1, 0)))
	assertPathSafetyBytes(t, leaf, before)
	require.DirExists(t, filepath.Dir(leaf))
}

func TestStatePathSafetyPreservesLegacyProviderDelimiter(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "my_provider_org.json")
	before := writePathSafetyState(t, legacy, "my_provider", "org")
	sm := NewStateManager(root)
	require.Equal(t, legacy, sm.GetStateFilePath("my_provider", "org"))
	state, err := sm.LoadState("my_provider", "org")
	require.NoError(t, err)
	state.MarkFailed()
	require.NoError(t, sm.SaveState(state))
	require.Equal(t, legacy, sm.GetStateFilePath("my_provider", "org"))
	updated, err := os.ReadFile(legacy)
	require.NoError(t, err)
	require.NotEqual(t, before, updated)
	_, err = sm.LoadState("my", "provider_org")
	require.ErrorContains(t, err, "conflicting state file")
	require.Error(t, sm.SaveState(NewCloneState("my", "provider_org", t.TempDir(), "pull", 1, 0)))
	require.Error(t, sm.DeleteState("my", "provider_org"))
	assertPathSafetyBytes(t, legacy, updated)
	require.NoError(t, sm.DeleteState("my_provider", "org"))
}

func TestStatePathSafetyDistinctProviderDelimiterIdentities(t *testing.T) {
	sm := NewStateManager(t.TempDir())
	for _, identity := range []struct{ provider, organization string }{
		{"my_provider", "org"}, {"my", "provider_org"}, {"My Provider", "org"}, {"my.provider", "org"},
	} {
		require.NoError(t, sm.SaveState(NewCloneState(identity.provider, identity.organization, t.TempDir(), "pull", 1, 0)))
		state, err := sm.LoadState(identity.provider, identity.organization)
		require.NoError(t, err)
		require.Equal(t, identity.provider, state.Provider)
		require.Equal(t, identity.organization, state.Organization)
	}
	states, err := sm.ListStates()
	require.NoError(t, err)
	require.Len(t, states, 4)
}

func TestStatePathSafetyRejectsInvalidUTF8BeforeCreatingRoot(t *testing.T) {
	for _, identity := range []struct{ provider, organization string }{{"bad\xff", "org"}, {"github", "bad\xff"}} {
		t.Run(identity.provider+identity.organization, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "not-created")
			sm := NewStateManager(root)
			assert.Error(t, sm.SaveState(NewCloneState(identity.provider, identity.organization, t.TempDir(), "pull", 1, 0)))
			_, err := os.Stat(root)
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestStatePathSafetySingleIdentityIgnoresUnrelatedEntries(t *testing.T) {
	for _, kind := range []string{"corrupt", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			sm := NewStateManager(root)
			state := NewCloneState("github", "org", t.TempDir(), "pull", 1, 0)
			require.NoError(t, sm.SaveState(state))
			path := sm.GetStateFilePath("github", "org")
			unrelated := filepath.Join(root, "unrelated.json")
			data := []byte("{preserve unrelated recovery data")
			if kind == "corrupt" {
				require.NoError(t, os.WriteFile(unrelated, data, 0o600))
			} else {
				unrelated = filepath.Join(root, "unrelated-directory")
				outside := filepath.Join(t.TempDir(), "outside.json")
				require.NoError(t, os.WriteFile(outside, data, 0o600))
				pathSafetySymlink(t, filepath.Dir(outside), unrelated)
			}
			require.True(t, sm.HasState("github", "org"))
			require.False(t, sm.HasState("github", "absent"))
			require.Equal(t, path, sm.GetStateFilePath("github", "org"))
			state, err := sm.LoadState("github", "org")
			require.NoError(t, err)
			state.MarkFailed()
			require.NoError(t, sm.SaveState(state))
			_, err = sm.ListStates()
			require.Error(t, err)
			require.NoError(t, sm.DeleteState("github", "org"))
			if kind == "corrupt" {
				assertPathSafetyBytes(t, unrelated, data)
			} else {
				info, err := os.Lstat(unrelated)
				require.NoError(t, err)
				require.NotZero(t, info.Mode()&os.ModeSymlink)
			}
		})
	}
}

func TestStatePathSafetyLongAndUnicodeSubgroupRoundTrip(t *testing.T) {
	for _, organization := range []string{strings.Repeat("a", 60) + "/" + strings.Repeat("b", 60), strings.Repeat("팀", 60) + "/" + strings.Repeat("조직", 60), strings.Repeat("x", 300)} {
		t.Run(organization, func(t *testing.T) {
			sm := NewStateManager(t.TempDir())
			require.NoError(t, sm.SaveState(NewCloneState("gitlab", organization, t.TempDir(), "pull", 1, 0)))
			state, err := sm.LoadState("gitlab", organization)
			require.NoError(t, err)
			require.Equal(t, organization, state.Organization)
			states, err := sm.ListStates()
			require.NoError(t, err)
			require.Len(t, states, 1)
			require.NoError(t, sm.DeleteState("gitlab", organization))
		})
	}
}

func TestStatePathSafetyPathLookupRejectsUnverifiableSnapshot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "github_org.json")
	data := []byte("{corrupt snapshot")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	sm := NewStateManager(root)
	require.Empty(t, sm.GetStateFilePath("github", "org"))
	resolved, err := sm.ResolveStateFilePath("github", "org")
	require.Empty(t, resolved)
	require.ErrorContains(t, err, "failed to unmarshal")
	assertPathSafetyBytes(t, path, data)
}

func TestStatePathSafetyPreservesLongFlatLegacy(t *testing.T) {
	root := t.TempDir()
	organization := strings.Repeat("a", 220)
	legacy := filepath.Join(root, "gitlab_"+organization+".json")
	writePathSafetyState(t, legacy, "gitlab", organization)
	sm := NewStateManager(root)
	require.Equal(t, legacy, sm.GetStateFilePath("gitlab", organization))
	state, err := sm.LoadState("gitlab", organization)
	require.NoError(t, err)
	state.MarkFailed()
	require.NoError(t, sm.SaveState(state))
	require.Equal(t, legacy, sm.GetStateFilePath("gitlab", organization))
	require.NoError(t, sm.DeleteState("gitlab", organization))
}

func TestStatePathSafetyDigestCollisionPreservesSnapshot(t *testing.T) {
	root := t.TempDir()
	sm := NewStateManager(root)
	path := sm.GetStateFilePath("gitlab", "team/platform")
	before := writePathSafetyState(t, path, "gitlab", "different/group")
	_, err := sm.LoadState("gitlab", "team/platform")
	require.Error(t, err)
	require.Error(t, sm.SaveState(NewCloneState("gitlab", "team/platform", t.TempDir(), "pull", 1, 0)))
	require.Error(t, sm.DeleteState("gitlab", "team/platform"))
	require.Empty(t, sm.GetStateFilePath("gitlab", "team/platform"))
	assertPathSafetyBytes(t, path, before)
}

package devenv

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
)

func isolateSSHOwner(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("PATH", t.TempDir())
	return home
}

func TestSSHOwnerStateSwitchRollback(t *testing.T) {
	home := isolateSSHOwner(t)
	var adapter SSHSwitcher
	require.Equal(t, "ssh", adapter.Name())
	ctx := context.Background()
	absent, err := adapter.GetCurrentState(ctx)
	require.NoError(t, err)
	require.Equal(t, &environment.SSHConfig{}, absent)
	first := filepath.Join(home, "first")
	second := filepath.Join(home, "second")
	require.NoError(t, os.WriteFile(first, []byte("Host first\n"), 0o600))
	require.NoError(t, os.WriteFile(second, []byte("Host second\n"), 0o600))
	require.NoError(t, adapter.Switch(ctx, &environment.SSHConfig{Config: first}))
	active := filepath.Join(home, ".ssh", "config")
	target, err := os.Readlink(active)
	require.NoError(t, err)
	require.Equal(t, first, target)
	previous, err := adapter.GetCurrentState(ctx)
	require.NoError(t, err)
	require.Equal(t, &environment.SSHConfig{Config: first}, previous)
	require.NoError(t, adapter.Switch(ctx, &environment.SSHConfig{Config: second}))
	require.NoError(t, adapter.Rollback(ctx, previous))
	target, err = os.Readlink(active)
	require.NoError(t, err)
	require.Equal(t, first, target)
	require.NoError(t, adapter.Rollback(ctx, absent))
	_, err = os.Lstat(active)
	require.True(t, os.IsNotExist(err))
	data, err := os.ReadFile(first)
	require.NoError(t, err)
	require.Equal(t, "Host first\n", string(data))
}

func TestSSHOwnerRefusesOrdinaryFile(t *testing.T) {
	home := isolateSSHOwner(t)
	active := filepath.Join(home, ".ssh", "config")
	require.NoError(t, os.MkdirAll(filepath.Dir(active), 0o700))
	require.NoError(t, os.WriteFile(active, []byte("user-owned"), 0o600))
	target := filepath.Join(home, "other")
	require.NoError(t, os.WriteFile(target, []byte("other"), 0o600))
	adapter := &SSHSwitcher{}
	require.ErrorContains(t, adapter.Switch(context.Background(), &environment.SSHConfig{Config: target}), "refusing to overwrite")
	require.ErrorContains(t, adapter.Rollback(context.Background(), &environment.SSHConfig{}), "refusing to remove")
	require.NoError(t, adapter.Switch(context.Background(), &environment.SSHConfig{Config: active}))
	data, err := os.ReadFile(active)
	require.NoError(t, err)
	require.Equal(t, "user-owned", string(data))
}

func TestSSHOwnerInputAndCancellation(t *testing.T) {
	for _, operation := range []string{"switch", "rollback"} {
		t.Run(operation, func(t *testing.T) {
			home := isolateSSHOwner(t)
			adapter := &SSHSwitcher{}
			call := adapter.Switch
			if operation == "rollback" {
				call = adapter.Rollback
			}
			for _, input := range []any{nil, (*environment.SSHConfig)(nil), "wrong"} {
				require.EqualError(t, call(context.Background(), input), "invalid SSH configuration type")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			require.ErrorIs(t, call(ctx, &environment.SSHConfig{Config: "missing"}), context.Canceled)
			state, err := adapter.GetCurrentState(ctx)
			require.Nil(t, state)
			require.ErrorIs(t, err, context.Canceled)
			_, err = os.Lstat(filepath.Join(home, ".ssh"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

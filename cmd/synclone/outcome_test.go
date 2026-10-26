// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/internal/app"
	pkgconfig "github.com/gizzahub/gzh-cli/pkg/config"
)

// 실제 설정 로딩부터 대상 실행까지 통과하되 원격 provider 호출만 호출별로 대체한다.
func outcomeConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	roots := []map[string]any{}
	for _, name := range []string{"first", "second", "third"} {
		roots = append(roots, map[string]any{"root_path": filepath.Join(dir, name), "provider": "github", "protocol": "https", "org_name": name})
	}
	data, err := yaml.Marshal(map[string]any{"version": "1.0", "default": map[string]any{"protocol": "https"}, "repo_roots": roots})
	require.NoError(t, err)
	path := filepath.Join(dir, "fixture.yaml")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func TestSyncloneOutcomeAggregatesAndContinues(t *testing.T) {
	for _, failures := range []int{0, 1, 2} {
		t.Run(string(rune('0'+failures)), func(t *testing.T) {
			path := outcomeConfig(t)
			firstErr := errors.New("first fixture failure")
			secondErr := errors.New("second fixture failure")
			seen := []string{}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			runner := func(got context.Context, target pkgconfig.BulkCloneTarget, destination string) error {
				require.Equal(t, ctx, got)
				require.Equal(t, target.CloneDir, destination)
				require.Equal(t, "github", target.Provider)
				seen = append(seen, target.Name)
				if len(seen) == 1 && failures > 0 {
					return firstErr
				}
				if len(seen) == 2 && failures > 1 {
					return secondErr
				}
				return nil
			}
			command := newSyncCloneCmdWithRunner(app.NewTestAppContext(), runner)
			command.SetContext(ctx)
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)
			command.SilenceErrors = true
			command.SetArgs([]string{"--config", path, "--strategy", "fetch"})
			err := command.Execute()
			require.Len(t, seen, 3, "later targets must still run")
			if failures == 0 {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, firstErr)
				require.ErrorContains(t, err, "github/"+seen[0])
			}
			if failures == 2 {
				require.ErrorIs(t, err, secondErr)
				require.ErrorContains(t, err, "github/"+seen[1])
			}
		})
	}
}

func TestSyncloneOutcomeCancellationAndNoTargets(t *testing.T) {
	for _, precanceled := range []bool{false, true} {
		path := outcomeConfig(t)
		ctx, cancel := context.WithCancel(context.Background())
		if precanceled {
			cancel()
		}
		calls := 0
		failure := errors.New("runner failed before cancellation")
		runner := func(context.Context, pkgconfig.BulkCloneTarget, string) error { calls++; cancel(); return failure }
		opts := defaultSyncCloneOptions()
		opts.configFile = path
		err := opts.runWithCentralConfigServiceUsing(ctx, runner)
		cancel()
		require.ErrorIs(t, err, context.Canceled)
		if precanceled {
			require.Zero(t, calls)
		} else {
			require.Equal(t, 1, calls)
			require.ErrorIs(t, err, failure)
		}
	}
	path := outcomeConfig(t)
	calls := 0
	opts := defaultSyncCloneOptions()
	opts.configFile = path
	opts.providerFilter = "gitlab"
	require.NoError(t, opts.runWithCentralConfigServiceUsing(context.Background(), func(context.Context, pkgconfig.BulkCloneTarget, string) error { calls++; return nil }))
	require.Zero(t, calls)
}

func TestLegacyGiteaNeverReportsNoOpSuccess(t *testing.T) {
	for _, strategy := range []string{"reset", "pull", "fetch"} {
		dir := t.TempDir()
		target := filepath.Join(dir, "not-created")
		command := newSyncCloneGiteaCmd(app.NewTestAppContext())
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SilenceErrors, command.SilenceUsage = true, true
		require.Contains(t, command.Deprecated, "gz synclone forge --provider gitea")
		command.SetArgs([]string{"--targetPath", target, "--orgName", "fixture", "--strategy", strategy})
		require.ErrorContains(t, command.Execute(), "legacy Gitea cloning is not supported")
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Empty(t, entries)
	}
	for _, args := range [][]string{{}, {"--targetPath", "unused", "--orgName", "fixture", "--strategy", "bogus"}} {
		command := newSyncCloneGiteaCmd(app.NewTestAppContext())
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SilenceErrors, command.SilenceUsage = true, true
		command.SetArgs(args)
		err := command.Execute()
		require.Error(t, err)
		require.NotContains(t, err.Error(), "not supported")
	}
	command := NewSyncCloneCmd(app.NewTestAppContext())
	legacy, _, err := command.Find([]string{"gitea"})
	require.NoError(t, err)
	for _, name := range []string{"targetPath", "orgName", "strategy"} {
		require.NotNil(t, legacy.Flags().Lookup(name))
	}
	command.SetOut(io.Discard)
	command.SetArgs([]string{"gitea", "--help"})
	require.NoError(t, command.Execute())
	forge, _, err := command.Find([]string{"forge"})
	require.NoError(t, err)
	require.Equal(t, "forge", forge.Name())
}

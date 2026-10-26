// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	"github.com/gizzahub/gzh-cli/internal/git/clone"
	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// 재귀 회귀가 생겨도 테스트 프로세스 전체가 스택 고갈로 종료되지 않도록 격리한다.
func TestCloneResumeDispatch(t *testing.T) {
	if os.Getenv("GZH_CLONE_RESUME_TEST_CHILD") != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCloneResumeDispatch$", "-test.count=1")
		command.Env = append(os.Environ(), "GZH_CLONE_RESUME_TEST_CHILD=1")
		output, err := command.CombinedOutput()
		require.NoError(t, ctx.Err(), "resume child timed out: %s", output)
		require.NoError(t, err, "resume child failed: %s", output)
		return
	}
	isolateCloneSession(t)
	opts := clone.DefaultCloneOptions()
	opts.Provider, opts.Org, opts.Target = "github", "testorg", filepath.Join(t.TempDir(), "target")
	opts.Parallel, opts.MaxRetries = 7, 8
	opts.Format = string(clone.FormatJSON)
	session := clone.NewSession(opts)
	session.AddRepository("testorg/completed")
	session.MarkCompleted("testorg/completed")
	session.AddRepository("testorg/pending")
	require.NoError(t, session.Initialize())

	fake := providermock.NewProvider("github")
	fake.SetupListResponse("testorg", []provider.Repository{
		{Name: "completed", FullName: "testorg/completed", CloneURL: "https://invalid.example/completed.git"},
		{Name: "pending", FullName: "testorg/pending", CloneURL: "https://invalid.example/pending.git"},
	})
	calls := 0
	realProvider, realRunner := getCloneProvider, cloneGitRunner
	t.Cleanup(func() { getCloneProvider, cloneGitRunner = realProvider, realRunner })
	getCloneProvider = func(got *clone.CloneOptions) (provider.GitProvider, error) {
		calls++
		require.Equal(t, session.ID, got.Resume)
		require.Equal(t, opts.Target, got.Target)
		require.Equal(t, "github", got.Provider)
		require.Equal(t, "testorg", got.Org)
		require.Equal(t, 2, got.Parallel)
		require.Equal(t, 1, got.MaxRetries)
		require.Equal(t, string(clone.FormatTable), got.Format)
		require.True(t, got.Quiet)
		return fake, nil
	}
	runner := &gitinternal.RecordingGitRunner{}
	cloneGitRunner = runner
	cmd := newRepoCloneCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"--resume", session.ID, "--parallel", "2", "--max-retries", "1", "--format", "table", "--quiet"})
	require.NoError(t, cmd.Execute())
	require.Equal(t, 1, calls)
	clones := runner.CloneArgs()
	require.Len(t, clones, 1)
	require.Contains(t, cloneArgsJoined(clones), "pending.git")
	require.NotContains(t, cloneArgsJoined(clones), "completed.git")
	restored := clone.NewSession(nil)
	require.NoError(t, restored.Load(session.ID))
	require.Equal(t, session.ID, restored.ID)
	require.True(t, restored.IsCompleted("testorg/completed"))
	require.True(t, restored.IsCompleted("testorg/pending"))
	ids, err := clone.ListSessions()
	require.NoError(t, err)
	require.Equal(t, []string{session.ID}, ids)
	fake.AssertExpectations(t)
}

// 사용자 설정·작업 디렉터리는 읽거나 변경하지 않는다.
func isolateCloneSession(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("APPDATA", root)
}

func TestCloneResumeInvalidSessions(t *testing.T) {
	for _, tt := range []struct{ name, content, want string }{
		{"missing", "", "failed to load session"},
		{"corrupt", "not-json", "failed to unmarshal session data"},
		{"null options", `{"id":"invalid","options":null}`, "saved options are missing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			isolateCloneSession(t)
			opts := clone.DefaultCloneOptions()
			session := clone.NewSession(opts)
			require.NoError(t, session.Initialize())
			if tt.content != "" {
				config, err := os.UserConfigDir()
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Join(config, "gzh-manager", "sessions", "invalid.json"), []byte(tt.content), 0o600))
			}
			real := getCloneProvider
			t.Cleanup(func() { getCloneProvider = real })
			getCloneProvider = func(*clone.CloneOptions) (provider.GitProvider, error) {
				t.Fatal("invalid session must fail before resolving a provider")
				return nil, nil
			}
			cmd := newRepoCloneCmd()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs([]string{"--resume", "invalid"})
			require.ErrorContains(t, cmd.Execute(), tt.want)
		})
	}
}

func TestCloneRequiredAndConflictingFlags(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--provider", "github"},
		{"--org", "testorg"},
		{"--resume", "session", "--provider", "github"},
		{"--resume", "session", "--org", "testorg"},
		{"--resume", "session", "--quiet", "--verbose"},
	} {
		t.Run(jsonName(args), func(t *testing.T) {
			real := getCloneProvider
			t.Cleanup(func() { getCloneProvider = real })
			getCloneProvider = func(*clone.CloneOptions) (provider.GitProvider, error) {
				t.Fatal("invalid flags must fail before resolving a provider")
				return nil, nil
			}
			cmd := newRepoCloneCmd()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(args)
			err := cmd.Execute()
			require.Error(t, err)
			require.Contains(t, err.Error(), "flags")
			require.NotContains(t, err.Error(), "failed to load session")
		})
	}
}

func jsonName(args []string) string {
	name, _ := json.Marshal(args)
	return string(name)
}

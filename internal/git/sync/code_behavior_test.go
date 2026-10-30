// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

type scenarioGitRunner func(context.Context, string, ...string) ([]byte, error)

func (f scenarioGitRunner) Run(ctx context.Context, directory string, arguments ...string) ([]byte, error) {
	return f(ctx, directory, arguments...)
}

func TestCodeSyncUsesOwnedCloneAndExplicitDestination(t *testing.T) {
	_, temporary := isolatedTrackerHome(t)
	runner := &gitinternal.RecordingGitRunner{}
	syncer := &CodeSyncer{source: provider.Repository{CloneURL: "https://source.invalid/team/repo.git"}, destination: &provider.Repository{CloneURL: "https://target.invalid/team/repo.git"}, runner: runner, options: SyncOptions{Verbose: true}}
	require.NoError(t, syncer.ValidateRepositories(t.Context()))
	require.NoError(t, syncer.Sync(t.Context()))
	require.Len(t, runner.Calls, 4)
	clone := runner.Calls[0]
	require.Empty(t, clone.Dir)
	require.Equal(t, []string{"clone", "--mirror", syncer.source.CloneURL}, clone.Args[:3])
	repository := clone.Args[3]
	require.Equal(t, temporary, filepath.Dir(filepath.Dir(repository)))
	require.Equal(t, gitinternal.GitCall{Dir: repository, Args: []string{"remote", "add", "destination", syncer.destination.CloneURL}}, runner.Calls[1])
	require.Equal(t, gitinternal.GitCall{Dir: repository, Args: []string{"push", "destination", "--all"}}, runner.Calls[2])
	require.Equal(t, gitinternal.GitCall{Dir: repository, Args: []string{"push", "destination", "--tags"}}, runner.Calls[3])
	_, err := os.Stat(filepath.Dir(repository))
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestCodeSyncRejectsMissingInputsAndPreservesSetupErrors(t *testing.T) {
	isolatedTrackerHome(t)
	runner := &gitinternal.RecordingGitRunner{}
	syncer := &CodeSyncer{runner: runner}
	require.ErrorContains(t, syncer.ValidateRepositories(t.Context()), "source repository clone URL")
	require.ErrorContains(t, syncer.Sync(t.Context()), "destination repository is required")
	require.Empty(t, runner.Calls)
	syncer.source.CloneURL = "https://source.invalid/repo.git"
	syncer.destination = &provider.Repository{}
	require.ErrorContains(t, syncer.ValidateRepositories(t.Context()), "destination repository clone URL")
	syncer.destination.CloneURL = "https://target.invalid/repo.git"
	for _, stage := range []string{"clone", "remote"} {
		sentinel := errors.New("fixture " + stage + " failure")
		calls := 0
		syncer.runner = scenarioGitRunner(func(_ context.Context, _ string, args ...string) ([]byte, error) {
			calls++
			if args[0] == stage {
				return []byte("owned diagnostic"), sentinel
			}
			return nil, nil
		})
		err := syncer.Sync(t.Context())
		require.ErrorIs(t, err, sentinel)
		require.ErrorContains(t, err, "owned diagnostic")
		if stage == "clone" {
			require.Equal(t, 1, calls)
		} else {
			require.Equal(t, 2, calls)
		}
	}
}

func TestCodeSyncPreservesExplicitForceOption(t *testing.T) {
	isolatedTrackerHome(t)
	runner := &gitinternal.RecordingGitRunner{}
	syncer := &CodeSyncer{source: provider.Repository{CloneURL: "https://source.invalid/repo.git"}, destination: &provider.Repository{CloneURL: "https://target.invalid/repo.git"}, runner: runner, options: SyncOptions{Force: true}}
	require.NoError(t, syncer.Sync(t.Context()))
	require.Len(t, runner.Calls, 4)
	require.Equal(t, []string{"push", "destination", "--all", "--force"}, runner.Calls[2].Args)
	require.Equal(t, []string{"push", "destination", "--tags", "--force"}, runner.Calls[3].Args)
}

// 일반 실패를 강제 덮어쓰기로 확대하지 않고 태그 실패도 원래 원인을 보존한다.
func TestCodeSyncNeverEscalatesRejectedPushAndPreservesTagFailure(t *testing.T) {
	for _, stage := range []string{"branches", "tags"} {
		t.Run(stage, func(t *testing.T) {
			isolatedTrackerHome(t)
			sentinel := errors.New("fixture rejected " + stage)
			var forced bool
			runner := scenarioGitRunner(func(_ context.Context, _ string, args ...string) ([]byte, error) {
				if slices.Contains(args, "--force") {
					forced = true
					return nil, nil
				}
				if args[0] == "push" && ((stage == "branches" && slices.Contains(args, "--all")) || (stage == "tags" && slices.Contains(args, "--tags"))) {
					return nil, sentinel
				}
				return nil, nil
			})
			syncer := &CodeSyncer{source: provider.Repository{CloneURL: "https://source.invalid/repo.git"}, destination: &provider.Repository{CloneURL: "https://target.invalid/repo.git"}, runner: runner}
			err := syncer.Sync(t.Context())
			if forced {
				t.Error("default sync retried a rejected push with --force")
			}
			require.ErrorIs(t, err, sentinel)
		})
	}
}

func TestCodeSyncCreationAndStatisticsUseInjectedRunner(t *testing.T) {
	isolatedTrackerHome(t)
	source := provider.Repository{Name: "repo", Description: "metadata", Private: true, CloneURL: "https://source.invalid/repo.git", Topics: []string{"go"}}
	destination := &provider.Repository{Name: "repo", FullName: "target/repo", CloneURL: "https://target.invalid/repo.git"}
	dst := providermock.NewProvider("gitlab")
	dst.On("CreateRepository", mock.Anything, mock.MatchedBy(func(request provider.CreateRepoRequest) bool {
		return request.Owner == "target" && request.Name == "repo" && request.Description == "metadata" && request.Private && request.HasIssues && request.HasWiki && len(request.Topics) == 1 && request.Topics[0] == "go"
	})).Return(destination, nil).Once()
	runner := &gitinternal.RecordingGitRunner{}
	syncer := &CodeSyncer{source: source, runner: runner, options: SyncOptions{Verbose: true}}
	_, err := syncer.SyncNewRepository(t.Context(), dst, nil)
	require.ErrorContains(t, err, "destination target is required")
	_, err = syncer.SyncNewRepository(t.Context(), dst, &SyncTarget{})
	require.ErrorContains(t, err, "destination organization")
	created, err := syncer.SyncNewRepository(t.Context(), dst, &SyncTarget{Org: "target"})
	require.NoError(t, err)
	require.Same(t, destination, created)
	dst.AssertExpectations(t)

	directory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(directory, "repo"), 0o700))
	syncer.runner = scenarioGitRunner(func(_ context.Context, path string, args ...string) ([]byte, error) {
		require.Equal(t, filepath.Join(directory, "repo"), path)
		if args[0] == "branch" {
			require.Equal(t, []string{"branch", "-r", "--list"}, args)
			return []byte("  origin/main\n\n  origin/topic\n"), nil
		}
		require.Equal(t, []string{"tag", "--list"}, args)
		return []byte("v1\n"), nil
	})
	statistics, err := syncer.GetSyncStats(t.Context(), directory)
	require.NoError(t, err)
	require.Equal(t, 2, statistics.BranchCount)
	require.Equal(t, 1, statistics.TagCount)
	require.Equal(t, "unknown", statistics.Size)
}

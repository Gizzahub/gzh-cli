// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

func TestSyncPullRequestsRejectedBeforeProviderConstruction(t *testing.T) {
	realProvider, realRunner := getGitProvider, syncGitRunner
	t.Cleanup(func() { getGitProvider, syncGitRunner = realProvider, realRunner })
	calls := 0
	sentinel := errors.New("fixture provider construction")
	getGitProvider = func(string, string) (provider.GitProvider, error) { calls++; return nil, sentinel }
	runner := &gitinternal.RecordingGitRunner{}
	syncGitRunner = runner
	for _, extra := range [][]string{
		{"--include-prs", "--include-code=false"},
		{"--include-prs"},
		{"--include-prs", "--include-code=false", "--dry-run"},
		{"--include-prs", "--dry-run"},
	} {
		command := newRepoSyncCmd()
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SilenceUsage, command.SilenceErrors = true, true
		command.SetArgs(append([]string{"--from", "github:source", "--to", "gitlab:destination"}, extra...))
		require.ErrorContains(t, command.Execute(), "--include-prs")
		require.Zero(t, calls)
		require.Empty(t, runner.Calls)
	}
	command := newRepoSyncCmd()
	require.Equal(t, "false", command.Flags().Lookup("include-prs").DefValue)
	require.Equal(t, "true", command.Flags().Lookup("include-code").DefValue)
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SilenceUsage, command.SilenceErrors = true, true
	command.SetArgs([]string{"--from", "github:source", "--to", "gitlab:destination"})
	require.ErrorIs(t, command.Execute(), sentinel, "default false must retain supported code dispatch")
	require.Equal(t, 1, calls)
}

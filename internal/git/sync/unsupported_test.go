// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// PR 단독·코드 혼합·dry-run 모두 검증과 직접 engine 경로에서 작업 전에 거부한다.
func TestUnsupportedPullRequestsBeforeOperations(t *testing.T) {
	for _, code := range []bool{false, true} {
		for _, dryRun := range []bool{false, true} {
			opts := SyncOptions{From: "github:source", To: "gitlab:destination", IncludePRs: true, IncludeCode: code, DryRun: dryRun}
			require.ErrorContains(t, opts.Validate(), "--include-prs")
			require.Zero(t, opts.Parallel, "unsupported validation must not normalize other fields")
			source, destination := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
			runner := &gitinternal.RecordingGitRunner{}
			engine := NewSyncEngine(source, destination, opts)
			engine.SetRunner(runner)
			require.ErrorContains(t, engine.Sync(context.Background()), "not supported")
			require.Empty(t, source.Calls, "no provider query before rejection")
			require.Empty(t, destination.Calls)
			require.Empty(t, runner.Calls)
		}
	}
}

// 직접 engine 호출은 전체 Validate를 새로 적용하거나 기존 기본값을 변경하지 않는다.
func TestSupportedCodeRetainsDirectEngineBehavior(t *testing.T) {
	opts := SyncOptions{From: "github:source", To: "gitlab:destination", IncludeCode: true}
	source, destination := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
	source.SetupListResponse("source", nil)
	engine := NewSyncEngine(source, destination, opts)
	require.NoError(t, engine.Sync(context.Background()))
	require.Len(t, source.Calls, 1)
	require.Empty(t, destination.Calls)
	require.Zero(t, engine.options.Parallel)
	require.Len(t, engine.createSyncActions(provider.Repository{}, nil), 1)
	source.AssertExpectations(t)
}

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

func TestEngineCreationRebindsActionsToActualDestination(t *testing.T) {
	source := provider.Repository{Name: "repo", FullName: "source/repo", Description: "source metadata", Private: true, Topics: []string{"go"}}
	destination := provider.Repository{Name: "renamed", FullName: "target/renamed", Description: "old metadata"}
	src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
	src.SetupGetResponse("source/repo", &source, nil)
	dst.SetupGetResponse("target/renamed", nil, errors.New("repository not found"))
	dst.On("CreateRepository", mock.Anything, mock.MatchedBy(func(request provider.CreateRepoRequest) bool {
		return request.Owner == "target" && request.Name == "renamed" && request.Private && request.HasIssues && request.HasWiki && request.Description == "source metadata" && len(request.Topics) == 1 && request.Topics[0] == "go"
	})).Return(&destination, nil).Once()
	dst.On("UpdateRepository", mock.Anything, "target/renamed", mock.MatchedBy(func(request provider.UpdateRepoRequest) bool {
		return request.Description != nil && *request.Description == "source metadata"
	})).Return(&destination, nil).Once()
	engine := NewSyncEngine(src, dst, SyncOptions{From: "github:source/repo", To: "gitlab:target/renamed", CreateMissing: true, IncludeSettings: true})
	require.NoError(t, engine.Sync(t.Context()))
	src.AssertExpectations(t)
	dst.AssertExpectations(t)
}

func TestEngineDryRunOnlyReadsFilteredRepositoryLists(t *testing.T) {
	src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
	src.SetupListResponse("source", []provider.Repository{
		{Name: "app-new", FullName: "source/app-new"},
		{Name: "app-existing", FullName: "source/app-existing"},
		{Name: "app-ignore", FullName: "source/app-ignore"},
		{Name: "other", FullName: "source/other"},
	})
	dst.SetupListResponse("target", []provider.Repository{{Name: "app-existing", FullName: "target/app-existing"}})
	for _, enabled := range []bool{false, true} {
		engine := NewSyncEngine(src, dst, SyncOptions{From: "github:source", To: "gitlab:target", IncludeSettings: true, Match: "app-*", Exclude: "*-ignore", CreateMissing: enabled, UpdateExisting: enabled, DryRun: true})
		require.NoError(t, engine.Sync(t.Context()))
	}
	// 변경 API는 기대에 등록하지 않아 잘못 호출하면 mock이 실패한다.
	for _, call := range append(src.Calls, dst.Calls...) {
		require.Equal(t, "ListRepositories", call.Method)
	}
	src.AssertExpectations(t)
	dst.AssertExpectations(t)
}

func TestEnginePropagatesSourceLookupFailure(t *testing.T) {
	src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
	sentinel := errors.New("source fixture denied read")
	src.SetupGetResponse("source/repo", nil, sentinel)
	engine := NewSyncEngine(src, dst, SyncOptions{From: "github:source/repo", To: "gitlab:target/repo", IncludeSettings: true})
	require.ErrorIs(t, engine.Sync(t.Context()), sentinel)
	require.Empty(t, dst.Calls)
	src.AssertExpectations(t)
}

// 실제 API나 Git 없이 atomic 관측으로 동시 실행의 완료와 오류 집계를 검증한다.
func TestEngineExecutesSequentialAndParallelPlans(t *testing.T) {
	for _, workers := range []int{1, 3} {
		for _, fail := range []bool{false, true} {
			var completed atomic.Int32
			destination := &provider.Repository{FullName: "target/repo"}
			first := RepoSync{Source: provider.Repository{FullName: "source/first"}, Destination: destination, Actions: []SyncAction{{Type: "fixture", Handler: func(context.Context) error {
				completed.Add(1)
				return nil
			}}}}
			second := RepoSync{Source: provider.Repository{FullName: "source/second"}, Destination: destination, Actions: []SyncAction{{Type: "fixture", Handler: func(context.Context) error {
				completed.Add(1)
				if fail {
					return errors.New("fixture failure")
				}
				return nil
			}}}}
			engine := NewSyncEngine(nil, nil, SyncOptions{Parallel: workers})
			err := engine.executeSyncPlan(t.Context(), SyncPlan{Create: []RepoSync{first}, Update: []RepoSync{second}})
			if fail {
				require.ErrorContains(t, err, "1 repositories")
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, int32(2), completed.Load())
			require.NoError(t, engine.executeSyncPlan(t.Context(), SyncPlan{}))
		}
	}
}

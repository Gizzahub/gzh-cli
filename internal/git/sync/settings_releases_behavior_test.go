// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

func TestSettingsSynchronizationPreservesProtectedFieldsUnlessForced(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "force"}[force], func(t *testing.T) {
			source := provider.Repository{FullName: "source/repo", Description: "new description", DefaultBranch: "main", Topics: []string{"go", "cli"}, Visibility: provider.VisibilityPrivate, Archived: true}
			destination := provider.Repository{FullName: "target/repo", Description: "old description", DefaultBranch: "legacy", Topics: []string{"other"}, Visibility: provider.VisibilityPublic}
			dst := providermock.NewProvider("gitlab")
			dst.On("UpdateRepository", mock.Anything, "target/repo", mock.MatchedBy(func(update provider.UpdateRepoRequest) bool {
				if update.Description == nil || *update.Description != source.Description || update.Archived == nil || !*update.Archived {
					return false
				}
				if len(update.Topics) != 2 || update.Topics[0] != "go" || update.Topics[1] != "cli" || update.Private != nil {
					return false
				}
				if force {
					return update.DefaultBranch != nil && *update.DefaultBranch == "main" && update.Visibility == provider.VisibilityPrivate
				}
				return update.DefaultBranch == nil && update.Visibility == ""
			})).Return(&destination, nil).Once()
			syncer := NewSettingsSyncer(nil, dst, SyncOptions{Force: force, Verbose: true})
			require.NoError(t, syncer.Sync(t.Context(), source, destination))
			dst.AssertExpectations(t)
		})
	}
}

func TestSettingsSynchronizationNoopAndProviderFailure(t *testing.T) {
	source := provider.Repository{FullName: "source/repo", Topics: []string{"cli", "go"}}
	destination := provider.Repository{FullName: "target/repo", Topics: []string{"go", "cli"}}
	dst := providermock.NewProvider("gitlab")
	require.NoError(t, NewSettingsSyncer(nil, dst, SyncOptions{Verbose: true}).Sync(t.Context(), source, destination))
	require.Empty(t, dst.Calls)
	source.Description = "changed"
	sentinel := errors.New("fixture denied mutation")
	dst.On("UpdateRepository", mock.Anything, "target/repo", mock.Anything).Return(nil, sentinel).Once()
	require.ErrorIs(t, NewSettingsSyncer(nil, dst, SyncOptions{}).Sync(t.Context(), source, destination), sentinel)
	dst.AssertExpectations(t)
}

// 실제 게시 없이 provider 호출과 업로드 바이트로 페이지·중복·필드 전달을 관측한다.
func TestReleaseSynchronizationPaginatesCreatesUpdatesAndSkips(t *testing.T) {
	src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
	newRelease := provider.Release{ID: "source-new", TagName: "v2", Name: "new", Body: "notes", TargetBranch: "main", Draft: true, Prerelease: true, Assets: []provider.Asset{
		{ID: "already", Name: "existing.txt"},
		{ID: "fresh", Name: "new.bin", Label: "binary", ContentType: "application/octet-stream"},
	}}
	changed := provider.Release{ID: "source-changed", TagName: "v1", Name: "renamed", Body: "current notes"}
	unchanged := provider.Release{ID: "source-same", TagName: "v0", Name: "same", Body: "same notes"}
	src.On("ListReleases", mock.Anything, "source/repo", provider.ListReleasesOptions{IncludeDrafts: true, IncludePrereleases: true, Page: 1, PerPage: 100}).Return(&provider.ReleaseList{Releases: []provider.Release{newRelease}, HasNext: true}, nil).Once()
	src.On("ListReleases", mock.Anything, "source/repo", provider.ListReleasesOptions{IncludeDrafts: true, IncludePrereleases: true, Page: 2, PerPage: 100}).Return(&provider.ReleaseList{Releases: []provider.Release{changed, unchanged}}, nil).Once()
	dst.On("ListReleases", mock.Anything, "target/repo", mock.Anything).Return(&provider.ReleaseList{Releases: []provider.Release{
		{ID: "target-changed", TagName: "v1", Name: "old"},
		{ID: "target-same", TagName: "v0", Name: "same", Body: "same notes"},
	}}, nil).Once()
	dst.On("CreateRelease", mock.Anything, "target/repo", provider.CreateReleaseRequest{TagName: "v2", Name: "new", Body: "notes", TargetBranch: "main", Draft: true, Prerelease: true}).Return(&provider.Release{ID: "target-new"}, nil).Once()
	dst.On("UpdateRelease", mock.Anything, "target/repo", "target-changed", mock.MatchedBy(func(update provider.UpdateReleaseRequest) bool {
		return update.Name != nil && *update.Name == "renamed" && update.Body != nil && *update.Body == "current notes" && update.Draft != nil && !*update.Draft && update.Prerelease != nil && !*update.Prerelease
	})).Return(&provider.Release{ID: "target-changed"}, nil).Once()
	dst.On("ListReleaseAssets", mock.Anything, "target/repo", "target-new").Return([]provider.Asset{{Name: "existing.txt"}}, nil).Once()
	src.On("DownloadReleaseAsset", mock.Anything, "source-new", "fresh").Return([]byte("owned payload"), nil).Once()
	dst.On("UploadReleaseAsset", mock.Anything, "target/repo", provider.UploadAssetRequest{ReleaseID: "target-new", FileName: "new.bin", Label: "binary", ContentType: "application/octet-stream", Content: []byte("owned payload")}).Return(&provider.Asset{ID: "uploaded"}, nil).Once()
	syncer := NewReleaseSyncer(src, dst, SyncOptions{Verbose: true})
	require.NoError(t, syncer.Sync(t.Context(), provider.Repository{FullName: "source/repo"}, provider.Repository{FullName: "target/repo"}))
	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	src.AssertNotCalled(t, "DownloadReleaseAsset", mock.Anything, "source-new", "already")
	dst.AssertNotCalled(t, "UpdateRelease", mock.Anything, "target/repo", "target-same", mock.Anything)
}

func TestReleaseSynchronizationEmptySourceAndListErrors(t *testing.T) {
	sentinel := errors.New("fixture unavailable")
	for _, scenario := range []string{"empty", "source failure", "destination failure"} {
		t.Run(scenario, func(t *testing.T) {
			src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
			switch scenario {
			case "empty":
				src.On("ListReleases", mock.Anything, "s/r", mock.Anything).Return(&provider.ReleaseList{}, nil).Once()
			case "source failure":
				src.On("ListReleases", mock.Anything, "s/r", mock.Anything).Return(nil, sentinel).Once()
			case "destination failure":
				src.On("ListReleases", mock.Anything, "s/r", mock.Anything).Return(&provider.ReleaseList{Releases: []provider.Release{{TagName: "v1"}}}, nil).Once()
				dst.On("ListReleases", mock.Anything, "d/r", mock.Anything).Return(nil, sentinel).Once()
			}
			err := NewReleaseSyncer(src, dst, SyncOptions{Verbose: true}).Sync(t.Context(), provider.Repository{FullName: "s/r"}, provider.Repository{FullName: "d/r"})
			if scenario == "empty" {
				require.NoError(t, err)
				require.Empty(t, dst.Calls)
			} else {
				require.ErrorIs(t, err, sentinel)
			}
			src.AssertExpectations(t)
			dst.AssertExpectations(t)
		})
	}
}

func TestReleaseMetadataAndForceTriggerUpdates(t *testing.T) {
	for _, field := range []string{"body", "draft", "prerelease", "force"} {
		t.Run(field, func(t *testing.T) {
			source := provider.Release{ID: "source", TagName: "v1", Name: "release"}
			destination := provider.Release{ID: "destination", TagName: "v1", Name: "release"}
			switch field {
			case "body":
				source.Body = "new notes"
			case "draft":
				source.Draft = true
			case "prerelease":
				source.Prerelease = true
			}
			src, dst := providermock.NewProvider("github"), providermock.NewProvider("gitlab")
			src.On("ListReleases", mock.Anything, "s/r", mock.Anything).Return(&provider.ReleaseList{Releases: []provider.Release{source}}, nil).Once()
			dst.On("ListReleases", mock.Anything, "d/r", mock.Anything).Return(&provider.ReleaseList{Releases: []provider.Release{destination}}, nil).Once()
			dst.On("UpdateRelease", mock.Anything, "d/r", "destination", mock.MatchedBy(func(update provider.UpdateReleaseRequest) bool {
				return update.Name != nil && *update.Name == source.Name && update.Body != nil && *update.Body == source.Body && update.Draft != nil && *update.Draft == source.Draft && update.Prerelease != nil && *update.Prerelease == source.Prerelease
			})).Return(&destination, nil).Once()
			require.NoError(t, NewReleaseSyncer(src, dst, SyncOptions{Force: field == "force"}).Sync(context.Background(), provider.Repository{FullName: "s/r"}, provider.Repository{FullName: "d/r"}))
			src.AssertExpectations(t)
			dst.AssertExpectations(t)
		})
	}
}

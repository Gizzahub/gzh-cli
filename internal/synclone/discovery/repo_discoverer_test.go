// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package discovery

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initRepo는 시험용 git 저장소를 하나 만든다. 원격 주소까지 붙여야
// parseRemoteURL 쪽까지 함께 지나간다.
func initRepo(t *testing.T, dir, remote string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o750))

	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", remote},
	} {
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
}

func TestDiscoverReposFindsRepositories(t *testing.T) {
	base := t.TempDir()
	initRepo(t, filepath.Join(base, "a"), "git@github.com:acme/a.git")
	initRepo(t, filepath.Join(base, "nested", "b"), "https://gitlab.com/acme/b.git")

	// 저장소가 아닌 디렉토리는 그냥 지나가야 한다.
	require.NoError(t, os.MkdirAll(filepath.Join(base, "plain"), 0o750))

	repos, err := NewRepoDiscoverer(base).DiscoverRepos(context.Background())
	require.NoError(t, err)
	require.Len(t, repos, 2)

	byOrgRepo := map[string]DiscoveredRepo{}
	for _, r := range repos {
		byOrgRepo[r.Provider+"/"+r.Org+"/"+r.RepoName] = r
	}

	assert.Contains(t, byOrgRepo, "github/acme/a")
	assert.Contains(t, byOrgRepo, "gitlab/acme/b")
}

// TestDiscoverReposStopsWhenCanceled는 맥락 취소가 훑기를 멈추는지 본다.
//
// 예전에는 세 번의 git 호출이 전부 context.Background()였다. 나무가 크면
// 바깥 프로세스가 수백 개 뜨는데 Ctrl+C로 끊을 길이 없었다.
func TestDiscoverReposStopsWhenCanceled(t *testing.T) {
	base := t.TempDir()
	initRepo(t, filepath.Join(base, "a"), "git@github.com:acme/a.git")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repos, err := NewRepoDiscoverer(base).DiscoverRepos(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, repos)
}

// TestAnalyzeRepositoryDropsPartialResultOnCancel는 취소로 git이 실패했을 때
// 빈 칸투성이 항목이 결과에 들어가지 않는지 본다.
//
// analyzeRepository는 git 실패를 일부러 삼킨다 -- 원격이 없는 저장소도 결과에
// 넣기 위해서다. 취소는 그 예외라서 따로 가른다.
func TestAnalyzeRepositoryDropsPartialResultOnCancel(t *testing.T) {
	base := t.TempDir()
	repoPath := filepath.Join(base, "a")
	initRepo(t, repoPath, "git@github.com:acme/a.git")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo, err := NewRepoDiscoverer(base).analyzeRepository(ctx, repoPath)

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, repo)
}

// TestDiscoverReposSkipsRepoWithoutRemote는 원격이 없어도 결과에 남는지 본다.
// 위의 취소 처리가 이 동작을 망가뜨리지 않았음을 못 박는다.
func TestDiscoverReposSkipsRepoWithoutRemote(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "bare")
	require.NoError(t, os.MkdirAll(dir, 0o750))

	out, err := exec.CommandContext(t.Context(), "git", "-C", dir, "init", "-q").CombinedOutput()
	require.NoError(t, err, "git init: %s", out)

	repos, err := NewRepoDiscoverer(base).DiscoverRepos(context.Background())
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Empty(t, repos[0].RemoteURL)
	assert.Equal(t, dir, repos[0].Path)
}

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package clone

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	providermock "github.com/gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// 세션 복원 이후 실제 runner 호출 수와 저장 옵션으로 덮어쓰기 회귀를 검증한다.
func TestExecutorResumeEffectiveOptions(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("APPDATA", root)
	saved := DefaultCloneOptions()
	saved.Provider, saved.Org, saved.Target = "github", "testorg", filepath.Join(root, "target")
	saved.MaxRetries, saved.Parallel, saved.RetryDelay = 8, 7, time.Millisecond
	session := NewSession(saved)
	session.AddRepository("testorg/completed")
	session.MarkCompleted("testorg/completed")
	session.AddRepository("testorg/pending")
	require.NoError(t, session.Initialize())
	effective := *saved
	effective.Resume, effective.MaxRetries, effective.Parallel = session.ID, 1, 2
	effective.Quiet = true
	fake := providermock.NewProvider("github")
	fake.SetupListResponse("testorg", []provider.Repository{
		{Name: "completed", FullName: "testorg/completed", CloneURL: "https://invalid.example/completed.git"},
		{Name: "pending", FullName: "testorg/pending", CloneURL: "https://invalid.example/pending.git"},
	})
	runner := &gitinternal.RecordingGitRunner{Err: ErrNetworkError}
	executor, err := NewCloneExecutorWithRunner(fake, &effective, runner)
	require.NoError(t, err)
	require.ErrorContains(t, executor.Execute(context.Background()), "1 of 2 repositories failed")
	require.Len(t, runner.CloneArgs(), 2, "one initial attempt plus requested one retry, not saved eight retries")
	require.NotContains(t, runner.CloneArgs()[0], "https://invalid.example/completed.git")
	require.Equal(t, 2, executor.options.Parallel)
	require.Equal(t, 1, executor.options.MaxRetries)
	restored := NewSession(nil)
	require.NoError(t, restored.Load(session.ID))
	require.Equal(t, session.ID, restored.ID)
	require.True(t, restored.IsCompleted("testorg/completed"))
	require.True(t, restored.IsFailed("testorg/pending"))
	require.Equal(t, 2, restored.Options.Parallel)
	require.Equal(t, 1, restored.Options.MaxRetries)
	require.Equal(t, session.ID, restored.Options.Resume)
	fake.AssertExpectations(t)
}

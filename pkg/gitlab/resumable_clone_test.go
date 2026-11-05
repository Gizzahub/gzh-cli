package gitlab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/httpclient"
	"github.com/gizzahub/gzh-cli/internal/workerpool"
	synclonepkg "github.com/gizzahub/gzh-cli/pkg/synclone"
)

// 테스트 바이너리를 git으로 실행하여 실제 네트워크와 Git을 사용하지 않는다.
func TestMain(m *testing.M) {
	if os.Getenv("GZH_GITLAB_RESUME_FIXTURE") == "1" {
		if dir := os.Getenv("GZH_GITLAB_BLOCK_STATE"); dir != "" {
			if os.Rename(dir, dir+".prior") != nil || os.WriteFile(dir, []byte("blocked"), 0o600) != nil {
				os.Exit(92)
			}
		}
		if os.Getenv("GZH_GITLAB_FAIL") == "1" {
			os.Exit(17)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type resumeRoundTripper func(*http.Request) (*http.Response, error)

func (f resumeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func resumeFixture(t *testing.T, repos string, beforeResponse func()) {
	t.Helper()
	client := httpclient.GetGlobalClient("gitlab")
	prior := client.Transport
	t.Cleanup(func() { client.Transport = prior })
	client.Transport = resumeRoundTripper(func(r *http.Request) (*http.Response, error) {
		if beforeResponse != nil {
			beforeResponse()
		}
		body := repos
		if strings.HasSuffix(r.URL.Path, "/subgroups") {
			body = `[]`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
}

func resumeFakeGit(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	exe, err := os.Executable()
	require.NoError(t, err)
	data, err := os.ReadFile(exe)
	require.NoError(t, err)
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
		t.Setenv("PATHEXT", ".EXE")
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), data, 0o750))
	t.Setenv("PATH", dir)
	t.Setenv("GZH_GITLAB_RESUME_FIXTURE", "1")
	t.Setenv("GZH_GITLAB_FAIL", "")
	t.Setenv("GZH_GITLAB_BLOCK_STATE", "")
}

func resumeManager(t *testing.T) *ResumableCloneManager {
	t.Helper()
	config := workerpool.DefaultRepositoryPoolConfig()
	config.RetryAttempts = 0
	manager := NewResumableCloneManager(config)
	manager.stateManager = synclonepkg.NewStateManager(t.TempDir())
	return manager
}

func TestResumeStatePathSafetyInitializationPreservesLoadErrors(t *testing.T) {
	for _, snapshot := range []string{"gitlab_example.json"} {
		for _, resume := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/resume=%v", snapshot, resume), func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, snapshot)
				data := []byte("{preserve corrupt snapshot")
				require.NoError(t, os.WriteFile(path, data, 0o600))
				manager := resumeManager(t)
				manager.stateManager = synclonepkg.NewStateManager(root)
				_, loadErr := manager.stateManager.LoadState("gitlab", "example")
				require.Error(t, loadErr)
				state, err := manager.initializeState("example", t.TempDir(), "pull", 1, 0, resume)
				require.Nil(t, state)
				require.ErrorContains(t, err, loadErr.Error())
				require.NotContains(t, err.Error(), "delete state file")
				after, err := os.ReadFile(path)
				require.NoError(t, err)
				require.Equal(t, data, after)
			})
		}
	}
}

func TestResumeStatePathSafetyInitializationAndResumeCompatibility(t *testing.T) {
	manager := resumeManager(t)
	target := t.TempDir()
	state, err := manager.initializeState("team/platform", target, "pull", 1, 0, false)
	require.NoError(t, err)
	require.NoError(t, manager.stateManager.SaveState(state))
	state, err = manager.initializeState("team/platform", target, "pull", 1, 0, false)
	require.Nil(t, state)
	require.ErrorContains(t, err, "existing state found")
	require.ErrorContains(t, err, "--resume")
	require.ErrorContains(t, err, manager.stateManager.GetStateFilePath("gitlab", "team/platform"))
	require.ErrorContains(t, err, "back up")
	require.NotContains(t, err.Error(), "delete state file")
	state, err = manager.initializeState("team/platform", target, "fetch", 2, 1, true)
	require.NoError(t, err)
	require.Equal(t, "fetch", state.Strategy)
	require.Equal(t, 2, state.Parallel)
	_, err = manager.initializeState("team/platform", t.TempDir(), "pull", 1, 0, true)
	require.ErrorContains(t, err, "target path mismatch")
	state, err = manager.initializeState("missing", target, "pull", 1, 0, true)
	require.NoError(t, err)
	require.Equal(t, "missing", state.Organization)
}

func TestResumeStatePathSafetyIgnoresUnrelatedCorruption(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "github_other.json")
	data := []byte("{preserve unrelated snapshot")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	manager := resumeManager(t)
	manager.stateManager = synclonepkg.NewStateManager(root)
	target := t.TempDir()
	state, err := manager.initializeState("example", target, "pull", 1, 0, false)
	require.NoError(t, err)
	require.NoError(t, manager.stateManager.SaveState(state))
	_, err = manager.initializeState("example", target, "pull", 1, 0, false)
	require.ErrorContains(t, err, "existing state found")
	require.ErrorContains(t, err, filepath.Join(root, "gitlab_example.json"))
	state, err = manager.initializeState("example", target, "fetch", 2, 0, true)
	require.NoError(t, err)
	require.Equal(t, "fetch", state.Strategy)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, after)
}

func TestResumeIntegrityFailureRetryAndCleanup(t *testing.T) {
	resumeFakeGit(t)
	resumeFixture(t, `[{"name":"repo"}]`, nil)
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "repo", ".git"), 0o750))
	t.Setenv("GZH_GITLAB_FAIL", "1")
	manager := resumeManager(t)
	err := manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet")
	require.ErrorContains(t, err, "1 operations failed")
	state, err := manager.stateManager.LoadState("gitlab", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
	t.Setenv("GZH_GITLAB_FAIL", "")
	require.NoError(t, manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"))
	require.False(t, manager.stateManager.HasState("gitlab", "example"))
}

func TestResumeIntegrityOmittedFailure(t *testing.T) {
	for _, body := range []string{`[]`, `[{"name":"done"}]`} {
		t.Run(body, func(t *testing.T) {
			target := t.TempDir()
			sm := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
			state := synclonepkg.NewCloneState("gitlab", "example", target, "pull", 1, 0)
			state.AddFailedRepository("missing", filepath.Join(target, "missing"), "pull", "original failure", 1)
			state.AddCompletedRepository("done", filepath.Join(target, "done"), "pull", "ok")
			require.NoError(t, sm.SaveState(state))
			resumeFixture(t, body, nil)
			manager := resumeManager(t)
			require.ErrorContains(t, manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "1 operations failed")
			state, err := sm.LoadState("gitlab", "example")
			require.NoError(t, err)
			require.Equal(t, "failed", state.Status)
			require.Equal(t, "original failure", state.FailedRepos[0].Error)
		})
	}
}

func TestResumeIntegrityNoWorkSaveFailure(t *testing.T) {
	target := t.TempDir()
	dir := filepath.Join(target, ".gzh", "state")
	sm := synclonepkg.NewStateManager(dir)
	state := synclonepkg.NewCloneState("gitlab", "example", target, "pull", 1, 0)
	state.AddCompletedRepository("done", target, "pull", "ok")
	require.NoError(t, sm.SaveState(state))
	blocked := false
	resumeFixture(t, `[{"name":"done"}]`, func() {
		if !blocked {
			require.NoError(t, os.Rename(dir, dir+".prior"))
			require.NoError(t, os.WriteFile(dir, []byte("blocked"), 0o600))
			blocked = true
		}
	})
	require.ErrorContains(t, resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "save final state")
}

func TestResumeIntegrityEmptyAPIWithPending(t *testing.T) {
	resumeFakeGit(t)
	t.Setenv("GZH_GITLAB_FAIL", "1")
	resumeFixture(t, `[]`, nil)
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "repo", ".git"), 0o750))
	sm := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	state := synclonepkg.NewCloneState("gitlab", "example", target, "pull", 1, 0)
	state.SetPendingRepositories([]string{"repo"})
	require.NoError(t, sm.SaveState(state))
	require.ErrorContains(t, resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "1 operations failed")
	state, err := sm.LoadState("gitlab", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
}

func TestResumeIntegrityNewWorkPreservesOmittedFailure(t *testing.T) {
	resumeFakeGit(t)
	resumeFixture(t, `[{"name":"new"}]`, nil)
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "new", ".git"), 0o750))
	sm := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	state := synclonepkg.NewCloneState("gitlab", "example", target, "pull", 1, 0)
	state.AddFailedRepository("omitted", target, "pull", "original failure", 1)
	require.NoError(t, sm.SaveState(state))
	require.ErrorContains(t, resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "1 operations failed")
	state, err := sm.LoadState("gitlab", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsCompleted("new"))
	require.True(t, state.IsFailed("omitted"))
	require.Equal(t, 2, state.TotalRepositories)
}

func TestResumeIntegrityEmptyNewGroup(t *testing.T) {
	resumeFixture(t, `[]`, nil)
	manager := resumeManager(t)
	target := t.TempDir()
	require.NoError(t, manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet"))
	require.False(t, manager.stateManager.HasState("gitlab", "example"))
	_, err := os.Stat(filepath.Join(target, ".gzh"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestResumeIntegrityCancellationSaveFailure(t *testing.T) {
	manager := resumeManager(t)
	blocked := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(blocked, []byte("retain"), 0o600))
	manager.stateManager = synclonepkg.NewStateManager(filepath.Join(blocked, "state"))
	state := synclonepkg.NewCloneState("gitlab", "example", t.TempDir(), "pull", 1, 0)
	err := manager.saveCancellation(state, context.Canceled)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorContains(t, err, "save cancelled state")
}

func TestResumeIntegrityFinalSaveFailure(t *testing.T) {
	resumeFakeGit(t)
	resumeFixture(t, `[{"name":"repo"}]`, nil)
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "repo", ".git"), 0o750))
	dir := filepath.Join(target, ".gzh", "state")
	t.Setenv("GZH_GITLAB_BLOCK_STATE", dir)
	require.ErrorContains(t, resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet"), "save final state")
	state, err := synclonepkg.NewStateManager(dir+".prior").LoadState("gitlab", "example")
	require.NoError(t, err)
	require.Equal(t, []string{"repo"}, state.PendingRepos)
}

func TestResumeIntegrityFailureAndFinalSaveFailure(t *testing.T) {
	resumeFakeGit(t)
	resumeFixture(t, `[{"name":"repo"}]`, nil)
	t.Setenv("GZH_GITLAB_FAIL", "1")
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "repo", ".git"), 0o750))
	t.Setenv("GZH_GITLAB_BLOCK_STATE", filepath.Join(target, ".gzh", "state"))
	err := resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet")
	require.ErrorContains(t, err, "save final state")
	require.ErrorContains(t, err, "1 operations failed")
	require.ErrorContains(t, err, "repository repo:")
	require.ErrorContains(t, err, "git pull failed")
}

func TestResumeIntegrityManyFailuresAndSaveFailure(t *testing.T) {
	const failureCount = 128
	target := t.TempDir()
	dir := filepath.Join(target, ".gzh", "state")
	sm := synclonepkg.NewStateManager(dir)
	state := synclonepkg.NewCloneState("gitlab", "example", target, "pull", 1, 0)
	want := []string{fmt.Sprintf("%d operations failed", failureCount)}
	for i := range failureCount {
		name := fmt.Sprintf("repo-%03d", i)
		message := fmt.Sprintf("failure detail %03d", i)
		state.AddFailedRepository(name, filepath.Join(target, name), "pull", message, 1)
		want = append(want, fmt.Sprintf("repository %s: %s", name, message))
	}
	require.NoError(t, sm.SaveState(state))
	blocked := false
	resumeFixture(t, `[]`, func() {
		if !blocked {
			require.NoError(t, os.Rename(dir, dir+".prior"))
			require.NoError(t, os.WriteFile(dir, []byte("blocked"), 0o600))
			blocked = true
		}
	})
	err := resumeManager(t).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet")
	require.Error(t, err)
	lines := strings.Split(err.Error(), "\n")
	require.Len(t, lines, failureCount+2)
	require.Equal(t, want, lines[:failureCount+1])
	require.Contains(t, lines[failureCount+1], "failed to save final state:")
	retained, loadErr := synclonepkg.NewStateManager(dir+".prior").LoadState("gitlab", "example")
	require.NoError(t, loadErr)
	require.Len(t, retained.FailedRepos, failureCount)
}

func TestResumeIntegrityTickerWaitsForResult(t *testing.T) {
	manager := resumeManager(t)
	pool := workerpool.NewRepositoryWorkerPool(workerpool.DefaultRepositoryPoolConfig())
	require.NoError(t, pool.Start())
	defer pool.Stop()
	state := synclonepkg.NewCloneState("gitlab", "example", t.TempDir(), "pull", 1, 0)
	state.SetPendingRepositories([]string{"slow"})
	job := workerpool.RepositoryJob{Repository: "slow", Operation: workerpool.OperationPull}
	release := make(chan struct{})
	defer close(release)
	require.NoError(t, pool.SubmitJob(job, func(ctx context.Context, _ workerpool.RepositoryJob) error {
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}))
	done := make(chan error, 1)
	go func() {
		success, failed, err := manager.trackResultsWithPeriodicSaving(t.Context(), pool, []workerpool.RepositoryJob{job}, state, synclonepkg.NewProgressTracker([]string{"slow"}, synclonepkg.DisplayModeQuiet))
		if err == nil && (success != 1 || failed != 0) {
			err = errors.New("missing actual result")
		}
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("returned before job completion: %v", err)
	case <-time.After(1200 * time.Millisecond):
	}
	release <- struct{}{}
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("result was not received")
	}
	require.True(t, state.IsCompleted("slow"))
}

func TestResumeIntegrityCancellation(t *testing.T) {
	manager := resumeManager(t)
	pool := workerpool.NewRepositoryWorkerPool(workerpool.DefaultRepositoryPoolConfig())
	require.NoError(t, pool.Start())
	defer pool.Stop()
	state := synclonepkg.NewCloneState("gitlab", "example", t.TempDir(), "pull", 1, 0)
	state.SetPendingRepositories([]string{"pending"})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err := manager.trackResultsWithPeriodicSaving(ctx, pool, []workerpool.RepositoryJob{{Repository: "pending"}}, state, synclonepkg.NewProgressTracker([]string{"pending"}, synclonepkg.DisplayModeQuiet))
	require.ErrorIs(t, err, context.Canceled)
	saved, err := manager.stateManager.LoadState("gitlab", "example")
	require.NoError(t, err)
	require.Equal(t, "canceled", saved.Status)
	require.Equal(t, []string{"pending"}, saved.PendingRepos)
}

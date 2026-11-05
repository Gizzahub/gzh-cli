package github

import (
	"context"
	"encoding/json"
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

// TestMain은 실제 Git 대신 테스트 바이너리로 명령 인자와 실패를 재현한다.
func TestMain(m *testing.M) {
	if logPath := os.Getenv("GZH_RESUME_GIT_LOG"); logPath != "" {
		data, err := json.Marshal(os.Args[1:])
		if err != nil {
			os.Exit(90)
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			os.Exit(91)
		}
		_, writeErr := file.Write(append(data, '\n'))
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			os.Exit(92)
		}
		if os.Getenv("GZH_RESUME_GIT_BLOCK") == "1" {
			time.Sleep(10 * time.Second)
		}
		if dir := os.Getenv("GZH_RESUME_BLOCK_STATE"); dir != "" {
			if os.Rename(dir, dir+".prior") != nil || os.WriteFile(dir, []byte("blocked"), 0o600) != nil {
				os.Exit(93)
			}
		}
		for _, arg := range os.Args[1:] {
			if arg == os.Getenv("GZH_RESUME_GIT_FAIL") {
				os.Exit(17)
			}
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func newResumeTestManager(t *testing.T) *ResumableCloneManager {
	t.Helper()
	manager := NewResumableCloneManager(DefaultBulkOperationsConfig())
	manager.stateManager = synclonepkg.NewStateManager(t.TempDir())
	return manager
}

func TestResumeStatePreservesCorruptFile(t *testing.T) {
	manager := newResumeTestManager(t)
	path := manager.stateManager.GetStateFilePath("github", "example")
	data := []byte("{broken state: retain for recovery")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	state, err := manager.initializeOrLoadState("example", t.TempDir(), "pull", 2, 1, false)
	require.Error(t, err)
	require.Nil(t, state)
	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, data, got)
}

func TestResumeStateLifecycle(t *testing.T) {
	manager := newResumeTestManager(t)
	target := t.TempDir()
	state, err := manager.initializeOrLoadState("example", target, "pull", 2, 1, false)
	require.NoError(t, err)
	require.Equal(t, target, state.TargetPath)
	state.SetPendingRepositories([]string{"done", "failed", "pending"})
	state.AddCompletedRepository("done", filepath.Join(target, "done"), "pull", "ok")
	state.AddFailedRepository("failed", filepath.Join(target, "failed"), "pull", "fixture", 1)
	require.NoError(t, manager.stateManager.SaveState(state))
	path := manager.stateManager.GetStateFilePath("github", "example")
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	_, err = manager.initializeOrLoadState("example", target, "pull", 2, 1, false)
	require.ErrorContains(t, err, "--resume")
	_, err = manager.initializeOrLoadState("example", t.TempDir(), "pull", 2, 1, true)
	require.ErrorContains(t, err, "target path mismatch")
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
	resumed, err := manager.initializeOrLoadState("example", target, "fetch", 4, 3, true)
	require.NoError(t, err)
	require.Equal(t, "fetch", resumed.Strategy)
	require.Equal(t, 4, resumed.Parallel)
	require.Equal(t, 3, resumed.MaxRetries)
	require.Equal(t, []string{"pending", "new"}, manager.getResumeRepositories(resumed, []string{"done", "failed", "pending", "new", "new"}))
	require.Len(t, resumed.CompletedRepos, 1)
	require.Equal(t, state.CompletedRepos[0].Name, resumed.CompletedRepos[0].Name)
	require.True(t, state.CompletedRepos[0].CompletedAt.Equal(resumed.CompletedRepos[0].CompletedAt))
	require.Len(t, resumed.FailedRepos, 1)
	require.Equal(t, state.FailedRepos[0].Error, resumed.FailedRepos[0].Error)
	_, err = manager.initializeOrLoadState("missing", target, "pull", 2, 1, true)
	require.ErrorContains(t, err, "failed to load state for resume")
}

func TestResumeIntegrityGuidanceNamesDedicatedSnapshot(t *testing.T) {
	target := t.TempDir()
	manager := newResumeTestManager(t)
	manager.stateManager = synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	state := synclonepkg.NewCloneState("github", "example", target, "pull", 1, 0)
	state.AddFailedRepository("repo", filepath.Join(target, "repo"), "pull", "connection failed", 1)
	state.AddFailedRepository("existing", filepath.Join(target, "existing directory"), "clone", "destination path already exists", 1)
	state.AddFailedRepository("remote", filepath.Join(target, "remote repository"), "pull", "remote unavailable", 1)
	require.NoError(t, manager.stateManager.SaveState(state))
	path := manager.stateManager.GetStateFilePath("github", "example")
	_, err := manager.createNewState("example", target, "pull", 1, 0)
	require.ErrorContains(t, err, "--resume")
	require.ErrorContains(t, err, path)
	require.ErrorContains(t, err, "back up")
	require.ErrorContains(t, err, "move this snapshot to a backup location")
	require.NotContains(t, err.Error(), "--all")
	require.NotContains(t, err.Error(), "gz synclone state")

	// 실패 안내의 실제 출력을 전용 파일에 모아 복구 경로와 보존 안내를 검증한다.
	output, err := os.Create(filepath.Join(t.TempDir(), "failure-output"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, output.Close()) })
	prior := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = prior }()
	manager.printFailureDetails(state)
	os.Stdout = prior
	_, err = output.Seek(0, io.SeekStart)
	require.NoError(t, err)
	data, err := io.ReadAll(output)
	require.NoError(t, err)
	message := string(data)
	require.Contains(t, message, path)
	require.Contains(t, message, "back up")
	require.Contains(t, message, "Preserve this file for recovery")
	require.Contains(t, message, "Clone destination already exists")
	require.NotContains(t, message, "not a valid git repository")
	require.Contains(t, message, filepath.Join(target, "existing directory"))
	require.Contains(t, message, filepath.Join(target, "remote repository"))
	require.NotContains(t, message, "rm -rf")
	require.NotContains(t, message, "--all")
	require.NotContains(t, message, "gz synclone state")
	require.NotContains(t, message, "review and clear")
	require.True(t, manager.stateManager.HasState("github", "example"))
}

func TestResumeIntegrityGuidanceReportsUnverifiableSnapshot(t *testing.T) {
	manager := newResumeTestManager(t)
	state := synclonepkg.NewCloneState("github", "example", t.TempDir(), "pull", 1, 0)
	state.AddFailedRepository("repo", filepath.Join(state.TargetPath, "repo"), "pull", "connection failed", 1)
	path := manager.stateManager.GetStateFilePath("github", "example")
	require.NoError(t, os.WriteFile(path, []byte("{preserve corrupt snapshot"), 0o600))
	output, err := os.Create(filepath.Join(t.TempDir(), "failure-output"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, output.Close()) })
	prior := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = prior }()
	manager.printFailureDetails(state)
	os.Stdout = prior
	_, err = output.Seek(0, io.SeekStart)
	require.NoError(t, err)
	data, err := io.ReadAll(output)
	require.NoError(t, err)
	require.Contains(t, string(data), "Snapshot path could not be verified:")
	require.NotContains(t, string(data), "Saved snapshot:")
}

func TestResumeStateDifferentTargetAndSaveFailure(t *testing.T) {
	manager := newResumeTestManager(t)
	state := synclonepkg.NewCloneState("github", "example", t.TempDir(), "pull", 2, 1)
	require.NoError(t, manager.stateManager.SaveState(state))
	target := t.TempDir()
	replacement, err := manager.createNewState("example", target, "fetch", 1, 0)
	require.NoError(t, err)
	require.Equal(t, target, replacement.TargetPath)
	require.False(t, manager.stateManager.HasState("github", "example"))
	blocked := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(blocked, []byte("retain"), 0o600))
	manager.stateManager = synclonepkg.NewStateManager(filepath.Join(blocked, "state"))
	require.Error(t, manager.stateManager.SaveState(replacement))
	data, err := os.ReadFile(blocked)
	require.NoError(t, err)
	require.Equal(t, "retain", string(data))
}

func resumeGitFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	executable, err := os.Executable()
	require.NoError(t, err)
	data, err := os.ReadFile(executable)
	require.NoError(t, err)
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
		t.Setenv("PATHEXT", ".EXE")
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), data, 0o750))
	logPath := filepath.Join(dir, "calls.jsonl")
	t.Setenv("PATH", dir)
	t.Setenv("GZH_RESUME_GIT_LOG", logPath)
	t.Setenv("GZH_RESUME_GIT_FAIL", "")
	t.Setenv("GZH_RESUME_GIT_BLOCK", "")
	t.Setenv("GZH_RESUME_BLOCK_STATE", "")
	return logPath
}

// 전역 HTTP 클라이언트는 테스트 동안만 메모리 응답으로 교체한다.
func resumeListFixture(t *testing.T, body string, beforeResponse func()) {
	t.Helper()
	client := httpclient.GetGlobalClient("github")
	prior := client.Transport
	t.Cleanup(func() { client.Transport = prior })
	client.Transport = providerBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		if beforeResponse != nil {
			beforeResponse()
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
}

func TestResumeIntegrityFailureAndNoRetry(t *testing.T) {
	logPath := resumeGitFixture(t)
	resumeListFixture(t, `[{"name":"repo"}]`, nil)
	t.Setenv("GZH_RESUME_GIT_FAIL", "pull")
	target := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(target, "repo"), 0o750))
	config := DefaultBulkOperationsConfig()
	config.PoolConfig.RetryAttempts = 0
	manager := NewResumableCloneManager(config)
	err := manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet")
	require.ErrorContains(t, err, "1 operations failed")
	state, err := manager.stateManager.LoadState("github", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
	require.Empty(t, state.PendingRepos)
	require.Len(t, resumeGitCalls(t, logPath), 1)
	err = manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet")
	require.ErrorContains(t, err, "1 operations failed")
	state, err = manager.stateManager.LoadState("github", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
	require.Len(t, resumeGitCalls(t, logPath), 1)
	resumeListFixture(t, `[]`, nil)
	require.ErrorContains(t, manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "1 operations failed")
	state, err = manager.stateManager.LoadState("github", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
}

func TestResumeIntegritySuccessAndFinalSaveFailure(t *testing.T) {
	for _, blockSave := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "save failure"}[blockSave], func(t *testing.T) {
			resumeGitFixture(t)
			resumeListFixture(t, `[{"name":"repo"}]`, nil)
			target := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(target, "repo"), 0o750))
			if blockSave {
				t.Setenv("GZH_RESUME_BLOCK_STATE", filepath.Join(target, ".gzh", "state"))
			}
			manager := NewResumableCloneManager(DefaultBulkOperationsConfig())
			err := manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet")
			if blockSave {
				require.ErrorContains(t, err, "save final state")
				prior := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state.prior"))
				state, loadErr := prior.LoadState("github", "example")
				require.NoError(t, loadErr)
				require.Equal(t, []string{"repo"}, state.PendingRepos)
			} else {
				require.NoError(t, err)
				require.False(t, manager.stateManager.HasState("github", "example"))
			}
		})
	}
}

func TestResumeIntegrityNoWorkSaveFailure(t *testing.T) {
	target := t.TempDir()
	dir := filepath.Join(target, ".gzh", "state")
	sm := synclonepkg.NewStateManager(dir)
	require.NoError(t, sm.SaveState(synclonepkg.NewCloneState("github", "example", target, "pull", 1, 0)))
	resumeListFixture(t, `[]`, func() {
		require.NoError(t, os.Rename(dir, dir+".prior"))
		require.NoError(t, os.WriteFile(dir, []byte("blocked"), 0o600))
	})
	manager := NewResumableCloneManager(DefaultBulkOperationsConfig())
	require.ErrorContains(t, manager.RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "save final state")
}

func TestResumeIntegrityFailureAndFinalSaveFailure(t *testing.T) {
	resumeGitFixture(t)
	resumeListFixture(t, `[{"name":"repo"}]`, nil)
	t.Setenv("GZH_RESUME_GIT_FAIL", "pull")
	target := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(target, "repo"), 0o750))
	t.Setenv("GZH_RESUME_BLOCK_STATE", filepath.Join(target, ".gzh", "state"))
	config := DefaultBulkOperationsConfig()
	config.PoolConfig.RetryAttempts = 0
	err := NewResumableCloneManager(config).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, false, "quiet")
	require.ErrorContains(t, err, "save final state")
	require.ErrorContains(t, err, "1 operations failed")
	require.ErrorContains(t, err, "repository repo:")
	require.ErrorContains(t, err, "exit status 17")
}

func TestResumeIntegrityManyFailuresAndSaveFailure(t *testing.T) {
	const failureCount = 128
	target := t.TempDir()
	dir := filepath.Join(target, ".gzh", "state")
	sm := synclonepkg.NewStateManager(dir)
	state := synclonepkg.NewCloneState("github", "example", target, "pull", 1, 0)
	want := []string{fmt.Sprintf("%d operations failed", failureCount)}
	for i := range failureCount {
		name := fmt.Sprintf("repo-%03d", i)
		message := fmt.Sprintf("failure detail %03d", i)
		state.AddFailedRepository(name, filepath.Join(target, name), "pull", message, 1)
		want = append(want, fmt.Sprintf("repository %s: %s", name, message))
	}
	require.NoError(t, sm.SaveState(state))
	resumeListFixture(t, `[]`, func() {
		require.NoError(t, os.Rename(dir, dir+".prior"))
		require.NoError(t, os.WriteFile(dir, []byte("blocked"), 0o600))
	})
	err := NewResumableCloneManager(DefaultBulkOperationsConfig()).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet")
	require.Error(t, err)
	lines := strings.Split(err.Error(), "\n")
	require.Len(t, lines, failureCount+2)
	require.Equal(t, want, lines[:failureCount+1])
	require.Contains(t, lines[failureCount+1], "failed to save final state:")
	retained, loadErr := synclonepkg.NewStateManager(dir+".prior").LoadState("github", "example")
	require.NoError(t, loadErr)
	require.Len(t, retained.FailedRepos, failureCount)
}

func TestResumeIntegrityCancellation(t *testing.T) {
	logPath := resumeGitFixture(t)
	t.Setenv("GZH_RESUME_GIT_BLOCK", "1")
	resumeListFixture(t, `[{"name":"repo"}]`, nil)
	target := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(target, "repo"), 0o750))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	manager := NewResumableCloneManager(DefaultBulkOperationsConfig())
	done := make(chan error, 1)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		done <- manager.RefreshAllResumable(ctx, target, "example", "pull", 1, 0, false, "quiet")
	}()
	defer func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
		}
	}()
	require.Eventually(t, func() bool { _, err := os.Stat(logPath); return err == nil }, 5*time.Second, 10*time.Millisecond)
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation did not finish")
	}
	state, err := manager.stateManager.LoadState("github", "example")
	require.NoError(t, err)
	require.Equal(t, "canceled", state.Status)
	require.Equal(t, []string{"repo"}, state.PendingRepos)
}

func TestResumeIntegrityEmptyAPIWithPending(t *testing.T) {
	resumeGitFixture(t)
	t.Setenv("GZH_RESUME_GIT_FAIL", "pull")
	resumeListFixture(t, `[]`, nil)
	target := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(target, "repo"), 0o750))
	sm := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	state := synclonepkg.NewCloneState("github", "example", target, "pull", 1, 0)
	state.SetPendingRepositories([]string{"repo"})
	require.NoError(t, sm.SaveState(state))
	config := DefaultBulkOperationsConfig()
	config.PoolConfig.RetryAttempts = 0
	require.ErrorContains(t, NewResumableCloneManager(config).RefreshAllResumable(t.Context(), target, "example", "pull", 1, 0, true, "quiet"), "1 operations failed")
	state, err := sm.LoadState("github", "example")
	require.NoError(t, err)
	require.Equal(t, "failed", state.Status)
	require.True(t, state.IsFailed("repo"))
}

func TestResumeIntegrityEmptyNewOrganization(t *testing.T) {
	resumeListFixture(t, `[]`, nil)
	manager := NewResumableCloneManager(DefaultBulkOperationsConfig())
	require.NoError(t, manager.RefreshAllResumable(t.Context(), t.TempDir(), "example", "pull", 1, 0, false, "quiet"))
	require.False(t, manager.stateManager.HasState("github", "example"))
}

func resumeGitCalls(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })
	decoder := json.NewDecoder(file)
	var calls [][]string
	for decoder.More() {
		var call []string
		require.NoError(t, decoder.Decode(&call))
		calls = append(calls, call)
	}
	return calls
}

func TestResumeRepositoryJobGitArguments(t *testing.T) {
	for _, test := range []struct {
		name string
		op   workerpool.RepositoryOperation
		args [][]string
	}{
		{"pull", workerpool.OperationPull, [][]string{{"pull"}}},
		{"fetch", workerpool.OperationFetch, [][]string{{"fetch"}}},
		{"reset", workerpool.OperationReset, [][]string{{"reset", "--hard", "HEAD"}, {"pull"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			logPath := resumeGitFixture(t)
			path := filepath.Join(t.TempDir(), "repo with spaces")
			require.NoError(t, processRepositoryJob(t.Context(), workerpool.RepositoryJob{Operation: test.op, Path: path}, "example"))
			want := make([][]string, len(test.args))
			for i, args := range test.args {
				want[i] = append([]string{"-C", path}, args...)
			}
			require.Equal(t, want, resumeGitCalls(t, logPath))
		})
	}
}

func TestResumeRepositoryJobFailures(t *testing.T) {
	for _, operation := range []workerpool.RepositoryOperation{workerpool.OperationPull, workerpool.OperationFetch, workerpool.OperationReset} {
		t.Run(string(operation), func(t *testing.T) {
			logPath := resumeGitFixture(t)
			t.Setenv("GZH_RESUME_GIT_FAIL", string(operation))
			err := processRepositoryJob(t.Context(), workerpool.RepositoryJob{Operation: operation, Path: t.TempDir()}, "example")
			require.ErrorContains(t, err, "git "+string(operation)+" failed")
			require.Len(t, resumeGitCalls(t, logPath), 1)
		})
	}
	for _, operation := range []workerpool.RepositoryOperation{workerpool.OperationConfig, "unknown"} {
		require.Error(t, processRepositoryJob(t.Context(), workerpool.RepositoryJob{Operation: operation}, "example"))
	}
}

func TestResumeRepositoryJobClone(t *testing.T) {
	logPath := resumeGitFixture(t)
	path := filepath.Join(t.TempDir(), "repository")
	job := workerpool.RepositoryJob{Operation: workerpool.OperationClone, Path: path, Repository: "repository"}
	require.NoError(t, processRepositoryJob(t.Context(), job, "example"))
	require.Equal(t, [][]string{{"-C", filepath.Dir(path), "clone", "https://github.com/example/repository.git", "repository"}}, resumeGitCalls(t, logPath))
	t.Setenv("GZH_RESUME_GIT_FAIL", "clone")
	require.ErrorContains(t, processRepositoryJob(t.Context(), job, "example"), "Clone Failed")
}

func TestResumeRepositoryJobResetPullFailure(t *testing.T) {
	logPath := resumeGitFixture(t)
	t.Setenv("GZH_RESUME_GIT_FAIL", "pull")
	path := t.TempDir()
	err := processRepositoryJob(t.Context(), workerpool.RepositoryJob{Operation: workerpool.OperationReset, Path: path}, "example")
	require.ErrorContains(t, err, "git pull failed")
	require.Equal(t, [][]string{{"-C", path, "reset", "--hard", "HEAD"}, {"-C", path, "pull"}}, resumeGitCalls(t, logPath))
}

func TestResumeRepositoryJobCancellation(t *testing.T) {
	logPath := resumeGitFixture(t)
	t.Setenv("GZH_RESUME_GIT_BLOCK", "1")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	path := t.TempDir()
	go func() {
		done <- processRepositoryJob(ctx, workerpool.RepositoryJob{Operation: workerpool.OperationFetch, Path: path}, "example")
	}()
	require.Eventually(t, func() bool { _, err := os.Stat(logPath); return err == nil }, 5*time.Second, 10*time.Millisecond)
	cancel()
	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Git child did not stop after cancellation")
	}
}

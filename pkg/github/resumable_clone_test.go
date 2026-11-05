package github

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	return logPath
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

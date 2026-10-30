// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSyncTargetAndOptionValidation(t *testing.T) {
	for _, input := range []string{"", "github", ":team", "github:", "github:/repo", "github:team/"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTarget(input)
			require.Error(t, err)
		})
	}
	for _, tc := range []struct {
		input, canonical, full string
		repository             bool
	}{
		{" github : team ", "github:team", "team", false},
		{"gitlab: team / project ", "gitlab:team/project", "team/project", true},
	} {
		target, err := ParseTarget(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.canonical, target.String())
		require.Equal(t, tc.full, target.FullName())
		require.Equal(t, tc.repository, target.IsRepository())
		require.Equal(t, !tc.repository, target.IsOrganization())
	}
	base := SyncOptions{From: "github:source/repo", To: "gitlab:target/repo", IncludeCode: true}
	for _, tc := range []struct {
		name, message string
		edit          func(*SyncOptions)
	}{
		{"missing source", "source", func(o *SyncOptions) { o.From = "" }},
		{"missing destination", "destination", func(o *SyncOptions) { o.To = "" }},
		{"bad source", "invalid source target", func(o *SyncOptions) { o.From = "missing-colon" }},
		{"bad destination", "invalid destination target", func(o *SyncOptions) { o.To = "gitlab:" }},
		{"mixed scope", "same type", func(o *SyncOptions) { o.To = "gitlab:target" }},
		{"too many workers", "exceed 20", func(o *SyncOptions) { o.Parallel = 21 }},
		{"no feature", "at least one", func(o *SyncOptions) { o.IncludeCode = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := base
			tc.edit(&opts)
			require.ErrorContains(t, opts.Validate(), tc.message)
		})
	}
	for _, parallel := range []int{-1, 0, 1, 20} {
		opts := base
		opts.Parallel = parallel
		require.NoError(t, opts.Validate())
		require.GreaterOrEqual(t, opts.Parallel, 1)
		source, err := opts.GetSourceTarget()
		require.NoError(t, err)
		require.Equal(t, "source/repo", source.FullName())
		destination, err := opts.GetDestinationTarget()
		require.NoError(t, err)
		require.Equal(t, "target/repo", destination.FullName())
	}
}

// HOME과 모든 임시 경로를 시험 소유 디렉터리로 고정한다.
func isolatedTrackerHome(t *testing.T) (string, string) {
	t.Helper()
	home, temporary := t.TempDir(), t.TempDir()
	for _, name := range []string{"HOME", "USERPROFILE"} {
		t.Setenv(name, home)
	}
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, temporary)
	}
	return home, temporary
}

func TestTrackerPersistsProgressErrorsAndTerminalStates(t *testing.T) {
	isolatedTrackerHome(t)
	tracker := NewSyncTracker("roundtrip", "github:source", "gitlab:target", SyncOptions{IncludeCode: true})
	require.Equal(t, StatusPending, tracker.Status)
	require.False(t, tracker.IsCompleted())
	require.False(t, tracker.HasErrors())
	require.Zero(t, tracker.GetOverallProgress())
	tracker.StartedAt = time.Now().Add(-2 * time.Minute)
	tracker.Progress = map[string]SyncProgress{"code": {Total: 4}}
	tracker.UpdateProgress("code", 2, 1)
	loaded, err := LoadSyncTracker(tracker.ID)
	require.NoError(t, err)
	require.Equal(t, 2, loaded.Statistics.CompletedRepositories)
	require.Equal(t, 1, loaded.Statistics.FailedRepositories)
	require.Equal(t, 1, loaded.Statistics.SkippedRepositories)
	require.Equal(t, 75.0, loaded.GetOverallProgress())
	require.Positive(t, loaded.Statistics.AverageDuration)
	tracker.AddError("source/repo", "code", "fixture rejected push")
	loaded, err = LoadSyncTracker(tracker.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Errors, 1)
	require.Equal(t, "fixture rejected push", loaded.Errors[0].Message)
	require.False(t, loaded.Errors[0].Timestamp.IsZero())
	for _, status := range []SyncStatus{StatusCompleted, StatusFailed, StatusCancelled} {
		tracker.SetStatus(status)
		loaded, err = LoadSyncTracker(tracker.ID)
		require.NoError(t, err)
		require.True(t, loaded.IsCompleted())
		require.NotNil(t, loaded.CompletedAt)
		require.Positive(t, loaded.Statistics.TotalDuration)
		summary := loaded.GetSummary()
		require.Equal(t, status, summary.Status)
		require.Equal(t, "roundtrip", summary.ID)
		require.Equal(t, 1, summary.ErrorCount)
		require.Equal(t, 75.0, summary.Progress)
	}
}

func TestTrackerListingIgnoresUnrelatedAndCorruptFiles(t *testing.T) {
	isolatedTrackerHome(t)
	listed, err := ListSyncTrackers()
	require.NoError(t, err)
	require.Empty(t, listed)
	directory := getSyncDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "broken.json"), []byte("{"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "notes.txt"), []byte("owned fixture"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(directory, "folder.json"), 0o700))
	tracker := NewSyncTracker("valid", "a", "b", SyncOptions{})
	tracker.SetStatus(StatusCompleted)
	listed, err = ListSyncTrackers()
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, "valid", listed[0].ID)
	_, err = LoadSyncTracker("broken")
	require.ErrorContains(t, err, "unmarshal")
	_, err = LoadSyncTracker("absent")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.NoError(t, DeleteSyncTracker("valid"))
	require.NoError(t, DeleteSyncTracker("valid"))
	_, err = LoadSyncTracker("valid")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestTrackerCleanupRetainsRecentStateAndOwnsFallback(t *testing.T) {
	home, temporary := isolatedTrackerHome(t)
	old := NewSyncTracker("old-completed", "a", "b", SyncOptions{})
	old.StartedAt = time.Now().Add(-48 * time.Hour)
	old.SetStatus(StatusCompleted)
	recent := NewSyncTracker("recent", "a", "b", SyncOptions{})
	recent.SetStatus(StatusInProgress)
	require.NoError(t, CleanupOldTrackers(24*time.Hour))
	_, err := LoadSyncTracker(old.ID)
	require.ErrorIs(t, err, fs.ErrNotExist)
	loaded, err := LoadSyncTracker(recent.ID)
	require.NoError(t, err)
	require.Equal(t, StatusInProgress, loaded.Status)

	blockedHome := filepath.Join(home, "blocked")
	require.NoError(t, os.Mkdir(blockedHome, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(blockedHome, ".config"), []byte("not a directory"), 0o600))
	t.Setenv("HOME", blockedHome)
	t.Setenv("USERPROFILE", blockedHome)
	require.Equal(t, filepath.Join(temporary, "gzh-sync"), getSyncDir())
}

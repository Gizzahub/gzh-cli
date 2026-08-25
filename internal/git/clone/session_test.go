// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package clone

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionConcurrentProgressSaves(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	session := NewSession(&CloneOptions{})
	const repositoryCount = 24
	for index := range repositoryCount {
		session.AddRepository(fmt.Sprintf("owner/repo-%d", index))
	}
	require.NoError(t, session.Initialize())

	var workers sync.WaitGroup
	saveErrors := make(chan error, repositoryCount)
	for index := range repositoryCount {
		workers.Add(1)
		go func() {
			defer workers.Done()

			repository := fmt.Sprintf("owner/repo-%d", index)
			session.MarkStarted(repository)
			session.MarkCompleted(repository)
			saveErrors <- session.Save()
		}()
	}
	workers.Wait()
	close(saveErrors)
	for err := range saveErrors {
		require.NoError(t, err)
	}

	require.NoError(t, session.Save())
	require.Equal(t, 100.0, session.GetProgress())

	loaded := NewSession(&CloneOptions{})
	require.NoError(t, loaded.Load(session.ID))
	require.Equal(t, repositoryCount, loaded.Statistics.TotalRepositories)
	require.Equal(t, repositoryCount, loaded.Statistics.CompletedCount)
	require.Zero(t, loaded.Statistics.FailedCount)
	require.Zero(t, loaded.Statistics.PendingCount)
	require.Zero(t, loaded.Statistics.InProgressCount)
}

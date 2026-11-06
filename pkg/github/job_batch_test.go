// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli/internal/workerpool"
)

func releaseOnce() (chan struct{}, func()) {
	release := make(chan struct{})
	var once sync.Once
	return release, func() {
		once.Do(func() { close(release) })
	}
}

func tinyClonePool(t *testing.T) *workerpool.RepositoryWorkerPool {
	t.Helper()

	pool := workerpool.NewRepositoryWorkerPool(workerpool.RepositoryPoolConfig{
		CloneWorkers:     1,
		UpdateWorkers:    1,
		ConfigWorkers:    1,
		OperationTimeout: 30 * time.Second,
		RetryAttempts:    0,
		RetryDelay:       time.Millisecond,
	})
	if err := pool.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	t.Cleanup(pool.Stop)

	return pool
}

func blockUntilRelease(release <-chan struct{}) func(context.Context, workerpool.RepositoryJob) error {
	return func(ctx context.Context, _ workerpool.RepositoryJob) error {
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func TestRunRepositoryJobsRetriesUntilTheBatchFinishes(t *testing.T) {
	pool := tinyClonePool(t)
	release, unlock := releaseOnce()
	t.Cleanup(unlock)

	const total = 6

	started := make(chan struct{}, total)
	block := func(ctx context.Context, job workerpool.RepositoryJob) error {
		started <- struct{}{}

		return blockUntilRelease(release)(ctx, job)
	}

	jobs := make([]workerpool.RepositoryJob, total)
	for i := range jobs {
		jobs[i] = workerpool.RepositoryJob{
			Repository: fmt.Sprintf("repo-%d", i),
			Operation:  workerpool.OperationClone,
		}
	}

	out := make(chan repositoryBatch, 1)
	go func() {
		out <- runRepositoryJobs(t.Context(), pool, jobs, block)
	}()

	select {
	case <-started:
	case batch := <-out:
		t.Fatalf("returned before a worker started: %v", batch.Err)
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start")
	}

	select {
	case batch := <-out:
		t.Fatalf("returned while workers were blocked: %v", batch.Err)
	case <-time.After(50 * time.Millisecond):
	}

	unlock()

	var batch repositoryBatch
	select {
	case batch = <-out:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for results the queue had already accepted")
	}

	if batch.Err != nil {
		t.Fatalf("runRepositoryJobs: %v", batch.Err)
	}

	if len(batch.Unsent) != 0 {
		t.Fatalf("unsent = %d, want 0", len(batch.Unsent))
	}

	if len(batch.Results) != total {
		t.Fatalf("results = %d, want %d", len(batch.Results), total)
	}

	for _, result := range batch.Results {
		if result.Error != nil {
			t.Fatalf("result %s: %v", result.Job.Repository, result.Error)
		}
	}
}

func TestRunRepositoryJobsReturnsJobsTheQueueNeverAccepted(t *testing.T) {
	pool := tinyClonePool(t)
	release, unlock := releaseOnce()
	t.Cleanup(unlock)

	started := make(chan struct{})
	var startOnce sync.Once
	block := func(ctx context.Context, job workerpool.RepositoryJob) error {
		startOnce.Do(func() { close(started) })
		return blockUntilRelease(release)(ctx, job)
	}
	first := workerpool.RepositoryJob{Repository: "held-0", Operation: workerpool.OperationClone}
	if err := pool.SubmitJob(first, block); err != nil {
		t.Fatalf("SubmitJob: %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start")
	}

	accepted := 1
	sawFull := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !sawFull {
		job := workerpool.RepositoryJob{
			Repository: fmt.Sprintf("held-%d", accepted),
			Operation:  workerpool.OperationClone,
		}
		err := pool.SubmitJob(job, block)
		if err == nil {
			accepted++
			continue
		}

		if !errors.Is(err, workerpool.ErrQueueFull) {
			t.Fatalf("SubmitJob: %v", err)
		}

		sawFull = true
	}

	if !sawFull || accepted < 1 {
		t.Fatalf("queue never filled (accepted=%d)", accepted)
	}

	extra := []workerpool.RepositoryJob{
		{Repository: "extra-0", Operation: workerpool.OperationClone},
		{Repository: "extra-1", Operation: workerpool.OperationClone},
	}
	out := make(chan repositoryBatch, 1)
	go func() {
		out <- runRepositoryJobs(t.Context(), pool, extra, block)
	}()

	var batch repositoryBatch
	select {
	case batch = <-out:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("waited for results of jobs the queue never accepted")
	}

	if !errors.Is(batch.Err, workerpool.ErrQueueFull) {
		t.Fatalf("error = %v, want ErrQueueFull", batch.Err)
	}

	if len(batch.Results) != 0 {
		t.Fatalf("results = %d, want 0", len(batch.Results))
	}

	if len(batch.Unsent) != len(extra) {
		t.Fatalf("unsent = %d, want %d", len(batch.Unsent), len(extra))
	}

	for i, job := range batch.Unsent {
		if job.Repository != extra[i].Repository {
			t.Fatalf("unsent[%d] = %s, want %s", i, job.Repository, extra[i].Repository)
		}

		err := submitFailure(job.Repository, batch.Err)
		if !errors.Is(err, workerpool.ErrQueueFull) {
			t.Fatalf("submitFailure(%s) = %v", job.Repository, err)
		}
	}
}

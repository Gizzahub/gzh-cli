// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package workerpool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func blockUntilRelease(release <-chan struct{}) func(context.Context, RepositoryJob) error {
	return func(ctx context.Context, _ RepositoryJob) error {
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func tinyClonePool(t *testing.T) *RepositoryWorkerPool {
	t.Helper()
	pool := NewRepositoryWorkerPool(RepositoryPoolConfig{
		CloneWorkers:     1,
		UpdateWorkers:    1,
		ConfigWorkers:    1,
		OperationTimeout: 5 * time.Second,
		RetryAttempts:    0,
		RetryDelay:       time.Millisecond,
	})
	if err := pool.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(pool.Stop)
	return pool
}

func TestSubmitStaysNonblockingWhenQueueIsFull(t *testing.T) {
	pool := tinyClonePool(t)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	block := blockUntilRelease(release)

	accepted := 0
	sawFull := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !sawFull {
		job := RepositoryJob{Repository: fmt.Sprintf("repo-%d", accepted), Operation: OperationClone}
		start := time.Now()
		err := pool.SubmitJob(job, block)
		elapsed := time.Since(start)
		if elapsed >= 200*time.Millisecond {
			t.Fatalf("Submit took %s; it must not wait for a free slot", elapsed)
		}
		if err == nil {
			accepted++
			continue
		}
		if !errors.Is(err, ErrQueueFull) {
			t.Fatalf("Submit error = %v, want ErrQueueFull", err)
		}
		sawFull = true
	}
	if !sawFull {
		t.Fatal("queue never filled")
	}
	if accepted < 1 {
		t.Fatal("no job was accepted before the queue filled")
	}
}

func TestSubmitDrainingRetriesAfterAResultFreesASlot(t *testing.T) {
	pool := tinyClonePool(t)
	release := make(chan struct{})
	const total = 6
	started := make(chan struct{}, total)
	block := func(ctx context.Context, _ RepositoryJob) error {
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	jobs := make([]RepositoryJob, total)
	for i := range jobs {
		jobs[i] = RepositoryJob{Repository: fmt.Sprintf("repo-%d", i), Operation: OperationClone}
	}
	got := make([]RepositoryResult, 0, total)
	errCh := make(chan error, 1)
	go func() {
		errCh <- pool.SubmitDraining(t.Context(), jobs, block, DrainConfig{
			OnResult: func(result RepositoryResult) {
				got = append(got, result)
			},
		})
	}()

	select {
	case <-started:
	case err := <-errCh:
		t.Fatalf("returned before a worker started: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start")
	}
	select {
	case err := <-errCh:
		t.Fatalf("rejected the batch before a result freed a slot: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("SubmitDraining: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the drained batch")
	}
	if len(got) != total {
		t.Fatalf("results = %d, want %d", len(got), total)
	}
	for _, result := range got {
		if result.Error != nil {
			t.Fatalf("result %s: %v", result.Job.Repository, result.Error)
		}
	}
}

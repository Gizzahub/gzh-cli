// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/gizzahub/gzh-cli/internal/workerpool"
)

// repositoryBatch is one drained submission.
// Unsent stays empty when the context is canceled. Those jobs may still be
// running, so callers must not record them as failures.
type repositoryBatch struct {
	Results []workerpool.RepositoryResult
	Unsent  []workerpool.RepositoryJob
	Err     error
}

// runRepositoryJobs submits jobs while reading results.
// Submit stays nonblocking. A full queue waits for one result, then retries.
// Jobs the queue never accepts are returned in Unsent with the drain error.
func runRepositoryJobs(
	ctx context.Context,
	pool *workerpool.RepositoryWorkerPool,
	jobs []workerpool.RepositoryJob,
	processFn func(context.Context, workerpool.RepositoryJob) error,
) repositoryBatch {
	results := make([]workerpool.RepositoryResult, 0, len(jobs))
	err := pool.SubmitDraining(ctx, jobs, processFn, workerpool.DrainConfig{
		OnResult: func(result workerpool.RepositoryResult) {
			results = append(results, result)
		},
	})

	batch := repositoryBatch{Results: results, Err: err}
	if err == nil || ctx.Err() != nil {
		return batch
	}

	batch.Unsent = jobsWithoutResults(jobs, results)

	return batch
}

func jobsWithoutResults(jobs []workerpool.RepositoryJob, results []workerpool.RepositoryResult) []workerpool.RepositoryJob {
	seen := make(map[string]int, len(results))
	for _, result := range results {
		seen[result.Job.Repository]++
	}

	unsent := make([]workerpool.RepositoryJob, 0)
	for _, job := range jobs {
		if seen[job.Repository] > 0 {
			seen[job.Repository]--
			continue
		}

		unsent = append(unsent, job)
	}

	return unsent
}

func submitFailure(repository string, drainErr error) error {
	cause := drainErr
	if errors.Is(drainErr, workerpool.ErrQueueFull) {
		cause = workerpool.ErrQueueFull
	}

	return fmt.Errorf("failed to submit job for %s: %w", repository, cause)
}

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package workerpool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Stop이 결과를 아무도 비우지 않는 상태에서도 끝나야 한다.
//
// pkg/github의 processBatch는 ctx.Done()에서 결과 수집 반복문을 중간에
// 버리고 나온다(optimized_synclone.go:364). 그러면 rp.results에 결과가
// 쌓인 채로 defer가 Close -> Stop을 부른다. Ctrl+C 경로가 정확히 이 모양이다.
func TestRepositoryWorkerPoolStopDoesNotHang(t *testing.T) {
	pool := NewRepositoryWorkerPool(RepositoryPoolConfig{
		CloneWorkers:     64,
		UpdateWorkers:    8,
		ConfigWorkers:    4,
		OperationTimeout: time.Minute,
		RetryAttempts:    0,
		RetryDelay:       time.Millisecond,
	})

	if err := pool.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// rp.results 버퍼가 100, clonePool.results가 CloneWorkers*2=128이다.
	// 둘을 다 채우려면 그보다 많은 일이 실제로 끝나야 한다. Submit은
	// 큐가 차면 default로 즉시 거절하므로(pool.go:117) 받아줄 때까지
	// 다시 넣는다.
	noop := func(context.Context, RepositoryJob) error { return nil }

	const wantAccepted = 300

	accepted := 0
	for i := 0; accepted < wantAccepted; i++ {
		job := RepositoryJob{
			Repository: fmt.Sprintf("repo-%d", i),
			Operation:  OperationClone,
		}

		if err := pool.SubmitJob(job, noop); err == nil {
			accepted++
			continue
		}

		time.Sleep(time.Millisecond)
	}

	t.Logf("accepted=%d", accepted)

	// 결과가 쌓여 채널이 다 차게 둔다.
	time.Sleep(300 * time.Millisecond)

	done := make(chan struct{})

	go func() {
		pool.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop()이 끝나지 않는다 -- Ctrl+C에 프로세스가 멈춘다")
	}
}

// Stop을 두 번 불러도 panic이 나면 안 된다.
//
// 안쪽 Pool.Stop은 started 깃발로 막고 있었는데 RepositoryWorkerPool.Stop만
// 맨몸으로 close(rp.results)를 불렀다. 두 번째 호출이 "close of closed
// channel"로 죽는다.
func TestRepositoryWorkerPoolStopIsIdempotent(t *testing.T) {
	pool := NewRepositoryWorkerPool(RepositoryPoolConfig{
		CloneWorkers:     2,
		UpdateWorkers:    2,
		ConfigWorkers:    1,
		OperationTimeout: time.Minute,
		RetryAttempts:    0,
		RetryDelay:       time.Millisecond,
	})

	if err := pool.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	pool.Stop()
	pool.Stop()
}

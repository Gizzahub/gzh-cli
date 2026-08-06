// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package monitor

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// TestRunMonitorFailsFastOnMissingWatchDir는 --watch-dir로 없는 경로를 준
// 경우를 본다.
//
// 예전에는 경고 한 줄을 찍고 "Watching 0 paths for changes"라고 한 뒤
// 감시 고리에 들어가 눌러앉았다. 아무 일도 일어날 수 없는데 일어나기를
// 기다리는 상태다. 쓰는 사람 눈에는 지켜보고 있는 것처럼 보인다.
func TestRunMonitorFailsFastOnMissingWatchDir(t *testing.T) {
	o := defaultMonitorOptions()
	o.watchDir = filepath.Join(t.TempDir(), "없는-디렉토리")

	// 취소되지 않는 맥락을 준다. 여기서 막히면 그것 자체가 고장이므로
	// ctx로 빠져나갈 구멍을 만들어 주지 않는다. 아래 시간 재기가 잡는다.
	done := make(chan error, 1)
	go func() { done <- o.runMonitor(context.Background(), nil, nil) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("없는 디렉토리를 줬는데 오류 없이 끝났다")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("없는 디렉토리를 줬는데 돌아오지 않는다 -- 감시 고리에 들어간 것이다")
	}
}

// TestRunMonitorStopsOnContextCancel은 취소가 실제로 감시 고리를 끊는지
// 본다.
//
// `gz ide monitor`는 Ctrl+C로 멈추지 않았다. apprunner가 cancel()을 부르는
// 것은 맞았지만 root.go가 ExecuteContext가 아니라 Execute를 불러서 cobra가
// context.Background()를 대신 넣었고, 취소가 여기까지 오지 않았다.
// "shutting down gracefully"만 찍고 계속 살아 있어 kill -9로 잡아야 했다.
func TestRunMonitorStopsOnContextCancel(t *testing.T) {
	o := defaultMonitorOptions()
	o.watchDir = t.TempDir() // 실제로 붙을 수 있는 디렉토리라야 고리까지 간다

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- o.runMonitor(ctx, nil, nil) }()

	// 고리에 들어갈 틈을 준 뒤 끊는다.
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("취소로 끝날 때는 오류가 없어야 한다: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("취소했는데 돌아오지 않는다")
	}
}

// TestCommandUsesExecutionContext는 RunE가 어느 맥락을 쓰는지 본다.
//
// 위의 TestRunMonitorStopsOnContextCancel은 runMonitor를 직접 부르면서
// 취소되는 맥락을 손에 쥐여 주므로 고치기 전에도 통과했다. 진짜 고장은
// 한 칸 위에 있었다 -- register.go가 NewCmd에 context.Background()를
// 주고(registry.Provider의 Command()에 ctx 자리가 없다) RunE가 그것을
// 붙잡고 있어서, 실행할 때 들어온 취소되는 맥락이 무시됐다.
//
// 그래서 여기서는 register.go와 똑같이 Background로 명령을 만든 뒤,
// 실행만 취소되는 맥락으로 한다. RunE가 cmd.Context()를 쓰지 않으면
// 이 시험은 돌아오지 않는다.
func TestCommandUsesExecutionContext(t *testing.T) {
	cmd := NewCmd() // register.go가 하는 그대로
	cmd.SetArgs([]string{"--watch-dir", t.TempDir()})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx) }()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("취소로 끝날 때는 오류가 없어야 한다: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("실행 맥락을 취소했는데 돌아오지 않는다 -- RunE가 붙잡아 둔 맥락을 쓰고 있다")
	}
}

// TestGetWatchDirectoriesPrefersExplicitDir는 --watch-dir를 준 경우 자동
// 탐지로 넘어가지 않는지 본다. 자동 탐지는 기계마다 결과가 달라서
// 시험하기 어렵지만, 콕 집어 준 경우는 결정적이다.
func TestGetWatchDirectoriesPrefersExplicitDir(t *testing.T) {
	dir := t.TempDir()

	o := defaultMonitorOptions()
	o.watchDir = dir

	dirs := o.getWatchDirectories()
	if len(dirs) != 1 || dirs[0] != dir {
		t.Fatalf("준 디렉토리 하나만 나와야 하는데 %v가 나왔다", dirs)
	}
}

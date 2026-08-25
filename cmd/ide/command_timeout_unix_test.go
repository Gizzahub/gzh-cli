// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package ide

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExecutableVersion_TerminatesChildProcessTree(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	script := filepath.Join(dir, "spawn-child.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nsleep 60 &\nchild=$!\nprintf '%s\\n' \"$child\" > \"$1\"\nwait \"$child\"\n"), 0o755))

	detector := NewIDEDetector()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan string, 1)
	go func() {
		result <- detector.getExecutableVersion(ctx, script, pidFile)
	}()
	var pid int
	require.Eventually(t, func() bool {
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			return false
		}
		pid, err = strconv.Atoi(strings.TrimSpace(string(pidData)))
		return err == nil && pid > 0
	}, 30*time.Second, 10*time.Millisecond, "timed out waiting for child process PID")

	start := time.Now()
	cancel()
	var version string
	select {
	case version = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("timed-out exec did not return")
	}
	require.Less(t, time.Since(start), 5*time.Second)
	assert.Equal(t, "unknown", version)

	require.Eventually(t, func() bool {
		err := syscall.Kill(pid, syscall.Signal(0))
		return errors.Is(err, syscall.ESRCH)
	}, 2*time.Second, 20*time.Millisecond, "timed-out command left child process %d running", pid)
}

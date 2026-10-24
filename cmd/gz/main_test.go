// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // 실제 main의 프로세스 종료 코드를 검증한다.
package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCLIExitStatus(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
		text string
	}{
		{"unknown root", []string{"does-not-exist"}, 1, "unknown command"},
		{"unknown group child", []string{"profile", "start"}, 1, "unknown command"},
		{"unknown flag", []string{"profile", "cpu", "--format", "flame"}, 1, "unknown flag"},
		{"root help", []string{"--help"}, 0, "Usage:"},
		{"group help", []string{"profile", "--help"}, 0, "Usage:"},
		{"bare group", []string{"profile"}, 0, "Usage:"},
		{"leaf help", []string{"profile", "cpu", "--help"}, 0, "Usage:"},
		{"wrapper alias help", []string{"q", "--help"}, 0, "Usage:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			args := append([]string{"-test.run=^TestCLIProcess$", "--"}, tc.args...)
			process := exec.CommandContext(ctx, os.Args[0], args...)
			process.Env = append(os.Environ(), "GZH_CLI_PROCESS_TEST=1", "GZH_DEBUG_SHELL=0", "HOME="+t.TempDir())
			output, err := process.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("CLI timed out: %s", output)
			}
			code := 0
			if err != nil {
				var exit *exec.ExitError
				if errors.As(err, &exit) {
					code = exit.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if code != tc.code || !strings.Contains(string(output), tc.text) {
				t.Fatalf("exit=%d want=%d, output=%s", code, tc.code, output)
			}
		})
	}
}

func TestCLIProcess(t *testing.T) {
	if os.Getenv("GZH_CLI_PROCESS_TEST") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"gz"}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing CLI arguments delimiter")
}

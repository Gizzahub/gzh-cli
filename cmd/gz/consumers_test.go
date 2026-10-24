// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // 실제 main을 실행하는 기존 프로세스 helper를 공유한다.
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runConsumerCLI(t *testing.T, directory string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)...)
	process.Dir = directory
	process.Env = append(os.Environ(), "GZH_CLI_PROCESS_TEST=1", "GZH_DEBUG_SHELL=0",
		"HOME="+directory, "XDG_CONFIG_HOME="+directory, "XDG_CACHE_HOME="+directory)
	output, err := process.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("consumer CLI timed out: %s", output)
	}
	return string(output), err
}

func TestReleasedQualityAnalyze(t *testing.T) {
	directory := t.TempDir()
	source := []byte("package fixture\n")
	path := filepath.Join(directory, "fixture.go")
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := runConsumerCLI(t, directory, "quality", "analyze")
	canonical, pathErr := filepath.EvalSymlinks(directory)
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if err != nil || !strings.Contains(output, "Go: 1") || !strings.Contains(output, canonical) {
		t.Fatalf("analysis failed: %v\n%s", err, output)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(source) {
		t.Fatalf("analysis changed source: %v", err)
	}
}

func TestReleasedShellforgeValidate(t *testing.T) {
	directory := t.TempDir()
	manifest := filepath.Join(directory, "manifest.yaml")
	if err := os.WriteFile(manifest, []byte("version: \"2\"\nshell:\n  type: zsh\nmodules:\n  - name: fixture\n    file: fixture.sh\n    target: zshrc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(directory, "fixture.sh")
	if err := os.WriteFile(module, []byte("export FIXTURE=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := runConsumerCLI(t, directory, "shellforge", "validate", "--manifest", manifest, "--config-dir", directory)
	if err != nil || !strings.Contains(output, "Validation successful!") {
		t.Fatalf("valid fixture rejected: %v\n%s", err, output)
	}
	if err := os.Remove(module); err != nil {
		t.Fatal(err)
	}
	output, err = runConsumerCLI(t, directory, "shellforge", "validate", "--manifest", manifest, "--config-dir", directory)
	if err == nil || !strings.Contains(output, "fixture.sh") {
		t.Fatalf("missing module accepted: %v\n%s", err, output)
	}
}

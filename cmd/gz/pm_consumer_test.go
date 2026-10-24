// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build pm_external

//nolint:testpackage // 실제 main의 외부 PM 경로를 검증한다.
package main

import (
	"strings"
	"testing"
)

func TestReleasedPMRejectsInvalidStrategy(t *testing.T) {
	output, err := runConsumerCLI(t, t.TempDir(), "pm", "update", "--managers", "mise", "--strategy", "invalid", "--dry-run")
	if err == nil || !strings.Contains(output, "strategy") || strings.Contains(output, "unknown flag") {
		t.Fatalf("external PM validation failed: %v\n%s", err, output)
	}
}

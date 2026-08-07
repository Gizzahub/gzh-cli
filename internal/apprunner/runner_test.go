// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package apprunner

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: 0},
		{name: "generic", err: errors.New("boom"), want: 1},
		{name: "canceled", err: context.Canceled, want: ExitInterrupted},
		{name: "wrapped canceled", err: fmt.Errorf("application execution failed: %w", context.Canceled), want: ExitInterrupted},
		{name: "deadline not interrupt", err: context.DeadlineExceeded, want: 1},
		{name: "synclone style", err: fmt.Errorf("operation canceled: %w", context.Canceled), want: ExitInterrupted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

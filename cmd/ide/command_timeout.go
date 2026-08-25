// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package ide

import (
	"context"
	"os/exec"
	"time"
)

const commandPipeWaitDelay = 500 * time.Millisecond

// commandContext creates a command whose cancellation terminates descendants
// and bounds waits on inherited stdout/stderr pipes.
func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	configureProcessTree(cmd)
	cmd.Cancel = func() error {
		return killProcessTree(cmd)
	}
	cmd.WaitDelay = commandPipeWaitDelay
	return cmd
}

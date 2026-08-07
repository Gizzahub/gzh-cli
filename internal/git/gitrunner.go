// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// GitRunner runs git with the given args. dir may be empty for no working dir.
// Implementations must not dial the network unless they intentionally wrap a
// real git binary (ExecGitRunner).
type GitRunner interface {
	Run(ctx context.Context, dir string, args ...string) (output []byte, err error)
}

// ExecGitRunner runs the real git binary via exec.CommandContext.
type ExecGitRunner struct{}

// Run executes git with the provided args, optionally in dir.
func (ExecGitRunner) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.CombinedOutput()
}

// GitCall records one GitRunner.Run invocation for tests.
type GitCall struct {
	Dir  string
	Args []string
}

// RecordingGitRunner records every Run call and optionally simulates clone by
// creating the target directory so path-existence asserts pass without network.
type RecordingGitRunner struct {
	mu    sync.Mutex
	Calls []GitCall
	// Err, when set, is returned from every Run (after recording the call).
	Err error
	// FailClone, when true, makes clone operations fail while other commands succeed.
	FailClone bool
}

// Run records the call. On successful clone-like invocations it creates the
// destination path so callers that check for a directory after clone succeed.
func (r *RecordingGitRunner) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := make([]string, len(args))
	copy(copied, args)
	r.Calls = append(r.Calls, GitCall{Dir: dir, Args: copied})

	if r.Err != nil {
		return nil, r.Err
	}

	if len(args) == 0 {
		return []byte(""), nil
	}

	if args[0] == "clone" {
		if r.FailClone {
			return []byte("fatal: clone failed"), fmt.Errorf("clone failed: permission denied")
		}
		// git clone [options...] <url> <path>
		target := args[len(args)-1]
		if target != "" && !strings.Contains(target, "://") && !strings.Contains(target, "@") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
			// Minimal content so the path looks like a checkout.
			_ = os.WriteFile(filepath.Join(target, "README.md"), []byte("# mock clone\n"), 0o644)
		}
		return []byte("Cloning into '" + target + "'...\n"), nil
	}

	// ls-remote and other probes: empty success (no remote content).
	return []byte(""), nil
}

// CloneArgs returns args from calls whose first arg is "clone".
func (r *RecordingGitRunner) CloneArgs() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out [][]string
	for _, c := range r.Calls {
		if len(c.Args) > 0 && c.Args[0] == "clone" {
			out = append(out, append([]string{}, c.Args...))
		}
	}
	return out
}

// Reset clears recorded calls and failure flags.
func (r *RecordingGitRunner) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Calls = nil
	r.Err = nil
	r.FailClone = false
}

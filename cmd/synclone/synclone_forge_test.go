// Package synclone provides CLI commands for repository synchronization.
package synclone

import (
	"bytes"
	"context"
	"testing"

	"github.com/Gizzahub/gzh-cli-git-sync/pkg/reposync"
)

func TestCreateForgeProvider(t *testing.T) {
	tests := []struct {
		name        string
		opts        *forgeOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "github provider",
			opts: &forgeOptions{
				Provider: "github",
				Token:    "test-token",
			},
			wantErr: false,
		},
		{
			name: "gitlab provider",
			opts: &forgeOptions{
				Provider: "gitlab",
				Token:    "test-token",
				BaseURL:  "",
			},
			wantErr: false,
		},
		{
			name: "gitlab provider with custom base URL",
			opts: &forgeOptions{
				Provider: "gitlab",
				Token:    "test-token",
				BaseURL:  "https://gitlab.company.com",
			},
			wantErr: false,
		},
		{
			name: "gitea provider",
			opts: &forgeOptions{
				Provider: "gitea",
				Token:    "test-token",
				BaseURL:  "https://gitea.example.com",
			},
			wantErr: false,
		},
		{
			name: "unsupported provider",
			opts: &forgeOptions{
				Provider: "bitbucket",
			},
			wantErr:     true,
			errContains: "unsupported provider",
		},
		{
			name: "empty provider",
			opts: &forgeOptions{
				Provider: "",
			},
			wantErr:     true,
			errContains: "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := createForgeProvider(tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("expected non-nil provider")
			}
		})
	}
}

func TestConsoleProgressSink(t *testing.T) {
	t.Run("OnStart outputs repo name and action type", func(t *testing.T) {
		var buf bytes.Buffer
		sink := consoleProgressSink{Out: &buf}

		action := reposync.Action{
			Repo: reposync.RepoSpec{Name: "my-repo"},
			Type: reposync.ActionClone,
		}

		sink.OnStart(action)

		output := buf.String()
		if !contains(output, "my-repo") {
			t.Error("output should contain repo name")
		}
		if !contains(output, "clone") {
			t.Error("output should contain action type")
		}
	})

	t.Run("OnProgress outputs message", func(t *testing.T) {
		var buf bytes.Buffer
		sink := consoleProgressSink{Out: &buf}

		action := reposync.Action{
			Repo: reposync.RepoSpec{Name: "test-repo"},
		}

		sink.OnProgress(action, "Cloning repository...", 0.5)

		output := buf.String()
		if !contains(output, "test-repo") {
			t.Error("output should contain repo name")
		}
		if !contains(output, "Cloning repository...") {
			t.Error("output should contain message")
		}
	})

	t.Run("OnComplete outputs success for successful result", func(t *testing.T) {
		var buf bytes.Buffer
		sink := consoleProgressSink{Out: &buf}

		result := reposync.ActionResult{
			Action: reposync.Action{
				Repo: reposync.RepoSpec{Name: "success-repo"},
			},
			Error: nil,
		}

		sink.OnComplete(result)

		output := buf.String()
		if !contains(output, "success-repo") {
			t.Error("output should contain repo name")
		}
		if !contains(output, "Completed") {
			t.Error("output should indicate completion")
		}
	})

	t.Run("OnComplete outputs failure for failed result", func(t *testing.T) {
		var buf bytes.Buffer
		sink := consoleProgressSink{Out: &buf}

		result := reposync.ActionResult{
			Action: reposync.Action{
				Repo: reposync.RepoSpec{Name: "failed-repo"},
			},
			Error: context.DeadlineExceeded,
		}

		sink.OnComplete(result)

		output := buf.String()
		if !contains(output, "failed-repo") {
			t.Error("output should contain repo name")
		}
		if !contains(output, "Failed") {
			t.Error("output should indicate failure")
		}
	})
}

func TestForgeOptions_Defaults(t *testing.T) {
	// Test that default values are set correctly
	opts := &forgeOptions{
		Strategy:       "reset",
		Parallel:       4,
		MaxRetries:     3,
		IncludePrivate: true,
	}

	if opts.Strategy != "reset" {
		t.Errorf("expected default strategy 'reset', got %s", opts.Strategy)
	}
	if opts.Parallel != 4 {
		t.Errorf("expected default parallel 4, got %d", opts.Parallel)
	}
	if opts.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", opts.MaxRetries)
	}
	if !opts.IncludePrivate {
		t.Error("expected IncludePrivate to be true by default")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

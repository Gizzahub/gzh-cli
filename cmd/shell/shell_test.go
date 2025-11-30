//nolint:testpackage // White-box testing needed for internal function access
package shell

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/pkg/gzhclient"
)

func TestShellCmdCreation(t *testing.T) {
	assert.Equal(t, "shell", ShellCmd.Use)
	assert.Equal(t, "Start interactive debugging shell (REPL)", ShellCmd.Short)
	assert.NotEmpty(t, ShellCmd.Long)
	assert.NotNil(t, ShellCmd.Run)

	// Check that flags are properly set up
	timeoutFlag := ShellCmd.Flags().Lookup("timeout")
	assert.NotNil(t, timeoutFlag)
	assert.Equal(t, "0s", timeoutFlag.DefValue)

	quietFlag := ShellCmd.Flags().Lookup("quiet")
	assert.NotNil(t, quietFlag)
	assert.Equal(t, "false", quietFlag.DefValue)

	noHistoryFlag := ShellCmd.Flags().Lookup("no-history")
	assert.NotNil(t, noHistoryFlag)
	assert.Equal(t, "false", noHistoryFlag.DefValue)
}

func TestNewShell(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	assert.NotNil(t, shell)
	assert.Equal(t, client, shell.client)
	assert.True(t, shell.running)
	assert.NotNil(t, shell.ctx)
	assert.NotNil(t, shell.cancel)
	assert.NotNil(t, shell.commands)
	assert.Empty(t, shell.history)

	// Verify built-in commands are registered
	expectedCommands := []string{
		"help", "exit", "quit", "status", "memory", "plugins",
		"config", "metrics", "trace", "profile", "history",
		"clear", "context", "logs",
	}

	for _, cmdName := range expectedCommands {
		_, exists := shell.commands[cmdName]
		assert.True(t, exists, "Command %s should be registered", cmdName)
	}
}

func TestShellStop(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)
	assert.True(t, shell.running)

	shell.Stop()
	assert.False(t, shell.running)

	// Context should be canceled
	select {
	case <-shell.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be canceled after Stop()")
	}
}

func TestAddToHistory(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	// Add first command
	shell.addToHistory("help")
	assert.Len(t, shell.history, 1)
	assert.Equal(t, "help", shell.history[0])

	// Add different command
	shell.addToHistory("status")
	assert.Len(t, shell.history, 2)
	assert.Equal(t, "status", shell.history[1])

	// Add duplicate consecutive command (should be ignored)
	shell.addToHistory("status")
	assert.Len(t, shell.history, 2)

	// Add different command after duplicate
	shell.addToHistory("exit")
	assert.Len(t, shell.history, 3)
}

func TestExecuteCommand(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty command",
			input:       "",
			expectError: false,
		},
		{
			name:        "help command",
			input:       "help",
			expectError: false,
		},
		{
			name:        "help with args",
			input:       "help status",
			expectError: false,
		},
		{
			name:        "unknown command",
			input:       "unknown",
			expectError: true,
			errorMsg:    "unknown command: unknown",
		},
		{
			name:        "exit command",
			input:       "exit",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shell.executeCommand(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShellCommands(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	t.Run("help command", func(t *testing.T) {
		err := handleHelp(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("help with specific command", func(t *testing.T) {
		err := handleHelp(shell, []string{"status"})
		assert.NoError(t, err)
	})

	t.Run("help with unknown command", func(t *testing.T) {
		err := handleHelp(shell, []string{"nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command: nonexistent")
	})

	t.Run("exit command", func(t *testing.T) {
		shell.running = true // Reset state
		err := handleExit(shell, []string{})
		assert.NoError(t, err)
		assert.False(t, shell.running)
	})

	t.Run("status command", func(t *testing.T) {
		err := handleStatus(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("status command with json", func(t *testing.T) {
		err := handleStatus(shell, []string{"--json"})
		assert.NoError(t, err)
	})

	t.Run("memory command", func(t *testing.T) {
		err := handleMemory(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("memory command with json", func(t *testing.T) {
		err := handleMemory(shell, []string{"--json"})
		assert.NoError(t, err)
	})

	t.Run("memory command with gc", func(t *testing.T) {
		err := handleMemory(shell, []string{"--gc"})
		assert.NoError(t, err)
	})

	t.Run("plugins command", func(t *testing.T) {
		err := handlePlugins(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("config command", func(t *testing.T) {
		err := handleConfig(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("clear command", func(t *testing.T) {
		err := handleClear(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("context command", func(t *testing.T) {
		err := handleContext(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("context command with json", func(t *testing.T) {
		err := handleContext(shell, []string{"--json"})
		assert.NoError(t, err)
	})

	t.Run("logs command", func(t *testing.T) {
		err := handleLogs(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("logs command with count", func(t *testing.T) {
		err := handleLogs(shell, []string{"--count", "5"})
		assert.NoError(t, err)
	})

	t.Run("logs command with level", func(t *testing.T) {
		err := handleLogs(shell, []string{"--level", "error"})
		assert.NoError(t, err)
	})
}

func TestHistoryCommand(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	// Add some history
	shell.addToHistory("help")
	shell.addToHistory("status")
	shell.addToHistory("memory")

	t.Run("show history", func(t *testing.T) {
		err := handleHistory(shell, []string{})
		assert.NoError(t, err)
	})

	t.Run("show history with count", func(t *testing.T) {
		err := handleHistory(shell, []string{"--count", "2"})
		assert.NoError(t, err)
	})

	t.Run("clear history", func(t *testing.T) {
		assert.Len(t, shell.history, 3)
		err := handleHistory(shell, []string{"--clear"})
		assert.NoError(t, err)
		assert.Empty(t, shell.history)
	})

	t.Run("empty history", func(t *testing.T) {
		err := handleHistory(shell, []string{})
		assert.NoError(t, err)
	})
}

func TestTraceCommand(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	tests := []struct {
		name string
		args []string
	}{
		{"start", []string{"start"}},
		{"stop", []string{"stop"}},
		{"status", []string{"status"}},
		{"default status", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleTrace(shell, tt.args)
			assert.NoError(t, err)
		})
	}

	t.Run("unknown trace command", func(t *testing.T) {
		err := handleTrace(shell, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown trace command: unknown")
	})
}

func TestProfileCommand(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	tests := []struct {
		name string
		args []string
	}{
		{"start", []string{"start"}},
		{"stop", []string{"stop"}},
		{"status", []string{"status"}},
		{"default status", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleProfile(shell, tt.args)
			assert.NoError(t, err)
		})
	}

	t.Run("unknown profile command", func(t *testing.T) {
		err := handleProfile(shell, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown profile command: unknown")
	})
}

func TestCompletion(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	t.Run("completeHelp", func(t *testing.T) {
		completions := completeHelp(shell, "st")
		assert.Contains(t, completions, "status")
	})

	t.Run("completeHelp no match", func(t *testing.T) {
		completions := completeHelp(shell, "xyz")
		assert.Empty(t, completions)
	})

	t.Run("completePlugins", func(t *testing.T) {
		completions := completePlugins(shell, "l")
		assert.Contains(t, completions, "list")
	})

	t.Run("completePlugins all", func(t *testing.T) {
		completions := completePlugins(shell, "")
		assert.Contains(t, completions, "list")
		assert.Contains(t, completions, "exec")
	})
}

func TestShellCommandParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedCmd string
	}{
		{"simple command", "help", "help"},
		{"command with args", "help status", "help"},
		{"multiple spaces", "   help   status   ", "help"},
		{"tabs and spaces", "\t  help\t status  \t", "help"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.input)
			if len(parts) > 0 {
				assert.Equal(t, tt.expectedCmd, parts[0])
			}
		})
	}
}

func TestHistoryLimit(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	// Add more than 100 commands to test history limit
	for i := 0; i < 110; i++ {
		shell.addToHistory("command" + string(rune(i)))
	}
	_ = shell // Use shell to avoid "declared and not used" error

	// Should only keep last 100
	assert.Len(t, shell.history, 100)
	assert.Equal(t, "command"+string(rune(10)), shell.history[0]) // Should start from command10
}

func TestShellCommandStructures(t *testing.T) {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	shell := NewShell(client)

	// Test that all commands have required fields
	for name, cmd := range shell.commands {
		t.Run("command "+name, func(t *testing.T) {
			assert.NotEmpty(t, cmd.Name)
			assert.NotEmpty(t, cmd.Description)
			assert.NotEmpty(t, cmd.Usage)
			assert.NotNil(t, cmd.Handler)
			// Completer is optional, so we don't check it
		})
	}
}

//nolint:testpackage // White-box testing needed for internal function access
package profile

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestNewProfileCmd(t *testing.T) {
	cmd := NewProfileCmd(app.NewTestAppContext())

	assert.Equal(t, "profile", cmd.Use)
	assert.Equal(t, "Performance profiling using standard Go pprof", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check that the long description contains expected content
	assert.Contains(t, cmd.Long, "Available commands:")
	assert.Contains(t, cmd.Long, "server")
	assert.Contains(t, cmd.Long, "cpu")
	assert.Contains(t, cmd.Long, "memory")
	assert.Contains(t, cmd.Long, "stats")

	// Verify subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)

	// Verify subcommands exist
	var serverCmd, cpuCmd, memoryCmd, statsCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "server":
			serverCmd = true
		case "cpu":
			cpuCmd = true
		case "memory":
			memoryCmd = true
		case "stats":
			statsCmd = true
		}
	}

	assert.True(t, serverCmd, "server subcommand should exist")
	assert.True(t, cpuCmd, "cpu subcommand should exist")
	assert.True(t, memoryCmd, "memory subcommand should exist")
	assert.True(t, statsCmd, "stats subcommand should exist")
}

func TestNewSimpleServerCmd(t *testing.T) {
	cmd := newSimpleServerCmd()

	assert.Equal(t, "server", cmd.Use)
	assert.Equal(t, "Start pprof HTTP server", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check that the long description contains expected endpoints
	assert.Contains(t, cmd.Long, "/debug/pprof/")
	assert.Contains(t, cmd.Long, "/debug/pprof/profile")
	assert.Contains(t, cmd.Long, "/debug/pprof/heap")
	assert.Contains(t, cmd.Long, "/debug/pprof/goroutine")

	// Check flags
	portFlag := cmd.Flags().Lookup("port")
	assert.NotNil(t, portFlag)
	assert.Equal(t, "6060", portFlag.DefValue)
}

func TestNewSimpleCPUCmd(t *testing.T) {
	cmd := newSimpleCPUCmd()

	assert.Equal(t, "cpu", cmd.Use)
	assert.Equal(t, "Collect CPU profile", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	durationFlag := cmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)
	assert.Equal(t, "30s", durationFlag.DefValue)
}

func TestNewSimpleMemoryCmd(t *testing.T) {
	cmd := newSimpleMemoryCmd()

	assert.Equal(t, "memory", cmd.Use)
	assert.Equal(t, "Collect memory profile", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Memory command should not have duration flag (it's a snapshot)
	durationFlag := cmd.Flags().Lookup("duration")
	assert.Nil(t, durationFlag)
}

func TestNewSimpleStatsCmd(t *testing.T) {
	cmd := newSimpleStatsCmd()

	assert.Equal(t, "stats", cmd.Use)
	assert.Equal(t, "Show runtime statistics", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Stats command should not have any flags
	assert.False(t, cmd.Flags().HasFlags())
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
		{"exact kilobyte", 1024, "1.0 KB"},
		{"exact megabyte", 1024 * 1024, "1.0 MB"},
		{"large value", 2048576, "2.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProfileCmdStructure(t *testing.T) {
	cmd := NewProfileCmd(app.NewTestAppContext())

	// Test that the command has proper structure
	assert.NotEmpty(t, cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test that examples are included in Long description
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz profile server")
	assert.Contains(t, cmd.Long, "gz profile cpu")
	assert.Contains(t, cmd.Long, "gz profile memory")
	assert.Contains(t, cmd.Long, "gz profile stats")
}

func TestProfileSubcommandDetails(t *testing.T) {
	cmd := NewProfileCmd(app.NewTestAppContext())
	subcommands := cmd.Commands()

	for _, subcmd := range subcommands {
		t.Run("subcommand "+subcmd.Use, func(t *testing.T) {
			assert.NotEmpty(t, subcmd.Use)
			assert.NotEmpty(t, subcmd.Short)
			assert.NotEmpty(t, subcmd.Long)
			assert.NotNil(t, subcmd.RunE)
		})
	}
}

func TestFormatBytesEdgeCases(t *testing.T) {
	// Test edge cases for formatBytes function
	testCases := []struct {
		input    uint64
		contains string // What the output should contain
	}{
		{1, "B"},                           // Should contain bytes unit
		{1023, "B"},                        // Just under 1KB
		{1025, "KB"},                       // Just over 1KB
		{uint64(1024 * 1024 * 1024), "GB"}, // Exactly 1GB
	}

	for _, tc := range testCases {
		result := formatBytes(tc.input)
		assert.Contains(t, result, tc.contains, "formatBytes(%d) = %s should contain %s", tc.input, result, tc.contains)
	}
}

func TestCommandCreationConsistency(t *testing.T) {
	// Test that all command creation functions return valid commands
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"server", newSimpleServerCmd},
		{"cpu", newSimpleCPUCmd},
		{"memory", newSimpleMemoryCmd},
		{"stats", newSimpleStatsCmd},
	}

	for _, tc := range commands {
		t.Run("command creation "+tc.name, func(t *testing.T) {
			cmd := tc.cmd()
			assert.NotNil(t, cmd)
			assert.NotEmpty(t, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)
		})
	}
}

func TestProfileCommandHelpText(t *testing.T) {
	cmd := NewProfileCmd(app.NewTestAppContext())

	// Verify help text mentions key profiling concepts
	longDesc := cmd.Long
	assert.Contains(t, longDesc, "performance profiling")
	assert.Contains(t, longDesc, "pprof")
	assert.Contains(t, longDesc, "CPU profile")
	assert.Contains(t, longDesc, "memory profile")
	assert.Contains(t, longDesc, "runtime statistics")

	// Verify examples show proper usage
	assert.Contains(t, longDesc, "--port")
	assert.Contains(t, longDesc, "--duration")
}

func TestSubcommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		cmd          func() *cobra.Command
		expectedFlag string
		defaultValue string
	}{
		{"server", newSimpleServerCmd, "port", "6060"},
		{"cpu", newSimpleCPUCmd, "duration", "30s"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" flags", func(t *testing.T) {
			cmd := tt.cmd()
			flag := cmd.Flags().Lookup(tt.expectedFlag)
			assert.NotNil(t, flag, "Flag %s should exist", tt.expectedFlag)
			assert.Equal(t, tt.defaultValue, flag.DefValue, "Default value for %s should be %s", tt.expectedFlag, tt.defaultValue)
		})
	}
}

func TestFormatBytesLargeNumbers(t *testing.T) {
	// Test very large numbers to ensure no overflow
	largeNumber := uint64(1024 * 1024 * 1024 * 1024) // 1TB
	result := formatBytes(largeNumber)
	assert.Contains(t, result, "TB")
	assert.NotContains(t, result, "NaN")
	assert.NotContains(t, result, "Inf")
}

func TestProfileCommandUsagePatterns(t *testing.T) {
	cmd := NewProfileCmd(app.NewTestAppContext())

	// Test that command descriptions follow consistent patterns
	for _, subcmd := range cmd.Commands() {
		t.Run("usage pattern "+subcmd.Use, func(t *testing.T) {
			// All commands should have proper capitalization in Short description
			assert.True(t, len(subcmd.Short) > 0, "Short description should not be empty")

			// Long description should be more detailed than short
			assert.True(t, len(subcmd.Long) > len(subcmd.Short), "Long description should be longer than short")
		})
	}
}

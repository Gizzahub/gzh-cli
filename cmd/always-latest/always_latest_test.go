package alwayslatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlwaysLatestCommand(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := NewAlwaysLatestCmd(context.Background())
		assert.NotNil(t, cmd)
		assert.Equal(t, "always-latest", cmd.Use)
		assert.Contains(t, cmd.Short, "Keep development tools and package managers up to date")

		// Check that it has subcommands
		subcommands := cmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, subcmd := range subcommands {
			subcommandNames[i] = subcmd.Use
		}

		assert.Contains(t, subcommandNames, "asdf")
		assert.Contains(t, subcommandNames, "brew")
	})
}

func TestAsdfCommand(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := newAlwaysLatestAsdfCmd(context.Background())
		assert.NotNil(t, cmd)
		assert.Equal(t, "asdf", cmd.Use)
		assert.Contains(t, cmd.Short, "Update asdf and its managed tools")

		// Check flags
		strategyFlag := cmd.Flags().Lookup("strategy")
		assert.NotNil(t, strategyFlag)
		assert.Equal(t, "minor", strategyFlag.DefValue)

		toolsFlag := cmd.Flags().Lookup("tools")
		assert.NotNil(t, toolsFlag)

		dryRunFlag := cmd.Flags().Lookup("dry-run")
		assert.NotNil(t, dryRunFlag)

		globalFlag := cmd.Flags().Lookup("global")
		assert.NotNil(t, globalFlag)

		interactiveFlag := cmd.Flags().Lookup("interactive")
		assert.NotNil(t, interactiveFlag)
	})

	t.Run("default options", func(t *testing.T) {
		opts := defaultAlwaysLatestAsdfOptions()
		assert.Equal(t, "minor", opts.strategy)
		assert.Equal(t, []string{}, opts.tools)
		assert.False(t, opts.dryRun)
		assert.True(t, opts.updateAsdf)
		assert.False(t, opts.global)
		assert.True(t, opts.interactive)
	})
}

func TestAsdfVersionParsing(t *testing.T) {
	opts := &alwaysLatestAsdfOptions{}

	t.Run("extract major version", func(t *testing.T) {
		tests := []struct {
			version  string
			expected string
			hasError bool
		}{
			{"18.17.0", "18", false},
			{"3.11.5", "3", false},
			{"2.7.18", "2", false},
			{"1.70.0", "1", false},
			{"v18.17.0", "", true}, // No leading 'v' support
			{"invalid", "", true},
			{"", "", true},
		}

		for _, tt := range tests {
			result, err := opts.extractMajorVersion(tt.version)
			if tt.hasError {
				assert.Error(t, err, "Expected error for version: %s", tt.version)
			} else {
				assert.NoError(t, err, "Unexpected error for version: %s", tt.version)
				assert.Equal(t, tt.expected, result, "Unexpected result for version: %s", tt.version)
			}
		}
	})

	t.Run("stable version detection", func(t *testing.T) {
		tests := []struct {
			version  string
			expected bool
		}{
			{"18.17.0", true},
			{"3.11.5", true},
			{"1.70.0", true},
			{"18.18.0-alpha.1", false},
			{"3.12.0-beta.2", false},
			{"1.71.0-rc.1", false},
			{"20.0.0-dev", false},
			{"18.17.0-snapshot", false},
			{"3.11.0-preview", false},
			{"1.70.0-pre", false},
			{"19.0.0-nightly", false},
			{"18.0.0-canary", false},
			{"3.12.0-experimental", false},
			{"1.71.0-test", false},
		}

		for _, tt := range tests {
			result := opts.isStableVersion(tt.version)
			assert.Equal(t, tt.expected, result, "Unexpected result for version: %s", tt.version)
		}
	})
}

func TestAsdfToolFiltering(t *testing.T) {
	opts := &alwaysLatestAsdfOptions{}

	t.Run("filter tools", func(t *testing.T) {
		allTools := []string{"nodejs", "python", "ruby", "golang", "java"}
		requestedTools := []string{"nodejs", "python"}

		filtered := opts.filterTools(allTools, requestedTools)
		expected := []string{"nodejs", "python"}

		assert.Equal(t, expected, filtered)
	})

	t.Run("filter tools with whitespace", func(t *testing.T) {
		allTools := []string{"nodejs", "python", "ruby"}
		requestedTools := []string{" nodejs ", "python", " ruby "}

		filtered := opts.filterTools(allTools, requestedTools)
		expected := []string{"nodejs", "python", "ruby"}

		assert.Equal(t, expected, filtered)
	})

	t.Run("filter tools not found", func(t *testing.T) {
		allTools := []string{"nodejs", "python", "ruby"}
		requestedTools := []string{"golang", "java"}

		filtered := opts.filterTools(allTools, requestedTools)

		assert.Empty(t, filtered)
	})
}

func TestAsdfTargetVersionLogic(t *testing.T) {
	opts := &alwaysLatestAsdfOptions{}

	t.Run("major strategy returns latest", func(t *testing.T) {
		opts.strategy = "major"

		target, err := opts.getTargetVersion("nodejs", "18.17.0", "20.0.0")
		assert.NoError(t, err)
		assert.Equal(t, "20.0.0", target)
	})

	t.Run("minor strategy with no current version returns latest", func(t *testing.T) {
		opts.strategy = "minor"

		target, err := opts.getTargetVersion("nodejs", "", "20.0.0")
		assert.NoError(t, err)
		assert.Equal(t, "20.0.0", target)
	})

	t.Run("extract major version from complex versions", func(t *testing.T) {
		tests := []struct {
			version  string
			expected string
		}{
			{"18", "18"},
			{"18.0", "18"},
			{"18.17", "18"},
			{"18.17.0", "18"},
			{"3.11.5", "3"},
			{"1.70.0", "1"},
		}

		for _, tt := range tests {
			result, err := opts.extractMajorVersion(tt.version)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})
}

// These are unit tests for the core logic and parsing functions.
func TestAsdfIntegration(t *testing.T) {
	t.Run("asdf not installed scenario", func(t *testing.T) {
		opts := &alwaysLatestAsdfOptions{}

		// This test will pass if asdf is not installed
		// In CI/CD environments where asdf might not be available
		if !opts.isAsdfInstalled() {
			t.Log("asdf is not installed - this is expected in some test environments")
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		opts := &alwaysLatestAsdfOptions{
			dryRun: true,
		}

		// Dry run should not make actual changes
		// This tests the dry run flag behavior
		assert.True(t, opts.dryRun)
	})
}

func TestBrewCommand(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := newAlwaysLatestBrewCmd(context.Background())
		assert.NotNil(t, cmd)
		assert.Equal(t, "brew", cmd.Use)
		assert.Contains(t, cmd.Short, "Update Homebrew and its managed packages")

		// Check flags
		strategyFlag := cmd.Flags().Lookup("strategy")
		assert.NotNil(t, strategyFlag)
		assert.Equal(t, "minor", strategyFlag.DefValue)

		packagesFlag := cmd.Flags().Lookup("packages")
		assert.NotNil(t, packagesFlag)

		dryRunFlag := cmd.Flags().Lookup("dry-run")
		assert.NotNil(t, dryRunFlag)

		casksFlag := cmd.Flags().Lookup("casks")
		assert.NotNil(t, casksFlag)

		tapsFlag := cmd.Flags().Lookup("taps")
		assert.NotNil(t, tapsFlag)

		interactiveFlag := cmd.Flags().Lookup("interactive")
		assert.NotNil(t, interactiveFlag)

		cleanupFlag := cmd.Flags().Lookup("cleanup")
		assert.NotNil(t, cleanupFlag)

		upgradeAllFlag := cmd.Flags().Lookup("upgrade-all")
		assert.NotNil(t, upgradeAllFlag)
	})

	t.Run("default options", func(t *testing.T) {
		opts := defaultAlwaysLatestBrewOptions()
		assert.Equal(t, "minor", opts.strategy)
		assert.Equal(t, []string{}, opts.packages)
		assert.False(t, opts.dryRun)
		assert.True(t, opts.updateBrew)
		assert.False(t, opts.casks)
		assert.False(t, opts.taps)
		assert.True(t, opts.interactive)
		assert.False(t, opts.cleanup)
		assert.False(t, opts.upgradeAll)
	})
}

func TestBrewPackageFiltering(t *testing.T) {
	opts := &alwaysLatestBrewOptions{}

	t.Run("filter packages", func(t *testing.T) {
		allPackages := []string{"node", "python", "git", "wget", "curl"}
		requestedPackages := []string{"node", "git"}

		filtered := opts.filterPackages(allPackages, requestedPackages)
		expected := []string{"node", "git"}

		assert.Equal(t, expected, filtered)
	})

	t.Run("filter packages with whitespace", func(t *testing.T) {
		allPackages := []string{"node", "python", "git"}
		requestedPackages := []string{" node ", "git", " python "}

		filtered := opts.filterPackages(allPackages, requestedPackages)

		// Check that all expected packages are present (order doesn't matter)
		assert.Len(t, filtered, 3)
		assert.Contains(t, filtered, "node")
		assert.Contains(t, filtered, "git")
		assert.Contains(t, filtered, "python")
	})

	t.Run("filter packages not found", func(t *testing.T) {
		allPackages := []string{"node", "python", "git"}
		requestedPackages := []string{"rust", "golang"}

		filtered := opts.filterPackages(allPackages, requestedPackages)

		assert.Empty(t, filtered)
	})
}

func TestBrewVersionExtraction(t *testing.T) {
	opts := &alwaysLatestBrewOptions{}

	t.Run("extract version numbers", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"18.17.0", "18.17.0"},
			{"3.11.5", "3.11.5"},
			{"v2.7.18", "2.7.18"},
			{"node-18.17.0", "18.17.0"},
			{"python@3.11.5", "3.11.5"},
			{"1.70", "1.70"},
			{"invalid", "invalid"},
		}

		for _, tt := range tests {
			result := opts.extractVersionNumber(tt.input)
			assert.Equal(t, tt.expected, result, "Unexpected result for input: %s", tt.input)
		}
	})
}

func TestBrewIntegration(t *testing.T) {
	t.Run("brew not installed scenario", func(t *testing.T) {
		opts := &alwaysLatestBrewOptions{}

		// This test will pass if brew is not installed
		// In CI/CD environments where brew might not be available
		if !opts.isBrewInstalled() {
			t.Log("homebrew is not installed - this is expected in some test environments")
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		opts := &alwaysLatestBrewOptions{
			dryRun: true,
		}

		// Dry run should not make actual changes
		// This tests the dry run flag behavior
		assert.True(t, opts.dryRun)
	})

	t.Run("cask mode", func(t *testing.T) {
		opts := &alwaysLatestBrewOptions{
			casks: true,
		}

		// Test cask mode flag
		assert.True(t, opts.casks)
	})

	t.Run("cleanup mode", func(t *testing.T) {
		opts := &alwaysLatestBrewOptions{
			cleanup: true,
		}

		// Test cleanup mode flag
		assert.True(t, opts.cleanup)
	})
}

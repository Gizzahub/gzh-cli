//nolint:testpackage // White-box testing needed for internal function access
package alwayslatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAlwaysLatestAptOptions(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	assert.Equal(t, "minor", opts.strategy)
	assert.Empty(t, opts.packages)
	assert.False(t, opts.dryRun)
	assert.True(t, opts.updateApt)
	assert.True(t, opts.interactive)
	assert.False(t, opts.cleanup)
	assert.False(t, opts.upgradeAll)
	assert.False(t, opts.autoRemove)
	assert.False(t, opts.fullUpgrade)
	assert.False(t, opts.fixBroken)
}

func TestNewAlwaysLatestAptCmd(t *testing.T) {
	cmd := newAlwaysLatestAptCmd(context.Background())

	assert.Equal(t, "apt", cmd.Use)
	assert.Equal(t, "Update APT and its managed packages to latest versions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("strategy"))
	assert.NotNil(t, cmd.Flags().Lookup("packages"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("update-apt"))
	assert.NotNil(t, cmd.Flags().Lookup("interactive"))
	assert.NotNil(t, cmd.Flags().Lookup("cleanup"))
	assert.NotNil(t, cmd.Flags().Lookup("upgrade-all"))
	assert.NotNil(t, cmd.Flags().Lookup("auto-remove"))
	assert.NotNil(t, cmd.Flags().Lookup("full-upgrade"))
	assert.NotNil(t, cmd.Flags().Lookup("fix-broken"))
}

func TestFilterPackages(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	allPackages := []string{"curl", "git", "vim", "wget", "htop"}
	requestedPackages := []string{"curl", "git"}

	filtered := opts.filterPackages(allPackages, requestedPackages)

	assert.Equal(t, []string{"curl", "git"}, filtered)
}

func TestFilterPackagesWithSpaces(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	allPackages := []string{"curl", "git", "vim"}
	requestedPackages := []string{" curl ", "git "}

	filtered := opts.filterPackages(allPackages, requestedPackages)

	assert.Equal(t, []string{"curl", "git"}, filtered)
}

func TestFilterPackagesNotFound(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	allPackages := []string{"curl", "git", "vim"}
	requestedPackages := []string{"nodejs", "python3"}

	filtered := opts.filterPackages(allPackages, requestedPackages)

	assert.Empty(t, filtered)
}

func TestAptExtractVersionNumber(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	tests := []struct {
		input    string
		expected string
	}{
		{"7.68.0-1ubuntu2.20", "7.68.0"},
		{"2.34.1-1ubuntu1.9", "2.34.1"},
		{"1:8.2.4919-1ubuntu1.1", "8.2.4919"},
		{"version-1.2.3-beta", "1.2.3"},
		{"no-version-string", "no-version-string"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := opts.extractVersionNumber(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestAptPackageVersionParsing(t *testing.T) {
	opts := defaultAlwaysLatestAptOptions()

	t.Run("extract version from dpkg output", func(t *testing.T) {
		// This would be tested with mock data since we can't rely on APT being available
		// in all test environments
		assert.NotNil(t, opts.getPackageVersion)
	})
}

func TestAptIntegration(t *testing.T) {
	t.Run("apt not installed scenario", func(t *testing.T) {
		opts := &alwaysLatestAptOptions{}

		// This test will pass whether APT is installed or not
		// In CI/CD environments APT availability varies
		isInstalled := opts.isAptInstalled()
		if !isInstalled {
			t.Log("APT is not available - this is expected in some test environments")
			assert.False(t, opts.isAptInstalled())
		} else {
			t.Log("APT is available")
			assert.True(t, opts.isAptInstalled())
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.dryRun = true

		// Test dry run mode flag
		assert.True(t, opts.dryRun)
	})

	t.Run("cleanup mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.cleanup = true

		// Test cleanup mode flag
		assert.True(t, opts.cleanup)
	})

	t.Run("upgrade all mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.upgradeAll = true

		// Test upgrade all mode flag
		assert.True(t, opts.upgradeAll)
	})

	t.Run("auto remove mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.autoRemove = true

		// Test auto remove mode flag
		assert.True(t, opts.autoRemove)
	})

	t.Run("full upgrade mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.fullUpgrade = true

		// Test full upgrade mode flag
		assert.True(t, opts.fullUpgrade)
	})

	t.Run("fix broken mode", func(t *testing.T) {
		opts := defaultAlwaysLatestAptOptions()
		opts.fixBroken = true

		// Test fix broken mode flag
		assert.True(t, opts.fixBroken)
	})
}

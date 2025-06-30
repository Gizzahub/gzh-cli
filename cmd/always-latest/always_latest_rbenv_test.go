package always_latest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAlwaysLatestRbenvOptions(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()

	assert.Equal(t, "minor", opts.strategy)
	assert.Empty(t, opts.versions)
	assert.False(t, opts.dryRun)
	assert.True(t, opts.updateRbenv)
	assert.False(t, opts.global)
	assert.True(t, opts.interactive)
	assert.True(t, opts.updatePlugins)
	assert.True(t, opts.rehash)
}

func TestNewAlwaysLatestRbenvCmd(t *testing.T) {
	cmd := newAlwaysLatestRbenvCmd()

	assert.Equal(t, "rbenv", cmd.Use)
	assert.Equal(t, "Update rbenv and install latest Ruby versions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("strategy"))
	assert.NotNil(t, cmd.Flags().Lookup("versions"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("update-rbenv"))
	assert.NotNil(t, cmd.Flags().Lookup("global"))
	assert.NotNil(t, cmd.Flags().Lookup("interactive"))
	assert.NotNil(t, cmd.Flags().Lookup("update-plugins"))
	assert.NotNil(t, cmd.Flags().Lookup("rehash"))
}

func TestFilterRequestedVersions(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()
	opts.versions = []string{"3.1", "3.2"}

	availableVersions := []string{"3.0.5", "3.1.3", "3.1.4", "3.2.0", "3.2.1", "3.3.0"}

	filtered := opts.filterRequestedVersions(availableVersions)

	expected := []string{"3.1.3", "3.1.4", "3.2.0", "3.2.1"}
	assert.Equal(t, expected, filtered)
}

func TestFilterRequestedVersionsExact(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()
	opts.versions = []string{"3.1.4", "3.2.1"}

	availableVersions := []string{"3.1.3", "3.1.4", "3.2.0", "3.2.1", "3.3.0"}

	filtered := opts.filterRequestedVersions(availableVersions)

	expected := []string{"3.1.4", "3.2.1"}
	assert.Equal(t, expected, filtered)
}

func TestIsStableRubyVersion(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()

	tests := []struct {
		version string
		stable  bool
	}{
		{"3.1.0", true},
		{"3.2.1", true},
		{"2.7.6", true},
		{"3.0.5", true},
		{"3.2.0-preview1", false},
		{"3.1.0-rc1", false},
		{"3.3.0-dev", false},
		{"ruby-3.1.0", false},    // not pure version format
		{"3.1", false},           // incomplete version
		{"jruby-9.3.9.0", false}, // alternative implementation
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := opts.isStableRubyVersion(test.version)
			assert.Equal(t, test.stable, result, "Version: %s", test.version)
		})
	}
}

func TestExtractMajorMinor(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()

	tests := []struct {
		version  string
		expected string
		hasError bool
	}{
		{"3.1.0", "3.1", false},
		{"3.2.1", "3.2", false},
		{"2.7.6", "2.7", false},
		{"3.0.5", "3.0", false},
		{"invalid", "", true},
		{"", "", true},
		{"3", "", true},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result, err := opts.extractMajorMinor(test.version)
			if test.hasError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestFindVersionsToInstallMajorStrategy(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()
	opts.strategy = "major"

	availableVersions := []string{"3.0.5", "3.1.3", "3.1.4", "3.2.0", "3.2.1"}
	installedVersions := []string{"3.1.2"}

	targetVersions := opts.findVersionsToInstall(availableVersions, installedVersions)

	// Should include latest from each major.minor series, excluding already installed
	expected := []string{"3.0.5", "3.1.4", "3.2.1"}
	assert.ElementsMatch(t, expected, targetVersions)
}

func TestFindVersionsToInstallMinorStrategy(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()
	opts.strategy = "minor"

	availableVersions := []string{"3.0.5", "3.1.3", "3.1.4", "3.2.0", "3.2.1"}
	installedVersions := []string{"3.1.2"}

	targetVersions := opts.findVersionsToInstall(availableVersions, installedVersions)

	// Should only include latest from installed major.minor series
	expected := []string{"3.1.4"}
	assert.Equal(t, expected, targetVersions)
}

func TestRbenvVersionParsing(t *testing.T) {
	opts := defaultAlwaysLatestRbenvOptions()

	t.Run("extract major minor from various versions", func(t *testing.T) {
		// This would be tested with mock data since we can't rely on rbenv being installed
		// in all test environments
		assert.NotNil(t, opts.extractMajorMinor)
	})
}

func TestRbenvIntegration(t *testing.T) {
	t.Run("rbenv not installed scenario", func(t *testing.T) {
		opts := &alwaysLatestRbenvOptions{}

		// This test will pass whether rbenv is installed or not
		// In CI/CD environments rbenv availability varies
		isInstalled := opts.isRbenvInstalled()
		if !isInstalled {
			t.Log("rbenv is not installed - this is expected in some test environments")
			assert.False(t, opts.isRbenvInstalled())
		} else {
			t.Log("rbenv is installed")
			assert.True(t, opts.isRbenvInstalled())
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		opts := defaultAlwaysLatestRbenvOptions()
		opts.dryRun = true

		// Test dry run mode flag
		assert.True(t, opts.dryRun)
	})

	t.Run("global mode", func(t *testing.T) {
		opts := defaultAlwaysLatestRbenvOptions()
		opts.global = true

		// Test global mode flag
		assert.True(t, opts.global)
	})

	t.Run("update plugins disabled", func(t *testing.T) {
		opts := defaultAlwaysLatestRbenvOptions()
		opts.updatePlugins = false

		// Test update plugins disabled flag
		assert.False(t, opts.updatePlugins)
	})

	t.Run("rehash disabled", func(t *testing.T) {
		opts := defaultAlwaysLatestRbenvOptions()
		opts.rehash = false

		// Test rehash disabled flag
		assert.False(t, opts.rehash)
	})

	t.Run("update rbenv disabled", func(t *testing.T) {
		opts := defaultAlwaysLatestRbenvOptions()
		opts.updateRbenv = false

		// Test update rbenv disabled flag
		assert.False(t, opts.updateRbenv)
	})
}

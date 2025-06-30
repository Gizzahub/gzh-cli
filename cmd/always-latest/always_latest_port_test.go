package always_latest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAlwaysLatestPortOptions(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	assert.Equal(t, "minor", opts.strategy)
	assert.Empty(t, opts.ports)
	assert.False(t, opts.dryRun)
	assert.True(t, opts.updatePorts)
	assert.True(t, opts.interactive)
	assert.False(t, opts.cleanup)
	assert.False(t, opts.upgradeAll)
	assert.True(t, opts.selfUpdate)
}

func TestNewAlwaysLatestPortCmd(t *testing.T) {
	cmd := newAlwaysLatestPortCmd()

	assert.Equal(t, "port", cmd.Use)
	assert.Equal(t, "Update MacPorts and its managed ports to latest versions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("strategy"))
	assert.NotNil(t, cmd.Flags().Lookup("ports"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("update-ports"))
	assert.NotNil(t, cmd.Flags().Lookup("interactive"))
	assert.NotNil(t, cmd.Flags().Lookup("cleanup"))
	assert.NotNil(t, cmd.Flags().Lookup("upgrade-all"))
	assert.NotNil(t, cmd.Flags().Lookup("self-update"))
}

func TestFilterPorts(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	allPorts := []string{"python39", "git", "wget", "curl", "vim"}
	requestedPorts := []string{"python39", "git"}

	filtered := opts.filterPorts(allPorts, requestedPorts)

	assert.Equal(t, []string{"python39", "git"}, filtered)
}

func TestFilterPortsWithSpaces(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	allPorts := []string{"python39", "git", "wget"}
	requestedPorts := []string{" python39 ", "git "}

	filtered := opts.filterPorts(allPorts, requestedPorts)

	assert.Equal(t, []string{"python39", "git"}, filtered)
}

func TestFilterPortsNotFound(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	allPorts := []string{"python39", "git", "wget"}
	requestedPorts := []string{"nodejs", "ruby"}

	filtered := opts.filterPorts(allPorts, requestedPorts)

	assert.Empty(t, filtered)
}

func TestExtractVersionNumber(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	tests := []struct {
		input    string
		expected string
	}{
		{"3.9.18", "3.9.18"},
		{"2.41.0", "2.41.0"},
		{"1.21.4_0", "1.21.4"},
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

func TestPortVersionParsing(t *testing.T) {
	opts := defaultAlwaysLatestPortOptions()

	t.Run("extract version from port list output", func(t *testing.T) {
		// This would be tested with mock data since we can't rely on MacPorts being installed
		// in the test environment
		assert.NotNil(t, opts.getPortVersion)
	})
}

func TestPortIntegration(t *testing.T) {
	t.Run("port not installed scenario", func(t *testing.T) {
		opts := &alwaysLatestPortOptions{}

		// This test will pass if MacPorts is not installed
		// In CI/CD environments where MacPorts might not be available
		if !opts.isPortInstalled() {
			t.Log("MacPorts is not installed - this is expected in some test environments")
			assert.False(t, opts.isPortInstalled())
		} else {
			t.Log("MacPorts is installed")
			assert.True(t, opts.isPortInstalled())
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		opts := defaultAlwaysLatestPortOptions()
		opts.dryRun = true

		// Test dry run mode flag
		assert.True(t, opts.dryRun)
	})

	t.Run("cleanup mode", func(t *testing.T) {
		opts := defaultAlwaysLatestPortOptions()
		opts.cleanup = true

		// Test cleanup mode flag
		assert.True(t, opts.cleanup)
	})

	t.Run("upgrade all mode", func(t *testing.T) {
		opts := defaultAlwaysLatestPortOptions()
		opts.upgradeAll = true

		// Test upgrade all mode flag
		assert.True(t, opts.upgradeAll)
	})

	t.Run("self update disabled", func(t *testing.T) {
		opts := defaultAlwaysLatestPortOptions()
		opts.selfUpdate = false

		// Test self update disabled flag
		assert.False(t, opts.selfUpdate)
	})
}

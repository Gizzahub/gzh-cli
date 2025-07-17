package alwayslatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAlwaysLatestSdkmanOptions(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	assert.Equal(t, "minor", opts.strategy)
	assert.Empty(t, opts.candidates)
	assert.False(t, opts.dryRun)
	assert.True(t, opts.updateSdk)
	assert.False(t, opts.global)
	assert.True(t, opts.interactive)
	assert.False(t, opts.flushBefore)
	assert.False(t, opts.cleanupOld)
}

func TestNewAlwaysLatestSdkmanCmd(t *testing.T) {
	cmd := newAlwaysLatestSdkmanCmd(context.Background())

	assert.Equal(t, "sdkman", cmd.Use)
	assert.Equal(t, "Update SDKMAN and its managed SDKs to latest versions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("strategy"))
	assert.NotNil(t, cmd.Flags().Lookup("candidates"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("update-sdk"))
	assert.NotNil(t, cmd.Flags().Lookup("global"))
	assert.NotNil(t, cmd.Flags().Lookup("interactive"))
	assert.NotNil(t, cmd.Flags().Lookup("flush-before"))
	assert.NotNil(t, cmd.Flags().Lookup("cleanup-old"))
}

func TestFilterCandidates(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	allCandidates := []string{"java", "gradle", "maven", "kotlin", "scala"}
	requestedCandidates := []string{"java", "gradle"}

	filtered := opts.filterCandidates(allCandidates, requestedCandidates)

	assert.Equal(t, []string{"java", "gradle"}, filtered)
}

func TestFilterCandidatesWithSpaces(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	allCandidates := []string{"java", "gradle", "maven"}
	requestedCandidates := []string{" java ", "gradle "}

	filtered := opts.filterCandidates(allCandidates, requestedCandidates)

	assert.Equal(t, []string{"java", "gradle"}, filtered)
}

func TestIsValidVersion(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	tests := []struct {
		version string
		valid   bool
	}{
		{"11.0.19", true},
		{"17.0.7", true},
		{"8.0.372", true},
		{"21.ea.35", true},
		{"invalid", false},
		{"", false},
		{"text-only", false},
		{"11", false}, // Too short
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := opts.isValidVersion(test.version)
			assert.Equal(t, test.valid, result, "Version: %s", test.version)
		})
	}
}

func TestFilterStableVersions(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	versions := []string{
		"11.0.19",
		"17.0.7",
		"21.ea.35",
		"11.0.20-beta",
		"17.0.8-rc.1",
		"8.0.372",
		"19.0.1-alpha",
	}

	stable := opts.filterStableVersions(versions)

	expected := []string{"11.0.19", "17.0.7", "8.0.372"}
	assert.Equal(t, expected, stable)
}

func TestIsStableVersion(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	tests := []struct {
		version string
		stable  bool
	}{
		{"11.0.19", true},
		{"17.0.7", true},
		{"21.ea.35", false},
		{"11.0.20-beta", false},
		{"17.0.8-rc.1", false},
		{"19.0.1-alpha", false},
		{"8.0.372-snapshot", false},
		{"17.preview.1", false},
		{"11.dev.123", false},
		{"21.experimental", false},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := opts.isStableVersion(test.version)
			assert.Equal(t, test.stable, result, "Version: %s", test.version)
		})
	}
}

func TestExtractMajorVersion(t *testing.T) {
	opts := defaultAlwaysLatestSdkmanOptions()

	tests := []struct {
		version  string
		expected string
		hasError bool
	}{
		{"11.0.19", "11", false},
		{"17.0.7", "17", false},
		{"8.0.372", "8", false},
		{"21.ea.35", "21", false},
		{"invalid", "", true},
		{"", "", true},
		{"text-only", "", true},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result, err := opts.extractMajorVersion(test.version)
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

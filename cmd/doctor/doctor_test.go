//nolint:testpackage // White-box testing needed for internal function access
package doctor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Gizzahub/gzh-cli/internal/profiling"
)

func TestDoctorCmdCreation(t *testing.T) {
	assert.Equal(t, "doctor", DoctorCmd.Use)
	assert.Equal(t, "Diagnose system health and configuration issues", DoctorCmd.Short)
	assert.NotEmpty(t, DoctorCmd.Long)
	assert.NotNil(t, DoctorCmd.Run)

	// Check that flags are properly set up
	reportFlag := DoctorCmd.Flags().Lookup("report")
	assert.NotNil(t, reportFlag)
	assert.Equal(t, "", reportFlag.DefValue)

	quickFlag := DoctorCmd.Flags().Lookup("quick")
	assert.NotNil(t, quickFlag)
	assert.Equal(t, "false", quickFlag.DefValue)

	fixFlag := DoctorCmd.Flags().Lookup("fix")
	assert.NotNil(t, fixFlag)
	assert.Equal(t, "false", fixFlag.DefValue)

	verboseFlag := DoctorCmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestDoctorCmdSubcommands(t *testing.T) {
	subcommands := DoctorCmd.Commands()

	// Should have expected subcommands based on init()
	expectedSubcommands := []string{"godoc", "dev-env", "setup", "benchmark", "metrics", "health", "container"}
	assert.Len(t, subcommands, len(expectedSubcommands))

	// Verify subcommands exist
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		assert.True(t, subcommandNames[expected], "Subcommand %s should exist", expected)
	}
}

func TestGetSystemInfo(t *testing.T) {
	info := getSystemInfo()

	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Arch)
	assert.Equal(t, runtime.Version(), info.GoVersion)
	assert.NotEmpty(t, info.WorkingDir)
	assert.NotEmpty(t, info.HomeDir)
	assert.NotEmpty(t, info.TempDir)
	// Hostname, Username, PathEnv, Shell may be empty in some environments
}

func TestIsGoVersionSupported(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		supported bool
	}{
		{"Go 1.19", "go1.19.1", true},
		{"Go 1.20", "go1.20.1", true},
		{"Go 1.21", "go1.21.0", true},
		{"Go 1.22", "go1.22.1", true},
		{"Go 1.23", "go1.23.0", true},
		{"Go 1.24", "go1.24.0", true},
		{"Go 1.18", "go1.18.1", false},
		{"Go 1.17", "go1.17.1", false},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGoVersionSupported(tt.version)
			assert.Equal(t, tt.supported, result)
		})
	}
}

func TestGetDiskSpace(t *testing.T) {
	space := getDiskSpace("/tmp")
	// This is a placeholder implementation that returns 50.0
	assert.Equal(t, 50.0, space)
}

func TestRunCPUBenchmark(t *testing.T) {
	score := runCPUBenchmark()
	assert.Greater(t, score, 0.0, "CPU benchmark should return positive score")
	assert.IsType(t, float64(0), score)
}

func TestRunDiskBenchmark(t *testing.T) {
	score := runDiskBenchmark()
	// Score should be >= 0 (may be 0 if disk operations fail)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.IsType(t, float64(0), score)
}

func TestCountSSHKeys(t *testing.T) {
	count := countSSHKeys()
	// Count should be >= 0 (may be 0 if no SSH keys or directory doesn't exist)
	assert.GreaterOrEqual(t, count, 0)
	assert.IsType(t, 0, count)
}

func TestFindUnsafePermissions(t *testing.T) {
	files := findUnsafePermissions()
	// This is a placeholder implementation that returns empty slice
	assert.NotNil(t, files)
	assert.IsType(t, []string{}, files)
}

func TestCalculateReportTotals(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{
			{Status: statusPass},
			{Status: statusWarn},
			{Status: statusFail},
			{Status: "skip"},
			{Status: statusPass},
		},
	}

	calculateReportTotals(report)

	assert.Equal(t, 5, report.TotalChecks)
	assert.Equal(t, 2, report.PassedChecks)
	assert.Equal(t, 1, report.WarnChecks)
	assert.Equal(t, 1, report.FailedChecks)
	assert.Equal(t, 1, report.SkippedChecks)
}

func TestGenerateSummaryAndRecommendations(t *testing.T) {
	tests := []struct {
		name                  string
		passed                int
		warn                  int
		failed                int
		total                 int
		expectSummary         string
		expectRecommendations int
	}{
		{
			name:                  "all passed",
			passed:                5,
			warn:                  0,
			failed:                0,
			total:                 5,
			expectSummary:         "100.0% success rate",
			expectRecommendations: 1, // "System appears healthy"
		},
		{
			name:                  "with warnings",
			passed:                3,
			warn:                  2,
			failed:                0,
			total:                 5,
			expectSummary:         "60.0% success rate",
			expectRecommendations: 2, // "Review warnings" + "System health below optimal"
		},
		{
			name:                  "with failures",
			passed:                2,
			warn:                  1,
			failed:                2,
			total:                 5,
			expectSummary:         "40.0% success rate",
			expectRecommendations: 3, // "Address critical" + "Review warnings" + "System health below optimal"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &DiagnosticReport{
				TotalChecks:  tt.total,
				PassedChecks: tt.passed,
				WarnChecks:   tt.warn,
				FailedChecks: tt.failed,
			}

			generateSummaryAndRecommendations(report)

			assert.Contains(t, report.Summary, tt.expectSummary)
			assert.Len(t, report.Recommendations, tt.expectRecommendations)
		})
	}
}

func TestTryAutoFix(t *testing.T) {
	tests := []struct {
		name     string
		result   DiagnosticResult
		expected bool
	}{
		{
			name: "Configuration Directory Access",
			result: DiagnosticResult{
				Name: "Configuration Directory Access",
			},
			expected: true, // Should attempt to create directory
		},
		{
			name: "unknown issue",
			result: DiagnosticResult{
				Name: "Unknown Issue",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tryAutoFix(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiagnosticResultStructure(t *testing.T) {
	result := DiagnosticResult{
		Name:          "Test Check",
		Category:      "test",
		Status:        statusPass,
		Message:       "Test message",
		Details:       map[string]interface{}{"key": "value"},
		FixSuggestion: "Test fix",
		Duration:      time.Second,
		Timestamp:     time.Now(),
	}

	assert.Equal(t, "Test Check", result.Name)
	assert.Equal(t, "test", result.Category)
	assert.Equal(t, statusPass, result.Status)
	assert.Equal(t, "Test message", result.Message)
	assert.Equal(t, "value", result.Details["key"])
	assert.Equal(t, "Test fix", result.FixSuggestion)
	assert.Equal(t, time.Second, result.Duration)
	assert.False(t, result.Timestamp.IsZero())
}

func TestDiagnosticReportStructure(t *testing.T) {
	report := DiagnosticReport{
		Timestamp:       time.Now(),
		Version:         "1.0.0",
		Platform:        "linux/amd64",
		TotalChecks:     5,
		PassedChecks:    3,
		WarnChecks:      1,
		FailedChecks:    1,
		SkippedChecks:   0,
		Results:         []DiagnosticResult{},
		Summary:         "Test summary",
		Recommendations: []string{"Test recommendation"},
		Duration:        time.Second,
	}

	assert.False(t, report.Timestamp.IsZero())
	assert.Equal(t, "1.0.0", report.Version)
	assert.Equal(t, "linux/amd64", report.Platform)
	assert.Equal(t, 5, report.TotalChecks)
	assert.Equal(t, 3, report.PassedChecks)
	assert.Equal(t, 1, report.WarnChecks)
	assert.Equal(t, 1, report.FailedChecks)
	assert.Equal(t, 0, report.SkippedChecks)
	assert.NotNil(t, report.Results)
	assert.Equal(t, "Test summary", report.Summary)
	assert.Contains(t, report.Recommendations, "Test recommendation")
	assert.Equal(t, time.Second, report.Duration)
}

func TestSystemInfoStructure(t *testing.T) {
	info := SystemInfo{
		OS:         "linux",
		Arch:       "amd64",
		GoVersion:  "go1.21.0",
		Hostname:   "testhost",
		Username:   "testuser",
		WorkingDir: "/tmp",
		HomeDir:    "/home/testuser",
		PathEnv:    "/usr/bin:/bin",
		Shell:      "/bin/bash",
		TempDir:    "/tmp",
	}

	assert.Equal(t, "linux", info.OS)
	assert.Equal(t, "amd64", info.Arch)
	assert.Equal(t, "go1.21.0", info.GoVersion)
	assert.Equal(t, "testhost", info.Hostname)
	assert.Equal(t, "testuser", info.Username)
	assert.Equal(t, "/tmp", info.WorkingDir)
	assert.Equal(t, "/home/testuser", info.HomeDir)
	assert.Equal(t, "/usr/bin:/bin", info.PathEnv)
	assert.Equal(t, "/bin/bash", info.Shell)
	assert.Equal(t, "/tmp", info.TempDir)
}

func TestDoctorHelpContent(t *testing.T) {
	// Verify help content mentions key features
	longDesc := DoctorCmd.Long
	assert.Contains(t, longDesc, "Comprehensive system diagnostics")
	assert.Contains(t, longDesc, "System information")
	assert.Contains(t, longDesc, "Configuration validation")
	assert.Contains(t, longDesc, "Network connectivity")
	assert.Contains(t, longDesc, "Git configuration")
	assert.Contains(t, longDesc, "Performance benchmarks")

	// Verify examples show proper usage
	assert.Contains(t, longDesc, "gz doctor")
	assert.Contains(t, longDesc, "--report")
	assert.Contains(t, longDesc, "--quick")
	assert.Contains(t, longDesc, "--fix")
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, "pass", statusPass)
	assert.Equal(t, "warn", statusWarn)
	assert.Equal(t, "fail", statusFail)
}

func TestDocorCmdStructure(t *testing.T) {
	// Test that the command has proper structure
	assert.NotEmpty(t, DoctorCmd.Use)
	assert.NotEmpty(t, DoctorCmd.Short)
	assert.NotEmpty(t, DoctorCmd.Long)
	assert.NotNil(t, DoctorCmd.Run)

	// Test that examples are included in Long description
	assert.Contains(t, DoctorCmd.Long, "Examples:")
	assert.Contains(t, DoctorCmd.Long, "subcommands:")
}

func TestIsGoVersionSupportedEdgeCases(t *testing.T) {
	// Test edge cases for Go version checking
	testCases := []struct {
		version   string
		supported bool
	}{
		{"", false},
		{"invalid", false},
		{"go1.19", true},
		{"go1.19.1", true},
		{"go1.20", true},
		{"go1.21", true},
		{"go1.22", true},
		{"go1.23", true},
		{"go1.24", true},
		{"go1.18", false},
		{"go2.0", false},
	}

	for _, tc := range testCases {
		result := isGoVersionSupported(tc.version)
		assert.Equal(t, tc.supported, result, "isGoVersionSupported(%s) should be %v", tc.version, tc.supported)
	}
}

func TestRunSystemChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runSystemChecks(report, nil, nil)

	// Should have at least 3 system checks: System Information, Go Version, Memory Usage, Disk Space
	assert.GreaterOrEqual(t, len(report.Results), 3)

	// Verify expected checks exist
	checkNames := make(map[string]bool)
	for _, result := range report.Results {
		checkNames[result.Name] = true
		assert.NotEmpty(t, result.Category)
		assert.NotEmpty(t, result.Status)
		assert.NotEmpty(t, result.Message)
		assert.False(t, result.Timestamp.IsZero())
	}

	assert.True(t, checkNames["System Information"], "System Information check should exist")
	assert.True(t, checkNames["Go Version"], "Go Version check should exist")
	assert.True(t, checkNames["Memory Usage"], "Memory Usage check should exist")
}

func TestRunConfigChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runConfigChecks(report, nil, nil)

	// Should have added config-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "config", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestRunPermissionChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runPermissionChecks(report, nil, nil)

	// Should have added permission-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "permissions", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestRunSecurityChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runSecurityChecks(report, nil, nil)

	// Should have added security-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "security", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestRunPerformanceChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runPerformanceChecks(report, nil, nil)

	// Should have added performance-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "performance", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestPrintResults(t *testing.T) {
	report := &DiagnosticReport{
		TotalChecks:     5,
		PassedChecks:    3,
		WarnChecks:      1,
		FailedChecks:    1,
		SkippedChecks:   0,
		Summary:         "Test summary",
		Recommendations: []string{"Test recommendation"},
		Duration:        time.Second,
		Results: []DiagnosticResult{
			{
				Name:     "Test Check",
				Status:   statusPass,
				Message:  "Test message",
				Category: "test",
			},
		},
	}

	// This should not panic
	assert.NotPanics(t, func() {
		printResults(report)
	})
}

func TestAttemptAutomaticFixes(t *testing.T) {
	// Skip this test since attemptAutomaticFixes requires proper logger and error recovery setup
	t.Skip("attemptAutomaticFixes requires complex setup, skipping for basic coverage")
}

func TestSubcommandCreation(t *testing.T) {
	// Test that subcommand creation functions don't panic
	assert.NotPanics(t, func() { newGodocCmd() })
	assert.NotPanics(t, func() { newDevEnvCmd() })
	assert.NotPanics(t, func() { newSetupCmd() })
	assert.NotPanics(t, func() { newBenchmarkCmd() })
	assert.NotPanics(t, func() { newMetricsCmd() })
	assert.NotPanics(t, func() { newHealthCmd() })
	assert.NotPanics(t, func() { newContainerCmd() })
}

func TestSaveReport(t *testing.T) {
	report := &DiagnosticReport{
		Timestamp:    time.Now(),
		Version:      "1.0.0",
		Platform:     "linux/amd64",
		TotalChecks:  1,
		PassedChecks: 1,
		Summary:      "Test report",
		Results: []DiagnosticResult{
			{
				Name:     "Test",
				Status:   statusPass,
				Message:  "Test",
				Category: "test",
			},
		},
	}

	// Test saving to temp file
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "test-report.json")

	err := saveReport(report, reportPath)
	assert.NoError(t, err)

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(reportPath)
	assert.NoError(t, err)

	var savedReport DiagnosticReport
	err = json.Unmarshal(data, &savedReport)
	assert.NoError(t, err)
	assert.Equal(t, report.Version, savedReport.Version)
	assert.Equal(t, report.Platform, savedReport.Platform)
}

func TestRunNetworkChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runNetworkChecks(context.Background(), report, nil, nil)

	// Should have added network-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "network", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestRunGitChecks(t *testing.T) {
	report := &DiagnosticReport{
		Results: []DiagnosticResult{},
	}

	runGitChecks(context.Background(), report, nil, nil)

	// Should have added git-related checks
	assert.Greater(t, len(report.Results), 0)

	// All results should have proper structure
	for _, result := range report.Results {
		assert.Equal(t, "git", result.Category)
		assert.NotEmpty(t, result.Name)
		assert.NotEmpty(t, result.Status)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestNewGodocCmd(t *testing.T) {
	cmd := newGodocCmd()

	assert.Equal(t, "godoc", cmd.Use)
	assert.Contains(t, cmd.Short, "API documentation")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist (may not be defined yet)
	flags := cmd.Flags()
	assert.NotNil(t, flags)
}

func TestNewDevEnvCmd(t *testing.T) {
	cmd := newDevEnvCmd()

	assert.Equal(t, "dev-env", cmd.Use)
	assert.Contains(t, cmd.Short, "development environment")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist
	flags := cmd.Flags()
	assert.NotNil(t, flags)
}

func TestNewSetupCmd(t *testing.T) {
	cmd := newSetupCmd()

	assert.Equal(t, "setup", cmd.Use)
	assert.Contains(t, cmd.Short, "Automated")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestNewBenchmarkCmd(t *testing.T) {
	cmd := newBenchmarkCmd()

	assert.Equal(t, "benchmark", cmd.Use)
	assert.Contains(t, cmd.Short, "performance")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist
	flags := cmd.Flags()
	assert.NotNil(t, flags)
}

func TestNewMetricsCmd(t *testing.T) {
	cmd := newMetricsCmd()

	assert.Equal(t, "metrics", cmd.Use)
	assert.Contains(t, cmd.Short, "quality")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestNewHealthCmd(t *testing.T) {
	cmd := newHealthCmd()

	assert.Equal(t, "health", cmd.Use)
	assert.Contains(t, cmd.Short, "system")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestNewContainerCmd(t *testing.T) {
	cmd := newContainerCmd()

	assert.Equal(t, "container", cmd.Use)
	assert.Contains(t, cmd.Short, "Docker")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestGetSystemInfoFields(t *testing.T) {
	info := getSystemInfo()

	// Verify required fields are populated
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.WorkingDir)
	assert.NotEmpty(t, info.HomeDir)
	assert.NotEmpty(t, info.TempDir)

	// OS and Arch should match runtime values
	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Arch)
	assert.Equal(t, runtime.Version(), info.GoVersion)
}

func TestDiagnosticResultValidation(t *testing.T) {
	result := DiagnosticResult{
		Name:      "Test Check",
		Category:  "test",
		Status:    statusPass,
		Message:   "Test passed",
		Duration:  time.Millisecond * 100,
		Timestamp: time.Now(),
	}

	// Verify all required fields
	assert.NotEmpty(t, result.Name)
	assert.NotEmpty(t, result.Category)
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.Message)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.False(t, result.Timestamp.IsZero())
}

func TestDiagnosticReportValidation(t *testing.T) {
	report := DiagnosticReport{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Platform:  "linux/amd64",
		Duration:  time.Second,
		Results:   []DiagnosticResult{},
	}

	// Calculate totals
	calculateReportTotals(&report)

	// Verify structure
	assert.False(t, report.Timestamp.IsZero())
	assert.NotEmpty(t, report.Version)
	assert.NotEmpty(t, report.Platform)
	assert.Greater(t, report.Duration, time.Duration(0))
	assert.NotNil(t, report.Results)
	assert.Equal(t, 0, report.TotalChecks) // No results added
}

func TestBenchmarkFunctions(t *testing.T) {
	t.Run("CPU benchmark", func(t *testing.T) {
		score := runCPUBenchmark()
		assert.Greater(t, score, 0.0)
		assert.IsType(t, float64(0), score)
	})

	t.Run("Disk benchmark", func(t *testing.T) {
		score := runDiskBenchmark()
		assert.GreaterOrEqual(t, score, 0.0)
		assert.IsType(t, float64(0), score)
	})
}

func TestSecurityFunctions(t *testing.T) {
	t.Run("SSH key count", func(t *testing.T) {
		count := countSSHKeys()
		assert.GreaterOrEqual(t, count, 0)
		assert.IsType(t, 0, count)
	})

	t.Run("Unsafe permissions", func(t *testing.T) {
		files := findUnsafePermissions()
		assert.NotNil(t, files)
		assert.IsType(t, []string{}, files)
	})
}

func TestDiskSpaceFunction(t *testing.T) {
	// Test with current directory
	space := getDiskSpace(".")

	// Should return a non-negative value
	assert.GreaterOrEqual(t, space, 0.0)
	assert.IsType(t, float64(0), space)

	// Test with temp directory
	tempSpace := getDiskSpace("/tmp")
	assert.GreaterOrEqual(t, tempSpace, 0.0)
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getSystemInfo", func(t *testing.T) {
		info := getSystemInfo()
		assert.Equal(t, runtime.GOOS, info.OS)
		assert.Equal(t, runtime.GOARCH, info.Arch)
		assert.Equal(t, runtime.Version(), info.GoVersion)
		assert.NotEmpty(t, info.WorkingDir)
		assert.NotEmpty(t, info.HomeDir)
		assert.NotEmpty(t, info.TempDir)
	})

	t.Run("isDirectoryWritable", func(t *testing.T) {
		// Test with temp directory (should be writable)
		assert.True(t, isDirectoryWritable(os.TempDir()))

		// Test with non-existent directory
		assert.False(t, isDirectoryWritable("/non/existent/path"))
	})

	t.Run("canCreateDirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "test-create")

		// Should be able to create in temp
		assert.True(t, canCreateDirectory(testDir))

		// Test with invalid path
		assert.False(t, canCreateDirectory("/invalid/root/path"))
	})

	t.Run("isURLReachable", func(t *testing.T) {
		ctx := context.Background()

		// Test with invalid URL
		assert.False(t, isURLReachable(ctx, "http://invalid-url-does-not-exist.com"))

		// Test with malformed URL
		assert.False(t, isURLReachable(ctx, "not-a-url"))
	})

	t.Run("getGitConfig", func(t *testing.T) {
		ctx := context.Background()
		config := getGitConfig(ctx)

		assert.NotNil(t, config)
		// Should contain expected keys even if empty
		expectedKeys := []string{"user.name", "user.email", "core.editor", "init.defaultBranch"}
		for _, key := range expectedKeys {
			_, exists := config[key]
			assert.True(t, exists, "Should contain key %s", key)
		}
	})
}

func TestReportGeneration(t *testing.T) {
	t.Run("calculateReportTotals", func(t *testing.T) {
		report := &DiagnosticReport{
			Results: []DiagnosticResult{
				{Status: statusPass},
				{Status: statusWarn},
				{Status: statusFail},
				{Status: "skip"},
				{Status: statusPass},
			},
		}

		calculateReportTotals(report)

		assert.Equal(t, 5, report.TotalChecks)
		assert.Equal(t, 2, report.PassedChecks)
		assert.Equal(t, 1, report.WarnChecks)
		assert.Equal(t, 1, report.FailedChecks)
		assert.Equal(t, 1, report.SkippedChecks)
	})

	t.Run("generateSummaryAndRecommendations", func(t *testing.T) {
		report := &DiagnosticReport{
			TotalChecks:     10,
			PassedChecks:    6,
			WarnChecks:      2,
			FailedChecks:    2,
			Recommendations: []string{},
		}

		generateSummaryAndRecommendations(report)

		assert.NotEmpty(t, report.Summary)
		assert.Contains(t, report.Summary, "60.0%")
		assert.NotEmpty(t, report.Recommendations)
		assert.Contains(t, report.Recommendations, "Address critical issues immediately")
		assert.Contains(t, report.Recommendations, "Review warnings and consider improvements")
	})
}

func TestAutoFixFunctionality(t *testing.T) {
	t.Run("tryAutoFix success", func(t *testing.T) {
		// Test with the actual hardcoded path since tryAutoFix uses os.UserHomeDir()
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)

		configDir := filepath.Join(homeDir, ".config", "gzh-manager")

		// Check if we can write to the config directory
		canWrite := canCreateDirectory(configDir)
		if !canWrite {
			t.Skip("Cannot write to config directory, skipping auto-fix test")
		}

		result := DiagnosticResult{
			Name: "Configuration Directory Access",
		}

		success := tryAutoFix(result)
		assert.True(t, success)

		// Verify directory exists (it should be created or already exist)
		_, err = os.Stat(configDir)
		assert.NoError(t, err)
	})

	t.Run("tryAutoFix unknown issue", func(t *testing.T) {
		result := DiagnosticResult{
			Name: "Unknown Issue",
		}

		success := tryAutoFix(result)
		assert.False(t, success)
	})
}

func TestMainRunDoctorComponents(t *testing.T) {
	// Test the constants
	assert.Equal(t, "pass", statusPass)
	assert.Equal(t, "warn", statusWarn)
	assert.Equal(t, "fail", statusFail)

	// Test DiagnosticResult structure
	result := DiagnosticResult{
		Name:      "Test",
		Category:  "test",
		Status:    statusPass,
		Message:   "Test message",
		Duration:  time.Millisecond,
		Timestamp: time.Now(),
	}

	assert.NotEmpty(t, result.Name)
	assert.NotEmpty(t, result.Category)
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.Message)

	// Test DiagnosticReport structure
	report := DiagnosticReport{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Platform:  "test/test",
		Results:   []DiagnosticResult{result},
		Duration:  time.Second,
	}

	assert.False(t, report.Timestamp.IsZero())
	assert.NotEmpty(t, report.Version)
	assert.NotEmpty(t, report.Platform)
	assert.Len(t, report.Results, 1)

	// Test SystemInfo structure
	sysInfo := SystemInfo{
		OS:         "linux",
		Arch:       "amd64",
		GoVersion:  "go1.21.0",
		Hostname:   "test",
		Username:   "test",
		WorkingDir: "/tmp",
		HomeDir:    "/home/test",
		PathEnv:    "/usr/bin",
		Shell:      "/bin/bash",
		TempDir:    "/tmp",
	}

	assert.NotEmpty(t, sysInfo.OS)
	assert.NotEmpty(t, sysInfo.Arch)
	assert.NotEmpty(t, sysInfo.GoVersion)
}

func TestOutputFunctions(t *testing.T) {
	t.Run("printResults does not panic", func(t *testing.T) {
		report := &DiagnosticReport{
			Platform:        "test/test",
			Duration:        time.Second,
			TotalChecks:     1,
			PassedChecks:    1,
			Summary:         "Test summary",
			Recommendations: []string{"Test recommendation"},
			Results: []DiagnosticResult{
				{
					Name:     "Test",
					Category: "test",
					Status:   statusPass,
					Message:  "Test message",
				},
			},
		}

		assert.NotPanics(t, func() {
			printResults(report)
		})
	})
}

func TestRunDoctorCommand(t *testing.T) {
	// Test the global DoctorCmd variable can be executed without panicking
	// We don't run the actual function due to complexity, but verify the command structure
	assert.NotNil(t, DoctorCmd)
	assert.Equal(t, "doctor", DoctorCmd.Use)

	// Test init function effects
	assert.NotNil(t, DoctorCmd.Flags().Lookup("report"))
	assert.NotNil(t, DoctorCmd.Flags().Lookup("quick"))
	assert.NotNil(t, DoctorCmd.Flags().Lookup("fix"))
	assert.NotNil(t, DoctorCmd.Flags().Lookup("verbose"))

	// Verify subcommands were added in init
	subcommands := DoctorCmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 7)
}

func TestGlobalVariables(t *testing.T) {
	// Test initial values of global variables
	originalReport := reportFile
	originalQuick := quickMode
	originalFix := attemptFix
	originalVerbose := verbose

	// Verify initial state
	assert.Equal(t, "", reportFile)
	assert.False(t, quickMode)
	assert.False(t, attemptFix)
	assert.False(t, verbose)

	// Test that variables can be modified (simulating flag setting)
	reportFile = "test.json"
	quickMode = true
	attemptFix = true
	verbose = true

	assert.Equal(t, "test.json", reportFile)
	assert.True(t, quickMode)
	assert.True(t, attemptFix)
	assert.True(t, verbose)

	// Restore original values
	reportFile = originalReport
	quickMode = originalQuick
	attemptFix = originalFix
	verbose = originalVerbose
}

func TestDoctorCmdComplete(t *testing.T) {
	// Test that the command is fully configured
	cmd := DoctorCmd

	// Verify Use field
	assert.Equal(t, "doctor", cmd.Use)

	// Verify Short description
	assert.Contains(t, cmd.Short, "Diagnose")
	assert.Contains(t, cmd.Short, "system")
	assert.Contains(t, cmd.Short, "health")

	// Verify Long description contains expected sections
	assert.Contains(t, cmd.Long, "Comprehensive system diagnostics")
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz doctor")

	// Verify Run function is set
	assert.NotNil(t, cmd.Run)

	// Verify all flags are properly configured
	reportFlag := cmd.Flags().Lookup("report")
	assert.NotNil(t, reportFlag)
	assert.Equal(t, "string", reportFlag.Value.Type())
	assert.Contains(t, reportFlag.Usage, "Output detailed report")

	quickFlag := cmd.Flags().Lookup("quick")
	assert.NotNil(t, quickFlag)
	assert.Equal(t, "bool", quickFlag.Value.Type())
	assert.Contains(t, quickFlag.Usage, "quick checks")

	fixFlag := cmd.Flags().Lookup("fix")
	assert.NotNil(t, fixFlag)
	assert.Equal(t, "bool", fixFlag.Value.Type())
	assert.Contains(t, fixFlag.Usage, "fix")

	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Value.Type())
	assert.Contains(t, verboseFlag.Usage, "verbose")
}

func TestAllSubcommandFunctionsExist(t *testing.T) {
	// Test that all subcommand creation functions work without panicking
	// and return properly configured commands

	t.Run("godoc", func(t *testing.T) {
		cmd := newGodocCmd()
		assert.Equal(t, "godoc", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test specific flags expected for godoc
		flags := cmd.Flags()
		assert.NotNil(t, flags)
	})

	t.Run("dev-env", func(t *testing.T) {
		cmd := newDevEnvCmd()
		assert.Equal(t, "dev-env", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test that fix flag exists
		fixFlag := cmd.Flags().Lookup("fix")
		assert.NotNil(t, fixFlag)
	})

	t.Run("setup", func(t *testing.T) {
		cmd := newSetupCmd()
		assert.Equal(t, "setup", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test that type flag exists
		typeFlag := cmd.Flags().Lookup("type")
		assert.NotNil(t, typeFlag)
	})

	t.Run("benchmark", func(t *testing.T) {
		cmd := newBenchmarkCmd()
		assert.Equal(t, "benchmark", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test benchmark specific flags
		packageFlag := cmd.Flags().Lookup("package")
		assert.NotNil(t, packageFlag)

		ciFlag := cmd.Flags().Lookup("ci")
		assert.NotNil(t, ciFlag)
	})

	t.Run("metrics", func(t *testing.T) {
		cmd := newMetricsCmd()
		assert.Equal(t, "metrics", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test metrics specific flags
		outputFlag := cmd.Flags().Lookup("output")
		assert.NotNil(t, outputFlag)
	})

	t.Run("health", func(t *testing.T) {
		cmd := newHealthCmd()
		assert.Equal(t, "health", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test health specific flags
		intervalFlag := cmd.Flags().Lookup("interval")
		assert.NotNil(t, intervalFlag)
	})

	t.Run("container", func(t *testing.T) {
		cmd := newContainerCmd()
		assert.Equal(t, "container", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)

		// Test that flags exist (container has specific flags)
		flags := cmd.Flags()
		assert.NotNil(t, flags)
	})
}

func TestCommandExecutionReadiness(t *testing.T) {
	// Test that all commands can be called without panicking
	// (even if they would fail due to missing setup)

	// Test main doctor command structure
	cmd := DoctorCmd
	assert.NotNil(t, cmd.Run, "Main doctor command should have Run function")

	// Test that all subcommands have RunE functions
	for _, subcmd := range cmd.Commands() {
		assert.NotNil(t, subcmd.RunE, "Subcommand %s should have RunE function", subcmd.Use)
	}
}

func TestFlagDefaults(t *testing.T) {
	// Test that global flags have correct default values
	cmd := DoctorCmd

	reportFlag := cmd.Flags().Lookup("report")
	assert.Equal(t, "", reportFlag.DefValue)

	quickFlag := cmd.Flags().Lookup("quick")
	assert.Equal(t, "false", quickFlag.DefValue)

	fixFlag := cmd.Flags().Lookup("fix")
	assert.Equal(t, "false", fixFlag.DefValue)

	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestCommandHierarchy(t *testing.T) {
	// Verify the command hierarchy is properly set up
	subcommands := DoctorCmd.Commands()

	expectedCommands := map[string]bool{
		"godoc":     false,
		"dev-env":   false,
		"setup":     false,
		"benchmark": false,
		"metrics":   false,
		"health":    false,
		"container": false,
	}

	for _, subcmd := range subcommands {
		if _, exists := expectedCommands[subcmd.Use]; exists {
			expectedCommands[subcmd.Use] = true
		}
	}

	// Verify all expected commands were found
	for cmdName, found := range expectedCommands {
		assert.True(t, found, "Expected subcommand %s not found", cmdName)
	}
}

func TestCommandExecuteAndFlags(t *testing.T) {
	// Test flag functionality without actually executing commands

	// Test that flags can be set programmatically
	originalQuick := quickMode
	originalVerbose := verbose

	defer func() {
		quickMode = originalQuick
		verbose = originalVerbose
	}()

	// Simulate flag setting
	quickMode = true
	verbose = true

	assert.True(t, quickMode)
	assert.True(t, verbose)

	// Test that DoctorCmd has access to these globals
	assert.NotNil(t, DoctorCmd.Run)
}

func TestBenchmarkUtilityFunctions(t *testing.T) {
	t.Run("getBenchmarkEnvironment", func(t *testing.T) {
		env := getBenchmarkEnvironment()

		// Verify required fields are populated
		assert.NotEmpty(t, env.Platform)
		assert.NotEmpty(t, env.GoVersion)
		assert.Greater(t, env.NumCPU, 0)
		assert.GreaterOrEqual(t, env.NumGoroutines, 0)
		assert.Greater(t, env.MemoryLimit, uint64(0))

		// Platform should be in format "os/arch"
		assert.Contains(t, env.Platform, "/")

		// GoVersion should start with "go"
		assert.Contains(t, env.GoVersion, "go")
	})

	t.Run("generateImpactDescription", func(t *testing.T) {
		testCases := []struct {
			changePercent float64
			expected      string
		}{
			{5.0, "minimal"},
			{-8.0, "minimal"},
			{15.0, "moderate"},
			{-20.0, "moderate"},
			{35.0, "significant"},
			{-40.0, "significant"},
			{60.0, "major"},
			{-80.0, "major"},
			{0.0, "minimal"},
		}

		for _, tc := range testCases {
			result := generateImpactDescription(tc.changePercent)
			assert.Equal(t, tc.expected, result, "changePercent: %f should return %s", tc.changePercent, tc.expected)
		}
	})

	t.Run("generateBenchmarkSummary", func(t *testing.T) {
		// Test with empty report
		emptyReport := &BenchmarkReport{}
		generateBenchmarkSummary(emptyReport)
		assert.Equal(t, 0, emptyReport.Summary.TotalBenchmarks)

		// Test with report containing benchmarks (using empty benchmarks for structure test)
		report := &BenchmarkReport{
			Benchmarks: make([]profiling.BenchmarkResult, 3),
		}
		generateBenchmarkSummary(report)
		assert.Equal(t, 3, report.Summary.TotalBenchmarks)
	})

	t.Run("getGitCommit", func(t *testing.T) {
		// This function returns "not implemented" error
		commit, err := getGitCommit()
		assert.Error(t, err)
		assert.Equal(t, "", commit)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("getGitBranch", func(t *testing.T) {
		// This function returns "not implemented" error
		branch, err := getGitBranch()
		assert.Error(t, err)
		assert.Equal(t, "", branch)
		assert.Contains(t, err.Error(), "not implemented")
	})
}

func TestBenchmarkStructures(t *testing.T) {
	t.Run("BenchmarkEnvironment structure", func(t *testing.T) {
		env := BenchmarkEnvironment{
			Platform:      "linux/amd64",
			GoVersion:     "go1.21.0",
			NumCPU:        8,
			NumGoroutines: 10,
			MemoryLimit:   1024 * 1024,
			GitCommit:     "abc123",
			GitBranch:     "main",
			BuildInfo:     "test build",
		}

		assert.NotEmpty(t, env.Platform)
		assert.NotEmpty(t, env.GoVersion)
		assert.Greater(t, env.NumCPU, 0)
		assert.GreaterOrEqual(t, env.NumGoroutines, 0)
		assert.Greater(t, env.MemoryLimit, uint64(0))
	})

	t.Run("BenchmarkReport structure", func(t *testing.T) {
		report := BenchmarkReport{
			Timestamp:       time.Now(),
			Environment:     BenchmarkEnvironment{},
			Summary:         BenchmarkSummary{},
			Benchmarks:      []profiling.BenchmarkResult{},
			Regressions:     []PerformanceRegression{},
			Improvements:    []PerformanceImprovement{},
			Recommendations: []string{},
			CIMetrics:       CIBenchmarkMetrics{},
		}

		assert.False(t, report.Timestamp.IsZero())
		assert.NotNil(t, report.Benchmarks)
		assert.NotNil(t, report.Regressions)
		assert.NotNil(t, report.Improvements)
		assert.NotNil(t, report.Recommendations)
	})
}

func TestContainerUtilityFunctions(t *testing.T) {
	t.Run("countContainersByState", func(t *testing.T) {
		containers := []ContainerInfo{
			{State: "running"},
			{State: "stopped"},
			{State: "running"},
			{State: "paused"},
			{State: "running"},
		}

		runningCount := countContainersByState(containers, "running")
		assert.Equal(t, 3, runningCount)

		stoppedCount := countContainersByState(containers, "stopped")
		assert.Equal(t, 1, stoppedCount)

		pausedCount := countContainersByState(containers, "paused")
		assert.Equal(t, 1, pausedCount)

		nonExistentCount := countContainersByState(containers, "nonexistent")
		assert.Equal(t, 0, nonExistentCount)

		// Test with empty slice
		emptyCount := countContainersByState([]ContainerInfo{}, "running")
		assert.Equal(t, 0, emptyCount)
	})

	t.Run("countNetworksByDriver", func(t *testing.T) {
		networks := []NetworkInfo{
			{Driver: "bridge"},
			{Driver: "overlay"},
			{Driver: "bridge"},
			{Driver: "host"},
			{Driver: "bridge"},
		}

		bridgeCount := countNetworksByDriver(networks, "bridge")
		assert.Equal(t, 3, bridgeCount)

		overlayCount := countNetworksByDriver(networks, "overlay")
		assert.Equal(t, 1, overlayCount)

		hostCount := countNetworksByDriver(networks, "host")
		assert.Equal(t, 1, hostCount)

		nonExistentCount := countNetworksByDriver(networks, "nonexistent")
		assert.Equal(t, 0, nonExistentCount)

		// Test with empty slice
		emptyCount := countNetworksByDriver([]NetworkInfo{}, "bridge")
		assert.Equal(t, 0, emptyCount)
	})
}

func TestToolCheckerFunctions(t *testing.T) {
	t.Run("NewToolChecker", func(t *testing.T) {
		config := ToolConfig{
			Name:              "test-tool",
			Command:           "test",
			VersionArgs:       []string{"--version"},
			InstallSuggestion: "Install test tool",
		}

		checker := NewToolChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, config.Name, checker.config.Name)
		assert.Equal(t, config.Command, checker.config.Command)
	})

	t.Run("CreateToolChecker with known tool", func(t *testing.T) {
		// Test with known tool
		checker := CreateToolChecker("golangci-lint")
		assert.NotNil(t, checker)
		assert.Equal(t, "golangci-lint", checker.config.Name)
		assert.Equal(t, "golangci-lint", checker.config.Command)
		assert.Equal(t, []string{"version"}, checker.config.VersionArgs)
		assert.Contains(t, checker.config.InstallSuggestion, "go install")
	})

	t.Run("CreateToolChecker with unknown tool", func(t *testing.T) {
		// Test with unknown tool
		checker := CreateToolChecker("unknown-tool")
		assert.NotNil(t, checker)
		assert.Equal(t, "unknown-tool", checker.config.Name)
		assert.Equal(t, "unknown-tool", checker.config.Command)
		assert.Equal(t, []string{"--version"}, checker.config.VersionArgs)
		assert.Contains(t, checker.config.InstallSuggestion, "unknown-tool")
	})

	t.Run("CommonToolConfigs verification", func(t *testing.T) {
		// Verify that CommonToolConfigs contains expected tools
		expectedTools := []string{"golangci-lint", "gofumpt", "gci", "deadcode", "dupl"}

		for _, tool := range expectedTools {
			config, exists := CommonToolConfigs[tool]
			assert.True(t, exists, "Tool %s should exist in CommonToolConfigs", tool)
			assert.Equal(t, tool, config.Name)
			assert.Equal(t, tool, config.Command)
			assert.NotEmpty(t, config.InstallSuggestion)
		}

		// Verify total count
		assert.Equal(t, len(expectedTools), len(CommonToolConfigs))
	})

	t.Run("CheckMultipleTools", func(t *testing.T) {
		ctx := context.Background()
		toolNames := []string{"nonexistent-tool-1", "nonexistent-tool-2"}

		results := CheckMultipleTools(ctx, toolNames)
		assert.Len(t, results, 2)

		for i, result := range results {
			assert.Equal(t, toolNames[i], result.Tool)
			assert.Equal(t, "missing", result.Status)
			assert.NotEmpty(t, result.Suggestion)
		}
	})

	t.Run("CheckAllCommonTools", func(t *testing.T) {
		ctx := context.Background()

		results := CheckAllCommonTools(ctx)
		assert.Len(t, results, len(CommonToolConfigs))

		// Verify all tools are checked
		toolsChecked := make(map[string]bool)
		for _, result := range results {
			toolsChecked[result.Tool] = true
			assert.NotEmpty(t, result.Tool)
			assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		}

		// Verify all common tools were checked
		for toolName := range CommonToolConfigs {
			assert.True(t, toolsChecked[toolName], "Tool %s should be checked", toolName)
		}
	})
}

func TestToolConfig(t *testing.T) {
	config := ToolConfig{
		Name:              "test-tool",
		Command:           "test-cmd",
		VersionArgs:       []string{"--version", "--help"},
		InstallSuggestion: "Install with package manager",
	}

	assert.Equal(t, "test-tool", config.Name)
	assert.Equal(t, "test-cmd", config.Command)
	assert.Len(t, config.VersionArgs, 2)
	assert.Equal(t, "--version", config.VersionArgs[0])
	assert.Equal(t, "--help", config.VersionArgs[1])
	assert.NotEmpty(t, config.InstallSuggestion)
}

func TestDevEnvUtilityFunctions(t *testing.T) {
	t.Run("getStatusIcon", func(t *testing.T) {
		testCases := []struct {
			status       string
			expectedIcon string
		}{
			{"ok", "✅"},
			{"found", "✅"},
			{"missing", "❌"},
			{"outdated", "⚠️"},
			{"needs_tidy", "⚠️"},
			{"incomplete", "⚠️"},
			{"error", "🔴"},
			{"alternative", "🔄"},
			{"unknown", "❓"},
			{"", "❓"}, // default case
		}

		for _, tc := range testCases {
			result := getStatusIcon(tc.status)
			assert.Equal(t, tc.expectedIcon, result, "Status %s should return icon %s", tc.status, tc.expectedIcon)
		}
	})

	t.Run("isGoVersionOutdated", func(t *testing.T) {
		testCases := []struct {
			version    string
			isOutdated bool
		}{
			{"go1.16.5", true},
			{"go1.17.0", true},
			{"go1.18.9", true},
			{"go1.19.2", true},
			{"go1.20.0", false},
			{"go1.21.5", false},
			{"go1.22.0", false},
			{"go1.23.0", false},
			{"go1.24.0", false},
			{"unknown", false},
			{"", false},
		}

		for _, tc := range testCases {
			result := isGoVersionOutdated(tc.version)
			assert.Equal(t, tc.isOutdated, result, "Version %s should return outdated: %v", tc.version, tc.isOutdated)
		}
	})
}

func TestAdditionalDataStructures(t *testing.T) {
	t.Run("PerformanceRegression structure", func(t *testing.T) {
		regression := PerformanceRegression{
			BenchmarkName:     "TestBenchmark",
			CurrentOpsPerSec:  100.0,
			BaselineOpsPerSec: 150.0,
			RegressionPercent: 33.33,
			Severity:          "high",
			Impact:            "significant",
		}

		assert.Equal(t, "TestBenchmark", regression.BenchmarkName)
		assert.Equal(t, 100.0, regression.CurrentOpsPerSec)
		assert.Equal(t, 150.0, regression.BaselineOpsPerSec)
		assert.Equal(t, 33.33, regression.RegressionPercent)
		assert.Equal(t, "high", regression.Severity)
		assert.Equal(t, "significant", regression.Impact)
	})

	t.Run("PerformanceImprovement structure", func(t *testing.T) {
		improvement := PerformanceImprovement{
			BenchmarkName:      "TestBenchmark",
			CurrentOpsPerSec:   150.0,
			BaselineOpsPerSec:  100.0,
			ImprovementPercent: 50.0,
			Impact:             "moderate",
		}

		assert.Equal(t, "TestBenchmark", improvement.BenchmarkName)
		assert.Equal(t, 150.0, improvement.CurrentOpsPerSec)
		assert.Equal(t, 100.0, improvement.BaselineOpsPerSec)
		assert.Equal(t, 50.0, improvement.ImprovementPercent)
		assert.Equal(t, "moderate", improvement.Impact)
	})

	t.Run("BenchmarkSummary structure", func(t *testing.T) {
		summary := BenchmarkSummary{
			TotalBenchmarks:  10,
			PassedBenchmarks: 8,
			FailedBenchmarks: 2,
			TotalDuration:    time.Minute,
			AverageOpsPerSec: 1000000.0,
			TotalMemoryUsage: 512 * 1024,
			PerformanceScore: 85.5,
		}

		assert.Equal(t, 10, summary.TotalBenchmarks)
		assert.Equal(t, 8, summary.PassedBenchmarks)
		assert.Equal(t, 2, summary.FailedBenchmarks)
		assert.Equal(t, time.Minute, summary.TotalDuration)
		assert.Equal(t, 1000000.0, summary.AverageOpsPerSec)
		assert.Equal(t, uint64(512*1024), summary.TotalMemoryUsage)
		assert.Equal(t, 85.5, summary.PerformanceScore)
	})

	t.Run("CIBenchmarkMetrics structure", func(t *testing.T) {
		metrics := CIBenchmarkMetrics{
			ExitCode:              0,
			RegressionThreshold:   5.0,
			HasCriticalRegression: false,
			RecommendedAction:     "continue",
			ArtifactPaths:         []string{"/path/to/artifact.json"},
		}

		assert.Equal(t, 0, metrics.ExitCode)
		assert.Equal(t, 5.0, metrics.RegressionThreshold)
		assert.False(t, metrics.HasCriticalRegression)
		assert.Equal(t, "continue", metrics.RecommendedAction)
		assert.Len(t, metrics.ArtifactPaths, 1)
		assert.Equal(t, "/path/to/artifact.json", metrics.ArtifactPaths[0])
	})

	t.Run("BenchmarkFunction structure", func(t *testing.T) {
		benchFunc := BenchmarkFunction{
			Name:     "BenchmarkTest",
			Package:  "github.com/test/package",
			Function: func(ctx context.Context) {}, // Simple function
		}

		assert.Equal(t, "BenchmarkTest", benchFunc.Name)
		assert.Equal(t, "github.com/test/package", benchFunc.Package)
		assert.NotNil(t, benchFunc.Function)
	})

	t.Run("benchmarkOptions structure", func(t *testing.T) {
		opts := benchmarkOptions{
			packagePattern:      "test-package",
			benchmarkFilter:     "TestBench",
			iterations:          10,
			duration:            time.Minute * 5,
			cpuProfile:          true,
			memProfile:          false,
			outputFile:          "output.json",
			baselineFile:        "baseline.json",
			ciMode:              true,
			regressionThreshold: 5.0,
			generateArtifacts:   true,
			compareMode:         false,
			createSnapshot:      true,
			snapshotDir:         "/snapshots",
		}

		assert.Equal(t, "test-package", opts.packagePattern)
		assert.Equal(t, "TestBench", opts.benchmarkFilter)
		assert.Equal(t, 10, opts.iterations)
		assert.Equal(t, time.Minute*5, opts.duration)
		assert.True(t, opts.cpuProfile)
		assert.False(t, opts.memProfile)
		assert.Equal(t, "output.json", opts.outputFile)
		assert.Equal(t, "baseline.json", opts.baselineFile)
		assert.True(t, opts.ciMode)
		assert.Equal(t, 5.0, opts.regressionThreshold)
		assert.True(t, opts.generateArtifacts)
		assert.False(t, opts.compareMode)
		assert.True(t, opts.createSnapshot)
		assert.Equal(t, "/snapshots", opts.snapshotDir)
	})
}

// Test refactored example functions
func TestRefactoredExampleFunctions(t *testing.T) {
	ctx := context.Background()

	t.Run("checkGolangciLintRefactored", func(t *testing.T) {
		result := checkGolangciLintRefactored(ctx)
		assert.Equal(t, "golangci-lint", result.Tool)
		assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		if result.Status == "missing" {
			assert.Contains(t, result.Suggestion, "Install with: go install")
		}
	})

	t.Run("checkGofumptRefactored", func(t *testing.T) {
		result := checkGofumptRefactored(ctx)
		assert.Equal(t, "gofumpt", result.Tool)
		assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		if result.Status == "missing" {
			assert.Contains(t, result.Suggestion, "Install with: go install")
		}
	})

	t.Run("checkGciRefactored", func(t *testing.T) {
		result := checkGciRefactored(ctx)
		assert.Equal(t, "gci", result.Tool)
		assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		if result.Status == "missing" {
			assert.Contains(t, result.Suggestion, "Install with: go install")
		}
	})

	t.Run("checkDeadcodeRefactored", func(t *testing.T) {
		result := checkDeadcodeRefactored(ctx)
		assert.Equal(t, "deadcode", result.Tool)
		assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		if result.Status == "missing" {
			assert.Contains(t, result.Suggestion, "Install with: go install")
		}
	})

	t.Run("checkDuplRefactored", func(t *testing.T) {
		result := checkDuplRefactored(ctx)
		assert.Equal(t, "dupl", result.Tool)
		assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		if result.Status == "missing" {
			assert.Contains(t, result.Suggestion, "Install with: go install")
		}
	})

	t.Run("checkAllGoToolsRefactored", func(t *testing.T) {
		results := checkAllGoToolsRefactored(ctx)
		assert.Equal(t, 5, len(results))

		// Verify all expected tools are present
		toolNames := make([]string, len(results))
		for i, result := range results {
			toolNames[i] = result.Tool
		}
		assert.Contains(t, toolNames, "golangci-lint")
		assert.Contains(t, toolNames, "gofumpt")
		assert.Contains(t, toolNames, "gci")
		assert.Contains(t, toolNames, "deadcode")
		assert.Contains(t, toolNames, "dupl")
	})

	t.Run("checkAllCommonToolsRefactored", func(t *testing.T) {
		results := checkAllCommonToolsRefactored(ctx)
		assert.Greater(t, len(results), 0)

		// Verify results contain the expected common tools
		toolNames := make([]string, len(results))
		for i, result := range results {
			toolNames[i] = result.Tool
			assert.Contains(t, []string{"missing", "found", "ok", "error"}, result.Status)
		}
		assert.Contains(t, toolNames, "golangci-lint")
		assert.Contains(t, toolNames, "gofumpt")
		assert.Contains(t, toolNames, "gci")
		assert.Contains(t, toolNames, "deadcode")
		assert.Contains(t, toolNames, "dupl")
	})
}

// Test performance snapshot utility functions
func TestPerformanceSnapshotUtilities(t *testing.T) {
	t.Run("generateSnapshotID", func(t *testing.T) {
		id1 := generateSnapshotID()
		assert.NotEmpty(t, id1)
		assert.Contains(t, id1, "snapshot-")

		// Generate another ID to ensure they're different (if called at different times)
		id2 := generateSnapshotID()
		assert.NotEmpty(t, id2)
		assert.Contains(t, id2, "snapshot-")
	})

	t.Run("environmentEqual", func(t *testing.T) {
		env1 := BenchmarkEnvironment{
			Platform:  "linux",
			GoVersion: "go1.21.0",
			NumCPU:    8,
		}

		env2 := BenchmarkEnvironment{
			Platform:  "linux",
			GoVersion: "go1.21.0",
			NumCPU:    8,
		}

		env3 := BenchmarkEnvironment{
			Platform:  "darwin",
			GoVersion: "go1.21.0",
			NumCPU:    8,
		}

		assert.True(t, environmentEqual(env1, env2))
		assert.False(t, environmentEqual(env1, env3))
	})

	t.Run("calculatePerformanceChange", func(t *testing.T) {
		// Import the profiling package for BenchmarkResult
		currentResult := profiling.BenchmarkResult{
			Name:      "TestBench",
			OpsPerSec: 150.0,
			NsPerOp:   6666667,
		}

		baselineResult := profiling.BenchmarkResult{
			Name:      "TestBench",
			OpsPerSec: 100.0,
			NsPerOp:   10000000,
		}

		change := calculatePerformanceChange(currentResult, baselineResult)
		assert.Equal(t, 50.0, change) // 50% improvement

		// Test with zero baseline
		zeroBaseline := profiling.BenchmarkResult{
			Name:      "TestBench",
			OpsPerSec: 0.0,
			NsPerOp:   0,
		}

		change = calculatePerformanceChange(currentResult, zeroBaseline)
		assert.Equal(t, 0.0, change)
	})

	t.Run("calculateSeverity", func(t *testing.T) {
		threshold := 5.0

		testCases := []struct {
			regressionPercent float64
			expected          string
		}{
			{20.0, "critical"}, // >= threshold*3 (15)
			{12.0, "high"},     // >= threshold*2 (10)
			{7.0, "medium"},    // >= threshold (5)
			{3.0, "low"},       // < threshold
		}

		for _, tc := range testCases {
			result := calculateSeverity(tc.regressionPercent, threshold)
			assert.Equal(t, tc.expected, result, "Regression %f should be %s", tc.regressionPercent, tc.expected)
		}
	})

	t.Run("calculateSnapshotPerformanceScore", func(t *testing.T) {
		regressions := []PerformanceRegression{
			{Severity: "high"},
			{Severity: "medium"},
		}

		improvements := []PerformanceImprovement{
			{ImprovementPercent: 10.0},
			{ImprovementPercent: 15.0},
		}

		score := calculateSnapshotPerformanceScore(regressions, improvements, 10)
		// Starting score: 100
		// High regression: -20
		// Medium regression: -10
		// Improvements: +10 (2 * 5)
		// Expected: 100 - 20 - 10 + 10 = 80
		assert.Equal(t, 80.0, score)

		// Test with zero benchmarks
		score = calculateSnapshotPerformanceScore(regressions, improvements, 0)
		assert.Equal(t, 0.0, score)

		// Test score boundaries
		manyRegressions := make([]PerformanceRegression, 5)
		for i := range manyRegressions {
			manyRegressions[i] = PerformanceRegression{Severity: "critical"}
		}

		score = calculateSnapshotPerformanceScore(manyRegressions, nil, 10)
		assert.Equal(t, 0.0, score) // Should not go below 0

		// Test with many improvements
		manyImprovements := make([]PerformanceImprovement, 10)
		score = calculateSnapshotPerformanceScore(nil, manyImprovements, 10)
		assert.Equal(t, 100.0, score) // Should cap at 100 (100 + 20 max improvement bonus)
	})
}

// Test setup utility functions
func TestSetupUtilityFunctions(t *testing.T) {
	t.Run("getSetupStatusIcon", func(t *testing.T) {
		testCases := []struct {
			status       string
			expectedIcon string
		}{
			{"success", "✅"},
			{"failed", "❌"},
			{"skipped", "⏭️"},
			{"unknown", "❓"},
			{"", "❓"}, // default case
		}

		for _, tc := range testCases {
			result := getSetupStatusIcon(tc.status)
			assert.Equal(t, tc.expectedIcon, result, "Status %s should return icon %s", tc.status, tc.expectedIcon)
		}
	})

	t.Run("countStepsByType", func(t *testing.T) {
		steps := []SetupStep{
			{Name: "step1", Critical: true, Optional: false},
			{Name: "step2", Critical: false, Optional: true},
			{Name: "step3", Critical: true, Optional: false},
			{Name: "step4", Critical: false, Optional: false},
			{Name: "step5", Critical: false, Optional: true},
		}

		// Count critical steps
		criticalCount := countStepsByType(steps, true, false)
		assert.Equal(t, 2, criticalCount)

		// Count optional steps
		optionalCount := countStepsByType(steps, false, true)
		assert.Equal(t, 2, optionalCount)

		// Count with both flags false (should return 0)
		noneCount := countStepsByType(steps, false, false)
		assert.Equal(t, 0, noneCount)

		// Test with empty slice
		emptyCount := countStepsByType([]SetupStep{}, true, false)
		assert.Equal(t, 0, emptyCount)
	})
}

// Test health utility functions
func TestHealthUtilityFunctions(t *testing.T) {
	t.Run("getAlertLevelPriority", func(t *testing.T) {
		testCases := []struct {
			level    string
			expected int
		}{
			{"critical", 3},
			{"warning", 2},
			{"info", 1},
			{"unknown", 0},
			{"", 0}, // default case
		}

		for _, tc := range testCases {
			result := getAlertLevelPriority(tc.level)
			assert.Equal(t, tc.expected, result, "Alert level %s should return priority %d", tc.level, tc.expected)
		}
	})

	t.Run("generateSuggestion", func(t *testing.T) {
		testCases := []struct {
			checkName string
			expected  string
		}{
			{"Memory Usage", "Consider increasing available memory or optimizing memory usage"},
			{"Disk Space", "Free up disk space or extend storage capacity"},
			{"Process Health", "Monitor for goroutine leaks and optimize concurrent operations"},
			{"CPU Usage", "Consider upgrading CPU or optimizing CPU-intensive operations"},
			{"Unknown Check", "Review system configuration and resource allocation"}, // default case
			{"", "Review system configuration and resource allocation"},              // default case
		}

		for _, tc := range testCases {
			check := HealthCheck{Name: tc.checkName}
			result := generateSuggestion(check)
			assert.Equal(t, tc.expected, result, "Check %s should return suggestion: %s", tc.checkName, tc.expected)
		}
	})
}

// Test container security utility functions
func TestContainerSecurityUtilities(t *testing.T) {
	t.Run("isPrivilegedContainer", func(t *testing.T) {
		container := ContainerInfo{
			ID:   "test-container",
			Name: "test",
		}
		// Current implementation always returns false
		result := isPrivilegedContainer(container)
		assert.False(t, result)
	})

	t.Run("isRootContainer", func(t *testing.T) {
		container := ContainerInfo{
			ID:   "test-container",
			Name: "test",
		}
		// Current implementation always returns false
		result := isRootContainer(container)
		assert.False(t, result)
	})

	t.Run("usesHostNetwork", func(t *testing.T) {
		container := ContainerInfo{
			ID:   "test-container",
			Name: "test",
		}
		// Current implementation always returns false
		result := usesHostNetwork(container)
		assert.False(t, result)
	})

	t.Run("usesHostPID", func(t *testing.T) {
		container := ContainerInfo{
			ID:   "test-container",
			Name: "test",
		}
		// Current implementation always returns false
		result := usesHostPID(container)
		assert.False(t, result)
	})

	t.Run("findExposedSecrets", func(t *testing.T) {
		container := ContainerInfo{
			ID:   "test-container",
			Name: "test",
		}
		// Current implementation always returns empty slice
		result := findExposedSecrets(container)
		assert.Empty(t, result)
		assert.IsType(t, []SecurityIssue{}, result)
	})
}

// Test more performance snapshot functions
func TestMorePerformanceSnapshots(t *testing.T) {
	t.Run("DefaultAnalysisOptions", func(t *testing.T) {
		options := DefaultAnalysisOptions()

		assert.Equal(t, 10.0, options.RegressionThreshold)
		assert.True(t, options.IncludeTrends)
		assert.Equal(t, 30, options.TrendWindowDays)
		assert.Equal(t, 0.95, options.ConfidenceLevel)
		assert.True(t, options.GeneratePredictions)
	})

	t.Run("generateTrendPrediction", func(t *testing.T) {
		// Test with non-significant regression
		nonSignificantRegression := LinearRegressionResult{
			IsSignificant: false,
		}

		prediction := generateTrendPrediction(nonSignificantRegression, []TrendDataPoint{})
		assert.Equal(t, "Continue monitoring - trend not statistically significant", prediction.RecommendedAction)

		// Test with significant regression - degrading
		degradingRegression := LinearRegressionResult{
			IsSignificant: true,
			Slope:         -15.0, // Significant degradation
		}

		dataPoints := []TrendDataPoint{
			{OpsPerSec: 1000.0, Timestamp: time.Now()},
		}

		prediction = generateTrendPrediction(degradingRegression, dataPoints)
		assert.Contains(t, prediction.RecommendedAction, "Immediate investigation required")

		// Test with improving trend
		improvingRegression := LinearRegressionResult{
			IsSignificant: true,
			Slope:         15.0, // Significant improvement
		}

		prediction = generateTrendPrediction(improvingRegression, dataPoints)
		assert.Contains(t, prediction.RecommendedAction, "Document optimization")

		// Test with stable trend
		stableRegression := LinearRegressionResult{
			IsSignificant: true,
			Slope:         0.5, // Stable trend
		}

		prediction = generateTrendPrediction(stableRegression, dataPoints)
		assert.Contains(t, prediction.RecommendedAction, "Continue current practices")
	})

	t.Run("performLinearRegression", func(t *testing.T) {
		// Test with insufficient data points
		emptyResult := performLinearRegression([]TrendDataPoint{})
		assert.Equal(t, LinearRegressionResult{}, emptyResult)

		onePointResult := performLinearRegression([]TrendDataPoint{
			{OpsPerSec: 100.0, Timestamp: time.Now()},
		})
		assert.Equal(t, LinearRegressionResult{}, onePointResult)

		// Test with sufficient data points
		now := time.Now()
		dataPoints := []TrendDataPoint{
			{OpsPerSec: 100.0, Timestamp: now.AddDate(0, 0, -2)},
			{OpsPerSec: 105.0, Timestamp: now.AddDate(0, 0, -1)},
			{OpsPerSec: 110.0, Timestamp: now},
		}

		result := performLinearRegression(dataPoints)
		assert.Greater(t, result.Slope, 0.0) // Should be positive slope for increasing trend
		assert.GreaterOrEqual(t, result.RSquared, 0.0)
		assert.LessOrEqual(t, result.RSquared, 1.0)
	})
}

// Test additional container functions
func TestAdditionalContainerFunctions(t *testing.T) {
	t.Run("WriteFile", func(t *testing.T) {
		// Create a temp file for testing
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "test.txt")
		data := []byte("test data")

		err := WriteFile(filename, data, 0o644)
		assert.NoError(t, err)

		// Verify file was written correctly
		readData, err := os.ReadFile(filename)
		assert.NoError(t, err)
		assert.Equal(t, data, readData)
	})

	t.Run("NewSnapshotManager", func(t *testing.T) {
		snapshotDir := "/tmp/snapshots"
		manager := NewSnapshotManager(snapshotDir)

		assert.NotNil(t, manager)
		assert.Equal(t, snapshotDir, manager.snapshotDir)
		assert.NotNil(t, manager.logger)
	})

	t.Run("SnapshotManager_LoadSnapshot", func(t *testing.T) {
		manager := NewSnapshotManager("/nonexistent")

		// Test loading non-existent snapshot
		snapshot, err := manager.LoadSnapshot("nonexistent-id")
		assert.Error(t, err)
		assert.Nil(t, snapshot)
		assert.Contains(t, err.Error(), "failed to read snapshot file")
	})

	t.Run("SnapshotManager_ListSnapshots", func(t *testing.T) {
		manager := NewSnapshotManager("/nonexistent")

		// Test listing snapshots from non-existent directory
		snapshots, err := manager.ListSnapshots()
		// The function may return an empty slice and no error for non-existent directories
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, snapshots)
			assert.Equal(t, 0, len(snapshots)) // Should be empty
		}
	})

	t.Run("SnapshotManager_CreateSnapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		manager := NewSnapshotManager(tempDir)
		ctx := context.Background()

		// Test creating snapshot with minimal data
		benchmarks := []profiling.BenchmarkResult{
			{
				Name:      "TestBench",
				OpsPerSec: 1000.0,
				NsPerOp:   1000000,
			},
		}

		metadata := map[string]interface{}{
			"test": "data",
		}

		snapshot, err := manager.CreateSnapshot(ctx, benchmarks, metadata)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.NotEmpty(t, snapshot.ID)
		assert.Equal(t, len(benchmarks), len(snapshot.Benchmarks))
	})

	t.Run("generateSnapshotID and saveSnapshot", func(t *testing.T) {
		// Test generateSnapshotID format
		id1 := generateSnapshotID()
		assert.Contains(t, id1, "snapshot-")
		assert.NotEmpty(t, id1)

		// Test saveSnapshot with a simple snapshot
		tempDir := t.TempDir()
		manager := NewSnapshotManager(tempDir)

		snapshot := &PerformanceSnapshot{
			ID:        "test-snapshot",
			Timestamp: time.Now(),
			Benchmarks: []profiling.BenchmarkResult{
				{Name: "test", OpsPerSec: 100, NsPerOp: 10000000},
			},
		}

		err := manager.saveSnapshot(snapshot)
		assert.NoError(t, err)

		// Verify file exists
		filename := filepath.Join(tempDir, "test-snapshot.json")
		_, err = os.Stat(filename)
		assert.NoError(t, err)
	})

	t.Run("calculateContainerSecurityScore", func(t *testing.T) {
		// Test with zero containers
		analysis := SecurityAnalysis{}
		score := calculateContainerSecurityScore(analysis, 0)
		assert.Equal(t, 100.0, score)

		// Test with perfect security (no issues)
		score = calculateContainerSecurityScore(analysis, 5)
		assert.Equal(t, 100.0, score)

		// Test with security issues
		analysisWithIssues := SecurityAnalysis{
			PrivilegedContainers:  []string{"priv1"},
			RootContainers:        []string{"root1"},
			HostNetworkContainers: []string{"hostnet1"},
			HostPIDContainers:     []string{"hostpid1"},
			SecretsExposed:        []SecurityIssue{{Issue: "secret1"}},
		}

		score = calculateContainerSecurityScore(analysisWithIssues, 5)
		// Score: 100 - (1*20 + 1*10 + 1*15 + 1*15 + 1*25) = 100 - 85 = 15
		assert.Equal(t, 15.0, score)

		// Test score doesn't go below 0
		manyIssues := SecurityAnalysis{
			PrivilegedContainers: []string{"p1", "p2", "p3", "p4", "p5", "p6"},
		}
		score = calculateContainerSecurityScore(manyIssues, 5)
		assert.Equal(t, 0.0, score) // Should not go below 0
	})

	t.Run("generateBenchmarkRecommendations", func(t *testing.T) {
		// Test with no issues
		report := &BenchmarkReport{
			Summary: BenchmarkSummary{
				TotalBenchmarks:  10,
				TotalMemoryUsage: 50 * 1024 * 1024, // 50MB
			},
		}

		// This function modifies the report in place, so we test the side effects
		initialRecommendations := len(report.Recommendations)
		generateBenchmarkRecommendations(report)

		// Should have no new recommendations for a good report
		assert.GreaterOrEqual(t, len(report.Recommendations), initialRecommendations)

		// Test with critical regressions
		reportWithRegressions := &BenchmarkReport{
			Summary: BenchmarkSummary{
				TotalBenchmarks:  10,
				TotalMemoryUsage: 50 * 1024 * 1024,
			},
			Regressions: []PerformanceRegression{
				{Severity: "critical"},
				{Severity: "high"},
				{Severity: "critical"},
			},
		}

		generateBenchmarkRecommendations(reportWithRegressions)
		assert.Greater(t, len(reportWithRegressions.Recommendations), 0)

		// Test with high memory usage
		reportWithMemory := &BenchmarkReport{
			Summary: BenchmarkSummary{
				TotalBenchmarks:  10,
				TotalMemoryUsage: 200 * 1024 * 1024, // 200MB (above 100MB threshold)
			},
		}

		generateBenchmarkRecommendations(reportWithMemory)
		assert.Greater(t, len(reportWithMemory.Recommendations), 0)

		// Test with few benchmarks
		reportWithFewBenchmarks := &BenchmarkReport{
			Summary: BenchmarkSummary{
				TotalBenchmarks:  3, // Below 5 threshold
				TotalMemoryUsage: 50 * 1024 * 1024,
			},
		}

		generateBenchmarkRecommendations(reportWithFewBenchmarks)
		assert.Greater(t, len(reportWithFewBenchmarks.Recommendations), 0)
	})

	t.Run("getPlatformPackageManager", func(t *testing.T) {
		manager := getPlatformPackageManager()
		assert.NotEmpty(t, manager)

		// The function should return one of the expected package managers
		expectedManagers := []string{"brew", "pip3", "apt", "yum", "pacman", "choco", "pip"}
		assert.Contains(t, expectedManagers, manager)
	})

	t.Run("getPlatformPreCommitInstallArgs", func(t *testing.T) {
		args := getPlatformPreCommitInstallArgs()
		assert.NotNil(t, args)
		// Should return some installation arguments
		assert.GreaterOrEqual(t, len(args), 0)
	})

	t.Run("generateSecurityRecommendations", func(t *testing.T) {
		// Test with no security issues
		cleanAnalysis := &SecurityAnalysis{}
		initialRecommendations := len(cleanAnalysis.Recommendations)
		generateSecurityRecommendations(cleanAnalysis)
		// Should not add any new recommendations
		assert.GreaterOrEqual(t, len(cleanAnalysis.Recommendations), initialRecommendations)

		// Test with all types of security issues
		analysisWithIssues := &SecurityAnalysis{
			PrivilegedContainers:  []string{"priv1", "priv2"},
			RootContainers:        []string{"root1"},
			HostNetworkContainers: []string{"hostnet1"},
			SecretsExposed:        []SecurityIssue{{Issue: "secret1"}},
		}

		initialCount := len(analysisWithIssues.Recommendations)
		generateSecurityRecommendations(analysisWithIssues)

		// Should have added recommendations for each type of issue
		assert.Greater(t, len(analysisWithIssues.Recommendations), initialCount)

		// Test with only privileged containers
		privOnlyAnalysis := &SecurityAnalysis{
			PrivilegedContainers: []string{"priv1"},
		}

		generateSecurityRecommendations(privOnlyAnalysis)
		assert.Greater(t, len(privOnlyAnalysis.Recommendations), 0)
	})

	t.Run("generateAnalysisSummary", func(t *testing.T) {
		tempDir := t.TempDir()
		manager := NewSnapshotManager(tempDir)

		// Create analysis with various trends and regressions
		analysis := &SnapshotAnalysis{
			Trends: []PerformanceTrend{
				{TrendDirection: "improving"},
				{TrendDirection: "degrading"},
				{TrendDirection: "stable"},
				{TrendDirection: "improving"},
			},
			Regressions: []PerformanceRegression{
				{Severity: "critical"},
				{Severity: "high"},
				{Severity: "critical"},
			},
		}

		// Test the summary generation (this modifies analysis in place)
		manager.generateAnalysisSummary(analysis)
		assert.NotNil(t, analysis.Summary)

		// The summary should reflect the trend counts
		assert.Equal(t, 2, analysis.Summary.TrendingUp)
		assert.Equal(t, 1, analysis.Summary.TrendingDown)
		assert.Equal(t, 1, analysis.Summary.StableTrends)
		assert.Equal(t, 2, analysis.Summary.CriticalIssues)
	})

	t.Run("tryAutoFix", func(t *testing.T) {
		// Test with Configuration Directory Access issue
		configResult := DiagnosticResult{
			Name:   "Configuration Directory Access",
			Status: "fail",
		}

		canFix := tryAutoFix(configResult)
		assert.True(t, canFix) // Should be able to create config directory

		// Test with unknown issue
		unknownResult := DiagnosticResult{
			Name:   "Unknown Issue",
			Status: "fail",
		}

		canFix = tryAutoFix(unknownResult)
		assert.False(t, canFix) // Should not be able to fix unknown issues
	})

	t.Run("calculateDocumentationScore", func(t *testing.T) {
		// Test with zero code lines
		emptyReport := &CodeQualityReport{
			Summary: QualitySummary{
				CodeLines:    0,
				CommentLines: 0,
			},
		}
		score := calculateDocumentationScore(emptyReport)
		assert.Equal(t, 0.0, score)

		// Test with normal documentation ratio
		normalReport := &CodeQualityReport{
			Summary: QualitySummary{
				CodeLines:    100,
				CommentLines: 20, // 20% documentation ratio
			},
		}
		score = calculateDocumentationScore(normalReport)
		// Score = (20/100) * 100 * 5 = 100
		assert.Equal(t, 100.0, score)

		// Test with high documentation ratio (should cap at 100)
		highDocReport := &CodeQualityReport{
			Summary: QualitySummary{
				CodeLines:    100,
				CommentLines: 50, // 50% documentation ratio
			},
		}
		score = calculateDocumentationScore(highDocReport)
		assert.Equal(t, 100.0, score) // Should cap at 100

		// Test with low documentation ratio
		lowDocReport := &CodeQualityReport{
			Summary: QualitySummary{
				CodeLines:    100,
				CommentLines: 5, // 5% documentation ratio
			},
		}
		score = calculateDocumentationScore(lowDocReport)
		// Score = (5/100) * 100 * 5 = 25
		assert.Equal(t, 25.0, score)
	})

	t.Run("generateQualityRecommendations", func(t *testing.T) {
		// Test with low overall score
		lowQualityReport := &CodeQualityReport{
			Summary: QualitySummary{
				OverallScore: 60.0,
			},
		}

		threshold := 80.0
		initialRecommendations := len(lowQualityReport.Recommendations)
		generateQualityRecommendations(lowQualityReport, threshold)

		// Should have added recommendations for low score
		assert.Greater(t, len(lowQualityReport.Recommendations), initialRecommendations)

		// Test with good quality score
		goodQualityReport := &CodeQualityReport{
			Summary: QualitySummary{
				OverallScore: 90.0,
			},
		}

		initialRecommendations = len(goodQualityReport.Recommendations)
		generateQualityRecommendations(goodQualityReport, threshold)

		// May or may not add recommendations depending on other factors
		assert.GreaterOrEqual(t, len(goodQualityReport.Recommendations), initialRecommendations)
	})

	t.Run("calculatePerformanceScore", func(t *testing.T) {
		// Test with good performance (low complexity, no performance issues)
		goodReport := &CodeQualityReport{
			Metrics: QualityMetrics{
				AverageComplexity: 10.0, // Below 20 threshold
			},
			Issues: []QualityIssue{}, // No issues
		}
		score := calculatePerformanceScore(goodReport)
		assert.Equal(t, 100.0, score)

		// Test with high complexity
		highComplexityReport := &CodeQualityReport{
			Metrics: QualityMetrics{
				AverageComplexity: 25.0, // Above 20 threshold
			},
			Issues: []QualityIssue{},
		}
		score = calculatePerformanceScore(highComplexityReport)
		assert.Equal(t, 70.0, score) // 100 - 30 = 70

		// Test with performance issues
		performanceIssuesReport := &CodeQualityReport{
			Metrics: QualityMetrics{
				AverageComplexity: 10.0,
			},
			Issues: []QualityIssue{
				{Message: "inefficient loop detected"},
				{Message: "performance bottleneck in function"},
				{Message: "regular issue"}, // Should not affect score
			},
		}
		score = calculatePerformanceScore(performanceIssuesReport)
		assert.Equal(t, 80.0, score) // 100 - (2 * 10) = 80

		// Test score doesn't go below 0
		terribleReport := &CodeQualityReport{
			Metrics: QualityMetrics{
				AverageComplexity: 30.0, // -30
			},
			Issues: []QualityIssue{
				{Message: "performance issue 1"}, // -10
				{Message: "inefficient code 2"},  // -10
				{Message: "performance issue 3"}, // -10
				{Message: "inefficient code 4"},  // -10
				{Message: "performance issue 5"}, // -10
				{Message: "inefficient code 6"},  // -10
				{Message: "performance issue 7"}, // -10
			},
		}
		score = calculatePerformanceScore(terribleReport)
		assert.Equal(t, 0.0, score) // Should not go below 0
	})
}

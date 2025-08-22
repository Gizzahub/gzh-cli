//nolint:testpackage // White-box testing needed for internal function access
package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

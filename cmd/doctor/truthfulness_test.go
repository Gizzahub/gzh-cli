// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 실측 backend가 없는 두 항목만 skip이며 실제 시스템·메모리·SSH 검사는 유지한다.
func TestDoctorUnmeasuredChecksPreserveRealChecks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	require.NoError(t, os.Mkdir(filepath.Join(home, ".ssh"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".ssh", "fixture.pub"), []byte("public fixture"), 0o600))
	report := &DiagnosticReport{}
	runSystemChecks(report, nil, nil)
	runSecurityChecks(report, nil, nil)
	byName := make(map[string]DiagnosticResult)
	for _, result := range report.Results {
		byName[result.Name] = result
	}
	require.Len(t, byName, 6)
	for _, name := range []string{"Disk Space", "File Permissions"} {
		result := byName[name]
		require.Equal(t, "skip", result.Status)
		require.Equal(t, false, result.Details["measured"])
		require.NotEmpty(t, result.Details["reason"])
		require.Contains(t, result.Message, "not")
		require.NotContains(t, result.Message, "50.00")
		require.NotContains(t, result.Message, "look secure")
		require.NotContains(t, result.Details, "unsafe_files")
		require.Empty(t, result.FixSuggestion, "unmeasured is not a detected repairable fault")
	}
	require.Equal(t, statusPass, byName["System Information"].Status)
	require.Contains(t, byName["Memory Usage"].Details, "memory_stats")
	require.Equal(t, statusPass, byName["SSH Keys"].Status)
	require.Equal(t, "SSH keys: 1 found", byName["SSH Keys"].Message)
	calculateReportTotals(report)
	require.Equal(t, 6, report.TotalChecks)
	require.Equal(t, 2, report.SkippedChecks)
	raw, err := json.Marshal(report)
	require.NoError(t, err)
	var saved DiagnosticReport
	require.NoError(t, json.Unmarshal(raw, &saved))
	for _, result := range saved.Results {
		if result.Status == "skip" {
			require.Equal(t, false, result.Details["measured"])
			require.NotEmpty(t, result.Details["reason"])
		}
	}
}

func TestDoctorSkippedSummaryDoesNotInventHealthVerdict(t *testing.T) {
	for _, counts := range [][3]int{{2, 0, 0}, {0, 0, 0}, {1, 1, 1}} {
		report := &DiagnosticReport{PassedChecks: counts[0], WarnChecks: counts[1], FailedChecks: counts[2], SkippedChecks: 2, TotalChecks: counts[0] + counts[1] + counts[2] + 2}
		generateSummaryAndRecommendations(report)
		require.Contains(t, report.Summary, "incomplete assessment: 2 checks were skipped")
		recommendations := strings.Join(report.Recommendations, "\n")
		require.Contains(t, recommendations, "not fully assessed")
		require.NotContains(t, recommendations, "System appears healthy")
		require.NotContains(t, recommendations, "System health is below optimal")
		if counts[2] > 0 {
			require.Contains(t, recommendations, "Address critical issues")
		}
		if counts[1] > 0 {
			require.Contains(t, recommendations, "Review warnings")
		}
	}
	report := &DiagnosticReport{PassedChecks: 2, SkippedChecks: 2, TotalChecks: 4}
	generateSummaryAndRecommendations(report)
	require.Contains(t, report.Summary, "50.0% success rate (2/4 checks passed)", "existing formula is unchanged")
}

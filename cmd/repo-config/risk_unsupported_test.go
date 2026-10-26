// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRiskDirectAssessmentUnsupported(t *testing.T) {
	for _, archived := range []bool{false, true} {
		result, err := performRiskAssessments("fixture-org", "fixture-token", archived)
		require.ErrorContains(t, err, "not supported")
		require.Nil(t, result)
	}
	finish := captureRiskStdout(t)
	require.ErrorContains(t, runRiskAssessmentCommand(GlobalFlags{Organization: "fixture", Verbose: true, DryRun: true}, "json", true, "high", filepath.Join(t.TempDir(), "absent.json"), 7), "not supported")
	require.Empty(t, finish())
	require.ErrorContains(t, runRiskAssessmentCommand(GlobalFlags{}, "json", false, "", "", 0), "organization is required")
}

// 공급된 fixture를 포맷·필터하는 유틸은 분석 생성기와 분리해 보존한다.
func TestRiskSuppliedFixtureUtilities(t *testing.T) {
	data := []RiskAssessment{{Repository: "provided-high", OverallScore: 8, Severity: "high", Categories: RiskCategories{AccessControl: 8}}, {Repository: "provided-low", OverallScore: 4, Severity: "low"}}
	require.Equal(t, data[:1], filterBySeverity(data, "high"))
	require.Equal(t, data[:1], filterByRiskThreshold(data, 8))
	require.Empty(t, filterByRiskThreshold(data, 9))
	summary := generateAssessmentSummary(data)
	require.Equal(t, 2, summary["total_repositories"])
	require.Equal(t, 6.0, summary["average_score"])
	require.Equal(t, map[string]int{"high": 1, "low": 1}, summary["severity_distribution"])
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "fixture.json")
	require.NoError(t, displayRiskAssessmentJSON(data, jsonPath))
	raw, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	var saved struct {
		Assessments []RiskAssessment `json:"assessments"`
	}
	require.NoError(t, json.Unmarshal(raw, &saved))
	require.Equal(t, data, saved.Assessments)
	csvPath := filepath.Join(dir, "fixture.csv")
	require.NoError(t, displayRiskAssessmentCSV(data, csvPath))
	raw, err = os.ReadFile(csvPath)
	require.NoError(t, err)
	rows, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 3)
	require.Equal(t, "provided-high", rows[1][0])
	require.Equal(t, "8.0", rows[1][1])
	htmlPath := filepath.Join(dir, "fixture.html")
	require.NoError(t, displayRiskAssessmentHTML(data, htmlPath, "supplied-org"))
	raw, err = os.ReadFile(htmlPath)
	require.NoError(t, err)
	require.Contains(t, string(raw), "provided-high")
	require.Contains(t, string(raw), "supplied-org")
	finish := captureRiskStdout(t)
	displayRiskAssessmentTable(data)
	displayRiskSummary(data, 7)
	rendered := finish()
	require.Contains(t, rendered, "provided-high")
	require.Contains(t, rendered, "6.0")
}

// 병렬 실행하지 않으며 fmt.Printf가 쓰는 실제 stdout도 관찰한다.
func captureRiskStdout(t *testing.T) func() string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	type captured struct {
		data []byte
		err  error
	}
	done := make(chan captured, 1)
	go func() { data, readErr := io.ReadAll(reader); done <- captured{data, readErr} }()
	finished := false
	var output string
	finish := func() string {
		if finished {
			return output
		}
		finished = true
		os.Stdout = original
		closeErr := writer.Close()
		result := <-done
		readerErr := reader.Close()
		output = string(result.data)
		require.NoError(t, closeErr)
		require.NoError(t, result.err)
		require.NoError(t, readerErr)
		return output
	}
	t.Cleanup(func() { finish() })
	return finish
}

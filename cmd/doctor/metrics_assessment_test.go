// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/internal/cli"
)

func TestMetricsAssessmentThresholdAcrossOutputs(t *testing.T) {
	for _, format := range []string{"table", "issues", "json", "yaml"} {
		for _, score := range []float64{74.99, 75, 75.01} {
			t.Run(format+"/"+fmtScore(score), func(t *testing.T) {
				report := &CodeQualityReport{HistoryAssessment: "not_assessed", Summary: QualitySummary{OverallScore: score}}
				output := filepath.Join(t.TempDir(), "report.json")
				var rendered bool
				err := finishQualityReport(report, metricsOptions{threshold: 75, outputFile: output}, func() error {
					rendered = true
					if format == "json" || format == "yaml" {
						return cli.NewOutputFormatterWithWriter(format, &bytes.Buffer{}).FormatOutput(report)
					}
					return displayQualityResults(report, metricsOptions{onlyIssues: format == "issues"})
				})
				require.True(t, rendered)
				if score < 75 {
					require.ErrorContains(t, err, "below threshold")
					require.ErrorContains(t, err, "74.99")
				} else {
					require.NoError(t, err)
				}
				raw, readErr := os.ReadFile(output)
				require.NoError(t, readErr)
				var payload map[string]any
				require.NoError(t, json.Unmarshal(raw, &payload))
				require.Equal(t, "not_assessed", payload["history_assessment"])
				require.Contains(t, payload, "trends")
			})
		}
	}
}

func fmtScore(score float64) string { return strconv.FormatFloat(score, 'f', 2, 64) }

func TestMetricsAssessmentErrorPrecedence(t *testing.T) {
	report := &CodeQualityReport{Summary: QualitySummary{OverallScore: 0}}
	called := false
	err := finishQualityReport(report, metricsOptions{threshold: 75, outputFile: filepath.Join(t.TempDir(), "absent", "report.json")}, func() error { called = true; return nil })
	require.ErrorContains(t, err, "failed to save report")
	require.False(t, called)
	sentinel := errors.New("renderer unavailable")
	require.ErrorIs(t, finishQualityReport(report, metricsOptions{threshold: 75}, func() error { return sentinel }), sentinel)
	for _, format := range []string{"table", "issues", "json", "yaml"} {
		err := finishQualityReport(report, metricsOptions{threshold: 75}, func() error {
			writer := assessmentErrorWriter{err: sentinel}
			if format == "json" || format == "yaml" {
				return cli.NewOutputFormatterWithWriter(format, writer).FormatOutput(report)
			}
			return displayQualityResultsWithWriter(report, metricsOptions{onlyIssues: format == "issues"}, writer)
		})
		require.ErrorContains(t, err, sentinel.Error())
		require.NotContains(t, err.Error(), "below threshold")
	}
	report.Summary.OverallScore = math.NaN()
	err = finishQualityReport(report, metricsOptions{threshold: 75}, func() error { return cli.NewOutputFormatterWithWriter("json", &bytes.Buffer{}).FormatOutput(report) })
	require.ErrorContains(t, err, "unsupported value")
}

type assessmentErrorWriter struct{ err error }

func (w assessmentErrorWriter) Write([]byte) (int, error) { return 0, w.err }

func TestMetricsHistoricalExplicitOnly(t *testing.T) {
	for _, value := range []string{"7", "30", "0"} {
		t.Run(value, func(t *testing.T) {
			cmd := newMetricsCmd()
			require.NoError(t, cmd.Flags().Set("path", filepath.Join(t.TempDir(), "absent")))
			require.NoError(t, cmd.Flags().Set("historical", value))
			require.ErrorContains(t, cmd.RunE(cmd, nil), "historical analysis is not supported")
		})
	}
	cmd := newMetricsCmd()
	require.False(t, cmd.Flags().Changed("historical"))
	days, err := cmd.Flags().GetInt("historical")
	require.NoError(t, err)
	require.Equal(t, 7, days)
	// 기본 실행이 사전 이력 거절에 걸리지 않고 원래 분석 경로로 진입하는지 확인한다.
	require.NoError(t, cmd.Flags().Set("path", filepath.Join(t.TempDir(), "absent")))
	require.ErrorContains(t, cmd.RunE(cmd, nil), "failed to open project root")
}

func TestMetricsAssessmentActualPipeline(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "fixture.go"), []byte("package fixture\n"), 0o600))
	for _, format := range []string{"table", "issues", "json", "yaml"} {
		output := filepath.Join(t.TempDir(), "report.json")
		calls := 0
		var rendered bytes.Buffer
		deps := metricsDependencies{output: &rendered, remove: os.Remove, run: func(_ context.Context, _, exe string, args ...string) ([]byte, error) {
			calls++
			if exe == "golangci-lint" {
				return []byte(`{"Issues":[]}`), nil
			}
			if len(args) > 0 && args[0] == "tool" {
				return []byte("total: (statements) 0.0%\n"), nil
			}
			return nil, nil
		}}
		selectedFormat := format
		if format == "issues" {
			selectedFormat = "table"
		}
		err := runQualityMetricsAnalysisWithDependencies(context.Background(), &cli.CommonFlags{Format: selectedFormat}, metricsOptions{projectPath: root, threshold: 100, outputFile: output, onlyIssues: format == "issues"}, deps)
		require.ErrorContains(t, err, "below threshold")
		require.Equal(t, 3, calls)
		data, err := os.ReadFile(output)
		require.NoError(t, err)
		require.Contains(t, string(data), `"history_assessment": "not_assessed"`)
		if format == "json" || format == "yaml" {
			var payload map[string]any
			if format == "json" {
				require.NoError(t, json.Unmarshal(rendered.Bytes(), &payload))
			} else {
				require.NoError(t, yaml.Unmarshal(rendered.Bytes(), &payload))
			}
			require.Equal(t, "not_assessed", payload["history_assessment"])
		} else {
			require.Contains(t, rendered.String(), "Historical trends: not_assessed")
		}
	}
}

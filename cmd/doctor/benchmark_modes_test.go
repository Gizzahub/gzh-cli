// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/logger"
	"github.com/gizzahub/gzh-cli/internal/profiling"
)

// 실제 Cobra 경로가 모든 실행 모드에서 측정·파일 부작용 전에 실패하는지 확인한다.
func TestBenchmarkModesUnsupportedBeforeSideEffects(t *testing.T) {
	cases := [][]string{
		{},
		{"--package", "./internal/..."},
		{"--filter", "BenchmarkClone"},
		{"--compare", "--baseline", "baseline.json"},
		{"--create-snapshot"},
		{"--analyze-snapshots", "--snapshot-id", "fixture"},
		{"--trend-analysis"},
		{"--compare", "--analyze-snapshots", "--trend-analysis"},
		{"--cpu-profile"},
		{"--mem-profile"},
		{"--cpu-profile", "--mem-profile", "--artifacts"},
		{"--ci", "--format", "json"},
		{"--ci", "--format", "yaml"},
		{"--ci", "--format", "table"},
	}
	for _, args := range cases {
		t.Run(testBenchmarkModeName(args), func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("USERPROFILE", root)
			path := filepath.Join(root, "empty-path")
			require.NoError(t, os.Mkdir(path, 0o700))
			t.Setenv("PATH", path)
			original := []byte("preserved fixture bytes")
			for _, file := range []string{"existing.json", "baseline.json"} {
				require.NoError(t, os.WriteFile(file, original, 0o600))
			}
			command := newBenchmarkCmd()
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)
			command.SilenceUsage, command.SilenceErrors = true, true
			argv := append([]string{}, args...)
			argv = append(argv, "--output", "existing.json", "--snapshot-dir", "snapshots")
			command.SetArgs(argv)
			require.ErrorContains(t, command.Execute(), "doctor benchmark execution is not supported")
			for _, file := range []string{"existing.json", "baseline.json"} {
				got, err := os.ReadFile(file)
				require.NoError(t, err)
				require.Equal(t, original, got)
			}
			for _, dir := range []string{"tmp", "benchmark-artifacts", "snapshots"} {
				_, err := os.Lstat(dir)
				require.True(t, os.IsNotExist(err), "unexpected artifact %s: %v", dir, err)
			}
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Len(t, entries, 3)
		})
	}
}

func testBenchmarkModeName(args []string) string {
	data, _ := json.Marshal(args)
	return string(data)
}

func TestPackageBenchmarkHelperRejectsFabricatedResults(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	report := &BenchmarkReport{}
	opts := benchmarkOptions{packagePattern: "./fixture/...", benchmarkFilter: "BenchmarkActual", iterations: 1, duration: time.Millisecond}
	require.ErrorContains(t, runPackageBenchmarks(context.Background(), nil, opts, report), "not supported")
	require.Empty(t, report.Benchmarks)
	require.Zero(t, report.Summary.PassedBenchmarks)
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestProfileArtifactHelperRejectsBeforeWriting(t *testing.T) {
	for _, opts := range []benchmarkOptions{{cpuProfile: true}, {memProfile: true}, {cpuProfile: true, memProfile: true}} {
		root := t.TempDir()
		t.Chdir(root)
		report := &BenchmarkReport{Timestamp: time.Now(), CIMetrics: CIBenchmarkMetrics{ArtifactPaths: []string{"previous-valid-artifact"}}}
		require.ErrorContains(t, generateBenchmarkArtifacts(nil, report, opts), "not supported")
		require.Equal(t, []string{"previous-valid-artifact"}, report.CIMetrics.ArtifactPaths)
		entries, err := os.ReadDir(root)
		require.NoError(t, err)
		require.Empty(t, entries)
	}
}

// 연결된 실제 JSON 보고서 저장 기능은 profile 요청 없이 유지한다.
func TestBenchmarkArtifactHelperWritesActualJSONReport(t *testing.T) {
	t.Chdir(t.TempDir())
	fixture := profiling.BenchmarkResult{Name: "supplied-test-fixture", Operations: 2, Duration: time.Second, OpsPerSec: 2}
	report := &BenchmarkReport{Timestamp: time.Now(), Benchmarks: []profiling.BenchmarkResult{fixture}}
	require.NoError(t, generateBenchmarkArtifacts(nil, report, benchmarkOptions{}))
	require.Len(t, report.CIMetrics.ArtifactPaths, 1)
	data, err := os.ReadFile(report.CIMetrics.ArtifactPaths[0])
	require.NoError(t, err)
	var saved BenchmarkReport
	require.NoError(t, json.Unmarshal(data, &saved))
	require.Equal(t, report.Benchmarks, saved.Benchmarks)
}

func TestBenchmarkUnsupportedHelpRemainsAvailable(t *testing.T) {
	command := newBenchmarkCmd()
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--help"})
	require.NoError(t, command.Execute())
	require.Contains(t, command.Long, "currently unsupported")
	for _, flag := range []string{"package", "filter", "iterations", "duration", "cpu-profile", "mem-profile", "output", "baseline", "ci", "regression-threshold", "artifacts", "compare", "create-snapshot", "snapshot-dir", "analyze-snapshots", "snapshot-id", "trend-analysis", "trend-window-days"} {
		require.NotNil(t, command.Flags().Lookup(flag))
	}
}

// 실제 저장 데이터 비교·분석 helper는 가상 CLI 실행과 분리해 유지한다.
func TestSavedBenchmarkAnalysisHelpersRemainFunctional(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	baseline := BenchmarkReport{Benchmarks: []profiling.BenchmarkResult{{Name: "fixture", OpsPerSec: 100}}}
	data, err := json.Marshal(baseline)
	require.NoError(t, err)
	baselineFile := filepath.Join(root, "baseline.json")
	require.NoError(t, os.WriteFile(baselineFile, data, 0o600))
	report := &BenchmarkReport{Timestamp: time.Now(), Benchmarks: []profiling.BenchmarkResult{{Name: "fixture", OpsPerSec: 60}}}
	require.NoError(t, compareAgainstBaseline(report, baselineFile, 10))
	require.Len(t, report.Regressions, 1)
	require.Equal(t, 40.0, report.Regressions[0].RegressionPercent)
	handleCIMode(report)
	require.Equal(t, 1, report.CIMetrics.ExitCode)
	require.NoError(t, displayBenchmarkResults(report, benchmarkOptions{}, nil))
	snapshot := PerformanceSnapshot{ID: "baseline", Timestamp: time.Now().Add(-time.Hour), Benchmarks: baseline.Benchmarks}
	data, err = json.Marshal(snapshot)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "baseline.json"), data, 0o600))
	manager := NewSnapshotManager(root)
	analysis, err := performSnapshotAnalysis(manager, report, benchmarkOptions{snapshotID: "baseline", regressionThreshold: 10}, logger.NewSimpleLogger("fixture"))
	require.NoError(t, err)
	require.Equal(t, report.Benchmarks, analysis.Current.Benchmarks)
	require.Len(t, analysis.Regressions, 1)
	require.NoError(t, displayBenchmarkResults(report, benchmarkOptions{}, analysis))
	displaySnapshotAnalysis(analysis)
}

// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/cli"
)

func TestMetricsProcessFixture(t *testing.T) {
	if os.Getenv("GZ_METRICS_FIXTURE") != "1" {
		return
	}
	switch os.Getenv("GZ_METRICS_MODE") {
	case "issue":
		fmt.Print(`{"Issues":[{"FromLinter":"gosec","Severity":"error","Text":"fixture issue","Pos":{"Filename":"fixture.go","Line":2}}]}`)
		os.Exit(1)
	case "fail":
		fmt.Fprint(os.Stderr, "controlled failure")
		os.Exit(2)
	case "wait":
		time.Sleep(30 * time.Second)
	case "observe":
		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(3)
		}
		fmt.Printf("%s\n%s", cwd, strings.Join(os.Args[3:], "|"))
	default:
		fmt.Print(`{"Issues":[]}`)
	}
	os.Exit(0)
}

func metricsFixtureRun(t *testing.T, mode string) metricsCommandRunner {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	return func(ctx context.Context, dir, _ string, args ...string) ([]byte, error) {
		commandArgs := append([]string{"-test.run=^TestMetricsProcessFixture$", "--"}, args...)
		cmd := exec.CommandContext(ctx, executable, commandArgs...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GZ_METRICS_FIXTURE=1", "GZ_METRICS_MODE="+mode)
		out, runErr := cmd.Output()
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		return out, runErr
	}
}

func TestMetricsLintIntegrity(t *testing.T) {
	root := t.TempDir()
	opts := metricsOptions{projectPath: root}
	for _, mode := range []string{"issue", "clean"} {
		t.Run(mode, func(t *testing.T) {
			report := &CodeQualityReport{}
			run := metricsFixtureRun(t, mode)
			err := collectLintingIssuesWithRunner(context.Background(), report, opts, func(ctx context.Context, dir, exe string, args ...string) ([]byte, error) {
				require.Equal(t, root, dir)
				require.Equal(t, "golangci-lint", exe)
				require.Equal(t, []string{"run", "--output.json.path", "stdout", "--show-stats=false", "--issues-exit-code=1", "./..."}, args)
				return run(ctx, dir, exe, args...)
			})
			require.NoError(t, err)
			if mode == "issue" {
				require.Len(t, report.Issues, 1)
				require.Equal(t, 1, report.Summary.CriticalIssues)
			} else {
				require.Empty(t, report.Issues)
			}
		})
	}
	sentinel := errors.New("spawn failure")
	for _, tc := range []struct {
		name, output string
		err          error
	}{
		{name: "spawn", err: sentinel},
		{name: "empty"},
		{name: "badJSON", output: "{"},
		{name: "mixedOutput", output: "text finding\n" + `{"Issues":[]}`},
		{name: "null", output: "null"},
		{name: "missing", output: "{}"},
		{name: "nullIssues", output: `{"Issues":null}`},
		{name: "wrongType", output: `{"Issues":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := collectLintingIssuesWithRunner(context.Background(), &CodeQualityReport{}, opts, func(context.Context, string, string, ...string) ([]byte, error) { return []byte(tc.output), tc.err })
			require.Error(t, err)
			if tc.err != nil {
				require.ErrorIs(t, err, sentinel)
			}
		})
	}
	require.Error(t, collectLintingIssuesWithRunner(context.Background(), &CodeQualityReport{}, opts, metricsFixtureRun(t, "fail")))
	for _, cancelFirst := range []bool{true, false} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		if cancelFirst {
			cancel()
		}
		err := collectLintingIssuesWithRunner(ctx, &CodeQualityReport{}, opts, metricsFixtureRun(t, "wait"))
		cancel()
		if cancelFirst {
			require.ErrorIs(t, err, context.Canceled)
		} else {
			require.ErrorIs(t, err, context.DeadlineExceeded)
		}
	}
}

func TestMetricsCoverageIntegrity(t *testing.T) {
	for _, failure := range []string{"none", "test", "cover", "parse", "cleanup", "test-cleanup", "cancel"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			old := filepath.Join(root, "coverage.tmp")
			require.NoError(t, os.WriteFile(old, []byte("user data"), 0o600))
			sentinel := errors.New("controlled " + failure)
			cleanupSentinel := errors.New("cleanup failure")
			calls := 0
			var owned string
			deps := metricsDependencies{remove: os.Remove}
			deps.run = func(ctx context.Context, dir, exe string, args ...string) ([]byte, error) {
				calls++
				require.Equal(t, root, dir)
				require.Equal(t, "go", exe)
				if calls == 1 {
					require.Equal(t, "test", args[0])
					require.Equal(t, "./...", args[2])
					profile := strings.TrimPrefix(args[1], "-coverprofile=")
					require.NotEqual(t, args[1], profile)
					require.True(t, filepath.IsAbs(profile))
					owned = profile
					require.NotEqual(t, old, owned)
					require.NoError(t, os.WriteFile(profile, []byte("partial or complete profile"), 0o600))
					if failure == "test" || failure == "test-cleanup" {
						return nil, sentinel
					}
					if failure == "cancel" {
						return nil, context.Canceled
					}
					return nil, nil
				}
				require.Equal(t, []string{"tool", "cover", "-func=" + owned}, args)
				if failure == "cover" {
					return nil, sentinel
				}
				if failure == "parse" {
					return []byte("bad total"), nil
				}
				return []byte("total: (statements) 87.5%\n"), nil
			}
			if failure == "cleanup" || failure == "test-cleanup" {
				deps.remove = func(path string) error { require.Equal(t, owned, path); return cleanupSentinel }
			}
			report := &CodeQualityReport{}
			err := collectCoverageMetricsWithDependencies(context.Background(), report, metricsOptions{projectPath: root}, deps)
			if failure == "none" {
				require.NoError(t, err)
				require.Equal(t, 87.5, report.Summary.TestCoverage)
			} else {
				require.Error(t, err)
			}
			if failure == "cleanup" || failure == "test-cleanup" {
				require.ErrorIs(t, err, cleanupSentinel)
				if failure == "test-cleanup" {
					require.ErrorIs(t, err, sentinel)
				}
				require.FileExists(t, owned)
				require.NoError(t, os.Remove(owned))
			} else {
				require.NoFileExists(t, owned)
			}
			data, readErr := os.ReadFile(old)
			require.NoError(t, readErr)
			require.Equal(t, "user data", string(data))
			entries, readErr := os.ReadDir(root)
			require.NoError(t, readErr)
			require.Len(t, entries, 1)
		})
	}
}

func TestMetricsCoverageTotals(t *testing.T) {
	for _, bad := range []string{"", "total:", "total: (statements) NaN%", "total: (statements) +Inf%", "total: (statements) 1e999%", "total: (statements) -1%", "total: (statements) 101%", "total: (statements) abc%", "total: (statements) 5", "total: (wrong) 5%", "total: (statements) 5%\ntotal: (statements) 6%"} {
		_, err := parseMetricsCoverage([]byte(bad))
		require.Error(t, err, bad)
	}
	for _, value := range []float64{0, 87.5, 100} {
		got, err := parseMetricsCoverage([]byte(fmt.Sprintf("total: (statements) %g%%", value)))
		require.NoError(t, err)
		require.Equal(t, value, got)
	}
}

func TestMetricsOuterFailureDoesNotSaveReport(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "fixture.go"), []byte("package fixture\n"), 0o600))
	output := filepath.Join(root, "report.json")
	require.NoError(t, os.WriteFile(output, []byte("keep"), 0o600))
	sentinel := errors.New("unavailable linter")
	calls := 0
	deps := metricsDependencies{run: func(context.Context, string, string, ...string) ([]byte, error) { calls++; return nil, sentinel }, remove: os.Remove}
	for _, format := range []string{"table", "json", "yaml"} {
		err := runQualityMetricsAnalysisWithDependencies(context.Background(), &cli.CommonFlags{Format: format}, metricsOptions{projectPath: root, outputFile: output}, deps)
		require.ErrorIs(t, err, sentinel)
		data, readErr := os.ReadFile(output)
		require.NoError(t, readErr)
		require.Equal(t, "keep", string(data))
	}
	require.Equal(t, 3, calls)
}

func TestMetricsCommandRunner(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	t.Setenv("GZ_METRICS_FIXTURE", "1")
	t.Setenv("GZ_METRICS_MODE", "observe")
	root := t.TempDir()
	output, err := runMetricsCommand(context.Background(), root, executable, "-test.run=^TestMetricsProcessFixture$", "--", "first", "second")
	require.NoError(t, err)
	require.Equal(t, root+"\nfirst|second", string(output))
	t.Setenv("GZ_METRICS_MODE", "wait")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = runMetricsCommand(ctx, root, executable, "-test.run=^TestMetricsProcessFixture$", "--")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	// 기존 helper 진입점의 시그니처와 취소 동작을 유지한다.
	ctx, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	require.Error(t, collectLintingIssues(ctx, &CodeQualityReport{}, metricsOptions{projectPath: root}))
	require.Error(t, collectCoverageMetrics(ctx, &CodeQualityReport{}, metricsOptions{projectPath: root}))
}

func TestMetricsCoverageFailureDoesNotSaveReport(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "report.json")
	require.NoError(t, os.WriteFile(output, []byte("keep"), 0o600))
	sentinel := errors.New("coverage unavailable")
	calls := 0
	deps := metricsDependencies{remove: os.Remove, run: func(_ context.Context, _, exe string, _ ...string) ([]byte, error) {
		calls++
		if exe == "golangci-lint" {
			return []byte(`{"Issues":[]}`), nil
		}
		return nil, sentinel
	}}
	require.ErrorIs(t, runQualityMetricsAnalysisWithDependencies(context.Background(), &cli.CommonFlags{Format: "json"}, metricsOptions{projectPath: root, outputFile: output}, deps), sentinel)
	require.Equal(t, 2, calls)
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

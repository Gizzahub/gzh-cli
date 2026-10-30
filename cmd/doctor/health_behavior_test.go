// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHealthChecksReportThresholdBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		check    HealthCheck
		status   string
		critical bool
	}{
		{name: "cpu low", check: checkCPUHealth(CPUMetrics{Count: 3}), status: "warning"},
		{name: "cpu sufficient", check: checkCPUHealth(CPUMetrics{Count: 4}), status: "healthy"},
		{name: "memory below warning", check: checkMemoryHealth(MemoryMetrics{UsedPercent: 69.9}), status: "healthy"},
		{name: "memory warning boundary", check: checkMemoryHealth(MemoryMetrics{UsedPercent: 70}), status: "warning"},
		{name: "memory critical boundary", check: checkMemoryHealth(MemoryMetrics{UsedPercent: 85}), status: "critical", critical: true},
		{name: "disk below warning", check: checkDiskHealth(DiskMetrics{UsedPercent: 74.9}), status: "healthy"},
		{name: "disk warning boundary", check: checkDiskHealth(DiskMetrics{UsedPercent: 75}), status: "warning"},
		{name: "disk critical boundary", check: checkDiskHealth(DiskMetrics{UsedPercent: 90}), status: "critical", critical: true},
		{name: "process below warning", check: checkProcessHealth(ProcessMetrics{Goroutines: 99}), status: "healthy"},
		{name: "process warning boundary", check: checkProcessHealth(ProcessMetrics{Goroutines: 100}), status: "warning"},
		{name: "process critical boundary", check: checkProcessHealth(ProcessMetrics{Goroutines: 1000}), status: "critical", critical: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.status, test.check.Status)
			require.Equal(t, test.critical, test.check.Critical)
			require.NotEmpty(t, test.check.Name)
			require.NotEmpty(t, test.check.Message)
			require.NotNil(t, test.check.Details)
		})
	}
}

func TestHealthReportAggregationFiltersAlertsAndRecommendsActions(t *testing.T) {
	report := &SystemHealthReport{
		Categories: []HealthCategory{
			{
				Name: "memory",
				Checks: []HealthCheck{
					checkMemoryHealth(MemoryMetrics{Used: 9, Total: 10, UsedPercent: 90}),
				},
			},
			{
				Name: "process",
				Checks: []HealthCheck{
					checkProcessHealth(ProcessMetrics{Goroutines: 150}),
				},
			},
		},
	}
	for i := range report.Categories {
		calculateCategoryHealth(&report.Categories[i])
	}
	calculateOverallHealth(report)
	generateHealthAlerts(report, "warning")
	generateHealthRecommendations(report)

	require.Equal(t, 65.0, report.Score)
	require.Equal(t, "critical", report.OverallStatus)
	require.Len(t, report.Alerts, 2)
	require.Equal(t, "critical", report.Alerts[0].Level)
	require.Equal(t, "memory", report.Alerts[0].Category)
	require.Equal(t, "Memory Usage: Memory usage: 90.0% (critical)", report.Alerts[0].Message)
	require.Equal(t, "Consider increasing available memory or optimizing memory usage", report.Alerts[0].Suggestion)
	require.Equal(t, "warning", report.Alerts[1].Level)
	require.Equal(t, "process", report.Alerts[1].Category)
	require.False(t, report.Alerts[0].Timestamp.IsZero())
	require.Contains(t, report.Recommendations, "Address 1 critical health alerts immediately")
	require.Contains(t, report.Recommendations, "Review 1 warning alerts for potential improvements")
	require.Contains(t, report.Recommendations, "Overall system health is below optimal - review critical issues")
	require.Contains(t, report.Recommendations, "Consider memory optimization or increasing available RAM")

	report.Alerts = nil
	generateHealthAlerts(report, "critical")
	require.Len(t, report.Alerts, 1)
	require.Equal(t, "critical", report.Alerts[0].Level)
	require.Equal(t, "memory", report.Alerts[0].Category)
}

func TestHealthCategoryScoreAndEmptyReportStates(t *testing.T) {
	category := HealthCategory{Checks: []HealthCheck{
		{Status: "warning"},
		{Status: "critical"},
		{Status: "critical"},
	}}
	calculateCategoryHealth(&category)
	require.Zero(t, category.Score)
	require.Equal(t, "critical", category.Status)

	category = HealthCategory{Checks: []HealthCheck{{Status: "critical"}, {Status: "critical"}}}
	calculateCategoryHealth(&category)
	require.Zero(t, category.Score, "scores are clamped at zero")
	require.Equal(t, "critical", category.Status)

	report := &SystemHealthReport{}
	calculateOverallHealth(report)
	require.Equal(t, "unknown", report.OverallStatus)
	require.Zero(t, report.Score)

	generateHealthRecommendations(report)
	require.Equal(t, []string{"Overall system health is below optimal - review critical issues"}, report.Recommendations)
}

func TestSaveHealthReportWritesPrivateJSONAndSurfacesFilesystemErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "health.json")
	report := &SystemHealthReport{
		OverallStatus: "warning",
		Score:         80,
		Categories:    []HealthCategory{{Name: "cpu", Status: "warning", Score: 80}},
		Timestamp:     time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, saveHealthReport(report, path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		// POSIX 권한 비트만 확인하며 Windows ACL을 증명하지 않는다.
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var saved SystemHealthReport
	require.NoError(t, json.Unmarshal(data, &saved))
	require.Equal(t, report.OverallStatus, saved.OverallStatus)
	require.Equal(t, report.Score, saved.Score)
	require.Equal(t, report.Categories, saved.Categories)
	require.Equal(t, report.Timestamp, saved.Timestamp)

	missingParent := filepath.Join(t.TempDir(), "absent", "health.json")
	require.Error(t, saveHealthReport(report, missingParent))
	_, err = os.Stat(missingParent)
	require.ErrorIs(t, err, os.ErrNotExist)
}

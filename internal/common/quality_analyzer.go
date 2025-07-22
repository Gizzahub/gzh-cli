// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package common provides shared quality analysis types and interfaces for repository analysis
package common

//go:generate mockgen -source=quality_analyzer.go -destination=mocks/quality_analyzer_mock.go -package=mocks QualityAnalyzer

import (
	"context"

	"go.uber.org/zap"
)

// Severity levels.
const (
	severityCritical = "critical"
	severityMajor    = "major"
	severityMinor    = "minor"
	severityInfo     = "info"
)

// QualityIssue represents a single quality issue found during analysis.
type QualityIssue struct {
	Type     string `json:"type"`     // "style", "bug", "security", "complexity", "type-error"
	Severity string `json:"severity"` // "critical", "major", "minor", "info"
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
	Tool     string `json:"tool"`
}

// QualityMetrics holds various quality metrics.
type QualityMetrics struct {
	TotalFiles       int     `json:"total_files"`
	TotalLinesOfCode int     `json:"total_lines_of_code"`
	AvgComplexity    float64 `json:"avg_complexity"`
	TestCoverage     float64 `json:"test_coverage"`
	DuplicationRate  float64 `json:"duplication_rate"`
}

// QualityResult represents the complete analysis result.
type QualityResult struct {
	Repository   string         `json:"repository"`
	Issues       []QualityIssue `json:"issues"`
	Metrics      QualityMetrics `json:"metrics"`
	OverallScore float64        `json:"overall_score"`
}

// FileMetrics holds file-level metrics.
type FileMetrics struct {
	TotalFiles int
	TotalLines int
}

// ComplexityMetrics holds complexity analysis results.
type ComplexityMetrics struct {
	AvgComplexity float64
}

// QualityAnalyzer defines the interface for language-specific quality analyzers.
type QualityAnalyzer interface {
	Name() string
	Language() string
	IsAvailable(ctx context.Context) bool
	Analyze(ctx context.Context, path string) (*QualityResult, error)
}

// BaseQualityAnalyzer provides common functionality for all analyzers.
type BaseQualityAnalyzer struct {
	logger *zap.Logger
}

// NewBaseQualityAnalyzer creates a new base analyzer.
func NewBaseQualityAnalyzer(logger *zap.Logger) *BaseQualityAnalyzer {
	return &BaseQualityAnalyzer{logger: logger}
}

// GetLogger returns the logger instance.
func (b *BaseQualityAnalyzer) GetLogger() *zap.Logger {
	return b.logger
}

// CalculateScore calculates quality score based on issues and metrics.
func (b *BaseQualityAnalyzer) CalculateScore(result *QualityResult) float64 {
	score := 100.0

	// Deduct points for issues based on severity
	for _, issue := range result.Issues {
		switch issue.Severity {
		case severityCritical:
			score -= 5.0
		case severityMajor:
			score -= 3.0
		case severityMinor:
			score -= 1.0
		case severityInfo:
			score -= 0.5
		}
	}

	// Factor in complexity
	if result.Metrics.AvgComplexity > 10 {
		score -= (result.Metrics.AvgComplexity - 10) * 2
	}

	// Factor in test coverage
	if result.Metrics.TestCoverage < 80 {
		score -= (80 - result.Metrics.TestCoverage) * 0.5
	}

	// Factor in duplication
	if result.Metrics.DuplicationRate > 5 {
		score -= result.Metrics.DuplicationRate * 0.5
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}

	if score > 100 {
		score = 100
	}

	return score
}

// MapSeverity maps tool-specific severity to standard severity.
func (b *BaseQualityAnalyzer) MapSeverity(severity string) string {
	switch severity {
	case "error", "fatal", "high":
		return "critical"
	case "warning", "medium":
		return "major"
	case "info", "low":
		return "minor"
	default:
		return "info"
	}
}

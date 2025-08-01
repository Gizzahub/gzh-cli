// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/analysis/godoc"
	"github.com/gizzahub/gzh-manager-go/internal/cli"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
)

// newGodocCmd creates the godoc subcommand for API documentation analysis.
func newGodocCmd() *cobra.Command {
	ctx := context.Background()

	var (
		packagePath     string
		allPackages     bool
		showCoverage    bool
		showMissing     bool
		showQuality     bool
		threshold       float64
		recommendations bool
		exportOnly      bool
	)

	cmd := cli.NewCommandBuilder(ctx, "godoc", "Analyze API documentation coverage and quality").
		WithLongDescription(`Analyze Go package documentation for coverage and quality issues.

This command provides comprehensive analysis of:
- Documentation coverage for exported symbols
- Missing documentation identification
- Code quality assessment
- Example function analysis
- Improvement recommendations

Examples:
  gz doctor godoc --package ./internal/logger
  gz doctor godoc --coverage --missing
  gz doctor godoc --all-packages --format json
  gz doctor godoc --threshold 80 --recommendations`).
		WithExample("gz doctor godoc --package ./internal/logger --coverage").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runGodocAnalysis(ctx, flags, godocOptions{
				packagePath:     packagePath,
				allPackages:     allPackages,
				showCoverage:    showCoverage,
				showMissing:     showMissing,
				showQuality:     showQuality,
				threshold:       threshold,
				recommendations: recommendations,
				exportOnly:      exportOnly,
			})
		}).
		Build()

	cmd.Flags().StringVar(&packagePath, "package", "", "Package path to analyze (relative to project root)")
	cmd.Flags().BoolVar(&allPackages, "all-packages", false, "Analyze all packages in the project")
	cmd.Flags().BoolVar(&showCoverage, "coverage", false, "Show detailed coverage statistics")
	cmd.Flags().BoolVar(&showMissing, "missing", false, "Show missing documentation items")
	cmd.Flags().BoolVar(&showQuality, "quality", false, "Show quality issues and suggestions")
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "Minimum coverage threshold (0-100)")
	cmd.Flags().BoolVar(&recommendations, "recommendations", false, "Show improvement recommendations")
	cmd.Flags().BoolVar(&exportOnly, "export-only", true, "Only analyze exported symbols")

	return cmd
}

type godocOptions struct {
	packagePath     string
	allPackages     bool
	showCoverage    bool
	showMissing     bool
	showQuality     bool
	threshold       float64
	recommendations bool
	exportOnly      bool
}

func runGodocAnalysis(ctx context.Context, flags *cli.CommonFlags, opts godocOptions) error {
	logger := logger.NewSimpleLogger("doctor-godoc")

	// Determine working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	analyzer := godoc.NewAnalyzer(workingDir)

	var packagesToAnalyze []string

	if opts.allPackages {
		// Find all Go packages in the project
		packagesToAnalyze, err = findGoPackages(workingDir)
		if err != nil {
			return fmt.Errorf("failed to find Go packages: %w", err)
		}
		logger.Info("Found packages", "count", len(packagesToAnalyze))
	} else if opts.packagePath != "" {
		packagesToAnalyze = []string{opts.packagePath}
	} else {
		// Default to current directory
		packagesToAnalyze = []string{"."}
	}

	results := make([]*godoc.PackageInfo, 0, len(packagesToAnalyze))
	failedPackages := make([]string, 0)

	// Analyze each package
	for _, pkgPath := range packagesToAnalyze {
		logger.Debug("Analyzing package", "path", pkgPath)

		pkgInfo, err := analyzer.AnalyzePackage(ctx, pkgPath)
		if err != nil {
			logger.Warn("Failed to analyze package", "path", pkgPath, "error", err)
			failedPackages = append(failedPackages, pkgPath)
			continue
		}

		// Apply threshold filter if specified
		if opts.threshold > 0 && pkgInfo.CoverageStats.CoveragePercentage < opts.threshold {
			logger.Debug("Package below threshold", "path", pkgPath,
				"coverage", pkgInfo.CoverageStats.CoveragePercentage, "threshold", opts.threshold)
		}

		results = append(results, pkgInfo)
	}

	if len(results) == 0 {
		return fmt.Errorf("no packages could be analyzed")
	}

	// Sort results by coverage percentage (lowest first for prioritization)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CoverageStats.CoveragePercentage < results[j].CoverageStats.CoveragePercentage
	})

	// Generate output based on format
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		// For structured formats, return the raw data
		output := map[string]interface{}{
			"packages":         results,
			"total_packages":   len(results),
			"failed_packages":  failedPackages,
			"analysis_options": opts,
		}
		return formatter.FormatOutput(output)

	default:
		// For table format, create custom output
		return displayGodocResults(results, failedPackages, opts)
	}
}

func findGoPackages(rootDir string) ([]string, error) {
	packages := make([]string, 0)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and vendor/node_modules
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
		}

		// Look for Go files in the directory
		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil // Skip directories we can't read
			}

			hasGoFiles := false
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
					hasGoFiles = true
					break
				}
			}

			if hasGoFiles {
				// Convert to relative path
				relPath, err := filepath.Rel(rootDir, path)
				if err != nil {
					return err
				}

				if relPath == "." {
					packages = append(packages, ".")
				} else {
					packages = append(packages, "./"+relPath)
				}
			}
		}

		return nil
	})

	return packages, err
}

func displayGodocResults(results []*godoc.PackageInfo, failedPackages []string, opts godocOptions) error {
	// Display summary
	totalSymbols := 0
	totalDocumented := 0
	totalIssues := 0

	for _, result := range results {
		totalSymbols += result.CoverageStats.TotalPublicSymbols
		totalDocumented += result.CoverageStats.DocumentedSymbols
		totalIssues += len(result.QualityIssues)
	}

	overallCoverage := 0.0
	if totalSymbols > 0 {
		overallCoverage = float64(totalDocumented) * 100.0 / float64(totalSymbols)
	}

	logger.SimpleInfo("📊 GoDoc Analysis Summary",
		"packages_analyzed", len(results),
		"failed_packages", len(failedPackages),
		"total_symbols", totalSymbols,
		"documented_symbols", totalDocumented,
		"overall_coverage", fmt.Sprintf("%.1f%%", overallCoverage),
		"quality_issues", totalIssues,
	)

	if len(failedPackages) > 0 {
		logger.SimpleWarn("❌ Failed to analyze packages:", "packages", strings.Join(failedPackages, ", "))
	}

	// Display per-package results
	for _, result := range results {
		displayPackageResults(result, opts)
	}

	// Display overall recommendations
	if opts.recommendations && len(results) > 0 {
		displayOverallRecommendations(results, overallCoverage, opts)
	}

	return nil
}

func displayPackageResults(result *godoc.PackageInfo, opts godocOptions) {
	// Package header
	statusIcon := "✅"
	if result.CoverageStats.CoveragePercentage < 70 {
		statusIcon = "⚠️"
	}
	if result.CoverageStats.CoveragePercentage < 50 {
		statusIcon = "❌"
	}

	logger.SimpleInfo(fmt.Sprintf("%s Package: %s", statusIcon, result.ImportPath),
		"coverage", fmt.Sprintf("%.1f%%", result.CoverageStats.CoveragePercentage),
		"symbols", result.CoverageStats.TotalPublicSymbols,
		"documented", result.CoverageStats.DocumentedSymbols,
		"issues", len(result.QualityIssues),
	)

	// Show detailed coverage if requested
	if opts.showCoverage {
		logger.SimpleInfo("  📈 Coverage Breakdown",
			"functions", fmt.Sprintf("%.1f%%", result.CoverageStats.FunctionCoverage),
			"types", fmt.Sprintf("%.1f%%", result.CoverageStats.TypeCoverage),
			"variables", fmt.Sprintf("%.1f%%", result.CoverageStats.VariableCoverage),
			"constants", fmt.Sprintf("%.1f%%", result.CoverageStats.ConstantCoverage),
			"examples", result.CoverageStats.ExampleCount,
		)
	}

	// Show missing documentation if requested
	if opts.showMissing {
		missing := make([]string, 0)

		for _, fn := range result.PublicFunctions {
			if fn.IsExported && !fn.HasDoc {
				missing = append(missing, fmt.Sprintf("func %s", fn.Name))
			}
		}

		for _, typ := range result.PublicTypes {
			if typ.IsExported && !typ.HasDoc {
				missing = append(missing, fmt.Sprintf("type %s", typ.Name))
			}

			for _, method := range typ.Methods {
				if method.IsExported && !method.HasDoc {
					missing = append(missing, fmt.Sprintf("method %s.%s", typ.Name, method.Name))
				}
			}
		}

		for _, v := range result.PublicVariables {
			if v.IsExported && !v.HasDoc {
				missing = append(missing, fmt.Sprintf("var %s", v.Name))
			}
		}

		for _, c := range result.PublicConstants {
			if c.IsExported && !c.HasDoc {
				missing = append(missing, fmt.Sprintf("const %s", c.Name))
			}
		}

		if len(missing) > 0 {
			logger.SimpleWarn("  📝 Missing Documentation", "items", strings.Join(missing, ", "))
		}
	}

	// Show quality issues if requested
	if opts.showQuality && len(result.QualityIssues) > 0 {
		for _, issue := range result.QualityIssues {
			severity := issue.Severity
			icon := "ℹ️"
			switch severity {
			case "high":
				icon = "🔴"
			case "medium":
				icon = "🟡"
			}

			logger.SimpleWarn(fmt.Sprintf("  %s %s", icon, issue.Message),
				"symbol", issue.Symbol,
				"line", issue.Line,
				"suggestion", issue.Suggestion,
			)
		}
	}

	// Show recommendations if requested
	if opts.recommendations && len(result.Recommendations) > 0 {
		for _, rec := range result.Recommendations {
			logger.SimpleInfo("  💡 Recommendation", "action", rec)
		}
	}
}

func displayOverallRecommendations(results []*godoc.PackageInfo, overallCoverage float64, opts godocOptions) {
	logger.SimpleInfo("🎯 Overall Recommendations")

	if overallCoverage < 80 {
		logger.SimpleInfo("  • Improve overall documentation coverage",
			"current", fmt.Sprintf("%.1f%%", overallCoverage),
			"target", "80%+")
	}

	// Find packages with lowest coverage
	if len(results) > 0 {
		worst := results[0] // Already sorted by coverage
		if worst.CoverageStats.CoveragePercentage < 60 {
			logger.SimpleInfo("  • Priority package for improvement",
				"package", worst.ImportPath,
				"coverage", fmt.Sprintf("%.1f%%", worst.CoverageStats.CoveragePercentage))
		}
	}

	// Count total high-severity issues
	highSeverityCount := 0
	for _, result := range results {
		for _, issue := range result.QualityIssues {
			if issue.Severity == "high" {
				highSeverityCount++
			}
		}
	}

	if highSeverityCount > 0 {
		logger.SimpleInfo("  • Address high-severity documentation issues",
			"count", highSeverityCount)
	}

	// Example recommendations
	totalFunctions := 0
	totalExamples := 0
	for _, result := range results {
		totalExamples += result.CoverageStats.ExampleCount
		for _, fn := range result.PublicFunctions {
			if fn.IsExported {
				totalFunctions++
			}
		}
	}

	if totalFunctions > 0 && float64(totalExamples)/float64(totalFunctions) < 0.3 {
		logger.SimpleInfo("  • Add more example functions",
			"current", totalExamples,
			"functions", totalFunctions,
			"recommended", totalFunctions/3)
	}
}

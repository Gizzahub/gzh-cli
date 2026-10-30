// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package godoc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzePackageSeparatesFactoriesAndReceiverMethods(t *testing.T) {
	root := t.TempDir()
	fixture := `// Package widgets describes widgets.
package widgets

// Widget is a sample widget.
type Widget struct{}

// NewWidget creates a pointer widget.
func NewWidget() *Widget { return &Widget{} }

func NewWidgetFrom(string) Widget { return Widget{} }

// Refresh checks the package operation.
func Refresh() {}

// Read returns widget data.
func (Widget) Read() string { return "data" }

func (*Widget) Write(string) {}
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "widgets.go"), []byte(fixture), 0o600))

	info, err := NewAnalyzer(root).AnalyzePackage(context.Background(), ".")
	require.NoError(t, err)

	functionCounts := make(map[string]int)
	for _, function := range info.PublicFunctions {
		functionCounts[function.Name]++
	}
	require.Equal(t, map[string]int{
		"NewWidget":     1,
		"NewWidgetFrom": 1,
		"Refresh":       1,
	}, functionCounts)

	require.Len(t, info.PublicTypes, 1)
	methods := make(map[string]FunctionInfo)
	for _, method := range info.PublicTypes[0].Methods {
		methods[method.Name] = method
	}
	require.Len(t, methods, 2)
	require.True(t, methods["Read"].IsExported)
	require.True(t, methods["Read"].HasDoc)
	require.True(t, methods["Write"].IsExported)
	require.False(t, methods["Write"].HasDoc)

	require.Equal(t, 6, info.CoverageStats.TotalPublicSymbols)
	require.Equal(t, 4, info.CoverageStats.DocumentedSymbols)
	require.Equal(t, 2, info.CoverageStats.UndocumentedSymbols)
	require.Equal(t, 60.0, info.CoverageStats.FunctionCoverage)
	issues := make(map[string]string)
	for _, issue := range info.QualityIssues {
		issues[issue.Type] = issue.Symbol
	}
	require.Equal(t, map[string]string{
		"missing_function_doc": "NewWidgetFrom",
		"missing_method_doc":   "Widget.Write",
	}, issues)
}

func TestCalculateCoverageStatsUsesExplicitFunctionTotals(t *testing.T) {
	analyzer := NewAnalyzer("")
	packageInfo := &PackageInfo{
		PublicFunctions: []FunctionInfo{
			{Name: "Documented", IsExported: true, HasDoc: true},
			{Name: "Missing", IsExported: true},
			{Name: "private", HasDoc: true},
		},
		PublicTypes: []TypeInfo{
			{
				Name: "PublicType", IsExported: true, HasDoc: true,
				Methods: []FunctionInfo{
					{Name: "Read", IsExported: true, HasDoc: true},
					{Name: "Write", IsExported: true},
					{Name: "privateMethod", HasDoc: true},
				},
			},
			{
				Name:    "privateType",
				Methods: []FunctionInfo{{Name: "ExportedMethod", IsExported: true, HasDoc: true}},
			},
		},
		PublicVariables: []VariableInfo{
			{Name: "DocumentedVar", IsExported: true, HasDoc: true},
			{Name: "MissingVar", IsExported: true},
			{Name: "privateVar", HasDoc: true},
		},
		PublicConstants: []ConstantInfo{
			{Name: "DocumentedConst", IsExported: true, HasDoc: true},
			{Name: "MissingConst", IsExported: true},
			{Name: "privateConst", HasDoc: true},
		},
		ExampleFunctions: []ExampleInfo{{Name: "ExampleDocumented"}, {Name: "ExampleMissing"}},
	}

	stats := analyzer.calculateCoverageStats(packageInfo)
	require.Equal(t, 9, stats.TotalPublicSymbols)
	require.Equal(t, 5, stats.DocumentedSymbols)
	require.Equal(t, 4, stats.UndocumentedSymbols)
	require.InDelta(t, 55.55555555555556, stats.CoveragePercentage, 0.000001)
	require.Equal(t, 50.0, stats.FunctionCoverage)
	require.Equal(t, 100.0, stats.TypeCoverage)
	require.Equal(t, 50.0, stats.VariableCoverage)
	require.Equal(t, 50.0, stats.ConstantCoverage)
	require.Equal(t, 2, stats.ExampleCount)
	require.Equal(t, 0.5, stats.ExamplesPerFunction)
}

func TestCalculateCoverageStatsLeavesFunctionMetricsZeroWithoutFunctions(t *testing.T) {
	packageInfo := &PackageInfo{
		PublicTypes:      []TypeInfo{{Name: "Widget", IsExported: true, HasDoc: true}},
		PublicVariables:  []VariableInfo{{Name: "Version", IsExported: true, HasDoc: true}},
		PublicConstants:  []ConstantInfo{{Name: "DefaultMode", IsExported: true, HasDoc: true}},
		ExampleFunctions: []ExampleInfo{{Name: "ExampleOne"}, {Name: "ExampleTwo"}},
	}

	stats := NewAnalyzer("").calculateCoverageStats(packageInfo)
	require.Equal(t, 3, stats.TotalPublicSymbols)
	require.Equal(t, 3, stats.DocumentedSymbols)
	require.Equal(t, 100.0, stats.CoveragePercentage)
	require.Zero(t, stats.FunctionCoverage)
	require.Zero(t, stats.ExamplesPerFunction)
	require.Equal(t, 2, stats.ExampleCount)
}

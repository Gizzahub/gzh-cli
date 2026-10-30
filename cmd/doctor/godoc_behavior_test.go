// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/analysis/godoc"
)

func TestFindGoPackagesIncludesSourceAndSkipsExcludedTrees(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, filepath.Join(root, "root.go"), "package root\n")
	writeGoFixture(t, filepath.Join(root, "internal", "child.go"), "package child\n")
	writeGoFixture(t, filepath.Join(root, "testonly", "only_test.go"), "package testonly\n")
	writeGoFixture(t, filepath.Join(root, ".cache", "hidden.go"), "package hidden\n")
	writeGoFixture(t, filepath.Join(root, "vendor", "dep", "vendor.go"), "package dep\n")
	writeGoFixture(t, filepath.Join(root, "node_modules", "dep", "module.go"), "package dep\n")

	packages, err := findGoPackages(root)
	require.NoError(t, err)
	require.Equal(t, []string{".", "./internal"}, packages)

	_, err = findGoPackages(filepath.Join(root, "missing"))
	require.Error(t, err)
}

func TestGodocAnalyzerReportsExportedDocumentation(t *testing.T) {
	root := t.TempDir()
	source := `// Package sample describes the public API.
package sample

// Public executes the documented path.
func Public() {}

func Missing() {}

func private() {}

// Tree represents a documented tree.
type Tree struct{}

// Open reads from a tree.
func (tree Tree) Open() {}

func (*Tree) Close() {}

// Count is a documented value.
var Count = 1
var MissingCount = 2

// Mode is a documented mode.
const Mode = "public"
const MissingMode = "missing"

`
	writeGoFixture(t, filepath.Join(root, "sample.go"), source)

	info, err := godoc.NewAnalyzer(root).AnalyzePackage(context.Background(), ".")
	require.NoError(t, err)
	require.Equal(t, "sample", info.Name)
	require.True(t, info.HasPackageDoc)
	require.Equal(t, []string{"sample.go"}, info.GoFiles)
	require.Equal(t, 9, info.CoverageStats.TotalPublicSymbols)
	require.Equal(t, 5, info.CoverageStats.DocumentedSymbols)
	require.Equal(t, 4, info.CoverageStats.UndocumentedSymbols)
	require.InDelta(t, 55.55555555555556, info.CoverageStats.CoveragePercentage, 0.000001)
	require.Zero(t, info.CoverageStats.ExampleCount)
	require.Empty(t, info.ExampleFunctions)

	functions := make(map[string]bool)
	for _, function := range info.PublicFunctions {
		functions[function.Name] = function.IsExported
	}
	require.True(t, functions["Public"])
	require.True(t, functions["Missing"])
	require.NotContains(t, functions, "Open")
	require.NotContains(t, functions, "Close")
	require.Contains(t, functions, "private")
	require.False(t, functions["private"])

	require.Len(t, info.PublicTypes, 1)
	require.Equal(t, "struct", info.PublicTypes[0].Kind)
	require.Len(t, info.PublicTypes[0].Methods, 2)
	methods := make(map[string]bool)
	for _, method := range info.PublicTypes[0].Methods {
		methods[method.Name] = method.HasDoc
	}
	require.Equal(t, map[string]bool{"Open": true, "Close": false}, methods)
	require.Equal(t, "missing_function_doc", info.QualityIssues[0].Type)
	require.Equal(t, "Missing", info.QualityIssues[0].Symbol)
	issues := make(map[string]string)
	for _, issue := range info.QualityIssues {
		issues[issue.Type] = issue.Symbol
	}
	require.Equal(t, map[string]string{
		"missing_function_doc": "Missing",
		"missing_method_doc":   "Tree.Close",
	}, issues)
	require.Contains(t, info.Recommendations, "Improve documentation coverage from 55.6% to at least 80%")
}

func TestGodocAnalyzerReturnsParseAndPackageErrors(t *testing.T) {
	root := t.TempDir()
	analyzer := godoc.NewAnalyzer(root)
	_, err := analyzer.AnalyzePackage(context.Background(), "./missing")
	require.ErrorContains(t, err, "package directory does not exist")

	writeGoFixture(t, filepath.Join(root, "broken.go"), "package broken\nfunc Broken( {\n")
	_, err = analyzer.AnalyzePackage(context.Background(), ".")
	require.ErrorContains(t, err, "failed to parse package")

	if err := os.Remove(filepath.Join(root, "broken.go")); err != nil {
		t.Fatal(err)
	}
	writeGoFixture(t, filepath.Join(root, "sample_test.go"), "package sample_test\n")
	_, err = analyzer.AnalyzePackage(context.Background(), ".")
	require.ErrorContains(t, err, "no non-test packages found")
}

func writeGoFixture(t *testing.T, path, source string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(source), 0o600))
}

// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package securityfinding

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type mapSourceResolver map[string]string

//go:embed testdata/gosec_sarif_fcdbaf25_excerpt.json
var historicalGosecSARIFExcerpt []byte

//go:embed testdata/gosec_sarif_b1a_excerpt.json
var currentGosecSARIFExcerpt []byte

func (resolver mapSourceResolver) ReadSource(path string) ([]byte, error) {
	contents, found := resolver[path]
	if !found {
		return nil, errors.New("source not found")
	}
	return []byte(contents), nil
}

func TestIngestSARIFConvertsGosecResultsAndCountsTopology(t *testing.T) {
	report, err := ingestSARIF(sarifDocumentWithResults(`[
		{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"test%2Ffoo.go"},"region":{"startLine":10,"startColumn":3,"snippet":{"text":"os.Open(name)"}}}}]},
		{"ruleId":"G304","message":{"text":" file  inclusion "},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"test%2Ffoo.go"},"region":{"startLine":10,"startColumn":3,"snippet":{"text":"os /* comment */ . Open(name)"}}}}]}
	]`), nil)
	require.NoError(t, err)
	require.Equal(t, 2, report.RawResultCount)
	require.Equal(t, 1, report.LocationUniqueCount)
	require.Len(t, report.Findings, 2)
	require.NotNil(t, report.Findings)
	require.Equal(t, "test%2Ffoo.go", report.Findings[0].Path)
	require.Equal(t, "os.Open(name)", report.Findings[0].SourceSnippet)
}

func TestIngestSARIFActualGosecReportExcerpts(t *testing.T) {
	for _, fixture := range []struct {
		name     string
		contents []byte
	}{
		{name: "fcdbaf25 excerpt", contents: historicalGosecSARIFExcerpt},
		{name: "B1a excerpt", contents: currentGosecSARIFExcerpt},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			report, err := ingestSARIF(fixture.contents, nil)
			require.NoError(t, err)
			require.Equal(t, 2, report.RawResultCount)
			require.Equal(t, 1, report.LocationUniqueCount)
			require.Len(t, report.Findings, 2)
			require.Equal(t, "test/e2e/helpers/filesystem.go", report.Findings[0].Path)
		})
	}
}

func TestIngestSARIFActualGosecReportExcerptsReconcile(t *testing.T) {
	baseline, err := ingestSARIF(historicalGosecSARIFExcerpt, nil)
	require.NoError(t, err)
	candidate, err := ingestSARIF(currentGosecSARIFExcerpt, nil)
	require.NoError(t, err)

	expected := expectation{
		BaselineSHA:    "fcdbaf253b61b4968297c3652aa04fe4ee4e8cb7",
		CandidateSHA:   "b43b053e8994e3e781d2c5c7c82baae4389a8a1d",
		ScannerVersion: "gosec-v2.28.0",
		ConfigHash:     "task116-fixture-config",
		FlagsHash:      "task116-fixture-flags",
	}
	result, err := compare(
		expected,
		&reportEnvelope{
			SourceSHA:      expected.BaselineSHA,
			ScannerVersion: expected.ScannerVersion,
			ConfigHash:     expected.ConfigHash,
			FlagsHash:      expected.FlagsHash,
			Findings:       baseline.Findings,
		},
		&reportEnvelope{
			SourceSHA:      expected.CandidateSHA,
			ScannerVersion: expected.ScannerVersion,
			ConfigHash:     expected.ConfigHash,
			FlagsHash:      expected.FlagsHash,
			Findings:       candidate.Findings,
		},
	)

	require.NoError(t, err)
	require.Equal(t, 2, baseline.RawResultCount)
	require.Equal(t, 1, baseline.LocationUniqueCount)
	require.Equal(t, 2, candidate.RawResultCount)
	require.Equal(t, 1, candidate.LocationUniqueCount)
	require.Equal(t, 2, result.BaselineRawCount)
	require.Equal(t, 2, result.CandidateRawCount)
	known, err := candidate.Findings[0].canonical()
	require.NoError(t, err)
	require.Equal(t, []finding{known}, result.Known)
	require.Empty(t, result.New)
	require.Empty(t, result.Unclassified)
}

func TestIngestSARIFAcceptsExplicitEmptyResults(t *testing.T) {
	report, err := ingestSARIF(sarifDocumentWithResults(`[]`), nil)
	require.NoError(t, err)
	require.NotNil(t, report.Findings)
	require.Empty(t, report.Findings)
	require.Zero(t, report.RawResultCount)
	require.Zero(t, report.LocationUniqueCount)
}

func TestIngestSARIFRejectsMissingNullAndInvalidRequiredStructures(t *testing.T) {
	for name, document := range map[string][]byte{
		"unsupported version":       []byte(`{"version":"2.0.0","runs":[]}`),
		"missing runs":              []byte(`{"version":"2.1.0"}`),
		"null runs":                 []byte(`{"version":"2.1.0","runs":null}`),
		"empty runs":                []byte(`{"version":"2.1.0","runs":[]}`),
		"multiple runs":             []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gosec"}},"results":[]},{"tool":{"driver":{"name":"gosec"}},"results":[]}]}`),
		"non gosec tool":            []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"other"}},"results":[]}]}`),
		"missing results":           []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gosec"}}}]}`),
		"null results":              []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gosec"}},"results":null}]}`),
		"non array results":         []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gosec"}},"results":{}}]}`),
		"missing message":           sarifDocumentWithResults(`[{"ruleId":"G304","locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":1,"startColumn":1,"snippet":{"text":"os.Open(name)"}}}}]}]`),
		"null locations":            sarifDocumentWithResults(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":null}]`),
		"missing physical location": sarifDocumentWithResults(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{}]}]`),
		"missing artifact location": sarifDocumentWithResults(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"region":{"startLine":1,"startColumn":1,"snippet":{"text":"os.Open(name)"}}}}]}]`),
		"missing region":            sarifDocumentWithResults(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"}}}]}]`),
		"non positive region":       sarifDocumentWithResults(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":0,"startColumn":1,"snippet":{"text":"os.Open(name)"}}}}]}]`),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ingestSARIF(document, nil)
			require.Error(t, err)
		})
	}
}

func TestIngestSARIFRejectsUnsafeArtifactURIs(t *testing.T) {
	for _, uri := range []string{
		"/absolute/main.go", "../outside.go", `\\server\share\main.go`,
		`cmd\test.go`, `C:\absolute\main.go`, "https://example.test/main.go", "cmd/\x00main.go",
		" cmd/test.go",
	} {
		t.Run(uri, func(t *testing.T) {
			_, err := ingestSARIF(sarifDocumentWithResults(resultWithURI(uri)), nil)
			require.Error(t, err)
		})
	}
}

func TestIngestSARIFResolvesMissingOrInvalidSnippet(t *testing.T) {
	resolver := mapSourceResolver{"cmd/app/main.go": "package app\n\tos.Open(name)\n"}
	for name, result := range map[string]string{
		"missing": `{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":2,"startColumn":2}}}]}`,
		"invalid": `{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":2,"startColumn":2,"snippet":{"text":"\"unterminated"}}}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			report, err := ingestSARIF(sarifDocumentWithResults("["+result+"]"), resolver)
			require.NoError(t, err)
			require.Equal(t, "\tos.Open(name)", report.Findings[0].SourceSnippet)
		})
	}
}

func TestIngestSARIFRejectsNULSnippet(t *testing.T) {
	_, err := ingestSARIF(sarifDocumentWithResults(`[
		{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":1,"startColumn":1,"snippet":{"text":"os\u0000.Open(name)"}}}}]}
	]`), nil)
	require.Error(t, err)
}

func TestIngestSARIFFailsClosedWhenSnippetCannotBeResolved(t *testing.T) {
	result := `[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":4,"startColumn":1}}}]}]`
	for name, resolver := range map[string]sourceResolver{
		"missing resolver": nil,
		"missing source":   mapSourceResolver{},
		"out of range":     mapSourceResolver{"cmd/app/main.go": "package app\n"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ingestSARIF(sarifDocumentWithResults(result), resolver)
			require.Error(t, err)
		})
	}
}

func TestIngestSARIFRejectsMultipleLocations(t *testing.T) {
	result := `[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[
		{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":1,"startColumn":1,"snippet":{"text":"os.Open(one)"}}}},
		{"physicalLocation":{"artifactLocation":{"uri":"cmd/app/main.go"},"region":{"startLine":2,"startColumn":1,"snippet":{"text":"os.Open(two)"}}}}
	]}]`
	_, err := ingestSARIF(sarifDocumentWithResults(result), nil)
	require.ErrorContains(t, err, "exactly one location")
}

func sarifDocumentWithResults(results string) []byte {
	return []byte(fmt.Sprintf(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gosec"}},"results":%s}]}`, results))
}

func resultWithURI(uri string) string {
	encodedURI, err := json.Marshal(uri)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`[{"ruleId":"G304","message":{"text":"file inclusion"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":%s},"region":{"startLine":1,"startColumn":1,"snippet":{"text":"os.Open(name)"}}}}]}]`, encodedURI)
}

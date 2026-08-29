package securityfinding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareFixtures(t *testing.T) {
	for _, testCase := range loadFixtures(t) {
		t.Run(testCase.Name, func(t *testing.T) {
			result, err := compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
			require.NoError(t, err)
			require.Equal(t, testCase.Want, result)
		})
	}
}

func TestCompareIsStableAcrossInputPermutation(t *testing.T) {
	testCase := loadFixture(t, "permutation remains stable")
	first, err := compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
	require.NoError(t, err)

	testCase.Baseline.Findings[0], testCase.Baseline.Findings[1] = testCase.Baseline.Findings[1], testCase.Baseline.Findings[0]
	testCase.Candidate.Findings[0], testCase.Candidate.Findings[1] = testCase.Candidate.Findings[1], testCase.Candidate.Findings[0]
	second, err := compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
	require.NoError(t, err)
	require.Equal(t, first, second)
}

func TestCompareRejectsTamperedEnvelopeMetadata(t *testing.T) {
	for _, field := range []string{"baseline SHA", "candidate SHA", "scanner version", "config hash", "flags hash"} {
		t.Run(field, func(t *testing.T) {
			testCase := loadFixture(t, "different revisions match trusted policy")
			switch field {
			case "baseline SHA":
				testCase.Baseline.SourceSHA = "tampered"
			case "candidate SHA":
				testCase.Candidate.SourceSHA = "tampered"
			case "scanner version":
				testCase.Candidate.ScannerVersion = "tampered"
			case "config hash":
				testCase.Candidate.ConfigHash = "tampered"
			case "flags hash":
				testCase.Candidate.FlagsHash = "tampered"
			}
			_, err := compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
			require.Error(t, err)
		})
	}
}

func TestCompareRejectsNilEnvelopeAndNilFindings(t *testing.T) {
	testCase := loadFixture(t, "different revisions match trusted policy")
	_, err := compare(testCase.Expected, nil, testCase.Candidate)
	require.ErrorContains(t, err, "baseline report is required")

	testCase.Candidate.Findings = nil
	_, err = compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
	require.ErrorContains(t, err, "candidate findings must be an explicit array")
}

func TestCompareRejectsMalformedFindings(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*reportEnvelope)
	}{
		{"missing rule", func(report *reportEnvelope) {
			report.Findings = []finding{{Path: "cmd/app/main.go", StartLine: 1, StartColumn: 1, Message: "message", SourceSnippet: "os.Open(name)"}}
		}},
		{"zero location", func(report *reportEnvelope) {
			report.Findings = []finding{{Rule: "G304", Path: "cmd/app/main.go", StartLine: 0, StartColumn: 1, Message: "message", SourceSnippet: "os.Open(name)"}}
		}},
		{"invalid snippet", func(report *reportEnvelope) {
			report.Findings = []finding{{Rule: "G304", Path: "cmd/app/main.go", StartLine: 1, StartColumn: 1, Message: "message", SourceSnippet: "\"unterminated"}}
		}},
		{"conflicting same location", func(report *reportEnvelope) {
			report.Findings = []finding{{Rule: "G304", Path: "cmd/app/main.go", StartLine: 1, StartColumn: 1, Message: "first", SourceSnippet: "os.Open(one)"}, {Rule: "G304", Path: "cmd/app/main.go", StartLine: 1, StartColumn: 1, Message: "second", SourceSnippet: "os.Open(two)"}}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := loadFixture(t, "different revisions match trusted policy")
			testCase.mutate(fixture.Candidate)
			_, err := compare(fixture.Expected, fixture.Baseline, fixture.Candidate)
			require.Error(t, err)
		})
	}
}

func TestCompareRejectsUnsafePaths(t *testing.T) {
	for _, unsafePath := range []string{
		"/absolute/main.go", "../traversal/main.go", "dir/../traversal/main.go", "\\\\server\\share\\main.go",
		"C:\\absolute\\main.go", "C:relative\\main.go", "https://example.test/main.go", "cmd/\x00main.go", "cmd/\x1fmain.go",
	} {
		t.Run(unsafePath, func(t *testing.T) {
			fixture := loadFixture(t, "different revisions match trusted policy")
			fixture.Candidate.Findings = []finding{{
				Rule: "G304", Path: unsafePath, StartLine: 1, StartColumn: 1, Message: "message", SourceSnippet: "os.Open(name)",
			}}
			_, err := compare(fixture.Expected, fixture.Baseline, fixture.Candidate)
			require.Error(t, err)
		})
	}
}

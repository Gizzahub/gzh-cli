package securityfinding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixture struct {
	Name       string         `json:"name"`
	Expected   metadata       `json:"expected"`
	Baseline   reportEnvelope `json:"baseline"`
	Candidate  reportEnvelope `json:"candidate"`
	WantKnown  int            `json:"want_known"`
	WantNew    int            `json:"want_new"`
	WantUnsafe int            `json:"want_unclassified"`
}

func TestCompareFixtures(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("testdata", "comparison_cases.json"))
	require.NoError(t, err)

	var fixtures []fixture
	require.NoError(t, json.Unmarshal(contents, &fixtures))

	for _, testCase := range fixtures {
		t.Run(testCase.Name, func(t *testing.T) {
			result, compareErr := compare(testCase.Expected, testCase.Baseline, testCase.Candidate)
			require.NoError(t, compareErr)
			require.Len(t, result.Known, testCase.WantKnown)
			require.Len(t, result.New, testCase.WantNew)
			require.Len(t, result.Unclassified, testCase.WantUnsafe)
		})
	}
}

func TestCompareRejectsUntrustedEnvelopeMetadata(t *testing.T) {
	expected := testMetadata()
	baseline := reportEnvelope{metadata: expected}
	candidate := reportEnvelope{metadata: expected}
	candidate.FlagsHash = "other-flags"

	_, err := compare(expected, baseline, candidate)
	require.ErrorContains(t, err, "does not match trusted metadata")
}

func TestCompareRejectsMissingExpectedMetadata(t *testing.T) {
	_, err := compare(metadata{}, reportEnvelope{}, reportEnvelope{})
	require.ErrorContains(t, err, "expected metadata must include")
}

func testMetadata() metadata {
	return metadata{
		SourceSHA:      "f4c1e2d3",
		ScannerVersion: "gosec-v2.28.0",
		ConfigHash:     "a-config-hash",
		FlagsHash:      "a-flags-hash",
	}
}

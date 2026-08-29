package securityfinding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixture struct {
	Name      string          `json:"name"`
	Expected  expectation     `json:"expected"`
	Baseline  *reportEnvelope `json:"baseline"`
	Candidate *reportEnvelope `json:"candidate"`
	Want      comparison      `json:"want"`
}

func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("testdata", "comparison_cases.json"))
	require.NoError(t, err)
	var fixtures []fixture
	require.NoError(t, json.Unmarshal(contents, &fixtures))
	return fixtures
}

func loadFixture(t *testing.T, name string) fixture {
	t.Helper()
	for _, testCase := range loadFixtures(t) {
		if testCase.Name == name {
			return testCase
		}
	}
	t.Fatalf("fixture %q not found", name)
	return fixture{}
}

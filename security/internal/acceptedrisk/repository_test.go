package acceptedrisk

import (
	"errors"
	"io/fs"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repositoryRoot is this package's path back to the repository root.
const repositoryRoot = "../../.."

const (
	policyPath   = "security/policy.yaml"
	registryPath = "security/accepted-risks.yaml"
)

// unwiredVerifier is the deliberate default for the tracked registry: no commit
// verifier is wired yet, so any record that claimed to be approved would be
// reported as unverified rather than accepted.
type unwiredVerifier struct{}

func (unwiredVerifier) VerifyCommit(string) (verifiedCommit, error) {
	return verifiedCommit{}, errors.New("no signed-commit verifier is wired for this repository yet")
}

func TestRepositoryPolicyAndRegistryDecode(t *testing.T) {
	root := os.DirFS(repositoryRoot)

	policyContents, err := fs.ReadFile(root, policyPath)
	require.NoError(t, err)
	current, err := decodePolicy(policyContents)
	require.NoError(t, err)
	assert.Equal(t, 90, current.Cadence.ReviewIntervalDays)
	assert.Equal(t, 180, current.Cadence.HardSunsetDays)
	require.Len(t, current.Approvers, 1)
	assert.Equal(t, int64(1732826), current.Approvers[0].ID)

	registryContents, err := fs.ReadFile(root, registryPath)
	require.NoError(t, err)
	records, err := decodeRegistry(registryContents)
	require.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestRepositorySuppressionsAndRecordsAreOneToOne(t *testing.T) {
	root := os.DirFS(repositoryRoot)
	records, suppressions := repositoryState(t, root)

	recordIDs := make([]string, 0, len(records))
	for _, current := range records {
		recordIDs = append(recordIDs, current.ID)
	}
	suppressionIDs := make([]string, 0, len(suppressions))
	for _, current := range suppressions {
		require.NotEmpty(t, current.RiskID, "malformed directive at %s: %s", current.location(), current.Raw)
		suppressionIDs = append(suppressionIDs, current.RiskID)
	}
	sort.Strings(recordIDs)
	sort.Strings(suppressionIDs)

	assert.Equal(t, recordIDs, suppressionIDs,
		"every tracked gosec:disable must reference exactly one record and every record exactly one site")
}

// TestRepositoryRegistryIsBlockedOnlyByPendingApproval pins the current authority
// state: the registry is structurally complete and fully linked, and the single
// remaining blocker on every record is that the policy owner has not yet approved
// it. When a real signed approval commit lands, this test must be updated with the
// record that is no longer pending.
func TestRepositoryRegistryIsBlockedOnlyByPendingApproval(t *testing.T) {
	root := os.DirFS(repositoryRoot)
	records, suppressions := repositoryState(t, root)

	policyContents, err := fs.ReadFile(root, policyPath)
	require.NoError(t, err)
	current, err := decodePolicy(policyContents)
	require.NoError(t, err)

	// The evaluation instant is fixed at the registry's creation date so that this
	// assertion isolates the approval blocker; the cadence rules themselves are
	// covered by TestValidateRegistryEnforcesCadence.
	violations, err := validateRegistry(validationInput{
		Policy:       current,
		Records:      records,
		Suppressions: suppressions,
		Verifier:     unwiredVerifier{},
		Now:          testDate(t, "2026-09-02"),
	})
	require.NoError(t, err)
	require.Len(t, violations, len(records))
	for _, found := range violations {
		assert.Equal(t, codeEvidencePending, found.Code, found.Subject)
	}
	assert.Error(t, violationsError(violations), "the tracked registry must not validate before the owner approves it")
}

func repositoryState(t *testing.T, root fs.FS) ([]record, []suppression) {
	t.Helper()

	registryContents, err := fs.ReadFile(root, registryPath)
	require.NoError(t, err)
	records, err := decodeRegistry(registryContents)
	require.NoError(t, err)

	suppressions, err := scanSuppressions(root)
	require.NoError(t, err)
	return records, suppressions
}

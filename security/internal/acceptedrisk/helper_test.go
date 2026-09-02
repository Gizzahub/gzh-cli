package acceptedrisk

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const approvedSHA = "0123456789abcdef0123456789abcdef01234567"

const validPolicyYAML = `version: 1
organization:
  id: 157453618
  login: Gizzahub
approvers:
  - id: 1732826
    login: archmagece
    type: User
    role: repository-owner
evidence:
  accepted_types:
    - signed-commit
cadence:
  review_interval_days: 90
  hard_sunset_days: 180
`

const recordTemplate = `  - id: %s
    rule: %s
    path: %s
    symbol: exampleSymbol
    threat: an example threat statement
    compensating_control: an example compensating control
    owner: archmagece
    approver:
      id: %d
      login: %s
    created_at: "%s"
    last_reviewed_at: "%s"
    evidence:
      type: %s
      sha: %s
    test_evidence: an example test reference
`

// recordSpec renders one registry record so a test can vary exactly one field.
type recordSpec struct {
	ID             string
	Rule           string
	Path           string
	ApproverID     int64
	ApproverLogin  string
	CreatedAt      string
	LastReviewedAt string
	EvidenceType   string
	EvidenceSHA    string
}

func defaultSpec() recordSpec {
	return recordSpec{
		ID:             "AR-2026-001",
		Rule:           "G304",
		Path:           "cmd/example/example.go",
		ApproverID:     1732826,
		ApproverLogin:  "archmagece",
		CreatedAt:      "2026-09-02",
		LastReviewedAt: "2026-09-02",
		EvidenceType:   evidenceTypeSignedCommit,
		EvidenceSHA:    approvedSHA,
	}
}

func (spec recordSpec) render() string {
	return fmt.Sprintf(recordTemplate, spec.ID, spec.Rule, spec.Path, spec.ApproverID, spec.ApproverLogin,
		spec.CreatedAt, spec.LastReviewedAt, spec.EvidenceType, spec.EvidenceSHA)
}

func registryDoc(specs ...recordSpec) string {
	var builder strings.Builder
	builder.WriteString("version: 1\nrecords:\n")
	for _, spec := range specs {
		builder.WriteString(spec.render())
	}
	return builder.String()
}

func testPolicy(t *testing.T) *policy {
	t.Helper()
	decoded, err := decodePolicy([]byte(validPolicyYAML))
	require.NoError(t, err)
	return decoded
}

func testRecords(t *testing.T, specs ...recordSpec) []record {
	t.Helper()
	records, err := decodeRegistry([]byte(registryDoc(specs...)))
	require.NoError(t, err)
	return records
}

func testDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(dateLayout, value)
	require.NoError(t, err)
	return parsed
}

// fakeVerifier stands in for `git verify-commit`. The zero value reports a
// verified signature for whatever commit it is asked about.
type fakeVerifier struct {
	err        error
	unverified bool
	reportSHA  string
}

func (f fakeVerifier) VerifyCommit(sha string) (verifiedCommit, error) {
	if f.err != nil {
		return verifiedCommit{}, f.err
	}
	reported := f.reportSHA
	if reported == "" {
		reported = sha
	}
	return verifiedCommit{SHA: reported, Verified: !f.unverified}, nil
}

// suppressionFor builds the directive a valid record expects.
func suppressionFor(spec recordSpec, line int) suppression {
	return suppression{
		Path:   spec.Path,
		Line:   line,
		Rule:   spec.Rule,
		RiskID: spec.ID,
		Reason: "an example reason",
		Raw:    fmt.Sprintf("//gosec:disable %s -- %s an example reason", spec.Rule, spec.ID),
	}
}

func violationCodes(values []violation) []string {
	codes := make([]string, 0, len(values))
	for _, current := range values {
		codes = append(codes, current.Code)
	}
	return codes
}

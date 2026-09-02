package acceptedrisk

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const approvedSHA = "0123456789abcdef0123456789abcdef01234567"

// registeredSigningKey is the fingerprint validPolicyYAML trusts. It is an
// obvious placeholder: no real key fingerprint belongs in test fixtures.
const registeredSigningKey = "SHA256:example-fingerprint-for-accepted-risk-tests"

// defaultApprovalMessage names every identifier the tests approve, because the
// validator requires an approval commit to name the record it ratifies.
const defaultApprovalMessage = "security: approve AR-2026-001 and AR-2026-002"

// toolsMakefilePath holds the pinned gosec scan flags the scan scope derives from.
const toolsMakefilePath = ".make/tools.mk"

const validPolicyYAML = `version: 1
organization:
  id: 157453618
  login: Gizzahub
approvers:
  - id: 1732826
    login: archmagece
    type: User
    role: repository-owner
    signing_keys:
      - SHA256:example-fingerprint-for-accepted-risk-tests
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

// testScanScope derives the scan scope from the repository's real pinned flags,
// so a change to GOSEC_SCAN_FLAGS is visible to every test that scans a tree.
func testScanScope(t *testing.T) scanScope {
	t.Helper()
	contents, err := fs.ReadFile(os.DirFS(repositoryRoot), toolsMakefilePath)
	require.NoError(t, err)
	scope, err := gosecScanScope(contents)
	require.NoError(t, err)
	return scope
}

// testBlanketTokens derives the blanket tags from the repository's real
// .gosec.json for the same reason testScanScope reads the real tool flags: the
// tags gosec honors are a property of that file, and a test that supplied its
// own would stop noticing when the file and the scanner disagreed.
func testBlanketTokens(t *testing.T) blanketTokens {
	t.Helper()
	contents, err := fs.ReadFile(os.DirFS(repositoryRoot), gosecConfigPath)
	require.NoError(t, err)
	tokens, err := gosecBlanketTokens(contents)
	require.NoError(t, err)
	return tokens
}

func testDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(dateLayout, value)
	require.NoError(t, err)
	return parsed
}

// fakeVerifier stands in for a production verifier. The zero value reports a
// signature that verifies, was made with the key validPolicyYAML registers, and
// covers a message naming the identifiers the tests approve.
type fakeVerifier struct {
	err             error
	unverified      bool
	reportSHA       string
	signerKey       string
	signerAccountID int64
	message         string
	// noSignerKey and noMessage distinguish "the verifier established nothing"
	// from "the test did not override the default".
	noSignerKey bool
	noMessage   bool
}

func (f fakeVerifier) VerifyCommit(sha string) (verifiedCommit, error) {
	if f.err != nil {
		return verifiedCommit{}, f.err
	}
	reported := f.reportSHA
	if reported == "" {
		reported = sha
	}
	signerKey := f.signerKey
	if signerKey == "" && !f.noSignerKey {
		signerKey = registeredSigningKey
	}
	message := f.message
	if message == "" && !f.noMessage {
		message = defaultApprovalMessage
	}
	return verifiedCommit{
		SHA:                     reported,
		Verified:                !f.unverified,
		VerifiedSignerKey:       signerKey,
		VerifiedSignerAccountID: f.signerAccountID,
		Message:                 message,
	}, nil
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

package acceptedrisk

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegistryAcceptsAnApprovedLinkedRecord(t *testing.T) {
	spec := defaultSpec()
	violations, err := validateRegistry(validationInput{
		Policy:       testPolicy(t),
		Records:      testRecords(t, spec),
		Suppressions: []suppression{suppressionFor(spec, 42)},
		Verifier:     fakeVerifier{},
		Now:          testDate(t, "2026-09-02"),
	})
	require.NoError(t, err)
	assert.Empty(t, violations)
	assert.NoError(t, violationsError(violations))
}

func TestValidateRegistryRequiresCompleteInput(t *testing.T) {
	spec := defaultSpec()
	complete := validationInput{
		Policy:       testPolicy(t),
		Records:      testRecords(t, spec),
		Suppressions: []suppression{suppressionFor(spec, 42)},
		Verifier:     fakeVerifier{},
		Now:          testDate(t, "2026-09-02"),
	}

	cases := map[string]func(input validationInput) validationInput{
		"nil policy":       func(input validationInput) validationInput { input.Policy = nil; return input },
		"nil records":      func(input validationInput) validationInput { input.Records = nil; return input },
		"nil suppressions": func(input validationInput) validationInput { input.Suppressions = nil; return input },
		"nil verifier":     func(input validationInput) validationInput { input.Verifier = nil; return input },
		"zero time":        func(input validationInput) validationInput { input.Now = time.Time{}; return input },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			violations, err := validateRegistry(mutate(complete))
			require.Error(t, err)
			assert.Nil(t, violations)
		})
	}
}

func TestValidateRegistryRejectsUnauthorizedApprovals(t *testing.T) {
	agentPolicy := testPolicy(t)
	agentPolicy.Approvers = append(agentPolicy.Approvers, policyApprover{
		ID: 99, Login: "release-bot", Type: "Bot", Role: "automation",
	})

	cases := []struct {
		name     string
		policy   *policy
		spec     recordSpec
		verifier commitVerifier
		want     string
	}{
		{
			name:     "approver absent from the trusted base",
			spec:     withApprover(defaultSpec(), 424242, "someone-else"),
			verifier: fakeVerifier{},
			want:     codeApproverNotInPolicy,
		},
		{
			name:     "approver is a bot identity",
			policy:   agentPolicy,
			spec:     withApprover(defaultSpec(), 99, "release-bot"),
			verifier: fakeVerifier{},
			want:     codeApproverNotHuman,
		},
		{
			name:     "registry login has drifted from the trusted base",
			spec:     withApprover(defaultSpec(), 1732826, "old-login"),
			verifier: fakeVerifier{},
			want:     codeApproverLoginMismatch,
		},
		{
			name:     "approval is still pending",
			spec:     withEvidence(defaultSpec(), evidenceTypeSignedCommit, pendingApprovalSentinel),
			verifier: fakeVerifier{},
			want:     codeEvidencePending,
		},
		{
			name:     "evidence type is mutable",
			spec:     withEvidence(defaultSpec(), "github-issue", approvedSHA),
			verifier: fakeVerifier{},
			want:     codeEvidenceTypeUnsupported,
		},
		{
			name:     "abbreviated sha",
			spec:     withEvidence(defaultSpec(), evidenceTypeSignedCommit, "0123456"),
			verifier: fakeVerifier{},
			want:     codeEvidenceSHAMalformed,
		},
		{
			name:     "uppercase sha",
			spec:     withEvidence(defaultSpec(), evidenceTypeSignedCommit, "0123456789ABCDEF0123456789abcdef01234567"),
			verifier: fakeVerifier{},
			want:     codeEvidenceSHAMalformed,
		},
		{
			name:     "signature does not verify",
			verifier: fakeVerifier{unverified: true},
			want:     codeEvidenceUnverified,
		},
		{
			name:     "verifier fails",
			verifier: fakeVerifier{err: errors.New("no signing key")},
			want:     codeEvidenceUnverified,
		},
		{
			name:     "verifier answers about a different commit",
			verifier: fakeVerifier{reportSHA: "89abcdef0123456789abcdef0123456789abcdef"},
			want:     codeEvidenceUnverified,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			spec := testCase.spec
			if spec.ID == "" {
				spec = defaultSpec()
			}
			current := testCase.policy
			if current == nil {
				current = testPolicy(t)
			}
			violations, err := validateRegistry(validationInput{
				Policy:       current,
				Records:      testRecords(t, spec),
				Suppressions: []suppression{suppressionFor(spec, 42)},
				Verifier:     testCase.verifier,
				Now:          testDate(t, "2026-09-02"),
			})
			require.NoError(t, err)
			assert.Contains(t, violationCodes(violations), testCase.want)
			assert.Error(t, violationsError(violations))
		})
	}
}

func TestValidateRegistryEnforcesCadence(t *testing.T) {
	spec := defaultSpec()
	cases := []struct {
		name string
		spec recordSpec
		now  string
		want []string
	}{
		{
			name: "the day the review falls due is still valid",
			spec: spec,
			now:  "2026-12-01",
			want: nil,
		},
		{
			name: "review overdue",
			spec: spec,
			now:  "2026-12-02",
			want: []string{codeReviewOverdue},
		},
		{
			name: "hard sunset passed",
			spec: spec,
			now:  "2027-03-02",
			want: []string{codeHardSunsetExpired, codeReviewOverdue},
		},
		{
			name: "renewal after the hard sunset",
			spec: withDates(spec, "2026-09-02", "2027-03-02"),
			now:  "2027-03-02",
			want: []string{codeHardSunsetExpired, codeRenewalAfterHardSunset},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			violations, err := validateRegistry(validationInput{
				Policy:       testPolicy(t),
				Records:      testRecords(t, testCase.spec),
				Suppressions: []suppression{suppressionFor(testCase.spec, 42)},
				Verifier:     fakeVerifier{},
				Now:          testDate(t, testCase.now),
			})
			require.NoError(t, err)
			assert.Equal(t, testCase.want, emptyToNil(violationCodes(violations)))
		})
	}
}

func TestValidateRegistryEnforcesSuppressionLinkage(t *testing.T) {
	spec := defaultSpec()
	other := defaultSpec()
	other.ID = "AR-2026-002"
	other.Path = "internal/example/other.go"

	cases := []struct {
		name         string
		records      []recordSpec
		suppressions []suppression
		want         []string
	}{
		{
			name:         "record nothing references",
			records:      []recordSpec{spec},
			suppressions: []suppression{},
			want:         []string{codeRecordOrphaned},
		},
		{
			name:         "suppression with no record",
			records:      []recordSpec{spec},
			suppressions: []suppression{suppressionFor(spec, 42), suppressionFor(other, 7)},
			want:         []string{codeSuppressionUnregistered},
		},
		{
			name:    "malformed directive",
			records: []recordSpec{spec},
			suppressions: []suppression{
				suppressionFor(spec, 42),
				{Path: "cmd/example/example.go", Line: 9, Raw: "//gosec:disable G304 -- no identifier"},
			},
			want: []string{codeSuppressionMalformed},
		},
		{
			name:         "two sites share one record",
			records:      []recordSpec{spec},
			suppressions: []suppression{suppressionFor(spec, 42), suppressionFor(spec, 99)},
			want:         []string{codeSuppressionDuplicateRef},
		},
		{
			name:    "rule does not match the record",
			records: []recordSpec{spec},
			suppressions: []suppression{
				withSuppressionRule(suppressionFor(spec, 42), "G302"),
			},
			want: []string{codeSuppressionRuleMismatch},
		},
		{
			name:    "path does not match the record",
			records: []recordSpec{spec},
			suppressions: []suppression{
				withSuppressionPath(suppressionFor(spec, 42), "internal/example/moved.go"),
			},
			want: []string{codeSuppressionPathMismatch},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			violations, err := validateRegistry(validationInput{
				Policy:       testPolicy(t),
				Records:      testRecords(t, testCase.records...),
				Suppressions: testCase.suppressions,
				Verifier:     fakeVerifier{},
				Now:          testDate(t, "2026-09-02"),
			})
			require.NoError(t, err)
			assert.Equal(t, testCase.want, emptyToNil(violationCodes(violations)))
			assert.Error(t, violationsError(violations))
		})
	}
}

func TestValidateRegistrySortsViolationsIndependentlyOfInputOrder(t *testing.T) {
	first := withEvidence(defaultSpec(), evidenceTypeSignedCommit, pendingApprovalSentinel)
	second := withEvidence(defaultSpec(), evidenceTypeSignedCommit, pendingApprovalSentinel)
	second.ID = "AR-2026-002"
	second.Path = "internal/example/other.go"

	forward, err := validateRegistry(validationInput{
		Policy:       testPolicy(t),
		Records:      testRecords(t, first, second),
		Suppressions: []suppression{suppressionFor(second, 7), suppressionFor(first, 42)},
		Verifier:     fakeVerifier{},
		Now:          testDate(t, "2026-09-02"),
	})
	require.NoError(t, err)

	reverse, err := validateRegistry(validationInput{
		Policy:       testPolicy(t),
		Records:      testRecords(t, second, first),
		Suppressions: []suppression{suppressionFor(first, 42), suppressionFor(second, 7)},
		Verifier:     fakeVerifier{},
		Now:          testDate(t, "2026-09-02"),
	})
	require.NoError(t, err)

	assert.Equal(t, forward, reverse)
	require.Len(t, forward, 2)
	assert.Equal(t, "AR-2026-001", forward[0].Subject)
	assert.Equal(t, "AR-2026-002", forward[1].Subject)
}

func withApprover(spec recordSpec, id int64, login string) recordSpec {
	spec.ApproverID = id
	spec.ApproverLogin = login
	return spec
}

func withEvidence(spec recordSpec, evidenceType, sha string) recordSpec {
	spec.EvidenceType = evidenceType
	spec.EvidenceSHA = sha
	return spec
}

func withSuppressionRule(current suppression, rule string) suppression {
	current.Rule = rule
	return current
}

func withSuppressionPath(current suppression, path string) suppression {
	current.Path = path
	return current
}

func emptyToNil(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return values
}

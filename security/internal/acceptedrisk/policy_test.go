package acceptedrisk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodePolicyAcceptsTheTrustedBaseSSOT(t *testing.T) {
	decoded, err := decodePolicy([]byte(validPolicyYAML))
	require.NoError(t, err)

	assert.Equal(t, supportedSchemaVersion, decoded.Version)
	assert.Equal(t, int64(157453618), decoded.Organization.ID)
	require.Len(t, decoded.Approvers, 1)
	assert.Equal(t, int64(1732826), decoded.Approvers[0].ID)
	assert.Equal(t, approverTypeUser, decoded.Approvers[0].Type)
	assert.True(t, decoded.acceptsEvidenceType(evidenceTypeSignedCommit))
	assert.False(t, decoded.acceptsEvidenceType("github-issue"))

	approver, found := decoded.approverByID(1732826)
	assert.True(t, found)
	assert.Equal(t, "archmagece", approver.Login)

	_, found = decoded.approverByID(1)
	assert.False(t, found)
}

func TestDecodePolicyFailsClosed(t *testing.T) {
	cases := []struct {
		name     string
		contents string
		message  string
	}{
		{
			name:     "unknown field",
			contents: validPolicyYAML + "maintainers:\n  - someone\n",
			message:  "field maintainers not found",
		},
		{
			name:     "unsupported version",
			contents: strings.Replace(validPolicyYAML, "version: 1", "version: 2", 1),
			message:  "version must be 1",
		},
		{
			name:     "bot approver type",
			contents: strings.Replace(validPolicyYAML, "type: User", "type: Bot", 1),
			message:  `approver type must be "User"`,
		},
		{
			name:     "agent approver login",
			contents: strings.Replace(validPolicyYAML, "login: archmagece\n    type", "login: claude-agent\n    type", 1),
			message:  "automation or agent identity",
		},
		{
			name: "duplicate approver id",
			contents: strings.Replace(validPolicyYAML, "    role: repository-owner\n",
				"    role: repository-owner\n  - id: 1732826\n    login: archmagece\n    type: User\n    role: security-maintainer\n", 1),
			message: "duplicate approver id",
		},
		{
			name:     "no approvers",
			contents: strings.Replace(validPolicyYAML, "approvers:\n  - id: 1732826\n    login: archmagece\n    type: User\n    role: repository-owner\n", "approvers: []\n", 1),
			message:  "at least one approver",
		},
		{
			name:     "mutable evidence type",
			contents: strings.Replace(validPolicyYAML, "- signed-commit", "- github-issue-url", 1),
			message:  "not an immutable approval evidence format",
		},
		{
			name:     "sunset before review interval",
			contents: strings.Replace(validPolicyYAML, "hard_sunset_days: 180", "hard_sunset_days: 30", 1),
			message:  "hard sunset must not precede",
		},
		{
			name:     "non positive cadence",
			contents: strings.Replace(validPolicyYAML, "review_interval_days: 90", "review_interval_days: 0", 1),
			message:  "cadence days must be positive",
		},
		{
			name:     "missing organization",
			contents: strings.Replace(validPolicyYAML, "id: 157453618", "id: 0", 1),
			message:  "organization requires a positive id",
		},
		{
			name:     "second document",
			contents: validPolicyYAML + "---\n" + validPolicyYAML,
			message:  "exactly one document",
		},
		{
			name:     "not yaml",
			contents: "\tnot: [yaml",
			message:  "decode security policy",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			decoded, err := decodePolicy([]byte(testCase.contents))
			require.Error(t, err)
			assert.Nil(t, decoded)
			assert.Contains(t, err.Error(), testCase.message)
		})
	}
}

func TestIsNonHumanLoginRejectsAutomationIdentities(t *testing.T) {
	for _, login := range []string{"dependabot[bot]", "renovate", "github-actions", "some-bot", "Claude", "codex", "svc_bot"} {
		assert.True(t, isNonHumanLogin(login), login)
	}
	for _, login := range []string{"archmagece", "celee", "alice"} {
		assert.False(t, isNonHumanLogin(login), login)
	}
}

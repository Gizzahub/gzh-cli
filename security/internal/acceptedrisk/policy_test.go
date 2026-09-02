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
	assert.True(t, approver.authorizesSigningKey(registeredSigningKey))
	assert.False(t, approver.authorizesSigningKey("SHA256:some-other-key"))
	assert.False(t, approver.authorizesSigningKey(""), "an empty fingerprint must never authorize")

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
			contents: strings.Replace(validPolicyYAML, approverBlock,
				approverBlock+"  - id: 1732826\n    login: archmagece\n    type: User\n    role: security-maintainer\n    signing_keys: []\n", 1),
			message: "duplicate approver id",
		},
		{
			name:     "no approvers",
			contents: strings.Replace(validPolicyYAML, approverBlock, "approvers: []\n", 1),
			message:  "at least one approver",
		},
		{
			name:     "signing keys omitted",
			contents: strings.Replace(validPolicyYAML, "    signing_keys:\n      - "+registeredSigningKey+"\n", "", 1),
			message:  "signing_keys must be an explicit list",
		},
		{
			name:     "duplicate signing key",
			contents: strings.Replace(validPolicyYAML, "      - "+registeredSigningKey+"\n", "      - "+registeredSigningKey+"\n      - "+registeredSigningKey+"\n", 1),
			message:  "duplicate signing key",
		},
		{
			name:     "blank signing key",
			contents: strings.Replace(validPolicyYAML, "      - "+registeredSigningKey+"\n", "      - \"  \"\n", 1),
			message:  "trimmed, non-empty fingerprint",
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
		// A trailing document is rejected on its presence, never on its shape.
		// None of these decode into a policy, so a guard that read a failed
		// decode as "nothing follows" would accept every one of them.
		{
			name:     "trailing scalar document",
			contents: validPolicyYAML + "--- 42\n",
			message:  "exactly one document",
		},
		{
			name:     "trailing sequence document",
			contents: validPolicyYAML + "---\n- approvers\n- cadence\n",
			message:  "exactly one document",
		},
		{
			name:     "trailing wrong-typed document",
			contents: validPolicyYAML + "---\nversion:\n  - 1\n",
			message:  "exactly one document",
		},
		{
			name:     "trailing document with unknown fields",
			contents: validPolicyYAML + "---\nmaintainers:\n  - someone\n",
			message:  "exactly one document",
		},
		{
			name:     "homoglyph approver login",
			contents: strings.Replace(validPolicyYAML, "login: archmagece\n    type", "login: сlaude-agent\n    type", 1),
			message:  "not a well-formed GitHub login",
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

// approverBlock is the approvers section of validPolicyYAML, so a fixture can
// replace it wholesale instead of matching fragments that move when the schema
// grows.
const approverBlock = `approvers:
  - id: 1732826
    login: archmagece
    type: User
    role: repository-owner
    signing_keys:
      - ` + registeredSigningKey + `
`

// TestPolicyApproverWithNoRegisteredKeyAuthorizesNothing pins the fail-closed
// meaning of an empty list: it is a valid policy, and it grants nothing.
func TestPolicyApproverWithNoRegisteredKeyAuthorizesNothing(t *testing.T) {
	contents := strings.Replace(validPolicyYAML,
		"    signing_keys:\n      - "+registeredSigningKey+"\n", "    signing_keys: []\n", 1)
	decoded, err := decodePolicy([]byte(contents))
	require.NoError(t, err)

	approver, found := decoded.approverByID(1732826)
	require.True(t, found)
	assert.Empty(t, approver.SigningKeys)
	assert.False(t, approver.authorizesSigningKey(registeredSigningKey))
}

func TestIsNonHumanLoginRejectsAutomationIdentities(t *testing.T) {
	for _, login := range []string{"dependabot[bot]", "renovate", "github-actions", "some-bot", "Claude", "codex", "svc_bot"} {
		assert.True(t, isNonHumanLogin(login), login)
	}
	for _, login := range []string{"archmagece", "celee", "alice"} {
		assert.False(t, isNonHumanLogin(login), login)
	}
}

// TestIsWellFormedLoginRejectsLoginsGitHubCouldNotIssue pins the character-level
// authority rule. The marker scan compares ASCII substrings, so a login spelled
// with a Cyrillic homoglyph renders exactly like a name a reviewer trusts and
// matches no marker at all; a login GitHub could never issue cannot be an
// authority, which closes that vector without depending on the marker list.
func TestIsWellFormedLoginRejectsLoginsGitHubCouldNotIssue(t *testing.T) {
	unissuable := map[string]string{
		"cyrillic homoglyph": "сlaude-agent",
		"leading hyphen":     "-archmagece",
		"trailing hyphen":    "archmagece-",
		"consecutive hyphen": "arch--magece",
		"underscore":         "arch_magece",
		"dot":                "arch.magece",
		"space":              "arch magece",
		"empty":              "",
		"too long":           strings.Repeat("a", maxGitHubLoginLength+1),
	}
	for name, login := range unissuable {
		t.Run(name, func(t *testing.T) {
			assert.False(t, isWellFormedLogin(login), login)
			assert.False(t, isHumanApproverLogin(login), login)
		})
	}

	for _, login := range []string{"a", "archmagece", "Arch-Magece-9", "celee", strings.Repeat("a", maxGitHubLoginLength)} {
		assert.True(t, isWellFormedLogin(login), login)
		assert.True(t, isHumanApproverLogin(login), login)
	}
}

// TestIsHumanApproverLoginLeavesMarkerlessAutomationToReview records what neither
// rule reaches. "gzh-release-automation" is a login GitHub would issue and carries
// no marker, so both checks accept it. Whether an account belongs to a person is a
// judgment about ownership, not about characters, and it stays with the review
// that adds an approver to security/policy.yaml.
func TestIsHumanApproverLoginLeavesMarkerlessAutomationToReview(t *testing.T) {
	const markerless = "gzh-release-automation"

	assert.True(t, isWellFormedLogin(markerless))
	assert.False(t, isNonHumanLogin(markerless))
	assert.True(t, isHumanApproverLogin(markerless),
		"a marker-less automation name is a documented gap, not a rule this package can close")
}

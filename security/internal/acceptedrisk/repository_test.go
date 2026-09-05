package acceptedrisk

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repositoryRoot is this package's path back to the repository root.
const repositoryRoot = "../../.."

const (
	policyPath      = "security/policy.yaml"
	registryPath    = "security/accepted-risks.yaml"
	gosecConfigPath = ".gosec.json"
)

// unwiredVerifier is the deliberate default for the tracked registry: no commit
// verifier is wired yet, so any record that claimed to be approved would be
// reported as unverified rather than accepted.
type unwiredVerifier struct{}

func (unwiredVerifier) VerifyCommit(string) (verifiedCommit, error) {
	return verifiedCommit{}, errors.New("no signed-commit verifier is wired for this repository yet")
}

// blessingVerifier reports a perfectly verified signature for anything it is
// asked about. It exists to prove that a verified signature alone is not enough:
// the tracked policy registers no signing key, so nothing it blesses is accepted.
type blessingVerifier struct{}

func (blessingVerifier) VerifyCommit(sha string) (verifiedCommit, error) {
	return verifiedCommit{
		SHA:               sha,
		Verified:          true,
		VerifiedSignerKey: "SHA256:an-unregistered-key-that-verifies-fine",
		Message:           "security: approve every accepted risk",
	}, nil
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
		require.False(t, current.Blanket,
			"unregistrable blanket suppression at %s: %s", current.location(), current.Raw)
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

// TestRepositoryRegistryRejectsAnUnregisteredSigner pins the property that a
// signature which merely verifies proves nothing about authority. Anyone able to
// push to the trusted base can sign a commit with their own key; the record is
// accepted only when that key is registered to the record's approver.
func TestRepositoryRegistryRejectsAnUnregisteredSigner(t *testing.T) {
	root := os.DirFS(repositoryRoot)
	records, suppressions := repositoryState(t, root)

	policyContents, err := fs.ReadFile(root, policyPath)
	require.NoError(t, err)
	current, err := decodePolicy(policyContents)
	require.NoError(t, err)

	// Replace the pending sentinel so evaluation reaches the signature checks.
	approved := make([]record, 0, len(records))
	for _, entry := range records {
		entry.Evidence.SHA = "0123456789abcdef0123456789abcdef01234567"
		approved = append(approved, entry)
	}

	violations, err := validateRegistry(validationInput{
		Policy:       current,
		Records:      approved,
		Suppressions: suppressions,
		Verifier:     blessingVerifier{},
		Now:          testDate(t, "2026-09-02"),
	})
	require.NoError(t, err)
	require.Len(t, violations, len(approved))
	for _, found := range violations {
		assert.Equal(t, codeSignerNotAuthorized, found.Code, found.Subject)
	}
}

// TestRepositoryGosecConfigDisablesBlanketSuppression couples this gate to the
// settings that decide which blanket tag the pinned gosec actually honors.
//
// The scanner reports every occurrence of a live tag, but nothing in this
// package can stop gosec from honoring one. That is decided entirely by
// .gosec.json, and measured against the pinned gosec v2.28.0 binary the
// mechanism is a rename rather than a switch: the live tag is "#" followed by
// the configured global.nosec value, so the tracked setting of false makes
// "#false" the live form and gosec's built-in spelling inert. Removing the line
// does not relax a check so much as move the tag back to the built-in spelling,
// which is why the scanner derives the tag from this file instead of assuming
// one.
//
// The value pin is therefore not stylistic. An earlier reading of this coupling,
// that the key's presence is what matters and the value is irrelevant, was
// wrong; false is the only workable value, for two separate measured reasons:
//
//   - Any other value renames the live tag to match it, so a setting of "off"
//     would put "#off" in play and leave every other spelling inert.
//   - true does something different and worse. It sets gosec's ignoreNosec,
//     which disables every suppression grammar including the directive form, so
//     this registry's own directives would stop suppressing and the security
//     gate would fail on risks it has already accepted. It is not available as a
//     hardening.
//
// The alternative key is pinned by the same test because gosec honors it in
// addition to the default tag rather than in place of it. Unpinned it could be
// added silently, putting a second blanket form in play; the scanner would
// follow it, but this repository has no use for one.
func TestRepositoryGosecConfigDisablesBlanketSuppression(t *testing.T) {
	contents, err := fs.ReadFile(os.DirFS(repositoryRoot), gosecConfigPath)
	require.NoError(t, err)

	var config struct {
		Global struct {
			Nosec       *bool           `json:"nosec"`
			Alternative json.RawMessage `json:"#nosec"`
		} `json:"global"`
	}
	require.NoError(t, json.Unmarshal(contents, &config))

	require.NotNil(t, config.Global.Nosec,
		"%s must declare global.nosec; without it gosec's built-in blanket tag is the live one", gosecConfigPath)
	assert.False(t, *config.Global.Nosec,
		"%s must keep global.nosec false: true disables the directive form too, and any other value renames the blanket tag to match it", gosecConfigPath)
	assert.Nil(t, config.Global.Alternative,
		"%s must not declare the %q alternative; it adds a second live blanket tag beside the default one", gosecConfigPath, gosecAlternativeKey)

	tokens, err := gosecBlanketTokens(contents)
	require.NoError(t, err)
	assert.Equal(t, blanketTokens{"#false", "#nosec"}, tokens,
		"the scanner must watch the tag this configuration makes live, plus the built-in spelling that returns if the setting is removed")
}

// TestRepositoryRegistryEnforcesCadenceAsTheClockAdvances exercises the cadence
// rules against the tracked registry rather than against a fixture. Pinning the
// evaluation instant to the records' own creation date, as the approval test
// does, makes review-overdue and hard-sunset structurally unreachable no matter
// how much time passes, so the clock is advanced here instead.
func TestRepositoryRegistryEnforcesCadenceAsTheClockAdvances(t *testing.T) {
	root := os.DirFS(repositoryRoot)
	records, suppressions := repositoryState(t, root)

	policyContents, err := fs.ReadFile(root, policyPath)
	require.NoError(t, err)
	current, err := decodePolicy(policyContents)
	require.NoError(t, err)

	cases := []struct {
		name string
		now  string
		want []string
	}{
		{
			name: "the day the 90-day review falls due is still valid",
			now:  "2026-12-01",
			want: []string{codeEvidencePending},
		},
		{
			name: "one day past the 90-day review interval",
			now:  "2026-12-02",
			want: []string{codeEvidencePending, codeReviewOverdue},
		},
		{
			name: "the day the 180-day hard sunset falls due is still valid",
			now:  "2027-03-01",
			want: []string{codeEvidencePending, codeReviewOverdue},
		},
		{
			name: "one day past the 180-day hard sunset",
			now:  "2027-03-02",
			want: []string{codeEvidencePending, codeHardSunsetExpired, codeReviewOverdue},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			violations, err := validateRegistry(validationInput{
				Policy:       current,
				Records:      records,
				Suppressions: suppressions,
				Verifier:     unwiredVerifier{},
				Now:          testDate(t, testCase.now),
			})
			require.NoError(t, err)
			require.Len(t, violations, len(records)*len(testCase.want))

			byRecord := make(map[string][]string, len(records))
			for _, found := range violations {
				byRecord[found.Subject] = append(byRecord[found.Subject], found.Code)
			}
			for _, entry := range records {
				codes := byRecord[entry.ID]
				sort.Strings(codes)
				assert.Equal(t, testCase.want, codes, entry.ID)
			}
		})
	}
}

func repositoryState(t *testing.T, root fs.FS) ([]record, []suppression) {
	t.Helper()

	registryContents, err := fs.ReadFile(root, registryPath)
	require.NoError(t, err)
	records, err := decodeRegistry(registryContents)
	require.NoError(t, err)

	contents, err := fs.ReadFile(root, toolsMakefilePath)
	require.NoError(t, err)
	scope, err := gosecScanScope(contents)
	require.NoError(t, err)

	suppressions, err := scanSuppressions(root, scope, testBlanketTokens(t))
	require.NoError(t, err)
	return records, suppressions
}

// TestPlatformTagsMatchThePinnedToolchain guards the one place this package
// restates something the toolchain owns. The build-constraint check treats a tag
// outside these lists as never set, so a new port would silently turn a normal
// platform file into one the scanner believes is unreachable. Asking the
// toolchain here means that drift fails a test instead.
func TestPlatformTagsMatchThePinnedToolchain(t *testing.T) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not on PATH; the toolchain cannot be asked what it supports")
	}

	output, err := exec.CommandContext(t.Context(), goBinary, "tool", "dist", "list").Output()
	require.NoError(t, err)

	oses := make(map[string]struct{})
	arches := make(map[string]struct{})
	for _, pair := range strings.Fields(string(output)) {
		platform, architecture, found := strings.Cut(pair, "/")
		require.True(t, found, pair)
		oses[platform] = struct{}{}
		arches[architecture] = struct{}{}
	}

	assert.ElementsMatch(t, keysOf(oses), goosTags,
		"goosTags must equal the GOOS values of the pinned toolchain")
	assert.ElementsMatch(t, keysOf(arches), goarchTags,
		"goarchTags must equal the GOARCH values of the pinned toolchain")
}

func keysOf(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

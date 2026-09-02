package acceptedrisk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeRegistryCanonicalizesRecords(t *testing.T) {
	records, err := decodeRegistry([]byte(registryDoc(defaultSpec())))
	require.NoError(t, err)
	require.Len(t, records, 1)

	current := records[0]
	assert.Equal(t, "AR-2026-001", current.ID)
	assert.Equal(t, "G304", current.Rule)
	assert.Equal(t, "cmd/example/example.go", current.Path)
	assert.Equal(t, "2026-09-02", current.CreatedAt.Format(dateLayout))
	assert.Equal(t, approvedSHA, current.Evidence.SHA)

	cadence := testPolicy(t).Cadence
	assert.Equal(t, "2026-12-01", current.reviewBy(cadence).Format(dateLayout))
	assert.Equal(t, "2027-03-01", current.hardSunset(cadence).Format(dateLayout))
}

func TestDecodeRegistryFailsClosed(t *testing.T) {
	duplicate := defaultSpec()
	second := defaultSpec()
	second.Path = "internal/example/other.go"

	cases := []struct {
		name     string
		contents string
		message  string
	}{
		{
			name:     "unknown field",
			contents: registryDoc(defaultSpec()) + "    reviewed_by: someone\n",
			message:  "field reviewed_by not found",
		},
		{
			name:     "unsupported version",
			contents: strings.Replace(registryDoc(defaultSpec()), "version: 1", "version: 2", 1),
			message:  "registry version must be 1",
		},
		{
			name:     "records not a list",
			contents: "version: 1\nrecords:\n",
			message:  "must be an explicit list",
		},
		{
			name:     "duplicate identifier",
			contents: registryDoc(duplicate, second),
			message:  "duplicate identifier AR-2026-001",
		},
		{
			name:     "malformed identifier",
			contents: registryDoc(withID(defaultSpec(), "AR-26-1")),
			message:  "must have the form AR-YYYY-NNN",
		},
		{
			name:     "identifier year mismatch",
			contents: registryDoc(withID(defaultSpec(), "AR-2025-001")),
			message:  "does not match created_at year",
		},
		{
			name:     "malformed rule",
			contents: registryDoc(withRule(defaultSpec(), "gosec-G304")),
			message:  "must be a gosec rule identifier",
		},
		{
			name:     "absolute path",
			contents: registryDoc(withPath(defaultSpec(), "/etc/passwd.go")),
			message:  "must be repository-relative",
		},
		{
			name:     "traversal path",
			contents: registryDoc(withPath(defaultSpec(), "../outside/example.go")),
			message:  "empty or relative segments",
		},
		{
			name:     "windows drive path",
			contents: registryDoc(withPath(defaultSpec(), "C:/repo/example.go")),
			message:  "must be repository-relative",
		},
		{
			name:     "backslash path",
			contents: registryDoc(withPath(defaultSpec(), `cmd\example\example.go`)),
			message:  "backslashes or control characters",
		},
		{
			name:     "non go path",
			contents: registryDoc(withPath(defaultSpec(), "cmd/example/example.txt")),
			message:  "must name a Go source file",
		},
		{
			name:     "missing narrative field",
			contents: strings.Replace(registryDoc(defaultSpec()), "    threat: an example threat statement\n", "    threat: \"\"\n", 1),
			message:  "threat is required",
		},
		{
			name:     "missing approver id",
			contents: strings.Replace(registryDoc(defaultSpec()), "id: 1732826", "id: 0", 1),
			message:  "approver requires a positive immutable numeric id",
		},
		{
			name:     "unparsable date",
			contents: registryDoc(withDates(defaultSpec(), "2026-13-40", "2026-09-02")),
			message:  "created_at must be a 2006-01-02 date",
		},
		{
			name:     "review before creation",
			contents: registryDoc(withDates(defaultSpec(), "2026-09-02", "2026-09-01")),
			message:  "must not precede created_at",
		},
		{
			name:     "second document",
			contents: registryDoc(defaultSpec()) + "---\n" + registryDoc(defaultSpec()),
			message:  "exactly one document",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			records, err := decodeRegistry([]byte(testCase.contents))
			require.Error(t, err)
			assert.Nil(t, records)
			assert.Contains(t, err.Error(), testCase.message)
		})
	}
}

func withID(spec recordSpec, id string) recordSpec {
	spec.ID = id
	return spec
}

func withRule(spec recordSpec, rule string) recordSpec {
	spec.Rule = rule
	return spec
}

func withPath(spec recordSpec, path string) recordSpec {
	spec.Path = path
	return spec
}

func withDates(spec recordSpec, createdAt, lastReviewedAt string) recordSpec {
	spec.CreatedAt = createdAt
	spec.LastReviewedAt = lastReviewedAt
	return spec
}

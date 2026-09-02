package acceptedrisk

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanSuppressionsReadsOnlyRealComments(t *testing.T) {
	tree := fstest.MapFS{
		"cmd/example/example.go": &fstest.MapFile{Data: []byte(`package example

import "os"

// Open reads a caller-supplied path.
func Open(path string) (*os.File, error) {
	//gosec:disable G304 -- AR-2026-001 the caller owns the path.
	return os.Open(path)
}

// documentation mentions the directive form //gosec:disable Gxxx -- AR-YYYY-NNN reason
func documented() string {
	return "//gosec:disable G999 -- AR-2026-999 inside a string literal"
}
`)},
		"internal/example/second.go": &fstest.MapFile{Data: []byte(`package example

func second() int {
	//gosec:disable G118 -- AR-2026-006 owned and released elsewhere.
	return 1
}
`)},
		"vendor/other/vendored.go": &fstest.MapFile{Data: []byte(`package other

func vendored() int {
	//gosec:disable G304 -- AR-2026-777 vendored code is out of scope.
	return 1
}
`)},
		"testdata/fixture.go": &fstest.MapFile{Data: []byte(`package fixture

func fixture() int {
	//gosec:disable G304 -- AR-2026-888 fixture code is out of scope.
	return 1
}
`)},
		"docs/notes.md": &fstest.MapFile{Data: []byte("//gosec:disable G304 -- AR-2026-999 not Go source\n")},
	}

	found, err := scanSuppressions(tree)
	require.NoError(t, err)
	require.Len(t, found, 2)

	assert.Equal(t, "cmd/example/example.go", found[0].Path)
	assert.Equal(t, 7, found[0].Line)
	assert.Equal(t, "G304", found[0].Rule)
	assert.Equal(t, "AR-2026-001", found[0].RiskID)
	assert.Equal(t, "the caller owns the path.", found[0].Reason)
	assert.Equal(t, "cmd/example/example.go:7", found[0].location())

	assert.Equal(t, "internal/example/second.go", found[1].Path)
	assert.Equal(t, "AR-2026-006", found[1].RiskID)
}

func TestScanSuppressionsReportsMalformedDirectives(t *testing.T) {
	cases := map[string]string{
		"missing identifier": "\t//gosec:disable G304 -- the caller owns the path.\n",
		"missing separator":  "\t//gosec:disable G304 AR-2026-001 the caller owns the path.\n",
		"missing reason":     "\t//gosec:disable G304 -- AR-2026-001\n",
		"rule list":          "\t//gosec:disable G304,G302 -- AR-2026-001 two rules at once.\n",
		"malformed rule":     "\t//gosec:disable G30 -- AR-2026-001 short rule.\n",
		"malformed risk id":  "\t//gosec:disable G304 -- AR-26-1 short identifier.\n",
		"bare directive":     "\t//gosec:disable\n",
		"block comment form": "\t/*gosec:disable G304 -- AR-2026-001 block comment.*/\n",
	}

	for name, directive := range cases {
		t.Run(name, func(t *testing.T) {
			tree := fstest.MapFS{
				"example.go": &fstest.MapFile{Data: []byte("package example\n\nfunc example() int {\n" + directive + "\treturn 1\n}\n")},
			}
			found, err := scanSuppressions(tree)
			require.NoError(t, err)
			if name == "block comment form" {
				// A block comment is not a gosec directive at all, so it is not a
				// suppression the registry has to account for.
				assert.Empty(t, found)
				return
			}
			require.Len(t, found, 1)
			assert.Empty(t, found[0].RiskID, "malformed directive must not resolve to an identifier")
		})
	}
}

func TestScanSuppressionsFailsClosed(t *testing.T) {
	_, err := scanSuppressions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source file system is required")

	broken := fstest.MapFS{"broken.go": &fstest.MapFile{Data: []byte("package example\n\nfunc (\n")}}
	_, err = scanSuppressions(broken)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse broken.go")
}

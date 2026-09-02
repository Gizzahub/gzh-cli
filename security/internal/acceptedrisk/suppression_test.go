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
		"build/gen.go": &fstest.MapFile{Data: []byte(`package build

func generated() int {
	//gosec:disable G302 -- AR-2026-009 build output is still scanned by gosec.
	return 1
}
`)},
		"internal/build/helper.go": &fstest.MapFile{Data: []byte(`package build

func helper() int {
	//gosec:disable G304 -- AR-2026-010 a nested build directory is scanned too.
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

	found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
	require.NoError(t, err)

	paths := make([]string, 0, len(found))
	for _, current := range found {
		paths = append(paths, current.location())
	}
	// build/ and internal/build/ are present because the pinned scan flags do
	// not exclude them; vendor/, testdata/ and the .md file are absent.
	// require, not assert: the assertions below index into found, and a panic
	// in one test aborts the whole test binary rather than just that test.
	require.Equal(t, []string{
		"build/gen.go:4",
		"cmd/example/example.go:7",
		"internal/build/helper.go:4",
		"internal/example/second.go:4",
	}, paths)

	assert.Equal(t, "cmd/example/example.go", found[1].Path)
	assert.Equal(t, "G304", found[1].Rule)
	assert.Equal(t, "AR-2026-001", found[1].RiskID)
	assert.Equal(t, "the caller owns the path.", found[1].Reason)
}

// TestScanSuppressionsReportsBlanketSuppressions pins the premise the registry
// rests on. A blanket tag is honored by gosec's default grammar, carries no
// accepted-risk identifier and can never carry one, so the scanner has to see it
// in every comment shape gosec does, including a block comment, which the
// directive grammar never matches.
func TestScanSuppressionsReportsBlanketSuppressions(t *testing.T) {
	tree := fstest.MapFS{
		"cmd/example/example.go": &fstest.MapFile{Data: []byte(`package example

import "os"

func lineComment(p string) ([]byte, error) {
	//#nosec G304
	return os.ReadFile(p)
}

func spacedLineComment(p string) ([]byte, error) {
	// #nosec
	return os.ReadFile(p)
}

func blockComment(p string) ([]byte, error) {
	/*
	   #nosec G304 hidden inside a block comment
	*/
	return os.ReadFile(p)
}

func upperCase(p string) ([]byte, error) {
	//#NOSEC G304
	return os.ReadFile(p)
}

// documentation may name the #nosec form mid-sentence, and a line that merely
// mentions #nosec is not a suppression the pinned gosec would ever honor.
func documented(p string) ([]byte, error) {
	return os.ReadFile(p)
}

func stringLiteral() string {
	return "#nosec G304 inside a string literal"
}
`)},
	}

	found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
	require.NoError(t, err)

	locations := make([]string, 0, len(found))
	for _, current := range found {
		assert.True(t, current.Blanket, current.location())
		assert.Empty(t, current.RiskID, "the blanket form can never carry an identifier")
		locations = append(locations, current.location())
	}
	// The string literal is absent because comments are read from the parsed
	// comment groups. The uppercase spelling and the two prose mentions are
	// absent because gosec does not honor them either: it compares bytes and
	// accepts a tag only at the start of a comment line.
	// require, not assert: the assertion below indexes into found, and a panic
	// in one test aborts the whole test binary rather than just that test.
	require.Equal(t, []string{
		"cmd/example/example.go:6",
		"cmd/example/example.go:11",
		"cmd/example/example.go:17",
	}, locations)
	assert.Equal(t, "#nosec G304 hidden inside a block comment", found[2].Raw,
		"a violation must quote the suppression line, not the whole block comment")
}

// TestScanSuppressionsMatchesGosecTagPlacement pins the matcher to gosec's own
// placement rule rather than to a looser one.
//
// Every case was measured against the pinned gosec v2.28.0 binary with a fixture
// whose call sites differ only by the comment above them: the reported cases are
// the ones gosec left unsuppressed. The rule is byte comparison at the start of a
// line within a comment, after optional spaces and tabs, with no separator
// required after the tag. Matching more loosely than this would report comments
// gosec never acts on, and would make an ordinary sentence about the blanket
// form unwriteable anywhere in the repository.
func TestScanSuppressionsMatchesGosecTagPlacement(t *testing.T) {
	cases := []struct {
		name       string
		comment    string
		suppressed bool
	}{
		{name: "at the start of the comment", comment: "\t//#nosec G304\n", suppressed: true},
		{name: "after the marker's space", comment: "\t// #nosec G304\n", suppressed: true},
		{name: "after extra indentation", comment: "\t//   #nosec G304\n", suppressed: true},
		{name: "with no separator before the rule", comment: "\t//#nosecG304\n", suppressed: true},
		{name: "on a later line of a comment group", comment: "\t// the caller owns this path.\n\t// #nosec G304\n", suppressed: true},
		{name: "on a later line of a block comment", comment: "\t/* the caller owns this path.\n\t   #nosec G304 */\n", suppressed: true},
		{name: "spelled in uppercase", comment: "\t//#NOSEC G304\n", suppressed: false},
		{name: "named mid-sentence", comment: "\t// note: #nosec G304 would apply here\n", suppressed: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			tree := fstest.MapFS{
				"example.go": &fstest.MapFile{Data: []byte("package example\n\nfunc example() int {\n" + testCase.comment + "\treturn 1\n}\n")},
			}
			found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
			require.NoError(t, err)
			if !testCase.suppressed {
				assert.Empty(t, found, "gosec does not honor this placement, so neither does the scanner")
				return
			}
			require.Len(t, found, 1)
			assert.True(t, found[0].Blanket)
		})
	}
}

// TestGosecBlanketTokensTrackTheConfiguration pins the derivation to gosec
// v2.28.0's own construction, measured with the pinned binary: the live blanket
// tag is "#" followed by the configured global.nosec value, and the alternative
// key adds a second tag rather than replacing the first. gosec's built-in
// spelling stays in the set unconditionally because removing the setting makes
// it live again.
func TestGosecBlanketTokensTrackTheConfiguration(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   blanketTokens
	}{
		{
			name:   "no global section leaves the built-in tag live",
			config: `{}`,
			want:   blanketTokens{"#nosec"},
		},
		{
			name:   "the tracked setting renames the live tag",
			config: `{"global": {"nosec": false}}`,
			want:   blanketTokens{"#false", "#nosec"},
		},
		{
			name:   "any other value renames it to match",
			config: `{"global": {"nosec": "skipme"}}`,
			want:   blanketTokens{"#nosec", "#skipme"},
		},
		{
			name:   "the alternative key adds a tag beside the default one",
			config: `{"global": {"nosec": false, "#nosec": "skipme"}}`,
			want:   blanketTokens{"#false", "#nosec", "#skipme"},
		},
		{
			name:   "a value equal to the built-in spelling is not listed twice",
			config: `{"global": {"nosec": "nosec"}}`,
			want:   blanketTokens{"#nosec"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			tokens, err := gosecBlanketTokens([]byte(testCase.config))
			require.NoError(t, err)
			assert.Equal(t, testCase.want, tokens)
		})
	}
}

// TestScanSuppressionsFollowsTheConfiguredBlanketTag is the test that goes red if
// the scanner ever stops tracking .gosec.json. Under the tracked configuration
// the live tag is not gosec's built-in spelling, so a scanner pinned to a fixed
// token walks straight past a real suppression.
func TestScanSuppressionsFollowsTheConfiguredBlanketTag(t *testing.T) {
	tree := fstest.MapFS{
		"cmd/example/example.go": &fstest.MapFile{Data: []byte(`package example

import "os"

func configuredTag(p string) ([]byte, error) {
	//#false G304
	return os.ReadFile(p)
}

func alternativeTag(p string) ([]byte, error) {
	//#skipme G304
	return os.ReadFile(p)
}
`)},
	}

	found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
	require.NoError(t, err)
	require.Len(t, found, 1, "the tag this repository's configuration makes live must be reported")
	assert.Equal(t, "cmd/example/example.go:6", found[0].location())
	assert.True(t, found[0].Blanket)

	alternative, err := gosecBlanketTokens([]byte(`{"global": {"nosec": false, "#nosec": "skipme"}}`))
	require.NoError(t, err)
	found, err = scanSuppressions(tree, testScanScope(t), alternative)
	require.NoError(t, err)
	require.Len(t, found, 2, "a configured alternative tag is live alongside the default one")
	assert.Equal(t, "cmd/example/example.go:11", found[1].location())
}

func TestGosecBlanketTokensFailClosed(t *testing.T) {
	_, err := gosecBlanketTokens([]byte("global:\n  nosec: false\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse the pinned gosec configuration")

	_, err = scanSuppressions(fstest.MapFS{}, testScanScope(t), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "derived from the pinned gosec configuration")
}

// TestScanSuppressionsSeparatesBlanketFromRegistrableForms checks that one file
// carrying both grammars yields both findings, each classified correctly.
func TestScanSuppressionsSeparatesBlanketFromRegistrableForms(t *testing.T) {
	tree := fstest.MapFS{
		"cmd/example/example.go": &fstest.MapFile{Data: []byte(`package example

func registered() int {
	//gosec:disable G304 -- AR-2026-001 the caller owns the path.
	return 1
}

func blanket() int {
	//#nosec G304
	return 1
}
`)},
	}

	found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
	require.NoError(t, err)
	require.Len(t, found, 2)

	assert.False(t, found[0].Blanket)
	assert.Equal(t, "AR-2026-001", found[0].RiskID)
	assert.True(t, found[1].Blanket)
	assert.Equal(t, 9, found[1].Line)
}

// TestGosecScanScopeMatchesPinnedScanFlags fails if the scanner's skip set ever
// drifts wider than what the pinned scan actually excludes. A directive in a
// directory gosec reads but the scanner does not is exactly the
// suppression-unregistered hole the registry exists to close.
func TestGosecScanScopeMatchesPinnedScanFlags(t *testing.T) {
	scope := testScanScope(t)

	assert.Equal(t, []string{".git", "node_modules", "tmp", "vendor"}, scope.excludedDirs,
		"the skip set is derived from GOSEC_SCAN_FLAGS; update both together")

	for _, scanned := range []string{"bin", "build", "dist", "internal/build", "pkg/dist", "cmd/bin", "security"} {
		assert.False(t, scope.skipsDir(scanned), "gosec scans %s, so the scanner must too", scanned)
	}
	for _, skipped := range []string{"vendor", "internal/vendor", "node_modules", ".git", "tmp", "testdata", "pkg/testdata", ".idea", "_scratch"} {
		assert.True(t, scope.skipsDir(skipped), "gosec does not scan %s", skipped)
	}
}

func TestGosecScanScopeFailsClosed(t *testing.T) {
	_, err := gosecScanScope([]byte("GOLANGCI_LINT_VERSION := v2.13.1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GOSEC_SCAN_FLAGS is not declared")

	_, err = gosecScanScope([]byte("GOSEC_SCAN_FLAGS := -conf=.gosec.json -tests\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no -exclude-dir value")

	_, err = scanSuppressions(fstest.MapFS{}, scanScope{}, testBlanketTokens(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "derived from the pinned gosec scan flags")
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
			found, err := scanSuppressions(tree, testScanScope(t), testBlanketTokens(t))
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
	_, err := scanSuppressions(nil, testScanScope(t), testBlanketTokens(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source file system is required")

	broken := fstest.MapFS{"broken.go": &fstest.MapFile{Data: []byte("package example\n\nfunc (\n")}}
	_, err = scanSuppressions(broken, testScanScope(t), testBlanketTokens(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse broken.go")
}

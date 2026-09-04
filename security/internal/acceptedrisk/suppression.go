// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package acceptedrisk

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// directivePrefix introduces a standalone gosec suppression.
const directivePrefix = "//gosec:disable"

// directivePattern is the only accepted suppression form: the directive prefix,
// exactly one gosec rule identifier, a "--" separator, one AR-YYYY-NNN
// accepted-risk identifier and a non-empty reason.
//
// A comma-separated rule list is rejected on purpose: one suppression must map to
// exactly one accepted-risk record, and a list would hide the rules it covers.
var directivePattern = regexp.MustCompile(`^//gosec:disable[ \t]+(G\d{3})[ \t]+--[ \t]+(AR-20\d{2}-\d{3})[ \t]+(\S.*)$`)

// gosecNosecKey and gosecAlternativeKey are the two .gosec.json global settings
// that decide which comment tag the pinned gosec honors as a blanket
// suppression. The names are gosec's own: config.Nosec and
// config.NoSecAlternative in v2.28.0.
const (
	gosecNosecKey       = "nosec"
	gosecAlternativeKey = "#nosec"
)

// defaultBlanketTag is the tag gosec falls back to when the configuration
// declares no global.nosec value, built the way gosec builds it: config.NoSecTag
// prefixes "#" to the setting name itself.
const defaultBlanketTag = "#" + gosecNosecKey

// blanketTokens are the comment tags the pinned gosec honors as blanket
// suppressions: the form that names no accepted-risk record and can never name
// one, so it is never a registrable suppression.
//
// The set is derived from .gosec.json rather than spelled here, because that is
// where gosec derives it from. Hardcoding one tag is not merely incomplete under
// the tracked configuration, it is exactly inverted: the configuration renames
// the tag, so a scanner pinned to gosec's built-in spelling would watch the one
// form gosec ignores while the live form passed it unseen.
type blanketTokens []string

// gosecBlanketTokens derives the live blanket tags from the contents of
// .gosec.json, reproducing the construction in gosec v2.28.0's Analyzer.ignore:
//
//   - global.nosec supplies the default tag as "#" plus its value, so the
//     tracked setting of false makes "#false" the live tag and gosec's built-in
//     spelling inert. Values reach gosec through fmt.Sprintf("%v", v) over the
//     decoded JSON, so a bool arrives as "true" or "false".
//   - the global "#nosec" key supplies an alternative tag, again as "#" plus its
//     value. gosec honors it in addition to the default tag rather than instead
//     of it, so a configuration carrying both has two live blanket forms.
//
// gosec's built-in spelling is always included even when the configuration
// renames it. It is inert under the tracked configuration, but deleting that one
// line from .gosec.json makes it live again, and a scanner that had stopped
// watching it would fall silent at precisely the moment it began to matter.
//
// A configuration this cannot parse is an error rather than an empty set, for
// the same reason scanSuppressions rejects an underived scan scope: guessing the
// tag would let a live suppression through unreported.
func gosecBlanketTokens(gosecConfig []byte) (blanketTokens, error) {
	var config struct {
		Global map[string]any `json:"global"`
	}
	if err := json.Unmarshal(gosecConfig, &config); err != nil {
		return nil, fmt.Errorf("parse the pinned gosec configuration: %w", err)
	}

	tokens := []string{defaultBlanketTag}
	for _, key := range []string{gosecNosecKey, gosecAlternativeKey} {
		value, declared := config.Global[key]
		if !declared {
			continue
		}
		tokens = append(tokens, fmt.Sprintf("#%v", value))
	}

	sort.Strings(tokens)
	return slices.Compact(tokens), nil
}

// live reports whether one comment line, already stripped of its markers and of
// leading indentation, opens with a blanket tag.
//
// The rule mirrors gosec's findNoSecTag, and the fidelity is the whole point.
// gosec compares bytes and accepts the tag only at the start of a line within a
// comment group, after optional spaces and tabs. A looser rule is not merely
// noisy: matching case-insensitively and at any position makes an ordinary
// sentence about the blanket form unwriteable in any comment in the repository,
// including the comments in this file, which is a cost paid for findings gosec
// would never act on.
//
// No separator is required after the tag because gosec requires none: a tag
// immediately followed by a rule identifier still suppresses, so a scanner that
// demanded a following space would miss a live suppression.
func (tokens blanketTokens) live(line string) bool {
	for _, token := range tokens {
		if strings.HasPrefix(line, token) {
			return true
		}
	}
	return false
}

// gosecScanFlagsPattern locates the pinned scan flags in .make/tools.mk, and
// gosecExcludeDirPattern extracts the directories they exclude. The skip set is
// derived from that single source of truth rather than restated here: a skip set
// wider than the scan is a hole, because gosec honors a directive in any
// directory it scans and the registry would never see it.
var (
	gosecScanFlagsPattern  = regexp.MustCompile(`(?m)^GOSEC_SCAN_FLAGS[ \t]*:?=(.*)$`)
	gosecExcludeDirPattern = regexp.MustCompile(`-exclude-dir=([^\s]+)`)
)

// goToolchainSkippedDir reports directories `go list ./...` never loads, so the
// pinned scan never sees them either. gosec resolves ./... through the Go
// toolchain, so this is a property of the scan, not a local preference.
func goToolchainSkippedDir(name string) bool {
	return name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanScope is the set of directories the pinned gosec scan does not look at.
type scanScope struct {
	// excludedDirs are the -exclude-dir values, in sorted order, for reporting
	// and for the drift test.
	excludedDirs []string
	// excludedPatterns mirror gosec's own construction of those values, so this
	// scanner cannot skip a directory the scan actually reads.
	excludedPatterns []*regexp.Regexp
}

// gosecScanScope derives the scope from the contents of .make/tools.mk.
func gosecScanScope(toolsMakefile []byte) (scanScope, error) {
	flags := gosecScanFlagsPattern.FindSubmatch(toolsMakefile)
	if flags == nil {
		return scanScope{}, fmt.Errorf("GOSEC_SCAN_FLAGS is not declared in the pinned tool configuration")
	}

	matches := gosecExcludeDirPattern.FindAllStringSubmatch(string(flags[1]), -1)
	if len(matches) == 0 {
		return scanScope{}, fmt.Errorf("GOSEC_SCAN_FLAGS declares no -exclude-dir value")
	}

	scope := scanScope{excludedDirs: make([]string, 0, len(matches))}
	for _, match := range matches {
		dir := match[1]
		scope.excludedDirs = append(scope.excludedDirs, dir)
		// gosec builds `([\\/])?<dir>([\\/])?` and matches it against the
		// path, so the same construction is used here.
		pattern, err := regexp.Compile(`([\\/])?` + strings.ReplaceAll(regexp.QuoteMeta(dir), "/", `\/`) + `([\\/])?`)
		if err != nil {
			return scanScope{}, fmt.Errorf("compile gosec exclusion %q: %w", dir, err)
		}
		scope.excludedPatterns = append(scope.excludedPatterns, pattern)
	}
	sort.Strings(scope.excludedDirs)
	return scope, nil
}

// skipsDir reports whether the pinned scan ignores this directory. relPath is
// relative to the tree root, so a nested directory is judged on its full path
// rather than on its base name alone.
func (scope scanScope) skipsDir(relPath string) bool {
	for _, pattern := range scope.excludedPatterns {
		if pattern.MatchString(relPath) {
			return true
		}
	}
	return goToolchainSkippedDir(path.Base(relPath))
}

// suppression is one gosec suppression found in the repository source, in either
// grammar the scanner understands. A directive that does not match the accepted
// form yields a suppression with an empty RiskID so that the validator reports it
// instead of ignoring it.
type suppression struct {
	Path   string
	Line   int
	Rule   string
	RiskID string
	Reason string
	Raw    string
	// Blanket marks the blanket-tag form. It carries no accepted-risk identifier
	// and can never carry one, so it is a violation in its own right rather than
	// a directive that failed to parse. The zero value is the registrable
	// directive form, which is the grammar every other field describes.
	Blanket bool
}

// scanSuppressions collects every gosec suppression under fsys, in both the
// registrable directive grammar and the blanket-tag grammar. It reads through
// io/fs so that the traversal cannot escape the tree it was handed and so that
// tests can supply an in-memory tree.
//
// Comments are taken from the parsed comment groups rather than from raw lines,
// so a directive spelled inside a string literal is not mistaken for a real
// suppression.
//
// Both the scan scope and the blanket tags are required and are derived from the
// pinned tool configuration by the caller. This scanner reports what the pinned
// gosec would actually honor, and neither of those facts is knowable from this
// package's own source.
func scanSuppressions(fsys fs.FS, scope scanScope, tokens blanketTokens) ([]suppression, error) {
	if fsys == nil {
		return nil, fmt.Errorf("source file system is required")
	}
	if len(scope.excludedPatterns) == 0 {
		return nil, fmt.Errorf("scan scope must be derived from the pinned gosec scan flags")
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("blanket tags must be derived from the pinned gosec configuration")
	}

	found := make([]suppression, 0)
	walkErr := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != "." && scope.skipsDir(name) {
				return fs.SkipDir
			}
			return nil
		}
		if path.Ext(name) != ".go" {
			return nil
		}
		contents, readErr := fs.ReadFile(fsys, name)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", name, readErr)
		}
		fileFound, parseErr := suppressionsInFile(name, contents, tokens)
		if parseErr != nil {
			return parseErr
		}
		found = append(found, fileFound...)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sortSuppressions(found)
	return found, nil
}

func suppressionsInFile(name string, contents []byte, tokens blanketTokens) ([]suppression, error) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, name, contents, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	found := make([]suppression, 0)
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			startLine := fileSet.Position(comment.Slash).Line
			found = append(found, tokens.blanketSuppressions(name, startLine, comment.Text)...)

			text := strings.TrimRight(comment.Text, " \t")
			if !strings.HasPrefix(text, directivePrefix) {
				continue
			}
			current := suppression{Path: name, Line: startLine, Raw: text}
			if matches := directivePattern.FindStringSubmatch(text); matches != nil {
				current.Rule = matches[1]
				current.RiskID = matches[2]
				current.Reason = strings.TrimSpace(matches[3])
			}
			found = append(found, current)
		}
	}
	return found, nil
}

// blanketSuppressions reports every blanket suppression written inside one
// comment. The comment markers are stripped first, because gosec matches against
// ast.CommentGroup.Text, which has already removed them, and each line is then
// judged on its own, because gosec judges each line on its own. An occurrence is
// attributed to the line it is written on rather than to the line a block
// comment opens on.
func (tokens blanketTokens) blanketSuppressions(name string, startLine int, text string) []suppression {
	body := text
	switch {
	case strings.HasPrefix(body, "//"):
		body = body[2:]
	case strings.HasPrefix(body, "/*"):
		body = strings.TrimSuffix(body[2:], "*/")
	}

	var found []suppression
	for offset, line := range strings.Split(body, "\n") {
		if !tokens.live(strings.TrimLeft(line, " \t")) {
			continue
		}
		found = append(found, suppression{
			Path:    name,
			Line:    startLine + offset,
			Raw:     strings.TrimSpace(line),
			Blanket: true,
		})
	}
	return found
}

func sortSuppressions(values []suppression) {
	sort.Slice(values, func(left, right int) bool {
		if values[left].Path != values[right].Path {
			return values[left].Path < values[right].Path
		}
		if values[left].Line != values[right].Line {
			return values[left].Line < values[right].Line
		}
		// One line can carry both grammars, so the raw text is the final
		// tiebreak; without it the order of two findings on one line would
		// depend on the sort implementation.
		return values[left].Raw < values[right].Raw
	})
}

// location renders a stable subject for a suppression violation.
func (current suppression) location() string {
	return fmt.Sprintf("%s:%d", current.Path, current.Line)
}

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package acceptedrisk

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"regexp"
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

// skippedDirs are never scanned for suppressions. They hold vendored, generated
// or build output rather than repository-owned source.
var skippedDirs = map[string]struct{}{
	".git":         {},
	"bin":          {},
	"build":        {},
	"dist":         {},
	"node_modules": {},
	"testdata":     {},
	"tmp":          {},
	"vendor":       {},
}

// suppression is one standalone gosec directive found in the repository source.
// A directive that does not match the accepted form yields a suppression with an
// empty RiskID so that the validator reports it instead of ignoring it.
type suppression struct {
	Path   string
	Line   int
	Rule   string
	RiskID string
	Reason string
	Raw    string
}

// scanSuppressions collects every standalone gosec directive under fsys. It reads
// through io/fs so that the traversal cannot escape the tree it was handed and so
// that tests can supply an in-memory tree.
//
// Comments are taken from the parsed comment groups rather than from raw lines,
// so a directive spelled inside a string literal is not mistaken for a real
// suppression.
func scanSuppressions(fsys fs.FS) ([]suppression, error) {
	if fsys == nil {
		return nil, fmt.Errorf("source file system is required")
	}

	found := make([]suppression, 0)
	walkErr := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != "." && isSkippedDir(entry.Name()) {
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
		fileFound, parseErr := suppressionsInFile(name, contents)
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

func suppressionsInFile(name string, contents []byte) ([]suppression, error) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, name, contents, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	found := make([]suppression, 0)
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			text := strings.TrimRight(comment.Text, " \t")
			if !strings.HasPrefix(text, directivePrefix) {
				continue
			}
			current := suppression{Path: name, Line: fileSet.Position(comment.Slash).Line, Raw: text}
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

func isSkippedDir(name string) bool {
	_, skipped := skippedDirs[name]
	return skipped
}

func sortSuppressions(values []suppression) {
	sort.Slice(values, func(left, right int) bool {
		if values[left].Path != values[right].Path {
			return values[left].Path < values[right].Path
		}
		return values[left].Line < values[right].Line
	})
}

// location renders a stable subject for a suppression violation.
func (current suppression) location() string {
	return fmt.Sprintf("%s:%d", current.Path, current.Line)
}

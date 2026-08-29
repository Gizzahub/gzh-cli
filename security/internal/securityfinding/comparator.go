// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package securityfinding compares prevalidated security-scan finding reports.
package securityfinding

import (
	"fmt"
	"go/scanner"
	"go/token"
	"path"
	"sort"
	"strings"
	"unicode"
)

type expectation struct {
	BaselineSHA    string
	CandidateSHA   string
	ScannerVersion string
	ConfigHash     string
	FlagsHash      string
}

type reportEnvelope struct {
	SourceSHA      string
	ScannerVersion string
	ConfigHash     string
	FlagsHash      string
	Findings       []finding
}

type finding struct {
	Rule          string
	Path          string
	StartLine     int
	StartColumn   int
	Message       string
	SourceSnippet string
}

type comparison struct {
	BaselineRawCount  int
	CandidateRawCount int
	Known             []finding
	New               []finding
	Unclassified      []finding
}

func compare(expected expectation, baseline, candidate *reportEnvelope) (comparison, error) {
	if err := expected.validate(); err != nil {
		return comparison{}, err
	}
	baselineFindings, baselineCount, err := normalizeReport(baseline, expected.BaselineSHA, expected, "baseline")
	if err != nil {
		return comparison{}, err
	}
	candidateFindings, candidateCount, err := normalizeReport(candidate, expected.CandidateSHA, expected, "candidate")
	if err != nil {
		return comparison{}, err
	}

	result := comparison{
		BaselineRawCount:  baselineCount,
		CandidateRawCount: candidateCount,
		Known:             []finding{},
		New:               []finding{},
		Unclassified:      []finding{},
	}
	baselineBySemanticKey := make(map[string][]finding, len(baselineFindings))
	for _, current := range baselineFindings {
		key, keyErr := current.semanticKey()
		if keyErr != nil {
			return comparison{}, keyErr
		}
		baselineBySemanticKey[key] = append(baselineBySemanticKey[key], current)
	}

	used := make(map[string]bool, len(baselineBySemanticKey))
	for _, current := range candidateFindings {
		key, keyErr := current.semanticKey()
		if keyErr != nil {
			return comparison{}, keyErr
		}
		matches := baselineBySemanticKey[key]
		switch {
		case len(matches) == 0:
			result.New = append(result.New, current)
		case len(matches) != 1 || used[key]:
			// A many-to-one match would hide a changed finding, so do not classify it as known.
			result.Unclassified = append(result.Unclassified, current)
		default:
			used[key] = true
			result.Known = append(result.Known, current)
		}
	}

	sortFindings(result.Known)
	sortFindings(result.New)
	sortFindings(result.Unclassified)
	return result, nil
}

func (expected expectation) validate() error {
	if strings.TrimSpace(expected.BaselineSHA) == "" || strings.TrimSpace(expected.CandidateSHA) == "" ||
		strings.TrimSpace(expected.ScannerVersion) == "" || strings.TrimSpace(expected.ConfigHash) == "" ||
		strings.TrimSpace(expected.FlagsHash) == "" {
		return fmt.Errorf("expectation must include baseline SHA, candidate SHA, scanner version, config hash, and flags hash")
	}
	return nil
}

func normalizeReport(report *reportEnvelope, expectedSHA string, expected expectation, label string) ([]finding, int, error) {
	if report == nil {
		return nil, 0, fmt.Errorf("%s report is required", label)
	}
	if report.Findings == nil {
		return nil, 0, fmt.Errorf("%s findings must be an explicit array", label)
	}
	if report.SourceSHA != expectedSHA {
		return nil, 0, fmt.Errorf("%s source SHA does not match its trusted revision", label)
	}
	if report.ScannerVersion != expected.ScannerVersion || report.ConfigHash != expected.ConfigHash || report.FlagsHash != expected.FlagsHash {
		return nil, 0, fmt.Errorf("%s scanner version, config hash, and flags hash must match trusted policy", label)
	}

	byTopology := make(map[string]finding, len(report.Findings))
	for index, current := range report.Findings {
		canonical, err := current.canonical()
		if err != nil {
			return nil, 0, fmt.Errorf("%s finding %d: %w", label, index, err)
		}
		topologyKey := canonical.topologyKey()
		if existing, found := byTopology[topologyKey]; found {
			existingKey, existingErr := existing.semanticKey()
			currentKey, currentErr := canonical.semanticKey()
			if existingErr != nil || currentErr != nil {
				return nil, 0, fmt.Errorf("%s finding %d has invalid semantic content", label, index)
			}
			if existingKey != currentKey {
				return nil, 0, fmt.Errorf("%s finding %d conflicts with another finding at the same location", label, index)
			}
			continue
		}
		byTopology[topologyKey] = canonical
	}

	unique := make([]finding, 0, len(byTopology))
	for _, current := range byTopology {
		unique = append(unique, current)
	}
	sortFindings(unique)
	return unique, len(report.Findings), nil
}

func (current finding) canonical() (finding, error) {
	if strings.TrimSpace(current.Rule) == "" || strings.TrimSpace(current.Message) == "" || strings.TrimSpace(current.SourceSnippet) == "" {
		return finding{}, fmt.Errorf("rule, message, and source snippet are required")
	}
	if current.StartLine <= 0 || current.StartColumn <= 0 {
		return finding{}, fmt.Errorf("start line and start column must be positive")
	}
	canonicalPath, err := canonicalPath(current.Path)
	if err != nil {
		return finding{}, err
	}
	snippet, err := normalizedSnippet(current.SourceSnippet)
	if err != nil {
		return finding{}, err
	}
	return finding{
		Rule:          strings.Join(strings.Fields(current.Rule), " "),
		Path:          canonicalPath,
		StartLine:     current.StartLine,
		StartColumn:   current.StartColumn,
		Message:       strings.Join(strings.Fields(current.Message), " "),
		SourceSnippet: snippet,
	}, nil
}

func (current finding) topologyKey() string {
	return strings.Join([]string{current.Rule, current.Path, fmt.Sprint(current.StartLine), fmt.Sprint(current.StartColumn)}, "\x00")
}

func (current finding) semanticKey() (string, error) {
	if current.SourceSnippet == "" {
		return "", fmt.Errorf("source snippet is required")
	}
	return strings.Join([]string{current.Rule, current.Path, current.Message, current.SourceSnippet}, "\x00"), nil
}

func canonicalPath(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.IndexFunc(trimmed, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("path must be non-empty and contain no control characters")
	}
	replaced := strings.ReplaceAll(trimmed, "\\", "/")
	if strings.HasPrefix(replaced, "//") || path.IsAbs(replaced) || isWindowsDrivePath(replaced) || hasURIScheme(replaced) {
		return "", fmt.Errorf("path must be repository-relative")
	}
	for _, segment := range strings.Split(replaced, "/") {
		if segment == ".." {
			return "", fmt.Errorf("path must not traverse outside the repository")
		}
	}
	cleaned := path.Clean(replaced)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("path must be repository-relative")
	}
	return cleaned, nil
}

func isWindowsDrivePath(value string) bool {
	return len(value) >= 2 && isASCIILetter(value[0]) && value[1] == ':'
}

func hasURIScheme(value string) bool {
	for index, current := range value {
		switch {
		case current == ':':
			return index > 0
		case !(current == '+' || current == '-' || current == '.' || unicode.IsLetter(current) || (index > 0 && unicode.IsDigit(current))):
			return false
		}
	}
	return false
}

func normalizedSnippet(value string) (string, error) {
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("snippet.go", -1, len(value))
	var scan scanner.Scanner
	var scanErr error
	scan.Init(file, []byte(value), func(_ token.Position, message string) {
		scanErr = fmt.Errorf("invalid source snippet: %s", message)
	}, 0)

	parts := make([]string, 0)
	for _, current, literal := scan.Scan(); current != token.EOF; _, current, literal = scan.Scan() {
		if current == token.SEMICOLON && literal == "\n" {
			continue
		}
		if literal == "" {
			literal = current.String()
		}
		parts = append(parts, literal)
	}
	if scanErr != nil {
		return "", scanErr
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("source snippet contains no Go tokens")
	}
	return strings.Join(parts, "\x00"), nil
}

func sortFindings(findings []finding) {
	sort.Slice(findings, func(left, right int) bool {
		leftKey := strings.Join([]string{findings[left].topologyKey(), findings[left].Message, findings[left].SourceSnippet}, "\x00")
		rightKey := strings.Join([]string{findings[right].topologyKey(), findings[right].Message, findings[right].SourceSnippet}, "\x00")
		return leftKey < rightKey
	})
}

func isASCIILetter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

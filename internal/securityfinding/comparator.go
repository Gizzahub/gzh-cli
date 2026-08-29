// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package securityfinding compares trusted security-scan finding reports.
package securityfinding

import (
	"fmt"
	"path"
	"strings"
)

// metadata identifies the exact trusted scan input and configuration.
// It is deliberately private until a caller owns report ingestion.
type metadata struct {
	SourceSHA      string `json:"source_sha"`
	ScannerVersion string `json:"scanner_version"`
	ConfigHash     string `json:"config_hash"`
	FlagsHash      string `json:"flags_hash"`
}

type reportEnvelope struct {
	metadata
	Findings []finding `json:"findings"`
}

type finding struct {
	Rule          string `json:"rule"`
	Path          string `json:"path"`
	Line          int    `json:"line"`
	Message       string `json:"message"`
	SourceSnippet string `json:"source_snippet"`
}

type comparison struct {
	Known        []finding
	New          []finding
	Unclassified []finding
}

func compare(expected metadata, baseline, candidate reportEnvelope) (comparison, error) {
	if err := validateMetadata(expected, "expected metadata"); err != nil {
		return comparison{}, err
	}
	if err := baseline.validate(expected, "baseline"); err != nil {
		return comparison{}, err
	}
	if err := candidate.validate(expected, "candidate"); err != nil {
		return comparison{}, err
	}

	baselineByKey := make(map[string][]finding)
	for _, current := range collapseRawDuplicates(baseline.Findings) {
		if !current.classifiable() {
			continue
		}
		key := current.semanticKey()
		baselineByKey[key] = append(baselineByKey[key], current)
	}

	result := comparison{}
	used := make(map[string]bool)
	for _, current := range collapseRawDuplicates(candidate.Findings) {
		if !current.classifiable() {
			result.Unclassified = append(result.Unclassified, current)
			continue
		}

		key := current.semanticKey()
		matches := baselineByKey[key]
		switch {
		case len(matches) == 0:
			result.New = append(result.New, current)
		case len(matches) != 1 || used[key]:
			// Multiple locations with the same semantic key cannot be reconciled safely.
			result.Unclassified = append(result.Unclassified, current)
		default:
			used[key] = true
			result.Known = append(result.Known, current)
		}
	}

	return result, nil
}

func (envelope reportEnvelope) validate(expected metadata, label string) error {
	if err := validateMetadata(envelope.metadata, label+" metadata"); err != nil {
		return err
	}
	if envelope.metadata != expected {
		return fmt.Errorf("%s metadata does not match trusted metadata", label)
	}
	return nil
}

func validateMetadata(current metadata, label string) error {
	if strings.TrimSpace(current.SourceSHA) == "" ||
		strings.TrimSpace(current.ScannerVersion) == "" ||
		strings.TrimSpace(current.ConfigHash) == "" ||
		strings.TrimSpace(current.FlagsHash) == "" {
		return fmt.Errorf("%s must include source SHA, scanner version, config hash, and flags hash", label)
	}
	return nil
}

func collapseRawDuplicates(findings []finding) []finding {
	seen := make(map[string]struct{}, len(findings))
	unique := make([]finding, 0, len(findings))
	for _, current := range findings {
		key := current.rawKey()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, current)
	}
	return unique
}

func (current finding) classifiable() bool {
	if strings.TrimSpace(current.Rule) == "" || strings.TrimSpace(current.Message) == "" || strings.TrimSpace(current.SourceSnippet) == "" {
		return false
	}
	_, ok := normalizedPath(current.Path)
	return ok
}

func (current finding) semanticKey() string {
	return strings.Join([]string{
		strings.TrimSpace(current.Rule),
		mustNormalizePath(current.Path),
		normalize(current.Message),
		normalize(current.SourceSnippet),
	}, "\x00")
}

func (current finding) rawKey() string {
	return strings.Join([]string{
		current.Rule,
		current.Path,
		fmt.Sprint(current.Line),
		current.Message,
		current.SourceSnippet,
	}, "\x00")
}

func normalize(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func normalizedPath(value string) (string, bool) {
	cleaned := path.Clean(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"))
	if cleaned == "." || path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}
	if len(cleaned) >= 3 && cleaned[1] == ':' && cleaned[2] == '/' && isASCIILetter(cleaned[0]) {
		return "", false
	}
	return cleaned, true
}

func mustNormalizePath(value string) string {
	cleaned, ok := normalizedPath(value)
	if !ok {
		panic("semantic key requested for an unclassifiable path")
	}
	return cleaned
}

func isASCIILetter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

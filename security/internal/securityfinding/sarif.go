// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package securityfinding

import (
	"encoding/json"
	"fmt"
	"strings"
)

const sarifVersion210 = "2.1.0"

// sourceResolver는 SARIF 결과에 사용할 수 있는 snippet이 없을 때만 source line을 제공한다.
// 구현체는 이미 선택된 repository revision에서 읽어야 하며, 이 adapter는 revision을 선택하거나 인증하지 않는다.
type sourceResolver interface {
	ReadSource(path string) ([]byte, error)
}

// ingestedReport는 comparator에 전달할 수 있는 SARIF report의 비신뢰 부분이다.
// scanner policy나 source revision metadata를 의도적으로 포함하지 않는다.
type ingestedReport struct {
	Findings            []finding
	RawResultCount      int
	LocationUniqueCount int
}

type sarifDocument struct {
	Version string          `json:"version"`
	Runs    json.RawMessage `json:"runs"`
}

type sarifRun struct {
	Tool    *sarifTool      `json:"tool"`
	Results json.RawMessage `json:"results"`
}

type sarifTool struct {
	Driver *sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name string `json:"name"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Message   *sarifMessage   `json:"message"`
	Locations json.RawMessage `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation *sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation *sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int           `json:"startLine"`
	StartColumn int           `json:"startColumn"`
	Snippet     *sarifMessage `json:"snippet"`
}

func ingestSARIF(contents []byte, resolver sourceResolver) (ingestedReport, error) {
	var document sarifDocument
	if err := json.Unmarshal(contents, &document); err != nil {
		return ingestedReport{}, fmt.Errorf("decode SARIF: %w", err)
	}
	if document.Version != sarifVersion210 {
		return ingestedReport{}, fmt.Errorf("SARIF version must be %s", sarifVersion210)
	}
	if !isExplicitJSONArray(document.Runs) {
		return ingestedReport{}, fmt.Errorf("SARIF runs must be an explicit array")
	}

	var runs []sarifRun
	if err := json.Unmarshal(document.Runs, &runs); err != nil {
		return ingestedReport{}, fmt.Errorf("decode SARIF runs: %w", err)
	}
	if len(runs) != 1 || runs[0].Tool == nil || runs[0].Tool.Driver == nil || runs[0].Tool.Driver.Name != "gosec" {
		return ingestedReport{}, fmt.Errorf("SARIF must contain exactly one gosec run")
	}
	if !isExplicitJSONArray(runs[0].Results) {
		return ingestedReport{}, fmt.Errorf("gosec results must be an explicit array")
	}

	var results []sarifResult
	if err := json.Unmarshal(runs[0].Results, &results); err != nil {
		return ingestedReport{}, fmt.Errorf("decode gosec results: %w", err)
	}

	report := ingestedReport{
		Findings:       make([]finding, 0, len(results)),
		RawResultCount: len(results),
	}
	uniqueTopology := make(map[string]struct{}, len(results))
	for index, result := range results {
		current, err := findingFromSARIFResult(result, resolver)
		if err != nil {
			return ingestedReport{}, fmt.Errorf("gosec result %d: %w", index, err)
		}
		report.Findings = append(report.Findings, current)
		uniqueTopology[current.topologyKey()] = struct{}{}
	}
	report.LocationUniqueCount = len(uniqueTopology)
	return report, nil
}

func findingFromSARIFResult(result sarifResult, resolver sourceResolver) (finding, error) {
	if result.Message == nil {
		return finding{}, fmt.Errorf("message is required")
	}
	if !isExplicitJSONArray(result.Locations) {
		return finding{}, fmt.Errorf("locations must be an explicit array")
	}
	var locations []sarifLocation
	if err := json.Unmarshal(result.Locations, &locations); err != nil {
		return finding{}, fmt.Errorf("decode locations: %w", err)
	}
	if len(locations) != 1 {
		return finding{}, fmt.Errorf("exactly one location is required")
	}
	location := locations[0].PhysicalLocation
	if location == nil || location.ArtifactLocation == nil || location.Region == nil {
		return finding{}, fmt.Errorf("physical location, artifact location, and region are required")
	}
	canonicalPath, err := canonicalSARIFURI(location.ArtifactLocation.URI)
	if err != nil {
		return finding{}, err
	}
	if location.Region.StartLine <= 0 || location.Region.StartColumn <= 0 {
		return finding{}, fmt.Errorf("region start line and start column must be positive")
	}

	snippet, err := usableSnippet(location.Region.Snippet)
	if err != nil {
		snippet, err = resolveSnippet(resolver, canonicalPath, location.Region.StartLine)
		if err != nil {
			return finding{}, err
		}
	}
	canonical, err := (finding{
		Rule:          result.RuleID,
		Path:          canonicalPath,
		StartLine:     location.Region.StartLine,
		StartColumn:   location.Region.StartColumn,
		Message:       result.Message.Text,
		SourceSnippet: snippet,
	}).canonical()
	if err != nil {
		return finding{}, err
	}
	// comparator가 신뢰 메타데이터 경계에서 snippet token normalization을 수행한다.
	canonical.SourceSnippet = snippet
	return canonical, nil
}

func canonicalSARIFURI(value string) (string, error) {
	if value != strings.TrimSpace(value) || strings.Contains(value, "\\") {
		return "", fmt.Errorf("artifact URI must be a trimmed repository-relative slash path")
	}
	return canonicalPath(value)
}

func usableSnippet(snippet *sarifMessage) (string, error) {
	if snippet == nil || strings.TrimSpace(snippet.Text) == "" {
		return "", fmt.Errorf("snippet is missing")
	}
	if _, err := normalizedSnippet(snippet.Text); err != nil {
		return "", err
	}
	return snippet.Text, nil
}

func resolveSnippet(resolver sourceResolver, repositoryPath string, startLine int) (string, error) {
	canonicalPath, err := canonicalPath(repositoryPath)
	if err != nil {
		return "", err
	}
	if resolver == nil {
		return "", fmt.Errorf("source resolver is required for a missing or invalid snippet")
	}
	contents, err := resolver.ReadSource(canonicalPath)
	if err != nil {
		return "", fmt.Errorf("resolve source %q: %w", canonicalPath, err)
	}
	lines := strings.Split(strings.ReplaceAll(string(contents), "\r\n", "\n"), "\n")
	if startLine > len(lines) {
		return "", fmt.Errorf("source line %d is outside %q", startLine, canonicalPath)
	}
	snippet := lines[startLine-1]
	if _, err := normalizedSnippet(snippet); err != nil {
		return "", fmt.Errorf("source line %d in %q is not a usable Go snippet: %w", startLine, canonicalPath, err)
	}
	return snippet, nil
}

func isExplicitJSONArray(value json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(value))
	return len(trimmed) >= 2 && trimmed[0] == '[' && trimmed[len(trimmed)-1] == ']'
}

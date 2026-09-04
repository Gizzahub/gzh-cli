// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package acceptedrisk

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

const (
	// dateLayout is the only accepted date form. A date carries no clock or zone,
	// so cadence arithmetic cannot drift with the machine that runs the check.
	dateLayout = "2006-01-02"

	// pendingApprovalSentinel marks a record whose risk analysis is written but
	// whose owner approval has not been given. It is rejected by the validator;
	// it exists so an unapproved record is explicit instead of looking like a
	// truncated or corrupted SHA.
	pendingApprovalSentinel = "pending-owner-approval"
)

var (
	// riskIDPattern is the immutable accepted-risk identifier form AR-YYYY-NNN.
	riskIDPattern = regexp.MustCompile(`^AR-(20\d{2})-(\d{3})$`)

	// gosecRulePattern is a gosec rule identifier such as G304.
	gosecRulePattern = regexp.MustCompile(`^G\d{3}$`)

	// commitSHAPattern is a full 40-hex git object name. An abbreviated SHA is
	// rejected because it is not a stable identity.
	commitSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// registry is the tracked accepted-risk registry read from
// security/accepted-risks.yaml.
type registry struct {
	Version int            `yaml:"version"`
	Records []registryYAML `yaml:"records"`
}

// registryYAML is the on-disk shape of one record. Dates stay as strings here so
// that a malformed date is reported by canonicalization rather than by the YAML
// decoder's own timestamp handling.
type registryYAML struct {
	ID                  string           `yaml:"id"`
	Rule                string           `yaml:"rule"`
	Path                string           `yaml:"path"`
	Symbol              string           `yaml:"symbol"`
	Threat              string           `yaml:"threat"`
	CompensatingControl string           `yaml:"compensating_control"`
	Owner               string           `yaml:"owner"`
	Approver            registryApprover `yaml:"approver"`
	CreatedAt           string           `yaml:"created_at"`
	LastReviewedAt      string           `yaml:"last_reviewed_at"`
	Evidence            registryEvidence `yaml:"evidence"`
	TestEvidence        string           `yaml:"test_evidence"`
}

type registryApprover struct {
	ID    int64  `yaml:"id"`
	Login string `yaml:"login"`
}

type registryEvidence struct {
	Type string `yaml:"type"`
	SHA  string `yaml:"sha"`
}

// record is a canonicalized registry record with parsed dates.
type record struct {
	ID                  string
	Rule                string
	Path                string
	Symbol              string
	Threat              string
	CompensatingControl string
	Owner               string
	Approver            registryApprover
	CreatedAt           time.Time
	LastReviewedAt      time.Time
	Evidence            registryEvidence
	TestEvidence        string
}

// decodeRegistry parses the accepted-risk registry. Like the policy decoder it is
// strictly closed, and it rejects duplicate identifiers at load time because a
// duplicate identifier makes every later suppression-to-record lookup ambiguous.
func decodeRegistry(contents []byte) ([]record, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)

	var decoded registry
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode accepted-risk registry: %w", err)
	}
	if err := requireSingleDocument(decoder, "accepted-risk registry"); err != nil {
		return nil, err
	}
	if decoded.Version != supportedSchemaVersion {
		return nil, fmt.Errorf("accepted-risk registry version must be %d", supportedSchemaVersion)
	}
	if decoded.Records == nil {
		return nil, fmt.Errorf("accepted-risk registry records must be an explicit list")
	}

	records := make([]record, 0, len(decoded.Records))
	seen := make(map[string]struct{}, len(decoded.Records))
	for index, entry := range decoded.Records {
		canonical, err := entry.canonical()
		if err != nil {
			return nil, fmt.Errorf("accepted-risk record %d: %w", index, err)
		}
		if _, duplicated := seen[canonical.ID]; duplicated {
			return nil, fmt.Errorf("accepted-risk record %d: duplicate identifier %s", index, canonical.ID)
		}
		seen[canonical.ID] = struct{}{}
		records = append(records, canonical)
	}
	return records, nil
}

func (entry registryYAML) canonical() (record, error) {
	createdYear, err := entry.validateIdentity()
	if err != nil {
		return record{}, err
	}
	canonicalPath, err := canonicalSourcePath(entry.Path)
	if err != nil {
		return record{}, err
	}
	if err := entry.validateNarrative(); err != nil {
		return record{}, err
	}

	createdAt, err := time.Parse(dateLayout, entry.CreatedAt)
	if err != nil {
		return record{}, fmt.Errorf("created_at must be a %s date", dateLayout)
	}
	lastReviewedAt, err := time.Parse(dateLayout, entry.LastReviewedAt)
	if err != nil {
		return record{}, fmt.Errorf("last_reviewed_at must be a %s date", dateLayout)
	}
	if lastReviewedAt.Before(createdAt) {
		return record{}, fmt.Errorf("last_reviewed_at must not precede created_at")
	}
	if createdAt.Year() != createdYear {
		return record{}, fmt.Errorf("identifier year %d does not match created_at year %d", createdYear, createdAt.Year())
	}

	return record{
		ID:                  entry.ID,
		Rule:                entry.Rule,
		Path:                canonicalPath,
		Symbol:              strings.TrimSpace(entry.Symbol),
		Threat:              strings.TrimSpace(entry.Threat),
		CompensatingControl: strings.TrimSpace(entry.CompensatingControl),
		Owner:               strings.TrimSpace(entry.Owner),
		Approver:            entry.Approver,
		CreatedAt:           createdAt,
		LastReviewedAt:      lastReviewedAt,
		Evidence: registryEvidence{
			Type: strings.TrimSpace(entry.Evidence.Type),
			SHA:  strings.TrimSpace(entry.Evidence.SHA),
		},
		TestEvidence: strings.TrimSpace(entry.TestEvidence),
	}, nil
}

// validateIdentity checks the immutable identifier, the rule and the approver
// reference, and returns the year encoded in the identifier.
func (entry registryYAML) validateIdentity() (int, error) {
	matches := riskIDPattern.FindStringSubmatch(entry.ID)
	if matches == nil {
		return 0, fmt.Errorf("identifier %q must have the form AR-YYYY-NNN", entry.ID)
	}
	year, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("identifier %q has an unreadable year", entry.ID)
	}
	if !gosecRulePattern.MatchString(entry.Rule) {
		return 0, fmt.Errorf("rule %q must be a gosec rule identifier", entry.Rule)
	}
	if entry.Approver.ID <= 0 || strings.TrimSpace(entry.Approver.Login) == "" {
		return 0, fmt.Errorf("approver requires a positive immutable numeric id and a login")
	}
	return year, nil
}

// validateNarrative requires the human-readable justification fields. A record
// without a threat, a compensating control, an owner or test evidence is not a
// risk acceptance, it is an undocumented suppression.
func (entry registryYAML) validateNarrative() error {
	required := map[string]string{
		"symbol":               entry.Symbol,
		"threat":               entry.Threat,
		"compensating_control": entry.CompensatingControl,
		"owner":                entry.Owner,
		"test_evidence":        entry.TestEvidence,
		"evidence.type":        entry.Evidence.Type,
		"evidence.sha":         entry.Evidence.SHA,
	}
	for _, field := range sortedKeys(required) {
		if strings.TrimSpace(required[field]) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	return nil
}

// reviewBy is derived, never stored, so a record cannot claim a review date that
// disagrees with the cadence the policy owner approved.
func (current record) reviewBy(cadence policyCadence) time.Time {
	return current.LastReviewedAt.AddDate(0, 0, cadence.ReviewIntervalDays)
}

// hardSunset is measured from the initial created_at, not from the latest review,
// so re-reviewing a record cannot extend it past the sunset.
func (current record) hardSunset(cadence policyCadence) time.Time {
	return current.CreatedAt.AddDate(0, 0, cadence.HardSunsetDays)
}

// canonicalSourcePath accepts only a repository-relative slash path to a Go
// source file. It is deliberately independent of the SARIF path rules in
// securityfinding: this registry addresses tracked source files only.
func canonicalSourcePath(value string) (string, error) {
	if value != strings.TrimSpace(value) || value == "" {
		return "", fmt.Errorf("path must be a trimmed non-empty repository-relative path")
	}
	if strings.ContainsRune(value, '\\') || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("path must not contain backslashes or control characters")
	}
	if path.IsAbs(value) || strings.HasPrefix(value, "//") || isWindowsDrivePath(value) {
		return "", fmt.Errorf("path must be repository-relative")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("path must not contain empty or relative segments")
		}
	}
	if !strings.HasSuffix(value, ".go") {
		return "", fmt.Errorf("path must name a Go source file")
	}
	return value, nil
}

func isWindowsDrivePath(value string) bool {
	return len(value) >= 2 && value[1] == ':' &&
		(value[0] >= 'a' && value[0] <= 'z' || value[0] >= 'A' && value[0] <= 'Z')
}

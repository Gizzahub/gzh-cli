// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package acceptedrisk validates the trusted-base accepted-risk registry and its
// linkage to the standalone gosec suppressions present in the repository source.
package acceptedrisk

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// supportedSchemaVersion is the only schema version this validator accepts.
	// A newer file must be handled by a newer validator rather than by a lenient
	// reader that silently ignores fields it does not understand.
	supportedSchemaVersion = 1

	// approverTypeUser is the only GitHub account type that may approve an
	// accepted risk. Bot and agent identities are excluded by contract.
	approverTypeUser = "User"

	// evidenceTypeSignedCommit is the only accepted approval evidence format. An
	// Issue or pull request URL is deliberately not accepted because its body
	// remains editable after the approval was given.
	evidenceTypeSignedCommit = "signed-commit"

	// maxGitHubLoginLength is the longest login GitHub issues.
	maxGitHubLoginLength = 39
)

// githubLoginPattern is the character-level shape of a GitHub login: ASCII
// alphanumerics separated by single hyphens, never starting or ending with one.
var githubLoginPattern = regexp.MustCompile(`^[0-9A-Za-z](?:-?[0-9A-Za-z])*$`)

// nonHumanLoginMarkers are login fragments that identify an automation or agent
// account. The list deliberately over-rejects: a rejected approval is a review
// task, while an accepted agent approval would silently defeat the authority
// gate this package exists to enforce.
var nonHumanLoginMarkers = []string{
	"[bot]",
	"-bot",
	"_bot",
	"dependabot",
	"renovate",
	"github-actions",
	"copilot",
	"claude",
	"codex",
	"gemini",
	"chatgpt",
}

// policy is the trusted-base authority SSOT read from security/policy.yaml.
type policy struct {
	Version      int              `yaml:"version"`
	Organization policyOrg        `yaml:"organization"`
	Approvers    []policyApprover `yaml:"approvers"`
	Evidence     policyEvidence   `yaml:"evidence"`
	Cadence      policyCadence    `yaml:"cadence"`
}

type policyOrg struct {
	ID    int64  `yaml:"id"`
	Login string `yaml:"login"`
}

// policyApprover identifies one authority. Matching is performed on ID because
// the immutable numeric GitHub user ID survives a login rename; Login is a
// display label only.
//
// SigningKeys is the trust mapping from this authority to the keys it may sign
// approval commits with. It must be present as an explicit list. An empty list
// means no signature can ever satisfy this approver, which is the fail-closed
// state to leave it in until a real fingerprint is registered.
type policyApprover struct {
	ID          int64    `yaml:"id"`
	Login       string   `yaml:"login"`
	Type        string   `yaml:"type"`
	Role        string   `yaml:"role"`
	SigningKeys []string `yaml:"signing_keys"`
}

type policyEvidence struct {
	AcceptedTypes []string `yaml:"accepted_types"`
}

type policyCadence struct {
	ReviewIntervalDays int `yaml:"review_interval_days"`
	HardSunsetDays     int `yaml:"hard_sunset_days"`
}

// decodePolicy parses the authority SSOT. The decoder is strictly closed: an
// unknown field is an error rather than an ignored value, so a typo in a field
// name can never silently widen the authority set.
func decodePolicy(contents []byte) (*policy, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)

	var decoded policy
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode security policy: %w", err)
	}
	if err := requireSingleDocument(decoder, "security policy"); err != nil {
		return nil, err
	}
	if err := decoded.validate(); err != nil {
		return nil, err
	}
	return &decoded, nil
}

// requireSingleDocument rejects anything that follows the first YAML document.
//
// Only a clean end of stream is accepted. Decoding the trailing document into a
// concrete type and reading any failure as "nothing follows" would let a scalar,
// a sequence, a wrong-typed mapping or a mapping with unknown fields ride along
// unnoticed — every shape that does not happen to decode is exactly the shape an
// appended document would take. The value is decoded into any so that what
// follows is judged on its presence, not on whether it happens to fit.
func requireSingleDocument(decoder *yaml.Decoder, subject string) error {
	var trailing any
	switch err := decoder.Decode(&trailing); {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return fmt.Errorf("%s must contain exactly one document", subject)
	default:
		return fmt.Errorf("%s must contain exactly one document: %w", subject, err)
	}
}

func (current policy) validate() error {
	if current.Version != supportedSchemaVersion {
		return fmt.Errorf("security policy version must be %d", supportedSchemaVersion)
	}
	if current.Organization.ID <= 0 || strings.TrimSpace(current.Organization.Login) == "" {
		return fmt.Errorf("security policy organization requires a positive id and a login")
	}
	if len(current.Approvers) == 0 {
		return fmt.Errorf("security policy must declare at least one approver")
	}
	seen := make(map[int64]struct{}, len(current.Approvers))
	for index, approver := range current.Approvers {
		if err := approver.validate(); err != nil {
			return fmt.Errorf("security policy approver %d: %w", index, err)
		}
		if _, duplicated := seen[approver.ID]; duplicated {
			return fmt.Errorf("security policy approver %d: duplicate approver id %d", index, approver.ID)
		}
		seen[approver.ID] = struct{}{}
	}
	if err := current.Evidence.validate(); err != nil {
		return err
	}
	return current.Cadence.validate()
}

func (current policyApprover) validate() error {
	if current.ID <= 0 {
		return fmt.Errorf("approver id must be a positive immutable numeric user id")
	}
	if strings.TrimSpace(current.Login) == "" || strings.TrimSpace(current.Role) == "" {
		return fmt.Errorf("approver login and role are required")
	}
	if current.Type != approverTypeUser {
		return fmt.Errorf("approver type must be %q", approverTypeUser)
	}
	if !isWellFormedLogin(current.Login) {
		return fmt.Errorf("approver login %q is not a well-formed GitHub login", current.Login)
	}
	if isNonHumanLogin(current.Login) {
		return fmt.Errorf("approver %q is an automation or agent identity", current.Login)
	}
	return current.validateSigningKeys()
}

// validateSigningKeys requires the trust mapping to be explicit. A missing list
// is an error rather than an empty default, because an omitted mapping and a
// deliberately empty one would otherwise be indistinguishable.
func (current policyApprover) validateSigningKeys() error {
	if current.SigningKeys == nil {
		return fmt.Errorf("approver signing_keys must be an explicit list; use [] when no key is registered yet")
	}
	seen := make(map[string]struct{}, len(current.SigningKeys))
	for _, key := range current.SigningKeys {
		if key != strings.TrimSpace(key) || key == "" {
			return fmt.Errorf("approver signing key %q must be a trimmed, non-empty fingerprint", key)
		}
		if _, duplicated := seen[key]; duplicated {
			return fmt.Errorf("approver declares duplicate signing key %q", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// authorizesSigningKey reports whether a fingerprint the verifier established
// from a signature belongs to this approver. An approver with no registered key
// authorizes nothing.
func (current policyApprover) authorizesSigningKey(fingerprint string) bool {
	if fingerprint == "" {
		return false
	}
	for _, key := range current.SigningKeys {
		if key == fingerprint {
			return true
		}
	}
	return false
}

func (current policyEvidence) validate() error {
	if len(current.AcceptedTypes) == 0 {
		return fmt.Errorf("security policy must declare at least one accepted evidence type")
	}
	for _, accepted := range current.AcceptedTypes {
		if accepted != evidenceTypeSignedCommit {
			return fmt.Errorf("evidence type %q is not an immutable approval evidence format", accepted)
		}
	}
	return nil
}

func (current policyCadence) validate() error {
	if current.ReviewIntervalDays <= 0 || current.HardSunsetDays <= 0 {
		return fmt.Errorf("security policy cadence days must be positive")
	}
	if current.HardSunsetDays < current.ReviewIntervalDays {
		return fmt.Errorf("security policy hard sunset must not precede the review interval")
	}
	return nil
}

// approverByID resolves an approver on the immutable numeric id only.
func (current policy) approverByID(id int64) (policyApprover, bool) {
	for _, approver := range current.Approvers {
		if approver.ID == id {
			return approver, true
		}
	}
	return policyApprover{}, false
}

func (current policy) acceptsEvidenceType(value string) bool {
	for _, accepted := range current.Evidence.AcceptedTypes {
		if accepted == value {
			return true
		}
	}
	return false
}

// isNonHumanLogin reports whether a login carries an automation marker. It is a
// substring scan, so it catches the names automation accounts are conventionally
// given and nothing else.
func isNonHumanLogin(login string) bool {
	lowered := strings.ToLower(strings.TrimSpace(login))
	for _, marker := range nonHumanLoginMarkers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
}

// isWellFormedLogin reports whether a login is one GitHub could actually issue:
// ASCII alphanumerics with single internal hyphens, at most 39 characters, no
// leading or trailing hyphen.
//
// The marker scan alone cannot see a Unicode homoglyph — a login spelled with a
// Cyrillic character passes every substring comparison against an ASCII marker
// while rendering identically to a name a reviewer would trust. A login that
// cannot exist cannot be an authority, so the character rule closes that vector
// completely, independently of what the marker list happens to contain.
func isWellFormedLogin(login string) bool {
	return len(login) <= maxGitHubLoginLength && githubLoginPattern.MatchString(login)
}

// isHumanApproverLogin combines both rules: a login must be one GitHub could
// issue and must carry no automation marker.
//
// Neither rule reaches an automation account whose name contains no marker —
// "gzh-release-automation" is a well-formed login and matches nothing in
// nonHumanLoginMarkers. That case is a human judgment about who an account
// belongs to, which no character or substring rule can decide; it is left to the
// review that adds an approver to security/policy.yaml.
func isHumanApproverLogin(login string) bool {
	return isWellFormedLogin(login) && !isNonHumanLogin(login)
}

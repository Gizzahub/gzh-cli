// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package acceptedrisk validates the trusted-base accepted-risk registry and its
// linkage to the standalone gosec suppressions present in the repository source.
package acceptedrisk

import (
	"bytes"
	"fmt"
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
)

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
type policyApprover struct {
	ID    int64  `yaml:"id"`
	Login string `yaml:"login"`
	Type  string `yaml:"type"`
	Role  string `yaml:"role"`
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
	if err := decoder.Decode(new(policy)); err == nil {
		return nil, fmt.Errorf("security policy must contain exactly one document")
	}
	if err := decoded.validate(); err != nil {
		return nil, err
	}
	return &decoded, nil
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
	if isNonHumanLogin(current.Login) {
		return fmt.Errorf("approver %q is an automation or agent identity", current.Login)
	}
	return nil
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

func isNonHumanLogin(login string) bool {
	lowered := strings.ToLower(strings.TrimSpace(login))
	for _, marker := range nonHumanLoginMarkers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
}

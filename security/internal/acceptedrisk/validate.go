// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package acceptedrisk

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Violation codes. Every code is a blocking condition: the registry is valid only
// when the returned slice is empty.
const (
	codeApproverNotInPolicy     = "approver-not-in-policy"
	codeApproverNotHuman        = "approver-not-human"
	codeApproverLoginMismatch   = "approver-login-mismatch"
	codeEvidencePending         = "approval-evidence-pending"
	codeEvidenceTypeUnsupported = "approval-evidence-type-unsupported"
	codeEvidenceSHAMalformed    = "approval-evidence-sha-malformed"
	codeEvidenceUnverified      = "approval-evidence-unverified"
	codeSignerNotAuthorized     = "approval-signer-not-authorized"
	codeApprovalScopeMismatch   = "approval-commit-scope-mismatch"
	codeSuppressionBlanket      = "suppression-blanket-form"
	codeSuppressionUnscanned    = "suppression-unscanned-file"
	codeSuppressionMalformed    = "suppression-malformed"
	codeSuppressionUnregistered = "suppression-unregistered"
	codeSuppressionRuleMismatch = "suppression-rule-mismatch"
	codeSuppressionPathMismatch = "suppression-path-mismatch"
	codeSuppressionDuplicateRef = "suppression-duplicate-reference"
	codeRecordOrphaned          = "record-orphaned"
	codeRecordDateInFuture      = "record-date-in-future"
	codeReviewOverdue           = "review-overdue"
	codeHardSunsetExpired       = "hard-sunset-expired"
	codeRenewalAfterHardSunset  = "renewal-after-hard-sunset"
)

// verifiedCommit is what a commit verifier established about one approval
// commit. Every field describes what the verifier *proved* from the signature,
// never what the commit claims about itself: a commit author or committer header
// is attacker-controlled text, so it is not represented here at all.
type verifiedCommit struct {
	// SHA is the commit the verifier actually inspected.
	SHA string
	// Verified is true only when the signature itself validated.
	Verified bool
	// VerifiedSignerKey is the fingerprint of the key the signature was made
	// with, as established by the verifier. It is empty when the verifier could
	// not establish one, which is a blocking condition.
	VerifiedSignerKey string
	// VerifiedSignerAccountID is the forge account the verifier resolved the
	// signing key to, or 0 when the verifier cannot resolve one. When it is
	// non-zero it must equal the record's approver id.
	VerifiedSignerAccountID int64
	// Message is the approval commit's message as read from the verified
	// commit. The signature covers it, so it can carry approval scope.
	Message string
}

// commitVerifier reports what it can prove about an approval commit.
// Implementations wrap `git verify-commit` (and the key-to-account mapping their
// forge exposes); this package requires one rather than shelling out itself so
// that the authority rules stay independent of any particular repository access
// mechanism.
//
// An implementation must never populate the signer fields from the commit's own
// author or committer headers. Those are unauthenticated strings that any pusher
// can set, and treating them as identity would defeat this whole gate.
type commitVerifier interface {
	VerifyCommit(sha string) (verifiedCommit, error)
}

// violation is one blocking finding about the registry or its linkage.
type violation struct {
	Code    string
	Subject string
	Detail  string
}

// validationInput is the complete, explicit input to a registry validation. Every
// field is required; a missing field is an error rather than a permissive default.
type validationInput struct {
	Policy       *policy
	Records      []record
	Suppressions []suppression
	Verifier     commitVerifier
	Now          time.Time
}

// validateRegistry returns every blocking violation, sorted so that the output is
// independent of map iteration and input order. A non-nil error means the input
// itself could not be evaluated, which is also a blocking condition.
func validateRegistry(input validationInput) ([]violation, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}

	violations := make([]violation, 0)
	byID := make(map[string]record, len(input.Records))
	for _, current := range input.Records {
		byID[current.ID] = current
		violations = append(violations, approvalViolations(*input.Policy, current, input.Verifier)...)
		violations = append(violations, cadenceViolations(input.Policy.Cadence, current, input.Now)...)
	}

	linkage, linkageViolations := suppressionViolations(input.Suppressions, byID)
	violations = append(violations, linkageViolations...)
	for _, current := range input.Records {
		if linkage[current.ID] == 0 {
			violations = append(violations, violation{
				Code:    codeRecordOrphaned,
				Subject: current.ID,
				Detail:  "no source suppression references this accepted-risk record",
			})
		}
	}

	sortViolations(violations)
	return violations, nil
}

func (input validationInput) validate() error {
	if input.Policy == nil {
		return fmt.Errorf("trusted-base security policy is required")
	}
	if input.Records == nil {
		return fmt.Errorf("accepted-risk records must be an explicit list")
	}
	if input.Suppressions == nil {
		return fmt.Errorf("source suppressions must be an explicit list")
	}
	if input.Verifier == nil {
		return fmt.Errorf("commit verifier is required to check approval evidence")
	}
	if input.Now.IsZero() {
		return fmt.Errorf("evaluation time is required")
	}
	return nil
}

// approvalViolations checks that the approver is a human authority declared in the
// trusted base and that the approval evidence is immutable and verifiable.
func approvalViolations(current policy, entry record, verifier commitVerifier) []violation {
	approver, found := current.approverByID(entry.Approver.ID)
	if !found {
		return []violation{{
			Code:    codeApproverNotInPolicy,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("approver id %d is not declared in the trusted-base security policy", entry.Approver.ID),
		}}
	}

	violations := make([]violation, 0)
	if approver.Type != approverTypeUser || !isHumanApproverLogin(approver.Login) || !isHumanApproverLogin(entry.Approver.Login) {
		violations = append(violations, violation{
			Code:    codeApproverNotHuman,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("approver id %d does not resolve to a human authority with a well-formed login", entry.Approver.ID),
		})
	}
	if approver.Login != entry.Approver.Login {
		// The numeric id is authoritative, but a stale label in the registry means
		// the record was not re-reviewed after a rename.
		violations = append(violations, violation{
			Code:    codeApproverLoginMismatch,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("registry login %q does not match policy login %q for id %d", entry.Approver.Login, approver.Login, approver.ID),
		})
	}
	return append(violations, evidenceViolations(current, approver, entry, verifier)...)
}

// evidenceViolations checks that the approval evidence is immutable, that it
// verifies, that the identity the signature proves is one this approver is
// trusted to sign with, and that the approval commit names the record it
// approves.
func evidenceViolations(current policy, approver policyApprover, entry record, verifier commitVerifier) []violation {
	if !current.acceptsEvidenceType(entry.Evidence.Type) {
		return []violation{{
			Code:    codeEvidenceTypeUnsupported,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("evidence type %q is not an accepted immutable approval format", entry.Evidence.Type),
		}}
	}
	if entry.Evidence.SHA == pendingApprovalSentinel {
		return []violation{{
			Code:    codeEvidencePending,
			Subject: entry.ID,
			Detail:  "approval evidence is the explicit pending sentinel; the policy owner has not approved this accepted risk",
		}}
	}
	if !commitSHAPattern.MatchString(entry.Evidence.SHA) {
		return []violation{{
			Code:    codeEvidenceSHAMalformed,
			Subject: entry.ID,
			Detail:  "approval evidence sha must be a full 40-character lowercase hex commit id",
		}}
	}

	verified, err := verifier.VerifyCommit(entry.Evidence.SHA)
	switch {
	case err != nil:
		return []violation{{
			Code:    codeEvidenceUnverified,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("approval commit signature could not be verified: %v", err),
		}}
	case !verified.Verified || verified.SHA != entry.Evidence.SHA:
		return []violation{{
			Code:    codeEvidenceUnverified,
			Subject: entry.ID,
			Detail:  "approval commit signature did not verify for the recorded commit",
		}}
	}

	if unauthorized := signerViolations(approver, entry, verified); unauthorized != nil {
		return unauthorized
	}
	return scopeViolations(entry, verified)
}

// signerViolations binds the approval to the approver. A signature that merely
// verifies proves only that somebody signed; it is accepted only when the key it
// was made with is registered to this record's approver in the trusted base.
func signerViolations(approver policyApprover, entry record, verified verifiedCommit) []violation {
	if len(approver.SigningKeys) == 0 {
		return []violation{{
			Code:    codeSignerNotAuthorized,
			Subject: entry.ID,
			Detail: fmt.Sprintf("no signing key is registered for approver id %d, so no signature can satisfy this approval",
				approver.ID),
		}}
	}
	if verified.VerifiedSignerKey == "" {
		return []violation{{
			Code:    codeSignerNotAuthorized,
			Subject: entry.ID,
			Detail:  "the verifier did not establish which key signed the approval commit",
		}}
	}
	if !approver.authorizesSigningKey(verified.VerifiedSignerKey) {
		return []violation{{
			Code:    codeSignerNotAuthorized,
			Subject: entry.ID,
			Detail: fmt.Sprintf("approval commit was signed with key %q, which is not registered to approver id %d",
				verified.VerifiedSignerKey, approver.ID),
		}}
	}
	if verified.VerifiedSignerAccountID != 0 && verified.VerifiedSignerAccountID != approver.ID {
		return []violation{{
			Code:    codeSignerNotAuthorized,
			Subject: entry.ID,
			Detail: fmt.Sprintf("approval commit was signed by account id %d, not by approver id %d",
				verified.VerifiedSignerAccountID, approver.ID),
		}}
	}
	return nil
}

// scopeViolations requires the signed approval commit to name the record it
// approves, so that one signature cannot silently ratify every record in the
// registry. The signature covers the message, so the identifier in it is as
// tamper-evident as the signature itself.
func scopeViolations(entry record, verified verifiedCommit) []violation {
	if strings.TrimSpace(verified.Message) == "" {
		return []violation{{
			Code:    codeApprovalScopeMismatch,
			Subject: entry.ID,
			Detail:  "the verifier did not report the approval commit message, so the approval scope cannot be established",
		}}
	}
	if !mentionsRiskID(verified.Message, entry.ID) {
		return []violation{{
			Code:    codeApprovalScopeMismatch,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("approval commit %s does not name %s in its message", verified.SHA, entry.ID),
		}}
	}
	return nil
}

// mentionsRiskID matches the identifier on non-alphanumeric boundaries so that
// AR-2026-001 is not satisfied by a longer token that merely starts with it.
func mentionsRiskID(message, riskID string) bool {
	pattern := regexp.MustCompile(`(^|[^0-9A-Za-z])` + regexp.QuoteMeta(riskID) + `([^0-9A-Za-z]|$)`)
	return pattern.MatchString(message)
}

// cadenceViolations applies the approved 90-day review and 180-day hard sunset.
//
// It also rejects a postdated record. The decoder can only check the dates
// against each other; whether a date is in the future is a question about the
// clock, so it is answered here, against the same instant every cadence rule
// uses. A postdated last_reviewed_at would push review-overdue out for as long
// as the date is ahead, and a postdated created_at would push the hard sunset
// out with it.
func cadenceViolations(cadence policyCadence, entry record, now time.Time) []violation {
	violations := make([]violation, 0)
	sunset := entry.hardSunset(cadence)
	for _, dated := range []struct {
		field string
		value time.Time
	}{
		{field: "created_at", value: entry.CreatedAt},
		{field: "last_reviewed_at", value: entry.LastReviewedAt},
	} {
		if dated.value.After(now) {
			violations = append(violations, violation{
				Code:    codeRecordDateInFuture,
				Subject: entry.ID,
				Detail: fmt.Sprintf("%s %s is later than the evaluation date %s", dated.field,
					dated.value.Format(dateLayout), now.Format(dateLayout)),
			})
		}
	}
	if reviewBy := entry.reviewBy(cadence); now.After(reviewBy) {
		violations = append(violations, violation{
			Code:    codeReviewOverdue,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("review was due %s", reviewBy.Format(dateLayout)),
		})
	}
	if now.After(sunset) {
		violations = append(violations, violation{
			Code:    codeHardSunsetExpired,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("hard sunset passed %s; a new risk analysis, approval and identifier are required", sunset.Format(dateLayout)),
		})
	}
	if entry.LastReviewedAt.After(sunset) {
		violations = append(violations, violation{
			Code:    codeRenewalAfterHardSunset,
			Subject: entry.ID,
			Detail:  fmt.Sprintf("last_reviewed_at %s renews an identifier past its %s hard sunset", entry.LastReviewedAt.Format(dateLayout), sunset.Format(dateLayout)),
		})
	}
	return violations
}

// suppressionViolations checks every directive against the registry and returns
// how many directives referenced each identifier.
func suppressionViolations(suppressions []suppression, byID map[string]record) (map[string]int, []violation) {
	linkage := make(map[string]int, len(byID))
	violations := make([]violation, 0)
	for _, current := range suppressions {
		if current.Unscanned {
			// Reported before the blanket and linkage checks because it
			// subsumes them: in a file the scan never loads, no directive of
			// any grammar suppresses anything. Registering an accepted risk
			// for it would record a risk gosec never evaluates, and silently
			// skipping it would let the file become a place to hide directives.
			violations = append(violations, violation{
				Code:    codeSuppressionUnscanned,
				Subject: current.location(),
				Detail: fmt.Sprintf("%q sits in a file the pinned scan never loads, because its build constraint requires a tag the pinned flags do not set; it suppresses nothing and must be removed rather than registered",
					current.Raw),
			})
			continue
		}
		if current.Blanket {
			// The blanket grammar carries no identifier and cannot be given
			// one, so it is never a malformed directive to be corrected: it is
			// a suppression that this registry cannot account for at all.
			violations = append(violations, violation{
				Code:    codeSuppressionBlanket,
				Subject: current.location(),
				Detail: fmt.Sprintf("%q is gosec's blanket suppression form, which names no accepted-risk record and can never be registered; replace it with a %s directive",
					current.Raw, directivePrefix),
			})
			continue
		}
		if current.RiskID == "" {
			violations = append(violations, violation{
				Code:    codeSuppressionMalformed,
				Subject: current.location(),
				Detail:  fmt.Sprintf("directive %q is not of the form //gosec:disable Gxxx -- AR-YYYY-NNN <reason>", current.Raw),
			})
			continue
		}
		linkage[current.RiskID]++
		entry, registered := byID[current.RiskID]
		if !registered {
			violations = append(violations, violation{
				Code:    codeSuppressionUnregistered,
				Subject: current.location(),
				Detail:  fmt.Sprintf("%s has no accepted-risk record", current.RiskID),
			})
			continue
		}
		if entry.Rule != current.Rule {
			violations = append(violations, violation{
				Code:    codeSuppressionRuleMismatch,
				Subject: current.location(),
				Detail:  fmt.Sprintf("%s covers rule %s, not %s", entry.ID, entry.Rule, current.Rule),
			})
		}
		if entry.Path != current.Path {
			violations = append(violations, violation{
				Code:    codeSuppressionPathMismatch,
				Subject: current.location(),
				Detail:  fmt.Sprintf("%s covers %s, not %s", entry.ID, entry.Path, current.Path),
			})
		}
	}

	for _, id := range sortedCounts(linkage) {
		if linkage[id] > 1 {
			violations = append(violations, violation{
				Code:    codeSuppressionDuplicateRef,
				Subject: id,
				Detail:  fmt.Sprintf("%d suppressions reference one accepted-risk record; each record covers exactly one site", linkage[id]),
			})
		}
	}
	return linkage, violations
}

func sortViolations(values []violation) {
	sort.Slice(values, func(left, right int) bool {
		leftKey := strings.Join([]string{values[left].Code, values[left].Subject, values[left].Detail}, "\x00")
		rightKey := strings.Join([]string{values[right].Code, values[right].Subject, values[right].Detail}, "\x00")
		return leftKey < rightKey
	})
}

// violationsError folds violations into one blocking error so a caller cannot
// mistake a non-empty result for success.
func violationsError(values []violation) error {
	if len(values) == 0 {
		return nil
	}
	joined := make([]error, 0, len(values))
	for _, current := range values {
		joined = append(joined, fmt.Errorf("%s [%s]: %s", current.Code, current.Subject, current.Detail))
	}
	return errors.Join(joined...)
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedCounts(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

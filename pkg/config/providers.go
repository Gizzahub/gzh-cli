// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package config

import "fmt"

// BulkCloneResult contains the results of a bulk clone operation.
type BulkCloneResult struct {
	TotalTargets      int            `json:"totalTargets"`
	SuccessfulTargets int            `json:"successfulTargets"`
	FailedTargets     int            `json:"failedTargets"`
	SkippedTargets    int            `json:"skippedTargets"`
	Results           []TargetResult `json:"results"`
}

// TargetResult contains the result of cloning a single target.
type TargetResult struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	CloneDir string `json:"cloneDir"`
	Strategy string `json:"strategy"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// GetSummary returns a summary of the bulk clone operation.
func (r *BulkCloneResult) GetSummary() string {
	return fmt.Sprintf("Total: %d, Successful: %d, Failed: %d, Skipped: %d",
		r.TotalTargets, r.SuccessfulTargets, r.FailedTargets, r.SkippedTargets)
}

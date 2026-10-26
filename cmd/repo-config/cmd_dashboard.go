// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"fmt"
	"net/http"
)

// runDashboardCommand rejects the unimplemented dashboard before any side effects.
func runDashboardCommand(_ GlobalFlags, _ int, _ bool, _ int) error {
	return fmt.Errorf("repository compliance dashboard is unsupported: no dashboard backend is implemented")
}

// handleDashboardHome reports the unsupported dashboard without fabricated HTML.
func handleDashboardHome(w http.ResponseWriter, _ *http.Request, _ string, _ bool, _ int) {
	http.Error(w, "repository compliance dashboard is unsupported", http.StatusNotImplemented)
}

// handleRepositoriesAPI reports the unsupported dashboard repository endpoint.
func handleRepositoriesAPI(w http.ResponseWriter, _ *http.Request, _, _ string) {
	http.Error(w, "dashboard repository reports are unsupported", http.StatusNotImplemented)
}

// handleComplianceAPI reports the unsupported dashboard compliance endpoint.
func handleComplianceAPI(w http.ResponseWriter, _ *http.Request, _, _ string) {
	http.Error(w, "dashboard compliance reports are unsupported", http.StatusNotImplemented)
}

// handleStaticAssets has no assets to serve.
func handleStaticAssets(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

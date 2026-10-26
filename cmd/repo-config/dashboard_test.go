// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func captureDashboardStdout(t *testing.T, run func()) string {
	t.Helper()
	read, write, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = original; _ = read.Close(); _ = write.Close() }()
	run()
	require.NoError(t, write.Close())
	data, err := io.ReadAll(read)
	require.NoError(t, err)
	return string(data)
}

func TestDashboardDirectUnsupported(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	file := filepath.Join(root, "existing")
	require.NoError(t, os.WriteFile(file, []byte("unchanged"), 0o600))
	output := captureDashboardStdout(t, func() {
		err := runDashboardCommand(GlobalFlags{Organization: "org", Verbose: true, DryRun: true}, -1, true, 1)
		require.ErrorContains(t, err, "unsupported")
	})
	require.Empty(t, output)
	data, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, "unchanged", string(data))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.ErrorContains(t, runDashboardCommand(GlobalFlags{}, -1, false, 0), "unsupported")
}

func TestDashboardHTTPUnsupported(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"home":         func(w http.ResponseWriter, r *http.Request) { handleDashboardHome(w, r, "org", true, 1) },
		"repositories": func(w http.ResponseWriter, r *http.Request) { handleRepositoriesAPI(w, r, "org", "unused-token") },
		"compliance":   func(w http.ResponseWriter, r *http.Request) { handleComplianceAPI(w, r, "org", "unused-token") },
	}
	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
			require.Equal(t, http.StatusNotImplemented, response.Code)
			require.Contains(t, response.Body.String(), "unsupported")
			for _, fake := range []string{"api-server", "web-frontend", "85.5", "overallScore", "complianceRate", "<!DOCTYPE"} {
				require.NotContains(t, response.Body.String(), fake)
			}
		})
	}
	response := httptest.NewRecorder()
	handleStaticAssets(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/static/", nil))
	require.Equal(t, http.StatusNotFound, response.Code)
}

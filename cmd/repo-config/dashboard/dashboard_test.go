// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package dashboard

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

func TestDashboardCommandUnsupported(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--org", "org"},
		{"--org", "org", "--dry-run", "--verbose", "--auto-refresh", "--refresh-rate", "1", "--port", "-1"},
		{"--org", "org", "--config-file", "missing", "--token", "unused", "--parallel", "2", "--timeout", "1s", "--port", "-1"},
	} {
		cmd := NewCmd()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		// Invalid port prevents binding even if the old listener implementation returns.
		require.NoError(t, cmd.Flags().Set("port", "-1"))
		cmd.SetArgs(args)
		output := captureDashboardStdout(t, func() { require.ErrorContains(t, cmd.Execute(), "unsupported") })
		require.Empty(t, output)
	}
	cmd := NewCmd()
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"--help"})
	require.NoError(t, cmd.Execute())
	for _, name := range []string{"org", "config-file", "token", "dry-run", "verbose", "parallel", "timeout", "port", "auto-refresh", "refresh-rate"} {
		require.NotNil(t, cmd.Flags().Lookup(name))
	}
}

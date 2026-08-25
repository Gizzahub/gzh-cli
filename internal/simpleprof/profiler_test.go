// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package simpleprof

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPprofServeMuxOwnsOnlyPprofRoutes(t *testing.T) {
	mux := newPprofServeMux()

	for _, path := range []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Errorf("mux has no handler for %q", path)
		}

		postRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
		postResponse := httptest.NewRecorder()
		mux.ServeHTTP(postResponse, postRequest)
		if postResponse.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s status = %d, want %d", path, postResponse.Code, http.StatusMethodNotAllowed)
		}
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated endpoint status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestStartHTTPServerUsesDedicatedIPv4LoopbackHandler(t *testing.T) {
	profiler := NewSimpleProfiler(t.TempDir())
	if err := profiler.StartHTTPServer(0); err != nil {
		t.Fatalf("StartHTTPServer() error: %v", err)
	}
	t.Cleanup(func() {
		if err := profiler.StopHTTPServer(context.Background()); err != nil {
			t.Errorf("StopHTTPServer() error: %v", err)
		}
	})

	if profiler.server.Addr != "127.0.0.1:0" {
		t.Fatalf("server address = %q, want explicit IPv4 loopback", profiler.server.Addr)
	}
	if profiler.server.Handler == nil || profiler.server.Handler == http.DefaultServeMux {
		t.Fatalf("server handler = %#v, want dedicated mux", profiler.server.Handler)
	}
}

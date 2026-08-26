// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package simpleprof

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLastGCTime(t *testing.T) {
	tests := []struct {
		name   string
		lastGC uint64
		want   time.Time
	}{
		{name: "zero preserves Unix epoch", lastGC: 0, want: time.Unix(0, 0)},
		{name: "normal timestamp", lastGC: 1_234_567_890, want: time.Unix(0, 1_234_567_890)},
		{name: "maximum signed timestamp", lastGC: math.MaxInt64, want: time.Unix(9_223_372_036, 854_775_807)},
		{name: "signed boundary plus one", lastGC: uint64(math.MaxInt64) + 1, want: time.Unix(9_223_372_036, 854_775_808)},
		{name: "maximum unsigned timestamp", lastGC: math.MaxUint64, want: time.Unix(18_446_744_073, 709_551_615)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastGCTime(tt.lastGC); !got.Equal(tt.want) {
				t.Errorf("LastGCTime(%d) = %v, want %v", tt.lastGC, got, tt.want)
			}
		})
	}
}

func TestPprofServeMuxOwnsOnlyPprofRoutes(t *testing.T) {
	assertPprofServeMuxContract(t, newPprofServeMux())
}

func TestPprofServeMuxSupportsGo121CompatibilityMode(t *testing.T) {
	t.Setenv("GODEBUG", "httpmuxgo121=1")
	assertPprofServeMuxContract(t, newPprofServeMux())
}

func assertPprofServeMuxContract(t *testing.T, mux *http.ServeMux) {
	t.Helper()

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
		if allow := postResponse.Header().Get("Allow"); allow != "GET, HEAD" {
			t.Errorf("POST %s Allow = %q, want %q", path, allow, "GET, HEAD")
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

package gzhclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHTTPClientWrapperOriginValidation(t *testing.T) {
	t.Run("same origin request allowed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		wrapper := newHTTPClientWrapper(server.Client(), server.URL)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/repos", http.NoBody)
		if err != nil {
			t.Fatalf("NewRequestWithContext() unexpected error: %v", err)
		}

		resp, err := wrapper.Do(req)
		if err != nil {
			t.Fatalf("Do() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("response body close failed: %v", err)
			}
		})
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Do() status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("wrong origin rejected before round trip", func(t *testing.T) {
		var roundTrips atomic.Int32
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			roundTrips.Add(1)
			return nil, fmt.Errorf("unexpected round trip")
		})}
		wrapper := newHTTPClientWrapper(client, "https://api.github.com")

		for _, rawURL := range []string{
			"http://api.github.com/repos",
			"https://example.com/repos",
			"https://api.github.com:8443/repos",
		} {
			t.Run(rawURL, func(t *testing.T) {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, http.NoBody)
				if err != nil {
					t.Fatalf("NewRequestWithContext() unexpected error: %v", err)
				}
				resp, err := wrapper.Do(req)
				if resp != nil {
					if closeErr := resp.Body.Close(); closeErr != nil {
						t.Errorf("response body close failed: %v", closeErr)
					}
				}
				if err == nil {
					t.Fatal("Do() expected origin error")
				}
			})
		}

		if got := roundTrips.Load(); got != 0 {
			t.Fatalf("RoundTrip calls = %d, want 0", got)
		}
	})

	t.Run("same origin redirect allowed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/start" {
				http.Redirect(w, r, "/finish", http.StatusFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		wrapper := newHTTPClientWrapper(server.Client(), server.URL)
		resp, err := wrapper.Get(server.URL + "/start")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("response body close failed: %v", err)
			}
		})
		if resp.Request.URL.Path != "/finish" {
			t.Fatalf("Get() final path = %q, want %q", resp.Request.URL.Path, "/finish")
		}
	})

	t.Run("cross origin redirect rejected", func(t *testing.T) {
		var destinationHits atomic.Int32
		destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			destinationHits.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer destination.Close()

		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination.URL, http.StatusFound)
		}))
		defer origin.Close()

		wrapper := newHTTPClientWrapper(origin.Client(), origin.URL)
		resp, err := wrapper.Get(origin.URL)
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("response body close failed: %v", closeErr)
			}
		}
		if err == nil {
			t.Fatal("Get() expected cross-origin redirect error")
		}
		if got := destinationHits.Load(); got != 0 {
			t.Fatalf("destination hits = %d, want 0", got)
		}
	})

	t.Run("Get and Post share validation", func(t *testing.T) {
		var roundTrips atomic.Int32
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			roundTrips.Add(1)
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Header:     make(http.Header),
			}, nil
		})}
		wrapper := newHTTPClientWrapper(client, "https://api.github.com")

		resp, err := wrapper.Get("https://example.com/repos")
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("response body close failed: %v", closeErr)
			}
		}
		if err == nil {
			t.Fatal("Get() expected origin error")
		}
		resp, err = wrapper.Post("https://example.com/repos", "application/json", map[string]string{"name": "repo"})
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("response body close failed: %v", closeErr)
			}
		}
		if err == nil {
			t.Fatal("Post() expected origin error")
		}
		if got := roundTrips.Load(); got != 0 {
			t.Fatalf("RoundTrip calls = %d, want 0", got)
		}
	})
}

func TestClient_UpdateConfig(t *testing.T) {
	tests := []struct {
		name   string
		update ClientConfig
	}{
		{
			name: "shorter timeout",
			update: func() ClientConfig {
				cfg := DefaultConfig()
				cfg.Timeout = 10 * time.Second
				return cfg
			}(),
		},
		{
			name: "change log level",
			update: func() ClientConfig {
				cfg := DefaultConfig()
				cfg.LogLevel = "debug"
				return cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(DefaultConfig())
			assert.NoError(t, err)

			err = c.UpdateConfig(tt.update)
			assert.NoError(t, err)
			assert.Equal(t, tt.update, c.GetConfig())
		})
	}
}

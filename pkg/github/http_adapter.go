package github

import (
	"context"
	"io"
	"net/http"
	"time"
)

// HTTPClientAdapter adapts the standard http.Client to the HTTPClient interface.
type HTTPClientAdapter struct {
	client *http.Client
}

// NewHTTPClientAdapter creates a new HTTP client adapter.
func NewHTTPClientAdapter() HTTPClient {
	return &HTTPClientAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewHTTPClientAdapterWithClient creates a new HTTP client adapter with a custom client.
func NewHTTPClientAdapterWithClient(client *http.Client) HTTPClient {
	return &HTTPClientAdapter{
		client: client,
	}
}

// Do performs an HTTP request.
func (a *HTTPClientAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.client.Do(req)
}

// Get performs a GET request.
func (a *HTTPClientAdapter) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return a.client.Do(req)
}

// Post performs a POST request.
func (a *HTTPClientAdapter) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return a.client.Do(req)
}

// Put performs a PUT request.
func (a *HTTPClientAdapter) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return a.client.Do(req)
}

// Patch performs a PATCH request.
func (a *HTTPClientAdapter) Patch(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return a.client.Do(req)
}

// Delete performs a DELETE request.
func (a *HTTPClientAdapter) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return a.client.Do(req)
}

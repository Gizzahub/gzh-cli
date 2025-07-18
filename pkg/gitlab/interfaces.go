package gitlab

import (
	"context"
	"io"
	"net/http"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	// Do performs an HTTP request
	Do(req *http.Request) (*http.Response, error)
	
	// Get performs a GET request
	Get(ctx context.Context, url string) (*http.Response, error)
	
	// Post performs a POST request
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	
	// Put performs a PUT request
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	
	// Patch performs a PATCH request
	Patch(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	
	// Delete performs a DELETE request  
	Delete(ctx context.Context, url string) (*http.Response, error)
}
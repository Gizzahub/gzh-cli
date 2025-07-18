package mocks

import (
	"bytes"
	"io"
	"net/http"
)

// MockHTTPClient provides a mock HTTP client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
	Calls  []http.Request
}

// Do implements the HTTPClient interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Calls = append(m.Calls, *req)
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
	}, nil
}

// Reset clears the recorded calls
func (m *MockHTTPClient) Reset() {
	m.Calls = nil
}

// MockRoundTripper provides a mock HTTP round tripper
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip implements the http.RoundTripper interface
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
	}, nil
}

// NewMockResponse creates a mock HTTP response
func NewMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     make(http.Header),
	}
}

// NewMockJSONResponse creates a mock HTTP response with JSON content type
func NewMockJSONResponse(statusCode int, body string) *http.Response {
	resp := NewMockResponse(statusCode, body)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}
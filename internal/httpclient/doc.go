// Package httpclient provides HTTP client abstraction interfaces and
// implementations for making external HTTP requests.
//
// This package defines the HTTPClient interface that abstracts HTTP operations,
// enabling easy testing through mocking and providing consistent HTTP behavior
// across different components of the GZH Manager system.
//
// Key Components:
//
// HTTPClient Interface:
//   - Standard HTTP methods (GET, POST, PUT, DELETE, PATCH)
//   - Request/response handling and transformation
//   - Authentication and authorization
//   - Error handling and retry logic
//
// Implementations:
//   - StandardHTTPClient: HTTP client based on net/http
//   - RateLimitedHTTPClient: Client with rate limiting
//   - RetryHTTPClient: Client with automatic retry logic
//   - MockHTTPClient: Generated mock for unit testing
//
// Features:
//   - Configurable timeouts and connection limits
//   - Automatic retry with exponential backoff
//   - Request/response logging and metrics
//   - Custom header and authentication handling
//   - SSL/TLS configuration and validation
//
// Authentication Support:
//   - Bearer token authentication
//   - API key authentication
//   - Basic authentication
//   - Custom authentication schemes
//
// Example usage:
//
//	client := httpclient.NewStandardHTTPClient(timeout)
//	resp, err := client.Get("https://api.github.com/user")
//	data, err := client.PostJSON(url, payload)
//
// The abstraction enables consistent HTTP operations throughout the
// application while supporting comprehensive testing and monitoring.
package httpclient
